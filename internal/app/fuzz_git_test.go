package app

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
)

// FuzzParseGitStatus feeds the Git section whatever git status might
// print, a repository's file names being anyone's to choose, and checks
// what it makes of it: no panic, here or in the row each change draws as,
// every change with its two-letter status and a path that is one of the
// entries, and a diff command the path cannot add an option to.
func FuzzParseGitStatus(f *testing.F) {
	for _, out := range []string{
		"## main...origin/main [ahead 1]\x00 M session.go\x00M  topbar.go\x00?? new file.go\x00",
		"## No commits yet on main\x00A  中文.txt\x00",
		"## HEAD (no branch)\x00R  new.txt\x00old.txt\x00 D gone.go\x00",
		"R  last-entry-renamed\x00",
		"## \x00##  \x00## ...\x00",
		"?? -option\x00?? \x00UU both\x00",
		" M a\n M b\n",
		"", "\x00\x00\x00", "## ", "abc", "abcd",
		" \xef00", // a status byte that is no letter, as the fuzzer found
	} {
		f.Add(out)
	}
	theme := ui.LightTheme()
	f.Fuzz(func(t *testing.T, out string) {
		branch, changes := parseGitStatus(out)
		if strings.ContainsAny(branch, " \x00") {
			t.Fatalf("branch %q holds more than a name", branch)
		}
		for _, ch := range changes {
			if len(ch.status) != 2 {
				t.Fatalf("status %q is not two letters", ch.status)
			}
			if !slices.ContainsFunc(strings.Split(out, "\x00"), func(entry string) bool {
				return len(entry) > 3 && entry[:2] == ch.status && entry[3:] == ch.path
			}) {
				t.Fatalf("change %q %q is not an entry of the output", ch.status, ch.path)
			}
			if label := gitStatusLabel(ch.status); utf8.RuneCountInString(label) != 1 {
				t.Fatalf("label %q of status %q is not one letter", label, ch.status)
			}
			_ = gitStatusColor(theme, ch.status)
			args := gitDiffArgs("/repo", ch.path, ch.status[0] != ' ')
			if n := len(args); args[n-2] != "--" || args[n-1] != ":(top,literal)"+ch.path {
				t.Fatalf("the diff of %q runs as %q", ch.path, args)
			}
		}
	})
}
