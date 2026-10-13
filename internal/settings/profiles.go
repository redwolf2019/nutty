package settings

import (
	"fmt"
	"strings"

	"nutty/internal/i18n"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
	"github.com/smallnest/uix/components/field"
	"github.com/smallnest/uix/components/input"
	uixselect "github.com/smallnest/uix/components/select"
	"github.com/smallnest/uix/components/textarea"
)

// pageProfiles is the Profiles category: the profiles on the left, the
// chosen one's Settings on the right. New panes launch the default
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
					ui.Text(c, i18n.Tr("New Profile")).FontSize(12).TextColor(t.Text)
				})
			})
			if ab.Clicked() {
				sw.s.Profiles = append(sw.s.Profiles, Profile{Name: sw.newProfileName()})
				sw.profileSel = len(sw.s.Profiles) - 1
			}
			for i := range sw.s.Profiles {
				i := i
				p := &sw.s.Profiles[i]
				rb := ui.ButtonBase(c.Key("profile-sel-"+fmt.Sprint(i))).FillWidth().
					Padding(5, 10).Radius(t.Radius).FocusRing(false)
				if sw.profileSel == i {
					rb.Background(t.SurfacePressed)
				}
				rb.Children(func() {
					ui.Row(c).Gap(6).Children(func() {
						ui.Text(c, p.Name).FontSize(12).SingleLine().Grow(1).
							TextColor(sw.profileColor(t, i))
						if p.Name == sw.s.DefaultProfile {
							ui.Text(c, i18n.Tr("Default")).FontSize(10).TextColor(t.TextMuted)
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
				sw.hint(c, i18n.Tr("Select a profile to edit it, or add one."))
				return
			}
			p := &sw.s.Profiles[sw.profileSel]
			sw.profileField(c, i18n.Tr("Name"), func() {
				sw.profileNameEditor(c, p)
			})
			sw.profileField(c, i18n.Tr("Command"), func() {
				input.Input(c, input.Props{Value: &p.Command, Placeholder: i18n.Tr("The shell new panes run")})
			})
			sw.profileField(c, i18n.Tr("Run After"), func() {
				input.Input(c, input.Props{Value: &p.RunAfter, Placeholder: i18n.Tr("A command to run when the pane opens")})
			})
			sw.profileField(c, i18n.Tr("Working Directory"), func() {
				input.Input(c, input.Props{Value: &p.CWD, Placeholder: i18n.Tr("Where the pane starts")})
			})
			sw.profileField(c, i18n.Tr("Title Prefix"), func() {
				input.Input(c, input.Props{Value: &p.TitlePrefix})
			})
			sw.profileField(c, i18n.Tr("Theme"), func() {
				sw.profileThemeSelect(c, &p.Theme)
			})
			sw.profileField(c, i18n.Tr("Dark Theme"), func() {
				sw.profileThemeSelect(c, &p.ThemeDark)
			})
			sw.profileField(c, i18n.Tr("Environment"), func() {
				sw.profileEnvEditor(c, p)
			})
			ui.Box(c).FillWidth().Height(4)
			ui.Row(c).Gap(8).Children(func() {
				if p.Name != sw.s.DefaultProfile && p.Name != "" {
					sw.advButton(c, "profile-default", i18n.Tr("Set as Default"), func() {
						sw.s.DefaultProfile = p.Name
					})
				}
				if len(sw.s.Profiles) > 1 {
					sw.advButton(c, "profile-dup", i18n.Tr("Duplicate"), func() {
						dup := *p
						dup.Name = dup.Name + i18n.Tr(" copy")
						for sw.profileNamed(dup.Name) {
							dup.Name += i18n.Tr(" copy") // a copy of a copy
						}
						sw.s.Profiles = append(sw.s.Profiles, dup)
						sw.profileSel = len(sw.s.Profiles) - 1
					})
				}
				sw.advButton(c, "profile-del", i18n.Tr("Delete"), func() {
					if sw.s.Profile() == p {
						sw.s.DefaultProfile = ""
					}
					sw.s.Profiles = append(sw.s.Profiles[:sw.profileSel], sw.s.Profiles[sw.profileSel+1:]...)
					sw.profileSel = -1
					// The editors' buffers are kept by a profile's index, and
					// the profiles after this one each moved down by one: drop
					// the buffers, or the next profile shown takes this one's
					// environment.
					sw.lists = nil
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
// "Follow system", as the Appearance page's selects. Theme names are
// proper nouns and stay as-is; only the follow-system option is
// translated, and the stored value stays the empty name either way.
func (sw *settingsWindow) profileThemeSelect(c *ui.Context, set *string) {
	names := ThemeNames()
	labels := make([]string, len(names))
	cur := *set
	show := ""
	for i, n := range names {
		labels[i] = n
		if n == "Follow system" {
			labels[i] = i18n.Tr(n)
		}
		if n == cur {
			show = labels[i]
		} else if cur == "" && n == "Follow system" {
			show = labels[i]
		}
	}
	uixselect.Select(c, uixselect.Props{Selected: &show, Options: labels})
	fs := i18n.Tr("Follow system")
	if show != "" && show != fs && *set != show {
		*set = show
	} else if show == fs && *set != "" {
		*set = ""
	}
}

// profileEnvEditor edits a profile's environment, one entry per line,
// with a buffer that outlives the frame (see renderList).
func (sw *settingsWindow) profileEnvEditor(c *ui.Context, p *Profile) {
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

// profileNameEditor edits a profile's name through a buffer of its own,
// kept with the list editors' so it is dropped when they are. The default
// profile is stored by its name, so a name has to stay one profile's: what
// is typed renames the profile only while it is not empty and no other
// profile has it, and the default follows its profile to the new name.
// Out of focus the field shows the name the profile has.
func (sw *settingsWindow) profileNameEditor(c *ui.Context, p *Profile) {
	key := fmt.Sprintf("profile-name-%d", sw.profileSel)
	if sw.lists == nil {
		sw.lists = map[string]*string{}
	}
	b, ok := sw.lists[key]
	if !ok {
		name := p.Name
		b = &name
		sw.lists[key] = b
	}
	in := input.Input(c, input.Props{Value: b})
	if in.Changed() {
		if *b != "" && !sw.profileNamed(*b) {
			if sw.s.Profile() == p {
				sw.s.DefaultProfile = *b
			}
			p.Name = *b
		}
	} else if !in.Focused() {
		*b = p.Name
	}
}

// profileNamed reports whether a profile has the name.
func (sw *settingsWindow) profileNamed(name string) bool {
	for i := range sw.s.Profiles {
		if sw.s.Profiles[i].Name == name {
			return true
		}
	}
	return false
}

// newProfileName is the next "Profile N" no profile has: counting the
// profiles alone gives a name twice once a profile was deleted.
func (sw *settingsWindow) newProfileName() string {
	for n := len(sw.s.Profiles) + 1; ; n++ {
		if name := fmt.Sprintf("Profile %d", n); !sw.profileNamed(name) {
			return name
		}
	}
}
