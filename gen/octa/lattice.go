package octa

import "math"

// Cell is a body-centered cubic lattice coordinate. All three components share parity.
// Its world center is (X, Y, Z) * span / 2; a cell's axis-aligned extent is span.
type Cell struct{ X, Y, Z int32 }

// MaxCoordinate bounds every component, matching the studio's validator.
const MaxCoordinate = 1_000_000

// CellRadius is the truncated octahedron circumradius for span 1: |(0, 1, 2)| / 4.
// Every point in space lies within CellRadius of some lattice center.
const CellRadius = 0.5590169943749475

// Neighbors are the 14 face-sharing offsets: 6 square faces, 8 hexagonal faces.
var Neighbors = [14]Cell{
	{2, 0, 0}, {-2, 0, 0}, {0, 2, 0}, {0, -2, 0}, {0, 0, 2}, {0, 0, -2},
	{1, 1, 1}, {1, 1, -1}, {1, -1, 1}, {1, -1, -1},
	{-1, 1, 1}, {-1, 1, -1}, {-1, -1, 1}, {-1, -1, -1},
}

func (c Cell) Valid() bool {
	return c.X&1 == c.Y&1 && c.Y&1 == c.Z&1 &&
		abs32(c.X) <= MaxCoordinate && abs32(c.Y) <= MaxCoordinate && abs32(c.Z) <= MaxCoordinate
}

func (c Cell) Add(o Cell) Cell { return Cell{c.X + o.X, c.Y + o.Y, c.Z + o.Z} }

// Center is the world-space center for cell span 1.
func (c Cell) Center() Vec3 { return Vec3{float64(c.X) * 0.5, float64(c.Y) * 0.5, float64(c.Z) * 0.5} }

// Nearest returns the lattice cell whose truncated octahedron contains p.
func Nearest(p Vec3) Cell {
	// The two candidate sublattices: all-even and all-odd coordinates.
	even := Cell{roundTo(p.X*2, 0), roundTo(p.Y*2, 0), roundTo(p.Z*2, 0)}
	odd := Cell{roundTo(p.X*2, 1), roundTo(p.Y*2, 1), roundTo(p.Z*2, 1)}
	if even.Center().Sub(p).Length() <= odd.Center().Sub(p).Length() {
		return even
	}
	return odd
}

// roundTo rounds v to the nearest integer with the given parity.
func roundTo(v float64, parity int32) int32 {
	n := int32(math.Floor((v-float64(parity))/2+0.5))*2 + parity
	return n
}

// eachCell calls f for every valid cell whose center lies inside the inclusive box.
func eachCell(b Box, f func(Cell, Vec3)) {
	if b.Empty() {
		return
	}
	y0, y1 := int32(math.Ceil(b.Min.Y*2)), int32(math.Floor(b.Max.Y*2))
	z0, z1 := int32(math.Ceil(b.Min.Z*2)), int32(math.Floor(b.Max.Z*2))
	x0, x1 := int32(math.Ceil(b.Min.X*2)), int32(math.Floor(b.Max.X*2))
	for y := y0; y <= y1; y++ {
		p := y & 1
		for z := z0 + ((z0&1)^p)&1; z <= z1; z += 2 {
			for x := x0 + ((x0&1)^p)&1; x <= x1; x += 2 {
				c := Cell{x, y, z}
				f(c, c.Center())
			}
		}
	}
}

// volumeSamples evenly fills one truncated octahedron: |x|,|y|,|z| ≤ ½ and |x|+|y|+|z| ≤ ¾.
var volumeSamples = func() []Vec3 {
	const steps = 8
	var result []Vec3
	for i := 0; i < steps; i++ {
		for j := 0; j < steps; j++ {
			for k := 0; k < steps; k++ {
				p := Vec3{(float64(i)+0.5)/steps - 0.5, (float64(j)+0.5)/steps - 0.5, (float64(k)+0.5)/steps - 0.5}
				if math.Abs(p.X)+math.Abs(p.Y)+math.Abs(p.Z) <= 0.75 {
					result = append(result, p)
				}
			}
		}
	}
	return result
}()

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
