package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// ColosseumRuins is the Flavian Amphitheatre as it stands today. Its elliptical shell
// (188 × 156 m proportions) keeps the full four-storey travertine façade only on the
// north side: three arcaded orders framed by engaged columns and entablatures, and the
// attic with small windows and corbels. East of it Stern's brick buttress ramps the wall
// down; to the west the break descends in a jagged diagonal. Elsewhere only pier stubs
// of the outer ring remain, so the second ring forms the ragged outside and exposes the
// vaulted ambulacra, the inner ring and the brick radial walls of the stripped cavea.
// A sector of marble seats is restored, a wooden arena deck covers the east end and the
// dark hypogeum is open in the rest. Around it: the travertine boundary stones, the
// Arch of Constantine, the Meta Sudans stump, the basalt Via Sacra and the Palatine
// slope with brick ruins and umbrella pines. East is +X, north is -Z.
type ColosseumRuins struct{}

func init() { Register(ColosseumRuins{}) }

func (ColosseumRuins) Name() string  { return "colosseum_ruins" }
func (ColosseumRuins) Title() string { return "Colosseum Ruins" }

const (
	colosseumA      = 22.0 // Arena semi-axes.
	colosseumB      = 12.7
	colosseumBays   = 80
	colosseumL1     = 5.9 // Storey tops.
	colosseumL2     = 12.2
	colosseumL3     = 18.4
	colosseumTop    = 25.8
	colosseumDeck   = 3.0 // Arena floor.
	colosseumPodium = 5.2
	colosseumW      = -2.35 // Intact outer façade spans these elliptic angles.
	colosseumE      = 0.05
	colosseumRuinSp = 2*math.Pi - (colosseumE - colosseumW)
	colosseumStep   = 2 * math.Pi / colosseumBays
	colosseumNone   = Color(255)
)

// colosseumArch is the Arch of Constantine's plan center; its passage runs along Z.
var colosseumArch = P(-30, 47)

// colosseumHill is the Palatine slope's summit.
var colosseumHill = P(-60, 56)

// colosseumRuinWalls are brick arcade fragments on the Palatine slope: plan ends,
// height above the terrain and number of arch tiers.
var colosseumRuinWalls = []struct {
	A, B   Vec2
	Height float64
	Tiers  int
}{
	{P(-44, 49), P(-49, 63), 7.2, 2},
	{P(-50, 43), P(-63, 41), 4.4, 1},
	{P(-56, 49), P(-57, 61), 4.2, 1},
	{P(-60, 46), P(-67, 53), 3.2, 1},
}

// colosseumPines are umbrella pine positions.
var colosseumPines = []Vec2{
	P(-52, 53), P(-58, 45), P(-64, 51), P(-53, 60), P(-61, 61),
	P(-50, 27), P(-14, 49), P(52, 21), P(56, -10), P(30, -41), P(-42, -33), P(-58, 6),
}

type colosseumColumn struct {
	x, z, d, phi       float64
	idx, pier          int
	fromArch, fromPier float64
	t                  float64 // Ruin depth in radians beyond the intact façade; < 0 intact.
	east               bool
	outerTop           float64
	r2Top, r3Top       float64
	site               float64 // Superellipse norm of the site; ≤ 1 inside.
	hill               float64 // Terrain top.
}

func colosseumSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func colosseumNoise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := colosseumSmooth(x-ix), colosseumSmooth(z-iz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*fx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*fx
	return a + (b-a)*fz
}

// ring maps a plan point to (d, φ): d is the offset from the arena ellipse in the family
// of ellipses with semi-axes (A + d, B + d); φ is the elliptic angle.
func (ColosseumRuins) ring(x, z float64) (float64, float64) {
	lo, hi := -colosseumB+1e-6, 120.0
	for i := 0; i < 44; i++ {
		mid := (lo + hi) / 2
		u, v := x/(colosseumA+mid), z/(colosseumB+mid)
		if u*u+v*v > 1 {
			lo = mid
		} else {
			hi = mid
		}
	}
	d := (lo + hi) / 2
	return d, math.Atan2(z/(colosseumB+d), x/(colosseumA+d))
}

func (ColosseumRuins) ruinDepth(phi float64) (float64, bool) {
	u := math.Mod(phi-colosseumE+4*math.Pi, 2*math.Pi)
	if u >= colosseumRuinSp {
		return -1, false
	}
	if u < colosseumRuinSp-u {
		return u, true
	}
	return colosseumRuinSp - u, false
}

// siteNorm is the site outline, ≤ 1 inside: an ellipse around the amphitheatre and a
// rounded lobe to the south-west for the arch and the Palatine slope.
func (ColosseumRuins) siteNorm(x, z float64) float64 {
	u, v := x/61, z/52
	a, b := (x+46)/24, (z-45)/21
	return math.Min(u*u+v*v, a*a*a*a+b*b*b*b)
}

// hillH is the terrain top: the Palatine slope tapering to the site rim.
func (s ColosseumRuins) hillH(x, z float64) float64 {
	r := math.Hypot(x-colosseumHill.X, (z-colosseumHill.Y)*1.15)
	bump := 7.5*colosseumSmooth(1-r/21) + 1.2*colosseumNoise(x/5, z/5, 31)*colosseumSmooth(1-r/18)
	taper := colosseumSmooth((1 - s.siteNorm(x, z)) / 0.35)
	return 0.5 + bump*taper
}

func (s ColosseumRuins) column(x, z float64) colosseumColumn {
	c := colosseumColumn{x: x, z: z}
	c.d, c.phi = s.ring(x, z)
	speed := math.Hypot((colosseumA+c.d)*math.Sin(c.phi), (colosseumB+c.d)*math.Cos(c.phi))
	w := c.phi + math.Pi
	c.idx = int(w/colosseumStep) % colosseumBays
	u := w - float64(int(w/colosseumStep))*colosseumStep
	c.fromPier = math.Min(u, colosseumStep-u) * speed
	c.fromArch = math.Abs(u-colosseumStep/2) * speed
	c.pier = int(math.Round(w/colosseumStep)) % colosseumBays
	half := colosseumStep / 2
	hidx := int(w / half)
	phiQ := (float64(hidx)+0.5)*half - math.Pi
	c.t, c.east = s.ruinDepth(phiQ)
	tSmooth, _ := s.ruinDepth(c.phi)
	switch {
	case c.t < 0:
		c.outerTop = colosseumTop
	case c.east:
		// Stern's buttress: a stepped brick ramp over about three bays.
		c.outerTop = colosseumTop - math.Floor(tSmooth/0.2*(colosseumTop-0.5)/1.6)*1.6
	default:
		c.outerTop = colosseumTop - c.t*52 - Hash01(hidx, 1)*2.6
	}
	if c.t < 0 {
		c.r2Top, c.r3Top = colosseumL3, 15.6
	} else {
		base := colosseumL3 - c.t*8
		if c.east {
			base = colosseumL3 - math.Max(0, c.t-0.2)*14
		}
		floor := 7.6 + 3.6*math.Sin(3.1*c.phi+0.7) + 2.4*Hash01(c.idx, 3)
		c.r2Top = math.Min(colosseumL3, math.Max(base, floor)) - Hash01(hidx, 4)*1.3
		c.r3Top = math.Max(15.6-c.t*6, 9.4+2.2*Hash01(c.idx, 5)) - Hash01(hidx, 6)*1.0
	}
	c.site = s.siteNorm(x, z)
	c.hill = s.hillH(x, z)
	return c
}

// open reports whether a ring wall point lies in an arch opening.
func colosseumOpen(y, fromArch, sill, crown, hw float64) bool {
	if y < sill || fromArch > hw || y > crown {
		return false
	}
	rise := y - (crown - hw)
	return rise <= 0 || fromArch*fromArch+rise*rise <= hw*hw
}

// cavea is the seating surface height at offset d.
func colosseumCavea(d float64) float64 { return colosseumPodium + (d-1)*0.55 }

// slab1 reports whether the outer ambulacrum's vault at a storey top survives.
func (c *colosseumColumn) slab1(top float64) bool {
	return c.outerTop >= top+0.4 && c.r2Top >= top
}

func (c *colosseumColumn) slab2(top float64) bool {
	return c.r2Top >= top && c.r3Top >= top-0.2
}

// roofed reports whether the corridor level containing y is vaulted over.
func (c *colosseumColumn) roofed(y float64, outer bool) bool {
	tops := []float64{colosseumL1, colosseumL2, colosseumL3}
	for _, top := range tops {
		if y <= top-0.8 {
			if outer {
				return c.slab1(top)
			}
			return top < colosseumL3 && c.slab2(top)
		}
	}
	return false
}

func (s ColosseumRuins) classify(c *colosseumColumn, x, y, z float64) Color {
	if c.site > 1 {
		return colosseumNone
	}
	d := c.d
	if y <= 0.5 {
		if c.d >= 1 && c.d < 19.3 && y < 0.2 {
			return colosseumNone // A single floor layer under the cavea vaults.
		}
		return s.ground(c, x, z)
	}
	h := Hash01(int(math.Floor(x*1.6)), int(math.Floor(y*1.6)), int(math.Floor(z*1.6)), 77)
	switch {
	case d < 0:
		return s.hypogeum(c, x, y, z)
	case d < 1:
		if y > colosseumPodium {
			return colosseumNone
		}
		if y > colosseumPodium-0.5 {
			return Sand
		}
		return Stone
	case d < 19.3:
		return s.cavea(c, y, h)
	case d < 20.4: // Inner ring: tuff and brick.
		if y > c.r3Top || colosseumOpen(y, c.fromArch, 0.5, 3.9, 0.9) || colosseumOpen(y, c.fromArch, 6.6, 9.8, 0.8) {
			return colosseumNone
		}
		if d > 19.85 && c.roofed(y, false) {
			return Charcoal
		}
		if Hash01(c.idx, 21) < 0.3 || Hash01(c.idx/2, int(y/3.2), 21) < 0.2 {
			return Brick
		}
		return Stone
	case d < 22.8: // Inner ambulacrum vaults.
		for _, top := range []float64{colosseumL1, colosseumL2} {
			if y <= top && y > top-0.8 && c.slab2(top) {
				if y < top-0.4 {
					return DarkStone
				}
				return Stone
			}
		}
		return colosseumNone
	case d < 24.0: // Second ring: travertine arcade.
		if y > c.r2Top {
			return colosseumNone
		}
		if colosseumOpen(y, c.fromArch, 0.5, 4.7, 1.0) || colosseumOpen(y, c.fromArch, 6.8, 10.8, 1.0) || colosseumOpen(y, c.fromArch, 13.1, 17.1, 1.0) {
			return colosseumNone
		}
		if d < 23.4 && c.roofed(y, false) || d >= 23.4 && c.roofed(y, true) {
			return Charcoal
		}
		if c.t > 0 && c.r2Top-y < 1.4 && Hash01(c.idx, 22) < 0.5 {
			return Brick
		}
		if Hash01(c.idx, int(y/1.8), 23) < 0.22 {
			return Stone
		}
		return Sand
	case d < 26.6: // Outer ambulacrum vaults and Stern's buttress.
		for _, top := range []float64{colosseumL1, colosseumL2, colosseumL3} {
			if y <= top && y > top-0.8 && c.slab1(top) {
				if y < top-0.4 {
					return DarkStone
				}
				return Stone
			}
		}
		return colosseumNone
	case d < 28.0: // Outer wall.
		return s.outerWall(c, y, h)
	case d < 28.7: // Engaged columns, entablatures, corbels.
		return s.outerFrame(c, y)
	}
	// Travertine boundary stones.
	if d >= 32.3 && d <= 33.1 && c.pier%2 == 0 && c.fromPier <= 0.45 && y <= 1.8 {
		return Sand
	}
	// Meta Sudans stump.
	if r := math.Hypot(x+44, z-35); r <= 2.6-(y-0.5)*0.4 && y <= 2.4 {
		return Brick
	}
	// Palatine slope and its ruins.
	if y <= c.hill && y > c.hill-1.6 {
		if colosseumNoise(x/3, z/3, 41) > 0.68 {
			return DarkStone
		}
		return Grass
	}
	if y > c.hill {
		for i, seg := range colosseumRuinWalls {
			ab := seg.B.Sub(seg.A)
			length := ab.Length()
			q := P(x, z).Sub(seg.A)
			along := q.Dot(ab) / length
			if along < 0 || along > length {
				continue
			}
			across := math.Abs(q.X*ab.Y-q.Y*ab.X) / length
			if across > 0.6 {
				continue
			}
			local := y - c.hill
			top := seg.Height - Hash01(i, int(along/1.5))*1.0 - math.Max(0, along-(length-4))*1.3
			m := math.Mod(along, 3.0)
			arch := m > 0.8 && m < 2.2 && (local <= 2.0 || math.Hypot(m-1.5, local-2.0) <= 0.7)
			if seg.Tiers > 1 && m > 0.8 && m < 2.2 && local >= 3.6 && (local <= 5.0 || math.Hypot(m-1.5, local-5.0) <= 0.7) {
				arch = true
			}
			if local <= top && !arch {
				if local > top-0.45 || local < 0.8 {
					return Stone
				}
				return Brick
			}
		}
	}
	return colosseumNone
}

func (s ColosseumRuins) ground(c *colosseumColumn, x, z float64) Color {
	d := c.d
	switch {
	case d < 0:
		return Charcoal
	case d < 28:
		if d >= 24 && d < 26.6 && c.slab1(colosseumL1) || d >= 20.4 && d < 22.8 && c.slab2(colosseumL1) {
			return DarkStone
		}
		if d < 19.3 {
			return DarkStone
		}
		return Stone
	case d < 33.5:
		return Stone
	}
	// Basalt roads: the Via Triumphalis through the arch and the Via Sacra to the west.
	if math.Abs(x-colosseumArch.X) <= 3.6 && z > 36 {
		return DarkStone
	}
	if x < -36 && x > -80 {
		center := 33 - (-36-x)*0.18
		if math.Abs(z-center) <= 2.8 {
			return DarkStone
		}
	}
	if math.Hypot(x+44, z-35) <= 4.2 {
		return Stone
	}
	if math.Abs(x-colosseumArch.X) <= 10 && math.Abs(z-colosseumArch.Y) <= 6 {
		return Sand
	}
	if c.hill > 1.7 {
		return colosseumNone // Terrain shell above covers it.
	}
	return Grass
}

func (s ColosseumRuins) hypogeum(c *colosseumColumn, x, y, z float64) Color {
	d := c.d
	if x >= 9 && y > colosseumDeck-0.5 && y <= colosseumDeck {
		if Hash01(int(math.Floor(x/2.2)), int(math.Floor(z/2.2)), 51) < 0.1 {
			return colosseumNone // Open trapdoors.
		}
		return Wood
	}
	if y > colosseumDeck-0.5 {
		return colosseumNone
	}
	top := colosseumDeck - 0.5
	wall := false
	if d >= -1.6 && d <= -0.85 || d >= -4.0 && d <= -3.25 {
		wall = true
		top -= Hash01(c.idx, 61) * 0.9
	}
	az := math.Abs(z)
	if !wall && az > 1.3 {
		k := math.Floor((az - 1.3) / 1.75)
		if az-1.3-k*1.75 <= 0.75 {
			wall = true
			top -= Hash01(int(k), int(math.Floor(x/3)), 62) * 1.1
		}
	}
	if !wall && math.Mod(math.Abs(x)+0.4, 5.2) <= 0.75 && Hash01(int(x/5.2), int(z/1.75), 63) > 0.4 {
		wall = true
		top -= Hash01(int(x/5.2), int(z/1.75), 64) * 0.8
	}
	if x >= 9 && d < -4.0 {
		return colosseumNone // Hidden under the deck.
	}
	if !wall || y > top {
		return colosseumNone
	}
	if y < 1.7 {
		return DarkStone
	}
	return Brick
}

func (s ColosseumRuins) cavea(c *colosseumColumn, y, h float64) Color {
	d := c.d
	hd := colosseumCavea(d)
	ruin := 0.0
	if c.t >= 0 {
		ruin = math.Min(c.t*3, 3) * Hash01(c.pier, 71)
	}
	// Restored marble seating sector on the north side.
	if c.phi > -1.42 && c.phi < -1.12 && d < 12 {
		row := math.Floor((d - 1) / 1.5)
		step := colosseumPodium + (row+1)*0.82
		if y <= step && y > step-1.3 {
			if int(row)%2 == 0 {
				return White
			}
			return Stone
		}
	}
	// Radial walls on every other pier.
	if c.pier%2 == 0 && c.fromPier <= 0.4 {
		top := hd + 1.1 - ruin - Hash01(c.pier, int(d/2.2), 72)*1.4
		door := (math.Abs(d-10.3) <= 1.3 || math.Abs(d-16.2) <= 1.1) && y <= 3.4
		if y <= top && !door {
			if y < 2.6 || d > 17.6 {
				return Stone
			}
			return Brick
		}
	}
	// Concentric walls of the substructure.
	if d >= 7.0 && d <= 7.8 || d >= 13.0 && d <= 13.8 {
		top := hd + 0.3 - ruin - Hash01(c.idx, int(d), 73)*1.2
		door := c.fromArch <= 0.8 && (y <= 2.8 || math.Hypot(c.fromArch, y-2.8) <= 0.8)
		if y <= top && !door {
			return Brick
		}
	}
	// Sloping vaults between the radial walls.
	sector := c.pier / 2
	if c.pier%2 == 1 {
		sector = c.idx / 2
	}
	keep := 0.8
	if c.t >= 0 {
		keep = 0.55 - c.t*0.25
	}
	if Hash01(sector, 74) < keep && Hash01(sector, int(d/3), 75) > 0.18 && y <= hd && y > hd-0.8 && d > 1.0 {
		if y < hd-0.4 {
			return DarkStone
		}
		return Stone
	}
	return colosseumNone
}

func (s ColosseumRuins) outerWall(c *colosseumColumn, y, h float64) Color {
	if c.outerTop < 1.6 {
		if c.fromPier <= 0.85 && y <= 0.9+1.3*Hash01(c.pier, 9) {
			return Sand
		}
		return colosseumNone
	}
	if y > c.outerTop {
		return colosseumNone
	}
	buttress := c.t >= 0 && c.east && c.t < 0.22
	if buttress {
		return Brick
	}
	if colosseumOpen(y, c.fromArch, 0.5, 4.7, 1.1) || colosseumOpen(y, c.fromArch, 6.8, 10.8, 1.1) || colosseumOpen(y, c.fromArch, 13.1, 17.1, 1.1) {
		return colosseumNone
	}
	if y > 20.3 && y < 22.3 && c.idx%2 == 0 && c.fromArch <= 0.75 {
		if c.d > 27.2 {
			return Charcoal // Attic windows, dark against the gallery behind.
		}
		return Sand
	}
	if c.d < 27.2 && c.roofed(y, true) {
		return DarkStone
	}
	if c.t > 0 && c.outerTop-y < 1.5 {
		return Brick
	}
	if c.d > 27.5 && y > 2 && h < 0.04 {
		return DarkStone // Putlog and clamp holes.
	}
	return Sand
}

func (s ColosseumRuins) outerFrame(c *colosseumColumn, y float64) Color {
	if c.outerTop < 1.6 || y > c.outerTop {
		return colosseumNone
	}
	if c.t >= 0 && c.east && c.t < 0.22 {
		return Brick
	}
	bands := [][2]float64{{4.9, 5.9}, {11.0, 12.2}, {17.3, 18.4}, {25.0, 25.8}}
	for _, b := range bands {
		if y >= b[0] && y <= b[1] {
			return White
		}
	}
	if c.fromPier <= 0.55 && y < colosseumL3 {
		return Sand // Engaged half-columns: Doric, Ionic, Corinthian.
	}
	if c.fromPier <= 0.5 && y >= colosseumL3 && c.d < 28.4 {
		return Sand // Attic pilasters.
	}
	if y >= 23.6 && y <= 24.4 && (c.fromArch <= 0.35 || math.Abs(c.fromArch-1.2) <= 0.35) && c.d < 28.45 {
		return White // Corbels for the velarium masts.
	}
	return colosseumNone
}

func (s ColosseumRuins) Build(g *Grid) Order {
	span := g.Span()
	toIdx := func(v float64) int32 { return int32(math.Floor(v * 2 / span)) }
	x0, x1 := toIdx(-76), toIdx(60)
	z0, z1 := toIdx(-50), toIdx(66)
	yMax := toIdx(colosseumTop + 0.5)
	for parity := int32(0); parity < 2; parity++ {
		for zi := z0 + ((z0&1)^parity)&1; zi <= z1; zi += 2 {
			for xi := x0 + ((x0&1)^parity)&1; xi <= x1; xi += 2 {
				x, z := float64(xi)*span/2, float64(zi)*span/2
				c := s.column(x, z)
				if c.site > 1 {
					continue
				}
				for yi := parity; yi <= yMax; yi += 2 {
					y := float64(yi) * span / 2
					col := s.classify(&c, x, y, z)
					if col == colosseumNone {
						continue
					}
					g.With(col, func() { g.Add(Cell{xi, yi, zi}) })
				}
			}
		}
	}
	s.archOfConstantine(g)
	for i, at := range colosseumPines {
		s.pine(g, at, i)
	}
	return Sweep(P(0, 0), -math.Pi/2)
}

func (s ColosseumRuins) archOfConstantine(g *Grid) {
	a := colosseumArch
	box := func(x0, y0, z0, x1, y1, z1 float64) Cuboid {
		return Block(V(a.X+x0, y0, a.Y+z0), V(a.X+x1, y1, a.Y+z1))
	}
	g.Use(Stone).Fill(box(-7.3, 0.5, -3.1, 7.3, 1.0, 3.1))
	g.Use(White).Fill(box(-6.9, 0.5, -2.0, 6.9, 7.0, 2.0))
	g.Fill(box(-7.1, 7.0, -2.3, 7.1, 7.8, 2.3))
	g.Fill(box(-6.9, 7.8, -2.0, 6.9, 11.0, 2.0))
	g.Fill(box(-7.1, 11.0, -2.3, 7.1, 11.5, 2.3))
	for _, cx := range []float64{-6.25, -2.65, 2.65, 6.25} {
		for _, sz := range []float64{-1, 1} {
			g.Use(White).Fill(box(cx-0.7, 0.5, sz*2.0, cx+0.7, 2.6, sz*3.1))
			g.Use(Sand).Fill(Cylinder{Center: P(a.X+cx, a.Y+sz*2.55), Radius: 0.55, Y0: 2.6, Y1: 6.5})
			g.Use(White).Fill(box(cx-0.7, 6.5, sz*2.0, cx+0.7, 7.8, sz*3.1))
			g.Use(Stone).Fill(box(cx-0.4, 7.8, sz*2.0, cx+0.4, 10.4, sz*2.9))
		}
	}
	for _, sz := range []float64{-1, 1} {
		g.Use(Stone).Fill(box(-2.2, 8.4, sz*2.0, 2.2, 10.5, sz*2.4))
		for _, cx := range []float64{-4.5, 4.5} {
			g.Paint(Stone, Func{box(cx-1, 4.6, sz*1.6, cx+1, 6.6, sz*2.5).Bounds(), func(p Vec3) bool {
				return math.Hypot(p.X-(a.X+cx), p.Y-5.6) <= 0.85
			}})
		}
	}
	g.Carve(
		Arch{Sill: a.At(0.5), HalfWidth: 1.8, Spring: 4.8, Depth: 10, Through: AxisZ},
		Arch{Sill: V(a.X-4.5, 0.5, a.Y), HalfWidth: 0.95, Spring: 3.2, Depth: 10, Through: AxisZ},
		Arch{Sill: V(a.X+4.5, 0.5, a.Y), HalfWidth: 0.95, Spring: 3.2, Depth: 10, Through: AxisZ},
	)
	g.Use(Stone).Fill(box(-1.8, 0.0, -3.0, 1.8, 0.5, 3.0))
}

func (s ColosseumRuins) pine(g *Grid, at Vec2, i int) {
	base := s.hillH(at.X, at.Y)
	height := 9.5 + Hash01(i, 81)*3
	lean := Polar(P(0, 0), 0.8+Hash01(i, 82)*1.6, Hash01(i, 83)*2*math.Pi)
	top := at.Add(lean)
	g.Use(Wood).Fill(Capsule{A: at.At(base - 0.3), B: top.At(base + height), Radius: 0.5})
	r := 3.8 + Hash01(i, 84)*1.4
	g.Use(Foliage).FillVolume(Ellipsoid{top.At(base + height + 0.7), V(r, 1.4, r)})
	for k := 0; k < 3; k++ {
		o := Polar(top, r*0.6, Hash01(i, k, 85)*2*math.Pi)
		g.FillVolume(Ellipsoid{o.At(base + height + 0.2 + Hash01(i, k, 86)*0.8), V(r*0.55, 1.1, r*0.55)})
	}
}
