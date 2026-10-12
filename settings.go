package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// settings holds the app's preferences: the otty settings this app
// implements, saved as JSON under the user's Application Support folder.
// Fields hold the values the settings window edits; the defaults below
// are otty's own.
type settings struct {
	// Appearance
	ThemeName       string  // "" follows the window appearance
	ThemeDark       string  // a separate theme for dark windows
	AutoThemeDark   bool    // use ThemeDark for dark windows
	FontFamily      string  // default JetBrains Mono, otty's font
	FontSize        float64 // in DIPs; 13, otty's default
	FontLigatures   bool    // otty's default is "dlig"
	FontThicken     bool    // a thicker stroke, as font-thicken
	AdjustCellHeight float64 // 0..0.4, rows this much taller
	CursorStyle     string  // "block", "bar" or "underline"
	CursorBlink     bool
	WindowOpacity   float64 // the whole window, 0..1
	TerminalOpacity float64 // the terminal's background, 0..1
	Background      string  // terminal colors; "" uses the theme's
	Foreground      string
	CursorColor     string
	CursorText      string
	Selection       string
	SelectionText   string

	// View
	SidebarVisible bool
	SidebarWidth   float64
	WindowWidth    float64
	WindowHeight   float64

	// Controls
	OptionAsAlt bool // Option types Meta to programs

	// Scroll
	ScrollbackLines float64 // otty's default is 10000

	// Shell
	Command          string // "" runs the user's shell
	WorkingDirectory string // "" uses the home directory

	// General
	ConfirmCloseTab bool
}

// defaultSettings are otty's own defaults, as its settings read.
func defaultSettings() *settings {
	return &settings{
		ThemeName:       "",
		ThemeDark:       "",
		AutoThemeDark:   false,
		FontFamily:      "JetBrains Mono",
		FontSize:        13,
		FontLigatures:   true,
		FontThicken:     false,
		AdjustCellHeight: 0,
		CursorStyle:     "block",
		CursorBlink:     false,
		WindowOpacity:   1,
		TerminalOpacity: 1,
		SidebarVisible:  true,
		SidebarWidth:    220, // the measured width of otty's tabs rail
		WindowWidth:     1120,
		WindowHeight:    740,
		OptionAsAlt:     false,
		ScrollbackLines: 10000,
		ConfirmCloseTab: false,
	}
}

// settingsPath is where the preferences are saved, as JSON.
func settingsPath() string {
	dir, err := os.UserConfigDir() // ~/Library/Application Support
	if err != nil {
		return "settings.json"
	}
	return filepath.Join(dir, "Nutty", "settings.json")
}

// loadSettings reads the saved preferences, or the defaults when there is
// no file yet. A file that does not parse is ignored.
func loadSettings() *settings {
	s := defaultSettings()
	if b, err := os.ReadFile(settingsPath()); err == nil {
		json.Unmarshal(b, s)
	}
	return s
}

func (a *app) saveSettings() {
	if a == nil || a.settings == nil {
		return
	}
	b, err := json.MarshalIndent(a.settings, "", "  ")
	if err != nil {
		return
	}
	os.MkdirAll(filepath.Dir(settingsPath()), 0o755)
	os.WriteFile(settingsPath(), b, 0o644)
}

// sidebarWidth is the width of the tabs rail, from the settings.
func (a *app) sidebarWidth() float32 { return float32(a.settings.SidebarWidth) }

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

// scrollbackBytes is the scrollback the setting's lines mean, in the bytes
// the terminal plugin measures; zero lines turns the scrollback off.
func (a *app) scrollbackBytes() int {
	n := int(a.settings.ScrollbackLines)
	if n <= 0 {
		return -1
	}
	return n * 100
}

// customColors reports whether any terminal color is set by hand.
func (s *settings) customColors() bool {
	return s.Foreground != "" || s.Background != "" || s.CursorColor != "" ||
		s.CursorText != "" || s.Selection != "" || s.SelectionText != ""
}

// applySettings puts the settings into effect: the terminals' theme and
// font, the window's opacity and size, and a redraw. The creation-time
// options (cursor, scrollback, Option-as-Alt) reach new sessions, in
// newTerm.
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
// for the terminals to follow the window's appearance.
func (a *app) termThemes() (light, dark *terminal.Theme) {
	lightName := a.settings.ThemeName
	darkName := a.settings.ThemeDark
	if !a.settings.AutoThemeDark || darkName == "" {
		darkName = lightName
	}
	if lightName == "" && darkName == "" && !a.settings.customColors() {
		return nil, nil
	}
	light = themeFor(lightName, false)
	dark = themeFor(darkName, true)
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
	return light, dark
}

// themeFor is a Ghostty theme by name, or the window's own colors when the
// name is empty or unknown.
func themeFor(name string, dark bool) *terminal.Theme {
	if name == "" {
		return ottyTermTheme(dark)
	}
	th, err := terminal.GhosttyTheme(name)
	if err != nil {
		return ottyTermTheme(dark)
	}
	return th
}

// setTermTheme picks a theme pair by name for both appearances; the palette
// and the View menu still use it. An empty name follows the window.
func (a *app) setTermTheme(name string) {
	if name == "" || name == "Follow system" {
		a.settings.ThemeName, a.settings.ThemeDark = "", ""
	} else {
		a.settings.ThemeName, a.settings.ThemeDark = name, name
	}
	a.applySettings()
	a.saveSettings()
}

// colorHex is "#RRGGBB" of an opaque color, for the settings file.
func colorHex(c ui.Color) string { return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B) }
