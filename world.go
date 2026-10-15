package main

import (
	"context"
	"fmt"
	"github.com/yaqyn/clibble-pokemon-cli/internal/pokeapi"
)

// meadowWorld is Clibble's bundled, original world. The legacy record shape is
// retained so existing journals and the read-only companion remain compatible.
type meadowWorld struct{}
type trailEntry struct {
	name      string
	creatures []string
}

var trails = []trailEntry{
	{"mosslight-woods", []string{"mosskin", "fernlet", "puddlehop"}},
	{"willow-brook", []string{"puddlehop", "reedwhisk", "dewbell"}},
	{"bramble-hollow", []string{"bramblebun", "mosskin", "cinderpip"}},
	{"moonfern-glade", []string{"glimwing", "dewbell", "hushroot"}},
	{"sunseed-meadow", []string{"fernlet", "sunmuff", "cinderpip"}},
	{"old-root-ridge", []string{"hushroot", "reedwhisk", "glimwing"}},
}
var wildlings = map[string]struct {
	habitat                 string
	shyness, height, weight int
}{
	"mosskin": {"moss", 25, 3, 12}, "fernlet": {"leaf", 40, 4, 15}, "puddlehop": {"brook", 55, 2, 8},
	"reedwhisk": {"brook", 75, 5, 22}, "dewbell": {"mist", 100, 3, 9}, "bramblebun": {"bramble", 65, 4, 18},
	"cinderpip": {"ember", 120, 2, 6}, "glimwing": {"moon", 160, 3, 4}, "hushroot": {"root", 190, 8, 60}, "sunmuff": {"sun", 85, 4, 20},
}

func (meadowWorld) ListLocations(ctx context.Context, _ *string) (pokeapi.Locations, error) {
	if err := ctx.Err(); err != nil {
		return pokeapi.Locations{}, err
	}
	result := pokeapi.Locations{}
	for _, trail := range trails {
		result.Results = append(result.Results, pokeapi.Resource{Name: trail.name})
	}
	return result, nil
}
func (meadowWorld) Explore(ctx context.Context, name string) (pokeapi.Area, error) {
	if err := ctx.Err(); err != nil {
		return pokeapi.Area{}, err
	}
	for _, trail := range trails {
		if trail.name == name {
			result := pokeapi.Area{}
			for _, name := range trail.creatures {
				result.PokemonEncounters = append(result.PokemonEncounters, struct {
					Pokemon pokeapi.Resource `json:"pokemon"`
				}{Pokemon: pokeapi.Resource{Name: name}})
			}
			return result, nil
		}
	}
	return pokeapi.Area{}, fmt.Errorf("unknown trail %q; use map to find a trail", name)
}
func (meadowWorld) GetPokemon(ctx context.Context, name string) (pokeapi.Pokemon, error) {
	if err := ctx.Err(); err != nil {
		return pokeapi.Pokemon{}, err
	}
	entry, ok := wildlings[name]
	if !ok {
		return pokeapi.Pokemon{}, fmt.Errorf("unknown wildling %q; explore a trail to meet one", name)
	}
	p := pokeapi.Pokemon{Name: name, BaseExperience: entry.shyness, Height: entry.height, Weight: entry.weight}
	p.Types = append(p.Types, struct {
		Type pokeapi.Resource `json:"type"`
	}{Type: pokeapi.Resource{Name: entry.habitat}})
	for _, stat := range []struct {
		name  string
		value int
	}{{"curiosity", 100 - entry.shyness/3}, {"spirit", 30 + entry.shyness/4}, {"nimbleness", 45 + entry.shyness/5}} {
		p.Stats = append(p.Stats, struct {
			BaseStat int              `json:"base_stat"`
			Stat     pokeapi.Resource `json:"stat"`
		}{BaseStat: stat.value, Stat: pokeapi.Resource{Name: stat.name}})
	}
	return p, nil
}

// Each wildling has a small, authored character moment; no remote lore or assets.
var wildlingStories = map[string]string{
	"mosskin":    "It peeks from a mossy hood, carrying a fallen leaf like an umbrella.",
	"fernlet":    "It uncurls a frond and waves as though it has been expecting you.",
	"puddlehop":  "It skips between raindrops, leaving tiny rings in the puddles.",
	"reedwhisk":  "It braids river reeds into a nest that floats behind it.",
	"dewbell":    "A clear note rings out whenever a dewdrop touches its petals.",
	"bramblebun": "It offers you a berry, then shyly hides behind its own brambles.",
	"cinderpip":  "Its little ember keeps a patch of grass warm through the rain.",
	"glimwing":   "Its wings collect moonlight and scatter it like dandelion seeds.",
	"hushroot":   "You almost mistake it for an old root until it opens one sleepy eye.",
	"sunmuff":    "It shakes the rain from its fluffy crown and turns toward the light.",
}
