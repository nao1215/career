package resume

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

	const doc = `
date: 2026年6月13日現在
profile:
  name: 履歴書 太郎
  name_kana: りれきしょ たろう
  birth_date: 1990年1月1日
  gender: 男
  email: taro@example.com
  address:
    zip: 100-0001
    text: 東京都千代田区千代田1-1-1
education:
  - { year: 2009, month: 4, value: 見本大学 入学 }
  - { year: "20XX", month: 3, value: 同 卒業 }
work:
  - { year: 2015, month: 4, value: 株式会社A 入社 }
licenses:
  - { year: 2010, month: 4, value: 普通自動車第一種運転免許 }
rireki:
  hobby: 読書
  motivation: 貴社の理念に共感したため
career:
  summary: バックエンドエンジニアとして10年。
  skills:
    - Go
    - AWS
  histories:
    - company: 株式会社A
      period: 2015年4月 - 2021年12月
      projects:
        - title: 決済基盤
          description: ISO8583の決済電文処理
          tech: [Go, AWS]
  certifications:
    - 応用情報技術者試験
  publications:
    - Software Design 2024年12月号
  self_pr: 継続的な学習を重視しています。
`

	res, err := Parse(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if got, want := res.Profile.Name.For(LangJA), "履歴書 太郎"; got != want {
		t.Errorf("Name = %q, want %q", got, want)
	}
	if got, want := len(res.Education), 2; got != want {
		t.Errorf("len(Education) = %d, want %d", got, want)
	}
	// Flex accepts both numeric and string scalars.
	if got, want := res.Education[0].Year.String(), "2009"; got != want {
		t.Errorf("Education[0].Year = %q, want %q", got, want)
	}
	if got, want := res.Education[1].Year.String(), "20XX"; got != want {
		t.Errorf("Education[1].Year = %q, want %q", got, want)
	}
	if got, want := len(res.Career.Skills), 2; got != want {
		t.Errorf("len(Skills) = %d, want %d", got, want)
	}
	if got, want := res.Career.Histories[0].Projects[0].Tech[0], "Go"; got != want {
		t.Errorf("tech[0] = %q, want %q", got, want)
	}
}

func TestTextLocalization(t *testing.T) {
	t.Parallel()

	const doc = `
profile:
  name:
    ja: 見本 太郎
    en: Taro Mihon
career:
  summary: shared summary
`
	res, err := Parse(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := res.Profile.Name.For(LangJA); got != "見本 太郎" {
		t.Errorf("Name.For(ja) = %q", got)
	}
	if got := res.Profile.Name.For(LangEN); got != "Taro Mihon" {
		t.Errorf("Name.For(en) = %q", got)
	}
	// A scalar applies to every language.
	if got := res.Career.Summary.For(LangEN); got != "shared summary" {
		t.Errorf("Summary.For(en) = %q, want fallback to the scalar value", got)
	}
}

// TestTextFallbackIsDeterministic guards the regression where a text written
// only in languages the templates do not ask for (neither ja nor en) printed a
// randomly chosen language on each run because the fallback walked a map.
func TestTextFallbackIsDeterministic(t *testing.T) {
	t.Parallel()

	const doc = `
profile:
  name:
    fr: Jean
    de: Hans
    it: Gianni
`
	res, err := Parse(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	for range 50 {
		if got := res.Profile.Name.For(LangJA); got != "Hans" {
			t.Fatalf("Name.For(ja) = %q, want %q (first language in key order)", got, "Hans")
		}
	}
}

func TestParseUnknownField(t *testing.T) {
	t.Parallel()

	const doc = `
profile:
  name: テスト
  nonexistent: oops
`
	if _, err := Parse(strings.NewReader(doc)); err == nil {
		t.Fatal("Parse() error = nil, want error for unknown field")
	}
}

func TestParseEmpty(t *testing.T) {
	t.Parallel()

	if _, err := Parse(strings.NewReader("")); err == nil {
		t.Fatal("Parse() error = nil, want error for empty document")
	}
}

func TestValidateRireki(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		res     Resume
		wantErr error
	}{
		{
			name:    "missing name",
			res:     Resume{},
			wantErr: ErrEmptyName,
		},
		{
			name:    "whitespace-only name",
			res:     Resume{Profile: Profile{Name: Plain("   ")}},
			wantErr: ErrEmptyName,
		},
		{
			name: "ok",
			res:  Resume{Profile: Profile{Name: Plain("x")}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.res.ValidateRireki()
			if tt.wantErr == nil && err != nil {
				t.Fatalf("ValidateRireki() = %v, want nil", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateRireki() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCareer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		res     Resume
		wantErr bool
	}{
		{
			name:    "missing name",
			res:     Resume{Career: Career{Summary: Plain("x")}},
			wantErr: true,
		},
		{
			name:    "no content",
			res:     Resume{Profile: Profile{Name: Plain("x")}},
			wantErr: true,
		},
		{
			name:    "whitespace-only summary, no history",
			res:     Resume{Profile: Profile{Name: Plain("x")}, Career: Career{Summary: Plain("   ")}},
			wantErr: true,
		},
		{
			name: "summary only",
			res:  Resume{Profile: Profile{Name: Plain("x")}, Career: Career{Summary: Plain("x")}},
		},
		{
			name: "history only",
			res:  Resume{Profile: Profile{Name: Plain("x")}, Career: Career{Histories: []CareerHistory{{Company: Plain("A")}}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.res.ValidateCareer()
			if tt.wantErr && err == nil {
				t.Fatal("ValidateCareer() = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ValidateCareer() = %v, want nil", err)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "r.yaml")
	if err := os.WriteFile(path, []byte("profile:\n  name: ロード太郎\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	res, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := res.Profile.Name.For(""); got != "ロード太郎" {
		t.Errorf("Name = %q", got)
	}

	if _, err := Load(filepath.Join(dir, "missing.yaml")); err == nil {
		t.Error("Load(missing) error = nil, want error")
	}
}

// texts collects every localized field of a parsed document.
func texts(res *Resume) []Text {
	out := []Text{res.Date, res.Profile.Name, res.Profile.Address.Text, res.Profile.Contact.Text}
	for _, list := range [][]HistoryItem{res.Education, res.Work, res.Licenses} {
		for _, it := range list {
			out = append(out, it.Value)
		}
	}
	c := res.Career
	out = append(out, c.Summary, c.SelfPR)
	out = append(out, c.Skills...)
	out = append(out, c.Certifications...)
	out = append(out, c.Publications...)
	out = append(out, c.Links...)
	for _, h := range c.Histories {
		out = append(out, h.Company, h.Period, h.Role, h.Summary)
		for _, p := range h.Projects {
			out = append(out, p.Title, p.Period, p.Role, p.Description)
		}
	}
	return out
}

// FuzzParse feeds arbitrary bytes to the resume YAML parser. Parsing may fail
// but must not panic, and a parsed document must render the same text every
// time: Text.For is deterministic for every language a template asks for, and
// Has agrees with the text it would print.
func FuzzParse(f *testing.F) {
	for _, path := range []string{
		"../../examples/resume.yaml", "../../examples/minimal.yaml", "../../cmd/templates/starter.yaml",
	} {
		b, err := os.ReadFile(path) //nolint:gosec // fixed seed paths
		if err != nil {
			f.Fatal(err)
		}
		f.Add(b)
	}
	for _, s := range []string{
		"", "profile:\n  name: テスト\n  nonexistent: oops\n",
		"profile:\n  name:\n    ja: 見本 太郎\n    en: Taro Mihon\n",
		"profile:\n  name:\n    fr: Jean\n    de: Hans\n",
		"education:\n  - { year: 2009, month: [4], value: {ja: x} }\n",
		"career: &a\n  summary: *a\n", "profile: [1, 2]\n",
	} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		res, err := Parse(bytes.NewReader(data))
		if err != nil {
			if res != nil {
				t.Fatalf("Parse returned a document with error %v", err)
			}
			return
		}
		for _, txt := range texts(res) {
			for _, lang := range []string{"", LangJA, LangEN, "fr"} {
				first := txt.For(lang)
				for range 8 {
					if got := txt.For(lang); got != first {
						t.Fatalf("Text.For(%q) is not deterministic: %q then %q", lang, first, got)
					}
				}
			}
			if got, want := txt.Has(), strings.TrimSpace(txt.For("")) != ""; got != want {
				t.Fatalf("Text.Has() = %v, but For(\"\") = %q", got, txt.For(""))
			}
		}
	})
}
