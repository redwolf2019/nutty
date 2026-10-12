package main

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
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

// TestSettingsPagesRender checks that every category draws its page, with
// the settings the page is known for. Long pages render taller than the
// window, so each one is drawn in a viewport tall enough for its content.
func TestSettingsPagesRender(t *testing.T) {
	a := &app{settings: defaultSettings(), sideOpen: true, agent: &demoAgent{}}
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
	}
	for _, cat := range settingsCategories {
		sw := &settingsWindow{a: a, s: a.settings, last: *a.settings, page: cat.id}
		tt := ui.NewTester(sw.view, 760, 1600)
		tt.Frame()
		for _, w := range want[cat.id] {
			if !tt.HasText(w) {
				t.Errorf("%s: missing %q", cat.id, w)
			}
		}
	}
}
