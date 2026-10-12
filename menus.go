package main

import (
	"github.com/egoist/mygo"
)

// setAppMenu builds the application menu, mirroring otty's layout. The
// handlers dispatch to the focused window's app, so each window keeps its
// own sessions. Edit roles go to the focused control, so a terminal gets
// its copy and paste.
func setAppMenu() {
	menu := mygo.NewMenu([]*mygo.MenuItem{
		{
			Label: "Nutty",
			Submenu: []*mygo.MenuItem{
				{Role: mygo.RoleAbout},
				mygo.Separator(),
				{Label: "Settings…", Accelerator: "CmdOrCtrl+,", Click: act(func(a *app) { a.openSettings() })},
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
				{Label: "New Window", Accelerator: "CmdOrCtrl+Shift+N", Click: func(_ *mygo.MenuItem, _ *mygo.Window) { newWindow() }},
				{Label: "New Tab", Accelerator: "CmdOrCtrl+T", Click: act(func(a *app) { a.newTab() })},
				{Label: "Duplicate Tab", Accelerator: "CmdOrCtrl+Shift+T", Click: act(func(a *app) { a.duplicateTab() })},
				mygo.Separator(),
				{Label: "Close Tab", Accelerator: "CmdOrCtrl+W", Click: act(func(a *app) { a.closeTab() })},
				{Role: mygo.RoleClose},
				mygo.Separator(),
				{Label: "Reveal in Finder", Accelerator: "CmdOrCtrl+Shift+R", Click: act(func(a *app) { a.revealInFinder() })},
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
				{Label: "Toggle Sidebar", Accelerator: "CmdOrCtrl+Shift+L", Click: act(func(a *app) { a.toggleSidebar() })},
				{Label: "Toggle Agent Chat", Accelerator: "CmdOrCtrl+Shift+A", Click: act(func(a *app) { a.toggleChat() })},
				{Label: "Command Palette", Accelerator: "CmdOrCtrl+Shift+P", Click: act(func(a *app) { a.paletteOpen = true })},
				mygo.Separator(),
				{Label: "Increase Font Size", Accelerator: "CmdOrCtrl+=", Click: act(func(a *app) { a.zoomFont(1) })},
				{Label: "Decrease Font Size", Accelerator: "CmdOrCtrl+-", Click: act(func(a *app) { a.zoomFont(-1) })},
				{Label: "Reset Font Size", Accelerator: "CmdOrCtrl+0", Click: act(func(a *app) { a.resetFont() })},
				mygo.Separator(),
				{Role: mygo.RoleToggleFullScreen},
			},
		},
		{
			Label: "Terminal",
			Submenu: []*mygo.MenuItem{
				{Label: "Rename Tab…", Accelerator: "CmdOrCtrl+R", Click: act(func(a *app) { a.beginRename() })},
				{Label: "Split Right", Accelerator: "CmdOrCtrl+D", Click: act(func(a *app) { a.splitRight() })},
				{Label: "Split Down", Accelerator: "CmdOrCtrl+Shift+D", Click: act(func(a *app) { a.splitDown() })},
				mygo.Separator(),
				{Label: "Clear Screen", Accelerator: "CmdOrCtrl+Shift+K", Click: act(func(a *app) { a.clearScreen() })},
				{Label: "Copy Path", Accelerator: "CmdOrCtrl+Shift+C", Click: act(func(a *app) { a.copyPath() })},
			},
		},
		{
			Label: "Window",
			Submenu: []*mygo.MenuItem{
				{Role: mygo.RoleMinimize},
				{Role: mygo.RoleZoom},
				mygo.Separator(),
				{Label: "Next Tab", Accelerator: "CmdOrCtrl+Shift+]", Click: act(func(a *app) { a.nextTab() })},
				{Label: "Previous Tab", Accelerator: "CmdOrCtrl+Shift+[", Click: act(func(a *app) { a.prevTab() })},
				{Label: "Focus Next Pane", Accelerator: "CmdOrCtrl+]", Click: act(func(a *app) { a.focusNextPane() })},
				{Label: "Focus Previous Pane", Accelerator: "CmdOrCtrl+[", Click: act(func(a *app) { a.focusPrevPane() })},
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
