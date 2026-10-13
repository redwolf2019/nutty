package settings

import (
	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// OttyTheme returns the window theme with the colors otty draws: warm
// surfaces, a warm border, and a background that matches the terminal's.
// The values are measured from otty's window, not mygo's defaults.
func OttyTheme(dark bool) *ui.Theme {
	if dark {
		t := ui.DarkTheme()
		t.Background = ui.Hex("#1b1b1b")
		t.Surface = ui.Hex("#262626")
		t.SurfaceHover = ui.Hex("#303030")
		t.SurfacePressed = ui.Hex("#3a3a3a")
		t.Border = ui.Hex("#3a3a38")
		t.TextMuted = ui.Hex("#8a8a88")
		return t
	}
	t := ui.LightTheme()
	t.Background = ui.Hex("#FCFAF8") // the terminal's light background
	t.Surface = ui.Hex("#F4F2EE")
	t.SurfaceHover = ui.Hex("#EAE8E1")
	t.SurfacePressed = ui.Hex("#E0DED5")
	t.Border = ui.Hex("#E0DED5")
	t.Text = ui.Hex("#3D3D3B")
	t.TextMuted = ui.Hex("#9D9B92")
	return t
}

// OttyTermTheme is the terminal palette of a window that follows the
// system appearance: the window's own background and text, so the
// terminal blends with the strip above it, as otty's does.
func OttyTermTheme(dark bool) *terminal.Theme {
	if dark {
		return &terminal.Theme{Foreground: ui.DarkTheme().Text, Background: ui.Hex("#1b1b1b"), Palette: terminal.DarkTheme().Palette}
	}
	return &terminal.Theme{Foreground: ui.Hex("#3D3D3B"), Background: ui.Hex("#FCFAF8"), Palette: terminal.LightTheme().Palette}
}

// ThemeFor is a Ghostty theme by name, or the window's own colors when the
// name is empty or unknown.
func ThemeFor(name string, dark bool) *terminal.Theme {
	th, err := ghosttyTheme(name)
	if err != nil {
		return OttyTermTheme(dark)
	}
	return th
}

// ghosttyTheme is the Ghostty theme of an otty theme name. The name itself
// goes first, so a theme file of the user's under otty's name wins; the
// themes Ghostty ships under another name are found by that one.
func ghosttyTheme(name string) (*terminal.Theme, error) {
	th, err := terminal.GhosttyTheme(name)
	if alias, ok := ghosttyNames[name]; ok && err != nil {
		return terminal.GhosttyTheme(alias)
	}
	return th, err
}

// ghosttyNames are the otty themes Ghostty ships under another name. The
// settings keep otty's name, the one the selects show and older files hold.
var ghosttyNames = map[string]string{
	"Ayu Dark":        "Ayu",
	"One Dark":        "Atom One Dark",
	"One Light":       "Atom One Light",
	"Rosé Pine":       "Rose Pine",
	"Solarized Dark":  "iTerm2 Solarized Dark",
	"Solarized Light": "iTerm2 Solarized Light",
	"Tokyo Night":     "TokyoNight",
}

// ThemeNames lists the themes the settings select from: otty's own list,
// as far as Ghostty's themes cover it, with follow-system first.
func ThemeNames() []string {
	names := []string{"Follow system"}
	return append(names, OttyThemeNames...)
}

// OttyThemeNames are the themes of otty's bundle this build can draw: the
// ones Ghostty ships too, by otty's name. Paper, otty's default, has no
// Ghostty theme, and the window's own colors, measured from otty under
// Paper, stand for it. otty's other themes of its own (Acrylic Dark,
// April, April Dark, Floating Card, Glass Dark, Glass Light, Newsprint,
// Night, Pink) are left out: picking one changed nothing, as ThemeFor has
// no colors for them. A settings file that names one still reads, and
// draws the window's colors as it did.
var OttyThemeNames = []string{
	"Owl", "Gruvbox Dark", "Paper",
	"Monokai Classic", "Ayu Dark", "Rosé Pine",
	"Ayu Light", "Seafoam Pastel", "Catppuccin Mocha", "Nord",
	"Solarized Dark", "Dracula", "One Dark", "Solarized Light",
	"One Light", "Tokyo Night",
}
