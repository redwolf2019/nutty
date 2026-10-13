package settings

import (
	"testing"

	"github.com/egoist/mygo/ui"
	uixsettings "github.com/smallnest/uix/components/settings"
)

// TestAwaitingInputBadges checks that the two awaiting-input badges are two
// settings with a control each: the Agents page edits the agents' badge,
// and the View page the commands'. Under one shared key the Agents page
// drew the View page's setting, and the agents' badge had no control.
func TestAwaitingInputBadges(t *testing.T) {
	s := Default()
	s.Language = "english"
	for _, k := range []struct {
		key   string
		field *bool
	}{
		{"privilege-badge-agent-awaiting-input", &s.AgentBadgeAwaitingInput},
		{"detect-awaiting-input", &s.BadgeAwaitingInput},
	} {
		d := defByKey(k.key)
		if d == nil {
			t.Errorf("no setting under %s", k.key)
		} else if d.ptr(s).(*bool) != k.field {
			t.Errorf("%s edits another setting's field (%q)", k.key, d.label)
		}
	}
	for page, label := range map[string]string{
		"agents": "Badge When Awaiting Input",
		"view":   "When Command Awaits Input",
	} {
		sw := &settingsWindow{app: mockApp{}, s: s, last: *s,
			state: uixsettings.SettingsState{Category: page}}
		tt := ui.NewTester(sw.view, 900, 1600)
		tt.Frame()
		if !tt.HasText(label) {
			t.Errorf("the %s page lacks %q", page, label)
		}
	}
}
