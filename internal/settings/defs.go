package settings

import "fmt"

// The Settings registry: every setting this build models, with the label
// and description otty shows, and closures that read and write its field.
// The tab pages render from the groups below and the Advanced page's All
// Settings list renders the whole registry, so the two always agree. The
// labels are otty's own, from its localization files.

// settingDef is one configurable setting.
type settingDef struct {
	key   string // the otty config key, e.g. "cursor-style"
	label string
	desc  string
	kind  string // "bool" | "enum" | "number" | "color" | "string" | "list"
	enum  []enumOpt
	def   any
	lo    float64
	hi    float64
	step  float64
	ptr   func(*Settings) any // the field a control binds to, when direct
	get   func(*Settings) any
	set   func(*Settings, any)
}

// enumOpt is one choice of an enum setting: the stored value and the
// label otty shows for it.
type enumOpt struct {
	value string
	label string
}

// settingGroup is one titled section of a Settings page.
type settingGroup struct {
	title string
	keys  []string
}

func dBool(key, label, desc string, p func(*Settings) *bool, def bool) settingDef {
	return settingDef{key: key, label: label, desc: desc, kind: "bool", def: def,
		ptr: func(s *Settings) any { return p(s) },
		get: func(s *Settings) any { return *p(s) },
		set: func(s *Settings, v any) { *p(s) = v.(bool) }}
}

func dEnum(key, label, desc string, p func(*Settings) *string, def string, opts ...enumOpt) settingDef {
	return settingDef{key: key, label: label, desc: desc, kind: "enum", enum: opts, def: def,
		ptr: func(s *Settings) any { return p(s) },
		get: func(s *Settings) any { return *p(s) },
		set: func(s *Settings, v any) { *p(s) = v.(string) }}
}

func dNum(key, label, desc string, p func(*Settings) *float64, def, lo, hi, step float64) settingDef {
	return settingDef{key: key, label: label, desc: desc, kind: "number", def: def,
		lo: lo, hi: hi, step: step,
		ptr: func(s *Settings) any { return p(s) },
		get: func(s *Settings) any { return *p(s) },
		set: func(s *Settings, v any) { *p(s) = v.(float64) }}
}

func dColor(key, label, desc string, p func(*Settings) *string, def string) settingDef {
	return settingDef{key: key, label: label, desc: desc, kind: "color", def: def,
		ptr: func(s *Settings) any { return p(s) },
		get: func(s *Settings) any { return *p(s) },
		set: func(s *Settings, v any) { *p(s) = v.(string) }}
}

func dText(key, label, desc string, p func(*Settings) *string, def string) settingDef {
	return settingDef{key: key, label: label, desc: desc, kind: "string", def: def,
		ptr: func(s *Settings) any { return p(s) },
		get: func(s *Settings) any { return *p(s) },
		set: func(s *Settings, v any) { *p(s) = v.(string) }}
}

// dList is a list-of-strings setting, edited one item per line.
func dList(key, label, desc string, p func(*Settings) *[]string, def []string) settingDef {
	return settingDef{key: key, label: label, desc: desc, kind: "list", def: def,
		get: func(s *Settings) any { return *p(s) },
		set: func(s *Settings, v any) { *p(s) = v.([]string) }}
}

// en is a shorthand for an enum option pair.
func en(value, label string) enumOpt { return enumOpt{value: value, label: label} }

// settingsDefs is the registry of every modeled setting, in the order the
// All Settings list shows them.
var settingsDefs = []settingDef{
	// --- Appearance: theme (the theme names join in init; the palette too) ---
	dBool("auto-theme-dark-mode", "Use separated theme for dark mode", "Follow the system color scheme.", func(s *Settings) *bool { return &s.AutoThemeDark }, true),

	// --- Shell: environment ---
	dList("env", "Environment Variables", "Extra variables new sessions run with, one per line (NAME=value).", func(s *Settings) *[]string { return &s.Env }, nil),
	dList("command-whitelist", "Command Whitelist", "The commands re-run on restore and shown in the command list, one per line.", func(s *Settings) *[]string { return &s.CommandWhitelist }, nil),
	dList("link-scheme-allowlist", "Custom Link Schemes", "The URL schemes to make clickable when Auto-Detect is Custom, one per line.", func(s *Settings) *[]string { return &s.LinkSchemesCustom }, nil),

	// --- Appearance: font ---
	dText("font-family", "Font Family", "The font of the terminal text.", func(s *Settings) *string { return &s.FontFamily }, "JetBrains Mono"),
	dText("font-family-fallback", "Font Family Fallback", "Comma-separated; used when a glyph is missing.", func(s *Settings) *string { return &s.FontFamilyFallback }, ""),
	dNum("font-size", "Font Size", "The text size in DIPs.", func(s *Settings) *float64 { return &s.FontSize }, 13, 6, 40, 1),
	dEnum("font-blending", "Font Blending", "How text pixels blend onto the background.", func(s *Settings) *string { return &s.FontBlending }, "srgb-over",
		en("srgb-over", "Default"), en("macos-like", "macOS-like"), en("linear", "Linear"), en("perceptual", "Perceptual")),
	dEnum("text-bold", "Bold", "How bold text renders.", func(s *Settings) *string { return &s.TextBold }, "auto",
		en("auto", "Auto"), en("off", "Off"), en("primary-only", "Primary Only"), en("synthetic", "Synthetic")),
	dEnum("text-italic", "Italic", "How italic text renders.", func(s *Settings) *string { return &s.TextItalic }, "auto",
		en("auto", "Auto"), en("off", "Off"), en("primary-only", "Primary Only"), en("synthetic", "Synthetic")),
	dBool("text-underline", "Underline", "Draw underlines under underlined text.", func(s *Settings) *bool { return &s.TextUnderline }, true),
	dBool("text-blink", "Blink", "Allow blinking text.", func(s *Settings) *bool { return &s.TextBlink }, false),
	dEnum("font-ligatures", "Ligatures", "Which ligature set the font uses.", func(s *Settings) *string { return &s.FontLigatures }, "dlig",
		en("off", "Off"), en("calt", "Standard (calt)"), en("dlig", "Discretionary (dlig)")),
	dBool("font-thicken", "Thicker Text", "A thicker stroke, as font-thicken does on macOS.", func(s *Settings) *bool { return &s.FontThicken }, false),
	dNum("line-height", "Custom Line Height", "Scale row height by this multiplier (1.0 = font default).", func(s *Settings) *float64 { return &s.LineHeight }, 0, 0, 2, 0.05),
	dNum("adjust-cell-height", "Adjust Cell Height", "Add this many pixels to each row (negative shrinks).", func(s *Settings) *float64 { return &s.AdjustCellHeight }, 0, -0.5, 0.5, 0.05),

	// --- Appearance: cursor ---
	dColor("cursor-color", "Cursor color", "The color of the cursor.", func(s *Settings) *string { return &s.CursorColor }, ""),
	dColor("cursor-text", "Text color under cursor", "The color of text under the cursor.", func(s *Settings) *string { return &s.CursorText }, ""),
	dNum("cursor-opacity", "Cursor opacity", "How opaque the cursor is.", func(s *Settings) *float64 { return &s.CursorOpacity }, 1, 0.2, 1, 0.05),
	dEnum("cursor-style", "Cursor Style", "The shape of the cursor.", func(s *Settings) *string { return &s.CursorStyle }, "block",
		en("block", "Block"), en("bar", "Bar"), en("underline", "Underline"), en("block-hollow", "Block (hollow)")),
	dEnum("cursor-style-blink", "Cursor Blink", "Whether the cursor blinks, and whether programs may change it.", func(s *Settings) *string { return &s.CursorBlinkStyle }, "default_off",
		en("default_off", "Default off"), en("default_on", "Default on"), en("always_off", "Always off"), en("always_on", "Always on")),
	dEnum("cursor-animation", "Cursor Animation", "Smooth glides the caret on same-row moves.", func(s *Settings) *string { return &s.CursorAnimation }, "off",
		en("off", "Off"), en("smooth", "Smooth")),

	// --- Appearance: colors ---
	dColor("foreground", "Foreground", "The terminal's text color.", func(s *Settings) *string { return &s.Foreground }, ""),
	dColor("background", "Background", "The terminal's background color.", func(s *Settings) *string { return &s.Background }, ""),
	dColor("selection-background", "Selection", "The background of a selection.", func(s *Settings) *string { return &s.Selection }, ""),
	dColor("selection-foreground", "Selection text", "The text color of a selection.", func(s *Settings) *string { return &s.SelectionText }, ""),
	dEnum("bold-color", "Bold Color", "The color of bold text.", func(s *Settings) *string { return &s.BoldColor }, "none",
		en("none", "None"), en("bright", "Bright variant")),
	dNum("faint-opacity", "Faint Opacity", "How faint faint text renders.", func(s *Settings) *float64 { return &s.FaintOpacity }, 0.5, 0, 1, 0.05),
	dNum("minimum-contrast", "Minimum Contrast", "The minimum contrast ratio for text.", func(s *Settings) *float64 { return &s.MinimumContrast }, 1, 1, 21, 0.5),

	// --- Layout ---
	dEnum("window-layout", "Layout", "Where the tabs sit in the window.", func(s *Settings) *string { return &s.WindowLayout }, "sidebar-left",
		en("sidebar-left", "Vertical Tabs"), en("tabs-top", "Tabs Top"), en("tabs-bottom", "Tabs Bottom")),
	dBool("sidebar-visible", "Show the tabs rail", "Whether the tabs rail is open.", func(s *Settings) *bool { return &s.SidebarVisible }, true),
	dNum("sidebar-width", "Rail width", "The width of the tabs rail.", func(s *Settings) *float64 { return &s.SidebarWidth }, 220, 160, 420, 4),
	dNum("details-panel-width", "Details panel width", "The width of the details panel.", func(s *Settings) *float64 { return &s.DetailsPanelWidth }, 220, 160, 420, 4),
	dEnum("window-size", "Window Size", "How new windows decide their initial dimensions.", func(s *Settings) *string { return &s.WindowSize }, "remember",
		en("remember", "Remember last size"), en("frame", "Fixed by frame"), en("grid", "Fixed by grid")),
	dNum("window-width-px", "Width", "The window's width.", func(s *Settings) *float64 { return &s.WindowWidth }, 1120, 680, 3200, 20),
	dNum("window-height-px", "Height", "The window's height.", func(s *Settings) *float64 { return &s.WindowHeight }, 740, 440, 2400, 20),
	dNum("window-cols", "Columns", "The window's width in columns.", func(s *Settings) *float64 { return &s.WindowCols }, 80, 20, 400, 1),
	dNum("window-rows", "Rows", "The window's height in rows.", func(s *Settings) *float64 { return &s.WindowRows }, 24, 10, 200, 1),
	dNum("unfocused-split-opacity", "Unfocused Split Opacity", "How much split panes fade when they don't have focus.", func(s *Settings) *float64 { return &s.UnfocusedSplitOpacity }, 0.6, 0, 1, 0.05),
	dNum("background-opacity", "Terminal background opacity", "The terminal background's opacity.", func(s *Settings) *float64 { return &s.TerminalOpacity }, 1, 0.5, 1, 0.05),
	dNum("window-opacity", "Window opacity", "The whole window's opacity.", func(s *Settings) *float64 { return &s.WindowOpacity }, 1, 0.5, 1, 0.05),
	dNum("container-padding", "Padding", "The padding around the terminal.", func(s *Settings) *float64 { return &s.ContainerPadding }, 0, 0, 24, 1),

	// --- Controls: keyboard & mouse ---
	dEnum("macos-option-as-alt", "Use Option Key as Meta", "Treat Option as Alt/Meta so apps see Esc-prefixed sequences.", func(s *Settings) *string { return &s.OptionAsAlt }, "false",
		en("false", "Off (typing accents)"), en("true", "Both Option keys"), en("left", "Left Option only"), en("right", "Right Option only")),
	dBool("focus-follows-mouse", "Mouse Over to Focus", "Focus the pane under the mouse cursor automatically.", func(s *Settings) *bool { return &s.FocusFollowsMouse }, false),
	dEnum("right-click-action", "Right-Click Action", "What right-click does in the terminal viewport.", func(s *Settings) *string { return &s.RightClickAction }, "context-menu",
		en("context-menu", "Context Menu"), en("copy", "Copy"), en("paste", "Paste"), en("copy-or-paste", "Copy or Paste"), en("ignore", "Ignore")),
	dBool("mouse-hide-while-typing", "Hide Mouse When Typing", "Hide the mouse cursor while the keyboard is in use.", func(s *Settings) *bool { return &s.MouseHideWhileTyping }, false),
	dBool("mouse-shift-to-select", "Allow Shift with Mouse Click", "Hold Shift to select text even when the running app captures the mouse.", func(s *Settings) *bool { return &s.MouseShiftToSelect }, true),
	dEnum("bypass-mouse-reporting", "Bypass Mouse Reporting", "Hold this modifier to select text or open a link past a mouse-capturing app.", func(s *Settings) *string { return &s.BypassMouseReporting }, "shift",
		en("none", "None"), en("shift", "Shift"), en("ctrl", "Ctrl"), en("alt", "Alt"), en("ctrl-shift", "Ctrl+Shift"), en("super", "Super")),
	dBool("link-click-over-mouse-mode", "Ctrl-Click Opens Links in Mouse Mode", "Ctrl-click opens a link even while an app captures the mouse.", func(s *Settings) *bool { return &s.LinkClickOverMouseMode }, true),
	dBool("cursor-click-to-move", "Cursor Click-to-Move", "Click in the prompt to move the shell cursor.", func(s *Settings) *bool { return &s.CursorClickToMove }, true),

	// --- Controls: open with & links ---
	dEnum("link-open-with", "Open Links With", "Where Cmd+click opens a URL.", func(s *Settings) *string { return &s.LinkOpenWith }, "browser",
		en("browser", "Browser"), en("otty", "Otty")),
	dEnum("file-open-with", "Open Files With", "Where Cmd+click opens a file path.", func(s *Settings) *string { return &s.FileOpenWith }, "default-app",
		en("default-app", "Default App"), en("otty", "Otty")),
	dEnum("folder-open-with", "Open Folders With", "Where Cmd+click opens a folder path.", func(s *Settings) *string { return &s.FolderOpenWith }, "default-app",
		en("default-app", "Finder"), en("otty", "Otty")),
	dText("default-git-client", "Default Git Client", "The git GUI opened from the Details Git toolbar.", func(s *Settings) *string { return &s.DefaultGitClient }, ""),
	dEnum("link-schemes", "Auto-Detect Link Schemes", "Which URL schemes become clickable on Cmd-hover.", func(s *Settings) *string { return &s.LinkSchemes }, "all",
		en("all", "All"), en("custom", "Custom")),
	dBool("link-previews", "Show Link Previews", "Show the full path or URL while Cmd-hovering a link.", func(s *Settings) *bool { return &s.LinkPreviews }, true),

	// --- Controls: secure input & scroll & clipboard ---
	dBool("auto-secure-input", "Auto secure input", "Turn on Secure Keyboard Entry at a password prompt.", func(s *Settings) *bool { return &s.AutoSecureInput }, true),
	dBool("secure-input-indication", "Indicate secure input", "Show a pill in the title bar while Secure Keyboard Entry is active.", func(s *Settings) *bool { return &s.SecureInputIndication }, true),
	dEnum("terminal-scroll-past-end", "Scroll Past Last Line", "Where the last row settles when you scroll past the bottom.", func(s *Settings) *string { return &s.ScrollPastEnd }, "disabled",
		en("disabled", "Disabled"), en("last-line-with-content", "Last Line With Content"), en("last-line-in-middle", "Last Line In Middle"), en("cursor-line", "Cursor Line")),
	dEnum("terminal-scroll-past-first-line", "Scroll Past First Line", "Where the first row settles when you scroll above the top.", func(s *Settings) *string { return &s.ScrollPastFirstLine }, "disabled",
		en("disabled", "Disabled"), en("same-as-last-line", "Same as Scroll Past Last Line"), en("first-line-with-content", "First Line With Content"), en("first-line-in-middle", "First Line In Middle")),
	dBool("terminal-scroll-smooth", "Smooth Scroll", "Scroll by pixel instead of snapping row by row.", func(s *Settings) *bool { return &s.ScrollSmooth }, true),
	dBool("copy-on-select", "Copy on Select", "Copy text to the clipboard automatically when you finish selecting.", func(s *Settings) *bool { return &s.CopyOnSelect }, false),
	dBool("clipboard-trim-trailing-spaces", "Trim Trailing Space on Copy", "Strip trailing whitespace from each line when copying.", func(s *Settings) *bool { return &s.ClipboardTrimTrailingSpaces }, false),
	dBool("clipboard-paste-protection", "Confirm When Pasting Unsafely", "Warn when pasted text would auto-execute commands.", func(s *Settings) *bool { return &s.ClipboardPasteProtection }, true),
	dBool("clipboard-paste-bracketed-safe", "Mark Bracketed Paste as Safe", "Skip the paste warning when the running app supports bracketed paste mode.", func(s *Settings) *bool { return &s.ClipboardPasteBracketedSafe }, true),
	dBool("selection-clear-on-typing", "Clear Selection on Typing", "Clear the active selection when you type.", func(s *Settings) *bool { return &s.SelectionClearOnTyping }, true),
	dBool("selection-clear-on-copy", "Clear Selection on Copy", "Clear the active selection after Cmd+C.", func(s *Settings) *bool { return &s.SelectionClearOnCopy }, true),
	dBool("selection-backspace-deletes", "Backspace Deletes Selection", "Backspace deletes the selection, not one character.", func(s *Settings) *bool { return &s.SelectionBackspaceDeletes }, true),
	dEnum("command-history-scope", "History Scope", "Whether panes share command history as you type.", func(s *Settings) *string { return &s.CommandHistoryScope }, "per-pane",
		en("follow-shell", "Follow Shell Config"), en("shared", "Shared Across Panes"), en("per-pane", "Per Pane")),
	dEnum("history-up-arrow", "Up Arrow Behavior", "What Up recalls at the prompt.", func(s *Settings) *string { return &s.HistoryUpArrow }, "auto",
		en("auto", "Follow Shell Config"), en("sequential", "Previous Command"), en("prefix", "Match What I Typed")),
	dNum("scrollback-lines", "Scrollback Lines", "How many lines of scrollback to keep.", func(s *Settings) *float64 { return &s.ScrollbackLines }, 10000, 0, 1000000, 100),

	// --- Shell ---
	dText("command", "Command", "The shell new sessions run. Empty uses your login shell.", func(s *Settings) *string { return &s.Command }, ""),
	dText("working-directory", "Working Directory", "Where new sessions start. Empty uses your home directory.", func(s *Settings) *string { return &s.WorkingDirectory }, ""),
	dText("term", "Term", "What kind of terminal to advertise to child processes.", func(s *Settings) *string { return &s.Term }, "auto"),
	dBool("shell-integration", "Provide Shell Integration", "Powers prompt marks, CWD tracking and command status.", func(s *Settings) *bool { return &s.ShellIntegration }, true),
	dBool("ssh-integration", "SSH Integration", "Forward environment and terminfo over SSH.", func(s *Settings) *bool { return &s.SSHIntegration }, true),
	dEnum("session-restore-processes", "Re-run Processes on Restore", "Re-launch the command that was running in each pane when a window is restored.", func(s *Settings) *string { return &s.SessionRestoreProcesses }, "none",
		en("none", "None"), en("whitelist", "Whitelisted only"), en("all", "All running processes")),
	dBool("session-restore-banner", "Restore banner", "Show a banner when a session was restored.", func(s *Settings) *bool { return &s.SessionRestoreBanner }, true),
	dBool("privilege-sound-shell", "Allow Terminal Bell", "Allow shell apps to play sounds via the BEL character.", func(s *Settings) *bool { return &s.SoundShell }, true),
	dBool("privilege-sound-on-error", "Beep on Error Exit", "Beep when a command exits with non-zero status.", func(s *Settings) *bool { return &s.SoundOnError }, false),
	dBool("privilege-notification-shell", "Allow App Notifications", "Allow shell apps to send system notifications.", func(s *Settings) *bool { return &s.NotifyShellApp }, true),
	dBool("privilege-notification-on-finish", "Notify on Command Finish", "Notify when a background command finishes.", func(s *Settings) *bool { return &s.NotifyOnFinish }, false),
	dBool("privilege-notification-on-error", "Notify on Error Exit", "Notify when a command fails.", func(s *Settings) *bool { return &s.NotifyOnError }, true),
	dEnum("notification-foreground", "Notify While Foreground", "Banner behavior while Otty is the foreground app.", func(s *Settings) *string { return &s.NotifyForeground }, "off",
		en("off", "Off (system default)"), en("tab-unfocused", "Only when source tab is unfocused"), en("always", "Always")),

	// --- General ---
	dEnum("language", "Language", "The language of the app's interface.", func(s *Settings) *string { return &s.Language }, "chinese",
		en("system", "System"), en("english", "English"), en("chinese", "中文")),
	dEnum("on-launch", "On Launch", "What Otty does when it starts.", func(s *Settings) *string { return &s.OnLaunch }, "restore_session",
		en("new_window", "New Window"), en("restore_session", "Restore Last Session")),
	dEnum("confirm-close-tab", "Closing Tab", "When closing a tab asks first.", func(s *Settings) *string { return &s.ConfirmCloseTab }, "process",
		en("always", "Always"), en("process", "Running Process"), en("never", "Never")),
	dEnum("confirm-close-window", "Closing Window", "When closing a window asks first.", func(s *Settings) *string { return &s.ConfirmCloseWindow }, "process",
		en("always", "Always"), en("process", "Running Process"), en("multiple_tabs", "Multiple Tabs"), en("never", "Never")),
	dEnum("confirm-close-pane", "Closing Pane", "When closing a pane asks first.", func(s *Settings) *string { return &s.ConfirmClosePane }, "process",
		en("always", "Always"), en("process", "Running Process"), en("never", "Never")),
	dEnum("new-tab-position", "New Tab Position", "Where a new tab opens.", func(s *Settings) *string { return &s.NewTabPosition }, "auto",
		en("end", "End"), en("auto", "Auto"), en("after-current", "After Current Tab")),
	dBool("quit-after-last-window-closed", "Quit When All Windows Closed", "Quit with the last window.", func(s *Settings) *bool { return &s.QuitAfterLastWindowClosed }, false),
	dBool("new-window-when-all-windows-closed", "New Window When All Windows Closed", "Open a new window when the last one closes.", func(s *Settings) *bool { return &s.NewWindowWhenAllWindowsClosed }, false),
	dBool("window-hide-dirty-indicator", "Hide Close Button Dot", "Never show the unsaved/running dot on the window close button.", func(s *Settings) *bool { return &s.WindowHideDirtyIndicator }, false),

	// --- Editor ---
	dBool("editor-line-wrap", "Soft Wrap", "Wrap long lines instead of scrolling horizontally.", func(s *Settings) *bool { return &s.EditorLineWrap }, true),
	dBool("editor-show-line-numbers", "Show Line Numbers", "Display the line-number gutter in text panes.", func(s *Settings) *bool { return &s.EditorShowLineNumbers }, true),
	dBool("editor-visible-whitespace", "Show Whitespace", "Render spaces, tabs, and newlines as visible glyphs.", func(s *Settings) *bool { return &s.EditorVisibleWhitespace }, false),
	dNum("editor-tab-size", "Tab Size", "Visual width of a tab character, in columns.", func(s *Settings) *float64 { return &s.EditorTabSize }, 4, 1, 16, 1),
	dBool("editor-scroll-past-end", "Scroll Past Last Line", "Allow scrolling past the last line in the editor.", func(s *Settings) *bool { return &s.EditorScrollPastEnd }, true),
	dBool("editor-vim", "Vim Keys", "Modal editing in file and editor panes.", func(s *Settings) *bool { return &s.EditorVim }, false),
	dBool("editor-default-to-preview-readonly", "Open in Preview / Read-Only", "Markdown / SVG / HTML open as a rendered preview.", func(s *Settings) *bool { return &s.EditorDefaultToPreviewReadonly }, true),

	// --- View ---
	dBool("title-from-osc", "Programs Can Set the Title", "Let programs rename the tab and window with OSC 0/1/2.", func(s *Settings) *bool { return &s.TitleFromOsc }, true),
	dEnum("tab-icon-placement", "Icon and Badge", "Whether the tab icon and status badge share one slot.", func(s *Settings) *string { return &s.TabIconPlacement }, "combined",
		en("combined", "Combined"), en("separate", "Separate")),
	dBool("privilege-badge-exit-status", "When Command Finishes", "Show accent dot on tab when a command finishes.", func(s *Settings) *bool { return &s.BadgeExitStatus }, true),
	dBool("privilege-badge-activity", "When Command Fails", "Show error alert on tab when a command fails.", func(s *Settings) *bool { return &s.BadgeActivity }, true),
	dBool("detect-awaiting-input", "When Command Awaits Input", "Detect [y/n], password and ENTER prompts and show the awaiting-input badge.", func(s *Settings) *bool { return &s.BadgeAwaitingInput }, true),
	dBool("web-pane-persist-cookies", "Stay Signed In", "Keep cookies and site data on disk, so sign-ins survive a restart.", func(s *Settings) *bool { return &s.WebPanePersistCookies }, true),

	// --- Recipes ---
	dEnum("recipe-replay-saved", "Saved Recipes", "How recipes you saved replay their commands when opened.", func(s *Settings) *string { return &s.RecipeReplaySaved }, "ask_once",
		en("auto", "Auto"), en("ask_once", "Ask Once"), en("manually", "Manually")),
	dEnum("recipe-replay-file", "Recipe Files", "How an external .ottyrecipe file replays its commands.", func(s *Settings) *string { return &s.RecipeReplayFile }, "manually",
		en("auto", "Auto"), en("ask_once", "Ask Once"), en("manually", "Manually")),

	// --- Agents ---
	dBool("privilege-badge-agent-processing", "Badge While Processing", "Show a tab badge while an agent is running a task.", func(s *Settings) *bool { return &s.AgentBadgeProcessing }, true),
	dBool("privilege-badge-agent-task-complete", "Badge When Task Completes", "Show a tab badge when an agent finishes a task.", func(s *Settings) *bool { return &s.AgentBadgeTaskComplete }, true),
	dBool("privilege-badge-agent-awaiting-input", "Badge When Awaiting Input", "Show a tab badge when an agent is waiting for you.", func(s *Settings) *bool { return &s.AgentBadgeAwaitingInput }, true),
	dBool("privilege-notify-agent-task-complete", "Notify When Task Completes", "Send a system notification when an agent finishes a task.", func(s *Settings) *bool { return &s.AgentNotifyTaskComplete }, true),
	dBool("privilege-notify-agent-awaiting-input", "Notify When Awaiting Input", "Send a system notification when an agent is waiting for you.", func(s *Settings) *bool { return &s.AgentNotifyAwaitingInput }, true),
	dBool("privilege-caffeinate-agent-processing", "Prevent Sleep While Processing", "Keep the Mac awake while an agent is running a task.", func(s *Settings) *bool { return &s.AgentPreventSleep }, false),
	dBool("privilege-resume-agent-session", "Resume Session on Recovery", "Reattach the agent's session when a window is restored.", func(s *Settings) *bool { return &s.AgentResumeSession }, true),
}

// settingsGroups is every Settings page's sections, in order. A key that
// appears in no page still shows in the Advanced tab's All Settings list.
var settingsGroups = map[string][]settingGroup{
	"appearance": {
		{"Theme", []string{"theme", "theme-dark", "auto-theme-dark-mode"}},
		{"Font", []string{"font-family", "font-family-fallback", "font-size", "font-blending", "text-bold", "text-italic", "text-underline", "text-blink", "font-ligatures", "font-thicken", "line-height", "adjust-cell-height"}},
		{"Cursor", []string{"cursor-style", "cursor-style-blink", "cursor-animation", "cursor-opacity", "cursor-color", "cursor-text"}},
		{"Colors", []string{"foreground", "background", "selection-background", "selection-foreground", "bold-color", "faint-opacity", "minimum-contrast"}},
		{"Layout", []string{"window-layout"}},
		{"Window", []string{"window-size", "window-width-px", "window-height-px", "window-cols", "window-rows", "unfocused-split-opacity", "window-opacity"}},
		{"Terminal", []string{"background-opacity", "container-padding"}},
	},
	"general": {
		{"Closing", []string{"confirm-close-tab", "confirm-close-window", "confirm-close-pane"}},
		{"Tabs", []string{"new-tab-position"}},
		{"Startup", []string{"on-launch", "quit-after-last-window-closed", "new-window-when-all-windows-closed"}},
		{"Window", []string{"window-hide-dirty-indicator"}},
		{"Language", []string{"language"}},
	},
	"shell": {
		{"Session", []string{"command", "working-directory", "term", "env"}},
		{"Integration", []string{"shell-integration", "ssh-integration"}},
		{"Session Restore", []string{"session-restore-processes", "session-restore-banner", "command-whitelist"}},
		{"Sound", []string{"privilege-sound-shell", "privilege-sound-on-error"}},
		{"Notifications", []string{"privilege-notification-shell", "privilege-notification-on-finish", "privilege-notification-on-error", "notification-foreground"}},
	},
	"controls": {
		{"Keyboard", []string{"macos-option-as-alt"}},
		{"Mouse", []string{"focus-follows-mouse", "right-click-action", "mouse-hide-while-typing", "mouse-shift-to-select", "bypass-mouse-reporting", "link-click-over-mouse-mode", "cursor-click-to-move"}},
		{"Open With", []string{"link-open-with", "file-open-with", "folder-open-with", "default-git-client"}},
		{"Link Schemes", []string{"link-schemes", "link-scheme-allowlist", "link-previews"}},
		{"Secure Input", []string{"auto-secure-input", "secure-input-indication"}},
		{"Scroll", []string{"terminal-scroll-past-end", "terminal-scroll-past-first-line", "terminal-scroll-smooth", "scrollback-lines"}},
		{"Clipboard", []string{"copy-on-select", "clipboard-trim-trailing-spaces", "clipboard-paste-protection", "clipboard-paste-bracketed-safe", "selection-clear-on-typing", "selection-clear-on-copy", "selection-backspace-deletes"}},
		{"Command History", []string{"command-history-scope", "history-up-arrow"}},
	},
	"editor": {
		{"Text", []string{"editor-line-wrap", "editor-show-line-numbers", "editor-visible-whitespace", "editor-tab-size"}},
		{"Scrolling", []string{"editor-scroll-past-end"}},
		{"Editing", []string{"editor-vim"}},
		{"Opening Files", []string{"editor-default-to-preview-readonly"}},
	},
	"view": {
		{"Tab & Title", []string{"title-from-osc", "tab-icon-placement"}},
		{"Tab Icon & Badge", []string{"privilege-badge-exit-status", "privilege-badge-activity", "detect-awaiting-input"}},
		{"Web Pane", []string{"web-pane-persist-cookies"}},
	},
	"recipes": {
		{"Command Replay", []string{"recipe-replay-saved", "recipe-replay-file"}},
	},
	"agents": {
		{"Agent Behavior", []string{"privilege-badge-agent-processing", "privilege-badge-agent-task-complete", "privilege-badge-agent-awaiting-input", "privilege-notify-agent-task-complete", "privilege-notify-agent-awaiting-input", "privilege-caffeinate-agent-processing", "privilege-resume-agent-session"}},
	},
}

// defByKey is the registry entry of a key, or nil.
func defByKey(key string) *settingDef {
	for i := range settingsDefs {
		if settingsDefs[i].key == key {
			return &settingsDefs[i]
		}
	}
	return nil
}

// init appends the Settings the pages build from a variable list: the two
// theme selects offer otty's themes with Follow system first (the empty
// name, which the terminals follow the window for), and the sixteen
// palette colors each get a well, as the Appearance page's palette grid.
func init() {
	themes := []enumOpt{{value: "", label: "Follow system"}}
	for _, n := range OttyThemeNames {
		themes = append(themes, en(n, n))
	}
	settingsDefs = append(settingsDefs,
		dEnum("theme", "Theme", "The color scheme for light windows.", func(s *Settings) *string { return &s.ThemeName }, "Paper", themes...),
		dEnum("theme-dark", "Dark Theme", "The color scheme for dark windows.", func(s *Settings) *string { return &s.ThemeDark }, "Nord", themes...),
	)
	for i := 0; i < 16; i++ {
		i := i
		key := fmt.Sprintf("palette-%d", i)
		settingsDefs = append(settingsDefs, settingDef{
			key: key, label: key, desc: "Palette color " + fmt.Sprint(i) + ".", kind: "color",
			ptr: func(s *Settings) any { return &s.Palette[i] },
			get: func(s *Settings) any { return s.Palette[i] },
			set: func(s *Settings, v any) { s.Palette[i] = v.(string) },
		})
	}
}
