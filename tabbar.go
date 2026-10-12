package main

import (
	"github.com/egoist/mygo/ui"
)

// renameDialog renames the active tab in place, as Finder renames a file.
func (a *app) renameDialog(c *ui.Context) {
	if !a.rename.open {
		return
	}
	t := c.Theme()
	w, h := c.Size()
	ui.DialogBase(c, &a.rename.open, func(back, wp ui.Element) {
		back.Background(ui.RGBA(0, 0, 0, 0.4))
		wp.Width(min(w-48, 340)).MaxHeight(h - 48).
			Radius(t.Radius + 2).Background(t.Background).
			Shadow(0, 10, 30, 0, ui.RGBA(0, 0, 0, 0.3))
		wp.Children(func() {
			ui.Column(c).Padding(16).Gap(12).Children(func() {
				ui.Text(c, "Rename Tab").Bold()
				in := ui.TextInput(c.Key("rename"), &a.rename.buf).
					Bind(&a.rename.input).AutoFocus()
				if in.Submitted() {
					a.commitRename()
				}
				ui.Row(c).FillWidth().Justify(ui.End).Gap(8).Children(func() {
					ui.Button(c, "Cancel").FocusRing(false)
					if ui.PrimaryButton(c, "Rename").Clicked() {
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
		wp.Width(min(w-48, 340)).Radius(t.Radius + 2).Background(t.Background).
			Shadow(0, 10, 30, 0, ui.RGBA(0, 0, 0, 0.3))
		wp.Children(func() {
			ui.Column(c).Padding(16).Gap(12).Children(func() {
				ui.Text(c, "Close Tab?").Bold()
				ui.Text(c, "“"+a.pendingClose.displayTitle()+"” and its processes will end.").
					FontSize(12).TextColor(t.TextMuted)
				ui.Row(c).FillWidth().Justify(ui.End).Gap(8).Children(func() {
					if ui.Button(c, "Cancel").Clicked() {
						a.pendingClose = nil
					}
					if ui.PrimaryButton(c, "Close Tab").Clicked() {
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
