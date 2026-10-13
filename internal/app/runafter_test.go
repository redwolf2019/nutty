package app

import (
	"testing"
	"time"

	"nutty/internal/settings"

	"github.com/egoist/mygo/plugins/terminal"
)

// typedConn stands for a pane's shell: it reads nothing, and hands what
// is typed into the pane to the test.
type typedConn struct{ typed chan string }

func (c *typedConn) Read([]byte) (int, error) { select {} } // a shell that says nothing
func (c *typedConn) Write(p []byte) (int, error) {
	c.typed <- string(p)
	return len(p), nil
}
func (c *typedConn) Close() error { return nil }

// TestProfileRunAfter checks that a pane opened under a default profile
// with a Run After gets the command typed into its shell, with the Enter
// that runs it, and that a pane without one gets nothing typed.
func TestProfileRunAfter(t *testing.T) {
	conn := &typedConn{typed: make(chan string, 8)}
	old := newTerminal
	t.Cleanup(func() { newTerminal = old })
	newTerminal = func(o terminal.Options) (*terminal.Terminal, error) {
		o.Conn = conn
		return terminal.New(o)
	}
	a := &app{settings: settings.Default(), sideOpen: true, agent: &demoAgent{}}
	a.settings.Profiles = []settings.Profile{{Name: "web", RunAfter: "npm run dev"}}

	// No default profile: the pane opens and nothing is typed.
	a.newSession(t.TempDir())
	select {
	case got := <-conn.typed:
		t.Fatalf("without a default profile %q was typed into the pane", got)
	case <-time.After(200 * time.Millisecond):
	}

	a.settings.DefaultProfile = "web"
	a.newSession(t.TempDir())
	select {
	case got := <-conn.typed:
		if got != "npm run dev\r" {
			t.Errorf("the pane got %q typed, want the Run After command and Enter", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the profile's Run After was not typed into the new pane")
	}
}
