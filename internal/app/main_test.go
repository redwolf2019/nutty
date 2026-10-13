package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nutty/internal/settings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// testApp makes an app whose terminals never start a shell.
func testApp() *app {
	newTerminal = func(terminal.Options) (*terminal.Terminal, error) {
		return nil, errors.New("stub: no terminal in tests")
	}
	return &app{settings: settings.Default(), sideOpen: true, agent: &demoAgent{}}
}

// TestWindowLayouts checks that the window draws in each of the three
// Window Layout settings, with a tab for every session.
func TestWindowLayouts(t *testing.T) {
	for _, layout := range []string{"sidebar-left", "tabs-top", "tabs-bottom"} {
		a := testApp()
		a.settings.WindowLayout = layout
		a.newSession("/tmp/nutty-a")
		a.newSession("/tmp/nutty-b")
		tt := ui.NewTester(a.view, 1000, 700)
		tt.Frame()
		if !tt.HasText("/tmp/nutty-b") {
			t.Errorf("%s: missing a session tab", layout)
		}
	}
}

func TestPaneTree(t *testing.T) {
	a := testApp()
	s := a.newSession("/tmp/nutty-test")
	p := s.root.term

	a.splitPane(p, false)
	if len(s.panes) != 2 {
		t.Fatalf("panes = %d, want 2", len(s.panes))
	}
	if s.root.split == nil || s.root.split.vertical {
		t.Fatal("expected a horizontal split at the root")
	}
	if s.root.split.a.term != p {
		t.Fatal("the split should keep the original pane on the left")
	}
	if a.activePane != s.panes[1] {
		t.Fatal("the new pane should take the focus")
	}

	// Split the new pane down: a vertical split inside the right side.
	np := s.panes[1]
	a.splitPane(np, true)
	if len(s.panes) != 3 {
		t.Fatalf("panes = %d, want 3", len(s.panes))
	}
	right := s.root.split.b
	if right.split == nil || !right.split.vertical {
		t.Fatal("expected a vertical split on the right")
	}
	if got := s.orderedPanes(); len(got) != 3 || got[1] != np {
		t.Fatalf("ordered panes wrong: %v", got)
	}

	// Focus navigation wraps around.
	a.activePane = np
	a.focusRelative(1)
	if a.activePane != right.split.b.term {
		t.Fatal("focus next should move to the pane below")
	}

	// Closing the middle pane collapses the split to its sibling.
	a.closePane(np)
	if len(s.panes) != 2 {
		t.Fatalf("panes = %d, want 2 after close", len(s.panes))
	}
	if s.root.split.b.term == nil {
		t.Fatal("the right side should be a leaf again")
	}

	// Closing the last panes closes the session.
	a.closePane(s.root.split.b.term)
	if len(s.panes) != 1 {
		t.Fatalf("panes = %d, want 1", len(s.panes))
	}
	a.closePane(s.root.term)
	if len(a.sessions) != 0 {
		t.Fatal("the session should close with its last pane")
	}
}

func TestDemoAgent(t *testing.T) {
	a := &demoAgent{}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Hello\n\n- a\n- b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for prompt, want := range map[string]string{
		"/help":   "我能做什么",
		"/status": "目录",
		"/readme": "Hello",
		"hi":      "你的提示",
	} {
		if got := a.answer(dir, prompt); !strings.Contains(got, want) {
			t.Errorf("%s: want %q in %q", prompt, want, got)
		}
	}
}

func TestPaletteItems(t *testing.T) {
	a := testApp()
	s := a.newSession("")
	labels := map[string]bool{}
	for _, it := range a.paletteItems() {
		labels[it.Label] = true
	}
	for _, want := range []string{
		"New Tab",
		"Split Right",
		"Toggle Agent Chat",
		"Switch to " + s.displayTitle(),
		"Theme: Tokyo Night",
		"Theme: Follow system",
	} {
		if !labels[want] {
			t.Errorf("palette missing %q", want)
		}
	}
}

func TestViewChrome(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh") // the rail names the shell
	a := testApp()
	tt := ui.NewTester(a.view, 1000, 700)
	if !tt.HasText("No sessions") {
		t.Fatal("expected the empty state")
	}

	a.newSession("")
	a.newSession("/u/d")
	tt.Frame()
	if tt.HasText("No sessions") {
		t.Fatal("sessions should have appeared")
	}
	// The rail shows its header and both sessions as directory + shell.
	if !tt.HasText("TABS") || !tt.HasText("~") || !tt.HasText("/u/d") || !tt.HasText("zsh") {
		t.Fatalf("expected the rail header and both sessions, texts: %v", tt.Texts())
	}

	// The info panel opens with its sections and closes again.
	a.setPanel(panelInfo)
	tt.Frame()
	for _, want := range []string{"Working Directory", "Process", "Ports"} {
		if !tt.HasText(want) {
			t.Fatalf("info panel missing %q, texts: %v", want, tt.Texts())
		}
	}
	a.setPanel(panelNone)
	tt.Frame()
	if tt.HasText("Working Directory") {
		t.Fatal("info panel should close")
	}

	// Cmd+K opens the palette; Escape closes it.
	tt.Key(ui.Cmd, ui.KeyK)
	if !tt.HasText("New Tab") || !tt.HasText("Split Right") {
		t.Fatalf("palette should be open, texts: %v", tt.Texts())
	}
	tt.Key(0, ui.KeyEscape)
	if tt.HasText("New Tab") {
		t.Fatal("palette should close on Escape")
	}

	// Renaming the active session works, and survives a shell title.
	a.beginRename()
	tt.Frame()
	if !tt.HasText("Rename Tab") {
		t.Fatal("rename dialog should be open")
	}
	tt.Type("newname")
	tt.Key(0, ui.KeyEnter)
	if a.rename.open {
		t.Fatal("rename dialog should commit on Enter")
	}
	if got := a.currentSession().displayTitle(); got != "newname" {
		t.Fatalf("display title = %q, want newname", got)
	}
}

// TestRenamePinsTab checks that a renamed tab draws its name on the rail
// and the tab bar, and keeps it when the shell changes directory: the
// path the tab would otherwise show is pinned by the name, as otty pins
// the name the user gives a tab. The window title still follows the
// directory (user@host:dir), so the check looks at the tab label alone.
func TestRenamePinsTab(t *testing.T) {
	for _, layout := range []string{"sidebar-left", "tabs-top", "tabs-bottom"} {
		a := testApp()
		a.settings.WindowLayout = layout
		s := a.newSession("/tmp/nutty-origin")
		s.name = "pinned"
		tt := ui.NewTester(a.view, 1000, 700)
		tt.Frame()
		if _, ok := tt.Find("pinned"); !ok {
			t.Fatalf("%s: renamed tab missing its name, texts: %v", layout, tt.Texts())
		}
		if _, ok := tt.Find("/tmp/nutty-origin"); ok {
			t.Fatalf("%s: renamed tab still shows the path, texts: %v", layout, tt.Texts())
		}
		// The shell changes directory; the tab keeps its pinned name.
		s.dir = "/elsewhere"
		a.refreshTitle()
		tt.Frame()
		if _, ok := tt.Find("pinned"); !ok {
			t.Fatalf("%s: renamed tab lost its name after the path changed, texts: %v", layout, tt.Texts())
		}
		if _, ok := tt.Find("/elsewhere"); ok {
			t.Fatalf("%s: renamed tab moved with the path, texts: %v", layout, tt.Texts())
		}
	}
}

// TestDoubleClickRenames checks that two clicks on a tab open its rename
// dialog: on the rail and on the tab bar, the tab the double click hit,
// not just the active one. Both presses must land within 5px of each
// other, so the second press uses the rect found before the first click
// activated the tab and re-rendered its label bold (which shifts the
// label's center).
func TestDoubleClickRenames(t *testing.T) {
	for _, layout := range []string{"sidebar-left", "tabs-top", "tabs-bottom"} {
		a := testApp()
		a.settings.WindowLayout = layout
		a.newSession("/tmp/nutty-first")
		second := a.newSession("/tmp/nutty-second")
		tt := ui.NewTester(a.view, 1000, 700)
		tt.Frame()
		// Double click the second tab, which is not the active one.
		r, ok := tt.Find("/tmp/nutty-second")
		if !ok {
			t.Fatalf("%s: tab label not found, texts: %v", layout, tt.Texts())
		}
		tt.ClickAt(r.X+r.W/2, r.Y+r.H/2)
		tt.ClickAt(r.X+r.W/2, r.Y+r.H/2)
		tt.Frame()
		if !a.rename.open {
			t.Fatalf("%s: double click did not open the rename dialog", layout)
		}
		if a.rename.target != second {
			t.Fatalf("%s: rename target = %v, want the second tab", layout, a.rename.target)
		}
	}
}
