package scenes

import (
	"math"

	"godocta/gen/kit"
	. "godocta/gen/octa"
)

// Dubrovnik is the walled old town on the Adriatic: pale limestone walls ringing a
// dense grid of terracotta-roofed stone houses, the straight polished Stradun running
// from the Pile Gate in the west to Luža square and the old harbor in the east. The
// town climbs north to the round Minčeta Tower; Fort Bokar and the Fort of St John
// guard the sea corners, Revelin the Ploče Gate, and Fort Lovrijenac stands on its own
// sea cliff across the cove. Inside: the Franciscan and Dominican bell towers, the
// Sponza Palace, the clock tower, St Blaise, the Rector's Palace, the domed cathedral
// and the Jesuit church above Gundulić square. The harbor holds a breakwater, small
// boats and a carrack replica. North is -Z, east +X: the sea lies south and east.
type Dubrovnik struct{}

func init() { Register(Dubrovnik{}) }

func (Dubrovnik) Name() string  { return "dubrovnik" }
func (Dubrovnik) Title() string { return "Dubrovnik" }

const (
	dubrovnikScale  = 0.9  // The plan is authored large, then shrunk onto the lattice.
	dubrovnikStreet = 2.5  // Stradun paving.
	dubrovnikRock   = 12.5 // Lovrijenac rock top.
)

var (
	// dubrovnikWalls is the wall's center line, clockwise from the Pile Gate.
	dubrovnikWalls = []Vec2{
		P(-46, -2), P(-47, 8), P(-45, 15), P(-32, 21), P(-16, 24), P(0, 25), P(14, 24), P(26, 21), P(37, 16),
		P(38, 8), P(37, -2), P(38, -10), P(30, -18), P(16, -24), P(0, -28), P(-18, -32), P(-30, -26), P(-40, -16),
	}
	dubrovnikLovrijenac = P(-64, 6)
	dubrovnikMinceta    = P(-18, -33)
)

// dubrovnikHit is the nearest point of the wall line to a plan point.
type dubrovnikHit struct {
	sd  float64 // Signed distance: positive outside the town.
	at  Vec2    // Nearest point on the wall's center line.
	arc float64 // Distance along the wall from the Pile Gate.
	out Vec2    // Outward normal of the nearest segment.
}

type dubrovnikRect struct{ x0, z0, x1, z1 float64 }

// dubrovnikLot is one house plot. Faces: 0 -X, 1 +X, 2 -Z, 3 +Z.
type dubrovnikLot struct {
	x0, x1, z0, z1 float64
	seed           int
	windows        [4]bool
	shops          int // Face with shop arches onto the Stradun, or -1.
}

// dubrovnikSite holds the grid, the sampled heightfield and the reserved plots.
type dubrovnikSite struct {
	g            *Grid
	normals      []Vec2
	x0, z0, step float64
	nx, nz       int
	h            []float64
	reserved     []dubrovnikRect
}

// Build authors the plan on a finer temporary lattice and merges its cells, which
// scales the whole site by dubrovnikScale relative to the cell size.
func (Dubrovnik) Build(out *Grid) Order {
	g := NewGridWithSpan(out.Span() / dubrovnikScale)
	defer out.Merge(g)
	s := &dubrovnikSite{g: g, x0: -82, z0: -58, step: 0.4}
	n := len(dubrovnikWalls)
	for i := 0; i < n; i++ {
		a, b := dubrovnikWalls[i], dubrovnikWalls[(i+1)%n]
		ab := b.Sub(a)
		nv := P(ab.Y, -ab.X).Scale(1 / ab.Length())
		if InsidePolygon(dubrovnikWalls, a.Add(ab.Scale(0.5)).Add(nv.Scale(0.5))) {
			nv = nv.Scale(-1)
		}
		s.normals = append(s.normals, nv)
	}
	s.nx, s.nz = int(146/s.step)+2, int(94/s.step)+2
	s.h = make([]float64, s.nx*s.nz)
	for j := 0; j < s.nz; j++ {
		for i := 0; i < s.nx; i++ {
			s.h[j*s.nx+i] = s.ground(s.x0+float64(i)*s.step, s.z0+float64(j)*s.step)
		}
	}
	s.reserved = []dubrovnikRect{
		{-47, -13.5, -30.5, -1.8}, // Franciscan monastery.
		{-47, -1.8, -37, 6.5},     // Pile square and Onofrio's fountain.
		{21.5, -3.5, 40, 5.5},     // Luža square.
		{22, -10.5, 32.5, -3.5},   // Sponza Palace and the clock tower.
		{22, 2.5, 29, 10},         // St Blaise.
		{29.5, 5, 38, 12.5},       // Rector's Palace.
		{18, 10.5, 33, 22},        // Cathedral.
		{-1.5, 2.5, 7.6, 23},      // Gundulić square, Jesuit stairs and church.
		{16, -17.5, 31, -11},      // Dominican monastery.
		{32.5, -12, 40, 12},       // Arsenal.
	}

	s.terrain()
	s.walls()
	s.forts()
	lots := s.lots()
	lots = append(lots, s.infill(lots)...)
	for _, l := range lots {
		s.house(l)
	}
	s.landmarks()
	s.lovrijenac()
	s.harbor()
	s.trees()
	return Sweep(P(-4, -4).Scale(dubrovnikScale), math.Pi)
}

// ---- Math helpers ----

func (*dubrovnikSite) clamp01(v float64) float64 { return math.Max(0, math.Min(1, v)) }
func (*dubrovnikSite) smooth(t float64) float64  { return t * t * (3 - 2*t) }

func (*dubrovnikSite) noise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

func (s *dubrovnikSite) fbm(x, z float64, seed int) float64 {
	return (s.noise(x, z, seed)*4 + s.noise(x*2.1, z*2.1, seed+1)*2 + s.noise(x*4.3, z*4.3, seed+2)) / 7
}

// hash is a per-cell random value for speckled paint.
func (*dubrovnikSite) hash(p Vec3, seed int) float64 {
	return Hash01(int(math.Round(p.X*4)), int(math.Round(p.Y*4)), int(math.Round(p.Z*4)), seed)
}

// ---- Site plan ----

// wall finds the nearest point of the town wall line.
func (s *dubrovnikSite) wall(q Vec2) dubrovnikHit {
	best := dubrovnikHit{sd: math.Inf(1)}
	arc := 0.0
	n := len(dubrovnikWalls)
	for i := 0; i < n; i++ {
		a, b := dubrovnikWalls[i], dubrovnikWalls[(i+1)%n]
		ab := b.Sub(a)
		l := ab.Length()
		t := s.clamp01(q.Sub(a).Dot(ab) / (l * l))
		at := a.Add(ab.Scale(t))
		if d := q.Sub(at).Length(); d < best.sd {
			best = dubrovnikHit{d, at, arc + t*l, s.normals[i]}
		}
		arc += l
	}
	if InsidePolygon(dubrovnikWalls, q) {
		best.sd = -best.sd
	}
	return best
}

// land is 1 on the landward (north and west) side of the walls, 0 towards the sea.
func (s *dubrovnikSite) land(x, z float64) float64 {
	line := 0.0
	switch {
	case x < -44:
		line = 3
	case x < -38:
		line = 3 - 3*(x+44)/6
	case x > 40:
		line = -6
	case x > 34:
		line = -(x - 34)
	}
	return s.smooth(s.clamp01((line-z)/4 + 0.5))
}

// rim is the distance inside the site's rectangular edge.
func (*dubrovnikSite) rim(x, z float64) float64 {
	return math.Min(math.Min(x+75, 58-x), math.Min(z+43, 32-z))
}

// town is the ground inside the walls: the Stradun valley, the steep north slope up
// to Minčeta, the gentler south slope up to the Jesuit church, the low harbor end.
func (s *dubrovnikSite) town(x, z float64) float64 {
	h := dubrovnikStreet
	switch {
	case z < -1.8:
		t := s.clamp01((-z - 1.8) / 26)
		k := 0.7 + 0.3*s.clamp01((14-x)/32)
		h += 10.5 * k * s.smooth(t)
	case z > 1.8:
		t := s.clamp01((z - 1.8) / 12)
		h += 3*s.smooth(t) + 2*math.Exp(-((x-3)*(x-3)/80+(z-17)*(z-17)/30))
	}
	east := s.smooth(s.clamp01((x-21)/7)) * (1 - s.smooth(s.clamp01((-6-z)/6)))
	return h + (dubrovnikStreet-h)*east
}

// ground is the whole site's terrain; values below 0.3 are sea.
func (s *dubrovnikSite) ground(x, z float64) float64 {
	w := s.wall(P(x, z))
	if w.sd < 0 {
		return s.town(x, z)
	}
	d := w.sd
	in := s.town(w.at.X, w.at.Y)
	n := s.fbm(x/6, z/6, 3) - 0.5
	// Landward: a dry ditch below the walls, then slopes easing down to the rim.
	foot := in - 1
	land := 1.4 + (foot-1.4)*(1-s.smooth(s.clamp01((d-2)/20))) + n*1.6*s.clamp01(d/6)
	land -= math.Min(2, land-0.8) * math.Exp(-(d-4.2)*(d-4.2)/3)
	land = 0.6 + (land-0.6)*s.smooth(s.clamp01(s.rim(x, z)/12))
	// Seaward: limestone rocks dropping into the sea.
	sea := math.Max(1.8, in-2.5) - 1.1*(d-1.4) + n*3.5
	lw := s.land(x, z)
	h := land*lw + sea*(1-lw)
	// Lovrijenac's sea cliff.
	q := P(x, z).Sub(dubrovnikLovrijenac)
	a := q.Angle()
	r := q.Length() / (9.8 + 2.5*(s.fbm(3*math.Cos(a)+20, 3*math.Sin(a)+20, 9)-0.5))
	if r < 1 {
		u := s.clamp01((r - 0.62) / 0.38)
		rock := dubrovnikRock*math.Pow(1-u*u, 0.6) + (s.fbm(x/3, z/3, 10)-0.5)*2.5*u
		h = math.Max(h, rock)
	}
	return h
}

// at samples the heightfield bilinearly.
func (s *dubrovnikSite) at(x, z float64) float64 {
	fx, fz := (x-s.x0)/s.step, (z-s.z0)/s.step
	if fx < 0 || fz < 0 || fx >= float64(s.nx-1) || fz >= float64(s.nz-1) {
		return -1
	}
	i, j := int(fx), int(fz)
	tx, tz := fx-float64(i), fz-float64(j)
	k := j*s.nx + i
	a := s.h[k] + (s.h[k+1]-s.h[k])*tx
	b := s.h[k+s.nx] + (s.h[k+s.nx+1]-s.h[k+s.nx])*tx
	return a + (b-a)*tz
}

// low is the lowest ground around a point, used to close steep faces into shells.
func (s *dubrovnikSite) low(x, z float64) float64 {
	lo := s.at(x, z)
	for k := 0; k < 8; k++ {
		a := float64(k) * math.Pi / 4
		lo = math.Min(lo, s.at(x+0.75*math.Cos(a), z+0.75*math.Sin(a)))
	}
	return lo
}

// span is the lowest and highest town ground under a rectangle.
func (s *dubrovnikSite) span(x0, x1, z0, z1 float64) (float64, float64) {
	lo, hi := math.Inf(1), math.Inf(-1)
	for i := 0; i <= 4; i++ {
		for j := 0; j <= 4; j++ {
			h := s.town(x0+(x1-x0)*float64(i)/4, z0+(z1-z0)*float64(j)/4)
			lo, hi = math.Min(lo, h), math.Max(hi, h)
		}
	}
	return lo, hi
}

// footing is the lowest map ground under a disc, never below the sea plate.
func (s *dubrovnikSite) footing(c Vec2, r float64) float64 {
	lo := s.at(c.X, c.Y)
	for k := 0; k < 12; k++ {
		a := float64(k) * math.Pi / 6
		for _, f := range []float64{0.5, 1} {
			lo = math.Min(lo, s.at(c.X+f*r*math.Cos(a), c.Y+f*r*math.Sin(a)))
		}
	}
	return math.Max(-0.2, lo-0.8)
}

func (*dubrovnikSite) inRect(x, z float64, r dubrovnikRect) bool {
	return x >= r.x0 && x <= r.x1 && z >= r.z0 && z <= r.z1
}

// ---- Terrain ----

func (s *dubrovnikSite) terrain() {
	g := s.g
	box := Box{V(-75, -0.3, -43), V(58, 16, 32)}
	g.Use(DarkStone).Fill(Func{box, func(p Vec3) bool {
		if s.rim(p.X, p.Z) < 0 {
			return false
		}
		h := s.at(p.X, p.Z)
		if h < 0.3 {
			return false
		}
		return p.Y <= h && p.Y >= math.Max(-0.2, math.Min(h, s.low(p.X, p.Z))-0.5)
	}})
	g.Use(Water).Fill(Func{Box{V(-75, -0.2, -43), V(58, 0.2, 32)}, func(p Vec3) bool {
		return s.rim(p.X, p.Z) >= 0 && s.at(p.X, p.Z) < 0.3
	}})
	for _, c := range []Color{Stone, Grass, Foliage, Sand, White} {
		c := c
		g.Paint(c, Func{box, func(p Vec3) bool { return p.Y > 0.25 && s.groundColor(p) == c }})
	}
}

// groundColor picks paving, gardens, meadow, scrub and rock.
func (s *dubrovnikSite) groundColor(p Vec3) Color {
	x, z := p.X, p.Z
	h := s.at(x, z)
	w := s.wall(P(x, z))
	if w.sd < 0 {
		switch {
		case math.Abs(z) < 1.8 && x < 22:
			return White
		case s.inRect(x, z, dubrovnikRect{21.5, -3.5, 40, 5.5}), s.inRect(x, z, dubrovnikRect{-46, -1.8, -37, 6.5}):
			return White
		case s.inRect(x, z, dubrovnikRect{0, 9, 5, 13.5}) && math.Mod(p.Y*2.5, 2) < 1:
			return White
		case w.sd > -7 && s.fbm(x/4, z/4, 14) > 0.5:
			if s.fbm(x/2, z/2, 15) > 0.62 {
				return Foliage
			}
			return Grass
		}
		return Stone
	}
	steep := h - s.low(x, z)
	if P(x, z).Sub(dubrovnikLovrijenac).Length() < 12 && h > 1.5 {
		if h > dubrovnikRock-0.6 && steep < 0.8 {
			return Stone
		}
		if Hash01(int(math.Floor(p.Y/1.2+3*s.noise(x/5, z/5, 16))), int(math.Floor(x/3)), int(math.Floor(z/3)), 17) < 0.3 {
			return Stone
		}
		return DarkStone
	}
	if s.land(x, z) > 0.5 && h > 0.7 {
		if steep > 1.5 {
			if s.hash(p, 18) < 0.3 {
				return Stone
			}
			return DarkStone
		}
		n := s.fbm(x/5, z/5, 7)
		switch {
		case n > 0.6:
			return Foliage
		case n > 0.5:
			return Grass
		case s.fbm(x/2.5, z/2.5, 22) > 0.5:
			return Stone
		}
		return Sand
	}
	if Hash01(int(math.Floor(p.Y/1.1+2*s.noise(x/4, z/4, 19))), int(math.Floor(x/2.5)), int(math.Floor(z/2.5)), 20) < 0.18 {
		return Stone
	}
	return DarkStone
}

// ---- Walls ----

// wallTop is the walkway height: higher on the landward side.
func (s *dubrovnikSite) wallTop(w dubrovnikHit) float64 {
	return s.town(w.at.X, w.at.Y) + 5 + 1.5*s.land(w.at.X, w.at.Y)
}

func (s *dubrovnikSite) wallHalf(w dubrovnikHit) float64 {
	return 0.9 + 0.3*s.land(w.at.X, w.at.Y)
}

// wallFoot is where the wall meets the rock or ditch outside it.
func (s *dubrovnikSite) wallFoot(w dubrovnikHit) float64 {
	o := w.at.Add(w.out.Scale(s.wallHalf(w) + 1.5))
	return math.Max(-0.2, math.Min(s.town(w.at.X, w.at.Y)-1, s.at(o.X, o.Y))-0.8)
}

func (s *dubrovnikSite) walls() {
	g := s.g
	box := Box{V(-52, -0.3, -38), V(44, 26, 30)}
	in := func(p Vec3) (bool, bool) {
		w := s.wall(p.Plan())
		if math.Abs(w.sd) > 3.5 {
			return false, false
		}
		half, top := s.wallHalf(w), s.wallTop(w)
		if p.Y > top+2.2 || p.Y < s.wallFoot(w) {
			return false, false
		}
		batter := math.Max(0, top-5-p.Y) * 0.15
		if w.sd < -half || w.sd > half+batter {
			return false, false
		}
		if p.Y <= top {
			return true, false
		}
		if w.sd < half-0.9 {
			return false, false
		}
		if p.Y <= top+1.2 {
			return true, false
		}
		return math.Mod(w.arc, 2.4) < 1.2, p.Y > top+1.7
	}
	g.Use(Stone).Fill(Func{box, func(p Vec3) bool { ok, _ := in(p); return ok }})
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		w := s.wall(p.Plan())
		if math.Abs(w.sd) > 3.5 {
			return false
		}
		return Hash01(int(math.Floor(w.arc/2.6)), int(math.Floor(p.Y/1.6)), 21) < 0.3
	}})
	g.Paint(White, Func{box, func(p Vec3) bool { _, cap := in(p); return cap }})
	// Rectangular towers on the landward walls, rounded bastions towards the sea.
	for _, q := range []Vec2{P(-40, -16), P(-30, -26), P(-8, -30), P(0, -28), P(8, -26), P(16, -24), P(30, -18), P(-46.6, 4.5)} {
		s.wallTower(q, 4.4, 2.4, 2.6)
	}
	for _, q := range []Vec2{P(-32, 21), P(-16, 24), P(0, 25), P(14, 24), P(26, 21)} {
		w := s.wall(q)
		c := w.at.Add(w.out.Scale(1))
		top := s.wallTop(w) + 0.8
		foot := s.footing(c, 3)
		g.Use(Stone).Fill(Cylinder{Center: c, Radius: 3, Y0: foot, Y1: top})
		g.Carve(Cylinder{Center: c, Radius: 1.8, Y0: foot + 1.5, Y1: top - 1})
		g.Build(kit.RingBattlements{Center: c, Radius: 3, Top: top - 0.5, Count: 8})
		s.slits(c, 3, foot+2, top-1, 3, w.out.Angle())
	}
}

// wallTower is a hollow rectangular tower astride the wall, projecting outward.
func (s *dubrovnikSite) wallTower(q Vec2, width, proj, rise float64) {
	g := s.g
	w := s.wall(q)
	half := s.wallHalf(w)
	top := s.wallTop(w) + rise
	out := w.out
	tan := P(-out.Y, out.X)
	at := w.at
	foot := s.footing(at.Add(out.Scale(half+proj/2)), width/2+1)
	frame := func(u0, u1, v0, v1, y0, y1 float64) Solid {
		return Func{Box{V(at.X-9, y0, at.Y-9), V(at.X+9, y1, at.Y+9)}, func(p Vec3) bool {
			d := p.Plan().Sub(at)
			u, v := d.Dot(tan), d.Dot(out)
			return u >= u0 && u <= u1 && v >= v0 && v <= v1 && p.Y >= y0 && p.Y <= y1
		}}
	}
	v0, v1 := -half, half+proj
	g.Use(Stone).Fill(frame(-width/2, width/2, v0, v1, foot, top))
	g.Carve(frame(-width/2+1, width/2-1, v0+1, v1-1, foot+1, top-1))
	g.Use(Stone).Fill(Func{Box{V(at.X-9, top, at.Y-9), V(at.X+9, top+2.2, at.Y+9)}, func(p Vec3) bool {
		d := p.Plan().Sub(at)
		u, v := d.Dot(tan), d.Dot(out)
		if math.Abs(u) > width/2 || v < v0 || v > v1 {
			return false
		}
		edge := math.Abs(u) > width/2-0.9 || v < v0+0.9 || v > v1-0.9
		if !edge {
			return false
		}
		if p.Y <= top+1.2 {
			return true
		}
		return math.Mod(u+v+40, 2.2) < 1.1
	}})
	// Slits facing out.
	for y := foot + 2.5; y < top-1.5; y += 3 {
		g.Carve(frame(-0.45, 0.45, v1-1.2, v1+0.3, y, y+1.3))
	}
}

// slits carves arrow slits around a round tower and darkens the openings.
func (s *dubrovnikSite) slits(c Vec2, r, y0, y1 float64, n int, start float64) {
	level := 0
	for y := y0; y+1.6 < y1; y += 3 {
		for k := 0; k < n; k++ {
			a := start + (float64(k)-float64(n-1)/2)*0.6 + float64(level%2)*0.3
			slit := Arch{Sill: V(r-0.6, y, 0), HalfWidth: 0.5, Spring: 1.0, Depth: 2.4, Through: AxisX}
			s.g.Carve(Translate(RotateY(slit, P(0, 0), a), c.At(0)))
		}
		level++
	}
}

// ---- Forts and gates ----

func (s *dubrovnikSite) forts() {
	g := s.g
	// Minčeta: a massive battered drum crowned by a slimmer round tower.
	c := dubrovnikMinceta
	foot := s.footing(c, 7)
	low := s.town(c.X, c.Y+3) + 7.5
	g.Use(Stone).Fill(Frustum{c, 7.2, 6.2, foot, low})
	g.Carve(Frustum{c, 6, 5, foot + 1, low - 1})
	g.Build(kit.RingBattlements{Center: c, Radius: 6.2, Top: low, Count: 16})
	g.Paint(Sand, Func{Box{V(c.X-8, foot, c.Y-8), V(c.X+8, low, c.Y+8)}, func(p Vec3) bool {
		return Hash01(int(math.Floor(p.Y/1.4)), int(math.Floor(p.Plan().Sub(c).Angle()*4)), 22) < 0.3
	}})
	upper := low + 8
	g.Use(Stone).Fill(Cylinder{Center: c, Radius: 4.3, Y0: low, Y1: upper})
	g.Carve(Cylinder{Center: c, Radius: 3, Y0: low + 1, Y1: upper - 2})
	g.Fill(Frustum{c, 4.3, 5.1, upper - 1.6, upper - 0.6}, Cylinder{Center: c, Radius: 5.1, Y0: upper - 0.6, Y1: upper})
	g.Build(kit.RingBattlements{Center: c, Radius: 5.1, Top: upper, Count: 12})
	g.Paint(White, Cylinder{Center: c, Radius: 5.2, Y0: upper + 0.1, Y1: upper + 3})
	// Lookout turret on the parapet.
	t := c.Add(P(-2.6, -2.2))
	g.Fill(Cylinder{Center: t, Radius: 1.4, Y0: upper, Y1: upper + 3.6})
	g.With(White, func() { g.Fill(Cylinder{Center: t, Radius: 1.7, Y0: upper + 3.6, Y1: upper + 4.2}) })
	s.slits(c, 7, foot+2.5, low-1, 5, -math.Pi/2)
	s.slits(c, 4.3, low+1.5, upper-2, 6, -math.Pi/2-0.3)
	s.darkRound(c, 5, foot+1, low-1)
	s.darkRound(c, 3, low+1, upper-2)

	// Bokar: a squat casemate drum at the south-west sea corner.
	c = P(-46.5, 16)
	foot = s.footing(c, 4.8)
	top := s.wallTop(s.wall(c)) + 0.5
	g.Use(Stone).Fill(Frustum{c, 5.2, 4.6, foot, top})
	g.Carve(Frustum{c, 4, 3.4, foot + 1, top - 1})
	g.Build(kit.RingBattlements{Center: c, Radius: 4.6, Top: top - 0.5, Count: 12})
	s.slits(c, 5, foot+2, top-1.5, 5, math.Pi*0.75)
	s.darkRound(c, 3.4, foot+1, top-1)

	// Fort of St John at the harbor mouth: a long casemate and a round bastion.
	g.Use(Stone).Fill(Block(V(31, -0.2, 13), V(40, 9, 21)))
	g.Carve(Block(V(32.5, 0.8, 14.5), V(38.5, 7.6, 19.5)))
	g.Build(kit.Battlements{Min: P(31, 13), Max: P(40, 21), Top: 8.5, Rim: 1, Period: 2.2})
	c = P(41, 18)
	g.Fill(Frustum{c, 5.4, 5, -0.2, 9.8})
	g.Carve(Cylinder{Center: c, Radius: 4.1, Y0: 0.8, Y1: 8.8})
	g.Fill(Block(V(31, -0.2, 19.5), V(40, 8.5, 21)), Block(V(31, -0.2, 13), V(40, 8.5, 14.5)))
	g.Build(kit.RingBattlements{Center: c, Radius: 5, Top: 9.3, Count: 14})
	s.slits(c, 5.2, 2, 8.5, 5, math.Pi/4)
	for _, x := range []float64{33, 35.5, 38} {
		for _, y := range []float64{2.5, 5.5} {
			g.Carve(Arch{Sill: V(x, y, 21), HalfWidth: 0.55, Spring: 0.8, Depth: 2.4, Through: AxisZ})
		}
	}
	s.darkInside(Box{V(30.5, -0.5, 12.5), V(46.5, 10, 23.5)}, Union(Block(V(32.5, 0.8, 14.5), V(38.5, 7.6, 19.5)),
		Subtract(Cylinder{Center: c, Radius: 4.1, Y0: 0.8, Y1: 8.8}, Block(V(30, -1, 19.4), V(40.2, 10, 24)), Block(V(30, -1, 10), V(40.2, 10, 14.6)))))

	// Pile Gate: gate tower, outer arch and a stone bridge over the ditch.
	gy := dubrovnikStreet
	g.Use(Stone).Fill(Block(V(-50, 0.5, -5.5), V(-43.5, 11, 1.5)))
	g.Build(kit.Battlements{Min: P(-50, -5.5), Max: P(-43.5, 1.5), Top: 11, Rim: 1, Period: 2.2})
	g.Carve(Arch{Sill: V(-46.5, gy, -2), HalfWidth: 1.5, Spring: 2.2, Depth: 9, Through: AxisX})
	g.With(Charcoal, func() { g.Fill(Block(V(-48, 8.5, -6), V(-45, 9.6, -5.4))) })
	g.With(White, func() { g.Fill(Block(V(-50.4, 6.6, -2.6), V(-49.9, 8.4, -1.4))) }) // St Blaise in his niche.
	g.Use(Stone).Fill(Block(V(-57, 0.2, -3.8), V(-50, gy, -0.2)))
	for _, x := range []float64{-55, -52} {
		g.Carve(Arch{Sill: V(x, 0, -2), HalfWidth: 1.1, Spring: 0.4, Depth: 5, Through: AxisZ})
	}
	g.Fill(Block(V(-57, gy, -3.8), V(-50, gy+0.9, -3.1)), Block(V(-57, gy, -0.9), V(-50, gy+0.9, -0.2)))
	g.With(Wood, func() { g.Fill(Block(V(-51.5, gy-0.4, -3.1), V(-50, gy, -0.9))) })

	// Ploče Gate and its bridge to Revelin.
	py := s.town(36, -10)
	g.Use(Stone).Fill(Block(V(35.5, py-1, -13), V(41, py+8.5, -7)))
	g.Build(kit.Battlements{Min: P(35.5, -13), Max: P(41, -7), Top: py + 8.5, Rim: 1, Period: 2.2})
	g.Carve(Arch{Sill: V(38.2, py, -10), HalfWidth: 1.4, Spring: 2, Depth: 8, Through: AxisX})
	g.Fill(Block(V(41, 0.5, -11.6), V(46, py, -8.4)))
	g.Carve(Arch{Sill: V(43.5, 0.4, -10), HalfWidth: 1.2, Spring: 0.4, Depth: 5, Through: AxisZ})
	g.Fill(Block(V(41, py, -11.6), V(46, py+0.9, -10.9)), Block(V(41, py, -9.1), V(46, py+0.9, -8.4)))

	// Revelin: a squat polygonal fort beyond the Ploče Gate, battered walls
	// around a terrace with its barracks.
	poly := []Vec2{P(45, -21), P(56, -22), P(57.5, -9), P(46, -7.5)}
	inner := []Vec2{P(46.5, -19.5), P(54.8, -20.3), P(56, -10.3), P(47.4, -9.1)}
	rf := s.footing(P(51, -14.5), 7)
	rt, deck := 8.5, 6.5
	cen := P(51.2, -14.6)
	batter := func(p Vec3, grow float64) bool {
		q := cen.Add(p.Plan().Sub(cen).Scale(1 / (1 + grow*math.Max(0, deck-p.Y)/40)))
		return InsidePolygon(poly, q)
	}
	g.Use(Stone).Fill(Func{Box{V(42, rf, -25), V(61, rt, -4)}, func(p Vec3) bool { return batter(p, 1) }})
	g.Carve(Prism{inner, rf + 1, rt + 3})
	g.With(Sand, func() { g.Fill(Prism{inner, deck - 0.8, deck}) })
	g.Fill(Func{Box{V(44, rt, -23), V(58, rt+2.2, -6)}, func(p Vec3) bool {
		if !InsidePolygon(poly, p.Plan()) || InsidePolygon(inner, p.Plan()) {
			return false
		}
		return p.Y <= rt+1.1 || math.Mod(p.X+p.Z+40, 2.4) < 1.2
	}})
	g.Paint(Sand, Func{Box{V(42, rf, -25), V(61, rt+3, -4)}, func(p Vec3) bool {
		return Hash01(int(math.Floor((p.X+p.Z)/2.6)), int(math.Floor(p.Y/1.6)), 23) < 0.3
	}})
	g.Paint(DarkStone, Func{Box{V(42, rf, -25), V(61, rf+1.6, -4)}, func(p Vec3) bool { return true }})
	s.hall(48.5, 54, -18.2, -12.4, deck, deck+4.5, Stone, 1, Terracotta, 2.6)
	for _, z := range []float64{-18, -14.5, -11} {
		g.Carve(Arch{Sill: V(57.5, 2.5, z), HalfWidth: 0.55, Spring: 0.8, Depth: 4, Through: AxisX})
	}
	for _, x := range []float64{48, 51.5} {
		g.Carve(Arch{Sill: V(x, 2.5, -21.8), HalfWidth: 0.55, Spring: 0.8, Depth: 4, Through: AxisZ})
	}
	g.Carve(Arch{Sill: V(46, py, -10), HalfWidth: 1.1, Spring: 1.4, Depth: 3, Through: AxisX})
	s.darkInside(Box{V(43, rf, -24), V(60, deck, -5)}, Prism{inner, rf + 1, deck - 0.8})
}

// darkRound paints the inner faces of a hollow round tower.
func (s *dubrovnikSite) darkRound(c Vec2, r, y0, y1 float64) {
	s.darkInside(Box{V(c.X-r-2, y0-1, c.Y-r-2), V(c.X+r+2, y1+1, c.Y+r+2)}, Cylinder{Center: c, Radius: r, Y0: y0, Y1: y1})
}

// darkInside paints Charcoal every cell that faces an empty cell inside the hollow,
// so windows and slits read as dark openings.
func (s *dubrovnikSite) darkInside(box Box, hollow Solid) {
	g := s.g
	sp := g.Span()
	h := sp / 2
	offs := []Vec3{V(sp, 0, 0), V(-sp, 0, 0), V(0, sp, 0), V(0, -sp, 0), V(0, 0, sp), V(0, 0, -sp)}
	for _, a := range []float64{-h, h} {
		for _, b := range []float64{-h, h} {
			for _, c := range []float64{-h, h} {
				offs = append(offs, V(a, b, c))
			}
		}
	}
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		c := Nearest(p.Scale(1 / sp))
		if g.World(c).Sub(p).Length() > 1e-6 {
			return false
		}
		for _, o := range offs {
			q := p.Add(o)
			if hollow.Contains(q) && !g.HasAt(q) {
				return true
			}
		}
		return false
	}})
}

// ---- Houses ----

// lots lays out the house plots: long blocks between parallel lanes running up from
// the Stradun, broken by a cross street on each side.
func (s *dubrovnikSite) lots() []dubrovnikLot {
	var lots []dubrovnikLot
	seed := 0
	valid := func(x0, x1, z0, z1 float64) bool {
		xm, zm := (x0+x1)/2, (z0+z1)/2
		for _, q := range []Vec2{P(x0, z0), P(x1, z0), P(x0, z1), P(x1, z1), P(xm, z0), P(xm, z1), P(x0, zm), P(x1, zm)} {
			if s.wall(q).sd > -2.3 {
				return false
			}
		}
		return true
	}
	blocked := func(x0, x1, z0, z1 float64) bool {
		for _, r := range s.reserved {
			if x1 > r.x0+0.1 && x0 < r.x1-0.1 && z1 > r.z0+0.1 && z0 < r.z1-0.1 {
				return true
			}
		}
		return false
	}
	// column fills one block outward from the Stradun; dir is -1 north, +1 south.
	column := func(x0, x1, dir, crossA, crossB float64) {
		start := len(lots)
		t := 0.0
		for t < 34 {
			seed++
			d := 4 + 0.5*math.Floor(Hash01(seed, 7)*3)
			if dir > 0 {
				d += 0.5
			}
			t0, t1 := t, t+d
			if t0 < crossA && t1 > crossA {
				t1 = crossA
				if t1-t0 < 3 {
					t = crossB
					continue
				}
			}
			toZ := func(v float64) float64 { return dir * (1.8 + v) }
			z0, z1 := math.Min(toZ(t0), toZ(t1)), math.Max(toZ(t0), toZ(t1))
			if blocked(x0, x1, z0, z1) {
				t = t1
				if t1 == crossA {
					t = crossB
				}
				continue
			}
			last, ok := false, false
			lx0, lx1 := x0, x1
			for _, v := range [][2]float64{{0, 0}, {1.5, 0}, {0, -1.5}, {1.5, -1.5}} {
				lx0, lx1 = x0+v[0], x1+v[1]
				for e := t1; e-t0 >= 3; e -= 0.5 {
					z0, z1 = math.Min(toZ(t0), toZ(e)), math.Max(toZ(t0), toZ(e))
					if valid(lx0, lx1, z0, z1) {
						ok, last = true, e < t1
						t1 = e
						break
					}
				}
				if ok {
					break
				}
			}
			if !ok {
				break
			}
			l := dubrovnikLot{x0: lx0, x1: lx1, z0: z0, z1: z1, seed: seed, shops: -1}
			l.windows[0], l.windows[1] = true, true
			near, far := 3, 2 // Faces towards and away from the Stradun.
			if dir > 0 {
				near, far = 2, 3
			}
			if t0 == 0 || t0 == crossB {
				l.windows[near] = true
			}
			if t0 == 0 {
				l.shops = near
			}
			if t1 == crossA || last {
				l.windows[far] = true
			}
			if xm := math.Round(lx0+lx1) / 2; lx1-lx0 >= 5.5 && Hash01(seed, 11) < 0.45 {
				a, b := l, l
				a.x1, b.x0 = xm, xm
				a.windows[1], b.windows[0] = false, false
				b.seed = seed + 5000
				lots = append(lots, a, b)
			} else {
				lots = append(lots, l)
			}
			t = t1
			if t1 == crossA {
				t = crossB
			}
			if last {
				break
			}
		}
		if n := len(lots); n > start {
			lots[n-1].windows[map[bool]int{true: 2, false: 3}[dir < 0]] = true
		}
	}
	for lx := -46.0; lx < 36; lx += 7.5 {
		column(lx+1.5, lx+7.5, -1, 8.2, 9.7)
		column(lx+1.5, lx+7.5, 1, 6.7, 8.2)
	}
	return lots
}

// infill packs small houses into the leftover ground along the inner walls,
// keeping the Stradun, the cross streets and the column lanes clear.
func (s *dubrovnikSite) infill(lots []dubrovnikLot) []dubrovnikLot {
	g := s.g
	var out []dubrovnikLot
	taken := func(x0, x1, z0, z1, gap float64) bool {
		for _, ls := range [][]dubrovnikLot{lots, out} {
			for _, l := range ls {
				if x1 > l.x0-gap && x0 < l.x1+gap && z1 > l.z0-gap && z0 < l.z1+gap {
					return true
				}
			}
		}
		for _, r := range s.reserved {
			if x1 > r.x0-0.5 && x0 < r.x1+0.5 && z1 > r.z0-0.5 && z0 < r.z1+0.5 {
				return true
			}
		}
		return false
	}
	street := func(a, b float64, bands [][2]float64) bool {
		for _, band := range bands {
			if b > band[0] && a < band[1] {
				return true
			}
		}
		return false
	}
	zBands := [][2]float64{{-2.3, 2.3}, {-11.7, -9.8}, {8.3, 10.2}}
	var xBands [][2]float64
	for lx := -46.0; lx < 36; lx += 7.5 {
		xBands = append(xBands, [2]float64{lx - 0.1, lx + 1.6})
	}
	seed := 9000
	for _, size := range [][2]float64{{5, 4.5}, {4.5, 3.5}, {3.5, 4.5}, {3.5, 3}, {3, 2.5}, {2.5, 3}} {
		for z := -34.0; z < 26; z += 0.5 {
			for x := -48.0; x < 40; x += 0.5 {
				x0, x1, z0, z1 := x, x+size[0], z, z+size[1]
				if street(z0, z1, zBands) || street(x0, x1, xBands) || taken(x0, x1, z0, z1, 1.1) {
					continue
				}
				ok := true
				for u := 0.0; u <= 1 && ok; u += 0.25 {
					for v := 0.0; v <= 1 && ok; v += 0.25 {
						q := P(x0+(x1-x0)*u, z0+(z1-z0)*v)
						if s.wall(q).sd > -1.4 {
							ok = false
							break
						}
						y := s.town(q.X, q.Y)
						if g.HasAt(q.At(y+1.5)) || g.HasAt(q.At(y+3.5)) {
							ok = false
						}
					}
				}
				if !ok {
					continue
				}
				seed++
				l := dubrovnikLot{x0: x0, x1: x1, z0: z0, z1: z1, seed: seed, shops: -1}
				l.windows = [4]bool{true, true, true, true}
				out = append(out, l)
			}
		}
	}
	return out
}

// facePatch is a block on one face of a rectangle: c is the position along the face,
// d0..d1 the depth measured inward from the face plane.
func (*dubrovnikSite) facePatch(x0, x1, z0, z1 float64, face int, c, halfw, y0, y1, d0, d1 float64) Cuboid {
	switch face {
	case 0:
		return Block(V(x0+d0, y0, c-halfw), V(x0+d1, y1, c+halfw))
	case 1:
		return Block(V(x1-d1, y0, c-halfw), V(x1-d0, y1, c+halfw))
	case 2:
		return Block(V(c-halfw, y0, z0+d0), V(c+halfw, y1, z0+d1))
	}
	return Block(V(c-halfw, y0, z1-d1), V(c+halfw, y1, z1-d0))
}

// faceInfo is the extent along a face and the ground at its middle.
func (s *dubrovnikSite) faceInfo(x0, x1, z0, z1 float64, face int) (float64, float64, float64) {
	switch face {
	case 0:
		return z0, z1, s.town(x0-0.8, (z0+z1)/2)
	case 1:
		return z0, z1, s.town(x1+0.8, (z0+z1)/2)
	case 2:
		return x0, x1, s.town((x0+x1)/2, z0-0.8)
	}
	return x0, x1, s.town((x0+x1)/2, z1+0.8)
}

// house is one tall stone house: shuttered windows, a door, shops on the Stradun,
// and a tiled roof.
func (s *dubrovnikSite) house(l dubrovnikLot) {
	g := s.g
	x0, x1, z0, z1 := l.x0, l.x1, l.z0, l.z1
	lo, hi := s.span(x0, x1, z0, z1)
	storeys := 3
	switch k := Hash01(l.seed, 1); {
	case k < 0.15:
		storeys = 4
	case k > 0.6:
		storeys = 2
	}
	top := math.Round((hi+0.6+2.1*float64(storeys))*2) / 2
	wall := Stone
	switch k := Hash01(l.seed, 2); {
	case k < 0.3:
		wall = Sand
	case k > 0.84:
		wall = White
	}
	g.Use(wall).Fill(Block(V(x0, lo-0.6, z0), V(x1, top, z1)))
	inner := Block(V(x0+0.9, lo+0.1, z0+0.9), V(x1-0.9, top-0.9, z1-0.9))
	g.Carve(inner)

	// Openings, and dark panels inside them so they read as deep windows.
	var holes, backs []Solid
	back := func(face int, a, b, c, halfw, y0, y1 float64) {
		halfw = math.Min(halfw, math.Min(c-a, b-c)-0.95)
		backs = append(backs, s.facePatch(x0, x1, z0, z1, face, c, halfw, y0, y1, 1.5, 2.1))
	}
	open := 0.55 + 0.3*Hash01(l.seed, 6)
	for face := 0; face < 4; face++ {
		if !l.windows[face] {
			continue
		}
		a, b, ground := s.faceInfo(x0, x1, z0, z1, face)
		length := b - a
		n := int((length - 0.4) / 2.6)
		if n < 1 && length >= 2.8 {
			n = 1
		}
		if n < 1 {
			continue
		}
		mid := (a + b) / 2
		door := face == l.shops || (l.shops < 0 && face == 0)
		for k := 0; ; k++ {
			sill := ground + 0.9 + 2.1*float64(k)
			if sill < lo+0.4 {
				continue
			}
			if sill+1.6 > top-0.6 {
				break
			}
			for i := 0; i < n; i++ {
				c := mid + (float64(i)-float64(n-1)/2)*2.6
				if k == 0 && face == l.shops {
					continue
				}
				if k == 0 && door && math.Abs(c-mid) < 1.3 {
					continue
				}
				if Hash01(l.seed, face, k, i, 8) < open {
					holes = append(holes, s.facePatch(x0, x1, z0, z1, face, c, 0.6, sill, sill+1.6, -0.4, 1.4))
					back(face, a, b, c, 1, sill-0.3, sill+1.9)
					g.Paint(Grass, s.facePatch(x0, x1, z0, z1, face, c-0.9, 0.3, sill, sill+1.6, -0.3, 0.45))
					g.Paint(Grass, s.facePatch(x0, x1, z0, z1, face, c+0.9, 0.3, sill, sill+1.6, -0.3, 0.45))
				} else {
					g.Paint(Grass, s.facePatch(x0, x1, z0, z1, face, c, 0.6, sill, sill+1.6, -0.3, 0.45))
				}
			}
		}
		if face == l.shops {
			for i := 0; i < n; i++ {
				c := mid + (float64(i)-float64(n-1)/2)*2.6
				sill := V(c, ground, z0)
				if face == 3 {
					sill = V(c, ground, z1)
				}
				holes = append(holes, Arch{Sill: sill, HalfWidth: 0.75, Spring: 1.3, Depth: 2.6, Through: AxisZ})
				back(face, a, b, c, 1.1, ground-0.3, ground+2.4)
			}
		} else if door {
			g.Paint(Wood, s.facePatch(x0, x1, z0, z1, face, mid, 0.5, ground, ground+1.9, -0.3, 0.45))
		}
	}
	g.Carve(holes...)
	g.With(Charcoal, func() { g.Fill(backs...) })

	// Roof.
	rc, alt := Terracotta, Brick
	switch k := Hash01(l.seed, 3); {
	case k < 0.14:
		rc, alt = Brick, Terracotta
	case k < 0.24:
		rc, alt = Red, Terracotta
	}
	lo2, hi2 := P(x0, z0-0.3), P(x1, z1+0.3)
	var roof Solid
	switch k := Hash01(l.seed, 4); {
	case x1-x0 < 4 && k < 0.8:
		roof = Gable{lo2, hi2, top, AxisZ}
	case k < 0.25:
		roof = HipRoof{lo2, hi2, top}
	case k < 0.4:
		roof = Gable{lo2, hi2, top, AxisZ}
	default:
		roof = Gable{lo2, hi2, top, AxisX}
	}
	g.With(rc, func() { g.Fill(roof) })
	g.Carve(Intersect(Translate(roof, V(0, -1.1, 0)), Block(V(x0, top+0.4, z0), V(x1, top+9, z1))))
	g.Paint(alt, Func{roof.Bounds(), func(p Vec3) bool { return roof.Contains(p) && s.hash(p, l.seed) < 0.12 }})
	if Hash01(l.seed, 5) < 0.35 {
		cx := x0 + 1.2 + (x1-x0-2.4)*Hash01(l.seed, 9)
		cz := z0 + 1.2 + (z1-z0-2.4)*Hash01(l.seed, 10)
		g.With(Stone, func() { g.Fill(BlockAt(P(cx, cz), 0.9, 0.9, top, top+math.Min(x1-x0, z1-z0)/2+0.8)) })
	}
}

// ---- Landmarks ----

// hall is a hollow masonry building with arched windows and a roof.
// roof: 0 flat, 1 hip, 2 gable along X, 3 gable along Z.
func (s *dubrovnikSite) hall(x0, x1, z0, z1, ground, top float64, wall Color, roof int, roofColor Color, spacing float64) {
	g := s.g
	lo, _ := s.span(x0, x1, z0, z1)
	base := math.Min(lo, ground) - 1
	g.Use(wall).Fill(Block(V(x0, base, z0), V(x1, top, z1)))
	hollow := Block(V(x0+1, ground+0.1, z0+1), V(x1-1, top-1, z1-1))
	g.Carve(hollow)
	for face := 0; face < 4; face++ {
		a, b, _ := s.faceInfo(x0, x1, z0, z1, face)
		n := int((b - a - 1) / spacing)
		for i := 0; i < n; i++ {
			c := (a+b)/2 + (float64(i)-float64(n-1)/2)*spacing
			for y := ground + 1.6; y+2.4 < top-0.6; y += 3.4 {
				var sill Vec3
				through := AxisX
				switch face {
				case 0:
					sill = V(x0+0.5, y, c)
				case 1:
					sill = V(x1-0.5, y, c)
				case 2:
					sill, through = V(c, y, z0+0.5), AxisZ
				default:
					sill, through = V(c, y, z1-0.5), AxisZ
				}
				g.Carve(Arch{Sill: sill, HalfWidth: 0.65, Spring: 1.4, Depth: 2, Through: through})
				g.With(Charcoal, func() {
					g.Fill(s.facePatch(x0, x1, z0, z1, face, c, math.Min(0.8, math.Min(c-a, b-c)-1.05), y-0.2, y+2.2, 1.6, 2.2))
				})
			}
		}
	}
	lo2, hi2 := P(x0-0.4, z0-0.4), P(x1+0.4, z1+0.4)
	var r Solid
	switch roof {
	case 1:
		r = HipRoof{lo2, hi2, top}
	case 2:
		r = Gable{lo2, hi2, top, AxisX}
	case 3:
		r = Gable{lo2, hi2, top, AxisZ}
	default:
		g.Build(kit.Battlements{Min: P(x0, z0), Max: P(x1, z1), Top: top - 0.5, Rim: 0.9, Period: 2})
		return
	}
	g.With(roofColor, func() { g.Fill(r) })
	g.Carve(Intersect(Translate(r, V(0, -1.1, 0)), Block(V(x0, top+0.4, z0), V(x1, top+12, z1))))
}

// belfry is a square bell tower with open arches and a small dome.
func (s *dubrovnikSite) belfry(c Vec2, w, ground, top float64, dome Color) {
	g := s.g
	x0, z0, x1, z1 := c.X-w/2, c.Y-w/2, c.X+w/2, c.Y+w/2
	g.Use(Stone).Fill(Block(V(x0, ground-1, z0), V(x1, top, z1)))
	g.Carve(Block(V(x0+0.9, ground+0.5, z0+0.9), V(x1-0.9, top-1, z1-0.9)))
	g.With(White, func() {
		g.Fill(Block(V(x0-0.3, top-3.6, z0-0.3), V(x1+0.3, top-3.2, z1+0.3)), Block(V(x0-0.3, top, z0-0.3), V(x1+0.3, top+0.5, z1+0.3)))
	})
	g.Carve(Arch{Sill: V(c.X, top-3, c.Y), HalfWidth: w/2 - 0.9, Spring: 1.2, Depth: w + 1, Through: AxisX},
		Arch{Sill: V(c.X, top-3, c.Y), HalfWidth: w/2 - 0.9, Spring: 1.2, Depth: w + 1, Through: AxisZ})
	g.With(Gold, func() { g.Fill(BlockAt(c, 0.9, 0.9, top-3, top-1.6)) })
	g.With(dome, func() {
		g.Fill(Ellipsoid{c.At(top + 0.5), V(w/2, w*0.55, w/2)})
		g.Fill(Cylinder{Center: c, Radius: 0.5, Y0: top + 0.5, Y1: top + 0.5 + w*0.55 + 1})
	})
	g.With(Gold, func() { g.Fill(Sphere(c.At(top+w*0.55+1.6), 0.6)) })
}

func (s *dubrovnikSite) landmarks() {
	g := s.g
	gy := dubrovnikStreet

	// Franciscan monastery: church along the Stradun, cloister behind, bell tower.
	s.hall(-44, -34, -7, -2.2, gy, gy+7.5, Stone, 2, Terracotta, 2.6)
	cl0, cl1 := V(-44, gy-1, -13), V(-34, gy+5, -7)
	g.Use(Sand).Fill(Block(cl0, cl1))
	g.Carve(Block(V(-43, gy+0.1, -12), V(-35, gy+4, -8)))
	g.Carve(Block(V(-41, gy+0.1, -11), V(-37, gy+8, -9)))
	g.With(Grass, func() { g.Fill(Block(V(-41, gy-0.5, -11), V(-37, gy, -9))) })
	g.With(Foliage, func() { g.Fill(Ellipsoid{V(-39, gy+1.2, -10), V(0.9, 1.2, 0.9)}) })
	for _, x := range []float64{-40, -38} {
		g.Carve(Arch{Sill: V(x, gy, -11.2), HalfWidth: 0.5, Spring: 1.2, Depth: 1.4, Through: AxisZ},
			Arch{Sill: V(x, gy, -8.8), HalfWidth: 0.5, Spring: 1.2, Depth: 1.4, Through: AxisZ})
	}
	g.With(Terracotta, func() {
		g.Fill(Subtract(HipRoof{P(-44.4, -13.4), P(-33.6, -6.6), gy + 5}, Block(V(-41.6, gy, -11.6), V(-36.4, gy+9, -8.4))))
	})
	s.belfry(P(-32.4, -4), 3.2, gy, gy+14.5, DarkStone)

	// Onofrio's fountain.
	f := P(-41.5, 3)
	g.Use(Stone).Fill(Cylinder{Center: f, Radius: 2.4, Y0: gy - 0.5, Y1: gy + 1.8})
	g.With(White, func() { g.Fill(Ellipsoid{f.At(gy + 1.8), V(2.4, 1.6, 2.4)}) })
	g.With(Water, func() { g.Fill(Cylinder{Center: f, Radius: 3, Inner: 2.4, Y0: gy, Y1: gy + 0.4}) })
	g.With(Stone, func() { g.Fill(Cylinder{Center: f, Radius: 3.5, Inner: 3, Y0: gy, Y1: gy + 0.8}) })

	// Sponza Palace: portico on the square, Gothic first floor, Renaissance attic.
	sx0, sx1, sz0, sz1 := 23.0, 31.0, -10.0, -4.0
	s.hall(sx0, sx1, sz0, sz1, gy, gy+8.5, Sand, 1, Terracotta, 2.2)
	g.Carve(Block(V(sx0+0.6, gy+0.1, sz1-1.6), V(sx1-0.6, gy+3.2, sz1+0.2)))
	g.Use(Sand)
	for k := 0; k < 6; k++ {
		x := sx0 + 0.9 + float64(k)*(sx1-sx0-1.8)/5
		g.Fill(Cylinder{Center: P(x, sz1-0.4), Radius: 0.5, Y0: gy, Y1: gy + 3.2})
	}
	g.With(Charcoal, func() { g.Fill(Block(V(sx0+1, gy, sz1-1.9), V(sx1-1, gy+3.2, sz1-1.6))) })
	g.With(White, func() { g.Fill(Block(V(sx0, gy+8.5, sz1-0.6), V(sx1, gy+9.4, sz1))) })

	// Clock tower at the end of the Stradun.
	ct := P(34.2, -4.2)
	s.belfry(ct, 3.4, gy, gy+15, DarkStone)
	g.With(Navy, func() {
		g.Fill(Func{Box{V(32.2, gy+10, -6), V(32.7, gy+13, -2.4)}, func(p Vec3) bool {
			return math.Hypot(p.Y-(gy+11.4), p.Z-ct.Y) < 1.3
		}})
	})
	g.With(Gold, func() { g.Fill(Block(V(32.2, gy+11.1, ct.Y-0.3), V(32.6, gy+11.7, ct.Y+0.3))) })
	g.With(Grass, func() {
		g.Fill(BlockAt(ct.Add(P(-0.9, 0)), 0.8, 0.8, gy+15.5, gy+16.6), BlockAt(ct.Add(P(0.9, 0)), 0.8, 0.8, gy+15.5, gy+16.6))
	})

	// Orlando's column.
	g.With(White, func() {
		g.Fill(BlockAt(P(23.5, 0), 1.6, 1.6, gy, gy+0.8))
		g.Fill(Cylinder{Center: P(23.5, 0), Radius: 0.55, Y0: gy, Y1: gy + 4.5})
	})
	g.With(Red, func() { g.Fill(BlockAt(P(23.5, 0.6), 0.6, 1, gy+4.5, gy+5.5)) })

	// St Blaise: baroque church facing the square, with a dome over the sanctuary.
	s.hall(22.5, 28.5, 3, 9.5, gy, gy+7, White, 3, Terracotta, 2.4)
	g.With(White, func() {
		g.Fill(Gable{P(22.2, 2.6), P(28.8, 3.6), gy + 7, AxisZ})
		g.Fill(Block(V(23, gy-0.4, 1.9), V(28, gy+0.4, 3)))
	})
	g.Carve(Arch{Sill: V(25.5, gy+0.4, 3.2), HalfWidth: 0.8, Spring: 1.6, Depth: 1.6, Through: AxisZ})
	dc := P(25.5, 7.5)
	g.Use(Stone).Fill(Cylinder{Center: dc, Radius: 1.9, Y0: gy + 7, Y1: gy + 10.4})
	g.With(DarkStone, func() { g.Fill(Ellipsoid{dc.At(gy + 10.4), V(1.9, 1.8, 1.9)}) })
	g.With(Gold, func() { g.Fill(Cylinder{Center: dc, Radius: 0.5, Y0: gy + 12, Y1: gy + 13.2}) })

	// Rector's Palace with its arcaded loggia.
	s.hall(30, 37, 5.5, 12, gy, gy+8.5, Stone, 1, Terracotta, 2.2)
	g.Carve(Block(V(29.8, gy+0.1, 6.2), V(31.6, gy+3.4, 11.3)))
	g.Use(White)
	for k := 0; k < 6; k++ {
		g.Fill(Cylinder{Center: P(30.4, 6.4+float64(k)*0.98), Radius: 0.5, Y0: gy, Y1: gy + 3.4})
	}
	g.With(Charcoal, func() { g.Fill(Block(V(31.6, gy, 6.2), V(31.9, gy+3.4, 11.3))) })

	// Cathedral of the Assumption: cross plan, facade to the north, lead dome.
	clo, chi := s.span(19, 30, 11, 20.5)
	cb := clo
	s.hall(21.6, 27.4, 11, 20.5, cb, chi+8.5, White, 3, Terracotta, 2.4)
	s.hall(19, 30, 14, 17.4, cb, chi+7.5, White, 2, Terracotta, 2.4)
	cc := P(24.5, 15.7)
	ctop := chi + 8.5 + 3
	g.Use(White).Fill(Cylinder{Center: cc, Radius: 2.7, Y0: chi + 7.5, Y1: ctop})
	for k := 0; k < 8; k++ {
		a := float64(k) * math.Pi / 4
		w := Arch{Sill: V(2.4, ctop-2.4, 0), HalfWidth: 0.45, Spring: 0.9, Depth: 1.6, Through: AxisX}
		g.Carve(Translate(RotateY(w, P(0, 0), a), cc.At(0)))
	}
	g.With(Charcoal, func() { g.Fill(Cylinder{Center: cc, Radius: 1.9, Y0: ctop - 3, Y1: ctop}) })
	g.With(DarkStone, func() { g.Fill(Ellipsoid{cc.At(ctop), V(2.9, 2.8, 2.9)}) })
	g.With(White, func() { g.Fill(Cylinder{Center: cc, Radius: 0.8, Y0: ctop + 2.6, Y1: ctop + 4}) })
	g.With(Gold, func() { g.Fill(Sphere(cc.At(ctop+4.5), 0.6)) })
	g.With(White, func() {
		g.Fill(Gable{P(21.2, 10.6), P(27.8, 11.6), chi + 8.5, AxisZ})
		g.Fill(Block(V(21.2, chi+8.5, 10.6), V(27.8, chi+9.3, 11.6)))
	})
	g.Carve(Arch{Sill: V(24.5, cb, 11), HalfWidth: 0.9, Spring: 1.8, Depth: 1.6, Through: AxisZ})

	// Gundulić square: market parasols around the poet's statue.
	sg := s.town(3, 6)
	g.With(Stone, func() { g.Fill(Cylinder{Center: P(3, 6.2), Radius: 0.9, Y0: sg - 0.5, Y1: sg + 1.6}) })
	g.With(DarkStone, func() { g.Fill(Cylinder{Center: P(3, 6.2), Radius: 0.5, Y0: sg + 1.6, Y1: sg + 3.3}) })
	for i, q := range []Vec2{P(-0.2, 4.4), P(2, 3.9), P(5.6, 4.2), P(6.4, 6.6), P(-0.4, 8), P(5.4, 8.4), P(0.6, 6.2)} {
		y := s.town(q.X, q.Y)
		g.With(Wood, func() { g.Fill(Cylinder{Center: q, Radius: 0.45, Y0: y, Y1: y + 2}) })
		c := White
		if i%2 == 0 {
			c = Red
		}
		g.With(c, func() { g.Fill(Cone(q, 1.5, y+2, 0.9)) })
	}

	// Jesuit church above its stairs and square.
	jlo, jhi := s.span(-1, 5.4, 13.5, 21)
	s.hall(-1, 5.4, 13.5, 21, jlo, jhi+9.5, Sand, 3, Terracotta, 2.4)
	g.With(Sand, func() { g.Fill(Gable{P(-1.4, 13.1), P(5.8, 14.1), jhi + 9.5, AxisZ}) })
	g.Carve(Arch{Sill: V(2.2, jlo, 13.5), HalfWidth: 0.9, Spring: 2, Depth: 1.6, Through: AxisZ})
	s.belfry(P(6.5, 19.5), 2.2, s.town(6.5, 19.5), jhi+12, Terracotta)

	// Dominican monastery: church and bell tower near the Ploče Gate.
	dlo, dhi := s.span(17, 27, -16.5, -12)
	s.hall(17, 27, -16.5, -12, dlo, dhi+7, Stone, 2, Terracotta, 2.6)
	s.belfry(P(28.8, -12.4), 3, s.town(28.8, -12.4), dhi+12, DarkStone)

	// Arsenal: three great arches opening onto the harbor.
	s.hall(33, 36.5, -2, 8, gy, gy+5.5, Stone, 1, Terracotta, 4)
	for _, z := range []float64{-0.4, 3, 6.4} {
		g.Carve(Arch{Sill: V(37.5, 1.6, z), HalfWidth: 1.3, Spring: 1.6, Depth: 4, Through: AxisX})
	}
	g.With(Charcoal, func() { g.Fill(Block(V(34, gy+0.1, -1.6), V(34.6, gy+3, 7.6))) })
	_ = cl0
	_ = cl1
}

// ---- Lovrijenac ----

func (s *dubrovnikSite) lovrijenac() {
	g := s.g
	poly := []Vec2{P(-69.5, 0.5), P(-59.5, -0.5), P(-57.5, 6), P(-60, 11.5), P(-68.5, 11)}
	inner := []Vec2{P(-68.2, 1.6), P(-60.4, 0.8), P(-58.9, 6), P(-60.8, 10.3), P(-67.4, 9.9)}
	base, top := dubrovnikRock-2.5, dubrovnikRock+9
	g.Use(Stone).Fill(Prism{poly, base, top})
	g.Carve(Prism{inner, dubrovnikRock + 0.1, top + 1})
	g.With(Stone, func() { g.Fill(Prism{inner, dubrovnikRock - 0.8, dubrovnikRock + 0.1}) })
	g.Fill(Func{Box{V(-70, top, -1), V(-57, top+2.2, 12)}, func(p Vec3) bool {
		if !InsidePolygon(poly, p.Plan()) || InsidePolygon(inner, p.Plan()) {
			return false
		}
		return p.Y <= top+1.2 || math.Mod(p.X-p.Z+40, 2.4) < 1.2
	}})
	g.Paint(Sand, Func{Box{V(-70, base, -1), V(-57, top, 12)}, func(p Vec3) bool {
		return Hash01(int(math.Floor(p.Y/1.5)), int(math.Floor((p.X+p.Z)/2.5)), 23) < 0.3
	}})
	g.Paint(White, Prism{poly, top + 0.1, top + 2.4})
	// Inner galleries: arcades on two storeys around the court.
	g.Use(Stone).Fill(Prism{[]Vec2{P(-60.4, 0.8), P(-58.9, 6), P(-60.8, 10.3), P(-62.6, 10.2), P(-62.2, 1)}, dubrovnikRock, top - 2})
	g.Carve(Prism{[]Vec2{P(-61.2, 2), P(-60, 6), P(-61.6, 9.2), P(-62.6, 9.2), P(-62.4, 2)}, dubrovnikRock + 0.1, top - 3})
	for _, z := range []float64{2.6, 4.6, 6.6, 8.6} {
		for _, y := range []float64{dubrovnikRock, dubrovnikRock + 4} {
			g.Carve(Arch{Sill: V(-62.4, y, z), HalfWidth: 0.6, Spring: 1.6, Depth: 1.2, Through: AxisX})
		}
	}
	// Openings: the great arch facing the city and windows on the sea sides.
	g.Carve(Arch{Sill: V(-58.5, dubrovnikRock+3, 5.5), HalfWidth: 1.2, Spring: 2.4, Depth: 3, Through: AxisX})
	for _, x := range []float64{-67, -64, -61.5} {
		g.Carve(Arch{Sill: V(x, dubrovnikRock+4, 0.2), HalfWidth: 0.55, Spring: 1, Depth: 3, Through: AxisZ})
		g.Carve(Arch{Sill: V(x, dubrovnikRock+4, 11.2), HalfWidth: 0.55, Spring: 1, Depth: 3, Through: AxisZ})
	}
	s.darkInside(Box{V(-58.4, dubrovnikRock, 2), V(-57, top, 9)}, Prism{[]Vec2{P(-61.2, 2), P(-60, 6), P(-61.6, 9.2), P(-62.6, 9.2), P(-62.4, 2)}, dubrovnikRock + 0.1, top - 3})
	g.Paint(Charcoal, Prism{inner, dubrovnikRock + 0.1, dubrovnikRock + 0.3})
	g.Paint(Stone, Prism{inner, dubrovnikRock - 0.8, dubrovnikRock + 0.1})
	// Flag.
	g.With(Wood, func() { g.Fill(Cylinder{Center: P(-65, 5), Radius: 0.45, Y0: dubrovnikRock, Y1: top + 5}) })
	g.With(Red, func() { g.Fill(Block(V(-64.6, top+3, 4.8), V(-62.6, top+4.6, 5.2))) })
	// Stair up the north flank from the land.
	for k := 0; k < 12; k++ {
		y := 1.4 + float64(k)*0.95
		z := -9.5 + float64(k)*0.75
		g.With(Stone, func() { g.Fill(Block(V(-64, math.Max(-0.2, y-1.5), z), V(-61.8, y, z+0.9))) })
	}
}

// ---- Harbor ----

func (s *dubrovnikSite) harbor() {
	g := s.g
	// Quay along the walls.
	g.Use(Stone).Fill(Block(V(37.8, -0.2, -6.5), V(42, 1.6, 12.5)))
	g.Paint(DarkStone, Block(V(41.2, -0.2, -6.5), V(42, 0.8, 12.5)))
	// Kaše breakwater.
	g.Use(DarkStone).Fill(Block(V(51, -0.2, -7), V(53.6, 1.6, 11.5)))
	g.Paint(Stone, Block(V(51, 1.2, -7), V(53.6, 1.6, 11.5)))
	g.With(White, func() { g.Fill(Cylinder{Center: P(52.3, 10.4), Radius: 0.8, Y0: 1.6, Y1: 4}) })
	g.With(Red, func() { g.Fill(Cylinder{Center: P(52.3, 10.4), Radius: 0.9, Y0: 4, Y1: 4.8}) })
	// Porporela pier from St John's.
	pier := Func{Box{V(42, -0.2, 11), V(50, 1.6, 22)}, func(p Vec3) bool {
		a, b := P(44, 19), P(48.5, 13.5)
		ab := b.Sub(a)
		t := s.clamp01(p.Plan().Sub(a).Dot(ab) / ab.Dot(ab))
		return p.Plan().Sub(a.Add(ab.Scale(t))).Length() < 1.2
	}}
	g.Use(Stone).Fill(pier)
	// Moored boats along the quay and a few out at sea.
	for i, z := range []float64{-4.5, -1.5, 1.5, 4.5, 7.5, 10.2} {
		s.boat(P(43.2, z), math.Pi/2*0+0.08*(Hash01(i, 30)-0.5), 3.6, 1.6, i)
	}
	s.boat(P(49, -3.5), math.Pi/2, 4, 1.7, 11)
	s.boat(P(10, 30), 0.3, 4.2, 1.8, 12)
	s.boat(P(-24, 29), -0.2, 3.8, 1.6, 13)
	s.boat(P(-53, 22), 1.2, 3.6, 1.6, 14)
	s.carrack(P(47, 3.5))
}

// boat is a small open fishing boat; angle 0 points it along +X.
func (s *dubrovnikSite) boat(at Vec2, angle, length, beam float64, seed int) {
	g := s.g
	hull := Func{Box{V(-length/2, -0.2, -beam/2), V(length/2, 1, beam/2)}, func(p Vec3) bool {
		u := 2 * math.Abs(p.X) / length
		w := beam / 2 * math.Sqrt(math.Max(0, 1-u*u*u))
		return math.Abs(p.Z) <= w
	}}
	place := func(sol Solid) Solid { return Translate(RotateY(sol, P(0, 0), angle), at.At(0)) }
	colors := []Color{White, Navy, Red, White, Wood, Glass}
	g.With(colors[seed%len(colors)], func() { g.FillVolume(place(hull)) })
	g.Paint(Wood, place(Block(V(-length/2+0.6, 0.6, -beam/2+0.4), V(length/2-0.6, 1, beam/2-0.4))))
	if seed%3 == 0 {
		g.With(White, func() { g.Fill(place(Block(V(-0.6, 0.8, -0.5), V(0.4, 1.8, 0.5)))) })
	}
}

// carrack is a replica of a Ragusan carrack, moored with sails set, bow to the north.
func (s *dubrovnikSite) carrack(at Vec2) {
	g := s.g
	const half, beam, deck = 6.2, 2.1, 2.4
	width := func(u, y float64) float64 {
		t := math.Abs(u) / half
		if u < 0 {
			t = math.Abs(u) / (half + 0.6)
		}
		return beam * math.Sqrt(math.Max(0, 1-math.Pow(t, 2.4))) * (0.72 + 0.28*s.clamp01(y/2.2))
	}
	box := Box{V(at.X-3, -0.2, at.Y-8), V(at.X+3, 6, at.Y+8)}
	g.Use(Wood).FillVolume(Func{box, func(p Vec3) bool {
		u, v := p.Z-at.Y, p.X-at.X
		w := width(u, p.Y)
		if math.Abs(v) > w {
			return false
		}
		top := deck
		switch {
		case u > 3.2:
			top = 4.4
		case u < -4.2:
			top = 3.6
		}
		if p.Y > top+0.8 {
			return false
		}
		if p.Y > top && math.Abs(v) < w-0.8 && (u < half-0.8 && u > -half+0.2) {
			return false
		}
		return true
	}})
	g.Carve(Func{box, func(p Vec3) bool {
		u, v := p.Z-at.Y, p.X-at.X
		return p.Y > 0.6 && p.Y < deck-0.8 && math.Abs(v) < width(u, p.Y)-1
	}})
	g.Paint(Charcoal, Func{box, func(p Vec3) bool { return p.Y > 1.2 && p.Y < 1.7 }})
	g.Paint(Gold, Func{box, func(p Vec3) bool { return p.Y > 4.8 && p.Z-at.Y > 3.2 }})
	g.Paint(Red, Func{box, func(p Vec3) bool { return p.Z-at.Y > half-0.7 && p.Y > 2.4 && p.Y < 4.6 }})
	// Masts, yards and sails.
	type mast struct{ u, h, sail float64 }
	for _, m := range []mast{{-3.4, 10.5, 3.6}, {0, 13, 4.6}, {3.9, 8.5, 0}} {
		c := P(at.X, at.Y+m.u)
		g.With(Wood, func() { g.Fill(Cylinder{Center: c, Radius: 0.5, Y0: deck, Y1: deck + m.h}) })
		g.With(Red, func() { g.Fill(Block(V(c.X-0.2, deck+m.h-0.2, c.Y), V(c.X+0.2, deck+m.h+0.9, c.Y+1.6))) })
		if m.sail == 0 {
			// Lateen mizzen.
			g.With(White, func() {
				g.Fill(Func{Box{V(c.X-0.5, deck+1.5, c.Y-0.4), V(c.X+0.5, deck+m.h, c.Y+3.2)}, func(p Vec3) bool {
					a := (p.Y - deck - 1.5) / (m.h - 1.5)
					return p.Z-c.Y < 3*(1-a) && p.Z-c.Y > 0.2
				}})
			})
			continue
		}
		for _, sail := range []struct{ y0, y1, w float64 }{{deck + 2.6, deck + 0.62*m.h, m.sail}, {deck + 0.66*m.h, deck + 0.93*m.h, m.sail * 0.7}} {
			sail := sail
			g.With(Wood, func() { g.Fill(Block(V(c.X-sail.w-0.3, sail.y1, c.Y+0.5), V(c.X+sail.w+0.3, sail.y1+0.5, c.Y+1.2))) })
			g.With(White, func() {
				g.Fill(Func{Box{V(c.X-sail.w, sail.y0, c.Y-2), V(c.X+sail.w, sail.y1, c.Y+1.5)}, func(p Vec3) bool {
					v := (p.X - c.X) / sail.w
					t := (p.Y - sail.y0) / (sail.y1 - sail.y0)
					belly := 1.1 * (1 - v*v) * math.Sin(math.Pi*(0.25+0.75*t))
					off := p.Z - (c.Y + 0.8 - belly)
					return math.Abs(v) <= 1 && off >= -0.45 && off <= 0.45
				}})
			})
			if sail.w == m.sail {
				g.Paint(Red, Func{Box{V(c.X-sail.w, sail.y0, c.Y-2), V(c.X+sail.w, sail.y1, c.Y+0.8)}, func(p Vec3) bool {
					return math.Abs(p.X-c.X) < 0.5 || math.Abs(p.Y-(sail.y0+sail.y1)/2) < 0.5
				}})
			}
		}
	}
	g.With(Wood, func() { g.Fill(Capsule{V(at.X, 3.4, at.Y-half), V(at.X, 5.2, at.Y-half-3.2), 0.45}) })
}

// ---- Trees ----

func (s *dubrovnikSite) trees() {
	g := s.g
	for gx := -78.0; gx < 62; gx += 4.6 {
		for gz := -54.0; gz < 32; gz += 4.6 {
			i, j := int(gx*10), int(gz*10)
			x := gx + 3*(Hash01(i, j, 40)-0.5)
			z := gz + 3*(Hash01(i, j, 41)-0.5)
			if s.rim(x, z) < 3 || s.land(x, z) < 0.8 {
				continue
			}
			w := s.wall(P(x, z))
			h := s.at(x, z)
			if w.sd < 6 || h < 1.2 || h-s.low(x, z) > 1.2 || P(x, z).Sub(dubrovnikLovrijenac).Length() < 9 {
				continue
			}
			if math.Abs(z+2) < 4 && x < -45 || x > 40 && z > -24 && z < -4 || math.Abs(x+63) < 3 && z < -2 {
				continue // Bridges, Revelin and the Lovrijenac stair.
			}
			if Hash01(i, j, 42) > 0.62 {
				continue
			}
			base := h - 0.4
			if Hash01(i, j, 43) < 0.55 {
				// Umbrella pine.
				ht := 5 + 2*Hash01(i, j, 44)
				g.With(Wood, func() { g.Fill(Cylinder{Center: P(x, z), Radius: 0.5, Y0: base, Y1: base + ht*0.75}) })
				g.With(Foliage, func() { g.Fill(Ellipsoid{V(x, base+ht*0.8, z), V(ht*0.42, ht*0.17, ht*0.42)}) })
			} else {
				// Cypress.
				ht := 5.5 + 2.5*Hash01(i, j, 45)
				g.With(Wood, func() { g.Fill(Cylinder{Center: P(x, z), Radius: 0.5, Y0: base, Y1: base + 1.2}) })
				g.With(Foliage, func() { g.Fill(Ellipsoid{V(x, base+ht*0.55, z), V(1, ht*0.48, 1)}) })
			}
		}
	}
	// A few cypresses in the town gardens.
	for k, q := range []Vec2{P(-34, -20), P(-24, -25), P(10, -22.5), P(-38, 15), P(14, 20)} {
		w := s.wall(q)
		if w.sd > -2 {
			continue
		}
		base := s.town(q.X, q.Y) - 0.3
		ht := 5 + Hash01(k, 46)*2
		g.With(Foliage, func() { g.Fill(Ellipsoid{V(q.X, base+ht*0.55, q.Y), V(1, ht*0.48, 1)}) })
	}
}
