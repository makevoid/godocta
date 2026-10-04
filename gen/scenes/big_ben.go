package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// BigBen is the Elizabeth Tower and the Palace of Westminster on the Thames. The river
// flows along X in front of the palace (+Z); the clock tower stands at the palace's
// north end (+X) beside Westminster Bridge, which crosses the river along Z. The
// Victoria Tower closes the river front at the far end (-X) and the octagonal Central
// Tower rises between the ranges. Westminster Hall, Parliament Square and the far bank
// complete the site; buses and cabs cross the green bridge and a cruise boat passes.
type BigBen struct{}

func init() { Register(BigBen{}) }

func (BigBen) Name() string  { return "big_ben" }
func (BigBen) Title() string { return "Big Ben" }

const (
	bigBenLand   = 2.4  // Ground level on both banks; the river surface is y = 0.
	bigBenBankZ  = 13.0 // Embankment wall of the palace bank.
	bigBenFarZ   = 39.0 // Embankment wall of the far bank.
	bigBenX0     = -66.0
	bigBenX1     = 58.0
	bigBenZ0     = -22.0
	bigBenZ1     = 45.0
	bigBenDeck   = 5.0  // Bridge road surface.
	bigBenBridge = 44.0 // Bridge centerline X.
	bigBenBridgW = 5.0  // Bridge half width.
)

// bigBenTower is the clock tower's plan center.
var bigBenTower = P(31, 6)

func (b BigBen) Build(g *Grid) Order {
	b.ground(g)
	b.palace(g)
	b.westminsterHall(g)
	b.centralTower(g, P(-13, -3))
	b.victoriaTower(g, P(-52, -5))
	b.clockTower(g, bigBenTower)
	b.bridge(g)
	b.square(g)
	b.vehicles(g)
	b.boat(g, P(-8, 27), 8)
	b.pier(g)
	// Masons work outward from the clock tower.
	return Outward(bigBenTower)
}

// raster sets every cell whose center lies in the box and for which f reports a color,
// without half-unit snapping.
func (BigBen) raster(g *Grid, bx Box, f func(Vec3) (Color, bool)) {
	s := g.Span()
	lo := func(v float64) int32 { return int32(math.Ceil(v * 2 / s)) }
	hi := func(v float64) int32 { return int32(math.Floor(v * 2 / s)) }
	previous := g.Brush()
	for y := lo(bx.Min.Y); y <= hi(bx.Max.Y); y++ {
		z0, x0 := lo(bx.Min.Z), lo(bx.Min.X)
		if (z0^y)&1 != 0 {
			z0++
		}
		if (x0^y)&1 != 0 {
			x0++
		}
		for z := z0; z <= hi(bx.Max.Z); z += 2 {
			for x := x0; x <= hi(bx.Max.X); x += 2 {
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

// arch is a Gothic pointed opening: jambs up to spring above the sill, then two arcs.
func (BigBen) arch(sill Vec3, hw, spring, depth float64, through Axis) Solid {
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
func (BigBen) octagon(center Vec2, r, y0, y1 float64) Prism {
	poly := make([]Vec2, 8)
	for i := range poly {
		poly[i] = Polar(center, r, (float64(i)+0.5)*math.Pi/4)
	}
	return Prism{Polygon: poly, Y0: y0, Y1: y1}
}

// spire is an octagonal pyramid.
func (BigBen) spire(center Vec2, r, base, height float64) Solid {
	bx := Box{V(center.X-r, base, center.Y-r), V(center.X+r, base+height, center.Y+r)}
	return Func{Box: bx, F: func(p Vec3) bool {
		rr := r * (1 - (p.Y-base)/height)
		d := p.Plan().Sub(center)
		a := math.Mod(math.Abs(d.Angle())+math.Pi/8, math.Pi/4)
		return d.Length()*math.Cos(a-math.Pi/8) <= rr*math.Cos(math.Pi/8)
	}}
}

// roof is a steep roof over a rectangle with horizontal inset k per unit of height:
// a pyramid, or a gable whose ridge runs along X or Z.
func (BigBen) roof(min, max Vec2, base, k float64, kind int) Solid {
	hx, hz := (max.X-min.X)/2, (max.Y-min.Y)/2
	rise := math.Min(hx, hz) / k
	if kind == bigBenGableX {
		rise = hz / k
	} else if kind == bigBenGableZ {
		rise = hx / k
	}
	bx := Box{min.At(base), max.At(base + rise)}
	c := min.Add(max).Scale(0.5)
	return Func{Box: bx, F: func(p Vec3) bool {
		in := (p.Y - base) * k
		dx, dz := math.Abs(p.X-c.X), math.Abs(p.Z-c.Y)
		switch kind {
		case bigBenGableX:
			return dz <= hz-in
		case bigBenGableZ:
			return dx <= hx-in
		}
		return dx <= hx-in && dz <= hz-in
	}}
}

const (
	bigBenPyramid = iota
	bigBenGableX
	bigBenGableZ
)

// face is one side of a rectangular block: its outward normal, center, along direction
// and half length.
type bigBenFace struct {
	n, c, t Vec2
	half    float64
	through Axis
}

func (BigBen) faces(min, max Vec2) []bigBenFace {
	c := min.Add(max).Scale(0.5)
	hx, hz := (max.X-min.X)/2, (max.Y-min.Y)/2
	return []bigBenFace{
		{P(0, 1), P(c.X, max.Y), P(1, 0), hx, AxisZ},
		{P(0, -1), P(c.X, min.Y), P(1, 0), hx, AxisZ},
		{P(1, 0), P(max.X, c.Y), P(0, 1), hz, AxisX},
		{P(-1, 0), P(min.X, c.Y), P(0, 1), hz, AxisX},
	}
}

// shell fills a hollow box with walls of the given thickness and paints its inner
// faces dark, so openings read as dark.
func (BigBen) shell(g *Grid, min, max Vec2, y0, y1, wall float64, color Color, floor bool) {
	g.With(color, func() { g.Fill(Block(min.At(y0), max.At(y1))) })
	f := 0.0
	if floor {
		f = 0.5
	}
	g.Carve(Block(V(min.X+wall, y0+f, min.Y+wall), V(max.X-wall, y1+0.1, max.Y-wall)))
	g.Paint(Charcoal, Block(V(min.X+wall-0.5, y0-0.6, min.Y+wall-0.5), V(max.X-wall+0.5, y1, max.Y-wall+0.5)))
}

// window carves a pointed window through a face and frames it in stone.
func (b BigBen) window(g *Grid, f bigBenFace, along, sill, hw, h float64, frame Color) {
	spring := math.Max(0.2, h-math.Sqrt(3)*hw)
	at := f.c.Add(f.t.Scale(along)).At(sill)
	g.Paint(frame, Subtract(b.arch(at.Sub(V(0, 0.45, 0)), hw+0.45, spring+0.2, 1.0, f.through), b.arch(at, hw, spring, 3, f.through)))
	g.Carve(b.arch(at, hw, spring, 4.0, f.through))
}

// block is a Gothic Revival range: a hollow sand-stone box with buttress piers and
// pinnacles between bays of tall pointed windows, string courses, a pierced parapet
// and a steep slate roof with gilded cresting.
func (b BigBen) block(g *Grid, min, max Vec2, eaves float64, kind int, k, bay float64, wall Color) {
	y0 := bigBenLand
	b.shell(g, min, max, y0, eaves, 1.2, wall, false)
	h := eaves - y0
	type row struct{ sill, h float64 }
	rows := []row{}
	for _, r := range []row{{1.0, 3.0}, {4.9, 3.6}, {9.3, 2.4}, {13.3, 2.6}} {
		if r.sill+r.h+0.8 <= h {
			rows = append(rows, r)
		}
	}
	// String courses.
	ring := func(y, t float64) Solid {
		return Subtract(Block(V(min.X-1, y, min.Y-1), V(max.X+1, y+t, max.Y+1)), Block(V(min.X+0.5, y, min.Y+0.5), V(max.X-0.5, y+t, max.Y-0.5)))
	}
	g.Paint(Stone, ring(y0+4.3, 0.4), ring(eaves-0.4, 0.4))
	for _, f := range b.faces(min, max) {
		n := math.Max(1, math.Round(2*f.half/bay))
		pitch := 2 * f.half / n
		for i := 0.0; i <= n; i++ {
			u := -f.half + i*pitch
			p := f.c.Add(f.t.Scale(u))
			w := 0.5
			if i == 0 || i == n {
				w = 0.8
			}
			lo, hi := p.Sub(f.t.Scale(w)).Sub(f.n.Scale(0.3)), p.Add(f.t.Scale(w)).Add(f.n.Scale(0.8))
			g.With(wall, func() {
				g.Fill(Block(lo.At(y0), hi.At(eaves+1.0)))
				q := p.Add(f.n.Scale(0.25))
				g.Fill(BlockAt(q, 1.0, 1.0, eaves+1.0, eaves+2.2))
			})
			q := p.Add(f.n.Scale(0.25))
			g.With(Stone, func() { g.Fill(Capsule{q.At(eaves + 2.2), q.At(eaves + 3.6), 0.5}) })
		}
		for i := 0.0; i < n; i++ {
			u := -f.half + (i+0.5)*pitch
			hw := math.Min(0.8, pitch/2-1.0)
			if hw < 0.45 {
				continue
			}
			for _, r := range rows {
				b.window(g, f, u, y0+r.sill, hw, r.h, Stone)
			}
		}
	}
	// Pierced parapet.
	g.With(wall, func() {
		g.Fill(Subtract(Block(V(min.X-0.2, eaves, min.Y-0.2), V(max.X+0.2, eaves+1.0, max.Y+0.2)),
			Block(V(min.X+0.9, eaves, min.Y+0.9), V(max.X-0.9, eaves+1.0, max.Y-0.9))))
	})
	for _, f := range b.faces(min, max) {
		for u := -f.half + 1.6; u < f.half-1.2; u += 1.6 {
			p := f.c.Add(f.t.Scale(u))
			g.Carve(BlockAt(p, 0.7, 0.7, eaves+0.5, eaves+1.0))
		}
	}
	// Steep slate roof, hollow, with sand gable ends and gilded cresting.
	rmin, rmax := min.Add(P(0.9, 0.9)), max.Sub(P(0.9, 0.9))
	r := b.roof(rmin, rmax, eaves, k, kind)
	g.With(Slate, func() { g.FillVolume(r) })
	g.Carve(Translate(r, V(0, -1.6, 0)))
	top := r.Bounds().Max.Y
	c := rmin.Add(rmax).Scale(0.5)
	switch kind {
	case bigBenGableX:
		g.With(wall, func() {
			g.Fill(Intersect(r, Block(V(rmin.X, eaves, rmin.Y), V(rmin.X+0.9, top, rmax.Y))),
				Intersect(r, Block(V(rmax.X-0.9, eaves, rmin.Y), V(rmax.X, top, rmax.Y))))
		})
		g.With(Gold, func() { g.Fill(Capsule{V(rmin.X+0.6, top+0.1, c.Y), V(rmax.X-0.6, top+0.1, c.Y), 0.5}) })
		for x := rmin.X + 2; x < rmax.X-1; x += 4 {
			g.With(Gold, func() { g.Fill(Capsule{V(x, top, c.Y), V(x, top+1.3, c.Y), 0.45}) })
		}
	case bigBenGableZ:
		g.With(wall, func() {
			g.Fill(Intersect(r, Block(V(rmin.X, eaves, rmin.Y), V(rmax.X, top, rmin.Y+0.9))),
				Intersect(r, Block(V(rmin.X, eaves, rmax.Y-0.9), V(rmax.X, top, rmax.Y))))
		})
		g.With(Gold, func() { g.Fill(Capsule{V(c.X, top+0.1, rmin.Y+0.6), V(c.X, top+0.1, rmax.Y-0.6), 0.5}) })
		for z := rmin.Y + 2; z < rmax.Y-1; z += 4 {
			g.With(Gold, func() { g.Fill(Capsule{V(c.X, top, z), V(c.X, top+1.3, z), 0.45}) })
		}
	default:
		g.With(Gold, func() { g.Fill(Capsule{c.At(top - 0.6), c.At(top + 2.0), 0.5}) })
	}
}

// ground lays the two banks as hollow plinths with a two-layer slab, and the river.
func (b BigBen) ground(g *Grid) {
	// River: a single water layer between the embankments.
	g.Use(Water).Fill(Block(V(bigBenX0, -0.1, bigBenBankZ), V(bigBenX1, 0.1, bigBenFarZ)))
	slab := func(z0, z1 float64, color func(Vec3) Color) {
		b.raster(g, Box{V(bigBenX0, bigBenLand-0.75, z0), V(bigBenX1, bigBenLand, z1)}, func(p Vec3) (Color, bool) {
			return color(p), true
		})
		g.With(DarkStone, func() {
			g.Fill(Block(V(bigBenX0, 0, z0), V(bigBenX0+0.8, bigBenLand, z1)),
				Block(V(bigBenX1-0.8, 0, z0), V(bigBenX1, bigBenLand, z1)))
		})
	}
	// Palace bank.
	slab(bigBenZ0, bigBenBankZ, func(p Vec3) Color {
		switch {
		case math.Abs(p.X-bigBenBridge) <= bigBenBridgW-0.6:
			return Charcoal // Bridge Street.
		case math.Abs(p.X-bigBenBridge) <= bigBenBridgW+1.2:
			return Stone
		case p.X > bigBenBridge && p.Z < 8:
			return Grass // Parliament Square.
		case p.X < -60:
			return Grass // Victoria Tower Gardens.
		case p.Z > 10.8:
			return Stone // River terrace.
		case p.Z < -17.5 && p.X > -60 && p.X < 22:
			return Charcoal // Abingdon Street.
		}
		return Stone
	})
	g.With(DarkStone, func() { g.Fill(Block(V(bigBenX0, 0, bigBenZ0), V(bigBenX1, bigBenLand, bigBenZ0+0.8))) })
	// Far bank: embankment walk and lawn with plane trees.
	slab(bigBenFarZ, bigBenZ1, func(p Vec3) Color {
		if math.Abs(p.X-bigBenBridge) <= bigBenBridgW-0.6 {
			return Charcoal
		}
		if p.Z < bigBenFarZ+3 || math.Abs(p.X-bigBenBridge) <= bigBenBridgW+1.2 {
			return Sand
		}
		return Grass
	})
	g.With(DarkStone, func() { g.Fill(Block(V(bigBenX0, 0, bigBenZ1-0.8), V(bigBenX1, bigBenLand, bigBenZ1))) })
	// Embankment walls with a granite parapet.
	for _, z := range []float64{bigBenBankZ, bigBenFarZ} {
		g.With(Stone, func() { g.Fill(Block(V(bigBenX0, 0, z-0.5), V(bigBenX1, bigBenLand, z+0.5))) })
		for _, span := range [][2]float64{{bigBenX0, bigBenBridge - bigBenBridgW - 0.5}, {bigBenBridge + bigBenBridgW + 0.5, bigBenX1}} {
			g.With(DarkStone, func() { g.Fill(Block(V(span[0], bigBenLand, z-0.4), V(span[1], bigBenLand+0.9, z+0.4))) })
		}
	}
	for x := bigBenX0 + 4; x < bigBenX1-2; x += 10 {
		if math.Abs(x-bigBenBridge) < 8 {
			continue
		}
		b.tree(g, P(x+2*Hash01(int(x), 1), bigBenZ1-3.0), 5.5+1.5*Hash01(int(x), 2))
	}
}

// tree is a plane tree: a trunk and a lumpy foliage crown.
func (b BigBen) tree(g *Grid, at Vec2, h float64) {
	y := bigBenLand
	g.With(Wood, func() { g.Fill(Cylinder{Center: at, Radius: 0.6, Y0: y, Y1: y + h*0.55}) })
	g.With(Foliage, func() {
		g.FillVolume(Ellipsoid{at.At(y + h*0.68), V(h*0.34, h*0.3, h*0.34)})
		g.FillVolume(Ellipsoid{at.Add(P(h*0.15, -h*0.1)).At(y + h*0.82), V(h*0.22, h*0.2, h*0.22)})
	})
}

// palace builds the river front, back range, cross ranges and courtyards.
func (b BigBen) palace(g *Grid) {
	// Courtyards between the ranges.
	g.Paint(Grass, Block(V(-37, bigBenLand-0.3, -8), V(-17, bigBenLand, 2)), Block(V(-9, bigBenLand-0.3, -8), V(13, bigBenLand, 2)))
	// River front.
	b.block(g, P(-44, 2), P(24, 11), 14, bigBenGableX, 0.7, 3.4, Sand)
	// Back range and cross ranges.
	b.block(g, P(-40, -17), P(20, -8), 13, bigBenGableX, 0.7, 3.4, Sand)
	b.block(g, P(-44, -12), P(-37, 3), 13, bigBenGableZ, 0.7, 3.4, Sand)
	b.block(g, P(-17, -12), P(-9, 3), 13, bigBenGableZ, 0.7, 3.4, Sand)
	b.block(g, P(13, -12), P(20, 3), 13, bigBenGableZ, 0.7, 3.4, Sand)
	// Pavilions on the river front.
	for _, x := range []float64{-40.4, -20.5, -1, 19.5} {
		b.block(g, P(x-3.6, 1), P(x+3.6, 12), 18.5, bigBenPyramid, 0.36, 3.4, Sand)
	}
	// Link to the clock tower.
	b.block(g, P(23, 3), P(27.5, 9), 9, bigBenGableX, 0.8, 4.5, Sand)
	// Terrace balustrade along the river.
	g.With(Stone, func() { g.Fill(Block(V(-45, bigBenLand, 12.4), V(25, bigBenLand+1.0, 13))) })
}

// westminsterHall is the medieval hall behind the clock tower: grey stone walls with
// buttresses, a great steep roof and a large end window between two turrets.
func (b BigBen) westminsterHall(g *Grid) {
	min, max := P(23.5, -20.5), P(32.5, -4)
	b.block(g, min, max, 11.5, bigBenGableZ, 0.5, 4.0, Stone)
	c := min.Add(max).Scale(0.5)
	// Great north window and flanking turrets.
	f := bigBenFace{P(0, -1), P(c.X, min.Y), P(1, 0), 4.5, AxisZ}
	b.window(g, f, 0, 9.5, 2.2, 7.5, Sand)
	g.With(Sand, func() {
		g.Fill(Block(V(c.X-0.4, 9.5, min.Y+0.6), V(c.X+0.4, 15.5, min.Y+1.8)))
	})
	b.window(g, f, 0, bigBenLand, 1.4, 4.0, Sand)
	for _, sx := range []float64{-1, 1} {
		tc := P(c.X+sx*4.5, min.Y+0.3)
		g.With(Sand, func() { g.Fill(b.octagon(tc, 1.5, bigBenLand, 21)) })
		g.With(Slate, func() { g.FillVolume(b.spire(tc, 1.7, 21, 4.5)) })
		g.With(Gold, func() { g.Fill(Capsule{tc.At(25), tc.At(26.5), 0.45}) })
	}
}

// centralTower is the octagonal lantern over the Central Lobby.
func (b BigBen) centralTower(g *Grid, c Vec2) {
	const r, top = 4.4, 29.0
	g.With(Sand, func() { g.Fill(b.octagon(c, r, bigBenLand, top)) })
	g.Carve(b.octagon(c, r-1.2, bigBenLand+0.5, top-0.5))
	g.Paint(Charcoal, b.octagon(c, r-0.7, bigBenLand, top))
	g.Paint(Stone, Subtract(b.octagon(c, r+0.5, 18, 18.5), b.octagon(c, r-0.5, 18, 18.5)))
	// Tall lantern windows on every face, with a pinnacle at every angle.
	for i := 0; i < 8; i++ {
		a := float64(i) * math.Pi / 4
		n := P(math.Cos(a), math.Sin(a))
		p := c.Add(n.Scale(r * math.Cos(math.Pi/8)))
		sill := 19.5
		sp := Rotated{b.arch(V(0, sill, 0), 0.9, 5.0-math.Sqrt(3)*0.9, 3, AxisX), P(0, 0), a}
		g.Paint(Stone, Translate(Rotated{Subtract(b.arch(V(0, sill-0.4, 0), 1.4, 5.0-math.Sqrt(3)*0.9, 1.0, AxisX), b.arch(V(0, sill, 0), 0.9, 5.0-math.Sqrt(3)*0.9, 3, AxisX)), P(0, 0), a}, p.At(0)))
		g.Carve(Translate(sp, p.At(0)))
		q := Polar(c, r+0.2, a+math.Pi/8)
		g.With(Sand, func() { g.Fill(Cylinder{Center: q, Radius: 0.65, Y0: top, Y1: top + 2.6}) })
		g.With(Stone, func() { g.Fill(Capsule{q.At(top + 2.6), q.At(top + 4.2), 0.5}) })
	}
	g.With(Sand, func() { g.Fill(Subtract(b.octagon(c, r+0.3, top, top+0.9), b.octagon(c, r-0.6, top, top+0.9))) })
	g.With(Sand, func() { g.FillVolume(b.spire(c, r-0.5, top, 13)) })
	g.Paint(Stone, Block(V(c.X-r, top+4, c.Y-r), V(c.X+r, top+4.4, c.Y+r)))
	g.With(Gold, func() { g.Fill(Capsule{c.At(top + 12), c.At(top + 15), 0.5}) })
}

// victoriaTower is the great square tower at the south end, with the Sovereign's
// Entrance arch, tiers of tall windows, octagonal corner turrets and a flag.
func (b BigBen) victoriaTower(g *Grid, c Vec2) {
	const h, top = 7.0, 45.0
	min, max := c.Sub(P(h, h)), c.Add(P(h, h))
	b.shell(g, min, max, bigBenLand, top, 1.2, Sand, false)
	ring := func(y, t float64) Solid {
		return Subtract(Block(V(min.X-1, y, min.Y-1), V(max.X+1, y+t, max.Y+1)), Block(V(min.X+0.5, y, min.Y+0.5), V(max.X-0.5, y+t, max.Y-0.5)))
	}
	for _, f := range b.faces(min, max) {
		// Pilaster strips dividing each face into three panels.
		for _, u := range []float64{-2.4, 2.4} {
			p := f.c.Add(f.t.Scale(u))
			lo, hi := p.Sub(f.t.Scale(0.5)).Sub(f.n.Scale(0.3)), p.Add(f.t.Scale(0.5)).Add(f.n.Scale(0.7))
			g.With(Sand, func() { g.Fill(Block(lo.At(bigBenLand+12), hi.At(top))) })
		}
		// Sovereign's Entrance on the south face; smaller doors elsewhere.
		if f.n.Y > 0 {
			b.window(g, f, 0, bigBenLand, 2.6, 8.5, Stone)
		} else {
			b.window(g, f, 0, bigBenLand, 1.4, 4.5, Stone)
		}
		for _, u := range []float64{-4.4, 0, 4.4} {
			b.window(g, f, u, 14, 0.75, 4.0, Stone)
			b.window(g, f, u, 20, 0.75, 4.0, Stone)
			b.window(g, f, u, 28, 0.8, 11.0, Stone)
		}
		for _, u := range []float64{-4.4, 4.4} {
			b.window(g, f, u, 7.5, 0.75, 3.5, Stone)
		}
	}
	g.Paint(Stone, ring(12, 0.5), ring(18.6, 0.4), ring(25.5, 0.5), ring(top-1, 0.5))
	// Parapet.
	g.With(Sand, func() {
		g.Fill(Subtract(Block(V(min.X-0.3, top, min.Y-0.3), V(max.X+0.3, top+1.4, max.Y+0.3)),
			Block(V(min.X+0.9, top, min.Y+0.9), V(max.X-0.9, top+1.4, max.Y-0.9))))
	})
	for _, f := range b.faces(min, max) {
		for u := -5.4; u < 5.5; u += 1.8 {
			g.Carve(BlockAt(f.c.Add(f.t.Scale(u)), 0.8, 0.8, top+0.7, top+1.4))
		}
		// Small parapet pinnacles mid-face.
		for _, u := range []float64{-2.4, 2.4} {
			q := f.c.Add(f.t.Scale(u)).Add(f.n.Scale(0.2))
			g.With(Sand, func() { g.Fill(BlockAt(q, 1.0, 1.0, top, top+2.4)) })
			g.With(Gold, func() { g.Fill(Capsule{q.At(top + 2.4), q.At(top + 3.6), 0.45}) })
		}
	}
	// Corner turrets with gilded crowns.
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			tc := P(c.X+sx*h, c.Y+sz*h)
			g.With(Sand, func() { g.Fill(b.octagon(tc, 1.8, bigBenLand, top+4)) })
			g.Carve(b.octagon(tc, 0.8, bigBenLand+0.5, top+3.5))
			g.Paint(Stone, b.octagon(tc, 1.9, top-1, top-0.5))
			g.With(Sand, func() { g.Fill(Subtract(b.octagon(tc, 2.1, top+4, top+4.8), b.octagon(tc, 1.0, top+4, top+4.8))) })
			g.With(Slate, func() { g.FillVolume(b.spire(tc, 1.6, top+4.8, 3.5)) })
			g.With(Gold, func() {
				g.Fill(Subtract(b.octagon(tc, 1.7, top+4.8, top+5.6), b.octagon(tc, 0.8, top+4.8, top+5.6)))
				g.Fill(Capsule{tc.At(top + 7.8), tc.At(top + 9.4), 0.5})
			})
		}
	}
	// Low roof and flagstaff with the Union flag flying towards +X.
	g.With(Slate, func() { g.FillVolume(b.roof(min.Add(P(1, 1)), max.Sub(P(1, 1)), top, 2.5, bigBenPyramid)) })
	pole := c
	g.With(White, func() { g.Fill(Capsule{pole.At(top), pole.At(top + 15), 0.45}) })
	g.With(Gold, func() { g.Fill(Sphere(pole.At(top+15.3), 0.6)) })
	fx0, fx1, fy0, fy1 := pole.X+0.4, pole.X+7.0, top+10.4, top+14.6
	b.raster(g, Box{V(fx0, fy0, pole.Y-0.5), V(fx1, fy1, pole.Y+0.5)}, func(p Vec3) (Color, bool) {
		u := (p.X - fx0) / (fx1 - fx0)
		v := (p.Y - fy0) / (fy1 - fy0)
		du, dv := math.Abs(u-0.5), math.Abs(v-0.5)
		switch {
		case du < 0.08 || dv < 0.14:
			return Red, true
		case du < 0.15 || dv < 0.26:
			return White, true
		case math.Abs(du*(fx1-fx0)-dv*(fy1-fy0)*1.0) < 0.6:
			return White, true
		}
		return Navy, true
	})
}

// clockTower is the Elizabeth Tower: a panelled shaft, the clock stage with four gilded
// dials, the open belfry and the steep cast-iron spire with lucarnes and a lantern.
func (b BigBen) clockTower(g *Grid, c Vec2) {
	y0 := bigBenLand
	const (
		hs     = 4.6  // Shaft half width.
		shaft  = 33.0 // Top of the panelled shaft.
		hc     = 5.0  // Clock stage half width.
		clock  = 41.5 // Top of the clock stage.
		hb     = 4.6  // Belfry half width.
		belfry = 48.5
	)
	sq := func(h, y0, y1 float64) Cuboid { return Block(V(c.X-h, y0, c.Y-h), V(c.X+h, y1, c.Y+h)) }
	// Plinth.
	g.With(Stone, func() { g.Fill(sq(hs+0.7, y0, y0+1.6)) })
	// Shaft with recessed panel tracery: sand ribs, stone panels with dark lancets.
	g.With(Sand, func() { g.Fill(sq(hs, y0, shaft)) })
	g.Carve(sq(hs-1.8, y0+0.5, belfry-0.5))
	g.Paint(Charcoal, sq(hs-1.3, y0, belfry))
	faces := b.faces(c.Sub(P(hs, hs)), c.Add(P(hs, hs)))
	for _, f := range faces {
		for _, u := range []float64{-2.4, 0, 2.4} {
			p := f.c.Add(f.t.Scale(u))
			lo, hi := p.Sub(f.t.Scale(0.8)).Sub(f.n.Scale(0.8)), p.Add(f.t.Scale(0.8)).Add(f.n.Scale(0.5))
			panel := Block(lo.At(y0+3), hi.At(shaft-1.2))
			g.Paint(Stone, Block(p.Sub(f.t.Scale(0.8)).Sub(f.n.Scale(1.3)).At(y0+3), p.Add(f.t.Scale(0.8)).Add(f.n.Scale(0.5)).At(shaft-1.2)))
			g.Carve(panel)
			// Transoms every tier.
			for y := y0 + 5.6; y < shaft-1.6; y += 3.2 {
				g.With(Sand, func() { g.Fill(Block(lo.At(y), hi.Sub(f.n.Scale(0.2)).At(y+0.5))) })
			}
			if u == 0 {
				for y := y0 + 6.6; y < shaft-3; y += 6.4 {
					b.window(g, f, u, y, 0.5, 1.8, Stone)
				}
			}
		}
		// Door at the foot of the tower.
		if f.n.X > 0 {
			b.window(g, f, 0, y0, 1.0, 2.8, Stone)
		}
	}
	g.Paint(Stone, Subtract(sq(hs+1, shaft-1.0, shaft-0.6), sq(hs-0.5, shaft-1.0, shaft-0.6)))
	// Clock stage.
	g.With(Sand, func() {
		g.Fill(Subtract(sq(hc, shaft, clock), sq(hc-1.4, shaft, clock)))
		g.Fill(Subtract(sq(hc+0.3, shaft-0.4, shaft+0.4), sq(hc-1.4, shaft-0.4, shaft+0.4)))
		g.Fill(Subtract(sq(hc+0.3, clock-0.6, clock+0.2), sq(hc-1.4, clock-0.6, clock+0.2)))
	})
	cy := (shaft + clock) / 2
	const rd = 3.1
	for _, f := range b.faces(c.Sub(P(hc, hc)), c.Add(P(hc, hc))) {
		// A point on the face: along, up, depth (outward from the face plane).
		at := func(u, v, d float64) Vec3 { return f.c.Add(f.t.Scale(u)).Add(f.n.Scale(d)).At(cy + v) }
		local := func(p Vec3) (u, v, d float64) {
			q := p.Plan().Sub(f.c)
			return q.Dot(f.t), p.Y - cy, q.Dot(f.n)
		}
		bx := Box{at(0, 0, 0).Sub(V(4, 4, 4)), at(0, 0, 0).Add(V(4, 4, 4))}
		// Gilded square surround standing proud of the face, then the dial.
		g.With(Gold, func() {
			g.Fill(Func{Box: bx, F: func(p Vec3) bool {
				u, v, d := local(p)
				r := math.Hypot(u, v)
				return d > -0.4 && d <= 0.5 && math.Abs(u) <= rd+0.9 && math.Abs(v) <= rd+0.9 && r > rd-0.2
			}})
		})
		b.raster(g, bx, func(p Vec3) (Color, bool) {
			u, v, d := local(p)
			r := math.Hypot(u, v)
			if d < -0.8 || d > 0.2 || r > rd {
				return 0, false
			}
			return White, true
		})
		// Hour marks and hands (ten past ten).
		for i := 0; i < 12; i++ {
			a := float64(i) * math.Pi / 6
			g.Paint(Charcoal, Sphere(at(math.Sin(a)*(rd-0.5), math.Cos(a)*(rd-0.5), 0), 0.3))
		}
		hour, minute := -2*math.Pi/6-math.Pi/36, 2*math.Pi/6
		g.Paint(Charcoal, Capsule{at(0, 0, 0), at(math.Sin(hour)*1.3, math.Cos(hour)*1.3, 0), 0.32},
			Capsule{at(0, 0, 0), at(math.Sin(minute)*2.1, math.Cos(minute)*2.1, 0), 0.3})
		g.Paint(Gold, Sphere(at(0, 0, 0), 0.35))
		// Gable over each dial.
		top := at(0, rd+1.2, 0.2)
		g.With(Sand, func() {
			g.Fill(Intersect(b.roof(top.Plan().Sub(f.t.Scale(rd+0.9)).Sub(f.n.Scale(0.8)), top.Plan().Add(f.t.Scale(rd+0.9)).Add(f.n.Scale(0.6)), clock-0.3, 0.8, map[Axis]int{AxisZ: bigBenGableZ, AxisX: bigBenGableX}[f.through]),
				Block(V(c.X-hc-1, clock-0.3, c.Y-hc-1), V(c.X+hc+1, clock+4, c.Y+hc+1))))
		})
		g.With(Gold, func() {
			g.Fill(Capsule{at(0, clock-cy-0.3+(rd+0.9)/0.8-0.4, 0.2), at(0, clock-cy-0.3+(rd+0.9)/0.8+1.2, 0.2), 0.45})
		})
	}
	// Corner pinnacles of the clock stage.
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			q := P(c.X+sx*(hc-0.2), c.Y+sz*(hc-0.2))
			g.With(Sand, func() { g.Fill(b.octagon(q, 0.9, shaft, clock+2.5)) })
			g.With(Gold, func() { g.FillVolume(b.spire(q, 1.0, clock+2.5, 2.6)) })
		}
	}
	// Belfry with three open arches per face, and the great bell inside.
	g.With(Charcoal, func() { g.Fill(sq(hb-1, clock-0.4, clock)) })
	g.With(Sand, func() { g.Fill(Subtract(sq(hb, clock, belfry), sq(hb-1.1, clock, belfry))) })
	g.Paint(Charcoal, Subtract(sq(hb-0.6, clock, belfry), sq(hb-1.2, clock, belfry)))
	for _, f := range b.faces(c.Sub(P(hb, hb)), c.Add(P(hb, hb))) {
		for _, u := range []float64{-2.3, 0, 2.3} {
			b.window(g, f, u, clock+0.9, 0.75, 4.4, Gold)
		}
	}
	g.With(Gold, func() {
		g.Fill(Subtract(sq(hb+0.3, belfry-0.6, belfry+0.2), sq(hb-1.2, belfry-0.6, belfry+0.2)))
	})
	g.With(DarkStone, func() {
		g.FillVolume(Frustum{Center: c, R0: 2.2, R1: 1.1, Y0: clock + 1.5, Y1: clock + 4.4})
	})
	// Spire: steep slate pyramid with gilded hips and lucarnes, a gilded lantern stage,
	// then the upper spire and finial.
	const k, lantern = 0.25, 56.5
	hipped := func(h, base, kk float64, y1 float64, color Color) {
		r := b.roof(c.Sub(P(h, h)), c.Add(P(h, h)), base, kk, bigBenPyramid)
		g.With(color, func() { g.FillVolume(Intersect(r, Block(V(c.X-h, base, c.Y-h), V(c.X+h, y1, c.Y+h)))) })
		g.Paint(Gold, Func{Box: r.Bounds(), F: func(p Vec3) bool {
			in := h - (p.Y-base)*kk
			dx, dz := math.Abs(p.X-c.X), math.Abs(p.Z-c.Y)
			return p.Y <= y1 && in-dx < 0.75 && in-dz < 0.75
		}})
	}
	hipped(hb, belfry, k, lantern, Slate)
	g.Paint(Gold, Block(V(c.X-hb, belfry+3.6, c.Y-hb), V(c.X+hb, belfry+3.9, c.Y+hb)))
	for _, f := range b.faces(c.Sub(P(hb-0.6, hb-0.6)), c.Add(P(hb-0.6, hb-0.6))) {
		p := f.c
		g.With(Gold, func() {
			g.Fill(BlockAt(p, 2.2, 2.2, belfry, belfry+2.6))
			kind := bigBenGableZ
			if f.through == AxisX {
				kind = bigBenGableX
			}
			g.Fill(b.roof(p.Sub(P(1.2, 1.2)), p.Add(P(1.2, 1.2)), belfry+2.6, 0.7, kind))
		})
		b.window(g, f, 0, belfry+0.6, 0.5, 1.6, Gold)
	}
	// Corner spirelets.
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			q := P(c.X+sx*(hb-0.1), c.Y+sz*(hb-0.1))
			g.With(Gold, func() { g.Fill(b.octagon(q, 0.8, belfry, belfry+1.6)) })
			g.With(Slate, func() { g.FillVolume(b.spire(q, 0.95, belfry+1.6, 3.6)) })
			g.With(Gold, func() { g.Fill(Capsule{q.At(belfry + 4.6), q.At(belfry + 5.8), 0.45}) })
		}
	}
	// Lantern (the Ayrton Light stage).
	const hl = 2.2
	g.With(Gold, func() {
		g.Fill(Subtract(sq(hl, lantern, lantern+3.2), sq(hl-1.0, lantern+0.4, lantern+2.8)))
		g.Fill(Subtract(sq(hl+0.5, lantern, lantern+0.5), sq(hl-1.0, lantern, lantern+0.5)))
	})
	for _, f := range b.faces(c.Sub(P(hl, hl)), c.Add(P(hl, hl))) {
		g.Carve(b.arch(f.c.At(lantern+0.6), 0.6, 1.2, 2.0, f.through))
	}
	g.With(White, func() { g.Fill(sq(0.7, lantern+0.5, lantern+2.4)) })
	hipped(hl+0.2, lantern+3.2, 0.24, lantern+3.2+(hl+0.2)/0.24, Slate)
	tip := lantern + 3.2 + (hl+0.2)/0.24
	g.With(Gold, func() {
		g.Fill(Capsule{c.At(tip - 1), c.At(tip + 2.6), 0.5})
		g.Fill(Capsule{c.Sub(P(0.9, 0)).At(tip + 1.6), c.Add(P(0.9, 0)).At(tip + 1.6), 0.4})
	})
}

// bridge is Westminster Bridge: seven shallow arches of green-painted iron on granite
// piers, a road with pavements, an openwork parapet and lamp standards over the piers.
func (b BigBen) bridge(g *Grid) {
	x0, x1 := bigBenBridge-bigBenBridgW, bigBenBridge+bigBenBridgW
	z0, z1 := bigBenBankZ, bigBenFarZ
	const n = 5
	pitch := (z1 - z0) / n
	const pier, crown = 1.3, 3.7
	// Piers with pointed cutwaters up- and downstream.
	piers := []float64{}
	for i := 1; i < n; i++ {
		z := z0 + float64(i)*pitch
		piers = append(piers, z)
		g.With(Stone, func() {
			g.Fill(Prism{Polygon: []Vec2{P(x0-1.8, z), P(x0-0.6, z-pier/2), P(x1+0.6, z-pier/2), P(x1+1.8, z), P(x1+0.6, z+pier/2), P(x0-0.6, z+pier/2)}, Y0: -0.2, Y1: 2.0})
		})
	}
	// Fascias: dark green spandrels with lighter arch ribs.
	inArch := func(p Vec3) (inside bool, rib bool) {
		u := (p.Z - z0) / pitch
		i := math.Floor(u)
		if i < 0 || i >= n {
			return false, false
		}
		zc := z0 + (i+0.5)*pitch
		hw := (pitch - pier) / 2
		dz := (p.Z - zc) / hw
		dy := (p.Y - 0.4) / (crown - 0.4)
		if p.Y < 0.4 {
			dy = 0
		}
		e := dz*dz + dy*dy
		return e <= 1, e > 1 && math.Sqrt(dz*dz+dy*dy) < 1+0.7/hw
	}
	b.raster(g, Box{V(x0, 0, z0), V(x1, bigBenDeck+1.1, z1)}, func(p Vec3) (Color, bool) {
		edge := p.X < x0+1.0 || p.X > x1-1.0
		if p.Y > bigBenDeck {
			// Openwork parapet: posts and a top rail.
			if !edge {
				return 0, false
			}
			if p.Y > bigBenDeck+0.75 || math.Mod(p.Z-z0, 1.6) < 0.8 {
				return Grass, true
			}
			return 0, false
		}
		if p.Y > bigBenDeck-0.8 {
			switch {
			case edge:
				return Grass, true
			case p.X < x0+2.6 || p.X > x1-2.6:
				return Sand, true
			case math.Abs(p.X-bigBenBridge) < 0.3 && math.Mod(p.Z, 3) < 1.6:
				return White, true
			}
			return Charcoal, true
		}
		inside, rib := inArch(p)
		if inside {
			return 0, false
		}
		if !edge {
			// Iron ribs under the deck, between the fascias.
			if rib && math.Abs(p.X-bigBenBridge) < 2.6 && math.Abs(p.X-bigBenBridge) > 1.8 {
				return Foliage, true
			}
			return 0, false
		}
		if rib {
			return Grass, true
		}
		if p.Y > bigBenDeck-1.4 {
			return Grass, true // Cornice.
		}
		return Foliage, true
	})
	// Approach ramps onto both banks.
	ramp := func(za, zb float64) {
		lo, hi := math.Min(za, zb), math.Max(za, zb)
		b.raster(g, Box{V(x0, bigBenLand-0.4, lo), V(x1, bigBenDeck, hi)}, func(p Vec3) (Color, bool) {
			t := (p.Z - za) / (zb - za)
			top := bigBenLand + (bigBenDeck-bigBenLand)*t
			if p.Y > top {
				return 0, false
			}
			if p.X < x0+1.0 || p.X > x1-1.0 {
				return Stone, true
			}
			if p.X < x0+2.6 || p.X > x1-2.6 {
				return Sand, true
			}
			return Charcoal, true
		})
	}
	ramp(z0-9, z0)
	ramp(z1+5, z1)
	// Lamp standards over the piers and at the abutments.
	for _, z := range append(piers, z0+0.4, z1-0.4) {
		for _, x := range []float64{x0 + 0.5, x1 - 0.5} {
			q := P(x, z)
			g.With(Charcoal, func() { g.Fill(Capsule{q.At(bigBenDeck + 1), q.At(bigBenDeck + 3.6), 0.45}) })
			g.With(White, func() { g.Fill(Sphere(q.At(bigBenDeck+3.9), 0.5)) })
			g.With(Gold, func() { g.Fill(Capsule{q.At(bigBenDeck + 4.4), q.At(bigBenDeck + 4.8), 0.4}) })
		}
	}
}

// square is Parliament Square and the Victoria Tower Gardens: lawns, paths, trees and
// statues on plinths.
func (b BigBen) square(g *Grid) {
	g.Paint(Sand, Block(V(51, bigBenLand-0.3, -18), V(57, bigBenLand, -16.8)), Block(V(51, bigBenLand-0.3, -18), V(52.2, bigBenLand, 4)),
		Block(V(51, bigBenLand-0.3, 2.8), V(57, bigBenLand, 4)))
	for i, z := range []float64{-19, -11.5, -4, 3.5, 9} {
		b.tree(g, P(55.5+1.0*Hash01(i, 3), z), 6+1.5*Hash01(i, 4))
	}
	for i, z := range []float64{-12, -3} {
		q := P(53.6, z)
		g.With(Stone, func() { g.Fill(BlockAt(q, 1.8, 1.8, bigBenLand, bigBenLand+1.6)) })
		g.With(DarkStone, func() { g.Fill(Capsule{q.At(bigBenLand + 1.8), q.At(bigBenLand + 3.4 + 0.4*float64(i)), 0.55}) })
	}
	for i, z := range []float64{-18, -9, 0, 9} {
		b.tree(g, P(-63, z), 6+1.5*Hash01(i, 5))
	}
	// Plane trees along the back road.
	for i, x := range []float64{-34, -20, -6, 8} {
		b.tree(g, P(x, -20), 5+1.0*Hash01(i, 6))
	}
}

// vehicles places red double-decker buses and black cabs on the bridge and street.
func (b BigBen) vehicles(g *Grid) {
	b.bus(g, P(bigBenBridge-1.2, 19.5), bigBenDeck, 1)
	b.bus(g, P(bigBenBridge+1.2, 32), bigBenDeck, -1)
	b.bus(g, P(bigBenBridge+1.2, -12), bigBenLand, -1)
	b.cab(g, P(bigBenBridge-1.2, 28), bigBenDeck)
	b.cab(g, P(bigBenBridge+1.2, 24), bigBenDeck)
	b.cab(g, P(bigBenBridge-1.2, -4), bigBenLand)
	b.cab(g, P(bigBenBridge+1.2, -18), bigBenLand)
}

// bus is a red double-decker running along Z, its front facing dir.
func (b BigBen) bus(g *Grid, c Vec2, y, dir float64) {
	const hl, hw = 3.6, 1.05
	g.With(Charcoal, func() {
		for _, dz := range []float64{-2.4, 2.2} {
			g.Fill(Block(V(c.X-hw, y, c.Y+dz-0.6), V(c.X+hw, y+0.7, c.Y+dz+0.6)))
		}
	})
	g.With(Red, func() { g.Fill(Block(V(c.X-hw, y+0.5, c.Y-hl), V(c.X+hw, y+3.4, c.Y+hl))) })
	g.Paint(Glass, Block(V(c.X-hw, y+1.4, c.Y-hl+0.5), V(c.X+hw, y+1.9, c.Y+hl-0.5)),
		Block(V(c.X-hw, y+2.6, c.Y-hl+0.4), V(c.X+hw, y+3.0, c.Y+hl-0.4)),
		Block(V(c.X-hw+0.4, y+1.4, c.Y+dir*hl-0.4), V(c.X+hw-0.4, y+3.0, c.Y+dir*hl+0.4)))
	g.Paint(White, Block(V(c.X-hw, y+2.2, c.Y-hl), V(c.X+hw, y+2.3, c.Y+hl)))
}

// cab is a black taxi.
func (b BigBen) cab(g *Grid, c Vec2, y float64) {
	g.With(Charcoal, func() {
		g.Fill(Block(V(c.X-0.8, y, c.Y-1.7), V(c.X+0.8, y+1.1, c.Y+1.7)))
		g.Fill(Block(V(c.X-0.7, y+1.1, c.Y-1.0), V(c.X+0.7, y+1.8, c.Y+0.9)))
	})
	g.Paint(Glass, Block(V(c.X-0.7, y+1.2, c.Y-0.9), V(c.X+0.7, y+1.6, c.Y+0.8)))
	g.Paint(Gold, Block(V(c.X-0.3, y+1.7, c.Y-0.2), V(c.X+0.3, y+1.8, c.Y+0.2)))
}

// boat is a river cruise boat heading towards +X: white hull with a navy waterline,
// a long glazed saloon and an open top deck with a rail.
func (b BigBen) boat(g *Grid, c Vec2, hl float64) {
	const hw = 1.9
	hull := func(y0, y1, inset float64) Prism {
		return Prism{Polygon: []Vec2{
			P(c.X-hl+inset, c.Y-hw+inset), P(c.X+hl-3, c.Y-hw+inset), P(c.X+hl-inset, c.Y),
			P(c.X+hl-3, c.Y+hw-inset), P(c.X-hl+inset, c.Y+hw-inset), P(c.X-hl-0.4+inset, c.Y),
		}, Y0: y0, Y1: y1}
	}
	g.With(Navy, func() { g.Fill(hull(-0.3, 0.5, 0)) })
	g.With(White, func() { g.Fill(hull(0.5, 1.4, 0)) })
	g.Paint(Red, hull(1.1, 1.4, -0.2))
	g.With(White, func() {
		g.Fill(Block(V(c.X-hl+1.2, 1.4, c.Y-hw+0.4), V(c.X+hl-4.0, 3.2, c.Y+hw-0.4)))
	})
	g.Paint(Glass, Block(V(c.X-hl+1.6, 1.9, c.Y-hw), V(c.X+hl-4.4, 2.7, c.Y+hw)))
	g.With(Navy, func() {
		g.Fill(Subtract(Block(V(c.X-hl+1.2, 3.2, c.Y-hw+0.4), V(c.X+hl-4.0, 3.9, c.Y+hw-0.4)),
			Block(V(c.X-hl+1.8, 3.2, c.Y-hw+1.0), V(c.X+hl-4.6, 3.9, c.Y+hw-1.0))))
	})
	// Wheelhouse forward.
	g.With(White, func() { g.Fill(Block(V(c.X+hl-5.6, 3.2, c.Y-1.0), V(c.X+hl-4.0, 4.6, c.Y+1.0))) })
	g.Paint(Glass, Block(V(c.X+hl-4.4, 3.6, c.Y-0.8), V(c.X+hl-4.0, 4.2, c.Y+0.8)))
	g.With(Red, func() { g.Fill(Block(V(c.X-hl+1.6, 3.2, c.Y-0.5), V(c.X-hl+2.6, 4.0, c.Y+0.5))) })
}

// pier is Westminster Pier: a pontoon below the bridge with a gangway to the
// embankment and a smaller tour boat moored alongside. The cruise boat's wake trails
// behind it in the river.
func (b BigBen) pier(g *Grid) {
	g.With(Navy, func() { g.Fill(Block(V(50.5, -0.2, 14.5), V(57, 0.5, 17))) })
	g.With(Wood, func() { g.Fill(Block(V(50.5, 0.5, 14.5), V(57, 0.9, 17))) })
	g.With(White, func() { g.Fill(Block(V(52, 0.9, 15), V(55.5, 2.6, 16.6))) })
	g.With(Slate, func() { g.Fill(Block(V(51.6, 2.6, 14.7), V(55.9, 3.0, 16.9))) })
	g.Paint(Glass, Block(V(52, 1.5, 15), V(55.5, 2.1, 16.6)))
	g.With(Wood, func() { g.Fill(Block(V(53, 0.9, 13), V(54.2, 1.4, 14.5))) })
	b.boat(g, P(53, 19.4), 4.4)
	// Wake: a foam trail and a spreading V behind the cruise boat.
	c := P(-8-8.4, 27)
	g.Paint(White, Func{Box: Box{V(c.X-18, -0.2, c.Y-6), V(c.X, 0.2, c.Y+6)}, F: func(p Vec3) bool {
		d := c.X - p.X
		dz := math.Abs(p.Z - c.Y)
		trail := d < 8 && dz < 1.0 && Hash01(int(p.X*4), int(p.Z*4)) < 0.95-d*0.07
		v := math.Abs(dz-0.8-d*0.3) < 0.6 && Hash01(int(p.X*4), int(p.Z*4), 1) < 1.0-d*0.045
		return trail || v
	}})
}
