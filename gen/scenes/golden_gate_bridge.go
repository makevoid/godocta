package scenes

import (
	"math"

	"godocta/gen/kit"
	. "godocta/gen/octa"
)

// GoldenGateBridge is the suspension bridge across the Golden Gate strait. The bridge
// runs along X: San Francisco is +X, Marin County -X; the Pacific lies towards -Z and
// the bay towards +Z. Two international-orange Art Deco towers with stepped legs and
// portal struts carry sagging main cables, vertical suspenders and a Warren-truss deck
// with four lanes of traffic. At the south end the deck crosses the steel arch over
// Fort Point to the Presidio bluff and toll plaza; at the north end it enters a cut in
// the Marin Headlands and the rainbow tunnel. The span is compressed and the towers
// exaggerated so the bridge fills the frame. A container ship passes under the main
// span and sailboats cross the strait.
type GoldenGateBridge struct{}

func init() { Register(GoldenGateBridge{}) }

func (GoldenGateBridge) Name() string  { return "golden_gate_bridge" }
func (GoldenGateBridge) Title() string { return "Golden Gate Bridge" }

const (
	ggbTowerX    = 60.0  // Tower centers at ±X.
	ggbAnchorX   = 86.0  // Anchorage centers (side span ends) at ±X.
	ggbDeck      = 15.5  // Road surface; the strait is y = 0.
	ggbSlab      = 14.8  // Underside of the road slab.
	ggbTrussBot  = 12.0  // Bottom chord of the stiffening truss.
	ggbRoad      = 5.0   // Road half width; truss inner face.
	ggbTrussOut  = 6.4   // Truss outer face.
	ggbCableZ    = 6.0   // Main cable plane.
	ggbCableR    = 0.8   // Main cable radius.
	ggbCableTop  = 56.2  // Cable center over the saddles.
	ggbCableLow  = 17.6  // Cable center at mid-span.
	ggbAnchorY   = 17.0  // Cable center where it enters the anchorages.
	ggbLegZ      = 6.75  // Tower leg centers at ±Z.
	ggbHalfZ     = 28.0  // Site extent along Z.
	ggbArchX     = 99.0  // Fort Point arch crown.
	ggbArchHalf  = 8.0   // Fort Point arch half span.
	ggbMarinEnd  = -96.0 // Tunnel portal face.
	ggbSiteMinX  = -124.0
	ggbSiteMaxX  = 134.0
	ggbBridgeEnd = 112.0 // South end of the truss deck, where the bluff takes over.
)

func (b GoldenGateBridge) Build(g *Grid) Order {
	b.terrain(g)
	b.water(g)
	b.fortPoint(g)
	b.deck(g)
	for _, side := range []float64{-1, 1} {
		b.tower(g, side*ggbTowerX, side > 0)
		b.anchorage(g, side)
	}
	b.cables(g)
	b.arch(g)
	b.pylons(g, ggbBridgeEnd-3)
	b.landRoad(g)
	b.tunnel(g)
	b.plaza(g)
	b.trees(g)
	b.traffic(g)
	b.ship(g, P(22, 6))
	b.sailboat(g, P(-28, 17), White, 1)
	b.sailboat(g, P(36, -16), Red, -1)
	// Crews work out from both shores towards mid-span.
	return func(p Vec3) float64 { return -math.Abs(p.X) + 0.01*p.Z }
}

// put sets every cell whose center lies inside a solid, without half-unit snapping,
// so slender members and odd-sized blocks stay symmetric.
func (b GoldenGateBridge) put(g *Grid, color Color, solids ...Solid) {
	for _, s := range solids {
		b.raster(g, s.Bounds(), func(p Vec3) (Color, bool) { return color, s.Contains(p) })
	}
}

// cut removes every cell whose center lies inside a solid, without snapping.
func (b GoldenGateBridge) cut(g *Grid, solids ...Solid) {
	for _, s := range solids {
		b.each(g, s.Bounds(), func(c Cell, p Vec3) {
			if s.Contains(p) {
				g.Remove(c)
			}
		})
	}
}

func (GoldenGateBridge) each(g *Grid, b Box, f func(Cell, Vec3)) {
	s := g.Span()
	lo := func(v float64) int32 { return int32(math.Ceil(v * 2 / s)) }
	hi := func(v float64) int32 { return int32(math.Floor(v * 2 / s)) }
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
				f(c, g.World(c))
			}
		}
	}
}

// raster sets every cell in the box for which f reports a color.
func (b GoldenGateBridge) raster(g *Grid, box Box, f func(Vec3) (Color, bool)) {
	previous := g.Brush()
	b.each(g, box, func(c Cell, p Vec3) {
		if color, ok := f(p); ok {
			g.Remove(c)
			g.Use(color).Add(c)
		}
	})
	g.Use(previous)
}

// line draws a connected one-cell line between two points.
func (GoldenGateBridge) line(g *Grid, a, c Vec3, color Color) {
	previous := g.Brush()
	g.Use(color)
	d := c.Sub(a)
	steps := int(math.Ceil(d.Length()/(g.Span()*0.25))) + 1
	for i := 0; i <= steps; i++ {
		cell := Nearest(a.Add(d.Scale(float64(i) / float64(steps))).Scale(1 / g.Span()))
		g.Remove(cell)
		g.Add(cell)
	}
	g.Use(previous)
}

// noise is smooth value noise in [0, 1) over a unit integer lattice.
func (GoldenGateBridge) noise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	fx, fz = fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	h := func(dx, dz int) float64 { return Hash01(int(ix)+dx, int(iz)+dz, seed) }
	top := h(0, 0) + (h(1, 0)-h(0, 0))*fx
	bottom := h(0, 1) + (h(1, 1)-h(0, 1))*fx
	return top + (bottom-top)*fz
}

// height is the terrain surface: the Marin Headlands (-X), the Fort Point rocks and
// the Presidio bluff (+X). Zero is open water.
func (b GoldenGateBridge) height(x, z float64) float64 {
	h := 0.0
	// Marin Headlands: a steep green-brown hill whose peak lies west of the road.
	{
		const px, pz, peak = -105.0, -7.0, 33.0
		dx := (x - px) / 37
		if x < px {
			dx = (x - px) / 19
		}
		dz := (z - pz) / 35
		if z < pz {
			dz = (z - pz) / 21
		}
		if r := math.Hypot(dx, dz); r < 1 {
			s := 1 - r
			prof := 0.45*s + 0.55*s*s*(3-2*s)
			n := b.noise(x/8, z/8, 3) - 0.5
			h = math.Max(h, peak*prof+7*n*s)
		}
	}
	// Presidio bluff: a plateau carrying the toll plaza.
	{
		const px, pz, top = 120.0, 2.0, 14.6
		dx := (x - px) / 12
		if x > px {
			dx = (x - px) / 14
		}
		dz := (z - pz) / 26
		if z < pz {
			dz = (z - pz) / 30
		}
		if r2 := dx*dx + dz*dz; r2 < 1 {
			v := math.Min(1, 1.45*(1-r2))
			n := b.noise(x/6, z/6, 5) - 0.5
			h = math.Max(h, top*v+3*n*v*(1-v)*4)
		}
	}
	// Fort Point: a low rocky shelf under the arch.
	{
		ax, az := math.Abs(x-98)/12, math.Abs(z)/13
		r := math.Pow(math.Pow(ax, 4)+math.Pow(az, 4), 0.25)
		if r < 1 {
			h = math.Max(h, 1.6*math.Min(1, (1-r)*6))
		}
	}
	return h
}

// inCut reports whether a plan point lies in the Marin road cut.
func (GoldenGateBridge) inCut(x, z float64) bool {
	return x < -ggbAnchorX+3 && x > ggbMarinEnd-3 && math.Abs(z) < 7.5
}

// terrain rasterizes the height field as a shell, thickened on steep slopes so the
// surface stays closed, and solid down to the road in the Marin cut.
func (b GoldenGateBridge) terrain(g *Grid) {
	type column struct {
		h, bottom float64
		color     Color
	}
	cache := map[[2]int32]column{}
	get := func(c Cell, p Vec3) column {
		key := [2]int32{c.X, c.Z}
		if col, ok := cache[key]; ok {
			return col
		}
		h := b.height(p.X, p.Z)
		const e = 0.5
		slope := math.Hypot(b.height(p.X+e, p.Z)-b.height(p.X-e, p.Z), b.height(p.X, p.Z+e)-b.height(p.X, p.Z-e)) / (2 * e)
		bottom := h - 1.3 - 0.9*slope
		cut := b.inCut(p.X, p.Z)
		if math.Abs(p.Z) < 11 && p.X < -ggbAnchorX+4 && p.X > ggbMarinEnd-6 {
			bottom = math.Min(bottom, ggbSlab-1.5)
		}
		if cut {
			h = math.Min(h, ggbSlab-0.2)
		}
		n := b.noise(p.X/5, p.Z/5, 11)
		fine := Hash01(int(c.X), int(c.Z), 17)
		var color Color
		switch {
		case h < 1.0 && p.X > 80:
			color = DarkStone
			if fine < 0.35 {
				color = Stone
			}
		case h < 1.0:
			color = Sand
			if n > 0.62 {
				color = DarkStone
			}
		case p.X > 112 && slope < 3.2:
			color = Foliage
			if n < 0.45 {
				color = Grass
			} else if fine < 0.15 {
				color = DarkStone
			}
		case slope > 1.5 || (cut && math.Abs(p.Z) > 6.6):
			color = DarkStone
			switch {
			case p.X > 0 && n > 0.55:
				color = Grass
			case n > 0.6:
				color = Brick
			case fine < 0.25:
				color = Stone
			}
		case p.X > 0:
			color = Grass
			if n > 0.55 {
				color = Foliage
			}
		default:
			switch {
			case n < 0.38:
				color = Sand
			case n < 0.7:
				color = Grass
			default:
				color = Foliage
			}
			if fine < 0.12 {
				color = Foliage
			}
		}
		col := column{h, math.Max(bottom, 0.15), color}
		cache[key] = col
		return col
	}
	box := Box{V(ggbSiteMinX, 0.1, -ggbHalfZ), V(ggbSiteMaxX, 40, ggbHalfZ)}
	b.raster(g, box, func(p Vec3) (Color, bool) {
		c := Nearest(p.Scale(1 / g.Span()))
		col := get(c, p)
		if col.h < 0.5 || p.Y > col.h || p.Y < col.bottom {
			return 0, false
		}
		return col.color, true
	})
}

// water is a single layer over the strait wherever the land is lower.
func (b GoldenGateBridge) water(g *Grid) {
	box := Box{V(ggbSiteMinX, -0.2, -ggbHalfZ), V(ggbSiteMaxX, 0.2, ggbHalfZ)}
	b.raster(g, box, func(p Vec3) (Color, bool) {
		if b.height(p.X, p.Z) >= 0.5 {
			return 0, false
		}
		return Water, true
	})
	// The container ship's wake.
	b.raster(g, Box{V(10, -0.2, 20), V(34, 0.2, ggbHalfZ)}, func(p Vec3) (Color, bool) {
		d := p.Z - 20
		w := math.Abs(math.Abs(p.X-22) - 3 - d*0.55)
		if w < 0.5+d*0.04 && Hash01(int(p.X*3), int(p.Z*3), 23) < 0.8 {
			return White, true
		}
		return 0, false
	})
}

// cableY is the height of the main cable's centerline at x.
func (GoldenGateBridge) cableY(x float64) float64 {
	ax := math.Abs(x)
	if ax <= ggbTowerX {
		u := ax / ggbTowerX
		return ggbCableLow + (ggbCableTop-ggbCableLow)*u*u
	}
	s := (ax - ggbTowerX) / (ggbAnchorX - ggbTowerX)
	return ggbCableTop + (ggbAnchorY-ggbCableTop)*s - 2.5*4*s*(1-s)
}

// deck is the road slab between two Warren stiffening trusses, from the Marin
// anchorage to the Presidio bluff, with lane markings and a railing.
func (b GoldenGateBridge) deck(g *Grid) {
	x0, x1 := -ggbAnchorX, ggbBridgeEnd
	const panel = 4.0
	for _, sz := range []float64{-1, 1} {
		zlo, zhi := ggbRoad, ggbTrussOut
		if sz < 0 {
			zlo, zhi = -ggbTrussOut, -ggbRoad
		}
		b.raster(g, Box{V(x0, ggbTrussBot, zlo), V(x1, ggbDeck, zhi)}, func(p Vec3) (Color, bool) {
			y := p.Y - ggbTrussBot
			depth := ggbDeck - ggbTrussBot
			if y < 0.6 || y > depth-0.65 {
				return Red, true
			}
			u := math.Mod(p.X-x0, panel)
			k := int(math.Floor((p.X - x0) / panel))
			v := y / depth * panel
			if k%2 == 1 {
				v = panel - v
			}
			if math.Abs(u-v) < 0.4 || u < 0.3 || u > panel-0.3 {
				return Red, true
			}
			return 0, false
		})
		// Railing on the outer chord, with lamp posts.
		rz := sz * (ggbTrussOut - 0.4)
		b.raster(g, Box{V(x0, ggbDeck, rz-0.4), V(x1, ggbDeck+0.9, rz+0.4)}, func(p Vec3) (Color, bool) {
			if p.Y > ggbDeck+0.6 || math.Mod(p.X-x0, 1.6) < 0.5 {
				return Red, true
			}
			return 0, false
		})
		for x := x0 + 4; x < x1; x += 8 {
			if math.Abs(math.Abs(x)-ggbTowerX) < 4 {
				continue
			}
			lz := sz * (ggbRoad + 0.3)
			b.line(g, V(x, ggbDeck, lz), V(x, ggbDeck+3.2, lz), Stone)
			b.line(g, V(x, ggbDeck+3.2, lz), V(x, ggbDeck+3.2, lz-sz*0.8), Gold)
		}
	}
	// Floor beams under the slab and the road slab with markings.
	b.raster(g, Box{V(x0, ggbSlab, -ggbRoad), V(x1, ggbDeck, ggbRoad)}, func(p Vec3) (Color, bool) {
		if p.Y < ggbDeck-g.Span()*0.5 {
			return Charcoal, true
		}
		dash := math.Mod(p.X+200, 4) < 2
		switch {
		case math.Abs(p.Z) < 0.45:
			return Gold, true
		case math.Abs(math.Abs(p.Z)-2.5) < 0.4 && dash:
			return White, true
		}
		return Charcoal, true
	})
	for x := x0 + 1.6; x < x1; x += panel {
		b.put(g, Red, Block(V(x-0.4, ggbTrussBot, -ggbRoad), V(x+0.4, ggbTrussBot+0.7, ggbRoad)))
	}
}

// tower builds one Art Deco tower: a pier, two stepped legs with fluted faces, a
// braced portal below the deck, four stepped portal struts above it and saddles.
func (b GoldenGateBridge) tower(g *Grid, cx float64, fender bool) {
	// Pier and, for the south tower, the fender ring.
	pier := Union(Block(V(cx-4.5, -0.6, -9), V(cx+4.5, 3, 9)),
		Cylinder{Center: P(cx, -9), Radius: 4.5, Y0: -0.6, Y1: 3},
		Cylinder{Center: P(cx, 9), Radius: 4.5, Y0: -0.6, Y1: 3})
	b.put(g, Stone, pier)
	b.put(g, DarkStone, Subtract(pier, Block(V(cx-10, 0.6, -20), V(cx+10, 5, 20))))
	if fender {
		ring := Subtract(EllipticCylinder{Center: P(cx, 0), A: 10.5, B: 17.5, Y0: -0.6, Y1: 2.2},
			EllipticCylinder{Center: P(cx, 0), A: 9.3, B: 16.3, Y0: -1, Y1: 3})
		b.put(g, Stone, ring)
	}
	type segment struct{ y0, y1, hx, hz float64 }
	segs := []segment{{3, 24.5, 2.5, 1.7}, {24.5, 34.5, 2.2, 1.5}, {34.5, 43.5, 1.9, 1.35}, {43.5, 55, 1.6, 1.2}}
	inner := func(s segment) float64 { return ggbLegZ - s.hz }
	for _, sz := range []float64{-1, 1} {
		lz := sz * ggbLegZ
		for i, s := range segs {
			b.put(g, Red, Block(V(cx-s.hx, s.y0, lz-s.hz), V(cx+s.hx, s.y1, lz+s.hz)))
			// Setback ledges in a darker tone mark each step.
			if i > 0 {
				b.put(g, Brick, Block(V(cx-segs[i-1].hx, s.y0-0.4, lz-segs[i-1].hz), V(cx+segs[i-1].hx, s.y0, lz+segs[i-1].hz)))
			}
		}
		// Stepped leg cap.
		b.put(g, Red, Block(V(cx-1.2, 55, lz-0.9), V(cx+1.2, 56, lz+0.9)))
		// Fluting: recessed vertical channels on the faces seen along and across the strait.
		for _, s := range segs {
			for _, sx := range []float64{-1, 1} {
				for _, off := range []float64{-s.hz * 0.45, s.hz * 0.45} {
					b.cut(g, Block(V(cx+sx*(s.hx-0.6), s.y0+1.2, lz+off-0.3), V(cx+sx*(s.hx+0.2), s.y1-1.0, lz+off+0.3)))
				}
			}
			b.cut(g, Block(V(cx-s.hx*0.35, s.y0+1.2, lz+sz*(s.hz-0.6)), V(cx+s.hx*0.35, s.y1-1.0, lz+sz*(s.hz+0.2))))
		}
	}
	// Portal struts between the legs, with stepped Art Deco brackets.
	strut := func(y0, y1, half float64, s segment) {
		zin := inner(s)
		b.put(g, Red, Block(V(cx-half, y0, -zin), V(cx+half, y1, zin)))
		for _, sz := range []float64{-1, 1} {
			for k, step := range []float64{2.2, 1.4, 0.7} {
				z0, z1 := sz*zin, sz*(zin-step)
				yb := y0 - float64(k+1)*0.8
				b.put(g, Red, Block(V(cx-half, yb, math.Min(z0, z1)), V(cx+half, y0, math.Max(z0, z1))))
			}
		}
		// Vertical ribs across the strut faces.
		for z := -zin + 1.2; z < zin-1; z += 1.6 {
			for _, sx := range []float64{-1, 1} {
				b.put(g, Brick, Block(V(cx+sx*half-0.3, y0+0.4, z-0.3), V(cx+sx*half+0.3, y1-0.4, z+0.3)))
			}
		}
	}
	strut(9.5, ggbTrussBot, 1.6, segs[0])
	strut(27.5, 30, 1.4, segs[1])
	strut(38.5, 40.5, 1.3, segs[2])
	strut(48, 50, 1.2, segs[3])
	// The top portal: a deep strut with a stepped crown between the leg caps.
	zin := inner(segs[3])
	b.put(g, Red, Block(V(cx-1.3, 52, -zin), V(cx+1.3, 55, zin)))
	b.put(g, Red, Block(V(cx-1.0, 55, -zin+1.5), V(cx+1.0, 55.8, zin-1.5)))
	for z := -zin + 1.0; z < zin-0.5; z += 1.5 {
		b.cut(g, Block(V(cx-2, 52.6, z-0.35), V(cx+2, 54.4, z+0.35)))
	}
	b.put(g, Red, Block(V(cx-0.5, 52, -zin), V(cx+0.5, 55, zin)))
	// X bracing below the deck.
	for _, sz := range []float64{-1, 1} {
		b.put(g, Red, Capsule{V(cx, 3.5, -sz*(ggbLegZ-1.7)), V(cx, 8.8, sz*(ggbLegZ-1.7)), 0.55})
	}
	b.put(g, Red, Block(V(cx-1.4, 3, -ggbLegZ), V(cx+1.4, 4, ggbLegZ)))
	// Cable saddle housings.
	for _, sz := range []float64{-1, 1} {
		b.put(g, Red, Block(V(cx-1.6, 55.5, sz*ggbCableZ-1.0), V(cx+1.6, 57.2, sz*ggbCableZ+1.0)))
	}
}

// anchorage is a concrete block where the cables end, with a pair of stepped pylons.
func (b GoldenGateBridge) anchorage(g *Grid, side float64) {
	cx := side * ggbAnchorX
	base := 0.0
	if side < 0 {
		base = 6
	}
	b.put(g, Stone, Block(V(cx-3.5, base, -9), V(cx+3.5, ggbTrussBot, 9)))
	b.put(g, DarkStone, Block(V(cx-3.6, ggbTrussBot-0.8, -9.1), V(cx+3.6, ggbTrussBot, 9.1)))
	// Recessed Art Deco panels on the long faces and the face towards the tower.
	for _, sz := range []float64{-1, 1} {
		for _, x := range []float64{cx - 2, cx, cx + 2} {
			b.cut(g, Block(V(x-0.4, base+1.5, sz*9-0.7), V(x+0.4, ggbTrussBot-1.6, sz*9+0.7)))
		}
	}
	for z := -6.0; z <= 6; z += 2 {
		b.cut(g, Block(V(cx-side*3.5-0.7, base+1.5, z-0.4), V(cx-side*3.5+0.7, ggbTrussBot-1.6, z+0.4)))
	}
	// Cable housings rising to the deck.
	for _, sz := range []float64{-1, 1} {
		b.put(g, Stone, Block(V(cx-3.0, ggbTrussBot, sz*ggbCableZ-1.3), V(cx+3.0, ggbAnchorY+0.6, sz*ggbCableZ+1.3)))
	}
	b.pylons(g, cx)
}

// pylons are twin Art Deco concrete columns flanking the deck.
func (b GoldenGateBridge) pylons(g *Grid, cx float64) {
	for _, sz := range []float64{-1, 1} {
		z := sz * 8.6
		base := math.Max(0, b.height(cx, z)-1)
		if cx < 0 {
			base = math.Max(base, 6)
		}
		b.put(g, Stone, Block(V(cx-1.8, base, z-1.6), V(cx+1.8, 21, z+1.6)))
		b.put(g, Stone, Block(V(cx-1.4, 21, z-1.2), V(cx+1.4, 22.4, z+1.2)))
		b.put(g, Stone, Block(V(cx-1.0, 22.4, z-0.8), V(cx+1.0, 23.4, z+0.8)))
		b.put(g, DarkStone, Block(V(cx-1.9, 20.2, z-1.7), V(cx+1.9, 20.8, z+1.7)))
		// Vertical channels on every face.
		for _, sx := range []float64{-1, 1} {
			b.cut(g, Block(V(cx+sx*1.3, ggbDeck+1, z-0.35), V(cx+sx*2.0, 19.6, z+0.35)))
			b.cut(g, Block(V(cx-0.35, ggbDeck+1, z+sx*1.1), V(cx+0.35, 19.6, z+sx*1.8)))
		}
	}
}

// cables draws the two main cables and their vertical suspenders.
func (b GoldenGateBridge) cables(g *Grid) {
	for _, sz := range []float64{-1, 1} {
		z := sz * ggbCableZ
		var prev Vec3
		for i := 0; ; i++ {
			x := -ggbAnchorX + float64(i)*0.6
			if x > ggbAnchorX {
				break
			}
			p := V(x, b.cableY(x), z)
			if i > 0 {
				b.put(g, Red, Capsule{prev, p, ggbCableR})
			}
			prev = p
		}
		for x := -ggbAnchorX + 4; x < ggbAnchorX-2; x += 4 {
			if math.Abs(math.Abs(x)-ggbTowerX) < 2.5 {
				continue
			}
			top := b.cableY(x) - ggbCableR + 0.2
			if top-ggbDeck < 0.8 {
				continue
			}
			b.line(g, V(x, ggbDeck, z), V(x, top, z), Red)
		}
	}
}

// arch is the steel arch that carries the deck over Fort Point.
func (b GoldenGateBridge) arch(g *Grid) {
	archY := func(x float64) float64 {
		u := (x - ggbArchX) / ggbArchHalf
		return 1.8 + (ggbTrussBot-2.2-1.8)*(1-u*u)
	}
	for _, sz := range []float64{-1, 1} {
		zlo, zhi := sz*ggbRoad, sz*ggbTrussOut
		if sz < 0 {
			zlo, zhi = zhi, zlo
		}
		b.raster(g, Box{V(ggbArchX-ggbArchHalf-0.5, 1.5, zlo), V(ggbArchX+ggbArchHalf+0.5, ggbTrussBot, zhi)}, func(p Vec3) (Color, bool) {
			y := archY(p.X)
			if math.Abs(p.X-ggbArchX) <= ggbArchHalf && p.Y >= y && p.Y <= y+1.3 {
				return Red, true
			}
			if p.Y > y && math.Mod(p.X-ggbArchX+40, 2.0) < 0.6 {
				return Red, true
			}
			return 0, false
		})
	}
	for _, x := range []float64{ggbArchX - ggbArchHalf - 1.5, ggbArchX + ggbArchHalf + 1.5} {
		b.put(g, Stone, Block(V(x-2, 0, -7.5), V(x+2, 3, 7.5)))
	}
}

// fortPoint is the brick Civil War fort beneath the arch: a casemated rectangle with
// an open parade ground, a seawall and a small lighthouse.
func (b GoldenGateBridge) fortPoint(g *Grid) {
	const x0, x1, z0, z1, base, top = 93.5, 104.5, -8.0, 8.0, 1.4, 7.6
	b.put(g, Brick, Block(V(x0, base, z0), V(x1, top, z1)))
	b.cut(g, Block(V(x0+1.6, base+0.6, z0+1.6), V(x1-1.6, top+1, z1-1.6)))
	// Casemate tiers on the outer faces.
	for _, sill := range []float64{base + 0.9, base + 3.5} {
		for z := z0 + 2.0; z <= z1-1.9; z += 2.0 {
			for _, x := range []float64{x0, x1} {
				b.cut(g, Arch{V(x, sill, z), 0.5, 0.9, 1.4, AxisX})
			}
		}
		for x := x0 + 2.0; x <= x1-1.9; x += 2.0 {
			for _, z := range []float64{z0, z1} {
				b.cut(g, Arch{V(x, sill, z), 0.5, 0.9, 1.4, AxisZ})
			}
		}
	}
	// Barbette tier band and parapet.
	b.put(g, Sand, Subtract(Block(V(x0, top-0.5, z0), V(x1, top, z1)), Block(V(x0+1, top-1, z0+1), V(x1-1, top+1, z1-1))))
	// Seawall.
	b.put(g, Stone, Subtract(Block(V(x0-2.5, 0, z0-2.5), V(x1+1.5, 2.0, z1+2.5)), Block(V(x0-1.5, 0.5, z0-1.5), V(x1+0.5, 3, z1+1.5))))
	// Lighthouse on the barbette corner.
	b.put(g, White, Cylinder{Center: P(x0+1.2, z1-1.2), Radius: 0.9, Y0: top, Y1: top + 2.6})
	b.put(g, Charcoal, Cylinder{Center: P(x0+1.2, z1-1.2), Radius: 0.6, Y0: top + 2.6, Y1: top + 3.3})
}

// landRoad continues the roadway on the Presidio plateau and in the Marin cut.
func (b GoldenGateBridge) landRoad(g *Grid) {
	road := func(p Vec3) (Color, bool) {
		if p.Y < ggbDeck-g.Span()*0.5 {
			return DarkStone, true
		}
		switch {
		case math.Abs(p.Z) < 0.45:
			return Gold, true
		case math.Abs(math.Abs(p.Z)-2.5) < 0.4 && math.Mod(p.X+200, 4) < 2:
			return White, true
		case math.Abs(p.Z) > ggbRoad+0.3:
			return Stone, true
		}
		return Charcoal, true
	}
	b.raster(g, Box{V(ggbMarinEnd-2, ggbSlab, -ggbRoad-1), V(-ggbAnchorX-3.5, ggbDeck, ggbRoad+1)}, road)
	b.raster(g, Box{V(ggbBridgeEnd, ggbSlab, -ggbRoad-1), V(ggbSiteMaxX, ggbDeck, ggbRoad+1)}, func(p Vec3) (Color, bool) {
		if b.height(p.X, p.Z) < 13.6 {
			return 0, false
		}
		return road(p)
	})
}

// tunnel is the Marin portal with its painted rainbow arch.
func (b GoldenGateBridge) tunnel(g *Grid) {
	x := ggbMarinEnd
	b.put(g, Stone, Block(V(x-2.5, ggbSlab-1, -8), V(x, ggbDeck+9.5, 8)))
	bands := []Color{Red, Terracotta, Gold, Grass, Water, Navy}
	for i, c := range bands {
		r := 5.9 - float64(i)*0.3
		b.put(g, c, Arch{V(x+0.2, ggbDeck, 0), r, 3.4, 0.9, AxisX})
	}
	b.cut(g, Arch{V(x, ggbDeck, 0), 4.1, 3.4, 6, AxisX})
	b.put(g, Charcoal, Block(V(x-4, ggbSlab, -4.5), V(x-3, ggbDeck+8, 4.5)))
}

// plaza is the toll plaza canopy and booths on the Presidio side.
func (b GoldenGateBridge) plaza(g *Grid) {
	const x0, x1 = 118.0, 122.0
	for _, z := range []float64{-5.0, -2.5, 0, 2.5, 5.0} {
		b.put(g, White, Block(V(x0+1, ggbDeck, z-0.4), V(x1-1, ggbDeck+2.4, z+0.4)))
	}
	b.put(g, Red, Block(V(x0, ggbDeck+2.4, -6.5), V(x1, ggbDeck+3.0, 6.5)))
	b.put(g, White, Block(V(x0+0.4, ggbDeck+3.0, -6.0), V(x1-0.4, ggbDeck+3.4, 6.0)))
}

// trees scatter cypresses and scrub over the Presidio and the headlands.
func (b GoldenGateBridge) trees(g *Grid) {
	for i := 0; i < 70; i++ {
		x := ggbSiteMinX + Hash01(i, 41)*(ggbSiteMaxX-ggbSiteMinX)
		z := -ggbHalfZ + 2 + Hash01(i, 43)*(2*ggbHalfZ-4)
		h := b.height(x, z)
		if h < 3 || math.Abs(z) < 11 || (x > 90 && x < 108) {
			continue
		}
		e := 0.6
		slope := math.Hypot(b.height(x+e, z)-b.height(x-e, z), b.height(x, z+e)-b.height(x, z-e)) / (2 * e)
		if slope > 1.2 {
			continue
		}
		if x < 0 && Hash01(i, 47) < 0.6 {
			// Marin: low coastal scrub.
			b.put(g, Foliage, Ellipsoid{V(x, h+0.3, z), V(1.6, 1.0, 1.6)})
			continue
		}
		g.Build(kit.Tree{At: P(x, z), Base: h - 0.5, Height: 3 + 2*Hash01(i, 53)})
	}
	// The Presidio's cypress and eucalyptus groves on the bluff.
	for i := 0; i < 60; i++ {
		x := 106 + Hash01(i, 91)*(ggbSiteMaxX-108)
		z := -ggbHalfZ + 2 + Hash01(i, 93)*(2*ggbHalfZ-4)
		h := b.height(x, z)
		if h < 4 || math.Abs(z) < 8.5 {
			continue
		}
		e := 0.6
		slope := math.Hypot(b.height(x+e, z)-b.height(x-e, z), b.height(x, z+e)-b.height(x, z-e)) / (2 * e)
		if slope > 2.2 {
			continue
		}
		g.Build(kit.Tree{At: P(x, z), Base: h - 0.8, Height: 3.5 + 2.5*Hash01(i, 97)})
	}
}

// traffic places cars and trucks in four lanes across the whole crossing.
func (b GoldenGateBridge) traffic(g *Grid) {
	colors := []Color{Red, White, Navy, Gold, Charcoal, Slate, Glass, Terracotta, White, Grass}
	lanes := []float64{-3.75, -1.25, 1.25, 3.75}
	for li, z := range lanes {
		x := -ggbAnchorX + 2 + 6*Hash01(li, 61)
		for k := 0; x < ggbBridgeEnd+4; k++ {
			c := colors[int(Hash01(li, k, 67)*float64(len(colors)))]
			length := 3.0
			truck := Hash01(li, k, 71) < 0.12
			if truck {
				length = 5.5
			}
			body := Block(V(x, ggbDeck, z-0.8), V(x+length, ggbDeck+0.9, z+0.8))
			b.put(g, c, body)
			if truck {
				b.put(g, White, Block(V(x, ggbDeck+0.9, z-0.8), V(x+length-1.4, ggbDeck+2.2, z+0.8)))
			} else {
				b.put(g, Glass, Block(V(x+0.8, ggbDeck+0.9, z-0.7), V(x+length-0.8, ggbDeck+1.5, z+0.7)))
			}
			x += length + 4 + 9*Hash01(li, k, 73)
			if math.Abs(math.Abs(x)-ggbTowerX) < 3 {
				x += 4
			}
		}
	}
}

// ship is a container ship steaming along -Z under the main span.
func (b GoldenGateBridge) ship(g *Grid, c Vec2) {
	const half, length = 3.2, 30.0
	bow, stern := c.Y-length/2, c.Y+length/2
	hull := Prism{Polygon: []Vec2{
		P(c.X, bow), P(c.X+half, bow+5), P(c.X+half, stern-1), P(c.X+half-1, stern),
		P(c.X-half+1, stern), P(c.X-half, stern-1), P(c.X-half, bow+5),
	}, Y0: -0.4, Y1: 3}
	b.put(g, Navy, hull)
	b.put(g, Red, Intersect(hull, Block(V(c.X-5, -0.4, bow-1), V(c.X+5, 0.5, stern+1))))
	b.cut(g, Intersect(hull, Block(V(c.X-half+1, 0.5, bow+4), V(c.X+half-1, 2.3, stern-1.5))))
	b.put(g, Charcoal, Intersect(hull, Block(V(c.X-half+0.6, 2.4, bow+3), V(c.X+half-0.6, 3, stern-1))))
	// Container stacks.
	colors := []Color{Red, Navy, Terracotta, Gold, Slate, White, Grass, Glass, Brick}
	bay := 0
	for z := bow + 5; z < stern-8; z += 3.3 {
		for row := 0; row < 4; row++ {
			x := c.X - 2.25 + float64(row)*1.5
			tiers := 2 + int(Hash01(bay, row, 81)*2.2)
			for t := 0; t < tiers; t++ {
				col := colors[int(Hash01(bay, row, t, 83)*float64(len(colors)))]
				b.put(g, col, Block(V(x-0.7, 3+float64(t)*1.2, z), V(x+0.7, 3+float64(t+1)*1.2-0.05, z+3.1)))
			}
		}
		bay++
	}
	// Bridge house, wings and funnel at the stern.
	b.put(g, White, Block(V(c.X-2.6, 3, stern-6.5), V(c.X+2.6, 9, stern-2.5)))
	b.put(g, Glass, Block(V(c.X-2.7, 7.6, stern-6.6), V(c.X+2.7, 8.4, stern-2.4)))
	b.put(g, White, Block(V(c.X-half-0.6, 8.4, stern-6.5), V(c.X+half+0.6, 9, stern-5.3)))
	b.put(g, Charcoal, Block(V(c.X-0.9, 9, stern-3.6), V(c.X+0.9, 10.6, stern-2.2)))
	b.put(g, Red, Block(V(c.X-0.9, 10.6, stern-3.6), V(c.X+0.9, 11.2, stern-2.2)))
}

// sailboat is a sloop heading along ±X with a mainsail and jib.
func (b GoldenGateBridge) sailboat(g *Grid, c Vec2, sail Color, dir float64) {
	hull := Prism{Polygon: []Vec2{
		P(c.X+dir*4, c.Y), P(c.X+dir*1.5, c.Y-1.1), P(c.X-dir*3, c.Y-1.0), P(c.X-dir*3, c.Y+1.0), P(c.X+dir*1.5, c.Y+1.1),
	}, Y0: -0.2, Y1: 1.0}
	b.put(g, White, hull)
	b.put(g, Navy, Intersect(hull, Block(V(c.X-5, 0.6, c.Y-2), V(c.X+5, 1.0, c.Y+2))))
	mast := c.X + dir*0.6
	b.line(g, V(mast, 1, c.Y), V(mast, 10.5, c.Y), Wood)
	// Mainsail aft of the mast, jib forward, each a thin triangle in the X-Y plane.
	tri := func(a, bb, cc Vec2) Solid {
		lo := V(math.Min(a.X, math.Min(bb.X, cc.X)), math.Min(a.Y, math.Min(bb.Y, cc.Y)), c.Y-0.45)
		hi := V(math.Max(a.X, math.Max(bb.X, cc.X)), math.Max(a.Y, math.Max(bb.Y, cc.Y)), c.Y+0.45)
		return Func{Box{lo, hi}, func(p Vec3) bool {
			return InsidePolygon([]Vec2{a, bb, cc}, P(p.X, p.Y))
		}}
	}
	b.put(g, sail, tri(P(mast-dir*0.5, 1.8), P(mast-dir*0.5, 10.2), P(mast-dir*3.6, 1.8)))
	b.put(g, White, tri(P(mast+dir*0.5, 2.0), P(mast+dir*0.5, 9.4), P(mast+dir*3.4, 1.5)))
}
