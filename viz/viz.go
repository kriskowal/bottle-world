package viz

import (
	"image"
	"image/color"
	"image/gif"
	"os"

	"github.com/kriskowal/bottle-world/sim"
)

func Gray(b uint8) color.Color {
	return color.RGBA{b, b, b, 0xff}
}

func NewGrayScale() color.Palette {
	pal := make(color.Palette, 0)
	for n := 0; n < 0xff; n++ {
		pal = append(pal, Gray(uint8(n)))
	}
	return pal
}

func Capture(w *sim.World, pal color.Palette, getColor func(c *sim.Cell) color.Color) *image.Paletted {
	width := w.Width * 5 / 4
	height := w.Height * 5 / 4
	img := image.NewPaletted(image.Rect(0, 0, width, height), pal)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, getColor(&w.Field[x%w.Width][y%w.Height]))
		}
	}
	return img
}

func Write(w *sim.World, file string, pal color.Palette, getColor func(c *sim.Cell) color.Color) {
	img := Capture(w, pal, getColor)
	f, _ := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	gif.EncodeAll(f, &gif.GIF{
		Image: []*image.Paletted{img},
		Delay: []int{0},
	})
}
