package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// Meteora is the cluster of Greek Orthodox monasteries perched on rounded sandstone
// pillars above the Thessalian valley. Seven hollow conglomerate pillars of different
// heights rise from wooded scree mounds: their rounded, fissured shafts are banded
// with gray strata, streaked dark by run-off and stained rust, with hermit caves and
// shrub tufts on the ledges. Great Meteoron crowns the tallest, Varlaam the next;
// Roussanou covers its narrow pillar to the cliff edge; Holy Trinity is reached by a
// stair spiralling round its shaft; St. Stephen is joined by a stone bridge to a
// chapel pinnacle. Byzantine churches with brick-banded walls and terracotta domes,
// whitewashed and stone cell ranges, wooden balconies over the drop and winch towers
// with hanging nets stand on top. In front lies the valley: a village of white
// houses with tiled roofs, a river with a stone bridge, roads, fields and olive groves.
// East (+X) is to the right; the village side (+Z) is the front.
type Meteora struct{}

func init() { Register(Meteora{}) }

func (Meteora) Name() string  { return "meteora" }
func (Meteora) Title() string { return "Meteora" }

// meteoraRock is a pillar with an elliptical plan of semi-axes a (X) and b (Z), a
// rounded top at height top, and a scree mound of height foot around its base.
type meteoraRock struct {
	c         Vec2
	a, b, top float64
	foot      float64
	seed      int
	shoulder  float64
	lean      Vec2 // Plan offset per unit of depth below the top.
}

var meteoraRocks = []meteoraRock{
	{c: P(-12, -10), a: 11, b: 8.5, top: 42, foot: 7, seed: 1, shoulder: 4, lean: P(0.03, 0.02)},         // Great Meteoron
	{c: P(12, -17), a: 8.5, b: 7, top: 36, foot: 6.5, seed: 2, shoulder: 3.5, lean: P(-0.04, 0.01)},      // Varlaam
	{c: P(-31, 4), a: 5.5, b: 4.5, top: 27, foot: 5, seed: 3, shoulder: 3, lean: P(0.05, -0.03)},         // Roussanou
	{c: P(22, 2), a: 5.5, b: 5, top: 31, foot: 5.5, seed: 4, shoulder: 3, lean: P(-0.03, -0.03)},         // Holy Trinity
	{c: P(37, -19), a: 6.5, b: 6, top: 30, foot: 6, seed: 5, shoulder: 3, lean: P(-0.05, 0.02)},          // St. Stephen
	{c: P(23.5, -25.5), a: 4.2, b: 4, top: 29.5, foot: 5.5, seed: 6, shoulder: 2.5, lean: P(0.04, 0.05)}, // Chapel pinnacle
	{c: P(-34, -16), a: 4.2, b: 4.4, top: 23, foot: 5, seed: 7, shoulder: 3, lean: P(0.07, 0.02)},        // Bare pinnacle
}

func meteoraSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

// meteoraNoise2 is smooth value noise in [0, 1).
func meteoraNoise2(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

func meteoraFBM(x, z float64, seed int) float64 {
	return (meteoraNoise2(x, z, seed)*4 + meteoraNoise2(x*2.1+5, z*2.1+3, seed+1)*2 + meteoraNoise2(x*4.3+1, z*4.3+7, seed+2)) / 7
}

// meteoraNoise3 is smooth 3D value noise in [0, 1).
func meteoraNoise3(x, y, z float64, seed int) float64 {
	ix, iy, iz := math.Floor(x), math.Floor(y), math.Floor(z)
	fx, fy, fz := x-ix, y-iy, z-iz
	sx, sy, sz := fx*fx*(3-2*fx), fy*fy*(3-2*fy), fz*fz*(3-2*fz)
	i, j, k := int(ix), int(iy), int(iz)
	lerp := func(a, b, t float64) float64 { return a + (b-a)*t }
	c := func(di, dj, dk int) float64 { return Hash01(i+di, j+dj, k+dk, seed) }
	x00 := lerp(c(0, 0, 0), c(1, 0, 0), sx)
	x10 := lerp(c(0, 1, 0), c(1, 1, 0), sx)
	x01 := lerp(c(0, 0, 1), c(1, 0, 1), sx)
	x11 := lerp(c(0, 1, 1), c(1, 1, 1), sx)
	return lerp(lerp(x00, x10, sy), lerp(x01, x11, sy), sz)
}

// e is the normalized elliptical radius of a plan point at the rock's top.
func (r meteoraRock) e(p Vec2) float64 {
	dx, dz := (p.X-r.c.X)/r.a, (p.Y-r.c.Y)/r.b
	return math.Sqrt(dx*dx + dz*dz)
}

// e3 is the normalized radius at a height, following the pillar's lean.
func (r meteoraRock) e3(p Vec3) float64 {
	return r.e(p.Plan().Sub(r.lean.Scale(r.top - p.Y)))
}

// eFoot is the normalized radius at the pillar foot.
func (r meteoraRock) eFoot(p Vec2) float64 { return r.e3(p.At(2)) }

// limit is the normalized radius of the rock surface near p.
func (r meteoraRock) limit(p Vec3) float64 {
	y := p.Y
	f := 1.0
	// Flared base.
	if k := 1 - (y-1)/(r.foot+3); k > 0 {
		f += 0.2 * k * k
	}
	// Gentle bulges along the shaft.
	f += 0.04*math.Sin(y*0.31+float64(r.seed)*1.7) - 0.05*y/r.top
	// Eroded strata: each band leans out toward its top, leaving up-facing ledges.
	band := (y + 1.6*meteoraNoise2(p.X/7, p.Z/7, r.seed+40)) / 6.5
	f += 0.03 * (band - math.Floor(band))
	// Vertical fissures and lumps.
	f += 0.14*(meteoraNoise3(p.X/1.7, y/12, p.Z/1.7, r.seed+10)-0.5) + 0.2*(meteoraNoise3(p.X/5, y/20, p.Z/5, r.seed+20)-0.5)
	// Rounded shoulder up to the flat top.
	if t := (y - (r.top - r.shoulder)) / r.shoulder; t > 0 {
		f *= math.Cbrt(math.Max(0, 1-t*t*t))
	}
	return f
}

func (r meteoraRock) inside(p Vec3) bool {
	if p.Y > r.top || p.Y < -1 {
		return false
	}
	e := r.e3(p)
	if e > 1.45 {
		return false
	}
	return e <= r.limit(p)
}

func (r meteoraRock) bounds() Box {
	shift := r.lean.Scale(r.top + 1)
	b := Box{V(r.c.X-r.a*1.45, -1, r.c.Y-r.b*1.45), V(r.c.X+r.a*1.45, r.top, r.c.Y+r.b*1.45)}
	return b.Union(b.Translate(V(shift.X, 0, shift.Y)))
}

// surface is the point on the rock face in plan direction phi at height y.
func (r meteoraRock) surface(phi, y float64) Vec3 {
	s, c := math.Sincos(phi)
	o := r.c.Add(r.lean.Scale(r.top - y))
	for e := 1.45; e > 0.2; e -= 0.02 {
		q := V(o.X+r.a*c*e, y, o.Y+r.b*s*e)
		if r.inside(q) {
			return q
		}
	}
	return V(r.c.X, y, r.c.Y)
}

// meteoraShellDirs are the 14 lattice neighbor directions, normalized.
var meteoraShellDirs = func() []Vec3 {
	var d []Vec3
	for _, n := range Neighbors {
		v := V(float64(n.X), float64(n.Y), float64(n.Z))
		d = append(d, v.Scale(1/v.Length()))
	}
	return d
}()

// shell keeps the part of s within t of its surface.
func (Meteora) shell(s Solid, t float64) Solid {
	return Func{s.Bounds(), func(p Vec3) bool {
		if !s.Contains(p) {
			return false
		}
		for _, d := range meteoraShellDirs {
			if !s.Contains(p.Add(d.Scale(t))) {
				return true
			}
		}
		return false
	}}
}

// meteoraMap is a sampled terrain heightfield with bilinear lookup.
type meteoraMap struct {
	x0, z0, step float64
	nx, nz       int
	h            []float64
}

func (m *meteoraMap) at(x, z float64) float64 {
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

func (m *meteoraMap) low(x, z float64) float64 {
	lo := m.at(x, z)
	for k := 0; k < 8; k++ {
		a := float64(k) * math.Pi / 4
		lo = math.Min(lo, m.at(x+1.2*math.Cos(a), z+1.2*math.Sin(a)))
	}
	return lo
}

func meteoraRiver(x float64) float64 { return 21 + 3*math.Sin(x/10+0.4) }

// meteoraRoads are the valley roads as polylines.
var meteoraRoads = [][]Vec2{
	{P(-52, 13), P(-38, 13.5), P(-24, 12), P(-13, 9.5), P(-4, 11), P(6, 12), P(16, 11.5), P(30, 11), P(40, 8), P(52, 7)},
	{P(-4, 11), P(-3, 5), P(0, -1), P(3, -6), P(4, -10)},
	{P(3, 38), P(3.5, 28), P(3.5, 18), P(5, 12)},
	{P(30, 11), P(33, 2), P(34, -8), P(33, -14)},
}

func meteoraSegDist(p, a, b Vec2) float64 {
	ab := b.Sub(a)
	t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
	return p.Sub(a.Add(ab.Scale(t))).Length()
}

func (Meteora) roadDist(p Vec2) float64 {
	best := math.Inf(1)
	for _, road := range meteoraRoads {
		for i := 0; i+1 < len(road); i++ {
			best = math.Min(best, meteoraSegDist(p, road[i], road[i+1]))
		}
	}
	return best
}

// rim is the normalized site radius; above 1 is off the plot.
func (Meteora) rim(x, z float64) float64 {
	ex, ez := x/52, (z+5)/36
	return math.Hypot(ex, ez) + 0.05*(meteoraFBM(x/9, z/9, 1)-0.5)
}

// height is the valley floor, scree mounds and river channel; zero is off the plot.
func (m Meteora) height(x, z float64) float64 {
	e := m.rim(x, z)
	if e > 1 {
		return 0
	}
	h := 1.9 + 1.1*meteoraFBM(x/14, z/14, 2) + 4.5*meteoraSmooth((-z-6)/24)
	for _, r := range meteoraRocks {
		d := math.Max(0, (r.eFoot(P(x, z))-1)*math.Min(r.a, r.b))
		mound := r.foot * math.Exp(-d/4.2) * (0.8 + 0.4*meteoraFBM(x/5, z/5, 3+r.seed))
		h = math.Max(h, 1.9+mound)
	}
	w := math.Abs(z - meteoraRiver(x))
	h = math.Min(h, 1.1+math.Max(0, w-2.4)*0.9)
	return math.Min(h, 0.6+25*(1-e))
}

func (m Meteora) Build(g *Grid) Order {
	mp := &meteoraMap{x0: -58, z0: -44, step: 0.4}
	mp.nx, mp.nz = int(116/mp.step)+2, int(82/mp.step)+2
	mp.h = make([]float64, mp.nx*mp.nz)
	for j := 0; j < mp.nz; j++ {
		for i := 0; i < mp.nx; i++ {
			mp.h[j*mp.nx+i] = m.height(mp.x0+float64(i)*mp.step, mp.z0+float64(j)*mp.step)
		}
	}
	m.terrain(g, mp)
	for _, r := range meteoraRocks {
		m.rock(g, mp, r)
	}
	m.greatMeteoron(g)
	m.varlaam(g)
	m.roussanou(g)
	m.holyTrinity(g)
	m.stStephen(g)
	m.village(g, mp)
	m.trees(g, mp)
	return Sweep(P(0, -6), math.Pi/2)
}

// terrain fills the heightfield as a hollow skin and colors fields, roads and river.
func (m Meteora) terrain(g *Grid, mp *meteoraMap) {
	box := Box{V(mp.x0, 0, mp.z0), V(mp.x0+116, 20, mp.z0+82)}
	t := NewGridWithSpan(g.Span())
	t.Use(Grass).Fill(Func{box, func(p Vec3) bool {
		h := mp.at(p.X, p.Z)
		if h <= 0 {
			return false
		}
		for _, r := range meteoraRocks {
			if r.eFoot(p.Plan()) < 0.85 {
				return false
			}
		}
		return p.Y <= h && p.Y >= math.Min(h, mp.low(p.X, p.Z))-0.8 && p.Y >= 0
	}})
	// Patchwork fields on the valley floor.
	t.Paint(Sand, Func{box, func(p Vec3) bool {
		fi, fj := int(math.Floor((p.X+200)/7.5)), int(math.Floor((p.Z+200)/5.5))
		return p.Z > 4 && Hash01(fi, fj, 50) < 0.3 && mp.at(p.X, p.Z) < 3.4 && math.Abs(p.Z-meteoraRiver(p.X)) > 4.5
	}})
	t.Paint(Wood, Func{box, func(p Vec3) bool {
		fi, fj := int(math.Floor((p.X+200)/7.5)), int(math.Floor((p.Z+200)/5.5))
		h := Hash01(fi, fj, 50)
		return p.Z > 4 && h >= 0.3 && h < 0.42 && mp.at(p.X, p.Z) < 3.4 && math.Abs(p.Z-meteoraRiver(p.X)) > 4.5
	}})
	// Wooded scree around the pillar feet, with bare rubble right at the rock.
	t.Paint(Foliage, Func{box, func(p Vec3) bool {
		h := mp.at(p.X, p.Z)
		return h > 3.6+1.2*meteoraFBM(p.X/6, p.Z/6, 60)
	}})
	t.Paint(DarkStone, Func{box, func(p Vec3) bool {
		for _, r := range meteoraRocks {
			if r.eFoot(p.Plan()) < 1.12+0.12*meteoraNoise2(p.X/2, p.Z/2, 61) {
				return true
			}
		}
		return false
	}})
	// Roads and river.
	t.Paint(Stone, Func{box, func(p Vec3) bool {
		return m.roadDist(p.Plan()) < 1.2 && p.Y >= mp.at(p.X, p.Z)-1
	}})
	t.Paint(Water, Func{box, func(p Vec3) bool {
		return math.Abs(p.Z-meteoraRiver(p.X)) < 2.6 && mp.at(p.X, p.Z) < 1.3
	}})
	g.Merge(t)
	// Stone bridge over the river.
	zr := meteoraRiver(3.5)
	g.With(Stone, func() {
		g.Fill(Subtract(Block(V(1.5, 0.6, zr-4), V(5.5, 2.6, zr+4)),
			Arch{Sill: V(3.5, 0.4, zr), HalfWidth: 2.2, Spring: 0.2, Depth: 6, Through: AxisX}))
		g.Fill(Block(V(1.5, 2.6, zr-4), V(2.2, 3.4, zr+4)), Block(V(4.8, 2.6, zr-4), V(5.5, 3.4, zr+4)))
	})
}

// rock builds one pillar as a hollow, weathered shell with caves and shrub tufts.
func (m Meteora) rock(g *Grid, mp *meteoraMap, r meteoraRock) {
	t := NewGridWithSpan(g.Span())
	const thick = 1.1
	t.Use(Sand).Fill(Func{r.bounds(), func(p Vec3) bool {
		if p.Y < mp.at(p.X, p.Z)-2 || !r.inside(p) {
			return false
		}
		for _, d := range meteoraShellDirs {
			if !r.inside(p.Add(d.Scale(thick))) {
				return true
			}
		}
		return false
	}})
	b := r.bounds()
	s := r.seed * 100
	// Gray conglomerate strata.
	t.Paint(Stone, Func{b, func(p Vec3) bool {
		band := int(math.Floor((p.Y + 2.5*meteoraNoise3(p.X/8, p.Y/8, p.Z/8, s+1)) / 2.3))
		return Hash01(band, s+2) < 0.22
	}})
	// Rust stains.
	t.Paint(Terracotta, Func{b, func(p Vec3) bool {
		return meteoraNoise3(p.X/3.5, p.Y/6, p.Z/3.5, s+3) > 0.74
	}})
	// Dark run-off streaks, denser toward the top.
	t.Paint(DarkStone, Func{b, func(p Vec3) bool {
		k := (p.Y - r.foot) / (r.top - r.foot)
		return meteoraNoise3(p.X/1.3, p.Y/18, p.Z/1.3, s+4) > 0.66-0.12*k
	}})
	// Hermit caves: shallow hollows that open into the dark core.
	for i := 0; i < 3; i++ {
		phi := 2 * math.Pi * Hash01(r.seed, i, 70)
		y := r.foot + 4 + (r.top-r.foot-10)*Hash01(r.seed, i, 71)
		q := r.surface(phi, y)
		t.Paint(DarkStone, Ellipsoid{q, V(2.2, 2.6, 2.2)})
		t.Carve(Ellipsoid{q, V(0.9, 1.3, 0.9)})
	}
	// Shrub tufts on ledges.
	n := 6 + int(r.a+r.b)/2
	for i := 0; i < n; i++ {
		phi := 2 * math.Pi * Hash01(r.seed, i, 72)
		y := r.foot + 3 + (r.top-r.foot-6)*Hash01(r.seed, i, 73)
		q := r.surface(phi, y)
		rad := 1.0 + 0.7*Hash01(r.seed, i, 74)
		t.With(Foliage, func() { t.Fill(Ellipsoid{q.Add(V(0, 0.3, 0)), V(rad, rad*0.7, rad)}) })
	}
	if r.seed == 7 {
		// The bare pinnacle keeps a scrubby crown.
		t.Paint(Foliage, Func{b, func(p Vec3) bool { return p.Y > r.top-1.6 && r.e3(p) < 0.75 }})
		for i := 0; i < 4; i++ {
			q := Polar(r.c, 1.8*Hash01(i, 75), 2*math.Pi*Hash01(i, 76))
			t.With(Foliage, func() { t.Fill(Ellipsoid{q.At(r.top + 0.8), V(1.6, 1.3, 1.6)}) })
		}
	}
	g.Merge(t)
}

// box is a hollow masonry block with walls about one unit thick.
func (m Meteora) box(g *Grid, c Color, lo, hi Vec3) {
	g.With(c, func() { g.Fill(m.shell(Block(lo, hi), 1.0)) })
}

// windows carves rows of openings through the long faces of a block (along X or Z).
func (Meteora) windows(g *Grid, lo, hi Vec3, ridge Axis, y0 float64, rows int, spacing float64) {
	for k := 0; k < rows; k++ {
		y := y0 + float64(k)*2.6
		if y+1.4 > hi.Y-0.3 {
			break
		}
		if ridge == AxisX {
			n := int((hi.X - lo.X - 1) / spacing)
			for i := 0; i < n; i++ {
				x := lo.X + (hi.X-lo.X)/2 + (float64(i)-float64(n-1)/2)*spacing
				g.Carve(Block(V(x-0.5, y, lo.Z-0.5), V(x+0.5, y+1.4, lo.Z+0.9)), Block(V(x-0.5, y, hi.Z-0.9), V(x+0.5, y+1.4, hi.Z+0.5)))
			}
		} else {
			n := int((hi.Z - lo.Z - 1) / spacing)
			for i := 0; i < n; i++ {
				z := lo.Z + (hi.Z-lo.Z)/2 + (float64(i)-float64(n-1)/2)*spacing
				g.Carve(Block(V(lo.X-0.5, y, z-0.5), V(lo.X+0.9, y+1.4, z+0.5)), Block(V(hi.X-0.9, y, z-0.5), V(hi.X+0.5, y+1.4, z+0.5)))
			}
		}
	}
}

// cells is a monastery range: masonry foundation down the cliff, plastered or stone
// storeys from floor to floor+h with windows, and a tiled gable roof.
func (m Meteora) cells(g *Grid, x0, z0, x1, z1, foot, floor, h float64, wall, roof Color, ridge Axis) {
	lo, hi := V(x0, foot, z0), V(x1, floor+h, z1)
	m.box(g, wall, lo, hi)
	g.Paint(Stone, Block(V(x0-1, foot, z0-1), V(x1+1, floor+0.6, z1+1)))
	m.windows(g, lo, hi, ridge, floor+1, 2, 2.2)
	g.With(roof, func() {
		g.Fill(m.shell(Gable{Min: P(x0-0.5, z0-0.5), Max: P(x1+0.5, z1+0.5), Base: floor + h - 0.2, Ridge: ridge}, 1.0))
	})
}

// church is a Byzantine cross-in-square katholikon facing east (+X): brick-banded
// walls, tiled gables, an apse, and terracotta domes on drums over the crossing and
// the narthex.
func (m Meteora) church(g *Grid, c Vec2, floor, length, width, h float64) {
	x0, x1 := c.X-length/2, c.X+length/2
	z0, z1 := c.Y-width/2, c.Y+width/2
	foot := floor - 3
	body := Union(
		Block(V(x0, foot, z0), V(x1, floor+h, z1)),
		Block(V(c.X-1.7, foot, z0-1.2), V(c.X+1.7, floor+h, z1+1.2)),
	)
	ar := width/2 - 0.3
	apse := Intersect(Cylinder{Center: P(x1, c.Y), Radius: ar, Y0: foot, Y1: floor + h - 0.8}, Block(V(x1-0.5, foot, z0), V(x1+ar+1, floor+h, z1)))
	g.With(Sand, func() { g.Fill(m.shell(body, 0.9), m.shell(apse, 0.9)) })
	// Cloisonné brick bands.
	bb := Box{V(x0-1, floor, z0-2), V(x1+ar+1, floor+h, z1+2)}
	g.Paint(Brick, Func{bb, func(p Vec3) bool {
		f := (p.Y - floor) / 1.6
		return f-math.Floor(f) < 0.3
	}})
	// Door and windows.
	g.Carve(Arch{Sill: V(x0, floor, c.Y), HalfWidth: 0.8, Spring: 1.3, Depth: 2, Through: AxisX})
	for _, z := range []float64{z0 - 1.2, z1 + 1.2} {
		g.Carve(Arch{Sill: V(c.X, floor+1, z), HalfWidth: 0.6, Spring: 1.0, Depth: 2, Through: AxisZ})
	}
	for _, x := range []float64{x0 + 1.4, x1 - 1.4} {
		for _, z := range []float64{z0, z1} {
			g.Carve(Arch{Sill: V(x, floor+1.2, z), HalfWidth: 0.5, Spring: 0.8, Depth: 1.8, Through: AxisZ})
		}
	}
	// Roofs.
	g.With(Terracotta, func() {
		g.Fill(m.shell(Gable{Min: P(x0-0.4, z0-0.4), Max: P(x1+0.4, z1+0.4), Base: floor + h - 0.2, Ridge: AxisX}, 1.0))
		g.Fill(m.shell(Gable{Min: P(c.X-2.1, z0-1.6), Max: P(c.X+2.1, z1+1.6), Base: floor + h - 0.2, Ridge: AxisZ}, 1.0))
		g.Fill(Intersect(Cone(P(x1, c.Y), ar+0.5, floor+h-1, ar+0.3), Block(V(x1-0.2, 0, z0-1), V(x1+ar+2, 99, z1+1))))
	})
	// Domes.
	ridge := floor + h - 0.2 + width/2 + 0.4
	m.dome(g, c, ridge-1, ridge+1.6, 1.7)
	m.dome(g, P(x0+1.4, c.Y), ridge-1, ridge+0.5, 1.15)
}

// dome is a drum with dark window slits under a terracotta hemisphere.
func (m Meteora) dome(g *Grid, c Vec2, y0, y1, r float64) {
	g.With(Sand, func() { g.Fill(Cylinder{Center: c, Radius: r, Y0: y0, Y1: y1}) })
	for k := 0; k < 8; k++ {
		q := Polar(c, r, float64(k)*math.Pi/4)
		g.Paint(DarkStone, Block(V(q.X-0.3, y1-1.4, q.Y-0.3), V(q.X+0.3, y1-0.4, q.Y+0.3)))
	}
	g.With(Terracotta, func() {
		g.FillVolume(Intersect(Ellipsoid{c.At(y1), V(r+0.3, r+0.5, r+0.3)}, Block(V(c.X-r-1, y1, c.Y-r-1), V(c.X+r+1, y1+r+1, c.Y+r+1))))
		g.Fill(Cylinder{Center: c, Radius: 0.45, Y0: y1 + r, Y1: y1 + r + 1.1})
	})
}

// balcony is a covered wooden gallery hung on a wall face from a to b, projecting
// depth along out at floor height y, propped by struts to the rock below.
func (m Meteora) balcony(g *Grid, a, b, out Vec2, y, depth float64) {
	far := out.Scale(depth)
	back := out.Scale(-0.6)
	quad := func(d0, d1 float64) []Vec2 {
		return []Vec2{a.Add(out.Scale(d0)), b.Add(out.Scale(d0)), b.Add(out.Scale(d1)), a.Add(out.Scale(d1))}
	}
	g.With(Wood, func() {
		g.Fill(Prism{Polygon: []Vec2{a.Add(back), b.Add(back), b.Add(far), a.Add(far)}, Y0: y - 0.8, Y1: y})
		g.Fill(Prism{Polygon: quad(depth-0.8, depth), Y0: y, Y1: y + 1.1})
		along := b.Sub(a)
		n := int(along.Length()/2.6) + 1
		for i := 0; i <= n; i++ {
			q := a.Add(along.Scale(float64(i) / float64(n)))
			tip := q.Add(out.Scale(depth - 0.4))
			g.Fill(Capsule{tip.At(y), tip.At(y + 3), 0.45})
			g.Fill(Capsule{tip.At(y - 0.6), q.Add(out.Scale(-0.3)).At(y - 3.4), 0.5})
		}
	})
	g.With(Terracotta, func() {
		g.Fill(Func{Prism{Polygon: quad(-0.6, depth+0.5), Y0: y + 2.8, Y1: y + 4.2}.Bounds(), func(p Vec3) bool {
			d := p.Plan().Sub(a).Dot(out)
			top := y + 4.2 - d*0.5
			return InsidePolygon(quad(-0.6, depth+0.5), p.Plan()) && p.Y <= top && p.Y >= top-0.9
		}})
	})
}

// winch is a stone tower with a timber hoist room projecting over the cliff along
// out, a beam, and a rope hanging down to a net.
func (m Meteora) winch(g *Grid, c, out Vec2, foot, floor, netY float64) {
	side := P(-out.Y, out.X)
	sq := func(center Vec2, half, depth0, depth1 float64) []Vec2 {
		return []Vec2{
			center.Add(out.Scale(depth0)).Add(side.Scale(-half)), center.Add(out.Scale(depth0)).Add(side.Scale(half)),
			center.Add(out.Scale(depth1)).Add(side.Scale(half)), center.Add(out.Scale(depth1)).Add(side.Scale(-half)),
		}
	}
	g.With(Stone, func() { g.Fill(m.shell(Prism{Polygon: sq(c, 1.8, -1.8, 1.8), Y0: foot, Y1: floor + 2.5}, 1.0)) })
	g.With(Wood, func() {
		g.Fill(m.shell(Prism{Polygon: sq(c, 1.6, 0.5, 4.2), Y0: floor + 2.5, Y1: floor + 5.5}, 0.9))
		g.Fill(Capsule{c.Add(out.Scale(1.8)).At(floor + 2.2), c.Add(out.Scale(0)).At(floor - 1.5), 0.5})
		tip := c.Add(out.Scale(5.6))
		g.Fill(Capsule{c.Add(out.Scale(3)).At(floor + 4.6), tip.At(floor + 4.6), 0.5})
		g.Fill(Capsule{tip.At(floor + 4.4), tip.At(netY + 1), 0.5})
		g.FillVolume(Ellipsoid{tip.At(netY), V(1.1, 1.2, 1.1)})
	})
	// Carve the hoist room's open front so the beam reads.
	g.Carve(Prism{Polygon: sq(c, 0.9, 3.4, 4.4), Y0: floor + 3.1, Y1: floor + 4.9})
	g.With(Terracotta, func() {
		plan := EmptyBox()
		for _, q := range sq(c, 2.2, -2.2, 4.7) {
			plan = plan.Union(Box{q.At(0), q.At(0)})
		}
		g.Fill(m.shell(HipRoof{Min: P(plan.Min.X, plan.Min.Z), Max: P(plan.Max.X, plan.Max.Z), Base: floor + 5.3}, 1.0))
	})
}

// bell is a small square bell tower with open arches and a hipped roof.
func (m Meteora) bell(g *Grid, c Vec2, foot, top float64) {
	lo, hi := V(c.X-1.3, foot, c.Y-1.3), V(c.X+1.3, top, c.Y+1.3)
	m.box(g, Stone, lo, hi)
	g.Carve(Block(V(c.X-0.5, top-2.4, lo.Z-1), V(c.X+0.5, top-0.8, hi.Z+1)), Block(V(lo.X-1, top-2.4, c.Y-0.5), V(hi.X+1, top-0.8, c.Y+0.5)))
	g.With(Terracotta, func() { g.Fill(HipRoof{Min: P(lo.X-0.4, lo.Z-0.4), Max: P(hi.X+0.4, hi.Z+0.4), Base: top}) })
}

// cypress is a slim dark evergreen.
func (Meteora) cypress(g *Grid, at Vec2, base, h float64) {
	g.With(Wood, func() { g.Fill(Cylinder{Center: at, Radius: 0.45, Y0: base - 0.5, Y1: base + 1}) })
	g.With(Foliage, func() { g.FillVolume(Ellipsoid{at.At(base + 0.8 + h/2), V(0.95, h/2, 0.95)}) })
}

// paving paints a rock top's courtyard surface.
func (Meteora) paving(g *Grid, r meteoraRock) {
	g.Paint(Stone, Func{Box{V(r.c.X-r.a, r.top-0.9, r.c.Y-r.b), V(r.c.X+r.a, r.top+0.1, r.c.Y+r.b)}, func(p Vec3) bool {
		return r.e(p.Plan()) < 0.6
	}})
	g.Paint(Grass, Func{Box{V(r.c.X-r.a, r.top-0.9, r.c.Y-r.b), V(r.c.X+r.a, r.top+0.1, r.c.Y+r.b)}, func(p Vec3) bool {
		e := r.e(p.Plan())
		return e >= 0.6 && meteoraNoise2(p.X/2.5, p.Z/2.5, r.seed+90) > 0.45
	}})
}

func (m Meteora) greatMeteoron(g *Grid) {
	r := meteoraRocks[0]
	T := r.top
	m.paving(g, r)
	m.church(g, P(-12.5, -11), T, 11, 4.4, 4.4)
	m.cells(g, -18, -17, -6, -14, T-7, T, 5.2, White, Terracotta, AxisX)
	m.cells(g, -18, -6, -10, -3, T-7, T, 4.6, Stone, Brick, AxisX)
	m.balcony(g, P(-17.5, -3), P(-10.5, -3), P(0, 1), T+2.4, 1.9)
	m.cells(g, -22, -12, -19, -6, T-6, T, 3.6, White, Terracotta, AxisZ)
	m.bell(g, P(-7.5, -5.5), T-3, T+8)
	m.winch(g, P(-3, -10), P(1, 0), T-6, T, T-18)
	m.cypress(g, P(-9, -7), T, 4.5)
	m.cypress(g, P(-16, -7.6), T, 3.8)
}

func (m Meteora) varlaam(g *Grid) {
	r := meteoraRocks[1]
	T := r.top
	m.paving(g, r)
	m.church(g, P(11.5, -18), T, 9, 4, 4)
	m.cells(g, 8, -13.5, 14, -11, T-6, T, 4.2, White, Terracotta, AxisX)
	m.balcony(g, P(8.5, -11), P(13.5, -11), P(0, 1), T+2.2, 1.8)
	m.cells(g, 7, -23.5, 16, -21, T-6, T, 4.4, Stone, Brick, AxisX)
	m.winch(g, P(18, -16.5), P(1, 0), T-6, T, T-16)
	m.cypress(g, P(15.5, -13), T, 3.6)
}

func (m Meteora) roussanou(g *Grid) {
	r := meteoraRocks[2]
	T := r.top
	m.paving(g, r)
	// A ring of cells built flush with the cliff, three storeys over the drop.
	outer := EllipticCylinder{Center: r.c, A: r.a * 0.98, B: r.b * 0.98, Y0: T - 7, Y1: T + 4.5}
	inner := EllipticCylinder{Center: r.c, A: r.a*0.98 - 2.6, B: r.b*0.98 - 2.6, Y0: T - 8, Y1: T + 9}
	ring := Subtract(outer, inner)
	g.With(White, func() { g.Fill(m.shell(ring, 1.0)) })
	g.Paint(Stone, Block(V(r.c.X-r.a-1, T-7, r.c.Y-r.b-1), V(r.c.X+r.a+1, T-1.5, r.c.Y+r.b+1)))
	for k := 0; k < 14; k++ {
		a := float64(k) * 2 * math.Pi / 14
		s, c := math.Sincos(a)
		for _, y := range []float64{T + 0.6, T - 2.4} {
			q := P(r.c.X+r.a*0.98*c, r.c.Y+r.b*0.98*s)
			g.Carve(Block(V(q.X-0.5, y, q.Y-0.5), V(q.X+0.5, y+1.4, q.Y+0.5)))
		}
	}
	// Ring roof: a ridge running round the middle of the range.
	g.With(Terracotta, func() {
		g.Fill(m.shell(Func{Box{V(r.c.X-r.a-1, T+4, r.c.Y-r.b-1), V(r.c.X+r.a+1, T+6.5, r.c.Y+r.b+1)}, func(p Vec3) bool {
			e := r.e(p.Plan())
			rr := math.Sqrt(r.a * r.b)
			dOut := (0.98 + 0.5/rr - e) * rr
			dIn := (e - (0.98 - 3.1/rr)) * rr
			return p.Y >= T+4.2 && p.Y-(T+4.2) <= math.Min(dOut, dIn)
		}}, 1.0))
	})
	m.church(g, r.c.Add(P(0, 0)), T, 4.6, 2.6, 3.4)
	m.balcony(g, P(-34.5, 7.5), P(-28.5, 7.5), P(0, 1), T+2, 1.8)
	m.balcony(g, P(-26, 1.5), P(-26, 6), P(1, 0), T-1, 1.6)
}

func (m Meteora) holyTrinity(g *Grid) {
	r := meteoraRocks[3]
	T := r.top
	m.paving(g, r)
	m.church(g, P(21.5, 1), T, 7, 3.6, 3.8)
	m.cells(g, 18.5, 4, 24.5, 6.5, T-5, T, 3.6, White, Terracotta, AxisX)
	m.balcony(g, P(19, 6.5), P(24, 6.5), P(0, 1), T+1.8, 1.6)
	m.bell(g, P(25.5, -1.5), T-3, T+6)
	m.cypress(g, P(18.5, -1.5), T, 3.4)
	// A stair cut round the shaft: a helical ledge with a parapet, two full turns.
	y0, y1 := r.foot+1.5, T-0.6
	pitch := (y1 - y0) / 2.15
	phi0 := 1.2
	g.With(Stone, func() {
		g.Fill(Func{Box{V(r.c.X-r.a*1.6, y0-1, r.c.Y-r.b*1.6), V(r.c.X+r.a*1.6, y1+1.2, r.c.Y+r.b*1.6)}, func(p Vec3) bool {
			d := p.Plan().Sub(r.c.Add(r.lean.Scale(r.top - p.Y)))
			phi := math.Atan2(d.Y/r.b, d.X/r.a)
			turn := math.Mod(math.Mod(phi-phi0, 2*math.Pi)+2*math.Pi, 2*math.Pi) / (2 * math.Pi)
			k := math.Floor((p.Y - y0 - turn*pitch + 1.0) / pitch)
			yh := y0 + (k+turn)*pitch
			if yh < y0 || yh > y1 {
				return false
			}
			// Steps: quantize the tread height.
			yh = math.Floor(yh/0.5) * 0.5
			lim := r.limit(V(p.X, yh, p.Z))
			e := r.e3(p)
			rr := math.Min(r.a, r.b)
			out := (e - lim) * rr
			if p.Y >= yh-0.9 && p.Y <= yh && out >= -0.6 && out <= 1.7 {
				return true
			}
			return p.Y > yh && p.Y <= yh+1.1 && out >= 1.0 && out <= 1.7
		}})
	})
	g.Paint(Brick, Func{Box{V(r.c.X-r.a*1.6, y0-1, r.c.Y-r.b*1.6), V(r.c.X+r.a*1.6, y1+1.2, r.c.Y+r.b*1.6)}, func(p Vec3) bool {
		d := p.Plan().Sub(r.c.Add(r.lean.Scale(r.top - p.Y)))
		phi := math.Atan2(d.Y/r.b, d.X/r.a)
		turn := math.Mod(math.Mod(phi-phi0, 2*math.Pi)+2*math.Pi, 2*math.Pi) / (2 * math.Pi)
		k := math.Floor((p.Y - y0 - turn*pitch + 1.0) / pitch)
		yh := math.Floor((y0+(k+turn)*pitch)/0.5) * 0.5
		out := (r.e3(p) - r.limit(V(p.X, yh, p.Z))) * math.Min(r.a, r.b)
		return p.Y > yh && p.Y <= yh+1.2 && out >= 0.9
	}})
}

func (m Meteora) stStephen(g *Grid) {
	r, p5 := meteoraRocks[4], meteoraRocks[5]
	T := r.top
	m.paving(g, r)
	m.paving(g, p5)
	m.church(g, P(39, -20.5), T, 7, 3.8, 3.8)
	m.cells(g, 34, -16, 40.5, -13.8, T-6, T, 4, White, Terracotta, AxisX)
	m.balcony(g, P(34.5, -13.8), P(40, -13.8), P(0, 1), T+2, 1.7)
	// Stone bridge from the chapel pinnacle to St. Stephen's west door.
	dir := r.c.Sub(p5.c)
	dir = dir.Scale(1 / dir.Length())
	side := P(-dir.Y, dir.X)
	a := p5.c.Add(P(dir.X*p5.a, dir.Y*p5.b).Scale(0.55))
	b := r.c.Add(P(-dir.X*r.a, -dir.Y*r.b).Scale(0.6))
	length := b.Sub(a).Length()
	deck := func(s float64) float64 { return p5.top + (T-p5.top)*meteoraSmooth(s/length) }
	bb := BoxOf(a.At(p5.top-9), b.At(T+1.2)).Grow(1.6)
	frame := func(p Vec3) (float64, float64) {
		d := p.Plan().Sub(a)
		return d.Dot(dir), d.Dot(side)
	}
	g.With(Stone, func() {
		g.Fill(Func{bb, func(p Vec3) bool {
			s, w := frame(p)
			if s < 0 || s > length || math.Abs(w) > 1.4 {
				return false
			}
			top := deck(s)
			if p.Y > top+1.1 {
				return false
			}
			if p.Y > top {
				return math.Abs(w) > 0.7
			}
			// The arch: thin at the crown, deepening to the springers on the rock.
			u := (s - length/2) / (length/2 - 0.6)
			if math.Abs(u) >= 1 {
				return p.Y >= top-8
			}
			return p.Y >= top-1.2-5*(1-math.Sqrt(1-u*u))
		}})
	})
	g.Paint(Brick, Func{bb, func(p Vec3) bool {
		s, w := frame(p)
		return s >= 0 && s <= length && math.Abs(w) <= 1.5 && p.Y > deck(s)-0.9 && p.Y <= deck(s)+0.1
	}})
	// Chapel on the pinnacle.
	m.church(g, P(22.4, -26.2), p5.top, 4, 2.6, 3)
	m.cypress(g, P(24.4, -28), p5.top, 3.2)
}

// village is a cluster of whitewashed houses with tiled roofs and a church.
func (m Meteora) village(g *Grid, mp *meteoraMap) {
	for i := -8; i <= 8; i++ {
		for j := 0; j <= 4; j++ {
			at := P(float64(i)*4.1+1.2*(Hash01(i, j, 80)-0.5), 4+float64(j)*3.5+0.8*(Hash01(i, j, 81)-0.5))
			if Hash01(i, j, 82) < 0.12 || m.roadDist(at) < 2.4 || math.Abs(at.X+2) > 27-float64(j) {
				continue
			}
			if math.Abs(at.X+2) < 4 && math.Abs(at.Y-17.5) < 4 {
				continue // Church square.
			}
			h0 := mp.at(at.X, at.Y)
			if h0 > 5.5 || h0 <= 0 || at.Y > meteoraRiver(at.X)-4.4 {
				continue
			}
			w := 2.8 + 1.2*Hash01(i, j, 83)
			d := 2.8 + 1.2*Hash01(i, j, 84)
			hh := 2.6 + 1.6*math.Floor(2*Hash01(i, j, 85))
			base := h0
			for _, q := range []Vec2{P(-w/2, -d/2), P(w/2, -d/2), P(-w/2, d/2), P(w/2, d/2)} {
				base = math.Min(base, mp.at(at.X+q.X, at.Y+q.Y))
			}
			nearRock := false
			for _, r := range meteoraRocks {
				nearRock = nearRock || r.eFoot(at) < 1.35
			}
			if nearRock {
				continue
			}
			lo := V(at.X-w/2, base-0.8, at.Y-d/2)
			hi := V(at.X+w/2, h0+hh, at.Y+d/2)
			wall := White
			if Hash01(i, j, 86) < 0.25 {
				wall = Stone
			}
			g.With(wall, func() { g.Fill(Block(lo, hi)) })
			// Door and windows on the street side (+Z).
			g.Paint(Wood, Block(V(at.X-0.4, h0, hi.Z-0.2), V(at.X+0.4, h0+1.6, hi.Z+0.3)))
			g.Paint(DarkStone, Block(V(lo.X+0.5, h0+hh-1.5, hi.Z-0.2), V(lo.X+1.2, h0+hh-0.6, hi.Z+0.3)))
			roof := Terracotta
			if Hash01(i, j, 87) < 0.3 {
				roof = Brick
			}
			g.With(roof, func() {
				if Hash01(i, j, 88) < 0.5 {
					g.Fill(HipRoof{Min: P(lo.X-0.4, lo.Z-0.4), Max: P(hi.X+0.4, hi.Z+0.4), Base: hi.Y})
				} else {
					g.Fill(Gable{Min: P(lo.X-0.4, lo.Z-0.4), Max: P(hi.X+0.4, hi.Z+0.4), Base: hi.Y, Ridge: AxisX})
				}
			})
		}
	}
	// Village church on its square.
	c := P(-2, 17)
	f := mp.at(c.X, c.Y) - 0.4
	m.church(g, c, f+0.4, 7, 3.6, 3.6)
	g.Paint(White, Block(V(c.X-4.5, f-3, c.Y-3.5), V(c.X+5.5, f+4, c.Y+3.5)))
	m.bell(g, P(-6.8, 14.5), f-1, f+7)
	m.cypress(g, P(3, 15), f, 4)
	m.cypress(g, P(-7.5, 19.5), f, 3.6)
}

// trees scatters olive groves over the valley and woods over the scree mounds.
func (m Meteora) trees(g *Grid, mp *meteoraMap) {
	for i := -12; i <= 12; i++ {
		for j := -9; j <= 9; j++ {
			at := P(float64(i)*4.2+1.4*(Hash01(i, j, 90)-0.5), float64(j)*4.2+1.4*(Hash01(i, j, 91)-0.5))
			h := mp.at(at.X, at.Y)
			if h <= 0 || m.rim(at.X, at.Y) > 0.9 || m.roadDist(at) < 2.4 || math.Abs(at.Y-meteoraRiver(at.X)) < 4 {
				continue
			}
			near := false
			for _, r := range meteoraRocks {
				if r.eFoot(at) < 1.25 {
					near = true
				}
			}
			if near || mp.low(at.X, at.Y) < h-1.6 {
				continue
			}
			inVillage := at.X > -24 && at.X < 30 && at.Y > 4 && at.Y < 20
			if h > 3.8 {
				// Woods on the scree.
				if Hash01(i, j, 92) < 0.35 {
					continue
				}
				rad := 1.5 + 0.9*Hash01(i, j, 93)
				g.With(Wood, func() { g.Fill(Cylinder{Center: at, Radius: 0.5, Y0: h - 0.5, Y1: h + 1.4}) })
				c := Foliage
				if Hash01(i, j, 94) < 0.3 {
					c = Grass
				}
				g.With(c, func() { g.Fill(Ellipsoid{at.At(h + 1.4 + rad*0.7), V(rad, rad*0.9, rad)}) })
				continue
			}
			if inVillage || Hash01(i, j, 95) < 0.25 {
				continue
			}
			// Olive tree: short gnarled trunk, wide low crown.
			g.With(Wood, func() { g.Fill(Cylinder{Center: at, Radius: 0.5, Y0: h - 0.5, Y1: h + 1.2}) })
			g.With(Foliage, func() { g.Fill(Ellipsoid{at.At(h + 2), V(1.5, 1.0, 1.5)}) })
		}
	}
}
