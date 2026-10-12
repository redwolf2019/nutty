package main

// TestSiteShots renders the screenshots the site in docs/ shows, with an
// English session for the pages. NUTTY_SNAPSHOT=1 go test -run TestSiteShots -v .
import (
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// siteScript is a lively English session: a git status, then a dev server.
const siteScript = "\x1b]0;nutty — web-app\x07" +
	"\x1b[38;5;245mLast login: Sun Oct 12 09:15:00 on ttys002\r\n" +
	"\x1b[1;32mnutty\x1b[0m \x1b[1;34m~/web-app\x1b[0m \x1b[1m❯\x1b[0m git status\r\n" +
	"\x1b[0mOn branch \x1b[1;36mmain\x1b[0m\r\n" +
	"\x1b[0mnothing to commit, working tree clean\r\n" +
	"\x1b[1;32mnutty\x1b[0m \x1b[1;34m~/web-app\x1b[0m \x1b[1m❯\x1b[0m npm run dev\r\n" +
	"\x1b[0;38;5;250m  VITE v6.0.0  ready in 321 ms\r\n" +
	"\x1b[0m  \x1b[1;36m➜\x1b[0m  Local:   http://localhost:5173/\r\n" +
	"\x1b[0m  \x1b[1;36m➜\x1b[0m  Network: use --host to expose\r\n" +
	"\x1b[1;32mnutty\x1b[0m \x1b[1;34m~/web-app\x1b[0m \x1b[1m❯\x1b[0m "

// siteChat is the agent exchange the hero screenshot shows.
var siteChat = []chatMsg{
	{user: true, text: "Watch the ports while I code"},
	{text: "## Ship faster\n\nI’ll keep an eye on the network while you type:\n\n- `5173` — Vite dev server\n- `8000` — mock API\n\n```sh\ncurl localhost:8000/health\n```\n\n> 4 checks passed"},
}

func TestSiteShots(t *testing.T) {
	if os.Getenv("NUTTY_SNAPSHOT") == "" {
		t.Skip("set NUTTY_SNAPSHOT=1 to render the screenshots")
	}
	render := func(name string, panel panelMode, dark bool) {
		registerFonts()
		lt, dt := ottyTermTheme(false), ottyTermTheme(true)
		term, err := terminal.New(terminal.Options{
			Conn:      &scriptConn{data: []byte(siteScript)},
			Font:      (&app{settings: defaultSettings()}).termFont(),
			Theme:     lt,
			DarkTheme: dt,
		})
		if err != nil {
			t.Fatal(err)
		}
		p := &pane{term: term}
		s := &session{
			id: 1, title: "preview", dir: "/Users/smallnest/web-app",
			started: time.Now().Add(-73 * time.Second),
			root:    &node{term: p},
			panes:   []*pane{p},
		}
		a := &app{settings: defaultSettings(), sideOpen: true, agent: &demoAgent{}}
		a.sessions = []*session{s}
		a.active = 0
		a.activePane = p
		a.sel = sessionID(1)
		a.panel = panel
		a.info = infoState{ports: []string{"5173", "8000"}, portsDone: true}
		a.chat.msgs = siteChat
		tt := ui.NewTester(a.view, 1280, 780)
		tt.SetTitleBar(ui.TitleBar{Height: 31, Left: 60})
		tt.SetDark(dark)
		// The toolbars appear on hover, so rest the pointer over them for
		// the shot: the chat column is 360 wide, the info column 220.
		if panel == panelChat {
			tt.Move(1100, 15)
		} else {
			tt.Move(1170, 15)
		}
		f, err := os.Create("/tmp/nutty-site-" + name + ".png")
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, tt.Image()); err != nil {
			f.Close()
			t.Fatal(err)
		}
		f.Close()
		t.Log("wrote /tmp/nutty-site-" + name + ".png")
	}
	render("hero", panelChat, false)
	render("hero-dark", panelChat, true)
	render("ports", panelInfo, false)
}
