package app

import (
	"fmt"
	"os"
	"testing"

	"github.com/egoist/mygo/plugins/terminal"
)

// TestMain gives the package's tests a home directory of their own. The
// app keeps files under the user's home — the settings, the shell
// integration it writes to the cache directory when a session starts —
// and reads some there, as the shell's history; a test that opens a
// session must neither leave files in the real one nor depend on what it
// holds.
func TestMain(m *testing.M) {
	// The terminal's library is found in the real cache directory, where
	// a test run downloads it once: load it before the home changes, or
	// every run would download it again into the new one.
	_ = terminal.Load()
	home, err := os.MkdirTemp("", "nutty-test-home-")
	if err == nil {
		for _, name := range []string{"HOME", "XDG_CONFIG_HOME", "XDG_CACHE_HOME"} {
			if err == nil {
				err = os.Setenv(name, home)
			}
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "nutty tests: no home directory of their own:", err)
		os.Exit(1)
	}
	code := m.Run()
	_ = os.RemoveAll(home)
	os.Exit(code)
}
