package i18n

import "testing"

// TestLangStates checks that SetLang resolves the Language settings a
// headless test can: "chinese" and "english" map to themselves. The
// "system" value follows the OS locale through mygo.App.Locale, which
// dispatches to a main goroutine that never runs under a test, so it is
// left to the running app.
func TestLangStates(t *testing.T) {
	SetLang("chinese")
	if Tr("Settings") != "设置" {
		t.Fatalf("chinese: Tr(Settings) = %q, want 设置", Tr("Settings"))
	}
	SetLang("english")
	if Tr("Settings") != "Settings" {
		t.Fatalf("english: Tr(Settings) = %q, want Settings", Tr("Settings"))
	}
	// Whatever the previous state, the next render sees it.
	SetLang("chinese")
	if Tr("New Tab") != "新建标签页" {
		t.Fatalf("chinese: Tr(New Tab) = %q, want 新建标签页", Tr("New Tab"))
	}
}

// TestTrRoundTrip checks that translated strings come back untranslated
// in English, so a value persisted through Tr still round-trips.
func TestTrRoundTrip(t *testing.T) {
	SetLang("chinese")
	for en, zh := range map[string]string{
		"New Tab":       "新建标签页",
		"Follow system": "跟随系统",
		"Close Tab?":    "关闭标签页？",
	} {
		if got := Tr(en); got != zh {
			t.Errorf("Tr(%q) = %q, want %q", en, got, zh)
		}
	}
	SetLang("english")
	for en := range map[string]string{
		"New Tab":       "新建标签页",
		"Follow system": "跟随系统",
		"Close Tab?":    "关闭标签页？",
	} {
		if got := Tr(en); got != en {
			t.Errorf("english Tr(%q) = %q, want %q", en, got, en)
		}
	}
}
