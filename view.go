package main

import (
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
	"github.com/smallnest/uix/components/command"
)

// view is the root of the window, laid out as otty is: a 44px top strip
// with the sessions rail's "TABS" header, the centered user@host:dir
// title and the right panel's toolbar; below it the rail of session tabs,
// the panes in the middle, and the info or chat panel on the right. The
// command palette overlays it all.
func (a *app) view(c *ui.Context) {
	c.SetTheme(ottyTheme(c.Theme().Dark))
	if a.settings.WindowLayout == "sidebar-left" {
		a.viewSidebar(c)
	} else {
		a.viewTabs(c)
	}

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

// viewSidebar is the Sidebar layout: the strip over the whole window, then
// the rail of sessions on the left, the panes in the middle and the open
// info or chat panel on the right, all below the strip.
func (a *app) viewSidebar(c *ui.Context) {
	ui.Column(c).Fill().Children(func() {
		a.topstrip(c)
		ui.Row(c).Grow(1).MinHeight(0).Children(func() {
			if a.sideOpen {
				a.railView(c)
			}
			ui.Box(c).Grow(1).MinWidth(0).FillHeight().Clip().Children(func() {
				a.contentView(c)
				a.cornerToggles(c)
			})
			switch a.panel {
			case panelInfo:
				a.infoView(c)
			case panelChat:
				a.chatView(c)
			}
		})
	})
}

// viewTabs is the Tabs Top and Tabs Bottom layouts: the strip with the tab
// bar's and the info panel's toggles at the title's height, the horizontal
// tab bar above or below the panes, and the open panel as a full-height
// column reaching the top of the window, as otty draws the layouts.
func (a *app) viewTabs(c *ui.Context) {
	// A Row centers its children vertically; the columns must fill the
	// window's height, so they stretch across it.
	ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Children(func() {
			a.topstrip(c)
			if a.sideOpen && a.settings.WindowLayout == "tabs-top" {
				a.tabBarView(c)
			}
			ui.Box(c).FillWidth().Grow(1).MinHeight(0).Clip().Children(func() {
				a.contentView(c)
			})
			if a.sideOpen && a.settings.WindowLayout == "tabs-bottom" {
				a.tabBarView(c)
			}
		})
		switch a.panel {
		case panelInfo:
			a.infoColumn(c)
		case panelChat:
			a.chatColumn(c)
		}
	})
}

// cornerToggles float at the top corners of the middle panel, the two
// places a hidden side can be brought back with the mouse: the rail's
// toggle at the left, the info panel's at the right. Each shows only
// while its side is hidden, so it never covers an open panel.
func (a *app) cornerToggles(c *ui.Context) {
	t := c.Theme()
	if !a.sideOpen {
		b := ui.ButtonBase(c.Key("corner-rail")).Padding(5).Radius(t.Radius).
			FocusRing(false).Tooltip("Toggle sidebar")
		b.Absolute().Top(8).Left(8).Background(t.Surface).BorderWidth(1).BorderColor(t.Border)
		b.Children(func() {
			ui.Image(c, lucide.PanelLeftOpen()).Size(15, 15)
		})
		if b.Clicked() {
			a.toggleSidebar()
		}
	}
	if a.panel == panelNone {
		b := ui.ButtonBase(c.Key("corner-info")).Padding(5).Radius(t.Radius).
			FocusRing(false).Tooltip("Info panel")
		b.Absolute().Top(8).Right(8).Background(t.Surface).BorderWidth(1).BorderColor(t.Border)
		b.Children(func() {
			ui.Image(c, lucide.PanelRightOpen()).Size(15, 15)
		})
		if b.Clicked() {
			a.setPanel(panelInfo)
		}
	}
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
