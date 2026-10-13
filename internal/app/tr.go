package app

import "nutty/internal/i18n"

// tr is a display string in the interface language (see the i18n
// package); the settings package calls i18n.Tr directly.
func tr(s string) string { return i18n.Tr(s) }

// setUILang resolves the Language setting to the interface language.
func setUILang(v string) { i18n.SetLang(v) }
