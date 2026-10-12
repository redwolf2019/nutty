package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Agent is the backend of the agent chat. demoAgent streams a canned
// markdown reply; to wire a real model, replace Reply with a call to it
// and keep streaming its text into out, or return an error string in
// markdown.
type Agent interface {
	// Reply streams markdown chunks of a reply to prompt, from cwd, and
	// closes out when it finishes. It must return promptly when ctx is
	// cancelled.
	Reply(ctx context.Context, cwd, prompt string, out chan<- string)
	// Stop cancels the reply in flight.
	Stop()
}

// demoAgent answers with a small markdown page. It streams the text in
// chunks so the chat's writing caret and thinking indicator work exactly
// as they will with a real model.
type demoAgent struct {
	mu     sync.Mutex
	cancel context.CancelFunc
}

func (a *demoAgent) Reply(ctx context.Context, cwd, prompt string, out chan<- string) {
	ctx, cancel := context.WithCancel(ctx)
	a.mu.Lock()
	a.cancel = cancel
	a.mu.Unlock()
	go func() {
		defer close(out)
		md := a.answer(cwd, prompt)
		for _, chunk := range chunks(md, 24) {
			select {
			case <-ctx.Done():
				return
			case out <- chunk:
			}
			time.Sleep(14 * time.Millisecond)
		}
	}()
}

func (a *demoAgent) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}
}

// answer is the demo's markdown reply for a prompt.
func (a *demoAgent) answer(cwd, prompt string) string {
	p := strings.TrimSpace(prompt)
	switch {
	case p == "/help":
		return "## 我能做什么\n\n" +
			"这是一个演示 Agent —— UI 骨架已就位，接入真实 LLM 只需替换 `agent.go` 里的 `Reply`。\n\n" +
			"- `/readme` —— 渲染当前目录的 README\n" +
			"- `/status` —— 看看当前目录\n" +
			"- 其他输入 —— 得到一张 markdown 卡片\n\n" +
			"试试 `/readme`。"
	case p == "/status":
		return "## 状态\n\n" +
			"- 目录: `" + cwd + "`\n" +
			"- 模型: `demo`（内置演示）\n\n" +
			"> 渲染结果由 onevra 的 markdown 插件绘制。"
	case p == "/readme":
		md, err := os.ReadFile(readmePath(cwd))
		if err != nil {
			return "## 没有找到 README\n\n在 `" + cwd + "` 下没有 `README.md`。\n\n> 写一个吧，这个面板会把它渲染成文档。"
		}
		return "## README\n\n" + string(md)
	default:
		return "## 你的提示\n\n> " + strings.ReplaceAll(p, "\n", " ") + "\n\n" +
			"这里是 **markdown 渲染** 的演示：\n\n" +
			"1. 列表\n2. `行内代码`\n3. 代码块\n\n" +
			"```go\n" +
			"func hello() string { return \"🌰\" }\n" +
			"```\n\n" +
			"---\n\n" +
			"接入真实后端后，这里会显示模型的回复。输入 `/help` 查看更多。"
	}
}

// readmePath is the first README found in dir.
func readmePath(dir string) string {
	for _, name := range []string{"README.md", "readme.md", "README.markdown"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return filepath.Join(dir, "README.md")
}

// chunks splits s into pieces of at most n runes, keeping spaces whole.
func chunks(s string, n int) []string {
	var out []string
	runes := []rune(s)
	for len(runes) > 0 {
		take := n
		if take > len(runes) {
			take = len(runes)
		}
		out = append(out, string(runes[:take]))
		runes = runes[take:]
	}
	return out
}
