package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// TigersNest is Paro Taktsang, the Tiger's Nest monastery of Bhutan, clinging to a
// sheer granite cliff. The cliff faces +Z and rises from a forested slope; two thirds
// of the way up a deep alcove shelters the monastery: white temple buildings stacked on
// rock ledges at different levels, each with a red kemar band, dark wooden trapezoidal
// windows and projecting bays, wide-eaved shingle roofs and gilded roof pavilions with
// finials. Stairs and walkways link the levels. To the east (+X) a cleft splits the
// cliff: a waterfall plunges into a pool, then drops again to a stream on the slope. A
// wooden bridge hung with prayer flags crosses the pool to a lower forested ridge,
// where stairs climb past a chorten and flag poles to a teahouse viewpoint. The cliff
// top is crowned with pines; every outer edge steps down to a low rim.
type TigersNest struct{}

func init() { Register(TigersNest{}) }

func (TigersNest) Name() string  { return "tigers_nest" }
func (TigersNest) Title() string { return "Tigers Nest" }

const (
	tigersNestX0, tigersNestX1 = -46.0, 38.0
	tigersNestZ0, tigersNestZ1 = -32.0, 23.0
	tigersNestYMax             = 74.0
	tigersNestStep             = 0.25
	tigersNestSkin             = 1.1
	tigersNestNone             = -100.0
	tigersNestCleftW           = 13.0  // West wall of the cleft (monastery side).
	tigersNestCleftE           = 19.5  // East wall of the cleft (ridge side).
	tigersNestBack             = -16.0 // Front of the rock block behind the cleft.
	tigersNestLip              = -5.0  // Front lip of the waterfall pool.
	tigersNestPool             = 31.0
	tigersNestDeck             = 34.0
	tigersNestBridgeZ          = -8.0
)

func (TigersNest) smooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func (TigersNest) noise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

func (s TigersNest) fbm(x, z float64, seed int) float64 {
	return (s.noise(x, z, seed)*4 + s.noise(x*2.1+5, z*2.1+3, seed+1)*2 + s.noise(x*4.3+1, z*4.3+7, seed+2)) / 7
}

// steps is an eroded terrace profile d units in from a foot: risers every bench units.
func (TigersNest) steps(d, bench, riser, jitter float64) float64 {
	if d <= 0 {
		return 0
	}
	return riser * (1 + math.Floor(d/bench+jitter))
}

// rim limits every height near the site edges so the terrain steps down to a low rim.
// Sides and back may step at different rates: benchX applies to the side edges.
func (s TigersNest) rim(x, z, benchX, benchZ, riser float64) float64 {
	n := 2.5 * (s.fbm(x/6, z/6, 11) - 0.5)
	b := 0.5 * s.noise(x/5, z/5, 12)
	j := s.noise(x/3, z/3, 13)
	dx := math.Min(x-tigersNestX0, tigersNestX1-x) + n
	dz := math.Min(z-tigersNestZ0, tigersNestZ1-z) + n
	return 0.6 + math.Min(s.steps(dx-1.2, benchX+b, riser, j), s.steps(dz-1.2, benchZ+b, riser, j))
}

// slope is the forested ground rising from the front rim to the cliff foot.
func (s TigersNest) slope(x, z float64) float64 {
	t := s.smooth((tigersNestZ1 - 1 - z) / 25)
	return 1 + 14.5*t + 2.4*(s.fbm(x/9, z/9, 21)-0.5)
}

// ridge is the lower forested spur east of the cleft, with a terrace at the bridgehead.
func (s TigersNest) ridge(x, z float64) float64 {
	if x < tigersNestCleftE+0.8*(s.noise(z/2.5, 1, 22)-0.5) {
		return tigersNestNone
	}
	front := math.Max(0, z+8)
	e := 40 + 3*(s.fbm(x/7, z/7, 23)-0.5) - 0.05*(x-28)*(x-28) - 0.75*front - 0.02*front*front
	return math.Min(e, tigersNestDeck+1.9*math.Max(0, x-21.5)+0.6*(s.noise(x/2, z/2, 24)-0.5))
}

// cleft is the waterfall pool floor between the cliff and the ridge.
func (s TigersNest) cleft(x, z float64) float64 {
	if x <= tigersNestCleftW-0.5 || x >= tigersNestCleftE+0.5 || z > tigersNestLip || z < tigersNestBack-1 {
		return tigersNestNone
	}
	return tigersNestPool + 0.4*(s.noise(x/2, z/2, 25)-0.5)
}

// massif is the cliff block's top: the main cliff west of the cleft and the block behind.
func (s TigersNest) massif(x, z float64) float64 {
	front := x <= tigersNestCleftW+0.5*(s.noise(z/2.5, 3, 26)-0.5)
	back := z <= tigersNestBack+0.6*(s.noise(x/3, 5, 27)-0.5)
	if !front && !back {
		return tigersNestNone
	}
	t := 68 + 4*(s.fbm(x/10, z/10, 28)-0.5) + 1.5*(s.noise(x/3.5, z/3.5, 29)-0.5) - 2.1*math.Max(0, -z-20)
	if !front {
		t = math.Min(t, math.Min(57+2*(s.noise(x/3, z/4, 35)-0.5), 42+s.steps(31-x, 1.0, 5, s.noise(x/3, z/3, 30))))
	}
	if x > 14.2 && x < 17.8 && z > tigersNestBack-5 {
		t -= 2.5 // The waterfall's notch.
	}
	return t
}

// ledge is the alcove floor under each monastery building; ceiling is its rock roof.
func (TigersNest) ledge(x float64) float64 {
	switch {
	case x < -24:
		return 33
	case x < -13.5:
		return 39
	case x < -2.5:
		return 36
	case x < 8:
		return 33
	}
	return tigersNestDeck
}

func (s TigersNest) ceiling(x float64) float64 {
	return 62 - 0.04*(x+9)*(x+9) + 1.5*(s.noise(x/3, 7, 31)-0.5)
}

// alcove weighs how deep the monastery recess is at x.
func (s TigersNest) alcove(x float64) float64 {
	return s.smooth((x+31.5)/2.5) * s.smooth((tigersNestCleftW+0.5-x)/0.5)
}

// face is the cliff's front plane (z) at x and height y, with overhangs and the alcove.
func (s TigersNest) face(x, y float64) float64 {
	f := -3 + 1.8*(s.noise(x/4, y/11, 32)-0.5) + 0.9*(s.noise(x/1.5, y/5, 33)-0.5)
	f += 2.8 * s.smooth((y-52)/12)
	if x < -33 {
		f -= 0.3 * (x + 33) * (x + 33) // The cliff wraps back at its west end.
	}
	if a := s.alcove(x); a > 0 {
		lo, hi := s.ledge(x), s.ceiling(x)
		switch {
		case y >= lo && y <= hi:
			back := -12.3 + 0.8*(s.noise(x/2, y/3, 34)-0.5)
			f += (back - f) * a
		case y < lo && y > lo-4:
			f += 3.2 * a * s.smooth((y-(lo-4))/4)
		}
	}
	return f
}

// tigersNestMap is a sampled 2D field with bilinear lookup.
type tigersNestMap struct {
	x0, y0, step float64
	nx, ny       int
	v            []float64
}

func (TigersNest) sample(x0, y0, x1, y1 float64, f func(a, b float64) float64) *tigersNestMap {
	m := &tigersNestMap{x0: x0, y0: y0, step: tigersNestStep}
	m.nx = int(math.Ceil((x1-x0)/m.step)) + 1
	m.ny = int(math.Ceil((y1-y0)/m.step)) + 1
	m.v = make([]float64, m.nx*m.ny)
	for j := 0; j < m.ny; j++ {
		for i := 0; i < m.nx; i++ {
			m.v[j*m.nx+i] = f(x0+float64(i)*m.step, y0+float64(j)*m.step)
		}
	}
	return m
}

func (m *tigersNestMap) at(x, y float64) float64 {
	fx, fy := (x-m.x0)/m.step, (y-m.y0)/m.step
	if fx < 0 || fy < 0 || fx >= float64(m.nx-1) || fy >= float64(m.ny-1) {
		return tigersNestNone
	}
	i, j := int(fx), int(fy)
	tx, ty := fx-float64(i), fy-float64(j)
	k := j*m.nx + i
	a := m.v[k] + (m.v[k+1]-m.v[k])*tx
	b := m.v[k+m.nx] + (m.v[k+m.nx+1]-m.v[k+m.nx])*tx
	return a + (b-a)*ty
}

// tigersNestTerrain holds the sampled ground, massif top and cliff face.
type tigersNestTerrain struct {
	ground, top, face *tigersNestMap
}

func (t *tigersNestTerrain) solid(p Vec3) bool {
	if p.Y <= t.ground.at(p.X, p.Z) {
		return true
	}
	return p.Y <= t.top.at(p.X, p.Z) && p.Z < t.face.at(p.X, p.Y)
}

// surface is the walkable height at a plan point: the massif top or the ground.
func (t *tigersNestTerrain) surface(x, z float64) float64 {
	h := t.ground.at(x, z)
	if m := t.top.at(x, z); m > h && z < t.face.at(x, m)-0.5 {
		return m
	}
	return h
}

func (t *tigersNestTerrain) gradient(m *tigersNestMap, x, z float64) float64 {
	gx := (m.at(x+0.6, z) - m.at(x-0.6, z)) / 1.2
	gz := (m.at(x, z+0.6) - m.at(x, z-0.6)) / 1.2
	return math.Hypot(gx, gz)
}

func (s TigersNest) terrain() *tigersNestTerrain {
	ground := func(x, z float64) float64 {
		h := math.Max(math.Max(s.slope(x, z), s.ridge(x, z)), s.cleft(x, z))
		return math.Min(h, s.rim(x, z, 1.15, 1.15, 4.2))
	}
	top := func(x, z float64) float64 {
		m := s.massif(x, z)
		if m == tigersNestNone {
			return m
		}
		return math.Min(m, s.rim(x, z, 1.3, 0.75, 5))
	}
	return &tigersNestTerrain{
		ground: s.sample(tigersNestX0, tigersNestZ0, tigersNestX1, tigersNestZ1, ground),
		top:    s.sample(tigersNestX0, tigersNestZ0, tigersNestX1, tigersNestZ1, top),
		face:   s.sample(tigersNestX0, -1, tigersNestX1, tigersNestYMax, s.face),
	}
}

func (s TigersNest) Build(g *Grid) Order {
	t := s.terrain()
	s.rock(g, t)
	s.water(g, t)
	s.monastery(g, t)
	s.crossing(g, t)
	s.forest(g, t)
	return Sweep(P(-8, -6), math.Pi/2)
}

// rock fills the terrain as a shell and colors it: vegetation on gentle ground, granite
// streaks on the cliffs, bushes on ledges and shadowed rock inside the alcove.
func (s TigersNest) rock(g *Grid, t *tigersNestTerrain) {
	var dirs []Vec3
	for _, n := range Neighbors {
		v := V(float64(n.X), float64(n.Y), float64(n.Z))
		dirs = append(dirs, v.Scale(tigersNestSkin/v.Length()))
	}
	inside := func(p Vec3) bool { return p.Y < 0 || t.solid(p) }
	tmp := NewGridWithSpan(g.Span())
	box := Box{V(tigersNestX0, 0, tigersNestZ0), V(tigersNestX1, tigersNestYMax, tigersNestZ1)}
	tmp.Fill(Func{box, func(p Vec3) bool {
		if p.Y < 0 || !t.solid(p) {
			return false
		}
		for _, d := range dirs {
			if !inside(p.Add(d)) {
				return true
			}
		}
		return false
	}})
	up := Cell{0, 2, 0}
	for _, c := range tmp.Cells() {
		p := g.World(c)
		g.Use(s.rockColor(t, p, !tmp.Has(c.Add(up)), c)).Add(c)
	}
}

func (s TigersNest) rockColor(t *tigersNestTerrain, p Vec3, exposed bool, c Cell) Color {
	h := Hash01(int(c.X), int(c.Y), int(c.Z), 41)
	gh := t.ground.at(p.X, p.Z)
	streak := s.noise(p.X/1.4+p.Z*0.45, p.Y/16, 42)
	stone := Stone
	switch {
	case streak > 0.8:
		stone = Charcoal
	case streak > 0.6 || h < 0.12:
		stone = DarkStone
	}
	if p.Y <= gh+0.35 {
		// Ground: vegetation unless steep.
		if t.gradient(t.ground, p.X, p.Z) > 2.2 && p.Y < gh-0.6 {
			return stone
		}
		if s.fbm(p.X/4, p.Z/4, 43) > 0.55 || h < 0.15 {
			return Foliage
		}
		return Grass
	}
	top := t.top.at(p.X, p.Z)
	face := t.face.at(p.X, p.Y)
	if p.Y > top-1.5 && p.Z < face-1.0 && t.gradient(t.top, p.X, p.Z) < 1.7 {
		if s.fbm(p.X/4, p.Z/4, 44) > 0.5 || h < 0.2 {
			return Foliage
		}
		return Grass
	}
	if a := s.alcove(p.X); a > 0.5 && p.Z < -7 && p.Y >= s.ledge(p.X)-0.5 && p.Y < s.ceiling(p.X)+2 {
		if h < 0.45 {
			return Charcoal
		}
		return DarkStone
	}
	if exposed && h < 0.3 && p.Y < 60 {
		return Foliage
	}
	return stone
}

// water paints the pool, the two falls and the stream down the slope.
func (s TigersNest) water(g *Grid, t *tigersNestTerrain) {
	// Upper fall down the back wall of the cleft, from the notch to the pool.
	g.Use(Water).Fill(Func{Box{V(14.3, tigersNestPool-0.5, tigersNestBack-1.8), V(17.7, 67, tigersNestBack+0.6)}, func(p Vec3) bool {
		w := 1.6 + 0.25*math.Sin(p.Y*0.7)
		return math.Abs(p.X-16) <= w && p.Y <= t.top.at(16, tigersNestBack-3)+0.3
	}})
	g.Paint(White, Func{Box{V(14, tigersNestPool-1, tigersNestBack-2), V(18, 68, tigersNestBack+0.8)}, func(p Vec3) bool {
		return s.noise(p.X/0.9, p.Y/5, 45) > 0.6
	}})
	// Pool.
	g.Paint(Water, Block(V(tigersNestCleftW+0.3, tigersNestPool-1.2, tigersNestBack), V(tigersNestCleftE-0.2, tigersNestPool+0.6, tigersNestLip)))
	g.Paint(White, Ellipsoid{V(16, tigersNestPool, tigersNestBack+1.2), V(2.6, 1.2, 1.8)})
	// Lower fall over the lip.
	ground := t.ground.at(16, tigersNestLip+2)
	g.Use(Water).Fill(Func{Box{V(14.2, ground-1, tigersNestLip-0.6), V(17.8, tigersNestPool+0.4, tigersNestLip+0.8)}, func(p Vec3) bool {
		return math.Abs(p.X-16) <= 1.4+0.2*math.Sin(p.Y*0.8)
	}})
	g.Paint(White, Func{Box{V(14, ground-1, tigersNestLip-1), V(18, tigersNestPool+1, tigersNestLip+1)}, func(p Vec3) bool {
		return s.noise(p.X/0.9, p.Y/4, 46) > 0.62
	}})
	// Stream winding down the slope to the rim.
	pts := s.stream()
	g.Paint(Water, Func{Box{V(tigersNestX0, 0, tigersNestLip), V(tigersNestX1, 20, tigersNestZ1)}, func(p Vec3) bool {
		return s.polyDist(pts, p.Plan()) < 1.3 && p.Y > t.ground.at(p.X, p.Z)-1.4
	}})
	g.Paint(White, Func{Box{V(tigersNestX0, 0, tigersNestLip), V(tigersNestX1, 20, tigersNestZ1)}, func(p Vec3) bool {
		return s.polyDist(pts, p.Plan()) < 1.3 && p.Y > t.ground.at(p.X, p.Z)-1.4 && s.noise(p.X/1.2, p.Z/1.2, 47) > 0.75
	}})
}

func (TigersNest) stream() []Vec2 {
	return []Vec2{P(16, tigersNestLip), P(16.5, -1), P(14.5, 4), P(16, 9), P(19.5, 13), P(18.5, 18), P(17, 24)}
}

func (TigersNest) segDist(p, a, b Vec2) float64 {
	ab := b.Sub(a)
	t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
	return p.Sub(a.Add(ab.Scale(t))).Length()
}

func (s TigersNest) polyDist(pts []Vec2, p Vec2) float64 {
	d := math.Inf(1)
	for i := 0; i+1 < len(pts); i++ {
		d = math.Min(d, s.segDist(p, pts[i], pts[i+1]))
	}
	return d
}

// tigersNestHall is one temple building backed against the alcove.
type tigersNestHall struct {
	x0, x1, z0, z1 float64
	base           float64
	storeys        int
	pavilion       int  // 0 none, 1 small gilded pavilion, 2 large two-tier pavilion.
	bays           bool // Projecting wooden window bays on the top storey.
	west, east     bool // Windows on the side walls.
	free           bool // Free-standing: battered on all four sides.
}

const (
	tigersNestStorey = 3.0
	tigersNestKemar  = 2.0
	tigersNestBatter = 0.05
)

func (s TigersNest) monastery(g *Grid, t *tigersNestTerrain) {
	halls := []tigersNestHall{
		{x0: -29.5, x1: -24.6, z0: -12.6, z1: -2.6, base: 33, storeys: 1, east: true, west: true},
		{x0: -23.5, x1: -14, z0: -12.6, z1: -2, base: 39, storeys: 2, pavilion: 1, bays: true, west: true},
		{x0: -23, x1: -17.5, z0: -12.6, z1: -7.5, base: 49, storeys: 1, pavilion: 1, west: true, east: true},
		{x0: -13, x1: -3, z0: -12.6, z1: -1.4, base: 36, storeys: 3, pavilion: 2, bays: true, west: true, east: true},
		{x0: -2, x1: 7.5, z0: -12.6, z1: -2, base: 33, storeys: 2, bays: true, east: true},
		{x0: -1, x1: 6, z0: -12.6, z1: -8, base: 43, storeys: 1, pavilion: 1, west: true, east: true},
		{x0: 8.5, x1: 12.6, z0: -11.5, z1: -3.5, base: tigersNestDeck, storeys: 1, east: true},
	}
	for _, h := range halls {
		s.hall(g, t, h)
	}
	// Walkways with white parapets in front of the halls, and stairs between levels.
	for _, h := range halls[:len(halls)-1] {
		if h.z1 > -3 {
			s.walkway(g, t, h.x0-0.6, h.x1+0.6, h.z1, h.base)
		}
	}
	s.walkway(g, t, 8, tigersNestCleftW-0.5, -3.5, tigersNestDeck)
	for _, st := range [][2]Vec3{
		{V(-0.2, 33, -1.2), V(-3.6, 36.2, -1.2)},
		{V(-11.6, 36, -0.4), V(-15, 39.2, -0.4)},
		{V(-26.8, 33, -1.6), V(-22.6, 39.2, -1.2)},
		{V(7.4, 33, -1.2), V(9.4, 34.2, -2.5)},
	} {
		s.stairs(g, st[0], st[1], 1.6, false)
	}
	// A small hermitage clinging to the cliff above the west end, on a rock corbel.
	g.With(DarkStone, func() { g.Fill(Frustum{P(-27.5, -2.8), 0.8, 3.0, 46.5, 50.6}) })
	s.hall(g, t, tigersNestHall{x0: -30, x1: -25.5, z0: -6.5, z1: -0.8, base: 50.5, storeys: 1, west: true, east: true})
}

// walkway is a stone terrace with a white parapet in front of a hall, propped on masonry.
func (s TigersNest) walkway(g *Grid, t *tigersNestTerrain, x0, x1, z, base float64) {
	g.With(White, func() {
		g.Fill(Func{Box{V(x0, base-4, z), V(x1, base-0.6, z+1.8)}, func(p Vec3) bool { return !t.solid(p) }})
		g.Fill(Block(V(x0, base, z+1.3), V(x1, base+1.0, z+1.8)))
	})
	g.With(Stone, func() { g.Fill(Block(V(x0, base-0.6, z), V(x1, base, z+1.8))) })
}

func (s TigersNest) hall(g *Grid, t *tigersNestTerrain, h tigersNestHall) {
	top := h.base + float64(h.storeys)*tigersNestStorey + tigersNestKemar
	inset := func(y float64) float64 { return math.Max(0, y-h.base) * tigersNestBatter }
	shape := func(p Vec3, extra float64) bool {
		in := inset(p.Y) + extra
		back := h.z0 + extra
		if h.free {
			back += inset(p.Y)
		}
		return p.X >= h.x0+in && p.X <= h.x1-in && p.Z >= back && p.Z <= h.z1-in
	}
	box := Box{V(h.x0, h.base-5, h.z0), V(h.x1, top, h.z1)}
	// Walls, and a foundation reaching down to the rock.
	g.Use(White).Fill(Func{box, func(p Vec3) bool {
		return shape(p, 0) && (p.Y >= h.base-0.3 || !t.solid(p))
	}})
	g.Carve(Func{box, func(p Vec3) bool { return p.Y >= h.base+0.6 && p.Y <= top-0.9 && shape(p, 1.1) }})
	g.Paint(Charcoal, Func{box, func(p Vec3) bool { return p.Y >= h.base && p.Y <= top-0.4 && shape(p, 0.6) }})
	// Kemar band with a dark cornice below and gilded discs.
	skin := func(y0, y1 float64) Func {
		return Func{Box{V(h.x0, y0, h.z0), V(h.x1, y1, h.z1)}, func(p Vec3) bool { return shape(p, 0) && !shape(p, 0.6) }}
	}
	g.Paint(Wood, skin(top-tigersNestKemar-0.6, top-tigersNestKemar))
	g.Paint(Red, skin(top-tigersNestKemar, top))
	zf := func(y float64) float64 { return h.z1 - inset(y) }
	discY := top - tigersNestKemar/2
	for x := h.x0 + 1.5; x <= h.x1-1.2; x += 2.4 {
		g.Paint(Gold, Block(V(x-0.35, discY-0.35, zf(discY)-0.7), V(x+0.35, discY+0.35, zf(discY)+0.3)))
	}
	// Windows on every storey: dark trapezoids in wooden frames.
	for k := 0; k < h.storeys; k++ {
		sill := h.base + float64(k)*tigersNestStorey + 0.8
		bay := h.bays && k == h.storeys-1
		width := h.x1 - h.x0
		n := int(math.Max(1, math.Floor((width-1.5)/2.7)))
		pitch := (width - 1.5) / float64(n)
		for i := 0; i < n; i++ {
			x := h.x0 + 0.75 + pitch*(float64(i)+0.5)
			s.window(g, P(x, zf(sill)), sill, false, bay)
		}
		depth := h.z1 - h.z0
		m := int(math.Max(1, math.Floor((depth-2)/2.8)))
		zp := (depth - 2) / float64(m)
		for i := 0; i < m; i++ {
			z := h.z1 - 1 - zp*(float64(i)+0.5)
			if h.west {
				s.window(g, P(h.x0+inset(sill), z), sill, true, false)
			}
			if h.east {
				s.window(g, P(h.x1-inset(sill), z), sill, true, false)
			}
		}
	}
	s.roof(g, h, top)
}

// window carves a trapezoidal opening at a wall point; side windows face ±X.
func (s TigersNest) window(g *Grid, at Vec2, sill float64, side, bay bool) {
	orient := func(sol Solid) Solid {
		if side {
			return Translate(RotateY(sol, P(0, 0), math.Pi/2), at.At(0))
		}
		return Translate(sol, at.At(0))
	}
	if bay {
		g.With(Wood, func() { g.Fill(orient(Block(V(-1.3, sill-0.7, -0.3), V(1.3, sill+2.3, 0.9)))) })
		g.With(Red, func() { g.Fill(orient(Block(V(-1.5, sill+2.3, -0.3), V(1.5, sill+2.7, 1.2)))) })
	} else {
		g.Paint(Wood, orient(Block(V(-1.0, sill-0.45, -0.7), V(1.0, sill+2.1, 0.5))))
	}
	g.Carve(orient(Block(V(-0.5, sill, -1.8), V(0.5, sill+0.9, 1.6))), orient(Block(V(-0.75, sill+0.9, -1.8), V(0.75, sill+1.7, 1.6))))
}

// roof adds the open loft, the wide-eaved shingle roof and the gilded pavilion.
func (s TigersNest) roof(g *Grid, h tigersNestHall, top float64) {
	const eave, pitch, rise = 1.6, 0.36, 1.1
	rb := top + rise
	lo, hi := P(h.x0-eave, h.z0), P(h.x1+eave, h.z1+eave)
	if h.free {
		lo.Y -= eave
	}
	dist := func(p Vec3) float64 {
		d := math.Min(math.Min(p.X-lo.X, hi.X-p.X), hi.Y-p.Z)
		if h.free {
			d = math.Min(d, p.Z-lo.Y)
		}
		return d
	}
	maxRise := pitch * math.Min((hi.X-lo.X)/2, hi.Y-lo.Y)
	if h.free {
		maxRise = pitch * math.Min(hi.X-lo.X, hi.Y-lo.Y) / 2
	}
	// Dark loft behind posts.
	wallBox := Box{V(h.x0, top, h.z0), V(h.x1, rb, h.z1)}
	g.With(Charcoal, func() {
		g.Fill(Func{wallBox, func(p Vec3) bool {
			in := (top-h.base)*tigersNestBatter + 0.9
			return p.X >= h.x0+in && p.X <= h.x1-in && p.Z <= h.z1-in
		}})
	})
	g.With(Wood, func() {
		in := (top - h.base) * tigersNestBatter
		for x := h.x0 + in + 0.4; x <= h.x1-in-0.3; x += 2.2 {
			g.Fill(Block(V(x-0.35, top, h.z1-in-0.8), V(x+0.35, rb, h.z1-in)))
		}
	})
	roofBox := Box{lo.At(rb - 1), hi.At(rb + maxRise + 0.5)}
	g.With(Slate, func() {
		g.Fill(Func{roofBox, func(p Vec3) bool {
			d := dist(p)
			if d < 0 {
				return false
			}
			y := rb + math.Min(pitch*d, maxRise)
			return p.Y <= y && p.Y >= y-0.95
		}})
	})
	// Dark eave edge.
	g.Paint(Charcoal, Func{roofBox, func(p Vec3) bool { return dist(p) < 0.5 }})
	cx := (h.x0 + h.x1) / 2
	cz := h.z0 + (h.z1-h.z0)*0.55
	if h.free {
		cz = (h.z0 + h.z1) / 2
	}
	ridge := rb + maxRise
	switch h.pavilion {
	case 0:
		g.With(Gold, func() {
			g.Fill(Cylinder{Center: P(cx, cz), Radius: 0.5, Y0: ridge - 0.5, Y1: ridge + 1.2})
			g.Fill(Sphere(V(cx, ridge+1.3, cz), 0.55))
		})
	default:
		w := 3.4
		if h.pavilion == 2 {
			w = 4.6
		}
		s.pavilion(g, P(cx, cz), w, rb, ridge+1.4, h.pavilion == 2)
	}
}

// pavilion is a small red-walled tower through the roof with a gilded hip roof and finial.
func (TigersNest) pavilion(g *Grid, c Vec2, w, y0, y1 float64, tiers bool) {
	g.With(Red, func() { g.Fill(BlockAt(c, w, w, y0, y1)) })
	g.With(Wood, func() { g.Fill(BlockAt(c, w+0.4, w+0.4, y1-0.5, y1)) })
	gold := func(half, base, pitch float64) float64 {
		peak := base + pitch*half
		g.With(Gold, func() {
			g.Fill(Func{Box{V(c.X-half, base, c.Y-half), V(c.X+half, peak, c.Y+half)}, func(p Vec3) bool {
				d := half - math.Max(math.Abs(p.X-c.X), math.Abs(p.Z-c.Y))
				return d >= 0 && p.Y <= base+0.5+pitch*d
			}})
		})
		return peak
	}
	peak := gold(w/2+1.1, y1+0.3, 0.55)
	if tiers {
		g.With(Red, func() { g.Fill(BlockAt(c, w*0.55, w*0.55, peak-1, peak+1.2)) })
		peak = gold(w*0.275+0.8, peak+1.2, 0.6)
	}
	g.With(Gold, func() {
		g.Fill(Cylinder{Center: c, Radius: 0.5, Y0: peak - 0.5, Y1: peak + 1.6})
		g.Fill(Sphere(c.At(peak+0.9), 0.75))
		g.Fill(Cylinder{Center: c, Radius: 0.4, Y0: peak + 1.6, Y1: peak + 2.4})
	})
}

// crossing is the bridge, prayer flags, stairs, chorten and teahouse on the ridge.
func (s TigersNest) crossing(g *Grid, t *tigersNestTerrain) {
	// Wooden bridge over the pool.
	z := tigersNestBridgeZ
	g.With(Wood, func() {
		g.Fill(Block(V(tigersNestCleftW-1, tigersNestDeck-0.6, z-1.1), V(tigersNestCleftE+1.5, tigersNestDeck, z+1.1)))
		for _, side := range []float64{-1, 1} {
			g.Fill(Block(V(tigersNestCleftW-1, tigersNestDeck+1, z+side*1.1-0.2), V(tigersNestCleftE+1.5, tigersNestDeck+1.4, z+side*1.1+0.2)))
			for x := tigersNestCleftW; x <= tigersNestCleftE+1.5; x += 2.2 {
				g.Fill(Block(V(x-0.3, tigersNestDeck, z+side*1.1-0.3), V(x+0.3, tigersNestDeck+1.4, z+side*1.1+0.3)))
			}
		}
	})
	// Prayer flags fanned over the bridge.
	for i, a := range []Vec3{V(12.6, 44, -11.5), V(12.6, 41, -6), V(12.8, 47, -9)} {
		b := []Vec3{V(20, 39, -6), V(20.5, 41, -10.5), V(21, 43, -7.5)}[i]
		s.flags(g, a, b, 2.2, i)
	}
	// Stairs climbing the ridge to the viewpoint.
	route := []Vec3{
		V(tigersNestCleftE+1.5, tigersNestDeck, z),
		V(23, tigersNestDeck+0.5, -6.5),
		V(25.5, 39.5, -4),
		V(28.5, 44, -6.8),
	}
	for i := 0; i+1 < len(route); i++ {
		s.stairs(g, route[i], route[i+1], 2, true)
	}
	for i := 1; i+1 < len(route); i++ {
		a, b := route[i], route[i+1]
		s.flags(g, a.Add(V(0, 4, 1.4)), b.Add(V(0, 4, 1.4)), 1.2, i)
		s.pole(g, a.Plan().Add(P(0, 1.4)), a.Y-0.5, a.Y+4.3, false)
	}
	// Chorten by the bridgehead.
	s.chorten(g, P(24, -10.5), t.surface(24, -10.5))
	// Tall white prayer flags around the viewpoint.
	for i, p := range []Vec2{P(31.5, -3.5), P(33, -6), P(22.5, -2.5), P(26, -13.5), P(34, -10)} {
		s.pole(g, p, t.surface(p.X, p.Y)-0.5, t.surface(p.X, p.Y)+7+float64(i%2), true)
	}
	// Teahouse at the viewpoint.
	base := math.Floor(t.surface(29.5, -11)*2) / 2
	g.With(DarkStone, func() {
		g.Fill(Func{Box{V(26.3, base-9, -14.2), V(32.7, base-0.6, -6.8)}, func(p Vec3) bool { return !t.solid(p) }})
	})
	g.With(Sand, func() { g.Fill(Block(V(26.3, base-0.6, -14.2), V(32.7, base, -6.8))) })
	s.hall(g, t, tigersNestHall{x0: 27, x1: 32, z0: -13.5, z1: -9, base: base, storeys: 1, west: true, east: true, free: true})
}

// stairs fills stone steps from a to b and clears the terrain above them.
func (s TigersNest) stairs(g *Grid, a, b Vec3, width float64, clear bool) {
	run := b.Plan().Sub(a.Plan())
	n := int(math.Max(1, math.Ceil(math.Abs(b.Y-a.Y)/0.5)))
	n = int(math.Max(float64(n), math.Ceil(run.Length()/0.6)))
	angle := run.Angle()
	step := run.Length() / float64(n)
	for i := 0; i <= n; i++ {
		f := float64(i) / float64(n)
		c := a.Plan().Add(run.Scale(f))
		y := a.Y + (b.Y-a.Y)*f
		if clear {
			g.Carve(Translate(RotateY(Block(V(-step/2-0.3, y+0.1, -width/2), V(step/2+0.3, y+3, width/2)), P(0, 0), angle), c.At(0)))
		}
		g.With(Stone, func() {
			g.Fill(Translate(RotateY(Block(V(-step/2-0.3, y-1.6, -width/2), V(step/2+0.3, y, width/2)), P(0, 0), angle), c.At(0)))
		})
	}
}

// flags strings a sagging line of five-colored prayer flags from a to b.
func (TigersNest) flags(g *Grid, a, b Vec3, sag float64, seed int) {
	colors := []Color{Navy, White, Red, Grass, Gold}
	n := int(math.Max(2, math.Round(b.Sub(a).Length()/0.85)))
	prev := a
	for i := 1; i <= n; i++ {
		f := float64(i) / float64(n)
		p := a.Add(b.Sub(a).Scale(f))
		p.Y -= sag * 4 * f * (1 - f)
		g.With(colors[(i+seed)%len(colors)], func() {
			g.Fill(Capsule{prev, p, 0.45}, Capsule{p, p.Sub(V(0, 0.9, 0)), 0.45})
		})
		prev = p
	}
}

// pole is a wooden flagpole; tall ones carry a long white banner.
func (TigersNest) pole(g *Grid, at Vec2, y0, y1 float64, banner bool) {
	g.With(Wood, func() { g.Fill(Cylinder{Center: at, Radius: 0.45, Y0: y0, Y1: y1}) })
	if banner {
		g.With(White, func() { g.Fill(Block(V(at.X+0.3, y1-5, at.Y-0.25), V(at.X+1.6, y1-0.4, at.Y+0.25))) })
	}
	g.With(Gold, func() { g.Fill(Sphere(at.At(y1+0.3), 0.5)) })
}

// chorten is a small white stupa with a gilded spire.
func (TigersNest) chorten(g *Grid, at Vec2, y float64) {
	g.With(White, func() {
		g.Fill(BlockAt(at, 3, 3, y-1, y+1))
		g.Fill(BlockAt(at, 2.2, 2.2, y+1, y+1.8))
		g.Fill(Ellipsoid{at.At(y + 2.4), V(1.3, 1.1, 1.3)})
	})
	g.With(Gold, func() { g.Fill(Cone(at, 0.7, y+3.2, 2.2)) })
	g.With(Red, func() { g.Fill(BlockAt(at, 1.2, 1.2, y+3.1, y+3.6)) })
}

// forest scatters pines over the slope, ridge and cliff top.
func (s TigersNest) forest(g *Grid, t *tigersNestTerrain) {
	stream := s.stream()
	const cell = 3.1
	for i := 0; float64(i)*cell+tigersNestX0 < tigersNestX1; i++ {
		for j := 0; float64(j)*cell+tigersNestZ0 < tigersNestZ1; j++ {
			x := tigersNestX0 + (float64(i)+0.2+0.6*Hash01(i, j, 51))*cell
			z := tigersNestZ0 + (float64(j)+0.2+0.6*Hash01(i, j, 52))*cell
			if x < tigersNestX0+3 || x > tigersNestX1-3 || z < tigersNestZ0+3 || z > tigersNestZ1-3 {
				continue
			}
			y := t.surface(x, z)
			onTop := t.top.at(x, z) > t.ground.at(x, z) && y > t.ground.at(x, z)
			grad := t.gradient(t.ground, x, z)
			if onTop {
				grad = t.gradient(t.top, x, z)
			}
			if grad > 2 || y < 1.5 || Hash01(i, j, 53) > 0.55 {
				continue
			}
			if s.polyDist(stream, P(x, z)) < 2.5 || (x > 11 && x < 22 && z < tigersNestLip+2) {
				continue
			}
			if x > 19 && x < 35 && z > -15 && z < -1 && y > 33 {
				continue // Viewpoint, stairs and bridgehead.
			}
			h := 4.5 + 3*Hash01(i, j, 54)
			s.pine(g, P(x, z), y, h)
		}
	}
}

func (TigersNest) pine(g *Grid, at Vec2, base, h float64) {
	g.With(Wood, func() { g.Fill(Cylinder{Center: at, Radius: 0.45, Y0: base - 0.8, Y1: base + h*0.4}) })
	g.With(Foliage, func() {
		for k := 0; k < 3; k++ {
			r := h * 0.3 * (1 - 0.22*float64(k))
			y0 := base + h*(0.2+0.24*float64(k))
			g.Fill(Cone(at, r, y0, h*0.45))
		}
	})
}
