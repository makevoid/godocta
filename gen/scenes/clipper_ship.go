package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// ClipperShip is a 19th-century tea clipper under full sail on the open sea: a long,
// sharp black hull with a gold sheer line, gilded figurehead and trailboards, a copper
// bottom showing at the boot top, a teak deck with white deckhouses, boats, poop and
// forecastle; a long bowsprit and jib boom; three raked masts with stacked, billowing
// square sails, jibs, staysails and a spanker; standing rigging and a red ensign.
// The ship heels to port on the starboard tack. The bow points to +X; y = 0 is the
// waterline of a thin sea plate with wave crests, a bow wave and a wake.
type ClipperShip struct{}

func init() { Register(ClipperShip{}) }

func (ClipperShip) Name() string  { return "clipper_ship" }
func (ClipperShip) Title() string { return "Clipper Ship" }

const (
	clipperShipHeel  = -0.085 // Heel to starboard (+Z, leeward), about 5°.
	clipperShipBrace = 0.36   // Yards braced with the starboard (lee) yardarms aft.
	clipperShipBeam  = 17.0
	clipperShipScale = 1.22                    // World units per ship-frame unit.
	clipperShipSkin  = 0.62 / clipperShipScale // Half-thickness of sails in the ship frame.
)

// toShip maps a world point into the upright, unscaled ship frame.
func (ClipperShip) toShip(p Vec3) Vec3 {
	s, c := math.Sincos(clipperShipHeel)
	return V(p.X, p.Y*c-p.Z*s, p.Z*c+p.Y*s).Scale(1 / clipperShipScale)
}

// toWorld maps a ship-frame point into the heeled world.
func (ClipperShip) toWorld(q Vec3) Vec3 {
	s, c := math.Sincos(clipperShipHeel)
	q = q.Scale(clipperShipScale)
	return V(q.X, q.Y*c+q.Z*s, q.Z*c-q.Y*s)
}

// heel turns a ship-frame solid into the heeled world.
func (s ClipperShip) heel(solid Solid) Solid {
	b := solid.Bounds()
	box := EmptyBox()
	for _, x := range []float64{b.Min.X, b.Max.X} {
		for _, y := range []float64{b.Min.Y, b.Max.Y} {
			for _, z := range []float64{b.Min.Z, b.Max.Z} {
				w := s.toWorld(V(x, y, z))
				box = box.Union(Box{w, w})
			}
		}
	}
	return Func{box.Grow(0.3), func(p Vec3) bool { return solid.Contains(s.toShip(p)) }}
}

// deckY is the main deck surface at x, with sheer rising toward both ends.
func (ClipperShip) deckY(x float64) float64 {
	u := x / 50
	if x > 0 {
		return 6.4 + 2.4*u*u
	}
	return 6.4 + 1.4*u*u
}

// railY is the top of the bulwark.
func (s ClipperShip) railY(x float64) float64 { return s.deckY(x) + 2.0 }

// bowX is the stem's X at height y: a raked, gently curved clipper stem.
func (ClipperShip) bowX(y float64) float64 {
	if y < 0 {
		return 47 + 0.6*y
	}
	return 47 + 4.5*math.Pow(y/10.8, 1.2)
}

// sternX is the counter stern's X at height y: the run ends at the waterline and the
// rounded counter overhangs above it.
func (ClipperShip) sternX(y float64) float64 {
	switch {
	case y < 0:
		return -44 - y
	case y < 3:
		return -44 - 6.5*math.Sqrt(y/3)
	}
	return -50.5 - 0.12*(y-3)
}

// halfWidth is the hull's outer half-beam at (x, y), or −1 outside its length.
func (s ClipperShip) halfWidth(x, y float64) float64 {
	xs, xb := s.sternX(y), s.bowX(y)
	if x < xs || x > xb {
		return -1
	}
	plan := 1.0
	if x > 4 { // Long, fine entry, flaring toward the rail.
		t := (x - 4) / (xb - 4)
		plan = 1 - math.Pow(t, 1.45+0.09*clamp01(y/10)*10)
	}
	if x < -10 { // Clean run into the rounded counter.
		t := (-10 - x) / (-10 - xs)
		k := clamp01(y / 3.5)
		plan = (1-k)*(1-math.Pow(t, 1.7)) + k*math.Sqrt(math.Max(0, 1-math.Pow(t, 2.4)))
	}
	sec := 1 - 0.004*y*y
	if y < 0 {
		sec = 1 - 0.05*y*y
	}
	return clipperShipBeam / 2 * plan * sec
}

func clamp01(v float64) float64 { return math.Max(0, math.Min(1, v)) }

// innerWidth is the half-width of the hollow interior, a skin about 1.1 thick.
func (s ClipperShip) innerWidth(x, y float64) float64 {
	const skin = 1.1
	a, b := s.halfWidth(x-skin, y), s.halfWidth(x+skin, y)
	if a < 0 || b < 0 {
		return -1
	}
	return math.Min(a, b) - skin
}

// clipperShipMast is one mast: its foot, rake, mast heads and yards.
type clipperShipMast struct {
	x, rake, deck     float64
	top, cross, truck float64      // Lower masthead (top), topmast head (crosstrees), truck.
	yards             [][2]float64 // Height and half-length, from the course up.
	courseFoot        float64
}

// at is the mast's X at height y.
func (m clipperShipMast) at(y float64) float64 { return m.x - m.rake*(y-m.deck) }

// point is a point on the mast axis offset forward by dx and across by z.
func (m clipperShipMast) point(y, dx, z float64) Vec3 { return V(m.at(y)+dx, y, z) }

// clipperShipBuilder applies ship-frame solids to the grid through the heel.
type clipperShipBuilder struct {
	s ClipperShip
	g *Grid
}

func (b clipperShipBuilder) fill(c Color, solids ...Solid) {
	b.g.Use(c)
	for _, x := range solids {
		b.g.Fill(b.s.heel(x))
	}
}

func (b clipperShipBuilder) vol(c Color, solids ...Solid) {
	b.g.Use(c)
	for _, x := range solids {
		b.g.FillVolume(b.s.heel(x))
	}
}

func (b clipperShipBuilder) paint(c Color, solids ...Solid) {
	for _, x := range solids {
		b.g.Paint(c, b.s.heel(x))
	}
}

func (b clipperShipBuilder) carve(solids ...Solid) {
	for _, x := range solids {
		b.g.Carve(b.s.heel(x))
	}
}

func (b clipperShipBuilder) line(c Color, r float64, pts ...Vec3) {
	for i := 1; i < len(pts); i++ {
		b.fill(c, Capsule{pts[i-1], pts[i], r})
	}
}

func (s ClipperShip) Build(g *Grid) Order {
	b := clipperShipBuilder{s, g}
	s.buildHull(b)
	s.buildDeck(b)
	s.buildBow(b)
	s.buildStern(b)
	s.buildRig(b)
	s.buildSea(g)
	// The sea plate is the lowest layer: nothing hangs below it.
	g.Carve(Block(V(-200, -10, -200), V(200, -0.45, 200)))
	return Outward(P(0, 0))
}

// buildHull adds the black shell with copper bottom, gold sheer line and teak rail.
func (s ClipperShip) buildHull(b clipperShipBuilder) {
	B := clipperShipBeam
	box := Box{V(-52, -3, -B/2-1), V(53, 12, B/2+1)}
	b.vol(Charcoal, Func{box, func(p Vec3) bool {
		return p.Y <= s.railY(p.X) && math.Abs(p.Z) <= s.halfWidth(p.X, p.Y)
	}})
	b.carve(Func{box.Grow(1), func(p Vec3) bool {
		return p.Y > -2.5 && math.Abs(p.Z) <= s.innerWidth(p.X, p.Y)
	}})
	// Copper below the boot top with a gold edge.
	shell := func(p Vec3) bool {
		w := s.halfWidth(p.X, p.Y)
		return w >= 0 && math.Abs(p.Z) <= w+0.6
	}
	b.paint(Terracotta, Func{box, func(p Vec3) bool { return p.Y < 1.5 && shell(p) }})
	b.paint(Gold, Func{box, func(p Vec3) bool { return p.Y >= 1.3 && p.Y < 1.7 && shell(p) }})
	// Gold sheer line just below the rail.
	b.vol(Gold, Func{box, func(p Vec3) bool {
		h := s.railY(p.X) - p.Y
		w := s.halfWidth(p.X, p.Y)
		z := math.Abs(p.Z)
		return h >= 1.0 && h <= 1.55 && w > 0 && z <= w+0.35 && z >= w-0.6
	}})
	// White inner bulwarks and a teak rail cap.
	b.paint(White, Func{box, func(p Vec3) bool {
		return p.Y > s.deckY(p.X)+0.3 && math.Abs(p.Z) <= s.innerWidth(p.X, p.Y)+0.6
	}})
	b.paint(Wood, Func{box, func(p Vec3) bool { return p.Y > s.railY(p.X)-0.45 && shell(p) }})
}

// buildDeck adds the teak deck, poop and forecastle, deckhouses, boats, hatches,
// capstans, wheel and skylights.
func (s ClipperShip) buildDeck(b clipperShipBuilder) {
	B := clipperShipBeam
	box := Box{V(-52, 4, -B/2), V(53, 13, B/2)}
	inside := func(p Vec3, inset float64) bool { return math.Abs(p.Z) <= s.innerWidth(p.X, p.Y)+0.4-inset }
	b.fill(Wood, Func{box, func(p Vec3) bool {
		return math.Abs(p.Y-s.deckY(p.X)) <= 0.5 && inside(p, 0)
	}})
	// Planking seams.
	b.paint(Sand, Func{box, func(p Vec3) bool {
		return math.Abs(p.Y-s.deckY(p.X)) <= 0.5 && math.Mod(p.Z+100.4, 2.4) < 0.5
	}})

	// Raised poop aft and topgallant forecastle forward, white fronted with doors.
	raised := func(x0, x1, face float64) {
		top := func(x float64) float64 { return s.deckY(x) + 1.9 }
		b.fill(Wood, Func{Box{V(x0, 5, -B/2), V(x1, 12, B/2)}, func(p Vec3) bool {
			return p.Y > s.deckY(p.X) && p.Y <= top(p.X) && inside(p, 0)
		}})
		b.paint(Sand, Func{Box{V(x0, 5, -B/2), V(x1, 12, B/2)}, func(p Vec3) bool {
			return p.Y > top(p.X)-0.6 && math.Mod(p.Z+100.4, 2.4) < 0.5
		}})
		b.paint(White, Func{Box{V(face-0.7, 5, -B/2), V(face+0.7, 12, B/2)}, func(p Vec3) bool {
			return p.Y > s.deckY(p.X)+0.4 && inside(p, 0.3)
		}})
		for _, z := range []float64{-3.6, 3.6} {
			b.carve(BlockAt(P(face, z), 2.2, 1.6, s.deckY(face)+0.5, s.deckY(face)+1.6))
			b.paint(Charcoal, BlockAt(P(face, z), 4.2, 1.6, s.deckY(face)+0.3, s.deckY(face)+1.8))
		}
		// White railing on top: stanchions and a rail.
		rail := func(p Vec3, h0, h1 float64) bool {
			d := p.Y - top(p.X)
			return d > h0 && d <= h1 && math.Abs(p.Z) <= s.innerWidth(p.X, p.Y)+0.4 &&
				math.Abs(p.Z) >= s.innerWidth(p.X, p.Y)-0.6
		}
		b.fill(White, Func{Box{V(x0, 6, -B/2), V(x1, 14, B/2)}, func(p Vec3) bool {
			return rail(p, 0.4, 1.9) && (rail(p, 1.2, 1.9) || math.Mod(p.X+100, 2.0) < 0.7)
		}})
	}
	raised(-52, -35, -35)
	raised(39, 52, 39)

	// Deckhouses: white walls, dark windows, teak roofs with white boats on skids.
	house := func(x0, x1, half float64) {
		y0 := s.deckY((x0+x1)/2) + 0.4
		y1 := y0 + 3.0
		b.fill(White, Subtract(Block(V(x0, y0, -half), V(x1, y1, half)),
			Block(V(x0+1, y0, -half+1), V(x1-1, y1-1, half-1))))
		b.fill(Wood, Block(V(x0-0.5, y1-0.6, -half-0.5), V(x1+0.5, y1+0.2, half+0.5)))
		for x := x0 + 2; x <= x1-2; x += 2.6 {
			for _, z := range []float64{-half, half} {
				b.paint(Charcoal, BlockAt(P(x, z), 1.0, 1.6, y0+1.0, y0+2.0))
			}
		}
		for _, z := range []float64{-2, 2} {
			b.carve(BlockAt(P(x1, z), 1.6, 1.4, y0, y0+2.2))
		}
		// Boats on skids above the roof.
		for _, z := range []float64{-half + 2.4, half - 2.4} {
			c := V((x0+x1)/2, y1+1.6, z)
			l := (x1 - x0) / 2
			b.fill(Wood, BlockAt(P(c.X-l*0.5, z), 0.8, 3.6, y1, y1+0.9), BlockAt(P(c.X+l*0.5, z), 0.8, 3.6, y1, y1+0.9))
			b.vol(White, Subtract(Ellipsoid{c, V(l-0.4, 1.5, 1.6)}, Block(V(c.X-l, c.Y, z-2), V(c.X+l, c.Y+2, z+2))))
			b.fill(Wood, Ellipsoid{c.Add(V(0, -0.2, 0)), V(l-1.6, 0.5, 0.8)})
		}
	}
	house(13, 24, 4.4)
	house(-22, -8, 4.6)

	// Main hatch, fore hatch and capstans.
	for _, x := range []float64{-3.5, 31} {
		y := s.deckY(x) + 0.4
		b.fill(Wood, BlockAt(P(x, 0), 4.5, 5, y, y+1.0))
		b.fill(Charcoal, BlockAt(P(x, 0), 3.3, 3.8, y+0.6, y+1.2))
	}
	for _, x := range []float64{-28, 44} {
		y := s.deckY(x) + 0.4
		if x > 0 || x < -34 {
			y += 1.9
		}
		b.fill(White, Cylinder{Center: P(x, 0), Radius: 1.1, Y0: y, Y1: y + 1.8})
		b.fill(Gold, Cylinder{Center: P(x, 0), Radius: 1.4, Y0: y + 1.6, Y1: y + 2.2})
	}
	// Windlass and anchor bitts just aft of the forecastle.
	wy := s.deckY(37) + 0.4
	b.fill(Wood, Capsule{V(37, wy+1, -3.5), V(37, wy+1, 3.5), 0.8})
	b.fill(White, BlockAt(P(37, -4), 1.2, 1.2, wy, wy+2.0), BlockAt(P(37, 4), 1.2, 1.2, wy, wy+2.0))

	// Poop: skylight, binnacle, wheel box and the ship's wheel.
	py := s.deckY(-42) + 2.3
	b.fill(White, BlockAt(P(-40, 0), 4.4, 3.4, py, py+1.2))
	b.fill(Glass, Gable{Min: P(-42, -1.7), Max: P(-38, 1.7), Base: py + 1.2, Ridge: AxisX})
	b.fill(Gold, Cylinder{Center: P(-44.5, 0), Radius: 0.7, Y0: py, Y1: py + 1.8})
	b.fill(Wood, BlockAt(P(-47.2, 0), 1.6, 2.4, py, py+1.8))
	wc := V(-46.2, py+2.0, 0)
	b.fill(Wood, Func{Box{wc.Sub(V(0.6, 2.2, 2.2)), wc.Add(V(0.6, 2.2, 2.2))}, func(p Vec3) bool {
		d := math.Hypot(p.Y-wc.Y, p.Z-wc.Z)
		a := math.Atan2(p.Y-wc.Y, p.Z-wc.Z)
		spoke := math.Abs(math.Sin(4*a)) < 0.35
		return math.Abs(p.X-wc.X) <= 0.5 && (d >= 1.3 && d <= 2.0 || spoke && d <= 2.3)
	}})
	b.fill(Gold, Sphere(wc, 0.6))
}

// buildBow adds the gilded figurehead, trailboards, bowsprit, jib boom and martingale.
func (s ClipperShip) buildBow(b clipperShipBuilder) {
	sy := s.deckY(47)
	xb := s.bowX(sy)
	// Knee of the head below the bowsprit, carrying the figurehead.
	b.fill(Charcoal, Func{Box{V(xb-3, 3, -0.8), V(xb+3, sy+2, 0.8)}, func(p Vec3) bool {
		return p.X <= s.bowX(p.Y)+1.6*clamp01((p.Y-3)/(sy-3)) && p.X >= s.bowX(p.Y)-2
	}})
	fh := V(s.bowX(sy-2)+1.6, sy-1.6, 0)
	b.vol(Gold, Ellipsoid{fh, V(1.6, 2.6, 1.1)})
	b.vol(Gold, Sphere(fh.Add(V(0.6, 2.6, 0)), 0.9))
	b.fill(Gold, Capsule{fh.Add(V(0.4, 1.4, 0.5)), fh.Add(V(3.2, 2.6, 0.4)), 0.5})
	// Gilded trailboards: scrolls on both bows below the sheer.
	for _, side := range []float64{-1, 1} {
		z := func(x, y float64) float64 { return side * (s.halfWidth(x, y) + 0.2) }
		var pts []Vec3
		for t := 0.0; t <= 1.0001; t += 0.1 {
			x := xb - 1.5 - 9*t
			y := sy - 1.5 - 1.6*math.Sin(math.Pi*t)
			pts = append(pts, V(x, y, z(x, y)))
		}
		b.line(Gold, 0.5, pts...)
		for _, c := range []Vec2{{xb - 4, sy - 3.5}, {xb - 8.5, sy - 3.0}} {
			for a := 0.0; a < 2*math.Pi; a += 0.5 {
				r := 0.6 + 0.12*a
				x, y := c.X+r*math.Cos(a), c.Y+r*math.Sin(a)
				b.fill(Gold, Sphere(V(x, y, z(x, y)), 0.5))
			}
		}
	}
	// Bowsprit and jib boom, steeved up about 12°.
	start, dir := s.spritStart(), s.spritDir()
	at := func(d float64) Vec3 { return start.Add(dir.Scale(d)) }
	b.fill(White, Capsule{at(-4), at(14), 1.0})
	b.fill(Wood, Capsule{at(9), at(31), 0.6})
	b.fill(Gold, Sphere(at(31.4), 0.7))
	// Martingale (dolphin striker) and its stays; bobstay to the stem.
	ms := at(13.5)
	low := ms.Add(V(-0.6, -5.5, 0))
	b.fill(Wood, Capsule{ms, low, 0.5})
	b.line(Charcoal, 0.45, at(31), low, V(s.bowX(2)+0.5, 2.2, 0))
	b.line(Charcoal, 0.45, at(24), low)
	// Footropes / guys spreading to the catheads.
	for _, side := range []float64{-1, 1} {
		b.line(Charcoal, 0.45, at(30), V(xb-5, s.railY(xb-5), side*(s.halfWidth(xb-5, s.railY(xb-5))+0.3)))
	}
}

func (s ClipperShip) spritStart() Vec3 {
	return V(s.bowX(s.deckY(48))-1.2, s.deckY(48)+1.6, 0)
}

func (ClipperShip) spritDir() Vec3 {
	a := 12 * math.Pi / 180
	return V(math.Cos(a), math.Sin(a), 0)
}

// buildStern adds gilded scrollwork on the counter, stern windows and the red ensign.
func (s ClipperShip) buildStern(b clipperShipBuilder) {
	B := clipperShipBeam
	ry := s.railY(-50)
	// Gold name board and scroll band around the counter.
	b.vol(Gold, Func{Box{V(-53, ry-3, -B/2), V(-42, ry, B/2)}, func(p Vec3) bool {
		w := s.halfWidth(p.X, p.Y)
		if w < 0 || p.X > -44 {
			return false
		}
		z := math.Abs(p.Z)
		d := ry - p.Y
		in := z >= w-0.8 && z <= w+0.3
		return in && (d >= 1.0 && d <= 1.55 || d > 1.55 && d < 2.6 && math.Abs(p.Z) < 3.5)
	}})
	// Stern windows below the scroll band.
	for _, z := range []float64{-4, -1.4, 1.4, 4} {
		y := ry - 4.2
		x := s.sternX(y) + 0.1
		b.paint(Glass, BlockAt(P(x, z), 1.6, 1.4, y-0.8, y+0.6))
	}
	// Ensign staff at the taffrail and the red ensign streaming aft and to leeward.
	fx := -48.6
	top := s.deckY(fx) + 2.0 + 9
	b.fill(Wood, Cylinder{Center: P(fx, 0), Radius: 0.5, Y0: s.deckY(fx) + 1.5, Y1: top + 0.6})
	b.fill(Gold, Sphere(V(fx, top+0.9, 0), 0.6))
	dir := P(-math.Cos(0.5), math.Sin(0.5))
	fw, fh := 9.0, 5.4
	flag := func(f func(u, v float64) bool) Solid {
		return Func{Box{V(fx-fw-1, top-fh-1, -1), V(fx+1, top+1, fw+1)}, func(p Vec3) bool {
			d := p.Plan().Sub(P(fx, 0))
			u := d.Dot(dir)
			if u < 0 || u > fw {
				return false
			}
			off := d.X*(-dir.Y) + d.Y*dir.X // Perpendicular offset across the flag.
			wave := 0.7 * math.Sin(u*0.8) * u / fw
			v := top - p.Y
			if v < 0 || v > fh {
				return false
			}
			return math.Abs(off-wave) <= 0.5 && f(u, v)
		}}
	}
	b.fill(Red, flag(func(u, v float64) bool { return true }))
	b.paint(Navy, flag(func(u, v float64) bool { return u < fw*0.45 && v < fh*0.5 }))
	b.paint(White, flag(func(u, v float64) bool {
		if u >= fw*0.45 || v >= fh*0.5 {
			return false
		}
		return math.Abs(u-fw*0.225) < 0.55 || math.Abs(v-fh*0.25) < 0.45
	}))
}

// buildRig adds masts, tops, yards, square sails, headsails, staysails, the spanker and
// standing rigging.
func (s ClipperShip) buildRig(b clipperShipBuilder) {
	mast := func(x, rake, scale float64, skysail bool, courseFoot float64) clipperShipMast {
		d := s.deckY(x)
		m := clipperShipMast{x: x, rake: rake, deck: d, courseFoot: courseFoot}
		h := func(v float64) float64 { return d + v*scale }
		m.top, m.cross, m.truck = h(21), h(40), h(70)
		m.yards = [][2]float64{
			{h(19.5), 18.5 * scale}, {h(28.5), 16.6 * scale}, {h(37), 14.6 * scale},
			{h(45.5), 11.8 * scale}, {h(53.5), 9.2 * scale},
		}
		if skysail {
			m.yards = append(m.yards, [2]float64{h(60.5), 6.6 * scale})
		}
		return m
	}
	fore := mast(28, 0.04, 0.92, false, s.railY(28)+3.2)
	main := mast(3, 0.05, 1.0, true, s.railY(3)+3.4)
	mizzen := mast(-22, 0.06, 0.8, false, s.railY(-22)+5.5)
	for _, m := range []clipperShipMast{fore, main, mizzen} {
		s.buildMast(b, m)
	}
	s.buildHeadsails(b, fore)
	s.buildStaysails(b, fore, main, mizzen)
	s.buildSpanker(b, mizzen)
	for _, m := range []clipperShipMast{fore, main, mizzen} {
		s.buildShrouds(b, m)
	}
	// Fore-and-aft stays between the mastheads.
	start, dir := s.spritStart(), s.spritDir()
	at := func(d float64) Vec3 { return start.Add(dir.Scale(d)) }
	b.line(Charcoal, 0.45, fore.point(fore.top, 0, 0), at(3))
	b.line(Charcoal, 0.45, main.point(main.top, 0, 0), fore.point(fore.deck+2, 0, 0))
	b.line(Charcoal, 0.45, main.point(main.cross, 0, 0), fore.point(fore.top+1, 0, 0))
	b.line(Charcoal, 0.45, main.point(main.truck-6, 0, 0), fore.point(fore.cross+1, 0, 0))
	b.line(Charcoal, 0.45, mizzen.point(mizzen.top, 0, 0), main.point(main.deck+5, 0, 0))
	b.line(Charcoal, 0.45, mizzen.point(mizzen.cross, 0, 0), main.point(main.top+1, 0, 0))
	b.line(Charcoal, 0.45, mizzen.point(mizzen.truck-5, 0, 0), main.point(main.cross+1, 0, 0))
	// House flag streaming from the main truck.
	tx, ty := main.at(main.truck), main.truck+1.2
	b.fill(Red, Func{Box{V(tx-7, ty-2, -1), V(tx+1, ty+2, 7)}, func(p Vec3) bool {
		d := p.Plan().Sub(P(tx, 0))
		dir := P(-math.Cos(0.6), math.Sin(0.6))
		u := d.Dot(dir)
		off := d.X*(-dir.Y) + d.Y*dir.X
		return u >= 0 && u <= 6 && math.Abs(off) <= 0.5 && math.Abs(p.Y-ty) <= 1.3*(1-u/7)+0.4
	}})
}

// buildMast adds a mast in three tapering pieces, its top, crosstrees, yards and sails.
func (s ClipperShip) buildMast(b clipperShipBuilder, m clipperShipMast) {
	shear := func(solid Solid) Solid { return Sheared{solid, P(-m.rake, 0), m.deck} }
	c := P(m.x, 0)
	b.fill(Sand, shear(Frustum{c, 1.35, 1.1, m.deck - 1, m.top + 2.5}))
	b.fill(Wood, shear(Frustum{c, 0.9, 0.7, m.top - 1, m.cross + 2}), shear(Frustum{c, 0.6, 0.5, m.cross - 1, m.truck}))
	b.fill(Gold, Sphere(m.point(m.truck+0.3, 0, 0), 0.7))
	// Top: a flat platform; crosstrees: a spreader.
	b.fill(Charcoal, shear(BlockAt(c.Add(P(-0.6, 0)), 4.2, 7, m.top-0.4, m.top+0.4)))
	b.fill(Charcoal, shear(BlockAt(c, 1.6, 6, m.cross-0.4, m.cross+0.3)))
	pivot := func(y float64) Vec2 { return P(m.at(y), 0) }
	for i, yd := range m.yards {
		y, half := yd[0], yd[1]
		b.fill(Charcoal, RotateY(Capsule{m.point(y, 1.1, -half), m.point(y, 1.1, half), 0.55}, pivot(y), clipperShipBrace))
		top := y - 0.7
		var bottom, wBottom float64
		if i == 0 {
			bottom, wBottom = m.courseFoot, half+0.5
		} else {
			bottom, wBottom = m.yards[i-1][0]+0.8, m.yards[i-1][1]-0.6
		}
		bulge := 0.16*(top-bottom) + 0.6
		roach := 0.5
		if i == 0 {
			roach = 1.8
		}
		sail := s.squareSail(m, top, bottom, half-0.6, wBottom, bulge, roach)
		color := White
		if i == 0 {
			color = Sand
		}
		b.vol(color, RotateY(sail, pivot((top+bottom)/2), clipperShipBrace))
	}
	if m.x < 0 {
		return
	}
	// Studding sails set outboard of the fore and main yards on booms run out from the
	// yardarms: topmast and topgallant stunsails on both sides.
	y := m.yards
	type stun struct {
		top, bottom, in0, in1, out0, out1 float64
	}
	stuns := []stun{
		{y[2][0] - 0.7, y[0][0] + 0.8, y[2][1] - 0.3, y[0][1] - 0.3, y[2][1] + 6, y[0][1] + 6.5},
		{y[3][0] - 0.7, y[2][0] + 0.8, y[3][1] - 0.3, y[2][1] - 0.3, y[3][1] + 4.5, y[2][1] + 5.5},
	}
	for _, side := range []float64{-1, 1} {
		for _, yd := range [][2]float64{{y[0][0], y[0][1] + 7}, {y[2][0], y[2][1] + 6.4}, {y[3][0], y[3][1] + 4.9}} {
			b.fill(Wood, RotateY(Capsule{m.point(yd[0], 1.1, side*(yd[1]-8)), m.point(yd[0], 1.1, side*yd[1]), 0.42}, pivot(yd[0]), clipperShipBrace))
		}
		for _, st := range stuns {
			sail := s.panel(m, st.top, st.bottom, side*st.in0, side*st.in1, side*st.out0, side*st.out1, 0.12*(st.top-st.bottom)+0.4, 0.3)
			b.vol(White, RotateY(sail, pivot((st.top+st.bottom)/2), clipperShipBrace))
		}
	}
}

// squareSail hangs from a yard at height top down to bottom, tapering between half-widths
// wTop and wBottom, billowing forward (+X) by up to bulge, with an arched foot.
func (s ClipperShip) squareSail(m clipperShipMast, top, bottom, wTop, wBottom, bulge, roach float64) Solid {
	return s.panel(m, top, bottom, -wTop, -wBottom, wTop, wBottom, bulge, roach)
}

// panel is a four-sided square-rigged sail between heights top and bottom whose edges
// run from l0/r0 at the top to l1/r1 at the foot (Z), set 1.1 forward of the mast.
func (ClipperShip) panel(m clipperShipMast, top, bottom, l0, l1, r0, r1, bulge, roach float64) Solid {
	zmin := math.Min(math.Min(l0, l1), math.Min(r0, r1))
	zmax := math.Max(math.Max(l0, l1), math.Max(r0, r1))
	box := Box{V(m.at(top)-1, bottom, zmin-1), V(m.at(bottom)+bulge+3, top+0.5, zmax+1)}
	return Func{box, func(p Vec3) bool {
		v := (top - p.Y) / (top - bottom)
		if v < 0 || v > 1 {
			return false
		}
		lo, hi := l0+(l1-l0)*v, r0+(r1-r0)*v
		if hi < lo {
			lo, hi = hi, lo
		}
		u := 2*(p.Z-lo)/(hi-lo) - 1
		if math.Abs(u) > 1 || p.Y < bottom+roach*(1-u*u) {
			return false
		}
		bl := bulge * (1 - 0.55*u*u) * math.Sin(math.Pi*0.5*math.Min(v*1.5, 1))
		return math.Abs(p.X-m.at(p.Y)-1.1-bl) <= clipperShipSkin
	}}
}

// triSail is a triangular sail tack-head-clew, bellying to leeward (+Z) by bulge.
func (ClipperShip) triSail(a, h, c Vec3, bulge float64) Solid {
	box := BoxOf(a, h).Union(BoxOf(c, c)).Grow(1)
	box.Max.Z += bulge
	area := (h.X-a.X)*(c.Y-a.Y) - (c.X-a.X)*(h.Y-a.Y)
	return Func{box, func(p Vec3) bool {
		l1 := ((h.X-p.X)*(c.Y-p.Y) - (c.X-p.X)*(h.Y-p.Y)) / area
		l2 := ((c.X-p.X)*(a.Y-p.Y) - (a.X-p.X)*(c.Y-p.Y)) / area
		l3 := 1 - l1 - l2
		if l1 < 0 || l2 < 0 || l3 < 0 {
			return false
		}
		z := l1*a.Z + l2*h.Z + l3*c.Z + bulge*27*l1*l2*l3
		return math.Abs(p.Z-z) <= clipperShipSkin
	}}
}

// buildHeadsails adds the fore topmast staysail and three jibs on their stays.
func (s ClipperShip) buildHeadsails(b clipperShipBuilder, fore clipperShipMast) {
	start, dir := s.spritStart(), s.spritDir()
	at := func(d float64) Vec3 { return start.Add(dir.Scale(d)) }
	type stay struct {
		mastY, boom float64
		clew        Vec3
	}
	stays := []stay{
		{fore.cross - 2, 12, V(36, s.railY(36)+2.5, 2.5)},
		{fore.cross + 6, 20, V(41, s.railY(41)+5.5, 3.5)},
		{fore.cross + 13, 26, V(46, s.railY(46)+9, 4.5)},
		{fore.truck - 4, 31, V(52, s.railY(50)+13, 5.5)},
	}
	for i, st := range stays {
		hi := fore.point(st.mastY, 0.6, 0)
		lo := at(st.boom)
		b.line(Charcoal, 0.45, hi, lo)
		tack := lo.Add(hi.Sub(lo).Scale(0.04))
		head := lo.Add(hi.Sub(lo).Scale(0.78 - 0.04*float64(i)))
		color := White
		if i%2 == 1 {
			color = Sand
		}
		b.vol(color, s.triSail(tack.Add(V(0, 0.6, 0)), head.Add(V(0, 0, -0.3)), st.clew, 1.6+0.3*float64(i)))
	}
}

// buildStaysails adds the triangular staysails set between the masts.
func (s ClipperShip) buildStaysails(b clipperShipBuilder, fore, main, mizzen clipperShipMast) {
	sails := []struct {
		hi, lo, clew Vec3
	}{
		{main.point(main.cross, 0.6, 0), fore.point(fore.top+1, 0.6, 0), main.point(main.yards[1][0]+1.5, 5, 3)},
		{main.point(main.truck-6, 0.6, 0), fore.point(fore.cross+1, 0.6, 0), main.point(main.yards[3][0]+1, 4, 3)},
		{mizzen.point(mizzen.cross, 0.6, 0), main.point(main.top+1, 0.6, 0), mizzen.point(mizzen.yards[1][0]+1.5, 4, 3)},
		{mizzen.point(mizzen.truck-5, 0.6, 0), main.point(main.cross+1, 0.6, 0), mizzen.point(mizzen.yards[3][0]+1, 3.5, 3)},
	}
	for _, sl := range sails {
		tack := sl.lo.Add(sl.hi.Sub(sl.lo).Scale(0.1))
		head := sl.lo.Add(sl.hi.Sub(sl.lo).Scale(0.85))
		b.vol(White, s.triSail(tack, head, sl.clew, 1.6))
	}
}

// buildSpanker adds the gaff sail aft of the mizzen, its gaff and boom swung to leeward.
func (s ClipperShip) buildSpanker(b clipperShipBuilder, m clipperShipMast) {
	tackY, throatY := s.deckY(m.x)+4.2, m.top-1.5
	xm := m.at(tackY) - 1
	clew := V(-53, tackY+0.8, 0)
	peak := V(-41, throatY+7, 0)
	pivot := P(m.at(tackY), 0)
	swing := func(solid Solid) Solid { return RotateY(solid, pivot, -0.22) }
	b.fill(Wood, swing(Capsule{V(xm, tackY-0.4, 0), clew.Add(V(-1, -0.4, 0)), 0.55}))
	b.fill(Wood, swing(Capsule{V(m.at(throatY)-1, throatY+0.5, 0), peak.Add(V(-0.5, 0.6, 0)), 0.5}))
	poly := []Vec2{P(xm, tackY), P(clew.X, clew.Y), P(peak.X, peak.Y), P(m.at(throatY)-1, throatY)}
	sail := Func{Box{V(clew.X-1, tackY-1, -1), V(xm+1, peak.Y+1, 4)}, func(p Vec3) bool {
		if !InsidePolygon(poly, P(p.X, p.Y)) {
			return false
		}
		u := (p.X - xm) / (clew.X - xm)
		v := (p.Y - tackY) / (peak.Y - tackY)
		return math.Abs(p.Z-2.2*math.Sin(math.Pi*clamp01(u))*math.Sin(math.Pi*clamp01(v))) <= clipperShipSkin
	}}
	b.vol(Sand, swing(sail))
}

// buildShrouds adds lower shrouds to channels on the hull sides, topmast shrouds,
// and long backstays from the upper mastheads.
func (s ClipperShip) buildShrouds(b clipperShipBuilder, m clipperShipMast) {
	for _, side := range []float64{-1, 1} {
		chainY := s.railY(m.x) - 1.0
		w := s.halfWidth(m.x, chainY)
		b.fill(Charcoal, Block(V(m.x-6, chainY-0.4, side*(w-0.4)), V(m.x+1.5, chainY+0.3, side*(w+1.2))))
		for i, dx := range []float64{-5, -3.2, -1.4, 0.4} {
			x := m.x + dx
			hw := s.halfWidth(x, chainY) + 0.8
			b.line(Charcoal, 0.45, m.point(m.top-0.5, -0.6+0.3*float64(i), side*1.2), V(x, chainY, side*hw))
		}
		for _, dx := range []float64{-1.4, 0.4} {
			b.line(Charcoal, 0.45, m.point(m.cross-0.5, dx*0.3, side*0.6), m.point(m.top+0.4, dx, side*3.2))
		}
		// Backstays from the topmast head and the royal masthead.
		for _, st := range [][2]float64{{m.cross + 1, -9}, {m.truck - 4, -13}} {
			x := m.x + st[1]
			y := s.railY(x) - 0.6
			b.line(Charcoal, 0.45, m.point(st[0], 0, side*0.5), V(x, y, side*(s.halfWidth(x, y)+0.6)))
		}
	}
}

// buildSea lays a thin sea plate around the ship with wind-wave crests, the bow wave
// and its diverging arms, a white wake and foam along the waterline.
func (s ClipperShip) buildSea(g *Grid) {
	span := g.Span()
	inSea := func(x, z float64) bool {
		u, v := (x/clipperShipScale-4)/76, z/clipperShipScale/27
		return u*u+v*v <= 1
	}
	inHull := func(p Vec3) bool {
		q := s.toShip(p)
		w := s.halfWidth(q.X, q.Y)
		return w >= 0 && math.Abs(q.Z) < w+0.1
	}
	hash := func(p Vec3, k int) float64 {
		c := Nearest(p.Scale(1 / span))
		return Hash01(int(c.X), int(c.Z), k)
	}
	water := func(p Vec3) bool {
		c, ok := g.ColorOf(Nearest(p.Scale(1 / span)))
		return ok && c == Water
	}
	// sea returns the surface level (a multiple of 0.4) and foam strength at p.
	sea := func(p Vec3) (float64, float64) {
		x, z := p.X/clipperShipScale, p.Z/clipperShipScale
		// Distance from the waterline of the (upright) hull, roughly.
		hw := s.halfWidth(x, 0.4)
		near := 99.0
		if hw >= 0 {
			near = math.Abs(z) - hw
		}
		h, foam := 0.0, 0.0
		// Wind waves running to leeward (−Z), crests broken along their length.
		a := 0.3
		zr := p.Z*math.Cos(a) + p.X*math.Sin(a)
		xr := p.X*math.Cos(a) - p.Z*math.Sin(a)
		crest := math.Sin(zr*2*math.Pi/9.5+1.3*math.Sin(xr*0.05)) *
			(0.5 + 0.5*math.Sin(xr*0.12+2*math.Sin(zr*0.08)))
		if near > 3 {
			if crest > 0.5 {
				h = 0.4
			}
			if crest > 0.8 {
				h, foam = 0.8, 0.85
			}
		}
		// Bow wave piled against the fore body, highest at the stem.
		if hw >= 0 && x > 26 && near < 2.4 {
			t := (x - 26) / 22
			h = math.Max(h, 0.4+0.4*math.Round(t))
			foam = math.Max(foam, 0.5+0.5*t)
		}
		// Foam along the waterline amidships and aft.
		if hw >= 0 && x <= 26 && near < 1.4 {
			foam = math.Max(foam, 0.45)
		}
		// Diverging arms of the bow wave and the quarter wave.
		for _, src := range [][3]float64{{48.5, 0.36, 1}, {-40, 0.36, 0.6}} {
			along := src[0] - x
			if along < 2 || along > 75 {
				continue
			}
			arm := 1 + along*src[1]
			if math.Abs(math.Abs(z)-arm) < 1.1+along*0.012 {
				fade := (1 - along/75) * src[2]
				h = math.Max(h, 0.4)
				if along < 20*src[2] {
					h = math.Max(h, 0.8)
				}
				foam = math.Max(foam, 0.25+0.7*fade)
			}
		}
		// Wake astern: a widening band of churned water drifting to windward.
		if along := -43 - x; along > 0 {
			zc := -0.05 * along
			if math.Abs(z-zc) < 2.6+0.14*along {
				foam = math.Max(foam, 0.85-0.012*along)
			}
		}
		return h, foam
	}
	box := Box{V(-92, -0.5, -34), V(100, 1.2, 34)}
	g.Use(Water).Fill(Func{box, func(p Vec3) bool {
		if !inSea(p.X, p.Z) || inHull(p) || g.HasAt(p) {
			return false
		}
		h, _ := sea(p)
		return p.Y >= -0.45 && p.Y <= h+0.15
	}})
	g.Paint(White, Func{box, func(p Vec3) bool {
		h, foam := sea(p)
		return p.Y > h-0.25 && hash(p, 11) < foam && water(p)
	}})
	g.Paint(Glass, Func{box, func(p Vec3) bool {
		h, foam := sea(p)
		return p.Y > h-0.25 && foam == 0 && hash(p, 12) < 0.05 && water(p)
	}})
}
