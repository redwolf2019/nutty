package settings

import (
	"testing"

	"nutty/internal/i18n"

	"github.com/egoist/mygo/ui"
	uixsettings "github.com/smallnest/uix/components/settings"
)

// TestSweepTranslations renders every settings category with Chinese
// selected and recording on, then fails on the strings still lacking a
// translation. The table in the i18n package is filled until only the
// exempt strings remain: settings keys, theme names, kept-English terms
// and 中文.
func TestSweepTranslations(t *testing.T) {
	i18n.SetLang("chinese")
	// The language is the process's: put English back for the other tests.
	t.Cleanup(func() { i18n.SetLang("english") })
	i18n.SetRecording(true)
	defer i18n.SetRecording(false)
	s := Default()
	for _, cat := range settingsCategories {
		sw := &settingsWindow{app: mockApp{}, s: s, last: *s,
			state: uixsettings.SettingsState{Category: cat.id}}
		if cat.id == "profiles" {
			sw.profileSel = 0
		}
		tt := ui.NewTester(sw.view, 760, 1600)
		tt.Frame()
	}
	for _, m := range i18n.Untranslated() {
		if !sweepExempt(m) {
			t.Errorf("missing translation for %q", m)
		}
	}
}

// sweepExempt reports the strings that legitimately have no translation:
// empty and already-Chinese strings, settings keys (lowercase kebab),
// theme names (proper nouns) and the terms Nutty keeps in English on
// purpose (Git, Shell, Shift, Term, Recipes, key names).
func sweepExempt(s string) bool {
	switch s {
	case "", "中文", "Nutty", "Otty", "Git", "Nord", "Paper", "Recipes",
		"Shell", "Shift", "Term", "Alt", "Ctrl", "Ctrl+Shift", "Super":
		return true
	}
	for _, n := range OttyThemeNames {
		if s == n {
			return true
		}
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return false
		}
	}
	return s != ""
}
