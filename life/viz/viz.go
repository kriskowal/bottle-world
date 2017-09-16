package viz

import (
	"image"
	"image/color"

	"github.com/kriskowal/bottle-world/life/sim"
)

func Gray(b uint8) color.Color {
	return color.RGBA{b, b, b, 0xff}
}

func NewGrayScale() color.Palette {
	pal := make(color.Palette, 0)
	for n := 0; n < 0x100; n++ {
		pal = append(pal, Gray(uint8(n)))
	}
	return pal
}

var GrayScale = NewGrayScale()

var Black = GrayScale[0]
var White = GrayScale[0xff]

func Capture(w *sim.World, pal color.Palette, getColor func(c *sim.Cell) color.Color) *image.Paletted {
	img := image.NewPaletted(image.Rect(0, 0, sim.Width, sim.Height), pal)

	var i sim.IntVec2
	for i.X = 0; i.X < sim.Width; i.X++ {
		for i.Y = 0; i.Y < sim.Height; i.Y++ {
			img.Set(i.X, i.Y, getColor(w.Field.At(i.AddMod(sim.IntVec2{}, sim.Size))))
		}
	}
	return img
}
