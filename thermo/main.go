package main

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"os"

	"github.com/kriskowal/bottle-world/sim"
	"github.com/kriskowal/bottle-world/viz"
)

func heat(w *sim.World) func(*sim.Cell) color.Color {
	return func(c *sim.Cell) color.Color {
		z := uint8(float64(c.SurfaceHeat) / float64(w.HottestSurface) * 0xff)
		return viz.Gray(z)
	}
}

func main() {
	prev := sim.NewWorld()
	next := prev

	pal := viz.NewGrayScale()
	images := make([]*image.Paletted, 0, next.Width)
	delays := make([]int, 0, next.Width)

	for t := 0; t < next.Width; t++ {
		fmt.Printf(".")

		next, prev = prev, next
		sim.Tick(next, prev, t)
		img := viz.Capture(next, pal, heat(next))
		images = append(images, img)
		delays = append(delays, 10)
	}

	f, _ := os.OpenFile("thermo.gif", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	gif.EncodeAll(f, &gif.GIF{
		Image: images,
		Delay: delays,
	})
}
