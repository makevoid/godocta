package octa

import "math"

// Solid is a region of world space. Bounds must contain every point Contains accepts.
type Solid interface {
	Contains(p Vec3) bool
	Bounds() Box
}

// Axis names a horizontal direction.
type Axis int

const (
	AxisX Axis = iota
	AxisZ
)

// Cuboid is an inclusive axis-aligned block.
type Cuboid struct{ Min, Max Vec3 }

// Block returns the cuboid spanning two corners in any order.
func Block(a, b Vec3) Cuboid {
	box := BoxOf(a, b)
	return Cuboid{box.Min, box.Max}
}

// BlockAt returns a cuboid centered in plan on center, from y0 to y1.
func BlockAt(center Vec2, width, depth, y0, y1 float64) Cuboid {
	return Block(V(center.X-width/2, y0, center.Y-depth/2), V(center.X+width/2, y1, center.Y+depth/2))
}

func (c Cuboid) Contains(p Vec3) bool { return Box(c).Contains(p) }
func (c Cuboid) Bounds() Box          { return Box(c) }

// Cylinder is vertical, from Y0 to Y1. Inner > 0 makes it a tube.
type Cylinder struct {
	Center        Vec2
	Radius, Inner float64
	Y0, Y1        float64
}

func (c Cylinder) Contains(p Vec3) bool {
	if p.Y < c.Y0 || p.Y > c.Y1 {
		return false
	}
	d := p.Plan().Sub(c.Center).Dot(p.Plan().Sub(c.Center))
	return d <= c.Radius*c.Radius && (c.Inner <= 0 || d > c.Inner*c.Inner)
}

func (c Cylinder) Bounds() Box {
	return Box{V(c.Center.X-c.Radius, c.Y0, c.Center.Y-c.Radius), V(c.Center.X+c.Radius, c.Y1, c.Center.Y+c.Radius)}
}

// Frustum is a vertical truncated cone: radius R0 at Y0 to R1 at Y1. R1 = 0 gives a cone.
type Frustum struct {
	Center Vec2
	R0, R1 float64
	Y0, Y1 float64
}

// Cone is a vertical cone standing on base with the given height.
func Cone(center Vec2, radius, base, height float64) Frustum {
	return Frustum{center, radius, 0, base, base + height}
}

func (f Frustum) Contains(p Vec3) bool {
	if p.Y < f.Y0 || p.Y > f.Y1 {
		return false
	}
	t := (p.Y - f.Y0) / math.Max(f.Y1-f.Y0, 1e-9)
	r := f.R0 + (f.R1-f.R0)*t
	return p.Plan().Sub(f.Center).Length() <= r
}

func (f Frustum) Bounds() Box {
	r := math.Max(f.R0, f.R1)
	return Box{V(f.Center.X-r, f.Y0, f.Center.Y-r), V(f.Center.X+r, f.Y1, f.Center.Y+r)}
}

// Ellipsoid is axis-aligned with semi-axes Radii.
type Ellipsoid struct {
	Center, Radii Vec3
}

// Sphere is an ellipsoid with equal radii.
func Sphere(center Vec3, radius float64) Ellipsoid {
	return Ellipsoid{center, V(radius, radius, radius)}
}

func (e Ellipsoid) Contains(p Vec3) bool {
	d := p.Sub(e.Center)
	x, y, z := d.X/e.Radii.X, d.Y/e.Radii.Y, d.Z/e.Radii.Z
	return x*x+y*y+z*z <= 1
}

func (e Ellipsoid) Bounds() Box { return Box{e.Center.Sub(e.Radii), e.Center.Add(e.Radii)} }

// EllipticCylinder is vertical with plan semi-axes A (X) and B (Z).
type EllipticCylinder struct {
	Center Vec2
	A, B   float64
	Y0, Y1 float64
}

func (e EllipticCylinder) Contains(p Vec3) bool {
	if p.Y < e.Y0 || p.Y > e.Y1 {
		return false
	}
	x, z := (p.X-e.Center.X)/e.A, (p.Z-e.Center.Y)/e.B
	return x*x+z*z <= 1
}

func (e EllipticCylinder) Bounds() Box {
	return Box{V(e.Center.X-e.A, e.Y0, e.Center.Y-e.B), V(e.Center.X+e.A, e.Y1, e.Center.Y+e.B)}
}

// Prism extrudes a simple plan polygon from Y0 to Y1.
type Prism struct {
	Polygon []Vec2
	Y0, Y1  float64
}

func (s Prism) Contains(p Vec3) bool {
	if p.Y < s.Y0 || p.Y > s.Y1 {
		return false
	}
	return InsidePolygon(s.Polygon, p.Plan())
}

func (s Prism) Bounds() Box {
	b := EmptyBox()
	for _, v := range s.Polygon {
		b = b.Union(Box{v.At(s.Y0), v.At(s.Y1)})
	}
	return b
}

// InsidePolygon is an even-odd point-in-polygon test.
func InsidePolygon(polygon []Vec2, p Vec2) bool {
	inside := false
	for i, j := 0, len(polygon)-1; i < len(polygon); j, i = i, i+1 {
		a, b := polygon[i], polygon[j]
		if (a.Y > p.Y) != (b.Y > p.Y) && p.X < (b.X-a.X)*(p.Y-a.Y)/(b.Y-a.Y)+a.X {
			inside = !inside
		}
	}
	return inside
}

// Gable is a pitched roof over a rectangular plan, rising from Base at the lattice's
// natural 45° pitch. Ridge runs along the given axis.
type Gable struct {
	Min, Max Vec2 // Plan footprint.
	Base     float64
	Ridge    Axis
}

func (g Gable) half() float64 {
	if g.Ridge == AxisZ {
		return (g.Max.X - g.Min.X) / 2
	}
	return (g.Max.Y - g.Min.Y) / 2
}

func (g Gable) Contains(p Vec3) bool {
	if p.X < g.Min.X || p.X > g.Max.X || p.Z < g.Min.Y || p.Z > g.Max.Y || p.Y < g.Base {
		return false
	}
	across := math.Abs(p.Z - (g.Min.Y+g.Max.Y)/2)
	if g.Ridge == AxisZ {
		across = math.Abs(p.X - (g.Min.X+g.Max.X)/2)
	}
	return across <= g.half()-(p.Y-g.Base)
}

func (g Gable) Bounds() Box { return Box{g.Min.At(g.Base), g.Max.At(g.Base + g.half())} }

// HipRoof is a pyramid roof over a rectangle at 45°.
type HipRoof struct {
	Min, Max Vec2
	Base     float64
}

func (h HipRoof) Contains(p Vec3) bool {
	if p.Y < h.Base {
		return false
	}
	inset := p.Y - h.Base
	return p.X >= h.Min.X+inset && p.X <= h.Max.X-inset && p.Z >= h.Min.Y+inset && p.Z <= h.Max.Y-inset
}

func (h HipRoof) Bounds() Box {
	rise := math.Min(h.Max.X-h.Min.X, h.Max.Y-h.Min.Y) / 2
	return Box{h.Min.At(h.Base), h.Max.At(h.Base + rise)}
}

// Arch is a round-headed opening, typically carved: vertical jambs up to Spring above
// the sill, then a semicircle. Through is the axis the opening passes along.
type Arch struct {
	Sill      Vec3 // Center of the opening's bottom edge.
	HalfWidth float64
	Spring    float64
	Depth     float64
	Through   Axis
}

func (a Arch) Contains(p Vec3) bool {
	d := p.Sub(a.Sill)
	across, along := d.X, d.Z
	if a.Through == AxisX {
		across, along = d.Z, d.X
	}
	if math.Abs(along) > a.Depth/2 || math.Abs(across) > a.HalfWidth || d.Y < 0 {
		return false
	}
	rise := d.Y - a.Spring
	return rise <= 0 || across*across+rise*rise <= a.HalfWidth*a.HalfWidth
}

func (a Arch) Bounds() Box {
	span := V(a.HalfWidth, 0, a.Depth/2)
	if a.Through == AxisX {
		span = V(a.Depth/2, 0, a.HalfWidth)
	}
	return Box{a.Sill.Sub(span), a.Sill.Add(span).Add(V(0, a.Spring+a.HalfWidth, 0))}
}

// Capsule is a segment with a radius. Radius ≥ CellRadius always yields a connected line.
type Capsule struct {
	A, B   Vec3
	Radius float64
}

func (c Capsule) Contains(p Vec3) bool {
	ab := c.B.Sub(c.A)
	t := clamp(p.Sub(c.A).Dot(ab)/math.Max(ab.Dot(ab), 1e-12), 0, 1)
	return p.Sub(c.A.Add(ab.Scale(t))).Length() <= c.Radius
}

func (c Capsule) Bounds() Box { return BoxOf(c.A, c.B).Grow(c.Radius) }

// Func adapts a predicate and its bounds into a Solid.
type Func struct {
	Box Box
	F   func(Vec3) bool
}

func (f Func) Contains(p Vec3) bool { return f.Box.Contains(p) && f.F(p) }
func (f Func) Bounds() Box          { return f.Box }

// Union contains points in any of its solids.
type UnionSolid []Solid

func Union(solids ...Solid) UnionSolid { return UnionSolid(solids) }

func (u UnionSolid) Contains(p Vec3) bool {
	for _, s := range u {
		if s.Bounds().Contains(p) && s.Contains(p) {
			return true
		}
	}
	return false
}

func (u UnionSolid) Bounds() Box {
	b := EmptyBox()
	for _, s := range u {
		b = b.Union(s.Bounds())
	}
	return b
}

// DifferenceSolid contains points of Base outside every Holes solid.
type DifferenceSolid struct {
	Base  Solid
	Holes []Solid
}

func Subtract(base Solid, holes ...Solid) DifferenceSolid { return DifferenceSolid{base, holes} }

func (d DifferenceSolid) Contains(p Vec3) bool {
	if !d.Base.Contains(p) {
		return false
	}
	for _, h := range d.Holes {
		if h.Bounds().Contains(p) && h.Contains(p) {
			return false
		}
	}
	return true
}

func (d DifferenceSolid) Bounds() Box { return d.Base.Bounds() }

// IntersectionSolid contains points inside all of its solids.
type IntersectionSolid []Solid

func Intersect(solids ...Solid) IntersectionSolid { return IntersectionSolid(solids) }

func (s IntersectionSolid) Contains(p Vec3) bool {
	for _, x := range s {
		if !x.Contains(p) {
			return false
		}
	}
	return len(s) > 0
}

func (s IntersectionSolid) Bounds() Box {
	if len(s) == 0 {
		return EmptyBox()
	}
	b := s[0].Bounds()
	for _, x := range s[1:] {
		b = b.Intersect(x.Bounds())
	}
	return b
}

// Translated moves a solid by Offset.
type Translated struct {
	Solid  Solid
	Offset Vec3
}

func Translate(s Solid, offset Vec3) Translated { return Translated{s, offset} }

func (t Translated) Contains(p Vec3) bool { return t.Solid.Contains(p.Sub(t.Offset)) }
func (t Translated) Bounds() Box          { return t.Solid.Bounds().Translate(t.Offset) }

// Rotated turns a solid around a vertical axis through Pivot by Angle radians.
type Rotated struct {
	Solid Solid
	Pivot Vec2
	Angle float64
}

func RotateY(s Solid, pivot Vec2, angle float64) Rotated { return Rotated{s, pivot, angle} }

func (r Rotated) Contains(p Vec3) bool {
	q := p.Plan().Sub(r.Pivot).Rotate(-r.Angle).Add(r.Pivot)
	return r.Solid.Contains(q.At(p.Y))
}

func (r Rotated) Bounds() Box {
	inner := r.Solid.Bounds()
	b := EmptyBox()
	for _, x := range []float64{inner.Min.X, inner.Max.X} {
		for _, z := range []float64{inner.Min.Z, inner.Max.Z} {
			q := P(x, z).Sub(r.Pivot).Rotate(r.Angle).Add(r.Pivot)
			b = b.Union(Box{q.At(inner.Min.Y), q.At(inner.Max.Y)})
		}
	}
	return b
}

// Sheared slants a solid: points move by Slope (plan offset per unit height) above Base.
// Useful for raked funnels, masts and leaning towers.
type Sheared struct {
	Solid Solid
	Slope Vec2
	Base  float64
}

func (s Sheared) Contains(p Vec3) bool {
	shift := s.Slope.Scale(p.Y - s.Base)
	return s.Solid.Contains(V(p.X-shift.X, p.Y, p.Z-shift.Y))
}

func (s Sheared) Bounds() Box {
	b := s.Solid.Bounds()
	lo, hi := s.Slope.Scale(b.Min.Y-s.Base), s.Slope.Scale(b.Max.Y-s.Base)
	return b.Union(b.Translate(V(lo.X, 0, lo.Y))).Union(b.Translate(V(hi.X, 0, hi.Y)))
}
