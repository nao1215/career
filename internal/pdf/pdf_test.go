package pdf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nao1215/career/internal/resume"
)

// sampleResume returns a small but complete document covering every output.
func sampleResume() *resume.Resume {
	p := resume.Plain
	return &resume.Resume{
		Date: p("2026年6月13日現在"),
		Profile: resume.Profile{
			Name:      p("履歴書 太郎"),
			NameKana:  "りれきしょ たろう",
			BirthDate: "1990年1月1日",
			Gender:    "男",
			Email:     "taro@example.com",
			Phone:     "090-0000-0000",
			Address:   resume.Address{Zip: "100-0001", Text: p("東京都千代田区千代田1-1-1")},
		},
		Education: []resume.HistoryItem{
			{Year: "2009", Month: "4", Value: p("見本大学 入学")},
			{Year: "2013", Month: "3", Value: p("見本大学 卒業")},
		},
		Work: []resume.HistoryItem{
			{Year: "2015", Month: "4", Value: p("株式会社A 入社")},
		},
		Licenses: []resume.HistoryItem{
			{Year: "2010", Month: "4", Value: p("普通自動車第一種運転免許")},
		},
		Rireki: resume.Rireki{Hobby: "読書", Motivation: "技術力を高めたいため"},
		Career: resume.Career{
			Summary: p("バックエンドエンジニアとして従事。"),
			Skills:  []resume.Text{p("Go"), p("AWS")},
			Histories: []resume.CareerHistory{
				{
					Company: p("株式会社A"),
					Period:  p("2015年4月 - 現在"),
					Projects: []resume.Project{
						{Title: p("決済基盤"), Description: p("決済電文処理"), Tech: []string{"Go", "AWS"}},
					},
				},
			},
			Certifications: []resume.Text{p("応用情報技術者試験")},
			Publications:   []resume.Text{p("Software Design 2024年12月号")},
			SelfPR:         p("継続的な学習を重視しています。"),
		},
	}
}

func assertPDF(t *testing.T, got []byte) {
	t.Helper()
	if len(got) == 0 {
		t.Fatal("rendered PDF is empty")
	}
	if !bytes.HasPrefix(got, []byte("%PDF-")) {
		t.Fatalf("output does not start with PDF header: %q", got[:min(8, len(got))])
	}
	if !bytes.Contains(got, []byte("%%EOF")) {
		t.Error("output is missing the EOF trailer")
	}
}

func TestRenderRirekisho(t *testing.T) {
	t.Parallel()
	got, err := RenderRirekisho(sampleResume(), options{lang: resume.LangJA})
	if err != nil {
		t.Fatalf("RenderRirekisho() error = %v", err)
	}
	assertPDF(t, got)
}

// pageCount renders the 履歴書 and reports how many pages it produced.
func pageCount(t *testing.T, res *resume.Resume) int {
	t.Helper()
	c, err := newCanvas()
	if err != nil {
		t.Fatalf("newCanvas() error = %v", err)
	}
	(&rirekishoRenderer{c: c, res: res, lang: resume.LangJA}).render()
	return c.pdf.GetNumberOfPages()
}

// TestRirekishoPaginatesWithoutDropping guards against the regression where rows
// beyond a single page were silently discarded. A long history must spill onto
// extra pages rather than being truncated.
func TestRirekishoPaginatesWithoutDropping(t *testing.T) {
	t.Parallel()

	small := sampleResume()
	if got := pageCount(t, small); got != 2 {
		t.Fatalf("small resume pages = %d, want 2", got)
	}

	big := sampleResume()
	for i := 0; i < 40; i++ {
		big.Work = append(big.Work, resume.HistoryItem{
			Year:  resume.Flex("2010"),
			Month: resume.Flex("4"),
			Value: resume.Plain("会社の職歴エントリ"),
		})
	}
	for i := 0; i < 40; i++ {
		big.Licenses = append(big.Licenses, resume.HistoryItem{
			Year:  resume.Flex("2011"),
			Month: resume.Flex("5"),
			Value: resume.Plain("資格エントリ"),
		})
	}
	// 80+ extra rows cannot fit on the two pages a short resume uses; the
	// renderer must add pages instead of dropping rows.
	if got := pageCount(t, big); got <= 2 {
		t.Fatalf("large resume pages = %d, want > 2 (rows must not be dropped)", got)
	}
}

// TestRirekishoFreeTextDoesNotOverflow guards the regression where a long
// free-text field overran its box and the page. A very long 志望動機 must push
// content onto additional pages rather than overlapping the next field.
func TestRirekishoFreeTextDoesNotOverflow(t *testing.T) {
	t.Parallel()

	res := sampleResume()
	res.Rireki.Motivation = strings.Repeat("長い志望動機のテキストをここに書き連ねます。", 120)

	if got := pageCount(t, res); got <= 2 {
		t.Fatalf("resume with a long free-text field pages = %d, want > 2", got)
	}
}

// TestRirekishoLicensesPaginate guards the second High regression: a long
// 免許・資格 list must spill onto extra pages, and because render() always draws
// the free-text fields after it, those fields cannot be pushed off the document.
func TestRirekishoLicensesPaginate(t *testing.T) {
	t.Parallel()

	res := sampleResume()
	for i := 0; i < 60; i++ {
		res.Licenses = append(res.Licenses, resume.HistoryItem{
			Year:  resume.Flex("2011"),
			Month: resume.Flex("5"),
			Value: resume.Plain("資格エントリ"),
		})
	}
	if got := pageCount(t, res); got <= 2 {
		t.Fatalf("resume with many licenses pages = %d, want > 2", got)
	}
}

// TestRirekishoOverlongCellRenders ensures a single very long table cell renders
// without error; the value is truncated to its cell width (covered in detail by
// TestTruncateToWidth) rather than crossing the page.
func TestRirekishoOverlongCellRenders(t *testing.T) {
	t.Parallel()

	res := sampleResume()
	res.Work = append(res.Work, resume.HistoryItem{
		Year:  resume.Flex("2016"),
		Month: resume.Flex("4"),
		Value: resume.Plain(strings.Repeat("非常に長い職歴の説明", 40)),
	})
	got, err := RenderRirekisho(res, options{lang: resume.LangJA})
	if err != nil {
		t.Fatalf("RenderRirekisho() error = %v", err)
	}
	assertPDF(t, got)
}

func TestRenderShokumukeirekisho(t *testing.T) {
	t.Parallel()
	got, err := RenderShokumukeirekisho(sampleResume(), options{accent: defaultAccent, accentOn: true})
	if err != nil {
		t.Fatalf("RenderShokumukeirekisho() error = %v", err)
	}
	assertPDF(t, got)
}

func TestRenderCV(t *testing.T) {
	t.Parallel()
	// Cover both the accented and the monochrome paths.
	for _, opts := range []options{
		{accent: defaultAccent, accentOn: true},
		{accentOn: false},
	} {
		got, err := RenderCV(sampleResume(), opts)
		if err != nil {
			t.Fatalf("RenderCV() error = %v", err)
		}
		assertPDF(t, got)
	}
}

// TestRenderMinimal ensures rendering does not panic on a document that only has
// the universally required name field.
func TestRenderMinimal(t *testing.T) {
	t.Parallel()
	minimal := &resume.Resume{Profile: resume.Profile{Name: resume.Plain("最小 太郎")}}

	r, err := RenderRirekisho(minimal, options{lang: resume.LangJA})
	if err != nil {
		t.Fatalf("RenderRirekisho(minimal) error = %v", err)
	}
	assertPDF(t, r)

	s, err := RenderShokumukeirekisho(minimal, options{})
	if err != nil {
		t.Fatalf("RenderShokumukeirekisho(minimal) error = %v", err)
	}
	assertPDF(t, s)

	v, err := RenderCV(minimal, options{})
	if err != nil {
		t.Fatalf("RenderCV(minimal) error = %v", err)
	}
	assertPDF(t, v)
}

// TestWrapBreaksLongText checks the line wrapper splits text with no spaces,
// which is the common case for Japanese.
func TestWrapBreaksLongText(t *testing.T) {
	t.Parallel()
	c, err := newCanvas()
	if err != nil {
		t.Fatalf("newCanvas() error = %v", err)
	}
	c.pdf.AddPage()
	c.setFont("mincho", 10)

	long := ""
	for range 200 {
		long += "あ"
	}
	lines := c.wrap(long, 100)
	if len(lines) < 2 {
		t.Fatalf("wrap() returned %d lines, want >= 2", len(lines))
	}
}

// TestWrapLineStartKinsoku checks 行頭禁則: no wrapped line may begin with a
// closing bracket or sentence punctuation.
func TestWrapLineStartKinsoku(t *testing.T) {
	t.Parallel()
	c, err := newCanvas()
	if err != nil {
		t.Fatalf("newCanvas() error = %v", err)
	}
	c.pdf.AddPage()
	c.setFont("mincho", 10)

	s := strings.Repeat("あいう。えお、", 60)
	lines := c.wrap(s, 30)
	if len(lines) < 2 {
		t.Fatalf("wrap() returned %d lines, want >= 2", len(lines))
	}
	for _, line := range lines {
		r := []rune(line)
		if len(r) > 0 && isLineStartProhibited(r[0]) {
			t.Errorf("line begins with prohibited rune %q: %q", string(r[0]), line)
		}
	}
}

// TestWrapLineEndKinsoku checks 行末禁則: no wrapped line may end with an opening
// bracket.
func TestWrapLineEndKinsoku(t *testing.T) {
	t.Parallel()
	c, err := newCanvas()
	if err != nil {
		t.Fatalf("newCanvas() error = %v", err)
	}
	c.pdf.AddPage()
	c.setFont("mincho", 10)

	s := strings.Repeat("あい「うえ」お", 60)
	lines := c.wrap(s, 30)
	if len(lines) < 2 {
		t.Fatalf("wrap() returned %d lines, want >= 2", len(lines))
	}
	for _, line := range lines {
		r := []rune(line)
		if len(r) > 0 && isLineEndProhibited(r[len(r)-1]) {
			t.Errorf("line ends with prohibited rune %q: %q", string(r[len(r)-1]), line)
		}
	}
}
