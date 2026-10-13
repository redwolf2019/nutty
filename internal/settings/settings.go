// Package settings holds the app's preferences: the otty settings this app
// implements, saved as JSON under the user's Application Support folder. The
// package renders the settings window (see NewWindow) and owns the registry
// of every modeled setting; the app keeps the methods that apply a change to
// the terminals and menus, through the App interface.
package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo/ui"
)

// Settings holds the app's preferences. Field names are the otty config
// keys in camel case, so the All Settings list in the Advanced tab can
// drive every one of them; the defaults are otty's own, read from its
// config. Fields without a dedicated control still persist and show there.
type Settings struct {
	// Appearance — themes, fonts, cursor, colors, opacity.
	ThemeName             string // "Paper", otty's default light theme
	ThemeDark             string // "Nord", otty's default dark theme
	AutoThemeDark         bool   // a separate theme for dark windows
	FontFamily            string // "JetBrains Mono", otty's font
	FontFamilyFallback    string
	FontSize              float64 // in DIPs; 13, otty's default
	FontBlending          string  // "srgb-over" | "macos-like" | "linear" | "perceptual"
	TextBold              string  // "auto" | "off" | "primary-only" | "synthetic"
	TextItalic            string
	TextUnderline         bool
	TextBlink             bool
	FontLigatures         string // "dlig" (otty's default) | "calt" | "off"
	FontLigaturesAlphabet bool
	FontThicken           bool
	LineHeight            float64 // 0 follows the font; else a row-height multiplier
	AdjustCellHeight      float64 // pixels added to each row (negative shrinks)
	CursorStyle           string  // "block" | "bar" | "underline" | "block-hollow"
	CursorBlinkStyle      string  // "default_off" | "default_on" | "always_off" | "always_on"
	CursorAnimation       string  // "off" | "smooth"
	CursorOpacity         float64
	WindowOpacity         float64
	TerminalOpacity       float64 // the terminal background's opacity, 0.5..1
	Background            string  // terminal colors; "" uses the theme's
	Foreground            string
	CursorColor           string
	CursorText            string
	Selection             string
	SelectionText         string
	Palette               [16]string // palette-0..15; "" uses the theme's
	BoldColor             string     // "none" | "bright" | a color
	FaintOpacity          float64
	MinimumContrast       float64

	// Layout — where the tabs sit, and the window's size.
	WindowLayout          string // "sidebar-left" | "tabs-top" | "tabs-bottom"
	SidebarVisible        bool
	SidebarWidth          float64 // 220 on macOS, otty's rail
	DetailsPanelWidth     float64 // 220 on macOS
	WindowSize            string  // "remember" | "frame" | "grid"
	WindowWidth           float64
	WindowHeight          float64
	WindowCols            float64
	WindowRows            float64
	UnfocusedSplitOpacity float64
	ContainerPadding      float64

	// Controls — keyboard, mouse, open-with, clipboard, scroll.
	OptionAsAlt                 string // "false" | "true" | "left" | "right"
	FocusFollowsMouse           bool
	RightClickAction            string // "context-menu" | "copy" | "paste" | "copy-or-paste" | "ignore"
	MouseHideWhileTyping        bool
	MouseShiftToSelect          bool
	BypassMouseReporting        string // "none" | "shift" | "ctrl" | "alt" | "ctrl-shift" | "super"
	LinkClickOverMouseMode      bool
	CursorClickToMove           bool
	LinkOpenWith                string // "browser" | "otty"
	FileOpenWith                string // "default-app" | "otty"
	FolderOpenWith              string
	DefaultGitClient            string
	OpenWithApps                []OpenWithApp
	LinkSchemes                 string // "all" | "custom"
	LinkSchemesCustom           []string
	LinkPreviews                bool
	AutoSecureInput             bool
	SecureInputIndication       bool
	ScrollPastEnd               string // "disabled" | "last-line-with-content" | "last-line-in-middle" | "cursor-line"
	ScrollPastFirstLine         string // "disabled" | "same-as-last-line" | "first-line-with-content" | "first-line-in-middle"
	ScrollSmooth                bool
	CopyOnSelect                bool
	ClipboardTrimTrailingSpaces bool
	ClipboardPasteProtection    bool
	ClipboardPasteBracketedSafe bool
	SelectionClearOnTyping      bool
	SelectionClearOnCopy        bool
	SelectionBackspaceDeletes   bool
	CommandHistoryScope         string  // "per-pane" | "shared" | "follow-shell"
	HistoryUpArrow              string  // "auto" | "sequential" | "prefix"
	ScrollbackLines             float64 // otty's default is 10000

	// Shell — what new sessions run, and the integrations.
	Command                 string // "" runs the user's shell
	WorkingDirectory        string // "" uses the home directory
	Env                     []string
	Term                    string // "auto", or a TERM value
	ShellIntegration        bool
	SSHIntegration          bool
	SessionRestoreProcesses string // "none" | "whitelist" | "all"
	SessionRestoreBanner    bool
	CommandWhitelist        []string
	SoundShell              bool
	SoundOnError            bool
	NotifyShellApp          bool // shell apps may send system notifications
	NotifyOnFinish          bool
	NotifyOnError           bool
	NotifyForeground        string // "off" | "tab-unfocused" | "always"

	// General — app behavior.
	Language                      string // "system" | "english" | "chinese"; default chinese
	OnLaunch                      string // "new_window" | "restore_session"
	QuitAfterLastWindowClosed     bool
	NewWindowWhenAllWindowsClosed bool
	NewTabPosition                string // "end" | "auto" | "after-current"
	ConfirmCloseTab               string // "always" | "process" | "never"
	ConfirmCloseWindow            string
	ConfirmClosePane              string
	WindowHideDirtyIndicator      bool

	// Editor — the file pane.
	EditorLineWrap                 bool
	EditorTabSize                  float64
	EditorVisibleWhitespace        bool
	EditorShowLineNumbers          bool
	EditorDefaultToPreviewReadonly bool
	EditorScrollPastEnd            bool
	EditorVim                      bool

	// View — tab titles, badges, the details panel.
	TitleFromOsc          bool
	TabIconPlacement      string // "combined" | "separate"
	BadgeExitStatus       bool
	BadgeActivity         bool
	BadgeAwaitingInput    bool
	DetailsViews          []DetailsView
	WebPanePersistCookies bool

	// Key Bindings — trigger overrides on otty's default set.
	KeybindOverrides map[string]string

	// Profiles — what new panes launch, and the default one.
	Profiles       []Profile
	DefaultProfile string

	// Recipes.
	RecipeReplaySaved string // "auto" | "ask_once" | "manually"
	RecipeReplayFile  string

	// Agents — badge and notification behavior.
	AgentBadgeProcessing     bool
	AgentBadgeTaskComplete   bool
	AgentBadgeAwaitingInput  bool
	AgentNotifyTaskComplete  bool
	AgentNotifyAwaitingInput bool
	AgentPreventSleep        bool
	AgentResumeSession       bool

	// Advanced.
	DebugMode bool

	// Extra holds the config keys that have no field of their own; the All
	// Settings list edits them, and they persist, even when this build does
	// not act on them.
	Extra map[string]any
}

// OpenWithApp is one app the "Open in…" menus offer.
type OpenWithApp struct {
	Name    string
	Bundle  string // a bundle id, or ""
	Command string // a TUI command, or ""
}

// DetailsView is one tab of the details panel, in display order.
type DetailsView struct {
	ID      string // "info" | "outline" | "git" | "files", or a custom id
	Kind    string // "builtin" | "tui" | "web"
	Title   string
	Command string
	URL     string
	Enabled bool
}

// Profile decides what a new pane launches, as otty's profiles do.
type Profile struct {
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

// Default are otty's own defaults, read from its config: the values otty
// shows in a fresh install.
func Default() *Settings {
	return &Settings{
		ThemeName:        "Paper",
		ThemeDark:        "Nord",
		AutoThemeDark:    true,
		FontFamily:       "JetBrains Mono",
		FontSize:         13,
		FontBlending:     "srgb-over",
		TextBold:         "auto",
		TextItalic:       "auto",
		TextUnderline:    true,
		FontLigatures:    "dlig",
		FontThicken:      false,
		CursorStyle:      "block",
		CursorBlinkStyle: "default_off",
		CursorAnimation:  "off",
		CursorOpacity:    1,
		WindowOpacity:    1,
		TerminalOpacity:  1,
		BoldColor:        "none",
		FaintOpacity:     0.5,
		MinimumContrast:  1,

		WindowLayout:          "sidebar-left",
		SidebarVisible:        true,
		SidebarWidth:          220, // the measured width of otty's tabs rail
		DetailsPanelWidth:     220,
		WindowSize:            "remember",
		WindowWidth:           1120,
		WindowHeight:          740,
		WindowCols:            80,
		WindowRows:            24,
		UnfocusedSplitOpacity: 0.6,

		OptionAsAlt:                 "false",
		RightClickAction:            "context-menu",
		BypassMouseReporting:        "shift",
		LinkClickOverMouseMode:      true,
		CursorClickToMove:           true,
		LinkOpenWith:                "browser",
		FileOpenWith:                "default-app",
		FolderOpenWith:              "default-app",
		LinkSchemes:                 "all",
		LinkPreviews:                true,
		AutoSecureInput:             true,
		SecureInputIndication:       true,
		ScrollPastEnd:               "disabled",
		ScrollPastFirstLine:         "disabled",
		ScrollSmooth:                true,
		ClipboardPasteProtection:    true,
		ClipboardPasteBracketedSafe: true,
		SelectionClearOnTyping:      true,
		SelectionClearOnCopy:        true,
		SelectionBackspaceDeletes:   true,
		CommandHistoryScope:         "per-pane",
		HistoryUpArrow:              "auto",
		ScrollbackLines:             10000,

		ShellIntegration:        true,
		SSHIntegration:          true,
		SessionRestoreProcesses: "none",
		SessionRestoreBanner:    true,
		SoundShell:              true,
		NotifyShellApp:          true,
		NotifyOnError:           true,
		NotifyForeground:        "off",

		Language:           "chinese",
		OnLaunch:           "restore_session",
		NewTabPosition:     "auto",
		ConfirmCloseTab:    "process",
		ConfirmCloseWindow: "process",
		ConfirmClosePane:   "process",

		EditorLineWrap:                 true,
		EditorTabSize:                  4,
		EditorShowLineNumbers:          true,
		EditorDefaultToPreviewReadonly: true,
		EditorScrollPastEnd:            true,

		TitleFromOsc:       true,
		TabIconPlacement:   "combined",
		BadgeExitStatus:    true,
		BadgeActivity:      true,
		BadgeAwaitingInput: true,
		DetailsViews: []DetailsView{
			{ID: "info", Kind: "builtin", Title: "Info", Enabled: true},
			{ID: "outline", Kind: "builtin", Title: "Outline", Enabled: true},
			{ID: "git", Kind: "builtin", Title: "Git", Enabled: true},
			{ID: "files", Kind: "builtin", Title: "Files", Enabled: true},
		},
		WebPanePersistCookies: true,

		RecipeReplaySaved: "ask_once",
		RecipeReplayFile:  "manually",

		AgentBadgeProcessing:     true,
		AgentBadgeTaskComplete:   true,
		AgentBadgeAwaitingInput:  true,
		AgentNotifyTaskComplete:  true,
		AgentNotifyAwaitingInput: true,
		AgentResumeSession:       true,
	}
}

// Path is where the preferences are saved, as JSON.
func Path() string {
	dir, err := os.UserConfigDir() // ~/Library/Application Support
	if err != nil {
		return "settings.json"
	}
	return filepath.Join(dir, "Nutty", "settings.json")
}

// Load reads the saved preferences, or the defaults when there is no file
// yet. A file that does not parse is ignored.
func Load() *Settings {
	s := Default()
	if b, err := os.ReadFile(Path()); err == nil {
		_ = json.Unmarshal(b, s)
	}
	s.Normalize()
	return s
}

// UnmarshalJSON reads a settings file into the current settings, turning
// the booleans older builds saved for the settings that are enums now
// (and the old CursorBlink name) into the strings Normalize understands,
// so an old file still reads instead of being dropped. The file overlays
// the current settings, as a plain unmarshal would.
func (s *Settings) UnmarshalJSON(b []byte) error {
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
	type alias Settings
	var a alias
	if err := json.Unmarshal(out, &a); err != nil {
		return err
	}
	*s = Settings(a)
	return nil
}

// LoadFrom parses a settings file, defaulting the rest, or nil when the
// file does not parse.
func LoadFrom(b []byte) *Settings {
	s := Default()
	if err := json.Unmarshal(b, s); err != nil {
		return nil
	}
	s.Normalize()
	return s
}

// Normalize makes an older or hand-edited settings file valid: it fills
// defaults that predate a field, and maps the booleans older builds saved
// for the settings that are enums now.
func (s *Settings) Normalize() {
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
		s.DetailsViews = Default().DetailsViews
	}
}

// CustomColors reports whether any terminal color is set by hand.
func (s *Settings) CustomColors() bool {
	return s.Foreground != "" || s.Background != "" || s.CursorColor != "" ||
		s.CursorText != "" || s.Selection != "" || s.SelectionText != "" || s.AnyPalette()
}

// AnyPalette reports whether any of the sixteen palette colors is set.
func (s *Settings) AnyPalette() bool {
	for _, c := range s.Palette {
		if c != "" {
			return true
		}
	}
	return false
}

// ColorHex is "#RRGGBB" of an opaque color, for the settings file.
func ColorHex(c ui.Color) string { return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B) }

// EnvList joins the shell environment for the terminal plugin.
func (s *Settings) EnvList() []string {
	if s == nil {
		return nil
	}
	// The Shell page's Term is the TERM new sessions advertise. "auto", and
	// the empty value of older files, leave it to the terminal. It goes
	// first: the terminal takes the list in order, so a TERM among the
	// variables still has the last word.
	if term := strings.TrimSpace(s.Term); term != "" && term != "auto" {
		return append([]string{"TERM=" + term}, ValidEnv(s.Env)...)
	}
	return ValidEnv(s.Env)
}

// ValidEnv is the environment entries of a list that have a value, the
// "NAME=value" ones.
func ValidEnv(list []string) []string {
	out := []string{}
	for _, e := range list {
		if strings.Contains(e, "=") {
			out = append(out, e)
		}
	}
	return out
}

// Profile is the default profile, or nil when none is chosen or it no
// longer exists.
func (s *Settings) Profile() *Profile {
	if s == nil || s.DefaultProfile == "" {
		return nil
	}
	for i := range s.Profiles {
		if s.Profiles[i].Name == s.DefaultProfile {
			return &s.Profiles[i]
		}
	}
	return nil
}

// ShellArgs splits a command line into arguments, honoring single and
// double quotes, so "sh -c 'echo hi'" keeps its two words together.
func ShellArgs(s string) []string {
	var out []string
	var cur strings.Builder
	var quote rune
	for _, r := range s {
		switch {
		case quote == '\'' && r == '\'' || quote == '"' && r == '"':
			quote = 0
		case quote != 0:
			cur.WriteRune(r)
		case r == '\'' || r == '"':
			quote = r
		case r == ' ' || r == '\t':
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}
