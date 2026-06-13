package pdf

import (
	"strings"

	"github.com/nao1215/career/internal/font"
	"github.com/nao1215/career/internal/resume"
	"github.com/signintech/gopdf"
)

// Rirekisho-specific layout constants in millimetres.
const (
	rkLeft     = 10.0  // left margin / left edge of the form
	rkRight    = 200.0 // right edge of the form
	rkWidth    = rkRight - rkLeft
	rkYearX    = 26.0  // right edge of the 年 column
	rkMonthX   = 40.0  // right edge of the 月 column
	rkValX     = 42.0  // left edge of the 内容 column
	rkRowH     = 8.0   // height of one table row
	rkTop      = 14.0  // top margin for continuation pages
	rkBottom   = 285.0 // y past which content flows to a new page
	rkTableTop = 104.0 // where the 学歴・職歴 table starts on page one
)

// RenderRirekisho renders a JIS-style 履歴書 to PDF bytes. The personal block and
// the 学歴・職歴 table open on page one; the 学歴・職歴 and 免許・資格 tables and the
// free-text fields flow onto additional pages as the data requires, so nothing
// is silently dropped.
func RenderRirekisho(res *resume.Resume, opts options) ([]byte, error) {
	c, err := newCanvas()
	if err != nil {
		return nil, err
	}

	r := &rirekishoRenderer{c: c, res: res, lang: opts.lang, photo: opts.photo}
	r.render()

	return c.bytes()
}

type rirekishoRenderer struct {
	c     *canvas
	res   *resume.Resume
	lang  string
	photo string // resolved portrait path, or ""
	y     float64
}

func (r *rirekishoRenderer) render() {
	r.c.pdf.AddPage()

	// Header: title on the left, "as of" date to the right of the text block.
	r.c.setFont(font.Gothic, 16)
	r.c.text(rkLeft, 11, "履　歴　書")
	if date := r.res.Date.For(r.lang); date != "" {
		r.c.textFit(100, 14, 60, date, 9, 7)
	}

	r.personalBlock()
	r.photoBox()

	// 学歴・職歴 fills the rest of page one (and continues onto new pages).
	r.y = rkTableTop
	r.flowGrid("学歴・職歴（各項目ごとにまとめて書く）", r.buildHistoryRows(), true, 0)

	// 免許・資格 sized to its content, then the free-text fields below it.
	r.flowGrid("免許・資格", r.licenseRows(), false, 8)
	r.y += 6

	r.ensure(16)
	r.summaryBox(r.y)
	r.y += 16 + 6

	r.freeField("趣味・特技", r.res.Rireki.Hobby, 36)
	r.freeField("志望動機", r.res.Rireki.Motivation, 40)
	r.freeField("本人希望記入欄", r.res.Rireki.Request, 40)
}

// newPage starts a fresh page and resets the cursor to the top margin.
func (r *rirekishoRenderer) newPage() {
	r.c.pdf.AddPage()
	r.y = rkTop
}

// ensure guarantees space mm remain before the bottom margin.
func (r *rirekishoRenderer) ensure(space float64) {
	if r.y+space > rkBottom {
		r.newPage()
	}
}

// personalBlock draws the bordered name/birth/address grid to the left of the
// photo. Long values shrink to fit their cell instead of overflowing.
func (r *rirekishoRenderer) personalBlock() {
	c := r.c
	p := r.res.Profile

	const (
		top     = 18.0
		left    = rkLeft
		right   = 162.0 // leave room for the photo on the right
		labelX  = left + 1.5
		valueX  = left + 22.0
		blockW  = right - left
		labelPt = 9.0
		smallPt = 8.0
	)

	yFuriganaName := top
	yName := top + 6
	yBirth := top + 20
	yFuriganaAddr := top + 32
	yAddr := top + 38
	yFuriganaContact := top + 52
	yContact := top + 58
	yPhoneRow := top + 72
	bottom := top + 80

	c.rect(left, top, blockW, bottom-top)

	// ふりがな (name)
	c.setFont(font.Mincho, smallPt)
	c.text(labelX, yFuriganaName+1, "ふりがな")
	c.textFit(valueX, yFuriganaName+1, right-valueX, p.NameKana, smallPt, 6)
	c.line(left, yName, right, yName)

	// 氏名
	c.setFont(font.Mincho, labelPt)
	c.text(labelX, yName+5, "氏　名")
	c.textFit(valueX, yName+4, right-valueX, p.Name.For(r.lang), 18, 11)
	c.line(left, yBirth, right, yBirth)

	// 生年月日 / 性別
	genderX := 122.0
	c.line(genderX, yBirth, genderX, yFuriganaAddr)
	c.setFont(font.Mincho, labelPt)
	c.text(labelX, yBirth+2, "生年月日")
	birth := p.BirthDate
	if p.Age != "" {
		birth += "　（" + p.Age + "）"
	}
	c.textFit(valueX, yBirth+6, genderX-valueX-2, birth, 12, 8)
	c.setFont(font.Mincho, labelPt)
	c.text(genderX+2, yBirth+2, "性別")
	c.textFit(genderX+10, yBirth+6, right-genderX-12, p.Gender, 12, 8)
	c.line(left, yFuriganaAddr, right, yFuriganaAddr)

	// ふりがな (address)
	c.setFont(font.Mincho, smallPt)
	c.text(labelX, yFuriganaAddr+1, "ふりがな")
	c.textFit(valueX, yFuriganaAddr+1, right-valueX, p.Address.Kana, smallPt, 6)
	c.line(left, yAddr, right, yAddr)

	// 現住所
	c.setFont(font.Mincho, labelPt)
	c.text(labelX, yAddr+2, "現住所")
	if p.Address.Zip != "" {
		c.text(labelX, yAddr+7, "〒 "+p.Address.Zip)
	}
	c.textFit(valueX, yAddr+8, right-valueX, p.Address.Text.For(r.lang), 11, 7)
	c.line(left, yFuriganaContact, right, yFuriganaContact)

	// ふりがな (contact)
	c.setFont(font.Mincho, smallPt)
	c.text(labelX, yFuriganaContact+1, "ふりがな")
	c.textFit(valueX, yFuriganaContact+1, right-valueX, p.Contact.Kana, smallPt, 6)
	c.line(left, yContact, right, yContact)

	// 連絡先
	c.setFont(font.Mincho, labelPt)
	c.text(labelX, yContact+2, "連絡先")
	if !p.Contact.Text.Has() {
		c.setFont(font.Mincho, smallPt)
		c.text(labelX, yContact+7, "（現住所に同じ）")
	} else {
		if p.Contact.Zip != "" {
			c.text(labelX, yContact+7, "〒 "+p.Contact.Zip)
		}
		c.textFit(valueX, yContact+8, right-valueX, p.Contact.Text.For(r.lang), 11, 7)
	}
	c.line(left, yPhoneRow, right, yPhoneRow)

	// 携帯電話 / E-MAIL
	emailX := 80.0
	c.line(emailX, yPhoneRow, emailX, bottom)
	c.setFont(font.Mincho, smallPt)
	c.text(labelX, yPhoneRow+1.5, "携帯電話")
	c.textFit(labelX+18, yPhoneRow+1.5, emailX-(labelX+18)-1, p.Phone, labelPt, 6)
	c.setFont(font.Mincho, smallPt)
	c.text(emailX+2, yPhoneRow+1.5, "E-MAIL")
	c.textFit(emailX+16, yPhoneRow+1.5, right-(emailX+16)-1, p.Email, labelPt, 6)
}

// photoBox draws the portrait placeholder, or the supplied image when present.
func (r *rirekishoRenderer) photoBox() {
	c := r.c
	const (
		x = 167.0
		y = 18.0
		w = 30.0
		h = 40.0
	)
	if r.photo != "" {
		if iw, ih, err := photoDims(r.photo); err == nil {
			// Fit the image inside the frame without distortion and center it.
			fw, fh, dx, dy := containRect(iw, ih, w, h)
			if err := c.pdf.Image(r.photo, x+dx, y+dy, &gopdf.Rect{W: fw, H: fh}); err == nil {
				c.rect(x, y, w, h) // frame around the (possibly letterboxed) photo
				return
			}
		}
		// Fall through to the placeholder when the image cannot be loaded.
	}
	c.rect(x, y, w, h)
	c.setFont(font.Mincho, 8)
	c.textCenter(x, w, y+14, "写真を貼る位置")
	c.setFont(font.Mincho, 7)
	c.textCenter(x, w, y+20, "縦36〜40mm")
	c.textCenter(x, w, y+24, "横24〜30mm")
}

// historyRow is one already-formatted line in a year/month/value table.
type historyRow struct {
	year    string
	month   string
	value   string
	center  bool // when true, value is centred (used for the 学歴/職歴 captions)
	righted bool // when true, value is right-aligned (used for 以上)
}

// buildHistoryRows turns the education and work lists into display rows with the
// conventional 学歴 / 職歴 captions and a trailing 以上 marker.
func (r *rirekishoRenderer) buildHistoryRows() []historyRow {
	var rows []historyRow
	if len(r.res.Education) > 0 {
		rows = append(rows, historyRow{value: "学歴", center: true})
		for _, e := range r.res.Education {
			rows = append(rows, historyRow{year: e.Year.String(), month: e.Month.String(), value: e.Value.For(r.lang)})
		}
		rows = append(rows, historyRow{})
	}
	if len(r.res.Work) > 0 {
		rows = append(rows, historyRow{value: "職歴", center: true})
		for _, w := range r.res.Work {
			rows = append(rows, historyRow{year: w.Year.String(), month: w.Month.String(), value: w.Value.For(r.lang)})
		}
	}
	rows = append(rows, historyRow{value: "以上", righted: true})
	return rows
}

// licenseRows turns the licenses list into display rows with a trailing 以上.
func (r *rirekishoRenderer) licenseRows() []historyRow {
	rows := make([]historyRow, 0, len(r.res.Licenses)+1)
	for _, l := range r.res.Licenses {
		rows = append(rows, historyRow{year: l.Year.String(), month: l.Month.String(), value: l.Value.For(r.lang)})
	}
	rows = append(rows, historyRow{value: "以上", righted: true})
	return rows
}

// flowGrid draws a bordered year/month/value table starting at the cursor,
// breaking to new pages until every row is drawn. Full pages are filled to the
// bottom. When fillLastToBottom is true the final page is filled too (used for
// the 学歴・職歴 table that should occupy page one); otherwise the final segment is
// sized to its content but at least minLastRows tall, so later content can follow.
func (r *rirekishoRenderer) flowGrid(caption string, rows []historyRow, fillLastToBottom bool, minLastRows int) {
	for {
		avail := int((rkBottom - r.y) / rkRowH) // rows that fit, including the header
		if avail < 2 {
			r.newPage()
			continue
		}
		bodyCap := avail - 1
		isLast := len(rows) <= bodyCap

		var chunk []historyRow
		if isLast {
			chunk, rows = rows, nil
		} else {
			chunk, rows = rows[:bodyCap], rows[bodyCap:]
		}

		segRows := avail // a full page by default
		if isLast && !fillLastToBottom {
			segRows = 1 + len(chunk)
			if segRows < minLastRows {
				segRows = minLastRows
			}
			if segRows > avail {
				segRows = avail
			}
		}

		r.drawGridSegment(r.y, segRows, caption, chunk)
		r.y += float64(segRows) * rkRowH

		if isLast {
			return
		}
		r.newPage()
	}
}

// drawGridSegment draws one page-segment of a table: the outer frame, the column
// separators, a header row, the row separators, and the supplied content rows.
func (r *rirekishoRenderer) drawGridSegment(top float64, segRows int, caption string, chunk []historyRow) {
	c := r.c
	height := float64(segRows) * rkRowH

	c.rect(rkLeft, top, rkWidth, height)
	c.line(rkYearX, top, rkYearX, top+height)
	c.line(rkMonthX, top, rkMonthX, top+height)

	// Header row.
	c.setFont(font.Mincho, 9)
	c.textCenter(rkLeft, rkYearX-rkLeft, top+2, "年")
	c.textCenter(rkYearX, rkMonthX-rkYearX, top+2, "月")
	c.textCenter(rkValX, rkRight-rkValX, top+2, caption)

	// Row separators (the rect supplies the bottom edge).
	for i := 1; i < segRows; i++ {
		y := top + float64(i)*rkRowH
		c.line(rkLeft, y, rkRight, y)
	}

	// Content rows start just below the header.
	c.setFont(font.Mincho, 11)
	for i, row := range chunk {
		y := top + float64(i+1)*rkRowH + 2
		switch {
		case row.center:
			c.textCenter(rkValX, rkRight-rkValX, y, row.value)
		case row.righted:
			c.textRight(rkRight-2, y, row.value)
		default:
			c.textRight(rkYearX-1, y, row.year)
			c.textRight(rkMonthX-1, y, row.month)
			c.textFit(rkValX, y, rkRight-rkValX-2, row.value, 11, 7)
		}
	}
}

// summaryBox draws the single-row commute/dependents/spouse table.
func (r *rirekishoRenderer) summaryBox(top float64) {
	c := r.c
	const h = 16.0
	rk := r.res.Rireki

	c.rect(rkLeft, top, rkWidth, h)
	x1 := rkLeft + 55
	x2 := rkLeft + 100
	x3 := rkLeft + 145
	c.line(x1, top, x1, top+h)
	c.line(x2, top, x2, top+h)
	c.line(x3, top, x3, top+h)

	c.setFont(font.Mincho, 9)
	c.text(rkLeft+2, top+2, "通勤時間")
	c.text(x1+2, top+2, "扶養家族数（配偶者を除く）")
	c.text(x2+2, top+2, "配偶者")
	c.text(x3+2, top+2, "配偶者の扶養義務")

	c.setFont(font.Mincho, 12)
	c.text(rkLeft+6, top+9, rk.CommutingTime)
	c.text(x1+6, top+9, rk.Dependents)
	c.text(x2+6, top+9, rk.Spouse)
	c.text(x3+6, top+9, rk.SupportingSpouse)
}

// freeField draws a labelled free-text box of the given height, breaking to a
// new page first when it would not fit, and advances the cursor below it.
func (r *rirekishoRenderer) freeField(label, body string, h float64) {
	r.ensure(h)
	c := r.c
	top := r.y
	c.rect(rkLeft, top, rkWidth, h)
	c.setFont(font.Mincho, 9)
	c.text(rkLeft+2, top+2, label)
	if strings.TrimSpace(body) != "" {
		c.setFont(font.Mincho, 11)
		c.paragraph(rkLeft+3, top+8, rkWidth-6, 6, strings.TrimRight(body, "\n"))
	}
	r.y = top + h + 5
}
