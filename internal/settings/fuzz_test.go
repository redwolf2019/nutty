package settings

import (
	"slices"
	"strings"
	"testing"
)

// FuzzShellArgs checks the command line splitter on any text: it never
// yields an empty argument, and text without quotes splits exactly at its
// spaces and tabs.
func FuzzShellArgs(f *testing.F) {
	for _, seed := range []string{"", "zsh", "docker compose up", "sh -c 'echo hi'", `bash -c "ls -la"`,
		`a "b c`, `''`, "a\tb  c ", `"a'b" 'c"d'`, "\x00 \xff'", `foo\ bar`, "\xf7 caf\xe9", "日本 語"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		args := ShellArgs(s)
		for _, a := range args {
			if a == "" {
				t.Errorf("ShellArgs(%q) holds an empty argument: %q", s, args)
			}
		}
		if !strings.ContainsAny(s, `'"`) {
			want := strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == '\t' })
			if !slices.Equal(args, want) {
				t.Errorf("ShellArgs(%q) = %q, want %q", s, args, want)
			}
		}
	})
}

// FuzzAcceleratorOf checks the trigger converter on any text: it gives no
// accelerator, or one with a part for each part of the trigger, the key
// last.
func FuzzAcceleratorOf(f *testing.F) {
	for _, kb := range defaultKeybinds {
		f.Add(kb.trigger)
	}
	for _, seed := range []string{"", "+", "cmd+", "+k", "cmd++k", "hyper+k", "cmd+f12", "cmd+f", "f", "cmd+shift+é", "a+b+c+d+e+f"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, trigger string) {
		acc := AcceleratorOf(trigger)
		if acc == "" {
			return
		}
		// "\\" and the other keys are one part; "+" is no key name.
		if got, want := strings.Count(acc, "+"), strings.Count(trigger, "+"); got != want {
			t.Errorf("AcceleratorOf(%q) = %q: %d separators, want %d", trigger, acc, got, want)
		}
	})
}

// TestDefaultTriggersConvert checks that every default trigger becomes an
// accelerator with a key.
func TestDefaultTriggersConvert(t *testing.T) {
	for _, kb := range defaultKeybinds {
		acc := AcceleratorOf(kb.trigger)
		if acc == "" || strings.HasSuffix(acc, "+") {
			t.Errorf("AcceleratorOf(%q) = %q, want an accelerator with a key", kb.trigger, acc)
		}
	}
}

// FuzzValidEnv checks the environment filter on any list: it keeps the
// entries with a value, in their order, and no others.
func FuzzValidEnv(f *testing.F) {
	for _, seed := range []string{"", "A=1", "A=1\nB\nC=", "=x\n\n=", "A\nB", "PATH=/a:/b\nTERM"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		list := strings.Split(s, "\n")
		var want []string
		for _, e := range list {
			if strings.Contains(e, "=") {
				want = append(want, e)
			}
		}
		if got := ValidEnv(list); !slices.Equal(got, want) {
			t.Errorf("ValidEnv(%q) = %q, want %q", list, got, want)
		}
	})
}

// FuzzSplitLines checks the list editors' splitter on any text: no line it
// gives is blank, has space at its ends or holds a line break, and the
// lines put back in the editor split to themselves, so an untouched
// editor never rewrites its list.
func FuzzSplitLines(f *testing.F) {
	for _, seed := range []string{"", "a", "a\nb", " a \n\n\tb\n", "\r\n a\r\n", "a\u00a0\n\u2028b", "\n\n\n"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		lines := splitLines(s)
		for _, ln := range lines {
			if ln == "" || ln != strings.TrimSpace(ln) || strings.Contains(ln, "\n") {
				t.Errorf("splitLines(%q) holds the line %q", s, ln)
			}
		}
		if again := splitLines(strings.Join(lines, "\n")); !slices.Equal(again, lines) {
			t.Errorf("splitLines(%q) = %q, which splits again to %q", s, lines, again)
		}
	})
}
