package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// settings holds the app's preferences: the otty settings this app
// implements, saved as JSON under the user's Application Support folder.
// Field names are the otty config keys in camel case, so the All Settings
// list in the Advanced tab can drive every one of them; the defaults are
// otty's own, read from its config. Fields without a dedicated control
// still persist and show there.
type settings struct {
	// Appearance — themes, fonts, cursor, colors, opacity.
	ThemeName       string  // "Paper", otty's default light theme
	ThemeDark       string  // "Nord", otty's default dark theme
	AutoThemeDark   bool    // a separate theme for dark windows
	FontFamily      string  // "JetBrains Mono", otty's font
	FontFamilyFallback string
	FontSize        float64 // in DIPs; 13, otty's default
	FontBlending    string  // "srgb-over" | "macos-like" | "linear" | "perceptual"
	TextBold        string  // "auto" | "off" | "primary-only" | "synthetic"
	TextItalic      string
	TextUnderline   bool
	TextBlink       bool
	FontLigatures   string  // "dlig" (otty's default) | "calt" | "off"
	FontLigaturesAlphabet bool
	FontThicken     bool
	LineHeight      float64 // 0 follows the font; else a row-height multiplier
	AdjustCellHeight float64 // pixels added to each row (negative shrinks)
	CursorStyle     string  // "block" | "bar" | "underline" | "block-hollow"
	CursorBlinkStyle string // "default_off" | "default_on" | "always_off" | "always_on"
	CursorAnimation string  // "off" | "smooth"
	CursorOpacity   float64
	WindowOpacity   float64
	TerminalOpacity float64 // the terminal background's opacity, 0.5..1
	Background      string  // terminal colors; "" uses the theme's
	Foreground      string
	CursorColor     string
	CursorText      string
	Selection       string
	SelectionText   string
	Palette         [16]string // palette-0..15; "" uses the theme's
	BoldColor       string     // "none" | "bright" | a color
	FaintOpacity    float64
	MinimumContrast float64

	// Layout — where the tabs sit, and the window's size.
	WindowLayout         string // "sidebar-left" | "tabs-top" | "tabs-bottom"
	SidebarVisible       bool
	SidebarWidth         float64 // 220 on macOS, otty's rail
	DetailsPanelWidth    float64 // 220 on macOS
	WindowSize           string  // "remember" | "frame" | "grid"
	WindowWidth          float64
	WindowHeight         float64
	WindowCols           float64
	WindowRows           float64
	UnfocusedSplitOpacity float64
	ContainerPadding     float64

	// Controls — keyboard, mouse, open-with, clipboard, scroll.
	OptionAsAlt          string // "false" | "true" | "left" | "right"
	FocusFollowsMouse    bool
	RightClickAction     string // "context-menu" | "copy" | "paste" | "copy-or-paste" | "ignore"
	MouseHideWhileTyping bool
	MouseShiftToSelect   bool
	BypassMouseReporting string // "none" | "shift" | "ctrl" | "alt" | "ctrl-shift" | "super"
	LinkClickOverMouseMode bool
	CursorClickToMove    bool
	LinkOpenWith         string // "browser" | "otty"
	FileOpenWith         string // "default-app" | "otty"
	FolderOpenWith       string
	DefaultGitClient     string
	OpenWithApps         []openWithApp
	LinkSchemes          string // "all" | "custom"
	LinkSchemesCustom    []string
	LinkPreviews         bool
	AutoSecureInput      bool
	SecureInputIndication bool
	ScrollPastEnd        string // "disabled" | "last-line-with-content" | "last-line-in-middle" | "cursor-line"
	ScrollPastFirstLine  string // "disabled" | "same-as-last-line" | "first-line-with-content" | "first-line-in-middle"
	ScrollSmooth         bool
	CopyOnSelect         bool
	ClipboardTrimTrailingSpaces bool
	ClipboardPasteProtection    bool
	ClipboardPasteBracketedSafe bool
	SelectionClearOnTyping      bool
	SelectionClearOnCopy        bool
	SelectionBackspaceDeletes   bool
	CommandHistoryScope  string // "per-pane" | "shared" | "follow-shell"
	HistoryUpArrow       string // "auto" | "sequential" | "prefix"
	ScrollbackLines      float64 // otty's default is 10000

	// Shell — what new sessions run, and the integrations.
	Command          string // "" runs the user's shell
	WorkingDirectory string // "" uses the home directory
	Env              []string
	Term             string // "auto", or a TERM value
	ShellIntegration bool
	SSHIntegration   bool
	SessionRestoreProcesses string // "none" | "whitelist" | "all"
	SessionRestoreBanner    bool
	CommandWhitelist        []string
	SoundShell       bool
	SoundOnError     bool
	NotifyShellApp   bool // shell apps may send system notifications
	NotifyOnFinish   bool
	NotifyOnError    bool
	NotifyForeground string // "off" | "tab-unfocused" | "always"

	// General — app behavior.
	Language                     string // "system" | "english" | "chinese"
	OnLaunch                     string // "new_window" | "restore_session"
	QuitAfterLastWindowClosed    bool
	NewWindowWhenAllWindowsClosed bool
	NewTabPosition               string // "end" | "auto" | "after-current"
	ConfirmCloseTab              string // "always" | "process" | "never"
	ConfirmCloseWindow           string
	ConfirmClosePane             string
	WindowHideDirtyIndicator     bool

	// Editor — the file pane.
	EditorLineWrap              bool
	EditorTabSize               float64
	EditorVisibleWhitespace     bool
	EditorShowLineNumbers       bool
	EditorDefaultToPreviewReadonly bool
	EditorScrollPastEnd         bool
	EditorVim                   bool

	// View — tab titles, badges, the details panel.
	TitleFromOsc      bool
	TabIconPlacement  string // "combined" | "separate"
	BadgeExitStatus   bool
	BadgeActivity     bool
	BadgeAwaitingInput bool
	DetailsViews      []detailsView
	WebPanePersistCookies bool

	// Key Bindings — trigger overrides on otty's default set.
	KeybindOverrides map[string]string

	// Profiles — what new panes launch, and the default one.
	Profiles       []profile
	DefaultProfile string

	// Recipes.
	RecipeReplaySaved string // "auto" | "ask_once" | "manually"
	RecipeReplayFile  string

	// Agents — badge and notification behavior.
	AgentBadgeProcessing   bool
	AgentBadgeTaskComplete bool
	AgentBadgeAwaitingInput bool
	AgentNotifyTaskComplete bool
	AgentNotifyAwaitingInput bool
	AgentPreventSleep      bool
	AgentResumeSession     bool

	// Advanced.
	DebugMode bool

	// Extra holds the config keys that have no field of their own; the All
	// Settings list edits them, and they persist, even when this build does
	// not act on them.
	Extra map[string]any
}

// openWithApp is one app the "Open in…" menus offer.
type openWithApp struct {
	Name    string
	Bundle  string // a bundle id, or ""
	Command string // a TUI command, or ""
}

// detailsView is one tab of the details panel, in display order.
type detailsView struct {
	ID      string // "info" | "outline" | "git" | "files", or a custom id
	Kind    string // "builtin" | "tui" | "web"
	Title   string
	Command string
	URL     string
	Enabled bool
}

// profile decides what a new pane launches, as otty's profiles do.
type profile struct {
	Name        string
	Icon        string
	Command     string
	RunAfter    string
	CWD         string
	Env         []string
	Theme       string
	ThemeDark   string
	TitlePrefix string
}

// defaultSettings are otty's own defaults, read from its config: the
// values otty shows in a fresh install.
func defaultSettings() *settings {
	return &settings{
		ThemeName:       "Paper",
		ThemeDark:       "Nord",
		AutoThemeDark:   true,
		FontFamily:      "JetBrains Mono",
		FontSize:        13,
		FontBlending:    "srgb-over",
		TextBold:        "auto",
		TextItalic:      "auto",
		TextUnderline:   true,
		FontLigatures:   "dlig",
		FontThicken:     false,
		CursorStyle:     "block",
		CursorBlinkStyle: "default_off",
		CursorAnimation: "off",
		CursorOpacity:   1,
		WindowOpacity:   1,
		TerminalOpacity: 1,
		BoldColor:       "none",
		FaintOpacity:    0.5,
		MinimumContrast: 1,

		WindowLayout:         "sidebar-left",
		SidebarVisible:       true,
		SidebarWidth:         220, // the measured width of otty's tabs rail
		DetailsPanelWidth:    220,
		WindowSize:           "remember",
		WindowWidth:          1120,
		WindowHeight:         740,
		WindowCols:           80,
		WindowRows:           24,
		UnfocusedSplitOpacity: 0.6,

		OptionAsAlt:          "false",
		RightClickAction:     "context-menu",
		BypassMouseReporting: "shift",
		LinkClickOverMouseMode: true,
		CursorClickToMove:    true,
		LinkOpenWith:         "browser",
		FileOpenWith:         "default-app",
		FolderOpenWith:       "default-app",
		LinkSchemes:          "all",
		LinkPreviews:         true,
		AutoSecureInput:      true,
		SecureInputIndication: true,
		ScrollPastEnd:        "disabled",
		ScrollPastFirstLine:  "disabled",
		ScrollSmooth:         true,
		ClipboardPasteProtection:    true,
		ClipboardPasteBracketedSafe: true,
		SelectionClearOnTyping:      true,
		SelectionClearOnCopy:        true,
		SelectionBackspaceDeletes:   true,
		CommandHistoryScope:  "per-pane",
		HistoryUpArrow:       "auto",
		ScrollbackLines:      10000,

		ShellIntegration:      true,
		SSHIntegration:        true,
		SessionRestoreProcesses: "none",
		SessionRestoreBanner:  true,
		SoundShell:            true,
		NotifyShellApp:        true,
		NotifyOnError:         true,
		NotifyForeground:      "off",

		Language:             "system",
		OnLaunch:             "restore_session",
		NewTabPosition:       "auto",
		ConfirmCloseTab:      "process",
		ConfirmCloseWindow:   "process",
		ConfirmClosePane:     "process",

		EditorLineWrap:              true,
		EditorTabSize:               4,
		EditorShowLineNumbers:       true,
		EditorDefaultToPreviewReadonly: true,
		EditorScrollPastEnd:         true,

		TitleFromOsc:        true,
		TabIconPlacement:    "combined",
		BadgeExitStatus:     true,
		BadgeActivity:       true,
		BadgeAwaitingInput:  true,
		DetailsViews: []detailsView{
			{ID: "info", Kind: "builtin", Title: "Info", Enabled: true},
			{ID: "outline", Kind: "builtin", Title: "Outline", Enabled: true},
			{ID: "git", Kind: "builtin", Title: "Git", Enabled: true},
			{ID: "files", Kind: "builtin", Title: "Files", Enabled: true},
		},
		WebPanePersistCookies: true,

		RecipeReplaySaved: "ask_once",
		RecipeReplayFile:  "manually",

		AgentBadgeProcessing:   true,
		AgentBadgeTaskComplete: true,
		AgentBadgeAwaitingInput: true,
		AgentNotifyTaskComplete: true,
		AgentNotifyAwaitingInput: true,
		AgentResumeSession:     true,
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
	s.normalize()
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

// UnmarshalJSON reads a settings file into the current settings, turning
// the booleans older builds saved for the settings that are enums now
// (and the old CursorBlink name) into the strings normalize() understands,
// so an old file still reads instead of being dropped. The file overlays
// the current settings, as a plain unmarshal would.
func (s *settings) UnmarshalJSON(b []byte) error {
	file := map[string]any{}
	if err := json.Unmarshal(b, &file); err != nil {
		return err
	}
	str := func(v any) string {
		if b, ok := v.(bool); ok {
			return map[bool]string{true: "true", false: "false"}[b]
		}
		return ""
	}
	for _, k := range []string{"FontLigatures", "OptionAsAlt", "ConfirmCloseTab", "ConfirmCloseWindow", "ConfirmClosePane"} {
		if v, ok := file[k]; ok {
			if sv := str(v); sv != "" {
				file[k] = sv
			}
		}
	}
	if v, ok := file["CursorBlink"]; ok {
		if sv := str(v); sv != "" {
			file["CursorBlinkStyle"] = sv
		}
		delete(file, "CursorBlink")
	}
	// Overlay the file's keys on the current settings, then read the merged
	// JSON into an alias (no methods), so this unmarshal does not recurse.
	cur, err := json.Marshal(s)
	if err != nil {
		return err
	}
	merged := map[string]any{}
	if err := json.Unmarshal(cur, &merged); err != nil {
		return err
	}
	for k, v := range file {
		merged[k] = v
	}
	out, err := json.Marshal(merged)
	if err != nil {
		return err
	}
	type alias settings
	var a alias
	if err := json.Unmarshal(out, &a); err != nil {
		return err
	}
	*s = settings(a)
	return nil
}

// normalize makes an older or hand-edited settings file valid: it fills
// defaults that predate a field, and maps the booleans older builds saved
// for the settings that are enums now.
func (s *settings) normalize() {
	if s.FontLigatures == "" && s.FontFamily != "" {
		s.FontLigatures = "dlig"
	}
	if s.FontLigatures == "true" || s.FontLigatures == "false" {
		if s.FontLigatures == "true" {
			s.FontLigatures = "dlig"
		} else {
			s.FontLigatures = "off"
		}
	}
	if s.OptionAsAlt == "true" || s.OptionAsAlt == "false" || s.OptionAsAlt == "" {
		if s.OptionAsAlt == "true" {
			s.OptionAsAlt = "true"
		} else {
			s.OptionAsAlt = "false"
		}
	}
	if s.ConfirmCloseTab == "true" {
		s.ConfirmCloseTab = "process"
	}
	if s.ConfirmCloseTab == "false" {
		s.ConfirmCloseTab = "never"
	}
	if s.CursorBlinkStyle == "" {
		s.CursorBlinkStyle = "default_off"
	}
	if s.CursorBlinkStyle == "true" {
		s.CursorBlinkStyle = "default_on"
	}
	if s.CursorBlinkStyle == "false" {
		s.CursorBlinkStyle = "default_off"
	}
	if len(s.DetailsViews) == 0 {
		s.DetailsViews = defaultSettings().DetailsViews
	}
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

// customColors reports whether any terminal color is set by hand.
func (s *settings) customColors() bool {
	return s.Foreground != "" || s.Background != "" || s.CursorColor != "" ||
		s.CursorText != "" || s.Selection != "" || s.SelectionText != "" || s.anyPalette()
}

// anyPalette reports whether any of the sixteen palette colors is set.
func (s *settings) anyPalette() bool {
	for _, c := range s.Palette {
		if c != "" {
			return true
		}
	}
	return false
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
	// The palette overrides the theme's 16 colors one at a time.
	for i, c := range s.Palette {
		if c == "" {
			continue
		}
		light.Palette[i], dark.Palette[i] = ui.Hex(c), ui.Hex(c)
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

// themeNames lists the themes the settings select from: otty's own list,
// which Ghostty's themes cover, with follow-system first.
func themeNames() []string {
	names := []string{"Follow system"}
	for _, n := range ottyThemeNames {
		names = append(names, n)
	}
	return names
}

// ottyThemeNames are the themes otty ships, read from its bundle.
var ottyThemeNames = []string{
	"Acrylic Dark", "Glass Light", "Owl", "April", "Gruvbox Dark", "Paper",
	"April Dark", "Monokai Classic", "Pink", "Ayu Dark", "Newsprint", "Rosé Pine",
	"Ayu Light", "Night", "Seafoam Pastel", "Catppuccin Mocha", "Nord",
	"Solarized Dark", "Dracula", "One Dark", "Solarized Light", "Floating Card",
	"One Light", "Tokyo Night", "Glass Dark",
}

// colorHex is "#RRGGBB" of an opaque color, for the settings file.
func colorHex(c ui.Color) string { return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B) }

// envList joins the shell environment for the terminal plugin.
func (s *settings) envList() []string {
	if s == nil {
		return nil
	}
	out := []string{}
	for _, e := range s.Env {
		if strings.Contains(e, "=") {
			out = append(out, e)
		}
	}
	return out
}
