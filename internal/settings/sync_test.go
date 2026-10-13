package settings

import (
	"testing"

	"github.com/egoist/mygo/ui"
	uixsettings "github.com/smallnest/uix/components/settings"
)

// savingApp is a host that counts what the settings window asks of it.
type savingApp struct{ applied, saved int }

func (a *savingApp) ApplySettings() { a.applied++ }
func (a *savingApp) SaveSettings()  { a.saved++ }
func (a *savingApp) ApplyKeybinds() {}

// TestSyncSavesInPlaceEdits checks that the window applies and saves an
// edit made inside a slice or a map of the settings, the way the pages
// make them: a profile's field through its pointer, a details view's
// switch, a recorded key binding, an Extra key. A snapshot that shares
// those with the settings never sees them change.
func TestSyncSavesInPlaceEdits(t *testing.T) {
	s := Default()
	s.Language = "english"
	s.Profiles = []Profile{{Name: "Web"}}
	s.KeybindOverrides = map[string]string{}
	s.Extra = map[string]any{}
	host := &savingApp{}
	sw := &settingsWindow{app: host, s: s, last: *s,
		state: uixsettings.SettingsState{Category: "general"}}
	tt := ui.NewTester(sw.view, 760, 540)
	tt.Frame()
	tt.Frame()
	if host.saved != 0 {
		t.Fatalf("an unchanged window saved %d times", host.saved)
	}
	edits := []struct {
		name string
		edit func()
	}{
		{"a profile's command", func() { s.Profiles[0].Command = "zsh" }},
		{"a profile's environment", func() { s.Profiles[0].Env = []string{"A=1"} }},
		{"a details view's switch", func() { s.DetailsViews[1].Enabled = false }},
		{"the details views' order", func() { sw.moveViewByID("info", 1) }},
		{"a recorded key binding", func() { s.KeybindOverrides["cmd+shift+k"] = "new_tab" }},
		{"a cleared key binding", func() { delete(s.KeybindOverrides, "cmd+shift+k") }},
		{"an Extra list", func() { s.Extra["tab-rules"] = []string{"x"} }},
	}
	for i, e := range edits {
		e.edit()
		tt.Frame()
		if host.saved != i+1 || host.applied != i+1 {
			t.Fatalf("%s: saved %d and applied %d times in all, want %d", e.name, host.saved, host.applied, i+1)
		}
		// The edit is the snapshot now: a frame without one saves nothing.
		tt.Frame()
		if host.saved != i+1 {
			t.Fatalf("%s: a frame without an edit saved again", e.name)
		}
	}
}
