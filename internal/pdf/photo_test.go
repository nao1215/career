package pdf

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// writePNG creates a w×h PNG at path for tests.
func writePNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	f, err := os.Create(path) //nolint:gosec // test path
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestContainRect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		iw, ih         int
		wantW, wantH   float64
		wantDX, wantDY float64
	}{
		{name: "exact 3:4 fills frame", iw: 300, ih: 400, wantW: 30, wantH: 40, wantDX: 0, wantDY: 0},
		{name: "wide letterboxes vertically", iw: 400, ih: 300, wantW: 30, wantH: 22.5, wantDX: 0, wantDY: 8.75},
		{name: "tall letterboxes horizontally", iw: 300, ih: 600, wantW: 20, wantH: 40, wantDX: 5, wantDY: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w, h, dx, dy := containRect(tt.iw, tt.ih, photoBoxW, photoBoxH)
			for _, c := range []struct {
				label    string
				got, exp float64
			}{
				{"w", w, tt.wantW}, {"h", h, tt.wantH}, {"dx", dx, tt.wantDX}, {"dy", dy, tt.wantDY},
			} {
				if math.Abs(c.got-c.exp) > 0.01 {
					t.Errorf("%s = %.3f, want %.3f", c.label, c.got, c.exp)
				}
			}
		})
	}
}

func TestCheckPhoto(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	ok34 := filepath.Join(dir, "p34.png")
	writePNG(t, ok34, 300, 400)
	wide := filepath.Join(dir, "wide.png")
	writePNG(t, wide, 400, 300)

	if ok, warn := CheckPhoto(ok34); !ok || warn != "" {
		t.Errorf("CheckPhoto(3:4) = (%v, %q), want (true, \"\")", ok, warn)
	}
	if ok, warn := CheckPhoto(wide); !ok || warn == "" {
		t.Errorf("CheckPhoto(wide) = (%v, %q), want (true, non-empty)", ok, warn)
	}
	if ok, _ := CheckPhoto(filepath.Join(dir, "missing.png")); ok {
		t.Error("CheckPhoto(missing) ok = true, want false")
	}
}

func TestRenderRirekishoWithPhoto(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	photo := filepath.Join(dir, "p.png")
	writePNG(t, photo, 300, 400)

	got, err := RenderRirekisho(sampleResume(), options{photo: photo})
	if err != nil {
		t.Fatalf("RenderRirekisho() error = %v", err)
	}
	assertPDF(t, got)
}
