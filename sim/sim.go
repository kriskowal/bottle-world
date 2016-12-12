package sim

import "github.com/ojrac/opensimplex-go"

// humidity
// area of water surface
// temperature of water surface
// temperature of air
// 540 calories per gram is latent heat of vaporization
// water convection
// airflow about the surface of the water
// (mass loss rate)/(unit area) = (vapor pressure - ambient partial pressure)*sqrt( (molecular weight)/(2*pi*R*T) )
// Nitrogen is 0.185 cm**2/sec at room temperature and 1 atm.

// TODO start with ice, then melt, then flow, then evaporate, then precipitate, then freeze
type Cell struct {
	Height           int
	Width            int
	SunLight         int
	SurfaceElevation int
	SurfaceHeat      int
	Water            int // Height of water column over terrain, under ice
	WaterElevation   int // Absolute height of water column
	WaterShed        uint8
	WaterSpeed       int
	// WaterDX     int
	// WaterDY     int
	// WaterHeat int
	// Steam     int
	// SteamHeat int
	// Air       int
	// AirHeat   int
}

type Field [width][height]Cell

type World struct {
	Height int
	Width  int

	HighestSurfaceElevation int
	LowestSurfaceElevation  int
	HottestSurface          int
	BrightestSurface        int
	Wettest                 int
	HighestWaterElevation   int
	LowestWaterElevation    int
	MostRapidWater          int
	// equatorial
	EquatorialMinimumSurfaceHeat int
	EquatorialMaximumSurfaceHeat int
	// latitudinal (height/4, height*3/4)
	Latminheat int
	Latmaxheat int
	Field      Field
}

func Reset(w *World) {
	w.Height = height
	w.Width = width

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

	w.LowestSurfaceElevation = 0
	w.HighestSurfaceElevation = 0
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			l := 0.0
			for _, n := range noises {
				l += n.scale * n.source.Eval2(float64(x), float64(y))
			}
			el := int(l)
			if el > w.HighestSurfaceElevation {
				w.HighestSurfaceElevation = el
			}
			if el < w.LowestSurfaceElevation {
				w.LowestSurfaceElevation = el
			}
			c := &w.Field[x][y]
			c.SurfaceElevation = el
			c.Water = totalWater / width / height
		}
	}
}

func NewWorld() *World {
	world := &World{}
	Reset(world)
	return world
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

func Tick(next, prev *World, t int) {
	next.HighestSurfaceElevation = prev.HighestSurfaceElevation
	next.LowestSurfaceElevation = prev.LowestSurfaceElevation
	next.MostRapidWater = 0

	sx := width - (t % width)
	sy := height / 2

	// Reset
	for x := 0; x < width; x++ {
		for y := 0; y < width; y++ {
			// Compute water gradient
			pc := &prev.Field[x][y]
			nc := &next.Field[x][y]
			pc.WaterElevation = pc.SurfaceElevation + pc.Water
			nc.SurfaceElevation = pc.SurfaceElevation
			nc.Water = pc.Water
		}
	}

	// Distribute water
	for x := 0; x < width; x++ {
		for y := 0; y < width; y++ {
			// Compute water gradient
			// {prev,next}cell{north,south,east,west}
			pc := &prev.Field[x][y]
			pcn := &prev.Field[x][(y+height-1)%height]
			pcs := &prev.Field[x][(y+1)%height]
			pcw := &prev.Field[(x+width-1)%width][y]
			pce := &prev.Field[(x+1)%width][y]

			nc := &next.Field[x][y]
			ncn := &next.Field[x][(y+height-1)%height]
			ncs := &next.Field[x][(y+1)%height]
			ncw := &next.Field[(x+width-1)%width][y]
			nce := &next.Field[(x+1)%width][y]

			pt := pc
			nt := nc
			nc.WaterShed = pc.WaterShed
			if pcn.WaterElevation < pt.WaterElevation {
				pt = pcn
				nt = ncn
				nc.WaterShed = 1
			}
			if pcs.WaterElevation < pt.WaterElevation {
				pt = pcs
				nt = ncs
				nc.WaterShed = 2
			}
			if pcw.WaterElevation < pt.WaterElevation {
				pt = pcw
				nt = ncw
				nc.WaterShed = 3
			}
			if pce.WaterElevation < pt.WaterElevation {
				pt = pce
				nt = nce
				nc.WaterShed = 4
			}

			equilibrium := pc.WaterElevation/2 + pt.WaterElevation/2
			delta := pc.WaterElevation - equilibrium
			if delta > pc.Water {
				delta = pc.Water
			}
			// dampen water flow
			if delta > 3 {
				delta = delta / 3
			}
			nt.Water += delta
			nc.Water -= delta

			nc.WaterSpeed = delta
			if nc.WaterSpeed > next.MostRapidWater {
				next.MostRapidWater = nc.WaterSpeed
			}

		}
	}

	// Bathymetry
	next.Wettest = 0
	next.HighestWaterElevation = 0
	next.LowestWaterElevation = 1000000000
	for x := 0; x < width; x++ {
		for y := 0; y < width; y++ {
			// Compute water gradient
			nc := &next.Field[x][y]
			nc.WaterElevation = nc.SurfaceElevation + nc.Water
			if nc.WaterElevation > next.HighestWaterElevation {
				next.HighestWaterElevation = nc.WaterElevation
			}
			if nc.WaterElevation < next.LowestWaterElevation {
				next.LowestWaterElevation = nc.WaterElevation
			}
			if nc.Water > next.Wettest {
				next.Wettest = nc.Water
			}
		}
	}

	// Distribute heat
	next.HottestSurface = 0
	next.BrightestSurface = 0
	for x := 0; x < width; x++ {
		for y := 0; y < width; y++ {
			nc := &next.Field[x][y]
			pc := &prev.Field[x][y]
			pcn := &prev.Field[x][(y+height-1)%height]
			pcs := &prev.Field[x][(y+1)%height]
			pcw := &prev.Field[(x+width-1)%width][y]
			pce := &prev.Field[(x+1)%width][y]

			// diffuse heat from prior turn
			heat := (pcn.SurfaceHeat + pcs.SurfaceHeat + pce.SurfaceHeat + pcw.SurfaceHeat + pc.SurfaceHeat) / 5
			// heat := pc.SurfaceHeat

			// distribute heat according to the distance from direct sunlight
			d := manhattan(sx, sy, x, y)
			dh := width*3/5 - d
			if dh < 0 {
				dh = 0
			}
			nc.SunLight = dh
			if nc.SunLight > next.BrightestSurface {
				next.BrightestSurface = nc.SunLight
			}

			// dissipate heat through radiation
			nc.SurfaceHeat = (heat + dh) * 100 / 102

			if nc.SurfaceHeat > next.HottestSurface {
				next.HottestSurface = nc.SurfaceHeat
			}
		}
	}

	// Recalculate equatorial minima and maxima
	y := height / 2
	next.EquatorialMinimumSurfaceHeat = next.HottestSurface
	next.EquatorialMaximumSurfaceHeat = 0
	for x := 0; x < width; x++ {
		heat := next.Field[x][y].SurfaceHeat
		if heat < next.EquatorialMinimumSurfaceHeat {
			next.EquatorialMinimumSurfaceHeat = heat
		}
		if heat > next.EquatorialMaximumSurfaceHeat {
			next.EquatorialMaximumSurfaceHeat = heat
		}
	}

	// Recalculate latitudinal maxima and minima
	y = height / 4
	next.Latminheat = next.HottestSurface
	next.Latmaxheat = 0
	for x := 0; x < width; x++ {
		heat := next.Field[x][y].SurfaceHeat
		if heat < next.Latminheat {
			next.Latminheat = heat
		}
		if heat > next.Latmaxheat {
			next.Latmaxheat = heat
		}
	}
}
