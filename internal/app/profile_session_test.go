package app

import (
	"errors"
	"testing"

	"nutty/internal/settings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// profileApp makes an app whose default profile is pr and whose terminals
// are never started: started records the options each would have had, and
// every start fails with err.
func profileApp(t *testing.T, pr settings.Profile, err error) (a *app, started *[]terminal.Options) {
	t.Helper()
	old := newTerminal
	t.Cleanup(func() { newTerminal = old })
	started = &[]terminal.Options{}
	newTerminal = func(o terminal.Options) (*terminal.Terminal, error) {
		*started = append(*started, o)
		return nil, err
	}
	s := settings.Default()
	s.Profiles = []settings.Profile{pr}
	s.DefaultProfile = pr.Name
	return &app{settings: s, sideOpen: true, agent: &demoAgent{}}, started
}

// TestProfileDirIsTheSessions checks that a new tab starts in the default
// profile's directory and says so, and that Duplicate Tab and a split
// start where their tab is: the profile's directory used to replace the
// session's for every terminal, behind a tab that still showed home.
func TestProfileDirIsTheSessions(t *testing.T) {
	a, started := profileApp(t, settings.Profile{Name: "work", CWD: "/tmp/nutty-profile"}, errors.New("not started"))

	s := a.newSession("")
	if s.dir != "/tmp/nutty-profile" {
		t.Errorf("a new tab's directory = %q, want the profile's", s.dir)
	}
	if got := (*started)[0].Dir; got != "/tmp/nutty-profile" {
		t.Errorf("a new tab's shell starts in %q, want the profile's directory", got)
	}

	other := a.newSession("/tmp/nutty-elsewhere")
	a.duplicateTab()
	if got := a.currentSession().dir; got != "/tmp/nutty-elsewhere" {
		t.Errorf("the duplicate's directory = %q, want its original's", got)
	}
	if got := (*started)[2].Dir; got != "/tmp/nutty-elsewhere" {
		t.Errorf("the duplicate's shell starts in %q, want its original's directory", got)
	}

	a.splitPane(other.panes[0], false)
	if got := (*started)[3].Dir; got != "/tmp/nutty-elsewhere" {
		t.Errorf("a split's shell starts in %q, want its tab's directory", got)
	}
}

// TestProfileThemeSurvivesRestyle checks that the themes the default
// profile picks are the ones every restyle applies, not only the ones a
// new pane starts with: a change of settings or of the system's appearance
// used to put the settings' own theme back on the pane.
func TestProfileThemeSurvivesRestyle(t *testing.T) {
	a, started := profileApp(t, settings.Profile{Name: "night", Theme: "Dracula"}, errors.New("not started"))
	want := settings.ThemeFor("Dracula", false).Background
	if base := settings.ThemeFor(a.settings.ThemeName, false).Background; base == want {
		t.Fatal("the test needs a profile theme that differs from the settings' theme")
	}

	a.newSession("")
	if got := (*started)[0].Theme; got == nil || got.Background != want {
		t.Fatal("a new pane should start with the profile's theme")
	}
	// What applyTermTheme sets on every pane, on ApplySettings and when
	// the appearance changes.
	light, dark := a.termThemes()
	if light == nil || light.Background != want {
		t.Error("a restyle puts the settings' light theme back in place of the profile's")
	}
	if dark == nil || dark.Background != want {
		t.Error("a restyle puts the settings' dark theme back in place of the profile's")
	}
}

// TestPaneShowsWhyItHasNoTerminal checks that a pane without a terminal
// gives the reason: the pane always blamed the terminal library, also for
// a command in the settings that is not installed.
func TestPaneShowsWhyItHasNoTerminal(t *testing.T) {
	a, _ := profileApp(t, settings.Profile{Name: "fish", Command: "fish"},
		errors.New(`exec: "fish": executable file not found in $PATH`))
	a.newSession("")
	tt := ui.NewTester(a.view, 1000, 700)
	tt.Frame()
	if !tt.HasText(tr("Terminal unavailable")) {
		t.Fatalf("the pane should say it has no terminal, texts: %v", tt.Texts())
	}
	if !tt.HasText(`exec: "fish": executable file not found in $PATH`) {
		t.Errorf("the pane does not say why it has no terminal, texts: %v", tt.Texts())
	}
}
