package main

import (
	"image"
	"image/png"
	"os"
	"testing"
)

// TestIconEdge checks the icon the tool renders, and the one committed in
// resources/, which MyGo packages: 1024 pixels square, transparent
// outside the tile, and dark all along the tile's outline. The tile used
// to be rounded twice, by the renderer and by the mask, and where the two
// curves disagreed the white background showed as slivers at the corners.
func TestIconEdge(t *testing.T) {
	src, err := os.ReadFile("../../internal/app/assets/nutty.svg")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("rendered", func(t *testing.T) { checkIcon(t, render(src)) })
	t.Run("committed", func(t *testing.T) {
		f, err := os.Open("../../resources/icon.png")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = f.Close() }()
		img, err := png.Decode(f)
		if err != nil {
			t.Fatal(err)
		}
		checkIcon(t, img)
	})
}

func checkIcon(t *testing.T, img image.Image) {
	t.Helper()
	b := img.Bounds()
	if b.Dx() != size || b.Dy() != size {
		t.Fatalf("icon is %dx%d, want %dx%d", b.Dx(), b.Dy(), size, size)
	}
	clear := func(x, y int) bool {
		if !image.Pt(x, y).In(b) {
			return true
		}
		_, _, _, a := img.At(x, y).RGBA()
		return a == 0
	}
	for _, p := range []image.Point{b.Min, {b.Max.X - 1, b.Min.Y}, {b.Min.X, b.Max.Y - 1}, {b.Max.X - 1, b.Max.Y - 1}} {
		if !clear(p.X, p.Y) {
			t.Errorf("the corner %v of the image is not transparent", p)
		}
	}
	// The outline is every opaque pixel that touches a transparent one:
	// the tile's black, or the mark's dark ink where the corner cuts it,
	// but never the background's white.
	const dark = 0x4000
	outline, stray := 0, 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if clear(x, y) {
				continue
			}
			if !clear(x-1, y) && !clear(x+1, y) && !clear(x, y-1) && !clear(x, y+1) {
				continue // inside the tile
			}
			outline++
			if r, g, bl, _ := img.At(x, y).RGBA(); r > dark || g > dark || bl > dark {
				if stray++; stray <= 3 {
					t.Errorf("the outline pixel (%d,%d) is light: %04x %04x %04x", x, y, r, g, bl)
				}
			}
		}
	}
	if outline == 0 {
		t.Fatal("the icon has no opaque tile")
	}
	if stray > 0 {
		t.Errorf("%d of the %d outline pixels are light", stray, outline)
	}
}
