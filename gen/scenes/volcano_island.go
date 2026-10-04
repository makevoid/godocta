package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// VolcanoIsland is a compact tropical island dominated by a smoking stratovolcano. The
// hollow cone has dark ash-banded upper slopes cut by radial gullies, a summit crater
// holding a lava lake, and a broad smoke column drifting east-north-east. A lava channel
// with black crusted levees spills through a notch in the eastern rim and runs down the
// flank into the sea, where it builds a small delta wrapped in steam; a red cinder cone
// sits on the south-eastern flank. Jungle, rainforest giants and palm groves clothe the
// lower slopes; a thatched fishing village with a jetty, an overwater hut and outrigger
// canoes sits on the south-western beach, a lighthouse crowns the south-eastern headland,
// and a broken fringing reef rings the shallow sea plate. North is -Z.
type VolcanoIsland struct{}

func init() { Register(VolcanoIsland{}) }

func (VolcanoIsland) Name() string  { return "volcano_island" }
func (VolcanoIsland) Title() string { return "Volcano Island" }

// Key dimensions.
var (
	volcanoIslandPeak     = P(2, -4)   // Cone axis.
	volcanoIslandVillage  = P(-21, 17) // Village lowland center.
	volcanoIslandHeadland = P(26, 28)  // Lighthouse headland center.
	volcanoIslandSea      = P(-1, 1)   // Sea plate center.
	volcanoIslandCinder   = P(25, 11)  // Parasitic cinder cone.
	// Lava channel from the rim notch down the eastern flank into the sea.
	volcanoIslandLava = []Vec2{P(8.5, -6.2), P(12, -6.5), P(16, -5), P(20, -6.6), P(24, -5.4),
		P(28, -7.2), P(32, -6.4), P(36, -8), P(40, -7.6), P(43.5, -8.8), P(47, -9.6)}
)

const (
	volcanoIslandRim      = 35.0 // Crater rim height.
	volcanoIslandCrater   = 8.0  // Crater rim radius.
	volcanoIslandFloor    = 30.0 // Lava lake level.
	volcanoIslandBase     = 42.0 // Cone radius where it meets the sea.
	volcanoIslandSeaA     = 59.0 // Sea plate semi-axes.
	volcanoIslandSeaB     = 55.0
	volcanoIslandVillageY = 2.5
	volcanoIslandHeadY    = 7.0
)

// volcanoIslandLand memoizes terrain heights per sample point.
type volcanoIslandLand struct {
	cache map[[2]int64]float64
}

func volcanoIslandSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func volcanoIslandLerp(a, b, t float64) float64 { return a + (b-a)*t }

// volcanoIslandNoise is smooth value noise in [-1, 1].
func volcanoIslandNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return volcanoIslandLerp(volcanoIslandLerp(a, b, sx), volcanoIslandLerp(c, d, sx), sz)*2 - 1
}

func volcanoIslandFBM(x, z float64, seed int) float64 {
	return 0.6*volcanoIslandNoise(x, z, seed) + 0.3*volcanoIslandNoise(x*2.03+17, z*2.03+5, seed+1) +
		0.1*volcanoIslandNoise(x*4.1+3, z*4.1+29, seed+2)
}

// inSea is the elliptical sea plate that holds the island.
func (VolcanoIsland) inSea(x, z float64) bool {
	dx, dz := (x-volcanoIslandSea.X)/volcanoIslandSeaA, (z-volcanoIslandSea.Y)/volcanoIslandSeaB
	return dx*dx+dz*dz <= 1
}

// polyline returns the plan distance to a polyline and the arc length at the nearest point.
func (VolcanoIsland) polyline(p Vec2, pts []Vec2) (dist, along float64) {
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

// cone is the volcano's height without the lava channel.
func (VolcanoIsland) cone(x, z float64) float64 {
	d := P(x, z).Sub(volcanoIslandPeak)
	r, a := d.Length(), d.Angle()
	lava := math.Atan2(-0.25, 1)
	// Notch in the eastern rim where the lava spills over.
	da := math.Remainder(a-lava, 2*math.Pi)
	notch := 2.6 * math.Exp(-da*da/0.09)
	rimY := volcanoIslandRim + 0.4*volcanoIslandNoise(math.Cos(a)*3, math.Sin(a)*3, 5) - notch
	if r <= volcanoIslandCrater {
		// Steep inner walls down to the lava lake.
		k := volcanoIslandSmooth((volcanoIslandCrater - r) / 2.8)
		return volcanoIslandLerp(rimY+0.5, volcanoIslandFloor, k)
	}
	base := volcanoIslandBase + 2.5*volcanoIslandNoise(math.Cos(a)*2.2+7, math.Sin(a)*2.2+3, 9)
	u := (base - r) / (base - volcanoIslandCrater)
	if u <= 0 {
		return u * 6 // Below sea level past the shore.
	}
	h := (rimY + 0.5) * math.Pow(u, 1.45)
	// Radial gullies, fading at the rim and the shore.
	fade := volcanoIslandSmooth((r-volcanoIslandCrater)/6) * volcanoIslandSmooth((base-r)/10)
	h -= 1.9 * fade * volcanoIslandGully(x, z)
	h += 0.5 * fade * volcanoIslandFBM(x*0.12, z*0.12, 17)
	// A parasitic cinder cone on the south-eastern flank, with its own small crater.
	if dc := P(x, z).Sub(volcanoIslandCinder).Length(); dc < 8 {
		mound := math.Max(h, volcanoIslandCone(x, z)+5.5*volcanoIslandSmooth((7.5-dc)/5)-
			2.4*volcanoIslandSmooth((2.2-dc)/1.4))
		h = volcanoIslandLerp(h, mound, volcanoIslandSmooth((8-dc)/3))
	}
	return h
}

// volcanoIslandCone is the flank height under the cinder cone, sampled at its center so the
// little cone stands level instead of sliding down the slope.
func volcanoIslandCone(x, z float64) float64 {
	c := volcanoIslandCinder
	r := c.Sub(volcanoIslandPeak).Length()
	u := (volcanoIslandBase - r) / (volcanoIslandBase - volcanoIslandCrater)
	center := (volcanoIslandRim + 0.5) * math.Pow(u, 1.45)
	// Blend toward the real slope at the rim of the cone.
	d := P(x, z).Sub(c)
	dir := c.Sub(volcanoIslandPeak).Scale(1 / r)
	return center - 0.25*d.Dot(dir)*(volcanoIslandRim+0.5)*1.45*math.Pow(u, 0.45)/(volcanoIslandBase-volcanoIslandCrater)
}

// volcanoIslandGully is 0 on the ridges of the cone and up to 1 in its radial gullies.
func volcanoIslandGully(x, z float64) float64 {
	d := P(x, z).Sub(volcanoIslandPeak)
	r, a := d.Length(), d.Angle()
	ridge := math.Sin(13*a + 1.6*math.Sin(3*a) + 1.1*volcanoIslandNoise(r*0.09, a*2, 13))
	return math.Max(0, ridge) * math.Max(0, ridge)
}

func (s VolcanoIsland) raw(x, z float64) float64 {
	if !s.inSea(x, z) {
		return -1
	}
	h := s.cone(x, z)
	coast := 1.6 * volcanoIslandFBM(x*0.09, z*0.09, 21)
	blob := func(c Vec2, rx, rz, height, w float64) float64 {
		r := math.Hypot((x-c.X)/rx, (z-c.Y)/rz)
		e := (1-r)*math.Min(rx, rz) + coast
		return height * volcanoIslandSmooth(e/w)
	}
	low := math.Max(
		blob(volcanoIslandVillage, 17, 12.5, volcanoIslandVillageY, 7), // Village lowland and beach.
		blob(volcanoIslandHeadland, 7.5, 6.5, volcanoIslandHeadY, 2.2), // Lighthouse headland.
	)
	low = math.Max(low, blob(P(-14, -30), 14, 9, 1.8, 6)) // Northern beach.
	if low > h {
		h = low
	}
	// Lava channel: a trough between raised levees, ending in a delta at the shore.
	if d, along := s.polyline(P(x, z), volcanoIslandLava); d < 4.5 {
		w := 1.4 + along*0.02
		levee := 1.0 * math.Exp(-(d-w-0.8)*(d-w-0.8)/0.6)
		trough := -0.7 * volcanoIslandSmooth((w-d)/0.8+0.5)
		end := volcanoIslandLava[len(volcanoIslandLava)-1]
		if h < 1 {
			// The delta: a low lobe of fresh lava pushing into the sea.
			dd := P(x, z).Sub(end.Add(P(-1.5, 0))).Length()
			h = math.Max(h, 1.6*volcanoIslandSmooth((4.2-dd)/1.5))
		} else {
			h += levee + trough
		}
	}
	return h
}

// height is the memoized terrain height.
func (s VolcanoIsland) height(land *volcanoIslandLand, x, z float64) float64 {
	key := [2]int64{int64(math.Round(x * 1000)), int64(math.Round(z * 1000))}
	if h, ok := land.cache[key]; ok {
		return h
	}
	h := s.raw(x, z)
	land.cache[key] = h
	return h
}

// shell fills the cells of a predicate within t of its outside (sideways or above).
func (VolcanoIsland) shell(g *Grid, box Box, t float64, in func(Vec3) bool) {
	g.Fill(Func{box, func(p Vec3) bool {
		if !in(p) {
			return false
		}
		return !(in(p.Add(V(t, 0, 0))) && in(p.Add(V(-t, 0, 0))) && in(p.Add(V(0, 0, t))) &&
			in(p.Add(V(0, 0, -t))) && in(p.Add(V(0, t, 0))))
	}})
}

// cell returns the lattice cell under a world point.
func (VolcanoIsland) cell(g *Grid, p Vec3) Cell { return Nearest(p.Scale(1 / g.Span())) }

func (s VolcanoIsland) Build(g *Grid) Order {
	land := &volcanoIslandLand{cache: make(map[[2]int64]float64)}
	H := func(x, z float64) float64 { return s.height(land, x, z) }
	site := Box{V(-61, -0.2, -55), V(59, volcanoIslandRim+2, 57)}
	plate := Box{V(-61, -0.2, -55), V(59, 0.2, 57)}
	slope := func(x, z float64) float64 {
		dx := (H(x+0.6, z) - H(x-0.6, z)) / 1.2
		dz := (H(x, z+0.6) - H(x, z-0.6)) / 1.2
		return math.Hypot(dx, dz)
	}

	// Sea plate with pale shallows, reef lagoons and surf.
	g.Use(Water).Fill(Func{plate, func(p Vec3) bool { return s.inSea(p.X, p.Z) }})
	// shore is the distance (in 0.5 steps, capped at 9) from a sea point to the land.
	shoreCache := map[Cell]float64{}
	shore := func(p Vec3) float64 {
		c := s.cell(g, V(p.X, 0, p.Z))
		if d, ok := shoreCache[c]; ok {
			return d
		}
		d := 9.0
		for r := 0.5; r < 9 && d == 9; r += 0.5 {
			for k := 0; k < 16; k++ {
				q := Polar(p.Plan(), r, float64(k)*math.Pi/8)
				if H(q.X, q.Y) >= 0.45 {
					d = r
					break
				}
			}
		}
		shoreCache[c] = d
		return d
	}
	lavaSea := func(x, z float64) bool { return x > 30 && z < 2 && z > -22 }
	// A fringing reef: a broken ring of coral a few units out, with a pale lagoon inside.
	reef := func(p Vec3) bool {
		if lavaSea(p.X, p.Z) || H(p.X, p.Z) >= 0.1 {
			return false
		}
		d := shore(p) + 1.5*volcanoIslandNoise(p.X*0.15, p.Z*0.15, 62)
		return d > 4.5 && d < 7 && volcanoIslandFBM(p.X*0.12, p.Z*0.12, 61) > -0.25
	}
	g.Paint(Glass, Func{plate, func(p Vec3) bool {
		c := s.cell(g, p)
		if Hash01(int(c.X), int(c.Z), 6) < 0.025 {
			return true
		}
		d := shore(p)
		if d < 5.5 && !lavaSea(p.X, p.Z) {
			return Hash01(int(c.X), int(c.Z), 7) < 0.85-0.1*d
		}
		return d < 2.5 && Hash01(int(c.X), int(c.Z), 7) < 0.7
	}})
	g.Paint(White, Func{plate, func(p Vec3) bool {
		h := H(p.X, p.Z)
		c := s.cell(g, p)
		return h > 0.0 && h < 0.45 && Hash01(int(c.X), int(c.Z), 5) < 0.55
	}})
	// Coral heads on the reef, one layer proud of the sea, with surf on the outer edge.
	coralBox := Box{V(-61, 0.25, -55), V(59, 0.6, 57)}
	coralAt := func(p Vec3) bool {
		c := s.cell(g, p)
		return s.inSea(p.X, p.Z) && reef(p) && Hash01(int(c.X), int(c.Z), 63) < 0.75
	}
	g.Use(Terracotta).Fill(Func{coralBox, coralAt})
	for i, col := range []Color{Sand, Red, Gold, Brick} {
		lo := -0.9 + 0.45*float64(i)
		g.Paint(col, Func{coralBox, func(p Vec3) bool {
			n := volcanoIslandNoise(p.X*0.4, p.Z*0.4, 64+i)
			return coralAt(p) && n > lo && n < lo+0.25
		}})
	}
	g.Paint(White, Func{plate, func(p Vec3) bool {
		c := s.cell(g, p)
		return !reef(p) && reef(p.Add(V(0.9, 0, 0))) && Hash01(int(c.X), int(c.Z), 66) < 0.6 ||
			!reef(p) && reef(p.Add(V(0, 0, 0.9))) && Hash01(int(c.X), int(c.Z), 67) < 0.4
	}})

	// Land: a hollow skin over the height field.
	inLand := func(p Vec3) bool {
		if p.Y < -0.2 {
			return false
		}
		h := H(p.X, p.Z)
		return h >= 0.45 && p.Y <= h
	}
	g.Use(Charcoal)
	s.shell(g, site, 1.2, inLand)

	top := func(p Vec3) bool { h := H(p.X, p.Z); return h >= 0.45 && p.Y > h-0.9 }
	// Upper slopes: dark scoria, pale ash streaks down the gullies, a few tephra bands.
	onLand := func(p Vec3) bool { return p.Y > 0.25 && H(p.X, p.Z) >= 0.45 }
	streak := func(p Vec3, seed int) float64 {
		d := p.Plan().Sub(volcanoIslandPeak)
		a, r := d.Angle(), d.Length()
		return volcanoIslandFBM(math.Cos(a)*7+r*0.05, math.Sin(a)*7+r*0.05, seed)
	}
	g.Paint(DarkStone, Func{site, func(p Vec3) bool { return onLand(p) && streak(p, 31) > 0.1 }})
	g.Paint(Stone, Func{site, func(p Vec3) bool {
		c := s.cell(g, p)
		gully := volcanoIslandGully(p.X, p.Z)
		return onLand(p) && p.Y > 9 && gully > 0.7 && Hash01(int(c.X), int(c.Y), int(c.Z), 32) < 0.6
	}})
	g.Paint(Terracotta, Func{site, func(p Vec3) bool {
		c := s.cell(g, p)
		return onLand(p) && p.Y > volcanoIslandRim-6 && streak(p, 36) > 0.15 &&
			Hash01(int(c.X), int(c.Y), int(c.Z), 35) < 0.5
	}})
	// Jungle: green on the lower slopes, climbing higher in the gullies.
	jungleTop := func(x, z float64) float64 { return 13 + 4*volcanoIslandFBM(x*0.1, z*0.1, 41) }
	g.Paint(Foliage, Func{site, func(p Vec3) bool {
		return top(p) && H(p.X, p.Z) < jungleTop(p.X, p.Z) && H(p.X, p.Z) > 1.2
	}})
	g.Paint(Grass, Func{site, func(p Vec3) bool {
		h := H(p.X, p.Z)
		if !top(p) || h <= 1.2 {
			return false
		}
		jt := jungleTop(p.X, p.Z)
		c := s.cell(g, p)
		if h > jt && h < jt+3 { // Scrub fringe above the forest.
			return Hash01(int(c.X), int(c.Y), int(c.Z), 43) < 0.35*(jt+3-h)/3
		}
		return h < jt && volcanoIslandFBM(p.X*0.15, p.Z*0.15, 45) > 0.15
	}})
	// Beaches.
	g.Paint(Sand, Func{site, func(p Vec3) bool {
		h := H(p.X, p.Z)
		return top(p) && h < 1.9 && slope(p.X, p.Z) < 1.2
	}})
	g.Paint(Sand, Func{site, func(p Vec3) bool {
		d := p.Plan().Sub(volcanoIslandVillage).Length()
		return top(p) && H(p.X, p.Z) <= volcanoIslandVillageY+0.3 && d < 12 &&
			volcanoIslandNoise(p.X*0.4, p.Z*0.4, 47) > -0.2
	}})
	// Headland cliffs in bare rock with a grassy cap.
	g.Paint(DarkStone, Func{site, func(p Vec3) bool {
		d := p.Plan().Sub(volcanoIslandHeadland).Length()
		return d < 10 && p.Y > 0.5 && p.Y < volcanoIslandHeadY-0.6 && slope(p.X, p.Z) > 1.2
	}})
	g.Paint(Grass, Func{site, func(p Vec3) bool {
		d := p.Plan().Sub(volcanoIslandHeadland).Length()
		return d < 9 && top(p) && H(p.X, p.Z) > volcanoIslandHeadY-0.8
	}})

	// Oxidized red-brown scoria on the cinder cone.
	g.Paint(Brick, Func{site, func(p Vec3) bool {
		c := s.cell(g, p)
		d := p.Plan().Sub(volcanoIslandCinder).Length()
		return top(p) && d < 5.5-1.5*Hash01(int(c.X), int(c.Z), 51)
	}})
	g.Paint(Terracotta, Func{site, func(p Vec3) bool {
		c := s.cell(g, p)
		d := p.Plan().Sub(volcanoIslandCinder).Length()
		return top(p) && d < 5.5 && Hash01(int(c.X), int(c.Y), int(c.Z), 52) < 0.25
	}})
	g.Paint(DarkStone, Func{site, func(p Vec3) bool {
		c := s.cell(g, p)
		d := p.Plan().Sub(volcanoIslandCinder).Length()
		return top(p) && d < 5.5 && Hash01(int(c.X), int(c.Y), int(c.Z), 53) < 0.22
	}})
	g.Paint(Charcoal, Func{site, func(p Vec3) bool {
		d := p.Plan().Sub(volcanoIslandCinder).Length()
		return top(p) && d < 1.6
	}})
	s.rocks(g, H)
	s.lava(g, H, site)
	s.crater(g, H)
	s.jungle(g, H, slope)
	s.village(g, H)
	s.lighthouse(g, H)
	s.plume(g)
	return Sweep(volcanoIslandPeak, math.Pi*0.75)
}

// lava paints the channel, its crusted levees, the delta and the steam at the sea.
func (s VolcanoIsland) lava(g *Grid, H func(x, z float64) float64, site Box) {
	top := func(p Vec3) bool { h := H(p.X, p.Z); return h >= 0.45 && p.Y > h-1.0 }
	end := volcanoIslandLava[len(volcanoIslandLava)-1]
	delta := func(p Vec2) float64 { return p.Sub(end.Add(P(-1.5, 0))).Length() }
	g.Paint(Charcoal, Func{site, func(p Vec3) bool {
		d, _ := s.polyline(p.Plan(), volcanoIslandLava)
		return top(p) && (d < 3.6 || delta(p.Plan()) < 5)
	}})
	g.Paint(Red, Func{site, func(p Vec3) bool {
		d, along := s.polyline(p.Plan(), volcanoIslandLava)
		w := 1.4 + along*0.02
		c := s.cell(g, p)
		if top(p) && (d < 3.6 || delta(p.Plan()) < 5) && Hash01(int(c.X), int(c.Y), int(c.Z), 71) < 0.1 {
			return true // Glowing cracks in the crust.
		}
		return top(p) && (d < w+0.1 || delta(p.Plan()) < 2.4)
	}})
	g.Paint(Gold, Func{site, func(p Vec3) bool {
		d, along := s.polyline(p.Plan(), volcanoIslandLava)
		w := 1.4 + along*0.02
		if !top(p) || d >= w {
			return false
		}
		band := int(math.Floor(along / 1.6))
		return d < 0.55*w && Hash01(band, 73) < 0.65
	}})
	g.Paint(Terracotta, Func{site, func(p Vec3) bool {
		d, along := s.polyline(p.Plan(), volcanoIslandLava)
		w := 1.4 + along*0.02
		return top(p) && d >= w-0.1 && d < w+0.7
	}})
	// Steam where the lava meets the sea.
	var puffs []Ellipsoid
	for i := 0; i < 12; i++ {
		t := float64(i) / 11
		c := end.Add(P(-1+6*t+2*(Hash01(i, 75)-0.5), -3*t+4*(Hash01(i, 76)-0.5)))
		r := 0.9 + 1.1*t + 0.4*Hash01(i, 77)
		puffs = append(puffs, Ellipsoid{c.At(1.0 + 8*t + Hash01(i, 74)), V(r*1.15, r, r*1.1)})
	}
	in := func(p Vec3) bool {
		for _, e := range puffs {
			if e.Contains(p) {
				return true
			}
		}
		return false
	}
	box := Box{V(end.X-6, 0.2, end.Y-10), V(end.X+12, 16, end.Y+8)}
	g.Use(White)
	s.shell(g, box, 1.2, in)
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		c := s.cell(g, p)
		col, _ := g.ColorOf(c)
		return col == White && p.Y > 0.6 && Hash01(int(c.X), int(c.Y), int(c.Z), 78) < 0.2
	}})
}

// crater paints the lava lake and the scorched inner walls.
func (s VolcanoIsland) crater(g *Grid, H func(x, z float64) float64) {
	box := Box{V(volcanoIslandPeak.X-10, volcanoIslandFloor-2, volcanoIslandPeak.Y-10),
		V(volcanoIslandPeak.X+10, volcanoIslandRim+2, volcanoIslandPeak.Y+10)}
	g.Paint(Red, Func{box, func(p Vec3) bool {
		r := p.Plan().Sub(volcanoIslandPeak).Length()
		return r < volcanoIslandCrater && p.Y < volcanoIslandFloor+1.6
	}})
	g.Paint(Terracotta, Func{box, func(p Vec3) bool {
		r := p.Plan().Sub(volcanoIslandPeak).Length()
		c := s.cell(g, p)
		return r < volcanoIslandCrater && p.Y >= volcanoIslandFloor+1.6 && p.Y < volcanoIslandFloor+3.2 &&
			Hash01(int(c.X), int(c.Y), int(c.Z), 81) < 0.5
	}})
	g.Paint(Gold, Func{box, func(p Vec3) bool {
		r := p.Plan().Sub(volcanoIslandPeak).Length()
		return r < volcanoIslandCrater-2.6 && p.Y < volcanoIslandFloor+0.6 &&
			volcanoIslandNoise(p.X*0.55, p.Z*0.55, 83) > -0.05
	}})
	// Scorched dark lip around the rim.
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		r := p.Plan().Sub(volcanoIslandPeak).Length()
		return r >= volcanoIslandCrater-0.6 && r < volcanoIslandCrater+1.8 && p.Y > H(p.X, p.Z)-0.9
	}})
}

// palm is a curved leaning trunk with drooping fronds.
func (s VolcanoIsland) palm(g *Grid, base Vec3, lean Vec2, height float64, seed int) {
	pts := make([]Vec3, 0, 6)
	for i := 0; i <= 5; i++ {
		t := float64(i) / 5
		off := lean.Scale(t * t * height * 0.45)
		pts = append(pts, V(base.X+off.X, base.Y+t*height, base.Z+off.Y))
	}
	g.With(Wood, func() {
		for i := 0; i+1 < len(pts); i++ {
			g.Fill(Capsule{pts[i], pts[i+1], 0.5})
		}
	})
	crown := pts[len(pts)-1]
	phase := Hash01(seed, 1) * math.Pi
	for k := 0; k < 7; k++ {
		a := phase + float64(k)*2*math.Pi/7
		tip := Polar(crown.Plan(), 3.4, a)
		mid := Polar(crown.Plan(), 1.9, a)
		col := Grass
		if k%2 == 1 {
			col = Foliage
		}
		g.With(col, func() {
			g.Fill(Capsule{crown.Add(V(0, 0.2, 0)), mid.At(crown.Y + 0.6), 0.5},
				Capsule{mid.At(crown.Y + 0.6), tip.At(crown.Y - 1.0), 0.45})
		})
	}
	g.With(Gold, func() { g.Fill(Sphere(crown.Add(V(0, -0.5, 0)), 0.6)) })
}

// jungle plants broadleaf canopy clumps on the lower slopes and palms along the shore.
func (s VolcanoIsland) jungle(g *Grid, H func(x, z float64) float64, slope func(x, z float64) float64) {
	lavaD := func(p Vec2) float64 { d, _ := s.polyline(p, volcanoIslandLava); return d }
	village := func(p Vec2) bool { return p.Sub(volcanoIslandVillage).Length() < 12 }
	headland := func(p Vec2) bool { return p.Sub(volcanoIslandHeadland).Length() < 8 }
	for i := -52; i <= 52; i += 3 {
		for k := -50; k <= 50; k += 3 {
			x, z := float64(i)+2*Hash01(i, k, 91), float64(k)+2*Hash01(i, k, 92)
			h := H(x, z)
			jt := 12 + 4*volcanoIslandFBM(x*0.1, z*0.1, 41)
			p := P(x, z)
			if h < 2.2 || h > jt || slope(x, z) > 1.5 || lavaD(p) < 5.5 || village(p) || headland(p) ||
				p.Sub(volcanoIslandCinder).Length() < 7 ||
				Hash01(i, k, 93) < 0.08 {
				continue
			}
			r := 1.7 + 1.2*Hash01(i, k, 94)
			col := Foliage
			if Hash01(i, k, 95) < 0.3 {
				col = Grass
			}
			if Hash01(i, k, 96) < 0.12 && slope(x, z) < 1.0 {
				// Emergent rainforest giant: tall trunk, broad layered crown.
				g.With(Wood, func() { g.Fill(Cylinder{Center: p, Radius: 0.6, Y0: h - 0.5, Y1: h + 5}) })
				g.With(Foliage, func() { g.Fill(Ellipsoid{V(x, h+5.2, z), V(3.4, 1.4, 3.4)}) })
				g.With(Grass, func() { g.Fill(Ellipsoid{V(x+0.4, h+6.4, z-0.3), V(2.3, 1.2, 2.3)}) })
				continue
			}
			g.With(Wood, func() { g.Fill(Cylinder{Center: p, Radius: 0.5, Y0: h - 0.5, Y1: h + 1.5}) })
			g.With(col, func() { g.Fill(Ellipsoid{V(x, h+1.9, z), V(r, r*0.75, r)}) })
		}
	}
	// Palm groves on the beaches and in the village.
	n := 0
	for i := -52; i <= 52; i += 4 {
		for k := -50; k <= 50; k += 4 {
			x, z := float64(i)+3*Hash01(i, k, 101), float64(k)+3*Hash01(i, k, 102)
			h := H(x, z)
			p := P(x, z)
			if h < 0.8 || h > 4.2 || slope(x, z) > 0.9 || lavaD(p) < 6 || headland(p) || Hash01(i, k, 103) < 0.35 {
				continue
			}
			if village(p) && p.Sub(volcanoIslandVillage).Length() < 9 {
				continue
			}
			// Lean toward the sea: away from the peak.
			out := p.Sub(volcanoIslandPeak)
			out = out.Scale(1 / out.Length()).Rotate((Hash01(i, k, 104) - 0.5) * 1.2)
			s.palm(g, V(x, h-0.4, z), out, 6.5+2.5*Hash01(i, k, 105), n)
			n++
		}
	}
}

// hut is a round thatched hut on a raised floor with a dark doorway.
func (s VolcanoIsland) hut(g *Grid, c Vec2, base float64, r float64, door float64) {
	floor := base + 0.6
	wallTop := floor + 2.4
	g.With(Wood, func() {
		g.Fill(Cylinder{Center: c, Radius: r + 0.4, Y0: base - 1, Y1: floor})
		g.Fill(Cylinder{Center: c, Radius: r, Inner: r - 1.1, Y0: floor, Y1: wallTop})
	})
	g.With(Charcoal, func() {
		g.Fill(Cylinder{Center: c, Radius: r - 1.1, Y0: floor, Y1: floor + 0.4})
	})
	dc := Polar(c, r-0.5, door)
	g.Carve(Translate(RotateY(BlockAt(P(0, 0), 2.4, 1.6, floor+0.1, floor+2.0), P(0, 0), door), dc.At(0)))
	g.With(Sand, func() {
		g.FillVolume(Frustum{c, r + 1.3, 0, wallTop - 0.6, wallTop + r + 1.9})
	})
	g.Carve(Frustum{c, r - 0.2, 0, wallTop - 0.7, wallTop + r - 0.6})
	g.Paint(Charcoal, Cylinder{Center: c, Radius: r - 0.6, Y0: floor + 0.4, Y1: wallTop - 0.6},
		Frustum{c, r + 0.5, 0, wallTop - 0.7, wallTop + r + 0.6})
	g.With(Wood, func() { g.Fill(Cylinder{Center: c, Radius: 0.45, Y0: wallTop + r + 1.2, Y1: wallTop + r + 2.6}) })
}

// canoe is an outrigger canoe: a slim hull, a float on one side and two booms.
func (s VolcanoIsland) canoe(g *Grid, c Vec2, y float64, angle float64, hull Color) {
	const L, beam = 7.0, 1.3
	local := func(p Vec3) (float64, float64) {
		q := p.Plan().Sub(c).Rotate(-angle)
		return q.X, q.Y
	}
	box := Box{V(c.X-5, y-0.3, c.Y-5), V(c.X+5, y+2, c.Y+5)}
	g.With(hull, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			u, v := local(p)
			k := u / (L / 2)
			if math.Abs(k) > 1 {
				return false
			}
			w := beam / 2 * math.Sqrt(1-k*k)
			sheer := y + 0.9 + 0.5*k*k
			return p.Y >= y && p.Y <= sheer && math.Abs(v) <= math.Max(w, 0.45)
		}})
	})
	g.Carve(Func{box, func(p Vec3) bool {
		u, v := local(p)
		return math.Abs(u) < L/2-1.4 && math.Abs(v) < 0.2 && p.Y > y+0.5
	}})
	at := func(u, v, h float64) Vec3 { return c.Add(P(u, v).Rotate(angle)).At(y + h) }
	g.With(Wood, func() {
		g.Fill(Capsule{at(-2.2, 2.3, 0.3), at(2.2, 2.3, 0.3), 0.45})
		g.Fill(Capsule{at(-1.2, 0, 1.0), at(-1.2, 2.3, 0.7), 0.4}, Capsule{at(1.2, 0, 1.0), at(1.2, 2.3, 0.7), 0.4})
	})
}

// village places thatched huts, a meeting house, the jetty and canoes on the lowland.
func (s VolcanoIsland) village(g *Grid, H func(x, z float64) float64) {
	c := volcanoIslandVillage
	ground := func(q Vec2, r float64) float64 {
		h := 0.0
		for k := 0; k < 8; k++ {
			p := Polar(q, r, float64(k)*math.Pi/4)
			h = math.Max(h, H(p.X, p.Y))
		}
		return math.Round(math.Max(h, H(q.X, q.Y))*2) / 2
	}
	// A ring of round huts facing the plaza.
	for i := 0; i < 7; i++ {
		a := -0.5*math.Pi + float64(i)*2*math.Pi/7 + 0.3*Hash01(i, 111)
		r := 7.5 + 1.5*Hash01(i, 112)
		hc := Polar(c, r, a)
		hr := 2.2 + 0.5*Hash01(i, 113)
		s.hut(g, hc, ground(hc, hr), hr, a+math.Pi)
	}
	// Outlying huts toward the beaches.
	for i, a := range []float64{0.9, 1.6, 3.3, 3.9} {
		hc := Polar(c, 13+1.5*Hash01(i, 114), a)
		hr := 2.0 + 0.4*Hash01(i, 115)
		s.hut(g, hc, ground(hc, hr), hr, a+math.Pi)
	}
	// Meeting house in the plaza: an open-sided long roof on posts.
	base := ground(c, 3)
	g.With(Wood, func() {
		g.Fill(BlockAt(c, 7, 4.4, base-1, base+0.5))
		for _, dx := range []float64{-3, 0, 3} {
			for _, dz := range []float64{-1.8, 1.8} {
				g.Fill(Cylinder{Center: c.Add(P(dx, dz)), Radius: 0.5, Y0: base, Y1: base + 3})
			}
		}
	})
	g.With(Sand, func() {
		g.Fill(Gable{c.Add(P(-4.5, -3.2)), c.Add(P(4.5, 3.2)), base + 2.8, AxisX})
	})
	g.Carve(Gable{c.Add(P(-3.8, -2.5)), c.Add(P(3.8, 2.5)), base + 2.2, AxisX})
	g.Paint(Charcoal, Gable{c.Add(P(-4.2, -2.6)), c.Add(P(4.2, 2.6)), base + 2.6, AxisX})
	g.With(Red, func() { g.Fill(Cylinder{Center: c, Radius: 1.2, Y0: base + 0.5, Y1: base + 1.1}) })
	g.With(Gold, func() { g.Fill(Cylinder{Center: c, Radius: 0.6, Y0: base + 1.1, Y1: base + 1.5}) })

	// Jetty: from the beach out toward the south-west.
	dir := P(-0.75, 0.66)
	dir = dir.Scale(1 / dir.Length())
	start := c
	for t := 0.0; t < 40; t += 0.25 {
		start = c.Add(dir.Scale(t))
		if H(start.X, start.Y) < 1.2 {
			break
		}
	}
	start = start.Sub(dir.Scale(2))
	const length = 12.0
	end := start.Add(dir.Scale(length))
	deckY := 1.3
	side := P(-dir.Y, dir.X)
	ang := dir.Angle()
	g.With(Wood, func() {
		deck := Func{BoxOf(start.At(deckY-0.4), end.At(deckY+0.2)).Grow(2), func(p Vec3) bool {
			q := p.Plan().Sub(start)
			u, v := q.Dot(dir), q.Dot(side)
			return u >= 0 && u <= length && math.Abs(v) <= 1.3 && p.Y >= deckY-0.4 && p.Y <= deckY+0.2
		}}
		g.Fill(deck)
		// T-head at the end.
		g.Fill(Func{BoxOf(end.At(deckY-0.4), end.At(deckY+0.2)).Grow(4), func(p Vec3) bool {
			q := p.Plan().Sub(end)
			u, v := q.Dot(dir), q.Dot(side)
			return u >= -2.4 && u <= 0 && math.Abs(v) <= 4 && p.Y >= deckY-0.4 && p.Y <= deckY+0.2
		}})
		for u := 2.0; u <= length; u += 3 {
			for _, v := range []float64{-1.3, 1.3} {
				q := start.Add(dir.Scale(u)).Add(side.Scale(v))
				g.Fill(Cylinder{Center: q, Radius: 0.45, Y0: -0.2, Y1: deckY + 0.9})
			}
		}
		for _, v := range []float64{-4, 4} {
			q := end.Add(side.Scale(v)).Sub(dir.Scale(1.2))
			g.Fill(Cylinder{Center: q, Radius: 0.5, Y0: -0.2, Y1: deckY + 1.2})
		}
	})
	// Canoes moored along the jetty and drawn up on the beach.
	s.canoe(g, start.Add(dir.Scale(6)).Add(side.Scale(3.4)), 0.0, ang, Red)
	s.canoe(g, start.Add(dir.Scale(8)).Add(side.Scale(-4.4)), 0.0, ang+0.15, Wood)
	s.canoe(g, end.Add(dir.Scale(3)).Add(side.Scale(-6.5)), 0.0, ang-0.5, Terracotta)
	// An overwater hut on piles beside the T-head.
	s.stiltHut(g, end.Add(side.Scale(7)).Sub(dir.Scale(1)), deckY, ang)
	for i, q := range []Vec2{start.Add(side.Scale(5)), start.Add(side.Scale(-6)).Add(dir.Scale(-1))} {
		h := H(q.X, q.Y)
		col := []Color{Red, Wood}[i]
		s.canoe(g, q, math.Max(h, 0.4)-0.2, ang+math.Pi/2+0.3*float64(i), col)
	}
}

// lighthouse stands on the south-eastern headland with a small keeper's hut.
func (s VolcanoIsland) lighthouse(g *Grid, H func(x, z float64) float64) {
	c := volcanoIslandHeadland.Add(P(1, 1))
	base := math.Round(H(c.X, c.Y)*2)/2 - 0.5
	topY := base + 17
	g.With(White, func() {
		g.FillVolume(Frustum{c, 2.7, 1.9, base, topY})
	})
	g.Carve(Frustum{c, 1.5, 0.8, base + 0.6, topY - 0.6})
	g.Paint(Red, Func{Box{V(c.X-3, base, c.Y-3), V(c.X+3, topY, c.Y+3)}, func(p Vec3) bool {
		return int(math.Floor((p.Y-base)/2.4))%2 == 1
	}})
	// Door and small windows.
	g.Carve(Arch{V(c.X-0.0, base+0.6, c.Y+2), 0.8, 1.4, 2.4, AxisZ})
	for _, y := range []float64{base + 5.5, base + 10, base + 14} {
		g.Carve(BlockAt(c.Add(P(1.6, 0.6)), 1.6, 1.0, y, y+1.2))
	}
	g.Paint(Charcoal, Frustum{c, 2.0, 1.3, base + 0.6, topY - 0.6})
	// Gallery, lantern and cap.
	g.With(Charcoal, func() {
		g.Fill(Cylinder{Center: c, Radius: 3.0, Y0: topY, Y1: topY + 0.6})
		g.Fill(Cylinder{Center: c, Radius: 3.0, Inner: 2.4, Y0: topY + 0.6, Y1: topY + 1.3})
	})
	g.With(Glass, func() { g.Fill(Cylinder{Center: c, Radius: 1.7, Y0: topY + 0.6, Y1: topY + 2.8}) })
	g.With(Gold, func() { g.Fill(Cylinder{Center: c, Radius: 0.7, Y0: topY + 0.6, Y1: topY + 2.4}) })
	g.With(Red, func() {
		g.FillVolume(Frustum{c, 2.3, 0, topY + 2.8, topY + 4.9})
		g.Fill(Sphere(c.At(topY+5.2), 0.55))
	})
	// Keeper's hut.
	k := c.Add(P(-4.5, -2.5))
	kb := math.Round(H(k.X, k.Y)*2)/2 - 0.3
	g.With(White, func() {
		g.Fill(BlockAt(k, 4.5, 3.6, kb-1, kb+2.6))
	})
	g.Carve(BlockAt(k, 2.5, 1.6, kb+0.5, kb+2.8))
	g.Carve(BlockAt(k.Add(P(0, 1.8)), 1.2, 1.4, kb+0.5, kb+2.1))
	g.Paint(Charcoal, BlockAt(k, 3.3, 2.4, kb, kb+2.6))
	g.With(Red, func() { g.Fill(Gable{k.Add(P(-2.8, -2.3)), k.Add(P(2.8, 2.3)), kb + 2.6, AxisX}) })
}

// plume is the ash and smoke column drifting from the crater.
func (s VolcanoIsland) plume(g *Grid) {
	drift := P(1, -0.55)
	drift = drift.Scale(1 / drift.Length())
	var puffs []Ellipsoid
	const n = 22
	for i := 0; i < n; i++ {
		t := float64(i) / (n - 1)
		along := 30 * math.Pow(t, 1.8)
		c := volcanoIslandPeak.Add(drift.Scale(along))
		y := volcanoIslandFloor + 2.5 + 27*math.Pow(t, 0.75)
		r := 2.0 + 4.8*t
		puffs = append(puffs, Ellipsoid{c.At(y), V(r, r*0.9, r)})
		// Billows around the core make a lumpy, cauliflower outline.
		for k := 0; k < 4; k++ {
			a := Hash01(i, k, 121) * 2 * math.Pi
			off := Polar(P(0, 0), r*(0.65+0.35*Hash01(i, k, 122)), a)
			rr := r * (0.4 + 0.3*Hash01(i, k, 123))
			dy := r * (Hash01(i, k, 124) - 0.3) * 0.8
			puffs = append(puffs, Ellipsoid{c.Add(off).At(y + dy), V(rr, rr*0.85, rr)})
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
	g.Use(DarkStone)
	s.shell(g, b, 1.2, in)
	// Dark ash at the vent, paling to white smoke aloft, mottled.
	g.Paint(Stone, Func{b, func(p Vec3) bool {
		c := s.cell(g, p)
		col, _ := g.ColorOf(c)
		k := (p.Y - volcanoIslandRim) / 22
		return in(p) && col == DarkStone && Hash01(int(c.X), int(c.Y), int(c.Z), 125) < k*1.2
	}})
	g.Paint(White, Func{b, func(p Vec3) bool {
		c := s.cell(g, p)
		col, _ := g.ColorOf(c)
		k := (p.Y - volcanoIslandRim - 8) / 16
		return in(p) && col == Stone && Hash01(int(c.X), int(c.Y), int(c.Z), 126) < k*1.3
	}})
	g.Paint(Charcoal, Func{b, func(p Vec3) bool {
		c := s.cell(g, p)
		col, _ := g.ColorOf(c)
		return in(p) && col == DarkStone && p.Y < volcanoIslandRim+4 && Hash01(int(c.X), int(c.Y), int(c.Z), 127) < 0.45
	}})
	// Glow of the lava lit underside at the vent.
	g.Paint(Terracotta, Func{b, func(p Vec3) bool {
		c := s.cell(g, p)
		return in(p) && p.Y < volcanoIslandFloor+4.5 && Hash01(int(c.X), int(c.Y), int(c.Z), 128) < 0.5
	}})
}

// stiltHut is a rectangular thatched hut on piles over the water, joined to the jetty.
func (s VolcanoIsland) stiltHut(g *Grid, c Vec2, deckY, angle float64) {
	rot := func(sol Solid) Solid { return Translate(RotateY(sol, P(0, 0), angle), c.At(0)) }
	g.With(Wood, func() {
		g.Fill(rot(BlockAt(P(0, 0), 6, 5, deckY-0.4, deckY+0.2)))
		g.Fill(rot(BlockAt(P(0, -3.2), 1.6, 2, deckY-0.4, deckY+0.2)))
		for _, x := range []float64{-2.6, 2.6} {
			for _, z := range []float64{-2.1, 2.1} {
				g.Fill(rot(Cylinder{Center: P(x, z), Radius: 0.45, Y0: -0.2, Y1: deckY}))
			}
		}
		g.Fill(rot(Subtract(BlockAt(P(0, 0), 4.6, 3.6, deckY, deckY+2.6), BlockAt(P(0, 0), 2.8, 1.8, deckY, deckY+3))))
	})
	g.Carve(rot(BlockAt(P(0, -1.6), 1.4, 1.2, deckY+0.3, deckY+2.2)))
	g.Carve(rot(BlockAt(P(-1.6, 0), 1.2, 1.0, deckY+1.0, deckY+2.0)), rot(BlockAt(P(1.6, 0), 1.2, 1.0, deckY+1.0, deckY+2.0)))
	g.Paint(Charcoal, rot(BlockAt(P(0, 0), 3.4, 2.4, deckY, deckY+2.6)))
	g.With(Sand, func() {
		g.Fill(rot(Gable{P(-3.3, -2.7), P(3.3, 2.7), deckY + 2.4, AxisX}))
	})
	g.Carve(rot(Gable{P(-2.4, -1.8), P(2.4, 1.8), deckY + 2.2, AxisX}))
}

// rocks scatters hollow black basalt boulders and sea stacks along the shore.
func (s VolcanoIsland) rocks(g *Grid, H func(x, z float64) float64) {
	near := func(q Vec2, r float64) bool {
		for k := 0; k < 8; k++ {
			p := Polar(q, r, float64(k)*math.Pi/4)
			if H(p.X, p.Y) >= 0.45 {
				return true
			}
		}
		return false
	}
	placed := 0
	for i := 0; i < 400 && placed < 18; i++ {
		a := Hash01(i, 131) * 2 * math.Pi
		r := 30 + 16*Hash01(i, 132)
		c := Polar(volcanoIslandPeak, r, a)
		if !s.inSea(c.X, c.Y) || H(c.X, c.Y) >= 0.1 || !near(c, 3.5) || near(c, 1.2) ||
			c.Sub(volcanoIslandVillage).Length() < 22 {
			continue
		}
		rad := 1.0 + 1.3*Hash01(i, 133)
		top := 1.5 + 3.5*Hash01(i, 134)
		seed := 140 + i
		in := func(p Vec3) bool {
			if p.Y < -0.2 || p.Y > top {
				return false
			}
			d := p.Plan().Sub(c)
			u := p.Y / top
			rr := rad * (1 - 0.3*u) * (0.85 + 0.25*volcanoIslandNoise(math.Cos(d.Angle())*2+p.Y*0.3, math.Sin(d.Angle())*2, seed))
			if u > 0.7 {
				k := (u - 0.7) / 0.3
				rr *= math.Sqrt(math.Max(1-k*k, 0))
			}
			return d.Length() <= rr
		}
		g.Use(Charcoal)
		s.shell(g, Box{V(c.X-rad*1.3, 0, c.Y-rad*1.3), V(c.X+rad*1.3, top, c.Y+rad*1.3)}, 1.2, in)
		g.Paint(DarkStone, Func{Box{V(c.X-rad*1.3, 0.6, c.Y-rad*1.3), V(c.X+rad*1.3, top, c.Y+rad*1.3)}, func(p Vec3) bool {
			q := s.cell(g, p)
			return Hash01(int(q.X), int(q.Y), int(q.Z), 135) < 0.3
		}})
		placed++
	}
}
