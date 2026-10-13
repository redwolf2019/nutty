package app

import (
	_ "embed"

	"nutty/internal/settings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// The terminal font: JetBrains Mono Nerd Font (SIL Open Font License),
// v3.5.1's monospaced build from the nerd-fonts release, which carries the
// icon glyphs shell prompts draw alongside JetBrains Mono's own faces. The
// Mono variant keeps every glyph — the icons included — at one cell width,
// so prompts with symbols do not break column alignment.
//
//go:embed assets/JetBrainsMonoNerdFontMono-Regular.ttf
var jetbrainsMonoTTF []byte

//go:embed assets/JetBrainsMonoNerdFontMono-Bold.ttf
var jetbrainsMonoBoldTTF []byte

//go:embed assets/JetBrainsMonoNerdFontMono-Italic.ttf
var jetbrainsMonoItalicTTF []byte

//go:embed assets/JetBrainsMonoNerdFontMono-BoldItalic.ttf
var jetbrainsMonoBoldItalicTTF []byte

// registerFonts makes the font available to text before any window draws.
// All four faces register under "JetBrains Mono", the name the settings
// use, so Core Text matches bold and italic by trait instead of the app
// synthesizing them.
func registerFonts() {
	ui.RegisterFont(jetbrainsMonoTTF, "JetBrains Mono")
	ui.RegisterFont(jetbrainsMonoBoldTTF, "JetBrains Mono")
	ui.RegisterFont(jetbrainsMonoItalicTTF, "JetBrains Mono")
	ui.RegisterFont(jetbrainsMonoBoldItalicTTF, "JetBrains Mono")
}

// termFont is the terminal font the settings pick: otty's JetBrains Mono
// at 13 by default, with the ligatures, thickness and row height they set.
func (a *app) termFont() terminal.Font {
	s := a.settings
	if s == nil {
		s = settings.Default()
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
