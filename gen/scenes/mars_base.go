package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// MarsBase is a near-future crewed outpost on the floor of a rusty Martian valley. A
// central two-storey hub links by pressurized tubes and striped habitat modules to three
// geodesic greenhouse domes whose white-framed glass panels (some left open) reveal crop
// rows and fruit trees inside. Behind them a tall stainless starship stands on its
// landing legs and pad, its dark heat shield turned away from the base; a lattice
// communications mast with a dish crowns a layered mesa, and a squat gold cargo lander
// waits beside its crates. In front, two six-wheeled pressurized rovers leave ruts
// through the dunes toward a drilling rig and a solar farm. The red ground has dunes,
// boulders, a raised crater rim along the back and mesas, and slopes down to a low rim.
type MarsBase struct{}

func init() { Register(MarsBase{}) }

func (MarsBase) Name() string  { return "mars_base" }
func (MarsBase) Title() string { return "Mars Base" }

const (
	marsBaseA = 60.0 // Site semi-axis along X.
	marsBaseB = 50.0 // Site semi-axis along Z.
	marsBaseH = 3.0  // Valley floor height above the rim.
)

// Plan positions of the base hardware.
var (
	marsBaseHub     = P(0, 2)
	marsBaseRocket  = P(-28, -22)
	marsBaseLander  = P(43, -6)
	marsBaseMast    = P(34, -28)
	marsBaseDrill   = P(-37, 19)
	marsBaseAirlock = P(1, 26)
	marsBaseWest    = P(-26, 3)
)

type marsBaseDisk struct {
	C Vec2
	R float64
}

type marsBaseMesa struct {
	C    Vec2
	R, H float64
}

// marsBaseScaled scales a solid about the origin.
type marsBaseScaled struct {
	S Solid
	F float64
}

func (s marsBaseScaled) Contains(p Vec3) bool { return s.S.Contains(p.Scale(1 / s.F)) }
func (s marsBaseScaled) Bounds() Box {
	b := s.S.Bounds()
	return Box{b.Min.Scale(s.F), b.Max.Scale(s.F)}
}

type marsBaseDome struct {
	C Vec2
	R float64
}

// marsBaseSite carries the derived site data used while building.
type marsBaseSite struct {
	g      *Grid
	flats  []marsBaseDisk
	mesas  []marsBaseMesa
	domes  []marsBaseDome
	tracks [][]Vec2 // Rover wheel ruts, densely sampled.
	rovers []marsBaseDisk
}

func (MarsBase) Build(g *Grid) Order {
	s := &marsBaseSite{g: g}
	s.plan()
	s.terrain()
	s.boulders()
	for _, d := range s.domes {
		s.dome(d.C, d.R)
	}
	s.hub()
	s.modules()
	s.rocket(marsBaseRocket)
	s.mast(marsBaseMast)
	s.lander(marsBaseLander)
	s.drill(marsBaseDrill)
	s.solar()
	s.plant()
	s.flags()
	s.astronaut(P(-31, 9), 2.8, false)
	s.astronaut(P(-28, 25), -0.4, true)
	s.astronaut(P(5, 32), 0.6, true)
	s.astronaut(P(17, 10), 1.8, false)
	for i, t := range s.tracks[:2] {
		end, prev := t[len(t)-1], t[len(t)-8]
		dir := end.Sub(prev)
		s.rover(end, math.Atan2(dir.Y, dir.X), i)
	}
	s.hollow(2)
	return Outward(marsBaseHub)
}

func (marsBaseSite) smooth(e0, e1, v float64) float64 {
	t := math.Max(0, math.Min(1, (v-e0)/(e1-e0)))
	return t * t * (3 - 2*t)
}

// noise is smooth value noise in [-1, 1].
func (marsBaseSite) noise(x, z float64, seed int) float64 {
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
func (s *marsBaseSite) edge(x, z float64) float64 {
	e := math.Hypot(x/marsBaseA, z/marsBaseB)
	a := math.Atan2(z, x)
	return e * (1 + 0.035*math.Sin(3*a+0.4) + 0.025*math.Sin(7*a+1.3) + 0.02*s.noise(x/9, z/9, 3))
}

// flat is 1 on the leveled base footprints, fading to 0 around them.
func (s *marsBaseSite) flat(x, z float64) float64 {
	f := 0.0
	for _, d := range s.flats {
		dx, dz := x-d.C.X, z-d.C.Y
		if math.Abs(dx) > d.R+6 || math.Abs(dz) > d.R+6 {
			continue
		}
		f = math.Max(f, s.smooth(d.R+6, d.R+0.5, math.Hypot(dx, dz)))
	}
	return f
}

// mesa is the height a mesa adds: a talus apron and a cliffed, terraced top.
func (s *marsBaseSite) mesa(m marsBaseMesa, x, z float64) float64 {
	dx, dz := x-m.C.X, z-m.C.Y
	if math.Abs(dx) > m.R+6 || math.Abs(dz) > m.R+6 {
		return 0
	}
	a := math.Atan2(dz, dx)
	d := math.Hypot(dx, dz) * (1 + 0.08*math.Sin(3*a+m.C.X) + 0.06*math.Sin(5*a+m.C.Y) + 0.05*s.noise(x/3, z/3, 17))
	apron := s.smooth(m.R+5, m.R+1.5, d)
	cliff := s.smooth(m.R+1.2, m.R+0.2, d)
	cap := s.smooth(m.R-1.5, m.R-3, d)
	return m.H * (0.3*apron*(1-s.flat(x, z)) + 0.5*cliff + 0.2*cap)
}

// ground is the surface height.
func (s *marsBaseSite) ground(x, z float64) float64 {
	e := s.edge(x, z)
	base := marsBaseH * (1 - s.smooth(0.82, 1.0, e))
	flat := s.flat(x, z)
	rim := 1 - s.smooth(0.84, 1.0, e)
	free := (1 - flat) * rim
	und := (0.9*s.noise(x/14, z/14, 1) + 0.35*s.noise(x/5, z/5, 2)) * free
	mask := s.smooth(-0.2, 0.35, s.noise(x/22, z/22, 5))
	rip := 0.5 + 0.5*math.Sin((x*0.83+z*0.55)/2.0+2.2*s.noise(x/10, z/10, 6))
	dune := 1.3 * mask * rip * rip * free
	// The raised crater rim curving along the back of the valley.
	back := s.smooth(0.3, -0.35, z/marsBaseB)
	ridge := (8.5 + 2.5*s.noise(x/8, z/8, 9)) * math.Exp(-math.Pow((e-0.85)/0.075, 2)) * back * (1 - flat)
	h := base + und + dune + ridge
	for _, m := range s.mesas {
		h = math.Max(h, base+s.mesa(m, x, z))
	}
	return h
}

// spline samples a Catmull-Rom curve through the points every step units.
func (marsBaseSite) spline(pts []Vec2, step float64) []Vec2 {
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
func (marsBaseSite) offset(line []Vec2, d float64) []Vec2 {
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
func (marsBaseSite) near(line []Vec2, p Vec2, limit float64) (float64, float64) {
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

// plan lays out the leveled footprints, mesas, domes and rover tracks.
func (s *marsBaseSite) plan() {
	s.domes = []marsBaseDome{{P(14, -17), 12.5}, {P(-9, -15), 10}, {P(28, 4), 8}, {P(-15, 23), 8}}
	for _, d := range s.domes {
		s.flats = append(s.flats, marsBaseDisk{d.C, d.R + 1.5})
	}
	s.flats = append(s.flats,
		marsBaseDisk{marsBaseHub, 8},
		marsBaseDisk{marsBaseRocket, 12},
		marsBaseDisk{marsBaseLander, 9},
		marsBaseDisk{marsBaseDrill, 8},
		marsBaseDisk{P(29, 26), 9},
		marsBaseDisk{P(-10, -32), 8},
		marsBaseDisk{P(-14, 3), 5}, marsBaseDisk{P(-22, 3), 5},
		marsBaseDisk{P(1, 15), 5}, marsBaseDisk{marsBaseAirlock, 5},
		marsBaseDisk{P(13, 3), 5},
		marsBaseDisk{P(-8, 14), 5},
	)
	s.mesas = []marsBaseMesa{{P(-46, -4), 8, 12}, {marsBaseMast, 7.5, 11}}

	// Ruts: the first two end at the rovers.
	pairs := func(pts ...Vec2) {
		line := s.spline(pts, 0.5)
		s.tracks = append(s.tracks, line)
	}
	pairs(P(46, 8), P(38, 14.5), P(24, 15.5), P(15, 21), P(12, 30))
	pairs(P(4, 33), P(-6, 38), P(-16, 38.5), P(-25, 33.5))
	pairs(P(12, 30), P(8, 34), P(4, 33), P(1, 30))
	pairs(P(-29, 5), P(-31, 12), P(-35, 17))
	pairs(P(-16, 38.5), P(-8, 42), P(8, 41), P(19, 37), P(30, 37))
}

// raster adds cells in the box for which f reports a color, replacing existing ones.
func (s *marsBaseSite) raster(b Box, f func(Vec3) (Color, bool)) {
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
func (s *marsBaseSite) recolor(b Box, f func(Vec3) (Color, bool)) {
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

// terrain rasterizes the ground shell, thick enough under slopes and cliffs to stay closed.
func (s *marsBaseSite) terrain() {
	A, B := marsBaseA*1.12, marsBaseB*1.12
	box := Box{V(-A, -3, -B), V(A, marsBaseH+16, B)}
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
				d := sp * 0.75
				lo := math.Min(math.Min(s.ground(x+d, z), s.ground(x-d, z)), math.Min(s.ground(x, z+d), s.ground(x, z-d)))
				col = column{h, math.Min(16, math.Max(sp*1.05, h-lo+sp*0.8))}
			}
			cache[key] = col
		}
		if math.IsNaN(col.h) || y > col.h+0.02 || y < col.h-col.t {
			return 0, false
		}
		return s.soil(x, y, z, col.h), true
	})
}

// soil picks the ground color: rust and brick mottling, sand dune crests, layered mesa
// strata, dark rover ruts and a scorched landing zone.
func (s *marsBaseSite) soil(x, y, z, h float64) Color {
	ix, iy, iz := int(math.Round(x*4)), int(math.Round(y*4)), int(math.Round(z*4))
	top := y > h-0.6
	p := P(x, z)
	e := s.edge(x, z)
	base := marsBaseH * (1 - s.smooth(0.82, 1.0, e))
	v := Hash01(ix, iy, iz, 9)
	// Mesa cliffs show horizontal strata.
	if y > base+1.2 && h > base+2.5 {
		mesa := false
		for _, m := range s.mesas {
			if s.mesa(m, x, z) > 2 {
				mesa = true
			}
		}
		if mesa {
			if top && h-base > 10.5 {
				if v < 0.12 {
					return DarkStone
				}
				return Terracotta
			}
			band := int(math.Floor((y-base)/1.3 + 0.4*s.noise(x/6, z/6, 13)))
			switch ((band % 5) + 5) % 5 {
			case 0, 3:
				return Brick
			case 2:
				return Sand
			}
			if v < 0.1 {
				return Brick
			}
			return Terracotta
		}
	}
	if top {
		for _, t := range s.tracks {
			if d, run := s.near(t, p, 3.2); d > 1.5 && d < 3.0 {
				if math.Mod(run, 1.2) < 0.45 {
					return Charcoal
				}
				return DarkStone
			}
		}
		// Scorched blast zone around the landing pad.
		if r := p.Sub(marsBaseRocket).Length(); r < 18 && Hash01(ix, iz, 21) < 0.55*s.smooth(18, 11, r) {
			return Charcoal
		}
	}
	// Raised crater rim: darker brick.
	ridge := h - base
	if s.flat(x, z) < 0.5 && ridge > 3.5 && z < 0 {
		if v < 0.55 {
			return Brick
		}
		if v < 0.62 {
			return DarkStone
		}
		return Terracotta
	}
	n := s.noise(x/12, z/12, 7) + 0.5*s.noise(x/4, z/4, 8)
	if top {
		mask := s.smooth(-0.2, 0.35, s.noise(x/22, z/22, 5))
		rip := 0.5 + 0.5*math.Sin((x*0.83+z*0.55)/2.0+2.2*s.noise(x/10, z/10, 6))
		if mask*(1-s.flat(x, z)) > 0.35 && rip > 0.72 {
			if v < 0.8 {
				return Sand
			}
			return Terracotta
		}
	}
	switch {
	case n < -0.35 && v < 0.75:
		return Brick
	case v < 0.06:
		return Brick
	case v < 0.09:
		return DarkStone
	case v > 0.965:
		return Sand
	}
	return Terracotta
}

// boulders scatters rocks across the open ground.
func (s *marsBaseSite) boulders() {
	g := s.g
	n := 0
	for i := 0; n < 60 && i < 800; i++ {
		x := (Hash01(i, 61)*2 - 1) * marsBaseA * 0.9
		z := (Hash01(i, 62)*2 - 1) * marsBaseB * 0.9
		if s.edge(x, z) > 0.86 || s.flat(x, z) > 0.05 {
			continue
		}
		r := 0.8 + 1.0*Hash01(i, 63)
		if Hash01(i, 64) < 0.15 {
			r = 2.2 + 1.8*Hash01(i, 65)
		}
		ok := true
		for _, t := range s.tracks {
			if d, _ := s.near(t, P(x, z), 6); d < 3.5+r {
				ok = false
			}
		}
		for _, m := range s.mesas {
			if s.mesa(m, x, z) > 0.5 {
				ok = false
			}
		}
		if P(x, z).Sub(P(29, 26)).Length() < 15 {
			ok = false // Solar farm.
		}
		if !ok {
			continue
		}
		n++
		h := s.ground(x, z)
		col := Brick
		switch k := Hash01(i, 66); {
		case k < 0.35:
			col = DarkStone
		case k < 0.45:
			col = Charcoal
		}
		g.Use(col).FillVolume(RotateY(Ellipsoid{V(x, h+r*0.15, z), V(r*1.25, r*0.7, r*0.9)}, P(x, z), Hash01(i, 67)*math.Pi))
	}
}

// panel classifies a point on a geodesic dome shell: frame, glass, or open panel.
func (s *marsBaseSite) panel(p, ctr Vec3, R float64, seed int) (Color, bool) {
	d := p.Sub(ctr)
	r := d.Length()
	lat := math.Asin(math.Max(-1, math.Min(1, d.Y/r)))
	lat = math.Max(lat, 0)
	if lat > math.Pi/2-1.4/R {
		return White, true
	}
	lon := math.Atan2(d.Z, d.X)
	size := 5.6
	quarter := math.Pi / 2 * R
	bands := math.Max(2, math.Round(quarter/size))
	bs := quarter / bands
	arc := lat * R
	i := math.Floor(arc / bs)
	f := arc/bs - i
	mid := (i + 0.5) * bs / R
	n := math.Max(5, math.Round(2*math.Pi*R*math.Cos(mid)/size))
	t := (lon+math.Pi)/(2*math.Pi)*n + 0.5*i
	w := 0.42
	wt := w / math.Max(0.5, 2*math.Pi*R*math.Cos(lat)/n)
	wf := w / bs
	dist := func(v float64) float64 { v -= math.Floor(v); return math.Min(v, 1-v) }
	if f < wf || f > 1-wf || dist(t-0.5*f) < wt || dist(t+0.5*f) < wt {
		return White, true
	}
	k1, k2 := int(math.Floor(t-0.5*f)), int(math.Floor(t+0.5*f))
	if i < bands-1 && Hash01(int(i), k1, k2, seed) < 0.35 {
		return 0, false
	}
	return Glass, true
}

// dome builds a geodesic greenhouse with crop beds and fruit trees inside.
func (s *marsBaseSite) dome(c Vec2, R float64) {
	g := s.g
	y0 := marsBaseH
	wall := 1.6
	cy := y0 + wall
	ctr := c.At(cy)
	th := 1.15
	seed := int(c.X*10 + c.Y)
	// Foundation ring with dark portholes.
	g.Use(Stone).Fill(Cylinder{c, R + 0.4, R - th, y0 - 1, cy})
	g.Use(DarkStone).Fill(Cylinder{c, R + 1.2, R - th, y0 - 1, y0 + 0.4})
	// Floor and crops.
	in := R - th
	g.Use(Sand).Fill(Cylinder{c, in, 0, y0 - 0.5, y0 + 0.3})
	s.raster(Box{V(c.X-in, y0+0.3, c.Y-in), V(c.X+in, y0+2.2, c.Y+in)}, func(p Vec3) (Color, bool) {
		dx, dz := p.X-c.X, p.Z-c.Y
		r := math.Hypot(dx, dz)
		if r > in-1.2 || math.Abs(dx) < 1.4 {
			return 0, false
		}
		row := math.Floor((dz + 100) / 3.0)
		if math.Mod(dz+100, 3.0) > 1.9 {
			return 0, false
		}
		ix, iy, iz := int(math.Round(p.X*4)), int(math.Round(p.Y*4)), int(math.Round(p.Z*4))
		hgt := 1.0 + 0.6*Hash01(int(row), int(math.Floor(p.X)), seed)
		if p.Y > y0+0.3+hgt {
			return 0, false
		}
		col := Grass
		if int(row)%2 == 0 {
			col = Foliage
		}
		if v := Hash01(ix, iy, iz, 31); p.Y > y0+0.8 && v < 0.09 {
			if int(row)%3 == 0 {
				return Red, true
			}
			return Gold, true
		}
		return col, true
	})
	for _, side := range []float64{-1, 1} {
		tc := P(c.X+side*0.0, c.Y+side*R*0.42)
		if R < 8 {
			tc = P(c.X, c.Y+side*R*0.38)
		}
		g.Use(Sand).Fill(Cylinder{tc, 2.0, 0, y0, y0 + 0.6})
		g.Use(Wood).Fill(Capsule{tc.At(y0), tc.At(y0 + R*0.38), 0.55})
		g.Use(Foliage).FillVolume(Ellipsoid{tc.At(y0 + R*0.45), V(R*0.24, R*0.2, R*0.24)})
		g.Use(Grass).Paint(Grass, Ellipsoid{tc.At(y0 + R*0.55), V(R*0.24, R*0.12, R*0.24)})
	}
	// Shell.
	s.raster(Box{V(c.X-R, cy, c.Y-R), V(c.X+R, cy+R, c.Y+R)}, func(p Vec3) (Color, bool) {
		d := p.Sub(ctr).Length()
		if d > R || d <= R-th {
			return 0, false
		}
		return s.panel(p, ctr, R, seed)
	})
	// Ventilation cap.
	g.Use(Stone).Fill(Cylinder{c, 1.3, 0, cy + R - 0.6, cy + R + 0.6})
}

// hub is the two-storey central habitat with a low domed roof.
func (s *marsBaseSite) hub() {
	g := s.g
	c := marsBaseHub
	y0 := marsBaseH
	R, h := 6.5, 10.0
	g.Use(DarkStone).Fill(Cylinder{c, R + 1.4, 0, y0 - 1, y0 + 0.7})
	g.Use(White).Fill(Cylinder{c, R, 0, y0, y0 + h})
	g.Use(White).FillVolume(Intersect(Ellipsoid{c.At(y0 + h), V(R, 2.6, R)}, Block(V(c.X-R, y0+h, c.Y-R), V(c.X+R, y0+h+3, c.Y+R))))
	g.Paint(Navy, Cylinder{c, R + 1, 0, y0 + h - 1.3, y0 + h - 0.3})
	g.Paint(Red, Cylinder{c, R + 1, 0, y0 + 1.0, y0 + 1.6})
	g.Paint(Stone, Cylinder{c, R + 1, 0, y0 + 5.0, y0 + 5.5})
	win := func(ya, yb float64) Solid {
		return Func{Box{V(c.X-R-1, ya, c.Y-R-1), V(c.X+R+1, yb, c.Y+R+1)}, func(p Vec3) bool {
			a := math.Atan2(p.Z-c.Y, p.X-c.X)
			k := math.Mod(a/(2*math.Pi)*14+100, 1)
			return k > 0.25 && k < 0.75
		}}
	}
	g.Paint(Charcoal, win(y0+2.4, y0+4.2), win(y0+6.4, y0+8.2))
	// Roof: hatch, small dish and whip antennas.
	g.Use(Stone).Fill(Cylinder{c, 1.6, 0, y0 + h + 2, y0 + h + 3.4})
	dc := P(c.X+3, c.Y+1)
	g.Use(Stone).Fill(Capsule{dc.At(y0 + h + 1), dc.At(y0 + h + 3), 0.5})
	g.Use(White).Fill(s.dish(dc.At(y0+h+3.2), V(0.5, 1, 0.6), 2.0, 0.7, 0.8))
	for _, a := range []float64{2.3, 3.4} {
		q := Polar(c, 3.5, a)
		g.Use(Stone).Fill(Capsule{q.At(y0 + h + 1), q.At(y0 + h + 6), 0.45})
		g.Use(Red).Fill(Sphere(q.At(y0+h+6.2), 0.6))
	}
}

// dish is a shallow paraboloid shell opening along axis.
func (marsBaseSite) dish(c, axis Vec3, radius, depth, thick float64) Solid {
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

// axial returns a solid for points whose (axial position, radial distance, height
// above the axis) relative to segment a→b satisfy f.
func (marsBaseSite) axial(a, b Vec3, r float64, f func(t, rad, up float64) bool) Solid {
	ab := b.Sub(a)
	l := ab.Length()
	u := ab.Scale(1 / l)
	return Func{BoxOf(a, b).Grow(r + 0.5), func(p Vec3) bool {
		d := p.Sub(a)
		t := d.Dot(u)
		q := d.Sub(u.Scale(t))
		return f(t, q.Length(), q.Y)
	}}
}

// module is a striped horizontal habitat cylinder with rounded ends, on saddles.
func (s *marsBaseSite) module(a, b Vec2) {
	g := s.g
	y0 := marsBaseH
	r := 3.3
	yc := y0 + r + 0.5
	A, B := a.At(yc), b.At(yc)
	l := b.Sub(a).Length()
	u := b.Sub(a).Scale(1 / l)
	for t := 2.0; t < l-1; t += 4.5 {
		q := a.Add(u.Scale(t))
		g.Use(DarkStone).Fill(BlockAt(q, 2.2, 2.2, y0-0.5, yc-1.5))
	}
	g.Use(White).FillVolume(Capsule{A, B, r})
	g.Paint(Navy, s.axial(A, B, r, func(t, rad, up float64) bool { return up > -0.5 && up < 0.4 }))
	g.Paint(Red, s.axial(A, B, r, func(t, rad, up float64) bool { return up > -1.6 && up < -1.0 }))
	g.Paint(Navy, s.axial(A, B, r, func(t, rad, up float64) bool {
		return (t > 0.2 && t < 1.0) || (t > l-1.0 && t < l-0.2)
	}))
	g.Paint(Charcoal, s.axial(A, B, r, func(t, rad, up float64) bool {
		k := math.Mod(t+0.5, 2.6)
		return t > 1.5 && t < l-1.5 && up > 0.9 && up < 2.1 && k > 0.6 && k < 2.0
	}))
}

// tube is a ribbed pressurized connector between two plan points.
func (s *marsBaseSite) tube(a, b Vec2) {
	g := s.g
	y := marsBaseH + 2.2
	A, B := a.At(y), b.At(y)
	g.Use(White).FillVolume(Capsule{A, B, 1.8})
	g.Paint(Stone, s.axial(A, B, 2, func(t, rad, up float64) bool { return math.Mod(t, 2.5) < 0.6 }))
	l := b.Sub(a).Length()
	u := b.Sub(a).Scale(1 / l)
	for t := 1.5; t < l; t += 5 {
		q := a.Add(u.Scale(t))
		g.Use(DarkStone).Fill(BlockAt(q, 1.4, 1.4, marsBaseH-0.5, y-1))
	}
}

// airlock is a short vertical chamber with an outward door and a ramp.
func (s *marsBaseSite) airlock(c, out Vec2) {
	g := s.g
	y0 := marsBaseH
	g.Use(DarkStone).Fill(Cylinder{c, 3.4, 0, y0 - 0.5, y0 + 0.6})
	g.Use(White).Fill(Cylinder{c, 2.8, 0, y0, y0 + 6.5})
	g.Use(White).FillVolume(Intersect(Sphere(c.At(y0+6.5), 2.8), Block(V(c.X-3, y0+6.5, c.Y-3), V(c.X+3, y0+9.5, c.Y+3))))
	g.Paint(Red, Cylinder{c, 4, 0, y0 + 5.6, y0 + 6.2})
	g.Paint(Navy, Cylinder{c, 4, 0, y0 + 0.6, y0 + 1.1})
	g.Use(Gold).Fill(Sphere(c.At(y0+9.0), 0.7))
	// Door: red frame, dark opening.
	n := out.Scale(1 / out.Length())
	side := P(-n.Y, n.X)
	door := func(w, hb, ht float64) Solid {
		return Func{Box{V(c.X-4, y0, c.Y-4), V(c.X+4, y0+ht+1, c.Y+4)}, func(p Vec3) bool {
			d := p.Plan().Sub(c)
			return d.Dot(n) > 1.0 && math.Abs(d.Dot(side)) < w && p.Y > hb && p.Y < ht
		}}
	}
	g.Paint(Red, door(1.9, y0+0.6, y0+5.4))
	g.Paint(Charcoal, door(1.2, y0+1.0, y0+4.7))
	// Ramp down to the ground.
	g.Use(Stone).Fill(Func{Box{V(c.X-8, y0-1, c.Y-8), V(c.X+8, y0+1.2, c.Y+8)}, func(p Vec3) bool {
		d := p.Plan().Sub(c)
		a := d.Dot(n) - 2.6
		if a < 0 || a > 4 || math.Abs(d.Dot(side)) > 1.8 {
			return false
		}
		return p.Y <= y0+1.0-a*0.25 && p.Y > y0-1
	}})
}

// modules places the habitat modules, tubes and airlocks around the hub.
func (s *marsBaseSite) modules() {
	hub := marsBaseHub
	s.module(P(-22, 3), P(-8, 3))
	s.module(P(1, 9), P(1, 21))
	s.module(P(8, 3), P(18, 3.5))
	s.airlock(marsBaseWest, P(-1, 0))
	s.airlock(marsBaseAirlock, P(0, 1))
	// Tubes from the hub to the two large domes.
	for _, d := range s.domes[:2] {
		dir := d.C.Sub(hub)
		dir = dir.Scale(1 / dir.Length())
		s.tube(hub.Add(dir.Scale(5.5)), d.C.Sub(dir.Scale(d.R-0.6)))
	}
	// Tube between the two big domes.
	a, b := s.domes[1], s.domes[0]
	dir := b.C.Sub(a.C)
	dir = dir.Scale(1 / dir.Length())
	s.tube(a.C.Add(dir.Scale(a.R-0.6)), b.C.Sub(dir.Scale(b.R-0.6)))
	// Tube from the front-left dome to the south module.
	d := s.domes[3]
	end := P(1, 19)
	dir = end.Sub(d.C)
	dir = dir.Scale(1 / dir.Length())
	s.tube(d.C.Add(dir.Scale(d.R-0.6)), end)
}

// astronaut is a suited figure facing angle, optionally waving.
func (s *marsBaseSite) astronaut(at Vec2, angle float64, wave bool) {
	g := s.g
	y0 := s.ground(at.X, at.Y) - 0.1
	place := func(sol Solid) Solid { return RotateY(Translate(marsBaseScaled{sol, 1.3}, at.At(y0)), at, angle) }
	g.Use(White).Fill(place(Capsule{V(0, 0.3, -0.38), V(0, 1.4, -0.38), 0.4}), place(Capsule{V(0, 0.3, 0.38), V(0, 1.4, 0.38), 0.4}))
	g.Use(White).FillVolume(place(Ellipsoid{V(0, 2.1, 0), V(0.55, 0.8, 0.75)}))
	g.Use(Stone).Fill(place(Block(V(-0.95, 1.6, -0.6), V(-0.4, 2.8, 0.6))))
	g.Use(White).Fill(place(Sphere(V(0.05, 3.1, 0), 0.6)))
	g.Paint(Gold, place(Block(V(0.3, 2.85, -0.45), V(0.8, 3.35, 0.45))))
	g.Use(White).Fill(place(Capsule{V(0, 2.6, -0.8), V(0.3, 1.5, -0.95), 0.32}))
	if wave {
		g.Use(White).Fill(place(Capsule{V(0, 2.6, 0.8), V(0.2, 3.9, 1.3), 0.32}))
	} else {
		g.Use(White).Fill(place(Capsule{V(0, 2.6, 0.8), V(0.3, 1.5, 0.95), 0.32}))
	}
	g.Paint(Red, place(Block(V(-0.7, 2.3, -0.8), V(0.7, 2.5, 0.8))))
	g.Use(Charcoal).Fill(place(Block(V(-0.2, 0, -0.75), V(0.45, 0.35, 0.75))))
}

// rocket is a landed stainless starship on its pad: six legs, engine skirt, belly heat
// shield facing -X, aft and forward flaps, and a ring of cabin windows.
func (s *marsBaseSite) rocket(c Vec2) {
	g := s.g
	y0 := marsBaseH + 0.9
	R := 4.5
	// Pad with markings and scorch.
	g.Use(DarkStone).Fill(Cylinder{c, 10, 0, marsBaseH - 1.5, y0})
	g.Paint(Stone, Cylinder{c, 9.2, 0, y0 - 0.4, y0 + 0.1})
	g.Paint(Gold, Cylinder{c, 7.8, 6.9, y0 - 0.4, y0 + 0.1})
	g.Paint(White, Cylinder{c, 10.2, 9.2, y0 - 0.4, y0 + 0.1})
	s.recolor(Box{V(c.X-7, y0-0.6, c.Y-7), V(c.X+7, y0+0.2, c.Y+7)}, func(p Vec3) (Color, bool) {
		r := p.Plan().Sub(c).Length()
		if r < 6.5 && Hash01(int(math.Round(p.X*4)), int(math.Round(p.Z*4)), 41) < 0.75*s.smooth(6.5, 3, r) {
			return Charcoal, true
		}
		return 0, false
	})
	// Legs and skirt.
	for k := 0; k < 6; k++ {
		a := float64(k)*math.Pi/3 + math.Pi/6
		foot := Polar(c, R+2.6, a)
		g.Use(DarkStone).Fill(Capsule{Polar(c, R-0.4, a).At(y0 + 7), foot.At(y0 + 0.9), 0.6})
		g.Use(Charcoal).Fill(Cylinder{foot, 1.3, 0, y0, y0 + 0.7})
	}
	g.Use(DarkStone).FillVolume(Frustum{c, 3.2, R, y0 + 2.2, y0 + 4.4})
	for k := 0; k < 3; k++ {
		e := Polar(c, 1.7, float64(k)*2*math.Pi/3+0.5)
		g.Use(Charcoal).FillVolume(Frustum{e, 1.3, 0.7, y0 + 1.0, y0 + 2.4})
	}
	// Body and nose.
	yb, yt, L := y0+4.4, y0+37.0, 11.0
	g.Use(White).FillVolume(Cylinder{c, R, 0, yb, yt})
	nr := func(y float64) float64 {
		if y < yt {
			return R
		}
		k := (y - yt) / L
		if k > 1 {
			return 0
		}
		return R * math.Sqrt(1-k*k)
	}
	g.Use(White).FillVolume(Func{Box{V(c.X-R, yt, c.Y-R), V(c.X+R, yt+L, c.Y+R)}, func(p Vec3) bool {
		return p.Plan().Sub(c).Length() <= nr(p.Y)
	}})
	for y := yb + 3.5; y < yt; y += 4.5 {
		g.Paint(Stone, Cylinder{c, R + 1, 0, y, y + 0.5})
	}
	// Heat shield on the belly (-X).
	g.Paint(Charcoal, Func{Box{V(c.X-R-1, yb, c.Y-R-1), V(c.X+R+1, yt+L, c.Y+R+1)}, func(p Vec3) bool {
		return p.X-c.X < -0.5
	}})
	// Flaps at ±Z on the belly side.
	flap := func(ya, yb, w0, w1 float64) Solid {
		return Func{Box{V(c.X-2, ya, c.Y-R-w0-1), V(c.X, yb, c.Y+R+w0+1)}, func(p Vec3) bool {
			dx, dz := p.X-c.X, math.Abs(p.Z-c.Y)
			t := (p.Y - ya) / (yb - ya)
			w := w0 + (w1-w0)*t
			rb := nr(p.Y)
			return dx > -1.7 && dx < -0.4 && dz > rb-0.8 && dz < rb+w
		}}
	}
	g.Use(Charcoal).Fill(flap(yb+0.5, yb+11, 3.6, 1.6))
	g.Use(Charcoal).Fill(flap(yt-2.5, yt+5.5, 2.8, 1.0))
	// Cabin windows on the leeward side.
	g.Paint(Charcoal, Func{Box{V(c.X, yt-3.3, c.Y-R-1), V(c.X+R+1, yt-1.6, c.Y+R+1)}, func(p Vec3) bool {
		a := math.Atan2(p.Z-c.Y, p.X-c.X)
		if math.Abs(a) > 1.1 {
			return false
		}
		k := math.Mod(a*R/2.2+100, 1)
		return k > 0.2 && k < 0.8
	}})
}

// mast is a red-and-white lattice tower with a dish and beacon on the mesa top.
func (s *marsBaseSite) mast(c Vec2) {
	g := s.g
	y0 := s.ground(c.X, c.Y) - 0.4
	h := 19.0
	pos := func(k int, y float64) Vec3 {
		t := (y - y0) / h
		return Polar(c, 3.0-1.9*t, float64(k)*2*math.Pi/3+0.4).At(y)
	}
	// Equipment shelter.
	sh := P(c.X+4.5, c.Y+3.5)
	g.Use(White).Fill(BlockAt(sh, 4.5, 3.5, y0, y0+3.2))
	g.Paint(Navy, BlockAt(sh, 5, 4, y0+2.2, y0+2.8))
	g.Use(DarkStone).Fill(Cylinder{c, 3.8, 0, y0 - 0.6, y0 + 0.6})
	for k := 0; k < 3; k++ {
		g.Use(White).Fill(Capsule{pos(k, y0), pos(k, y0+h), 0.55})
		for y := y0; y < y0+h-0.5; y += 3.2 {
			g.Use(White).Fill(Capsule{pos(k, y), pos((k+1)%3, y+3.2), 0.42}, Capsule{pos((k+1)%3, y), pos(k, y+3.2), 0.42})
		}
	}
	for y := y0; y < y0+h; y += 6.4 {
		g.Paint(Red, Cylinder{c, 4, 0, y + 3.2, y + 6.4})
	}
	top := c.At(y0 + h)
	g.Use(DarkStone).Fill(Cylinder{c, 2.0, 0, y0 + h, y0 + h + 0.8})
	g.Use(Stone).Fill(Capsule{top, top.Add(V(0, 5, 0)), 0.5})
	g.Use(Red).Fill(Sphere(top.Add(V(0, 5.6, 0)), 0.8))
	axis := V(0.55, 0.62, 0.56)
	axis = axis.Scale(1 / axis.Length())
	dc := top.Add(V(0.6, 2.4, 0.6))
	g.Use(White).Fill(s.dish(dc, axis, 4.2, 1.4, 0.9))
	g.Use(Stone).Fill(Capsule{dc, dc.Add(axis.Scale(3.4)), 0.45})
	g.Use(Gold).Fill(Sphere(dc.Add(axis.Scale(3.6)), 0.7))
	g.Use(Charcoal).Fill(Capsule{top.Add(V(0, 0.8, 0)), dc.Sub(axis.Scale(0.3)), 0.6})
	// Second smaller dish on a side arm.
	sd := pos(0, y0+11).Add(V(1.5, 0, 1.5))
	g.Use(Stone).Fill(Capsule{pos(0, y0+11), sd, 0.45})
	g.Use(White).Fill(s.dish(sd.Add(V(0.6, 0.3, 0.6)), V(1, 0.5, 0.3), 1.8, 0.6, 0.8))
}

// lander is a squat gold-foil cargo lander on four splayed legs with a ramp and crates.
func (s *marsBaseSite) lander(c Vec2) {
	g := s.g
	y0 := marsBaseH
	g.Use(DarkStone).Fill(Cylinder{c, 7.5, 0, y0 - 1, y0 + 0.3})
	g.Paint(Charcoal, Cylinder{c, 3.5, 0, y0 - 0.3, y0 + 0.4})
	w, cut := 4.4, 1.6
	oct := []Vec2{}
	for _, v := range []Vec2{P(w, -(w - cut)), P(w, w-cut), P(w-cut, w), P(-(w - cut), w), P(-w, w-cut), P(-w, -(w - cut)), P(-(w - cut), -w), P(w-cut, -w)} {
		oct = append(oct, c.Add(v))
	}
	for k := 0; k < 4; k++ {
		a := float64(k)*math.Pi/2 + math.Pi/4
		foot := Polar(c, 7.5, a)
		g.Use(Stone).Fill(Capsule{Polar(c, 4.2, a).At(y0 + 5), foot.At(y0 + 0.9), 0.55})
		g.Use(Stone).Fill(Capsule{Polar(c, 4.6, a).At(y0 + 3.3), foot.At(y0 + 0.9), 0.45})
		g.Use(Charcoal).Fill(Cylinder{foot, 1.3, 0, y0 + 0.2, y0 + 0.8})
	}
	for k := 0; k < 4; k++ {
		e := Polar(c, 2.0, float64(k)*math.Pi/2+math.Pi/4)
		g.Use(Charcoal).FillVolume(Frustum{e, 1.1, 0.6, y0 + 1.5, y0 + 3.0})
	}
	g.Use(Gold).Fill(Prism{oct, y0 + 3.0, y0 + 8.5})
	g.Use(White).Fill(Prism{oct, y0 + 8.5, y0 + 9.6})
	g.Use(White).FillVolume(Frustum{c, 3.4, 1.5, y0 + 9.6, y0 + 11.5})
	g.Use(Stone).Fill(Capsule{c.At(y0 + 11.5), c.At(y0 + 13.5), 0.5})
	g.Use(Gold).FillVolume(Sphere(Polar(c, 3.6, 0.6).At(y0+10.6), 1.4), Sphere(Polar(c, 3.6, 0.6+math.Pi).At(y0+10.6), 1.4))
	// Panel seams on the foil.
	g.Paint(DarkStone, Prism{oct, y0 + 5.6, y0 + 6.0})
	// Cargo door facing -X (toward the base) and ramp.
	g.Paint(Charcoal, Block(V(c.X-w-1, y0+3.8, c.Y-1.6), V(c.X-w+1, y0+7.6, c.Y+1.6)))
	g.Use(Stone).Fill(Func{Box{V(c.X-w-7, y0-0.5, c.Y-1.8), V(c.X-w+0.5, y0+4, c.Y+1.8)}, func(p Vec3) bool {
		a := (c.X - w + 0.3) - p.X
		top := y0 + 3.6 - a*0.55
		return p.Y <= top && p.Y >= top-0.9 && p.Y >= y0-0.5
	}})
	// Crates.
	for i, q := range []Vec2{P(c.X-10, c.Y-5), P(c.X-12.5, c.Y-5.5), P(c.X-11, c.Y-8), P(c.X-11.2, c.Y-6.5)} {
		col := White
		if i%2 == 1 {
			col = Navy
		}
		y := y0
		if i == 3 {
			y = y0 + 2.2
		}
		g.Use(col).Fill(BlockAt(q, 2.2, 2.2, y, y+2.2))
	}
}

// drill is a red derrick on a platform with drawworks, pipe rack and tailings pile.
func (s *marsBaseSite) drill(c Vec2) {
	g := s.g
	y0 := marsBaseH
	g.Use(DarkStone).Fill(BlockAt(c, 10, 10, y0-0.5, y0+1.0))
	g.Paint(Stone, BlockAt(c, 9, 9, y0+0.6, y0+1.0))
	h := 19.0
	pos := func(k int, y float64) Vec3 {
		t := (y - y0 - 1) / h
		half := 3.6 - 2.6*t
		sx := []float64{-1, 1, 1, -1}[k]
		sz := []float64{-1, -1, 1, 1}[k]
		return V(c.X+sx*half, y, c.Y+sz*half)
	}
	for k := 0; k < 4; k++ {
		g.Use(Red).Fill(Capsule{pos(k, y0+1), pos(k, y0+1+h), 0.55})
		for y := y0 + 1; y < y0+h-0.5; y += 3.6 {
			g.Use(White).Fill(Capsule{pos(k, y), pos((k+1)%4, y+3.6), 0.4}, Capsule{pos((k+1)%4, y), pos(k, y+3.6), 0.4})
		}
	}
	g.Use(Gold).Fill(BlockAt(c, 3.2, 3.2, y0+h+0.6, y0+h+2.2))
	g.Use(Stone).Fill(Cylinder{c, 0.6, 0, y0 - 0.5, y0 + h})
	g.Use(Charcoal).Fill(BlockAt(c, 2.6, 2.6, y0+5, y0+7))
	g.Use(Charcoal).Fill(BlockAt(P(c.X-3, c.Y+3.2), 3, 2.4, y0+1, y0+3.4))
	g.Use(White).Fill(BlockAt(P(c.X+2.8, c.Y+3.0), 3.2, 2.6, y0+1, y0+3.8))
	g.Paint(Navy, BlockAt(P(c.X+2.8, c.Y+3.0), 3.6, 3.0, y0+2.8, y0+3.3))
	// Pipe rack.
	for i := 0; i < 3; i++ {
		y := y0 + 0.5 + 1.0*float64(i)
		z := c.Y - 7 + 1.0*float64(i%2)
		g.Use(Stone).Fill(Capsule{V(c.X-4, y, z), V(c.X+5, y, z), 0.5}, Capsule{V(c.X-4, y, z+1.1), V(c.X+5, y, z+1.1), 0.5})
	}
	// Tailings and a storage tank.
	g.Use(Brick).FillVolume(Ellipsoid{P(c.X+8, c.Y-2).At(y0), V(4, 2.6, 3.4)})
	g.Paint(DarkStone, Ellipsoid{P(c.X+8, c.Y-2).At(y0 + 2.4), V(2.4, 0.6, 2.0)})
	g.Use(White).FillVolume(Capsule{V(c.X-8, y0+2, c.Y-3), V(c.X-8, y0+2, c.Y+3), 1.8})
	g.Paint(Red, Cylinder{P(c.X-8, c.Y), 3, 0, y0 + 2.6, y0 + 3.1})
}

// solar is a farm of tilted navy panel rows on posts, facing +Z.
func (s *marsBaseSite) solar() {
	g := s.g
	tilt := 0.5
	n := V(0, math.Cos(tilt), math.Sin(tilt))
	along := V(0, -math.Sin(tilt), math.Cos(tilt)) // Up-slope direction is -along.
	for row := 0; row < 3; row++ {
		for col := 0; col < 2; col++ {
			cx := 22 + 9.2*float64(col) - 1.5*float64(row)
			cz := 19 + 5.6*float64(row)
			hgt := s.ground(cx, cz)
			ctr := V(cx, hgt+2.6, cz)
			half := 4.3
			slab := Func{Box{ctr.Sub(V(half+0.5, 2.5, 2.5)), ctr.Add(V(half+0.5, 2.5, 2.5))}, func(p Vec3) bool {
				d := p.Sub(ctr)
				return math.Abs(d.Dot(n)) < 0.45 && math.Abs(d.Dot(along)) < 2.2 && math.Abs(d.X) < half
			}}
			for _, x := range []float64{-half + 1.2, half - 1.2} {
				q := V(cx+x, 0, cz)
				g.Use(Stone).Fill(Capsule{V(q.X, s.ground(q.X, q.Z)-0.5, q.Z), V(q.X, ctr.Y-0.4, q.Z), 0.5})
			}
			g.Use(Navy).Fill(slab)
			g.Paint(White, Func{slab.Box, func(p Vec3) bool {
				d := p.Sub(ctr)
				return slab.F(p) && (math.Abs(d.Dot(along)) > 1.75 || math.Abs(d.X) > half-0.5)
			}})
			g.Paint(Slate, Func{slab.Box, func(p Vec3) bool {
				d := p.Sub(ctr)
				k := math.Mod(d.X+100, 2.3)
				return slab.F(p) && k < 0.35
			}})
		}
	}
	// Inverter cabinet.
	q := P(42, 22)
	h := s.ground(q.X, q.Y)
	g.Use(White).Fill(BlockAt(q, 3, 2, h-0.5, h+2.6))
	g.Paint(Navy, BlockAt(q, 3.4, 2.4, h+1.6, h+2.1))
}

// plant is the propellant plant behind the domes: spherical tanks on legs, a process
// building with radiator fins, and a pipe run to the landing pad.
func (s *marsBaseSite) plant() {
	g := s.g
	y0 := marsBaseH
	b := P(-4, -33)
	g.Use(DarkStone).Fill(BlockAt(P(-10, -32), 20, 8, y0-0.6, y0+0.5))
	g.Use(White).Fill(BlockAt(b, 6, 4.5, y0, y0+4.5))
	g.Paint(Navy, BlockAt(b, 6.6, 5, y0+3.2, y0+3.8))
	g.Paint(Charcoal, BlockAt(P(b.X, b.Y+2.25), 1.6, 0.8, y0+0.5, y0+2.6))
	for i := 0; i < 4; i++ {
		x := b.X - 2.2 + 1.5*float64(i)
		g.Use(Stone).Fill(Block(V(x-0.35, y0+4.5, b.Y-2), V(x+0.35, y0+7, b.Y+2)))
	}
	for i, x := range []float64{-17, -12} {
		c := P(x, -32)
		yc := y0 + 4.2
		for k := 0; k < 4; k++ {
			q := Polar(c, 1.8, float64(k)*math.Pi/2+math.Pi/4)
			g.Use(DarkStone).Fill(Capsule{q.At(y0), q.At(yc - 1), 0.45})
		}
		g.Use(White).FillVolume(Sphere(c.At(yc), 2.5))
		g.Paint(Stone, Cylinder{c, 3, 0, yc - 0.25, yc + 0.25})
		if i == 1 {
			g.Paint(Red, Cylinder{c, 3, 0, yc + 1.2, yc + 1.7})
		} else {
			g.Paint(Navy, Cylinder{c, 3, 0, yc + 1.2, yc + 1.7})
		}
	}
	// Pipe run from the tanks to the pad, on low supports.
	pts := []Vec3{V(-17, y0+1.2, -32), V(-21, y0+1.2, -32), V(-21, y0+1.2, -29)}
	for i := 0; i+1 < len(pts); i++ {
		g.Use(Stone).Fill(Capsule{pts[i], pts[i+1], 0.45})
	}
	g.Use(Stone).Fill(Capsule{V(-12, y0+1.2, -32), V(-7, y0+1.2, -32), 0.45})
}

// flags is a row of three poles with generic mission flags.
func (s *marsBaseSite) flags() {
	g := s.g
	y0 := marsBaseH
	for i := 0; i < 3; i++ {
		q := P(5.5+3.0*float64(i), 14.5-0.5*float64(i))
		g.Use(DarkStone).Fill(Cylinder{q, 1.1, 0, y0 - 0.5, y0 + 0.6})
		g.Use(White).Fill(Capsule{q.At(y0), q.At(y0 + 10), 0.5})
		g.Use(Gold).Fill(Sphere(q.At(y0+10.4), 0.6))
		f := Block(V(q.X+0.3, y0+6.6, q.Y-0.4), V(q.X+4.6, y0+9.6, q.Y+0.4))
		switch i {
		case 0:
			g.Use(Navy).Fill(f)
			g.Paint(White, Block(V(q.X+0.3, y0+7.8, q.Y-0.5), V(q.X+4.6, y0+8.4, q.Y+0.5)))
		case 1:
			g.Use(Red).Fill(f)
			g.Paint(White, Block(V(q.X+1.8, y0+6.6, q.Y-0.5), V(q.X+2.8, y0+9.6, q.Y+0.5)))
		default:
			g.Use(White).Fill(f)
			g.Paint(Red, Sphere(V(q.X+2.4, y0+8.1, q.Y), 1.0))
		}
	}
}

// rover is a six-wheeled pressurized rover, forward along angle, at a plan point.
func (s *marsBaseSite) rover(at Vec2, angle float64, idx int) {
	g := s.g
	y0 := s.ground(at.X, at.Y) - 0.2
	place := func(sol Solid) Solid { return RotateY(Translate(marsBaseScaled{sol, 1.25}, at.At(y0)), at, angle) }
	body := Func{Box{V(-5.6, 1.6, -2.9), V(5.6, 5.6, 2.9)}, func(p Vec3) bool {
		x, y, z := p.X/5.4, (p.Y-3.5)/1.9, p.Z/2.7
		x2, y2, z2 := x*x, y*y, z*z
		return x2*x2+y2*y2+z2*z2 <= 1
	}}
	g.Use(White).FillVolume(place(body))
	g.Paint(Navy, place(Block(V(-6, 1.4, -3.2), V(6, 2.4, 3.2))))
	g.Paint(Red, place(Block(V(-6, 3.1, -3.2), V(6, 3.6, 3.2))))
	// Windscreen and side windows.
	g.Paint(Charcoal, place(Func{Box{V(3, 3.8, -3), V(6, 5.4, 3)}, func(p Vec3) bool { return p.X > 3.9 && math.Abs(p.Z) < 2.2 }}))
	g.Paint(Charcoal, place(Func{Box{V(-3, 3.9, -3.2), V(3.4, 4.8, 3.2)}, func(p Vec3) bool {
		return math.Mod(p.X+10, 2.2) < 1.3 && math.Abs(p.Z) > 2.0
	}}))
	g.Paint(Charcoal, place(Block(V(-6, 2.6, -1.0), V(-4.9, 4.6, 1.0))))
	// Wheels.
	for _, wx := range []float64{-3.6, 0, 3.6} {
		for _, sz := range []float64{-1, 1} {
			c := V(wx, 1.4, sz*3.0)
			g.Use(Charcoal).Fill(place(Func{Box{c.Sub(V(1.5, 1.5, 0.6)), c.Add(V(1.5, 1.5, 0.6))}, func(p Vec3) bool {
				return math.Hypot(p.X-c.X, p.Y-c.Y) <= 1.45
			}}))
			g.Paint(Stone, place(Func{Box{c.Sub(V(0.6, 0.6, 0.7)), c.Add(V(0.6, 0.6, 0.7))}, func(p Vec3) bool {
				return math.Hypot(p.X-c.X, p.Y-c.Y) <= 0.55 && math.Abs(p.Z) > 3.2
			}}))
		}
	}
	// Roof gear: solar deck, mast, light bar.
	g.Use(Navy).Fill(place(Block(V(-4.2, 5.4, -1.8), V(0.5, 5.9, 1.8))))
	g.Use(Stone).Fill(place(Capsule{V(-3.5, 5.4, 1.6), V(-3.5, 9, 1.6), 0.45}))
	g.Use(Red).Fill(place(Sphere(V(-3.5, 9.3, 1.6), 0.6)))
	g.Use(Gold).Fill(place(Block(V(2.0, 5.3, -1.6), V(2.8, 5.9, 1.6))))
	if idx == 0 {
		g.Use(White).Fill(place(s.dish(V(-1.5, 6.8, -1.0), V(0.2, 1, 0.3), 1.5, 0.5, 0.8)))
	}
}

// hollow removes cells buried deeper than keep layers.
func (s *marsBaseSite) hollow(keep int) {
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
