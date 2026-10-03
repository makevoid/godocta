package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// WildWestTown is a frontier town on the desert floor. A dusty main street runs east
// (+X) to west between two rows of wooden false-front buildings with boardwalks,
// porch posts, balconies and painted signs: livery stable and corral, general store,
// saloon, brick bank, hotel, sheriff's office and stone jail on the north side; church
// with steeple, blacksmith, barber, mercantile and boarding house on the south side.
// A gap in the south row opens onto the depot, where a steam train stands on the
// railroad beside a wooden water tower. Red sandstone mesas, a butte and a spire close
// the northern horizon (-Z); a windmill pump, wagons, horses and saguaros dot the land.
type WildWestTown struct{}

func init() { Register(WildWestTown{}) }

func (WildWestTown) Name() string  { return "wild_west_town" }
func (WildWestTown) Title() string { return "Wild West Town" }

const (
	wildWestTownTrack = 31.0 // Railroad centerline (z).
	wildWestTownLoco  = 27.0 // Locomotive pilot tip (x); the train faces +X.
)

// wildWestTownShop is a frontier building facing the street. U runs along X; V is the
// depth from the facade's outer face (negative toward the street).
type wildWestTownShop struct {
	X0, X1      float64
	Front, Dir  float64 // Facade z; Dir is the direction the building extends (±1).
	Depth, Eave float64
	Storeys     int
	FalseTop    float64 // Top of the false front; 0 shows the gable end instead.
	Step        bool    // Raised center pediment on the false front.
	Flat        bool    // Flat roof behind a parapet (bank).
	Arched      bool    // Round-headed windows.
	WideDoor    bool    // Open workshop front.
	Wall, Trim  Color
	Roof        Color
	Sign, Text  Color // Sign board and lettering; Sign == Stone means no sign.
	Porch       int   // 0 none, 1 boardwalk and awning, 2 boardwalk and balcony.
	Hitch       bool
	Seed        int
}

func (b wildWestTownShop) box(u0, u1, y0, y1, v0, v1 float64) Cuboid {
	return Block(V(u0, y0, b.Front+b.Dir*v0), V(u1, y1, b.Front+b.Dir*v1))
}

func (b wildWestTownShop) zr(v0, v1 float64) (float64, float64) {
	a, c := b.Front+b.Dir*v0, b.Front+b.Dir*v1
	return math.Min(a, c), math.Max(a, c)
}

// v converts a world z into the building's depth coordinate.
func (b wildWestTownShop) v(z float64) float64 { return (z - b.Front) * b.Dir }

func wildWestTownNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	lerp := func(a, b, t float64) float64 { return a + (b-a)*t }
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return lerp(lerp(a, b, sx), lerp(c, d, sx), sz)*2 - 1
}

func wildWestTownFBM(x, z float64, seed int) float64 {
	return 0.6*wildWestTownNoise(x, z, seed) + 0.3*wildWestTownNoise(x*2.03+17, z*2.03+5, seed+1) +
		0.1*wildWestTownNoise(x*4.1+3, z*4.1+29, seed+2)
}

// cell index of a world coordinate, for hashing.
func wildWestTownIdx(v float64) int { return int(math.Floor(v * 2)) }

func (w WildWestTown) Build(g *Grid) Order {
	w.ground(g)
	w.mesas(g)
	for _, b := range w.shops() {
		w.shop(g, b)
	}
	w.jail(g)
	w.church(g)
	w.livery(g)
	w.windmill(g, P(-57, -31))
	w.station(g)
	w.railroad(g)
	w.train(g)
	w.waterTower(g, P(37, 37.5))
	w.street(g)
	w.stagecoach(g, P(6, 3.6))
	w.props(g)
	w.cacti(g)
	return Outward(P(0, 0))
}

// ---------------------------------------------------------------------------------
// Terrain.

func (WildWestTown) inSite(x, z float64) bool {
	const hx, hz, cx, cz, r = 63, 47, -2, -5, 14
	n := 1.6 * wildWestTownNoise(x*0.08, z*0.08, 91)
	dx, dz := math.Max(math.Abs(x-cx)-(hx-r), 0), math.Max(math.Abs(z-cz)-(hz-r), 0)
	return math.Hypot(dx, dz) <= r+n
}

type wildWestTownMesa struct {
	C               Vec2
	RX, RZ          float64
	Cap, Talus, TW  float64
	Seed            int
	Power, Rotation float64
}

var wildWestTownMesas = []wildWestTownMesa{
	{P(-38, -45), 21, 7, 21, 5, 5, 1, 3, 0.05},
	{P(-37, -45.5), 24.5, 9.5, 8, 4, 4, 5, 3, 0.05}, // Bench below the big mesa's cliff.
	{P(-1, -46), 8.5, 4.2, 29, 6, 5, 2, 4, -0.12},
	{P(-1, -46), 10.5, 6.2, 8, 4, 4, 6, 3, -0.12},
	{P(13, -44.5), 1.8, 2.6, 23, 4.5, 4, 3, 2, 0},
	{P(44, -46), 14, 6, 16, 4.5, 5, 4, 3, -0.08},
}

// mesaHeight is the height of the rock at a plan point (0 off the rock).
func (WildWestTown) mesaHeight(x, z float64) float64 {
	h := 0.0
	for _, m := range wildWestTownMesas {
		d := P(x, z).Sub(m.C).Rotate(-m.Rotation)
		if math.Abs(d.X) > m.RX+m.TW+2 || math.Abs(d.Y) > m.RZ+m.TW+2 {
			continue
		}
		ang := math.Atan2(d.Y/m.RZ, d.X/m.RX)
		r := math.Pow(math.Pow(math.Abs(d.X)/m.RX, m.Power)+math.Pow(math.Abs(d.Y)/m.RZ, m.Power), 1/m.Power)
		r += 0.05*math.Sin(ang*9+float64(m.Seed)) + 0.035*math.Sin(ang*17+2*float64(m.Seed))
		r += 0.04 * wildWestTownNoise(x*0.3, z*0.3, 40+m.Seed)
		e := (1 - r) * math.Min(m.RX, m.RZ) // Distance inside the cliff edge.
		var v float64
		switch {
		case e >= 0:
			v = m.Cap + 0.5*wildWestTownNoise(x*0.2, z*0.2, 50+m.Seed)
		case e > -m.TW:
			t := 1 + e/m.TW
			v = m.Talus * math.Pow(t, 1.4) * (1 + 0.15*wildWestTownNoise(x*0.4, z*0.4, 60+m.Seed))
		}
		h = math.Max(h, v)
	}
	return h
}

func (w WildWestTown) ground(g *Grid) {
	box := Box{V(-70, -0.8, -58), V(66, -0.4, 48)}
	g.Use(Sand).Fill(Func{box, func(p Vec3) bool {
		return w.inSite(p.X, p.Z) && w.mesaHeight(p.X, p.Z) < 0.5
	}})
	// Mottled desert: red soil patches near the rocks, pale pebbles, scrub.
	g.Paint(Terracotta, Func{box, func(p Vec3) bool {
		if math.Abs(p.Z) < 8.5 && p.X > -64 && p.X < 66 {
			return false
		}
		n := wildWestTownFBM(p.X*0.07, p.Z*0.07, 7) - 0.012*(p.Z+20)
		return n > 0.3
	}})
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		return math.Abs(p.Z) > 9 && Hash01(wildWestTownIdx(p.X), wildWestTownIdx(p.Z), 8) < 0.04
	}})
}

func (w WildWestTown) mesas(g *Grid) {
	const t = 1.0
	box := Box{V(-70, -0.4, -62), V(66, 30, -30)}
	g.Use(Terracotta).Fill(Func{box, func(p Vec3) bool {
		h := w.mesaHeight(p.X, p.Z)
		if h < 0.4 || p.Y > h {
			return false
		}
		low := h
		for k := 0; k < 8; k++ {
			a := float64(k) * math.Pi / 4
			low = math.Min(low, w.mesaHeight(p.X+t*math.Cos(a), p.Z+t*math.Sin(a)))
		}
		return p.Y > low-0.8
	}})
	// Strata: thin brick bands in orange sandstone, dark cap rock, talus and scrub.
	g.Paint(Brick, Func{box, func(p Vec3) bool {
		h := w.mesaHeight(p.X, p.Z)
		if h < 0.4 {
			return false
		}
		yy := p.Y + 0.9*wildWestTownNoise(p.X*0.07, p.Z*0.07, 70)
		if h > 9 && p.Y > h-1.2 { // Cap rock rim.
			return true
		}
		if h < 8 { // Talus: mixed scree.
			return Hash01(wildWestTownIdx(p.X), wildWestTownIdx(p.Y), wildWestTownIdx(p.Z), 71) < 0.35
		}
		m := math.Mod(yy, 5.2)
		return m < 1.0 || (m > 3.0 && m < 3.4)
	}})
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		h := w.mesaHeight(p.X, p.Z)
		return h > 0.4 && h < 3.5 && Hash01(wildWestTownIdx(p.X), wildWestTownIdx(p.Y), wildWestTownIdx(p.Z), 72) < 0.4
	}})
	g.Paint(Sand, Func{box, func(p Vec3) bool { // Pale sandstone band.
		h := w.mesaHeight(p.X, p.Z)
		m := math.Mod(p.Y+0.9*wildWestTownNoise(p.X*0.07, p.Z*0.07, 70), 5.2)
		return h > 9 && p.Y < h-1.2 && p.Y > 6 && m > 4.2 && m < 4.7
	}})
	g.Paint(Terracotta, Func{box, func(p Vec3) bool { // Top surface.
		h := w.mesaHeight(p.X, p.Z)
		return h > 9 && p.Y > h-0.45
	}})
	g.Paint(Foliage, Func{box, func(p Vec3) bool {
		h := w.mesaHeight(p.X, p.Z)
		return h > 9 && p.Y > h-0.45 && Hash01(wildWestTownIdx(p.X), wildWestTownIdx(p.Z), 73) < 0.06
	}})
}

// ---------------------------------------------------------------------------------
// Buildings.

func (WildWestTown) shops() []wildWestTownShop {
	n, s := -11.0, 11.0
	return []wildWestTownShop{
		// North row, facing south (+Z).
		{X0: -33, X1: -23, Front: n, Dir: -1, Depth: 13, Eave: 5.4, Storeys: 1, FalseTop: 8.6,
			Wall: Wood, Trim: White, Roof: DarkStone, Sign: White, Text: Charcoal, Porch: 1, Hitch: true, Seed: 1},
		{X0: -20, X1: -5, Front: n, Dir: -1, Depth: 16, Eave: 9.2, Storeys: 2, FalseTop: 12, Step: true,
			Wall: Terracotta, Trim: White, Roof: DarkStone, Sign: Gold, Text: Charcoal, Porch: 2, Hitch: true, Seed: 2},
		{X0: -2, X1: 8, Front: n, Dir: -1, Depth: 13, Eave: 9.2, Storeys: 2, FalseTop: 10.6, Flat: true, Arched: true,
			Wall: Brick, Trim: Stone, Roof: Charcoal, Sign: Charcoal, Text: Gold, Seed: 3},
		{X0: 11, X1: 26, Front: n, Dir: -1, Depth: 15, Eave: 13.6, Storeys: 3, FalseTop: 15.6, Step: true,
			Wall: White, Trim: Wood, Roof: Slate, Sign: Navy, Text: White, Porch: 2, Seed: 4},
		{X0: 29, X1: 36, Front: n, Dir: -1, Depth: 12, Eave: 5.4, Storeys: 1, FalseTop: 8,
			Wall: Stone, Trim: Wood, Roof: DarkStone, Sign: White, Text: Charcoal, Porch: 1, Hitch: true, Seed: 5},
		{X0: 44, X1: 51, Front: n, Dir: -1, Depth: 10, Eave: 5, Storeys: 1, FalseTop: 7.4,
			Wall: Navy, Trim: White, Roof: DarkStone, Sign: White, Text: Navy, Porch: 1, Seed: 6},
		// South row, facing north (-Z).
		{X0: -40, X1: -31, Front: s, Dir: 1, Depth: 12, Eave: 5.4, Storeys: 1, WideDoor: true,
			Wall: Wood, Trim: Charcoal, Roof: DarkStone, Seed: 7},
		{X0: -28, X1: -22, Front: s, Dir: 1, Depth: 10, Eave: 5, Storeys: 1, FalseTop: 7.6,
			Wall: White, Trim: Red, Roof: DarkStone, Sign: Red, Text: White, Porch: 1, Seed: 8},
		{X0: -19, X1: -7, Front: s, Dir: 1, Depth: 13, Eave: 9.2, Storeys: 2, FalseTop: 11, Step: true,
			Wall: Stone, Trim: Brick, Roof: Slate, Sign: White, Text: Brick, Porch: 1, Hitch: true, Seed: 9},
		{X0: 22, X1: 34, Front: s, Dir: 1, Depth: 12, Eave: 9.2, Storeys: 2,
			Wall: Brick, Trim: White, Roof: Slate, Porch: 2, Seed: 10},
		{X0: 37, X1: 43, Front: s, Dir: 1, Depth: 10, Eave: 5, Storeys: 1, FalseTop: 7.2,
			Wall: Wood, Trim: Terracotta, Roof: DarkStone, Sign: Terracotta, Text: White, Porch: 1, Seed: 11},
	}
}

// pitchedRoof fills a hollow roof over a rectangle with its ridge along the given axis.
// top(a) is the roof height at plan distance a from the ridge line, for a ≤ half.
func (WildWestTown) pitchedRoof(g *Grid, min, max Vec2, ridge Axis, half float64, top func(a float64) float64, thick float64) {
	mid := min.Add(max).Scale(0.5)
	inside := func(a, y float64) bool { return a <= half && y <= top(a) && y >= top(half)-thick }
	g.Fill(Func{Box{min.At(top(half) - thick), max.At(top(0) + 0.1)}, func(p Vec3) bool {
		a := math.Abs(p.Z - mid.Y)
		if ridge == AxisZ {
			a = math.Abs(p.X - mid.X)
		}
		if !inside(a, p.Y) {
			return false
		}
		return !inside(a+thick, p.Y) || !inside(a, p.Y+thick) || !inside(a+0.7*thick, p.Y+0.7*thick)
	}})
}

// sign fills a board on the facade with pseudo-lettering.
func (WildWestTown) sign(g *Grid, b wildWestTownShop, u0, u1, y0, y1 float64, seed int) {
	g.With(b.Sign, func() { g.Fill(b.box(u0, u1, y0, y1, -0.75, 0.2)) })
	z0, z1 := b.zr(-0.8, -0.35)
	lx := u0 + 0.9
	g.Paint(b.Text, Func{Box{V(lx, y0+0.5, z0), V(u1-0.9, y1-0.5, z1)}, func(p Vec3) bool {
		c := int(math.Floor((p.X - lx) / 0.8))
		row := int(math.Floor((p.Y - y0 - 0.5) / 0.8))
		if c%4 == 3 {
			return false
		}
		return Hash01(seed, c/4, c%4, row) > 0.4
	}})
}

func (w WildWestTown) shop(g *Grid, b wildWestTownShop) {
	uc, hw := (b.X0+b.X1)/2, (b.X1-b.X0)/2
	D, H := b.Depth, b.Eave
	const floorH = 4.4
	top := H
	if b.Flat {
		top = b.FalseTop
	}
	// Walls.
	g.Use(b.Wall).Fill(Subtract(b.box(b.X0, b.X1, 0, top, 0, D), b.box(b.X0+1, b.X1-1, 0, top+1, 1, D-1)))
	// Roof.
	if b.Flat {
		g.Use(b.Roof).Fill(b.box(b.X0+0.5, b.X1-0.5, H-0.6, H, 0.5, D-0.5))
		g.Use(b.Trim).Fill(b.box(b.X0-0.4, b.X1+0.4, top-0.8, top, -0.6, D+0.4))
		g.Carve(b.box(b.X0+0.6, b.X1-0.6, top-0.9, top+0.5, 0.6, D-0.6))
	} else {
		half := hw + 0.7
		roofTop := func(a float64) float64 { return H + 0.5*(half-a) }
		z0, z1 := b.zr(0.6, D+0.7)
		g.Use(b.Roof)
		w.pitchedRoof(g, P(b.X0-0.7, z0), P(b.X1+0.7, z1), AxisZ, half, roofTop, 1)
		// Gable ends.
		gable := func(v0, v1 float64) Solid {
			z0, z1 := b.zr(v0, v1)
			return Func{Box{V(b.X0, H, z0), V(b.X1, roofTop(0), z1)}, func(p Vec3) bool {
				return p.Y <= roofTop(math.Abs(p.X-uc))-0.4
			}}
		}
		g.Use(b.Wall).Fill(gable(D-1, D))
		if b.FalseTop == 0 {
			g.Fill(gable(0, 1))
			g.Use(b.Trim)
			w.pitchedRoof(g, P(b.X0-0.7, math.Min(b.Front-b.Dir*0.4, b.Front+b.Dir*0.6)), P(b.X1+0.7, math.Max(b.Front-b.Dir*0.4, b.Front+b.Dir*0.6)), AxisZ, half, func(a float64) float64 { return roofTop(a) + 0.4 }, 0.8)
		}
	}
	// False front with cornice, pediment and corner boards.
	if b.FalseTop > 0 && !b.Flat {
		g.Use(b.Wall).Fill(b.box(b.X0, b.X1, H, b.FalseTop, 0, 1))
		g.Use(b.Trim).Fill(b.box(b.X0-0.5, b.X1+0.5, b.FalseTop-0.8, b.FalseTop, -0.6, 1))
		if b.Step {
			pw := hw * 0.45
			g.Use(b.Wall).Fill(b.box(uc-pw, uc+pw, b.FalseTop, b.FalseTop+1.8, 0, 1))
			g.Use(b.Trim).Fill(b.box(uc-pw-0.4, uc+pw+0.4, b.FalseTop+1.2, b.FalseTop+1.9, -0.6, 1))
		}
	}
	if b.FalseTop > 0 {
		g.Use(b.Trim).Fill(b.box(b.X0-0.3, b.X0+0.6, 0, b.FalseTop-0.6, -0.4, 0.5), b.box(b.X1-0.6, b.X1+0.3, 0, b.FalseTop-0.6, -0.4, 0.5))
	}
	// Ground floor: door and display windows.
	door := 1.1
	if b.WideDoor {
		door = hw - 1.8
	}
	g.Carve(b.box(uc-door, uc+door, 0.6, 3.6, -0.5, 1.5))
	if b.WideDoor {
		g.Use(b.Wall).Fill(b.box(uc-door-0.3, uc+door+0.3, 3.6, 4.2, -0.5, 1)) // Lintel.
	}
	if b.Wall == Terracotta { // Saloon batwing doors.
		g.Use(Wood).Fill(b.box(uc-door, uc+door, 1.4, 2.8, 0.4, 0.9))
	}
	if !b.WideDoor {
		for _, side := range []float64{-1, 1} {
			a0, a1 := uc+side*(door+1), uc+side*(hw-1.3)
			lo, hi := math.Min(a0, a1), math.Max(a0, a1)
			if hi-lo < 1.4 {
				continue
			}
			if b.Arched {
				for _, u := range []float64{(lo + hi) / 2} {
					g.Carve(Arch{Sill: V(u, 1.2, b.Front+b.Dir*0.5), HalfWidth: math.Min(1.3, (hi-lo)/2), Spring: 1.8, Depth: 3, Through: AxisZ})
					g.Use(Glass).Fill(b.box(u-1.3, u+1.3, 1.2, 4.2, 0.7, 1))
				}
				continue
			}
			g.Carve(b.box(lo, hi, 1.2, 3.6, -0.5, 1.5))
			g.Use(Glass).Fill(b.box(lo, hi, 1.2, 3.6, 0.7, 1))
		}
	}
	// Upper storeys.
	for k := 1; k < b.Storeys; k++ {
		y0 := float64(k)*floorH + 1.3
		n := int(math.Floor((2*hw - 1.5) / 3.4))
		if n < 1 {
			n = 1
		}
		for i := 0; i < n; i++ {
			u := b.X0 + 2*hw*(float64(i)+0.5)/float64(n)
			if b.Arched {
				g.Carve(Arch{Sill: V(u, y0, b.Front+b.Dir*0.5), HalfWidth: 0.85, Spring: 1.5, Depth: 3, Through: AxisZ})
			} else {
				g.Carve(b.box(u-0.85, u+0.85, y0, y0+2.4, -0.5, 1.5))
			}
			if Hash01(b.Seed, k, i) < 0.7 {
				g.Use(Glass).Fill(b.box(u-0.85, u+0.85, y0, y0+2.6, 0.7, 1))
			}
			if b.Wall == Brick { // Shutters.
				g.Use(b.Trim).Fill(b.box(u-1.6, u-0.9, y0, y0+2.4, -0.3, 0.4), b.box(u+0.9, u+1.6, y0, y0+2.4, -0.3, 0.4))
			}
		}
	}
	// Side and back windows.
	for k := 0; k < b.Storeys; k++ {
		y0 := float64(k)*floorH + 1.4
		for _, v := range []float64{D * 0.35, D * 0.75} {
			z0, z1 := b.zr(v-0.8, v+0.8)
			if Hash01(b.Seed, k, int(v), 5) < 0.75 {
				g.Carve(Block(V(b.X0-0.5, y0, z0), V(b.X1+0.5, y0+2.2, z1)))
			}
		}
		g.Carve(b.box(uc-0.8+hw*0.4, uc+0.8+hw*0.4, y0, y0+2.2, D-1.5, D+0.5))
	}
	g.Carve(b.box(uc-hw*0.4-1, uc-hw*0.4+1, 0.6, 3.4, D-1.5, D+0.5)) // Back door.
	// Porch: boardwalk, posts, awning or balcony.
	if b.Porch > 0 {
		g.Use(Wood).Fill(b.box(b.X0, b.X1, 0, 0.4, -3.2, 0))
		n := int(math.Ceil((b.X1 - b.X0 - 1) / 4.6))
		postTop := floorH
		if b.Porch == 2 {
			postTop = H - 0.3
		}
		for i := 0; i <= n; i++ {
			u := b.X0 + 0.5 + (b.X1-b.X0-1)*float64(i)/float64(n)
			g.Use(b.Trim).Fill(b.box(u-0.45, u+0.45, 0.4, postTop, -3.2, -2.3))
		}
		g.Use(b.Roof).Fill(b.box(b.X0-0.3, b.X1+0.3, floorH, floorH+0.45, -3.6, 0))
		if b.Porch == 2 {
			g.Use(Wood).Paint(Wood, b.box(b.X0-0.3, b.X1+0.3, floorH+0.4, floorH+0.45, -3.6, 0))
			g.Use(b.Trim).Fill(b.box(b.X0-0.3, b.X1+0.3, floorH+1.7, floorH+2.1, -3.6, -3))
			for u := b.X0 + 0.8; u < b.X1-0.4; u += 1.6 {
				g.Fill(b.box(u-0.25, u+0.25, floorH+0.4, floorH+1.8, -3.6, -3))
			}
			g.Use(b.Roof).Fill(b.box(b.X0-0.3, b.X1+0.3, H-0.4, H+0.1, -3.6, 0))
			// Balcony door.
			g.Carve(b.box(uc-0.9, uc+0.9, floorH+0.6, floorH+3.4, -0.5, 1.5))
		}
	}
	if b.Hitch {
		u0, u1 := b.X0+1, b.X1-1
		g.Use(Wood).Fill(b.box(u0, u0+0.8, 0, 1.6, -5, -4.2), b.box(u1-0.8, u1, 0, 1.6, -5, -4.2),
			b.box(u0, u1, 1.2, 1.6, -5, -4.2))
	}
	// Sign.
	if b.Sign != Stone {
		y0, y1 := H+0.6, b.FalseTop-1.1
		if b.Flat {
			y0, y1 = H-2.6, H-0.6
			if b.Storeys > 1 {
				y0, y1 = floorH-0.2, floorH+1.0
			}
		}
		if b.FalseTop == 0 || y1-y0 < 1.6 {
			y0, y1 = floorH+0.5, floorH+2.3
			if b.Porch > 0 {
				y0, y1 = floorH+0.5, floorH+2.1
			}
		}
		if b.Storeys == 3 {
			y0, y1 = H+0.4, b.FalseTop-1.0
		}
		w.sign(g, b, b.X0+1.2, b.X1-1.2, y0, y1, b.Seed)
	}
	// Workshop chimney, barber pole, barrels.
	if b.WideDoor {
		z0, z1 := b.zr(D-3, D-1)
		g.Use(Brick).Fill(Block(V(b.X1-3, 0, z0), V(b.X1-1, H+4.5, z1)))
		g.Use(Charcoal).Fill(b.box(uc-1, uc+1, 0, 1.2, 4, 5.4)) // Anvil and forge.
		g.Use(Brick).Fill(b.box(uc+1.8, uc+3.6, 0, 1.6, 6, 8))
	}
	if b.Trim == Red {
		for y := 0.4; y < 3.6; y += 0.8 {
			c := White
			if int(math.Round(y/0.8))%2 == 0 {
				c = Red
			}
			g.Use(c).Fill(b.box(b.X1-1.4, b.X1-0.6, y, y+0.8, -1.4, -0.6))
		}
	}
	if b.Porch > 0 && Hash01(b.Seed, 9) < 0.7 {
		u := b.X0 + 1.2
		z0, z1 := b.zr(-1.6, -0.4)
		g.Use(Wood).Fill(Cylinder{Center: P(u, (z0+z1)/2), Radius: 0.6, Y0: 0.4, Y1: 1.6})
	}
}

func (w WildWestTown) jail(g *Grid) {
	b := wildWestTownShop{X0: 36.4, X1: 42, Front: -11.8, Dir: -1, Depth: 10, Wall: DarkStone}
	g.Use(DarkStone).Fill(Subtract(b.box(b.X0, b.X1, 0, 5.4, 0, b.Depth), b.box(b.X0+1, b.X1-1, 0, 4.8, 1, b.Depth-1)))
	g.Use(Stone).Fill(b.box(b.X0-0.3, b.X1+0.3, 5.0, 5.8, -0.3, b.Depth+0.3))
	g.Carve(b.box(b.X0+0.8, b.X1-0.8, 5.4, 6, 0.5, b.Depth-0.5))
	bars := func(u0, u1, y0, y1 float64, z0, z1 float64, alongX bool) {
		if alongX {
			g.Carve(Block(V(u0, y0, z0-0.6), V(u1, y1, z1+0.6)))
			for u := u0 + 0.4; u < u1; u += 0.8 {
				g.Use(Charcoal).Fill(Block(V(u-0.2, y0, (z0+z1)/2-0.2), V(u+0.2, y1, (z0+z1)/2+0.2)))
			}
			return
		}
		g.Carve(Block(V(z0-0.6, y0, u0), V(z1+0.6, y1, u1)))
		for u := u0 + 0.4; u < u1; u += 0.8 {
			g.Use(Charcoal).Fill(Block(V((z0+z1)/2-0.2, y0, u-0.2), V((z0+z1)/2+0.2, y1, u+0.2)))
		}
	}
	bars(37.6, 40.8, 2, 3.8, -12.8, -11.8, true)
	bars(-19, -16.6, 2, 3.8, b.X1-1, b.X1, false)
	bars(-15, -12.6, 2, 3.8, b.X1-1, b.X1, false)
}

func (w WildWestTown) church(g *Grid) {
	const x0, x1, z0, z1, H = -58.0, -47.0, 15.0, 30.0, 7.0
	xc := (x0 + x1) / 2
	g.Use(White).Fill(Subtract(Block(V(x0, 0, z0), V(x1, H, z1)), Block(V(x0+1, 0, z0+1), V(x1-1, H+1, z1-1))))
	// Gable ends and 45° roof.
	half := (x1-x0)/2 + 0.7
	for _, z := range [][2]float64{{z0, z0 + 1}, {z1 - 1, z1}} {
		g.Use(White).Fill(Func{Box{V(x0, H, z[0]), V(x1, H+half, z[1])}, func(p Vec3) bool {
			return p.Y <= H+(half-math.Abs(p.X-xc))-0.6
		}})
	}
	g.Use(Slate)
	w.pitchedRoof(g, P(x0-0.7, z0-0.6), P(x1+0.7, z1+0.6), AxisZ, half, func(a float64) float64 { return H + (half - a) }, 1.1)
	// Tall arched windows.
	for z := z0 + 3.5; z < z1-1.5; z += 3.6 {
		for _, x := range []float64{x0 + 0.5, x1 - 0.5} {
			g.Carve(Arch{Sill: V(x, 1.6, z), HalfWidth: 0.8, Spring: 2.6, Depth: 2.5, Through: AxisX})
			gx := x + 0.35*math.Copysign(1, xc-x)
			g.Use(Glass).Fill(Block(V(gx-0.2, 1.6, z-0.8), V(gx+0.2, 4.9, z+0.8)))
		}
	}
	// Steeple at the street end.
	const s = 2.2
	tz := z0 - 0.6
	g.Use(White).Fill(Subtract(BlockAt(P(xc, tz), 2*s, 2*s, 0, 15), BlockAt(P(xc, tz), 2*s-2, 2*s-2, 0, 15)))
	g.Carve(Arch{Sill: V(xc, 0.6, tz-s), HalfWidth: 1, Spring: 2.4, Depth: 2, Through: AxisZ})
	g.Carve(Arch{Sill: V(xc, 7, tz-s), HalfWidth: 0.7, Spring: 1.4, Depth: 2, Through: AxisZ})
	g.Use(Glass).Fill(BlockAt(P(xc, tz-s+1.2), 1.4, 0.4, 7, 9))
	// Belfry with open arches, a bell and cornices.
	g.Use(Wood).Fill(BlockAt(P(xc, tz), 2*s+0.8, 2*s+0.8, 10.8, 11.4), BlockAt(P(xc, tz), 2*s+0.8, 2*s+0.8, 14.6, 15.2))
	g.Carve(Arch{Sill: V(xc, 11.4, tz), HalfWidth: 1.1, Spring: 1.4, Depth: 6, Through: AxisZ},
		Arch{Sill: V(xc, 11.4, tz), HalfWidth: 1.1, Spring: 1.4, Depth: 6, Through: AxisX})
	g.Use(Gold).Fill(Sphere(V(xc, 12.6, tz), 0.8))
	// Spire and cross.
	g.Use(Slate).Fill(Func{Box{V(xc-s-0.4, 15, tz-s-0.4), V(xc+s+0.4, 23.5, tz+s+0.4)}, func(p Vec3) bool {
		r := (s + 0.4) * (1 - (p.Y-15)/8.5)
		return math.Max(math.Abs(p.X-xc), math.Abs(p.Z-tz)) <= r
	}})
	g.Use(Gold).Fill(Block(V(xc-0.3, 23.4, tz-0.3), V(xc+0.3, 26.2, tz+0.3)), Block(V(xc-1.1, 24.8, tz-0.3), V(xc+1.1, 25.4, tz+0.3)))
	// Steps and a picket fence with a gate around the churchyard.
	g.Use(Wood).Fill(BlockAt(P(xc, tz-s-0.8), 3, 1.4, 0, 0.4))
	g.Use(White)
	for x := x0 - 2.5; x <= x1+2.5; x += 0.8 {
		if math.Abs(x-xc) < 2 {
			continue
		}
		g.Fill(Block(V(x-0.2, 0, z0-4.6), V(x+0.2, 1.4, z0-4.2)))
	}
	g.Fill(Block(V(x0-2.5, 0.8, z0-4.6), V(xc-2, 1.1, z0-4.2)), Block(V(xc+2, 0.8, z0-4.6), V(x1+2.5, 1.1, z0-4.2)))
	for z := z0 - 4.4; z <= z1+1.5; z += 0.8 {
		for _, x := range []float64{x0 - 2.5, x1 + 2.5} {
			g.Fill(Block(V(x-0.2, 0, z-0.2), V(x+0.2, 1.4, z+0.2)))
		}
	}
	// Graves behind the church.
	for i := 0; i < 6; i++ {
		x := x0 - 1.5 + float64(i%3)*1.6
		z := z1 - 6 + float64(i/3)*2.8
		g.Use(Stone).Fill(Block(V(x-0.4, 0, z-0.3), V(x+0.4, 1.2, z+0.3)))
	}
}

func (w WildWestTown) livery(g *Grid) {
	const x0, x1, z0, z1, H = -50.0, -36.0, -27.0, -11.0, 6.0
	xc, half := (x0+x1)/2, (x1-x0)/2+0.7
	// Gambrel profile: steep lower slopes, shallow upper slopes.
	roof := func(a float64) float64 {
		if a > 4 {
			return H + (half-a)*1.5
		}
		return H + (half-4)*1.5 + (4-a)*0.55
	}
	g.Use(Red).Fill(Subtract(Block(V(x0, 0, z0), V(x1, H, z1)), Block(V(x0+1, 0, z0+1), V(x1-1, H+1, z1-1))))
	for _, z := range [][2]float64{{z0, z0 + 1}, {z1 - 1, z1}} {
		g.Fill(Func{Box{V(x0, H, z[0]), V(x1, roof(0), z[1])}, func(p Vec3) bool { return p.Y <= roof(math.Abs(p.X-xc))-0.5 }})
	}
	g.Use(DarkStone)
	w.pitchedRoof(g, P(x0-0.7, z0-0.6), P(x1+0.7, z1+0.6), AxisZ, half, roof, 1.1)
	// White trim along the front gable and corners.
	g.Use(White)
	w.pitchedRoof(g, P(x0-0.7, z1-0.2), P(x1+0.7, z1+0.6), AxisZ, half, func(a float64) float64 { return roof(a) + 0.3 }, 0.8)
	g.Fill(Block(V(x0-0.3, 0, z1-0.6), V(x0+0.6, H, z1+0.3)), Block(V(x1-0.6, 0, z1-0.6), V(x1+0.3, H, z1+0.3)))
	// Big door, opened leaves, hayloft door and hoist beam.
	g.Carve(Block(V(xc-2.6, 0.6, z1-1.5), V(xc+2.6, 5.2, z1+0.5)))
	g.Use(Wood).Fill(Block(V(xc-6, 0.4, z1+0.1), V(xc-3, 5.2, z1+0.6)), Block(V(xc+3, 0.4, z1+0.1), V(xc+6, 5.2, z1+0.6)))
	g.Use(White).Fill(Capsule{V(xc-5.8, 0.8, z1+0.5), V(xc-3.2, 4.8, z1+0.5), 0.3}, Capsule{V(xc+3.2, 0.8, z1+0.5), V(xc+5.8, 4.8, z1+0.5), 0.3})
	g.Use(White).Fill(Block(V(xc-2.9, 5.2, z1-0.6), V(xc+2.9, 5.8, z1+0.4)))
	g.Carve(Block(V(xc-1.3, 7.2, z1-1.5), V(xc+1.3, 9.8, z1+0.5)))
	g.Use(Wood).Fill(Block(V(xc-0.4, 11, z1-2), V(xc+0.4, 11.6, z1+1.6)))
	g.Use(Sand).Fill(Block(V(xc-1, 7.2, z1-2.2), V(xc+1, 8.0, z1-1.2))) // Hay.
	// Side windows.
	for z := z0 + 3; z < z1-2; z += 4 {
		g.Carve(Block(V(x0-0.5, 2, z-0.8), V(x1+0.5, 3.6, z+0.8)))
	}
	// Sign over the door.
	b := wildWestTownShop{Front: z1, Dir: -1, Sign: White, Text: Charcoal}
	w.sign(g, b, xc-3.5, xc+3.5, 5.9, 7.0, 31)
	// Corral to the west: post-and-rail fence, trough, horses.
	const cx0, cx1, cz0, cz1 = -62.0, -50.6, -26.0, -11.0
	g.Use(Wood)
	post := func(x, z float64) { g.Fill(Block(V(x-0.4, 0, z-0.4), V(x+0.4, 2.4, z+0.4))) }
	for x := cx0; x <= cx1; x += 2.85 {
		post(x, cz0)
		post(x, cz1)
	}
	for z := cz0; z <= cz1; z += 3 {
		post(cx0, z)
	}
	for _, y := range []float64{1.0, 2.0} {
		g.Fill(Block(V(cx0, y, cz0-0.25), V(cx1, y+0.4, cz0+0.25)), Block(V(cx0, y, cz1-0.25), V(cx1, y+0.4, cz1+0.25)),
			Block(V(cx0-0.25, y, cz0), V(cx0+0.25, y+0.4, cz1)))
	}
	g.Carve(Block(V(-57.5, 0.6, cz1-1), V(-54.5, 3, cz1+1))) // Gate.
	g.Use(Wood).Fill(Block(V(-59, 0, -20), V(-58, 1.2, -15)))
	g.Use(Water).Fill(Block(V(-58.8, 0.8, -19.6), V(-58.2, 1.2, -15.4)))
	w.horse(g, P(-55, -16), 0, Wood)
	w.horse(g, P(-55.5, -22), math.Pi*0.8, Charcoal)
	w.horse(g, P(-60, -23.5), -math.Pi/2, White)
}

// horse is a simple horse standing at c, facing angle (0 = +X).
func (WildWestTown) horse(g *Grid, c Vec2, angle float64, coat Color) {
	at := func(a, y, b float64) Vec3 { return P(a, b).Rotate(angle).Add(c).At(y) }
	parts := []Solid{
		Capsule{at(-0.9, 2.0, 0), at(0.9, 2.1, 0), 0.75},
		Capsule{at(1.1, 2.3, 0), at(1.8, 3.3, 0), 0.45},
		Capsule{at(1.9, 3.4, 0), at(2.6, 2.9, 0), 0.4},
	}
	for _, a := range []float64{-1.0, 1.0} {
		for _, b := range []float64{-0.35, 0.35} {
			parts = append(parts, Capsule{at(a, 0, b), at(a, 1.7, b), 0.38})
		}
	}
	g.Use(coat).Fill(parts...)
	mane := Charcoal
	if coat == Charcoal {
		mane = Wood
	}
	g.Use(mane).Fill(Capsule{at(-1.5, 2.3, 0), at(-1.9, 1.2, 0), 0.32}, Capsule{at(1.0, 2.9, 0), at(1.7, 3.8, 0), 0.3})
}

func (w WildWestTown) windmill(g *Grid, c Vec2) {
	const top = 11.0
	g.Use(Wood)
	for _, d := range []Vec2{P(-1, -1), P(1, -1), P(1, 1), P(-1, 1)} {
		g.Fill(Capsule{c.Add(d.Scale(2)).At(0), c.Add(d.Scale(0.6)).At(top), 0.45})
	}
	for _, y := range []float64{3.5, 7} {
		r := 2 - 1.4*y/top
		for _, s := range [][2]Vec2{{P(-1, -1), P(1, -1)}, {P(1, -1), P(1, 1)}, {P(1, 1), P(-1, 1)}, {P(-1, 1), P(-1, -1)}} {
			g.Fill(Capsule{c.Add(s[0].Scale(r)).At(y), c.Add(s[1].Scale(r)).At(y), 0.3})
		}
	}
	g.Fill(BlockAt(c, 2, 2, top, top+0.5))
	// Rotor facing south (+Z), slatted blades in a ring, and a tail vane.
	hub := c.At(top + 2.2).Add(V(0, 0, 0.9))
	g.Use(Charcoal).Fill(Capsule{hub.Add(V(0, 0, 0.3)), hub.Add(V(0, 0, -2)), 0.5}, Block(V(c.X-0.4, top, c.Y-0.4), V(c.X+0.4, hub.Y, c.Y+0.4)))
	g.Use(White).Fill(Func{Box{hub.Sub(V(3.4, 3.4, 0.35)), hub.Add(V(3.4, 3.4, 0.35))}, func(p Vec3) bool {
		d := P(p.X-hub.X, p.Y-hub.Y)
		r := d.Length()
		if r > 3.3 || r < 0.8 {
			return false
		}
		f := math.Mod((d.Angle()+math.Pi)*12/(2*math.Pi), 1)
		return f < 0.55 || r > 2.95
	}})
	g.Use(Red).Fill(Block(V(c.X-0.25, hub.Y-0.6, hub.Z-2.2), V(c.X+0.25, hub.Y+1.4, hub.Z-5)))
	// Stock tank.
	tank := c.Add(P(5.5, 1))
	g.Use(DarkStone).Fill(Cylinder{Center: tank, Radius: 2.2, Inner: 1.5, Y0: 0, Y1: 1.4})
	g.Use(Water).Fill(Cylinder{Center: tank, Radius: 1.6, Y0: 0, Y1: 1.0})
	g.Use(Charcoal).Fill(Capsule{c.At(0.6), tank.At(0.6), 0.3})
}

func (w WildWestTown) station(g *Grid) {
	const x0, x1, z0, z1, H = 1.0, 17.0, 20.0, 26.0, 5.6
	xc := (x0 + x1) / 2
	g.Use(Terracotta).Fill(Subtract(Block(V(x0, 0, z0), V(x1, H, z1)), Block(V(x0+1, 0, z0+1), V(x1-1, H+1, z1-1))))
	// Bay window toward the tracks for the operator.
	g.Fill(BlockAt(P(xc+3, z1+0.6), 3.6, 1.6, 0, H))
	g.Carve(BlockAt(P(xc+3, z1+1.4), 2.4, 1, 1.4, 3.6))
	g.Use(Glass).Fill(BlockAt(P(xc+3, z1+0.9), 2.4, 0.4, 1.4, 3.6))
	// Doors and windows on both long sides.
	for _, z := range []float64{z0, z1} {
		for x := x0 + 2.2; x < x1-1; x += 3.2 {
			if math.Abs(x-(xc+3)) < 2 && z == z1 {
				continue
			}
			if math.Abs(x-xc+0.6) < 1.7 {
				g.Carve(Block(V(x-1, 0.6, z-1.5), V(x+1, 3.8, z+1.5)))
				continue
			}
			g.Carve(Block(V(x-0.8, 1.4, z-1.5), V(x+0.8, 3.6, z+1.5)))
			gz := z + 0.35*math.Copysign(1, (z0+z1)/2-z)
			g.Use(Glass).Fill(Block(V(x-0.8, 1.4, gz-0.2), V(x+0.8, 3.6, gz+0.2)))
		}
	}
	// White trim band and wide hip roof on brackets, overhanging the platform.
	g.Use(White).Paint(White, Block(V(x0-1, H-0.8, z0-1), V(x1+1, H, z1+2)))
	g.Use(Slate)
	roofMin, roofMax := P(x0-2.5, z0-2.5), P(x1+2.5, z1+3.5)
	g.Fill(Func{Box{roofMin.At(H), roofMax.At(H + 4)}, func(p Vec3) bool {
		in := math.Min(math.Min(p.X-roofMin.X, roofMax.X-p.X), math.Min(p.Z-roofMin.Y, roofMax.Y-p.Z))
		t := H + 0.6*in
		return in >= 0 && p.Y <= t && p.Y > t-1.1 && p.Y <= H+3.2
	}})
	g.Fill(Block(V(x0+3.5, H+2.4, z0+2.5), V(x1-3.5, H+3.2, z1-1.5)))
	g.Use(White)
	for x := x0; x <= x1+0.1; x += 4 {
		for _, z := range []float64{z0 - 0.2, z1 + 0.2} {
			dz := -1.6
			if z > z0 {
				dz = 2.8
			}
			g.Fill(Capsule{V(x, H-1.6, z), V(x, H+0.2, z+dz), 0.35})
		}
	}
	// Name boards on the gable ends.
	for _, x := range []float64{x0 - 0.3, x1 + 0.3} {
		g.Use(White).Fill(Block(V(x-0.3, H-2.4, z0+1.5), V(x+0.3, H-1.0, z1-1.5)))
	}
	g.Use(Charcoal).Paint(Charcoal, Func{Box{V(x0-0.7, H-2, z0+2), V(x1+0.7, H-1.4, z1-2)}, func(p Vec3) bool {
		return Hash01(wildWestTownIdx(p.Z), 77) < 0.6 && (p.X < x0 || p.X > x1)
	}})
	// Platform with a freight dock, baggage carts and crates.
	g.Use(Wood).Fill(Block(V(-8, 0, z1-6.5), V(25, 0.8, wildWestTownTrack-2.2)))
	g.Use(Wood).Fill(Block(V(18, 0.8, z0), V(24, 1.6, z1)))
	for i, p := range []Vec3{V(19, 1.6, 21), V(20.4, 1.6, 21.2), V(19.6, 2.8, 21.1), V(22.5, 1.6, 24.5), V(3, 0.8, 27.5)} {
		c := Wood
		if i%2 == 1 {
			c = Terracotta
		}
		g.Use(c).Fill(Block(p.Sub(V(0.6, 0, 0.6)), p.Add(V(0.6, 1.2, 0.6))))
	}
	g.Use(Charcoal).Fill(Cylinder{Center: P(-4, 27.6), Radius: 0.6, Y0: 0.8, Y1: 2.2})
	// Path from the street to the depot.
	g.Paint(Sand, Block(V(-4, -0.4, 8), V(21, -0.4, z0)))
}

func (w WildWestTown) railroad(g *Grid) {
	const tz = wildWestTownTrack
	const x0, x1 = -64.0, 62.0
	ok := func(x float64) bool { return w.inSite(x, tz-3) && w.inSite(x, tz+3) }
	g.Use(DarkStone).Fill(Func{Box{V(x0-4, -0.4, tz-3.2), V(x1+4, 0.4, tz+3.2)}, func(p Vec3) bool {
		if !ok(p.X) {
			return false
		}
		return math.Abs(p.Z-tz) <= 3.2-1.6*p.Y
	}})
	g.Use(Wood).Fill(Func{Box{V(x0-4, 0.4, tz-2.0), V(x1+4, 0.8, tz+2.0)}, func(p Vec3) bool {
		return ok(p.X) && math.Mod(p.X+200, 2.0) < 0.9
	}})
	g.Use(Charcoal).Fill(Func{Box{V(x0-4, 0.8, tz-1.2), V(x1+4, 1.2, tz+1.2)}, func(p Vec3) bool {
		return ok(p.X) && math.Abs(math.Abs(p.Z-tz)-1.0) < 0.25
	}})
	// Telegraph poles beside the line.
	for x := -58.0; x < 60; x += 13 {
		if !w.inSite(x, tz+5.5) || (x > 30 && x < 44) {
			continue
		}
		g.Use(Wood).Fill(Block(V(x-0.4, 0, tz+5.1), V(x+0.4, 8, tz+5.9)), Block(V(x-0.4, 6.8, tz+3.9), V(x+0.4, 7.3, tz+7.1)))
		g.Use(Glass).Fill(Block(V(x-0.3, 7.3, tz+4.0), V(x+0.3, 7.7, tz+4.4)), Block(V(x-0.3, 7.3, tz+6.6), V(x+0.3, 7.7, tz+7.0)))
	}
}

// xCyl is a horizontal cylinder along X.
func (WildWestTown) xCyl(y, z, r, x0, x1 float64) Solid {
	return Func{Box{V(x0, y-r, z-r), V(x1, y+r, z+r)}, func(p Vec3) bool {
		return (p.Y-y)*(p.Y-y)+(p.Z-z)*(p.Z-z) <= r*r
	}}
}

// wheel is a vertical disk in the XY plane.
func (WildWestTown) wheel(x, y, z, r, t float64) Solid {
	return Func{Box{V(x-r, y-r, z-t/2), V(x+r, y+r, z+t/2)}, func(p Vec3) bool {
		return (p.X-x)*(p.X-x)+(p.Y-y)*(p.Y-y) <= r*r
	}}
}

func (w WildWestTown) train(g *Grid) {
	const tz, lf = wildWestTownTrack, wildWestTownLoco
	// Locomotive: frame, boiler, smokebox, stack, domes, headlamp, cab, pilot.
	g.Use(Charcoal).Fill(Block(V(lf-14, 2.2, tz-1.5), V(lf-1.2, 2.8, tz+1.5)))
	g.Use(Charcoal).Fill(w.xCyl(4.2, tz, 1.55, lf-11, lf-2.2))
	g.Use(DarkStone).Fill(w.xCyl(4.2, tz, 1.7, lf-3.2, lf-1.2))
	g.Use(Gold).Paint(Gold, w.xCyl(4.2, tz, 1.8, lf-9.2, lf-8.8), w.xCyl(4.2, tz, 1.8, lf-5.6, lf-5.2))
	g.Use(Charcoal).Fill(Frustum{Center: P(lf-2.4, tz), R0: 0.6, R1: 1.4, Y0: 5.4, Y1: 8.6})
	g.Use(Charcoal).Fill(Cylinder{Center: P(lf-2.4, tz), Radius: 1.5, Y0: 8.2, Y1: 8.8})
	g.Use(Gold).Fill(Cylinder{Center: P(lf-7, tz), Radius: 0.8, Y0: 5.2, Y1: 6.4}, Sphere(V(lf-7, 6.4, tz), 0.8),
		Cylinder{Center: P(lf-4.8, tz), Radius: 0.6, Y0: 5.2, Y1: 6.0})
	g.Use(Gold).Fill(Block(V(lf-2.2, 5.6, tz-0.7), V(lf-0.8, 6.8, tz+0.7)))
	g.Use(White).Fill(Block(V(lf-0.9, 5.8, tz-0.4), V(lf-0.6, 6.6, tz+0.4)))
	g.Use(Red).Fill(Func{Box{V(lf-1.4, 1.0, tz-1.6), V(lf+1.2, 2.6, tz+1.6)}, func(p Vec3) bool {
		t := (p.X - (lf - 1.4)) / 2.6
		return math.Abs(p.Z-tz) <= 1.6*(1-t) && p.Y <= 1.0+1.6*(1-t)+0.4
	}})
	g.Use(Red).Fill(Block(V(lf-1.6, 2.2, tz-1.7), V(lf-1.0, 3.0, tz+1.7))) // Buffer beam.
	// Cab.
	g.Use(Red).Fill(Subtract(Block(V(lf-14.4, 2.8, tz-1.9), V(lf-10.8, 7, tz+1.9)), Block(V(lf-13.6, 2.8, tz-1.1), V(lf-11.6, 7, tz+1.1))))
	g.Carve(Block(V(lf-13.6, 4.4, tz-2.5), V(lf-11.6, 6.2, tz+2.5)))
	g.Use(Charcoal).Fill(Func{Box{V(lf-15, 7, tz-2.4), V(lf-10.2, 7.9, tz+2.4)}, func(p Vec3) bool {
		return p.Y <= 7.9-0.15*math.Abs(p.Z-tz)
	}})
	// Wheels: two driver pairs, a leading truck and tender trucks.
	for _, side := range []float64{-1.35, 1.35} {
		z := tz + side
		g.Use(Red).Fill(w.wheel(lf-9.6, 2.5, z, 1.5, 0.5), w.wheel(lf-6.2, 2.5, z, 1.5, 0.5))
		g.Use(Charcoal).Fill(w.wheel(lf-3.6, 1.9, z, 0.85, 0.5), w.wheel(lf-1.8, 1.9, z, 0.85, 0.5))
		g.Use(Charcoal).Fill(Block(V(lf-9.8, 2.2, z+math.Copysign(0.3, side)), V(lf-6, 2.7, z+math.Copysign(0.6, side)))) // Side rod.
	}
	// Tender with firewood.
	g.Use(Charcoal).Fill(Block(V(lf-21, 2.0, tz-1.8), V(lf-15, 4.6, tz+1.8)))
	g.Use(Wood).Fill(Func{Box{V(lf-20.6, 4.6, tz-1.5), V(lf-17, 5.8, tz+1.5)}, func(p Vec3) bool {
		return p.Y < 5.0 || Hash01(wildWestTownIdx(p.X), wildWestTownIdx(p.Z), 80) < 0.7
	}})
	// Passenger coach and caboose.
	coach := func(x0, x1 float64, body, roof Color, cupola bool) {
		g.Use(body).Fill(Subtract(Block(V(x0, 2.0, tz-1.9), V(x1, 5.6, tz+1.9)), Block(V(x0+0.8, 2.0, tz-1.1), V(x1-0.8, 5.6, tz+1.1))))
		for x := x0 + 1.6; x < x1-1.2; x += 1.6 {
			g.Carve(Block(V(x-0.5, 3.6, tz-2.5), V(x+0.5, 4.8, tz+2.5)))
		}
		g.Use(Gold).Paint(Gold, Block(V(x0, 2.8, tz-2.1), V(x1, 3.1, tz+2.1)))
		g.Use(roof).Fill(Block(V(x0-0.4, 5.6, tz-2.2), V(x1+0.4, 6.1, tz+2.2)))
		if cupola {
			c := (x0 + x1) / 2
			g.Use(body).Fill(Block(V(c-1.4, 6.1, tz-1.2), V(c+1.4, 7.6, tz+1.2)))
			g.Carve(Block(V(c-0.8, 6.6, tz-2), V(c+0.8, 7.2, tz+2)))
			g.Use(roof).Fill(Block(V(c-1.8, 7.6, tz-1.6), V(c+1.8, 8.0, tz+1.6)))
		} else {
			g.Use(roof).Fill(Block(V(x0+1, 6.1, tz-1.0), V(x1-1, 6.8, tz+1.0)))
		}
		g.Use(Charcoal).Fill(Block(V(x0-1, 2.0, tz-1.4), V(x0, 2.4, tz+1.4)), Block(V(x1, 2.0, tz-1.4), V(x1+1, 2.4, tz+1.4)))
		for _, xx := range []float64{x0 + 2, x0 + 3.6, x1 - 3.6, x1 - 2} {
			for _, side := range []float64{-1.35, 1.35} {
				g.Fill(w.wheel(xx, 1.8, tz+side, 0.8, 0.5))
			}
		}
	}
	coach(lf-38.5, lf-22.5, Wood, Charcoal, false)
	coach(lf-50.5, lf-40.5, Red, Charcoal, true)
	// Steam drifting back over the train.
	g.Use(White)
	for _, p := range []struct {
		c Vec3
		r float64
	}{{V(lf-2.6, 10, tz), 1.3}, {V(lf-5, 11.6, tz+0.4), 1.7}, {V(lf-8.6, 12.8, tz+0.9), 2.1}, {V(lf-13, 13.6, tz+1.4), 2.4}, {V(lf-18, 14, tz+2), 2.2}} {
		g.FillVolume(Ellipsoid{p.c, V(p.r*1.25, p.r*0.85, p.r)})
	}
}

func (w WildWestTown) waterTower(g *Grid, c Vec2) {
	const legTop, tankTop, R = 9.5, 15.5, 3.6
	g.Use(Wood)
	corners := []Vec2{P(-2.6, -2.6), P(2.6, -2.6), P(2.6, 2.6), P(-2.6, 2.6)}
	for _, d := range corners {
		g.Fill(Capsule{c.Add(d.Scale(1.15)).At(0), c.Add(d).At(legTop), 0.55})
	}
	for i := range corners {
		a, b := c.Add(corners[i].Scale(1.1)), c.Add(corners[(i+1)%4].Scale(1.1))
		g.Fill(Capsule{a.At(1), b.At(legTop - 1), 0.35}, Capsule{b.At(1), a.At(legTop - 1), 0.35})
	}
	g.Fill(Cylinder{Center: c, Radius: R + 0.4, Y0: legTop - 0.5, Y1: legTop})
	g.Fill(Cylinder{Center: c, Radius: R, Inner: R - 1, Y0: legTop, Y1: tankTop})
	g.Use(Charcoal).Paint(Charcoal, Cylinder{Center: c, Radius: R + 0.3, Inner: R - 0.5, Y0: 11.2, Y1: 11.5},
		Cylinder{Center: c, Radius: R + 0.3, Inner: R - 0.5, Y0: 13.6, Y1: 13.9})
	g.Use(DarkStone).Fill(Subtract(Cone(c, R+0.7, tankTop, 3.2), Translate(Cone(c, R+0.7, tankTop, 3.2), V(0, -1.1, 0))))
	g.Use(Wood).Fill(Sphere(c.At(tankTop+3.1), 0.5))
	// Spout swung toward the track.
	g.Use(Charcoal).Fill(Capsule{c.Add(P(0, -R)).At(11), V(c.X, 9.4, wildWestTownTrack+0.6), 0.5})
	g.Use(Charcoal).Fill(Capsule{V(c.X, 11.5, c.Y-R), V(c.X, 13.5, c.Y-R-1), 0.25})
}

// street paints the main street, ruts, and parks wagons and horses.
func (w WildWestTown) street(g *Grid) {
	g.Paint(Sand, Block(V(-66, -0.4, -8.5), V(66, -0.4, 8.5)))
	g.Paint(Stone, Func{Box{V(-66, -0.4, -8), V(66, -0.4, 8)}, func(p Vec3) bool {
		return math.Abs(math.Abs(p.Z-0.4)-1.6) < 0.3 && Hash01(wildWestTownIdx(p.X), 83) < 0.8
	}})
	w.wagon(g, P(-12, 2.2), 1, true)
	w.wagon(g, P(30, -3), -1, false)
	w.wagon(g, P(56, 4), 1, true)
	for _, h := range []struct {
		c Vec2
		a float64
		k Color
	}{{P(-17, -5.8), math.Pi / 2, Wood}, {P(-13.5, -5.8), math.Pi / 2, Charcoal}, {P(-28, -5.6), math.Pi / 2, White},
		{P(-12, 5.8), -math.Pi / 2, Wood}, {P(33, -5.5), math.Pi / 2, Wood}} {
		w.horse(g, h.c, h.a, h.k)
	}
}

// wagon is a covered wagon (or buckboard) heading along ±X with its team.
func (w WildWestTown) wagon(g *Grid, c Vec2, dir float64, covered bool) {
	x, z := c.X, c.Y
	g.Use(Wood).Fill(Block(V(x-2.6, 1.4, z-1.1), V(x+2.6, 2.4, z+1.1)))
	g.Carve(Block(V(x-2.0, 1.9, z-0.5), V(x+2.0, 2.5, z+0.5)))
	for _, wx := range []float64{x - 1.8*dir, x + 1.8*dir} {
		r := 1.2
		if wx == x+1.8*dir {
			r = 0.9
		}
		for _, side := range []float64{-1.4, 1.4} {
			g.Use(Charcoal).Fill(w.wheel(wx, r, z+side, r, 0.45))
			g.Use(Wood).Fill(w.wheel(wx, r, z+side, r-0.45, 0.6))
		}
	}
	if covered {
		g.Use(White).Fill(Func{Box{V(x-2.4, 2.2, z-1.8), V(x+2.4, 4.3, z+1.8)}, func(p Vec3) bool {
			d := math.Hypot(p.Y-2.2, (p.Z-z)*1.1)
			return d <= 2.0 && d >= 1.2
		}})
	} else {
		g.Use(Wood).Fill(Block(V(x+1.0*dir-0.4, 2.4, z-1.1), V(x+1.0*dir+0.4, 3.4, z+1.1)))
		g.Use(Sand).Fill(Block(V(x-2.0, 2.0, z-0.6), V(x-0.4, 3.0, z+0.6)))
	}
	g.Use(Wood).Fill(Capsule{V(x+2.6*dir, 1.6, z), V(x+5.4*dir, 1.9, z), 0.3})
	a := 0.0
	if dir < 0 {
		a = math.Pi
	}
	w.horse(g, P(x+4.6*dir, z-0.85), a, Wood)
	if covered {
		w.horse(g, P(x+4.6*dir, z+0.85), a, Charcoal)
	}
}

// cacti scatters saguaros, barrel cacti and boulders over the open desert.
func (w WildWestTown) cacti(g *Grid) {
	free := func(x, z float64) bool {
		if !w.inSite(x+2, z) || !w.inSite(x-2, z) || !w.inSite(x, z+2) || !w.inSite(x, z-2) || w.mesaHeight(x, z) > 0.2 {
			return false
		}
		if z > -30 && z < 28 && x > -64 && x < 53 { // Town.
			return false
		}
		if math.Abs(z-wildWestTownTrack) < 7 { // Railroad.
			return false
		}
		if z > -36 && z < -24 && x < -48 { // Windmill.
			return false
		}
		if x > 30 && x < 45 && z > 30 { // Water tower.
			return false
		}
		return math.Abs(z) > 10 || x > 60
	}
	for gx := -66.0; gx <= 66; gx += 4.5 {
		for gz := -56.0; gz <= 48; gz += 4.5 {
			i, j := int(gx), int(gz)
			x := gx + (Hash01(i, j, 1)-0.5)*3
			z := gz + (Hash01(i, j, 2)-0.5)*3
			if !free(x, z) {
				continue
			}
			r := Hash01(i, j, 3)
			switch {
			case r < 0.22:
				w.saguaro(g, P(x, z), 4.5+3*Hash01(i, j, 4), i*31+j)
			case r < 0.36:
				g.Use(Foliage).FillVolume(Ellipsoid{V(x, 0.3, z), V(0.9, 1.0, 0.9)})
			case r < 0.46:
				c := Terracotta
				if Hash01(i, j, 5) < 0.5 {
					c = Brick
				}
				g.Use(c).FillVolume(Ellipsoid{V(x, 0, z), V(1.2+Hash01(i, j, 6), 0.9+0.6*Hash01(i, j, 7), 1.1)})
			case r < 0.58:
				g.Use(Grass).FillVolume(Ellipsoid{V(x, 0.1, z), V(0.8, 0.6, 0.8)})
			}
		}
	}
	// A few landmark saguaros by the street ends and the depot.
	for i, p := range []Vec2{P(56, -14), P(58, 16), P(-63, 4), P(-4, 43), P(22, 44), P(50, 22)} {
		if free(p.X, p.Y) || math.Abs(p.Y) > 10 {
			w.saguaro(g, p, 7+float64(i%3), 900+i)
		}
	}
}

func (WildWestTown) saguaro(g *Grid, c Vec2, h float64, seed int) {
	g.Use(Foliage).Fill(Capsule{c.At(0), c.At(h), 0.7})
	arms := 1 + int(Hash01(seed, 1)*2.5)
	for k := 0; k < arms; k++ {
		a := Hash01(seed, k, 2)*2*math.Pi + float64(k)*math.Pi
		d := P(math.Cos(a), math.Sin(a))
		y := h * (0.35 + 0.25*Hash01(seed, k, 3))
		out := c.Add(d.Scale(1.7))
		g.Fill(Capsule{c.At(y), out.At(y + 0.3), 0.55}, Capsule{out.At(y + 0.3), out.At(y + 1.8 + 1.4*Hash01(seed, k, 4)), 0.55})
	}
}

// stagecoach is a red Concord coach heading west (-X) with a four-horse team.
func (w WildWestTown) stagecoach(g *Grid, c Vec2) {
	x, z := c.X, c.Y
	g.Use(Red).Fill(Func{Box{V(x-1.9, 1.8, z-1.1), V(x+1.9, 4.4, z+1.1)}, func(p Vec3) bool {
		bulge := 0.35 * (1 - math.Pow((p.Y-3.1)/1.3, 2))
		return math.Abs(p.X-x) <= 1.6+bulge
	}})
	g.Carve(Block(V(x-0.9, 3.0, z-1.6), V(x+0.9, 4.0, z+1.6)))
	g.Use(Gold).Paint(Gold, Block(V(x-2.1, 2.4, z-1.2), V(x+2.1, 2.6, z+1.2)))
	g.Use(Charcoal).Fill(Block(V(x-1.8, 4.4, z-1.2), V(x+1.8, 4.8, z+1.2)))
	g.Use(Wood).Fill(Block(V(x-1.4, 4.8, z-0.8), V(x+0.6, 5.6, z+0.8)))
	g.Use(Red).Fill(Block(V(x-3.2, 3.0, z-1.0), V(x-1.9, 3.6, z+1.0)), Block(V(x-2.8, 3.6, z-1.0), V(x-1.9, 4.4, z+1.0)))
	g.Use(Wood).Fill(Block(V(x+1.9, 2.4, z-0.9), V(x+2.8, 3.4, z+0.9)))
	for _, wh := range []struct{ x, r float64 }{{x + 1.6, 1.3}, {x - 1.8, 1.0}} {
		for _, side := range []float64{-1.4, 1.4} {
			g.Use(Charcoal).Fill(w.wheel(wh.x, wh.r, z+side, wh.r, 0.45))
			g.Use(Gold).Fill(w.wheel(wh.x, wh.r, z+side, 0.45, 0.6))
		}
	}
	g.Use(Wood).Fill(Capsule{V(x-2.8, 1.6, z), V(x-10, 1.8, z), 0.3})
	for i, hx := range []float64{x - 5, x - 8.6} {
		k := Wood
		if i == 1 {
			k = Charcoal
		}
		w.horse(g, P(hx, z-0.85), math.Pi, k)
		w.horse(g, P(hx, z+0.85), math.Pi, White)
	}
}

// props adds a well, troughs, outhouses and woodpiles around town.
func (w WildWestTown) props(g *Grid) {
	well := P(-4, 0)
	g.Use(Stone).Fill(Cylinder{Center: well, Radius: 1.6, Inner: 0.8, Y0: 0, Y1: 1.4})
	g.Use(Water).Fill(Cylinder{Center: well, Radius: 0.9, Y0: 0, Y1: 0.4})
	g.Use(Wood).Fill(Block(V(well.X-1.6, 0, well.Y-0.3), V(well.X-1.0, 4, well.Y+0.3)), Block(V(well.X+1.0, 0, well.Y-0.3), V(well.X+1.6, 4, well.Y+0.3)),
		Block(V(well.X-1.6, 3.2, well.Y-0.25), V(well.X+1.6, 3.6, well.Y+0.25)))
	g.Use(DarkStone)
	w.pitchedRoof(g, P(well.X-2.2, well.Y-1.8), P(well.X+2.2, well.Y+1.8), AxisX, 1.8, func(a float64) float64 { return 4 + (1.8 - a) }, 0.8)
	g.Use(Wood).Fill(Block(V(well.X+1.8, 0, well.Y-0.6), V(well.X+4.6, 1.0, well.Y+0.6)))
	g.Use(Water).Fill(Block(V(well.X+2.1, 0.6, well.Y-0.25), V(well.X+4.3, 1.0, well.Y+0.25)))
	// Outhouses and woodpiles behind the rows.
	for i, p := range []Vec2{P(-27, -28.5), P(-11, -31), P(20, -30), P(47, -26), P(-24, 25.5), P(-13, 28), P(28, 27)} {
		g.Use(Wood).Fill(Subtract(BlockAt(p, 1.8, 1.8, 0, 3), BlockAt(p, 0.6, 0.6, 0, 3)))
		dz := -0.9
		if p.Y < 0 {
			dz = 0.9
		}
		g.Carve(BlockAt(P(p.X, p.Y+dz), 0.8, 0.6, 0.6, 2.4))
		g.Use(DarkStone).Fill(BlockAt(p, 2.4, 2.4, 3, 3.5))
		if i%2 == 0 {
			g.Use(Wood).Fill(Func{Box{V(p.X+2, 0, p.Y-1.5), V(p.X+4.4, 1.6, p.Y+1.5)}, func(q Vec3) bool {
				return q.Y < 0.8 || Hash01(wildWestTownIdx(q.X), wildWestTownIdx(q.Z), i) < 0.6
			}})
		}
	}
}
