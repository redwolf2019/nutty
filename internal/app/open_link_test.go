package app

import (
	"os"
	"path/filepath"
	"testing"

	"nutty/internal/settings"
)

// TestLinkPath checks that the paths a terminal passes for a Cmd+click,
// as printed, resolve to the file they name: from ~, relative to the
// shell's directory, and with the :line or :line:col a compiler prints.
// Only an absolute path without a suffix used to open.
func TestLinkPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, "proj")
	if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "src", "main.go")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	for link, want := range map[string]string{
		file:                      file,
		file + ":12":              file,
		file + ":12:3":            file,
		"file://" + file:          file,
		"src/main.go":             file,
		"./src/main.go:12":        file,
		"src/main.go:12:3":        file,
		"../proj/src/main.go":     file,
		"~/proj/src/main.go":      file,
		"~/proj/src/main.go:7":    file,
		"src":                     filepath.Join(dir, "src"),
		"~":                       home,
		"src/missing.go":          "",
		"src/missing.go:12":       "",
		"/no/such/file":           "",
		"~other/proj/src/main.go": "",
	} {
		if got := linkPath(link, dir); got != want {
			t.Errorf("linkPath(%q) = %q, want %q", link, got, want)
		}
	}
}

// TestOpenLinkFromSessionDir checks that a relative path opens from the
// directory of the session the pane belongs to, not from the app's own
// working directory, and that a path naming no file is left alone.
func TestOpenLinkFromSessionDir(t *testing.T) {
	var opened []string
	old := openFile
	t.Cleanup(func() { openFile = old })
	openFile = func(path string) { opened = append(opened, path) }

	dir := t.TempDir()
	file := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	p := &pane{}
	s := &session{id: 1, dir: dir, root: &node{term: p}, panes: []*pane{p}}
	a := &app{settings: settings.Default(), sessions: []*session{s}, activePane: p}

	if !a.openLink(p, "notes.txt:3") {
		t.Fatal("a file in the session's directory did not open")
	}
	if len(opened) != 1 || opened[0] != file {
		t.Fatalf("opened %q, want %q", opened, file)
	}
	if a.openLink(p, "nothing-here.txt") {
		t.Fatal("a path naming no file reported that it opened")
	}
	if len(opened) != 1 {
		t.Fatalf("a path naming no file was opened: %q", opened)
	}
}
