package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"nutty/settings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// SaveSettings writes the preferences to the settings file.
func (a *app) SaveSettings() {
	if a == nil || a.settings == nil {
		return
	}
	b, err := json.MarshalIndent(a.settings, "", "  ")
	if err != nil {
		return
	}
	os.MkdirAll(filepath.Dir(settings.Path()), 0o755)
	os.WriteFile(settings.Path(), b, 0o644)
}

// sidebarWidth is the width of the tabs rail, from the settings.
func (a *app) sidebarWidth() float32 { return float32(a.settings.SidebarWidth) }

// detailsPanelWidth is the width of the right panel, from the settings.
func (a *app) detailsPanelWidth() float32 { return float32(a.settings.DetailsPanelWidth) }

// cursorStyle maps the setting to the terminal's cursor.
func (a *app) cursorStyle() terminal.CursorStyle {
	switch a.settings.CursorStyle {
	case "bar":
		return terminal.CursorBar
	case "underline":
		return terminal.CursorUnderline
	}
	return terminal.CursorBlock
}

// optionAsAlt maps the four-state setting to the terminal's Option key.
func (a *app) optionAsAlt() bool {
	return a.settings.OptionAsAlt == "true" || a.settings.OptionAsAlt == "left" ||
		a.settings.OptionAsAlt == "right"
}

// confirmClosingTab is whether closing a tab asks first, per the setting.
func (a *app) confirmClosingTab() bool { return a.settings.ConfirmCloseTab != "never" }

// scrollbackBytes is the scrollback the setting's lines mean, in the bytes
// the terminal plugin measures; zero lines turns the scrollback off.
func (a *app) scrollbackBytes() int {
	n := int(a.settings.ScrollbackLines)
	if n <= 0 {
		return -1
	}
	return n * 100
}

// ApplySettings puts the settings into effect: the terminals' theme and
// font, the window's opacity and size, and a redraw. The creation-time
// options (cursor, scrollback, Option-as-Alt) reach new sessions, in
// newTerm.
func (a *app) ApplySettings() {
	if a.settings == nil {
		return
	}
	a.applyTermTheme()
	a.applyFont()
	if a.win == nil {
		return
	}
	a.win.SetOpacity(a.settings.WindowOpacity)
	// Size follows the setting only when the setting itself changed (the
	// settings window edits it); a resize of the window by hand does not
	// rewrite the setting, so it does not snap the window back.
	if int(a.settings.WindowWidth) != a.winSizeW || int(a.settings.WindowHeight) != a.winSizeH {
		a.win.SetContentSize(int(a.settings.WindowWidth), int(a.settings.WindowHeight))
		a.winSizeW, a.winSizeH = int(a.settings.WindowWidth), int(a.settings.WindowHeight)
	}
	a.win.Invalidate()
}

// applyTermTheme sets the terminals' theme from the settings: the light
// theme, the dark theme when a separate one is chosen, with the hand-picked
// colors overlaid, or leaves them to follow the window when none is chosen.
func (a *app) applyTermTheme() {
	light, dark := a.termThemes()
	for _, s := range a.sessions {
		for _, p := range s.panes {
			if p.term != nil {
				p.term.SetTheme(light, dark)
			}
		}
	}
}

// termThemes is the pair of terminal themes the settings pick, or nil,nil
// for the terminals to follow the window's appearance.
func (a *app) termThemes() (light, dark *terminal.Theme) {
	lightName := a.settings.ThemeName
	darkName := a.settings.ThemeDark
	if !a.settings.AutoThemeDark || darkName == "" {
		darkName = lightName
	}
	if lightName == "" && darkName == "" && !a.settings.CustomColors() {
		return nil, nil
	}
	light = settings.ThemeFor(lightName, false)
	dark = settings.ThemeFor(darkName, true)
	// The hand-picked colors apply to both appearances.
	s := a.settings
	if s.Foreground != "" {
		light.Foreground, dark.Foreground = ui.Hex(s.Foreground), ui.Hex(s.Foreground)
	}
	if s.Background != "" {
		light.Background, dark.Background = ui.Hex(s.Background), ui.Hex(s.Background)
	}
	if s.CursorColor != "" {
		light.Cursor, dark.Cursor = ui.Hex(s.CursorColor), ui.Hex(s.CursorColor)
	}
	if s.CursorText != "" {
		light.CursorText, dark.CursorText = ui.Hex(s.CursorText), ui.Hex(s.CursorText)
	}
	if s.Selection != "" {
		light.Selection, dark.Selection = ui.Hex(s.Selection), ui.Hex(s.Selection)
	}
	if s.SelectionText != "" {
		light.SelectionText, dark.SelectionText = ui.Hex(s.SelectionText), ui.Hex(s.SelectionText)
	}
	// The palette overrides the theme's 16 colors one at a time.
	for i, c := range s.Palette {
		if c == "" {
			continue
		}
		light.Palette[i], dark.Palette[i] = ui.Hex(c), ui.Hex(c)
	}
	return light, dark
}

// setTermTheme picks a theme pair by name for both appearances; the palette
// and the View menu still use it. An empty name follows the window.
func (a *app) setTermTheme(name string) {
	if name == "" || name == "Follow system" {
		a.settings.ThemeName, a.settings.ThemeDark = "", ""
	} else {
		a.settings.ThemeName, a.settings.ThemeDark = name, name
	}
	a.ApplySettings()
	a.SaveSettings()
}
