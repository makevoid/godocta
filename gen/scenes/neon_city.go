package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// NeonCity is a dense, generic sci-fi city block at night: a tight vertical cluster of
// mismatched megastructures on a stepped city plate. A stepped Charcoal tower, a Navy
// slab with a rooftop landing pad and a giant holographic billboard, a Slate cylinder
// with bulging pods and a spire, a hexagonal needle and a stack of offset boxes rise
// over lower tenement blocks with carved windows, balconies, AC units and pipes. Glass
// sky bridges link the towers at several heights, an elevated monorail loops through
// tunnel portals in three towers with a train on its south-east arc, and an elevated
// highway ramp climbs along the east edge. At street level a narrow canyon street of
// market stalls under striped awnings and lantern strings crosses the block, and a canal
// with a footbridge and boats runs to the front edge. Vertical sign boards stacked up the
// facades, lit window rows and neon trims supply the bright accents. Front is +Z.
type NeonCity struct{}

func init() { Register(NeonCity{}) }

func (NeonCity) Name() string  { return "neon_city" }
func (NeonCity) Title() string { return "Neon City" }

const (
	neonCityG     = 2.0  // Street level.
	neonCityX     = 36.0 // Plate half extent along X.
	neonCityZ     = 30.0 // Plate half extent along Z.
	neonCityRail  = 24.4 // Underside of the monorail beam.
	neonCityRailA = 27.0 // Monorail ellipse semi-axis along X.
	neonCityRailB = 17.0 // Monorail ellipse semi-axis along Z.
)

var neonCityRailC = P(-3, -4)

// neonCityLot is a building footprint and its roof height.
type neonCityLot struct {
	lo, hi Vec2
	top    float64
	tower  bool
}

type neonCitySite struct {
	g    *Grid
	lots []neonCityLot
}

// neonCityFace is one vertical facade of a box: o is its left corner, dir runs along the
// facade and n points outward. Facade-local blocks use u along, y up and t outward.
type neonCityFace struct {
	o, dir, n Vec2
	length    float64
}

func (f neonCityFace) block(u0, u1, y0, y1, t0, t1 float64) Cuboid {
	a := f.o.Add(f.dir.Scale(u0)).Add(f.n.Scale(t0))
	b := f.o.Add(f.dir.Scale(u1)).Add(f.n.Scale(t1))
	return Block(a.At(y0), b.At(y1))
}

// local returns the facade coordinates (u, t) of a point.
func (f neonCityFace) local(p Vec3) (float64, float64) {
	d := p.Plan().Sub(f.o)
	return d.Dot(f.dir), d.Dot(f.n)
}

// neonCityFaces lists a box's facades: south (+Z), east (+X), north (-Z), west (-X).
func neonCityFaces(lo, hi Vec2) [4]neonCityFace {
	w, d := hi.X-lo.X, hi.Y-lo.Y
	return [4]neonCityFace{
		{P(lo.X, hi.Y), P(1, 0), P(0, 1), w},
		{P(hi.X, hi.Y), P(0, -1), P(1, 0), d},
		{P(hi.X, lo.Y), P(-1, 0), P(0, -1), w},
		{P(lo.X, lo.Y), P(0, 1), P(-1, 0), d},
	}
}

// neonCitySkin selects the side skin of a box, within d of its vertical faces.
func neonCitySkin(lo, hi Vec2, y0, y1, d float64) Solid {
	return Func{Box{lo.At(y0), hi.At(y1)}, func(p Vec3) bool {
		return p.X <= lo.X+d || p.X >= hi.X-d || p.Z <= lo.Y+d || p.Z >= hi.Y-d
	}}
}

func (NeonCity) Build(g *Grid) Order {
	s := &neonCitySite{g: g}
	s.plate()
	s.canal()
	s.steppedTower()
	s.slabTower()
	s.podTower()
	s.hexTower()
	s.stackTower()
	s.tenements()
	s.monorail()
	s.bridges()
	s.highway()
	s.market()
	s.cables()
	s.extras()
	s.hollow(2)
	return Outward(P(-3, -6))
}

// recolor paints occupied cells inside box with the color f picks, trying each listed color.
func (s *neonCitySite) recolor(b Box, colors []Color, f func(Vec3) (Color, bool)) {
	for _, k := range colors {
		k := k
		s.g.Paint(k, Func{b, func(p Vec3) bool { c, ok := f(p); return ok && c == k }})
	}
}

var neonCityNeon = []Color{Red, Gold, Glass, White, Water, Navy, Charcoal, Slate}

// lit picks a lit window color from a hash, or false for a dark pane.
func neonCityLit(r float64) (Color, bool) {
	switch {
	case r < 0.36:
		return Gold, true
	case r < 0.5:
		return Glass, true
	case r < 0.6:
		return Water, true
	case r < 0.65:
		return White, true
	case r < 0.69:
		return Red, true
	}
	return 0, false
}

func (s *neonCitySite) lot(lo, hi Vec2, top float64, tower bool) {
	s.lots = append(s.lots, neonCityLot{lo, hi, top, tower})
	s.g.Paint(Slate, Block(lo.Sub(P(1.2, 1.2)).At(1.75), hi.Add(P(1.2, 1.2)).At(2.1)))
}

// shell fills a hollow box with walls and roof wall thick; inner faces are Charcoal.
func (s *neonCitySite) shell(lo, hi Vec2, y0, y1, wall float64, c Color) {
	g := s.g
	g.Use(c).Fill(Block(lo.At(y0), hi.At(y1)))
	if hi.X-lo.X > 2*wall+1 && hi.Y-lo.Y > 2*wall+1 && y1-y0 > 2*wall+1 {
		in := P(wall, wall)
		inner := Block(lo.Add(in).At(y0+0.5), hi.Sub(in).At(y1-wall))
		g.Carve(inner)
		s.darken(Box{lo.At(y0), hi.At(y1)}, inner)
	}
}

// darken paints Charcoal every cell in b that faces an empty cell inside interior, so
// openings show a dark inside.
func (s *neonCitySite) darken(b Box, interior Solid) {
	g := s.g
	span := g.Span()
	g.Paint(Charcoal, Func{b, func(p Vec3) bool {
		c := Nearest(p.Scale(1 / span))
		if g.World(c).Sub(p).Length() > 1e-6 {
			return false
		}
		for _, n := range Neighbors {
			q := c.Add(n)
			if !g.Has(q) && interior.Contains(g.World(q)) {
				return true
			}
		}
		return false
	}})
}

// litWindows paints rows of lit panes on the outer skin of a box tower.
func (s *neonCitySite) litWindows(lo, hi Vec2, y0, y1, storey float64, seed int) {
	for fi, f := range neonCityFaces(lo, hi) {
		for fl := 0; y0+float64(fl)*storey+2.4 <= y1; fl++ {
			wy := y0 + float64(fl)*storey + 1
			for k := 0; 1.2+float64(k)*2+1.4 <= f.length-1; k++ {
				u := 1.2 + float64(k)*2
				if c, ok := neonCityLit(Hash01(seed, fi, fl, k)); ok {
					s.g.Paint(c, f.block(u, u+1.4, wy, wy+1.3, -0.45, 0.1))
				}
			}
		}
	}
}

// radialWindows paints lit panes on the skin of a round or polygonal tower, in rows
// of the given storey and angular sectors around center.
func (s *neonCitySite) radialWindows(center Vec2, b Box, skin func(Vec3) bool, y0, storey float64, sectors, seed int) {
	s.recolor(b, []Color{Gold, Glass, Water, White, Red}, func(p Vec3) (Color, bool) {
		if !skin(p) {
			return 0, false
		}
		v := math.Mod(p.Y-y0, storey)
		if v < 1 || v > 2.3 {
			return 0, false
		}
		a := p.Plan().Sub(center).Angle()/(2*math.Pi) + 0.5
		k := a * float64(sectors)
		if fr := k - math.Floor(k); fr < 0.18 || fr > 0.82 {
			return 0, false
		}
		return neonCityLit(Hash01(seed, int(math.Floor(k)), int(math.Floor((p.Y-y0)/storey))))
	})
}

// vsign is a vertical sign board standing out from a facade at u, from y0 to y1: a dark
// board with a bright frame and columns of glyphs.
func (s *neonCitySite) vsign(f neonCityFace, u, y0, y1, depth float64, seed int) {
	g := s.g
	themes := [][2]Color{{Red, Gold}, {Water, White}, {Gold, Red}, {Glass, White}, {White, Red}, {Red, White}, {Gold, Water}}
	th := themes[int(Hash01(seed, 1)*float64(len(themes)))]
	board := f.block(u-0.5, u+0.5, y0, y1, 0, depth)
	g.Use(Navy).Fill(board)
	g.Fill(f.block(u-0.6, u+0.6, y0+0.4, y0+0.9, -0.2, 0.6)) // Bracket into the wall.
	s.recolor(Box(board), []Color{th[0], th[1]}, func(p Vec3) (Color, bool) {
		_, t := f.local(p)
		if p.Y < y0+0.5 || p.Y > y1-0.5 || t > depth-0.5 {
			return th[0], true
		}
		gy := (p.Y - y0 - 0.9) / 1.7
		row := int(math.Floor(gy))
		if gy-float64(row) > 0.78 || t < 0.4 {
			return 0, false
		}
		ia := int(math.Floor((p.Y - y0) / 0.42))
		ib := int(math.Floor(t / 0.42))
		if Hash01(seed, row, ia, ib) < 0.62 {
			return th[1], true
		}
		return 0, false
	})
}

// billboard is a flat screen on a facade with a frame and one of three patterns.
func (s *neonCitySite) billboard(f neonCityFace, u0, u1, y0, y1 float64, seed int) {
	g := s.g
	board := f.block(u0, u1, y0, y1, 0, 0.8)
	g.Use(Navy).Fill(board)
	w, h := u1-u0, y1-y0
	cu, cv := w/2, h/2
	kind := seed % 3
	s.recolor(Box(board).Grow(0.2), []Color{White, Glass, Water, Red, Gold}, func(p Vec3) (Color, bool) {
		u, _ := f.local(p)
		u -= u0
		v := p.Y - y0
		if u < 0.6 || u > w-0.6 || v < 0.6 || v > h-0.6 {
			return White, true
		}
		switch kind {
		case 0: // Scanlines behind a ringed disk logo.
			r := math.Hypot(u-cu, v-cv)
			R := math.Min(w, h) * 0.32
			if r < R*0.55 {
				return Red, true
			}
			if r < R*0.8 {
				return Gold, true
			}
			if int(math.Floor(v/0.8))%2 == 0 {
				return Glass, true
			}
			return Water, true
		case 1: // Diagonal stripes.
			k := int(math.Floor((u + v) / 1.6))
			switch ((k % 3) + 3) % 3 {
			case 0:
				return Red, true
			case 1:
				return Gold, true
			}
			return 0, false
		}
		// Big glyphs in a grid.
		gu, gv := int(math.Floor((u-1)/2.6)), int(math.Floor((v-1)/2.6))
		fu, fv := math.Mod(u-1, 2.6), math.Mod(v-1, 2.6)
		if fu > 2 || fv > 2 {
			return 0, false
		}
		if Hash01(seed, gu, gv, int(fu/0.5), int(fv/0.5)) < 0.55 {
			if Hash01(seed, gu, gv) < 0.5 {
				return Water, true
			}
			return Gold, true
		}
		return 0, false
	})
}

// plate is the stepped city plinth with asphalt, lane markings and a neon rim.
func (s *neonCitySite) plate() {
	g := s.g
	g.Use(DarkStone)
	for i := 0; i < 4; i++ {
		f := float64(i)
		g.Fill(Block(V(-neonCityX+f, f*0.5, -neonCityZ+f), V(neonCityX-f, f*0.5+0.5, neonCityZ-f)))
	}
	for i := 0; i < 3; i++ {
		f := float64(i) + 1.6
		g.Carve(Block(V(-neonCityX+f, f*0.5-1, -neonCityZ+f), V(neonCityX-f, float64(i)*0.5+0.5, neonCityZ-f)))
	}
	g.Carve(Block(V(-31, -1, -25), V(31, 1.3, 25)))
	top := Block(V(-33, 1.75, -27), V(33, 2.1, 27))
	g.Paint(Charcoal, top)
	g.Paint(Water, Subtract(top, Block(V(-32.4, 1, -26.4), V(32.4, 3, 26.4))))
	g.Paint(Red, Subtract(Block(V(-35, 0.2, -29), V(35, 0.6, 29)), Block(V(-34.4, 0, -28.4), V(34.4, 1, 28.4))))
	// Canyon street (along X) and cross street (along Z) markings.
	for x := -32.0; x < 32; x += 3 {
		g.Paint(Gold, Block(V(x, 1.75, 6), V(x+1.5, 2.1, 7)))
	}
	for z := -26.0; z < 26; z += 3 {
		g.Paint(Gold, Block(V(8, 1.75, z), V(9, 2.1, z+1.5)))
	}
	for x := 7.0; x < 10; x += 1.2 {
		g.Paint(White, Block(V(x, 1.75, 2.2), V(x+0.6, 2.1, 3.4)), Block(V(x, 1.75, 9.6), V(x+0.6, 2.1, 10.8)))
	}
}

// canal runs from the canyon street to the front edge, with quays, a footbridge and boats.
func (s *neonCitySite) canal() {
	g := s.g
	g.Carve(Block(V(-8, 0.9, 9.5), V(-3, 4, 31)))
	g.Use(DarkStone).Fill(Block(V(-9, 0.4, 9.5), V(-8, 2, 28.5)), Block(V(-3, 0.4, 9.5), V(-2, 2, 28.5)),
		Block(V(-9, 0.4, 8.6), V(-2, 2, 9.5)))
	g.Use(Water).Fill(Block(V(-8, 0.4, 9.5), V(-3, 0.85, 30)))
	g.Paint(Water, Block(V(-9, 1.7, 9), V(-2, 2.1, 9.5)))
	// Arched footbridge with railings.
	h := func(x float64) float64 { u := (x + 5.5) / 4.2; return neonCityG + 1.3*(1-u*u) }
	g.Use(Red).Fill(Func{Box{V(-9.7, 1.5, 15.8), V(-1.3, 4.5, 18.2)}, func(p Vec3) bool {
		y := h(p.X)
		return p.Y <= y && p.Y >= y-0.6
	}})
	g.Use(White).Fill(Func{Box{V(-9.7, 1.5, 15.8), V(-1.3, 5.5, 18.2)}, func(p Vec3) bool {
		y := h(p.X)
		return p.Y > y && p.Y <= y+0.9 && (p.Z < 16.4 || p.Z > 17.6)
	}})
	// Two small boats with canopies.
	for i, z := range []float64{21.5, 26} {
		at := P(-5.5+float64(i)*0.6-0.3, z)
		hull := Intersect(Ellipsoid{at.At(0.9), V(1.3, 0.9, 2.2)}, Block(V(at.X-2, 0.85, z-3), V(at.X+2, 1.6, z+3)))
		g.Use([]Color{Navy, White}[i]).FillVolume(hull)
		g.Use([]Color{Red, Gold}[i]).Fill(Block(V(at.X-0.8, 1.6, z-0.8), V(at.X+0.8, 2.6, z+1)))
		g.Carve(Block(V(at.X-0.4, 1.6, z-0.8), V(at.X+0.4, 2.1, z+1)))
	}
}

// steppedTower is the tallest tower: four setback tiers with red cornices and masts.
func (s *neonCitySite) steppedTower() {
	g := s.g
	lo, hi := P(-31, -25), P(-15, -9)
	tiers := []struct{ h, in float64 }{{28, 0}, {20, 1.5}, {18, 3}, {8, 4.5}}
	y := neonCityG
	for i, t := range tiers {
		a, b := lo.Add(P(t.in, t.in)), hi.Sub(P(t.in, t.in))
		s.shell(a, b, y, y+t.h, 1.0, Charcoal)
		s.litWindows(a, b, y, y+t.h-1, 3, 100+i)
		g.Paint([]Color{Red, Water, Red, Gold}[i], neonCitySkin(a, b, y+t.h-0.8, y+t.h, 0.6))
		y += t.h
	}
	s.lot(lo, hi, y, true)
	// Masts, beacons and a rooftop sign.
	g.Use(Charcoal).Fill(Capsule{V(-24, y, -18), V(-24, y+17, -18), 0.6}, Capsule{V(-21.5, y, -15.5), V(-21.5, y+10, -15.5), 0.55})
	g.Use(Red).Fill(Sphere(V(-24, y+17.4, -18), 0.8), Sphere(V(-21.5, y+10.4, -15.5), 0.7))
	g.Use(Gold).Fill(Cylinder{Center: P(-24, -18), Radius: 1.2, Y0: y + 8, Y1: y + 8.8})
	roof := neonCityFace{P(-26, -14.6), P(1, 0), P(0, 1), 6}
	g.Use(Charcoal).Fill(Block(V(-25.5, y, -15), V(-25, y+1.5, -14.6)), Block(V(-20.5, y, -15), V(-20, y+1.5, -14.6)))
	s.billboard(roof, 0, 6, y+1.5, y+6, 7)
	// Vertical signs on the east face and upper tiers.
	east := neonCityFaces(lo, hi)[1]
	s.vsign(east, 3, 5, 25, 2.6, 11)
	s.vsign(east, 12, 8, 26, 2.4, 12)
	t2 := neonCityFaces(lo.Add(P(1.5, 1.5)), hi.Sub(P(1.5, 1.5)))
	s.vsign(t2[0], 10, 32, 47, 2.6, 13)
	s.vsign(t2[1], 9, 33, 46, 2.4, 14)
	t3 := neonCityFaces(lo.Add(P(3, 3)), hi.Sub(P(3, 3)))
	s.vsign(t3[0], 3, 52, 66, 2.4, 15)
	s.billboard(t3[1], 1, 9, 54, 63, 16)
}

// slabTower is a thin Navy slab with a holographic billboard and a rooftop landing pad.
func (s *neonCitySite) slabTower() {
	g := s.g
	lo, hi := P(-10, -25), P(6, -17)
	top := neonCityG + 60
	s.shell(lo, hi, neonCityG, top, 1.0, Navy)
	s.litWindows(lo, hi, neonCityG, top-1, 2.5, 200)
	for y := neonCityG + 10; y < top; y += 12.5 {
		g.Paint(Water, neonCitySkin(lo, hi, y, y+0.6, 0.6))
	}
	s.lot(lo, hi, top, true)
	f := neonCityFaces(lo, hi)
	s.billboard(f[0], 1.5, 14.5, 46, 58.5, 21)
	s.vsign(f[1], 6, 8, 34, 2.6, 22)
	s.vsign(f[0], 14.2, 6, 40, 2.6, 24)
	s.vsign(f[3], 4, 12, 30, 2.4, 23)
	// Landing pad on a flared stalk.
	c := P(-2, -21)
	g.Use(Navy).FillVolume(Frustum{c, 3, 6.5, top - 2, top})
	g.Use(Charcoal).Fill(Cylinder{Center: c, Radius: 6.8, Y0: top, Y1: top + 0.8})
	g.Paint(White, Cylinder{Center: c, Radius: 5.3, Inner: 4.3, Y0: top + 0.4, Y1: top + 1})
	g.Paint(Red, Cylinder{Center: c, Radius: 1.6, Y0: top + 0.4, Y1: top + 1})
	for k := 0; k < 10; k++ {
		g.Use(Gold).Fill(Sphere(Polar(c, 6.3, float64(k)*math.Pi/5).At(top+1.2), 0.5))
	}
	// A parked flying car.
	car := c.Add(P(-1.5, 2.2))
	g.Use(White).FillVolume(Ellipsoid{car.At(top + 1.8), V(2.3, 0.9, 1.2)})
	g.Use(Glass).FillVolume(Ellipsoid{car.Add(P(0.4, 0)).At(top + 2.5), V(1.2, 0.7, 0.9)})
	g.Use(Red).Fill(Sphere(car.Add(P(-2.3, 0)).At(top+1.9), 0.5))
	g.Use(Charcoal).Fill(Capsule{V(5, top, -24), V(5, top+9, -24), 0.55})
	g.Use(Red).Fill(Sphere(V(5, top+9.4, -24), 0.7))
}

// podTower is a Slate cylinder with three bulging pods, neon equators and a spire.
func (s *neonCitySite) podTower() {
	g := s.g
	c := P(18, -16)
	top := neonCityG + 78
	pods := []struct{ y, r, h float64 }{{18, 8, 3}, {44, 7.6, 3}, {64, 7.2, 2.6}}
	shape := func(inset float64) Solid {
		u := UnionSolid{Cylinder{Center: c, Radius: 6 - inset, Y0: neonCityG + inset*0.5, Y1: top - inset}}
		for _, p := range pods {
			u = append(u, Ellipsoid{c.At(p.y), V(p.r-inset, p.h-inset, p.r-inset)})
		}
		return u
	}
	g.Use(Slate).FillVolume(Subtract(shape(0), shape(0.9)))
	s.darken(Box{V(c.X-8.5, neonCityG, c.Y-8.5), V(c.X+8.5, top, c.Y+8.5)}, shape(0.9))
	outer := func(y float64) float64 {
		r := 6.0
		for _, p := range pods {
			if d := (y - p.y) / p.h; math.Abs(d) < 1 {
				r = math.Max(r, p.r*math.Sqrt(1-d*d))
			}
		}
		return r
	}
	b := Box{V(c.X-8.5, neonCityG, c.Y-8.5), V(c.X+8.5, top, c.Y+8.5)}
	skin := func(p Vec3) bool { return p.Plan().Sub(c).Length() >= outer(p.Y)-0.55 }
	s.radialWindows(c, b, skin, neonCityG, 3, 36, 300)
	for i, p := range pods {
		p := p
		g.Paint([]Color{Water, Red, Water}[i], Func{b, func(q Vec3) bool { return math.Abs(q.Y-p.y) <= 0.6 && skin(q) }})
		g.Paint(White, Func{b, func(q Vec3) bool { return math.Abs(math.Abs(q.Y-p.y)-1.6) <= 0.3 && skin(q) }})
	}
	s.lot(c.Sub(P(6, 6)), c.Add(P(6, 6)), top, true)
	// Dome cap and spire.
	g.Use(Navy).FillVolume(Intersect(Ellipsoid{c.At(top), V(6, 2.8, 6)}, Block(V(c.X-7, top, c.Y-7), V(c.X+7, top+3, c.Y+7))))
	g.Use(White).Fill(Capsule{c.At(top), c.At(top + 16), 0.7})
	g.Use(Red).Fill(Sphere(c.At(top+16.4), 0.9), Cylinder{Center: c, Radius: 1.5, Y0: top + 8, Y1: top + 8.8})
	g.Use(Gold).Fill(Cylinder{Center: c, Radius: 1.3, Y0: top + 12, Y1: top + 12.6})
	// A curved holographic screen wrapping the south-east side between the pods.
	s.recolor(b, []Color{White, Glass, Water, Red, Gold}, func(p Vec3) (Color, bool) {
		if p.Y < 30.5 || p.Y > 40.5 || !skin(p) {
			return 0, false
		}
		a := p.Plan().Sub(c).Angle()
		if a < 0.15 || a > 1.45 {
			return 0, false
		}
		if p.Y < 31.1 || p.Y > 39.9 || a < 0.25 || a > 1.35 {
			return White, true
		}
		if p.Y > 33.5 && p.Y < 37.5 {
			k := int(math.Floor((a - 0.25) / 0.16))
			if math.Mod(a-0.25, 0.16) < 0.11 && Hash01(k, int((p.Y-33.5)/0.5)) < 0.6 {
				return []Color{Red, Gold}[k%2], true
			}
		}
		if int(math.Floor((p.Y-30.5)/0.8))%2 == 0 {
			return Glass, true
		}
		return Water, true
	})
	// Signs standing out radially to the east and south.
	s.vsign(neonCityFace{P(23.6, -16), P(0, 1), P(1, 0), 1}, 0, 23, 40, 2.8, 31)
	s.vsign(neonCityFace{P(18, -10.4), P(1, 0), P(0, 1), 1}, 0, 50, 62, 2.6, 32)
}

// hexTower is a hexagonal DarkStone needle with neon corner strips.
func (s *neonCitySite) hexTower() {
	g := s.g
	c := P(-1, -8)
	top := neonCityG + 42
	hex := func(r float64) []Vec2 {
		var v []Vec2
		for k := 0; k < 6; k++ {
			v = append(v, Polar(c, r, float64(k)*math.Pi/3))
		}
		return v
	}
	g.Use(DarkStone).Fill(Prism{hex(4.6), neonCityG, top})
	g.Carve(Prism{hex(3.75), neonCityG + 0.5, top - 1})
	s.darken(Box{V(c.X-5, neonCityG, c.Y-5), V(c.X+5, top, c.Y+5)}, Prism{hex(3.75), neonCityG + 0.5, top - 1})
	b := Box{V(c.X-5, neonCityG, c.Y-5), V(c.X+5, top, c.Y+5)}
	skinPoly := hex(4.1)
	skin := func(p Vec3) bool { return !InsidePolygon(skinPoly, p.Plan()) }
	s.radialWindows(c, b, skin, neonCityG, 3, 24, 400)
	g.Paint(Water, Func{b, func(p Vec3) bool {
		d := p.Plan().Sub(c)
		a := math.Mod(d.Angle()+2*math.Pi, math.Pi/3)
		return skin(p) && (a < 0.11 || a > math.Pi/3-0.11)
	}})
	for y := neonCityG + 12; y < top; y += 12 {
		g.Paint(Red, Prism{hex(5), y, y + 0.6})
	}
	s.lot(c.Sub(P(4.6, 4)), c.Add(P(4.6, 4)), top, true)
	g.Use(DarkStone).FillVolume(Frustum{c, 3.8, 1.2, top, top + 5})
	g.Paint(Water, Frustum{c, 3.8, 1.2, top + 2, top + 2.6})
	g.Use(Glass).Fill(Capsule{c.At(top + 5), c.At(top + 13), 0.6})
	g.Use(Red).Fill(Sphere(c.At(top+13.4), 0.7))
}

// stackTower is a stack of offset boxes with colored trims and big signs.
func (s *neonCitySite) stackTower() {
	g := s.g
	G := neonCityG
	boxes := []struct {
		lo, hi  Vec2
		y0, y1  float64
		c, trim Color
	}{
		{P(-31, 12), P(-18, 25), G, G + 12, DarkStone, Gold},
		{P(-30, 11.5), P(-18, 22), G + 12, G + 22, Slate, Red},
		{P(-32, 11), P(-20, 21), G + 22, G + 32, Charcoal, Water},
		{P(-28, 14), P(-17, 24), G + 32, G + 42, Navy, Gold},
		{P(-30, 15), P(-22, 22), G + 42, G + 48, DarkStone, Red},
	}
	for i, b := range boxes {
		s.shell(b.lo, b.hi, b.y0, b.y1, 1.0, b.c)
		s.litWindows(b.lo, b.hi, b.y0, b.y1-1, 3, 500+i)
		g.Paint(b.trim, neonCitySkin(b.lo, b.hi, b.y0, b.y0+0.8, 0.6))
	}
	s.lot(P(-32, 11), P(-17, 25), G+48, true)
	f0 := neonCityFaces(boxes[0].lo, boxes[0].hi)
	s.vsign(f0[0], 3, G+1.5, G+11, 2.2, 51)
	s.vsign(f0[0], 9.5, G+3, G+11.5, 2.2, 52)
	f1 := neonCityFaces(boxes[1].lo, boxes[1].hi)
	s.vsign(f1[1], 2, G+13, G+21.5, 1.8, 53)
	f3 := neonCityFaces(boxes[3].lo, boxes[3].hi)
	s.billboard(f3[0], 1, 10, G+33, G+41, 54)
	f2 := neonCityFaces(boxes[2].lo, boxes[2].hi)
	s.vsign(f2[0], 6, G+23, G+31.5, 2.4, 55)
	// Roof: antenna array and a dish.
	y := G + 48
	for k := 0; k < 3; k++ {
		x := -28.5 + float64(k)*2.5
		g.Use(Charcoal).Fill(Capsule{V(x, y, 17), V(x, y+4+float64(k)*2, 17), 0.5})
		g.Use(Red).Fill(Sphere(V(x, y+4.3+float64(k)*2, 17), 0.6))
	}
	g.Use(White).FillVolume(Intersect(Sphere(V(-25, y+5.5, 20), 3), Subtract(Block(V(-29, y+1, 16), V(-21, y+4, 24)), Sphere(V(-25, y+5.5, 20), 2.3))))
	g.Use(Slate).Fill(Cylinder{Center: P(-25, 20), Radius: 0.6, Y0: y, Y1: y + 3})
}

// tenements are the lower stacked blocks around the towers.
func (s *neonCitySite) tenements() {
	G := neonCityG
	type ten struct {
		lo, hi Vec2
		h      float64
		c      Color
		roof   int // 0 plain, 1 water tanks, 2 billboard, 3 fan
	}
	list := []ten{
		{P(-31, -6.5), P(-21, 2.5), 18, DarkStone, 1},
		{P(-20, -6.5), P(-12, 2.5), 22, Slate, 2},
		{P(-11, -3), P(-6.5, 2.5), 13, Navy, 0},
		{P(-5.5, -2.5), P(5.5, 2.5), 10, Charcoal, 1},
		{P(11.5, -6), P(26, 2.5), 15, DarkStone, 2},
		{P(-16, 10.5), P(-10, 26), 16, Slate, 1},
		{P(-1.5, 10.5), P(6, 18), 20, DarkStone, 2},
		{P(-1.5, 19), P(6, 26), 11, Charcoal, 1},
		{P(11.5, 10.5), P(26, 25.5), 12, Navy, 3},
		{P(11.5, 10.5), P(17, 16), 18, Slate, 1},
	}
	for i, t := range list {
		s.tenement(t.lo, t.hi, G, G+t.h, t.c, 600+i*10)
		s.roof(t.lo, t.hi, G+t.h, t.roof, 600+i*10, t.c)
	}
	// Vertical signs over the canyon street on both sides.
	signs := []struct {
		lo, hi Vec2
		face   int
		u, h   float64
	}{
		{list[0].lo, list[0].hi, 0, 2.5, 11}, {list[0].lo, list[0].hi, 0, 7.5, 8},
		{list[1].lo, list[1].hi, 0, 2, 14}, {list[1].lo, list[1].hi, 0, 6, 9},
		{list[2].lo, list[2].hi, 0, 2.2, 7},
		{list[3].lo, list[3].hi, 0, 3, 6}, {list[3].lo, list[3].hi, 0, 8.5, 5},
		{list[4].lo, list[4].hi, 0, 3, 10}, {list[4].lo, list[4].hi, 0, 8, 8}, {list[4].lo, list[4].hi, 0, 12.5, 9},
		{list[5].lo, list[5].hi, 2, 3, 10},
		{list[6].lo, list[6].hi, 2, 2, 13}, {list[6].lo, list[6].hi, 2, 6, 9},
		{list[9].lo, list[9].hi, 2, 2.5, 11}, {list[8].lo, list[8].hi, 2, 2.5, 7},
		{list[8].lo, list[8].hi, 0, 3, 7}, {list[8].lo, list[8].hi, 0, 9, 6},
		{list[8].lo, list[8].hi, 1, 4, 7}, {list[8].lo, list[8].hi, 1, 11, 6},
		{list[7].lo, list[7].hi, 0, 4, 6}, {list[5].lo, list[5].hi, 0, 3, 9},
		{list[6].lo, list[6].hi, 1, 3, 12}, {list[4].lo, list[4].hi, 1, 3, 9},
	}
	for i, sg := range signs {
		f := neonCityFaces(sg.lo, sg.hi)[sg.face]
		s.vsign(f, sg.u, G+4, G+4+sg.h, 2, 700+i)
	}
}

// tenement is a hollow block with shopfronts, carved and lit windows, balconies, AC
// units and pipes.
func (s *neonCitySite) tenement(lo, hi Vec2, y0, y1 float64, c Color, seed int) {
	g := s.g
	s.shell(lo, hi, y0, y1, 1.0, c)
	s.lot(lo, hi, y1, false)
	for fi, f := range neonCityFaces(lo, hi) {
		// Shopfronts with neon fascia bands.
		for k := 0; 0.6+float64(k)*3+2.4 <= f.length-0.6; k++ {
			u := 0.6 + float64(k)*3
			if Hash01(seed, fi, k, 1) < 0.75 {
				g.Carve(f.block(u+0.3, u+2.4, y0+0.1, y0+2.2, -1.4, 0.2))
			}
			col := []Color{Red, Gold, Water, Glass, White}[int(Hash01(seed, fi, k, 2)*5)]
			g.Paint(col, f.block(u, u+2.7, y0+2.4, y0+3.1, -0.45, 0.1))
		}
		// Upper storeys.
		for fl := 1; y0+float64(fl)*3+2.6 <= y1; fl++ {
			wy := y0 + float64(fl)*3 + 0.9
			for k := 0; 1+float64(k)*2.2+1.2 <= f.length-1; k++ {
				u := 1 + float64(k)*2.2
				r := Hash01(seed, fi, fl, k)
				switch {
				case r < 0.42:
					g.Carve(f.block(u, u+1.2, wy, wy+1.4, -1.4, 0.2))
					if Hash01(seed, fi, fl, k, 3) < 0.3 {
						g.Use(Slate).Fill(f.block(u-0.4, u+1.6, wy-0.6, wy-0.1, 0, 1.2))
						g.Use([]Color{White, Red, Gold}[int(Hash01(seed, fi, fl, k, 4)*3)]).Fill(f.block(u-0.4, u+1.6, wy-0.1, wy+0.6, 0.65, 1.2))
					}
				case r < 0.72:
					if lc, ok := neonCityLit(Hash01(seed, fi, fl, k, 5) * 0.69); ok {
						g.Paint(lc, f.block(u, u+1.2, wy, wy+1.4, -0.45, 0.1))
					}
				}
				if Hash01(seed, fi, fl, k, 6) < 0.18 {
					g.Use(White).Fill(f.block(u+1.3, u+2.1, wy-0.2, wy+0.6, 0, 0.8))
				}
			}
		}
		// A drainpipe on some facades.
		if Hash01(seed, fi, 9) < 0.6 && f.length > 6 {
			u := 0.5 + math.Floor(Hash01(seed, fi, 10)*(f.length-2))
			a := f.o.Add(f.dir.Scale(u)).Add(f.n.Scale(0.45))
			g.Use([]Color{Slate, DarkStone, Charcoal}[int(Hash01(seed, fi, 11)*3)]).Fill(Capsule{a.At(y0 + 3.2), a.At(y1 - 0.3), 0.5})
		}
	}
}

// roof adds a parapet and rooftop furniture.
func (s *neonCitySite) roof(lo, hi Vec2, y float64, kind, seed int, col Color) {
	g := s.g
	g.Use(col).Fill(Subtract(Block(lo.At(y), hi.At(y+0.8)), Block(lo.Add(P(0.7, 0.7)).At(y), hi.Sub(P(0.7, 0.7)).At(y+1))))
	trim := []Color{Gold, Red, Water, White, Gold}[int(Hash01(seed, 31)*5)]
	g.Paint(trim, Subtract(Block(lo.At(y+0.4), hi.At(y+0.8)), Block(lo.Add(P(0.5, 0.5)).At(y), hi.Sub(P(0.5, 0.5)).At(y+1))))
	c := lo.Add(hi).Scale(0.5)
	w, d := hi.X-lo.X, hi.Y-lo.Y
	// A stair shack and AC boxes.
	sh := c.Add(P(-w*0.25, -d*0.2))
	g.Use(DarkStone).Fill(BlockAt(sh, 2.4, 2, y, y+2.4))
	g.Use(White).Fill(BlockAt(c.Add(P(w*0.2, -d*0.25)), 1.4, 1, y, y+1), BlockAt(c.Add(P(w*0.3, -d*0.25)), 1.4, 1, y, y+1))
	switch kind {
	case 1:
		for k := 0; k < 2 && (k == 0 || w > 8); k++ {
			t := c.Add(P(w*0.2-float64(k)*3.4, d*0.18))
			g.Use(Charcoal).Fill(Cylinder{Center: t, Radius: 1.25, Inner: 0.6, Y0: y, Y1: y + 1.2})
			g.Use(Slate).Fill(Cylinder{Center: t, Radius: 1.35, Y0: y + 1.2, Y1: y + 3.6})
			g.Use(Charcoal).FillVolume(Cone(t, 1.6, y+3.6, 1.2))
		}
	case 2:
		f := neonCityFace{P(c.X-3.5, c.Y+d*0.15), P(1, 0), P(0, 1), 7}
		g.Use(Charcoal).Fill(BlockAt(P(c.X-2.8, c.Y+d*0.15-0.3), 0.6, 0.6, y, y+2), BlockAt(P(c.X+2.8, c.Y+d*0.15-0.3), 0.6, 0.6, y, y+2))
		s.billboard(f, 0, 7, y+2, y+6, seed)
	case 3:
		fc := P(21.5, 19)
		g.Use(Slate).Fill(Cylinder{Center: fc, Radius: 4.4, Inner: 3.5, Y0: y, Y1: y + 1.8})
		g.Paint(Red, Cylinder{Center: fc, Radius: 4.5, Inner: 3.4, Y0: y + 1.3, Y1: y + 1.8})
		g.Use(Charcoal).Fill(Cylinder{Center: fc, Radius: 3.6, Y0: y - 0.4, Y1: y + 0.1})
		g.Use(Gold).Fill(Cylinder{Center: fc, Radius: 0.9, Y0: y, Y1: y + 1.8})
		for k := 0; k < 5; k++ {
			a := float64(k) * 2 * math.Pi / 5
			blade := Block(V(0.6, y+0.8, -0.6), V(3.4, y+1.3, 0.6))
			g.Use(White).Fill(Translate(RotateY(blade, P(0, 0), a), V(fc.X, 0, fc.Y)))
		}
		g.Use(Slate).Fill(Capsule{fc.Add(P(-4, 0)).At(y + 1.9), fc.Add(P(4, 0)).At(y + 1.9), 0.45},
			Capsule{fc.Add(P(0, -4)).At(y + 1.9), fc.Add(P(0, 4)).At(y + 1.9), 0.45})
	}
	if Hash01(seed, 77) < 0.6 {
		at := c.Add(P(w*0.3, d*0.3))
		g.Use(Charcoal).Fill(Capsule{at.At(y), at.At(y + 4 + Hash01(seed, 78)*4), 0.5})
	}
}

// railDist is the approximate signed distance from the monorail centerline, and the
// ellipse parameter of the nearest point.
func neonCityRailDist(p Vec3) (float64, float64) {
	dx, dz := (p.X-neonCityRailC.X)/neonCityRailA, (p.Z-neonCityRailC.Y)/neonCityRailB
	t := math.Atan2(dz, dx)
	e := math.Hypot(dx, dz)
	lr := math.Hypot(neonCityRailA*math.Cos(t), neonCityRailB*math.Sin(t))
	return (e - 1) * lr, t
}

func neonCityRailAt(t float64) Vec2 {
	return P(neonCityRailC.X+neonCityRailA*math.Cos(t), neonCityRailC.Y+neonCityRailB*math.Sin(t))
}

// ring selects the band of the monorail loop of width w between heights and parameters.
func neonCityRing(w, y0, y1, t0, t1 float64) Solid {
	b := Box{V(neonCityRailC.X-neonCityRailA-w, y0, neonCityRailC.Y-neonCityRailB-w),
		V(neonCityRailC.X+neonCityRailA+w, y1, neonCityRailC.Y+neonCityRailB+w)}
	return Func{b, func(p Vec3) bool {
		d, t := neonCityRailDist(p)
		if math.Abs(d) > w/2 {
			return false
		}
		if t1-t0 < 2*math.Pi {
			x := math.Mod(t-t0+4*math.Pi, 2*math.Pi)
			return x <= t1-t0
		}
		return true
	}}
}

// monorail cuts portals through the towers, then lays the beam, pillars and a train.
func (s *neonCitySite) monorail() {
	g := s.g
	y := neonCityRail
	var towers []Solid
	for _, l := range s.lots {
		if l.tower {
			towers = append(towers, Block(l.lo.Sub(P(3, 3)).At(0), l.hi.Add(P(3, 3)).At(200)))
		}
	}
	full := 2 * math.Pi
	g.Paint(Gold, Intersect(neonCityRing(6, y-1.4, y+5.8, 0, full), Union(towers...)))
	g.Carve(neonCityRing(4, y-0.6, y+5, 0, full))
	g.Use(Slate).Fill(neonCityRing(1.6, y, y+1.2, 0, full))
	g.Paint(Water, Func{Box{V(-31, y, -22), V(25, y+1.2, 14)}, func(p Vec3) bool {
		d, _ := neonCityRailDist(p)
		return math.Abs(d) >= 0.45 && p.Y >= y+0.35 && p.Y <= y+0.8
	}})
	// Pillars where the ground below is free (or from a tenement roof).
	for k := 0; k < 24; k++ {
		t := float64(k) * full / 24
		at := neonCityRailAt(t)
		base, skip := neonCityG, false
		for _, l := range s.lots {
			if at.X >= l.lo.X-0.5 && at.X <= l.hi.X+0.5 && at.Y >= l.lo.Y-0.5 && at.Y <= l.hi.Y+0.5 {
				if l.tower || l.top > y-2 {
					skip = true
				}
				base = math.Max(base, l.top)
			}
		}
		if skip || (at.X > -9.5 && at.X < -1.5 && at.Y > 9) {
			continue
		}
		g.Use(Slate).Fill(Cylinder{Center: at, Radius: 0.9, Y0: base, Y1: y})
		g.Use(Slate).Fill(Cylinder{Center: at, Radius: 1.4, Y0: y - 1, Y1: y})
		g.Paint(Red, Cylinder{Center: at, Radius: 1.5, Y0: base, Y1: base + 0.5})
	}
	// A four-car train on the south-east arc.
	t0, car, gap := 0.32, 0.17, 0.018
	for i := 0; i < 4; i++ {
		a := t0 + float64(i)*(car+gap)
		body := neonCityRing(2.3, y+1.2, y+3.5, a, a+car)
		g.Use(White).Fill(body)
		g.Paint(Glass, Intersect(body, neonCityRing(2.4, y+2.2, y+2.9, a, a+car)))
		g.Paint(Red, Intersect(body, neonCityRing(2.4, y+1.3, y+1.7, a, a+car)))
		if i == 0 || i == 3 {
			e := a
			if i == 3 {
				e = a + car - 0.03
			}
			g.Paint(Navy, Intersect(body, neonCityRing(2.4, y+1.9, y+3.5, e, e+0.03)))
		}
	}
}

// skyBridge is a glass tube with white ribs between two towers.
func (s *neonCitySite) skyBridge(lo, hi Vec3, alongX bool) {
	g := s.g
	g.Use(Glass).Fill(Block(lo, hi))
	for v := 0.0; ; v += 2.5 {
		var rib Cuboid
		if alongX {
			if lo.X+v > hi.X {
				break
			}
			rib = Block(V(lo.X+v, lo.Y-0.2, lo.Z-0.2), V(lo.X+v+0.6, hi.Y+0.2, hi.Z+0.2))
		} else {
			if lo.Z+v > hi.Z {
				break
			}
			rib = Block(V(lo.X-0.2, lo.Y-0.2, lo.Z+v), V(hi.X+0.2, hi.Y+0.2, lo.Z+v+0.6))
		}
		g.Use(White).Fill(rib)
	}
	g.Paint(Navy, Block(lo, V(hi.X, lo.Y+0.5, hi.Z)))
	in := V(0, 0.8, 0.8)
	if !alongX {
		in = V(0.8, 0.8, 0)
	}
	g.Carve(Block(lo.Add(in), hi.Sub(in)))
}

func (s *neonCitySite) bridges() {
	s.skyBridge(V(-17, 36, -22), V(-9.6, 39.2, -19), true)
	s.skyBridge(V(5.6, 40, -20), V(13.2, 43.2, -17.2), true)
	s.skyBridge(V(-27, 31, -10.8), V(-23.8, 34.2, 11.6), false)
	s.skyBridge(V(-3, 30, -17.6), V(1, 33.2, -11.5), false)
	s.skyBridge(V(1.2, 36, -12.8), V(14.8, 39.2, -9.8), true)
	// Open walkways over the canyon street.
	g := s.g
	for _, w := range []struct{ x0, x1, y float64 }{{-15.5, -13, 12}, {1, 3.5, 10}, {20, 22.5, 11}} {
		y := neonCityG + w.y
		g.Use(Slate).Fill(Block(V(w.x0, y-0.6, 2), V(w.x1, y, 11)))
		g.Paint(Water, Block(V(w.x0, y-0.6, 2), V(w.x1, y-0.3, 11)))
		g.Use(Red).Fill(Block(V(w.x0, y, 2.5), V(w.x0+0.6, y+1, 10.5)), Block(V(w.x1-0.6, y, 2.5), V(w.x1, y+1, 10.5)))
		g.Carve(Block(V(w.x0+0.7, y, 1.2), V(w.x1-0.7, y+2.2, 2.6)), Block(V(w.x0+0.7, y, 10.4), V(w.x1-0.7, y+2.2, 11.8)))
	}
}

// highway is an elevated ramp along the east edge with barriers, pillars, cars and a gantry.
func (s *neonCitySite) highway() {
	g := s.g
	h := func(z float64) float64 {
		if z <= -2 {
			return neonCityG + 12
		}
		return neonCityG + 12*math.Max(0, 26.5-z)/28.5
	}
	x0, x1 := 28.0, 33.0
	b := Box{V(x0, neonCityG-0.3, -27), V(x1, neonCityG+14, 26.5)}
	g.Use(Charcoal).Fill(Func{b, func(p Vec3) bool {
		y := h(p.Z)
		return p.Y <= y && p.Y >= math.Max(neonCityG-0.3, y-0.8)
	}})
	g.Use(White).Fill(Func{Box{b.Min, b.Max.Add(V(0, 1.2, 0))}, func(p Vec3) bool {
		y := h(p.Z)
		return p.Y > y && p.Y <= y+0.9 && (p.X <= x0+0.6 || p.X >= x1-0.6)
	}})
	g.Paint(Red, Func{Box{b.Min, b.Max.Add(V(0, 1.2, 0))}, func(p Vec3) bool {
		y := h(p.Z)
		return p.Y > y+0.5 && math.Mod(p.Z+40, 3) < 0.8
	}})
	g.Paint(Gold, Func{b, func(p Vec3) bool {
		y := h(p.Z)
		return p.Y >= y-0.35 && p.X >= 30 && p.X <= 31 && math.Mod(p.Z+40, 3) < 1.5
	}})
	for z := 18.0; z > -27; z -= 6 {
		y := h(z) - 0.8
		if y-neonCityG < 2 {
			continue
		}
		g.Use(Slate).Fill(Cylinder{Center: P(30.5, z), Radius: 1.1, Y0: neonCityG, Y1: y})
		g.Fill(Block(V(x0, y-0.8, z-0.6), V(x1, y, z+0.6)))
		g.Paint(Water, Cylinder{Center: P(30.5, z), Radius: 1.2, Y0: neonCityG + 2, Y1: neonCityG + 2.5})
	}
	// Cars.
	for _, c := range []struct {
		x, z float64
		col  Color
	}{{29.3, 19, Red}, {31.7, 9, White}, {29.3, -5, Gold}, {31.7, -16, Water}, {29.3, -23, White}} {
		y := h(c.z)
		g.Use(c.col).Fill(Block(V(c.x-0.8, y, c.z-1.6), V(c.x+0.8, y+0.9, c.z+1.6)))
		g.Use(Glass).Fill(Block(V(c.x-0.7, y+0.9, c.z-0.8), V(c.x+0.7, y+1.5, c.z+0.6)))
		g.Use(Red).Fill(Block(V(c.x-0.8, y+0.3, c.z+1.6), V(c.x+0.8, y+0.7, c.z+1.9)))
	}
	// Overhead sign gantry.
	y := h(-20)
	g.Use(Slate).Fill(Block(V(27.2, y-0.8, -20.4), V(27.9, y+6, -19.6)), Block(V(33.1, y-0.8, -20.4), V(33.8, y+6, -19.6)))
	g.Fill(Block(V(27.2, y+5.2, -20.4), V(33.8, y+6, -19.6)))
	s.billboard(neonCityFace{P(28, -19.6), P(1, 0), P(0, 1), 5}, 0, 5, y+2.6, y+5.2, 81)
}

// market fills the canyon street with stalls under striped awnings and lantern strings.
func (s *neonCitySite) market() {
	g := s.g
	G := neonCityG
	pairs := [][2]Color{{Red, White}, {Gold, Red}, {Water, White}, {Glass, Navy}, {Red, Gold}, {White, Water}}
	for side := 0; side < 2; side++ {
		z0, sgn := 2.5, 1.0
		if side == 1 {
			z0, sgn = 10.5, -1
		}
		for k := 0; ; k++ {
			x := -31.5 + float64(k)*3.3
			if x > 24 {
				break
			}
			if (x > 4.5 && x < 10.5) || (side == 1 && x > -10.5 && x < -0.5) || Hash01(k, side, 1) < 0.15 {
				continue
			}
			pc := pairs[int(Hash01(k, side, 2)*float64(len(pairs)))]
			zf := z0 + sgn*1.6
			g.Use(DarkStone).Fill(Block(V(x+0.2, G, z0), V(x+2.8, G+1.1, zf)))
			for j := 0; j < 3; j++ {
				gc := []Color{Gold, Red, Glass, White, Water}[int(Hash01(k, side, j, 3)*5)]
				g.Use(gc).Fill(Block(V(x+0.4+float64(j)*0.8, G+1.1, z0+sgn*0.6), V(x+1.0+float64(j)*0.8, G+1.6, z0+sgn*1.3)))
			}
			zt := z0 + sgn*2.6
			g.Use(Charcoal).Fill(Capsule{V(x+0.4, G, zt), V(x+0.4, G+3, zt), 0.45}, Capsule{V(x+2.6, G, zt), V(x+2.6, G+3, zt), 0.45})
			zlo, zhi := math.Min(z0, z0+sgn*3), math.Max(z0, z0+sgn*3)
			aw := Func{Box{V(x, G+2, zlo), V(x+3, G+4.2, zhi)}, func(p Vec3) bool {
				y := G + 3.8 - 0.35*math.Abs(p.Z-z0)
				return p.Y <= y && p.Y >= y-0.5
			}}
			g.Use(pc[0]).Fill(aw)
			g.Paint(pc[1], Intersect(aw, Func{aw.Box, func(p Vec3) bool { return math.Mod(p.X-x, 1.5) >= 0.75 }}))
		}
	}
	// Lantern strings across the street.
	for k := 0; k < 8; k++ {
		x := -29.0 + float64(k)*7.4
		if x > 4 && x < 11 {
			x += 6
		}
		y := G + 6.5 + Hash01(k, 9)*2
		var pts []Vec3
		for i := 0; i <= 10; i++ {
			u := float64(i) / 10
			pts = append(pts, V(x+u*1.5, y-1.2*(1-(2*u-1)*(2*u-1)), 2.5+u*8))
		}
		for i := 1; i < len(pts); i++ {
			g.Use(Charcoal).Fill(Capsule{pts[i-1], pts[i], 0.45})
		}
		for i := 2; i < len(pts)-1; i += 2 {
			g.Use([]Color{Red, Gold}[(i/2+k)%2]).Fill(Sphere(pts[i].Sub(V(0, 0.8, 0)), 0.6))
		}
	}
}

// cables are sagging power lines high across the canyon street.
func (s *neonCitySite) cables() {
	g := s.g
	G := neonCityG
	for _, c := range []struct{ a, b Vec3 }{
		{V(-29, G+16, 2.5), V(-14, G+13, 10.5)},
		{V(-18, G+19, 2.5), V(-11, G+15, 10.5)},
		{V(-5, G+9, 2.5), V(1, G+17, 10.5)},
		{V(14, G+13, 2.5), V(13, G+15, 10.5)},
		{V(24, G+12, 2.5), V(18, G+10, 10.5)},
	} {
		prev := c.a
		for i := 1; i <= 12; i++ {
			u := float64(i) / 12
			p := c.a.Add(c.b.Sub(c.a).Scale(u)).Sub(V(0, 1.5*(1-(2*u-1)*(2*u-1)), 0))
			g.Use(Charcoal).Fill(Capsule{prev, p, 0.45})
			prev = p
		}
	}
}

// hollow removes cells buried deeper than keep layers.
func (s *neonCitySite) hollow(keep int) {
	g := s.g
	inner := map[Cell]bool{}
	for _, c := range g.Cells() {
		all := true
		for _, n := range Neighbors {
			if !g.Has(c.Add(n)) {
				all = false
				break
			}
		}
		if all {
			inner[c] = true
		}
	}
	for k := 1; k < keep; k++ {
		next := map[Cell]bool{}
		for c := range inner {
			all := true
			for _, n := range Neighbors {
				if !inner[c.Add(n)] {
					all = false
					break
				}
			}
			if all {
				next[c] = true
			}
		}
		inner = next
	}
	for c := range inner {
		g.Remove(c)
	}
}

// extras adds a monorail station over the east tenement and hovering flying cars.
func (s *neonCitySite) extras() {
	g := s.g
	y := neonCityRail + 1.2
	t0, t1 := -0.42, 0.12
	plat := Subtract(neonCityRing(6.4, y-0.6, y, t0, t1), neonCityRing(2.6, y-1, y+1, t0, t1))
	g.Use(Charcoal).Fill(plat)
	g.Paint(Gold, Subtract(plat, neonCityRing(5.6, y-1, y+1, t0, t1)))
	g.Use(Glass).Fill(neonCityRing(7, y+4.2, y+4.8, t0, t1))
	g.Paint(White, Intersect(neonCityRing(7, y+4.2, y+4.8, t0, t1), Subtract(neonCityRing(8, y, y+6, t0, t1), neonCityRing(6.2, y, y+6, t0, t1))))
	for k := 0; k <= 4; k++ {
		t := t0 + (t1-t0)*float64(k)/4
		for _, off := range []float64{-2.9, 2.9} {
			at := neonCityRailAt(t)
			dir := at.Sub(neonCityRailC)
			dir = P(dir.X/neonCityRailA/neonCityRailA, dir.Y/neonCityRailB/neonCityRailB)
			dir = dir.Scale(1 / dir.Length())
			q := at.Add(dir.Scale(off))
			g.Use(White).Fill(Capsule{q.At(y), q.At(y + 4.4), 0.45})
			base := neonCityG
			for _, l := range s.lots {
				if q.X >= l.lo.X && q.X <= l.hi.X && q.Y >= l.lo.Y && q.Y <= l.hi.Y {
					base = math.Max(base, l.top)
				}
			}
			if k%2 == 0 {
				g.Use(Slate).Fill(Cylinder{Center: q, Radius: 0.7, Y0: base, Y1: y - 0.6})
			}
		}
	}
	// Flying cars between the towers.
	for i, c := range []struct {
		at  Vec3
		col Color
		yaw float64
	}{{V(-8, 47, -8), Red, 0.4}, {V(8, 54, -6), White, -0.3}, {V(-12, 30, 5), Gold, 1.2}, {V(26, 44, -4), Water, 1.6}} {
		body := RotateY(Ellipsoid{V(0, 0, 0), V(2, 0.8, 1.1)}, P(0, 0), c.yaw)
		canopy := RotateY(Ellipsoid{V(0.4, 0.6, 0), V(1.0, 0.6, 0.8)}, P(0, 0), c.yaw)
		g.Use(c.col).FillVolume(Translate(body, c.at))
		g.Use(Glass).FillVolume(Translate(canopy, c.at))
		tail := P(-2, 0).Rotate(c.yaw)
		g.Use([]Color{Gold, Red, Red, White}[i]).Fill(Sphere(c.at.Add(V(tail.X, 0, tail.Y)), 0.5))
	}
}
