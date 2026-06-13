// Package font embeds the IPAex Japanese fonts and registers them with a gopdf
// document. The fonts are bundled so the tool produces identical PDFs anywhere
// without requiring the user to install or download fonts first.
//
// IPAex Mincho and IPAex Gothic are distributed under the IPA Font License v1.0.
// The license text is kept alongside the font files in the assets directory.
package font

import (
	_ "embed"
	"fmt"

	"github.com/signintech/gopdf"
)

// Family names registered into the PDF document.
const (
	// Mincho is a serif typeface used for body text on formal documents.
	Mincho = "mincho"
	// Gothic is a sans-serif typeface used for headings and emphasis.
	Gothic = "gothic"
)

//go:embed assets/ipaexm.ttf
var minchoTTF []byte

//go:embed assets/ipaexg.ttf
var gothicTTF []byte

// Register adds the embedded IPAex fonts to the document under the Mincho and
// Gothic family names. Call it once after gopdf.Start and before SetFont.
func Register(pdf *gopdf.GoPdf) error {
	if err := pdf.AddTTFFontData(Mincho, minchoTTF); err != nil {
		return fmt.Errorf("register mincho font: %w", err)
	}
	if err := pdf.AddTTFFontData(Gothic, gothicTTF); err != nil {
		return fmt.Errorf("register gothic font: %w", err)
	}
	return nil
}
