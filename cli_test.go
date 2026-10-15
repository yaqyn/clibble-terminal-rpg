package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/yaqyn/clibble/internal/pokeapi"
)

func TestCLIHelpVersionAndInvalidCommandsHaveNoSaveSideEffects(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "pokedex.json")
	for _, tc := range []struct {
		args      []string
		wantError bool
		want      string
	}{
		{[]string{"-data", path, "help"}, false, "catch <pokemon>"},
		{[]string{"-data", path, "-version"}, false, "clibble dev"},
		{[]string{"-data", path, "what"}, true, ""},
		{[]string{"-data", path, "catch"}, true, ""},
		{[]string{"-data", path, "catch", "../secret"}, true, ""},
		{[]string{"-data", path, "-ephemeral", "pokedex"}, true, ""},
	} {
		var out, errOut bytes.Buffer
		err := runCLI(context.Background(), tc.args, strings.NewReader(""), &out, &errOut, false)
		if (err != nil) != tc.wantError {
			t.Fatalf("%v: %v", tc.args, err)
		}
		if !strings.Contains(out.String(), tc.want) {
			t.Errorf("%v: output %s", tc.args, out.String())
		}
	}
	if _, err := os.Stat(filepath.Dir(path)); !os.IsNotExist(err) {
		t.Fatalf("help or bad args created directory: %v", err)
	}
}
func TestCLIRestoresCollectionForSingleCommand(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pokedex.json")
	if err := savePokedex(path, map[string]pokeapi.Pokemon{"pidgey": {Name: "pidgey", Height: 3, Weight: 18}}); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if err := runCLI(context.Background(), []string{"-data", path, "inspect", "PIDGEY"}, strings.NewReader(""), &out, &errOut, false); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Name: pidgey") || strings.Contains(out.String(), "clibble >") {
		t.Fatal(out.String())
	}
	if err := runCLI(context.Background(), []string{"-data", path, "inspect", "pikachu"}, strings.NewReader(""), &out, &errOut, false); err == nil {
		t.Fatal("uncaught inspection succeeded")
	}
}
func TestCLIEphemeralAndAutomaticDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "save.json")
	var out, errOut bytes.Buffer
	if err := runCLI(context.Background(), []string{"-data", path, "pokedex"}, strings.NewReader(""), &out, &errOut, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".lock"); err != nil {
		t.Fatal(err)
	}
	if err := runCLI(context.Background(), []string{"-ephemeral", "pokedex"}, strings.NewReader(""), &out, &errOut, false); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "No Pokemon caught yet") {
		t.Fatal(out.String())
	}
}
func TestCLIDefaultSaveDirectory(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "windows" {
		t.Skip("platform config location cannot be isolated without changing HOME")
	}
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("AppData", dir)
	var out, errOut bytes.Buffer
	if err := runCLI(context.Background(), []string{"pokedex"}, strings.NewReader(""), &out, &errOut, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "clibble", "collection.json.lock")); err != nil {
		t.Fatal(err)
	}
}
func TestCLIBatchReportsErrorsAndContinues(t *testing.T) {
	var out, errOut bytes.Buffer
	err := runCLI(context.Background(), []string{"-ephemeral"}, strings.NewReader("unknown\npokedex\nexit\n"), &out, &errOut, false)
	if !errors.Is(err, errCommandsFailed) {
		t.Fatalf("got %v", err)
	}
	if !strings.Contains(errOut.String(), "Unknown command") || !strings.Contains(out.String(), "Your collection:") {
		t.Fatal("batch did not continue after error")
	}
	if strings.Contains(out.String(), "clibble >") {
		t.Fatal("prompt polluted piped output")
	}
}

func TestCLICountUsesSavedCollectionOffline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pokedex.json")
	if err := savePokedex(path, map[string]pokeapi.Pokemon{"pidgey": {Name: "pidgey"}, "pikachu": {Name: "pikachu"}}); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if err := runCLI(context.Background(), []string{"-data", path, "count"}, strings.NewReader(""), &out, &errOut, false); err != nil {
		t.Fatal(err)
	}
	if out.String() != "Caught Pokemon: 2\n" {
		t.Fatal(out.String())
	}
	out.Reset()
	if err := runCLI(context.Background(), []string{"-ephemeral", "count"}, strings.NewReader(""), &out, &errOut, false); err != nil {
		t.Fatal(err)
	}
	if out.String() != "Caught Pokemon: 0\n" {
		t.Fatal(out.String())
	}
}
