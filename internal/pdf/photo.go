package pdf

import (
	"fmt"
	"image"
	"math"
	"os"

	// Register the decoders for the portrait formats we accept.
	_ "image/jpeg"
	_ "image/png"
)

// The JIS 履歴書 photo frame is 30mm wide by 40mm tall (a 3:4 portrait).
const (
	photoBoxW   = 30.0
	photoBoxH   = 40.0
	photoAspect = photoBoxW / photoBoxH // 0.75
	// aspectTolerance is how far a photo's width/height ratio may stray from 3:4
	// before we warn the user.
	aspectTolerance = 0.03
)

// photoDims returns the pixel width and height of the image at path.
func photoDims(path string) (w, h int, err error) {
	f, err := os.Open(path) //nolint:gosec // path is a user-supplied portrait by design
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}

// containRect fits an imgW×imgH image inside a boxW×boxH frame without
// distortion, returning the drawn size and the top-left offset that centers it.
func containRect(imgW, imgH int, boxW, boxH float64) (w, h, dx, dy float64) {
	if imgW <= 0 || imgH <= 0 {
		return boxW, boxH, 0, 0
	}
	ar := float64(imgW) / float64(imgH)
	if ar > boxW/boxH {
		w, h = boxW, boxW/ar // limited by width
	} else {
		w, h = boxH*ar, boxH // limited by height
	}
	return w, h, (boxW - w) / 2, (boxH - h) / 2
}

// CheckPhoto inspects a portrait file before rendering. ok is false when the
// file cannot be read or decoded, so the caller can fall back to the placeholder.
// warn is a non-empty message when the image is readable but its aspect ratio
// differs from the 3:4 履歴書 frame.
func CheckPhoto(path string) (ok bool, warn string) {
	w, h, err := photoDims(path)
	if err != nil {
		return false, ""
	}
	if h == 0 {
		return false, ""
	}
	ar := float64(w) / float64(h)
	if math.Abs(ar-photoAspect) > aspectTolerance {
		return true, fmt.Sprintf(
			"photo is %dx%d (ratio %.2f); the 履歴書 frame is 3:4 (0.75), so it will be centered with margins — crop to 3:4 for a full frame",
			w, h, ar)
	}
	return true, ""
}
