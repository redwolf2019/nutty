package app

import (
	"strings"
	"testing"
	"unicode"

	"nutty/internal/i18n"

	"github.com/egoist/mygo/ui"
	"github.com/smallnest/uix/components/tree"
)

// langApp is a test app in the given interface language, with English put
// back afterwards and the home directory pointed away from the real
// settings file.
func langApp(t *testing.T, lang string) *app {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	i18n.SetLang(lang)
	t.Cleanup(func() { i18n.SetLang("english") })
	registerFonts()
	return testApp()
}

// TestFilesHeadingTranslated checks that the Files tab of the info panel
// draws its heading in the interface language, as the other tabs do: the
// heading was a bare literal, so a Chinese window showed "Files" over the
// tree.
func TestFilesHeadingTranslated(t *testing.T) {
	a := langApp(t, "chinese")
	a.panel = panelInfo
	a.info = infoState{tab: infoTabFiles, filesDone: true, files: []tree.Item{{Label: "a.go"}}}
	tt := ui.NewTester(a.view, 1280, 780)
	tt.Frame()
	if !tt.HasText("a.go") {
		t.Fatalf("the Files tab did not draw, texts: %v", tt.Texts())
	}
	if tt.HasText("Files") {
		t.Errorf("the Files heading is still English, texts: %v", tt.Texts())
	}
}

// TestPaletteThemeTranslated checks that the palette's "Follow system"
// theme reads in the interface language, while the theme names, which
// are proper nouns, stay as they are.
func TestPaletteThemeTranslated(t *testing.T) {
	a := langApp(t, "chinese")
	labels := map[string]bool{}
	for _, it := range a.paletteItems() {
		labels[it.Label] = true
	}
	for _, want := range []string{tr("Theme: ") + "跟随系统", tr("Theme: ") + "Tokyo Night"} {
		if !labels[want] {
			t.Errorf("palette missing %q", want)
		}
	}
	if labels[tr("Theme: ")+"Follow system"] {
		t.Error("palette still offers Follow system in English")
	}
}

// TestChatEmptyStateLanguage checks that the chat panel's empty state
// follows the interface language: it was written in Chinese in the
// source, so an English window showed Chinese.
func TestChatEmptyStateLanguage(t *testing.T) {
	a := langApp(t, "english")
	a.panel = panelChat
	tt := ui.NewTester(a.view, 1280, 780)
	tt.Frame()
	if !tt.HasText("Agent Panel") {
		t.Errorf("the empty state lacks its heading, texts: %v", tt.Texts())
	}
	for _, s := range tt.Texts() {
		if strings.ContainsFunc(s, func(r rune) bool { return unicode.Is(unicode.Han, r) }) {
			t.Errorf("the English window shows Chinese: %q", s)
		}
	}

	i18n.SetLang("chinese")
	tt.Frame()
	if !tt.HasText("Agent 面板") {
		t.Errorf("the Chinese empty state lacks its heading, texts: %v", tt.Texts())
	}
}
