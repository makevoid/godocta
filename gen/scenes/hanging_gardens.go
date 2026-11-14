package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// HangingGardens is the legendary terraced garden of Babylon, reconstructed as a stepped
// garden mountain. Five receding mud-brick storeys stand on open arcades of round arches
// under navy and gold glazed friezes; every terrace is planted with date palms,
// cypresses, broadleaf and blossom trees and flowering shrubs, and vines spill over the
// parapets in hanging curtains. A lifting tower with a great wheel and a chain of
// buckets stands at the back (-Z) and feeds an aqueduct to the summit reservoir, from
// which cascades fall terrace to terrace on both sides of the grand stair on the front
// (+Z) face and down the east (+X) face into basins and channels. A blue-glazed gate
// with gold and white beasts opens through the crenellated palace wall at the front
// left, a colonnaded court with a pool lies front right, and the Euphrates runs along
// the east edge with a quay, a jetty, reed boats and palms.
type HangingGardens struct{}

func init() { Register(HangingGardens{}) }

func (HangingGardens) Name() string  { return "hanging_gardens" }
func (HangingGardens) Title() string { return "Hanging Gardens" }

const (
	hangingGardensG      = 1.0   // Ground level.
	hangingGardensCX     = -8.0  // Garden axis: lifting tower, reservoir and grand stair.
	hangingGardensLevels = 5     // Arcaded storeys, each carrying a terrace.
	hangingGardensStorey = 7.0   // Storey height.
	hangingGardensBay    = 6.0   // Arcade bay; also the terrace depth in front and at the sides.
	hangingGardensBack   = 3.0   // Terrace depth at the back.
	hangingGardensHalf   = 37.0  // Half width of the ground storey.
	hangingGardensFront  = 22.0  // Front (+Z) face of the ground storey.
	hangingGardensRear   = -36.0 // Back face of the ground storey.
	hangingGardensWall   = 1.2   // Arcade wall thickness.
	hangingGardensRail   = 1.2   // Parapet height.
	hangingGardensArch   = 2.0   // Arch half width.
	hangingGardensSpring = 2.4   // Arch springing above the storey floor.
	hangingGardensQuay   = 42.5  // West bank of the Euphrates.
	hangingGardensSiteX  = 57.0  // Site half extents and corner radius.
	hangingGardensSiteZ  = 50.0
	hangingGardensSiteR  = 9.0
	hangingGardensGateX  = -37.0 // Axis of the glazed gate.
	hangingGardensWallZ  = 44.4  // Front palace wall.
)

// hangingGardensRect is a plan rectangle.
type hangingGardensRect struct{ x0, x1, z0, z1 float64 }

func (r hangingGardensRect) has(x, z float64) bool {
	return x >= r.x0 && x <= r.x1 && z >= r.z0 && z <= r.z1
}

func (r hangingGardensRect) inset(d float64) hangingGardensRect {
	return hangingGardensRect{r.x0 + d, r.x1 - d, r.z0 + d, r.z1 - d}
}

func (r hangingGardensRect) box(y0, y1 float64) Box {
	return Box{V(r.x0, y0, r.z0), V(r.x1, y1, r.z1)}
}

// hangingGardensLevel is the footprint of storey i.
func hangingGardensLevel(i int) hangingGardensRect {
	f := float64(i)
	return hangingGardensRect{
		hangingGardensCX - hangingGardensHalf + hangingGardensBay*f,
		hangingGardensCX + hangingGardensHalf - hangingGardensBay*f,
		hangingGardensRear + hangingGardensBack*f,
		hangingGardensFront - hangingGardensBay*f,
	}
}

// hangingGardensFloor is the floor of storey i, which is the terrace on storey i-1.
func hangingGardensFloor(i int) float64 {
	return hangingGardensG + hangingGardensStorey*float64(i)
}

// hangingGardensNoise is smooth value noise in [0, 1).
func hangingGardensNoise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

// hangingGardensFBM is three octaves of value noise in [0, 1).
func hangingGardensFBM(x, z float64, seed int) float64 {
	return (hangingGardensNoise(x, z, seed)*4 + hangingGardensNoise(x*2.1, z*2.1, seed+1)*2 +
		hangingGardensNoise(x*4.3, z*4.3, seed+2)) / 7
}

func hangingGardensHash(c Cell, seed int) float64 {
	return Hash01(int(c.X), int(c.Y), int(c.Z), seed)
}

func hangingGardensMod(v, m float64) float64 { return math.Mod(math.Mod(v, m)+m, m) }

// hangingGardensBeast draws the striding beasts of the glazed gate: three rows of
// pixel animals on a face, u from the left of the animal and v above the ground.
func hangingGardensBeast(u, v, dir float64) (Color, bool) {
	const px, pitch = 0.8, 4.0
	row := int(math.Floor((v - 1.4) / pitch))
	if row < 0 || row > 2 {
		return 0, false
	}
	i, j := int(math.Floor(u/px)), int(math.Floor((v-1.4-pitch*float64(row))/px))
	if i < 0 || i > 5 || j > 3 {
		return 0, false
	}
	if dir < 0 {
		i = 5 - i
	}
	on := false
	switch j {
	case 3:
		on = i >= 4 // Horned head.
	case 2:
		on = true
	case 1:
		on = i <= 4
	case 0:
		on = i == 0 || i == 4 // Legs.
	}
	if !on {
		return 0, false
	}
	if row%2 == 1 {
		return White, true
	}
	return Gold, true
}

// each visits every lattice cell whose world center lies in the box. The helpers built
// on it work on exact centers, without the grid's half-unit snapping, so thin decks and
// shells keep a predictable number of courses.
func (HangingGardens) each(g *Grid, b Box, f func(c Cell, p Vec3)) {
	s := g.Span()
	lo := func(v float64) int32 { return int32(math.Ceil(v*2/s - 1e-9)) }
	hi := func(v float64) int32 { return int32(math.Floor(v*2/s + 1e-9)) }
	y0, y1 := lo(b.Min.Y), hi(b.Max.Y)
	z0, z1 := lo(b.Min.Z), hi(b.Max.Z)
	x0, x1 := lo(b.Min.X), hi(b.Max.X)
	for y := y0; y <= y1; y++ {
		par := y & 1
		for z := z0 + ((z0&1)^par)&1; z <= z1; z += 2 {
			for x := x0 + ((x0&1)^par)&1; x <= x1; x += 2 {
				c := Cell{X: x, Y: y, Z: z}
				f(c, V(float64(x), float64(y), float64(z)).Scale(s/2))
			}
		}
	}
}

func (HangingGardens) set(g *Grid, c Cell, col Color) { g.Remove(c); g.Use(col).Add(c) }

// plot occupies cells in a box from a per-cell function.
func (s HangingGardens) plot(g *Grid, b Box, f func(c Cell, p Vec3) (Color, bool)) {
	s.each(g, b, func(c Cell, p Vec3) {
		if col, ok := f(c, p); ok {
			s.set(g, c, col)
		}
	})
}

// shade recolors occupied cells in a box from a per-cell function.
func (s HangingGardens) shade(g *Grid, b Box, f func(c Cell, p Vec3) (Color, bool)) {
	s.each(g, b, func(c Cell, p Vec3) {
		if !g.Has(c) {
			return
		}
		if col, ok := f(c, p); ok {
			s.set(g, c, col)
		}
	})
}

// cut removes the cells of a box that a predicate selects.
func (s HangingGardens) cut(g *Grid, b Box, f func(p Vec3) bool) {
	s.each(g, b, func(c Cell, p Vec3) {
		if f(p) {
			g.Remove(c)
		}
	})
}

// fill occupies the cells whose centers lie in any of the solids.
func (s HangingGardens) fill(g *Grid, col Color, solids ...Solid) {
	for _, sol := range solids {
		s.each(g, sol.Bounds(), func(c Cell, p Vec3) {
			if sol.Contains(p) {
				s.set(g, c, col)
			}
		})
	}
}

// course is the index of the highest lattice course at or below y.
func (HangingGardens) course(g *Grid, y float64) int32 {
	return int32(math.Floor(y*2/g.Span() + 1e-6))
}

// deck lays n courses, the highest at or just below top, over a rectangle.
func (s HangingGardens) deck(g *Grid, r hangingGardensRect, top float64, n int32, f func(c Cell, p Vec3) (Color, bool)) {
	k, h := s.course(g, top), g.Span()/2
	s.plot(g, r.box(float64(k-n+1)*h-0.01, float64(k)*h+0.01), f)
}

// site is the rounded rectangle of the ground plate.
func (HangingGardens) site(x, z float64) bool {
	const r = hangingGardensSiteR
	ax, az := math.Abs(x), math.Abs(z)
	if ax > hangingGardensSiteX || az > hangingGardensSiteZ {
		return false
	}
	dx, dz := ax-(hangingGardensSiteX-r), az-(hangingGardensSiteZ-r)
	return dx <= 0 || dz <= 0 || dx*dx+dz*dz <= r*r
}

// bank is the far shore of the river at z.
func (HangingGardens) bank(z float64) float64 {
	return 52.4 + 1.1*math.Sin(z*0.14+0.6) + 0.7*math.Sin(z*0.31)
}

// stream colors river water: long pale streaks drawn out along the current.
func (HangingGardens) stream(p Vec3) Color {
	if hangingGardensFBM(p.X*0.5, p.Z*0.09, 21) > 0.62 {
		return Glass
	}
	return Water
}

// soil colors open ground: sand with soft pale patches.
func (HangingGardens) soil(p Vec3) Color {
	if p.X > hangingGardensQuay && hangingGardensFBM(p.X*0.3, p.Z*0.3, 25) > 0.45 {
		return Grass // The reedy far bank.
	}
	if hangingGardensFBM(p.X*0.09, p.Z*0.09, 23) > 0.62 {
		return Stone
	}
	return Sand
}

// turf colors lawns and planting beds.
func (HangingGardens) turf(p Vec3) Color {
	if hangingGardensFBM(p.X*0.17, p.Z*0.17, 31) > 0.55 {
		return Foliage
	}
	return Grass
}

// slabs colors limestone paving with a brick border around r.
func (HangingGardens) slabs(c Cell, p Vec3, r hangingGardensRect) Color {
	switch {
	case !r.inset(1.1).has(p.X, p.Z):
		return Brick
	case hangingGardensMod(p.X, 2.4) < 0.45 || hangingGardensMod(p.Z, 2.4) < 0.45:
		return Stone
	case Hash01(int(math.Floor(p.X/2.4)), int(math.Floor(p.Z/2.4)), 27) < 0.12:
		return Sand // The odd darker flagstone.
	}
	return White
}

func (s HangingGardens) ground(g *Grid) {
	const G = hangingGardensG
	k := s.course(g, G)
	hall := hangingGardensLevel(0).inset(hangingGardensWall + 0.1)
	box := Box{V(-hangingGardensSiteX, -1, -hangingGardensSiteZ), V(hangingGardensSiteX, G, hangingGardensSiteZ)}
	s.each(g, box, func(c Cell, p Vec3) {
		if !s.site(p.X, p.Z) {
			return
		}
		switch {
		case p.X > hangingGardensQuay && p.X < s.bank(p.Z):
			// The Euphrates lies one course below the land.
			if c.Y == k-1 || c.Y == k-2 {
				s.set(g, c, s.stream(p))
			}
		case hall.has(p.X, p.Z):
			// Under the vaults a single dark course is enough.
			if c.Y == k {
				s.set(g, c, Charcoal)
			}
		case c.Y == k || c.Y == k-1:
			s.set(g, c, s.soil(p))
		}
	})
}

// grounds lays out lawns and paving around the garden mountain.
func (s HangingGardens) grounds(g *Grid) {
	const G, CX, GX = hangingGardensG, hangingGardensCX, hangingGardensGateX
	k := s.course(g, G)
	paint := func(r hangingGardensRect, f func(c Cell, p Vec3) Color) {
		s.shade(g, r.box(0, G), func(c Cell, p Vec3) (Color, bool) {
			if c.Y != k && c.Y != k-1 {
				return 0, false
			}
			return f(c, p), true
		})
	}
	for _, r := range []hangingGardensRect{
		{-50.5, -46.4, -44, 23},      // West walk.
		{-50.5, -22.5, 24.4, 38},     // Forecourt garden.
		{-50.5, 40, -44.6, -40.4},    // Behind the mountain.
		{30.6, 39.4, -35, 21.5},      // River garden.
		{6.5, 39.4, 23.6, 29.4},      // Before the court.
		{-21, -14.5, 23.6, 33},       // West of the stair.
		{-1.5, 5, 23.6, 33},          // East of the stair.
		{CX - 34.5, CX - 10, 23, 24}, // Planting strips at the foot of the arcade.
		{CX + 10, CX + 36, 23, 24},
	} {
		paint(r, func(c Cell, p Vec3) Color { return s.turf(p) })
	}
	for _, r := range []hangingGardensRect{
		{CX - 13, CX + 13, 33.2, 48.6},   // Plaza at the foot of the stair.
		{GX + 3, CX - 13, 39.2, 43.4},    // Way from the gate.
		{GX - 3, GX + 3, 39.2, 49.4},     // Gate passage.
		{CX - 4.6, CX + 4.6, 22.6, 33.2}, // Under the stair.
		{CX + 14.5, 39.4, 30.2, 48.4},    // Colonnaded court.
		{-50.5, GX - 3, 39.2, 43.4},      // Inside the front wall.
	} {
		rr := r
		paint(r, func(c Cell, p Vec3) Color { return s.slabs(c, p, rr) })
	}
}

// pool builds a rimmed basin whose water stands one course proud of a floor at y.
func (s HangingGardens) pool(g *Grid, r hangingGardensRect, y float64) {
	k, h := s.course(g, y), g.Span()/2
	in := r.inset(1.0)
	s.plot(g, r.box(float64(k+1)*h-0.01, float64(k+2)*h+0.01), func(c Cell, p Vec3) (Color, bool) {
		if !in.has(p.X, p.Z) {
			return White, true
		}
		if c.Y != k+1 {
			return 0, false
		}
		if hangingGardensFBM(p.X*0.45, p.Z*0.45, 33) > 0.63 {
			return Glass, true
		}
		return Water, true
	})
}

// breach opens a pool rim where a channel passes through it.
func (s HangingGardens) breach(g *Grid, r hangingGardensRect, y float64) {
	k := s.course(g, y)
	s.each(g, r.box(y, y+1.2), func(c Cell, p Vec3) {
		if !g.Has(c) {
			return
		}
		if c.Y == k+1 {
			s.set(g, c, Water)
		} else if c.Y > k+1 {
			g.Remove(c)
		}
	})
}

// rill recolors a strip of deck or ground at y as running water, with foam around the
// point where a fall lands on it.
func (s HangingGardens) rill(g *Grid, r hangingGardensRect, y float64, foam *Vec2) {
	k := s.course(g, y)
	s.shade(g, r.box(y-1.2, y+0.05), func(c Cell, p Vec3) (Color, bool) {
		if c.Y > k || c.Y < k-1 {
			return 0, false
		}
		if foam != nil && p.Plan().Sub(*foam).Length() < 2.2 && hangingGardensHash(c, 35) < 0.7 {
			return White, true
		}
		return Water, true
	})
}

// fall hangs a sheet of water with foam at its foot in a box against a wall face.
func (s HangingGardens) fall(g *Grid, b Box) {
	s.plot(g, b, func(c Cell, p Vec3) (Color, bool) {
		h := hangingGardensHash(c, 37)
		if (p.Y < b.Min.Y+1.7 && h < 0.8) || h < 0.14 {
			return White, true
		}
		if h > 0.9 {
			return Glass, true
		}
		return Water, true
	})
}

// basins lays the ground-level pools and channels that collect the cascades.
func (s HangingGardens) basins(g *Grid) {
	const G, CX, Q = hangingGardensG, hangingGardensCX, hangingGardensQuay
	front := hangingGardensFront
	for _, sx := range []float64{-1, 1} {
		a, b := CX+sx*5.1, CX+sx*10.5
		s.pool(g, hangingGardensRect{min(a, b), max(a, b), front - 1, front + 5.5}, G)
		s.rill(g, hangingGardensRect{CX + sx*7.8 - 1, CX + sx*7.8 + 1, front + 5.4, 40}, G, nil)
		s.breach(g, hangingGardensRect{CX + sx*7.8 - 1, CX + sx*7.8 + 1, front + 4.4, front + 5.6}, G)
	}
	// The cross channel feeds the court pool, which drains to the river.
	s.rill(g, hangingGardensRect{CX - 8.8, 20, 38, 40}, G, nil)
	s.pool(g, hangingGardensRect{19, 33, 36, 42}, G)
	s.breach(g, hangingGardensRect{18.9, 20.1, 38, 40}, G)
	s.breach(g, hangingGardensRect{31.9, 33.1, 38, 40}, G)
	s.rill(g, hangingGardensRect{33, Q, 38, 40}, G, nil)
	// The east basin and its outfall.
	e := hangingGardensLevel(0).x1
	s.pool(g, hangingGardensRect{e - 1, e + 7.5, -13.5, -4.5}, G)
	s.breach(g, hangingGardensRect{e + 6.4, e + 7.6, -10, -8}, G)
	s.rill(g, hangingGardensRect{e + 7.4, Q, -10, -8}, G, nil)
	// The supply canal from the river to the foot of the lifting tower.
	s.rill(g, hangingGardensRect{CX - 5.6, Q, -38.8, -36.8}, G, nil)
}

// bay returns the offset of a wall point from the axis of its arch, or false where the
// face is blank. Faces: 0 front (+Z), 1 back, 2 east (+X), 3 west.
func (HangingGardens) bay(i, face int, p Vec3) (float64, bool) {
	const m = hangingGardensBay
	r := hangingGardensLevel(i)
	if face < 2 {
		u := p.X - hangingGardensCX
		if math.Abs(u) > hangingGardensHalf-m*float64(i)-1 || (face == 0 && math.Abs(u) < m) {
			return 0, false // Corner piers, and the bays behind the grand stair.
		}
		return hangingGardensMod(u, m) - m/2, true
	}
	// Side arcades are set out from the front corner; what is left at the back is blank.
	w := r.z1 - 1 - p.Z
	if w < 0 || (math.Floor(w/m)+1)*m+3 > r.z1-r.z0 {
		return 0, false
	}
	return hangingGardensMod(w, m) - m/2, true
}

// masonry colors a wall cell v above its storey floor and d behind the face.
func (HangingGardens) masonry(c Cell, p Vec3, v, d, a float64, arcade bool) Color {
	const hw, spring = hangingGardensArch, hangingGardensSpring
	switch {
	case d > 0.8 && v < hangingGardensStorey-0.1:
		return Charcoal // The shadowed inner face of the vaults.
	case v >= hangingGardensStorey+hangingGardensRail-0.42:
		return Terracotta // Coping.
	case v >= 6.05 && v < 6.5:
		return Gold
	case v >= 4.85 && v < 6.05:
		if arcade && math.Abs(a) > hangingGardensBay/2-0.75 && v >= 5.1 && v < 5.85 {
			return Gold // A rosette over every pier.
		}
		return Navy
	case v < 0.8:
		return Terracotta
	}
	if arcade && v > spring-0.45 {
		da := math.Abs(a) - hw
		if v > spring {
			da = math.Hypot(a, v-spring) - hw
		}
		if da < 0.75 {
			return Brick // Voussoirs.
		}
	}
	if hangingGardensFBM((p.X+p.Z)*0.3, p.Y*0.5, 11) > 0.64 {
		return Stone
	}
	if hangingGardensHash(c, 12) < 0.03 {
		return Terracotta
	}
	return Sand
}

// storey builds one arcaded storey: the terrace deck it carries, then the wall of open
// arches with its glazed frieze and parapet.
func (s HangingGardens) storey(g *Grid, i int) {
	const hw, spring = hangingGardensArch, hangingGardensSpring
	r := hangingGardensLevel(i)
	y0, y1 := hangingGardensFloor(i), hangingGardensFloor(i+1)
	in := r.inset(hangingGardensWall)
	open := i+1 < hangingGardensLevels
	var hole hangingGardensRect
	if open {
		hole = hangingGardensLevel(i + 1).inset(hangingGardensWall)
	}
	s.deck(g, r, y1, 2, func(c Cell, p Vec3) (Color, bool) {
		if open && hole.has(p.X, p.Z) {
			return 0, false
		}
		return s.turf(p), true
	})
	s.plot(g, r.box(y0, y1+hangingGardensRail), func(c Cell, p Vec3) (Color, bool) {
		if in.has(p.X, p.Z) {
			return 0, false
		}
		// The nearest face and the depth behind it.
		face, d := 0, r.z1-p.Z
		if e := p.Z - r.z0; e < d {
			face, d = 1, e
		}
		if e := r.x1 - p.X; e < d {
			face, d = 2, e
		}
		if e := p.X - r.x0; e < d {
			face, d = 3, e
		}
		v := p.Y - y0
		a, ok := s.bay(i, face, p)
		if ok && math.Abs(a) <= hw && (v <= spring || a*a+(v-spring)*(v-spring) <= hw*hw) {
			return 0, false
		}
		return s.masonry(c, p, v, d, a, ok), true
	})
}

// stair builds the grand stair: one straight flight from the plaza to the summit, with
// glazed balustrades and a pair of gold lions at its foot.
func (s HangingGardens) stair(g *Grid) {
	const G, CX = hangingGardensG, hangingGardensCX
	const hw, wall = 5.0, 1.0
	top := hangingGardensFloor(hangingGardensLevels)
	zTop := hangingGardensLevel(hangingGardensLevels - 1).z1
	zFoot := zTop + top - G
	ramp := func(z float64) float64 { return G + zFoot - z }
	// terrace is the top of the masonry the flight passes over at z.
	terrace := func(z float64) float64 {
		y := G
		for i := 0; i < hangingGardensLevels; i++ {
			if z <= hangingGardensLevel(i).z1 {
				y = hangingGardensFloor(i + 1)
			}
		}
		return y
	}
	box := Box{V(CX-hw, G-0.3, zTop), V(CX+hw, top+1.4, zFoot+0.4)}
	s.plot(g, box, func(c Cell, p Vec3) (Color, bool) {
		y := ramp(p.Z)
		if math.Abs(p.X-CX) > hw-wall {
			// Side walls rising into balustrades.
			if p.Y > y+1.3 || p.Y < terrace(p.Z)-0.5 {
				return 0, false
			}
			switch {
			case p.Y > y+0.85:
				return Gold, true
			case p.Y > y-0.7:
				return Navy, true
			case hangingGardensMod(p.Y-G, hangingGardensStorey) < 0.8:
				return Terracotta, true
			case hangingGardensFBM(p.Z*0.3, p.Y*0.5, 13) > 0.64:
				return Stone, true
			}
			return Sand, true
		}
		if p.Y > y || p.Y < y-1.5 {
			return 0, false
		}
		switch {
		case hangingGardensMod(p.Y-G, hangingGardensStorey) > hangingGardensStorey-1.2:
			return Brick, true // A landing at every terrace.
		case int(math.Floor((p.Y-G)/0.8))%2 == 0:
			return White, true
		}
		return Sand, true
	})
	// Hollow the flight, and open the summit parapet where it arrives.
	s.cut(g, box, func(p Vec3) bool {
		return math.Abs(p.X-CX) < hw-wall-0.05 && p.Y > G+0.2 && p.Y < ramp(p.Z)-1.5
	})
	s.cut(g, Box{V(CX-hw+wall, top+0.05, zTop-hangingGardensWall-0.3), V(CX+hw-wall, top+3, zTop+0.2)},
		func(Vec3) bool { return true })
	// Glazed newels where the balustrades pass each terrace, taller at the summit.
	for i := 1; i <= hangingGardensLevels; i++ {
		y := hangingGardensFloor(i)
		z := zFoot - (y - G)
		h := 2.3
		if i == hangingGardensLevels {
			h = 3.9
		}
		for _, sx := range []float64{-1, 1} {
			x := CX + sx*(hw-0.7)
			s.fill(g, Navy, Block(V(x-0.7, y-1, z-0.9), V(x+0.7, y+h, z+0.9)))
			s.fill(g, Gold, Block(V(x-0.7, y+h, z-0.9), V(x+0.7, y+h+0.45, z+0.9)))
		}
	}
	// Newel pedestals with couchant lions.
	for _, sx := range []float64{-1, 1} {
		x := CX + sx*(hw-wall/2)
		s.fill(g, Navy, Block(V(x-1.1, G, zFoot-0.2), V(x+1.1, G+2.4, zFoot+3.4)))
		s.fill(g, Gold, Block(V(x-1.1, G+2.4, zFoot-0.2), V(x+1.1, G+2.85, zFoot+3.4)),
			Block(V(x-0.7, G+2.85, zFoot+0.2), V(x+0.7, G+4.0, zFoot+3.0)))
		g.Use(Gold).Fill(Sphere(V(x, G+4.4, zFoot+2.6), 0.85))
	}
}

// summit paves the top terrace around the reservoir.
func (s HangingGardens) summit(g *Grid) {
	const CX = hangingGardensCX
	y := hangingGardensFloor(hangingGardensLevels)
	top := hangingGardensLevel(hangingGardensLevels - 1)
	k := s.course(g, y)
	paved := hangingGardensRect{CX - 8.6, CX + 8.6, top.z0 + 4, top.z1}
	s.shade(g, top.box(y-1.2, y+0.05), func(c Cell, p Vec3) (Color, bool) {
		if c.Y > k || c.Y < k-1 || !paved.has(p.X, p.Z) {
			return 0, false
		}
		return s.slabs(c, p, paved), true
	})
	s.pool(g, hangingGardensRect{CX - 6, CX + 6, -17, -8}, y)
}

// cascades runs the water from the reservoir down the front, on both sides of the
// stair, and down the east face, falling from terrace to terrace.
func (s HangingGardens) cascades(g *Grid) {
	const CX = hangingGardensCX
	const wall, rail = hangingGardensWall, hangingGardensRail
	n := hangingGardensLevels
	all := func(Vec3) bool { return true }
	for _, sx := range []float64{-1, 1} {
		x0, x1 := CX+sx*6-1, CX+sx*6+1
		s.breach(g, hangingGardensRect{x0, x1, -9.1, -7.9}, hangingGardensFloor(n))
		s.rill(g, hangingGardensRect{x0, x1, -8, hangingGardensLevel(n - 1).z1}, hangingGardensFloor(n), nil)
		for j := n; j >= 1; j-- {
			f := hangingGardensLevel(j - 1).z1
			hi, lo := hangingGardensFloor(j), hangingGardensFloor(j-1)
			s.cut(g, Box{V(x0, hi+0.05, f-wall-0.05), V(x1, hi+rail+0.8, f+0.05)}, all)
			if j > 1 {
				foot := P(CX+sx*6, f+0.6)
				s.rill(g, hangingGardensRect{x0, x1, f, hangingGardensLevel(j - 2).z1}, lo, &foot)
			}
			s.fall(g, Box{V(x0, lo-0.3, f+0.02), V(x1, hi+0.05, f+0.85)})
		}
	}
	const z0, z1 = -10.0, -8.0
	s.breach(g, hangingGardensRect{CX + 4.9, CX + 6.1, z0, z1}, hangingGardensFloor(n))
	s.rill(g, hangingGardensRect{CX + 6, hangingGardensLevel(n - 1).x1, z0, z1}, hangingGardensFloor(n), nil)
	for j := n; j >= 1; j-- {
		e := hangingGardensLevel(j - 1).x1
		hi, lo := hangingGardensFloor(j), hangingGardensFloor(j-1)
		s.cut(g, Box{V(e-wall-0.05, hi+0.05, z0), V(e+0.05, hi+rail+0.8, z1)}, all)
		if j > 1 {
			foot := P(e+0.6, (z0+z1)/2)
			s.rill(g, hangingGardensRect{e, hangingGardensLevel(j - 2).x1, z0, z1}, lo, &foot)
		}
		s.fall(g, Box{V(e+0.02, lo-0.3, z0), V(e+0.85, hi+0.05, z1)})
	}
}

// merlons crowns a rectangle with stepped Babylonian merlons along its four edges.
func (s HangingGardens) merlons(g *Grid, r hangingGardensRect, top, w, h float64, col, tip Color) {
	const t = 0.9
	row := func(alongX bool, a0, a1, c0, c1 float64) {
		n := int((a1 - a0 + w*0.5) / (w * 1.5))
		if n < 1 {
			return
		}
		step := 0.0
		if n > 1 {
			step = (a1 - a0 - w) / float64(n-1)
		}
		for i := 0; i < n; i++ {
			m := a0 + w/2 + float64(i)*step
			if n == 1 {
				m = (a0 + a1) / 2
			}
			for k, half := range []float64{w / 2, w / 4} {
				y0, y1 := top+float64(k)*h/2, top+float64(k+1)*h/2
				cc := col
				if k == 1 {
					cc = tip
				}
				if alongX {
					s.fill(g, cc, Block(V(m-half, y0, c0), V(m+half, y1, c1)))
				} else {
					s.fill(g, cc, Block(V(c0, y0, m-half), V(c1, y1, m+half)))
				}
			}
		}
	}
	row(true, r.x0, r.x1, r.z1-t, r.z1)
	row(true, r.x0, r.x1, r.z0, r.z0+t)
	row(false, r.z0, r.z1, r.x0, r.x0+t)
	row(false, r.z0, r.z1, r.x1-t, r.x1)
}

// tower builds the lifting tower behind the mountain: a tall mud-brick shaft with a
// great wheel and a chain of buckets on its south face, and the aqueduct that carries
// the raised water over the back terraces to the summit reservoir.
func (s HangingGardens) tower(g *Grid) {
	const G, CX = hangingGardensG, hangingGardensCX
	const hw, z0, z1, h = 3.5, -46.0, -39.0, 50.0
	r := hangingGardensRect{CX - hw, CX + hw, z0, z1}
	in := r.inset(0.85)
	s.plot(g, r.box(G, G+h), func(c Cell, p Vec3) (Color, bool) {
		v := p.Y - G
		if in.has(p.X, p.Z) && v < h-0.85 {
			return 0, false
		}
		switch {
		case v > h-0.85:
			return Terracotta, true
		case v >= 48.6 && v < 49.15:
			return Gold, true
		case v >= 47.3 && v < 48.6:
			return Navy, true
		case v < 0.8 || hangingGardensMod(v, 5.6) < 0.8:
			return Terracotta, true
		case v > 8 && hangingGardensMod(v, 11.2) > 7.2 && hangingGardensMod(v, 11.2) < 9.6 &&
			(math.Abs(p.Z-(z0+z1)/2) < 0.55 || (math.Abs(p.X-CX) < 0.55 && p.Z < z0+1)):
			return Charcoal, true // Slit windows.
		case hangingGardensFBM((p.X+p.Z)*0.3, p.Y*0.5, 15) > 0.64:
			return Stone, true
		}
		return Sand, true
	})
	s.merlons(g, r, G+h, 1.6, 1.8, Sand, Terracotta)

	// The wheel on the south face, wider than the shaft that carries it.
	const R = 5.0
	wy, wz := G+43.0, z1+0.6
	s.plot(g, Box{V(CX-R-0.6, wy-R-0.6, wz-0.5), V(CX+R+0.6, wy+R+0.6, wz+0.5)}, func(c Cell, p Vec3) (Color, bool) {
		dx, dy := p.X-CX, p.Y-wy
		rr := math.Hypot(dx, dy)
		switch {
		case rr <= 1.15:
			return Gold, true
		case math.Abs(rr-R) <= 0.55:
			return Wood, true
		case rr < R:
			for k := 0; k < 2; k++ {
				sn, cs := math.Sincos(float64(k)*math.Pi/2 + 0.4)
				if math.Abs(dx*sn-dy*cs) <= 0.5 {
					return Wood, true
				}
			}
		}
		return 0, false
	})
	// The endless chain with its buckets: full ones rise on the east side.
	for _, sx := range []float64{-1, 1} {
		x := CX + sx*R
		s.fill(g, Charcoal, Block(V(x-0.5, G, wz-0.5), V(x+0.5, wy, wz+0.5)))
		for y := G + 3.0 + 1.75*(sx+1); y < wy-1.5; y += 3.5 {
			s.fill(g, Terracotta, Block(V(x-0.95, y, wz-0.95), V(x+0.95, y+1.4, wz+0.95)))
			if sx > 0 {
				s.fill(g, Water, Block(V(x-0.5, y+1.0, wz-0.5), V(x+0.5, y+1.8, wz+0.5)))
			}
		}
	}

	// The aqueduct: two arches over the back terraces, then a flume across the summit.
	const pier0, pier1 = -32.8, -30.2
	tr := G + 38.4
	under := tr - 1.5
	summit := hangingGardensFloor(hangingGardensLevels)
	rear := hangingGardensLevel(hangingGardensLevels - 1).z0
	s.plot(g, Box{V(CX-1.5, hangingGardensFloor(2), z1), V(CX+1.5, tr, -16.1)}, func(c Cell, p Vec3) (Color, bool) {
		ring := false
		switch {
		case p.Z > rear:
			if p.Y < summit {
				return 0, false
			}
		case p.Z < pier0 || p.Z > pier1:
			a, half := p.Z-(z1+pier0)/2, (pier0-z1)/2
			if p.Z > pier1 {
				a, half = p.Z-(pier1+rear)/2, (rear-pier1)/2
			}
			spring := under - 0.9 - half
			if p.Y < spring {
				return 0, false
			}
			dist := math.Hypot(a, p.Y-spring) - half
			if dist <= 0 {
				return 0, false
			}
			ring = dist < 0.8
		}
		switch {
		case p.Y > tr-0.85 && math.Abs(p.X-CX) < 0.6:
			return Water, true
		case p.Y > tr-0.45:
			return Terracotta, true
		case ring:
			return Brick, true
		case p.Y > under-0.2 && p.Y < under+0.5:
			return Navy, true
		}
		return Sand, true
	})
	// The spout into the reservoir.
	s.fall(g, Box{V(CX-0.6, summit, -16.1), V(CX+0.6, tr-0.5, -15.3)})
}

// curtain builds a crenellated mud-brick wall along X or Z from a0 to a1 at t.
func (s HangingGardens) curtain(g *Grid, alongX bool, a0, a1, t float64) {
	const G, th, h = hangingGardensG, 0.9, 6.0
	b := Box{V(a0, G, t-th/2), V(a1, G+h+1.05, t+th/2)}
	if !alongX {
		b = Box{V(t-th/2, G, a0), V(t+th/2, G+h+1.05, a1)}
	}
	s.plot(g, b, func(c Cell, p Vec3) (Color, bool) {
		u, v := p.X, p.Y-G
		if !alongX {
			u = p.Z
		}
		switch {
		case v > h:
			if hangingGardensMod(u, 2.4) > 1.2 {
				return 0, false
			}
			return Sand, true
		case v > h-0.85 || v < 0.8:
			return Terracotta, true
		case hangingGardensFBM(u*0.3, p.Y*0.5, 17) > 0.64:
			return Stone, true
		}
		return Sand, true
	})
}

// buttress is a projecting wall tower with a glazed band under its battlements.
func (s HangingGardens) buttress(g *Grid, x, z, wx, wz float64) {
	const G, h = hangingGardensG, 8.6
	r := hangingGardensRect{x - wx/2, x + wx/2, z - wz/2, z + wz/2}
	in := r.inset(0.85)
	s.plot(g, r.box(G, G+h+1.05), func(c Cell, p Vec3) (Color, bool) {
		v := p.Y - G
		switch {
		case v > h:
			dx, dz := math.Abs(p.X-x), math.Abs(p.Z-z)
			if in.has(p.X, p.Z) || (dx > 0.6 && dx < wx/2-1.2) || (dz > 0.6 && dz < wz/2-1.2) {
				return 0, false
			}
			return Sand, true
		case in.has(p.X, p.Z) && v < h-0.85:
			return 0, false
		case v > h-0.85 || v < 0.8:
			return Terracotta, true
		case v > h-2.2 && v <= h-1.25:
			return Navy, true
		case hangingGardensFBM((p.X+p.Z)*0.3, p.Y*0.5, 17) > 0.64:
			return Stone, true
		}
		return Sand, true
	})
}

// palace builds the crenellated palace walls on the west, north and front-left sides.
func (s HangingGardens) palace(g *Grid) {
	const GX, WZ = hangingGardensGateX, hangingGardensWallZ
	const west, north, east = -52.0, -46.0, 41.0
	s.curtain(g, false, north, WZ, west)
	s.curtain(g, true, west, east, north)
	s.curtain(g, true, west, GX-8, WZ)
	s.curtain(g, true, GX+8, -21, WZ)
	for _, z := range []float64{-30.8, -15.6, -0.4, 14.8, 29.8} {
		s.buttress(g, west-0.6, z, 3.2, 4.2)
	}
	for _, x := range []float64{-36.5, -21.5, 12, 26.5} {
		s.buttress(g, x, north-0.6, 4.2, 3.2)
	}
	for _, p := range []Vec2{P(west, north), P(west, WZ), P(east, north), P(-21, WZ)} {
		s.buttress(g, p.X, p.Y, 4.2, 4.2)
	}
}

// glazed builds one hollow blue-glazed gate tower with rows of beasts on every face.
func (s HangingGardens) glazed(g *Grid, r hangingGardensRect, h, dir float64) {
	const G = hangingGardensG
	in := r.inset(0.85)
	s.plot(g, r.box(G, G+h), func(c Cell, p Vec3) (Color, bool) {
		v := p.Y - G
		switch {
		case in.has(p.X, p.Z) && v < h-0.85:
			return 0, false
		case v > h-0.85:
			if in.has(p.X, p.Z) {
				return Sand, true // Roof terrace inside the white cornice.
			}
			return White, true
		case v > h-2.05 && v <= h-1.25:
			return Gold, true
		}
		// Beasts on the Z faces stride toward the passage, the others toward the front.
		u, w, f := p.X-r.x0, r.x1-r.x0, dir
		if d := min(r.x1-p.X, p.X-r.x0); d < min(r.z1-p.Z, p.Z-r.z0) {
			u, w, f = p.Z-r.z0, r.z1-r.z0, 1
		}
		if col, ok := hangingGardensBeast(u-(w-4.8)/2, v, f); ok {
			return col, true
		}
		return Navy, true
	})
	s.merlons(g, r, G+h, 1.6, 1.8, Navy, Gold)
}

// gate builds the blue-glazed gate in the front wall.
func (s HangingGardens) gate(g *Grid) {
	const G, GX = hangingGardensG, hangingGardensGateX
	const z0, z1, h, hc = 41.5, 48.5, 15.0, 11.0
	const half, spring = 2.2, 4.2
	mid := hangingGardensRect{GX - 3.2, GX + 3.2, z0 + 1, z1 - 1}
	s.plot(g, mid.box(G, G+hc), func(c Cell, p Vec3) (Color, bool) {
		a, v := p.X-GX, p.Y-G
		d := math.Abs(a) - half
		if v > spring {
			d = math.Hypot(a, v-spring) - half
		}
		face := p.Z > mid.z1-0.85 || p.Z < mid.z0+0.85
		switch {
		case d <= 0:
			return 0, false
		case v > hc-0.85:
			if !face {
				return Sand, true
			}
			return White, true
		case v > hc-2.05 && v <= hc-1.25:
			return Gold, true
		case face && d < 0.85:
			return Gold, true
		case face && d < 1.7 && hangingGardensMod(math.Atan2(v-spring, a)*3.2, 2) < 1 && v > spring:
			return White, true
		}
		return Navy, true
	})
	s.merlons(g, hangingGardensRect{mid.x0 + 0.4, mid.x1 - 0.4, mid.z0, mid.z1}, G+hc, 1.6, 1.8, Navy, Gold)
	for _, sx := range []float64{-1, 1} {
		a, b := GX+sx*3, GX+sx*9
		s.glazed(g, hangingGardensRect{min(a, b), max(a, b), z0, z1}, h, -sx)
	}
}

// court builds the colonnaded court around its pool.
func (s HangingGardens) court(g *Grid) {
	const G = hangingGardensG
	r := hangingGardensRect{10, 38, 31, 47}
	for x := r.x0; x <= r.x1; x += 4 {
		for z := r.z0; z <= r.z1; z += 4 {
			edge := x == r.x0 || x == r.x1 || z == r.z0 || z == r.z1
			if !edge || (z == 39 && (x == r.x0 || x == r.x1)) {
				continue
			}
			g.Use(Terracotta).Fill(Cylinder{Center: P(x, z), Radius: 0.85, Y0: G, Y1: G + 0.6})
			g.Use(White).Fill(Cylinder{Center: P(x, z), Radius: 0.6, Y0: G + 0.6, Y1: G + 4.4})
			s.fill(g, Gold, Block(V(x-0.85, G+4.3, z-0.85), V(x+0.85, G+4.85, z+0.85)))
		}
	}
	out, in := r.inset(-0.75), r.inset(0.75)
	s.plot(g, out.box(G+4.85, G+6.1), func(c Cell, p Vec3) (Color, bool) {
		if in.has(p.X, p.Z) {
			return 0, false
		}
		if p.Y > G+5.65 {
			return Gold, true
		}
		return Navy, true
	})
}

// boat is a reed boat with upturned ends, optionally under a square sail.
func (s HangingGardens) boat(g *Grid, at Vec2, angle, length float64, sail bool, cargo Color) {
	y := float64(s.course(g, hangingGardensG)) * g.Span() / 2 // Water surface.
	half := length / 2
	s.plot(g, Box{V(at.X-half-1, y-0.1, at.Y-half-1), V(at.X+half+1, y+5, at.Y+half+1)}, func(c Cell, p Vec3) (Color, bool) {
		q := p.Plan().Sub(at).Rotate(-angle)
		k := q.X / half
		if math.Abs(k) > 1 {
			return 0, false
		}
		k4 := k * k * k * k
		w := math.Max(1.35*math.Sqrt(1-k*k), 0.5)
		lift := 2.6 * k4 // Prow and stern curl up.
		if math.Abs(q.Y) > w || p.Y < y+lift-0.05 || p.Y > y+0.95+lift*1.1 {
			return 0, false
		}
		if math.Abs(k) < 0.55 && math.Abs(q.Y) < w-0.55 && p.Y > y+0.5 {
			if p.Y < y+0.95 && math.Abs(k) < 0.4 {
				return cargo, true
			}
			return 0, false
		}
		if hangingGardensMod(q.X, 1.7) < 0.45 {
			return Wood, true // Lashings.
		}
		return Sand, true
	})
	if !sail {
		return
	}
	turn := func(sol Solid) Solid { return Translate(RotateY(sol, P(0, 0), angle), at.At(0)) }
	g.Use(Wood).Fill(turn(Cylinder{Center: P(0.6, 0), Radius: 0.5, Y0: y + 0.3, Y1: y + 7.4}))
	g.Use(Wood).Fill(turn(Block(V(0.2, y+6.3, -2.4), V(1.0, y+7.0, 2.4))))
	g.Use(White).Fill(turn(Block(V(0.9, y+2.6, -2.2), V(1.6, y+6.3, 2.2))))
	g.Paint(Red, turn(Block(V(0.8, y+4.0, -2.3), V(1.7, y+4.9, 2.3))))
}

// river dresses the Euphrates: the quay, a jetty, mooring posts, reed boats and the
// reedy far bank.
func (s HangingGardens) river(g *Grid) {
	const G, Q = hangingGardensG, hangingGardensQuay
	k, h := s.course(g, G), g.Span()/2
	s.each(g, Box{V(Q-2.6, float64(k+1)*h-0.01, -hangingGardensSiteZ), V(Q, float64(k+1)*h+0.01, hangingGardensSiteZ)}, func(c Cell, p Vec3) {
		if !s.site(p.X, p.Z) {
			return
		}
		col := Stone
		switch {
		case p.X > Q-0.85:
			col = Brick
		case hangingGardensMod(p.Z, 3.2) < 0.5:
			col = DarkStone
		case hangingGardensHash(c, 41) < 0.25:
			col = White
		}
		s.set(g, c, col)
	})
	for z := -40.0; z <= 44; z += 7 {
		g.Use(Wood).Fill(Cylinder{Center: P(Q-0.6, z), Radius: 0.5, Y0: G + 0.4, Y1: G + 2.2})
	}
	// Jetty.
	jetty := hangingGardensRect{Q, Q + 6.5, 9, 12}
	s.deck(g, jetty, float64(k+1)*h, 2, func(c Cell, p Vec3) (Color, bool) { return Wood, true })
	for _, p := range []Vec2{P(Q+6, 9.4), P(Q+6, 11.6), P(Q+3, 9.4), P(Q+3, 11.6)} {
		g.Use(Wood).Fill(Cylinder{Center: p, Radius: 0.5, Y0: G - 0.6, Y1: G + 2.0})
	}
	s.boat(g, P(Q+3.6, 15.2), math.Pi/2, 9, false, Terracotta)
	s.boat(g, P(48.2, -21), math.Pi/2+0.12, 10, true, Red)
	s.boat(g, P(46.2, 33), math.Pi/2-0.2, 8.5, false, Gold)
	s.boat(g, P(47, -41), math.Pi/2+0.3, 8, false, Terracotta)
	// Reeds along the far bank.
	for i := 0; i < 44; i++ {
		z := -47 + 2.15*float64(i) + 0.8*Hash01(i, 43)
		x := s.bank(z) + 0.3 + 1.6*Hash01(i, 44)
		if !s.site(x+1, z) || Hash01(i, 45) < 0.25 {
			continue
		}
		top := G + 1.4 + 1.3*Hash01(i, 46)
		g.Use(Grass).Fill(Capsule{V(x, G-0.3, z), V(x+0.3, top, z+0.2), 0.5})
		g.Use(Gold).Fill(Sphere(V(x+0.3, top+0.3, z+0.2), 0.45))
	}
}

// palm is a date palm with a leaning trunk and drooping fronds.
func (s HangingGardens) palm(g *Grid, at Vec3, h float64, seed int) {
	a := Hash01(seed, 1) * 2 * math.Pi
	lean := 0.4 + 0.8*Hash01(seed, 2)
	sn, cs := math.Sincos(a)
	mid := at.Add(V(cs*lean*0.3, h*0.55, sn*lean*0.3))
	top := at.Add(V(cs*lean, h, sn*lean))
	g.Use(Wood).Fill(Capsule{at, mid, 0.5}, Capsule{mid, top, 0.5})
	g.Use(Terracotta).Fill(Sphere(top.Add(V(0, -0.6, 0)), 0.7))
	for k := 0; k < 7; k++ {
		b := a + 2*math.Pi*float64(k)/7
		d := V(math.Cos(b), 0, math.Sin(b))
		m := top.Add(d.Scale(1.6)).Add(V(0, 0.5, 0))
		tip := top.Add(d.Scale(3.0)).Add(V(0, -1.0, 0))
		col := Foliage
		if k%3 == 0 {
			col = Grass
		}
		g.Use(col).Fill(Capsule{top, m, 0.5}, Capsule{m, tip, 0.45})
	}
}

// cypress is a dark column of foliage.
func (s HangingGardens) cypress(g *Grid, at Vec3, h float64) {
	g.Use(Wood).Fill(Cylinder{Center: at.Plan(), Radius: 0.5, Y0: at.Y, Y1: at.Y + 1.2})
	g.Use(Foliage).FillVolume(Frustum{at.Plan(), 1.3, 0.45, at.Y + 1.0, at.Y + h})
	g.Use(Foliage).Fill(Sphere(at.Add(V(0, h, 0)), 0.5))
}

// tree is a broadleaf tree; some are in white or red blossom.
func (s HangingGardens) tree(g *Grid, at Vec3, r float64, seed int) {
	th := 1.5 + 1.2*Hash01(seed, 3)
	g.Use(Wood).Fill(Cylinder{Center: at.Plan(), Radius: 0.55, Y0: at.Y, Y1: at.Y + th + 0.8})
	c := at.Add(V(0, th+0.75*r, 0))
	g.Use(Foliage).FillVolume(Ellipsoid{c, V(r, 0.8*r, r)})
	sn, cs := math.Sincos(Hash01(seed, 4) * 2 * math.Pi)
	crown := Grass
	switch b := Hash01(seed, 5); {
	case b < 0.14:
		crown = White
	case b < 0.26:
		crown = Red
	}
	g.Use(crown).FillVolume(Ellipsoid{c.Add(V(cs*0.3*r, 0.42*r, sn*0.3*r)), V(0.64*r, 0.5*r, 0.64*r)})
}

// shrub is a low bush with a cap of flowers.
func (s HangingGardens) shrub(g *Grid, at Vec3, r float64, bloom Color) {
	c := at.Add(V(0, r*0.6, 0))
	g.Use(Foliage).Fill(Sphere(c, r))
	s.shade(g, Box{c.Sub(V(r, r, r)), c.Add(V(r, r, r))}.Grow(0.3), func(cc Cell, p Vec3) (Color, bool) {
		col, _ := g.ColorOf(cc)
		return bloom, col == Foliage && p.Y > c.Y+0.15 && p.Sub(c).Length() <= r+0.3
	})
}

// plant sets one plant of a hashed kind; small keeps to cypresses and shrubs.
func (s HangingGardens) plant(g *Grid, at Vec3, seed int, small bool) {
	blooms := []Color{Red, Gold, White}
	bloom := blooms[int(Hash01(seed, 7)*3)%3]
	k := Hash01(seed, 6)
	switch {
	case small && k < 0.4:
		s.cypress(g, at, 4.6+1.6*Hash01(seed, 8))
	case small:
		s.shrub(g, at, 1.0+0.3*Hash01(seed, 8), bloom)
	case k < 0.27:
		s.palm(g, at, 5.2+2.2*Hash01(seed, 8), seed)
	case k < 0.46:
		s.cypress(g, at, 5.6+2.6*Hash01(seed, 8))
	case k < 0.84:
		s.tree(g, at, 1.9+0.75*Hash01(seed, 8), seed)
	default:
		s.shrub(g, at, 1.25, bloom)
	}
}

// row plants a line from a to b at height y, with flowering shrubs toward out between
// the larger plants. skip masks stretches that must stay clear.
func (s HangingGardens) row(g *Grid, a, b, out Vec2, y float64, seed int, small bool, skip func(p Vec2) bool) {
	d := b.Sub(a)
	length := d.Length()
	if length < 0.5 {
		return
	}
	dir := d.Scale(1 / length)
	blooms := []Color{Red, Gold, White}
	for t, n := 0.0, 0; t <= length+0.1; n++ {
		p := a.Add(dir.Scale(t))
		step := 4.8 + 1.6*Hash01(seed, n, 1)
		if skip == nil || !skip(p) {
			s.plant(g, p.At(y), seed*131+n, small)
		}
		m := p.Add(dir.Scale(step / 2)).Add(out.Scale(1.5))
		if !small && t+step/2 <= length && (skip == nil || !skip(m)) && Hash01(seed, n, 2) < 0.7 {
			s.shrub(g, m.At(y), 0.95+0.25*Hash01(seed, n, 3), blooms[int(Hash01(seed, n, 4)*3)%3])
		}
		t += step
	}
}

// planting fills the terraces and the grounds with trees and shrubs.
func (s HangingGardens) planting(g *Grid) {
	const G, CX, Q = hangingGardensG, hangingGardensCX, hangingGardensQuay
	water := func(p Vec2) bool { // Keep the stair, the cascades and the aqueduct clear.
		return math.Abs(p.X-CX) < 9.6 || (p.X > CX && p.Y > -13.4 && p.Y < -4.6)
	}
	for j := 1; j < hangingGardensLevels; j++ {
		o := hangingGardensLevel(j - 1)
		y := hangingGardensFloor(j) - 0.3
		const m = 3.6
		s.row(g, P(o.x0+m, o.z1-m), P(o.x1-m, o.z1-m), P(0, 1), y, 10*j+1, false, water)
		s.row(g, P(o.x1-m, o.z1-m-4.8), P(o.x1-m, o.z0+4.5), P(1, 0), y, 10*j+2, false, water)
		s.row(g, P(o.x0+m, o.z1-m-4.8), P(o.x0+m, o.z0+4.5), P(-1, 0), y, 10*j+3, false, nil)
		s.row(g, P(o.x0+4.5, o.z0+2.1), P(o.x1-4.5, o.z0+2.1), P(0, -1), y, 10*j+4, true,
			func(p Vec2) bool { return math.Abs(p.X-CX) < 3.2 })
	}
	// The summit: cypresses at the corners, palms and blossom beside the reservoir.
	top := hangingGardensLevel(hangingGardensLevels - 1)
	y := hangingGardensFloor(hangingGardensLevels) - 0.3
	for i, sx := range []float64{-1, 1} {
		x := CX + sx*10.6
		s.cypress(g, V(x, y, top.z1-3), 6.5)
		s.cypress(g, V(x, y, top.z0+3), 7.5)
		s.palm(g, V(x, y, -17.5), 6.4, 500+i)
		if sx < 0 {
			s.tree(g, V(x-0.2, y, -9.5), 2.1, 510)
		}
		s.shrub(g, V(CX+sx*8, y, top.z0+2.6), 1.1, Red)
		s.shrub(g, V(CX+sx*4.5, y, top.z0+2.6), 1.1, White)
	}

	// The grounds.
	for i, z := range []float64{-33, -26, -19, -1, 6, 20, 27} {
		s.palm(g, V(Q-4.3+0.8*Hash01(i, 61), G-0.3, z), 7+2*Hash01(i, 62), 600+i)
	}
	river := func(p Vec2) bool { return p.Y > -15 && p.Y < -3 }
	s.row(g, P(33.6, -33), P(33.6, 20), P(-1, 0), G-0.3, 71, true, river)
	s.row(g, P(9, 26.4), P(37, 26.4), P(0, -1), G-0.3, 72, false, nil)
	s.row(g, P(-48.4, -42), P(-48.4, 20), P(1, 0), G-0.3, 73, false, nil)
	s.row(g, P(-48, 27.5), P(-25, 27.5), P(0, 1), G-0.3, 74, false, nil)
	s.row(g, P(-48, 35), P(-25, 35), P(0, -1), G-0.3, 75, false, nil)
	s.row(g, P(-46, -42.5), P(-14, -42.5), P(0, 1), G-0.3, 76, true, nil)
	s.row(g, P(10, -42.5), P(38, -42.5), P(0, 1), G-0.3, 77, true, nil)
	for _, sx := range []float64{-1, 1} {
		s.cypress(g, V(CX+sx*11.5, G-0.3, 36), 7.5)
		s.cypress(g, V(CX+sx*11.5, G-0.3, 46), 7.5)
		s.cypress(g, V(CX+sx*17.5, G-0.3, 28), 6.5)
	}
	for i, p := range []Vec2{P(54.5, -30), P(55, -8), P(54.6, 17), P(54.2, 36)} {
		if s.site(p.X+1, p.Y) {
			s.palm(g, p.At(G-0.3), 6.5+1.5*Hash01(i, 63), 620+i)
		}
	}
	for i, p := range []Vec2{P(40.5, 33), P(40.5, 45), P(7.5, 45), P(7.5, 33)} {
		s.palm(g, p.At(G-0.3), 6.5+1.5*Hash01(i, 64), 630+i)
	}
}

// vines hangs curtains of creeper over the parapets of every terrace.
func (s HangingGardens) vines(g *Grid) {
	const CX = hangingGardensCX
	for i := 0; i < hangingGardensLevels; i++ {
		r := hangingGardensLevel(i)
		top := hangingGardensFloor(i+1) + hangingGardensRail
		out := r.inset(-0.85)
		in := r.inset(hangingGardensWall)
		s.plot(g, out.box(top-6.5, top+0.5), func(c Cell, p Vec3) (Color, bool) {
			if in.has(p.X, p.Z) {
				return 0, false
			}
			// The face a strand hangs on and its place along it.
			face, u, d := 0, p.X, p.Z-r.z1
			if e := r.z0 - p.Z; e > d {
				face, u, d = 1, p.X, e
			}
			if e := p.X - r.x1; e > d {
				face, u, d = 2, p.Z, e
			}
			if e := r.x0 - p.X; e > d {
				face, u, d = 3, p.Z, e
			}
			if (d <= 0 && p.Y <= top) || (face < 2 && !(p.X > r.x0 && p.X < r.x1)) || (face >= 2 && !(p.Z > r.z0 && p.Z < r.z1)) {
				return 0, false // Inside the wall, or out past a corner.
			}
			switch {
			case face == 0 && math.Abs(u-CX) < 8.4:
				return 0, false // The stair and its cascades.
			case face == 1 && math.Abs(u-CX) < 5.5:
				return 0, false // The lifting chain and aqueduct.
			case face == 2 && u > -12 && u < -6:
				return 0, false // The east cascade.
			}
			n := hangingGardensFBM(u*0.17+float64(i)*7.3, float64(face)*13.7+float64(i)*3.1, 51)
			length := math.Min(15*(n-0.45), 5.8)
			if length < 0.7 {
				return 0, false
			}
			length += 1.1 * (Hash01(int(math.Floor(u/0.8)), i, face, 53) - 0.5)
			drop := top + 0.5 - p.Y
			if drop > length {
				return 0, false
			}
			// Lighter new growth at the hems and in drifts, with sprays of blossom.
			switch m := hangingGardensFBM(u*0.8+float64(face)*5.1, p.Y*0.8, 55); {
			case m > 0.71:
				return White, true
			case drop > length-1.0 || m < 0.4:
				return Grass, true
			}
			return Foliage, true
		})
	}
}

// finials sets a glazed pier with a gold cap on every corner of the terrace parapets.
func (s HangingGardens) finials(g *Grid) {
	const w = 1.8
	for i := 0; i < hangingGardensLevels; i++ {
		r := hangingGardensLevel(i)
		y := hangingGardensFloor(i+1) + hangingGardensRail
		for _, x := range []float64{r.x0, r.x1 - w} {
			for _, z := range []float64{r.z0, r.z1 - w} {
				s.fill(g, Navy, Block(V(x, y-0.45, z), V(x+w, y+1.3, z+w)))
				s.fill(g, Gold, Block(V(x, y+1.3, z), V(x+w, y+1.75, z+w)))
			}
		}
	}
}

func (s HangingGardens) Build(g *Grid) Order {
	s.ground(g)
	s.grounds(g)
	s.basins(g)
	for i := 0; i < hangingGardensLevels; i++ {
		s.storey(g, i)
	}
	s.stair(g)
	s.summit(g)
	s.cascades(g)
	s.tower(g)
	s.palace(g)
	s.gate(g)
	s.court(g)
	s.river(g)
	s.planting(g)
	s.vines(g)
	s.finials(g)
	return Sweep(P(hangingGardensCX, -7), -math.Pi/2)
}
