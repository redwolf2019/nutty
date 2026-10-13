package settings

import (
	"testing"

	"github.com/egoist/mygo/ui"
	uixsettings "github.com/smallnest/uix/components/settings"
)

// TestColorWellSetsUnsetColor checks that picking a color in the well of a
// color the settings leave to the theme makes the setting, and that the
// palette's reset hands the well back to the theme. It drives the picker
// itself: the well of the first palette color, then the Red swatch.
func TestColorWellSetsUnsetColor(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // no theme files of the user's
	s := Default()
	s.Language = "english"
	sw := &settingsWindow{app: mockApp{}, s: s, last: *s,
		state: uixsettings.SettingsState{Category: "appearance"}}
	tt := ui.NewTester(sw.view, 900, 3000)
	tt.Frame()
	tt.Frame()
	theme := sw.wells.state["palette-0"].color
	if err := tt.Click("palette-0"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Red"); err != nil {
		t.Fatalf("the picker did not open: %v", err)
	}
	tt.Frame()
	tt.Frame()
	if got := s.Palette[0]; got != "#EF4444" {
		t.Fatalf("after picking Red, palette-0 = %q, want #EF4444", got)
	}
	// The picker covers the reset below the grid until it closes.
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if err := tt.Click("Reset palette"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Frame()
	if got := s.Palette[0]; got != "" {
		t.Fatalf("after the reset, palette-0 = %q, want it unset", got)
	}
	if got := sw.wells.state["palette-0"].color; got != theme {
		t.Fatalf("after the reset, the well shows %v, want the theme's %v", got, theme)
	}
}

// TestColorWellThemeColors checks that the wells of the colors the settings
// leave to the theme show what the terminal draws for each: the theme's
// background for the background, and for the text under the cursor, which
// the terminal draws in the background when the theme names no color.
func TestColorWellThemeColors(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // no theme files of the user's
	s := Default()
	s.Language = "english"
	s.ThemeName = "Dracula"
	sw := &settingsWindow{app: mockApp{}, s: s, last: *s,
		state: uixsettings.SettingsState{Category: "appearance"}}
	tt := ui.NewTester(sw.view, 900, 3000)
	tt.Frame()
	tt.Frame()
	th := ThemeFor("Dracula", false)
	if th.Background == th.Foreground {
		t.Fatal("the theme's colors do not tell the wells apart")
	}
	cursorText := th.CursorText
	if cursorText.A == 0 {
		cursorText = th.Background
	}
	for key, want := range map[string]ui.Color{
		"foreground":  th.Foreground,
		"background":  th.Background,
		"cursor-text": cursorText,
	} {
		if got := sw.wells.state[key].color; got != want {
			t.Errorf("the %s well shows %v, want the theme's %v", key, got, want)
		}
	}
}
