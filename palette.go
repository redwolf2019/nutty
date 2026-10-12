package main

import (
	"github.com/oneo-me/onevra/source/icons/lucide"
	"github.com/smallnest/uix/components/command"
)

// paletteItems are the commands of the ⌘K palette: the window's actions,
// the sessions to jump to, and the terminal themes.
func (a *app) paletteItems() []command.Item {
	items := []command.Item{
		{Label: "New Tab", Keywords: []string{"session", "shell"}, Icon: lucide.Plus(), OnSelect: a.newTab},
		{Label: "New Window", Keywords: []string{"session"}, Icon: lucide.SquarePlus(), OnSelect: func() { newWindow() }},
		{Label: "Duplicate Tab", Keywords: []string{"clone"}, Icon: lucide.Copy(), OnSelect: a.duplicateTab},
		{Label: "Close Tab", Keywords: []string{"quit", "kill"}, Icon: lucide.X(), OnSelect: a.closeTab},
		{Label: "Split Right", Keywords: []string{"pane", "vertical"}, Icon: lucide.PanelRight(), OnSelect: a.splitRight},
		{Label: "Split Down", Keywords: []string{"pane", "horizontal"}, Icon: lucide.PanelRight(), OnSelect: a.splitDown},
		{Label: "Focus Next Pane", Keywords: []string{"navigate"}, Icon: lucide.ChevronRight(), OnSelect: a.focusNextPane},
		{Label: "Focus Previous Pane", Keywords: []string{"navigate"}, Icon: lucide.ChevronLeft(), OnSelect: a.focusPrevPane},
		{Label: "Next Tab", Keywords: []string{"cycle"}, Icon: lucide.ChevronsRight(), OnSelect: a.nextTab},
		{Label: "Previous Tab", Keywords: []string{"cycle"}, Icon: lucide.ChevronsLeft(), OnSelect: a.prevTab},
		{Label: "Toggle Sidebar", Keywords: []string{"rail", "sessions"}, Icon: lucide.PanelLeft(), OnSelect: a.toggleSidebar},
		{Label: "Toggle Agent Chat", Keywords: []string{"chat", "composer"}, Icon: lucide.MessageSquare(), OnSelect: a.toggleChat},
		{Label: "Rename Tab…", Keywords: []string{"title"}, Icon: lucide.Pen(), OnSelect: a.beginRename},
		{Label: "Clear Screen", Keywords: []string{"clear", "reset"}, Icon: lucide.Eraser(), OnSelect: a.clearScreen},
		{Label: "Copy Path", Keywords: []string{"directory", "cwd"}, Icon: lucide.Folder(), OnSelect: a.copyPath},
		{Label: "Reveal in Finder", Keywords: []string{"folder", "open", "cwd"}, Icon: lucide.FolderOpen(), OnSelect: a.revealInFinder},
		{Label: "Increase Font Size", Keywords: []string{"zoom", "bigger"}, Icon: lucide.ZoomIn(), OnSelect: func() { a.zoomFont(1) }},
		{Label: "Decrease Font Size", Keywords: []string{"zoom", "smaller"}, Icon: lucide.ZoomOut(), OnSelect: func() { a.zoomFont(-1) }},
		{Label: "Reset Font Size", Keywords: []string{"zoom", "default"}, Icon: lucide.RotateCcw(), OnSelect: a.resetFont},
		{Label: "Appearance: System", Keywords: []string{"light", "dark", "theme"}, Icon: lucide.Monitor(), OnSelect: a.appearanceSystem},
		{Label: "Appearance: Light", Keywords: []string{"theme"}, Icon: lucide.Sun(), OnSelect: a.appearanceLight},
		{Label: "Appearance: Dark", Keywords: []string{"theme"}, Icon: lucide.Moon(), OnSelect: a.appearanceDark},
	}
	for _, name := range themeNames() {
		name := name
		items = append(items, command.Item{
			Label:    "Theme: " + name,
			Keywords: []string{"terminal", "color", "palette"},
			Icon:     lucide.Palette(),
			OnSelect: func() { a.setTermTheme(name) },
		})
	}
	for _, s := range a.sessions {
		s := s
		items = append(items, command.Item{
			Label:    "Switch to " + s.displayTitle(),
			Keywords: []string{"session", "tab", s.dir},
			Icon:     lucide.Terminal(),
			OnSelect: func() { a.activate(s) },
		})
	}
	return items
}
