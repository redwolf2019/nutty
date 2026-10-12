package main

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
	"github.com/smallnest/uix/components/field"
	"github.com/smallnest/uix/components/input"
	uixselect "github.com/smallnest/uix/components/select"
	"github.com/smallnest/uix/components/textarea"
)

// pageProfiles is the Profiles category: the profiles on the left, the
// chosen one's settings on the right. New panes launch the default
// profile, which sets the command, the directory, the environment and the
// themes of a session (see newTerm).
func (sw *settingsWindow) pageProfiles(c *ui.Context) {
	if sw.profileSel >= len(sw.s.Profiles) {
		sw.profileSel = -1
	}
	t := c.Theme()
	ui.Row(c).FillWidth().Gap(16).Children(func() {
		// The list: an add button, then one row per profile.
		ui.Column(c).Width(180).Shrink(0).Gap(6).Children(func() {
			ab := ui.ButtonBase(c.Key("profile-add")).FillWidth().Padding(5, 10).
				Radius(t.Radius).FocusRing(false)
			ab.Children(func() {
				ui.Row(c).Gap(6).Children(func() {
					ui.Image(c, lucide.Plus()).Size(13, 13)
					ui.Text(c, "New Profile").FontSize(12).TextColor(t.Text)
				})
			})
			if ab.Clicked() {
				sw.s.Profiles = append(sw.s.Profiles, profile{Name: fmt.Sprintf("Profile %d", len(sw.s.Profiles)+1)})
				sw.profileSel = len(sw.s.Profiles) - 1
			}
			for i := range sw.s.Profiles {
				i := i
				p := &sw.s.Profiles[i]
				rb := ui.ButtonBase(c.Key("profile-sel-" + fmt.Sprint(i))).FillWidth().
					Padding(5, 10).Radius(t.Radius).FocusRing(false)
				if sw.profileSel == i {
					rb.Background(t.SurfacePressed)
				}
				rb.Children(func() {
					ui.Row(c).Gap(6).Children(func() {
						ui.Text(c, p.Name).FontSize(12).SingleLine().Grow(1).
							TextColor(sw.profileColor(t, i))
						if p.Name == sw.s.DefaultProfile {
							ui.Text(c, "Default").FontSize(10).TextColor(t.TextMuted)
						}
					})
				})
				if rb.Clicked() {
					sw.profileSel = i
				}
			}
		})
		// The editor of the chosen profile.
		ui.Column(c).Grow(1).MinWidth(0).Gap(8).Children(func() {
			if sw.profileSel < 0 {
				sw.hint(c, "Select a profile to edit it, or add one.")
				return
			}
			p := &sw.s.Profiles[sw.profileSel]
			sw.profileField(c, "Name", func() {
				input.Input(c, input.Props{Value: &p.Name})
			})
			sw.profileField(c, "Command", func() {
				input.Input(c, input.Props{Value: &p.Command, Placeholder: "The shell new panes run"})
			})
			sw.profileField(c, "Run After", func() {
				input.Input(c, input.Props{Value: &p.RunAfter, Placeholder: "A command to run when the pane opens"})
			})
			sw.profileField(c, "Working Directory", func() {
				input.Input(c, input.Props{Value: &p.CWD, Placeholder: "Where the pane starts"})
			})
			sw.profileField(c, "Title Prefix", func() {
				input.Input(c, input.Props{Value: &p.TitlePrefix})
			})
			sw.profileField(c, "Theme", func() {
				sw.profileThemeSelect(c, &p.Theme)
			})
			sw.profileField(c, "Dark Theme", func() {
				sw.profileThemeSelect(c, &p.ThemeDark)
			})
			sw.profileField(c, "Environment", func() {
				sw.profileEnvEditor(c, p)
			})
			ui.Box(c).FillWidth().Height(4)
			ui.Row(c).Gap(8).Children(func() {
				if p.Name != sw.s.DefaultProfile && p.Name != "" {
					sw.advButton(c, "profile-default", "Set as Default", func() {
						sw.s.DefaultProfile = p.Name
					})
				}
				if len(sw.s.Profiles) > 1 {
					sw.advButton(c, "profile-dup", "Duplicate", func() {
						dup := *p
						dup.Name = dup.Name + " copy"
						sw.s.Profiles = append(sw.s.Profiles, dup)
						sw.profileSel = len(sw.s.Profiles) - 1
					})
				}
				sw.advButton(c, "profile-del", "Delete", func() {
					if sw.s.DefaultProfile == p.Name {
						sw.s.DefaultProfile = ""
					}
					sw.s.Profiles = append(sw.s.Profiles[:sw.profileSel], sw.s.Profiles[sw.profileSel+1:]...)
					sw.profileSel = -1
				})
			})
		})
	})
}

func (sw *settingsWindow) profileColor(t *ui.Theme, i int) ui.Color {
	if sw.profileSel == i {
		return t.Text
	}
	return t.TextMuted
}

// profileField is a labeled field of the profile editor.
func (sw *settingsWindow) profileField(c *ui.Context, label string, fn func()) {
	field.Field(c, field.Props{Label: label}, fn)
}

// profileThemeSelect is a theme select of a profile; the empty name is
// "Follow system", as the Appearance page's selects.
func (sw *settingsWindow) profileThemeSelect(c *ui.Context, set *string) {
	labels := make([]string, len(themeNames()))
	cur := *set
	show := ""
	for i, n := range themeNames() {
		labels[i] = n
		if n == cur {
			show = n
		} else if cur == "" && n == "Follow system" {
			show = n
		}
	}
	uixselect.Select(c, uixselect.Props{Selected: &show, Options: labels})
	if show != "" && show != "Follow system" && *set != show {
		*set = show
	} else if show == "Follow system" && *set != "" {
		*set = ""
	}
}

// profileEnvEditor edits a profile's environment, one entry per line,
// with a buffer that outlives the frame (see renderList).
func (sw *settingsWindow) profileEnvEditor(c *ui.Context, p *profile) {
	key := fmt.Sprintf("profile-env-%d", sw.profileSel)
	if sw.lists == nil {
		sw.lists = map[string]*string{}
	}
	b, ok := sw.lists[key]
	cur := strings.Join(p.Env, "\n")
	if !ok {
		b = &cur
		sw.lists[key] = b
	}
	textarea.Textarea(c, textarea.Props{Value: b, Rows: 3})
	if *b != cur {
		p.Env = splitLines(*b)
	}
}
