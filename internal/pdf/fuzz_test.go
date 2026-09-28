package pdf

import (
	"encoding/hex"
	"strings"
	"testing"
	"unicode"
)

// FuzzParseHexColor checks theme.accent parsing against encoding/hex as an
// oracle: a value is accepted exactly when, after trimming and an optional
// leading '#', it is six hex digits, and the decoded bytes are the channels.
func FuzzParseHexColor(f *testing.F) {
	for _, s := range []string{
		"#1f4e79", "ff0000", "  #00ff00  ", "#fff", "#zzzzzz", "none", "",
		"+f+f+f", "0x1f4e", " 1 2 3", "#-1-1-1", "#FFFFFF", "##ffffff",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		got, err := parseHexColor(s)

		h := strings.TrimPrefix(strings.TrimSpace(s), "#")
		want, decErr := hex.DecodeString(h)
		valid := len(h) == 6 && decErr == nil

		if valid != (err == nil) {
			t.Fatalf("parseHexColor(%q) err = %v, oracle valid = %v", s, err, valid)
		}
		if valid && (got.r != want[0] || got.g != want[1] || got.b != want[2]) {
			t.Fatalf("parseHexColor(%q) = %v, want %v", s, got, want)
		}
	})
}

// nonSpace drops every Unicode space so text can be compared regardless of
// where fold and wrap moved or removed whitespace.
func nonSpace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

// FuzzFold checks the invariants of the soft-wrap normalizer: it is
// idempotent, never drops or reorders visible characters, and leaves no
// leading/trailing blank line or run of more than one blank line.
func FuzzFold(f *testing.F) {
	for _, s := range []string{
		"組み込み開発から\n始まりました。", "web application\ndevelopment", "Go を\n使う",
		"前半。\n\n\n後半。", "  あいう\n  えお", "前置き。\n・項目1\n・項目2",
		"- item\n- item2\nnot-a-bullet\n-x", "\r\n\r\n", "\n\n a \n\n", "",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		got := fold(s)
		if again := fold(got); again != got {
			t.Fatalf("fold is not idempotent:\n in    %q\n once  %q\n twice %q", s, got, again)
		}
		if nonSpace(got) != nonSpace(s) {
			t.Fatalf("fold(%q) = %q changed the visible characters", s, got)
		}
		if strings.Contains(got, "\n\n\n") || strings.HasPrefix(got, "\n") || strings.HasSuffix(got, "\n") {
			t.Fatalf("fold(%q) = %q has stray blank lines", s, got)
		}
	})
}

// FuzzWrap checks the line wrapper used for every free-text field: each line
// fits maxWidth unless it is a single character that cannot be split, and the
// lines together keep every visible character of the folded input in order.
func FuzzWrap(f *testing.F) {
	for _, s := range []string{
		strings.Repeat("あいう。えお、", 10), strings.Repeat("あい「うえ」お", 10),
		"web application development with Go and AWS", "前置き。\n・項目1\n・項目2",
		"supercalifragilisticexpialidocious", "前半。\n\n後半。", "", "   ",
	} {
		f.Add(s, uint8(30))
	}
	c, err := newCanvas()
	if err != nil {
		f.Fatalf("newCanvas() error = %v", err)
	}
	c.pdf.AddPage()
	c.setFont("mincho", 10)

	f.Fuzz(func(t *testing.T, s string, width uint8) {
		// gopdf does not cache the lookup of a glyph missing from the font, so
		// a long run of such runes costs about half a second per kilobyte.
		// Short inputs reach every branch of the wrapper while keeping the
		// fuzzer fast.
		if len(s) > 512 {
			t.Skip()
		}
		maxWidth := float64(width%200) + 1
		lines := c.wrap(s, maxWidth)

		if got, want := nonSpace(strings.Join(lines, "")), nonSpace(fold(s)); got != want {
			t.Fatalf("wrap(%q, %v) lost text: got %q, want %q", s, maxWidth, got, want)
		}
		for _, line := range lines {
			if strings.Contains(line, "\n") {
				t.Fatalf("wrap(%q, %v) returned a line with a newline: %q", s, maxWidth, line)
			}
			if len([]rune(line)) > 1 && c.textWidth(line) > maxWidth {
				t.Fatalf("wrap(%q, %v) line %q is %.2f wide", s, maxWidth, line, c.textWidth(line))
			}
		}
	})
}
