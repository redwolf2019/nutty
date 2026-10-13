// Package app is Nutty: a native window of shells, split into tabs and
// panes, with a command palette, Ghostty themes and an agent chat panel.
// Built on MyGo, the uix components and onevra's markdown.
package app

import (
	"github.com/egoist/mygo"
)

// Run starts the application: it loads the settings, applies the UI
// language and menu, and opens the window.
func Run() error {
	mygo.App.WhenReady(func() {
		registerFonts()
		s := sharedSettings()
		setUILang(s.Language)
		setAppMenu(s)
		newWindow()
	})
	return mygo.App.Run()
}
