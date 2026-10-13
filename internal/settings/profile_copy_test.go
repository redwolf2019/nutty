package settings

import (
	"testing"

	"nutty/internal/i18n"

	"github.com/egoist/mygo/ui"
	uixsettings "github.com/smallnest/uix/components/settings"
)

// TestProfileDuplicateSuffix checks that a duplicated profile is named
// in the interface language. The suffix goes through Tr and had no
// Chinese entry, so a Chinese window named the copy "Web copy".
func TestProfileDuplicateSuffix(t *testing.T) {
	t.Cleanup(func() { i18n.SetLang("english") })
	for lang, want := range map[string]string{"chinese": "Web 副本", "english": "Web copy"} {
		s := Default()
		s.Language = lang
		s.Profiles = []Profile{{Name: "Web"}, {Name: "API"}}
		sw := &settingsWindow{app: mockApp{}, s: s, last: *s,
			state: uixsettings.SettingsState{Category: "profiles"}}
		tt := ui.NewTester(sw.view, 760, 1600)
		tt.Frame()
		if err := tt.Click(i18n.Tr("Duplicate")); err != nil {
			t.Fatalf("%s: %v, texts: %v", lang, err, tt.Texts())
		}
		if n := len(s.Profiles); n != 3 {
			t.Fatalf("%s: profiles = %d, want 3", lang, n)
		}
		if got := s.Profiles[2].Name; got != want {
			t.Errorf("%s: duplicate is named %q, want %q", lang, got, want)
		}
	}
}
