package main

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
	ui.Row(c).FillWidth().Shrink(0).Height(44).DragWindow().Children(func() {
		if a.sideOpen {
			// The rail's column continues through the strip: its surface
			// and its right border, empty under the traffic lights.
			ui.Box(c).Width(a.sidebarWidth()).FillHeight().Shrink(0).Background(t.Surface).
				BorderWidth(0, 1, 0, 0).BorderColor(t.Border)
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

		// The right surface column: its toolbar when a panel is open, and
		// the bare column otherwise — the info panel's toggle floats at the
		// corner of the middle panel then.
		switch a.panel {
		case panelInfo:
			a.infoToolbar(c)
		case panelChat:
			a.chatToolbar(c)
		default:
			ui.Box(c).Width(220).FillHeight().Shrink(0).Background(t.Surface).
				BorderWidth(0, 0, 0, 1).BorderColor(t.Border)
		}
	})
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
// the info panel: an "Info" pill and icons for the outline, git, files
// and panel toggles. Our clone only implements the info and chat panels,
// so the extra icons switch back to the info panel and the panel toggle
// shows or hides it. The column itself is always there; the icons appear
// only when the pointer is over it, as otty reveals its toolbar.
func (a *app) infoToolbar(c *ui.Context) {
	t := c.Theme()
	box := ui.Box(c).Width(220).FillHeight().Shrink(0).Background(t.Surface).
		BorderWidth(0, 0, 0, 1).BorderColor(t.Border)
	if box.Hovered() {
		box.Children(func() {
			ui.Row(c).FillWidth().Height(a.stripBar(c)).Padding(0, 12).Gap(2).Children(func() {
				pill := ui.ButtonBase(c.Key("toolbar-info")).Padding(5, 9).
					Radius(999).FocusRing(false).Tooltip("Info")
				if a.panel == panelInfo {
					pill.Background(t.SurfaceHover)
				}
				pill.Children(func() {
					ui.Row(c).Gap(5).Children(func() {
						ui.Image(c, lucide.Info()).Size(14, 14)
						ui.Text(c, "Info").FontSize(12).TextColor(t.Text)
					})
				})
				if pill.Clicked() {
					a.toggleInfo()
				}
				a.iconButton(c, "toolbar-outline", lucide.List(), "Outline", func() {
					a.setPanel(panelInfo)
				})
				a.iconButton(c, "toolbar-git", lucide.GitBranch(), "Git", func() {
					a.setPanel(panelInfo)
				})
				a.iconButton(c, "toolbar-files", lucide.Folder(), "Files", func() {
					a.setPanel(panelInfo)
				})
				ui.Box(c).Grow(1)
				a.iconButton(c, "toolbar-panel", lucide.PanelRightClose(), "Toggle panel", func() {
					a.toggleInfo()
				})
			})
		})
	}
}

// chatToolbar is the chat panel's header row, in the same place the info
// toolbar sits so the two panels swap in one spot. Like the other
// toolbars it appears only on hover over the column.
func (a *app) chatToolbar(c *ui.Context) {
	t := c.Theme()
	box := ui.Box(c).Width(360).FillHeight().Shrink(0).Background(t.Surface).
		BorderWidth(0, 0, 0, 1).BorderColor(t.Border)
	if box.Hovered() {
		box.Children(func() {
			ui.Row(c).FillWidth().Height(a.stripBar(c)).Padding(0, 14).Gap(8).Children(func() {
				nutTile(c, 18)
				ui.Text(c, "Agent").Bold()
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
		})
	}
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
