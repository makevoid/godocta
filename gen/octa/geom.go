// Package octa builds truncated-octahedron structures on a body-centered cubic lattice
// and writes OCTA v1 (uncolored) or v2 (colored) files for the Godot studio.
//
// The pipeline is: describe shapes as Solids (CSG primitives in world units), rasterize
// them into a Grid of lattice cells, group them into reusable Components, then compile
// an ordered Structure. The order is the assembly order the studio animates.
package octa

import "math"

// Vec3 is a world-space point. Y is up; recipe dimensions are independent of cell span.
type Vec3 struct{ X, Y, Z float64 }

func V(x, y, z float64) Vec3 { return Vec3{x, y, z} }

func (a Vec3) Add(b Vec3) Vec3      { return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z} }
func (a Vec3) Sub(b Vec3) Vec3      { return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z} }
func (a Vec3) Scale(s float64) Vec3 { return Vec3{a.X * s, a.Y * s, a.Z * s} }
func (a Vec3) Dot(b Vec3) float64   { return a.X*b.X + a.Y*b.Y + a.Z*b.Z }
func (a Vec3) Length() float64      { return math.Sqrt(a.Dot(a)) }

// Plan drops the height: the point seen from above (X east, Y = world Z).
func (a Vec3) Plan() Vec2 { return Vec2{a.X, a.Z} }

// Vec2 is a plan-view point. Its Y is the world Z axis.
type Vec2 struct{ X, Y float64 }

func P(x, y float64) Vec2 { return Vec2{x, y} }

func (a Vec2) Add(b Vec2) Vec2      { return Vec2{a.X + b.X, a.Y + b.Y} }
func (a Vec2) Sub(b Vec2) Vec2      { return Vec2{a.X - b.X, a.Y - b.Y} }
func (a Vec2) Scale(s float64) Vec2 { return Vec2{a.X * s, a.Y * s} }
func (a Vec2) Dot(b Vec2) float64   { return a.X*b.X + a.Y*b.Y }
func (a Vec2) Length() float64      { return math.Hypot(a.X, a.Y) }

// Angle is measured from +X towards +Y (world +Z), in radians.
func (a Vec2) Angle() float64 { return math.Atan2(a.Y, a.X) }

// Rotate turns the point around the origin by angle radians.
func (a Vec2) Rotate(angle float64) Vec2 {
	s, c := math.Sincos(angle)
	return Vec2{a.X*c - a.Y*s, a.X*s + a.Y*c}
}

// At lifts a plan point to a world point at height y.
func (a Vec2) At(y float64) Vec3 { return Vec3{a.X, y, a.Y} }

// Polar returns the plan point at radius r and angle a around center.
func Polar(center Vec2, r, a float64) Vec2 {
	s, c := math.Sincos(a)
	return Vec2{center.X + r*c, center.Y + r*s}
}

// Box is an inclusive axis-aligned box.
type Box struct{ Min, Max Vec3 }

// BoxOf returns the box spanning two corners in any order.
func BoxOf(a, b Vec3) Box {
	return Box{
		Vec3{math.Min(a.X, b.X), math.Min(a.Y, b.Y), math.Min(a.Z, b.Z)},
		Vec3{math.Max(a.X, b.X), math.Max(a.Y, b.Y), math.Max(a.Z, b.Z)},
	}
}

// EmptyBox is the identity for Union.
func EmptyBox() Box {
	inf := math.Inf(1)
	return Box{Vec3{inf, inf, inf}, Vec3{-inf, -inf, -inf}}
}

func (b Box) Empty() bool { return b.Min.X > b.Max.X || b.Min.Y > b.Max.Y || b.Min.Z > b.Max.Z }
func (b Box) Size() Vec3  { return b.Max.Sub(b.Min) }

func (b Box) Center() Vec3 { return b.Min.Add(b.Max).Scale(0.5) }

func (b Box) Contains(p Vec3) bool {
	return p.X >= b.Min.X && p.X <= b.Max.X && p.Y >= b.Min.Y && p.Y <= b.Max.Y && p.Z >= b.Min.Z && p.Z <= b.Max.Z
}

func (b Box) Union(o Box) Box {
	if b.Empty() {
		return o
	}
	if o.Empty() {
		return b
	}
	return Box{
		Vec3{math.Min(b.Min.X, o.Min.X), math.Min(b.Min.Y, o.Min.Y), math.Min(b.Min.Z, o.Min.Z)},
		Vec3{math.Max(b.Max.X, o.Max.X), math.Max(b.Max.Y, o.Max.Y), math.Max(b.Max.Z, o.Max.Z)},
	}
}

func (b Box) Intersect(o Box) Box {
	return Box{
		Vec3{math.Max(b.Min.X, o.Min.X), math.Max(b.Min.Y, o.Min.Y), math.Max(b.Min.Z, o.Min.Z)},
		Vec3{math.Min(b.Max.X, o.Max.X), math.Min(b.Max.Y, o.Max.Y), math.Min(b.Max.Z, o.Max.Z)},
	}
}

func (b Box) Grow(d float64) Box {
	g := Vec3{d, d, d}
	return Box{b.Min.Sub(g), b.Max.Add(g)}
}

func (b Box) Translate(v Vec3) Box { return Box{b.Min.Add(v), b.Max.Add(v)} }

func clamp(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

// Hash01 is a deterministic pseudo-random value in [0, 1) for integer inputs.
// Use it for weathering, ruins and variation that must not change between runs.
func Hash01(values ...int) float64 {
	h := uint64(0x9E3779B97F4A7C15)
	for _, v := range values {
		h ^= uint64(int64(v)) + 0x9E3779B97F4A7C15 + (h << 6) + (h >> 2)
		h *= 0xBF58476D1CE4E5B9
		h ^= h >> 31
	}
	return float64(h>>11) / float64(1<<53)
}
