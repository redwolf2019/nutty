package settings

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	uixsettings "github.com/smallnest/uix/components/settings"
)

// brokenFile is a settings file a hand edit left unparsable: the value the
// user set is in it, and a stray comma after it.
const brokenFile = `{
  "FontSize": 18,
}`

// writeSettingsFile puts a settings file under a HOME of the test's own,
// never the user's, and returns its path.
func writeSettingsFile(t *testing.T, content string) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	path := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestLoadKeepsAFileItCannotRead checks that a settings file that does
// not parse is not left where the next save replaces it with the defaults
// Load returns for it: it is kept, whole, beside the settings file.
func TestLoadKeepsAFileItCannotRead(t *testing.T) {
	path := writeSettingsFile(t, brokenFile)
	if s := Load(); s == nil || s.FontSize != 13 || len(s.DetailsViews) == 0 {
		t.Fatalf("Load of a broken file = %+v, want the defaults", s)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("the broken file is still the settings file (%v): the next save replaces it", err)
	}
	if kept, err := os.ReadFile(path + ".bad"); err != nil || string(kept) != brokenFile {
		t.Errorf("the broken file was not kept as settings.json.bad: %q, %v", kept, err)
	}
}

// TestLoadReadsTheFile checks the ordinary loads: a file that parses is
// read and left alone, and no file gives the defaults.
func TestLoadReadsTheFile(t *testing.T) {
	path := writeSettingsFile(t, `{"FontSize": 18}`)
	if s := Load(); s.FontSize != 18 || s.FontFamily != "JetBrains Mono" {
		t.Fatalf("Load = FontSize %v, FontFamily %q", s.FontSize, s.FontFamily)
	}
	if _, err := os.Stat(path + ".bad"); !os.IsNotExist(err) {
		t.Error("a file that parses was moved aside")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if s := Load(); s.FontSize != 13 {
		t.Fatalf("Load without a file = FontSize %v, want the default", s.FontSize)
	}
}

// fileApp is a host that counts the saves the settings window asks for.
type fileApp struct{ saved int }

func (*fileApp) ApplySettings()  {}
func (a *fileApp) SaveSettings() { a.saved++ }
func (*fileApp) ApplyKeybinds()  {}

// TestReloadKeepsSettingsOnABrokenFile checks that Reload of a file that
// does not parse leaves the settings as they are, saves nothing, and
// leaves the file for the user to mend.
func TestReloadKeepsSettingsOnABrokenFile(t *testing.T) {
	path := writeSettingsFile(t, brokenFile)
	s := Default()
	s.Language = "english" // the button is found by its English label
	s.FontSize = 16
	host := &fileApp{}
	sw := &settingsWindow{app: host, s: s, last: *s,
		state: uixsettings.SettingsState{Category: "advanced"}}
	tt := ui.NewTester(sw.view, 900, 800)
	tt.Frame()
	if err := tt.Click("Reload"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if s.FontSize != 16 || s.Language != "english" {
		t.Errorf("Reload of a broken file changed the settings: FontSize %v, Language %q", s.FontSize, s.Language)
	}
	if host.saved != 0 {
		t.Errorf("Reload of a broken file saved %d times", host.saved)
	}
	if b, err := os.ReadFile(path); err != nil || string(b) != brokenFile {
		t.Errorf("Reload touched the broken file: %q, %v", b, err)
	}
}
