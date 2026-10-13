package app

import (
	"testing"

	"nutty/internal/settings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// osc7App is a window with one session in /start whose terminal has read
// an OSC 7 report with the given payload, all of it, before any frame.
func osc7App(t *testing.T, payload string) (*app, *session) {
	t.Helper()
	registerFonts()
	term, err := terminal.New(terminal.Options{
		Conn:  &scriptConn{data: []byte("\x1b]7;" + payload + "\x07")},
		Font:  (&app{settings: settings.Default()}).termFont(),
		Theme: settings.OttyTermTheme(false),
	})
	if err != nil {
		t.Fatal(err)
	}
	<-term.Done()
	p := &pane{term: term}
	s := &session{id: 1, dir: "/start", root: &node{term: p}, panes: []*pane{p}}
	a := &app{settings: settings.Default(), sideOpen: true, sessions: []*session{s}, active: 0, activePane: p}
	return a, s
}

// TestTrackWorkingDirAbsoluteOnly checks that the session takes only an
// absolute path from the shell's report. The report is whatever a program
// in the terminal wrote, and the directory is handed to open and git: one
// that reads as an option, or a relative one, must not become the
// session's directory.
func TestTrackWorkingDirAbsoluteOnly(t *testing.T) {
	for _, payload := range []string{
		"-aCalculator",              // would run as: open -aCalculator
		"relative/dir",              // would resolve against the app's own directory
		"kitty-shell-cwd://h/tmp/x", // not a file URL: the terminal passes it through
	} {
		a, s := osc7App(t, payload)
		a.trackWorkingDir()
		if s.dir != "/start" {
			t.Errorf("OSC 7 %q: dir = %q, want it to stay /start", payload, s.dir)
		}
	}
	// A file URL still moves the session, with or without a host.
	for payload, want := range map[string]string{
		"file://localhost/tmp/elsewhere": "/tmp/elsewhere",
		"file:///tmp/with%20space":       "/tmp/with space",
	} {
		a, s := osc7App(t, payload)
		a.trackWorkingDir()
		if s.dir != want {
			t.Errorf("OSC 7 %q: dir = %q, want %q", payload, s.dir, want)
		}
	}
}

// TestDirDrawnInFirstFrame checks that the frame that finds the shell in a
// new directory draws that directory: the terminal has read the report
// before the frame, so the tab and the title must not show the old one and
// wait for another frame, which nothing asks for.
func TestDirDrawnInFirstFrame(t *testing.T) {
	a, s := osc7App(t, "file://localhost/tmp/elsewhere")
	tt := ui.NewTester(a.view, 1000, 700) // builds the first frame
	if s.dir != "/tmp/elsewhere" {
		t.Fatalf("dir = %q, want /tmp/elsewhere", s.dir)
	}
	if !tt.HasText("/tmp/elsewhere") || tt.HasText("/start") {
		t.Fatalf("the first frame drew the old directory, texts: %v", tt.Texts())
	}
}
