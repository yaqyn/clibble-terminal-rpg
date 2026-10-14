package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/yaqyn/pokedex/internal/pokeapi"
	"github.com/yaqyn/pokedex/internal/pokecache"
)

func main() {
	if err := run(); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	dataPath := flag.String("data", "", "Optional JSON save file (one writer at a time)")
	timeout := flag.Duration("timeout", 10*time.Second, "Maximum duration of each API request")
	flag.Parse()
	if flag.NArg() != 0 {
		return fmt.Errorf("unexpected arguments; use -h for options")
	}
	client, err := pokeapi.NewClient(pokeapi.BaseURL, *timeout)
	if err != nil {
		return err
	}
	cache := pokecache.NewCache(5 * time.Minute)
	defer cache.Close()
	client.SetCache(cache)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg := &config{commands: getCommands(), client: client, ctx: ctx, out: os.Stdout, pokedex: make(map[string]pokeapi.Pokemon), random: rand.Float64}
	if *dataPath != "" {
		path, err := filepath.Abs(*dataPath)
		if err != nil {
			return err
		}
		// Resolve an existing symlink so saves and locks use the same target.
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
	// Closing stdin unblocks Scan when interrupted at an idle prompt.
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			_ = os.Stdin.Close()
		case <-done:
		}
	}()
	return runREPL(cfg, os.Stdin)
}
