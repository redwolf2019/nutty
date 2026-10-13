package app

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/egoist/mygo/plugins/terminal"
)

// startedTerminals makes the app's terminals record the options each
// would start with, and start none.
func startedTerminals(t *testing.T) *[]terminal.Options {
	t.Helper()
	old := newTerminal
	t.Cleanup(func() { newTerminal = old })
	started := &[]terminal.Options{}
	newTerminal = func(o terminal.Options) (*terminal.Terminal, error) {
		*started = append(*started, o)
		return nil, os.ErrInvalid
	}
	return started
}

// TestWorkingDirectorySetting checks that a new tab starts in the Shell
// page's Working Directory, with ~ as the home directory, that a tab told
// where to open still opens there, and that a setting naming no directory
// falls back to home.
func TestWorkingDirectorySetting(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	work := filepath.Join(home, "work")
	if err := os.Mkdir(work, 0o755); err != nil {
		t.Fatal(err)
	}
	a := testApp()
	started := startedTerminals(t)
	for setting, want := range map[string]string{
		"":                 home,
		work:               work,
		"~/work":           work,
		"~":                home,
		" " + work + " ":   work,
		"/no/such/dir/yet": home,
	} {
		a.settings.WorkingDirectory = setting
		s := a.newSession("")
		if s.dir != want {
			t.Errorf("Working Directory %q: the tab's directory is %q, want %q", setting, s.dir, want)
		}
		if got := (*started)[len(*started)-1].Dir; got != want {
			t.Errorf("Working Directory %q: the shell starts in %q, want %q", setting, got, want)
		}
	}
	a.settings.WorkingDirectory = work
	if s := a.newSession(home); s.dir != home {
		t.Errorf("a tab opened in %q is in %q", home, s.dir)
	}
}

// TestTermSetting checks that the Shell page's Term reaches a new shell
// as its TERM, and that "auto" leaves TERM to the terminal.
func TestTermSetting(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	a := testApp()
	started := startedTerminals(t)
	a.settings.Env = []string{"EDITOR=vi"}
	for term, want := range map[string]bool{"xterm-kitty": true, "auto": false, "": false} {
		a.settings.Term = term
		a.newSession("")
		env := (*started)[len(*started)-1].Env
		if got := slices.Contains(env, "TERM="+term); got != want {
			t.Errorf("Term %q: the shell's environment is %q", term, env)
		}
		if !slices.Contains(env, "EDITOR=vi") {
			t.Errorf("Term %q: the environment lost the settings' variables: %q", term, env)
		}
	}
}
