package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// StatueOfLiberty is Liberty Island in New York Harbor. The robed copper figure in
// verdigris patina raises the torch with its gold flame in her right hand and holds the
// tablet in her left, under the seven-ray crown. She stands on Hunt's granite pedestal
// with its loggias, shield band and balcony, on the stepped concrete foundation inside
// the eleven-pointed star of Fort Wood. Lawns, tree groves, paths, the museum, the
// flagpole and the ferry dock with a moored ferry fill the island, ringed by a granite
// seawall and promenade in a harbor water layer. The statue faces +Z.
type StatueOfLiberty struct{}

func init() { Register(StatueOfLiberty{}) }

func (StatueOfLiberty) Name() string  { return "statue_of_liberty" }
func (StatueOfLiberty) Title() string { return "Statue of Liberty" }

const (
	statueOfLibertyWater   = 0.6  // Top of the harbor layer.
	statueOfLibertyGround  = 1.6  // Island lawn level.
	statueOfLibertyFortTop = 7.6  // Fort Wood terrace.
	statueOfLibertyFound   = 15.0 // Top of the concrete foundation.
	statueOfLibertyPedTop  = 39.0 // Pedestal deck: the statue's feet.
	statueOfLibertyStarOut = 21.0 // Star point radius.
	statueOfLibertyStarIn  = 16.0 // Star notch radius.
	statueOfLibertyK       = 1.4  // Statue scale over its authored local units.
	statueOfLibertyScale   = 1.1  // Whole-site scale over the authored units.
	statueOfLibertyGroove  = Water
)

// star is Fort Wood's eleven-pointed plan, a point facing +Z.
func (StatueOfLiberty) star() []Vec2 {
	var pts []Vec2
	for k := 0; k < 22; k++ {
		r := statueOfLibertyStarOut
		if k%2 == 1 {
			r = statueOfLibertyStarIn
		}
		pts = append(pts, Polar(P(0, 0), r, math.Pi/2+float64(k)*2*math.Pi/22))
	}
	return pts
}

// polyDist is the signed plan distance to a polygon's boundary, positive inside.
func (StatueOfLiberty) polyDist(poly []Vec2, p Vec2) float64 {
	d := math.Inf(1)
	for i, j := 0, len(poly)-1; i < len(poly); j, i = i, i+1 {
		a, b := poly[j], poly[i]
		ab := b.Sub(a)
		t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
		d = math.Min(d, p.Sub(a.Add(ab.Scale(t))).Length())
	}
	if InsidePolygon(poly, p) {
		return d
	}
	return -d
}

// islandQ is the normalized island radius at a plan point (< 1 on the island).
func (StatueOfLiberty) islandQ(p Vec2) float64 {
	c := P(1, -2)
	d := p.Sub(c)
	a := d.Angle()
	wobble := 1 + 0.035*math.Sin(3*a+1) + 0.025*math.Sin(5*a+2.3) + 0.02*math.Sin(7*a+0.4)
	x, z := d.X/40, d.Y/31
	return math.Hypot(x, z) / wobble
}

// islandEdge approximates the plan distance in from the shore.
func (s StatueOfLiberty) islandEdge(p Vec2) float64 { return (1 - s.islandQ(p)) * 31 }

// segDist is the plan distance from p to segment ab.
func (StatueOfLiberty) segDist(p, a, b Vec2) float64 {
	ab := b.Sub(a)
	t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
	return p.Sub(a.Add(ab.Scale(t))).Length()
}

// lerpKnots interpolates a piecewise-linear profile.
func (StatueOfLiberty) lerpKnots(ys, vs []float64, y float64) float64 {
	if y <= ys[0] {
		return vs[0]
	}
	for i := 1; i < len(ys); i++ {
		if y <= ys[i] {
			t := (y - ys[i-1]) / (ys[i] - ys[i-1])
			return vs[i-1] + (vs[i]-vs[i-1])*t
		}
	}
	return vs[len(vs)-1]
}

// Statue profile knots (local height above the pedestal deck).
var (
	statueOfLibertyKY = []float64{0, 1.5, 6, 12, 14.5, 17, 19.5}
	statueOfLibertyKA = []float64{5.0, 4.85, 4.6, 4.1, 3.9, 4.3, 4.6}
	statueOfLibertyKB = []float64{4.2, 4.0, 3.7, 3.3, 3.1, 3.3, 3.4}
)

// hem is the height of the overgarment's diagonal lower edge, lowest on her left (+X).
func (StatueOfLiberty) hem(x float64) float64 { return 10.5 - 0.55*x }

// fold is a drapery modulation: rounded ridges and sharp grooves, in [-1, 1].
func (StatueOfLiberty) fold(theta, y float64, over bool) float64 {
	if over {
		return 2*math.Abs(math.Cos(4.5*theta+0.32*y)) - 1
	}
	return 2*math.Abs(math.Cos(5.5*theta+0.7*math.Sin(0.3*y))) - 1
}

// robeQ returns the normalized radius of a local point in the robe and the fold value.
func (s StatueOfLiberty) robeQ(p Vec3) (q, f float64, over bool) {
	a := s.lerpKnots(statueOfLibertyKY, statueOfLibertyKA, p.Y)
	b := s.lerpKnots(statueOfLibertyKY, statueOfLibertyKB, p.Y)
	theta := math.Atan2(p.Z/b, p.X/a)
	over = p.Y >= s.hem(p.X) && p.Y <= 20
	f = s.fold(theta, p.Y, over)
	return math.Hypot(p.X/a, p.Z/b), f, over
}

// figure is the statue body in local coordinates (feet at y = 0, facing +Z).
func (s StatueOfLiberty) figure(p Vec3) bool {
	if p.Y < 0 {
		return false
	}
	if p.Y <= 19.5 {
		q, f, over := s.robeQ(p)
		limit := 1 + 0.13*f
		if over {
			limit = 1.12 + 0.1*f
		}
		if p.Y < 1.2 { // Hem flares a little over the feet.
			limit += 0.04 * (1.2 - p.Y)
		}
		if q <= limit {
			return true
		}
	}
	// Torso and shoulders.
	d := p.Sub(V(0, 19.5, 0))
	if (d.X/4.7)*(d.X/4.7)+(d.Y/4.3)*(d.Y/4.3)+(d.Z/3.5)*(d.Z/3.5) <= 1 {
		return true
	}
	// Drapery cascading from her left forearm.
	d = p.Sub(V(4.9, 10, 0.5))
	if (d.X/1.7)*(d.X/1.7)+(d.Y/6.5)*(d.Y/6.5)+(d.Z/2.5)*(d.Z/2.5) <= 1 {
		return true
	}
	// Neck, head and hair.
	if p.Y >= 22 && p.Y <= 26.2 && math.Hypot(p.X, p.Z-0.1) <= 1.75 {
		return true
	}
	d = p.Sub(V(0, 27.5, 0.2))
	if (d.X/2.0)*(d.X/2.0)+(d.Y/2.5)*(d.Y/2.5)+(d.Z/2.3)*(d.Z/2.3) <= 1 {
		return true
	}
	d = p.Sub(V(0, 27.6, -0.6))
	if (d.X/2.3)*(d.X/2.3)+(d.Y/2.4)*(d.Y/2.4)+(d.Z/2.2)*(d.Z/2.2) <= 1 {
		return true
	}
	// Hair falling to the nape and the bun.
	d = p.Sub(V(0, 25.4, -1.3))
	if (d.X/2.0)*(d.X/2.0)+(d.Y/1.8)*(d.Y/1.8)+(d.Z/1.4)*(d.Z/1.4) <= 1 {
		return true
	}
	if p.Sub(V(0, 26.2, -2.5)).Length() <= 1.2 {
		return true
	}
	return false
}

// local maps a world point into the statue frame.
func (StatueOfLiberty) local(p Vec3) Vec3 {
	return V(p.X, p.Y-statueOfLibertyPedTop, p.Z).Scale(1 / statueOfLibertyK)
}

// at maps a local statue point into the world.
func (StatueOfLiberty) at(x, y, z float64) Vec3 {
	return V(x, y, z).Scale(statueOfLibertyK).Add(V(0, statueOfLibertyPedTop, 0))
}

// tablet is the tablet in her left hand as a local box test, with face coordinates.
func (StatueOfLiberty) tablet(p Vec3) (inside bool, u, v, w float64) {
	c := V(6.1, 19.4, 2.0)
	const t = 12 * math.Pi / 180
	n0 := V(math.Sin(40*math.Pi/180), 0, math.Cos(40*math.Pi/180))
	vAxis := V(n0.Z, 0, -n0.X)
	wAxis := V(0, math.Cos(t), 0).Sub(n0.Scale(math.Sin(t)))
	nAxis := n0.Scale(math.Cos(t)).Add(V(0, math.Sin(t), 0))
	d := p.Sub(c)
	u, v, w = d.Dot(nAxis), d.Dot(vAxis), d.Dot(wAxis)
	return math.Abs(u) <= 0.6 && math.Abs(v) <= 2.0 && math.Abs(w) <= 3.4, u, v, w
}

func (s StatueOfLiberty) Build(g *Grid) Order {
	// Author on a finer lattice, then reinterpret its cells at the grid's span: the
	// whole site comes out statueOfLibertyScale times larger than its authored units.
	a := NewGridWithSpan(g.Span() / statueOfLibertyScale)
	s.harbor(a)
	s.fort(a)
	s.foundation(a)
	s.pedestal(a)
	s.statue(a)
	s.grounds(a)
	s.ferry(a)
	g.Merge(a)
	return Sweep(P(0, 0), math.Pi/2)
}

// harbor lays the water ring, the island with its seawall, lawns and paths.
func (s StatueOfLiberty) harbor(g *Grid) {
	g.Use(Water).Fill(Func{Box{V(-52, 0, -42), V(58, statueOfLibertyWater, 38)}, func(p Vec3) bool {
		d := p.Plan().Sub(P(3, -2))
		return math.Hypot(d.X/54, d.Y/39) <= 1 && s.islandQ(p.Plan()) > 1
	}})
	island := Func{Box{V(-44, statueOfLibertyWater, -38), V(46, statueOfLibertyGround, 34)}, func(p Vec3) bool {
		return s.islandQ(p.Plan()) <= 1
	}}
	g.Use(Grass).Fill(island)
	// Granite seawall and the promenade along it.
	g.Paint(DarkStone, Func{island.Box, func(p Vec3) bool {
		e := s.islandEdge(p.Plan())
		return e >= 0 && e < 0.9
	}})
	g.Paint(Stone, Func{island.Box, func(p Vec3) bool {
		e := s.islandEdge(p.Plan())
		return e >= 0.9 && e < 3.2 && p.Y > statueOfLibertyGround-0.5
	}})
	// Promenade railing.
	g.Use(Charcoal).Fill(Func{Box{V(-44, statueOfLibertyGround, -38), V(46, statueOfLibertyGround+0.9, 34)}, func(p Vec3) bool {
		e := s.islandEdge(p.Plan())
		return e >= 0.5 && e < 1.0 && s.segDist(p.Plan(), P(38, -12), P(48, -12)) > 2.5
	}})
}

// fort builds Fort Wood: battered star walls, a hollow casemate interior with dark
// embrasures, the terrace deck, its parapet and the sally port on the north notch.
func (s StatueOfLiberty) fort(g *Grid) {
	star := s.star()
	const batter, thick, deck = 0.15, 1.4, 6.4
	top := statueOfLibertyFortTop
	box := Box{V(-23, statueOfLibertyGround-0.6, -23), V(23, top+1, 23)}
	wall := Func{box, func(p Vec3) bool {
		if p.Y < statueOfLibertyGround-0.6 || p.Y > top {
			return false
		}
		d := s.polyDist(star, p.Plan())
		out := batter * (top - p.Y)
		return d >= -out && (d <= thick || p.Y >= deck)
	}}
	g.Use(DarkStone).Fill(wall)
	// Coursed stone: lighter bands every two units.
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		d := s.polyDist(star, p.Plan())
		return d < 0.8 && math.Mod(p.Y-statueOfLibertyGround+20, 2.0) < 0.45 && p.Y < top-0.2
	}})
	// Terrace paving.
	g.Paint(Stone, Func{box, func(p Vec3) bool { return p.Y > top-0.5 && s.polyDist(star, p.Plan()) > 0.8 }})
	// Parapet with stone coping.
	g.Use(DarkStone).Fill(Func{box, func(p Vec3) bool {
		d := s.polyDist(star, p.Plan())
		return p.Y > top && p.Y <= top+1 && d >= 0 && d <= 0.9
	}})
	g.Paint(Stone, Func{box, func(p Vec3) bool { return p.Y > top+0.6 && s.polyDist(star, p.Plan()) >= -0.2 }})
	// Embrasures on every wall face, the sally port at the north notch.
	for i := 0; i < 22; i++ {
		a, b := star[i], star[(i+1)%22]
		mid := a.Add(b).Scale(0.5)
		dir := b.Sub(a).Scale(1 / b.Sub(a).Length())
		for _, t := range []float64{-2.2, 2.2} {
			c := mid.Add(dir.Scale(t))
			g.Carve(Func{Box{V(c.X-3, 3.4, c.Y-3), V(c.X+3, 5.4, c.Y+3)}, func(p Vec3) bool {
				q := p.Plan().Sub(c)
				return math.Abs(q.Dot(dir)) <= 0.55 && math.Abs(s.polyDist(star, p.Plan())) < 2.5
			}})
		}
	}
	g.Carve(Arch{Sill: V(0, statueOfLibertyGround, -statueOfLibertyStarIn), HalfWidth: 1.7, Spring: 2.2, Depth: 6, Through: AxisZ})
	// Dark casemates seen through the openings.
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		d := s.polyDist(star, p.Plan())
		return p.Y < deck+0.6 && d > 0.6 && p.Y > statueOfLibertyGround-0.2
	}})
	// Inner floor of the casemates.
	g.Use(Charcoal).Fill(Func{box, func(p Vec3) bool {
		return p.Y >= statueOfLibertyGround-0.6 && p.Y <= statueOfLibertyGround-0.2 && s.polyDist(star, p.Plan()) > 0
	}})
}

// foundation is the battered concrete base rising from the fort terrace.
func (s StatueOfLiberty) foundation(g *Grid) {
	y0, y1 := statueOfLibertyFortTop-0.5, statueOfLibertyFound
	half := func(y float64) float64 {
		h := 11.2 - 0.2*(y-y0)
		if y > 10.4 {
			h -= 0.5
		}
		if y > 13.2 {
			h -= 0.4
		}
		return h
	}
	box := Box{V(-11.5, y0, -11.5), V(11.5, y1, 11.5)}
	g.Use(Stone).Fill(Func{box, func(p Vec3) bool {
		h := half(p.Y)
		return math.Abs(p.X) <= h && math.Abs(p.Z) <= h
	}})
	g.Carve(Func{box, func(p Vec3) bool {
		h := half(p.Y) - 1.4
		return p.Y > y0+0.5 && p.Y < y1-1.2 && math.Abs(p.X) <= h && math.Abs(p.Z) <= h
	}})
	// Joint lines on the step edges.
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		m := math.Mod(p.Y-y0, 2.8)
		return m < 0.4 && p.Y > y0+1
	}})
	// Entrance doors on the north face.
	g.Carve(Arch{Sill: V(0, statueOfLibertyFortTop, -10.5), HalfWidth: 1.4, Spring: 2.0, Depth: 3, Through: AxisZ})
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		h := half(p.Y) - 0.6
		return p.Y > y0+0.5 && p.Y < y1-0.6 && math.Abs(p.X) <= h && math.Abs(p.Z) <= h
	}})
}

// pedestal is Hunt's granite pedestal: plinth, rusticated dado with corner quoins, the
// shield band, loggias with columns and balustrades, the cornice and the top balcony.
func (s StatueOfLiberty) pedestal(g *Grid) {
	type tier struct{ half, y0, y1 float64 }
	tiers := []tier{
		{9.4, statueOfLibertyFound, 16.2},
		{9.0, 16.2, 17.0},
		{8.4, 17.0, 25.0},
		{9.0, 25.0, 25.8},
		{8.2, 25.8, 28.0},
		{8.6, 28.0, 28.6},
		{8.0, 28.6, 34.6},
		{8.5, 34.6, 35.1},
		{9.0, 35.1, 35.9},
		{7.8, 35.9, 38.2},
		{8.4, 38.2, 39.0},
	}
	g.Use(Sand)
	for _, t := range tiers {
		g.Fill(BlockAt(P(0, 0), 2*t.half, 2*t.half, t.y0, t.y1))
	}
	// Corner quoins on the dado and loggia storeys.
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			g.Fill(Block(V(sx*8.8, 17.0, sz*8.8), V(sx*6.4, 25.0, sz*6.4)))
			g.Fill(Block(V(sx*8.4, 28.6, sz*8.4), V(sx*6.6, 34.6, sz*6.6)))
		}
	}
	// Rustication: alternate courses in grey granite.
	g.Paint(Stone, Func{Box{V(-10, 16, -10), V(10, 25, 10)}, func(p Vec3) bool {
		return math.Mod(p.Y-17.0, 1.6) < 0.5
	}})
	// Cornices and string courses in grey.
	for _, t := range tiers {
		if t.half >= 8.5 {
			g.Paint(Stone, BlockAt(P(0, 0), 2*t.half+0.5, 2*t.half+0.5, t.y0, t.y1))
		}
	}
	// Hollow interior with dark inner faces.
	g.Carve(BlockAt(P(0, 0), 13.0, 13.0, statueOfLibertyFound+0.6, 38.0))
	// Each face: a loggia opening into the dark interior, with three columns over a
	// balustrade, and a band of shields below.
	faces := func(f func(face func(Solid) Solid)) {
		for k := 0; k < 4; k++ {
			ang := float64(k) * math.Pi / 2
			f(func(sol Solid) Solid { return RotateY(sol, P(0, 0), ang) })
		}
	}
	faces(func(face func(Solid) Solid) { g.Carve(face(Block(V(-4.4, 29.4, 6.3), V(4.4, 33.6, 9)))) })
	g.Paint(Charcoal, Func{Box{V(-7.4, statueOfLibertyFound, -7.4), V(7.4, 38.6, 7.4)}, func(p Vec3) bool {
		return math.Max(math.Abs(p.X), math.Abs(p.Z)) <= 7.2 && p.Y > statueOfLibertyFound+0.2 && p.Y < 38.4
	}})
	faces(func(face func(Solid) Solid) {
		g.With(Sand, func() {
			for _, x := range []float64{-2.2, 0, 2.2} {
				g.Fill(face(Cylinder{Center: P(x, 7.7), Radius: 0.55, Y0: 29.4, Y1: 33.6}))
			}
			g.Fill(face(Block(V(-4.4, 33.2, 7.3), V(4.4, 33.8, 8.0))))
		})
		g.With(Stone, func() {
			g.Fill(face(Block(V(-5.0, 28.6, 8.0), V(5.0, 29.4, 8.7)))) // Balcony slab.
			g.Fill(face(Block(V(-5.0, 29.4, 8.3), V(5.0, 30.2, 8.7)))) // Balustrade.
			for _, x := range []float64{-6.0, -2.0, 2.0, 6.0} {        // Shields.
				g.FillVolume(face(Cylinder{Center: P(x, 8.2), Radius: 0.6, Y0: 26.2, Y1: 27.6}))
			}
		})
	})
	// Top balcony parapet with openings, around the statue's feet.
	g.Use(Stone).Fill(Func{Box{V(-8.5, statueOfLibertyPedTop, -8.5), V(8.5, statueOfLibertyPedTop+1.1, 8.5)}, func(p Vec3) bool {
		m := math.Max(math.Abs(p.X), math.Abs(p.Z))
		along := math.Min(math.Abs(p.X), math.Abs(p.Z))
		return m >= 7.6 && (math.Mod(along+0.6, 2.4) < 1.2 || p.Y > statueOfLibertyPedTop+0.6)
	}})
	// Small plinth under the statue.
	g.Use(Sand).Fill(BlockAt(P(0, 0), 11.5, 10, statueOfLibertyPedTop-0.4, statueOfLibertyPedTop+0.5))
}

// statue builds the copper figure with its arms, tablet, torch and crown.
func (s StatueOfLiberty) statue(g *Grid) {
	k := statueOfLibertyK
	box := Box{s.at(-8, -0.2, -6), s.at(9, 31, 7)}
	g.Use(Glass).FillVolume(Func{box, func(p Vec3) bool { return s.figure(s.local(p)) }})
	// Feet: the left foot steps forward out of the hem.
	g.FillVolume(Capsule{s.at(1.7, 0.5, 3.3), s.at(1.9, 0.5, 5.3), 0.75 * k})
	// Raised right arm: sleeve bunched at the shoulder, forearm and fist.
	shoulder, elbow, hand := s.at(-4.0, 21.6, 0), s.at(-5.0, 27.0, 0.3), s.at(-5.2, 32.0, 0.6)
	g.FillVolume(Capsule{shoulder, elbow, 1.55 * k}, Capsule{elbow, hand, 1.1 * k}, Sphere(hand, 1.3*k))
	g.FillVolume(Ellipsoid{s.at(-4.2, 22.4, 0), V(1.9, 1.6, 2.0).Scale(k)})
	// Left arm bent at the elbow, forearm forward under the tablet.
	lShoulder, lElbow, lHand := s.at(4.4, 21.0, -0.2), s.at(5.6, 16.0, -0.6), s.at(6.1, 15.8, 2.4)
	g.FillVolume(Capsule{lShoulder, lElbow, 1.3 * k}, Capsule{lElbow, lHand, 1.05 * k}, Sphere(lHand, 1.0*k))
	// Tablet with its inscription band.
	tab := Func{Box{s.at(3, 15, -2), s.at(9, 24, 5)}, func(p Vec3) bool {
		in, _, _, _ := s.tablet(s.local(p))
		return in
	}}
	g.FillVolume(tab)
	g.Paint(Water, Func{tab.Box, func(p Vec3) bool {
		in, u, v, w := s.tablet(s.local(p))
		return in && u > 0.1 && math.Abs(v) < 1.4 && math.Abs(w) < 2.6 && math.Mod(w+10, 1.3) < 0.55
	}})
	// Drapery grooves in a deeper patina.
	g.Paint(statueOfLibertyGroove, Func{box, func(p Vec3) bool {
		l := s.local(p)
		if l.Y > 19.5 || l.Y < 0.9 {
			return false
		}
		q, f, over := s.robeQ(l)
		if over && l.Y < s.hem(l.X)+0.5 {
			return true // Shadowed edge of the overgarment.
		}
		return f < -0.6 && q > 0.75
	}})
	// Face: dark eyes and the nose.
	g.Paint(Water, Sphere(s.at(-0.75, 27.9, 2.1), 0.32*k), Sphere(s.at(0.75, 27.9, 2.1), 0.32*k))
	g.Fill(Capsule{s.at(0, 28.0, 2.2), s.at(0, 27.1, 2.65), 0.42 * k})
	// Crown: diadem band with its windows and seven rays.
	g.Fill(Func{Box{s.at(-2.9, 28.9, -2.5), s.at(2.9, 30.3, 3.0)}, func(p Vec3) bool {
		l := s.local(p)
		return math.Hypot(l.X, l.Z-0.15) <= 2.75 && l.Z > -1.6
	}})
	g.Paint(Water, Func{Box{s.at(-2.9, 29.25, 0), s.at(2.9, 29.75, 3.0)}, func(p Vec3) bool {
		l := s.local(p)
		a := math.Atan2(l.X, l.Z)
		return math.Hypot(l.X, l.Z-0.15) > 2.1 && math.Abs(a) < 1.1 && math.Mod(a+4*math.Pi+0.1, 0.36) < 0.18
	}})
	for i := 0; i < 7; i++ {
		az := (float64(i) - 3) * 27 * math.Pi / 180
		el := (36 - 2*math.Abs(float64(i)-3)) * math.Pi / 180
		dir := V(math.Sin(az)*math.Cos(el), math.Sin(el), math.Cos(az)*math.Cos(el))
		base := s.at(2.2*math.Sin(az), 29.9, 0.15+2.2*math.Cos(az))
		mid := base.Add(dir.Scale(1.8 * k))
		tip := base.Add(dir.Scale(4.0 * k))
		g.Fill(Capsule{base, mid, 0.62 * k}, Capsule{mid, tip, 0.4 * k})
	}
	// Torch: handle, cup, balcony ring and gold flame.
	torch, y := hand.Plan(), hand.Y
	g.Fill(Cylinder{Center: torch, Radius: 0.75 * k, Y0: y, Y1: y + 2.9*k})
	g.FillVolume(Frustum{torch, 0.9 * k, 1.9 * k, y + 2.7*k, y + 4.2*k})
	g.Fill(Cylinder{Center: torch, Radius: 2.5 * k, Inner: 1.6 * k, Y0: y + 4.0*k, Y1: y + 4.6*k})
	g.Use(Gold).FillVolume(Ellipsoid{torch.At(y + 5.9*k), V(1.45, 1.9, 1.45).Scale(k)})
	g.FillVolume(Frustum{torch, 1.1 * k, 0, y + 6.3*k, y + 9.0*k})
}

// grounds adds the paths, trees, museum, gift building and flagpole.
func (s StatueOfLiberty) grounds(g *Grid) {
	star := s.star()
	gy := statueOfLibertyGround
	dockA, dockB := P(23, -7), P(39, -12)
	paths := []([2]Vec2){
		{dockA, dockB},
		{P(-21, 3), P(-36, 10)},
		{P(-19, -9), P(-33, -17)},
		{P(0, -20), P(2, -31)},
		{P(16, 15), P(24, 24)},
	}
	onPath := func(p Vec2) bool {
		d := s.polyDist(star, p)
		if d < -0.6 && d > -3.4 {
			return true
		}
		for _, sg := range paths {
			if s.segDist(p, sg[0], sg[1]) < 1.4 {
				return true
			}
		}
		return false
	}
	g.Paint(Sand, Func{Box{V(-44, gy-0.5, -38), V(46, gy, 34)}, func(p Vec3) bool {
		return s.islandEdge(p.Plan()) >= 3.2 && onPath(p.Plan())
	}})
	// Museum: low building with a planted roof and a glazed front.
	museum := P(27, 13)
	g.Use(Stone).Fill(BlockAt(museum, 12, 7, gy, gy+3.2))
	g.Paint(Glass, BlockAt(museum.Add(P(0, 3.2)), 10, 0.8, gy+0.4, gy+2.6))
	g.Paint(Glass, BlockAt(museum.Add(P(-5.7, 0)), 0.8, 5, gy+0.4, gy+2.6))
	g.Use(Grass).Fill(Func{Box{V(museum.X-6, gy+3.2, museum.Y-3.5), V(museum.X+6, gy+5, museum.Y+3.5)}, func(p Vec3) bool {
		return p.Y <= gy+3.2+(museum.Y+3.5-p.Z)*0.25+0.2
	}})
	g.Paint(Foliage, Func{Box{V(museum.X-6, gy+3.2, museum.Y-3.5), V(museum.X+6, gy+5, museum.Y+3.5)}, func(p Vec3) bool {
		return math.Sin(p.X*0.9)+math.Sin(p.Z*1.3+p.X*0.4) > 1.1
	}})
	// Ticket and gift building by the dock, slate hipped roof.
	shop := P(31, -21)
	g.Use(White).Fill(BlockAt(shop, 7, 4.5, gy, gy+2.8))
	g.Paint(Charcoal, BlockAt(shop, 7.2, 1.0, gy, gy+2.0))
	g.Use(Slate).Fill(HipRoof{P(shop.X-3.9, shop.Y-2.6), P(shop.X+3.9, shop.Y+2.6), gy + 2.8})
	// Flagpole with the flag streaming east.
	pole := P(22, -14)
	g.Use(Stone).Fill(Cylinder{Center: pole, Radius: 1.1, Y0: gy, Y1: gy + 0.6})
	g.Use(White).Fill(Cylinder{Center: pole, Radius: 0.5, Y0: gy, Y1: gy + 14})
	g.Use(Gold).Fill(Sphere(pole.At(gy+14.3), 0.5))
	flag := Block(V(pole.X+0.5, gy+10.4, pole.Y-0.4), V(pole.X+5.5, gy+13.6, pole.Y+0.4))
	g.Use(Red).Fill(flag)
	g.Paint(White, Func{flag.Bounds(), func(p Vec3) bool { return math.Mod(p.Y-gy-10.4, 1.0) < 0.5 }})
	g.Paint(Navy, Block(V(pole.X+0.5, gy+12.0, pole.Y-0.5), V(pole.X+2.7, gy+13.6, pole.Y+0.5)))
	// Lamp posts along the dock path.
	for t := 0.15; t < 1; t += 0.23 {
		c := dockA.Add(dockB.Sub(dockA).Scale(t)).Add(P(0.7, 2.0))
		g.Use(Charcoal).Fill(Cylinder{Center: c, Radius: 0.4, Y0: gy, Y1: gy + 3})
		g.Use(White).Fill(Sphere(c.At(gy+3.2), 0.5))
	}
	// Tree groves on the lawns.
	for i := -11; i <= 11; i++ {
		for j := -9; j <= 9; j++ {
			seed := i*101 + j*7
			c := P(float64(i)*3.9+(Hash01(seed, 1)-0.5)*2.2, float64(j)*3.9+(Hash01(seed, 2)-0.5)*2.2)
			if s.islandEdge(c) < 4.2 || s.polyDist(star, c) > -4.8 || Hash01(seed, 3) < 0.3 {
				continue
			}
			near := false
			for _, sg := range paths {
				if s.segDist(c, sg[0], sg[1]) < 3.2 {
					near = true
				}
			}
			for _, b := range []Vec2{museum, shop, pole} {
				if c.Sub(b).Length() < 9 {
					near = true
				}
			}
			if near || (c.X > 30 && c.Y < -4) {
				continue
			}
			h := 4 + Hash01(seed, 4)*2.2
			g.Use(Wood).Fill(Cylinder{Center: c, Radius: 0.5, Y0: gy, Y1: gy + h*0.55})
			leaf := Foliage
			if Hash01(seed, 5) < 0.3 {
				leaf = Grass
			}
			g.Use(leaf).FillVolume(Ellipsoid{c.At(gy + h*0.72), V(h*0.36, h*0.3, h*0.36)})
		}
	}
}

// ferry is the pier and a two-deck ferry moored alongside it.
func (s StatueOfLiberty) ferry(g *Grid) {
	gy := statueOfLibertyGround
	// Pier on wooden piles.
	g.Use(Wood).Fill(Block(V(37, gy-0.5, -14), V(50, gy, -10)))
	for x := 38.0; x <= 50; x += 3 {
		for _, z := range []float64{-13.6, -10.4} {
			g.Fill(Cylinder{Center: P(x, z), Radius: 0.5, Y0: 0, Y1: gy})
		}
	}
	g.Use(Charcoal).Fill(Block(V(38, gy, -14), V(50, gy+0.8, -13.6)))
	// Hull: pointed bow to +X, square stern.
	const x0, x1, half = 34.0, 52.0, 2.9
	hullHalf := func(x float64) float64 {
		if x > x1-5 {
			t := (x - (x1 - 5)) / 5
			return half * math.Sqrt(math.Max(0, 1-t*t))
		}
		return half
	}
	cz := -6.4
	hull := Func{Box{V(x0, 0.2, cz-half), V(x1, 2.4, cz+half)}, func(p Vec3) bool {
		return math.Abs(p.Z-cz) <= hullHalf(p.X)
	}}
	g.Use(White).FillVolume(hull)
	g.Paint(Navy, Func{hull.Box, func(p Vec3) bool { return p.Y < 1.2 }})
	g.Paint(Red, Func{hull.Box, func(p Vec3) bool { return p.Y >= 1.2 && p.Y < 1.6 }})
	// Main and upper deck cabins with window bands, wheelhouse and open top deck.
	g.Use(White).Fill(Block(V(x0+0.8, 2.4, cz-2.4), V(x1-4.5, 4.4, cz+2.4)))
	g.Paint(Charcoal, Block(V(x0+1.2, 3.0, cz-2.5), V(x1-5.0, 3.9, cz+2.5)))
	g.Fill(Block(V(x0+2.5, 4.4, cz-2.1), V(x1-7, 6.2, cz+2.1)))
	g.Paint(Glass, Block(V(x0+3, 5.0, cz-2.2), V(x1-7.5, 5.7, cz+2.2)))
	g.Use(Navy).Fill(Block(V(x1-9.5, 6.2, cz-1.5), V(x1-7, 7.4, cz+1.5)))
	g.Use(Charcoal).Fill(Func{Box{V(x0+2.5, 6.2, cz-2.1), V(x1-7, 7.0, cz+2.1)}, func(p Vec3) bool {
		return math.Abs(p.Z-cz) > 1.6 || p.X < x0+3
	}})
}
