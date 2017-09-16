package main

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"os"

	"github.com/kriskowal/bottle-world/life/sim"
	"github.com/kriskowal/bottle-world/life/viz"
)

const file = "life.gif"

func getColor(c *sim.Cell) color.Color {
	if c.Life > 0 {
		return viz.Black
	}
	return viz.White
}

func main() {
	var worlds [2]sim.World
	a := &worlds[0]
	b := &worlds[1]
	// a.Reset()
	// b.Reset()

	pal := viz.NewGrayScale()
	_ = pal

	var imgs []*image.Paletted
	var dels []int

	speed := 1

	for t := 0; t < 250*speed; t++ {
		fmt.Printf(".")
		sim.Tick(a, b)

		if t%speed == 0 {
			img := viz.Capture(b, viz.GrayScale, getColor)
			imgs = append(imgs, img)
			dels = append(dels, 10)
		}

		a, b = b, a
	}

	f, _ := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	gif.EncodeAll(f, &gif.GIF{
		Image: imgs,
		Delay: dels,
	})
}
