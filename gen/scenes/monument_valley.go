package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// MonumentValley is the red-rock desert on the Arizona/Utah border, seen down the long
// straight highway toward the buttes. Three sandstone buttes rise sheer from banded
// talus slopes at the back (-Z): West Mitten and East Mitten with their slender
// "thumb" spires, and the massive Merrick Butte. Each is a hollow heightfield shell:
// Organ Rock talus in dark red bands, an orange De Chelly cliff with desert-varnish
// streaks, and a pale caprock. The asphalt road with its gold centre line runs
// straight from the front rim between the Mittens, with a vintage car heading toward
// them. A Navajo hogan with a corral and horses, the Totem Pole spire group, sand
// dunes, junipers and sage scrub dot the desert floor, which slopes to a low rim.
type MonumentValley struct{}

func init() { Register(MonumentValley{}) }

func (MonumentValley) Name() string  { return "monument_valley" }
func (MonumentValley) Title() string { return "Monument Valley" }

const (
	monumentValleyX0, monumentValleyX1 = -56.0, 56.0
	monumentValleyZ0, monumentValleyZ1 = -42.0, 30.0
	monumentValleyRoadX                = -11.0 // Road centerline (x); it runs along Z.
	monumentValleyRoadHalf             = 3.4   // Half width of the asphalt.
	monumentValleyStep                 = 0.25  // Heightfield cache resolution.
)

var (
	monumentValleyHogan  = P(22, 10)
	monumentValleyCorral = P(34, 16)
)

// monumentValleyLobe is one superelliptic rock column of a butte, with a flat (or
// domed) top at Top above the desert floor. Cap marks a caprock-topped summit.
type monumentValleyLobe struct {
	C                     Vec2
	RX, RZ, Pow, Rot, Top float64
	Dome                  float64
	Cap                   bool
}

// monumentValleyButte is a group of lobes standing on a shared talus apron that
// reaches Talus at the cliff foot and fades out TW units away.
type monumentValleyButte struct {
	Lobes     []monumentValleyLobe
	Talus, TW float64
	Seed      int
}

var monumentValleyButtes = []monumentValleyButte{
	{ // West Mitten: the hand, its thumb to the south (+X), joined by a low web.
		Lobes: []monumentValleyLobe{
			{C: P(-34.5, -17.5), RX: 10.5, RZ: 6.5, Pow: 3, Rot: 0.05, Top: 38.5, Cap: true},
			{C: P(-19.2, -16.4), RX: 2.1, RZ: 2.5, Pow: 2, Top: 35.8, Dome: 1.5, Cap: true},
			{C: P(-21.8, -16.7), RX: 3.4, RZ: 3, Pow: 2, Top: 19.2},
		},
		Talus: 14.9, TW: 11, Seed: 1,
	},
	{ // East Mitten.
		Lobes: []monumentValleyLobe{
			{C: P(5, -24.5), RX: 10, RZ: 6.4, Pow: 3, Rot: -0.06, Top: 37.4, Cap: true},
			{C: P(19.6, -23.2), RX: 2.1, RZ: 2.5, Pow: 2, Top: 34.1, Dome: 1.5, Cap: true},
			{C: P(17, -23.5), RX: 3.4, RZ: 3, Pow: 2, Top: 19.8},
		},
		Talus: 14.9, TW: 11, Seed: 2,
	},
	{ // Merrick Butte: a broad block with a lower shoulder.
		Lobes: []monumentValleyLobe{
			{C: P(38.5, -10.5), RX: 10.5, RZ: 7.8, Pow: 3.2, Rot: -0.25, Top: 33, Cap: true},
			{C: P(45.5, -2), RX: 4, RZ: 3.4, Pow: 2.4, Rot: -0.3, Top: 24.2, Cap: true},
		},
		Talus: 13.2, TW: 9.5, Seed: 3,
	},
	{ // The Totem Pole and the Yei Bi Chei spires on a shared dune-swept apron.
		Lobes: []monumentValleyLobe{
			{C: P(-41, 2), RX: 1.5, RZ: 1.4, Pow: 2, Top: 31, Dome: 1, Cap: true},
			{C: P(-37.2, 4.5), RX: 1.8, RZ: 1.6, Pow: 2, Top: 20, Dome: 1.5, Cap: true},
			{C: P(-34.4, 1.8), RX: 1.7, RZ: 1.5, Pow: 2, Top: 16, Dome: 1.5, Cap: true},
			{C: P(-39, 7.6), RX: 1.6, RZ: 1.5, Pow: 2, Top: 14.5, Dome: 1.5, Cap: true},
			{C: P(-44.8, 5.6), RX: 1.8, RZ: 1.6, Pow: 2, Top: 22, Dome: 1.5, Cap: true},
			{C: P(-46.5, 0.5), RX: 1.6, RZ: 1.5, Pow: 2, Top: 17, Dome: 1.5, Cap: true},
			{C: P(-36.6, -1.5), RX: 1.5, RZ: 1.4, Pow: 2, Top: 12.5, Dome: 1.5, Cap: true},
		},
		Talus: 6.5, TW: 7, Seed: 4,
	},
}

func monumentValleyNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	lerp := func(a, b, t float64) float64 { return a + (b-a)*t }
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return lerp(lerp(a, b, sx), lerp(c, d, sx), sz)*2 - 1
}

func monumentValleyFBM(x, z float64, seed int) float64 {
	return 0.6*monumentValleyNoise(x, z, seed) + 0.3*monumentValleyNoise(x*2.03+17, z*2.03+5, seed+1) +
		0.1*monumentValleyNoise(x*4.1+3, z*4.1+29, seed+2)
}

func monumentValleySmooth(e0, e1, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-e0)/(e1-e0)))
	return t * t * (3 - 2*t)
}

func monumentValleyLerp(a, b, t float64) float64 { return a + (b-a)*t }

// eval returns the scaled distance inside the lobe's edge (negative outside) and its
// summit height at a plan point. Angular noise flutes the cliffs vertically.
func (l monumentValleyLobe) eval(x, z float64, seed int) (e, h float64) {
	d := P(x, z).Sub(l.C).Rotate(-l.Rot)
	ax, az := math.Abs(d.X)/l.RX, math.Abs(d.Y)/l.RZ
	if ax > 8 || az > 8 {
		return -1e9, 0
	}
	r := math.Pow(math.Pow(ax, l.Pow)+math.Pow(az, l.Pow), 1/l.Pow)
	ang := math.Atan2(d.Y/l.RZ, d.X/l.RX)
	k := (l.RX + l.RZ) * 0.45
	r += 0.07*monumentValleyNoise(math.Cos(ang)*k+40, math.Sin(ang)*k+40, seed+20) +
		0.03*monumentValleyNoise(math.Cos(ang)*k*3+80, math.Sin(ang)*k*3+80, seed+21)
	e = (1 - r) * math.Min(l.RX, l.RZ)
	h = l.Top - l.Dome*r*r + 0.25*monumentValleyNoise(x*0.4, z*0.4, seed+30)
	return e, h
}

// eval returns the butte's height above the floor, the distance inside its footprint
// (negative on the talus), the local summit height and whether that summit is capped.
func (b monumentValleyButte) eval(x, z float64) (h, e, top float64, cap bool) {
	e = -1e9
	for _, l := range b.Lobes {
		le, lh := l.eval(x, z, b.Seed)
		e = math.Max(e, le)
		if le >= 0 && lh > h {
			h, top, cap = lh, lh, l.Cap
		}
	}
	if e >= 0 {
		return h, e, top, cap
	}
	if e <= -b.TW {
		return 0, e, 0, false
	}
	t := 1 + e/b.TW
	ridge := 0.14 * monumentValleyNoise(x*0.3, z*0.3, b.Seed+60)
	return b.Talus * math.Pow(t, 1.6) * (1 + ridge), e, 0, false
}

// monumentValleyTerrain caches the site heightfield (desert floor, dunes, buttes, road
// and pads) on a fine plan grid.
type monumentValleyTerrain struct {
	x0, z0 float64
	nx, nz int
	h      []float64
}

func newMonumentValleyTerrain() *monumentValleyTerrain {
	s := monumentValleyStep
	t := &monumentValleyTerrain{x0: monumentValleyX0 - 3, z0: monumentValleyZ0 - 3}
	t.nx = int((monumentValleyX1-monumentValleyX0+6)/s) + 2
	t.nz = int((monumentValleyZ1-monumentValleyZ0+6)/s) + 2
	t.h = make([]float64, t.nx*t.nz)
	for j := 0; j < t.nz; j++ {
		for i := 0; i < t.nx; i++ {
			t.h[j*t.nx+i] = t.raw(t.x0+float64(i)*s, t.z0+float64(j)*s)
		}
	}
	return t
}

// edge is the distance inside the site's rounded, ragged outline.
func (monumentValleyTerrain) edge(x, z float64) float64 {
	cx, cz := (monumentValleyX0+monumentValleyX1)/2, (monumentValleyZ0+monumentValleyZ1)/2
	hx, hz := (monumentValleyX1-monumentValleyX0)/2, (monumentValleyZ1-monumentValleyZ0)/2
	const r = 12
	qx, qz := math.Abs(x-cx)-(hx-r), math.Abs(z-cz)-(hz-r)
	out := math.Hypot(math.Max(qx, 0), math.Max(qz, 0)) + math.Min(math.Max(qx, qz), 0) - r
	return -out + 1.6*monumentValleyNoise(x*0.08, z*0.08, 9)
}

// base is the smooth desert floor: rolling swells that slope down to the rim.
func (t monumentValleyTerrain) base(x, z float64) float64 {
	s := monumentValleySmooth(0, 9, t.edge(x, z))
	return 0.25 + s*(1.3+0.45*math.Sin(z*0.13+0.6)+0.25*math.Sin(x*0.09+1.3))
}

// dune is the height of the sand sea west of the road: transverse ridges with gentle
// windward slopes and steep slip faces.
func (monumentValleyTerrain) dune(x, z float64) float64 {
	d := P(x, z).Sub(P(-28, 19))
	env := 1 - monumentValleySmooth(0.35, 1, math.Hypot(d.X/16, d.Y/12))
	if env <= 0 {
		return 0
	}
	u := (x*0.8+z*0.6)/6.5 + 0.35*monumentValleyNoise(x*0.12, z*0.12, 11)
	f := u - math.Floor(u)
	var p float64
	if f < 0.72 {
		p = f / 0.72
	} else {
		p = (1 - f) / 0.28
	}
	p = p * p * (3 - 2*p)
	return env * (0.5 + 3.0*p)
}

func (t monumentValleyTerrain) raw(x, z float64) float64 {
	d := t.edge(x, z)
	if d < 0 {
		return -10
	}
	s := monumentValleySmooth(0, 9, d)
	base := t.base(x, z)
	h := base + s*(0.45*monumentValleyFBM(x*0.05, z*0.05, 3)+0.2) + s*t.dune(x, z)
	for _, b := range monumentValleyButtes {
		if bh, _, _, _ := b.eval(x, z); bh > 0 {
			h = math.Max(h, base+bh*monumentValleySmooth(0, 4, d))
		}
	}
	for _, pad := range []struct {
		C Vec2
		R float64
	}{{monumentValleyHogan, 8}, {monumentValleyCorral, 8.5}} {
		w := monumentValleySmooth(pad.R, pad.R+4, P(x, z).Sub(pad.C).Length())
		h = monumentValleyLerp(base+0.3, h, w)
	}
	w := monumentValleySmooth(monumentValleyRoadHalf+1.4, monumentValleyRoadHalf+6, math.Abs(x-monumentValleyRoadX))
	return monumentValleyLerp(base+0.35, h, w)
}

// H is the cached terrain height (bilinear); -10 outside the site.
func (t *monumentValleyTerrain) H(x, z float64) float64 {
	fx, fz := (x-t.x0)/monumentValleyStep, (z-t.z0)/monumentValleyStep
	i, j := int(math.Floor(fx)), int(math.Floor(fz))
	if i < 0 || j < 0 || i >= t.nx-1 || j >= t.nz-1 {
		return -10
	}
	u, v := fx-float64(i), fz-float64(j)
	a, b := t.h[j*t.nx+i], t.h[j*t.nx+i+1]
	c, d := t.h[(j+1)*t.nx+i], t.h[(j+1)*t.nx+i+1]
	return monumentValleyLerp(monumentValleyLerp(a, b, u), monumentValleyLerp(c, d, u), v)
}

// fill builds the terrain as a shell about one unit thick, then colors every cell.
func (t *monumentValleyTerrain) fill(g *Grid) {
	box := Box{V(monumentValleyX0-2, -0.7, monumentValleyZ0-2), V(monumentValleyX1+2, 44, monumentValleyZ1+2)}
	g.Use(Terracotta).Fill(Func{box, func(p Vec3) bool {
		h := t.H(p.X, p.Z)
		if p.Y > h || p.Y < -0.6 {
			return false
		}
		low := h
		for k := 0; k < 8; k++ {
			a := float64(k) * math.Pi / 4
			low = math.Min(low, t.H(p.X+1.4*math.Cos(a), p.Z+1.4*math.Sin(a)))
		}
		return p.Y > low-0.85
	}})
	for _, c := range g.Cells() {
		col := t.color(g.World(c), c)
		g.Remove(c)
		g.Use(col).Add(c)
	}
}

func (t *monumentValleyTerrain) color(p Vec3, c Cell) Color {
	x, z := p.X, p.Z
	base := t.base(x, z)
	hash := func(salt int) float64 { return Hash01(int(c.X), int(c.Y), int(c.Z), salt) }

	// Road: asphalt with a dashed gold centre line, white edge lines, sand shoulders.
	if dx := math.Abs(x - monumentValleyRoadX); dx < monumentValleyRoadHalf+1.5 && p.Y > base-0.7 && t.edge(x, z) > 0.5 {
		switch {
		case dx < 0.42 && math.Mod(z+200, 5) < 3:
			return Gold
		case dx > monumentValleyRoadHalf-0.45 && dx < monumentValleyRoadHalf:
			return White
		case dx < monumentValleyRoadHalf:
			return Charcoal
		default:
			return Sand
		}
	}

	// Buttes and spires.
	bestH, best := 0.0, -1
	var bestE, bestTop float64
	var bestCap bool
	for i, b := range monumentValleyButtes {
		if h, e, top, cp := b.eval(x, z); h > bestH {
			bestH, best, bestE, bestTop, bestCap = h, i, e, top, cp
		}
	}
	if best >= 0 && bestH > 0.7 && base+bestH >= t.H(x, z)-0.6 {
		b := monumentValleyButtes[best]
		y := p.Y - base
		yy := y + 0.7*monumentValleyNoise(x*0.1, z*0.1, 80)
		if bestE >= -0.25 && y > b.Talus-0.3 { // Cliff and summit.
			if bestCap && y > bestTop-1.1 {
				switch {
				case bestE > 1.6 && hash(1) < 0.04:
					return Foliage
				case bestE > 1.2 && monumentValleyFBM(x*0.35, z*0.35, 81) > 0.2:
					return Terracotta
				}
				return Sand
			}
			if bestCap && y > bestTop-2.5 {
				return Brick
			}
			col := Hash01(int(math.Floor(x)), int(math.Floor(z)), b.Seed, 5)
			if col < 0.22 && y > bestTop-3-14*Hash01(int(math.Floor(x)), int(math.Floor(z)), b.Seed, 6) {
				return DarkStone
			}
			if m := math.Mod(yy+100, 6.4); m < 0.75 {
				return Brick
			}
			return Terracotta
		}
		// Talus: banded Organ Rock shale, scree aprons at the foot.
		if y < b.Talus*0.35 && hash(2) < 0.4 {
			if hash(3) < 0.5 {
				return Sand
			}
			return Terracotta
		}
		if m := math.Mod(yy+100, 2.6); m < 0.85 {
			return Terracotta
		}
		return Brick
	}

	if t.dune(x, z) > 0.55 {
		if hash(4) < 0.07 {
			return Terracotta
		}
		return Sand
	}
	n := monumentValleyFBM(x*0.08, z*0.08, 7)
	switch {
	case n > 0.4:
		return Sand
	case n < -0.42:
		return Brick
	case hash(5) < 0.04:
		return Sand
	}
	return Terracotta
}

func (m MonumentValley) Build(g *Grid) Order {
	t := newMonumentValleyTerrain()
	t.fill(g)
	m.track(g, t)
	m.boulders(g, t)
	m.scrub(g, t)
	m.car(g, t, P(monumentValleyRoadX+1.7, 8))
	m.hogan(g, t, monumentValleyHogan)
	m.corral(g, t, monumentValleyCorral)
	m.pickup(g, t, P(13.5, 15))
	m.sheep(g, t)
	return Along(P(0, -1))
}

// scrub scatters sage bushes and a few junipers over the open floor and lower talus.
func (m MonumentValley) scrub(g *Grid, t *monumentValleyTerrain) {
	for iz := int(monumentValleyZ0); iz <= int(monumentValleyZ1); iz++ {
		for ix := int(monumentValleyX0); ix <= int(monumentValleyX1); ix++ {
			r := Hash01(ix, iz, 300)
			if r > 0.085 {
				continue
			}
			x := float64(ix) + Hash01(ix, iz, 301) - 0.5
			z := float64(iz) + Hash01(ix, iz, 302) - 0.5
			if t.edge(x, z) < 2 || monumentValleyTrackDist(x, z) < 2 || math.Abs(x-monumentValleyRoadX) < monumentValleyRoadHalf+2.5 ||
				P(x, z).Sub(monumentValleyHogan).Length() < 7 || P(x, z).Sub(monumentValleyCorral).Length() < 7.5 {
				continue
			}
			h := t.H(x, z)
			if h-t.base(x, z) > 4 || t.dune(x, z) > 0.8 && r > 0.02 {
				continue
			}
			if r < 0.008 { // Juniper.
				g.Use(Wood).Fill(Cylinder{P(x, z), 0.45, 0, h, h + 1.3})
				g.Use(Foliage).FillVolume(Ellipsoid{V(x, h+2, z), V(1.3, 1.1, 1.3)})
				continue
			}
			col := Foliage
			if Hash01(ix, iz, 303) < 0.45 {
				col = Grass
			}
			g.Use(col).Fill(Sphere(V(x, h+0.35, z), 0.6+0.25*Hash01(ix, iz, 304)))
		}
	}
}

// car is a two-tone fifties sedan heading up the road toward the Mittens (-Z).
func (MonumentValley) car(g *Grid, t *monumentValleyTerrain, c Vec2) {
	y := t.H(c.X, c.Y) + 0.1
	x0, x1 := c.X-1.25, c.X+1.25
	front, back := c.Y-2.9, c.Y+2.9
	// Wheels.
	g.Use(Charcoal)
	for _, wz := range []float64{front + 1.0, back - 1.1} {
		g.Fill(Capsule{V(x0+0.1, y+0.55, wz), V(x1-0.1, y+0.55, wz), 0.55})
	}
	// Body with long hood and trunk, tail fins.
	g.Use(Glass).Fill(Block(V(x0, y+0.6, front), V(x1, y+1.45, back)))
	g.Fill(Block(V(x0, y+1.45, back-0.6), V(x0+0.4, y+1.9, back)), Block(V(x1-0.4, y+1.45, back-0.6), V(x1, y+1.9, back)))
	// Cabin: dark glazing below a white roof.
	g.Use(Navy).Fill(Block(V(x0+0.2, y+1.45, c.Y-1.2), V(x1-0.2, y+2.2, c.Y+1.5)))
	g.Use(White).Fill(Block(V(x0+0.15, y+2.2, c.Y-1.0), V(x1-0.15, y+2.6, c.Y+1.4)))
	// Chrome bumpers, headlights and a white side stripe.
	g.Use(Stone).Fill(Block(V(x0-0.1, y+0.5, front-0.35), V(x1+0.1, y+0.9, front)),
		Block(V(x0-0.1, y+0.5, back), V(x1+0.1, y+0.9, back+0.35)))
	g.Paint(Gold, Block(V(x0, y+1.0, front-0.1), V(x0+0.5, y+1.4, front+0.3)), Block(V(x1-0.5, y+1.0, front-0.1), V(x1, y+1.4, front+0.3)))
	g.Paint(Red, Block(V(x0, y+1.0, back-0.3), V(x0+0.5, y+1.4, back+0.1)), Block(V(x1-0.5, y+1.0, back-0.3), V(x1, y+1.4, back+0.1)))
}

// hogan is an eight-sided log dwelling with a cribbed earthen dome, its door facing
// the sunrise (+X), beside a brush-roofed ramada.
func (MonumentValley) hogan(g *Grid, t *monumentValleyTerrain, c Vec2) {
	y := t.H(c.X, c.Y)
	oct := func(r float64) []Vec2 {
		pts := make([]Vec2, 8)
		for i := range pts {
			pts[i] = Polar(c, r, (float64(i)+0.5)*math.Pi/4)
		}
		return pts
	}
	const r, wall = 4.2, 2.8
	g.Use(Wood).Fill(Subtract(Prism{oct(r), y, y + wall}, Prism{oct(r - 1.3), y - 1, y + wall + 1}))
	// Log courses with pale mud chinking.
	for k := 0.0; k < wall; k += 0.9 {
		g.Paint(Sand, Subtract(Prism{oct(r), y + k + 0.55, y + k + 0.8}, Prism{oct(r - 0.6), y - 1, y + 9}))
	}
	// Cribbed dome covered with earth.
	g.Use(DarkStone).FillVolume(Subtract(
		Ellipsoid{V(c.X, y+wall-0.3, c.Y), V(r+0.4, 3.0, r+0.4)},
		Ellipsoid{V(c.X, y+wall-1.4, c.Y), V(r-0.9, 3.0, r-0.9)},
		Block(V(c.X-r-1, y-2, c.Y-r-1), V(c.X+r+1, y+wall-0.3, c.Y+r+1)),
	))
	g.Carve(Cylinder{c, 0.7, 0, y + wall, y + wall + 4}) // Smoke hole.
	// East door with a log frame and a blanket-dark interior.
	g.Use(Wood).Fill(Block(V(c.X+r-0.5, y, c.Y-1.4), V(c.X+r+0.9, y+2.6, c.Y+1.4)))
	g.Carve(Block(V(c.X+r-2, y+0.3, c.Y-0.75), V(c.X+r+1, y+2.0, c.Y+0.75)))
	g.Paint(Charcoal, Cylinder{c, r - 0.75, 0, y - 0.5, y + wall - 0.4})
	g.Paint(Charcoal, Block(V(c.X+r-2, y+0.2, c.Y-0.8), V(c.X+r-0.3, y+2.1, c.Y+0.8)))
	// Ramada: four posts and a brush roof, to the south.
	rc := c.Add(P(1, 8))
	g.Use(Wood)
	for _, d := range []Vec2{P(-2.4, -1.8), P(2.4, -1.8), P(-2.4, 1.8), P(2.4, 1.8)} {
		p := rc.Add(d)
		g.Fill(Cylinder{p, 0.4, 0, t.H(p.X, p.Y), y + 2.8})
	}
	g.Use(Wood).Fill(BlockAt(rc, 5.8, 4.6, y+2.8, y+3.2))
	g.Use(Wood).Fill(BlockAt(rc, 5.4, 4.2, y+3.2, y+3.6))
	g.Paint(Foliage, Func{BlockAt(rc, 5.4, 4.2, y+3.2, y+3.6).Bounds(), func(p Vec3) bool {
		return Hash01(int(math.Floor(p.X*1.3)), int(math.Floor(p.Z*1.3)), 310) < 0.4
	}})
	// Woodpile and a water barrel.
	g.Use(Wood).Fill(Capsule{V(c.X-5.5, y+0.4, c.Y-2), V(c.X-5.5, y+0.4, c.Y+2), 0.5},
		Capsule{V(c.X-6.4, y+0.4, c.Y-2), V(c.X-6.4, y+0.4, c.Y+2), 0.5},
		Capsule{V(c.X-5.95, y+1.1, c.Y-1.6), V(c.X-5.95, y+1.1, c.Y+1.6), 0.5})
	g.Use(Slate).Fill(Cylinder{c.Add(P(r+1.6, 2.6)), 0.7, 0, y, y + 1.4})
}

// corral is a round pole corral with three horses.
func (m MonumentValley) corral(g *Grid, t *monumentValleyTerrain, c Vec2) {
	y := t.H(c.X, c.Y)
	const r = 6.2
	g.Paint(Sand, Cylinder{c, r - 0.3, 0, y - 0.6, y + 0.1})
	gate := func(a float64) bool { // Gate opening toward the hogan (-X).
		d := math.Remainder(a-math.Pi, 2*math.Pi)
		return math.Abs(d) < 0.22
	}
	g.Use(Wood)
	for k := 0; k < 18; k++ {
		a := float64(k) * 2 * math.Pi / 18
		if gate(a) {
			continue
		}
		g.Fill(Cylinder{Polar(c, r, a), 0.4, 0, y - 0.3, y + 2.3})
	}
	rails := Func{Box{V(c.X-r-1, y, c.Y-r-1), V(c.X+r+1, y+2.4, c.Y+r+1)}, func(p Vec3) bool {
		d := p.Plan().Sub(c)
		l := d.Length()
		if l < r-0.4 || l > r+0.4 || gate(d.Angle()) {
			return false
		}
		yy := p.Y - y
		return (yy > 0.9 && yy < 1.35) || (yy > 1.85 && yy < 2.3)
	}}
	g.Use(Wood).Fill(rails)
	m.horse(g, V(c.X-1.5, y, c.Y-2), 0.4, Wood, Charcoal)
	m.horse(g, V(c.X+2, y, c.Y+1.5), 2.6, White, Wood)
	m.horse(g, V(c.X-1.8, y, c.Y+2.6), -0.6, Charcoal, Charcoal)
	// Feed trough.
	g.Use(Wood).Fill(Block(V(c.X+3.3, y, c.Y-3.6), V(c.X+4.3, y+0.9, c.Y-1.0)))
	g.Paint(Grass, Block(V(c.X+3.3, y+0.6, c.Y-3.6), V(c.X+4.3, y+0.9, c.Y-1.0)))
}

// horse stands at base facing angle (radians, plan direction), with mane/tail color.
func (MonumentValley) horse(g *Grid, base Vec3, angle float64, coat, mane Color) {
	piv := base.Plan()
	at := func(x, y, z float64) Vec3 { return V(base.X+x, base.Y+y, base.Z+z) }
	rot := func(s Solid) Solid { return RotateY(s, piv, angle) }
	g.Use(coat)
	g.Fill(rot(Union(
		Ellipsoid{at(0, 1.65, 0), V(1.15, 0.55, 0.45)},
		Capsule{at(0.75, 1.2, 0.3), at(0.75, 0.1, 0.3), 0.3},
		Capsule{at(0.75, 1.2, -0.3), at(0.75, 0.1, -0.3), 0.3},
		Capsule{at(-0.75, 1.2, 0.3), at(-0.75, 0.1, 0.3), 0.3},
		Capsule{at(-0.75, 1.2, -0.3), at(-0.75, 0.1, -0.3), 0.3},
		Capsule{at(0.9, 1.9, 0), at(1.35, 2.65, 0), 0.35},
		Capsule{at(1.4, 2.75, 0), at(1.95, 2.35, 0), 0.3},
	)))
	g.Use(mane).Fill(rot(Union(
		Capsule{at(-1.1, 1.9, 0), at(-1.45, 1.0, 0), 0.25},
		Capsule{at(0.95, 2.2, 0), at(1.3, 2.9, 0), 0.22},
	)))
}

// monumentValleyTrack is the dirt track from the highway to the hogan and corral.
var monumentValleyTrack = []Vec2{P(-7, 1), P(2, 2.5), P(10, 6), P(15, 9.5), P(17.5, 10)}

func monumentValleyTrackDist(x, z float64) float64 {
	best := math.Inf(1)
	p := P(x, z)
	for i := 0; i+1 < len(monumentValleyTrack); i++ {
		a, b := monumentValleyTrack[i], monumentValleyTrack[i+1]
		ab := b.Sub(a)
		u := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
		best = math.Min(best, p.Sub(a.Add(ab.Scale(u))).Length())
	}
	return best
}

func (MonumentValley) track(g *Grid, t *monumentValleyTerrain) {
	box := Box{V(-8, -1, -2), V(20, 6, 14)}
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		return monumentValleyTrackDist(p.X, p.Z) < 1.3 && p.Y > t.H(p.X, p.Z)-0.9
	}})
	g.Paint(Stone, Func{box, func(p Vec3) bool { // Pale wheel ruts.
		d := monumentValleyTrackDist(p.X, p.Z)
		return d > 0.45 && d < 0.9 && p.Y > t.H(p.X, p.Z)-0.5
	}})
}

// boulders are fallen sandstone blocks around the talus feet and on the floor.
func (MonumentValley) boulders(g *Grid, t *monumentValleyTerrain) {
	for i := 0; i < 40; i++ {
		x := monumentValleyX0 + 6 + Hash01(i, 400)*(monumentValleyX1-monumentValleyX0-12)
		z := monumentValleyZ0 + 6 + Hash01(i, 401)*(monumentValleyZ1-monumentValleyZ0-12)
		if math.Abs(x-monumentValleyRoadX) < monumentValleyRoadHalf+3 || monumentValleyTrackDist(x, z) < 3 ||
			P(x, z).Sub(monumentValleyHogan).Length() < 9 || P(x, z).Sub(monumentValleyCorral).Length() < 9 {
			continue
		}
		h := t.H(x, z)
		if h-t.base(x, z) > 6 || t.dune(x, z) > 0.6 {
			continue
		}
		r := 0.9 + 0.9*Hash01(i, 402)
		col := Brick
		if Hash01(i, 403) < 0.4 {
			col = Terracotta
		}
		g.Use(col).FillVolume(RotateY(Ellipsoid{V(x, h+0.2, z), V(r*1.3, r*0.8, r)}, P(x, z), Hash01(i, 404)*math.Pi))
	}
}

// pickup is a red pickup truck parked by the hogan, facing +X.
func (MonumentValley) pickup(g *Grid, t *monumentValleyTerrain, c Vec2) {
	y := t.H(c.X, c.Y) + 0.1
	z0, z1 := c.Y-1.15, c.Y+1.15
	g.Use(Charcoal)
	for _, wx := range []float64{c.X - 1.6, c.X + 1.6} {
		g.Fill(Capsule{V(wx, y+0.5, z0+0.1), V(wx, y+0.5, z1-0.1), 0.5})
	}
	g.Use(Red).Fill(Block(V(c.X-2.7, y+0.55, z0), V(c.X+2.7, y+1.35, z1)))
	g.Fill(Block(V(c.X+0.1, y+1.35, z0+0.1), V(c.X+1.6, y+2.3, z1-0.1)))            // Cab.
	g.Use(Navy).Fill(Block(V(c.X+1.6, y+1.45, z0+0.2), V(c.X+1.75, y+2.1, z1-0.2))) // Windshield.
	g.Use(Red).Fill(Block(V(c.X-2.7, y+1.35, z0), V(c.X-0.1, y+1.8, z0+0.35)),
		Block(V(c.X-2.7, y+1.35, z1-0.35), V(c.X-0.1, y+1.8, z1)))
	g.Use(Wood).Fill(Block(V(c.X-2.4, y+1.35, z0+0.5), V(c.X-1.2, y+1.8, z1-0.5))) // Firewood load.
	g.Use(Stone).Fill(Block(V(c.X+2.7, y+0.5, z0), V(c.X+3.0, y+0.9, z1)))
}

// sheep is a small Navajo-Churro flock grazing south of the hogan.
func (MonumentValley) sheep(g *Grid, t *monumentValleyTerrain) {
	for i := 0; i < 9; i++ {
		x := 6 + 9*Hash01(i, 500)
		z := 17 + 7*Hash01(i, 501)
		y := t.H(x, z)
		a := Hash01(i, 502) * 2 * math.Pi
		d := P(math.Cos(a), math.Sin(a))
		wool := White
		if Hash01(i, 503) < 0.25 {
			wool = Charcoal
		}
		g.Use(wool).Fill(Ellipsoid{V(x, y+0.75, z), V(0.75, 0.5, 0.75)})
		g.Use(Charcoal).Fill(Sphere(V(x+d.X*0.8, y+1.0, z+d.Y*0.8), 0.35))
	}
}
