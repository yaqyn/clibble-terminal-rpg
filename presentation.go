package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/yaqyn/dexrove/internal/brand"
	"github.com/yaqyn/dexrove/internal/pokeapi"
)

type terminalView struct {
	out   io.Writer
	color bool
	width int
}

func (v *terminalView) accent(text string) string {
	if !v.color {
		return text
	}
	hex := brand.Get().Colors.Azure
	var r, g, b int
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", r, g, b, text)
}
func (v *terminalView) title(text string) { fmt.Fprintf(v.out, "\n%s\n", v.accent(text)) }
func (v *terminalView) welcome(count int, persistent bool) {
	v.title(brand.Get().Name)
	fmt.Fprintln(v.out, brand.Get().Tagline)
	mode := "Temporary session"
	if persistent {
		mode = "Collection saved automatically"
	}
	fmt.Fprintf(v.out, "\n%d caught · %s\n", count, mode)
	fmt.Fprintln(v.out, "Start with map, catch pidgey, or collection.\nTab completes commands. Up/down recalls this session.\nUse help for commands; exit or Ctrl+C to leave.")
}
func (v *terminalView) help() {
	v.title("Dexrove field guide")
	for _, group := range []struct {
		title    string
		commands [][2]string
	}{
		{"Explore", [][2]string{{"map", "Next 20 areas"}, {"mapb", "Previous areas"}, {"explore <area>", "Find Pokemon in an area"}}},
		{"Collect", [][2]string{{"catch <pokemon>", "Attempt a catch"}, {"collection", "Browse your caught Pokemon"}, {"inspect <pokemon>", "Read a creature's field record"}}},
		{"Session", [][2]string{{"help", "Show this field guide"}, {"exit", "Leave Dexrove"}}},
	} {
		fmt.Fprintf(v.out, "\n%s\n", v.accent(group.title))
		for _, cmd := range group.commands {
			if v.width < 60 {
				fmt.Fprintf(v.out, "  %s\n    %s\n", cmd[0], cmd[1])
			} else {
				fmt.Fprintf(v.out, "  %-22s %s\n", cmd[0], cmd[1])
			}
		}
	}
	fmt.Fprintln(v.out, "\nBrowse saved records in a browser: dexrove serve\nThe pokedex command remains an alias for collection.")
}
func (v *terminalView) inspect(p pokeapi.Pokemon) {
	v.title(p.Name)
	types := make([]string, 0, len(p.Types))
	for _, kind := range p.Types {
		types = append(types, kind.Type.Name)
	}
	if len(types) > 0 {
		fmt.Fprintf(v.out, "Type     %s\n", strings.Join(types, " / "))
	}
	fmt.Fprintf(v.out, "Height   %.1f m\nWeight   %.1f kg\n", float64(p.Height)/10, float64(p.Weight)/10)
	v.title("Base stats")
	for _, stat := range p.Stats {
		if v.width < 60 {
			fmt.Fprintf(v.out, "  %s: %d\n", stat.Stat.Name, stat.BaseStat)
			continue
		}
		length := stat.BaseStat * 20 / 255
		if length > 20 {
			length = 20
		}
		fmt.Fprintf(v.out, "  %-17s %3d  %s\n", stat.Stat.Name, stat.BaseStat, v.accent(strings.Repeat("━", length)+strings.Repeat("·", 20-length)))
	}
	fmt.Fprintln(v.out, "\nStored in this collection. Read without a network request.")
}
func (v *terminalView) collection(pokemon map[string]pokeapi.Pokemon) {
	v.title("Your collection")
	if len(pokemon) == 0 {
		fmt.Fprintln(v.out, "No Pokemon caught yet. Try catch pidgey.")
		return
	}
	fmt.Fprintf(v.out, "%d caught\n\n", len(pokemon))
	for _, name := range sortedKeys(pokemon) {
		p := pokemon[name]
		types := make([]string, 0, len(p.Types))
		for _, kind := range p.Types {
			types = append(types, kind.Type.Name)
		}
		if v.width < 60 {
			fmt.Fprintf(v.out, "  %s\n", name)
		} else {
			fmt.Fprintf(v.out, "  %-24s %s\n", name, strings.Join(types, " / "))
		}
	}
	fmt.Fprintln(v.out, "\nUse inspect <pokemon> to open a record.")
}

func (v *terminalView) locations(areas []pokeapi.Resource) {
	v.title("Location areas")
	columns := 1
	if v.width >= 90 {
		columns = 2
		for _, area := range areas {
			if len(area.Name)+4 > v.width/2 {
				columns = 1
				break
			}
		}
	}
	for index, area := range areas {
		if columns == 2 && index%2 == 0 && index+1 < len(areas) {
			fmt.Fprintf(v.out, "  %-*s", v.width/2-2, area.Name)
		} else {
			fmt.Fprintf(v.out, "  %s\n", area.Name)
		}
	}
	fmt.Fprintln(v.out, "\nmap: next areas  ·  mapb: previous areas\nexplore <area>: see who lives here")
}
