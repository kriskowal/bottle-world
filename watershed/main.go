package main

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"os"

	"github.com/husl-colors/husl-go"
	"github.com/kriskowal/bottle-world/sim"
	"github.com/kriskowal/bottle-world/viz"
)

func newColor(n uint8) color.Color {
	r, g, b := husl.HuslToRGB(float64(n-1)/4*360, 100.0, 50.0)
	return color.RGBA{uint8(r * 0xff), uint8(g * 0xff), uint8(b * 0xff), 0xff}
}

var pal = color.Palette{
	color.RGBA{0xff, 0xff, 0xff, 0xff},
	newColor(0),
	newColor(1),
	newColor(2),
	newColor(3),
}

func render(w *sim.World) func(*sim.Cell) color.Color {
	return func(c *sim.Cell) color.Color {
		return pal[c.WaterShed]
	}
}

func main() {
	prev := sim.NewWorld()
	next := prev

	images := make([]*image.Paletted, 0, next.Width)
	delays := make([]int, 0, next.Width)

	speed := 100
	duration := 4
	overture := 0

	// Overture
	t := 0
	for ; t < overture; t++ {
		next, prev = prev, next
		sim.Tick(next, prev, t)
	}

	// Show
	for ; t < overture+next.Width*speed*duration; t++ {
		next, prev = prev, next
		sim.Tick(next, prev, t)
		if t%speed == 0 {
			fmt.Printf(".")
			images = append(images, viz.Capture(next, pal, render(next)))
			delays = append(delays, 10)
		}
	}

	f, _ := os.OpenFile("watershed.gif", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	gif.EncodeAll(f, &gif.GIF{
		Image: images,
		Delay: delays,
	})
}
