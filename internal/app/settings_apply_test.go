package app

import (
	"testing"

	"nutty/internal/settings"

	"github.com/egoist/mygo/plugins/terminal"
)

// TestSettingsApply checks the pieces the settings turn into terminal
// options: the font, the cursor, the scrollback.
func TestSettingsApply(t *testing.T) {
	a := &app{settings: settings.Default()}

	if a.settings.Language != "chinese" {
		t.Fatalf("default language = %q, want chinese", a.settings.Language)
	}

	f := a.termFont()
	if f.Family != "JetBrains Mono" || f.Size != 13 {
		t.Fatalf("default font = %+v", f)
	}
	if len(f.Features) != 1 || f.Features[0] != "dlig" {
		t.Fatalf("default ligatures = %v, want [dlig]", f.Features)
	}
	a.settings.FontLigatures = "off"
	if f := a.termFont(); len(f.Features) != 3 || f.Features[0] != "-calt" {
		t.Fatalf("ligatures off = %v", f.Features)
	}
	a.settings.FontLigatures = "dlig"
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
