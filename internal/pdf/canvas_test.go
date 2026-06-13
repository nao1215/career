package pdf

import (
	"strings"
	"testing"

	"github.com/nao1215/career/internal/font"
)

// newTestCanvas returns a canvas with a page and the Mincho font selected.
func newTestCanvas(t *testing.T) *canvas {
	t.Helper()
	c, err := newCanvas()
	if err != nil {
		t.Fatalf("newCanvas() error = %v", err)
	}
	c.pdf.AddPage()
	c.setFont(font.Mincho, 11)
	return c
}

// TestTruncateToWidth guards the regression where an over-long table cell or
// personal-block value crossed its border. The result must fit the limit and be
// marked with an ellipsis.
func TestTruncateToWidth(t *testing.T) {
	t.Parallel()
	c := newTestCanvas(t)

	const maxWidth = 40.0
	long := strings.Repeat("あ", 200)
	got := c.truncateToWidth(long, maxWidth)

	if w := c.textWidth(got); w > maxWidth {
		t.Errorf("truncated width = %.2f, want <= %.2f", w, maxWidth)
	}
	if !strings.HasSuffix(got, "…") {
		t.Errorf("truncated text %q should end with an ellipsis", got)
	}

	// A short string that already fits is returned unchanged.
	if got := c.truncateToWidth("あ", maxWidth); got != "あ" {
		t.Errorf("short string changed: %q", got)
	}
}

// TestTextFitRestoresFont guards the regression where a shrink inside textFit
// leaked the smaller size into later draws.
func TestTextFitRestoresFont(t *testing.T) {
	t.Parallel()
	c := newTestCanvas(t)
	c.setFont(font.Mincho, 11)

	// A long string in a narrow cell forces textFit to shrink.
	c.textFit(10, 10, 5, strings.Repeat("職歴", 50), 11, 7)

	if c.curSize != 11 {
		t.Errorf("font size after textFit = %.1f, want 11 (must not leak)", c.curSize)
	}
}
