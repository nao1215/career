package resume

// Generated properties for Text, the localized string behind every field of
// the résumé YAML. A failure prints the seed to replay it.

import (
	"sort"
	"testing"

	"gopkg.in/yaml.v3"
	"pgregory.net/rapid"
)

var propLangs = []string{"ja", "en", "fr", "de", "zh"}

func genLangMap() *rapid.Generator[map[string]string] {
	return rapid.MapOfN(
		rapid.SampledFrom(propLangs),
		rapid.SampledFrom([]string{"", "見本", "sample", "x"}),
		0, len(propLangs),
	)
}

// wantFor is the fallback order For documents, written independently: the
// asked language, then ja, then en, then the remaining languages in key order,
// skipping empty values.
func wantFor(m map[string]string, lang string) string {
	order := []string{lang, LangJA, LangEN}
	rest := make([]string, 0, len(m))
	for k := range m {
		rest = append(rest, k)
	}
	sort.Strings(rest)
	for _, k := range append(order, rest...) {
		if v := m[k]; v != "" {
			return v
		}
	}
	return ""
}

// A language map written in YAML and decoded into Text resolves every
// language the way the documented fallback order says.
func TestText_LanguageMapFollowsTheFallbackOrder(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		m := genLangMap().Draw(t, "map")
		lang := rapid.SampledFrom(propLangs).Draw(t, "lang")
		doc, err := yaml.Marshal(m)
		if err != nil {
			t.Fatalf("yaml.Marshal(%v) error = %v", m, err)
		}
		var text Text
		if err := yaml.Unmarshal(doc, &text); err != nil {
			t.Fatalf("yaml.Unmarshal(%q) error = %v", doc, err)
		}
		if got, want := text.For(lang), wantFor(m, lang); got != want {
			t.Fatalf("For(%q) on %v = %q, want %q", lang, m, got, want)
		}
	})
}

// A scalar applies to every language: For returns it whatever is asked.
func TestText_ScalarAppliesToEveryLanguage(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		s := rapid.StringMatching(`[a-zA-Z0-9 見本。]{1,20}`).Draw(t, "s")
		lang := rapid.SampledFrom(propLangs).Draw(t, "lang")
		doc, err := yaml.Marshal(s)
		if err != nil {
			t.Fatalf("yaml.Marshal(%q) error = %v", s, err)
		}
		var text Text
		if err := yaml.Unmarshal(doc, &text); err != nil {
			t.Fatalf("yaml.Unmarshal(%q) error = %v", doc, err)
		}
		if got := text.For(lang); got != s {
			t.Fatalf("For(%q) on scalar %q = %q", lang, s, got)
		}
	})
}
