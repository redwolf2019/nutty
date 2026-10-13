package app

// detailsTabs is the sections the info panel offers: the built-in views in
// the order of the View page's Details Panel list, without the ones it
// switches off. A section the list does not name shows, after the ones it
// does. The panel has to show something, so with every section switched
// off Info stays; there is always a first one.
func (a *app) detailsTabs() []infoTab {
	all := []infoTab{infoTabInfo, infoTabOutline, infoTabGit, infoTabFiles}
	tabs := []infoTab{}
	listed := map[infoTab]bool{}
	if a.settings != nil {
		for _, v := range a.settings.DetailsViews {
			for _, tab := range all {
				if infoTabID(tab) == v.ID && !listed[tab] {
					listed[tab] = true
					if v.Enabled {
						tabs = append(tabs, tab)
					}
				}
			}
		}
	}
	for _, tab := range all {
		if !listed[tab] {
			tabs = append(tabs, tab)
		}
	}
	if len(tabs) == 0 {
		tabs = append(tabs, infoTabInfo)
	}
	return tabs
}
