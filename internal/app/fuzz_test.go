package app

import (
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
)

// The fuzz tests feed the functions that read text from outside the app —
// what a shell, a program in the terminal or a file on disk wrote — and
// check that none panics and that each keeps what it promises. A plain
// `go test` runs their seeds; `go test -run '^$' -fuzz FuzzName` searches
// for more.

// FuzzChunks checks the pieces a reply streams in: together they are the
// text, rune for rune, and none is longer than asked or empty.
func FuzzChunks(f *testing.F) {
	f.Add("## 你的提示\n\n> hello 🌰 world", uint8(24))
	f.Add("", uint8(1))
	f.Add("\xff\xfe broken \xe4\xb8", uint8(3))
	f.Fuzz(func(t *testing.T, s string, size uint8) {
		n := int(size)%64 + 1 // the callers ask for a positive size
		pieces := chunks(s, n)
		if got, want := strings.Join(pieces, ""), string([]rune(s)); got != want {
			t.Fatalf("the pieces join to %q, want %q", got, want)
		}
		for _, p := range pieces {
			if c := utf8.RuneCountInString(p); c == 0 || c > n {
				t.Fatalf("a piece of %d runes, want 1 to %d: %q", c, n, p)
			}
		}
	})
}

// FuzzHomeShown checks the directory the tabs and the title show: the
// directory itself, or "~" in place of the home directory at its start.
func FuzzHomeShown(f *testing.F) {
	home := homeDir()
	for _, dir := range []string{"", "/", home, home + "/code", home + "2/code", home + "/", "relative", "~", "/tmp/\x00"} {
		f.Add(dir)
	}
	f.Fuzz(func(t *testing.T, dir string) {
		shown := homeShown(dir)
		if shown == dir {
			return
		}
		rest, ok := strings.CutPrefix(shown, "~")
		if !ok || home+rest != dir {
			t.Fatalf("homeShown(%q) = %q with the home %q", dir, shown, home)
		}
	})
}

// FuzzPortOf checks the port read from an lsof name: a piece of the name's
// end, with no address left in front of it.
func FuzzPortOf(f *testing.F) {
	for _, name := range []string{"*:8000", "*:8000 (LISTEN)", "127.0.0.1:5173", "[::1]:8080", "[fe80::1%lo0]:22", "*:*", "NAME", "", ":", "]:", "[::]:*"} {
		f.Add(name)
	}
	f.Fuzz(func(t *testing.T, name string) {
		port := portOf(name)
		if !strings.HasSuffix(strings.TrimSuffix(strings.TrimSpace(name), " (LISTEN)"), port) {
			t.Fatalf("portOf(%q) = %q, which is not the name's end", name, port)
		}
	})
}

// FuzzAgentAnswer checks the demo agent's replies for any prompt and any
// working directory, an odd one included: the README it looks for is
// always inside the directory, and the reply is never empty.
func FuzzAgentAnswer(f *testing.F) {
	dir := f.TempDir()
	for _, cwd := range []string{dir, "", "/", "/nonexistent/dir", "relative/dir", "..", "/tmp/a b/文", "/tmp/\x00nul", "-option", "~"} {
		for _, prompt := range []string{"/help", "/status", "/readme", " /readme\n", "hi", "", "a\nb", "```go\nfunc"} {
			f.Add(cwd, prompt)
		}
	}
	f.Fuzz(func(t *testing.T, cwd, prompt string) {
		if p := readmePath(cwd); filepath.Dir(p) != filepath.Clean(cwd) {
			t.Fatalf("readmePath(%q) = %q, outside the directory", cwd, p)
		}
		if (&demoAgent{}).answer(cwd, prompt) == "" {
			t.Fatalf("no answer to %q in %q", prompt, cwd)
		}
	})
}

// FuzzChatMarkdown draws a reply of any text as the chat does, parsed as
// markdown: /readme puts whatever file is called README.md on the panel,
// so no text may take the window down.
func FuzzChatMarkdown(f *testing.F) {
	registerFonts()
	for _, text := range []string{
		"## README\n\n- a\n- b\n\n```go\nfunc hello() string { return \"🌰\" }\n```\n\n> quote\n\n---\n",
		"| a | b |\n|---|---|\n| 1 | 2 |\n", "[link](http://x) ![img](y.png) <b>html</b>", "\xff\xfe\x00", "```", "1. \n   - [ ] x\n", "",
	} {
		f.Add(text)
	}
	f.Fuzz(func(t *testing.T, text string) {
		a := &app{}
		m := &chatMsg{text: text, streaming: true}
		ui.NewTester(func(c *ui.Context) { a.chatBubble(c, m) }, 360, 400).Frame()
	})
}
