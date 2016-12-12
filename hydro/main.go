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

func newPalette() color.Palette {
	pal := make(color.Palette, 0)
	for h := 0; h < 16; h++ {
		for w := 0; w < 16; w++ {
			pal = append(pal, newColor(float64(h)/16, float64(w)/16))
		}
	}
	return pal
}

func newColor(w, h float64) color.Color {
	r, g, b := husl.HuslToRGB(240, 10+w*80, 10+h*80)
	return color.RGBA{
		uint8(r * 0xff),
		uint8(g * 0xff),
		uint8(b * 0xff),
		0xff,
	}
}

func render(w *sim.World) func(*sim.Cell) color.Color {
	return func(c *sim.Cell) color.Color {
		saturation := 0.0
		lightness := 0.0
		hydraulicLightness := float64(c.WaterElevation-w.LowestWaterElevation) / float64(w.HighestWaterElevation-w.LowestWaterElevation)
		topographicLightness := float64(c.SurfaceElevation-w.LowestSurfaceElevation) / float64(w.HighestSurfaceElevation-w.LowestSurfaceElevation)
		if c.Water < 10 {
			saturation = 0
			lightness = topographicLightness
		} else if c.Water < 20 {
			lightness = hydraulicLightness/2 + topographicLightness/2
			saturation = 0.5
		} else {
			lightness = hydraulicLightness
			saturation = 1
		}
		return newColor(saturation, lightness)
	}
}

func main() {
	prev := sim.NewWorld()
	next := prev

	pal := newPalette()
	images := make([]*image.Paletted, 0, next.Width)
	delays := make([]int, 0, next.Width)

	speed := 5
	duration := 1
	overture := 20000

	// Overture
	t := 0
	for ; t < overture; t++ {
		next, prev = prev, next
		sim.Tick(next, prev, t)
	}

	// Show
	for ; t < overture+next.Width*speed*duration; t++ {
		fmt.Printf(".")

		next, prev = prev, next
		sim.Tick(next, prev, t)
		if t%speed == 0 {
			fmt.Printf(".")
			images = append(images, viz.Capture(next, pal, render(next)))
			delays = append(delays, 10)
		}
	}

	f, _ := os.OpenFile("hydro.gif", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	gif.EncodeAll(f, &gif.GIF{
		Image: images,
		Delay: delays,
	})
}
