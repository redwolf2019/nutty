package main

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
	"github.com/smallnest/uix/components/searchfield"
)

// keybindPage is the Key Bindings page's state: the filter and the action
// being recorded, if any.
type keybindPage struct {
	search    string
	recording string // the action whose shortcut is being recorded
}

// keybindRow is one action of the key binding list: its label, its
// effective trigger (the recorded one, else otty's default), and whether
// the trigger is a recording override.
type keybindRow struct {
	action     string
	label      string
	trigger    string
	overridden bool
}

// pageKeyBindings is the Key Bindings category: a search box over otty's
// actions, grouped as otty groups them, each row's trigger a chip that
// records a new one. The bindings save into KeybindOverrides and rebuild
// the menu's accelerators.
func (sw *settingsWindow) pageKeyBindings(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		ui.Box(c).Grow(1).MinWidth(0).Children(func() {
			searchfield.SearchField(c, searchfield.Props{Value: &sw.keys.search})
		})
		if len(sw.s.KeybindOverrides) > 0 {
			rb := ui.ButtonBase(c.Key("kb-reset-all")).Padding(4, 10).Radius(t.Radius).
				FocusRing(false).Tooltip("Restore every otty default")
			rb.Children(func() {
				ui.Row(c).Gap(4).Children(func() {
					ui.Image(c, lucide.RotateCcw()).Size(12, 12)
					ui.Text(c, "Reset to Default").FontSize(12).TextColor(t.TextMuted)
				})
			})
			if rb.Clicked() {
				sw.s.KeybindOverrides = map[string]string{}
				sw.rebind()
			}
		}
	})
	if sw.keys.recording != "" {
		ui.Row(c).Gap(6).Children(func() {
			ui.Image(c, lucide.Keyboard()).Size(13, 13)
			ui.Text(c, "Press a shortcut…  (Esc to cancel)").FontSize(11).TextColor(t.Accent)
		})
	}
	sw.keybindList(c)
}

// keybindList draws the actions in their otty groups, each with a chip
// that records its shortcut, and records one while that is active.
func (sw *settingsWindow) keybindList(c *ui.Context) {
	rows := sw.keybindRows()
	if len(rows) == 0 {
		sw.hint(c, "No matching key bindings")
		return
	}
	group := ""
	for _, r := range rows {
		g := keybindGroupOf(r.action)
		if g != group {
			group = g
			t := c.Theme()
			ui.Box(c).FillWidth().Height(10)
			ui.Text(c, group).FontSize(12).Bold().TextColor(t.TextMuted)
		}
		sw.keybindRow(c, r)
	}
	// The recorder watches the keys while an action is being bound.
	if sw.keys.recording != "" {
		sw.captureCombo(c)
	}
}

// keybindRows is the visible actions in group order: the ones otty binds
// by default first, then the rest, filtered by the search.
func (sw *settingsWindow) keybindRows() []keybindRow {
	rows := []keybindRow{}
	seen := map[string]bool{}
	add := func(action string) {
		if seen[action] {
			return
		}
		seen[action] = true
		label := keybindActions[strings.ReplaceAll(action, ":", "_")]
		if label == "" {
			label = action
		}
		trigger, overridden := sw.s.actionTrigger(action)
		rows = append(rows, keybindRow{action: action, label: label, trigger: trigger, overridden: overridden})
	}
	for _, kb := range defaultKeybinds {
		add(kb.action)
	}
	for action := range keybindActions {
		add(action)
	}
	if q := strings.ToLower(sw.keys.search); q != "" {
		filtered := rows[:0]
		for _, r := range rows {
			if strings.Contains(strings.ToLower(r.label), q) || strings.Contains(strings.ToLower(r.trigger), q) {
				filtered = append(filtered, r)
			}
		}
		rows = filtered
	}
	return rows
}

// keybindRow is one action's row: its name, the trigger chip that records
// a new shortcut, and a clear button when the trigger is an override.
func (sw *settingsWindow) keybindRow(c *ui.Context, r keybindRow) {
	t := c.Theme()
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		ui.Text(c, r.label).FontSize(13).SingleLine().Grow(1).MinWidth(0).TextColor(t.Text)
		if r.overridden && r.trigger != "" {
			cb := ui.ButtonBase(c.Key("kb-clear-" + r.action)).Padding(4, 5).Radius(t.Radius).
				FocusRing(false).Tooltip("Use the default shortcut")
			cb.Children(func() { ui.Image(c, lucide.X()).Size(12, 12) })
			if cb.Clicked() {
				delete(sw.s.KeybindOverrides, r.trigger)
				sw.rebind()
			}
		}
		chip := ui.ButtonBase(c.Key("kb-" + r.action)).Padding(4, 10).Radius(999).
			FocusRing(false).Tooltip("Click to rebind")
		if sw.keys.recording == r.action {
			chip.Background(t.Accent)
			chip.Children(func() {
				ui.Text(c, "Press a shortcut…").FontSize(12).TextColor(t.AccentText)
			})
		} else if r.trigger == "" {
			chip.Border(1, t.Border)
			chip.Children(func() {
				ui.Text(c, "No shortcut").FontSize(12).TextColor(t.TextMuted)
			})
		} else {
			chip.Border(1, t.Border)
			chip.Children(func() {
				ui.Text(c, r.trigger).FontSize(12).TextColor(t.Text)
			})
		}
		if chip.Clicked() {
			if sw.keys.recording == r.action {
				sw.keys.recording = ""
			} else {
				sw.keys.recording = r.action
			}
		}
		if other, ok := sw.s.triggerAction(r.trigger); sw.keys.recording != r.action && ok && other != r.action {
			ui.Text(c, "Conflicts with: "+keybindActions[strings.ReplaceAll(other, ":", "_")]).
				FontSize(10).TextColor(t.Danger)
		}
	})
}

// captureCombo watches the keys while a shortcut is being recorded, and
// binds the first combination pressed. Escape cancels.
func (sw *settingsWindow) captureCombo(c *ui.Context) {
	if c.Shortcut(0, ui.KeyEscape) {
		sw.keys.recording = ""
		return
	}
	for _, m := range []ui.Modifiers{0, ui.Shift, ui.Ctrl, ui.Alt, ui.Super,
		ui.Ctrl | ui.Shift, ui.Alt | ui.Shift, ui.Super | ui.Shift,
		ui.Super | ui.Ctrl, ui.Super | ui.Alt, ui.Alt | ui.Ctrl,
		ui.Super | ui.Ctrl | ui.Shift, ui.Super | ui.Alt | ui.Shift,
		ui.Ctrl | ui.Alt | ui.Shift, ui.Super | ui.Ctrl | ui.Alt,
		ui.Super | ui.Ctrl | ui.Alt | ui.Shift} {
		for _, key := range recordKeys {
			if c.Shortcut(m, key) {
				sw.recordCombo(modName(m) + keyName(key))
				return
			}
		}
	}
}

// recordKeys are the keys the recorder accepts, otty's key names.
var recordKeys = []ui.Key{
	ui.KeyA, ui.KeyB, ui.KeyC, ui.KeyD, ui.KeyE, ui.KeyF, ui.KeyG, ui.KeyH,
	ui.KeyI, ui.KeyJ, ui.KeyK, ui.KeyL, ui.KeyM, ui.KeyN, ui.KeyO, ui.KeyP,
	ui.KeyQ, ui.KeyR, ui.KeyS, ui.KeyT, ui.KeyU, ui.KeyV, ui.KeyW, ui.KeyX,
	ui.KeyY, ui.KeyZ,
	ui.Key0, ui.Key1, ui.Key2, ui.Key3, ui.Key4, ui.Key5, ui.Key6, ui.Key7,
	ui.Key8, ui.Key9,
	ui.KeyEnter, ui.KeyEscape, ui.KeyBackspace, ui.KeyTab, ui.KeySpace,
	ui.KeyLeft, ui.KeyRight, ui.KeyUp, ui.KeyDown, ui.KeyHome, ui.KeyEnd,
	ui.KeyPageUp, ui.KeyPageDown, ui.KeyInsert, ui.KeyDelete,
	ui.KeyMinus, ui.KeyEqual, ui.KeyComma, ui.KeyPeriod, ui.KeySlash,
	ui.KeySemicolon, ui.KeyQuote, ui.KeyBracketLeft, ui.KeyBracketRight,
	ui.KeyBackslash, ui.KeyBackquote,
	ui.KeyF1, ui.KeyF2, ui.KeyF3, ui.KeyF4, ui.KeyF5, ui.KeyF6, ui.KeyF7,
	ui.KeyF8, ui.KeyF9, ui.KeyF10, ui.KeyF11, ui.KeyF12,
}

// modName is an otty trigger's modifier list, cmd before ctrl before opt
// before shift, as otty writes them.
func modName(m ui.Modifiers) string {
	var parts []string
	if m&ui.Super != 0 {
		parts = append(parts, "cmd")
	}
	if m&ui.Ctrl != 0 {
		parts = append(parts, "ctrl")
	}
	if m&ui.Alt != 0 {
		parts = append(parts, "opt")
	}
	if m&ui.Shift != 0 {
		parts = append(parts, "shift")
	}
	return strings.Join(parts, "+")
}

// keyName is a key's otty trigger name.
func keyName(k ui.Key) string {
	switch {
	case k >= ui.KeyA && k <= ui.KeyZ:
		return string(rune('a' + k - ui.KeyA))
	case k >= ui.Key0 && k <= ui.Key9:
		return string(rune('0' + k - ui.Key0))
	case k >= ui.KeyF1 && k <= ui.KeyF12:
		return fmt.Sprintf("f%d", k-ui.KeyF1+1)
	}
	switch k {
	case ui.KeyEnter:
		return "enter"
	case ui.KeyEscape:
		return "escape"
	case ui.KeyBackspace:
		return "backspace"
	case ui.KeyDelete:
		return "forward_delete"
	case ui.KeyTab:
		return "tab"
	case ui.KeySpace:
		return "space"
	case ui.KeyLeft:
		return "left"
	case ui.KeyRight:
		return "right"
	case ui.KeyUp:
		return "up"
	case ui.KeyDown:
		return "down"
	case ui.KeyHome:
		return "home"
	case ui.KeyEnd:
		return "end"
	case ui.KeyPageUp:
		return "page_up"
	case ui.KeyPageDown:
		return "page_down"
	case ui.KeyInsert:
		return "insert"
	case ui.KeyMinus:
		return "minus"
	case ui.KeyEqual:
		return "equal"
	case ui.KeyComma:
		return "comma"
	case ui.KeyPeriod:
		return "period"
	case ui.KeySlash:
		return "slash"
	case ui.KeySemicolon:
		return "semicolon"
	case ui.KeyQuote:
		return "quote"
	case ui.KeyBracketLeft:
		return "left_bracket"
	case ui.KeyBracketRight:
		return "right_bracket"
	case ui.KeyBackslash:
		return "backslash"
	case ui.KeyBackquote:
		return "backtick"
	}
	return ""
}

// recordCombo binds a recorded combination to the action being recorded,
// dropping the action's old binding and any other action on the combo,
// then rebuilds the menu.
func (sw *settingsWindow) recordCombo(combo string) {
	if combo == "" {
		return
	}
	action := sw.keys.recording
	for t, a := range sw.s.KeybindOverrides {
		if a == action || t == combo {
			delete(sw.s.KeybindOverrides, t)
		}
	}
	sw.s.KeybindOverrides[combo] = action
	sw.keys.recording = ""
	sw.rebind()
}

// rebind applies the key bindings to the menu after a change, when there
// is a window; the tests build the page without one.
func (sw *settingsWindow) rebind() {
	if sw.a != nil {
		sw.a.applyKeybinds()
	}
}

// applyKeybinds rebuilds the application menu, whose accelerators come
// from the key bindings.
func (a *app) applyKeybinds() {
	setAppMenu(a.settings)
}

// actionTrigger is the trigger bound to an action: the recorded override,
// or otty's default. The second result reports whether it is an override.
func (s *settings) actionTrigger(action string) (string, bool) {
	for t, a := range s.KeybindOverrides {
		if a == action {
			return t, true
		}
	}
	for _, kb := range defaultKeybinds {
		if kb.action == action {
			return kb.trigger, false
		}
	}
	return "", false
}

// triggerAction is the action bound to a trigger: the recorded override,
// or otty's default. The second result reports whether it is an override.
func (s *settings) triggerAction(trigger string) (string, bool) {
	if a, ok := s.KeybindOverrides[trigger]; ok {
		return a, true
	}
	for _, kb := range defaultKeybinds {
		if kb.trigger == trigger {
			return kb.action, false
		}
	}
	return "", false
}

// keybindGroupOf is the otty group an action sits under on the page.
func keybindGroupOf(action string) string {
	prefix := func(p string) bool {
		return strings.HasPrefix(action, p)
	}
	switch {
	case action == "new_tab" || action == "duplicate_tab" || action == "close_tab" ||
		action == "reopen_closed_tab" || action == "next_tab" || action == "previous_tab" ||
		prefix("goto_tab") || prefix("switch"):
		return "Tabs"
	case prefix("split") || prefix("focus_pane") || prefix("move_divider") ||
		prefix("equalize") || prefix("toggle_split") || action == "close_pane":
		return "Pane"
	case prefix("cursor") || prefix("delete") || prefix("selection") || prefix("word"):
		return "Text Editing"
	case prefix("scroll") || prefix("page") || action == "scroll_to_top" || action == "scroll_to_bottom":
		return "Scroll"
	case action == "copy" || action == "paste" || action == "select_all" || prefix("find"):
		return "Edit"
	case prefix("font") || prefix("theme") || action == "command_palette" ||
		action == "toggle_tabs_panel" || action == "toggle_details_panel" ||
		action == "toggle_info_panel" || action == "toggle_full_screen" ||
		prefix("toggle_agent") || action == "open_settings" || action == "open_config_file" ||
		action == "reload_config" || prefix("appearance") || prefix("details_"):
		return "View"
	case action == "new_window" || action == "minimize" || prefix("toggle_window"):
		return "Window"
	case prefix("recipe") || prefix("command_") || prefix("run_"):
		return "Commands / Recipe"
	case prefix("send_"):
		return "Text / Sequence"
	case prefix("agent"):
		return "Code Agents"
	case action == "global_hotkey_open_window" || action == "global_hotkey_toggle_visibility":
		return "Global Hotkey"
	}
	return "General"
}

// acceleratorOf is a mygo menu accelerator for an otty trigger, or "".
func acceleratorOf(trigger string) string {
	var mods []string
	var key string
	parts := strings.Split(trigger, "+")
	for i, p := range parts {
		if i < len(parts)-1 {
			switch p {
			case "cmd":
				mods = append(mods, "CmdOrCtrl")
			case "ctrl":
				mods = append(mods, "Ctrl")
			case "opt", "alt":
				mods = append(mods, "Option")
			case "shift":
				mods = append(mods, "Shift")
			default:
				return "" // an unknown modifier means no accelerator
			}
			continue
		}
		key = p
	}
	names := map[string]string{
		"left_bracket":  "[",
		"right_bracket": "]",
		"comma":         ",",
		"period":        ".",
		"equal":         "=",
		"minus":         "-",
		"semicolon":     ";",
		"quote":         "'",
		"backslash":     "\\",
		"slash":         "/",
		"backtick":      "`",
		"enter":         "Enter",
		"space":         "Space",
		"tab":           "Tab",
		"backspace":     "Backspace",
		"forward_delete": "Delete",
		"delete":        "Delete",
		"escape":        "Escape",
		"home":          "Home",
		"end":           "End",
		"up":            "Up",
		"down":          "Down",
		"left":          "Left",
		"right":         "Right",
		"page_up":       "PageUp",
		"page_down":     "PageDown",
		"insert":        "Insert",
	}
	if v, ok := names[key]; ok {
		key = v
	} else if len(key) == 1 {
		key = strings.ToUpper(key) // letters, digits, and punctuation mygo uppercases
	} else if len(key) >= 2 && key[0] == 'f' && key[1] >= '1' && key[1] <= '9' {
		key = strings.ToUpper(key)
	}
	mods = append(mods, key)
	return strings.Join(mods, "+")
}

