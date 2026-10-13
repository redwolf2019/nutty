package app

import (
	"context"
	"math"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components/markdown"
	"github.com/smallnest/uix/components/composer"
	"github.com/smallnest/uix/components/thinking"
)

// chatState is the agent chat panel: the transcript, the composer and
// where the scroll is.
type chatState struct {
	msgs    []chatMsg
	value   string
	busy    bool
	turn    int // counts the prompts sent; the reply of the last one owns busy
	started time.Time
	scroll  ui.ScrollState
}

// chatMsg is one line of the transcript. The assistant's markdown parses
// once per change and renders in the panel.
type chatMsg struct {
	user      bool
	text      string
	streaming bool
	md        *markdown.MarkdownDocument
	mdLen     int
}

// chatView is the agent panel that slides in on the right. Its header —
// the nut, the "Agent" title and the close button — lives in the top
// strip's toolbar.
func (a *app) chatView(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Width(360).FillHeight().Shrink(0).Background(t.Surface).
		BorderWidth(0, 0, 0, 1).BorderColor(t.Border).Children(func() {
		a.chatBody(c)
	})
}

// chatColumn is the agent panel of the Tabs layouts: a full-height column
// that reaches the top of the window, with the header at the strip's
// height and the transcript below, as otty draws the panel beside the
// horizontal tab bar.
func (a *app) chatColumn(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Width(360).FillHeight().Shrink(0).Background(t.Surface).
		BorderWidth(0, 0, 0, 1).BorderColor(t.Border).Children(func() {
		ui.Box(c).FillWidth().Shrink(0).Height(a.stripBar(c)).Children(func() {
			a.chatToolbarRow(c)
		})
		a.chatBody(c)
	})
}

// chatBody is the panel's transcript and composer.
func (a *app) chatBody(c *ui.Context) {
	t := c.Theme()
	ui.Scroll(c).Grow(1).MinHeight(0).TrackScroll(&a.chat.scroll).
		Children(func() {
			ui.Column(c).FillWidth().Padding(12, 14).Gap(12).Children(func() {
				if len(a.chat.msgs) == 0 {
					ui.Column(c).FillWidth().Gap(8).Padding(8, 0).Children(func() {
						nutTile(c, 56)
						ui.Text(c, "Agent 面板").TextColor(t.Text).Bold()
						ui.Text(c, "这是一个 UI 骨架：回复由演示 Agent 生成，markdown 用 onevra 渲染。接入真实模型时替换 agent.go。").TextColor(t.TextMuted).FontSize(12)
					})
				}
				for i := range a.chat.msgs {
					a.chatBubble(c, &a.chat.msgs[i])
				}
				if a.chat.busy && a.chat.thinking() {
					ui.Box(c).Padding(8, 0).Children(func() {
						thinking.Thinking(c, thinking.Props{
							Label:     tr("Thinking"),
							ShowTimer: true,
							Started:   a.chat.started,
						})
					})
				}
			})
		})
	ui.Box(c).Shrink(0).Padding(12, 14).Children(func() {
		composer.Composer(c, composer.Props{
			Value:        &a.chat.value,
			Busy:         a.chat.busy,
			OnSubmit:     a.chatSubmit,
			OnStop:       a.chatStop,
			Model:        "demo",
			Provider:     "built-in",
			MessageCount: len(a.chat.msgs),
			Placeholder:  tr("Ask me anything"),
		})
	})
}

// chatBubble draws one turn: the user's in an accent bubble, the
// assistant's as markdown.
func (a *app) chatBubble(c *ui.Context, m *chatMsg) {
	t := c.Theme()
	if m.user {
		ui.Row(c).FillWidth().Justify(ui.End).Children(func() {
			ui.Box(c).MaxWidthPercent(82).Padding(9, 13).
				Radius(t.Radius + 4).Background(t.Accent).Children(func() {
				ui.Text(c, m.text).TextColor(t.AccentText)
			})
		})
		return
	}
	if m.text == "" && !m.streaming {
		return
	}
	ui.Column(c).FillWidth().Children(func() {
		if m.md == nil || len(m.text) != m.mdLen {
			m.md = markdown.ParseMarkdown(m.text)
			m.mdLen = len(m.text)
		}
		if m.md != nil {
			m.md.Render(c)
		}
		if m.streaming {
			ui.Box(c).Size(2, t.FontSize).Shrink(0).Draw(func(pp *ui.Painter, r ui.Rect) {
				phase := float64(pp.Now().UnixMilli()%800) / 800
				alpha := float32(math.Abs(1 - 2*phase))
				pp.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W, H: r.H}, t.Text.Alpha(alpha), 1)
				pp.After(80 * time.Millisecond)
			})
		}
	})
}

// thinking reports whether the assistant is about to write its first word.
func (s *chatState) thinking() bool {
	return len(s.msgs) == 0 || s.msgs[len(s.msgs)-1].text == ""
}

// chatSubmit sends the composer's text to the agent. The composer submits
// on Enter while a reply is still streaming too: that reply is cut short,
// and each reply keeps writing to its own message.
func (a *app) chatSubmit(text string) {
	dir := ""
	if s := a.currentSession(); s != nil {
		dir = s.dir
	}
	a.agent.Stop()
	a.chat.msgs = append(a.chat.msgs, chatMsg{user: true, text: text})
	a.chat.msgs = append(a.chat.msgs, chatMsg{})
	idx := len(a.chat.msgs) - 1
	a.chat.turn++
	turn := a.chat.turn
	a.chat.busy = true
	a.chat.started = time.Now()
	a.chat.value = ""
	out := make(chan string, 16)
	a.agent.Reply(context.Background(), dir, text, out)
	go func() {
		for chunk := range out {
			a.win.Update(func() { a.chatAppend(idx, chunk) })
		}
		a.win.Update(func() { a.chatFinish(idx, turn) })
	}()
}

// chatAppend adds a chunk of a reply to the reply's message.
func (a *app) chatAppend(idx int, chunk string) {
	m := &a.chat.msgs[idx]
	m.text += chunk
	m.streaming = true
	// Follow the end while the reader is already there.
	if a.chat.scroll.Y >= a.chat.scroll.MaxY {
		a.chat.scroll.Y = math.MaxFloat32
	}
}

// chatFinish ends a reply. The panel stays busy when a later prompt's
// reply is still coming.
func (a *app) chatFinish(idx, turn int) {
	a.chat.msgs[idx].streaming = false
	if turn == a.chat.turn {
		a.chat.busy = false
	}
}

// chatStop cancels the reply in flight.
func (a *app) chatStop() {
	a.agent.Stop()
}
