package settings

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	uixsettings "github.com/smallnest/uix/components/settings"

	"nutty/i18n"
)

// App is the host the settings window applies and saves through. The main
// package's app implements it; the tests use a mock.
type App interface {
	ApplySettings()
	SaveSettings()
	ApplyKeybinds()
}

// NewWindow opens the Settings window editing s, applying and saving every
// change through app. The window's closing is the caller's to observe, so
// it can forget the window it keeps open (the app's single-window rule).
func NewWindow(app App, s *Settings) *mygo.Window {
	sw := &settingsWindow{
		app:   app,
		s:     s,
		last:  *s,
		state: uixsettings.SettingsState{Category: "appearance"},
	}
	sw.win = mygo.NewWindow(mygo.WindowOptions{
		Title:     i18n.Tr("Settings"),
		Width:     760,
		Height:    540,
		MinWidth:  600,
		MinHeight: 420,
		Content:   ui.View(sw.view),
	})
	return sw.win
}
