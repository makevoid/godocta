package scenes

import (
	"math"

	"godocta/gen/kit"
	. "godocta/gen/octa"
)

// TowerBridge is London's Tower Bridge over the Thames, seen with its bascules open.
// The bridge runs along X across the river, which flows along Z. Two Gothic stone
// towers on boat-shaped piers carry the high-level lattice walkways; the bascule
// leaves are raised between them; blue and white suspension chains curve down from
// the towers to the shore abutment towers, and approach roads continue onto the
// embankments. On the north bank (-X) stands a hint of the Tower of London: the White
// Tower keep with its four corner turrets inside a towered curtain wall. A red
// double-decker bus and a few cars cross the side spans, a tour boat passes under the
// open leaves, and HMS Belfast lies moored off the south bank. North is -X, so the
// river flows towards -Z; the Tower of London stands upstream of the bridge.
type TowerBridge struct{}

func init() { Register(TowerBridge{}) }

func (TowerBridge) Name() string  { return "tower_bridge" }
func (TowerBridge) Title() string { return "Tower Bridge" }

const (
	towerBridgeDeck    = 4.5  // Road surface height; the river surface is y = 0.
	towerBridgeRoad    = 4.0  // Half width of the roadway between the girders.
	towerBridgeMainX   = 20.5 // Main tower centers at ±X.
	towerBridgeMainHX  = 5.0  // Main tower half extents.
	towerBridgeMainHZ  = 7.0
	towerBridgeMainTop = 28.0 // Main tower parapet.
	towerBridgeShoreX  = 70.0 // Abutment tower centers at ±X.
	towerBridgeShoreHX = 3.5
	towerBridgeShoreHZ = 5.5
	towerBridgeShoreTo = 17.0 // Abutment tower parapet.
	towerBridgeBank    = 67.5 // Embankment wall at ±X.
	towerBridgeBankTop = 3.0  // Embankment ground level.
	towerBridgeNorth   = 112.0
	towerBridgeSouth   = 94.0
	towerBridgeHalfZ   = 34.0 // River and banks span ±Z.
	towerBridgeWalkY   = 21.6 // Walkway floor.
	towerBridgeWalkTop = 25.8 // Walkway roof.
	towerBridgeLift    = 62 * math.Pi / 180
)

func (t TowerBridge) Build(g *Grid) Order {
	t.river(g)
	t.banks(g)
	for _, side := range []float64{-1, 1} {
		t.mainTower(g, side)
		t.shoreTower(g, side)
		t.sideSpan(g, side)
		t.chain(g, side)
		t.bascule(g, side)
		t.approach(g, side)
	}
	t.walkways(g)
	t.towerOfLondon(g)
	t.dock(g)
	t.southBank(g)
	t.bus(g, P(46, -1.9))
	t.car(g, P(-50, 1.9), Charcoal)
	t.car(g, P(-37, -1.9), White)
	t.car(g, P(33, 1.9), Navy)
	t.boat(g, P(-1.5, 15))
	t.belfast(g)
	// Masons work in from both banks towards the open span.
	return func(p Vec3) float64 { return -math.Abs(p.X) + 0.01*p.Z }
}

// raster sets every cell whose center lies in the box and for which f reports a color,
// without half-unit snapping, so slanted girders and lattices stay crisp.
func (TowerBridge) raster(g *Grid, b Box, f func(Vec3) (Color, bool)) {
	s := g.Span()
	lo := func(v float64) int32 { return int32(math.Ceil(v * 2 / s)) }
	hi := func(v float64) int32 { return int32(math.Floor(v * 2 / s)) }
	previous := g.Brush()
	for y := lo(b.Min.Y); y <= hi(b.Max.Y); y++ {
		z0, x0 := lo(b.Min.Z), lo(b.Min.X)
		if (z0^y)&1 != 0 {
			z0++
		}
		if (x0^y)&1 != 0 {
			x0++
		}
		for z := z0; z <= hi(b.Max.Z); z += 2 {
			for x := x0; x <= hi(b.Max.X); x += 2 {
				c := Cell{X: x, Y: y, Z: z}
				if color, ok := f(g.World(c)); ok {
					g.Remove(c)
					g.Use(color).Add(c)
				}
			}
		}
	}
	g.Use(previous)
}

// pointed is a Gothic (equilateral) pointed arch opening: jambs up to Spring above the
// sill, then two arcs meeting at a point.
func (TowerBridge) pointed(sill Vec3, hw, spring, depth float64, through Axis) Solid {
	top := spring + math.Sqrt(3)*hw
	span := V(hw, 0, depth/2)
	if through == AxisX {
		span = V(depth/2, 0, hw)
	}
	box := Box{sill.Sub(span), sill.Add(span).Add(V(0, top, 0))}
	return Func{Box: box, F: func(p Vec3) bool {
		d := p.Sub(sill)
		across := d.X
		if through == AxisX {
			across = d.Z
		}
		if math.Abs(across) > hw || d.Y < 0 {
			return false
		}
		rise := d.Y - spring
		if rise <= 0 {
			return true
		}
		r := 2 * hw
		return (across+hw)*(across+hw)+rise*rise <= r*r && (across-hw)*(across-hw)+rise*rise <= r*r
	}}
}

// octagon is a vertical octagonal prism.
func (TowerBridge) octagon(center Vec2, r, y0, y1 float64) Prism {
	poly := make([]Vec2, 8)
	for i := range poly {
		poly[i] = Polar(center, r, (float64(i)+0.5)*math.Pi/4)
	}
	return Prism{Polygon: poly, Y0: y0, Y1: y1}
}

// spire is an octagonal pyramid of the given height.
func (TowerBridge) spire(center Vec2, r, base, height float64) Solid {
	b := Box{V(center.X-r, base, center.Y-r), V(center.X+r, base+height, center.Y+r)}
	return Func{Box: b, F: func(p Vec3) bool {
		rr := r * (1 - (p.Y-base)/height)
		d := p.Plan().Sub(center)
		// Octagon "radius" along the nearest face normal.
		a := math.Mod(math.Abs(d.Angle()), math.Pi/4)
		return d.Length()*math.Cos(a-math.Pi/8) <= rr*math.Cos(math.Pi/8)
	}}
}

// steepRoof is a hipped roof over a rectangle with horizontal inset k per unit height.
func (TowerBridge) steepRoof(min, max Vec2, base, k float64) Solid {
	rise := math.Min(max.X-min.X, max.Y-min.Y) / 2 / k
	b := Box{min.At(base), max.At(base + rise)}
	return Func{Box: b, F: func(p Vec3) bool {
		in := (p.Y - base) * k
		return p.X >= min.X+in && p.X <= max.X-in && p.Z >= min.Y+in && p.Z <= max.Y-in
	}}
}

// river is a single layer of water between and beyond the embankments.
func (TowerBridge) river(g *Grid) {
	g.Use(Water).Fill(Block(V(-towerBridgeBank, -0.1, -towerBridgeHalfZ), V(towerBridgeBank, 0.1, towerBridgeHalfZ)))
}

// banks are hollow embankment plinths: a retaining wall facing the river, skirt walls
// and a two-layer ground slab, with a riverside parapet.
func (t TowerBridge) banks(g *Grid) {
	for _, side := range []float64{-1, 1} {
		far := towerBridgeSouth
		if side < 0 {
			far = towerBridgeNorth
		}
		lo, hi := math.Min(side*towerBridgeBank, side*far), math.Max(side*towerBridgeBank, side*far)
		top := towerBridgeBankTop
		g.Use(DarkStone)
		// River wall, skirts.
		g.Fill(Block(V(lo, -0.3, -towerBridgeHalfZ), V(hi, top, -towerBridgeHalfZ+1)))
		g.Fill(Block(V(lo, -0.3, towerBridgeHalfZ-1), V(hi, top, towerBridgeHalfZ)))
		edge := side * towerBridgeBank
		g.Fill(Block(V(edge-0.6, -0.3, -towerBridgeHalfZ), V(edge+0.6, top, towerBridgeHalfZ)))
		g.Fill(Block(V(side*far-0.5, -0.3, -towerBridgeHalfZ), V(side*far+0.5, top, towerBridgeHalfZ)))
		// Ground slab: grass, with a paved riverside walk.
		// One lattice layer, just below the ground level that components stand on.
		g.Use(Grass)
		t.raster(g, Box{V(lo, top-0.3, -towerBridgeHalfZ), V(hi, top-0.1, towerBridgeHalfZ)}, func(Vec3) (Color, bool) { return Grass, true })
		walk := Block(V(edge-side*0.6, top-0.8, -towerBridgeHalfZ), V(edge+side*5, top, towerBridgeHalfZ))
		g.Paint(Sand, walk)
		g.Paint(DarkStone, Block(V(lo, top-0.8, -towerBridgeHalfZ), V(hi, top, -towerBridgeHalfZ+1)),
			Block(V(lo, top-0.8, towerBridgeHalfZ-1), V(hi, top, towerBridgeHalfZ)))
		// Riverside parapet, broken for the abutment.
		g.Use(Stone)
		for _, z := range [][2]float64{{-towerBridgeHalfZ, -towerBridgeShoreHZ - 2}, {towerBridgeShoreHZ + 2, towerBridgeHalfZ}} {
			g.Fill(Block(V(edge-0.5, top, z[0]), V(edge+0.5, top+0.9, z[1])))
		}
	}
}

// towerBody builds a Gothic tower: a hollow rectangular shell with a road arch through
// X, lancet windows, string courses, octagonal corner turrets with slate spires and gold
// finials, and a steep hipped slate roof with a flèche.
func (t TowerBridge) towerBody(g *Grid, c Vec2, hx, hz, y0, top, turret float64, main bool) {
	wall := 1.2
	g.Use(Stone)
	body := Block(V(c.X-hx, y0, c.Y-hz), V(c.X+hx, top, c.Y+hz))
	g.Fill(body)
	g.Carve(Block(V(c.X-hx+wall, y0+0.5, c.Y-hz+wall), V(c.X+hx-wall, top-0.8, c.Y+hz-wall)))
	// Road arch through the tower.
	hw, spring := towerBridgeRoad-0.2, 3.6
	if !main {
		spring = 1.6
	}
	g.Carve(t.pointed(V(c.X, towerBridgeDeck, c.Y), hw, spring, 2*hx+2, AxisX))
	archTop := towerBridgeDeck + spring + math.Sqrt(3)*hw
	// Vault ceiling over the roadway.
	g.Fill(Block(V(c.X-hx, archTop+0.4, c.Y-hz), V(c.X+hx, archTop+1.2, c.Y+hz)))
	// Sand voussoir band around the arch on both faces.
	for _, sx := range []float64{-1, 1} {
		ring := Subtract(t.pointed(V(c.X+sx*hx, towerBridgeDeck, c.Y), hw+0.9, spring, 0.8, AxisX),
			t.pointed(V(c.X+sx*hx, towerBridgeDeck, c.Y), hw, spring, 2, AxisX))
		g.Paint(Sand, ring)
	}
	// String courses.
	shell := func(y float64) Solid {
		return Subtract(Block(V(c.X-hx, y, c.Y-hz), V(c.X+hx, y+0.4, c.Y+hz)),
			Block(V(c.X-hx+0.5, y, c.Y-hz+0.5), V(c.X+hx-0.5, y+0.4, c.Y+hz-0.5)))
	}
	courses := []float64{towerBridgeDeck - 0.4, archTop + 1.6, top - 1.2}
	if main {
		courses = append(courses, towerBridgeWalkY-0.6)
	}
	for _, y := range courses {
		g.Paint(Sand, shell(y))
	}
	// Corbelled parapet band.
	g.With(Sand, func() {
		g.Fill(Subtract(Block(V(c.X-hx-0.5, top-0.5, c.Y-hz-0.5), V(c.X+hx+0.5, top+0.6, c.Y+hz+0.5)),
			Block(V(c.X-hx+0.6, top-0.5, c.Y-hz+0.6), V(c.X+hx-0.6, top+0.6, c.Y+hz-0.6))))
	})
	// Lancet windows with sand surrounds: paired on the river faces (±Z), and above
	// the arch on the road faces (±X). Each row is a sill height and an opening height.
	type row struct{ sill, h float64 }
	rows := []row{{towerBridgeBankTop + 2.2, 3.2}, {archTop + 1.0, 2.0}}
	xs, zs := []float64{-hx / 2.6, hx / 2.6}, []float64{0}
	if main {
		rows = []row{{towerBridgeDeck + 1.2, 3.6}, {towerBridgeDeck + 6.6, 3.6}, {archTop + 2.2, 3.4}, {towerBridgeWalkY + 0.6, 2.8}}
		xs, zs = []float64{-hx / 2.2, hx / 2.2}, []float64{-hz / 2.6, hz / 2.6}
	}
	const lw = 0.95
	window := func(sill Vec3, h float64, through Axis) {
		g.Paint(Sand, Subtract(t.pointed(sill.Sub(V(0, 0.5, 0)), lw+0.55, h-lw*1.73+0.5, 3, through),
			t.pointed(sill, lw, h-lw*1.73, 4, through)))
		g.Carve(t.pointed(sill, lw, h-lw*1.73, 4, through))
	}
	for i, r := range rows {
		for _, dx := range xs {
			for _, sz := range []float64{-1, 1} {
				window(V(c.X+dx, r.sill, c.Y+sz*hz), r.h, AxisZ)
			}
		}
		if r.sill < archTop {
			continue
		}
		for _, dz := range zs {
			if main && i == 3 {
				dz = 0 // The walkways meet the tower on either side.
			}
			for _, sx := range []float64{-1, 1} {
				window(V(c.X+sx*hx, r.sill, c.Y+dz), r.h, AxisX)
			}
		}
	}
	// Corner turrets.
	rise := 2.6
	if !main {
		rise = 1.8
	}
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			tc := P(c.X+sx*hx, c.Y+sz*hz)
			g.Use(Stone).Fill(t.octagon(tc, turret, y0, top+rise))
			g.Carve(t.octagon(tc, turret-1.0, y0+0.5, top+rise-0.6))
			g.Paint(Sand, t.octagon(tc, turret+0.1, top-1.2, top-0.8))
			g.With(Sand, func() { g.Fill(t.octagon(tc, turret+0.45, top+rise-0.6, top+rise)) })
			// Little lancet slits on the outer faces.
			for y := archTop + 3.0; y+2 < top; y += 5.5 {
				g.Carve(t.pointed(V(tc.X+sx*turret, y, tc.Y), 0.7, 1.0, 1.6, AxisX))
			}
			height := turret * 3.2
			g.With(Slate, func() { g.FillVolume(t.spire(tc, turret+0.4, top+rise, height)) })
			g.With(Gold, func() {
				g.Fill(Capsule{tc.At(top + rise + height - 0.4), tc.At(top + rise + height + 1.2), 0.45})
			})
		}
	}
	// Steep slate roof, hollow, with dormers and a gold flèche.
	k := 0.45
	if !main {
		k = 0.75
	}
	min, max := P(c.X-hx+0.6, c.Y-hz+0.6), P(c.X+hx-0.6, c.Y+hz-0.6)
	roofTop := top + 0.6 + (max.X-min.X)/2/k
	g.With(Slate, func() {
		g.FillVolume(t.steepRoof(min, max, top+0.6, k))
		g.Carve(Translate(t.steepRoof(min, max, top+0.6, k), V(0, -1.4, 0)))
	})
	if main {
		// Gabled dormers on the river faces.
		for _, sz := range []float64{-1, 1} {
			z := c.Y + sz*(hz-2.2)
			g.With(Stone, func() {
				g.Fill(Block(V(c.X-1.2, top+0.6, z-1), V(c.X+1.2, top+3.2, z+1)))
			})
			g.With(Slate, func() {
				g.Fill(Gable{Min: P(c.X-1.6, z-1.2), Max: P(c.X+1.6, z+1.2), Base: top + 3.2, Ridge: AxisZ})
			})
			g.Carve(t.pointed(V(c.X, top+1.0, z+sz*1.0), 0.5, 0.9, 1.2, AxisZ))
		}
		g.With(Sand, func() { g.Fill(t.octagon(c, 0.9, roofTop-1.2, roofTop+1.5)) })
		g.With(Slate, func() { g.FillVolume(t.spire(c, 1.0, roofTop+1.5, 3.5)) })
		g.With(Gold, func() { g.Fill(Capsule{c.At(roofTop + 4.6), c.At(roofTop + 6.6), 0.45}) })
	} else {
		g.With(Gold, func() { g.Fill(Capsule{c.At(roofTop - 0.4), c.At(roofTop + 1.6), 0.45}) })
	}
}

// mainTower is a river tower on its pier, with the roadway through it.
func (t TowerBridge) mainTower(g *Grid, side float64) {
	cx := side * towerBridgeMainX
	// Boat-shaped granite pier with cutwaters up- and downstream.
	pier := func(hx, hz, tip, y0, y1 float64) Prism {
		return Prism{Polygon: []Vec2{
			P(cx-hx, -hz), P(cx, -hz-tip), P(cx+hx, -hz), P(cx+hx, hz), P(cx, hz+tip), P(cx-hx, hz),
		}, Y0: y0, Y1: y1}
	}
	g.Use(DarkStone).Fill(pier(6.6, 9.0, 3.6, -0.3, towerBridgeDeck-1.0))
	g.Carve(pier(5.4, 7.8, 2.8, 0.5, towerBridgeDeck-1.8))
	g.Paint(Stone, pier(6.7, 9.1, 3.7, towerBridgeDeck-1.6, towerBridgeDeck-1.0))
	t.towerBody(g, P(cx, 0), towerBridgeMainHX, towerBridgeMainHZ, towerBridgeDeck-1.0, towerBridgeMainTop, 1.9, true)
	// Roadway through the tower.
	g.With(Charcoal, func() {
		g.Fill(Block(V(cx-towerBridgeMainHX, towerBridgeDeck-0.8, -towerBridgeRoad), V(cx+towerBridgeMainHX, towerBridgeDeck, towerBridgeRoad)))
	})
}

// shoreTower is an abutment tower at the river's edge, anchoring the chains.
func (t TowerBridge) shoreTower(g *Grid, side float64) {
	cx := side * towerBridgeShoreX
	g.Use(DarkStone).Fill(Block(V(cx-towerBridgeShoreHX-0.8, -0.3, -towerBridgeShoreHZ-0.8), V(cx+towerBridgeShoreHX+0.8, towerBridgeBankTop, towerBridgeShoreHZ+0.8)))
	t.towerBody(g, P(cx, 0), towerBridgeShoreHX, towerBridgeShoreHZ, towerBridgeBankTop, towerBridgeShoreTo, 1.3, false)
	g.With(Charcoal, func() {
		g.Fill(Block(V(cx-towerBridgeShoreHX, towerBridgeDeck-0.8, -towerBridgeRoad), V(cx+towerBridgeShoreHX, towerBridgeDeck, towerBridgeRoad)))
	})
	// Solid footing under the roadway inside the tower.
	g.With(DarkStone, func() {
		g.Fill(Block(V(cx-towerBridgeShoreHX, towerBridgeBankTop, -towerBridgeRoad), V(cx+towerBridgeShoreHX, towerBridgeDeck-0.8, towerBridgeRoad)))
	})
}

// sideSpan is the suspended deck between a main tower and its abutment: a charcoal road
// between navy stiffening girders with white rails.
func (t TowerBridge) sideSpan(g *Grid, side float64) {
	x0 := towerBridgeMainX + towerBridgeMainHX
	x1 := towerBridgeShoreX - towerBridgeShoreHX
	lo, hi := math.Min(side*x0, side*x1), math.Max(side*x0, side*x1)
	g.With(Charcoal, func() {
		g.Fill(Block(V(lo, towerBridgeDeck-0.8, -towerBridgeRoad), V(hi, towerBridgeDeck, towerBridgeRoad)))
	})
	g.With(Sand, func() {
		g.Fill(Block(V(lo, towerBridgeDeck-0.8, -towerBridgeRoad), V(hi, towerBridgeDeck, -towerBridgeRoad+1.2)),
			Block(V(lo, towerBridgeDeck-0.8, towerBridgeRoad-1.2), V(hi, towerBridgeDeck, towerBridgeRoad)))
	})
	g.With(Navy, func() {
		for _, sz := range []float64{-1, 1} {
			z0, z1 := sz*towerBridgeRoad, sz*(towerBridgeRoad+1.0)
			g.Fill(Block(V(lo, towerBridgeDeck-1.8, math.Min(z0, z1)), V(hi, towerBridgeDeck+1.0, math.Max(z0, z1))))
		}
		// Cross girders under the deck.
		for x := lo + 2; x < hi; x += 4 {
			g.Fill(Block(V(x-0.5, towerBridgeDeck-1.8, -towerBridgeRoad), V(x+0.5, towerBridgeDeck-0.8, towerBridgeRoad)))
		}
	})
	g.With(White, func() {
		for _, sz := range []float64{-1, 1} {
			z0, z1 := sz*towerBridgeRoad, sz*(towerBridgeRoad+1.0)
			g.Fill(Block(V(lo, towerBridgeDeck+1.0, math.Min(z0, z1)), V(hi, towerBridgeDeck+1.3, math.Max(z0, z1))))
		}
	})
}

// chainY is the height of the suspension chain's centerline at |x|.
func (TowerBridge) chainY(x float64) float64 {
	x0 := towerBridgeMainX + towerBridgeMainHX
	x1 := towerBridgeShoreX - towerBridgeShoreHX
	y0, y1, ymin := towerBridgeWalkY+1.9, towerBridgeShoreTo+0.4, towerBridgeDeck+3.3
	r0, r1 := math.Sqrt(y0-ymin), math.Sqrt(y1-ymin)
	sa := (r0 + r1) / (x1 - x0)
	vx := x0 + r0/sa
	return ymin + sa*sa*(x-vx)*(x-vx)
}

// chain is a pair of stiffened lattice suspension chains, navy chords with white
// Warren bracing, with navy hangers dropping to the deck.
func (t TowerBridge) chain(g *Grid, side float64) {
	x0 := towerBridgeMainX + towerBridgeMainHX - 1.0
	x1 := towerBridgeShoreX - towerBridgeShoreHX + 1.0
	const zc, half, depth, panel = towerBridgeRoad + 0.5, 0.55, 1.1, 3.0
	b := Box{V(math.Min(side*x0, side*x1), towerBridgeDeck, -zc-1), V(math.Max(side*x0, side*x1), towerBridgeWalkY+4, zc+1)}
	t.raster(g, b, func(p Vec3) (Color, bool) {
		if math.Abs(math.Abs(p.Z)-zc) > half {
			return 0, false
		}
		ax := math.Abs(p.X)
		dy := p.Y - t.chainY(ax)
		if math.Abs(dy) > depth+0.5 {
			// Hangers.
			f := math.Mod(ax-x0, panel)
			if dy < 0 && p.Y >= towerBridgeDeck && math.Abs(f-panel/2) <= 0.4 && ax > x0+1 && ax < x1-1 {
				return Navy, true
			}
			return 0, false
		}
		if math.Abs(dy) > depth-0.5 {
			return Navy, true
		}
		u := (ax - x0) / panel
		f := u - math.Floor(u)
		target := -depth + 2*depth*f
		if int(math.Floor(u))%2 == 1 {
			target = -target
		}
		if math.Abs(dy-target) <= 0.5 || f < 0.12 || f > 0.88 {
			return White, true
		}
		return 0, false
	})
}

// bascule is one raised leaf of the opening span, hinged at its tower's inner face.
func (t TowerBridge) bascule(g *Grid, side float64) {
	px := side * (towerBridgeMainX - towerBridgeMainHX)
	length := towerBridgeMainX - towerBridgeMainHX - 0.4
	cs, sn := math.Cos(towerBridgeLift), math.Sin(towerBridgeLift)
	dir := -side // Leaves extend towards the center line.
	b := Box{V(math.Min(px, px+dir*length*cs)-2, towerBridgeDeck-2, -towerBridgeRoad-1.2), V(math.Max(px, px+dir*length*cs)+2, towerBridgeDeck+length*sn+2, towerBridgeRoad+1.2)}
	t.raster(g, b, func(p Vec3) (Color, bool) {
		dx, dy := (p.X-px)*dir, p.Y-towerBridgeDeck
		u := dx*cs + dy*sn
		v := -dx*sn + dy*cs
		az := math.Abs(p.Z)
		if u < 0 || u > length || az > towerBridgeRoad+0.8 {
			return 0, false
		}
		switch {
		case v >= -0.8 && v <= 0 && az <= towerBridgeRoad-1.0:
			if az < 0.35 && math.Mod(u, 3) < 1.6 {
				return White, true
			}
			return Charcoal, true
		case v >= -0.8 && v <= 0:
			return Sand, true
		case v >= -2.2 && v < -0.8 && (az >= towerBridgeRoad-0.9 || math.Mod(u+0.5, 3.75) < 1.0):
			return Navy, true
		case v > 0 && v <= 1.1 && az >= towerBridgeRoad-0.1:
			if v > 0.75 {
				return White, true
			}
			return Navy, true
		}
		return 0, false
	})
}

// walkways are the two high-level lattice footbridges between the main towers.
func (t TowerBridge) walkways(g *Grid) {
	x0 := towerBridgeMainX - towerBridgeMainHX
	const z0, z1, panel = 3.8, 7.0, 3.875
	y0, y1 := towerBridgeWalkY, towerBridgeWalkTop
	b := Box{V(-x0-0.5, y0, -z1), V(x0+0.5, y1+1.2, z1)}
	t.raster(g, b, func(p Vec3) (Color, bool) {
		az := math.Abs(p.Z)
		if az < z0 || az > z1 {
			return 0, false
		}
		// Floor and roof chords.
		if p.Y <= y0+0.8 || p.Y >= y1-0.8 {
			if p.Y > y1 {
				// Low ridge cap over the roof.
				if math.Abs(az-(z0+z1)/2) <= 0.9 {
					return Slate, true
				}
				return 0, false
			}
			return Navy, true
		}
		if az > z0+1.0 && az < z1-1.0 {
			return 0, false // Hollow inside.
		}
		u := (p.X + x0) / panel
		f := u - math.Floor(u)
		if f < 0.1 || f > 0.9 {
			return Navy, true
		}
		h := (p.Y - y0 - 0.8) / (y1 - y0 - 1.6)
		if math.Abs(f-h) < 0.11 || math.Abs(1-f-h) < 0.11 {
			return Slate, true
		}
		return 0, false
	})
}

// approach is the raised approach road running from an abutment onto the bank.
func (t TowerBridge) approach(g *Grid, side float64) {
	x0 := towerBridgeShoreX + towerBridgeShoreHX
	far := towerBridgeSouth
	if side < 0 {
		far = towerBridgeNorth
	}
	ramp := 12.0
	top := func(ax float64) float64 {
		return towerBridgeDeck - math.Max(0, math.Min(1, (ax-x0)/ramp))*(towerBridgeDeck-towerBridgeBankTop)
	}
	b := Box{V(math.Min(side*x0, side*far), towerBridgeBankTop-0.9, -towerBridgeRoad-0.8), V(math.Max(side*x0, side*far), towerBridgeDeck+1.2, towerBridgeRoad+0.8)}
	t.raster(g, b, func(p Vec3) (Color, bool) {
		ax, az := math.Abs(p.X), math.Abs(p.Z)
		h := top(ax)
		if az <= towerBridgeRoad {
			if p.Y > h || p.Y < h-0.85 {
				return 0, false
			}
			if az > towerBridgeRoad-1.2 {
				return Sand, true
			}
			if az < 0.35 && math.Mod(ax, 3) < 1.6 {
				return White, true
			}
			return Charcoal, true
		}
		// Parapet walls while the road is raised.
		if p.Y <= h+0.9 && p.Y >= towerBridgeBankTop-0.9 && h > towerBridgeBankTop+0.2 {
			return Stone, true
		}
		return 0, false
	})
	// Lamp posts with gold lanterns along the approach.
	for x := x0 + 4; x < far-4; x += 7 {
		for _, sz := range []float64{-1, 1} {
			at := P(side*x, sz*(towerBridgeRoad+0.4))
			base := top(x)
			g.With(Charcoal, func() { g.Fill(Cylinder{Center: at, Radius: 0.4, Y0: base, Y1: base + 3}) })
			g.With(Gold, func() { g.Fill(Cylinder{Center: at, Radius: 0.6, Y0: base + 3, Y1: base + 3.8}) })
		}
	}
}

// towerOfLondon places the White Tower inside a towered curtain wall on the north bank,
// upstream (+Z) of the bridge. North is -X, so east (downstream) is -Z.
func (t TowerBridge) towerOfLondon(g *Grid) {
	ground := towerBridgeBankTop
	minX, maxX, minZ, maxZ := -108.0, -80.0, 7.0, 30.0
	// Inner ward lawn and paths.
	g.Paint(Sand, Block(V(minX, ground-0.8, minZ), V(maxX, ground, maxZ)))
	g.Paint(Grass, Block(V(minX+2.5, ground-0.8, minZ+2.5), V(maxX-2.5, ground, maxZ-2.5)))
	g.Paint(Sand, Block(V(-88, ground-0.8, 18.5), V(maxX, ground, 21.5)))
	g.Use(Stone)
	g.Build(
		kit.Wall{Min: V(minX, ground, minZ), Max: V(maxX, ground+5, minZ+1.4)},
		kit.Wall{Min: V(minX, ground, maxZ-1.4), Max: V(maxX, ground+5, maxZ)},
		kit.Wall{Min: V(minX, ground, minZ), Max: V(minX+1.4, ground+5, maxZ)},
		kit.Wall{Min: V(maxX-1.4, ground, minZ), Max: V(maxX, ground+5, maxZ)},
	)
	for _, x := range []float64{minX, maxX} {
		for _, z := range []float64{minZ, maxZ} {
			g.Build(kit.Tower{Center: P(x, z), Radius: 2.4, Base: ground, Top: ground + 7.5, Storey: 4, Merlons: 8})
		}
	}
	g.Build(kit.Tower{Center: P(-95, minZ), Radius: 2.0, Base: ground, Top: ground + 6.5, Storey: 4, Merlons: 8},
		kit.Tower{Center: P(-95, maxZ), Radius: 2.0, Base: ground, Top: ground + 6.5, Storey: 4, Merlons: 8},
		kit.Tower{Center: P(minX, 20), Radius: 2.0, Base: ground, Top: ground + 6.5, Storey: 4, Merlons: 8})
	// Riverside gatehouse facing the bridge.
	gate := P(maxX, 20)
	g.Fill(Block(V(gate.X-2, ground, gate.Y-3.5), V(gate.X+2, ground+7, gate.Y+3.5)))
	g.Carve(Arch{Sill: V(gate.X, ground, gate.Y), HalfWidth: 1.5, Spring: 2.0, Depth: 6, Through: AxisX})
	g.Build(kit.Battlements{Min: P(gate.X-2, gate.Y-3.5), Max: P(gate.X+2, gate.Y+3.5), Top: ground + 7})

	// The White Tower: pale Norman keep, crenellated, with three square and one round
	// corner turret under lead ogee cupolas with gold vanes.
	k0, k1 := P(-102, 14), P(-88, 27)
	top := ground + 12
	g.Use(White)
	g.Fill(Block(k0.At(ground), k1.At(top)))
	g.Carve(Block(V(k0.X+1.2, ground+0.5, k0.Y+1.2), V(k1.X-1.2, top-0.8, k1.Y-1.2)))
	g.Fill(Block(V(k0.X+1.2, ground+5.6, k0.Y+1.2), V(k1.X-1.2, ground+6.2, k1.Y-1.2)))
	g.Build(kit.Battlements{Min: k0, Max: k1, Top: top, Period: 2.4})
	// Pilaster strips and round-headed windows.
	for _, sill := range []float64{ground + 2.6, ground + 7.6} {
		for x := k0.X + 2.5; x <= k1.X-2.5; x += 3 {
			g.Carve(Arch{Sill: V(x, sill, k0.Y), HalfWidth: 0.6, Spring: 1.4, Depth: 3, Through: AxisZ},
				Arch{Sill: V(x, sill, k1.Y), HalfWidth: 0.6, Spring: 1.4, Depth: 3, Through: AxisZ})
		}
		for z := k0.Y + 2.5; z <= k1.Y-2.5; z += 2.7 {
			g.Carve(Arch{Sill: V(k0.X, sill, z), HalfWidth: 0.6, Spring: 1.4, Depth: 3, Through: AxisX},
				Arch{Sill: V(k1.X, sill, z), HalfWidth: 0.6, Spring: 1.4, Depth: 3, Through: AxisX})
		}
	}
	// Forebuilding stair with a wooden door.
	g.Fill(Block(V(-97, ground, k0.Y-2.4), V(-93, ground+5, k0.Y)))
	g.With(Wood, func() { g.Fill(Block(V(-96, ground+0.4, k0.Y-2.6), V(-94, ground+2.8, k0.Y-2.4))) })
	for i, corner := range []Vec2{k0, P(k1.X, k0.Y), k1, P(k0.X, k1.Y)} {
		ttop := top + 3.6
		g.Use(White)
		if i == 0 {
			g.Fill(Cylinder{Center: corner, Radius: 2.0, Y0: ground, Y1: ttop})
		} else {
			g.Fill(BlockAt(corner, 3.4, 3.4, ground, ttop))
		}
		// Ogee cupola: a swelling lead bulb drawing up to a point.
		cup := Func{Box: Box{V(corner.X-2.2, ttop, corner.Y-2.2), V(corner.X+2.2, ttop+4.5, corner.Y+2.2)}, F: func(p Vec3) bool {
			h := (p.Y - ttop) / 4.5
			r := 2.0 * math.Sin(math.Pi*math.Pow(h, 0.7)) * (1 - h*0.6)
			if h < 0.18 {
				r = math.Max(r, 1.7)
			}
			return p.Plan().Sub(corner).Length() <= r
		}}
		g.With(Slate, func() { g.FillVolume(cup) })
		g.With(Gold, func() { g.Fill(Capsule{corner.At(ttop + 4.0), corner.At(ttop + 6.0), 0.45}) })
	}
	// Waterloo Block along the north wall and the chapel in the north-west corner.
	g.Use(Sand).Build(kit.Hall{Min: V(-106.2, ground, 10), Max: V(-103.2, ground+4, 20), Ridge: AxisZ, RoofColor: Slate})
	g.Use(Brick).Build(kit.Hall{Min: V(-106.2, ground, 22), Max: V(-103.2, ground+3.5, 27.5), Ridge: AxisZ, RoofColor: Slate})
	for _, at := range []Vec2{P(-84, 26.5), P(-84, 11)} {
		g.Build(kit.Tree{At: at, Base: ground, Height: 3.5})
	}
	// Trees on the rest of the north bank.
	for i, at := range []Vec2{P(-74, 28), P(-76, -12), P(-76, -22), P(-76, -31), P(-88, -32), P(-98, -32), P(-108, -32), P(-90, -6.5), P(-100, -6.5)} {
		g.Build(kit.Tree{At: at, Base: ground, Height: 3.5 + 2*Hash01(i, 31)})
	}
}

// southBank adds riverside trees and brick warehouses (Butler's Wharf) on the south bank.
func (t TowerBridge) southBank(g *Grid) {
	ground := towerBridgeBankTop
	for i, z := range [][2]float64{{-32, -21}, {-19, -8}, {22, 32}} {
		h := 6.5 + 2.5*Hash01(i, 7)
		lo, hi := P(81, z[0]), P(92, z[1])
		color := Brick
		if i%2 == 1 {
			color = Sand
		}
		t.warehouse(g, lo, hi, h, color)
	}
	for i, at := range []Vec2{P(76, -28), P(76, -18), P(76, -10), P(76, 12), P(84, 12), P(90, 16), P(76, 22), P(76, 30)} {
		g.Build(kit.Tree{At: at, Base: ground, Height: 3.5 + 1.5*Hash01(i, 77)})
	}
}

// warehouse is a hollow flat-roofed block with rows of lintelled windows.
func (t TowerBridge) warehouse(g *Grid, lo, hi Vec2, h float64, color Color) {
	ground := towerBridgeBankTop
	g.Use(color)
	g.Fill(Block(lo.At(ground), hi.At(ground+h)))
	g.Carve(Block(V(lo.X+1, ground+0.5, lo.Y+1), V(hi.X-1, ground+h-0.8, hi.Y-1)))
	// Rows of tall windows with stone lintels on every face.
	for y := ground + 1.0; y+2.2 < ground+h; y += 2.6 {
		for zz := lo.Y + 2.0; zz <= hi.Y-1.9; zz += 2.4 {
			for _, x := range []float64{lo.X, hi.X} {
				g.Paint(Stone, BlockAt(P(x, zz), 0.6, 1.8, y+1.6, y+2.0))
				g.Carve(BlockAt(P(x, zz), 2, 1.2, y, y+1.5))
			}
		}
		for xx := lo.X + 2.0; xx <= hi.X-1.9; xx += 2.4 {
			for _, z := range []float64{lo.Y, hi.Y} {
				g.Paint(Stone, BlockAt(P(xx, z), 1.8, 0.6, y+1.6, y+2.0))
				g.Carve(BlockAt(P(xx, z), 1.2, 2, y, y+1.5))
			}
		}
	}
	g.With(Stone, func() {
		g.Fill(Subtract(Block(V(lo.X-0.3, ground+h, lo.Y-0.3), V(hi.X+0.3, ground+h+0.8, hi.Y+0.3)),
			Block(V(lo.X+0.7, ground+h, lo.Y+0.7), V(hi.X-0.7, ground+h+0.8, hi.Y-0.7))))
	})
	g.Paint(Slate, Block(V(lo.X+0.7, ground+h-0.8, lo.Y+0.7), V(hi.X-0.7, ground+h, hi.Y-0.7)))
}

// bus is a red double-decker crossing a side span along X.
func (t TowerBridge) bus(g *Grid, at Vec2) {
	y := towerBridgeDeck
	g.With(Red, func() { g.Fill(BlockAt(at, 6.6, 2.2, y+0.5, y+4.0)) })
	g.Paint(Glass, BlockAt(at, 6.8, 2.4, y+1.7, y+2.3), BlockAt(at, 6.8, 2.4, y+3.0, y+3.5))
	g.Paint(Red, BlockAt(at, 0.8, 2.4, y+1.7, y+3.5))
	g.Paint(White, BlockAt(at, 6.8, 2.4, y+2.6, y+2.7))
	g.With(Charcoal, func() {
		for _, dx := range []float64{-2.2, 2.2} {
			g.Fill(BlockAt(P(at.X+dx, at.Y), 1.2, 2.4, y, y+0.9))
		}
	})
}

// boat is a white river cruiser with a red band, heading downstream (-Z) for the bascules.
func (t TowerBridge) boat(g *Grid, at Vec2) {
	const half, length = 2.4, 8.0
	hull := Func{Box: Box{V(at.X-half, -0.3, at.Y-length), V(at.X+half, 1.8, at.Y+length)}, F: func(p Vec3) bool {
		u := (at.Y - p.Z) / length
		w := half
		if u > 0.35 {
			w = half * math.Sqrt(math.Max(0, 1-(u-0.35)/0.65*((u-0.35)/0.65)))
		}
		if u < -0.85 {
			w = half * 0.9
		}
		return math.Abs(p.X-at.X) <= w
	}}
	g.With(White, func() { g.FillVolume(hull) })
	g.Paint(Navy, Block(V(at.X-half, -0.3, at.Y-length), V(at.X+half, 0.5, at.Y+length)))
	g.Paint(Red, Block(V(at.X-half, 1.0, at.Y-length), V(at.X+half, 1.3, at.Y+length)))
	g.With(White, func() { g.Fill(BlockAt(P(at.X, at.Y+1.5), 3.4, 9, 1.8, 3.6)) })
	g.Paint(Glass, BlockAt(P(at.X, at.Y+1.5), 3.6, 9.2, 2.3, 3.1))
	g.With(Red, func() { g.Fill(BlockAt(P(at.X, at.Y+1.5), 3.8, 9.6, 3.6, 4.0)) })
	g.With(Wood, func() { g.Fill(BlockAt(P(at.X, at.Y-5), 2.6, 3, 1.8, 2.0)) })
}

// car is a small vehicle on the deck along X.
func (t TowerBridge) car(g *Grid, at Vec2, color Color) {
	y := towerBridgeDeck
	g.With(color, func() {
		g.Fill(BlockAt(at, 3.0, 1.6, y+0.4, y+1.3), BlockAt(at, 1.8, 1.6, y+1.3, y+2.0))
	})
	g.Paint(Glass, BlockAt(at, 1.9, 1.7, y+1.4, y+1.8))
	g.With(Charcoal, func() {
		for _, dx := range []float64{-0.9, 0.9} {
			g.Fill(BlockAt(P(at.X+dx, at.Y), 0.8, 1.7, y, y+0.5))
		}
	})
}

// belfast is a hint of HMS Belfast moored upstream off the south bank: a dazzle-grey
// cruiser hull, superstructure, raked funnels, masts and four gun turrets.
func (t TowerBridge) belfast(g *Grid) {
	const cx, cz, half, beam, deck = 58.5, 19.5, 13.0, 2.3, 2.0
	hull := Func{Box: Box{V(cx-beam, -0.3, cz-half), V(cx+beam, deck, cz+half)}, F: func(p Vec3) bool {
		u := (p.Z - cz) / half
		w := beam
		if u > 0.25 {
			w = beam * math.Sqrt(math.Max(0, 1-(u-0.25)/0.75))
		} else if u < -0.7 {
			w = beam * (1 - (-0.7-u)*1.2)
		}
		// The bow rises a little.
		top := deck + math.Max(0, u-0.5)*1.2
		return math.Abs(p.X-cx) <= w && p.Y <= top
	}}
	g.With(DarkStone, func() { g.FillVolume(hull) })
	// Dazzle camouflage: slanted bands of grey tones, a dark boot-top at the waterline.
	g.Paint(Stone, Func{Box: hull.Box, F: func(p Vec3) bool {
		return math.Mod(math.Abs(p.Z-cz+0.8*p.Y), 6) < 2
	}})
	g.Paint(Slate, Func{Box: hull.Box, F: func(p Vec3) bool {
		return math.Mod(math.Abs(p.Z-cz+0.8*p.Y), 6) > 4.4
	}})
	g.Paint(Charcoal, Block(V(cx-beam, -0.3, cz-half), V(cx+beam, 0.45, cz+half)))
	g.Paint(Stone, Block(V(cx-beam, deck-0.3, cz-half), V(cx+beam, deck+1, cz+half)))
	// Superstructure, bridge and funnels.
	g.Use(Stone)
	g.Fill(BlockAt(P(cx, cz+0.5), 3.0, 11, deck, deck+1.6))
	g.Fill(BlockAt(P(cx, cz+4.0), 2.8, 3.2, deck+1.6, deck+3.6))
	g.Paint(Charcoal, BlockAt(P(cx, cz+4.0), 3.0, 3.4, deck+2.9, deck+3.2))
	for _, z := range []float64{cz + 0.8, cz - 2.8} {
		g.With(DarkStone, func() {
			g.Fill(Sheared{Solid: Cylinder{Center: P(cx, z), Radius: 0.95, Y0: deck + 1.6, Y1: deck + 5.0}, Slope: P(0, -0.25), Base: deck + 1.6})
		})
		g.With(Charcoal, func() {
			g.Fill(Cylinder{Center: P(cx, z-0.85), Radius: 0.95, Y0: deck + 5.0, Y1: deck + 5.5})
		})
	}
	g.With(Charcoal, func() {
		g.Fill(Capsule{V(cx, deck+3.6, cz+2.8), V(cx, deck+8.5, cz+2.4), 0.45},
			Capsule{V(cx, deck+1.6, cz-5.0), V(cx, deck+7.0, cz-5.3), 0.45})
	})
	// Turrets A, B forward and X, Y aft, guns trained along the hull.
	for _, z := range []float64{cz + 9.0, cz + 6.6, cz - 7.0, cz - 9.4} {
		dir := 1.0
		if z < cz {
			dir = -1
		}
		lift := 0.0
		if z == cz+6.6 || z == cz-7.0 {
			lift = 0.6 // Superfiring turrets.
		}
		g.With(Slate, func() {
			g.Fill(BlockAt(P(cx, z), 2.0, 2.0, deck, deck+1.2+lift))
		})
		g.With(Charcoal, func() {
			g.Fill(Capsule{V(cx, deck+0.8+lift, z+dir*1.0), V(cx, deck+0.9+lift, z+dir*3.0), 0.45})
		})
	}
}

// dock is a hint of St Katharine Docks downstream on the north bank: a sunken basin
// with quay walls, moored yachts and the sand-coloured Ivory House warehouse.
func (t TowerBridge) dock(g *Grid) {
	lo, hi := P(-102, -27), P(-82, -12)
	ground := towerBridgeBankTop
	g.Carve(Block(lo.At(ground-1), hi.At(ground+0.5)))
	g.With(DarkStone, func() {
		g.Fill(Subtract(Block(V(lo.X-0.8, 0.8, lo.Y-0.8), V(hi.X+0.8, ground, hi.Y+0.8)), Block(lo.At(0.8), hi.At(ground+1))))
	})
	g.Paint(Stone, Subtract(Block(V(lo.X-1.6, ground-0.8, lo.Y-1.6), V(hi.X+1.6, ground, hi.Y+1.6)), Block(lo.At(0), hi.At(ground+1))))
	t.raster(g, Box{lo.At(1.1), hi.At(1.3)}, func(Vec3) (Color, bool) { return Water, true })
	for i, at := range []Vec2{P(-97, -16), P(-92, -23), P(-87, -16), P(-97, -23)} {
		y := 1.2
		hull := Func{Box: Box{V(at.X-2.2, y, at.Y-0.9), V(at.X+2.2, y+1.0, at.Y+0.9)}, F: func(p Vec3) bool {
			u := (p.X - at.X) / 2.2
			return math.Abs(p.Z-at.Y) <= 0.9*math.Sqrt(math.Max(0, 1-math.Max(0, u)*math.Max(0, u)))
		}}
		color := White
		if i == 1 {
			color = Navy
		}
		g.With(color, func() { g.FillVolume(hull) })
		g.With(White, func() {
			g.Fill(Capsule{at.At(y + 1.0), at.At(y + 5.0 + Hash01(i, 5)), 0.4})
		})
	}
	t.warehouse(g, P(-110, -28), P(-104.5, -11), 6.5, Sand)
}
