package pdf

import (
	"fmt"

	"github.com/nao1215/career/internal/resume"
)

// RenderOptions are the caller-supplied settings for a render.
type RenderOptions struct {
	// Accent sets the accent color: "" default, "none" monochrome, or "#rrggbb".
	Accent string
	// Photo is a resolved path to a portrait image, or "" for none. Only the
	// 履歴書 uses it.
	Photo string
}

// options carries resolved rendering settings passed to a renderer.
type options struct {
	accent   rgb
	accentOn bool
	lang     string // language code the renderer should request from Text fields
	photo    string // resolved portrait path, or ""
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

	lang      string // language requested from localized Text fields
	usesPhoto bool   // whether this template renders a portrait
	render    func(*resume.Resume, options) ([]byte, error)
	validate  func(*resume.Resume) error
}

// UsesPhoto reports whether the template renders a portrait image.
func (t Template) UsesPhoto() bool { return t.usesPhoto }

// templates lists the available document templates in display order. CV is first
// so the project reads as a general resume tool that also covers the Japanese
// formats.
var templates = []Template{
	{
		Name:          "cv",
		Aliases:       nil,
		Description:   "English curriculum vitae / résumé",
		DefaultOutput: "cv.pdf",
		lang:          resume.LangEN,
		render:        RenderCV,
		validate:      (*resume.Resume).ValidateCareer,
	},
	{
		Name:          "japanese-resume",
		Aliases:       []string{"履歴書"},
		Description:   "JIS-style Japanese 履歴書 (A4)",
		DefaultOutput: "japanese-resume.pdf",
		lang:          resume.LangJA,
		usesPhoto:     true,
		render:        RenderRirekisho,
		validate:      (*resume.Resume).ValidateRireki,
	},
	{
		Name:          "work-history",
		Aliases:       []string{"職務経歴書", "career-history"},
		Description:   "Japanese 職務経歴書 (work history with projects)",
		DefaultOutput: "work-history.pdf",
		lang:          resume.LangJA,
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

// Render validates res for this template and renders it to PDF bytes using the
// given options.
func (t Template) Render(res *resume.Resume, ro RenderOptions) ([]byte, error) {
	if err := t.validate(res); err != nil {
		return nil, fmt.Errorf("%s: %w", t.Name, err)
	}
	color, on, err := accent(ro.Accent)
	if err != nil {
		return nil, err
	}
	return t.render(res, options{accent: color, accentOn: on, lang: t.lang, photo: ro.Photo})
}
