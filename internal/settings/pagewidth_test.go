package settings

import (
	"testing"

	"github.com/egoist/mygo/ui"
	uixsettings "github.com/smallnest/uix/components/settings"
)

// TestPagesFitTheWindow checks that a page that draws itself fits the
// settings window at the width it opens with and at its smallest: the
// trigger chip at the right end of a key binding's row, the only way to
// rebind it, is inside the window, and the page still takes the row of a
// wider window.
func TestPagesFitTheWindow(t *testing.T) {
	const pad = 20 // the shell's padding around a page
	// NewWindow's MinWidth and Width, and a wider window.
	for _, width := range []int{600, 760, 1000} {
		s := Default()
		s.Language = "english"
		sw := &settingsWindow{app: mockApp{}, s: s, last: *s,
			state: uixsettings.SettingsState{Category: "key_bindings"}}
		sw.keys.search = "new tab" // one row, wherever the list puts it
		tt := ui.NewTester(sw.view, width, 540)
		tt.Frame()
		tt.Frame()
		r, ok := tt.Find("cmd+t")
		if !ok {
			t.Fatalf("width %d: no chip for New Tab", width)
		}
		right := r.X + r.W
		if r.W == 0 || right > float32(width-pad) {
			t.Errorf("width %d: the chip ends at %.0f (%.0f wide), past the page's edge at %d", width, right, r.W, width-pad)
		}
		if right < float32(width-pad)-40 {
			t.Errorf("width %d: the chip ends at %.0f, the page does not reach its edge at %d", width, right, width-pad)
		}
	}
}
