package main

import (
	_ "embed"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// The fonts otty ships, so Nutty's terminal shows the same text: JetBrains
// Mono (SIL Open Font License) in its regular and italic faces.
//
//go:embed assets/JetBrainsMono.ttf
var jetbrainsMonoTTF []byte

//go:embed assets/JetBrainsMono-Italic.ttf
var jetbrainsMonoItalicTTF []byte

// registerFonts makes JetBrains Mono available to text before any window
// draws. The italic registers under the same family, so Core Text matches
// it by trait, as the regular's bold is synthesized by weight.
func registerFonts() {
	ui.RegisterFont(jetbrainsMonoTTF, "JetBrains Mono")
	ui.RegisterFont(jetbrainsMonoItalicTTF, "JetBrains Mono")
}

// termFont is the terminal font the settings pick: otty's JetBrains Mono
// at 13 by default, with the ligatures, thickness and row height they set.
func (a *app) termFont() terminal.Font {
	s := a.settings
	if s == nil {
		s = defaultSettings()
	}
	f := terminal.Font{Family: s.FontFamily, Size: float32(s.FontSize)}
	if f.Family == "" {
		f.Family = "JetBrains Mono"
	}
	if f.Size <= 0 {
		f.Size = 13
	}
	switch s.FontLigatures {
	case "calt":
		f.Features = []string{"calt"}
	case "off":
		f.Features = []string{"-calt", "-liga", "-dlig"}
	default: // "dlig", otty's default ligature set
		f.Features = []string{"dlig"}
	}
	if s.FontThicken {
		f.Thicken = true
	}
	if s.AdjustCellHeight > 0 {
		f.LineHeight = 1 + float32(s.AdjustCellHeight)
	}
	return f
}
