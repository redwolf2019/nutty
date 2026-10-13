package app

import (
	"testing"

	"nutty/internal/settings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
	"github.com/smallnest/uix/components/tree"
)

// TestTrackWorkingDir renders a terminal whose shell reports a directory
// (OSC 7) and checks that the session takes it, so the tab, the title and
// the info panel show where the shell is.
func TestTrackWorkingDir(t *testing.T) {
	registerFonts()
	term, err := terminal.New(terminal.Options{
		Conn:  &scriptConn{data: []byte("\x1b]7;file://localhost/tmp/elsewhere\x07")},
		Font:  (&app{settings: settings.Default()}).termFont(),
		Theme: settings.OttyTermTheme(false),
	})
	if err != nil {
		t.Fatal(err)
	}
	// The terminal reads its output on a goroutine of its own: wait for it
	// to reach the end of the script, or the frame can come before the
	// directory is known.
	<-term.Done()
	p := &pane{term: term}
	s := &session{id: 1, dir: "/start", root: &node{term: p}, panes: []*pane{p}}
	a := &app{settings: settings.Default(), sessions: []*session{s}, active: 0, activePane: p}
	ui.NewTester(a.view, 800, 400).Frame()
	a.trackWorkingDir()
	if s.dir != "/tmp/elsewhere" {
		t.Fatalf("dir = %q, want /tmp/elsewhere", s.dir)
	}
}

// TestInfoFollowsDir renders a terminal whose shell reports a new directory
// and checks that the Git and Files panels are marked stale, so they reload
// for the new directory instead of showing the previous one's state.
func TestInfoFollowsDir(t *testing.T) {
	registerFonts()
	term, err := terminal.New(terminal.Options{
		Conn:  &scriptConn{data: []byte("\x1b]7;file://localhost/tmp/elsewhere\x07")},
		Font:  (&app{settings: settings.Default()}).termFont(),
		Theme: settings.OttyTermTheme(false),
	})
	if err != nil {
		t.Fatal(err)
	}
	<-term.Done() // the script is read before the frame, as above
	p := &pane{term: term}
	s := &session{id: 1, dir: "/start", root: &node{term: p}, panes: []*pane{p}}
	a := &app{settings: settings.Default(), sessions: []*session{s}, active: 0, activePane: p}
	a.info = infoState{
		tab:       infoTabGit,
		git:       gitState{done: true, branch: "old", changes: []gitChange{{" M", "old.go"}}},
		filesDone: true,
		files:     []tree.Item{{Label: "old"}},
	}
	ui.NewTester(a.view, 800, 400).Frame()
	if a.info.git.done {
		t.Fatal("git panel kept its state after the directory changed")
	}
	if a.info.filesDone {
		t.Fatal("files panel kept its state after the directory changed")
	}
}
