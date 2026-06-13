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
	pdf     *gopdf.GoPdf
	curFont string
	curSize float64
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

// textWidth measures the rendered width of s at the current font.
func (c *canvas) textWidth(s string) float64 {
	w, err := c.pdf.MeasureTextWidth(s)
	if err != nil {
		return 0
	}
	return w
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
// newlines are honoured, and long runs are broken per rune because Japanese
// text has no spaces to break on.
func (c *canvas) wrap(s string, maxWidth float64) []string {
	var lines []string
	for _, paragraph := range strings.Split(s, "\n") {
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}
		var line []rune
		for _, r := range paragraph {
			candidate := string(append(line, r))
			if c.textWidth(candidate) > maxWidth && len(line) > 0 {
				lines = append(lines, string(line))
				line = []rune{r}
				continue
			}
			line = append(line, r)
		}
		lines = append(lines, string(line))
	}
	return lines
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
