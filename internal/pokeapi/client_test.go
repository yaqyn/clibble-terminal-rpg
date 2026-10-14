package pokeapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yaqyn/dexrove/internal/pokecache"
)

func TestResponsesAreCached(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		switch r.URL.Path {
		case "/api/v2/location-area/":
			if r.URL.Query().Get("limit") != "20" {
				t.Error("missing page size")
			}
			fmt.Fprint(w, `{"next":null,"previous":null,"results":[{"name":"area"}]}`)
		case "/api/v2/location-area/area/":
			fmt.Fprint(w, `{"pokemon_encounters":[{"pokemon":{"name":"pidgey"}}]}`)
		case "/api/v2/pokemon/pidgey/":
			fmt.Fprint(w, `{"name":"pidgey","base_experience":50}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client, err := NewClient(server.URL+"/api/v2/", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	cache := pokecache.NewCache(time.Hour)
	defer cache.Close()
	client.SetCache(cache)
	for i := 0; i < 2; i++ {
		page, err := client.ListLocations(context.Background(), nil)
		if err != nil || len(page.Results) != 1 {
			t.Fatalf("page=%+v err=%v", page, err)
		}
		area, err := client.Explore(context.Background(), "area")
		if err != nil || len(area.PokemonEncounters) != 1 {
			t.Fatalf("area=%+v err=%v", area, err)
		}
		pokemon, err := client.GetPokemon(context.Background(), "pidgey")
		if err != nil || pokemon.Name != "pidgey" {
			t.Fatalf("pokemon=%+v err=%v", pokemon, err)
		}
	}
	if hits.Load() != 3 {
		t.Fatalf("got %d requests; want 3", hits.Load())
	}
}
func TestFailuresAreNotCached(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"not found", `{}`, 404}, {"rate limited", `{}`, 429}, {"server error", `{}`, 500},
		{"malformed", `{`, 200}, {"trailing data", `{} {}`, 200}, {"oversize", strings.Repeat("x", maxResponseBytes+1), 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			count := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count++
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			client, _ := NewClient(server.URL+"/", time.Second)
			cache := pokecache.NewCache(time.Hour)
			defer cache.Close()
			client.SetCache(cache)
			for i := 0; i < 2; i++ {
				if _, err := client.ListLocations(context.Background(), nil); err == nil {
					t.Fatal("expected error")
				}
			}
			if count != 2 {
				t.Fatalf("error cached: %d requests", count)
			}
		})
	}
}
func TestURLBoundaryAndRedirect(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		http.Redirect(w, r, "https://example.com/", http.StatusFound)
	}))
	defer server.Close()
	client, _ := NewClient(server.URL+"/api/v2/", time.Second)
	for _, page := range []string{"https://example.com/api/v2/", server.URL + "/private/", server.URL + "/api/v2/../private/", server.URL + "/api/v2/%2e%2e/private/", server.URL + "/api/v2/foo%2f..%2f..%2fprivate/"} {
		if _, err := client.ListLocations(context.Background(), &page); err == nil {
			t.Fatalf("accepted %s", page)
		}
	}
	if hits != 0 {
		t.Fatal("out-of-bound URL requested")
	}
	if _, err := client.ListLocations(context.Background(), nil); err == nil {
		t.Fatal("redirect accepted")
	}
}
func TestCancellationAndTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	client, _ := NewClient(server.URL+"/", 20*time.Millisecond)
	if _, err := client.ListLocations(context.Background(), nil); err == nil {
		t.Fatal("timeout missing")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.ListLocations(ctx, nil); err == nil {
		t.Fatal("cancellation missing")
	}
}
func TestInvalidConfiguration(t *testing.T) {
	for _, base := range []string{"", "://", "ftp://example.com/", "https://example.com", "https://user@example.com/", "https://example.com/?key=x"} {
		if _, err := NewClient(base, time.Second); err == nil {
			t.Fatalf("accepted %q", base)
		}
	}
	if _, err := NewClient(BaseURL, 0); err == nil {
		t.Fatal("zero timeout accepted")
	}
}

func TestInvalidIdentifiersDoNotReachNetwork(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits++; fmt.Fprint(w, `{}`) }))
	defer server.Close()
	client, _ := NewClient(server.URL+"/", time.Second)
	for _, name := range []string{"", "../secret", "pikachu\x1b[31m", "-foo", strings.Repeat("x", 101)} {
		if _, err := client.Explore(context.Background(), name); err == nil {
			t.Fatalf("area accepted %q", name)
		}
		if _, err := client.GetPokemon(context.Background(), name); err == nil {
			t.Fatalf("Pokemon accepted %q", name)
		}
	}
	if hits != 0 {
		t.Fatal("invalid identifier reached network")
	}
}
func TestSemanticallyInvalidResponsesNotCached(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		fmt.Fprint(w, `{"name":"pidgey","types":[{"type":{"name":"\u001b[31m"}}]}`)
	}))
	defer server.Close()
	client, _ := NewClient(server.URL+"/", time.Second)
	cache := pokecache.NewCache(time.Hour)
	defer cache.Close()
	client.SetCache(cache)
	for i := 0; i < 2; i++ {
		if _, err := client.GetPokemon(context.Background(), "pidgey"); err == nil {
			t.Fatal("terminal control accepted")
		}
	}
	if hits.Load() != 2 {
		t.Fatal("invalid response cached")
	}
}
