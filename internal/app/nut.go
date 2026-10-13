package app

import (
	_ "embed"

	"github.com/egoist/mygo/ui"
)

//go:embed assets/nutty.svg
var nuttySVGSource string

// nuttySVG is the app's mark: the SVG Repo nut — a white beech-nut with
// a gray cap on black outlines, which reads on the black tile below.
var nuttySVG = ui.MustParseSVG([]byte(nuttySVGSource))

// nutTile draws the mark on a black tile, matching the app icon, so it
// keeps its contrast on the surface in both appearances.
func nutTile(c *ui.Context, size float32) ui.Element {
	return ui.Box(c).Size(size, size).Radius(size / 5).
		Background(ui.Hex("#000000")).Clip().Children(func() {
			ui.Column(c).Fill().Center().Children(func() {
				ui.Image(c, nuttySVG).Size(size*0.78, size*0.78).Shrink(0)
			})
		})
}
