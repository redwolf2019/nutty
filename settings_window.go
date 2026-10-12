package main

import (
	"fmt"
	"reflect"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
	"github.com/smallnest/uix/components/colorwell"
	"github.com/smallnest/uix/components/field"
	"github.com/smallnest/uix/components/input"
	"github.com/smallnest/uix/components/numberinput"
	uixselect "github.com/smallnest/uix/components/select"
	"github.com/smallnest/uix/components/slider"
	uixswitch "github.com/smallnest/uix/components/switch"
)

// settingsWindow is the Settings window: otty's categories on the left,
// the chosen one's controls on the right, every change applied live and
// saved. It edits the app's settings in place, so the main window and the
// settings agree.
type settingsWindow struct {
	a     *app
	s     *settings
	win   *mygo.Window
	last  settings // the applied snapshot; a difference means a change
	page  string   // the chosen category
	wells colorWells
}

// colorWells hold the buffers of the color pickers. An unused well shows
// the theme's color; editing it makes the setting, and Reset forgets it.
type colorWells struct {
	init         bool
	foreground, background, cursor, cursorText, selection, selectionText colorWellState
}

type colorWellState struct {
	color ui.Color
	base  ui.Color
	inUse bool
}

// settingsCategories are otty's own, in its order; the ones without a
// page of their own show a "not in this build" notice.
var settingsCategories = []struct{ id, title string }{
	{"general", "General"},
	{"profiles", "Profiles"},
	{"shell", "Shell"},
	{"remote", "Remote"},
	{"controls", "Controls"},
	{"editor", "Editor"},
	{"view", "View"},
	{"appearance", "Appearance"},
	{"key_bindings", "Key Bindings"},
	{"advanced", "Advanced"},
	{"agents", "Agents"},
	{"recipes", "Recipes"},
	{"data", "Data"},
	{"account", "Account"},
}

// newSettingsWindow opens the settings of the app: one window, reused.
func newSettingsWindow(a *app) *mygo.Window {
	sw := &settingsWindow{
		a:    a,
		s:    a.settings,
		last: *a.settings,
		page: "appearance",
	}
	sw.win = mygo.NewWindow(mygo.WindowOptions{
		Title:      "Settings",
		Width:      760,
		Height:     540,
		MinWidth:   600,
		MinHeight:  420,
		Content:    ui.View(sw.view),
	})
	sw.win.OnClosed(func() { a.settingsWin = nil })
	return sw.win
}

func (sw *settingsWindow) view(c *ui.Context) {
	c.SetTheme(ottyTheme(c.Theme().Dark))
	t := c.Theme()
	ui.Row(c).Fill().Children(func() {
		// The categories, a column the surface color of the app's rail.
		ui.Column(c).Width(200).FillHeight().Shrink(0).Background(t.Surface).
			BorderWidth(0, 0, 0, 1).BorderColor(t.Border).Children(func() {
				ui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
					ui.Column(c).FillWidth().Padding(10).Gap(1).Children(func() {
						for _, cat := range settingsCategories {
							sw.categoryItem(c, cat.id, cat.title)
						}
					})
				})
			})
		// The chosen category's controls.
		ui.Scroll(c).Grow(1).MinWidth(0).Children(func() {
			ui.Column(c).FillWidth().Padding(24, 28).MaxWidth(560).Gap(18).Children(func() {
				for _, cat := range settingsCategories {
					if cat.id == sw.page {
						ui.Text(c, cat.title).FontSize(17).Bold()
					}
				}
				switch sw.page {
				case "general":
					sw.pageGeneral(c)
				case "shell":
					sw.pageShell(c)
				case "controls":
					sw.pageControls(c)
				case "view":
					sw.pageView(c)
				case "appearance":
					sw.pageAppearance(c)
				default:
					sw.pagePlaceholder(c)
				}
			})
		})
	})
	// A change the controls made applies and saves.
	sw.sync()
}

// categoryItem is one row of the category list; the chosen one is shown.
func (sw *settingsWindow) categoryItem(c *ui.Context, id, title string) {
	t := c.Theme()
	b := ui.ButtonBase(c.Key("set-cat-" + id)).FillWidth().
		Padding(6, 10).Radius(t.Radius).FocusRing(false)
	if sw.page == id {
		b.Background(t.SurfacePressed)
	}
	b.Children(func() {
		ui.Text(c, title).FontSize(13).SingleLine().TextColor(sw.categoryColor(t, id))
	})
	if b.Clicked() {
		sw.page = id
	}
}

func (sw *settingsWindow) categoryColor(t *ui.Theme, id string) ui.Color {
	if sw.page == id {
		return t.Text
	}
	return t.TextMuted
}

// pageGeneral is the General category: how tabs close.
func (sw *settingsWindow) pageGeneral(c *ui.Context) {
	sw.group(c, "Closing", func() {
		uixswitch.Switch(c, uixswitch.Props{
			On: &sw.s.ConfirmCloseTab, Label: "Confirm before closing a tab",
		})
		sw.hint(c, "Ask before a tab with a running program closes.")
	})
}

// pageShell is the Shell category: what new sessions run and where.
func (sw *settingsWindow) pageShell(c *ui.Context) {
	sw.group(c, "Session", func() {
		field.Field(c, field.Props{
			Label: "Command", Hint: "The shell new sessions run. Empty uses your login shell.",
		}, func() {
			input.Input(c, input.Props{Value: &sw.s.Command, Placeholder: shellName()})
		})
		field.Field(c, field.Props{
			Label: "Working Directory", Hint: "Where new sessions start. Empty uses your home directory.",
		}, func() {
			input.Input(c, input.Props{Value: &sw.s.WorkingDirectory, Placeholder: homeDir()})
		})
	})
}

// pageControls is the Controls category: how the keyboard behaves.
func (sw *settingsWindow) pageControls(c *ui.Context) {
	sw.group(c, "Keyboard", func() {
		uixswitch.Switch(c, uixswitch.Props{
			On: &sw.s.OptionAsAlt, Label: "Option as Meta",
		})
		sw.hint(c, "The Option key types Meta to programs, not accented letters and symbols.")
	})
}

// pageView is the View category: the rail and the window.
func (sw *settingsWindow) pageView(c *ui.Context) {
	sw.group(c, "Sidebar", func() {
		uixswitch.Switch(c, uixswitch.Props{
			On: &sw.s.SidebarVisible, Label: "Show the tabs rail",
		})
		sw.row(c, "Rail width", func() {
			numberinput.NumberInput(c, numberinput.Props{Value: &sw.s.SidebarWidth, Lo: 160, Hi: 420, Step: 4})
		})
	})
	sw.group(c, "Window", func() {
		sw.row(c, "Width", func() {
			numberinput.NumberInput(c, numberinput.Props{Value: &sw.s.WindowWidth, Lo: 680, Hi: 3200, Step: 20})
		})
		sw.row(c, "Height", func() {
			numberinput.NumberInput(c, numberinput.Props{Value: &sw.s.WindowHeight, Lo: 440, Hi: 2400, Step: 20})
		})
	})
}

// pageAppearance is the Appearance category: the themes, the font, the
// cursor, the colors and the opacity, otty's appearance settings.
func (sw *settingsWindow) pageAppearance(c *ui.Context) {
	sw.group(c, "Theme", func() {
		sw.row(c, "Theme", func() {
			uixselect.Select(c, uixselect.Props{
				Selected: &sw.s.ThemeName, Options: themeNames(),
				Placeholder: "Follow system",
			})
		})
		uixswitch.Switch(c, uixswitch.Props{
			On: &sw.s.AutoThemeDark, Label: "Use a separate dark theme",
		})
		sw.row(c, "Dark theme", func() {
			sel := uixselect.Select(c, uixselect.Props{
				Selected: &sw.s.ThemeDark, Options: themeNames(),
				Placeholder: "Same as light",
			})
			if !sw.s.AutoThemeDark {
				sel.Disabled(true)
			}
		})
	})
	sw.group(c, "Font", func() {
		sw.row(c, "Font family", func() {
			uixselect.Select(c, uixselect.Props{
				Selected: &sw.s.FontFamily,
				Options:  []string{"JetBrains Mono", "Menlo", "Monaco", "SF Mono", "monospace"},
			})
		})
		sw.row(c, "Font size", func() {
			numberinput.NumberInput(c, numberinput.Props{Value: &sw.s.FontSize, Lo: 8, Hi: 26, Step: 1})
		})
		uixswitch.Switch(c, uixswitch.Props{
			On: &sw.s.FontLigatures, Label: "Ligatures",
		})
		uixswitch.Switch(c, uixswitch.Props{
			On: &sw.s.FontThicken, Label: "Thicker text",
		})
		slider.Slider(c, slider.Props{
			Value: &sw.s.AdjustCellHeight, Min: 0, Max: 0.4, Step: 0.05,
			Label: "Cell height", Format: func(v float64) string { return percent(v) },
		})
	})
	sw.group(c, "Cursor", func() {
		sw.row(c, "Style", func() {
			uixselect.Select(c, uixselect.Props{
				Selected: &sw.s.CursorStyle,
				Options:  []string{"block", "bar", "underline"},
			})
		})
		uixswitch.Switch(c, uixswitch.Props{
			On: &sw.s.CursorBlink, Label: "Blink",
		})
	})
	sw.group(c, "Colors", func() {
		sw.hint(c, "Set a terminal color by hand; Reset returns it to the theme's.")
		sw.colorRow(c, "Foreground", &sw.wells.foreground, &sw.s.Foreground)
		sw.colorRow(c, "Background", &sw.wells.background, &sw.s.Background)
		sw.colorRow(c, "Cursor", &sw.wells.cursor, &sw.s.CursorColor)
		sw.colorRow(c, "Cursor text", &sw.wells.cursorText, &sw.s.CursorText)
		sw.colorRow(c, "Selection", &sw.wells.selection, &sw.s.Selection)
		sw.colorRow(c, "Selection text", &sw.wells.selectionText, &sw.s.SelectionText)
	})
	sw.group(c, "Opacity", func() {
		slider.Slider(c, slider.Props{
			Value: &sw.s.WindowOpacity, Min: 0.4, Max: 1, Step: 0.05,
			Label: "Window", Format: percent,
		})
		slider.Slider(c, slider.Props{
			Value: &sw.s.TerminalOpacity, Min: 0.4, Max: 1, Step: 0.05,
			Label: "Terminal background", Format: percent,
		})
		sw.hint(c, "The terminal background's opacity reaches new sessions.")
	})
}

// pagePlaceholder is a category this build does not implement: the app
// has no account, remote, recipe or agent work, and no built-in editor.
func (sw *settingsWindow) pagePlaceholder(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).FillWidth().Gap(4).Children(func() {
		ui.Text(c, "Not included in this build").TextColor(t.TextMuted)
		ui.Text(c, "This setting needs a feature Nutty does not have yet.").
			FontSize(12).TextColor(t.TextMuted)
	})
}

// --- layout helpers --------------------------------------------------------

// group is a titled section of a page, with the controls inside.
func (sw *settingsWindow) group(c *ui.Context, title string, fn func()) {
	t := c.Theme()
	ui.Column(c).FillWidth().Gap(10).Children(func() {
		ui.Text(c, title).FontSize(12).Bold().TextColor(t.TextMuted)
		ui.Column(c).FillWidth().Gap(8).Children(fn)
	})
}

// row is a labeled control on one line.
func (sw *settingsWindow) row(c *ui.Context, label string, fn func()) {
	t := c.Theme()
	ui.Row(c).FillWidth().Gap(12).Children(func() {
		ui.Text(c, label).FontSize(13).SingleLine().Width(120).Shrink(0).TextColor(t.Text)
		ui.Box(c).Grow(1).MinWidth(0).Children(fn)
	})
}

// hint is a muted explanation under a control.
func (sw *settingsWindow) hint(c *ui.Context, text string) {
	ui.Text(c, text).FontSize(11).TextColor(c.Theme().TextMuted)
}

// colorRow is a named color well with a Reset that forgets the setting.
func (sw *settingsWindow) colorRow(c *ui.Context, label string, st *colorWellState, set *string) {
	t := c.Theme()
	ui.Row(c).FillWidth().Gap(12).Children(func() {
		ui.Text(c, label).FontSize(13).SingleLine().Width(120).Shrink(0).TextColor(t.Text)
		colorwell.ColorWell(c, colorwell.Props{Color: &st.color, Label: label})
		if st.inUse {
			rb := ui.ButtonBase(c.Key("reset-"+label)).Padding(2, 8).Radius(t.Radius).
				FocusRing(false).Tooltip("Use the theme's color")
			rb.Children(func() {
				ui.Row(c).Gap(4).Children(func() {
					ui.Image(c, lucide.RotateCcw()).Size(12, 12)
					ui.Text(c, "Reset").FontSize(12).TextColor(t.TextMuted)
				})
			})
			if rb.Clicked() {
				*set = ""
				st.inUse = false
			}
		}
	})
}

// --- applying changes ------------------------------------------------------

// sync turns the color wells' buffers into settings, and applies and saves
// whatever differs from the last applied snapshot. The selects store their
// "Follow system" choice as that literal; here it becomes the empty name
// that means following the window.
func (sw *settingsWindow) sync() {
	sw.syncColors()
	if sw.s.ThemeName == "Follow system" {
		sw.s.ThemeName = ""
	}
	if sw.s.ThemeDark == "Follow system" {
		sw.s.ThemeDark = ""
	}
	if !reflect.DeepEqual(*sw.s, sw.last) {
		sw.last = *sw.s
		sw.a.applySettings()
		sw.a.saveSettings()
	}
}

// syncColors writes the color wells back into the settings: an unused well
// follows the theme's color, editing one makes the setting, and a change
// of theme refreshes the unused ones.
func (sw *settingsWindow) syncColors() {
	th := themeFor(sw.s.ThemeName, false)
	if th == nil {
		th = ottyTermTheme(false)
	}
	// The theme's cursor and selection are zero in the window colors; the
	// terminal falls back to the foreground for the cursor and an accent
	// for the selection, so the wells show what would render.
	cursorDef := th.Cursor
	if cursorDef.A == 0 {
		cursorDef = th.Foreground
	}
	selDef := th.Selection
	if selDef.A == 0 {
		selDef = ui.Hex("#A8C7FA")
	}
	sync := func(st *colorWellState, hex string, def ui.Color, set *string) {
		if !sw.wells.init {
			if hex != "" {
				st.color = ui.Hex(hex)
				st.inUse = true
			} else {
				st.color = def
			}
			st.base = st.color
			return
		}
		if !st.inUse {
			st.base, st.color = def, def
		}
		if st.color != st.base {
			st.inUse = true
			st.base = st.color
			*set = colorHex(st.color)
		} else if st.inUse {
			*set = colorHex(st.color)
		}
	}
	sync(&sw.wells.foreground, sw.s.Foreground, th.Foreground, &sw.s.Foreground)
	sync(&sw.wells.background, sw.s.Background, th.Background, &sw.s.Background)
	sync(&sw.wells.cursor, sw.s.CursorColor, cursorDef, &sw.s.CursorColor)
	sync(&sw.wells.cursorText, sw.s.CursorText, th.Foreground, &sw.s.CursorText)
	sync(&sw.wells.selection, sw.s.Selection, selDef, &sw.s.Selection)
	sync(&sw.wells.selectionText, sw.s.SelectionText, th.Foreground, &sw.s.SelectionText)
	sw.wells.init = true
}

func percent(v float64) string { return fmt.Sprintf("%.0f%%", v*100) }
