package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/yaqyn/pokedex/internal/pokeapi"
	"github.com/yaqyn/pokedex/internal/pokecache"
	"golang.org/x/term"
)

var version = "dev"

func main() {
	if err := run(); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	interactive := term.IsTerminal(int(os.Stdin.Fd()))
	return runCLI(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr, interactive)
}
func runCLI(ctx context.Context, args []string, in io.Reader, out, errOut io.Writer, interactive bool) error {
	flags := flag.NewFlagSet("pokedex", flag.ContinueOnError)
	flags.SetOutput(errOut)
	showVersion := flags.Bool("version", false, "Print the build version")
	dataPath := flags.String("data", "", "JSON save path (default: user configuration directory/pokedex/pokedex.json)")
	ephemeral := flags.Bool("ephemeral", false, "Use a temporary collection without loading or writing saves")
	timeout := flags.Duration("timeout", 10*time.Second, "Maximum duration of each API request")
	flags.Usage = func() {
		fmt.Fprintln(errOut, "Usage: pokedex [flags] [command [argument]]\n\nStart an interactive session or run one command. Use 'pokedex help' for commands.\n\nFlags:")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); errors.Is(err, flag.ErrHelp) {
		return nil
	} else if err != nil {
		return err
	}
	if *showVersion {
		fmt.Fprintln(out, "pokedex", version)
		return nil
	}
	if *ephemeral && *dataPath != "" {
		return fmt.Errorf("-data and -ephemeral cannot be combined")
	}
	words := flags.Args()
	for i := range words {
		words[i] = strings.ToLower(words[i])
	}
	cfg := &config{commands: getCommands(), ctx: ctx, out: out, errOut: errOut, prompt: interactive, batch: !interactive, pokedex: make(map[string]pokeapi.Pokemon), random: rand.Float64}
	// Help and malformed one-shot commands should not create or lock save files.
	if len(words) > 0 {
		cmd, ok := cfg.commands[words[0]]
		if !ok || len(words[1:]) != cmd.args || len(words) > 1 && !validName(words[1]) || words[0] == "help" || words[0] == "exit" {
			err := execute(cfg, words)
			if errors.Is(err, errExit) {
				return nil
			}
			return err
		}
	}
	client, err := pokeapi.NewClient(pokeapi.BaseURL, *timeout)
	if err != nil {
		return err
	}
	cache := pokecache.NewCache(5 * time.Minute)
	defer cache.Close()
	client.SetCache(cache)
	cfg.client = client
	if !*ephemeral {
		path := *dataPath
		if path == "" {
			dir, err := os.UserConfigDir()
			if err != nil {
				return fmt.Errorf("find save directory: %w", err)
			}
			path = filepath.Join(dir, "pokedex", "pokedex.json")
		}
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return fmt.Errorf("create save directory: %w", err)
		}
		if resolved, err := filepath.EvalSymlinks(path); err == nil {
			path = resolved
		} else if !os.IsNotExist(err) {
			return err
		}
		release, err := lockSave(path)
		if err != nil {
			return err
		}
		defer release()
		cfg.pokedex, err = loadPokedex(path)
		if err != nil {
			return err
		}
		cfg.save = func(p map[string]pokeapi.Pokemon) error { return savePokedex(path, p) }
	}
	if len(words) > 0 {
		err := execute(cfg, words)
		if errors.Is(err, errExit) {
			return nil
		}
		return err
	}
	return runREPL(cfg, in)
}
