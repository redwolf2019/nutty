// Package i18n translates the app's display strings. The host calls
// SetLang with the Language setting ("system" | "english" | "chinese")
// and renders through Tr, which returns the string unchanged when no
// translation exists. Persisted values never go through Tr.
package i18n

import (
	"sort"
	"strings"
	"sync/atomic"

	"github.com/egoist/mygo"
)

// lang is the resolved interface language: "chinese" or "english".
// It starts empty, so before SetLang runs Tr returns English.
var lang atomic.Value

// SetLang resolves the Language setting to a concrete language.
// "system" follows the OS locale; mygo.App.Locale() panics before the
// app is ready, so a crash during startup resolves to Chinese rather
// than hiding the home language.
func SetLang(v string) {
	switch v {
	case "english":
		lang.Store("english")
	case "chinese":
		lang.Store("chinese")
	default: // "system"
		lang.Store(locale())
	}
}

func locale() string {
	defer func() { _ = recover() }()
	if strings.HasPrefix(mygo.App.Locale(), "zh") {
		return "chinese"
	}
	return "english"
}

// recording collects every string Tr saw, so a sweep test can render
// the whole app and list what still lacks a translation. Only tests
// enable it, before they start rendering.
var recording bool
var seen = map[string]int{}

// SetRecording turns Tr's recording on or off.
func SetRecording(on bool) {
	recording = on
	if !on {
		seen = map[string]int{}
	}
}

// Seen is the strings Tr saw while recording.
func Seen() map[string]int {
	out := make(map[string]int, len(seen))
	for s, n := range seen {
		out[s] = n
	}
	return out
}

// Untranslated is the strings Tr saw while recording that still lack a
// translation, sorted. The sweep tests drive the table to empty with it.
func Untranslated() []string {
	out := []string{}
	for s := range seen {
		if _, ok := table[s]; !ok {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// Tr returns s in the interface language, unchanged when a translation
// is missing. Only display strings call it.
func Tr(s string) string {
	if recording {
		seen[s]++
	}
	if lang.Load() != "chinese" || s == "" {
		return s
	}
	if zh, ok := table[s]; ok {
		return zh
	}
	return s
}

// table is the hand-written Chinese translation table in zh.go, keyed by
// the English display string. The sweep tests drive it to empty.
