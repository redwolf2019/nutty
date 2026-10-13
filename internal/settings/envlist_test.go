package settings

import (
	"reflect"
	"testing"
)

// TestEnvListTerm checks the environment new sessions get from the Shell
// page: the Term setting as TERM, ahead of the variables, so one of them
// that sets TERM is the one the terminal keeps.
func TestEnvListTerm(t *testing.T) {
	s := Default()
	s.Env = []string{"A=1", "novalue", "TERM=vt100"}
	for term, want := range map[string][]string{
		"":             {"A=1", "TERM=vt100"},
		"auto":         {"A=1", "TERM=vt100"},
		"xterm-kitty":  {"TERM=xterm-kitty", "A=1", "TERM=vt100"},
		" xterm-ghost": {"TERM=xterm-ghost", "A=1", "TERM=vt100"},
	} {
		s.Term = term
		if got := s.EnvList(); !reflect.DeepEqual(got, want) {
			t.Errorf("Term %q: EnvList = %q, want %q", term, got, want)
		}
	}
}
