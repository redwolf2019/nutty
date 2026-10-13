package app

import (
	"os"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
)

// stripBar is the room the traffic lights take: the strip's content lines
// up with them, not with the center of the 44px strip, as otty draws it.
// Without a title bar (in tests) the lights are moot, so center then.
func (a *app) stripBar(c *ui.Context) float32 {
	if h := c.TitleBar().Height; h > 0 {
		return h
	}
	return 44
}

// topstrip is the 44px strip across the top of the window: the surface
// color of each column continues through it, with the traffic lights over
// the rail's column, the user@host:dir title centered, and the right
// panel's toolbar on the right, as otty draws them. The strip drags the
// window.
func (a *app) topstrip(c *ui.Context) {
	t := c.Theme()
	bar := c.TitleBar()
	// A press on the strip or its buttons leaves the keyboard focus in the
	// terminal, as a press on the rail does (see railView).
	ui.Row(c).FillWidth().Shrink(0).Height(44).DragWindow().KeepFocus().Children(func() {
		if a.sideOpen && a.settings.WindowLayout == "sidebar-left" {
			// The rail's column continues through the strip: its surface and
			// its right border, with a new-tab button and the rail's toggle
			// near its right edge, revealed on hover, as otty draws them.
			rail := ui.Box(c).Width(a.sidebarWidth()).FillHeight().Shrink(0).Background(t.Surface).
				BorderWidth(0, 1, 0, 0).BorderColor(t.Border)
			if rail.Hovered() {
				rail.Children(func() {
					a.railStripToggles(c)
				})
			}
		} else if a.settings.WindowLayout == "sidebar-left" {
			// The rail is hidden, so the traffic lights sit over the title;
			// hovering the strip's left shows the toggle that brings the
			// rail back, right of the lights.
			left := ui.Box(c).Width(bar.Left + 6 + 24).FillHeight().Shrink(0)
			if left.Hovered() {
				left.Children(func() {
					ui.Row(c).FillHeight().Height(a.stripBar(c)).Gap(2).Children(func() {
						ui.Box(c).Width(bar.Left + 6)
						a.stripButton(c, "strip-rail-open", lucide.PanelLeft(), 12, tr("Show sidebar"), func() {
							a.toggleSidebar()
						})
					})
				})
			}
		} else {
			// Without the rail the traffic lights sit over the title.
			ui.Box(c).Width(bar.Left + 6).Shrink(0)
		}

		// The centered title, aligned to the traffic lights.
		ui.Box(c).Grow(1).FillHeight().Children(func() {
			ui.Box(c).FillWidth().Height(a.stripBar(c)).Children(func() {
				ui.Column(c).Fill().Center().Children(func() {
					ui.Text(c, a.titleText()).SingleLine().FillWidth().TextAlign(ui.Center).
						TextColor(t.TextMuted)
				})
			})
		})

		// The strip's hidden controls, as otty reveals them: a new-tab
		// button and the sessions toggle at the left, over the rail, and
		// the tab bar's or the rail's hide toggle with the info panel's
		// toggle at the right. All sit at the title's height and appear
		// only while the pointer is over the strip.
		switch {
		case a.settings.WindowLayout != "sidebar-left":
			a.stripControls(c)
		case a.panel == panelInfo:
			a.infoToolbar(c)
		case a.panel == panelChat:
			a.chatToolbar(c)
		default:
			a.sidebarStripToggles(c)
		}
	})
}

// stripControls is the right end of the Tabs layouts' strip: the toggle
// that hides the tab bar, and — while no panel is open — the toggle that
// opens the info panel, at the title's height. The closed panel leaves no
// column behind it, so the strip is the title's color there, and the
// controls appear only while the pointer is over the room the panel would
// take, as otty reveals its title-strip controls. An open panel draws its
// own toggle in its toolbar at the same height.
func (a *app) stripControls(c *ui.Context) {
	box := ui.Box(c).Width(220).FillHeight().Shrink(0)
	if box.Hovered() {
		box.Children(func() {
			ui.Row(c).FillHeight().Height(a.stripBar(c)).Padding(0, 12).Gap(2).Children(func() {
				ui.Box(c).Grow(1)
				a.stripButton(c, "tabbar-toggle", lucide.PanelLeft(), 12, tr("Toggle tabs"), func() {
					a.toggleSidebar()
				})
				if a.panel == panelNone {
					a.stripButton(c, "strip-panel", lucide.PanelRight(), 12, tr("Info panel"), func() {
						a.toggleInfo()
					})
				}
			})
		})
	}
}

// sidebarStripToggles is the right end of the Sidebar layout's strip while
// no panel is open: the toggle that opens the info panel, at the title's
// height. The closed panel leaves no column behind it, so the strip is the
// title's color there, and the toggle appears only while the pointer is
// over the room the panel would take, as otty reveals its title-strip
// controls. Hiding the rail is the strip's left pair, next to the new-tab
// button, as otty draws them. An open panel's toolbar lives here instead.
func (a *app) sidebarStripToggles(c *ui.Context) {
	box := ui.Box(c).Width(220).FillHeight().Shrink(0)
	if box.Hovered() {
		box.Children(func() {
			ui.Row(c).FillHeight().Height(a.stripBar(c)).Padding(0, 12).Gap(2).Children(func() {
				ui.Box(c).Grow(1)
				a.stripButton(c, "strip-panel", lucide.PanelRight(), 12, tr("Info panel"), func() {
					a.toggleInfo()
				})
			})
		})
	}
}

// railStripToggles is the left end of the Sidebar layout's strip, over the
// sessions rail: a new-tab button and the toggle that hides the rail, at
// the title's height and near the rail's right edge, as otty shows them.
// They appear only while the pointer is over the rail's column.
func (a *app) railStripToggles(c *ui.Context) {
	ui.Row(c).FillHeight().Height(a.stripBar(c)).Padding(0, 12).Gap(2).Children(func() {
		ui.Box(c).Grow(1)
		a.stripButton(c, "strip-newtab", lucide.Plus(), 11, tr("New tab"), func() {
			a.newTab()
		})
		a.stripButton(c, "strip-rail", lucide.PanelLeft(), 12, tr("Toggle sidebar"), func() {
			a.toggleSidebar()
		})
	})
}

// stripButton is a bare icon button of the title strip, at the small size
// otty draws its strip controls.
func (a *app) stripButton(c *ui.Context, key string, icon *ui.SVG, size float32, tip string, onClick func()) {
	t := c.Theme()
	b := ui.ButtonBase(c.Key(key)).Padding(4, 4).Radius(t.Radius).FocusRing(false).Tooltip(tip)
	b.Children(func() {
		ui.Image(c, icon).Size(size, size)
	})
	if b.Clicked() {
		onClick()
	}
}

// titleText is the window title: user@host:dir, as otty shows it.
func (a *app) titleText() string {
	s := a.currentSession()
	if s == nil {
		return "Nutty"
	}
	u := os.Getenv("USER")
	if u == "" {
		u = "user"
	}
	h, err := os.Hostname()
	if err != nil {
		h = "host"
	}
	return u + "@" + h + ":" + homeShown(s.dir)
}

// infoToolbar is the right panel's mode switch, as otty shows it above
// the info panel: an "Info" pill and icons for the outline, git and files
// sections, then the panel toggle. The open section is highlighted; the
// pill's click closes the panel when the Info section is open, and the
// panel toggle shows or hides it. The open panel's toolbar stays on the
// strip, as otty always shows it.
func (a *app) infoToolbar(c *ui.Context) {
	a.checkInfoTab()
	t := c.Theme()
	ui.Box(c).Width(a.detailsPanelWidth()).FillHeight().Shrink(0).Background(t.Surface).
		BorderWidth(0, 0, 0, 1).BorderColor(t.Border).Children(func() {
		a.infoToolbarRow(c)
	})
}

// infoToolbarRow is the info panel's mode switch: the section that is open
// shows its label on a pill, the others are bare icons, and the panel
// toggle sits at the right, at the strip's height. The Sidebar layout
// draws it in the top strip; the Tabs layouts draw it at the top of the
// panel's full-height column.
func (a *app) infoToolbarRow(c *ui.Context) {
	// The Tabs layouts draw this row outside the strip, at the top of the
	// panel's column, so it keeps the focus in the terminal itself: its
	// buttons would take it otherwise, as the strip's would (see topstrip).
	ui.Row(c).KeepFocus().FillWidth().Height(a.stripBar(c)).Padding(0, 12).Gap(2).Children(func() {
		// The sections come in the order of the View page's Details Panel
		// list, and the ones it switches off do not draw their button, as
		// otty drops them.
		for _, tab := range a.detailsTabs() {
			switch tab {
			case infoTabInfo:
				a.infoTabButton(c, "toolbar-info", lucide.Info(), tr("Info"), infoTabInfo, func() {
					if a.info.tab == infoTabInfo {
						a.toggleInfo()
					} else {
						a.setInfoTab(infoTabInfo)
					}
				})
			case infoTabOutline:
				a.infoTabButton(c, "toolbar-outline", lucide.List(), tr("Outline"), infoTabOutline, nil)
			case infoTabGit:
				a.infoTabButton(c, "toolbar-git", lucide.GitBranch(), tr("Git"), infoTabGit, nil)
			case infoTabFiles:
				a.infoTabButton(c, "toolbar-files", lucide.Folder(), tr("Files"), infoTabFiles, nil)
			}
		}
		ui.Box(c).Grow(1)
		a.iconButton(c, "toolbar-panel", lucide.PanelRightClose(), tr("Toggle panel"), func() {
			a.toggleInfo()
		})
	})
}

// infoTabButton is one section of the info panel's mode switch: the open
// section is a labeled pill, the others are bare icons, so only the open
// section's text shows, as otty shows the switch. The Info button closes
// the panel when it is the open section.
func (a *app) infoTabButton(c *ui.Context, key string, icon *ui.SVG, label string, tab infoTab, onClick func()) {
	t := c.Theme()
	b := ui.ButtonBase(c.Key(key)).Radius(999).FocusRing(false).Tooltip(label)
	if a.info.tab == tab {
		b.Background(t.SurfaceHover).Padding(5, 9)
	} else {
		b.Padding(5, 5)
	}
	b.Children(func() {
		ui.Row(c).Gap(5).Children(func() {
			ui.Image(c, icon).Size(14, 14)
			if a.info.tab == tab {
				ui.Text(c, label).FontSize(12).TextColor(t.Text)
			}
		})
	})
	if b.Clicked() {
		if onClick != nil {
			onClick()
		} else {
			a.setInfoTab(tab)
		}
	}
}

// chatToolbar is the chat panel's header row, in the same place the info
// toolbar sits so the two panels swap in one spot. The open panel's
// header stays on the strip, as otty always shows it.
func (a *app) chatToolbar(c *ui.Context) {
	t := c.Theme()
	ui.Box(c).Width(360).FillHeight().Shrink(0).Background(t.Surface).
		BorderWidth(0, 0, 0, 1).BorderColor(t.Border).Children(func() {
		a.chatToolbarRow(c)
	})
}

// chatToolbarRow is the chat panel's header: the nut, the "Agent" title
// and the close button, at the strip's height. The Sidebar layout draws
// it in the top strip; the Tabs layouts at the top of the panel's
// full-height column.
func (a *app) chatToolbarRow(c *ui.Context) {
	t := c.Theme()
	// As the info panel's row, it keeps the focus in the terminal where the
	// Tabs layouts draw it outside the strip (see infoToolbarRow).
	ui.Row(c).KeepFocus().FillWidth().Height(a.stripBar(c)).Padding(0, 14).Gap(8).Children(func() {
		nutTile(c, 18)
		ui.Text(c, tr("Agent")).Bold()
		ui.Box(c).Grow(1)
		cl := ui.Button(c, "×").Padding(0, 5).FocusRing(false)
		cl.TextColor(t.TextMuted)
		if cl.Hovered() {
			cl.TextColor(t.Text)
		}
		if cl.Clicked() {
			a.setPanel(panelNone)
		}
	})
}

// iconButton is a bare icon button of a toolbar.
func (a *app) iconButton(c *ui.Context, key string, icon *ui.SVG, tip string, onClick func()) {
	t := c.Theme()
	b := ui.ButtonBase(c.Key(key)).Padding(5, 5).Radius(t.Radius).
		FocusRing(false).Tooltip(tip)
	b.Children(func() {
		ui.Image(c, icon).Size(15, 15)
	})
	if b.Clicked() {
		onClick()
	}
}
