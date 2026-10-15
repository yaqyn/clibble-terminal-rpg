package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yaqyn/clibble-terminal-rpg/internal/brand"
)

// defaultCollection preserves the existing collection and its original lock.
func defaultCollection(dir string) (string, error) {
	path := filepath.Join(dir, brand.Get().Slug, "collection.json")
	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	for _, legacy := range []string{
		filepath.Join(dir, "dexrove", "collection.json"),
		filepath.Join(dir, "pokedex", "pokedex.json"),
	} {
		if _, err := os.Stat(legacy); err == nil {
			return legacy, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}
	return path, nil
}
func collectionPath(explicit string) (string, error) {
	path := explicit
	if path == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("find collection directory: %w", err)
		}
		path, err = defaultCollection(dir)
		if err != nil {
			return "", err
		}
	}
	return filepath.Abs(path)
}
