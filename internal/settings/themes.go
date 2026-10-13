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
	if name == "" {
		return OttyTermTheme(dark)
	}
	th, err := terminal.GhosttyTheme(name)
	if err != nil {
		return OttyTermTheme(dark)
	}
	return th
}

// ThemeNames lists the themes the settings select from: otty's own list,
// which Ghostty's themes cover, with follow-system first.
func ThemeNames() []string {
	names := []string{"Follow system"}
	return append(names, OttyThemeNames...)
}

// OttyThemeNames are the themes otty ships, read from its bundle.
var OttyThemeNames = []string{
	"Acrylic Dark", "Glass Light", "Owl", "April", "Gruvbox Dark", "Paper",
	"April Dark", "Monokai Classic", "Pink", "Ayu Dark", "Newsprint", "Rosé Pine",
	"Ayu Light", "Night", "Seafoam Pastel", "Catppuccin Mocha", "Nord",
	"Solarized Dark", "Dracula", "One Dark", "Solarized Light", "Floating Card",
	"One Light", "Tokyo Night", "Glass Dark",
}
