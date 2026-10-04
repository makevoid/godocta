package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// CinqueTerre is a Manarola/Vernazza-style Ligurian fishing village. Tall, narrow tower
// houses in red, terracotta, ochre, sand and white, with dark green shutters and slate or
// terracotta roofs, are packed up both flanks of a steep ravine and over a rocky
// promontory (east). The main street runs down the ravine floor to a little harbour cove
// with a slipway where wooden boats are pulled up. A church with a striped bell tower
// closes the top of the village; a round Doria tower guards the western headland. Dry
// stone vineyard terraces climb the hills above; dark cliffs drop into the sea with
// surf at their feet, and a cliff path runs east along the promontory. The sea is at the
// front (+Z); the land slopes down to the sea at the sides and to a low rim at the back.
type CinqueTerre struct{}

func init() { Register(CinqueTerre{}) }

func (CinqueTerre) Name() string  { return "cinque_terre" }
func (CinqueTerre) Title() string { return "Cinque Terre" }

// Ground kinds.
const (
	cinqueTerreSea = iota
	cinqueTerreRock
	cinqueTerreTown
	cinqueTerreStreet
	cinqueTerreVine
	cinqueTerreScrub
	cinqueTerrePath
	cinqueTerreSlip
)

// Site extent (plan) and heightmap resolution.
const (
	cinqueTerreX0  = -33.0
	cinqueTerreX1  = 33.0
	cinqueTerreZ0  = -39.0
	cinqueTerreZ1  = 22.0
	cinqueTerreRes = 0.25
	cinqueTerreSkn = 1.1  // Ground skin thickness.
	cinqueTerreSty = 2.5  // Storey height.
	cinqueTerreTer = 1.5  // Vineyard terrace riser.
	cinqueTerreWal = 1.05 // House wall thickness.
	cinqueTerreK   = 1.15 // World units per design unit.
)

func cinqueTerreSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func cinqueTerreG(u, w float64) float64 { return math.Exp(-(u / w) * (u / w)) }

func cinqueTerreNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return ((a+(b-a)*sx)*(1-sz)+(c+(d-c)*sx)*sz)*2 - 1
}

// cinqueTerreSite caches the terrain heightmap and kinds.
type cinqueTerreSite struct {
	nx, nz int
	h      []float64
	k      []uint8
}

// coast is the shoreline's z at plan x: land lies at smaller z.
func (CinqueTerre) coast(x float64) float64 {
	ax := math.Max(0, math.Abs(x)-24)
	return 6 + 11*cinqueTerreG(x-16, 7.5) + 8*cinqueTerreG(x+23, 5) - 2*cinqueTerreG(x+1, 4) - 0.12*ax*ax
}

// ravine is the x of the ravine floor (the main street) at z.
func (CinqueTerre) ravine(z float64) float64 { return -1 + 1.4*math.Sin(z*0.1) }

// rim fades the land down to a low rim at the sides and back.
func (CinqueTerre) rim(x, z float64) float64 {
	return cinqueTerreSmooth((cinqueTerreX1-math.Abs(x))/11) * cinqueTerreSmooth((z-cinqueTerreZ0)/9)
}

// raw is the smooth elevation (negative in the sea).
func (c CinqueTerre) raw(x, z float64) float64 {
	d := c.coast(x) - z
	if d <= 0 {
		return -1
	}
	cliff := 5.5 + 3*cinqueTerreG(x-16, 6)
	h := cliff*cinqueTerreSmooth(d/2.5) + 22*(1-math.Exp(-math.Max(d-2.5, 0)/20))
	// The vineyard hill behind the village.
	h += 9 * cinqueTerreSmooth((-z-16)/16) * cinqueTerreSmooth((30-math.Abs(x+2))/14)
	// The western headland knob.
	h += 5 * math.Exp(-((x+23)*(x+23)+(z-9)*(z-9))/30)
	// The ravine.
	depth := (6 + 2.5*cinqueTerreSmooth((z+1)/5)) * cinqueTerreSmooth((z+28)/8)
	h -= depth * cinqueTerreG(x-c.ravine(z), 5)
	h *= c.rim(x, z)
	h += 0.8 * cinqueTerreNoise(x*0.2, z*0.2, 3) * cinqueTerreSmooth(d/4)
	return math.Max(h, 0.6)
}

// town reports whether a plan point lies in the village.
func (c CinqueTerre) town(x, z float64) bool {
	e1 := ((x+3)/24)*((x+3)/24) + ((z+4)/16.5)*((z+4)/16.5)
	e2 := ((x-16)/11)*((x-16)/11) + ((z-3)/14)*((z-3)/14)
	return e1 < 1 || e2 < 1
}

// street reports the main street along the ravine floor.
func (c CinqueTerre) street(x, z float64) bool {
	return z > -17 && z < c.coast(x)-0.5 && math.Abs(x-c.ravine(z)) < 2.6
}

// slip reports the harbour slipway and piazza at the head of the cove.
func (c CinqueTerre) slip(x, z float64) bool {
	return x > -4.5 && x < 2.5 && z > c.coast(x)-6.5 && z < c.coast(x)+2.5
}

// Church site.
var (
	cinqueTerreChurchMin = P(-6, -27)
	cinqueTerreChurchMax = P(2.5, -17)
	cinqueTerrePiazzaMin = P(-7, -17)
	cinqueTerrePiazzaMax = P(10, -13)
	cinqueTerreBellAt    = P(7.5, -15.5)
	cinqueTerreTowerAt   = P(-23, 9)
)

// The cliff path along the east of the promontory, as plan points at cinqueTerrePathY.
var cinqueTerrePathPts = []Vec2{P(20, 13), P(23, 9), P(26, 4), P(29, 0.5), P(31.5, -3)}

const cinqueTerrePathY = 5.5

func (CinqueTerre) pathDist(x, z float64) float64 {
	best := math.Inf(1)
	for i := 0; i+1 < len(cinqueTerrePathPts); i++ {
		a, b := cinqueTerrePathPts[i], cinqueTerrePathPts[i+1]
		ab := b.Sub(a)
		t := math.Max(0, math.Min(1, P(x, z).Sub(a).Dot(ab)/ab.Dot(ab)))
		best = math.Min(best, P(x, z).Sub(a.Add(ab.Scale(t))).Length())
	}
	return best
}

func cinqueTerreInBox(p, a, b Vec2) bool { return p.X >= a.X && p.X <= b.X && p.Y >= a.Y && p.Y <= b.Y }

// churchFloor is the level of the church and its piazza.
func (c CinqueTerre) churchFloor() float64 { return c.raw(-1.5, -17) + 0.3 }

// classify computes the final height and kind at a plan point.
func (c CinqueTerre) classify(x, z float64) (float64, uint8) {
	d := c.coast(x) - z
	if c.slip(x, z) {
		return 0.1 + 2.4*cinqueTerreSmooth((c.coast(x)+2.5-z)/8.5), cinqueTerreSlip
	}
	if c.pathDist(x, z) < 1.3 && d > -1.5 {
		return cinqueTerrePathY, cinqueTerrePath
	}
	h := c.raw(x, z)
	if h < 0 || d <= 0 {
		return -1, cinqueTerreSea
	}
	p := P(x, z)
	if cinqueTerreInBox(p, cinqueTerrePiazzaMin, cinqueTerrePiazzaMax) ||
		cinqueTerreInBox(p, cinqueTerreChurchMin.Sub(P(1, 1)), cinqueTerreChurchMax.Add(P(1, 0))) {
		return c.churchFloor(), cinqueTerreTown
	}
	if c.street(x, z) {
		return c.raw(c.ravine(z), z), cinqueTerreStreet
	}
	if c.town(x, z) && d > 0.8 {
		return h, cinqueTerreTown
	}
	slope := math.Hypot(c.raw(x+0.5, z)-c.raw(x-0.5, z), c.raw(x, z+0.5)-c.raw(x, z-0.5))
	if d < 3.2 || (slope > 1.5 && c.rim(x, z) > 0.97) || h < 2 {
		return h, cinqueTerreRock
	}
	if math.Hypot(x-cinqueTerreTowerAt.X, z-cinqueTerreTowerAt.Y) < 8 || slope > 1.5 {
		return h, cinqueTerreScrub
	}
	if cinqueTerreNoise(x*0.12, z*0.12, 21) > 0.5 || h < 5 {
		return h, cinqueTerreScrub
	}
	// Vineyard terraces.
	return math.Floor(h/cinqueTerreTer)*cinqueTerreTer + 0.5, cinqueTerreVine
}

func (c CinqueTerre) site() *cinqueTerreSite {
	s := &cinqueTerreSite{
		nx: int((cinqueTerreX1-cinqueTerreX0)/cinqueTerreRes) + 1,
		nz: int((cinqueTerreZ1-cinqueTerreZ0)/cinqueTerreRes) + 1,
	}
	s.h = make([]float64, s.nx*s.nz)
	s.k = make([]uint8, s.nx*s.nz)
	for j := 0; j < s.nz; j++ {
		for i := 0; i < s.nx; i++ {
			h, k := c.classify(cinqueTerreX0+float64(i)*cinqueTerreRes, cinqueTerreZ0+float64(j)*cinqueTerreRes)
			s.h[j*s.nx+i], s.k[j*s.nx+i] = h, k
		}
	}
	return s
}

func (s *cinqueTerreSite) at(x, z float64) (float64, uint8) {
	i := int(math.Round((x - cinqueTerreX0) / cinqueTerreRes))
	j := int(math.Round((z - cinqueTerreZ0) / cinqueTerreRes))
	if i < 0 || j < 0 || i >= s.nx || j >= s.nz {
		return -1, cinqueTerreSea
	}
	return s.h[j*s.nx+i], s.k[j*s.nx+i]
}

func (s *cinqueTerreSite) height(x, z float64) float64 { h, _ := s.at(x, z); return h }

// Build models the scene in design units on a finer grid, then merges its cells: the
// lattice coordinates carry over, so the whole scene scales up by cinqueTerreK.
func (c CinqueTerre) Build(world *Grid) Order {
	g := NewGridWithSpan(world.Span() / cinqueTerreK)
	s := c.site()
	c.sea(g, s)
	c.terrain(g, s)
	c.village(g, s)
	c.church(g, s)
	c.doria(g, s)
	c.harbour(g, s)
	c.railing(g, s)
	c.rocks(g, s)
	c.vines(g, s)
	c.trees(g, s)
	world.Merge(g)
	return Outward(P(-1, 12).Scale(cinqueTerreK))
}

// sea is a water layer with surf along the shore.
func (CinqueTerre) sea(g *Grid, s *cinqueTerreSite) {
	box := Box{V(cinqueTerreX0, -0.2, cinqueTerreZ0), V(cinqueTerreX1, 0.2, cinqueTerreZ1)}
	g.Use(Water).Fill(Func{box, func(p Vec3) bool { return s.height(p.X, p.Z) < 0 }})
	g.Paint(White, Func{box, func(p Vec3) bool {
		if s.height(p.X, p.Z) >= 0 {
			return false
		}
		near := 0
		for k := 0; k < 8; k++ {
			a := float64(k) * math.Pi / 4
			for _, r := range []float64{0.8, 1.6} {
				if s.height(p.X+r*math.Cos(a), p.Z+r*math.Sin(a)) >= 0 {
					near++
				}
			}
		}
		return near > 0 && Hash01(int(math.Floor(p.X*1.3)), int(math.Floor(p.Z*1.3)), 5) < 0.25+0.1*float64(near)
	}})
}

// terrain builds the hollow ground skin and colors it.
func (c CinqueTerre) terrain(g *Grid, s *cinqueTerreSite) {
	in := func(p Vec3) bool {
		h := s.height(p.X, p.Z)
		return h >= 0 && p.Y >= -0.2 && p.Y <= h
	}
	k := cinqueTerreSkn
	box := Box{V(cinqueTerreX0, -0.2, cinqueTerreZ0), V(cinqueTerreX1, 36, cinqueTerreZ1)}
	g.Use(DarkStone).Fill(Func{box, func(p Vec3) bool {
		if !in(p) {
			return false
		}
		return !(in(p.Add(V(k, 0, 0))) && in(p.Add(V(-k, 0, 0))) && in(p.Add(V(0, 0, k))) &&
			in(p.Add(V(0, 0, -k))) && in(p.Add(V(0, k, 0))))
	}})
	color := func(p Vec3) Color {
		h, kind := s.at(p.X, p.Z)
		if h < 0 {
			return PaletteSize
		}
		top := p.Y > h-0.6
		rock := func() Color {
			wob := 0.7*math.Sin(p.X*0.13+p.Z*0.05) + 0.5*cinqueTerreNoise(p.X*0.3, p.Z*0.3, 9)
			band := int(math.Floor((p.Y + wob) / 0.9))
			r := Hash01(band, int(math.Floor((p.X+p.Z*0.4)/9)), 4)
			switch {
			case r < 0.6:
				return DarkStone
			case r < 0.78:
				return Charcoal
			case r < 0.96:
				return Stone
			default:
				return Wood
			}
		}
		switch kind {
		case cinqueTerreStreet, cinqueTerreSlip, cinqueTerrePath:
			if top {
				return Stone
			}
		case cinqueTerreTown:
			slope := math.Hypot(s.height(p.X+0.6, p.Z)-s.height(p.X-0.6, p.Z), s.height(p.X, p.Z+0.6)-s.height(p.X, p.Z-0.6))
			if top && slope < 1.3 {
				switch r := Hash01(int(math.Floor(p.X/1.5)), int(math.Floor(p.Z/1.5)), 8); {
				case r < 0.1:
					return Foliage // Kitchen gardens.
				case r < 0.25:
					return Sand
				}
				return Stone
			}
		case cinqueTerreVine:
			if top {
				return Grass
			}
			return Stone // Dry-stone retaining walls.
		case cinqueTerreScrub:
			if top {
				if cinqueTerreNoise(p.X*0.5, p.Z*0.5, 12) > 0.1 {
					return Foliage
				}
				return Grass
			}
		}
		return rock()
	}
	for _, col := range []Color{Stone, Sand, Grass, Foliage, Charcoal, Wood} {
		g.Paint(col, Func{box, func(p Vec3) bool { return color(p) == col }})
	}
}

// vines are rows of vines on the terraces, running along the contours.
func (c CinqueTerre) vines(g *Grid, s *cinqueTerreSite) {
	box := Box{V(cinqueTerreX0, 0, cinqueTerreZ0), V(cinqueTerreX1, 36, cinqueTerreZ1)}
	g.Use(Foliage).Fill(Func{box, func(p Vec3) bool {
		h, kind := s.at(p.X, p.Z)
		if kind != cinqueTerreVine || p.Y <= h || p.Y > h+1.1 {
			return false
		}
		r := c.raw(p.X, p.Z)
		f := r/cinqueTerreTer - math.Floor(r/cinqueTerreTer)
		// Keep a strip at the terrace's edge free, and gaps between plants.
		return f > 0.25 && f < 0.6 && Hash01(int(math.Floor(p.X*0.8)), int(math.Floor(p.Z*0.8)), 31) < 0.85
	}})
}

// house builds a narrow tower house on the plan rectangle [x0,x1]×[z0,z1].
func (c CinqueTerre) house(g *Grid, s *cinqueTerreSite, x0, x1, z0, z1 float64, storeys int, seed int) {
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, x := range []float64{x0, (x0 + x1) / 2, x1} {
		for _, z := range []float64{z0, (z0 + z1) / 2, z1} {
			h := s.height(x, z)
			lo, hi = math.Min(lo, h), math.Max(hi, h)
		}
	}
	y0 := lo - 0.6
	y1 := hi + float64(storeys)*cinqueTerreSty + 0.4
	r := Hash01(seed, 1)
	wall := Red
	switch {
	case r < 0.2:
		wall = Red
	case r < 0.36:
		wall = Terracotta
	case r < 0.5:
		wall = Gold
	case r < 0.62:
		wall = Sand
	case r < 0.75:
		wall = White
	case r < 0.85:
		wall = Brick
	case r < 0.92:
		wall = Grass
	default:
		wall = Glass
	}
	w := cinqueTerreWal
	g.With(wall, func() { g.Fill(Block(V(x0, y0, z0), V(x1, y1, z1))) })
	g.Paint(Charcoal, Block(V(x0+0.5, y0+0.5, z0+0.5), V(x1-0.5, y1-0.3, z1-0.5)))
	g.Carve(Block(V(x0+w, y0+0.6, z0+w), V(x1-w, y1-0.8, z1-w)))
	// A dark cross wall keeps facing side windows from showing the sky through the house.
	mx := (x0 + x1) / 2
	g.With(Charcoal, func() { g.Fill(Block(V(mx-0.5, y0+0.6, z0+w), V(mx+0.5, y1-0.8, z1-w))) })
	// Roof: slate or terracotta gable with a slight eave.
	roof := Slate
	if Hash01(seed, 2) < 0.35 {
		roof = Terracotta
	}
	ridge := AxisX
	if Hash01(seed, 3) < 0.5 {
		ridge = AxisZ
	}
	if Hash01(seed, 5) < 0.18 {
		// A flat roof terrace with a parapet and a pergola.
		g.With(wall, func() {
			g.Fill(Subtract(Block(V(x0, y1, z0), V(x1, y1+1.0, z1)), Block(V(x0+0.7, y1, z0+0.7), V(x1-0.7, y1+1.1, z1-0.7))))
		})
		g.With(Terracotta, func() { g.Fill(Block(V(x0+0.7, y1-0.2, z0+0.7), V(x1-0.7, y1+0.2, z1-0.7))) })
		px := x0 + 0.9 + 1.4*Hash01(seed, 6)
		g.With(Wood, func() {
			g.Fill(Block(V(px, y1, z1-1.7), V(px+0.7, y1+2.4, z1-1)))
			g.Fill(Block(V(px, y1+2.0, z0+1), V(px+0.7, y1+2.6, z1-1)))
		})
		g.With(Foliage, func() { g.Fill(Ellipsoid{V(x1-1.4, y1+0.9, z0+1.4), V(0.9, 0.8, 0.9)}) })
	} else {
		g.With(roof, func() {
			g.Fill(Gable{Min: P(x0-0.3, z0-0.3), Max: P(x1+0.3, z1+0.3), Base: y1, Ridge: ridge})
		})
	}
	// Windows with green shutters on the front (+Z) and sides.
	type face struct {
		axis       Axis // Facade normal axis.
		pos, out   float64
		a0, a1     float64
		groundSide func(a float64) float64
	}
	faces := []face{
		{AxisZ, z1, 1, x0, x1, func(a float64) float64 { return s.height(a, z1+0.6) }},
		{AxisX, x1, 1, z0, z1, func(a float64) float64 { return s.height(x1+0.6, a) }},
		{AxisX, x0, -1, z0, z1, func(a float64) float64 { return s.height(x0-0.6, a) }},
	}
	shutter := Foliage
	if Hash01(seed, 4) < 0.25 {
		shutter = Grass
	}
	if wall == Grass || wall == Glass {
		shutter = Wood
	}
	for fi, f := range faces {
		width := f.a1 - f.a0
		n := int(math.Max(1, math.Floor((width-0.6)/2.6)))
		step := width / float64(n)
		for st := 0; ; st++ {
			sill := y1 - 0.4 - float64(st+1)*cinqueTerreSty + 0.7
			if sill < y0+0.5 {
				break
			}
			for i := 0; i < n; i++ {
				a := f.a0 + step*(float64(i)+0.5)
				if sill < f.groundSide(a)+0.6 {
					continue
				}
				// Window opening through the wall.
				var hole, sh1, sh2 Solid
				if f.axis == AxisZ {
					hole = Block(V(a-0.5, sill, f.pos-1.2), V(a+0.5, sill+1.4, f.pos+0.2))
					sh1 = Block(V(a-1.3, sill, f.pos-0.4), V(a-0.55, sill+1.4, f.pos+0.2))
					sh2 = Block(V(a+0.55, sill, f.pos-0.4), V(a+1.3, sill+1.4, f.pos+0.2))
				} else {
					lo, hi := f.pos-0.2, f.pos+1.2
					slo, shi := f.pos-0.2, f.pos+0.4
					if f.out > 0 {
						lo, hi = f.pos-1.2, f.pos+0.2
						slo, shi = f.pos-0.4, f.pos+0.2
					}
					hole = Block(V(lo, sill, a-0.5), V(hi, sill+1.4, a+0.5))
					sh1 = Block(V(slo, sill, a-1.3), V(shi, sill+1.4, a-0.55))
					sh2 = Block(V(slo, sill, a+0.55), V(shi, sill+1.4, a+1.3))
				}
				if Hash01(seed, fi, st, i, 7) < 0.3 {
					// Closed shutters.
					g.Paint(shutter, hole)
					continue
				}
				g.Paint(shutter, sh1, sh2)
				g.Carve(hole)
			}
		}
	}
	// Front door at street level.
	fy := s.height((x0+x1)/2, z1+0.6)
	if fy > y0 && fy < y1-3 {
		a := (x0 + x1) / 2
		g.Carve(Block(V(a-0.6, fy, z1-1.2), V(a+0.6, fy+2.1, z1+0.2)))
	}
}

// village packs tower-house lots row by row over the town area: each row walks east,
// placing a lot wherever its footprint fits, shrinking its depth near the cliffs.
func (c CinqueTerre) village(g *Grid, s *cinqueTerreSite) {
	ok := func(x, z float64) bool {
		h, kind := s.at(x, z)
		return h >= 0 && kind == cinqueTerreTown && c.town(x, z) && !c.street(x, z) && !c.slip(x, z) &&
			!cinqueTerreInBox(P(x, z), cinqueTerreChurchMin.Sub(P(1.5, 1.5)), cinqueTerrePiazzaMax.Add(P(1, 1)))
	}
	fits := func(a0, a1, b0, b1 float64) bool {
		for _, x := range []float64{a0, (a0 + a1) / 2, a1} {
			for _, z := range []float64{b0, (b0 + b1) / 2, b1} {
				if !ok(x, z) {
					return false
				}
			}
		}
		return true
	}
	row := 0
	for z1 := 19.0; z1 > -27; row++ {
		depth := 4.6 + 1.2*Hash01(row, 11)
		z0 := z1 - depth
		x := -26 + 2*Hash01(row, 12)
		for col := 0; x < 28; col++ {
			w := 3.6 + 1.4*Hash01(row, col, 13)
			jz := (Hash01(row, col, 14) - 0.5) * 1.4
			a0, a1, b0, b1 := x, x+w, z0+jz, z1+jz
			placed := false
			for _, cut := range []float64{0, 1.2, 2.2} {
				if b1-cut-b0 >= 3.4 && fits(a0, a1, b0, b1-cut) {
					b1 -= cut
					placed = true
					break
				}
			}
			if !placed {
				x += 1
				continue
			}
			x += w
			if Hash01(row, col, 15) < 0.05 {
				continue // A small garden gap.
			}
			storeys := 2 + int(3*Hash01(row, col, 16))
			c.house(g, s, a0, a1, b0, b1, storeys, row*100+col)
		}
		z1 = z0
	}
}

// church builds the church at the head of the village and its striped bell tower.
func (c CinqueTerre) church(g *Grid, s *cinqueTerreSite) {
	fy := c.churchFloor()
	a, b := cinqueTerreChurchMin, cinqueTerreChurchMax
	g.Use(White).Fill(Block(V(a.X, fy-3, a.Y), V(b.X, fy+9, b.Y)))
	g.Paint(Charcoal, Block(V(a.X+0.5, fy, a.Y+0.5), V(b.X-0.5, fy+8.7, b.Y-0.5)))
	g.Carve(Block(V(a.X+1.1, fy, a.Y+1.1), V(b.X-1.1, fy+8.5, b.Y-1.1)))
	g.Use(Slate).Fill(Gable{Min: P(a.X-0.4, a.Y-0.4), Max: P(b.X+0.4, b.Y), Base: fy + 9, Ridge: AxisZ})
	// Facade gable in white with a rose window and portal.
	cx := (a.X + b.X) / 2
	g.Use(White).Fill(Gable{Min: P(a.X, b.Y-1), Max: P(b.X, b.Y), Base: fy + 9, Ridge: AxisZ})
	g.Carve(Arch{Sill: V(cx, fy, b.Y), HalfWidth: 1.1, Spring: 2.4, Depth: 2.6, Through: AxisZ})
	g.Paint(Charcoal, Func{Box{V(cx-2, fy+5, b.Y-1), V(cx+2, fy+10, b.Y+0.1)}, func(p Vec3) bool {
		return math.Hypot(p.X-cx, p.Y-fy-7.5) < 1.6
	}})
	g.Carve(Func{Box{V(cx-2, fy+5, b.Y-1.2), V(cx+2, fy+10, b.Y+0.1)}, func(p Vec3) bool {
		return math.Hypot(p.X-cx, p.Y-fy-7.5) < 0.8
	}})
	// Side windows.
	for z := a.Y + 3.0; z < b.Y-2; z += 3.5 {
		for _, x := range []float64{a.X, b.X} {
			g.Carve(Arch{Sill: V(x, fy+4, z), HalfWidth: 0.6, Spring: 1.6, Depth: 2.6, Through: AxisX})
		}
	}
	// Bell tower in the piazza: white and dark stone stripes.
	t := cinqueTerreBellAt
	tw := 2.3
	ty := s.height(t.X, t.Y)
	top := ty + 20
	g.Use(White).Fill(BlockAt(t, 2*tw, 2*tw, ty-1, top))
	g.Paint(Charcoal, BlockAt(t, 2*tw-1, 2*tw-1, ty, top-0.3))
	g.Carve(BlockAt(t, 2*tw-2.2, 2*tw-2.2, ty+0.5, top-0.5))
	for y := ty + 1.2; y < top-0.6; y += 2.4 {
		g.Paint(Slate, Block(V(t.X-tw, y, t.Y-tw), V(t.X+tw, y+1.2, t.Y+tw)))
	}
	// Belfry openings.
	g.Carve(Arch{Sill: V(t.X, top-4.5, t.Y), HalfWidth: 0.9, Spring: 1.8, Depth: 2 * tw, Through: AxisZ})
	g.Carve(Arch{Sill: V(t.X, top-4.5, t.Y), HalfWidth: 0.9, Spring: 1.8, Depth: 2 * tw, Through: AxisX})
	g.Use(White).Fill(BlockAt(t, 2*tw+0.6, 2*tw+0.6, top, top+0.6))
	g.Use(Slate).FillVolume(Cone(t, tw+0.2, top+0.6, 5.5))
	g.Use(Gold).Fill(Cylinder{Center: t, Radius: 0.45, Y0: top + 5.5, Y1: top + 7})
}

// doria builds the round castle tower and its curtain on the western headland.
func (c CinqueTerre) doria(g *Grid, s *cinqueTerreSite) {
	t := cinqueTerreTowerAt
	base := s.height(t.X, t.Y)
	top := base + 13
	g.Use(Sand).Fill(Cylinder{Center: t, Radius: 3.4, Y0: base - 4, Y1: top})
	// Battered plinth and string courses.
	g.Use(Stone).Fill(Frustum{t, 4.4, 3.4, base - 5, base + 2.5})
	g.Paint(Stone, Cylinder{Center: t, Radius: 3.6, Y0: base + 7, Y1: base + 7.6})
	g.Use(Sand)
	g.Paint(Charcoal, Cylinder{Center: t, Radius: 2.9, Y0: base, Y1: top - 0.5})
	g.Carve(Cylinder{Center: t, Radius: 2.3, Y0: base, Y1: top - 0.5})
	g.Fill(Frustum{t, 3.4, 4.1, top - 1.5, top})
	g.Fill(Cylinder{Center: t, Radius: 4.1, Inner: 3.0, Y0: top, Y1: top + 1})
	g.Fill(Func{Box{V(t.X-4.2, top+1, t.Y-4.2), V(t.X+4.2, top+2.2, t.Y+4.2)}, func(p Vec3) bool {
		d := p.Plan().Sub(t)
		turn := math.Mod(d.Angle()/(2*math.Pi)*20+20, 2)
		return d.Length() <= 4.1 && d.Length() > 3.0 && turn < 1
	}})
	g.Paint(Stone, Frustum{t, 3.5, 4.2, top - 1.5, top})
	for k := 0; k < 6; k++ {
		a := float64(k)*math.Pi/3 + 0.3
		q := Polar(t, 3.1, a)
		y := base + 4 + float64(k%2)*4
		g.Carve(Func{Box{q.At(y).Sub(V(1.2, 0, 1.2)), q.At(y + 2.2).Add(V(1.2, 0, 1.2))}, func(p Vec3) bool {
			d := p.Plan().Sub(t)
			return math.Abs(d.Rotate(-a).Y) < 0.55
		}})
	}
	// Arched door on the landward side.
	g.Carve(Arch{Sill: V(t.X, base+0.3, t.Y-3.2), HalfWidth: 0.9, Spring: 1.6, Depth: 2.4, Through: AxisZ})
	// A low curtain wall on the landward side of the knob.
	g.Use(Stone).Fill(Func{Box{V(t.X-8, base-6, t.Y-8), V(t.X+8, base+3, t.Y+8)}, func(p Vec3) bool {
		q := p.Plan().Sub(t)
		d := q.Length()
		if d < 5.6 || d > 6.6 || q.Y > 1 {
			return false
		}
		h := s.height(p.X, p.Z)
		top := h + 1.6
		if math.Mod(q.Angle()*6/math.Pi*2+40, 2) < 1 {
			top += 0.8
		}
		return h >= 0 && p.Y <= top && p.Y >= h-1.5
	}})
}

// railing lines the seaward edge of the cliff path.
func (c CinqueTerre) railing(g *Grid, s *cinqueTerreSite) {
	box := Box{V(18, cinqueTerrePathY-0.4, -4), V(37, cinqueTerrePathY+1.2, 15)}
	g.Use(Wood).Fill(Func{box, func(p Vec3) bool {
		d := c.pathDist(p.X, p.Z)
		if d < 1.3 || d > 1.8 || s.height(p.X, p.Z) > cinqueTerrePathY-0.2 {
			return false
		}
		post := math.Mod(math.Abs(p.X+p.Z), 2.4) < 0.7
		return p.Y > cinqueTerrePathY+0.7 || (post && p.Y > cinqueTerrePathY-0.1)
	}})
}

// harbour builds the slipway boats and a few moored boats.
func (c CinqueTerre) harbour(g *Grid, s *cinqueTerreSite) {
	colors := []Color{Red, Navy, White, Gold, Glass, Red, Navy, White}
	spots := []struct {
		p     Vec2
		angle float64
	}{
		{P(-3.2, 2), math.Pi/2 + 0.05}, {P(-0.6, 3), math.Pi/2 + 0.1}, {P(1.4, 0.5), math.Pi/2 - 0.1},
		{P(-3, -2.5), math.Pi/2 + 0.15}, {P(0.2, -3.5), math.Pi / 2}, {P(-6.5, 12.5), 0.3}, {P(5.5, 15), -0.4},
		{P(-12, 15), 0.1}, {P(-2, 9), 1.75}, {P(2.5, 11.5), 1.2},
	}
	// Boats parked along the main street.
	for k := 0; k < 3; k++ {
		z := -5.5 - 4.4*float64(k)
		side := 0.8
		if k%2 == 1 {
			side = -0.8
		}
		x := c.ravine(z) + side
		spots = append(spots, struct {
			p     Vec2
			angle float64
		}{P(x, z), math.Pi/2 + 0.08*side})
	}
	for i, sp := range spots {
		y := math.Max(s.height(sp.p.X, sp.p.Y), 0.1)
		if sp.p.Y < -5 {
			y = s.height(sp.p.X, sp.p.Y+2.5)
		}
		c.boat(g, sp.p, y, sp.angle, colors[i%len(colors)], i)
	}
}

// boat builds a double-ended wooden gozzo along angle (radians from +X) at height y.
func (CinqueTerre) boat(g *Grid, at Vec2, y, angle float64, color Color, seed int) {
	L, W := 2.8, 1.25
	hull := func(p Vec3, inset float64) bool {
		q := p.Plan().Sub(at).Rotate(-angle)
		t := q.X / L
		if math.Abs(t) > 1 {
			return false
		}
		k := (p.Y - y) / 1.3
		if k < 0 || k > 1 {
			return false
		}
		hw := W*math.Sqrt(1-t*t)*(0.55+0.45*k) - inset
		return math.Abs(q.Y) <= hw
	}
	box := Box{V(at.X-L-1, y, at.Y-L-1), V(at.X+L+1, y+1.4, at.Y+L+1)}
	g.With(color, func() { g.FillVolume(Func{box, func(p Vec3) bool { return hull(p, 0) }}) })
	g.Paint(Wood, Func{box, func(p Vec3) bool { return p.Y > y+1.0 && hull(p, 0) }})
	g.Carve(Func{box, func(p Vec3) bool { return p.Y > y+0.6 && hull(p, 0.55) }})
}

// rocks are sea stacks off the headlands, ringed with surf.
func (c CinqueTerre) rocks(g *Grid, s *cinqueTerreSite) {
	stacks := []struct {
		at   Vec2
		r, h float64
	}{{P(-28.5, 6.5), 2.4, 3.6}, {P(-27.5, 13.5), 1.6, 2.2}, {P(26, 15.5), 2.2, 3.2}, {P(9.5, 17), 1.5, 1.8}, {P(28, 10), 1.8, 2.4}}
	for i, st := range stacks {
		at, r, h := st.at, st.r, st.h
		box := Box{V(at.X-r-1, -0.2, at.Y-r-1), V(at.X+r+1, h+1, at.Y+r+1)}
		g.Use(DarkStone).FillVolume(Func{box, func(p Vec3) bool {
			q := p.Plan().Sub(at)
			k := p.Y / h
			rr := r * (1 - 0.55*k*k) * (1 + 0.25*cinqueTerreNoise(q.Angle()*1.5+float64(i)*7, p.Y*0.7, 40+i))
			return p.Y >= -0.2 && p.Y <= h && q.Length() <= rr
		}})
		g.Paint(Charcoal, Func{box, func(p Vec3) bool { return Hash01(int(math.Floor(p.Y/0.8)), i, 41) < 0.35 }})
		g.Paint(White, Func{box.Grow(1.5), func(p Vec3) bool {
			if p.Y > 0.25 || s.height(p.X, p.Z) >= 0 {
				return false
			}
			d := p.Plan().Sub(at).Length()
			return d > r*0.9 && d < r+1.3 && Hash01(int(math.Floor(p.X*1.5)), int(math.Floor(p.Z*1.5)), 42) < 0.6
		}})
	}
}

// trees are olive and lemon trees dotted over the scrub and the village edges.
func (c CinqueTerre) trees(g *Grid, s *cinqueTerreSite) {
	n := 0
	for k := 0; k < 400 && n < 22; k++ {
		x := cinqueTerreX0 + 4 + (cinqueTerreX1-cinqueTerreX0-8)*Hash01(k, 51)
		z := cinqueTerreZ0 + 4 + (cinqueTerreZ1-cinqueTerreZ0-8)*Hash01(k, 52)
		h, kind := s.at(x, z)
		if kind != cinqueTerreScrub || h < 3 || g.HasAt(V(x, h+1.2, z)) ||
			math.Hypot(x-cinqueTerreTowerAt.X, z-cinqueTerreTowerAt.Y) < 6 {
			continue
		}
		size := 1.0 + 0.5*Hash01(k, 53)
		g.With(Wood, func() { g.Fill(Cylinder{Center: P(x, z), Radius: 0.5, Y0: h - 0.5, Y1: h + 1.6*size}) })
		crown := Foliage
		if Hash01(k, 54) < 0.4 {
			crown = Grass
		}
		g.With(crown, func() {
			g.FillVolume(Ellipsoid{V(x, h+2.2*size, z), V(1.5*size, 1.1*size, 1.5*size)})
		})
		n++
	}
}
