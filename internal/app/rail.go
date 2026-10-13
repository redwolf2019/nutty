package app

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
)

// railView is the sessions rail on the left: a "TABS" header with its menu
// under the top strip, then one compact tab per session with its directory
// and shell on a row, as otty draws them. The selected tab is a white card
// with a shadow; the rest are plain text. A click activates a session, a
// hover shows its close button.
func (a *app) railView(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Width(a.sidebarWidth()).FillHeight().Shrink(0).Background(t.Surface).
		BorderWidth(0, 1, 0, 0).BorderColor(t.Border).Children(func() {
		// The section header: "TABS" and a filter icon that opens the
		// command palette, as otty draws them. Hiding the rail is the
		// strip's left pair, next to the new-tab button.
		ui.Row(c).FillWidth().Padding(0, 9, 0, 18).Gap(6).Children(func() {
			ui.Text(c, tr("TABS")).Bold().FontSize(11).TextColor(faint(t)).
				LetterSpacing(1).Grow(1)
			ib := ui.ButtonBase(c.Key("tabs-menu")).Padding(0).
				Radius(t.Radius).FocusRing(false).Tooltip(tr("Tabs menu"))
			ib.Children(func() {
				ui.Image(c, lucide.ListFilter()).Size(15, 15)
			})
			if ib.Clicked() {
				a.paletteOpen = true
			}
		})
		closeID := 0
		ui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
			ui.Column(c).FillWidth().Padding(4, 9, 8, 9).Gap(4).Children(func() {
				for i, s := range a.sessions {
					if a.railItem(c, i, s) {
						closeID = s.id
					}
				}
			})
		})
		// A close clicked while building must not shift the loop;
		// close after the rail is built.
		if closeID != 0 {
			for _, s := range a.sessions {
				if s.id == closeID {
					a.requestClose(s)
					break
				}
			}
		}
	})
}

// railItem draws one session tab, reporting whether its close button was
// clicked. The rows are as tall as the selected card, so a hover and a
// click cover the same room every tab, as otty's do.
func (a *app) railItem(c *ui.Context, i int, s *session) (close bool) {
	t := c.Theme()
	on := i == a.active
	b := ui.ButtonBase(c.Key("rail"+sessionID(s.id))).FillWidth().
		Padding(9, 10, 9, 17).Radius(t.Radius).FocusRing(false)
	if on {
		b.Background(selectedCard(t)).Shadow(0, 1, 4, 0, ui.RGBA(0, 0, 0, 0.08)).
			BorderWidth(1).BorderColor(t.Border)
	} else if b.Hovered() {
		b.Background(t.SurfaceHover)
	}
	if b.Clicked() {
		a.activate(s)
	}
	if b.DoubleClicked() {
		a.beginRenameSession(s)
	}
	b.Children(func() {
		ui.Row(c).FillWidth().Gap(6).Children(func() {
			name := ui.Text(c, tabLabel(s)).SingleLine().Grow(1).MinWidth(0)
			if on {
				name.Bold()
			} else {
				name.TextColor(t.TextMuted)
			}
			ui.Text(c, shellName()).SingleLine().FontSize(11).TextColor(faint(t))
			if b.Hovered() {
				cl := ui.Button(c, "×").Padding(0, 4).FocusRing(false)
				cl.TextColor(t.TextMuted)
				if cl.Hovered() {
					cl.TextColor(t.Text)
				}
				if cl.Clicked() {
					close = true
				}
			}
		})
	})
	return close
}

// faint is the lightest muted tone, for labels like "TABS" and the
// shell's name, which otty draws barely darker than the surface.
func faint(t *ui.Theme) ui.Color {
	if t.Dark {
		return ui.Hex("#6E6E6C")
	}
	return ui.Hex("#C1BFB6")
}

// selectedCard is the fill of the selected tab: white over the warm
// surface on light, a lighter gray than the surface on dark, as otty
// raises the active tab.
func selectedCard(t *ui.Theme) ui.Color {
	if t.Dark {
		return ui.Hex("#333333")
	}
	return ui.Hex("#FFFFFF")
}

// homeShown writes dir with the home directory shortened to "~", as the
// rail and the info panel show it.
func homeShown(dir string) string {
	home := homeDir()
	if strings.HasPrefix(dir, home) {
		return "~" + strings.TrimPrefix(dir, home)
	}
	return dir
}
