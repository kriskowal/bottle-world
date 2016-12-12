package sim

type Source interface {
	Eval2(x, y float64) float64
}

func NewTesselation(source Source, width, height float64) Source {
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

func NewScale(source Source, s float64) Source {
	return &scale{source: source, scale: s}
}

type scale struct {
	source Source
	scale  float64
}

func (s *scale) Eval2(x, y float64) float64 {
	return s.source.Eval2(x*s.scale, y*s.scale)
}
