package settings

// validate puts back what a hand-edited or imported file can hold and no
// control can make: a number outside its range, as a window opacity of 0
// (an invisible window) or a negative width, goes to the nearest end of
// the range, and an enum's value that is not one of its options goes to
// its default. The registry drives both, so a new setting is covered by
// its def alone. Then the values that refer to other values: the details
// views, each once and of a kind this build knows, and the default
// profile, which names a profile or nothing. It runs after the booleans
// of older files have become their enums.
func (s *Settings) validate() {
	for i := range settingsDefs {
		d := &settingsDefs[i]
		switch d.kind {
		case "number":
			v := d.get(s).(float64)
			if in := min(max(v, d.lo), d.hi); in != v {
				d.set(s, in)
			}
		case "enum":
			if !d.open && !d.offers(d.get(s).(string)) {
				d.set(s, d.def)
			}
		}
	}

	// A view is found by its id, so a second view of an id is never drawn
	// and its switch moves the first. Beside the built-in views a file
	// may hold otty's own kinds of view, which keep their place; any
	// other id belongs to nothing. A built-in view the file lacks comes
	// back, shown, since the window draws it and the page must offer its
	// switch.
	builtin := Default().DetailsViews
	seen := map[string]bool{}
	views := []DetailsView{}
	for _, v := range s.DetailsViews {
		known := v.Kind == "tui" || v.Kind == "web"
		for _, b := range builtin {
			known = known || v.ID == b.ID
		}
		if v.ID == "" || seen[v.ID] || !known {
			continue
		}
		seen[v.ID] = true
		views = append(views, v)
	}
	for _, b := range builtin {
		if !seen[b.ID] {
			views = append(views, b)
		}
	}
	s.DetailsViews = views

	if s.Profile() == nil {
		s.DefaultProfile = ""
	}
}

// offers reports whether v is one of an enum's options.
func (d *settingDef) offers(v string) bool {
	for _, o := range d.enum {
		if o.value == v {
			return true
		}
	}
	return false
}
