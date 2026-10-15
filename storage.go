package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/yaqyn/clibble-terminal-rpg/internal/pokeapi"
)

const maxSaveBytes = 4 << 20

type diskPokedex struct {
	Version  int                        `json:"version"`
	Pokemon  map[string]pokeapi.Pokemon `json:"pokemon"`
	Progress *gameProgress              `json:"progress,omitempty"`
}

func loadPokedex(path string) (map[string]pokeapi.Pokemon, error) {
	p, _, err := loadGame(path)
	return p, err
}
func loadGame(path string) (map[string]pokeapi.Pokemon, gameProgress, error) {
	pokemon := make(map[string]pokeapi.Pokemon)
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return pokemon, freshProgress(time.Now()), nil
	}
	if err != nil {
		return nil, gameProgress{}, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxSaveBytes+1))
	if err != nil {
		return nil, gameProgress{}, err
	}
	if len(data) > maxSaveBytes {
		return nil, gameProgress{}, fmt.Errorf("save exceeds 4 MiB")
	}
	var saved diskPokedex
	if err := json.Unmarshal(data, &saved); err != nil {
		return nil, gameProgress{}, fmt.Errorf("invalid save: %w", err)
	}
	if (saved.Version != 1 && saved.Version != 2) || saved.Pokemon == nil {
		return nil, gameProgress{}, fmt.Errorf("unsupported or incomplete save")
	}
	for name, p := range saved.Pokemon {
		if name != p.Name {
			return nil, gameProgress{}, fmt.Errorf("save key does not match wildling name")
		}
		if err := pokeapi.ValidatePokemon(p); err != nil {
			return nil, gameProgress{}, fmt.Errorf("invalid saved wildling: %w", err)
		}
	}
	progress := freshProgress(time.Now())
	if saved.Version == 2 {
		if saved.Progress == nil {
			return nil, gameProgress{}, fmt.Errorf("missing game progress")
		}
		progress = *saved.Progress
		if progress.Seeds < 0 || progress.Seeds > seedCapacity || progress.RefillAt.IsZero() || progress.Encounters == nil {
			return nil, gameProgress{}, fmt.Errorf("invalid game progress")
		}
		for name, e := range progress.Encounters {
			if name == "" || len(name) > 100 || e.Offers < 1 || e.Offers > 3 || e.Until.IsZero() {
				return nil, gameProgress{}, fmt.Errorf("invalid encounter budget")
			}
		}
	}
	return saved.Pokemon, progress, nil
}
func savePokedex(path string, pokemon map[string]pokeapi.Pokemon) error {
	data, err := json.MarshalIndent(diskPokedex{Version: 1, Pokemon: pokemon}, "", "  ")
	if err != nil {
		return err
	}
	return writeSave(path, data)
}
func saveGame(path string, pokemon map[string]pokeapi.Pokemon, progress gameProgress) error {
	data, err := json.MarshalIndent(diskPokedex{Version: 2, Pokemon: pokemon, Progress: &progress}, "", "  ")
	if err != nil {
		return err
	}
	return writeSave(path, data)
}
func writeSave(path string, data []byte) error {
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
