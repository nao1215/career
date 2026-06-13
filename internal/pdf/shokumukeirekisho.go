package pdf

import (
	"strings"

	"github.com/nao1215/career/internal/font"
	"github.com/nao1215/career/internal/resume"
)

// 職務経歴書 flow-layout constants in millimetres.
const (
	skLeft   = 20.0  // left text margin
	skRight  = 190.0 // right text margin
	skWidth  = skRight - skLeft
	skTop    = 20.0  // top margin for body content
	skBottom = 282.0 // y past which content flows to a new page
	skLineH  = 6.0   // default line height for body text
	skBodyPt = 10.5  // default body font size
	skMetaH  = 5.6   // line height for the smaller history body / meta lines
	skMetaPt = 9.5   // font size for history body / meta lines
)

// RenderShokumukeirekisho renders a flowing 職務経歴書 to PDF bytes, adding pages
// as the content requires.
func RenderShokumukeirekisho(res *resume.Resume, opts options) ([]byte, error) {
	c, err := newCanvas()
	if err != nil {
		return nil, err
	}
	s := &shokumuRenderer{c: c, res: res, opts: opts}
	s.render()
	return c.bytes()
}

type shokumuRenderer struct {
	c    *canvas
	res  *resume.Resume
	opts options
	y    float64
}

func (s *shokumuRenderer) render() {
	s.c.pdf.AddPage()
	s.header()
	s.summarySection()
	s.skillsSection()
	s.historySection()
	s.listSection("資格", s.res.Career.Certifications)
	s.listSection("出版・登壇", s.res.Career.Publications)
	s.prSection()
}

// newPage starts a fresh page and resets the cursor to the top margin.
func (s *shokumuRenderer) newPage() {
	s.c.pdf.AddPage()
	s.y = skTop
}

// ensure guarantees space mm remain before the bottom margin, breaking to a new
// page when they do not.
func (s *shokumuRenderer) ensure(space float64) {
	if s.y+space > skBottom {
		s.newPage()
	}
}

// tr resolves a localized Text field in the renderer's language.
func (s *shokumuRenderer) tr(t resume.Text) string { return t.For(s.opts.lang) }

func (s *shokumuRenderer) header() {
	c := s.c
	c.setFont(font.Gothic, 18)
	c.textCenter(skLeft, skWidth, skTop, "職 務 経 歴 書")
	s.y = skTop + 12

	c.setFont(font.Mincho, 10)
	if date := s.tr(s.res.Date); date != "" {
		c.textRight(skRight, s.y, date)
	}
	s.y += 5
	if name := s.tr(s.res.Profile.Name); name != "" {
		c.textRight(skRight, s.y, "氏名　"+name)
	}
	s.y += 8
}

// heading draws a section heading: an accent bar, bold gothic title, and an
// underline rule spanning the text width. The accent color is used when enabled,
// otherwise everything is black.
func (s *shokumuRenderer) heading(title string) {
	s.ensure(skLineH * 3)
	c := s.c
	const barW = 1.8
	barColor := black
	if s.opts.accentOn {
		barColor = s.opts.accent
	}
	c.fillRect(skLeft, s.y, barW, 5.2, barColor)
	c.setFont(font.Gothic, 12.5)
	c.setColor(barColor)
	c.text(skLeft+3.5, s.y+0.3, title)
	c.setColor(black)
	s.y += 6
	c.line(skLeft, s.y, skRight, s.y)
	s.y += 4
}

func (s *shokumuRenderer) summarySection() {
	summary := s.tr(s.res.Career.Summary)
	if strings.TrimSpace(summary) == "" {
		return
	}
	s.heading("職務要約")
	s.bodyText(summary)
	s.y += 4
}

func (s *shokumuRenderer) skillsSection() {
	skills := s.res.Career.Skills
	if len(skills) == 0 {
		return
	}
	s.heading("活かせる経験・知識・技術")
	c := s.c
	c.setFont(font.Mincho, skBodyPt)
	for _, skill := range skills {
		s.ensure(skLineH)
		c.text(skLeft+1, s.y, "・")
		s.flow(skLeft+5, skLineH, s.tr(skill))
	}
	s.y += 4
}

func (s *shokumuRenderer) historySection() {
	histories := s.res.Career.Histories
	if len(histories) == 0 {
		return
	}
	s.heading("職務経歴")
	for i := range histories {
		s.companyBlock(&histories[i])
	}
}

func (s *shokumuRenderer) companyBlock(h *resume.CareerHistory) {
	c := s.c
	s.ensure(skLineH * 3)

	// Company name with an accent marker, period right-aligned on the same line.
	c.setFont(font.Gothic, 11.5)
	mark := "■"
	c.setColor(s.markColor())
	c.text(skLeft, s.y, mark)
	c.setColor(black)
	c.text(skLeft+c.textWidth(mark)+1.8, s.y, s.tr(h.Company))
	if period := s.tr(h.Period); period != "" {
		c.setFont(font.Mincho, skMetaPt)
		c.textRight(skRight, s.y+0.8, period)
	}
	s.y += 6.5

	if role := s.tr(h.Role); role != "" {
		s.metaLine(skLeft+2, skMetaH, "役職", role)
	}
	if summary := s.tr(h.Summary); strings.TrimSpace(summary) != "" {
		c.setFont(font.Mincho, skMetaPt)
		s.flow(skLeft+2, skMetaH, summary)
	}
	s.y += 2

	for i := range h.Projects {
		s.projectBlock(&h.Projects[i])
	}
	s.y += 4
}

func (s *shokumuRenderer) projectBlock(p *resume.Project) {
	c := s.c
	s.ensure(skLineH * 2)
	s.y += 1 // breathing room above each project

	// Project title with an accent marker; the period follows in a lighter tone.
	c.setFont(font.Gothic, 10)
	mark := "▸"
	c.setColor(s.markColor())
	c.text(skLeft+3, s.y, mark)
	c.setColor(black)
	titleX := skLeft + 3 + c.textWidth(mark) + 1.5
	title := s.tr(p.Title)
	if period := s.tr(p.Period); period != "" {
		title += "（" + period + "）"
	}
	s.flow(titleX, skLineH, title)

	const projIndent = skLeft + 7
	if role := s.tr(p.Role); role != "" {
		s.metaLine(projIndent, skMetaH, "役割・規模", role)
	}
	if desc := s.tr(p.Description); strings.TrimSpace(desc) != "" {
		c.setFont(font.Mincho, skMetaPt)
		s.flow(projIndent, skMetaH, desc)
	}
	if len(p.Tech) > 0 {
		s.metaLine(projIndent, skMetaH, "使用技術", strings.Join(p.Tech, " / "))
	}
	s.y += 2
}

// markColor returns the accent color when enabled, else black. It tints the
// small ■ / ▸ markers so company and project entries are easy to scan without
// adding loud color to the body text.
func (s *shokumuRenderer) markColor() rgb {
	if s.opts.accentOn {
		return s.opts.accent
	}
	return black
}

// metaLine draws a small gray caption label followed by its value as body text.
// The value wraps with a hanging indent so continuation lines align under the
// first value line rather than under the label, keeping 役割・規模 and 使用技術
// visually distinct from the surrounding description.
func (s *shokumuRenderer) metaLine(x, lineH float64, label, value string) {
	c := s.c
	// The label shares the body size and baseline so it reads as part of the
	// line; only its gothic weight and gray tone set it apart from the value.
	c.setFont(font.Gothic, skMetaPt)
	valueX := x + c.textWidth(label+"：")
	for i, line := range c.wrap(strings.TrimRight(value, "\n"), skRight-valueX) {
		s.ensure(lineH)
		if i == 0 {
			c.setFont(font.Gothic, skMetaPt)
			c.setColor(metaLabel)
			c.text(x, s.y, label+"：")
			c.setColor(black)
			c.setFont(font.Mincho, skMetaPt)
		}
		c.text(valueX, s.y, line)
		s.y += lineH
	}
}

func (s *shokumuRenderer) listSection(title string, items []resume.Text) {
	if len(items) == 0 {
		return
	}
	s.heading(title)
	c := s.c
	c.setFont(font.Mincho, skBodyPt)
	for _, item := range items {
		s.ensure(skLineH)
		c.text(skLeft+1, s.y, "・")
		s.flow(skLeft+5, skLineH, s.tr(item))
	}
	s.y += 4
}

func (s *shokumuRenderer) prSection() {
	pr := s.tr(s.res.Career.SelfPR)
	if strings.TrimSpace(pr) == "" {
		return
	}
	s.heading("自己PR")
	s.bodyText(pr)
}

// bodyText draws a wrapped paragraph at the body font from the left margin.
func (s *shokumuRenderer) bodyText(text string) {
	s.c.setFont(font.Mincho, skBodyPt)
	s.flow(skLeft, skLineH, text)
}

// flow draws wrapped text starting at absolute x, breaking pages line by line so
// long passages never overflow the bottom margin. The caller selects the font
// first; on a page break the font carries over.
func (s *shokumuRenderer) flow(x, lineH float64, text string) {
	c := s.c
	for _, line := range c.wrap(strings.TrimRight(text, "\n"), skRight-x) {
		s.ensure(lineH)
		c.text(x, s.y, line)
		s.y += lineH
	}
}
