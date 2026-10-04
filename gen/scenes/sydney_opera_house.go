package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// SydneyOperaHouse is the opera house on Bennelong Point. It is authored with north
// (the harbor) towards +Z and west (Sydney Cove, Circular Quay) towards +X, then turned
// so north is +X and west -Z: the front view is the classic profile from the east, as
// from Mrs Macquarie's Chair, and the hero view looks in from the north-east. A stepped granite podium carries two halls
// of interlocking white shells, each a pair of spherical sections meeting at a pointed
// ridge, with chevron tile ribs; their mouths step down towards the harbor and are
// closed by bronze glass walls. The Concert Hall is west, the Joan Sutherland Theatre
// east, and the small restaurant shells sit at the south-west corner beside the broad
// monumental stair. The forecourt and the edge of the Botanic Garden lie south; the
// broadwalk runs around the point; a ferry and sailboats cross the harbor.
type SydneyOperaHouse struct{}

func init() { Register(SydneyOperaHouse{}) }

func (SydneyOperaHouse) Name() string  { return "sydney_opera_house" }
func (SydneyOperaHouse) Title() string { return "Sydney Opera House" }

const (
	sydneyOperaHouseWalk   = 1.6  // Broadwalk, forecourt and land level.
	sydneyOperaHouseDeck   = 7.0  // Podium platform where the shells spring.
	sydneyOperaHouseStairS = -38. // Podium south edge, top of the stair.
	sydneyOperaHouseStairN = -49. // Foot of the stair.
	sydneyOperaHouseSiteX  = 48.0 // Site half width.
	sydneyOperaHouseSiteZ0 = -58.0
	sydneyOperaHouseSiteZ1 = 54.0
)

// sydneyOperaHouseShell is one sail: two spherical sections mirrored about a pointed
// ridge. Apex is the plan point of the tip, Face the plan angle the mouth opens to,
// Y0 the springing level and H the tip's height above it.
type sydneyOperaHouseShell struct {
	Apex  Vec2
	Face  float64
	Y0, H float64
}

// geometry of the canonical shell, all proportional to H.
func (s sydneyOperaHouseShell) params() (yc, zc, d, r2, k float64) {
	h := s.H
	yc = s.Y0 - 0.35*h
	zc = 0.491 * h
	rr := 1.437 * h
	d = 0.8 * h
	r2 = rr*rr + d*d
	k = 0.45
	return
}

// mouth is the along-axis position of the open edge at height y: the tip juts forward
// and the edge sweeps back in a concave curve to the feet.
func (s sydneyOperaHouseShell) mouth(y float64) float64 {
	_, _, _, _, k := s.params()
	t := math.Min(math.Max((y-s.Y0)/s.H, 0), 1)
	return -k * s.H * math.Pow(1-t, 0.7)
}

// local returns across (u), along the facing (w) coordinates relative to the apex.
func (s sydneyOperaHouseShell) local(p Vec3) (u, w float64) {
	sn, cs := math.Sincos(s.Face)
	dx, dz := p.X-s.Apex.X, p.Z-s.Apex.Y
	return -dx*sn + dz*cs, dx*cs + dz*sn
}

func (s sydneyOperaHouseShell) box() Box {
	r := 1.1 * s.H
	return Box{V(s.Apex.X-r, s.Y0, s.Apex.Y-r), V(s.Apex.X+r, s.Y0+s.H+1, s.Apex.Y+r)}
}

// measure reports the squared distance from the half's sphere center and whether the
// point lies behind the mouth and above the springing.
func (s sydneyOperaHouseShell) measure(p Vec3) (dist2 float64, ok bool) {
	yc, zc, d, _, _ := s.params()
	if p.Y < s.Y0 {
		return 0, false
	}
	u, w := s.local(p)
	if w > s.mouth(p.Y) {
		return 0, false
	}
	a, b, c := math.Abs(u)+d, w-zc, p.Y-yc
	return a*a + b*b + c*c, true
}

const sydneyOperaHouseT = 1.3 // Shell thickness.

func (s sydneyOperaHouseShell) skin() Solid {
	_, _, _, r2, _ := s.params()
	r := math.Sqrt(r2)
	in2 := (r - sydneyOperaHouseT) * (r - sydneyOperaHouseT)
	return Func{s.box(), func(p Vec3) bool {
		d2, ok := s.measure(p)
		return ok && d2 <= r2 && d2 >= in2
	}}
}

// glass is a slab just inside the mouth, under the shell.
func (s sydneyOperaHouseShell) glass(depth float64) Solid {
	_, _, _, r2, _ := s.params()
	r := math.Sqrt(r2) - sydneyOperaHouseT*0.6
	return Func{s.box(), func(p Vec3) bool {
		d2, ok := s.measure(p)
		if !ok || d2 > r*r {
			return false
		}
		_, w := s.local(p)
		return w >= s.mouth(p.Y)-depth
	}}
}

// ribs selects the cream joints of the tile lids, fanning from the front feet and
// meeting at the ridge in chevrons.
func (s sydneyOperaHouseShell) ribs() Solid {
	_, _, _, _, k := s.params()
	wf := -k * s.H
	step := 5.0 / s.H
	return Func{s.box(), func(p Vec3) bool {
		_, w := s.local(p)
		dy, dw := p.Y-s.Y0, wf-w
		if dy*dy+dw*dw < 9 {
			return false
		}
		phi := math.Atan2(dy, dw)
		f := phi/step - math.Floor(phi/step)
		return f < 0.26
	}}
}

func (s SydneyOperaHouse) Build(out *Grid) Order {
	g := NewGridWithSpan(out.Span())
	s.water(g)
	s.land(g)
	s.podium(g)
	s.stair(g)
	s.halls(g)
	s.garden(g)
	s.lamps(g)
	s.ferry(g, P(40, -24), math.Pi/2+0.15)
	s.sailboat(g, P(-28, 46), 0.4, White)
	s.sailboat(g, P(16, 47), 2.6, White)
	s.sailboat(g, P(-42, -22), math.Pi/2, Red)
	// Turn the plan a quarter: north to +X and west to -Z.
	for _, c := range g.Cells() {
		color, _ := g.ColorOf(c)
		out.Use(color).Add(Cell{X: c.Z, Y: c.Y, Z: -c.X})
	}
	return Along(P(1, 0))
}

// inLand is the plan of the point and the shore along the south edge.
func (SydneyOperaHouse) inLand(x, z float64) bool {
	if z <= -46 && x <= 38 {
		return true
	}
	point := []Vec2{P(-36, -47), P(38, -47), P(38, 10), P(28, 35), P(-26, 35), P(-36, 10)}
	return InsidePolygon(point, P(x, z))
}

func (SydneyOperaHouse) inSite(x, z float64) bool {
	const r = 14.0
	ax := math.Max(math.Abs(x)-(sydneyOperaHouseSiteX-r), 0)
	cz := (sydneyOperaHouseSiteZ0 + sydneyOperaHouseSiteZ1) / 2
	hz := (sydneyOperaHouseSiteZ1 - sydneyOperaHouseSiteZ0) / 2
	az := math.Max(math.Abs(z-cz)-(hz-r), 0)
	return ax*ax+az*az <= r*r
}

func (s SydneyOperaHouse) siteBox(y0, y1 float64) Box {
	return Box{V(-sydneyOperaHouseSiteX, y0, sydneyOperaHouseSiteZ0), V(sydneyOperaHouseSiteX, y1, sydneyOperaHouseSiteZ1)}
}

func (s SydneyOperaHouse) water(g *Grid) {
	box := s.siteBox(-0.2, 0.3)
	g.Use(Water).Fill(Func{box, func(p Vec3) bool {
		return s.inSite(p.X, p.Z) && !s.inLand(p.X, p.Z)
	}})
	// Glints and a few wakes.
	g.Paint(Glass, Func{box, func(p Vec3) bool {
		c := Nearest(p.Scale(1 / g.Span()))
		return Hash01(int(c.X), int(c.Z), 41) < 0.04
	}})
}

func (s SydneyOperaHouse) land(g *Grid) {
	top := sydneyOperaHouseWalk
	box := s.siteBox(-0.2, top)
	in := func(x, z float64) bool { return s.inSite(x, z) && s.inLand(x, z) }
	// Paving slab and harbor wall: a hollow skin.
	g.Use(Stone).Fill(Func{box, func(p Vec3) bool {
		if !in(p.X, p.Z) {
			return false
		}
		if p.Y >= top-0.9 {
			return true
		}
		for _, d := range []Vec2{P(1.2, 0), P(-1.2, 0), P(0, 1.2), P(0, -1.2)} {
			if !in(p.X+d.X, p.Z+d.Y) {
				return true
			}
		}
		return false
	}})
	// Sandstone sea wall face and paved forecourt in warm granite.
	g.Paint(DarkStone, Func{box, func(p Vec3) bool { return p.Y < top-0.5 && in(p.X, p.Z) }})
	g.Paint(Sand, Func{s.siteBox(top-0.5, top), func(p Vec3) bool {
		c := Nearest(p.Scale(1 / g.Span()))
		gx, gz := math.Floor(p.X/4), math.Floor(p.Z/4)
		return int(gx+gz)%2 == 0 || Hash01(int(c.X), int(c.Z), 3) < 0.2
	}})
}

// podium is the stepped granite platform: a brick plinth, terracotta walls with dark
// openings and a projecting sand cornice, paved on top.
func (s SydneyOperaHouse) podium(g *Grid) {
	walk, deck := sydneyOperaHouseWalk, sydneyOperaHouseDeck
	poly := func(inset float64) []Vec2 {
		return []Vec2{
			P(-32+inset, sydneyOperaHouseStairS+inset), P(34-inset, sydneyOperaHouseStairS+inset),
			P(34-inset, 8), P(25-inset, 30-inset), P(-23+inset, 30-inset), P(-32+inset, 8),
		}
	}
	box := Box{V(-33, walk, -40), V(35, deck+0.6, 36)}
	inPoly := func(inset float64, p Vec3) bool { return InsidePolygon(poly(inset), p.Plan()) }
	// Plinth course.
	g.Use(Brick).Fill(Func{box, func(p Vec3) bool {
		return p.Y <= walk+1.2 && inPoly(-0.8, p) && !inPoly(1.2, p)
	}})
	// Walls.
	g.Use(Terracotta).Fill(Func{box, func(p Vec3) bool {
		return p.Y <= deck-0.4 && inPoly(0, p) && !inPoly(1.2, p)
	}})
	// Deck slab and cornice.
	g.Use(Sand).Fill(Func{box, func(p Vec3) bool {
		return p.Y > deck-0.9 && p.Y <= deck && inPoly(-0.6, p)
	}})
	// Paving joints on the deck.
	g.Paint(Stone, Func{Box{V(-33, deck-0.4, -40), V(35, deck, 36)}, func(p Vec3) bool {
		return inPoly(1, p) && (math.Mod(p.X+100.5, 9) < 0.8 || math.Mod(p.Z+100.5, 9) < 0.8)
	}})
	// Openings into the concourse along the sides: tall dark bays between piers.
	g.Carve(Func{box, func(p Vec3) bool {
		if p.Y < walk+1.3 || p.Y > walk+3.6 || !inPoly(-0.1, p) || inPoly(1.3, p) {
			return false
		}
		along := p.X + p.Z
		if p.X > 33 || p.X < -31 {
			along = p.Z
		}
		return math.Mod(along+200, 4) < 2.4
	}})
	g.Use(Charcoal).Fill(Func{box, func(p Vec3) bool {
		return p.Y >= walk+1.3 && p.Y <= walk+3.6 && inPoly(1.2, p) && !inPoly(1.9, p)
	}})
}

// stair is the monumental flight from the forecourt to the podium, with granite
// cheeks; the steps are a hollow folded slab.
func (s SydneyOperaHouse) stair(g *Grid) {
	walk, deck := sydneyOperaHouseWalk, sydneyOperaHouseDeck
	z0, z1 := sydneyOperaHouseStairN, sydneyOperaHouseStairS
	x0, x1 := -24.0, 19.0
	rise := deck - walk
	steps := 11.0
	tread := (z1 - z0) / steps
	stepTop := func(z float64) float64 {
		n := math.Floor((z-z0)/tread) + 1
		return walk + math.Min(n, steps)*rise/steps
	}
	box := Box{V(x0-2, walk, z0), V(x1+2, deck, z1)}
	g.Use(Sand).Fill(Func{box, func(p Vec3) bool {
		if p.X < x0 || p.X > x1 {
			return false
		}
		t := stepTop(p.Z)
		return p.Y <= t && p.Y >= t-1.3
	}})
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		t := stepTop(p.Z)
		return p.Y > t-0.4 && math.Mod(p.Z-z0, tread) < 0.3
	}})
	// Cheek walls.
	for _, x := range []float64{x0 - 1, x1 + 1} {
		g.Use(Terracotta).Fill(Func{Box{V(x-1, walk, z0), V(x+1, deck, z1)}, func(p Vec3) bool {
			return p.Y <= walk+(p.Z-z0)/(z1-z0)*rise+0.8
		}})
	}
}

// halls places the shell sets. Shells are listed by their distance along each hall's
// axis; negative heights face back towards the stair.
func (s SydneyOperaHouse) halls(g *Grid) {
	deck := sydneyOperaHouseDeck
	type spec struct{ s, h float64 }
	var shells []sydneyOperaHouseShell
	hall := func(pivot Vec2, angle, scale float64, list []spec) {
		dir := P(math.Cos(angle), math.Sin(angle))
		for _, sp := range list {
			face := angle
			h := sp.h
			if h < 0 {
				face += math.Pi
				h = -h
			}
			shells = append(shells, sydneyOperaHouseShell{
				Apex: pivot.Add(dir.Scale(sp.s * scale)), Face: face, Y0: deck, H: h * scale,
			})
		}
	}
	// Concert Hall (west, +X) and Joan Sutherland Theatre (east).
	main := []spec{{-11, -13}, {24, 28}, {34.4, 21}, {41.5, 15}}
	hall(P(16.5, -21), math.Pi/2-0.05, 1.07, main)
	hall(P(-15.5, -20), math.Pi/2+0.05, 0.96, main)
	// Bennelong restaurant shells facing the city.
	shells = append(shells,
		sydneyOperaHouseShell{Apex: P(28.5, -31.5), Face: -math.Pi / 2, Y0: deck, H: 9},
		sydneyOperaHouseShell{Apex: P(28.5, -35.6), Face: -math.Pi / 2, Y0: deck, H: 6.2},
		sydneyOperaHouseShell{Apex: P(28.5, -19.5), Face: math.Pi / 2, Y0: deck, H: 5},
	)
	for _, sh := range shells {
		g.Use(Glass).Fill(sh.glass(0.9))
	}
	for _, sh := range shells {
		g.Paint(Charcoal, Intersect(sh.glass(0.9), Func{sh.box(), func(p Vec3) bool {
			u, _ := sh.local(p)
			return math.Mod(math.Abs(u)+0.6, 3.2) < 1.2 || math.Mod(p.Y-sh.Y0+0.5, 4) < 1.0
		}}))
	}
	for _, sh := range shells {
		g.Use(White).FillVolume(sh.skin())
	}
	for _, sh := range shells {
		g.Paint(Stone, Intersect(sh.skin(), sh.ribs()))
	}
}

// lamps lines the broadwalk around the point with dark posts and gold lanterns.
func (s SydneyOperaHouse) lamps(g *Grid) {
	walk := sydneyOperaHouseWalk
	edge := []Vec2{P(-34.5, -40), P(-34.5, 9.5), P(-25, 33.5), P(27, 33.5), P(36.5, 9.5), P(36.5, -40)}
	for i := 0; i+1 < len(edge); i++ {
		a, b := edge[i], edge[i+1]
		n := int(math.Round(b.Sub(a).Length() / 7))
		for j := 0; j <= n; j++ {
			if j == n && i+2 < len(edge) {
				continue
			}
			q := a.Add(b.Sub(a).Scale(float64(j) / float64(n)))
			g.Use(Charcoal).Fill(Cylinder{Center: q, Radius: 0.45, Y0: walk, Y1: walk + 3})
			g.Use(Gold).Fill(Sphere(q.At(walk+3.3), 0.6))
		}
	}
}

// garden is the edge of the Royal Botanic Garden along the south-east shore.
func (s SydneyOperaHouse) garden(g *Grid) {
	walk := sydneyOperaHouseWalk
	lawn := func(p Vec3) bool { return p.X < -26 && p.Z < -52 && s.inSite(p.X, p.Z) }
	g.Paint(Grass, Func{s.siteBox(walk-0.5, walk), lawn})
	for i := 0; i < 9; i++ {
		x := -28 - 18*Hash01(i, 1)
		z := -51 - 6*Hash01(i, 2)
		if !s.inSite(x, z) {
			continue
		}
		h := 4 + 2.5*Hash01(i, 3)
		g.Use(Wood).Fill(Cylinder{Center: P(x, z), Radius: 0.6, Y0: walk, Y1: walk + h*0.55})
		g.Use(Foliage).FillVolume(Ellipsoid{P(x, z).At(walk + h*0.75), V(h*0.4, h*0.3, h*0.4)})
	}
}

// ferry is a green-and-cream harbor ferry heading along angle.
func (s SydneyOperaHouse) ferry(g *Grid, c Vec2, angle float64) {
	const L, B = 16.0, 5.0
	at := func(p Vec3) (float64, float64) {
		d := p.Plan().Sub(c).Rotate(-angle)
		return d.X, d.Y
	}
	box := Box{V(c.X-L, 0, c.Y-L), V(c.X+L, 6, c.Y+L)}
	hull := func(p Vec3) bool {
		a, b := at(p)
		// Double-ended: pointed at both ends.
		half := B / 2 * math.Min(1, (L/2-math.Abs(a))/3)
		return math.Abs(a) <= L/2 && math.Abs(b) <= half
	}
	g.Use(Foliage).Fill(Func{box, func(p Vec3) bool { return p.Y <= 1.6 && hull(p) }})
	g.Use(Sand).Fill(Func{box, func(p Vec3) bool {
		a, b := at(p)
		return p.Y > 1.6 && p.Y <= 3.6 && math.Abs(a) <= L/2-3 && math.Abs(b) <= B/2-0.4
	}})
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		a, _ := at(p)
		return p.Y > 2.3 && p.Y <= 3.0 && math.Mod(math.Abs(a)+0.3, 1.6) < 1.0
	}})
	g.Use(Foliage).Fill(Func{box, func(p Vec3) bool {
		a, b := at(p)
		return p.Y > 3.6 && p.Y <= 4.2 && math.Abs(a) <= L/2-2.6 && math.Abs(b) <= B/2-0.1
	}})
	g.Use(White).Fill(Func{box, func(p Vec3) bool {
		a, b := at(p)
		return p.Y > 4.2 && p.Y <= 5.6 && math.Abs(a) <= 1.6 && math.Abs(b) <= 1.4
	}})
	g.Use(Gold).Fill(Cylinder{Center: c, Radius: 0.6, Y0: 5.6, Y1: 6.6})
}

// sailboat is a small yacht with mainsail and jib.
func (s SydneyOperaHouse) sailboat(g *Grid, c Vec2, angle float64, sail Color) {
	const L = 8.5
	at := func(p Vec3) (float64, float64) {
		d := p.Plan().Sub(c).Rotate(-angle)
		return d.X, d.Y
	}
	box := Box{V(c.X-L, 0, c.Y-L), V(c.X+L, 12, c.Y+L)}
	g.Use(White).Fill(Func{box, func(p Vec3) bool {
		a, b := at(p)
		if p.Y > 1.2 || math.Abs(a) > L/2 {
			return false
		}
		half := 1.3 * math.Min(1, (L/2-a)/2.5)
		return math.Abs(b) <= half
	}})
	g.Paint(Navy, Func{box, func(p Vec3) bool { return p.Y > 0.6 && p.Y <= 1.0 }})
	g.Use(Wood).Fill(Capsule{c.At(1), c.At(12.5), 0.5})
	g.Use(sail).Fill(Func{box, func(p Vec3) bool {
		a, b := at(p)
		if p.Y < 2 || p.Y > 12 || math.Abs(b) > 0.45 {
			return false
		}
		t := (p.Y - 2) / 10
		// Main abaft the mast, jib forward.
		return (a <= -0.5 && a >= -0.5-4*(1-t)) || (a >= 0.5 && a <= 0.5+3.2*(1-t) && p.Y < 11)
	}})
}
