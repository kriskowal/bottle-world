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

func newColor(direction uint8, speed float64) color.Color {
	if direction == 0 {
		return color.RGBA{0, 0, 0, 0xff}
	}
	r, g, b := husl.HuslToRGB((float64(direction-1))*360/4, 50, speed*100)
	return color.RGBA{
		uint8(r * 0xff),
		uint8(g * 0xff),
		uint8(b * 0xff),
		0xff,
	}
}

func newPalette() color.Palette {
	pal := make(color.Palette, 0)
	for s := 0.0; s < 16.0; s++ {
		pal = append(pal, newColor(0, s/16))
	}
	var d uint8
	for d = 0; d < 4; d++ {
		for s := 0.0; s < 16.0; s++ {
			pal = append(pal, newColor(d, s/16))
		}
	}
	return pal
}

func render(w *sim.World) func(*sim.Cell) color.Color {
	return func(c *sim.Cell) color.Color {
		return newColor(
			c.WaterShed,
			float64(c.WaterSpeed)/float64(w.MostRapidWater),
		)
	}
}

func main() {
	prev := sim.NewWorld()
	next := prev

	pal := newPalette()
	images := make([]*image.Paletted, 0, next.Width)
	delays := make([]int, 0, next.Width)

	speed := 50
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

	f, _ := os.OpenFile("waterspeed.gif", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	gif.EncodeAll(f, &gif.GIF{
		Image: images,
		Delay: delays,
	})
}
