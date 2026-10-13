package app

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"nutty/internal/settings"
)

// The zsh half of the shell integration, written out as the .zshenv of a
// directory ZDOTDIR points zsh at (see shellEnv).
//
//go:embed assets/shell-integration.zsh
var zshIntegration []byte

// bashIntegration is the bash half, a PROMPT_COMMAND: bash has no startup
// file an environment variable can add without replacing the user's own,
// but it runs PROMPT_COMMAND from the environment before each prompt. The
// command defines its function each time, so it needs no file, and the
// function percent-encodes $PWD byte by byte, as macOS's own
// update_terminal_cwd does, with what bash 3.2 has. It returns the status
// it was handed, so whatever the user's startup files append still sees
// the status of the command line in $?, and it ends without a semicolon,
// so "$PROMPT_COMMAND; more" stays a valid command.
const bashIntegration = `__nutty_s=$?; ` +
	`__nutty_report_cwd() { ` +
	`local LC_CTYPE=C LC_COLLATE=C LC_ALL= i c h u=; ` +
	`for ((i = 0; i < ${#PWD}; i++)); do c=${PWD:i:1}; ` +
	`if [[ $c =~ [/._~A-Za-z0-9-] ]]; then u+=$c; ` +
	`else printf -v h %02X "'$c"; u+=%${h: -2:2}; fi; done; ` +
	`printf '\033]7;file://%s%s\a' "$HOSTNAME" "$u"; ` +
	`return "$1"; }; ` +
	`__nutty_report_cwd "$__nutty_s"`

// shellEnv is the environment of a new session's shell: env, the settings'
// and the profile's, with the shell integration added when the settings
// provide it. A stock zsh or bash on macOS reports its working directory
// (OSC 7) to Terminal.app only, so without the integration a session
// started from the Finder never learns where its shell went, and the tab,
// the title and the info panel stay at the directory the session started
// in. command is what the session runs, empty for the user's shell.
//
// zsh gets a ZDOTDIR of ours, whose .zshenv puts the user's ZDOTDIR back,
// reads the user's own .zshenv and adds the hook; bash gets the hook as
// PROMPT_COMMAND; fish reports its directory by itself. Any other
// program, and a zsh told to read no startup files, runs as it is.
func (a *app) shellEnv(command string, env []string) []string {
	if !a.settings.ShellIntegration {
		return env
	}
	shell, args := shellName(), settings.ShellArgs(command)
	if len(args) > 0 {
		shell = filepath.Base(args[0])
	}
	switch shell {
	case "zsh":
		if slices.Contains(args, "-f") || slices.Contains(args, "--no-rcs") {
			return env
		}
		dir, err := zshIntegrationDir()
		if err != nil {
			// Without the directory the shell starts as it always did.
			return env
		}
		env = slices.Clone(env)
		if user, ok := envValue(env, "ZDOTDIR"); ok {
			env = append(env, "NUTTY_ZSH_ZDOTDIR="+user)
		}
		return append(env, "ZDOTDIR="+dir)
	case "bash":
		hook := bashIntegration
		if user, _ := envValue(env, "PROMPT_COMMAND"); user != "" {
			hook += "; " + user
		}
		return append(slices.Clone(env), "PROMPT_COMMAND="+hook)
	}
	return env
}

// envValue is the value the shell would get for name: the last of env's,
// which the terminal lays over this process's environment, else this
// process's own.
func envValue(env []string, name string) (string, bool) {
	for i := len(env) - 1; i >= 0; i-- {
		if v, ok := strings.CutPrefix(env[i], name+"="); ok {
			return v, true
		}
	}
	return os.LookupEnv(name)
}

// zshIntegrationDir is the directory zsh's ZDOTDIR points at, in the
// user's cache directory, with the integration as its .zshenv. The file is
// written when it is missing or differs from this build's, through a
// rename, so a shell starting from another window never reads half of it.
func zshIntegrationDir() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cache, "Nutty", "shell-integration", "zsh")
	file := filepath.Join(dir, ".zshenv")
	if have, err := os.ReadFile(file); err == nil && bytes.Equal(have, zshIntegration) {
		return dir, nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(dir, ".zshenv-*")
	if err != nil {
		return "", err
	}
	_, err = tmp.Write(zshIntegration)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(tmp.Name(), file)
	}
	if err != nil {
		_ = os.Remove(tmp.Name())
		return "", err
	}
	return dir, nil
}
