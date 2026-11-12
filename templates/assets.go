package templates

import (
	"context"
	"net/url"
	"strings"
	"sync"
)

// Stylesheet is the page's stylesheet URL. The production build replaces it with a
// content-hashed name (see cmd/build) so the file can be cached forever.
var Stylesheet = "/static/style.css"

// faviconHref renders the Favicon component once and inlines it as an SVG data URI,
// so the icon ships inside the static HTML with no separate file.
var faviconHref = sync.OnceValue(func() string {
	var b strings.Builder
	if err := Favicon().Render(context.Background(), &b); err != nil {
		panic(err)
	}
	return "data:image/svg+xml," + url.PathEscape(b.String())
})
