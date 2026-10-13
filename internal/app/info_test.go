package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestParseGitStatus checks what the Git section reads from git status
// --porcelain=v1 -b -z: the branch, and one change per entry with its raw
// two-letter status kept so staged and unstaged stay distinct.
func TestParseGitStatus(t *testing.T) {
	branch, changes := parseGitStatus("## main...origin/main\x00" +
		" M session.go\x00" +
		"M  topbar.go\x00" +
		"?? settings.go\x00" +
		"A  info.go\x00" +
		"MM both.go\x00")
	if branch != "main" {
		t.Fatalf("branch = %q, want %q", branch, "main")
	}
	want := []gitChange{
		{" M", "session.go"},
		{"M ", "topbar.go"},
		{"??", "settings.go"},
		{"A ", "info.go"},
		{"MM", "both.go"},
	}
	if len(changes) != len(want) {
		t.Fatalf("changes = %v, want %v", changes, want)
	}
	for i, w := range want {
		if changes[i] != w {
			t.Errorf("changes[%d] = %v, want %v", i, changes[i], w)
		}
	}
}

// TestParseGitStatusDetached checks the branch shown for a detached HEAD,
// and that a header with ahead/behind notes still parses.
func TestParseGitStatusDetached(t *testing.T) {
	branch, _ := parseGitStatus("## HEAD (no branch)\x00")
	if branch != "HEAD" {
		t.Fatalf("branch = %q, want %q", branch, "HEAD")
	}
	branch, _ = parseGitStatus("## feature/x [ahead 2]\x00")
	if branch != "feature/x" {
		t.Fatalf("branch = %q, want %q", branch, "feature/x")
	}
	branch, _ = parseGitStatus("## No commits yet on release-1.2\x00")
	if branch != "release-1.2" {
		t.Fatalf("branch = %q, want %q", branch, "release-1.2")
	}
}

// TestGitTabInSubdirectory runs the Git section's two commands in a real
// repository, from a directory below its root: every change is listed
// under the path git knows it by, a rename as its new path alone, and the
// diff of each tracked change is not empty.
func TestGitTabInSubdirectory(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	root := t.TempDir()
	sub := filepath.Join(root, "sub", "dir")
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_SYSTEM="+os.DevNull,
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
		return string(out)
	}
	write := func(name, text string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	git("-C", root, "init", "-q", "-b", "main")
	names := []string{"sub/dir/tracked.txt", "has space.txt", "中文.txt", "old.txt"}
	for _, n := range names {
		write(n, "one\n")
	}
	git("-C", root, "add", "-A")
	git("-C", root, "commit", "-q", "-m", "init")
	for _, n := range names[:3] {
		write(n, "one\ntwo\n")
	}
	git("-C", root, "mv", "old.txt", "new.txt")

	branch, changes := parseGitStatus(git(gitStatusArgs(sub)...))
	if branch != "main" {
		t.Errorf("branch = %q, want main", branch)
	}
	want := map[string]string{"sub/dir/tracked.txt": " M", "has space.txt": " M", "中文.txt": " M", "new.txt": "R "}
	if len(changes) != len(want) {
		t.Fatalf("changes = %q, want %d of them", changes, len(want))
	}
	for _, ch := range changes {
		if want[ch.path] != ch.status {
			t.Errorf("change %q %q, want one of %v", ch.status, ch.path, want)
		}
		if diff := git(gitDiffArgs(sub, ch.path, ch.status[0] != ' ')...); diff == "" {
			t.Errorf("%q: the diff is empty", ch.path)
		}
	}
}

// TestGitStatusLabel checks the letter a change shows, and that the raw
// status is what decides how the diff is loaded.
func TestGitStatusLabel(t *testing.T) {
	for in, want := range map[string]string{
		" M": "M",
		"M ": "M",
		"??": "?",
		"A ": "A",
		"MM": "M",
	} {
		if got := gitStatusLabel(in); got != want {
			t.Errorf("gitStatusLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestStaleInfoCommit checks when a Git or Files load's commit must be
// dropped: the active session switched while the load ran, or the session's
// directory moved on, so a newer load owns the panel's data.
func TestStaleInfoCommit(t *testing.T) {
	same := &session{id: 1, dir: "/a"}
	moved := &session{id: 2, dir: "/a"}
	moved.dir = "/b" // the directory moved on while the load ran
	other := &session{id: 3, dir: "/b"}
	tests := []struct {
		name   string
		cur    *session
		loaded *session
		dir    string
		want   bool
	}{
		{"directory unchanged", same, same, "/a", false},
		{"directory moved on", moved, moved, "/a", true},
		{"session switched", other, same, "/a", true},
		{"no session", nil, nil, "", false},
	}
	for _, tc := range tests {
		if got := staleInfoCommit(tc.cur, tc.loaded, tc.dir); got != tc.want {
			t.Errorf("%s: staleInfoCommit = %v, want %v", tc.name, got, tc.want)
		}
	}
}
