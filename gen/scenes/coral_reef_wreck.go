package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// CoralReefWreck is an underwater diorama with no water volume: the open air of the model
// is the sea. An elliptical plate of rippled sand, cut at the rim to show sand over stone
// strata, carries a rocky reef ridge around its back (-Z) edge, with a low pass behind
// the masthead. Across the middle lies a broken-backed wooden galleon, bow toward +X,
// heeled toward +Z so its side and decks face the hero camera: a hollow plank hull with a
// dark wale, gun ports and a gilded stern castle, ribs showing where the planking is gone,
// and a gap where the back broke and the bow section slewed away. The main mast still
// stands with its top, a cocked yard and a rag of sail; the foremast has fallen across
// the sand. Cannons, an anchor on its chain and an open chest of gold lie spilled in
// front, watched by a helmeted diver whose hose runs to a small yellow submersible. Brain
// corals, staghorn, sea fans, tube sponges and table corals crowd the ridge and two
// knolls, kelp groves flank the wreck, and a shark, a manta ray, a turtle, an octopus on
// the quarterdeck and three fish schools (a ring at the masthead, a bait ball and an arch
// fleeing the shark) swim above.
type CoralReefWreck struct{}

func init() { Register(CoralReefWreck{}) }

func (CoralReefWreck) Name() string  { return "coral_reef_wreck" }
func (CoralReefWreck) Title() string { return "Coral Reef Wreck" }

const (
	coralReefWreckA     = 57.0 // Seabed plate semi-axes.
	coralReefWreckB     = 45.0
	coralReefWreckFloor = -7.0 // Underside of the plate's edge skirt.
	coralReefWreckSkin  = 1.4  // Thickness of the hollow seabed skin.
	coralReefWreckPlank = 1.3  // Hull planking thickness.
)

var (
	// Crest line of the reef ridge around the back of the plate.
	coralReefWreckRidge    = []Vec2{P(-51, 7), P(-45.5, -12.5), P(-31, -27), P(-9, -35), P(14, -36), P(35, -28), P(47, -11.5)}
	coralReefWreckRidgeLen = coralReefWreckLength(coralReefWreckRidge)
	// Coral knolls on the open sand: x, z, radius, height.
	coralReefWreckKnolls = [][4]float64{{-33, 27, 10, 4.5}, {41, 23, 8, 3.5}}
	// Stations where the galleon's decks step.
	coralReefWreckSteps = [3]float64{-25, -12, 19}
)

// coralReefWreckField is the seabed height at a plan point.
type coralReefWreckField func(x, z float64) float64

func coralReefWreckSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func coralReefWreckLerp(a, b, t float64) float64 { return a + (b-a)*t }

// coralReefWreckNoise is smooth value noise in [-1, 1].
func coralReefWreckNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return coralReefWreckLerp(coralReefWreckLerp(a, b, sx), coralReefWreckLerp(c, d, sx), sz)*2 - 1
}

func coralReefWreckFBM(x, z float64, seed int) float64 {
	return 0.6*coralReefWreckNoise(x, z, seed) + 0.3*coralReefWreckNoise(x*2.03+17, z*2.03+5, seed+1) +
		0.1*coralReefWreckNoise(x*4.1+3, z*4.1+29, seed+2)
}

func coralReefWreckLength(pts []Vec2) float64 {
	total := 0.0
	for i := 0; i+1 < len(pts); i++ {
		total += pts[i+1].Sub(pts[i]).Length()
	}
	return total
}

// coralReefWreckPolyline returns the plan distance to a polyline and the arc length at the
// nearest point.
func coralReefWreckPolyline(p Vec2, pts []Vec2) (dist, along float64) {
	dist = math.Inf(1)
	run := 0.0
	for i := 0; i+1 < len(pts); i++ {
		a, b := pts[i], pts[i+1]
		ab := b.Sub(a)
		l := ab.Length()
		t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/(l*l)))
		if d := p.Sub(a.Add(ab.Scale(t))).Length(); d < dist {
			dist, along = d, run+t*l
		}
		run += l
	}
	return
}

func coralReefWreckUnit(v Vec3) Vec3 { return v.Scale(1 / math.Max(v.Length(), 1e-9)) }

// coralReefWreckFrame is a rigid local coordinate system: U forward, V up, W to the side,
// with a uniform scale K.
type coralReefWreckFrame struct {
	O, U, V, W Vec3
	K          float64
}

// coralReefWreckNewFrame yaws the frame in plan, pitches its nose up, then heels it over
// toward its W side.
func coralReefWreckNewFrame(o Vec3, yaw, pitch, heel float64) coralReefWreckFrame {
	u0 := V(math.Cos(yaw), 0, math.Sin(yaw))
	w0 := V(-math.Sin(yaw), 0, math.Cos(yaw))
	v0 := V(0, 1, 0)
	u := u0.Scale(math.Cos(pitch)).Add(v0.Scale(math.Sin(pitch)))
	v1 := v0.Scale(math.Cos(pitch)).Sub(u0.Scale(math.Sin(pitch)))
	v := v1.Scale(math.Cos(heel)).Add(w0.Scale(math.Sin(heel)))
	w := w0.Scale(math.Cos(heel)).Sub(v1.Scale(math.Sin(heel)))
	return coralReefWreckFrame{o, u, v, w, 1}
}

// coralReefWreckHeading points U along dir, keeps W horizontal, then rolls about U.
func coralReefWreckHeading(o, dir Vec3, roll float64) coralReefWreckFrame {
	u := coralReefWreckUnit(dir)
	w := coralReefWreckUnit(V(-u.Z, 0, u.X))
	v := V(w.Y*u.Z-w.Z*u.Y, w.Z*u.X-w.X*u.Z, w.X*u.Y-w.Y*u.X)
	c, s := math.Cos(roll), math.Sin(roll)
	return coralReefWreckFrame{o, u, v.Scale(c).Add(w.Scale(s)), w.Scale(c).Sub(v.Scale(s)), 1}
}

func (f coralReefWreckFrame) local(p Vec3) Vec3 {
	d := p.Sub(f.O)
	return V(d.Dot(f.U), d.Dot(f.V), d.Dot(f.W)).Scale(1 / f.K)
}

func (f coralReefWreckFrame) world(l Vec3) Vec3 {
	return f.O.Add(f.U.Scale(l.X * f.K)).Add(f.V.Scale(l.Y * f.K)).Add(f.W.Scale(l.Z * f.K))
}

// at places a solid authored in the frame's local coordinates into the world.
func (f coralReefWreckFrame) at(s Solid) Solid { return coralReefWreckPlaced{f, s} }

type coralReefWreckPlaced struct {
	F coralReefWreckFrame
	S Solid
}

func (s coralReefWreckPlaced) Contains(p Vec3) bool { return s.S.Contains(s.F.local(p)) }

func (s coralReefWreckPlaced) Bounds() Box {
	in := s.S.Bounds()
	b := EmptyBox()
	for _, x := range [2]float64{in.Min.X, in.Max.X} {
		for _, y := range [2]float64{in.Min.Y, in.Max.Y} {
			for _, z := range [2]float64{in.Min.Z, in.Max.Z} {
				q := s.F.world(V(x, y, z))
				b = b.Union(Box{q, q})
			}
		}
	}
	return b
}

// coralReefWreckTaper is a capsule whose radius runs from RA at A to RB at B.
type coralReefWreckTaper struct {
	A, B   Vec3
	RA, RB float64
}

func (c coralReefWreckTaper) Contains(p Vec3) bool {
	ab := c.B.Sub(c.A)
	t := math.Max(0, math.Min(1, p.Sub(c.A).Dot(ab)/math.Max(ab.Dot(ab), 1e-12)))
	return p.Sub(c.A.Add(ab.Scale(t))).Length() <= c.RA+(c.RB-c.RA)*t
}

func (c coralReefWreckTaper) Bounds() Box { return BoxOf(c.A, c.B).Grow(math.Max(c.RA, c.RB)) }

// coralReefWreckChain strings tapers through points with a radius at each.
func coralReefWreckChain(pts []Vec3, radii []float64) UnionSolid {
	u := make(UnionSolid, 0, len(pts))
	for i := 0; i+1 < len(pts); i++ {
		u = append(u, coralReefWreckTaper{pts[i], pts[i+1], radii[i], radii[i+1]})
	}
	return u
}

// coralReefWreckRock is how firmly a plan point belongs to the reef ridge or a knoll (0 on
// open sand, 1 on a crest) and the rock height there.
func coralReefWreckRock(x, z float64) (weight, height float64) {
	d, along := coralReefWreckPolyline(P(x, z), coralReefWreckRidge)
	w := 9 + 2*coralReefWreckNoise(along*0.07, 3, 8)
	ends := coralReefWreckSmooth(math.Min(along, coralReefWreckRidgeLen-along)/9 + 0.3)
	// A low sandy pass behind the masthead keeps the standing mast clear against open sea.
	pass := 1 - 0.8*coralReefWreckSmooth(1-math.Abs(along-38.5)/11)
	weight = coralReefWreckSmooth(1-d/w) * ends * pass
	height = weight * (10 + 3.5*coralReefWreckNoise(along*0.09, 0, 7) + 1.5*math.Sin(along*0.21))
	for _, n := range coralReefWreckKnolls {
		k := coralReefWreckSmooth(1 - math.Hypot(x-n[0], z-n[1])/n[2])
		weight = math.Max(weight, k)
		height = math.Max(height, k*n[3])
	}
	return
}

// coralReefWreckRipple is the sand ripple phase in [-1, 1]: crests run roughly east-west.
func coralReefWreckRipple(x, z float64) float64 {
	return math.Sin(0.95*(0.35*x+z) + 2.2*coralReefWreckNoise(x*0.07, z*0.07, 5))
}

func coralReefWreckRaw(x, z float64) float64 {
	weight, rock := coralReefWreckRock(x, z)
	h := 1.1 * coralReefWreckFBM(x*0.05, z*0.05, 3)
	h += 0.4 * (1 - weight) * coralReefWreckRipple(x, z)
	return h + rock + 1.3*weight*coralReefWreckFBM(x*0.24, z*0.24, 11)
}

// coralReefWreckBed memoizes seabed heights per sample point.
type coralReefWreckBed struct{ cache map[[2]int64]float64 }

func (b *coralReefWreckBed) height(x, z float64) float64 {
	k := [2]int64{int64(math.Round(x * 8)), int64(math.Round(z * 8))}
	if h, ok := b.cache[k]; ok {
		return h
	}
	h := coralReefWreckRaw(x, z)
	b.cache[k] = h
	return h
}

// Hull profile in ship coordinates: x from the stern (-35) to the stem (+35), y up from
// the keel, z athwartships with +z the side that faces the camera.

// coralReefWreckSheer is the top of the bulwark.
func coralReefWreckSheer(x float64) float64 {
	switch {
	case x < -25:
		return 17.5
	case x < -12:
		return 14.5
	case x < 19:
		return 11
	}
	return 13.5 + 0.08*(x-19)
}

// coralReefWreckDeck is the weather deck: poop, quarterdeck, waist and forecastle.
func coralReefWreckDeck(x float64) float64 {
	switch {
	case x < -25:
		return 15.5
	case x < -12:
		return 12.5
	case x < 19:
		return 8.5
	}
	return 11.5 + 0.08*(x-19)
}

// coralReefWreckBeam is the hull's half-width at a station and height, or negative
// outside the hull's length.
func coralReefWreckBeam(x, y float64) float64 {
	if y < 0 {
		return -1
	}
	xe := x
	if x > 0 {
		xe += 0.45 * math.Max(0, 10-y) // Raked stem.
	} else {
		xe -= 0.3 * math.Max(0, 9-y) // Counter under the stern.
	}
	var b float64
	switch {
	case xe > 35 || xe < -35:
		return -1
	case xe > 8:
		b = 7.5 * (1 - math.Pow((xe-8)/27, 2.3))
	case xe < -10:
		t := (-10 - xe) / 25
		b = 7.5 - 2.3*t*t
	default:
		b = 7.5
	}
	s := math.Pow(math.Min(1, y/6.5), 0.42)
	if y > 9 {
		s -= 0.035 * (y - 9) // Tumblehome.
	}
	return b * s
}

// coralReefWreckHull is one surviving section of the galleon.
type coralReefWreckHull struct {
	F      coralReefWreckFrame
	X0, X1 float64      // Surviving length; the break is ragged.
	Seed   int          // Varies the weathering and the ragged edges.
	Gaps   [][6]float64 // Blobs where the planking is gone: center x, y, z, then radii.
}

func (h coralReefWreckHull) kept(l Vec3) bool {
	n := 2.2 * coralReefWreckNoise(l.Y*0.33+float64(h.Seed)*5, l.Z*0.33, 17)
	low := 0.7 * math.Max(0, 4.5-l.Y) // The bottom planks reach further into the break.
	return l.X >= h.X0+n-low && l.X <= h.X1+n+low
}

func (h coralReefWreckHull) rib(l Vec3) bool { return math.Mod(l.X+70, 3.5) < 1.15 }

func (h coralReefWreckHull) gone(l Vec3) bool {
	m := -1.0
	for _, b := range h.Gaps {
		dx, dy, dz := (l.X-b[0])/b[3], (l.Y-b[1])/b[4], (l.Z-b[2])/b[5]
		m = math.Max(m, 1-math.Sqrt(dx*dx+dy*dy+dz*dz))
	}
	return m+0.3*coralReefWreckNoise(l.X*0.23+9, (l.Y+0.7*l.Z)*0.23, 31+h.Seed) > 0.12
}

// solid is the hollow hull: planked sides, decks, bulkheads and keel, with only the ribs
// and deck beams left where the planking is gone.
func (h coralReefWreckHull) solid(l Vec3) bool {
	if !h.kept(l) {
		return false
	}
	x, y, z := l.X, l.Y, math.Abs(l.Z)
	if y >= -1.2 && y <= 1.4 && z <= 0.85 && x >= -32 && x <= 30 {
		return true // Keel.
	}
	beam := coralReefWreckBeam(x, y)
	if y > coralReefWreckSheer(x) || z > beam {
		return false
	}
	deck := coralReefWreckDeck(x)
	side := z > beam-coralReefWreckPlank || y < coralReefWreckPlank || coralReefWreckBeam(x-coralReefWreckPlank, y) < 0
	floor := y <= deck && y > deck-1.1
	bulk := false
	for _, step := range coralReefWreckSteps {
		if math.Abs(x-step) <= 0.65 && y > 7.5 && y <= math.Max(coralReefWreckDeck(x-1), coralReefWreckDeck(x+1)) {
			bulk = true
		}
	}
	if !side && !floor && !bulk {
		return false
	}
	if h.gone(l) {
		return (side || floor) && h.rib(l)
	}
	return true
}

// tint colors a hull cell: grey deck planks, a dark wale, gun ports, the gilded stern
// castle with its cabin windows, and weathered strakes.
func (h coralReefWreckHull) tint(l Vec3) Color {
	x, y, z := l.X, l.Y, math.Abs(l.Z)
	if y < 0 {
		return DarkStone
	}
	deck, sheer := coralReefWreckDeck(x), coralReefWreckSheer(x)
	if z <= coralReefWreckBeam(x, y)-coralReefWreckPlank && coralReefWreckBeam(x-coralReefWreckPlank, y) >= 0 {
		if y <= deck && y > deck-1.1 {
			plank := int(math.Floor(l.Z / 1.2))
			if Hash01(plank, int(math.Floor((x+1.7*float64(plank))/6)), h.Seed) < 0.22 {
				return DarkStone
			}
			return Stone
		}
		return Wood
	}
	switch {
	case y > sheer-0.9 && x < -25:
		return Gold
	case y > sheer-0.9:
		return DarkStone
	case x < -26 && x > -34 && y > 13 && y <= 14.8 && math.Mod(x+40, 2.4) < 1.3:
		return Glass
	case x < -25 && y > 12 && y <= 13:
		return Gold
	case y > 4 && y < 5.1:
		return Charcoal
	case y > 5.9 && y < 7.6 && x > -31 && x < 17 && math.Mod(x+40, 6) < 1.8:
		return Charcoal
	}
	strake := int(math.Floor(y / 1.1))
	r := Hash01(strake, int(math.Floor((x+2.3*float64(strake))/5.5)), h.Seed, 5)
	switch {
	case r < 0.07:
		return Charcoal
	case r < 0.23+0.25*coralReefWreckSmooth((3.5-y)/3):
		return DarkStone
	}
	return Wood
}

func (h coralReefWreckHull) build(g *Grid) {
	box := Box{V(math.Max(h.X0-5, -36), -1.5, -8.5), V(math.Min(h.X1+5, 36), 19, 8.5)}
	g.Use(Wood).Fill(h.F.at(Func{box, h.solid}))
	for _, c := range []Color{Stone, DarkStone, Charcoal, Gold, Glass} {
		c := c
		g.Paint(c, h.F.at(Func{box, func(l Vec3) bool { return h.solid(l) && h.tint(l) == c }}))
	}
}

func (s CoralReefWreck) Build(g *Grid) Order {
	bed := &coralReefWreckBed{cache: map[[2]int64]float64{}}
	H := coralReefWreckField(bed.height)
	s.seabed(g, H)

	yaw := -0.26
	stern := coralReefWreckHull{
		F:  coralReefWreckNewFrame(V(-2, -0.8, -5), yaw, 0, 0.31),
		X0: -40, X1: 2, Seed: 1,
		Gaps: [][6]float64{{2, 9, 0, 8, 10, 12}, {-5, 7, 7.5, 6, 4, 4}, {-8, 8.5, -2, 5, 2.5, 4}},
	}
	// The bow section broke off, slewed toward the camera, nosed into the sand and fell
	// further onto its side.
	bf := coralReefWreckNewFrame(V(0, 0, 0), yaw+0.3, -0.07, 0.48)
	bf.O = stern.F.world(V(10.5, 0, 3.5)).Sub(bf.U.Scale(6))
	bf.O.Y = 0.6
	bow := coralReefWreckHull{
		F:  bf,
		X0: 6, X1: 40, Seed: 2,
		Gaps: [][6]float64{{6, 9, 0, 7, 10, 12}, {19, 7, 6.5, 7, 4.5, 4}, {27, 12, 0, 4, 2.5, 4}},
	}
	stern.build(g)
	bow.build(g)

	s.rig(g, H, stern, bow)
	s.spill(g, H, bow)
	s.reef(g, H)
	s.weeds(g, H)
	s.life(g, stern)
	s.explorers(g, H)
	return Along(P(0.97, -0.26))
}

// seabed is a hollow skin over the height field, closed at the rim by a skirt that shows
// sand over stone strata.
func (s CoralReefWreck) seabed(g *Grid, H coralReefWreckField) {
	box := Box{V(-coralReefWreckA, coralReefWreckFloor, -coralReefWreckB), V(coralReefWreckA, 18, coralReefWreckB)}
	rim := func(p Vec3) float64 {
		return (1 - math.Hypot(p.X/coralReefWreckA, p.Z/coralReefWreckB)) * coralReefWreckB
	}
	g.Use(Sand).Fill(Func{box, func(p Vec3) bool {
		e := rim(p)
		if e < 0 {
			return false
		}
		h := H(p.X, p.Z)
		if p.Y > h {
			return false
		}
		if e < 1.5 || p.Y >= h-coralReefWreckSkin {
			return true
		}
		// Keep steep rock faces closed.
		lo := math.Min(math.Min(H(p.X+1.1, p.Z), H(p.X-1.1, p.Z)), math.Min(H(p.X, p.Z+1.1), H(p.X, p.Z-1.1)))
		return p.Y >= lo-0.3
	}})
	tint := func(p Vec3) Color {
		if rim(p) < 1.7 && p.Y < -1.2 {
			y := p.Y + 0.7*coralReefWreckNoise(p.X*0.11+p.Z*0.07, 0.5, 41)
			switch {
			case y > -2.4:
				return Sand
			case y > -3.6:
				return Stone
			case y > -5.6:
				return DarkStone
			case y > -6.5:
				return Stone
			}
			return DarkStone
		}
		weight, _ := coralReefWreckRock(p.X, p.Z)
		if weight > 0.3 {
			n := coralReefWreckFBM(p.X*0.2, p.Z*0.2, 23) + 0.25*math.Sin(p.Y*1.1)
			switch {
			case H(p.X, p.Z)-p.Y < 1 && n > 0.25:
				return Stone
			case n < -0.3:
				return Slate
			}
			return DarkStone
		}
		if coralReefWreckRipple(p.X, p.Z) < -0.62 && coralReefWreckNoise(p.X*0.05, p.Z*0.05, 29) > -0.15 {
			return Stone
		}
		return Sand
	}
	for _, c := range []Color{Stone, DarkStone, Slate} {
		c := c
		g.Paint(c, Func{box, func(p Vec3) bool { return tint(p) == c }})
	}
}

// rig raises the surviving main mast with its top, yard and sail scrap, the broken stumps,
// the bowsprit, and the foremast lying across the sand.
func (s CoralReefWreck) rig(g *Grid, H coralReefWreckField, stern, bow coralReefWreckHull) {
	f := stern.F
	mast := P(-7, 0)
	g.Use(Wood).Fill(f.at(Frustum{mast, 1.15, 0.75, 1, 40}))
	g.Use(DarkStone).Fill(f.at(Cylinder{mast, 2.9, 0, 28, 29}), f.at(Cylinder{mast, 2.9, 2, 29, 30.6}))
	ya, yb := V(-16, 25.5, -2.5), V(2, 22.5, 2.5)
	g.Use(Wood).Fill(f.at(Capsule{ya, yb, 0.75}))
	sail := Func{Box{V(-16, 8, -5), V(2, 26, 6)}, func(l Vec3) bool {
		t := (l.X - ya.X) / (yb.X - ya.X)
		if t < 0.04 || t > 0.96 {
			return false
		}
		drop := ya.Y + (yb.Y-ya.Y)*t - l.Y
		if drop < 0 || drop > 4+9*math.Abs(coralReefWreckNoise(t*3.3, 0.5, 61)) {
			return false
		}
		belly := 1.6 * math.Sin(math.Pi*t) * math.Min(1, drop/6)
		if math.Abs(l.Z-(ya.Z+(yb.Z-ya.Z)*t+belly)) > 0.6 {
			return false
		}
		return drop < 1.2 || coralReefWreckNoise(l.X*0.45+7, l.Y*0.45, 63) < 0.45
	}}
	g.Use(White).Fill(f.at(sail))
	g.Paint(Stone, f.at(Func{sail.Box, func(l Vec3) bool {
		return sail.Contains(l) && coralReefWreckNoise(l.X*0.3, l.Y*0.3, 67) > 0.35
	}}))
	// Shrouds still holding the mast.
	g.Use(Charcoal).Fill(
		f.at(Capsule{V(-7, 28, 2.4), V(-11, 11, 7), 0.5}), f.at(Capsule{V(-7, 28, 2.4), V(-14.5, 14.5, 6), 0.5}),
		f.at(Capsule{V(-7, 28, -2.4), V(-11, 11, -7), 0.5}), f.at(Capsule{V(-7, 28, -2.4), V(-3.5, 11, -7), 0.5}),
	)
	// Mizzen stump and the stern lantern.
	g.Use(Wood).Fill(f.at(Frustum{P(-21, 0), 0.95, 0.85, 12, 19.5}), f.at(Cylinder{P(-34.3, 0), 0.6, 0, 17, 19.5}))
	g.Use(Gold).Fill(f.at(Sphere(V(-34.3, 20.2, 0), 1)))

	b := bow.F
	g.Use(Wood).Fill(b.at(Frustum{P(22, 0), 1, 0.9, 11, 15.5}), b.at(Capsule{V(32, 12.5, 0), V(43, 17, 0), 0.8}))
	g.Use(Gold).Fill(b.at(Ellipsoid{V(35.6, 11.4, 0), V(1.4, 1.5, 1.1)}))

	// The fallen foremast, its top and its yard.
	top := b.world(V(22, 15.5, 1.5))
	foot := V(top.X+3, 0, top.Z+24)
	foot.Y = H(foot.X, foot.Z) + 0.9
	g.Use(Wood).Fill(Capsule{top, foot, 0.9})
	m := coralReefWreckHeading(top, foot.Sub(top), 0)
	length := foot.Sub(top).Length()
	g.Use(DarkStone).Fill(m.at(Func{Box{V(length*0.7-0.55, -2.5, -2.5), V(length*0.7+0.55, 2.5, 2.5)},
		func(l Vec3) bool { return math.Hypot(l.Y, l.Z) <= 2.5 }}))
	yc := m.world(V(length*0.87, 0, 0))
	p1, p2 := yc.Add(m.W.Scale(-5.5)), yc.Add(m.W.Scale(5.5))
	p1.Y, p2.Y = H(p1.X, p1.Z)+0.8, H(p2.X, p2.Z)+0.8
	g.Use(Wood).Fill(Capsule{p1, p2, 0.75})
	rag := yc.Add(m.W.Scale(-3)).Add(m.U.Scale(-3))
	g.Use(White).Fill(Ellipsoid{V(rag.X, H(rag.X, rag.Z)+0.4, rag.Z), V(3.2, 0.7, 2.2)})
}

// cannon lays a gun on the sand; mounted guns keep their wooden carriage.
func (s CoralReefWreck) cannon(g *Grid, H coralReefWreckField, c Vec2, yaw float64, mounted bool) {
	at, pitch := c.At(H(c.X, c.Y)+0.8), 0.0
	if mounted {
		at.Y += 0.9
		pitch = 0.16
	}
	f := coralReefWreckHeading(at, V(math.Cos(yaw)*math.Cos(pitch), math.Sin(pitch), math.Sin(yaw)*math.Cos(pitch)), 0)
	if mounted {
		g.Use(Wood).Fill(f.at(Block(V(-2.4, -1.9, -1.4), V(0.8, -0.6, 1.4))))
	}
	g.Use(Charcoal).Fill(f.at(Union(
		coralReefWreckTaper{V(-2.3, 0, 0), V(2.9, 0, 0), 1.05, 0.7},
		Sphere(V(3, 0, 0), 0.95), Sphere(V(-2.9, 0, 0), 0.7),
	)))
}

// spill scatters what fell out of the broken hull: guns, the anchor on its chain and the
// treasure chest.
func (s CoralReefWreck) spill(g *Grid, H coralReefWreckField, bow coralReefWreckHull) {
	s.cannon(g, H, P(7, 7.5), 0.4, false)
	s.cannon(g, H, P(13.5, 11.5), 2.1, true)
	s.cannon(g, H, P(3.5, 12.5), -0.9, true)
	s.cannon(g, H, P(20, 12), 1.2, false)

	// Starfish on the open sand.
	for i, c := range []Vec2{P(-24, 14), P(15, 22), P(-3, 24), P(31, 17)} {
		y := H(c.X, c.Y) + 0.4
		g.Use(Terracotta)
		if i%2 == 1 {
			g.Use(Red)
		}
		for k := 0; k < 5; k++ {
			tip := Polar(c, 2.5, 2*math.Pi*float64(k)/5+float64(i))
			g.Fill(Capsule{c.At(y), tip.At(H(tip.X, tip.Y) + 0.3), 0.55})
		}
	}

	// Anchor lying flat, propped on its wooden stock.
	ring := P(41.5, 5.5)
	a := coralReefWreckHeading(ring.At(H(ring.X, ring.Y)+0.2), V(0.85, 0, 0.53), 0)
	g.Use(Charcoal).Fill(a.at(Union(
		Capsule{V(0.8, 0.8, 0), V(9.5, 0.8, 0), 0.7},
		Capsule{V(9.5, 0.8, 0), V(9.1, 0.8, 2.4), 0.7}, Capsule{V(9.1, 0.8, 2.4), V(7.4, 0.8, 4.2), 0.7},
		Capsule{V(9.5, 0.8, 0), V(9.1, 0.8, -2.4), 0.7}, Capsule{V(9.1, 0.8, -2.4), V(7.4, 0.8, -4.2), 0.7},
		Ellipsoid{V(7, 0.9, 4.4), V(1.3, 0.8, 1.2)}, Ellipsoid{V(7, 0.9, -4.4), V(1.3, 0.8, 1.2)},
	)))
	g.Use(Wood).Fill(a.at(Capsule{V(1.6, -0.3, 0), V(1.6, 4.6, 0), 0.7}))
	g.Use(DarkStone).Fill(a.at(Sphere(V(0, 0.8, 0), 1)))
	hawse, end := bow.F.world(V(31, 9.5, 3.2)), a.world(V(0, 0.8, 0))
	prev := hawse
	for i := 1; i <= 16; i++ {
		t := float64(i) / 16
		p := hawse.Add(end.Sub(hawse).Scale(t))
		ground := H(p.X, p.Z) + 0.6
		p.Y = ground + (hawse.Y-ground)*math.Pow(1-t, 3)
		g.Use(Charcoal)
		if i%2 == 0 {
			g.Use(DarkStone)
		}
		g.Fill(Capsule{prev, p, 0.55})
		prev = p
	}

	// Open chest heaped with gold, lid thrown back, coins spilled toward the camera.
	c := P(-2, 14.5)
	f := coralReefWreckHeading(c.At(H(c.X, c.Y)-0.4), V(math.Cos(-0.5), 0, math.Sin(-0.5)), 0.08)
	body := Block(V(-2.7, 0, -1.9), V(2.7, 3, 1.9))
	lid := Func{Box{V(-2.7, 2.9, -4.6), V(2.7, 6.2, -0.8)}, func(l Vec3) bool {
		back := -1.9 - (l.Y-2.9)*0.45
		return l.Z <= back+0.1 && l.Z >= back-1
	}}
	g.Use(Wood).Fill(f.at(Subtract(body, Block(V(-1.7, 1.1, -0.9), V(1.7, 3.5, 0.9)))), f.at(lid))
	g.Paint(Gold, f.at(Intersect(Union(body, lid),
		Union(Block(V(-2.8, 0.6, -6), V(-2, 8, 3)), Block(V(-0.45, 0.6, -6), V(0.45, 8, 3)), Block(V(2, 0.6, -6), V(2.8, 8, 3))))))
	g.Use(Gold).Fill(f.at(Ellipsoid{V(0, 2.5, 0), V(1.65, 1.5, 0.9)}), f.at(Ellipsoid{V(0.4, 0.5, 3.4), V(2.4, 0.7, 1.5)}))
	g.Paint(Red, Sphere(f.world(V(-0.8, 3.7, 0.3)), 0.5), Sphere(f.world(V(1.6, 1, 3.6)), 0.5))
	g.Paint(Glass, Sphere(f.world(V(0.7, 3.8, -0.2)), 0.5), Sphere(f.world(V(-0.9, 1, 3.9)), 0.5))
	g.Paint(White, Sphere(f.world(V(0, 4, 0.4)), 0.5))
}

// brain is a domed brain coral with meandering grooves.
func (s CoralReefWreck) brain(g *Grid, H coralReefWreckField, c Vec2, r float64, col, groove Color) {
	base := c.At(H(c.X, c.Y) - 0.6)
	var dome Solid = Ellipsoid{base, V(r, r*0.85, r)}
	if r > 4 {
		dome = Subtract(dome, Ellipsoid{base, V(r-1.6, r*0.85-1.6, r-1.6)})
	}
	dome = Intersect(dome, Block(V(c.X-r, base.Y, c.Y-r), V(c.X+r, base.Y+r, c.Y+r)))
	g.Use(col).FillVolume(dome)
	g.Paint(groove, Func{dome.Bounds(), func(p Vec3) bool {
		d := p.Sub(base)
		return dome.Contains(p) && math.Sin(2.1*d.X+1.7*math.Sin(1.1*d.Z+0.8*d.Y)) > 0.45
	}})
}

// branch grows one staghorn limb and forks it.
func (s CoralReefWreck) branch(g *Grid, a, dir Vec3, length, r float64, depth int, col, tip Color, seed int) {
	b := a.Add(dir.Scale(length))
	if depth == 0 {
		g.Use(tip).Fill(Capsule{a, b, r})
		return
	}
	g.Use(col).Fill(Capsule{a, b, r})
	turn := 2 * math.Pi * Hash01(seed, depth)
	for i := 0; i < 2; i++ {
		ang := turn + math.Pi*float64(i) + 0.8*Hash01(seed, depth, i)
		next := coralReefWreckUnit(dir.Add(V(0.75*math.Cos(ang), 0.1, 0.75*math.Sin(ang))))
		s.branch(g, b, next, length*0.8, math.Max(0.55, r*0.82), depth-1, col, tip, seed*7+i+1)
	}
}

// staghorn is a thicket of branching coral.
func (s CoralReefWreck) staghorn(g *Grid, H coralReefWreckField, c Vec2, size float64, col, tip Color, seed int) {
	base := c.At(H(c.X, c.Y) - 0.5)
	for i := 0; i < 4; i++ {
		a := 2 * math.Pi * (float64(i) + Hash01(seed, i)) / 4
		s.branch(g, base, coralReefWreckUnit(V(0.55*math.Cos(a), 1, 0.55*math.Sin(a))), size, 0.9, 2, col, tip, seed*13+i)
	}
}

// fan is a sea fan: a flat lattice of veins on a short stalk, facing across yaw.
func (s CoralReefWreck) fan(g *Grid, H coralReefWreckField, c Vec2, yaw, r float64, col Color) {
	f := coralReefWreckHeading(c.At(H(c.X, c.Y)-0.5), V(math.Cos(yaw), 0, math.Sin(yaw)), 0)
	g.Use(col).Fill(f.at(Func{Box{V(-r-1, 0, -0.6), V(r+1, r+2.5, 0.6)}, func(l Vec3) bool {
		if l.Y < 2 {
			return math.Abs(l.X) <= 0.85
		}
		d, a := math.Hypot(l.X, l.Y), math.Atan2(l.X, l.Y)
		edge := r * (1 + 0.1*math.Cos(5*a))
		if d > edge+1 || math.Abs(a) > 1.2 {
			return false
		}
		return d > edge || math.Abs(math.Remainder(a, 0.4))*d < 0.55 || math.Abs(math.Remainder(d, 2.6)) < 0.55
	}}))
}

// tubes is a cluster of tube sponges with dark throats.
func (s CoralReefWreck) tubes(g *Grid, H coralReefWreckField, c Vec2, n int, col Color, seed int) {
	for i := 0; i < n; i++ {
		p := c
		if i > 0 {
			p = Polar(c, 2.4+0.6*Hash01(seed, i, 2), 2*math.Pi*float64(i)/float64(n-1)+Hash01(seed, i))
		}
		rad := 1.5 + 0.4*Hash01(seed, i, 3)
		y := H(p.X, p.Y) - 0.6
		top := y + 5 + 4.5*Hash01(seed, i, 4)
		g.Use(col).Fill(Cylinder{p, rad, 0, y, top - 1.6}, Cylinder{p, rad, rad - 0.9, top - 1.6, top})
		g.Use(Charcoal).Fill(Cylinder{p, rad - 0.9, 0, top - 2.3, top - 1.6})
	}
}

// table is a plate coral on a stout stalk.
func (s CoralReefWreck) table(g *Grid, H coralReefWreckField, c Vec2, r float64, col Color) {
	y := H(c.X, c.Y) - 0.6
	g.Use(DarkStone).Fill(Frustum{c, 1.1, 1.7, y, y + 3})
	g.Use(col).FillVolume(Ellipsoid{c.At(y + 3.5), V(r, 0.9, r)})
}

// reef plants the corals in deliberate clumps: along the ridge crest and on the knolls.
func (s CoralReefWreck) reef(g *Grid, H coralReefWreckField) {
	// Ridge crest, west to east.
	s.fan(g, H, P(-49, -1), -0.4, 5.5, Brick)
	s.fan(g, H, P(-45, -10), -0.55, 5.5, Red)
	s.brain(g, H, P(-41, -14.5), 4, Terracotta, Brick)
	// (The pass behind the masthead stays bare.)
	s.tubes(g, H, P(-22.5, -28), 4, Glass, 2)
	s.brain(g, H, P(-16, -30), 4.6, Gold, Terracotta)
	s.fan(g, H, P(-9.5, -32), -0.6, 6.5, Terracotta)
	s.staghorn(g, H, P(-3.5, -32.5), 3.4, Red, White, 3)
	s.table(g, H, P(3.5, -32.5), 4.8, White)
	s.brain(g, H, P(10, -33.5), 4.4, Brick, Terracotta)
	s.tubes(g, H, P(17, -32.5), 5, Gold, 4)
	s.fan(g, H, P(23, -30), -0.8, 6.5, Red)
	s.staghorn(g, H, P(29.5, -27), 3, White, Glass, 5)
	s.brain(g, H, P(35.5, -22.5), 4.2, Terracotta, Gold)
	s.tubes(g, H, P(40, -18), 4, Terracotta, 6)
	s.table(g, H, P(44.5, -11.5), 4, Glass)
	// Western knoll.
	s.brain(g, H, P(-34, 27.5), 4.6, Gold, Terracotta)
	s.staghorn(g, H, P(-33, 20.5), 2.8, Red, White, 7)
	s.tubes(g, H, P(-39.5, 22), 4, Glass, 8)
	s.brain(g, H, P(-38.5, 28), 3, Brick, Red)
	// Eastern knoll.
	s.brain(g, H, P(42, 22.5), 4, Brick, Terracotta)
	s.staghorn(g, H, P(36.5, 26), 2.6, Gold, White, 9)
	s.tubes(g, H, P(44.5, 19), 3, Glass, 10)
	s.fan(g, H, P(37.5, 30), -0.9, 4.2, Red)
}

// kelp is one swaying strand with alternating blades.
func (s CoralReefWreck) kelp(g *Grid, H coralReefWreckField, c Vec2, height float64, seed int) {
	base := c.At(H(c.X, c.Y) - 0.5)
	phase := 2 * math.Pi * Hash01(seed, 1)
	sway := Polar(P(0, 0), 1, 2*math.Pi*Hash01(seed, 2))
	prev := base
	for i := 1; float64(i)*2.2 <= height; i++ {
		y := float64(i) * 2.2
		off := 1.5 * math.Sin(y*0.24+phase) * math.Min(1, y/6)
		p := V(base.X+sway.X*off+0.1*y, base.Y+y, base.Z+sway.Y*off)
		g.Use(Foliage).Fill(Capsule{prev, p, 0.7})
		if i >= 2 {
			side := 2.4
			if i%2 == 0 {
				side = -2.4
			}
			g.Use(Grass)
			if i%3 == 0 {
				g.Use(Foliage)
			}
			g.Fill(Capsule{p, p.Add(V(sway.Y*side, 1.6, -sway.X*side)), 0.6})
		}
		prev = p
	}
}

// weeds raises the kelp groves behind the wreck and tufts of sea grass on the sand.
func (s CoralReefWreck) weeds(g *Grid, H coralReefWreckField) {
	kelp := [][3]float64{
		{-43, 0, 20}, {-41.5, -5, 17}, {-38.5, -9.5, 15}, {-46.5, 4, 14}, // West grove.
		{30, -19, 26}, {34, -15, 26}, {38, -12, 26}, {42.5, -10.5, 18}, {26.5, -22.5, 16}, // East grove.
		{12.5, -20, 14}, {16, -23.5, 16}, {19.5, -19, 12}, // Behind the bow.
		{-45, 16, 14}, {-41.5, 19.5, 12}, // Front-west corner.
	}
	for i, k := range kelp {
		s.kelp(g, H, P(k[0], k[1]), k[2], 100+i)
	}
	grass := []Vec2{P(-16, 22), P(-19, 25.5), P(-22, 34), P(33, 10), P(35.5, 14), P(-28, -11), P(-24, -15),
		P(-6, -22), P(4, -20), P(24, -16), P(-41, 8), P(-12, 9)}
	for i, c := range grass {
		for j := 0; j < 6; j++ {
			a := 2 * math.Pi * Hash01(i, j, 71)
			p := Polar(c, 1.6*Hash01(i, j, 72), a)
			y := H(p.X, p.Y) - 0.3
			g.Use(Grass)
			if j%3 == 0 {
				g.Use(Foliage)
			}
			g.Fill(Capsule{p.At(y), V(p.X+0.8*math.Cos(a), y+2.2+2.2*Hash01(i, j, 73), p.Y+0.8*math.Sin(a)), 0.5})
		}
	}
}

// fish is one schooling fish: a body, a tail fin and a stripe.
func (s CoralReefWreck) fish(g *Grid, at, dir Vec3, k float64, body, stripe Color) {
	f := coralReefWreckHeading(at, dir, 0)
	f.K = k
	g.Use(body).Fill(f.at(Union(
		Ellipsoid{V(0, 0, 0), V(1.6, 0.85, 0.62)},
		Func{Box{V(-2.7, -1.1, -0.5), V(-1.2, 1.1, 0.5)}, func(l Vec3) bool { return math.Abs(l.Y) <= 0.25+0.75*(-1.2-l.X) }},
	)))
	g.Paint(stripe, f.at(Block(V(-0.2, -1, -0.8), V(0.55, 1, 0.8))))
}

// shark is a grey reef shark about thirteen units long.
func (s CoralReefWreck) shark(g *Grid, at, dir Vec3) {
	f := coralReefWreckHeading(at, dir, 0.12)
	f.K = 0.92
	body := coralReefWreckChain(
		[]Vec3{V(6.6, -0.1, 0), V(4.6, 0, 0), V(2, 0.1, 0), V(-1.5, 0.1, 0), V(-4.6, 0.3, 0)},
		[]float64{0.4, 1.25, 1.75, 1.35, 0.6})
	g.Use(Slate).FillVolume(f.at(body))
	g.Fill(f.at(Union(
		Func{Box{V(-0.9, 1, -0.5), V(2.4, 5, 0.5)}, func(l Vec3) bool { return l.Y <= 1.3+(2.4-l.X)*1.1 }},
		coralReefWreckTaper{V(-4.4, 0.3, 0), V(-7.4, 3.6, 0), 0.7, 0.5},
		coralReefWreckTaper{V(-4.4, 0.3, 0), V(-6.4, -1.8, 0), 0.6, 0.45},
		coralReefWreckTaper{V(2.4, -0.7, 1.2), V(0.4, -1.9, 4.4), 0.8, 0.5},
		coralReefWreckTaper{V(2.4, -0.7, -1.2), V(0.4, -1.9, -4.4), 0.8, 0.5},
	)))
	g.Paint(White, f.at(Func{body.Bounds(), func(l Vec3) bool { return l.Y < -0.35 && body.Contains(l) }}))
	g.Paint(Charcoal, Sphere(f.world(V(5, 0.45, 1.05)), 0.5), Sphere(f.world(V(5, 0.45, -1.05)), 0.5))
}

// manta is a manta ray with upswept wings, head fins and a whip tail.
func (s CoralReefWreck) manta(g *Grid, at, dir Vec3) {
	f := coralReefWreckHeading(at, dir, 0.2)
	lift := func(z float64) float64 { return 0.035 * z * z }
	wing := Func{Box{V(-4.5, -1.5, -8.3), V(4.5, 4, 8.3)}, func(l Vec3) bool {
		z := math.Abs(l.Z)
		if l.X > 3.4-0.1*z-0.05*z*z || l.X < -3.2+0.2*z {
			return false
		}
		return math.Abs(l.Y-lift(z)) <= math.Max(0.55, 1.25*(1-z/7))
	}}
	g.Use(Navy).Fill(f.at(wing),
		f.at(Capsule{V(3, 0, 1.3), V(4.9, -0.2, 1.7), 0.6}), f.at(Capsule{V(3, 0, -1.3), V(4.9, -0.2, -1.7), 0.6}),
		f.at(Capsule{V(-3, 0.2, 0), V(-11, 0.8, 0), 0.5}))
	g.Paint(White, f.at(Func{wing.Box, func(l Vec3) bool {
		z := math.Abs(l.Z)
		if !wing.Contains(l) {
			return false
		}
		return l.Y < lift(z)-0.15 || math.Hypot(l.X-1.4, z-2.4) < 1.1
	}}))
}

// turtle is a green sea turtle, flippers swept back in mid-stroke.
func (s CoralReefWreck) turtle(g *Grid, at, dir Vec3) {
	f := coralReefWreckHeading(at, dir, -0.15)
	f.K = 0.9
	shell := Ellipsoid{V(0, 0.3, 0), V(3.4, 1.5, 2.8)}
	g.Use(Grass).Fill(f.at(Union(
		Sphere(V(4.3, 0.3, 0), 1.1), Capsule{V(3, 0.2, 0), V(4.3, 0.3, 0), 0.8},
		coralReefWreckTaper{V(2.2, 0, 2.2), V(0.2, -0.4, 6.6), 0.85, 0.55},
		coralReefWreckTaper{V(2.2, 0, -2.2), V(0.2, -0.4, -6.6), 0.85, 0.55},
		coralReefWreckTaper{V(-2.8, 0, 1.6), V(-4.8, -0.2, 2.8), 0.7, 0.5},
		coralReefWreckTaper{V(-2.8, 0, -1.6), V(-4.8, -0.2, -2.8), 0.7, 0.5},
	)))
	g.Use(Foliage).FillVolume(f.at(shell))
	g.Paint(Sand, f.at(Func{shell.Bounds(), func(l Vec3) bool { return l.Y < -0.3 && shell.Contains(l) }}))
	g.Paint(Grass, f.at(Func{shell.Bounds(), func(l Vec3) bool {
		q := math.Hypot(l.X/3.4, l.Z/2.8)
		return shell.Contains(l) && l.Y >= -0.3 && (q > 0.82 || (q > 0.38 && q < 0.5))
	}}))
	g.Paint(Charcoal, Sphere(f.world(V(4.7, 0.6, 0.9)), 0.5), Sphere(f.world(V(4.7, 0.6, -0.9)), 0.5))
}

// octopus sprawls over the quarterdeck rail, arms trailing down the hull and across the deck.
func (s CoralReefWreck) octopus(g *Grid, f coralReefWreckFrame) {
	arms := [][]Vec3{
		{V(-20, 15.2, 5.8), V(-21.6, 14.4, 6.6), V(-22.8, 12, 7.3), V(-23.4, 9.8, 7.9)},
		{V(-19, 15.2, 6.2), V(-19.5, 14, 6.9), V(-19.9, 11.5, 7.5), V(-20.6, 9.2, 8.2)},
		{V(-18, 15.2, 6.2), V(-17.5, 14, 6.9), V(-17, 11.5, 7.6), V(-16.2, 9.4, 8.2)},
		{V(-17, 15.2, 5.8), V(-15.4, 14.4, 6.6), V(-14.2, 12.6, 7.2), V(-13.6, 10.4, 7.9)},
		{V(-20.3, 15.3, 5), V(-22.5, 15.2, 5.6), V(-24.6, 15.6, 5.4), V(-26.4, 17.8, 5)},
		{V(-16.7, 15.3, 5), V(-14.5, 15.2, 5.8), V(-12.6, 13.6, 6.4), V(-11, 11.8, 7.1)},
		{V(-19.3, 15, 3.6), V(-20, 13.3, 1.8), V(-21.2, 13.1, -0.4)},
		{V(-17.7, 15, 3.6), V(-17, 13.3, 1.6), V(-15.8, 13.1, -0.8)},
	}
	g.Use(Red)
	for _, arm := range arms {
		radii := make([]float64, len(arm))
		for i := range arm {
			radii[i] = 0.9 - 0.38*float64(i)/float64(len(arm)-1)
		}
		g.Fill(f.at(coralReefWreckChain(arm, radii)))
	}
	g.FillVolume(f.at(Ellipsoid{V(-18.5, 16.5, 5), V(2.1, 2.4, 2.1)}))
	for _, x := range [2]float64{-19.5, -17.5} {
		g.Use(White).Fill(Sphere(f.world(V(x, 16.2, 6.8)), 0.7))
		g.Paint(Charcoal, Sphere(f.world(V(x, 16.2, 7.4)), 0.5))
	}
}

// life adds the swimmers: three fish schools, the shark, the manta, the turtle and the
// octopus on the wreck.
func (s CoralReefWreck) life(g *Grid, stern coralReefWreckHull) {
	// Gold fish circling the masthead.
	top := stern.F.world(V(-7, 35, 0))
	for i := 0; i < 11; i++ {
		a := 0.5 + 0.55*float64(i)
		p := V(top.X+8.5*math.Cos(a), top.Y-1.2+0.24*float64(i), top.Z+8.5*math.Sin(a))
		s.fish(g, p, V(-math.Sin(a), 0.03, math.Cos(a)), 0.85, Gold, White)
	}
	// A red bait ball over the western knoll.
	ball := V(-38, 21, 23)
	for i := 0; i < 14; i++ {
		y := 1 - 2*(float64(i)+0.5)/14
		r := math.Sqrt(1 - y*y)
		a := 2.39996 * float64(i)
		s.fish(g, ball.Add(V(5.6*r*math.Cos(a), 4.6*y, 5.6*r*math.Sin(a))), V(-math.Sin(a), 0, math.Cos(a)), 1.05, Red, White)
	}
	// Pale blue fish fleeing the shark in a tall arch above the eastern kelp, turned to
	// face the hero camera.
	right := V(0.78, 0, -0.63)
	for i := 0; i < 5; i++ {
		a := 2.88 - 0.655*float64(i)
		for _, r := range [2]float64{7.5, 10} {
			p := V(40, 29+r*math.Sin(a), -16).Add(right.Scale(r * math.Cos(a)))
			s.fish(g, p, right.Scale(math.Sin(a)).Add(V(0, -math.Cos(a), 0)), 1, Glass, Navy)
		}
	}
	s.shark(g, V(10, 34, -18), V(0.8, -0.04, -0.6))
	s.manta(g, V(-22, 38, 10), V(0.75, 0.05, 0.66))
	s.turtle(g, V(20, 10, 38), V(-0.85, 0.08, -0.5))
	s.octopus(g, stern.F)
}

// explorers adds the yellow submersible with its bubble column and the helmeted diver at
// the chest, joined by an air hose.
func (s CoralReefWreck) explorers(g *Grid, H coralReefWreckField) {
	sub := coralReefWreckHeading(V(-8, 7, 36), V(0.78, -0.05, -0.63), 0)
	hull := Ellipsoid{V(0, 0, 0), V(4.2, 2.3, 2.3)}
	g.Use(Gold).FillVolume(sub.at(hull))
	g.Fill(sub.at(Cylinder{P(-0.6, 0), 1.3, 0, 1.8, 3.4}))
	g.Use(Red).Fill(sub.at(Cylinder{P(-0.6, 0), 1.5, 0, 3.4, 4.1}))
	g.Paint(Red, sub.at(Intersect(hull, Block(V(-2.5, -3, -3), V(-1.6, 3, 3)))))
	g.Use(Glass).FillVolume(sub.at(Sphere(V(2.9, 0.3, 0), 1.9)))
	g.Use(Charcoal).Fill(sub.at(Union(
		Capsule{V(-2.6, -2.9, -1.5), V(2.6, -2.9, -1.5), 0.55}, Capsule{V(-2.6, -2.9, 1.5), V(2.6, -2.9, 1.5), 0.55},
		Capsule{V(-1.6, -1.6, -1.3), V(-1.6, -2.9, -1.5), 0.5}, Capsule{V(-1.6, -1.6, 1.3), V(-1.6, -2.9, 1.5), 0.5},
		Capsule{V(1.6, -1.6, -1.3), V(1.6, -2.9, -1.5), 0.5}, Capsule{V(1.6, -1.6, 1.3), V(1.6, -2.9, 1.5), 0.5},
		Func{Box{V(-5.6, -2.1, -2.1), V(-4.4, 2.1, 2.1)}, func(l Vec3) bool {
			r := math.Hypot(l.Y, l.Z)
			return r <= 2.1 && r >= 1.1
		}},
		Capsule{V(-3.9, 0, 0), V(-5.4, 0, 0), 0.6},
		Capsule{V(3.4, -1.6, 1.2), V(5.8, -2.6, 1.6), 0.5},
	)))
	g.Use(White).Fill(sub.at(Sphere(V(3.4, -1.5, -1.3), 0.7)))
	// Bubble column rising from the conning tower and trailing astern.
	vent := sub.world(V(-0.6, 4.6, 0))
	for i := 0; i < 6; i++ {
		g.Use(White)
		if i%2 == 1 {
			g.Use(Glass)
		}
		rise := float64(i)
		g.Fill(Sphere(vent.Add(sub.U.Scale(-0.3-0.25*rise)).Add(V(0, 2.2+2.4*rise, 0.5*math.Cos(1.9*rise))), 0.6+0.09*rise))
	}

	// The diver, reaching for the chest.
	feet := P(-9.5, 16.5)
	d := coralReefWreckHeading(feet.At(H(feet.X, feet.Y)+0.2), V(1, 0, -0.25), 0)
	d.K = 0.95
	g.Use(Charcoal).Fill(d.at(Block(V(-0.7, 0, -1.5), V(1.2, 1, -0.3))), d.at(Block(V(-0.7, 0, 0.3), V(1.2, 1, 1.5))))
	g.Use(White).Fill(d.at(Union(
		Capsule{V(0, 1, -0.9), V(0, 2.7, -0.7), 0.65}, Capsule{V(0, 1, 0.9), V(0, 2.7, 0.7), 0.65},
		Capsule{V(0, 3, 0), V(0, 4.4, 0), 1.3},
		Capsule{V(0, 4.5, -1.6), V(0.6, 2.9, -2.2), 0.62}, Capsule{V(0, 4.5, 1.6), V(2.1, 4.1, 2.1), 0.62},
	)))
	g.Use(Charcoal).Fill(d.at(Block(V(0.9, 3.2, -0.7), V(1.6, 4.3, 0.7))))
	g.Use(Gold).Fill(d.at(Cylinder{P(0, 0), 1.5, 0, 5, 5.5}), d.at(Sphere(V(0, 6.4, 0), 1.45)))
	g.Paint(Glass, Sphere(d.world(V(1.3, 6.5, 0)), 0.6))
	helmet := d.world(V(0, 7.6, 0))
	for i := 0; i < 3; i++ {
		g.Use(White).Fill(Sphere(helmet.Add(V(0.6*math.Sin(2.1*float64(i)), 1.6+2.1*float64(i), 0.5*math.Cos(1.7*float64(i)))), 0.55+0.06*float64(i)))
	}
	// Air hose sagging across the sand from the submersible's keel to the helmet.
	a, b := sub.world(V(-2.4, -2.2, 0)), d.world(V(-1.3, 6.4, 0))
	prev := a
	for i := 1; i <= 14; i++ {
		t := float64(i) / 14
		p := a.Add(b.Sub(a).Scale(t))
		p.Y = math.Max(H(p.X, p.Z)+0.6, p.Y-7*math.Sin(math.Pi*t))
		g.Use(Charcoal).Fill(Capsule{prev, p, 0.5})
		prev = p
	}
}
