// Icon renders assets/nutty.svg into resources/icon.png, the app icon,
// with MyGo's own headless renderer: go run ./tools/icon
//
// The icon follows the reference format: the nut mark on a rounded black
// tile with a transparent background.
package main

import (
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/egoist/mygo/ui"
)

func main() {
	src, err := os.ReadFile("assets/nutty.svg")
	if err != nil {
		log.Fatal(err)
	}
	svg := ui.MustParseSVG(src)
	w, h := svg.Size()
	tile := 0.84 * float32(w)
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Center().Children(func() {
			ui.Box(c).Size(tile, tile).Radius(tile*0.22).
				Background(ui.Hex("#000000")).Clip().Children(func() {
					ui.Column(c).Fill().Center().Children(func() {
						ui.Image(c, svg).Size(tile, tile)
					})
				})
		})
	}, int(w), int(h))
	img := tt.Image()
	// The headless renderer paints the theme's white background; drop
	// everything outside the black tile, whose rounded corners it keeps.
	// The logo's own white stays: it lies inside the tile.
	ts := int(tile)
	r := int(float32(ts) * 0.22)
	x0, y0 := (int(w)-ts)/2, (int(h)-ts)/2
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
	if err := os.MkdirAll("resources", 0o755); err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("resources/icon.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote resources/icon.png %dx%d", int(w), int(h))
}
