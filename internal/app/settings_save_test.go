package app

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"nutty/internal/settings"
)

// TestSaveSettingsReplacesTheFile checks that a save never empties or
// rewrites the settings file that is there: it puts a whole new file in
// its place. A reader that opened the file before the save still reads
// all of the old settings, where a write in place hands it the file
// mid-rewrite. Nothing but the settings file is left in the folder.
func TestSaveSettingsReplacesTheFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir()) // never the user's own settings file
	a := &app{settings: settings.Default()}
	a.SaveSettings()
	path := settings.Path()
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	old, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = old.Close() }()

	a.settings.FontSize = 18
	a.SaveSettings()

	if got, _ := io.ReadAll(old); !bytes.Equal(got, before) {
		t.Error("the save rewrote the old file in place: a reader that had it open no longer reads the old settings")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var saved struct{ FontSize float64 }
	if err := json.Unmarshal(after, &saved); err != nil || saved.FontSize != 18 {
		t.Errorf("the saved file holds FontSize %v (%v), want 18", saved.FontSize, err)
	}
	if st, _ := os.Stat(path); st.Mode().Perm() != 0o644 {
		t.Errorf("the saved file's mode is %v, want 0644", st.Mode().Perm())
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Errorf("the save left %d files in the folder, want the settings file alone", len(entries))
	}
}

// TestSaveSettingsKeepsALink checks that a settings file that is a
// symbolic link stays one, and the save reaches the file it points to.
func TestSaveSettingsKeepsALink(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	target := filepath.Join(t.TempDir(), "nutty.json")
	if err := os.WriteFile(target, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := settings.Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	a := &app{settings: settings.Default()}
	a.settings.FontSize = 18
	a.SaveSettings()
	if st, err := os.Lstat(path); err != nil || st.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("the settings file is no longer a link (%v)", err)
	}
	b, _ := os.ReadFile(target)
	var saved struct{ FontSize float64 }
	if err := json.Unmarshal(b, &saved); err != nil || saved.FontSize != 18 {
		t.Errorf("the linked file holds FontSize %v (%v), want 18", saved.FontSize, err)
	}
}
