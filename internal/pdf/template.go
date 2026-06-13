package pdf

import (
	"fmt"

	"github.com/nao1215/career/internal/resume"
)

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

	render   func(*resume.Resume) ([]byte, error)
	validate func(*resume.Resume) error
}

// templates lists the available document templates in display order.
var templates = []Template{
	{
		Name:          "rirekisho",
		Aliases:       []string{"rireki", "resume"},
		Description:   "JIS規格スタイルの履歴書（A4・2ページ）",
		DefaultOutput: "rirekisho.pdf",
		render:        RenderRirekisho,
		validate:      (*resume.Resume).ValidateRireki,
	},
	{
		Name:          "shokumukeirekisho",
		Aliases:       []string{"shokureki", "career", "cv"},
		Description:   "職務経歴書（職務要約・スキル・職務経歴・自己PR）",
		DefaultOutput: "shokumukeirekisho.pdf",
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
func (t Template) Render(res *resume.Resume) ([]byte, error) {
	if err := t.validate(res); err != nil {
		return nil, fmt.Errorf("%s: %w", t.Name, err)
	}
	return t.render(res)
}
