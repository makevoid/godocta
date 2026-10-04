package scenes

import (
	"math"

	"godocta/gen/kit"
	. "godocta/gen/octa"
)

// Rothenburg is Rothenburg ob der Tauber, the walled medieval town on its plateau above
// the Tauber valley. North is -Z. The market square with the white-towered town hall
// and St. Jakob's twin spires sits at the center; streets of half-timbered gabled houses
// radiate to the gate towers of the roofed town wall. In the south the Plönlein fork
// splits around its narrow yellow house, the left street passing under the Siebersturm,
// the right running down to the Kobolzeller gate; from there a lane descends the
// vineyard slope to the two-tiered double bridge over the river.
type Rothenburg struct{}

func init() { Register(Rothenburg{}) }

func (Rothenburg) Name() string  { return "rothenburg" }
func (Rothenburg) Title() string { return "Rothenburg" }

// rothenburgP is the plateau (street) level.
const rothenburgP = 6.0

// rothenburgWall is the town wall's outer face, a closed polygon.
var rothenburgWall = []Vec2{
	{-41, -8.1}, {-36.7, -25.4}, {-23.8, -36.2}, {0, -39.4}, {19.4, -36.2}, {33.5, -26.4},
	{40, -10.2}, {38.9, 6}, {30.2, 16.8}, {15.1, 23.2}, {-4.3, 24.3}, {-21.6, 20}, {-34.6, 10.3},
}

// rothenburgFrame places local solids: local +Z points along plan angle theta and the
// local origin lands on c.
type rothenburgFrame struct {
	c     Vec2
	theta float64
}

func (f rothenburgFrame) at(s Solid) Solid {
	return Translate(RotateY(s, P(0, 0), f.theta-math.Pi/2), f.c.At(0))
}

func (f rothenburgFrame) plan(x, z float64) Vec2 {
	return P(x, z).Rotate(f.theta - math.Pi/2).Add(f.c)
}

// rothenburgStreet is a straight street band of width w.
type rothenburgStreet struct {
	a, b Vec2
	w    float64
}

func (s rothenburgStreet) has(p Vec2, margin float64) bool {
	d := s.b.Sub(s.a)
	l := d.Length()
	u := d.Scale(1 / l)
	q := p.Sub(s.a)
	t := q.Dot(u)
	n := q.Dot(P(-u.Y, u.X))
	return t >= 0 && t <= l && math.Abs(n) < s.w/2+margin
}

// rothenburgHouse is one burgher house: facade on local z = 0 facing +Z, w wide and
// d deep, standing on the plateau.
type rothenburgHouse struct {
	f                           rothenburgFrame
	w, d                        float64
	floors                      int
	plaster, timber, roof, base Color
	gable, boxes                bool
	seed                        int
}

// rothenburgTown plans the house plots along the streets and the wall.
type rothenburgTown struct {
	r        Rothenburg
	streets  []rothenburgStreet
	reserved [][4]float64
	feet     [][]Vec2
	houses   []rothenburgHouse
	seed     int
}

func (Rothenburg) smooth(a, b, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-a)/(b-a)))
	return t * t * (3 - 2*t)
}

// wallInset is the distance inside the wall polygon (negative outside) and the arc
// position of the nearest wall point.
func (Rothenburg) wallInset(p Vec2) (float64, float64) {
	best, bestS, arc := math.Inf(1), 0.0, 0.0
	n := len(rothenburgWall)
	for i := 0; i < n; i++ {
		a, b := rothenburgWall[i], rothenburgWall[(i+1)%n]
		d := b.Sub(a)
		l := d.Length()
		t := math.Max(0, math.Min(1, p.Sub(a).Dot(d)/(l*l)))
		if dist := p.Sub(a.Add(d.Scale(t))).Length(); dist < best {
			best, bestS = dist, arc+t*l
		}
		arc += l
	}
	if !InsidePolygon(rothenburgWall, p) {
		best = -best
	}
	return best, bestS
}

// wallNormal is the outward unit normal of the wall edge nearest p.
func (Rothenburg) wallNormal(p Vec2) Vec2 {
	best, normal := math.Inf(1), P(0, 1)
	n := len(rothenburgWall)
	for i := 0; i < n; i++ {
		a, b := rothenburgWall[i], rothenburgWall[(i+1)%n]
		d := b.Sub(a)
		l := d.Length()
		t := math.Max(0, math.Min(1, p.Sub(a).Dot(d)/(l*l)))
		if dist := p.Sub(a.Add(d.Scale(t))).Length(); dist < best {
			best = dist
			normal = P(d.Y/l, -d.X/l)
			mid := a.Add(d.Scale(0.5))
			if InsidePolygon(rothenburgWall, mid.Add(normal)) {
				normal = normal.Scale(-1)
			}
		}
	}
	return normal
}

func (Rothenburg) riverZ(x float64) float64 { return 37.5 + 1.6*math.Sin(x/13) }

const rothenburgRiver = 3.2 // River half-width.

func (r Rothenburg) inSite(x, z float64) bool {
	if u, _ := r.wallInset(P(x, z)); u >= -7 {
		return true
	}
	a, b := x/44, (z-31)/14.2
	return a*a*a*a+b*b*b*b <= 1
}

// height is the terrain surface: the plateau inside the wall and a shelf outside it,
// steep slopes to a low rim, and in the south a long vineyard slope to the river.
func (r Rothenburg) height(x, z float64) float64 {
	u, _ := r.wallInset(P(x, z))
	if u >= -1.8 {
		return rothenburgP
	}
	d := -u - 1.8
	s := r.smooth(4, 16, z)
	k := 1.2 + (0.6-1.2)*s
	h := rothenburgP - d*k
	dz := z - r.riverZ(x)
	if math.Abs(dz) < rothenburgRiver {
		return 0.5
	}
	floor := 1.0
	if dz > 0 {
		floor = math.Min(1.8, 1.0+(dz-rothenburgRiver)*0.15)
	}
	return math.Max(h, floor)
}

func (r Rothenburg) Build(g *Grid) Order {
	town := r.plan()
	r.terrain(g, town)
	r.bridge(g)
	r.wall(g)
	for _, h := range town.houses {
		r.house(g, h)
	}
	r.church(g)
	r.rathaus(g)
	r.plonlein(g)
	r.gates(g)
	r.trees(g, town)
	return Sweep(P(0, -6), math.Pi/2)
}

// plan lays out the streets, reserved plots and house rows.
func (r Rothenburg) plan() *rothenburgTown {
	t := &rothenburgTown{r: r}
	t.streets = []rothenburgStreet{
		{P(-8, -10.5), P(6, -10.5), 11},    // Markt.
		{P(-1, -16), P(-1, -5), 14},        // Markt.
		{P(-8.5, -2.5), P(-41, -2.5), 4.4}, // Herrngasse to the Burgtor.
		{P(-1, -5), P(-1, 2.5), 4},         // Schmiedgasse.
		{P(-1, 2), P(13.5, 22.5), 3.6},     // Plönlein, left: Siebersturm.
		{P(-1, 2), P(-11.5, 24.6), 3.6},    // Plönlein, right: Kobolzeller Steige.
		{P(6, -10), P(41, -8), 4},          // Rödergasse to the Rödertor.
		{P(-26, -2.5), P(-26, -35), 3.6},   // Klingengasse.
		{P(20, -9.5), P(26.6, -34), 3.6},   // Galgengasse.
		{P(17.5, 4), P(38, 3), 3.2},        // Lane east.
	}
	t.reserved = [][4]float64{
		{-12, -36, 12.8, -19},  // St. Jakob.
		{-21.5, -18.5, -8, -5}, // Town hall.
		{3.2, 8.2, 10.4, 15.0}, // Siebersturm.
		{-2.4, 5.2, 2.6, 9.4},  // Plönlein fountain.
		{-3.0, 9.3, 3.3, 16.0}, // Plönlein house.
		{-3.4, 15.6, 3.6, 30},  // Its yard, open towards the valley.
	}
	// The yellow Plönlein house in the fork, gable to the north.
	t.houses = append(t.houses, rothenburgHouse{
		f: rothenburgFrame{P(0.15, 9.5), -math.Pi / 2}, w: 5.4, d: 6.2, floors: 3,
		plaster: Gold, timber: Wood, roof: Terracotta, base: Gold, gable: true, boxes: true, seed: 999,
	})
	for _, s := range t.streets {
		d := s.b.Sub(s.a)
		u := d.Scale(1 / d.Length())
		n := P(-u.Y, u.X)
		for _, side := range []float64{1, -1} {
			off := n.Scale(side * s.w / 2)
			t.row(s.a.Add(off), s.b.Add(off), n.Scale(-side), 0.75, 1)
		}
	}
	// Houses backing onto the lane behind the wall.
	m := len(rothenburgWall)
	for i := 0; i < m; i++ {
		a, b := rothenburgWall[i], rothenburgWall[(i+1)%m]
		in := r.wallNormal(a.Add(b).Scale(0.5)).Scale(-1)
		t.row(a.Add(in.Scale(3.5)), b.Add(in.Scale(3.5)), in.Scale(-1), 0.8, 2)
	}
	// Back buildings filling the remaining yards.
	for z := -38.0; z <= 23; z += 0.5 {
		for x := -40.0; x <= 39; x += 0.5 {
			p := P(x, z)
			if !t.free(p) {
				continue
			}
			t.seed++
			k := t.seed
			sizes := [][2]float64{{4.4 + 1.6*Hash01(k, 1), 4.6 + 2.4*Hash01(k, 2)}, {4.6, 5.0}, {4.2, 4.4}}
		fill:
			for _, wd := range sizes {
				w, d := wd[0], wd[1]
				for o := 0; o < 4; o++ {
					theta := float64((o+int(Hash01(k, 3)*4))%4) * math.Pi / 2
					f := rothenburgFrame{p, theta}
					if t.fits(f, w, d) {
						t.feet = append(t.feet, []Vec2{f.plan(-w/2, 0), f.plan(w/2, 0), f.plan(w/2, -d), f.plan(-w/2, -d)})
						h := r.style(f, w, d, k, Hash01(k, 4) < 0.5)
						if h.floors > 2 {
							h.floors = 2
						}
						t.houses = append(t.houses, h)
						break fill
					}
				}
			}
		}
	}
	return t
}

func (t *rothenburgTown) free(p Vec2) bool {
	if u, _ := t.r.wallInset(p); u < 3.4 {
		return false
	}
	for _, s := range t.streets {
		if s.has(p, -0.2) {
			return false
		}
	}
	for _, b := range t.reserved {
		if p.X > b[0] && p.X < b[2] && p.Y > b[1] && p.Y < b[3] {
			return false
		}
	}
	for _, f := range t.feet {
		if InsidePolygon(f, p) {
			return false
		}
	}
	return true
}

func (t *rothenburgTown) fits(f rothenburgFrame, w, d float64) bool {
	nx, nz := int(math.Ceil(w/0.9)), int(math.Ceil(d/0.9))
	for i := 0; i <= nx; i++ {
		for j := 0; j <= nz; j++ {
			x := -w/2 + 0.15 + (w-0.3)*float64(i)/float64(nx)
			z := -0.15 - (d-0.3)*float64(j)/float64(nz)
			if !t.free(f.plan(x, z)) {
				return false
			}
		}
	}
	return true
}

// row lines houses along the facade line a→b, their facades facing the unit normal.
func (t *rothenburgTown) row(a, b, facing Vec2, gableOdds float64, minFloors int) {
	d := b.Sub(a)
	l := d.Length()
	dir := d.Scale(1 / l)
	theta := facing.Angle()
	for s := 0.2; s < l-4.2; {
		t.seed++
		k := t.seed
		w := 4.4 + 1.8*Hash01(k, 1)
		if s+w > l {
			w = l - s
		}
		placed := false
		for _, depth := range []float64{5.8 + 2*Hash01(k, 2), 5.2, 4.4} {
			f := rothenburgFrame{a.Add(dir.Scale(s + w/2)), theta}
			if !t.fits(f, w, depth) {
				continue
			}
			t.feet = append(t.feet, []Vec2{f.plan(-w/2, 0), f.plan(w/2, 0), f.plan(w/2, -depth), f.plan(-w/2, -depth)})
			h := t.r.style(f, w, depth, k, Hash01(k, 3) < gableOdds)
			if h.floors < minFloors && depth > 5.5 {
				h.floors = minFloors
			}
			t.houses = append(t.houses, h)
			placed = true
			break
		}
		if placed {
			s += w
		} else {
			s += 1
		}
	}
}

// style picks a house's storeys and colors deterministically.
func (Rothenburg) style(f rothenburgFrame, w, d float64, k int, gable bool) rothenburgHouse {
	h := rothenburgHouse{f: f, w: w, d: d, gable: gable, seed: k, boxes: Hash01(k, 4) < 0.6}
	switch x := Hash01(k, 5); {
	case x < 0.25:
		h.floors = 1
	case x < 0.85:
		h.floors = 2
	default:
		h.floors = 3
	}
	if d < 5.5 {
		h.floors = 1
	}
	switch x := Hash01(k, 6); {
	case x < 0.28:
		h.plaster, h.timber = White, Wood
	case x < 0.4:
		h.plaster, h.timber = White, Charcoal
	case x < 0.58:
		h.plaster, h.timber = Sand, Wood
	case x < 0.74:
		h.plaster, h.timber = Gold, Wood
	case x < 0.82:
		h.plaster, h.timber = Terracotta, Charcoal
	case x < 0.88:
		h.plaster, h.timber = Red, Charcoal
	default:
		h.plaster, h.timber = Stone, Wood
	}
	h.roof = Brick
	if Hash01(k, 7) < 0.35 {
		h.roof = Terracotta
	}
	h.base = h.plaster
	if Hash01(k, 8) < 0.45 {
		h.base = Sand
		if h.plaster == Sand {
			h.base = Stone
		}
	}
	return h
}

// terrain lays the hill shell, river, vineyards, paths and the paved streets.
func (r Rothenburg) terrain(g *Grid, town *rothenburgTown) {
	box := Box{V(-47, -1, -45), V(47, rothenburgP+0.5, 48)}
	g.Use(Grass).Fill(Func{box, func(p Vec3) bool {
		if !r.inSite(p.X, p.Z) {
			return false
		}
		h := r.height(p.X, p.Z)
		if p.Y > h || p.Y < h-4 {
			return false
		}
		gx := r.height(p.X+0.5, p.Z) - r.height(p.X-0.5, p.Z)
		gz := r.height(p.X, p.Z+0.5) - r.height(p.X, p.Z-0.5)
		return p.Y >= h-0.65-0.75*math.Hypot(gx, gz)
	}})
	// River.
	g.Paint(Water, Func{box, func(p Vec3) bool {
		u, _ := r.wallInset(p.Plan())
		return u < -1.8 && math.Abs(p.Z-r.riverZ(p.X)) < rothenburgRiver
	}})
	// Vineyard terraces on the valley slope: contour rows of vines.
	g.Paint(Foliage, Func{box, func(p Vec3) bool {
		u, _ := r.wallInset(p.Plan())
		if p.Z < 6 || u > -3 || math.Abs(p.Z-r.riverZ(p.X)) < rothenburgRiver+1.5 {
			return false
		}
		if p.X > -22 && p.X < -7 {
			return false // Meadow around the lane to the bridge.
		}
		h := r.height(p.X, p.Z)
		return h > 1.6 && math.Mod(-u, 1.6) < 0.8
	}})
	// Lane from the Kobolzeller gate down to the double bridge.
	lane := []Vec2{P(-11.2, 24.8), P(-13.5, 25.3), P(-15, 26)}
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		q := p.Plan()
		for i := 0; i+1 < len(lane); i++ {
			a, b := lane[i], lane[i+1]
			d := b.Sub(a)
			t := math.Max(0, math.Min(1, q.Sub(a).Dot(d)/d.Dot(d)))
			if q.Sub(a.Add(d.Scale(t))).Length() < 1.3 {
				return true
			}
		}
		return false
	}})
	// Cobbled streets, market and the lane behind the wall.
	paved := func(p Vec3) bool {
		if p.Y < rothenburgP-1.5 {
			return false
		}
		q := p.Plan()
		if u, _ := r.wallInset(q); u >= 1 && u < 3.5 {
			return true
		}
		for _, s := range town.streets {
			if s.has(q, 0.3) {
				return true
			}
		}
		return false
	}
	g.Paint(Stone, Func{box, paved})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		return paved(p) && Hash01(int(math.Floor(p.X*1.3)), int(math.Floor(p.Z*1.3)), 11) < 0.22
	}})
}

// wall builds the stone town wall with its roofed timber gallery and loopholes.
func (r Rothenburg) wall(g *Grid) {
	y0 := rothenburgP
	box := Box{V(-40, y0-2, -39), V(39, y0+9, 24)}
	g.Use(Stone).Fill(Func{box, func(p Vec3) bool {
		u, _ := r.wallInset(p.Plan())
		return u >= 0 && u <= 1.15 && p.Y >= y0-1 && p.Y <= y0+5.2
	}})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		u, s := r.wallInset(p.Plan())
		return u >= -0.5 && u <= 1.6 && Hash01(int(math.Floor(s*1.2)), int(math.Floor(p.Y*1.6)), 21) < 0.18
	}})
	// Loopholes under the gallery.
	g.Carve(Func{box, func(p Vec3) bool {
		u, s := r.wallInset(p.Plan())
		return u >= -0.4 && u <= 1.5 && p.Y >= y0+3.4 && p.Y <= y0+4.4 && math.Mod(s, 3.2) < 0.75
	}})
	g.Use(Wood).Fill(Func{box, func(p Vec3) bool {
		u, s := r.wallInset(p.Plan())
		switch {
		case u < 1.15 || u > 3.1:
			return false
		case p.Y >= y0+2.6 && p.Y <= y0+3.2:
			return true // Walkway.
		}
		return u >= 2.3 && p.Y >= y0 && p.Y <= y0+5.4 && math.Mod(s+1.6, 3.2) < 0.8 // Posts.
	}})
	g.Use(Terracotta).Fill(Func{box, func(p Vec3) bool {
		u, _ := r.wallInset(p.Plan())
		if u < -0.4 || u > 3.4 {
			return false
		}
		ys := y0 + 6.6 - u*0.5
		return p.Y <= ys && p.Y >= ys-0.8
	}})
}

// opening is a rectangular window across X centred on x, through Z from z0 to z1.
func (Rothenburg) opening(x, hw, y0, y1, z0, z1 float64) Solid {
	return Block(V(x-hw, y0, z0), V(x+hw, y1, z1))
}

// openingX is a rectangular window across Z centred on z, through X from x0 to x1.
func (Rothenburg) openingX(z, hw, y0, y1, x0, x1 float64) Solid {
	return Block(V(x0, y0, z-hw), V(x1, y1, z+hw))
}

func (r Rothenburg) house(g *Grid, h rothenburgHouse) {
	at := h.f.at
	hw, d := h.w/2, h.d
	const (
		gf  = 3.0  // Ground floor height.
		fh  = 2.6  // Upper floor height.
		t   = 1.0  // Wall thickness.
		k   = 1.45 // Roof pitch.
		jet = 0.5  // Upper-floor jetty over the street.
		ov  = 0.4  // Roof overhang.
	)
	y0 := rothenburgP
	eave := y0 + gf + float64(h.floors)*fh
	th := 0.85 * math.Sqrt(1+k*k)

	g.With(h.base, func() { g.Fill(at(Block(V(-hw, y0-0.6, -d), V(hw, y0+gf, -jet)))) })
	g.With(h.plaster, func() { g.Fill(at(Block(V(-hw, y0+gf, -d), V(hw, eave, 0)))) })
	inner := Block(V(-hw+t, y0+0.2, -d+t), V(hw-t, eave+10, -jet-t))
	g.Carve(at(inner))

	// Roof and gables; ridge(x, z) is the roof surface.
	var ridge func(x, z float64) float64
	if h.gable {
		ridge = func(x, z float64) float64 { return eave + k*(hw+ov-math.Abs(x)) }
		g.With(h.roof, func() {
			g.Fill(at(Func{Box{V(-hw-ov, eave-1.5, -d-0.3), V(hw+ov, eave+k*(hw+ov), 0.4)}, func(p Vec3) bool {
				y := ridge(p.X, p.Z)
				return p.Y <= y && p.Y >= y-th
			}}))
		})
		g.With(h.plaster, func() {
			for _, z := range [][2]float64{{-t, 0}, {-d, -d + t}} {
				g.Fill(at(Func{Box{V(-hw, eave, z[0]), V(hw, eave+k*hw, z[1])}, func(p Vec3) bool {
					return p.Y <= ridge(p.X, p.Z)-1.0
				}}))
			}
		})
	} else {
		ridge = func(x, z float64) float64 { return eave + k*(d/2+ov-math.Abs(z+d/2)) }
		g.With(h.roof, func() {
			g.Fill(at(Func{Box{V(-hw-0.3, eave-1.5, -d-ov), V(hw+0.3, eave+k*(d/2+ov), ov)}, func(p Vec3) bool {
				y := ridge(p.X, p.Z)
				return p.Y <= y && p.Y >= y-th
			}}))
		})
		g.With(h.plaster, func() {
			for _, x := range [][2]float64{{-hw, -hw + t}, {hw - t, hw}} {
				g.Fill(at(Func{Box{V(x[0], eave, -d), V(x[1], eave+k*d/2, 0)}, func(p Vec3) bool {
					return p.Y <= ridge(p.X, p.Z)-1.0
				}}))
			}
		})
		// Cross gables (Zwerchhäuser) rising from the facade through the eaves.
		n := 1
		if h.w > 6.4 {
			n = 2
		}
		for i := 0; i < n; i++ {
			x := 0.0
			if n == 2 {
				x = -hw/2 + float64(i)*hw
			}
			const cw = 1.3
			top := eave + 2.4
			cg := func(px float64) float64 { return top + k*(cw+ov-math.Abs(px-x)) }
			g.With(h.plaster, func() {
				g.Fill(at(Block(V(x-cw, eave-0.5, -3.4), V(x+cw, top, 0))))
				g.Fill(at(Func{Box{V(x-cw, top, -1), V(x+cw, top+k*cw, 0)}, func(p Vec3) bool { return p.Y <= cg(p.X)-1.0 }}))
			})
			g.With(h.roof, func() {
				g.Fill(at(Func{Box{V(x-cw-ov, top-1.5, -4.2), V(x+cw+ov, top+k*(cw+ov), 0.4)}, func(p Vec3) bool {
					y := cg(p.X)
					return p.Y <= y && p.Y >= y-th && p.Y >= ridge(p.X, p.Z)-th
				}}))
			})
			g.Paint(h.timber, at(Func{Box{V(x-cw, eave-0.45, -0.45), V(x+cw, top+k*cw, 0.2)}, func(p Vec3) bool {
				return math.Abs(p.X-x) >= cw-0.45 || math.Abs(p.Y-top) < 0.25 || math.Abs(p.Y-eave) < 0.25
			}}))
			g.Carve(at(r.opening(x, 0.5, eave+0.5, eave+1.9, -1.6, 0.3)))
		}
	}

	// Interior faces dark, so openings read as windows.
	g.Paint(Charcoal, at(Block(V(-hw+t-0.2, y0-0.3, -d+t-0.2), V(hw-t+0.2, eave-0.3, -jet-t+0.2))))

	// Timber frame on the upper storeys of the front facade (and its gable).
	n := int(math.Max(2, math.Round(h.w/2.3)))
	bay := h.w / float64(n)
	yb := y0 + gf
	brace := func(j, floor int) bool { return n >= 3 && (j+floor+h.seed)%2 == 1 && j != 0 && j != n-1 }
	frameTop := eave
	if h.gable {
		frameTop = eave + k*hw
	}
	for _, face := range [][2]float64{{-0.45, 0.2}, {-d - 0.2, -d + 0.45}} {
		g.Paint(h.timber, at(Func{Box{V(-hw, yb, face[0]), V(hw, frameTop, face[1])}, func(p Vec3) bool {
			x, y := p.X, p.Y-yb
			if p.Y > eave-0.5 && p.Y > ridge(p.X, p.Z)-1.0 {
				return false
			}
			if math.Abs(x) >= hw-0.5 || math.Mod(y, fh) < 0.45 || (p.Y >= eave-0.45 && p.Y <= eave+0.1) {
				return true
			}
			if p.Y > eave {
				// Gable: king post and collar rail.
				return math.Abs(x) < 0.3 || math.Abs(p.Y-(eave+2.2)) < 0.25
			}
			j := int(math.Floor((x + hw) / bay))
			xc := -hw + bay*(float64(j)+0.5)
			if math.Abs(math.Abs(x-xc)-bay/2) < 0.28 {
				return true // Posts between bays.
			}
			floor := int(math.Floor(y / fh))
			if !brace(j, floor) {
				return false
			}
			yc := float64(floor)*fh + fh/2 + 0.2
			bw, bh := bay/2-0.25, fh/2-0.25
			return math.Abs(math.Abs(x-xc)*bh-math.Abs(y-yc)*bw)/math.Hypot(bw, bh) < 0.3
		}}))
	}

	// Upper windows front and back, with flower boxes.
	for i := 0; i < h.floors; i++ {
		sill := yb + float64(i)*fh + 0.85
		for j := 0; j < n; j++ {
			x := -hw + bay*(float64(j)+0.5)
			if !brace(j, i) {
				g.Carve(at(r.opening(x, 0.55, sill, sill+1.45, -1.6, 0.3)))
				if h.boxes && Hash01(h.seed, i, j, 9) < 0.6 {
					box := Red
					if h.plaster == Red {
						box = Foliage
					}
					g.With(box, func() { g.Fill(at(Block(V(x-0.75, sill-0.6, 0.1), V(x+0.75, sill-0.05, 0.75)))) })
				}
			}
			if !brace(j, i) {
				g.Carve(at(r.opening(x, 0.55, sill, sill+1.45, -d-0.3, -d+1.3)))
			}
		}
	}
	// Gable windows.
	if h.gable {
		for _, z := range [][2]float64{{-1.6, 0.3}, {-d - 0.3, -d + 1.6}} {
			for _, x := range []float64{-1.1, 1.1} {
				g.Carve(at(r.opening(x, 0.45, eave+0.6, eave+1.9, z[0], z[1])))
			}
			g.Carve(at(r.opening(0, 0.4, eave+2.8, eave+3.8, z[0], z[1])))
		}
	}

	// Ground floor: an arched door and shop windows.
	door := -hw + bay*0.5
	if Hash01(h.seed, 10) < 0.5 {
		door = hw - bay*0.5
	}
	g.Carve(at(Arch{V(door, y0+0.2, -jet-0.5), 0.7, 1.5, 1.8, AxisZ}))
	g.With(Wood, func() { g.Fill(at(Block(V(door-0.8, y0+0.2, -jet-1.3), V(door+0.8, y0+2.6, -jet-0.9)))) })
	for j := 0; j < n; j++ {
		x := -hw + bay*(float64(j)+0.5)
		if math.Abs(x-door) > 0.1 {
			g.Carve(at(r.opening(x, 0.6, y0+1.0, y0+2.4, -jet-1.4, -jet+0.3)))
		}
	}

	// Chimney.
	if Hash01(h.seed, 12) < 0.4 {
		cx, cz := hw*0.45, -d*0.6
		g.With(Brick, func() { g.Fill(at(Block(V(cx-0.5, eave, cz-0.5), V(cx+0.5, ridge(cx, cz)+1.3, cz+0.5)))) })
	}
}

// lancet is a pointed opening across X centred on x, through Z from z0 to z1.
func (Rothenburg) lancet(x, hw, sill, spring, z0, z1 float64) Func {
	apex := spring + math.Sqrt(3)*hw
	return Func{Box{V(x-hw, sill, z0), V(x+hw, apex, z1)}, func(p Vec3) bool {
		ax := math.Abs(p.X - x)
		if ax > hw {
			return false
		}
		rise := p.Y - spring
		if rise <= 0 {
			return true
		}
		d := ax + hw
		return d*d+rise*rise <= 4*hw*hw
	}}
}

// pyramid is a steep hipped tower roof slab centred on the local origin.
func (Rothenburg) pyramid(half, base, k float64) Func {
	return Func{Box{V(-half, base, -half), V(half, base+k*half, half)}, func(p Vec3) bool {
		ys := base + k*(half-math.Max(math.Abs(p.X), math.Abs(p.Z)))
		return p.Y <= ys && p.Y >= ys-1.6
	}}
}

// church is St. Jakob: a Gothic nave with buttresses and lancets, a narrow choir with
// an apse, and twin spired towers flanking the choir.
func (r Rothenburg) church(g *Grid) {
	y0 := rothenburgP
	x0, x1, z0, z1 := -10.5, 3.5, -30.0, -22.0
	eave := y0 + 12
	zc := (z0 + z1) / 2
	const k = 1.5
	half := (z1-z0)/2 + 0.4
	g.Use(Sand).Fill(Block(V(x0, y0-0.5, z0), V(x1, eave, z1)))
	g.Carve(Block(V(x0+1.2, y0+0.2, z0+1.2), V(x1-1.2, eave+12, z1-1.2)))
	nave := func(p Vec3) float64 { return eave + k*(half-math.Abs(p.Z-zc)) }
	g.With(Brick, func() {
		g.Fill(Func{Box{V(x0-0.4, eave-2, z0-0.4), V(x1, eave+k*half, z1+0.4)}, func(p Vec3) bool {
			y := nave(p)
			return p.Y <= y && p.Y >= y-1.6
		}})
	})
	g.Fill(Func{Box{V(x0, eave, z0), V(x0+1.2, eave+k*half, z1)}, func(p Vec3) bool { return p.Y <= nave(p)-1.0 }})
	// Buttresses and lancet windows on both long sides.
	for x := x0; x <= x1+0.1; x += 4 {
		for _, s := range []float64{-1, 1} {
			face := zc + s*(z1-z0)/2
			g.Fill(Func{Box{V(x-0.6, y0-0.5, math.Min(face, face+s*1.6)), V(x+0.6, eave-1, math.Max(face, face+s*1.6))}, func(p Vec3) bool {
				out := math.Abs(p.Z - face)
				return p.Y <= eave-1-out*1.6
			}})
		}
		if x+2 < x1 {
			g.Carve(r.lancet(x+2, 0.85, y0+3.5, y0+8.5, z0-0.5, z0+1.5), r.lancet(x+2, 0.85, y0+3.5, y0+8.5, z1-1.5, z1+0.5))
		}
	}
	// West front: portal and great window, through X.
	swap := func(f Func) Func {
		b := f.Box
		return Func{Box{V(b.Min.Z, b.Min.Y, b.Min.X), V(b.Max.Z, b.Max.Y, b.Max.X)}, func(p Vec3) bool { return f.Contains(V(p.Z, p.Y, p.X)) }}
	}
	g.Carve(swap(r.lancet(zc, 1.2, y0+0.2, y0+2.4, x0-0.5, x0+1.7)), swap(r.lancet(zc, 1.5, y0+5, y0+9, x0-0.5, x0+1.7)))
	g.Fill(swap(r.lancet(zc, 1.2, y0+0.2, y0+2.4, x0+0.9, x0+1.3)))
	g.Paint(Wood, swap(r.lancet(zc, 1.2, y0+0.2, y0+2.4, x0+0.9, x0+1.3)))

	// Choir and apse.
	c0, c1 := zc-2.5, zc+2.5
	ce := x1 + 6
	g.Fill(Block(V(x1, y0-0.5, c0), V(ce, eave, c1)))
	g.Fill(Intersect(Cylinder{Center: P(ce, zc), Radius: 2.5, Y0: y0 - 0.5, Y1: eave}, Block(V(ce, y0-1, c0), V(ce+3, eave+1, c1))))
	g.Carve(Block(V(x1-1.2, y0+0.2, c0+1.1), V(ce, eave+8, c1-1.1)), Intersect(Cylinder{Center: P(ce, zc), Radius: 1.4, Y0: y0 + 0.2, Y1: eave + 8}, Block(V(ce-0.1, y0, c0), V(ce+3, eave+9, c1))))
	choir := func(p Vec3) float64 { return eave + k*(2.9-math.Abs(p.Z-zc)) }
	g.With(Brick, func() {
		g.Fill(Func{Box{V(x1-1, eave-2, c0-0.4), V(ce, eave+k*2.9, c1+0.4)}, func(p Vec3) bool {
			y := choir(p)
			return p.Y <= y && p.Y >= y-1.6
		}})
		g.Fill(Intersect(Frustum{P(ce, zc), 2.9, 0, eave, eave + k*2.9}, Block(V(ce, eave, c0-1), V(ce+3, eave+5, c1+1))))
	})
	for _, a := range []float64{-0.6, 0, 0.6} {
		g.Carve(Translate(RotateY(Block(V(1.5, y0+3.5, -0.45), V(3, y0+8.5, 0.45)), P(0, 0), a), V(ce, 0, zc)))
	}
	g.Carve(r.lancet(ce-1.6, 0.7, y0+3.5, y0+8, c0-0.5, c0+1.5), r.lancet(ce-1.6, 0.7, y0+3.5, y0+8, c1-1.5, c1+0.5))

	// Twin towers flanking the choir, with belfries and slate spires.
	for _, tz := range []float64{c0 - 1.8, c1 + 1.8} {
		tc := P(x1+3, tz)
		top := y0 + 21.0
		g.Use(Sand).Fill(BlockAt(tc, 3.6, 3.6, y0-0.5, top))
		g.Fill(BlockAt(tc, 4.3, 4.3, top-0.7, top))
		g.Carve(BlockAt(tc, 1.6, 1.6, y0+0.2, top-0.3))
		for _, y := range []float64{y0 + 9.5, y0 + 15} {
			g.Carve(r.lancet(tc.X, 0.55, y, y+2.6, tc.Y-2.5, tc.Y+2.5), swap(r.lancet(tc.Y, 0.55, y, y+2.6, tc.X-2.5, tc.X+2.5)))
		}
		g.With(Slate, func() { g.FillVolume(Frustum{tc, 2.0, 0.25, top, top + 12}) })
		g.With(Gold, func() { g.Fill(Cylinder{Center: tc, Radius: 0.5, Y0: top + 11.5, Y1: top + 13.2}) })
	}
	g.Paint(Charcoal, Block(V(x0+1, y0-0.3, z0+1), V(x1-1, eave-0.3, z1-1)), Block(V(x1-1.3, y0-0.3, c0+0.9), V(ce+1.6, eave-0.3, c1-0.9)))
}

// rathaus is the town hall on the market's west side: a white Gothic hall with a steep
// roof, the Renaissance arcade towards the square and the tall white tower.
func (r Rothenburg) rathaus(g *Grid) {
	y0 := rothenburgP
	x0, x1, z0, z1 := -21.0, -9.0, -18.0, -5.0
	eave := y0 + 9.0
	xc := (x0 + x1) / 2
	half := (x1-x0)/2 + 0.4
	const k = 1.3
	top := func(p Vec3) float64 { return eave + k*(half-math.Abs(p.X-xc)) }
	g.Use(White).Fill(Block(V(x0, y0-0.5, z0), V(x1, eave, z1)))
	g.Carve(Block(V(x0+1.1, y0+0.2, z0+1.1), V(x1-1.1, eave+20, z1-1.1)))
	g.With(Brick, func() {
		g.Fill(Func{Box{V(x0-0.4, eave-2, z0-0.4), V(x1+0.4, eave+k*half, z1+0.4)}, func(p Vec3) bool {
			y := top(p)
			return p.Y <= y && p.Y >= y-1.6
		}})
	})
	g.Fill(Func{Box{V(x0, eave, z0), V(x1, eave+k*half, z0+1.1)}, func(p Vec3) bool { return p.Y <= top(p)-1.0 }})
	g.Fill(Func{Box{V(x0, eave, z1-1.1), V(x1, eave+k*half, z1)}, func(p Vec3) bool { return p.Y <= top(p)-1.0 }})
	g.Paint(Stone, Block(V(x0-0.2, y0-0.5, z0-0.2), V(x1+0.2, y0+0.8, z1+0.2)))
	// Windows on the long sides and the south gable.
	for i := 0; i < 2; i++ {
		y := y0 + 3.4 + float64(i)*2.9
		for j := 0; j < 5; j++ {
			z := z0 + 1.8 + float64(j)*(z1-z0-3.6)/4
			g.Carve(r.openingX(z, 0.55, y, y+1.8, x1-1.5, x1+0.3), r.openingX(z, 0.55, y, y+1.8, x0-0.3, x0+1.5))
		}
		for j := 0; j < 4; j++ {
			x := x0 + 2.2 + float64(j)*(x1-x0-4.4)/3
			g.Carve(r.opening(x, 0.55, y, y+1.8, z1-1.5, z1+0.3))
		}
	}
	for _, x := range []float64{xc - 2, xc + 2} {
		g.Carve(r.opening(x, 0.5, eave+1, eave+2.6, z1-1.5, z1+0.3))
	}
	g.Carve(r.opening(xc, 0.5, eave+4, eave+5.5, z1-1.5, z1+0.3))
	g.Paint(Charcoal, Block(V(x0+0.9, y0-0.3, z0+0.9), V(x1-0.9, eave-0.3, z1-0.9)))
	// Arcade porch on the market.
	g.Use(Sand).Fill(Block(V(x1, y0-0.3, -14.5), V(x1+2.8, y0+4.4, -7.5)))
	g.Carve(Block(V(x1, y0+0.2, -13.7), V(x1+1.9, y0+3.4, -8.3)))
	for _, z := range []float64{-12.3, -11, -9.7} {
		g.Carve(Arch{V(x1+2.4, y0+0.2, z), 0.55, 1.8, 1.4, AxisX})
	}
	g.Carve(Arch{V(x1+1, y0+0.2, -14.1), 0.6, 1.8, 1.4, AxisZ}, Arch{V(x1+1, y0+0.2, -7.9), 0.6, 1.8, 1.4, AxisZ})
	g.Paint(Charcoal, Block(V(x1-0.3, y0-0.3, -13.9), V(x1+0.4, y0+3.6, -8.1)))
	g.Fill(Block(V(x1, y0+4.4, -14.5), V(x1+2.8, y0+5.2, -14)), Block(V(x1, y0+4.4, -8), V(x1+2.8, y0+5.2, -7.5)), Block(V(x1+2.3, y0+4.4, -14.5), V(x1+2.8, y0+5.2, -7.5)))
	g.Paint(Red, Block(V(x1+2.3, y0+4.4, -14), V(x1+2.8, y0+5.2, -8)))

	// The tower rising from the north end.
	tc := P(xc, z0+2.4)
	t := y0 + 30.0
	g.Use(White).Fill(BlockAt(tc, 4.4, 4.4, eave, t))
	g.Carve(BlockAt(tc, 2.2, 2.2, eave, t-0.3))
	g.Paint(Charcoal, BlockAt(tc, 2.6, 2.6, eave, t))
	for y := eave + 3.0; y < t-3; y += 4 {
		g.Carve(r.opening(tc.X, 0.4, y, y+1.5, tc.Y-3, tc.Y+3), r.openingX(tc.Y, 0.4, y, y+1.5, tc.X-3, tc.X+3))
	}
	g.Use(Stone).Fill(BlockAt(tc, 5.4, 5.4, t-0.8, t))
	g.Use(Charcoal).Fill(Subtract(BlockAt(tc, 5.4, 5.4, t, t+1.1), BlockAt(tc, 4.2, 4.2, t-1, t+2)))
	g.Use(White).Fill(Cylinder{Center: tc, Radius: 1.6, Y0: t, Y1: t + 3})
	for _, a := range []float64{0, math.Pi / 2} {
		g.Carve(Translate(RotateY(Arch{V(0, t+0.6, 0), 0.5, 1.2, 4, AxisZ}, P(0, 0), a), tc.At(0)))
	}
	g.Use(Charcoal).FillVolume(Frustum{tc, 1.9, 0.3, t + 3, t + 6.5})
	g.Use(Gold).Fill(Cylinder{Center: tc, Radius: 0.5, Y0: t + 6.3, Y1: t + 8})
	// St. George's fountain in the market.
	fc := P(3, -13.5)
	g.Use(Stone).Fill(Cylinder{Center: fc, Radius: 1.8, Inner: 1.1, Y0: y0, Y1: y0 + 1.1}, Cylinder{Center: fc, Radius: 0.5, Y0: y0, Y1: y0 + 3})
	g.Paint(Water, Cylinder{Center: fc, Radius: 1.1, Y0: y0 - 0.5, Y1: y0 + 0.2})
	g.Use(Gold).Fill(Cylinder{Center: fc, Radius: 0.6, Y0: y0 + 3, Y1: y0 + 3.8})
}

// plonlein adds the little fountain in the fork in front of the yellow house.
func (Rothenburg) plonlein(g *Grid) {
	y0 := rothenburgP
	fc := P(0.1, 7.2)
	g.Use(Stone).Fill(Cylinder{Center: fc, Radius: 1.3, Inner: 0.7, Y0: y0, Y1: y0 + 1.0}, Cylinder{Center: fc, Radius: 0.45, Y0: y0, Y1: y0 + 2.4})
	g.Paint(Water, Cylinder{Center: fc, Radius: 0.7, Y0: y0 - 0.5, Y1: y0 + 0.2})
	g.Use(Red).Fill(Cylinder{Center: fc, Radius: 0.7, Y0: y0 + 2.4, Y1: y0 + 3.0})
}

// squareTower is a gate or wall tower in frame f: a hollow masonry shaft with small
// windows, a corbelled cornice and a steep hipped roof. Gates pass along local Z.
func (r Rothenburg) squareTower(g *Grid, f rothenburgFrame, half, top, roofK float64, wall, roof Color, gate bool, foot float64) {
	at := f.at
	y0 := rothenburgP
	g.With(wall, func() {
		g.Fill(at(Block(V(-half, y0-foot, -half), V(half, top, half))))
		g.Fill(at(Block(V(-half-0.45, top-0.9, -half-0.45), V(half+0.45, top, half+0.45))))
	})
	g.Carve(at(Block(V(-half+1.1, y0+0.3, -half+1.1), V(half-1.1, top-0.4, half-1.1))))
	for y := y0 + 5.5; y < top-2.4; y += 3.6 {
		for k := 0; k < 4; k++ {
			g.Carve(at(RotateY(Block(V(-0.45, y, half-1.5), V(0.45, y+1.5, half+0.3)), P(0, 0), float64(k)*math.Pi/2)))
		}
	}
	if gate {
		g.Carve(at(Arch{V(0, y0+0.3, 0), 1.5, 2.3, 2*half + 9, AxisZ}))
	}
	g.Paint(Charcoal, at(Block(V(-half+0.9, y0-0.1, -half+0.9), V(half-0.9, top+0.5, half-0.9))))
	g.With(roof, func() { g.Fill(at(r.pyramid(half+0.7, top, roofK))) })
	apex := top + roofK*(half+0.7)
	g.With(Gold, func() { g.Fill(at(Cylinder{Center: P(0, 0), Radius: 0.5, Y0: apex - 0.6, Y1: apex + 1.2})) })
}

// roundTower is a wall tower with a conical tiled roof.
func (Rothenburg) roundTower(g *Grid, c Vec2, rad, top, roofH float64, roof Color) {
	y0 := rothenburgP
	g.Use(Stone).Fill(Cylinder{Center: c, Radius: rad, Y0: y0 - 4, Y1: top})
	g.Fill(Frustum{c, rad, rad + 0.5, top - 1.4, top - 0.6}, Cylinder{Center: c, Radius: rad + 0.5, Y0: top - 0.6, Y1: top})
	g.Carve(Cylinder{Center: c, Radius: rad - 1.1, Y0: y0 + 0.3, Y1: top - 0.4})
	for i, y := range []float64{y0 + 2, y0 + 5.5} {
		for k := 0; k < 5; k++ {
			a := float64(k)*2*math.Pi/5 + float64(i)*0.6
			g.Carve(Translate(RotateY(Block(V(rad-1.6, y, -0.4), V(rad+0.3, y+1.4, 0.4)), P(0, 0), a), c.At(0)))
		}
	}
	g.Paint(Charcoal, Cylinder{Center: c, Radius: rad - 0.9, Y0: y0 - 0.1, Y1: top})
	g.Use(roof).FillVolume(Frustum{c, rad + 0.9, 0, top - 0.2, top + roofH})
	g.Use(Gold).Fill(Cylinder{Center: c, Radius: 0.5, Y0: top + roofH - 0.6, Y1: top + roofH + 1})
}

// onWall returns a frame centred on the wall near p, local +Z pointing out of town.
func (r Rothenburg) onWall(p Vec2) rothenburgFrame {
	n := r.wallNormal(p)
	u, _ := r.wallInset(p)
	return rothenburgFrame{p.Add(n.Scale(u - 0.65)), n.Angle()}
}

// gates places the gate towers, the Siebersturm and the wall towers.
func (r Rothenburg) gates(g *Grid) {
	y0 := rothenburgP
	r.squareTower(g, r.onWall(P(-39, -2.5)), 2.6, y0+15, 1.9, Stone, Terracotta, true, 2)  // Burgtor.
	r.squareTower(g, r.onWall(P(-26, -34.2)), 2.4, y0+14, 2.0, Stone, Brick, true, 2)      // Klingentor.
	r.squareTower(g, r.onWall(P(26, -31.5)), 2.4, y0+13, 1.8, Stone, Terracotta, true, 2)  // Galgentor.
	r.squareTower(g, r.onWall(P(-10.6, 22.7)), 2.3, y0+11, 1.8, Sand, Terracotta, true, 2) // Kobolzeller Tor.
	// Rödertor with its outer bastion and two round guard turrets.
	roder := r.onWall(P(39.8, -8))
	r.squareTower(g, roder, 2.7, y0+17, 2.1, Stone, Brick, true, 2)
	g.Use(Stone).Fill(roder.at(Block(V(-3.6, y0-5, 2.6), V(3.6, y0+4, 7.4))))
	g.Carve(roder.at(Block(V(-2.6, y0+0.3, 2.6), V(2.6, y0+5, 6.4))))
	g.Carve(roder.at(Arch{V(0, y0+0.3, 6.9), 1.3, 1.6, 1.6, AxisZ}))
	for _, x := range []float64{-3.6, 3.6} {
		r.roundTower(g, roder.f(x, 7.2), 1.7, y0+6, 4.2, Terracotta)
	}
	g.Paint(Sand, roder.at(Block(V(-2.6, y0-0.5, 2.6), V(2.6, y0+0.2, 6.4))))
	// The Siebersturm across the Plönlein's left street.
	sieber := rothenburgFrame{P(6.8, 11.6), P(13, 16).Angle()}
	r.squareTower(g, sieber, 2.3, y0+15, 2.1, Sand, Terracotta, true, 1)
	g.Paint(White, sieber.at(Block(V(-0.9, y0+11, 2.0), V(0.9, y0+12.8, 2.6))))
	g.Paint(Gold, sieber.at(Block(V(-0.3, y0+11.6, 2.0), V(0.3, y0+12.2, 2.6))))
	// Wall towers at the corners.
	towers := []struct {
		at    Vec2
		round bool
	}{
		{P(-34.6, 10.3), true}, {P(30.2, 16.8), false}, {P(38.9, 6), false},
		{P(19.4, -36.2), false}, {P(0, -39.4), false}, {P(-36.7, -25.4), true},
	}
	for i, t := range towers {
		roof := Terracotta
		if i%3 == 1 {
			roof = Brick
		}
		if t.round {
			r.roundTower(g, t.at, 2.4, y0+9, 4.6, roof)
		} else {
			f := r.onWall(t.at)
			f.c = t.at.Add(r.wallNormal(t.at).Scale(-0.4))
			r.squareTower(g, f, 2.0, y0+10, 1.7, Stone, roof, false, 4)
		}
	}
}

// f maps a local point to plan.
func (f rothenburgFrame) f(x, z float64) Vec2 { return f.plan(x, z) }

// bridge is the double bridge: two tiers of arches carrying the lane over the river.
func (r Rothenburg) bridge(g *Grid) {
	const xb, deck, hw, ramp = -15.0, 4.8, 1.7, 4.5
	zr := r.riverZ(xb)
	zN := 20.0
	for r.height(xb, zN) > deck {
		zN += 0.25
	}
	zS := zr + rothenburgRiver + 1.2
	floor := r.height(xb, zS+ramp)
	g.Use(Stone).Fill(Block(V(xb-hw, 0.8, zN-1.5), V(xb+hw, deck, zS)))
	const period = 3.4
	n := int((zS - zN) / period)
	z0 := zN + (zS-zN-float64(n)*period)/2
	for i := 0; i <= n; i++ {
		z := z0 + float64(i)*period
		g.Fill(Block(V(xb-hw-0.4, -0.5, z-0.65), V(xb+hw+0.4, 2.7, z+0.65)))
	}
	for i := 0; i < n; i++ {
		zc := z0 + (float64(i)+0.5)*period
		g.Carve(Arch{V(xb, 0.8, zc), 1.05, 0.4, 2*hw + 2, AxisX}, Arch{V(xb, 2.9, zc), 0.95, 0.2, 2*hw + 2, AxisX})
	}
	// Ramp down to the far bank.
	g.Fill(Func{Box{V(xb-hw, floor-2, zS), V(xb+hw, deck, zS+ramp)}, func(p Vec3) bool {
		y := deck - (p.Z-zS)*(deck-floor)/ramp
		return p.Y <= y && p.Y >= y-1.6
	}})
	g.Paint(Sand, Func{Box{V(xb-hw+0.6, 0, zN-1.5), V(xb+hw-0.6, deck, zS+ramp)}, func(p Vec3) bool {
		y := deck
		if p.Z > zS {
			y = deck - (p.Z-zS)*(deck-floor)/ramp
		}
		return p.Y > y-0.45
	}})
	g.Fill(Func{Box{V(xb-hw, deck, zN-1.5), V(xb+hw, deck+0.8, zS+ramp)}, func(p Vec3) bool {
		if math.Abs(p.X-xb) < hw-0.55 {
			return false
		}
		y := deck
		if p.Z > zS {
			y = deck - (p.Z-zS)*(deck-floor)/ramp
		}
		return p.Y <= y+0.8
	}})
}

// trees fills the gardens behind the houses and dots the slopes and far meadow.
func (r Rothenburg) trees(g *Grid, town *rothenburgTown) {
	clear := func(p Vec2) bool {
		for _, s := range town.streets {
			if s.has(p, 1.2) {
				return false
			}
		}
		for _, d := range []Vec2{P(0, 0), P(1.6, 0), P(-1.6, 0), P(0, 1.6), P(0, -1.6)} {
			if !town.free(p.Add(d)) {
				return false
			}
		}
		return true
	}
	for x := -40.0; x <= 40; x += 3.2 {
		for z := -42.0; z <= 46; z += 3.2 {
			i, j := int(x*10), int(z*10)
			p := P(x+1.4*(Hash01(i, j, 31)-0.5), z+1.4*(Hash01(i, j, 32)-0.5))
			h := 3.6 + 1.6*Hash01(i, j, 33)
			u, _ := r.wallInset(p)
			odds := Hash01(i, j, 34)
			switch {
			case u > 4.4:
				if odds < 0.55 && clear(p) {
					g.Build(kit.Tree{At: p, Base: rothenburgP, Height: h})
				}
			case u < -2.6 && u > -8 && r.inSite(p.X, p.Y):
				if p.Y > 6 && (p.X < -23 || p.X > -6) {
					continue // Vineyards.
				}
				if p.X > -18 && p.X < -11 && p.Y > 19 {
					continue // Lane and bridge.
				}
				if odds < 0.45 {
					g.Build(kit.Tree{At: p, Base: r.height(p.X, p.Y) - 0.3, Height: h})
				}
			case u <= -8 && p.Y > r.riverZ(p.X)+rothenburgRiver+1.5 && r.inSite(p.X, p.Y) && math.Abs(p.X+15) > 3:
				if odds < 0.3 {
					g.Build(kit.Tree{At: p, Base: r.height(p.X, p.Y) - 0.3, Height: h})
				}
			}
		}
	}
}
