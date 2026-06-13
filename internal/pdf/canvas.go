// Package pdf renders a resume document into a PDF. It draws the JIS-style
// 履歴書 and a flowing 職務経歴書 with millimetre coordinates on A4 pages using
// the embedded IPAex fonts.
package pdf

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/nao1215/career/internal/font"
	"github.com/signintech/gopdf"
)

// solid resets dashed line drawing back to a continuous stroke. gopdf treats an
// empty line type as solid.
const solid = ""

// canvas wraps a gopdf document with millimetre-oriented helpers. The origin is
// the upper-left corner of the page and y grows downward.
type canvas struct {
	pdf      *gopdf.GoPdf
	curFont  string
	curSize  float64
	curColor rgb
}

// newCanvas starts an A4 document in millimetre units with the embedded fonts
// registered.
func newCanvas() (*canvas, error) {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{Unit: gopdf.UnitMM, PageSize: *gopdf.PageSizeA4})
	if err := font.Register(pdf); err != nil {
		return nil, err
	}
	return &canvas{pdf: pdf}, nil
}

// bytes renders the document to a byte slice.
func (c *canvas) bytes() ([]byte, error) {
	var buf bytes.Buffer
	if _, err := c.pdf.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("render pdf: %w", err)
	}
	return buf.Bytes(), nil
}

// setFont selects a font family and size, skipping the call when nothing
// changed so repeated text draws stay cheap.
func (c *canvas) setFont(family string, size float64) {
	if c.curFont == family && c.curSize == size {
		return
	}
	// SetFont only fails when the family was never registered, which is a
	// programming error given the fixed set of embedded fonts.
	if err := c.pdf.SetFont(family, "", size); err != nil {
		panic(fmt.Sprintf("career: set font %q: %v", family, err))
	}
	c.curFont = family
	c.curSize = size
}

// text draws s with its left edge at x and its top at y.
func (c *canvas) text(x, y float64, s string) {
	if s == "" {
		return
	}
	c.pdf.SetXY(x, y)
	_ = c.pdf.Cell(nil, s)
}

// setColor selects the text color, skipping the call when nothing changed.
func (c *canvas) setColor(color rgb) {
	if c.curColor == color {
		return
	}
	c.pdf.SetTextColor(color.r, color.g, color.b)
	c.curColor = color
}

// fillRect draws a filled rectangle in the given color, then restores black as
// the fill color so later strokes are unaffected.
func (c *canvas) fillRect(x, y, w, h float64, color rgb) {
	c.pdf.SetFillColor(color.r, color.g, color.b)
	c.pdf.RectFromUpperLeftWithStyle(x, y, w, h, "F")
	c.pdf.SetFillColor(black.r, black.g, black.b)
}

// textWidth measures the rendered width of s at the current font.
func (c *canvas) textWidth(s string) float64 {
	w, err := c.pdf.MeasureTextWidth(s)
	if err != nil {
		return 0
	}
	return w
}

// textFit draws s at (x, y) in the Mincho font, shrinking from size down toward
// minSize until the text fits within maxWidth. It keeps long values (addresses,
// emails) from overflowing their cell.
func (c *canvas) textFit(x, y, maxWidth float64, s string, size, minSize float64) {
	if s == "" {
		return
	}
	sz := size
	for sz > minSize {
		c.setFont(font.Mincho, sz)
		if c.textWidth(s) <= maxWidth {
			break
		}
		sz -= 0.5
	}
	c.setFont(font.Mincho, sz)
	c.text(x, y, s)
}

// textRight draws s so its right edge sits at xRight.
func (c *canvas) textRight(xRight, y float64, s string) {
	c.text(xRight-c.textWidth(s), y, s)
}

// textCenter draws s centred horizontally between x and x+width.
func (c *canvas) textCenter(x, width, y float64, s string) {
	c.text(x+(width-c.textWidth(s))/2, y, s)
}

// line draws a solid line between two points.
func (c *canvas) line(x1, y1, x2, y2 float64) {
	c.pdf.SetLineType(solid)
	c.pdf.Line(x1, y1, x2, y2)
}

// rect draws an unfilled rectangle with its upper-left corner at (x, y).
func (c *canvas) rect(x, y, w, h float64) {
	c.pdf.SetLineType(solid)
	c.pdf.RectFromUpperLeftWithStyle(x, y, w, h, "D")
}

// wrap breaks s into lines no wider than maxWidth at the current font. Explicit
// newlines are honoured. Latin text breaks at spaces; Japanese text, which has
// no spaces, may break between any two characters; an unbreakable run that is
// still too long is split per character.
func (c *canvas) wrap(s string, maxWidth float64) []string {
	var lines []string
	for _, paragraph := range strings.Split(s, "\n") {
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}
		lines = append(lines, c.wrapParagraph(paragraph, maxWidth)...)
	}
	return lines
}

func (c *canvas) wrapParagraph(paragraph string, maxWidth float64) []string {
	runes := []rune(paragraph)
	var lines []string
	start := 0
	lastBreak := -1 // rune index at which the current line may break

	for i := start; i < len(runes); i++ {
		if i > start && canBreakBefore(runes[i-1], runes[i]) {
			lastBreak = i
		}
		seg := strings.TrimRight(string(runes[start:i+1]), " ")
		if i > start && c.textWidth(seg) > maxWidth {
			brk := lastBreak
			if brk <= start {
				brk = i // no break opportunity: hard-split before the current rune
			}
			lines = append(lines, strings.TrimRight(string(runes[start:brk]), " "))
			start = brk
			for start < len(runes) && runes[start] == ' ' {
				start++
			}
			lastBreak = -1
			i = start - 1 // rescan from the new line start
		}
	}
	if start < len(runes) {
		lines = append(lines, strings.TrimRight(string(runes[start:]), " "))
	}
	if len(lines) == 0 {
		lines = append(lines, "")
	}
	return lines
}

// canBreakBefore reports whether a line may break between prev and next: after a
// space, or adjacent to a CJK character (which carries no spaces of its own).
func canBreakBefore(prev, next rune) bool {
	return prev == ' ' || isCJK(prev) || isCJK(next)
}

// isCJK reports whether r is a CJK ideograph, kana, or CJK punctuation that may
// start or end a line.
func isCJK(r rune) bool {
	switch {
	case r >= 0x3040 && r <= 0x30ff: // hiragana + katakana
		return true
	case r >= 0x3400 && r <= 0x9fff: // CJK unified ideographs (incl. ext A)
		return true
	case r >= 0xff00 && r <= 0xffef: // full-width forms
		return true
	case r >= 0x3000 && r <= 0x303f: // CJK symbols and punctuation
		return true
	}
	return false
}

// paragraph draws wrapped text starting at (x, y) within maxWidth, advancing y
// by lineHeight per line. It returns the y position just below the last line.
func (c *canvas) paragraph(x, y, maxWidth, lineHeight float64, s string) float64 {
	for _, line := range c.wrap(s, maxWidth) {
		c.text(x, y, line)
		y += lineHeight
	}
	return y
}
