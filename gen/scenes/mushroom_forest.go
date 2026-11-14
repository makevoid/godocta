package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// MushroomForest is a fairy-tale grove of giant mushrooms with tiny homes in their hollow
// stalks, on an elliptical plate of mossy forest floor. A great red fly agaric with white
// warts and a skirt ring stands back-centre: a round red door, lit windows, a spiral stair
// winding up to a balcony, a crooked chimney and lanterns under the gilled cap. Around it
// are a young round-capped agaric, a tall gold-scaled parasol joined to the great cap by a
// rope bridge, a closed drumstick parasol, a fat terracotta bolete cottage with its pup, a
// cluster of pale blue bell caps and a clump of upturned golden funnels. Behind, a mossy
// hollow log leans on its dead stump with bracket fungi stepping up both. A brook runs
// from a spring under an arched stone bridge and past stepping stones into a lily pond
// with a jetty, a frog, a giant snail and ladybirds. The composition is laid out in the
// hero camera's frame: the doors face +X/+Z and the tall caps stand at the back (-X/-Z).
type MushroomForest struct{}

func init() { Register(MushroomForest{}) }

func (MushroomForest) Name() string  { return "mushroom_forest" }
func (MushroomForest) Title() string { return "Mushroom Forest" }

const (
	mushroomForestA     = 53.0 // Plate semi-axes.
	mushroomForestB     = 49.0
	mushroomForestWater = 1.3 // Brook and pond surface.
	mushroomForestPondR = 10.5
)

// mushroomForestAt maps hero-view coordinates (u toward the camera's right, d toward the
// camera) to a plan point, so the grove can be composed as the hero camera sees it.
func mushroomForestAt(u, d float64) Vec2 {
	s, c := math.Sincos(0.68)
	return P(u*c+d*s, -u*s+d*c)
}

func mushroomForestLine(pts ...Vec2) []Vec2 {
	out := make([]Vec2, len(pts))
	for i, p := range pts {
		out[i] = mushroomForestAt(p.X, p.Y)
	}
	return out
}

// mushroomForestSpot is a plan circle: a levelled terrace or an area kept clear.
type mushroomForestSpot struct {
	c    Vec2
	r, h float64
}

var (
	mushroomForestFace   = mushroomForestAt(0, 1).Angle() // Plan angle toward the hero camera.
	mushroomForestGreat  = mushroomForestAt(-8, -16)      // Great fly agaric.
	mushroomForestYoung  = mushroomForestAt(-36, 2)       // Young fly agaric.
	mushroomForestTall   = mushroomForestAt(36, -12)      // Open parasol.
	mushroomForestDrum   = mushroomForestAt(43, 5)        // Closed parasol.
	mushroomForestBun    = mushroomForestAt(16, 18)       // Bolete cottage.
	mushroomForestPup    = mushroomForestAt(35, 15)       // Small bolete.
	mushroomForestBells  = mushroomForestAt(-18, 26)      // Bell cap cluster.
	mushroomForestHorns  = mushroomForestAt(-29, -29)     // Funnel caps.
	mushroomForestStump  = mushroomForestAt(9, -34)       // Dead stump.
	mushroomForestLogEnd = mushroomForestAt(30, -24)      // Low end of the fallen log.
	mushroomForestPond   = mushroomForestAt(1, 34)
	mushroomForestSnail  = mushroomForestAt(27, 32)
	mushroomForestBug    = mushroomForestAt(-25, 37)    // Ladybird on the lawn.
	mushroomForestArch   = mushroomForestAt(-20, 0)     // Bridge over the brook.
	mushroomForestStones = mushroomForestAt(-3.5, 20.5) // Stepping stones.

	mushroomForestStream = mushroomForestLine(P(-41, -21), P(-34, -15), P(-24, -6), P(-16, 6), P(-6, 16),
		P(-1, 25), P(1, 30))
	mushroomForestPaths = [][]Vec2{
		mushroomForestLine(P(-7.5, -7), P(-8, -1), P(-2, 6), P(2, 15), P(7, 22.5), P(11.5, 27.5), P(13, 32.5)),
		mushroomForestLine(P(-8, -1), P(-13, -3.6), P(-15.2, -3.2)),
		mushroomForestLine(P(-24.8, 3.2), P(-28, 4.6), P(-32.4, 5.6)),
		mushroomForestLine(P(11.5, 27.5), P(15.5, 27)),
		mushroomForestLine(P(2, 15), P(0.2, 18.2)),
		mushroomForestLine(P(-7.3, 22.8), P(-11, 25)),
		mushroomForestLine(P(-2, 6), P(12, 1), P(25, -6), P(32, -9.5)),
	}
	// Levelled terraces under the homes and the deadwood.
	mushroomForestTerraces = []mushroomForestSpot{
		{mushroomForestGreat, 17, 4.8}, {mushroomForestYoung, 10, 3.6}, {mushroomForestBun, 13, 3.4},
		{mushroomForestStump, 11, 3.4}, {mushroomForestTall, 7, 3.6}, {mushroomForestSnail, 9, 2.8},
	}
	// Ground kept clear of ferns and clutter.
	mushroomForestKeep = []mushroomForestSpot{
		{mushroomForestGreat, 12.5, 0}, {mushroomForestYoung, 6.5, 0}, {mushroomForestTall, 5, 0},
		{mushroomForestDrum, 4.5, 0}, {mushroomForestBun, 9.5, 0}, {mushroomForestPup, 5, 0},
		{mushroomForestBells, 8.5, 0}, {mushroomForestHorns, 9.5, 0}, {mushroomForestStump, 10.5, 0},
		{mushroomForestSnail, 9, 0}, {mushroomForestArch, 7, 0}, {mushroomForestBug, 4.5, 0},
	}
	// Fairy rings of little toadstools.
	mushroomForestRings = []mushroomForestSpot{
		{mushroomForestAt(13, -9), 4.2, 0}, {mushroomForestAt(-36, 20), 3.6, 0}, {mushroomForestAt(-22, 13), 3.4, 0},
	}
)

func mushroomForestSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func mushroomForestLerp(a, b, t float64) float64 { return a + (b-a)*t }

func mushroomForestUnit(v Vec2) Vec2 { return v.Scale(1 / math.Max(v.Length(), 1e-9)) }

// mushroomForestE is 1 on the rim of the plate and less inside.
func mushroomForestE(x, z float64) float64 {
	return math.Hypot(x/mushroomForestA, z/mushroomForestB)
}

// mushroomForestNoise is smooth value noise in [-1, 1].
func mushroomForestNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return mushroomForestLerp(mushroomForestLerp(a, b, sx), mushroomForestLerp(c, d, sx), sz)*2 - 1
}

func mushroomForestFBM(x, z float64, seed int) float64 {
	return 0.6*mushroomForestNoise(x, z, seed) + 0.3*mushroomForestNoise(x*2.03+17, z*2.03+5, seed+1) +
		0.1*mushroomForestNoise(x*4.1+3, z*4.1+29, seed+2)
}

// mushroomForestPolyline returns the plan distance to a polyline and the arc length there.
func mushroomForestPolyline(p Vec2, pts []Vec2) (dist, along float64) {
	dist = math.Inf(1)
	run := 0.0
	for i := 0; i+1 < len(pts); i++ {
		a, b := pts[i], pts[i+1]
		ab := b.Sub(a)
		l := ab.Length()
		t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/(l*l)))
		if d := p.Sub(a.Add(ab.Scale(t))).Length(); d < dist {
			dist, along = d, run+t*l
		}
		run += l
	}
	return
}

// mushroomForestMap is the terrain height sampled on a regular grid.
type mushroomForestMap struct {
	x0, z0, step float64
	nx, nz       int
	h            []float64
}

func (m *mushroomForestMap) at(x, z float64) float64 {
	fx := math.Max(0, math.Min(float64(m.nx)-1.001, (x-m.x0)/m.step))
	fz := math.Max(0, math.Min(float64(m.nz)-1.001, (z-m.z0)/m.step))
	i, j := int(fx), int(fz)
	tx, tz := fx-float64(i), fz-float64(j)
	k := j*m.nx + i
	a := m.h[k] + (m.h[k+1]-m.h[k])*tx
	b := m.h[k+m.nx] + (m.h[k+m.nx+1]-m.h[k+m.nx])*tx
	return a + (b-a)*tz
}

func (m *mushroomForestMap) on(q Vec2) float64 { return m.at(q.X, q.Y) }

// mushroomForestFrame places local coordinates (x forward, y up, z to the left) in the world.
type mushroomForestFrame struct {
	o Vec3
	f Vec2    // Unit heading.
	k float64 // Rise per unit forward, to stand on a slope.
}

func (fr mushroomForestFrame) w(x, y, z float64) Vec3 {
	return V(fr.o.X+fr.f.X*x-fr.f.Y*z, fr.o.Y+y+fr.k*x, fr.o.Z+fr.f.Y*x+fr.f.X*z)
}

func (fr mushroomForestFrame) l(p Vec3) Vec3 {
	q := p.Sub(fr.o)
	x := q.X*fr.f.X + q.Z*fr.f.Y
	return V(x, q.Y-fr.k*x, -q.X*fr.f.Y+q.Z*fr.f.X)
}

// ell is an ellipsoid with local center c and local radii r.
func (fr mushroomForestFrame) ell(c, r Vec3) Solid {
	o := fr.w(c.X, c.Y, c.Z)
	k := math.Max(r.X, math.Max(r.Y, r.Z)) * (1 + math.Abs(fr.k))
	return Func{Box{o.Sub(V(k, k, k)), o.Add(V(k, k, k))}, func(p Vec3) bool {
		q := fr.l(p).Sub(c)
		x, y, z := q.X/r.X, q.Y/r.Y, q.Z/r.Z
		return x*x+y*y+z*z <= 1
	}}
}

// mushroomForestStem is a curved, tapered stalk with a swollen foot; hollow when wall > 0.
type mushroomForestStem struct {
	foot   Vec3    // Center of the foot, a little below the ground.
	h      float64 // Height from the foot to the underside of the cap.
	r0, r1 float64 // Radius at the foot and under the cap.
	bulb   float64 // Extra radius of the swollen foot.
	lean   Vec2    // Plan offset of the top: the stalk curves toward it.
	wall   float64 // Shell thickness; 0 is a solid stalk.
}

func (st mushroomForestStem) t(y float64) float64 {
	return math.Max(0, math.Min(1, (y-st.foot.Y)/st.h))
}

func (st mushroomForestStem) axis(y float64) Vec2 {
	t := st.t(y)
	return st.foot.Plan().Add(st.lean.Scale(t * t))
}

func (st mushroomForestStem) rad(y float64) float64 {
	return st.r0 + (st.r1-st.r0)*st.t(y) + st.bulb*math.Exp(-math.Max(0, y-st.foot.Y)/3)
}

func (st mushroomForestStem) top() Vec3 {
	y := st.foot.Y + st.h
	return st.axis(y).At(y)
}

func (st mushroomForestStem) box() Box {
	return BoxOf(st.foot, st.top()).Grow(math.Max(st.r0, st.r1) + st.bulb + 0.5)
}

// around returns a point's distance from the stalk's axis and its plan angle around it.
func (st mushroomForestStem) around(p Vec3) (d, a float64) {
	q := p.Plan().Sub(st.axis(p.Y))
	return q.Length(), q.Angle()
}

func (st mushroomForestStem) solid() Solid {
	return Func{st.box(), func(p Vec3) bool {
		if p.Y < st.foot.Y || p.Y > st.foot.Y+st.h {
			return false
		}
		d, _ := st.around(p)
		r := st.rad(p.Y)
		return d <= r && (st.wall <= 0 || d > r-st.wall)
	}}
}

// surface is the point on the stalk's skin at height y in plan direction a, pushed out by off.
func (st mushroomForestStem) surface(y, a, off float64) Vec3 {
	return Polar(st.axis(y), st.rad(y)+off, a).At(y)
}

// mushroomForestCap is a body of revolution around a slightly tilted axis, kept as a skin.
type mushroomForestCap struct {
	c      Vec3                       // Local origin: the top of the stalk.
	R      float64                    // Largest rim radius.
	in     func(r, a, y float64) bool // The flesh before hollowing, in local coordinates.
	top    func(r float64) float64    // Outer surface height, for placing things on the cap.
	t      float64                    // Flesh thickness.
	dir    Vec2                       // The rim rises toward dir by ang radians.
	ang    float64
	lo, hi float64 // Extent below and above the origin.
}

func (c *mushroomForestCap) local(p Vec3) (r, a, y float64) {
	q := p.Sub(c.c)
	al := q.X*c.dir.X + q.Z*c.dir.Y
	ac := -q.X*c.dir.Y + q.Z*c.dir.X
	sn, cs := math.Sincos(c.ang)
	fw := al*cs + q.Y*sn
	return math.Hypot(fw, ac), math.Atan2(ac, fw), -al*sn + q.Y*cs
}

func (c *mushroomForestCap) world(r, a, y float64) Vec3 {
	fw, ac := r*math.Cos(a), r*math.Sin(a)
	sn, cs := math.Sincos(c.ang)
	al := fw*cs - y*sn
	return V(c.c.X+c.dir.X*al-c.dir.Y*ac, c.c.Y+fw*sn+y*cs, c.c.Z+c.dir.Y*al+c.dir.X*ac)
}

// on is the point lift above the outer surface at local radius r toward plan angle a.
func (c *mushroomForestCap) on(r, a, lift float64) Vec3 {
	return c.world(r, a-c.dir.Angle(), c.top(r)+lift)
}

func (c *mushroomForestCap) box() Box {
	pad := c.R*math.Abs(math.Sin(c.ang)) + 1
	return Box{V(c.c.X-c.R-1.5, c.c.Y+c.lo-pad, c.c.Z-c.R-1.5), V(c.c.X+c.R+1.5, c.c.Y+c.hi+pad, c.c.Z+c.R+1.5)}
}

// flesh is the skin of the cap: its outer and upper t units.
func (c *mushroomForestCap) flesh() Solid {
	t := c.t
	return Func{c.box(), func(p Vec3) bool {
		r, a, y := c.local(p)
		if !c.in(r, a, y) {
			return false
		}
		return !(c.in(r+t, a, y) && c.in(r, a, y+t) && c.in(r+0.7*t, a, y+0.7*t))
	}}
}

// gills are n radial plates under the cap, from the stalk (radius rs) to the rim, hanging
// down to drop(r). Every other plate is a short one that starts halfway out.
func (c *mushroomForestCap) gills(n int, rs float64, drop func(r float64) float64) Solid {
	step := 2 * math.Pi / float64(n)
	return Func{c.box(), func(p Vec3) bool {
		r, a, y := c.local(p)
		if r < rs || r > c.R-0.4 || y < drop(r) || y > math.Min(c.top(r), drop(r)+3.2) {
			return false
		}
		k := math.Round(a / step)
		if int(k)&1 != 0 && r < 0.5*c.R {
			return false
		}
		return r*math.Abs(math.Sin(a-k*step)) <= 0.55
	}}
}

// noise helpers and site predicates.

func (MushroomForest) pondD(p Vec2) float64 {
	q := p.Sub(mushroomForestPond)
	a := q.Angle()
	return q.Length() - mushroomForestPondR*(1+0.14*mushroomForestNoise(math.Cos(a)*1.7+3, math.Sin(a)*1.7, 23))
}

func (MushroomForest) streamD(p Vec2) float64 {
	d, _ := mushroomForestPolyline(p, mushroomForestStream)
	return d
}

func (MushroomForest) pathD(p Vec2) float64 {
	best := math.Inf(1)
	for _, line := range mushroomForestPaths {
		if d, _ := mushroomForestPolyline(p, line); d < best {
			best = d
		}
	}
	return best
}

func (s MushroomForest) raw(x, z float64) float64 {
	p := P(x, z)
	h := 2.7 + 1.1*mushroomForestFBM(x*0.07, z*0.07, 11)
	for _, k := range mushroomForestTerraces {
		d := p.Sub(k.c).Length()
		h = mushroomForestLerp(h, k.h, mushroomForestSmooth((k.r-d)/(0.42*k.r)))
	}
	h += 0.28 * mushroomForestNoise(x*0.31, z*0.31, 15)
	if d, along := mushroomForestPolyline(p, mushroomForestStream); d < 4.5 {
		w := 0.7 + 0.014*along
		h = mushroomForestLerp(0.5, h, mushroomForestSmooth((d-w)/2.6))
	}
	if d := s.pondD(p); d < 2 {
		h = mushroomForestLerp(0.3, h, mushroomForestSmooth((d+2.2)/3.4))
	}
	return mushroomForestLerp(1.7, h, mushroomForestSmooth((1-mushroomForestE(x, z))/0.09))
}

// heights samples the terrain on a half-unit grid.
func (s MushroomForest) heights() *mushroomForestMap {
	m := &mushroomForestMap{x0: -mushroomForestA - 2, z0: -mushroomForestB - 2, step: 0.5}
	m.nx = int(math.Round((2*mushroomForestA+4)/m.step)) + 1
	m.nz = int(math.Round((2*mushroomForestB+4)/m.step)) + 1
	m.h = make([]float64, m.nx*m.nz)
	for j := 0; j < m.nz; j++ {
		for i := 0; i < m.nx; i++ {
			m.h[j*m.nx+i] = s.raw(m.x0+float64(i)*m.step, m.z0+float64(j)*m.step)
		}
	}
	return m
}

func (s MushroomForest) wet(m *mushroomForestMap, x, z float64) bool {
	p := P(x, z)
	return m.at(x, z) < mushroomForestWater-0.05 && (s.streamD(p) < 4.5 || s.pondD(p) < 2)
}

// shell fills the cells of a predicate that lie within t of its outside; under also skins
// the underside.
func (MushroomForest) shell(g *Grid, box Box, t float64, under bool, in func(Vec3) bool) {
	g.Fill(Func{box, func(p Vec3) bool {
		if !in(p) {
			return false
		}
		return !(in(p.Add(V(t, 0, 0))) && in(p.Add(V(-t, 0, 0))) && in(p.Add(V(0, 0, t))) &&
			in(p.Add(V(0, 0, -t))) && in(p.Add(V(0, t, 0))) && (!under || in(p.Add(V(0, -t, 0)))))
	}})
}

// hash is a deterministic per-cell value.
func (MushroomForest) hash(g *Grid, p Vec3, seed int) float64 {
	c := Nearest(p.Scale(1 / g.Span()))
	return Hash01(int(c.X), int(c.Y), int(c.Z), seed)
}

// part builds one piece in its own grid, so its paints never touch its neighbours.
func (MushroomForest) part(g *Grid, build func(t *Grid)) {
	t := NewGridWithSpan(g.Span())
	build(t)
	g.Merge(t)
}

func (s MushroomForest) Build(g *Grid) Order {
	m := s.heights()
	s.terrain(g, m)
	s.deadwood(g, m)
	var great, tall *mushroomForestCap
	s.part(g, func(t *Grid) { great = s.greatAgaric(t, m) })
	s.part(g, func(t *Grid) { s.youngAgaric(t, m) })
	s.part(g, func(t *Grid) { tall = s.parasol(t, m) })
	s.part(g, func(t *Grid) { s.drumstick(t, m) })
	s.part(g, func(t *Grid) { s.bolete(t, m) })
	s.part(g, func(t *Grid) { s.pup(t, m) })
	s.bells(g, m)
	s.horns(g, m)
	toTall := tall.c.Plan().Sub(great.c.Plan()).Angle()
	s.part(g, func(t *Grid) {
		s.ropeBridge(t, great.on(14.5, toTall, 0.2), tall.on(10.5, toTall+math.Pi, 0.2))
	})
	s.crossings(g, m)
	s.pond(g, m)
	s.part(g, func(t *Grid) { s.snail(t, m) })
	s.ladybird(g, tall.on(7.5, mushroomForestFace+0.2, 0.1), mushroomForestAt(1, 0.25), 1.1, 0)
	s.ladybird(g, mushroomForestBug.At(m.on(mushroomForestBug)+0.3), mushroomForestAt(1, -0.5), 1.3, 0)
	s.floor(g, m)
	return Sweep(mushroomForestGreat, math.Pi*0.75)
}

// terrain is a hollow skin of moss over the height field, with the brook, pond and paths.
func (s MushroomForest) terrain(g *Grid, m *mushroomForestMap) {
	site := Box{V(-mushroomForestA, 0, -mushroomForestB), V(mushroomForestA, 9, mushroomForestB)}
	in := func(p Vec3) bool {
		return p.Y >= 0 && mushroomForestE(p.X, p.Z) <= 1 && p.Y <= m.at(p.X, p.Z)
	}
	g.Use(Wood)
	s.shell(g, site, 0.85, false, in)
	top := func(p Vec3) bool { return p.Y > m.at(p.X, p.Z)-0.95 }
	// Soil strata on the rim.
	g.Paint(DarkStone, Func{site, func(p Vec3) bool { return !top(p) && p.Y < 0.7 }})
	// Moss: dark cushions with brighter drifts and patches of leaf litter.
	g.Paint(Foliage, Func{site, top})
	g.Paint(Grass, Func{site, func(p Vec3) bool {
		if !top(p) {
			return false
		}
		n := mushroomForestFBM(p.X*0.085, p.Z*0.085, 31)
		return n > 0.1 || n > -0.12 && s.hash(g, p, 32) < 0.2
	}})
	g.Paint(Wood, Func{site, func(p Vec3) bool {
		return top(p) && mushroomForestFBM(p.X*0.11+40, p.Z*0.11, 35) > 0.4 && s.hash(g, p, 36) < 0.8
	}})
	// Pebbled banks and paths.
	pebbles := func(p Vec3, seed int) Color {
		switch r := s.hash(g, p, seed); {
		case r < 0.3:
			return Stone
		case r < 0.45:
			return DarkStone
		}
		return Sand
	}
	for _, c := range []Color{Sand, Stone, DarkStone} {
		g.Paint(c, Func{site, func(p Vec3) bool {
			if !top(p) || s.wet(m, p.X, p.Z) {
				return false
			}
			q := p.Plan()
			if s.pathD(q) < 1.5+0.4*mushroomForestNoise(p.X*0.5, p.Z*0.5, 37) {
				return pebbles(p, 38) == c
			}
			bank := s.streamD(q) < 3.3 || s.pondD(q) < 1.3
			return bank && c != Sand && s.hash(g, p, 39) < 0.5 && pebbles(p, 38) == c
		}})
	}
	// Water, with a few pale ripples.
	pool := Box{site.Min, V(mushroomForestA, mushroomForestWater, mushroomForestB)}
	g.Use(Water).Fill(Func{pool, func(p Vec3) bool {
		return s.wet(m, p.X, p.Z) && p.Y > m.at(p.X, p.Z)-0.3
	}})
	g.Paint(Glass, Func{pool, func(p Vec3) bool {
		if !s.wet(m, p.X, p.Z) || p.Y < mushroomForestWater-0.5 {
			return false
		}
		_, along := mushroomForestPolyline(p.Plan(), mushroomForestStream)
		return s.pondD(p.Plan()) > 0 && math.Mod(along, 5) < 0.9 && s.hash(g, p, 40) < 0.7
	}})
}

// streaks paints short vertical fibres on a stalk.
func (s MushroomForest) streaks(g *Grid, st mushroomForestStem, c Color, seed int, density float64) {
	n := math.Max(8, math.Round((st.r0+st.r1)*math.Pi/1.3))
	g.Paint(c, Func{st.box(), func(p Vec3) bool {
		if p.Y < st.foot.Y || p.Y > st.foot.Y+st.h {
			return false
		}
		d, a := st.around(p)
		if d > st.rad(p.Y)+0.3 {
			return false
		}
		sector := int(math.Floor((a + math.Pi) / (2 * math.Pi) * n))
		band := int(math.Floor(p.Y/3.4 + 3*Hash01(sector, seed)))
		return Hash01(sector, band, seed+1) < density
	}})
}

// skirt is the ring left on the stalk by the veil: a flared cone hanging from height y.
func (s MushroomForest) skirt(g *Grid, st mushroomForestStem, y, drop, flare float64, body, hem Color) {
	in := func(p Vec3) bool {
		if p.Y > y || p.Y < y-drop {
			return false
		}
		d, _ := st.around(p)
		r := st.rad(p.Y) + 0.4 + (y-p.Y)*flare
		return d <= r && d > r-1.1
	}
	g.With(body, func() { g.Fill(Func{st.box().Grow(drop*flare + 1), in}) })
	g.Paint(hem, Func{st.box().Grow(drop*flare + 1), func(p Vec3) bool { return in(p) && p.Y < y-drop+0.8 }})
}

// door is a round door set into a stalk, facing plan angle a, with a frame, knob and step.
func (s MushroomForest) door(g *Grid, st mushroomForestStem, a, sill, rd float64, leaf, frame Color) {
	yc := sill + rd*0.8
	o, rc := st.axis(yc), st.rad(yc)
	n := P(math.Cos(a), math.Sin(a))
	tn := P(-n.Y, n.X)
	box := BoxOf(o.Add(n.Scale(rc-2)).At(sill-1), o.Add(n.Scale(rc+3)).At(yc+rd+1.5)).Grow(rd + 1.5)
	loc := func(p Vec3) (u, v, w float64) {
		q := p.Plan().Sub(o)
		v = q.Dot(tn)
		return q.Dot(n) - rc, v, math.Hypot(v, p.Y-yc)
	}
	// Clear the swollen foot in front of the door so it reads as set into the stalk.
	g.Carve(Func{box, func(p Vec3) bool {
		u, _, w := loc(p)
		return u > -1.0 && w <= rd+0.9 && p.Y > sill
	}})
	g.With(frame, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			u, _, w := loc(p)
			return u > -1.3 && u <= 0.5 && w > rd && w <= rd+0.9 && p.Y > sill-0.4
		}})
	})
	g.With(leaf, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			u, _, w := loc(p)
			return u > -1.3 && u <= -0.2 && w <= rd && p.Y > sill-0.4
		}})
	})
	knob := o.Add(n.Scale(rc + 0.1)).Add(tn.Scale(rd * 0.45)).At(yc)
	g.With(Gold, func() { g.Fill(Sphere(knob, 0.55)) })
	g.With(Stone, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			u, v, _ := loc(p)
			return u > -0.8 && u < 2.2 && math.Abs(v) < rd+0.7 && p.Y > sill-1.2 && p.Y <= sill
		}})
	})
}

// window is a round lit window with a wooden frame and sill.
func (s MushroomForest) window(g *Grid, st mushroomForestStem, a, yc, r float64) {
	o, rc := st.axis(yc), st.rad(yc)
	n := P(math.Cos(a), math.Sin(a))
	tn := P(-n.Y, n.X)
	box := BoxOf(o.Add(n.Scale(rc-2)).At(yc), o.Add(n.Scale(rc+1.5)).At(yc)).Grow(r + 2)
	loc := func(p Vec3) (u, v, w float64) {
		q := p.Plan().Sub(o)
		v = q.Dot(tn)
		return q.Dot(n) - rc, v, math.Hypot(v, p.Y-yc)
	}
	g.Paint(Gold, Func{box, func(p Vec3) bool {
		u, _, w := loc(p)
		return u > -1.8 && w <= r
	}})
	g.With(Wood, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			u, v, w := loc(p)
			if u < -1.0 {
				return false
			}
			if w > r && w <= r+0.8 && u <= 0.45 {
				return true
			}
			return math.Abs(v) <= r+1.0 && p.Y < yc-r-0.1 && p.Y >= yc-r-0.9 && u <= 1.0
		}})
	})
}

// balcony is a railed half-ring deck on a stalk at height y, facing plan angle a.
func (s MushroomForest) balcony(g *Grid, st mushroomForestStem, a, half, y float64) {
	o, r := st.axis(y), st.rad(y)
	box := BoxOf(o.At(y-4.5), o.At(y+2.2)).Grow(r + 4)
	box.Min.Y, box.Max.Y = y-4.5, y+2.2
	loc := func(p Vec3) (d, da, arc float64) {
		q := p.Plan().Sub(o)
		d = q.Length()
		rel := math.Remainder(q.Angle()-a, 2*math.Pi)
		return d, math.Abs(rel), (rel + math.Pi) * (r + 3)
	}
	g.With(Wood, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			d, da, arc := loc(p)
			if da > half || d < r-0.4 || d > r+3.3 {
				return false
			}
			if p.Y <= y && p.Y >= y-0.7 {
				return true
			}
			if d < r+2.6 || p.Y <= y || p.Y > y+2.0 {
				return false
			}
			return p.Y >= y+1.4 || math.Mod(arc, 2.4) < 0.8
		}})
		// Braces from the stalk to the outer edge.
		for _, k := range []float64{-0.75, 0, 0.75} {
			g.Fill(Capsule{st.surface(y-4, a+k*half, -0.3), Polar(o, r+2.7, a+k*half).At(y - 0.6), 0.5})
		}
	})
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		d, da, arc := loc(p)
		return da <= half && d > r && d <= r+3.3 && p.Y <= y && p.Y >= y-0.7 && math.Mod(arc, 2.4) >= 1.2
	}})
}

// stair is a spiral of plank steps climbing a stalk from y0 to y1, turning toward smaller
// plan angles and arriving at angle aTop after sweep radians.
func (s MushroomForest) stair(g *Grid, st mushroomForestStem, aTop, y0, y1, sweep float64) {
	n := int(math.Round(sweep * (st.rad((y0+y1)/2) + 1.3) / 2.4))
	da := sweep / float64(n)
	rise := (y1 - y0) / float64(n)
	box := st.box().Grow(3.5)
	box.Min.Y, box.Max.Y = y0-1, y1+2.5
	// phi returns the unwound positions along the stair that pass over a plan angle.
	phi := func(a float64) [3]float64 {
		rel := math.Mod(aTop+sweep-a, 2*math.Pi)
		if rel < 0 {
			rel += 2 * math.Pi
		}
		return [3]float64{rel, rel + 2*math.Pi, rel + 4*math.Pi}
	}
	g.With(Sand, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			d, a := st.around(p)
			for _, f := range phi(a) {
				k := int(math.Floor(f / da))
				if k >= n {
					break
				}
				ytop := y0 + float64(k+1)*rise
				r := st.rad(ytop)
				if p.Y <= ytop+0.05 && p.Y >= ytop-0.75 && d >= r-0.4 && d <= r+2.6 {
					return true
				}
			}
			return false
		}})
	})
	g.With(Wood, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			d, a := st.around(p)
			for _, f := range phi(a) {
				if f > sweep {
					break
				}
				k := math.Floor(f / da)
				ytop := y0 + (k+1)*rise
				yr := y0 + (f/da+0.5)*rise + 1.7
				if math.Abs(d-(st.rad(ytop)+2.3)) > 0.42 {
					continue
				}
				if math.Abs(p.Y-yr) <= 0.4 {
					return true
				}
				if int(k)%3 == 1 && f/da-k > 0.25 && f/da-k < 0.75 && p.Y > ytop && p.Y < yr {
					return true
				}
			}
			return false
		}})
	})
}

// lantern hangs on a cord from a point.
func (s MushroomForest) lantern(g *Grid, at Vec3, drop float64) {
	g.With(Wood, func() { g.Fill(Capsule{at, at.Add(V(0, -drop, 0)), 0.45}) })
	g.With(Gold, func() { g.FillVolume(Sphere(at.Add(V(0, -drop-0.7, 0)), 1.0)) })
}

// lanterns hangs n lanterns from the long gills of a cap at radius r.
func (s MushroomForest) lanterns(g *Grid, hat *mushroomForestCap, gills, n int, r, from, drop, phase float64) {
	step := 2 * math.Pi / float64(gills)
	for i := 0; i < n; i++ {
		k := 2 * math.Round((phase+float64(i)*2*math.Pi/float64(n))/step/2)
		s.lantern(g, hat.world(r, k*step, from), drop)
	}
}

// chimney is a crooked stone flue with a pot and a drifting trail of smoke.
func (s MushroomForest) chimney(g *Grid, base Vec3, lean Vec2, h float64, seed int) {
	mid := base.Add(V(lean.X, h*0.55, lean.Y))
	tip := base.Add(V(lean.X*0.2, h, lean.Y*0.2))
	box := BoxOf(base, tip).Grow(2.5)
	flue := Union(Capsule{base, mid, 1.05}, Capsule{mid, tip, 1.0})
	g.With(Stone, func() { g.Fill(flue) })
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		return flue.Contains(p) && int(math.Floor(p.Y/0.8))%2 == 0 && s.hash(g, p, seed) < 0.6
	}})
	g.With(Charcoal, func() {
		g.Fill(Cylinder{Center: tip.Plan(), Radius: 1.4, Y0: tip.Y - 0.1, Y1: tip.Y + 0.8})
	})
	drift := mushroomForestUnit(mushroomForestAt(1, -0.35))
	for i := 0; i < 4; i++ {
		f := float64(i)
		r := 0.95 + 0.42*f
		c := tip.Add(V(drift.X*(0.5*f*f+0.4*f), 1.6+1.75*f, drift.Y*(0.5*f*f+0.4*f)))
		col := White
		if i%2 == 1 {
			col = Stone
		}
		g.With(col, func() { g.FillVolume(Ellipsoid{c, V(r*1.15, r*0.9, r*1.15)}) })
	}
}

// warts studs a cap with raised white patches on a sunflower spiral.
func (s MushroomForest) warts(g *Grid, hat *mushroomForestCap, n int, size float64, seed int) {
	g.With(White, func() {
		for i := 0; i < n; i++ {
			r := hat.R * 0.95 * math.Sqrt((float64(i)+0.6)/float64(n))
			a := float64(i)*2.39996 + 0.5*Hash01(i, seed)
			rad := size * (0.75 + 0.5*Hash01(i, seed+1))
			g.FillVolume(Sphere(hat.world(r, a, hat.top(r)-0.2*rad), rad))
		}
	})
}

// greatAgaric is the tall fly agaric home at the back of the grove.
func (s MushroomForest) greatAgaric(g *Grid, m *mushroomForestMap) *mushroomForestCap {
	c := mushroomForestGreat
	y0 := m.on(c)
	face := mushroomForestFace
	st := mushroomForestStem{foot: c.At(y0 - 1.4), h: 40, r0: 6.0, r1: 4.8, bulb: 2.8,
		lean: mushroomForestAt(2, 1.2), wall: 1.2}
	hat := &mushroomForestCap{c: st.top(), R: 19, t: 1.2, dir: mushroomForestUnit(mushroomForestAt(-0.4, -1)),
		ang: 0.07, lo: -3, hi: 16.5}
	hat.top = func(r float64) float64 {
		return 15.5 * math.Pow(math.Max(0, 1-math.Pow(r/hat.R, 2.3)), 0.55)
	}
	hat.in = func(r, a, y float64) bool { return r <= hat.R && y >= 0 && y <= hat.top(r) }
	drop := func(r float64) float64 { return -1.1 - math.Sin(math.Pi*(r-4)/(hat.R-4)) }

	g.Use(White).Fill(st.solid())
	s.streaks(g, st, Stone, 201, 0.14)
	// Ridged rings around the swollen foot.
	for i, y := range []float64{y0 + 1.6, y0 + 3.4, y0 + 5.2} {
		ring := func(p Vec3) bool {
			d, _ := st.around(p)
			return math.Abs(p.Y-y) <= 0.5 && d > st.rad(y)-0.3 && d <= st.rad(y)+0.75-0.1*float64(i)
		}
		g.With(Sand, func() { g.Fill(Func{st.box().Grow(1), ring}) })
	}
	s.skirt(g, st, hat.c.Y-7.5, 4.2, 0.85, White, Sand)

	g.Use(White).Fill(hat.gills(28, 4.2, drop))
	g.Use(Red).Fill(hat.flesh())
	g.Paint(Terracotta, Func{hat.box(), func(p Vec3) bool {
		r, a, y := hat.local(p)
		return hat.in(r, a, y) && y < 1.0 && r > hat.R-2.2
	}})
	s.warts(g, hat, 46, 1.3, 210)
	s.lanterns(g, hat, 28, 7, 14, -1.6, 3.0, face)

	// The home: door, windows, balcony with its own door, and the spiral stair.
	s.door(g, st, face, y0+0.3, 2.3, Red, Wood)
	yb := y0 + 21.3
	s.stair(g, st, face+1.0, y0+0.3, yb, 8.8)
	s.balcony(g, st, face, 1.0, yb)
	s.door(g, st, face, yb, 1.6, Red, Wood)
	s.window(g, st, face+0.95, y0+13.5, 1.1)
	s.window(g, st, face-0.95, y0+5.6, 1.1)
	s.window(g, st, face-0.75, y0+15.5, 1.1)
	s.window(g, st, face+0.5, y0+28.5, 1.1)
	s.window(g, st, face-0.6, y0+29.5, 1.1)
	// Lantern beside the door.
	post := st.surface(y0, face-0.62, 2.0)
	g.With(Wood, func() { g.Fill(Cylinder{Center: post.Plan(), Radius: 0.45, Y0: y0 - 0.5, Y1: y0 + 3.4}) })
	g.With(Gold, func() { g.FillVolume(Sphere(post.Add(V(0, 4.0, 0)), 0.9)) })

	s.chimney(g, hat.on(13, face+1.9, -1.2), mushroomForestAt(-1.2, 0.4), 6, 215)
	return hat
}

// youngAgaric is a smaller, rounder fly agaric home on the left.
func (s MushroomForest) youngAgaric(g *Grid, m *mushroomForestMap) {
	c := mushroomForestYoung
	y0 := m.on(c)
	face := mushroomForestAt(1, 1).Angle()
	st := mushroomForestStem{foot: c.At(y0 - 1.2), h: 17.5, r0: 3.9, r1: 3.2, bulb: 1.9,
		lean: mushroomForestAt(-1.2, 0.6), wall: 1.1}
	hat := &mushroomForestCap{c: st.top(), R: 10.2, t: 1.25, dir: mushroomForestUnit(mushroomForestAt(-1, -0.6)),
		ang: 0.12, lo: -2.5, hi: 11.5}
	hat.top = func(r float64) float64 {
		return 10.5 * math.Pow(math.Max(0, 1-math.Pow(r/hat.R, 2.7)), 0.5)
	}
	hat.in = func(r, a, y float64) bool { return r <= hat.R && y >= 0 && y <= hat.top(r) }
	drop := func(r float64) float64 { return -0.9 - 0.6*math.Sin(math.Pi*(r-3)/(hat.R-3)) }

	g.Use(White).Fill(st.solid())
	s.streaks(g, st, Stone, 221, 0.14)
	for _, y := range []float64{y0 + 1.5, y0 + 3.2} {
		ring := func(p Vec3) bool {
			d, _ := st.around(p)
			return math.Abs(p.Y-y) <= 0.45 && d > st.rad(y)-0.3 && d <= st.rad(y)+0.7
		}
		g.With(Sand, func() { g.Fill(Func{st.box().Grow(1), ring}) })
	}
	s.skirt(g, st, hat.c.Y-3.6, 2.8, 0.8, White, Sand)
	g.Use(White).Fill(hat.gills(18, 2.8, drop))
	g.Use(Red).Fill(hat.flesh())
	g.Paint(Terracotta, Func{hat.box(), func(p Vec3) bool {
		r, a, y := hat.local(p)
		return hat.in(r, a, y) && y < 0.9 && r > hat.R-2
	}})
	s.warts(g, hat, 26, 1.15, 230)
	s.lanterns(g, hat, 18, 4, 7.2, -1.3, 2.2, face)

	s.door(g, st, face, y0+0.3, 1.9, Navy, Wood)
	s.window(g, st, face-1.2, y0+5.2, 1.0)
	s.window(g, st, face+0.9, y0+8.6, 1.0)
	s.window(g, st, face-0.3, y0+10.6, 1.0)
	s.chimney(g, hat.on(5.2, face+1.9, -1.0), mushroomForestAt(-0.9, -0.3), 5, 235)
}

// parasol is the tall slender parasol with a broad, flat, scaly cap and a knob at its center.
func (s MushroomForest) parasol(g *Grid, m *mushroomForestMap) *mushroomForestCap {
	c := mushroomForestTall
	y0 := m.on(c)
	st := mushroomForestStem{foot: c.At(y0 - 1), h: 43.5, r0: 2.5, r1: 2.0, bulb: 2.3, lean: mushroomForestAt(-2, 2.5)}
	hat := &mushroomForestCap{c: st.top(), R: 13.5, t: 1.3, dir: mushroomForestUnit(mushroomForestAt(-0.3, -1)),
		ang: 0.1, lo: -5, hi: 5.5}
	hat.top = func(r float64) float64 {
		k := r / hat.R
		return 1.9*(1-k*k) + 2.6*math.Exp(-(r/3.4)*(r/3.4))
	}
	hat.in = func(r, a, y float64) bool { return r <= hat.R && y <= hat.top(r) && y >= hat.top(r)-1.3 }
	drop := func(r float64) float64 {
		return hat.top(r) - 1.9 - 2.2*math.Pow(math.Max(0, math.Sin(math.Pi*(r-2)/(hat.R-2))), 0.7)
	}

	g.Use(Sand).Fill(st.solid())
	// Snakeskin bands on the stalk.
	g.Paint(Wood, Func{st.box(), func(p Vec3) bool {
		_, a := st.around(p)
		return p.Y < hat.c.Y-11 && math.Mod(p.Y+1.1*math.Sin(3*a)+100, 3.6) < 1.2
	}})
	// The loose double ring.
	ringY := hat.c.Y - 10
	ring := func(lo, hi, out float64) Solid {
		return Func{st.box().Grow(2), func(p Vec3) bool {
			d, _ := st.around(p)
			return p.Y >= ringY+lo && p.Y <= ringY+hi && d <= st.rad(p.Y)+out
		}}
	}
	g.With(White, func() { g.Fill(ring(0, 0.7, 1.9)) })
	g.With(Gold, func() { g.Fill(ring(-0.8, 0, 1.3)) })

	g.Use(White).Fill(hat.gills(22, 2, drop))
	g.Use(Sand).Fill(hat.flesh())
	// Concentric rings of scales around a dark knob.
	g.Paint(Gold, Func{hat.box(), func(p Vec3) bool {
		r, a, y := hat.local(p)
		if !hat.in(r, a, y) || y < hat.top(r)-0.7 {
			return false
		}
		if r < 3.3 {
			return true
		}
		band := int(math.Floor((r - 3.3) / 1.7))
		if band%2 == 0 || r > hat.R-1 {
			return false
		}
		sectors := float64(6 + 3*band)
		k := int(math.Floor((a + math.Pi) / (2 * math.Pi) * sectors))
		return Hash01(band, k, 241) < 0.62
	}})
	g.Paint(Wood, Func{hat.box(), func(p Vec3) bool {
		r, a, y := hat.local(p)
		return hat.in(r, a, y) && r < 1.5
	}})
	s.lanterns(g, hat, 22, 4, 9.5, -4.2, 2.4, 0.4)
	return hat
}

// drumstick is a young parasol whose egg-shaped cap has not opened yet.
func (s MushroomForest) drumstick(g *Grid, m *mushroomForestMap) {
	c := mushroomForestDrum
	y0 := m.on(c)
	st := mushroomForestStem{foot: c.At(y0 - 1), h: 23, r0: 2.1, r1: 1.8, bulb: 2.0, lean: mushroomForestAt(1.5, 1)}
	hat := &mushroomForestCap{c: st.top(), R: 5.2, t: 1.2, dir: P(1, 0), lo: -3.5, hi: 11.5}
	egg := func(r, y float64) float64 {
		x, v := r/hat.R, (y-4.2)/6.8
		return x*x + v*v
	}
	hat.top = func(r float64) float64 { return 4.2 + 6.8*math.Sqrt(math.Max(0, 1-r*r/(hat.R*hat.R))) }
	hat.in = func(r, a, y float64) bool { return egg(r, y) <= 1 }
	g.Use(Sand).Fill(st.solid())
	g.Paint(Wood, Func{st.box(), func(p Vec3) bool {
		_, a := st.around(p)
		return math.Mod(p.Y+1.1*math.Sin(3*a)+100, 3.6) < 1.2
	}})
	g.Use(Sand).Fill(hat.flesh())
	g.Paint(Gold, Func{hat.box(), func(p Vec3) bool {
		r, a, y := hat.local(p)
		if egg(r, y) > 1 {
			return false
		}
		if y > 8.4 {
			return true
		}
		band := int(math.Floor((y + 4) / 1.5))
		k := int(math.Floor((a+math.Pi)/(2*math.Pi)*9 + 0.5*float64(band%2)))
		return band%2 == 0 && y > 0.5 && Hash01(band, k, 251) < 0.6
	}})
	g.Paint(Wood, Func{hat.box(), func(p Vec3) bool {
		r, _, y := hat.local(p)
		return egg(r, y) <= 1 && y > 10.1
	}})
}

// bun returns a bolete: a thick cushion cap over a sponge of pores, on a fat stalk.
func (s MushroomForest) bun(g *Grid, st mushroomForestStem, R, H float64, dir Vec2, ang float64, crown Color, seed int) *mushroomForestCap {
	hat := &mushroomForestCap{c: st.top(), R: R, t: 1.3, dir: mushroomForestUnit(dir), ang: ang, lo: -4, hi: H + 1}
	hat.top = func(r float64) float64 {
		return H * math.Pow(math.Max(0, 1-math.Pow(r/R, 2.2)), 0.6)
	}
	hat.in = func(r, a, y float64) bool { return r <= R && y >= -0.8 && y <= hat.top(r) }
	// The sponge bulges below the rim.
	pores := func(r float64) float64 {
		k := r / (R - 0.7)
		return -0.8 - 0.22*R*math.Sqrt(math.Max(0, 1-k*k))
	}
	g.With(Gold, func() {
		g.Fill(Func{hat.box(), func(p Vec3) bool {
			r, _, y := hat.local(p)
			return r <= R-0.7 && r >= st.r1-1 && y >= pores(r) && y <= pores(r)+1.3
		}})
	})
	g.Paint(Sand, Func{hat.box(), func(p Vec3) bool {
		r, a, y := hat.local(p)
		return y < -0.2 && r < R && mushroomForestNoise(r*0.5, a*3, seed) > 0.25
	}})
	g.Use(Terracotta).Fill(hat.flesh())
	g.Paint(crown, Func{hat.box(), func(p Vec3) bool {
		r, a, y := hat.local(p)
		if !hat.in(r, a, y) || y < 0 {
			return false
		}
		n := mushroomForestFBM(math.Cos(a)*r*0.2+7, math.Sin(a)*r*0.2, seed+1)
		return r < 0.36*R+0.25*R*n || n > 0.42
	}})
	g.Paint(Sand, Func{hat.box(), func(p Vec3) bool {
		r, a, y := hat.local(p)
		return hat.in(r, a, y) && y < 0.1 && r > R-1.6
	}})
	return hat
}

// bolete is the fat terracotta bolete cottage in the foreground.
func (s MushroomForest) bolete(g *Grid, m *mushroomForestMap) {
	c := mushroomForestBun
	y0 := m.on(c)
	face := mushroomForestFace
	st := mushroomForestStem{foot: c.At(y0 - 1.2), h: 13.4, r0: 6.5, r1: 5.6, bulb: 1.4,
		lean: mushroomForestAt(0.6, 0.4), wall: 1.2}
	g.Use(Sand).Fill(st.solid())
	s.streaks(g, st, White, 261, 0.3)
	g.Paint(Terracotta, Func{st.box(), func(p Vec3) bool {
		return p.Y > st.foot.Y+st.h-2.4 && s.hash(g, p, 263) < 0.55
	}})
	hat := s.bun(g, st, 11.5, 8.8, mushroomForestAt(0.5, -1), 0.06, Brick, 265)
	s.door(g, st, face, y0+0.3, 2.2, Foliage, Wood)
	s.window(g, st, face+1.05, y0+5.0, 1.15)
	s.window(g, st, face-1.05, y0+5.0, 1.15)
	s.window(g, st, face+2.1, y0+6.0, 1.0)
	s.window(g, st, face-2.1, y0+6.0, 1.0)
	// Porch lantern on a bracket over the door.
	arm := st.surface(y0+7.6, face, -0.3)
	out := st.surface(y0+7.6, face, 2.0)
	g.With(Wood, func() { g.Fill(Capsule{arm, out, 0.45}) })
	s.lantern(g, out, 0.6)
	s.chimney(g, hat.on(6.4, mushroomForestAt(1, -0.2).Angle(), -1.2), mushroomForestAt(1.1, 0.3), 6, 267)
}

// pup is the small bolete beside the cottage.
func (s MushroomForest) pup(g *Grid, m *mushroomForestMap) {
	c := mushroomForestPup
	y0 := m.on(c)
	st := mushroomForestStem{foot: c.At(y0 - 1), h: 6.6, r0: 3.4, r1: 2.9, bulb: 0.8,
		lean: mushroomForestAt(0.8, 0.3), wall: 1.1}
	g.Use(Sand).Fill(st.solid())
	s.streaks(g, st, White, 271, 0.3)
	s.bun(g, st, 6.2, 5.0, mushroomForestAt(1, 0.3), 0.14, Brick, 275)
}

// bells is the cluster of pale blue bell caps on thin leaning stalks.
func (s MushroomForest) bells(g *Grid, m *mushroomForestMap) {
	for i, b := range [][4]float64{{-6, -2, 19, 4.7}, {0, -3.5, 14, 4.1}, {5.5, 0, 10.5, 3.7}, {-1.5, 3, 7.6, 3.2}, {4, 5, 5.2, 2.7}} {
		s.part(g, func(t *Grid) {
			c := mushroomForestBells.Add(mushroomForestAt(b[0], b[1]))
			y0 := m.on(c)
			H, R := b[2], b[3]
			out := mushroomForestUnit(mushroomForestAt(b[0]+0.3, b[1]+1))
			st := mushroomForestStem{foot: c.At(y0 - 0.8), h: H + 0.8, r0: 1.05, r1: 0.9, bulb: 0.5,
				lean: out.Scale(0.14 * H)}
			t.Use(White).Fill(st.solid())
			t.Paint(Glass, Func{st.box(), func(p Vec3) bool { return p.Y < y0+0.25*H }})
			bh := 1.35 * R
			hat := &mushroomForestCap{c: st.top().Add(V(0, -0.45*bh, 0)), R: R, t: 1.1, dir: out, ang: -0.16,
				lo: -1.5, hi: bh + 0.5}
			// A bell: a round crown, straight flanks and a flared lip.
			wide := func(y float64) float64 {
				x := (bh - y) / bh
				if x < 0.38 {
					k := 1 - x/0.38
					return 0.64 * R * math.Sqrt(1-k*k)
				}
				k := (x - 0.38) / 0.62
				return R * (0.64 + 0.1*k + 0.26*k*k*k)
			}
			hat.top = func(r float64) float64 { return bh }
			hat.in = func(r, a, y float64) bool { return y >= 0 && y <= bh && r <= wide(y) }
			t.Use(White).Fill(Func{hat.box(), func(p Vec3) bool {
				r, a, y := hat.local(p)
				k := math.Round(a / (math.Pi / 4))
				return y >= -0.5 && y < 0.5*bh && r >= 1 && r <= wide(math.Max(y, 0))-0.3 &&
					r*math.Abs(math.Sin(a-k*math.Pi/4)) <= 0.5
			}})
			t.Use(Glass).Fill(hat.flesh())
			stripes := float64(8 + 2*(i%2))
			t.Paint(Navy, Func{hat.box(), func(p Vec3) bool {
				r, a, y := hat.local(p)
				if !hat.in(r, a, y) {
					return false
				}
				if y > 0.84*bh {
					return true
				}
				step := 2 * math.Pi / stripes
				return y > 1.0 && y < 0.6*bh && r*math.Abs(math.Sin(a-math.Round(a/step)*step)) <= 0.5
			}})
			t.Paint(White, Func{hat.box(), func(p Vec3) bool {
				r, a, y := hat.local(p)
				return hat.in(r, a, y) && y < 0.75
			}})
		})
	}
}

// horns is the clump of tall golden funnel caps at the back, leaning toward the camera so
// their hollows show, with gills running down the outside of each trumpet.
func (s MushroomForest) horns(g *Grid, m *mushroomForestMap) {
	for i, b := range [][4]float64{{-2, -2, 19, 7.4}, {6.5, 0, 13, 5.8}, {2, 6, 8, 4.2}} {
		s.part(g, func(t *Grid) {
			c := mushroomForestHorns.Add(mushroomForestAt(b[0], b[1]))
			y0 := m.on(c) - 0.8
			H, R := b[2]+0.8, b[3]
			rs := 0.16*R + 0.9
			phase := 1.7 * float64(i)
			step := 2 * math.Pi / math.Round(1.4*R)
			hat := &mushroomForestCap{c: c.At(y0), R: R*1.1 + 0.17*H, t: 1.2, dir: mushroomForestUnit(mushroomForestAt(0.3, 1)),
				ang: -0.16, lo: -1, hi: H + 1.5}
			rim := func(a float64) (float64, float64) {
				return R * (1 + 0.08*math.Sin(4*a+phase)), H + 0.45*math.Sin(4*a+phase+1.3)
			}
			// The trumpet flares to a rim with a blunt edge a unit tall.
			wide := func(a, y float64) float64 {
				ra, ha := rim(a)
				return rs + (ra-rs)*math.Pow(math.Max(0, math.Min(1, y/(ha-1.1))), 2.6)
			}
			lip := func(r, a float64) float64 {
				ra, ha := rim(a)
				k := math.Min(1, r/ra)
				return ha - 0.32*H*(1-k*k)
			}
			hat.top = func(r float64) float64 { return H }
			hat.in = func(r, a, y float64) bool { return y >= 0 && y <= lip(r, a) && r <= wide(a, y) }
			t.Use(Gold).Fill(hat.flesh())
			t.Paint(Terracotta, Func{hat.box(), func(p Vec3) bool {
				r, a, y := hat.local(p)
				_, ha := rim(a)
				if y > lip(r, a)-0.9 && r < 0.3*R {
					return true // The throat.
				}
				if y < 0.52*H || y > ha-1.3 || r < wide(a, y)-1.0 {
					return false
				}
				return r*math.Abs(math.Sin(a-math.Round(a/step)*step)) <= 0.5
			}})
			t.Paint(Sand, Func{hat.box(), func(p Vec3) bool {
				r, a, y := hat.local(p)
				ra, ha := rim(a)
				return y < 1.6 || y > ha-1.0 && r > ra-1.5
			}})
		})
	}
}

// ropeBridge is a sagging plank bridge with rope handrails between two caps.
func (s MushroomForest) ropeBridge(g *Grid, p1, p2 Vec3) {
	d := p2.Plan().Sub(p1.Plan())
	L := d.Length()
	u := d.Scale(1 / L)
	n := P(-u.Y, u.X)
	sag := 0.07*L + 0.5
	deck := func(t float64) float64 { return p1.Y + (p2.Y-p1.Y)*t - sag*4*t*(1-t) }
	rope := func(t float64) float64 { return deck(t) + 2.0 - 0.35*4*t*(1-t) }
	box := BoxOf(p1, p2).Grow(3)
	box.Min.Y -= sag + 1
	local := func(p Vec3) (float64, float64) {
		r := p.Plan().Sub(p1.Plan())
		return r.Dot(u), r.Dot(n)
	}
	g.Use(Wood).Fill(Func{box, func(p Vec3) bool {
		sl, a := local(p)
		if sl < -0.6 || sl > L+0.6 || math.Abs(a) > 1.1 {
			return false
		}
		y := deck(math.Max(0, math.Min(1, sl/L)))
		return p.Y <= y && p.Y >= y-0.75
	}})
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		sl, _ := local(p)
		return math.Mod(sl+10, 1.6) < 0.8
	}})
	hangers := math.Round(L / 2.4)
	g.Use(Sand).Fill(Func{box, func(p Vec3) bool {
		sl, a := local(p)
		if sl < 0 || sl > L {
			return false
		}
		t := sl / L
		for _, side := range []float64{-1.3, 1.3} {
			if math.Hypot(a-side, p.Y-rope(t)) <= 0.5 {
				return true
			}
			f := t * hangers
			if math.Abs(a-side) < 0.42 && p.Y > deck(t) && p.Y < rope(t) && math.Abs(f-math.Round(f))*L/hangers < 0.4 {
				return true
			}
		}
		return false
	}})
	for _, e := range []Vec3{p1, p2} {
		for _, side := range []float64{-1.3, 1.3} {
			q := e.Plan().Add(n.Scale(side))
			g.With(Wood, func() { g.Fill(Cylinder{Center: q, Radius: 0.5, Y0: e.Y - 1.6, Y1: e.Y + 2.3}) })
			g.With(Gold, func() { g.FillVolume(Sphere(q.At(e.Y+2.9), 0.8)) })
		}
	}
}

// bracket is a shelf fungus: a flat half-disc growing out of a trunk at a point.
func (s MushroomForest) bracket(g *Grid, at Vec3, out Vec2, w, reach float64, seed int) {
	tn := P(-out.Y, out.X)
	k := w + reach
	box := Box{V(at.X-k, at.Y-0.7, at.Z-k), V(at.X+k, at.Y+0.5, at.Z+k)}
	zone := func(p Vec3) (float64, float64) {
		q := p.Plan().Sub(at.Plan())
		o := q.Dot(out)
		return o, math.Hypot(o/reach, q.Dot(tn)/w)
	}
	g.With(White, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			o, z := zone(p)
			return o > -1.3 && z <= 1 && p.Y <= at.Y+0.4 && p.Y >= at.Y-0.55
		}})
	})
	for i, c := range []Color{Sand, Terracotta, Brick} {
		lim := []float64{0.84, 0.62, 0.32}[i] + 0.08*(Hash01(seed, i)-0.5)
		g.Paint(c, Func{box, func(p Vec3) bool {
			o, z := zone(p)
			return o > -1.3 && z <= lim && p.Y > at.Y-0.15
		}})
	}
}

// ladybird is a spotted red beetle standing at a point.
func (s MushroomForest) ladybird(g *Grid, at Vec3, heading Vec2, k, slope float64) {
	fr := mushroomForestFrame{at, mushroomForestUnit(heading), slope}
	s.part(g, func(t *Grid) {
		t.With(Charcoal, func() {
			for _, x := range []float64{-1.0, 0.2, 1.3} {
				for _, z := range []float64{-1, 1} {
					t.Fill(Capsule{fr.w(x*k, 0.6*k, z*1.2*k), fr.w(x*k*1.2, -0.3, z*2.2*k), 0.45})
				}
			}
			t.FillVolume(fr.ell(V(1.9*k, 0.7*k, 0), V(1.0*k, 0.8*k, 1.2*k)))
		})
		t.With(Red, func() {
			body := fr.ell(V(0, 0.25*k, 0), V(2.2*k, 1.6*k, 1.9*k))
			t.FillVolume(Intersect(body, Func{body.Bounds(), func(p Vec3) bool { return fr.l(p).Y >= -0.2 }}))
		})
		for _, q := range [][3]float64{{0.9, 1.5, 0.9}, {0.9, 1.5, -0.9}, {-0.7, 1.6, 1.0}, {-0.7, 1.6, -1.0}, {-1.8, 0.9, 0}, {0.1, 1.9, 0}} {
			t.Paint(Charcoal, Sphere(fr.w(q[0]*k, q[1]*k, q[2]*k), 0.62*k))
		}
		t.Paint(White, Sphere(fr.w(2.3*k, 1.2*k, 0.75*k), 0.5), Sphere(fr.w(2.3*k, 1.2*k, -0.75*k), 0.5))
	})
}

// deadwood is the hollow stump, the fallen log leaning on it and their bracket fungi.
func (s MushroomForest) deadwood(g *Grid, m *mushroomForestMap) {
	c := mushroomForestStump
	end := mushroomForestLogEnd
	y0 := m.on(c)
	hinge := end.Sub(c).Angle()
	const R0, R1, H = 6.4, 5.3, 12.0
	rad := func(a, y float64) float64 {
		t := math.Max(0, math.Min(1, (y-y0)/H))
		root := math.Cos(2.5*a + 0.7)
		return R0 + (R1-R0)*t + 3.2*root*root*math.Exp(-math.Max(0, y-y0)/2.0) + 0.3*math.Sin(7*a+1)
	}
	crest := func(a float64) float64 {
		da := math.Remainder(a-hinge, 2*math.Pi)
		return y0 + H + 2.2*mushroomForestNoise(math.Cos(a)*2.2+5, math.Sin(a)*2.2, 51) + 3*math.Exp(-da*da/0.2)
	}
	box := Box{V(c.X-10.5, y0-1.2, c.Y-10.5), V(c.X+10.5, y0+H+6, c.Y+10.5)}
	polar := func(p Vec3) (float64, float64) {
		q := p.Plan().Sub(c)
		return q.Length(), q.Angle()
	}
	s.part(g, func(t *Grid) {
		t.Use(Wood).Fill(Func{box, func(p Vec3) bool {
			d, a := polar(p)
			r := rad(a, p.Y)
			return p.Y >= y0-1.2 && p.Y <= crest(a) && d <= r && d > r-1.3
		}})
		t.Paint(DarkStone, Func{box, func(p Vec3) bool {
			_, a := polar(p)
			sector := int(math.Floor((a + math.Pi) / (2 * math.Pi) * 30))
			band := int(math.Floor(p.Y/4.2 + 3*Hash01(sector, 52)))
			return Hash01(sector, band, 53) < 0.3
		}})
		t.Paint(Charcoal, Func{box, func(p Vec3) bool {
			d, a := polar(p)
			return d < rad(a, p.Y)-0.7 && p.Y > y0+3
		}})
		t.Paint(Sand, Func{box, func(p Vec3) bool {
			_, a := polar(p)
			return p.Y > crest(a)-0.8
		}})
		t.Paint(Foliage, Func{box, func(p Vec3) bool {
			d, a := polar(p)
			if d < rad(a, p.Y)-0.7 {
				return false
			}
			n := mushroomForestFBM(math.Cos(a)*2.4+p.Y*0.12, math.Sin(a)*2.4, 54)
			return p.Y < y0+2.6+4.5*n || n > 0.34
		}})
		t.Paint(Grass, Func{box, func(p Vec3) bool {
			_, a := polar(p)
			return p.Y < y0+2.2 && mushroomForestNoise(math.Cos(a)*3+p.Y*0.3, math.Sin(a)*3, 55) > 0.2
		}})
	})

	// The log, from the stump's broken top down to the ground.
	a3 := c.Add(mushroomForestUnit(end.Sub(c)).Scale(2.5)).At(y0 + H + 1.6)
	b3 := end.At(m.on(end) + 3.3)
	ab := b3.Sub(a3)
	L := ab.Length()
	ax := ab.Scale(1 / L)
	hz := math.Hypot(ax.X, ax.Z)
	side := V(-ax.Z/hz, 0, ax.X/hz) // Horizontal, toward the camera.
	up := V(-side.Z*ax.Y, side.Z*ax.X-side.X*ax.Z, side.X*ax.Y)
	loc := func(p Vec3) (sl, d, phi float64) {
		q := p.Sub(a3)
		sl = q.Dot(ax)
		r := q.Sub(ax.Scale(sl))
		return sl, r.Length(), math.Atan2(r.Dot(up), r.Dot(side))
	}
	lr := func(sl float64) float64 { return 4.4 - 0.7*sl/L }
	lbox := BoxOf(a3, b3).Grow(6.5)
	point := func(sl float64, phi, off float64) Vec3 {
		r := lr(sl) + off
		return a3.Add(ax.Scale(sl)).Add(side.Scale(r * math.Cos(phi))).Add(up.Scale(r * math.Sin(phi)))
	}
	s.part(g, func(t *Grid) {
		t.Use(Wood).Fill(Func{lbox, func(p Vec3) bool {
			sl, d, _ := loc(p)
			return sl >= -1.5 && sl <= L && d <= lr(sl) && d > lr(sl)-1.2
		}})
		// Broken branch stubs.
		for _, b := range [][3]float64{{8, 1.3, 3.2}, {16.5, 2.2, 2.8}} {
			t.Fill(Capsule{point(b[0], b[1], -0.5), point(b[0]+1, b[1], b[2]), 1.25})
			t.With(Sand, func() { t.FillVolume(Sphere(point(b[0]+1, b[1], b[2]+0.5), 1.0)) })
		}
		t.Paint(DarkStone, Func{lbox, func(p Vec3) bool {
			sl, _, phi := loc(p)
			sector := int(math.Floor((phi + math.Pi) / (2 * math.Pi) * 22))
			band := int(math.Floor(sl/5 + 3*Hash01(sector, 56)))
			return sl < L && Hash01(sector, band, 57) < 0.3
		}})
		t.Paint(Charcoal, Func{lbox, func(p Vec3) bool {
			sl, d, _ := loc(p)
			return d < lr(sl)-0.65 && sl < L-0.4
		}})
		t.Paint(Sand, Func{lbox, func(p Vec3) bool {
			sl, d, _ := loc(p)
			return sl > L-0.8 && d <= lr(sl)+0.2
		}})
		t.Paint(Foliage, Func{lbox, func(p Vec3) bool {
			sl, d, phi := loc(p)
			if d < lr(sl)-0.65 || d > lr(sl)+0.3 || sl > L-0.8 {
				return false
			}
			return math.Sin(phi) > 0.3+0.9*mushroomForestFBM(sl*0.16, phi*1.3, 58)
		}})
		t.Paint(Grass, Func{lbox, func(p Vec3) bool {
			sl, d, phi := loc(p)
			if d < lr(sl)-0.65 || d > lr(sl)+0.3 || sl > L-0.8 {
				return false
			}
			return math.Sin(phi) > 0.72 && mushroomForestNoise(sl*0.35, phi*2, 59) > 0.1
		}})
	})
	// Splinters where the trunk snapped.
	g.With(Sand, func() {
		for i, da := range []float64{-0.5, 0.1, 0.55} {
			a := hinge + da
			foot := Polar(c, R1-0.6, a).At(crest(a) - 1)
			g.Fill(Capsule{foot, foot.Add(V(0, 2.6+1.2*Hash01(i, 60), 0)), 0.55})
		}
	})
	// Bracket fungi: a stair winding up the stump, then stepping along the log.
	for i := 0; i < 6; i++ {
		a := mushroomForestFace + 1.9 - 0.48*float64(i)
		y := y0 + 2.0 + 1.9*float64(i)
		s.bracket(g, Polar(c, rad(a, y)-0.5, a).At(y), P(math.Cos(a), math.Sin(a)), 3.2-0.15*float64(i%2), 3.3, 61+i)
	}
	out := mushroomForestUnit(side.Plan())
	for i := 0; i < 4; i++ {
		sl := 5.5 + 4.6*float64(i)
		s.bracket(g, point(sl, 0.12, -0.5), out, 2.9+0.5*Hash01(i, 68), 3.0+0.6*Hash01(i, 69), 70+i)
	}
	s.ladybird(g, point(11.5, math.Pi/2, 0.2), ax.Plan(), 1.15, ax.Y/hz)
}

// crossings are the arched stone bridge and the stepping stones over the brook.
func (s MushroomForest) crossings(g *Grid, m *mushroomForestMap) {
	c := mushroomForestArch
	dir := mushroomForestUnit(mushroomForestAt(0.832, -0.555))
	side := P(-dir.Y, dir.X)
	const half, width, span = 5.6, 1.9, 3.5
	y0 := math.Min(m.on(c.Add(dir.Scale(half))), m.on(c.Sub(dir.Scale(half)))) - 0.1
	deck := func(sl float64) float64 { return y0 + 0.5 + 2.3*(1-(sl/half)*(sl/half)) }
	box := BoxOf(c.Add(dir.Scale(half)).At(0), c.Sub(dir.Scale(half)).At(y0+5)).Grow(3)
	box.Min.Y, box.Max.Y = 0.3, y0+5
	local := func(p Vec3) (float64, float64) {
		q := p.Plan().Sub(c)
		return q.Dot(dir), q.Dot(side)
	}
	soffit := func(sl float64) float64 {
		if math.Abs(sl) >= span {
			return -1
		}
		return mushroomForestWater + 0.2 + 2.7*math.Sqrt(1-(sl/span)*(sl/span))
	}
	s.part(g, func(t *Grid) {
		t.Use(Stone).Fill(Func{box, func(p Vec3) bool {
			sl, v := local(p)
			if math.Abs(sl) > half+0.6 || math.Abs(v) > width {
				return false
			}
			d := deck(math.Max(-half, math.Min(half, sl)))
			if math.Abs(v) > width-0.75 && p.Y > d && p.Y <= d+1.2 {
				return true
			}
			if p.Y > d || p.Y < soffit(sl) {
				return false
			}
			return math.Abs(sl) < span || p.Y > m.at(p.X, p.Z)-0.6
		}})
		t.Paint(DarkStone, Func{box, func(p Vec3) bool {
			sl, _ := local(p)
			ring := math.Abs(sl) < span+0.9 && p.Y < soffit(math.Max(-span+0.01, math.Min(span-0.01, sl)))+0.9
			return ring && int(math.Floor((sl+20)/1.1))%2 == 0 || s.hash(t, p, 81) < 0.14
		}})
		t.Paint(Sand, Func{box, func(p Vec3) bool {
			sl, v := local(p)
			d := deck(math.Max(-half, math.Min(half, sl)))
			return math.Abs(v) < width-0.75 && p.Y > d-0.45 && s.hash(t, p, 82) < 0.75
		}})
	})
	for _, e := range []float64{-half, half} {
		q := c.Add(dir.Scale(e)).Add(side.Scale(width - 0.4))
		y := m.on(q)
		g.With(Wood, func() { g.Fill(Cylinder{Center: q, Radius: 0.5, Y0: y - 0.5, Y1: y + 3.2}) })
		g.With(Gold, func() { g.FillVolume(Sphere(q.At(y+3.8), 0.9)) })
	}
	// Stepping stones.
	across := mushroomForestUnit(mushroomForestAt(0.874, -0.486))
	for k := -2; k <= 2; k++ {
		q := mushroomForestStones.Add(across.Scale(1.9 * float64(k)))
		col := Stone
		if k%2 != 0 {
			col = DarkStone
		}
		g.With(col, func() {
			g.FillVolume(EllipticCylinder{q, 1.0 + 0.2*Hash01(k, 83), 0.9, 0.3, mushroomForestWater + 0.55})
		})
	}
}

// frog sits at a point, facing a heading.
func (s MushroomForest) frog(g *Grid, at Vec3, heading Vec2) {
	fr := mushroomForestFrame{at, mushroomForestUnit(heading), 0}
	s.part(g, func(t *Grid) {
		t.With(Foliage, func() {
			for _, z := range []float64{-1.7, 1.7} {
				t.FillVolume(fr.ell(V(-1.1, 0.8, z), V(1.8, 0.85, 0.95)))
				t.Fill(Capsule{fr.w(1.2, 1.2, z*0.8), fr.w(2.1, 0.2, z*1.1), 0.5})
			}
		})
		t.With(Grass, func() {
			t.FillVolume(fr.ell(V(0, 1.5, 0), V(2.4, 1.5, 1.9)))
			t.FillVolume(fr.ell(V(1.9, 2.3, 0), V(1.5, 1.2, 1.6)))
		})
		t.Paint(White, Func{Box{at.Sub(V(5, 1, 5)), at.Add(V(5, 2.0, 5))}, func(p Vec3) bool {
			q := fr.l(p)
			return q.X > 0.4 && q.Y < 1.9 && math.Abs(q.Z) < 1.3
		}})
		for _, z := range []float64{-1.0, 1.0} {
			t.With(Gold, func() { t.FillVolume(Sphere(fr.w(2.0, 3.4, z), 0.7)) })
			t.Paint(Charcoal, Sphere(fr.w(2.5, 3.5, z*1.3), 0.42))
		}
	})
}

// pond dresses the pond: lily pads, the frog, reeds and the jetty.
func (s MushroomForest) pond(g *Grid, m *mushroomForestMap) {
	c := mushroomForestPond
	w := mushroomForestWater
	pads := [][3]float64{{-4.4, 1.6, 2.7}, {0.8, -4.4, 1.9}, {3.9, 4.2, 1.8}, {-1.3, 6.4, 1.6}, {-6, -3.9, 1.7}, {5, -5.2, 1.4}}
	for i, pd := range pads {
		q := c.Add(mushroomForestAt(pd[0], pd[1]))
		col := Grass
		if i%2 == 1 {
			col = Foliage
		}
		notch := Hash01(i, 91) * 2 * math.Pi
		g.With(col, func() {
			g.Fill(Func{Box{V(q.X-3, w-0.2, q.Y-3), V(q.X+3, w+0.45, q.Y+3)}, func(p Vec3) bool {
				d := p.Plan().Sub(q)
				if d.Length() > pd[2] || !s.wet(m, p.X, p.Z) {
					return false
				}
				return d.Length() < 0.5 || math.Abs(math.Remainder(d.Angle()-notch, 2*math.Pi)) > 0.3
			}})
		})
		if i > 0 && i%2 == 0 || i == 1 {
			f := Polar(q, 0.3*pd[2], notch+2.5)
			g.With(White, func() { g.FillVolume(Ellipsoid{f.At(w + 0.85), V(0.95, 0.6, 0.95)}) })
			g.With(Gold, func() { g.Fill(Sphere(f.At(w+1.35), 0.45)) })
		}
	}
	first := c.Add(mushroomForestAt(pads[0][0], pads[0][1]))
	s.frog(g, first.At(w+0.35), mushroomForestAt(1, 0.5))

	// Jetty from the east bank, toward the middle of the pond.
	dir := mushroomForestUnit(mushroomForestAt(-1, 0.1))
	side := P(-dir.Y, dir.X)
	start := c.Add(mushroomForestAt(12, -1.5))
	deck := m.on(start) + 0.1
	const length = 8.5
	g.With(Wood, func() {
		g.Fill(Func{BoxOf(start.At(deck-0.6), start.Add(dir.Scale(length)).At(deck)).Grow(2.5), func(p Vec3) bool {
			q := p.Plan().Sub(start)
			u, v := q.Dot(dir), q.Dot(side)
			return u >= -1.5 && u <= length && math.Abs(v) <= 1.3 && p.Y >= deck-0.6 && p.Y <= deck
		}})
		for _, u := range []float64{3.8, length - 0.4} {
			for _, v := range []float64{-1.3, 1.3} {
				q := start.Add(dir.Scale(u)).Add(side.Scale(v))
				top := deck + 0.3
				if u > 5 && v > 0 {
					top = deck + 3.0
				}
				g.Fill(Cylinder{Center: q, Radius: 0.5, Y0: 0.2, Y1: top})
			}
		}
	})
	g.Paint(Sand, Func{BoxOf(start.At(deck-0.6), start.Add(dir.Scale(length)).At(deck)).Grow(2.5), func(p Vec3) bool {
		q := p.Plan().Sub(start)
		u, v := q.Dot(dir), q.Dot(side)
		return u >= -1.5 && u <= length && math.Abs(v) <= 1.3 && p.Y >= deck-0.6 && p.Y <= deck && math.Mod(u+10, 1.8) < 0.9
	}})
	tip := start.Add(dir.Scale(length - 0.4)).Add(side.Scale(1.3))
	g.With(Gold, func() { g.FillVolume(Sphere(tip.At(deck+3.6), 0.9)) })

	// Reeds and cattails in clumps on the banks.
	for i := 0; i < 9; i++ {
		a := mushroomForestFace + []float64{1.4, 2.2, -2.4}[i%3] + 0.17*float64(i/3) + 0.1*Hash01(i, 92)
		q := c.Add(Polar(P(0, 0), 20, a))
		// Walk in to the water's edge.
		for k := 0; k < 60 && s.pondD(q) > -0.3; k++ {
			q = q.Add(mushroomForestUnit(c.Sub(q)).Scale(0.3))
		}
		h := 3.2 + 2.4*Hash01(i, 93)
		tipR := q.Add(Polar(P(0, 0), 0.5, a+1.5)).At(m.on(q) + h)
		g.With(Foliage, func() { g.Fill(Capsule{q.At(m.on(q) - 0.3), tipR, 0.45}) })
		if i%2 == 0 {
			g.With(Wood, func() { g.Fill(Capsule{tipR.Add(V(0, -0.2, 0)), tipR.Add(V(0, 1.3, 0)), 0.62}) })
		}
	}
}

// snail is the giant snail on the lawn in front of the cottage.
func (s MushroomForest) snail(g *Grid, m *mushroomForestMap) {
	c := mushroomForestSnail
	fr := mushroomForestFrame{c.At(m.on(c) - 0.3), mushroomForestUnit(mushroomForestAt(-1, 0.12)), 0}
	g.With(Stone, func() {
		g.FillVolume(fr.ell(V(0, 1.0, 0), V(7.0, 1.35, 2.2)))
		g.FillVolume(Capsule{fr.w(3.8, 1.4, 0), fr.w(6.0, 3.6, 0), 1.3})
		g.FillVolume(Sphere(fr.w(6.4, 4.0, 0), 1.45))
		for _, z := range []float64{-0.8, 0.8} {
			g.Fill(Capsule{fr.w(6.6, 4.8, z), fr.w(7.6, 7.0, z*1.7), 0.5})
			g.Fill(Capsule{fr.w(7.3, 3.6, z*0.7), fr.w(8.3, 3.2, z*1.2), 0.45})
		}
	})
	g.Paint(DarkStone, Func{Box{fr.o.Sub(V(9, 1, 9)), fr.o.Add(V(9, 1.1, 9))}, func(p Vec3) bool {
		return math.Mod(fr.l(p).X+20, 2.4) < 0.9
	}})
	for _, z := range []float64{-0.8, 0.8} {
		g.With(Charcoal, func() { g.FillVolume(Sphere(fr.w(7.7, 7.3, z*1.75), 0.72)) })
	}
	// The shell: a flattened ball coiled in a spiral.
	sc, sr := V(-0.9, 5.4, 0), V(4.7, 4.7, 2.7)
	body := fr.ell(sc, sr)
	g.Use(Terracotta)
	s.shell(g, body.Bounds(), 1.1, true, body.Contains)
	coil := func(p Vec3) (float64, float64) {
		q := fr.l(p).Sub(sc)
		rho := math.Hypot(q.X, q.Y)
		turn := rho/2.5 - math.Atan2(q.Y, -q.X)/(2*math.Pi)
		return rho, turn - math.Floor(turn)
	}
	g.Paint(Sand, Func{body.Bounds(), func(p Vec3) bool {
		rho, f := coil(p)
		return body.Contains(p) && rho > 0.9 && f < 0.36
	}})
	g.Paint(Brick, Func{body.Bounds(), func(p Vec3) bool {
		rho, f := coil(p)
		return body.Contains(p) && (rho <= 0.9 || f > 0.7 && rho > 2)
	}})
}

// floor scatters ferns, clover, toadstool rings, tufts and mossy stones.
func (s MushroomForest) floor(g *Grid, m *mushroomForestMap) {
	logA, logB := mushroomForestStump, mushroomForestLogEnd
	free := func(q Vec2, clear float64) bool {
		if mushroomForestE(q.X, q.Y) > 0.94 || s.streamD(q) < 3.2+0.3*clear || s.pondD(q) < 1.6+0.3*clear ||
			s.pathD(q) < 1.7+0.5*clear {
			return false
		}
		if d, _ := mushroomForestPolyline(q, []Vec2{logA, logB}); d < 6+clear {
			return false
		}
		for _, k := range mushroomForestKeep {
			if q.Sub(k.c).Length() < k.r+clear {
				return false
			}
		}
		return true
	}
	toadstool := func(q Vec2, size float64, top Color) {
		h := m.on(q)
		g.With(White, func() {
			g.Fill(Cylinder{Center: q, Radius: 0.42 + 0.14*size, Y0: h - 0.5, Y1: h + 1.5*size})
		})
		hat := Ellipsoid{q.At(h + 1.5*size), V(1.0*size+0.35, 0.7*size+0.25, 1.0*size+0.35)}
		g.With(top, func() {
			g.FillVolume(Intersect(hat, Cuboid{V(q.X-9, h+1.5*size-0.2, q.Y-9), V(q.X+9, h+20, q.Y+9)}))
		})
	}
	fern := func(q Vec2, size float64, seed int) {
		h := m.on(q)
		col := Foliage
		if Hash01(seed, 1) < 0.45 {
			col = Grass
		}
		g.With(col, func() {
			for k := 0; k < 6; k++ {
				a := float64(k)*math.Pi/3 + Hash01(seed, 2)
				mid := Polar(q, 1.2*size, a).At(h + 1.5*size)
				g.Fill(Capsule{q.At(h - 0.2), mid, 0.45}, Capsule{mid, Polar(q, 2.5*size, a+0.2).At(h + 0.7*size), 0.42})
			}
		})
	}
	clover := func(q Vec2, seed int) {
		for k := 0; k < 3; k++ {
			c := Polar(q, 1.5, float64(k)*2.1+Hash01(seed, 3)*3)
			h := m.on(c)
			top := h + 1.2 + 0.5*Hash01(seed, k, 4)
			g.With(Foliage, func() { g.Fill(Cylinder{Center: c, Radius: 0.42, Y0: h - 0.3, Y1: top}) })
			g.With(Grass, func() {
				for j := 0; j < 3; j++ {
					leaf := Polar(c, 0.85, float64(j)*2*math.Pi/3+Hash01(seed, k, 5))
					g.FillVolume(Ellipsoid{leaf.At(top), V(0.95, 0.42, 0.95)})
				}
			})
		}
		if Hash01(seed, 6) < 0.6 {
			h := m.on(q)
			g.With(Foliage, func() { g.Fill(Cylinder{Center: q, Radius: 0.42, Y0: h - 0.3, Y1: h + 2.2}) })
			g.With(White, func() { g.FillVolume(Sphere(q.At(h+2.6), 0.8)) })
		}
	}
	// Fairy rings of little toadstools.
	for i, ring := range mushroomForestRings {
		top := []Color{Red, Sand, Terracotta}[i]
		n := int(math.Round(ring.r * 2.2))
		for k := 0; k < n; k++ {
			a := float64(k)*2*math.Pi/float64(n) + 0.3*Hash01(i, k, 101)
			toadstool(Polar(ring.c, ring.r+0.4*Hash01(i, k, 102), a), 0.55+0.35*Hash01(i, k, 103), top)
		}
	}
	inRing := func(q Vec2) bool {
		for _, r := range mushroomForestRings {
			if q.Sub(r.c).Length() < r.r+2.5 {
				return true
			}
		}
		return false
	}
	for gx := -52.0; gx <= 52; gx += 4.3 {
		for gz := -48.0; gz <= 48; gz += 4.3 {
			ix, iz := int(gx*10), int(gz*10)
			q := P(gx+(Hash01(ix, iz, 104)-0.5)*3.2, gz+(Hash01(ix, iz, 105)-0.5)*3.2)
			if !free(q, 1.5) || inRing(q) {
				continue
			}
			r := Hash01(ix, iz, 106)
			n := mushroomForestFBM(q.X/11, q.Y/11, 107)
			h := m.on(q)
			switch {
			case r < 0.3+0.25*n:
				fern(q, 0.8+0.5*Hash01(ix, iz, 108), ix*7+iz)
			case r < 0.48:
				clover(q, ix*5+iz)
			case r < 0.56:
				toadstool(q, 0.5+0.5*Hash01(ix, iz, 109), []Color{Red, Sand, Terracotta, White}[int(Hash01(ix, iz, 110)*4)])
			case r < 0.66:
				g.With(Grass, func() {
					for k := 0; k < 4; k++ {
						b := Polar(q, 0.7, float64(k)*1.7+Hash01(ix, iz, 111)*3)
						g.Fill(Capsule{b.At(h - 0.2), Polar(q, 1.3, float64(k)*1.7).At(h + 1.6 + 0.8*Hash01(ix, iz, k, 112)), 0.45})
					}
				})
			case r > 0.93:
				col := DarkStone
				if Hash01(ix, iz, 113) < 0.5 {
					col = Stone
				}
				rock := Ellipsoid{q.At(h + 0.1), V(1.3+Hash01(ix, iz, 114), 0.9+0.6*Hash01(ix, iz, 115), 1.2+Hash01(ix, iz, 116))}
				g.With(col, func() { g.FillVolume(rock) })
				g.Paint(Foliage, Func{rock.Bounds(), func(p Vec3) bool {
					return rock.Contains(p) && p.Y > h+0.55*rock.Radii.Y && s.hash(g, p, 117) < 0.7
				}})
			}
		}
	}
}
