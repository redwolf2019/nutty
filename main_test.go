package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// testApp makes an app whose terminals never start a shell.
func testApp() *app {
	newTerminal = func(terminal.Options) (*terminal.Terminal, error) {
		return nil, errors.New("stub: no terminal in tests")
	}
	return &app{settings: defaultSettings(), sideOpen: true, agent: &demoAgent{}}
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
		"Theme: TokyoNight",
		"Theme: Follow system",
	} {
		if !labels[want] {
			t.Errorf("palette missing %q", want)
		}
	}
}

func TestViewChrome(t *testing.T) {
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
