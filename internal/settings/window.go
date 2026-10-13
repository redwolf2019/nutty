package settings

import (
	"bytes"
	"encoding/json"
	"fmt"

	"nutty/internal/i18n"

	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
	"github.com/smallnest/uix/components/colorwell"
	"github.com/smallnest/uix/components/field"
	"github.com/smallnest/uix/components/input"
	"github.com/smallnest/uix/components/numberinput"
	uixselect "github.com/smallnest/uix/components/select"
	uixsettings "github.com/smallnest/uix/components/settings"
	uixswitch "github.com/smallnest/uix/components/switch"
	"github.com/smallnest/uix/components/textarea"
)

// settingsWindow is the Settings window. It is the shell of onevra's
// Settings component: otty's categories on the left, searchable, each
// category's groups and controls on the right, every change applied live
// and saved. It edits the app's settings in place, so the main window and
// the settings agree.
type settingsWindow struct {
	app            App
	s              *Settings
	win            *mygo.Window
	last           Settings                  // the applied snapshot; a difference means a change
	lastJSON       []byte                    // the snapshot as JSON, what sync compares
	state          uixsettings.SettingsState // the category and search the shell shows
	lastCat        string                    // the category of the last frame, the breath's hook
	wells          colorWells
	lists          map[string]*string // def key -> list editor buffer
	keys           keybindPage
	settingsSearch string // the Advanced page's filter
	profileSel     int    // the profile the editor shows, or -1
	dataBreath     uint64 // clicks of the Data row, its icon's breath count
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

func (sw *settingsWindow) view(c *ui.Context) {
	c.SetTheme(OttyTheme(c.Theme().Dark))
	i18n.SetLang(sw.s.Language) // the shell's chrome follows the setting
	sw.syncColors()             // the wells show the theme's colors before the controls draw
	uixsettings.Settings(c, &sw.state, sw.settingsCategoriesFor(), uixsettings.SettingsOptions{
		Title:             i18n.Tr("Settings"),
		SearchLabel:       i18n.Tr("Search settings"),
		SearchPlaceholder: i18n.Tr("Search settings"),
		ResultsTitle:      i18n.Tr("Search results"),
		EmptyText:         i18n.Tr("No settings found"),
	})
	// The shell's own row click sets the selection; a click of the Data
	// row also breathes its database icon, one cycle per click.
	if sw.state.Category == "data" && sw.state.Category != sw.lastCat {
		sw.dataBreath++
	}
	sw.lastCat = sw.state.Category
	// A change the controls made applies and saves.
	sw.sync()
}

// settingsCategoriesFor is the shell's navigation: otty's categories in
// their order, each with the groups and items of its page. It is rebuilt
// every frame, so the items' closures see this frame's values; nothing is
// cached.
func (sw *settingsWindow) settingsCategoriesFor() []uixsettings.SettingsCategory {
	cats := make([]uixsettings.SettingsCategory, 0, len(settingsCategories))
	for _, cat := range settingsCategories {
		cats = append(cats, uixsettings.SettingsCategory{
			ID:      cat.id,
			Title:   i18n.Tr(cat.title),
			Leading: sw.categoryLeading(cat.id),
			Groups:  sw.groupsFor(cat.id),
		})
	}
	return cats
}

// categoryLeading is a sidebar row's icon: otty's own SVG, muted or lit
// with the selection, and the Data row's icon breathes on each click of
// the row (see breatheData).
func (sw *settingsWindow) categoryLeading(id string) func(c *ui.Context, size float32) {
	return func(c *ui.Context, size float32) {
		if icon := settingsCategoryIcon(id); icon != nil {
			img := ui.Image(c, icon).Size(size, size).TextColor(sw.categoryColor(c.Theme(), id))
			if id == "data" {
				sw.breatheData(c, img)
			}
		}
	}
}

func (sw *settingsWindow) categoryColor(t *ui.Theme, id string) ui.Color {
	if sw.state.Category == id {
		return t.Text
	}
	return t.TextMuted
}

// breatheData plays the data icon's breath: a click of the Data row puffs
// its database once, as otty's SMIL morph replays on the click of the row.
// mygo parses SVG statically, so the puff is a Keyframes scale on the icon's
// pose instead of the path morph: one cycle per click, on otty's 0.983 s
// period, held at rest until the next click.
func (sw *settingsWindow) breatheData(c *ui.Context, img ui.Element) {
	if sw.dataBreath == 0 {
		return // the row was never clicked; nothing to breathe
	}
	scale := img.Keyframes("breath", []ui.Keyframe{
		{At: 0, Value: 1, Ease: ui.EaseInOut},
		{At: 333 * time.Millisecond, Value: 1.08, Ease: ui.EaseInOut},
		{At: 667 * time.Millisecond, Value: 1.08, Ease: ui.Linear},
		{At: 983 * time.Millisecond, Value: 1, Ease: ui.EaseInOut},
	}, ui.KeyframeOptions{Run: sw.dataBreath})
	img.AnimatePose(ui.Pose{Scale: scale}, ui.Smooth)
}

// --- the categories of the shell -------------------------------------------

// groupsFor is one category's groups for the shell. The pages the
// registry drives map straight to groups of items; Profiles, Key
// Bindings and Advanced draw their whole page in one item, as do the
// notices of the pages without a form.
func (sw *settingsWindow) groupsFor(id string) []uixsettings.SettingsGroup {
	switch id {
	case "appearance":
		return sw.appearanceGroups()
	case "view":
		return sw.viewGroups()
	case "profiles":
		return []uixsettings.SettingsGroup{pageItem(func(c *ui.Context) { sw.pageProfiles(c) })}
	case "key_bindings":
		return []uixsettings.SettingsGroup{pageItem(func(c *ui.Context) { sw.pageKeyBindings(c) })}
	case "advanced":
		return []uixsettings.SettingsGroup{pageItem(func(c *ui.Context) { sw.pageAdvanced(c) })}
	case "data":
		return noticeGroup(i18n.Tr("Sync & Backup"), i18n.Tr("Not available in this build"))
	case "account":
		return noticeGroup(i18n.Tr("Account"), i18n.Tr("Not signed in"))
	case "remote":
		return noticeGroup(i18n.Tr("Remote"), i18n.Tr("Not available in this build."))
	default:
		return sw.registryGroups(id)
	}
}

// pageItem is the group of a category that draws its whole page itself:
// an unheaded group with one item that fills the row. The shell's row
// shows an empty title beside the page, and both ask for the row's whole
// width. The page barely shrinks, so the empty title gives its share up
// first and the page takes the row. A fixed minimum did that too, but it
// was wider than the row of the window at its default size, where the
// page's right side, the key bindings' shortcuts with it, ran off the
// window.
func pageItem(build func(c *ui.Context)) uixsettings.SettingsGroup {
	return uixsettings.SettingsGroup{ID: "page", Items: []uixsettings.SettingsItem{{
		ID: "page",
		Build: func(c *ui.Context) {
			ui.Column(c).FillWidth().Shrink(0.0001).Children(func() { build(c) })
		},
	}}}
}

// noticeGroup is a category without a form: a heading and the muted
// notice otty shows for the absent feature.
func noticeGroup(title, notice string) []uixsettings.SettingsGroup {
	return []uixsettings.SettingsGroup{{
		ID: title, Title: title,
		Items: []uixsettings.SettingsItem{{ID: "notice", Description: notice}},
	}}
}

// registryGroups is the shell's groups of one page: each otty section of
// the registry becomes a group, each setting an item whose row shows its
// label and description and whose control edits the value.
func (sw *settingsWindow) registryGroups(id string) []uixsettings.SettingsGroup {
	groups := make([]uixsettings.SettingsGroup, 0, len(settingsGroups[id]))
	for _, g := range settingsGroups[id] {
		items := make([]uixsettings.SettingsItem, 0, len(g.keys))
		for _, key := range g.keys {
			if d := defByKey(key); d != nil {
				d := *d
				items = append(items, uixsettings.SettingsItem{
					ID: d.key, Title: i18n.Tr(d.label), Description: i18n.Tr(d.desc),
					Keywords: []string{d.key},
					Build:    func(c *ui.Context) { sw.control(c, d) },
				})
			}
		}
		groups = append(groups, uixsettings.SettingsGroup{ID: g.title, Title: i18n.Tr(g.title), Items: items})
	}
	return groups
}

// appearanceGroups is the Appearance page: the registry's groups, with
// the sixteen palette colors as a grid inside the Colors section, as
// otty draws them.
func (sw *settingsWindow) appearanceGroups() []uixsettings.SettingsGroup {
	groups := sw.registryGroups("appearance")
	for i := range groups {
		if groups[i].ID == "Colors" {
			groups[i].Items = append(groups[i].Items, uixsettings.SettingsItem{
				ID: "palette", Title: i18n.Tr("Palette"),
				Description: i18n.Tr("The sixteen ANSI colors of the terminal palette."),
				Keywords:    []string{"palette", "ansi", "color"},
				Build:       func(c *ui.Context) { sw.paletteGrid(c) },
			})
		}
	}
	return groups
}

// viewGroups is the View page: the registry's groups, and the Details
// Panel's tabs, each with its enable switch and the arrows to reorder it.
func (sw *settingsWindow) viewGroups() []uixsettings.SettingsGroup {
	groups := sw.registryGroups("view")
	details := uixsettings.SettingsGroup{ID: "Details Panel", Title: i18n.Tr("Details Panel")}
	for _, v := range sw.s.DetailsViews {
		id, title := v.ID, v.Title
		details.Items = append(details.Items, uixsettings.SettingsItem{
			ID: id, Title: i18n.Tr(title),
			Description: i18n.Tr("A tab of the details panel on the right; the arrows reorder it."),
			Keywords:    []string{"details", "panel"},
			Build: func(c *ui.Context) {
				p := sw.DetailsView(id)
				if p == nil {
					return
				}
				idx := sw.detailsViewIndex(id)
				ui.Row(c).Gap(8).Children(func() {
					uixswitch.Switch(c, uixswitch.Props{On: &p.Enabled})
					ui.Box(c).Grow(1)
					sw.moveButton(c, "up-"+id, lucide.ChevronUp(), idx > 0, func() { sw.moveViewByID(id, -1) })
					sw.moveButton(c, "down-"+id, lucide.ChevronDown(), idx < len(sw.s.DetailsViews)-1, func() { sw.moveViewByID(id, 1) })
				})
			},
		})
	}
	return append(groups, details)
}

// DetailsView is the details panel's tab of an id, or nil. The item
// looks it up on the frame it builds, so a reorder never leaves a stale
// pointer behind.
func (sw *settingsWindow) DetailsView(id string) *DetailsView {
	for i := range sw.s.DetailsViews {
		if sw.s.DetailsViews[i].ID == id {
			return &sw.s.DetailsViews[i]
		}
	}
	return nil
}

func (sw *settingsWindow) detailsViewIndex(id string) int {
	for i := range sw.s.DetailsViews {
		if sw.s.DetailsViews[i].ID == id {
			return i
		}
	}
	return -1
}

// moveViewByID shifts a details view by its id in the list, clamping at
// the ends.
func (sw *settingsWindow) moveViewByID(id string, delta int) {
	i := sw.detailsViewIndex(id)
	v := sw.s.DetailsViews
	j := i + delta
	if i < 0 || j < 0 || j >= len(v) {
		return
	}
	v[i], v[j] = v[j], v[i]
}

// --- the Settings' controls ------------------------------------------------

// control draws one setting's control alone: the shell's item row shows
// the label and description, so only the control belongs here. The
// Advanced page's All Settings list keeps the field wrapper (see
// renderSetting).
func (sw *settingsWindow) control(c *ui.Context, d settingDef) {
	switch d.kind {
	case "bool":
		uixswitch.Switch(c, uixswitch.Props{On: d.ptr(sw.s).(*bool)})
	case "enum":
		cur := d.get(sw.s).(string)
		show := enumLabel(d, cur)
		labels := make([]string, len(d.enum))
		for i, o := range d.enum {
			labels[i] = i18n.Tr(o.label)
		}
		// show is this frame's local: Changed applies the choice to it now.
		// Without the query the choice lands after the frame, in a variable
		// nothing reads again.
		if uixselect.Select(c, uixselect.Props{Selected: &show, Options: labels}).Changed() {
			d.set(sw.s, enumValue(d, show))
		}
	case "number":
		numberinput.NumberInput(c, numberinput.Props{
			Value: d.ptr(sw.s).(*float64), Lo: d.lo, Hi: d.hi, Step: d.step,
		})
	case "string":
		input.Input(c, input.Props{Value: d.ptr(sw.s).(*string)})
	case "list":
		sw.renderList(c, d)
	case "color":
		sw.colorControls(c, d)
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
			On: d.ptr(sw.s).(*bool), Label: i18n.Tr(d.label),
		})
		if d.desc != "" {
			sw.hint(c, i18n.Tr(d.desc))
		}
	case "enum":
		field.Field(c, field.Props{Label: i18n.Tr(d.label), Hint: i18n.Tr(d.desc)}, func() {
			labels := make([]string, len(d.enum))
			for i, o := range d.enum {
				labels[i] = i18n.Tr(o.label)
			}
			show := enumLabel(d, d.get(sw.s).(string))
			if uixselect.Select(c, uixselect.Props{Selected: &show, Options: labels}).Changed() {
				d.set(sw.s, enumValue(d, show))
			}
		})
	case "number":
		field.Field(c, field.Props{Label: i18n.Tr(d.label), Hint: i18n.Tr(d.desc)}, func() {
			numberinput.NumberInput(c, numberinput.Props{
				Value: d.ptr(sw.s).(*float64), Lo: d.lo, Hi: d.hi, Step: d.step,
			})
		})
	case "string":
		field.Field(c, field.Props{Label: i18n.Tr(d.label), Hint: i18n.Tr(d.desc)}, func() {
			input.Input(c, input.Props{Value: d.ptr(sw.s).(*string)})
		})
	case "list":
		field.Field(c, field.Props{Label: i18n.Tr(d.label), Hint: i18n.Tr(d.desc)}, func() {
			sw.renderList(c, d)
		})
	case "color":
		sw.renderColorRow(c, d)
	}
}

// enumLabel is the select of an enum setting: it shows the option
// labels in the interface language but stores the option values, so a
// value with a friendlier label (like "Off (typing accents)" for
// false) still saves the key.
func enumLabel(d settingDef, v string) string {
	for _, o := range d.enum {
		if o.value == v {
			return i18n.Tr(o.label)
		}
	}
	return v
}

// enumValue maps a shown label back to its stored value. It accepts the
// label in either language, so a select toggled mid-transition still
// round-trips.
func enumValue(d settingDef, label string) string {
	for _, o := range d.enum {
		if o.label == label || i18n.Tr(o.label) == label {
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

// renderColorRow is one named color setting of the All Settings list:
// the well, and a Reset that forgets the setting and returns the well to
// the theme's color.
func (sw *settingsWindow) renderColorRow(c *ui.Context, d settingDef) {
	field.Field(c, field.Props{Label: i18n.Tr(d.label), Hint: i18n.Tr(d.desc)}, func() {
		sw.colorControls(c, d)
	})
}

// colorControls is one color setting's control: the well, and a Reset
// that forgets the setting and returns the well to the theme's color.
func (sw *settingsWindow) colorControls(c *ui.Context, d settingDef) {
	t := c.Theme()
	st := sw.wellState(d)
	ui.Row(c).Gap(12).Children(func() {
		colorwell.ColorWell(c, colorwell.Props{Color: &st.color, Label: i18n.Tr(d.label)})
		if st.inUse {
			rb := ui.ButtonBase(c.Key("reset-"+d.key)).Padding(2, 8).Radius(t.Radius).
				FocusRing(false).Tooltip(i18n.Tr("Use the theme's color"))
			rb.Children(func() {
				ui.Row(c).Gap(4).Children(func() {
					ui.Image(c, lucide.RotateCcw()).Size(12, 12)
					ui.Text(c, i18n.Tr("Reset")).FontSize(12).TextColor(t.TextMuted)
				})
			})
			if rb.Clicked() {
				d.set(sw.s, "")
				st.inUse = false
			}
		}
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

// moveButton is a small arrow that calls onClick, dimmed when not enabled.
func (sw *settingsWindow) moveButton(c *ui.Context, key string, icon *ui.SVG, enabled bool, onClick func()) {
	t := c.Theme()
	b := ui.ButtonBase(c.Key(key)).Padding(4, 4).Radius(t.Radius).FocusRing(false)
	if !enabled {
		b.Disabled(true)
	}
	b.Children(func() { ui.Image(c, icon).Size(14, 14) })
	if b.Clicked() {
		onClick()
	}
}

// paletteGrid is the sixteen palette wells in a grid, and a reset that
// returns them all to the theme's colors. The item's row supplies the
// "Palette" title.
func (sw *settingsWindow) paletteGrid(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).FillWidth().Gap(8).Children(func() {
		for row := 0; row < 4; row++ {
			ui.Row(c).Gap(10).Children(func() {
				for col := 0; col < 4; col++ {
					i := row*4 + col
					d := *defByKey(fmt.Sprintf("palette-%d", i))
					st := sw.wellState(d)
					colorwell.ColorWell(c, colorwell.Props{Color: &st.color, Label: i18n.Tr(d.label)})
				}
			})
		}
		if sw.s.AnyPalette() {
			rb := ui.ButtonBase(c.Key("reset-palette")).Padding(2, 8).Radius(t.Radius).
				FocusRing(false).Tooltip(i18n.Tr("Use the theme's colors"))
			rb.Children(func() {
				ui.Row(c).Gap(4).Children(func() {
					ui.Image(c, lucide.RotateCcw()).Size(12, 12)
					ui.Text(c, i18n.Tr("Reset palette")).FontSize(12).TextColor(t.TextMuted)
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
// that means following the window. The snapshot it compares is the
// settings' JSON: a copy of the struct shares its slices and maps with the
// settings, so an edit made in place (a profile's field, a details view, a
// key binding, an Extra key) never differed from the copy, and was not saved.
func (sw *settingsWindow) sync() {
	now, err := json.Marshal(sw.s)
	if err != nil {
		return
	}
	if sw.lastJSON == nil {
		sw.lastJSON = now // the first frame: what the window opened with
		return
	}
	if !bytes.Equal(now, sw.lastJSON) {
		// A change of language redraws the menu with translated labels.
		if sw.last.Language != sw.s.Language {
			i18n.SetLang(sw.s.Language)
			sw.app.ApplyKeybinds()
			// The window's title is native, set once as the window was
			// made: it does not redraw with the view, so retitle it.
			if sw.win != nil {
				sw.win.SetTitle(i18n.Tr("Settings"))
			}
		}
		sw.last = *sw.s
		sw.lastJSON = now
		sw.app.ApplySettings()
		sw.app.SaveSettings()
	}
}

// syncColors writes the color wells back into the Settings: an unused
// well follows the theme's color, editing one makes the setting, and a
// change of theme refreshes the unused ones.
func (sw *settingsWindow) syncColors() {
	th := ThemeFor(sw.s.ThemeName, false)
	if th == nil {
		th = OttyTermTheme(false)
	}
	// The theme's cursor and selection are zero in the window colors; the
	// terminal falls back to the foreground for the cursor and an accent
	// for the selection, so the wells show what would render. The text
	// under the cursor falls back to the background the same way.
	cursorDef := th.Cursor
	if cursorDef.A == 0 {
		cursorDef = th.Foreground
	}
	cursorTextDef := th.CursorText
	if cursorTextDef.A == 0 {
		cursorTextDef = th.Background
	}
	selDef := th.Selection
	if selDef.A == 0 {
		selDef = ui.Hex("#A8C7FA")
	}
	selTextDef := th.SelectionText
	if selTextDef.A == 0 {
		selTextDef = th.Foreground // selected text keeps its own colors
	}
	defColor := func(d settingDef) ui.Color {
		switch d.key {
		case "background":
			return th.Background
		case "cursor-color":
			return cursorDef
		case "cursor-text":
			return cursorTextDef
		case "selection-background":
			return selDef
		case "selection-foreground":
			return selTextDef
		}
		var i int
		if n, _ := fmt.Sscanf(d.key, "palette-%d", &i); n == 1 {
			return th.Palette[i]
		}
		return th.Foreground
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
		// An unused well follows the theme, so a change of theme refreshes
		// it. It is refreshed only then: the picker writes its color into
		// the well between frames, and resetting the well on every frame
		// erased the pick before the comparison below could see it.
		if !st.inUse && st.base != def {
			st.base, st.color = def, def
		}
		if st.color != st.base {
			st.inUse = true
			st.base = st.color
			set(ColorHex(st.color))
		} else if st.inUse {
			set(ColorHex(st.color))
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
