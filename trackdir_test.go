package main

import (
	"testing"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// TestTrackWorkingDir renders a terminal whose shell reports a directory
// (OSC 7) and checks that the session takes it, so the tab, the title and
// the info panel show where the shell is.
func TestTrackWorkingDir(t *testing.T) {
	registerFonts()
	term, err := terminal.New(terminal.Options{
		Conn:  &scriptConn{data: []byte("\x1b]7;file://localhost/tmp/elsewhere\x07")},
		Font:  (&app{settings: defaultSettings()}).termFont(),
		Theme: ottyTermTheme(false),
	})
	if err != nil {
		t.Fatal(err)
	}
	p := &pane{term: term}
	s := &session{id: 1, dir: "/start", root: &node{term: p}, panes: []*pane{p}}
	a := &app{settings: defaultSettings(), sessions: []*session{s}, active: 0, activePane: p}
	ui.NewTester(a.view, 800, 400).Frame()
	a.trackWorkingDir()
	if s.dir != "/tmp/elsewhere" {
		t.Fatalf("dir = %q, want /tmp/elsewhere", s.dir)
	}
}
