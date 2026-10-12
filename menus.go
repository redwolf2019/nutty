package main

import (
	"github.com/egoist/mygo"
)

// setAppMenu builds the application menu, mirroring otty's layout. The
// accelerators come from the key bindings: a recorded override, else
// otty's default, so a change on the Key Bindings page rebuilds the menu
// (see applyKeybinds). The handlers dispatch to the focused window's app,
// so each window keeps its own sessions. Edit roles go to the focused
// control, so a terminal gets its copy and paste.
func setAppMenu(s *settings) {
	acc := func(action string) string {
		if t, ok := s.actionTrigger(action); ok {
			return acceleratorOf(t)
		}
		return ""
	}
	menu := mygo.NewMenu([]*mygo.MenuItem{
		{
			Label: "Nutty",
			Submenu: []*mygo.MenuItem{
				{Role: mygo.RoleAbout},
				mygo.Separator(),
				{Label: "Settings…", Accelerator: acc("open_settings"), Click: act(func(a *app) { a.openSettings() })},
				mygo.Separator(),
				{Role: mygo.RoleServices},
				mygo.Separator(),
				{Role: mygo.RoleHide},
				{Role: mygo.RoleHideOthers},
				{Role: mygo.RoleUnhide},
				mygo.Separator(),
				{Role: mygo.RoleQuit},
			},
		},
		{
			Label: "File",
			Submenu: []*mygo.MenuItem{
				{Label: "New Window", Accelerator: acc("new_window"), Click: func(_ *mygo.MenuItem, _ *mygo.Window) { newWindow() }},
				{Label: "New Tab", Accelerator: acc("new_tab"), Click: act(func(a *app) { a.newTab() })},
				{Label: "Duplicate Tab", Accelerator: acc("duplicate_tab"), Click: act(func(a *app) { a.duplicateTab() })},
				mygo.Separator(),
				{Label: "Close Tab", Accelerator: acc("close_tab"), Click: act(func(a *app) { a.closeTab() })},
				{Role: mygo.RoleClose},
				mygo.Separator(),
				{Label: "Reveal in Finder", Click: act(func(a *app) { a.revealInFinder() })},
			},
		},
		{
			Label: "Edit",
			Submenu: []*mygo.MenuItem{
				{Role: mygo.RoleUndo},
				{Role: mygo.RoleRedo},
				mygo.Separator(),
				{Role: mygo.RoleCut},
				{Role: mygo.RoleCopy},
				{Role: mygo.RolePaste},
				{Role: mygo.RoleSelectAll},
			},
		},
		{
			Label: "View",
			Submenu: []*mygo.MenuItem{
				{Label: "Toggle Sidebar", Accelerator: acc("toggle_tabs_panel"), Click: act(func(a *app) { a.toggleSidebar() })},
				{Label: "Toggle Agent Chat", Accelerator: acc("open_agent"), Click: act(func(a *app) { a.toggleChat() })},
				{Label: "Command Palette", Accelerator: acc("command_palette"), Click: act(func(a *app) { a.paletteOpen = true })},
				mygo.Separator(),
				{Label: "Increase Font Size", Accelerator: acc("increase_font_size"), Click: act(func(a *app) { a.zoomFont(1) })},
				{Label: "Decrease Font Size", Accelerator: acc("decrease_font_size"), Click: act(func(a *app) { a.zoomFont(-1) })},
				{Label: "Reset Font Size", Accelerator: acc("reset_font_size"), Click: act(func(a *app) { a.resetFont() })},
				mygo.Separator(),
				{Role: mygo.RoleToggleFullScreen},
			},
		},
		{
			Label: "Terminal",
			Submenu: []*mygo.MenuItem{
				{Label: "Rename Tab…", Accelerator: acc("rename_tab"), Click: act(func(a *app) { a.beginRename() })},
				{Label: "Split Right", Accelerator: acc("split_right"), Click: act(func(a *app) { a.splitRight() })},
				{Label: "Split Down", Accelerator: acc("split_down"), Click: act(func(a *app) { a.splitDown() })},
				mygo.Separator(),
				{Label: "Clear Screen", Accelerator: acc("clear_screen"), Click: act(func(a *app) { a.clearScreen() })},
				{Label: "Copy Path", Accelerator: acc("copy_path"), Click: act(func(a *app) { a.copyPath() })},
			},
		},
		{
			Label: "Window",
			Submenu: []*mygo.MenuItem{
				{Role: mygo.RoleMinimize},
				{Role: mygo.RoleZoom},
				mygo.Separator(),
				{Label: "Next Tab", Accelerator: acc("next_tab"), Click: act(func(a *app) { a.nextTab() })},
				{Label: "Previous Tab", Accelerator: acc("previous_tab"), Click: act(func(a *app) { a.prevTab() })},
				{Label: "Focus Next Pane", Accelerator: acc("focus_pane_next"), Click: act(func(a *app) { a.focusNextPane() })},
				{Label: "Focus Previous Pane", Accelerator: acc("focus_pane_previous"), Click: act(func(a *app) { a.focusPrevPane() })},
			},
		},
		{Role: mygo.RoleHelp},
	})
	mygo.App.SetMenu(menu)
}

// act wraps an app action with the focused window's app.
func act(fn func(a *app)) func(*mygo.MenuItem, *mygo.Window) {
	return func(_ *mygo.MenuItem, win *mygo.Window) {
		if a := appOf(win); a != nil {
			fn(a)
		}
	}
}
