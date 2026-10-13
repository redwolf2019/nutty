package app

import (
	"testing"

	"nutty/internal/settings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// focusApp makes an app whose panes hold real terminals on a connection
// that never ends, so the terminal views take the keyboard focus as they
// do in the window.
func focusApp(t *testing.T) *app {
	t.Helper()
	registerFonts()
	old := newTerminal
	t.Cleanup(func() { newTerminal = old })
	newTerminal = func(o terminal.Options) (*terminal.Terminal, error) {
		o.Conn = &scriptConn{data: []byte("$ "), hold: true}
		return terminal.New(o)
	}
	return &app{settings: settings.Default(), sideOpen: true, agent: &demoAgent{}}
}

// focusFrames draws enough frames for a request for the focus to land.
func focusFrames(tt *ui.Tester) {
	for range 4 {
		tt.Frame()
	}
}

// TestSplitFocusesNewPane checks that Split Right moves the keyboard focus
// to the new pane: the pane that had the focus used to claim activePane
// back as it drew, so the next split and the typing stayed in the old one.
func TestSplitFocusesNewPane(t *testing.T) {
	a := focusApp(t)
	s := a.newSession("/tmp/nutty-focus")
	focused := map[*pane]bool{}
	tt := ui.NewTester(func(c *ui.Context) {
		a.view(c)
		for _, p := range s.panes {
			focused[p] = p.handle.Focused(c)
		}
	}, 1000, 700)
	focusFrames(tt)
	if !focused[s.panes[0]] {
		t.Fatal("the first pane should start with the focus")
	}
	a.splitRight()
	focusFrames(tt)
	if len(s.panes) != 2 {
		t.Fatalf("panes = %d, want 2", len(s.panes))
	}
	if a.activePane != s.panes[1] {
		t.Error("the new pane should be the active one once the split is drawn")
	}
	if !focused[s.panes[1]] || focused[s.panes[0]] {
		t.Errorf("the new pane should have the focus: old=%v new=%v", focused[s.panes[0]], focused[s.panes[1]])
	}
}

// TestTabClickKeepsTerminalFocus clicks a tab in the rail and in the tab
// bar and checks that the terminal still has the keyboard focus: the tab,
// a button, used to take it, so nothing typed after the click reached the
// shell.
func TestTabClickKeepsTerminalFocus(t *testing.T) {
	for _, layout := range []string{"sidebar-left", "tabs-top"} {
		a := focusApp(t)
		a.settings.WindowLayout = layout
		a.newSession("/tmp/nutty-focus-a")
		a.newSession("/tmp/nutty-focus-b")
		tt := ui.NewTester(a.view, 1000, 700)
		focusFrames(tt)
		if !tt.Focused("Terminal") {
			t.Fatalf("%s: the terminal should start with the focus", layout)
		}
		// One tab, then the other: two clicks on the same tab are a double
		// click, which opens the rename dialog.
		for want, tab := range []string{"/tmp/nutty-focus-a", "/tmp/nutty-focus-b"} {
			if err := tt.Click(tab); err != nil {
				t.Fatal(err)
			}
			focusFrames(tt)
			if a.active != want {
				t.Fatalf("%s: the click did not activate tab %d", layout, want)
			}
			if !tt.Focused("Terminal") {
				t.Errorf("%s: a click on a tab took the focus from the terminal", layout)
			}
		}
	}
}

// TestRenameReturnsFocus renames a tab from the keyboard and checks that
// the dialog's field takes what is typed and that the terminal has the
// focus again once the dialog closes.
func TestRenameReturnsFocus(t *testing.T) {
	a := focusApp(t)
	s := a.newSession("/tmp/nutty-focus")
	tt := ui.NewTester(a.view, 1000, 700)
	focusFrames(tt)
	a.beginRename()
	focusFrames(tt)
	tt.Type("web")
	tt.Key(0, ui.KeyEnter)
	focusFrames(tt)
	if a.rename.open || s.name != "web" {
		t.Fatalf("the rename did not commit: open=%v name=%q", a.rename.open, s.name)
	}
	if !tt.Focused("Terminal") {
		t.Error("the terminal did not get the focus back after the rename")
	}
}
