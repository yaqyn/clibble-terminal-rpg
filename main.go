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

	"github.com/yaqyn/clibble-pokemon-cli/internal/brand"
	"github.com/yaqyn/clibble-pokemon-cli/internal/pokeapi"
	"github.com/yaqyn/clibble-pokemon-cli/internal/pokecache"
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
	interactive := term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
	return runCLI(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr, interactive)
}
func runCLI(ctx context.Context, args []string, in io.Reader, out, errOut io.Writer, interactive bool) error {
	flags := flag.NewFlagSet(brand.Get().Slug, flag.ContinueOnError)
	flags.SetOutput(errOut)
	plain := flags.Bool("plain", false, "Use the command prompt instead of the RPG screen")
	port := flags.Int("port", 8787, "Local web companion port (serve only)")
	colorMode := flags.String("color", "auto", "Terminal color: auto, always, or never")
	showVersion := flags.Bool("version", false, "Print the build version")
	dataPath := flags.String("data", "", "JSON save path (default: user configuration directory/clibble/collection.json)")
	ephemeral := flags.Bool("ephemeral", false, "Use a temporary collection without loading or writing saves")
	timeout := flags.Duration("timeout", 10*time.Second, "Maximum duration of each API request")
	flags.Usage = func() {
		fmt.Fprintln(errOut, "Usage: clibble [flags] [command [argument]]\n\nStart an interactive session or run one command. Use 'clibble help' for commands.\n\nFlags:")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); errors.Is(err, flag.ErrHelp) {
		return nil
	} else if err != nil {
		return err
	}
	if *colorMode != "auto" && *colorMode != "always" && *colorMode != "never" {
		return fmt.Errorf("-color must be auto, always, or never")
	}
	if *showVersion {
		fmt.Fprintln(out, brand.Get().Slug, version)
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
	if interactive {
		width := 80
		color := *colorMode == "always"
		if file, ok := out.(*os.File); ok {
			if w, _, err := term.GetSize(int(file.Fd())); err == nil && w > 0 {
				width = w
			}
			if *colorMode == "auto" {
				_, noColor := os.LookupEnv("NO_COLOR")
				color = term.IsTerminal(int(file.Fd())) && !noColor && os.Getenv("TERM") != "dumb"
			}
		}
		cfg.view = &terminalView{out: out, color: color, width: width}
	}
	if len(words) > 0 && words[0] == "serve" {
		if len(words) != 1 {
			return fmt.Errorf("Usage: clibble [flags] serve")
		}
		if *ephemeral {
			return fmt.Errorf("serve reads saved collections; omit -ephemeral")
		}
		path, err := collectionPath(*dataPath)
		if err != nil {
			return err
		}
		return serveCompanion(ctx, path, *port, out)
	}
	if _, noColor := os.LookupEnv("NO_COLOR"); noColor && cfg.view != nil {
		cfg.view.color = false
	}
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
		path, err := collectionPath(*dataPath)
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
	if interactive && !*plain && supportsGame(in, out) {
		return runGame(cfg, in, out)
	}
	if cfg.view != nil {
		restore, err := enableEditor(cfg, in, out)
		if err != nil {
			return err
		}
		defer restore()
	}
	if cfg.view != nil {
		cfg.view.welcome(len(cfg.pokedex), cfg.save != nil)
	}
	return runREPL(cfg, in)
}
