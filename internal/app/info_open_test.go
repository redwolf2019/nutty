package app

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// stubOpen replaces the open command for a test, so nothing opens: the
// channel receives the arguments of each call, and a call fails when fail
// says so.
func stubOpen(t *testing.T, fail func(args []string) bool) <-chan []string {
	t.Helper()
	calls := make(chan []string, 8)
	real := openCmd
	openCmd = func(args ...string) error {
		calls <- args
		if fail != nil && fail(args) {
			return errors.New("stub: open failed")
		}
		return nil
	}
	t.Cleanup(func() { openCmd = real })
	return calls
}

// nextOpen is the next call of the stubbed open, or nil when none comes.
func nextOpen(calls <-chan []string, wait time.Duration) []string {
	select {
	case args := <-calls:
		return args
	case <-time.After(wait):
		return nil
	}
}

// TestOpenZedFallsBack checks that a file still opens on a Mac without
// Zed: when open finds no Zed, the plain open follows, so the click on a
// file of the Files section is not without effect.
func TestOpenZedFallsBack(t *testing.T) {
	const path = "/nonexistent/nutty-open-test/main.go"
	calls := stubOpen(t, func(args []string) bool { return args[0] == "-a" })
	openZed(path)
	if got := nextOpen(calls, 2*time.Second); !slices.Equal(got, []string{"-a", "Zed", path}) {
		t.Fatalf("first open = %q, want it to ask for Zed", got)
	}
	if got := nextOpen(calls, 2*time.Second); !slices.Equal(got, []string{path}) {
		t.Fatalf("second open = %q, want the plain open of %s", got, path)
	}

	// With Zed there, Zed is all that opens.
	calls = stubOpen(t, nil)
	openZed(path)
	if got := nextOpen(calls, 2*time.Second); !slices.Equal(got, []string{"-a", "Zed", path}) {
		t.Fatalf("open = %q, want it to ask for Zed", got)
	}
	if got := nextOpen(calls, 50*time.Millisecond); got != nil {
		t.Fatalf("a second open ran with Zed installed: %q", got)
	}
}

// TestRevealWaitsForOpen checks that Reveal in Finder, from the menu and
// from the info panel, goes through openCmd, which waits for the child:
// started and never waited for, each open stayed a zombie until the app
// quit.
func TestRevealWaitsForOpen(t *testing.T) {
	setUILang("english") // the test clicks the action by its English label
	const dir = "/nonexistent/nutty-open-test"
	calls := stubOpen(t, nil)
	a := testApp()
	a.newSession(dir)

	a.revealInFinder()
	if got := nextOpen(calls, 2*time.Second); !slices.Equal(got, []string{dir}) {
		t.Fatalf("the menu's reveal ran open %q, want %q through openCmd", got, dir)
	}

	a.setPanel(panelInfo)
	tt := ui.NewTester(a.view, 1000, 700)
	tt.Frame()
	if err := tt.Click("Reveal in Finder"); err != nil {
		t.Fatal(err)
	}
	if got := nextOpen(calls, 2*time.Second); !slices.Equal(got, []string{dir}) {
		t.Fatalf("the panel's reveal ran open %q, want %q through openCmd", got, dir)
	}
}
