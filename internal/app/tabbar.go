package app

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// tabBarView is the horizontal tab bar of the Tabs Top and Tabs Bottom
// layouts: one tab per session in a row, the active one a raised card, a
// toggle at the end that hides the bar, as otty's layouts show the tabs
// across the top or bottom of the window instead of the left rail.
func (a *app) tabBarView(c *ui.Context) {
	t := c.Theme()
	top := a.settings.WindowLayout == "tabs-top"
	closeID := 0
	bar := ui.Row(c).FillWidth().Height(38).Shrink(0).Background(t.Surface)
	// The border separates the bar from the terminal: below it when the
	// bar is at the top, above it when at the bottom.
	if top {
		bar.BorderWidth(0, 0, 1, 0)
	} else {
		bar.BorderWidth(1, 0, 0, 0)
	}
	bar.BorderColor(t.Border)
	bar.Children(func() {
		ui.ScrollHorizontal(c).Grow(1).MinWidth(0).Children(func() {
			ui.Row(c).Padding(5, 6).Gap(4).Children(func() {
				for i, s := range a.sessions {
					if a.tabItem(c, i, s) {
						closeID = s.id
					}
				}
			})
		})
	})
	// A close clicked while building must not shift the loop; close after
	// the bar is built.
	if closeID != 0 {
		for _, s := range a.sessions {
			if s.id == closeID {
				a.closeSession(s)
				break
			}
		}
	}
}

// tabItem draws one horizontal tab, reporting whether its close button
// was clicked.
func (a *app) tabItem(c *ui.Context, i int, s *session) (close bool) {
	t := c.Theme()
	on := i == a.active
	b := ui.ButtonBase(c.Key("tab"+sessionID(s.id))).Padding(6, 12).
		Radius(t.Radius).FocusRing(false)
	if on {
		b.Background(selectedCard(t)).BorderWidth(1).BorderColor(t.Border)
	} else if b.Hovered() {
		b.Background(t.SurfaceHover)
	}
	if b.Clicked() {
		a.activate(s)
	}
	b.Children(func() {
		ui.Row(c).Gap(6).Children(func() {
			name := ui.Text(c, tabLabel(s)).SingleLine().FontSize(12)
			if !on {
				name.TextColor(t.TextMuted)
			}
			if b.Hovered() {
				cl := ui.Button(c, "×").Padding(0, 3).FocusRing(false)
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

// renameDialog renames the active tab in place, as Finder renames a file.
func (a *app) renameDialog(c *ui.Context) {
	if !a.rename.open {
		return
	}
	t := c.Theme()
	w, h := c.Size()
	ui.DialogBase(c, &a.rename.open, func(back, wp ui.Element) {
		back.Background(ui.RGBA(0, 0, 0, 0.4))
		wp.Width(min(w-48, 340)).MaxHeight(h-48).
			Radius(t.Radius+2).Background(t.Background).
			Shadow(0, 10, 30, 0, ui.RGBA(0, 0, 0, 0.3))
		wp.Children(func() {
			ui.Column(c).Padding(16).Gap(12).Children(func() {
				ui.Text(c, tr("Rename Tab")).Bold()
				in := ui.TextInput(c.Key("rename"), &a.rename.buf).
					Bind(&a.rename.input).AutoFocus()
				if in.Submitted() {
					a.commitRename()
				}
				ui.Row(c).FillWidth().Justify(ui.End).Gap(8).Children(func() {
					ui.Button(c, tr("Cancel")).FocusRing(false)
					if ui.PrimaryButton(c, tr("Rename")).Clicked() {
						a.commitRename()
					}
				})
			})
		})
	})
}

// confirmDialog asks before closing a tab, when the settings ask it to.
// The close waits for the button; the backdrop and Escape cancel it.
func (a *app) confirmDialog(c *ui.Context) {
	if a.pendingClose == nil {
		return
	}
	t := c.Theme()
	w, _ := c.Size()
	open := true
	ui.DialogBase(c, &open, func(back, wp ui.Element) {
		back.Background(ui.RGBA(0, 0, 0, 0.4))
		wp.Width(min(w-48, 340)).Radius(t.Radius+2).Background(t.Background).
			Shadow(0, 10, 30, 0, ui.RGBA(0, 0, 0, 0.3))
		wp.Children(func() {
			ui.Column(c).Padding(16).Gap(12).Children(func() {
				ui.Text(c, tr("Close Tab?")).Bold()
				ui.Text(c, fmt.Sprintf(tr("“%s” and its processes will end."), a.pendingClose.displayTitle())).
					FontSize(12).TextColor(t.TextMuted)
				ui.Row(c).FillWidth().Justify(ui.End).Gap(8).Children(func() {
					if ui.Button(c, tr("Cancel")).Clicked() {
						a.pendingClose = nil
					}
					if ui.PrimaryButton(c, tr("Close Tab")).Clicked() {
						s := a.pendingClose
						a.pendingClose = nil
						a.closeSession(s)
					}
				})
			})
		})
	})
	// The backdrop or Escape closed the dialog without a button.
	if !open && a.pendingClose != nil {
		a.pendingClose = nil
	}
}
