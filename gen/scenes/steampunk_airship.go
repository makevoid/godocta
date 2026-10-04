package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// SteampunkAirship is a Victorian airship moored nose-first to an iron-lattice mast on a
// hilltop sky-dock. The patchwork canvas envelope is banded in brass, with striped
// rudders on cruciform fins; a wooden ship-hull gondola hangs beneath on struts, with
// portholes, a glazed bridge, a saloon, a figurehead and four brass steam engines on
// outriggers, each with a smokestack and a pusher propeller. A gangway links the
// gondola to the mast's landing; the cobbled plaza has a warehouse, crane and cargo.
// Recipes are authored with the bow toward +X; Build mirrors the site so the finished
// build has its bow and mast toward -X, and the hero view looks along it from the tail.
type SteampunkAirship struct{}

func init() { Register(SteampunkAirship{}) }

func (SteampunkAirship) Name() string  { return "steampunk_airship" }
func (SteampunkAirship) Title() string { return "Steampunk Airship" }

const (
	steampunkAirshipAxisY = 36.0  // Envelope axis height.
	steampunkAirshipR     = 13.0  // Envelope radius.
	steampunkAirshipNose  = 42.5  // Nose tip X.
	steampunkAirshipTail  = -48.0 // Tail tip X.
	steampunkAirshipShell = 1.3   // Envelope skin thickness.
	steampunkAirshipDeck  = 14.0  // Gondola deck height.
	steampunkAirshipKeel  = 8.0   // Gondola keel height.
	steampunkAirshipBeam  = 5.0   // Gondola half beam.
	steampunkAirshipPlaza = 5.0   // Plaza height.
	steampunkAirshipPlazR = 17.0  // Plaza radius.
	steampunkAirshipSiteA = 61.0  // Site ellipse semi-axes and center X.
	steampunkAirshipSiteB = 28.0
	steampunkAirshipSiteX = 6.0
)

// steampunkAirshipMast is the mooring mast's plan center.
var steampunkAirshipMast = P(46, 0)

// radius is the envelope's radius at x, shrunk by inset; negative outside the envelope.
func (SteampunkAirship) radius(x, inset float64) float64 {
	nose0, tail0 := 24.0, -12.0
	noseEnd, tailEnd := steampunkAirshipNose-inset, steampunkAirshipTail+inset
	r := steampunkAirshipR - inset
	switch {
	case x > noseEnd || x < tailEnd:
		return -1
	case x > nose0:
		t := (x - nose0) / (noseEnd - nose0)
		return r * math.Sqrt(1-t*t)
	case x < tail0:
		t := (tail0 - x) / (tail0 - tailEnd)
		return r * math.Pow(1-t*t, 0.6)
	}
	return r
}

// hullWidth is the gondola's plan half-width at x and height y, shrunk by inset.
func (SteampunkAirship) hullWidth(x, y, inset float64) float64 {
	const stern, bowStart = -16.0, 17.0
	bowEnd := 29 - (steampunkAirshipDeck-y)*0.5
	if x < stern || x > bowEnd || y < steampunkAirshipKeel {
		return -1
	}
	w := steampunkAirshipBeam
	switch {
	case x > bowStart:
		t := (x - bowStart) / (bowEnd - bowStart)
		w *= 1 - t*t
	case x < stern+8:
		t := (stern + 8 - x) / 8
		w -= 1.2 * t * t
	}
	if y < steampunkAirshipKeel+2.8 {
		t := (steampunkAirshipKeel + 2.8 - y) / 2.8
		w *= 1 - 0.6*t*t
	}
	return w - inset
}

// ground is the terrain height: a cobbled plateau around the mast falling smoothly to a
// low rim on the site ellipse. ok is false outside the site.
func (SteampunkAirship) ground(x, z float64) (h, s float64, ok bool) {
	ex, ez := (x-steampunkAirshipSiteX)/steampunkAirshipSiteA, z/steampunkAirshipSiteB
	if ex*ex+ez*ez > 1 {
		return 0, 1, false
	}
	d := P(x, z).Sub(steampunkAirshipMast)
	dist := d.Length()
	if dist <= steampunkAirshipPlazR {
		return steampunkAirshipPlaza, 0, true
	}
	u := d.Scale(1 / dist)
	A2, B2 := steampunkAirshipSiteA*steampunkAirshipSiteA, steampunkAirshipSiteB*steampunkAirshipSiteB
	qx, qz := steampunkAirshipMast.X-steampunkAirshipSiteX, steampunkAirshipMast.Y
	a := u.X*u.X/A2 + u.Y*u.Y/B2
	b := 2 * (qx*u.X/A2 + qz*u.Y/B2)
	c := qx*qx/A2 + qz*qz/B2 - 1
	tb := (-b + math.Sqrt(b*b-4*a*c)) / (2 * a)
	s = math.Min(1, math.Max(0, (dist-steampunkAirshipPlazR)/(tb-steampunkAirshipPlazR)))
	fall := s * s * (3 - 2*s)
	wave := 0.7 * math.Sin(x*0.17+1.3) * math.Sin(z*0.21+0.4) * s * (1 - s) * 4
	return 0.6 + (steampunkAirshipPlaza-0.6)*(1-fall) + wave, s, true
}

func (a SteampunkAirship) Build(g *Grid) Order {
	a.terrain(g)
	a.plaza(g)
	a.mast(g)
	a.envelope(g)
	a.fins(g)
	a.gondola(g)
	a.engines(g)
	a.rigging(g)
	a.mirror(g)
	return Outward(P(-steampunkAirshipMast.X, steampunkAirshipMast.Y))
}

// terrain lays a hollow grassy hill with meadow patches, a sand path and a rocky rim.
func (a SteampunkAirship) terrain(g *Grid) {
	box := Box{V(steampunkAirshipSiteX-steampunkAirshipSiteA, 0, -steampunkAirshipSiteB),
		V(steampunkAirshipSiteX+steampunkAirshipSiteA, steampunkAirshipPlaza+1, steampunkAirshipSiteB)}
	g.Use(Grass).Fill(Func{box, func(p Vec3) bool {
		h, _, ok := a.ground(p.X, p.Z)
		return ok && p.Y <= h && p.Y >= h-1.1
	}})
	g.Paint(Foliage, Func{box, func(p Vec3) bool {
		n := math.Sin(p.X*0.23+p.Z*0.11) + math.Sin(p.Z*0.31-p.X*0.07+2) + 0.6*math.Sin(p.X*0.5+p.Z*0.45)
		return n > 1.2
	}})
	// Rocky rim.
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		ex, ez := (p.X-steampunkAirshipSiteX)/steampunkAirshipSiteA, p.Z/steampunkAirshipSiteB
		return math.Hypot(ex, ez) > 0.93-0.03*Hash01(int(math.Floor(p.X)), int(math.Floor(p.Z)))
	}})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		ex, ez := (p.X-steampunkAirshipSiteX)/steampunkAirshipSiteA, p.Z/steampunkAirshipSiteB
		return math.Hypot(ex, ez) > 0.91 && Hash01(int(math.Floor(p.X*1.3)), int(math.Floor(p.Z*1.3)), 3) < 0.3
	}})
	// Winding sand path from the plaza down the +Z slope.
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		if p.Z < 10 {
			return false
		}
		cx := 50 + 4*math.Sin(p.Z/4.5)
		return math.Abs(p.X-cx) < 1.4
	}})
	// Gas lamps along the path.
	for k, z := range []float64{13, 18.5, 24} {
		side := float64(k%2*2 - 1)
		c := P(50+4*math.Sin(z/4.5)+side*2.4, z)
		h, _, ok := a.ground(c.X, c.Y)
		if !ok {
			continue
		}
		g.With(Charcoal, func() { g.Fill(Cylinder{Center: c, Radius: 0.4, Y0: h - 0.5, Y1: h + 3.6}) })
		g.With(Gold, func() { g.Fill(BlockAt(c, 1.1, 1.1, h+3.6, h+4.6)) })
		g.With(Charcoal, func() { g.Fill(Cone(c, 0.75, h+4.6, 0.7)) })
	}
	// Boulders along the rim.
	for i := 0; i < 26; i++ {
		ang := Hash01(i, 31) * 2 * math.Pi
		x := steampunkAirshipSiteX + math.Cos(ang)*steampunkAirshipSiteA*0.86
		z := math.Sin(ang) * steampunkAirshipSiteB * 0.86
		h, _, ok := a.ground(x, z)
		if !ok {
			continue
		}
		r := 1.2 + 1.2*Hash01(i, 32)
		col := Stone
		if Hash01(i, 33) < 0.5 {
			col = DarkStone
		}
		g.With(col, func() { g.FillVolume(Ellipsoid{V(x, h, z), V(r*1.3, r*0.8, r)}) })
	}
	// Trees on the flanks, clear of the airship.
	for i := 0; i < 160; i++ {
		x := steampunkAirshipSiteX + (Hash01(i, 11)*2-1)*steampunkAirshipSiteA*0.9
		z := (Hash01(i, 12)*2 - 1) * steampunkAirshipSiteB * 0.9
		az := math.Abs(z)
		switch {
		case x > 27 && (az < 20 || x < 62):
			continue // The plaza and mast.
		case x > -20 && az < 12.5:
			continue // Under the gondola, paddle wheels and engines.
		case math.Abs(x-50-4*math.Sin(z/4.5)) < 4 && z > 0:
			continue // The path.
		}
		h, s, ok := a.ground(x, z)
		if !ok || s > 0.8 {
			continue
		}
		size := 2.2 + 1.4*Hash01(i, 13)
		c := P(x, z)
		g.With(Wood, func() { g.Fill(Cylinder{Center: c, Radius: 0.6, Y0: h - 0.5, Y1: h + size*1.2}) })
		g.With(Foliage, func() { g.FillVolume(Ellipsoid{c.At(h + size*1.5), V(size, size*0.85, size)}) })
	}
}

// plaza cobbles the hilltop and adds the warehouse, crane, cargo, lamps and station.
func (a SteampunkAirship) plaza(g *Grid) {
	m := steampunkAirshipMast
	top := steampunkAirshipPlaza
	plaza := Cylinder{Center: m, Radius: steampunkAirshipPlazR, Y0: top - 0.6, Y1: top}
	g.Paint(Stone, plaza)
	g.Paint(DarkStone, Func{plaza.Bounds(), func(p Vec3) bool {
		return plaza.Contains(p) && Hash01(int(math.Floor(p.X*1.26)), int(math.Floor(p.Z*1.26)), 5) < 0.28
	}})
	// Curb ring with lamp posts.
	g.Use(DarkStone).Fill(Cylinder{Center: m, Radius: steampunkAirshipPlazR, Inner: steampunkAirshipPlazR - 0.9, Y0: top, Y1: top + 0.6})
	for k := 0; k < 8; k++ {
		ang := float64(k)*math.Pi/4 + math.Pi/8
		c := Polar(m, steampunkAirshipPlazR-0.5, ang)
		if c.X < 32 && math.Abs(c.Y) < 8 {
			continue
		}
		g.With(Charcoal, func() { g.Fill(Cylinder{Center: c, Radius: 0.45, Y0: top, Y1: top + 4}) })
		g.With(Gold, func() { g.Fill(BlockAt(c, 1.2, 1.2, top+4, top+5)) })
		g.With(Charcoal, func() { g.Fill(Cone(c, 0.8, top+5, 0.8)) })
	}
	// Warehouse: brick hall with gold-lit windows, slate roof and a dark cargo door.
	w0, w1 := P(38, -15), P(54, -8)
	g.Use(Brick).Fill(Block(w0.At(top-2), w1.At(top+7)))
	g.Carve(Block(V(w0.X+1, top, w0.Y+1), V(w1.X-1, top+7, w1.Y-1)))
	g.Paint(Charcoal, Func{Box{w0.At(top), w1.At(top + 7)}, func(p Vec3) bool {
		return p.X > w0.X+0.5 && p.X < w1.X-0.5 && p.Z > w0.Y+0.5 && p.Z < w1.Y-0.5
	}})
	g.Use(Slate).Fill(Gable{Min: P(w0.X-0.6, w0.Y-0.6), Max: P(w1.X+0.6, w1.Y+0.6), Base: top + 7, Ridge: AxisX})
	g.Carve(Gable{Min: P(w0.X+0.4, w0.Y+0.8), Max: P(w1.X-0.4, w1.Y-0.8), Base: top + 7, Ridge: AxisX})
	// Gable ends in brick.
	g.Use(Brick).Fill(Gable{Min: P(w0.X, w0.Y), Max: P(w0.X+1, w1.Y), Base: top + 7, Ridge: AxisX},
		Gable{Min: P(w1.X-1, w0.Y), Max: P(w1.X, w1.Y), Base: top + 7, Ridge: AxisX})
	g.Use(Brick).Fill(BlockAt(P(50, -11.5), 1.4, 1.4, top+7, top+12))
	g.Paint(Charcoal, BlockAt(P(50, -11.5), 1.4, 1.4, top+11.5, top+12))
	for x := w0.X + 2.5; x < w1.X-1.5; x += 3 {
		if math.Abs(x-46) < 3 {
			continue
		}
		for _, z := range []float64{w0.Y, w1.Y} {
			g.Paint(Gold, BlockAt(P(x, z), 1.2, 1, top+2.5, top+5))
		}
	}
	g.Paint(Gold, BlockAt(P(w0.X, -11.5), 1, 1.6, top+3, top+5.5), BlockAt(P(w1.X, -11.5), 1, 1.6, top+3, top+5.5))
	g.Paint(Gold, Block(V(w0.X, top+8.5, -11.8), V(w0.X+0.5, top+9.5, -11.2)), Block(V(w1.X-0.5, top+8.5, -11.8), V(w1.X, top+9.5, -11.2)))
	g.Carve(Arch{Sill: V(46, top, w1.Y-0.5), HalfWidth: 1.8, Spring: 3, Depth: 1.6, Through: AxisZ})
	g.Paint(Wood, Block(V(44, top+5, w1.Y-0.5), V(48, top+5.5, w1.Y)))
	// Crane: wooden post and jib over the cargo, with a hanging crate.
	cr := P(35, 11)
	g.Use(Wood).Fill(Cylinder{Center: cr, Radius: 0.9, Y0: top, Y1: top + 10})
	g.Use(DarkStone).Fill(Cylinder{Center: cr, Radius: 1.6, Y0: top, Y1: top + 1})
	tip := V(29.5, top+11.5, 15)
	g.Use(Wood).Fill(Capsule{cr.At(top + 9), tip, 0.6}, Capsule{cr.At(top + 6), V(31.5, top+10.8, 14), 0.5})
	g.Use(Charcoal).Fill(Capsule{cr.At(top + 10), cr.At(top + 9).Add(V(2, 0, -0.5)), 0.45})
	g.Fill(Block(V(cr.X+1.6, top+7.5, cr.Y-1.2), V(cr.X+3.2, top+9.5, cr.Y+0.2)))
	g.Fill(Capsule{tip, tip.Sub(V(0, 5, 0)), 0.42})
	g.Use(Wood).Fill(BlockAt(P(29.5, 15), 1.8, 1.8, top+3.6, top+5.4))
	g.Paint(Gold, BlockAt(P(29.5, 15), 1.8, 1.8, top+5, top+5.4))
	// Cargo: crates and barrels.
	for i := 0; i < 26; i++ {
		c := P(32+Hash01(i, 21)*14, 7.5+Hash01(i, 22)*8)
		if c.Sub(m).Length() > steampunkAirshipPlazR-1.5 || c.Sub(cr).Length() < 2 || (c.X > 37.5 && c.Y < 9) {
			continue
		}
		if Hash01(i, 23) < 0.55 {
			col := Wood
			if Hash01(i, 24) < 0.3 {
				col = Terracotta
			}
			s := 1.2 + 0.6*Hash01(i, 25)
			stack := 1.0
			if Hash01(i, 26) < 0.35 {
				stack = 2
			}
			g.With(col, func() { g.Fill(BlockAt(c, s, s, top, top+s*stack)) })
			g.Paint(Sand, BlockAt(c, s, s, top+s*stack-0.3, top+s*stack))
		} else {
			g.With(Wood, func() { g.Fill(Cylinder{Center: c, Radius: 0.75, Y0: top, Y1: top + 1.8}) })
			g.Paint(Charcoal, Cylinder{Center: c, Radius: 0.8, Y0: top + 0.4, Y1: top + 0.6}, Cylinder{Center: c, Radius: 0.8, Y0: top + 1.3, Y1: top + 1.5})
		}
	}
	a.station(g)
}

// station is the dock's booking office: a brick house with gold-lit windows, a hipped
// slate roof and a clock turret.
func (SteampunkAirship) station(g *Grid) {
	top := steampunkAirshipPlaza
	c := P(54, 9.5)
	w, d := 7.0, 6.0
	g.Use(Sand).Fill(BlockAt(c, w, d, top-1, top+5))
	g.Carve(BlockAt(c, w-2, d-2, top, top+5))
	g.Paint(Charcoal, BlockAt(c, w-1, d-1, top, top+5))
	g.Paint(Brick, BlockAt(c, w, d, top-1, top+0.6), BlockAt(c, w, d, top+4.4, top+5))
	g.Paint(Brick, Func{BlockAt(c, w, d, top, top+5).Bounds(), func(p Vec3) bool {
		return math.Abs(math.Abs(p.X-c.X)-w/2) < 0.6 && math.Abs(math.Abs(p.Z-c.Y)-d/2) < 0.6
	}})
	for _, dx := range []float64{-1.8, 1.8} {
		g.Paint(Gold, BlockAt(P(c.X+dx, c.Y), 1.2, d, top+1.8, top+3.6))
	}
	g.Paint(Gold, BlockAt(c, w, 1.2, top+1.8, top+3.6))
	g.Carve(Arch{Sill: V(c.X, top, c.Y-d/2+0.5), HalfWidth: 0.9, Spring: 2, Depth: 1.2, Through: AxisZ})
	g.Use(Slate).Fill(HipRoof{Min: P(c.X-w/2-0.6, c.Y-d/2-0.6), Max: P(c.X+w/2+0.6, c.Y+d/2+0.6), Base: top + 5})
	g.Carve(HipRoof{Min: P(c.X-w/2+0.6, c.Y-d/2+0.6), Max: P(c.X+w/2-0.6, c.Y+d/2-0.6), Base: top + 5})
	// Clock turret.
	g.Use(Brick).Fill(BlockAt(c, 2.4, 2.4, top+6, top+10))
	g.Paint(Gold, BlockAt(c, 2.4, 2.4, top+8, top+9.4))
	g.Paint(White, BlockAt(c, 1.2, 2.4, top+8.2, top+9.2), BlockAt(c, 2.4, 1.2, top+8.2, top+9.2))
	g.Use(Charcoal).Fill(HipRoof{Min: P(c.X-1.7, c.Y-1.7), Max: P(c.X+1.7, c.Y+1.7), Base: top + 10})
	g.Use(Gold).Fill(Cylinder{Center: c, Radius: 0.4, Y0: top + 11.5, Y1: top + 13})
}

// mastHalf is the half-width of the mooring mast at height y.
func (SteampunkAirship) mastHalf(y float64) float64 {
	return 5 - 1.4*(y-steampunkAirshipPlaza)/(34-steampunkAirshipPlaza)
}

// mast is an iron-lattice mooring tower with a spiral stair, a gangway landing, a top
// platform with the mooring arm and cup, and a beacon lantern.
func (a SteampunkAirship) mast(g *Grid) {
	m := steampunkAirshipMast
	base, top := steampunkAirshipPlaza, 34.0
	levels := []float64{base + 0.4, steampunkAirshipDeck - 0.3, 20.5, 27.5, top}
	corners := [4][2]float64{{1, 1}, {1, -1}, {-1, -1}, {-1, 1}}
	corner := func(k int, y float64) Vec3 {
		h := a.mastHalf(y)
		return V(m.X+corners[k%4][0]*h, y, m.Y+corners[k%4][1]*h)
	}
	g.Use(Charcoal)
	// Stone footings and legs.
	for k := 0; k < 4; k++ {
		c := corner(k, base)
		g.With(DarkStone, func() { g.Fill(BlockAt(c.Plan(), 2.2, 2.2, base-0.5, base+1)) })
		g.Fill(Capsule{corner(k, base), corner(k, top), 0.65})
	}
	// Girder rings and X-bracing on every face.
	for i, y := range levels {
		h := a.mastHalf(y)
		g.Fill(Func{Box{V(m.X-h-0.7, y-0.4, m.Y-h-0.7), V(m.X+h+0.7, y+0.4, m.Y+h+0.7)}, func(p Vec3) bool {
			d := math.Max(math.Abs(p.X-m.X), math.Abs(p.Z-m.Y))
			return math.Abs(d-h) <= 0.55
		}})
		if i == 0 {
			continue
		}
		y0 := levels[i-1]
		for k := 0; k < 4; k++ {
			if (i+k)%2 == 0 {
				g.Fill(Capsule{corner(k, y0), corner(k+1, y), 0.45})
			} else {
				g.Fill(Capsule{corner(k+1, y0), corner(k, y), 0.45})
			}
		}
	}
	// Spiral stair around a central column.
	g.Fill(Cylinder{Center: m, Radius: 0.75, Y0: base, Y1: top + 5})
	const pitch, steps = 5.5, 18.0
	g.With(Wood, func() {
		g.Fill(Func{Box{V(m.X-2.8, base, m.Y-2.8), V(m.X+2.8, top, m.Y+2.8)}, func(p Vec3) bool {
			d := p.Plan().Sub(m)
			r := d.Length()
			if r < 0.7 || r > 2.8 {
				return false
			}
			turn := math.Floor((d.Angle()+math.Pi)/(2*math.Pi)*steps) / steps
			return math.Mod(p.Y-base-turn*pitch+100*pitch, pitch) < 0.7
		}})
	})
	g.With(Gold, func() {
		g.Fill(Func{Box{V(m.X-3.1, base+1, m.Y-3.1), V(m.X+3.1, top, m.Y+3.1)}, func(p Vec3) bool {
			d := p.Plan().Sub(m)
			r := d.Length()
			if r < 2.5 || r > 3.1 {
				return false
			}
			turn := (d.Angle() + math.Pi) / (2 * math.Pi)
			return math.Mod(p.Y-base-turn*pitch-1.2+100*pitch, pitch) < 0.6
		}})
	})
	// Gangway landing: a deck ring around the stair, with a balcony on the +Z side.
	deck := steampunkAirshipDeck
	hd := a.mastHalf(deck) + 0.6
	g.With(Wood, func() {
		g.Fill(Subtract(Block(V(m.X-hd, deck-0.6, m.Y-hd), V(m.X+hd, deck, m.Y+hd+2)), Cylinder{Center: m, Radius: 3, Y0: deck - 1, Y1: deck + 1}))
	})
	g.With(Gold, func() {
		g.Fill(Func{Box{V(m.X-hd, deck, m.Y-hd), V(m.X+hd, deck+1.6, m.Y+hd+2)}, func(p Vec3) bool {
			edge := math.Abs(p.X-m.X) > hd-0.6 || p.Z < m.Y-hd+0.6 || p.Z > m.Y+hd+1.4
			if !edge || (p.X < m.X-hd+0.8 && p.Z > m.Y+3.2) {
				return false
			}
			return p.Y > deck+1 || math.Mod(p.X+p.Z+100, 1.6) < 0.7
		}})
	})
	// Top platform.
	ht := a.mastHalf(top) + 0.8
	g.With(Wood, func() { g.Fill(Block(V(m.X-ht, top-0.6, m.Y-ht), V(m.X+ht, top, m.Y+ht))) })
	g.With(Gold, func() {
		g.Fill(Func{Box{V(m.X-ht, top, m.Y-ht), V(m.X+ht, top+1.6, m.Y+ht)}, func(p Vec3) bool {
			edge := math.Max(math.Abs(p.X-m.X), math.Abs(p.Z-m.Y)) > ht-0.6
			return edge && (p.Y > top+1 || math.Mod(p.X+p.Z+100, 1.6) < 0.7)
		}})
	})
	// Mooring arm and cup gripping the envelope nose.
	ay := steampunkAirshipAxisY
	g.Use(Charcoal).Fill(Capsule{V(m.X, ay, m.Y), V(steampunkAirshipNose+0.6, ay, m.Y), 0.85})
	g.Fill(Capsule{V(m.X-ht+0.6, top, m.Y), V(steampunkAirshipNose+1.5, ay-0.6, m.Y), 0.55})
	g.Fill(Func{Box{V(steampunkAirshipNose-2.5, ay-3.5, -3.5), V(steampunkAirshipNose+1, ay+3.5, 3.5)}, func(p Vec3) bool {
		r := math.Hypot(p.Y-ay, p.Z)
		outer := 1.6 + (steampunkAirshipNose+1-p.X)*0.75
		return r <= outer && r > outer-0.8
	}})
	// Lantern and beacon.
	lt := ay + 1.5
	g.Use(Gold).Fill(Cylinder{Center: m, Radius: 1.2, Y0: lt, Y1: lt + 1.6})
	g.Use(Red).Fill(Sphere(m.At(lt+2.4), 0.9))
	g.Use(Charcoal).Fill(Cone(m, 1.5, lt+1.6, 0.5), Cylinder{Center: m, Radius: 0.45, Y0: lt + 3, Y1: lt + 5})
	g.Use(Gold).Fill(Sphere(m.At(lt+5.3), 0.6))
}

// envelope is the hollow cigar of patchwork canvas with brass bands, ribs and caps.
func (a SteampunkAirship) envelope(g *Grid) {
	ay, R := steampunkAirshipAxisY, steampunkAirshipR
	box := Box{V(steampunkAirshipTail, ay-R, -R), V(steampunkAirshipNose, ay+R, R)}
	shell := Func{box, func(p Vec3) bool {
		r := math.Hypot(p.Y-ay, p.Z)
		return r <= a.radius(p.X, 0) && r > a.radius(p.X, steampunkAirshipShell)
	}}
	g.Use(Sand).FillVolume(shell)
	panel := func(p Vec3) (int, int) {
		theta := math.Atan2(p.Z, p.Y-ay)
		return int(math.Floor((theta + math.Pi) / (2 * math.Pi) * 16)), int(math.Floor((p.X - steampunkAirshipTail) / 5.5))
	}
	// Alternating canvas gores, with a few patched panels in leather and sand.
	g.Paint(Terracotta, Func{box, func(p Vec3) bool {
		gore, sec := panel(p)
		h := Hash01(gore, sec, 7)
		return (gore%2 == 0 && h > 0.1) || (gore%2 == 1 && h < 0.06)
	}})
	g.Paint(Wood, Func{box, func(p Vec3) bool {
		gore, sec := panel(p)
		return Hash01(gore, sec, 7) > 0.94
	}})
	// Brass hoops, a longitudinal belt on each flank and the ridge, nose and tail caps.
	g.Paint(Gold, Func{box, func(p Vec3) bool {
		x := p.X
		dy := p.Y - ay
		r := math.Hypot(dy, p.Z)
		hoop := math.Mod(x-steampunkAirshipTail+1000, 11) < 0.9 && x > steampunkAirshipTail+4 && x < steampunkAirshipNose-4
		theta := math.Atan2(p.Z, dy)
		belt := false
		for _, t := range []float64{-math.Pi / 2, math.Pi / 2, 0} {
			if math.Abs(theta-t)*r < 0.5 {
				belt = true
			}
		}
		return hoop || belt || x > steampunkAirshipNose-2.6 || x < steampunkAirshipTail+1.6
	}})
	// Lookout on the crown: a railed platform with a flag mast and pennant.
	lc := V(24, ay+a.radius(24, 0)-0.6, 0)
	g.Use(Wood).Fill(Cylinder{Center: lc.Plan(), Radius: 2.8, Y0: lc.Y, Y1: lc.Y + 0.6})
	g.Use(Gold).Fill(Cylinder{Center: lc.Plan(), Radius: 2.8, Inner: 2.2, Y0: lc.Y + 1.6, Y1: lc.Y + 2.1})
	for k := 0; k < 8; k++ {
		q := Polar(lc.Plan(), 2.5, float64(k)*math.Pi/4)
		g.Fill(Cylinder{Center: q, Radius: 0.35, Y0: lc.Y + 0.6, Y1: lc.Y + 1.8})
	}
	g.Use(Charcoal).Fill(Cylinder{Center: lc.Plan(), Radius: 0.45, Y0: lc.Y, Y1: lc.Y + 9})
	g.Use(Gold).Fill(Sphere(lc.Add(V(0, 9.4, 0)), 0.6))
	g.Use(Red).Fill(Func{Box{lc.Add(V(-6.5, 5.2, -0.9)), lc.Add(V(0, 8.8, 0.9))}, func(p Vec3) bool {
		t := (lc.X - p.X) / 6.5
		mid := lc.Y + 7.2 - 0.6*t
		return math.Abs(p.Y-mid) <= 1.6*(1-t)+0.3 && math.Abs(p.Z+0.4*math.Sin(t*5)) < 0.5
	}})
	// Brass keel walkway along the belly.
	g.Use(Gold).Fill(Func{Box{V(-30, ay-R-1, -1), V(32, ay-R+2, 1)}, func(p Vec3) bool {
		r := math.Hypot(p.Y-ay, p.Z)
		return math.Abs(p.Z) <= 0.8 && p.Y < ay && r <= a.radius(p.X, 0)+0.9 && r > a.radius(p.X, 0)-0.4
	}})
}

// fins are four cruciform tail fins in terracotta with brass edges and striped rudders.
func (a SteampunkAirship) fins(g *Grid) {
	ay := steampunkAirshipAxisY
	x0, x1, xr, span := -47.0, -28.0, -42.0, 19.5
	reach := func(x float64) float64 {
		lead := a.radius(x1, 0)
		if x > -37 {
			return lead + (span-lead)*(x1-x)/(x1+37)
		}
		return span
	}
	dirs := [][2]float64{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	box := Box{V(x0, ay-span, -span), V(x1, ay+span, span)}
	inFin := func(p Vec3) (along float64, ok bool) {
		if p.X < x0 || p.X > x1 {
			return 0, false
		}
		dy, dz := p.Y-ay, p.Z
		for _, d := range dirs {
			al := dy*d[0] + dz*d[1]
			across := -dy*d[1] + dz*d[0]
			if math.Abs(across) <= 0.6 && al > a.radius(p.X, steampunkAirshipShell) && al <= reach(p.X) {
				return al, true
			}
		}
		return 0, false
	}
	g.Use(Terracotta).Fill(Func{box, func(p Vec3) bool { _, ok := inFin(p); return ok }})
	g.Paint(Gold, Func{box, func(p Vec3) bool {
		al, ok := inFin(p)
		return ok && (al > reach(p.X)-0.8 || (p.X > xr-0.4 && p.X < xr+0.5) || p.X > x1-0.8)
	}})
	g.Paint(Red, Func{box, func(p Vec3) bool {
		al, ok := inFin(p)
		return ok && p.X < xr-0.4 && al <= reach(p.X)-0.8 && math.Mod(al, 2.4) < 1.2
	}})
	g.Paint(White, Func{box, func(p Vec3) bool {
		al, ok := inFin(p)
		return ok && p.X < xr-0.4 && al <= reach(p.X)-0.8 && math.Mod(al, 2.4) >= 1.2
	}})
	// Cut a hinge slot between fin and rudder so it reads as a separate control surface.
	g.Carve(Func{box, func(p Vec3) bool {
		al, ok := inFin(p)
		return ok && math.Abs(p.X-xr) < 0.35 && al > a.radius(p.X, 0)+1.2 && al < reach(p.X)-1.2
	}})
}

// gondola is a wooden ship hull with portholes, rails, a bridge, saloon and figurehead.
func (a SteampunkAirship) gondola(g *Grid) {
	deck, keel := steampunkAirshipDeck, steampunkAirshipKeel
	const stern, bow = -16.0, 29.0
	box := Box{V(stern, keel, -steampunkAirshipBeam), V(bow, deck, steampunkAirshipBeam)}
	outer := func(p Vec3, inset float64) bool { return math.Abs(p.Z) > a.hullWidth(p.X, p.Y, inset) }
	g.Use(Wood).FillVolume(Func{box, func(p Vec3) bool { return !outer(p, 0) }})
	g.Carve(Func{box, func(p Vec3) bool {
		return p.Y > keel+1.1 && p.X > stern+1.1 && !outer(p, 1.1)
	}})
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		return p.Y > keel+0.6 && p.X > stern+0.6 && !outer(p, 0.6)
	}})
	// Strakes: a brass rubbing strake under the gunwale, a dark keel band.
	g.Paint(Gold, Func{box, func(p Vec3) bool {
		return p.Y >= deck-1.0 && p.Y <= deck-0.6 && outer(p, 0.7)
	}})
	g.Paint(Charcoal, Func{box, func(p Vec3) bool { return p.Y < keel+1.2 && outer(p, 0.7) }})
	g.Paint(Gold, Block(V(stern, deck-1, -steampunkAirshipBeam), V(stern+0.5, deck-0.6, steampunkAirshipBeam)))
	// Two rows of portholes with brass rims; stern windows on the transom.
	for _, py := range []float64{keel + 2.9, keel + 4.6} {
		for x := stern + 3.0; x <= 23; x += 2.6 {
			xc, yc := x, py
			if yc > keel+4 {
				xc += 1.3
			}
			g.Paint(Gold, Func{Box{V(xc-1.2, yc-1.2, -steampunkAirshipBeam), V(xc+1.2, yc+1.2, steampunkAirshipBeam)}, func(p Vec3) bool {
				return math.Hypot(p.X-xc, p.Y-yc) <= 1.0 && outer(p, 0.7)
			}})
			g.Carve(Func{Box{V(xc-0.8, yc-0.8, -steampunkAirshipBeam), V(xc+0.8, yc+0.8, steampunkAirshipBeam)}, func(p Vec3) bool {
				return math.Hypot(p.X-xc, p.Y-yc) <= 0.62 && outer(p, 1.3)
			}})
		}
	}
	for _, z := range []float64{-2.4, 0, 2.4} {
		g.Paint(Gold, Block(V(stern, keel+3, z-0.9), V(stern+0.4, keel+5, z+0.9)))
		g.Carve(Block(V(stern, keel+3.4, z-0.5), V(stern+1.2, keel+4.6, z+0.5)))
	}
	// Deck.
	g.Use(Wood).Fill(Func{box, func(p Vec3) bool {
		return p.Y >= deck-0.6 && math.Abs(p.Z) <= a.hullWidth(p.X, deck, 0)
	}})
	// Bulwark rail with balusters and a brass cap, sweeping up fore and aft like a ship's
	// sheer, open at the boarding gate.
	sheer := func(x float64) float64 {
		t := (x - 6.5) / 22.5
		return deck + 1.8 + 1.8*t*t
	}
	railBox := Box{V(stern, deck, -steampunkAirshipBeam), V(bow, deck+3.8, steampunkAirshipBeam)}
	g.Fill(Func{railBox, func(p Vec3) bool {
		w := a.hullWidth(p.X, deck, 0)
		if w < 0 || math.Abs(p.Z) < w-0.8 && p.X > stern+0.8 || p.Y > sheer(p.X) {
			return false
		}
		if p.Z > 0 && p.X > 17 && p.X < 20.5 {
			return false // Boarding gate for the gangway.
		}
		return p.Y > sheer(p.X)-0.6 || p.Y < deck+0.5+(sheer(p.X)-deck-1.8) || math.Mod(p.X+100, 1.5) < 0.75
	}})
	g.Paint(Gold, Func{railBox, func(p Vec3) bool { return p.Y > sheer(p.X)-0.6 }})
	// Anchor hanging from the bow on its chain.
	g.Use(Charcoal).Fill(
		Capsule{V(25.5, deck-1.5, 2.4), V(26.8, keel-0.5, 3.0), 0.4},
		Capsule{V(26.8, keel-0.5, 3.0), V(26.8, keel-3.4, 3.0), 0.5},
		Capsule{V(26.8, keel-3.4, 3.0), V(25.4, keel-2.4, 3.0), 0.45},
		Capsule{V(26.8, keel-3.4, 3.0), V(28.2, keel-2.4, 3.0), 0.45},
		Capsule{V(26.8, keel-0.9, 2.0), V(26.8, keel-0.9, 4.0), 0.4},
	)
	// Bridge: glazed wheelhouse with a dark roof, brass trim and a ship's wheel.
	b0, b1 := 9.0, 16.0
	g.Use(Wood).Fill(Block(V(b0, deck, -3.2), V(b1, deck+4.2, 3.2)))
	g.Carve(Block(V(b0+0.8, deck, -2.4), V(b1-0.8, deck+4.2, 2.4)))
	g.Paint(Charcoal, Block(V(b0+0.4, deck, -2.8), V(b1-0.4, deck+4.2, 2.8)))
	g.Paint(Glass, Func{Box{V(b0, deck+1.6, -3.2), V(b1, deck+3.6, 3.2)}, func(p Vec3) bool {
		edge := p.X > b1-0.5 || p.X < b0+0.5 || math.Abs(p.Z) > 2.7
		mullion := math.Mod(p.X+100, 1.6) < 0.5 && math.Abs(p.Z) > 2.7 || math.Abs(p.Z) < 0.4
		return edge && !mullion
	}})
	g.Use(Charcoal).Fill(Block(V(b0-0.6, deck+4.2, -3.8), V(b1+0.6, deck+4.8, 3.8)))
	g.Paint(Gold, Block(V(b0-0.6, deck+4.2, -3.8), V(b1+0.6, deck+4.3, 3.8)))
	g.Use(Gold).Fill(Cylinder{Center: P(b0+3.5, 0), Radius: 1.0, Y0: deck + 4.8, Y1: deck + 5.6}, Sphere(V(b0+3.5, deck+6.1, 0), 0.6))
	g.Fill(Cylinder{Center: P(b1+1.6, 0), Radius: 0.4, Y0: deck, Y1: deck + 1.8}, Sphere(V(b1+1.6, deck+2.3, 0), 0.7))
	// Saloon aft: a raised quarterdeck house with sand walls and a glazed band.
	s0, s1 := stern+1.2, -2.0
	g.Use(Sand).Fill(Block(V(s0, deck, -3.6), V(s1, deck+3.8, 3.6)))
	g.Carve(Block(V(s0+0.8, deck, -2.8), V(s1-0.8, deck+3.8, 2.8)))
	g.Paint(Charcoal, Block(V(s0+0.4, deck, -3.2), V(s1-0.4, deck+3.8, 3.2)))
	g.Paint(Glass, Func{Box{V(s0, deck+1.4, -3.6), V(s1, deck+2.9, 3.6)}, func(p Vec3) bool {
		return (math.Abs(p.Z) > 3.1 && math.Mod(p.X+100, 2.2) > 0.7) || (p.X > s1-0.5 && math.Mod(p.Z+100, 2.2) > 0.7)
	}})
	g.Use(Wood).Fill(Block(V(s0-0.5, deck+3.8, -4.2), V(s1+0.5, deck+4.4, 4.2)))
	g.Paint(Gold, Block(V(s0-0.5, deck+3.8, -4.2), V(s1+0.5, deck+3.9, 4.2)))
	g.Use(Gold).Fill(Func{Box{V(s0-0.5, deck+4.4, -4.2), V(s1+0.5, deck+5.6, 4.2)}, func(p Vec3) bool {
		edge := p.X < s0 || p.X > s1 || math.Abs(p.Z) > 3.7
		return edge && (p.Y > deck+5 || math.Mod(p.X+p.Z+100, 1.6) < 0.7)
	}})
	// Brass boiler amidships.
	g.Use(Gold).Fill(Func{Box{V(1, deck, -2.2), V(7, deck+4, 2.2)}, func(p Vec3) bool {
		return math.Hypot(p.Y-(deck+2), p.Z) <= 2
	}})
	g.Paint(Charcoal, Func{Box{V(1, deck, -2.2), V(7, deck+4, 2.2)}, func(p Vec3) bool {
		return math.Mod(p.X+100, 2) < 0.5
	}})
	g.Use(Charcoal).Fill(Cylinder{Center: P(5.5, 0), Radius: 0.8, Y0: deck + 3.5, Y1: deck + 6.5})
	g.Fill(Capsule{V(4, deck+2, 1.8), V(4, deck+1.2, 4), 0.45}, Capsule{V(4, deck+2, -1.8), V(4, deck+1.2, -4), 0.45})
	// Figurehead: a brass winged figure under the bowsprit.
	g.Use(Gold).Fill(Capsule{V(26.4, 10.8, 0), V(29.6, 12.8, 0), 0.8}, Sphere(V(30.4, 13.6, 0), 0.9))
	g.Fill(Capsule{V(28.2, 12, 0), V(26.6, 14.2, 2.0), 0.5}, Capsule{V(28.2, 12, 0), V(26.6, 14.2, -2.0), 0.5})
	g.Use(Wood).Fill(Capsule{V(26, deck+0.5, 0), V(33.5, deck+2.6, 0), 0.55})
	g.Use(Charcoal).Fill(Capsule{V(33.2, deck+2.4, 0), V(29.5, keel+2.5, 0), 0.4})
	// Side paddle wheels under brass paddle boxes.
	for _, s := range []float64{-1, 1} {
		c := V(4, keel+3.6, s*6.3)
		wheel := Box{V(c.X-3.6, c.Y-3.6, c.Z-0.7), V(c.X+3.6, c.Y+3.6, c.Z+0.7)}
		g.Use(Wood).Fill(Func{wheel, func(p Vec3) bool {
			dx, dy := p.X-c.X, p.Y-c.Y
			r := math.Hypot(dx, dy)
			if r > 3.4 {
				return false
			}
			ang := math.Mod(math.Atan2(dy, dx)+2*math.Pi, math.Pi/4)
			paddle := math.Min(ang, math.Pi/4-ang)*r < 0.45 && r > 1.2
			return r > 2.7 || paddle
		}})
		g.Use(Gold).Fill(Capsule{V(c.X, c.Y, s*4.6), V(c.X, c.Y, c.Z+s*0.9), 0.75})
		g.Fill(Func{Box{V(c.X-4.6, c.Y, c.Z-1.2), V(c.X+4.6, c.Y+4.6, c.Z+1.2)}, func(p Vec3) bool {
			r := math.Hypot(p.X-c.X, p.Y-c.Y)
			side := math.Abs(p.Z-(c.Z+s*0.9)) < 0.4
			return r <= 4.5 && (r > 3.9 || side) && p.Y >= c.Y+0.3
		}})
		g.Paint(Red, Func{Box{V(c.X-4.6, c.Y, c.Z-1.2), V(c.X+4.6, c.Y+4.6, c.Z+1.2)}, func(p Vec3) bool {
			r := math.Hypot(p.X-c.X, p.Y-c.Y)
			return r < 3.2 && r > 2.2 && math.Abs(p.Z-(c.Z+s*0.9)) < 0.4
		}})
	}
}

// engines are four brass steam engines on outriggers between gondola and envelope, each
// with a smokestack, steam puffs and a pusher propeller.
func (a SteampunkAirship) engines(g *Grid) {
	deck := steampunkAirshipDeck
	const zc, py = 15.6, 21.0
	for i, px := range []float64{-5.0, 12.0} {
		for j, s := range []float64{-1, 1} {
			z := s * zc
			// Outrigger truss from the gondola and struts up to the envelope.
			g.Use(Charcoal).Fill(
				Capsule{V(px, deck+1.6, s*3.8), V(px, py-0.8, s*(zc-1.4)), 0.55},
				Capsule{V(px, 12.4, s*3.6), V(px, py-1.4, s*(zc-0.4)), 0.45},
				Capsule{V(px-2.4, py+1.2, z-s*0.9), V(px-2.4, 30.6, s*10.8), 0.5},
				Capsule{V(px+2.4, py+1.2, z-s*0.9), V(px+2.4, 30.6, s*10.8), 0.5},
			)
			// Pod.
			pod := Func{Box{V(px-3.2, py-1.8, z-1.8), V(px+3.2, py+1.8, z+1.8)}, func(p Vec3) bool {
				return math.Hypot(p.Y-py, p.Z-z) <= 1.75
			}}
			g.Use(Gold).FillVolume(pod)
			g.FillVolume(Ellipsoid{V(px+3.2, py, z), V(1.8, 1.6, 1.6)})
			g.Paint(Charcoal, Func{pod.Box, func(p Vec3) bool {
				return math.Abs(p.X-px+1.5) < 0.35 || math.Abs(p.X-px-1.5) < 0.35
			}})
			g.Use(Charcoal).Fill(Sphere(V(px+4.9, py, z), 0.6))
			// Smokestack with flared crown.
			sc := P(px, z)
			stackTop := 31.0 + 1.5*float64(i)
			g.Use(Charcoal).Fill(Cylinder{Center: sc, Radius: 0.95, Y0: py, Y1: stackTop})
			g.Fill(Frustum{Center: sc, R0: 0.95, R1: 1.6, Y0: stackTop - 1, Y1: stackTop + 0.6})
			g.Carve(Cylinder{Center: sc, Radius: 0.7, Y0: stackTop - 0.4, Y1: stackTop + 1})
			g.Paint(Gold, Cylinder{Center: sc, Radius: 1.0, Y0: (py + stackTop) / 2, Y1: (py+stackTop)/2 + 0.6})
			// Steam puffs drifting aft.
			k := float64(i*2 + j)
			g.Use(White).FillVolume(
				Sphere(V(sc.X-0.4, stackTop+2.0, z), 1.3),
				Sphere(V(sc.X-1.8, stackTop+3.6+0.3*k, z+s*0.4), 1.7),
				Sphere(V(sc.X-3.9, stackTop+4.8+0.2*k, z+s*0.2), 1.3),
			)
			// Pusher propeller: hub and four wooden blades with brass tips.
			hub := V(px-3.8, py, z)
			g.Use(Charcoal).Fill(Capsule{V(px-3.2, py, z), hub, 0.6}, Sphere(hub, 0.85))
			for b := 0; b < 4; b++ {
				ang := float64(b)*math.Pi/2 + math.Pi/4 + 0.3*float64(j)
				dir := V(0, math.Cos(ang), math.Sin(ang))
				tip := hub.Add(dir.Scale(5.2))
				g.Use(Wood).Fill(Capsule{hub.Add(dir.Scale(0.8)), tip, 0.6})
				g.FillVolume(Ellipsoid{hub.Add(dir.Scale(3.1)), V(0.6, 0.6+1.6*math.Abs(dir.Y), 0.6+1.6*math.Abs(dir.Z))})
				g.Paint(Gold, Sphere(tip, 1.1))
			}
		}
	}
}

// mirror flips the whole site across x = 0, so the hero view looks along the ship from
// its tail with the mast beyond the nose. Lattice parity is preserved.
func (SteampunkAirship) mirror(g *Grid) {
	cells := g.Cells()
	colors := make([]Color, len(cells))
	for i, c := range cells {
		colors[i], _ = g.ColorOf(c)
		g.Remove(c)
	}
	brush := g.Brush()
	for i, c := range cells {
		g.Use(colors[i]).Add(Cell{-c.X, c.Y, c.Z})
	}
	g.Use(brush)
}

// rigging hangs the gondola from the envelope and ties the ship to the mast and ground.
func (a SteampunkAirship) rigging(g *Grid) {
	deck, ay := steampunkAirshipDeck, steampunkAirshipAxisY
	g.Use(Charcoal)
	xs := []float64{-13, -5, 3, 11, 19}
	for _, s := range []float64{-1, 1} {
		for i, x := range xs {
			w := a.hullWidth(x, deck, 0)
			low := V(x, deck+1.6, s*(w-0.4))
			r := a.radius(x*1.05, 0)
			high := V(x*1.05, ay-math.Sqrt(r*r-36)+0.6, s*6)
			g.Fill(Capsule{low, high, 0.45})
			if i+1 < len(xs) {
				x2 := xs[i+1]
				r2 := a.radius(x2*1.05, 0)
				g.Fill(Capsule{low, V(x2*1.05, ay-math.Sqrt(r2*r2-36)+0.6, s*6), 0.4})
			}
		}
	}
	// Gangway from the mast landing to the gondola's boarding gate.
	m := steampunkAirshipMast
	gx0, gx1 := 18.0, m.X-a.mastHalf(deck)
	g.Use(Wood).Fill(Block(V(gx0, deck-0.6, 4.6), V(gx1, deck, 6.8)))
	g.Use(Gold)
	for _, z := range []float64{4.6, 6.8} {
		g.Fill(Capsule{V(gx0+2.5, deck+1.4, z), V(gx1, deck+1.4, z), 0.45})
		for x := gx0 + 2.5; x <= gx1; x += 2.2 {
			g.Fill(Capsule{V(x, deck, z), V(x, deck+1.4, z), 0.4})
		}
	}
	g.Use(Charcoal).Fill(Capsule{V(gx0+6, deck-0.6, 5.7), V(gx0+12, deck-3, 5.7), 0.4}, Capsule{V(gx0+12, deck-3, 5.7), V(gx1-1, deck-0.6, 5.7), 0.4})
	// Nose lines down to the mast's landing corners.
	for _, s := range []float64{-1, 1} {
		r := a.radius(34, 0)
		g.Fill(Capsule{V(34, ay-r*0.4, s*r*0.85), V(m.X-a.mastHalf(deck), deck+1, s*a.mastHalf(deck)), 0.42})
	}
}
