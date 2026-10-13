package app

import (
	"encoding/hex"
	"slices"
	"strings"
	"testing"
)

// zshHistoryHex is a history file as zsh 5.9 wrote it with EXTENDED_HISTORY
// (print -s …; fc -W), in hex since it is not valid UTF-8: a command with a
// second ";" in it, one with Chinese in it, whose bytes zsh metafied, one
// that spans three lines, one that ends in a backslash, which zsh saves
// with a space after it, and a plain one.
const zshHistoryHex = "" +
	"3a20313739313837343033313a303b" + "6364207765623b206e706d2072756e20646576" + "0a" +
	"3a20313739313837343033313a303b" + "67697420636f6d6d6974202d6d20" + "e4bfaee5a483ade683b683a7e68381a3" + "0a" +
	"3a20313739313837343033313a303b" + "666f72206620696e202a3b20646f5c0a" + "20206563686f2024665c0a" + "646f6e65" + "0a" +
	"3a20313739313837343033313a303b" + "6563686f207461696c5c20" + "0a" +
	"3a20313739313837343033313a303b" + "6c73" + "0a"

// TestParseZshHistory checks what the Outline section reads from the
// shell's history: each command whole, with the text the user typed.
func TestParseZshHistory(t *testing.T) {
	file, err := hex.DecodeString(zshHistoryHex)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"cd web; npm run dev", // not "npm run dev": the stamp ends at the first ";"
		"git commit -m 修复文档",
		"for f in *; do\n  echo $f\ndone", // one command, not three
		`echo tail\`,
		"ls",
	}
	if got := parseZshHistory(file); !slices.Equal(got, want) {
		t.Errorf("extended history:\n got %q\nwant %q", got, want)
	}

	// Without EXTENDED_HISTORY there is no stamp, and zsh escapes the colon
	// of a command that starts with one.
	plain := "cd web; npm run dev\n" + `\: > out; echo done` + "\n\n"
	want = []string{"cd web; npm run dev", ": > out; echo done"}
	if got := parseZshHistory([]byte(plain)); !slices.Equal(got, want) {
		t.Errorf("plain history:\n got %q\nwant %q", got, want)
	}
}

// TestListeningPorts checks what the Ports section reads from lsof: each
// port once, whichever address and family it listens on, in the order of
// the numbers rather than of the strings.
func TestListeningPorts(t *testing.T) {
	out := strings.Join([]string{
		"COMMAND     PID USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME",
		"rapportd   1083 me    11u  IPv4 0x235d79cda004009e      0t0  TCP *:55516 (LISTEN)",
		"rapportd   1083 me    12u  IPv6 0xf8a906e637010f56      0t0  TCP *:55516 (LISTEN)",
		"ControlCe  1229 me     9u  IPv4  0xf3965d74576dadf      0t0  TCP *:7000 (LISTEN)",
		"ControlCe  1229 me    11u  IPv4 0xd0f5cc17bddb23f8      0t0  TCP *:5000 (LISTEN)",
		`Code\x20H 64983 me    42u  IPv4 0xec4cca769636a413      0t0  TCP 127.0.0.1:10000 (LISTEN)`,
		"node      70001 me    23u  IPv6 0x7c8d9ddb37f75b08      0t0  TCP [::1]:8080 (LISTEN)",
		"node      70001 me    24u  IPv4 0x7c8d9ddb37f75b09      0t0  TCP 127.0.0.1:8080 (LISTEN)",
		"",
	}, "\n")
	want := []string{"5000", "7000", "8080", "10000", "55516"}
	if got := listeningPorts(out); !slices.Equal(got, want) {
		t.Errorf("ports = %v, want %v", got, want)
	}
	if got := listeningPorts(""); len(got) != 0 {
		t.Errorf("ports of no output = %v, want none", got)
	}
}
