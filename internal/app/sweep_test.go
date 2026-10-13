package app

import (
	"testing"
	"time"

	"nutty/internal/i18n"
	"nutty/internal/settings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
	"github.com/smallnest/uix/components/tree"
)

// TestSweepChrome renders the main window in its states — the pane, the
// info sections, the chat panel, the empty state, the dialogs and the
// palette — with Chinese selected and recording on, then fails on the
// strings still lacking a translation. The tooltip strings only draw on a
// hover and are added to the table by hand; everything that renders is
// caught here.
func TestSweepChrome(t *testing.T) {
	i18n.SetLang("chinese")
	i18n.SetRecording(true)
	defer i18n.SetRecording(false)
	registerFonts()

	render := func(name string, fn func(a *app)) {
		a := testApp()
		a.settings = settings.Default()
		a.sideOpen = true
		a.agent = &demoAgent{}
		fn(a)
		tt := ui.NewTester(a.view, 1280, 780)
		tt.SetTitleBar(ui.TitleBar{Height: 31, Left: 60})
		tt.Frame()
	}

	// A pane with a shell, the rail, the title, and the info and chat
	// panels each open, then the empty window.
	render("pane", func(a *app) {
		p := &pane{}
		p.term = scriptTerm(t)
		s := &session{id: 1, title: "sweep", dir: "/Users/smallnest/ai/nutty",
			started: time.Now(), root: &node{term: p}, panes: []*pane{p}}
		a.sessions = []*session{s}
		a.active = 0
		a.activePane = p
	})
	render("info", func(a *app) {
		a.panel = panelInfo
		a.info = infoState{ports: []string{"5173", "8000"}, portsDone: true,
			outline: []string{"git status", "npm run dev"}, outlineDone: true,
			git:   gitState{branch: "main", changes: []gitChange{{" M", "go.mod"}}, done: true},
			files: []tree.Item{{Label: "a.go"}, {Label: "lib", Children: []tree.Item{{Label: "b.go"}}}}, filesDone: true}
	})
	render("info-empty", func(a *app) {
		a.panel = panelInfo
		a.info = infoState{portsDone: true, outlineDone: true, git: gitState{done: true}, filesDone: true}
	})
	render("git-diff", func(a *app) {
		a.panel = panelInfo
		a.info.tab = infoTabGit
		a.info.git = gitState{branch: "main", done: true,
			diffOpen: "go.mod", diffLines: []string{"+module nutty", "-module old"}}
	})
	render("chat", func(a *app) {
		a.panel = panelChat
		a.chat.busy = true
		a.chat.started = time.Now()
	})
	render("empty", func(a *app) {})
	render("rename", func(a *app) {
		a.rename.open = true
		a.rename.buf = "web"
	})
	render("confirm", func(a *app) {
		p := &pane{}
		p.term = scriptTerm(t)
		a.pendingClose = &session{id: 1, title: "web", dir: "/Users/smallnest/ai/nutty",
			root: &node{term: p}, panes: []*pane{p}}
	})
	render("palette", func(a *app) {
		a.paletteOpen = true
	})

	for _, m := range i18n.Untranslated() {
		if !sweepChromeExempt(m) {
			t.Errorf("missing translation for %q", m)
		}
	}
}

// sweepChromeExempt is the chrome half of the settings sweep's exemption
// list: the strings the main window legitimately keeps untranslated.
func sweepChromeExempt(s string) bool {
	switch s {
	case "", "中文", "Nutty", "Otty", "Git", "Nord", "Paper", "Recipes",
		"Shell", "Shift", "Term", "Alt", "Ctrl", "Ctrl+Shift", "Super":
		return true
	}
	for _, n := range settings.OttyThemeNames {
		if s == n {
			return true
		}
	}
	return false
}

// scriptTerm is a terminal that renders an idle prompt without a shell.
func scriptTerm(t *testing.T) *terminal.Terminal {
	t.Helper()
	lt, dt := settings.OttyTermTheme(false), settings.OttyTermTheme(true)
	term, err := terminal.New(terminal.Options{
		Conn:      &scriptConn{data: []byte("\x1b]0;sweep — nutty\x07\x1b[1;32msweep\x1b[0m ❯ ")},
		Font:      (&app{settings: settings.Default()}).termFont(),
		Theme:     lt,
		DarkTheme: dt,
	})
	if err != nil {
		t.Fatal(err)
	}
	return term
}
