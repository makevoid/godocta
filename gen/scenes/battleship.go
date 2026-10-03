package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// Battleship is a WWII fast battleship in the style of the Iowa class, at sea: a long,
// narrow hollow hull with a raked clipper bow, forward sheer and flare, a rounded
// cruiser stern, red antifouling, a charcoal boot-topping and a planked teak deck;
// three triple 16-inch turrets (two forward, superfiring, one aft); ten twin 5-inch
// mounts; a stepped tower bridge with conning tower, gun directors, a tripod foremast
// with a bedspring radar; two raked funnels; a mainmast with the ensign; quad 40 mm
// tubs; boats; anchors; the hull number; catapults with a floatplane and a crane on
// the fantail. y = 0 is the flat bottom; the waterline is about 3.4 ship units (4.4
// world units after battleshipScale) above it. Bow points +X.
type Battleship struct{}

func init() { Register(Battleship{}) }

func (Battleship) Name() string  { return "battleship" }
func (Battleship) Title() string { return "Battleship" }

const (
	battleshipHalfBeam = 10.0
	battleshipDeck     = 9.8 // Main deck height amidships.
	battleshipBoot     = 3.1 // Top of the red antifouling.
)

// deckY is the main deck height at x: a strong sheer forward, slightly lower aft.
func (Battleship) deckY(x float64) float64 {
	switch {
	case x > 8:
		t := (x - 8) / 72
		return battleshipDeck + 4.3*t*t
	case x < -20:
		t := (-20 - x) / 60
		return battleshipDeck - 0.5*t*t
	}
	return battleshipDeck
}

// stemX is the clipper stem's X at height y: cut away below, raking forward above.
func (Battleship) stemX(y float64) float64 {
	t := math.Min(math.Max(y/15, 0), 1)
	return 70 + 10*math.Pow(t, 1.3)
}

// sternX is the cruiser stern's X at height y, with the run cut up below the waterline.
func (Battleship) sternX(y float64) float64 {
	if y < 3.5 {
		return -80 + (3.5-y)*1.6
	}
	return -80
}

// halfWidth is the outer half-beam at (x, y), negative outside the hull's length.
func (s Battleship) halfWidth(x, y float64) float64 {
	xs, xb := s.sternX(y), s.stemX(y)
	if x < xs || x > xb {
		return -1
	}
	w := battleshipHalfBeam
	switch {
	case x > 12: // Long fine entry; fuller (flared) sections higher up.
		t := (x - 12) / (xb - 12)
		k := 1.35 + 1.2*math.Min(math.Max(y/14, 0), 1)
		w *= 1 - math.Pow(t, k)
	case x < -38: // Narrowing run into a rounded stern.
		t := (-38 - x) / (-38 - xs)
		w *= (1 - 0.3*t*t) * math.Pow(math.Max(1-math.Pow(t, 6), 0), 1.0/6)
	}
	if y < 3 { // Rounded bilge over a flat bottom.
		t := (3 - y) / 3
		w *= 1 - 0.4*t*t
	}
	return w
}

// innerWidth is the hollow interior's half-width: a skin about 1.2 thick.
func (s Battleship) innerWidth(x, y float64) float64 {
	const skin = 1.2
	a, b := s.halfWidth(x-skin, y), s.halfWidth(x+skin, y)
	if a < 0 || b < 0 {
		return -1
	}
	return math.Min(a, b) - skin
}

// place turns a solid authored facing +X around the origin to the angle, then moves it
// to (x, z).
func (Battleship) place(sol Solid, x, z, angle float64) Solid {
	return Translate(RotateY(sol, P(0, 0), angle), V(x, 0, z))
}

// battleshipScale enlarges the whole authored ship uniformly: recipe coordinates
// below are in "ship units" and every solid is scaled into world units.
const battleshipScale = 1.3

// battleshipScaled scales a solid about the origin by K.
type battleshipScaled struct {
	S Solid
	K float64
}

func (b battleshipScaled) Contains(p Vec3) bool { return b.S.Contains(p.Scale(1 / b.K)) }
func (b battleshipScaled) Bounds() Box {
	bb := b.S.Bounds()
	return Box{bb.Min.Scale(b.K), bb.Max.Scale(b.K)}
}

// battleshipGrid forwards fills to a grid, scaling every solid by k.
type battleshipGrid struct {
	g *Grid
	k float64
}

func (b battleshipGrid) scale(solids []Solid) []Solid {
	out := make([]Solid, len(solids))
	for i, s := range solids {
		out[i] = battleshipScaled{s, b.k}
	}
	return out
}

func (b battleshipGrid) Use(c Color) battleshipGrid { b.g.Use(c); return b }
func (b battleshipGrid) With(c Color, f func())     { b.g.With(c, f) }
func (b battleshipGrid) Fill(s ...Solid) battleshipGrid {
	b.g.Fill(b.scale(s)...)
	return b
}
func (b battleshipGrid) FillVolume(s ...Solid) battleshipGrid {
	b.g.FillVolume(b.scale(s)...)
	return b
}
func (b battleshipGrid) Paint(c Color, s ...Solid) battleshipGrid {
	b.g.Paint(c, b.scale(s)...)
	return b
}
func (b battleshipGrid) Carve(s ...Solid) battleshipGrid {
	b.g.Carve(b.scale(s)...)
	return b
}

// box is a block from x0 to x1 and y0 to y1, centered across the ship.
func (Battleship) box(x0, x1, half, y0, y1 float64) Cuboid {
	return Block(V(x0, y0, -half), V(x1, y1, half))
}

func (s Battleship) Build(grid *Grid) Order {
	g := battleshipGrid{grid, battleshipScale}
	s.buildHull(g)
	s.buildTurrets(g)
	s.buildSuperstructure(g)
	s.buildSecondaries(g)
	s.buildFantail(g)
	s.buildBowGear(g)
	return Along(P(-1, 0))
}

// buildHull lays the shell, deck, livery, portholes and hull number.
func (s Battleship) buildHull(g battleshipGrid) {
	hullBox := Box{V(-81, 0, -11), V(81, 16, 11)}
	g.Use(Slate).FillVolume(Func{hullBox, func(p Vec3) bool {
		return p.Y <= s.deckY(p.X) && math.Abs(p.Z) <= s.halfWidth(p.X, p.Y)
	}})
	g.Carve(Func{hullBox, func(p Vec3) bool {
		return p.Y >= 1.2 && p.Y <= s.deckY(p.X)-0.9 && math.Abs(p.Z) <= s.innerWidth(p.X, p.Y)
	}})
	// Livery: red antifouling, charcoal boot-topping at the waterline.
	g.Paint(Red, Block(V(-81, 0, -11), V(81, battleshipBoot, 11)))
	g.Paint(Charcoal, Block(V(-81, battleshipBoot+0.1, -11), V(81, battleshipBoot+1.0, 11)))
	// Measure 22: navy-blue hull below the deck's lowest point, haze gray above.
	g.Paint(DarkStone, Block(V(-81, battleshipDeck-0.8, -11), V(81, 16, 11)))
	// Teak deck with planking seams, leaving a gray margin along the deck edge.
	g.Paint(Sand, Func{hullBox, func(p Vec3) bool {
		return p.Y > s.deckY(p.X)-0.8 && math.Abs(p.Z) < s.halfWidth(p.X, p.Y)-0.6
	}})
	g.Paint(Wood, Func{hullBox, func(p Vec3) bool {
		return p.Y > s.deckY(p.X)-0.8 && math.Abs(p.Z) < s.halfWidth(p.X, p.Y)-0.6 &&
			math.Mod(p.Z+100.4, 2.4) < 0.5
	}})
	// Portholes fore and aft of the armored belt: recesses with dark backs.
	for _, run := range [][2]float64{{-76, -42}, {16, 52}} {
		for x := run[0]; x <= run[1]; x += 3 {
			y := s.deckY(x) - 2.4
			hole := Box{V(x-0.4, y-0.4, -11), V(x+0.4, y+0.4, 11)}
			g.Carve(Func{hole, func(p Vec3) bool { return math.Abs(p.Z) > s.halfWidth(p.X, p.Y)-0.6 }})
			g.Paint(Charcoal, Func{hole, func(p Vec3) bool { return math.Abs(p.Z) > s.halfWidth(p.X, p.Y)-1.5 }})
		}
	}
	// Hull number on both bows, reading left to right from either side.
	glyphs := map[byte][5]string{
		'6': {"###", "#..", "###", "#.#", "###"},
		'1': {".#.", "##.", ".#.", ".#.", "###"},
	}
	const text, px, x0, top = "61", 0.9, 55.0, battleshipDeck - 1.0
	width := float64(len(text))*4*px - px
	lit := func(along, y float64) bool {
		row := int(math.Floor((top - y) / px))
		col := int(math.Floor(along / px))
		if row < 0 || row > 4 || col < 0 {
			return false
		}
		ch, c := col/4, col%4
		if ch >= len(text) || c == 3 {
			return false
		}
		return glyphs[text[ch]][row][c] == '#'
	}
	g.Paint(White, Func{Box{V(x0, top-5*px, -11), V(x0+width, top, 11)}, func(p Vec3) bool {
		if math.Abs(p.Z) < s.halfWidth(p.X, p.Y)-0.9 {
			return false
		}
		if p.Z > 0 {
			return lit(p.X-x0, p.Y)
		}
		return lit(x0+width-p.X, p.Y)
	}})
}

// mainTurret is a triple 16-inch turret authored facing +X at the origin on base.
func (s Battleship) mainTurret(base float64) (body, barrels, ears Solid) {
	const h = 3.4
	plan := []Vec2{{-5.6, -3.4}, {-4.6, -4.3}, {1.2, -4.3}, {4.6, -3.0}, {4.6, 3.0}, {1.2, 4.3}, {-4.6, 4.3}, {-5.6, 3.4}}
	body = Func{Box{V(-5.6, base, -4.3), V(4.6, base+h, 4.3)}, func(p Vec3) bool {
		return InsidePolygon(plan, p.Plan()) && p.Y <= base+h-math.Max(p.X-2.4, 0)*0.75
	}}
	var tubes UnionSolid
	for _, z := range []float64{-2.1, 0, 2.1} {
		a := V(3.6, base+1.5, z)
		dir := V(1, 0.07, 0)
		tubes = append(tubes,
			Capsule{a, a.Add(dir.Scale(4)), 0.75},
			Capsule{a, a.Add(dir.Scale(12.5)), 0.5},
			Capsule{a.Add(dir.Scale(11.9)), a.Add(dir.Scale(12.5)), 0.62},
		)
	}
	ears = Union(Block(V(-3.8, base+1.9, -5.4), V(-2.2, base+2.9, 5.4)))
	return body, tubes, ears
}

// buildTurrets places the barbettes and the three main turrets.
func (s Battleship) buildTurrets(g battleshipGrid) {
	type mount struct{ x, lift, angle float64 }
	for _, m := range []mount{{40, 0.6, 0}, {25, 4.6, 0}, {-48, 0.6, math.Pi}} {
		deck := s.deckY(m.x)
		base := deck + m.lift
		g.Use(DarkStone).Fill(Cylinder{Center: P(m.x, 0), Radius: 4, Y0: deck - 1, Y1: base})
		body, barrels, ears := s.mainTurret(base)
		g.Use(DarkStone).Fill(s.place(body, m.x, 0, m.angle), s.place(ears, m.x, 0, m.angle))
		g.Use(Stone).Fill(s.place(barrels, m.x, 0, m.angle))
		// Roof plating and the gun-port shields.
		g.Paint(Stone, s.place(Block(V(-5.6, base+3.0, -4.3), V(2.6, base+3.4, 4.3)), m.x, 0, m.angle))
		g.Paint(Charcoal, s.place(Block(V(3.4, base+1.0, -3.0), V(4.6, base+2.0, 3.0)), m.x, 0, m.angle))
		if m.lift > 1 { // Quad 40 mm on the superfiring turret's roof.
			s.aaTub(g, m.x-3.6, 0, base+3.4, m.angle)
		} else if m.angle != 0 {
			s.aaTub(g, m.x+3.6, 0, base+3.4, m.angle)
		}
	}
}

// aaTub is a quad 40 mm mount in a round splinter shield, guns pointing at angle.
func (s Battleship) aaTub(g battleshipGrid, x, z, y, angle float64) {
	g.Use(DarkStone).Fill(Cylinder{Center: P(x, z), Radius: 1.8, Inner: 1.15, Y0: y, Y1: y + 1.1})
	g.Use(Stone).Fill(Cylinder{Center: P(x, z), Radius: 0.8, Y0: y, Y1: y + 1.0})
	var guns UnionSolid
	for _, side := range []float64{-0.5, 0.5} {
		guns = append(guns, Capsule{V(0.3, y+1.0, side), V(2.6, y+2.2, side), 0.4})
	}
	g.Use(Charcoal).Fill(s.place(guns, x, z, angle))
}

// fiveInch is a twin 5-inch/38 gun house at (x, z) on base, guns pointing at angle.
func (s Battleship) fiveInch(g battleshipGrid, x, z, base, angle float64) {
	house := Func{Box{V(-2.0, base, -1.5), V(1.8, base+2.1, 1.5)}, func(p Vec3) bool {
		if p.X < -0.6 && P(p.X+0.6, p.Z).Length() > 1.5 { // Rounded back.
			return false
		}
		return p.Y <= base+2.1-math.Max(p.X-0.6, 0)*0.7
	}}
	g.Use(DarkStone).Fill(Cylinder{Center: P(x, z), Radius: 1.3, Y0: base - 0.5, Y1: base + 0.3})
	g.Use(Stone).Fill(s.place(house, x, z, angle))
	var guns UnionSolid
	for _, side := range []float64{-0.65, 0.65} {
		guns = append(guns, Capsule{V(1.2, base+1.0, side), V(5.0, base+1.25, side), 0.42})
	}
	g.Use(DarkStone).Fill(s.place(guns, x, z, angle))
}

// director is a Mk 37 fire-control director with its radar dish, facing angle.
func (s Battleship) director(g battleshipGrid, x, z, base, angle float64) {
	body := Func{Box{V(-1.6, base, -1.4), V(1.8, base+1.8, 1.4)}, func(p Vec3) bool {
		return p.X < 0.4 || P(p.X-0.4, p.Z).Length() <= 1.4
	}}
	g.Use(Stone).Fill(s.place(body, x, z, angle))
	g.Use(Charcoal).Fill(s.place(Block(V(-0.2, base+1.8, -1.5), V(0.6, base+3.3, 1.5)), x, z, angle))
	g.Paint(Glass, s.place(Block(V(1.0, base+0.8, -1.4), V(1.8, base+1.3, 1.4)), x, z, angle))
}

// rangeDirector is a Mk 38 main-battery director with a long rangefinder.
func (s Battleship) rangeDirector(g battleshipGrid, x, base float64) {
	g.Use(Stone).Fill(s.box(x-1.9, x+1.9, 1.8, base, base+2.0))
	g.Fill(Capsule{V(x+0.2, base+1.3, -4.8), V(x+0.2, base+1.3, 4.8), 0.55})
	g.Fill(Block(V(x-0.6, base+0.7, -5.4), V(x+1.0, base+2.0, -4.4)), Block(V(x-0.6, base+0.7, 4.4), V(x+1.0, base+2.0, 5.4)))
	g.Use(Charcoal).Fill(s.box(x-0.4, x+0.4, 1.6, base+2.0, base+3.4))
	g.Paint(Glass, s.box(x+1.9, x+1.9, 1.6, base+0.7, base+1.2))
}

// funnel is a raked elliptical stack with a dark cap and an open top.
func (s Battleship) funnel(g battleshipGrid, x, a, b, y0, y1 float64) {
	rake := P(-0.14, 0)
	shape := func(da, db, lo, hi float64) Solid {
		return Sheared{Func{Box{V(x-a, lo, -b), V(x+a, hi+1.5, b)}, func(p Vec3) bool {
			u, v := (p.X-x)/(a-da), p.Z/(b-db)
			// Cap raked to match: higher at the front.
			return u*u+v*v <= 1 && p.Y <= hi+0.25*(p.X-x)
		}}, rake, y0}
	}
	g.Use(DarkStone).Fill(shape(0, 0, y0, y1))
	g.Paint(Charcoal, Sheared{Block(V(x-a-1, y1-2, -b-1), V(x+a+1, y1+2, b+1)), rake, y0})
	g.Carve(Sheared{Func{Box{V(x-a, y1-3, -b), V(x+a, y1+2, b)}, func(p Vec3) bool {
		u, v := (p.X-x)/(a-0.9), p.Z/(b-0.9)
		return u*u+v*v <= 1
	}}, rake, y0})
	// Steam pipes up the leading edge.
	g.Use(Stone).Fill(Sheared{Cylinder{Center: P(x+a-0.2, 0), Radius: 0.5, Y0: y0, Y1: y1 + 0.8}, rake, y0})
}

// buildSuperstructure adds the deckhouses, tower bridge, directors, funnels and masts.
func (s Battleship) buildSuperstructure(g battleshipGrid) {
	const dk = battleshipDeck
	l1, l2, l3 := dk+2.6, dk+5.0, dk+7.4
	// 01 level: broad deckhouse with a chamfered front.
	g.Use(DarkStone).Fill(Prism{Polygon: []Vec2{{-37, -7.4}, {17, -7.4}, {20.5, -4.6}, {20.5, 4.6}, {17, 7.4}, {-37, 7.4}}, Y0: dk - 0.5, Y1: l1})
	g.Carve(s.box(-36, 16.5, 6.5, dk+0.4, l1-0.8))
	// 02 level.
	g.Fill(Prism{Polygon: []Vec2{{-31, -4.4}, {13, -4.4}, {16, -2.6}, {16, 2.6}, {13, 4.4}, {-31, 4.4}}, Y0: l1, Y1: l2})
	g.Carve(s.box(-30, 12.5, 3.5, l1, l2-0.8))
	// Port rows on the deckhouse sides.
	for _, lv := range []struct{ x0, x1, half, y float64 }{{-35, 15, 7.4, dk + 1.1}, {-29, 11, 4.4, l1 + 1.0}} {
		for x := lv.x0; x <= lv.x1; x += 2.4 {
			for _, side := range []float64{-1, 1} {
				win := Block(V(x-0.4, lv.y, side*(lv.half-0.5)), V(x+0.4, lv.y+0.6, side*(lv.half+0.5)))
				g.Carve(win)
				g.Paint(Charcoal, Block(V(x-0.4, lv.y, side*(lv.half-1.4)), V(x+0.4, lv.y+0.6, side*(lv.half-0.5))))
			}
		}
	}

	// Tower bridge: armored conning tower, stepped levels and the bridge with wings.
	g.Use(DarkStone).Fill(EllipticCylinder{P(15.8, 0), 2.6, 3.2, l1, l3 + 3.0})
	g.Paint(Charcoal, EllipticCylinder{P(15.8, 0), 2.6, 3.2, l3 + 1.8, l3 + 2.4})
	g.Use(Stone).Fill(Prism{Polygon: []Vec2{{-1, -5.6}, {4, -5.6}, {4, -4}, {11.5, -4}, {14, -2.5}, {14, 2.5}, {11.5, 4}, {4, 4}, {4, 5.6}, {-1, 5.6}}, Y0: l2, Y1: l3})
	g.Carve(s.box(0, 13, 3.1, l2, l3-0.8))
	l4, l5, l6 := l3+2.6, l3+5.2, l3+7.8
	g.Fill(Prism{Polygon: []Vec2{{4.5, -4.4}, {11.5, -4.4}, {13.6, -2.4}, {13.6, 2.4}, {11.5, 4.4}, {4.5, 4.4}}, Y0: l3, Y1: l4})
	g.Fill(s.box(9.5, 12, 6.6, l4-0.6, l4)) // Bridge wings.
	g.Paint(Glass, Prism{Polygon: []Vec2{{4.5, -4.4}, {11.5, -4.4}, {13.6, -2.4}, {13.6, 2.4}, {11.5, 4.4}, {4.5, 4.4}}, Y0: l3 + 1.0, Y1: l3 + 1.6})
	g.Fill(Prism{Polygon: []Vec2{{5.5, -3.2}, {10.5, -3.2}, {12, -1.8}, {12, 1.8}, {10.5, 3.2}, {5.5, 3.2}}, Y0: l4, Y1: l5})
	g.Paint(Glass, Prism{Polygon: []Vec2{{5.5, -3.2}, {10.5, -3.2}, {12, -1.8}, {12, 1.8}, {10.5, 3.2}, {5.5, 3.2}}, Y0: l4 + 1.0, Y1: l4 + 1.6})
	// Open navigating bridge, fire-control tower and the forward Mk 38 director.
	g.Fill(Prism{Polygon: []Vec2{{5, -2.6}, {9.4, -2.6}, {10.6, -1.4}, {10.6, 1.4}, {9.4, 2.6}, {5, 2.6}}, Y0: l5, Y1: l6})
	g.Paint(Glass, Prism{Polygon: []Vec2{{5, -2.6}, {9.4, -2.6}, {10.6, -1.4}, {10.6, 1.4}, {9.4, 2.6}, {5, 2.6}}, Y0: l5 + 1.0, Y1: l5 + 1.6})
	g.Fill(Cylinder{Center: P(7.2, 0), Radius: 1.9, Y0: l6, Y1: l6 + 5})
	g.Use(DarkStone).Fill(Cylinder{Center: P(7.6, 0), Radius: 3.2, Y0: l6 + 2.2, Y1: l6 + 2.8})
	g.Use(Stone).Fill(Cylinder{Center: P(7.6, 0), Radius: 3.2, Inner: 2.6, Y0: l6 + 2.8, Y1: l6 + 3.6})
	s.rangeDirector(g, 7.2, l6+5)
	// Mk 37 directors: forward atop the bridge, abreast the tower on the wings.
	s.director(g, 11.4, 0, l5, 0)
	for _, side := range []float64{-1, 1} {
		s.director(g, 1.6, side*4.4, l3, 0)
	}
	// Tripod foremast with yard, bedspring air-search radar and topmast.
	g.Use(DarkStone).Fill(
		Capsule{V(6.4, l6+8.2, 0), V(5.8, l6+17, 0), 0.55},
		Capsule{V(4.0, l4, -2.6), V(5.9, l6+11, 0), 0.45},
		Capsule{V(4.0, l4, 2.6), V(5.9, l6+11, 0), 0.45},
		Capsule{V(5.9, l6+11.6, -4.5), V(5.9, l6+11.6, 4.5), 0.45},
		Capsule{V(5.8, l6+17, 0), V(5.6, l6+21.5, 0), 0.42},
	)
	radar := Block(V(6.5, l6+12.6, -3.0), V(7.3, l6+17.4, 3.0))
	g.Use(Charcoal).Fill(radar)
	g.Carve(Func{radar.Bounds(), func(p Vec3) bool {
		return math.Mod(p.Z+30.6, 1.6) < 0.7 && math.Mod(p.Y-l6+0.3, 1.6) < 0.7 &&
			math.Abs(p.Z) < 2.4 && p.Y > l6+13.2 && p.Y < l6+16.8
	}})

	// Funnels, each ringed by a searchlight platform.
	s.funnel(g, -3.2, 3.9, 2.8, l2, dk+17.4)
	s.funnel(g, -19.5, 3.7, 2.7, l2, dk+16.4)
	for _, f := range []struct{ x, y float64 }{{-3.2, dk + 10.4}, {-19.5, dk + 9.6}} {
		g.Use(DarkStone).Fill(EllipticCylinder{P(f.x+0.2, 0), 5.6, 4.4, f.y, f.y + 0.6})
		for _, side := range []float64{-1, 1} {
			g.Use(Stone).Fill(Cylinder{Center: P(f.x+0.6, side*3.6), Radius: 0.7, Y0: f.y + 0.6, Y1: f.y + 1.8})
			g.Paint(Glass, Block(V(f.x+0.9, f.y+0.9, side*3.6-0.4), V(f.x+1.3, f.y+1.5, side*3.6+0.4)))
		}
	}
	// Mainmast aft of the second funnel: pole, yard, surface radar, gaff and ensign.
	mx := -24.6
	g.Use(DarkStone).Fill(
		Capsule{V(mx, l2, 0), V(mx+0.6, dk+27, 0), 0.55},
		Capsule{V(mx-3.4, l2, 0), V(mx+0.3, dk+19, 0), 0.42},
		Capsule{V(mx+0.4, dk+22, -4), V(mx+0.4, dk+22, 4), 0.45},
		Capsule{V(mx+0.5, dk+20, 0), V(mx-4, dk+23, 0), 0.42},
	)
	g.Use(Charcoal).Fill(s.box(mx-0.3, mx+1.5, 1.4, dk+27, dk+28.2))
	flag := Block(V(mx-6.6, dk+21.6, -0.4), V(mx-3.6, dk+23.6, 0.4))
	g.Use(White).Fill(flag)
	g.Paint(Red, Func{flag.Bounds(), func(p Vec3) bool { return math.Mod(p.Y-dk+20, 0.8) < 0.4 }})
	g.Paint(Navy, Block(V(mx-4.9, dk+22.6, -0.4), V(mx-3.6, dk+23.6, 0.4)))

	// After superstructure: tower carrying the aft Mk 38 director.
	g.Use(Stone).Fill(s.box(-32, -26.5, 2.8, l1, l3))
	g.Carve(s.box(-31.2, -27.3, 2.0, l2, l3-0.8))
	g.Fill(Cylinder{Center: P(-29.2, 0), Radius: 1.7, Y0: l3, Y1: l3 + 2})
	s.rangeDirector(g, -29.2, l3+2)
	g.Use(Stone).Fill(Cylinder{Center: P(-11.6, 0), Radius: 1.4, Y0: l2, Y1: l2 + 1.6})
	s.director(g, -11.6, 0, l2+1.6, math.Pi)
	// Quad 40 mm tubs between the funnels and at the after deckhouse.
	for _, side := range []float64{-1, 1} {
		s.aaTub(g, -8.4, side*3.4, l2, side*1.2)
		s.aaTub(g, -15, side*3.4, l2, side*1.9)
		s.aaTub(g, -34.6, side*4.2, l1, side*2.2)
		s.aaTub(g, 19.2, side*3.1, l1, side*0.6)
	}
	// Boats in chocks with boat cranes.
	for _, bx := range []float64{-22.5, -28.5} {
		for _, side := range []float64{-1, 1} {
			z := side * 6.0
			g.Use(DarkStone).Fill(Block(V(bx-1.6, l1, z-0.7), V(bx+1.6, l1+0.5, z+0.7)))
			g.Use(White).Fill(Intersect(Ellipsoid{V(bx, l1+1.5, z), V(2.7, 1.1, 1.1)}, Block(V(bx-3, l1, z-2), V(bx+3, l1+1.6, z+2))))
			g.Paint(Wood, Block(V(bx-2.0, l1+1.2, z-0.6), V(bx+2.0, l1+1.6, z+0.6)))
		}
	}
	for _, side := range []float64{-1, 1} {
		g.Use(DarkStone).Fill(
			Capsule{V(-25.5, l2, side*3.4), V(-25.5, l2+4, side*3.4), 0.5},
			Capsule{V(-25.5, l2+4, side*3.4), V(-24, l2+3, side*6.2), 0.42},
		)
	}
}

// buildSecondaries places the ten twin 5-inch mounts along the 01 level and the
// shielded 20 mm guns along the deck edges.
func (s Battleship) buildSecondaries(g battleshipGrid) {
	l1 := battleshipDeck + 2.6
	for _, run := range [][2]float64{{-44, -38}, {-34, 18}, {28, 34}} {
		for x := run[0]; x <= run[1]; x += 4 {
			d := s.deckY(x)
			for _, side := range []float64{-1, 1} {
				z := side * (s.halfWidth(x, d) - 1.3)
				g.Use(Stone).Fill(Block(V(x-0.5, d, z-0.5), V(x+0.5, d+1.0, z+0.5)))
				g.Fill(Block(V(x+0.1, d+0.6, z-0.9), V(x+0.6, d+1.8, z+0.9)))
				g.Use(Charcoal).Fill(Capsule{V(x, d+1.3, z), V(x+2.0, d+2.0, z), 0.4})
			}
		}
	}
	for i, x := range []float64{16, 8.5, 1, -6.5, -14} {
		for _, side := range []float64{-1, 1} {
			angle := side * 0.35
			if i >= 3 {
				angle = side * (math.Pi - 0.35)
			}
			s.fiveInch(g, x, side*5.9, l1, angle)
		}
	}
}

// buildFantail adds the catapults, floatplane, aircraft crane, 40 mm tubs and ensign staff.
func (s Battleship) buildFantail(g battleshipGrid) {
	for _, side := range []float64{-1, 1} {
		x, z := -70.0, side*5.0
		deck := s.deckY(x)
		angle := side * (math.Pi - 0.6) // Launch end outboard and aft.
		g.Use(DarkStone).Fill(Cylinder{Center: P(x, z), Radius: 1.3, Y0: deck - 0.4, Y1: deck + 0.6})
		beam := Block(V(-4.2, deck+0.6, -0.6), V(7.6, deck+1.4, 0.6))
		g.Fill(s.place(beam, x, z, angle))
		g.Carve(s.place(Func{beam.Bounds(), func(p Vec3) bool {
			return math.Mod(p.X+40.4, 1.6) < 0.7 && p.Y < deck+1.2 && math.Abs(p.Z) > 0.2
		}}, x, z, angle))
		if side > 0 {
			s.floatplane(g, x, z, deck+1.4, angle)
		}
		s.aaTub(g, -57, side*6.4, s.deckY(-57), side*0.9)
	}
	// Aircraft crane at the stern.
	cx, deck := -77.0, s.deckY(-77)
	g.Use(DarkStone).Fill(Cylinder{Center: P(cx, 0), Radius: 0.9, Y0: deck - 0.4, Y1: deck + 5})
	g.Fill(Capsule{V(cx, deck+4.6, 0), V(cx+7, deck+9.5, 0), 0.5})
	g.Use(Charcoal).Fill(Capsule{V(cx+7, deck+9.5, 0), V(cx+7, deck+6.5, 0), 0.4})
	// Ensign staff on the stern.
	g.Use(Wood).Fill(Capsule{V(-79.2, deck, 0), V(-79.2, deck+4.5, 0), 0.42})
}

// floatplane is an OS2U Kingfisher on its catapult, nose along angle.
func (s Battleship) floatplane(g battleshipGrid, x, z, y, angle float64) {
	at := func(sol Solid) Solid { return s.place(Translate(sol, V(1.6, 0, 0)), x, z, angle) }
	g.Use(Navy).Fill(
		at(Capsule{V(-2.3, y+0.5, 0), V(2.4, y+0.5, 0), 0.55}),   // Main float.
		at(Block(V(-0.2, y+0.5, -0.3), V(0.6, y+1.6, 0.3))),      // Strut.
		at(Capsule{V(-3.2, y+2.0, 0), V(2.6, y+2.0, 0), 0.68}),   // Fuselage.
		at(Block(V(-0.2, y+2.3, -4.4), V(1.2, y+2.9, 4.4))),      // Wing.
		at(Block(V(-3.6, y+2.0, -1.6), V(-2.7, y+2.5, 1.6))),     // Tailplane.
		at(Block(V(-3.6, y+2.0, -0.3), V(-2.6, y+3.6, 0.3))),     // Fin.
		at(Capsule{V(0, y+1.6, -3.8), V(1.4, y+1.6, -3.8), 0.4}), // Wingtip floats.
		at(Capsule{V(0, y+1.6, 3.8), V(1.4, y+1.6, 3.8), 0.4}),
	)
	g.Use(Glass).Fill(at(Block(V(-1.4, y+2.4, -0.4), V(0.2, y+3.0, 0.4))))
	g.Use(Charcoal).Fill(at(Block(V(2.9, y+1.0, -1.3), V(3.4, y+3.0, 1.3))))
	g.Paint(White, at(Block(V(-0.2, y+2.6, 2.4), V(1.2, y+3.0, 3.6))), at(Block(V(-0.2, y+2.6, -3.6), V(1.2, y+3.0, -2.4))))
}

// buildBowGear adds the breakwater, anchors with chains, windlasses and the jackstaff.
func (s Battleship) buildBowGear(g battleshipGrid) {
	// V-shaped breakwater forward of turret 1.
	bw := 48.8
	g.Use(DarkStone).Fill(Func{Box{V(bw-3, battleshipDeck-1, -7.5), V(bw+1, battleshipDeck+5, 7.5)}, func(p Vec3) bool {
		wall := bw - 0.3*math.Abs(p.Z)
		d := s.deckY(p.X)
		return math.Abs(p.X-wall) <= 0.5 && p.Y >= d-0.4 && p.Y <= d+1.3 &&
			math.Abs(p.Z) <= s.halfWidth(p.X, p.Y)-0.4
	}})
	for _, side := range []float64{-1, 1} {
		ax := 64.0
		ay := s.deckY(ax) - 2.4
		w := s.halfWidth(ax, ay)
		// Hawse pipe recess and the stowed anchor.
		g.Carve(Func{Box{V(ax-0.8, ay, -11), V(ax+0.8, ay+1.6, 11)}, func(p Vec3) bool {
			return p.Z*side > s.halfWidth(p.X, p.Y)-0.7
		}})
		zc := side * (w + 0.2)
		g.Use(Charcoal).Fill(
			Block(V(ax-0.45, ay-2.8, zc-0.45), V(ax+0.45, ay+0.8, zc+0.45)),
			Block(V(ax-1.4, ay-3.4, zc-0.5), V(ax+1.4, ay-2.6, zc+0.5)),
			Block(V(ax-1.4, ay-3.4, zc-0.5), V(ax-0.8, ay-2.0, zc+0.5)),
			Block(V(ax+0.8, ay-3.4, zc-0.5), V(ax+1.4, ay-2.0, zc+0.5)),
		)
		// Chains along the forecastle to the windlasses.
		g.Fill(Capsule{V(ax, s.deckY(ax)+0.3, side*(w-1.4)), V(56, s.deckY(56)+0.3, side*1.4), 0.42})
		g.Use(DarkStone).Fill(Cylinder{Center: P(55, side*1.4), Radius: 0.9, Y0: s.deckY(55) - 0.4, Y1: s.deckY(55) + 1.2})
	}
	// Bow 40 mm tubs and the jackstaff.
	for _, side := range []float64{-1, 1} {
		s.aaTub(g, 58.5, side*2.6, s.deckY(58.5), side*0.4)
	}
	g.Use(Wood).Fill(Capsule{V(78.6, s.deckY(78.6)-0.2, 0), V(78.9, s.deckY(78.6)+4.2, 0), 0.42})
}
