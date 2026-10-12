package main

import (
	"context"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
)

// infoState is the info panel's data: the ports of the last scan.
type infoState struct {
	ports     []string
	portsDone bool
}

// infoView is the panel on the right with what otty shows there: the
// working directory with actions, the running process, and the ports in
// use. Its toolbar lives in the top strip; here are only the sections.
func (a *app) infoView(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Width(220).FillHeight().Shrink(0).Background(t.Surface).
		BorderWidth(0, 0, 0, 1).BorderColor(t.Border).Children(func() {
			ui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
				ui.Column(c).FillWidth().Padding(14, 18).Gap(20).Children(func() {
					a.infoDirectory(c)
					a.infoProcess(c)
					a.infoPorts(c)
				})
			})
		})
}

// infoSection is one titled block of the info panel. The title is plain
// text, small and muted, as otty draws its section headers.
func (a *app) infoSection(c *ui.Context, title string, fn func()) {
	t := c.Theme()
	ui.Column(c).FillWidth().Gap(6).Children(func() {
		ui.Text(c, title).FontSize(11).TextColor(t.TextMuted)
		fn()
	})
}

func (a *app) infoDirectory(c *ui.Context) {
	t := c.Theme()
	s := a.currentSession()
	a.infoSection(c, "Working Directory", func() {
		if s == nil {
			ui.Text(c, "—").TextColor(t.TextMuted)
			return
		}
		ui.Column(c).FillWidth().Gap(12).Children(func() {
			ui.Text(c, homeShown(s.dir)).SingleLine().FontSize(12).TextColor(t.Text)
			ui.Column(c).FillWidth().Gap(4).Children(func() {
				a.infoAction(c, "Copy Path", lucide.Copy(), func() {
					mygo.Clipboard.WriteText(s.dir)
				})
				a.infoAction(c, "Reveal in Finder", lucide.FolderOpen(), func() {
					exec.Command("open", s.dir).Start()
				})
				a.infoAction(c, "Open in Zed", lucide.ArrowUpRight(), func() {
					openZed(s.dir)
				})
			})
		})
	})
}

func (a *app) infoProcess(c *ui.Context) {
	t := c.Theme()
	a.infoSection(c, "Process", func() {
		// Otty lists the children of the session's shell here. MyGo's
		// terminal plugin exposes no child process, so show the idle
		// state, as otty does when nothing runs.
		ui.Text(c, "No child processes").FontSize(11).TextColor(t.TextMuted)
	})
}

func (a *app) infoPorts(c *ui.Context) {
	t := c.Theme()
	a.infoSection(c, "Ports", func() {
		switch {
		case !a.info.portsDone:
			ui.Text(c, "Scanning…").TextColor(t.TextMuted)
		case len(a.info.ports) == 0:
			ui.Text(c, "No listening ports").TextColor(t.TextMuted)
		default:
			ui.Column(c).FillWidth().Gap(4).Children(func() {
				for _, p := range a.info.ports {
					ui.Text(c, p).FontSize(11).TextColor(t.Text)
				}
			})
		}
	})
}

// infoAction is one row of the working directory's actions: an icon and a
// green label, as otty draws them.
func (a *app) infoAction(c *ui.Context, label string, icon *ui.SVG, onClick func()) {
	t := c.Theme()
	b := ui.ButtonBase(c.Key("info-" + strings.ReplaceAll(label, " ", "-"))).
		FillWidth().Padding(5, 0).Radius(t.Radius).FocusRing(false)
	b.Children(func() {
		ui.Row(c).Gap(7).Children(func() {
			ui.Image(c, icon).Size(14, 14)
			ui.Text(c, label).FontSize(12).TextColor(actionGreen(t))
		})
	})
	if b.Clicked() {
		onClick()
	}
}

// actionGreen is the green of the info panel's action rows: a muted green
// on light, a brighter one on dark, as otty colors them.
func actionGreen(t *ui.Theme) ui.Color {
	if t.Dark {
		return ui.Hex("#7fbf8a")
	}
	return ui.Hex("#78877C")
}

// loadPorts scans the ports the window's processes listen on, in the
// background, and puts them on the panel when the scan ends.
func (a *app) loadPorts() {
	if a.win == nil {
		return
	}
	a.info.ports = nil
	a.info.portsDone = false
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		out, err := exec.CommandContext(ctx, "lsof", "-nP", "-iTCP", "-sTCP:LISTEN").Output()
		var ports []string
		if err == nil {
			seen := map[string]bool{}
			for _, line := range strings.Split(string(out), "\n") {
				f := strings.Fields(line)
				if len(f) < 9 {
					continue
				}
				if p := portOf(f[8]); p != "" && !seen[p] {
					seen[p] = true
					ports = append(ports, p)
				}
			}
			sort.Strings(ports)
		}
		a.win.Update(func() {
			a.info.ports = ports
			a.info.portsDone = true
		})
	}()
}

// portOf extracts the port from an lsof name like "*:8000 (LISTEN)" or
// "[::1]:8000 (LISTEN)", or "" when there is none.
func portOf(name string) string {
	name = strings.TrimSuffix(strings.TrimSpace(name), " (LISTEN)")
	if i := strings.LastIndex(name, "]:"); i >= 0 { // IPv6
		return name[i+2:]
	}
	if i := strings.LastIndex(name, ":"); i >= 0 {
		p := name[i+1:]
		if p != "" && p != "*" {
			return p
		}
	}
	return ""
}

// openZed opens dir in Zed, when it is installed.
func openZed(dir string) {
	if _, err := os.Stat("/Applications/Zed.app"); err != nil {
		return
	}
	exec.Command("open", "-a", "Zed", dir).Start()
}
