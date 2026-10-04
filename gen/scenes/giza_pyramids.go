package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// GizaPyramids is the Giza plateau with compressed distances: Khufu's stepped,
// weathered pyramid (north-east), Khafre's on its raised terrace keeping the smooth
// casing cap, and the small Menkaure with its red granite base courses and the gash
// in its north face. Queens' pyramids flank Khufu (east) and Menkaure (south). The
// Great Sphinx lies in its excavated enclosure facing east, with the Sphinx and
// Valley temples below the causeway from Khafre's mortuary temple. Mastaba fields,
// causeways, a camel caravan, palms at the valley edge and rolling dunes complete
// the site. North is -Z, east is +X.
type GizaPyramids struct{}

func init() { Register(GizaPyramids{}) }

func (GizaPyramids) Name() string  { return "giza_pyramids" }
func (GizaPyramids) Title() string { return "Giza Pyramids" }

const (
	gizaPyramidsK       = 0.22 // World units per metre for the pyramids.
	gizaPyramidsPlateau = 2.6  // Plateau level around Khufu.
	gizaPyramidsLow     = 0.5  // Valley floor: Sphinx pit and the temples.
	gizaPyramidsRim     = 0.35 // Height at the site edge.
	gizaPyramidsSphinx  = 1.55 // Sphinx exaggeration over its authored size.
)

// gizaPyramidsPyr is one pyramid: a hollow shell of stepped block courses.
type gizaPyramidsPyr struct {
	c        Vec2
	half, h  float64 // Base half width and full (original) height.
	y0       float64 // Base level.
	ch       float64 // Course height.
	top      float64 // Fraction of the height still standing.
	cap      float64 // Fraction above which the smooth casing survives (0 = none).
	thick    float64 // Horizontal shell thickness.
	notch    float64 // Probability of a missing outer block.
	granite  int     // Courses of red granite casing at the base.
	seed     int
	solidTop float64 // Depth below the top that is filled solid.
	dark     bool    // Paint the inner face dark, for the entrances.
}

func (p gizaPyramidsPyr) halfAt(y float64) (float64, bool, int) {
	run := p.half / p.h
	yc := math.Max(y, 0)
	ci := int(math.Floor(yc / p.ch))
	if p.cap > 0 && yc >= p.cap*p.h {
		return p.half - yc*run, true, ci
	}
	half := p.half - (float64(ci)+0.5)*p.ch*run
	if p.cap > 0 {
		half -= 0.35 // The stepped core has receded below the surviving casing.
	}
	return half, false, ci
}

// shape reports whether a world point lies in the shell and its block color.
func (p gizaPyramidsPyr) shape(q Vec3) (bool, Color) {
	y := q.Y - p.y0
	if y < -1.0 || y > p.h*p.top {
		return false, 0
	}
	half, capped, ci := p.halfAt(y)
	dx, dz := q.X-p.c.X, q.Z-p.c.Y
	ax, az := math.Abs(dx), math.Abs(dz)
	m := math.Max(ax, az)
	if m > half {
		return false, 0
	}
	d := half - m
	solid := y > p.h*p.top-p.solidTop
	if d > p.thick && !solid {
		return false, 0
	}
	face, t := 0, dz
	if ax >= az {
		if dx < 0 {
			face = 1
		}
	} else {
		face, t = 2, dx
		if dz < 0 {
			face = 3
		}
	}
	const bw = 1.7
	j := int(math.Floor((t+float64(ci%2)*bw/2)/bw + 1000))
	hv := Hash01(p.seed, ci, j, face)
	if !capped && d < 0.6 && hv < p.notch && y > 0.4 {
		return false, 0
	}
	if p.dark && d > p.thick-0.3 && !solid {
		return true, Charcoal
	}
	h2 := Hash01(p.seed+7, ci, j, face)
	switch {
	case capped:
		if h2 < 0.18 {
			return true, Sand
		}
		return true, White
	case ci < p.granite && (face != 3 || h2 > 0.5) && Hash01(p.seed, ci, j, face, 5) > 0.3:
		if h2 < 0.15 {
			return true, DarkStone
		}
		return true, Brick
	case h2 < 0.015:
		return true, DarkStone
	}
	course, other := Stone, Sand
	if Hash01(p.seed, ci, 99) < 0.4 {
		course, other = Sand, Stone
	}
	if h2 < 0.13 {
		return true, other
	}
	return true, course
}

func (p gizaPyramidsPyr) build(g *Grid) {
	box := Box{V(p.c.X-p.half-1, p.y0-1.2, p.c.Y-p.half-1), V(p.c.X+p.half+1, p.y0+p.h+1, p.c.Y+p.half+1)}
	for _, col := range []Color{Stone, Sand, DarkStone, Charcoal, White, Brick} {
		col := col
		g.Use(col).Fill(Func{box, func(q Vec3) bool {
			in, c := p.shape(q)
			return in && c == col
		}})
	}
}

// gizaPyramidsMap is a sampled heightfield with bilinear lookup.
type gizaPyramidsMap struct {
	x0, z0, step float64
	nx, nz       int
	h            []float64
}

func (m *gizaPyramidsMap) at(x, z float64) float64 {
	fx, fz := (x-m.x0)/m.step, (z-m.z0)/m.step
	if fx < 0 || fz < 0 || fx >= float64(m.nx-1) || fz >= float64(m.nz-1) {
		return 0
	}
	i, j := int(fx), int(fz)
	tx, tz := fx-float64(i), fz-float64(j)
	k := j*m.nx + i
	a := m.h[k] + (m.h[k+1]-m.h[k])*tx
	b := m.h[k+m.nx] + (m.h[k+m.nx+1]-m.h[k+m.nx])*tx
	return a + (b-a)*tz
}

func (m *gizaPyramidsMap) low(x, z float64) float64 {
	lo := m.at(x, z)
	for k := 0; k < 8; k++ {
		a := float64(k) * math.Pi / 4
		lo = math.Min(lo, m.at(x+0.75*math.Cos(a), z+0.75*math.Sin(a)))
	}
	return lo
}

// gizaPyramidsWay is a causeway: a straight ramp between two deck levels.
type gizaPyramidsWay struct {
	a, b   Vec2
	ya, yb float64
	w      float64 // Half width of the deck.
}

func (w gizaPyramidsWay) near(x, z float64) (float64, float64) {
	ab := w.b.Sub(w.a)
	p := P(x, z)
	t := math.Max(0, math.Min(1, p.Sub(w.a).Dot(ab)/ab.Dot(ab)))
	return p.Sub(w.a.Add(ab.Scale(t))).Length(), t
}

// gizaPyramidsSite holds the layout shared by all the build steps.
type gizaPyramidsSite struct {
	khufu, khafre, menkaure gizaPyramidsPyr
	queens                  []gizaPyramidsPyr
	ways                    []gizaPyramidsWay
	sphinx                  Vec2
	pit                     [2]Vec2 // Sphinx enclosure plan rectangle.
	low                     [2]Vec2 // Valley zone east of the Sphinx.
	fields                  [][2]Vec2
	temples                 [][2]Vec2
	camelA, camelB          Vec2
	m                       *gizaPyramidsMap
}

func (GizaPyramids) noise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

func (s GizaPyramids) fbm(x, z float64, seed int) float64 {
	return (s.noise(x, z, seed)*4 + s.noise(x*2.1, z*2.1, seed+1)*2 + s.noise(x*4.3, z*4.3, seed+2)) / 7
}

func (GizaPyramids) smooth(a, b, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-a)/(b-a)))
	return t * t * (3 - 2*t)
}

// rect is the plan distance from a point to a rectangle (0 inside).
func (GizaPyramids) rect(r [2]Vec2, x, z float64) float64 {
	dx := math.Max(math.Max(r[0].X-x, x-r[1].X), 0)
	dz := math.Max(math.Max(r[0].Y-z, z-r[1].Y), 0)
	return math.Hypot(dx, dz)
}

func (s GizaPyramids) foot(p gizaPyramidsPyr, x, z float64) float64 {
	return s.rect([2]Vec2{P(p.c.X-p.half, p.c.Y-p.half), P(p.c.X+p.half, p.c.Y+p.half)}, x, z)
}

func (s GizaPyramids) layout() *gizaPyramidsSite {
	const k = gizaPyramidsK
	st := &gizaPyramidsSite{}
	st.khufu = gizaPyramidsPyr{c: P(0, 0), half: 230 * k / 2, h: 146.6 * k, y0: gizaPyramidsPlateau,
		ch: 0.8, top: 0.945, thick: 1.6, notch: 0.06, seed: 1, solidTop: 1.4, dark: true}
	kh := st.khufu
	kfHalf := 215 * k / 2
	st.khafre = gizaPyramidsPyr{c: P(-kh.half-4-kfHalf, kh.half+4+kfHalf), half: kfHalf, h: 143.5 * k, y0: gizaPyramidsPlateau + 2.2,
		ch: 0.8, top: 0.975, cap: 0.76, thick: 1.6, notch: 0.05, seed: 2, solidTop: 1.4, dark: true}
	kf := st.khafre
	mc := P(kf.c.X-kf.half-4-104*k/2, kf.c.Y+kf.half+6+104*k/2)
	st.menkaure = gizaPyramidsPyr{c: mc, half: 104 * k / 2, h: 65.5 * k, y0: gizaPyramidsPlateau + 0.9,
		ch: 0.8, top: 0.97, thick: 1.4, notch: 0.08, granite: 2, seed: 3, solidTop: 1.2, dark: true}
	for i, z := range []float64{4, 14, 24} {
		st.queens = append(st.queens, gizaPyramidsPyr{c: P(kh.half+6.5, z), half: 4.5, h: 5.8, y0: gizaPyramidsPlateau,
			ch: 0.8, top: 0.8 + 0.1*Hash01(i, 5), thick: 1.3, notch: 0.12, seed: 10 + i, solidTop: 1.2})
	}
	mq := st.menkaure
	for i, dx := range []float64{10, 0, -10} {
		q := gizaPyramidsPyr{c: P(mq.c.X+dx, mq.c.Y+mq.half+6.5), half: 4.3, h: 6.4, y0: mq.y0 - 0.2,
			ch: 0.8, top: 0.9, thick: 1.3, notch: 0.1, seed: 20 + i, solidTop: 1.2}
		if i > 0 {
			q.ch, q.top = 1.4, 0.88 // Unfinished step pyramids.
		}
		st.queens = append(st.queens, q)
	}
	sk := gizaPyramidsSphinx
	st.sphinx = P(4, kh.half+17)
	sp := st.sphinx
	st.pit = [2]Vec2{P(sp.X-8.8*sk-2.2, sp.Y-2.9*sk-2.6), P(sp.X+8.6*sk+3.5, sp.Y+2.9*sk+2.6)}
	pe := st.pit[1].X
	st.low = [2]Vec2{P(pe-1, st.pit[0].Y-4), P(pe+60, sp.Y+40)}
	st.fields = [][2]Vec2{{P(-kh.half-30, -kh.half-2), P(-kh.half-5, kf.c.Y-kf.half-5)}, {P(kh.half+13, -kh.half-5), P(kh.half+29, -kh.half+11)}}
	st.temples = [][2]Vec2{
		{P(kf.c.X+kf.half, kf.c.Y-5), P(kf.c.X+kf.half+9.5, kf.c.Y+5)}, // Khafre mortuary temple.
		{P(pe, st.pit[0].Y+0.5), P(pe+8.5, st.pit[1].Y-0.5)},           // Sphinx temple.
		{P(pe-2, st.pit[1].Y+2), P(pe+12, st.pit[1].Y+16)},             // Valley temple.
		{P(kh.half+1.2, -12), P(kh.half+8.5, -2)},                      // Khufu mortuary pavement.
		{P(mq.c.X+mq.half, mq.c.Y-5), P(mq.c.X+mq.half+7.5, mq.c.Y+5)}, // Menkaure mortuary temple.
	}
	vt := st.temples[2]
	st.camelA, st.camelB = P(kf.c.X+kf.half+5, kf.c.Y+kf.half+11), P(pe-6, vt[1].Y+9)
	st.ways = []gizaPyramidsWay{
		{P(kf.c.X+kf.half+9.5, kf.c.Y), P(vt[0].X, (vt[0].Y+vt[1].Y)/2), kf.y0, gizaPyramidsLow, 1.1},
		{P(kh.half+8.5, -7), P(kh.half+30, 0), gizaPyramidsPlateau, gizaPyramidsPlateau - 1.2, 1.1},
		{P(mq.c.X+mq.half+7.5, mq.c.Y), P(mq.c.X+mq.half+26, mq.c.Y+5), mq.y0, mq.y0 - 1.2, 1.0},
	}
	return st
}

// sdf is the signed distance to the site outline (negative inside): a margin of
// desert around the monuments, plus the caravan route and the palm grove.
func (s GizaPyramids) sdf(st *gizaPyramidsSite, x, z float64) float64 {
	d := s.monument(st, x, z) - 8
	p := P(x, z)
	for _, c := range []struct {
		a, b Vec2
		r    float64
	}{
		{st.camelA, st.camelB, 8},
		{st.khafre.c, st.sphinx, 20},
		{P(st.khafre.c.X+st.khafre.half, st.khafre.c.Y+st.khafre.half), st.temples[2][0].Add(P(0, 8)), 12},
		{P(st.menkaure.c.X-8, st.khafre.c.Y-30), P(st.menkaure.c.X-12, st.menkaure.c.Y-10), 14},
		{P(st.khafre.c.X+4, st.khafre.c.Y+st.khafre.half+14), P(st.menkaure.c.X+st.menkaure.half+14, st.menkaure.c.Y+12), 12},
		{P(st.low[0].X+6, st.low[0].Y+12), P(st.low[0].X+10, st.low[0].Y+34), 10},
	} {
		ab := c.b.Sub(c.a)
		t := math.Max(0, math.Min(1, p.Sub(c.a).Dot(ab)/ab.Dot(ab)))
		e := p.Sub(c.a.Add(ab.Scale(t))).Length() - c.r
		const kk = 10.0
		hh := math.Max(kk-math.Abs(d-e), 0) / kk
		d = math.Min(d, e) - hh*hh*kk/4
	}
	return d + 4*(s.fbm(x/11, z/11, 41)-0.5)
}

// monument is the plan distance to the nearest built feature (dunes keep clear).
func (s GizaPyramids) monument(st *gizaPyramidsSite, x, z float64) float64 {
	d := math.Min(s.foot(st.khufu, x, z), s.foot(st.khafre, x, z))
	d = math.Min(d, s.foot(st.menkaure, x, z))
	for _, q := range st.queens {
		d = math.Min(d, s.foot(q, x, z))
	}
	for _, r := range st.fields {
		d = math.Min(d, s.rect(r, x, z))
	}
	for _, r := range st.temples {
		d = math.Min(d, s.rect(r, x, z))
	}
	d = math.Min(d, s.rect(st.pit, x, z))
	for _, w := range st.ways {
		e, _ := w.near(x, z)
		d = math.Min(d, e-w.w)
	}
	return math.Max(d, 0)
}

// height is the raw terrain; zero means outside the site.
func (s GizaPyramids) height(st *gizaPyramidsSite, x, z float64) float64 {
	in := -s.sdf(st, x, z)
	if in <= 0 {
		return 0
	}
	h := gizaPyramidsPlateau
	h += 2.2 * (1 - s.smooth(2, 14, s.foot(st.khafre, x, z)))
	h += 0.9 * (1 - s.smooth(2, 12, s.foot(st.menkaure, x, z)))
	h += (gizaPyramidsLow - h) * (1 - s.smooth(0, 9, s.rect(st.low, x, z)))
	md := s.monument(st, x, z)
	// Rolling sand and dunes away from the monuments.
	h += 0.7 * (s.fbm(x/10, z/10, 3) - 0.5) * s.smooth(1, 6, md)
	warp := 3 * s.noise(x/14, z/14, 7)
	u := (x*0.42 + z*0.91 + 6*warp) / 11
	f := u - math.Floor(u)
	ridge := math.Min(f/0.72, (1-f)/0.28) // Gentle windward slope, steep lee.
	h += 2.8 * ridge * (0.3 + s.fbm(x/16, z/16, 9)) * s.smooth(3, 9, md)
	// Flat pads for the pyramids, fields and temples.
	pad := func(level, d, r0, r1 float64) {
		w := 1 - s.smooth(r0, r1, d)
		h += (level - h) * w
	}
	for _, r := range st.fields {
		pad(h, s.rect(r, x, z), 0, 3) // Smooth only: keep the local level.
	}
	for _, q := range st.queens {
		pad(q.y0, s.foot(q, x, z), 1, 4)
	}
	pad(st.khufu.y0, s.foot(st.khufu, x, z), 1.5, 6)
	pad(st.khafre.y0, s.foot(st.khafre, x, z), 1.5, 6)
	pad(st.menkaure.y0, s.foot(st.menkaure, x, z), 1.5, 5)
	// Causeway embankments.
	for _, w := range st.ways {
		d, t := w.near(x, z)
		pad(w.ya+(w.yb-w.ya)*t, d, w.w+0.3, w.w+3.5)
	}
	// The Sphinx enclosure is cut into the rock.
	if s.rect(st.pit, x, z) <= 0 {
		h = gizaPyramidsLow
	}
	h = math.Min(h, gizaPyramidsRim+0.42*in)
	return math.Max(h, gizaPyramidsRim)
}

func (s GizaPyramids) Build(g *Grid) Order {
	st := s.layout()
	m := &gizaPyramidsMap{x0: -125, z0: -60, step: 0.4}
	m.nx = int(215/m.step) + 2
	m.nz = int(195/m.step) + 2
	m.h = make([]float64, m.nx*m.nz)
	for j := 0; j < m.nz; j++ {
		for i := 0; i < m.nx; i++ {
			m.h[j*m.nx+i] = s.height(st, m.x0+float64(i)*m.step, m.z0+float64(j)*m.step)
		}
	}
	st.m = m
	s.terrain(g, st)
	for _, p := range append([]gizaPyramidsPyr{st.khufu, st.khafre, st.menkaure}, st.queens...) {
		p.build(g)
	}
	s.details(g, st)
	s.sphinx(g, st)
	s.temples(g, st)
	s.mastabas(g, st)
	s.caravan(g, st)
	s.palms(g, st)
	return Along(P(-1, 1).Scale(math.Sqrt2 / 2))
}

func (s GizaPyramids) terrain(g *Grid, st *gizaPyramidsSite) {
	m := st.m
	box := Box{V(m.x0, 0, m.z0), V(m.x0+float64(m.nx-1)*m.step, 12, m.z0+float64(m.nz-1)*m.step)}
	under := func(q Vec3) bool {
		for _, p := range []gizaPyramidsPyr{st.khufu, st.khafre, st.menkaure} {
			if math.Abs(q.X-p.c.X) < p.half-p.thick-1 && math.Abs(q.Z-p.c.Y) < p.half-p.thick-1 {
				return true
			}
		}
		return false
	}
	g.Use(Sand).Fill(Func{box, func(q Vec3) bool {
		h := m.at(q.X, q.Z)
		if h <= 0 || under(q) {
			return false
		}
		return q.Y <= h && q.Y >= math.Min(h, m.low(q.X, q.Z))-0.6
	}})
	top := func(q Vec3) bool { return q.Y >= m.at(q.X, q.Z)-0.9 }
	// Gravel and bedrock on the plateau around the monuments.
	g.Paint(Stone, Func{box, func(q Vec3) bool {
		md := s.monument(st, q.X, q.Z)
		return top(q) && md < 4 && s.fbm(q.X/4, q.Z/4, 31) > 0.62+0.03*md
	}})
	g.Paint(DarkStone, Func{box, func(q Vec3) bool {
		return top(q) && s.monument(st, q.X, q.Z) < 9 && Hash01(int(math.Floor(q.X*1.3)), int(math.Floor(q.Z*1.3)), 32) < 0.015
	}})
	// Sphinx enclosure walls: banded limestone strata.
	pit := Func{Box{V(st.pit[0].X-1.5, 0, st.pit[0].Y-1.5), V(st.pit[1].X+1, 6, st.pit[1].Y+1.5)}, func(q Vec3) bool {
		return true
	}}
	g.Paint(Stone, pit)
	g.Paint(DarkStone, Intersect(pit, Func{box, func(q Vec3) bool {
		if q.Y < gizaPyramidsLow+0.25 {
			return Hash01(int(math.Floor(q.X*1.5)), int(math.Floor(q.Z*1.5)), 33) < 0.2
		}
		return int(math.Floor(q.Y/0.8))%2 == 0
	}}))
	// Causeway decks and their side walls.
	for _, w := range st.ways {
		w := w
		g.Paint(Stone, Func{box, func(q Vec3) bool {
			d, _ := w.near(q.X, q.Z)
			return d < w.w && top(q)
		}})
		deck := func(q Vec3) float64 { _, t := w.near(q.X, q.Z); return w.ya + (w.yb-w.ya)*t }
		g.Use(Stone).Fill(Func{Box{V(math.Min(w.a.X, w.b.X)-2, 0, math.Min(w.a.Y, w.b.Y)-2), V(math.Max(w.a.X, w.b.X)+2, 8, math.Max(w.a.Y, w.b.Y)+2)}, func(q Vec3) bool {
			d, t := w.near(q.X, q.Z)
			if d < w.w || d > w.w+0.9 || t <= 0 || t >= 1 || m.at(q.X, q.Z) <= 0 {
				return false
			}
			y := deck(q)
			wall := 0.9
			if Hash01(int(math.Floor(t*40)), 77) < 0.35 {
				wall = 0.3 // Broken stretches.
			}
			return q.Y > y-0.5 && q.Y <= y+wall && q.Y > m.at(q.X, q.Z)-0.5
		}})
	}
}

// details carves the entrances and the Menkaure gash and adds Khufu's survey mast.
func (s GizaPyramids) details(g *Grid, st *gizaPyramidsSite) {
	kh := st.khufu
	ey := 2.7
	half, _, _ := kh.halfAt(ey)
	ex := kh.c.X + 1.1
	g.Carve(Block(V(ex-0.65, kh.y0+ey, kh.c.Y-kh.half-1), V(ex+0.65, kh.y0+ey+1.6, kh.c.Y-half+kh.thick+0.4)))
	// Gabled blocks above the entrance.
	g.Use(Sand).Fill(Gable{Min: P(ex-1.5, kh.c.Y-half-0.2), Max: P(ex+1.5, kh.c.Y-half+1.2), Base: kh.y0 + ey + 1.6, Ridge: AxisZ})
	// Al-Ma'mun's robbers' tunnel, lower and west of centre.
	ry := 1.0
	half, _, _ = kh.halfAt(ry)
	g.Carve(Block(V(kh.c.X-0.4, kh.y0+ry, kh.c.Y-kh.half-1), V(kh.c.X+0.9, kh.y0+ry+1.3, kh.c.Y-half+kh.thick+0.4)))
	// Survey mast on the truncated summit.
	topY := kh.y0 + kh.h*kh.top
	g.Use(Wood).Fill(Cylinder{Center: kh.c, Radius: 0.5, Y0: topY - 0.5, Y1: topY + 2.6})
	// Khafre's entrance.
	kf := st.khafre
	half, _, _ = kf.halfAt(1.2)
	g.Carve(Block(V(kf.c.X-0.6, kf.y0+1.2, kf.c.Y-kf.half-1), V(kf.c.X+0.6, kf.y0+2.6, kf.c.Y-half+kf.thick+0.4)))
	// Menkaure's north face gash.
	mk := st.menkaure
	gash := Func{Box{V(mk.c.X-1.0, mk.y0+0.8, mk.c.Y-mk.half-1), V(mk.c.X+1.0, mk.y0+mk.h*0.55, mk.c.Y)}, func(q Vec3) bool {
		half, _, _ := mk.halfAt(q.Y - mk.y0)
		return math.Abs(q.X-mk.c.X) < 0.9 && q.Z < mk.c.Y-half+1.1
	}}
	g.Carve(gash)
	g.Paint(DarkStone, Func{gash.Box.Grow(0.6), func(q Vec3) bool {
		half, _, _ := mk.halfAt(q.Y - mk.y0)
		return math.Abs(q.X-mk.c.X) < 1.8 && q.Z < mk.c.Y-half+1.8 && q.Y > mk.y0+0.6
	}})
}

// sphinx is the recumbent lion facing east with the nemes headdress.
func (s GizaPyramids) sphinx(g *Grid, st *gizaPyramidsSite) {
	o := st.sphinx.At(gizaPyramidsLow)
	const sk = gizaPyramidsSphinx
	at := func(u, w, v float64) Vec3 { return o.Add(V(u, w, v).Scale(sk)) }
	blk := func(u0, w0, v0, u1, w1, v1 float64) Cuboid { return Block(at(u0, w0, v0), at(u1, w1, v1)) }
	local := func(q Vec3) (float64, float64, float64) {
		return (q.X - o.X) / sk, (q.Y - o.Y) / sk, (q.Z - o.Z) / sk
	}
	region := func(u0, w0, v0, u1, w1, v1 float64, f func(u, w, v float64) bool) Func {
		return Func{Box{at(u0, w0, v0), at(u1, w1, v1)}, func(q Vec3) bool {
			u, w, v := local(q)
			return f(u, w, math.Abs(v))
		}}
	}
	// Bedrock plinth.
	g.Use(Stone).Fill(blk(-8.8, -0.4, -2.9, 7.2, 0.7, 2.9))
	// Body: a long rounded block, haunches lower at the rear, shoulders rising to the neck.
	g.Use(Sand).FillVolume(region(-8.8, 0.5, -2.7, 3.8, 5.2, 2.7, func(u, w, v float64) bool {
		top := 3.9 - 0.7*s.smooth(-5.5, -8.0, u) + 0.6*s.smooth(0, 2.5, u)
		hw := 2.25 + 0.2*s.smooth(-4.5, -6.5, u)
		if u > 2.2 {
			hw -= (u - 2.2) * 0.9 // Rounded chest.
		}
		if u < -7.6 {
			hw -= (-7.6 - u) * 1.4 // Rounded rump.
		}
		if hw <= 0 || w < 0.5 || w > top {
			return false
		}
		a, b := v/hw, (w-0.5)/(top-0.5)
		return a*a*a*a+b*b*b*b <= 1
	}))
	// Tail curled along the right haunch.
	g.FillVolume(Capsule{at(-8.2, 1.2, 2.45), at(-5.0, 1.0, 2.65), 0.45 * sk})
	// Forelegs and paws of restored limestone blocks.
	g.Use(Stone).Fill(blk(2.4, 0.5, 0.7, 8.6, 1.6, 2.2), blk(2.4, 0.5, -2.2, 8.6, 1.6, -0.7))
	// Dream Stele of Thutmose IV between the paws.
	g.Use(Brick).Fill(blk(4.4, 0.5, -0.5, 5.1, 2.9, 0.5))
	// Nemes headdress: lappets flaring onto the shoulders, then the rounded head cloth.
	nemes := region(1.6, 3.6, -2.2, 4.5, 6.9, 2.2, func(u, w, v float64) bool {
		if w < 4.0 {
			return false
		}
		hw := 1.15 + 0.75*s.smooth(5.4, 4.0, w)
		if w < 5.3 {
			return u > 3.2 && u < 4.3 && v < hw
		}
		r := (w - 5.3) / 1.5
		return u > 2.0 && u < 4.4 && v < hw*math.Sqrt(math.Max(1-r*r, 0)+0.05)
	})
	g.Use(Sand).FillVolume(nemes, Ellipsoid{at(2.7, 4.5, 0), V(1.4, 1.1, 1.25).Scale(sk)})
	g.Paint(Gold, Intersect(nemes, region(1.6, 3.6, -2.2, 4.5, 6.9, 2.2, func(u, w, v float64) bool {
		return int(math.Floor(w/0.32))%2 == 0 && (v > 0.95 || u < 3.6)
	})))
	// Face, with eyes and the shadow of the broken nose.
	g.Use(Terracotta).FillVolume(blk(4.1, 4.3, -0.9, 4.95, 6.2, 0.9))
	g.Paint(DarkStone, blk(4.75, 5.3, -0.7, 5.1, 5.6, -0.2), blk(4.75, 5.3, 0.2, 5.1, 5.6, 0.7))
	// Weathered strata on the body.
	g.Paint(Terracotta, region(-9, 0.5, -3, 3.8, 4.6, 3, func(u, w, v float64) bool {
		return math.Mod(w+0.3*s.noise(u, v, 51), 0.9) < 0.22
	}))
	g.Paint(DarkStone, region(-9, -0.4, -3, 7.2, 0.7, 3, func(u, w, v float64) bool {
		return Hash01(int(math.Floor(u*2)), int(math.Floor(v*2)), 52) < 0.25
	}))
}

// temples are roofless courts: Khafre's mortuary temple, the Sphinx and Valley temples
// (red granite pillars, alabaster floor), Khufu's basalt pavement and Menkaure's temple.
func (s GizaPyramids) temples(g *Grid, st *gizaPyramidsSite) {
	court := func(r [2]Vec2, floor, wall, thick float64, seed int, pillars Color, doorW, doorE bool) {
		lo, hi := r[0], r[1]
		y0 := floor - 0.8
		walls := Func{Box{lo.At(y0), hi.At(floor + wall)}, func(q Vec3) bool {
			in := q.X > lo.X+thick && q.X < hi.X-thick && q.Z > lo.Y+thick && q.Z < hi.Y-thick
			if in {
				return false
			}
			cz := (lo.Y + hi.Y) / 2
			if math.Abs(q.Z-cz) < 1.0 && q.Y > floor && ((doorW && q.X < lo.X+thick) || (doorE && q.X > hi.X-thick)) {
				return false
			}
			seg := int(math.Floor((q.X + q.Z) / 2.2))
			hmax := floor + wall*(0.55+0.45*Hash01(seg, seed))
			return q.Y <= hmax
		}}
		g.Use(Stone).Fill(walls)
		g.Paint(Sand, Intersect(walls, Func{walls.Box, func(q Vec3) bool {
			return int(math.Floor(q.Y/0.8)+math.Floor(q.X/1.7)+math.Floor(q.Z/1.7))%3 == 0
		}}))
		// Floor slab.
		g.Use(White).Fill(Block(V(lo.X+thick, y0, lo.Y+thick), V(hi.X-thick, floor, hi.Y-thick)))
		if pillars != Stone {
			for x := lo.X + thick + 1.6; x < hi.X-thick-1.0; x += 3.0 {
				for z := lo.Y + thick + 1.6; z < hi.Y-thick-1.0; z += 3.0 {
					g.Use(pillars).Fill(BlockAt(P(x, z), 0.9, 0.9, floor, floor+wall*0.85))
				}
			}
		}
	}
	kf := st.khafre
	court(st.temples[0], kf.y0, 2.4, 1.0, 1, Brick, true, true)
	court(st.temples[1], gizaPyramidsLow, 2.2, 1.0, 2, Stone, true, true)
	court(st.temples[2], gizaPyramidsLow, 3.4, 1.6, 3, Brick, true, true)
	court(st.temples[4], st.menkaure.y0, 1.8, 0.9, 4, DarkStone, true, true)
	// Khufu's mortuary temple: only the basalt pavement and a few wall stubs remain.
	r := st.temples[3]
	g.Use(Charcoal).Fill(Block(r[0].At(gizaPyramidsPlateau-0.5), r[1].At(gizaPyramidsPlateau+0.3)))
	for i := 0; i < 6; i++ {
		x := r[0].X + 0.8 + (r[1].X-r[0].X-1.6)*Hash01(i, 61)
		z := r[0].Y + 0.8 + (r[1].Y-r[0].Y-1.6)*Hash01(i, 62)
		g.Use(Brick).Fill(BlockAt(P(x, z), 0.9, 0.9, gizaPyramidsPlateau, gizaPyramidsPlateau+0.8+Hash01(i, 63)))
	}
	// The Valley temple's floor is alabaster; the Sphinx temple's court is paved stone.
	g.Paint(Stone, Block(V(st.temples[1][0].X, gizaPyramidsLow-0.3, st.temples[1][0].Y), V(st.temples[1][1].X, gizaPyramidsLow, st.temples[1][1].Y)))
}

// mastabas fill the West and East fields with rows of low battered tombs.
func (s GizaPyramids) mastabas(g *Grid, st *gizaPyramidsSite) {
	m := st.m
	n := 0
	for _, r := range st.fields {
		for x := r[0].X + 2.5; x < r[1].X-1.5; x += 5.4 {
			for z := r[0].Y + 3.4; z < r[1].Y-2.5; z += 7.2 {
				n++
				if Hash01(n, 71) < 0.12 {
					continue
				}
				cx := x + 0.6*(Hash01(n, 72)-0.5)
				w := 2.6 + 0.8*Hash01(n, 73)
				l := 4.4 + 1.6*Hash01(n, 74)
				hgt := 1.2 + 0.9*Hash01(n, 75)
				if Hash01(n, 76) < 0.2 {
					hgt = 0.7 // Ruined.
				}
				base := m.at(cx, z)
				body := Func{Box{V(cx-w/2, base-0.8, z-l/2), V(cx+w/2, base+hgt, z+l/2)}, func(q Vec3) bool {
					dy := math.Max(q.Y-base, 0) * 0.22
					return math.Abs(q.X-cx) <= w/2-dy && math.Abs(q.Z-z) <= l/2-dy
				}}
				col := Sand
				if Hash01(n, 77) < 0.45 {
					col = Stone
				}
				g.Use(col).Fill(body)
				// Burial shaft on the roof and the false door on the east face.
				g.Paint(Charcoal, BlockAt(P(cx, z-l/4), 0.8, 0.8, base+hgt-0.3, base+hgt+0.3))
				g.Paint(DarkStone, Block(V(cx+w/2-0.7, base, z+l/2-1.6), V(cx+w/2+0.2, base+hgt-0.3, z+l/2-0.8)))
			}
		}
	}
}

// caravan is a line of camels crossing the sand south of the Sphinx.
func (s GizaPyramids) caravan(g *Grid, st *gizaPyramidsSite) {
	m := st.m
	a, b := st.camelA, st.camelB
	dir := b.Sub(a).Scale(1 / b.Sub(a).Length())
	heading := math.Atan2(dir.Y, dir.X)
	blankets := []Color{Red, Navy, Gold, Red, Navy, Gold, Red}
	for i := 0; i < 7; i++ {
		t := float64(i) / 6
		c := a.Add(b.Sub(a).Scale(t)).Add(P(0, 1.2*math.Sin(t*math.Pi*1.3)))
		y := m.at(c.X, c.Y)
		place := func(sol Solid) Solid {
			return Translate(RotateY(sol, P(0, 0), heading), c.At(y))
		}
		const k = 1.35
		body := Union(
			Ellipsoid{V(0, 1.9*k, 0), V(1.15*k, 0.5*k, 0.5*k)},
			Ellipsoid{V(-0.1*k, 2.35*k, 0), V(0.6*k, 0.45*k, 0.45*k)},
			Capsule{V(0.75*k, 1.7*k, 0), V(0.8*k, 0.1, 0), 0.5},
			Capsule{V(-0.75*k, 1.7*k, 0), V(-0.85*k, 0.1, 0), 0.5},
			Capsule{V(1.0*k, 2.0*k, 0), V(1.5*k, 2.75*k, 0), 0.45},
			Capsule{V(1.5*k, 2.75*k, 0), V(2.0*k, 2.7*k, 0), 0.42},
		)
		col := Wood
		if Hash01(i, 81) < 0.4 {
			col = Terracotta
		}
		g.Use(col).FillVolume(place(body))
		g.Paint(blankets[i], place(Ellipsoid{V(-0.1*k, 2.45*k, 0), V(0.75*k, 0.45*k, 0.75*k)}))
		if i%2 == 0 {
			g.Use(White).FillVolume(place(Capsule{V(-0.1*k, 3.0*k, 0), V(-0.1*k, 3.7*k, 0), 0.5}))
			g.Use(Sand).FillVolume(place(Sphere(V(-0.1*k, 4.15*k, 0), 0.45)))
		} else {
			g.Use(Wood).FillVolume(place(Block(V(-0.5*k, 2.6*k, -0.75*k), V(0.3*k, 3.2*k, 0.75*k))))
		}
	}
}

// palms mark the edge of the green valley beyond the Valley temple.
func (s GizaPyramids) palms(g *Grid, st *gizaPyramidsSite) {
	m := st.m
	vt := st.temples[2]
	for i, d := range []Vec2{P(3, -6), P(5.5, 0), P(2.5, 5), P(7, 7), P(3, 12), P(8, -10), P(9, 2)} {
		c := P(vt[1].X, vt[0].Y).Add(d)
		y := m.at(c.X, c.Y)
		if y <= 0 {
			continue
		}
		h := 4.5 + 1.5*Hash01(i, 91)
		lean := P(0.25*(Hash01(i, 92)-0.5), 0.25*(Hash01(i, 93)-0.5))
		topP := c.Add(lean.Scale(h)).At(y + h)
		g.Use(Wood).Fill(Capsule{c.At(y - 0.3), topP, 0.5})
		g.Use(Foliage).FillVolume(Ellipsoid{topP.Add(V(0, 0.2, 0)), V(1.9, 0.6, 1.9)})
		for k := 0; k < 6; k++ {
			a := float64(k)*math.Pi/3 + Hash01(i, k, 94)
			tip := Polar(topP.Plan(), 2.6, a).At(topP.Y - 1.1)
			g.FillVolume(Capsule{topP, tip, 0.45})
		}
		g.Paint(Grass, Cylinder{Center: c, Radius: 2.5, Y0: y - 0.8, Y1: y + 0.3})
	}
}
