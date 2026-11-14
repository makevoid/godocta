package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// PaddleSteamer is a grand Mississippi sternwheel packet under way past a levee landing.
// A long, shallow white hull with a spoon bow and a red boot top carries three tiers of
// open galleries (main, boiler and texas decks) with posts, bracketed valances, railings
// and red trim in front of cabin walls with green doors and dark windows. A glazed pilot
// house with a red dome rides on the texas roof behind twin black smokestacks with gold
// bands, feathered gold crowns and a spreader bar, trailing smoke aft. The red sternwheel
// with its buckets, spoked flanges, cranks and white pitman arms hangs from the fantail,
// throwing spray and a foaming wake. On the foredeck are cotton bales, the grand
// staircase, a capstan, the jackstaff and twin landing stages slung from booms; lifeboats
// and a steam calliope sit on the texas roof and passengers line the rails. The bow points
// to +X. The far bank (−Z) is a cobbled levee with a wharf, bales and barrels, a brick
// cotton warehouse, willows and bald cypress; y = 0 is the river plate.
type PaddleSteamer struct{}

func init() { Register(PaddleSteamer{}) }

func (PaddleSteamer) Name() string  { return "paddle_steamer" }
func (PaddleSteamer) Title() string { return "Paddle Steamer" }

const (
	paddleSteamerHalf   = 11.0  // Hull half-beam.
	paddleSteamerStern  = -40.0 // Transom.
	paddleSteamerBow    = 48.0  // Stem head.
	paddleSteamerMain   = 3.2   // Deck surfaces: main deck,
	paddleSteamerBoiler = 9.5   // boiler deck,
	paddleSteamerHurr   = 15.5  // hurricane deck, which carries the texas,
	paddleSteamerTexas  = 21.0  // and the texas roof.
	paddleSteamerStackX = 21.0  // Smokestack axis.
	paddleSteamerStackZ = 5.0
	paddleSteamerStackY = 45.0 // Crown tips.
	paddleSteamerWheelX = -48.2
	paddleSteamerWheelY = 6.2
	paddleSteamerWheelR = 7.8
	paddleSteamerShore  = -25.5 // Mean waterline of the far bank.
)

// paddleSteamerDeck is the plan outline of one deck: square at the stern, rounded at the bow.
type paddleSteamerDeck struct {
	x0, x1 float64 // Stern edge and bow tip.
	half   float64 // Half-beam.
	r      float64 // Length of the rounded bow end.
	bays   int     // Post bays along one side, stern corner to bow tip.
}

var (
	// The boiler and hurricane decks share one outline; the texas roof is smaller.
	paddleSteamerGuard = paddleSteamerDeck{x0: -39, x1: 31, half: 10.5, r: 10.5, bays: 15}
	paddleSteamerRoof  = paddleSteamerDeck{x0: -30, x1: 13, half: 8, r: 6, bays: 10}
)

// paddleSteamerRail selects what a gallery carries along its deck edge.
type paddleSteamerRail struct {
	from, to float64 // Edge range with a railing; the stern edge is negative.
	gap      float64 // Clear opening at the bow tip.
	lamps    bool    // Gold lanterns on every other post.
}

// length is the edge length from the stern corner to the bow tip.
func (d paddleSteamerDeck) length() float64 {
	return d.x1 - d.r - d.x0 + math.Pi/2*(d.r+d.half)/2
}

// bay is the post spacing at edge position e.
func (d paddleSteamerDeck) bay(e float64) float64 {
	if e < 0 {
		return d.half / math.Round(d.half/4.5)
	}
	return d.length() / float64(d.bays)
}

// edge maps a plan point to its position along the deck edge (0 at the stern corner,
// negative across the stern, growing toward the bow tip) and its depth inside the edge.
func (d paddleSteamerDeck) edge(x, z float64) (e, depth float64) {
	az := math.Abs(z)
	xc := d.x1 - d.r
	if x > xc {
		u, v := (x-xc)/d.r, az/d.half
		rho := math.Hypot(u, v)
		if rho < 1e-6 {
			return d.length(), math.Min(d.r, d.half)
		}
		return xc - d.x0 + (math.Pi/2-math.Atan2(v, u))*(d.r+d.half)/2, math.Hypot(x-xc, az) * (1 - rho) / rho
	}
	side, end := d.half-az, x-d.x0
	if end < side {
		return -side, end
	}
	return x - d.x0, side
}

func paddleSteamerSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func paddleSteamerLerp(a, b, t float64) float64 { return a + (b-a)*t }

// paddleSteamerNoise is smooth value noise in [-1, 1].
func paddleSteamerNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return paddleSteamerLerp(paddleSteamerLerp(a, b, sx), paddleSteamerLerp(c, d, sx), sz)*2 - 1
}

// deckY is the main deck surface, with a little sheer toward the stem.
func (PaddleSteamer) deckY(x float64) float64 {
	if x <= 30 {
		return paddleSteamerMain
	}
	t := (x - 30) / (paddleSteamerBow - 30)
	return paddleSteamerMain + 1.2*t*t
}

// bowX is the stem's X at height y: a spoon bow sweeping forward to the deck.
func (PaddleSteamer) bowX(y float64) float64 {
	t := math.Max(0, math.Min(1, y/4))
	return paddleSteamerBow - 5*math.Pow(1-t, 1.6)
}

// halfWidth is the hull's half-beam at (x, y), or −1 beyond its ends.
func (s PaddleSteamer) halfWidth(x, y float64) float64 {
	xb := s.bowX(y)
	if x < paddleSteamerStern || x > xb {
		return -1
	}
	w := paddleSteamerHalf * (0.93 + 0.07*math.Max(0, math.Min(1, y/3)))
	if x > 14 {
		t := (x - 14) / (xb - 14)
		w *= math.Pow(1-t*t, 0.7)
	}
	if x < -35 {
		t := (-35 - x) / 5
		w *= 1 - 0.08*t*t
	}
	return w
}

// inPlate is the rounded river plate.
func (PaddleSteamer) inPlate(x, z float64) bool {
	u, v := math.Abs(x+2.5)/65.5, math.Abs(z+11)/36
	return u*u*u+v*v*v <= 1
}

// shore is the Z of the far bank's waterline.
func (PaddleSteamer) shore(x float64) float64 {
	return paddleSteamerShore + 2.2*paddleSteamerNoise(x*0.06, 3.3, 11)
}

// ground is the bank's height, or −1 over water and off the plate.
func (s PaddleSteamer) ground(x, z float64) float64 {
	if !s.inPlate(x, z) {
		return -1
	}
	d := s.shore(x) - z
	if d < 0 {
		return -1
	}
	return 4.6*paddleSteamerSmooth(d/7) + 0.5*paddleSteamerNoise(x*0.11, z*0.11, 12)*paddleSteamerSmooth(d/5)
}

// cells visits every lattice cell whose true center lies in the box.
func (PaddleSteamer) cells(g *Grid, b Box, f func(c Cell, p Vec3)) {
	h := g.Span() / 2
	lo := func(v float64) int32 { return int32(math.Ceil(v/h - 1e-9)) }
	hi := func(v float64) int32 { return int32(math.Floor(v/h + 1e-9)) }
	for y := lo(b.Min.Y); y <= hi(b.Max.Y); y++ {
		for z := lo(b.Min.Z); z <= hi(b.Max.Z); z++ {
			if (z^y)&1 != 0 {
				continue
			}
			for x := lo(b.Min.X); x <= hi(b.Max.X); x++ {
				if (x^y)&1 != 0 {
					continue
				}
				c := Cell{x, y, z}
				f(c, g.World(c))
			}
		}
	}
}

// put sets one cell to a color.
func (PaddleSteamer) put(g *Grid, c Cell, col Color) {
	g.Remove(c)
	g.With(col, func() { g.Add(c) })
}

// shell fills the cells of a predicate within t of its outside (sideways or above).
func (PaddleSteamer) shell(g *Grid, box Box, t float64, in func(Vec3) bool) {
	g.Fill(Func{box, func(p Vec3) bool {
		if !in(p) {
			return false
		}
		return !(in(p.Add(V(t, 0, 0))) && in(p.Add(V(-t, 0, 0))) && in(p.Add(V(0, 0, t))) &&
			in(p.Add(V(0, 0, -t))) && in(p.Add(V(0, t, 0))) && in(p.Add(V(0, -t, 0))))
	}})
}

// line draws a polyline of capsules.
func (PaddleSteamer) line(g *Grid, c Color, r float64, pts ...Vec3) {
	g.With(c, func() {
		for i := 1; i < len(pts); i++ {
			g.Fill(Capsule{pts[i-1], pts[i], r})
		}
	})
}

func (s PaddleSteamer) Build(g *Grid) Order {
	s.hull(g)
	s.houses(g)
	s.decks(g)
	s.pilotHouse(g)
	s.stacks(g)
	s.wheel(g)
	s.foredeck(g)
	s.fittings(g)
	s.crowd(g)
	s.bank(g)
	s.landing(g)
	s.river(g)
	s.smoke(g)
	// The river plate is the lowest layer: nothing hangs below it.
	g.Carve(Block(V(-200, -10, -200), V(200, -0.6, 200)))
	return Along(P(1, 0))
}

// hull is the shallow white shell with a red boot top and a planked main deck.
func (s PaddleSteamer) hull(g *Grid) {
	B := paddleSteamerHalf
	box := Box{V(paddleSteamerStern-1, -1.2, -B-1), V(paddleSteamerBow+1, 5.5, B+1)}
	g.Use(White).FillVolume(Func{box, func(p Vec3) bool {
		return p.Y <= s.deckY(p.X)-0.2 && math.Abs(p.Z) <= s.halfWidth(p.X, p.Y)
	}})
	g.Carve(Func{box, func(p Vec3) bool {
		if p.Y > s.deckY(p.X)-1.1 {
			return false
		}
		a, b := s.halfWidth(p.X-1.1, p.Y), s.halfWidth(p.X+1.1, p.Y)
		return math.Abs(p.Z) <= math.Min(a, b)-1.1
	}})
	g.Paint(Red, Block(V(box.Min.X, -1.2, -B-1), V(box.Max.X, 0.95, B+1)))
	g.Paint(Wood, Func{box, func(p Vec3) bool {
		return p.Y > s.deckY(p.X)-0.75 && p.X > paddleSteamerStern+0.9 &&
			math.Abs(p.Z) <= s.halfWidth(p.X+0.9, p.Y)-0.9
	}})
}

// house is one hollow white deckhouse.
func (PaddleSteamer) house(g *Grid, x0, x1, half, y0, y1 float64) {
	g.Use(White).Fill(Subtract(Block(V(x0, y0, -half), V(x1, y1, half)),
		Block(V(x0+1, y0-1, -half+1), V(x1-1, y1+1, half-1))))
}

// pane paints an opening on both side walls of a deckhouse of the given half-width.
func (PaddleSteamer) pane(g *Grid, c Color, x0, x1, y0, y1, half float64) {
	for _, side := range []float64{-1, 1} {
		g.Paint(c, Block(V(x0, y0, side*(half-1.3)), V(x1, y1, side*(half+0.3))))
	}
}

// houses are the engine house, the boiler-deck cabin and the texas, with their doors
// and windows set between the gallery posts.
func (s PaddleSteamer) houses(g *Grid) {
	main, boiler, hurr, texas := paddleSteamerMain, paddleSteamerBoiler, paddleSteamerHurr, paddleSteamerTexas
	s.house(g, -36, 17, 7, main-0.3, boiler-0.6)
	s.house(g, -34, 15, 7, boiler-0.3, hurr-0.6)
	s.house(g, -27, 9, 5, hurr-0.3, texas-0.6)

	d := paddleSteamerGuard
	bay := d.bay(1)
	for k := 1; k <= 10; k++ {
		xc := d.x0 + (float64(k)+0.5)*bay
		// Main deck: freight doors alternate with engine-room windows.
		if k%2 == 1 {
			s.pane(g, Charcoal, xc-1.5, xc+1.5, main, main+3.9, 7)
		} else {
			s.pane(g, Navy, xc-1.1, xc+1.1, main+2.1, main+3.7, 7)
		}
		// Boiler deck: a louvred stateroom door and a window in every bay.
		s.pane(g, Foliage, xc-1.8, xc-0.5, boiler, boiler+3.2, 7)
		if xc+1.8 < 14 {
			s.pane(g, Navy, xc+0.6, xc+1.8, boiler+1.5, boiler+3.1, 7)
		}
	}
	r := paddleSteamerRoof
	bay = r.bay(1)
	for k := 1; k <= 7; k++ {
		xc := r.x0 + (float64(k)+0.5)*bay
		if k == 2 || k == 5 {
			s.pane(g, Foliage, xc-0.7, xc+0.7, hurr, hurr+3.0, 5)
		} else {
			s.pane(g, Navy, xc-1.0, xc+1.0, hurr+1.5, hurr+3.0, 5)
		}
	}
	// Forward ends: the boiler fronts with glowing fire doors, the saloon entrance
	// and the texas front windows.
	g.Paint(Charcoal, Block(V(15.7, main, -5), V(17.3, main+4.2, 5)))
	for _, z := range []float64{-3.4, -1.1, 1.1, 3.4} {
		g.Paint(Terracotta, Block(V(15.7, main+0.6, z-0.6), V(17.3, main+1.8, z+0.6)))
		g.Paint(Gold, Block(V(15.7, main+0.9, z-0.3), V(17.3, main+1.4, z+0.3)))
	}
	g.Paint(Foliage, Block(V(13.7, boiler, -1.4), V(15.3, boiler+3.4, 1.4)))
	g.Paint(Glass, Block(V(13.7, boiler+1.4, -5.6), V(15.3, boiler+3.2, -2.6)),
		Block(V(13.7, boiler+1.4, 2.6), V(15.3, boiler+3.2, 5.6)))
	g.Paint(Glass, Block(V(7.7, hurr+1.4, -3.8), V(9.3, hurr+3.0, -0.8)),
		Block(V(7.7, hurr+1.4, 0.8), V(9.3, hurr+3.0, 3.8)))
	// After ends.
	g.Paint(Charcoal, Block(V(-36.3, main, -2.2), V(-34.7, main+3.8, 2.2)))
	g.Paint(Foliage, Block(V(-34.3, boiler, -1.2), V(-32.7, boiler+3.2, 1.2)))
	g.Paint(Navy, Block(V(-34.3, boiler+1.5, -5.4), V(-32.7, boiler+3.1, -3.2)),
		Block(V(-34.3, boiler+1.5, 3.2), V(-32.7, boiler+3.1, 5.4)))
	g.Paint(Foliage, Block(V(-27.3, hurr, -1), V(-25.7, hurr+3.0, 1)))
}

// slab is a two-layer deck with a white fascia over a red trim line; its upper face
// takes the floor color with lengthwise seams.
func (s PaddleSteamer) slab(g *Grid, d paddleSteamerDeck, top float64, floor, seam Color) {
	h := g.Span() / 2
	t0 := (math.Round(top/h) - 1) * h // Center of the top layer.
	s.cells(g, Box{V(d.x0, t0-h-0.2, -d.half), V(d.x1, t0+0.2, d.half)}, func(c Cell, p Vec3) {
		_, depth := d.edge(p.X, p.Z)
		if depth < 0 {
			return
		}
		upper := p.Y > t0-0.2
		col := White
		switch {
		case depth > 0.9 && upper && math.Mod(math.Abs(p.Z)+0.4, 2.4) < 0.8:
			col = seam
		case depth > 0.9 && upper:
			col = floor
		case depth <= 0.9 && !upper:
			col = Red
		}
		s.put(g, c, col)
	})
}

// gallery stands posts with bracketed valances along a deck edge, between a floor and
// the deck above, and hangs a railing between them.
func (s PaddleSteamer) gallery(g *Grid, d paddleSteamerDeck, floor, ceil float64, o paddleSteamerRail) {
	h := g.Span() / 2
	f0 := (math.Round(floor/h) - 1) * h // Center of the floor's top layer.
	c0 := (math.Round(ceil/h) - 2) * h  // Center of the upper slab's bottom layer.
	s.cells(g, Box{V(d.x0, f0+0.2, -d.half), V(d.x1, c0-0.2, d.half)}, func(c Cell, p Vec3) {
		e, depth := d.edge(p.X, p.Z)
		if depth < 0 || depth > 1 || o.gap > 0 && e > d.length()-o.gap {
			return
		}
		bay := d.bay(e)
		k := math.Round(e / bay)
		near := math.Abs(e - bay*k)
		up, down := p.Y-f0, c0-p.Y
		switch {
		case near <= 0.5 && o.lamps && int(k)%2 == 0 && down > 1.0 && down < 1.9:
			s.put(g, c, Gold)
		case near <= 0.5, down < 0.6 && near <= 1.75, down < 1.0 && near <= 1.1:
			s.put(g, c, White)
		case e < o.from || e > o.to:
		case up > 1.4 && up < 2.2:
			s.put(g, c, White)
		}
	})
}

// decks lays the boiler deck, hurricane deck and texas roof and their galleries.
func (s PaddleSteamer) decks(g *Grid) {
	d, r := paddleSteamerGuard, paddleSteamerRoof
	s.slab(g, d, paddleSteamerBoiler, Wood, Wood)
	s.slab(g, d, paddleSteamerHurr, DarkStone, Stone)
	s.slab(g, r, paddleSteamerTexas, Slate, DarkStone)
	all := 1000.0
	s.gallery(g, d, paddleSteamerMain, paddleSteamerBoiler, paddleSteamerRail{from: -0.6, to: 55.5, gap: 2.8})
	s.gallery(g, d, paddleSteamerBoiler, paddleSteamerHurr, paddleSteamerRail{from: -all, to: all, gap: 2.4, lamps: true})
	s.gallery(g, r, paddleSteamerHurr, paddleSteamerTexas, paddleSteamerRail{from: -all, to: all})
}

// pilotHouse is the glazed wheelhouse on the texas roof under a red dome.
func (s PaddleSteamer) pilotHouse(g *Grid) {
	const x0, x1, half = -1.0, 7.0, 3.5
	y0, y1 := paddleSteamerTexas, paddleSteamerTexas+4.6
	cx := (x0 + x1) / 2
	g.Use(White).Fill(Subtract(Block(V(x0, y0-0.3, -half), V(x1, y1, half)),
		Block(V(x0+1, y0-0.3, -half+1), V(x1-1, y1+1, half-1))))
	g.Paint(Glass, Func{Box{V(x0, y0+1.5, -half), V(x1, y0+3.7, half)}, func(p Vec3) bool {
		dx, dz := math.Abs(p.X-cx), math.Abs(p.Z)
		if dx > (x1-x0)/2-0.9 && dz > half-0.9 {
			return false // Corner posts.
		}
		return dx > 0.45 && dz > 0.45 // Mullions.
	}})
	g.Use(White).Fill(Block(V(x0-0.8, y1, -half-0.8), V(x1+0.8, y1+0.8, half+0.8)))
	dome := Ellipsoid{V(cx, y1+0.6, 0), V(4.6, 3.6, 4.1)}
	g.Use(Red).FillVolume(Intersect(dome, Block(V(x0-2, y1+0.7, -half-2), V(x1+2, y1+6, half+2))))
	g.Carve(Intersect(Ellipsoid{dome.Center, V(3.5, 2.5, 3.0)}, Block(V(x0-2, y1+0.9, -half-2), V(x1+2, y1+6, half+2))))
	g.Use(Gold).Fill(Cylinder{Center: P(cx, 0), Radius: 0.5, Y0: y1 + 3.8, Y1: y1 + 6.0}, Sphere(V(cx, y1+4.9, 0), 0.9))
}

// stacks are the twin smokestacks with gold bands, flared feathered crowns and the
// spreader bar between them.
func (s PaddleSteamer) stacks(g *Grid) {
	x, top := paddleSteamerStackX, paddleSteamerStackY
	for _, z := range []float64{-paddleSteamerStackZ, paddleSteamerStackZ} {
		c := P(x, z)
		g.Use(Charcoal).Fill(Cylinder{Center: c, Radius: 1.6, Y0: paddleSteamerMain, Y1: top - 4.5})
		g.FillVolume(Frustum{c, 1.6, 2.6, top - 4.5, top - 2})
		g.Use(Gold).Fill(Func{Box{V(x-3.6, top-2, z-3.6), V(x+3.6, top+0.2, z+3.6)}, func(p Vec3) bool {
			d := p.Plan().Sub(c)
			r, k := d.Length(), (p.Y-(top-2))/2.2
			if r < 1.5+0.3*k || r > 2.7+0.6*k {
				return false
			}
			return math.Cos(6*d.Angle()) > -0.35+0.9*k
		}})
		g.Carve(Cylinder{Center: c, Radius: 1.1, Y0: top - 3.2, Y1: top + 1})
		for _, y := range []float64{top - 6.4, top - 16} {
			g.Use(Gold).Fill(Cylinder{Center: c, Radius: 2.0, Inner: 1.0, Y0: y, Y1: y + 0.8})
		}
	}
	z := paddleSteamerStackZ
	g.Use(Charcoal).Fill(Block(V(x-0.5, 33.6, -z), V(x+0.5, 34.6, z)))
	g.Use(Gold).Fill(Sphere(V(x, 34.1, 0), 1.2))
}

// wheel is the sternwheel on its fantail timbers, with cranks and pitman arms.
func (s PaddleSteamer) wheel(g *Grid) {
	const half, n = 7.4, 10
	cx, cy, R := paddleSteamerWheelX, paddleSteamerWheelY, paddleSteamerWheelR
	step := 2 * math.Pi / n
	box := Box{V(cx-R-0.3, cy-R-0.3, -half), V(cx+R+0.3, cy+R+0.3, half)}
	g.Use(Red).Fill(Func{box, func(p Vec3) bool {
		dx, dy := p.X-cx, p.Y-cy
		r := math.Hypot(dx, dy)
		if r > R {
			return false
		}
		a := math.Atan2(dy, dx) + step/2
		off := math.Abs(a-step*math.Round(a/step)) * r // Arc distance to the nearest arm.
		if r >= 5.6 && off <= 0.5 {
			return true // Bucket plank across the whole wheel.
		}
		if math.Abs(p.Z-6.7*math.Round(p.Z/6.7)) > 0.55 {
			return false
		}
		// Flanges: arms, hub, a circle inside the buckets and a rim on the end flanges.
		return off <= 0.5 || r <= 1.7 || r >= 4.3 && r <= 5.2 || r >= R-0.9 && math.Abs(p.Z) > 3
	}})
	g.Use(Charcoal).Fill(Capsule{V(cx, cy, -10.6), V(cx, cy, 10.6), 0.8})
	// Fantail: cylinder timbers, a tie beam astern and the shaft bearings.
	for _, side := range []float64{-1, 1} {
		g.Use(White).Fill(Block(V(cx-R-1.8, 1.8, side*9.7), V(paddleSteamerStern+0.5, 3.2, side*10.9)))
		g.Use(Charcoal).Fill(Block(V(cx-1.2, 3.2, side*9.7), V(cx+1.2, cy+1.0, side*10.9)))
	}
	g.Use(White).Fill(Block(V(cx-R-2.0, 1.8, -10.9), V(cx-R-0.9, 3.2, 10.9)))
	// Cranks a quarter turn apart, driven by pitman arms from the engine room.
	for i, side := range []float64{1, -1} {
		a := 0.75 + float64(i)*math.Pi/2
		z := side * 8.6
		pin := V(cx+2.4*math.Cos(a), cy+2.4*math.Sin(a), z)
		s.line(g, Charcoal, 0.6, V(cx, cy, z), pin)
		s.line(g, White, 0.55, pin, V(-37, cy+0.4, z))
		g.Use(Charcoal).Fill(Block(V(-40, cy-0.3, side*6.8), V(-34.5, cy+1.1, side*9.2)))
		g.Use(Gold).Fill(Sphere(pin, 0.7))
	}
}

// plank is a sloping gangplank from a to b with low white curbs.
func (PaddleSteamer) plank(g *Grid, a, b Vec3, half float64) {
	ab := b.Sub(a)
	length := ab.Length()
	dir := ab.Scale(1 / length)
	hz := math.Hypot(dir.X, dir.Z)
	side := V(-dir.Z/hz, 0, dir.X/hz)
	up := V(-side.Z*dir.Y, side.Z*dir.X-side.X*dir.Z, side.X*dir.Y)
	local := func(p Vec3) (u, v, w float64) {
		q := p.Sub(a)
		return q.Dot(dir), math.Abs(q.Dot(side)), q.Dot(up)
	}
	box := BoxOf(a, b).Grow(half + 1.5)
	g.Use(Wood).Fill(Func{box, func(p Vec3) bool {
		u, v, w := local(p)
		return u >= 0 && u <= length && v <= half && math.Abs(w) <= 0.45
	}})
	g.Use(White).Fill(Func{box, func(p Vec3) bool {
		u, v, w := local(p)
		return u >= 0 && u <= length && v >= half-0.45 && v <= half+0.35 && w >= -0.45 && w <= 0.85
	}})
}

// bale is one cotton bale; a stack mixes burlap and bare cotton.
func (PaddleSteamer) bale(g *Grid, x, y, z float64, seed int) {
	col := Sand
	if Hash01(seed, 7) < 0.3 {
		col = White
	}
	g.With(col, func() { g.Fill(Block(V(x-1.15, y, z-0.85), V(x+1.15, y+1.5, z+0.85))) })
}

// bales stacks cotton in a stepped pile: nx by nz bales at the bottom, tiers high.
func (s PaddleSteamer) bales(g *Grid, x, y, z float64, nx, nz, tiers, seed int) {
	for t := 0; t < tiers; t++ {
		for i := 0; i < nx-t; i++ {
			for k := 0; k < nz; k++ {
				if t > 0 && Hash01(seed, t, i, k) < 0.2 {
					continue
				}
				s.bale(g, x+2.5*(float64(i)+0.5*float64(t)), y+1.6*float64(t), z+1.9*float64(k), seed*31+t*7+i*3+k)
			}
		}
	}
}

// foredeck carries the grand staircase, cotton, the capstan, the jackstaff and the
// twin landing stages on their booms.
func (s PaddleSteamer) foredeck(g *Grid) {
	boiler := paddleSteamerBoiler
	// Grand staircase up to the boiler deck's rounded front.
	stairs := Box{V(30.6, 3, -2.2), V(38.4, boiler+1.6, 2.2)}
	rise := func(x float64) float64 { return boiler - (x-31)*(boiler-3.3)/7.2 }
	g.Use(Wood).Fill(Func{stairs, func(p Vec3) bool {
		top := rise(p.X)
		return math.Abs(p.Z) < 1.6 && p.Y <= top && p.Y >= top-1.3
	}})
	g.Use(White).Fill(Func{stairs, func(p Vec3) bool {
		top := rise(p.X)
		return math.Abs(p.Z) >= 1.5 && p.Y <= top+1.3 && p.Y >= top-1.3
	}})
	// Cotton bales under the boiler deck's overhang and out on the bow.
	for _, side := range []float64{-1, 1} {
		s.bales(g, 22.5, paddleSteamerMain-0.2, side*5.6-0.95, 4, 2, 3, 3+int(side))
		s.bales(g, 37.6, s.deckY(39)-0.2, side*3.9, 2, 1, 2, 9+int(side))
	}
	// Capstan and bitts.
	deck := s.deckY(42)
	g.Use(Charcoal).Fill(Cylinder{Center: P(42, 0), Radius: 1.0, Y0: deck - 0.3, Y1: deck + 1.3})
	g.Use(Gold).Fill(Cylinder{Center: P(42, 0), Radius: 0.6, Y0: deck + 1.3, Y1: deck + 1.9})
	// Jackstaff with a gold ball and a red pennant streaming aft.
	deck = s.deckY(46.2)
	g.Use(White).Fill(Cylinder{Center: P(46.2, 0), Radius: 0.5, Y0: deck - 0.4, Y1: 20})
	g.Use(Gold).Fill(Sphere(V(46.2, 20.4, 0), 0.8), Sphere(V(46.2, 12, 0), 0.75))
	g.Use(Red).Fill(Func{Box{V(40, 16, -0.5), V(45.8, 19.4, 0.5)}, func(p Vec3) bool {
		return math.Abs(p.Y-17.7) <= 0.35+1.3*(p.X-40)/5.8
	}})
	// Landing stages: a king post and boom each side, the stage hanging in its falls.
	for _, side := range []float64{-1, 1} {
		mast := P(35, side*5.6)
		foot := s.deckY(35)
		head := mast.At(23)
		tip := V(53, 19, side*7.4)
		g.Use(White).Fill(Cylinder{Center: mast, Radius: 0.6, Y0: foot - 0.3, Y1: head.Y})
		g.Use(Gold).Fill(Sphere(head.Add(V(0, 0.5, 0)), 0.75))
		s.line(g, White, 0.6, mast.At(foot+1.2), tip)
		s.line(g, Charcoal, 0.5, head, tip)
		a, b := V(43, s.deckY(43)+0.5, side*3.1), V(59, 9.6, side*7.4)
		s.plank(g, a, b, 1.7)
		s.line(g, Charcoal, 0.5, tip, a.Add(b.Sub(a).Scale(0.625)).Add(V(0, 0.4, 0)))
	}
}

// fittings are the lifeboats, the steam calliope, the bell and the stern flag.
func (s PaddleSteamer) fittings(g *Grid) {
	texas, hurr := paddleSteamerTexas, paddleSteamerHurr
	// Lifeboats in chocks on the texas roof.
	for _, side := range []float64{-1, 1} {
		c := V(-13, texas+1.0, side*5.6)
		g.Use(Charcoal).Fill(Block(V(c.X-2.2, texas-0.2, c.Z-0.6), V(c.X-1.4, texas+0.5, c.Z+0.6)),
			Block(V(c.X+1.4, texas-0.2, c.Z-0.6), V(c.X+2.2, texas+0.5, c.Z+0.6)))
		g.Use(White).FillVolume(Ellipsoid{c, V(3.3, 1.05, 1.3)})
		g.Paint(Red, Block(V(c.X-3.5, c.Y-0.2, c.Z-1.5), V(c.X+3.5, c.Y+0.2, c.Z+1.5)))
	}
	// Steam calliope: a red chest with a rank of brass whistles and puffs of steam.
	g.Use(Red).Fill(Block(V(-26, texas-0.2, -1.6), V(-20.5, texas+1.0, 1.6)))
	for i := 0; i < 6; i++ {
		x := -25.5 + 0.9*float64(i)
		top := texas + 3.6 - 0.35*float64(i)
		for _, z := range []float64{-0.8, 0.8} {
			g.Use(Gold).Fill(Cylinder{Center: P(x, z), Radius: 0.45, Y0: texas + 1.0, Y1: top})
		}
	}
	g.Use(White).Fill(Sphere(V(-25.3, texas+4.6, 0), 1.1), Sphere(V(-23.9, texas+5.2, 0.3), 0.9))
	// The roof bell on the hurricane deck's forward end.
	for _, z := range []float64{-1.3, 1.3} {
		g.Use(Charcoal).Fill(Cylinder{Center: P(26.5, z), Radius: 0.5, Y0: hurr - 0.2, Y1: hurr + 3.0})
	}
	g.Use(Charcoal).Fill(Block(V(26, hurr+2.6, -1.8), V(27, hurr+3.3, 1.8)))
	g.Use(Gold).Fill(Sphere(V(26.5, hurr+1.7, 0), 0.9))
	// Verge staff and ensign at the stern, streaming aft.
	g.Use(White).Fill(Cylinder{Center: P(-37.6, 0), Radius: 0.5, Y0: hurr - 0.2, Y1: 29.4})
	g.Use(Gold).Fill(Sphere(V(-37.6, 29.8, 0), 0.75))
	flag := Box{V(-45.6, 23.6, -0.5), V(-38.2, 28.6, 0.5)}
	g.Use(Red).Fill(Cuboid(flag))
	g.Paint(White, Func{flag, func(p Vec3) bool { return int(math.Floor((p.Y-23.6)/0.82))%2 == 1 }})
	g.Paint(Navy, Block(V(-41.6, 26.0, -0.5), V(-38.2, 28.6, 0.5)))
}

// person is a tiny figure: a coat and a head.
func (PaddleSteamer) person(g *Grid, x, y, z float64, coat Color) {
	g.With(coat, func() { g.Fill(Capsule{V(x, y+0.1, z), V(x, y+1.4, z), 0.5}) })
	g.With(Sand, func() { g.Fill(Sphere(V(x, y+2.2, z), 0.5)) })
}

// crowd places passengers at the rails and deckhands forward.
func (s PaddleSteamer) crowd(g *Grid) {
	coats := []Color{Navy, Red, Charcoal, Wood, Terracotta, Gold, Slate, Brick}
	n := 0
	stand := func(x, y, z float64) {
		s.person(g, x, y, z, coats[n%len(coats)])
		n++
	}
	boiler, hurr, texas := paddleSteamerBoiler, paddleSteamerHurr, paddleSteamerTexas
	for _, x := range []float64{-33.5, -26, -18.3, -13.5, -3.2, 2, 6.8, 12.2} {
		stand(x, boiler, 8.6)
	}
	for _, x := range []float64{-29, -8.4, 9.6} {
		stand(x, boiler, -8.6)
	}
	// Forward promenade of the boiler deck and the hurricane deck above it.
	stand(18, boiler, 2)
	stand(25.5, boiler, -3.2)
	stand(26.4, boiler, 3.8)
	stand(17.5, hurr, 7.5)
	stand(22.4, hurr, 0.6)
	stand(29, hurr, -0.5)
	stand(-34.5, hurr, 5)
	for _, x := range []float64{-20.5, -6.2, 5.9} {
		stand(x, hurr, 6.4)
	}
	stand(-6, texas, 2.5)
	stand(-21, texas, 4)
	// Deckhands on the foredeck and a stoker by the boilers.
	stand(33.5, s.deckY(33.5), 1.2)
	stand(40, s.deckY(40), -0.6)
	stand(44.4, s.deckY(44.4), 0.6)
	stand(19, paddleSteamerMain, 0.5)
}

// bank is the far shore: a hollow levee skin with a muddy foreshore and a cobbled landing.
func (s PaddleSteamer) bank(g *Grid) {
	site := Box{V(-70, -0.5, -48), V(65, 7, paddleSteamerShore+4)}
	H := s.ground
	inLand := func(p Vec3) bool {
		h := H(p.X, p.Z)
		return h >= 0.25 && p.Y >= -0.5 && p.Y <= h
	}
	g.Use(Wood)
	s.shell(g, site, 1.2, func(p Vec3) bool { return p.Y < -0.5 || inLand(p) })
	top := func(p Vec3) bool { h := H(p.X, p.Z); return h >= 0.25 && p.Y > h-0.9 }
	g.Paint(Grass, Func{site, top})
	g.Paint(Foliage, Func{site, func(p Vec3) bool {
		return top(p) && paddleSteamerNoise(p.X*0.12, p.Z*0.12, 31) > 0.2
	}})
	g.Paint(Sand, Func{site, func(p Vec3) bool {
		return top(p) && H(p.X, p.Z) < 1.5+0.6*paddleSteamerNoise(p.X*0.2, p.Z*0.2, 32)
	}})
	// The landing: stone courses up the levee and a packed-earth yard on top.
	paved := func(p Vec3) bool {
		edge := 1.5 * paddleSteamerNoise(p.Z*0.2, 7.7, 33)
		return top(p) && p.X > 9+edge && p.X < 47+edge
	}
	g.Paint(Stone, Func{site, paved})
	g.Paint(DarkStone, Func{site, func(p Vec3) bool {
		return paved(p) && p.Y < 4 && int(math.Floor(p.Y/0.8+10))%2 == 0
	}})
	g.Paint(Sand, Func{site, func(p Vec3) bool {
		return paved(p) && p.Y >= 4 && paddleSteamerNoise(p.X*0.25, p.Z*0.25, 34) > -0.1
	}})
}

// cypress is a bald cypress: a buttressed trunk with knees and a feathery tiered crown.
func (s PaddleSteamer) cypress(g *Grid, c Vec2, height float64, seed int) {
	base := math.Max(s.ground(c.X, c.Y), 0) - 0.4
	g.With(Wood, func() {
		g.FillVolume(Frustum{c, 1.9, 0.8, base, base + 4.5})
		g.Fill(Cylinder{Center: c, Radius: 0.8, Y0: base + 4, Y1: base + height - 3})
		for k := 0; k < 3; k++ {
			q := Polar(c, 2.5+0.8*Hash01(seed, k, 1), Hash01(seed, k, 2)*2*math.Pi)
			g.Fill(Frustum{q, 0.75, 0.45, math.Max(s.ground(q.X, q.Y), 0) - 0.4, base + 1.8 + Hash01(seed, k, 3)})
		}
	})
	const tiers = 4
	for i := 0; i < tiers; i++ {
		t := float64(i) / (tiers - 1)
		y := base + height*(0.5+0.46*t)
		r := 3.9 - 2.3*t + 0.5*Hash01(seed, i, 4)
		col := Foliage
		if (i+seed)%2 == 1 {
			col = Grass
		}
		off := Polar(P(0, 0), 0.5, Hash01(seed, i, 5)*2*math.Pi)
		g.With(col, func() { g.FillVolume(Ellipsoid{c.Add(off).At(y), V(r, 1.4, r)}) })
	}
}

// willow is a weeping willow: a hollow domed crown hung with trailing curtains.
func (s PaddleSteamer) willow(g *Grid, c Vec2, seed int) {
	base := s.ground(c.X, c.Y) - 0.4
	g.With(Wood, func() { g.Fill(Cylinder{Center: c, Radius: 0.9, Y0: base, Y1: base + 6}) })
	crown := c.At(base + 8.2)
	g.With(Grass, func() { g.FillVolume(Ellipsoid{crown, V(5.2, 3.4, 5.2)}) })
	g.Carve(Ellipsoid{crown.Add(V(0, -0.4, 0)), V(4.0, 2.5, 4.0)})
	g.Paint(Foliage, Func{Ellipsoid{crown, V(5.4, 3.6, 5.4)}.Bounds(), func(p Vec3) bool {
		return paddleSteamerNoise(p.X*0.5+p.Y*0.3, p.Z*0.5, 40+seed) > 0.15
	}})
	for k := 0; k < 12; k++ {
		a := (float64(k) + 0.6*Hash01(seed, k, 1)) * 2 * math.Pi / 12
		q := Polar(c, 4.5+0.5*Hash01(seed, k, 2), a)
		drop := 4.2 + 2.6*Hash01(seed, k, 3)
		col := Grass
		if k%2 == 1 {
			col = Foliage
		}
		s.line(g, col, 0.55, q.At(crown.Y-0.6), q.At(crown.Y-0.6-drop))
	}
}

// landing builds the wharf with its freight, the cotton warehouse and the trees.
func (s PaddleSteamer) landing(g *Grid) {
	// Wharf: a timber deck on piles with mooring posts.
	const x0, x1, deck = 13.0, 43.0, 3.0
	const z0, z1 = paddleSteamerShore - 4.5, paddleSteamerShore + 6.5
	g.Use(Wood).Fill(Block(V(x0, deck-0.8, z0), V(x1, deck, z1)))
	g.Paint(Sand, Func{Box{V(x0, deck-0.4, z0), V(x1, deck, z1)}, func(p Vec3) bool {
		return math.Mod(p.X-x0, 3) < 0.8
	}})
	for x := x0 + 0.6; x <= x1; x += 4.8 {
		g.Use(Wood).Fill(Cylinder{Center: P(x, z1-0.6), Radius: 0.6, Y0: -0.5, Y1: deck + 1.2},
			Cylinder{Center: P(x, z1-4.5), Radius: 0.6, Y0: -0.5, Y1: deck})
	}
	s.bales(g, 15.5, deck-0.2, z1-5, 3, 2, 3, 21)
	s.bales(g, 25, deck-0.2, z1-8, 2, 2, 2, 22)
	for i := 0; i < 7; i++ {
		c := P(32.5+1.9*float64(i%4), z1-3-1.9*float64(i/4)-0.4*float64(i%2))
		col := Brick
		if i%3 == 1 {
			col = Wood
		}
		g.Use(col).Fill(Cylinder{Center: c, Radius: 0.8, Y0: deck - 0.2, Y1: deck + 1.6})
		g.Use(Charcoal).Fill(Cylinder{Center: c, Radius: 0.8, Y0: deck + 0.5, Y1: deck + 0.9})
	}
	for _, x := range []float64{x0 + 0.8, x1 - 0.8} {
		g.Use(Charcoal).Fill(Cylinder{Center: P(x, z0+1.5), Radius: 0.5, Y0: deck - 0.2, Y1: deck + 4.6})
		g.Use(Gold).Fill(Sphere(V(x, deck+5.1, z0+1.5), 0.8))
	}
	s.person(g, 23.5, deck, z1-3, Navy)
	s.person(g, 29.6, deck, z1-5.5, Red)
	s.person(g, 39.5, deck, z1-7, Charcoal)
	s.person(g, 20, deck, z1-9.1, Terracotta)

	// Cotton warehouse on the levee crown.
	const wx0, wx1, wy0, wy1 = 16.0, 38.0, 3.8, 11.0
	const wz0, wz1 = paddleSteamerShore - 16, paddleSteamerShore - 7
	g.Use(Brick).Fill(Subtract(Block(V(wx0, wy0, wz0), V(wx1, wy1, wz1)),
		Block(V(wx0+1, wy0-1, wz0+1), V(wx1-1, wy1+1, wz1-1))))
	g.Paint(Terracotta, Func{Box{V(wx0, wy0, wz0), V(wx1, wy1, wz1)}, func(p Vec3) bool {
		return int(math.Floor(p.Y/1.6))%2 == 0 && paddleSteamerNoise(p.X*0.5, p.Y*0.5+p.Z, 51) > 0.1
	}})
	roof := Gable{P(wx0-1, wz0-1), P(wx1+1, wz1+1), wy1, AxisX}
	g.Use(Slate).Fill(roof)
	g.Carve(Gable{P(wx0-1, wz0+0.4), P(wx1+1, wz1-0.4), wy1 - 0.4, AxisX})
	g.Use(Brick).Fill(Intersect(Gable{P(wx0, wz0), P(wx1, wz1), wy1, AxisX},
		Union(Block(V(wx0, wy1, wz0), V(wx0+1, wy1+8, wz1)), Block(V(wx1-1, wy1, wz0), V(wx1, wy1+8, wz1)))))
	g.Paint(DarkStone, Func{roof.Bounds(), func(p Vec3) bool {
		return p.Y > wy1 && math.Mod(p.X-wx0+40, 3.2) < 0.8
	}})
	for i, x := range []float64{20.5, 27, 33.5} {
		g.Paint(Charcoal, Arch{V(x, wy0, wz1-0.5), 1.5, 2.6, 1.6, AxisZ})
		if i == 1 {
			g.Paint(Wood, Arch{V(x, wy0, wz1-0.5), 1.5, 2.6, 1.6, AxisZ})
		}
	}
	g.Paint(White, Block(V(21, wy1-1.9, wz1-1.2), V(33, wy1-0.9, wz1+0.3)))
	g.Paint(Navy, Block(V(wx1-1.2, wy0+3, wz0+2.5), V(wx1+0.3, wy0+4.8, wz0+4)),
		Block(V(wx1-1.2, wy0+3, wz0+5), V(wx1+0.3, wy0+4.8, wz0+6.5)))
	g.Use(Brick).Fill(Block(V(19, wy1+2, wz0+2.8), V(20.6, wy1+7.6, wz0+4.4)))
	g.Use(Charcoal).Fill(Block(V(19, wy1+7.6, wz0+2.8), V(20.6, wy1+8.2, wz0+4.4)))

	// Bald cypress at the water's edge and willows on the levee, placed by X and by
	// distance behind the waterline; the tall ones stand clear of the boat's ends.
	for i, t := range [][3]float64{
		{-60, -4, 18}, {-56, -9, 23}, {-52.5, -1.5, 19}, {-49, -13, 25}, {-36, -4, 13},
		{-17, -14, 14}, {5, -16, 15}, {51, -6, 17}, {55, -1.6, 14},
	} {
		s.cypress(g, P(t[0], paddleSteamerShore+t[1]), t[2], i+1)
	}
	for i, c := range []Vec2{P(-27, -9.5), P(-6, -8.5), P(45, -12)} {
		s.willow(g, c.Add(P(0, paddleSteamerShore)), i+1)
	}
}

// river lays the water plate with current streaks, the bow wave, the wake and the
// spray thrown by the wheel.
func (s PaddleSteamer) river(g *Grid) {
	box := Box{V(-70, -0.5, -48), V(65, 0.6, 26)}
	hash := func(c Cell, k int) float64 { return Hash01(int(c.X), int(c.Z), k) }
	astern := paddleSteamerWheelX - paddleSteamerWheelR
	// surface returns the water level (0 or one layer proud) and the foam strength.
	surface := func(p Vec3) (float64, float64) {
		x, az := p.X, math.Abs(p.Z)
		h, foam := 0.0, 0.0
		hw := s.halfWidth(x, 0.5)
		near := 99.0
		if hw >= 0 {
			near = az - hw
		}
		if x > 30 && near < 1.9 { // Bow wave piled against the spoon bow.
			h, foam = 0.4, 0.45+0.5*(x-30)/18
		} else if near < 1.2 {
			foam = 0.3
		}
		if along := paddleSteamerBow + 1 - x; along > 1 && along < 62 { // Diverging arms.
			arm := 1.5 + along*0.42
			if math.Abs(az-arm) < 0.9+along*0.012 && near > 1.5 {
				if along < 34 {
					h = 0.4
				}
				foam = math.Max(foam, 0.2+0.7*(1-along/62))
			}
		}
		if back := astern - x; back > -1 { // Churned wake astern of the wheel.
			if az < 8.5+0.13*back {
				foam = math.Max(foam, 0.95-0.035*back)
				if math.Sin(back*1.15+0.5*math.Sin(p.Z*0.5)) > 0.2 {
					h = 0.4
				}
			}
		}
		return h, foam
	}
	s.cells(g, box, func(c Cell, p Vec3) {
		if g.Has(c) || !s.inPlate(p.X, p.Z) || s.ground(p.X, p.Z) > 0.6 {
			return
		}
		if w := s.halfWidth(p.X, 0.5); w >= 0 && math.Abs(p.Z) < w-0.8 {
			return // Inside the hull.
		}
		h, foam := surface(p)
		if p.Y > h+0.15 {
			return
		}
		col := Water
		if p.Y > h-0.25 {
			switch {
			case hash(c, 11) < foam:
				col = White
			case foam > 0 && hash(c, 12) < 0.3:
				col = Glass
			case paddleSteamerNoise(p.X*0.06, p.Z*0.6, 21) > 0.5:
				col = Glass // Long streaks drawn out by the current.
			case s.shore(p.X)-p.Z > -3.5 && hash(c, 13) < 0.5:
				col = Glass // Shallows under the bank.
			case paddleSteamerNoise(p.X*0.09+40, p.Z*0.7, 22) > 0.72:
				col = White
			}
		}
		s.put(g, c, col)
	})
	// Spray: a ragged heap of white water thrown up behind the buckets.
	heap := func(p Vec3) bool {
		back := astern + 0.6 - p.X
		if back < 0 || back > 10.5 {
			return false
		}
		k := 1 - p.Z*p.Z/90
		top := 6.2 * math.Exp(-back/4.2) * k * (0.75 + 0.35*paddleSteamerNoise(p.X*0.6, p.Z*0.6, 23))
		return k > 0 && p.Y <= top
	}
	g.Use(White)
	s.shell(g, Box{V(astern-11, 0.2, -9.5), V(astern+0.6, 7.5, 9.5)}, 1.2, heap)
}

// smoke trails aft from both stacks, merging into one paling plume.
func (s PaddleSteamer) smoke(g *Grid) {
	var puffs []Ellipsoid
	const n = 12
	for i := 0; i < n; i++ {
		t := float64(i) / (n - 1)
		x := paddleSteamerStackX - 1.0 - 38*math.Pow(t, 1.1)
		y := paddleSteamerStackY + 1.5 + 4.4*math.Pow(t, 0.6) + 1.1*math.Sin(t*9)*t
		r := 1.6 + 2.5*t + 0.9*Hash01(i, 60)*t
		spread := paddleSteamerStackZ * math.Max(0, 1-t*3.4)
		for _, side := range []float64{-1, 1} {
			if spread == 0 && side > 0 {
				continue
			}
			// The breeze carries the plume aft and off toward the bank.
			c := V(x, y, side*spread-9*t)
			puffs = append(puffs, Ellipsoid{c, V(r*1.2, r*0.85, r)})
			for k := 0; k < 4 && i > 1; k++ {
				a := Hash01(i, k, 61) * 2 * math.Pi
				rr := r * (0.5 + 0.35*Hash01(i, k, 62))
				off := V(r*0.9*math.Cos(a), r*0.7*(Hash01(i, k, 63)-0.25), r*0.8*math.Sin(a))
				puffs = append(puffs, Ellipsoid{c.Add(off), V(rr*1.1, rr*0.85, rr)})
			}
		}
	}
	b := EmptyBox()
	for _, e := range puffs {
		b = b.Union(e.Bounds())
	}
	in := func(p Vec3) bool {
		for _, e := range puffs {
			if e.Bounds().Contains(p) && e.Contains(p) {
				return true
			}
		}
		return false
	}
	g.Use(Charcoal)
	s.shell(g, b, 1.4, in)
	// Sooty at the stacks, paling to grey and white as it thins. Nothing else reaches
	// this high, so every cell above the crowns belongs to the plume.
	b.Min.Y = paddleSteamerStackY - 0.6
	age := func(p Vec3) float64 { return (paddleSteamerStackX - p.X) / 39 }
	tint := func(col, from Color, lo, hi float64, seed int) {
		s.cells(g, b, func(c Cell, p Vec3) {
			was, ok := g.ColorOf(c)
			k := (age(p) - lo) / (hi - lo)
			if ok && was == from && Hash01(int(c.X), int(c.Y), int(c.Z), seed) < k {
				s.put(g, c, col)
			}
		})
	}
	tint(DarkStone, Charcoal, 0.02, 0.35, 64)
	tint(Stone, DarkStone, 0.25, 0.7, 65)
	tint(White, Stone, 0.55, 1.1, 66)
}
