package settings

import (
	"encoding/json"
	"maps"
	"reflect"
	"slices"
	"strings"
)

// overlay reads a file's keys into the settings, one field at a time. A
// value that does not fit its field, as "18" where a number goes, is left
// out and the field keeps the value it has: read as one whole, a single
// such value failed the file, and every setting in it was dropped for the
// defaults. A field the file names is replaced whole, not merged into; a
// null, and a key no field has, change nothing. Keys match fields as
// encoding/json matches them, by name or else by name ignoring case, and
// are taken in order, so of two spellings of one key the same one wins
// each time.
func (s *Settings) overlay(file map[string]any) {
	v := reflect.ValueOf(s).Elem()
	for _, key := range slices.Sorted(maps.Keys(file)) {
		if file[key] == nil {
			continue
		}
		f, ok := v.Type().FieldByNameFunc(func(name string) bool { return strings.EqualFold(name, key) })
		if !ok {
			continue
		}
		raw, err := json.Marshal(file[key])
		if err != nil {
			continue
		}
		// A value of the field's own type, so a value that fits only in
		// part, as a list with one bad item, leaves nothing behind.
		fresh := reflect.New(f.Type)
		if json.Unmarshal(raw, fresh.Interface()) == nil {
			v.FieldByIndex(f.Index).Set(fresh.Elem())
		}
	}
}
