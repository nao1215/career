package pdf

import (
	"strings"

	"github.com/nao1215/career/internal/font"
	"github.com/nao1215/career/internal/resume"
)

// RenderCV renders an English curriculum vitae / résumé to PDF bytes, adding
// pages as the content requires. It reads the same data as the 職務経歴書 plus the
// education list, and renders it with English section headings.
func RenderCV(res *resume.Resume, opts options) ([]byte, error) {
	c, err := newCanvas()
	if err != nil {
		return nil, err
	}
	cv := &cvRenderer{c: c, res: res, opts: opts}
	cv.render()
	return c.bytes()
}

type cvRenderer struct {
	c    *canvas
	res  *resume.Resume
	opts options
	y    float64
}

func (cv *cvRenderer) render() {
	cv.c.pdf.AddPage()
	cv.header()
	cv.summary()
	cv.skills()
	cv.list("Links", cv.res.Career.Links)
	cv.experience()
	cv.education()
	cv.list("Certifications", cv.res.Career.Certifications)
	cv.list("Publications", cv.res.Career.Publications)
}

func (cv *cvRenderer) newPage() {
	cv.c.pdf.AddPage()
	cv.y = skTop
}

func (cv *cvRenderer) ensure(space float64) {
	if cv.y+space > skBottom {
		cv.newPage()
	}
}

func (cv *cvRenderer) accentColor() rgb {
	if cv.opts.accentOn {
		return cv.opts.accent
	}
	return black
}

// tr resolves a localized Text field in the renderer's language.
func (cv *cvRenderer) tr(t resume.Text) string { return t.For(cv.opts.lang) }

func (cv *cvRenderer) header() {
	c := cv.c
	p := cv.res.Profile

	c.setFont(font.Gothic, 22)
	c.setColor(cv.accentColor())
	c.text(skLeft, skTop, cv.tr(p.Name))
	c.setColor(black)
	cv.y = skTop + 10

	// Contact line: email · phone · location.
	var parts []string
	for _, v := range []string{p.Email, p.Phone, cv.tr(p.Address.Text)} {
		if strings.TrimSpace(v) != "" {
			parts = append(parts, v)
		}
	}
	if len(parts) > 0 {
		c.setFont(font.Mincho, 9.5)
		c.text(skLeft, cv.y, strings.Join(parts, "  ·  "))
		cv.y += 6
	}
	c.line(skLeft, cv.y, skRight, cv.y)
	cv.y += 5
}

// heading draws an English section heading with an accent underline.
func (cv *cvRenderer) heading(title string) {
	cv.ensure(skLineH * 3)
	c := cv.c
	c.setFont(font.Gothic, 12)
	c.setColor(cv.accentColor())
	c.text(skLeft, cv.y, title)
	c.setColor(black)
	cv.y += 5.5
	c.line(skLeft, cv.y, skRight, cv.y)
	cv.y += 4
}

func (cv *cvRenderer) summary() {
	text := cv.tr(cv.res.Career.Summary)
	if strings.TrimSpace(text) == "" {
		text = cv.tr(cv.res.Career.SelfPR)
	}
	if strings.TrimSpace(text) == "" {
		return
	}
	cv.heading("Summary")
	cv.c.setFont(font.Mincho, skBodyPt)
	cv.flow(skLeft, skLineH, text)
	cv.y += 4
}

func (cv *cvRenderer) skills() {
	skills := cv.res.Career.Skills
	if len(skills) == 0 {
		return
	}
	cv.heading("Skills")
	c := cv.c
	c.setFont(font.Mincho, skBodyPt)
	for _, skill := range skills {
		cv.ensure(skLineH)
		c.text(skLeft+1, cv.y, "•")
		cv.flow(skLeft+5, skLineH, cv.tr(skill))
	}
	cv.y += 4
}

func (cv *cvRenderer) experience() {
	histories := cv.res.Career.Histories
	if len(histories) == 0 {
		return
	}
	cv.heading("Experience")
	for i := range histories {
		if i > 0 {
			cv.companyDivider()
		}
		cv.companyBlock(&histories[i])
	}
}

// companyDivider draws a short centered rule between company entries, with room
// above and below so it reads as a separator rather than crowding either block.
func (cv *cvRenderer) companyDivider() {
	cv.ensure(skLineH * 2)
	cv.y += 5
	cv.c.centerRule(cv.y, 24, divider)
	cv.y += 9
}

func (cv *cvRenderer) companyBlock(h *resume.CareerHistory) {
	c := cv.c
	cv.ensure(skLineH * 3)

	// Company name at the left margin; role and summary hang one rail in.
	c.setFont(font.Gothic, 11.5)
	c.text(skLeft, cv.y, cv.tr(h.Company))
	if period := cv.tr(h.Period); period != "" {
		c.setFont(font.Mincho, skMetaPt)
		c.textRight(skRight, cv.y+0.8, period)
	}
	cv.y += 6.5

	if role := cv.tr(h.Role); role != "" {
		c.setFont(font.Mincho, skMetaPt)
		cv.flow(skHistL1, skMetaH, role)
	}
	if summary := cv.tr(h.Summary); strings.TrimSpace(summary) != "" {
		c.setFont(font.Mincho, skMetaPt)
		cv.flow(skHistL1, skMetaH, summary)
	}
	cv.y += 3

	for i := range h.Projects {
		cv.projectBlock(&h.Projects[i])
	}
	cv.y += 4
}

func (cv *cvRenderer) projectBlock(p *resume.Project) {
	c := cv.c
	cv.ensure(skLineH * 2)
	cv.y += 1.5 // breathing room above each project

	// Bullet hangs on the L1 rail; the title and body align on the L2 rail.
	c.setFont(font.Gothic, 10)
	c.setColor(cv.accentColor())
	c.text(skHistL1, cv.y, "•")
	c.setColor(black)
	header := cv.tr(p.Title)
	if period := cv.tr(p.Period); period != "" {
		header += " (" + period + ")"
	}
	cv.flow(skHistL2, skLineH, header)

	// Role, then tech, then the description as its own spaced-out paragraph.
	if role := cv.tr(p.Role); role != "" {
		c.setFont(font.Mincho, skMetaPt)
		cv.flow(skHistL2, skMetaH, role)
	}
	if len(p.Tech) > 0 {
		cv.metaLine(skHistL2, skMetaH, "Tech", strings.Join(p.Tech, " / "))
	}
	if desc := cv.tr(p.Description); strings.TrimSpace(desc) != "" {
		cv.y += 1.2 // separate the detail from the metadata above
		c.setFont(font.Mincho, skMetaPt)
		cv.flow(skHistL2, skMetaH, desc)
	}
	cv.y += 2.5
}

// metaLine draws a gray "Tech: …" caption and its value, sharing the cursor and
// pagination of the renderer.
func (cv *cvRenderer) metaLine(x, lineH float64, label, value string) {
	drawMetaLine(cv.c, x, lineH, label, ": ", value, &cv.y, cv.ensure)
}

func (cv *cvRenderer) education() {
	edu := cv.res.Education
	if len(edu) == 0 {
		return
	}
	cv.heading("Education")
	c := cv.c
	c.setFont(font.Mincho, skBodyPt)
	for _, e := range edu {
		cv.ensure(skLineH)
		date := strings.TrimSpace(e.Year.String() + " " + e.Month.String())
		if date != "" {
			c.setFont(font.Mincho, 9)
			c.text(skLeft, cv.y+0.5, date)
		}
		c.setFont(font.Mincho, skBodyPt)
		cv.flow(skLeft+22, skLineH, cv.tr(e.Value))
	}
	cv.y += 4
}

func (cv *cvRenderer) list(title string, items []resume.Text) {
	if len(items) == 0 {
		return
	}
	cv.heading(title)
	c := cv.c
	c.setFont(font.Mincho, skBodyPt)
	for _, item := range items {
		cv.ensure(skLineH)
		c.text(skLeft+1, cv.y, "•")
		cv.flow(skLeft+5, skLineH, cv.tr(item))
	}
	cv.y += 4
}

// flow draws wrapped text starting at absolute x, breaking pages line by line.
// The caller selects the font first.
func (cv *cvRenderer) flow(x, lineH float64, text string) {
	c := cv.c
	for _, line := range c.wrap(strings.TrimRight(text, "\n"), skRight-x) {
		cv.ensure(lineH)
		c.text(x, cv.y, line)
		cv.y += lineH
	}
}
