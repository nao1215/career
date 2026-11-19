package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitWritesStarter(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	app, stdout, stderr := newTestApp(dir)
	if code := app.Run([]string{"init"}); code != 0 {
		t.Fatalf("init exit = %d, want 0 (stderr=%q)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "wrote resume.yaml") {
		t.Errorf("stdout = %q", stdout.String())
	}

	data, err := os.ReadFile(filepath.Join(dir, "resume.yaml")) //nolint:gosec // test path
	if err != nil {
		t.Fatalf("starter not written: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "profile:") {
		t.Error("starter is missing the profile section")
	}
	// The date placeholder must be replaced with a real date at write time.
	if strings.Contains(body, "__DATE__") {
		t.Error("starter still contains the __DATE__ placeholder")
	}
	if !strings.Contains(body, "年") || !strings.Contains(body, "現在") {
		t.Errorf("starter date was not filled in: %q", body)
	}
	// The guidance must use the current canonical template name, not the old one.
	if !strings.Contains(body, "work-history") {
		t.Error("starter guidance does not mention the work-history template")
	}
	if strings.Contains(body, "career-history") {
		t.Error("starter guidance still references the old career-history name")
	}
}

func TestInitCustomPath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	app, _, stderr := newTestApp(dir)
	if code := app.Run([]string{"init", "cv-source.yaml"}); code != 0 {
		t.Fatalf("init exit = %d, want 0 (stderr=%q)", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "cv-source.yaml")); err != nil {
		t.Errorf("custom path not written: %v", err)
	}
}

func TestInitNoOverwriteWithoutForce(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "resume.yaml")
	if err := os.WriteFile(path, []byte("keep me\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app, _, stderr := newTestApp(dir)
	if code := app.Run([]string{"init"}); code != 1 {
		t.Fatalf("init exit = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "already exists") {
		t.Errorf("stderr = %q", stderr.String())
	}
	data, _ := os.ReadFile(path) //nolint:gosec // test path
	if string(data) != "keep me\n" {
		t.Error("init overwrote an existing file without --force")
	}
}

func TestInitForceOverwrites(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "resume.yaml")
	if err := os.WriteFile(path, []byte("old\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app, _, stderr := newTestApp(dir)
	if code := app.Run([]string{"init", "--force"}); code != 0 {
		t.Fatalf("init --force exit = %d, want 0 (stderr=%q)", code, stderr.String())
	}
	data, _ := os.ReadFile(path) //nolint:gosec // test path
	if string(data) == "old\n" {
		t.Error("init --force did not overwrite the file")
	}
}

// TestInitOutputGenerates checks the scaffold init writes is itself valid input
// for generate.
func TestInitOutputGenerates(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	app, _, stderr := newTestApp(dir)
	if code := app.Run([]string{"init"}); code != 0 {
		t.Fatalf("init exit = %d", code)
	}
	for _, tmpl := range []string{"cv", "japanese-resume", "work-history"} {
		out := tmpl + ".pdf"
		if code := app.Run([]string{"generate", "resume.yaml", "-t", tmpl, "-o", out}); code != 0 {
			t.Fatalf("generate %s exit = %d (stderr=%q)", tmpl, code, stderr.String())
		}
		assertPDFFile(t, filepath.Join(dir, out))
	}
}
