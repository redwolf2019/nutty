package app

import (
	"path/filepath"

	"nutty/internal/settings"

	"github.com/egoist/mygo/ui"
	"github.com/smallnest/uix/components/command"
)

// view is the root of the window, laid out as otty is: a 44px top strip
// with the sessions rail's "TABS" header, the centered user@host:dir
// title and the right panel's toolbar; below it the rail of session tabs,
// the panes in the middle, and the info or chat panel on the right. The
// command palette overlays it all.
func (a *app) view(c *ui.Context) {
	c.SetTheme(settings.OttyTheme(c.Theme().Dark))
	// The terminals read their output on goroutines of their own, so what
	// the shell reported is known before the frame is built: follow it into
	// a new directory first, so this frame's tab, title and info panel show
	// where it is. Tracking after the build would draw the old directory
	// and wait for another frame.
	a.trackWorkingDir()
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
	// The commands are built only while the palette shows them, not for
	// every frame of a terminal's output.
	var items []command.Item
	if a.paletteOpen {
		items = a.paletteItems()
	}
	command.Command(c, command.Props{
		Open:  &a.paletteOpen,
		State: &a.palette,
		Items: items,
	})
}

// trackWorkingDir follows the active shell's directory: when the terminal
// reports a new working directory (OSC 7), the session takes it, so the
// rail tab, the strip's user@host:dir title and the info panel's Working
// Directory show where the shell is, as otty follows its shell. The report
// is whatever a program in the terminal wrote, and the directory goes to
// git and open as an argument, so only an absolute path is taken: a
// relative one, or one that reads as an option ("-a…"), is ignored.
func (a *app) trackWorkingDir() {
	s := a.currentSession()
	if s == nil || a.activePane == nil || a.activePane.term == nil {
		return
	}
	if d := a.activePane.term.Dir(); filepath.IsAbs(d) && d != s.dir {
		s.dir = d
		a.refreshTitle()
		// The Git and Files panels describe the working directory; a new
		// directory makes what they hold stale.
		a.infoDirChanged()
	}
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

// contentView draws the active session's pane tree, or a gentle empty
// state when there is none.
func (a *app) contentView(c *ui.Context) {
	t := c.Theme()
	if len(a.sessions) == 0 {
		ui.Column(c).Fill().Center().Gap(12).Children(func() {
			nutTile(c, 88)
			ui.Text(c, tr("No sessions")).TextColor(t.TextMuted)
			ui.Text(c, tr("Press ⌘T to open a new session")).TextColor(t.TextMuted).FontSize(12)
		})
		return
	}
	a.renderNode(c, a.sessions[a.active].root)
}
