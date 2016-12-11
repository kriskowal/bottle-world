package main

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"log"
	"os"

	"github.com/husl-colors/husl-go"
	"github.com/ojrac/opensimplex-go"
)

type Source interface {
	Eval2(x, y float64) float64
}

func NewTesselation(source Source, width, height float64) *tesselation {
	return &tesselation{
		source: source,
		width:  width,
		height: height,
	}
}

type tesselation struct {
	source        Source
	width, height float64
}

func (t *tesselation) Eval2(x, y float64) float64 {
	a := t.source.Eval2(float64(x), float64(y))
	b := t.source.Eval2(float64(x-t.width), float64(y))
	ab := a*(1.0-float64(x)/t.width) + b*(float64(x)/t.width)
	c := t.source.Eval2(float64(x), float64(y-t.height))
	d := t.source.Eval2(float64(x-t.width), float64(y-t.height))
	cd := c*(1.0-float64(x)/t.width) + d*(float64(x)/t.width)
	return ab*(1.0-float64(y)/height) + cd*float64(y)/height
}

func NewScale(source Source, s float64) *scale {
	return &scale{source: source, scale: s}
}

type scale struct {
	source Source
	scale  float64
}

func (s *scale) Eval2(x, y float64) float64 {
	return s.source.Eval2(x*s.scale, y*s.scale)
}

// humidity
// area of water surface
// temperature of water surface
// temperature of air
// 540 calories per gram is latent heat of vaporization
// water convection
// airflow about the surface of the water
// (mass loss rate)/(unit area) = (vapor pressure - ambient partial pressure)*sqrt( (molecular weight)/(2*pi*R*T) )
// Nitrogen is 0.185 cm**2/sec at room temperature and 1 atm.

type point2 struct {
	x int
	y int
}

// TODO start with ice, then melt, then flow, then evaporate, then precipitate, then freeze
type cell struct {
	Light       int
	Terrain     int
	SurfaceHeat int
	Water       int // Height of water column over terrain, under ice
	WaterHeight int // Absolute height of water column
	Watershed   uint8
	// WaterDX     int
	// WaterDY     int
	// WaterHeat int
	// Steam     int
	// SteamHeat int
	// Air       int
	// AirHeat   int
}

const seed = 4
const seed2 = 5
const width = 128
const height = 128
const terrainScale = 100
const terrainScale2 = 10
const simplexScale = 1.0 / 20
const simplexScale2 = 1.0 / 8
const dissipation = 10 / 11
const totalWater = width * height * 100

type field [width][height]cell

type world struct {
	height    int
	depth     int
	hottest   int
	brightest int
	wettest   int
	highest   int
	lowest    int
	// equatorial
	eqminheat int
	eqmaxheat int
	// latitudinal (height/4, height*3/4)
	latminheat int
	latmaxheat int
	field      field
}

func reset(w *world) {
	scales := []struct {
		seed                       int64
		terrainScale, simplexScale float64
	}{
		{2, 125, 1.0 / 80},
		{3, 100, 1.0 / 40},
		{5, 75, 1.0 / 30},
		{4, 50, 1.0 / 10},
		{4, 20, 1.0 / 6},
		{5, 10, 1.0 / 4},
		{5, 5, 1.0 / 2},
	}
	noises := make([]struct {
		scale  float64
		source Source
	}, 0, len(scales))
	for _, s := range scales {
		noises = append(noises, struct {
			scale  float64
			source Source
		}{
			scale:  s.terrainScale * 10,
			source: NewTesselation(NewScale(opensimplex.NewWithSeed(s.seed), s.simplexScale), width, height),
		})
	}

	w.depth = 0
	w.height = 0
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			l := 0.0
			for _, n := range noises {
				l += n.scale * n.source.Eval2(float64(x), float64(y))
			}
			el := int(l)
			if el > w.height {
				w.height = el
			}
			if el < w.depth {
				w.depth = el
			}
			c := &w.field[x][y]
			c.Terrain = el
			c.Water = totalWater / width / height
		}
	}
}

func manhattan(x1, y1, x2, y2 int) int {
	dx := x2 - x1
	if dx < 0 {
		dx = -dx
	}
	if dx > width/2 {
		dx = width - dx
	}
	dy := y2 - y1
	if dy < 0 {
		dy = -dy
	}
	if dy > height/2 {
		dy = height - dy
	}
	d := dx + dy
	return d
}

func tick(next, prev *world, t int) {
	next.height = prev.height
	next.depth = prev.depth

	sx := width - (t % width)
	sy := height / 2

	// Reset
	for x := 0; x < width; x++ {
		for y := 0; y < width; y++ {
			// Compute water gradient
			pc := &prev.field[x][y]
			nc := &next.field[x][y]
			pc.WaterHeight = pc.Terrain + pc.Water
			nc.Terrain = pc.Terrain
			nc.Water = pc.Water
		}
	}

	// Distribute water
	for x := 0; x < width; x++ {
		for y := 0; y < width; y++ {
			// Compute water gradient
			// {prev,next}cell{north,south,east,west}
			pc := &prev.field[x][y]
			pcn := &prev.field[x][(y+height-1)%height]
			pcs := &prev.field[x][(y+1)%height]
			pcw := &prev.field[(x+width-1)%width][y]
			pce := &prev.field[(x+1)%width][y]

			nc := &next.field[x][y]
			ncn := &next.field[x][(y+height-1)%height]
			ncs := &next.field[x][(y+1)%height]
			ncw := &next.field[(x+width-1)%width][y]
			nce := &next.field[(x+1)%width][y]

			pt := pc
			nt := nc
			nc.Watershed = pc.Watershed
			if pcn.WaterHeight < pt.WaterHeight {
				pt = pcn
				nt = ncn
				nc.Watershed = 1
			}
			if pcs.WaterHeight < pt.WaterHeight {
				pt = pcs
				nt = ncs
				nc.Watershed = 2
			}
			if pcw.WaterHeight < pt.WaterHeight {
				pt = pcw
				nt = ncw
				nc.Watershed = 3
			}
			if pce.WaterHeight < pt.WaterHeight {
				pt = pce
				nt = nce
				nc.Watershed = 4
			}

			equilibrium := pc.WaterHeight/2 + pt.WaterHeight/2
			delta := pc.WaterHeight - equilibrium
			if delta > pc.Water {
				delta = pc.Water
			}
			// dampen water flow
			if delta > 2 {
				delta = delta / 2
			}
			nt.Water += delta
			nc.Water -= delta

		}
	}

	// Bathymetry
	next.wettest = 0
	next.highest = 0
	next.lowest = 1000000000
	for x := 0; x < width; x++ {
		for y := 0; y < width; y++ {
			// Compute water gradient
			nc := &next.field[x][y]
			nc.WaterHeight = nc.Terrain + nc.Water
			if nc.WaterHeight > next.highest {
				next.highest = nc.WaterHeight
			}
			if nc.WaterHeight < next.lowest {
				next.lowest = nc.WaterHeight
			}
			if nc.Water > next.wettest {
				next.wettest = nc.Water
			}
		}
	}

	// Distribute heat
	next.hottest = 0
	next.brightest = 0
	for x := 0; x < width; x++ {
		for y := 0; y < width; y++ {
			nc := &next.field[x][y]
			pc := &prev.field[x][y]
			pcn := &prev.field[x][(y+height-1)%height]
			pcs := &prev.field[x][(y+1)%height]
			pcw := &prev.field[(x+width-1)%width][y]
			pce := &prev.field[(x+1)%width][y]

			// diffuse heat from prior turn
			heat := (pcn.SurfaceHeat + pcs.SurfaceHeat + pce.SurfaceHeat + pcw.SurfaceHeat + pc.SurfaceHeat) / 5
			// heat := pc.SurfaceHeat

			// distribute heat according to the distance from direct sunlight
			d := manhattan(sx, sy, x, y)
			dh := width*3/5 - d
			if dh < 0 {
				dh = 0
			}
			nc.Light = dh
			if nc.Light > next.brightest {
				next.brightest = nc.Light
			}

			// dissipate heat through radiation
			nc.SurfaceHeat = (heat + dh) * 100 / 102

			if nc.SurfaceHeat > next.hottest {
				next.hottest = nc.SurfaceHeat
			}
		}
	}

	// Recalculate equatorial minima and maxima
	y := height / 2
	next.eqminheat = next.hottest
	next.eqmaxheat = 0
	for x := 0; x < width; x++ {
		heat := next.field[x][y].SurfaceHeat
		if heat < next.eqminheat {
			next.eqminheat = heat
		}
		if heat > next.eqmaxheat {
			next.eqmaxheat = heat
		}
	}

	// Recalculate latitudinal maxima and minima
	y = height / 4
	next.latminheat = next.hottest
	next.latmaxheat = 0
	for x := 0; x < width; x++ {
		heat := next.field[x][y].SurfaceHeat
		if heat < next.latminheat {
			next.latminheat = heat
		}
		if heat > next.latmaxheat {
			next.latmaxheat = heat
		}
	}
}

func gray(b uint8) color.Color {
	return color.RGBA{b, b, b, 0xff}
}

func capture(w *world, pal color.Palette, getColor func(c *cell) color.Color) *image.Paletted {
	img := image.NewPaletted(image.Rect(0, 0, width*1.25, height*1.25), pal)
	for x := 0; x < width*1.25; x++ {
		for y := 0; y < height*1.25; y++ {
			img.Set(x, y, getColor(&w.field[x%width][y%height]))
		}
	}
	return img
}

func write(w *world, file string, getColor func(c *cell) color.Color) {
	img := capture(w, nil, getColor)
	f, _ := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	gif.EncodeAll(f, &gif.GIF{
		Image: []*image.Paletted{img},
		Delay: []int{0},
	})
}

func writeTerrain(w *world) {
	breadth := float64(w.height - w.depth)
	write(w, "terrain.gif", func(c *cell) color.Color {
		z := uint8(float64(c.Terrain-w.depth) / breadth * 0xff)
		return gray(z)
	})
}

func writeHeat(w *world) {
	write(w, "heat.gif", func(c *cell) color.Color {
		z := uint8(float64(c.SurfaceHeat) / float64(w.hottest) * 0xff)
		return gray(z)
	})
}

func bathcolor(w, h float64) color.Color {
	r, g, b := husl.HuslToRGB(240, 10+w*80, 10+h*80)
	return color.RGBA{
		uint8(r * 0xff),
		uint8(g * 0xff),
		uint8(b * 0xff),
		0xff,
	}
}

func newBathscale() color.Palette {
	pal := make(color.Palette, 0)
	for h := 0; h < 16; h++ {
		for w := 0; w < 16; w++ {
			pal = append(pal, bathcolor(float64(h)/16, float64(w)/16))
		}
	}
	return pal
}

var bathscale = newBathscale()

func heat(w *world) func(*cell) color.Color {
	return func(c *cell) color.Color {
		z := uint8(float64(c.SurfaceHeat) / float64(w.hottest) * 0xff)
		return gray(z)
	}
}

func terrainModel() {
	next := &world{}
	reset(next)
	writeTerrain(next)
}

func thermalModel() {
	prev := &world{}
	next := &world{}
	reset(next)
	t := 0
	for ; t < width*height; t++ {
		next, prev = prev, next
		tick(next, prev, t)
	}
	images := make([]*image.Paletted, 0, width)
	delays := make([]int, 0, width)
	for ; t < width*height+width; t++ {
		fmt.Printf(".")
		next, prev = prev, next
		tick(next, prev, t)
		if t%4 == 0 {
			images = append(images, capture(next, bathscale, heat(next)))
			delays = append(delays, 10)
		}
	}

	f, _ := os.OpenFile("day.gif", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	gif.EncodeAll(f, &gif.GIF{
		Image: images,
		Delay: delays,
	})

	fmt.Println(next.eqminheat, next.eqmaxheat, next.latminheat, next.latmaxheat)
}

func bathymetry(w *world) func(*cell) color.Color {
	return func(c *cell) color.Color {
		saturation := 0.0
		lightness := 0.0
		hydraulicLightness := 0.5 - 0.5*float64(c.Water)/float64(w.wettest)
		topographicLightness := float64(c.Terrain-w.depth) / float64(w.height-w.depth)
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
		return bathcolor(
			saturation,
			lightness,
		)
	}
}

func bathymetryModel() {
	prev := &world{}
	next := &world{}
	reset(next)

	writeTerrain(next)

	images := make([]*image.Paletted, 0, width)
	delays := make([]int, 0, width)

	speed := 1
	duration := 1
	prime := 1000

	t := 0
	for ; t < prime; t++ {
		next, prev = prev, next
		tick(next, prev, t)
	}
	for ; t < prime+width*speed*duration; t++ {
		next, prev = prev, next
		tick(next, prev, t)
		if t%speed == 0 {
			fmt.Printf(".")
			images = append(images, capture(next, bathscale, bathymetry(next)))
			delays = append(delays, 10)
		}
	}

	fmt.Println(next.highest, next.wettest)

	f, err := os.OpenFile("bathymetry.gif", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	if err := gif.EncodeAll(f, &gif.GIF{
		Image: images,
		Delay: delays,
	}); err != nil {
		log.Fatalln(err)
	}
}

func sealevels(w *world) func(*cell) color.Color {
	return func(c *cell) color.Color {
		saturation := 0.0
		lightness := 0.0
		hydraulicLightness := float64(c.WaterHeight-w.lowest) / float64(w.highest-w.lowest)
		topographicLightness := float64(c.Terrain-w.depth) / float64(w.height-w.depth)
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
		return bathcolor(
			saturation,
			lightness,
		)
	}
}

func waterLevelModel() {
	prev := &world{}
	next := &world{}
	reset(next)

	writeTerrain(next)

	images := make([]*image.Paletted, 0, width)
	delays := make([]int, 0, width)

	speed := 5
	duration := 1
	prime := 20000

	t := 0
	for ; t < prime; t++ {
		next, prev = prev, next
		tick(next, prev, t)
	}
	for ; t < prime+width*speed*duration; t++ {
		next, prev = prev, next
		tick(next, prev, t)
		if t%speed == 0 {
			fmt.Printf(".")
			images = append(images, capture(next, bathscale, sealevels(next)))
			delays = append(delays, 10)
		}
	}

	fmt.Println(next.highest, next.wettest)

	f, err := os.OpenFile("sealevels.gif", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	if err := gif.EncodeAll(f, &gif.GIF{
		Image: images,
		Delay: delays,
	}); err != nil {
		log.Fatalln(err)
	}
}

func watershedModel() {
	prev := &world{}
	next := &world{}
	reset(next)

	images := make([]*image.Paletted, 0, width)
	delays := make([]int, 0, width)

	speed := 5
	duration := 1
	prime := 20000

	t := 0
	for ; t < prime; t++ {
		next, prev = prev, next
		tick(next, prev, t)
	}
	for ; t < prime+width*speed*duration; t++ {
		next, prev = prev, next
		tick(next, prev, t)
		if t%speed == 0 {
			fmt.Printf(".")
			images = append(images, capture(next, bathscale, sealevels(next)))
			delays = append(delays, 10)
		}
	}

	fmt.Println(next.highest, next.wettest)

	f, err := os.OpenFile("sealevels.gif", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	if err := gif.EncodeAll(f, &gif.GIF{
		Image: images,
		Delay: delays,
	}); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	terrainModel()
}
