package app

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"nutty/internal/settings"

	"github.com/egoist/mygo/plugins/terminal"
)

// integrationHome makes a home directory for a shell under test, with the
// startup files given, and points this process's HOME at it, so the
// integration's files land in it too. ZDOTDIR and PROMPT_COMMAND are
// cleared from this process's environment, which shellEnv falls back on.
func integrationHome(t *testing.T, files map[string]string) string {
	t.Helper()
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	for _, name := range []string{"ZDOTDIR", "PROMPT_COMMAND"} {
		t.Setenv(name, "") // restored when the test ends
		if err := os.Unsetenv(name); err != nil {
			t.Fatal(err)
		}
	}
	for name, content := range files {
		file := filepath.Join(home, name)
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return home
}

// shellPrompt is the prompt the startup files of the tests set: the
// session below types its next line when it has seen one more of it.
const shellPrompt = "nutty-test> "

// shellSession runs a real shell in a terminal of its own (script(1)
// gives it one) with a clean environment plus env, starting in home, and
// types the lines one at a time, each when the shell has drawn its next
// prompt. It returns everything the terminal received.
func shellSession(t *testing.T, home string, env []string, argv []string, lines ...string) string {
	t.Helper()
	if runtime.GOOS != "darwin" {
		t.Skip("script(1) takes its command this way on macOS only")
	}
	if _, err := exec.LookPath("script"); err != nil {
		t.Skip("no script(1) to give the shell a terminal")
	}
	if _, err := os.Stat(argv[0]); err != nil {
		t.Skipf("no %s on this machine", argv[0])
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	args := []string{"-q", "/dev/null", "env", "-i",
		"HOME=" + home, "TERM=xterm-256color", "PATH=/usr/bin:/bin", "LANG=en_US.UTF-8"}
	args = append(append(args, env...), argv...)
	cmd := exec.CommandContext(ctx, "script", args...)
	cmd.Dir = home
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var out bytes.Buffer
	cmd.Stdout = writerFunc(func(p []byte) (int, error) {
		mu.Lock()
		defer mu.Unlock()
		return out.Write(p)
	})
	output := func() string {
		mu.Lock()
		defer mu.Unlock()
		return out.String()
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	for i, line := range append(lines, "exit") {
		for strings.Count(output(), shellPrompt) <= i {
			if ctx.Err() != nil {
				_ = cmd.Wait()
				t.Fatalf("the shell drew no prompt %d, output: %q", i+1, output())
			}
			time.Sleep(10 * time.Millisecond)
		}
		if _, err := stdin.Write([]byte(line + "\n")); err != nil {
			t.Fatal(err)
		}
	}
	_ = cmd.Wait()
	return output()
}

type writerFunc func(p []byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

// osc7 is the directories the shell reported in out, as it wrote them:
// the path of each file URL, still percent-encoded.
func osc7(out string) []string {
	var paths []string
	for _, m := range regexp.MustCompile(`\x1b\]7;file://[^/\x07]*(/[^\x07]*)\x07`).FindAllStringSubmatch(out, -1) {
		paths = append(paths, m[1])
	}
	return paths
}

// zshReport is a line the tests type to learn what the shell sees: its
// ZDOTDIR, whether the integration's own variable leaked into it, and
// which of the user's startup files ran. The quotes inside the first word
// keep the echo of the typed line from matching the answer.
const zshReport = `print -r -- "R""EPORT zdotdir=[${ZDOTDIR-unset}] leak=[${NUTTY_ZSH_ZDOTDIR-unset}] read=[$T_ENV $T_PROFILE $T_RC]"`

// TestZshIntegration starts the real zsh as a session does, with a stock
// configuration, and checks that it reports its directory when it starts
// and after a cd — which the zsh of macOS does for Terminal.app only, so a
// session started from the Finder never followed its shell — and that the
// user's startup files and ZDOTDIR are what they would be without Nutty.
func TestZshIntegration(t *testing.T) {
	a := &app{settings: settings.Default()}
	rc := "T_RC=zshrc\nPS1='" + shellPrompt + "'\n"

	t.Run("stock", func(t *testing.T) {
		home := integrationHome(t, map[string]string{
			".zshenv":       "T_ENV=zshenv\n",
			".zprofile":     "T_PROFILE=zprofile\n",
			".zshrc":        rc,
			"a b 文/.keep":   "",
			"sub/dir/.keep": "",
		})
		out := shellSession(t, home, a.shellEnv("/bin/zsh -l", nil), []string{"/bin/zsh", "-l"},
			`cd "a b 文"`, `cd ~/sub && cd dir && true`, zshReport)
		want := []string{home, home + "/a%20b%20%E6%96%87", home + "/sub", home + "/sub/dir"}
		for _, dir := range want {
			if !slices.Contains(osc7(out), dir) {
				t.Errorf("zsh did not report %s, reported %q", dir, osc7(out))
			}
		}
		if want := "REPORT zdotdir=[unset] leak=[unset] read=[zshenv zprofile zshrc]"; !strings.Contains(out, want) {
			t.Errorf("want %q in the shell's output: %q", want, out)
		}
	})

	// A ZDOTDIR of the user's, from the settings' environment: the shell
	// must read its startup files there and find it set again.
	t.Run("user ZDOTDIR", func(t *testing.T) {
		home := integrationHome(t, map[string]string{
			".zshenv":      "T_ENV=wrong-zshenv\n",
			"zdot/.zshenv": "T_ENV=zshenv\n",
			"zdot/.zshrc":  rc,
		})
		env := a.shellEnv("/bin/zsh", []string{"ZDOTDIR=" + home + "/zdot"})
		out := shellSession(t, home, env, []string{"/bin/zsh"}, "cd /", zshReport)
		if got := osc7(out); !slices.Contains(got, home) || !slices.Contains(got, "/") {
			t.Errorf("zsh reported %q, want %s and /", got, home)
		}
		if want := "REPORT zdotdir=[" + home + "/zdot] leak=[unset] read=[zshenv  zshrc]"; !strings.Contains(out, want) {
			t.Errorf("want %q in the shell's output: %q", want, out)
		}
	})

	// A ZDOTDIR the user's own .zshenv sets, as many configurations do.
	t.Run("ZDOTDIR set in .zshenv", func(t *testing.T) {
		home := integrationHome(t, map[string]string{
			".zshenv":     "T_ENV=zshenv\nZDOTDIR=$HOME/zdot\n",
			".zshrc":      "T_RC=wrong-zshrc\n",
			"zdot/.zshrc": rc,
		})
		out := shellSession(t, home, a.shellEnv("", nil), []string{"/bin/zsh"}, zshReport)
		if got := osc7(out); !slices.Contains(got, home) {
			t.Errorf("zsh reported %q, want %s", got, home)
		}
		if want := "REPORT zdotdir=[" + home + "/zdot] leak=[unset] read=[zshenv  zshrc]"; !strings.Contains(out, want) {
			t.Errorf("want %q in the shell's output: %q", want, out)
		}
	})

	// Options of the user's that change how zsh reads a script, and a
	// standard output that points elsewhere while the prompt is drawn, as
	// prompts that draw early leave it: the report still reaches the
	// terminal, and only the terminal.
	t.Run("user options and redirected stdout", func(t *testing.T) {
		home := integrationHome(t, map[string]string{
			".zshrc": rc + "setopt ksh_arrays sh_word_split no_unset\nexec 1>$HOME/stdout.txt\n",
		})
		out := shellSession(t, home, a.shellEnv("", nil), []string{"/bin/zsh"}, "cd /", "print done")
		if got := osc7(out); !slices.Contains(got, home) || !slices.Contains(got, "/") {
			t.Errorf("zsh reported %q, want %s and /", got, home)
		}
		if stdout, err := os.ReadFile(filepath.Join(home, "stdout.txt")); err != nil || string(stdout) != "done\n" {
			t.Errorf("the redirected stdout holds %q (%v), want only what the user printed", stdout, err)
		}
	})

	// A script run by the session's zsh is not interactive: it must get
	// the user's .zshenv and ZDOTDIR, and no report in its output.
	t.Run("not interactive", func(t *testing.T) {
		if _, err := os.Stat("/bin/zsh"); err != nil {
			t.Skip("no /bin/zsh on this machine")
		}
		home := integrationHome(t, map[string]string{".zshenv": "T_ENV=zshenv\n"})
		cmd := exec.Command("/bin/zsh", "-c", `print -r -- "zdotdir=[${ZDOTDIR-unset}] read=[$T_ENV]"`)
		cmd.Env = append([]string{"HOME=" + home, "PATH=/usr/bin:/bin"}, a.shellEnv("/bin/zsh", nil)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err, string(out))
		}
		if got, want := string(out), "zdotdir=[unset] read=[zshenv]\n"; got != want {
			t.Errorf("zsh -c printed %q, want %q", got, want)
		}
	})
}

// TestBashIntegration starts macOS's bash as a session does and checks
// that it reports its directory, and that the hook, which reaches bash as
// PROMPT_COMMAND, leaves a user's own PROMPT_COMMAND working whichever way
// the startup files set it: added after ours, before it, or in its place.
func TestBashIntegration(t *testing.T) {
	a := &app{settings: settings.Default()}
	ps1 := "PS1='" + shellPrompt + "'\n"
	for name, tc := range map[string]struct {
		profile  string // the user's .bash_profile, after PS1
		reports  bool   // whether the directory is still reported
		wantText string // what the user's PROMPT_COMMAND prints after "false"
	}{
		"stock":       {"", true, ""},
		"nounset":     {"set -u\n", true, ""},
		"appended":    {`PROMPT_COMMAND="$PROMPT_COMMAND"'; echo "user saw $?"'` + "\n", true, "user saw 1"},
		"prepended":   {`PROMPT_COMMAND='echo "user saw $?"; '"$PROMPT_COMMAND"` + "\n", true, "user saw 1"},
		"overwritten": {`PROMPT_COMMAND='echo "user saw $?"'` + "\n", false, "user saw 1"},
	} {
		t.Run(name, func(t *testing.T) {
			home := integrationHome(t, map[string]string{
				".bash_profile": ps1 + tc.profile,
				"a b 文/.keep":   "",
			})
			out := shellSession(t, home, a.shellEnv("/bin/bash -l", nil), []string{"/bin/bash", "-l"},
				`cd "a b 文"`, "false", "true")
			got := osc7(out)
			if tc.reports && (!slices.Contains(got, home) || !slices.Contains(got, home+"/a%20b%20%E6%96%87")) {
				t.Errorf("bash reported %q, want %s and its a%%20b%%20%%E6%%96%%87", got, home)
			}
			if !tc.reports && len(got) != 0 {
				t.Errorf("bash reported %q with the user's PROMPT_COMMAND in place of ours", got)
			}
			if !strings.Contains(out, tc.wantText) {
				t.Errorf("want %q from the user's PROMPT_COMMAND: %q", tc.wantText, out)
			}
			for _, bad := range []string{"syntax error", "command not found", "unbound variable"} {
				if strings.Contains(out, bad) {
					t.Errorf("bash complained (%s): %q", bad, out)
				}
			}
		})
	}
}

// TestShellEnv checks which shells get the integration and what it adds
// to their environment, and that a session without it runs as before.
func TestShellEnv(t *testing.T) {
	home := integrationHome(t, nil)
	a := &app{settings: settings.Default()}
	user := []string{"FOO=bar"}

	t.Setenv("SHELL", "/bin/zsh")
	for _, command := range []string{"", "/bin/zsh", "/opt/homebrew/bin/zsh -l"} {
		env := a.shellEnv(command, user)
		dir, ok := envValue(env, "ZDOTDIR")
		if !ok || !strings.HasPrefix(dir, home) || !slices.Contains(env, "FOO=bar") {
			t.Fatalf("command %q: env = %q, want FOO and a ZDOTDIR under %s", command, env, home)
		}
		if got, err := os.ReadFile(filepath.Join(dir, ".zshenv")); err != nil || !bytes.Equal(got, zshIntegration) {
			t.Fatalf("command %q: the integration is not at %s/.zshenv: %v", command, dir, err)
		}
	}
	if len(user) != 1 {
		t.Fatalf("shellEnv changed the caller's list: %q", user)
	}
	// The user's ZDOTDIR travels beside ours, to be put back.
	if env := a.shellEnv("zsh", []string{"ZDOTDIR=/u/zdot"}); !slices.Contains(env, "NUTTY_ZSH_ZDOTDIR=/u/zdot") {
		t.Errorf("env = %q, want the user's ZDOTDIR kept in NUTTY_ZSH_ZDOTDIR", env)
	}

	// bash gets the hook before a PROMPT_COMMAND of the user's.
	env := a.shellEnv("bash", []string{"PROMPT_COMMAND=history -a"})
	if pc, _ := envValue(env, "PROMPT_COMMAND"); pc != bashIntegration+"; history -a" {
		t.Errorf("PROMPT_COMMAND = %q", pc)
	}

	// Everything else runs as it is: a shell that reports by itself, a
	// program that is no shell, a zsh that reads no startup files, and
	// any shell with the setting off.
	t.Setenv("SHELL", "/opt/homebrew/bin/fish")
	for _, command := range []string{"", "ssh host", "/bin/zsh -f", "zsh --no-rcs"} {
		if env := a.shellEnv(command, user); !slices.Equal(env, user) {
			t.Errorf("command %q: env = %q, want it unchanged", command, env)
		}
	}
	a.settings.ShellIntegration = false
	for _, command := range []string{"/bin/zsh", "/bin/bash"} {
		if env := a.shellEnv(command, user); !slices.Equal(env, user) {
			t.Errorf("setting off, command %q: env = %q, want it unchanged", command, env)
		}
	}
}

// TestSessionFollowsShell runs zsh in a real terminal, as a session does,
// and checks the whole chain: the shell reports where it is, the terminal
// reads the report, and the session's directory follows a cd.
func TestSessionFollowsShell(t *testing.T) {
	if _, err := os.Stat("/bin/zsh"); err != nil {
		t.Skip("no /bin/zsh on this machine")
	}
	registerFonts()
	home := integrationHome(t, map[string]string{"sub/.keep": ""})
	a := &app{settings: settings.Default()}
	term, err := terminal.New(terminal.Options{
		Command: []string{"/bin/zsh"},
		Dir:     home,
		Env:     a.shellEnv("/bin/zsh", nil),
		Font:    a.termFont(),
		Theme:   settings.OttyTermTheme(false),
	})
	if err != nil {
		t.Skip("no terminal to run a shell in:", err)
	}
	defer func() { _ = term.Close() }()
	p := &pane{term: term}
	s := &session{id: 1, dir: "/start", root: &node{term: p}, panes: []*pane{p}}
	a.sessions, a.activePane = []*session{s}, p
	follows := func(want string) {
		t.Helper()
		for deadline := time.Now().Add(20 * time.Second); s.dir != want; time.Sleep(10 * time.Millisecond) {
			if time.Now().After(deadline) {
				t.Fatalf("the session is in %q, want %q", s.dir, want)
			}
			a.trackWorkingDir()
		}
	}
	follows(home)
	term.Send([]byte("cd sub\n"))
	follows(home + "/sub")
	// Let the shell end by itself, and wait for it: hung up when the
	// terminal closes, it would still be saving its history into the
	// home directory the test is about to remove.
	term.Send([]byte("exit\n"))
	select {
	case <-term.Done():
	case <-time.After(20 * time.Second):
		t.Error("the shell did not exit")
	}
}
