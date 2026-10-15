package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/yaqyn/clibble-pokemon-cli/internal/pokeapi"
)

var errExit = errors.New("exit requested")

type cliCommand struct {
	name        string
	description string
	args        int
	callback    func(*config, ...string) error
}

type pokemonAPI interface {
	ListLocations(context.Context, *string) (pokeapi.Locations, error)
	Explore(context.Context, string) (pokeapi.Area, error)
	GetPokemon(context.Context, string) (pokeapi.Pokemon, error)
}
type config struct {
	progress     gameProgress
	clock        func() time.Time
	saveProgress func(map[string]pokeapi.Pokemon, gameProgress) error
	noMotion     bool

	readLine func() (string, error)
	view     *terminalView
	errOut   io.Writer
	prompt   bool
	batch    bool
	commands map[string]cliCommand
	client   pokemonAPI
	ctx      context.Context
	out      io.Writer
	next     *string
	previous *string
	started  bool
	pokedex  map[string]pokeapi.Pokemon
	random   func() float64
	save     func(map[string]pokeapi.Pokemon) error
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help":       {"help", "Displays a help message", 0, commandHelp},
		"exit":       {"exit", "Leave Clibble", 0, commandExit},
		"map":        {"map", "Display the next 20 location areas", 0, commandMap},
		"mapb":       {"mapb", "Display the previous 20 location areas", 0, commandMapBack},
		"explore":    {"explore <area>", "List Pokemon in a location area", 1, commandExplore},
		"catch":      {"catch <pokemon>", "Attempt to catch a Pokemon", 1, commandCatch},
		"inspect":    {"inspect <pokemon>", "Show a caught Pokemon's details", 1, commandInspect},
		"collection": {"collection", "List your caught Pokemon", 0, commandPokedex},
		"pokedex":    {"pokedex", "Alias for collection", 0, commandPokedex},
		"count":      {"count", "Show the number of caught Pokemon", 0, commandCount},
	}
}
func commandExit(cfg *config, args ...string) error {
	fmt.Fprintln(cfg.out, "Leaving Clibble. See you on the next trail.")
	return errExit
}
func commandHelp(cfg *config, args ...string) error {
	if cfg.view != nil {
		cfg.view.help()
		return nil
	}
	fmt.Fprint(cfg.out, "Clibble — A little wild in your terminal.\nUsage:\n\n")
	for _, key := range sortedKeys(cfg.commands) {
		cmd := cfg.commands[key]
		fmt.Fprintf(cfg.out, "%s: %s\n", cmd.name, cmd.description)
	}
	fmt.Fprintln(cfg.out, "\nserve: Open the local collection companion (clibble serve)")
	return nil
}
func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
func commandMap(cfg *config, args ...string) error {
	if cfg.started && cfg.next == nil {
		fmt.Fprintln(cfg.out, "you're on the last page")
		return nil
	}
	return showLocations(cfg, cfg.next)
}
func commandMapBack(cfg *config, args ...string) error {
	if cfg.previous == nil {
		fmt.Fprintln(cfg.out, "you're on the first page")
		return nil
	}
	return showLocations(cfg, cfg.previous)
}
func showLocations(cfg *config, page *string) error {
	result, err := cfg.client.ListLocations(cfg.ctx, page)
	if err != nil {
		return err
	}
	if cfg.view != nil {
		cfg.view.locations(result.Results)
	} else {
		for _, area := range result.Results {
			fmt.Fprintln(cfg.out, area.Name)
		}
	}
	cfg.next, cfg.previous, cfg.started = result.Next, result.Previous, true
	return nil
}
func commandExplore(cfg *config, args ...string) error {
	fmt.Fprintf(cfg.out, "Exploring %s...\n", args[0])
	area, err := cfg.client.Explore(cfg.ctx, args[0])
	if err != nil {
		return err
	}
	fmt.Fprintln(cfg.out, "Found Pokemon:")
	for _, encounter := range area.PokemonEncounters {
		fmt.Fprintf(cfg.out, " - %s\n", encounter.Pokemon.Name)
	}
	return nil
}
func catchProbability(experience int) float64 {
	if experience < 0 {
		experience = 0
	}
	return 100.0 / (100.0 + float64(experience))
}
func commandCatch(cfg *config, args ...string) error {
	if _, ok := cfg.pokedex[args[0]]; ok {
		fmt.Fprintf(cfg.out, "%s is already in your collection.\n", args[0])
		return nil
	}
	fmt.Fprintf(cfg.out, "Throwing a Pokeball at %s...\n", args[0])
	result, err := attemptCatch(cfg.ctx, cfg.client, cfg.pokedex, cfg.random, cfg.persist, cfg.progress, cfg.now(), args[0])
	if err != nil {
		return err
	}
	if result.already {
		fmt.Fprintf(cfg.out, "%s is already in your collection.\n", result.pokemon.Name)
		return nil
	}
	cfg.progress = result.progress
	fmt.Fprintf(cfg.out, "Pokeballs: %d/%d; tries left: %d/3\n", cfg.progress.Balls, ballCapacity, 3-cfg.progress.Encounters[result.pokemon.Name].Throws)
	if !result.caught {
		fmt.Fprintf(cfg.out, "%s escaped!\n", result.pokemon.Name)
		return nil
	}
	cfg.pokedex = result.owned
	fmt.Fprintf(cfg.out, "%s was caught!\nOpen its record with inspect <pokemon>.\n", result.pokemon.Name)
	return nil
}
func commandInspect(cfg *config, args ...string) error {
	pokemon, ok := cfg.pokedex[args[0]]
	if !ok {
		return fmt.Errorf("you have not caught that pokemon; try catch %s first", args[0])
	}
	if cfg.view != nil {
		cfg.view.inspect(pokemon)
		return nil
	}
	fmt.Fprintf(cfg.out, "Name: %s\nHeight: %d\nWeight: %d\nStats:\n", pokemon.Name, pokemon.Height, pokemon.Weight)
	for _, stat := range pokemon.Stats {
		fmt.Fprintf(cfg.out, "  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Fprintln(cfg.out, "Types:")
	for _, kind := range pokemon.Types {
		fmt.Fprintf(cfg.out, "  - %s\n", kind.Type.Name)
	}
	return nil
}
func commandPokedex(cfg *config, args ...string) error {
	if cfg.view != nil {
		cfg.view.collection(cfg.pokedex)
		return nil
	}
	fmt.Fprintln(cfg.out, "Your collection:")
	for _, name := range sortedKeys(cfg.pokedex) {
		fmt.Fprintf(cfg.out, " - %s\n", name)
	}
	if len(cfg.pokedex) == 0 {
		fmt.Fprintln(cfg.out, "No Pokemon caught yet. Try catch <pokemon>.")
	}
	return nil
}
func validName(name string) bool { return pokeapi.ValidName(name) }

func commandCount(cfg *config, args ...string) error {
	_, err := fmt.Fprintf(cfg.out, "Caught Pokemon: %d\n", len(cfg.pokedex))
	return err
}
