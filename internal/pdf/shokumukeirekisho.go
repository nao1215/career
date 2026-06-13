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

func (s *shokumuRenderer) header() {
	c := s.c
	c.setFont(font.Gothic, 18)
	c.textCenter(skLeft, skWidth, skTop, "職 務 経 歴 書")
	s.y = skTop + 12

	c.setFont(font.Mincho, 10)
	if s.res.Date != "" {
		c.textRight(skRight, s.y, s.res.Date)
	}
	s.y += 5
	name := s.res.Profile.Name
	if name != "" {
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
	if strings.TrimSpace(s.res.Career.Summary) == "" {
		return
	}
	s.heading("職務要約")
	s.bodyText(s.res.Career.Summary)
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
		s.flow(skLeft+5, skLineH, skill)
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

	// Company name and period on one line.
	c.setFont(font.Gothic, 11.5)
	c.text(skLeft, s.y, "■ "+h.Company)
	if h.Period != "" {
		c.setFont(font.Mincho, 9.5)
		c.textRight(skRight, s.y+0.8, h.Period)
	}
	s.y += 6

	if h.Role != "" {
		c.setFont(font.Mincho, 9.5)
		s.flow(skLeft+2, skLineH, "役職: "+h.Role)
	}
	if strings.TrimSpace(h.Summary) != "" {
		c.setFont(font.Mincho, 9.5)
		s.flow(skLeft+2, 5, h.Summary)
	}
	s.y += 1

	for i := range h.Projects {
		s.projectBlock(&h.Projects[i])
	}
	s.y += 3
}

func (s *shokumuRenderer) projectBlock(p *resume.Project) {
	c := s.c
	s.ensure(skLineH * 2)

	c.setFont(font.Gothic, 10)
	header := "▸ " + p.Title
	if p.Period != "" {
		header += "（" + p.Period + "）"
	}
	s.flow(skLeft+3, skLineH, header)

	c.setFont(font.Mincho, 9.5)
	if p.Role != "" {
		s.flow(skLeft+6, 5, "役割・規模: "+p.Role)
	}
	if strings.TrimSpace(p.Description) != "" {
		s.flow(skLeft+6, 5, p.Description)
	}
	if len(p.Tech) > 0 {
		s.flow(skLeft+6, 5, "使用技術: "+strings.Join(p.Tech, ", "))
	}
	s.y += 1.5
}

func (s *shokumuRenderer) listSection(title string, items []string) {
	if len(items) == 0 {
		return
	}
	s.heading(title)
	c := s.c
	c.setFont(font.Mincho, skBodyPt)
	for _, item := range items {
		s.ensure(skLineH)
		c.text(skLeft+1, s.y, "・")
		s.flow(skLeft+5, skLineH, item)
	}
	s.y += 4
}

func (s *shokumuRenderer) prSection() {
	if strings.TrimSpace(s.res.Career.SelfPR) == "" {
		return
	}
	s.heading("自己PR")
	s.bodyText(s.res.Career.SelfPR)
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
