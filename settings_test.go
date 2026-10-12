package main

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/egoist/mygo/plugins/terminal"
)

// TestSettingsJSON checks that the settings survive a save and load, and
// that a file with unknown fields still reads.
func TestSettingsJSON(t *testing.T) {
	s := defaultSettings()
	s.ThemeName = "TokyoNight"
	s.FontSize = 15
	s.WindowOpacity = 0.8
	s.Foreground = "#123456"
	s.ConfirmCloseTab = true
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var back settings
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(s, &back) {
		t.Fatalf("round trip changed settings:\n%+v\n%+v", s, &back)
	}
	// A hand-edited file, with an unknown key, still loads.
	if err := json.Unmarshal([]byte(`{"FontSize":14,"futureSetting":true}`), &back); err != nil {
		t.Fatal(err)
	}
	if back.FontSize != 14 || back.FontFamily != "JetBrains Mono" {
		t.Fatalf("merged wrong: %+v", back)
	}
}

// TestSettingsApply checks the pieces the settings turn into terminal
// options: the font, the cursor, the scrollback.
func TestSettingsApply(t *testing.T) {
	a := &app{settings: defaultSettings()}

	f := a.termFont()
	if f.Family != "JetBrains Mono" || f.Size != 13 {
		t.Fatalf("default font = %+v", f)
	}
	if len(f.Features) != 1 || f.Features[0] != "dlig" {
		t.Fatalf("default ligatures = %v, want [dlig]", f.Features)
	}
	a.settings.FontLigatures = false
	if f := a.termFont(); len(f.Features) != 3 || f.Features[0] != "-calt" {
		t.Fatalf("ligatures off = %v", f.Features)
	}
	a.settings.FontLigatures = true
	a.settings.FontSize = 16
	a.settings.FontThicken = true
	f = a.termFont()
	if f.Size != 16 || !f.Thicken {
		t.Fatalf("font options = %+v", f)
	}

	if a.cursorStyle() != terminal.CursorBlock {
		t.Fatal("default cursor should be block")
	}
	a.settings.CursorStyle = "bar"
	if a.cursorStyle() != terminal.CursorBar {
		t.Fatal("bar cursor")
	}

	if n := a.scrollbackBytes(); n != 1_000_000 { // 10000 lines
		t.Fatalf("scrollback bytes = %d, want 1,000,000", n)
	}
	a.settings.ScrollbackLines = 0
	if a.scrollbackBytes() != -1 {
		t.Fatal("zero lines should turn the scrollback off")
	}
}

// TestThemeNames checks the theme lists the settings select from.
func TestThemeNames(t *testing.T) {
	names := themeNames()
	if len(names) != 1+2*len(curatedThemes) {
		t.Fatalf("theme names = %d, want %d", len(names), 1+2*len(curatedThemes))
	}
	if names[0] != "Follow system" {
		t.Fatalf("first theme = %q, want Follow system", names[0])
	}
}
