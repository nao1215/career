package pdf

import (
	"fmt"
	"strings"
)

// rgb is a 24-bit color.
type rgb struct {
	r, g, b uint8
}

var (
	// black is the default text and rule color.
	black = rgb{0, 0, 0}
	// defaultAccent is a muted slate blue used when no accent is configured. It
	// stays legible in print and does not read as decorative.
	defaultAccent = rgb{0x1f, 0x4e, 0x79}
)

// accent resolves a theme accent setting into a color and whether the accent is
// enabled at all. An empty string selects the default accent; "none" (any case)
// disables it; otherwise the value is parsed as a #rrggbb / rrggbb hex color.
func accent(setting string) (color rgb, enabled bool, err error) {
	s := strings.TrimSpace(setting)
	switch {
	case s == "":
		return defaultAccent, true, nil
	case strings.EqualFold(s, "none"):
		return black, false, nil
	default:
		c, err := parseHexColor(s)
		if err != nil {
			return black, false, err
		}
		return c, true, nil
	}
}

// parseHexColor parses "#rrggbb" or "rrggbb" into an rgb value.
func parseHexColor(s string) (rgb, error) {
	h := strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(h) != 6 {
		return rgb{}, fmt.Errorf("invalid hex color %q: want #rrggbb", s)
	}
	var c rgb
	if _, err := fmt.Sscanf(h, "%02x%02x%02x", &c.r, &c.g, &c.b); err != nil {
		return rgb{}, fmt.Errorf("invalid hex color %q: %w", s, err)
	}
	return c, nil
}
