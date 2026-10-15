package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/yaqyn/clibble/internal/pokeapi"
)

const maxSaveBytes = 4 << 20

type diskPokedex struct {
	Version int                        `json:"version"`
	Pokemon map[string]pokeapi.Pokemon `json:"pokemon"`
}

func loadPokedex(path string) (map[string]pokeapi.Pokemon, error) {
	pokemon := make(map[string]pokeapi.Pokemon)
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return pokemon, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxSaveBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxSaveBytes {
		return nil, fmt.Errorf("save exceeds 4 MiB")
	}
	var saved diskPokedex
	if err := json.Unmarshal(data, &saved); err != nil {
		return nil, fmt.Errorf("invalid save: %w", err)
	}
	if saved.Version != 1 || saved.Pokemon == nil {
		return nil, fmt.Errorf("unsupported or incomplete save")
	}
	for name, p := range saved.Pokemon {
		if name != p.Name {
			return nil, fmt.Errorf("save key does not match Pokemon name")
		}
		if err := pokeapi.ValidatePokemon(p); err != nil {
			return nil, fmt.Errorf("invalid saved Pokemon: %w", err)
		}
	}
	return saved.Pokemon, nil
}
func savePokedex(path string, pokemon map[string]pokeapi.Pokemon) error {
	data, err := json.MarshalIndent(diskPokedex{Version: 1, Pokemon: pokemon}, "", "  ")
	if err != nil {
		return err
	}
	if len(data)+1 > maxSaveBytes {
		return fmt.Errorf("save exceeds 4 MiB")
	}
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, ".pokedex-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	defer temp.Close()
	if _, err := temp.Write(append(data, '\n')); err != nil {
		return err
	}
	if err := temp.Sync(); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return replaceSave(temp.Name(), path)
}
