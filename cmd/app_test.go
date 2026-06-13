package cmd

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleYAML = `date: 2026年6月13日現在
profile:
  name: 履歴書 太郎
  birth_date: 1990年1月1日
  gender: 男
education:
  - { year: 2009, month: 4, value: 見本大学 入学 }
work:
  - { year: 2015, month: 4, value: 株式会社A 入社 }
licenses:
  - { year: 2010, month: 4, value: 普通自動車第一種運転免許 }
career:
  summary: バックエンドエンジニアとして従事。
  skills: [Go, AWS]
  self_pr: 学習を継続しています。
`

// newTestApp returns an App writing to in-memory buffers rooted at dir.
func newTestApp(dir string) (*App, *bytes.Buffer, *bytes.Buffer) {
	var stdout, stderr bytes.Buffer
	app := NewApp(&stdout, &stderr, strings.NewReader(""), dir)
	return app, &stdout, &stderr
}

func writeSample(t *testing.T, dir string) {
	t.Helper()
	path := filepath.Join(dir, "resume.yaml")
	if err := os.WriteFile(path, []byte(sampleYAML), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestRunNoArgs(t *testing.T) {
	t.Parallel()
	app, stdout, _ := newTestApp(t.TempDir())
	if code := app.Run(nil); code != 0 {
		t.Fatalf("Run(nil) = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Errorf("root help missing Usage, got %q", stdout.String())
	}
}

func TestRunVersion(t *testing.T) {
	t.Parallel()
	app, stdout, _ := newTestApp(t.TempDir())
	if code := app.Run([]string{"version"}); code != 0 {
		t.Fatalf("version exit = %d, want 0", code)
	}
	if !strings.HasPrefix(stdout.String(), "career ") {
		t.Errorf("version output = %q", stdout.String())
	}
}

func TestRunTemplates(t *testing.T) {
	t.Parallel()
	app, stdout, _ := newTestApp(t.TempDir())
	if code := app.Run([]string{"templates"}); code != 0 {
		t.Fatalf("templates exit = %d, want 0", code)
	}
	out := stdout.String()
	for _, want := range []string{"cv", "japanese-resume", "career-history", "aliases"} {
		if !strings.Contains(out, want) {
			t.Errorf("templates output missing %q", want)
		}
	}
}

func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()
	app, _, stderr := newTestApp(t.TempDir())
	if code := app.Run([]string{"frobnicate"}); code != 1 {
		t.Fatalf("unknown command exit = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Errorf("stderr = %q", stderr.String())
	}
}

func TestGenerateRirekisho(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSample(t, dir)

	app, stdout, stderr := newTestApp(dir)
	code := app.Run([]string{"generate", "resume.yaml", "-t", "japanese-resume", "-o", "out.pdf"})
	if code != 0 {
		t.Fatalf("generate exit = %d, want 0 (stderr=%q)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "wrote out.pdf") {
		t.Errorf("stdout = %q", stdout.String())
	}
	assertPDFFile(t, filepath.Join(dir, "out.pdf"))
}

func TestGenerateCareerHistoryDefaultOutput(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSample(t, dir)

	app, _, stderr := newTestApp(dir)
	// Input via --input, no --output so the default name is used.
	code := app.Run([]string{"generate", "--input", "resume.yaml", "--template", "career-history"})
	if code != 0 {
		t.Fatalf("generate exit = %d, want 0 (stderr=%q)", code, stderr.String())
	}
	assertPDFFile(t, filepath.Join(dir, "career-history.pdf"))
}

func TestGenerateCV(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSample(t, dir)

	app, _, stderr := newTestApp(dir)
	// Default template is cv; also exercise the --accent flag.
	code := app.Run([]string{"generate", "resume.yaml", "--accent", "#2c6e6e", "-o", "cv.pdf"})
	if code != 0 {
		t.Fatalf("generate exit = %d, want 0 (stderr=%q)", code, stderr.String())
	}
	assertPDFFile(t, filepath.Join(dir, "cv.pdf"))
}

func TestGenerateAll(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSample(t, dir)

	app, stdout, stderr := newTestApp(dir)
	if code := app.Run([]string{"generate", "resume.yaml", "-t", "all"}); code != 0 {
		t.Fatalf("generate -t all exit = %d (stderr=%q)", code, stderr.String())
	}
	for _, name := range []string{"cv.pdf", "japanese-resume.pdf", "career-history.pdf"} {
		assertPDFFile(t, filepath.Join(dir, name))
	}
	if n := strings.Count(stdout.String(), "wrote "); n != 3 {
		t.Errorf("wrote count = %d, want 3 (%q)", n, stdout.String())
	}
}

func TestGenerateMultiple(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSample(t, dir)

	app, _, stderr := newTestApp(dir)
	// Repeated and comma-separated values combine; duplicates collapse.
	if code := app.Run([]string{"generate", "resume.yaml", "-t", "cv,career-history", "-t", "cv"}); code != 0 {
		t.Fatalf("generate exit = %d (stderr=%q)", code, stderr.String())
	}
	assertPDFFile(t, filepath.Join(dir, "cv.pdf"))
	assertPDFFile(t, filepath.Join(dir, "career-history.pdf"))
	if _, err := os.Stat(filepath.Join(dir, "japanese-resume.pdf")); err == nil {
		t.Error("japanese-resume.pdf was generated but not requested")
	}
}

func TestGenerateOutputWithMultipleFails(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSample(t, dir)

	app, _, stderr := newTestApp(dir)
	if code := app.Run([]string{"generate", "resume.yaml", "-t", "all", "-o", "x.pdf"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "--output cannot be used with multiple") {
		t.Errorf("stderr = %q", stderr.String())
	}
}

// writeTestPNG creates a w×h PNG for photo tests.
func writeTestPNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	f, err := os.Create(path) //nolint:gosec // test path
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestGeneratePhotoFlag(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSample(t, dir)
	writeTestPNG(t, filepath.Join(dir, "face.png"), 300, 400)

	app, _, stderr := newTestApp(dir)
	code := app.Run([]string{"generate", "resume.yaml", "-t", "japanese-resume", "--photo", "face.png", "-o", "out.pdf"})
	if code != 0 {
		t.Fatalf("generate exit = %d (stderr=%q)", code, stderr.String())
	}
	assertPDFFile(t, filepath.Join(dir, "out.pdf"))
	if stderr.Len() != 0 {
		t.Errorf("unexpected stderr for a 3:4 photo: %q", stderr.String())
	}
}

func TestGeneratePhotoWrongAspectWarns(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSample(t, dir)
	writeTestPNG(t, filepath.Join(dir, "wide.png"), 400, 300)

	app, _, stderr := newTestApp(dir)
	code := app.Run([]string{"generate", "resume.yaml", "-t", "japanese-resume", "--photo", "wide.png", "-o", "out.pdf"})
	if code != 0 {
		t.Fatalf("generate exit = %d (stderr=%q)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "3:4") {
		t.Errorf("expected an aspect-ratio warning, got %q", stderr.String())
	}
	assertPDFFile(t, filepath.Join(dir, "out.pdf"))
}

func TestGeneratePhotoMissingWarns(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSample(t, dir)

	app, _, stderr := newTestApp(dir)
	code := app.Run([]string{"generate", "resume.yaml", "-t", "japanese-resume", "--photo", "nope.png", "-o", "out.pdf"})
	if code != 0 {
		t.Fatalf("generate exit = %d (stderr=%q)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "cannot read photo") {
		t.Errorf("expected a missing-photo warning, got %q", stderr.String())
	}
	// Still produces the PDF with a placeholder.
	assertPDFFile(t, filepath.Join(dir, "out.pdf"))
}

func TestGenerateCVIgnoresPhoto(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSample(t, dir)
	writeTestPNG(t, filepath.Join(dir, "wide.png"), 400, 300)

	app, _, stderr := newTestApp(dir)
	// cv does not use a photo, so even a bad-aspect photo must not warn.
	code := app.Run([]string{"generate", "resume.yaml", "-t", "cv", "--photo", "wide.png", "-o", "cv.pdf"})
	if code != 0 {
		t.Fatalf("generate exit = %d (stderr=%q)", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("cv should ignore the photo, got stderr %q", stderr.String())
	}
}

func TestGenerateInvalidAccent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSample(t, dir)

	app, _, stderr := newTestApp(dir)
	if code := app.Run([]string{"generate", "resume.yaml", "-t", "cv", "--accent", "bogus"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "hex color") {
		t.Errorf("stderr = %q", stderr.String())
	}
}

func TestGenerateMissingInput(t *testing.T) {
	t.Parallel()
	app, _, stderr := newTestApp(t.TempDir())
	if code := app.Run([]string{"generate", "-t", "japanese-resume"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "no input file") {
		t.Errorf("stderr = %q", stderr.String())
	}
}

func TestGenerateUnknownTemplate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSample(t, dir)
	app, _, stderr := newTestApp(dir)
	if code := app.Run([]string{"generate", "resume.yaml", "-t", "bogus"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "unknown template") {
		t.Errorf("stderr = %q", stderr.String())
	}
}

func TestGenerateFileNotFound(t *testing.T) {
	t.Parallel()
	app, _, stderr := newTestApp(t.TempDir())
	if code := app.Run([]string{"generate", "missing.yaml", "-t", "japanese-resume"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if stderr.Len() == 0 {
		t.Error("expected an error message on stderr")
	}
}

func TestGenerateValidationError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// A document with no name fails validation for either template.
	path := filepath.Join(dir, "empty.yaml")
	if err := os.WriteFile(path, []byte("date: x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	app, _, stderr := newTestApp(dir)
	if code := app.Run([]string{"generate", "empty.yaml", "-t", "japanese-resume"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "name is required") {
		t.Errorf("stderr = %q", stderr.String())
	}
}

func TestHelpForGenerate(t *testing.T) {
	t.Parallel()
	app, stdout, _ := newTestApp(t.TempDir())
	if code := app.Run([]string{"help", "generate"}); code != 0 {
		t.Fatalf("help generate exit = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "--template") {
		t.Errorf("help output = %q", stdout.String())
	}
}

func assertPDFFile(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Errorf("%s is not a PDF", path)
	}
}
