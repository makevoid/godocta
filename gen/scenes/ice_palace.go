package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// IcePalace is a winter-festival ice palace on the shore of a frozen lake, in the
// tradition of the great carnival ice castles. Every wall is laid from cut ice blocks in
// a running bond: translucent Glass, deep Water-blue and a few frosted White blocks, with
// frosted courses, white copings dripping icicles and dark arched windows. Crenellated
// curtain walls join four corner towers with lantern stages and pointed caps; a grand
// gatehouse with flanking towers opens on the lake; and a stepped central keep rises
// through two crenellated tiers and an octagonal drum into a striped onion dome and a
// gold spire. Around it lie a snowy shore with snowbanks and snow-laden spruces, a
// cracked lake with skaters, an ice slide, bear and swan ice sculptures, red-roofed
// wooden market stalls and lantern posts. The gate faces +Z, towards the hero camera.
type IcePalace struct{}

func init() { Register(IcePalace{}) }

func (IcePalace) Name() string  { return "ice_palace" }
func (IcePalace) Title() string { return "Ice Palace" }

const (
	icePalaceG   = 2.2  // Snow plateau.
	icePalaceL   = 1.3  // Lake ice surface.
	icePalaceF   = 2.8  // Palace floor.
	icePalaceHX  = 41.0 // Site half width.
	icePalaceZ0  = -40.0
	icePalaceZ1  = 35.0
	icePalaceRim = 7.0 // Width of the slope down to the site rim.

	icePalaceWX0, icePalaceWX1 = -21.0, 21.0 // Curtain wall outer faces.
	icePalaceWZ0, icePalaceWZ1 = -24.0, 8.0
	icePalaceWT                = 1.6              // Curtain wall thickness.
	icePalaceW                 = icePalaceF + 9.0 // Curtain wall top.

	icePalaceBW = 2.6 // Ice block length.
	icePalaceBH = 1.2 // Ice block course height.
)

var (
	icePalaceChuteA = P(-36, -4.6)
	icePalaceChuteB = P(-25, 20)
	icePalaceKeepC  = P(0, -8)
	icePalaceLakeC  = P(0, 21.5)
)

const icePalaceLakeA, icePalaceLakeB = 29.0, 9.5

// icePalaceRound is a round tower used to lay its blocks around the circumference.
type icePalaceRound struct {
	c Vec2
	r float64
}

func (IcePalace) smooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

// siteDist is the signed distance inside the rounded-rectangle site.
func (IcePalace) siteDist(x, z float64) float64 {
	const r = 12.0
	cz := (icePalaceZ0 + icePalaceZ1) / 2
	hz := (icePalaceZ1 - icePalaceZ0) / 2
	qx := math.Abs(x) - (icePalaceHX - r)
	qz := math.Abs(z-cz) - (hz - r)
	out := math.Hypot(math.Max(qx, 0), math.Max(qz, 0)) + math.Min(math.Max(qx, qz), 0)
	return r - out
}

// lakeE is the normalized elliptical radius of the lake (1 at its edge).
func (IcePalace) lakeE(x, z float64) float64 {
	dx, dz := (x-icePalaceLakeC.X)/icePalaceLakeA, (z-icePalaceLakeC.Y)/icePalaceLakeB
	return math.Sqrt(dx*dx + dz*dz)
}

// palaceMask is 1 on and around the palace footprint, fading to 0 outside it.
func (s IcePalace) palaceMask(x, z float64) float64 {
	dx := math.Max(math.Max(icePalaceWX0-6-x, x-(icePalaceWX1+6)), 0)
	dz := math.Max(math.Max(icePalaceWZ0-6-z, z-(icePalaceWZ1+5)), 0)
	return 1 - s.smooth(math.Hypot(dx, dz)/5)
}

// height is the snow (or lake ice) surface.
func (s IcePalace) height(x, z float64) float64 {
	e := s.lakeE(x, z)
	bumps := 0.8*math.Sin(x*0.23+1.3)*math.Sin(z*0.19+0.4) + 0.45*math.Sin((x+z)*0.37+2.1) +
		0.3*math.Sin(x*0.61-z*0.43)
	bumps = math.Max(bumps, -0.2) * (1 - s.palaceMask(x, z))
	base := icePalaceG + bumps
	// Snowbanks pushed off the rink, open in front of the gate.
	bank := 1.0 * math.Exp(-math.Pow((e-1.32)/0.13, 2))
	if math.Abs(x) < 6 && z < icePalaceLakeC.Y {
		bank *= s.smooth((math.Abs(x) - 4) / 2)
	}
	if e < 1 {
		base = icePalaceL
	} else if e < 1.4 {
		t := s.smooth((e - 1) / 0.25)
		base = icePalaceL + (base+bank-icePalaceL)*t
	}
	rim := s.smooth(s.siteDist(x, z) / icePalaceRim)
	return 0.3 + (base-0.3)*rim
}

// iceColor lays ice blocks in a running bond.
func (IcePalace) iceColor(p Vec3, rounds []icePalaceRound, seed int) Color {
	along := p.X + p.Z
	for _, t := range rounds {
		q := p.Plan().Sub(t.c)
		if q.Length() < t.r+1.0 {
			along = (math.Atan2(q.Y, q.X) + math.Pi) * t.r
			seed += 1000 + int(t.c.X*7+t.c.Y*13)
			break
		}
	}
	k := int(math.Floor((p.Y + 0.2) / icePalaceBH))
	shift := 0.0
	if k%2 != 0 {
		shift = icePalaceBW / 2
	}
	b := int(math.Floor((along + shift) / icePalaceBW))
	h := Hash01(b, k, seed)
	if k%8 == 3 {
		if h < 0.6 {
			return White
		}
		return Glass
	}
	switch {
	case h < 0.46:
		return Glass
	case h < 0.86:
		return Water
	case h < 0.94:
		return Navy
	}
	return White
}

func (IcePalace) set(g *Grid, c Cell, col Color) {
	prev := g.Brush()
	g.Remove(c)
	g.Use(col)
	g.Add(c)
	g.Use(prev)
}

// topExposed reports whether nothing rests on the cell.
func (IcePalace) topExposed(g *Grid, c Cell) bool {
	for _, d := range []Cell{{0, 2, 0}, {1, 1, 1}, {1, 1, -1}, {-1, 1, 1}, {-1, 1, -1}} {
		if g.Has(c.Add(d)) {
			return false
		}
	}
	return true
}

// snowCap whitens exposed tops of eligible cells.
func (s IcePalace) snowCap(g *Grid, eligible func(Color) bool, chance float64, seed int) {
	var tops []Cell
	for _, c := range g.Cells() {
		col, _ := g.ColorOf(c)
		if !eligible(col) || !s.topExposed(g, c) {
			continue
		}
		if chance < 1 && Hash01(int(c.X), int(c.Y), int(c.Z), seed) >= chance {
			continue
		}
		tops = append(tops, c)
	}
	for _, c := range tops {
		s.set(g, c, White)
	}
}

// faceArch is an arched opening on the face dist from c at a plan angle, offset along it.
func (IcePalace) faceArch(c Vec2, dist, off, angle, sill, hw, spring, depth float64) Solid {
	return RotateY(Arch{V(c.X+dist, sill, c.Y+off), hw, spring, depth, AxisX}, c, angle)
}

// window carves a framed arched window and darkens the room behind it.
func (s IcePalace) window(g *Grid, c Vec2, dist, off, angle, sill, hw, spring, depth float64) {
	g.Paint(White, s.faceArch(c, dist, off, angle, sill-0.4, hw+0.45, spring+0.35, depth))
	g.Carve(s.faceArch(c, dist, off, angle, sill, hw, spring, depth))
}

// blind carves a recessed arch into a face and darkens its back.
func (s IcePalace) blind(g *Grid, c Vec2, dist, off, angle, sill, hw, spring float64) {
	g.Paint(White, s.faceArch(c, dist-0.4, off, angle, sill-0.4, hw+0.45, spring+0.35, 1.0))
	g.Carve(s.faceArch(c, dist, off, angle, sill, hw, spring, 1.4))
	g.Paint(Navy, s.faceArch(c, dist-1.1, off, angle, sill, hw, spring, 0.8))
}

func (IcePalace) square(c Vec2, half, y0, y1 float64) Cuboid {
	return BlockAt(c, 2*half, 2*half, y0, y1)
}

// squareShell is a hollow square storey with a darkened inner face.
func (s IcePalace) squareShell(g *Grid, c Vec2, half, thick, y0, y1 float64) {
	g.Fill(Subtract(s.square(c, half, y0, y1), s.square(c, half-thick, y0-1, y1+1)))
}

func (s IcePalace) squareDark(g *Grid, c Vec2, half, thick, y0, y1 float64) {
	in := half - thick
	g.Paint(Navy, Subtract(s.square(c, in+0.7, y0, y1), s.square(c, in, y0-1, y1+1)))
}

// squareTop adds a white corbelled coping and a crenellated parapet above y.
func (s IcePalace) squareTop(g *Grid, c Vec2, half, y float64) {
	g.With(White, func() {
		g.Fill(Subtract(s.square(c, half+0.6, y-0.8, y), s.square(c, half-1.4, y-2, y+1)))
	})
	g.Fill(Func{Box{V(c.X-half-0.6, y, c.Y-half-0.6), V(c.X+half+0.6, y+1.4, c.Y+half+0.6)}, func(p Vec3) bool {
		dx, dz := p.X-c.X, p.Z-c.Y
		if math.Max(math.Abs(dx), math.Abs(dz)) < half-0.8 {
			return false
		}
		a := dx
		if math.Abs(dx) > math.Abs(dz) {
			a = dz
		}
		return math.Mod(a+half+100*2.8, 2.8) < 1.6
	}})
	s.icicles(g, func(t float64) (Vec2, bool) {
		per := 8 * half
		d := t * per
		side := int(d / (2 * half))
		u := math.Mod(d, 2*half) - half
		o := half + 0.35
		switch side {
		case 0:
			return c.Add(P(u, o)), true
		case 1:
			return c.Add(P(o, -u)), true
		case 2:
			return c.Add(P(-u, -o)), true
		}
		return c.Add(P(-o, u)), true
	}, int(8*half/1.1), y-0.8, int(c.X*31+c.Y*17+y*5))
}

// icicles hangs white icicles below a coping at n points along a loop.
func (IcePalace) icicles(g *Grid, at func(t float64) (Vec2, bool), n int, y float64, seed int) {
	g.With(White, func() {
		for i := 0; i < n; i++ {
			h := Hash01(i, seed, 5)
			if h < 0.25 {
				continue
			}
			q, ok := at((float64(i) + 0.5) / float64(n))
			if !ok {
				continue
			}
			l := 0.6 + 2.0*Hash01(i, seed, 9)
			g.Fill(Capsule{q.At(y - 0.2), q.At(y - l), 0.42})
		}
	})
}

// roundTower is a hollow round tower with dark windows, a coping and a parapet.
func (s IcePalace) roundTower(g *Grid, c Vec2, r, y0, top float64, rows []float64, n int) {
	g.Fill(Cylinder{c, r, r - 1.3, y0, top})
	g.Fill(Cylinder{c, r - 1.0, 0, top - 1.2, top})
	for row, sill := range rows {
		for k := 0; k < n; k++ {
			a := (float64(k) + 0.5*float64(row%2)) * 2 * math.Pi / float64(n)
			s.window(g, c, r-0.6, 0, a, sill, 0.85, 2.0, 2.6)
		}
	}
	g.Paint(Navy, Cylinder{c, r - 1.3 + 0.01, r - 2.1, y0, top - 1.2})
	g.With(White, func() { g.Fill(Cylinder{c, r + 0.7, r - 1.0, top - 0.9, top}) })
	m := int(math.Max(8, math.Round(r*2.6)))
	g.Fill(Func{Box{V(c.X-r-0.7, top, c.Y-r-0.7), V(c.X+r+0.7, top+1.4, c.Y+r+0.7)}, func(p Vec3) bool {
		q := p.Plan().Sub(c)
		d := q.Length()
		if d > r+0.7 || d < r-0.5 {
			return false
		}
		a := (math.Atan2(q.Y, q.X) + math.Pi) / (2 * math.Pi) * float64(m)
		return a-math.Floor(a) < 0.55
	}})
	s.icicles(g, func(t float64) (Vec2, bool) { return Polar(c, r+0.35, t*2*math.Pi), true },
		int(2*math.Pi*r/1.1), top-0.9, int(c.X*13+c.Y*29))
}

// cap is a hollow pointed cap in deep ice with white ribs and a gold finial.
func (IcePalace) cap(g *Grid, c Vec2, r, base, h float64, ribs int) {
	g.With(Water, func() {
		g.FillVolume(Subtract(Cone(c, r, base, h), Frustum{c, r - 1.3, 0, base - 0.01, base + h - 2.2}))
	})
	g.Paint(White, Func{Box{V(c.X-r, base, c.Y-r), V(c.X+r, base+h, c.Y+r)}, func(p Vec3) bool {
		q := p.Plan().Sub(c)
		a := (math.Atan2(q.Y, q.X) + math.Pi) / (2 * math.Pi) * float64(ribs)
		return a-math.Floor(a) < 0.22 || p.Y < base+0.7
	}})
	g.With(Gold, func() {
		g.Fill(Capsule{c.At(base + h - 0.8), c.At(base + h + 1.6), 0.45})
		g.Fill(Sphere(c.At(base+h+1.0), 0.75))
	})
}

// onion is a bulb dome profile from the neck at base.
func (IcePalace) onion(c Vec2, base, neck, R, H float64) Solid {
	const tm = 0.36
	return Func{Box{V(c.X-R, base, c.Y-R), V(c.X+R, base+H, c.Y+R)}, func(p Vec3) bool {
		t := (p.Y - base) / H
		var r float64
		if t < tm {
			u := (tm - t) / tm
			r = neck + (R-neck)*math.Sqrt(1-u*u)
		} else {
			v := (t - tm) / (1 - tm)
			r = R * math.Pow(1-v, 1.25) * (1 + 1.2*v)
		}
		return p.Plan().Sub(c).Length() <= r
	}}
}

// octShell is an octagonal prism shell with apothem a.
func (IcePalace) oct(c Vec2, a, y0, y1 float64) Solid {
	r := a / math.Cos(math.Pi/8)
	return Func{Box{V(c.X-r, y0, c.Y-r), V(c.X+r, y1, c.Y+r)}, func(p Vec3) bool {
		q := p.Plan().Sub(c)
		d := math.Max(math.Max(math.Abs(q.X), math.Abs(q.Y)), (math.Abs(q.X)+math.Abs(q.Y))/math.Sqrt2)
		return d <= a
	}}
}

func (s IcePalace) Build(g *Grid) Order {
	s.terrain(g)
	pg := NewGridWithSpan(g.Span())
	s.palace(pg)
	// The palace floor replaces the snow beneath it.
	g.Carve(Block(V(icePalaceWX0-1.5, -1, icePalaceWZ0-1.5), V(icePalaceWX1+1.5, icePalaceF, icePalaceWZ1+3.6)))
	g.Merge(pg)
	tg := NewGridWithSpan(g.Span())
	s.trees(tg)
	g.Merge(tg)
	s.slide(g)
	s.stalls(g)
	s.sculptures(g)
	s.lake(g)
	s.lanterns(g)
	s.snowmen(g)
	return Sweep(icePalaceKeepC, math.Pi/2)
}

// terrain is a snow shell over the heightfield, with the lake ice set into it.
func (s IcePalace) terrain(g *Grid) {
	box := Box{V(-icePalaceHX, 0, icePalaceZ0), V(icePalaceHX, icePalaceG+2.5, icePalaceZ1)}
	g.Use(White).Fill(Func{box, func(p Vec3) bool {
		if s.siteDist(p.X, p.Z) < 0 {
			return false
		}
		h := s.height(p.X, p.Z)
		return p.Y <= h && p.Y >= math.Max(0, h-1.3)
	}})
	// Lake ice: translucent Glass, frosted near the banks.
	g.Paint(Glass, Func{box, func(p Vec3) bool {
		e := s.lakeE(p.X, p.Z)
		return e < 1.02 && p.Y <= icePalaceL+0.1
	}})
	g.Paint(White, Func{box, func(p Vec3) bool {
		e := s.lakeE(p.X, p.Z)
		n := math.Sin(p.X*0.5+1.1)*math.Sin(p.Z*0.7+0.3) + 0.5*math.Sin(p.X*1.3-p.Z*0.9)
		return e < 1.02 && e > 0.72 && n+(e-0.72)*3 > 1.1
	}})
	// Cracks: deterministic random walks in deep Water blue.
	g.Use(Water)
	for i := 0; i < 16; i++ {
		a := Hash01(i, 1) * 2 * math.Pi
		r := math.Sqrt(Hash01(i, 2)) * 0.8
		p := P(icePalaceLakeC.X+r*icePalaceLakeA*math.Cos(a), icePalaceLakeC.Y+r*icePalaceLakeB*math.Sin(a))
		dir := Hash01(i, 3) * 2 * math.Pi
		for j := 0; j < 6; j++ {
			dir += (Hash01(i, j, 4) - 0.5) * 1.6
			q := p.Add(P(math.Cos(dir), math.Sin(dir)).Scale(1.6 + 1.6*Hash01(i, j, 5)))
			if s.lakeE(q.X, q.Y) > 0.95 {
				break
			}
			g.Paint(Water, Capsule{p.At(icePalaceL), q.At(icePalaceL), 0.45})
			if Hash01(i, j, 6) < 0.35 {
				b := dir + 1.2*(Hash01(i, j, 7)-0.5) + math.Pi/3
				e := p.Add(P(math.Cos(b), math.Sin(b)).Scale(2.2))
				if s.lakeE(e.X, e.Y) < 0.95 {
					g.Paint(Water, Capsule{p.At(icePalaceL), e.At(icePalaceL), 0.45})
				}
			}
			p = q
		}
	}
	g.Use(White)
}

func (s IcePalace) palace(g *Grid) {
	F, W, T := icePalaceF, icePalaceW, icePalaceF+15.0
	x0, x1, z0, z1, wt := icePalaceWX0, icePalaceWX1, icePalaceWZ0, icePalaceWZ1, icePalaceWT
	c := P((x0+x1)/2, (z0+z1)/2)
	hx, hz := (x1-x0)/2, (z1-z0)/2

	// Floor: a snow-covered platform with an ice-tiled forecourt.
	g.Use(White).Fill(Block(V(x0-1.5, F-1.0, z0-1.5), V(x1+1.5, F, z1+3.6)))

	// Curtain walls with blind arcades, coping, icicles and merlons.
	g.Use(Glass)
	outer := BlockAt(c, 2*hx, 2*hz, F-1, W)
	inner := BlockAt(c, 2*(hx-wt), 2*(hz-wt), F-2, W+3)
	g.Fill(Subtract(outer, inner))
	for _, side := range []struct {
		angle, dist, half float64
	}{{math.Pi / 2, hz, hx}, {-math.Pi / 2, hz, hx}, {0, hx, hz}, {math.Pi, hx, hz}} {
		for off := -side.half + 8.5; off <= side.half-8.5+0.01; off += 3.6 {
			if side.angle == math.Pi/2 && math.Abs(off) < 12 {
				continue // Gatehouse.
			}
			if side.dist == hx && math.Abs(off+c.Y-icePalaceKeepC.Y) < 4 {
				continue // Side turrets.
			}
			if side.angle == -math.Pi/2 && math.Abs(math.Abs(off)-9) < 3.5 {
				continue // Back turrets.
			}
			s.blind(g, c, side.dist-0.3, off, side.angle, F+2.2, 0.95, 3.0)
		}
	}
	g.With(White, func() {
		g.Fill(Subtract(BlockAt(c, 2*hx+1.2, 2*hz+1.2, W-0.8, W), BlockAt(c, 2*(hx-wt)-0.2, 2*(hz-wt)-0.2, W-2, W+1)))
	})
	g.Fill(Func{Box{V(x0-0.6, W, z0-0.6), V(x1+0.6, W+1.4, z1+0.6)}, func(p Vec3) bool {
		onZ := p.Z <= z0+wt+0.1 || p.Z >= z1-wt-0.1
		onX := p.X <= x0+wt+0.1 || p.X >= x1-wt-0.1
		if !onZ && !onX {
			return false
		}
		if p.X < x0-0.1 || p.X > x1+0.1 || p.Z < z0-0.1 || p.Z > z1+0.1 {
			return false
		}
		a := p.X
		if onX && !onZ {
			a = p.Z
		}
		return math.Mod(a+280, 2.8) < 1.6
	}})
	s.icicles(g, func(t float64) (Vec2, bool) {
		per := 4*hx + 4*hz
		d := t * per
		o := 0.35
		switch {
		case d < 2*hx:
			return P(x0+d, z1+o), true
		case d < 2*hx+2*hz:
			return P(x1+o, z1-(d-2*hx)), true
		case d < 4*hx+2*hz:
			return P(x1-(d-2*hx-2*hz), z0-o), true
		}
		return P(x0-o, z0+(d-4*hx-2*hz)), true
	}, int((4*hx+4*hz)/1.1), W-0.8, 77)

	// Courtyard halls flanking the keep.
	for _, sx := range []float64{-1, 1} {
		hc := P(sx*14.5, -10)
		g.Use(Glass)
		g.Fill(Subtract(BlockAt(hc, 6, 18, F-1, F+6), BlockAt(hc, 3.6, 15.6, F-2, F+5)))
		g.Fill(BlockAt(hc, 6, 18, F+5, F+6))
		for off := -6.0; off <= 6.01; off += 3.0 {
			s.window(g, hc, 3-0.6, off, math.Pi*(1-sx)/2+math.Pi, F+1.2, 0.8, 2.0, 2.6)
			s.window(g, hc, 3-0.6, off, math.Pi*(1-sx)/2, F+1.2, 0.8, 2.0, 2.6)
		}
		g.Paint(Navy, Subtract(BlockAt(hc, 3.6+1.4, 15.6+1.4, F, F+5), BlockAt(hc, 3.6, 15.6, F-2, F+6)))
	}

	// Keep: two crenellated square tiers and an octagonal drum.
	k := icePalaceKeepC
	K1, K2, K3 := F+15.0, F+26.0, F+33.0
	g.Use(Glass)
	s.squareShell(g, k, 8, 1.5, F-1, K1)
	g.Fill(s.square(k, 7, K1-1.2, K1))
	s.squareShell(g, k, 5.5, 1.4, K1-1.2, K2)
	g.Fill(s.square(k, 4.6, K2-1.2, K2))
	for q := 0; q < 4; q++ {
		a := float64(q) * math.Pi / 2
		for _, sill := range []float64{F + 2.5, F + 8.5} {
			for _, off := range []float64{-4.5, 0, 4.5} {
				if q == 1 && sill < F+5 && off == 0 {
					// The door faces the gate.
					g.Paint(White, s.faceArch(k, 7.3, 0, a, F-0.5, 2.3, 3.6, 2.6))
					g.Carve(s.faceArch(k, 7.3, 0, a, F, 1.7, 3.2, 2.6))
					continue
				}
				s.window(g, k, 7.3, off, a, sill, 1.0, 2.6, 2.6)
			}
		}
		for _, off := range []float64{-2.4, 2.4} {
			s.window(g, k, 4.9, off, a, K1+2.2, 0.95, 2.8, 2.6)
		}
	}
	s.squareDark(g, k, 8, 1.5, F-1, K1-1.2)
	s.squareDark(g, k, 5.5, 1.4, K1, K2-1.2)
	s.squareTop(g, k, 8, K1)
	s.squareTop(g, k, 5.5, K2)
	g.Use(Glass)
	g.Fill(Subtract(s.oct(k, 4.0, K2-0.5, K3), s.oct(k, 2.9, K2, K3+1)))
	g.Fill(s.oct(k, 3.2, K3-1.2, K3))
	for q := 0; q < 8; q++ {
		s.window(g, k, 3.5, 0, float64(q)*math.Pi/4, K2+1.4, 0.85, 2.3, 2.2)
	}
	g.Paint(Navy, Subtract(s.oct(k, 2.9+0.01, K2, K3-1.2), s.oct(k, 2.2, K2-1, K3)))

	// Corner towers, side turrets and back turrets.
	var rounds []icePalaceRound
	for _, tc := range []Vec2{P(x0, z0), P(x1, z0), P(x0, z1), P(x1, z1)} {
		g.Use(Glass)
		s.roundTower(g, tc, 4.4, F-1, T, []float64{F + 3, F + 8.2}, 6)
		g.Fill(Cylinder{tc, 3.0, 2.0, T, T + 4.5})
		for q := 0; q < 6; q++ {
			s.window(g, tc, 2.5, 0, float64(q)*math.Pi/3+0.3, T+1, 0.8, 1.9, 2)
		}
		g.Paint(Navy, Cylinder{tc, 2.01, 1.2, T, T + 4.5})
		g.With(White, func() { g.Fill(Cylinder{tc, 3.5, 2.0, T + 3.9, T + 4.5}) })
		rounds = append(rounds, icePalaceRound{tc, 4.4})
	}
	for _, tc := range []Vec2{P(x0, k.Y), P(x1, k.Y), P(-9, z0), P(9, z0)} {
		g.Use(Glass)
		s.roundTower(g, tc, 2.6, F-1, W+4, []float64{F + 5}, 4)
		rounds = append(rounds, icePalaceRound{tc, 2.6})
	}
	// Corbelled turrets on the keep's corners.
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			tc := k.Add(P(sx*8, sz*8))
			g.Use(Glass)
			g.FillVolume(Frustum{tc, 0.5, 2.0, F + 4, F + 7})
			g.Fill(Cylinder{tc, 2.0, 0, F + 7, K1 + 4.5})
			g.With(White, func() { g.Fill(Cylinder{tc, 2.5, 0, K1 + 3.9, K1 + 4.5}) })
			for q := 0; q < 4; q++ {
				g.Paint(Navy, s.faceArch(tc, 1.6, 0, float64(q)*math.Pi/2+math.Pi/4, K1+0.8, 0.6, 1.6, 1))
			}
			rounds = append(rounds, icePalaceRound{tc, 2.0})
		}
	}

	// Gatehouse with its flanking towers and a dark passage to the doors.
	gc := P(0, 8)
	GH := F + 13
	g.Use(Glass)
	g.Fill(Subtract(BlockAt(gc, 15, 6.4, F-1, GH), BlockAt(gc, 12.2, 3.6, F-2, GH-1.2)))
	g.Fill(BlockAt(gc, 15, 6.4, GH-1.2, GH))
	for _, off := range []float64{-4, 0, 4} {
		s.window(g, gc, 2.6, off, math.Pi/2, F+9.0, 0.9, 1.9, 2.2)
	}
	g.Paint(Navy, Subtract(BlockAt(gc, 13.6, 5.0, F, GH-1.2), BlockAt(gc, 12.2, 3.6, F-2, GH)))
	for _, sx := range []float64{-1, 1} {
		tc := P(sx*8, 10.6)
		g.Use(Glass)
		s.roundTower(g, tc, 2.8, F-1, F+17, []float64{F + 4, F + 10}, 4)
		rounds = append(rounds, icePalaceRound{tc, 2.8})
	}
	gate := func(hw, spring, z, depth float64) Solid {
		return Arch{V(0, F, z), hw, spring, depth, AxisZ}
	}
	g.Paint(Navy, gate(3.9, 4.6, 8.3, 5.4))
	g.Paint(White, gate(4.2, 4.6, 10.9, 1.0))
	g.Carve(gate(3.0, 4.6, 8.8, 5.2))
	g.Paint(Navy, gate(3.0, 4.6, 5.6, 1.4))
	g.Paint(Gold, Capsule{V(0, F+0.5, 5.6), V(0, F+6, 5.6), 0.3})

	// Lay the ice blocks, then let snow settle on every ledge.
	for _, c := range g.Cells() {
		if col, _ := g.ColorOf(c); col == Glass {
			s.set(g, c, s.iceColor(g.World(c), rounds, 17))
		}
	}
	s.snowCap(g, func(col Color) bool { return col == Glass || col == Water || col == Navy }, 0.75, 3)

	// Roofs, caps and the onion dome.
	for _, sx := range []float64{-1, 1} {
		hc := P(sx*14.5, -10)
		hg := NewGridWithSpan(g.Span())
		hg.Use(Water).Fill(Gable{P(hc.X-3.6, hc.Y-9.6), P(hc.X+3.6, hc.Y+9.6), F + 6, AxisZ})
		roof := Gable{P(hc.X-3.6, hc.Y-9.6), P(hc.X+3.6, hc.Y+9.6), F + 6, AxisZ}
		hg.Paint(White, Func{roof.Bounds(), func(p Vec3) bool {
			h := Hash01(int(math.Round(p.X*4)), int(math.Round(p.Y*4)), int(math.Round(p.Z*4)), 19)
			return !roof.Contains(p.Add(V(0, 0.9, 0))) && (h < 0.55 || p.Y > F+8.6)
		}})
		g.Merge(hg)
	}
	for _, tc := range []Vec2{P(x0, z0), P(x1, z0), P(x0, z1), P(x1, z1)} {
		s.cap(g, tc, 3.7, T+4.5, 10, 8)
	}
	for _, tc := range []Vec2{P(x0, k.Y), P(x1, k.Y), P(-9, z0), P(9, z0)} {
		g.Use(Glass).Fill(Cylinder{tc, 1.7, 0, W + 4, W + 5.5})
		s.cap(g, tc, 2.8, W+5.5, 6.5, 6)
	}
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			s.cap(g, k.Add(P(sx*8, sz*8)), 2.4, K1+4.5, 5.5, 6)
		}
		s.cap(g, P(sx*8, 10.6), 3.3, F+18.4, 8, 6)
		g.Use(Glass).Fill(Cylinder{P(sx*8, 10.6), 2.2, 0, F + 17, F + 18.4})
	}
	g.With(Glass, func() {
		g.FillVolume(Subtract(s.onion(k, K3, 3.3, 5.4, 11), s.onion(k, K3-0.01, 2.1, 4.2, 9.4)))
	})
	g.Paint(Water, Func{Box{V(k.X-6, K3, k.Y-6), V(k.X+6, K3+11, k.Y+6)}, func(p Vec3) bool {
		q := p.Plan().Sub(k)
		a := (math.Atan2(q.Y, q.X)+math.Pi)/(2*math.Pi)*12 + (p.Y-K3)*0.18
		return a-math.Floor(a) < 0.45
	}})
	g.Paint(White, Cylinder{k, 6, 0, K3, K3 + 1.0})
	g.With(Gold, func() {
		g.Fill(Capsule{k.At(K3 + 10.2), k.At(K3 + 14.8), 0.5})
		g.Fill(Sphere(k.At(K3+11.6), 0.9))
	})
	// Lamps at the gate.
	g.With(Gold, func() {
		for _, sx := range []float64{-1, 1} {
			g.Fill(Sphere(V(sx*4.9, F+5.2, 11.6), 0.55))
		}
	})
	// Steps from the gate down to the shore.
	g.Use(White).Fill(Block(V(-5, icePalaceG-0.6, 11), V(5, F-0.4, 12.4)))
}

// spruce is a tiered snow-laden conifer.
func (IcePalace) spruce(g *Grid, c Vec2, base, H float64) {
	g.Use(Wood).Fill(Cylinder{c, 0.55, 0, base - 0.8, base + 1.8})
	g.Use(Foliage)
	R := H * 0.3
	crown := H - 1.2
	for i := 0; i < 4; i++ {
		fi := float64(i)
		yb := base + 1.2 + fi*crown*0.21
		h := crown * 0.4
		if i == 3 {
			h = base + H - yb
		}
		g.FillVolume(Cone(c, R*(1-0.21*fi), yb, h))
	}
}

func (s IcePalace) trees(g *Grid) {
	A, B := icePalaceChuteA, icePalaceChuteB
	for gz := icePalaceZ0; gz <= icePalaceZ1; gz += 5.2 {
		for gx := -icePalaceHX; gx <= icePalaceHX; gx += 5.2 {
			i, j := int(gx*10), int(gz*10)
			x := gx + (Hash01(i, j, 1)-0.5)*3.6
			z := gz + (Hash01(i, j, 2)-0.5)*3.6
			q := P(x, z)
			ab := B.Sub(A)
			t := math.Max(0, math.Min(1, q.Sub(A).Dot(ab)/ab.Dot(ab)))
			switch {
			case s.siteDist(x, z) < 3.5,
				x > icePalaceWX0-5.5 && x < icePalaceWX1+5.5 && z > icePalaceWZ0-5.5 && z < icePalaceWZ1+6,
				s.lakeE(x, z) < 1.45,
				q.Sub(A.Add(ab.Scale(t))).Length() < 5,
				x > -42 && x < -31 && z > -22 && z < -2,
				x > 25 && x < 34.5 && z > -24 && z < 9,
				q.Sub(P(33, 21)).Length() < 7, q.Sub(P(24, 17)).Length() < 5, q.Sub(P(-30, 27)).Length() < 4,
				z > 23 && math.Abs(x) < 30,
				Hash01(i, j, 3) > 0.85:
				continue
			}
			h := 7 + 4*Hash01(i, j, 4)
			s.spruce(g, q, s.height(x, z), h)
		}
	}
	s.snowCap(g, func(col Color) bool { return col == Foliage }, 0.75, 11)
}

// chute is the ice slide's channel frame.
func (s IcePalace) chute(a, b Vec2, ya, yb float64) (func(p Vec3) (t, u, ys float64, ok bool), Box) {
	d := b.Sub(a)
	l := d.Length()
	dir := d.Scale(1 / l)
	n := P(-dir.Y, dir.X)
	f := func(p Vec3) (float64, float64, float64, bool) {
		q := p.Plan().Sub(a)
		t, u := q.Dot(dir), q.Dot(n)
		if t < 0 || t > l {
			return t, u, 0, false
		}
		f := t / l
		return t, u, yb + (ya-yb)*math.Pow(1-f, 1.5), true
	}
	box := BoxOf(a.At(yb-2), b.At(ya+3)).Grow(3)
	return f, box
}

// slide is an ice slide: a crenellated tower with stairs and a long railed chute.
func (s IcePalace) slide(g *Grid) {
	sg := NewGridWithSpan(g.Span())
	tc := P(-36, -8)
	top := icePalaceG + 7.5
	A, B := icePalaceChuteA, icePalaceChuteB
	frame, box := s.chute(A, B, top-0.4, icePalaceL+0.4)
	// Clear any snow that pokes into the channel.
	g.Carve(Func{box, func(p Vec3) bool {
		_, u, ys, ok := frame(p)
		return ok && math.Abs(u) <= 2.6 && p.Y > ys-0.9 && p.Y < ys+3
	}})
	sg.Use(Glass)
	sg.Fill(Subtract(BlockAt(tc, 6, 6, icePalaceG-1, top), BlockAt(tc, 3.6, 3.6, icePalaceG-2, top-1.2)))
	sg.Fill(BlockAt(tc, 6, 6, top-1.2, top))
	s.squareTop(sg, tc, 3, top)
	sg.Carve(Block(V(tc.X-1.6, top, tc.Y+1.8), V(tc.X+1.6, top+2, tc.Y+3.6)))
	for _, a := range []float64{0, math.Pi} {
		s.window(sg, tc, 2.4, 0, a, icePalaceG+1.5, 0.8, 2.0, 2.4)
	}
	sg.Paint(Navy, Subtract(BlockAt(tc, 5, 5, icePalaceG-1, top-1.2), BlockAt(tc, 3.6, 3.6, icePalaceG-2, top)))
	for i := 0; i < 8; i++ {
		z1 := tc.Y - 3 - float64(i)*1.05
		y := top - float64(i+1)*0.95
		sg.Fill(Block(V(tc.X-1.5, icePalaceG-0.6, z1-1.05), V(tc.X+1.5, y, z1)))
	}
	sg.Fill(Func{box, func(p Vec3) bool {
		_, u, ys, ok := frame(p)
		return ok && math.Abs(u) <= 1.7 && p.Y <= ys && p.Y >= ys-0.9
	}})
	l := B.Sub(A).Length()
	for t := 3.0; t < l-2; t += 4.5 {
		q := A.Add(B.Sub(A).Scale(t / l))
		_, _, ys, _ := frame(q.At(0))
		ground := s.height(q.X, q.Y)
		if ys-0.9-ground > 0.8 {
			sg.Fill(BlockAt(q, 1.4, 1.4, ground-0.6, ys-0.8))
		}
	}
	for _, c := range sg.Cells() {
		if col, _ := sg.ColorOf(c); col == Glass {
			s.set(sg, c, s.iceColor(sg.World(c), nil, 23))
		}
	}
	s.snowCap(sg, func(col Color) bool { return col == Glass || col == Water || col == Navy }, 0.9, 5)
	sg.Use(Water).Fill(Func{box, func(p Vec3) bool {
		_, u, ys, ok := frame(p)
		return ok && math.Abs(u) > 1.7 && math.Abs(u) <= 2.5 && p.Y <= ys+1.1 && p.Y >= ys-0.9
	}})
	g.Merge(sg)
}

// stalls are red-roofed wooden market stalls facing the palace.
func (s IcePalace) stalls(g *Grid) {
	sg := NewGridWithSpan(g.Span())
	for idx, zc := range []float64{-21, -12, -3, 6} {
		x0, x1 := 28.5, 32.5
		base := s.height(30.5, zc)
		sg.Use(Wood)
		sg.Fill(Block(V(x0, base-0.8, zc-2.4), V(x1, base+0.1, zc+2.4)))
		sg.Fill(Block(V(x1-0.8, base, zc-2.4), V(x1, base+3.4, zc+2.4)))
		sg.Fill(Block(V(x0+1.4, base, zc-2.4), V(x1, base+3.4, zc-1.7)), Block(V(x0+1.4, base, zc+1.7), V(x1, base+3.4, zc+2.4)))
		sg.Fill(Cylinder{P(x0+0.4, zc-2.0), 0.45, 0, base, base + 3.4}, Cylinder{P(x0+0.4, zc+2.0), 0.45, 0, base, base + 3.4})
		sg.Fill(Block(V(x0, base, zc-1.7), V(x0+0.9, base+1.3, zc+1.7)))
		goods := []Color{Red, Gold, White}
		for k := -1; k <= 1; k++ {
			sg.Use(goods[(idx+k+3)%3]).Fill(Sphere(V(x0+0.5, base+1.8, zc+float64(k)*1.1), 0.5))
		}
		sg.Use(Red).Fill(Gable{P(x0-0.9, zc-2.9), P(x1+0.7, zc+2.9), base + 3.2, AxisZ})
		sg.Use(Gold).Fill(Sphere(V(x0-0.3, base+2.9, zc), 0.5))
	}
	s.snowCap(sg, func(col Color) bool { return col == Red }, 0.4, 13)
	g.Merge(sg)
}

// scaled grows a local solid by k about the origin and lifts it to base b.
func (IcePalace) scaled(sol Solid, k, b float64) Solid {
	bb := sol.Bounds()
	box := Box{V(bb.Min.X*k, bb.Min.Y*k+b, bb.Min.Z*k), V(bb.Max.X*k, bb.Max.Y*k+b, bb.Max.Z*k)}
	return Func{box, func(p Vec3) bool { return sol.Contains(V(p.X/k, (p.Y-b)/k, p.Z/k)) }}
}

// place turns a local solid (facing +X) to face angle and moves it to c.
func (IcePalace) place(sol Solid, c Vec2, angle float64) Solid {
	return RotateY(Translate(sol, V(c.X, 0, c.Y)), c, angle)
}

// sculptures are a bear on the shore and a swan on the ice.
func (s IcePalace) sculptures(g *Grid) {
	// Bear walking on a snow plinth, in profile to the hero view.
	bc, ba := P(33, 21), 2.46
	ground := s.height(bc.X, bc.Y)
	g.Use(White).Fill(s.place(BlockAt(P(0, 0), 10.4, 5.2, ground-0.8, ground+1.6), bc, ba))
	g.Use(Water).Fill(s.place(Subtract(BlockAt(P(0, 0), 10.4, 5.2, ground+0.8, ground+1.6), BlockAt(P(0, 0), 9.0, 3.8, ground, ground+2)), bc, ba))
	b := ground + 1.6
	bear := func(sol Solid) Solid { return s.place(s.scaled(sol, 1.3, b), bc, ba) }
	g.Use(Glass)
	for _, sol := range []Solid{
		Ellipsoid{V(0, 3.0, 0), V(3.0, 1.9, 1.7)},
		Ellipsoid{V(1.0, 4.0, 0), V(1.4, 1.0, 1.3)},
		Ellipsoid{V(3.3, 3.6, 0), V(1.2, 1.1, 1.0)},
		Ellipsoid{V(4.5, 3.3, 0), V(0.9, 0.6, 0.6)},
		Sphere(V(3.0, 4.6, 0.65), 0.5), Sphere(V(3.0, 4.6, -0.65), 0.5),
	} {
		g.FillVolume(bear(sol))
	}
	for _, lx := range []float64{-1.9, 1.9} {
		for _, lz := range []float64{-0.9, 0.9} {
			g.Fill(bear(Capsule{V(lx, 2.6, lz), V(lx+0.3*lz, 0.3, lz), 0.75}))
		}
	}
	g.Paint(Water, bear(Block(V(-3.5, 0, -2), V(3.5, 2.0, 2))))
	g.Paint(Navy, bear(Sphere(V(5.3, 3.4, 0), 0.45)))

	// Swan with raised wings on a deep-ice pedestal.
	sc, sa := P(-11, 27), 2.46
	g.Use(Water).Fill(s.place(BlockAt(P(0, 0), 6, 3.6, icePalaceL-0.5, icePalaceL+0.9), sc, sa))
	b = icePalaceL + 0.9
	g.Use(White)
	g.FillVolume(s.place(Ellipsoid{V(0, b+1.3, 0), V(2.5, 1.2, 1.4)}, sc, sa))
	g.FillVolume(s.place(Ellipsoid{V(-2.4, b+2.0, 0), V(0.9, 0.7, 0.6)}, sc, sa))
	for _, q := range [][2]Vec3{{V(1.8, b+1.6, 0), V(2.6, b+3.0, 0)}, {V(2.6, b+3.0, 0), V(2.2, b+4.3, 0)}, {V(2.2, b+4.3, 0), V(2.6, b+5.3, 0)}} {
		g.Fill(s.place(Capsule{q[0], q[1], 0.55}, sc, sa))
	}
	g.FillVolume(s.place(Ellipsoid{V(3.0, b+5.4, 0), V(0.85, 0.55, 0.5)}, sc, sa))
	g.Use(Gold).Fill(s.place(Capsule{V(3.6, b+5.3, 0), V(4.3, b+5.0, 0), 0.38}, sc, sa))
	g.Use(Glass)
	for _, wz := range []float64{-1.0, 1.0} {
		g.FillVolume(s.place(Ellipsoid{V(-0.6, b+2.7, wz), V(2.1, 1.2, 0.5)}, sc, sa))
	}
}

// snowmen stand about the shore with red scarves and dark hats.
func (s IcePalace) snowmen(g *Grid) {
	for i, c := range []Vec2{P(24, 17), P(-30, 27), P(36, 13)} {
		base := s.height(c.X, c.Y)
		g.Use(White)
		g.FillVolume(Sphere(c.At(base+1.0), 1.4), Sphere(c.At(base+2.9), 1.05), Sphere(c.At(base+4.4), 0.8))
		g.Use(Red).Fill(Cylinder{c, 1.1, 0, base + 3.6, base + 3.9})
		g.Use(Navy).Fill(Cylinder{c, 0.9, 0, base + 5.0, base + 5.2}, Cylinder{c, 0.6, 0, base + 5.0, base + 6.2})
		dir := Polar(c, 0.9, 0.9+float64(i)*0.3)
		g.Use(Gold).Fill(Sphere(dir.At(base+4.4), 0.4))
	}
}

// lake adds skaters on the ice.
func (s IcePalace) lake(g *Grid) {
	coats := []Color{Red, Gold, Navy, Red, White}
	n := 0
	for i := 0; n < 11 && i < 200; i++ {
		x := icePalaceLakeC.X + (Hash01(i, 31)-0.5)*2*icePalaceLakeA*0.8
		z := icePalaceLakeC.Y + (Hash01(i, 32)-0.5)*2*icePalaceLakeB*0.8
		if s.lakeE(x, z) > 0.75 || P(x, z).Sub(P(-11, 27)).Length() < 6 || math.Abs(x) < 5 && z < 17 {
			continue
		}
		L := icePalaceL
		g.Use(Navy).Fill(Capsule{V(x, L+0.3, z), V(x, L+1.3, z), 0.45})
		g.Use(coats[n%len(coats)]).Fill(Capsule{V(x, L+1.6, z), V(x, L+2.4, z), 0.6})
		g.Use(White).Fill(Sphere(V(x, L+3.1, z), 0.5))
		n++
	}
}

// lanterns line the shore behind the rink and the steps to the gate.
func (s IcePalace) lanterns(g *Grid) {
	k := 0
	for a := math.Pi + 0.25; a < 2*math.Pi-0.2; a += 0.3 {
		x := icePalaceLakeC.X + 1.12*icePalaceLakeA*math.Cos(a)
		z := icePalaceLakeC.Y + 1.12*icePalaceLakeB*math.Sin(a)
		if math.Abs(x) < 6 || x < -24 && z < 18 {
			continue
		}
		base := s.height(x, z)
		g.Use(Wood).Fill(Capsule{V(x, base-0.4, z), V(x, base+3.2, z), 0.45})
		lamp := Gold
		if k%2 == 1 {
			lamp = Red
		}
		g.Use(lamp).Fill(Sphere(V(x, base+3.7, z), 0.65))
		g.Use(White).Fill(Sphere(V(x, base+4.4, z), 0.45))
		k++
	}
	for _, sx := range []float64{-1, 1} {
		g.Use(Wood).Fill(Capsule{V(sx*5.8, icePalaceG-0.4, 12.6), V(sx*5.8, icePalaceG+3.4, 12.6), 0.45})
		g.Use(Gold).Fill(Sphere(V(sx*5.8, icePalaceG+3.9, 12.6), 0.65))
	}
}
