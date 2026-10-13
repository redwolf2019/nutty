package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	uixsettings "github.com/smallnest/uix/components/settings"
)

// menuApp is a host that counts the menu rebuilds the window asks for.
type menuApp struct{ keybinds int }

func (*menuApp) ApplySettings()   {}
func (*menuApp) SaveSettings()    {}
func (a *menuApp) ApplyKeybinds() { a.keybinds++ }

// allSettingsRow is the Advanced page with the All Settings list narrowed
// to one key, so the row's editor is the only one of its kind drawn.
func allSettingsRow(s *Settings, key string) *ui.Tester {
	s.Language = "english"
	sw := &settingsWindow{app: mockApp{}, s: s, last: *s,
		state: uixsettings.SettingsState{Category: "advanced"}}
	sw.settingsSearch = key
	tt := ui.NewTester(sw.view, 900, 1200)
	tt.Frame()
	return tt
}

// TestExtraEditorsEdit checks that the All Settings editors of the keys
// without a field change the key: the switch, the number's step button and
// the text field each edit a variable of the frame that draws them, and
// the edit has to reach Settings.Extra from there.
func TestExtraEditorsEdit(t *testing.T) {
	s := Default()
	tt := allSettingsRow(s, "auto-update")
	if err := tt.Click("auto-update"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if got := s.Extra["auto-update"]; got != true {
		t.Errorf("after the switch, auto-update = %v, want true", got)
	}

	s = Default()
	tt = allSettingsRow(s, "max-fps")
	for range 2 {
		if err := tt.Click("Increase"); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
	}
	if got := s.Extra["max-fps"]; got != 2.0 {
		t.Errorf("after two steps up, max-fps = %v, want 2", got)
	}

	s = Default()
	tt = allSettingsRow(s, "ui-accent")
	r, ok := tt.Find("ui-accent")
	if !ok {
		t.Fatal("no ui-accent row")
	}
	tt.ClickAt(r.X+40, r.Y+r.H+14) // the field under its label
	tt.Type("ab")
	tt.Frame()
	tt.Type("c")
	tt.Frame()
	if got := s.Extra["ui-accent"]; got != "abc" {
		t.Errorf("after typing, ui-accent = %v, want abc", got)
	}
}

// TestReloadRebuildsTheMenu checks that Reload hands the file's key
// bindings to the menu, as Reset does: the accelerators are the menu's,
// and only a rebuild shows a binding the file changed.
func TestReloadRebuildsTheMenu(t *testing.T) {
	t.Setenv("HOME", t.TempDir()) // never the user's own settings file
	file := Default()
	file.Language = "english"
	file.KeybindOverrides = map[string]string{"cmd+shift+k": "new_tab"}
	b, err := json.Marshal(file)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(Path()), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(), b, 0o644); err != nil {
		t.Fatal(err)
	}
	s := Default()
	s.Language = "english" // as the file's: a change of language rebuilds the menu by itself
	host := &menuApp{}
	sw := &settingsWindow{app: host, s: s, last: *s,
		state: uixsettings.SettingsState{Category: "advanced"}}
	tt := ui.NewTester(sw.view, 900, 800)
	tt.Frame()
	if err := tt.Click("Reload"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if s.KeybindOverrides["cmd+shift+k"] != "new_tab" {
		t.Fatalf("Reload did not read the file's key bindings: %v", s.KeybindOverrides)
	}
	if host.keybinds == 0 {
		t.Error("Reload read new key bindings and did not rebuild the menu")
	}
}

// TestNormalizeClosingBooleans checks that the booleans older builds saved
// for the three closing settings all read as the enums they are now, the
// window's and the pane's as the tab's.
func TestNormalizeClosingBooleans(t *testing.T) {
	s := LoadFrom([]byte(`{"ConfirmCloseTab": false, "ConfirmCloseWindow": true, "ConfirmClosePane": false}`))
	if s == nil {
		t.Fatal("the file did not load")
	}
	if s.ConfirmCloseTab != "never" || s.ConfirmCloseWindow != "process" || s.ConfirmClosePane != "never" {
		t.Errorf("closing settings = %q, %q, %q, want never, process, never",
			s.ConfirmCloseTab, s.ConfirmCloseWindow, s.ConfirmClosePane)
	}
}
