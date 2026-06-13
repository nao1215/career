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

func (cv *cvRenderer) header() {
	c := cv.c
	p := cv.res.Profile

	c.setFont(font.Gothic, 22)
	c.setColor(cv.accentColor())
	c.text(skLeft, skTop, p.Name)
	c.setColor(black)
	cv.y = skTop + 10

	// Contact line: email · phone · location.
	var parts []string
	for _, v := range []string{p.Email, p.Phone, p.Address.Text} {
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
	text := cv.res.Career.Summary
	if strings.TrimSpace(text) == "" {
		text = cv.res.Career.SelfPR
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
		cv.flow(skLeft+5, skLineH, skill)
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
		cv.companyBlock(&histories[i])
	}
}

func (cv *cvRenderer) companyBlock(h *resume.CareerHistory) {
	c := cv.c
	cv.ensure(skLineH * 3)

	c.setFont(font.Gothic, 11.5)
	c.text(skLeft, cv.y, h.Company)
	if h.Period != "" {
		c.setFont(font.Mincho, 9.5)
		c.textRight(skRight, cv.y+0.8, h.Period)
	}
	cv.y += 6

	if h.Role != "" {
		c.setFont(font.Mincho, 9.5)
		cv.flow(skLeft+2, skLineH, h.Role)
	}
	if strings.TrimSpace(h.Summary) != "" {
		c.setFont(font.Mincho, 9.5)
		cv.flow(skLeft+2, 5, h.Summary)
	}
	cv.y += 1

	for i := range h.Projects {
		cv.projectBlock(&h.Projects[i])
	}
	cv.y += 3
}

func (cv *cvRenderer) projectBlock(p *resume.Project) {
	c := cv.c
	cv.ensure(skLineH * 2)

	c.setFont(font.Gothic, 10)
	header := p.Title
	if p.Period != "" {
		header += " (" + p.Period + ")"
	}
	cv.flow(skLeft+3, skLineH, "• "+header)

	c.setFont(font.Mincho, 9.5)
	if p.Role != "" {
		cv.flow(skLeft+6, 5, p.Role)
	}
	if strings.TrimSpace(p.Description) != "" {
		cv.flow(skLeft+6, 5, p.Description)
	}
	if len(p.Tech) > 0 {
		cv.flow(skLeft+6, 5, "Tech: "+strings.Join(p.Tech, ", "))
	}
	cv.y += 1.5
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
		cv.flow(skLeft+22, skLineH, e.Value)
	}
	cv.y += 4
}

func (cv *cvRenderer) list(title string, items []string) {
	if len(items) == 0 {
		return
	}
	cv.heading(title)
	c := cv.c
	c.setFont(font.Mincho, skBodyPt)
	for _, item := range items {
		cv.ensure(skLineH)
		c.text(skLeft+1, cv.y, "•")
		cv.flow(skLeft+5, skLineH, item)
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
