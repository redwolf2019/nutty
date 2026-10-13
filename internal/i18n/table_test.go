package i18n

import (
	"regexp"
	"slices"
	"strings"
	"testing"
	"unicode"
)

// TestTableEntries checks every entry of the Chinese table for the
// mistakes an exact-match table hides until the string is drawn: an empty
// translation, format verbs that differ from the key's, which fmt would
// print as %!s(MISSING) or %!(EXTRA ...), and a space at either end that
// the translation does not keep, where a caller joins the string to
// another.
func TestTableEntries(t *testing.T) {
	verb := regexp.MustCompile(`%[-+# 0]*\d*(\.\d+)?[a-zA-Z%]`)
	spaced := func(s string) (lead, trail bool) {
		r := []rune(s)
		return unicode.IsSpace(r[0]), unicode.IsSpace(r[len(r)-1])
	}
	for en, zh := range table {
		if en == "" || strings.TrimSpace(zh) == "" {
			t.Errorf("%q has an empty key or translation", en)
			continue
		}
		if a, b := verb.FindAllString(en, -1), verb.FindAllString(zh, -1); !slices.Equal(a, b) {
			t.Errorf("%q has the verbs %q, its translation %q has %q", en, a, zh, b)
		}
		el, et := spaced(en)
		zl, zt := spaced(zh)
		if el != zl || et != zt {
			t.Errorf("%q and its translation %q differ in the space around them", en, zh)
		}
	}
}
