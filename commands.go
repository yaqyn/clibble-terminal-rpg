package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/yaqyn/clibble-terminal-rpg/internal/pokeapi"
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
		"explore":    {"explore <area>", "Meet wildlings on a trail", 1, commandExplore},
		"catch":      {"catch <wildling>", "Attempt to catch a wildling", 1, commandCatch},
		"inspect":    {"inspect <wildling>", "Show a caught wildling's details", 1, commandInspect},
		"collection": {"collection", "List your caught wildlings", 0, commandPokedex},
		"pokedex":    {"pokedex", "Legacy alias for collection", 0, commandPokedex},
		"count":      {"count", "Show the number of caught wildling", 0, commandCount},
		"types":      {"types", "Summarize your collection by wildling type", 0, commandTypes},
		"type":       {"type <type>", "List caught wildlings with a type", 1, commandType},
		"largest":    {"largest", "Show your heaviest caught wildling", 0, commandLargest},
		"tallest":    {"tallest", "Show your tallest caught wildling", 0, commandTallest},
		"search":     {"search <text>", "Find caught wildlings by name", 1, commandSearch},
		"supplies":   {"supplies", "Show available catch seeds without spending any", 0, commandSupplies},
		"where":      {"where <wildling>", "List bundled trails where a wildling lives", 1, commandWhere},
		"trails":     {"trails", "List bundled trails and their wildling counts", 0, commandTrails},
		"trail":      {"trail <trail>", "List wildlings on a bundled trail", 1, commandTrail},
		"progress":   {"progress", "Show progress through the bundled wildlings", 0, commandProgress},
		"missing":    {"missing <trail>", "List uncaught wildlings on a bundled trail", 1, commandMissing},
		"strongest":  {"strongest", "Show the caught wildling with the highest total stats", 0, commandStrongest},
		"coverage":   {"coverage", "Show collection progress for each bundled trail", 0, commandCoverage},
		"surprise":   {"surprise", "Choose a caught wildling for your next adventure", 0, commandSurprise},
	}
}

func commandSurprise(cfg *config, args ...string) error {
	names := sortedKeys(cfg.pokedex)
	if len(names) == 0 {
		_, err := fmt.Fprintln(cfg.out, "No wildling caught yet. Try catch <wildling>.")
		return err
	}
	choice := 0
	if cfg.random != nil {
		choice = min(int(cfg.random()*float64(len(names))), len(names)-1)
	}
	_, err := fmt.Fprintf(cfg.out, "Adventure companion: %s\n", names[choice])
	return err
}

func commandCoverage(cfg *config, args ...string) error {
	for _, trail := range trails {
		caught := 0
		for _, name := range trail.creatures {
			if _, ok := cfg.pokedex[name]; ok {
				caught++
			}
		}
		if _, err := fmt.Fprintf(cfg.out, "%s: %d/%d caught\n", trail.name, caught, len(trail.creatures)); err != nil {
			return err
		}
	}
	return nil
}

func commandStrongest(cfg *config, args ...string) error {
	if len(cfg.pokedex) == 0 {
		_, err := fmt.Fprintln(cfg.out, "No wildling caught yet. Try catch <wildling>.")
		return err
	}
	var strongest pokeapi.Pokemon
	best := -1
	for _, pokemon := range cfg.pokedex {
		total := 0
		for _, stat := range pokemon.Stats {
			total += stat.BaseStat
		}
		if total > best || total == best && pokemon.Name < strongest.Name {
			strongest, best = pokemon, total
		}
	}
	_, err := fmt.Fprintf(cfg.out, "Strongest wildling: %s (%d total base stats)\n", strongest.Name, best)
	return err
}

func commandMissing(cfg *config, args ...string) error {
	for _, trail := range trails {
		if !strings.EqualFold(trail.name, args[0]) {
			continue
		}
		missing := make([]string, 0, len(trail.creatures))
		for _, name := range trail.creatures {
			if _, caught := cfg.pokedex[name]; !caught {
				missing = append(missing, name)
			}
		}
		if len(missing) == 0 {
			_, err := fmt.Fprintf(cfg.out, "%s complete: every wildling is in your collection.\n", trail.name)
			return err
		}
		_, err := fmt.Fprintf(cfg.out, "Missing from %s: %s\n", trail.name, strings.Join(missing, ", "))
		return err
	}
	return fmt.Errorf("unknown trail %q; use trails to list bundled trails", args[0])
}

func commandProgress(cfg *config, args ...string) error {
	available := map[string]bool{}
	for _, trail := range trails {
		for _, name := range trail.creatures {
			available[name] = true
		}
	}
	caught := 0
	for name := range cfg.pokedex {
		if available[name] {
			caught++
		}
	}
	percent := 0
	if len(available) > 0 {
		percent = caught * 100 / len(available)
	}
	_, err := fmt.Fprintf(cfg.out, "Bundled wildling progress: %d/%d (%d%%)\n", caught, len(available), percent)
	return err
}

func commandTrail(cfg *config, args ...string) error {
	for _, trail := range trails {
		if strings.EqualFold(trail.name, args[0]) {
			_, err := fmt.Fprintf(cfg.out, "%s wildlings: %s\n", trail.name, strings.Join(trail.creatures, ", "))
			return err
		}
	}
	return fmt.Errorf("unknown trail %q; use trails to list bundled trails", args[0])
}

func commandTallest(cfg *config, args ...string) error {
	if len(cfg.pokedex) == 0 {
		_, err := fmt.Fprintln(cfg.out, "No wildling caught yet. Try catch <wildling>.")
		return err
	}
	var tallest pokeapi.Pokemon
	for _, pokemon := range cfg.pokedex {
		if pokemon.Height > tallest.Height || pokemon.Height == tallest.Height && pokemon.Name < tallest.Name {
			tallest = pokemon
		}
	}
	_, err := fmt.Fprintf(cfg.out, "Tallest wildling: %s (%d dm)\n", tallest.Name, tallest.Height)
	return err
}

func commandSearch(cfg *config, args ...string) error {
	query := strings.ToLower(args[0])
	var names []string
	for name := range cfg.pokedex {
		if strings.Contains(strings.ToLower(name), query) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		_, err := fmt.Fprintf(cfg.out, "No caught wildlings match %q.\n", args[0])
		return err
	}
	_, err := fmt.Fprintln(cfg.out, strings.Join(names, ", "))
	return err
}

func commandLargest(cfg *config, args ...string) error {
	if len(cfg.pokedex) == 0 {
		_, err := fmt.Fprintln(cfg.out, "No wildling caught yet. Try catch <wildling>.")
		return err
	}
	var largest pokeapi.Pokemon
	for _, pokemon := range cfg.pokedex {
		if pokemon.Weight > largest.Weight || pokemon.Weight == largest.Weight && pokemon.Name < largest.Name {
			largest = pokemon
		}
	}
	_, err := fmt.Fprintf(cfg.out, "Heaviest wildling: %s (%d hg)\n", largest.Name, largest.Weight)
	return err
}

func commandType(cfg *config, args ...string) error {
	var names []string
	for name, pokemon := range cfg.pokedex {
		for _, kind := range pokemon.Types {
			if kind.Type.Name == args[0] {
				names = append(names, name)
				break
			}
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		_, err := fmt.Fprintf(cfg.out, "No caught %s wildlings.\n", args[0])
		return err
	}
	_, err := fmt.Fprintf(cfg.out, "%s wildlings: %s\n", args[0], strings.Join(names, ", "))
	return err
}

func commandTrails(cfg *config, args ...string) error {
	if _, err := fmt.Fprintln(cfg.out, "Bundled trails:"); err != nil {
		return err
	}
	for _, trail := range trails {
		if _, err := fmt.Fprintf(cfg.out, " - %s: %d wildlings\n", trail.name, len(trail.creatures)); err != nil {
			return err
		}
	}
	return nil
}

func commandWhere(cfg *config, args ...string) error {
	var locations []string
	for _, trail := range trails {
		for _, name := range trail.creatures {
			if name == args[0] {
				locations = append(locations, trail.name)
				break
			}
		}
	}
	if len(locations) == 0 {
		return fmt.Errorf("unknown wildling %q; explore a trail to meet one", args[0])
	}
	for _, location := range locations {
		if _, err := fmt.Fprintln(cfg.out, location); err != nil {
			return err
		}
	}
	return nil
}

func commandSupplies(cfg *config, args ...string) error {
	now := cfg.now()
	progress := cfg.progress.refreshed(now)
	if _, err := fmt.Fprintf(cfg.out, "Catch seeds: %d/%d\n", progress.Seeds, seedCapacity); err != nil {
		return err
	}
	if progress.Seeds < seedCapacity {
		_, err := fmt.Fprintf(cfg.out, "Next seed in: %s\n", progress.RefillAt.Add(seedInterval).Sub(now).Round(time.Second))
		return err
	}
	return nil
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
	fmt.Fprintln(cfg.out, "Wildlings on this trail:")
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
	fmt.Fprintf(cfg.out, "Offering a catch seed to %s...\n", args[0])
	result, err := attemptCatch(cfg.ctx, cfg.client, cfg.pokedex, cfg.random, cfg.persist, cfg.progress, cfg.now(), args[0])
	if err != nil {
		return err
	}
	if result.already {
		fmt.Fprintf(cfg.out, "%s is already in your collection.\n", result.pokemon.Name)
		return nil
	}
	cfg.progress = result.progress
	fmt.Fprintf(cfg.out, "Catch seeds: %d/%d; tries left: %d/3\n", cfg.progress.Seeds, seedCapacity, 3-cfg.progress.Encounters[result.pokemon.Name].Offers)
	if !result.caught {
		fmt.Fprintf(cfg.out, "%s escaped!\n", result.pokemon.Name)
		return nil
	}
	cfg.pokedex = result.owned
	fmt.Fprintf(cfg.out, "%s was caught!\nOpen its record with inspect <wildling>.\n", result.pokemon.Name)
	return nil
}

func commandInspect(cfg *config, args ...string) error {
	pokemon, ok := cfg.pokedex[args[0]]
	if !ok {
		return fmt.Errorf("you have not caught that wildling; try catch %s first", args[0])
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
		fmt.Fprintln(cfg.out, "No wildling caught yet. Try catch <wildling>.")
	}
	return nil
}
func validName(name string) bool { return pokeapi.ValidName(name) }

func commandCount(cfg *config, args ...string) error {
	_, err := fmt.Fprintf(cfg.out, "Befriended wildlings: %d\n", len(cfg.pokedex))
	return err
}

func commandTypes(cfg *config, args ...string) error {
	counts := map[string]int{}
	for _, pokemon := range cfg.pokedex {
		for _, kind := range pokemon.Types {
			counts[kind.Type.Name]++
		}
	}
	if len(counts) == 0 {
		_, err := fmt.Fprintln(cfg.out, "No caught wildling types yet. Try catch <wildling>.")
		return err
	}
	if _, err := fmt.Fprintln(cfg.out, "Collection types:"); err != nil {
		return err
	}
	for _, name := range sortedKeys(counts) {
		if _, err := fmt.Fprintf(cfg.out, " - %s: %d\n", name, counts[name]); err != nil {
			return err
		}
	}
	return nil
}
