package main

import "testing"

// TestParseGitStatus checks what the Git section reads from git status
// --porcelain=v1 -b: the branch, and one change per line with its raw
// two-letter status kept so staged and unstaged stay distinct.
func TestParseGitStatus(t *testing.T) {
	branch, changes := parseGitStatus("## main...origin/main\n" +
		" M session.go\n" +
		"M  topbar.go\n" +
		"?? settings.go\n" +
		"A  info.go\n" +
		"MM both.go\n")
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
	branch, _ := parseGitStatus("## HEAD (no branch)\n")
	if branch != "HEAD" {
		t.Fatalf("branch = %q, want %q", branch, "HEAD")
	}
	branch, _ = parseGitStatus("## feature/x [ahead 2]\n")
	if branch != "feature/x" {
		t.Fatalf("branch = %q, want %q", branch, "feature/x")
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
