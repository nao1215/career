// Package resume defines the data model for a person's career information and
// loads it from a YAML file. A single document holds everything needed to
// render both a 履歴書 (rirekisho) and a 職務経歴書 (shokumukeirekisho); each
// generator reads only the sections it needs.
package resume

import "gopkg.in/yaml.v3"

// Resume is the root document loaded from YAML.
type Resume struct {
	// Date is printed in the header, e.g. "2026年6月13日現在".
	Date Text `yaml:"date"`
	// Theme controls optional styling such as the accent color.
	Theme Theme `yaml:"theme"`
	// Profile holds personal information shared by both documents.
	Profile Profile `yaml:"profile"`

	// Education is the 学歴 list, oldest first.
	Education []HistoryItem `yaml:"education"`
	// Work is the 職歴 list, oldest first.
	Work []HistoryItem `yaml:"work"`
	// Licenses is the 免許・資格 list.
	Licenses []HistoryItem `yaml:"licenses"`

	// Rireki holds fields unique to the 履歴書.
	Rireki Rireki `yaml:"rireki"`
	// Career holds fields unique to the 職務経歴書.
	Career Career `yaml:"career"`
}

// Theme controls optional styling shared by the CV and 職務経歴書.
type Theme struct {
	// Accent is the accent color as a hex string ("#1f4e79"). Empty uses the
	// default; "none" disables the accent and renders in monochrome. The JIS
	// 履歴書 always renders in black regardless of this setting.
	Accent string `yaml:"accent"`
}

// Profile is the personal information block.
type Profile struct {
	Name      Text    `yaml:"name"`       // 氏名 (localizable: kanji for JP, romaji for the CV)
	NameKana  string  `yaml:"name_kana"`  // ふりがな (Japanese only)
	BirthDate string  `yaml:"birth_date"` // 生年月日, e.g. "1990年1月1日"
	Age       string  `yaml:"age"`        // optional, e.g. "満 35 歳"
	Gender    string  `yaml:"gender"`     // 性別
	Email     string  `yaml:"email"`
	Phone     string  `yaml:"phone"` // 携帯電話番号
	Photo     string  `yaml:"photo"` // optional path to a JPEG/PNG portrait
	Address   Address `yaml:"address"`
	Contact   Address `yaml:"contact"` // 連絡先, optional
}

// Address is a postal address with optional phone and fax.
type Address struct {
	Zip  string `yaml:"zip"`  // 〒
	Kana string `yaml:"kana"` // ふりがな
	Text Text   `yaml:"text"` // 住所 (localizable)
	Tel  string `yaml:"tel"`
	Fax  string `yaml:"fax"`
}

// HistoryItem is one dated row in a year/month/value table such as 学歴・職歴
// or 免許・資格.
type HistoryItem struct {
	Year  Flex `yaml:"year"`
	Month Flex `yaml:"month"`
	Value Text `yaml:"value"`
}

// Rireki holds the 履歴書-only fields.
type Rireki struct {
	CommutingTime    string `yaml:"commuting_time"`    // 通勤時間
	Dependents       string `yaml:"dependents"`        // 扶養家族数(配偶者を除く)
	Spouse           string `yaml:"spouse"`            // 配偶者
	SupportingSpouse string `yaml:"supporting_spouse"` // 配偶者の扶養義務
	Hobby            string `yaml:"hobby"`             // 趣味・特技
	Motivation       string `yaml:"motivation"`        // 志望動機
	Request          string `yaml:"request"`           // 本人希望記入欄
}

// Career holds the 職務経歴書-only fields. Its text is localizable so the same
// section drives both the Japanese 職務経歴書 and the English CV.
type Career struct {
	Summary        Text            `yaml:"summary"`        // 職務要約
	Skills         []Text          `yaml:"skills"`         // 活かせる経験・知識・技術
	Histories      []CareerHistory `yaml:"histories"`      // 職務経歴
	Certifications []Text          `yaml:"certifications"` // 資格
	Publications   []Text          `yaml:"publications"`   // 出版・登壇など
	Links          []Text          `yaml:"links"`          // リンク（GitHub、ブログなど）
	SelfPR         Text            `yaml:"self_pr"`        // 自己PR
}

// CareerHistory is one employer block in the 職務経歴.
type CareerHistory struct {
	Company  Text      `yaml:"company"`  // 会社名
	Period   Text      `yaml:"period"`   // 在籍期間, e.g. "2022年1月 - 2025年2月"
	Role     Text      `yaml:"role"`     // 役職, optional
	Summary  Text      `yaml:"summary"`  // 事業内容・概要, optional
	Projects []Project `yaml:"projects"` // 担当プロジェクト
}

// Project is one assignment within a CareerHistory.
type Project struct {
	Title       Text     `yaml:"title"`       // 案件名・テーマ
	Period      Text     `yaml:"period"`      // 期間, optional
	Role        Text     `yaml:"role"`        // 役割・規模, optional
	Description Text     `yaml:"description"` // 業務内容
	Tech        []string `yaml:"tech"`        // 使用技術 (language-neutral)
}

// Flex is a YAML scalar that accepts either a string or a number and stores it
// as a string. It lets a user write `year: 2009` or `year: "20XX"` freely
// without worrying about quoting.
type Flex string

// UnmarshalYAML stores the raw scalar value regardless of its YAML type.
func (f *Flex) UnmarshalYAML(value *yaml.Node) error {
	*f = Flex(value.Value)
	return nil
}

// String returns the scalar as a plain string.
func (f Flex) String() string { return string(f) }
