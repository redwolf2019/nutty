package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"nutty/internal/settings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
	"github.com/smallnest/uix/components/split"
)

// session is one tab: a running shell, split into panes.
type session struct {
	id      int
	name    string // the user's name; empty follows the shell's title
	title   string // the latest title the shell set
	dir     string
	started time.Time // when the shell started, for the info panel
	root    *node
	panes   []*pane
}

// displayTitle is the session's title: the user's name when set, else the
// shell's title, else the shell's name. The tabs show tabLabel instead.
func (s *session) displayTitle() string {
	if s.name != "" {
		return s.name
	}
	if s.title != "" {
		return s.title
	}
	return shellName()
}

// tabLabel is what the rail and the tab bar draw on a tab: the user's
// name when the tab was renamed, else the directory the shell runs in.
// A renamed tab keeps its name when the shell changes directory, as
// otty pins the name the user gives a tab.
func tabLabel(s *session) string {
	if s.name != "" {
		return s.name
	}
	return homeShown(s.dir)
}

// node is the pane tree of a session: a terminal (leaf) or a split.
type node struct {
	term  *pane
	split *splitNode
}

// splitNode is an internal node: two children with a divider.
type splitNode struct {
	vertical bool
	size     *float32
	a, b     *node
}

// pane is a terminal with the handle that gives it focus.
type pane struct {
	term   *terminal.Terminal
	handle ui.Handle
	dead   bool
}

// newTerminal is terminal.New, a seam for tests that must not start a
// shell.
var newTerminal = terminal.New

var sessionIDs int

// newSession opens a session in dir, starting a shell there.
func (a *app) newSession(dir string) *session {
	sessionIDs++
	s := &session{id: sessionIDs, dir: dir, started: time.Now()}
	if s.dir == "" {
		s.dir = homeDir()
	}
	p := &pane{}
	p.term = a.newTerm(s, p)
	s.root = &node{term: p}
	s.panes = []*pane{p}
	// Auto and after-current open the new tab beside the active one, as
	// otty's New Tab Position setting says; End appends it.
	if (a.settings.NewTabPosition == "auto" || a.settings.NewTabPosition == "after-current") &&
		a.active < len(a.sessions) {
		a.sessions = append(a.sessions[:a.active+1],
			append([]*session{s}, a.sessions[a.active+1:]...)...)
	} else {
		a.sessions = append(a.sessions, s)
	}
	a.activate(s)
	return s
}

// newTerm starts a terminal running a shell in the session's directory.
// Its callbacks run on the terminal's goroutine, so they cross to the UI
// thread before touching app state. p is the pane that shows the
// terminal, whose dead flag the exit callback sets.
func (a *app) newTerm(s *session, p *pane) *terminal.Terminal {
	theme, darkTheme := a.termThemes()
	dir := s.dir
	command := a.settings.Command
	env := a.settings.EnvList()
	// The default profile overrides what new sessions run and where, and
	// may pick its own themes, as otty's profiles do.
	if pr := a.settings.Profile(); pr != nil {
		if pr.Command != "" {
			command = pr.Command
		}
		if pr.CWD != "" {
			dir = pr.CWD
		}
		env = append(append([]string{}, env...), settings.ValidEnv(pr.Env)...)
		if pr.Theme != "" || pr.ThemeDark != "" {
			light := settings.ThemeFor(pr.Theme, false)
			dark := settings.ThemeFor(pr.ThemeDark, true)
			if pr.ThemeDark == "" {
				dark = settings.ThemeFor(pr.Theme, true)
			}
			theme, darkTheme = light, dark
		}
	}
	term, err := newTerminal(terminal.Options{
		Command:     settings.ShellArgs(command),
		Dir:         dir,
		Env:         env,
		Font:        a.termFont(),
		Theme:       theme,
		DarkTheme:   darkTheme,
		Cursor:      a.cursorStyle(),
		NoBlink:     a.settings.CursorBlinkStyle == "default_off" || a.settings.CursorBlinkStyle == "always_off",
		Scrollback:  a.scrollbackBytes(),
		OptionAsAlt: a.optionAsAlt(),
		Transparent: a.settings.TerminalOpacity < 1,
		OnTitle: func(title string) {
			a.win.Update(func() {
				s.title = title
			})
		},
		OnExit: func(code int) {
			a.win.Update(func() {
				p.dead = true
				if sessionAllDead(s) {
					a.closeSession(s)
				}
			})
		},
		OpenLink: a.openLink,
	})
	if err != nil {
		// Without its library the terminal cannot run; show the pane as
		// closed rather than crash.
		return nil
	}
	return term
}

// renderNode draws the pane tree of a session.
func (a *app) renderNode(c *ui.Context, n *node) {
	if n.split != nil {
		sn := n.split
		split.Split(c, split.Props{
			Size:     sn.size,
			Vertical: sn.vertical,
			First:    func(c *ui.Context) { a.renderNode(c, sn.a) },
			Second:   func(c *ui.Context) { a.renderNode(c, sn.b) },
		}).Fill()
		return
	}
	a.renderPane(c, n.term)
}

// renderPane draws one terminal, filling the pane with no ring around it —
// otty draws none, and a ring around a terminal reads as a border.
func (a *app) renderPane(c *ui.Context, p *pane) {
	on := p.handle.Focused(c)
	if on {
		a.activePane = p
	}
	ui.Box(c).Fill().Clip().Children(func() {
		if p.term == nil {
			t := c.Theme()
			ui.Column(c).Fill().Center().Gap(8).Children(func() {
				ui.Text(c, tr("Terminal unavailable")).TextColor(t.TextMuted)
				ui.Text(c, tr("the terminal library could not be loaded")).TextColor(t.TextMuted).FontSize(11)
			})
			return
		}
		tv := terminal.View(c, p.term)
		if p == a.activePane && a.panel == panelNone && !a.paletteOpen {
			tv.AutoFocus()
		}
		tv.Fill().Bind(&p.handle)
	})
}

// splitPane splits p with a new pane beside it and focuses the new pane.
func (a *app) splitPane(p *pane, vertical bool) {
	if p == nil {
		p = a.activePane
		if p == nil {
			return
		}
	}
	s := a.sessionOf(p)
	if s == nil {
		return
	}
	np := &pane{}
	np.term = a.newTerm(s, np)
	size := a.defaultSplitSize(vertical)
	split := &node{split: &splitNode{
		vertical: vertical,
		size:     &size,
		a:        nil,
		b:        &node{term: np},
	}}
	if s.root.term == p {
		split.split.a = s.root
		s.root = split
	} else {
		s.replaceChild(p, func(n *node) *node {
			split.split.a = n
			return split
		})
	}
	s.panes = append(s.panes, np)
	a.activePane = np
}

// defaultSplitSize puts the divider near the middle of the window.
func (a *app) defaultSplitSize(vertical bool) float32 {
	if a.win == nil {
		return 300
	}
	w, h := a.win.ContentSize()
	if vertical {
		if h > 80 {
			return float32(h)/2 - 20
		}
		return 200
	}
	if w > 80 {
		return float32(w)/2 - 20
	}
	return 300
}

// replaceChild replaces the node holding p (as a direct child of a split)
// with the result of fn.
func (s *session) replaceChild(p *pane, fn func(*node) *node) bool {
	var rec func(n *node) bool
	rec = func(n *node) bool {
		if n.split == nil {
			return false
		}
		if n.split.a.term == p {
			n.split.a = fn(n.split.a)
			return true
		}
		if n.split.b.term == p {
			n.split.b = fn(n.split.b)
			return true
		}
		return rec(n.split.a) || rec(n.split.b)
	}
	return rec(s.root)
}

// closePane closes p, collapsing its split to the sibling, or closing the
// whole session when it is the last pane.
func (a *app) closePane(p *pane) {
	if p == nil {
		return
	}
	s := a.sessionOf(p)
	if s == nil {
		return
	}
	if len(s.panes) == 1 {
		a.closeSession(s)
		return
	}
	s.removePane(p)
	if p.term != nil {
		_ = p.term.Close()
	}
	if a.activePane == p {
		a.activePane = s.panes[0]
	}
}

// removePane removes p from the tree, collapsing its split to the sibling.
func (s *session) removePane(p *pane) {
	var rec func(n *node) bool
	rec = func(n *node) bool {
		if n.split == nil {
			return false
		}
		if n.split.a.term == p {
			*n = *n.split.b
			return true
		}
		if n.split.b.term == p {
			*n = *n.split.a
			return true
		}
		return rec(n.split.a) || rec(n.split.b)
	}
	rec(s.root)
	s.panes = collectPanes(s.root)
}

// orderedPanes lists the panes left to right, top to bottom, for focus
// navigation.
func (s *session) orderedPanes() []*pane {
	return collectPanes(s.root)
}

func collectPanes(n *node) []*pane {
	if n == nil {
		return nil
	}
	if n.term != nil {
		return []*pane{n.term}
	}
	return append(collectPanes(n.split.a), collectPanes(n.split.b)...)
}

// sessionAllDead reports whether every pane of the session has exited.
func sessionAllDead(s *session) bool {
	if s == nil {
		return true
	}
	for _, p := range s.panes {
		if !p.dead {
			return false
		}
	}
	return true
}

// openLink opens a file path the terminal printed, reporting whether it
// did; URLs open with the system on their own.
func (a *app) openLink(link string) bool {
	link = strings.TrimPrefix(link, "file://")
	if _, err := os.Stat(link); err == nil {
		_ = exec.Command("open", link).Start()
		return true
	}
	return false
}

func homeDir() string {
	d, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return d
}

func shellName() string {
	sh := os.Getenv("SHELL")
	if sh == "" {
		sh = "/bin/zsh"
	}
	return filepath.Base(sh)
}

// sessionID is the sidebar's selection id of a session.
func sessionID(id int) string {
	return "session:" + strconv.Itoa(id)
}
