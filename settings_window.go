package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
	"github.com/smallnest/uix/components/colorwell"
	"github.com/smallnest/uix/components/field"
	"github.com/smallnest/uix/components/input"
	"github.com/smallnest/uix/components/numberinput"
	uixselect "github.com/smallnest/uix/components/select"
	"github.com/smallnest/uix/components/textarea"
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
	lists map[string]*string // def key -> list editor buffer
}

// colorWells hold the buffers of the color pickers, one per color
// setting. An unused well shows the theme's color; editing it makes the
// setting, and Reset forgets it.
type colorWells struct {
	init  bool
	state map[string]*colorWellState
}

type colorWellState struct {
	color ui.Color
	base  ui.Color
	inUse bool
}

// settingsCategories are otty's own, in its order; the pages without a
// real implementation show a "not in this build" notice.
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
	sw.syncColors() // the wells show the theme's colors before the controls draw
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
				case "appearance":
					sw.pageAppearance(c)
				case "general":
					sw.pageFromGroups(c, "general")
				case "shell":
					sw.pageFromGroups(c, "shell")
				case "controls":
					sw.pageFromGroups(c, "controls")
				case "editor":
					sw.pageFromGroups(c, "editor")
				case "view":
					sw.pageView(c)
				case "recipes":
					sw.pageFromGroups(c, "recipes")
				case "agents":
					sw.pageFromGroups(c, "agents")
				case "data":
					sw.pageData(c)
				case "account":
					sw.pageAccount(c)
				case "remote":
					sw.pageRemote(c)
				case "profiles":
					sw.pageProfiles(c)
				case "key_bindings":
					sw.pageKeyBindings(c)
				case "advanced":
					sw.pageAdvanced(c)
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

// --- the registry-driven pages --------------------------------------------

// pageFromGroups renders one category's groups from the settings registry:
// the sections of settingsGroups[id], each setting drawn from its def.
func (sw *settingsWindow) pageFromGroups(c *ui.Context, id string) {
	for _, g := range settingsGroups[id] {
		sw.group(c, g.title, func() {
			for _, key := range g.keys {
				if d := defByKey(key); d != nil {
					sw.renderSetting(c, *d)
				}
			}
		})
	}
}

// renderSetting draws one setting from its def: a switch, a select, a
// number, a color well, a text field or a list box. A direct field
// binding keeps the live controls' state across frames; the others write
// their value back through the def's setter when they change.
func (sw *settingsWindow) renderSetting(c *ui.Context, d settingDef) {
	switch d.kind {
	case "bool":
		uixswitch.Switch(c, uixswitch.Props{
			On: d.ptr(sw.s).(*bool), Label: d.label,
		})
		if d.desc != "" {
			sw.hint(c, d.desc)
		}
	case "enum":
		field.Field(c, field.Props{Label: d.label, Hint: d.desc}, func() {
			labels := make([]string, len(d.enum))
			for i, o := range d.enum {
				labels[i] = o.label
			}
			cur := d.get(sw.s).(string)
			show := enumLabel(d, cur)
			uixselect.Select(c, uixselect.Props{Selected: &show, Options: labels})
			if show != enumLabel(d, cur) {
				d.set(sw.s, enumValue(d, show))
			}
		})
	case "number":
		field.Field(c, field.Props{Label: d.label, Hint: d.desc}, func() {
			numberinput.NumberInput(c, numberinput.Props{
				Value: d.ptr(sw.s).(*float64), Lo: d.lo, Hi: d.hi, Step: d.step,
			})
		})
	case "string":
		field.Field(c, field.Props{Label: d.label, Hint: d.desc}, func() {
			input.Input(c, input.Props{Value: d.ptr(sw.s).(*string)})
		})
	case "list":
		field.Field(c, field.Props{Label: d.label, Hint: d.desc}, func() {
			sw.renderList(c, d)
		})
	case "color":
		sw.renderColorRow(c, d)
	}
}

// renderEnum is the select of an enum setting: it shows the option
// labels but stores the option values, so a value with a friendlier
// label (like "Off (typing accents)" for false) still saves the key.
func enumLabel(d settingDef, v string) string {
	for _, o := range d.enum {
		if o.value == v {
			return o.label
		}
	}
	return v
}

func enumValue(d settingDef, label string) string {
	for _, o := range d.enum {
		if o.label == label {
			return o.value
		}
	}
	return label
}

// renderList edits a list-of-strings setting as one item per line. The
// buffer lives in the window, because the text box must keep the same
// pointer across frames or its caret resets; it writes the split list
// back only when the text changed.
func (sw *settingsWindow) renderList(c *ui.Context, d settingDef) {
	if sw.lists == nil {
		sw.lists = map[string]*string{}
	}
	b, ok := sw.lists[d.key]
	if !ok {
		joined := strings.Join(d.get(sw.s).([]string), "\n")
		b = &joined
		sw.lists[d.key] = b
	}
	textarea.Textarea(c, textarea.Props{Value: b, Rows: 4})
	if cur := strings.Join(d.get(sw.s).([]string), "\n"); *b != cur {
		d.set(sw.s, splitLines(*b))
	}
}

// splitLines is a list-of-strings from a text box's lines, blank and
// whitespace-only lines dropped.
func splitLines(s string) []string {
	var out []string
	for _, ln := range strings.Split(s, "\n") {
		if ln = strings.TrimSpace(ln); ln != "" {
			out = append(out, ln)
		}
	}
	return out
}

// renderColorRow is one named color setting: the well, and a Reset that
// forgets the setting and returns the well to the theme's color.
func (sw *settingsWindow) renderColorRow(c *ui.Context, d settingDef) {
	t := c.Theme()
	st := sw.wellState(d)
	field.Field(c, field.Props{Label: d.label, Hint: d.desc}, func() {
		ui.Row(c).Gap(12).Children(func() {
			colorwell.ColorWell(c, colorwell.Props{Color: &st.color, Label: d.label})
			if st.inUse {
				rb := ui.ButtonBase(c.Key("reset-" + d.key)).Padding(2, 8).Radius(t.Radius).
					FocusRing(false).Tooltip("Use the theme's color")
				rb.Children(func() {
					ui.Row(c).Gap(4).Children(func() {
						ui.Image(c, lucide.RotateCcw()).Size(12, 12)
						ui.Text(c, "Reset").FontSize(12).TextColor(t.TextMuted)
					})
				})
				if rb.Clicked() {
					d.set(sw.s, "")
					st.inUse = false
				}
			}
		})
	})
}

// wellState is the picker state of one color setting.
func (sw *settingsWindow) wellState(d settingDef) *colorWellState {
	if sw.wells.state == nil {
		sw.wells.state = map[string]*colorWellState{}
	}
	st, ok := sw.wells.state[d.key]
	if !ok {
		st = &colorWellState{}
		sw.wells.state[d.key] = st
	}
	return st
}

// pageView is the View category: the registry's tab and badge settings,
// plus the details panel's views, each with its own enable switch.
func (sw *settingsWindow) pageView(c *ui.Context) {
	sw.pageFromGroups(c, "view")
	sw.group(c, "Details Panel", func() {
		sw.hint(c, "The tabs of the panel on the right. Disable one to hide it.")
		for i := range sw.s.DetailsViews {
			i := i
			v := &sw.s.DetailsViews[i]
			uixswitch.Switch(c, uixswitch.Props{On: &v.Enabled, Label: v.Title})
		}
	})
}

// pageAppearance is the Appearance category: the registry's theme, font,
// cursor, window and terminal groups, with the sixteen palette colors
// drawn as a grid inside the Colors section, as otty's appearance does.
func (sw *settingsWindow) pageAppearance(c *ui.Context) {
	for _, g := range settingsGroups["appearance"] {
		sw.group(c, g.title, func() {
			for _, key := range g.keys {
				if d := defByKey(key); d != nil {
					sw.renderSetting(c, *d)
				}
			}
			if g.title == "Colors" {
				sw.paletteGrid(c)
			}
		})
	}
}

// paletteGrid is the sixteen palette wells in a grid, and a reset that
// returns them all to the theme's colors.
func (sw *settingsWindow) paletteGrid(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).FillWidth().Gap(8).Children(func() {
		ui.Text(c, "Palette").FontSize(11).Bold().TextColor(t.TextMuted)
		for row := 0; row < 4; row++ {
			ui.Row(c).Gap(10).Children(func() {
				for col := 0; col < 4; col++ {
					i := row*4 + col
					d := *defByKey(fmt.Sprintf("palette-%d", i))
					st := sw.wellState(d)
					colorwell.ColorWell(c, colorwell.Props{Color: &st.color, Label: d.label})
				}
			})
		}
		if sw.s.anyPalette() {
			rb := ui.ButtonBase(c.Key("reset-palette")).Padding(2, 8).Radius(t.Radius).
				FocusRing(false).Tooltip("Use the theme's colors")
			rb.Children(func() {
				ui.Row(c).Gap(4).Children(func() {
					ui.Image(c, lucide.RotateCcw()).Size(12, 12)
					ui.Text(c, "Reset palette").FontSize(12).TextColor(t.TextMuted)
				})
			})
			if rb.Clicked() {
				for i := range sw.s.Palette {
					sw.s.Palette[i] = ""
				}
				for i := 0; i < 16; i++ {
					if st := sw.wells.state[fmt.Sprintf("palette-%d", i)]; st != nil {
						st.inUse = false
					}
				}
			}
		}
	})
}

// --- pages without a real implementation yet ------------------------------

// pageProfiles is the Profiles category; the editor lands in a later pass.
func (sw *settingsWindow) pageProfiles(c *ui.Context) {
	sw.pageUnavailable(c, "Profiles", "Profiles come next; the default profile starts every new pane.")
}

// pageKeyBindings is the Key Bindings category; the recorder lands in a
// later pass.
func (sw *settingsWindow) pageKeyBindings(c *ui.Context) {
	sw.pageUnavailable(c, "Key Bindings", "The key binding recorder comes next.")
}

// pageAdvanced is the Advanced category; the All Settings list lands in a
// later pass.
func (sw *settingsWindow) pageAdvanced(c *ui.Context) {
	sw.pageUnavailable(c, "Advanced", "The All Settings list comes next.")
}

// pageData is the Data category: cloud sync, which this build has no
// backend for. Otty shows the same notice when the feature is absent.
func (sw *settingsWindow) pageData(c *ui.Context) {
	sw.pageUnavailable(c, "Sync & Backup", "Not available in this build")
}

// pageAccount is the Account category: the app is signed out.
func (sw *settingsWindow) pageAccount(c *ui.Context) {
	sw.pageUnavailable(c, "Account", "Not signed in")
}

// pageRemote is the Remote category, otty's own notice for it.
func (sw *settingsWindow) pageRemote(c *ui.Context) {
	sw.pageUnavailable(c, "Remote", "Not available in this build.")
}

// pageUnavailable is a category with a heading and a muted notice.
func (sw *settingsWindow) pageUnavailable(c *ui.Context, title, notice string) {
	t := c.Theme()
	sw.group(c, title, func() {
		ui.Text(c, notice).FontSize(12).TextColor(t.TextMuted)
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

// hint is a muted explanation under a control.
func (sw *settingsWindow) hint(c *ui.Context, text string) {
	ui.Text(c, text).FontSize(11).TextColor(c.Theme().TextMuted)
}

// --- applying changes ------------------------------------------------------

// sync applies and saves whatever differs from the last applied snapshot.
// The selects store their "Follow system" theme choice as the empty name
// that means following the window.
func (sw *settingsWindow) sync() {
	if !reflect.DeepEqual(*sw.s, sw.last) {
		sw.last = *sw.s
		sw.a.applySettings()
		sw.a.saveSettings()
	}
}

// syncColors writes the color wells back into the settings: an unused
// well follows the theme's color, editing one makes the setting, and a
// change of theme refreshes the unused ones.
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
	defColor := func(d settingDef) ui.Color {
		switch d.key {
		case "cursor-color":
			return cursorDef
		case "selection-background":
			return selDef
		}
		var i int
		if n, _ := fmt.Sscanf(d.key, "palette-%d", &i); n == 1 {
			return th.Palette[i]
		}
		return th.Foreground // foreground, cursor text, selection text
	}
	sync := func(d settingDef, set func(string)) {
		st := sw.wellState(d)
		hex := d.get(sw.s).(string)
		def := defColor(d)
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
			set(colorHex(st.color))
		} else if st.inUse {
			set(colorHex(st.color))
		}
	}
	for _, key := range []string{"foreground", "background", "cursor-color", "cursor-text", "selection-background", "selection-foreground"} {
		key := key
		d := *defByKey(key)
		sync(d, func(v string) { d.set(sw.s, v) })
	}
	for i := 0; i < 16; i++ {
		i := i
		d := *defByKey(fmt.Sprintf("palette-%d", i))
		sync(d, func(v string) { d.set(sw.s, v) })
	}
	sw.wells.init = true
}
