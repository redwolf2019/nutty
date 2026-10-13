package main

import (
	"github.com/egoist/mygo"
)

// The window appearance the terminals' colors are seen against.
func (a *app) appearanceSystem() { mygo.Theme.SetSource(mygo.ThemeSystem) }
func (a *app) appearanceLight()  { mygo.Theme.SetSource(mygo.ThemeLight) }
func (a *app) appearanceDark()   { mygo.Theme.SetSource(mygo.ThemeDark) }

// applyTheme re-applies the terminals' colors: after a change of
// appearance, so the "follow system" terminals keep blending with the
// strip, as the window theme changes.
func (a *app) applyTheme() {
	a.applyTermTheme()
}
