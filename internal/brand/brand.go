// Package brand owns Clibble's shared product identity and palette.
package brand

import (
	"embed"
	"encoding/json"
)

//go:embed identity.json
var files embed.FS

type Palette struct{ Primary, Ink, Mist, Cloud, Slate, Fern string }
type ForestPalette struct{ Pine, Canopy, Fern, Leaf, Selection, Ivory string }
type Identity struct {
	Forest              ForestPalette
	Name, Slug, Tagline string
	Colors              Palette
}

var identity = func() Identity {
	data, err := files.ReadFile("identity.json")
	if err != nil {
		panic(err)
	}
	var value Identity
	if err := json.Unmarshal(data, &value); err != nil {
		panic(err)
	}
	return value
}()

// Get returns a value copy; callers cannot mutate the shared identity.
func Get() Identity { return identity }
