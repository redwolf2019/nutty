package app

import (
	"os"
	"testing"

	"nutty/internal/settings"
)

// TestZoomKeepsTheSettingsRange checks that Increase and Decrease Font
// Size move one step within the sizes the Settings page allows, 6 to 40:
// the zoom used to clamp to 8..26, so at a size of 32 Increase Font Size
// made the text smaller.
func TestZoomKeepsTheSettingsRange(t *testing.T) {
	// The zoom saves the settings: keep the file out of the real home.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	a := &app{settings: settings.Default()}

	for _, step := range []struct {
		from  float64
		delta float32
		want  float64
	}{
		{32, 1, 33},
		{32, -1, 31},
		{40, 1, 40},
		{7, -1, 6},
		{6, -1, 6},
		{13, 1, 14},
	} {
		a.settings.FontSize = step.from
		a.zoomFont(step.delta)
		if a.settings.FontSize != step.want {
			t.Errorf("size %v zoomed by %v = %v, want %v", step.from, step.delta, a.settings.FontSize, step.want)
		}
	}

	a.settings.FontSize = 32
	a.resetFont()
	if want := settings.Default().FontSize; a.settings.FontSize != want {
		t.Errorf("Reset Font Size = %v, want the default %v", a.settings.FontSize, want)
	}
	if _, err := os.Stat(settings.Path()); err != nil {
		t.Errorf("the zoom should save the settings: %v", err)
	}
}

// TestAdjustCellHeightBothWays checks that the row height follows a
// negative Adjust Cell Height as it does a positive one: the setting
// allows -0.5 to 0.5, and the negative half used to do nothing.
func TestAdjustCellHeightBothWays(t *testing.T) {
	a := &app{settings: settings.Default()}
	for adjust, want := range map[float64]float32{0.25: 1.25, -0.25: 0.75, 0: 0} {
		a.settings.AdjustCellHeight = adjust
		if got := a.termFont().LineHeight; got != want {
			t.Errorf("Adjust Cell Height %v: line height = %v, want %v", adjust, got, want)
		}
	}
}
