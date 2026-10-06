// Package web embeds the built frontend (web/dist, produced by the Vite build).
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// Dist returns the built frontend files. Without a frontend build it only contains .gitkeep.
func Dist() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err) // "dist" is a compile-time constant embedded above
	}
	return sub
}
