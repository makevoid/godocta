package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// AircraftCarrier is a generic nuclear supercarrier in the style of the Nimitz class, at
// sea: a long hollow hull with a flared bow and a transom stern, red antifouling and a
// charcoal boot-top; an overhanging charcoal flight deck carried on sloped sponsons, with
// the 9° angled landing deck, white and gold markings, arresting wires, four catapult
// tracks with blast deflectors and the fictional hull number 92; four deck-edge elevators
// (one lowered to the hangar with a jet on it); the island on the starboard side with
// bridge windows, Pri-Fly, a mast with planar and rotating radars and radomes; parked
// strike fighters with folded wings, an airborne early-warning plane, helicopters, deck
// tractors and a crane; and weapon sponsons with CIWS mounts and missile launchers.
// y = 0 is the flat bottom; the waterline is 3.5 above it. Bow points +X; starboard is +Z.
type AircraftCarrier struct{}

func init() { Register(AircraftCarrier{}) }

func (AircraftCarrier) Name() string  { return "aircraft_carrier" }
func (AircraftCarrier) Title() string { return "Aircraft Carrier" }

const (
	aircraftCarrierBoot  = 3.4  // Top of the red antifouling.
	aircraftCarrierHD    = 8.5  // Hangar deck top.
	aircraftCarrierFD    = 15.0 // Flight deck top.
	aircraftCarrierSkin  = 1.2  // Shell thickness.
	aircraftCarrierAngle = 0.157
)

// aircraftCarrierDeck is the flight deck outline in plan (X forward, Y = world Z, starboard +).
var aircraftCarrierDeck = []Vec2{
	{92, 6.5}, {87, 10}, {78, 12.6}, {62, 14.6}, {44, 15.5},
	{-4, 15.5}, {-4, 18.5}, {-23, 18.5}, {-23, 15.5}, {-88, 15.5}, {-92, 14},
	{-92, -13}, {-89, -15}, {-12, -27.2}, {2, -27.5}, {8, -26.8},
	{40, -14.5}, {58, -13.2}, {76, -11.6}, {86, -9.5}, {92, -6.5},
}

// aircraftCarrierElevator is a deck-edge elevator platform; lowered ones sit at the
// hangar deck.
type aircraftCarrierElevator struct {
	x0, x1, z0, z1 float64
	lowered        bool
}

var aircraftCarrierElevators = []aircraftCarrierElevator{
	{16, 28, 13.5, 21, false},
	{-2.5, 9.5, 13.5, 21, false},
	{-40, -28, 13.5, 21, true},
	{-60, -48, -27.5, -19.5, false},
}

// hullW is the hull's outer half-beam at (x, y), negative outside its length.
func (AircraftCarrier) hullW(x, y float64) float64 {
	stem := 80 + 10*math.Pow(math.Min(math.Max(y/15, 0), 1), 1.2)
	stern := -88.0
	if y < 4 {
		stern += (4 - y) * 1.6
	}
	if x > stem || x < stern {
		return -1
	}
	w := 11.4 + 1.6*math.Min(math.Max((y-3.5)/10.5, 0), 1)
	if x > 20 { // Long entry, fuller (flared) higher up.
		t := (x - 20) / (stem - 20)
		k := 1.6 + 1.6*math.Min(math.Max(y/14, 0), 1)
		w *= 1 - math.Pow(t, k)
	}
	if x < -58 { // Run narrowing low down into the transom.
		t := (-58 - x) / 30
		low := math.Min(math.Max((9-y)/9, 0), 1)
		w *= 1 - 0.4*t*t*low
	}
	if y < 2.5 { // Rounded bilge.
		t := (2.5 - y) / 2.5
		w *= 1 - 0.3*t*t
	}
	return w
}

// innerW is the hollow hull's interior half-width at the given inset.
func (s AircraftCarrier) innerW(x, y, inset float64) float64 {
	a, b := s.hullW(x-inset, y), s.hullW(x+inset, y)
	if a < 0 || b < 0 {
		return -1
	}
	return math.Min(a, b) - inset
}

func (AircraftCarrier) segDist(p, a, b Vec2) float64 {
	ab := b.Sub(a)
	t := math.Min(math.Max(p.Sub(a).Dot(ab)/ab.Dot(ab), 0), 1)
	return p.Sub(a.Add(ab.Scale(t))).Length()
}

// edgeDist is the plan distance to the flight deck outline.
func (s AircraftCarrier) edgeDist(p Vec2) float64 {
	d := math.Inf(1)
	poly := aircraftCarrierDeck
	for i, j := 0, len(poly)-1; i < len(poly); j, i = i, i+1 {
		d = math.Min(d, s.segDist(p, poly[j], poly[i]))
	}
	return d
}

// inElevator reports whether a plan point lies in an elevator's footprint grown by grow.
func (AircraftCarrier) inElevator(p Vec2, grow float64) bool {
	for _, e := range aircraftCarrierElevators {
		if p.X >= e.x0-grow && p.X <= e.x1+grow && p.Y >= e.z0-grow && p.Y <= e.z1+grow {
			return true
		}
	}
	return false
}

// underY is the underside height of the overhang sponsons: a skirt at the deck edge,
// then sloping down toward the hull.
func (s AircraftCarrier) underY(p Vec2) float64 {
	return math.Max(aircraftCarrierFD-2.6-0.6*s.edgeDist(p), 7.5)
}

// landing returns the (along, across) coordinates in the angled landing area's frame:
// along runs forward from the stern, across is positive to starboard.
func (AircraftCarrier) landing(p Vec2) (float64, float64) {
	n := math.Hypot(1, aircraftCarrierAngle)
	d := P(1/n, -aircraftCarrierAngle/n)
	q := p.Sub(P(-92, 0.5))
	return q.Dot(d), q.Dot(P(-d.Y, d.X))
}

// place turns a solid authored facing +X at the origin to the angle, then moves it to (x, z).
func (AircraftCarrier) place(sol Solid, x, z, angle float64) Solid {
	return Translate(RotateY(sol, P(0, 0), angle), V(x, 0, z))
}

// line is a plan segment painted or filled between y0 and y1.
func (s AircraftCarrier) line(a, b Vec2, half, y0, y1 float64) Solid {
	box := BoxOf(a.At(y0), b.At(y1)).Grow(half)
	box.Min.Y, box.Max.Y = y0, y1
	return Func{box, func(p Vec3) bool { return s.segDist(p.Plan(), a, b) <= half }}
}

// glyphLit reports whether (u, v) lies on a stroke of text set in 3×5 glyphs of size px;
// u runs left to right and v down from the top.
func (AircraftCarrier) glyphLit(text string, px, u, v float64) bool {
	glyphs := map[byte][5]string{
		'9': {"###", "#.#", "###", "..#", "###"},
		'2': {"###", "..#", "###", "#..", "###"},
	}
	if u < 0 || v < 0 {
		return false
	}
	row, col := int(v/px), int(u/px)
	if row > 4 {
		return false
	}
	ch, c := col/4, col%4
	if ch >= len(text) || c == 3 {
		return false
	}
	return glyphs[text[ch]][row][c] == '#'
}

func (s AircraftCarrier) Build(g *Grid) Order {
	s.buildHull(g)
	s.buildDeck(g)
	s.buildMarkings(g)
	s.buildIsland(g)
	s.buildSponsons(g)
	s.buildAircraft(g)
	return Along(P(1, 0))
}

// buildHull lays the hollow hull and overhang sponsons, livery, hangar openings and anchors.
func (s AircraftCarrier) buildHull(g *Grid) {
	const fd, skin = aircraftCarrierFD, aircraftCarrierSkin
	hullBox := Box{V(-89, 0, -14), V(91, fd-1, 14)}
	g.Use(Slate).FillVolume(Func{hullBox, func(p Vec3) bool {
		return math.Abs(p.Z) <= s.hullW(p.X, p.Y)
	}})
	deckBox := Box{V(-92, 7.5, -27.5), V(92, fd-1, 18.5)}
	inDeck := func(p Vec2) bool { return InsidePolygon(aircraftCarrierDeck, p) && !s.inElevator(p, 0) }
	g.Fill(Func{deckBox, func(p Vec3) bool {
		return inDeck(p.Plan()) && p.Y >= s.underY(p.Plan())
	}})
	// Hollow everything out, leaving a skin.
	interior := func(inset float64) Solid {
		return Func{Box{V(-92, 0, -28), V(92, fd-1, 19)}, func(p Vec3) bool {
			if p.Y < inset || p.Y > fd-1-inset/2 {
				return false
			}
			if math.Abs(p.Z) < s.innerW(p.X, p.Y, inset) {
				return true
			}
			q := p.Plan()
			return inDeck(q) && !s.inElevator(q, inset) && s.edgeDist(q) > inset && p.Y > s.underY(q)+inset
		}}
	}
	g.Carve(interior(skin))
	// Livery: red antifouling and a charcoal boot-top at the waterline.
	g.Paint(Red, Block(V(-92, 0, -28), V(92, aircraftCarrierBoot, 28)))
	g.Paint(Charcoal, Block(V(-92, aircraftCarrierBoot+0.1, -28), V(92, aircraftCarrierBoot+0.9, 28)))
	g.Paint(Charcoal, interior(skin-0.45))

	// Hangar openings behind the starboard elevators, with a dark hangar floor inside.
	for _, e := range aircraftCarrierElevators {
		if e.z0 < 0 {
			continue
		}
		g.Carve(Block(V(e.x0+0.8, aircraftCarrierHD, 9), V(e.x1-0.8, fd-1.8, 15)))
		g.Use(Charcoal).Fill(Block(V(e.x0-1, aircraftCarrierHD, 2), V(e.x1+1, fd-1, 3)))
		g.Use(Charcoal).Fill(Block(V(e.x0-3, aircraftCarrierHD-0.8, -2), V(e.x1+3, aircraftCarrierHD, s.innerW((e.x0+e.x1)/2, aircraftCarrierHD, skin))))
	}
	// Fantail opening under the stern overhang.
	g.Carve(Block(V(-89.5, 9.5, -8), V(-86, 13, 8)))
	g.Use(DarkStone).Fill(Block(V(-88, 8.7, -8), V(-80, 9.5, 8)))

	// Portholes in two long rows, skipping the hangar openings.
	for x := -84.0; x <= 76; x += 3.2 {
		for _, y := range []float64{6.4, 10.4} {
			hole := Box{V(x-0.4, y-0.4, -15), V(x+0.4, y+0.4, 15)}
			g.Paint(Charcoal, Func{hole, func(p Vec3) bool {
				if p.Z > 0 && y > 8 && s.inElevator(P(p.X, 17), 1.2) {
					return false
				}
				return math.Abs(p.Z) > s.hullW(p.X, p.Y)-0.7
			}})
		}
	}
	// Recessed side ports: refuelling stations and boat bays.
	for _, port := range []struct{ x, side float64 }{{42, 1}, {-50, 1}, {-72, 1}, {32, -1}, {-30, -1}, {-66, -1}} {
		box := Box{V(port.x-1.8, 9.4, -15), V(port.x+1.8, 11.6, 15)}
		g.Carve(Func{box, func(p Vec3) bool { return p.Z*port.side > s.hullW(p.X, p.Y)-0.7 }})
		g.Paint(Charcoal, Func{box, func(p Vec3) bool { return p.Z*port.side > s.hullW(p.X, p.Y)-1.6 }})
	}
	// Catwalk groove under the overhanging deck edge.
	groove := func(d0, d1 float64) Solid {
		return Func{Box{V(-92, fd-2.6, -28), V(92, fd-1.8, 19)}, func(p Vec3) bool {
			q := p.Plan()
			if !InsidePolygon(aircraftCarrierDeck, q) || s.inElevator(q, 0.8) || math.Abs(p.Z) < s.hullW(p.X, p.Y)+0.4 {
				return false
			}
			e := s.edgeDist(q)
			return e >= d0 && e < d1
		}}
	}
	g.Carve(groove(0, 0.6))
	g.Paint(Charcoal, groove(0.6, 1.3))

	// Anchors in their hawse pockets on both bows.
	for _, side := range []float64{-1, 1} {
		ax, ay := 74.0, 10.5
		w := s.hullW(ax, ay)
		g.Carve(Func{Box{V(ax-1, ay-1, -15), V(ax+1, ay+1, 15)}, func(p Vec3) bool {
			return p.Z*side > s.hullW(p.X, p.Y)-0.7
		}})
		zc := side * (w + 0.1)
		g.Use(Charcoal).Fill(
			Block(V(ax-0.45, ay-3.0, zc-0.45), V(ax+0.45, ay+0.6, zc+0.45)),
			Block(V(ax-1.5, ay-3.6, zc-0.5), V(ax+1.5, ay-2.8, zc+0.5)),
			Block(V(ax-1.5, ay-3.6, zc-0.5), V(ax-0.9, ay-2.2, zc+0.5)),
			Block(V(ax+0.9, ay-3.6, zc-0.5), V(ax+1.5, ay-2.2, zc+0.5)),
		)
	}
}

// buildDeck lays the flight deck, elevators, catapult blast deflectors and deck-edge gear.
func (s AircraftCarrier) buildDeck(g *Grid) {
	const fd = aircraftCarrierFD
	g.Use(Charcoal).Fill(Func{Box{V(-92, fd-1, -27.5), V(92, fd, 18.5)}, func(p Vec3) bool {
		q := p.Plan()
		return InsidePolygon(aircraftCarrierDeck, q) && !s.inElevator(q, 0)
	}})
	for _, e := range aircraftCarrierElevators {
		top := fd
		if e.lowered {
			top = aircraftCarrierHD
		}
		g.Use(Charcoal).Fill(Block(V(e.x0+0.2, top-0.9, e.z0+0.2), V(e.x1-0.2, top, e.z1-0.2)))
		// Hazard edging on the outboard edges.
		g.Paint(Gold, Func{Box{V(e.x0, top-0.4, e.z0), V(e.x1, top, e.z1)}, func(p Vec3) bool {
			return p.X < e.x0+0.8 || p.X > e.x1-0.8 || p.Z < e.z0+0.8 || p.Z > e.z1-0.8
		}})
		if e.lowered { // Guide rails up the hull side.
			for _, x := range []float64{e.x0 - 0.3, e.x1 + 0.3} {
				g.Use(DarkStone).Fill(Block(V(x-0.5, aircraftCarrierHD-1.5, e.z0-0.4), V(x+0.5, fd-1, e.z0+0.8)))
			}
		}
	}
}

// buildMarkings paints the landing area, catapult tracks, wires and the deck number.
func (s AircraftCarrier) buildMarkings(g *Grid) {
	const fd = aircraftCarrierFD
	top := func(f func(q Vec2) bool) Func {
		return Func{Box{V(-92, fd-0.45, -28), V(92, fd, 19)}, func(p Vec3) bool {
			q := p.Plan()
			return InsidePolygon(aircraftCarrierDeck, q) && !s.inElevator(q, 0) && f(q)
		}}
	}
	// Angled landing area: white edges, dashed centerline, gold foul line, stern ramp.
	g.Paint(White, top(func(q Vec2) bool {
		along, across := s.landing(q)
		if along < 0 {
			return false
		}
		a := math.Abs(across)
		return (a > 5.6 && a < 6.4) || (a < 0.35 && math.Mod(along, 6) < 3.4)
	}))
	g.Paint(Gold, top(func(q Vec2) bool {
		along, across := s.landing(q)
		return along > 2 && across > 7.8 && across < 8.5
	}))
	g.Paint(White, top(func(q Vec2) bool {
		return q.X < -90.8 && math.Mod(q.Y+40, 2.4) < 1.2
	}))
	// Arresting wires across the landing area.
	g.Paint(Stone, top(func(q Vec2) bool {
		along, across := s.landing(q)
		if math.Abs(across) > 5.4 {
			return false
		}
		for _, w := range []float64{26, 31, 36, 41} {
			if math.Abs(along-w) < 0.4 {
				return true
			}
		}
		return false
	}))
	// Catapult tracks: two on the bow, two in the waist.
	type track struct{ a, b Vec2 }
	cats := []track{{P(42, 3.0), P(91, 5.0)}, {P(42, -3.0), P(91, -5.2)}, {P(-34, -14.2), P(7, -21.4)}, {P(-28, -19.6), P(9, -25.9)}}
	for _, c := range cats {
		g.Paint(Stone, s.line(c.a, c.b, 0.42, fd-0.45, fd))
	}
	// Blast deflectors behind each catapult: raised on cat 1, flush outlines on the rest.
	for i, c := range cats {
		d := c.b.Sub(c.a).Scale(1 / c.b.Sub(c.a).Length())
		n := P(-d.Y, d.X)
		at := c.a.Sub(d.Scale(2.5))
		if i == 0 {
			g.Use(Slate).Fill(Sheared{s.line(at.Add(n.Scale(3)), at.Sub(n.Scale(3)), 0.45, fd, fd+2.2), d.Scale(-0.45), fd})
			continue
		}
		g.Paint(Gold, s.line(at.Add(n.Scale(3)), at.Sub(n.Scale(3)), 0.45, fd-0.45, fd))
	}
	// Taxi line along the starboard side.
	g.Paint(Gold, top(func(q Vec2) bool { return q.X > -86 && q.X < 86 && math.Abs(q.Y-7) < 0.35 && math.Mod(q.X+200, 5) < 3 }))
	// Deck number on the bow, read facing forward.
	const text, px, topX = "92", 1.6, 89.0
	width := float64(len(text))*4*px - px
	g.Paint(White, top(func(q Vec2) bool { return s.glyphLit(text, px, q.Y+width/2, topX-q.X) }))
}

// buildIsland raises the island on the starboard deck edge with its bridges and masts.
func (s AircraftCarrier) buildIsland(g *Grid) {
	const fd = aircraftCarrierFD
	x0, x1, z0, z1 := -21.0, -6.0, 11.5, 17.5
	plan := func(in float64, fwd float64) []Vec2 {
		return []Vec2{{x0 + in, z0 + in}, {x1 + fwd - 2.5 - in*0.4, z0 + in}, {x1 + fwd - in, z0 + 2.2}, {x1 + fwd - in, z1 - in}, {x0 + in, z1 - in}}
	}
	l1, l2, l3, l4, l5 := fd+2.7, fd+5.4, fd+8.1, fd+10.8, fd+13.4
	g.Use(Slate).Fill(Prism{plan(0, 0), fd, l3})
	g.Fill(Prism{plan(0, 0.8), l3, l4}) // Navigation bridge, jutting forward.
	g.Fill(Prism{plan(0, 0.8), l4 - 0.1, l4 + 0.4})
	g.Carve(Prism{plan(1, 0), fd, l4 - 0.8})
	g.Paint(Charcoal, Prism{plan(0.55, 0), fd, l4 - 0.4})
	// Pri-Fly and the flag levels on top, set back from the front.
	pri := []Vec2{{x0 + 1, z0 + 0.4}, {x1 - 3.5, z0 + 0.4}, {x1 - 3.5, z1 - 1}, {x0 + 1, z1 - 1}}
	g.Fill(Prism{pri, l4, l5})
	g.Use(DarkStone).Fill(Prism{[]Vec2{{x0 + 0.6, z0}, {x1 - 3, z0}, {x1 - 3, z1 - 0.6}, {x0 + 0.6, z1 - 0.6}}, l5, l5 + 0.5})
	// Level ledges and window bands.
	for _, y := range []float64{l1, l2, l3} {
		g.Use(DarkStone).Fill(Prism{plan(-0.3, 0.3), y - 0.3, y + 0.1})
	}
	g.Paint(Glass, Func{Box{V(x0, l3+0.7, z0), V(x1+0.8, l3+1.7, z1)}, func(p Vec3) bool {
		return p.X > x1-5 || p.Z < z0+0.8
	}})
	g.Paint(Glass, Func{Box{V(x0, l4+0.8, z0), V(x1, l4+1.8, z1)}, func(p Vec3) bool {
		return p.Z < z0+1.2 || p.X < x0+1.6 || p.X > x1-4.3
	}})
	// Hull number on both faces, read from outboard on each side.
	const text, px = "92", 1.3
	width := float64(len(text))*4*px - px
	nx := (x0+x1)/2 - width/2 - 0.5
	ny := fd + 7.8
	g.Paint(White, Func{Box{V(nx, ny-5*px, z0-1), V(nx+width, ny, z1+1)}, func(p Vec3) bool {
		if p.Z > (z0+z1)/2 {
			return p.Z > z1-0.5 && s.glyphLit(text, px, p.X-nx, ny-p.Y)
		}
		return p.Z < z0+0.5 && s.glyphLit(text, px, nx+width-p.X, ny-p.Y)
	}})

	// Main mast: tripod, platform, planar air-search array, yard and topmast.
	mx, mz := -11.5, 14.5
	g.Use(Slate).Fill(
		Capsule{V(mx+1.5, l5, mz-2), V(mx, l5+5, mz), 0.5},
		Capsule{V(mx+1.5, l5, mz+2), V(mx, l5+5, mz), 0.5},
		Capsule{V(mx-2, l5, mz), V(mx, l5+5, mz), 0.5},
		Capsule{V(mx, l5+5, mz), V(mx, l5+13.5, mz), 0.45},
	)
	g.Use(DarkStone).Fill(Block(V(mx-2.2, l5+4, mz-2.4), V(mx+2.4, l5+4.6, mz+2.4)))
	array := Sheared{Block(V(mx+1.4, l5+4.6, mz-2.3), V(mx+2.1, l5+8.8, mz+2.3)), P(-0.25, 0), l5 + 4.6}
	g.Use(Charcoal).Fill(array)
	g.Paint(Slate, Func{array.Bounds(), func(p Vec3) bool {
		return array.Contains(p) && (math.Mod(p.Y+0.2, 1.3) < 0.45 || math.Mod(p.Z+50.2, 1.5) < 0.45)
	}})
	g.Use(Slate).Fill(Capsule{V(mx, l5+10, mz-4), V(mx, l5+10, mz+4), 0.42})
	g.Use(White).Fill(Sphere(V(mx, l5+14, mz), 0.8))
	// Aft lattice mast with the rotating air-search antenna.
	ax := -19.0
	for _, dz := range []float64{-1.2, 1.2} {
		g.Use(Slate).Fill(Capsule{V(ax-1, l5, mz+dz), V(ax, l5+7, mz+dz*0.4), 0.45})
	}
	g.Fill(Capsule{V(ax-1, l5+2, mz-1.2), V(ax, l5+4, mz+0.5), 0.42}, Capsule{V(ax-1, l5+2, mz+1.2), V(ax, l5+4, mz-0.5), 0.42})
	dish := Block(V(ax-0.6, l5+7, mz-3.5), V(ax+0.2, l5+9.6, mz+3.5))
	g.Use(Charcoal).Fill(dish)
	g.Carve(Func{dish.Bounds(), func(p Vec3) bool {
		return math.Abs(p.Z-mz) < 3 && p.Y > l5+7.5 && p.Y < l5+9.1 && math.Mod(p.Z+50.1, 1.4) < 0.6
	}})
	// Radomes on platforms off the island's sides and top.
	for _, r := range []struct{ x, z, y float64 }{{x0 - 0.6, z1 - 1.2, l3}, {x0 - 0.2, z0 + 1, l4}, {-15, z1 - 0.2, l5 + 0.5}} {
		g.Use(DarkStone).Fill(Cylinder{Center: P(r.x, r.z), Radius: 1.3, Y0: r.y - 0.5, Y1: r.y})
		g.Use(White).Fill(Sphere(V(r.x, r.y+1.1, r.z), 1.2))
	}
	// Whip antennas at the deck edge.
	for _, x := range []float64{-50, -66, 34, 48} {
		g.Use(Slate).Fill(Capsule{V(x, fd-2, 16), V(x, fd+4, 16.3), 0.42})
	}
}

// ciws is a close-in weapon system: pedestal, white radome and gun, facing angle.
func (s AircraftCarrier) ciws(g *Grid, x, z, y, angle float64) {
	g.Use(Stone).Fill(Cylinder{Center: P(x, z), Radius: 0.9, Y0: y, Y1: y + 0.9})
	g.Use(White).Fill(Capsule{V(x, y+1.4, z), V(x, y+2.3, z), 0.8})
	g.Use(Charcoal).Fill(s.place(Capsule{V(0.6, y+1.2, 0), V(2.8, y+1.4, 0), 0.42}, x, z, angle))
}

// launcher is a boxy missile launcher with a dark cell face, facing angle.
func (s AircraftCarrier) launcher(g *Grid, x, z, y, angle float64) {
	g.Use(Stone).Fill(Cylinder{Center: P(x, z), Radius: 0.9, Y0: y, Y1: y + 0.8})
	g.Fill(s.place(Block(V(-1.2, y+0.8, -1.1), V(1.2, y+2.4, 1.1)), x, z, angle))
	g.Paint(Charcoal, s.place(Block(V(0.9, y+1.0, -0.9), V(1.2, y+2.2, 0.9)), x, z, angle))
}

// buildSponsons hangs weapon sponsons off the deck edges, plus the landing aids.
func (s AircraftCarrier) buildSponsons(g *Grid) {
	const fd = aircraftCarrierFD
	type mount struct {
		x, z, angle float64
		gun         bool
	}
	for _, m := range []mount{{84, 11.5, 0.6, true}, {80, -11.5, -0.6, false}, {-90, 15.5, 2.6, true}, {-90, -14.5, -2.6, true}, {-70, -18.2, -1.9, false}, {50, 17.2, 1.2, false}} {
		y := fd - 3.4
		dir := P(math.Cos(m.angle), math.Sin(m.angle))
		c := P(m.x, m.z).Add(dir.Scale(1.2))
		g.Use(Slate).Fill(Cylinder{Center: c, Radius: 2.0, Y0: y - 0.8, Y1: y})
		g.Fill(Cylinder{Center: c, Radius: 2.0, Inner: 1.5, Y0: y, Y1: y + 0.6})
		if m.gun {
			s.ciws(g, c.X, c.Y, y, m.angle)
		} else {
			s.launcher(g, c.X, c.Y, y, m.angle)
		}
	}
	// Landing signal officer's platform and the optical landing aid on the port side.
	g.Use(Slate).Fill(Block(V(-73, fd-0.8, -20), V(-67, fd, -17)))
	g.Use(White).Fill(Block(V(-73, fd, -20), V(-72.4, fd+1.2, -17.5)))
	lx, lz := -40.0, -22.6
	g.Use(DarkStone).Fill(Block(V(lx-0.6, fd, lz-2.2), V(lx+0.6, fd+2.8, lz+2.2)))
	g.Paint(Grass, Block(V(lx+0.4, fd+1.6, lz-2.2), V(lx+0.6, fd+2.0, lz+2.2)))
	g.Paint(Gold, Block(V(lx+0.4, fd+1.4, lz-0.5), V(lx+0.6, fd+2.2, lz+0.5)))
}

// jet is a twin-tail strike fighter with folded wings, authored facing +X.
func (s AircraftCarrier) jet(g *Grid, x, z, y, angle float64, body, accent Color) {
	at := func(sol Solid) Solid { return s.place(sol, x, z, angle) }
	g.Use(body).Fill(
		at(Capsule{V(-4.2, y+1.2, 0), V(3.0, y+1.2, 0), 0.72}),
		at(Capsule{V(3.0, y+1.15, 0), V(5.0, y+1.0, 0), 0.45}),
		at(Block(V(-4.4, y+0.6, -1.1), V(1.4, y+1.5, 1.1))),
		at(Prism{[]Vec2{{-2.4, -2.4}, {0.2, -2.4}, {2.2, -0.9}, {2.2, 0.9}, {0.2, 2.4}, {-2.4, 2.4}}, y + 1.0, y + 1.6}),
		at(Block(V(-2.1, y+1.0, -2.9), V(0.0, y+3.0, -2.3))),
		at(Block(V(-2.1, y+1.0, 2.3), V(0.0, y+3.0, 2.9))),
		at(Block(V(-5.2, y+0.9, -2.3), V(-3.6, y+1.5, 2.3))),
		at(Block(V(-4.8, y+1.4, -1.2), V(-3.0, y+3.4, -0.6))),
		at(Block(V(-4.8, y+1.4, 0.6), V(-3.0, y+3.4, 1.2))),
		at(Block(V(2.6, y, -0.3), V(3.4, y+0.8, 0.3))),
		at(Block(V(-1.8, y, -1.3), V(-0.8, y+0.8, 1.3))),
	)
	g.Use(Glass).Fill(at(Block(V(1.2, y+1.7, -0.45), V(3.0, y+2.3, 0.45))))
	g.Use(Charcoal).Fill(at(Block(V(-4.9, y+0.7, -0.9), V(-4.3, y+1.5, 0.9))))
	if accent != body {
		g.Paint(accent, at(Block(V(-4.8, y+2.6, -1.3), V(-3.0, y+3.4, 1.3))))
	}
}

// hawkeye is an airborne early-warning plane with its rotodome and folded wings.
func (s AircraftCarrier) hawkeye(g *Grid, x, z, y, angle float64) {
	at := func(sol Solid) Solid { return s.place(sol, x, z, angle) }
	g.Use(Stone).Fill(
		at(Capsule{V(-4.6, y+1.6, 0), V(3.6, y+1.4, 0), 0.9}),
		at(Capsule{V(3.6, y+1.4, 0), V(5.0, y+1.2, 0), 0.6}),
		at(Block(V(-0.4, y+1.8, -3.0), V(1.4, y+2.4, 3.0))),
		at(Block(V(-5.0, y+2.0, -1.9), V(-0.4, y+2.6, -1.3))),
		at(Block(V(-5.0, y+2.0, 1.3), V(-0.4, y+2.6, 1.9))),
		at(Capsule{V(-1.0, y+1.6, -2.4), V(2.6, y+1.6, -2.4), 0.6}),
		at(Capsule{V(-1.0, y+1.6, 2.4), V(2.6, y+1.6, 2.4), 0.6}),
		at(Block(V(-6.4, y+1.8, -3.0), V(-5.2, y+2.3, 3.0))),
		at(Block(V(-1.0, y+2.4, -0.4), V(0.0, y+3.0, 0.4))),
		at(Block(V(2.8, y, -0.3), V(3.4, y+0.8, 0.3))),
		at(Block(V(-0.2, y, -2.6), V(0.6, y+1.2, 2.6))),
	)
	for _, fz := range []float64{-2.9, -1.0, 1.0, 2.9} {
		g.Fill(at(Block(V(-6.4, y+1.4, fz-0.3), V(-5.2, y+3.6, fz+0.3))))
	}
	g.Use(White).Fill(at(Cylinder{Center: P(-0.5, 0), Radius: 2.4, Y0: y + 3.0, Y1: y + 3.7}))
	g.Paint(Charcoal, at(Cylinder{Center: P(-0.5, 0), Radius: 2.4, Inner: 1.8, Y0: y + 3.0, Y1: y + 3.7}))
	g.Use(Charcoal).Fill(at(Block(V(2.8, y+0.6, -3.6), V(3.2, y+2.6, -1.2))), at(Block(V(2.8, y+0.6, 1.2), V(3.2, y+2.6, 3.6))))
	g.Use(Glass).Fill(at(Block(V(3.4, y+1.8, -0.5), V(4.4, y+2.3, 0.5))))
}

// helicopter is a utility helicopter with its rotor blades folded aft.
func (s AircraftCarrier) helicopter(g *Grid, x, z, y, angle float64) {
	at := func(sol Solid) Solid { return s.place(sol, x, z, angle) }
	g.Use(Slate).Fill(
		at(Ellipsoid{V(0.8, y+1.5, 0), V(2.8, 1.2, 1.15)}),
		at(Capsule{V(-1.2, y+1.8, 0), V(-6.0, y+2.2, 0), 0.5}),
		at(Block(V(-6.6, y+1.8, -0.3), V(-5.6, y+3.6, 0.3))),
		at(Cylinder{Center: P(0.8, 0), Radius: 0.6, Y0: y + 2.4, Y1: y + 3.2}),
		at(Block(V(2.0, y, -0.9), V(2.6, y+0.6, 0.9))),
	)
	g.Use(DarkStone).Fill(
		at(Capsule{V(0.8, y+3.1, 0), V(-6.0, y+3.0, -0.9), 0.42}),
		at(Capsule{V(0.8, y+3.1, 0), V(-6.0, y+3.0, 0.9), 0.42}),
	)
	g.Use(Glass).Fill(at(Block(V(2.6, y+1.6, -0.8), V(3.4, y+2.3, 0.8))))
}

// tractor is a small deck tractor.
func (s AircraftCarrier) tractor(g *Grid, x, z, y, angle float64) {
	at := func(sol Solid) Solid { return s.place(sol, x, z, angle) }
	g.Use(Gold).Fill(at(Block(V(-1.1, y, -0.8), V(1.1, y+0.9, 0.8))))
	g.Use(Charcoal).Fill(at(Block(V(-1.0, y+0.9, -0.5), V(-0.2, y+1.4, 0.5))))
}

// buildAircraft parks the air wing clear of the landing area.
func (s AircraftCarrier) buildAircraft(g *Grid) {
	const fd = aircraftCarrierFD
	type spot struct {
		x, z, angle float64
		accent      Color
	}
	nose := -1.05 // Nose inboard and forward, tail toward the starboard edge.
	spots := []spot{
		{47, 3.4, 0.04, Gold}, // On catapult 1.
		{57, 8.6, nose, Slate}, {66, 8.6, nose, White}, {75, 8.2, nose, Slate},
		{58, -8.4, -nose, Slate}, {67, -7.8, -nose, Slate},
		{22, 16.8, -math.Pi / 2, Slate}, {3.5, 16.8, -math.Pi / 2, Gold},
		{76, -6.8, -nose, Gold},
		{-47, 10.6, nose, Slate}, {-56, 10.6, nose, White}, {-65, 10.6, nose, Slate}, {-74, 10.6, nose, Slate}, {-83, 10.6, nose, Gold},
	}
	for _, sp := range spots {
		s.jet(g, sp.x, sp.z, fd, sp.angle, Slate, sp.accent)
	}
	// Jet on the lowered elevator, nosing into the hangar.
	s.jet(g, -34, 17.4, aircraftCarrierHD, -math.Pi/2, Slate, Slate)
	s.jet(g, 22, 7.5, aircraftCarrierHD, 0.2, Slate, White) // In the hangar.
	s.hawkeye(g, 12, 8.0, fd, math.Pi+0.35)
	s.helicopter(g, -27, 3.5, fd, 0)
	s.helicopter(g, -27, 8.6, fd, 0)
	for _, t := range []struct{ x, z, a float64 }{{-3, 10.5, 0.3}, {52, 6.5, 1.2}, {-60, 5.6, -0.4}, {-24, 13, 1.57}} {
		s.tractor(g, t.x, t.z, fd, t.a)
	}
	// Crash and salvage fire truck parked beside the island.
	g.Use(Red).Fill(Block(V(-16, fd, 8.6), V(-12, fd+1.6, 10.4)))
	g.Use(White).Fill(Block(V(-16, fd+1.6, 8.8), V(-14.6, fd+2.0, 10.2)))
	g.Paint(White, Block(V(-16, fd+0.6, 8.6), V(-12, fd+0.9, 10.4)))
	// Mobile crane beside the island.
	g.Use(Gold).Fill(Block(V(-4.8, fd, 12), V(-1.8, fd+1.6, 14.4)))
	g.Use(Gold).Fill(Capsule{V(-2.6, fd+1.6, 13.2), V(-6.5, fd+6.2, 13.2), 0.45})
	g.Use(Charcoal).Fill(Capsule{V(-6.5, fd+6.2, 13.2), V(-6.5, fd+3.6, 13.2), 0.42})
}
