package resume

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Language codes understood by the templates.
const (
	LangJA = "ja"
	LangEN = "en"
)

// defaultLangKey is the map key for a value written as a plain scalar, i.e. one
// that applies to every language.
const defaultLangKey = ""

// Text is a localized string. In YAML it is written either as a plain scalar,
// which applies to every language, or as a mapping of language code to string:
//
//	summary: applies to every language
//
//	summary:
//	  ja: 日本語の本文
//	  en: English text
//
// This lets a single document hold both the Japanese 履歴書/職務経歴書 content and
// the English CV content; each template asks for the language it needs.
type Text struct {
	byLang map[string]string
}

// UnmarshalYAML accepts either a scalar or a language-keyed mapping.
func (t *Text) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		t.byLang = map[string]string{defaultLangKey: value.Value}
		return nil
	case yaml.MappingNode:
		m := map[string]string{}
		if err := value.Decode(&m); err != nil {
			return fmt.Errorf("line %d: %w", value.Line, err)
		}
		t.byLang = m
		return nil
	default:
		return fmt.Errorf("line %d: text must be a string or a language map", value.Line)
	}
}

// For returns the string for lang. When that language is missing it falls back
// to the unkeyed default, then Japanese, then English, then any value, so a
// document written in a single language still renders under every template.
func (t Text) For(lang string) string {
	if t.byLang == nil {
		return ""
	}
	if v := t.byLang[lang]; v != "" {
		return v
	}
	for _, k := range []string{defaultLangKey, LangJA, LangEN} {
		if v := t.byLang[k]; v != "" {
			return v
		}
	}
	for _, v := range t.byLang {
		if v != "" {
			return v
		}
	}
	return ""
}

// Has reports whether the text holds any non-empty value.
func (t Text) Has() bool {
	return t.For(defaultLangKey) != ""
}

// Plain builds a Text that applies to every language. It is mainly used in tests
// and by the init scaffold.
func Plain(s string) Text {
	return Text{byLang: map[string]string{defaultLangKey: s}}
}
