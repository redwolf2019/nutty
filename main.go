// Nutty is a terminal app in the spirit of otty: a native window of
// shells, split into tabs and panes, with a command palette, Ghostty
// themes and an agent chat panel. Built on MyGo, the uix components and
// onevra's markdown.
package main

import (
	"log"

	"nutty/settings"

	"github.com/egoist/mygo"
)

func main() {
	mygo.App.WhenReady(func() {
		registerFonts()
		s := settings.Load()
		setUILang(s.Language)
		setAppMenu(s)
		newWindow()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
