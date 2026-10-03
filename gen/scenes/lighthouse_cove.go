package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// LighthouseCove is a rocky coastal headland around a sheltered cove. A red and white
// striped lighthouse with a gallery, glazed lantern and red cap stands on the western
// promontory beside the keeper's cottage; cliff stairs drop to a sandy beach; a fishing
// village with a stone quay, net sheds and a boathouse lines the eastern shore, behind
// a curved breakwater with a harbor light. Boats are moored at a wooden jetty and the
// quay; sea stacks and boulders stand off the cliffs in a shallow water layer at y = 0.
// North is -Z; the cove opens south (+Z).
type LighthouseCove struct{}

func init() { Register(LighthouseCove{}) }

func (LighthouseCove) Name() string  { return "lighthouse_cove" }
func (LighthouseCove) Title() string { return "Lighthouse Cove" }

// Key positions.
var (
	lighthouseCoveLight   = P(-35, 20)  // Lighthouse center.
	lighthouseCoveCottage = P(-39, 4.5) // Keeper's cottage center.
	lighthouseCoveChapel  = P(3, -41)   // Chapel on the hill.
)

const (
	lighthouseCovePlateau = 13.0 // Building pad height on the headland.
	lighthouseCoveQuay    = 4.2  // Village and quay level.
)

// lighthouseCoveLand is the terrain height field, memoized per sample point.
type lighthouseCoveLand struct {
	cache map[[2]int64]float64
}

func lighthouseCoveSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func lighthouseCoveLerp(a, b, t float64) float64 { return a + (b-a)*t }

// noise is smooth value noise in [-1, 1].
func lighthouseCoveNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return lighthouseCoveLerp(lighthouseCoveLerp(a, b, sx), lighthouseCoveLerp(c, d, sx), sz)*2 - 1
}

func lighthouseCoveFBM(x, z float64, seed int) float64 {
	return 0.6*lighthouseCoveNoise(x, z, seed) + 0.3*lighthouseCoveNoise(x*2.03+17, z*2.03+5, seed+1) +
		0.1*lighthouseCoveNoise(x*4.1+3, z*4.1+29, seed+2)
}

// inBounds is the rounded-rectangle site, water included.
func (LighthouseCove) inBounds(x, z float64) bool {
	const hx, hz, r = 62, 45, 16
	dx, dz := math.Max(math.Abs(x)-(hx-r), 0), math.Max(math.Abs(z+3)-(hz-r), 0)
	return dx*dx+dz*dz <= r*r
}

// segment returns the plan distance from p to segment ab and the parameter along it.
func (LighthouseCove) segment(p, a, b Vec2) (float64, float64) {
	ab := b.Sub(a)
	t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
	return p.Sub(a.Add(ab.Scale(t))).Length(), t
}

// polyline returns the distance to a polyline, the arc length at the nearest point, and
// the signed side (negative to the right of travel, seen with +Z down the page).
func (s LighthouseCove) polyline(p Vec2, pts []Vec2) (dist, along, side float64) {
	dist = math.Inf(1)
	run := 0.0
	for i := 0; i+1 < len(pts); i++ {
		a, b := pts[i], pts[i+1]
		d, t := s.segment(p, a, b)
		l := b.Sub(a).Length()
		if d < dist {
			dir := b.Sub(a).Scale(1 / l)
			q := p.Sub(a)
			dist, along, side = d, run+t*l, dir.X*q.Y-dir.Y*q.X
		}
		run += l
	}
	return
}

// Stair: from the beach up the back cliffs to the clifftop path.
var (
	lighthouseCoveStairA = P(-1, -21)
	lighthouseCoveStairB = P(-11, -35)
)

const (
	lighthouseCoveStairY0 = 1.6
	lighthouseCoveStairY1 = 11.5
)

func (s LighthouseCove) raw(x, z float64) float64 {
	if !s.inBounds(x, z) {
		return 0
	}
	coast := 2.2 * lighthouseCoveFBM(x*0.07, z*0.07, 11)
	blob := func(cx, cz, rx, rz, height, w float64) float64 {
		r := math.Hypot((x-cx)/rx, (z-cz)/rz)
		e := (1-r)*math.Min(rx, rz) + coast
		return height * lighthouseCoveSmooth(e/w)
	}
	// rect has dressed quay walls on its west and south sides and rocky shores elsewhere.
	rect := func(x0, z0, x1, z1, height float64) float64 {
		quay := math.Min(x-x0, z1-z)
		shore := math.Min(x1-x, z-z0) + coast
		return height * math.Min(lighthouseCoveSmooth(quay/0.8), lighthouseCoveSmooth(shore/2.5))
	}
	h := 0.0
	for _, v := range []float64{
		blob(-38, -16, 22, 28, 13.5, 3.5),         // Western headland.
		blob(-35, 17, 11, 16, 13, 3),              // Lighthouse promontory.
		blob(-6, -38, 26, 9, 11.5, 3.5),           // Back cliffs behind the beach.
		blob(12, -34, 20, 11, 8, 9),               // Hill falling to the village.
		blob(0, -18, 16, 9, 2.6, 7),               // Beach.
		rect(16, -42, 56, -8, lighthouseCoveQuay), // Village quay.
		rect(44, -42, 57, 8, lighthouseCoveQuay),  // Eastern arm.
	} {
		h = math.Max(h, v)
	}
	if h > 5 { // Rolling turf on the high ground.
		h += 0.8 * lighthouseCoveFBM(x*0.13, z*0.13, 23) * lighthouseCoveSmooth((h-5)/4)
	}
	// Level pads for the lighthouse and the cottage.
	for _, pad := range []struct {
		c    Vec2
		r, y float64
	}{{lighthouseCoveLight, 8, lighthouseCovePlateau}, {lighthouseCoveCottage, 9, lighthouseCovePlateau},
		{lighthouseCoveChapel, 8, 11.5}} {
		if h > 8 {
			d := P(x, z).Sub(pad.c).Length()
			h = lighthouseCoveLerp(h, pad.y, lighthouseCoveSmooth((pad.r-d)/3))
		}
	}
	// Cliff stair: a stepped ledge cut into (and built out from) the rock face.
	if d, t := s.segment(P(x, z), lighthouseCoveStairA, lighthouseCoveStairB); d < 1.5 {
		steps := 14.0
		y := lighthouseCoveStairY0 + math.Floor(t*steps+0.5)/steps*(lighthouseCoveStairY1-lighthouseCoveStairY0)
		h = math.Min(h, y)
	}
	return h
}

// height is the memoized terrain height.
func (s LighthouseCove) height(land *lighthouseCoveLand, x, z float64) float64 {
	key := [2]int64{int64(math.Round(x * 1000)), int64(math.Round(z * 1000))}
	if h, ok := land.cache[key]; ok {
		return h
	}
	h := s.raw(x, z)
	land.cache[key] = h
	return h
}

// shell fills the cells of a solid predicate that lie within t of its outside (above
// or sideways): a hollow skin, open underneath.
func (LighthouseCove) shell(g *Grid, box Box, t float64, in func(Vec3) bool) {
	g.Fill(Func{box, func(p Vec3) bool {
		if !in(p) {
			return false
		}
		return !(in(p.Add(V(t, 0, 0))) && in(p.Add(V(-t, 0, 0))) && in(p.Add(V(0, 0, t))) &&
			in(p.Add(V(0, 0, -t))) && in(p.Add(V(0, t, 0))))
	}})
}

// rock fills a hollow, noisy boulder or stack: radius r at the base tapering by taper
// toward height top.
func (s LighthouseCove) rock(g *Grid, c Vec2, r, top, taper float64, seed int) {
	in := func(p Vec3) bool {
		if p.Y < -0.2 || p.Y > top {
			return false
		}
		d := p.Plan().Sub(c)
		a := d.Angle()
		u := p.Y / top
		rr := r * (1 - taper*u) * (0.85 + 0.25*lighthouseCoveNoise(math.Cos(a)*2+p.Y*0.25, math.Sin(a)*2, seed))
		if u > 0.75 { // Rounded crown.
			k := (u - 0.75) / 0.25
			rr *= math.Sqrt(math.Max(1-k*k, 0))
		}
		return d.Length() <= rr
	}
	box := Box{V(c.X-r*1.2, 0, c.Y-r*1.2), V(c.X+r*1.2, top, c.Y+r*1.2)}
	s.shell(g, box, 1.2, in)
}

// house is a hollow cottage with glazed windows, a door, a gable roof and a chimney.
func (s LighthouseCove) house(g *Grid, land *lighthouseCoveLand, c Vec2, w, d, wall float64, ridge Axis, walls, roof Color, door int) {
	x0, x1, z0, z1 := c.X-w/2, c.X+w/2, c.Y-d/2, c.Y+d/2
	base := 0.0
	low := math.Inf(1)
	for _, q := range []Vec2{P(x0, z0), P(x1, z0), P(x0, z1), P(x1, z1), c} {
		h := s.height(land, q.X, q.Y)
		base, low = math.Max(base, h), math.Min(low, h)
	}
	base = math.Round(base*2) / 2
	top := base + wall
	g.With(Stone, func() {
		g.Fill(Block(V(x0-0.3, low-1.5, z0-0.3), V(x1+0.3, base+0.5, z1+0.3)))
	})
	g.With(walls, func() {
		g.Fill(Block(V(x0, base, z0), V(x1, top, z1)))
		g.Carve(Block(V(x0+1, base+0.6, z0+1), V(x1-1, top+3, z1-1)))
	})
	// Windows along both long walls, glazed in the middle of the wall.
	sill, head := base+1.6, math.Min(base+3.2, top-0.6)
	if ridge == AxisX {
		for x := x0 + 2.2; x <= x1-2; x += 3 {
			for _, z := range []float64{z0, z1} {
				g.Carve(Block(V(x-0.6, sill, z-0.6), V(x+0.6, head, z+0.6)))
				g.With(Glass, func() { g.Fill(Block(V(x-0.6, sill, z-0.2), V(x+0.6, head, z+0.2))) })
			}
		}
	} else {
		for z := z0 + 2.2; z <= z1-2; z += 3 {
			for _, x := range []float64{x0, x1} {
				g.Carve(Block(V(x-0.6, sill, z-0.6), V(x+0.6, head, z+0.6)))
				g.With(Glass, func() { g.Fill(Block(V(x-0.2, sill, z-0.6), V(x+0.2, head, z+0.6))) })
			}
		}
	}
	// Door: 0 = +Z, 1 = +X, 2 = -Z, 3 = -X.
	dc := []Vec3{V(c.X, base, z1), V(x1, base, c.Y), V(c.X, base, z0), V(x0, base, c.Y)}[door]
	through := AxisZ
	if door%2 == 1 {
		through = AxisX
	}
	g.Carve(Arch{dc, 0.8, 1.8, 1.4, through})
	g.With(Wood, func() { g.Fill(Arch{dc, 0.8, 1.8, 0.4, through}) })
	// Roof with overhanging eaves, hollowed underneath.
	g.With(roof, func() {
		g.Fill(Gable{P(x0-0.7, z0-0.7), P(x1+0.7, z1+0.7), top, ridge})
		g.Carve(Gable{P(x0+0.6, z0+0.6), P(x1-0.6, z1-0.6), top - 0.6, ridge})
	})
	// Chimney at one gable end.
	var ch Vec2
	half := d / 2
	if ridge == AxisX {
		ch = P(x1-1.2, c.Y)
	} else {
		ch = P(c.X, z1-1.2)
		half = w / 2
	}
	g.With(Stone, func() { g.Fill(BlockAt(ch, 1.4, 1.4, top, top+half+1.6)) })
	g.With(Charcoal, func() { g.Fill(BlockAt(ch, 0.8, 0.8, top+half+1.6, top+half+2.2)) })
}

// boat is a small fishing boat at c, bow toward angle, built in the water layer.
func (s LighthouseCove) boat(g *Grid, c Vec2, angle, length, beam float64, hull, band Color, cabin bool) {
	L := length
	halfw := func(u, y float64) float64 {
		if u < -L/2 || u > L/2 {
			return -1
		}
		w := beam / 2
		if u0 := 0.05 * L; u > u0 {
			t := (u - u0) / (L/2 - u0)
			w *= math.Sqrt(math.Max(1-t*t, 0))
		} else if u < -0.3*L {
			t := (-0.3*L - u) / (0.2 * L)
			w *= 1 - 0.25*t*t
		}
		if y < 1.2 {
			t := (1.2 - y) / 1.2
			w *= 1 - 0.45*t*t
		}
		return w
	}
	sheer := func(u float64) float64 { k := u / (L / 2); return 0.25 + 0.7*k*k + 0.3*k }
	depth := 2.0
	local := func(p Vec3) (float64, float64) {
		q := p.Plan().Sub(c).Rotate(-angle)
		return q.X, q.Y
	}
	box := Box{V(c.X-L/2-1, 0, c.Y-L/2-1), V(c.X+L/2+1, depth+1.5, c.Y+L/2+1)}
	g.With(hull, func() {
		g.FillVolume(Func{box, func(p Vec3) bool {
			u, v := local(p)
			return p.Y >= -0.2 && p.Y <= depth+sheer(u) && math.Abs(v) <= halfw(u, p.Y)
		}})
	})
	interior := Func{box, func(p Vec3) bool {
		u, v := local(p)
		w := math.Min(halfw(u-0.9, p.Y), halfw(u+0.9, p.Y)) - 0.9
		return p.Y >= 0.8 && math.Abs(v) <= w
	}}
	g.Carve(interior)
	// Gunwale band and a wooden floor.
	g.Paint(band, Func{box, func(p Vec3) bool {
		u, _ := local(p)
		return p.Y >= depth+sheer(u)-0.6
	}})
	g.Paint(Wood, Func{box, func(p Vec3) bool {
		u, v := local(p)
		w := math.Min(halfw(u-0.9, 1.0), halfw(u+0.9, 1.0)) - 0.9
		return p.Y < 0.8 && p.Y > 0.1 && math.Abs(v) <= w
	}})
	if !cabin {
		// Thwarts (seats) across the open boat.
		for _, k := range []float64{-0.18, 0.12} {
			g.With(Wood, func() {
				g.Fill(Func{box, func(p Vec3) bool {
					u, v := local(p)
					return math.Abs(u-k*L) < 0.45 && p.Y >= 1.0 && p.Y <= 1.5 && math.Abs(v) <= halfw(u, p.Y)
				}})
			})
		}
		return
	}
	at := func(u, v float64) Vec2 { return c.Add(P(u, v).Rotate(angle)) }
	// Wheelhouse aft of amidships, with a glazed band and a roof.
	wh := Func{box, func(p Vec3) bool {
		u, v := local(p)
		return u >= -0.28*L && u <= -0.02*L && math.Abs(v) <= beam/2-1 && p.Y >= 0.8 && p.Y <= 4.4
	}}
	g.With(White, func() { g.Fill(wh) })
	g.Paint(Glass, Func{box, func(p Vec3) bool {
		u, v := local(p)
		return p.Y >= 2.9 && p.Y <= 3.7 && wh.F(p) && (math.Abs(v) > beam/2-1.6 || u > -0.02*L-0.6)
	}})
	g.With(band, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			u, v := local(p)
			return u >= -0.3*L && u <= 0 && math.Abs(v) <= beam/2-0.6 && p.Y >= 4.4 && p.Y <= 4.9
		}})
	})
	// Mast forward and a short gantry aft.
	m := at(0.18*L, 0)
	g.With(Wood, func() { g.Fill(Cylinder{Center: m, Radius: 0.5, Y0: 0.5, Y1: 8}) })
	g.With(Wood, func() { g.Fill(Capsule{at(0.18*L, 0).At(6.5), at(-0.05*L, 0).At(4.6), 0.45}) })
	g.With(Red, func() { g.Fill(Sphere(m.At(8.3), 0.6)) })
}

func (s LighthouseCove) Build(g *Grid) Order {
	land := &lighthouseCoveLand{cache: make(map[[2]int64]float64)}
	H := func(x, z float64) float64 { return s.height(land, x, z) }
	site := Box{V(-64, -0.2, -48), V(64, 16, 48)}

	// Sea: a single shallow layer with surf where it laps the land.
	g.Use(Water).Fill(Func{Box{V(-64, -0.2, -48), V(64, 0.2, 48)}, func(p Vec3) bool {
		return s.inBounds(p.X, p.Z)
	}})
	// Paler shallows along the shore and scattered glints.
	g.Paint(Glass, Func{Box{V(-64, -0.2, -48), V(64, 0.2, 48)}, func(p Vec3) bool {
		c := Nearest(p.Scale(1 / g.Span()))
		if Hash01(int(c.X), int(c.Z), 6) < 0.03 {
			return true
		}
		for k := 0; k < 8; k++ {
			q := Polar(p.Plan(), 2.6, float64(k)*math.Pi/4)
			if H(q.X, q.Y) >= 0.45 {
				return Hash01(int(c.X), int(c.Z), 7) < 0.7
			}
		}
		return false
	}})
	g.Paint(White, Func{Box{V(-64, -0.2, -48), V(64, 0.2, 48)}, func(p Vec3) bool {
		h := H(p.X, p.Z)
		c := Nearest(p.Scale(1 / g.Span()))
		return h > 0.08 && h < 0.45 && Hash01(int(c.X), int(c.Z), 5) < 0.55
	}})

	// Land: a hollow skin over the height field.
	inLand := func(p Vec3) bool {
		if p.Y < -0.2 || !s.inBounds(p.X, p.Z) {
			return false
		}
		h := H(p.X, p.Z)
		return h >= 0.45 && p.Y <= h
	}
	g.Use(DarkStone)
	s.shell(g, site, 1.2, inLand)

	slope := func(x, z float64) float64 {
		dx := (H(x+0.6, z) - H(x-0.6, z)) / 1.2
		dz := (H(x, z+0.6) - H(x, z-0.6)) / 1.2
		return math.Hypot(dx, dz)
	}
	top := func(p Vec3) bool { h := H(p.X, p.Z); return h >= 0.45 && p.Y > h-0.9 }
	// Rock strata in the cliffs, dark and wet near the water.
	g.Paint(Stone, Func{site, func(p Vec3) bool {
		band := int(math.Floor((p.Y + 0.9*lighthouseCoveFBM(p.X*0.15, p.Z*0.15, 31)) / 1.3))
		return p.Y > 1 && Hash01(band, 3) < 0.6
	}})
	village := func(x, z float64) bool { return x >= 15.5 && z >= -40.5 && (z <= -7.5 || x >= 43.5) }
	g.Paint(Grass, Func{site, func(p Vec3) bool {
		return top(p) && H(p.X, p.Z) >= 3 && slope(p.X, p.Z) < 1.1
	}})
	g.Paint(Sand, Func{site, func(p Vec3) bool {
		return top(p) && H(p.X, p.Z) < 3.4 && slope(p.X, p.Z) < 1.5 && !village(p.X, p.Z)
	}})
	// Quay walls and paved waterfront in dressed stone.
	g.Paint(Stone, Func{site, func(p Vec3) bool {
		if !village(p.X, p.Z) || p.Y > lighthouseCoveQuay+0.5 || H(p.X, p.Z) < 0.45 {
			return false
		}
		edge := math.Min(p.X-16, -8-p.Z)
		if p.X >= 44 {
			edge = 8 - p.Z
			if p.Z > -8 {
				edge = math.Min(edge, p.X-44)
			}
		}
		return edge < 3.5
	}})
	// Footpath from the village over the hill to the lighthouse, and the cliff stair.
	path := []Vec2{P(40, -26), P(28, -30), P(16, -34), P(4, -36), P(-8, -33), P(-16, -28), P(-23, -20),
		P(-30, -10), P(-36, -2), P(-34, 8), lighthouseCoveLight}
	g.Paint(Sand, Func{site, func(p Vec3) bool {
		d, _, _ := s.polyline(p.Plan(), path)
		return d < 1.1 && top(p)
	}})
	g.Paint(Stone, Func{site, func(p Vec3) bool {
		d, _ := s.segment(p.Plan(), lighthouseCoveStairA, lighthouseCoveStairB)
		return d < 1.5 && p.Y > 1 && top(p)
	}})

	s.lighthouse(g)
	s.cottage(g, land)
	s.harbor(g, land)

	// Sea stacks and boulders off the cliffs.
	g.Use(DarkStone)
	s.rock(g, P(-22, 38), 4, 14, 0.35, 1)
	s.rock(g, P(-13, 33), 2.4, 8, 0.3, 2)
	s.rock(g, P(-56, 22), 2.2, 6, 0.3, 4)
	for i := 0; i < 26; i++ {
		a := Hash01(i, 41) * 2 * math.Pi
		c := P(-35+40*math.Cos(a)*0.9, -10+38*math.Sin(a)*0.95)
		if h := H(c.X, c.Y); h > 1 || !s.inBounds(c.X, c.Y) {
			continue
		}
		s.rock(g, c, 1.2+1.4*Hash01(i, 42), 1.5+2.5*Hash01(i, 43), 0.2, 10+i)
	}
	// A natural arch: a rock fin off the headland pierced by the sea.
	archA, archB := P(-59, 31), P(-45, 31)
	hole := Arch{V(-52, -1, 31), 3.0, 4.0, 12, AxisZ}
	s.shell(g, Box{V(-64, -0.2, 24), V(-40, 13, 38)}, 1.2, func(p Vec3) bool {
		if p.Y < -0.2 || p.Y > 12 || hole.Contains(p) {
			return false
		}
		d, t := s.segment(p.Plan(), archA, archB)
		u := p.Y / 12
		end := math.Min(t, 1-t) * 14 // Taper the fin toward both ends.
		r := (3.6 - 1.1*u) * (0.85 + 0.25*lighthouseCoveNoise(p.X*0.35, p.Y*0.3, 7))
		crown := 12 - 0.35*math.Max(0, 5-end)*2 + 1.2*lighthouseCoveNoise(p.X*0.2, 3, 8)
		return d <= r && p.Y <= crown
	})
	g.Paint(Stone, Func{Box{V(-64, 2, -48), V(64, 16, 48)}, func(p Vec3) bool {
		c := Nearest(p.Scale(1 / g.Span()))
		col, _ := g.ColorOf(c)
		return col == DarkStone && Hash01(int(math.Floor(p.Y/1.3)), 9) < 0.4
	}})
	// Gorse on the high ground: dark clumps flecked with yellow.
	pathD := func(p Vec2) float64 { d, _, _ := s.polyline(p, path); return d }
	for i := -60; i <= 60; i += 3 {
		for k := -46; k <= 40; k += 3 {
			x, z := float64(i)+1.5*Hash01(i, k, 81), float64(k)+1.5*Hash01(i, k, 82)
			h := H(x, z)
			yard := x > -49 && x < -26 && z > -3 && z < 30
			if Hash01(i, k, 83) > 0.2 || h < 5 || slope(x, z) > 0.6 || yard || pathD(P(x, z)) < 2.5 {
				continue
			}
			r := 1.0 + 0.7*Hash01(i, k, 84)
			g.With(Foliage, func() { g.Fill(Ellipsoid{V(x, h+0.3, z), V(r*1.2, r*0.8, r)}) })
		}
	}
	g.Paint(Gold, Func{site, func(p Vec3) bool {
		c := Nearest(p.Scale(1 / g.Span()))
		col, _ := g.ColorOf(c)
		return col == Foliage && p.Y > H(p.X, p.Z)+0.6 && Hash01(int(c.X), int(c.Y), int(c.Z), 85) < 0.18
	}})
	// Village trees.
	for _, t := range []Vec2{P(22, -27), P(46, -24), P(36, -38), P(18, -38), P(-14, -42)} {
		h := H(t.X, t.Y)
		g.With(Wood, func() { g.Fill(Cylinder{Center: t, Radius: 0.6, Y0: h - 0.5, Y1: h + 3.5}) })
		g.With(Foliage, func() { g.FillVolume(Ellipsoid{t.At(h + 5), V(2.6, 2.4, 2.6)}) })
	}
	return Sweep(P(0, -5), math.Pi)
}

// lighthouse is the striped tower on the promontory.
func (s LighthouseCove) lighthouse(g *Grid) {
	c := lighthouseCoveLight
	y0 := lighthouseCovePlateau
	// Plinth.
	g.Use(Stone).Fill(Cylinder{Center: c, Radius: 6.5, Y0: y0 - 2.5, Y1: y0 + 1.2})
	g.Carve(Cylinder{Center: c, Radius: 5.3, Y0: y0 - 2, Y1: y0 + 0.2})
	base, T := y0+1.2, y0+33.0
	g.Use(White).Fill(Frustum{c, 4.5, 3.0, base, T})
	g.Carve(Frustum{c, 3.3, 1.8, base + 0.6, T - 0.6})
	g.Paint(Red, Func{Box{V(c.X-5, base, c.Y-5), V(c.X+5, T, c.Y+5)}, func(p Vec3) bool {
		return int(math.Floor((p.Y-base)/3.8))%2 == 1
	}})
	// Door facing east and small windows climbing the tower.
	g.Carve(Arch{V(c.X+4.3, base, c.Y), 1.0, 2.0, 2.5, AxisX})
	g.With(Wood, func() { g.Fill(Arch{V(c.X+3.6, base, c.Y), 1.0, 2.0, 0.5, AxisX}) })
	for k := 0; k < 6; k++ {
		y := base + 5 + float64(k)*4.2
		a := 0.4 + float64(k)*1.9
		r := 4.5 - 1.5*(y-base)/(T-base)
		slit := Arch{Sill: V(r-0.6, y, 0), HalfWidth: 0.6, Spring: 1.0, Depth: 2.4, Through: AxisX}
		g.Carve(Translate(RotateY(slit, P(0, 0), a), c.At(0)))
	}
	// Gallery: corbel, deck and railing.
	g.Use(Charcoal)
	g.Fill(Frustum{c, 3.0, 4.9, T - 1.6, T})
	g.Fill(Cylinder{Center: c, Radius: 4.9, Y0: T, Y1: T + 0.7})
	for k := 0; k < 12; k++ {
		post := Polar(c, 4.45, float64(k)*math.Pi/6)
		g.Fill(Cylinder{Center: post, Radius: 0.45, Y0: T + 0.7, Y1: T + 2.6})
	}
	g.Fill(Cylinder{Center: c, Radius: 4.9, Inner: 4.0, Y0: T + 2.1, Y1: T + 2.7})
	// Lantern: a charcoal base, glazing with mullions, a lamp, a red dome and vent.
	g.Fill(Cylinder{Center: c, Radius: 2.9, Y0: T + 0.7, Y1: T + 1.8})
	g.With(Glass, func() { g.Fill(Cylinder{Center: c, Radius: 2.7, Inner: 1.7, Y0: T + 1.8, Y1: T + 5.4}) })
	for k := 0; k < 8; k++ {
		m := Polar(c, 2.55, float64(k)*math.Pi/4+math.Pi/8)
		g.Fill(Cylinder{Center: m, Radius: 0.45, Y0: T + 1.8, Y1: T + 5.4})
	}
	g.With(Gold, func() { g.Fill(Sphere(c.At(T+3.6), 1.2)) })
	g.Fill(Cylinder{Center: c, Radius: 3.3, Y0: T + 5.4, Y1: T + 6.0})
	g.With(Red, func() {
		g.Fill(Intersect(Ellipsoid{c.At(T + 6.0), V(3.2, 2.8, 3.2)}, Block(V(c.X-4, T+6.0, c.Y-4), V(c.X+4, T+9, c.Y+4))))
	})
	g.Fill(Sphere(c.At(T+9.3), 0.9))
	g.Fill(Cylinder{Center: c, Radius: 0.4, Y0: T + 9.5, Y1: T + 11})
}

// cottage is the keeper's house beside the lighthouse with a whitewashed yard wall.
func (s LighthouseCove) cottage(g *Grid, land *lighthouseCoveLand) {
	c := lighthouseCoveCottage
	s.house(g, land, c, 13, 7, 4.5, AxisX, White, Terracotta, 0)
	// A lean-to store on the west end.
	s.house(g, land, P(c.X-8.5, c.Y+0.5), 4, 5, 3, AxisZ, White, Terracotta, 0)
	// Yard wall enclosing cottage and tower, with a gate toward the path.
	y := lighthouseCovePlateau
	wall := []Vec2{P(-47, -1), P(-28, -1), P(-27, 13), P(-27, 28), P(-43, 28), P(-48, 12), P(-47, -1)}
	g.Use(White).Fill(Func{Box{V(-50, y-1, -3), V(-25, y+1.3, 30)}, func(p Vec3) bool {
		d, _, _ := s.polyline(p.Plan(), wall)
		gate := p.Plan().Sub(P(-34, -1)).Length() < 1.8 || p.Plan().Sub(P(-27, 10)).Length() < 1.8
		return d < 0.55 && p.Y >= s.height(land, p.X, p.Z)-1 && !gate
	}})
}

// harbor builds the village, quay furniture, jetty, breakwater, light and boats.
func (s LighthouseCove) harbor(g *Grid, land *lighthouseCoveLand) {
	q := lighthouseCoveQuay
	// Village houses up the slope, whitewashed and colored.
	type hs struct {
		c        Vec2
		w, d, h  float64
		ridge    Axis
		wall, rf Color
		door     int
	}
	for _, h := range []hs{
		{P(28, -20), 9, 6, 4, AxisX, White, Terracotta, 0},
		{P(39, -21), 8, 6, 4.5, AxisX, Sand, Slate, 0},
		{P(50, -14), 6, 9, 4, AxisZ, White, Terracotta, 3},
		{P(50, -30), 6, 8, 5, AxisZ, White, Slate, 3},
		{P(31, -32), 8, 6, 4, AxisX, White, Slate, 0},
		{P(42, -34), 7, 6, 4, AxisX, Sand, Terracotta, 0},
	} {
		s.house(g, land, h.c, h.w, h.d, h.h, h.ridge, h.wall, h.rf, h.door)
	}
	// Harbor office on the eastern arm.
	s.house(g, land, P(51, 0), 6, 7, 4, AxisZ, White, Red, 3)
	// Whitewashed chapel on the hill, with a bell tower at its west end.
	ch := lighthouseCoveChapel
	s.house(g, land, ch, 11, 6, 5, AxisX, White, Slate, 1)
	ty := 11.5
	g.With(White, func() {
		g.Fill(BlockAt(P(ch.X-7, ch.Y), 3.4, 3.4, ty-1, ty+11))
		g.Carve(BlockAt(P(ch.X-7, ch.Y), 1.6, 1.6, ty, ty+10.5))
		g.Carve(Arch{V(ch.X-7, ty+8, ch.Y), 0.7, 1.2, 4, AxisX}, Arch{V(ch.X-7, ty+8, ch.Y), 0.7, 1.2, 4, AxisZ})
	})
	g.With(Slate, func() { g.Fill(HipRoof{P(ch.X-9.1, ch.Y-2.1), P(ch.X-4.9, ch.Y+2.1), ty + 11}) })
	g.With(Gold, func() { g.Fill(Cylinder{Center: P(ch.X-7, ch.Y), Radius: 0.45, Y0: ty + 13, Y1: ty + 15}) })
	// Tall tarred net sheds along the quay.
	for i, x := range []float64{36, 39.5, 43} {
		c := P(x, -12.5)
		hh := 6.5 + float64(i%2)
		g.With(Charcoal, func() {
			g.Fill(BlockAt(c, 3, 3, q, q+hh))
			g.Carve(BlockAt(c, 1.4, 1.4, q+0.5, q+hh+2))
			g.Fill(Gable{P(c.X-1.9, c.Y-1.9), P(c.X+1.9, c.Y+1.9), q + hh, AxisZ})
		})
		g.Carve(Block(V(c.X-0.6, q, c.Y+1), V(c.X+0.6, q+2.4, c.Y+2)))
		g.With(Wood, func() { g.Fill(Block(V(c.X-0.6, q, c.Y+1.1), V(c.X+0.6, q+2.4, c.Y+1.4))) })
	}
	// Boathouse on the quay with a slipway down to the water.
	bc := P(21, -15)
	g.With(Wood, func() {
		g.Fill(BlockAt(bc, 8, 12, q, q+5))
		g.Carve(BlockAt(bc, 6, 10, q+0.5, q+8))
	})
	g.Carve(Arch{V(bc.X, q, bc.Y+6), 2.4, 2.6, 3, AxisZ})
	g.With(Red, func() {
		g.Fill(Gable{P(bc.X-4.6, bc.Y-6.6), P(bc.X+4.6, bc.Y+6.6), q + 5, AxisZ})
		g.Carve(Gable{P(bc.X-3.4, bc.Y-6.6), P(bc.X+3.4, bc.Y+6.6), q + 4.6, AxisZ})
	})
	g.With(Stone, func() {
		g.Fill(Func{Box{V(bc.X-2.5, 0, -8), V(bc.X+2.5, q, 2)}, func(p Vec3) bool {
			return p.Y <= q-(p.Z+8)*q/10 && p.Y >= -0.2
		}})
	})
	// Bollards and crates on the quay.
	g.Use(Charcoal)
	for x := 27.0; x <= 52; x += 5 {
		g.Fill(Cylinder{Center: P(x, -9), Radius: 0.5, Y0: q, Y1: q + 1.2})
	}
	for i := 0; i < 8; i++ {
		c := P(26+Hash01(i, 61)*16, -10.5-Hash01(i, 62)*2)
		col := []Color{Wood, Gold, Navy, Red}[i%4]
		g.With(col, func() { g.Fill(BlockAt(c, 1.2, 1.2, q, q+1.0+float64(i%2)*0.8)) })
	}
	// Wooden jetty from the beach into the cove.
	jx, j0, j1 := 9.0, -14.0, 12.0
	deck := 2.4
	g.Use(Wood).Fill(Block(V(jx-1.6, deck-0.5, j0), V(jx+1.6, deck, j1)))
	for z := j0 + 1; z <= j1; z += 3.2 {
		for _, side := range []float64{-1.3, 1.3} {
			g.Fill(Cylinder{Center: P(jx+side, z), Radius: 0.5, Y0: -0.2, Y1: deck + 1.0})
		}
	}
	g.Fill(Block(V(jx-3.5, deck-0.5, j1-4), V(jx+3.5, deck, j1)))
	// Stone breakwater: a curved arm with a seaward parapet, riprap and a harbor light.
	arm := []Vec2{P(50, 6), P(51, 20), P(46, 30), P(32, 35)}
	box := Box{V(25, -0.2, 0), V(60, 6, 42)}
	g.Use(Stone).Fill(Func{box, func(p Vec3) bool {
		d, _, side := s.polyline(p.Plan(), arm)
		if d > 2.6 || p.Y < -0.2 {
			return false
		}
		if p.Y <= 3.4 {
			return d > 1.2 || p.Y >= 2.4
		}
		return side < 0 && d > 1.5 && p.Y <= 4.6
	}})
	g.Use(DarkStone)
	for i := 0; i < 34; i++ {
		t := float64(i)/34*44 + Hash01(i, 71)
		// Point along the arm at arc length t, pushed to seaward.
		run := 0.0
		for k := 0; k+1 < len(arm); k++ {
			l := arm[k+1].Sub(arm[k]).Length()
			if t <= run+l || k == len(arm)-2 {
				dir := arm[k+1].Sub(arm[k]).Scale(1 / l)
				pt := arm[k].Add(dir.Scale(t - run))
				out := P(dir.Y, -dir.X) // Seaward: the negative side of travel.
				c := pt.Add(out.Scale(3.4 + Hash01(i, 72)))
				r := 1.2 + 0.8*Hash01(i, 73)
				g.Fill(Ellipsoid{c.At(0.4), V(r, r*0.9, r)})
				break
			}
			run += l
		}
	}
	lc := P(32, 35)
	g.Use(Stone).Fill(Cylinder{Center: lc, Radius: 2.6, Y0: -0.2, Y1: 3.4})
	g.Use(White).Fill(Cylinder{Center: lc, Radius: 1.6, Y0: 3.4, Y1: 9})
	g.Paint(Red, Block(V(lc.X-2, 6.5, lc.Y-2), V(lc.X+2, 9, lc.Y+2)))
	g.With(Glass, func() { g.Fill(Cylinder{Center: lc, Radius: 1.2, Y0: 9, Y1: 10.6}) })
	g.With(Red, func() { g.Fill(Cone(lc, 1.8, 10.6, 1.8)) })
	// Boats: moored along the jetty and the quay, one hauled up on the beach.
	s.boat(g, P(13.5, 2), math.Pi/2, 10, 4, Navy, White, true)
	s.boat(g, P(4.5, 4), math.Pi/2+0.1, 7, 3, Red, White, false)
	s.boat(g, P(28, -4.5), 0, 11, 4.2, Red, White, true)
	s.boat(g, P(40, -4.2), 0.05, 9, 3.6, White, Navy, true)
	s.boat(g, P(46.5, 10), -math.Pi/2+0.1, 7, 3, Gold, Wood, false)
	s.boat(g, P(26, 18), 2.5, 10, 4, Terracotta, White, true)
	s.boat(g, P(-4, -14.5), -1.9, 6, 2.8, Navy, White, false)
	// A sailing dinghy heading in past the stacks, and the harbor-mouth buoys.
	s.boat(g, P(4, 26), -2.2, 8, 3.2, White, Navy, false)
	at := func(u, v float64) Vec2 { return P(4, 26).Add(P(u, v).Rotate(-2.2)) }
	g.With(Wood, func() { g.Fill(Cylinder{Center: at(1.4, 0), Radius: 0.45, Y0: 0.5, Y1: 11}) })
	g.With(White, func() {
		g.Fill(Func{Box{V(-6, 1, 16), V(14, 11, 36)}, func(p Vec3) bool {
			q := p.Plan().Sub(P(4, 26)).Rotate(2.2)
			u, v := q.X, q.Y
			t := (p.Y - 2.2) / 8.6
			if t < 0 || t > 1 || u > 1.0 {
				return false
			}
			reach := 6.0 * (1 - t)
			belly := 0.9 * math.Sin(math.Pi*math.Max(0, 1.0-u)/math.Max(reach, 0.1)) * (1 - t)
			return 1.0-u <= reach && math.Abs(v-belly) <= 0.45
		}})
	})
	g.With(Red, func() { g.Fill(Sphere(V(22, 0.8, 30), 1.0), Sphere(V(12, 0.8, 38), 1.0)) })
}
