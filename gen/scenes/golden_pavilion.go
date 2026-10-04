package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// GoldenPavilion is Kinkaku-ji in Kyoto on the Kyoko-chi mirror pond: a three-storey
// pavilion with a natural wood and white plaster ground floor and an open veranda,
// two gold-leafed upper floors with balconies, flaring shingled eaves and a bronze
// phoenix on the pyramidal roof, plus the small fishing pavilion over the water. The
// pond holds pine islands (including the crane and turtle islands) and rocks; maples,
// moss, gravel paths, a stone lantern and a tea house lead up into wooded hills.
type GoldenPavilion struct{}

func init() { Register(GoldenPavilion{}) }

func (GoldenPavilion) Name() string  { return "golden_pavilion" }
func (GoldenPavilion) Title() string { return "Golden Pavilion" }

const (
	goldenPavilionWater = 0.5   // Top of the pond surface.
	goldenPavilionBank  = 1.6   // Lawn level beside the pond.
	goldenPavilionHX    = 37.0  // Site half-width (X).
	goldenPavilionZ0    = -44.0 // Back (north) edge of the site.
	goldenPavilionZ1    = 30.0  // Front (south) edge of the site.
)

// goldenPavilionCenter is the pavilion's plan center; it faces south (+Z) over the pond.
var goldenPavilionCenter = P(0, -7)

// goldenPavilionTea is the tea house site on the north-east hillside.
var goldenPavilionTea = P(23, -22)

// goldenPavilionPaths are the gravel walks (polylines in plan).
var goldenPavilionPaths = [][]Vec2{
	{P(-38, 26.5), P(-20, 27.2), P(0, 26.4), P(18, 27.2), P(30, 25)},
	{P(30, 25), P(29, 14), P(27, 4), P(22, -2), P(18, -7), P(17, -14), P(19, -19), goldenPavilionTea},
}

func (gp GoldenPavilion) Build(g *Grid) Order {
	gp.terrain(g)
	gp.reflection(g)
	gp.islands(g)
	gp.shoreRocks(g)
	gp.pavilion(g)
	gp.fishing(g)
	gp.lantern(g, P(18.5, -3))
	gp.teaHouse(g)
	gp.bridge(g)
	gp.fence(g)
	gp.trees(g)
	return Outward(goldenPavilionCenter)
}

// sm is a smoothstep from e0 to e1 (either direction).
func (GoldenPavilion) sm(e0, e1, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-e0)/(e1-e0)))
	return t * t * (3 - 2*t)
}

func (GoldenPavilion) clamp01(v float64) float64 { return math.Max(0, math.Min(1, v)) }

// pond is an approximate signed distance to the pond shore (negative on the water).
func (GoldenPavilion) pond(x, z float64) float64 {
	dx, dz := x/28, (z-11.5)/14
	if z > 11.5 {
		dz = (z - 11.5) / 12.5
	}
	a := math.Atan2(dz, dx)
	r := 1 + 0.07*math.Sin(3*a+0.5) + 0.05*math.Sin(5*a+1.3)
	e := (math.Hypot(dx, dz)/r - 1) * 14
	lx, lz := (x+19)/7.5, (z+2.5)/6
	e2 := (math.Hypot(lx, lz) - 1) * 7
	return math.Min(e, e2)
}

// goldenPavilionStream is the brook that tumbles off the east hill into the pond.
var goldenPavilionStream = [][]Vec2{{P(31, -21), P(29.5, -12), P(25, -4), P(20.5, 3)}}

// pathDist is the plan distance to the nearest gravel walk.
func (gp GoldenPavilion) pathDist(x, z float64) float64 {
	return gp.lineDist(goldenPavilionPaths, x, z)
}

// streamDist is the plan distance to the brook's center line.
func (gp GoldenPavilion) streamDist(x, z float64) float64 {
	return gp.lineDist(goldenPavilionStream, x, z)
}

// lineDist is the plan distance to the nearest of several polylines.
func (GoldenPavilion) lineDist(lines [][]Vec2, x, z float64) float64 {
	best := math.Inf(1)
	p := P(x, z)
	for _, line := range lines {
		for i := 1; i < len(line); i++ {
			a, b := line[i-1], line[i]
			ab := b.Sub(a)
			t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
			best = math.Min(best, p.Sub(a.Add(ab.Scale(t))).Length())
		}
	}
	return best
}

// hills is the raw hill height above the lawn: a ridge behind the pavilion and lower
// wooded slopes along the east and west sides.
func (gp GoldenPavilion) hills(x, z float64) float64 {
	north := 15 * gp.sm(-18, -34, z)
	side := 6 * gp.sm(23, 31, math.Abs(x)) * gp.sm(18, 4, z)
	h := math.Max(north, side)
	return h + 0.9*math.Sin(x*0.23+1)*math.Sin(z*0.19+2)*gp.clamp01(h/4)
}

// rim fades the terrain down to a low edge around the site.
func (gp GoldenPavilion) rim(x, z, h float64) float64 {
	e := math.Min(goldenPavilionHX-math.Abs(x), math.Min(goldenPavilionZ1-z, z-goldenPavilionZ0))
	return gp.sm(0, 3+0.5*h, e)
}

// ground is the terrain height, or NaN on the pond.
func (gp GoldenPavilion) ground(x, z float64) float64 {
	d := gp.pond(x, z)
	if d < 0 {
		return math.NaN()
	}
	h := gp.hills(x, z)
	// A level terrace for the tea house.
	t := goldenPavilionTea
	ht := gp.hills(t.X, t.Y)
	h = ht + (h-ht)*gp.sm(5, 8, math.Hypot(x-t.X, z-t.Y))
	y := 0.4 + (goldenPavilionBank+h-0.4)*gp.rim(x, z, h)
	y -= 0.9 * gp.sm(1.9, 0.7, gp.streamDist(x, z))
	return goldenPavilionWater + (y-goldenPavilionWater)*gp.sm(0, 2.5, d)
}

// forest is how strongly the woodland covers a plan point.
func (gp GoldenPavilion) forest(x, z float64) float64 {
	f := gp.sm(1.5, 3.5, gp.hills(x, z))
	f *= gp.sm(6, 8.5, math.Hypot(x-goldenPavilionTea.X, z-goldenPavilionTea.Y))
	f *= gp.sm(1.6, 3.2, gp.pathDist(x, z))
	f *= gp.sm(1.6, 3.6, gp.streamDist(x, z))
	return f
}

// canopy is the woodland crown height over a plan point and the crown's species.
func (gp GoldenPavilion) canopy(x, z float64) (float64, int) {
	const cs = 3.6
	ci, ck := int(math.Floor(x/cs)), int(math.Floor(z/cs))
	best, kind := math.Inf(-1), -1
	here := gp.ground(x, z)
	if math.IsNaN(here) {
		return best, kind
	}
	for di := -1; di <= 1; di++ {
		for dk := -1; dk <= 1; dk++ {
			i, k := ci+di, ck+dk
			tx := (float64(i) + 0.15 + 0.7*Hash01(i, k, 1)) * cs
			tz := (float64(k) + 0.15 + 0.7*Hash01(i, k, 2)) * cs
			f := gp.forest(tx, tz)
			if f < 0.25 {
				continue
			}
			r := (2.6 + 1.4*Hash01(i, k, 3)) * (0.6 + 0.4*f)
			d := math.Hypot(x-tx, z-tz)
			if d >= r {
				continue
			}
			rimF := gp.rim(tx, tz, gp.hills(tx, tz))
			top := here + ((0.8+1.8*Hash01(i, k, 4))+1.15*r*math.Pow(1-d*d/(r*r), 0.5))*f*(0.4+0.6*rimF)
			if top > best {
				best, kind = top, int(Hash01(i, k, 5)*100)
			}
		}
	}
	return best, kind
}

// surface is the top of the land: ground or tree crowns (NaN on the pond).
func (gp GoldenPavilion) surface(x, z float64) float64 {
	gy := gp.ground(x, z)
	if math.IsNaN(gy) {
		return gy
	}
	c, _ := gp.canopy(x, z)
	return math.Max(gy, c)
}

// terrain lays the pond water, the lawn shell and the woodland canopy as one
// heightfield shell, then colors it: moss, gravel walks and autumn crowns.
func (gp GoldenPavilion) terrain(g *Grid) {
	box := Box{V(-goldenPavilionHX, 0, goldenPavilionZ0), V(goldenPavilionHX, 32, goldenPavilionZ1)}
	g.Use(Water).Fill(Func{Box{V(-goldenPavilionHX, 0, goldenPavilionZ0), V(goldenPavilionHX, goldenPavilionWater, goldenPavilionZ1)}, func(p Vec3) bool {
		return gp.pond(p.X, p.Z) < 1.2
	}})
	land := NewGridWithSpan(g.Span())
	land.Use(Grass).Fill(Func{box, func(p Vec3) bool {
		s := gp.surface(p.X, p.Z)
		if math.IsNaN(s) || p.Y > s {
			return false
		}
		lo := s
		for _, o := range []Vec2{P(0.9, 0), P(-0.9, 0), P(0, 0.9), P(0, -0.9)} {
			n := gp.surface(p.X+o.X, p.Z+o.Y)
			if math.IsNaN(n) {
				n = 0
			}
			lo = math.Min(lo, n)
		}
		return p.Y >= math.Max(0, lo-1.1)
	}})
	for _, c := range land.Cells() {
		p := land.World(c)
		gy := gp.ground(p.X, p.Z)
		top, kind := gp.canopy(p.X, p.Z)
		col := Grass
		switch {
		case top > gy+0.4:
			switch {
			case kind < 70:
				col = Foliage
			case kind < 92:
				col = Grass
			case kind < 97:
				col = Red
			default:
				col = Terracotta
			}
			if col == Foliage && Hash01(int(c.X), int(c.Y), int(c.Z)) < 0.12 {
				col = Grass
			}
		case gp.streamDist(p.X, p.Z) < 1.0 && p.Y > gy-0.9:
			col = Water
		case gp.streamDist(p.X, p.Z) < 1.8 && p.Y > gy-0.9:
			col = Stone
			if Hash01(int(c.X), int(c.Z), 5) < 0.4 {
				col = DarkStone
			}
		case gp.pathDist(p.X, p.Z) < 1.3 && p.Y > gy-1:
			col = Sand
		case Hash01(int(c.X), int(c.Y), int(c.Z), 3) < 0.14:
			col = Foliage // Moss variation.
		}
		g.Use(col).Add(c)
	}
	// A gravel court east of the pavilion.
	g.Paint(Sand, Func{Box{V(14, 0, -20), V(24, 4, -6)}, func(p Vec3) bool {
		gy := gp.ground(p.X, p.Z)
		return !math.IsNaN(gy) && p.Y > gy-0.9 && gp.pond(p.X, p.Z) > 1.5
	}})
}

// reflection paints the pavilion's rippled mirror image onto the pond surface: Kyoko-chi
// means "mirror pond". The image runs away from the podium's front edge.
func (gp GoldenPavilion) reflection(g *Grid) {
	c := goldenPavilionCenter
	edge := c.Y + 10.7
	type band struct {
		y0, y1, w0, w1 float64
		col            Color
	}
	bands := []band{
		{1.4, 8.0, 11.5, 11.5, Wood},
		{8.0, 9.8, 13.7, 12.5, Navy},
		{9.8, 15.4, 11.6, 11.6, Gold},
		{15.4, 18.8, 13.6, 7.0, Navy},
		{18.8, 23.6, 7.4, 7.4, Gold},
		{23.6, 29.6, 10.0, 1.2, Navy},
		{29.6, 34.0, 1.6, 1.6, Gold},
	}
	box := Box{V(c.X-15, 0.2, edge), V(c.X+15, goldenPavilionWater, edge+30)}
	for _, b := range bands {
		b := b
		g.Paint(b.col, Func{box, func(p Vec3) bool {
			col, ok := g.ColorOf(Nearest(p.Scale(1 / g.Span())))
			if !ok || col != Water {
				return false
			}
			y := (p.Z-edge)*1.2 + 1.4
			if y < b.y0 || y > b.y1 {
				return false
			}
			w := b.w0 + (b.w1-b.w0)*(y-b.y0)/(b.y1-b.y0)
			x := p.X - c.X + 0.7*math.Sin(p.Z*1.9)
			if math.Abs(x) > w {
				return false
			}
			row := int(math.Floor(p.Z * 1.26))
			return row%3 != 0 && Hash01(int(math.Round(p.X*2)), row, 71) > 0.3
		}})
	}
}

// rock places a weathered boulder.
func (GoldenPavilion) rock(g *Grid, at Vec3, rx, ry, rz float64, seed int) {
	e := Ellipsoid{at, V(rx, ry, rz)}
	s := Func{e.Bounds(), func(p Vec3) bool { return e.Contains(p) && p.Y >= 0 }}
	col := Stone
	if Hash01(seed, 9) < 0.45 {
		col = DarkStone
	}
	g.Use(col).FillVolume(s)
	other := DarkStone
	if col == DarkStone {
		other = Stone
	}
	g.Paint(other, Func{e.Bounds(), func(p Vec3) bool {
		return e.Contains(p) && Hash01(int(math.Round(p.X*2)), int(math.Round(p.Y*2)), int(math.Round(p.Z*2)), seed) < 0.2
	}})
}

// island is a low mossy dome with a stony waterline.
func (gp GoldenPavilion) island(g *Grid, c Vec2, a, b, h float64, seed int) {
	e := Ellipsoid{c.At(-0.4), V(a, h+0.4, b)}
	g.Use(Grass).FillVolume(Func{e.Bounds(), func(p Vec3) bool { return e.Contains(p) && p.Y >= 0 }})
	g.Paint(Stone, Func{e.Bounds(), func(p Vec3) bool { return e.Contains(p) && p.Y < 1.0 }})
	g.Paint(DarkStone, Func{e.Bounds(), func(p Vec3) bool {
		return e.Contains(p) && p.Y < 1.0 && Hash01(int(math.Round(p.X*1.5)), int(math.Round(p.Z*1.5)), seed) < 0.35
	}})
	g.Paint(Foliage, Func{e.Bounds(), func(p Vec3) bool {
		return e.Contains(p) && p.Y >= 1.0 && Hash01(int(math.Round(p.X*2)), int(math.Round(p.Z*2)), seed, 1) < 0.2
	}})
	n := 3 + int(Hash01(seed, 2)*3)
	for k := 0; k < n; k++ {
		ang := Hash01(seed, k, 3) * 2 * math.Pi
		at := P(c.X+a*0.92*math.Cos(ang), c.Y+b*0.92*math.Sin(ang))
		r := 0.9 + 0.7*Hash01(seed, k, 4)
		gp.rock(g, at.At(0.3), r, 0.7+0.6*Hash01(seed, k, 5), r*0.85, seed*10+k)
	}
}

// islands builds Ashihara island, the turtle and crane islands and standing rocks.
func (gp GoldenPavilion) islands(g *Grid) {
	gp.island(g, P(-15, 13), 8.5, 5, 2.0, 1)
	gp.pine(g, P(-18.5, 13), 1.7, 6.8, 1)
	gp.pine(g, P(-12.5, 12.2), 1.7, 5.4, 2)
	gp.pine(g, P(-15, 15.8), 1.5, 4.4, 3)
	gp.pine(g, P(-9.5, 14.5), 1.2, 3.6, 7)

	// Turtle island (Kame-jima): a domed shell with head and flipper stones.
	tc := P(8, 11)
	gp.island(g, tc, 3.8, 3.2, 1.6, 2)
	gp.rock(g, V(tc.X-0.5, 0.6, tc.Y+4.3), 1.0, 0.9, 1.3, 21)
	for _, o := range []Vec2{P(-3.2, 2.4), P(3.2, 2.4), P(-3.0, -2.2), P(3.0, -2.2)} {
		gp.rock(g, tc.Add(o).At(0.2), 0.9, 0.6, 0.9, 22+int(o.X))
	}
	gp.pine(g, P(8.3, 10.6), 1.4, 5.4, 4)

	// Crane island (Tsuru-jima): long and low, with a pine stretched out like a neck.
	cc := P(21, 17)
	gp.island(g, cc, 5, 2.4, 1.3, 3)
	gp.rock(g, V(cc.X+4.6, 1.0, cc.Y-0.4), 1.0, 2.0, 1.0, 31)
	gp.pine(g, P(19.5, 17.2), 1.2, 5.8, 5)

	// Standing stones in the water.
	for i, s := range []Vec3{V(-3, 1.1, 19), V(-1.5, 0.8, 20.5), V(-26, 1.0, 20), V(14, 1.0, 22), V(-8, 0.9, 6), V(27, 0.8, 8), V(-30, 0.9, 9), V(2, 0.8, 25)} {
		gp.rock(g, V(s.X, 0.3, s.Z), 1.0+0.4*Hash01(i, 1), s.Y+0.4, 1.0+0.3*Hash01(i, 2), 40+i)
	}
	// A tiny pine islet with two stones.
	gp.island(g, P(-7, 21), 2.6, 2.0, 1.2, 4)
	gp.pine(g, P(-6.7, 21), 1.1, 3.8, 6)
}

// shoreRocks lines the pond edge with boulders.
func (gp GoldenPavilion) shoreRocks(g *Grid) {
	for k := 0; k < 64; k++ {
		a := float64(k) / 64 * 2 * math.Pi
		r := 1 + 0.07*math.Sin(3*a+0.5) + 0.05*math.Sin(5*a+1.3)
		x, z := 28*r*math.Cos(a), 11.5+14*r*math.Sin(a)
		if math.Abs(x) < 15 && z < 6 || Hash01(k, 77) < 0.35 {
			continue
		}
		s := 0.8 + 0.8*Hash01(k, 78)
		gp.rock(g, V(x, 0.6, z), s*1.2, s*0.8, s, 100+k)
	}
}

// eaveSurface is a flared hip-roof skirt height over a plan point: from yRidge at
// `inner` inside the rectangle down a concave curve to yEave at `over` outside, with
// corners swept up by `up`.
func (gp GoldenPavilion) eaveSurface(c Vec2, hx, hz, inner, over, yEave, yRidge, up, x, z float64) (s, u, v, d float64, ok bool) {
	u, v = math.Abs(x-c.X)-hx, math.Abs(z-c.Y)-hz
	d = math.Max(u, v)
	if d < -inner || d > over {
		return 0, u, v, d, false
	}
	t := (d + inner) / (inner + over)
	s = yEave + (yRidge-yEave)*math.Pow(1-t, 1.7)
	ext := hx + over
	if u > v {
		ext = hz + over
	}
	near := gp.clamp01(1 - (d-math.Min(u, v))/ext)
	s += up * math.Pow(near, 3) * (0.25 + 0.75*t)
	return s, u, v, d, true
}

// roof fills a shingled skirt: charcoal cypress bark, dark hip ridges, a thick brown
// eave edge and wooden rafters underneath.
func (gp GoldenPavilion) roof(g *Grid, c Vec2, hx, hz, inner, over, yEave, yRidge, up, thick float64) {
	r := over + 0.6
	box := Box{V(c.X-hx-r, yEave-thick-1, c.Y-hz-r), V(c.X+hx+r, yRidge+up+1, c.Y+hz+r)}
	at := func(p Vec3) (float64, float64, float64, float64, bool) {
		return gp.eaveSurface(c, hx, hz, inner, over, yEave, yRidge, up, p.X, p.Z)
	}
	g.Use(Charcoal).FillVolume(Func{box, func(p Vec3) bool {
		s, _, _, _, ok := at(p)
		return ok && p.Y <= s && p.Y >= s-thick
	}})
	g.Paint(Wood, Func{box, func(p Vec3) bool {
		s, _, _, d, ok := at(p)
		return ok && d > 0.4 && p.Y < s-thick+0.45
	}})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		s, u, v, d, ok := at(p)
		return ok && p.Y >= s-0.6 && math.Abs(u-v) < 0.5 && d > -inner+0.5
	}})
	g.Use(Wood).Fill(Func{box, func(p Vec3) bool {
		s, _, _, d, ok := at(p)
		return ok && d > over-0.6 && p.Y >= s-thick-0.1 && p.Y <= s+0.1
	}})
}

// storey is a hollow storey: an outer wall shell with a dark lining inside it and a
// dark floor and ceiling, so openings carved later show a shadowed interior.
func (GoldenPavilion) storey(g *Grid, c Vec2, hx, hz, y0, y1 float64, wall Color) {
	g.Use(Charcoal).Fill(Subtract(BlockAt(c, 2*hx-1.6, 2*hz-1.6, y0, y1), BlockAt(c, 2*hx-3.4, 2*hz-3.4, y0+0.6, y1-0.6)))
	g.Use(wall).Fill(Subtract(BlockAt(c, 2*hx, 2*hz, y0, y1), BlockAt(c, 2*hx-1.8, 2*hz-1.8, y0-1, y1+1)))
}

// openings carves a row of openings into each face of a rectangular storey. xs lists the
// opening centers along X faces (offsets from c.X), zs along Z faces; shape returns
// whether a point (across offset a, height y) lies in the opening.
func (GoldenPavilion) openings(g *Grid, c Vec2, hx, hz float64, faces []int, xs, zs []float64, y0, y1 float64, shape func(a, y float64) bool) {
	for _, f := range faces {
		var box Box
		var along []float64
		switch f {
		case 0, 2: // +Z, -Z faces: openings spread along X.
			z := c.Y + hz
			if f == 2 {
				z = c.Y - hz
			}
			box, along = Box{V(c.X-hx, y0, z-2.0), V(c.X+hx, y1, z+0.6)}, xs
		default: // +X, -X faces: openings spread along Z.
			x := c.X + hx
			if f == 3 {
				x = c.X - hx
			}
			box, along = Box{V(x-2.0, y0, c.Y-hz), V(x+0.6, y1, c.Y+hz)}, zs
		}
		for _, o := range along {
			o := o
			g.Carve(Func{box, func(p Vec3) bool {
				a := p.X - c.X - o
				if f == 1 || f == 3 {
					a = p.Z - c.Y - o
				}
				return shape(a, p.Y)
			}})
		}
	}
}

// rectShape is a rectangular opening of half-width w between y0 and y1.
func (GoldenPavilion) rectShape(w, y0, y1 float64) func(a, y float64) bool {
	return func(a, y float64) bool { return math.Abs(a) <= w && y >= y0 && y <= y1 }
}

// bellShape is a katomado (cusped bell) window: straight jambs rising to a pointed,
// flared head.
func (GoldenPavilion) bellShape(w, y0, spring, y1 float64) func(a, y float64) bool {
	return func(a, y float64) bool {
		if y < y0 || y > y1 {
			return false
		}
		if y <= spring {
			return math.Abs(a) <= w
		}
		t := (y - spring) / (y1 - spring)
		hw := w * (1 + 0.25*math.Sin(math.Pi*t*0.7)) * math.Cos(t*math.Pi/2)
		return math.Abs(a) <= math.Max(hw, 0.45)
	}
}

// ring is a rectangular band between two half-size rectangles.
func (GoldenPavilion) ring(c Vec2, hxo, hzo, hxi, hzi, y0, y1 float64) Solid {
	return Subtract(BlockAt(c, 2*hxo, 2*hzo, y0, y1), BlockAt(c, 2*hxi, 2*hzi, y0-1, y1+1))
}

// balcony lays a deck ring around a storey with a railing at its outer edge.
func (gp GoldenPavilion) balcony(g *Grid, c Vec2, hx, hz, out, y float64, rail Color) {
	g.Use(Wood).Fill(gp.ring(c, hx+out, hz+out, hx-0.5, hz-0.5, y-0.6, y))
	ox, oz := hx+out, hz+out
	g.Use(rail).Fill(gp.ring(c, ox, oz, ox-0.9, oz-0.9, y+1.0, y+1.5))
	for _, side := range []float64{-1, 1} {
		for x := -ox; x <= ox+0.01; x += 2.2 {
			g.Fill(BlockAt(P(c.X+x, c.Y+side*(oz-0.45)), 0.9, 0.9, y, y+1.5))
		}
		for z := -oz; z <= oz+0.01; z += 2.2 {
			g.Fill(BlockAt(P(c.X+side*(ox-0.45), c.Y+z), 0.9, 0.9, y, y+1.5))
		}
	}
}

// pavilion builds the three storeys, the roofs and the phoenix.
// pilasters adds slightly proud wall posts at the given offsets along each face.
func (GoldenPavilion) pilasters(g *Grid, c Vec2, hx, hz float64, xs, zs []float64, y0, y1 float64) {
	for _, side := range []float64{-1, 1} {
		for _, x := range xs {
			g.Fill(BlockAt(P(c.X+x, c.Y+side*hz), 1.0, 1.1, y0, y1))
		}
		for _, z := range zs {
			g.Fill(BlockAt(P(c.X+side*hx, c.Y+z), 1.1, 1.0, y0, y1))
		}
	}
}

// pavilion builds the three storeys, the roofs and the phoenix.
func (gp GoldenPavilion) pavilion(g *Grid) {
	c := goldenPavilionCenter

	// Stone podium (kidan) standing in the pond, hollow below its rim.
	podium := BlockAt(c, 27, 21.4, 0, 1.4)
	g.Use(Stone).Fill(Subtract(podium, BlockAt(c, 24, 18.4, -1, 0.8)))
	g.Paint(DarkStone, Func{podium.Bounds(), func(p Vec3) bool {
		course := int(math.Floor(p.Y / 0.7))
		return Hash01(int(math.Floor((p.X+float64(course%2))/1.6)), course, int(math.Floor((p.Z+float64(course%2))/1.6))) < 0.3
	}})

	// Ground floor (Hosui-in): veranda deck, white plaster walls framed in natural wood.
	const f0, f1 = 1.9, 8.8
	const phx, phz = 12.0, 9.2 // Veranda post line.
	g.Use(Wood).Fill(BlockAt(c, 2*phx+1.2, 2*phz+1.2, 1.3, f0))
	wallC, whx, whz := P(c.X, c.Y-0.8), 11.6, 8.0
	gp.storey(g, wallC, whx, whz, f0, f1, White)
	g.Use(Wood).Fill(gp.ring(wallC, whx+0.2, whz+0.2, whx-0.9, whz-0.9, f0, f0+0.8))
	g.Fill(gp.ring(wallC, whx+0.3, whz+0.3, whx-0.9, whz-0.9, f1-0.9, f1))

	gp.pilasters(g, wallC, whx, whz, []float64{-whx + 0.3, -7.2, -2.4, 2.4, 7.2, whx - 0.3}, []float64{-whz + 0.3, whz - 0.3}, f0, f1)
	// Open sliding doors on the south front, propped-up shutters elsewhere.
	gp.openings(g, wallC, whx, whz, []int{0}, []float64{-4.8, 0, 4.8}, nil, f0, f1, gp.rectShape(1.7, f0+0.3, f1-3.2))
	gp.openings(g, wallC, whx, whz, []int{0}, []float64{-9.4, 9.4}, nil, f0, f1, gp.rectShape(0.9, f0+3.4, f1-1.6))
	gp.openings(g, wallC, whx, whz, []int{2}, []float64{-9.6, -4.8, 0, 4.8, 9.6}, nil, f0, f1, gp.rectShape(1.3, f0+3.4, f1-1.6))
	gp.openings(g, wallC, whx, whz, []int{1, 3}, nil, []float64{-6.1, -1.5, 3.1}, f0, f1, gp.rectShape(1.3, f0+3.4, f1-1.6))
	gp.openings(g, wallC, whx, whz, []int{1, 3}, nil, []float64{6.7}, f0, f1, gp.rectShape(0.6, f0+3.4, f1-1.6))
	// Veranda posts and the ring beam carrying the first eave.
	g.Use(Wood)
	for _, x := range []float64{-phx, -7.2, -2.4, 2.4, 7.2, phx} {
		for _, side := range []float64{-1, 1} {
			g.Fill(BlockAt(P(c.X+x, c.Y+side*phz), 1.0, 1.0, f0, f1))
		}
	}
	for _, z := range []float64{-4.6, 0, 4.6} {
		for _, side := range []float64{-1, 1} {
			g.Fill(BlockAt(P(c.X+side*phx, c.Y+z), 1.0, 1.0, f0, f1))
		}
	}
	g.Fill(gp.ring(c, phx+0.5, phz+0.5, phx-0.5, phz-0.5, f1-0.8, f1))
	g.Fill(gp.ring(c, phx, phz, whx, whz, f1-0.4, f1)) // Veranda ceiling.

	// First (pent) eave, then the second storey (Choon-do) in gold with a balcony.
	gp.roof(g, c, phx, phz, 1.4, 1.7, f1-0.1, f1+1.0, 0.5, 0.9)
	const s0, s1 = 10.0, 16.4
	shx, shz := 10.6, 7.8
	g.Use(Wood).Fill(BlockAt(c, 2*shx+2.4, 2*shz+2.4, s0-0.6, s0))
	gp.storey(g, c, shx, shz, s0, s1, Gold)
	gp.balcony(g, c, shx, shz, 1.2, s0, Gold)
	g.Use(Gold).Fill(gp.ring(c, shx+0.3, shz+0.3, shx-0.9, shz-0.9, s1-0.9, s1))
	gp.pilasters(g, c, shx, shz, []float64{-shx + 0.3, -7.2, -2.4, 2.4, 7.2, shx - 0.3}, []float64{-shz + 0.3, -2.6, 2.6, shz - 0.3}, s0, s1)
	gp.openings(g, c, shx, shz, []int{0}, []float64{0}, nil, s0, s1, gp.rectShape(1.5, s0+0.1, s1-1.3))
	gp.openings(g, c, shx, shz, []int{0, 2}, []float64{-8.9, -4.8, 4.8, 8.9}, nil, s0, s1, gp.rectShape(1.0, s0+1.8, s1-1.6))
	gp.openings(g, c, shx, shz, []int{2}, []float64{0}, nil, s0, s1, gp.rectShape(1.0, s0+1.8, s1-1.6))
	gp.openings(g, c, shx, shz, []int{1, 3}, nil, []float64{-5.2, 0, 5.2}, s0, s1, gp.rectShape(1.0, s0+1.8, s1-1.6))

	// Second eave rising to the third storey (Kukkyo-cho): square, gold, bell windows.
	thx := 6.2
	gp.roof(g, c, shx, shz, 4.8, 3.0, s1-0.3, s1+2.6, 1.1, 1.0)
	const t0, t1 = 19.2, 24.2
	gp.storey(g, c, thx, thx, s1-0.5, t1, Gold)
	gp.balcony(g, c, thx, thx, 1.3, t0, Gold)
	g.Use(Gold).Fill(gp.ring(c, thx+0.3, thx+0.3, thx-0.9, thx-0.9, t1-0.8, t1))
	posts := []float64{-thx + 0.3, -2.1, 2.1, thx - 0.3}
	gp.pilasters(g, c, thx, thx, posts, posts, t0, t1)
	all := []int{0, 1, 2, 3}
	gp.openings(g, c, thx, thx, []int{0, 2}, []float64{0}, nil, t0, t1, gp.rectShape(1.1, t0+0.1, t1-1.1))
	gp.openings(g, c, thx, thx, []int{1, 3}, nil, []float64{0}, t0, t1, gp.rectShape(1.1, t0+0.1, t1-1.1))
	gp.openings(g, c, thx, thx, all, []float64{-4.15, 4.15}, []float64{-4.15, 4.15}, t0, t1, gp.bellShape(0.95, t0+0.8, t0+2.3, t1-0.9))

	// Pyramidal (hogyo) roof and the gold phoenix.
	gp.roof(g, c, thx, thx, thx, 3.8, t1-0.3, t1+5.6, 1.6, 1.1)
	top := t1 + 5.6
	g.Use(Gold).Fill(Cylinder{Center: c, Radius: 1.0, Y0: top - 1.0, Y1: top + 0.6})
	g.Fill(Sphere(c.At(top+0.6), 0.8))
	gp.phoenix(g, c.At(top+1.0))
}

// phoenix is the gilded bronze hoo bird facing south, wings raised and tail high.
func (GoldenPavilion) phoenix(g *Grid, base Vec3) {
	g.Use(Gold)
	o := func(x, y, z float64) Vec3 { return base.Add(V(x, y, z)) }
	g.Fill(Capsule{o(0, 0, 0), o(0, 0.9, 0), 0.55})
	g.Fill(Ellipsoid{o(0, 1.6, 0), V(0.85, 1.1, 1.3)})
	g.Fill(Capsule{o(0, 2.2, 0.7), o(0, 3.4, 1.2), 0.55})
	g.Fill(Sphere(o(0, 3.6, 1.4), 0.7))
	g.Fill(Capsule{o(0, 3.6, 1.6), o(0, 3.4, 2.3), 0.45}, Capsule{o(0, 3.8, 1.2), o(0, 4.5, 0.7), 0.45})
	g.Fill(Capsule{o(0, 1.3, -1.0), o(0, 1.9, -2.6), 0.6}, Capsule{o(0, 1.9, -2.6), o(0, 3.1, -3.6), 0.5})
	for _, s := range []float64{-1, 1} {
		g.Fill(Capsule{o(s*0.6, 2.2, 0), o(s*2.2, 2.8, -0.6), 0.6}, Capsule{o(s*2.2, 2.8, -0.6), o(s*3.0, 3.8, -1.4), 0.5})
		g.Fill(Capsule{o(s*0.6, 1.8, -0.4), o(s*2.4, 2.2, -1.4), 0.5})
		g.Fill(Capsule{o(0, 1.9, -2.4), o(s*0.9, 2.5, -3.3), 0.45})
	}
}

// fishing builds the Sosei fishing pavilion jutting west over the water.
func (gp GoldenPavilion) fishing(g *Grid) {
	c := P(-16.0, -2.5)
	const hx, hz, d0 = 3.4, 3.4, 2.0
	g.Use(Wood)
	for _, x := range []float64{-hx + 0.5, hx - 0.5} {
		for _, z := range []float64{-hz + 0.5, hz - 0.5} {
			g.Fill(BlockAt(P(c.X+x, c.Y+z), 1.0, 1.0, 0, 6.2))
		}
	}
	g.Fill(BlockAt(c, 2*hx+0.6, 2*hz, d0-0.6, d0))
	g.Fill(BlockAt(P(c.X+hx+1.0, c.Y), 2.4, 2*hz, d0-0.6, d0)) // Link to the veranda.
	// Low railing and a white back screen on the north.
	g.Fill(gp.ring(c, hx+0.3, hz, hx-0.6, hz-0.9, d0+0.8, d0+1.3))
	g.Carve(BlockAt(P(c.X+hx, c.Y), 2, 3, d0+0.5, d0+1.5))
	g.Use(White).Fill(Block(V(c.X-hx+0.5, d0, c.Y-hz+0.1), V(c.X+hx-0.5, 5.4, c.Y-hz+1.0)))
	g.Use(Wood).Fill(gp.ring(c, hx, hz, hx-0.9, hz-0.9, 5.4, 6.2))
	gp.roof(g, c, hx, hz, 2.6, 1.8, 5.8, 8.3, 0.6, 1.0)
}

// bridge lays a granite slab bridge where the east walk crosses the brook.
func (gp GoldenPavilion) bridge(g *Grid) {
	at := P(22.74, -1.11)
	dir := P(-5, -6).Scale(1 / math.Hypot(5, 6))
	y := gp.ground(at.X+dir.X*3.5, at.Y+dir.Y*3.5)
	g.Use(Stone).Fill(Func{Box{V(at.X-4, y-1, at.Y-4), V(at.X+4, y+1, at.Y+4)}, func(p Vec3) bool {
		q := P(p.X, p.Z).Sub(at)
		along, across := q.Dot(dir), q.Dot(P(-dir.Y, dir.X))
		return math.Abs(along) <= 2.8 && math.Abs(across) <= 1.4 && p.Y >= y-0.7 && p.Y <= y+0.2
	}})
}

// fence runs a low bamboo fence (Kinkakuji-gaki) between the south walk and the pond.
func (gp GoldenPavilion) fence(g *Grid) {
	line := goldenPavilionPaths[0]
	pathZ := func(x float64) float64 {
		for i := 1; i < len(line); i++ {
			if x <= line[i].X {
				t := (x - line[i-1].X) / (line[i].X - line[i-1].X)
				return line[i-1].Y + t*(line[i].Y-line[i-1].Y)
			}
		}
		return line[len(line)-1].Y
	}
	g.Use(Wood)
	for x := -25.0; x <= 23; x += 0.4 {
		z := pathZ(x) - 1.8
		y := gp.ground(x, z)
		if math.IsNaN(y) || gp.pond(x, z) < 0.6 {
			continue
		}
		post := math.Mod(x+25, 2.0) < 0.5
		y0 := y + 0.75
		if post {
			y0 = y - 0.3
		}
		g.Fill(Block(V(x-0.25, y0, z-0.4), V(x+0.25, y+1.3, z+0.4)))
	}
}

// lantern is a stone toro: base, shaft, platform, firebox, flared cap and finial.
func (gp GoldenPavilion) lantern(g *Grid, at Vec2) {
	y := gp.ground(at.X, at.Y)
	if math.IsNaN(y) {
		y = goldenPavilionWater
	}
	y -= 0.3
	g.Use(Stone).Fill(Cylinder{Center: at, Radius: 1.2, Y0: y, Y1: y + 0.7})
	g.Fill(Cylinder{Center: at, Radius: 0.6, Y0: y + 0.7, Y1: y + 2.6})
	g.Fill(BlockAt(at, 2.4, 2.4, y+2.6, y+3.2))
	g.Fill(BlockAt(at, 1.8, 1.8, y+3.2, y+4.6))
	g.Paint(Charcoal, BlockAt(at, 2.0, 1.0, y+3.5, y+4.3), BlockAt(at, 1.0, 2.0, y+3.5, y+4.3))
	g.Use(DarkStone).Fill(Func{Box{V(at.X-1.9, y+4.6, at.Y-1.9), V(at.X+1.9, y+6.0, at.Y+1.9)}, func(p Vec3) bool {
		r := math.Max(math.Abs(p.X-at.X), math.Abs(p.Z-at.Y))
		return p.Y <= y+5.8-r*0.7
	}})
	g.Use(Stone).Fill(Sphere(at.At(y+6.1), 0.55))
}

// teaHouse is a small thatched tea hut on a stone terrace on the hillside.
func (gp GoldenPavilion) teaHouse(g *Grid) {
	c := goldenPavilionTea
	y := goldenPavilionBank + gp.hills(c.X, c.Y)
	g.Use(Stone).Fill(BlockAt(c, 10, 8, y-1.2, y+0.4))
	const hx, hz = 3.2, 2.4
	g.Use(Wood).Fill(BlockAt(c, 2*hx+1.4, 2*hz+1.4, y+0.4, y+0.9))
	gp.storey(g, c, hx, hz, y+0.9, y+4.6, White)
	for _, x := range []float64{-hx + 0.3, 0, hx - 0.3} {
		for _, side := range []float64{-1, 1} {
			g.Use(Wood).Fill(BlockAt(P(c.X+x, c.Y+side*hz), 0.9, 1.2, y+0.9, y+4.6))
		}
	}
	g.Fill(gp.ring(c, hx+0.2, hz+0.2, hx-0.9, hz-0.9, y+3.8, y+4.6))
	gp.openings(g, c, hx, hz, []int{0}, []float64{-1.9, 1.9}, nil, y+1, y+4, gp.rectShape(0.9, y+1.0, y+3.6))
	gp.openings(g, c, hx, hz, []int{1, 3}, nil, []float64{0}, y+1, y+4, gp.rectShape(0.9, y+2.2, y+3.4))
	// Thatch: a thick hipped roof with a dark ridge.
	g.Use(Sand).FillVolume(Func{Box{V(c.X-hx-2, y+4, c.Y-hz-2), V(c.X+hx+2, y+9, c.Y+hz+2)}, func(p Vec3) bool {
		u, v := math.Abs(p.X-c.X)-hx-1.6, math.Abs(p.Z-c.Y)-hz-1.6
		s := y + 4.4 - 1.05*math.Max(u+1.0, v)
		return u <= 0 && v <= 0 && p.Y <= s && p.Y >= s-1.4 && p.Y >= y+4.0
	}})
	g.Use(Charcoal).Fill(Capsule{V(c.X-1.6, y+8.3, c.Y), V(c.X+1.6, y+8.3, c.Y), 0.6})
}

// pine is a cloud-pruned pine: a twisting trunk with flat foliage pads on its branches.
func (gp GoldenPavilion) pine(g *Grid, at Vec2, base, h float64, seed int) {
	a := Hash01(seed, 1) * 2 * math.Pi
	p0 := at.At(base)
	p1 := at.Add(Polar(P(0, 0), h*0.18, a)).At(base + h*0.4)
	p2 := at.Add(Polar(P(0, 0), h*0.1, a+2.2)).At(base + h*0.72)
	p3 := at.Add(Polar(P(0, 0), h*0.2, a+0.6)).At(base + h)
	g.Use(Wood).Fill(Capsule{p0, p1, 0.7}, Capsule{p1, p2, 0.6}, Capsule{p2, p3, 0.55})
	pads := []Vec3{p3}
	for k, src := range []Vec3{p1, p2, p2} {
		dir := Polar(P(0, 0), h*0.32, a+math.Pi*0.9+float64(k)*2.3)
		end := src.Plan().Add(dir).At(src.Y + 0.4)
		g.Use(Wood).Fill(Capsule{src, end, 0.5})
		pads = append(pads, end)
	}
	for k, pd := range pads {
		r := h * (0.34 - 0.05*float64(k%2))
		if k == 0 {
			r = h * 0.3
		}
		g.Use(Foliage).Fill(Ellipsoid{pd.Add(V(0, 0.5, 0)), V(r, 0.95, r*0.85)})
		g.Fill(Ellipsoid{pd.Add(V(0, 1.2, 0)), V(r*0.6, 0.7, r*0.5)})
	}
}

// maple is a round-crowned autumn maple of clustered lobes.
func (GoldenPavilion) maple(g *Grid, at Vec2, base, h float64, seed int) {
	main, speck := Red, Terracotta
	switch v := Hash01(seed, 1); {
	case v < 0.3:
		main, speck = Terracotta, Red
	case v < 0.0:
		main, speck = Gold, Terracotta
	}
	top := at.At(base + h*0.62)
	g.Use(Wood).Fill(Capsule{at.At(base), top, 0.6})
	r := h * 0.36
	lobes := []Solid{Ellipsoid{top.Add(V(0, 0.6, 0)), V(r, r*0.75, r)}}
	for k := 0; k < 3; k++ {
		o := Polar(P(0, 0), r*0.6, Hash01(seed, k, 2)*2*math.Pi)
		lobes = append(lobes, Ellipsoid{top.Add(V(o.X, -0.4+0.8*Hash01(seed, k, 3), o.Y)), V(r*0.7, r*0.55, r*0.7)})
	}
	crown := Union(lobes...)
	g.Use(main).Fill(crown)
	g.Paint(speck, Func{crown.Bounds(), func(p Vec3) bool {
		return crown.Contains(p) && Hash01(int(math.Round(p.X*2)), int(math.Round(p.Y*2)), int(math.Round(p.Z*2)), seed) < 0.25
	}})
}

// cedar is a tall conical Japanese cedar (sugi) rising out of the woodland.
func (GoldenPavilion) cedar(g *Grid, at Vec2, base, h float64) {
	g.Use(Wood).Fill(Capsule{at.At(base), at.At(base + h*0.5), 0.6})
	r := h * 0.27
	g.Use(Foliage).FillVolume(Frustum{at, r, 0.5, base + h*0.25, base + h})
	g.FillVolume(Frustum{at, r * 1.15, 0.2, base + h*0.25, base + h*0.6})
}

// broadleaf is a round evergreen oak crown.
func (GoldenPavilion) broadleaf(g *Grid, at Vec2, base, h float64, seed int) {
	g.Use(Wood).Fill(Capsule{at.At(base), at.At(base + h*0.6), 0.6})
	crown := Ellipsoid{at.At(base + h*0.68), V(h*0.4, h*0.32, h*0.4)}
	g.Use(Foliage).Fill(crown)
	g.Paint(Grass, Func{crown.Bounds(), func(p Vec3) bool {
		return crown.Contains(p) && Hash01(int(math.Round(p.X*2)), int(math.Round(p.Y*2)), int(math.Round(p.Z*2)), seed) < 0.3
	}})
}

// emergents scatters cedars, oaks and maples that rise above the woodland canopy,
// giving the hills a broken, tree-by-tree skyline.
func (gp GoldenPavilion) emergents(g *Grid) {
	const cs = 4.6
	for i := -9; i <= 9; i++ {
		for k := -10; k <= 5; k++ {
			x := (float64(i) + 0.2 + 0.6*Hash01(i, k, 61)) * cs
			z := (float64(k) + 0.2 + 0.6*Hash01(i, k, 62)) * cs
			f := gp.forest(x, z)
			if f < 0.6 || Hash01(i, k, 63) > 0.72 || math.Abs(x) > goldenPavilionHX-4 || z < goldenPavilionZ0+4 {
				continue
			}
			if gp.rim(x, z, gp.hills(x, z)) < 0.5 {
				continue
			}
			s := gp.surface(x, z)
			if math.IsNaN(s) {
				continue
			}
			rimF := gp.rim(x, z, gp.hills(x, z))
			h := (6 + 4*Hash01(i, k, 64)) * (0.5 + 0.5*rimF)
			switch v := Hash01(i, k, 65); {
			case v < 0.5:
				gp.cedar(g, P(x, z), s-2.5, h+3)
			case v < 0.8:
				gp.broadleaf(g, P(x, z), s-2.5, h, 500+i*31+k)
			default:
				gp.maple(g, P(x, z), s-2.5, h, 600+i*31+k)
			}
		}
	}
}

// trees places pines and maples around the pond and behind the pavilion.
func (gp GoldenPavilion) trees(g *Grid) {
	gp.emergents(g)
	for i, t := range []Vec3{
		V(30, 5.5, 5), V(33, 6, -5), V(27, 5, -12), V(32, 5.5, 14), V(30, 4.5, 21),
		V(-31, 5.5, 6), V(-30, 5, -10), V(-33, 6, 15), V(-26, 5, -14), V(-31, 4.5, 22),
		V(34, 5, 0), V(34.5, 5.5, 9), V(-34.5, 5, 0), V(-35, 5.5, 9), V(33, 4, 21), V(-26, 4, 24),

		V(-14, 6.5, -21), V(13, 6.5, -22), V(-21, 6, -16), V(24, 6, -11),
	} {
		y := gp.ground(t.X, t.Z)
		if math.IsNaN(y) {
			continue
		}
		gp.maple(g, P(t.X, t.Z), y-0.3, t.Y+1.5*Hash01(i, 7), 200+i)
	}
	for i, t := range []Vec3{V(-19, 7, -20), V(7, 7.5, -21), V(-5, 7, -21.5), V(-27, 6, -6), V(16, 5, 2), V(23, 5, 2), V(-34, 5, -3), V(-20, 4.5, 24), V(19, 4.5, 24.5), V(29, 5, 12)} {
		y := gp.ground(t.X, t.Z)
		if math.IsNaN(y) {
			continue
		}
		gp.pine(g, P(t.X, t.Z), y-0.3, t.Y, 300+i)
	}
}
