package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// MarinaBaySands is the Singapore resort and its waterfront, laid out as on the map
// with north towards +Z and the bay (west) towards +X, so the hero view is the classic
// one from the Helix bridge and the side view is the postcard from across the bay.
// Three 55-storey hotel towers stand in a north-south row; each is two slabs, a leaning
// garden-side slab with planted balcony grids and a curved bay-side slab with a banded
// glass curtain wall, splayed apart over a glazed atrium at the foot and merged higher
// up. The SkyPark lies across their roofs: a white ship-shaped hull that cantilevers far
// past the north tower, with an infinity pool along the bay edge, palms and gardens,
// and a timber observation deck at the prow. West of the hotel the low Shoppes and
// convention podium has four vaulted roofs with glazed arcs over a glass arcade, then
// the tree-lined promenade, the stepped event plaza and two faceted Crystal Pavilions
// on the water. The lotus-shaped ArtScience Museum (ten upturned white fingers over a
// lily pond) sits on the north-west promontory, where the double-helix footbridge
// crosses the channel to the far bank. East of the towers are Dragonfly Lake, the
// Supertree Grove with its yellow aerial walkway, and the two conservatory domes.
// Bumboats and yachts cross the built bay plate.
type MarinaBaySands struct{}

func init() { Register(MarinaBaySands{}) }

func (MarinaBaySands) Name() string  { return "marina_bay_sands" }
func (MarinaBaySands) Title() string { return "Marina Bay Sands" }

const (
	marinaBaySandsG      = 1.6  // Land level.
	marinaBaySandsRoof   = 74.0 // Tower roofs, where the SkyPark's keel rests.
	marinaBaySandsDeck   = 78.5 // SkyPark deck.
	marinaBaySandsMerge  = 0.44 // Height fraction where a tower's two slabs meet.
	marinaBaySandsPodX0  = 7.0  // Podium back wall, towards the hotel.
	marinaBaySandsPodX1  = 27.0 // Podium gables, towards the bay.
	marinaBaySandsPodZ0  = -44.0
	marinaBaySandsPodZ1  = 26.0
	marinaBaySandsArcade = 31.0 // Front of the Shoppes arcade.
	marinaBaySandsEave   = 6.2  // Podium eaves and link roofs.
)

// Key positions.
var (
	marinaBaySandsPlate  = P(6, 13)     // Center of the bay plate.
	marinaBaySandsMuseum = P(38, 41)    // ArtScience Museum.
	marinaBaySandsPlaza  = P(35.5, -10) // Event plaza on the promenade.
	marinaBaySandsFlower = P(-34, 43)   // Flower Dome.
	marinaBaySandsCloud  = P(-18, 47)   // Cloud Forest dome.
	// Podium halls: south edge, north edge and crown height of each vault.
	marinaBaySandsHalls = [4][3]float64{{-44, -25, 13.2}, {-22, -5, 12.2}, {-2, 12, 11.4}, {15, 26, 10.4}}
	// Supertrees: plan position and height.
	marinaBaySandsGrove = [6][3]float64{{-39, 13, 25}, {-43, 1, 21}, {-31, 22, 19}, {-43, 25, 17}, {-28, 32, 15}, {-34, 2, 13.5}}
)

func marinaBaySandsClamp(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

func marinaBaySandsSmooth(t float64) float64 {
	t = marinaBaySandsClamp(t, 0, 1)
	return t * t * (3 - 2*t)
}

func marinaBaySandsMod(v, n int) int { return ((v % n) + n) % n }

func marinaBaySandsFmod(v, n float64) float64 { return v - n*math.Floor(v/n) }

// marinaBaySandsNoise is smooth value noise in [-1, 1].
func marinaBaySandsNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return (a+(b-a)*sx+(c+(d-c)*sx-a-(b-a)*sx)*sz)*2 - 1
}

// marinaBaySandsSegment is the plan distance from p to the segment ab.
func marinaBaySandsSegment(p, a, b Vec2) float64 {
	ab := b.Sub(a)
	t := marinaBaySandsClamp(p.Sub(a).Dot(ab)/math.Max(ab.Dot(ab), 1e-9), 0, 1)
	return p.Sub(a.Add(ab.Scale(t))).Length()
}

// marinaBaySandsInPlate is the rounded bay plate that carries the whole site.
func marinaBaySandsInPlate(x, z float64) bool {
	dx, dz := math.Abs(x-marinaBaySandsPlate.X)/56, math.Abs(z-marinaBaySandsPlate.Y)/65
	return math.Pow(dx, 2.5)+math.Pow(dz, 2.5) <= 1
}

// marinaBaySandsShore is the bay edge of the promenade, stepping back north of the podium.
func marinaBaySandsShore(z float64) float64 {
	d := (z + 9) / 36
	return 35 + 1.6*d*d - 6.5*marinaBaySandsSmooth((z-25)/7)
}

// marinaBaySandsLand reports whether a plan point is land: the resort, the museum
// promontory, the event plaza and the far bank beyond the channel.
func marinaBaySandsLand(x, z float64) bool {
	if !marinaBaySandsInPlate(x, z) {
		return false
	}
	if z >= 68 {
		return true
	}
	if z > 56+0.8*math.Sin(x*0.12) {
		return false
	}
	if P(x, z).Sub(marinaBaySandsMuseum).Length() <= 14 || P(x, z).Sub(marinaBaySandsPlaza).Length() <= 7 {
		return true
	}
	return x <= marinaBaySandsShore(z)
}

// marinaBaySandsLake is Dragonfly Lake, the long pond between the hotel and the gardens.
func marinaBaySandsLake(x, z float64) bool {
	c := -25.6 + 1.1*math.Sin(z*0.13)
	end := math.Max(math.Max(-41-z, z-21), 0)
	return math.Hypot(x-c, end) < 2.0
}

// marinaBaySandsParkAxis is the SkyPark's gently bowed center line.
func marinaBaySandsParkAxis(z float64) float64 {
	d := (z - 4) / 46
	return -12 - 2.4*d*d
}

// marinaBaySandsParkHalf is the SkyPark's half width at z, and how far (0 to 1) z lies
// into the rounded stern or the pointed prow.
func marinaBaySandsParkHalf(z float64) (half, end float64) {
	const stern, aft, fore, prow, beam = -44.0, -30.0, 20.0, 51.0, 8.5
	switch {
	case z < stern || z > prow:
		return 0, 1
	case z < aft:
		s := (aft - z) / (aft - stern)
		return beam * math.Pow(1-math.Pow(s, 2.4), 0.6), s
	case z > fore:
		s := (z - fore) / (prow - fore)
		return beam * math.Pow(1-math.Pow(s, 1.9), 0.75), s
	}
	return beam, 0
}

// marinaBaySandsSite carries the grid and the shared layout through the build.
type marinaBaySandsSite struct {
	g      *Grid
	towers [3]marinaBaySandsTower
}

func (MarinaBaySands) Build(g *Grid) Order {
	s := &marinaBaySandsSite{g: g}
	splay := [3]float64{9.5, 8.5, 9}
	for i, z := range [3][2]float64{{-36, -20}, {-14, 2}, {8, 24}} {
		s.towers[i] = marinaBaySandsTower{X: marinaBaySandsParkAxis((z[0] + z[1]) / 2), Z0: z[0], Z1: z[1], Splay: splay[i]}
	}
	s.ground()
	for _, t := range s.towers {
		s.tower(t)
	}
	s.skypark()
	s.podium()
	s.promenade()
	s.crystal(P(42, -29), 0.25, 1)
	s.crystal(P(43.5, 9), -0.3, 2)
	s.museum()
	s.helix()
	s.gardens()
	s.dome(marinaBaySandsFlower, 9, 11, 9.5, false)
	s.dome(marinaBaySandsCloud, 7, 7, 13.5, true)
	s.boats()
	return Along(P(0, 1))
}

// each visits every lattice cell whose true center lies inside the box.
func (s *marinaBaySandsSite) each(b Box, f func(c Cell, p Vec3)) {
	span := s.g.Span()
	lo := func(v float64) int32 { return int32(math.Floor(v / span * 2)) }
	hi := func(v float64) int32 { return int32(math.Ceil(v / span * 2)) }
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
				if p := s.g.World(c); b.Contains(p) {
					f(c, p)
				}
			}
		}
	}
}

// put occupies, in one color, the cells whose true centers satisfy in. Unlike Fill it
// does not snap to the half-unit grid, so thin skins keep their authored thickness.
func (s *marinaBaySandsSite) put(col Color, b Box, in func(p Vec3) bool) {
	g := s.g
	brush := g.Brush()
	g.Use(col)
	s.each(b, func(c Cell, p Vec3) {
		if in(p) {
			g.Remove(c)
			g.Add(c)
		}
	})
	g.Use(brush)
}

// tint recolors occupied cells in the box with the color pick returns for them.
func (s *marinaBaySandsSite) tint(b Box, pick func(c Cell, p Vec3) (Color, bool)) {
	g := s.g
	brush := g.Brush()
	s.each(b, func(c Cell, p Vec3) {
		if !g.Has(c) {
			return
		}
		if col, ok := pick(c, p); ok {
			g.Remove(c)
			g.Use(col).Add(c)
		}
	})
	g.Use(brush)
}

// shell occupies the cells of a predicate within t of its outside. With floor false the
// underside is left open, for skins that stand on the ground.
func (s *marinaBaySandsSite) shell(col Color, b Box, t float64, floor bool, in func(Vec3) bool) {
	s.put(col, b, func(p Vec3) bool {
		if !in(p) {
			return false
		}
		return !(in(p.Add(V(t, 0, 0))) && in(p.Add(V(-t, 0, 0))) && in(p.Add(V(0, 0, t))) &&
			in(p.Add(V(0, 0, -t))) && in(p.Add(V(0, t, 0))) && (!floor || in(p.Add(V(0, -t, 0)))))
	})
}

// covered reports ground hidden inside a closed building, where no paving is laid.
func (s *marinaBaySandsSite) covered(x, z float64) bool {
	if x > marinaBaySandsPodX0+1.2 && x < marinaBaySandsArcade-1.2 && z > marinaBaySandsPodZ0+1.2 && z < marinaBaySandsPodZ1-1.2 {
		return true
	}
	for _, t := range s.towers {
		if z > t.Z0+2.6 && z < t.Z1-2.6 && x > t.X-7.8 && x < t.X+4+t.Splay {
			return true
		}
	}
	for _, d := range [2][4]float64{{marinaBaySandsFlower.X, marinaBaySandsFlower.Y, 9, 11}, {marinaBaySandsCloud.X, marinaBaySandsCloud.Y, 7, 7}} {
		if math.Hypot((x-d[0])/(d[2]-1.3), (z-d[1])/(d[3]-1.3)) < 1 {
			return true
		}
	}
	return false
}

// ground lays the bay plate: one layer of water, the quay walls and the paved land.
func (s *marinaBaySandsSite) ground() {
	const top = marinaBaySandsG
	box := func(y0, y1 float64) Box { return Box{V(-51, y0, -53), V(63, y1, 79)} }
	land := marinaBaySandsLand
	edge := func(x, z, r float64) bool {
		for k := 0; k < 8; k++ {
			if q := Polar(P(x, z), r, float64(k)*math.Pi/4); !land(q.X, q.Y) {
				return true
			}
		}
		return false
	}
	s.put(Water, box(-0.2, 0.2), func(p Vec3) bool {
		return marinaBaySandsInPlate(p.X, p.Z) && !land(p.X, p.Z)
	})
	s.tint(box(-0.2, 0.2), func(c Cell, p Vec3) (Color, bool) {
		return Glass, Hash01(int(c.X), int(c.Z), 2) < 0.03
	})
	s.put(DarkStone, box(-0.2, top-0.35), func(p Vec3) bool { return land(p.X, p.Z) && edge(p.X, p.Z, 1.0) })
	s.put(Grass, box(top-0.3, top), func(p Vec3) bool { return land(p.X, p.Z) && !s.covered(p.X, p.Z) })
	s.tint(box(top-0.3, top), func(c Cell, p Vec3) (Color, bool) {
		if edge(p.X, p.Z, 0.9) {
			return Stone, true
		}
		return s.paving(c, p), true
	})
}

// paving picks the surface color of a land cell.
func (s *marinaBaySandsSite) paving(c Cell, p Vec3) Color {
	x, z := p.X, p.Z
	hx, hz := int(c.X), int(c.Z)
	lawn := func(seed int) Color {
		if marinaBaySandsNoise(x*0.17, z*0.17, seed) > 0.2 {
			return Foliage
		}
		return Grass
	}
	if z >= 68 { // The far bank: a waterfront path and park.
		if z < 70 {
			return Sand
		}
		return lawn(3)
	}
	if d := p.Plan().Sub(marinaBaySandsMuseum); d.Length() <= 14.3 {
		switch r := d.Length(); {
		case r < 3.6:
			return Stone
		case r < 10.4: // Lily pond.
			if marinaBaySandsNoise(x*0.6, z*0.6, 7) > 0.28 {
				if Hash01(hx, hz, 8) < 0.2 {
					return White
				}
				return Foliage
			}
			return Water
		case r < 11.3:
			return White
		}
		if marinaBaySandsFmod(d.Angle(), math.Pi/7) < 0.12 {
			return Stone
		}
		return Sand
	}
	if r := p.Plan().Sub(marinaBaySandsPlaza).Length(); r <= 7.2 && x > 32 { // Event plaza rings.
		switch {
		case r < 1.6:
			return Terracotta
		case marinaBaySandsMod(int(math.Floor(r/1.3)), 2) == 0:
			return Stone
		}
		return Sand
	}
	switch {
	case x > marinaBaySandsArcade: // Promenade.
		if marinaBaySandsFmod(z, 6) < 0.8 {
			return Stone
		}
		return Sand
	case x > 6.6: // Around the podium.
		if z > marinaBaySandsPodZ1+2 {
			if z > 40 {
				return lawn(4)
			}
			return Sand
		}
		return Stone
	case x > 2.4: // Bayfront Avenue.
		if math.Abs(x-4.5) < 0.42 && marinaBaySandsFmod(z, 4.5) < 2 {
			return White
		}
		return DarkStone
	case x > -21.6: // Hotel forecourts.
		if z > 40 {
			return lawn(5)
		}
		if z > 24 || z < -36 {
			if marinaBaySandsMod(int(math.Floor(x/3.2))+int(math.Floor(z/3.2)), 2) == 0 {
				return Sand
			}
		}
		return Stone
	}
	// Gardens by the Bay.
	if marinaBaySandsLake(x, z) {
		return Water
	}
	for _, t := range marinaBaySandsGrove {
		if d := p.Plan().Sub(P(t[0], t[1])).Length(); d < 4.2 {
			return Stone
		}
	}
	for i := 0; i+1 < len(marinaBaySandsPaths); i++ {
		if marinaBaySandsSegment(p.Plan(), marinaBaySandsPaths[i], marinaBaySandsPaths[i+1]) < 1.0 {
			return Sand
		}
	}
	return lawn(6)
}

// marinaBaySandsPaths is the garden path from the lake bridge around the grove.
var marinaBaySandsPaths = []Vec2{P(-21.6, 13), P(-39, 13), P(-43, 1), P(-34, 2), P(-39, 13), P(-43, 25),
	P(-31, 22), P(-28, 32), P(-22, 36), P(-10, 40), P(4.5, 42)}

// marinaBaySandsTower is one hotel tower: a leaning garden slab and a curved bay slab.
type marinaBaySandsTower struct {
	X      float64 // Axis of the merged slabs at the roof.
	Z0, Z1 float64 // Extent along the row.
	Splay  float64 // How far the bay slab kicks out at its foot.
}

func (t marinaBaySandsTower) frac(y float64) float64 {
	return marinaBaySandsClamp((y-marinaBaySandsG)/(marinaBaySandsRoof-marinaBaySandsG), 0, 1)
}

// faces returns the outer faces of the garden (east) and bay (west) slabs at height y.
func (t marinaBaySandsTower) faces(y float64) (east, west float64) {
	h := t.frac(y)
	east = t.X - 9 + 3.5*h
	west = t.X + 5.5
	if h < 0.62 {
		west += t.Splay * math.Pow(1-h/0.62, 2.2)
	}
	return east, west
}

// atrium returns the pointed void between the two slabs at height y.
func (t marinaBaySandsTower) atrium(y float64) (lo, hi float64, ok bool) {
	h := t.frac(y)
	if h >= marinaBaySandsMerge {
		return 0, 0, false
	}
	east, west := t.faces(y)
	em, wm := t.faces(marinaBaySandsG + marinaBaySandsMerge*(marinaBaySandsRoof-marinaBaySandsG))
	k := h / marinaBaySandsMerge
	lo = east + 4
	hi = west - 4.5 - (wm-em-8.5)*k*k
	return lo, hi, hi > lo+0.5
}

// tower builds one hollow tower: two long skins, end walls that show the slabs'
// profile, and a recessed glass wall closing the atrium at each end.
func (s *marinaBaySandsSite) tower(t marinaBaySandsTower) {
	const wall = 0.9
	b := Box{V(t.X-9.5, marinaBaySandsG, t.Z0), V(t.X+6+t.Splay, marinaBaySandsRoof+2.6, t.Z1)}
	s.put(White, b, func(p Vec3) bool {
		east, west := t.faces(p.Y)
		if p.X < east || p.X > west {
			return false
		}
		if p.X <= east+wall || p.X >= west-wall-0.2 {
			return true
		}
		end := math.Min(p.Z-t.Z0, t.Z1-p.Z)
		if end > 2.3 {
			return false
		}
		lo, hi, open := t.atrium(p.Y)
		switch {
		case !open:
			return end <= wall
		case p.X > lo && p.X < hi:
			return end >= 1.4 // Atrium glazing, set back in the opening.
		case end <= wall:
			return true
		}
		return p.X >= lo-wall && p.X <= hi+wall // Reveals of the opening.
	})
	seed := int(t.Z0)
	s.tint(b, func(c Cell, p Vec3) (Color, bool) {
		east, west := t.faces(p.Y)
		end := math.Min(p.Z-t.Z0, t.Z1-p.Z)
		floor := marinaBaySandsMod(int(c.Y), 8)
		storey := (int(c.Y) - floor) / 8
		line := floor < 2
		lo, hi, open := t.atrium(p.Y)
		switch {
		case open && p.X > lo && p.X < hi: // Cable-net glazing.
			if line || marinaBaySandsFmod(p.X-t.X, 3.2) < 0.8 {
				return Slate, true
			}
			return Glass, true
		case end < 1.0 || (p.X > east+wall && p.X < west-wall-0.2): // End walls.
			if t.frac(p.Y) > marinaBaySandsMerge+0.04 && math.Abs(p.X-(east+west)/2) < 1.0 && end < 1.0 {
				if line {
					return Slate, true
				}
				return Glass, true
			}
			if line {
				return Stone, true
			}
		case p.X >= west-wall-0.2: // Bay side: glass curtain wall banded by floor lines.
			if line {
				return Slate, true
			}
			return Glass, true
		default: // Garden side: balcony grid with planters.
			zz := p.Z - t.Z0 - 1
			bay := int(math.Floor(zz / 2.8))
			switch {
			case line || marinaBaySandsFmod(zz, 2.8) < 0.75:
				return White, true
			case floor < 4 && Hash01(bay, storey, seed) < 0.55:
				return Foliage, true
			}
			return Slate, true
		}
		return 0, false
	})
}

// skypark builds the ship-shaped deck across the three roofs.
func (s *marinaBaySandsSite) skypark() {
	g := s.g
	const deck = marinaBaySandsDeck
	local := func(p Vec3) (u, w, end float64) {
		w, end = marinaBaySandsParkHalf(p.Z)
		return p.X - marinaBaySandsParkAxis(p.Z), w, end
	}
	b := Box{V(-27, deck-5.2, -45), V(-2, deck, 52)}
	in := func(p Vec3) bool {
		if p.Y > deck {
			return false
		}
		u, w, end := local(p)
		if w <= 0 || math.Abs(u) > w {
			return false
		}
		depth := 1.8 + 3.2*(1-0.75*end*end)*math.Pow(1-(u/w)*(u/w), 0.55)
		return p.Y >= deck-depth
	}
	s.shell(White, b, 0.9, true, in)
	const pool0, pool1 = -28.0, 14.0
	s.tint(b, func(c Cell, p Vec3) (Color, bool) {
		u, w, _ := local(p)
		side := w - u // Distance in from the bay edge.
		pool := p.Z > pool0 && p.Z < pool1 && side < 4.2
		if p.Y < deck-0.9 || (math.Abs(u) > w-0.95 && !pool) { // Hull: white sides, pale belly.
			return Stone, p.Y < deck-3
		}
		switch {
		case pool:
			return Water, true
		case p.Z > pool0-1 && p.Z < pool1+1 && side < 6.4:
			if side > 4.7 && side < 5.8 && marinaBaySandsFmod(p.Z, 2.4) < 0.8 {
				return White, true // Sun loungers.
			}
			return Sand, true
		case p.Z > 24.5: // Observation deck.
			return Wood, true
		case p.Z < -31 || (u < -w+5.6 && marinaBaySandsFmod(p.Z+3, 11) < 8):
			if marinaBaySandsNoise(p.X*0.5, p.Z*0.5, 21) > 0.2 {
				return Foliage, true
			}
			return Grass, true
		case u < -w+5.6:
			return Wood, true // Timber terraces between the gardens.
		case marinaBaySandsFmod(p.Z, 5.5) < 0.8:
			return Stone, true // Paving joints.
		}
		return Sand, true
	})
	// Glass balustrade, left off the infinity edge of the pool.
	s.put(Glass, Box{b.Min, V(b.Max.X, deck+0.75, b.Max.Z)}, func(p Vec3) bool {
		u, w, _ := local(p)
		if p.Y <= deck || w <= 0 || math.Abs(u) > w || math.Abs(u) < w-0.95 {
			return false
		}
		return !(u > 0 && p.Z > pool0 && p.Z < pool1)
	})
	at := func(u, z float64) Vec3 { return V(marinaBaySandsParkAxis(z)+u, deck, z) }
	// Palms along the garden edge and across the stern.
	for z := -40.0; z < 23; z += 5.5 {
		if z < -30 || marinaBaySandsFmod(z+3, 11) < 8 {
			s.palm(at(-5.4, z), 2.6+0.8*Hash01(int(z), 31))
		}
	}
	s.palm(at(1.5, -38.5), 3.2)
	s.palm(at(4, -34), 2.8)
	s.palm(at(-1.5, -33), 3)
	// Restaurant pavilions and lift lobbies over each tower.
	for i, t := range s.towers {
		c := at(-0.8, (t.Z0+t.Z1)/2+2)
		w, d := 5.0, 6.0
		if i == 1 {
			d = 8
		}
		s.shell(White, Box{c.Add(V(-w/2, 0, -d/2)), c.Add(V(w/2, 2.6, d/2))}, 0.85, false, func(Vec3) bool { return true })
		s.tint(Box{c.Add(V(-w/2, 0.5, -d/2)), c.Add(V(w/2, 1.7, d/2))}, func(Cell, Vec3) (Color, bool) { return Slate, true })
	}
	// Canopy at the entrance to the observation deck, and visitors along the rail.
	g.Use(White).Fill(Block(at(-5, 25).Add(V(0, 3, 0)), at(5, 28.5).Add(V(0, 3.6, 0))))
	for _, u := range []float64{-4.5, 4.5} {
		g.Use(Stone).Fill(Capsule{at(u, 26.7), at(u, 26.7).Add(V(0, 3, 0)), 0.5})
	}
	for i, q := range [][2]float64{{-3.5, 33}, {2.8, 35}, {-1.5, 39}, {1.2, 43}, {0, 47}, {3.6, 31}, {-2, 30}} {
		g.Use([]Color{Red, Navy, Gold, White}[i%4]).AddAt(at(q[0], q[1]).Add(V(0, 0.45, 0)))
	}
}

// palm is a slim trunk under a flat crown of fronds.
func (s *marinaBaySandsSite) palm(at Vec3, h float64) {
	g := s.g
	g.Use(Wood).Fill(Capsule{at, at.Add(V(0, h, 0)), 0.5})
	g.Use(Foliage).FillVolume(Ellipsoid{at.Add(V(0, h+0.4, 0)), V(1.9, 0.85, 1.9)})
}

// tree is a round-crowned park tree.
func (s *marinaBaySandsSite) tree(at Vec2, h float64, crown Color) {
	g := s.g
	g.Use(Wood).Fill(Cylinder{Center: at, Radius: 0.55, Y0: marinaBaySandsG, Y1: marinaBaySandsG + h*0.5})
	g.Use(crown).FillVolume(Ellipsoid{at.At(marinaBaySandsG + h*0.7), V(h*0.36, h*0.32, h*0.36)})
}

// marinaBaySandsVault is the podium roof height: flat links at the eaves and a
// parabolic vault over each hall, falling gently towards the bay.
func marinaBaySandsVault(x, z float64) (y float64, hall int) {
	y, hall = marinaBaySandsEave, -1
	for i, h := range marinaBaySandsHalls {
		if z >= h[0] && z <= h[1] {
			v := (z - (h[0]+h[1])/2) / ((h[1] - h[0]) / 2)
			crown := h[2] - 2.4*marinaBaySandsClamp((x-marinaBaySandsPodX0)/20, 0, 1.1)
			return marinaBaySandsEave + (crown-marinaBaySandsEave)*(1-v*v), i
		}
	}
	return y, hall
}

// podium builds the Shoppes and convention halls: one hollow mass with four vaults,
// glazed lunettes facing the bay, and the long glass-roofed arcade in front.
func (s *marinaBaySandsSite) podium() {
	const x0, x1, z0, z1 = marinaBaySandsPodX0, marinaBaySandsPodX1, marinaBaySandsPodZ0, marinaBaySandsPodZ1
	const base, front = marinaBaySandsG, marinaBaySandsArcade
	arcade := func(x float64) float64 { // Barrel skylight over the arcade.
		d := (x - (x1+front)/2) / ((front - x1) / 2)
		return base + 4.3 + 1.1*(1-d*d)
	}
	b := Box{V(x0, base, z0), V(front+1, 14, z1)}
	hall := func(p Vec3) bool {
		if p.X > x1 || p.Y < base {
			return false
		}
		y, _ := marinaBaySandsVault(p.X, p.Z)
		return p.Y <= y
	}
	s.shell(White, b, 0.9, false, func(p Vec3) bool {
		// The gable is hidden behind the arcade below its roof.
		return hall(p) && !(p.X > x1-1 && p.Y < base+4)
	})
	// Roof edge oversailing the lunettes.
	s.put(White, b, func(p Vec3) bool {
		y, i := marinaBaySandsVault(x1, p.Z)
		return i >= 0 && p.X > x1 && p.X <= x1+1.5 && p.Y <= y && p.Y > y-0.9 && y > marinaBaySandsEave+0.6
	})
	s.shell(White, Box{V(x1, base, z0), V(front, base+6, z1)}, 0.9, false, func(p Vec3) bool {
		// Open at the back, where the arcade meets the halls.
		return p.Y <= arcade(p.X) || p.X < x1+1
	})
	s.put(White, Box{V(x1, base, z0), V(x1+1, base+6, z1)}, func(p Vec3) bool { return p.Y <= arcade(p.X) && p.Y > arcade(p.X)-0.9 })
	s.tint(b, func(c Cell, p Vec3) (Color, bool) {
		y, i := marinaBaySandsVault(p.X, p.Z)
		switch {
		case p.X > x1: // Arcade.
			top := arcade(p.X)
			switch {
			case p.Y > top-0.95 && p.X < front-0.9:
				if marinaBaySandsFmod(p.Z, 2.4) < 0.8 || p.X < x1+0.9 {
					return White, true
				}
				return Glass, true
			case p.X >= front-0.9 && p.Y < base+3.4:
				if marinaBaySandsFmod(p.Z+1, 5) < 0.8 {
					return White, true
				}
				if p.Y < base+2 && marinaBaySandsFmod(p.Z+1, 15) > 6.5 && marinaBaySandsFmod(p.Z+1, 15) < 8.5 {
					return Charcoal, true
				}
				return Glass, true
			case p.Z < z0+0.9 || p.Z > z1-0.9:
				return Stone, true
			}
			return 0, false
		case p.Y > y-1.0: // Roofs.
			if i < 0 {
				if marinaBaySandsFmod(p.X, 3.2) < 0.8 {
					return White, true
				}
				return Glass, true // Glazed links between the halls.
			}
			h := marinaBaySandsHalls[i]
			if math.Abs(p.Z-(h[0]+h[1])/2) < 0.9 && p.X > x0+2 && p.X < x1-1.5 {
				return Glass, true // Ridge skylight.
			}
			if marinaBaySandsFmod(p.X-x0, 4) < 0.8 || p.X > x1-0.9 {
				return White, true // Standing seams, drawn as arcs, and the eaves.
			}
			return Stone, true
		case p.X > x1-1: // Lunettes facing the bay.
			if i < 0 || marinaBaySandsFmod(p.Z-marinaBaySandsHalls[i][0], 3.2) < 0.8 {
				return White, true
			}
			return Glass, true
		case p.Y > base+3.2 && p.Y < base+4.4: // Clerestory band on the other walls.
			return Glass, true
		}
		return Stone, true
	})
}

// promenade plants the waterfront walk and dresses the event plaza and the avenue.
func (s *marinaBaySandsSite) promenade() {
	g := s.g
	const base = marinaBaySandsG
	for z := -42.0; z <= 24; z += 6 {
		if math.Abs(z-marinaBaySandsPlaza.Y) < 8 {
			continue
		}
		s.tree(P(marinaBaySandsShore(z)-1.8, z+3), 4.4, Foliage)
	}
	// The plaza's lower step into the bay, kiosks under red parasols and a crowd.
	s.put(Stone, Box{V(34, 0.5, -20), V(46, 1.0, 0)}, func(p Vec3) bool {
		r := p.Plan().Sub(marinaBaySandsPlaza).Length()
		return r > 7 && r <= 8.8 && p.X > marinaBaySandsShore(p.Z)
	})
	for i, z := range []float64{-15.5, -4.5} {
		c := P(34, z)
		g.Use(White).Fill(Capsule{c.At(base), c.At(base + 2.4), 0.5})
		g.Use([]Color{Red, Gold}[i]).FillVolume(Cone(c, 2.1, base+2.2, 1.5))
	}
	for i, q := range [][2]float64{{36, -12}, {38, -9}, {37, -6.5}, {39.5, -11.5}, {33, -9}, {35.5, -14}, {40, -8}} {
		g.Use([]Color{Red, Navy, White, Gold, Charcoal}[i%5]).AddAt(V(q[0], base+0.45, q[1]))
	}
	for i, z := range []float64{-39, -30, -21, -2, 8, 19, 30, 46} {
		g.Use([]Color{Navy, Red, White, Charcoal}[i%4]).AddAt(V(33+1.2*Hash01(i, 41), base+0.45, z))
	}
	// Traffic on Bayfront Avenue.
	for i, z := range []float64{-46, -33, -19, -7, 6, 17, 33, 47, 60} {
		x := 3.5 + 2*float64(i%2)
		col := []Color{Red, White, Gold, Navy, Glass}[i%5]
		g.Use(col).Fill(Block(V(x-0.5, base, z-1), V(x+0.5, base+0.8, z+1)))
	}
}

// crystal is one Crystal Pavilion: a faceted glass gem on a plinth in the bay, joined
// to the promenade by a short pier.
func (s *marinaBaySandsSite) crystal(c Vec2, rot float64, seed int) {
	const base, skin = 1.4, 0.95
	var normals []Vec3
	var dists []float64
	add := func(n Vec3, d float64) {
		normals = append(normals, n.Scale(1/n.Length()))
		dists = append(dists, d)
	}
	for j := 0; j < 6; j++ {
		a := float64(j)*math.Pi/3 + 0.5*(Hash01(seed, j, 1)-0.5)
		tilt := 0.1 + 0.42*Hash01(seed, j, 2)
		add(V(math.Cos(a)*math.Cos(tilt), math.Sin(tilt), math.Sin(a)*math.Cos(tilt)), 3.7+1.1*Hash01(seed, j, 3))
	}
	add(V(0.32, 1, 0.14), 6.2)
	add(V(-0.38, 1, -0.2), 6.6)
	// measure returns how deep p lies inside the gem and how many facets it touches.
	measure := func(p Vec3) (depth float64, near int) {
		q := p.Plan().Sub(c).Rotate(-rot)
		l := V(q.X/1.45, p.Y-base, q.Y)
		depth = math.Inf(1)
		for i, n := range normals {
			e := dists[i] - n.Dot(l)
			depth = math.Min(depth, e)
			if e < skin {
				near++
			}
		}
		return depth, near
	}
	b := Box{V(c.X-9, 0.3, c.Y-9), V(c.X+9, base+8, c.Y+9)}
	plinth := func(p Vec3) bool {
		q := p.Plan().Sub(c).Rotate(-rot)
		return p.Y <= base && math.Hypot(q.X/7.6, q.Y/5.6) <= 1
	}
	s.shell(Stone, b, 0.9, false, plinth)
	s.g.Use(Wood).Fill(Block(V(marinaBaySandsShore(c.Y)-0.5, 0.9, c.Y-1), V(c.X-5.5, 1.5, c.Y+1)))
	s.put(Glass, b, func(p Vec3) bool {
		d, _ := measure(p)
		return p.Y >= base && d >= 0 && d < skin
	})
	s.tint(Box{V(b.Min.X, base, b.Min.Z), b.Max}, func(_ Cell, p Vec3) (Color, bool) {
		_, near := measure(p)
		return White, near >= 2
	})
}

// museum builds the ArtScience Museum: ten hollow fingers of different lengths curving
// up and out from a glass lobby drum, with skylights in their cut tips.
func (s *marinaBaySandsSite) museum() {
	m := marinaBaySandsMuseum
	const base = marinaBaySandsG
	const hub = base + 5.6
	type finger struct {
		dir        Vec2
		reach, tip float64
	}
	// The tallest fingers stand at the back of the hero view so the bowl stays open to it.
	tall := math.Atan2(-0.69, -0.56)
	var fingers [10]finger
	for i := range fingers {
		a := tall + (float64(i)+0.5)*math.Pi/5
		k := 0.5 + 0.5*math.Cos(a-tall)
		alt := float64(i%2)*2 - 1
		fingers[i] = finger{P(math.Cos(a), math.Sin(a)), 9.6 + 4.6*k + 0.9*alt, hub + 5 + 11*k + 1.8*alt}
	}
	// measure reports where p sits in a finger: e ≤ 1 inside its section, k along its
	// reach, v above its center line as a fraction of the half height and w across it
	// as a fraction of the half width.
	measure := func(f finger, p Vec3) (e, k, v, w float64, ok bool) {
		d := p.Plan().Sub(m)
		r := d.Dot(f.dir)
		if r < 0.9 || r > f.reach {
			return 0, 0, 0, 0, false
		}
		u := d.Y*f.dir.X - d.X*f.dir.Y
		k = r / f.reach
		hw := 0.9 + 1.5*k
		if math.Abs(u) > hw {
			return 0, 0, 0, 0, false
		}
		rise := f.tip - hub
		slope := 1.6 * rise / f.reach * math.Pow(k, 0.6)
		hh := (1.2 + 0.9*k) * math.Sqrt(1+slope*slope)
		v = (p.Y - hub - rise*math.Pow(k, 1.6)) / hh
		w = math.Abs(u) / hw
		e = w*w + v*v
		return e, k, v, w, e <= 1
	}
	in := func(p Vec3) bool {
		for _, f := range fingers {
			if _, _, _, _, ok := measure(f, p); ok {
				return true
			}
		}
		return false
	}
	b := Box{V(m.X-16, hub-3, m.Y-16), V(m.X+16, hub+22, m.Y+16)}
	s.shell(White, b, 0.9, true, in)
	s.tint(b, func(_ Cell, p Vec3) (Color, bool) {
		best, shade, tip := 2.0, false, false
		for _, f := range fingers {
			if e, k, v, w, ok := measure(f, p); ok && e < best {
				// Shaded undersides, and creases where the petals merge into the bowl.
				best, shade = e, v < -0.45 || (w > 0.7 && k < 0.5)
				tip = f.reach*(1-k) < 1.0 && e < 0.5
			}
		}
		switch {
		case tip:
			return Glass, true
		case shade:
			return Stone, true
		}
		return 0, false
	})
	// Lobby drum under the bowl.
	drum := Box{V(m.X-3.4, base, m.Y-3.4), V(m.X+3.4, hub-0.5, m.Y+3.4)}
	s.put(Glass, drum, func(p Vec3) bool {
		r := p.Plan().Sub(m).Length()
		return r <= 3.3 && r > 2.4
	})
	s.tint(drum, func(_ Cell, p Vec3) (Color, bool) {
		return White, marinaBaySandsFmod(p.Plan().Sub(m).Angle(), math.Pi/4) < 0.28 || p.Y > hub-1.4
	})
}

// helix builds the double-helix footbridge across the channel, with the plain road
// bridge of Bayfront Avenue beside it.
func (s *marinaBaySandsSite) helix() {
	g := s.g
	const base = marinaBaySandsG
	a, b := P(46, 51.5), P(11, 72)
	ab := b.Sub(a)
	out := P(ab.Y, -ab.X).Scale(1 / ab.Length()) // Towards the bay.
	ctrl := a.Add(ab.Scale(0.5)).Add(out.Scale(4.5))
	at := func(t float64) Vec2 {
		u := 1 - t
		return a.Scale(u * u).Add(ctrl.Scale(2 * u * t)).Add(b.Scale(t * t))
	}
	// The deck arches high enough for the whole tube to clear the water.
	deck := func(t float64) float64 { return base + 0.9 + 3.3*math.Sin(math.Pi*t) }
	normal := func(t float64) Vec3 {
		d := at(t + 0.01).Sub(at(t - 0.01))
		d = d.Scale(1 / d.Length())
		return V(d.Y, 0, -d.X)
	}
	const steps, radius, turns = 160, 3.2, 2.75
	// Piers in the channel.
	for _, t := range []float64{0.36, 0.64} {
		g.Use(DarkStone).Fill(Capsule{at(t).At(0), at(t).At(deck(t) - 1), 0.7})
	}
	// Dark deck, three members wide, so the pale spirals read against it.
	for k := 0; k < steps; k++ {
		t0, t1 := float64(k)/steps, float64(k+1)/steps
		p0, p1 := at(t0).At(deck(t0)-0.4), at(t1).At(deck(t1)-0.4)
		n := normal(t0)
		for _, off := range []float64{-1, 0, 1} {
			g.Use(DarkStone).Fill(Capsule{p0.Add(n.Scale(off)), p1.Add(n.Scale(off)), 0.5})
		}
	}
	// Two intertwined spirals around the deck, narrowing at the landings.
	strand := func(t, phase float64) Vec3 {
		ang := 2*math.Pi*turns*t + phase
		r := radius * (0.55 + 0.45*marinaBaySandsSmooth(math.Min(t, 1-t)/0.14))
		c := at(t).At(deck(t) + 1.3)
		return c.Add(normal(t).Scale(r * math.Cos(ang))).Add(V(0, r*math.Sin(ang), 0))
	}
	for h, col := range []Color{White, Stone} {
		phase := float64(h) * math.Pi
		for k := 5; k < steps-5; k++ {
			t0, t1 := float64(k)/steps, float64(k+1)/steps
			g.Use(col).Fill(Capsule{strand(t0, phase), strand(t1, phase), 0.5})
		}
	}
	// Bayfront Bridge: a flat road deck on two piers.
	g.Use(DarkStone).Fill(Block(V(2.4, base-0.3, 55), V(6.6, base, 69)))
	g.Use(Stone).Fill(Block(V(2.4, base, 55), V(3.0, base+0.8, 69)), Block(V(6.0, base, 55), V(6.6, base+0.8, 69)))
	g.Use(DarkStone).Fill(Block(V(3.4, 0, 59.5), V(5.6, base-0.3, 60.5)), Block(V(3.4, 0, 64), V(5.6, base-0.3, 65)))
}

// supertree is one Supertree: a planted trunk and a flared funnel of branching ribs.
func (s *marinaBaySandsSite) supertree(c Vec2, h float64, seed int) {
	g := s.g
	const base = marinaBaySandsG
	neck, top := base+0.56*h, base+h
	rim := 0.26*h + 1
	trunk := Box{V(c.X-2.6, base, c.Y-2.6), V(c.X+2.6, neck, c.Y+2.6)}
	s.put(Foliage, trunk, func(p Vec3) bool {
		r := 2.3 - 0.8*(p.Y-base)/(neck-base)
		d := p.Plan().Sub(c).Length()
		return d <= r && d > r-1.15
	})
	// Planting panels in swatches that spiral up the trunk.
	s.tint(trunk, func(_ Cell, p Vec3) (Color, bool) {
		a := p.Plan().Sub(c).Angle()
		sector := int(math.Floor((a+math.Pi)/(2*math.Pi)*7 + (p.Y-base)*0.22))
		switch r := Hash01(sector, int(math.Floor((p.Y-base)/2.4)), seed); {
		case r < 0.2:
			return Terracotta, true
		case r < 0.36:
			return Red, true
		case r < 0.54:
			return Grass, true
		}
		return 0, false
	})
	n := 6
	if h > 22 {
		n = 8
	} else if h > 16 {
		n = 7
	}
	spread := math.Pi / float64(n) / 2
	rib := func(a0, branch, t float64) Vec3 {
		a := a0 + branch*spread*marinaBaySandsSmooth((t-0.4)/0.6)
		r := 1.3 + (rim-1.3)*t
		return Polar(c, r, a).At(neck - 1 + (top-neck+1)*(1-math.Pow(1-t, 2.2)))
	}
	const steps = 8
	for i := 0; i < n; i++ {
		a0 := 2*math.Pi*float64(i)/float64(n) + 0.4*float64(seed)
		for _, branch := range []float64{-1, 1} {
			for k := 0; k < steps; k++ {
				t0, t1 := float64(k)/steps, float64(k+1)/steps
				g.Use(Red).Fill(Capsule{rib(a0, branch, t0), rib(a0, branch, t1), 0.5})
			}
		}
	}
	// Rim ring tying the rib tips together.
	const ring = 40
	for k := 0; k < ring; k++ {
		a0, a1 := 2*math.Pi*float64(k)/ring, 2*math.Pi*float64(k+1)/ring
		g.Use(Terracotta).Fill(Capsule{Polar(c, rim, a0).At(top), Polar(c, rim, a1).At(top), 0.5})
	}
}

// gardens builds Gardens by the Bay: the Supertree Grove, the aerial walkway between
// two of the trees, the lake footbridge and the park trees.
func (s *marinaBaySandsSite) gardens() {
	g := s.g
	const base = marinaBaySandsG
	for i, t := range marinaBaySandsGrove {
		s.supertree(P(t[0], t[1]), t[2], i+1)
	}
	// The yellow skyway, curving between the two tallest neighbours.
	a, b := marinaBaySandsGrove[0], marinaBaySandsGrove[2]
	pa, pb := P(a[0], a[1]), P(b[0], b[1])
	ab := pb.Sub(pa)
	ctrl := pa.Add(ab.Scale(0.5)).Add(P(-ab.Y, ab.X).Scale(0.55))
	walk := func(t float64) Vec3 {
		u := 1 - t
		return pa.Scale(u * u).Add(ctrl.Scale(2 * u * t)).Add(pb.Scale(t * t)).At(base + 9.4)
	}
	for k := 0; k < 24; k++ {
		t0, t1 := float64(k)/24, float64(k+1)/24
		g.Use(Gold).Fill(Capsule{walk(t0), walk(t1), 0.75})
	}
	// Footbridge over Dragonfly Lake.
	g.Use(Wood).Fill(Block(V(-29, base, 12.2), V(-22.5, base+0.5, 13.8)))
	g.Use(Red).Fill(Block(V(-28.5, base+0.5, 12.2), V(-23, base+1.2, 12.6)), Block(V(-28.5, base+0.5, 13.4), V(-23, base+1.2, 13.8)))
	// Park trees on a jittered grid, clear of water, paths, trunks and buildings.
	clear := func(q Vec2) bool {
		if !marinaBaySandsLand(q.X, q.Y) || !marinaBaySandsLand(q.X-2.5, q.Y) || !marinaBaySandsLand(q.X, q.Y+2.5) ||
			!marinaBaySandsLand(q.X, q.Y-2.5) || marinaBaySandsLake(q.X, q.Y) || marinaBaySandsLake(q.X+2, q.Y) ||
			marinaBaySandsLake(q.X-2, q.Y) || s.covered(q.X, q.Y) {
			return false
		}
		for _, t := range marinaBaySandsGrove {
			if q.Sub(P(t[0], t[1])).Length() < 6 {
				return false
			}
		}
		for i := 0; i+1 < len(marinaBaySandsPaths); i++ {
			if marinaBaySandsSegment(q, marinaBaySandsPaths[i], marinaBaySandsPaths[i+1]) < 2.4 {
				return false
			}
		}
		for _, d := range []Vec2{marinaBaySandsFlower, marinaBaySandsCloud} {
			if q.Sub(d).Length() < 13 {
				return false
			}
		}
		return true
	}
	for ix := 0; ix < 6; ix++ {
		for iz := 0; iz < 19; iz++ {
			q := P(-48.5+5.4*float64(ix)+2.6*Hash01(ix, iz, 51), -47+5.6*float64(iz)+2.6*Hash01(ix, iz, 52))
			if q.X > -22.5 || Hash01(ix, iz, 53) < 0.25 || !clear(q) {
				continue
			}
			crown := Foliage
			if Hash01(ix, iz, 54) < 0.3 {
				crown = Grass
			}
			s.tree(q, 4+2*Hash01(ix, iz, 55), crown)
		}
	}
	// Trees on the north lawn, the far bank and the southern forecourt.
	for i, q := range [][2]float64{{-8, 44}, {-3, 50}, {12, 46}, {18, 51}, {11, 34}, {20, 38}, {-9, 73}, {-1, 74.5}, {20, 73},
		{27, 71.5}, {-17, 71.5}, {-16, 30}, {-4, 31}, {-15, -42}, {-6, -45}, {12, -48}, {22, -47.5}, {-24, -44}} {
		s.tree(P(q[0], q[1]), 4.2+1.6*Hash01(i, 56), Foliage)
	}
}

// dome is a conservatory: a hollow glass shell banded by white arch ribs.
func (s *marinaBaySandsSite) dome(c Vec2, rx, rz, h float64, round bool) {
	const base = marinaBaySandsG
	b := Box{V(c.X-rx, base, c.Y-rz), V(c.X+rx, base+h, c.Y+rz)}
	s.shell(Glass, b, 0.9, false, func(p Vec3) bool {
		dx, dy, dz := (p.X-c.X)/rx, (p.Y-base)/h, (p.Z-c.Y)/rz
		return dx*dx+dy*dy+dz*dz <= 1
	})
	s.tint(b, func(_ Cell, p Vec3) (Color, bool) {
		if p.Y < base+0.8 {
			return Stone, true
		}
		if round { // Ribs fanning from the apex.
			a := p.Plan().Sub(c).Angle()
			return White, marinaBaySandsFmod(a, math.Pi/5) < 0.16 && p.Plan().Sub(c).Length() > 1.5
		}
		return White, marinaBaySandsFmod(p.Z-c.Y+0.4, 3.6) < 0.8
	})
}

// boats puts bumboats and yachts on the bay and the channel.
func (s *marinaBaySandsSite) boats() {
	s.bumboat(P(49, -17), 1.45, Red)
	s.bumboat(P(53, 24), -1.2, Foliage)
	s.bumboat(P(-16, 62), 0.1, Red)
	s.yacht(P(56, 2), 1.8, White)
	s.yacht(P(49, -37), 1.2, Red)
}

// bumboat is a wooden river taxi with a painted gunwale and a canopy; angle is its heading.
func (s *marinaBaySandsSite) bumboat(c Vec2, angle float64, trim Color) {
	g := s.g
	const length, beam = 8.0, 2.8
	at := func(p Vec3) (float64, float64) {
		d := p.Plan().Sub(c).Rotate(-angle)
		return d.X, d.Y
	}
	b := Box{V(c.X-length, 0.2, c.Y-length), V(c.X+length, 3.4, c.Y+length)}
	hull := func(p Vec3) bool {
		a, w := at(p)
		return math.Abs(a) <= length/2 && math.Abs(w) <= beam/2*math.Min(1, (length/2-math.Abs(a))/1.8+0.3)
	}
	g.Use(Wood).Fill(Func{b, func(p Vec3) bool { return p.Y <= 1.2 && hull(p) }})
	g.Paint(trim, Func{b, func(p Vec3) bool { return p.Y > 0.8 && p.Y <= 1.2 && hull(p) }})
	g.Use(Sand).Fill(Func{b, func(p Vec3) bool {
		a, w := at(p)
		return p.Y > 1.2 && p.Y <= 2.2 && a > -2.6 && a < 1.4 && math.Abs(w) <= 0.9
	}})
	g.Use(trim).Fill(Func{b, func(p Vec3) bool {
		a, w := at(p)
		return p.Y > 2.2 && p.Y <= 2.8 && a > -3 && a < 1.8 && math.Abs(w) <= 1.3
	}})
	// Wake.
	g.Paint(White, Func{Box{V(c.X-12, -0.2, c.Y-12), V(c.X+12, 0.2, c.Y+12)}, func(p Vec3) bool {
		a, w := at(p)
		return a < -length/2 && a > -length/2-5 && math.Abs(math.Abs(w)-0.35*(-length/2-a)-0.6) < 0.5
	}})
}

// yacht is a sailing yacht with a mainsail and a jib; angle is its heading.
func (s *marinaBaySandsSite) yacht(c Vec2, angle float64, sail Color) {
	g := s.g
	const length = 10.0
	at := func(p Vec3) (float64, float64) {
		d := p.Plan().Sub(c).Rotate(-angle)
		return d.X, d.Y
	}
	b := Box{V(c.X-length, 0.2, c.Y-length), V(c.X+length, 14, c.Y+length)}
	g.Use(White).Fill(Func{b, func(p Vec3) bool {
		a, w := at(p)
		return p.Y <= 1.4 && math.Abs(a) <= length/2 && math.Abs(w) <= 1.5*math.Min(1, (length/2-a)/3)
	}})
	g.Paint(Navy, Func{b, func(p Vec3) bool { return p.Y > 0.7 && p.Y <= 1.1 }})
	mast := c.Add(P(0.6, 0).Rotate(angle))
	g.Use(Wood).Fill(Capsule{mast.At(1.2), mast.At(13.5), 0.5})
	g.Use(sail).Fill(Func{b, func(p Vec3) bool {
		a, w := at(p)
		a -= 0.6
		if p.Y < 2.4 || p.Y > 13 || math.Abs(w) > 0.5 {
			return false
		}
		t := (p.Y - 2.4) / 10.6
		return (a <= -0.5 && a >= -0.5-4.2*(1-t)) || (a >= 0.5 && a <= 0.5+3.6*(1-t) && p.Y < 12)
	}})
}
