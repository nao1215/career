package resume

import (
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
