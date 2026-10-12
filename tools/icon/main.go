// Icon renders assets/nutty.svg into resources/icon.png, the app icon,
// with MyGo's own headless renderer: go run ./tools/icon
//
// The icon follows the reference format: the acorn on a rounded black
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
	// The headless renderer paints the theme's white background; drop it
	// so the icon keeps a transparent background around the black tile.
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			p := img.RGBAAt(x, y)
			if p.A > 0 && p.R > 0xF0 && p.G > 0xF0 && p.B > 0xF0 {
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
