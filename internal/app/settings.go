package app

import (
	"encoding/json"
	"os"
	"path/filepath"

	"nutty/internal/settings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// SaveSettings writes the preferences to the settings file. It writes a
// file beside it and renames that over it, so the settings file is whole
// at every moment: writing it in place empties it first, and a crash or a
// full disk then leaves a file that does not parse, which reads as the
// defaults at the next launch.
func (a *app) SaveSettings() {
	if a == nil || a.settings == nil {
		return
	}
	b, err := json.MarshalIndent(a.settings, "", "  ")
	if err != nil {
		return
	}
	path := settings.Path()
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	// A settings file that is a link, as into a dotfiles folder, stays
	// one: the rename replaces the file it points to.
	if target, err := filepath.EvalSymlinks(path); err == nil {
		path = target
	}
	f, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return
	}
	_, err = f.Write(b)
	if err == nil {
		err = f.Chmod(0o644) // a temporary file is 0600; the settings file is not
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(f.Name(), path)
	}
	if err != nil {
		_ = os.Remove(f.Name())
	}
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

// ApplySettings puts the settings into effect in every window: they are
// one set (see sharedSettings), so a change made from one window, or from
// its Settings window, shows in the others as well.
func (a *app) ApplySettings() {
	a.applySettings()
	appsMu.Lock()
	var others []*app
	for _, o := range apps {
		if o != a && o.settings == a.settings {
			others = append(others, o)
		}
	}
	appsMu.Unlock()
	for _, o := range others {
		o.applySettings()
	}
}

// applySettings puts the settings into effect in this window: the
// terminals' theme and font, the window's opacity and size, and a redraw.
// The creation-time options (cursor, scrollback, Option-as-Alt) reach new
// sessions, in newTerm.
func (a *app) applySettings() {
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
// for the terminals to follow the window's appearance. The default
// profile's themes, when it picks any, take the place of the settings'
// whole, as otty's profiles do. They are decided here, for new panes and
// for the ones a change of settings or of appearance restyles alike: a
// new pane used to get the profile's themes and lose them again on the
// next restyle.
func (a *app) termThemes() (light, dark *terminal.Theme) {
	if pr := a.settings.Profile(); pr != nil && (pr.Theme != "" || pr.ThemeDark != "") {
		light = settings.ThemeFor(pr.Theme, false)
		dark = settings.ThemeFor(pr.ThemeDark, true)
		if pr.ThemeDark == "" {
			dark = settings.ThemeFor(pr.Theme, true)
		}
		return light, dark
	}
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
