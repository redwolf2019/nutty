package app

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
)

// confirmApp makes an app with two tabs, a and b, that asks before it
// closes a tab; a is the active one.
func confirmApp() (w *app, a, b *session) {
	w = testApp()
	w.settings.ConfirmCloseTab = "always"
	w.settings.NewTabPosition = "end"
	a = w.newSession("/tmp/nutty-close-a")
	b = w.newSession("/tmp/nutty-close-b")
	w.activate(a)
	return w, a, b
}

// TestCloseTabClosesTheTabAsked answers the Close Tab dialog with a second
// Close Tab after the active tab changed under it, as Next Tab's shortcut
// still does, and checks that the tab the dialog named closes: the second
// Close Tab used to close whichever tab was active, unasked.
func TestCloseTabClosesTheTabAsked(t *testing.T) {
	w, a, b := confirmApp()
	w.closeTab()
	if w.pendingClose != a || len(w.sessions) != 2 {
		t.Fatal("Close Tab should ask about the active tab first")
	}
	w.nextTab()
	w.closeTab()
	if slices.Contains(w.sessions, a) {
		t.Error("the tab the dialog asked about is still open")
	}
	if !slices.Contains(w.sessions, b) {
		t.Error("a tab the dialog did not ask about was closed")
	}
	if w.pendingClose != nil {
		t.Error("the dialog should be gone once it was answered")
	}
}

// TestCloseButtonAsksFirst clicks a tab's close button in the rail and in
// the tab bar and checks that it asks as Close Tab does, where the button
// used to close the tab at once whatever the setting said.
func TestCloseButtonAsksFirst(t *testing.T) {
	for _, layout := range []string{"sidebar-left", "tabs-top"} {
		w, _, b := confirmApp()
		w.settings.WindowLayout = layout
		tt := ui.NewTester(w.view, 1000, 700)
		tt.Frame()
		// The close button shows while the pointer is over its tab.
		r, ok := tt.Find("/tmp/nutty-close-b")
		if !ok {
			t.Fatalf("%s: no tab for the second session", layout)
		}
		tt.Move(r.X+r.W/2, r.Y+r.H/2)
		tt.Frame()
		if err := tt.Click("×"); err != nil {
			t.Fatalf("%s: %v", layout, err)
		}
		tt.Frame()
		if len(w.sessions) != 2 {
			t.Fatalf("%s: the close button closed the tab without asking", layout)
		}
		if w.pendingClose != b {
			t.Fatalf("%s: the close button should ask about its own tab", layout)
		}
		if !tt.HasText(tr("Close Tab?")) {
			t.Errorf("%s: the dialog is not shown", layout)
		}

		// With the question turned off the button closes the tab at once.
		w.pendingClose = nil
		w.settings.ConfirmCloseTab = "never"
		tt.Frame()
		tt.Move(r.X+r.W/2, r.Y+r.H/2)
		tt.Frame()
		if err := tt.Click("×"); err != nil {
			t.Fatalf("%s: %v", layout, err)
		}
		tt.Frame()
		if slices.Contains(w.sessions, b) {
			t.Errorf("%s: the close button should close the tab when nothing asks", layout)
		}
	}
}

// TestClosedTabDropsItsDialogs closes a tab the way its shell's exit does
// while the Close Tab and Rename Tab dialogs are about it, and checks that
// neither is left asking about a tab that is gone.
func TestClosedTabDropsItsDialogs(t *testing.T) {
	w, a, b := confirmApp()
	w.closeTab()
	w.beginRenameSession(a)
	w.closeSession(a)
	if w.pendingClose != nil {
		t.Error("the Close Tab dialog still names the closed tab")
	}
	if w.rename.open || w.rename.target != nil {
		t.Error("the Rename Tab dialog is still open for the closed tab")
	}
	// The next Close Tab asks about the tab that is active now.
	w.closeTab()
	if w.pendingClose != b || len(w.sessions) != 1 {
		t.Error("Close Tab should ask again, about the tab that is left")
	}
}
