package app

import (
	"slices"
	"testing"

	"nutty/internal/settings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// TestExitedPaneCloses ends the shell of the middle pane of three and
// checks that the pane leaves the tree, that the focus still cycles through
// the panes that remain, and that the tab closes with its last pane. The
// pane used to stay, marked dead: Focus Next Pane stopped in front of it,
// and nothing but closing the tab removed it.
func TestExitedPaneCloses(t *testing.T) {
	a := testApp()
	s := a.newSession("/tmp/nutty-exit")
	first := s.panes[0]
	a.splitPane(first, false)
	middle := s.panes[1]
	a.splitPane(middle, false)
	last := s.panes[2]

	a.activePane = middle
	a.paneExited(middle)
	if got := s.orderedPanes(); len(got) != 2 || slices.Contains(got, middle) {
		t.Fatalf("the exited pane is still in the tree: %d panes", len(got))
	}
	if len(s.panes) != 2 {
		t.Fatalf("panes = %d, want 2", len(s.panes))
	}
	if a.activePane == middle || !slices.Contains(s.panes, a.activePane) {
		t.Fatal("the active pane should be one that is still open")
	}

	// The focus goes round the two panes left, both ways.
	a.activePane = first
	a.focusNextPane()
	if a.activePane != last {
		t.Fatal("Focus Next Pane did not reach the pane after the one that exited")
	}
	a.focusNextPane()
	if a.activePane != first {
		t.Fatal("Focus Next Pane did not wrap around")
	}
	a.focusPrevPane()
	if a.activePane != last {
		t.Fatal("Focus Previous Pane did not wrap around")
	}

	// The tab closes once its last shell ends.
	a.paneExited(first)
	if len(a.sessions) != 1 || len(s.panes) != 1 {
		t.Fatalf("sessions = %d, panes = %d, want 1 and 1", len(a.sessions), len(s.panes))
	}
	a.paneExited(last)
	if len(a.sessions) != 0 {
		t.Fatal("the tab should close with its last pane")
	}
}

// TestExitedPaneHandsFocusOn ends the shell of the pane with the keyboard
// focus and checks that a pane still open has the focus afterwards, so the
// next key reaches a shell without a click.
func TestExitedPaneHandsFocusOn(t *testing.T) {
	registerFonts()
	old := newTerminal
	t.Cleanup(func() { newTerminal = old })
	// Real terminals, on a connection that never ends: only a terminal
	// view takes the keyboard focus.
	newTerminal = func(o terminal.Options) (*terminal.Terminal, error) {
		o.Conn = &scriptConn{data: []byte("$ "), hold: true}
		return terminal.New(o)
	}
	a := &app{settings: settings.Default(), sideOpen: true, agent: &demoAgent{}}
	s := a.newSession("/tmp/nutty-exit")
	tt := ui.NewTester(a.view, 1000, 700)
	frames := func() {
		for range 4 {
			tt.Frame()
		}
	}
	frames()
	a.splitPane(s.panes[0], false)
	frames()
	if !tt.Focused("Terminal") {
		t.Fatal("a terminal should have the focus before the exit")
	}
	a.paneExited(a.activePane)
	frames()
	if len(s.panes) != 1 {
		t.Fatalf("panes = %d, want 1", len(s.panes))
	}
	if !tt.Focused("Terminal") {
		t.Error("no terminal has the focus after the focused pane exited")
	}
}
