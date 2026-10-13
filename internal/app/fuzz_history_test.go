package app

import (
	"bytes"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// FuzzParseZshHistory feeds the Outline section any history file — the
// file holds whatever was typed, and zsh's own encoding of it — and
// checks the commands it reads: no panic, no empty command, none with
// space around it, and no more commands than the file has lines.
func FuzzParseZshHistory(f *testing.F) {
	if real, err := hex.DecodeString(zshHistoryHex); err == nil {
		f.Add(real)
	}
	for _, file := range []string{
		": 1700000000:0;ls\n", "ls\ncd web; npm run dev\n", `\: > out`, ": \n", ": 1:2", ": 1:2;", ";", "\\", "\\\n", "a\\\n\\\n",
		"\x83", "\x83\n", "ls \x83", ": 17\x83\x1a:0;x", "", "\n\n\n",
	} {
		f.Add([]byte(file))
	}
	f.Fuzz(func(t *testing.T, file []byte) {
		cmds := parseZshHistory(file)
		if lines := bytes.Count(file, []byte("\n")) + 1; len(cmds) > lines {
			t.Fatalf("%d commands from %d lines", len(cmds), lines)
		}
		for _, cmd := range cmds {
			if cmd == "" || cmd != strings.TrimSpace(cmd) {
				t.Fatalf("command %q is empty or has space around it", cmd)
			}
		}
	})
}

// FuzzListeningPorts feeds the Ports section any lsof output and checks
// the ports it lists: no panic, each port once, none empty, in the order
// of their numbers.
func FuzzListeningPorts(f *testing.F) {
	for _, out := range []string{
		"COMMAND     PID USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME\n" +
			"rapportd   1083 me    11u  IPv4 0x235d79cda004009e      0t0  TCP *:55516 (LISTEN)\n" +
			"node      70001 me    23u  IPv6 0x7c8d9ddb37f75b08      0t0  TCP [::1]:8080 (LISTEN)\n" +
			`Code\x20H 64983 me    42u  IPv4 0xec4cca769636a413      0t0  TCP 127.0.0.1:10000 (LISTEN)` + "\n",
		"a b c d e f g h *:http (LISTEN)\n", "a b c d e f g h [::]:* x\n", "a b c d e f g h :99999999999999999999\n",
		"a b c d e f g h -1:-1\na b c d e f g h x:+7\n", "", "\n", "1 2 3 4 5 6 7 8 9",
	} {
		f.Add(out)
	}
	f.Fuzz(func(t *testing.T, out string) {
		ports := listeningPorts(out)
		seen := map[string]bool{}
		for _, p := range ports {
			if p == "" || seen[p] {
				t.Fatalf("ports %q hold an empty or a repeated port", ports)
			}
			seen[p] = true
		}
		if !sort.SliceIsSorted(ports, func(i, j int) bool {
			a, _ := strconv.Atoi(ports[i])
			b, _ := strconv.Atoi(ports[j])
			return a < b
		}) {
			t.Fatalf("ports %q are not in the order of their numbers", ports)
		}
	})
}
