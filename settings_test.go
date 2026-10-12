package main

import (
	"bytes"
	"encoding/json"
	"image"
	"reflect"
	"testing"
	"time"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// TestSettingsJSON checks that the settings survive a save and load, and
// that a file with unknown fields still reads.
func TestSettingsJSON(t *testing.T) {
	s := defaultSettings()
	s.ThemeName = "TokyoNight"
	s.FontSize = 15
	s.WindowOpacity = 0.8
	s.Foreground = "#123456"
	s.ConfirmCloseTab = "always"
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var back settings
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(s, &back) {
		t.Fatalf("round trip changed settings:\n%+v\n%+v", s, &back)
	}
	// A hand-edited file, with an unknown key, still loads.
	if err := json.Unmarshal([]byte(`{"FontSize":14,"futureSetting":true}`), &back); err != nil {
		t.Fatal(err)
	}
	if back.FontSize != 14 || back.FontFamily != "JetBrains Mono" {
		t.Fatalf("merged wrong: %+v", back)
	}
}

// TestSettingsApply checks the pieces the settings turn into terminal
// options: the font, the cursor, the scrollback.
func TestSettingsApply(t *testing.T) {
	a := &app{settings: defaultSettings()}

	f := a.termFont()
	if f.Family != "JetBrains Mono" || f.Size != 13 {
		t.Fatalf("default font = %+v", f)
	}
	if len(f.Features) != 1 || f.Features[0] != "dlig" {
		t.Fatalf("default ligatures = %v, want [dlig]", f.Features)
	}
	a.settings.FontLigatures = "off"
	if f := a.termFont(); len(f.Features) != 3 || f.Features[0] != "-calt" {
		t.Fatalf("ligatures off = %v", f.Features)
	}
	a.settings.FontLigatures = "dlig"
	a.settings.FontSize = 16
	a.settings.FontThicken = true
	f = a.termFont()
	if f.Size != 16 || !f.Thicken {
		t.Fatalf("font options = %+v", f)
	}

	if a.cursorStyle() != terminal.CursorBlock {
		t.Fatal("default cursor should be block")
	}
	a.settings.CursorStyle = "bar"
	if a.cursorStyle() != terminal.CursorBar {
		t.Fatal("bar cursor")
	}

	if n := a.scrollbackBytes(); n != 1_000_000 { // 10000 lines
		t.Fatalf("scrollback bytes = %d, want 1,000,000", n)
	}
	a.settings.ScrollbackLines = 0
	if a.scrollbackBytes() != -1 {
		t.Fatal("zero lines should turn the scrollback off")
	}
}

// TestSettingsNormalize checks that a settings file written by an older
// build — with booleans where the enums live now — still reads as the
// right enum, and that the details views have their defaults.
func TestSettingsNormalize(t *testing.T) {
	var s settings
	if err := json.Unmarshal([]byte(`{
		"FontLigatures": true,
		"OptionAsAlt": false,
		"ConfirmCloseTab": true,
		"CursorBlink": true
	}`), &s); err != nil {
		t.Fatal(err)
	}
	s.normalize()
	if s.FontLigatures != "dlig" {
		t.Fatalf("FontLigatures = %q, want dlig", s.FontLigatures)
	}
	if s.OptionAsAlt != "false" {
		t.Fatalf("OptionAsAlt = %q, want false", s.OptionAsAlt)
	}
	if s.ConfirmCloseTab != "process" {
		t.Fatalf("ConfirmCloseTab = %q, want process", s.ConfirmCloseTab)
	}
	if len(s.DetailsViews) != 4 {
		t.Fatalf("details views = %d, want 4", len(s.DetailsViews))
	}
}

// TestThemeNames checks the theme lists the settings select from.
func TestThemeNames(t *testing.T) {
	names := themeNames()
	if len(names) != 1+len(ottyThemeNames) {
		t.Fatalf("theme names = %d, want %d", len(names), 1+len(ottyThemeNames))
	}
	if names[0] != "Follow system" {
		t.Fatalf("first theme = %q, want Follow system", names[0])
	}
}

// TestSettingsRegistry checks the settings registry: every modeled key's
// get and set round-trip, every page group references a key the registry
// holds, and the enum values all have a label.
func TestSettingsRegistry(t *testing.T) {
	s := defaultSettings()
	for _, d := range settingsDefs {
		orig := d.get(s)
		switch d.kind {
		case "enum":
			if enumLabel(d, orig.(string)) == "" {
				t.Errorf("%s: value %q has no label", d.key, orig)
			}
			for _, o := range d.enum {
				if enumValue(d, o.label) != o.value {
					t.Errorf("%s: label %q does not round-trip to %q", d.key, o.label, o.value)
				}
			}
		case "color", "string":
			if _, ok := orig.(string); !ok {
				t.Errorf("%s: %s get = %T, want string", d.key, d.kind, orig)
			}
		case "number":
			if _, ok := orig.(float64); !ok {
				t.Errorf("%s: number get = %T, want float64", d.key, orig)
			}
		case "bool":
			if _, ok := orig.(bool); !ok {
				t.Errorf("%s: bool get = %T, want bool", d.key, orig)
			}
		}
		// A set of a different value reads back as that value.
		var v any
		switch d.kind {
		case "bool":
			v = !orig.(bool)
		case "string", "color":
			v = orig.(string) + "x"
		case "number":
			v = orig.(float64) + 1
		case "list":
			v = append(append([]string{}, orig.([]string)...), "x")
		case "enum":
			v = d.enum[0].value
		}
		if reflect.DeepEqual(v, orig) {
			continue
		}
		d.set(s, v)
		if !reflect.DeepEqual(d.get(s), v) {
			t.Errorf("%s: set(%v) then get = %v", d.key, v, d.get(s))
		}
		d.set(s, orig)
	}
	for page, groups := range settingsGroups {
		for _, g := range groups {
			for _, key := range g.keys {
				if defByKey(key) == nil {
					t.Errorf("page %s group %q: no def for %q", page, g.title, key)
				}
			}
		}
	}
}

// TestKeybindRows checks the key binding list: every default binding is
// there with its trigger, and a recorded override wins over the default.
func TestKeybindRows(t *testing.T) {
	s := defaultSettings()
	sw := &settingsWindow{s: s}
	rows := sw.keybindRows()
	byAction := map[string]keybindRow{}
	for _, r := range rows {
		byAction[r.action] = r
	}
	if r, ok := byAction["goto_tab:1"]; !ok || r.trigger != "cmd+1" {
		t.Fatalf("goto_tab:1 row = %+v, want cmd+1", r)
	}
	if r, ok := byAction["new_tab"]; !ok || r.trigger != "cmd+t" {
		t.Fatalf("new_tab row = %+v, want cmd+t", r)
	}
	// An override replaces the trigger and marks the row.
	s.KeybindOverrides = map[string]string{"cmd+alt+t": "new_tab"}
	sw = &settingsWindow{s: s}
	for _, r := range sw.keybindRows() {
		if r.action == "new_tab" {
			if r.trigger != "cmd+alt+t" || !r.overridden {
				t.Fatalf("overridden new_tab = %+v", r)
			}
		}
	}
}

// TestKeybindRecord checks that recording a combo binds it, drops the
// action's old binding and any other action on the combo, and that
// conflicts are detected.
func TestKeybindRecord(t *testing.T) {
	s := defaultSettings()
	s.KeybindOverrides = map[string]string{"cmd+d": "new_tab"}
	sw := &settingsWindow{s: s}
	sw.keys.recording = "split_right"
	sw.recordCombo("cmd+d")
	if got := s.KeybindOverrides["cmd+d"]; got != "split_right" {
		t.Fatalf("cmd+d = %q, want split_right", got)
	}
	if tr, ok := s.actionTrigger("new_tab"); ok {
		t.Fatalf("new_tab still bound to %q, want unbound", tr)
	}
	if a, _ := s.triggerAction("cmd+shift+d"); a != "split_down" {
		t.Fatalf("default split_down = %q, want split_down", a)
	}
}

// TestAccelerator checks that otty triggers convert to menu accelerators.
func TestAccelerator(t *testing.T) {
	cases := map[string]string{
		"cmd+comma":             "CmdOrCtrl+,",
		"cmd+shift+l":           "CmdOrCtrl+Shift+L",
		"cmd+opt+d":             "CmdOrCtrl+Option+D",
		"cmd+shift+right_bracket": "CmdOrCtrl+Shift+]",
		"cmd+ctrl+down":         "CmdOrCtrl+Ctrl+Down",
		"cmd+equal":             "CmdOrCtrl+=",
		"cmd+0":                 "CmdOrCtrl+0",
		"cmd+t":                 "CmdOrCtrl+T",
	}
	for trigger, want := range cases {
		if got := acceleratorOf(trigger); got != want {
			t.Errorf("acceleratorOf(%q) = %q, want %q", trigger, got, want)
		}
	}
}

// TestShellArgs checks that a profile's command line splits into
// arguments with its quotes honored.
func TestShellArgs(t *testing.T) {
	cases := []struct{ in string; want int }{
		{"", 0},
		{"zsh", 1},
		{"docker compose up", 3},
		{"sh -c 'echo hi'", 3},
		{`bash -c "ls -la"`, 3},
	}
	for _, tc := range cases {
		if got := len(shellArgs(tc.in)); got != tc.want {
			t.Errorf("shellArgs(%q) = %d args, want %d", tc.in, got, tc.want)
		}
	}
}

// TestProfile checks that the default profile is found by name and that
// newTerm's profile overrides apply.
func TestProfile(t *testing.T) {
	s := defaultSettings()
	if s.profile() != nil {
		t.Fatal("no default profile should be chosen")
	}
	s.Profiles = []profile{{Name: "Web", Command: "docker compose up", CWD: "/srv"}}
	s.DefaultProfile = "Web"
	pr := s.profile()
	if pr == nil || pr.Name != "Web" {
		t.Fatalf("profile() = %+v, want Web", pr)
	}
	// A default that points nowhere is ignored.
	s.DefaultProfile = "Gone"
	if s.profile() != nil {
		t.Fatal("a missing default profile should be ignored")
	}
}

// TestSettingsPagesRender checks that every category draws its page, with
// the settings the page is known for. Long pages render taller than the
// window, so each one is drawn in a viewport tall enough for its content.
func TestSettingsPagesRender(t *testing.T) {
	a := &app{settings: defaultSettings(), sideOpen: true, agent: &demoAgent{}}
	a.settings.Profiles = []profile{{Name: "Web", Command: "docker compose up"}}
	want := map[string][]string{
		"appearance":  {"Theme", "Font Family", "Cursor Style", "Palette"},
		"general":     {"Language", "Closing Window", "On Launch"},
		"shell":       {"Command", "Working Directory", "Environment Variables"},
		"controls":    {"Use Option Key as Meta", "Right-Click Action", "Scrollback Lines"},
		"editor":      {"Soft Wrap", "Vim Keys"},
		"view":        {"Details Panel", "Outline"},
		"recipes":     {"Saved Recipes"},
		"agents":      {"Prevent Sleep While Processing"},
		"data":        {"Not available in this build"},
		"account":     {"Not signed in"},
		"remote":      {"Not available in this build."},
		"key_bindings": {"New Tab", "cmd+t", "Tab 1", "cmd+1"},
		"advanced":     {"Config File", "All Settings", "font-family", "auto-update"},
		"profiles":     {"New Profile", "Web", "Command", "Set as Default"},
	}
	for _, cat := range settingsCategories {
		sw := &settingsWindow{
			a: a, s: a.settings, last: *a.settings,
			state: components.SettingsState{Category: cat.id},
		}
		if cat.id == "profiles" {
			sw.profileSel = 0
		}
		tt := ui.NewTester(sw.view, 760, 1600)
		tt.Frame()
		for _, w := range want[cat.id] {
			if !tt.HasText(w) {
				t.Errorf("%s: missing %q", cat.id, w)
			}
		}
	}
}

// TestSettingsCategoryIcons checks that every settings category has an
// icon and that the otty paths parse, so a bad extraction or a mygo SVG
// limitation shows as a test failure, not a blank row.
func TestSettingsCategoryIcons(t *testing.T) {
	for _, cat := range settingsCategories {
		if icon := settingsCategoryIcon(cat.id); icon == nil {
			t.Errorf("%s: no icon", cat.id)
		}
	}
}

// TestDataBreath checks that a click of the Data row breathes the database
// icon, and that the breath comes back to rest: at the puff the icon's
// pixels differ from the rested icon's, and the rested icon matches another
// rested one. The pose spring in the headless tester shows a frame behind
// the keyframes (it advances by the time between frames), so the capture
// sequence is: a frame sets the puff target, the next renders it, and two
// more frames with a gap settle the icon. The selection the click also
// causes is the same in every frame compared here, so only the puff shows.
func TestDataBreath(t *testing.T) {
	a := &app{settings: defaultSettings()}
	sw := &settingsWindow{a: a, s: a.settings, last: *a.settings,
		state: components.SettingsState{Category: "appearance"}}
	tt := ui.NewTester(sw.view, 760, 800)
	tt.Frame()
	if err := tt.Click("Data"); err != nil {
		t.Fatal(err)
	}
	// The click starts the breath; this frame's keyframe is the puff's peak,
	// which the spring shows on the next frame.
	time.Sleep(350 * time.Millisecond)
	tt.Frame()
	// The spring catches the peak across the gap: the icon puffs here.
	time.Sleep(400 * time.Millisecond)
	tt.Frame()
	puff := breathCrop(t, tt)
	// The breath (0.983 s) is over; this frame tells the spring to rest...
	time.Sleep(600 * time.Millisecond)
	tt.Frame()
	// ...and with a gap the spring settles there.
	time.Sleep(700 * time.Millisecond)
	tt.Frame()
	rest1 := breathCrop(t, tt)
	// A rested icon later is the same again.
	time.Sleep(1200 * time.Millisecond)
	tt.Frame()
	rest1b := breathCrop(t, tt)
	if bytes.Equal(puff, rest1) {
		t.Fatal("the icon does not change through the breath")
	}
	if !bytes.Equal(rest1, rest1b) {
		t.Fatal("the icon does not return to rest after the breath")
	}
}

// breathCrop is the pixels of the settings sidebar around the Data row's
// icon: the icon's square, the only thing that moves as it breathes. The
// row's background and label are static across the frames a breath
// compares.
func breathCrop(t *testing.T, tt *ui.Tester) []byte {
	t.Helper()
	r, ok := tt.Find("Data")
	if !ok {
		t.Fatal("no Data row")
	}
	// Find returns the row's button, which spans the sidebar; the icon
	// sits just right of its left padding. Crop that square, ending where
	// the label starts.
	crop := image.Rect(int(r.X), int(r.Y)-6, int(r.X)+34, int(r.Y)+int(r.H)+6)
	img := tt.Image()
	var b []byte
	for y := max(crop.Min.Y, 0); y < crop.Max.Y && y < img.Bounds().Dy(); y++ {
		for x := max(crop.Min.X, 0); x < crop.Max.X && x < img.Bounds().Dx(); x++ {
			i := img.PixOffset(x, y)
			b = append(b, img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3])
		}
	}
	return b
}
