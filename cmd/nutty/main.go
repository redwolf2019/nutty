// Nutty is a terminal app in the spirit of otty: a native window of
// shells, split into tabs and panes, with a command palette, Ghostty
// themes and an agent chat panel. Built on MyGo, the uix components and
// onevra's markdown.
package main

import (
	"log"

	"nutty/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
