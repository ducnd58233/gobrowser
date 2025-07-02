package types

type Bounds struct {
	X, Y, Width, Height float64
}

func (b *Bounds) Contains(x, y float64) bool {
	return x >= b.X && x <= b.X+b.Width && y >= b.Y && y <= b.Y+b.Height
}

func (b *Bounds) Intersects(other *Bounds) bool {
	return !(b.X+b.Width < other.X || other.X+other.Width < b.X ||
		b.Y+b.Height < other.Y || other.Y+other.Height < b.Y)
}
