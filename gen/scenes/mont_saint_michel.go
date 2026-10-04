package scenes

import (
	"math"

	"godocta/gen/kit"
	. "godocta/gen/octa"
)

// MontSaintMichel is the tidal island abbey: a granite rock rising from sand flats
// threaded by water channels, ringed on its south and east sides by ramparts with
// round towers and the King's Gate, a village of slate-roofed stone and half-timbered
// houses climbing the Grande Rue around the south-east slope, and the abbey on the
// summit: the Gothic church with its long slate roof, the buttressed Merveille on the
// north cliff, the abbot's lodgings and Châtelet to the south, and the slender spire
// crowned by the gilded Saint Michael. Woods cover the north-west slope; a causeway
// bridge runs south to the mainland. North is -Z, east +X.
type MontSaintMichel struct{}

func init() { Register(MontSaintMichel{}) }

func (MontSaintMichel) Name() string  { return "mont_saint_michel" }
func (MontSaintMichel) Title() string { return "Mont Saint Michel" }

const (
	montSaintMichelSummit  = 24.0 // Rock summit and church floor.
	montSaintMichelTerrace = 3.2  // Village ground behind the ramparts.
	montSaintMichelWall    = 5.0  // Rampart walkway.
	montSaintMichelDeck    = 2.6  // Causeway deck top.
	montSaintMichelStretch = 1.2  // The island is wider east-west.
)

var montSaintMichelCenter = P(0, -3)

// montSaintMichelMap is a sampled heightfield with bilinear lookup.
type montSaintMichelMap struct {
	x0, z0, step float64
	nx, nz       int
	h            []float64
}

func (m *montSaintMichelMap) at(x, z float64) float64 {
	fx, fz := (x-m.x0)/m.step, (z-m.z0)/m.step
	if fx < 0 || fz < 0 || fx >= float64(m.nx-1) || fz >= float64(m.nz-1) {
		return 0
	}
	i, j := int(fx), int(fz)
	tx, tz := fx-float64(i), fz-float64(j)
	k := j*m.nx + i
	a := m.h[k] + (m.h[k+1]-m.h[k])*tx
	b := m.h[k+m.nx] + (m.h[k+m.nx+1]-m.h[k+m.nx])*tx
	return a + (b-a)*tz
}

// low is the lowest height around a point, used to close steep faces into shells.
func (m *montSaintMichelMap) low(x, z float64) float64 {
	lo := m.at(x, z)
	for k := 0; k < 8; k++ {
		a := float64(k) * math.Pi / 4
		lo = math.Min(lo, m.at(x+1.2*math.Cos(a), z+1.2*math.Sin(a)))
	}
	return lo
}

// span returns the lowest and highest ground under a rectangle rotated by angle.
func (m *montSaintMichelMap) span(center Vec2, w, d, angle float64) (float64, float64) {
	lo, hi := math.Inf(1), math.Inf(-1)
	for i := 0; i <= 4; i++ {
		for j := 0; j <= 4; j++ {
			q := P((float64(i)/4-0.5)*w, (float64(j)/4-0.5)*d).Rotate(angle).Add(center)
			h := m.at(q.X, q.Y)
			lo, hi = math.Min(lo, h), math.Max(hi, h)
		}
	}
	return lo, hi
}

// noise is smooth value noise in [0, 1).
func (MontSaintMichel) noise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

func (s MontSaintMichel) fbm(x, z float64, seed int) float64 {
	return (s.noise(x, z, seed)*4 + s.noise(x*2.1, z*2.1, seed+1)*2 + s.noise(x*4.3, z*4.3, seed+2)) / 7
}

// polar returns the stretched radius and angle of a plan point around the island.
func (MontSaintMichel) polar(x, z float64) (float64, float64) {
	q := P((x-montSaintMichelCenter.X)/montSaintMichelStretch, z-montSaintMichelCenter.Y)
	return q.Length(), q.Angle()
}

// point is the inverse of polar.
func (MontSaintMichel) point(r, a float64) Vec2 {
	return montSaintMichelCenter.Add(P(montSaintMichelStretch*r*math.Cos(a), r*math.Sin(a)))
}

// radius is the rock's foot: wide on the south-east village slope, tight under the
// northern cliff.
func (MontSaintMichel) radius(a float64) float64 {
	return 27 + 6*math.Max(0, math.Cos(a-math.Pi/4)) - 5*math.Max(0, math.Cos(a+math.Pi/2))
}

// profile is the dome's height fraction at a relative radius.
func (MontSaintMichel) profile(u float64) float64 {
	t := math.Max(0, math.Min(1, (u-0.08)/0.92))
	return 0.5 * (1 + math.Cos(math.Pi*t))
}

// dome is the rock: a convex granite dome with steep flanks, easing into the gentler
// village slope on the south-east.
func (s MontSaintMichel) dome(r, a float64) float64 {
	u := r / s.radius(a)
	t := math.Max(0, math.Min(1, (u-0.08)/0.92))
	steep := math.Pow(math.Max(0, 1-math.Pow(t, 2.2)), 0.9)
	k := math.Pow(math.Max(0, math.Cos(a-math.Pi/4)), 0.7)
	return montSaintMichelSummit * (steep + (s.profile(u)-steep)*k)
}

// ramparts span from the west-south-west round the south and east to the north.
func (MontSaintMichel) inRamparts(a float64) bool {
	return a >= -55*math.Pi/180 && a <= 150*math.Pi/180
}

func (s MontSaintMichel) rampart(a float64) float64 { return s.radius(a) - 2 }

// Grande Rue: a spiral from the King's Gate at the south round the east slope.
const (
	montSaintMichelStreetA0 = 92 * math.Pi / 180
	montSaintMichelStreetA1 = 14 * math.Pi / 180
)

func (MontSaintMichel) streetR(a float64) float64 {
	t := (montSaintMichelStreetA0 - a) / (montSaintMichelStreetA0 - montSaintMichelStreetA1)
	return 24.5 + (13-24.5)*t
}

func (s MontSaintMichel) streetY(a float64) float64 {
	return math.Max(montSaintMichelTerrace, s.dome(s.streetR(a), a))
}

func (MontSaintMichel) onStreet(a float64) bool {
	return a <= montSaintMichelStreetA0 && a >= montSaintMichelStreetA1
}

// height is the island ground; zero means sand flats.
func (s MontSaintMichel) height(x, z float64) float64 {
	r, a := s.polar(x, z)
	h := s.dome(r, a)
	if h > 0 {
		t := r / s.radius(a)
		rough := (s.fbm(x/5, z/5, 3) - 0.5) * 4 * math.Sin(math.Pi*math.Min(1, t))
		h = math.Max(0, h+rough*math.Min(1, h/2))
	}
	if s.inRamparts(a) {
		rw := s.rampart(a)
		if r < rw-0.5 {
			h = math.Max(h, montSaintMichelTerrace+0.4*(s.fbm(x/4, z/4, 5)-0.5))
		}
		if r > rw+0.6 {
			h = 0
		}
	}
	if s.onStreet(a) && math.Abs(r-s.streetR(a)) < 1.4 {
		h = s.streetY(a)
	}
	return h
}

func (s MontSaintMichel) Build(g *Grid) Order {
	m := &montSaintMichelMap{x0: -44, z0: -36, step: 0.4}
	m.nx, m.nz = int(88/m.step)+2, int(70/m.step)+2
	m.h = make([]float64, m.nx*m.nz)
	for j := 0; j < m.nz; j++ {
		for i := 0; i < m.nx; i++ {
			m.h[j*m.nx+i] = s.height(m.x0+float64(i)*m.step, m.z0+float64(j)*m.step)
		}
	}
	s.flats(g, m)
	s.rock(g, m)
	s.ramparts(g)
	s.causeway(g)
	s.village(g, m)
	s.abbey(g, m)
	s.woods(g, m)
	s.boats(g)
	return Sweep(montSaintMichelCenter, math.Pi/2)
}

// flats lays the tidal sand with river channels, salt meadows and the mainland bank.
func (s MontSaintMichel) flats(g *Grid, m *montSaintMichelMap) {
	const cx, cz, ax, az = 0.0, 4.0, 44.0, 47.0
	inside := func(x, z float64) float64 {
		dx, dz := (x-cx)/ax, (z-cz)/az
		return math.Sqrt(dx*dx+dz*dz) + 0.03*(s.fbm(x/8, z/8, 21)-0.5)
	}
	channel := func(x, z float64) bool {
		river := []Vec2{P(-8, 52), P(-18, 42), P(-30, 28), P(-36, 10), P(-35, -8), P(-28, -26), P(-20, -46)}
		creek := []Vec2{P(36, 44), P(33, 28), P(37, 10), P(34, -10), P(38, -28), P(32, -46)}
		dist := func(path []Vec2) float64 {
			best := math.Inf(1)
			q := P(x, z)
			for i := 1; i < len(path); i++ {
				a, b := path[i-1], path[i]
				ab := b.Sub(a)
				t := math.Max(0, math.Min(1, q.Sub(a).Dot(ab)/ab.Dot(ab)))
				best = math.Min(best, q.Sub(a.Add(ab.Scale(t))).Length())
			}
			return best
		}
		return dist(river) < 2.4+1.2*s.fbm(x/6, z/6, 22) || dist(creek) < 1.1+0.8*s.fbm(x/5, z/5, 23)
	}
	bank := func(x, z float64) float64 {
		e := inside(x, z)
		b := math.Min(2.2, (z-44-2*math.Sin(x/9))*0.45)
		return math.Min(b, (1-e)*30)
	}
	box := Box{V(cx-ax-2, -1, cz-az-2), V(cx+ax+2, 3, cz+az+2)}
	g.Use(Sand).Fill(Func{box, func(p Vec3) bool {
		if inside(p.X, p.Z) > 1 || m.at(p.X, p.Z) > 0.6 {
			return false
		}
		top := math.Max(0.45, bank(p.X, p.Z))
		low := 0.0
		if channel(p.X, p.Z) && top <= 0.45 {
			low, top = -0.6, -0.2
		}
		return p.Y >= low && p.Y <= top && p.Y >= top-1.2
	}})
	g.Paint(Water, Func{box, func(p Vec3) bool { return channel(p.X, p.Z) && p.Y < 0.5 }})
	// Salt meadows towards the mainland, then the grassy bank.
	g.Paint(Grass, Func{box, func(p Vec3) bool {
		return !channel(p.X, p.Z) && (p.Z > 30 && s.fbm(p.X/7, p.Z/7, 24) > 0.56-0.004*(p.Z-30) || bank(p.X, p.Z) > 0.45)
	}})
	// Road from the causeway to the edge.
	g.Paint(Charcoal, Func{Box{V(-1.6, -1, 44), V(1.6, 4, 60)}, func(p Vec3) bool { return p.Y > 0.3 }})
	for i := 0; i < 9; i++ {
		x := -30 + float64(i)*7.5 + 3*Hash01(i, 31)
		z := 50 + 3*Hash01(i, 32) + math.Abs(x)*0.03
		if math.Abs(x) < 5 || inside(x, z) > 0.97 {
			continue
		}
		g.Build(kit.Tree{At: P(x, z), Base: bank(x, z) - 0.3, Height: 3.5 + 1.5*Hash01(i, 33)})
	}
}

// rock fills the island heightfield as a hollow granite shell.
func (s MontSaintMichel) rock(g *Grid, m *montSaintMichelMap) {
	box := Box{V(m.x0, 0, m.z0), V(m.x0+88, montSaintMichelSummit+2, m.z0+70)}
	g.Use(DarkStone).Fill(Func{box, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		if h <= 0.6 {
			return false
		}
		return p.Y <= h && p.Y >= math.Min(h, m.low(p.X, p.Z))-1.2
	}})
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		if m.at(p.X, p.Z) <= 0.6 {
			return false
		}
		band := int(math.Floor(p.Y/1.3 + 2*s.noise(p.X/7, p.Z/7, 11)))
		return Hash01(band, int(math.Floor(p.X/4)), int(math.Floor(p.Z/4)), 12) < 0.25
	}})
	// Village ground, paved lanes and garden plots, and the Grande Rue.
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		r, a := s.polar(p.X, p.Z)
		near := r > s.rampart(a)-5.5 || s.onStreet(a) && math.Abs(r-s.streetR(a)) < 6
		return near && s.inRamparts(a) && r < s.rampart(a)-0.5 && p.Y >= m.at(p.X, p.Z)-1 && m.at(p.X, p.Z)-m.low(p.X, p.Z) < 1.2
	}})
	g.Paint(Grass, Func{box, func(p Vec3) bool {
		r, a := s.polar(p.X, p.Z)
		return s.inRamparts(a) && r < s.rampart(a)-0.5 && p.Y >= m.at(p.X, p.Z)-1 && m.at(p.X, p.Z)-m.low(p.X, p.Z) < 1.6 &&
			s.fbm(p.X/4, p.Z/4, 14) > 0.5
	}})
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		r, a := s.polar(p.X, p.Z)
		return s.onStreet(a) && math.Abs(r-s.streetR(a)) < 1.2 && p.Y >= m.at(p.X, p.Z)-0.6
	}})
}

// ramparts raise the curtain wall, its towers and the King's Gate.
func (s MontSaintMichel) ramparts(g *Grid) {
	g.Use(Stone)
	box := Box{V(-44, 0, -36), V(44, montSaintMichelWall+2, 34)}
	g.Fill(Func{box, func(p Vec3) bool {
		r, a := s.polar(p.X, p.Z)
		if !s.inRamparts(a) {
			return false
		}
		rw := s.rampart(a)
		dr := r - rw
		if dr < -0.75 || dr > 0.75 {
			return false
		}
		switch {
		case p.Y <= montSaintMichelWall:
			return true
		case dr < 0:
			return false
		case p.Y <= montSaintMichelWall+1:
			return true
		}
		return p.Y <= montSaintMichelWall+1.8 && math.Mod(a*rw+400, 2.4) < 1.2
	}})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		r, a := s.polar(p.X, p.Z)
		return p.Y > 0.5 && p.Y < 1.4 && s.inRamparts(a) && r > s.rampart(a)-1.5
	}})
	// Towers, west to north: Gabriel, Arcade-side, King, Liberty, Basse, Boucle, Nord.
	type tower struct {
		deg, radius, top float64
		cone             bool
	}
	for _, t := range []tower{
		{150, 2.6, 8, true}, {120, 2.2, 7, false}, {75, 2.5, 8, false}, {50, 2.3, 7.5, true},
		{24, 2.2, 7, false}, {-8, 3.2, 7.5, false}, {-32, 2.3, 8, true}, {-55, 2.5, 8.5, true},
	} {
		a := t.deg * math.Pi / 180
		at := s.point(s.rampart(a)+0.6, a)
		roof := kit.RoofBattlements
		if t.cone {
			roof = kit.RoofCone
		}
		g.Build(kit.Tower{Center: at, Radius: t.radius, Base: 0, Top: t.top, Storey: 4, Roof: roof, RoofHeight: 2.2 * t.radius, Merlons: 10, Windows: true})
	}
	// King's Gate: a tall gatehouse astride the wall with a forecourt before it.
	gz := s.point(s.rampart(math.Pi/2), math.Pi/2).Y
	g.Fill(Block(V(-3.5, 0, gz-2.5), V(3.5, 8, gz+2.5)))
	g.Carve(Block(V(-2.3, 2.6, gz-1.3), V(2.3, 7, gz+1.3)))
	g.With(Slate, func() { g.Fill(HipRoof{P(-4, gz-3), P(4, gz+3), 8}) })
	g.Carve(Arch{V(0, montSaintMichelDeck-0.4, gz), 1.6, 2.2, 8, AxisZ})
	g.Paint(Charcoal, Arch{V(0, montSaintMichelDeck-0.4, gz-2), 1.6, 2.2, 0.8, AxisZ})
	for _, x := range []float64{-2, 2} {
		g.Carve(Arch{V(x, 5, gz+2.5), 0.5, 1.2, 1.5, AxisZ})
	}
	// Forecourt (the Avancée), floored at deck level.
	g.Fill(Subtract(Block(V(-5, 0, gz+2.5), V(5, montSaintMichelDeck+2.5, gz+7.5)), Block(V(-4, montSaintMichelDeck, gz+2.5), V(4, 10, gz+6.5))))
	g.Build(kit.Battlements{Min: P(-5, gz+2.5), Max: P(5, gz+7.5), Top: montSaintMichelDeck + 2.5, Rim: 1, Period: 2})
	g.Carve(Block(V(-4, montSaintMichelDeck+0.1, gz+2.5), V(4, 10, gz+6.5)))
	g.Carve(Arch{V(0, montSaintMichelDeck, gz+7), 1.6, 1.6, 2, AxisZ})
	g.With(Sand, func() { g.Fill(Block(V(-4, montSaintMichelDeck-0.5, gz+2.5), V(4, montSaintMichelDeck, gz+6.5))) })
}

// causeway is the pile bridge from the forecourt to the mainland bank.
func (s MontSaintMichel) causeway(g *Grid) {
	gz := s.point(s.rampart(math.Pi/2), math.Pi/2).Y + 7.5
	const end = 47.0
	g.With(Wood, func() { g.Fill(Block(V(-1.8, montSaintMichelDeck-0.8, gz), V(1.8, montSaintMichelDeck, end))) })
	g.With(Charcoal, func() {
		for _, x := range []float64{-1.8, 1.8} {
			g.Fill(Block(V(x-0.25, montSaintMichelDeck, gz), V(x+0.25, montSaintMichelDeck+0.9, end)))
		}
		for z := gz + 2; z < end-1; z += 3 {
			for _, x := range []float64{-1.1, 1.1} {
				g.Fill(Cylinder{Center: P(x, z), Radius: 0.5, Y0: 0, Y1: montSaintMichelDeck - 0.8})
			}
		}
	})
}

// house is one village house, rotated to face the street.
func (s MontSaintMichel) house(g *Grid, m *montSaintMichelMap, at Vec2, angle, w, d, street float64, outer bool, seed int) {
	place := func(sol Solid) Solid { return Translate(RotateY(sol, P(0, 0), angle), at.At(0)) }
	lo, hi := m.span(at, w, d, angle)
	base := lo - 1
	var top float64
	if outer {
		top = street + 4.2 + 2.2*Hash01(seed, 1)
	} else {
		top = math.Max(hi, street) + 3.4 + 2*Hash01(seed, 1)
	}
	kind := Hash01(seed, 2)
	wall := Stone
	switch {
	case kind < 0.35:
		wall = White
	case kind > 0.78:
		wall = DarkStone
	}
	g.Use(wall).Fill(place(Block(V(-w/2, base, -d/2), V(w/2, top, d/2))))
	g.Carve(place(Block(V(-w/2+1, base+0.5, -d/2+1), V(w/2-1, top-1, d/2-1))))
	if wall == White {
		// Half-timbering over a stone ground floor.
		ground := math.Max(hi, street) + 1.6
		g.Paint(Stone, place(Block(V(-w/2, base, -d/2), V(w/2, ground, d/2))))
		g.Paint(Wood, place(Func{Box{V(-w/2, ground, -d/2), V(w/2, top, d/2)}, func(p Vec3) bool {
			band := math.Mod(p.Y-ground+20, 2.4) < 0.45 || p.Y > top-0.45
			post := math.Mod(p.X+w/2+20, 1.7) < 0.45 || math.Mod(p.Z+d/2+20, 1.7) < 0.45
			return band || post && p.Y < top-0.4
		}}))
	}
	// Windows on the street and back fronts, a door on the street.
	floor := math.Max(hi, street) + 1.4
	for y := floor; y < top-1.4; y += 2.4 {
		for _, x := range []float64{-w / 4, w / 4} {
			g.Paint(Navy, place(Block(V(x-0.4, y, -d/2-0.5), V(x+0.4, y+1, d/2+0.5))))
		}
	}
	// Local +Z points away from the island center: outer houses face inward.
	streetFace := d/2 + 0.5
	if outer {
		streetFace = -streetFace
	}
	g.Paint(Wood, place(Block(V(-0.5, street, streetFace-1.0), V(0.5, street+1.9, streetFace+1.0))))
	// Shop awnings over the Grande Rue.
	if h := Hash01(seed, 6); h < 0.6 && math.Abs(hi-street) < 1.5 {
		awning := []Color{Red, Navy, Terracotta, Red, White}[int(h*8)%5]
		out := math.Copysign(1, streetFace)
		g.With(awning, func() {
			g.Fill(place(Block(V(-w/2+0.6, street+2.2, math.Min(streetFace, streetFace+out*0.9)), V(w/2-0.6, street+2.7, math.Max(streetFace, streetFace+out*0.9)))))
		})
	}
	// Steep slate roof; gable ends to the street on some houses.
	ridgeX := Hash01(seed, 3) < 0.6
	half, run := d/2+0.4, w/2+0.4
	if !ridgeX {
		half, run = run, half
	}
	const pitch = 1.15
	roof := Func{Box{V(-run, top, -half), V(run, top+half*pitch, half)}, func(p Vec3) bool {
		across := math.Abs(p.Z)
		if !ridgeX {
			across = math.Abs(p.X)
		}
		return across <= half-(p.Y-top)/pitch
	}}
	if !ridgeX {
		roof.Box = Box{V(-half, top, -run), V(half, top+half*pitch, run)}
	}
	g.With(Slate, func() { g.Fill(place(roof)) })
	if Hash01(seed, 4) < 0.5 {
		cx := (Hash01(seed, 5) - 0.5) * (w - 2)
		g.With(Stone, func() { g.Fill(place(Block(V(cx-0.45, top, -0.45), V(cx+0.45, top+half*pitch+0.6, 0.45)))) })
	}
}

// village packs houses along both sides of the Grande Rue and in rows above and below.
func (s MontSaintMichel) village(g *Grid, m *montSaintMichelMap) {
	type plot struct {
		at Vec2
		r  float64
	}
	var placed []plot
	free := func(at Vec2, r float64) bool {
		for _, p := range placed {
			if p.at.Sub(at).Length() < p.r+r-0.5 {
				return false
			}
		}
		return true
	}
	seed := 0
	// Parish church of Saint-Pierre on the inner side, halfway up.
	{
		a := 58 * math.Pi / 180
		r := s.streetR(a)
		c := s.point(r, a)
		tangent := s.point(r, a-0.01).Sub(c)
		tangent = tangent.Scale(1 / tangent.Length())
		normal := c.Sub(montSaintMichelCenter)
		normal = normal.Scale(1 / normal.Length())
		at := c.Sub(normal.Scale(4))
		angle := tangent.Angle()
		place := func(sol Solid) Solid { return Translate(RotateY(sol, P(0, 0), angle), at.At(0)) }
		street := s.streetY(a)
		lo, _ := m.span(at, 8, 4.5, angle)
		g.Use(Stone).Fill(place(Block(V(-4, lo-1, -2.2), V(4, street+5, 2.2))))
		g.Carve(place(Block(V(-3, lo, -1.2), V(3, street+4, 1.2))))
		g.With(Slate, func() {
			g.Fill(place(Func{Box{V(-4.4, street+5, -2.6), V(4.4, street+8, 2.6)}, func(p Vec3) bool {
				return math.Abs(p.Z) <= 2.6-(p.Y-street-5)/1.15
			}}))
		})
		for _, x := range []float64{-2.2, 0, 2.2} {
			g.Carve(place(Arch{V(x, street+1.5, 2.2), 0.5, 1.8, 1.5, AxisZ}))
		}
		// Bell tower at the uphill end with a slate spire.
		g.Fill(place(Block(V(-6.5, lo-1, -1.5), V(-3.5, street+9, 1.5))))
		g.Carve(place(Arch{V(-5, street+6.5, 1.5), 0.6, 1.2, 4, AxisZ}), place(Arch{V(-6.5, street+6.5, 0), 0.6, 1.2, 4, AxisX}))
		g.With(Slate, func() {
			g.Fill(place(Func{Box{V(-6.8, street+9, -1.8), V(-3.2, street+13, 1.8)}, func(p Vec3) bool {
				k := 1.8 * (1 - (p.Y-street-9)/4)
				return math.Abs(p.X+5) <= k && math.Abs(p.Z) <= k
			}}))
		})
		g.With(Gold, func() { g.Fill(place(Block(V(-5.3, street+13, -0.3), V(-4.7, street+14, 0.3)))) })
		placed = append(placed, plot{at, 4}, plot{at.Add(tangent.Scale(-5)), 2})
	}
	for a := montSaintMichelStreetA0 - 0.02; a > montSaintMichelStreetA1+0.04; {
		r := s.streetR(a)
		c := s.point(r, a)
		ahead := s.point(s.streetR(a-0.01), a-0.01)
		tangent := ahead.Sub(c)
		tangent = tangent.Scale(1 / tangent.Length())
		normal := P(tangent.Y, -tangent.X) // Points outward for a clockwise walk.
		if normal.Dot(c.Sub(montSaintMichelCenter)) < 0 {
			normal = normal.Scale(-1)
		}
		street := s.streetY(a)
		angle := tangent.Angle()
		w := 3.6 + 1.4*Hash01(seed, 7)
		for row, off := range []float64{1.4, -1.4, 6.4, -6.6, -11.2} {
			seed++
			d := 3.6 + 0.8*Hash01(seed, 8)
			shift := off + math.Copysign(d/2, off)
			at := c.Add(normal.Scale(shift)).Add(tangent.Scale((Hash01(seed, 9) - 0.5) * 1.2))
			if row >= 2 && Hash01(seed, 10) < 0.25 {
				continue
			}
			rr, ra := s.polar(at.X, at.Y)
			if rr+d*0.6 > s.rampart(ra)-0.9 || s.inAbbey(at, 2.5) {
				continue
			}
			radius := math.Hypot(w, d) / 2 * 0.8
			if !free(at, radius) {
				continue
			}
			placed = append(placed, plot{at, radius})
			ground := street
			if row >= 2 {
				ground = m.at(at.X, at.Y)
			}
			s.house(g, m, at, angle, w, d, ground, off > 0, seed)
		}
		a -= (w + 0.5) / (r * 1.1)
	}
	// Houses behind the ramparts: west of the gate and along the north-east walls.
	for k := 0; k < 14; k++ {
		a := (100 + float64(k)*9) * math.Pi / 180
		if k >= 6 {
			a = (12 - float64(k-6)*8.5) * math.Pi / 180
		}
		r := s.rampart(a) - 3.4
		at := s.point(r, a)
		tangent := s.point(r, a-0.01).Sub(at)
		w, d := 3.8+Hash01(k, 41), 3.6
		radius := math.Hypot(w, d) / 2 * 0.8
		if !free(at, radius) || s.inAbbey(at, 2) {
			continue
		}
		placed = append(placed, plot{at, radius})
		s.house(g, m, at, tangent.Angle(), w, d, m.at(at.X, at.Y), true, 500+k)
	}
	// Garden trees wherever the terraces leave room.
	for i := -14; i <= 14; i++ {
		for j := -14; j <= 14; j++ {
			at := P(float64(i)*2.5+Hash01(i, j, 60), float64(j)*2.5+Hash01(i, j, 61))
			r, a := s.polar(at.X, at.Y)
			if !s.inRamparts(a) || r > s.rampart(a)-1.8 || s.inAbbey(at, 1.5) || Hash01(i, j, 62) > 0.45 {
				continue
			}
			if s.onStreet(a) && math.Abs(r-s.streetR(a)) < 2.2 || !free(at, 1.4) {
				continue
			}
			h := m.at(at.X, at.Y)
			if h-m.low(at.X, at.Y) > 1.5 {
				continue
			}
			placed = append(placed, plot{at, 1.4})
			g.Build(kit.Tree{At: at, Base: h - 0.4, Height: 3 + 1.2*Hash01(i, j, 63)})
		}
	}
}

// abbey footprints in plan, used to keep houses clear.
func (MontSaintMichel) inAbbey(p Vec2, margin float64) bool {
	boxes := [][4]float64{
		{-20, 15, -9, 2.5}, // Terrace, church and choir.
		{-3, 4, -12.5, 5},  // Transept.
		{-2, 15, -18, -9},  // Merveille.
		{-1, 16.5, 2.5, 8}, // Abbot's lodgings and Châtelet.
	}
	for _, b := range boxes {
		if p.X > b[0]-margin && p.X < b[1]+margin && p.Y > b[2]-margin && p.Y < b[3]+margin {
			return true
		}
	}
	return false
}

// block is a hollow masonry mass from the ground under its footprint up to top.
func (s MontSaintMichel) block(g *Grid, m *montSaintMichelMap, min, max Vec2, top float64) float64 {
	c := min.Add(max).Scale(0.5)
	lo, _ := m.span(c, max.X-min.X, max.Y-min.Y, 0)
	g.Fill(Block(min.At(lo-1), max.At(top)))
	g.Carve(Block(min.Add(P(1.2, 1.2)).At(lo-0.5), max.Sub(P(1.2, 1.2)).At(top-1)))
	// Buttresses down to the local ground, tall windows between them.
	rib := func(at Vec2, dx, dz float64) {
		base := m.at(at.X, at.Y) - 1
		if top-base < 3 {
			return
		}
		g.Fill(Block(V(at.X-dx, base, at.Y-dz), V(at.X+dx, top-1.2, at.Y+dz)))
	}
	for x := min.X + 1.2; x < max.X-0.6; x += 2.8 {
		rib(P(x, min.Y-0.45), 0.5, 0.45)
		rib(P(x, max.Y+0.45), 0.5, 0.45)
		for y := m.at(x+1.4, min.Y) + 1.5; y < top-3; y += 3.4 {
			g.Paint(Navy, Block(V(x+0.9, y, min.Y-0.3), V(x+1.9, y+1.8, min.Y+0.5)))
		}
		for y := m.at(x+1.4, max.Y) + 1.5; y < top-3; y += 3.4 {
			g.Paint(Navy, Block(V(x+0.9, y, max.Y-0.5), V(x+1.9, y+1.8, max.Y+0.3)))
		}
	}
	for z := min.Y + 1.2; z < max.Y-0.6; z += 2.8 {
		rib(P(min.X-0.45, z), 0.45, 0.5)
		rib(P(max.X+0.45, z), 0.45, 0.5)
		for y := m.at(min.X, z+1.4) + 1.5; y < top-3; y += 3.4 {
			g.Paint(Navy, Block(V(min.X-0.3, y, z+0.9), V(min.X+0.5, y+1.8, z+1.9)))
		}
		for y := m.at(max.X, z+1.4) + 1.5; y < top-3; y += 3.4 {
			g.Paint(Navy, Block(V(max.X-0.5, y, z+0.9), V(max.X+0.3, y+1.8, z+1.9)))
		}
	}
	g.Paint(DarkStone, Block(min.Sub(P(1, 1)).At(top-1), max.Add(P(1, 1)).At(top-0.6)))
	return lo
}

// steep is a pitched roof over a rectangle with the given rise per unit across.
func (MontSaintMichel) steep(min, max Vec2, base, pitch float64, ridge Axis) Func {
	half := (max.Y - min.Y) / 2
	if ridge == AxisZ {
		half = (max.X - min.X) / 2
	}
	return Func{Box{min.At(base), max.At(base + half*pitch)}, func(p Vec3) bool {
		across := math.Abs(p.Z - (min.Y+max.Y)/2)
		if ridge == AxisZ {
			across = math.Abs(p.X - (min.X+max.X)/2)
		}
		return across <= half-(p.Y-base)/pitch
	}}
}

// abbey crowns the summit.
func (s MontSaintMichel) abbey(g *Grid, m *montSaintMichelMap) {
	const floor = montSaintMichelSummit
	g.Use(Stone)
	// Substructures: crypts and the west terrace carry the church floor.
	s.block(g, m, P(-13, -8.5), P(12, 2.5), floor)
	s.block(g, m, P(-2.5, -11), P(4.5, 5), floor)
	s.block(g, m, P(-18.5, -7), P(-13, 1), floor)
	g.With(Sand, func() { g.Fill(Block(V(-18.5, floor-0.5, -7), V(-13, floor, 1))) })
	g.Fill(Subtract(Block(V(-18.5, floor, -7), V(-13, floor+1, 1)), Block(V(-17.7, floor, -6.2), V(-12, floor+2, 0.2))))

	// Nave.
	const nz0, nz1, nave = -6.5, 0.5, floor + 3
	g.Fill(Block(V(-13, floor, nz0), V(-2.5, nave, nz1)))
	g.Carve(Block(V(-12, floor, nz0+1), V(-2.5, nave-1, nz1-1)))
	for x := -11.0; x < -3; x += 2.8 {
		g.Fill(Block(V(x-0.5, floor, nz0-1), V(x+0.5, nave-0.5, nz0)), Block(V(x-0.5, floor, nz1), V(x+0.5, nave-0.5, nz1+1)))
		g.Carve(Arch{V(x+1.4, floor+1.8, nz0), 0.6, 1.4, 2.5, AxisZ}, Arch{V(x+1.4, floor+1.8, nz1), 0.6, 1.4, 2.5, AxisZ})
	}
	g.Carve(Arch{V(-13, floor, -3), 1.2, 1.8, 2.5, AxisX})
	g.With(Slate, func() { g.Fill(s.steep(P(-13.4, nz0-0.5), P(-2.5, nz1+0.5), nave, 0.95, AxisX)) })
	g.Fill(Intersect(Block(V(-13.5, nave, nz0), V(-12.7, nave+5, nz1)), s.steep(P(-13.5, nz0-0.5), P(-12.7, nz1+0.5), nave, 0.95, AxisX)))

	// Transept arms with gabled roofs.
	g.Fill(Block(V(-2.5, floor, -10.5), V(4.5, nave, 4.5)))
	g.Carve(Block(V(-1.5, floor, -9.5), V(3.5, nave-1, 3.5)))
	for _, z := range []float64{-10.5, 4.5} {
		g.Carve(Arch{V(1, floor+1.3, z), 0.8, 1.6, 2.5, AxisZ})
	}
	g.With(Slate, func() { g.Fill(s.steep(P(-3, -11), P(5, 5), nave, 0.95, AxisZ)) })

	// Gothic choir, taller, ending in an apse ringed by flying buttresses.
	const choir = floor + 3.5
	apse := P(9, -3)
	g.Fill(Block(V(4.5, floor, -7), V(9, choir, 1)), Cylinder{Center: apse, Radius: 4, Y0: floor, Y1: choir})
	g.Carve(Block(V(4.5, floor, -6), V(9, choir-1, 0)), Cylinder{Center: apse, Radius: 3, Y0: floor, Y1: choir - 1})
	for _, x := range []float64{5.8, 7.6} {
		g.Carve(Arch{V(x, floor+2, -7), 0.5, 2.6, 2.5, AxisZ}, Arch{V(x, floor+2, 1), 0.5, 2.6, 2.5, AxisZ})
	}
	g.With(Slate, func() {
		g.Fill(s.steep(P(4.5, -7.5), P(9, 1.5), choir, 0.9, AxisX))
		g.Fill(Intersect(Cone(apse, 4.5, choir, 4.5*0.9), Block(V(9, choir, -8), V(14, choir+6, 2))))
	})
	for k := 0; k < 7; k++ {
		a := -math.Pi/2 + float64(k)*math.Pi/6
		if k == 0 || k == 6 {
			a = -math.Pi/2 - 0.0001
		}
		var foot, wall Vec2
		switch k {
		case 0:
			foot, wall = P(6.7, -10), P(6.7, -7)
		case 6:
			foot, wall = P(6.7, 4), P(6.7, 1)
		default:
			foot, wall = Polar(apse, 7, a), Polar(apse, 4, a)
		}
		g.Fill(Block(V(foot.X-0.55, floor-0.5, foot.Y-0.55), V(foot.X+0.55, choir+0.5, foot.Y+0.55)))
		g.With(DarkStone, func() { g.Fill(Cone(foot, 0.8, choir+0.5, 1.8)) })
		g.Fill(Capsule{foot.At(choir - 0.8), wall.At(choir - 0.3), 0.5})
		g.Carve(Arch{V(foot.X, floor, foot.Y), 0.45, 1.5, 1.5, AxisX})
	}

	// Crossing tower, openwork lantern, spire and Saint Michael.
	tc := P(1, -3)
	const towerTop = floor + 7.8
	g.Fill(BlockAt(tc, 6, 6, floor, towerTop))
	g.Carve(BlockAt(tc, 4, 4, floor, towerTop-1))
	for _, off := range []float64{-1.3, 1.3} {
		g.Carve(Arch{V(tc.X+off, towerTop-3.6, tc.Y-3), 0.5, 2, 2, AxisZ}, Arch{V(tc.X+off, towerTop-3.6, tc.Y+3), 0.5, 2, 2, AxisZ})
		g.Carve(Arch{V(tc.X-3, towerTop-3.6, tc.Y+off), 0.5, 2, 2, AxisX}, Arch{V(tc.X+3, towerTop-3.6, tc.Y+off), 0.5, 2, 2, AxisX})
	}
	g.Fill(Block(V(tc.X-3.5, towerTop-0.5, tc.Y-3.5), V(tc.X+3.5, towerTop, tc.Y+3.5)))
	g.Build(kit.Battlements{Min: P(tc.X-3.5, tc.Y-3.5), Max: P(tc.X+3.5, tc.Y+3.5), Top: towerTop - 0.5, Rim: 0.6, Period: 1.6})
	for _, dx := range []float64{-3, 3} {
		for _, dz := range []float64{-3, 3} {
			g.Fill(Cylinder{Center: P(tc.X+dx, tc.Y+dz), Radius: 0.6, Y0: towerTop - 2, Y1: towerTop + 1.5})
		}
	}
	const lantern = towerTop + 1.2
	g.Fill(Cylinder{Center: tc, Radius: 2, Y0: towerTop, Y1: lantern})
	for k := 0; k < 4; k++ {
		a := float64(k)*math.Pi/2 + math.Pi/4
		at := Polar(tc, 2, a)
		g.Carve(Func{Box{V(at.X-1, towerTop+0.4, at.Y-1), V(at.X+1, lantern-0.4, at.Y+1)}, func(p Vec3) bool {
			return p.Plan().Sub(at).Length() < 0.7
		}})
	}
	g.With(Slate, func() {
		g.FillVolume(Frustum{tc, 1.35, 0.4, lantern, lantern + 5.1})
		g.Fill(Cylinder{Center: tc, Radius: 0.45, Y0: lantern, Y1: lantern + 5.5})
	})
	statue := lantern + 5.5
	g.With(Gold, func() {
		g.Fill(Capsule{tc.At(statue), tc.At(statue + 1.4), 0.55})
		g.Fill(Block(V(tc.X-0.3, statue+0.7, tc.Y-1.1), V(tc.X+0.3, statue+1.3, tc.Y+1.1)))
		g.Fill(Capsule{V(tc.X, statue+1.2, tc.Y+0.6), V(tc.X, statue+2.1, tc.Y+0.6), 0.4})
	})

	// The Merveille: a sheer, buttressed wall block on the north cliff.
	const mx0, mx1, mz0, mz1, mtop = -2.0, 15.0, -18.0, -10.5, floor + 3
	lo := s.block(g, m, P(mx0, mz0), P(mx1, mz1), mtop)
	// Refectory roof on the east half, the cloister garden on the west.
	g.With(Slate, func() { g.Fill(s.steep(P(5.5, mz0-0.4), P(mx1+0.4, mz1+0.4), mtop, 1.05, AxisX)) })
	court := Block(V(mx0+2.6, mtop-1.5, mz0+2.6), V(3, mtop+6, mz1-2.6))
	g.With(Slate, func() {
		g.Fill(Subtract(Intersect(HipRoof{P(mx0-0.4, mz0-0.4), P(9, mz1+0.4), mtop}, Block(V(mx0-0.4, mtop, mz0-0.4), V(5.5, mtop+5, mz1+0.4))), court))
	})
	g.Carve(court)
	g.With(Grass, func() { g.Fill(Block(V(mx0+2.6, mtop-1.5, mz0+2.6), V(3, mtop-1, mz1-2.6))) })

	// Abbot's lodgings to the south, and the Châtelet gate at their east end.
	const lz0, lz1, ltop = 2.5, 8.0, floor + 4
	lo = s.block(g, m, P(-1, lz0), P(12, lz1), ltop)
	g.With(Slate, func() { g.Fill(s.steep(P(-1.4, lz0-0.4), P(12.4, lz1+0.4), ltop, 1.05, AxisX)) })
	lo = s.block(g, m, P(12, 2), P(15.5, 8.5), floor+1)
	for _, z := range []float64{2.5, 8} {
		g.Fill(Cylinder{Center: P(15.5, z), Radius: 1.4, Y0: lo - 1, Y1: floor + 2.5})
		g.With(Slate, func() { g.FillVolume(Cone(P(15.5, z), 1.7, floor+2.5, 3)) })
	}
	g.With(Slate, func() { g.Fill(s.steep(P(11.8, 1.8), P(15.8, 8.7), floor+1, 1.05, AxisZ)) })
	streetEnd := s.streetY(montSaintMichelStreetA1)
	g.Carve(Arch{V(15.5, streetEnd, 5.25), 1.2, 2, 3, AxisX})
	// Grand Degré: steps from the top of the Grande Rue into the Châtelet.
	end := s.point(s.streetR(montSaintMichelStreetA1), montSaintMichelStreetA1)
	steps := int(math.Ceil(end.X - 16))
	for k := 0; k <= steps; k++ {
		x := 16 + float64(k)
		g.With(Sand, func() { g.Fill(Block(V(x, streetEnd-1.5, 3.8), V(x+1, streetEnd, 6.7))) })
	}
	g.Fill(Block(V(16, streetEnd-1.5, 3.1), V(end.X, streetEnd+1, 3.8)), Block(V(16, streetEnd-1.5, 6.7), V(end.X, streetEnd+1, 7.4)))
}

// woods plant trees and lawns on the north-west slope and the chapel of Saint-Aubert.
func (s MontSaintMichel) woods(g *Grid, m *montSaintMichelMap) {
	box := Box{V(m.x0, 0, m.z0), V(m.x0+88, montSaintMichelSummit, m.z0+70)}
	g.Paint(Grass, Func{box, func(p Vec3) bool {
		_, a := s.polar(p.X, p.Z)
		h := m.at(p.X, p.Z)
		if s.inRamparts(a) || h < 1 || p.Y < h-1 || s.inAbbey(p.Plan(), 0.5) {
			return false
		}
		return p.Y <= h+0.3 && h-m.low(p.X, p.Z) < 1.6 && s.fbm(p.X/6, p.Z/6, 40) > 0.38
	}})
	for i := -15; i <= 15; i++ {
		for j := -15; j <= 15; j++ {
			at := P(float64(i)*2.6+1.2*Hash01(i, j, 50), float64(j)*2.6+1.2*Hash01(i, j, 51))
			r, a := s.polar(at.X, at.Y)
			h := m.at(at.X, at.Y)
			if s.inRamparts(a) || h < 1.2 || h > 18 || s.inAbbey(at, 1.2) || r > s.radius(a)-1 || Hash01(i, j, 52) < 0.15 {
				continue
			}
			g.Build(kit.Tree{At: at, Base: h - 0.5, Height: 4.4 + 2*Hash01(i, j, 53)})
		}
	}
	// Chapel of Saint-Aubert on a boulder at the north-west foot.
	a := -140 * math.Pi / 180
	at := s.point(s.radius(a)-1.5, a)
	h := m.at(at.X, at.Y)
	g.Use(DarkStone).FillVolume(Ellipsoid{at.At(h - 0.5), V(3.2, 2.2, 3.2)})
	g.Use(Stone).Fill(BlockAt(at, 3.2, 2.4, h, h+2.8))
	g.With(Slate, func() { g.Fill(s.steep(at.Sub(P(1.9, 1.5)), at.Add(P(1.9, 1.5)), h+2.8, 1.1, AxisX)) })
	g.With(White, func() { g.Fill(Cylinder{Center: at.Add(P(1.9, 0)), Radius: 0.5, Y0: h + 2.8, Y1: h + 5.5}) })
}

// boats lie stranded on the sand by the channels at low tide.
func (MontSaintMichel) boats(g *Grid) {
	for k, b := range []struct {
		at    Vec2
		angle float64
	}{{P(-27, 25), 0.9}, {P(31, 20), -1.3}, {P(-31, -20), 2.0}, {P(29, -16), 1.7}} {
		place := func(sol Solid) Solid { return Translate(RotateY(sol, P(0, 0), b.angle), b.at.At(0)) }
		hull := Intersect(Ellipsoid{V(0, 2, 0), V(3.2, 2, 1.3)}, Block(V(-4, 0.4, -2), V(4, 1.9, 2)))
		g.With(Wood, func() { g.Fill(place(hull)) })
		g.Carve(place(Intersect(Ellipsoid{V(0, 2, 0), V(2.4, 1.4, 0.6)}, Block(V(-4, 1.2, -2), V(4, 2, 2)))))
		if k%2 == 1 {
			g.Paint(Navy, place(Block(V(-4, 0, -2), V(4, 1.5, 2))))
		}
		g.Paint(White, place(Block(V(-4, 1.5, -2), V(4, 1.9, 2))))
		g.With(Wood, func() { g.Fill(place(Capsule{V(0.5, 1, 0), V(0.5, 5, 0), 0.4})) })
	}
}
