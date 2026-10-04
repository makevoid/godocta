package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// StaveChurch is a Borgund-style Norwegian stave church on a churchyard terrace above
// a fjord. The dark tarred church stacks its roofs: the low svalgang gallery with small
// arched openings wraps the whole building, then the aisle lean-to around the raised
// nave, the steep nave and chancel roofs, the round apse with its own little tower, and
// the pagoda-like ridge turret with two skirt roofs, gablets and a spire. Carved dragon
// heads spring from the gables. Around it: a low stone churchyard wall with gravestones
// and wooden crosses, the separate bell tower, turf-roofed log cabins and stabbur
// storehouses on stilts, birch and spruce, a steep mountainside behind (-Z) with a
// waterfall and stream, and the fjord in front (+Z) with a jetty and rowing boat.
type StaveChurch struct{}

func init() { Register(StaveChurch{}) }

func (StaveChurch) Name() string  { return "stave_church" }
func (StaveChurch) Title() string { return "Stave Church" }

const (
	staveChurchX0, staveChurchX1 = -34.0, 35.2
	staveChurchZ0, staveChurchZ1 = -42.0, 29.2
	staveChurchGround            = 3.0 // Churchyard terrace.
	staveChurchFloor             = 4.0 // Church floor.
	staveChurchSea               = 1.0 // Fjord surface.
	staveChurchStep              = 0.4 // Terrain raster resolution.
)

// staveChurchRect is a plan rectangle.
type staveChurchRect struct{ x0, z0, x1, z1 float64 }

// dist is the signed Chebyshev distance from the rectangle (negative inside), and
// whether the nearest face is an X face (normal along X).
func (r staveChurchRect) dist(x, z float64) (float64, bool) {
	dx := math.Max(r.x0-x, x-r.x1)
	dz := math.Max(r.z0-z, z-r.z1)
	return math.Max(dx, dz), dx > dz
}

func (r staveChurchRect) inset(t float64) staveChurchRect {
	return staveChurchRect{r.x0 + t, r.z0 + t, r.x1 - t, r.z1 - t}
}

func (r staveChurchRect) box(y0, y1 float64) Cuboid {
	return Block(V(r.x0, y0, r.z0), V(r.x1, y1, r.z1))
}

// The church plan: raised nave core inside its aisle, chancel core inside its aisle,
// round apse at the east end. The svalgang wraps them all.
var (
	staveChurchNave        = staveChurchRect{-9, -4.5, 3, 4.5}
	staveChurchNaveAisle   = staveChurchRect{-12, -7.5, 6, 7.5}
	staveChurchChancel     = staveChurchRect{6, -3.5, 13, 3.5}
	staveChurchChancelAisl = staveChurchRect{6, -5.5, 14, 5.5}
	staveChurchApse        = P(14.5, 0)
	staveChurchTower       = P(-3, 0)
	staveChurchBell        = P(-27, 9)
)

const (
	staveChurchApseR = 3.0
	staveChurchScale = 1.15
)

// staveChurchMap is a sampled terrain heightfield with bilinear lookup.
type staveChurchMap struct {
	nx, nz int
	h      []float64
}

func (m *staveChurchMap) at(x, z float64) float64 {
	fx, fz := (x-staveChurchX0)/staveChurchStep, (z-staveChurchZ0)/staveChurchStep
	fx = math.Max(0, math.Min(fx, float64(m.nx-1)-1e-6))
	fz = math.Max(0, math.Min(fz, float64(m.nz-1)-1e-6))
	i, j := int(fx), int(fz)
	tx, tz := fx-float64(i), fz-float64(j)
	k := j*m.nx + i
	a := m.h[k] + (m.h[k+1]-m.h[k])*tx
	b := m.h[k+m.nx] + (m.h[k+m.nx+1]-m.h[k+m.nx])*tx
	return a + (b-a)*tz
}

// low is the lowest terrain height around a point, used to close slopes into shells.
func (m *staveChurchMap) low(x, z float64) float64 {
	lo := m.at(x, z)
	for k := 0; k < 8; k++ {
		a := float64(k) * math.Pi / 4
		lo = math.Min(lo, m.at(x+1.0*math.Cos(a), z+1.0*math.Sin(a)))
	}
	return lo
}

func (StaveChurch) inSite(x, z float64) bool {
	return x >= staveChurchX0 && x <= staveChurchX1 && z >= staveChurchZ0 && z <= staveChurchZ1
}

// noise is smooth value noise in [0, 1).
func (StaveChurch) noise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

func (s StaveChurch) fbm(x, z float64, seed int) float64 {
	return (s.noise(x, z, seed)*4 + s.noise(x*2.1, z*2.1, seed+1)*2 + s.noise(x*4.3, z*4.3, seed+2)) / 7
}

func (StaveChurch) clamp(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

// shore is the fjord's waterline (z at x).
func (StaveChurch) shore(x float64) float64 {
	return 21 + 1.8*math.Sin(x/7+0.5) + 1.0*math.Sin(x/3.3+1.2)
}

// stream is the brook's course (x at z), from the waterfall to the fjord.
func (s StaveChurch) stream(z float64) float64 {
	t := s.clamp((-z-11)/9, 0, 1)
	return 28 - 6*t*t*(3-2*t) + 1.2*math.Sin(z/5)
}

// yard is the signed distance from the churchyard wall's outer face (rounded rectangle).
func (StaveChurch) yard(x, z float64) float64 {
	const cx, cz, hx, hz, r = 1.0, 0.0, 23.5, 16.0, 4.0
	qx, qz := math.Abs(x-cx)-(hx-r), math.Abs(z-cz)-(hz-r)
	return math.Hypot(math.Max(qx, 0), math.Max(qz, 0)) + math.Min(math.Max(qx, qz), 0) - r
}

// height is the terrain: the churchyard terrace, the mountainside behind with its
// waterfall gully, the brook, and the beach falling into the fjord.
func (s StaveChurch) height(x, z float64) float64 {
	h := staveChurchGround + 1.4*(s.fbm(x/11, z/11, 1)-0.5)
	w := s.clamp((s.yard(x, z)-1)/7, 0, 1)
	h = staveChurchGround + (h-staveChurchGround)*w
	// Mountainside.
	t := s.clamp((-z-18)/13, 0, 1)
	env := s.clamp(math.Min(staveChurchX1-x, x-staveChurchX0)/11+0.15*(s.noise(z/6, 2.5, 5)-0.5), 0, 1)
	peak := (32 + 8*(s.fbm(x/13, 4.2, 3)-0.5)) * env * env * (3 - 2*env)
	ledges := 1.6 * math.Floor(t*7+2*s.noise(x/7, z/7, 4)) / 7
	mh := staveChurchGround + peak*(0.75*math.Pow(t, 1.3)+0.25*ledges) + 4*(s.fbm(x/5, z/5, 2)-0.5)*t
	mh -= 5 * math.Exp(-math.Pow((x-s.stream(z))/3.2, 2)) * t
	h = math.Max(h, mh)
	// Beach into the fjord.
	if z > s.shore(x)-4 {
		h = math.Min(h, 3-(z-(s.shore(x)-4))*0.5)
	}
	// Site rims.
	h = math.Min(h, 0.4+2.4*math.Min(staveChurchX1-x, x-staveChurchX0))
	h = math.Min(h, 0.4+3.3*(z-staveChurchZ0))
	// Brook channel.
	if z > -20 {
		h -= 0.9 * s.clamp(1.9-math.Abs(x-s.stream(z)), 0, 1)
	}
	return h
}

func (s StaveChurch) Build(g *Grid) Order {
	m := &staveChurchMap{}
	m.nx = int((staveChurchX1-staveChurchX0)/staveChurchStep) + 2
	m.nz = int((staveChurchZ1-staveChurchZ0)/staveChurchStep) + 2
	m.h = make([]float64, m.nx*m.nz)
	for j := 0; j < m.nz; j++ {
		for i := 0; i < m.nx; i++ {
			m.h[j*m.nx+i] = s.height(staveChurchX0+float64(i)*staveChurchStep, staveChurchZ0+float64(j)*staveChurchStep)
		}
	}
	s.terrain(g, m)
	s.churchyard(g, m)
	// The church is built on a finer lattice and pasted cell for cell, which scales it
	// up by staveChurchScale about the origin; the shift settles it onto the terrace.
	cg := NewGridWithSpan(g.Span() / staveChurchScale)
	s.church(cg)
	shift := Cell{X: -2, Y: -2, Z: 0}
	for _, c := range cg.Cells() {
		col, _ := cg.ColorOf(c)
		n := c.Add(shift)
		g.Remove(n)
		g.With(col, func() { g.Add(n) })
	}
	s.bellTower(g, m)
	s.stabbur(g, m, P(-28, -8), 5, 4.2, AxisX)
	s.stabbur(g, m, P(31.5, -6), 4.2, 5, AxisZ)
	s.cabin(g, m, P(-29.5, 1.5), 6, 7.5, AxisZ)
	s.cabin(g, m, P(31.2, 6.5), 5.0, 6.5, AxisZ)
	s.jetty(g, m)
	s.naust(g)
	s.trees(g, m)
	return Outward(P(2, 0))
}

// terrain fills the heightfield as a shell and colors meadow, forest, rock, beach,
// waterfall, brook and fjord.
func (s StaveChurch) terrain(g *Grid, m *staveChurchMap) {
	box := Box{V(staveChurchX0, 0, staveChurchZ0), V(staveChurchX1, 50, staveChurchZ1)}
	g.Use(Grass).Fill(Func{box, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		if h <= staveChurchSea-0.2 {
			return false
		}
		return p.Y <= h && p.Y >= math.Max(0, math.Min(h, m.low(p.X, p.Z))-1.0)
	}})
	steep := func(p Vec3) float64 { return m.at(p.X, p.Z) - m.low(p.X, p.Z) }
	// Rock on steep faces, banded.
	g.Paint(DarkStone, Func{box, func(p Vec3) bool { return steep(p) > 3.1 }})
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		if steep(p) <= 3.1 {
			return false
		}
		band := int(math.Floor(p.Y/1.3 + 2*s.noise(p.X/8, p.Z/8, 11)))
		return Hash01(band, int(math.Floor(p.X/4)), 12) < 0.3
	}})
	// Forest floor and moss on the mountainside.
	g.Paint(Foliage, Func{box, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		return p.Z < -14 && p.Y >= h-1.2 && steep(p) < 3.0 && s.fbm(p.X/8, p.Z/8, 13) > 0.45
	}})
	// Shingle beach.
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		return h < 1.9 && p.Z > 10
	}})
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		return h < 1.9 && p.Z > 10 && Hash01(int(math.Floor(p.X*1.3)), int(math.Floor(p.Z*1.3)), 14) < 0.3
	}})
	// Paths: west gate to the bell tower and cabins, south gate to the jetty.
	path := func(a, b Vec2, w float64) Solid {
		return Func{Box{V(math.Min(a.X, b.X)-w, 0, math.Min(a.Y, b.Y)-w), V(math.Max(a.X, b.X)+w, 8, math.Max(a.Y, b.Y)+w)},
			func(p Vec3) bool {
				ab := b.Sub(a)
				t := s.clamp(p.Plan().Sub(a).Dot(ab)/ab.Dot(ab), 0, 1)
				return p.Plan().Sub(a.Add(ab.Scale(t))).Length() <= w && p.Y >= m.at(p.X, p.Z)-0.9
			}}
	}
	g.Paint(Sand, path(P(-22, 0), P(-26, 1), 1.2), path(P(-26, 1), P(-28, 5), 1.0), path(P(-26, 1), P(-28, -5), 1.0),
		path(P(-4.2, 15), P(1, 18), 1.2), path(P(1, 18), P(7.5, 19.8), 1.2))
	// Waterfall: a raised strip of white water down the gully, then the brook.
	fall := Func{Box{V(16, 0, staveChurchZ0), V(33, 50, -14)}, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		return math.Abs(p.X-s.stream(p.Z)) < 1.5 && p.Y >= h-0.8 && p.Y <= h+0.6 && h > 4
	}}
	g.Use(White).Fill(fall)
	g.Paint(Water, Intersect(fall, Func{box, func(p Vec3) bool {
		return steep(p) < 0.9 || Hash01(int(math.Floor(p.X*1.6)), int(math.Floor(p.Z/6)), 15) < 0.3
	}}))
	g.Paint(Water, Func{Box{V(17, 0, -24), V(34, 12, staveChurchZ1)}, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		return math.Abs(p.X-s.stream(p.Z)) < 1.3 && p.Y >= h-0.9
	}})
	// Plunge pool foam.
	g.Paint(White, Func{Box{V(17, 0, -22), V(30, 10, -15)}, func(p Vec3) bool {
		d := math.Hypot(p.X-s.stream(p.Z), p.Z+18.5)
		return d < 2.2 && p.Y >= m.at(p.X, p.Z)-0.9 && Hash01(int(math.Floor(p.X*2)), int(math.Floor(p.Z*2)), 16) < 0.5
	}})
	// The fjord: one water layer.
	g.Use(Water).Fill(Func{Box{V(staveChurchX0, 0, 5), V(staveChurchX1, staveChurchSea, staveChurchZ1)}, func(p Vec3) bool {
		return m.at(p.X, p.Z) < staveChurchSea
	}})
}

// churchyard is the low dry-stone wall with two gates, gravestones and wooden crosses.
func (s StaveChurch) churchyard(g *Grid, m *staveChurchMap) {
	box := Box{V(-23, 0, -16), V(25, 7, 16)}
	gate := func(p Vec3) bool {
		return (p.X < -18 && math.Abs(p.Z) < 1.8) || (p.Z > 11 && math.Abs(p.X+4.2) < 1.8)
	}
	wall := Func{box, func(p Vec3) bool {
		d := s.yard(p.X, p.Z)
		h := m.at(p.X, p.Z)
		return d <= 0 && d >= -1.1 && p.Y >= h-0.8 && p.Y <= h+1.2 && !gate(p)
	}}
	g.Use(Stone).Fill(wall)
	g.Paint(DarkStone, Intersect(wall, Func{box, func(p Vec3) bool {
		row := int(math.Floor(p.Y / 0.7))
		return Hash01(row, int(math.Floor((p.X+p.Z+float64(row)*0.6)/1.3)), 31) < 0.55
	}}))
	// Gate posts.
	g.Use(Wood)
	for _, at := range []Vec2{P(-22, -2.4), P(-22, 2.4), P(-6.4, 15.5), P(-2.0, 15.5)} {
		h := m.at(at.X, at.Y)
		g.Fill(Cylinder{Center: at, Radius: 0.6, Y0: h - 0.5, Y1: h + 2.4})
	}
	// Graves in rows round the church, in the strips inside the wall.
	grave := func(at Vec2, k int, facing Axis) {
		h := m.at(at.X, at.Y)
		switch r := Hash01(k, 33); {
		case r < 0.4: // Gravestone with a rounded head.
			c := Stone
			if Hash01(k, 34) < 0.5 {
				c = DarkStone
			}
			g.With(c, func() {
				if facing == AxisZ {
					g.Fill(Block(V(at.X-0.7, h-0.3, at.Y-0.4), V(at.X+0.7, h+1.2, at.Y+0.4)))
					g.FillVolume(Ellipsoid{at.At(h + 1.2), V(0.75, 0.55, 0.45)})
				} else {
					g.Fill(Block(V(at.X-0.4, h-0.3, at.Y-0.7), V(at.X+0.4, h+1.2, at.Y+0.7)))
					g.FillVolume(Ellipsoid{at.At(h + 1.2), V(0.45, 0.55, 0.75)})
				}
			})
		default: // Wooden cross.
			c := Wood
			if Hash01(k, 35) < 0.25 {
				c = White
			}
			g.With(c, func() {
				g.Fill(Block(V(at.X-0.35, h-0.3, at.Y-0.35), V(at.X+0.35, h+2.4, at.Y+0.35)))
				if facing == AxisZ {
					g.Fill(Block(V(at.X-1.0, h+1.5, at.Y-0.35), V(at.X+1.0, h+2.0, at.Y+0.35)))
				} else {
					g.Fill(Block(V(at.X-0.35, h+1.5, at.Y-1.0), V(at.X+0.35, h+2.0, at.Y+1.0)))
				}
			})
		}
	}
	k := 0
	for x := -19.5; x <= 22; x += 2.6 {
		k++
		if math.Abs(x+4.2) > 3 && Hash01(k, 36) < 0.6 {
			grave(P(x+0.4*(Hash01(k, 37)-0.5), 13.6), k, AxisZ)
		}
		k++
		if Hash01(k, 36) < 0.55 {
			grave(P(x+0.4*(Hash01(k, 37)-0.5), -13.6), k, AxisZ)
		}
	}
	for z := -11.0; z <= 11; z += 2.6 {
		k++
		if math.Abs(z) > 3 && Hash01(k, 36) < 0.6 {
			grave(P(-20.4, z), k, AxisX)
		}
		k++
		if math.Abs(z) > 4.5 && Hash01(k, 36) < 0.6 {
			grave(P(22.4, z), k, AxisX)
		}
	}
}

// shingles paints a roof solid with Wood flecks on its Charcoal shingles.
func (StaveChurch) shingles(g *Grid, roof Solid) {
	g.With(Charcoal, func() { g.Fill(roof) })
	g.Paint(Wood, Intersect(roof, Func{Box{V(-100, -10, -100), V(100, 100, 100)}, func(p Vec3) bool {
		return Hash01(int(math.Floor(p.X*2)), int(math.Floor(p.Y*2)), int(math.Floor(p.Z*2)), 41) < 0.07
	}}))
}

// gable is a steep pitched roof shell over a plan rectangle: eaves at eave, rising k
// per unit across, th thick vertically. The ridge runs along the given axis.
func (StaveChurch) gable(min, max Vec2, eave, k, th float64, ridge Axis) Func {
	half, c := (max.Y-min.Y)/2, (max.Y+min.Y)/2
	if ridge == AxisZ {
		half, c = (max.X-min.X)/2, (max.X+min.X)/2
	}
	return Func{Box{min.At(eave - th), max.At(eave + half*k)}, func(p Vec3) bool {
		a := math.Abs(p.Z - c)
		if ridge == AxisZ {
			a = math.Abs(p.X - c)
		}
		y := eave + (half-a)*k
		return p.Y <= y && p.Y >= y-th
	}}
}

// gableWall is the solid triangle under a gable roof (same half, center, eave and
// pitch), within a thin slab given by min/max, from base up.
func (StaveChurch) gableWall(min, max Vec2, base, eave, half, c, k float64, ridge Axis) Func {
	return Func{Box{min.At(base), max.At(eave + half*k)}, func(p Vec3) bool {
		a := math.Abs(p.Z - c)
		if ridge == AxisZ {
			a = math.Abs(p.X - c)
		}
		return p.Y <= eave+(half-a)*k-0.4
	}}
}

// leanTo is a roof shell sloping down from top at d = 0 by k per unit of the distance
// d, out to dmax; dist < 0 is the core it leans against. With a point core it is a
// pyramid.
func (StaveChurch) leanTo(box Box, dist func(x, z float64) float64, top, k, dmax, th float64) Func {
	return Func{box, func(p Vec3) bool {
		d := dist(p.X, p.Z)
		if d < -0.3 || d > dmax {
			return false
		}
		y := top - math.Max(d, 0)*k
		return p.Y <= y && p.Y >= y-th
	}}
}

// shell fills a hollow rectangular wall of thickness t with vertical battens.
func (StaveChurch) shell(g *Grid, r staveChurchRect, y0, y1, t float64) {
	solid := Subtract(r.box(y0, y1), r.inset(t).box(y0-1, y1+1))
	g.With(Wood, func() { g.Fill(solid) })
	g.Paint(Charcoal, Intersect(solid, Func{solid.Bounds(), func(p Vec3) bool {
		fx := math.Min(p.X-r.x0, r.x1-p.X)
		fz := math.Min(p.Z-r.z0, r.z1-p.Z)
		if math.Min(fx, fz) < 0.6 && math.Max(fx, fz) < 0.9 {
			return true // Corner post.
		}
		return false
	}}))
}

// dragon is a carved dragon head springing from at, facing out (plan unit vector).
func (StaveChurch) dragon(g *Grid, at Vec3, out Vec2, size float64) {
	o := V(out.X, 0, out.Y)
	up := V(0, 1, 0)
	p1 := at.Add(o.Scale(1.1 * size)).Add(up.Scale(1.5 * size))
	g.With(Wood, func() { g.Fill(Capsule{at, p1, 0.6}) })
	g.With(Gold, func() {
		g.Fill(Capsule{p1, p1.Add(o.Scale(2.0 * size)).Add(up.Scale(0.3 * size)), 0.7})
		g.Fill(Capsule{p1.Add(o.Scale(0.5 * size)).Sub(up.Scale(0.6 * size)), p1.Add(o.Scale(1.8 * size)).Sub(up.Scale(0.7 * size)), 0.5})
		g.Fill(Capsule{p1, p1.Sub(o.Scale(0.9 * size)).Add(up.Scale(1.1 * size)), 0.45})
	})
}

// cross is a wooden gable cross.
func (StaveChurch) cross(g *Grid, at Vec3, across Axis, c Color) {
	g.With(c, func() {
		g.Fill(Capsule{at, at.Add(V(0, 2.6, 0)), 0.5})
		bar := V(0.9, 0, 0)
		if across == AxisZ {
			bar = V(0, 0, 0.9)
		}
		g.Fill(Capsule{at.Add(V(0, 1.8, 0)).Sub(bar), at.Add(V(0, 1.8, 0)).Add(bar), 0.45})
	})
}

// svDist is the distance from the church's outer aisle walls (union of the nave aisle,
// chancel aisle and apse), with an along-wall coordinate and a corner flag.
func (StaveChurch) svDist(x, z float64) (d, u float64, corner bool) {
	d = math.Inf(1)
	for _, r := range []staveChurchRect{staveChurchNaveAisle, staveChurchChancelAisl} {
		rd, xFace := r.dist(x, z)
		if rd < d {
			d = rd
			u = x
			if xFace {
				u = z
			}
			dx := math.Max(r.x0-x, x-r.x1)
			dz := math.Max(r.z0-z, z-r.z1)
			corner = math.Min(dx, dz) > -0.9
		}
	}
	a := P(x, z).Sub(staveChurchApse)
	if ad := a.Length() - staveChurchApseR; ad < d {
		d, u, corner = ad, a.Angle()*(staveChurchApseR+2), false
	}
	return d, u, corner
}

// church builds the stave church.
func (s StaveChurch) church(g *Grid) {
	const F = staveChurchFloor
	big := Box{V(-20, 0, -14), V(22, 60, 14)}
	sv := func(x, z float64) float64 { d, _, _ := s.svDist(x, z); return d }
	// Stone sill and plank floor.
	g.Use(DarkStone).Fill(Func{big, func(p Vec3) bool { return p.Y >= 2.0 && p.Y <= F-0.6 && sv(p.X, p.Z) <= 2.9 }})
	g.Paint(Stone, Func{big, func(p Vec3) bool {
		return p.Y <= F-0.6 && sv(p.X, p.Z) <= 2.9 && Hash01(int(math.Floor(p.Y/0.7)), int(math.Floor((p.X+p.Z)/1.4)), 42) < 0.4
	}})
	g.Use(Wood).Fill(Func{big, func(p Vec3) bool { return p.Y > F-0.8 && p.Y <= F && sv(p.X, p.Z) <= 2.6 }})

	// Svalgang outer wall: a solid parapet, then an arcade of small round arches.
	svWall := Func{big, func(p Vec3) bool {
		d := sv(p.X, p.Z)
		return p.Y > F && p.Y <= F+3.9 && d >= 1.4 && d <= 2.5
	}}
	g.Use(Wood).Fill(svWall)
	g.Paint(Charcoal, Func{big, func(p Vec3) bool {
		d, u, corner := s.svDist(p.X, p.Z)
		return d >= 1.4 && d <= 2.5 && p.Y > F && (corner || (p.Y > F+1.0 && p.Y < F+1.6 && math.Mod(u+100.3, 1.1) < 0.4))
	}})
	g.Carve(Func{big, func(p Vec3) bool {
		d, u, corner := s.svDist(p.X, p.Z)
		if d < 1.3 || d > 2.6 || corner || p.Y < F+1.6 {
			return false
		}
		if (p.X < -13 && math.Abs(p.Z) < 2.6) || (p.Z > 8 && math.Abs(p.X+3) < 2.6) {
			return false // Porches.
		}
		a := math.Mod(u+100.3, 2.2) - 1.1 - 0.25
		y := p.Y - (F + 2.9)
		return math.Abs(a) < 0.7 && (y < 0 || a*a+y*y < 0.49)
	}})

	// Aisle walls, core walls, apse.
	s.shell(g, staveChurchNaveAisle, F, F+10.5, 1.2)
	s.shell(g, staveChurchChancelAisl, F, F+8.2, 1.2)
	s.shell(g, staveChurchNave, F, F+18.4, 1.2)
	s.shell(g, staveChurchChancel, F, F+12.4, 1.2)
	g.Use(Wood).Fill(Cylinder{Center: staveChurchApse, Radius: staveChurchApseR, Inner: staveChurchApseR - 1.2, Y0: F, Y1: F + 7.0})
	// Gable walls.
	nv, ch := staveChurchNave, staveChurchChancel
	g.Use(Wood).Fill(
		s.gableWall(P(nv.x0, nv.z0), P(nv.x0+1.2, nv.z1), F+18, F+18.1, 5.4, 0, 1.85, AxisX),
		s.gableWall(P(nv.x1-1.2, nv.z0), P(nv.x1, nv.z1), F+18, F+18.1, 5.4, 0, 1.85, AxisX),
		s.gableWall(P(ch.x1-1.2, ch.z0), P(ch.x1, ch.z1), F+12, F+12.1, 4.3, 0, 1.85, AxisX),
	)
	// Clerestory: small round windows high on the nave walls.
	for _, x := range []float64{-7.5, -4.5, -1.5, 1.5} {
		g.Carve(Capsule{V(x, F+16.0, nv.z0-1), V(x, F+16.0, nv.z0+1.6), 0.75}, Capsule{V(x, F+16.0, nv.z1+1), V(x, F+16.0, nv.z1-1.6), 0.75})
	}
	for _, z := range []float64{-2, 2} {
		g.Carve(Capsule{V(nv.x0-1, F+16.0, z), V(nv.x0+1.6, F+16.0, z), 0.75})
	}

	// Tower stages.
	tw := staveChurchTower
	tF := F + 1.0
	sq := func(h float64) staveChurchRect { return staveChurchRect{tw.X - h, tw.Y - h, tw.X + h, tw.Y + h} }
	cheb := func(x, z float64) float64 { return math.Max(math.Abs(x-tw.X), math.Abs(z-tw.Y)) }
	s.shell(g, sq(2.6), tF+20, tF+30, 1.1)
	s.shell(g, sq(1.8), tF+30.5, tF+35, 1.0)
	s.shell(g, sq(1.1), tF+35.5, tF+37.6, 1.1)
	for _, sign := range []float64{-1, 1} {
		g.Carve(Arch{Sill: V(tw.X, tF+32.2, tw.Y+sign*1.3), HalfWidth: 0.65, Spring: 1.2, Depth: 1.8, Through: AxisZ},
			Arch{Sill: V(tw.X+sign*1.3, tF+32.2, tw.Y), HalfWidth: 0.65, Spring: 1.2, Depth: 1.8, Through: AxisX})
		g.Carve(Arch{Sill: V(tw.X, tF+27.4, tw.Y+sign*2.1), HalfWidth: 0.5, Spring: 0.8, Depth: 1.4, Through: AxisZ},
			Arch{Sill: V(tw.X+sign*2.1, tF+27.4, tw.Y), HalfWidth: 0.5, Spring: 0.8, Depth: 1.4, Through: AxisX})
	}

	// Dark interiors so openings read as shadow.
	dark := Charcoal
	g.Paint(dark, Func{big, func(p Vec3) bool {
		if p.Y <= F {
			return false
		}
		in := func(r staveChurchRect, top float64) bool {
			d, _ := r.dist(p.X, p.Z)
			return d < -0.6 && p.Y < top
		}
		naveRoof := F + 18.1 + (5.4-math.Abs(p.Z))*1.85 - 1.0
		chRoof := F + 12.1 + (4.3-math.Abs(p.Z))*1.85 - 1.0
		if in(staveChurchNave, naveRoof) && !(cheb(p.X, p.Z) < 2.6 && p.Y > tF+24) {
			return true
		}
		if in(staveChurchNaveAisle, F+12.5) || in(staveChurchChancelAisl, F+9.2) || in(staveChurchChancel, chRoof) {
			return true
		}
		if P(p.X, p.Z).Sub(staveChurchApse).Length() < staveChurchApseR-0.6 && p.Y < F+9 {
			return true
		}
		d := sv(p.X, p.Z)
		if d > -0.6 && d < 1.9 && p.Y < F+3.9+(2.5-math.Max(d, 0))*1.0-0.8 {
			return true
		}
		c := cheb(p.X, p.Z)
		return (c < 1.2 && p.Y > tF+30.5 && p.Y < tF+35) || (c < 2.0 && p.Y > tF+25 && p.Y < tF+30)
	}})

	// Roofs, bottom to top.
	s.shingles(g, Intersect(s.leanTo(big, sv, F+3.9+2.5*1.0, 1.0, 3.1, 1.5), Block(V(-30, F+3.75, -20), V(30, F+12, 20))))
	s.shingles(g, s.leanTo(Box{V(-13, F+8, -9), V(7, F+15, 9)}, func(x, z float64) float64 {
		d, _ := staveChurchNave.dist(x, z)
		return d
	}, F+10.5+3*1.15, 1.15, 3.6, 1.6))
	s.shingles(g, s.leanTo(Box{V(6, F+6, -7), V(15, F+12, 7)}, func(x, z float64) float64 {
		d, _ := staveChurchChancel.dist(x, z)
		return d
	}, F+8.2+2*1.25, 1.25, 2.7, 1.6))
	s.shingles(g, s.gable(P(3, -4.3), P(13.8, 4.3), F+12.1, 1.85, 2.0, AxisX))
	s.shingles(g, s.gable(P(-9.9, -5.4), P(3.9, 5.4), F+18.1, 1.85, 2.0, AxisX))
	// Apse: conical roof, little round tower and cap.
	g.Use(Charcoal)
	g.FillVolume(Subtract(Frustum{staveChurchApse, 3.8, 1.6, F + 6.6, F + 9.4}, Frustum{staveChurchApse, 2.6, 0.6, F + 6.0, F + 8.8}))
	g.Use(Wood).Fill(Cylinder{Center: staveChurchApse, Radius: 1.6, Y0: F + 8.6, Y1: F + 11.6})
	s.shingles(g, Subtract(Frustum{staveChurchApse, 2.2, 0, F + 11.4, F + 16.2}, Frustum{staveChurchApse, 1.2, 0, F + 11.0, F + 14.5}))
	g.Use(Gold).Fill(Capsule{staveChurchApse.At(F + 14.5), staveChurchApse.At(F + 17.2), 0.5})
	g.Paint(Charcoal, Cylinder{Center: staveChurchApse, Radius: 1.7, Y0: F + 10.0, Y1: F + 10.6})
	// Tower skirts with gablets, then the spire.
	towerBox := func(h, y0, y1 float64) Box { return Box{V(tw.X-h, y0, tw.Y-h), V(tw.X+h, y1, tw.Y+h)} }
	s.shingles(g, s.leanTo(towerBox(3.5, tF+27, tF+33), func(x, z float64) float64 { return cheb(x, z) - 1.8 }, tF+29.6+1.6*1.3, 1.3, 1.6, 1.4))
	for _, dir := range []Vec2{P(1, 0), P(-1, 0), P(0, 1), P(0, -1)} {
		c := tw.Add(dir.Scale(2.8))
		var min, max Vec2
		gr := AxisX
		if dir.Y != 0 {
			gr = AxisZ
		}
		if gr == AxisX {
			min, max = P(c.X-0.9, c.Y-1.4), P(c.X+0.9, c.Y+1.4)
		} else {
			min, max = P(c.X-1.4, c.Y-0.9), P(c.X+1.4, c.Y+0.9)
		}
		s.shingles(g, s.gable(min, max, tF+29.8, 1.5, 1.4, gr))
		tip := c.Add(dir.Scale(0.9))
		apex := tF + 29.8 + 1.4*1.5
		wallMin, wallMax := tip.Sub(dir.Scale(0.6)), tip
		if gr == AxisX {
			wallMin, wallMax = P(math.Min(wallMin.X, wallMax.X), c.Y-1.4), P(math.Max(wallMin.X, wallMax.X), c.Y+1.4)
			g.Use(Wood).Fill(s.gableWall(wallMin, wallMax, tF+29, tF+29.8, 1.4, c.Y, 1.5, AxisX))
		} else {
			wallMin, wallMax = P(c.X-1.4, math.Min(wallMin.Y, wallMax.Y)), P(c.X+1.4, math.Max(wallMin.Y, wallMax.Y))
			g.Use(Wood).Fill(s.gableWall(wallMin, wallMax, tF+29, tF+29.8, 1.4, c.X, 1.5, AxisZ))
		}
		s.dragon(g, tip.At(apex-0.3), dir, 0.7)
	}
	s.shingles(g, s.leanTo(towerBox(2.6, tF+33, tF+37), func(x, z float64) float64 { return cheb(x, z) - 1.1 }, tF+34.6+1.5*1.3, 1.3, 1.5, 1.4))
	g.Use(Charcoal).FillVolume(Frustum{tw, 1.6, 0, tF + 37.3, tF + 46.5})
	g.Paint(Wood, Func{Box{V(tw.X-2, tF+37, tw.Y-2), V(tw.X+2, tF+47, tw.Y+2)}, func(p Vec3) bool {
		return Hash01(int(math.Floor(p.Y/0.8)), int(math.Floor((p.X+p.Z)/0.9)), 43) < 0.2
	}})
	g.Use(Charcoal).Fill(Capsule{tw.At(tF + 43), tw.At(tF + 46.5), 0.5})
	g.Use(Gold).Fill(Sphere(tw.At(tF+46.8), 0.7))
	s.cross(g, tw.At(tF+47), AxisX, Gold)

	// Gable crosses and dragon heads.
	s.cross(g, V(nv.x0-0.6, F+28.0, 0), AxisZ, Wood)
	s.cross(g, V(nv.x1+0.6, F+28.0, 0), AxisZ, Wood)
	s.cross(g, V(ch.x1+0.5, F+20.0, 0), AxisZ, Wood)
	for _, z := range []float64{-1, 1} {
		s.dragon(g, V(nv.x0-0.6, F+27.0, z*0.6), P(-0.8, z*0.6), 1.2)
		s.dragon(g, V(nv.x1+0.6, F+27.0, z*0.6), P(0.8, z*0.6), 1.2)
	}
	s.dragon(g, V(ch.x1+0.5, F+19.2, 0), P(1, 0), 0.9)

	// Porches with carved portals: west (ridge along X) and south (ridge along Z).
	g.Carve(Arch{Sill: V(-13.95, F, 0), HalfWidth: 1.2, Spring: 2.0, Depth: 1.8, Through: AxisX},
		Arch{Sill: V(-3, F, 9.45), HalfWidth: 1.2, Spring: 2.0, Depth: 1.8, Through: AxisZ})
	g.Carve(Arch{Sill: V(-11.4, F, 0), HalfWidth: 1.3, Spring: 2.8, Depth: 1.6, Through: AxisX},
		Arch{Sill: V(-3, F, 6.9), HalfWidth: 1.3, Spring: 2.8, Depth: 1.6, Through: AxisZ})
	g.Paint(Gold,
		Subtract(Arch{Sill: V(-12.0, F, 0), HalfWidth: 2.0, Spring: 2.8, Depth: 0.9, Through: AxisX}, Arch{Sill: V(-12.0, F, 0), HalfWidth: 1.3, Spring: 2.8, Depth: 2, Through: AxisX}),
		Subtract(Arch{Sill: V(-3, F, 7.5), HalfWidth: 2.0, Spring: 2.8, Depth: 0.9, Through: AxisZ}, Arch{Sill: V(-3, F, 7.5), HalfWidth: 1.3, Spring: 2.8, Depth: 2, Through: AxisZ}))
	g.Paint(Charcoal, Block(V(-11.0, F, -1.3), V(-10.4, F+4.2, 1.3)), Block(V(-4.3, F, 6.1), V(-1.7, F+4.2, 6.6)))
	// Porch roofs, front gables and posts.
	s.shingles(g, s.gable(P(-16.0, -2.4), P(-13.0, 2.4), F+4.0, 1.4, 1.4, AxisX))
	g.Use(Wood).Fill(s.gableWall(P(-15.8, -2.4), P(-15.2, 2.4), F+3.8, F+4.0, 2.4, 0, 1.4, AxisX))
	s.shingles(g, s.gable(P(-5.4, 9.0), P(-0.6, 12.0), F+4.0, 1.4, 1.4, AxisZ))
	g.Use(Wood).Fill(s.gableWall(P(-5.4, 11.2), P(-0.6, 11.8), F+3.8, F+4.0, 2.4, -3, 1.4, AxisZ))
	g.Use(Charcoal)
	for _, at := range []Vec2{P(-15.5, -2.0), P(-15.5, 2.0), P(-5.0, 11.5), P(-1.0, 11.5)} {
		g.Fill(Cylinder{Center: at, Radius: 0.5, Y0: F - 1, Y1: F + 4.2})
	}
	g.Carve(Arch{Sill: V(-15.5, F, 0), HalfWidth: 1.0, Spring: 1.4, Depth: 1.2, Through: AxisX},
		Arch{Sill: V(-3, F, 11.5), HalfWidth: 1.0, Spring: 1.4, Depth: 1.2, Through: AxisZ})
	g.Paint(Gold, Subtract(Arch{Sill: V(-15.5, F, 0), HalfWidth: 1.6, Spring: 1.4, Depth: 1.2, Through: AxisX}, Arch{Sill: V(-15.5, F, 0), HalfWidth: 1.0, Spring: 1.4, Depth: 2, Through: AxisX}),
		Subtract(Arch{Sill: V(-3, F, 11.5), HalfWidth: 1.6, Spring: 1.4, Depth: 1.2, Through: AxisZ}, Arch{Sill: V(-3, F, 11.5), HalfWidth: 1.0, Spring: 1.4, Depth: 2, Through: AxisZ}))
	s.cross(g, V(-16.0, F+7.0, 0), AxisZ, Wood)
	s.cross(g, V(-3, F+7.0, 12.0), AxisX, Wood)
}

// bellTower is the free-standing klokketårn: flared lower walls, an open belfry and a
// steep pyramid roof.
func (s StaveChurch) bellTower(g *Grid, m *staveChurchMap) {
	c := staveChurchBell
	h := m.at(c.X, c.Y)
	cheb := func(x, z float64) float64 { return math.Max(math.Abs(x-c.X), math.Abs(z-c.Y)) }
	box := Box{V(c.X-5, h-2, c.Y-5), V(c.X+5, h+22, c.Y+5)}
	g.Use(DarkStone).Fill(Block(V(c.X-3.4, h-1.5, c.Y-3.4), V(c.X+3.4, h+0.6, c.Y+3.4)))
	half := func(y float64) float64 { return 3.2 - 0.9*s.clamp((y-h)/7, 0, 1) }
	walls := Func{box, func(p Vec3) bool {
		d := cheb(p.X, p.Z)
		return p.Y > h+0.6 && p.Y <= h+10.4 && d <= half(p.Y) && d >= half(p.Y)-1.1
	}}
	g.Use(Wood).Fill(walls)
	g.Paint(Charcoal, Intersect(walls, Func{box, func(p Vec3) bool {
		u := p.X
		if math.Abs(p.X-c.X) > math.Abs(p.Z-c.Y) {
			u = p.Z
		}
		return math.Mod(u+100.3, 2.0) < 0.5 || (p.Y > h+6.8 && p.Y < h+7.5)
	}}))
	for _, sign := range []float64{-1, 1} {
		g.Carve(Arch{Sill: V(c.X, h+7.6, c.Y+sign*2.0), HalfWidth: 0.9, Spring: 1.0, Depth: 1.6, Through: AxisZ},
			Arch{Sill: V(c.X+sign*2.0, h+7.6, c.Y), HalfWidth: 0.9, Spring: 1.0, Depth: 1.6, Through: AxisX})
	}
	g.Carve(Arch{Sill: V(c.X, h+0.6, c.Y+3.0), HalfWidth: 0.8, Spring: 1.6, Depth: 1.6, Through: AxisZ})
	g.Paint(Charcoal, Func{box, func(p Vec3) bool { return cheb(p.X, p.Z) < half(p.Y)-0.6 }})
	s.shingles(g, s.leanTo(box, cheb, h+10.0+3.0*2.3, 2.3, 3.0, 1.8))
	g.Use(Charcoal).Fill(Capsule{c.At(h + 14), c.At(h + 17), 0.5})
	g.Use(Gold).Fill(Sphere(c.At(h+17.2), 0.6))
	s.cross(g, c.At(h+17.2), AxisX, Gold)
}

// turf is a low pitched roof shell of Wood boards topped with grass.
func (s StaveChurch) turf(g *Grid, min, max Vec2, eave float64, ridge Axis) {
	roof := s.gable(min, max, eave, 0.75, 1.3, ridge)
	g.Use(Wood).Fill(roof)
	half, c := (max.Y-min.Y)/2, (max.Y+min.Y)/2
	if ridge == AxisZ {
		half, c = (max.X-min.X)/2, (max.X+min.X)/2
	}
	g.Paint(Grass, Func{roof.Box, func(p Vec3) bool {
		a := math.Abs(p.Z - c)
		if ridge == AxisZ {
			a = math.Abs(p.X - c)
		}
		inEdge := p.X > min.X+0.5 && p.X < max.X-0.5 && p.Z > min.Y+0.5 && p.Z < max.Y-0.5
		return inEdge && p.Y > eave+(half-a)*0.75-0.7
	}})
	g.Paint(Foliage, Func{roof.Box, func(p Vec3) bool {
		return Hash01(int(math.Floor(p.X*1.4)), int(math.Floor(p.Z*1.4)), 51) < 0.12 && p.Y > eave+0.5
	}})
}

// logWalls fills a hollow log box with darker log seams and a door and window.
func (s StaveChurch) logWalls(g *Grid, r staveChurchRect, y0, y1 float64, door Vec3, through Axis) {
	solid := Subtract(r.box(y0, y1), r.inset(1.0).box(y0-1, y1+1))
	g.Use(Wood).Fill(solid)
	g.Paint(Charcoal, Intersect(solid, Func{solid.Bounds(), func(p Vec3) bool {
		return math.Mod(p.Y-y0+10, 1.6) < 0.4
	}}))
	g.Paint(Charcoal, Func{r.inset(0.5).box(y0, y1+3).Bounds(), func(p Vec3) bool { return true }})
	g.Carve(Arch{Sill: door, HalfWidth: 0.75, Spring: 1.6, Depth: 2.4, Through: through})
}

// stabbur is a storehouse raised on stilts with an overhanging upper storey and turf
// roof.
func (s StaveChurch) stabbur(g *Grid, m *staveChurchMap, c Vec2, w, d float64, ridge Axis) {
	h := m.at(c.X, c.Y)
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			at := P(c.X+sx*(w/2-0.6), c.Y+sz*(d/2-0.6))
			gh := m.at(at.X, at.Y)
			g.Use(DarkStone).Fill(Cylinder{Center: at, Radius: 0.8, Y0: math.Min(gh, h) - 1, Y1: h + 0.6})
			g.Use(Wood).Fill(Cylinder{Center: at, Radius: 0.55, Y0: h + 0.6, Y1: h + 2.2})
		}
	}
	base := h + 2.2
	r := staveChurchRect{c.X - w/2, c.Y - d/2, c.X + w/2, c.Y + d/2}
	door := V(c.X, base, r.z1-0.5)
	thr := AxisZ
	if ridge == AxisZ {
		door, thr = V(r.x0+0.5, base, c.Y), AxisX
	}
	s.logWalls(g, r, base, base+3.2, door, thr)
	up := r.inset(-0.6)
	g.Use(Wood).Fill(Subtract(up.box(base+3.2, base+4.6), up.inset(1.0).box(base, base+6)))
	g.Paint(Charcoal, up.box(base+3.2, base+3.6))
	g.Use(Wood).Fill(Block(V(r.x0-0.6, base-0.6, r.z0-0.6), V(r.x1+0.6, base, r.z1+0.6)))
	min, max := P(up.x0-0.6, up.z0-0.6), P(up.x1+0.6, up.z1+0.6)
	s.turf(g, min, max, base+4.4, ridge)
	// Gable infill.
	if ridge == AxisX {
		g.Use(Wood).Fill(s.gableWall(P(up.x0, up.z0), P(up.x0+1, up.z1), base+4.4, base+4.4, (max.Y-min.Y)/2, c.Y, 0.75, AxisX),
			s.gableWall(P(up.x1-1, up.z0), P(up.x1, up.z1), base+4.4, base+4.4, (max.Y-min.Y)/2, c.Y, 0.75, AxisX))
	} else {
		g.Use(Wood).Fill(s.gableWall(P(up.x0, up.z0), P(up.x1, up.z0+1), base+4.4, base+4.4, (max.X-min.X)/2, c.X, 0.75, AxisZ),
			s.gableWall(P(up.x0, up.z1-1), P(up.x1, up.z1), base+4.4, base+4.4, (max.X-min.X)/2, c.X, 0.75, AxisZ))
	}
}

// cabin is a turf-roofed log house on a stone footing, with a stone chimney.
func (s StaveChurch) cabin(g *Grid, m *staveChurchMap, c Vec2, w, d float64, ridge Axis) {
	h := m.at(c.X, c.Y)
	r := staveChurchRect{c.X - w/2, c.Y - d/2, c.X + w/2, c.Y + d/2}
	lo := math.Min(math.Min(m.at(r.x0, r.z0), m.at(r.x1, r.z1)), math.Min(m.at(r.x0, r.z1), m.at(r.x1, r.z0)))
	g.Use(DarkStone).Fill(r.inset(-0.3).box(lo-1, h+0.6))
	base := h + 0.6
	door := V(c.X, base, r.z1-0.5)
	thr := AxisZ
	if ridge == AxisZ {
		door, thr = V(r.x0+0.5, base, c.Y), AxisX
	}
	s.logWalls(g, r, base, base+4.0, door, thr)
	// Windows with white frames.
	for _, sign := range []float64{-1, 1} {
		var at Vec3
		if ridge == AxisZ {
			at = V(c.X+sign*w/2, base+1.6, c.Y+d/4)
		} else {
			at = V(c.X+w/4, base+1.6, c.Y+sign*d/2)
		}
		frame := Block(at.Sub(V(0.9, 0.9, 0.9)), at.Add(V(0.9, 0.9, 0.9)))
		g.Paint(White, frame)
		g.Paint(Charcoal, Block(at.Sub(V(0.45, 0.45, 0.45)), at.Add(V(0.45, 0.45, 0.45))))
	}
	min, max := P(r.x0-0.8, r.z0-0.8), P(r.x1+0.8, r.z1+0.8)
	s.turf(g, min, max, base+3.8, ridge)
	if ridge == AxisX {
		g.Use(Wood).Fill(s.gableWall(P(r.x0, r.z0), P(r.x0+1, r.z1), base+3.8, base+3.8, (max.Y-min.Y)/2, c.Y, 0.75, AxisX),
			s.gableWall(P(r.x1-1, r.z0), P(r.x1, r.z1), base+3.8, base+3.8, (max.Y-min.Y)/2, c.Y, 0.75, AxisX))
	} else {
		g.Use(Wood).Fill(s.gableWall(P(r.x0, r.z0), P(r.x1, r.z0+1), base+3.8, base+3.8, (max.X-min.X)/2, c.X, 0.75, AxisZ),
			s.gableWall(P(r.x0, r.z1-1), P(r.x1, r.z1), base+3.8, base+3.8, (max.X-min.X)/2, c.X, 0.75, AxisZ))
	}
	ch := P(c.X+w/4, c.Y-d/4)
	g.Use(Stone).Fill(BlockAt(ch, 1.4, 1.4, base+3, base+3.8+w/2*0.75+1.2))
}

// jetty is a plank pier on posts with a moored rowing boat.
func (s StaveChurch) jetty(g *Grid, m *staveChurchMap) {
	const x = 8.0
	z0 := s.shore(x) - 2.5
	g.Use(Wood).Fill(Block(V(x-1.3, 1.6, z0), V(x+1.3, 2.2, 27.5)))
	g.Use(Charcoal)
	for z := z0 + 2; z <= 27.5; z += 3 {
		for _, dx := range []float64{-1.2, 1.2} {
			g.Fill(Cylinder{Center: P(x+dx, z), Radius: 0.5, Y0: 0, Y1: 2.8})
		}
	}
	// Rowing boat alongside, bow toward the fjord.
	c := V(x+3.6, 1.1, 24.5)
	hull := Intersect(Ellipsoid{c, V(1.5, 1.4, 3.6)}, Block(V(c.X-2, 0.2, c.Z-4), V(c.X+2, c.Y+0.8, c.Z+4)))
	inner := Ellipsoid{c.Add(V(0, 0.5, 0)), V(0.8, 1.2, 2.8)}
	g.Use(Wood).FillVolume(Subtract(hull, inner))
	g.Paint(White, Block(V(c.X-2, c.Y+0.3, c.Z-4), V(c.X+2, c.Y+0.9, c.Z+4)))
	g.Use(Wood).Fill(Block(V(c.X-0.8, c.Y, c.Z-0.3), V(c.X+0.8, c.Y+0.4, c.Z+0.3)))
	g.Fill(Capsule{V(c.X+0.6, c.Y+0.5, c.Z), V(c.X+2.6, c.Y-0.2, c.Z+1.6), 0.45})
}

// spruce is a trunk with stacked foliage cones.
func (StaveChurch) spruce(g *Grid, at Vec2, base, h float64) {
	g.With(Wood, func() { g.Fill(Cylinder{Center: at, Radius: 0.55, Y0: base - 1, Y1: base + h*0.3}) })
	g.With(Foliage, func() {
		g.Fill(Cone(at, h*0.25, base+h*0.15, h*0.45), Cone(at, h*0.19, base+h*0.4, h*0.38), Cone(at, h*0.12, base+h*0.65, h*0.35))
	})
}

// birch is a white trunk with a light, airy crown.
func (StaveChurch) birch(g *Grid, at Vec2, base, h float64, k int) {
	g.With(White, func() { g.Fill(Cylinder{Center: at, Radius: 0.55, Y0: base - 1, Y1: base + h*0.65}) })
	g.Paint(Charcoal, Func{Box{V(at.X-1, base, at.Y-1), V(at.X+1, base+h, at.Y+1)}, func(p Vec3) bool {
		return Hash01(int(math.Floor(p.Y*1.3)), k, 61) < 0.25
	}})
	g.With(Grass, func() {
		g.Fill(Ellipsoid{at.At(base + h*0.72), V(h*0.26, h*0.32, h*0.26)})
		g.Fill(Ellipsoid{at.Add(P(h*0.12, h*0.08)).At(base + h*0.55), V(h*0.18, h*0.2, h*0.18)})
	})
	g.Paint(Foliage, Func{Box{V(at.X-h, base, at.Y-h), V(at.X+h, base+h*1.2, at.Y+h)}, func(p Vec3) bool {
		return Hash01(int(math.Floor(p.X*1.2)), int(math.Floor(p.Y*1.2)), int(math.Floor(p.Z*1.2)), 62) < 0.25
	}})
}

// trees scatters spruce on the mountainside and birch and spruce on the flats.
func (s StaveChurch) trees(g *Grid, m *staveChurchMap) {
	busy := []Vec3{ // Plan center and clearance radius of buildings.
		V(staveChurchBell.X, 0, staveChurchBell.Y).Add(V(0, 5, 0)), V(-28, 5, -8), V(31.5, 5, -6), V(-29.5, 6, 1.5), V(31.2, 5.5, 6.5), V(-23, 6, 20),
	}
	for i := -12; i <= 12; i++ {
		for j := -13; j <= 9; j++ {
			at := P(float64(i)*3.2+2.4*(Hash01(i, j, 71)-0.5), float64(j)*3.2+2.4*(Hash01(i, j, 72)-0.5))
			if !s.inSite(at.X, at.Y) || at.X < staveChurchX0+2 || at.X > staveChurchX1-2 || at.Y < staveChurchZ0+5 {
				continue
			}
			h := m.at(at.X, at.Y)
			mountain := at.Y < -17
			chance := 0.32
			if mountain {
				chance = 0.75
			}
			if Hash01(i, j, 73) > chance || h < 1.8 || h > 36 || h-m.low(at.X, at.Y) > 3.2 {
				continue
			}
			if s.yard(at.X, at.Y) < 2.5 || math.Abs(at.X-s.stream(at.Y)) < 3.2 {
				continue
			}
			if at.X < -18 && math.Abs(at.Y) < 3 || at.X > -8 && at.X < 10 && at.Y > 15 {
				continue // Paths.
			}
			blocked := false
			for _, b := range busy {
				if P(b.X, b.Z).Sub(at).Length() < b.Y {
					blocked = true
				}
			}
			if blocked {
				continue
			}
			size := 5 + 4*Hash01(i, j, 74)
			if mountain || Hash01(i, j, 75) < 0.45 {
				s.spruce(g, at, h-0.3, size+1.5)
			} else {
				s.birch(g, at, h-0.3, size, i*100+j)
			}
		}
	}
}

// naust is a boathouse at the water's edge: stone footing, plank walls, an open
// gable toward the fjord and a turf roof.
func (s StaveChurch) naust(g *Grid) {
	r := staveChurchRect{-26, 16.5, -20, 24}
	const y0, y1 = 0.2, 4.4
	g.Use(Wood).Fill(Subtract(r.box(y0, y1), r.inset(1.0).box(y0-1, y1+1)))
	g.Paint(Charcoal, Func{r.box(y0, y1).Bounds(), func(p Vec3) bool { return math.Mod(p.Y+10, 1.6) < 0.4 }})
	g.Use(DarkStone).Fill(Subtract(r.box(y0, 1.6), r.inset(1.0).box(y0-1, 3)))
	g.Paint(Stone, Func{r.box(y0, 1.6).Bounds(), func(p Vec3) bool {
		return Hash01(int(math.Floor(p.X/1.2)), int(math.Floor(p.Y/0.7)), int(math.Floor(p.Z/1.2)), 52) < 0.4
	}})
	g.Carve(Arch{Sill: V(-23, 0.6, r.z1-0.5), HalfWidth: 1.8, Spring: 1.6, Depth: 2.0, Through: AxisZ})
	g.Paint(Charcoal, r.inset(0.6).box(y0, y1+3))
	min, max := P(r.x0-0.7, r.z0-0.7), P(r.x1+0.7, r.z1+0.7)
	s.turf(g, min, max, y1-0.2, AxisZ)
	g.Use(Wood).Fill(s.gableWall(P(r.x0, r.z0), P(r.x1, r.z0+1), y1-0.2, y1-0.2, (max.X-min.X)/2, -23, 0.75, AxisZ),
		s.gableWall(P(r.x0, r.z1-1), P(r.x1, r.z1), y1-0.2, y1-0.2, (max.X-min.X)/2, -23, 0.75, AxisZ))
	// Slipway planks into the water.
	g.Use(Wood).Fill(Block(V(-24.2, 0.4, 24), V(-21.8, 0.9, 26.5)))
}
