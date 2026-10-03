package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// PontDuGard is the three-tiered Roman aqueduct bridge over the Gardon: six broad
// arches in the lower tier (the widest spanning the river), eleven in the middle tier
// and thirty-five small arches carrying the covered water channel on top. The 1743
// road bridge hugs the lower tier's downstream (+Z) side. The bridge runs along X;
// the green river flows along Z through a limestone gorge with garrigue slopes,
// holm oaks and cypresses, and a pebble beach on the upstream right bank.
type PontDuGard struct{}

func init() { Register(PontDuGard{}) }

func (PontDuGard) Name() string  { return "pont_du_gard" }
func (PontDuGard) Title() string { return "Pont Du Gard" }

const (
	pontDuGardScale = 0.65                                 // World units per metre.
	pontDuGardWater = 2.0                                  // River surface.
	pontDuGardLow   = 2.6 + 21.9*pontDuGardScale           // Top of the lower tier and road deck.
	pontDuGardMid   = pontDuGardLow + 19.5*pontDuGardScale // Top of the middle tier.
	pontDuGardTop   = pontDuGardMid + 7.4*pontDuGardScale  // Top of the channel's cover slabs.
	// Tier depths are slightly exaggerated so the thinnest tier stays three cells deep.
	pontDuGardDL    = 6.6 * pontDuGardScale // Roman lower tier, centred on z = 0.
	pontDuGardRoad  = 5.3 * pontDuGardScale // Road bridge on the downstream (+Z) side.
	pontDuGardDM    = 4.6 * pontDuGardScale
	pontDuGardDT    = 3.6 * pontDuGardScale
	pontDuGardHalfZ = 40 * pontDuGardScale // Half depth of the terrain plot.
)

// pontDuGardBay is one arch: opening centre and half width, in world units.
type pontDuGardBay struct{ c, hw float64 }

// pontDuGardPlan is the bridge layout along X plus the terrain heightfield.
type pontDuGardPlan struct {
	lower, middle, top           []pontDuGardBay
	xL0, xL1, xM0, xM1, xT0, xT1 float64
	m                            *pontDuGardMap
}

// pontDuGardMap is a sampled terrain heightfield with bilinear lookup.
type pontDuGardMap struct {
	x0, z0, step float64
	nx, nz       int
	h            []float64
}

func (m *pontDuGardMap) at(x, z float64) float64 {
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

// low is the lowest terrain around a point, used to close slopes into shells.
func (m *pontDuGardMap) low(x, z float64) float64 {
	lo := m.at(x, z)
	for k := 0; k < 8; k++ {
		a := float64(k) * math.Pi / 4
		lo = math.Min(lo, m.at(x+0.75*math.Cos(a), z+0.75*math.Sin(a)))
	}
	return lo
}

// noise is smooth value noise in [0, 1).
func (PontDuGard) noise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

// fbm is three octaves of value noise in [0, 1).
func (s PontDuGard) fbm(x, z float64, seed int) float64 {
	return (s.noise(x, z, seed)*4 + s.noise(x*2.1, z*2.1, seed+1)*2 + s.noise(x*4.3, z*4.3, seed+2)) / 7
}

func (PontDuGard) smooth(a, b, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-a)/(b-a)))
	return t * t * (3 - 2*t)
}

// plan lays out the arches from real proportions (metres), centred on the top tier.
func (PontDuGard) plan() *pontDuGardPlan {
	const k = pontDuGardScale
	spans := []float64{15.5, 19.2, 24.5, 21.0, 17.6, 15.5}
	const pier, extra = 5.0, 16.0
	var centres []float64
	x := 0.0
	for i, s := range spans {
		centres = append(centres, x+s/2)
		x += s
		if i < len(spans)-1 {
			x += pier
		}
	}
	end := x
	type arch struct{ c, s float64 }
	var mid []arch
	for j := 1; j >= 0; j-- {
		mid = append(mid, arch{-pier - extra/2 - float64(j)*(extra+pier), extra})
	}
	for i, s := range spans {
		mid = append(mid, arch{centres[i], s})
	}
	for j := 0; j < 3; j++ {
		mid = append(mid, arch{end + pier + extra/2 + float64(j)*(extra+pier), extra})
	}
	m0 := mid[0].c - extra/2 - pier
	m1 := mid[len(mid)-1].c + extra/2 + pier
	shift := (m0 + m1) / 2
	u := func(xm float64) float64 { return (xm - shift) * k }
	p := &pontDuGardPlan{}
	for i, s := range spans {
		p.lower = append(p.lower, pontDuGardBay{u(centres[i]), s * k / 2})
	}
	for _, a := range mid {
		p.middle = append(p.middle, pontDuGardBay{u(a.c), a.s * k / 2})
	}
	const length, arches = 275.0, 35
	bay := length / arches
	t0 := shift - length/2
	for i := 0; i < arches; i++ {
		p.top = append(p.top, pontDuGardBay{u(t0 + (float64(i)+0.5)*bay), 2.3 * k})
	}
	p.xL0, p.xL1 = u(-6), u(end+6)
	p.xM0, p.xM1 = u(m0), u(m1)
	p.xT0, p.xT1 = u(t0), u(t0+length)
	return p
}

// profile is the valley cross-section along the bridge axis: the river channel under
// the great arch, rock shelves, and hillsides rising to each tier's ends.
func (s PontDuGard) profile(p *pontDuGardPlan, x float64) float64 {
	r := p.lower[2]
	type pt struct{ x, y float64 }
	pts := []pt{
		{p.xT0 - 40, pontDuGardMid + 3.2},
		{p.xT0 - 3, pontDuGardMid + 2.6},
		{p.xT0 + 5, pontDuGardMid + 1.2},
		{p.xM0 + 1.5, pontDuGardMid - 0.8},
		{p.xL0 + 1.5, pontDuGardLow - 0.4},
		{p.lower[0].c, pontDuGardLow - 7.5},
		{r.c - r.hw - 0.4, 3.4},
		{r.c - r.hw + 1.6, 0.8},
		{r.c + r.hw + 3.5, 0.8},
		{r.c + r.hw + 6.5, 2.7},
		{p.lower[4].c, 3.2},
		{p.lower[5].c, 5.6},
		{p.xL1 - 1.5, pontDuGardLow - 0.4},
		{p.xM1 - 1.5, pontDuGardMid - 0.8},
		{p.xT1 - 5, pontDuGardMid + 1.2},
		{p.xT1 + 3, pontDuGardMid + 2.6},
		{p.xT1 + 40, pontDuGardMid + 3.2},
	}
	if x <= pts[0].x {
		return pts[0].y
	}
	for i := 1; i < len(pts); i++ {
		if x <= pts[i].x {
			a, b := pts[i-1], pts[i]
			t := (x - a.x) / (b.x - a.x)
			return a.y + (b.y-a.y)*t
		}
	}
	return pts[len(pts)-1].y
}

// rim is the signed distance inside the plot's rounded rectangle (negative outside).
func (s PontDuGard) rim(p *pontDuGardPlan, x, z float64) float64 {
	cx := (p.xT0 + p.xT1) / 2
	hx := (p.xT1-p.xT0)/2 + 7
	const r = 8.0
	dx := math.Max(math.Abs(x-cx)-(hx-r), 0)
	dz := math.Max(math.Abs(z)-(pontDuGardHalfZ-r), 0)
	return r - math.Hypot(dx, dz) + 2.5*(s.fbm(x/7, z/7, 21)-0.5)
}

// bend shifts the valley sideways away from the bridge, so the river meanders.
func (s PontDuGard) bend(z float64) float64 {
	return 7 * math.Sin(z/15+0.5) * s.smooth(5, 16, math.Abs(z))
}

// height is the raw terrain; zero means outside the plot.
func (s PontDuGard) height(p *pontDuGardPlan, x, z float64) float64 {
	rim := s.rim(p, x, z)
	if rim <= 0 {
		return 0
	}
	// The gorge opens out away from the bridge, so its cut faces frame the arches.
	rc := p.lower[2].c
	w := 1 + 0.5*s.smooth(5, 22, math.Abs(z))
	h := s.profile(p, rc+(x-s.bend(z)-rc)/w)
	// Rolling garrigue and broken limestone ledges, calm under the bridge itself.
	calm := (0.35 + 0.65*s.smooth(3, 10, math.Abs(z))) * (0.3 + 0.7*s.smooth(pontDuGardWater, pontDuGardWater+3, h))
	h += calm * (3.2*(s.fbm(x/11, z/11, 1)-0.5) + 1.2*math.Floor(2*s.fbm(x/6, z/6, 2)))
	if h > pontDuGardWater+0.3 {
		h = math.Max(h, pontDuGardWater+0.3)
	}
	h = math.Min(h, 0.5+3.5*rim)
	return math.Max(h, 0.4)
}

func (s PontDuGard) Build(g *Grid) Order {
	p := s.plan()
	m := &pontDuGardMap{x0: p.xT0 - 16, z0: -pontDuGardHalfZ - 2, step: 0.4}
	m.nx = int((p.xT1-p.xT0+32)/m.step) + 2
	m.nz = int((2*pontDuGardHalfZ+4)/m.step) + 2
	raw := make([]float64, m.nx*m.nz)
	for j := 0; j < m.nz; j++ {
		for i := 0; i < m.nx; i++ {
			raw[j*m.nx+i] = s.height(p, m.x0+float64(i)*m.step, m.z0+float64(j)*m.step)
		}
	}
	// Two box-blur passes soften the piecewise profile; the plot outline stays sharp.
	for pass := 0; pass < 2; pass++ {
		next := make([]float64, len(raw))
		for j := 0; j < m.nz; j++ {
			for i := 0; i < m.nx; i++ {
				k := j*m.nx + i
				if raw[k] <= 0 || i == 0 || j == 0 || i == m.nx-1 || j == m.nz-1 {
					next[k] = raw[k]
					continue
				}
				sum, n := 0.0, 0.0
				for dj := -1; dj <= 1; dj++ {
					for di := -1; di <= 1; di++ {
						if v := raw[k+dj*m.nx+di]; v > 0 {
							sum += v
							n++
						}
					}
				}
				next[k] = sum / n
			}
		}
		raw = next
	}
	m.h = raw
	p.m = m
	s.terrain(g, p)
	s.aqueduct(g, p)
	s.trees(g, p)
	return Along(P(1, 0))
}

// terrain fills the heightfield as a hollow shell: limestone strata, garrigue on the
// gentle slopes, pale rock shelves and a beach by the river, and the green water.
func (s PontDuGard) terrain(g *Grid, p *pontDuGardPlan) {
	m := p.m
	box := Box{V(m.x0, 0, m.z0), V(m.x0+float64(m.nx-1)*m.step, 40, m.z0+float64(m.nz-1)*m.step)}
	g.Use(Stone).Fill(Func{box, func(q Vec3) bool {
		h := m.at(q.X, q.Z)
		if h <= 0 {
			return false
		}
		return q.Y <= h && q.Y >= math.Min(h, m.low(q.X, q.Z))-0.6
	}})
	// Weathered strata on the cliffs.
	// Limestone strata: pale, grey and dark beds, gently warped.
	stratum := func(q Vec3) int {
		return int(math.Floor(q.Y/1.1 + 0.9*s.noise(q.X/13, q.Z/13, 11)))
	}
	g.Paint(DarkStone, Func{box, func(q Vec3) bool { return Hash01(stratum(q), 12) < 0.3 }})
	// Vegetation stays off the plot's cut faces.
	gentle := func(q Vec3, h float64) bool {
		return q.Y >= h-1.2 && h-m.low(q.X, q.Z) < 1.7 && 0.5+3.5*s.rim(p, q.X, q.Z) > h+1
	}
	// Garrigue: dry grass with dense green scrub, bare limestone outcrops.
	g.Paint(Grass, Func{box, func(q Vec3) bool {
		h := m.at(q.X, q.Z)
		return h > pontDuGardWater+2.5 && gentle(q, h)
	}})
	g.Paint(Foliage, Func{box, func(q Vec3) bool {
		h := m.at(q.X, q.Z)
		return h > pontDuGardWater+3 && q.Y >= h-1.2 && h-m.low(q.X, q.Z) < 2.2 && s.fbm(q.X/8, q.Z/8, 13) > 0.5 &&
			0.5+3.5*s.rim(p, q.X, q.Z) > h+1
	}})
	g.Paint(Stone, Func{box, func(q Vec3) bool {
		h := m.at(q.X, q.Z)
		return q.Y >= h-1.2 && s.fbm(q.X/5, q.Z/5, 14) > 0.66
	}})
	// Pale limestone shelves along the river, wet and dark at the waterline.
	g.Paint(Stone, Func{box, func(q Vec3) bool {
		return m.at(q.X, q.Z) < pontDuGardWater+2.4 && q.Y > pontDuGardWater-0.5
	}})
	// Tufts of scrub and grass that root in the cracks of the shelves.
	g.Paint(Foliage, Func{box, func(q Vec3) bool {
		h := m.at(q.X, q.Z)
		return h > pontDuGardWater+1 && h < pontDuGardWater+6 && q.Y >= h-1 && h-m.low(q.X, q.Z) < 1.7 &&
			s.rim(p, q.X, q.Z) > 2 && s.fbm(q.X/3.5, q.Z/3.5, 16) > 0.64
	}})
	g.Paint(Grass, Func{box, func(q Vec3) bool {
		h := m.at(q.X, q.Z)
		v := s.fbm(q.X/3.5, q.Z/3.5, 16)
		return h > pontDuGardWater+1 && h < pontDuGardWater+6 && q.Y >= h-1 && h-m.low(q.X, q.Z) < 1.7 &&
			s.rim(p, q.X, q.Z) > 2 && v > 0.58 && v <= 0.64
	}})
	g.Paint(DarkStone, Func{box, func(q Vec3) bool {
		return q.Y > pontDuGardWater-0.5 && q.Y < pontDuGardWater+0.6 && Hash01(int(math.Floor(q.X)), int(math.Floor(q.Z)), 15) < 0.6
	}})
	// Pebble beach on the upstream right bank, a smaller one downstream on the left.
	r := p.lower[2]
	g.Paint(Sand, Func{box, func(q Vec3) bool {
		x := q.X - s.bend(q.Z)
		h := m.at(q.X, q.Z)
		if h > pontDuGardWater+1.8 || q.Y < pontDuGardWater-0.2 {
			return false
		}
		up := q.Z < -7 && q.Z > -26 && x > r.c+r.hw-2 && x < r.c+r.hw+9
		down := q.Z > 12 && q.Z < 22 && x < r.c-r.hw+3 && x > r.c-r.hw-6
		return up || down
	}})
	// The river: green-blue water, paler where shallow.
	water := Func{box, func(q Vec3) bool {
		h := m.at(q.X, q.Z)
		return h > 0 && h < pontDuGardWater-0.15 && q.Y <= pontDuGardWater && q.Y >= pontDuGardWater-1.0 && s.rim(p, q.X, q.Z) > 0.6
	}}
	g.Use(Water).Fill(water)
	g.Paint(Glass, Intersect(water, Func{box, func(q Vec3) bool {
		return m.at(q.X, q.Z) > pontDuGardWater-0.45 && q.Y >= pontDuGardWater-0.5
	}}))
	// Tracks from the road bridge up both banks.
	zc := pontDuGardDL/2 + pontDuGardRoad/2
	for _, seg := range [][2]Vec2{
		{P(p.xL0+1.5, zc), P(p.xL0-7, pontDuGardHalfZ)},
		{P(p.xL1-1.5, zc), P(p.xL1+5, pontDuGardHalfZ)},
	} {
		a, b := seg[0], seg[1]
		g.Paint(Stone, Func{box, func(q Vec3) bool {
			ab := b.Sub(a)
			t := math.Max(0, math.Min(1, q.Plan().Sub(a).Dot(ab)/ab.Dot(ab)))
			return q.Plan().Sub(a.Add(ab.Scale(t))).Length() < 1.1 && q.Y >= m.at(q.X, q.Z)-1.0
		}})
	}
	// Canoes on the river.
	for i, z := range []float64{-15, -9.5, 11, 17.5} {
		x := r.c + s.bend(z) + 3*(Hash01(i, 34)-0.5)
		d := Polar(P(0, 0), 1.3, math.Pi/2+0.5*(Hash01(i, 35)-0.5))
		color := Red
		if i%2 == 1 {
			color = Gold
		}
		g.With(color, func() {
			g.Fill(Capsule{V(x-d.X, pontDuGardWater+0.3, z-d.Y), V(x+d.X, pontDuGardWater+0.3, z+d.Y), 0.55})
		})
	}
	// Boulders strewn in the shallows and on the shelves.
	for i := 0; i < 14; i++ {
		z := -26 + 52*Hash01(i, 31)
		x := r.c + s.bend(z) + (2*Hash01(i, 32)-1)*(r.hw+3)
		if math.Abs(z) < 4 {
			continue
		}
		rad := 0.9 + 0.8*Hash01(i, 33)
		y := math.Max(m.at(x, z), pontDuGardWater-0.6)
		g.With(Stone, func() { g.FillVolume(Ellipsoid{V(x, y, z), V(rad*1.3, rad*0.8, rad)}) })
	}
}

// aqueduct builds the three tiers and the road bridge in a separate grid, colors the
// masonry block by block, then merges it into the scene.
func (s PontDuGard) aqueduct(g *Grid, p *pontDuGardPlan) {
	a := NewGridWithSpan(g.Span())
	ground := Func{Box{V(p.xT0-3, -2, -12), V(p.xT1+3, 40, 12)}, func(q Vec3) bool {
		return q.Y >= p.m.at(q.X, q.Z)-1.0
	}}
	clip := func(solids ...Solid) Solid { return Intersect(Union(solids...), ground) }
	const dl, dm, dt = pontDuGardDL / 2, pontDuGardDM / 2, pontDuGardDT / 2
	zr := dl + pontDuGardRoad
	opening := func(b pontDuGardBay, crown float64) Arch {
		return Arch{Sill: V(b.c, -2, 0), HalfWidth: b.hw, Spring: crown - b.hw + 2, Depth: 24, Through: AxisZ}
	}
	crownL, crownM, crownT := pontDuGardLow-1.5, pontDuGardMid-1.3, pontDuGardMid+2.7
	a.Use(Sand)

	// Lower tier with the road bridge, cutwaters, imposts and cornices.
	var holes []Solid
	for _, b := range p.lower {
		holes = append(holes, opening(b, crownL))
	}
	a.Fill(clip(Subtract(Block(V(p.xL0, -1.5, -dl), V(p.xL1, pontDuGardLow, zr)), holes...)))
	for i := 0; i+1 < len(p.lower); i++ {
		l, r := p.lower[i].c+p.lower[i].hw, p.lower[i+1].c-p.lower[i+1].hw
		c := (l + r) / 2
		cut := Prism{Polygon: []Vec2{P(l, -dl+0.2), P(r, -dl+0.2), P(c, -dl-2.2)}, Y0: -1.5, Y1: pontDuGardWater + 4.5}
		nose := Func{Box{V(l, pontDuGardWater+4.5, -dl-2.2), V(r, pontDuGardWater+6.5, -dl)}, func(q Vec3) bool {
			return InsidePolygon(cut.Polygon, q.Plan()) && q.Y-(pontDuGardWater+4.5) <= (q.Z+dl+2.2)*0.9
		}}
		a.Fill(clip(cut, nose))
	}
	for _, b := range p.lower {
		spring := crownL - b.hw
		a.Fill(clip(
			Block(V(b.c-b.hw-1.4, spring-0.7, -dl-0.5), V(b.c-b.hw+0.4, spring, zr+0.5)),
			Block(V(b.c+b.hw-0.4, spring-0.7, -dl-0.5), V(b.c+b.hw+1.4, spring, zr+0.5))))
		s.studs(a, p, b, spring, -dl, zr, ground)
	}
	a.Fill(clip(
		Block(V(p.xL0, pontDuGardLow-0.8, -dl-0.6), V(p.xL1, pontDuGardLow, -dl)),
		Block(V(p.xL0, pontDuGardLow-0.8, zr), V(p.xL1, pontDuGardLow, zr+0.6)),
		// Road parapet.
		Block(V(p.xL0, pontDuGardLow, zr-1.0), V(p.xL1, pontDuGardLow+1.1, zr+0.4))))

	// Middle tier.
	holes = nil
	for _, b := range p.middle {
		holes = append(holes, opening(b, crownM))
	}
	body := Union(
		Block(V(p.xL0, pontDuGardLow-0.3, -dm), V(p.xL1, pontDuGardMid, dm)),
		Block(V(p.xM0, -1.5, -dm), V(p.xL0, pontDuGardMid, dm)),
		Block(V(p.xL1, -1.5, -dm), V(p.xM1, pontDuGardMid, dm)))
	a.Fill(clip(Subtract(body, holes...)))
	for _, b := range p.middle {
		spring := crownM - b.hw
		a.Fill(clip(
			Block(V(b.c-b.hw-1.2, spring-0.6, -dm-0.5), V(b.c-b.hw+0.4, spring, dm+0.5)),
			Block(V(b.c+b.hw-0.4, spring-0.6, -dm-0.5), V(b.c+b.hw+1.2, spring, dm+0.5))))
		s.studs(a, p, b, spring, -dm, dm, ground)
	}
	a.Fill(clip(Block(V(p.xM0, pontDuGardMid-0.7, -dm-0.6), V(p.xM1, pontDuGardMid, dm+0.6))))

	// Top tier: small arches under the covered channel.
	holes = nil
	for _, b := range p.top {
		holes = append(holes, opening(b, crownT))
	}
	body = Union(
		Block(V(p.xM0, pontDuGardMid-0.3, -dt), V(p.xM1, pontDuGardTop-0.6, dt)),
		Block(V(p.xT0, -1.5, -dt), V(p.xM0, pontDuGardTop-0.6, dt)),
		Block(V(p.xM1, -1.5, -dt), V(p.xT1, pontDuGardTop-0.6, dt)))
	a.Fill(clip(Subtract(body, holes...)))
	a.Fill(clip(Block(V(p.xT0, pontDuGardTop-0.7, -dt-0.5), V(p.xT1, pontDuGardTop, dt+0.5))))

	s.masonry(a, p)
	g.Merge(a)
}

// studs adds the protruding scaffold stones on the pier faces either side of an arch.
func (PontDuGard) studs(a *Grid, p *pontDuGardPlan, b pontDuGardBay, spring, z0, z1 float64, ground Solid) {
	for _, side := range []float64{-1, 1} {
		x := b.c + side*(b.hw+0.9)
		for k, y := 0, spring-2.2; k < 4 && y > p.m.at(x, 0)+1; k, y = k+1, y-2.8 {
			if Hash01(int(x*10), k, 41) < 0.35 {
				continue
			}
			a.Fill(Intersect(Union(
				Block(V(x-0.4, y-0.4, z0-0.7), V(x+0.4, y+0.4, z0)),
				Block(V(x-0.4, y-0.4, z1), V(x+0.4, y+0.4, z1+0.7))), ground))
		}
	}
}

// masonry recolors the bridge block by block: honey sandstone with paler and darker
// blocks, golden patches, archivolt rings, flood staining, paving and cover slabs.
func (s PontDuGard) masonry(a *Grid, p *pontDuGardPlan) {
	const dl = pontDuGardDL / 2
	ring := func(q Vec3, bays []pontDuGardBay, crown float64) (int, bool) {
		for _, b := range bays {
			spring := crown - b.hw
			d := math.Hypot(q.X-b.c, q.Y-spring)
			if q.Y >= spring && d >= b.hw-0.1 && d < b.hw+1.0 {
				seg := int(math.Floor(math.Atan2(q.Y-spring, q.X-b.c) * (b.hw + 0.5) / 1.1))
				return seg, true
			}
		}
		return 0, false
	}
	for _, c := range a.Cells() {
		q := a.World(c)
		course := int(math.Floor(q.Y / 0.8))
		stagger := 0.0
		if course%2 != 0 {
			stagger = 0.9
		}
		bx := int(math.Floor((q.X + stagger) / 1.8))
		bz := int(math.Floor((q.Z + stagger) / 1.8))
		h := Hash01(course, bx, bz, 7)
		color := Sand
		road := q.Z > dl+0.1 && q.Y < pontDuGardLow+1.5
		switch {
		case h < 0.14:
			color = Stone
		case h < 0.18:
			color = DarkStone
		}
		if road && h < 0.4 {
			color = Stone
		}
		if seg, ok := ring(q, p.lower, pontDuGardLow-1.5); ok && q.Y < pontDuGardLow {
			color = Sand
			if seg%2 == 0 {
				color = Stone
			}
		}
		if seg, ok := ring(q, p.middle, pontDuGardMid-1.3); ok && q.Y > pontDuGardLow && q.Y < pontDuGardMid {
			color = Sand
			if seg%2 == 0 {
				color = Stone
			}
		}
		// Flood line and damp bases.
		if q.Y < pontDuGardWater+1.8+0.8*Hash01(bx, 43) {
			color = DarkStone
		}
		// Rain streaks below the cornices.
		for _, top := range []float64{pontDuGardLow, pontDuGardMid, pontDuGardTop} {
			if q.Y < top-0.8 && q.Y > top-3.2 && Hash01(int(math.Floor(q.X/0.9)), int(top), 44) < 0.1 {
				color = DarkStone
			}
		}
		// Road paving, and the channel's cover slabs.
		if q.Y > pontDuGardLow-0.5 && q.Y < pontDuGardLow+0.1 && q.Z > dl-1 && q.Z < dl+pontDuGardRoad-1 {
			color = Stone
			if Hash01(int(math.Floor(q.X/1.2)), int(math.Floor(q.Z)), 45) < 0.3 {
				color = DarkStone
			}
		}
		if q.Y > pontDuGardTop-0.7 {
			color = Sand
			if Hash01(int(math.Floor(q.X/1.3)), 46) < 0.2 {
				color = Stone
			}
		}
		a.Remove(c)
		a.Use(color).Add(c)
	}
}

// trees scatters holm oaks and cypresses over the garrigue, clear of the bridge,
// the river and the cliffs.
func (s PontDuGard) trees(g *Grid, p *pontDuGardPlan) {
	m := p.m
	for i := -32; i <= 32; i++ {
		for j := -10; j <= 10; j++ {
			at := P(float64(i)*3.2+2.6*(Hash01(i, j, 51)-0.5), float64(j)*3.2+2.6*(Hash01(i, j, 52)-0.5))
			if Hash01(i, j, 50) > 0.2+0.6*s.fbm(at.X/14, at.Y/14, 53) {
				continue
			}
			h := m.at(at.X, at.Y)
			if h < 9 || h-m.low(at.X, at.Y) > 1.0 || s.rim(p, at.X, at.Y) < 3 {
				continue
			}
			if math.Abs(at.Y) < 7.5 && at.X > p.xT0-4 && at.X < p.xT1+4 {
				continue
			}
			if Hash01(i, j, 54) < 0.3 {
				s.cypress(g, at, h-0.5, 4.5+2.5*Hash01(i, j, 55))
			} else {
				s.oak(g, at, h-0.5, 1.8+1.0*Hash01(i, j, 56), i*100+j)
			}
		}
	}
}

func (PontDuGard) cypress(g *Grid, at Vec2, base, h float64) {
	g.With(Wood, func() { g.Fill(Cylinder{Center: at, Radius: 0.5, Y0: base - 0.5, Y1: base + 1}) })
	g.With(Foliage, func() {
		g.FillVolume(Ellipsoid{at.At(base + 0.6 + h*0.45), V(1.0, h*0.5, 1.0)})
	})
}

func (PontDuGard) oak(g *Grid, at Vec2, base, r float64, seed int) {
	g.With(Wood, func() { g.Fill(Cylinder{Center: at, Radius: 0.6, Y0: base - 0.5, Y1: base + r*0.7}) })
	g.With(Foliage, func() {
		c := at.At(base + r*1.15)
		g.FillVolume(Subtract(Ellipsoid{c, V(r, r*0.8, r)}, Ellipsoid{c, V(r-1.1, r*0.8-1.1, r-1.1)}))
		a := 2 * math.Pi * Hash01(seed, 57)
		o := Polar(at, r*0.55, a).At(base + r*1.4)
		g.FillVolume(Subtract(Ellipsoid{o, V(r*0.7, r*0.6, r*0.7)}, Ellipsoid{o, V(r*0.7-1.1, r*0.6-1.1, r*0.7-1.1)}))
	})
}
