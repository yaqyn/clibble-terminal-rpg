package main

import (
	"bytes"
	"context"
	"errors"
	"github.com/yaqyn/dexrove/internal/pokeapi"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCleanInput(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  []string
	}{
		{"  hello  world  ", []string{"hello", "world"}},
		{"Charmander Bulbasaur PIKACHU", []string{"charmander", "bulbasaur", "pikachu"}},
		{"\tMAP\nmapb\r", []string{"map", "mapb"}},
		{"", []string{}},
	} {
		if got := cleanInput(tc.input); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("cleanInput(%q) = %v; want %v", tc.input, got, tc.want)
		}
	}
}

type fakeAPI struct {
	calls   int
	page    pokeapi.Locations
	pokemon pokeapi.Pokemon
	err     error
}

func (f *fakeAPI) ListLocations(context.Context, *string) (pokeapi.Locations, error) {
	f.calls++
	return f.page, f.err
}
func (f *fakeAPI) Explore(context.Context, string) (pokeapi.Area, error) {
	f.calls++
	return pokeapi.Area{}, f.err
}
func (f *fakeAPI) GetPokemon(context.Context, string) (pokeapi.Pokemon, error) {
	f.calls++
	return f.pokemon, f.err
}
func testConfig() (*config, *bytes.Buffer, *fakeAPI) {
	output := new(bytes.Buffer)
	api := &fakeAPI{pokemon: pokeapi.Pokemon{Name: "pidgey", Height: 3, Weight: 18, BaseExperience: 50}}
	cfg := &config{commands: getCommands(), client: api, ctx: context.Background(), out: output, pokedex: make(map[string]pokeapi.Pokemon), random: func() float64 { return 0 }}
	return cfg, output, api
}
func TestREPLCatchInspectPokedex(t *testing.T) {
	cfg, out, api := testConfig()
	err := runREPL(cfg, strings.NewReader("\ninspect pidgey\nCATCH PIDGEY\ninspect pidgey\npokedex\nexit\ncatch pikachu\n"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"you have not caught that pokemon", "pidgey was caught!", "Name: pidgey", "Height: 3", "Weight: 18", "Your collection:\n - pidgey", "Leaving Dexrove. See you on the next trail."} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q in %s", want, out.String())
		}
	}
	if api.calls != 1 {
		t.Fatalf("inspect or exit triggered requests: %d", api.calls)
	}
}
func TestCatchEscapeAndSaveFailure(t *testing.T) {
	cfg, out, _ := testConfig()
	cfg.random = func() float64 { return 0.99 }
	if err := commandCatch(cfg, "pidgey"); err != nil {
		t.Fatal(err)
	}
	if len(cfg.pokedex) != 0 || !strings.Contains(out.String(), "pidgey escaped!") {
		t.Fatal("escape added a catch")
	}
	cfg.random = func() float64 { return 0 }
	cfg.save = func(map[string]pokeapi.Pokemon) error { return errors.New("disk full") }
	if err := commandCatch(cfg, "pidgey"); err == nil {
		t.Fatal("save failure hidden")
	}
	if len(cfg.pokedex) != 0 {
		t.Fatal("failed save retained catch")
	}
}
func TestREPLInvalidArgumentsAndUnknown(t *testing.T) {
	cfg, out, api := testConfig()
	if err := runREPL(cfg, strings.NewReader("what\ncatch\ncatch pidgey extra\nmap extra\ncatch ../secret\nexit\n")); err != nil {
		t.Fatal(err)
	}
	if api.calls != 0 {
		t.Fatal("invalid command made requests")
	}
	if !strings.Contains(out.String(), "Unknown command") || strings.Count(out.String(), "Usage:") != 3 {
		t.Fatal(out.String())
	}
}
func TestPaginationBoundariesAndFailure(t *testing.T) {
	cfg, out, api := testConfig()
	commandMapBack(cfg)
	if api.calls != 0 || !strings.Contains(out.String(), "first page") {
		t.Fatal("mapb requested before first page")
	}
	next := "next"
	api.page = pokeapi.Locations{Next: &next}
	if err := commandMap(cfg); err != nil {
		t.Fatal(err)
	}
	api.err = errors.New("offline")
	if err := commandMap(cfg); err == nil {
		t.Fatal("request failure hidden")
	}
	if cfg.next != &next || !cfg.started {
		t.Fatal("failure changed pagination")
	}
	api.err = nil
	api.page = pokeapi.Locations{}
	commandMap(cfg)
	calls := api.calls
	commandMap(cfg)
	if api.calls != calls || !strings.Contains(out.String(), "last page") {
		t.Fatal("last page requested again")
	}
}
func TestProbabilityDecreases(t *testing.T) {
	if catchProbability(50) <= catchProbability(300) {
		t.Fatal("stronger Pokemon should be harder")
	}
	if catchProbability(-1) != 1 {
		t.Fatal("negative experience invalid")
	}
}
func TestREPLEOFAndInputLimit(t *testing.T) {
	cfg, _, _ := testConfig()
	if err := runREPL(cfg, strings.NewReader("")); err != nil {
		t.Fatal(err)
	}
	if err := runREPL(cfg, strings.NewReader(strings.Repeat("x", 70*1024))); err == nil {
		t.Fatal("oversize input accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cfg.ctx = ctx
	if err := runREPL(cfg, strings.NewReader("map\n")); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func TestCatchAliasPreservesExistingPokemon(t *testing.T) {
	cfg, out, _ := testConfig()
	original := pokeapi.Pokemon{Name: "pidgey", Height: 99}
	cfg.pokedex["pidgey"] = original
	cfg.save = func(map[string]pokeapi.Pokemon) error { t.Fatal("duplicate should not save"); return nil }
	cfg.random = func() float64 { t.Fatal("duplicate should not roll again"); return 0 }
	if err := commandCatch(cfg, "16"); err != nil {
		t.Fatal(err)
	}
	if cfg.pokedex["pidgey"].Height != 99 || !strings.Contains(out.String(), "already in your collection") {
		t.Fatal("alias replaced existing catch")
	}
}

func TestREPLCancellationWhileInputIsBlocked(t *testing.T) {
	cfg, _, _ := testConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg.ctx = ctx
	input, writer := io.Pipe()
	defer input.Close()
	defer writer.Close()
	result := make(chan error, 1)
	go func() { result <- runREPL(cfg, input) }()
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation waited for input")
	}
}
