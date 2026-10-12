package main

import (
	"image/png"
	"io"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
	"github.com/smallnest/uix/components/tree"
)

// scriptConn is a read-only terminal source for previews: the terminal
// shows script as the program's output, and typed input goes nowhere.
type scriptConn struct {
	data []byte
	off  int
}

func (c *scriptConn) Read(p []byte) (int, error) {
	if c.off >= len(c.data) {
		return 0, io.EOF
	}
	n := copy(p, c.data[c.off:])
	c.off += n
	return n, nil
}
func (c *scriptConn) Write(p []byte) (int, error) { return len(p), nil }
func (c *scriptConn) Close() error                { return nil }

const previewScript = "\x1b]0;preview — nutty\x07" +
	"\x1b[38;5;245mLast login: Sat Oct 12 09:15:00 on ttys001\r\n" +
	"\x1b[1;32mnutty\x1b[0m \x1b[1;34m~/ai/nutty\x1b[0m \x1b[1m❯\x1b[0m " +
	"ls\r\n" +
	"\x1b[0;38;5;250massets\x1b[0m   \x1b[1;34mresources\x1b[0m   " +
	"\x1b[0;38;5;250mmain.go\x1b[0m   \x1b[1;34msession.go\x1b[0m\r\n" +
	"\x1b[1;32mnutty\x1b[0m \x1b[1;34m~/ai/nutty\x1b[0m \x1b[1m❯\x1b[0m "

// TestSnapshot renders the window to /tmp/nutty-preview.png for a visual
// check without a window: NUTTY_SNAPSHOT=1 go test -run TestSnapshot.
func TestSnapshot(t *testing.T) {
	if os.Getenv("NUTTY_SNAPSHOT") == "" {
		t.Skip("set NUTTY_SNAPSHOT=1 to render the preview")
	}
	registerFonts()
	lt, dt := ottyTermTheme(false), ottyTermTheme(true)
	term, err := terminal.New(terminal.Options{
		Conn:      &scriptConn{data: []byte(previewScript)},
		Font:      (&app{settings: defaultSettings()}).termFont(),
		Theme:     lt,
		DarkTheme: dt,
	})
	if err != nil {
		t.Fatal(err)
	}
	p := &pane{term: term}
	s := &session{
		id: 1, title: "preview", dir: "/Users/smallnest/ai/nutty",
		started: time.Now().Add(-73 * time.Second),
		root:    &node{term: p},
		panes:   []*pane{p},
	}
	a := &app{settings: defaultSettings(), sideOpen: true, agent: &demoAgent{}}
	a.sessions = []*session{s}
	a.active = 0
	a.activePane = p
	a.sel = sessionID(1)
	a.panel = panelChat
	a.info = infoState{ports: []string{"5173", "8000"}, portsDone: true}
	a.chat.msgs = []chatMsg{
		{user: true, text: "把会话渲染成 markdown 看看"},
		{text: "## Markdown 渲染\n\n这是 **onevra** 渲染的列表：\n\n- 第一项\n- `行内代码`\n\n```go\nfmt.Println(\"hi\")\n```\n\n> 引用块"},
	}
	tt := ui.NewTester(a.view, 1280, 780)
	// The room the traffic lights take in a real window with the Y:7
	// position: the strip's content lines up with them.
	tt.SetTitleBar(ui.TitleBar{Height: 31, Left: 60})
	for panel, base := range map[panelMode]string{
		panelChat: "preview",
		panelInfo: "preview-info",
		panelNone: "preview-none",
	} {
		a.panel = panel
		for name, dark := range map[string]bool{base: false, base + "-dark": true} {
			tt.SetDark(dark)
			f, err := os.Create("/tmp/nutty-" + name + ".png")
			if err != nil {
				t.Fatal(err)
			}
			if err := png.Encode(f, tt.Image()); err != nil {
				f.Close()
				t.Fatal(err)
			}
			f.Close()
			t.Log("wrote /tmp/nutty-" + name + ".png")
		}
	}
}

// TestStripToggles renders the title strip's two toggles — the one that
// hides the sessions and the one that opens the info panel — in every
// layout, at the title's height: NUTTY_SNAPSHOT=1 go test -run
// TestStripToggles.
func TestStripToggles(t *testing.T) {
	if os.Getenv("NUTTY_SNAPSHOT") == "" {
		t.Skip("set NUTTY_SNAPSHOT=1 to render the preview")
	}
	registerFonts()
	lt, dt := ottyTermTheme(false), ottyTermTheme(true)
	term, err := terminal.New(terminal.Options{
		Conn:      &scriptConn{data: []byte(previewScript)},
		Font:      (&app{settings: defaultSettings()}).termFont(),
		Theme:     lt,
		DarkTheme: dt,
	})
	if err != nil {
		t.Fatal(err)
	}
	p := &pane{term: term}
	s := &session{
		id: 1, title: "preview", dir: "/Users/smallnest/ai/nutty",
		started: time.Now().Add(-73 * time.Second),
		root:    &node{term: p},
		panes:   []*pane{p},
	}
	render := func(name string, layout string) {
		a := &app{settings: defaultSettings(), sideOpen: true, agent: &demoAgent{}}
		a.settings.WindowLayout = layout
		a.sessions = []*session{s}
		a.active = 0
		a.activePane = p
		a.sel = sessionID(1)
		tt := ui.NewTester(a.view, 1280, 780)
		tt.SetTitleBar(ui.TitleBar{Height: 31, Left: 60})
		f, err := os.Create("/tmp/nutty-" + name + ".png")
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, tt.Image()); err != nil {
			f.Close()
			t.Fatal(err)
		}
		f.Close()
		t.Log("wrote /tmp/nutty-" + name + ".png")
	}
	render("strip-toggles-sidebar", "sidebar-left")
	render("strip-toggles-tabs-top", "tabs-top")
	render("strip-toggles-tabs-bottom", "tabs-bottom")
}

// TestInfoTabs renders the right panel's three sections that load real
// data, with sample state seeded in place of a live session: NUTTY_SNAPSHOT=1
// go test -run TestInfoTabs. The loaders no-op without a window, so the
// renders show exactly what the seeded data draws.
func TestInfoTabs(t *testing.T) {
	if os.Getenv("NUTTY_SNAPSHOT") == "" {
		t.Skip("set NUTTY_SNAPSHOT=1 to render the preview")
	}
	registerFonts()
	lt, dt := ottyTermTheme(false), ottyTermTheme(true)
	term, err := terminal.New(terminal.Options{
		Conn:      &scriptConn{data: []byte(previewScript)},
		Font:      (&app{settings: defaultSettings()}).termFont(),
		Theme:     lt,
		DarkTheme: dt,
	})
	if err != nil {
		t.Fatal(err)
	}
	p := &pane{term: term}
	s := &session{
		id: 1, title: "preview", dir: "/Users/smallnest/ai/nutty",
		started: time.Now().Add(-73 * time.Second),
		root:    &node{term: p},
		panes:   []*pane{p},
	}
	render := func(name string, a *app, dark bool) {
		tt := ui.NewTester(a.view, 1280, 780)
		tt.SetTitleBar(ui.TitleBar{Height: 31, Left: 60})
		tt.SetDark(dark)
		// Rest the pointer over the toolbar so its icons show, as the site
		// shots do.
		tt.Move(1170, 15)
		f, err := os.Create("/tmp/nutty-" + name + ".png")
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, tt.Image()); err != nil {
			f.Close()
			t.Fatal(err)
		}
		f.Close()
		t.Log("wrote /tmp/nutty-" + name + ".png")
	}
	base := func() *app {
		a := &app{settings: defaultSettings(), sideOpen: true, agent: &demoAgent{}}
		a.sessions = []*session{s}
		a.active = 0
		a.activePane = p
		a.sel = sessionID(1)
		a.panel = panelInfo
		return a
	}
	open := true
	render("outline", func() *app {
		a := base()
		a.info = infoState{
			tab:         infoTabOutline,
			outline:     []string{"git commit -m \"outline marks\"", "npm run dev", "ls -la", "cd web-app"},
			outlineDone: true,
		}
		return a
	}(), false)
	render("git", func() *app {
		a := base()
		a.info = infoState{
			tab: infoTabGit,
			git: gitState{
				branch:  "main",
				changes: []gitChange{{" M", "session.go"}, {"M ", "topbar.go"}, {"??", "settings.go"}, {"A ", "info.go"}},
				done:    true,
			},
		}
		return a
	}(), false)
	render("git-diff", func() *app {
		a := base()
		a.info = infoState{
			tab: infoTabGit,
			git: gitState{
				branch:  "main",
				changes: []gitChange{{" M", "session.go"}, {"??", "settings.go"}},
				done:    true,
				diffOpen: "session.go",
				diffLines: []string{
					"--- a/session.go",
					"+++ b/session.go",
					"@@ -1,4 +1,5 @@",
					" func (a *app) renderPane(c *ui.Context) {",
					"+	a.win.SetTitle(\"rendered\")",
					"-	return nil",
				},
			},
		}
		return a
	}(), false)
	render("files", func() *app {
		a := base()
		a.info = infoState{
			tab: infoTabFiles,
			files: []tree.Item{
				{Label: "assets", Open: &open, Children: []tree.Item{
					{Label: "nutty.svg"},
					{Label: "JetBrainsMonoNerdFontMono.ttf"},
				}},
				{Label: "main.go"},
				{Label: "session.go"},
				{Label: "themes.go"},
			},
			filesDone: true,
		}
		return a
	}(), false)
}

// TestLayoutSnapshot renders the window in each Window Layout setting for
// a visual check, with the info panel closed and open. The open panel must
// reach the top of the window in every layout.
func TestLayoutSnapshot(t *testing.T) {
	if os.Getenv("NUTTY_SNAPSHOT") == "" {
		t.Skip("set NUTTY_SNAPSHOT=1 to render the preview")
	}
	for _, layout := range []string{"sidebar-left", "tabs-top", "tabs-bottom"} {
		for _, open := range []string{"", "-panel"} {
			a := testApp()
			a.settings.WindowLayout = layout
			a.newSession("/tmp/nutty-a")
			a.newSession("/tmp/nutty-b")
			if open != "" {
				a.setPanel(panelInfo)
			}
			tt := ui.NewTester(a.view, 1000, 700)
			tt.Frame()
			f, err := os.Create("/tmp/nutty-layout-" + layout + open + ".png")
			if err != nil {
				t.Fatal(err)
			}
			if err := png.Encode(f, tt.Image()); err != nil {
				f.Close()
				t.Fatal(err)
			}
			f.Close()
			t.Log("wrote /tmp/nutty-layout-" + layout + open + ".png")
		}
	}
}

// TestSettingsSnapshot renders every settings category for a visual check.
func TestSettingsSnapshot(t *testing.T) {
	if os.Getenv("NUTTY_SNAPSHOT") == "" {
		t.Skip("set NUTTY_SNAPSHOT=1 to render the preview")
	}
	a := &app{settings: defaultSettings(), sideOpen: true, agent: &demoAgent{}}
	for _, cat := range settingsCategories {
		sw := &settingsWindow{a: a, s: a.settings, last: *a.settings, page: cat.id}
		tt := ui.NewTester(sw.view, 760, 540)
		tt.Frame()
		f, err := os.Create("/tmp/nutty-settings-" + cat.id + ".png")
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, tt.Image()); err != nil {
			f.Close()
			t.Fatal(err)
		}
		f.Close()
		t.Log("wrote /tmp/nutty-settings-" + cat.id + ".png")
	}
}
