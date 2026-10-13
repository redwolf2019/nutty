package app

import (
	"slices"
	"testing"

	"nutty/internal/settings"

	"github.com/egoist/mygo/ui"
)

// detailsViews is a Details Panel list of the built-in views in the order
// of ids, each switched on unless its id comes with a "-".
func detailsViews(ids ...string) []settings.DetailsView {
	views := []settings.DetailsView{}
	for _, id := range ids {
		on := id[0] != '-'
		if !on {
			id = id[1:]
		}
		views = append(views, settings.DetailsView{ID: id, Kind: "builtin", Enabled: on})
	}
	return views
}

// TestDetailsTabsFollowTheSettings checks that the info panel offers its
// sections as the View page's Details Panel list has them: in its order,
// without the ones switched off, Info among them, and Info alone when
// every one is off.
func TestDetailsTabsFollowTheSettings(t *testing.T) {
	a := testApp()
	for _, c := range []struct {
		views []settings.DetailsView
		want  []infoTab
	}{
		{detailsViews("info", "outline", "git", "files"), []infoTab{infoTabInfo, infoTabOutline, infoTabGit, infoTabFiles}},
		{detailsViews("files", "-git", "info", "outline"), []infoTab{infoTabFiles, infoTabInfo, infoTabOutline}},
		{detailsViews("-info", "git", "outline", "files"), []infoTab{infoTabGit, infoTabOutline, infoTabFiles}},
		{detailsViews("-info", "-outline", "-git", "-files"), []infoTab{infoTabInfo}},
		// A view the list does not name shows, after the ones it names; a
		// view of no built-in id, and a second one of an id, are no tab.
		{detailsViews("git", "docs", "-git"), []infoTab{infoTabGit, infoTabInfo, infoTabOutline, infoTabFiles}},
	} {
		a.settings.DetailsViews = c.views
		if got := a.detailsTabs(); !slices.Equal(got, c.want) {
			t.Errorf("details views %+v: tabs %v, want %v", c.views, got, c.want)
		}
	}
}

// TestDetailsPanelDrawsTheOrder checks the panel itself: its toolbar draws
// the sections in the settings' order, and opening the panel with Info
// switched off opens the first section that is on.
func TestDetailsPanelDrawsTheOrder(t *testing.T) {
	// pillX is where the open section's labeled pill sits in the toolbar.
	pillX := func(a *app, tab infoTab, label string) float32 {
		t.Helper()
		a.panel, a.info.tab = panelInfo, tab
		tt := ui.NewTester(a.view, 1280, 780)
		tt.Frame()
		r, ok := tt.Find(label)
		if !ok {
			t.Fatalf("the toolbar shows no %q pill", label)
		}
		return r.X
	}
	a := testApp()
	a.newSession(t.TempDir())
	if info, files := pillX(a, infoTabInfo, "Info"), pillX(a, infoTabFiles, "Files"); info >= files {
		t.Fatalf("by default Info (at %v) should sit left of Files (at %v)", info, files)
	}
	a.settings.DetailsViews = detailsViews("files", "git", "outline", "info")
	if info, files := pillX(a, infoTabInfo, "Info"), pillX(a, infoTabFiles, "Files"); files >= info {
		t.Errorf("with Files first in the settings, Files (at %v) should sit left of Info (at %v)", files, info)
	}

	a.settings.DetailsViews = detailsViews("-info", "outline", "git", "files")
	a.info.tab = infoTabInfo
	a.setPanel(panelInfo)
	if a.info.tab != infoTabOutline {
		t.Errorf("with Info switched off the panel opens on tab %v, want Outline", a.info.tab)
	}
	tt := ui.NewTester(a.view, 1280, 780)
	tt.Frame()
	if !tt.HasText("Outline") {
		t.Error("the open Outline section has no pill in the toolbar")
	}
}
