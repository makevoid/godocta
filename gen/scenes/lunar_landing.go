package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// LunarLanding is an Apollo J-mission landing site, built large so the craft is the
// subject. The Lunar Module stands at the center facing +Z: the gold-foil octagonal
// descent stage with its engine bell, four splayed legs on round footpads, the ladder and
// porch on the forward leg and the deployed equipment bay; above it the faceted ascent
// stage with its paneled cabin, triangular windows, forward hatch, propellant-tank
// cheeks, docking tunnel, rendezvous radar, steerable S-band dish, VHF whips and the four
// RCS thruster quads. Around it: the Lunar Roving Vehicle with wire wheels, fenders,
// seats, console, TV camera and umbrella high-gain antenna; two suited astronauts (one
// saluting the flag), a striped flag, the solar-wind sheet, a TV tripod, and the
// experiment package (central station, RTG, seismometer, laser reflector) with its
// cables. The grey regolith is cratered and strewn with boulders, crossed by footprint
// trails and rover tracks, and slopes down to a low rim. A small half-lit Earth hangs
// in the sky above the western rim.
type LunarLanding struct{}

func init() { Register(LunarLanding{}) }

func (LunarLanding) Name() string  { return "lunar_landing" }
func (LunarLanding) Title() string { return "Lunar Landing" }

const (
	lunarLandingK    = 7.0               // World units per meter for the craft.
	lunarLandingProp = lunarLandingK / 5 // Scale of props authored at 5 units per meter.
	lunarLandingA    = 72.0              // Site semi-axis along X.
	lunarLandingB    = 64.0              // Site semi-axis along Z.
	lunarLandingH    = 3.0               // Plateau height above the rim.
)

// Plan positions of the surface hardware.
var (
	lunarLandingRover  = P(40, 32)
	lunarLandingFlag   = P(-25, 40)
	lunarLandingSalute = P(-14, 45)
	lunarLandingScoop  = P(26, 48)
	lunarLandingSheet  = P(-27, 16)
	lunarLandingTV     = P(20, 31)
	lunarLandingCS     = P(-36, -30)
	lunarLandingRTG    = P(-44, -22)
	lunarLandingPSE    = P(-28, -42)
	lunarLandingLRRR   = P(-46, -37)
	lunarLandingClear  = []Vec2{P(0, 0), lunarLandingRover, lunarLandingFlag, lunarLandingSalute, lunarLandingScoop,
		lunarLandingSheet, lunarLandingTV, lunarLandingCS, lunarLandingRTG, lunarLandingPSE, lunarLandingLRRR}
)

// lunarLandingScaled scales a solid about the origin.
type lunarLandingScaled struct {
	S Solid
	F float64
}

func (s lunarLandingScaled) Contains(p Vec3) bool { return s.S.Contains(p.Scale(1 / s.F)) }
func (s lunarLandingScaled) Bounds() Box {
	b := s.S.Bounds()
	return Box{b.Min.Scale(s.F), b.Max.Scale(s.F)}
}

type lunarLandingCrater struct{ X, Z, R, D float64 }

type lunarLandingPrint struct{ P, Dir Vec2 }

// lunarLandingSite carries the derived site data used while building.
type lunarLandingSite struct {
	g       *Grid
	craters []lunarLandingCrater
	prints  []lunarLandingPrint
	tracks  [][]Vec2 // Rover wheel ruts, densely sampled.
	cables  [][]Vec2
}

func (LunarLanding) Build(g *Grid) Order {
	s := &lunarLandingSite{g: g}
	s.plan()
	s.terrain()
	s.lunarModule()
	s.rover(lunarLandingRover, -0.5)
	s.flag(lunarLandingFlag, -0.25)
	s.astronaut(lunarLandingSalute, 1.2, true, true)
	s.astronaut(lunarLandingScoop, 0.9, false, false)
	s.experiments()
	s.boulders()
	s.earth()
	s.hollow(2)
	return Outward(P(0, 0))
}

func (lunarLandingSite) smooth(e0, e1, v float64) float64 {
	t := math.Max(0, math.Min(1, (v-e0)/(e1-e0)))
	return t * t * (3 - 2*t)
}

// noise is smooth value noise in [-1, 1].
func (lunarLandingSite) noise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	ab, cd := a+(b-a)*sx, c+(d-c)*sx
	return (ab+(cd-ab)*sz)*2 - 1
}

// edge is the normalized, slightly wobbly elliptical site radius (1 at the rim).
func (s *lunarLandingSite) edge(x, z float64) float64 {
	e := math.Hypot(x/lunarLandingA, z/lunarLandingB)
	a := math.Atan2(z, x)
	return e * (1 + 0.035*math.Sin(3*a+0.7) + 0.025*math.Sin(7*a+2) + 0.02*s.noise(x/9, z/9, 3))
}

// craterAt is the height offset of one crater with a raised rim.
func (lunarLandingSite) craterAt(c lunarLandingCrater, x, z float64) float64 {
	d := math.Hypot(x-c.X, z-c.Z) / c.R
	rim := c.D * 0.32
	if d < 1 {
		return c.D*(d*d-1) + rim
	}
	t := (d - 1) / 0.5
	return rim * math.Exp(-t*t)
}

// ground is the regolith surface height.
func (s *lunarLandingSite) ground(x, z float64) float64 {
	e := s.edge(x, z)
	base := lunarLandingH * (1 - s.smooth(0.84, 1.0, e))
	r := math.Hypot(x, z)
	calm := 0.3 + 0.7*s.smooth(16, 42, r)
	und := (0.55*s.noise(x/15, z/15, 1) + 0.25*s.noise(x/6, z/6, 2)) * calm * (1 - s.smooth(0.8, 1.0, e))
	h := base + und
	fade := 1 - s.smooth(0.85, 1.0, e)
	for _, c := range s.craters {
		if math.Abs(x-c.X) < c.R*2.2 && math.Abs(z-c.Z) < c.R*2.2 {
			h += s.craterAt(c, x, z) * fade
		}
	}
	return h
}

// spline samples a Catmull-Rom curve through the points every step units.
func (lunarLandingSite) spline(pts []Vec2, step float64) []Vec2 {
	var out []Vec2
	for i := 0; i+1 < len(pts); i++ {
		p0, p1, p2, p3 := pts[max(i-1, 0)], pts[i], pts[i+1], pts[min(i+2, len(pts)-1)]
		n := max(1, int(p2.Sub(p1).Length()/step))
		for j := 0; j < n; j++ {
			t := float64(j) / float64(n)
			t2, t3 := t*t, t*t*t
			f := func(a, b, c, d float64) float64 {
				return 0.5 * (2*b + (-a+c)*t + (2*a-5*b+4*c-d)*t2 + (-a+3*b-3*c+d)*t3)
			}
			out = append(out, P(f(p0.X, p1.X, p2.X, p3.X), f(p0.Y, p1.Y, p2.Y, p3.Y)))
		}
	}
	return append(out, pts[len(pts)-1])
}

// offset shifts a dense polyline sideways (positive to the left of travel).
func (lunarLandingSite) offset(line []Vec2, d float64) []Vec2 {
	out := make([]Vec2, len(line))
	for i := range line {
		a, b := line[max(i-1, 0)], line[min(i+1, len(line)-1)]
		t := b.Sub(a)
		t = t.Scale(1 / math.Max(t.Length(), 1e-9))
		out[i] = line[i].Add(P(-t.Y, t.X).Scale(d))
	}
	return out
}

// near reports the distance from p to a polyline and the arc length at the closest point.
func (lunarLandingSite) near(line []Vec2, p Vec2, limit float64) (float64, float64) {
	best, at, run := math.Inf(1), 0.0, 0.0
	for i := 0; i+1 < len(line); i++ {
		a, b := line[i], line[i+1]
		ab := b.Sub(a)
		l := ab.Length()
		if math.Abs(p.X-a.X) < l+limit && math.Abs(p.Y-a.Y) < l+limit {
			t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/math.Max(l*l, 1e-12)))
			if d := p.Sub(a.Add(ab.Scale(t))).Length(); d < best {
				best, at = d, run+t*l
			}
		}
		run += l
	}
	return best, at
}

// plan lays out craters, footprint trails, rover tracks and experiment cables.
func (s *lunarLandingSite) plan() {
	s.craters = []lunarLandingCrater{
		{40, -30, 16, 6.5},
		{-54, 6, 10, 3.8},
		{10, -48, 7, 2.8},
		{-44, 36, 6, 2.4},
		{56, 10, 6, 2.2},
	}
	for i := 0; len(s.craters) < 36 && i < 500; i++ {
		x := (Hash01(i, 11)*2 - 1) * lunarLandingA * 0.9
		z := (Hash01(i, 12)*2 - 1) * lunarLandingB * 0.9
		r := 1.6 + 3.4*math.Pow(Hash01(i, 13), 2)
		if s.edge(x, z) > 0.82 {
			continue
		}
		ok := math.Hypot(x, z) > 38+r
		for _, k := range lunarLandingClear {
			if P(x, z).Sub(k).Length() < 11+r {
				ok = false
			}
		}
		for _, c := range s.craters {
			if P(x, z).Sub(P(c.X, c.Z)).Length() < c.R+r+2 {
				ok = false
			}
		}
		if ok {
			s.craters = append(s.craters, lunarLandingCrater{x, z, r, r * 0.4})
		}
	}

	walk := func(pts ...Vec2) {
		line := s.spline(pts, 0.5)
		run, side := 0.0, 1.0
		for i := 1; i < len(line); i++ {
			d := line[i].Sub(line[i-1])
			run += d.Length()
			if run >= 1.8 {
				run = 0
				t := d.Scale(1 / math.Max(d.Length(), 1e-9))
				n := P(-t.Y, t.X).Scale(0.6 * side)
				j := Hash01(len(s.prints), 77) * 0.3
				s.prints = append(s.prints, lunarLandingPrint{line[i].Add(n).Add(t.Scale(j)), t})
				side = -side
			}
		}
	}
	walk(P(0, 37), P(-7, 41), P(-14, 43))
	walk(P(-16, 44), P(-22, 42))
	walk(P(-20, 38), P(-10, 38), P(-2, 38))
	walk(P(2, 38), P(12, 38), P(20, 42), P(25, 47))
	walk(P(27, 46), P(34, 42), P(42, 36))
	walk(P(-1, 36), P(-12, 29), P(-22, 21), P(-27, 17))
	walk(P(-4, 36), P(-24, 30), P(-40, 14), P(-44, -6), P(-37, -25))
	walk(P(-36, -30), P(-43, -23))
	walk(P(-35, -31), P(-29, -41))
	walk(P(-37, -31), P(-45, -36))
	walk(P(-32, -27), P(-26, -16), P(-28, -2), P(-33, 8))
	walk(P(2, 36), P(12, 33), P(19, 32))
	walk(P(19, 28), P(24, 16), P(30, 4), P(30, -8), P(28, -14))
	walk(P(30, -13), P(40, -10), P(44, 2), P(36, 16), P(38, 27))
	walk(P(-5, 37), P(-3, 44), P(5, 46), P(8, 40))

	// The rover came in from the southwest along a lazy curve and parked heading away.
	h := P(math.Cos(-0.5), math.Sin(-0.5))
	park := lunarLandingRover
	tread := 4.6 * lunarLandingProp
	route := s.spline([]Vec2{P(-42, 64), P(-18, 57), P(6, 50), park.Sub(h.Scale(20)), park.Sub(h.Scale(7))}, 0.6)
	out := s.spline([]Vec2{park.Add(h.Scale(7)), P(58, 16), P(68, 0), P(73, -14), P(76, -28)}, 0.6)
	for _, line := range [][]Vec2{route, out} {
		s.tracks = append(s.tracks, s.offset(line, tread), s.offset(line, -tread))
	}

	cs := lunarLandingCS
	for _, to := range []Vec2{lunarLandingRTG, lunarLandingPSE, lunarLandingLRRR} {
		mid := cs.Add(to).Scale(0.5).Add(P(Hash01(int(to.X), 5)*3-1.5, Hash01(int(to.Y), 6)*3-1.5))
		s.cables = append(s.cables, s.spline([]Vec2{cs, mid, to}, 0.5))
	}
}

// raster sets every cell whose center lies in the box and for which f reports a color,
// without half-unit snapping, replacing existing cells.
func (s *lunarLandingSite) raster(b Box, f func(Vec3) (Color, bool)) {
	g := s.g
	sp := g.Span()
	lo := func(v float64) int32 { return int32(math.Ceil(v * 2 / sp)) }
	hi := func(v float64) int32 { return int32(math.Floor(v * 2 / sp)) }
	previous := g.Brush()
	for y := lo(b.Min.Y); y <= hi(b.Max.Y); y++ {
		z0, x0 := lo(b.Min.Z), lo(b.Min.X)
		if (z0^y)&1 != 0 {
			z0++
		}
		if (x0^y)&1 != 0 {
			x0++
		}
		for z := z0; z <= hi(b.Max.Z); z += 2 {
			for x := x0; x <= hi(b.Max.X); x += 2 {
				c := Cell{X: x, Y: y, Z: z}
				if color, ok := f(g.World(c)); ok {
					g.Remove(c)
					g.Use(color).Add(c)
				}
			}
		}
	}
	g.Use(previous)
}

// recolor changes the color of existing cells in the box for which f reports one.
func (s *lunarLandingSite) recolor(b Box, f func(Vec3) (Color, bool)) {
	g := s.g
	previous := g.Brush()
	for _, c := range g.Cells() {
		p := g.World(c)
		if !b.Contains(p) {
			continue
		}
		if color, ok := f(p); ok {
			g.Remove(c)
			g.Use(color).Add(c)
		}
	}
	g.Use(previous)
}

// terrain rasterizes the regolith shell with craters, speckles and surface marks.
func (s *lunarLandingSite) terrain() {
	A, B := lunarLandingA*1.12, lunarLandingB*1.12
	box := Box{V(-A, -4, -B), V(A, lunarLandingH+4, B)}
	sp := s.g.Span()
	type column struct{ h, t float64 }
	cache := map[[2]int32]column{}
	s.raster(box, func(p Vec3) (Color, bool) {
		x, y, z := p.X, p.Y, p.Z
		key := [2]int32{int32(math.Round(x * 4)), int32(math.Round(z * 4))}
		col, ok := cache[key]
		if !ok {
			if s.edge(x, z) > 1 {
				col = column{math.NaN(), 0}
			} else {
				h := s.ground(x, z)
				gx := s.ground(x+0.4, z) - s.ground(x-0.4, z)
				gz := s.ground(x, z+0.4) - s.ground(x, z-0.4)
				slope := math.Hypot(gx, gz) / 0.8
				col = column{h, sp*1.02 + math.Min(slope*0.8, 3)}
			}
			cache[key] = col
		}
		if math.IsNaN(col.h) || y > col.h+0.02 || y < col.h-col.t {
			return 0, false
		}
		return s.soil(x, y, z, col.h), true
	})
}

// soil picks the regolith color: speckled grey, darker patches and crater floors,
// a bright blast halo around the lander, and dark footprints, ruts and cables on top.
func (s *lunarLandingSite) soil(x, y, z, h float64) Color {
	ix, iy, iz := int(math.Round(x*4)), int(math.Round(y*4)), int(math.Round(z*4))
	top := y > h-0.6
	p := P(x, z)
	if top {
		for _, f := range s.prints {
			d := p.Sub(f.P)
			if math.Abs(d.X) < 1 && math.Abs(d.Y) < 1 {
				a, b := d.Dot(f.Dir), d.X*f.Dir.Y-d.Y*f.Dir.X
				if (a*a)/(0.8*0.8)+(b*b)/(0.45*0.45) <= 1 {
					return Charcoal
				}
			}
		}
		for _, t := range s.tracks {
			if d, run := s.near(t, p, 0.7); d < 0.7 {
				if math.Mod(run, 1.4) < 0.55 {
					return Charcoal
				}
				return DarkStone
			}
		}
		for _, c := range s.cables {
			if d, _ := s.near(c, p, 0.4); d < 0.4 {
				return White
			}
		}
	}
	r := math.Hypot(x, z)
	// Trampled ground at the ladder foot.
	if top && p.Sub(P(0, 37)).Length() < 6 && Hash01(ix, iz, 21) < 0.4 {
		return DarkStone
	}
	n := s.noise(x/14, z/14, 7) + 0.5*s.noise(x/5, z/5, 8)
	dark, white := 0.06, 0.006
	if n < -0.25 {
		dark += math.Min(0.5, (-n-0.25)*1.1)
	}
	if n > 0.4 {
		white += math.Min(0.25, (n-0.4)*0.7)
	}
	for _, c := range s.craters {
		d := math.Hypot(x-c.X, z-c.Z) / c.R
		if d < 0.9 {
			dark += 0.3 + 0.25*(1-d)
		} else if d < 1.6 && c.R > 3 {
			white += 0.12 * (1.6 - d)
			// Bright ejecta rays from the fresh large crater.
			if c.R > 12 && d < 2.6 {
				a := math.Atan2(z-c.Z, x-c.X)
				if math.Mod(a*7+20, 2) < 0.6 {
					white += 0.3
				}
			}
		} else if c.R > 12 && d < 2.8 {
			a := math.Atan2(z-c.Z, x-c.X)
			if math.Mod(a*7+20, 2) < 0.6 {
				white += 0.3 * (2.8 - d)
			}
		}
	}
	if r < 34 {
		k := s.smooth(34, 18, r)
		dark *= 1 - 0.8*k
		white += 0.1 * k
	}
	v := Hash01(ix, iy, iz, 9)
	switch {
	case v < dark:
		return DarkStone
	case v > 1-white:
		return White
	}
	return Stone
}

// local converts a world point into the frame of an object at, turned by angle.
func (lunarLandingSite) local(p Vec3, at Vec2, base, angle float64) Vec3 {
	q := p.Plan().Sub(at).Rotate(-angle)
	return V(q.X, p.Y-base, q.Y).Scale(1 / lunarLandingProp)
}

// place moves a solid authored in a local frame (forward +X) to at, turned by angle.
func (lunarLandingSite) place(sol Solid, at Vec2, base, angle float64) Solid {
	return RotateY(Translate(lunarLandingScaled{sol, lunarLandingProp}, at.At(base)), at, angle)
}

// dish is a shallow paraboloid shell opening along axis.
func (lunarLandingSite) dish(c, axis Vec3, radius, depth, thick float64) Solid {
	axis = axis.Scale(1 / axis.Length())
	return Func{Box{c, c}.Grow(radius + depth + thick), func(p Vec3) bool {
		d := p.Sub(c)
		a := d.Dot(axis)
		r := d.Sub(axis.Scale(a)).Length()
		if r > radius {
			return false
		}
		cup := depth * (r / radius) * (r / radius)
		return a >= cup-thick/2 && a <= cup+thick/2
	}}
}

// lunarModule builds the LM at the origin, its front (ladder, hatch) facing +Z.
func (s *lunarLandingSite) lunarModule() {
	g := s.g
	m := func(v float64) float64 { return v * lunarLandingK }
	g0 := s.ground(0, 0)
	w, cut := m(2.1), m(0.85)
	oct := []Vec2{P(w, -(w - cut)), P(w, w-cut), P(w-cut, w), P(-(w - cut), w), P(-w, w-cut), P(-w, -(w - cut)), P(-(w - cut), -w), P(w-cut, -w)}
	y0, y1 := g0+m(1.35), g0+m(3.0)

	// Descent stage body and engine.
	g.Use(Gold).Fill(Prism{oct, y0, y1})
	s.recolor(Box{V(-w-1, y0-1, -w-1), V(w+1, y1+1, w+1)}, func(p Vec3) (Color, bool) {
		ax, az := math.Abs(p.X), math.Abs(p.Z)
		ix, iy, iz := int(math.Round(p.X*4)), int(math.Round(p.Y*4)), int(math.Round(p.Z*4))
		switch {
		case p.Y < y0+0.7:
			return Charcoal, true
		case ax > w-cut-0.3 && az > w-cut-0.3 && ax+az > 2*w-cut-1.2:
			// Corner bays: black blankets below, gold above.
			if p.Y < y0+(y1-y0)*0.55 {
				return Charcoal, true
			}
			if Hash01(ix, iy, iz, 31) < 0.3 {
				return Sand, true
			}
			return Gold, true
		case Hash01(int(math.Floor(p.X/2.2)), int(math.Floor(p.Y/2.2)), int(math.Floor(p.Z/2.2)), 33) < 0.12:
			return Stone, true
		case Hash01(ix, iy, iz, 32) < 0.3:
			return Sand, true
		}
		return Gold, true
	})
	bell := Subtract(Frustum{P(0, 0), m(0.8), m(0.5), g0 + m(0.55), y0},
		Frustum{P(0, 0), m(0.8) - 0.9, m(0.5) - 0.9, g0 + m(0.55) - 0.1, y0 - 0.9})
	g.Use(Charcoal).Fill(bell)
	g.Paint(DarkStone, Cylinder{P(0, 0), m(0.85), m(0.8) - 1, g0 + m(0.5), g0 + m(0.55) + 0.8})

	// Legs: primary strut (foil sleeve over a bright piston), two secondary struts, pads.
	for i := 0; i < 4; i++ {
		a := float64(i) * math.Pi / 2
		d := P(math.Cos(a), math.Sin(a))
		e := P(-d.Y, d.X)
		fp := d.Scale(m(4.7))
		gf := s.ground(fp.X, fp.Y)
		u := d.Scale(w + 0.6).At(y1 - 0.8)
		f := fp.At(gf + 1.4)
		mid := u.Add(f.Sub(u).Scale(0.52))
		g.Use(Gold).Fill(Capsule{u, mid, 1.0})
		g.Use(Stone).Fill(Capsule{mid, f, 0.75})
		g.Use(White).Fill(Sphere(f, 0.95))
		knee := u.Add(f.Sub(u).Scale(0.45))
		for _, side := range []float64{-1, 1} {
			root := d.Scale(w - 0.2).Add(e.Scale(side * m(1.05))).At(y0 + 0.8)
			g.Use(Gold).Fill(Capsule{root, knee, 0.6})
		}
		// Downlock truss from the knee to the body bottom center line.
		g.Use(DarkStone).Fill(Capsule{d.Scale(w - 0.2).At(y0 + 0.4), knee, 0.5})
		g.Use(Stone).Fill(Frustum{fp, m(0.38), m(0.47), gf - 0.6, gf + 0.9})
		g.Paint(White, Cylinder{fp, m(0.47) + 0.2, m(0.47) - 0.8, gf + 0.3, gf + 1})
		if i != 1 {
			// Bent contact probe lying beside the pad.
			tip := fp.Add(e.Scale(3.2)).Add(d.Scale(1.5))
			g.Use(DarkStone).Fill(Capsule{fp.Add(e.Scale(1.5)).At(gf + 0.3), tip.At(s.ground(tip.X, tip.Y) + 0.2), 0.45})
		}
	}

	// Ladder on the forward (+Z) leg, porch and handrails.
	{
		fp := P(0, m(4.7))
		gf := s.ground(fp.X, fp.Y)
		u := P(0, w+0.6).At(y1 - 0.8)
		f := fp.At(gf + 1.4)
		dir := f.Sub(u).Scale(1 / f.Sub(u).Length())
		n := V(0, dir.Z, -dir.Y) // Outward and up, perpendicular to the strut.
		if n.Y < 0 {
			n = n.Scale(-1)
		}
		l := f.Sub(u).Length()
		at := func(t, x float64) Vec3 { return u.Add(dir.Scale(t)).Add(n.Scale(1.25)).Add(V(x, 0, 0)) }
		g.Use(White)
		for _, x := range []float64{-1.7, 1.7} {
			g.Fill(Capsule{at(0.4, x), at(l-2.4, x), 0.5})
			g.Fill(Capsule{at(0.4, x), u.Add(dir.Scale(0.4)).Add(V(x, 0, 0)), 0.45})
		}
		for t := 1.2; t < l-2.4; t += 1.55 {
			g.Fill(Capsule{at(t, -1.7), at(t, 1.7), 0.45})
		}
		g.Use(Stone).Fill(Block(V(-2.6, y1-0.7, w-1), V(2.6, y1, w+2.2)))
		g.Use(White)
		for _, x := range []float64{-2.6, 2.6} {
			g.Fill(Capsule{V(x, y1, w+1.9), V(x, y1+2.6, w+1.9), 0.45})
			g.Fill(Capsule{V(x, y1+2.6, w+1.9), V(x, y1+2.6, m(1.7)), 0.45})
		}
	}

	// Deployed equipment bay (MESA) in the front-left corner bay.
	{
		ang := 3 * math.Pi / 4
		dist := (2*w - cut) / math.Sqrt2
		at := P(math.Cos(ang), math.Sin(ang)).Scale(dist)
		g.Use(Stone).Fill(s.place(Block(V(-0.2, m(1.45), -2.2), V(1.4, m(2.4), 2.2)), at, g0, ang))
		g.Use(Charcoal).Fill(s.place(Block(V(0.8, m(1.7), -1.8), V(1.5, m(1.85), 1.8)), at, g0, ang))
		g.Use(White).Fill(s.place(Block(V(0.2, m(1.35), -2.4), V(2.6, m(1.45), 2.4)), at, g0, ang))
	}

	// Ascent stage, in meters relative to the descent deck.
	A := y1
	M := func(x, y, z float64) Vec3 { return V(m(x), A+m(y), m(z)) }
	front := func(yl float64) float64 {
		if yl > 1.3 {
			return 1.65 - (yl-1.3)*(0.65/1.05)
		}
		return 1.65
	}
	cabin := Func{Box{M(-1.5, 0.2, -0.3), M(1.5, 2.36, 1.66)}, func(p Vec3) bool {
		yl, zm, xm := (p.Y-A)/lunarLandingK, p.Z/lunarLandingK, math.Abs(p.X)/lunarLandingK
		if zm > front(yl) {
			return false
		}
		hx := 1.45
		if zm > 0.9 {
			hx -= (zm - 0.9) * 0.6
		}
		if yl > 2.0 {
			hx -= (yl - 2.0) * 1.2
		}
		return xm <= hx
	}}
	g.Use(Stone).FillVolume(cabin)
	g.Fill(Block(M(-1.45, 0.2, -1.3), M(1.45, 2.2, -0.3)))
	g.Fill(Block(M(-1.15, 0.55, -2.25), M(1.15, 1.95, -1.3)))
	for _, sx := range []float64{-1, 1} {
		g.FillVolume(Capsule{M(sx*1.5, 1.05, -1.05), M(sx*1.5, 1.05, 0.1), m(0.6)})
	}
	g.Use(Charcoal).Fill(Block(M(-0.9, -0.05, -0.9), M(0.9, 0.25, 0.9)))
	g.Use(Stone).Fill(Cylinder{P(0, m(0.25)), m(0.55), 0, A + m(2.1), A + m(2.7)})
	g.Paint(Charcoal, Cylinder{P(0, m(0.25)), m(0.32), 0, A + m(2.6), A + m(2.75)})

	s.recolor(Box{M(-2.3, 0.1, -2.4), M(2.3, 2.8, 1.8)}, func(p Vec3) (Color, bool) {
		xl, yl, zl := p.X/lunarLandingK, (p.Y-A)/lunarLandingK, p.Z/lunarLandingK
		ax := math.Abs(xl)
		if yl > 2.15 && math.Hypot(xl, zl-0.25) < 0.6 {
			return 0, false
		}
		ix, iy, iz := int(math.Floor(p.X/1.7)), int(math.Floor(p.Y/1.7)), int(math.Floor(p.Z/1.7))
		fz := front(yl)
		switch {
		case yl < 0.42:
			if Hash01(int(p.X*4), int(p.Z*4), 41) < 0.25 {
				return Sand, true
			}
			return Gold, true
		case zl > fz-0.3 && yl > 1.3:
			// Slanted upper front: dark panel with the two triangular windows.
			wx := ax - 0.18
			wy := yl - 1.42
			if wx > 0 && wy > 0 && wx < 0.78 && wy < 0.7 && wy > 0.7*wx/0.78-0.04 {
				return Charcoal, true
			}
			return DarkStone, true
		case zl > fz-0.3 && ax < 0.42 && yl > 0.42 && yl < 1.2:
			// Forward hatch with its handle.
			if ax < 0.12 && yl > 0.7 && yl < 0.85 {
				return Charcoal, true
			}
			return DarkStone, true
		case zl > fz-0.3 && ax < 0.55 && yl > 0.32 && yl < 1.3:
			return White, true
		case zl < -1.25:
			// Aft equipment bay: black radiator fins and gold blankets.
			if yl < 0.9 {
				return Gold, true
			}
			if int(math.Floor(p.X/1.2))%2 == 0 {
				return Charcoal, true
			}
			return DarkStone, true
		case yl > 2.15 && math.Hypot(xl, zl-0.25) < 0.75:
			return White, true
		}
		switch v := Hash01(ix, iy, iz, 43); {
		case v < 0.14:
			return White, true
		case v < 0.22:
			return DarkStone, true
		}
		return Stone, true
	})

	// RCS thruster quads on the four corners.
	for _, sx := range []float64{-1, 1} {
		for _, zq := range []float64{1.0, -1.2} {
			c := M(sx*1.95, 2.0, zq)
			g.Use(Stone).Fill(Capsule{M(sx*1.3, 2.0, zq*0.8), c, 0.6})
			g.Use(White).Fill(BlockAt(c.Plan(), 1.8, 1.8, c.Y-0.9, c.Y+0.9))
			g.Use(Charcoal)
			for _, d := range []Vec3{V(0, 1, 0), V(0, -1, 0), V(sx, 0, 0), V(0, 0, math.Copysign(1, zq))} {
				g.Fill(Capsule{c.Add(d.Scale(0.9)), c.Add(d.Scale(1.9)), 0.5})
			}
		}
	}

	// Antennas: rendezvous radar, steerable S-band, VHF whips, docking target.
	rr := M(0.75, 2.55, 0.95)
	g.Use(DarkStone).Fill(Capsule{M(0.75, 2.0, 0.85), rr, 0.6})
	g.Use(White).Fill(s.dish(rr.Add(V(0, 0.3, 0.3)), V(0, 0.55, 0.85), m(0.32), 0.7, 0.75))
	g.Use(Charcoal).Fill(Capsule{rr, rr.Add(V(0, 1.1, 1.7)), 0.45})
	sb := M(1.85, 3.0, -0.75)
	g.Use(Stone).Fill(Capsule{M(1.35, 2.1, -0.6), sb, 0.5})
	g.Use(White).Fill(s.dish(sb, V(0.3, 0.9, 0.35), m(0.34), 0.7, 0.75))
	g.Use(Charcoal).Fill(Capsule{sb, sb.Add(V(0.4, 1.4, 0.5)), 0.45})
	g.Use(Stone).Fill(Capsule{M(-1.0, 2.2, -1.0), M(-1.25, 3.4, -1.3), 0.45})
	g.Fill(Capsule{M(1.0, 1.9, -1.9), M(1.45, 2.9, -2.2), 0.45})
	g.Use(DarkStone).Fill(Capsule{M(-0.8, 2.25, 0.8), M(-0.8, 2.75, 0.8), 0.45})
	g.Use(Charcoal).Fill(Capsule{M(-0.95, 2.75, 0.8), M(-0.65, 2.75, 0.8), 0.45})
	g.Fill(Capsule{M(-0.8, 2.7, 0.8), M(-0.8, 2.95, 0.8), 0.45})
}

// rover builds the Lunar Roving Vehicle parked at at, its front turned by angle.
func (s *lunarLandingSite) rover(at Vec2, angle float64) {
	g := s.g
	base := s.ground(at.X, at.Y) - 0.2
	put := func(c Color, sol Solid) { g.Use(c).Fill(s.place(sol, at, base, angle)) }
	wheel := func(c Vec3, r0, r1, half float64) Solid {
		return Func{Box{c.Sub(V(r1, r1, half)), c.Add(V(r1, r1, half))}, func(p Vec3) bool {
			d := math.Hypot(p.X-c.X, p.Y-c.Y)
			return math.Abs(p.Z-c.Z) <= half && d <= r1 && d >= r0
		}}
	}
	const wr = 2.05
	for _, wx := range []float64{-5.75, 5.75} {
		for _, wz := range []float64{-4.6, 4.6} {
			c := V(wx, wr, wz)
			put(Charcoal, wheel(c, wr-0.75, wr, 0.65))
			// Titanium chevrons on the wire-mesh tire.
			s.recolor(s.place(wheel(c, wr-0.8, wr+0.1, 0.8), at, base, angle).Bounds(), func(p Vec3) (Color, bool) {
				q := s.local(p, at, base, angle)
				d := math.Hypot(q.X-c.X, q.Y-c.Y)
				if math.Abs(q.Z-c.Z) > 0.8 || d < wr-0.85 || d > wr+0.15 {
					return 0, false
				}
				if math.Mod(math.Atan2(q.Y-c.Y, q.X-c.X)+math.Pi+math.Abs(q.Z-c.Z)*0.4, math.Pi/6) < math.Pi/14 {
					return DarkStone, true
				}
				return 0, false
			})
			put(Gold, wheel(c, 0, 0.9, 0.5))
			for k := 0; k < 3; k++ {
				a := float64(k) * math.Pi / 3
				d := V(math.Cos(a), math.Sin(a), 0).Scale(wr - 0.5)
				put(Stone, Capsule{c.Sub(d), c.Add(d), 0.45})
			}
			// Fender over the top of the wheel.
			fz := wz
			put(Stone, Func{Box{c.Sub(V(3, 0, 1)), c.Add(V(3, 3, 1))}, func(p Vec3) bool {
				d := math.Hypot(p.X-c.X, p.Y-c.Y)
				a := math.Atan2(p.Y-c.Y, p.X-c.X)
				return math.Abs(p.Z-fz) <= 0.7 && d >= wr+0.3 && d <= wr+0.85 && a > 0.5 && a < math.Pi-0.5
			}})
			// Suspension arms from the chassis.
			sz := math.Copysign(2.4, wz)
			put(DarkStone, Capsule{V(wx*0.75, 2.7, sz), V(wx, wr, wz-math.Copysign(0.7, wz)), 0.5})
		}
	}
	// Chassis, gold-blanketed forward bay, rear pallet.
	put(DarkStone, Block(V(-6.6, 2.4, -2.5), V(6.6, 3.1, 2.5)))
	put(Gold, Block(V(3.4, 3.1, -2.2), V(6.4, 4.0, 2.2)))
	put(Stone, Block(V(-7.0, 3.1, -2.4), V(-5.0, 4.8, 2.4)))
	put(DarkStone, Block(V(-7.2, 4.8, -2.0), V(-5.4, 5.6, 2.0)))
	// Seats with their frames.
	for _, sz := range []float64{-1.3, 1.3} {
		put(Glass, Block(V(-2.8, 3.1, sz-0.95), V(-0.6, 3.7, sz+0.95)))
		put(Glass, Block(V(-3.2, 3.1, sz-0.95), V(-2.6, 5.4, sz+0.95)))
		put(White, Capsule{V(-3.2, 5.4, sz-0.95), V(-3.2, 5.4, sz+0.95), 0.45})
	}
	// Console and hand controller.
	put(DarkStone, Capsule{V(1.4, 3.1, 0), V(1.0, 5.0, 0), 0.5})
	put(Charcoal, Block(V(0.5, 5.0, -1.2), V(1.7, 6.2, 1.2)))
	put(Charcoal, Capsule{V(-0.3, 3.1, 0), V(-0.3, 4.4, 0), 0.45})
	// High-gain umbrella antenna, low-gain helix and the TV camera.
	mast := V(5.2, 3.9, 1.6)
	top := V(5.2, 9.0, 1.6)
	put(Stone, Capsule{mast, top, 0.45})
	axis := V(-0.35, 0.85, 0.4)
	put(White, s.dish(top.Add(V(0, 0.3, 0)), axis, 2.6, 1.2, 0.75))
	put(Charcoal, Capsule{top, top.Add(axis.Scale(1.4)), 0.45})
	put(Stone, Capsule{V(4.6, 4.0, -1.7), V(4.4, 7.0, -1.9), 0.55})
	put(DarkStone, Capsule{V(6.6, 4.0, 0), V(6.6, 5.4, 0), 0.45})
	put(White, Block(V(6.0, 5.4, -0.7), V(7.4, 6.5, 0.7)))
	put(Gold, Block(V(7.4, 5.6, -0.4), V(7.9, 6.3, 0.4)))
}

// flag plants a striped flag on a pole with a top crossbar; the cloth extends along
// the local +X axis, turned by angle.
func (s *lunarLandingSite) flag(at Vec2, angle float64) {
	g := s.g
	base := s.ground(at.X, at.Y)
	g.Use(White).Fill(s.place(Capsule{V(0, -0.8, 0), V(0, 15.4, 0), 0.5}, at, base, angle))
	g.Fill(s.place(Capsule{V(0, 14.8, 0), V(10.2, 14.8, 0), 0.45}, at, base, angle))
	const top, bottom, x0, x1 = 14.4, 8.2, 0.6, 10.0
	cloth := Func{Box{V(x0, bottom, -1.2), V(x1, top, 1.2)}, func(p Vec3) bool {
		wave := 0.45 * math.Sin(p.X*0.85) * (p.X / x1)
		return math.Abs(p.Z-wave) <= 0.42
	}}
	sol := s.place(cloth, at, base, angle)
	g.Use(Red).Fill(sol)
	stripe := (top - bottom) / 7
	s.recolor(sol.Bounds(), func(p Vec3) (Color, bool) {
		q := s.local(p, at, base, angle)
		if !cloth.Contains(q) && !cloth.Contains(V(math.Round(q.X*2)/2, math.Round(q.Y*2)/2, math.Round(q.Z*2)/2)) {
			return 0, false
		}
		i := int(math.Floor((top - q.Y) / stripe))
		if q.X < 4.6 && i < 4 {
			if Hash01(int(math.Round(p.X*4)), int(math.Round(p.Y*4)), int(math.Round(p.Z*4)), 51) < 0.22 {
				return White, true
			}
			return Navy, true
		}
		if i%2 == 0 {
			return Red, true
		}
		return White, true
	})
}

// astronaut builds a suited figure facing angle; salute raises the right arm and
// commander adds the red rank bands.
func (s *lunarLandingSite) astronaut(at Vec2, angle float64, salute, commander bool) {
	g := s.g
	base := s.ground(at.X, at.Y) - 0.3
	put := func(c Color, sol Solid) { g.Use(c).Fill(s.place(sol, at, base, angle)) }
	putV := func(c Color, sol Solid) { g.Use(c).FillVolume(s.place(sol, at, base, angle)) }
	for _, z := range []float64{-0.85, 0.85} {
		put(White, Block(V(-0.6, 0, z-0.6), V(1.0, 1.3, z+0.6)))
		putV(White, Capsule{V(0, 1.2, z), V(0, 4.3, z*0.9), 0.85})
	}
	putV(White, Ellipsoid{V(0, 5.7, 0), V(1.2, 1.8, 1.6)})
	put(Stone, Block(V(-2.7, 3.9, -1.45), V(-1.0, 7.6, 1.45)))
	put(White, Block(V(-2.5, 7.6, -1.25), V(-1.1, 8.3, 1.25)))
	put(DarkStone, Block(V(0.9, 5.3, -0.8), V(1.8, 6.3, 0.8)))
	head := V(0.15, 8.15, 0)
	putV(White, Sphere(head, 1.35))
	s.recolor(s.place(Sphere(head, 1.6), at, base, angle).Bounds(), func(p Vec3) (Color, bool) {
		q := s.local(p, at, base, angle).Sub(head)
		if q.Length() > 1.7 || q.X < 0.45 || q.Y < -0.85 || q.Y > 0.9 {
			return 0, false
		}
		return Gold, true
	})
	// Arms.
	left := V(0, 6.6, -1.8)
	right := V(0, 6.6, 1.8)
	putV(White, Capsule{left, V(0.7, 3.9, -2.1), 0.7})
	if salute {
		elbow := V(0.9, 7.0, 3.1)
		putV(White, Capsule{right, elbow, 0.7})
		putV(White, Capsule{elbow, V(0.9, 8.4, 1.4), 0.65})
	} else {
		hand := V(1.6, 4.6, 2.0)
		putV(White, Capsule{right, hand, 0.7})
		// Long-handled scoop resting on the ground.
		put(Stone, Capsule{hand, V(3.6, 0.6, 2.6), 0.45})
		put(DarkStone, Block(V(3.3, 0.1, 2.0), V(4.4, 0.8, 3.2)))
	}
	if commander {
		s.recolor(s.place(Block(V(-1.5, 2.6, -3.6), V(1.5, 7.2, 3.6)), at, base, angle).Bounds(), func(p Vec3) (Color, bool) {
			q := s.local(p, at, base, angle)
			if (math.Abs(q.Z) > 1.55 && q.Y > 5.3 && q.Y < 6.0 && q.X > -0.8) || (q.Y > 2.7 && q.Y < 3.4 && q.X > -0.9 && q.X < 1.0 && math.Abs(q.Z) < 1.8) {
				return Red, true
			}
			return 0, false
		})
		s.recolor(s.place(Sphere(head, 1.7), at, base, angle).Bounds(), func(p Vec3) (Color, bool) {
			q := s.local(p, at, base, angle).Sub(head)
			if q.Length() < 1.8 && q.Y > 0.9 && math.Abs(q.Z) < 0.5 {
				return Red, true
			}
			return 0, false
		})
	}
}

// experiments sets out the surface experiment package, the solar-wind sheet and the
// TV camera tripod.
func (s *lunarLandingSite) experiments() {
	g := s.g
	at := func(x, z float64) float64 { return s.ground(x, z) }
	// Central station with its sunshield and antenna.
	cx, cz := lunarLandingCS.X, lunarLandingCS.Y
	b := at(cx, cz) - 0.3
	g.Use(Gold).Fill(BlockAt(P(cx, cz), 3.6, 2.6, b, b+2.6))
	g.Use(Stone).Fill(BlockAt(P(cx, cz), 3.0, 2.0, b+2.6, b+3.0))
	g.Use(White)
	for _, dx := range []float64{-2.2, 2.2} {
		for _, dz := range []float64{-1.6, 1.6} {
			g.Fill(Capsule{V(cx+dx, b+0.5, cz+dz), V(cx+dx, b+5.0, cz+dz), 0.45})
		}
	}
	g.Fill(BlockAt(P(cx, cz), 5.6, 4.2, b+5.0, b+5.6))
	g.Use(Stone).Fill(Capsule{V(cx+1.2, b+5.6, cz), V(cx+1.6, b+9.0, cz+0.4), 0.55})
	g.Use(Charcoal).Fill(Sphere(V(cx+1.6, b+9.2, cz+0.4), 0.7))
	// Radioisotope generator with cooling fins.
	rx, rz := lunarLandingRTG.X, lunarLandingRTG.Y
	rb := at(rx, rz) - 0.3
	g.Use(Charcoal).Fill(Func{Box{V(rx-2, rb, rz-2), V(rx+2, rb+3.4, rz+2)}, func(p Vec3) bool {
		d := P(p.X-rx, p.Z-rz)
		r := d.Length()
		if r <= 1.0 {
			return true
		}
		a := math.Mod(d.Angle()+2*math.Pi, math.Pi/4)
		return r <= 2.0 && p.Y < rb+2.9 && (a < 0.32 || a > math.Pi/4-0.32)
	}})
	g.Use(DarkStone).Fill(Cylinder{P(rx, rz), 0.8, 0, rb + 3.4, rb + 4.0})
	// Passive seismometer under its foil skirt.
	px, pz := lunarLandingPSE.X, lunarLandingPSE.Y
	pb := at(px, pz)
	g.Use(Stone).Fill(Cylinder{P(px, pz), 2.8, 0, pb - 0.2, pb + 0.25})
	g.Use(Gold).Fill(Cylinder{P(px, pz), 1.0, 0, pb, pb + 2.0})
	g.Use(White).Fill(Ellipsoid{V(px, pb+2.0, pz), V(1.1, 0.8, 1.1)})
	// Laser ranging retroreflector: a tilted panel of corner cubes.
	lx, lz := lunarLandingLRRR.X, lunarLandingLRRR.Y
	lb := at(lx, lz) - 0.2
	g.Use(Stone).Fill(BlockAt(P(lx, lz), 2.4, 2.0, lb, lb+1.4))
	panel := Func{Box{V(lx-3, lb+0.8, lz-3), V(lx+3, lb+4.5, lz+3)}, func(p Vec3) bool {
		d := V(p.X-lx, p.Y-(lb+2.4), p.Z-lz)
		// Panel tilted 35 degrees toward +Z (the Earth side).
		n := V(0, math.Cos(0.6), math.Sin(0.6))
		u := V(0, -math.Sin(0.6), math.Cos(0.6))
		return math.Abs(d.Dot(n)) <= 0.45 && math.Abs(d.X) <= 2.6 && math.Abs(d.Dot(u)) <= 1.9
	}}
	g.Use(Stone).Fill(panel)
	s.recolor(panel.Bounds(), func(p Vec3) (Color, bool) {
		if !panel.Contains(p) {
			return 0, false
		}
		if int(math.Round(p.X*1.5))%2 == 0 && int(math.Round(p.Y*1.5+p.Z*1.5))%2 == 0 {
			return Charcoal, true
		}
		return 0, false
	})
	// Solar-wind composition sheet.
	sx, sz := lunarLandingSheet.X, lunarLandingSheet.Y
	sbase := at(sx, sz)
	g.Use(Stone).Fill(Capsule{V(sx, sbase-0.5, sz), V(sx, sbase+9.4, sz), 0.45})
	g.Use(White).Fill(Block(V(sx-1.6, sbase+3.6, sz-0.4), V(sx+1.6, sbase+9.0, sz+0.4)))
	// TV camera tripod near the lander.
	tx, tz := lunarLandingTV.X, lunarLandingTV.Y
	tb := at(tx, tz)
	for k := 0; k < 3; k++ {
		a := float64(k)*2*math.Pi/3 + 0.4
		foot := Polar(P(tx, tz), 2.0, a)
		g.Use(DarkStone).Fill(Capsule{foot.At(s.ground(foot.X, foot.Y)), V(tx, tb+4.6, tz), 0.45})
	}
	g.Use(White).Fill(BlockAt(P(tx, tz), 1.6, 2.2, tb+4.6, tb+6.0))
	g.Use(Charcoal).Fill(BlockAt(P(tx-0.2, tz+1.4), 0.9, 0.9, tb+5.0, tb+5.7))
}

// boulders scatters rocks: a few large blocks near the big crater rim and many small
// ones over the plain, kept clear of the hardware.
func (s *lunarLandingSite) boulders() {
	g := s.g
	n := 0
	for i := 0; n < 70 && i < 600; i++ {
		x := (Hash01(i, 61)*2 - 1) * lunarLandingA * 0.88
		z := (Hash01(i, 62)*2 - 1) * lunarLandingB * 0.88
		if s.edge(x, z) > 0.86 {
			continue
		}
		r := 0.7 + 1.0*Hash01(i, 63)
		if Hash01(i, 64) < 0.12 {
			r = 2.2 + 2.0*Hash01(i, 65)
		}
		ok := math.Hypot(x, z) > 40
		for _, c := range lunarLandingClear {
			if P(x, z).Sub(c).Length() < 8+r {
				ok = false
			}
		}
		for _, c := range s.craters {
			if d := P(x, z).Sub(P(c.X, c.Z)).Length(); d < c.R*0.9 {
				ok = false
			}
		}
		for _, t := range s.tracks {
			if d, _ := s.near(t, P(x, z), 3); d < 2+r {
				ok = false
			}
		}
		if !ok {
			continue
		}
		n++
		s.rock(V(x, s.ground(x, z), z), r, i)
	}
	for k, b := range []struct{ X, Z, R float64 }{{24, -20, 4.2}, {30, -48, 3.2}, {-20, -50, 3.0}, {-56, -14, 4.6}, {-48, 24, 2.6}} {
		s.rock(V(b.X, s.ground(b.X, b.Z), b.Z), b.R, 900+k)
	}
	_ = g
}

// rock places one lumpy boulder half-buried at p.
func (s *lunarLandingSite) rock(p Vec3, r float64, seed int) {
	g := s.g
	rx := r * (0.8 + 0.4*Hash01(seed, 1))
	rz := r * (0.8 + 0.4*Hash01(seed, 2))
	ry := r * (0.55 + 0.25*Hash01(seed, 3))
	c := p.Add(V(0, ry*0.35, 0))
	main := Ellipsoid{c, V(rx, ry, rz)}
	knob := Sphere(c.Add(V(rx*0.4*(Hash01(seed, 4)*2-1), ry*0.4, rz*0.4*(Hash01(seed, 5)*2-1))), r*0.6)
	color := DarkStone
	if Hash01(seed, 6) < 0.35 {
		color = Stone
	}
	g.Use(color).FillVolume(RotateY(Union(main, knob), c.Plan(), Hash01(seed, 7)*math.Pi))
	if color == DarkStone {
		g.Paint(Charcoal, Func{Box{c.Sub(V(rx, ry, rz)), c.Add(V(rx, ry, rz))}, func(q Vec3) bool {
			return Hash01(int(math.Round(q.X*4)), int(math.Round(q.Y*4)), int(math.Round(q.Z*4)), seed) < 0.18
		}})
	}
}

// earth floats a small half-lit Earth high above the western rim; it arrives last.
func (s *lunarLandingSite) earth() {
	g := s.g
	c := V(-64, 48, -40)
	const r = 6.0
	sun := V(0.75, 0.2, -0.62)
	sun = sun.Scale(1 / sun.Length())
	g.Use(Water).FillVolume(Sphere(c, r))
	s.recolor(Box{c.Sub(V(r+1, r+1, r+1)), c.Add(V(r+1, r+1, r+1))}, func(p Vec3) (Color, bool) {
		d := p.Sub(c)
		n := d.Scale(1 / math.Max(d.Length(), 1e-9))
		lat := math.Asin(n.Y)
		lon := math.Atan2(n.Z, n.X)
		land := s.noise(lon*2.2+4, lat*2.6+4, 81) + 0.4*s.noise(lon*5, lat*5, 82)
		cloud := s.noise(lon*3.5+9, lat*4+9, 83)
		lit := n.Dot(sun) > -0.05
		switch {
		case math.Abs(lat) > 1.15:
			if lit {
				return White, true
			}
			return Slate, true
		case cloud > 0.45:
			if lit {
				return White, true
			}
			return Slate, true
		case land > 0.25:
			if !lit {
				return Charcoal, true
			}
			if math.Abs(lat) < 0.55 && land > 0.55 {
				return Sand, true
			}
			return Grass, true
		}
		if lit {
			return Water, true
		}
		return Navy, true
	})
}

// hollow removes cells buried deeper than keep layers so solids become shells.
func (s *lunarLandingSite) hollow(keep int) {
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
