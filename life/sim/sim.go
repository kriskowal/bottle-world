package sim

import (
	"math/rand"
)

const Width = 128
const Height = 128
const Area = Width * Height
const Deus = Area / 20

var Size = IntVec2{Width, Height}

type IntVec2 struct {
	X, Y int
}

func (i IntVec2) AddMod(k, m IntVec2) (j IntVec2) {
	j.X = (i.X + k.X + m.X) % m.X
	j.Y = (i.Y + k.Y + m.Y) % m.Y
	return
}

type Cell struct {
	Life int
}

type IntVec2Field [Width][Height]IntVec2

func (f *IntVec2Field) At(i IntVec2) *IntVec2 {
	return &f[i.X][i.Y]
}

type Field [Width][Height]Cell

func (f *Field) At(i IntVec2) *Cell {
	return &f[i.X][i.Y]
}

func NewIntVec2Field(r IntVec2) (f IntVec2Field) {
	var i IntVec2
	for i.X = 0; i.X < Width; i.X++ {
		for i.Y = 0; i.Y < Height; i.Y++ {
			*f.At(i) = i.AddMod(r, Size)
		}
	}
	return
}

var Neighbors []IntVec2

func init() {
	for x := 0; x < 3; x++ {
		for y := 0; y < 3; y++ {
			if !(x == 1 && y == 1) {
				Neighbors = append(Neighbors, IntVec2{x - 1, y - 1})
			}
		}
	}
}

func SumOfLife(f Field) (life int) {
	var i IntVec2
	for i.X = 0; i.X < Width; i.X++ {
		for i.Y = 0; i.Y < Height; i.Y++ {
			life += f.At(i).Life
		}
	}
	return
}

type World struct {
	Field Field
}

func (w *World) Reset() {
	var i IntVec2
	w.Field = Field{}
	for i.X = 0; i.X < Width; i.X++ {
		for i.Y = 0; i.Y < Height; i.Y++ {
			if rand.Intn(2) == 0 {
				w.Field.At(i).Life = 1
			}
		}
	}
}

func SumOfLifeAbout(f Field, i IntVec2) (life int) {
	for _, n := range Neighbors {
		life += f.At(i.AddMod(n, Size)).Life
	}
	return
}

func Tick(next, prev *World) {
	var i IntVec2

	total := SumOfLife(prev.Field)

	for i.X = 0; i.X < Width; i.X++ {
		for i.Y = 0; i.Y < Height; i.Y++ {
			s := SumOfLifeAbout(prev.Field, i)
			life := prev.Field.At(i).Life
			if life > 0 {
				if s < 2 || s > 3 {
					life = 0
				}
			} else {
				s += rand.Intn(3) - 1
				if s == 0 && total == 0 && rand.Intn(Area) == 1 {
					life = 1
				} else if s < 3 && total < 10 && rand.Intn(20) == 0 {
					life = 1
				} else if s == 3 {
					life = 1
				}
			}
			next.Field.At(i).Life = life
		}
	}
}
