package templates

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"
)

// The logo is a "P" from IBM Plex Mono Bold, ink on an amber square with an inked border.
// It is drawn on a 32-unit square: the border is 1.5 units wide and the glyph's ink box is
// 20 units tall, centred. Colours mirror the CSS tokens.
const (
	fontRegular = "static/fonts/IBMPlexMono-Regular.ttf"
	fontBold    = "static/fonts/IBMPlexMono-Bold.ttf"

	logoBox    = 32.0
	logoBorder = 1.5
	logoGlyph  = 20.0
)

var (
	colInk    = color.RGBA{0x11, 0x11, 0x11, 0xff} // --ink
	colMuted  = color.RGBA{0x5c, 0x5c, 0x5c, 0xff} // --muted
	colPaper  = color.RGBA{0xff, 0xff, 0xff, 0xff} // --paper
	colCanvas = color.RGBA{0xe8, 0xe7, 0xe3, 0xff} // --canvas
	colShadow = color.RGBA{0xb8, 0xb7, 0xb2, 0xff} // --shadow
	colAmber  = color.RGBA{0xff, 0xb7, 0x00, 0xff} // --amber
)

// segment is one outline command of the glyph, in logo units (0–32, Y down).
type segment struct {
	op  sfnt.SegmentOp
	pts [][2]float64
}

// logoOutline loads the "P" from the bold font and fits it into the logo square.
// An SVG used as an image can't load web fonts, so the glyph is always drawn as a path.
func logoOutline() ([]segment, error) {
	data, err := os.ReadFile(fontBold)
	if err != nil {
		return nil, err
	}
	f, err := sfnt.Parse(data)
	if err != nil {
		return nil, err
	}
	var buf sfnt.Buffer
	idx, err := f.GlyphIndex(&buf, 'P')
	if err != nil {
		return nil, err
	}
	// Load at one pixel per font unit; sfnt's Y axis points down, like SVG's.
	segs, err := f.LoadGlyph(&buf, idx, fixed.I(int(f.UnitsPerEm())), nil)
	if err != nil {
		return nil, err
	}

	b := segs.Bounds()
	minX, minY := unit(b.Min.X), unit(b.Min.Y)
	maxX, maxY := unit(b.Max.X), unit(b.Max.Y)
	s := logoGlyph / (maxY - minY)
	tx := logoBox/2 - (minX+maxX)/2*s
	ty := logoBox/2 - (minY+maxY)/2*s

	out := make([]segment, len(segs))
	for i, seg := range segs {
		n := map[sfnt.SegmentOp]int{sfnt.SegmentOpMoveTo: 1, sfnt.SegmentOpLineTo: 1, sfnt.SegmentOpQuadTo: 2, sfnt.SegmentOpCubeTo: 3}[seg.Op]
		out[i].op = seg.Op
		for _, p := range seg.Args[:n] {
			out[i].pts = append(out[i].pts, [2]float64{unit(p.X)*s + tx, unit(p.Y)*s + ty})
		}
	}
	return out, nil
}

func unit(v fixed.Int26_6) float64 { return float64(v) / 64 }

// LogoSVG returns the logo as a standalone SVG document.
func LogoSVG() (string, error) {
	segs, err := logoOutline()
	if err != nil {
		return "", err
	}
	pt := func(p [2]float64) string { return fmt.Sprintf("%.2f %.2f", p[0], p[1]) }
	var d strings.Builder
	for _, seg := range segs {
		switch seg.op {
		case sfnt.SegmentOpMoveTo:
			if d.Len() > 0 {
				d.WriteString("Z")
			}
			d.WriteString("M" + pt(seg.pts[0]))
		case sfnt.SegmentOpLineTo:
			d.WriteString("L" + pt(seg.pts[0]))
		case sfnt.SegmentOpQuadTo:
			d.WriteString("Q" + pt(seg.pts[0]) + " " + pt(seg.pts[1]))
		case sfnt.SegmentOpCubeTo:
			d.WriteString("C" + pt(seg.pts[0]) + " " + pt(seg.pts[1]) + " " + pt(seg.pts[2]))
		}
	}
	d.WriteString("Z")

	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">`+
		`<rect x="%[1]g" y="%[1]g" width="%[2]g" height="%[2]g" fill="#ffb700" stroke="#111" stroke-width="%[3]g"/>`+
		`<path fill="#111" d="%[4]s"/></svg>`,
		logoBorder/2, logoBox-logoBorder, logoBorder, d.String()), nil
}

// drawLogo paints the logo as a size×size square with its top-left corner at (x, y).
func drawLogo(dst *image.RGBA, x, y, size int) error {
	segs, err := logoOutline()
	if err != nil {
		return err
	}
	k := float32(size) / logoBox
	r := image.Rect(x, y, x+size, y+size)
	fill := func(c color.RGBA, path func(z *vector.Rasterizer)) {
		z := vector.NewRasterizer(size, size)
		path(z)
		z.Draw(dst, r, image.NewUniform(c), image.Point{})
	}
	rect := func(z *vector.Rasterizer, inset float32) {
		lo, hi := inset*k, (logoBox-inset)*k
		z.MoveTo(lo, lo)
		z.LineTo(hi, lo)
		z.LineTo(hi, hi)
		z.LineTo(lo, hi)
		z.ClosePath()
	}

	fill(colInk, func(z *vector.Rasterizer) { rect(z, 0) })
	fill(colAmber, func(z *vector.Rasterizer) { rect(z, logoBorder) })
	fill(colInk, func(z *vector.Rasterizer) {
		p := func(i int, s segment) (float32, float32) { return float32(s.pts[i][0]) * k, float32(s.pts[i][1]) * k }
		for i, s := range segs {
			switch s.op {
			case sfnt.SegmentOpMoveTo:
				if i > 0 {
					z.ClosePath()
				}
				z.MoveTo(p(0, s))
			case sfnt.SegmentOpLineTo:
				z.LineTo(p(0, s))
			case sfnt.SegmentOpQuadTo:
				ax, ay := p(0, s)
				bx, by := p(1, s)
				z.QuadTo(ax, ay, bx, by)
			case sfnt.SegmentOpCubeTo:
				ax, ay := p(0, s)
				bx, by := p(1, s)
				cx, cy := p(2, s)
				z.CubeTo(ax, ay, bx, by, cx, cy)
			}
		}
		z.ClosePath()
	})
	return nil
}

// LogoPNG returns the logo as a size×size PNG.
func LogoPNG(size int) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	if err := drawLogo(img, 0, 0, size); err != nil {
		return nil, err
	}
	return encodePNG(img)
}

// FaviconICO returns a favicon.ico holding 16, 32 and 48 px PNGs, for clients
// that ask for /favicon.ico instead of reading the page's SVG icon.
func FaviconICO() ([]byte, error) {
	sizes := []int{16, 32, 48}
	var imgs [][]byte
	for _, s := range sizes {
		b, err := LogoPNG(s)
		if err != nil {
			return nil, err
		}
		imgs = append(imgs, b)
	}
	var out bytes.Buffer
	w := func(v any) { binary.Write(&out, binary.LittleEndian, v) }
	w([3]uint16{0, 1, uint16(len(sizes))}) // reserved, type icon, count
	offset := 6 + 16*len(sizes)
	for i, s := range sizes {
		w([4]uint8{uint8(s), uint8(s), 0, 0}) // width, height, palette, reserved
		w([2]uint16{1, 32})                   // planes, bits per pixel
		w([2]uint32{uint32(len(imgs[i])), uint32(offset)})
		offset += len(imgs[i])
	}
	for _, b := range imgs {
		out.Write(b)
	}
	return out.Bytes(), nil
}

// OGWidth and OGHeight are the Open Graph card's size, the 1.91:1 ratio link previews use.
const OGWidth, OGHeight = 1200, 630

// OGImage returns the link-preview card: the logo, name, intro and domain on a paper
// sheet with the site's hard offset shadow.
func OGImage(name, intro, domain string) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, OGWidth, OGHeight))
	box := func(r image.Rectangle, c color.RGBA) {
		draw.Draw(img, r, image.NewUniform(c), image.Point{}, draw.Src)
	}

	// Canvas, shadow, inked border, paper.
	box(img.Bounds(), colCanvas)
	sheet := image.Rect(64, 64, OGWidth-64, OGHeight-64)
	box(sheet.Add(image.Pt(10, 10)), colShadow)
	box(sheet, colInk)
	box(sheet.Inset(3), colPaper)

	const logo = 176
	lx, ly := sheet.Min.X+96, (OGHeight-logo)/2
	if err := drawLogo(img, lx, ly, logo); err != nil {
		return nil, err
	}

	text := func(path string, size float64, c color.RGBA, x, y int, s string) error {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		f, err := opentype.Parse(data)
		if err != nil {
			return err
		}
		face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingNone})
		if err != nil {
			return err
		}
		defer face.Close()
		d := font.Drawer{Dst: img, Src: image.NewUniform(c), Face: face, Dot: fixed.P(x, y)}
		d.DrawString(s)
		return nil
	}

	// Text column: name on the logo's cap line, domain on its baseline.
	tx := lx + logo + 56
	for _, t := range []struct {
		path  string
		size  float64
		c     color.RGBA
		y     int
		value string
	}{
		{fontBold, 64, colInk, ly + 62, name},
		{fontRegular, 26, colInk, ly + 112, intro},
		{fontRegular, 24, colMuted, ly + logo - 4, domain},
	} {
		if err := text(t.path, t.size, t.c, tx, t.y, t.value); err != nil {
			return nil, err
		}
	}
	return encodePNG(img)
}

func encodePNG(img image.Image) ([]byte, error) {
	var b bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if err := enc.Encode(&b, img); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
