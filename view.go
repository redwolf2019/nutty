package main

import (
	"github.com/egoist/mygo/ui"
	"github.com/smallnest/uix/components/command"
)

// view is the root of the window, laid out as otty is: a 44px top strip
// with the sessions rail's "TABS" header, the centered user@host:dir
// title and the right panel's toolbar; below it the rail of session tabs,
// the panes in the middle, and the info or chat panel on the right. The
// command palette overlays it all.
func (a *app) view(c *ui.Context) {
	c.SetTheme(ottyTheme(c.Theme().Dark))
	ui.Column(c).Fill().Children(func() {
		a.topstrip(c)
		ui.Row(c).Grow(1).MinHeight(0).Children(func() {
			if a.sideOpen {
				a.railView(c)
			}
			ui.Box(c).Grow(1).MinWidth(0).FillHeight().Clip().Children(func() {
				a.contentView(c)
			})
			switch a.panel {
			case panelInfo:
				a.infoView(c)
			case panelChat:
				a.chatView(c)
			}
		})
	})

	// The palette opens with Cmd+K as well as the menu's Cmd+Shift+P.
	if c.Shortcut(ui.Cmd, ui.KeyK) {
		a.paletteOpen = true
	}

	a.renameDialog(c)
	a.confirmDialog(c)
	command.Command(c, command.Props{
		Open:  &a.paletteOpen,
		State: &a.palette,
		Items: a.paletteItems(),
	})
}

// contentView draws the active session's pane tree, or a gentle empty
// state when there is none.
func (a *app) contentView(c *ui.Context) {
	t := c.Theme()
	if len(a.sessions) == 0 {
		ui.Column(c).Fill().Center().Gap(12).Children(func() {
			nutTile(c, 88)
			ui.Text(c, "No sessions").TextColor(t.TextMuted)
			ui.Text(c, "Press ⌘T to open a new session").TextColor(t.TextMuted).FontSize(12)
		})
		return
	}
	a.renderNode(c, a.sessions[a.active].root)
}
