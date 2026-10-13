package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// longReadme makes a directory whose README takes the demo agent about a
// minute to stream, so a test can stop the reply while it is under way.
func longReadme(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	text := strings.Repeat("a line of the README, long enough to take a few chunks to stream\n", 1500)
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// drain reads a reply to its end, which must come within wait.
func drain(t *testing.T, out <-chan string, wait time.Duration) string {
	t.Helper()
	var reply strings.Builder
	deadline := time.After(wait)
	for {
		select {
		case chunk, ok := <-out:
			if !ok {
				return reply.String()
			}
			reply.WriteString(chunk)
		case <-deadline:
			t.Fatalf("the reply did not end within %v, after %d bytes", wait, reply.Len())
		}
	}
}

// TestDemoAgentReply checks what the chat gets from Reply: the answer,
// whole and in order, in chunks, and then the end of the channel.
func TestDemoAgentReply(t *testing.T) {
	a := &demoAgent{}
	out := make(chan string, 16)
	a.Reply(context.Background(), t.TempDir(), "/help", out)
	if got, want := drain(t, out, 20*time.Second), a.answer("", "/help"); got != want {
		t.Fatalf("the reply streamed %q, want %q", got, want)
	}
	// Stop after the reply ended is harmless, twice too.
	a.Stop()
	a.Stop()
}

// TestDemoAgentStop checks that Stop ends the reply in flight: the
// channel closes at once, well before the answer was streamed.
func TestDemoAgentStop(t *testing.T) {
	dir := longReadme(t)
	a := &demoAgent{}
	out := make(chan string, 16)
	a.Reply(context.Background(), dir, "/readme", out)
	if _, ok := <-out; !ok {
		t.Fatal("the reply ended before its first chunk")
	}
	a.Stop()
	if got, whole := drain(t, out, 5*time.Second), a.answer(dir, "/readme"); len(got) >= len(whole)/2 {
		t.Fatalf("the reply streamed %d of %d bytes after Stop", len(got), len(whole))
	}
}

// TestDemoAgentContext checks the other way a reply ends early: Reply
// must return promptly when its context is cancelled, as the Agent
// interface asks of every backend.
func TestDemoAgentContext(t *testing.T) {
	dir := longReadme(t)
	a := &demoAgent{}
	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan string, 16)
	a.Reply(ctx, dir, "/readme", out)
	if _, ok := <-out; !ok {
		t.Fatal("the reply ended before its first chunk")
	}
	cancel()
	if got, whole := drain(t, out, 5*time.Second), a.answer(dir, "/readme"); len(got) >= len(whole)/2 {
		t.Fatalf("the reply streamed %d of %d bytes after its context was cancelled", len(got), len(whole))
	}
}
