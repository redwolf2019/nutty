package app

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"
)

// TestRemovePaneKeepsTheLastPane asks removePane for a session's only
// pane and checks that the tree is left whole: it used to swap the root
// for an empty node, which the next walk of the tree dereferenced. The
// last pane closes with its session, in closePane.
func TestRemovePaneKeepsTheLastPane(t *testing.T) {
	a := testApp()
	s := a.newSession("/tmp/nutty-last-pane")
	only := s.panes[0]
	s.removePane(only)
	if got := s.orderedPanes(); len(got) != 1 || got[0] != only {
		t.Fatalf("the tree lost its only pane: %v", got)
	}
}

// TestTabNamedAsItShows checks that the palette's "Switch to" and the
// Close Tab dialog name a tab as the rail and the tab bar draw it, by its
// directory: they used to name it by the shell's title, which no tab shows.
func TestTabNamedAsItShows(t *testing.T) {
	a := testApp()
	s := a.newSession("/tmp/nutty-named")
	s.title = "vim" // what a program set as the terminal's title

	want := tr("Switch to ") + "/tmp/nutty-named"
	found := false
	for _, it := range a.paletteItems() {
		found = found || it.Label == want
	}
	if !found {
		t.Errorf("the palette has no %q", want)
	}

	a.pendingClose = s
	tt := ui.NewTester(a.view, 1000, 700)
	tt.Frame()
	if msg := fmt.Sprintf(tr("“%s” and its processes will end."), "/tmp/nutty-named"); !tt.HasText(msg) {
		t.Errorf("the Close Tab dialog does not name the tab as it shows, texts: %v", tt.Texts())
	}
}
