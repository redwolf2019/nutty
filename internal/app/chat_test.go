package app

import (
	"context"
	"testing"
)

// heldAgent is an agent whose replies never come: the test plays them.
type heldAgent struct{ stops int }

func (h *heldAgent) Reply(context.Context, string, string, chan<- string) {}
func (h *heldAgent) Stop()                                                { h.stops++ }

// TestChatSubmitWhileStreaming checks a prompt sent while a reply is still
// streaming: the first reply is stopped, what is left of it lands in its
// own message, and the panel stays busy until the second reply ends.
func TestChatSubmitWhileStreaming(t *testing.T) {
	a := testApp()
	agent := &heldAgent{}
	a.agent = agent
	a.chatSubmit("one")
	a.chatAppend(1, "first ")
	a.chatSubmit("two")
	if agent.stops != 2 {
		t.Errorf("Stop ran %d times, want once a prompt", agent.stops)
	}
	a.chatAppend(1, "reply") // a chunk of the first reply that was on its way
	a.chatAppend(3, "second reply")
	if got := a.chat.msgs[1].text; got != "first reply" {
		t.Errorf("first reply = %q", got)
	}
	if got := a.chat.msgs[3].text; got != "second reply" {
		t.Errorf("second reply = %q", got)
	}
	a.chatFinish(1, 1)
	if a.chat.msgs[1].streaming || !a.chat.busy {
		t.Errorf("after the first reply ended: streaming=%v busy=%v, want false true",
			a.chat.msgs[1].streaming, a.chat.busy)
	}
	a.chatFinish(3, 2)
	if a.chat.msgs[3].streaming || a.chat.busy {
		t.Errorf("after the second reply ended: streaming=%v busy=%v, want false false",
			a.chat.msgs[3].streaming, a.chat.busy)
	}
}
