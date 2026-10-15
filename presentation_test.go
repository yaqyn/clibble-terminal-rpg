package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yaqyn/clibble-pokemon-cli/internal/pokeapi"
)

func TestLegacyCollectionIsReused(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, "pokedex", "pokedex.json")
	if err := os.MkdirAll(filepath.Dir(legacy), 0700); err != nil {
		t.Fatal(err)
	}
	if err := savePokedex(legacy, map[string]pokeapi.Pokemon{"pidgey": {Name: "pidgey"}}); err != nil {
		t.Fatal(err)
	}
	selected, err := defaultCollection(dir)
	if err != nil || selected != legacy {
		t.Fatalf("selected %s: %v", selected, err)
	}
	previous := filepath.Join(dir, "dexrove", "collection.json")
	if err := os.MkdirAll(filepath.Dir(previous), 0700); err != nil {
		t.Fatal(err)
	}
	if err := savePokedex(previous, map[string]pokeapi.Pokemon{"pidgey": {Name: "pidgey"}}); err != nil {
		t.Fatal(err)
	}
	selected, err = defaultCollection(dir)
	if err != nil || selected != previous {
		t.Fatalf("Dexrove collection should take priority over Pokedex: %s: %v", selected, err)
	}
	current := filepath.Join(dir, "clibble", "collection.json")
	if err := os.MkdirAll(filepath.Dir(current), 0700); err != nil {
		t.Fatal(err)
	}
	if err := savePokedex(current, map[string]pokeapi.Pokemon{}); err != nil {
		t.Fatal(err)
	}
	selected, err = defaultCollection(dir)
	if err != nil || selected != current {
		t.Fatalf("new collection should take priority: %v", err)
	}
}
func TestRichCLIHasReadableNarrowLayout(t *testing.T) {
	cfg, out, _ := testConfig()
	cfg.view = &terminalView{out: out, width: 40}
	cfg.pokedex["pidgey"] = pokeapi.Pokemon{Name: "pidgey", Height: 3, Weight: 18}
	commandHelp(cfg)
	commandInspect(cfg, "pidgey")
	commandPokedex(cfg)
	for _, want := range []string{"Explore", "Collect", "Session", "0.3 m", "1.8 kg", "Your collection"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %s", want)
		}
	}
	if strings.Contains(out.String(), "\x1b[") {
		t.Fatal("color leaked into plain view")
	}
}
func TestPipesStayPlainEvenWhenColorRequested(t *testing.T) {
	var out, errOut bytes.Buffer
	if err := runCLI(context.Background(), []string{"-ephemeral", "-color", "always", "collection"}, strings.NewReader(""), &out, &errOut, false); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "\x1b[") || strings.Contains(out.String(), "clibble >") {
		t.Fatal("script output has terminal presentation")
	}
	if !strings.Contains(out.String(), "Your collection:") {
		t.Fatal(out.String())
	}
}
