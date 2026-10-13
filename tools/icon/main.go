// Icon renders assets/nutty.svg into resources/icon.png, the app icon,
// with MyGo's own headless renderer: go run ./tools/icon
//
// The icon follows the reference format: the nut mark on a rounded black
// tile with a transparent background.
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/egoist/mygo/ui"
)

// size is the icon's side in pixels: the largest size MyGo packages into
// the app's .icns, so none of them is scaled up from a smaller render.
const size = 1024

func main() {
	src, err := os.ReadFile("internal/app/assets/nutty.svg")
	if err != nil {
		log.Fatal(err)
	}
	img := render(src)
	if err := os.MkdirAll("resources", 0o755); err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("resources/icon.png")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote resources/icon.png %dx%d", size, size)
}

// render draws the icon from the mark's SVG. The tile is drawn square:
// the mask below is what rounds its corners. Rounding it in the renderer
// as well left white slivers in the icon, where the renderer's corner
// curve and the mask's circle disagree.
func render(src []byte) *image.RGBA {
	svg := ui.MustParseSVG(src)
	tile := float32(0.84 * size)
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Center().Children(func() {
			ui.Box(c).Size(tile, tile).
				Background(ui.Hex("#000000")).Clip().Children(func() {
				ui.Column(c).Fill().Center().Children(func() {
					ui.Image(c, svg).Size(tile, tile)
				})
			})
		})
	}, size, size)
	img := tt.Image()
	// The headless renderer paints the theme's white background; drop
	// everything outside the tile and round its corners. The logo's own
	// white stays: it lies inside the tile.
	ts := int(tile)
	r := int(float32(ts) * 0.22)
	x0, y0 := (size-ts)/2, (size-ts)/2
	inTile := func(x, y int) bool {
		if x < x0 || y < y0 || x >= x0+ts || y >= y0+ts {
			return false
		}
		// Only the four corner quadrants cut in: a pixel there is inside
		// when it lies within the corner's circle; everywhere else in the
		// square is inside.
		cx, cy := x0+r, y0+r
		corner := false
		switch {
		case x < x0+r && y < y0+r:
			corner = true
		case x >= x0+ts-r && y < y0+r:
			cx, cy, corner = x0+ts-r, y0+r, true
		case x < x0+r && y >= y0+ts-r:
			cx, cy, corner = x0+r, y0+ts-r, true
		case x >= x0+ts-r && y >= y0+ts-r:
			cx, cy, corner = x0+ts-r, y0+ts-r, true
		}
		if !corner {
			return true
		}
		dx, dy := x-cx, y-cy
		return dx*dx+dy*dy <= r*r
	}
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			if !inTile(x, y) {
				img.SetRGBA(x, y, color.RGBA{})
			}
		}
	}
	return img
}
