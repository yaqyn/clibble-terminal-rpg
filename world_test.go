package main

import (
	"context"
	"github.com/yaqyn/clibble-terminal-rpg/internal/pokeapi"
	"testing"
)

func TestOriginalWorldIsCompleteAndSaveable(t *testing.T) {
	w := meadowWorld{}
	locations, err := w.ListLocations(context.Background(), nil)
	if err != nil || len(locations.Results) != 6 {
		t.Fatal(locations, err)
	}
	seen := map[string]bool{}
	for _, location := range locations.Results {
		area, err := w.Explore(context.Background(), location.Name)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range area.PokemonEncounters {
			p, err := w.GetPokemon(context.Background(), e.Pokemon.Name)
			if err != nil {
				t.Fatal(err)
			}
			if err := pokeapi.ValidatePokemon(p); err != nil {
				t.Fatal(p.Name, err)
			}
			seen[p.Name] = true
		}
	}
	if len(seen) != len(wildlings) {
		t.Fatal("unreachable wildling")
	}
	if _, err := w.GetPokemon(context.Background(), "pikachu"); err == nil {
		t.Fatal("legacy species remains in world")
	}
}
