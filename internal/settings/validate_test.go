package settings

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

// TestNormalizeValidates checks what a hand-edited file can hold and no
// control can make: numbers outside their range come back to it, an enum
// value nothing offers goes to the default, the details views are each
// there once, and the default profile names a profile.
func TestNormalizeValidates(t *testing.T) {
	s := LoadFrom([]byte(`{
		"WindowOpacity": 0,
		"WindowWidth": -5,
		"SidebarWidth": 100000,
		"DetailsPanelWidth": 1,
		"FontSize": 0,
		"ScrollbackLines": 1e15,
		"WindowLayout": "upside-down",
		"CursorStyle": "",
		"ThemeName": "GitHub Dark",
		"BoldColor": "#FF0000",
		"DetailsViews": [
			{"ID": "git", "Kind": "builtin", "Title": "Git", "Enabled": false},
			{"ID": "git", "Kind": "builtin", "Title": "Git", "Enabled": true},
			{"ID": "nope", "Kind": "builtin", "Title": "?", "Enabled": true},
			{"ID": "docs", "Kind": "web", "Title": "Docs", "URL": "https://example.com", "Enabled": true},
			{"ID": "", "Kind": "web", "Enabled": true}
		],
		"Profiles": [{"Name": "Web"}],
		"DefaultProfile": "Gone"
	}`))
	if s == nil {
		t.Fatal("the file did not load")
	}
	for _, c := range []struct {
		name      string
		got, want any
	}{
		{"WindowOpacity", s.WindowOpacity, 0.5},
		{"WindowWidth", s.WindowWidth, 680.0},
		{"SidebarWidth", s.SidebarWidth, 420.0},
		{"DetailsPanelWidth", s.DetailsPanelWidth, 160.0},
		{"FontSize", s.FontSize, 6.0},
		{"ScrollbackLines", s.ScrollbackLines, 1000000.0},
		{"WindowLayout", s.WindowLayout, "sidebar-left"},
		{"CursorStyle", s.CursorStyle, "block"},
		// A theme is any Ghostty theme, and bold text may have a color:
		// their selects offer only some of the values they take.
		{"ThemeName", s.ThemeName, "GitHub Dark"},
		{"BoldColor", s.BoldColor, "#FF0000"},
		{"DefaultProfile", s.DefaultProfile, ""},
	} {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
	var ids []string
	for _, v := range s.DetailsViews {
		ids = append(ids, v.ID)
	}
	// The file's views keep their order, the first of an id wins, and the
	// built-in ones it lacks follow.
	if want := []string{"git", "docs", "info", "outline", "files"}; !reflect.DeepEqual(ids, want) {
		t.Errorf("details views = %v, want %v", ids, want)
	}
	if s.DetailsViews[0].Enabled {
		t.Error("the first git view was replaced by its duplicate")
	}
	// A default that names a profile stays.
	if s := LoadFrom([]byte(`{"Profiles": [{"Name": "Web"}], "DefaultProfile": "Web"}`)); s.DefaultProfile != "Web" {
		t.Errorf("DefaultProfile = %q, want Web", s.DefaultProfile)
	}
}

// TestDefaultsAreValid checks that the defaults pass their own validation:
// Normalize leaves a fresh install's settings as they are, apart from the
// maps it makes.
func TestDefaultsAreValid(t *testing.T) {
	s := Default()
	checkValid(t, s)
	before, _ := json.Marshal(s)
	s.validate()
	if after, _ := json.Marshal(s); !bytes.Equal(before, after) {
		t.Errorf("validate changed the defaults:\n%s\n%s", before, after)
	}
}

// checkValid fails for settings a control could not have made: a number
// outside its range, an enum value its select does not offer, a def whose
// value does not read back, a details view without an id of its own, a
// default profile that is not there.
func checkValid(t *testing.T, s *Settings) {
	t.Helper()
	for _, d := range settingsDefs {
		v := d.get(s)
		switch d.kind {
		case "number":
			if n, ok := v.(float64); !ok || !(n >= d.lo && n <= d.hi) {
				t.Errorf("%s = %v, outside %v..%v", d.key, v, d.lo, d.hi)
			}
		case "enum":
			if e, ok := v.(string); !ok || !d.open && !d.offers(e) {
				t.Errorf("%s = %q, not one of its options", d.key, v)
			}
		}
		d.set(s, v)
		if !reflect.DeepEqual(d.get(s), v) {
			t.Errorf("%s: set(%v) reads back as %v", d.key, v, d.get(s))
		}
	}
	seen := map[string]bool{}
	for _, v := range s.DetailsViews {
		if v.ID == "" || seen[v.ID] {
			t.Errorf("details view %q is empty or there twice", v.ID)
		}
		seen[v.ID] = true
	}
	for _, id := range []string{"info", "outline", "git", "files"} {
		if !seen[id] {
			t.Errorf("no %s details view", id)
		}
	}
	if s.DefaultProfile != "" && s.Profile() == nil {
		t.Errorf("the default profile %q is no profile", s.DefaultProfile)
	}
}

// FuzzLoadFrom checks that whatever a settings file holds, it either does
// not load or loads as settings the pages could have made (see checkValid)
// that are stable: a second Normalize changes nothing, and saved and read
// again they are the same settings.
func FuzzLoadFrom(f *testing.F) {
	def, _ := json.Marshal(Default())
	f.Add(def)
	for _, seed := range []string{
		`{}`, `null`, `[]`, `{"FontSize": "18"}`, `{"FontSize": 1e400}`,
		`{"WindowOpacity": 0, "WindowWidth": -1, "ScrollbackLines": 1e15, "FontSize": 1e-9}`,
		`{"FontLigatures": true, "OptionAsAlt": false, "ConfirmCloseTab": true, "CursorBlink": true}`,
		`{"WindowLayout": "x", "Language": "", "CursorBlinkStyle": "maybe", "ThemeName": "x"}`,
		`{"DetailsViews": [{"ID": "git"}, {"ID": "git"}, {"ID": "x", "Kind": "web"}, {}]}`,
		`{"Profiles": [{"Name": "a", "Env": ["A=1"]}, {"Name": "a"}], "DefaultProfile": "b"}`,
		`{"Palette": ["#000", "red"], "Foreground": "#12345", "Extra": {"a": [1, {"b": null}], "c": 1e21}}`,
		`{"KeybindOverrides": {"cmd+k": "new_tab"}, "Env": null, "fontsize": 7, "FontSize": 8}`,
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		s := LoadFrom(b)
		if s == nil {
			return
		}
		checkValid(t, s)
		before, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("the loaded settings do not save: %v", err)
		}
		s.Normalize()
		if after, _ := json.Marshal(s); !bytes.Equal(before, after) {
			t.Errorf("a second Normalize changed the settings:\n%s\n%s", before, after)
		}
		back := LoadFrom(before)
		if back == nil {
			t.Fatalf("the saved settings do not load:\n%s", before)
		}
		if again, _ := json.Marshal(back); !bytes.Equal(before, again) {
			t.Errorf("saved and read again, the settings differ:\n%s\n%s", before, again)
		}
	})
}
