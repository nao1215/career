package pdf

import (
	"fmt"

	"github.com/nao1215/career/internal/resume"
)

// options carries resolved rendering settings shared by the colored templates.
type options struct {
	accent   rgb
	accentOn bool
}

// Template describes one renderable document kind together with the validation
// and rendering it needs.
type Template struct {
	// Name is the canonical template identifier.
	Name string
	// Aliases are alternative names accepted on the command line.
	Aliases []string
	// Description is a short human-readable summary.
	Description string
	// DefaultOutput is the file name used when the user does not pass --output.
	DefaultOutput string

	render   func(*resume.Resume, options) ([]byte, error)
	validate func(*resume.Resume) error
}

// templates lists the available document templates in display order. CV is first
// so the project reads as a general resume tool that also covers the Japanese
// formats.
var templates = []Template{
	{
		Name:          "cv",
		Aliases:       nil,
		Description:   "English curriculum vitae / résumé",
		DefaultOutput: "cv.pdf",
		render:        RenderCV,
		validate:      (*resume.Resume).ValidateCareer,
	},
	{
		Name:          "japanese-resume",
		Aliases:       []string{"履歴書"},
		Description:   "JIS-style Japanese 履歴書 (A4, 2 pages)",
		DefaultOutput: "japanese-resume.pdf",
		render:        func(r *resume.Resume, _ options) ([]byte, error) { return RenderRirekisho(r) },
		validate:      (*resume.Resume).ValidateRireki,
	},
	{
		Name:          "career-history",
		Aliases:       []string{"職務経歴書"},
		Description:   "Japanese 職務経歴書 (work history with projects)",
		DefaultOutput: "career-history.pdf",
		render:        RenderShokumukeirekisho,
		validate:      (*resume.Resume).ValidateCareer,
	},
}

// Templates returns the available templates in display order.
func Templates() []Template {
	out := make([]Template, len(templates))
	copy(out, templates)
	return out
}

// Lookup resolves a template by its canonical name or any alias.
func Lookup(name string) (Template, bool) {
	for _, t := range templates {
		if t.Name == name {
			return t, true
		}
		for _, alias := range t.Aliases {
			if alias == name {
				return t, true
			}
		}
	}
	return Template{}, false
}

// Render validates res for this template and renders it to PDF bytes.
// accentSetting controls the accent color of the colored templates: "" uses the
// default, "none" is monochrome, and "#rrggbb" sets a custom color. It is
// ignored by the always-black 履歴書.
func (t Template) Render(res *resume.Resume, accentSetting string) ([]byte, error) {
	if err := t.validate(res); err != nil {
		return nil, fmt.Errorf("%s: %w", t.Name, err)
	}
	color, on, err := accent(accentSetting)
	if err != nil {
		return nil, err
	}
	return t.render(res, options{accent: color, accentOn: on})
}
