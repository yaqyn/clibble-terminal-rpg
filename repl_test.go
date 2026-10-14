package main

import (
	"reflect"
	"testing"
)

func TestCleanInput(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  []string
	}{
		{"  hello  world  ", []string{"hello", "world"}},
		{"Charmander Bulbasaur PIKACHU", []string{"charmander", "bulbasaur", "pikachu"}},
		{"\tMAP\nmapb\r", []string{"map", "mapb"}},
		{"", []string{}},
	} {
		if got := cleanInput(tc.input); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("cleanInput(%q) = %v; want %v", tc.input, got, tc.want)
		}
	}
}
