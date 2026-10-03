package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// PiazzaDeiMiracoli is Pisa's Square of Miracles on its lawn: the Romanesque cathedral
// (a Latin cross with striped flanks, a four-tier arcaded façade, an elliptical dome over
// the crossing and apses on the choir and both transept arms), the round baptistery
// (blind arcade, open loggia, gables and pinnacles, a half-tiled, half-leaded dome
// pierced by its cone), the leaning tower (six open galleries round the cella and the
// bell chamber, tilted about four degrees southwards), the long Camposanto cloister and
// the city wall with its gate. The façade looks east (+X) at the baptistery; the
// layout is mirrored from the real square so the default views catch the façade.
type PiazzaDeiMiracoli struct{}

func init() { Register(PiazzaDeiMiracoli{}) }

func (PiazzaDeiMiracoli) Name() string  { return "piazza_dei_miracoli" }
func (PiazzaDeiMiracoli) Title() string { return "Piazza Dei Miracoli" }

const (
	miracoliBase    = 0.6   // Bottom of every building, on the two-layer lawn.
	miracoliFacadeX = 30.0  // Cathedral façade plane (its outer face).
	miracoliApseX   = -30.0 // Choir end, where the main apse begins.
	miracoliHalfW   = 11.0  // Half width of nave and aisles.
	miracoliNaveW   = 5.5   // Half width of the nave.
	miracoliTransX  = -13.0 // Transept axis.
	miracoliTransW  = 6.0   // Half width of a transept arm.
	miracoliTransZ  = 20.0  // Transept arm ends, where their apses begin.
	miracoliLean    = 0.08  // Tower tilt in radians (about 4.6°).
)

var (
	miracoliBaptistery = P(54, 0)
	miracoliTower      = P(-46, 16)
)

// duomoTop is the roof height of the cathedral's envelope at plan (x, z) when every
// wall is moved inwards by t, and the eave of the part that is highest there.
func (PiazzaDeiMiracoli) duomoTop(x, z, t float64) (top, eave float64, ok bool) {
	a := math.Abs(z)
	consider := func(h, e float64) {
		if !ok || h > top {
			top, eave, ok = h, e, true
		}
	}
	if x >= miracoliApseX+t && x <= miracoliFacadeX-t && a <= miracoliHalfW-t {
		if b := a + t; b <= miracoliNaveW {
			consider(19+(miracoliNaveW-b)*0.6-t, 19)
		} else {
			consider(11.5+(miracoliHalfW-b)*0.636-t, 11.5)
		}
	}
	if ax := math.Abs(x - miracoliTransX); ax <= miracoliTransW-t && a <= miracoliTransZ-t {
		consider(15.5+(miracoliTransW-ax-t)*0.6-t, 15.5)
	}
	if x <= miracoliApseX+t {
		if d := math.Hypot(x-miracoliApseX, z); d <= miracoliNaveW-t {
			consider(17+(miracoliNaveW-d-t)*0.6-t, 17)
		}
	}
	if a >= miracoliTransZ-t {
		if d := math.Hypot(x-miracoliTransX, a-miracoliTransZ); d <= 4.5-t {
			consider(13+(4.5-d-t)*0.6-t, 13)
		}
	}
	return
}

// facadeTop is the façade screen's raked outline: a gable over the nave and sloping
// shoulders over the aisles.
func (PiazzaDeiMiracoli) facadeTop(a float64) float64 {
	if a <= miracoliNaveW {
		return 23.4 - a*0.6
	}
	return 16.6 - (a-miracoliNaveW)*0.6
}

// arcade is a ring of n round-headed openings between radii r0 and r1, centred on the
// angles (k + phase) × 2π/n; widths are measured at the mean radius, so the piers keep
// a constant width through the ring.
func (PiazzaDeiMiracoli) arcade(c Vec2, r0, r1 float64, n int, phase, hw, sill, spring float64) Func {
	bay := 2 * math.Pi / float64(n)
	rm := (r0 + r1) / 2
	return Func{Box{V(c.X-r1, sill, c.Y-r1), V(c.X+r1, sill+spring+hw, c.Y+r1)}, func(p Vec3) bool {
		d := p.Plan().Sub(c)
		r := d.Length()
		if r < r0 || r > r1 {
			return false
		}
		f := d.Angle()/bay - phase
		u := (f - math.Round(f)) * bay * rm
		if math.Abs(u) > hw {
			return false
		}
		rise := p.Y - sill - spring
		return rise <= 0 || u*u+rise*rise <= hw*hw
	}}
}

// ribs selects n thin meridians (angular half-width w at radius 1) of a round shape.
func (PiazzaDeiMiracoli) ribs(c Vec2, n int, w float64, box Box) Func {
	bay := 2 * math.Pi / float64(n)
	return Func{box, func(p Vec3) bool {
		d := p.Plan().Sub(c)
		f := d.Angle() / bay
		return math.Abs(f-math.Round(f))*bay*d.Length() <= w
	}}
}

// gableSlab is a roof slab of the given vertical thickness over a rectangle, pitched at
// rise per run from base, its ridge along the given axis.
func (PiazzaDeiMiracoli) gableSlab(min, max Vec2, base, pitch, thick float64, ridge Axis) Func {
	half := (max.Y - min.Y) / 2
	if ridge == AxisZ {
		half = (max.X - min.X) / 2
	}
	return Func{Box{min.At(base), max.At(base + half*pitch)}, func(p Vec3) bool {
		across := math.Abs(p.Z - (min.Y+max.Y)/2)
		if ridge == AxisZ {
			across = math.Abs(p.X - (min.X+max.X)/2)
		}
		top := base + (half-across)*pitch
		return p.Y <= top && p.Y >= top-thick
	}}
}

// tilt leans a solid about a horizontal axis through pivot: its top moves towards the
// plan direction dir (a unit vector) by angle radians.
func (PiazzaDeiMiracoli) tilt(s Solid, pivot Vec3, dir Vec2, angle float64) Solid {
	sn, cs := math.Sincos(angle)
	side := P(-dir.Y, dir.X)
	// rotate turns a point by ±angle in the vertical plane through pivot along dir.
	rotate := func(p Vec3, sn float64) Vec3 {
		d := p.Sub(pivot)
		u, v := d.Plan().Dot(dir), d.Plan().Dot(side)
		nu, ny := u*cs+d.Y*sn, d.Y*cs-u*sn
		q := dir.Scale(nu).Add(side.Scale(v))
		return V(pivot.X+q.X, pivot.Y+ny, pivot.Z+q.Y)
	}
	b := s.Bounds()
	out := EmptyBox()
	for _, x := range []float64{b.Min.X, b.Max.X} {
		for _, y := range []float64{b.Min.Y, b.Max.Y} {
			for _, z := range []float64{b.Min.Z, b.Max.Z} {
				q := rotate(V(x, y, z), sn)
				out = out.Union(Box{q, q})
			}
		}
	}
	return Func{out.Grow(0.5), func(p Vec3) bool { return s.Contains(rotate(p, -sn)) }}
}

// ground lays the lawn, the paved paths and the marble aprons round each building.
func (m PiazzaDeiMiracoli) ground(g *Grid) {
	const x0, x1, z0, z1 = -57.0, 73.0, -58.0, 31.0
	layer := func(ax, az, bx, bz float64) Cuboid { return Block(V(ax, -0.1, az), V(bx, 0.45, bz)) }
	g.Use(Grass).Fill(layer(x0, z0, x1, z1))
	g.Paint(Sand,
		layer(x0, 26.5, x1, 29.2),   // The road along the south edge.
		layer(30, -2.5, 42, 2.5),    // Façade to baptistery.
		layer(58.5, -54, 61.5, -11), // From the gate to the baptistery.
		layer(52.5, 12, 55.5, 26.5), // Baptistery to the road.
		layer(-47.5, 22, -44.5, 26.5),
		layer(-6, 22, -3, 26.5), // Transept door to the road.
		layer(-6, -28, -3, -22), // Transept door to the Camposanto.
		layer(38.5, -28, 41.5, -14),
	)
	// Raised white marble aprons, one layer above the lawn.
	apron := func(p Vec3) bool {
		if _, _, ok := m.duomoTop(p.X, p.Z, -1.6); ok {
			if _, _, in := m.duomoTop(p.X, p.Z, 0); !in {
				return true
			}
		}
		return p.X >= miracoliFacadeX && p.X <= 35.5 && math.Abs(p.Z) <= 13
	}
	g.Use(White).Fill(Func{Box{V(-38, 0.6, -28), V(36, 0.85, 28)}, func(p Vec3) bool { return apron(p) }})
	g.Fill(Cylinder{Center: miracoliBaptistery, Radius: 13.6, Inner: 11.6, Y0: 0.6, Y1: 0.85})
	g.Fill(Cylinder{Center: miracoliTower, Radius: 9, Inner: 7, Y0: 0.6, Y1: 0.85})
	// No lawn under the closed interiors.
	g.Carve(Func{Box{V(-38, -0.2, -28), V(31, 0.5, 28)}, func(p Vec3) bool {
		_, _, ok := m.duomoTop(p.X, p.Z, 1.2)
		return ok
	}})
	g.Carve(Cylinder{Center: miracoliBaptistery, Radius: 9.6, Y0: -0.2, Y1: 0.5})
	g.Carve(Cylinder{Center: miracoliTower, Radius: 4, Y0: -0.2, Y1: 0.5})
}

// cathedral builds the Duomo: one striped marble shell for nave, aisles, transept and
// apses, then roofs, blind arcades, windows, the crossing dome and the façade.
func (m PiazzaDeiMiracoli) cathedral(g *Grid) {
	const base = miracoliBase
	box := Box{V(-36, base, -25), V(miracoliFacadeX, 24, 25)}
	g.Use(White).FillVolume(Func{box, func(p Vec3) bool {
		top, _, ok := m.duomoTop(p.X, p.Z, 0)
		if !ok || p.Y > top {
			return false
		}
		inner, _, in := m.duomoTop(p.X, p.Z, 1.2)
		return !in || p.Y > inner
	}})
	// Dark marble bands on the walls, then roofs: lead over nave, transept and apses,
	// terracotta over the aisles.
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		_, eave, ok := m.duomoTop(p.X, p.Z, 0)
		return ok && p.Y < eave-0.6 && p.X < miracoliFacadeX-1.3 && math.Mod(p.Y-base, 2.0) < 0.45
	}})
	g.Paint(Slate, Func{box, func(p Vec3) bool {
		top, eave, ok := m.duomoTop(p.X, p.Z, 0)
		return ok && p.Y > eave+0.05 && p.Y > top-1.8 && p.X < miracoliFacadeX-1.3
	}})
	g.Paint(Terracotta, Func{box, func(p Vec3) bool {
		top, eave, ok := m.duomoTop(p.X, p.Z, 0)
		return ok && eave == 11.5 && p.Y > eave+0.05 && p.Y > top-1.8 && p.X < miracoliFacadeX-1.3
	}})

	// Flank blind arcades on the aisles, with a window in every bay.
	var piers []float64
	for x := 28.4; x > miracoliApseX; x -= 3.6 {
		if x < miracoliTransX-miracoliTransW-0.6 || x > miracoliTransX+miracoliTransW+0.6 {
			piers = append(piers, x)
		}
	}
	for _, s := range []float64{-1, 1} {
		face := s * miracoliHalfW
		for i, x := range piers {
			g.Fill(Block(V(x-0.45, base, face), V(x+0.45, 11.4, face+s*0.7)))
			if i == 0 || piers[i-1]-x > 3.7 {
				continue
			}
			mid := x + 1.8
			g.Fill(Block(V(x, 8.8, face), V(x+3.6, 11.4, face+s*0.7)))
			g.Carve(Arch{V(mid, base+0.5, face+s*0.4), 1.3, 10.4 - 1.3 - base - 0.5, 0.6, AxisZ})
			g.Carve(Arch{V(mid, 4, face-s*0.6), 0.55, 2.6, 3, AxisZ})
			g.Carve(Arch{V(mid, 15.4, s*4.9), 0.55, 1.7, 3, AxisZ}) // Clerestory.
		}
	}
	// Transept side walls: pilasters and windows; arm apses with blind arcades.
	for _, s := range []float64{-1, 1} {
		for _, side := range []float64{-1, 1} {
			face := miracoliTransX + side*miracoliTransW
			for _, z := range []float64{11.6, 15.8, 19.6} {
				g.Fill(Block(V(face, base, s*z-0.45), V(face+side*0.7, 14.9, s*z+0.45)))
			}
			for _, z := range []float64{13.7, 17.7} {
				g.Carve(Arch{V(face-side*0.6, 4.5, s*z), 0.55, 3.2, 3, AxisX})
				g.Carve(Arch{V(face-side*0.6, 10, s*z), 0.55, 2.2, 3, AxisX})
			}
		}
		c := P(miracoliTransX, s*miracoliTransZ)
		half := Func{Box{V(-20, base, -26), V(-6, 13, 26)}, func(p Vec3) bool { return p.Z*s >= miracoliTransZ }}
		g.Fill(Intersect(Cylinder{Center: c, Radius: 5.1, Inner: 4.3, Y0: base, Y1: 12.4}, half))
		g.Carve(Intersect(m.arcade(c, 4.45, 5.2, 10, 0.5, 0.8, base, 9.2), half))
		g.Carve(Intersect(m.arcade(c, 3.0, 4.7, 10, 0.5, 0.5, 5.5, 2.2), half))
	}
	// Main apse: blind arcade and an upper gallery of small windows.
	apse := P(miracoliApseX, 0)
	behind := Func{Box{V(-37, base, -7), V(miracoliApseX, 17, 7)}, func(p Vec3) bool { return p.X <= miracoliApseX }}
	g.Fill(Intersect(Cylinder{Center: apse, Radius: 6.2, Inner: 5.3, Y0: base, Y1: 13.4}, behind))
	g.Carve(Intersect(m.arcade(apse, 5.45, 6.3, 14, 0.5, 0.85, base, 10.3), behind))
	g.Carve(Intersect(m.arcade(apse, 3.8, 5.7, 14, 0.5, 0.5, 14, 1.3), behind))
	g.Carve(Intersect(m.arcade(apse, 3.8, 5.7, 7, 0.5, 0.55, 4.5, 3), behind))

	// Crossing dome: windowed drum, lead dome with marble ribs, lantern.
	cross := P(miracoliTransX, 0)
	g.Use(White)
	g.FillVolume(Cylinder{Center: cross, Radius: 5.3, Inner: 4.0, Y0: 18.5, Y1: 25})
	g.Paint(DarkStone, Func{Box{V(-19, 18.5, -6), V(-7, 25, 6)}, func(p Vec3) bool { return math.Mod(p.Y-base, 2.0) < 0.45 }})
	g.Carve(m.arcade(cross, 3.8, 5.5, 12, 0, 0.6, 21.6, 1.5))
	g.FillVolume(Cylinder{Center: cross, Radius: 5.9, Inner: 4.0, Y0: 24.4, Y1: 25.1})
	dome := Intersect(Subtract(Ellipsoid{cross.At(25), V(5.6, 6.2, 5.6)}, Ellipsoid{cross.At(25), V(4.4, 5.0, 4.4)}),
		Block(V(-19, 25, -6), V(-7, 32, 6)))
	g.With(Slate, func() { g.FillVolume(dome) })
	g.Paint(White, Intersect(dome, m.ribs(cross, 8, 0.07, dome.Bounds())))
	g.FillVolume(Cylinder{Center: cross, Radius: 1.4, Y0: 30.4, Y1: 32.4})
	g.With(Gold, func() {
		g.FillVolume(Cone(cross, 1.6, 32.4, 1.6))
		g.Fill(Cylinder{Center: cross, Radius: 0.5, Y0: 33.5, Y1: 35})
	})

	m.facade(g)
}

// facade builds the west front (here facing +X): blind arcade with three bronze portals,
// four tiers of open loggias stepping with the roofline, a raked coping and statues.
func (m PiazzaDeiMiracoli) facade(g *Grid) {
	const base = miracoliBase
	const back, front = miracoliFacadeX - 1.2, miracoliFacadeX
	g.Use(White)
	g.Fill(Func{Box{V(back, base, -11), V(front, 24, 11)}, func(p Vec3) bool {
		return math.Abs(p.Z) <= miracoliHalfW && p.Y <= m.facadeTop(math.Abs(p.Z))
	}})
	// Ground order: seven blind arches on pilasters.
	g.Fill(Block(V(front, base, -11), V(front+0.9, 8.6, 11)))
	bay := 2 * miracoliHalfW / 7
	for i := 0; i < 7; i++ {
		z := -miracoliHalfW + (float64(i)+0.5)*bay
		g.Carve(Arch{V(front+0.5, base+0.5, z), 1.1, 4.9, 0.8, AxisX})
		g.Paint(Stone, Arch{V(front-0.2, base+0.4, z), 1.1, 5.4, 0.6, AxisX})
	}
	for _, d := range []struct{ z, hw, spring float64 }{{0, 1.6, 4.6}, {-3 * bay, 1.1, 3.6}, {3 * bay, 1.1, 3.6}} {
		g.Carve(Arch{V(front, base+0.5, d.z), d.hw, d.spring - 0.5, 3.4, AxisX})
		g.With(Wood, func() { g.Fill(Arch{V(back+0.3, base, d.z), d.hw, d.spring, 0.5, AxisX}) })
	}
	// Loggias: a ledge, a band of columns and arches, the wall behind in shadow tone.
	tiers := []struct{ y0, y1, half float64 }{{9.3, 12.4, 11}, {13.1, 16, 11}, {16.7, 19.4, 5.5}, {20.1, 23, 5.5}}
	for _, t := range tiers {
		t := t
		top := func(a float64) float64 { return math.Min(t.y1, m.facadeTop(a)-0.7) }
		g.Fill(Func{Box{V(back, t.y0-0.7, -t.half), V(front+2.2, t.y0, t.half)}, func(p Vec3) bool {
			return math.Abs(p.Z) <= t.half && p.Y <= m.facadeTop(math.Abs(p.Z))
		}})
		g.Fill(Func{Box{V(front+1, t.y0, -t.half), V(front+2, t.y1, t.half)}, func(p Vec3) bool {
			return math.Abs(p.Z) <= t.half && p.Y <= top(math.Abs(p.Z))
		}})
		g.Paint(Stone, Func{Box{V(front-0.6, t.y0, -t.half), V(front+0.05, t.y1, t.half)}, func(p Vec3) bool {
			return math.Abs(p.Z) <= t.half-0.4 && p.Y <= top(math.Abs(p.Z))
		}})
		n := math.Round(2 * t.half / 2.4)
		sp := 2 * t.half / n
		hw := (sp - 1.0) / 2
		for k := 0.0; k < n; k++ {
			z := -t.half + (k+0.5)*sp
			spring := top(math.Abs(z)+hw) - 0.7 - hw - t.y0
			if spring < 0.4 {
				continue
			}
			g.Carve(Arch{V(front+1.5, t.y0, z), hw, spring, 1.4, AxisX})
		}
	}
	// Raked coping along the outline, then statues on the apex and the shoulders.
	g.Fill(Func{Box{V(back, 12, -11.3), V(front+2.2, 24, 11.3)}, func(p Vec3) bool {
		top := m.facadeTop(math.Min(math.Abs(p.Z), miracoliHalfW))
		return math.Abs(p.Z) <= 11.3 && p.Y <= top && p.Y >= top-0.7
	}})
	g.With(Gold, func() { g.Fill(Capsule{V(front, 23.6, 0), V(front, 25.4, 0), 0.6}) })
	for _, s := range []float64{-1, 1} {
		for _, a := range []float64{miracoliNaveW, miracoliHalfW} {
			y := m.facadeTop(a)
			if a == miracoliNaveW {
				y = m.facadeTop(a + 0.01)
			}
			g.Fill(Capsule{V(front, y+0.2, s*a), V(front, y+1.5, s*a), 0.55})
		}
	}
}

// tower builds the leaning campanile: a plinth, the blind-arcaded ground storey, six
// open galleries around the cella, the bell chamber, all tilted together.
func (m PiazzaDeiMiracoli) tower(g *Grid) {
	c := miracoliTower
	const y0 = 1.0
	dir := P(-0.6, 0.8)
	lean := func(s Solid) Solid { return m.tilt(s, c.At(y0), dir, miracoliLean) }
	g.Use(White)
	g.FillVolume(Cylinder{Center: c, Radius: 7.2, Inner: 4.0, Y0: miracoliBase, Y1: 1.5})
	// Ground storey: a stone wall behind fifteen engaged columns and arches.
	g.FillVolume(lean(Cylinder{Center: c, Radius: 5.6, Inner: 4.3, Y0: y0, Y1: 10.4}))
	g.FillVolume(lean(Cylinder{Center: c, Radius: 6.3, Inner: 5.5, Y0: y0, Y1: 10.4}))
	g.Carve(lean(m.arcade(c, 5.62, 6.4, 15, 0.5, 0.75, y0, 6.8)))
	g.Paint(Stone, lean(m.arcade(c, 4.9, 5.7, 15, 0.5, 0.75, y0+0.4, 6.8)))
	door := func(r0, r1 float64) Solid {
		return lean(Intersect(m.arcade(c, r0, r1, 15, 0, 0.9, y0, 3.0), Block(V(c.X+2, y0, c.Y-1.5), V(c.X+7, y0+4, c.Y+1.5))))
	}
	g.Carve(door(4.6, 6.5)) // Door facing +X.
	g.With(Wood, func() { g.Fill(door(4.2, 4.7)) })
	// Galleries.
	const storey = 4.3
	for k := 0; k <= 6; k++ {
		y := 10.4 + float64(k)*storey
		g.FillVolume(lean(Cylinder{Center: c, Radius: 6.5, Inner: 2.7, Y0: y, Y1: y + 0.8}))
		if k == 6 {
			break
		}
		g.With(Stone, func() { g.FillVolume(lean(Cylinder{Center: c, Radius: 3.9, Inner: 2.7, Y0: y + 0.8, Y1: y + storey})) })
		g.Carve(lean(m.arcade(c, 2.5, 4.1, 4, 0.5*float64(k%2), 0.6, y+0.8, 1.4)))
		g.FillVolume(lean(Cylinder{Center: c, Radius: 6.0, Inner: 4.9, Y0: y + 0.8, Y1: y + storey}))
		g.Carve(lean(m.arcade(c, 4.8, 6.1, 12, 0.5, 0.95, y+0.8, 1.85)))
	}
	// Terrace parapet and the bell chamber with its bells.
	top := 10.4 + 6*storey + 0.8
	g.FillVolume(lean(Cylinder{Center: c, Radius: 6.5, Inner: 5.6, Y0: top, Y1: top + 0.9}))
	g.FillVolume(lean(Cylinder{Center: c, Radius: 4.1, Inner: 2.9, Y0: top, Y1: top + 4.6}))
	g.Carve(lean(m.arcade(c, 2.8, 4.2, 6, 0, 0.85, top+0.6, 1.8)))
	g.FillVolume(lean(Cylinder{Center: c, Radius: 4.6, Inner: 2.0, Y0: top + 4.6, Y1: top + 5.3}))
	g.FillVolume(lean(Cylinder{Center: c, Radius: 4.6, Inner: 3.6, Y0: top + 5.3, Y1: top + 6.0}))
	g.With(Stone, func() { g.FillVolume(lean(Cylinder{Center: c, Radius: 3.0, Y0: top + 4.2, Y1: top + 5.0})) })
	g.With(Gold, func() {
		for k := 0; k < 6; k++ {
			g.Fill(lean(Sphere(Polar(c, 1.6, float64(k)*math.Pi/3).At(top+2.6), 0.75)))
		}
	})
}

// baptistery builds the round baptistery: a striped blind arcade with four portals, the
// open loggia, a crown of gables and pinnacles, the drum and the dome — tiled towards
// the cathedral, leaded towards the sea — pierced by the leaded cone and lantern.
func (m PiazzaDeiMiracoli) baptistery(g *Grid) {
	c := miracoliBaptistery
	const base = miracoliBase
	g.Use(White)
	g.FillVolume(Cylinder{Center: c, Radius: 11, Inner: 9.8, Y0: base, Y1: 11})
	g.Paint(DarkStone, Func{Box{V(c.X-11.1, base+0.8, c.Y-11.1), V(c.X+11.1, 10, c.Y+11.1)}, func(p Vec3) bool {
		return p.Plan().Sub(c).Length() <= 11.05 && math.Mod(p.Y-base, 2.0) < 0.45
	}})
	g.FillVolume(Cylinder{Center: c, Radius: 11.7, Inner: 10.9, Y0: base, Y1: 11})
	g.Carve(m.arcade(c, 10.95, 11.8, 20, 0, 1.15, base+0.5, 6.5))
	g.Carve(m.arcade(c, 10.4, 12, 4, 0, 1.6, base+0.5, 4.1))
	g.With(Wood, func() { g.Fill(m.arcade(c, 9.9, 10.5, 4, 0, 1.6, base, 4.6)) })
	// Loggia between two cornices.
	g.FillVolume(Cylinder{Center: c, Radius: 12.0, Inner: 8.6, Y0: 11, Y1: 11.8})
	g.With(Stone, func() { g.FillVolume(Cylinder{Center: c, Radius: 9.8, Inner: 8.6, Y0: 11.8, Y1: 16.8}) })
	g.FillVolume(Cylinder{Center: c, Radius: 11.8, Inner: 10.8, Y0: 11.8, Y1: 16.8})
	g.Carve(m.arcade(c, 10.7, 11.9, 30, 0.5, 0.65, 11.8, 2.55))
	g.FillVolume(Cylinder{Center: c, Radius: 12.0, Inner: 8.6, Y0: 16.8, Y1: 17.5})
	// Gables and pinnacles in front of the drum.
	const n = 15
	bay := 2 * math.Pi / n
	g.FillVolume(Func{Box{V(c.X-12, 17.5, c.Y-12), V(c.X+12, 21.2, c.Y+12)}, func(p Vec3) bool {
		d := p.Plan().Sub(c)
		if r := d.Length(); r < 10.9 || r > 11.9 {
			return false
		}
		f := d.Angle() / bay
		u := math.Abs(f-math.Round(f)) * bay * 11.4
		return p.Y-17.5 <= (2.2-u)*1.6
	}})
	for k := 0; k < n; k++ {
		at := Polar(c, 11.4, (float64(k)+0.5)*bay)
		g.Fill(Cylinder{Center: at, Radius: 0.55, Y0: 17.5, Y1: 21.2}, Cone(at, 0.75, 21.2, 1.6))
	}
	g.FillVolume(Cylinder{Center: c, Radius: 10.0, Inner: 8.8, Y0: 17.5, Y1: 21.5})
	// Dome, half tiled and half leaded, with marble ribs; the cone and lantern.
	dome := Intersect(Subtract(Ellipsoid{c.At(21.5), V(10.3, 8.5, 10.3)}, Ellipsoid{c.At(21.5), V(9.0, 7.2, 9.0)}),
		Block(V(c.X-11, 21.5, c.Y-11), V(c.X+11, 31, c.Y+11)))
	g.With(Terracotta, func() { g.FillVolume(dome) })
	g.Paint(Slate, Intersect(dome, Block(V(c.X, 21.5, c.Y-11), V(c.X+11, 31, c.Y+11))))
	g.With(Slate, func() {
		g.FillVolume(Subtract(Frustum{c, 5.4, 2.0, 26, 34.5}, Frustum{c, 4.2, 0.8, 26, 34.5}))
	})
	g.FillVolume(Cylinder{Center: c, Radius: 2.3, Inner: 1.2, Y0: 34.5, Y1: 36.6})
	g.Carve(m.arcade(c, 1.0, 2.5, 4, 0.5, 0.6, 34.9, 0.8))
	g.With(Slate, func() {
		g.FillVolume(Intersect(Sphere(c.At(36.6), 2.4), Block(V(c.X-3, 36.6, c.Y-3), V(c.X+3, 39.2, c.Y+3))))
	})
	g.With(Gold, func() { g.Fill(Capsule{c.At(38.8), c.At(40.8), 0.6}) })
}

// camposanto builds the monumental cemetery: an outer marble wall with a blind arcade on
// the south front and a tabernacle over the door, the courtyard arcade, and roofs over
// the four galleries round the lawn.
func (m PiazzaDeiMiracoli) camposanto(g *Grid) {
	const base = miracoliBase
	const x0, x1, z0, z1, h = -16.0, 52.0, -52.0, -28.0, 7.4
	const gx0, gx1, gz0, gz1 = x0 + 6, x1 - 6, z0 + 6, z1 - 6
	g.Use(White)
	g.Fill(Subtract(Block(V(x0, base, z0), V(x1, h, z1)), Block(V(x0+1.1, base-1, z0+1.1), V(x1-1.1, h+1, z1-1.1))))
	g.Fill(Subtract(Block(V(gx0, base, gz0), V(gx1, h, gz1)), Block(V(gx0+1, base-1, gz0+1), V(gx1-1, h+1, gz1-1))))
	// No lawn under the galleries.
	g.Carve(Subtract(Block(V(x0+1.1, -0.2, z0+1.1), V(x1-1.1, 0.5, z1-1.1)), Block(V(gx0, -0.3, gz0), V(gx1, 0.6, gz1))))
	// Courtyard arcades.
	for x := gx0 + 2.7; x < gx1-1; x += 3.4 {
		for _, z := range []float64{gz0 + 0.5, gz1 - 0.5} {
			g.Carve(Arch{V(x, base+0.6, z), 1.2, 3.6, 2, AxisZ})
		}
	}
	for z := gz0 + 2.6; z < gz1-1; z += 3.4 {
		for _, x := range []float64{gx0 + 0.5, gx1 - 0.5} {
			g.Carve(Arch{V(x, base+0.6, z), 1.2, 3.6, 2, AxisX})
		}
	}
	// South front: pilasters and blind arches.
	for x := x0 + 1.5; x <= x1-1.4; x += 3.0 {
		g.Fill(Block(V(x-0.45, base, z1), V(x+0.45, h, z1+0.6)))
		if x+3.0 <= x1-1.4 {
			g.Fill(Block(V(x, 5.4, z1), V(x+3.0, h, z1+0.6)))
			g.Carve(Arch{V(x+1.5, base, z1+0.35), 1.0, 5.0, 0.6, AxisZ})
		}
	}
	// Doors and the tabernacle over the main one.
	for _, x := range []float64{-2.5, 40} {
		g.Carve(Arch{V(x, base, z1), 1.4, 2.8, 3, AxisZ})
		g.With(Wood, func() { g.Fill(Arch{V(x, base, z1-0.8), 1.4, 2.8, 0.4, AxisZ}) })
	}
	g.Fill(Block(V(38.2, h, z1-0.6), V(41.8, h+2.6, z1+0.8)))
	g.Carve(Arch{V(40, h+0.3, z1+0.6), 0.7, 0.9, 0.8, AxisZ})
	g.With(Slate, func() { g.Fill(HipRoof{P(37.8, z1-1), P(42.2, z1+1.2), h + 2.6}) })
	g.With(Gold, func() { g.Fill(Cylinder{Center: P(40, z1+0.1), Radius: 0.5, Y0: h + 4.2, Y1: h + 5.4}) })
	// Gallery roofs.
	g.Use(Slate)
	g.FillVolume(
		m.gableSlab(P(x0-0.4, z0-0.4), P(x1+0.4, gz0+0.4), h, 0.6, 1.1, AxisX),
		m.gableSlab(P(x0-0.4, gz1-0.4), P(x1+0.4, z1+0.4), h, 0.6, 1.1, AxisX),
		m.gableSlab(P(x0-0.4, z0-0.4), P(gx0+0.4, z1+0.4), h, 0.6, 1.1, AxisZ),
		m.gableSlab(P(gx1-0.4, z0-0.4), P(x1+0.4, z1+0.4), h, 0.6, 1.1, AxisZ),
	)
}

// walls builds the city wall along the north edge, with two towers and the gate.
func (m PiazzaDeiMiracoli) walls(g *Grid) {
	const base = miracoliBase
	const x0, x1, z0, z1, h = -57.0, 73.0, -57.0, -55.4, 6.6
	merlons := func(ax, az, bx, bz, y float64) Func {
		return Func{Box{V(ax, y, az), V(bx, y+1.2, bz)}, func(p Vec3) bool {
			onZ := p.Z < az+1.1 || p.Z > bz-1.1
			onX := p.X < ax+1.1 || p.X > bx-1.1
			return (onZ && math.Mod(p.X-ax, 2.4) < 1.3) || (onX && math.Mod(p.Z-az, 2.4) < 1.3)
		}}
	}
	g.Use(Sand)
	g.Fill(Block(V(x0, base, z0), V(x1, h, z1)))
	g.Fill(merlons(x0, z0, x1, z1, h))
	for _, x := range []float64{-36, 14} {
		g.Fill(Subtract(Block(V(x-2.4, base, z0-2.6), V(x+2.4, h+3, z1+0.6)), Block(V(x-1.3, base, z0-1.5), V(x+1.3, h+4, z1-0.5))))
		g.Fill(merlons(x-2.4, z0-2.6, x+2.4, z1+0.6, h+3))
		g.Carve(Arch{V(x, h-0.8, z0-2.6), 0.55, 1.2, 2, AxisZ})
	}
	// Porta Santa Maria: a tall gate tower over an arched passage, a path leading in.
	const gx = 60.0
	g.Fill(Subtract(Block(V(gx-3.6, base, z0-2.4), V(gx+3.6, h+4.4, z1+1.4)), Block(V(gx-2.4, base+1, z0-1.2), V(gx+2.4, h+5, z1+0.2))))
	g.Fill(merlons(gx-3.6, z0-2.4, gx+3.6, z1+1.4, h+4.4))
	g.Carve(Arch{V(gx, base, (z0+z1)/2), 1.9, 3.4, 8, AxisZ})
	g.Carve(Arch{V(gx, h+1.2, z1+1.4), 0.6, 1.4, 2, AxisZ}, Arch{V(gx, h+1.2, z0-2.4), 0.6, 1.4, 2, AxisZ})
}

func (m PiazzaDeiMiracoli) Build(g *Grid) Order {
	m.ground(g)
	m.camposanto(g)
	m.walls(g)
	m.cathedral(g)
	m.baptistery(g)
	m.tower(g)
	// Course by course: the cathedral from its façade, then baptistery and tower
	// sweeping round, the Camposanto from the east, then the walls.
	return func(p Vec3) float64 {
		q := p.Plan()
		switch {
		case p.X > -37 && p.X < 33 && math.Abs(p.Z) < 26:
			return (miracoliFacadeX + 3 - p.X) / 100
		case q.Sub(miracoliBaptistery).Length() < 14.5:
			return 1 + Sweep(miracoliBaptistery, 0)(p)/(2*math.Pi)
		case q.Sub(miracoliTower).Length() < 10:
			return 2 + Sweep(miracoliTower, 0)(p)/(2*math.Pi)
		case p.Z < -26 && p.Z > -54:
			return 3 + (60-p.X)/200
		}
		return 4 + (80-p.X)/200
	}
}
