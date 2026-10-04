package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// EasterIsland is Ahu Tongariki on the south-east coast of Rapa Nui. Fifteen moai of
// dark volcanic tuff stand in a row on the long stone ahu with their backs to the
// Pacific: long faces under heavy brows, long noses, pouting lips and jutting chins,
// elongated ears and arms carved down their sides, a few still wearing red scoria pukao
// topknots. A paved ramp runs from the platform down to the grassy coastal plain, where
// horses graze among small ruined ahu, a fallen moai, a lone moai near the entrance and
// a toppled pukao. Behind, to the north-west, the crater of Rano Raraku rises with its
// tuff quarry face, half-buried heads on its slope and a reed-fringed crater lake. A
// black lava coastline with boulders and white surf meets the sea layer behind the ahu.
// The moai face +Z; the sea lies to -Z.
type EasterIsland struct{}

func init() { Register(EasterIsland{}) }

func (EasterIsland) Name() string  { return "easter_island" }
func (EasterIsland) Title() string { return "Easter Island" }

const (
	easterIslandSea      = 0.5   // Top of the sea layer.
	easterIslandGround   = 2.2   // Plain level around the ahu.
	easterIslandAhuTop   = 4.6   // Platform top: the moai bases.
	easterIslandAhuX0    = -38.0 // Platform ends.
	easterIslandAhuX1    = 60.0
	easterIslandAhuBack  = -14.0 // Seaward wall.
	easterIslandAhuFront = -6.5  // Front edge of the platform top.
	easterIslandRamp     = -2.0  // Foot of the paved ramp.
	easterIslandSpacing  = 6.6   // Moai spacing along the ahu.
)

var (
	easterIslandCrater = P(-61, -18) // Rano Raraku hill center.
	easterIslandLake   = P(-63, -21) // Crater bowl center.
)

// ---------------------------------------------------------------------------------
// Helpers.

func (EasterIsland) smooth(e0, e1, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-e0)/(e1-e0)))
	return t * t * (3 - 2*t)
}

func (EasterIsland) noise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	top := a + (b-a)*sx
	bot := c + (d-c)*sx
	return (top+(bot-top)*sz)*2 - 1
}

func (e EasterIsland) fbm(x, z float64, seed int) float64 {
	return 0.6*e.noise(x, z, seed) + 0.3*e.noise(x*2.03+17, z*2.03+5, seed+1) +
		0.1*e.noise(x*4.1+3, z*4.1+29, seed+2)
}

// cellHash is a deterministic value per lattice cell.
func (EasterIsland) cellHash(c Cell, seed int) float64 {
	return Hash01(int(c.X), int(c.Y), int(c.Z), seed)
}

// each visits every lattice cell of g whose center lies in the world box.
func (EasterIsland) each(g *Grid, b Box, f func(Cell, Vec3)) {
	s := g.Span()
	inv := 2 / s
	x0, x1 := int32(math.Ceil(b.Min.X*inv)), int32(math.Floor(b.Max.X*inv))
	y0, y1 := int32(math.Ceil(b.Min.Y*inv)), int32(math.Floor(b.Max.Y*inv))
	z0, z1 := int32(math.Ceil(b.Min.Z*inv)), int32(math.Floor(b.Max.Z*inv))
	for y := y0; y <= y1; y++ {
		p := y & 1
		for z := z0 + ((z0&1)^p)&1; z <= z1; z += 2 {
			for x := x0 + ((x0&1)^p)&1; x <= x1; x += 2 {
				c := Cell{x, y, z}
				f(c, g.World(c))
			}
		}
	}
}

// put occupies or recolors one cell.
func (EasterIsland) put(g *Grid, c Cell, col Color) {
	if !c.Valid() {
		return
	}
	g.Remove(c)
	prev := g.Brush()
	g.Use(col).Add(c)
	g.Use(prev)
}

func (EasterIsland) cross(a, b Vec3) Vec3 {
	return V(a.Y*b.Z-a.Z*b.Y, a.Z*b.X-a.X*b.Z, a.X*b.Y-a.Y*b.X)
}

func (EasterIsland) unit(a Vec3) Vec3 { return a.Scale(1 / a.Length()) }

// ---------------------------------------------------------------------------------
// Terrain.

// rim is the inside distance to the rounded-rectangle site edge.
func (EasterIsland) rim(x, z float64) float64 {
	const x0, x1, z0, z1, r = -90.0, 70.0, -46.0, 24.0, 14.0
	cx, cz := (x0+x1)/2, (z0+z1)/2
	hx, hz := (x1-x0)/2-r, (z1-z0)/2-r
	dx, dz := math.Max(math.Abs(x-cx)-hx, 0), math.Max(math.Abs(z-cz)-hz, 0)
	outside := math.Hypot(dx, dz)
	if outside > 0 {
		return r - outside
	}
	return r + math.Min(hx-math.Abs(x-cx), hz-math.Abs(z-cz))
}

// coast is the shoreline z behind the ahu: land lies at larger z.
func (e EasterIsland) coast(x float64) float64 {
	return -20 + 2.2*e.fbm(x*0.09, 3.7, 41)
}

// crater returns the Rano Raraku height at a point (0 away from the hill).
func (e EasterIsland) crater(x, z float64) float64 {
	d := P(x, z).Sub(easterIslandCrater)
	r := math.Hypot(d.X, d.Y*1.1)
	wobble := 1.2 * e.fbm(x*0.12, z*0.12, 57)
	// Outer cone with rounded shoulders.
	h := 24 * e.smooth(29, 9.5, r+wobble)
	// The rim is higher on the far (north-west) side.
	dir := d.Scale(1 / math.Max(d.Length(), 1e-6))
	tilt := -0.6*dir.X - 0.8*dir.Y
	h *= 1 + 0.13*tilt
	// Crater bowl.
	b := P(x, z).Sub(easterIslandLake)
	rb := math.Hypot(b.X, b.Y)
	bowl := 11 + 15*e.smooth(6, 13, rb+0.5*wobble)
	return math.Min(h, bowl)
}

// land is the terrain height; values below the sea top are open water.
func (e EasterIsland) land(x, z float64) float64 {
	rim := e.rim(x, z)
	if rim <= 0 {
		return -1
	}
	shore := z - e.coast(x)
	roll := 0.9 * e.fbm(x*0.055, z*0.055, 23)
	plain := easterIslandGround + 1.6*e.smooth(0, 22, z) + roll*e.smooth(-6, 10, z)
	// Level the ahu site and the plaza in front of it.
	if d := e.ahuDist(x, z); d < 14 {
		plain += (easterIslandGround - plain) * e.smooth(14, 4, d)
	}
	h := plain * math.Min(e.smooth(-1.5, 4.5, shore), 1)
	h = math.Max(h, e.crater(x, z))
	// Everything falls to the sea at the site edge.
	return h*e.smooth(1.5, 9, rim) - 0.2*(1-e.smooth(1.5, 9, rim))
}

// ahuDist is the plan distance to the ahu and its ramp.
func (EasterIsland) ahuDist(x, z float64) float64 {
	dx := math.Max(math.Max(easterIslandAhuX0-x, x-easterIslandAhuX1), 0)
	dz := math.Max(math.Max(easterIslandAhuBack-z, z-easterIslandRamp), 0)
	return math.Hypot(dx, dz)
}

// slope is the terrain gradient magnitude.
func (e EasterIsland) slope(x, z float64) float64 {
	const d = 0.6
	gx := (e.land(x+d, z) - e.land(x-d, z)) / (2 * d)
	gz := (e.land(x, z+d) - e.land(x, z-d)) / (2 * d)
	return math.Hypot(gx, gz)
}

// terrain fills the ground shell, the crater lake and the sea layer.
func (e EasterIsland) terrain(g *Grid) {
	box := Box{V(-91, -0.5, -47), V(71, 0, 25)}
	s := g.Span()
	inv := 2 / s
	x0, x1 := int32(math.Ceil(box.Min.X*inv)), int32(math.Floor(box.Max.X*inv))
	z0, z1 := int32(math.Ceil(box.Min.Z*inv)), int32(math.Floor(box.Max.Z*inv))
	for cz := z0; cz <= z1; cz++ {
		for cx := x0; cx <= x1; cx++ {
			if (cx^cz)&1 != 0 {
				continue
			}
			par := cx & 1
			x, z := float64(cx)*s/2, float64(cz)*s/2
			if e.rim(x, z) <= 0 {
				continue
			}
			h := e.land(x, z)
			sl := e.slope(x, z)
			lo := math.Max(h-1.15-0.9*sl, 0)
			top := h
			water := 0.0
			if h < easterIslandSea {
				lo, top = 0, easterIslandSea
				water = 1
			}
			// Crater lake.
			lb := P(x, z).Sub(easterIslandLake).Length()
			lake := lb < 9 && h < 11.8
			if lake {
				top = 11.8
			}
			j0 := int32(math.Ceil(lo * inv))
			if (j0 & 1) != par&1 {
				j0++
			}
			for cy := j0; float64(cy)*s/2 <= top+1e-9; cy += 2 {
				c := Cell{cx, cy, cz}
				y := float64(cy) * s / 2
				col := e.groundColor(c, x, y, z, h, sl)
				if water > 0 {
					col = e.seaColor(c, x, z)
				} else if lake && y > h-0.1 {
					col = Water
					if lb > 6.4 && Hash01(int(cx), int(cz), 77) < 0.55 {
						col = Foliage // Totora reeds.
					}
				}
				e.put(g, c, col)
			}
		}
	}
}

func (e EasterIsland) groundColor(c Cell, x, y, z, h, sl float64) Color {
	shore := z - e.coast(x)
	rim := e.rim(x, z)
	hash := e.cellHash(c, 3)
	// Black lava rock along the shore and where the land drops into the sea.
	if h < 1.6 || (h < 3 && (shore < 3.5 || rim < 9)) {
		switch {
		case hash < 0.45:
			return Charcoal
		case hash < 0.8:
			return DarkStone
		default:
			return Slate
		}
	}
	// Bare tuff on the quarry face and crater cliffs, grass on gentler ground.
	rock := e.fbm(x*0.25, z*0.25+y*0.2, 9)
	if h > 4 && sl > 1.45-0.45*rock {
		switch {
		case hash < 0.12:
			return Charcoal
		case rock > 0.2:
			return DarkStone
		case rock < -0.25:
			return Sand
		default:
			return Stone
		}
	}
	// A worn dirt track from the plain to the ahu ramp.
	if d := math.Abs(x - (-6 + 9*math.Sin(z*0.12))); d < 1.1+0.5*e.noise(x, z, 71) && z > -2.5 {
		return Sand
	}
	n := e.fbm(x*0.08, z*0.08, 61)
	switch {
	case n > 0.35 && hash < 0.8:
		return Foliage
	case n < -0.38 && hash < 0.75:
		return Sand
	case hash < 0.08:
		return Foliage
	case hash > 0.97:
		return Sand
	}
	return Grass
}

func (e EasterIsland) seaColor(c Cell, x, z float64) Color {
	// Surf where the sea meets land.
	near := false
	for _, d := range [][2]float64{{1.6, 0}, {-1.6, 0}, {0, 1.6}, {0, -1.6}, {1.1, 1.1}, {-1.1, 1.1}, {1.1, -1.1}, {-1.1, -1.1}} {
		if e.land(x+d[0], z+d[1]) > 0.9 {
			near = true
			break
		}
	}
	hash := e.cellHash(c, 5)
	if near && hash < 0.7 {
		return White
	}
	// Broken lines of breakers offshore.
	shore := e.coast(x) - z
	wave := math.Sin(shore*1.15 + e.noise(x*0.15, z*0.15, 13)*2.5)
	if wave > 0.93 && hash < 0.6 && shore > 3 {
		return White
	}
	return Water
}

// ---------------------------------------------------------------------------------
// Moai.

// moaiPart classifies a point in moai units (height 14 to the crown, facing +z):
// -1 outside, 0 body, 1 arm/hand, 2 eye socket, 3 pukao, 4 mouth.
func (e EasterIsland) moaiPart(x, u, z float64, pukao bool) int {
	ax := math.Abs(x)
	if pukao && u >= 13.7 {
		r := math.Hypot(x, z+0.25)
		if u <= 15.5 && r <= 1.65 {
			return 3
		}
		if u <= 16.1 && r <= 0.85 {
			return 3
		}
	}
	if u < 0 || u > 14 {
		return -1
	}
	// Arms carved down the sides, long fingers meeting over the belly.
	if u < 6.0 {
		p := V(ax, u, z)
		for _, sg := range [][2]Vec3{
			{V(1.9, 5.6, -0.25), V(2.0, 2.4, 0.2)},
			{V(2.0, 2.4, 0.2), V(0.55, 1.6, 1.15)},
		} {
			if (Capsule{sg[0], sg[1], 0.36}).Contains(p) {
				return 1
			}
		}
	}
	// Torso: a rounded slab, broad at the shoulders.
	if u <= 6.6 {
		a := 1.8 + 0.2*e.smooth(0, 5, u) - 0.3*e.smooth(5.6, 6.6, u)
		b := 1.25
		if z < 0 {
			b = 1.15
		}
		if math.Pow(ax/a, 2.5)+math.Pow(math.Abs(z)/b, 2.5) <= 1 {
			return 0
		}
	}
	if u < 6.2 {
		return -1
	}
	// Head.
	hw := 1.7 - 0.12*e.smooth(12.6, 14, u)
	back := 1.3
	if u > 13.4 {
		f := 1 - 0.25*(u-13.4)/0.6
		hw *= f
		back *= f
	}
	// Elongated ears.
	if ax >= hw-0.3 && ax <= hw+0.38 && u >= 8.1 && u <= 11.7 && z >= -0.65 && z <= 0.45 {
		return 0
	}
	if ax > hw {
		return -1
	}
	var front float64
	part := 0
	switch {
	case u < 7.0: // Jutting chin.
		front = 1.8
	case u < 7.4: // Mouth line.
		front = 1.3
		part = 4
	case u < 8.0: // Pouting lips.
		front = 1.65
	case u < 10.4: // Cheeks.
		front = 1.2
	case u < 11.3: // Heavy brow.
		front = 1.75
	default: // Receding forehead.
		front = 1.5 - 0.55*(u-11.3)/2.7
	}
	front -= 0.6 * (ax / hw) * (ax / hw)
	eye := ax >= 0.45 && ax <= 1.4 && u >= 9.3 && u < 10.4
	if eye {
		front = math.Min(front, 0.75)
	}
	// Long nose, widening and jutting to its base.
	if u >= 8.0 && u <= 10.9 {
		t := (10.9 - u) / 2.9
		if ax <= 0.35+0.45*t {
			front = math.Max(front, 1.4+1.5*t)
			if u < 8.4 && z > 1.2 && ax > 0.15 {
				part = 4 // Shadowed nostrils under the nose.
			}
		}
	}
	if z > front {
		return -1
	}
	if z < 0 && math.Pow(ax/hw, 3)+math.Pow(-z/back, 3) > 1 {
		return -1
	}
	if eye && z > 0.2 {
		return 2
	}
	if part == 4 && z > 0.9 && ax < 0.95 {
		return 4
	}
	return 0
}

// easterIslandMoai places one moai: origin on the base center, up axis, facing axis.
type easterIslandMoai struct {
	origin   Vec3
	up, face Vec3
	scale    float64
	pukao    bool
	seed     int
	clip     bool // Clip below the terrain (buried heads).
	stone    Color
}

func (e EasterIsland) moai(g *Grid, m easterIslandMoai) {
	up := e.unit(m.up)
	face := e.unit(m.face.Sub(up.Scale(m.face.Dot(up))))
	right := e.cross(up, face)
	local := func(p Vec3) (float64, float64, float64) {
		d := p.Sub(m.origin)
		return d.Dot(right) / m.scale, d.Dot(up) / m.scale, d.Dot(face) / m.scale
	}
	box := EmptyBox()
	for _, lx := range []float64{-2.5, 2.5} {
		for _, lu := range []float64{-0.5, 16.2} {
			for _, lz := range []float64{-1.8, 2.8} {
				q := m.origin.Add(right.Scale(lx * m.scale)).Add(up.Scale(lu * m.scale)).Add(face.Scale(lz * m.scale))
				box = box.Union(Box{q, q})
			}
		}
	}
	clip := func(p Vec3) bool {
		return !m.clip || p.Y > e.land(p.X, p.Z)-1.2
	}
	shape := Func{box, func(p Vec3) bool {
		x, u, z := local(p)
		return e.moaiPart(x, u, z, m.pukao) >= 0 && clip(p)
	}}
	tmp := NewGridWithSpan(g.Span())
	tmp.FillVolume(shape)
	for _, c := range tmp.Cells() {
		p := tmp.World(c)
		x, u, z := local(p)
		part := e.moaiPart(x, u, z, m.pukao)
		hash := e.cellHash(c, m.seed)
		patch := e.noise(p.X*0.6+float64(m.seed), p.Y*0.6+p.Z*0.4, 91)
		col := m.stone
		switch {
		case part == 1 || part == 2 || part == 4:
			col = Charcoal
		case part == 3:
			col = Red
			if hash < 0.3 {
				col = Terracotta
			}
		case hash < 0.03:
			col = Charcoal
		case patch < -0.55 || hash > 0.97:
			col = Stone
		}
		e.put(g, c, col)
	}
}

// ---------------------------------------------------------------------------------
// Ahu and props.

func (e EasterIsland) ahu(g *Grid) {
	top := easterIslandAhuTop
	base := easterIslandGround - 1.2
	box := Box{V(easterIslandAhuX0-9, base, easterIslandAhuBack), V(easterIslandAhuX1+9, top, easterIslandRamp)}
	e.each(g, box, func(c Cell, p Vec3) {
		// Main platform with tapering wings at both ends.
		end := math.Max(easterIslandAhuX0-p.X, p.X-easterIslandAhuX1)
		h := top
		front := easterIslandAhuFront
		if end > 0 {
			h = top - end*0.3
			front = easterIslandAhuFront - end*0.25
		}
		if h < easterIslandGround+0.3 {
			return
		}
		inPlatform := p.Z <= front && p.Y <= h
		// Paved ramp sloping down to the plain.
		ramp := false
		if p.Z > front {
			t := (p.Z - front) / (easterIslandRamp - front)
			rh := h - (h-easterIslandGround)*t
			ramp = p.Y <= rh+0.1 && t <= 1
		}
		if !inPlatform && !ramp {
			return
		}
		// Hollow interior.
		if inPlatform && p.Z > easterIslandAhuBack+1.3 && p.Z < front-1.3 && p.Y < h-1.3 && end < -1.3 {
			return
		}
		hash := e.cellHash(c, 17)
		col := DarkStone
		switch {
		case p.Z < easterIslandAhuBack+1.0:
			// Seaward wall of large fitted slabs.
			bx := int(math.Floor((p.X + 0.6*math.Floor(p.Y/1.2)) / 2.2))
			by := int(math.Floor(p.Y / 1.2))
			r := Hash01(bx, by, 23)
			switch {
			case r < 0.45:
				col = Stone
			case r < 0.8:
				col = DarkStone
			default:
				col = Sand
			}
		default:
			// Paving of rounded beach stones (poro) in irregular patches, grass in the joints.
			tx, tz := int(math.Floor(p.X/1.7+0.5*math.Floor(p.Z/1.7))), int(math.Floor(p.Z/1.7))
			r := Hash01(tx, tz, 29)
			switch {
			case hash < 0.06:
				col = Charcoal
			case ramp && hash > 0.93:
				col = Grass
			case r < 0.3:
				col = Stone
			case r < 0.85:
				col = DarkStone
			default:
				col = Charcoal
			}
		}
		e.put(g, c, col)
	})
}

// horse grazes with its head down; dir is the plan heading.
func (e EasterIsland) horse(g *Grid, at Vec2, dir float64, col Color, grazing bool) {
	y0 := e.land(at.X, at.Y) - 0.2
	f := P(math.Cos(dir), math.Sin(dir))
	side := P(-f.Y, f.X)
	w := func(a, h, b float64) Vec3 {
		q := at.Add(f.Scale(a)).Add(side.Scale(b))
		return q.At(y0 + h)
	}
	parts := []Solid{
		Capsule{w(-0.9, 2.5, 0), w(0.9, 2.6, 0), 0.62},
		Capsule{w(0.85, 0, 0.3), w(0.85, 2.3, 0.3), 0.33},
		Capsule{w(0.85, 0, -0.3), w(0.85, 2.3, -0.3), 0.33},
		Capsule{w(-0.9, 0, 0.3), w(-0.9, 2.3, 0.3), 0.33},
		Capsule{w(-0.9, 0, -0.3), w(-0.9, 2.3, -0.3), 0.33},
	}
	if grazing {
		parts = append(parts, Capsule{w(1.2, 2.8, 0), w(2.1, 1.3, 0), 0.42}, Capsule{w(2.1, 1.3, 0), w(2.4, 0.5, 0), 0.4})
	} else {
		parts = append(parts, Capsule{w(1.2, 2.8, 0), w(1.9, 3.8, 0), 0.42}, Capsule{w(1.9, 3.8, 0), w(2.6, 3.4, 0), 0.4})
	}
	tail := Capsule{w(-1.5, 2.6, 0), w(-1.9, 1.4, 0), 0.3}
	g.With(col, func() { g.Fill(parts...) })
	mane := Charcoal
	if col == Charcoal {
		mane = Wood
	}
	g.With(mane, func() { g.Fill(tail) })
}

// ruin is a small tumbled ahu: a low broken platform with scattered blocks.
func (e EasterIsland) ruin(g *Grid, at Vec2, length, depth float64, angle float64, seed int) {
	y := e.land(at.X, at.Y)
	f := P(math.Cos(angle), math.Sin(angle))
	n := int(length / 1.6)
	for i := 0; i < n; i++ {
		for j := 0; j < int(depth/1.6); j++ {
			r := Hash01(i, j, seed)
			if r < 0.25 {
				continue
			}
			h := 0.8 + 1.4*Hash01(i, j, seed+1)
			c := at.Add(f.Scale(float64(i)*1.6 - length/2)).Add(P(-f.Y, f.X).Scale(float64(j)*1.6 - depth/2))
			col := DarkStone
			if r > 0.7 {
				col = Stone
			}
			g.With(col, func() {
				g.Fill(RotateY(BlockAt(c, 1.5, 1.5, y-0.8, y+h), c, angle))
			})
		}
	}
	// Tumbled blocks around it.
	for k := 0; k < 7; k++ {
		a := Hash01(k, seed, 3) * 2 * math.Pi
		d := length/2 + 1 + 3*Hash01(k, seed, 4)
		c := Polar(at, d, a)
		yy := e.land(c.X, c.Y)
		g.With(DarkStone, func() {
			g.Fill(RotateY(BlockAt(c, 1.6, 1.2, yy-0.6, yy+0.9), c, a))
		})
	}
}

func (e EasterIsland) boulders(g *Grid) {
	for k := 0; k < 1600; k++ {
		x := -90 + 160*Hash01(k, 1, 99)
		z := -46 + 70*Hash01(k, 2, 99)
		h := e.land(x, z)
		if h < 0.1 || h > 3.2 {
			continue
		}
		// Only along the lava shore, never on the plain.
		if z-e.coast(x) > 5 && (e.rim(x, z) > 6 || z > 0 || Hash01(k, 5, 99) < 0.6) {
			continue
		}
		if e.ahuDist(x, z) < 2.5 {
			continue
		}
		r := 0.7 + 1.2*Hash01(k, 3, 99)
		col := Charcoal
		if Hash01(k, 4, 99) < 0.35 {
			col = DarkStone
		}
		g.With(col, func() {
			g.FillVolume(Ellipsoid{V(x, h+0.1, z), V(r*1.2, r*0.8, r)})
		})
	}
}

// ---------------------------------------------------------------------------------

func (e EasterIsland) Build(g *Grid) Order {
	e.terrain(g)
	e.boulders(g)
	e.ahu(g)

	// The fifteen moai of Ahu Tongariki.
	for i := 0; i < 15; i++ {
		x := easterIslandAhuX0 + 2.75 + float64(i)*easterIslandSpacing + (Hash01(i, 6)-0.5)*0.6
		scale := 1.28 + 0.3*Hash01(i, 7)
		if i == 7 {
			scale = 1.62
		}
		lean := V((Hash01(i, 8)-0.5)*0.08, 1, (Hash01(i, 9)-0.5)*0.12)
		yaw := (Hash01(i, 10) - 0.5) * 0.12
		e.moai(g, easterIslandMoai{
			origin: V(x, easterIslandAhuTop-0.2, -10.3),
			up:     lean,
			face:   V(math.Sin(yaw), 0, math.Cos(yaw)),
			scale:  scale,
			pukao:  i == 3 || i == 9 || i == 12,
			seed:   100 + i,
			stone:  DarkStone,
		})
	}

	// Half-buried heads on the outer slope of Rano Raraku, looking out from the hill.
	for k, a := range []float64{-0.15, 0.2, 0.55, 0.9, 1.25, 1.6} {
		dir := P(math.Cos(a), math.Sin(a))
		r := 14.5 + 3*Hash01(k, 31)
		at := easterIslandCrater.Add(dir.Scale(r))
		gy := e.land(at.X, at.Y)
		scale := 1.15 + 0.25*Hash01(k, 32)
		e.moai(g, easterIslandMoai{
			origin: V(at.X, gy-7.0*scale, at.Y),
			up:     V(dir.X*0.15, 1, dir.Y*0.15),
			face:   dir.At(0),
			scale:  scale,
			seed:   200 + k,
			clip:   true,
			stone:  DarkStone,
		})
	}

	// A fallen moai lying on its back by a ruined ahu, with its pukao rolled away.
	e.ruin(g, P(28, 13), 11, 3.5, 0.12, 5)
	fallAt := P(12, 15)
	fy := e.land(fallAt.X, fallAt.Y)
	e.moai(g, easterIslandMoai{
		origin: V(fallAt.X, fy+1.0, fallAt.Y),
		up:     V(1, 0.03, -0.25),
		face:   V(0, 1, 0),
		scale:  1.25,
		seed:   300,
		stone:  DarkStone,
	})
	pk := P(40, 7)
	py := e.land(pk.X, pk.Y)
	g.With(Red, func() {
		g.FillVolume(Capsule{V(pk.X-0.7, py+1.4, pk.Y-0.5), V(pk.X+0.7, py+1.4, pk.Y+0.5), 1.6})
	})

	// The lone "travelling" moai on its small ahu near the entrance.
	lone := P(-26, 12)
	ly := e.land(lone.X, lone.Y)
	g.With(Stone, func() {
		g.Fill(BlockAt(lone, 7, 5, ly-0.8, ly+1.2))
	})
	e.moai(g, easterIslandMoai{
		origin: V(lone.X, ly+1.0, lone.Y),
		up:     V(0, 1, 0.03),
		face:   V(0.35, 0, 1),
		scale:  1.05,
		seed:   400,
		stone:  DarkStone,
	})
	e.ruin(g, P(52, 15), 8, 3, -0.4, 9)

	// Stone footings of boat-shaped hare paenga houses.
	for _, hp := range []struct {
		at    Vec2
		angle float64
	}{{P(-46, 11), 0.2}, {P(-36, 18), -0.3}, {P(30, 2), 0.1}} {
		y := e.land(hp.at.X, hp.at.Y)
		ring := Func{Box{V(hp.at.X-7, y-0.6, hp.at.Y-7), V(hp.at.X+7, y+0.7, hp.at.Y+7)}, func(p Vec3) bool {
			q := p.Plan().Sub(hp.at).Rotate(-hp.angle)
			r := math.Hypot(q.X/5.5, q.Y/1.9)
			return r <= 1 && r >= 0.62
		}}
		g.With(Stone, func() { g.Fill(ring) })
	}

	// Horses grazing on the plain.
	for _, h := range []struct {
		at  Vec2
		dir float64
		col Color
		eat bool
	}{
		{P(-2, 8), 0.4, Wood, true},
		{P(3, 13), 2.6, Charcoal, true},
		{P(44, 16), -0.6, White, false},
		{P(-14, 15), 3.6, Wood, true},
		{P(-40, 4), 1.2, Sand, true},
		{P(22, 20), 0.9, Wood, true},
		{P(-52, 16), -2.2, Charcoal, false},
	} {
		e.horse(g, h.at, h.dir, h.col, h.eat)
	}

	return Along(P(1, 0.25))
}
