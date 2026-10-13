package settings

import (
	"testing"

	"github.com/egoist/mygo/ui"
	uixsettings "github.com/smallnest/uix/components/settings"
)

// countingApp is a host that counts what the settings window asks of it.
// mockApp answers every call and counts none, so a test built on it
// cannot tell a window that applies and saves from one that does not.
type countingApp struct{ applied, saved, keybinds int }

func (a *countingApp) ApplySettings() { a.applied++ }
func (a *countingApp) SaveSettings()  { a.saved++ }
func (a *countingApp) ApplyKeybinds() { a.keybinds++ }

// TestWindowAppliesAndSaves checks the window's contract with its host: a
// frame without a change asks for nothing, a control's change is applied
// and saved once, a change of language rebuilds the menu as well, and
// Reset All Settings does all three and leaves the defaults.
func TestWindowAppliesAndSaves(t *testing.T) {
	s := Default()
	s.Language = "english" // the controls are found by their English labels
	host := &countingApp{}
	sw := &settingsWindow{app: host, s: s, last: *s,
		state: uixsettings.SettingsState{Category: "advanced"}}
	tt := ui.NewTester(sw.view, 900, 800)
	tt.Frame()
	tt.Frame()
	if *host != (countingApp{}) {
		t.Fatalf("an unchanged window asked its host for %+v", *host)
	}

	if err := tt.Click("Debug mode"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !s.DebugMode {
		t.Fatal("the switch did not turn Debug mode on")
	}
	if host.applied != 1 || host.saved != 1 || host.keybinds != 0 {
		t.Errorf("after one switch the host was asked for %+v, want one apply and one save", *host)
	}

	// The language select's choice lands in the setting; the menu's labels
	// are in the old language until it is rebuilt.
	s.Language = "chinese"
	tt.Frame()
	if host.applied != 2 || host.saved != 2 || host.keybinds != 1 {
		t.Errorf("after a change of language the host was asked for %+v, want a second apply and save and a menu rebuild", *host)
	}
	s.Language = "english"
	tt.Frame()

	s.FontSize = 20
	sw.settingsSearch = "no such key" // an empty All Settings list, so the Reset below it is in the window
	tt.Frame()
	before := *host
	if err := tt.Click("Reset All Settings"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if host.applied <= before.applied || host.saved <= before.saved || host.keybinds <= before.keybinds {
		t.Errorf("Reset All Settings asked the host for %+v after %+v, want an apply, a save and a menu rebuild", *host, before)
	}
	if s.FontSize != Default().FontSize || s.DebugMode {
		t.Errorf("after the reset FontSize is %v and Debug mode %v, want the defaults", s.FontSize, s.DebugMode)
	}
}

// TestKeybindConflictShown checks what TestKeybindRecord's comment
// promises and its body does not: a trigger recorded for one action is
// reported on the row of the action that has it by default, and the
// override's own row is there, marked, without a conflict.
func TestKeybindConflictShown(t *testing.T) {
	s := Default()
	s.Language = "english"
	s.KeybindOverrides = map[string]string{"cmd+d": "new_tab"} // Split Right's default
	sw := &settingsWindow{app: mockApp{}, s: s, last: *s,
		state: uixsettings.SettingsState{Category: "key_bindings"}}

	var row *keybindRow
	for _, r := range sw.keybindRows() {
		if r.action == "new_tab" {
			row = &r
		}
	}
	if row == nil || row.trigger != "cmd+d" || !row.overridden {
		t.Fatalf("the New Tab row is %+v, want it there with the recorded cmd+d", row)
	}

	sw.keys.search = "split right"
	tt := ui.NewTester(sw.view, 1000, 800)
	tt.Frame()
	if !tt.HasText("Conflicts with: New Tab") {
		t.Errorf("Split Right's row does not report the conflict; it shows %q", tt.Texts())
	}
	sw.keys.search = "new tab"
	tt.Frame()
	if tt.HasText("Conflicts with: New Tab") || tt.HasText("Conflicts with: Split Right") {
		t.Errorf("New Tab's own row reports a conflict; it shows %q", tt.Texts())
	}
}

// TestDataBreathCount checks, without a clock, what starts the Data icon's
// breath: each move onto the Data row counts one breath, and frames that
// stay on a row count none. TestDataBreath checks the breath's pixels,
// against the wall clock.
func TestDataBreathCount(t *testing.T) {
	s := Default()
	s.Language = "english"
	sw := &settingsWindow{app: mockApp{}, s: s, last: *s,
		state: uixsettings.SettingsState{Category: "appearance"}}
	tt := ui.NewTester(sw.view, 760, 800)
	tt.Frame()
	for i, click := range []string{"Data", "General", "Data"} {
		if err := tt.Click(click); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
		tt.Frame()
		if want := uint64(i/2 + 1); sw.dataBreath != want {
			t.Fatalf("after the click on %s (%d), %d breaths, want %d", click, i, sw.dataBreath, want)
		}
	}
}
