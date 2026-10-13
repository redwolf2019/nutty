package settings

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestLoadKeepsTheRestOfAFile checks that a value of the wrong type costs
// its own setting and no other: the setting keeps its default, and the
// rest of the file is read. One such value used to fail the whole file,
// and every setting in it was replaced by the defaults.
func TestLoadKeepsTheRestOfAFile(t *testing.T) {
	s := LoadFrom([]byte(`{
		"FontSize": "18",
		"WindowOpacity": true,
		"Env": "A=1",
		"Profiles": [{"Name": "Web"}, {"Name": 5}],
		"DetailsViews": {"ID": "git"},
		"Palette": [1, 2],
		"FontFamily": null,
		"ThemeName": "Dracula",
		"ScrollbackLines": 500,
		"KeybindOverrides": {"cmd+shift+k": "new_tab"},
		"CommandWhitelist": ["vim"]
	}`))
	if s == nil {
		t.Fatal("a file with values of the wrong type did not load at all")
	}
	def := Default()
	// The values that do not fit keep the defaults.
	if s.FontSize != def.FontSize || s.WindowOpacity != def.WindowOpacity || s.FontFamily != def.FontFamily ||
		len(s.Env) != 0 || len(s.Profiles) != 0 || len(s.DetailsViews) != len(def.DetailsViews) || s.AnyPalette() {
		t.Errorf("a value of the wrong type changed its setting: FontSize %v, WindowOpacity %v, FontFamily %q, Env %v, Profiles %v, DetailsViews %v, Palette %v",
			s.FontSize, s.WindowOpacity, s.FontFamily, s.Env, s.Profiles, s.DetailsViews, s.Palette)
	}
	// The rest of the file is read.
	if s.ThemeName != "Dracula" || s.ScrollbackLines != 500 ||
		s.KeybindOverrides["cmd+shift+k"] != "new_tab" || !reflect.DeepEqual(s.CommandWhitelist, []string{"vim"}) {
		t.Errorf("the rest of the file was not read: ThemeName %q, ScrollbackLines %v, KeybindOverrides %v, CommandWhitelist %v",
			s.ThemeName, s.ScrollbackLines, s.KeybindOverrides, s.CommandWhitelist)
	}
}

// TestLoadFailsOnWhatIsNoObject checks the one failure left: a file that
// is not a JSON object at all.
func TestLoadFailsOnWhatIsNoObject(t *testing.T) {
	for _, file := range []string{``, `[]`, `"FontSize"`, `18`, `{"FontSize": 18,}`, `{"FontSize": 18`} {
		if LoadFrom([]byte(file)) != nil {
			t.Errorf("%q loaded as settings", file)
		}
	}
}

// TestUnmarshalOverlays checks that reading a file into settings that hold
// values replaces the fields it names, whole, and leaves the others.
func TestUnmarshalOverlays(t *testing.T) {
	s := Default()
	s.FontSize = 15
	s.Env = []string{"A=1", "B=2"}
	s.KeybindOverrides = map[string]string{"cmd+k": "clear_screen"}
	if err := json.Unmarshal([]byte(`{"Env": ["C=3"], "KeybindOverrides": {"cmd+j": "jump_to"}, "fontfamily": "Menlo", "unknown": 1}`), s); err != nil {
		t.Fatal(err)
	}
	if s.FontSize != 15 || s.FontFamily != "Menlo" {
		t.Errorf("FontSize %v, FontFamily %q, want 15 and Menlo", s.FontSize, s.FontFamily)
	}
	if !reflect.DeepEqual(s.Env, []string{"C=3"}) || !reflect.DeepEqual(s.KeybindOverrides, map[string]string{"cmd+j": "jump_to"}) {
		t.Errorf("a named list or map was merged, not replaced: Env %v, KeybindOverrides %v", s.Env, s.KeybindOverrides)
	}
}
