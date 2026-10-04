package scenes

import (
	"math"

	"godocta/gen/kit"
	. "godocta/gen/octa"
)

// EileanDonan is the Highland castle on its tidal islet where three sea lochs meet.
// A tall rubble-stone tower house with corbelled bartizans, a parapet walk and
// crow-stepped gables rises from the north side of the rock; a curtain wall follows the
// islet's outline with a hornwork at the eastern tip and a round bastion to the south
// west, enclosing a small courtyard and lower crow-gabled ranges. A long, low,
// multi-arched stone footbridge curves north east from the gate to the mainland shore,
// where whitewashed cottages stand below conifers and heather hills that slope down to
// a low rim. The loch is a single water layer at y = 0. North is -Z.
type EileanDonan struct{}

func init() { Register(EileanDonan{}) }

func (EileanDonan) Name() string  { return "eilean_donan" }
func (EileanDonan) Title() string { return "Eilean Donan" }

const (
	eileanDonanPlateau = 4.0  // Islet top inside the curtain, in castle units.
	eileanDonanScale   = 1.15 // Castle units to world units (about the origin).
	eileanDonanCurtain = 10.0 // Curtain wall walkway height.
)

// eileanDonanIslet is the curtain wall line; the rock skirt spreads beyond it.
var eileanDonanIslet = []Vec2{
	P(-25, -3), P(-20, -12), P(-8, -15), P(4, -13), P(13, -9), P(21, -2),
	P(18, 6), P(8, 11), P(-6, 13), P(-18, 11), P(-25, 5),
}

var (
	eileanDonanCenter = P(-2, -1)
	eileanDonanGate   = P(17, -5.5)                                  // Gate on the north-east edge.
	eileanDonanBridge = [3]Vec2{P(20.6, -8), P(33, -16), P(46, -13)} // Quadratic Bézier, world units.
)

func eileanDonanSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func eileanDonanLerp(a, b, t float64) float64 { return a + (b-a)*t }

// eileanDonanNoise is smooth value noise in [-1, 1].
func eileanDonanNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return eileanDonanLerp(eileanDonanLerp(a, b, sx), eileanDonanLerp(c, d, sx), sz)*2 - 1
}

func eileanDonanFBM(x, z float64, seed int) float64 {
	return 0.6*eileanDonanNoise(x, z, seed) + 0.3*eileanDonanNoise(x*2.03+17, z*2.03+5, seed+1) +
		0.1*eileanDonanNoise(x*4.1+3, z*4.1+29, seed+2)
}

// eileanDonanLand memoizes terrain heights.
type eileanDonanLand struct {
	cache map[[2]int64]float64
}

// segment returns the plan distance from p to segment ab and the parameter along it.
func (EileanDonan) segment(p, a, b Vec2) (float64, float64) {
	ab := b.Sub(a)
	t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
	return p.Sub(a.Add(ab.Scale(t))).Length(), t
}

// inside is the signed distance to the islet polygon, positive inside.
func (s EileanDonan) inside(p Vec2) float64 {
	d := math.Inf(1)
	n := len(eileanDonanIslet)
	for i := 0; i < n; i++ {
		e, _ := s.segment(p, eileanDonanIslet[i], eileanDonanIslet[(i+1)%n])
		d = math.Min(d, e)
	}
	if InsidePolygon(eileanDonanIslet, p) {
		return d
	}
	return -d
}

// along is the distance along the nearest islet edge, for crenel spacing.
func (s EileanDonan) along(p Vec2) float64 {
	best, at := math.Inf(1), 0.0
	n := len(eileanDonanIslet)
	for i := 0; i < n; i++ {
		a, b := eileanDonanIslet[i], eileanDonanIslet[(i+1)%n]
		d, t := s.segment(p, a, b)
		if d < best {
			best, at = d, t*b.Sub(a).Length()
		}
	}
	return at
}

// inland is the distance from the shoreline into the mainland, corrected for its slope.
func (s EileanDonan) inland(x, z float64) float64 {
	d := (s.shore(x+0.5) - s.shore(x-0.5))
	return (s.shore(x) - z) / math.Sqrt(1+d*d)
}

// rim is the signed distance inside the rounded-rectangle site.
func (EileanDonan) rim(x, z float64) float64 {
	const cx, cz, hx, hz, r = 12, -18.5, 52, 42.5, 14
	dx, dz := math.Abs(x-cx)-(hx-r), math.Abs(z-cz)-(hz-r)
	out := math.Hypot(math.Max(dx, 0), math.Max(dz, 0)) + math.Min(math.Max(dx, dz), 0)
	return r - out
}

// shore is the mainland shoreline: land lies north (lower Z) of it.
func (EileanDonan) shore(x float64) float64 {
	return -25 + 19*eileanDonanSmooth((x-26)/30) + 1.6*eileanDonanFBM(x*0.09, 3, 51)
}

func (s EileanDonan) raw(x, z float64) float64 {
	rim := s.rim(x, z)
	if rim < 0 {
		return 0
	}
	// Islet: flat top inside the curtain, a craggy rock skirt beyond it.
	const k = eileanDonanScale
	d := k * s.inside(P(x/k, z/k))
	islet := 0.0
	if d >= 0 {
		islet = eileanDonanPlateau * k
	} else {
		n := 1.4*eileanDonanFBM(x*0.18, z*0.18, 7) + 0.5*eileanDonanNoise(x*0.6, z*0.6, 8)
		islet = eileanDonanPlateau*k*eileanDonanSmooth(1+(d+n)/4.2) + 0.4*eileanDonanNoise(x*0.5, z*0.5, 9)
	}
	// Mainland: a shore strip, then heather hills behind, falling to a low rim.
	e := s.inland(x, z)
	main := 2.3*eileanDonanSmooth(e/2.6) + 0.3*eileanDonanFBM(x*0.2, z*0.2, 12)
	hill := (15 + 6*eileanDonanFBM(x*0.035, z*0.05, 13)) * eileanDonanSmooth((e-6)/16) * (1 - 0.35*eileanDonanSmooth((x-30)/25))
	hill += 1.4 * eileanDonanFBM(x*0.12, z*0.12, 14) * eileanDonanSmooth((e-8)/6)
	main += hill * eileanDonanSmooth((rim+1)/13)
	if e < 0 {
		main = 2.3 * eileanDonanSmooth(1+e/2.5)
	}
	h := math.Max(islet, main)
	// Smooth minimum against the rim slope, so the cap leaves no crease.
	b := 0.6 + rim*0.9
	w := math.Max(6-math.Abs(h-b), 0) / 6
	return math.Max(math.Min(h, b)-w*w*1.5, math.Min(h, 0.3))
}

func (s EileanDonan) height(land *eileanDonanLand, x, z float64) float64 {
	key := [2]int64{int64(math.Round(x * 1000)), int64(math.Round(z * 1000))}
	if h, ok := land.cache[key]; ok {
		return h
	}
	h := s.raw(x, z)
	land.cache[key] = h
	return h
}

// shell fills the cells of a predicate within t of its outside (sideways or above).
func (EileanDonan) shell(g *Grid, box Box, t float64, in func(Vec3) bool) {
	g.Fill(Func{box, func(p Vec3) bool {
		if !in(p) {
			return false
		}
		return !(in(p.Add(V(t, 0, 0))) && in(p.Add(V(-t, 0, 0))) && in(p.Add(V(0, 0, t))) &&
			in(p.Add(V(0, 0, -t))) && in(p.Add(V(0, t, 0))))
	}})
}

// recolor visits every occupied cell once and repaints it when pick returns true.
func (EileanDonan) recolor(g *Grid, pick func(c Cell, p Vec3, col Color) (Color, bool)) {
	brush := g.Brush()
	for _, c := range g.Cells() {
		col, _ := g.ColorOf(c)
		if next, ok := pick(c, g.World(c), col); ok && next != col {
			g.Use(next)
			g.Remove(c)
			g.Add(c)
		}
	}
	g.Use(brush)
}

// gableEnd fills a crow-stepped (or plain) gable wall rising from eave over a wall
// slab: across is the half-width perpendicular to the ridge.
func (EileanDonan) gableEnd(g *Grid, slab Box, eave float64, ridge Axis, crow bool) float64 {
	c := slab.Center()
	half := (slab.Max.Z - slab.Min.Z) / 2
	if ridge == AxisZ {
		half = (slab.Max.X - slab.Min.X) / 2
	}
	const step = 1.2
	apex := eave + half
	g.Fill(Func{Box{V(slab.Min.X, eave, slab.Min.Z), V(slab.Max.X, apex+0.5, slab.Max.Z)}, func(p Vec3) bool {
		across := math.Abs(p.Z - c.Z)
		if ridge == AxisZ {
			across = math.Abs(p.X - c.X)
		}
		ext := half - (p.Y - eave)
		if crow {
			ext = half - math.Floor((p.Y-eave)/step)*step - 0.2
		}
		return across <= ext && ext >= 0.5
	}})
	return apex
}

// house is a hollow gabled range or cottage with small dark windows and chimneys.
// Long walls run along the ridge; gable ends are crow-stepped when crow is set.
func (s EileanDonan) house(g *Grid, x0, z0, x1, z1, base, eave float64, ridge Axis, crow bool, wall, roof Color, rows []float64, seed int) {
	const t = 1.2
	g.With(wall, func() {
		g.Fill(Block(V(x0, base-1.5, z0), V(x1, eave, z1)))
	})
	g.Carve(Block(V(x0+t, base+0.3, z0+t), V(x1-t, eave+20, z1-t)))
	// Gable ends.
	ends := []Box{{V(x0, 0, z0), V(x0+t, 0, z1)}, {V(x1-t, 0, z0), V(x1, 0, z1)}}
	if ridge == AxisZ {
		ends = []Box{{V(x0, 0, z0), V(x1, 0, z0+t)}, {V(x0, 0, z1-t), V(x1, 0, z1)}}
	}
	apex := 0.0
	g.With(wall, func() {
		for _, e := range ends {
			apex = s.gableEnd(g, e, eave, ridge, crow)
		}
	})
	// Dark inner faces so that the openings read.
	g.Paint(Charcoal, Block(V(x0+t-0.6, base-1.5, z0+t-0.6), V(x1-t+0.6, apex, z1-t+0.6)))
	// Windows on the long walls and one in each gable.
	for ri, y := range rows {
		if ridge == AxisX {
			for x := x0 + 2.4; x <= x1-2.2; x += 2.8 {
				if Hash01(int(x*2), ri, seed) < 0.15 {
					continue
				}
				for _, z := range []float64{z0, z1} {
					g.Carve(Block(V(x-0.5, y, z-1.5), V(x+0.5, y+1.4, z+1.5)))
				}
			}
		} else {
			for z := z0 + 2.4; z <= z1-2.2; z += 2.8 {
				if Hash01(int(z*2), ri, seed) < 0.15 {
					continue
				}
				for _, x := range []float64{x0, x1} {
					g.Carve(Block(V(x-1.5, y, z-0.5), V(x+1.5, y+1.4, z+0.5)))
				}
			}
		}
	}
	gy := eave + 0.8
	if ridge == AxisX {
		cz := (z0 + z1) / 2
		g.Carve(Block(V(x0-1, gy, cz-0.5), V(x0+1.6, gy+1.4, cz+0.5)), Block(V(x1-1.6, gy, cz-0.5), V(x1+1, gy+1.4, cz+0.5)))
	} else {
		cx := (x0 + x1) / 2
		g.Carve(Block(V(cx-0.5, gy, z0-1), V(cx+0.5, gy+1.4, z0+1.6)), Block(V(cx-0.5, gy, z1-1.6), V(cx+0.5, gy+1.4, z1+1)))
	}
	// Roof between the gables (crow-stepped) or overhanging them (plain).
	g.With(roof, func() {
		if ridge == AxisX {
			if crow {
				g.Fill(Gable{P(x0+t, z0-0.3), P(x1-t, z1+0.3), eave - 0.5, AxisX})
			} else {
				g.Fill(Gable{P(x0-0.4, z0-0.4), P(x1+0.4, z1+0.4), eave - 0.4, AxisX})
			}
		} else {
			if crow {
				g.Fill(Gable{P(x0-0.3, z0+t), P(x1+0.3, z1-t), eave - 0.5, AxisZ})
			} else {
				g.Fill(Gable{P(x0-0.4, z0-0.4), P(x1+0.4, z1+0.4), eave - 0.4, AxisZ})
			}
		}
	})
	// Chimney stacks on the gable apexes.
	for _, e := range ends {
		c := e.Center()
		g.With(wall, func() { g.Fill(BlockAt(P(c.X, c.Z), 1.6, 1.6, apex-1.5, apex+1.3)) })
		g.With(Charcoal, func() { g.Fill(BlockAt(P(c.X, c.Z), 1.0, 1.0, apex+1.3, apex+1.8)) })
	}
}

// bezier samples the bridge's center line.
func (EileanDonan) bezier(n int) []Vec2 {
	a, b, c := eileanDonanBridge[0], eileanDonanBridge[1], eileanDonanBridge[2]
	pts := make([]Vec2, n+1)
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		u := 1 - t
		pts[i] = a.Scale(u * u).Add(b.Scale(2 * u * t)).Add(c.Scale(t * t))
	}
	return pts
}

// polyline returns the distance to a polyline and the arc length of the nearest point.
func (s EileanDonan) polyline(p Vec2, pts []Vec2) (dist, along float64) {
	dist = math.Inf(1)
	run := 0.0
	for i := 0; i+1 < len(pts); i++ {
		d, t := s.segment(p, pts[i], pts[i+1])
		l := pts[i+1].Sub(pts[i]).Length()
		if d < dist {
			dist, along = d, run+t*l
		}
		run += l
	}
	return
}

func (s EileanDonan) Build(g *Grid) Order {
	land := &eileanDonanLand{cache: make(map[[2]int64]float64)}
	H := func(x, z float64) float64 { return s.height(land, x, z) }
	site := Box{V(-41, -0.3, -62), V(65, 30, 25)}

	// Loch: one water layer.
	g.Use(Water).Fill(Func{Box{V(-41, -0.2, -62), V(65, 0.2, 25)}, func(p Vec3) bool {
		return s.rim(p.X, p.Z) >= 0
	}})

	// Land: a hollow skin over the height field.
	inLand := func(p Vec3) bool {
		if p.Y < -0.2 {
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
	isletD := func(p Vec3) float64 { return eileanDonanScale * s.inside(p.Plan().Scale(1/eileanDonanScale)) }
	s.recolor(g, func(c Cell, p Vec3, col Color) (Color, bool) {
		if col != DarkStone {
			return 0, false
		}
		h := H(p.X, p.Z)
		top := p.Y > h-0.9
		e := s.shore(p.X) - p.Z
		hsh := Hash01(int(c.X), int(c.Y), int(c.Z), 3)
		onIslet := isletD(p) > -9 && e < -2
		if onIslet {
			if isletD(p) >= 0 { // Courtyard and wall-walk ground.
				if Hash01(int(c.X), int(c.Z), 4) < 0.3 {
					return DarkStone, true
				}
				return Grass, true
			}
			switch {
			case p.Y < 0.9: // Seaweed tide line.
				if hsh < 0.5 {
					return Charcoal, true
				}
				return Foliage, true
			case top && slope(p.X, p.Z) < 0.7 && p.Y > 2:
				return Grass, true
			case hsh < 0.12:
				return Stone, true
			case hsh < 0.22:
				return Charcoal, true
			}
			return DarkStone, true
		}
		// Mainland.
		if p.Y < 0.9 {
			if hsh < 0.45 {
				return Charcoal, true
			}
			return DarkStone, true
		}
		if !top {
			if hsh < 0.3 {
				return Stone, true
			}
			return DarkStone, true
		}
		sl := slope(p.X, p.Z)
		if sl > 1.25 {
			if hsh < 0.4 {
				return Stone, true
			}
			return DarkStone, true
		}
		if h < 1.4 {
			return Stone, true // Shingle.
		}
		if h > 4 {
			heather := eileanDonanFBM(p.X*0.08, p.Z*0.08, 61)
			switch {
			case hsh < 0.05+0.06*math.Max(heather, 0):
				return Red, true
			case hsh < 0.09+0.06*math.Max(heather, 0):
				return Terracotta, true
			case heather+0.3*eileanDonanNoise(p.X*0.4, p.Z*0.4, 62) > 0.05:
				return Foliage, true
			}
		}
		return Grass, true
	})

	// The castle is built in castle units on a finer grid, then merged cell for cell:
	// world = castle units × scale, on the same lattice.
	const k = eileanDonanScale
	cg := NewGridWithSpan(g.Span() / k)
	s.castle(cg, func(x, z float64) float64 { return H(x*k, z*k) / k })
	g.Merge(cg)
	// Paving from the gate across the courtyard.
	g.Paint(DarkStone, Func{Box{V(-8, 3, -8), V(22, eileanDonanPlateau*k+0.2, 7)}, func(p Vec3) bool {
		d, _ := s.segment(p.Plan(), eileanDonanGate.Scale(k), P(-3, 0).Scale(k))
		return d < 1.6 && p.Y > eileanDonanPlateau*k-0.9
	}})
	s.bridge(g, H)
	s.shoreline(g, H)

	// Rubble masonry: mottle the light stone with darker stones.
	castleBox := Box{V(-36, 0.5, -24), V(50, 40, 22)}
	s.recolor(g, func(c Cell, p Vec3, col Color) (Color, bool) {
		if col != Stone || !castleBox.Contains(p) {
			return 0, false
		}
		patch := Hash01(int(math.Floor(p.X/1.7)), int(math.Floor(p.Y/0.9)), int(math.Floor(p.Z/1.7)), 71)
		if Hash01(int(c.X), int(c.Y), int(c.Z), 72) < 0.08 || patch < 0.05 {
			return DarkStone, true
		}
		return 0, false
	})
	return Sweep(eileanDonanCenter, math.Pi*0.25)
}

// castle builds the tower house, curtain, bastion, hornwork, ranges and gate.
func (s EileanDonan) castle(g *Grid, H func(x, z float64) float64) {
	py := eileanDonanPlateau
	box := Box{V(-30, -0.5, -20), V(26, 14, 18)}
	horn := P(21, -2)

	// Curtain wall following the islet, higher around the hornwork, with a parapet.
	wallTop := func(p Vec3) float64 {
		return eileanDonanCurtain + 1.8*eileanDonanSmooth((7-p.Plan().Sub(horn).Length())/3)
	}
	g.Use(Stone).Fill(Func{box, func(p Vec3) bool {
		d := s.inside(p.Plan())
		if d < 0 || d > 1.4 {
			return false
		}
		top := wallTop(p)
		if p.Y < 0.2 || p.Y > top+1.3 {
			return false
		}
		if p.Y <= top {
			return true
		}
		if d > 0.7 {
			return false
		}
		// Crenels in the parapet.
		return p.Y < top+0.6 || math.Mod(s.along(p.Plan()), 2.4) < 1.4
	}})
	// Plinth course at the foot of the wall, where it meets the rock.
	g.With(DarkStone, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			d := s.inside(p.Plan())
			return d >= -0.6 && d < 0.4 && p.Y >= 0.2 && p.Y <= math.Max(H(p.X, p.Z)+0.6, 1)
		}})
	})
	// Arrow slits and a few small windows through the curtain.
	for i, v := range eileanDonanIslet {
		w := eileanDonanIslet[(i+1)%len(eileanDonanIslet)]
		dir := w.Sub(v)
		l := dir.Length()
		n := int(l / 4.5)
		for k := 1; k < n; k++ {
			q := v.Add(dir.Scale(float64(k) / float64(n)))
			if q.Sub(eileanDonanGate).Length() < 4 || (q.X > -18 && q.X < 0 && q.Y < -11) {
				continue
			}
			ang := math.Atan2(dir.Y, dir.X) + math.Pi/2
			slit := Translate(RotateY(Block(V(-1.5, 0, -0.45), V(1.5, 1.6, 0.45)), P(0, 0), ang), q.At(py+2.2))
			g.Carve(slit)
		}
	}
	// Gate: an arch through the curtain with a stepped landing.
	gateDir := P(7, -8).Scale(1 / math.Hypot(7, 8))
	ga := math.Atan2(gateDir.Y, gateDir.X)
	g.Carve(Translate(RotateY(Arch{V(0, 0, 0), 1.3, 2.0, 5, AxisX}, P(0, 0), ga), eileanDonanGate.At(py+0.3)))
	g.With(DarkStone, func() {
		g.Fill(Translate(RotateY(Block(V(-1, -0.3, -1.4), V(2.5, 0.3, 1.4)), P(0, 0), ga), eileanDonanGate.At(py)))
	})
	// Square gate tower inside the gate.
	gt := eileanDonanGate.Sub(gateDir.Scale(3.2))
	gtSolid := Translate(RotateY(Block(V(-2.4, 0, -3.2), V(2.4, 1, 3.2)), P(0, 0), ga), V(gt.X, 0, gt.Y))
	g.Fill(Func{Box{V(gt.X-5, py-1, gt.Y-5), V(gt.X+5, 14.5, gt.Y+5)}, func(p Vec3) bool {
		return gtSolid.Contains(V(p.X, 0.5, p.Z)) && p.Y <= 13
	}})
	g.Paint(Charcoal, Func{Box{V(gt.X-5, py-1, gt.Y-5), V(gt.X+5, 12.4, gt.Y+5)}, func(p Vec3) bool {
		return gtSolid.Contains(V(p.X, 0.5, p.Z).Add(gateDir.At(0).Scale(0.6))) && gtSolid.Contains(V(p.X, 0.5, p.Z).Sub(gateDir.At(0).Scale(0.6))) &&
			gtSolid.Contains(V(p.X, 0.5, p.Z).Add(P(gateDir.Y, -gateDir.X).At(0).Scale(0.6))) && gtSolid.Contains(V(p.X, 0.5, p.Z).Sub(P(gateDir.Y, -gateDir.X).At(0).Scale(0.6)))
	}})
	g.Carve(Translate(RotateY(Arch{V(0, 0, 0), 1.3, 2.0, 12, AxisX}, P(0, 0), ga), gt.At(py+0.3)))
	g.With(Slate, func() {
		g.Fill(Func{Box{V(gt.X-5, 13, gt.Y-5), V(gt.X+5, 17, gt.Y+5)}, func(p Vec3) bool {
			q := p.Plan().Sub(gt).Rotate(-ga)
			return math.Abs(q.X) <= 2.8-(p.Y-13) && math.Abs(q.Y) <= 3.6-(p.Y-13)
		}})
	})
	for _, k := range []float64{-1, 1} {
		w := gt.Add(P(gateDir.Y, -gateDir.X).Scale(3.21 * k)).Add(gateDir.Scale(-0.5))
		g.Carve(Translate(RotateY(Block(V(-0.5, 0, -1.5), V(0.5, 1.4, 1.5)), P(0, 0), ga+math.Pi/2), w.At(py+5.5)))
		g.Carve(Translate(RotateY(Block(V(-0.5, 0, -1.5), V(0.5, 1.4, 1.5)), P(0, 0), ga+math.Pi/2), w.At(py+8.5)))
	}

	// Round bastion at the south-west corner.
	g.Use(Stone).Build(kit.Tower{Center: P(-21.5, 7.5), Radius: 4.2, Base: 0.5, Top: 12, Storey: 5.5, Merlons: 14, Windows: true})
	g.Paint(Charcoal, Cylinder{Center: P(-21.5, 7.5), Radius: 3.0, Y0: 1, Y1: 10.5})

	// Lower ranges with crow-stepped gables around the courtyard.
	s.house(g, -19, 4.6, -5, 10.6, py, 11, AxisX, true, Stone, Slate, []float64{py + 2, py + 5}, 21)
	s.house(g, 1, -11, 12, -4.6, py, 10.5, AxisX, true, Stone, Slate, []float64{py + 2, py + 5}, 22)
	s.house(g, 2, 4.4, 9, 9.6, py, 8.5, AxisX, true, Stone, Slate, []float64{py + 2}, 23)

	s.keep(g)

	// Round stair turret on the keep's courtyard face, and watch turrets on the curtain.
	g.Use(Stone)
	st := P(-12.5, -2.0)
	g.Fill(Cylinder{Center: st, Radius: 2.0, Y0: py - 1, Y1: 26})
	g.Paint(Charcoal, Cylinder{Center: st, Radius: 1.2, Y0: py, Y1: 25.4})
	g.Carve(Cylinder{Center: st, Radius: 1.0, Y0: py + 0.3, Y1: 25.4})
	for k, y := range []float64{py + 4, py + 9, py + 14, py + 18.5} {
		a := 0.5 + float64(k)*0.9
		q := Polar(st, 1.8, a)
		g.Carve(BlockAt(q, 1.6, 1.6, y, y+1.4))
	}
	g.With(DarkStone, func() { g.Fill(Cylinder{Center: st, Radius: 2.4, Y0: 25.2, Y1: 26}) })
	g.With(Slate, func() { g.FillVolume(Cone(st, 2.4, 26, 4.2)) })
	for _, t := range []struct {
		c    Vec2
		y, r float64
	}{{P(20, -2), 11.8, 1.9}, {P(-20, -12), 10, 1.6}, {P(8, 11), 10, 1.5}, {P(4, -13), 10, 1.5}} {
		g.With(DarkStone, func() { g.FillVolume(Frustum{t.c, 0.6, t.r, t.y - 3, t.y - 0.5}) })
		g.Fill(Cylinder{Center: t.c, Radius: t.r, Y0: t.y - 0.5, Y1: t.y + 2.4})
		g.Carve(Cylinder{Center: t.c, Radius: t.r - 0.9, Y0: t.y, Y1: t.y + 3})
		g.Paint(Charcoal, Cylinder{Center: t.c, Radius: t.r - 0.4, Y0: t.y - 0.5, Y1: t.y + 2.4})
		for k := 0; k < 4; k++ {
			g.Carve(BlockAt(Polar(t.c, t.r-0.3, float64(k)*math.Pi/2+0.4), 0.8, 0.8, t.y+0.8, t.y+1.8))
		}
		g.With(Slate, func() { g.FillVolume(Cone(t.c, t.r+0.3, t.y+2.4, t.r*1.9)) })
	}

	// Well and paving in the courtyard.
	g.With(DarkStone, func() { g.Fill(Cylinder{Center: P(-1, 3), Radius: 1.2, Inner: 0.5, Y0: py, Y1: py + 1.2}) })
}

// keep is the tall tower house with parapet walks, bartizans and crow-stepped gables.
func (s EileanDonan) keep(g *Grid) {
	x0, x1, z0, z1 := -17.0, -1.0, -14.5, -3.0
	floor := eileanDonanPlateau
	top := 24.0
	const t = 1.4
	g.Use(Stone)
	g.Fill(Block(V(x0, 0.5, z0), V(x1, top, z1)))
	g.Carve(Block(V(x0+t, floor+0.3, z0+t), V(x1-t, top-0.6, z1-t)))
	// Corbelled parapets on the long sides; the walk sits on the wall top.
	for _, z := range []float64{z0, z1} {
		out := -1.0
		if z == z1 {
			out = 1
		}
		g.With(DarkStone, func() {
			g.Fill(Block(V(x0, top-1.5, z), V(x1, top-0.4, z+out*0.6)))
		})
		g.Fill(Func{Box{V(x0, top-0.5, math.Min(z, z+out*0.7)), V(x1, top+1.5, math.Max(z+out*0.7, z-out*0.2))}, func(p Vec3) bool {
			return p.Y < top+0.7 || math.Mod(p.X-x0+0.6, 2.4) < 1.4
		}})
	}
	// Crow-stepped gables at both ends.
	apex := 0.0
	for _, e := range []Box{{V(x0, 0, z0), V(x0+t, 0, z1)}, {V(x1-t, 0, z0), V(x1, 0, z1)}} {
		apex = s.gableEnd(g, e, top, AxisX, true)
	}
	// Dark interior.
	g.Paint(Charcoal, Block(V(x0+t-0.6, floor-1.5, z0+t-0.6), V(x1-t+0.6, top-0.7, z1-t+0.6)))
	g.Paint(Charcoal, Block(V(x0+t-0.6, top-0.6, z0+t-0.6), V(x1-t+0.6, apex, z1-t+0.6)))
	// Windows: small, rows per storey, larger on the hall floor.
	rows := []struct{ y, h float64 }{{floor + 2.5, 1.4}, {floor + 6.4, 2.4}, {floor + 10.8, 1.8}, {floor + 14.4, 1.6}, {floor + 17.4, 1.4}}
	for ri, r := range rows {
		for x := x0 + 2.8; x <= x1-2.6; x += 2.6 {
			for zi, z := range []float64{z0, z1} {
				if Hash01(int(x*4), ri, zi, 31) < 0.3 {
					continue
				}
				g.Carve(Block(V(x-0.5, r.y, z-2), V(x+0.5, r.y+r.h, z+2)))
			}
		}
		for zi, z := range []float64{(z0+z1)/2 - 2.4, (z0 + z1) / 2, (z0+z1)/2 + 2.4} {
			for xi, x := range []float64{x0, x1} {
				if Hash01(int(z*4), ri, xi, zi, 32) < 0.45 {
					continue
				}
				g.Carve(Block(V(x-2, r.y, z-0.5), V(x+2, r.y+r.h, z+0.5)))
			}
		}
	}
	// Gable windows above the roof line and the arched door to the courtyard.
	for _, x := range []float64{x0, x1} {
		g.Carve(Block(V(x-2, top+1.2, (z0+z1)/2-0.5), V(x+2, top+2.6, (z0+z1)/2+0.5)))
	}
	g.Carve(Arch{V(-6, floor+0.3, z1), 1.1, 1.8, 4, AxisZ})
	g.With(Wood, func() { g.Fill(Arch{V(-6, floor+0.3, z1-0.7), 1.1, 1.8, 0.5, AxisZ}) })
	// Forestair up to the door.
	g.Use(Stone)
	g.Fill(Block(V(-7.6, floor, z1), V(-4.4, floor+0.3, z1+2.4)))
	// Slate roof between the gables, behind the parapets.
	g.With(Slate, func() {
		g.Fill(Gable{P(x0+t, z0+0.7), P(x1-t, z1-0.7), top - 0.2, AxisX})
	})
	g.Carve(Gable{P(x0+t, z0+2.2), P(x1-t, z1-2.2), top - 0.6, AxisX})
	// Dormers on the south slope.
	for _, x := range []float64{-12, -6} {
		g.Fill(Block(V(x-1.2, top, z1-3.4), V(x+1.2, top+2.6, z1-2.0)))
		g.Carve(Block(V(x-0.5, top+0.6, z1-3.6), V(x+0.5, top+1.9, z1-1.5)))
		g.With(Slate, func() { g.Fill(Gable{P(x-1.6, z1-5.5), P(x+1.6, z1-1.6), top + 2.4, AxisZ}) })
	}
	// Chimneys on the gable apexes.
	for _, x := range []float64{x0 + t/2, x1 - t/2} {
		g.Fill(BlockAt(P(x, (z0+z1)/2), 1.8, 2.2, apex-1.5, apex+1.6))
		g.With(Charcoal, func() { g.Fill(BlockAt(P(x, (z0+z1)/2), 1.2, 1.6, apex+1.6, apex+2.1)) })
	}
	// Corbelled bartizans with conical slate caps at the four corners.
	for _, x := range []float64{x0 - 0.3, x1 + 0.3} {
		for _, z := range []float64{z0 - 0.3, z1 + 0.3} {
			c := P(x, z)
			g.With(DarkStone, func() { g.FillVolume(Frustum{c, 0.6, 1.6, top - 4.5, top - 1.5}) })
			g.Fill(Cylinder{Center: c, Radius: 1.7, Y0: top - 1.5, Y1: top + 2.2})
			g.Carve(Cylinder{Center: c, Radius: 0.8, Y0: top - 0.5, Y1: top + 2.5})
			for k := 0; k < 4; k++ {
				a := float64(k)*math.Pi/2 + math.Pi/4
				q := Polar(c, 1.4, a)
				g.Carve(BlockAt(q, 0.9, 0.9, top+0.4, top+1.6))
			}
			g.With(Slate, func() { g.FillVolume(Cone(c, 2.0, top+2.2, 3.4)) })
		}
	}
}

// bridge is the long arched footbridge from the gate to the mainland.
func (s EileanDonan) bridge(g *Grid, H func(x, z float64) float64) {
	pts := s.bezier(160)
	total := 0.0
	for i := 0; i+1 < len(pts); i++ {
		total += pts[i+1].Sub(pts[i]).Length()
	}
	const half = 1.8
	const arches = 5
	const abut, pier = 3.5, 1.6
	spanL := (total - 2*abut - float64(arches-1)*pier) / arches
	deck := func(along float64) float64 {
		return eileanDonanLerp(eileanDonanPlateau*eileanDonanScale+0.4, 2.9, along/total)
	}
	underside := func(along float64) float64 {
		local := along - abut
		if local < 0 || local > total-2*abut {
			return -1
		}
		k := math.Floor(local / (spanL + pier))
		u := local - k*(spanL+pier)
		if u > spanL {
			return -1 // Pier.
		}
		a := spanL / 2
		x := (u - a) / a
		rise := math.Min(2.6, deck(along)-1.3)
		return 0.1 + rise*math.Sqrt(math.Max(0, 1-x*x))
	}
	b := EmptyBox()
	for _, p := range pts {
		b = b.Union(Box{p.At(-0.3), p.At(6)})
	}
	b = b.Grow(3)
	g.Use(Stone).Fill(Func{b, func(p Vec3) bool {
		d, along := s.polyline(p.Plan(), pts)
		top := deck(along)
		if d > half+0.6 || p.Y < -0.2 {
			return false
		}
		if along <= 0.01 || along >= total-0.01 {
			if d > half {
				return false
			}
		}
		// Cutwaters at the piers, pointed up and down stream.
		local := along - abut
		k := math.Floor(local / (spanL + pier))
		pc := abut + k*(spanL+pier) + spanL + pier/2
		if local > 0 && k < arches-1 && math.Abs(along-pc) < pier/2+0.8 && p.Y < 2.2 {
			if d+math.Abs(along-pc)*1.2 < half+0.9 {
				return true
			}
		}
		if d > half {
			return false
		}
		if p.Y <= top {
			return p.Y >= underside(along)
		}
		// Parapets.
		return d > half-0.6 && p.Y <= top+1.1
	}})
	// Coping and the deck path in darker stone.
	g.Paint(DarkStone, Func{b, func(p Vec3) bool {
		d, along := s.polyline(p.Plan(), pts)
		top := deck(along)
		return d <= half && ((d > half-0.6 && p.Y > top+0.6 && p.Y <= top+1.1) || (d < half-0.6 && p.Y > top-0.5 && p.Y <= top))
	}})
	g.Paint(Charcoal, Func{b, func(p Vec3) bool {
		d, along := s.polyline(p.Plan(), pts)
		u := underside(along)
		return d <= half && u > 0 && p.Y >= u-0.1 && p.Y < u+0.7
	}})
}

// shoreline adds the cottages, a road, conifers and boulders on the mainland.
func (s EileanDonan) shoreline(g *Grid, H func(x, z float64) float64) {
	type cottage struct {
		x, w, d float64
		crow    bool
	}
	var spots []Vec2
	for i, c := range []cottage{{28, 9, 5, false}, {37, 7, 5, true}, {54, 10, 5.5, false}, {-16, 8, 5, false}, {-3, 7, 5, false}, {9, 8, 5, false}} {
		z := s.shore(c.x) - 4.4
		h := math.Max(math.Max(H(c.x-c.w/2, z-c.d/2), H(c.x+c.w/2, z+c.d/2)), math.Max(H(c.x-c.w/2, z+c.d/2), H(c.x+c.w/2, z-c.d/2)))
		h = math.Round(h*2) / 2
		roof := Slate
		walls := White
		if i == 1 {
			walls = Stone
		}
		s.house(g, c.x-c.w/2, z-c.d/2, c.x+c.w/2, z+c.d/2, h, h+3.6, AxisX, c.crow, walls, roof, []float64{h + 1}, 40+i)
		// Door on the loch side.
		g.Carve(Block(V(c.x-0.6, h+0.3, z+c.d/2-1.5), V(c.x+0.6, h+2.4, z+c.d/2+1.5)))
		g.With(Wood, func() { g.Fill(Block(V(c.x-0.6, h+0.3, z+c.d/2-0.9), V(c.x+0.6, h+2.4, z+c.d/2-0.6))) })
		spots = append(spots, P(c.x, z))
	}
	// Shore road from the bridge along the strip.
	g.Paint(DarkStone, Func{Box{V(-41, 0.5, -62), V(65, 6, 25)}, func(p Vec3) bool {
		e := s.shore(p.X) - p.Z
		h := H(p.X, p.Z)
		return math.Abs(e-1.9) < 0.9 && p.Y > h-0.9 && p.X > -30 && p.X < 62 && h > 1.3
	}})
	// Conifer plantations on the lower hill slopes.
	for i := -36; i <= 62; i += 3 {
		for k := -58; k <= -6; k += 3 {
			x, z := float64(i)+2*Hash01(i, k, 81), float64(k)+2*Hash01(i, k, 82)
			e := s.inland(x, z)
			if e < 6.5 || e > 30 || s.rim(x, z) < 5 {
				continue
			}
			if f := eileanDonanFBM(x*0.07, z*0.07, 83); (f < -0.02 && Hash01(i, k, 87) > 0.07) || Hash01(i, k, 84) > 0.78 {
				continue
			}
			near := false
			for _, q := range spots {
				if q.Sub(P(x, z)).Length() < 7 {
					near = true
				}
			}
			if near {
				continue
			}
			h := H(x, z)
			tall := 5 + 3*Hash01(i, k, 85)
			r := 1.4 + 0.5*Hash01(i, k, 86)
			c := P(x, z)
			g.With(Wood, func() { g.Fill(Cylinder{Center: c, Radius: 0.5, Y0: h - 0.6, Y1: h + 1.2}) })
			g.With(Foliage, func() {
				g.FillVolume(Frustum{c, r, r * 0.45, h + 1, h + 1 + tall*0.55})
				g.FillVolume(Cone(c, r*0.8, h+tall*0.5, tall*0.55))
			})
		}
	}
	// A timber jetty below the cottages with a moored boat, and two boats on the loch.
	jx := 22.0
	jz := s.shore(jx)
	g.With(Wood, func() {
		g.Fill(Block(V(jx-1, 1.2, jz-2), V(jx+1, 1.6, jz+7)))
		for _, z := range []float64{jz + 1, jz + 3.5, jz + 6.5} {
			g.Fill(Cylinder{Center: P(jx-0.9, z), Radius: 0.45, Y0: -0.2, Y1: 2.2}, Cylinder{Center: P(jx+0.9, z), Radius: 0.45, Y0: -0.2, Y1: 2.2})
		}
	})
	s.boat(g, P(jx+3, jz+4.5), math.Pi/2, 6, 2.4, Red)
	s.boat(g, P(-31, 18), 0.5, 5, 2.2, White)
	s.boat(g, P(33, 4), -0.3, 6.5, 2.6, Navy)
	// A few boulders along the shore and around the islet.
	g.Use(DarkStone)
	for i := 0; i < 22; i++ {
		var c Vec2
		if i < 12 {
			a := Hash01(i, 91) * 2 * math.Pi
			c = Polar(eileanDonanCenter, 29+6*Hash01(i, 92), a)
			c = P(c.X, eileanDonanCenter.Y+(c.Y-eileanDonanCenter.Y)*0.75)
		} else {
			x := -36 + 96*Hash01(i, 93)
			c = P(x, s.shore(x)+0.5+1.5*Hash01(i, 94))
		}
		if s.rim(c.X, c.Y) < 2 || H(c.X, c.Y) > 1.5 {
			continue
		}
		r := 0.9 + 0.9*Hash01(i, 95)
		g.FillVolume(Ellipsoid{c.At(0.2), V(r*1.2, r*0.8, r)})
		if Hash01(i, 96) < 0.5 {
			g.Paint(Charcoal, Ellipsoid{c.At(0.2), V(r*1.3, 0.6, r*1.1)})
		}
	}
}

// boat is a small open rowing boat on the loch, bow toward angle.
func (EileanDonan) boat(g *Grid, c Vec2, angle, length, beam float64, band Color) {
	box := Box{V(c.X-length, -0.3, c.Y-length), V(c.X+length, 1.6, c.Y+length)}
	width := func(p Vec3, shrink float64) bool {
		q := p.Plan().Sub(c).Rotate(-angle)
		u := q.X / (length / 2)
		if math.Abs(u) > 1 {
			return false
		}
		w := beam / 2 * math.Sqrt(1-u*u*u*u) * (0.7 + 0.3*math.Min(1, p.Y+0.3))
		return math.Abs(q.Y) <= w-shrink
	}
	g.With(Wood, func() {
		g.FillVolume(Func{box, func(p Vec3) bool { return p.Y >= -0.2 && p.Y <= 1.2 && width(p, 0) }})
	})
	g.Carve(Func{box, func(p Vec3) bool { return p.Y > 0.5 && width(p, 0.8) }})
	g.Paint(band, Func{box, func(p Vec3) bool { return p.Y > 0.7 && p.Y <= 1.2 && width(p, 0) && !width(p, 0.7) }})
}
