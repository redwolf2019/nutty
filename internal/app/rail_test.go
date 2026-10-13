package app

import "testing"

// TestHomeShown checks that only the home directory and the paths below it
// are shortened to "~": a sibling directory whose name starts with the
// home directory's keeps its path.
func TestHomeShown(t *testing.T) {
	home := homeDir()
	for in, want := range map[string]string{
		home:                 "~",
		home + "/src/nutty":  "~/src/nutty",
		home + "-backup/src": home + "-backup/src",
		"/tmp/elsewhere":     "/tmp/elsewhere",
	} {
		if got := homeShown(in); got != want {
			t.Errorf("homeShown(%q) = %q, want %q", in, got, want)
		}
	}
}
