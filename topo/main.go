package main

import (
	"image/color"

	"github.com/kriskowal/bottle-world/sim"
	"github.com/kriskowal/bottle-world/viz"
)

func main() {
	w := sim.NewWorld()

	breadth := float64(w.HighestSurfaceElevation - w.LowestSurfaceElevation)
	viz.Write(w, "topo.gif", viz.NewGrayScale(), func(c *sim.Cell) color.Color {
		z := uint8(float64(c.SurfaceElevation-w.LowestSurfaceElevation) / breadth * 0xff)
		return viz.Gray(z)
	})
}
