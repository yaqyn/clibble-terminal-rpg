package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"time"

	"github.com/yaqyn/pokedex/internal/pokeapi"
	"github.com/yaqyn/pokedex/internal/pokecache"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	client, err := pokeapi.NewClient(pokeapi.BaseURL, 10*time.Second)
	if err != nil {
		return err
	}
	cache := pokecache.NewCache(5 * time.Minute)
	defer cache.Close()
	client.SetCache(cache)
	cfg := &config{commands: getCommands(), client: client, ctx: context.Background(), out: os.Stdout, pokedex: make(map[string]pokeapi.Pokemon), random: rand.Float64}
	return runREPL(cfg, os.Stdin)
}
