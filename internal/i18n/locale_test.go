package i18n

import "testing"

// TestSystemLocale checks how the "system" Language setting resolves: a
// Chinese locale in any of its spellings is Chinese, another locale is
// English, and a locale that cannot be read is Chinese, as SetLang
// promises, not the empty language that Tr would render in English.
func TestSystemLocale(t *testing.T) {
	read := systemLocale
	t.Cleanup(func() {
		systemLocale = read
		SetLang("english")
	})
	for _, c := range []struct {
		name   string
		locale func() string
		want   string
	}{
		{"zh-Hans-CN", func() string { return "zh-Hans-CN" }, "设置"},
		{"zh_TW", func() string { return "zh_TW" }, "设置"},
		{"en-US", func() string { return "en-US" }, "Settings"},
		{"empty", func() string { return "" }, "Settings"},
		{"panic", func() string { panic("mygo: App.Locale called before App.Run") }, "设置"},
	} {
		systemLocale = c.locale
		SetLang("system")
		if got := Tr("Settings"); got != c.want {
			t.Errorf("%s: Tr(Settings) = %q, want %q", c.name, got, c.want)
		}
	}
}
