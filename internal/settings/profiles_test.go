package settings

import (
	"reflect"
	"testing"

	"github.com/egoist/mygo/ui"
	uixsettings "github.com/smallnest/uix/components/settings"
)

// profilesPage is the Profiles page over s, the first profile chosen, as
// the window opens it.
func profilesPage(s *Settings) (*settingsWindow, *ui.Tester) {
	s.Language = "english" // the page is driven by its English labels
	sw := &settingsWindow{app: mockApp{}, s: s, last: *s,
		state: uixsettings.SettingsState{Category: "profiles"}}
	tt := ui.NewTester(sw.view, 900, 1600)
	tt.Frame()
	return sw, tt
}

// focusProfileField gives the focus to the field under a label of the
// profile editor, where a field draws its input.
func focusProfileField(t *testing.T, tt *ui.Tester, label string) {
	t.Helper()
	r, ok := tt.Find(label)
	if !ok {
		t.Fatalf("no %q field", label)
	}
	tt.ClickAt(r.X+40, r.Y+r.H+14)
}

// TestProfileDeleteKeepsNextEnv checks that deleting a profile leaves the
// next one's environment alone. The environment editor's buffer is kept by
// the profile's index, and the next profile takes the deleted one's index.
func TestProfileDeleteKeepsNextEnv(t *testing.T) {
	s := Default()
	s.Profiles = []Profile{
		{Name: "Api", Env: []string{"X=1"}},
		{Name: "Web", Env: []string{"Y=2"}},
	}
	_, tt := profilesPage(s)
	if err := tt.Click("Delete"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("Web"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Frame()
	if len(s.Profiles) != 1 || !reflect.DeepEqual(s.Profiles[0].Env, []string{"Y=2"}) {
		t.Fatalf("after deleting Api, the profiles are %+v, want Web with Y=2", s.Profiles)
	}
}

// TestProfileRenameKeepsDefault checks that the default profile stays the
// default while its name is edited: the default is stored by name, so it
// follows the rename, through an emptied field too, and a name another
// profile has is not taken.
func TestProfileRenameKeepsDefault(t *testing.T) {
	s := Default()
	s.Profiles = []Profile{{Name: "Web", Command: "zsh"}, {Name: "Api"}}
	s.DefaultProfile = "Web"
	_, tt := profilesPage(s)
	focusProfileField(t, tt, "Name")
	tt.Type("2")
	tt.Frame()
	if pr := s.Profile(); s.Profiles[0].Name != "Web2" || pr == nil || pr.Command != "zsh" {
		t.Fatalf("after typing 2: name %q, default %q", s.Profiles[0].Name, s.DefaultProfile)
	}
	// Emptying the field to type another name does not unname the profile:
	// it keeps the last name it had, and stays the default.
	for range "Web2" {
		tt.Key(0, ui.KeyBackspace)
	}
	tt.Frame()
	if pr := s.Profile(); s.Profiles[0].Name != "W" || pr == nil || pr.Command != "zsh" {
		t.Fatalf("with the field emptied: name %q, default %q", s.Profiles[0].Name, s.DefaultProfile)
	}
	// "Api" is the other profile's name on the way to "Api2".
	tt.Type("Api")
	tt.Frame()
	if s.Profiles[0].Name != "W" {
		t.Fatalf("typing another profile's name gave %q, want the name left at W", s.Profiles[0].Name)
	}
	tt.Type("2")
	tt.Frame()
	if pr := s.Profile(); s.Profiles[0].Name != "Api2" || pr == nil || pr.Command != "zsh" {
		t.Fatalf("after typing Api2: name %q, default %q", s.Profiles[0].Name, s.DefaultProfile)
	}
}

// TestProfileNamesStayUnique checks that the names the page makes up are
// not ones a profile already has: a new profile after a deletion, and a
// second copy of a profile.
func TestProfileNamesStayUnique(t *testing.T) {
	s := Default()
	_, tt := profilesPage(s)
	for _, click := range []string{"New Profile", "New Profile", "Profile 1", "Delete", "New Profile",
		"Profile 2", "Duplicate", "Profile 2", "Duplicate"} {
		if err := tt.Click(click); err != nil {
			t.Fatalf("click %q: %v", click, err)
		}
		tt.Frame()
	}
	seen := map[string]bool{}
	for _, p := range s.Profiles {
		if seen[p.Name] {
			t.Errorf("two profiles are named %q", p.Name)
		}
		seen[p.Name] = true
	}
	if len(s.Profiles) != 4 {
		t.Fatalf("profiles = %d, want 4", len(s.Profiles))
	}
}
