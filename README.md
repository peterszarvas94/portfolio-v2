# Péter Szarvas

A deliberately small, single-page portfolio and résumé, built as a static site. A templ template is rendered to `public/index.html` at build time; Caddy serves the files. IBM Plex Mono is self-hosted; its license is in `static/fonts/OFL.txt`.

## Develop

Install the tools in `mise.toml`, then:

- `mise run dev` is the live-reload dev server on http://localhost:8080. templ watches templates, Go and `static/`, rebuilds `public/` on every change and reloads the browser. Set `DEV_PORT` to change the port and `DEV_BIND=0.0.0.0` to reach it from other machines. It is a plain file server, so Caddy's redirects are not applied.
- `mise run build` renders `public/` only.
- `mise run docker` builds and runs the production image on http://localhost:8080, with the real Caddy config.
- `mise run check` checks template formatting and runs Go vet.

Career content lives in `templates/page.templ`. Employer year ranges come from the previous résumé; IBM/Kyndryl is grouped because separate dates were not available. Bandcash is explicitly shown as independent work alongside employment.

## Typography and styles

`static/style.css` uses native CSS nesting and shared `:root` variables. IBM Plex Mono has four text roles: title (24px, bold), section heading (20px, bold), body (16px, regular; company and project headings bold), and secondary text (13px, regular). Screen sizes use rem units. Dates, notes and project descriptions share the same muted style. Print overrides the four size tokens to 18/12/10/9pt. Change the role tokens rather than adding one-off sizes.

## Favicon

The favicon is a "P" in IBM Plex Mono Bold, black on an amber square with an inked border. `templates/favicon.go` reads the bold font file at build time and converts the glyph outline to an SVG path (an SVG image cannot load web fonts), and the page inlines the result as an SVG data URI, so there is no favicon file.

## Résumé PDF

`static/resume.pdf` is generated from the same page and print stylesheet with headless Chromium, and committed. Regenerate it with `mise run resume-pdf` after changing résumé content or print styles (needs `chromium` on PATH, or set `CHROMIUM_BIN`). The build fails if the PDF is missing, so the deploy never needs Chromium.

**Save resume** downloads the PDF. **Print resume** opens the browser print dialog with the same layout. Homepage extras and controls are excluded from print. The print stylesheet requests A4 portrait with a zero page margin (16mm/18mm spacing comes from the content), so browsers have no room for their date/URL headers and footers. Margins and Scale in the print dialog should stay on Default.

## Deploy

Coolify builds the `Dockerfile` (build pack: Dockerfile, exposed port 8080). The build stage generates the templ code and runs `go run ./cmd/build`; the final image is `caddy` serving `public/` with the `Caddyfile`.

Routes handled by Caddy:

| Path | Result |
| --- | --- |
| `/` | The page |
| `/resume`, `/resume/`, `/resume/index.html` | 301 to `/` (the old résumé page) |
| `/resume.pdf` | The PDF, downloaded as `Peter-Szarvas-resume.pdf` |
| `/static/Szarvas_Peter_resume.pdf` | 301 to `/resume.pdf` (the old PDF link) |
| `/demo/datastar-tiptap/*` | 301 to `https://tiptap.peterszarvas.hu/*` |

The build fingerprints the stylesheet and the script: `static/style.css` and `static/sheet.js` (the spec sheet toggle, an ES module) are published as `/static/style.<hash>.css` and `/static/sheet.<hash>.js` (first 10 hex characters of their SHA-256), and the page links those names. The hashed files and the fonts are cached for a year as immutable; any change gets a new URL. HTML and the PDF are served with `Cache-Control: no-cache` (revalidated by ETag), so a deploy shows up immediately. The local preview server (`-serve`, used by `mise run dev`) also sends `no-cache`.
