package main

import (
	"context"
	"fmt"

	"github.com/yaqyn/clibble-pokemon-cli/internal/pokeapi"
)

type catchResult struct {
	pokemon pokeapi.Pokemon
	owned   map[string]pokeapi.Pokemon
	caught  bool
	already bool
}

func copyCollection(source map[string]pokeapi.Pokemon) map[string]pokeapi.Pokemon {
	result := make(map[string]pokeapi.Pokemon, len(source)+1)
	for name, p := range source {
		result[name] = p
	}
	return result
}

// attemptCatch works on a snapshot. Neither interface publishes a new catch
// until its save succeeds, and background tasks never mutate the live collection.
func attemptCatch(ctx context.Context, client pokemonAPI, owned map[string]pokeapi.Pokemon, random func() float64, save func(map[string]pokeapi.Pokemon) error, name string) (catchResult, error) {
	if p, ok := owned[name]; ok {
		return catchResult{pokemon: p, already: true}, nil
	}
	p, err := client.GetPokemon(ctx, name)
	if err != nil {
		return catchResult{}, err
	}
	if _, ok := owned[p.Name]; ok {
		return catchResult{pokemon: p, already: true}, nil
	}
	if err := ctx.Err(); err != nil {
		return catchResult{}, err
	}
	if random() >= catchProbability(p.BaseExperience) {
		return catchResult{pokemon: p}, nil
	}
	next := copyCollection(owned)
	next[p.Name] = p
	if save != nil {
		if err := save(next); err != nil {
			return catchResult{}, fmt.Errorf("save catch: %w", err)
		}
	}
	return catchResult{pokemon: p, owned: next, caught: true}, nil
}
