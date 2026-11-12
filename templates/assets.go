package templates

import (
	"fmt"
	"net/url"
	"sync"
)

// Stylesheet is the page's stylesheet URL. The production build replaces it with a
// content-hashed name (see cmd/build) so the file can be cached forever.
var Stylesheet = "/static/style.css"

// Script is the page's ES module (the spec sheet toggle), fingerprinted the same way.
var Script = "/static/sheet.js"

// faviconHref draws the favicon once and inlines it as an SVG data URI,
// so the icon ships inside the static HTML with no separate file.
var faviconHref = sync.OnceValue(func() string {
	svg, err := faviconSVG()
	if err != nil {
		panic(fmt.Errorf("favicon: %w", err))
	}
	return "data:image/svg+xml," + url.PathEscape(svg)
})
