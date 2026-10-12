package main

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// The window appearance the terminals' colors are seen against.
func (a *app) appearanceSystem() { mygo.Theme.SetSource(mygo.ThemeSystem) }
func (a *app) appearanceLight()  { mygo.Theme.SetSource(mygo.ThemeLight) }
func (a *app) appearanceDark()   { mygo.Theme.SetSource(mygo.ThemeDark) }

// ottyTheme returns the window theme with the colors otty draws: warm
// surfaces, a warm border, and a background that matches the terminal's.
// The values are measured from otty's window, not mygo's defaults.
func ottyTheme(dark bool) *ui.Theme {
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

// ottyTermTheme is the terminal palette of a window that follows the
// system appearance: the window's own background and text, so the
// terminal blends with the strip above it, as otty's does.
func ottyTermTheme(dark bool) *terminal.Theme {
	if dark {
		return &terminal.Theme{Foreground: ui.DarkTheme().Text, Background: ui.Hex("#1b1b1b"), Palette: terminal.DarkTheme().Palette}
	}
	return &terminal.Theme{Foreground: ui.Hex("#3D3D3B"), Background: ui.Hex("#FCFAF8"), Palette: terminal.LightTheme().Palette}
}

// applyTheme re-applies the terminals' colors: after a change of
// appearance, so the "follow system" terminals keep blending with the
// strip, as the window theme changes.
func (a *app) applyTheme() {
	a.applyTermTheme()
}
