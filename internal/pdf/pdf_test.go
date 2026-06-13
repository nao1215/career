package pdf

import (
	"bytes"
	"testing"

	"github.com/nao1215/career/internal/resume"
)

// sampleResume returns a small but complete document covering both outputs.
func sampleResume() *resume.Resume {
	return &resume.Resume{
		Date: "2026年6月13日現在",
		Profile: resume.Profile{
			Name:      "履歴書 太郎",
			NameKana:  "りれきしょ たろう",
			BirthDate: "1990年1月1日",
			Gender:    "男",
			Email:     "taro@example.com",
			Phone:     "090-0000-0000",
			Address:   resume.Address{Zip: "100-0001", Text: "東京都千代田区千代田1-1-1"},
		},
		Education: []resume.HistoryItem{
			{Year: "2009", Month: "4", Value: "見本大学 入学"},
			{Year: "2013", Month: "3", Value: "見本大学 卒業"},
		},
		Work: []resume.HistoryItem{
			{Year: "2015", Month: "4", Value: "株式会社A 入社"},
		},
		Licenses: []resume.HistoryItem{
			{Year: "2010", Month: "4", Value: "普通自動車第一種運転免許"},
		},
		Rireki: resume.Rireki{Hobby: "読書", Motivation: "技術力を高めたいため"},
		Career: resume.Career{
			Summary: "バックエンドエンジニアとして従事。",
			Skills:  []string{"Go", "AWS"},
			Histories: []resume.CareerHistory{
				{
					Company: "株式会社A",
					Period:  "2015年4月 - 現在",
					Projects: []resume.Project{
						{Title: "決済基盤", Description: "決済電文処理", Tech: []string{"Go", "AWS"}},
					},
				},
			},
			Certifications: []string{"応用情報技術者試験"},
			Publications:   []string{"Software Design 2024年12月号"},
			SelfPR:         "継続的な学習を重視しています。",
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
	got, err := RenderRirekisho(sampleResume())
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
	minimal := &resume.Resume{Profile: resume.Profile{Name: "最小 太郎"}}

	r, err := RenderRirekisho(minimal)
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
