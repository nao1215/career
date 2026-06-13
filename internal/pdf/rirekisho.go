package pdf

import (
	"strings"

	"github.com/nao1215/career/internal/font"
	"github.com/nao1215/career/internal/resume"
	"github.com/signintech/gopdf"
)

// Rirekisho-specific layout constants in millimetres.
const (
	rkLeft   = 10.0  // left margin / left edge of the form
	rkRight  = 200.0 // right edge of the form
	rkWidth  = rkRight - rkLeft
	rkYearX  = 26.0 // right edge of the 年 column
	rkMonthX = 40.0 // right edge of the 月 column
	rkValX   = 42.0 // left edge of the 内容 column
	rkRowH   = 8.0  // height of one table row
)

// RenderRirekisho renders a JIS-style 履歴書 to PDF bytes. The layout follows
// the conventional two-page A4 format: page one carries the personal block and
// the 学歴・職歴 table, page two the 免許・資格 table and the free-text fields.
func RenderRirekisho(res *resume.Resume, opts options) ([]byte, error) {
	c, err := newCanvas()
	if err != nil {
		return nil, err
	}

	rireki := &rirekishoRenderer{c: c, res: res, lang: opts.lang}
	rireki.page1()
	rireki.page2()

	return c.bytes()
}

type rirekishoRenderer struct {
	c    *canvas
	res  *resume.Resume
	lang string
}

func (r *rirekishoRenderer) page1() {
	c := r.c
	c.pdf.AddPage()

	// Header: title on the left, "as of" date on the right of the text block.
	c.setFont(font.Gothic, 16)
	c.text(rkLeft, 11, "履　歴　書")
	if date := r.res.Date.For(r.lang); date != "" {
		c.setFont(font.Mincho, 9)
		c.textRight(160, 14, date)
	}

	r.personalBlock()
	r.photoBox()
	r.educationExperienceTable()
}

// personalBlock draws the bordered name/birth/address grid to the left of the
// photo.
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

	// Row boundaries from the top down.
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
	c.text(valueX, yFuriganaName+1, p.NameKana)
	c.line(left, yName, right, yName)

	// 氏名
	c.setFont(font.Mincho, labelPt)
	c.text(labelX, yName+5, "氏　名")
	c.setFont(font.Mincho, 18)
	c.text(valueX, yName+4, p.Name.For(r.lang))
	c.line(left, yBirth, right, yBirth)

	// 生年月日 / 性別
	genderX := 122.0
	c.line(genderX, yBirth, genderX, yFuriganaAddr)
	c.setFont(font.Mincho, labelPt)
	c.text(labelX, yBirth+2, "生年月日")
	c.setFont(font.Mincho, 12)
	birth := p.BirthDate
	if p.Age != "" {
		birth += "　（" + p.Age + "）"
	}
	c.text(valueX, yBirth+6, birth)
	c.setFont(font.Mincho, labelPt)
	c.text(genderX+2, yBirth+2, "性別")
	c.setFont(font.Mincho, 12)
	c.text(genderX+10, yBirth+6, p.Gender)
	c.line(left, yFuriganaAddr, right, yFuriganaAddr)

	// ふりがな (address)
	c.setFont(font.Mincho, smallPt)
	c.text(labelX, yFuriganaAddr+1, "ふりがな")
	c.text(valueX, yFuriganaAddr+1, p.Address.Kana)
	c.line(left, yAddr, right, yAddr)

	// 現住所
	c.setFont(font.Mincho, labelPt)
	c.text(labelX, yAddr+2, "現住所")
	if p.Address.Zip != "" {
		c.text(labelX, yAddr+7, "〒 "+p.Address.Zip)
	}
	c.setFont(font.Mincho, 11)
	c.text(valueX, yAddr+8, p.Address.Text.For(r.lang))
	c.line(left, yFuriganaContact, right, yFuriganaContact)

	// ふりがな (contact)
	c.setFont(font.Mincho, smallPt)
	c.text(labelX, yFuriganaContact+1, "ふりがな")
	c.text(valueX, yFuriganaContact+1, p.Contact.Kana)
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
		c.setFont(font.Mincho, 11)
		c.text(valueX, yContact+8, p.Contact.Text.For(r.lang))
	}
	c.line(left, yPhoneRow, right, yPhoneRow)

	// 携帯電話 / E-MAIL
	emailX := 80.0
	c.line(emailX, yPhoneRow, emailX, bottom)
	c.setFont(font.Mincho, smallPt)
	c.text(labelX, yPhoneRow+1.5, "携帯電話")
	c.setFont(font.Mincho, labelPt)
	c.text(labelX+18, yPhoneRow+1.5, p.Phone)
	c.setFont(font.Mincho, smallPt)
	c.text(emailX+2, yPhoneRow+1.5, "E-MAIL")
	c.setFont(font.Mincho, labelPt)
	c.text(emailX+16, yPhoneRow+1.5, p.Email)
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
	if r.res.Profile.Photo != "" {
		if err := c.pdf.Image(r.res.Profile.Photo, x, y, &gopdf.Rect{W: w, H: h}); err == nil {
			return
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

// educationExperienceTable draws the 学歴・職歴 grid filling the rest of page one.
func (r *rirekishoRenderer) educationExperienceTable() {
	const top = 104.0
	const bottom = 285.0

	rows := r.buildHistoryRows()
	r.drawHistoryGrid(top, bottom, "学歴・職歴（各項目ごとにまとめて書く）", rows)
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

// drawHistoryGrid renders a bordered year/month/value table between top and
// bottom, painting empty rows to fill the remaining space.
func (r *rirekishoRenderer) drawHistoryGrid(top, bottom float64, caption string, rows []historyRow) {
	c := r.c
	height := bottom - top
	rowCount := int(height / rkRowH)

	// Outer frame and column separators.
	c.rect(rkLeft, top, rkWidth, float64(rowCount)*rkRowH)
	c.line(rkYearX, top, rkYearX, top+float64(rowCount)*rkRowH)
	c.line(rkMonthX, top, rkMonthX, top+float64(rowCount)*rkRowH)

	// Header row.
	c.setFont(font.Mincho, 9)
	c.textCenter(rkLeft, rkYearX-rkLeft, top+2, "年")
	c.textCenter(rkYearX, rkMonthX-rkYearX, top+2, "月")
	c.textCenter(rkValX, rkRight-rkValX, top+2, caption)
	c.line(rkLeft, top+rkRowH, rkRight, top+rkRowH)

	// Body rows.
	for i := 1; i < rowCount; i++ {
		y := top + float64(i)*rkRowH
		c.line(rkLeft, y, rkRight, y)
	}

	c.setFont(font.Mincho, 11)
	for i, row := range rows {
		if i+1 >= rowCount {
			break // ran out of room; remaining rows are dropped from this page
		}
		y := top + float64(i+1)*rkRowH + 2
		switch {
		case row.center:
			c.textCenter(rkValX, rkRight-rkValX, y, row.value)
		case row.righted:
			c.textRight(rkRight-2, y, row.value)
		default:
			c.textRight(rkYearX-1, y, row.year)
			c.textRight(rkMonthX-1, y, row.month)
			c.text(rkValX, y, row.value)
		}
	}
}

func (r *rirekishoRenderer) page2() {
	c := r.c
	c.pdf.AddPage()

	// 免許・資格 table at the top of page two.
	licTop := 14.0
	licRows := make([]historyRow, 0, len(r.res.Licenses)+1)
	for _, l := range r.res.Licenses {
		licRows = append(licRows, historyRow{year: l.Year.String(), month: l.Month.String(), value: l.Value.For(r.lang)})
	}
	licRows = append(licRows, historyRow{value: "以上", righted: true})
	licBottom := licTop + rkRowH*float64(maxInt(len(licRows)+1, 8))
	r.drawHistoryGrid(licTop, licBottom, "免許・資格", licRows)

	y := licBottom + 8

	// 通勤時間 / 扶養家族 / 配偶者 summary box.
	y = r.summaryBox(y)
	y += 6

	// Free-text fields.
	y = r.textField(y, "趣味・特技", r.res.Rireki.Hobby, 36)
	y += 5
	y = r.textField(y, "志望動機", r.res.Rireki.Motivation, 40)
	y += 5
	_ = r.textField(y, "本人希望記入欄", r.res.Rireki.Request, 40)
}

// summaryBox draws the single-row commute/dependents/spouse table and returns
// the y just below it.
func (r *rirekishoRenderer) summaryBox(top float64) float64 {
	c := r.c
	const h = 16.0
	rk := r.res.Rireki

	c.rect(rkLeft, top, rkWidth, h)
	// Four equal-ish cells.
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
	return top + h
}

// textField draws a labelled free-text box of the given height and returns the
// y just below it.
func (r *rirekishoRenderer) textField(top float64, label, body string, h float64) float64 {
	c := r.c
	c.rect(rkLeft, top, rkWidth, h)
	c.setFont(font.Mincho, 9)
	c.text(rkLeft+2, top+2, label)
	if strings.TrimSpace(body) != "" {
		c.setFont(font.Mincho, 11)
		c.paragraph(rkLeft+3, top+8, rkWidth-6, 6, strings.TrimRight(body, "\n"))
	}
	return top + h
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
