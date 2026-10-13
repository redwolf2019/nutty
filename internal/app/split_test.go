package app

import (
	"testing"

	"nutty/internal/settings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// splitTester draws an app whose panes hold real terminals on a
// connection that never ends, so each pane lays out as it does in the
// window, and records where every pane was drawn.
func splitTester(t *testing.T, w, h int) (*app, *session, *ui.Tester, map[*pane]ui.Rect) {
	t.Helper()
	registerFonts()
	old := newTerminal
	t.Cleanup(func() { newTerminal = old })
	newTerminal = func(o terminal.Options) (*terminal.Terminal, error) {
		o.Conn = &scriptConn{data: []byte("$ "), hold: true}
		return terminal.New(o)
	}
	a := &app{settings: settings.Default(), sideOpen: true, agent: &demoAgent{}}
	s := a.newSession("/tmp/nutty-split")
	drawn := map[*pane]ui.Rect{}
	tt := ui.NewTester(func(c *ui.Context) {
		a.view(c)
		for _, p := range s.panes {
			if el := c.Resolve(p.handle); el.Valid() {
				drawn[p] = el.Bounds()
			}
		}
	}, w, h)
	return a, s, tt, drawn
}

// settleFrames draws enough frames for a change of layout to reach the
// bounds the view reads back, which are the previous frame's.
func settleFrames(tt *ui.Tester) {
	for range 4 {
		tt.Frame()
	}
}

// TestSplitHalvesThePane splits a pane that is already one side of a split
// and checks that the new pane gets room: the divider used to sit at half
// of the window, which is all of the pane being split, so the new pane was
// zero wide.
func TestSplitHalvesThePane(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		a, s, tt, drawn := splitTester(t, 1120, 740)
		settleFrames(tt)
		first := s.panes[0]
		a.splitPane(first, vertical)
		settleFrames(tt)
		a.splitPane(first, vertical)
		settleFrames(tt)
		if len(s.panes) != 3 {
			t.Fatalf("vertical=%v: panes = %d, want 3", vertical, len(s.panes))
		}
		for i, p := range s.orderedPanes() {
			r := drawn[p]
			if r.W < leastPane || r.H < leastPane {
				t.Errorf("vertical=%v: pane %d has no room: %+v", vertical, i, r)
			}
			if r.X+r.W > 1120 || r.Y+r.H > 740 {
				t.Errorf("vertical=%v: pane %d is drawn outside the window: %+v", vertical, i, r)
			}
		}
	}
}

// TestSplitKeepsItsRoom narrows the window under a split whose divider is
// further right than the narrower window is wide, and checks that both
// panes stay inside it.
func TestSplitKeepsItsRoom(t *testing.T) {
	a, s, tt, drawn := splitTester(t, 1120, 740)
	settleFrames(tt)
	a.splitPane(s.panes[0], false)
	*s.root.split.size = 600 // where a drag left the divider
	settleFrames(tt)
	tt.SetSize(680, 740) // the window's least width
	settleFrames(tt)
	for i, p := range s.orderedPanes() {
		r := drawn[p]
		if r.W < leastPane {
			t.Errorf("pane %d has no room in the narrow window: %+v", i, r)
		}
		if r.X+r.W > 680 {
			t.Errorf("pane %d is drawn outside the narrow window: %+v", i, r)
		}
	}
}

// leastPane is the least width or height a pane keeps: the 40
// DIPs the Split widget's divider leaves either side.
const leastPane = 40
