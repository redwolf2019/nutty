package app

import (
	"strings"

	"github.com/egoist/mygo/plugins/terminal"
)

// runAfter types the default profile's Run After into a pane that just
// opened, with the Enter that runs it: "A command to run when the pane
// opens", as the Profiles page says. The terminal keeps what is sent
// until the shell is there to read it, so the command runs at the shell's
// first prompt.
func (a *app) runAfter(term *terminal.Terminal) {
	if pr := a.settings.Profile(); pr != nil && strings.TrimSpace(pr.RunAfter) != "" {
		term.Send([]byte(pr.RunAfter + "\r"))
	}
}
