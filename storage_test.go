package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/yaqyn/dexrove/internal/pokeapi"
)

func TestSaveRoundTripAndPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pokedex.json")
	initial, err := loadPokedex(path)
	if err != nil || len(initial) != 0 {
		t.Fatalf("missing save: %v", err)
	}
	pokemon := map[string]pokeapi.Pokemon{"pidgey": {Name: "pidgey", Height: 3, Weight: 18}}
	if err := savePokedex(path, pokemon); err != nil {
		t.Fatal(err)
	}
	loaded, err := loadPokedex(path)
	if err != nil || loaded["pidgey"].Height != 3 {
		t.Fatalf("round trip: %v %v", loaded, err)
	}
	info, _ := os.Stat(path)
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatalf("mode %v", info.Mode())
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatal("temporary file left behind")
	}
}
func TestCorruptSaveRejectedWithoutOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pokedex.json")
	for _, content := range []string{`{`, `{"version":2,"pokemon":{}}`, `{"version":1}`, `{"version":1,"pokemon":{"pidgey":{"name":"pikachu"}}}`, `{"version":1,"pokemon":{"pidgey":{"name":"pidgey","stats":[{"stat":{"name":"\u001b[31m"}}]}}}`} {
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadPokedex(path); err == nil {
			t.Fatalf("accepted %s", content)
		}
		data, _ := os.ReadFile(path)
		if string(data) != content {
			t.Fatal("corrupt save modified")
		}
	}
}
func TestSaveFailurePreservesPreviousFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pokedex.json")
	original := map[string]pokeapi.Pokemon{"pidgey": {Name: "pidgey"}}
	if err := savePokedex(path, original); err != nil {
		t.Fatal(err)
	}
	huge := map[string]pokeapi.Pokemon{strings.Repeat("x", maxSaveBytes): {Name: "pidgey"}}
	if err := savePokedex(path, huge); err == nil {
		t.Fatal("oversize save accepted")
	}
	got, err := loadPokedex(path)
	if err != nil || len(got) != 1 {
		t.Fatalf("original lost: %v", err)
	}
}
func TestExclusiveSaveLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "save.json")
	release, err := lockSave(path)
	if err != nil {
		t.Fatal(err)
	}
	if second, err := lockSave(path); err == nil {
		second()
		release()
		t.Fatal("second writer allowed")
	}
	release()
	next, err := lockSave(path)
	if err != nil {
		t.Fatal(err)
	}
	next()
}

func TestSaveReplacesExistingCollection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pokedex.json")
	for _, pokemon := range []pokeapi.Pokemon{{Name: "pidgey", Height: 3}, {Name: "pikachu", Height: 4}} {
		if err := savePokedex(path, map[string]pokeapi.Pokemon{pokemon.Name: pokemon}); err != nil {
			t.Fatal(err)
		}
		loaded, err := loadPokedex(path)
		if err != nil || len(loaded) != 1 || loaded[pokemon.Name].Height != pokemon.Height {
			t.Fatalf("replacement failed: %v %v", loaded, err)
		}
	}
}
