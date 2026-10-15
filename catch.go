package main

import (
	"context"
	"fmt"
	"time"

	"github.com/yaqyn/clibble-pokemon-cli/internal/pokeapi"
)

type catchResult struct {
	progress gameProgress
	pokemon  pokeapi.Pokemon
	owned    map[string]pokeapi.Pokemon
	caught   bool
	already  bool
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
func attemptCatch(ctx context.Context, client pokemonAPI, owned map[string]pokeapi.Pokemon, random func() float64, save func(map[string]pokeapi.Pokemon, gameProgress) error, progress gameProgress, now time.Time, name string) (catchResult, error) {
	if p, ok := owned[name]; ok {
		return catchResult{pokemon: p, already: true}, nil
	}
	if err := progress.allowance(name, now); err != nil {
		return catchResult{}, err
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
	if err := progress.allowance(p.Name, now); err != nil {
		return catchResult{}, err
	}
	budget := progress.refreshed(now)
	budget.Seeds--
	encounter := budget.Encounters[p.Name]
	if encounter.Until.IsZero() {
		encounter.Until = now.Add(encounterWindow)
	}
	encounter.Offers++
	budget.Encounters[p.Name] = encounter
	next := copyCollection(owned)
	caught := random() < catchProbability(p.BaseExperience)
	if caught {
		next[p.Name] = p
	}
	if save != nil {
		if err := save(next, budget); err != nil {
			return catchResult{}, fmt.Errorf("save catch: %w", err)
		}
	}
	return catchResult{pokemon: p, owned: next, progress: budget, caught: caught}, nil
}
