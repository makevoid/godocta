package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// BalloonFestival is a dawn mass ascent from a riverside meadow. Nine hot-air balloons
// are caught at different stages: one half-inflated on its side with the burner roaring
// into its mouth, one upright on three ground tethers, one just lifting on handling
// lines, and six climbing to different heights. Every envelope is a hollow inverted
// teardrop with its own pattern (rainbow gores, chequers, chevrons, spirals, bands,
// harlequin diamonds) over a dark skirt, a burner flame, load lines and a wicker basket
// with passengers. The rounded meadow plate has a mown, striped launch field ringed by
// rough grass, hedges and trees; a river winds across the front with a footbridge to the
// chase-van car park on the near bank, a promenade of spectators and flags follows the
// far bank, and striped marquees stand on the western side. The camera side is +X/+Z:
// the river and car park are along +Z, the marquee village is toward -X.
type BalloonFestival struct{}

func init() { Register(BalloonFestival{}) }

func (BalloonFestival) Name() string  { return "balloon_festival" }
func (BalloonFestival) Title() string { return "Balloon Festival" }

const (
	balloonFestivalA    = 58.0 // Meadow plate semi-axes along X and Z.
	balloonFestivalB    = 48.0
	balloonFestivalSkin = 1.1  // Envelope fabric thickness.
	balloonFestivalStep = 0.05 // Height step of the envelope profile tables.
	balloonFestivalTall = 2.35 // Envelope height in equator radii.
	balloonFestivalDeep = 1.4  // Mouth-to-equator height in equator radii.
)

// Envelope patterns.
const (
	balloonFestivalGores = iota
	balloonFestivalBands
	balloonFestivalSpiral
	balloonFestivalChequer
	balloonFestivalZigzag
	balloonFestivalDiamonds
)

// Rigging to the ground.
const (
	balloonFestivalFree     = iota // Free flight.
	balloonFestivalTethered        // Basket on the grass, three ropes to stakes.
	balloonFestivalHandled         // Just lifting, two handling lines held by the crew.
)

// balloonFestivalBalloon describes one upright balloon.
type balloonFestivalBalloon struct {
	at      Vec2    // Axis in plan.
	floor   float64 // Basket floor height.
	r       float64 // Envelope equator radius.
	pattern int
	colors  []Color // Pattern colors in order.
	cap     Color   // Crown cap and mouth band.
	count   int     // Gores, spiral arms, chequer columns, chevron teeth or bands.
	turn    float64 // Pattern rotation in radians.
	rig     int
}

var (
	// balloonFestivalFleet lists the upright balloons from the ground up.
	balloonFestivalFleet = []balloonFestivalBalloon{
		{P(-26, -10), 0, 11.5, balloonFestivalChequer, []Color{Red, White}, Navy, 12, 0.13, balloonFestivalTethered},
		{P(35, 2), 6, 12, balloonFestivalZigzag, []Color{Gold, Red, Gold, Navy}, Red, 8, 0.2, balloonFestivalHandled},
		{P(-9, -30), 13, 10, balloonFestivalSpiral, []Color{Navy, Glass, White}, White, 9, 0.4, balloonFestivalFree},
		{P(50, -16), 24, 10.5, balloonFestivalDiamonds, []Color{Water, Gold}, Navy, 8, 0.1, balloonFestivalFree},
		{P(-42, -25), 28, 10.5, balloonFestivalBands, []Color{Navy, White, Red, White}, Gold, 8, 0, balloonFestivalFree},
		{P(11, -41), 31, 13, balloonFestivalGores, []Color{Red, Terracotta, Gold, Grass, Water, Navy}, White, 12, 0.26, balloonFestivalFree},
		{P(-44, 26), 41, 10, balloonFestivalGores, []Color{Charcoal, Gold}, Red, 12, 0.13, balloonFestivalFree},
		{P(-22, -34), 46, 9, balloonFestivalSpiral, []Color{White, Red}, Red, 8, 1.1, balloonFestivalFree},
	}
	// balloonFestivalRiver is the river's course before smoothing, west to east.
	balloonFestivalRiver = []Vec2{P(-49, 31), P(-38, 27), P(-28, 31), P(-16, 27), P(-4, 30), P(8, 26),
		P(20, 29), P(30, 23), P(40, 21), P(50, 15), P(61, 10)}
	balloonFestivalField  = P(-4, -6)  // Launch field center.
	balloonFestivalLying  = P(-4, 13)  // Mouth of the balloon lying on its side.
	balloonFestivalBridge = P(-22, 29) // Footbridge center.
)

// balloonFestivalNoise is smooth value noise in [-1, 1].
func balloonFestivalNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	top, bottom := a+(b-a)*sx, c+(d-c)*sx
	return (top+(bottom-top)*sz)*2 - 1
}

// balloonFestivalProfile is the envelope radius at height h above the mouth for an
// equator radius r, shrunk by inset along the surface normal; negative outside.
func balloonFestivalProfile(r, h, inset float64) float64 {
	lower, crown := balloonFestivalDeep*r, (balloonFestivalTall-balloonFestivalDeep)*r
	mouth := 0.3 * r
	switch {
	case h < 0:
		return -1
	case h <= lower:
		// A cosine taper: straight near the mouth, swelling into the equator.
		t := (1 - h/lower) * math.Pi / 2
		slope := (r - mouth) * math.Sin(t) * math.Pi / 2 / lower
		return mouth + (r-mouth)*math.Cos(t) - inset*math.Sqrt(1+slope*slope)
	}
	t := (h - lower) / (crown - inset)
	if t >= 1 {
		return -1
	}
	return (r - inset) * math.Sqrt(1-t*t)
}

// balloonFestivalShape tabulates squared outer and inner envelope radii by height.
type balloonFestivalShape struct {
	r, height    float64
	outer, inner []float64
}

func balloonFestivalShapeOf(r float64) *balloonFestivalShape {
	s := &balloonFestivalShape{r: r, height: balloonFestivalTall * r}
	for i := 0; float64(i)*balloonFestivalStep < s.height; i++ {
		h := (float64(i) + 0.5) * balloonFestivalStep
		o := math.Max(0, balloonFestivalProfile(r, h, 0))
		in := math.Max(0, balloonFestivalProfile(r, h, balloonFestivalSkin))
		s.outer = append(s.outer, o*o)
		s.inner = append(s.inner, in*in)
	}
	return s
}

// at returns the squared outer and inner radii at height h above the mouth.
func (s *balloonFestivalShape) at(h float64) (outer, inner float64, ok bool) {
	i := int(math.Floor(h / balloonFestivalStep))
	if h < 0 || i >= len(s.outer) {
		return 0, 0, false
	}
	return s.outer[i], s.inner[i], true
}

// tint is the index into the balloon's colors at plan angle a and height h above the mouth.
func (b balloonFestivalBalloon) tint(a, h float64) int {
	n := len(b.colors)
	a = math.Mod(a+b.turn, 2*math.Pi)
	if a < 0 {
		a += 2 * math.Pi
	}
	u := a / (2 * math.Pi) * float64(b.count) // Column coordinate around the envelope.
	w := 2 * math.Pi * b.r / float64(b.count) // Column width at the equator.
	switch b.pattern {
	case balloonFestivalBands:
		return int(h/(balloonFestivalTall*b.r)*float64(b.count)) % n
	case balloonFestivalSpiral:
		return int(u+h/w) % n
	case balloonFestivalChequer:
		return (int(u) + int(h/w)) % n
	case balloonFestivalZigzag:
		band := 0.3 * b.r
		return int((h+0.85*band*math.Abs(2*(u-math.Floor(u))-1))/band) % n
	case balloonFestivalDiamonds:
		return (int(u+h/w) + int(u-h/w+64)) % n
	}
	return int(u) % n
}

// balloonFestivalSite carries the smoothed river course.
type balloonFestivalSite struct {
	river []Vec2
}

func balloonFestivalNewSite() *balloonFestivalSite {
	pts := balloonFestivalRiver
	// Two rounds of corner cutting turn the polyline into smooth meanders.
	for round := 0; round < 2; round++ {
		next := []Vec2{pts[0]}
		for i := 0; i+1 < len(pts); i++ {
			a, b := pts[i], pts[i+1]
			next = append(next, a.Add(b.Sub(a).Scale(0.25)), a.Add(b.Sub(a).Scale(0.75)))
		}
		pts = append(next, pts[len(pts)-1])
	}
	return &balloonFestivalSite{river: pts}
}

// course returns the plan distance from p to the river centerline, the arc length there
// and whether p lies on the launch-field (northern) bank.
func (s *balloonFestivalSite) course(p Vec2) (dist, along float64, north bool) {
	dist = math.Inf(1)
	run := 0.0
	for i := 0; i+1 < len(s.river); i++ {
		a, b := s.river[i], s.river[i+1]
		ab := b.Sub(a)
		l := ab.Length()
		t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/(l*l)))
		q := a.Add(ab.Scale(t))
		if d := p.Sub(q).Length(); d < dist {
			dist, along, north = d, run+t*l, p.Y < q.Y
		}
		run += l
	}
	return
}

// width is the river's half width at arc length along.
func (s *balloonFestivalSite) width(along float64) float64 {
	return 2.7 + 0.5*math.Sin(along*0.13)
}

// bank is the signed distance from the water's edge: negative in the river.
func (s *balloonFestivalSite) bank(p Vec2) (d float64, north bool) {
	dist, along, north := s.course(p)
	return dist - s.width(along), north
}

// rim is the plate's normalized radius: 1 on the edge of the meadow.
func (BalloonFestival) rim(x, z float64) float64 {
	return math.Hypot(x/balloonFestivalA, z/balloonFestivalB)
}

// cell returns the lattice cell under a world point.
func (BalloonFestival) cell(g *Grid, p Vec3) Cell { return Nearest(p.Scale(1 / g.Span())) }

// inField is the launch field's normalized radius: below 1 on the mown grass.
func (BalloonFestival) inField(x, z float64) float64 {
	return math.Hypot((x-balloonFestivalField.X)/40, (z-balloonFestivalField.Y)/29)
}

// promenade reports whether a plan point is on the spectators' path along the far bank.
func (f BalloonFestival) promenade(s *balloonFestivalSite, p Vec2) bool {
	d, north := s.bank(p)
	return north && d >= 1.4 && d <= 4.6 && p.X > -26 && p.X < 37
}

// plaza reports whether a plan point is on the trodden ground of the marquee village,
// the bridge approaches or the car park.
func (f BalloonFestival) plaza(p Vec2) bool {
	switch {
	case p.Sub(P(-43, 6)).Length() < 8: // Round marquee.
		return true
	case p.Sub(P(-47, -8)).Length() < 6.5:
		return true
	case p.X > -41 && p.X < -22.5 && p.Y > 11.5 && p.Y < 23: // Long marquee.
		return true
	case math.Abs(p.X-balloonFestivalBridge.X) < 1.8 && p.Y > 21 && p.Y < 37: // Bridge path.
		return true
	case p.X > -24 && p.X < 19 && p.Y > 33.6 && p.Y < 36: // Car park lane.
		return true
	case p.X > -18 && p.X < 18.5 && p.Y >= 36 && p.Y < 46: // Car park.
		return true
	}
	return false
}

func (f BalloonFestival) Build(g *Grid) Order {
	site := balloonFestivalNewSite()
	f.meadow(g, site)
	f.water(g, site)
	f.bridge(g)
	f.hedges(g, site)
	f.village(g)
	f.carPark(g)
	f.trees(g)
	f.lineside(g, site)
	for i, b := range balloonFestivalFleet {
		f.balloon(g, b, i)
	}
	f.lying(g)
	f.crowd(g)
	return Sweep(P(0, 0), math.Pi/2)
}

// meadow lays the grass plate: two cells thick, deeper around the rim, with a mown
// striped launch field, rough tufted grass and flower drifts outside it, and sand paths.
func (f BalloonFestival) meadow(g *Grid, s *balloonFestivalSite) {
	A, B := balloonFestivalA, balloonFestivalB
	g.Use(Grass).Fill(Func{Box{V(-A, -1.5, -B), V(A, 0, B)}, func(p Vec3) bool {
		e := f.rim(p.X, p.Z)
		return e <= 1 && (p.Y >= -0.5 || e > 0.962)
	}})
	top := Box{V(-A, -0.6, -B), V(A, 0.2, B)}
	// Earth shows around the rim below the turf.
	g.Paint(Wood, Func{Box{V(-A, -1.6, -B), V(A, -0.6, B)}, func(p Vec3) bool { return true }})
	// Mown stripes on the launch field.
	g.Paint(Foliage, Func{top, func(p Vec3) bool {
		return f.inField(p.X, p.Z) < 1 && math.Mod(p.X+203.5, 7) < 3.5
	}})
	// Rough meadow: darker tufts in soft patches, with drifts of buttercups and daisies.
	g.Paint(Foliage, Func{top, func(p Vec3) bool {
		return f.inField(p.X, p.Z) >= 1.04 && balloonFestivalNoise(p.X*0.17, p.Z*0.17, 11) > 0.28
	}})
	for i, col := range []Color{Gold, White} {
		g.Paint(col, Func{top, func(p Vec3) bool {
			c := f.cell(g, p)
			return p.Y > -0.2 && f.inField(p.X, p.Z) >= 1.06 &&
				balloonFestivalNoise(p.X*0.11+40*float64(i), p.Z*0.11, 17+i) > 0.3 &&
				Hash01(int(c.X), int(c.Z), 19+i) < 0.16
		}})
	}
	// Sand: the riverside promenade, the village plaza and the car park.
	g.Paint(Sand, Func{top, func(p Vec3) bool {
		return f.promenade(s, p.Plan()) || f.plaza(p.Plan())
	}})
	g.Paint(Stone, Func{top, func(p Vec3) bool {
		c := f.cell(g, p)
		return p.X > -18 && p.X < 18.5 && p.Z >= 36 && p.Z < 46 && Hash01(int(c.X), int(c.Z), 23) < 0.3
	}})
}

// water cuts the river into the plate: a sunken channel with shimmer along the current,
// sandy inner banks and reed clumps.
func (f BalloonFestival) water(g *Grid, s *balloonFestivalSite) {
	A, B := balloonFestivalA, balloonFestivalB
	inPlate := func(p Vec3) bool { return f.rim(p.X, p.Z) <= 1 }
	g.Carve(Func{Box{V(-A, -0.55, -B), V(A, 0.3, B)}, func(p Vec3) bool {
		d, _ := s.bank(p.Plan())
		return d < 0
	}})
	bed := Box{V(-A, -1.5, -B), V(A, -0.6, B)}
	g.Use(Water).Fill(Func{bed, func(p Vec3) bool {
		d, _ := s.bank(p.Plan())
		return d < 1 && inPlate(p)
	}})
	g.Paint(Glass, Func{bed, func(p Vec3) bool {
		dist, along, north := s.course(p.Plan())
		side := dist
		if north {
			side = -dist
		}
		return dist < s.width(along)-0.6 && balloonFestivalNoise(along*0.22, side*0.9, 31) > 0.3
	}})
	// Beaches where the bank is low, and reeds standing at the water's edge.
	top := Box{V(-A, -0.6, -B), V(A, 0.2, B)}
	g.Paint(Sand, Func{top, func(p Vec3) bool {
		d, _ := s.bank(p.Plan())
		return d >= 0 && d < 1.3 && balloonFestivalNoise(p.X*0.2, p.Z*0.2, 33) > 0.1
	}})
	g.Use(Foliage).Fill(Func{Box{V(-A, 0.2, -B), V(A, 1.3, B)}, func(p Vec3) bool {
		d, north := s.bank(p.Plan())
		if d < 0.2 || d > 1.1 || !inPlate(p) || f.rim(p.X, p.Z) > 0.95 || math.Abs(p.X-balloonFestivalBridge.X) < 3.5 {
			return false
		}
		seed := 35
		if north {
			seed = 36
		}
		return balloonFestivalNoise(p.X*0.45, p.Z*0.45, seed) > 0.42
	}})
}

// bridge is an arched wooden footbridge with white handrails.
func (f BalloonFestival) bridge(g *Grid) {
	c := balloonFestivalBridge
	const half = 5.5
	deck := func(z float64) float64 {
		k := (z - c.Y) / half
		return 0.2 + 1.3*(1-k*k)
	}
	g.Use(Wood).Fill(Func{Box{V(c.X-1.6, 0, c.Y-half), V(c.X+1.6, 2.2, c.Y+half)}, func(p Vec3) bool {
		y := deck(p.Z)
		return p.Y <= y && p.Y >= y-0.8
	}})
	for _, dx := range []float64{-1.6, 1.6} {
		var prev Vec3
		for i := 0; i <= 8; i++ {
			z := c.Y - half + 2*half*float64(i)/8
			q := V(c.X+dx, deck(z)+1.4, z)
			if i > 0 {
				g.Use(White).Fill(Capsule{prev, q, 0.45})
			}
			if i%2 == 0 {
				g.Use(Wood).Fill(Capsule{V(c.X+dx, deck(z)-0.4, z), q, 0.45})
			}
			prev = q
		}
	}
}

// hedges ring the meadow just inside the rim, broken by gates, the river and the car park.
func (f BalloonFestival) hedges(g *Grid, s *balloonFestivalSite) {
	A, B := balloonFestivalA, balloonFestivalB
	in := func(p Vec3) bool {
		e := f.rim(p.X, p.Z)
		if e < 0.918 || e > 0.95 {
			return false
		}
		if d, _ := s.bank(p.Plan()); d < 3 {
			return false
		}
		a := math.Atan2(p.Z/B, p.X/A)
		switch {
		case p.Z > 30 && math.Abs(p.X) < 26: // Car park frontage.
			return false
		case math.Abs(a+math.Pi/2) < 0.09: // Back gate.
			return false
		case math.Abs(a+0.35) < 0.07: // East gate.
			return false
		}
		return true
	}
	g.Use(Foliage).Fill(Func{Box{V(-A, 0.2, -B), V(A, 2.0, B)}, func(p Vec3) bool {
		return in(p) && p.Y <= 1.5+0.5*balloonFestivalNoise(p.X*0.3, p.Z*0.3, 41)
	}})
	g.Paint(Grass, Func{Box{V(-A, 1.0, -B), V(A, 2.2, B)}, func(p Vec3) bool {
		return in(p) && balloonFestivalNoise(p.X*0.35, p.Z*0.35, 43) > 0.25
	}})
}

// tent raises a round marquee: striped wall and cone roof, scalloped valance, pennant.
func (f BalloonFestival) tent(g *Grid, c Vec2, r, wall, roof float64, a, b Color, stripes int, door float64) {
	stripe := func(p Vec3) bool {
		t := p.Plan().Sub(c).Angle()/(2*math.Pi)*float64(stripes) + 100
		return int(math.Floor(t))%2 == 0
	}
	ring := Cylinder{Center: c, Radius: r, Inner: r - 1, Y0: 0.2, Y1: wall}
	g.Use(a).Fill(ring)
	g.Paint(b, Func{ring.Bounds(), stripe})
	cone := Subtract(Frustum{c, r + 1, 0, wall, wall + roof}, Frustum{c, r + 1, 0, wall - 1.2, wall + roof - 1.2})
	g.Use(a).FillVolume(cone)
	g.Paint(b, Func{cone.Bounds().Grow(0.3), func(p Vec3) bool { return p.Y > wall-0.2 && stripe(p) }})
	g.Use(Gold).Fill(Cylinder{Center: c, Radius: r + 1.1, Inner: r + 0.1, Y0: wall - 0.7, Y1: wall + 0.1})
	dc := Polar(c, r-0.5, door)
	g.Carve(Translate(RotateY(BlockAt(P(0, 0), 2.4, 2.4, 0.3, 2.6), P(0, 0), door), dc.At(0)))
	apex := wall + roof
	g.Use(White).Fill(Capsule{c.At(apex - 1), c.At(apex + 2.6), 0.45})
	g.Use(b).Fill(Block(V(c.X+0.4, apex+1.3, c.Y-0.3), V(c.X+2.4, apex+2.6, c.Y+0.3)))
}

// village pitches the marquees on the western side of the field.
func (f BalloonFestival) village(g *Grid) {
	f.tent(g, P(-43, 6), 5, 3.4, 4.6, Red, White, 16, 0.2)
	f.tent(g, P(-47, -8), 4, 3, 4, Navy, Gold, 12, 0.5)
	// Long marquee: white walls under a striped ridge roof, open doorways to the south.
	c := P(-32, 17)
	const hx, hz, wall = 7.0, 3.5, 3.0
	g.Use(White).Fill(Block(V(c.X-hx, 0.2, c.Y-hz), V(c.X+hx, wall, c.Y+hz)))
	g.Carve(Block(V(c.X-hx+1, 0.2, c.Y-hz+1), V(c.X+hx-1, wall+0.2, c.Y+hz-1)))
	roof := Gable{P(c.X-hx-0.5, c.Y-hz-0.5), P(c.X+hx+0.5, c.Y+hz+0.5), wall, AxisX}
	g.Use(White).Fill(roof)
	g.Carve(Gable{P(c.X-hx+0.6, c.Y-hz+0.6), P(c.X+hx-0.6, c.Y+hz-0.6), wall - 0.1, AxisX})
	g.Paint(Water, Func{roof.Bounds().Grow(0.3), func(p Vec3) bool {
		return p.Y > wall-0.2 && math.Mod(p.X-c.X+hx+40.5, 4) < 2
	}})
	for _, dx := range []float64{-3.5, 3.5} {
		g.Carve(Arch{Sill: V(c.X+dx, 0.3, c.Y+hz-0.5), HalfWidth: 1.2, Spring: 1.2, Depth: 1.6, Through: AxisZ})
	}
	for i, dx := range []float64{-hx, 0, hx} {
		top := V(c.X+dx, wall+hz+0.4, c.Y)
		g.Use(White).Fill(Capsule{top, top.Add(V(0, 2.6, 0)), 0.45})
		g.Use([]Color{Red, Gold, Red}[i]).Fill(Block(top.Add(V(0.4, 1.4, -0.3)), top.Add(V(2.2, 2.6, 0.3))))
	}
}

// van parks a chase van with its nose toward -Z; trailer adds a low basket trailer.
func (f BalloonFestival) van(g *Grid, at Vec2, body Color, trailer bool, load Color) {
	x, z := at.X, at.Y
	g.Use(body).Fill(Block(V(x-1.2, 0.6, z-2.5), V(x+1.2, 3, z+2.5)))
	g.Carve(Block(V(x-1.3, 2.2, z-2.6), V(x+1.3, 3.1, z-1.9)))
	g.Paint(Glass, Block(V(x-1.2, 1.6, z-2), V(x+1.2, 2.5, z-1.4)))
	g.Paint(White, Block(V(x-1.2, 2.6, z-1.5), V(x+1.2, 3.1, z+2.5)))
	g.Use(Charcoal)
	for _, dz := range []float64{-1.5, 1.5} {
		g.Fill(Block(V(x-1.3, 0.1, z+dz-0.5), V(x+1.3, 0.9, z+dz+0.5)))
	}
	if !trailer {
		return
	}
	g.Fill(Block(V(x-0.3, 0.6, z+2.5), V(x+0.3, 1, z+3.6)))
	g.Fill(Block(V(x-1.3, 0.1, z+4.6), V(x+1.3, 0.9, z+5.6)))
	g.Use(Slate).Fill(Block(V(x-1.1, 0.6, z+3.6), V(x+1.1, 1.5, z+6.6)))
	g.Use(load).Fill(Ellipsoid{V(x, 1.9, z+5.1), V(1, 0.9, 1.3)})
}

// carPark lines the chase vans up on the near bank, some with trailers still hitched.
func (f BalloonFestival) carPark(g *Grid) {
	bodies := []Color{White, Red, Navy, Gold, White, Terracotta, Slate}
	loads := []Color{Red, Gold, Water, Sand, Navy, White, Terracotta}
	for i, body := range bodies {
		x := -15 + 5*float64(i)
		f.van(g, P(x, 39), body, i >= 1 && i <= 4, loads[i])
	}
}

// tree plants a round-crowned tree of total height h and crown radius r.
func (f BalloonFestival) tree(g *Grid, at Vec2, h, r float64, seed int) {
	g.Use(Wood).Fill(Capsule{at.At(0.2), at.At(h - r), 0.6})
	c := at.At(h - r)
	lobe := c.Add(V((Hash01(seed, 1)-0.5)*r, -0.25*r, (Hash01(seed, 2)-0.5)*r))
	crown := Union(
		Subtract(Ellipsoid{c, V(r, 0.95*r, r)}, Ellipsoid{c, V(r-1.1, 0.95*r-1.1, r-1.1)}),
		Subtract(Sphere(lobe, 0.72*r), Sphere(lobe, 0.72*r-1.1)),
	)
	g.Use(Foliage).FillVolume(crown)
	// Sunlit side.
	g.Paint(Grass, Intersect(Sphere(c.Add(V(0.55*r, 0.6*r, 0.2*r)), 0.85*r), Ellipsoid{c, V(r+0.4, r+0.4, r+0.4)}))
}

// trees stand along the hedges and at both ends of the near bank.
func (f BalloonFestival) trees(g *Grid) {
	list := []struct {
		at   Vec2
		h, r float64
	}{
		{P(-33, -33), 10, 3.6}, {P(-17, -40), 8.5, 3.1}, {P(-2, -43), 9.5, 3.4}, {P(25, -38), 10.5, 3.8},
		{P(37, -31), 8, 3}, {P(46, -21), 9.5, 3.5}, {P(52, -9), 8, 3}, {P(-51, -17), 9, 3.3},
		{P(-53, 3), 8, 3}, {P(-50, 16), 9.5, 3.4}, {P(-37, 37), 7.5, 2.9}, {P(-29, 39.5), 8.5, 3.2},
		{P(26, 38), 8, 3.1}, {P(34, 33), 7, 2.7}, {P(43, 27), 8, 3}, {P(52, 6), 7.5, 2.8},
	}
	for i, t := range list {
		f.tree(g, t.at, t.h, t.r, i)
	}
}

// figure stands a tiny person on height y: dark legs, a colored top and a sand head.
func (f BalloonFestival) figure(g *Grid, at Vec2, y float64, shirt Color, tall bool) {
	g.Use(Charcoal).AddAt(at.At(y + 0.4))
	g.Use(shirt).AddAt(at.At(y + 0.8))
	head := y + 1.2
	if tall {
		g.AddAt(at.At(head))
		head += 0.4
	}
	g.Use(Sand).AddAt(at.At(head))
}

// lineside dresses the edge of the launch field: a line of posts along the promenade
// with a flag every few bays, and a windsock on the eastern side.
func (f BalloonFestival) lineside(g *Grid, s *balloonFestivalSite) {
	flags := []Color{Red, Gold, Navy, White, Water, Terracotta, Red, Gold}
	run, next, n := 0.0, 0.0, 0
	for i := 0; i+1 < len(s.river); i++ {
		a, b := s.river[i], s.river[i+1]
		ab := b.Sub(a)
		l := ab.Length()
		side := P(ab.Y, -ab.X).Scale(1 / l) // Toward the field.
		for ; next <= run+l; next += 2.6 {
			q := a.Add(ab.Scale((next - run) / l))
			p := q.Add(side.Scale(s.width(next) + 5.4))
			if p.X < -25 || p.X > 36 || math.Abs(p.X-balloonFestivalBridge.X) < 2.2 {
				continue
			}
			n++
			if n%4 == 2 {
				g.Use(White).Fill(Capsule{p.At(0.2), p.At(8.2), 0.5})
				g.Use(Gold).Fill(Sphere(p.At(8.6), 0.6))
				g.Use(flags[(n/4)%len(flags)]).Fill(Block(V(p.X+0.4, 5.6, p.Y-0.3), V(p.X+3, 7.8, p.Y+0.3)))
				continue
			}
			g.Use(White).Fill(Capsule{p.At(0.2), p.At(1.1), 0.4})
			g.Use(Red).AddAt(p.At(1.5))
		}
		run += l
	}
	// Windsock.
	w := P(-4, -22)
	g.Use(White).Fill(Capsule{w.At(0.2), w.At(9), 0.5})
	sock := Func{Box{V(w.X, 7, w.Y-2), V(w.X+6, 10.4, w.Y+2)}, func(p Vec3) bool {
		t := (p.X - w.X - 0.3) / 5.4
		if t < 0 || t > 1 {
			return false
		}
		return math.Hypot(p.Y-(8.9-0.5*t), p.Z-w.Y) <= 1.25-0.7*t
	}}
	g.Use(Red).Fill(sock)
	g.Paint(White, Func{sock.Box, func(p Vec3) bool { return math.Mod(p.X-w.X+0.3, 2.4) > 1.2 }})
}

// balloon builds one upright balloon: envelope, skirt, burner, load lines and basket.
func (f BalloonFestival) balloon(g *Grid, b balloonFestivalBalloon, seed int) {
	const basketH = 2.8
	half := 1.5
	if b.r >= 10.5 {
		half = 2
	}
	c := b.at
	rim := b.floor + basketH
	mouthY := rim + 0.35*b.r + 2
	mouth := 0.3 * b.r
	shape := balloonFestivalShapeOf(b.r)

	// Envelope: a hollow skin, painted by angle and height.
	box := Box{V(c.X-b.r, mouthY, c.Y-b.r), V(c.X+b.r, mouthY+shape.height, c.Y+b.r)}
	g.Use(b.colors[0]).FillVolume(Func{box, func(p Vec3) bool {
		o, in, ok := shape.at(p.Y - mouthY)
		if !ok {
			return false
		}
		dx, dz := p.X-c.X, p.Z-c.Y
		d := dx*dx + dz*dz
		return d <= o && d > in
	}})
	height := func(p Vec3) float64 {
		return math.Max(0, math.Min(shape.height-balloonFestivalStep, p.Y-mouthY))
	}
	near := func(p Vec3) bool {
		o, _, _ := shape.at(height(p))
		d := p.Plan().Sub(c).Length()
		return d <= math.Sqrt(o)+0.6
	}
	wide := box.Grow(0.5)
	for i := 1; i < len(b.colors); i++ {
		if b.colors[i] == b.colors[0] {
			continue
		}
		g.Paint(b.colors[i], Func{wide, func(p Vec3) bool {
			return near(p) && b.colors[b.tint(p.Plan().Sub(c).Angle(), height(p))] == b.colors[i]
		}})
	}
	g.Paint(b.cap, Func{wide, func(p Vec3) bool {
		h := height(p)
		return near(p) && (h < 0.1*b.r+0.6 || h > balloonFestivalDeep*b.r && p.Plan().Sub(c).Length() < 0.3*b.r)
	}})

	// Skirt below the mouth.
	g.Use(Charcoal).FillVolume(Subtract(
		Frustum{c, 0.8 * mouth, mouth, mouthY - 1.2, mouthY + 0.3},
		Frustum{c, 0.8*mouth - 0.9, mouth - 0.9, mouthY - 1.3, mouthY + 0.4}))

	// Load lines from the basket corners to the skirt, tied by the burner frame.
	inner := (half - 0.3) * math.Sqrt2
	lineTop := mouthY - 0.9
	frameY := rim + 1.3
	for k := 0; k < 4; k++ {
		a := math.Pi/4 + float64(k)*math.Pi/2
		g.Use(Charcoal).Fill(Capsule{Polar(c, inner, a).At(rim - 0.3), Polar(c, 0.85*mouth, a).At(lineTop), 0.5})
	}
	t := (frameY - rim + 0.3) / (lineTop - rim + 0.3)
	fr := (inner+(0.85*mouth-inner)*t)/math.Sqrt2 + 0.25
	g.Use(Slate).Fill(Subtract(BlockAt(c, 2*fr, 2*fr, frameY-0.3, frameY+0.3), BlockAt(c, 2*fr-1.6, 2*fr-1.6, frameY-0.4, frameY+0.4)))
	g.Fill(BlockAt(c, 2*fr, 0.8, frameY-0.3, frameY+0.3))
	g.Use(Stone).Fill(Cylinder{Center: c, Radius: 0.7, Y0: frameY + 0.3, Y1: frameY + 0.9})
	g.Use(Red).Fill(Frustum{c, 0.95, 0.7, frameY + 0.9, frameY + 1.9})
	g.Use(Gold).Fill(Frustum{c, 0.7, 0.35, frameY + 1.9, mouthY + 0.8})

	// Wicker basket with leather trim, and passengers looking over the rim.
	g.Use(Sand).Fill(BlockAt(c, 2*half, 2*half, b.floor, rim))
	g.Carve(BlockAt(c, 2*half-1.6, 2*half-1.6, b.floor+1.4, rim+0.2))
	g.Paint(Wood, BlockAt(c, 2*half+0.2, 2*half+0.2, rim-0.55, rim+0.1), BlockAt(c, 2*half+0.2, 2*half+0.2, b.floor-0.1, b.floor+0.45))
	shirts := []Color{Red, Navy, White, Gold, Water, Terracotta}
	d := half - 1.0
	for k, o := range []Vec2{P(-d, -d), P(d, d), P(d, -d)} {
		if k == 2 && half < 2 {
			break
		}
		q := c.Add(o)
		g.Use(shirts[(seed+2*k)%len(shirts)]).AddAt(q.At(rim - 0.2))
		g.AddAt(q.At(rim + 0.2))
		g.Use(Sand).AddAt(q.At(rim + 0.6))
	}

	switch b.rig {
	case balloonFestivalTethered:
		for k := 0; k < 3; k++ {
			a := 1.75 + float64(k)*2*math.Pi/3
			stake := Polar(c, b.r+4, a)
			g.Use(White).Fill(Capsule{Polar(c, 0.9*mouth, a).At(mouthY - 0.8), stake.At(0.5), 0.5})
			g.Use(Wood).Fill(Cylinder{Center: stake, Radius: 0.6, Y0: 0.2, Y1: 1.4})
			f.figure(g, Polar(c, b.r+5.6, a+0.12), 0, shirts[(seed+k)%len(shirts)], k != 1)
		}
		for k := 0; k < 4; k++ {
			a := 0.4 + float64(k)*1.3
			f.figure(g, Polar(c, half+2.2+0.5*float64(k%2), a), 0, shirts[(seed+k+3)%len(shirts)], k%2 == 0)
		}
	case balloonFestivalHandled:
		for k, sx := range []float64{-1, 1} {
			hand := P(c.X+sx*(half+5.5), c.Y+2.5*sx)
			g.Use(White).Fill(Capsule{V(c.X+sx*half, b.floor+0.3, c.Y), hand.At(1.3), 0.5})
			f.figure(g, hand.Add(P(sx*0.8, 0)), 0, shirts[(seed+k)%len(shirts)], true)
			f.figure(g, hand.Add(P(sx*0.2, 1.1)), 0, shirts[(seed+k+2)%len(shirts)], false)
		}
	}
}

// lying builds the half-inflated balloon on its side: a sagging teardrop bag on the
// grass, its tipped basket and burner firing into the mouth, an inflation fan and crew.
func (f BalloonFestival) lying(g *Grid) {
	const r = 9.5
	m := balloonFestivalLying
	dir := P(0.778, -0.629)
	dir = dir.Scale(1 / dir.Length())
	side := P(-dir.Y, dir.X)
	shape := balloonFestivalShapeOf(r)
	// Local frame: s along the axis from the mouth, w across, y up.
	local := func(p Vec3) (float64, float64) {
		q := p.Plan().Sub(m)
		return q.Dot(dir), q.Dot(side)
	}
	world := func(s, w, y float64) Vec3 { return m.Add(dir.Scale(s)).Add(side.Scale(w)).At(y) }
	// The bag is wider than tall and sags further toward the crown.
	sag := func(s float64) float64 { k := s / shape.height; return 0.64 - 0.2*k*k }
	skin := func(p Vec3, shell bool) bool {
		s, w := local(p)
		o2, i2, ok := shape.at(s)
		if !ok {
			return false
		}
		o, in := math.Sqrt(o2), math.Sqrt(i2)
		a, b := 1.06*o, sag(s)*o+0.5
		dy := p.Y - b
		if (w*w)/(a*a)+(dy*dy)/(b*b) > 1 {
			return false
		}
		if !shell {
			return true
		}
		ai, bi := a-(o-in), b-(o-in)
		return in <= 0 || ai <= 0 || bi <= 0 || (w*w)/(ai*ai)+(dy*dy)/(bi*bi) > 1
	}
	reach := shape.height + 1
	box := BoxOf(world(-1, -1.1*r, 0.2), world(reach, 1.1*r, 0.2)).Union(BoxOf(world(-1, 1.1*r, 0.2), world(reach, -1.1*r, 1.5*r)))
	g.Use(Terracotta).FillVolume(Func{box, func(p Vec3) bool { return p.Y >= 0.2 && skin(p, true) }})
	wide := box.Grow(0.5)
	near := func(p Vec3) bool {
		q := p
		q.Y = math.Max(0.2, q.Y-0.4)
		return skin(q, false) || skin(p, false)
	}
	gore := func(p Vec3) int {
		s, w := local(p)
		o2, _, _ := shape.at(math.Max(0, math.Min(shape.height-balloonFestivalStep, s)))
		b := sag(s)*math.Sqrt(o2) + 0.5
		a := math.Atan2(p.Y-b, w) + math.Pi
		return int(a/(2*math.Pi)*12+0.5) % 12
	}
	g.Paint(Gold, Func{wide, func(p Vec3) bool { return near(p) && gore(p)%2 == 1 }})
	g.Paint(White, Func{wide, func(p Vec3) bool { return near(p) && gore(p)%4 == 3 }})
	g.Paint(Navy, Func{wide, func(p Vec3) bool {
		s, _ := local(p)
		return near(p) && (s < 1.6 || s > shape.height-0.26*r)
	}})

	// Tipped basket, load lines and the burner blasting into the mouth.
	turn := func(s Solid) Solid { return RotateY(s, m, dir.Angle()) }
	g.Use(Sand).Fill(turn(Block(V(m.X-7.2, 0.2, m.Y-1.5), V(m.X-4.2, 3.2, m.Y+1.5))))
	g.Carve(turn(Block(V(m.X-6, 0.9, m.Y-0.8), V(m.X-4.1, 2.5, m.Y+0.8))))
	g.Paint(Wood, turn(Block(V(m.X-4.7, 0.1, m.Y-1.6), V(m.X-4.1, 3.3, m.Y+1.6))),
		turn(Block(V(m.X-7.3, 0.1, m.Y-1.6), V(m.X-6.8, 3.3, m.Y+1.6))))
	for _, sw := range []float64{-1, 1} {
		for _, y := range []float64{0.7, 2.9} {
			g.Use(Charcoal).Fill(Capsule{world(-4.3, 1.2*sw, y), world(0.6, 2.6*sw, y+0.4), 0.5})
		}
	}
	g.Use(Slate).Fill(turn(Block(V(m.X-3.9, 1.1, m.Y-0.8), V(m.X-2.9, 2.5, m.Y+0.8))))
	g.Use(Red).Fill(Capsule{world(-2.6, 0, 1.9), world(-1.4, 0, 2), 0.75})
	g.Use(Gold).Fill(Capsule{world(-1, 0, 2), world(1.6, 0, 2.2), 0.6})

	// Inflation fan beside the basket, and the crew.
	fan := world(-4.5, 5.2, 1.8)
	g.Use(Charcoal).Fill(Func{BoxOf(fan, fan).Grow(2.2), func(p Vec3) bool {
		s, w := local(p)
		rr := math.Hypot(w-5.2, p.Y-1.8)
		return math.Abs(s+4.5) <= 0.6 && rr <= 1.7 && rr >= 0.9
	}})
	g.Use(Stone).Fill(Capsule{world(-4.5, 5.2, 1.8), world(-4.2, 5.2, 1.8), 0.6})
	g.Use(Charcoal).Fill(Capsule{world(-4.5, 5.2, 0.2), world(-4.5, 5.2, 0.4), 0.5})
	g.Use(White).Fill(Capsule{world(shape.height-0.6, 0, 0.9), world(shape.height+6.5, 0.6, 1.2), 0.5})
	shirts := []Color{Red, Navy, White, Water, Gold, Red}
	for i, q := range []Vec3{world(0.6, 4.6, 0), world(0.4, -4.6, 0), world(-5.5, -3.2, 0), world(shape.height+7.4, 0.6, 0),
		world(shape.height+6.4, 1.7, 0), world(-8.6, 1.2, 0)} {
		f.figure(g, q.Plan(), 0, shirts[i], i%2 == 0)
	}
}

// crowd scatters tiny spectators over the sand: the promenade, the village and the car park.
func (f BalloonFestival) crowd(g *Grid) {
	span := g.Span()
	shirts := []Color{Red, Navy, Gold, White, Water, Terracotta, Brick, Slate}
	for x := -55.0; x <= 55; x += 1.5 {
		for z := -30.0; z <= 46; z += 1.5 {
			i, j := int(math.Floor(x*2)), int(math.Floor(z*2))
			jx := x + (Hash01(i, j, 61)-0.5)*1.0
			jz := z + (Hash01(i, j, 62)-0.5)*1.0
			if Hash01(i, j, 63) > 0.34 {
				continue
			}
			ground, ok := g.ColorOf(Nearest(V(jx, 0, jz).Scale(1 / span)))
			if !ok || (ground != Sand && ground != Stone) {
				continue
			}
			blocked := false
			for _, dx := range []float64{-0.8, 0, 0.8} {
				for _, dz := range []float64{-0.8, 0, 0.8} {
					for _, y := range []float64{0.4, 0.8, 1.2, 1.6, 2.0, 2.4} {
						if g.HasAt(V(jx+dx, y, jz+dz)) {
							blocked = true
						}
					}
				}
			}
			if blocked {
				continue
			}
			f.figure(g, P(jx, jz), 0, shirts[int(Hash01(i, j, 65)*float64(len(shirts)))], Hash01(i, j, 64) < 0.7)
		}
	}
}
