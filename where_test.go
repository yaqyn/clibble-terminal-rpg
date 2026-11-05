package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yaqyn/clibble-terminal-rpg/internal/pokeapi"
)

func TestTrailsListsBundledGuideWithoutAnAPI(t *testing.T) {
	var output bytes.Buffer
	if err := commandTrails(&config{out: &output}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "mosslight-woods: 3 wildlings") ||
		!strings.Contains(output.String(), "old-root-ridge: 3 wildlings") {
		t.Fatal(output.String())
	}
}

func TestWhereListsBundledHabitatsWithoutAnAPI(t *testing.T) {
	var output bytes.Buffer
	cfg := &config{out: &output}
	if err := commandWhere(cfg, "mosskin"); err != nil {
		t.Fatal(err)
	}
	if output.String() != "mosslight-woods\nbramble-hollow\n" {
		t.Fatal(output.String())
	}
	output.Reset()
	if err := commandWhere(cfg, "unknown"); err == nil {
		t.Fatal("unknown wildling accepted")
	}
	if output.Len() != 0 {
		t.Fatal("unknown wildling produced misleading locations")
	}
}

func TestNextTrailRecommendsTheLargestCollectionGap(t *testing.T) {
	var output bytes.Buffer
	cfg := &config{out: &output, pokedex: map[string]pokeapi.Pokemon{}}
	if err := commandNextTrail(cfg); err != nil {
		t.Fatal(err)
	}
	if output.String() != "Recommended trail: bramble-hollow (3 uncaught wildlings)\n" {
		t.Fatal(output.String())
	}
}

func TestHabitatsListsOnlyWildlingsSharedByTrails(t *testing.T) {
	var output bytes.Buffer
	if err := commandHabitats(&config{out: &output}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "mosskin: mosslight-woods, bramble-hollow\n") {
		t.Fatal(output.String())
	}
	if strings.Contains(output.String(), "bramblebun:") {
		t.Fatal("single-trail wildling was reported")
	}
}
