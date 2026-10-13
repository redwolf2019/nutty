package app

import (
	"testing"

	"nutty/internal/settings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// TestSettingsAreShared checks that the settings are loaded once and
// handed to every caller, so the windows and the menu hold the same ones.
func TestSettingsAreShared(t *testing.T) {
	// Loading reads the settings file: keep to an empty home.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	old := shared
	t.Cleanup(func() { shared = old })
	shared = nil

	first := sharedSettings()
	if first == nil {
		t.Fatal("no settings")
	}
	first.FontSize = 21 // a change made from one window
	if second := sharedSettings(); second != first || second.FontSize != 21 {
		t.Fatal("a second window got settings of its own")
	}
}

// TestApplySettingsReachesEveryWindow changes the font size from one
// window and checks that the terminal of another window takes it: each
// window used to apply the settings to its own terminals only.
func TestApplySettingsReachesEveryWindow(t *testing.T) {
	registerFonts()
	oldTerm := newTerminal
	t.Cleanup(func() { newTerminal = oldTerm })
	newTerminal = func(o terminal.Options) (*terminal.Terminal, error) {
		o.Conn = &scriptConn{data: []byte("$ "), hold: true}
		return terminal.New(o)
	}
	// Two windows on one set of settings, registered as newWindow does.
	s := settings.Default()
	one := &app{settings: s, agent: &demoAgent{}}
	two := &app{settings: s, agent: &demoAgent{}}
	appsMu.Lock()
	apps[-1], apps[-2] = one, two
	appsMu.Unlock()
	t.Cleanup(func() {
		appsMu.Lock()
		delete(apps, -1)
		delete(apps, -2)
		appsMu.Unlock()
	})

	term := two.newSession("/tmp/nutty-shared").panes[0].term
	tt := ui.NewTester(two.view, 1000, 700)
	tt.Frame()
	tt.Frame()
	before, _ := term.Size()

	s.FontSize = 26
	one.ApplySettings()
	tt.Frame()
	tt.Frame()
	if after, _ := term.Size(); after >= before {
		t.Fatalf("the other window's terminal kept its font: %d columns before, %d after", before, after)
	}
}
