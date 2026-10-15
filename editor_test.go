package main

import "testing"

func TestCommandCompletion(t *testing.T) {
	commands := []string{"catch", "collection", "exit", "help", "inspect", "map", "mapb", "pokedex"}
	for _, tc := range []struct {
		line string
		pos  int
		key  rune
		want string
		ok   bool
	}{
		{"col", 3, '\t', "collection ", true},
		{"INS", 3, '\t', "inspect ", true},
		{"m", 1, '\t', "map", true},
		{"catch pi", 8, '\t', "catch pi", false},
		{"help", 2, '\t', "help", false},
		{"xyz", 3, '\t', "xyz", false},
	} {
		got, _, ok := completeCommand(commands, tc.line, tc.pos, tc.key)
		if got != tc.want || ok != tc.ok {
			t.Errorf("%q => %q %v", tc.line, got, ok)
		}
	}
}
