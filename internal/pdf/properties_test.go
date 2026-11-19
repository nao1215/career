package pdf

// Generated properties and metamorphic relations for the text layout that
// every free-text field goes through: fold (soft-wrap normalization) and wrap
// (line breaking with 禁則処理). FuzzFold and FuzzWrap check each call on its
// own; these relate two calls whose outputs must agree. A failure prints the
// seed to replay it.

import (
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// A small alphabet that mixes Latin words, spaces, kana, ideographs, the
// punctuation and brackets 禁則処理 handles, list bullets and newlines, so
// break opportunities, prohibited line starts and paragraph gaps meet often.
func genLayoutText() *rapid.Generator[string] {
	pieces := []string{
		"a", "go", "web", " ", "あ", "い", "漢", "字", "。", "、", "「", "」",
		"ー", "っ", "（", "）", "・", "- ", "\n", "\n\n",
	}
	return rapid.Custom(func(t *rapid.T) string {
		parts := rapid.SliceOfN(rapid.SampledFrom(pieces), 0, 40).Draw(t, "parts")
		return strings.Join(parts, "")
	})
}

// Folding two paragraphs separately and joining them with a blank line gives
// the same text as folding them joined: a blank line is a paragraph boundary
// that nothing on either side can move.
func TestFold_ParagraphsFoldIndependently(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		a := fold(genLayoutText().Draw(t, "a"))
		b := fold(genLayoutText().Draw(t, "b"))
		if a == "" || b == "" {
			t.Skip("an empty side has no paragraph to keep apart")
		}
		if got, want := fold(a+"\n\n"+b), a+"\n\n"+b; got != want {
			t.Fatalf("fold(a + blank line + b) = %q, want %q", got, want)
		}
	})
}

// A wider column never needs more lines: every break that fits the narrow
// column also fits the wide one, so greedy breaking can only merge lines.
func TestWrap_WiderColumnNeverNeedsMoreLines(t *testing.T) {
	t.Parallel()
	c, err := newCanvas()
	if err != nil {
		t.Fatalf("newCanvas() error = %v", err)
	}
	c.pdf.AddPage()
	c.setFont("mincho", 10)
	rapid.Check(t, func(t *rapid.T) {
		s := genLayoutText().Draw(t, "s")
		narrow := rapid.Float64Range(1, 150).Draw(t, "narrow")
		wide := narrow + rapid.Float64Range(0, 150).Draw(t, "extra")

		nLines, wLines := c.wrap(s, narrow), c.wrap(s, wide)
		if len(wLines) > len(nLines) {
			t.Fatalf("wrap(%q) has %d lines at width %.2f but %d at the wider %.2f:\n narrow %q\n wide   %q",
				s, len(nLines), narrow, len(wLines), wide, nLines, wLines)
		}
	})
}

// Wrapping the output of wrap again at the same width changes nothing: each
// line already fits, so a second pass finds no line to break.
func TestWrap_RewrappingEachLineIsStable(t *testing.T) {
	t.Parallel()
	c, err := newCanvas()
	if err != nil {
		t.Fatalf("newCanvas() error = %v", err)
	}
	c.pdf.AddPage()
	c.setFont("mincho", 10)
	rapid.Check(t, func(t *rapid.T) {
		s := genLayoutText().Draw(t, "s")
		width := rapid.Float64Range(1, 200).Draw(t, "width")
		for _, line := range c.wrap(s, width) {
			if line == "" {
				continue
			}
			again := c.wrapParagraph(line, width)
			if len(again) != 1 || again[0] != line {
				t.Fatalf("line %q of wrap(%q, %.2f) wraps again into %q", line, s, width, again)
			}
		}
	})
}
