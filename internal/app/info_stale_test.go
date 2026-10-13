package app

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/smallnest/uix/components/tree"
)

// loadedInfo is an info panel that has loaded a session's repository and
// files.
func loadedInfo(tab infoTab) infoState {
	return infoState{
		tab:       tab,
		git:       gitState{done: true, branch: "main", changes: []gitChange{{" M", "old.go"}}, diffOpen: "old.go"},
		files:     []tree.Item{{Label: "old.go"}},
		filesDone: true,
	}
}

// TestLastSessionClearsInfo checks that closing the last session takes its
// repository and its files off the info panel: with no session left there
// is no directory they could describe, and a click on one of the changes
// would wait for a diff that never comes.
func TestLastSessionClearsInfo(t *testing.T) {
	a := testApp()
	s := a.newSession("/tmp/nutty-last")
	a.info = loadedInfo(infoTabGit)
	a.closeSession(s)
	if len(a.sessions) != 0 {
		t.Fatal("the session should be closed")
	}
	if g := a.info.git; g.branch != "" || g.changes != nil || g.diffOpen != "" {
		t.Errorf("the Git panel kept the closed session's repository: %+v", g)
	}
	if a.info.files != nil || a.info.filesDone {
		t.Errorf("the Files panel kept the closed session's files: %v", a.info.files)
	}
}

// TestHiddenDetailsViewLeavesAtOnce checks that a section the settings
// hide while the panel shows it gives way to Info on the next frame, in
// every layout, rather than staying up, without its button, until the
// panel is closed and opened again.
func TestHiddenDetailsViewLeavesAtOnce(t *testing.T) {
	for _, layout := range []string{"sidebar-left", "tabs-top"} {
		a := testApp()
		a.settings.WindowLayout = layout
		a.newSession("/tmp/nutty-hidden")
		a.panel = panelInfo
		a.info = loadedInfo(infoTabGit)
		tt := ui.NewTester(a.view, 1000, 700)
		tt.Frame()
		if !tt.HasText("old.go") {
			t.Fatalf("%s: the Git section should be showing, texts: %v", layout, tt.Texts())
		}
		for i := range a.settings.DetailsViews {
			if a.settings.DetailsViews[i].ID == "git" {
				a.settings.DetailsViews[i].Enabled = false
			}
		}
		tt.Frame()
		if a.info.tab != infoTabInfo || tt.HasText("old.go") {
			t.Errorf("%s: the hidden Git section is still open (tab %d), texts: %v", layout, a.info.tab, tt.Texts())
		}
	}
}

// TestClosedPanelLoadsNothing checks that a change of directory, or of the
// active tab, marks the Git and Files sections stale without loading them
// while the panel is closed. The loaders need a window, so the test sees
// the load through what loadInfoTab does first: it moves a section the
// settings hide to Info, which a panel that loads nothing leaves alone.
func TestClosedPanelLoadsNothing(t *testing.T) {
	a := testApp()
	a.newSession("/tmp/nutty-closed")
	for i := range a.settings.DetailsViews {
		if a.settings.DetailsViews[i].ID == "git" {
			a.settings.DetailsViews[i].Enabled = false
		}
	}
	a.info = loadedInfo(infoTabGit)
	a.infoDirChanged()
	if a.info.git.done || a.info.filesDone {
		t.Fatal("the sections should be stale after the directory changed")
	}
	if a.info.tab != infoTabGit {
		t.Error("the closed panel loaded its open section")
	}
	// The open panel does load it.
	a.panel = panelInfo
	a.infoDirChanged()
	if a.info.tab != infoTabInfo {
		t.Error("the open panel did not load its open section")
	}
}
