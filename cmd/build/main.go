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

	// The résumé PDF lives at the site root; everything else under /static/.
	// The favicon is an SVG data URI inlined in the page (templates/favicon.templ).
	if err := copyFile("static/resume.pdf", filepath.Join(outDir, "resume.pdf")); err != nil {
		return fmt.Errorf("resume.pdf missing; run `mise run resume-pdf`: %w", err)
	}
	err := filepath.WalkDir("static", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "static/resume.pdf" || path == "static/style.css" {
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

	// Fingerprint the stylesheet: a content-hashed name lets Caddy cache it forever,
	// and every change gets a new URL, so a deploy can never serve stale CSS.
	css, err := os.ReadFile("static/style.css")
	if err != nil {
		return err
	}
	sum := sha256.Sum256(css)
	name := fmt.Sprintf("style.%x.css", sum[:5])
	if err := os.WriteFile(filepath.Join(outDir, "static", name), css, 0o644); err != nil {
		return err
	}
	templates.Stylesheet = "/static/" + name

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
