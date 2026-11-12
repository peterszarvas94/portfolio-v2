package templates

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// faviconFont is the page's bold face; the favicon's letter is drawn from it.
const faviconFont = "static/fonts/IBMPlexMono-Bold.ttf"

// faviconSVG draws a "P" from IBM Plex Mono Bold on an amber square with an inked border.
// An SVG used as an image can't load web fonts, so the glyph's outline is converted to a
// path at build time. Colours mirror the CSS tokens (--amber, --ink).
func faviconSVG() (string, error) {
	data, err := os.ReadFile(faviconFont)
	if err != nil {
		return "", err
	}
	f, err := sfnt.Parse(data)
	if err != nil {
		return "", err
	}
	var buf sfnt.Buffer
	idx, err := f.GlyphIndex(&buf, 'P')
	if err != nil {
		return "", err
	}
	// Load at one pixel per font unit; sfnt's Y axis points down, like SVG's.
	segs, err := f.LoadGlyph(&buf, idx, fixed.I(int(f.UnitsPerEm())), nil)
	if err != nil {
		return "", err
	}

	// Fit the glyph's ink box to 20 of 32 units tall, centred.
	const box, height = 32.0, 20.0
	b := segs.Bounds()
	minX, minY := unit(b.Min.X), unit(b.Min.Y)
	maxX, maxY := unit(b.Max.X), unit(b.Max.Y)
	s := height / (maxY - minY)
	tx := box/2 - (minX+maxX)/2*s
	ty := box/2 - (minY+maxY)/2*s
	pt := func(p fixed.Point26_6) string {
		return fmt.Sprintf("%.2f %.2f", unit(p.X)*s+tx, unit(p.Y)*s+ty)
	}

	var d strings.Builder
	for _, seg := range segs {
		switch seg.Op {
		case sfnt.SegmentOpMoveTo:
			if d.Len() > 0 {
				d.WriteString("Z")
			}
			d.WriteString("M" + pt(seg.Args[0]))
		case sfnt.SegmentOpLineTo:
			d.WriteString("L" + pt(seg.Args[0]))
		case sfnt.SegmentOpQuadTo:
			d.WriteString("Q" + pt(seg.Args[0]) + " " + pt(seg.Args[1]))
		case sfnt.SegmentOpCubeTo:
			d.WriteString("C" + pt(seg.Args[0]) + " " + pt(seg.Args[1]) + " " + pt(seg.Args[2]))
		}
	}
	d.WriteString("Z")

	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">` +
		`<rect x="0.75" y="0.75" width="30.5" height="30.5" fill="#ffb700" stroke="#111" stroke-width="1.5"/>` +
		`<path fill="#111" d="` + d.String() + `"/></svg>`, nil
}

func unit(v fixed.Int26_6) float64 { return float64(v) / 64 }
