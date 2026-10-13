package settings

import (
	"reflect"
	"testing"
)

// TestRegistryDefaultsAgree checks that the default a def carries is the
// one Default sets: they are two lists of the same otty defaults, and
// nothing kept them from drifting apart.
func TestRegistryDefaultsAgree(t *testing.T) {
	s := Default()
	for _, d := range settingsDefs {
		if d.def == nil {
			continue // a list, or a palette color: empty by default
		}
		if got := d.get(s); !reflect.DeepEqual(got, d.def) {
			t.Errorf("%s: Default gives %v, its def says %v", d.key, got, d.def)
		}
	}
}

// TestNormalizeOptionAsAlt checks the Option key's setting through
// Normalize: an empty value is off, and the four values stay.
func TestNormalizeOptionAsAlt(t *testing.T) {
	for in, want := range map[string]string{"": "false", "false": "false", "true": "true", "left": "left", "right": "right"} {
		s := Default()
		s.OptionAsAlt = in
		s.Normalize()
		if s.OptionAsAlt != want {
			t.Errorf("OptionAsAlt %q normalizes to %q, want %q", in, s.OptionAsAlt, want)
		}
	}
}
