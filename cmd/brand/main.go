// Brand renders the logo, icons and link-preview card into static/brand/; commit the result.
//
//	go run ./cmd/brand
package main

import (
	"log"
	"os"
	"path/filepath"

	"portfolio/templates"
)

const outDir = "static/brand"

func main() {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Fatal(err)
	}
	files := []struct {
		name   string
		render func() ([]byte, error)
	}{
		{"logo.svg", func() ([]byte, error) { s, err := templates.LogoSVG(); return []byte(s + "\n"), err }},
		{"logo.png", func() ([]byte, error) { return templates.LogoPNG(1024) }},
		{"apple-touch-icon.png", func() ([]byte, error) { return templates.LogoPNG(180) }},
		{"favicon.ico", templates.FaviconICO},
		{"og.png", func() ([]byte, error) {
			return templates.OGImage("Péter Szarvas", "I build software. And the things around it.", "peterszarvas.hu")
		}},
	}
	for _, f := range files {
		b, err := f.render()
		if err != nil {
			log.Fatalf("%s: %v", f.name, err)
		}
		path := filepath.Join(outDir, f.name)
		if err := os.WriteFile(path, b, 0o644); err != nil {
			log.Fatal(err)
		}
		log.Printf("Wrote %s", path)
	}
}
