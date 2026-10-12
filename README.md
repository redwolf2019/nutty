# Nutty 🌰

Nutty is a native terminal app for macOS. It is a clone of
[otty.sh](https://otty.sh/), built in Go with
[MyGo](https://github.com/egoist/mygo), the
[uix](https://github.com/smallnest/uix) components and
[onevra](https://github.com/oneo-me/onevra) markdown.

The terminal emulator is Ghostty's, loaded through MyGo's terminal plugin.
The app draws its own UI with MyGo's native toolkit. It uses no webview.

## Features

- **Tabs and panes.** Each tab is a session that keeps running when you
  switch away. Split a pane right or down, drag the dividers, and move the
  focus with Cmd+] and Cmd+[.
- **Sessions rail.** One compact tab per session — its directory and
  shell — with a hover close button.
- **Info panel.** The ⓘ button opens the right panel otty shows there: the
  working directory with actions (copy path, reveal in Finder, open in
  Zed), the running process, and the ports in use.
- **Command palette.** Cmd+K opens it. It runs window actions, jumps to
  sessions, and picks a terminal theme.
- **Ghostty themes.** The palette offers 28 curated palettes, or "Follow
  system" to match the window's light or dark appearance.
- **Agent chat.** A panel on the right renders the chat as markdown. The
  built-in agent is a demo; `agent.go` defines the interface to wire a
  real model.
- **Keyboard first.** Tabs, panes, palette, font size, rail and chat
  are all reachable from the keyboard. The menus show the shortcuts.

## Run

```sh
go run .
```

Go 1.27 downloads itself if needed. The first run fetches
`libghostty-vt` (Ghostty's emulator) into the MyGo cache.

## Package

```sh
go run github.com/egoist/mygo/cmd/mygo@v0.4.2 build
```

The result is a signed `Nutty.app` and a DMG under `dist/`.

## Tests

```sh
go test ./...
```

The tests run headless: they need no window. To render a preview image of
the window (a visual check):

```sh
NUTTY_SNAPSHOT=1 go test -run TestSnapshot
```

It writes `/tmp/nutty-preview.png` and `/tmp/nutty-preview-dark.png`.

## Layout

| File | What it does |
| --- | --- |
| `app.go` | Window state and actions |
| `session.go` | Sessions, the split tree, pane focus |
| `view.go` | The root view layout |
| `topbar.go` | The top bar: title and icon buttons |
| `rail.go` | The sessions rail |
| `info.go` | The info panel: directory, process, ports |
| `tabbar.go` | The rename dialog |
| `palette.go` | The commands of the palette |
| `themes.go` | The curated Ghostty themes |
| `menus.go` | The application menu |
| `chat.go` | The agent chat panel |
| `agent.go` | The `Agent` interface and the demo agent |
| `nut.go` | The app's nut icon (`assets/nutty.svg`) |

## Agent backend

`agent.go` declares `Agent`, the interface the chat panel streams from.
`demoAgent` answers with a small markdown page so the UI (streaming caret,
thinking indicator, markdown rendering) works end to end. To use a real
model, replace `demoAgent` with an implementation that calls the model and
streams its text into the channel.
