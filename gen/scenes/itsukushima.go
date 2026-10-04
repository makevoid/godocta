package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// Itsukushima is the shrine on Miyajima at high tide. The great ryobu-style O-torii
// (two main vermilion pillars braced by four auxiliary legs, the nuki tie beam, the
// curved black kasagi lintel and the name plaque) stands in the sea in front. Behind
// it the shrine floats on a grid of piers: the hirabutai platform with the takabutai
// dance stage and the hitasaki pier, the haraiden, haiden and honden with sweeping
// cypress-bark roofs, the zig-zag vermilion kairo corridors running east and west to
// the shore, the Marodo sub-shrine, the noh stage, and the arched Sorihashi bridge.
// A festival barge floats beside the shrine. Stone lanterns, shore pines and deer line
// the beach, where the Mitarai stream comes down the maple valley under a footbridge;
// the five-storey pagoda and the Senjokaku hall stand on a hill terrace to the east and
// the Tahoto on a terrace to the west. Mt. Misen's wooded slopes, speckled with autumn
// maples and crowned by granite boulders, rise behind and fall to a low rim.
type Itsukushima struct{}

func init() { Register(Itsukushima{}) }

func (Itsukushima) Name() string  { return "itsukushima" }
func (Itsukushima) Title() string { return "Itsukushima" }

const (
	itsukushimaWater = 0.4   // Top of the sea layer.
	itsukushimaDeck  = 2.0   // Top of the shrine's floor boards.
	itsukushimaHX    = 44.0  // Site half-width (X).
	itsukushimaZ0    = -57.0 // Back (north) edge of the site.
	itsukushimaZ1    = 35.0  // Front (south, seaward) edge of the site.
	itsukushimaTorii = 25.0  // Plan Z of the O-torii; it stands on the axis X = 0.
)

// Roof kinds.
const (
	itsukushimaGable   = iota // Kirizuma: plain gable.
	itsukushimaHip            // Yosemune: hipped.
	itsukushimaIrimoya        // Hip-and-gable.
)

// itsukushimaCorridors are the kairo center lines over the water.
var itsukushimaCorridors = [][]Vec2{
	// East corridor: from the main complex out to the shore entrance.
	{P(10, 0), P(16, 0), P(16, -4), P(29, -4), P(29, -10), P(35, -10), P(35, -20)},
	// West corridor: past the noh stage to the shore exit.
	{P(-10, 0), P(-18, 0), P(-18, -5), P(-29, -5), P(-29, -12), P(-35, -12), P(-35, -20)},
	// Sajiki: the side galleries flanking the hirabutai.
	{P(13.5, 0), P(13.5, 13)},
	{P(-13.5, 0), P(-13.5, 13)},
}

// itsukushimaPlatforms are pier-supported floors: min, max corners in plan and top.
var itsukushimaPlatforms = []struct {
	Min, Max Vec2
	Top      float64
}{
	{P(-11, -18), P(11, 4.5), itsukushimaDeck},         // Main complex.
	{P(-13, 4.5), P(13, 14), itsukushimaDeck - 0.4},    // Hirabutai.
	{P(-2.2, 14), P(2.2, 19.5), itsukushimaDeck - 0.4}, // Hitasaki pier.
	{P(16.5, -15.5), P(27.5, -5.5), itsukushimaDeck},   // Marodo shrine.
	{P(-32, -3.4), P(-19.5, 9.5), itsukushimaDeck},     // Noh stage court.
	{P(-20, -8), P(-16, -6.6), itsukushimaDeck},        // Sorihashi landing.
	{P(-27.5, -15), P(-21, -6.8), itsukushimaDeck},     // Tenjin-sha.
	{P(18, -2.2), P(28.5, 5.5), itsukushimaDeck},       // Asazaya.
}

func (it Itsukushima) Build(g *Grid) Order {
	it.terrain(g)
	for _, pl := range itsukushimaPlatforms {
		it.platform(g, pl.Min, pl.Max, pl.Top)
	}
	for _, line := range itsukushimaCorridors {
		it.corridor(g, line, 2.0)
	}
	it.mainShrine(g)
	it.marodo(g)
	it.subShrines(g)
	it.tahoto(g, itsukushimaTahoto)
	it.gozabune(g, P(21, 15), -0.45)
	it.archBridge(g, P(-17.2, -24), P(-10.8, -24), 1.3)
	it.noh(g)
	it.sorihashi(g)
	it.torii(g)
	it.shoreLanterns(g)
	it.pagoda(g, P(24.5, -31))
	it.senjokaku(g, P(9.5, -30.5))
	it.trees(g)
	it.summit(g)
	it.cedars(g)
	it.deer(g)
	return Outward(P(0, 0))
}

func (Itsukushima) sm(e0, e1, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-e0)/(e1-e0)))
	return t * t * (3 - 2*t)
}

func (Itsukushima) clamp01(v float64) float64 { return math.Max(0, math.Min(1, v)) }

func (Itsukushima) h3(p Vec3, seed int) float64 {
	return Hash01(int(math.Round(p.X*2)), int(math.Round(p.Y*2)), int(math.Round(p.Z*2)), seed)
}

// shore is the plan Z of the waterline: the shrine sits in a shallow cove whose
// sides reach forward.
func (it Itsukushima) shore(x float64) float64 {
	return -19.5 + 9*it.sm(24, 44, math.Abs(x)) + 0.8*math.Sin(x*0.21+0.7)
}

// plateau is how far a point lies on a level hill terrace (0..1): the pagoda hill
// to the east, or the Tahoto terrace cut into the western slope.
func (it Itsukushima) plateau(x, z float64) float64 {
	return math.Max(it.eastHill(x, z), it.westHill(x, z))
}

func (it Itsukushima) eastHill(x, z float64) float64 {
	dx := math.Max(0, math.Abs(x-16)-13)
	dz := math.Max(0, math.Abs(z+30.5)-5.5)
	return it.sm(5, 0, math.Hypot(dx, dz))
}

func (it Itsukushima) westHill(x, z float64) float64 {
	return it.sm(9, 5, math.Hypot(x-itsukushimaTahoto.X, z-itsukushimaTahoto.Y))
}

// itsukushimaStream is the Mitarai stream's course from the maple valley to the sea.
var itsukushimaStream = []Vec2{P(-10, -46), P(-12, -36), P(-13, -29), P(-14, -23), P(-14.5, -17)}

// streamDist is the plan distance to the stream's center line.
func (Itsukushima) streamDist(x, z float64) float64 {
	best := math.Inf(1)
	p := P(x, z)
	for i := 1; i < len(itsukushimaStream); i++ {
		a, b := itsukushimaStream[i-1], itsukushimaStream[i]
		ab := b.Sub(a)
		t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
		best = math.Min(best, p.Sub(a.Add(ab.Scale(t))).Length())
	}
	return best
}

// itsukushimaTahoto is the plan site of the two-storey Tahoto pagoda.
var itsukushimaTahoto = P(-27, -31)

// raw is the land height above the sea before the rim fade.
func (it Itsukushima) raw(x, z float64) float64 {
	e := it.shore(x) - z
	beach := 1.3 * it.sm(0, 4, e)
	// Mt. Misen: slopes rising to a broken ridge at the back.
	m := 27*it.sm(-20, -43, z)*math.Exp(-math.Pow((x+4)/22, 2)) + 9*it.sm(-20, -40, z)
	m += 2.4 * math.Sin(x*0.19+1.1) * math.Sin(z*0.23+0.4) * it.sm(-24, -36, z)
	// Momijidani: the maple valley cutting down behind the shrine's west side.
	m -= 6 * math.Exp(-math.Pow((x+12)/6, 2)) * it.sm(-24, -40, z) * it.sm(-58, -46, z)
	m = math.Max(0, m) * it.sm(4, 10, e)
	// The pagoda hill: a level terrace at +6.5 with slopes.
	h := math.Max(m, 6.5*it.eastHill(x, z))
	w := it.westHill(x, z)
	h = beach + h*(1-w) + 7*w
	// The Mitarai stream's ravine down the maple valley.
	return h - math.Min(h, 1.6)*it.sm(2.8, 0.9, it.streamDist(x, z))
}

// rim fades the land down to a low edge around the site.
func (it Itsukushima) rim(x, z, h float64) float64 {
	e := math.Min(itsukushimaHX-math.Abs(x), z-itsukushimaZ0)
	return it.sm(0, 2.5+0.45*h, e)
}

// ground is the land surface, or NaN over the sea.
func (it Itsukushima) ground(x, z float64) float64 {
	if z > it.shore(x) {
		return math.NaN()
	}
	h := it.raw(x, z)
	return itsukushimaWater - 0.2 + (h+0.2)*it.rim(x, z, h)
}

// forest is how strongly woodland covers a plan point.
func (it Itsukushima) forest(x, z float64) float64 {
	h := it.raw(x, z)
	f := it.sm(2.0, 3.6, h) * (1 - it.plateau(x, z)) * it.sm(1.5, 3.5, it.streamDist(x, z))
	return f
}

// canopy is the woodland crown height over a point and the crown's species key.
func (it Itsukushima) canopy(x, z float64) (float64, int) {
	const cs = 3.4
	ci, ck := int(math.Floor(x/cs)), int(math.Floor(z/cs))
	best, kind := math.Inf(-1), -1
	here := it.ground(x, z)
	if math.IsNaN(here) {
		return best, kind
	}
	for di := -1; di <= 1; di++ {
		for dk := -1; dk <= 1; dk++ {
			i, k := ci+di, ck+dk
			tx := (float64(i) + 0.15 + 0.7*Hash01(i, k, 1)) * cs
			tz := (float64(k) + 0.15 + 0.7*Hash01(i, k, 2)) * cs
			f := it.forest(tx, tz)
			if f < 0.25 {
				continue
			}
			r := (2.4 + 1.3*Hash01(i, k, 3)) * (0.6 + 0.4*f)
			d := math.Hypot(x-tx, z-tz)
			if d >= r {
				continue
			}
			rimF := it.rim(tx, tz, it.raw(tx, tz))
			top := here + ((0.8+1.6*Hash01(i, k, 4))+1.1*r*math.Sqrt(1-d*d/(r*r)))*f*(0.4+0.6*rimF)
			if top > best {
				best, kind = top, int(Hash01(i, k, 5)*1000)
				// More maples down the Momijidani valley.
				if math.Abs(tx+12) < 7 && tz < -24 && Hash01(i, k, 6) < 0.35 {
					kind = 900 + kind%100
				}
			}
		}
	}
	return best, kind
}

// surface is the top of the land: ground or tree crowns (NaN over the sea).
func (it Itsukushima) surface(x, z float64) float64 {
	gy := it.ground(x, z)
	if math.IsNaN(gy) {
		return gy
	}
	c, _ := it.canopy(x, z)
	return math.Max(gy, c)
}

// terrain lays the sea and a hollow skin of beach, hill and woodland canopy.
func (it Itsukushima) terrain(g *Grid) {
	sea := Box{V(-itsukushimaHX, 0, itsukushimaZ0), V(itsukushimaHX, itsukushimaWater, itsukushimaZ1)}
	g.Use(Water).Fill(Func{sea, func(p Vec3) bool {
		x := p.X / itsukushimaHX
		return p.Z > it.shore(p.X)-1.5 && p.Z < itsukushimaZ1-9*x*x
	}})
	// Glints and a paler band of shallows along the beach.
	g.Paint(Glass, Func{sea, func(p Vec3) bool {
		c := Nearest(p.Scale(1 / g.Span()))
		if Hash01(int(c.X), int(c.Z), 6) < 0.025 {
			return true
		}
		return p.Z < it.shore(p.X)+2.2 && Hash01(int(c.X), int(c.Z), 7) < 0.6
	}})

	land := NewGridWithSpan(g.Span())
	box := Box{V(-itsukushimaHX, 0, itsukushimaZ0), V(itsukushimaHX, 40, -5)}
	land.Use(Grass).Fill(Func{box, func(p Vec3) bool {
		s := it.surface(p.X, p.Z)
		if math.IsNaN(s) || p.Y > s {
			return false
		}
		lo := s
		for _, o := range []Vec2{P(0.9, 0), P(-0.9, 0), P(0, 0.9), P(0, -0.9)} {
			n := it.surface(p.X+o.X, p.Z+o.Y)
			if math.IsNaN(n) {
				n = 0
			}
			lo = math.Min(lo, n)
		}
		return p.Y >= math.Max(0, lo-1.1)
	}})
	for _, c := range land.Cells() {
		p := land.World(c)
		gy := it.ground(p.X, p.Z)
		top, kind := it.canopy(p.X, p.Z)
		e := it.shore(p.X) - p.Z
		col := Grass
		switch {
		case top > gy+0.4:
			v := Hash01(int(c.X), int(c.Y), int(c.Z))
			switch {
			case kind >= 900:
				// Autumn maples: speckled crowns of red, gold and rust.
				switch {
				case kind%4 == 0 && v < 0.6:
					col = Gold
				case v < 0.55:
					col = Red
				case v < 0.75:
					col = Terracotta
				default:
					col = Foliage
				}
			case kind < 640:
				col = Foliage
				if v < 0.1 {
					col = Grass
				}
			default:
				col = Grass
				if v < 0.15 {
					col = Foliage
				}
			}
			if v > 0.985 {
				col = Red
			} else if v > 0.975 {
				col = Gold
			}
		case it.streamDist(p.X, p.Z) < 1.0 && p.Y > gy-0.9:
			col = Water
		case it.streamDist(p.X, p.Z) < 1.9 && p.Y > gy-0.9:
			col = Stone
			if Hash01(int(c.X), int(c.Z), 5) < 0.4 {
				col = DarkStone
			}
		case math.Abs(p.X-10) < 1.7 && it.plateau(p.X, p.Z) > 0.03 && it.plateau(p.X, p.Z) < 0.97:
			col = Stone // Stone stair up to the Senjokaku.
			if int(math.Floor(p.Y/0.8))%2 == 0 {
				col = DarkStone
			}
		case e < 5.5 && gy < 2.2:
			col = Sand
			if e < 1.2 && Hash01(int(c.X), int(c.Z), 3) < 0.3 {
				col = Stone
			}
		case it.plateau(p.X, p.Z) > 0.96:
			col = Sand // Gravel court on the hill.
			if Hash01(int(c.X), int(c.Z), 4) < 0.25 {
				col = Stone
			}
		case Hash01(int(c.X), int(c.Y), int(c.Z), 3) < 0.15:
			col = Foliage
		}
		g.Use(col).Add(c)
	}
}

// seg gives a plan point's distance along a segment, its signed offset across it, and
// the segment length.
func (Itsukushima) seg(a, b Vec2, x, z float64) (along, across, length float64) {
	ab := b.Sub(a)
	length = ab.Length()
	dir := ab.Scale(1 / length)
	d := P(x, z).Sub(a)
	return d.Dot(dir), dir.X*d.Y - dir.Y*d.X, length
}

// segBox is the plan bounds of a segment grown by r, from y0 to y1.
func (Itsukushima) segBox(a, b Vec2, r, y0, y1 float64) Box {
	return Box{V(math.Min(a.X, b.X)-r, y0, math.Min(a.Y, b.Y)-r), V(math.Max(a.X, b.X)+r, y1, math.Max(a.Y, b.Y)+r)}
}

// piers fills the dark posts that carry a floor: a grid under a plan rectangle.
func (it Itsukushima) piers(g *Grid, lo, hi Vec2, top float64) {
	nx := int(math.Max(1, math.Round((hi.X-lo.X-1)/2.1)))
	nz := int(math.Max(1, math.Round((hi.Y-lo.Y-1)/2.1)))
	for i := 0; i <= nx; i++ {
		for k := 0; k <= nz; k++ {
			x := lo.X + 0.5 + (hi.X-lo.X-1)*float64(i)/float64(nx)
			z := lo.Y + 0.5 + (hi.Y-lo.Y-1)*float64(k)/float64(nz)
			it.post(g, P(x, z), top)
		}
	}
}

// post is one pier post with a stone footing, standing in the sea or on the beach.
func (it Itsukushima) post(g *Grid, at Vec2, top float64) {
	base := 0.0
	if gy := it.ground(at.X, at.Y); !math.IsNaN(gy) {
		base = math.Max(0, gy-0.5)
	}
	g.Use(Charcoal).Fill(Cylinder{Center: at, Radius: 0.5, Y0: base, Y1: top})
	g.Use(DarkStone).Fill(Cylinder{Center: at, Radius: 0.7, Y0: base, Y1: itsukushimaWater + 0.5})
}

// platform is a wooden floor on piers with a vermilion edge beam.
func (it Itsukushima) platform(g *Grid, lo, hi Vec2, top float64) {
	it.piers(g, lo, hi, top-0.7)
	g.Use(Wood).Fill(Block(lo.At(top-0.7), hi.At(top)))
	g.Paint(Red, Subtract(Block(lo.At(top-0.75), hi.At(top-0.35)), Block(lo.Add(P(0.6, 0.6)).At(top-1), hi.Sub(P(0.6, 0.6)).At(top))))
}

// corridor is a kairo gallery along a polyline: floor on piers, vermilion posts and
// head beams, a white frieze, and a gabled cypress-bark roof.
func (it Itsukushima) corridor(g *Grid, line []Vec2, half float64) {
	const roofHalf = 2.9
	deck := itsukushimaDeck
	head := deck + 2.9
	for i := 1; i < len(line); i++ {
		a, b := line[i-1], line[i]
		_, _, L := it.seg(a, b, a.X, a.Y)
		box := it.segBox(a, b, roofHalf+0.5, -1, deck+7)
		in := func(p Vec3, ext, w float64) (float64, float64, bool) {
			s, t, _ := it.seg(a, b, p.X, p.Z)
			return s, t, s >= -ext && s <= L+ext && math.Abs(t) <= w
		}
		// Floor and piers.
		g.Use(Wood).Fill(Func{box, func(p Vec3) bool {
			_, _, ok := in(p, half, half)
			return ok && p.Y >= deck-0.7 && p.Y <= deck
		}})
		n := int(math.Max(1, math.Round(L/2.6)))
		dir := b.Sub(a).Scale(1 / L)
		nrm := P(-dir.Y, dir.X)
		for k := 0; k <= n; k++ {
			c := a.Add(dir.Scale(L * float64(k) / float64(n)))
			for _, side := range []float64{-1, 1} {
				at := c.Add(nrm.Scale(side * (half - 0.5)))
				it.post(g, at, deck-0.7)
				g.Use(Red).Fill(Cylinder{Center: at, Radius: 0.5, Y0: deck, Y1: head})
			}
			// Hanging bronze lanterns between alternate posts.
			if k < n && k%2 == 0 {
				m := a.Add(dir.Scale(L * (float64(k) + 0.5) / float64(n)))
				for _, side := range []float64{-1, 1} {
					at := m.Add(nrm.Scale(side * (half - 0.5)))
					g.Use(Gold).Fill(BlockAt(at, 0.9, 0.9, head-1.7, head-0.9))
					g.Use(Charcoal).Fill(BlockAt(at, 1.1, 1.1, head-0.9, head-0.6))
				}
			}
		}
		// Head beams, white frieze and a low vermilion floor rail.
		g.Use(Red).Fill(Func{box, func(p Vec3) bool {
			_, t, ok := in(p, half-0.5, half)
			return ok && math.Abs(t) >= half-0.95 && p.Y >= head-0.7 && p.Y <= head
		}})
		g.Use(White).Fill(Func{box, func(p Vec3) bool {
			_, t, ok := in(p, half-0.5, half)
			return ok && math.Abs(t) >= half-0.95 && p.Y > head && p.Y <= head+0.7
		}})
		// Gabled roof: concave, flatter at the eaves.
		eave, rise := head+1.1, 1.6
		g.Use(Charcoal).FillVolume(Func{box, func(p Vec3) bool {
			_, t, ok := in(p, roofHalf, roofHalf)
			if !ok {
				return false
			}
			w := 1 - math.Abs(t)/roofHalf
			s := eave + rise*math.Pow(w, 1.4)
			return p.Y <= s && p.Y >= s-0.95
		}})
		g.Paint(Wood, Func{box, func(p Vec3) bool {
			_, t, ok := in(p, roofHalf, roofHalf)
			return ok && math.Abs(t) > roofHalf-0.45 && p.Y < eave+0.4
		}})
	}
}

// roofAt is a Japanese roof height over a plan point: concave slopes from yE at the
// eave line (the body plus `over`) up `rise`, with lightly swept corners. It returns
// the height and the distances from the long eave (across) and from the end (ends).
func (it Itsukushima) roofAt(c Vec2, hx, hz, over, yE, rise, up float64, ridge Axis, kind int, x, z float64) (s, across, ends, depth float64, ok bool) {
	ex, ez := hx+over-math.Abs(x-c.X), hz+over-math.Abs(z-c.Y)
	if ex < 0 || ez < 0 {
		return 0, 0, 0, 0, false
	}
	across, ends, depth = ez, ex, hz+over
	if ridge == AxisZ {
		across, ends, depth = ex, ez, hx+over
	}
	f := func(e float64) float64 { return rise * math.Pow(it.clamp01(e/depth), 1.45) }
	switch kind {
	case itsukushimaGable:
		s = yE + f(across)
	case itsukushimaHip:
		s = yE + f(math.Min(across, ends))
	default:
		if ends > 0.5*depth {
			s = yE + f(across)
		} else {
			s = yE + f(math.Min(across, ends))
		}
	}
	if kind != itsukushimaGable {
		near := it.clamp01(1 - math.Max(across, ends)/(0.4*depth))
		s += up * near * near
	}
	return s, across, ends, depth, true
}

// roof fills a cypress-bark roof with a wooden eave edge, a ridge, and white gable
// pediments with vermilion bargeboards.
func (it Itsukushima) roof(g *Grid, c Vec2, hx, hz, over, yE, rise, up float64, ridge Axis, kind int, wallTop float64, col Color) {
	r := over + 0.5
	box := Box{V(c.X-hx-r, yE-2, c.Y-hz-r), V(c.X+hx+r, yE+rise+up+2, c.Y+hz+r)}
	at := func(p Vec3) (float64, float64, float64, float64, bool) {
		return it.roofAt(c, hx, hz, over, yE, rise, up, ridge, kind, p.X, p.Z)
	}
	// Pediments: under the gable at the irimoya break, or above the end walls.
	g.Use(White).Fill(Func{box, func(p Vec3) bool {
		s, across, ends, depth, ok := at(p)
		if !ok || p.Y > s-1.0 {
			return false
		}
		if kind == itsukushimaIrimoya {
			lo := yE + rise*math.Pow(it.clamp01(math.Min(across, 0.5*depth)/depth), 1.45)
			return ends >= 0.5*depth && ends <= 0.5*depth+0.9 && p.Y >= lo-0.4
		}
		if kind == itsukushimaGable {
			return ends >= over && ends <= over+0.9 && p.Y >= wallTop-0.2
		}
		return false
	}})
	g.Use(col).FillVolume(Func{box, func(p Vec3) bool {
		s, _, _, _, ok := at(p)
		return ok && p.Y <= s && p.Y >= s-1.0
	}})
	// Bargeboards along the gable edges.
	if kind == itsukushimaGable {
		g.Paint(Red, Func{box, func(p Vec3) bool {
			s, _, ends, _, ok := at(p)
			return ok && p.Y >= s-1.1 && ends <= 0.5
		}})
	}
	// Eave edge.
	g.Paint(Wood, Func{box, func(p Vec3) bool {
		s, across, ends, _, ok := at(p)
		if !ok {
			return false
		}
		e := across
		if kind != itsukushimaGable {
			e = math.Min(across, ends)
		}
		return e < 0.45 && p.Y < s+0.1
	}})
	// Ridge.
	top := yE + rise
	var a, b Vec2
	l := hx + over
	if kind == itsukushimaIrimoya {
		l = hx + over - 0.5*(hz+over)
	}
	if ridge == AxisZ {
		l = hz + over
		if kind == itsukushimaIrimoya {
			l = hz + over - 0.5*(hx+over)
		}
		a, b = P(c.X, c.Y-l+0.3), P(c.X, c.Y+l-0.3)
	} else {
		a, b = P(c.X-l+0.3, c.Y), P(c.X+l-0.3, c.Y)
	}
	if kind != itsukushimaHip {
		g.Use(Charcoal).Fill(Capsule{a.At(top), b.At(top), 0.6})
	}
}

// hall is a shrine hall body: white plaster shell with a dark lining, vermilion
// columns, sill and head beams, and dark door bays.
func (it Itsukushima) hall(g *Grid, c Vec2, hx, hz, y0, y1 float64) {
	g.Use(Charcoal).Fill(Subtract(BlockAt(c, 2*hx-1.6, 2*hz-1.6, y0, y1), BlockAt(c, 2*hx-3.4, 2*hz-3.4, y0+0.6, y1-0.6)))
	g.Use(White).Fill(Subtract(BlockAt(c, 2*hx, 2*hz, y0, y1), BlockAt(c, 2*hx-1.8, 2*hz-1.8, y0-1, y1+1)))
	shell := Subtract(BlockAt(c, 2*hx+0.2, 2*hz+0.2, y0, y1), BlockAt(c, 2*hx-1.8, 2*hz-1.8, y0-1, y1+1))
	g.Use(Red).Fill(Intersect(shell, Union(
		Block(V(c.X-hx-1, y0, c.Y-hz-1), V(c.X+hx+1, y0+0.6, c.Y+hz+1)),
		Block(V(c.X-hx-1, y1-0.8, c.Y-hz-1), V(c.X+hx+1, y1, c.Y+hz+1)),
	)))
	nx := int(math.Max(1, math.Round(2*hx/3.2)))
	nz := int(math.Max(1, math.Round(2*hz/3.2)))
	bayX, bayZ := 2*hx/float64(nx), 2*hz/float64(nz)
	// Door and window bays between the columns.
	for i := 0; i < nx; i++ {
		x := c.X - hx + bayX*(float64(i)+0.5)
		g.Carve(BlockAt(P(x, c.Y), bayX-1.6, 2*hz+1, y0+0.9, y1-1.3))
	}
	for k := 0; k < nz; k++ {
		z := c.Y - hz + bayZ*(float64(k)+0.5)
		g.Carve(BlockAt(P(c.X, z), 2*hx+1, bayZ-1.6, y0+0.9, y1-1.3))
	}
	// Re-line the bays so openings read dark, with a vermilion lattice rail.
	g.Use(Charcoal).Fill(Subtract(BlockAt(c, 2*hx-1.6, 2*hz-1.6, y0, y1), BlockAt(c, 2*hx-3.4, 2*hz-3.4, y0-1, y1+1)))
	g.Use(Red).Fill(Subtract(BlockAt(c, 2*hx-0.6, 2*hz-0.6, y0+2.4, y1-2.0), BlockAt(c, 2*hx-1.6, 2*hz-1.6, y0-1, y1+1)))
	for i := 0; i <= nx; i++ {
		x := c.X - hx + bayX*float64(i)
		for _, z := range []float64{c.Y - hz, c.Y + hz} {
			g.Use(Red).Fill(Cylinder{Center: P(x, z), Radius: 0.6, Y0: y0, Y1: y1})
		}
	}
	for k := 0; k <= nz; k++ {
		z := c.Y - hz + bayZ*float64(k)
		for _, x := range []float64{c.X - hx, c.X + hx} {
			g.Use(Red).Fill(Cylinder{Center: P(x, z), Radius: 0.6, Y0: y0, Y1: y1})
		}
	}
}

// openHall is a pillared hall open on all sides, with a dark ceiling under the roof.
func (it Itsukushima) openHall(g *Grid, c Vec2, hx, hz, y0, y1 float64, col Color) {
	nx := int(math.Max(1, math.Round(2*hx/2.8)))
	nz := int(math.Max(1, math.Round(2*hz/2.8)))
	for i := 0; i <= nx; i++ {
		for k := 0; k <= nz; k++ {
			if i != 0 && i != nx && k != 0 && k != nz {
				continue
			}
			x := c.X - hx + 2*hx*float64(i)/float64(nx)
			z := c.Y - hz + 2*hz*float64(k)/float64(nz)
			g.Use(col).Fill(Cylinder{Center: P(x, z), Radius: 0.55, Y0: y0, Y1: y1})
		}
	}
	g.Use(col).Fill(Subtract(BlockAt(c, 2*hx+0.6, 2*hz+0.6, y1-0.7, y1), BlockAt(c, 2*hx-1.0, 2*hz-1.0, y1-1, y1+1)))
	g.Use(Charcoal).Fill(BlockAt(c, 2*hx-0.4, 2*hz-0.4, y1-0.3, y1+0.3))
}

// rail is a low vermilion railing around a plan rectangle, open where gaps say.
func (it Itsukushima) rail(g *Grid, lo, hi Vec2, y float64, open func(x, z float64) bool) {
	outer := Block(lo.At(y), hi.At(y+1.1))
	inner := Block(lo.Add(P(0.8, 0.8)).At(y-1), hi.Sub(P(0.8, 0.8)).At(y+2))
	g.Use(Red).Fill(Func{outer.Bounds(), func(p Vec3) bool {
		if !outer.Contains(p) || inner.Contains(p) || (open != nil && open(p.X, p.Z)) {
			return false
		}
		return p.Y >= y+0.7 || math.Mod(math.Abs(p.X+p.Z)+100, 1.6) < 0.8
	}})
}

// mainShrine builds the honden, heiden, haiden, haraiden, the hirabutai with the
// takabutai stage and the hitasaki pier with its lantern.
func (it Itsukushima) mainShrine(g *Grid) {
	d := itsukushimaDeck
	// Honden: the long main hall with a gabled roof.
	it.hall(g, P(0, -13.5), 8.5, 3.6, d, d+4.8)
	it.roof(g, P(0, -13.5), 8.5, 3.6, 2.0, d+4.5, 4.6, 0, AxisX, itsukushimaGable, d+4.8, Charcoal)
	// Heiden link and haiden.
	it.hall(g, P(0, -8.3), 2.8, 1.8, d, d+4.0)
	it.roof(g, P(0, -8.3), 2.8, 1.8, 0.8, d+4.0, 2.0, 0, AxisZ, itsukushimaHip, d+4.0, Charcoal)
	it.hall(g, P(0, -3.8), 10, 2.8, d, d+4.4)
	it.roof(g, P(0, -3.8), 10, 2.8, 2.0, d+4.0, 4.4, 0.7, AxisX, itsukushimaIrimoya, d+4.4, Charcoal)
	// Haraiden: open purification hall, gable toward the sea.
	it.openHall(g, P(0, 2.2), 5.6, 2.3, d, d+4.0, Red)
	it.roof(g, P(0, 2.2), 5.6, 2.3, 1.8, d+3.7, 4.0, 0.7, AxisZ, itsukushimaIrimoya, d+4.0, Charcoal)

	// Hirabutai rail, open on the axis toward the hitasaki.
	hy := d - 0.4
	it.rail(g, P(-13, 4.5), P(13, 14), hy, func(x, z float64) bool {
		return z < 6 || math.Abs(x) > 12 || (math.Abs(x) < 2.4 && z > 12)
	})
	it.rail(g, P(-2.2, 14), P(2.2, 19.5), hy, func(x, z float64) bool { return z < 15 })
	// Takabutai: a raised black-lacquered dance stage with vermilion railings and stairs.
	tc := P(0, 9.2)
	g.Use(Charcoal).Fill(BlockAt(tc, 6.2, 5.2, hy, hy+0.9))
	g.Use(Wood).Fill(BlockAt(tc, 5.4, 4.4, hy+0.9, hy+1.2))
	it.rail(g, P(-3.1, 6.6), P(3.1, 11.8), hy+1.1, func(x, z float64) bool { return math.Abs(x) < 1.1 })
	g.Use(Red).Fill(BlockAt(P(0, 12.5), 2.2, 1.4, hy, hy+0.5), BlockAt(P(0, 5.9), 2.2, 1.4, hy, hy+0.5))
	for _, x := range []float64{-3.1, 3.1} {
		for _, z := range []float64{6.6, 11.8} {
			g.Use(Gold).Fill(Sphere(V(x, hy+2.3, z), 0.6))
		}
	}
	// Hitasaki lantern: a bronze lantern on the pier's tip.
	it.bronzeLantern(g, P(0, 18.2), hy)
	// Tall lanterns (ishidoro) flanking the hirabutai front.
	for _, x := range []float64{-11, 11} {
		it.bronzeLantern(g, P(x, 12.6), hy)
	}
}

// bronzeLantern is a hanging-style bronze lantern on a post.
func (it Itsukushima) bronzeLantern(g *Grid, at Vec2, y float64) {
	g.Use(Charcoal).Fill(Cylinder{Center: at, Radius: 0.5, Y0: y, Y1: y + 2.2})
	g.Use(DarkStone).Fill(BlockAt(at, 1.6, 1.6, y+2.2, y+3.4))
	g.Use(Gold).Fill(BlockAt(at, 1.0, 1.0, y+2.4, y+3.2))
	g.Use(Charcoal).Fill(Frustum{at, 1.4, 0.3, y + 3.4, y + 4.3})
	g.Use(Gold).Fill(Sphere(at.At(y+4.5), 0.45))
}

// marodo is the Marodo sub-shrine on the east corridor.
func (it Itsukushima) marodo(g *Grid) {
	d := itsukushimaDeck
	c := P(25, -12)
	it.hall(g, c, 4.6, 2.6, d, d+3.8)
	it.roof(g, c, 4.6, 2.6, 1.6, d+3.5, 3.4, 0.5, AxisX, itsukushimaIrimoya, d+3.8, Charcoal)
	it.openHall(g, P(25, -7.4), 3.4, 1.4, d, d+3.2, Red)
	it.roof(g, P(25, -7.4), 3.4, 1.4, 1.2, d+3.0, 2.6, 0.4, AxisZ, itsukushimaIrimoya, d+3.2, Charcoal)
}

// noh is the noh stage court: a raised square stage under an irimoya roof with the
// painted pine on its back wall, the hashigakari bridgeway and the gakuya.
func (it Itsukushima) noh(g *Grid) {
	d := itsukushimaDeck
	c := P(-24.5, 5.5)
	sy := d + 0.6
	g.Use(Wood).Fill(BlockAt(c, 6.4, 6.4, d, sy))
	g.Paint(Charcoal, Subtract(BlockAt(c, 6.4, 6.4, d, sy-0.2), BlockAt(c, 5.4, 5.4, d-1, sy+1)))
	it.openHall(g, c, 3.0, 3.0, sy, sy+4.0, Wood)
	// Kagami-ita: the back wall with its pine.
	back := c.Y - 3.0
	g.Use(Wood).Fill(Block(V(c.X-3.0, sy, back-0.5), V(c.X+3.0, sy+4.0, back+0.4)))
	pine := Union(
		Ellipsoid{V(c.X, sy+2.8, back+0.5), V(2.3, 0.8, 0.5)},
		Ellipsoid{V(c.X-1.0, sy+1.6, back+0.5), V(1.6, 0.6, 0.5)},
		Ellipsoid{V(c.X+1.2, sy+1.9, back+0.5), V(1.2, 0.5, 0.5)},
	)
	g.Paint(Foliage, pine)
	g.Use(Foliage).Fill(Intersect(pine, Block(V(c.X-4, sy, back+0.3), V(c.X+4, sy+4, back+0.7))))
	it.roof(g, c, 3.0, 3.0, 1.6, sy+3.8, 3.8, 0.6, AxisZ, itsukushimaIrimoya, sy+4.0, Charcoal)
	// Hashigakari: a narrow covered bridgeway to the gakuya (green room).
	g.Use(Wood).Fill(Block(V(-32.5, d, 1.6), V(-27.5, sy-0.1, 3.6)))
	for _, x := range []float64{-32, -29.8, -27.8} {
		for _, z := range []float64{1.9, 3.3} {
			g.Use(Wood).Fill(Cylinder{Center: P(x, z), Radius: 0.45, Y0: sy - 0.1, Y1: sy + 2.6})
		}
	}
	g.Use(Charcoal).FillVolume(Func{Box{V(-33, sy+2, 0.6), V(-27, sy+4.5, 4.6)}, func(p Vec3) bool {
		s := sy + 3.6 - 0.55*math.Abs(p.Z-2.6)
		return math.Abs(p.Z-2.6) <= 1.9 && p.Y <= s && p.Y >= s-0.9
	}})
	gc := P(-35.2, 2.6)
	it.platform(g, P(-38.4, -1.2), P(-32, 6.2), d)
	it.hall(g, gc, 2.6, 2.8, d, d+3.6)
	it.roof(g, gc, 2.6, 2.8, 1.2, d+3.4, 2.8, 0, AxisX, itsukushimaGable, d+3.6, Charcoal)
}

// sorihashi is the steep arched vermilion bridge from the west corridor to the shore.
func (it Itsukushima) sorihashi(g *Grid) {
	const x = -18.0
	z0, z1 := -7.0, -21.0
	mid, L := (z0+z1)/2, z0-z1
	deck := func(z float64) float64 {
		t := (z - mid) / (L / 2)
		return itsukushimaDeck + 4.2*(1-t*t)
	}
	box := Box{V(x-2.5, 0, z1-1), V(x+2.5, 12, z0+1)}
	g.Use(Wood).Fill(Func{box, func(p Vec3) bool {
		y := deck(p.Z)
		return math.Abs(p.X-x) <= 1.6 && p.Z >= z1 && p.Z <= z0 && p.Y <= y && p.Y >= y-0.8
	}})
	// Vermilion side girders under the deck and railings above it.
	g.Use(Red).Fill(Func{box, func(p Vec3) bool {
		y := deck(p.Z)
		side := math.Abs(math.Abs(p.X-x) - 1.6)
		if side > 0.45 || p.Z < z1 || p.Z > z0 {
			return false
		}
		girder := p.Y <= y-0.3 && p.Y >= y-1.6
		top := p.Y >= y+0.9 && p.Y <= y+1.4
		baluster := p.Y > y && p.Y < y+0.9 && math.Mod(p.Z-z1+100, 1.8) < 0.7
		return girder || top || baluster
	}})
	for _, z := range []float64{z1 + 0.3, z0 - 0.3} {
		for _, s := range []float64{-1, 1} {
			g.Use(Red).Fill(Cylinder{Center: P(x+1.6*s, z), Radius: 0.45, Y0: deck(z), Y1: deck(z) + 1.8})
			g.Use(Gold).Fill(Sphere(V(x+1.6*s, deck(z)+2.0, z), 0.5))
		}
	}
	// Two pairs of piers under the arch.
	for _, z := range []float64{mid - 3.5, mid + 3.5} {
		for _, s := range []float64{-1, 1} {
			it.post(g, P(x+1.1*s, z), deck(z)-0.8)
		}
	}
}

// torii is the great O-torii in the sea: two main pillars with four auxiliary legs,
// the nuki tie beam, the plaque, the shimaki and the curved black kasagi.
func (it Itsukushima) torii(g *Grid) {
	zc := itsukushimaTorii
	const k = 1.25     // Scale over the authored ~17-unit gate.
	const px = 5.6 * k // Main pillar half spacing.
	y := func(v float64) float64 { return v * k }
	// Main pillars with a slight inward lean.
	for _, s := range []float64{-1, 1} {
		g.Use(Red).FillVolume(Func{Box{V(s*px-2.5, 0, zc-2), V(s*px+2.5, y(15.2), zc+2)}, func(p Vec3) bool {
			t := p.Y / y(15)
			cx := s * (px + 0.4 - 0.4*t)
			r := (1.15 - 0.15*t) * k
			return math.Hypot(p.X-cx, p.Z-zc) <= r
		}})
		// A dark waterline band.
		g.Use(Charcoal).FillVolume(Cylinder{Center: P(s*(px+0.4), zc), Radius: 1.4, Y0: 0, Y1: 1.3})
		// Sode-bashira: auxiliary legs fore and aft with small roofs, tied by two beams.
		for _, dz := range []float64{-3.8 * k, 3.8 * k} {
			at := P(s*px, zc+dz)
			g.Use(Red).FillVolume(Cylinder{Center: at, Radius: 0.75, Y0: 0, Y1: y(7.6)})
			g.Use(Charcoal).FillVolume(Cylinder{Center: at, Radius: 0.9, Y0: 0, Y1: 1.0})
			top := y(7.6)
			g.Use(Charcoal).FillVolume(Func{Box{V(at.X-1.5, top-0.2, at.Y-1.5), V(at.X+1.5, top+1.4, at.Y+1.5)}, func(p Vec3) bool {
				r := math.Max(math.Abs(p.X-at.X), math.Abs(p.Z-at.Y))
				return p.Y <= top+1.3-0.7*r && r <= 1.4
			}})
		}
		for _, yy := range []float64{y(3.0), y(6.4)} {
			g.Use(Red).Fill(Block(V(s*px-0.5, yy, zc-4.4*k), V(s*px+0.5, yy+0.9, zc+4.4*k)))
		}
	}
	// Nuki: the lower tie beam passing through the pillars.
	g.Use(Red).Fill(Block(V(-9.4*k, y(10.4), zc-0.6), V(9.4*k, y(10.4)+1.2, zc+0.6)))
	// Plaque post and the name plaque: gold frame on navy with gold characters.
	g.Use(Red).Fill(Block(V(-0.7, y(11.4), zc-0.5), V(0.7, y(14.4), zc+0.5)))
	plaque := Block(V(-1.7, y(11.4)+0.3, zc-0.9), V(1.7, y(11.4)+3.7, zc+0.9))
	g.Use(Gold).Fill(plaque)
	pb := plaque.Min.Y
	g.Use(Navy).Fill(Subtract(Block(V(-1.1, pb+0.5, zc-1.0), V(1.1, pb+2.9, zc+1.0)), Block(V(-1.1, pb, zc-0.6), V(1.1, pb+4, zc+0.6))))
	g.Paint(Gold, Func{plaque.Bounds(), func(p Vec3) bool {
		return math.Abs(p.X) < 0.35 && p.Y > pb+0.8 && p.Y < pb+2.7 && math.Mod(p.Y+10, 0.8) < 0.4
	}})
	// Shimaki and kasagi: vermilion lower lintel, black upper lintel with upswept ends.
	L := 12.5 * k
	lift := func(x float64) float64 { return 1.5 * math.Pow(math.Abs(x)/L, 2.4) }
	sb := y(14.4) + 0.4
	g.Use(Red).FillVolume(Func{Box{V(-L, sb-0.5, zc-0.8), V(L, sb+3, zc+0.8)}, func(p Vec3) bool {
		b := sb + 0.6*lift(p.X)
		return math.Abs(p.X) <= L-1.4 && math.Abs(p.Z-zc) <= 0.75 && p.Y >= b && p.Y <= b+1.1
	}})
	kb := sb + 1.05
	g.Use(Charcoal).FillVolume(Func{Box{V(-L-0.5, kb-0.5, zc-1.5), V(L+0.5, kb+4, zc+1.5)}, func(p Vec3) bool {
		ax := math.Abs(p.X)
		b := kb + lift(p.X)
		top := b + 1.6 + 0.3*it.clamp01((ax-L+3.5)/3.5)
		return ax <= L && math.Abs(p.Z-zc) <= 1.2 && p.Y >= b && p.Y <= top
	}})
}

// stoneLantern is a stone toro on the beach.
func (it Itsukushima) stoneLantern(g *Grid, at Vec2) {
	y := it.ground(at.X, at.Y)
	if math.IsNaN(y) {
		return
	}
	y -= 0.3
	g.Use(Stone).Fill(BlockAt(at, 2.0, 2.0, y, y+0.7))
	g.Fill(Cylinder{Center: at, Radius: 0.55, Y0: y + 0.7, Y1: y + 2.4})
	g.Fill(BlockAt(at, 2.0, 2.0, y+2.4, y+2.9))
	g.Fill(BlockAt(at, 1.5, 1.5, y+2.9, y+4.1))
	g.Paint(Charcoal, BlockAt(at, 1.7, 0.8, y+3.1, y+3.9), BlockAt(at, 0.8, 1.7, y+3.1, y+3.9))
	g.Use(DarkStone).Fill(Func{Box{V(at.X-1.6, y+4.1, at.Y-1.6), V(at.X+1.6, y+5.4, at.Y+1.6)}, func(p Vec3) bool {
		r := math.Max(math.Abs(p.X-at.X), math.Abs(p.Z-at.Y))
		return r <= 1.5 && p.Y <= y+5.1-r*0.6
	}})
	g.Use(Stone).Fill(Sphere(at.At(y+5.4), 0.5))
}

// shoreLanterns lines the beach behind the shrine with stone lanterns.
func (it Itsukushima) shoreLanterns(g *Grid) {
	for x := -33.0; x <= 33; x += 4.4 {
		if math.Abs(x+18) < 3 || math.Abs(x) < 2 {
			continue
		}
		it.stoneLantern(g, P(x, it.shore(x)-2.6))
	}
}

// pagoda is the five-storey pagoda on its stone plinth.
func (it Itsukushima) pagoda(g *Grid, c Vec2) {
	y := it.ground(c.X, c.Y)
	g.Use(Stone).Fill(BlockAt(c, 8.4, 8.4, y-1, y+1.0))
	y += 1.0
	for i := 0; i < 5; i++ {
		hs := 2.9 - 0.28*float64(i)
		h := 2.4
		if i == 0 {
			h = 3.0
		}
		it.storey(g, c, hs, y, y+h)
		over := 2.0 - 0.1*float64(i)
		it.roof(g, c, hs, hs, over, y+h-0.2, 1.7, 0.7, AxisX, itsukushimaHip, y+h, Slate)
		y += h + 0.9
	}
	// Sorin: a bronze finial with gold rings.
	g.Use(Charcoal).Fill(Cylinder{Center: c, Radius: 0.45, Y0: y - 1, Y1: y + 5.5})
	g.Fill(BlockAt(c, 1.6, 1.6, y-0.9, y+0.2))
	for k := 0; k < 5; k++ {
		g.Use(Gold).Fill(Cylinder{Center: c, Radius: 0.85, Y0: y + 0.8 + 0.8*float64(k), Y1: y + 1.2 + 0.8*float64(k)})
	}
	g.Use(Gold).Fill(Sphere(c.At(y+5.9), 0.6))
}

// storey is one pagoda storey: vermilion corner posts and beams with white panels
// and a dark door on each face.
func (it Itsukushima) storey(g *Grid, c Vec2, hs, y0, y1 float64) {
	g.Use(Charcoal).Fill(BlockAt(c, 2*hs-1.0, 2*hs-1.0, y0, y1))
	g.Use(White).Fill(Subtract(BlockAt(c, 2*hs, 2*hs, y0, y1), BlockAt(c, 2*hs-1.6, 2*hs-1.6, y0-1, y1+1)))
	g.Use(Red).Fill(Subtract(BlockAt(c, 2*hs+0.2, 2*hs+0.2, y1-0.7, y1), BlockAt(c, 2*hs-1.6, 2*hs-1.6, y0-1, y1+1)))
	g.Use(Red).Fill(Subtract(BlockAt(c, 2*hs+0.2, 2*hs+0.2, y0, y0+0.5), BlockAt(c, 2*hs-1.6, 2*hs-1.6, y0-1, y1+1)))
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			g.Use(Red).Fill(Cylinder{Center: P(c.X+sx*(hs-0.2), c.Y+sz*(hs-0.2)), Radius: 0.55, Y0: y0, Y1: y1})
		}
	}
	w := math.Min(1.4, hs-1)
	g.Carve(BlockAt(c, 2*w, 2*hs+1, y0+0.5, y1-0.9), BlockAt(c, 2*hs+1, 2*w, y0+0.5, y1-0.9))
	g.Use(Red).Fill(Subtract(BlockAt(c, 2*w+0.6, 2*hs+0.2, y0+0.5, y1-0.7), BlockAt(c, 2*w-0.6, 2*hs+1, y0, y1-1.3), BlockAt(c, 2*w+1, 2*hs-1.6, y0-1, y1+1)))
	g.Use(Red).Fill(Subtract(BlockAt(c, 2*hs+0.2, 2*w+0.6, y0+0.5, y1-0.7), BlockAt(c, 2*hs+1, 2*w-0.6, y0, y1-1.3), BlockAt(c, 2*hs-1.6, 2*w+1, y0-1, y1+1)))
}

// senjokaku is the great unpainted Hall of a Thousand Mats: an open forest of wooden
// pillars on a raised floor under a grey tiled hip-and-gable roof.
func (it Itsukushima) senjokaku(g *Grid, c Vec2) {
	y := it.ground(c.X, c.Y)
	const hx, hz = 7.6, 4.6
	g.Use(Stone).Fill(BlockAt(c, 2*hx+1.6, 2*hz+1.6, y-1, y+0.6))
	for i := 0; i <= 6; i++ {
		for k := 0; k <= 3; k++ {
			x := c.X - hx + 2*hx*float64(i)/6
			z := c.Y - hz + 2*hz*float64(k)/3
			g.Use(Wood).Fill(Cylinder{Center: P(x, z), Radius: 0.45, Y0: y + 0.6, Y1: y + 1.8})
		}
	}
	g.Use(Wood).Fill(BlockAt(c, 2*hx+0.8, 2*hz+0.8, y+1.8, y+2.4))
	it.openHall(g, c, hx, hz, y+2.4, y+6.4, Wood)
	g.Use(Charcoal).Fill(BlockAt(c, 2*hx-2, 2*hz-2, y+2.4, y+2.6))
	it.roof(g, c, hx, hz, 1.9, y+6.1, 5.0, 0.6, AxisX, itsukushimaIrimoya, y+6.4, Slate)
}

// tree is a lone shore pine with a few flat foliage pads.
func (it Itsukushima) pine(g *Grid, at Vec2, h float64, seed int) {
	base := it.ground(at.X, at.Y)
	if math.IsNaN(base) {
		return
	}
	a := Hash01(seed, 1) * 2 * math.Pi
	p0 := at.At(base - 0.3)
	p1 := at.Add(Polar(P(0, 0), h*0.22, a)).At(base + h*0.5)
	p2 := at.Add(Polar(P(0, 0), h*0.3, a+0.5)).At(base + h)
	g.Use(Wood).Fill(Capsule{p0, p1, 0.6}, Capsule{p1, p2, 0.55})
	for k, pd := range []Vec3{p2, p1.Add(Polar(P(0, 0), h*0.3, a+2.4).At(0.6)), p2.Add(Polar(P(0, 0), h*0.25, a-2).At(-1))} {
		r := h * (0.32 - 0.06*float64(k%2))
		g.Use(Wood).Fill(Capsule{p1, pd, 0.45})
		g.Use(Foliage).Fill(Ellipsoid{pd.Add(V(0, 0.4, 0)), V(r, 0.9, r*0.85)})
	}
}

// trees adds shore pines along the beach.
func (it Itsukushima) trees(g *Grid) {
	for i, x := range []float64{-38, -31, -24.5, -13, -6.5, 6.5, 15, 30, 38} {
		z := it.shore(x) - 5.0 - 1.5*Hash01(i, 3)
		it.pine(g, P(x, z), 5+2*Hash01(i, 4), 700+i)
	}
}

// deer places a few sika deer grazing on the beach.
func (it Itsukushima) deer(g *Grid) {
	for i, d := range []Vec3{V(5.5, 0.6, -22.5), V(9, -2.2, -21.8), V(-8, 2.4, -22.8), V(28, 1.0, -21), V(-27, -0.5, -21.5)} {
		at := P(d.X, d.Z)
		y := it.ground(at.X, at.Y)
		if math.IsNaN(y) {
			continue
		}
		dir := Polar(P(0, 0), 1, d.Y)
		side := P(-dir.Y, dir.X)
		body := at.At(y + 1.5)
		g.Use(Wood).Fill(RotateY(Ellipsoid{body, V(1.25, 0.6, 0.5)}, at, d.Y))
		for _, s := range []float64{-1, 1} {
			for _, f := range []float64{-0.75, 0.75} {
				leg := at.Add(dir.Scale(f)).Add(side.Scale(0.3 * s))
				g.Fill(Capsule{leg.At(y), leg.At(y + 1.2), 0.3})
			}
		}
		neck := at.Add(dir.Scale(1.0)).At(y + 2.5)
		if i%2 == 0 {
			neck = at.Add(dir.Scale(1.6)).At(y + 0.9) // Grazing: head down.
		}
		g.Fill(Capsule{at.Add(dir.Scale(0.8)).At(y + 1.6), neck, 0.4})
		g.Fill(Sphere(neck, 0.5))
		g.Use(White).Fill(Sphere(at.Sub(dir.Scale(1.1)).At(y+1.6), 0.45))
	}
}

// summit piles the great granite boulders that crown Mt. Misen.
func (it Itsukushima) summit(g *Grid) {
	for i := 0; i < 11; i++ {
		a := Hash01(i, 1) * 2 * math.Pi
		r := 5.5 * math.Sqrt(Hash01(i, 2))
		at := Polar(P(-4, -42), r, a)
		y := it.ground(at.X, at.Y)
		if math.IsNaN(y) {
			continue
		}
		s := 1.6 + 1.6*Hash01(i, 3)
		e := Ellipsoid{at.At(y + 2.6 + 0.8*Hash01(i, 4)), V(s, s*(0.8+0.5*Hash01(i, 5)), s*0.85)}
		col := Stone
		if Hash01(i, 6) < 0.35 {
			col = DarkStone
		}
		g.Use(col).FillVolume(e)
		g.Paint(DarkStone, Func{e.Bounds(), func(p Vec3) bool {
			return e.Contains(p) && it.h3(p, i) < 0.18
		}})
	}
}

// subShrines adds the Asazaya on the east waterfront and the Tenjin-sha on the west.
func (it Itsukushima) subShrines(g *Grid) {
	d := itsukushimaDeck
	a := P(23.2, 1.8)
	it.hall(g, a, 4.2, 2.2, d, d+3.8)
	it.roof(g, a, 4.2, 2.2, 1.6, d+3.5, 3.2, 0.5, AxisX, itsukushimaIrimoya, d+3.8, Charcoal)
	it.rail(g, P(18, -2.2), P(28.5, 5.5), d, func(x, z float64) bool { return z < 0 || math.Abs(x-23.2) < 1.2 })
	t := P(-24.2, -11.2)
	it.hall(g, t, 2.6, 2.6, d, d+3.6)
	it.roof(g, t, 2.6, 2.6, 1.3, d+3.3, 2.8, 0.4, AxisZ, itsukushimaIrimoya, d+3.6, Charcoal)
}

// tahoto is the two-storey Tahoto pagoda: a square vermilion and white lower storey
// under a wide roof, the white mortar dome (kamebara), a round upper storey, a square
// upper roof and a bronze finial.
func (it Itsukushima) tahoto(g *Grid, c Vec2) {
	y := it.ground(c.X, c.Y)
	g.Use(Stone).Fill(BlockAt(c, 8.2, 8.2, y-1, y+0.9))
	y += 0.9
	it.storey(g, c, 2.9, y, y+3.4)
	it.roof(g, c, 2.9, 2.9, 2.0, y+3.2, 2.0, 0.7, AxisX, itsukushimaHip, y+3.4, Charcoal)
	y += 4.6
	g.Use(White).FillVolume(Ellipsoid{c.At(y), V(2.4, 1.5, 2.4)})
	g.Use(Red).Fill(Cylinder{Center: c, Radius: 1.9, Y0: y + 0.6, Y1: y + 2.6})
	g.Use(Charcoal).Fill(Cylinder{Center: c, Radius: 1.5, Y0: y + 1.0, Y1: y + 2.2})
	g.Use(Red).Fill(Cylinder{Center: c, Radius: 1.9, Inner: 1.3, Y0: y + 0.6, Y1: y + 2.6})
	for k := 0; k < 4; k++ {
		g.Carve(Arch{c.Add(Polar(P(0, 0), 1.6, float64(k)*math.Pi/2+math.Pi/4)).At(y + 1.0), 0.5, 0.6, 1.2, Axis(k % 2)})
	}
	it.roof(g, c, 2.2, 2.2, 1.9, y+2.5, 2.0, 0.7, AxisX, itsukushimaHip, y+2.6, Charcoal)
	y += 4.5
	g.Use(Charcoal).Fill(Cylinder{Center: c, Radius: 0.45, Y0: y - 0.5, Y1: y + 4.5})
	for k := 0; k < 4; k++ {
		g.Use(Gold).Fill(Cylinder{Center: c, Radius: 0.8, Y0: y + 0.6 + 0.8*float64(k), Y1: y + 1.0 + 0.8*float64(k)})
	}
	g.Use(Gold).Fill(Sphere(c.At(y+4.9), 0.55))
}

// gozabune is a vermilion festival barge: a wooden hull with an upswept bow and a
// roofed pavilion amidships, at plan center c heading along angle a.
func (it Itsukushima) gozabune(g *Grid, c Vec2, a float64) {
	const L, W = 5.5, 1.7
	local := func(p Vec3) Vec2 { return p.Plan().Sub(c).Rotate(-a) }
	box := Box{V(c.X-7, 0, c.Y-7), V(c.X+7, 7, c.Y+7)}
	sheer := func(u float64) float64 {
		return 1.5 + 1.1*math.Pow(math.Max(0, u/L), 3) + 0.3*math.Pow(math.Max(0, -u/L), 2)
	}
	half := func(u, y float64) float64 {
		t := u / L
		w := W * math.Sqrt(math.Max(0, 1-math.Pow(math.Max(0, t), 2))) * math.Sqrt(math.Max(0, 1-math.Pow(math.Min(0, t), 6)))
		return w * (0.55 + 0.45*it.clamp01((y-0.2)/1.2))
	}
	hull := Func{box, func(p Vec3) bool {
		q := local(p)
		return math.Abs(q.X) <= L && p.Y >= 0.2 && p.Y <= sheer(q.X) && math.Abs(q.Y) <= half(q.X, p.Y)
	}}
	g.Use(Wood).FillVolume(hull)
	g.Carve(Func{box, func(p Vec3) bool {
		q := local(p)
		return math.Abs(q.X) <= L-1.2 && p.Y > 1.2 && math.Abs(q.Y) <= half(q.X, p.Y)-0.7
	}})
	g.Paint(Red, Func{box, func(p Vec3) bool {
		q := local(p)
		return math.Abs(q.X) <= L && p.Y > sheer(q.X)-0.5 && hull.F(p)
	}})
	// Pavilion: vermilion posts and frame under a small gabled roof.
	for _, u := range []float64{-1.6, 1.6} {
		for _, v := range []float64{-0.9, 0.9} {
			at := c.Add(P(u, v).Rotate(a))
			g.Use(Red).Fill(Cylinder{Center: at, Radius: 0.45, Y0: 1.0, Y1: 3.6})
		}
	}
	g.Use(Red).Fill(RotateY(Subtract(BlockAt(c, 4.0, 2.6, 3.2, 3.7), BlockAt(c, 2.6, 1.2, 3, 4)), c, a))
	g.Use(Charcoal).Fill(RotateY(Gable{P(c.X-2.6, c.Y-1.9), P(c.X+2.6, c.Y+1.9), 3.6, AxisX}, c, a))
	for _, u := range []float64{-2.4, 2.4} {
		g.Use(Gold).Fill(Sphere(c.Add(P(u, 0).Rotate(a)).At(5.4), 0.45))
	}
}

// cedars raises tall conical sugi above the woodland canopy for a broken skyline.
func (it Itsukushima) cedars(g *Grid) {
	const cs = 6.5
	for i := -6; i <= 6; i++ {
		for k := -9; k <= -3; k++ {
			x := (float64(i) + 0.2 + 0.6*Hash01(i, k, 61)) * cs
			z := (float64(k) + 0.2 + 0.6*Hash01(i, k, 62)) * cs
			if it.forest(x, z) < 0.8 || Hash01(i, k, 63) > 0.4 || math.Abs(x) > itsukushimaHX-5 || z < itsukushimaZ0+5 {
				continue
			}
			if math.Hypot(x+4, z+42) < 8 {
				continue // Leave the summit boulders clear.
			}
			s := it.surface(x, z)
			if math.IsNaN(s) {
				continue
			}
			rimF := it.rim(x, z, it.raw(x, z))
			h := (3 + 2.5*Hash01(i, k, 64)) * (0.4 + 0.6*rimF)
			r := 1.5 + 0.5*Hash01(i, k, 65)
			at := P(x, z)
			g.Use(Foliage).FillVolume(Frustum{at, r, 0.4, s - 2, s + h})
			g.Paint(Grass, Func{Box{V(x-r, s-2, z-r), V(x+r, s+h, z+r)}, func(p Vec3) bool {
				return it.h3(p, i*31+k) < 0.12
			}})
		}
	}
}

// archBridge is a small arched vermilion footbridge from a to b (plan) with a given rise.
func (it Itsukushima) archBridge(g *Grid, a, b Vec2, rise float64) {
	ya, yb := it.ground(a.X, a.Y), it.ground(b.X, b.Y)
	box := it.segBox(a, b, 2.5, 0, math.Max(ya, yb)+rise+3)
	_, _, L := it.seg(a, b, a.X, a.Y)
	deck := func(s float64) float64 {
		t := it.clamp01(s / L)
		return ya + (yb-ya)*t + rise*4*t*(1-t)
	}
	g.Use(Wood).Fill(Func{box, func(p Vec3) bool {
		s, t, _ := it.seg(a, b, p.X, p.Z)
		y := deck(s)
		return s >= 0 && s <= L && math.Abs(t) <= 1.3 && p.Y <= y && p.Y >= y-0.7
	}})
	g.Use(Red).Fill(Func{box, func(p Vec3) bool {
		s, t, _ := it.seg(a, b, p.X, p.Z)
		y := deck(s)
		edge := math.Abs(math.Abs(t)-1.3) <= 0.45
		return s >= 0 && s <= L && edge && ((p.Y > y+0.8 && p.Y <= y+1.3) || (p.Y > y-1.2 && p.Y <= y) || (p.Y > y && math.Mod(s+100, 1.6) < 0.6 && p.Y <= y+0.8))
	}})
}
