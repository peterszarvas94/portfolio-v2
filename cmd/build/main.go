// Build renders the site into public/ as plain files for Caddy to serve.
//
//	go run ./cmd/build          build public/
//	go run ./cmd/build -serve   build, then serve public/ on $PORT (default 8080) for local preview
package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"portfolio/templates"
)

const outDir = "public"

func main() {
	serve := flag.Bool("serve", false, "serve public/ after building (local preview only; no redirects)")
	flag.Parse()

	if err := build(); err != nil {
		log.Fatal(err)
	}
	log.Printf("Built %s/", outDir)

	if *serve {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		log.Printf("Preview: http://localhost:%s", port)
		// Revalidate every request, as Caddy does in production, so edits never show stale CSS.
		// Page requests rebuild first: templ's watch mode hot-reloads text-only template edits
		// without re-running this command, so the static output would otherwise go stale.
		files := http.FileServer(http.Dir(outDir))
		var mu sync.RWMutex
		log.Fatal(http.ListenAndServe(":"+port, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" || strings.HasSuffix(r.URL.Path, ".html") {
				mu.Lock()
				err := build()
				mu.Unlock()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
			mu.RLock()
			defer mu.RUnlock()
			w.Header().Set("Cache-Control", "no-cache")
			files.ServeHTTP(w, r)
		})))
	}
}

func build() error {
	if err := os.RemoveAll(outDir); err != nil {
		return err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	// The résumé PDF and the icons clients look for by name live at the site root;
	// everything else under /static/. The SVG favicon is inlined in the page (templates/assets.go).
	if err := copyFile("static/resume.pdf", filepath.Join(outDir, "resume.pdf")); err != nil {
		return fmt.Errorf("resume.pdf missing; run `mise run resume-pdf`: %w", err)
	}
	for _, name := range []string{"favicon.ico", "apple-touch-icon.png"} {
		if err := copyFile(filepath.Join("static/brand", name), filepath.Join(outDir, name)); err != nil {
			return fmt.Errorf("%s missing; run `mise run brand`: %w", name, err)
		}
	}
	err := filepath.WalkDir("static", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "static/resume.pdf" || fingerprinted[path] {
			return nil
		}
		target := filepath.Join(outDir, path)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
	if err != nil {
		return err
	}

	// Fingerprint the stylesheet and script: content-hashed names let Caddy cache them
	// forever, and every change gets a new URL, so a deploy can never serve stale assets.
	if templates.Stylesheet, err = fingerprint("style.css"); err != nil {
		return err
	}
	if templates.Script, err = fingerprint("sheet.js"); err != nil {
		return err
	}

	page, err := os.Create(filepath.Join(outDir, "index.html"))
	if err != nil {
		return err
	}
	if err := templates.Page().Render(context.Background(), page); err != nil {
		page.Close()
		return fmt.Errorf("render page: %w", err)
	}
	return page.Close()
}

// fingerprinted are the static files published only under a content-hashed name.
var fingerprinted = map[string]bool{"static/style.css": true, "static/sheet.js": true}

// fingerprint publishes static/<name> as /static/<base>.<hash><ext> (first 10 hex
// characters of its SHA-256) and returns that URL.
func fingerprint(name string) (string, error) {
	data, err := os.ReadFile(filepath.Join("static", name))
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	ext := filepath.Ext(name)
	hashed := fmt.Sprintf("%s.%x%s", strings.TrimSuffix(name, ext), sum[:5], ext)
	if err := os.WriteFile(filepath.Join(outDir, "static", hashed), data, 0o644); err != nil {
		return "", err
	}
	return "/static/" + hashed, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
