package scenes

import (
	"math"

	"godocta/gen/kit"
	. "godocta/gen/octa"
)

// Neuschwanstein is the fairy-tale Bavarian castle on its rocky crag: the tall white
// Palas with rows of arched windows and a steep slate roof, the slender stair tower and
// north tower, the bower, knights' house and square tower around the upper courtyard,
// the brick gatehouse over the lower courtyard, and, across the Pöllat gorge with its
// waterfall, an iron truss bridge on stone abutments. The ridge runs along X: the Palas
// stands at -X, the gatehouse faces +X, the front (+Z) is the picturesque side.
type Neuschwanstein struct{}

func init() { Register(Neuschwanstein{}) }

func (Neuschwanstein) Name() string  { return "neuschwanstein" }
func (Neuschwanstein) Title() string { return "Neuschwanstein" }

const (
	neuschwansteinLower = 20.0 // Crag plateau: lower courtyard and gatehouse.
	neuschwansteinUpper = 23.0 // Upper courtyard terrace and Palas floor.
	neuschwansteinDeck  = 24.0 // Bridge deck.
)

// neuschwansteinMap is a sampled terrain heightfield with bilinear lookup.
type neuschwansteinMap struct {
	x0, z0, step float64
	nx, nz       int
	h            []float64
}

func (m *neuschwansteinMap) at(x, z float64) float64 {
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

// low is the lowest terrain height around a point, used to close cliffs into shells.
func (m *neuschwansteinMap) low(x, z float64) float64 {
	lo := m.at(x, z)
	for k := 0; k < 8; k++ {
		a := float64(k) * math.Pi / 4
		lo = math.Min(lo, m.at(x+1.2*math.Cos(a), z+1.2*math.Sin(a)))
	}
	return lo
}

// noise is smooth value noise in [0, 1).
func (Neuschwanstein) noise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

// fbm is three octaves of value noise in [0, 1).
func (n Neuschwanstein) fbm(x, z float64, seed int) float64 {
	return (n.noise(x, z, seed)*4 + n.noise(x*2.1, z*2.1, seed+1)*2 + n.noise(x*4.3, z*4.3, seed+2)) / 7
}

// height is the terrain: forested slopes, the castle crag, the eastern saddle and hill,
// and the gorge cut between them. Zero means outside the plot.
func (n Neuschwanstein) height(x, z float64) float64 {
	ex, ez := (x-11)/56, z/27.5
	e := math.Sqrt(ex*ex+ez*ez) + 0.1*(n.fbm(x/10, z/10, 1)-0.5)
	if e > 1 {
		return 0
	}
	// Distance from the castle's rectangular footprint, roughened on the cliffs.
	dx := math.Max(math.Max(-35.5-x, x-33.5), 0)
	dz := math.Max(math.Abs(z)-13, 0)
	d := math.Hypot(dx, dz)
	if d > 0 {
		d += 3.5 * (n.fbm(x/6, z/6, 3) - 0.5) * math.Min(d, 1)
		d = math.Max(d, 0.01)
	}
	crag := neuschwansteinLower
	if d > 0 {
		// Steep cliffs broken by ledges.
		ledge := 1.2 * math.Floor(d/2.5)
		crag = neuschwansteinLower - 2.4*d - ledge*(0.6+n.noise(x/8, z/8, 4))
	}
	forest := 1 + 2.5*n.fbm(x/14, z/14, 2) + 9*math.Exp(-d/7)
	h := math.Max(forest, crag)
	// Saddle from the gatehouse to the gorge rim, carrying the road.
	if x > 30 {
		level := neuschwansteinLower + (neuschwansteinDeck-neuschwansteinLower)*math.Min(1, math.Max(0, (x-33)/12))
		hw := 5.5 + 3*(n.fbm(x/7, z/7, 5)-0.5)
		h = math.Max(h, level-2.6*math.Max(0, math.Abs(z)-hw))
	}
	// Rocky knoll beyond the gorge.
	r := math.Hypot(x-60, z-3)
	h = math.Max(h, neuschwansteinDeck+2-0.09*r*r-3*(n.fbm(x/6, z/6, 6)-0.5))
	// Everything slopes down to the plot's rim.
	h = math.Min(h, 1.5+99*(1-e))
	// The gorge, with a waterfall step under the bridge.
	cx := 52 + 1.5*math.Sin(z/9)
	w := math.Abs(x-cx) - (4.5 + 2*(n.fbm(x/5, z/5, 8)-0.5))
	bottom := 1.0
	if z < -4 {
		bottom = 9
	}
	gorge := bottom + math.Max(0, w)*3.2
	return math.Min(h, gorge)
}

func (n Neuschwanstein) Build(g *Grid) Order {
	m := &neuschwansteinMap{x0: -52, z0: -42, step: 0.4}
	m.nx, m.nz = int(126/m.step)+2, int(84/m.step)+2
	m.h = make([]float64, m.nx*m.nz)
	for j := 0; j < m.nz; j++ {
		for i := 0; i < m.nx; i++ {
			m.h[j*m.nx+i] = n.height(m.x0+float64(i)*m.step, m.z0+float64(j)*m.step)
		}
	}
	n.terrain(g, m)
	n.palas(g)
	n.towers(g)
	n.courtyards(g)
	n.gatehouse(g)
	n.bridge(g)
	n.trees(g, m)
	return Sweep(P(0, 0), math.Pi/2)
}

// terrain fills the heightfield as a hollow shell and colors rock, meadow, forest,
// road and water.
func (n Neuschwanstein) terrain(g *Grid, m *neuschwansteinMap) {
	box := Box{V(m.x0, 0, m.z0), V(m.x0+126, 34, m.z0+84)}
	ground := Func{box, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		if h <= 0 {
			return false
		}
		return p.Y <= h && p.Y >= math.Min(h, m.low(p.X, p.Z))-0.8
	}}
	g.Use(DarkStone).Fill(ground)
	// Rock strata.
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		band := int(math.Floor(p.Y/1.4 + 2*n.noise(p.X/9, p.Z/9, 11)))
		return Hash01(band, int(math.Floor(p.X/5)), int(math.Floor(p.Z/5)), 12) < 0.4
	}})
	// Meadow and forest floor on gentle tops.
	g.Paint(Grass, Func{box, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		return p.Y >= h-1.5 && h-m.low(p.X, p.Z) < 1.3 && h < neuschwansteinLower-1.5 || (h > neuschwansteinDeck+0.5 && p.Y >= h-1.5 && h-m.low(p.X, p.Z) < 1.3)
	}})
	g.Paint(Foliage, Func{box, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		return p.Y >= h-1.5 && h-m.low(p.X, p.Z) < 1.6 && n.fbm(p.X/9, p.Z/9, 13) > 0.5
	}})
	// Road from the gatehouse to the bridge.
	g.Paint(Sand, Func{Box{V(30, 15, -2.2), V(48, 30, 2.2)}, func(p Vec3) bool {
		return p.Y >= m.at(p.X, p.Z)-1.2
	}})
	// Stream and waterfall.
	stream := func(p Vec3) bool { return math.Abs(p.X-52-1.5*math.Sin(p.Z/9)) < 1.8 }
	g.Paint(Water, Func{Box{V(46, 0, m.z0), V(58, 10, m.z0+84)}, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		return stream(p) && h < 10 && p.Y >= h-1 || stream(p) && p.Z < -3 && p.Z > -6 && p.Y > 1
	}})
}

// window is a round-headed opening on a wall's outer face: a shallow recess with a
// dark pane behind it. out is +1 or -1, the face's outward direction along the axis.
func (Neuschwanstein) window(g *Grid, at Vec3, half, spring float64, through Axis, out float64) {
	shift := func(s float64) Vec3 {
		if through == AxisX {
			return at.Add(V(out*s, 0, 0))
		}
		return at.Add(V(0, 0, out*s))
	}
	g.Carve(Arch{Sill: shift(0.3), HalfWidth: half, Spring: spring, Depth: 1.5, Through: through})
	g.Paint(Navy, Arch{Sill: shift(-1.25), HalfWidth: half, Spring: spring, Depth: 1.6, Through: through})
}

// steep is a pitched roof with the given rise per unit across, ridge along an axis.
func (Neuschwanstein) steep(min, max Vec2, base, pitch float64, ridge Axis) Func {
	half := (max.Y - min.Y) / 2
	if ridge == AxisZ {
		half = (max.X - min.X) / 2
	}
	return Func{Box{min.At(base), max.At(base + half*pitch)}, func(p Vec3) bool {
		across := math.Abs(p.Z - (min.Y+max.Y)/2)
		if ridge == AxisZ {
			across = math.Abs(p.X - (min.X+max.X)/2)
		}
		return across <= half-(p.Y-base)/pitch
	}}
}

// roof fills a hollow steep roof shell of horizontal thickness t.
func (n Neuschwanstein) roof(g *Grid, min, max Vec2, base, pitch, t float64, ridge Axis) {
	outer := n.steep(min, max, base, pitch, ridge)
	var inner Func
	if ridge == AxisZ {
		inner = n.steep(P(min.X+t, min.Y-1), P(max.X-t, max.Y+1), base, pitch, ridge)
		inner.Box.Max.Z, inner.Box.Min.Z = max.Y-t, min.Y+t
	} else {
		inner = n.steep(P(min.X-1, min.Y+t), P(max.X+1, max.Y-t), base, pitch, ridge)
		inner.Box.Max.X, inner.Box.Min.X = max.X-t, min.X+t
	}
	g.Fill(Subtract(outer, inner))
}

// finial is a gold spike.
func (Neuschwanstein) finial(g *Grid, at Vec2, y, h float64) {
	g.With(Gold, func() {
		g.Fill(Cylinder{Center: at, Radius: 0.6, Y0: y - 0.5, Y1: y + h})
		g.Fill(Sphere(at.At(y+h*0.4), 0.9))
	})
}

// turret is a slender round tower with corbelled base, a cornice and a pointed roof.
func (n Neuschwanstein) turret(g *Grid, at Vec2, r, base, top, roofH, corbel float64, roof Color) {
	if corbel > 0 {
		g.Fill(Frustum{at, 0.6, r, base - corbel, base})
	}
	g.Fill(Cylinder{Center: at, Radius: r, Y0: base, Y1: top})
	if r > 2 {
		g.Carve(Cylinder{Center: at, Radius: r - 1.2, Y0: base + 1, Y1: top - 1.5})
	}
	g.Fill(Cylinder{Center: at, Radius: r + 0.5, Y0: top - 1, Y1: top})
	g.With(roof, func() { g.FillVolume(Cone(at, r+0.6, top, roofH)) })
	n.finial(g, at, top+roofH, 2)
}

// palas is the main residential block: five storeys of arched windows under a steep
// slate roof, with white gable walls and corner turrets.
func (n Neuschwanstein) palas(g *Grid) {
	const x0, x1, z0, z1 = -34.0, -14.0, -8.0, 8.0
	const base, top = neuschwansteinUpper, 52.0
	const t = 1.2
	g.Use(White)
	g.Fill(Subtract(Block(V(x0, 17, z0), V(x1, top, z1)), Block(V(x0+t, 17, z0+t), V(x1-t, top-1, z1-t))))
	g.Carve(Block(V(x0+t, 17, z0+t), V(x1-t, base-0.6, z1-t)))
	g.Fill(Block(V(x0, base-0.5, z0), V(x1, base, z1)))
	// Stone plinth and string courses.
	g.Paint(Stone, Block(V(x0-1, 0, z0-1), V(x1+1, base+1, z1+1)))
	storey := (top - base) / 5
	for s := 1; s <= 5; s++ {
		y := base + float64(s)*storey - 0.5
		g.Paint(Stone, Block(V(x0-1, y, z0-1), V(x1+1, y+0.4, z1+1)))
	}
	// Windows: paired arches per bay, a full arcade on the Singers' Hall storey.
	for s := 0; s < 5; s++ {
		sill := base + float64(s)*storey + 1.5
		if s == 4 {
			for x := x0 + 2; x <= x1-1.5; x += 2 {
				n.window(g, V(x, sill, z0), 0.7, 2.2, AxisZ, -1)
				n.window(g, V(x, sill, z1), 0.7, 2.2, AxisZ, 1)
			}
			continue
		}
		for x := x0 + 3; x <= x1-2; x += 4.5 {
			for _, off := range []float64{-0.9, 0.9} {
				if s == 0 && x > x0+10 && x < x0+13 {
					continue
				}
				n.window(g, V(x+off, sill, z0), 0.6, 1.8, AxisZ, -1)
				n.window(g, V(x+off, sill, z1), 0.6, 1.8, AxisZ, 1)
			}
		}
		for _, z := range []float64{-4.5, 0, 4.5} {
			n.window(g, V(x0, sill, z), 0.65, 2, AxisX, -1)
			if s > 0 {
				n.window(g, V(x1, sill, z), 0.65, 2, AxisX, 1)
			}
		}
	}
	// Steep roof between white gable walls that rise slightly above it.
	const pitch = 1.7
	g.With(White, func() {
		for _, gx := range []float64{x0, x1 - t} {
			g.Fill(Intersect(Block(V(gx, top, z0), V(gx+t, top+20, z1)), n.steep(P(gx, z0-0.8), P(gx+t, z1+0.8), top, pitch, AxisX)))
		}
	})
	g.With(Slate, func() { n.roof(g, P(x0+t, z0-0.8), P(x1-t, z1+0.8), top-0.5, pitch, 1.6, AxisX) })
	ridge := top + 8.8*pitch
	for _, gx := range []float64{x0 + 0.6, x1 - 0.6} {
		g.Fill(Cylinder{Center: P(gx, 0), Radius: 0.8, Y0: ridge - 2, Y1: ridge + 1.5})
		n.finial(g, P(gx, 0), ridge+1.5, 2)
	}
	// Dormers on both roof slopes.
	for _, x := range []float64{-29, -24, -19} {
		for _, side := range []float64{-1, 1} {
			z := side * 5.2
			zs := math.Min(z, z+side*3)
			g.With(White, func() { g.Fill(Block(V(x-1.3, top+3, zs), V(x+1.3, top+6, zs+3))) })
			g.With(Slate, func() { g.Fill(n.steep(P(x-1.8, zs), P(x+1.8, zs+3), top+6, 1.3, AxisZ)) })
			n.window(g, V(x, top+3.6, z+side*3), 0.6, 1.2, AxisZ, side)
		}
	}
	// Corner turrets on the western gable.
	for _, z := range []float64{z0, z1} {
		g.Use(White)
		n.turret(g, P(x0, z), 1.8, 40, 57, 8, 3, Slate)
	}
	g.Use(White)
	n.turret(g, P(x1, z0), 1.5, 44, 56, 6.5, 3, Slate)
	// Balcony on the west front.
	g.Use(Stone).Fill(Block(V(x0-2, 33, -4), V(x0, 33.6, 4)))
	g.Fill(Block(V(x0-2, 33.6, -4), V(x0-1.4, 34.8, 4)))
	g.Fill(Frustum{P(x0-1, -3), 0.4, 1, 30, 33}, Frustum{P(x0-1, 3), 0.4, 1, 30, 33})
}

// towers raises the slender stair tower and the north tower beside the Palas.
func (n Neuschwanstein) towers(g *Grid) {
	g.Use(White)
	stair := P(-12.5, 7.5)
	const stairTop = 74.0
	g.Fill(Cylinder{Center: stair, Radius: 2.7, Inner: 1.5, Y0: 17, Y1: stairTop})
	for k := 0; k < 14; k++ {
		y := neuschwansteinUpper + 2 + float64(k)*3.5
		if y > stairTop-6 {
			break
		}
		a := float64(k)*0.9 + 0.5
		at := Polar(stair, 2.7, a)
		g.Carve(Func{Box{V(at.X-1.5, y, at.Y-1.5), V(at.X+1.5, y+2.2, at.Y+1.5)}, func(p Vec3) bool {
			return p.Plan().Sub(at).Length() < 0.8
		}})
	}
	g.Paint(Stone, Cylinder{Center: stair, Radius: 4, Y0: 16, Y1: neuschwansteinUpper + 1})
	g.Fill(Frustum{stair, 2.7, 3.4, stairTop - 3, stairTop - 1.5}, Cylinder{Center: stair, Radius: 3.4, Inner: 2.2, Y0: stairTop - 1.5, Y1: stairTop + 1})
	g.With(Slate, func() { g.FillVolume(Cone(stair, 3.1, stairTop+0.5, 13)) })
	n.finial(g, stair, stairTop+13.5, 2.5)
	// Little companion turret corbelled at the top.
	g.Use(White)
	n.turret(g, Polar(stair, 3.1, -math.Pi/4), 1.3, stairTop-8, stairTop+3, 6, 3, Slate)

	// North tower at the back of the Palas.
	north := P(-21, -9.5)
	g.Fill(Cylinder{Center: north, Radius: 3.8, Inner: 2.6, Y0: 14, Y1: 62})
	for k := 0; k < 7; k++ {
		y := neuschwansteinUpper + 3 + float64(k)*5.5
		for _, a := range []float64{-math.Pi / 2, -math.Pi/2 + 1.2, -math.Pi/2 - 1.2} {
			at := Polar(north, 3.8, a+float64(k%2)*0.6)
			g.Carve(Func{Box{V(at.X-1.5, y, at.Y-1.5), V(at.X+1.5, y+2.4, at.Y+1.5)}, func(p Vec3) bool {
				return p.Plan().Sub(at).Length() < 0.85
			}})
		}
	}
	g.Paint(Stone, Cylinder{Center: north, Radius: 4.5, Y0: 0, Y1: neuschwansteinUpper + 1})
	g.Fill(Frustum{north, 3.8, 4.6, 58.5, 60}, Cylinder{Center: north, Radius: 4.6, Inner: 3.4, Y0: 60, Y1: 62.5})
	g.With(Slate, func() { g.FillVolume(Cone(north, 4.3, 62, 12)) })
	n.finial(g, north, 74.5, 2)
}

// courtyards builds the raised upper courtyard with the bower, knights' house and
// square tower, then the walled lower courtyard.
func (n Neuschwanstein) courtyards(g *Grid) {
	up, low := neuschwansteinUpper, neuschwansteinLower
	// Upper terrace: hollow retaining walls under a paved top.
	g.Use(Stone)
	g.Carve(Block(V(-12.8, 17, -11.3), V(8.8, up-1, 11.3)))
	g.Fill(Subtract(Block(V(-14, 17, -12.5), V(10, up, 12.5)), Block(V(-12.8, 17, -11.3), V(8.8, up-1, 11.3))))
	g.With(Sand, func() { g.Fill(Block(V(-14, up-0.5, -12.5), V(10, up, 12.5))) })
	// Steps down to the lower courtyard.
	for k := 0; k < 4; k++ {
		x := 10 + float64(k)*0.9
		g.Fill(Block(V(10, low, -3), V(x+0.9, up-0.75*float64(k), 3)))
	}

	// Outline of the unbuilt keep's foundations, with a lawn inside.
	g.Fill(Subtract(BlockAt(P(-2, 0), 8, 8, up, up+1), BlockAt(P(-2, 0), 6, 6, up-1, up+2)))
	g.With(Grass, func() { g.Fill(BlockAt(P(-2, 0), 6, 6, up-0.5, up)) })
	// Bower on the front side.
	g.Use(White)
	n.house(g, V(-9, up, 6.5), V(2, up+14, 12.5), AxisZ)
	n.turret(g, P(-9, 12.5), 1.5, up+8, up+19, 6, 2.5, Navy)
	// Knights' house at the back, with an open arcade to the courtyard.
	n.house(g, V(-6, up, -12.5), V(9, up+12, -6.5), AxisX)
	for x := -4.0; x <= 7; x += 2.6 {
		g.Carve(Arch{V(x, up, -6.5), 0.9, 2, 2.4, AxisZ})
	}
	// Low battlemented walls closing the courtyard between the buildings.
	g.Use(White)
	g.Fill(Block(V(-14, up, -12.5), V(-6, up+5, -11.3)))
	g.Build(kit.Battlements{Min: P(-14, -12.5), Max: P(-6, -11.3), Top: up + 5, Period: 2})

	// Square tower at the junction of the courtyards.
	g.Use(White)
	const sx0, sx1, sz0, sz1 = 3.0, 9.5, 5.5, 12.5
	const stTop = 54.0
	g.Fill(Subtract(Block(V(sx0, 17, sz0), V(sx1, stTop, sz1)), Block(V(sx0+1.2, 17, sz0+1.2), V(sx1-1.2, stTop-1, sz1-1.2))))
	g.Paint(Stone, Block(V(sx0-1, 0, sz0-1), V(sx1+1, up+1, sz1+1)))
	for y := up + 4; y < stTop-5; y += 6 {
		n.window(g, V((sx0+sx1)/2, y, sz1), 0.7, 1.8, AxisZ, 1)
		n.window(g, V(sx1, y, (sz0+sz1)/2), 0.7, 1.8, AxisX, 1)
		n.window(g, V(sx0, y, (sz0+sz1)/2), 0.7, 1.8, AxisX, -1)
	}
	g.Fill(Block(V(sx0-0.8, stTop-1.5, sz0-0.8), V(sx1+0.8, stTop, sz1+0.8)))
	g.Build(kit.Battlements{Min: P(sx0-0.8, sz0-0.8), Max: P(sx1+0.8, sz1+0.8), Top: stTop, Period: 2.2})
	n.turret(g, P(sx1+0.4, sz1+0.4), 1.7, stTop-6, stTop+8, 7, 3, Navy)

	// Lower courtyard: paved, with crenellated curtain walls.
	g.With(Sand, func() { g.Fill(Block(V(10, low-0.5, -12.5), V(22, low, 12.5))) })
	g.Use(White)
	for _, z := range []float64{-13, 12} {
		g.Fill(Block(V(9.5, 16, z), V(22, low+6, z+1.2)))
		g.Build(kit.Battlements{Min: P(9.5, z), Max: P(22, z+1.2), Top: low + 6, Period: 2})
		g.Paint(Stone, Block(V(9, 0, z-1), V(22.5, low, z+2)))
	}
	for x := 12.0; x < 21; x += 3 {
		n.window(g, V(x, low+2.5, 13.2), 0.5, 1.2, AxisZ, 1)
	}
	// Bower-like Knights' annex on the lower court's back wall.
	n.house(g, V(11, low, -12.5), V(20, low+9, -8), AxisX)
}

// house is a hollow white block with arched windows and a steep slate roof.
func (n Neuschwanstein) house(g *Grid, lo, hi Vec3, ridge Axis) {
	g.Use(White)
	g.Fill(Subtract(Block(lo, hi), Block(lo.Add(V(1.2, 0.5, 1.2)), hi.Sub(V(1.2, 1, 1.2)))))
	g.Paint(Stone, Block(V(lo.X-1, lo.Y, lo.Z-1), V(hi.X+1, lo.Y+0.6, hi.Z+1)))
	for y := lo.Y + 2; y+2.5 < hi.Y; y += 4.5 {
		for x := lo.X + 2; x <= hi.X-1.5; x += 2.6 {
			n.window(g, V(x, y, lo.Z), 0.6, 1.6, AxisZ, -1)
			n.window(g, V(x, y, hi.Z), 0.6, 1.6, AxisZ, 1)
		}
		for z := lo.Z + 2; z <= hi.Z-1.5; z += 2.6 {
			n.window(g, V(lo.X, y, z), 0.6, 1.6, AxisX, -1)
			n.window(g, V(hi.X, y, z), 0.6, 1.6, AxisX, 1)
		}
	}
	const pitch = 1.5
	min, max := P(lo.X-0.6, lo.Z-0.6), P(hi.X+0.6, hi.Z+0.6)
	if ridge == AxisX {
		min.X, max.X = lo.X, hi.X
		half := (hi.Z - lo.Z) / 2
		for _, gx := range []float64{lo.X, hi.X - 1.2} {
			g.Fill(Intersect(Block(V(gx, hi.Y, lo.Z), V(gx+1.2, hi.Y+half*pitch, hi.Z)), n.steep(P(gx, lo.Z-0.6), P(gx+1.2, hi.Z+0.6), hi.Y, pitch, AxisX)))
		}
		min.X, max.X = lo.X+1.2, hi.X-1.2
	} else {
		half := (hi.X - lo.X) / 2
		for _, gz := range []float64{lo.Z, hi.Z - 1.2} {
			g.Fill(Intersect(Block(V(lo.X, hi.Y, gz), V(hi.X, hi.Y+half*pitch, gz+1.2)), n.steep(P(lo.X-0.6, gz), P(hi.X+0.6, gz+1.2), hi.Y, pitch, AxisZ)))
		}
		min.Y, max.Y = lo.Z+1.2, hi.Z-1.2
	}
	g.With(Slate, func() { n.roof(g, min, max, hi.Y-0.5, pitch, 1.5, ridge) })
}

// gatehouse is the brick gate building with stepped gables, a crenellated gate block
// and corner turrets on its outer face.
func (n Neuschwanstein) gatehouse(g *Grid) {
	const x0, x1, z0, z1 = 22.0, 30.0, -12.0, 12.0
	const base, top = neuschwansteinLower, 33.0
	g.Use(Brick)
	g.Fill(Subtract(Block(V(x0, 16, z0), V(x1, top, z1)), Block(V(x0+1.2, 16, z0+1.2), V(x1-1.2, top-1, z1-1.2))))
	// Stepped gables at both ends, slate roof between.
	for _, gz := range []float64{z0, z1 - 1.2} {
		for k := 0; k < 4; k++ {
			hw := 4.6 - 1.1*float64(k)
			g.Fill(Block(V(26-hw, top, gz), V(26+hw, top+1.7*float64(k+1), gz+1.2)))
		}
	}
	g.With(Slate, func() { n.roof(g, P(x0-0.5, z0+1.2), P(x1+0.5, z1-1.2), top-0.5, 1.4, 1.5, AxisZ) })
	// Central gate block rising above the roof.
	g.Fill(Subtract(Block(V(x0-0.5, 16, -4.5), V(x1+1, top+7, 4.5)), Block(V(x0+0.7, 16, -3.3), V(x1-0.2, top+6, 3.3))))
	g.Build(kit.Battlements{Min: P(x0-0.5, -4.5), Max: P(x1+1, 4.5), Top: top + 7, Period: 2})
	g.With(Slate, func() { g.Fill(Block(V(x0+0.7, top+6.5, -3.3), V(x1-0.2, top+7, 3.3))) })
	// Light stone bands and trims.
	for _, y := range []float64{base + 5, base + 9.5, top - 0.5} {
		g.Paint(Terracotta, Block(V(x0-1, y, z0-1), V(x1+2, y+0.5, z1+1)))
	}
	g.Paint(Stone, Block(V(x0-1, 0, z0-1), V(x1+2, base+1, z1+1)))
	// Gate passage along X.
	g.Carve(Arch{V(26, base, 0), 2, 3.5, 12, AxisX})
	g.With(Charcoal, func() { g.Fill(Block(V(x1+0.6, base+3.6, -2), V(x1+1, base+5.5, 2))) })
	for _, z := range []float64{-9, -6.5, 6.5, 9} {
		for _, y := range []float64{base + 6, base + 10.5} {
			n.window(g, V(x1, y, z), 0.6, 1.6, AxisX, 1)
			n.window(g, V(x0, y, z), 0.6, 1.6, AxisX, -1)
		}
	}
	for _, x := range []float64{24, 28} {
		for _, y := range []float64{base + 6, base + 10.5} {
			n.window(g, V(x, y, z1), 0.6, 1.6, AxisZ, 1)
			n.window(g, V(x, y, z0), 0.6, 1.6, AxisZ, -1)
		}
		n.window(g, V(26, top+1.2, z1), 0.5, 1.2, AxisZ, 1)
		n.window(g, V(26, top+1.2, z0), 0.5, 1.2, AxisZ, -1)
	}
	for _, y := range []float64{base + 6, base + 10.5, top + 2} {
		n.window(g, V(x1+1, y, 0), 0.7, 1.7, AxisX, 1)
	}
	// Outer corner turrets.
	for _, z := range []float64{z0, z1} {
		g.Use(Brick)
		n.turret(g, P(x1+0.3, z), 1.6, base+6, top+5, 6, 3, Navy)
	}
}

// bridge is a Marienbrücke-style iron bridge on stone abutments across the gorge: a
// plank deck on a deep X-braced girder, with lattice railings.
func (n Neuschwanstein) bridge(g *Grid) {
	const xa, xb = 45.0, 59.0
	deck := neuschwansteinDeck
	g.Use(Stone)
	for _, x := range []float64{xa - 0.5, xb + 0.5} {
		g.Fill(Block(V(x-1.5, deck-7, -3), V(x+1.5, deck, 3)))
		g.Fill(Block(V(x-1.5, deck, -3), V(x+1.5, deck+1.2, -2.2)), Block(V(x-1.5, deck, 2.2), V(x+1.5, deck+1.2, 3)))
	}
	g.With(Wood, func() { g.Fill(Block(V(xa, deck-0.7, -2), V(xb, deck, 2))) })
	g.Use(Charcoal)
	const depth, rail = 3.0, 1.4
	for _, z := range []float64{-1.6, 1.6} {
		g.Fill(Capsule{V(xa, deck-0.8, z), V(xb, deck-0.8, z), 0.5})
		g.Fill(Capsule{V(xa+1, deck-depth, z), V(xb-1, deck-depth, z), 0.5})
		for x, k := xa+1, 0; x < xb-1; x, k = x+3, k+1 {
			g.Fill(Capsule{V(x, deck-depth, z), V(x, deck-0.8, z), 0.45})
			g.Fill(Capsule{V(x, deck-depth, z), V(x+3, deck-0.8, z), 0.4})
			g.Fill(Capsule{V(x, deck-0.8, z), V(x+3, deck-depth, z), 0.4})
		}
		g.Fill(Capsule{V(xb-1, deck-depth, z), V(xb-1, deck-0.8, z), 0.45})
	}
	for _, z := range []float64{-2.1, 2.1} {
		g.Fill(Capsule{V(xa, deck+rail, z), V(xb, deck+rail, z), 0.4})
		for x := xa + 1; x < xb; x += 1.5 {
			g.Fill(Capsule{V(x, deck, z), V(x, deck+rail, z), 0.35})
		}
	}
}

// trees scatters pines over the gentle forested slopes.
func (n Neuschwanstein) trees(g *Grid, m *neuschwansteinMap) {
	for i := -14; i <= 18; i++ {
		for j := -10; j <= 10; j++ {
			if Hash01(i, j, 21) < 0.35 {
				continue
			}
			at := P(float64(i)*4+3.2*(Hash01(i, j, 22)-0.5), float64(j)*4+3.2*(Hash01(i, j, 23)-0.5))
			h := m.at(at.X, at.Y)
			if h <= 0.5 || h-m.low(at.X, at.Y) > 1.4 {
				continue
			}
			if at.X > -38 && at.X < 36 && math.Abs(at.Y) < 15 {
				continue // Castle crag.
			}
			if at.X > 30 && at.X < 48 && math.Abs(at.Y) < 4 {
				continue // Road.
			}
			if math.Abs(at.X-52) < 7 {
				continue // Gorge.
			}
			edge := math.Pow(math.Abs(at.X-12)/60, 3) + math.Pow(math.Abs(at.Y)/36, 3)
			if edge > 0.8 {
				continue
			}
			size := 6 + 5*Hash01(i, j, 24)
			n.pine(g, at, h-0.5, size)
		}
	}
}

// pine is a trunk with three stacked foliage cones.
func (Neuschwanstein) pine(g *Grid, at Vec2, base, h float64) {
	g.With(Wood, func() { g.Fill(Cylinder{Center: at, Radius: 0.55, Y0: base - 1, Y1: base + h*0.3}) })
	g.With(Foliage, func() {
		g.Fill(Cone(at, h*0.27, base+h*0.18, h*0.42), Cone(at, h*0.21, base+h*0.42, h*0.36), Cone(at, h*0.14, base+h*0.66, h*0.34))
	})
}
