package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
	"github.com/smallnest/uix/components/colorwell"
	"github.com/smallnest/uix/components/field"
	"github.com/smallnest/uix/components/input"
	"github.com/smallnest/uix/components/numberinput"
	"github.com/smallnest/uix/components/searchfield"
	"github.com/smallnest/uix/components/textarea"
	uixswitch "github.com/smallnest/uix/components/switch"
)

// pageAdvanced is the Advanced category: the config file, import and
// export, and the All Settings list, every key of otty's config with an
// inline editor.
func (sw *settingsWindow) pageAdvanced(c *ui.Context) {
	sw.group(c, "Config File", func() {
		t := c.Theme()
		ui.Text(c, settingsPath()).FontSize(11).TextColor(t.TextMuted).SingleLine()
		ui.Row(c).Gap(8).Children(func() {
			sw.advButton(c, "adv-open-config", "Open", func() {
				exec.Command("open", settingsPath()).Start()
			})
			sw.advButton(c, "adv-reload-config", "Reload", func() {
				*sw.s = *loadSettings()
				sw.wells = colorWells{}
				sw.lists = nil
				sw.a.applySettings()
			})
		})
	})
	sw.group(c, "Import / Export", func() {
		ui.Row(c).Gap(8).Children(func() {
			sw.advButton(c, "adv-export", "Export…", func() {
				path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
					Parent: sw.win, Title: "Export Settings",
					DefaultPath: "nutty-settings.json",
					Filters:     []mygo.FileFilter{{Name: "JSON", Extensions: []string{"json"}}},
				})
				if err != nil || path == "" {
					return
				}
				if b, err := json.MarshalIndent(sw.s, "", "  "); err == nil {
					os.WriteFile(path, b, 0o644)
				}
			})
			sw.advButton(c, "adv-import", "Import…", func() {
				paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
					Parent: sw.win, Title: "Import Settings",
					Filters: []mygo.FileFilter{{Name: "JSON", Extensions: []string{"json"}}},
				})
				if err != nil || len(paths) == 0 {
					return
				}
				if b, err := os.ReadFile(paths[0]); err == nil {
					if loaded := loadSettingsFrom(b); loaded != nil {
						*sw.s = *loaded
						sw.wells = colorWells{}
						sw.lists = nil
						sw.a.applySettings()
						sw.a.saveSettings()
					}
				}
			})
		})
		sw.hint(c, "Import replaces the current settings with the file's.")
	})
	sw.group(c, "Debug", func() {
		uixswitch.Switch(c, uixswitch.Props{
			On: &sw.s.DebugMode, Label: "Debug mode",
		})
		sw.hint(c, "Writes extra logs to the console and shows the developer tools.")
	})
	sw.group(c, "All Settings", func() {
		searchfield.SearchField(c, searchfield.Props{Value: &sw.settingsSearch})
		sw.allSettingsList(c)
	})
	sw.group(c, "Reset", func() {
		sw.advButton(c, "adv-reset", "Reset All Settings", func() {
			*sw.s = *defaultSettings()
			sw.wells = colorWells{}
			sw.lists = nil
			sw.a.applySettings()
			sw.a.saveSettings()
			sw.a.applyKeybinds()
		})
		sw.hint(c, "Returns every setting to otty's defaults.")
	})
}

// advButton is a small labeled button of the Advanced page.
func (sw *settingsWindow) advButton(c *ui.Context, key, label string, onClick func()) {
	t := c.Theme()
	b := ui.ButtonBase(c.Key(key)).Padding(4, 12).Radius(t.Radius).FocusRing(false)
	b.Children(func() {
		ui.Text(c, label).FontSize(12).TextColor(t.Text)
	})
	if b.Clicked() {
		onClick()
	}
}

// allSettingsList is every modeled setting and every unmodeled key, with
// an inline editor, filtered by the search.
func (sw *settingsWindow) allSettingsList(c *ui.Context) {
	q := strings.ToLower(strings.TrimSpace(sw.settingsSearch))
	shown := 0
	for _, d := range settingsDefs {
		if q != "" && !strings.Contains(strings.ToLower(d.key), q) {
			continue
		}
		shown++
		// The list shows the config key as the row's name, as otty's does.
		d2 := d
		d2.label = d.key
		d2.desc = ""
		sw.renderSetting(c, d2)
	}
	for _, x := range extraSettings {
		if q != "" && !strings.Contains(strings.ToLower(x.key), q) {
			continue
		}
		shown++
		sw.renderExtra(c, x)
	}
	if shown == 0 {
		sw.hint(c, "No matching settings")
	}
}

// renderExtra draws one unmodeled key's inline editor, reading and
// writing it through settings.Extra.
func (sw *settingsWindow) renderExtra(c *ui.Context, x extraDef) {
	t := c.Theme()
	key := "extra:" + x.key
	switch x.kind {
	case "bool":
		on := extraBool(sw.s, x.key)
		uixswitch.Switch(c, uixswitch.Props{On: &on, Label: x.key})
		if on != extraBool(sw.s, x.key) {
			sw.s.Extra[x.key] = on
		}
	case "number":
		field.Field(c, field.Props{Label: x.key}, func() {
			v := extraFloat(sw.s, x.key)
			numberinput.NumberInput(c, numberinput.Props{Value: &v, Lo: -1e12, Hi: 1e12, Step: 1})
			if v != extraFloat(sw.s, x.key) {
				sw.s.Extra[x.key] = v
			}
		})
	case "color":
		field.Field(c, field.Props{Label: x.key}, func() {
			st := sw.extraWellState(key)
			colorwell.ColorWell(c, colorwell.Props{Color: &st.color, Label: x.key})
			if st.inUse {
				rb := ui.ButtonBase(c.Key("reset-extra-" + x.key)).Padding(2, 8).Radius(t.Radius).
					FocusRing(false).Tooltip("Use the default color")
				rb.Children(func() {
					ui.Row(c).Gap(4).Children(func() {
						ui.Image(c, lucide.RotateCcw()).Size(12, 12)
						ui.Text(c, "Reset").FontSize(12).TextColor(t.TextMuted)
					})
				})
				if rb.Clicked() {
					delete(sw.s.Extra, x.key)
					st.inUse = false
				}
			}
		})
	case "list":
		field.Field(c, field.Props{Label: x.key}, func() {
			sw.extraListEditor(c, key, x.key)
		})
	default: // string
		field.Field(c, field.Props{Label: x.key}, func() {
			v := extraString(sw.s, x.key)
			input.Input(c, input.Props{Value: &v})
			if v != extraString(sw.s, x.key) {
				sw.s.Extra[x.key] = v
			}
		})
	}
}

func extraBool(s *settings, key string) bool {
	if v, ok := s.Extra[key].(bool); ok {
		return v
	}
	return false
}

func extraFloat(s *settings, key string) float64 {
	if v, ok := s.Extra[key].(float64); ok {
		return v
	}
	return 0
}

func extraString(s *settings, key string) string {
	if v, ok := s.Extra[key].(string); ok {
		return v
	}
	return ""
}

// extraWellState is the color well state of an unmodeled color key, so the
// All Settings list can edit colors through settings.Extra.
func (sw *settingsWindow) extraWellState(key string) *colorWellState {
	if sw.wells.state == nil {
		sw.wells.state = map[string]*colorWellState{}
	}
	st, ok := sw.wells.state[key]
	if !ok {
		st = &colorWellState{}
		sw.wells.state[key] = st
	}
	return st
}

// extraListEditor edits an unmodeled list key as one item per line, with a
// buffer that outlives the frame.
func (sw *settingsWindow) extraListEditor(c *ui.Context, bufKey, key string) {
	if sw.lists == nil {
		sw.lists = map[string]*string{}
	}
	b, ok := sw.lists[bufKey]
	cur := extraList(sw.s, key)
	if !ok {
		joined := strings.Join(cur, "\n")
		b = &joined
		sw.lists[bufKey] = b
	}
	textarea.Textarea(c, textarea.Props{Value: b, Rows: 3})
	if joined := strings.Join(cur, "\n"); *b != joined {
		sw.s.Extra[key] = splitLines(*b)
	}
}

func extraList(s *settings, key string) []string {
	if v, ok := s.Extra[key].([]string); ok {
		return v
	}
	if v, ok := s.Extra[key].([]any); ok {
		out := make([]string, 0, len(v))
		for _, x := range v {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}
