// Generate the downloadable resume from the same template and print stylesheet.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/a-h/templ"
	"portfolio/templates"
)

func main() {
	if err := generate(); err != nil {
		log.Fatal(err)
	}
}

func generate() error {
	browser := os.Getenv("CHROMIUM_BIN")
	if browser == "" {
		browser = "chromium"
	}
	mux := http.NewServeMux()
	mux.Handle("/", templ.Handler(templates.Page()))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	server := httptest.NewServer(mux)
	defer server.Close()
	profile, err := os.MkdirTemp("", "portfolio-pdf-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(profile)
	output, err := filepath.Abs("static/resume.pdf")
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, browser, "--headless", "--no-sandbox", "--disable-gpu", "--no-pdf-header-footer", "--user-data-dir="+profile, "--print-to-pdf="+output, server.URL)
	if data, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("generate resume: %w\n%s", err, data)
	}
	log.Printf("Generated %s", output)
	return nil
}
