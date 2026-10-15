// Package webui embeds the local, offline-capable collection interface.
package webui

import (
	"embed"
	"io/fs"
)

//go:embed assets
var files embed.FS

func Assets() fs.FS {
	assets, err := fs.Sub(files, "assets")
	if err != nil {
		panic(err)
	}
	return assets
}
