package app

import (
	"nutty/internal/settings"

	"github.com/oneo-me/onevra/source/icons/lucide"
	"github.com/smallnest/uix/components/command"
)

// paletteItems are the commands of the ⌘K palette: the window's actions,
// the sessions to jump to, and the terminal themes.
func (a *app) paletteItems() []command.Item {
	items := []command.Item{
		{Label: tr("New Tab"), Keywords: []string{"session", "shell"}, Icon: lucide.Plus(), OnSelect: a.newTab},
		{Label: tr("New Window"), Keywords: []string{"session"}, Icon: lucide.SquarePlus(), OnSelect: func() { newWindow() }},
		{Label: tr("Duplicate Tab"), Keywords: []string{"clone"}, Icon: lucide.Copy(), OnSelect: a.duplicateTab},
		{Label: tr("Close Tab"), Keywords: []string{"quit", "kill"}, Icon: lucide.X(), OnSelect: a.closeTab},
		{Label: tr("Split Right"), Keywords: []string{"pane", "vertical"}, Icon: lucide.PanelRight(), OnSelect: a.splitRight},
		{Label: tr("Split Down"), Keywords: []string{"pane", "horizontal"}, Icon: lucide.PanelRight(), OnSelect: a.splitDown},
		{Label: tr("Focus Next Pane"), Keywords: []string{"navigate"}, Icon: lucide.ChevronRight(), OnSelect: a.focusNextPane},
		{Label: tr("Focus Previous Pane"), Keywords: []string{"navigate"}, Icon: lucide.ChevronLeft(), OnSelect: a.focusPrevPane},
		{Label: tr("Next Tab"), Keywords: []string{"cycle"}, Icon: lucide.ChevronsRight(), OnSelect: a.nextTab},
		{Label: tr("Previous Tab"), Keywords: []string{"cycle"}, Icon: lucide.ChevronsLeft(), OnSelect: a.prevTab},
		{Label: tr("Toggle Sidebar"), Keywords: []string{"rail", "sessions"}, Icon: lucide.PanelLeft(), OnSelect: a.toggleSidebar},
		{Label: tr("Toggle Agent Chat"), Keywords: []string{"chat", "composer"}, Icon: lucide.MessageSquare(), OnSelect: a.toggleChat},
		{Label: tr("Rename Tab…"), Keywords: []string{"title"}, Icon: lucide.Pen(), OnSelect: a.beginRename},
		{Label: tr("Clear Screen"), Keywords: []string{"clear", "reset"}, Icon: lucide.Eraser(), OnSelect: a.clearScreen},
		{Label: tr("Copy Path"), Keywords: []string{"directory", "cwd"}, Icon: lucide.Folder(), OnSelect: a.copyPath},
		{Label: tr("Reveal in Finder"), Keywords: []string{"folder", "open", "cwd"}, Icon: lucide.FolderOpen(), OnSelect: a.revealInFinder},
		{Label: tr("Increase Font Size"), Keywords: []string{"zoom", "bigger"}, Icon: lucide.ZoomIn(), OnSelect: func() { a.zoomFont(1) }},
		{Label: tr("Decrease Font Size"), Keywords: []string{"zoom", "smaller"}, Icon: lucide.ZoomOut(), OnSelect: func() { a.zoomFont(-1) }},
		{Label: tr("Reset Font Size"), Keywords: []string{"zoom", "default"}, Icon: lucide.RotateCcw(), OnSelect: a.resetFont},
		{Label: tr("Appearance: System"), Keywords: []string{"light", "dark", "theme"}, Icon: lucide.Monitor(), OnSelect: a.appearanceSystem},
		{Label: tr("Appearance: Light"), Keywords: []string{"theme"}, Icon: lucide.Sun(), OnSelect: a.appearanceLight},
		{Label: tr("Appearance: Dark"), Keywords: []string{"theme"}, Icon: lucide.Moon(), OnSelect: a.appearanceDark},
	}
	for _, name := range settings.ThemeNames() {
		name := name
		items = append(items, command.Item{
			Label:    tr("Theme: ") + name,
			Keywords: []string{"terminal", "color", "palette"},
			Icon:     lucide.Palette(),
			OnSelect: func() { a.setTermTheme(name) },
		})
	}
	for _, s := range a.sessions {
		s := s
		items = append(items, command.Item{
			Label:    tr("Switch to ") + tabLabel(s),
			Keywords: []string{"session", "tab", s.dir},
			Icon:     lucide.Terminal(),
			OnSelect: func() { a.activate(s) },
		})
	}
	return items
}
