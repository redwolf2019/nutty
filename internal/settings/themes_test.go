package settings

import (
	"testing"
)

// TestOfferedThemesResolve checks that every theme the selects offer has
// colors to draw: a Ghostty theme, under otty's name or the name Ghostty
// ships it by. Paper is the exception, the window's own colors. A name
// without a theme drew the window's colors silently, so picking it looked
// like a select that did nothing.
func TestOfferedThemesResolve(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // no theme files of the user's
	window := *OttyTermTheme(false)
	for _, name := range OttyThemeNames {
		if name == "Paper" {
			continue
		}
		if *ThemeFor(name, false) == window {
			t.Errorf("the selects offer %q, which has no theme of its own", name)
		}
	}
}

// TestThemeForKeepsOttyNames checks that the names older settings files
// hold still read: an otty name Ghostty spells otherwise draws Ghostty's
// theme, and a theme only otty has draws the window's colors, as before.
func TestThemeForKeepsOttyNames(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	window := OttyTermTheme(true)
	if th := ThemeFor("Tokyo Night", true); th.Background == window.Background {
		t.Error("Tokyo Night draws the window's colors, want Ghostty's TokyoNight")
	}
	for _, name := range []string{"Paper", "Night", "Glass Dark", "no such theme"} {
		if th := ThemeFor(name, true); th.Background != window.Background {
			t.Errorf("%q draws %v, want the window's own colors", name, th.Background)
		}
	}
}
