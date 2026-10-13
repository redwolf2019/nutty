package app

import (
	"strconv"
	"testing"

	"nutty/internal/settings"

	"github.com/egoist/mygo"
)

// menuShortcuts walks the application menu and returns the accelerator of
// every item, by its label, and of every role item, by its role.
func menuShortcuts(items []*mygo.MenuItem, byLabel map[string]string, byRole map[mygo.MenuRole]string) {
	for _, it := range items {
		if it.Role != "" {
			byRole[it.Role] = it.Accelerator
		} else if it.Label != "" {
			byLabel[it.Label] = it.Accelerator
		}
		menuShortcuts(it.Submenu, byLabel, byRole)
	}
}

// TestMenuWiresBoundActions checks that the actions the default key
// bindings bind and the app can do are in the menu, with their shortcuts:
// Cmd+1 to Cmd+9 for the tabs, Cmd+Shift+R for the details panel and
// Close Pane were bound, or listed, with no menu item to run them.
func TestMenuWiresBoundActions(t *testing.T) {
	s := settings.Default()
	byLabel, byRole := map[string]string{}, map[mygo.MenuRole]string{}
	menuShortcuts(menuItems(s), byLabel, byRole)

	for n := 1; n <= 9; n++ {
		label := tr("Tab " + strconv.Itoa(n))
		if got, want := byLabel[label], "CmdOrCtrl+"+strconv.Itoa(n); got != want {
			t.Errorf("%s: accelerator = %q, want %q", label, got, want)
		}
	}
	if got := byLabel[tr("Toggle Details Panel")]; got != "CmdOrCtrl+Shift+R" {
		t.Errorf("Toggle Details Panel: accelerator = %q, want CmdOrCtrl+Shift+R", got)
	}
	if _, ok := byLabel[tr("Close Pane")]; !ok {
		t.Error("the menu has no Close Pane")
	}

	// The roles follow the key bindings as well, so an override recorded on
	// the Key Bindings page moves the shortcut.
	s.KeybindOverrides = map[string]string{"cmd+shift+m": "minimize", "cmd+opt+f": "toggle_full_screen", "cmd+shift+9": "goto_tab:2"}
	byLabel, byRole = map[string]string{}, map[mygo.MenuRole]string{}
	menuShortcuts(menuItems(s), byLabel, byRole)
	if got := byRole[mygo.RoleMinimize]; got != "CmdOrCtrl+Shift+M" {
		t.Errorf("Minimize: accelerator = %q, want the override CmdOrCtrl+Shift+M", got)
	}
	if got := byRole[mygo.RoleToggleFullScreen]; got != "CmdOrCtrl+Option+F" {
		t.Errorf("Toggle Full Screen: accelerator = %q, want the override CmdOrCtrl+Option+F", got)
	}
	if got := byLabel[tr("Tab 2")]; got != "CmdOrCtrl+Shift+9" {
		t.Errorf("Tab 2: accelerator = %q, want the override CmdOrCtrl+Shift+9", got)
	}
}

// TestGotoTab checks that a tab is chosen by its place, counting from 1,
// and that a place with no tab changes nothing.
func TestGotoTab(t *testing.T) {
	a := testApp()
	a.settings.NewTabPosition = "end"
	first := a.newSession("/tmp/nutty-goto-1")
	second := a.newSession("/tmp/nutty-goto-2")
	a.gotoTab(1)
	if a.currentSession() != first {
		t.Fatal("Tab 1 should be the first tab")
	}
	a.gotoTab(2)
	if a.currentSession() != second {
		t.Fatal("Tab 2 should be the second tab")
	}
	for _, n := range []int{0, 3, 9, -1} {
		a.gotoTab(n)
		if a.currentSession() != second {
			t.Fatalf("Tab %d is not there: the active tab should stay", n)
		}
	}
	a.closeSession(first)
	a.closeSession(second)
	a.gotoTab(1) // no tabs: nothing to do, and no panic
}
