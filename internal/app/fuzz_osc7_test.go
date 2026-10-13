package app

import (
	"path/filepath"
	"testing"

	"nutty/internal/settings"

	"github.com/egoist/mygo/plugins/terminal"
)

// FuzzTrackWorkingDir lets a program in the terminal report any working
// directory (OSC 7) and checks what the session takes from it: its
// directory moves to an absolute path or stays where it was. The
// directory is handed to open and to git, so anything else — a relative
// path, an option, another kind of URL — must not get through.
func FuzzTrackWorkingDir(f *testing.F) {
	registerFonts()
	font := (&app{settings: settings.Default()}).termFont()
	for _, payload := range []string{
		"file://localhost/tmp/elsewhere", "file:///tmp/with%20space", "file://host", "file://", "file:///", "file://h/%", "file://h/%zz%4",
		"-aCalculator", "relative/dir", "kitty-shell-cwd://h/tmp/x", "/abs", "", "~", "\x1b]7;nested\x07", "a\x00b",
	} {
		f.Add(payload)
	}
	f.Fuzz(func(t *testing.T, payload string) {
		term, err := terminal.New(terminal.Options{
			Conn:  &scriptConn{data: []byte("\x1b]7;" + payload + "\x07")},
			Font:  font,
			Theme: settings.OttyTermTheme(false),
		})
		if err != nil {
			t.Skip("no terminal:", err)
		}
		defer func() { _ = term.Close() }()
		<-term.Done()
		p := &pane{term: term}
		s := &session{id: 1, dir: "/start", root: &node{term: p}, panes: []*pane{p}}
		a := &app{settings: settings.Default(), sessions: []*session{s}, active: 0, activePane: p}
		a.trackWorkingDir()
		if s.dir != "/start" && !filepath.IsAbs(s.dir) {
			t.Fatalf("OSC 7 %q moved the session to %q", payload, s.dir)
		}
	})
}
