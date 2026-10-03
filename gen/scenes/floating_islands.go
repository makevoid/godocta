package scenes

import (
	"math"

	"godocta/gen/kit"
	. "godocta/gen/octa"
)

// FloatingIslands is a fantasy sky archipelago: craggy inverted rock cones with grassy
// tops float at different heights. The largest carries a white round temple with a gold
// dome and a tall white tower; the others hold a red pagoda with a torii, a windmill
// over striped fields, a striped lighthouse crowned by a floating crystal (the lowest
// island, whose tip touches the ground), a ruined arch on the highest rock, and a dock
// where a red-and-gold airship is moored. Rope and chain bridges and floating stepping
// stones link the islands; waterfalls pour off their rims into the void and roots and
// vines dangle beneath. +Z is the front.
type FloatingIslands struct{}

func init() { Register(FloatingIslands{}) }

func (FloatingIslands) Name() string  { return "floating_islands" }
func (FloatingIslands) Title() string { return "Floating Islands" }

// floatingIslandsIsle is one floating island: a domed grassy top over an inverted,
// ridged rock cone whose tip lies depth below the rim.
type floatingIslandsIsle struct {
	c                Vec2
	r, top, depth    float64
	seed             int
	roots, crystals  int
	grassy, hollowed bool
}

// floatingIslandsNoise2 is smooth 2D value noise in [0, 1).
func floatingIslandsNoise2(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

// floatingIslandsNoise3 is smooth 3D value noise in [0, 1).
func floatingIslandsNoise3(x, y, z float64, seed int) float64 {
	ix, iy, iz := math.Floor(x), math.Floor(y), math.Floor(z)
	fx, fy, fz := x-ix, y-iy, z-iz
	sx, sy, sz := fx*fx*(3-2*fx), fy*fy*(3-2*fy), fz*fz*(3-2*fz)
	i, j, k := int(ix), int(iy), int(iz)
	lerp := func(a, b, t float64) float64 { return a + (b-a)*t }
	h := func(a, b, c int) float64 { return Hash01(a, b, c, seed) }
	x00 := lerp(h(i, j, k), h(i+1, j, k), sx)
	x10 := lerp(h(i, j+1, k), h(i+1, j+1, k), sx)
	x01 := lerp(h(i, j, k+1), h(i+1, j, k+1), sx)
	x11 := lerp(h(i, j+1, k+1), h(i+1, j+1, k+1), sx)
	return lerp(lerp(x00, x10, sy), lerp(x01, x11, sy), sz)
}

// radius is the rim radius in a plan direction.
func (s floatingIslandsIsle) radius(theta float64) float64 {
	cs, sn := math.Cos(theta), math.Sin(theta)
	n := floatingIslandsNoise2(1.5*cs+10, 1.5*sn+10, s.seed)
	n2 := floatingIslandsNoise2(4*cs+20, 4*sn+20, s.seed+1)
	return s.r * (0.88 + 0.22*n + 0.08*n2)
}

// topAt is the grass surface height at a plan point inside the rim.
func (s floatingIslandsIsle) topAt(x, z float64) float64 {
	d := P(x, z).Sub(s.c)
	k := d.Length() / s.radius(d.Angle())
	if k > 1 {
		return s.top
	}
	bump := floatingIslandsNoise2(x/5, z/5, s.seed+2)
	return s.top + (1-k*k)*(0.8+s.r*0.05+1.2*bump)
}

// crag is the ridged radius factor of the underside at a direction and height.
func (s floatingIslandsIsle) crag(theta, y float64) float64 {
	cs, sn := math.Cos(theta), math.Sin(theta)
	f := s.r / 3.5
	n := floatingIslandsNoise3(cs*f+50, y/6, sn*f+50, s.seed+3)
	n2 := floatingIslandsNoise3(cs*f*2.3+80, y/3, sn*f*2.3+80, s.seed+4)
	return 0.66 + 0.42*n + 0.16*n2
}

// underRadius is the rock's radius at height y below the rim (0 below the tip).
func (s floatingIslandsIsle) underRadius(theta, y float64) float64 {
	t := (y - (s.top - s.depth)) / s.depth
	if t <= 0 {
		return 0
	}
	if t > 1 {
		t = 1
	}
	crag := s.crag(theta, y)
	lip := math.Max(0, math.Min(1, (t-0.82)/0.18))
	crag += (1 - crag) * lip * lip
	return s.radius(theta) * math.Pow(t, 1.1) * crag
}

func (s floatingIslandsIsle) contains(p Vec3) bool {
	d := p.Plan().Sub(s.c)
	dist, theta := d.Length(), d.Angle()
	if p.Y >= s.top {
		return dist <= s.radius(theta) && p.Y <= s.topAt(p.X, p.Z)
	}
	return dist <= s.underRadius(theta, p.Y)
}

func (s floatingIslandsIsle) bounds() Box {
	r := s.r * 1.2
	return Box{V(s.c.X-r, s.top-s.depth, s.c.Y-r), V(s.c.X+r, s.top+3+s.r*0.05, s.c.Y+r)}
}

// rim is the plan point at a fraction of the rim radius in a direction.
func (s floatingIslandsIsle) rim(theta, k float64) Vec2 {
	return Polar(s.c, s.radius(theta)*k, theta)
}

// build fills the island as a hollow shell and colors its grass, soil and rock strata,
// moss, dangling roots and vines, and underside crystals.
func (s floatingIslandsIsle) build(g *Grid) {
	box := s.bounds()
	const w = 1.3
	offsets := []Vec3{V(w, 0, 0), V(-w, 0, 0), V(0, w, 0), V(0, -w, 0), V(0, 0, w), V(0, 0, -w)}
	shell := Func{box, func(p Vec3) bool {
		if !s.contains(p) {
			return false
		}
		if !s.hollowed {
			return true
		}
		for _, o := range offsets {
			if !s.contains(p.Add(o)) {
				return true
			}
		}
		return false
	}}
	g.Use(DarkStone).Fill(shell)
	// Rock strata: bands tilted by noise.
	band := func(p Vec3) int {
		return int(math.Floor(p.Y/1.3 + 1.6*floatingIslandsNoise2(p.X/9, p.Z/9, s.seed+5)))
	}
	g.Paint(Stone, Func{box, func(p Vec3) bool { return s.contains(p) && Hash01(band(p), s.seed, 1) < 0.35 }})
	g.Paint(Brick, Func{box, func(p Vec3) bool {
		h := Hash01(band(p), s.seed, 1)
		return s.contains(p) && h >= 0.35 && h < 0.5
	}})
	g.Paint(Terracotta, Func{box, func(p Vec3) bool {
		h := Hash01(band(p), s.seed, 1)
		return s.contains(p) && h >= 0.5 && h < 0.56
	}})
	// Moss on the upper underside.
	g.Paint(Foliage, Func{box, func(p Vec3) bool {
		return s.contains(p) && p.Y < s.top && p.Y > s.top-s.depth*0.45 &&
			floatingIslandsNoise3(p.X/2.5, p.Y/2.5, p.Z/2.5, s.seed+6) > 0.7
	}})
	// Soil band under the turf, then grass.
	g.Paint(Wood, Func{box, func(p Vec3) bool {
		return s.contains(p) && p.Y >= s.top-1.2 && p.Y < s.topAt(p.X, p.Z)-0.9
	}})
	if s.grassy {
		g.Paint(Grass, Func{box, func(p Vec3) bool {
			return s.contains(p) && p.Y >= s.topAt(p.X, p.Z)-0.9
		}})
		// Flower meadows.
		if s.r >= 6 {
			flowers := []Color{Red, Gold, White, Gold}
			for i, color := range flowers {
				g.Paint(color, Func{box, func(p Vec3) bool {
					if !s.contains(p) || p.Y < s.topAt(p.X, p.Z)-0.6 {
						return false
					}
					if floatingIslandsNoise2(p.X/3, p.Z/3, s.seed+60) < 0.62 {
						return false
					}
					cell := Nearest(p.Scale(1.4))
					return int(Hash01(int(cell.X), int(cell.Z), s.seed, 61)*float64(len(flowers)*3)) == i
				}})
			}
		}
	}
	// Roots and vines hanging off the underside.
	for k := 0; k < s.roots; k++ {
		theta := (float64(k) + 0.8*Hash01(k, s.seed, 7)) * 2 * math.Pi / float64(s.roots)
		u := 0.08 + 0.4*Hash01(k, s.seed, 8)
		y0 := s.top - s.depth*u
		r0 := s.underRadius(theta, y0) - 0.4
		if r0 < 1 {
			continue
		}
		a := Polar(s.c, r0, theta).At(y0)
		length := (2.5 + 7*Hash01(k, s.seed, 9)) * math.Min(1, s.r/7)
		sway := Polar(P(0, 0), 0.6+0.8*Hash01(k, s.seed, 10), theta+1.2*(Hash01(k, s.seed, 11)-0.5))
		b := a.Add(V(sway.X, -length*0.5, sway.Y))
		c := b.Add(V(-sway.X*0.4, -length*0.5, -sway.Y*0.4))
		color := Foliage
		if Hash01(k, s.seed, 12) < 0.35 {
			color = Wood
		}
		g.With(color, func() { g.Fill(Capsule{a, b, 0.55}, Capsule{b, c, 0.5}) })
		if color == Foliage && Hash01(k, s.seed, 13) < 0.4 {
			g.With(Gold, func() { g.Fill(Sphere(c, 0.7)) })
		}
	}
	// Glowing crystal shards poking out of the underside.
	for k := 0; k < s.crystals; k++ {
		theta := (float64(k) + Hash01(k, s.seed, 14)) * 2 * math.Pi / float64(s.crystals)
		y0 := s.top - s.depth*(0.3+0.35*Hash01(k, s.seed, 15))
		r0 := s.underRadius(theta, y0) - 0.8
		if r0 < 1 {
			continue
		}
		a := Polar(s.c, r0, theta).At(y0)
		tip := Polar(s.c, r0+2.2+Hash01(k, s.seed, 16), theta).At(y0 - 2.2 - Hash01(k, s.seed, 17))
		g.With(Glass, func() { g.Fill(Capsule{a, tip, 0.75}, Capsule{a, a.Add(tip.Sub(a).Scale(0.5)), 1.0}) })
	}
}

// waterfall paints a stream from a pond to the rim, then a falling sheet of water that
// curves outward, widens and breaks into droplets.
func (s floatingIslandsIsle) waterfall(g *Grid, theta, width, length, pond float64) {
	dir := Polar(P(0, 0), 1, theta)
	tan := P(-dir.Y, dir.X)
	r0 := s.radius(theta)
	box := s.bounds()
	pondAt := s.rim(theta, pond)
	// Pond and meandering stream on the top surface.
	g.Paint(Water, Func{box, func(p Vec3) bool {
		if !s.contains(p) || p.Y < s.topAt(p.X, p.Z)-1.0 {
			return false
		}
		q := p.Plan().Sub(s.c)
		along, across := q.Dot(dir), q.Dot(tan)
		across -= 0.8 * math.Sin(along*0.7)
		if p.Plan().Sub(pondAt).Length() < 2.6 {
			return true
		}
		return along > pond*r0 && math.Abs(across) <= width/2
	}})
	rimY := s.top
	fall := Func{Box{V(s.c.X-s.r*2, rimY-length-1, s.c.Y-s.r*2), V(s.c.X+s.r*2, rimY+1.5, s.c.Y+s.r*2)}, func(p Vec3) bool {
		q := p.Plan().Sub(s.c)
		along, across := q.Dot(dir)-r0, q.Dot(tan)
		dy := rimY + 0.6 - p.Y
		if dy < 0 || dy > length {
			return false
		}
		off := -0.8 + 0.008*dy*dy
		half := width/2 + 0.06*dy
		if math.Abs(across) > half || along < off || along > off+1.3 {
			return false
		}
		if dy > 0.65*length {
			fade := (dy - 0.65*length) / (0.35 * length)
			cell := Nearest(p.Scale(2))
			return Hash01(int(cell.X), int(cell.Y), int(cell.Z), s.seed, 20) > fade
		}
		return true
	}}
	g.With(Water, func() { g.Fill(fall) })
	// Foam streaks and the white lip.
	g.Paint(White, Func{fall.Box, func(p Vec3) bool {
		if !fall.F(p) {
			return false
		}
		q := p.Plan().Sub(s.c)
		across := q.Dot(tan)
		dy := rimY + 0.6 - p.Y
		streak := Hash01(int(math.Floor(across/1.1)), int(math.Floor(dy/4)), s.seed, 21) < 0.3
		return dy < 1.6 || streak || dy > 0.6*length && Hash01(int(math.Floor(p.X*3)), int(math.Floor(p.Y*3)), int(math.Floor(p.Z*3)), 22) < 0.4
	}})
}

// trees scatters round, conifer and colorful magic trees over an island top, skipping
// the clear circles (x, z, radius).
func (FloatingIslands) trees(g *Grid, s floatingIslandsIsle, count int, clear [][3]float64) {
	var placed []Vec2
	for k := 0; k < count*6 && len(placed) < count; k++ {
		theta := Hash01(k, s.seed, 30) * 2 * math.Pi
		at := s.rim(theta, 0.25+0.62*math.Sqrt(Hash01(k, s.seed, 31)))
		ok := true
		for _, c := range clear {
			if at.Sub(P(c[0], c[1])).Length() < c[2] {
				ok = false
			}
		}
		for _, q := range placed {
			if at.Sub(q).Length() < 3.4 {
				ok = false
			}
		}
		if !ok {
			continue
		}
		placed = append(placed, at)
		base := s.topAt(at.X, at.Y) - 0.4
		h := 4.5 + 3.5*Hash01(k, s.seed, 32)
		kind := Hash01(k, s.seed, 33)
		g.With(Wood, func() { g.Fill(Cylinder{Center: at, Radius: 0.6, Y0: base, Y1: base + h*0.5}) })
		if kind < 0.35 {
			// Conifer: stacked cones.
			g.With(Foliage, func() {
				for i := 0; i < 3; i++ {
					y := base + h*(0.3+0.22*float64(i))
					rad := h * (0.32 - 0.07*float64(i))
					g.FillVolume(Cone(at, rad, y, h*0.42))
				}
			})
			continue
		}
		colors := []Color{Foliage, Grass, Foliage, Gold, Red, Terracotta}
		color := colors[int(Hash01(k, s.seed, 34)*float64(len(colors)))]
		g.With(color, func() {
			g.FillVolume(Ellipsoid{at.At(base + h*0.68), V(h*0.34, h*0.3, h*0.34)})
		})
	}
}

// bridge spans two points with a sagging plank deck, side ropes or chains and posts.
func (FloatingIslands) bridge(g *Grid, a, b Vec3, width, sag float64, rail Color) {
	ab := b.Plan().Sub(a.Plan())
	length := ab.Length()
	dir := ab.Scale(1 / length)
	side := P(-dir.Y, dir.X)
	deckY := func(t float64) float64 { return a.Y + (b.Y-a.Y)*t - sag*4*t*(1-t) }
	box := BoxOf(a, b).Grow(width + 1)
	box.Min.Y -= sag + 1
	box.Max.Y += 3
	deck := Func{box, func(p Vec3) bool {
		q := p.Plan().Sub(a.Plan())
		t := q.Dot(dir) / length
		if t < 0 || t > 1 || math.Abs(q.Dot(side)) > width/2 {
			return false
		}
		y := deckY(t)
		return p.Y <= y && p.Y >= y-0.9
	}}
	g.With(Wood, func() { g.Fill(deck) })
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		return int(math.Floor(p.Plan().Sub(a.Plan()).Dot(dir)/1.2))%3 == 0 && deck.F(p)
	}})
	n := int(length / 0.8)
	point := func(t, offset, lift float64) Vec3 {
		return a.Plan().Add(dir.Scale(t * length)).Add(side.Scale(offset)).At(deckY(t) + lift - sag*0.25*4*t*(1-t))
	}
	g.With(rail, func() {
		for _, o := range []float64{-width/2 - 0.2, width/2 + 0.2} {
			for i := 0; i < n; i++ {
				t0, t1 := float64(i)/float64(n), float64(i+1)/float64(n)
				g.Fill(Capsule{point(t0, o, 2.0), point(t1, o, 2.0), 0.5})
			}
			posts := int(length/3.2) + 1
			for i := 0; i <= posts; i++ {
				t := float64(i) / float64(posts)
				top := point(t, o, 2.0)
				foot := a.Plan().Add(dir.Scale(t * length)).Add(side.Scale(o)).At(deckY(t) - 0.4)
				g.Fill(Capsule{foot, top, 0.45})
			}
		}
	})
	// Heavy anchor posts at both ends.
	g.With(Wood, func() {
		for _, t := range []float64{0, 1} {
			for _, o := range []float64{-width/2 - 0.4, width/2 + 0.4} {
				at := a.Plan().Add(dir.Scale(t * length)).Add(side.Scale(o))
				g.Fill(Cylinder{Center: at, Radius: 0.7, Y0: deckY(t) - 1.5, Y1: deckY(t) + 3})
			}
		}
	})
}

// radialArch is an arched opening through a round wall at angle a.
func (FloatingIslands) radialArch(center Vec2, radius, a, sill, half, spring float64) Solid {
	arch := Arch{Sill: V(radius, sill, 0), HalfWidth: half, Spring: spring, Depth: 3.2, Through: AxisX}
	return Translate(RotateY(arch, P(0, 0), a), center.At(0))
}

// temple is the white round temple with a colonnade and a gold dome, and the tall white
// tower behind it, on the main island.
func (f FloatingIslands) temple(g *Grid, c Vec2, tower Vec2, ground float64) {
	L := ground + 4.6
	// Stepped hollow plinth.
	g.Use(Stone)
	for i, r := range []float64{10.6, 9.6, 8.6} {
		top := L - 1.6 + 0.8*float64(i)
		g.Fill(Cylinder{Center: c, Radius: r, Inner: r - 1.3, Y0: ground - 2, Y1: top})
		g.Fill(Cylinder{Center: c, Radius: r, Y0: top - 0.9, Y1: top})
	}
	g.Paint(White, Cylinder{Center: c, Radius: 8.7, Y0: L - 0.5, Y1: L})
	// Colonnade.
	for k := 0; k < 12; k++ {
		at := Polar(c, 7.2, float64(k)*math.Pi/6+math.Pi/12)
		g.With(White, func() { g.Fill(Cylinder{Center: at, Radius: 0.75, Y0: L, Y1: L + 9}) })
		g.With(Gold, func() {
			g.Fill(Cylinder{Center: at, Radius: 1.15, Y0: L, Y1: L + 0.6})
			g.Fill(Cylinder{Center: at, Radius: 1.15, Y0: L + 8.4, Y1: L + 9})
		})
	}
	// Cella with door and windows.
	g.With(White, func() {
		g.Fill(Cylinder{Center: c, Radius: 4.8, Inner: 3.6, Y0: L, Y1: L + 9})
		g.Fill(Cylinder{Center: c, Radius: 8.2, Inner: 3.6, Y0: L + 9, Y1: L + 10.6})
		g.Fill(Cylinder{Center: c, Radius: 5.6, Inner: 4.3, Y0: L + 10.6, Y1: L + 13.2})
	})
	g.Paint(Gold, Cylinder{Center: c, Radius: 8.3, Inner: 7.3, Y0: L + 9.6, Y1: L + 10.0})
	g.Carve(f.radialArch(c, 4.2, math.Pi/2, L, 1.4, 3.6))
	for k := 0; k < 4; k++ {
		g.Carve(f.radialArch(c, 4.2, math.Pi/2+float64(k+1)*2*math.Pi/5, L+2.5, 0.8, 2.4))
	}
	for k := 0; k < 8; k++ {
		g.Carve(f.radialArch(c, 4.9, float64(k)*math.Pi/4, L+11, 0.6, 0.8))
	}
	// Gold dome with white ribs, lantern and spire.
	domeY := L + 13.2
	dome := Intersect(
		Subtract(Ellipsoid{c.At(domeY), V(6, 6.4, 6)}, Ellipsoid{c.At(domeY), V(4.8, 5.2, 4.8)}),
		Block(V(c.X-7, domeY, c.Y-7), V(c.X+7, domeY+7, c.Y+7)))
	g.With(Gold, func() { g.FillVolume(dome) })
	g.Paint(White, Func{dome.Bounds(), func(p Vec3) bool {
		q := p.Plan().Sub(c)
		a := math.Mod(q.Angle()+2*math.Pi, math.Pi/4)
		return math.Min(a, math.Pi/4-a)*q.Length() < 0.45 && p.Y < domeY+5.2
	}})
	g.With(White, func() { g.Fill(Cylinder{Center: c, Radius: 1.4, Y0: domeY + 5.4, Y1: domeY + 8}) })
	g.With(Gold, func() {
		g.FillVolume(Sphere(c.At(domeY+8.6), 1.6))
		g.Fill(Capsule{c.At(domeY + 9), c.At(domeY + 13), 0.55})
	})

	// The tall white tower.
	t0, t1 := ground-1.5, L+27.0
	g.With(White, func() {
		g.Fill(Cylinder{Center: tower, Radius: 4.4, Inner: 3.1, Y0: t0, Y1: L + 1})
		g.Fill(Cylinder{Center: tower, Radius: 3.3, Inner: 2.0, Y0: L, Y1: t1})
		// Balcony.
		g.Fill(Cylinder{Center: tower, Radius: 4.6, Inner: 2.0, Y0: t1 - 6, Y1: t1 - 5.2})
		g.Fill(Cylinder{Center: tower, Radius: 4.6, Inner: 3.6, Y0: t1 - 5.2, Y1: t1 - 4.2})
	})
	g.Carve(f.radialArch(tower, 4.0, math.Pi/2-0.3, ground+1.5, 1.2, 2.6))
	for row := 0; row < 6; row++ {
		y := L + 2 + float64(row)*3.4
		for k := 0; k < 3; k++ {
			a := float64(k)*2*math.Pi/3 + float64(row)*0.6
			g.Carve(f.radialArch(tower, 3.0, a, y, 0.65, 1.2))
		}
	}
	for k := 0; k < 6; k++ {
		g.Carve(f.radialArch(tower, 3.0, float64(k)*math.Pi/3, t1-4.0, 0.8, 1.6))
		g.Carve(f.radialArch(tower, 4.3, float64(k)*math.Pi/3+math.Pi/6, t1-5.2, 0.7, 0.2))
	}
	for _, y := range []float64{L + 0.2, L + 8.6, L + 15.4, t1 - 5.8, t1 - 0.6} {
		g.Paint(Gold, Cylinder{Center: tower, Radius: 4.7, Inner: 2.4, Y0: y, Y1: y + 0.6})
	}
	g.With(Gold, func() {
		g.Fill(Cylinder{Center: tower, Radius: 3.8, Inner: 2.0, Y0: t1, Y1: t1 + 0.8})
		onion := Func{Box{V(tower.X-4.2, t1, tower.Y-4.2), V(tower.X+4.2, t1+9, tower.Y+4.2)}, func(p Vec3) bool {
			h := p.Y - t1
			var r float64
			if h < 3.5 {
				r = 3.0 + 1.1*math.Sin(h/3.5*math.Pi/2)
			} else {
				r = 4.1 * math.Pow(1-(h-3.5)/5.5, 1.4)
			}
			d := p.Plan().Sub(tower).Length()
			return d <= r && (d >= r-1.3 || h > 6.5)
		}}
		g.FillVolume(onion)
		g.Fill(Capsule{tower.At(t1 + 8.5), tower.At(t1 + 13), 0.55})
		g.Fill(Sphere(tower.At(t1+11), 0.9))
	})
}

// pagoda is a five-tier red pagoda with slate roofs and a gold spire, plus a torii.
func (FloatingIslands) pagoda(g *Grid, c Vec2, ground float64, torii Vec2, toriiAngle float64) {
	L := ground + 3.4
	g.With(Stone, func() {
		g.Fill(Subtract(BlockAt(c, 11, 11, ground-2, L), BlockAt(c, 8.6, 8.6, ground-2, L-1)))
	})
	y := L
	for i := 0; i < 5; i++ {
		hs := 3.9 - 0.5*float64(i)
		h := 3.2 - 0.2*float64(i)
		g.With(Red, func() {
			g.Fill(Subtract(BlockAt(c, 2*hs, 2*hs, y, y+h), BlockAt(c, 2*hs-2.2, 2*hs-2.2, y, y+h)))
		})
		// Doors / windows on each face.
		for _, axis := range []Axis{AxisX, AxisZ} {
			sill := y
			if i > 0 {
				sill = y + 0.6
			}
			g.Carve(Arch{Sill: c.At(sill), HalfWidth: 0.9, Spring: h - 2.0, Depth: 2*hs + 2, Through: axis})
		}
		g.Paint(White, Subtract(BlockAt(c, 2*hs, 2*hs, y+h-0.9, y+h), BlockAt(c, 2*hs-0.8, 2*hs-0.8, y, y+h)))
		// Flared roof.
		ry := y + h
		ho := hs + 2.6 - 0.15*float64(i)
		roof := Func{Box{V(c.X-ho, ry, c.Y-ho), V(c.X+ho, ry+2.2, c.Y+ho)}, func(p Vec3) bool {
			q := p.Plan().Sub(c)
			m := math.Max(math.Abs(q.X), math.Abs(q.Y))
			e := ho - m // Distance in from the eave.
			if e < 0 {
				return false
			}
			surface := ry + 0.6 + math.Min(e, 2.0)*0.55 + math.Max(0, e-2.0)*0.9
			return p.Y <= surface && p.Y >= surface-1.1
		}}
		g.With(Slate, func() { g.Fill(roof) })
		g.With(Gold, func() {
			for _, sx := range []float64{-1, 1} {
				for _, sz := range []float64{-1, 1} {
					a := c.Add(P(sx*(ho-0.6), sz*(ho-0.6))).At(ry + 0.8)
					b := c.Add(P(sx*(ho+0.4), sz*(ho+0.4))).At(ry + 2.0)
					g.Fill(Capsule{a, b, 0.5})
				}
			}
		})
		y = ry + 1.3
	}
	g.With(Gold, func() {
		g.Fill(Cylinder{Center: c, Radius: 0.55, Y0: y, Y1: y + 8})
		for k := 0; k < 5; k++ {
			ry := y + 1.2 + float64(k)*1.2
			g.Fill(Cylinder{Center: c, Radius: 1.3 - 0.1*float64(k), Y0: ry, Y1: ry + 0.5})
		}
		g.FillVolume(Sphere(c.At(y+8.4), 0.9))
	})
	// Torii gate.
	tor := func(s Solid) Solid { return Translate(RotateY(s, P(0, 0), toriiAngle), torii.At(0)) }
	tg := ground - 1
	g.With(Red, func() {
		g.Fill(tor(Cylinder{Center: P(-2.6, 0), Radius: 0.6, Y0: tg, Y1: tg + 7.5}))
		g.Fill(tor(Cylinder{Center: P(2.6, 0), Radius: 0.6, Y0: tg, Y1: tg + 7.5}))
		g.Fill(tor(Block(V(-3.6, tg+5.6, -0.4), V(3.6, tg+6.3, 0.4))))
	})
	g.With(Charcoal, func() {
		g.Fill(tor(Func{Box{V(-4.8, tg+7.3, -0.6), V(4.8, tg+9.5, 0.6)}, func(p Vec3) bool {
			base := tg + 7.4 + 0.045*p.X*p.X
			return p.Y >= base && p.Y <= base+0.9
		}}))
	})
}

// windmill is a white tower windmill with a terracotta cap and four lattice sails
// facing face (a plan unit vector), over striped fields.
func (FloatingIslands) windmill(g *Grid, c Vec2, ground float64, face Vec2) {
	top := ground + 14
	g.With(Brick, func() {
		g.Fill(Subtract(Frustum{c, 4.4, 4.1, ground - 1.5, ground + 2.5}, Frustum{c, 3.2, 2.9, ground - 1.5, ground + 2.5}))
	})
	g.With(White, func() {
		g.Fill(Subtract(Frustum{c, 4.1, 2.8, ground + 2.5, top}, Frustum{c, 2.9, 1.6, ground + 2.5, top}))
	})
	// Door and windows.
	at := func(a, y, half, spring float64) Solid {
		arch := Arch{Sill: V(3.6, y, 0), HalfWidth: half, Spring: spring, Depth: 3, Through: AxisX}
		return Translate(RotateY(arch, P(0, 0), a), c.At(0))
	}
	fa := face.Angle()
	g.Carve(at(fa+0.9, ground+0.2, 1.1, 2.2))
	g.Carve(at(fa+2.6, ground+6, 0.6, 1.0), at(fa-1.8, ground+9.5, 0.6, 1.0), at(fa+0.4, ground+11, 0.6, 0.8))
	// Wooden gallery.
	g.With(Wood, func() {
		g.Fill(Cylinder{Center: c, Radius: 5.4, Inner: 3.0, Y0: ground + 6.2, Y1: ground + 6.9})
		for k := 0; k < 10; k++ {
			p := Polar(c, 5.0, float64(k)*math.Pi/5)
			g.Fill(Capsule{p.At(ground + 6.9), p.At(ground + 7.9), 0.45})
		}
		g.Fill(Cylinder{Center: c, Radius: 5.4, Inner: 4.5, Y0: ground + 7.9, Y1: ground + 8.4})
	})
	// Cap.
	g.With(Terracotta, func() { g.FillVolume(Ellipsoid{c.At(top), V(3.4, 3.6, 3.4)}) })
	g.Carve(Cylinder{Center: c, Radius: 1.6, Y0: top - 1, Y1: top + 2})
	// Hub and sails.
	f3 := V(face.X, 0, face.Y)
	hub := c.Add(face.Scale(3.4)).At(top + 0.8)
	g.With(Charcoal, func() {
		g.Fill(Capsule{c.At(top + 0.8), hub.Add(f3.Scale(0.8)), 0.8})
	})
	tan := V(-face.Y, 0, face.X)
	up := V(0, 1, 0)
	for k := 0; k < 4; k++ {
		phi := math.Pi/4 + float64(k)*math.Pi/2 + 0.2
		d := tan.Scale(math.Cos(phi)).Add(up.Scale(math.Sin(phi)))
		perp := V(f3.Y*d.Z-f3.Z*d.Y, f3.Z*d.X-f3.X*d.Z, f3.X*d.Y-f3.Y*d.X)
		g.With(Wood, func() { g.Fill(Capsule{hub, hub.Add(d.Scale(10.5)), 0.5}) })
		sail := Func{BoxOf(hub, hub.Add(d.Scale(11))).Grow(3.5), func(p Vec3) bool {
			r := p.Sub(hub)
			n := r.Dot(f3)
			s, w := r.Dot(d), r.Dot(perp)
			return math.Abs(n-0.2) <= 0.55 && s >= 2.2 && s <= 10.3 && w >= 0.4 && w <= 2.9
		}}
		g.With(White, func() { g.Fill(sail) })
		g.Paint(Wood, Func{sail.Box, func(p Vec3) bool {
			r := p.Sub(hub)
			s, w := r.Dot(d), r.Dot(perp)
			return sail.F(p) && (math.Mod(s-2.2, 2.0) < 0.55 || w > 2.4)
		}})
	}
}

// lighthouse is a red-striped white tower carrying a floating glass crystal.
func (FloatingIslands) lighthouse(g *Grid, c Vec2, ground float64) {
	top := ground + 17
	g.With(White, func() {
		g.Fill(Subtract(Frustum{c, 3.4, 2.4, ground - 1.5, top}, Frustum{c, 2.2, 1.2, ground - 1.5, top}))
	})
	g.Paint(Red, Func{Box{V(c.X-4, ground, c.Y-4), V(c.X+4, top, c.Y+4)}, func(p Vec3) bool {
		return int(math.Floor((p.Y-ground)/2.2))%2 == 1
	}})
	g.Carve(Arch{Sill: c.At(ground - 0.2).Add(V(0, 0, 3)), HalfWidth: 1.0, Spring: 2, Depth: 2.5, Through: AxisZ})
	g.With(Charcoal, func() {
		g.Fill(Cylinder{Center: c, Radius: 3.8, Y0: top, Y1: top + 0.8})
		g.Fill(Cylinder{Center: c, Radius: 3.8, Inner: 3.0, Y0: top + 0.8, Y1: top + 1.8})
	})
	g.With(Gold, func() {
		g.Fill(Cylinder{Center: c, Radius: 1.8, Y0: top + 0.8, Y1: top + 1.6})
		for k := 0; k < 6; k++ {
			p := Polar(c, 1.9, float64(k)*math.Pi/3)
			g.Fill(Capsule{p.At(top + 1.6), Polar(c, 2.6, float64(k)*math.Pi/3).At(top + 4.5), 0.45})
		}
	})
	crystal := func(center Vec3, rad, half float64, angle float64) Solid {
		return Func{Box{center.Sub(V(rad*1.5, half, rad*1.5)), center.Add(V(rad*1.5, half, rad*1.5))}, func(p Vec3) bool {
			q := p.Sub(center)
			pl := P(q.X, q.Z).Rotate(-angle)
			return (math.Abs(pl.X)+math.Abs(pl.Y))/rad+math.Abs(q.Y)/half <= 1
		}}
	}
	g.With(Glass, func() {
		g.FillVolume(crystal(c.At(top+7.6), 2.6, 5.4, math.Pi/4))
		for k := 0; k < 4; k++ {
			p := Polar(c, 5.2, float64(k)*math.Pi/2+0.4)
			g.FillVolume(crystal(p.At(top+5.5+float64(k%2)*2.5), 1.0, 1.9, 0))
		}
	})
}

// ruins is a broken arch, a ring of broken columns and fallen drums on the highest rock.
func (FloatingIslands) ruins(g *Grid, c Vec2, ground float64) {
	L := ground + 1.2
	g.With(Sand, func() {
		g.Fill(Func{Box{V(c.X-6, ground-1.5, c.Y-6), V(c.X+6, L, c.Y+6)}, func(p Vec3) bool {
			q := p.Plan().Sub(c)
			return q.Length() < 5.6 && Hash01(int(math.Floor(q.X/1.6)), int(math.Floor(q.Y/1.6)), 40) < 0.75 && p.Y > L-1
		}})
	})
	arch := Subtract(Block(V(-4.8, L, -0.9), V(4.8, L+11, 0.9)),
		Arch{Sill: V(0, L, 0), HalfWidth: 3.0, Spring: 6.0, Depth: 3, Through: AxisZ})
	broken := Subtract(arch, Sphere(V(3.6, L+11.5, 0), 2.6), Block(V(-4.8, L+10.2, -1), V(-2, L+12, 1)))
	g.With(Stone, func() { g.Fill(Translate(RotateY(broken, P(0, 0), 0.35), c.Add(P(0, -1.5)).At(0))) })
	for k := 0; k < 9; k++ {
		a := float64(k) * 2 * math.Pi / 9
		at := Polar(c.Add(P(0, 1)), 4.6, a)
		if math.Sin(a) < -0.5 {
			continue
		}
		h := 1.5 + 6*Hash01(k, 41)
		g.With(White, func() { g.Fill(Cylinder{Center: at, Radius: 0.75, Y0: ground - 1, Y1: L + h}) })
	}
	g.With(White, func() {
		g.Fill(Capsule{c.Add(P(-3.5, 4.5)).At(L + 0.6), c.Add(P(-0.5, 6.2)).At(L + 0.6), 0.7})
		g.Fill(Capsule{c.Add(P(3, -4.8)).At(L + 0.6), c.Add(P(5.2, -2.6)).At(L + 0.6), 0.7})
	})
	// Ivy.
	g.Paint(Foliage, Func{Box{V(c.X-8, L, c.Y-8), V(c.X+8, L+12, c.Y+8)}, func(p Vec3) bool {
		return floatingIslandsNoise3(p.X/1.8, p.Y/1.8, p.Z/1.8, 42) > 0.68
	}})
}

// airship is a red balloon envelope with gold bands and fins over a wooden gondola.
func (FloatingIslands) airship(g *Grid, gc Vec3, length float64) {
	ec := gc.Add(V(-1, 8.4, 0))
	ra := V(length/2, 4.8, 4.8)
	env := Subtract(Ellipsoid{ec, ra}, Ellipsoid{ec, ra.Sub(V(1.3, 1.3, 1.3))})
	g.With(Red, func() { g.FillVolume(env) })
	g.Paint(Gold, Func{env.Bounds(), func(p Vec3) bool {
		x := p.X - ec.X
		band := math.Mod(x+100, 4.4) < 0.9 && math.Abs(x) < ra.X-1
		return env.Contains(p) && (band || math.Abs(x) > ra.X-1.6 || math.Abs(p.Y-ec.Y) < 0.45)
	}})
	// Cruciform tail fins at the -X end.
	tail := ec.X - ra.X*0.62
	fin := func(p Vec3, a, b float64) bool {
		x := tail - p.X
		return x >= 0 && x <= 4.5 && math.Abs(b) <= 0.5 && a >= 2.5 && a <= 7.2-0.6*x
	}
	finBox := Box{V(tail-4.6, ec.Y-7.5, ec.Z-7.5), V(tail, ec.Y+7.5, ec.Z+7.5)}
	g.With(Red, func() {
		g.Fill(Func{finBox, func(p Vec3) bool {
			dy, dz := p.Y-ec.Y, p.Z-ec.Z
			return fin(p, math.Abs(dy), dz) || fin(p, math.Abs(dz), dy)
		}})
	})
	g.Paint(Gold, Func{finBox, func(p Vec3) bool {
		dy, dz := p.Y-ec.Y, p.Z-ec.Z
		x := tail - p.X
		return x > 3.4 || math.Max(math.Abs(dy), math.Abs(dz)) > 6.0-0.6*x
	}})
	// Gondola hull.
	hl := length * 0.42
	hull := Intersect(
		Subtract(Ellipsoid{gc, V(hl, 2.6, 2.4)}, Ellipsoid{gc, V(hl-1.2, 1.6, 1.3)}),
		Block(V(gc.X-hl, gc.Y-3, gc.Z-3), V(gc.X+hl, gc.Y, gc.Z+3)))
	g.With(Wood, func() {
		g.FillVolume(hull)
		g.Fill(Ellipsoid{gc.Add(V(0, -0.5, 0)), V(hl-1, 0.5, 1.4)})
	})
	g.Paint(Gold, Block(V(gc.X-hl, gc.Y-0.5, gc.Z-3), V(gc.X+hl, gc.Y, gc.Z+3)))
	// Cabin.
	g.With(White, func() {
		g.Fill(Subtract(BlockAt(P(gc.X-0.5, gc.Z), 4, 2.6, gc.Y, gc.Y+2.4), BlockAt(P(gc.X-0.5, gc.Z), 2.6, 1.2, gc.Y, gc.Y+2)))
	})
	g.With(Navy, func() { g.Fill(BlockAt(P(gc.X-0.5, gc.Z), 4.6, 3.2, gc.Y+2.4, gc.Y+3.0)) })
	// Rigging.
	g.With(Charcoal, func() {
		for _, x := range []float64{-hl + 1.5, 0, hl - 1.5} {
			for _, z := range []float64{-1.6, 1.6} {
				a := gc.Add(V(x, 0, z))
				b := ec.Add(V(x*1.15, -4.2, z*1.4))
				g.Fill(Capsule{a, b, 0.42})
			}
		}
		// Bow propeller.
		hub := gc.Add(V(hl+0.6, 0.2, 0))
		g.Fill(Capsule{gc.Add(V(hl-1, 0.2, 0)), hub, 0.6})
		g.Fill(Capsule{hub.Add(V(0.3, 2.4, 0)), hub.Add(V(0.3, -2.4, 0)), 0.5})
		g.Fill(Capsule{hub.Add(V(0.3, 0, 2.4)), hub.Add(V(0.3, 0, -2.4)), 0.5})
	})
}

// giantTree is a great golden tree with flared roots, spreading branches and a crown
// of hollow leaf clusters.
func (FloatingIslands) giantTree(g *Grid, c Vec2, ground float64) {
	g.With(Wood, func() {
		g.FillVolume(Frustum{c, 2.8, 1.5, ground - 1.5, ground + 12})
		for k := 0; k < 6; k++ {
			a := float64(k)*math.Pi/3 + 0.3
			g.Fill(Capsule{Polar(c, 1.5, a).At(ground + 2.5), Polar(c, 5.2, a+0.2).At(ground - 0.6), 0.8})
		}
	})
	type blob struct {
		a, r, h, size float64
		color         Color
	}
	blobs := []blob{
		{0, 5.5, 15.5, 4.6, Gold}, {1.3, 5.8, 14.5, 4.4, Gold}, {2.6, 5.4, 16, 4.6, Terracotta},
		{3.8, 5.8, 14.8, 4.4, Gold}, {5.0, 5.5, 15.2, 4.5, Red}, {0, 0, 19, 5.2, Gold},
	}
	for _, b := range blobs {
		tip := Polar(c, b.r, b.a).At(ground + b.h - 1)
		g.With(Wood, func() { g.Fill(Capsule{c.At(ground + 9.5), tip, 0.8}) })
		center := Polar(c, b.r, b.a).At(ground + b.h)
		radii := V(b.size, b.size*0.75, b.size)
		g.With(b.color, func() {
			g.FillVolume(Subtract(Ellipsoid{center, radii}, Ellipsoid{center, radii.Sub(V(1.3, 1.3, 1.3))}))
		})
	}
	// Glowing fruit.
	for k := 0; k < 10; k++ {
		a := float64(k) * 2.4
		p := Polar(c, 6+2*Hash01(k, 90), a).At(ground + 11.5 + 2*Hash01(k, 91))
		g.With(Glass, func() { g.Fill(Sphere(p, 0.75)) })
	}
}

func (f FloatingIslands) Build(g *Grid) Order {
	main := floatingIslandsIsle{c: P(0, -4), r: 22, top: 36, depth: 34, seed: 101, roots: 30, crystals: 7, grassy: true, hollowed: true}
	pag := floatingIslandsIsle{c: P(-46, 8), r: 12.5, top: 31, depth: 23, seed: 202, roots: 16, crystals: 3, grassy: true, hollowed: true}
	mill := floatingIslandsIsle{c: P(42, 18), r: 11, top: 28, depth: 21, seed: 303, roots: 14, crystals: 3, grassy: true, hollowed: true}
	light := floatingIslandsIsle{c: P(-18, 37), r: 10, top: 15, depth: 15, seed: 404, roots: 11, crystals: 3, grassy: true, hollowed: true}
	ruin := floatingIslandsIsle{c: P(22, -40), r: 9, top: 52, depth: 19, seed: 505, roots: 11, crystals: 2, grassy: true, hollowed: true}
	dock := floatingIslandsIsle{c: P(48, -14), r: 7, top: 35, depth: 14, seed: 606, roots: 8, crystals: 1, grassy: true, hollowed: true}
	grove := floatingIslandsIsle{c: P(-50, -28), r: 9, top: 41, depth: 18, seed: 707, roots: 12, crystals: 2, grassy: true, hollowed: true}
	isles := []floatingIslandsIsle{main, pag, mill, light, ruin, dock, grove}
	named := len(isles)
	// Small floating rocks.
	rocks := [][4]float64{{-12, -44, 46, 3.0}, {-28, 18, 24, 2.4}, {30, 40, 18, 2.6}, {-30, -12, 21, 2.4},
		{6, -46, 28, 2.4}, {-44, 38, 36, 2.2}, {60, 34, 22, 2.4}}
	for i, r := range rocks {
		isles = append(isles, floatingIslandsIsle{c: P(r[0], r[1]), r: r[3], top: r[2], depth: r[3] * 2.4,
			seed: 800 + i, roots: 3, grassy: true})
	}
	for _, s := range isles {
		s.build(g)
	}

	// Main island: temple, tower, pond and waterfalls.
	templeC, towerC := P(1, -3), P(-3, -17)
	f.temple(g, templeC, towerC, main.top)
	frontFall := 1.78
	main.waterfall(g, frontFall, 2.6, 26, 0.72)
	main.waterfall(g, 3.75, 2.2, 22, 0.66)
	pag.waterfall(g, 1.95, 2.0, 18, 0.55)
	mill.waterfall(g, 0.25, 2.0, 17, 0.6)
	ruin.waterfall(g, 0.9, 1.8, 26, 0.45)
	grove.waterfall(g, 2.6, 1.8, 20, 0.5)

	// Bridges, anchored just inside the rims.
	anchor := func(s floatingIslandsIsle, toward Vec2, inset float64) Vec3 {
		theta := toward.Sub(s.c).Angle()
		p := s.rim(theta, 1).Sub(Polar(P(0, 0), inset, theta))
		return p.At(s.topAt(p.X, p.Y) + 0.2)
	}
	links := []struct {
		to   floatingIslandsIsle
		sag  float64
		rail Color
	}{{pag, 2.2, Sand}, {mill, 1.8, Charcoal}, {dock, 2.4, Sand}, {grove, 1.6, Charcoal}}
	for _, l := range links {
		f.bridge(g, anchor(main, l.to.c, 2), anchor(l.to, main.c, 2), 2.4, l.sag, l.rail)
		// Sand path from the bridge head to the temple.
		a := anchor(main, l.to.c, 1)
		g.Paint(Sand, Func{main.bounds(), func(p Vec3) bool {
			q, ab := p.Plan().Sub(a.Plan()), templeC.Sub(a.Plan())
			t := math.Max(0, math.Min(1, q.Dot(ab)/ab.Dot(ab)))
			return p.Plan().Sub(a.Plan().Add(ab.Scale(t))).Length() < 1.3 && p.Y >= main.topAt(p.X, p.Z)-0.9 && main.contains(p)
		}})
	}
	// Floating stepping stones from the main island up to the ruins.
	from, to := anchor(main, ruin.c, 3), anchor(ruin, main.c, 2.5)
	side := P(-(to.Z - from.Z), to.X-from.X).Scale(1 / to.Plan().Sub(from.Plan()).Length())
	for k := 1; k <= 8; k++ {
		t := float64(k) / 9
		at := from.Plan().Add(to.Plan().Sub(from.Plan()).Scale(t)).Add(side.Scale(5 * math.Sin(t*math.Pi)))
		y := from.Y + (to.Y-from.Y)*t
		g.With(Stone, func() { g.Fill(Cylinder{Center: at, Radius: 1.6, Y0: y - 0.9, Y1: y}) })
		g.With(DarkStone, func() { g.FillVolume(Frustum{at, 0.3, 1.5, y - 3, y - 0.8}) })
		if k%3 == 1 {
			g.Paint(Grass, Cylinder{Center: at, Radius: 1.7, Y0: y - 0.4, Y1: y})
		}
	}

	// Pagoda island.
	pagC := pag.c.Add(P(-1, -1))
	toriiAt := pag.rim(0.3, 0.62)
	f.pagoda(g, pagC, pag.top+2.4, toriiAt, -0.3+math.Pi/2)

	// Windmill island with striped fields and a farmhouse.
	face := P(0.55, 0.83).Scale(1 / math.Hypot(0.55, 0.83))
	millC := mill.c.Add(P(-1.5, -1))
	field := func(p Vec3) (Vec2, bool) {
		q := p.Plan().Sub(mill.c)
		return q, mill.contains(p) && p.Y >= mill.topAt(p.X, p.Z)-0.9 && q.Y < -2.5
	}
	g.Paint(Gold, Func{mill.bounds(), func(p Vec3) bool {
		q, ok := field(p)
		return ok && q.X > -6 && int(math.Floor((q.X+q.Y)/1.6))%2 == 0
	}})
	g.Paint(Terracotta, Func{mill.bounds(), func(p Vec3) bool {
		q, ok := field(p)
		return ok && q.X <= -6 && int(math.Floor((q.X-q.Y)/1.6))%3 == 0
	}})
	f.windmill(g, millC, mill.top+2.2, face)
	g.With(Stone, func() {
		g.Build(kit.Hall{Min: V(mill.c.X+2.5, mill.top+0.8, mill.c.Y+2), Max: V(mill.c.X+8, mill.top+4.2, mill.c.Y+6.5), Ridge: AxisX, RoofColor: Brick})
	})

	// Lighthouse island with the keeper's cottage.
	f.lighthouse(g, light.c.Add(P(1, -1)), light.top+2)
	g.With(White, func() {
		g.Build(kit.Hall{Min: V(light.c.X-7, light.top+0.4, light.c.Y+0.5), Max: V(light.c.X-2.5, light.top+3.8, light.c.Y+5), Ridge: AxisZ, RoofColor: Navy})
	})

	f.ruins(g, ruin.c, ruin.top+1.5)
	f.giantTree(g, grove.c, grove.top+1.8)

	// Dock: a pier running -Z with the airship moored broadside at its end.
	pierY := dock.top + 1.2
	pz0, pz1 := dock.c.Y-3, dock.c.Y-13
	px := dock.c.X
	g.With(Wood, func() {
		g.Fill(Block(V(px-1.4, pierY-0.9, pz1), V(px+1.4, pierY, pz0)))
		g.Fill(Capsule{V(px, pierY-0.6, pz1+1), V(px, dock.top-5, dock.c.Y-4), 0.6})
		g.Fill(Capsule{V(px, pierY-0.6, pz1+6), V(px, dock.top-3, dock.c.Y-3), 0.6})
		for _, x := range []float64{-1.6, 1.6} {
			g.Fill(Cylinder{Center: P(px+x, pz1+0.6), Radius: 0.6, Y0: pierY - 0.5, Y1: pierY + 2.2})
			g.Fill(Cylinder{Center: P(px+x, pz1+5.5), Radius: 0.6, Y0: pierY - 0.5, Y1: pierY + 2.2})
		}
	})
	g.Paint(Charcoal, Func{Box{V(px-1.5, pierY-1, pz1), V(px+1.5, pierY, pz0)}, func(p Vec3) bool {
		return int(math.Floor(p.Z/1.2))%3 == 0
	}})
	gondola := V(px+1.5, pierY+0.8, pz1-3.0)
	f.airship(g, gondola, 22)
	g.With(Sand, func() {
		g.Fill(Capsule{V(px-1.6, pierY+2.2, pz1+0.6), gondola.Add(V(-6, 0.4, 1.0)), 0.45})
		g.Fill(Capsule{V(px+1.6, pierY+2.2, pz1+0.6), gondola.Add(V(5, 0.4, 1.0)), 0.45})
	})
	g.With(Wood, func() {
		g.Build(kit.Hall{Min: V(dock.c.X-4.5, dock.top+0.6, dock.c.Y-1), Max: V(dock.c.X-0.5, dock.top+3.6, dock.c.Y+3), Ridge: AxisX, RoofColor: Red})
	})

	// Trees.
	pond := main.rim(frontFall, 0.72)
	f.trees(g, main, 26, [][3]float64{{templeC.X, templeC.Y, 13.5}, {towerC.X, towerC.Y, 6}, {pond.X, pond.Y, 4}})
	f.trees(g, pag, 8, [][3]float64{{pagC.X, pagC.Y, 8.5}, {toriiAt.X, toriiAt.Y, 4.5}})
	f.trees(g, mill, 4, [][3]float64{{millC.X, millC.Y, 7}, {mill.c.X + 5, mill.c.Y + 4, 5}, {mill.c.X, mill.c.Y - 6, 6}})
	f.trees(g, light, 4, [][3]float64{{light.c.X + 1, light.c.Y - 1, 5}, {light.c.X - 4.7, light.c.Y + 2.7, 4}})
	f.trees(g, ruin, 3, [][3]float64{{ruin.c.X, ruin.c.Y, 6.5}})
	f.trees(g, dock, 2, [][3]float64{{dock.c.X - 2.5, dock.c.Y + 1, 3.5}, {dock.c.X, dock.c.Y - 4, 3}})
	f.trees(g, grove, 3, [][3]float64{{grove.c.X, grove.c.Y, 9}})
	for i := range rocks {
		if i%2 == 0 {
			f.trees(g, isles[named+i], 1, nil)
		}
	}
	return Sweep(P(0, 0), 0)
}
