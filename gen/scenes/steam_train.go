package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// SteamTrain is an express hauled by a blue Pacific steam locomotive crossing a tall,
// curved concrete railway viaduct over a highland glen. The track follows a circular
// curve bowing toward +Z (the hero side); the viaduct's slender arches stride over a
// river that winds from the back of the glen to a loch in the foreground. Heather and
// pine covered hills rise on both sides; the track runs through cuttings into tunnel
// portals in the hills. A memorial tower stands on a promontory at the loch head, with
// whitewashed cottages, a kirk, a road bridge, drystone dykes, sheep and rowing boats.
//
// Railway parts are authored in curve coordinates (s along the track, n across it,
// positive outward, y up) and mapped onto the curve by steamTrainBent. The train heads
// toward +s (+X); its plume trails back along the carriages.
type SteamTrain struct{}

func init() { Register(SteamTrain{}) }

func (SteamTrain) Name() string  { return "steam_train" }
func (SteamTrain) Title() string { return "Steam Train" }

const (
	steamTrainR      = 72.0  // Track curve radius.
	steamTrainCZ     = -72.0 // Curve center z; the track crosses x = 0 at z = 0.
	steamTrainEnd    = 46.0  // Viaduct half length along s.
	steamTrainArches = 14
	steamTrainPortal = 50.0 // Tunnel portals at s = ±50.
	steamTrainBed    = 20.0 // Deck / formation level under the ballast.
	steamTrainRail   = 21.0 // Rail top.
	steamTrainSpring = 14.6 // Arch springing height.
	steamTrainHalfX  = 62.0 // Site half width.
	steamTrainBack   = -38.0
	steamTrainFront  = 35.0
	steamTrainNose   = 41.0 // Locomotive front (s).
)

// steamTrainMonument is the memorial tower's plan position.
var steamTrainMonument = P(-17, 18)

// steamTrainBent maps a solid authored in curve coordinates (X = s, Y = y, Z = n) onto
// the track curve.
type steamTrainBent struct{ S Solid }

func steamTrainLocal(x, z float64) (s, n float64) {
	dx, dz := x, z-steamTrainCZ
	return steamTrainR * math.Atan2(dx, dz), math.Hypot(dx, dz) - steamTrainR
}

func steamTrainAt(s, n float64) Vec2 {
	th, r := s/steamTrainR, steamTrainR+n
	return P(r*math.Sin(th), steamTrainCZ+r*math.Cos(th))
}

func (b steamTrainBent) Contains(p Vec3) bool {
	s, n := steamTrainLocal(p.X, p.Z)
	return b.S.Contains(V(s, p.Y, n))
}

func (b steamTrainBent) Bounds() Box {
	lb := b.S.Bounds()
	out := EmptyBox()
	const steps = 24
	for i := 0; i <= steps; i++ {
		s := lb.Min.X + (lb.Max.X-lb.Min.X)*float64(i)/steps
		for _, n := range []float64{lb.Min.Z, lb.Max.Z} {
			q := steamTrainAt(s, n)
			out = out.Union(Box{q.At(lb.Min.Y), q.At(lb.Max.Y)})
		}
	}
	return out.Grow(0.6)
}

func steamTrainSmooth(a, b, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-a)/(b-a)))
	return t * t * (3 - 2*t)
}

func steamTrainNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	lerp := func(a, b, t float64) float64 { return a + (b-a)*t }
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return lerp(lerp(a, b, sx), lerp(c, d, sx), sz)*2 - 1
}

func steamTrainFBM(x, z float64, seed int) float64 {
	return 0.6*steamTrainNoise(x, z, seed) + 0.3*steamTrainNoise(x*2.03+17, z*2.03+5, seed+1) +
		0.1*steamTrainNoise(x*4.1+3, z*4.1+29, seed+2)
}

// Local-frame shapes (X = s, Y = y, Z = n).
func steamTrainBox(s0, y0, n0, s1, y1, n1 float64) Solid { return Block(V(s0, y0, n0), V(s1, y1, n1)) }

// steamTrainTube is a cylinder lying along the track.
func steamTrainTube(s0, s1, yc, nc, r float64) Solid {
	return Func{Box{V(s0, yc-r, nc-r), V(s1, yc+r, nc+r)}, func(p Vec3) bool {
		return (p.Y-yc)*(p.Y-yc)+(p.Z-nc)*(p.Z-nc) <= r*r
	}}
}

// steamTrainDisc is a wheel: a disc in the s-y plane.
func steamTrainDisc(sc, yc, nc, r, t float64) Solid {
	return Func{Box{V(sc-r, yc-r, nc-t/2), V(sc+r, yc+r, nc+t/2)}, func(p Vec3) bool {
		return (p.X-sc)*(p.X-sc)+(p.Y-yc)*(p.Y-yc) <= r*r
	}}
}

// steamTrainPost is an upright cylinder.
func steamTrainPost(sc, nc, r, y0, y1 float64) Solid {
	return Func{Box{V(sc-r, y0, nc-r), V(sc+r, y1, nc+r)}, func(p Vec3) bool {
		return (p.X-sc)*(p.X-sc)+(p.Z-nc)*(p.Z-nc) <= r*r
	}}
}

// waterLevel is the river surface: it climbs gently toward the back of the glen.
func (SteamTrain) waterLevel(z float64) float64 { return 1.0 + 0.055*math.Max(0, -z-2) }

// riverX is the river's centerline.
func (SteamTrain) riverX(z float64) float64 {
	return 2*math.Sin(z*0.045) + 2.6*math.Sin(z*0.11+0.8)
}

// height is the terrain surface; it is negative outside the site.
func (t SteamTrain) height(x, z float64) float64 {
	ex := steamTrainHalfX - math.Abs(x)
	ez := math.Min(z-steamTrainBack, steamTrainFront-z)
	if ex < 0 || ez < 0 {
		return -100
	}
	xg := 2 * math.Sin(z*0.045)
	u := math.Abs(x - xg)
	floor := 2.0 + 0.055*math.Max(0, -z-2)
	floor += 0.35 * steamTrainFBM(x*0.15, z*0.15, 3)
	// The loch: an irregular basin in front of the viaduct.
	lq := math.Hypot((x-3)/30, (z-25)/13) + 0.12*steamTrainFBM(x*0.08, z*0.08, 17)
	floor -= 2.6 * steamTrainSmooth(1.05, 0.8, lq)
	// Ridges on both sides: crests near the tunnel portals, easing toward the rims.
	crest := 47 + 2*steamTrainFBM(z*0.05, 0, 19)
	hill := 31 * steamTrainSmooth(18, crest, u) * (1 - 0.92*steamTrainSmooth(crest-1, crest+16, u))
	hill *= 1 - 0.75*steamTrainSmooth(-4, 32, z)
	hill *= 1 - 0.8*steamTrainSmooth(-16, -36, z)
	// The shoulders in front of the track fall away so the train stays in view.
	_, tn := steamTrainLocal(x, z)
	hill *= 1 - 0.6*steamTrainSmooth(5, 26, tn)
	hill += 3.0 * steamTrainFBM(x*0.07, z*0.07, 11) * steamTrainSmooth(16, 34, u)
	h := floor + math.Max(hill, 0)
	// River channel.
	if z < 14 {
		d := math.Abs(x - t.riverX(z))
		h = math.Min(h, t.waterLevel(z)-0.7+math.Max(0, d-1.8)*0.9)
	}
	// Promontory for the memorial at the loch head.
	if d := P(x, z).Sub(steamTrainMonument).Length(); d < 10 {
		h = math.Max(h, 2.0-math.Max(0, d-4.5)*0.45)
	}
	// Rim: everything eases down to a low edge.
	e := math.Min(ex, ez)
	h = 1.2 + (h-1.2)*steamTrainSmooth(0, 7, e)
	// Track formation: cuttings and embankments beyond the viaduct.
	s, n := steamTrainLocal(x, z)
	if as := math.Abs(s); as >= steamTrainEnd-1 && as <= steamTrainPortal+0.5 {
		w := math.Abs(n) - 3.6
		if w < 0 {
			h = steamTrainBed
		} else {
			h = math.Max(steamTrainBed-w, math.Min(steamTrainBed+w*1.1, h))
		}
	}
	return h
}

func (t SteamTrain) Build(g *Grid) Order {
	t.terrain(g)
	t.water(g)
	t.viaduct(g)
	t.portals(g)
	t.monument(g)
	t.settlement(g)
	t.trees(g)
	t.track(g)
	t.locomotive(g)
	t.tender(g)
	sT := steamTrainNose - 29.4
	const L, gap = 15.0, 0.9
	for i := 0; i < 4; i++ {
		s1 := sT - gap - float64(i)*(L+gap)
		t.carriage(g, s1-L, s1, i == 3)
	}
	t.steam(g)
	return Along(P(1, 0))
}

// ---------------------------------------------------------------------------------
// Terrain.

// terrain fills the height field as a skin: only cells with an empty face neighbor are
// kept (space below y = 0 counts as filled), so hills cost their surface, not volume.
func (t SteamTrain) terrain(g *Grid) {
	step := g.Span() / 2
	x0 := int(math.Floor(-steamTrainHalfX/step)) - 3
	x1 := int(math.Ceil(steamTrainHalfX/step)) + 3
	z0 := int(math.Floor(steamTrainBack/step)) - 3
	z1 := int(math.Ceil(steamTrainFront/step)) + 3
	w := x1 - x0 + 1
	cols := make([]float64, w*(z1-z0+1))
	for iz := z0; iz <= z1; iz++ {
		for ix := x0; ix <= x1; ix++ {
			cols[(iz-z0)*w+ix-x0] = t.height(float64(ix)*step, float64(iz)*step)
		}
	}
	hAt := func(ix, iz int) float64 {
		if ix < x0 || ix > x1 || iz < z0 || iz > z1 {
			return -100
		}
		return cols[(iz-z0)*w+ix-x0]
	}
	occ := func(ix, iy, iz int) bool {
		return iy < 0 || float64(iy)*step <= hAt(ix, iz)
	}
	for iz := z0; iz <= z1; iz++ {
		for ix := x0; ix <= x1; ix++ {
			if (ix^iz)&1 != 0 {
				continue
			}
			h := hAt(ix, iz)
			if h < 0 {
				continue
			}
			gx := (hAt(ix+2, iz) - hAt(ix-2, iz)) / (4 * step)
			gz := (hAt(ix, iz+2) - hAt(ix, iz-2)) / (4 * step)
			slope := math.Hypot(gx, gz)
			x, z := float64(ix)*step, float64(iz)*step
			for iy := ix & 1; float64(iy)*step <= h; iy += 2 {
				exposed := false
				for _, d := range Neighbors {
					if !occ(ix+int(d.X), iy+int(d.Y), iz+int(d.Z)) {
						exposed = true
						break
					}
				}
				if !exposed {
					continue
				}
				y := float64(iy) * step
				g.Use(t.groundColor(x, y, z, h, slope, ix, iy, iz))
				g.Add(Cell{int32(ix), int32(iy), int32(iz)})
			}
		}
	}
}

func (t SteamTrain) groundColor(x, y, z, h, slope float64, ix, iy, iz int) Color {
	r := Hash01(ix, iy, iz, 41)
	s, n := steamTrainLocal(x, z)
	inCut := math.Abs(s) >= steamTrainEnd-1 && math.Abs(s) <= steamTrainPortal+0.5
	switch {
	case inCut && math.Abs(n) < 3.7:
		return DarkStone
	case h-y > 1.4+1.4*slope:
		if r < 0.3 {
			return Stone
		}
		return DarkStone
	case slope > 3.2 && !inCut:
		if r < 0.25 {
			return Stone
		}
		return DarkStone
	case slope > 1.5 && !inCut && r < 0.08+0.14*steamTrainSmooth(1.5, 3.2, slope):
		return DarkStone
	case h < t.waterLevel(z)+0.55 && z > 8, h < t.waterLevel(z)+0.25 && (z > 4 || math.Abs(x-t.riverX(z)) < 4):
		return Sand
	}
	heather := steamTrainFBM(x*0.09+40, z*0.09, 23) + 0.25*steamTrainSmooth(6, 20, h)
	if h > 4 && heather > 0.32 {
		switch {
		case r < 0.5:
			return Terracotta
		case r < 0.68:
			return Red
		case r < 0.85:
			return Foliage
		}
		return Grass
	}
	if steamTrainFBM(x*0.12, z*0.12+30, 29) > 0.15 && h > 3 {
		if r < 0.75 {
			return Foliage
		}
		return Grass
	}
	if r < 0.12 {
		return Foliage
	}
	return Grass
}

// water fills the river and the loch up to the water level.
func (t SteamTrain) water(g *Grid) {
	g.Use(Water).Fill(Func{Box{V(-steamTrainHalfX, 0, steamTrainBack), V(steamTrainHalfX, 4.5, steamTrainFront)}, func(p Vec3) bool {
		h := t.height(p.X, p.Z)
		near := p.Z > 4 || math.Abs(p.X-t.riverX(p.Z)) < 3.5
		return near && h > -50 && p.Y > h && p.Y <= t.waterLevel(p.Z)
	}})
}

// ---------------------------------------------------------------------------------
// Viaduct and track.

func (t SteamTrain) groundUnder(s, n float64) float64 {
	q := steamTrainAt(s, n)
	return t.height(q.X, q.Y)
}

func (t SteamTrain) viaduct(g *Grid) {
	const pw = 1.9 // Pier width along the track at the springing.
	sp := 2 * steamTrainEnd / steamTrainArches
	a0 := (sp - pw) / 2
	opening := func(s, y float64) (bool, float64, float64) {
		j := math.Floor((s + steamTrainEnd) / sp)
		if j < 0 || j >= steamTrainArches {
			return false, 0, 0
		}
		c := -steamTrainEnd + (j+0.5)*sp
		bat := math.Max(0, steamTrainSpring-y)
		a := a0 - 0.035*bat
		ds := s - c
		dy := y - steamTrainSpring
		if math.Abs(ds) < a && (dy < 0 || ds*ds+dy*dy < a*a) {
			return true, ds, a
		}
		return false, ds, a
	}
	body := Func{Box{V(-steamTrainEnd-1.6, 0, -4.2), V(steamTrainEnd+1.6, steamTrainBed, 4.2)}, func(p Vec3) bool {
		s, y, n := p.X, p.Y, p.Z
		bat := math.Max(0, steamTrainSpring-y)
		if math.Abs(n) > 3.0+0.05*bat {
			return false
		}
		if y < t.groundUnder(s, 0)-1.5 {
			return false
		}
		in, _, _ := opening(s, y)
		return !in
	}}
	g.Use(Stone).Fill(steamTrainBent{body})
	// Weathering: darker patches on the piers and damp feet.
	g.Paint(DarkStone, steamTrainBent{Func{body.Box, func(p Vec3) bool {
		if !body.F(p) {
			return false
		}
		ground := t.groundUnder(p.X, 0)
		if p.Y < ground+2.5 {
			return Hash01(int(math.Floor(p.X*1.3)), int(math.Floor(p.Y*1.6)), int(math.Floor(p.Z*1.3)), 5) < 0.6
		}
		return Hash01(int(math.Floor(p.X/1.2)), int(math.Floor(p.Y/0.9)), int(math.Floor(p.Z/1.4)), 7) < 0.14
	}}})
	// Arch rings (voussoirs).
	g.Paint(DarkStone, steamTrainBent{Func{body.Box, func(p Vec3) bool {
		if p.Y < steamTrainSpring {
			return false
		}
		_, ds, a := opening(p.X, p.Y)
		r := math.Hypot(ds, p.Y-steamTrainSpring)
		return r >= a && r < a+0.75
	}}})
	// String course, parapets and copings.
	ends := steamTrainEnd + 1.6
	g.Use(DarkStone).Fill(steamTrainBent{steamTrainBox(-ends, 18.8, -3.5, ends, 19.4, 3.5)})
	g.Use(Stone).Fill(steamTrainBent{steamTrainBox(-ends, steamTrainBed, 2.5, ends, 21.0, 3.2)},
		steamTrainBent{steamTrainBox(-ends, steamTrainBed, -3.2, ends, 21.0, -2.5)})
	g.Use(White).Fill(steamTrainBent{steamTrainBox(-ends, 21.0, 2.5, ends, 21.4, 3.3)},
		steamTrainBent{steamTrainBox(-ends, 21.0, -3.3, ends, 21.4, -2.5)})
	// Pier refuges: the pilasters run up the faces of every pier.
	for k := 0; k <= steamTrainArches; k++ {
		s := -steamTrainEnd + float64(k)*sp
		for _, side := range []float64{-1, 1} {
			g.Use(Stone).Fill(steamTrainBent{Func{Box{V(s-0.8, 0, -4.4), V(s+0.8, 18.8, 4.4)}, func(p Vec3) bool {
				bat := math.Max(0, steamTrainSpring-p.Y)
				hw := 3.0 + 0.05*bat
				return p.Z*side > hw-0.2 && p.Z*side < hw+0.5 && p.Y > t.groundUnder(p.X, 0)-1.0
			}}})
		}
	}
}

func (t SteamTrain) track(g *Grid) {
	lo, hi := -steamTrainPortal, steamTrainPortal
	g.Use(DarkStone).Fill(steamTrainBent{steamTrainBox(lo, steamTrainBed, -2.3, hi, 20.4, 2.3)})
	g.Use(Wood)
	for s := math.Floor(lo/1.6) * 1.6; s < hi; s += 1.6 {
		g.Fill(steamTrainBent{steamTrainBox(s, 20.4, -1.7, s+0.8, 20.6, 1.7)})
	}
	g.Use(Slate).Fill(steamTrainBent{steamTrainBox(lo, 20.5, -1.35, hi, steamTrainRail, -0.9)},
		steamTrainBent{steamTrainBox(lo, 20.5, 0.9, hi, steamTrainRail, 1.35)})
}

// portals are masonry tunnel mouths at both ends of the cuttings.
func (t SteamTrain) portals(g *Grid) {
	for _, dir := range []float64{-1, 1} {
		s0 := dir * steamTrainPortal
		a, b := s0, s0+dir*1.2
		lo, hi := math.Min(a, b), math.Max(a, b)
		wall := Func{Box{V(lo, steamTrainBed-1, -7.5), V(hi, 31, 7.5)}, func(p Vec3) bool {
			return p.Y <= 29.5-0.25*math.Abs(p.Z)
		}}
		g.Use(Stone).Fill(steamTrainBent{wall})
		g.Paint(DarkStone, steamTrainBent{Func{wall.Box, func(p Vec3) bool {
			return wall.F(p) && Hash01(int(math.Floor(p.Y/0.8)), int(math.Floor(p.Z/1.3)), int(dir), 9) < 0.2
		}}})
		// Coping and opening.
		g.Use(DarkStone).Fill(steamTrainBent{Func{Box{V(lo-0.2, 27, -7.6), V(hi+0.2, 30.2, 7.6)}, func(p Vec3) bool {
			top := 29.5 - 0.25*math.Abs(p.Z)
			return p.Y > top-0.6 && p.Y <= top+0.4
		}}})
		open := Func{Box{V(lo-0.5, steamTrainBed, -2.8), V(hi+0.5, 27, 2.8)}, func(p Vec3) bool {
			dy := p.Y - 24.0
			return math.Abs(p.Z) <= 2.8 && (dy <= 0 || p.Z*p.Z+dy*dy <= 2.8*2.8)
		}}
		g.Carve(steamTrainBent{open})
		// Arch ring around the mouth.
		g.Paint(DarkStone, steamTrainBent{Func{Box{V(lo, 20, -4), V(hi, 28, 4)}, func(p Vec3) bool {
			dy := p.Y - 24.0
			r := math.Hypot(p.Z, math.Max(dy, 0))
			return r < 3.7
		}}})
		// Darkness inside the bore.
		da, db := s0+dir*1.2, s0+dir*2.2
		g.Use(Charcoal).Fill(steamTrainBent{Func{Box{V(math.Min(da, db), steamTrainBed, -2.8), V(math.Max(da, db), 27, 2.8)}, func(p Vec3) bool {
			return open.F(p)
		}}})
		// Track runs into the bore.
		g.Use(DarkStone).Fill(steamTrainBent{steamTrainBox(lo-0.5, steamTrainBed, -2.3, hi+0.5, 20.4, 2.3)})
	}
}

// ---------------------------------------------------------------------------------
// Monument and trees.

func (t SteamTrain) monument(g *Grid) {
	c := steamTrainMonument
	base := 1.6
	// Enclosure wall and plinth.
	g.Use(Stone).Fill(Cylinder{Center: c, Radius: 4.6, Inner: 3.8, Y0: base, Y1: base + 1.2})
	g.Use(Sand).Fill(Cylinder{Center: c, Radius: 3.8, Y0: base, Y1: base + 0.5})
	g.Use(Stone).Fill(Cylinder{Center: c, Radius: 2.1, Y0: base, Y1: base + 1.4})
	// Tower with a doorway and a crenellated parapet.
	g.Use(Stone).Fill(Cylinder{Center: c, Radius: 1.5, Inner: 0.6, Y0: base + 1.4, Y1: base + 12})
	g.Use(Charcoal).Fill(Cylinder{Center: c, Radius: 0.8, Y0: base + 1.4, Y1: base + 12})
	g.Carve(Block(V(c.X-0.5, base+1.4, c.Y), V(c.X+0.5, base+3.2, c.Y+2)))
	g.Use(DarkStone).Fill(Cylinder{Center: c, Radius: 1.9, Y0: base + 11.6, Y1: base + 12.3})
	g.Use(DarkStone).Fill(Func{Box{V(c.X-2, base+12.3, c.Y-2), V(c.X+2, base+13.1, c.Y+2)}, func(p Vec3) bool {
		d := P(p.X, p.Z).Sub(c)
		r := d.Length()
		k := int(math.Floor((d.Angle() + math.Pi) / (2 * math.Pi) * 10))
		return r <= 1.9 && r >= 1.2 && k%2 == 0
	}})
	// The highlander on top.
	g.Use(Stone).Fill(Cylinder{Center: c, Radius: 0.6, Y0: base + 12.3, Y1: base + 14.2}, Sphere(c.At(base+14.6), 0.5))
	g.Paint(DarkStone, Cylinder{Center: c, Radius: 0.7, Y0: base + 12.3, Y1: base + 13.1})
}

func (t SteamTrain) tree(g *Grid, c Vec2, base, height float64, seed int) {
	trunkTop := base + height*0.7
	g.Use(Wood).Fill(Capsule{c.At(base - 1.5), c.At(trunkTop), 0.42})
	g.Paint(Terracotta, Cylinder{Center: c, Radius: 0.5, Y0: base + height*0.4, Y1: trunkTop})
	g.Use(Foliage)
	for k := 0; k < 3; k++ {
		a := Hash01(seed, k, 1) * 2 * math.Pi
		off := Polar(P(0, 0), 0.6+0.9*Hash01(seed, k, 2), a)
		y := trunkTop - 0.4 + float64(k)*0.7
		r := 1.5 + 0.8*Hash01(seed, k, 3) - 0.25*float64(k)
		g.FillVolume(Ellipsoid{c.Add(off).At(y), V(r, r*0.55, r)})
	}
}

func (t SteamTrain) trees(g *Grid) {
	const cell = 4.6
	for i := -14; i <= 14; i++ {
		for j := -9; j <= 8; j++ {
			jx, jz := Hash01(i, j, 61), Hash01(i, j, 62)
			x := (float64(i) + jx - 0.5) * cell
			z := (float64(j) + jz - 0.5) * cell
			h := t.height(x, z)
			if h < 0 || math.Abs(x) > steamTrainHalfX-4 || z < steamTrainBack+4 || z > steamTrainFront-4 {
				continue
			}
			if h < t.waterLevel(z)+0.8 {
				continue
			}
			s, n := steamTrainLocal(x, z)
			if n > -8 && n < 14 && math.Abs(s) < steamTrainPortal+3 {
				continue
			}
			if math.Abs(x-t.riverX(z)) < 3.5 || P(x, z).Sub(steamTrainMonument).Length() < 7 {
				continue
			}
			sl := math.Hypot(t.height(x+1, z)-t.height(x-1, z), t.height(x, z+1)-t.height(x, z-1)) / 2
			if sl > 1.7 || t.nearBuildings(x, z, 5) {
				continue
			}
			// Pines gather in woods on the lower hillsides.
			dens := 0.35 + 0.6*steamTrainSmooth(-0.1, 0.35, steamTrainFBM(x*0.05+7, z*0.05, 71))
			if Hash01(i, j, 63) > dens {
				continue
			}
			t.tree(g, P(x, z), h, 6.5+3*Hash01(i, j, 64), i*31+j)
		}
	}
}

// ---------------------------------------------------------------------------------
// Train.

func (t SteamTrain) fill(g *Grid, c Color, solids ...Solid) {
	g.Use(c)
	for _, s := range solids {
		g.Fill(steamTrainBent{s})
	}
}

func (t SteamTrain) paint(g *Grid, c Color, solids ...Solid) {
	for _, s := range solids {
		g.Paint(c, steamTrainBent{s})
	}
}

// spokes paints the gaps between a wheel's spokes dark.
func (t SteamTrain) spokes(g *Grid, sc, yc, nc, r, th float64, count int) {
	t.paint(g, Charcoal, Func{Box{V(sc-r, yc-r, nc-th/2), V(sc+r, yc+r, nc+th/2)}, func(p Vec3) bool {
		ds, dy := p.X-sc, p.Y-yc
		d := math.Hypot(ds, dy)
		if d < r*0.36 || d > r*0.74 {
			return false
		}
		sector := 2 * math.Pi / float64(count)
		a := math.Mod(math.Atan2(dy, ds)+2*math.Pi, sector)
		return math.Min(a, sector-a)*d > 0.22
	}})
}

func (t SteamTrain) locomotive(g *Grid) {
	f := steamTrainNose
	const rail = steamTrainRail
	const by, br = 26.2, 1.55 // Boiler axis height and radius.
	// Frames and running plate.
	t.fill(g, Charcoal, steamTrainBox(f-19.6, 21.8, -0.9, f-1.0, 24.3, 0.9))
	t.fill(g, Charcoal, steamTrainBox(f-20.0, 24.2, -2.15, f-0.9, 24.6, 2.15))
	// Boiler, firebox and smokebox.
	t.fill(g, Navy, steamTrainTube(f-15.6, f-4.0, by, 0, br))
	t.fill(g, Navy, steamTrainBox(f-15.6, 24.6, -1.7, f-12.2, 27.3, 1.7))
	t.paint(g, Gold,
		steamTrainBox(f-5.0, 24.6, -2, f-4.5, 28, 2), steamTrainBox(f-8.0, 24.6, -2, f-7.5, 28, 2),
		steamTrainBox(f-11.4, 24.6, -2, f-10.9, 28, 2), steamTrainBox(f-15.6, 24.6, -2, f-15.1, 28, 2))
	t.fill(g, Charcoal, steamTrainTube(f-4.0, f-1.0, by, 0, br+0.15), steamTrainBox(f-4.0, 24.6, -1.6, f-1.0, by, 1.6))
	t.fill(g, Charcoal, steamTrainDisc(f-0.9, by, 0, 1.35, 0.5))
	t.fill(g, Stone, steamTrainBox(f-0.8, by-0.3, -0.3, f-0.5, by+0.3, 0.3))
	// Chimney, dome and safety valves.
	t.fill(g, Charcoal, steamTrainPost(f-2.4, 0, 0.75, by, 29.0))
	t.fill(g, Gold, steamTrainPost(f-2.4, 0, 0.95, 28.7, 29.3))
	t.fill(g, Gold, steamTrainPost(f-8.8, 0, 0.85, by, 28.1))
	t.fill(g, Gold, Ellipsoid{V(f-8.8, 28.1, 0), V(0.85, 0.5, 0.85)})
	t.fill(g, Gold, steamTrainPost(f-13.6, 0, 0.45, by, 28.4))
	// Cylinders, buffer beam, buffers and lamps.
	for _, side := range []float64{-1, 1} {
		t.fill(g, Navy, steamTrainTube(f-6.4, f-3.4, 23.0, side*1.75, 0.75))
	}
	t.fill(g, Red, steamTrainBox(f-1.0, 21.9, -2.2, f-0.2, 23.8, 2.2))
	for _, side := range []float64{-1, 1} {
		t.fill(g, Charcoal, steamTrainTube(f-0.2, f+0.5, 22.85, side*1.45, 0.35))
		t.fill(g, Stone, steamTrainTube(f+0.5, f+0.8, 22.85, side*1.45, 0.5))
		t.fill(g, White, steamTrainBox(f-0.9, 23.8, side*1.5-0.3, f-0.3, 24.4, side*1.5+0.3))
	}
	t.fill(g, White, steamTrainBox(f-1.9, 27.8, -0.3, f-1.3, 28.4, 0.3))
	// Cab with side windows, open back and a dark interior.
	t.fill(g, Navy, steamTrainBox(f-20.0, 24.6, -2.15, f-15.6, 28.7, 2.15))
	g.Carve(steamTrainBent{steamTrainBox(f-20.5, 24.6, -1.4, f-16.3, 28.4, 1.4)})
	g.Carve(steamTrainBent{steamTrainBox(f-18.4, 26.3, -3, f-16.6, 27.9, 3)})
	t.paint(g, Charcoal, steamTrainBox(f-20.5, 24.2, -1.9, f-15.9, 28.6, 1.9))
	t.paint(g, Navy, steamTrainBox(f-20.5, 24.6, -2.2, f-15.5, 28.8, -1.8), steamTrainBox(f-20.5, 24.6, 1.8, f-15.5, 28.8, 2.2))
	t.fill(g, Glass, steamTrainBox(f-15.9, 27.4, -1.5, f-15.6, 28.2, -0.7), steamTrainBox(f-15.9, 27.4, 0.7, f-15.6, 28.2, 1.5))
	t.fill(g, Charcoal, Func{Box{V(f-20.6, 28.7, -2.45), V(f-15.2, 29.5, 2.45)}, func(p Vec3) bool {
		return p.Y <= 29.5-0.12*p.Z*p.Z
	}})
	t.paint(g, Gold, steamTrainBox(f-20.0, 25.5, -2.3, f-15.6, 25.8, -2.0), steamTrainBox(f-20.0, 25.5, 2.0, f-15.6, 25.8, 2.3))
	// Wheels: leading bogie, three coupled drivers, trailing truck.
	for _, side := range []float64{-1, 1} {
		n := side * 1.4
		for _, s := range []float64{f - 2.4, f - 4.6} {
			t.fill(g, Red, steamTrainDisc(s, rail+0.85, n, 0.85, 0.55))
		}
		for _, s := range []float64{f - 7.4, f - 10.8, f - 14.2} {
			t.fill(g, Red, steamTrainDisc(s, rail+1.55, n, 1.55, 0.6))
			t.spokes(g, s, rail+1.55, n, 1.55, 0.6, 8)
			t.fill(g, Charcoal, steamTrainTube(s-0.25, s+0.25, rail+1.55, side*1.85, 0.3))
		}
		t.fill(g, Red, steamTrainDisc(f-17.6, rail+1.0, n, 1.0, 0.55))
		// Coupling rod, connecting rod and crosshead slide bars.
		t.fill(g, Stone, steamTrainBox(f-14.5, rail+1.1, side*1.75-0.2, f-7.1, rail+1.5, side*1.75+0.2))
		t.fill(g, Stone, Capsule{V(f-5.6, 23.0, side*2.05), V(f-10.8, rail+1.4, side*2.05), 0.28})
		t.fill(g, Stone, steamTrainBox(f-7.0, 23.2, side*2.05-0.2, f-5.6, 23.5, side*2.05+0.2))
	}
	// Steam escaping from the cylinder drain cocks.
	for _, side := range []float64{-1, 1} {
		t.fill(g, White, Ellipsoid{V(f-3.2, 21.9, side*2.9), V(1.2, 0.7, 0.9)})
	}
}

func (t SteamTrain) tender(g *Grid) {
	f := steamTrainNose
	s0, s1 := f-29.4, f-20.6
	t.fill(g, Charcoal, steamTrainBox(s0+0.3, 21.9, -1.6, s1-0.3, 22.7, 1.6))
	t.fill(g, Navy, steamTrainBox(s0, 22.6, -2.05, s1, 27.0, 2.05))
	t.paint(g, Gold, steamTrainBox(s0, 24.3, -2.2, s1, 24.6, 2.2))
	t.fill(g, Navy, steamTrainBox(s1-1.2, 27.0, -2.05, s1, 27.8, 2.05))
	t.fill(g, Charcoal, Func{Box{V(s0+0.6, 26.8, -1.7), V(s1-1.2, 28.6, 1.7)}, func(p Vec3) bool {
		u := (p.X - (s0+s1-0.6)/2) / ((s1 - s0 - 1.8) / 2)
		top := 27.2 + 1.2*(1-u*u)*(1-0.3*(p.Z*p.Z)/2.9)
		top += 0.3 * Hash01(int(math.Floor(p.X*1.5)), int(math.Floor(p.Z*1.5)), 13)
		return p.Y <= top
	}})
	for _, side := range []float64{-1, 1} {
		for _, s := range []float64{s0 + 1.6, s0 + 4.4, s0 + 7.2} {
			t.fill(g, Red, steamTrainDisc(s, steamTrainRail+0.85, side*1.4, 0.85, 0.55))
		}
	}
}

func (t SteamTrain) carriage(g *Grid, s0, s1 float64, last bool) {
	const lo, top = 22.4, 26.8
	t.fill(g, Charcoal, steamTrainBox(s0+0.4, 21.8, -1.6, s1-0.4, lo, 1.6))
	t.fill(g, Brick, steamTrainBox(s0, lo, -2.05, s1, top, 2.05))
	g.Carve(steamTrainBent{steamTrainBox(s0+0.7, lo+0.5, -1.3, s1-0.7, top-0.3, 1.3)})
	t.paint(g, Charcoal, steamTrainBox(s0+0.3, lo+0.2, -1.7, s1-0.3, top, 1.7))
	t.paint(g, White, steamTrainBox(s0, 24.6, -2.2, s1, top, -1.6), steamTrainBox(s0, 24.6, 1.6, s1, top, 2.2),
		steamTrainBox(s0-0.1, 24.6, -2.2, s0+0.6, top, 2.2), steamTrainBox(s1-0.6, 24.6, -2.2, s1+0.1, top, 2.2))
	t.paint(g, Gold, steamTrainBox(s0, 24.3, -2.2, s1, 24.6, -1.6), steamTrainBox(s0, 24.3, 1.6, s1, 24.6, 2.2))
	for x := s0 + 1.8; x+1.0 < s1-1.4; x += 2.1 {
		t.paint(g, Glass, steamTrainBox(x, 25.0, -2.2, x+1.0, 26.2, -1.6), steamTrainBox(x, 25.0, 1.6, x+1.0, 26.2, 2.2))
	}
	// Doors at the ends.
	for _, d := range []float64{s0 + 0.6, s1 - 1.4} {
		t.paint(g, Charcoal, steamTrainBox(d, 22.8, -2.2, d+0.8, 26.4, -1.6), steamTrainBox(d, 22.8, 1.6, d+0.8, 26.4, 2.2))
	}
	t.fill(g, Charcoal, Func{Box{V(s0-0.1, top, -2.3), V(s1+0.1, top+1.3, 2.3)}, func(p Vec3) bool {
		q := p.Z / 2.3
		return p.Y-top <= 1.2*math.Sqrt(math.Max(0, 1-q*q))
	}})
	for x := s0 + 2.5; x < s1-2; x += 3.2 {
		t.fill(g, DarkStone, steamTrainBox(x, top+1.1, -0.35, x+0.6, top+1.6, 0.35))
	}
	// Gangway to the vehicle ahead.
	t.fill(g, Charcoal, steamTrainBox(s1, 22.6, -1.2, s1+0.9, 26.4, 1.2))
	// Bogies.
	for _, c := range []float64{s0 + 2.6, s1 - 2.6} {
		t.fill(g, Charcoal, steamTrainBox(c-1.8, 21.5, -1.65, c+1.8, 22.0, 1.65))
		for _, side := range []float64{-1, 1} {
			for _, s := range []float64{c - 1.1, c + 1.1} {
				t.fill(g, Charcoal, steamTrainDisc(s, steamTrainRail+0.7, side*1.35, 0.7, 0.5))
			}
		}
	}
	if last {
		t.fill(g, Red, steamTrainBox(s0-0.4, 23.0, -0.4, s0, 23.6, 0.4))
	}
}

// steam is the plume trailing back from the chimney, drifting away from the viewer.
func (t SteamTrain) steam(g *Grid) {
	chim := steamTrainNose - 2.4
	for i := 0; i < 9; i++ {
		fi := float64(i)
		s := chim - 0.6 - fi*2.9
		y := 29.9 + fi*0.75 + 0.12*fi*fi
		n := -0.3 * fi
		r := 0.85 + 0.27*fi
		t.fill(g, White, Ellipsoid{V(s, y, n), V(r*1.2, r*0.85, r)})
		if i > 1 {
			a := Hash01(i, 3) * 2 * math.Pi
			t.fill(g, White, Sphere(V(s-r*0.5, y+r*0.45+0.3*math.Sin(a), n+r*0.5*math.Cos(a)), r*0.7))
		}
	}
}

// steamTrainHouses are the cottages and the visitor centre at the loch head.
var steamTrainHouses = []struct {
	C          Vec2
	W, D, Wall float64
	Roof       Color
	Ridge      Axis
}{
	{P(21, 9), 7, 4.4, 2.8, Slate, AxisX},
	{P(29, 14.5), 5, 4, 2.6, Slate, AxisZ},
	{P(-30, 4), 5.4, 4, 2.6, Slate, AxisX},
	{P(-33, 15), 8, 4.6, 2.4, Charcoal, AxisZ},
	{P(42, 22), 13, 9, 0, Slate, AxisX}, // Church plot (built separately).
}

// church is a small Gothic kirk with a west tower and spire.
func (t SteamTrain) church(g *Grid, c Vec2) {
	x0, x1, z0, z1 := c.X-4.5, c.X+5, c.Y-2.6, c.Y+2.6
	base := 0.0
	for _, q := range []Vec2{c, P(x0-3, z0), P(x1, z0), P(x0-3, z1), P(x1, z1)} {
		base = math.Max(base, t.height(q.X, q.Y))
	}
	top := base + 4.2
	g.Use(DarkStone).Fill(Block(V(x0-3.6, base-2.5, z0-0.6), V(x1+0.6, base+0.4, z1+0.6)))
	g.Use(Stone).Fill(Subtract(Block(V(x0, base, z0), V(x1, top, z1)), Block(V(x0+0.8, base, z0+0.8), V(x1-0.8, top, z1-0.8))))
	g.Use(Charcoal).Fill(Block(V(x0+0.8, base, z0+0.8), V(x1-0.8, base+0.4, z1-0.8)))
	g.Paint(Charcoal, Subtract(Block(V(x0+0.4, base, z0+0.4), V(x1-0.4, top, z1-0.4)), Block(V(x0+0.8, base, z0+0.8), V(x1-0.8, top, z1-0.8))))
	for x := x0 + 1.8; x < x1-1; x += 2.2 {
		g.Carve(Arch{Sill: V(x, base+1.2, z0+0.4), HalfWidth: 0.5, Spring: 1.6, Depth: 1, Through: AxisZ},
			Arch{Sill: V(x, base+1.2, z1-0.4), HalfWidth: 0.5, Spring: 1.6, Depth: 1, Through: AxisZ})
	}
	g.Carve(Arch{Sill: V(x1-0.4, base+1.4, c.Y), HalfWidth: 0.9, Spring: 1.4, Depth: 1, Through: AxisX})
	g.Use(Slate).Fill(Gable{Min: P(x0-0.3, z0-0.5), Max: P(x1+0.4, z1+0.5), Base: top, Ridge: AxisX})
	// West tower with belfry and spire.
	tx := x0 - 1.4
	g.Use(Stone).Fill(Block(V(tx-1.6, base, c.Y-1.6), V(tx+1.6, base+9, c.Y+1.6)))
	g.Use(Charcoal).Fill(Block(V(tx-0.9, base+6.6, c.Y-0.9), V(tx+0.9, base+8.4, c.Y+0.9)))
	g.Carve(Block(V(tx-0.6, base+6.8, c.Y-2), V(tx+0.6, base+8.2, c.Y+2)), Block(V(tx-2, base+6.8, c.Y-0.6), V(tx+2, base+8.2, c.Y+0.6)))
	g.Carve(Arch{Sill: V(tx-1.4, base, c.Y), HalfWidth: 0.7, Spring: 1.2, Depth: 1, Through: AxisX})
	g.Use(DarkStone).Fill(Block(V(tx-1.9, base+9, c.Y-1.9), V(tx+1.9, base+9.5, c.Y+1.9)))
	g.Use(Slate).Fill(Func{Box{V(tx-1.7, base+9.5, c.Y-1.7), V(tx+1.7, base+15, c.Y+1.7)}, func(p Vec3) bool {
		k := 1.7 * (1 - (p.Y-base-9.5)/5.5)
		return math.Abs(p.X-tx) <= k && math.Abs(p.Z-c.Y) <= k
	}})
	g.Use(Gold).Fill(Block(V(tx-0.3, base+15, c.Y-0.3), V(tx+0.3, base+16, c.Y+0.3)))
	g.Use(Red).Fill(Block(V(tx-0.5, base, c.Y-0.6), V(tx-1.7, base+1.9, c.Y+0.6)))
}

func (t SteamTrain) nearBuildings(x, z, margin float64) bool {
	for _, hs := range steamTrainHouses {
		if math.Abs(x-hs.C.X) < hs.W/2+margin && math.Abs(z-hs.C.Y) < hs.D/2+margin {
			return true
		}
	}
	return false
}

// roadX is the glen road's centerline; it follows the west bank up the glen.
func (t SteamTrain) roadX(z float64) float64 { return -7 + 2.5*math.Sin(z*0.07) - 0.06*z }

func (t SteamTrain) settlement(g *Grid) {
	// Road: a paved strip on the terrain skin, from the loch head up the glen.
	g.Paint(DarkStone, Func{Box{V(-20, 0, steamTrainBack), V(10, 8, 24)}, func(p Vec3) bool {
		if math.Abs(p.X-t.roadX(p.Z)) > 1.2 {
			return false
		}
		h := t.height(p.X, p.Z)
		return p.Y > h-0.9 && h > t.waterLevel(p.Z)+0.3
	}})
	// Road bridge over the river.
	zb := 11.0
	rx := t.riverX(zb)
	g.Use(Stone).Fill(Subtract(Block(V(rx-5, 0.5, zb-1.4), V(rx+5, 2.9, zb+1.4)),
		Arch{Sill: V(rx, 0, zb), HalfWidth: 2.2, Spring: 0.6, Depth: 4, Through: AxisZ}))
	g.Use(DarkStone).Fill(Block(V(rx-5, 2.9, zb-1.4), V(rx+5, 3.4, zb-0.9)), Block(V(rx-5, 2.9, zb+0.9), V(rx+5, 3.4, zb+1.4)))
	for _, hs := range steamTrainHouses {
		if hs.Wall == 0 {
			t.church(g, hs.C)
			continue
		}
		c := hs.C
		x0, x1 := c.X-hs.W/2, c.X+hs.W/2
		z0, z1 := c.Y-hs.D/2, c.Y+hs.D/2
		base := math.Max(t.height(c.X, c.Y), t.waterLevel(c.Y)+0.5)
		for _, q := range []Vec2{P(x0, z0), P(x1, z0), P(x0, z1), P(x1, z1)} {
			base = math.Max(base, t.height(q.X, q.Y))
		}
		top := base + hs.Wall
		g.Use(DarkStone).Fill(Block(V(x0-0.3, base-2, z0-0.3), V(x1+0.3, base+0.4, z1+0.3)))
		g.Use(White).Fill(Subtract(Block(V(x0, base, z0), V(x1, top, z1)), Block(V(x0+0.8, base, z0+0.8), V(x1-0.8, top, z1-0.8))))
		g.Use(Charcoal).Fill(Block(V(x0+0.8, base, z0+0.8), V(x1-0.8, top-0.4, z1-0.8)))
		g.Carve(Block(V(x0+0.8, base+0.4, z0+0.8), V(x1-0.8, top-0.4, z1-0.8)))
		// Windows and a door on the long sides.
		if hs.Ridge == AxisX {
			for x := x0 + 1.2; x < x1-1.4; x += 2.2 {
				g.Paint(Glass, Block(V(x, base+1.0, z0-0.2), V(x+0.9, base+2.0, z1+0.2)))
			}
			g.Paint(Red, Block(V(c.X-0.4, base, z1-0.5), V(c.X+0.4, base+1.9, z1+0.2)))
		} else {
			for z := z0 + 1.2; z < z1-1.4; z += 2.2 {
				g.Paint(Glass, Block(V(x0-0.2, base+1.0, z), V(x1+0.2, base+2.0, z+0.9)))
			}
			g.Paint(Red, Block(V(x1-0.5, base, c.Y-0.4), V(x1+0.2, base+1.9, c.Y+0.4)))
		}
		g.Use(hs.Roof).Fill(Gable{Min: P(x0-0.4, z0-0.4), Max: P(x1+0.4, z1+0.4), Base: top, Ridge: hs.Ridge})
		// Gable-end chimneys.
		if hs.Ridge == AxisX {
			for _, x := range []float64{x0 + 0.5, x1 - 0.5} {
				g.Use(Stone).Fill(Block(V(x-0.5, top, c.Y-0.5), V(x+0.5, top+hs.D/2+1.2, c.Y+0.5)))
			}
		} else {
			for _, z := range []float64{z0 + 0.5, z1 - 0.5} {
				g.Use(Stone).Fill(Block(V(c.X-0.5, top, z-0.5), V(c.X+0.5, top+hs.W/2+1.2, z+0.5)))
			}
		}
	}
	// Drystone dykes around the glen-floor fields.
	dykes := [][]Vec2{
		{P(-38, -2), P(-24, -6), P(-14, -4), P(-11, 6), P(-22, 10), P(-36, 9)},
		{P(-14, -4), P(-12, -18), P(-24, -30), P(-30, -24)},
		{P(8, -6), P(20, -10), P(34, -6)},
		{P(9, 6), P(16, 4), P(30, 6), P(40, 12)},
		{P(10, -14), P(14, -26), P(26, -32), P(36, -26)},
	}
	g.Use(DarkStone)
	for _, line := range dykes {
		for i := 0; i+1 < len(line); i++ {
			a, b := line[i], line[i+1]
			box := Box{V(math.Min(a.X, b.X)-1, 0, math.Min(a.Y, b.Y)-1), V(math.Max(a.X, b.X)+1, 30, math.Max(a.Y, b.Y)+1)}
			g.Fill(Func{box, func(p Vec3) bool {
				ab := b.Sub(a)
				q := P(p.X, p.Z)
				u := math.Max(0, math.Min(1, q.Sub(a).Dot(ab)/ab.Dot(ab)))
				if q.Sub(a.Add(ab.Scale(u))).Length() > 0.5 {
					return false
				}
				if math.Abs(p.X-t.riverX(p.Z)) < 2.4 || math.Abs(p.X-t.roadX(p.Z)) < 1.4 {
					return false
				}
				h := t.height(p.X, p.Z)
				return h > t.waterLevel(p.Z)+0.3 && p.Y > h-0.5 && p.Y <= h+1.1
			}})
		}
	}
	// Boulders scattered over the glen and hillsides.
	for i := 0; i < 60; i++ {
		x := (Hash01(i, 91)*2 - 1) * (steamTrainHalfX - 5)
		z := steamTrainBack + 4 + Hash01(i, 92)*(steamTrainFront-steamTrainBack-8)
		h := t.height(x, z)
		s, n := steamTrainLocal(x, z)
		if h < t.waterLevel(z)+0.6 || (n > -6 && n < 8 && math.Abs(s) < steamTrainPortal+3) || t.nearBuildings(x, z, 2) ||
			math.Abs(x-t.roadX(z)) < 2.5 || math.Abs(x-t.riverX(z)) < 2.5 {
			continue
		}
		r := 0.9 + 1.3*Hash01(i, 93)
		c := DarkStone
		if Hash01(i, 94) < 0.35 {
			c = Stone
		}
		g.Use(c).FillVolume(Ellipsoid{V(x, h-0.2, z), V(r*1.2, r*0.8, r)})
	}
	// Rowing boats on the loch.
	for i, c := range []Vec2{P(10, 26), P(-2, 30)} {
		a := 0.6 + float64(i)*1.7
		hull := RotateY(Func{Box{V(c.X-2.4, 0.5, c.Y-1.1), V(c.X+2.4, 2.0, c.Y+1.1)}, func(p Vec3) bool {
			u := (p.X - c.X) / 2.4
			return math.Abs(p.Z-c.Y) <= 1.0*(1-u*u)+0.15 && p.Y <= 1.9
		}}, c, a)
		g.Use(Wood).Fill(hull)
		g.Carve(RotateY(Block(V(c.X-1.6, 1.4, c.Y-0.5), V(c.X+1.6, 2.0, c.Y+0.5)), c, a))
		g.Use(Red).Paint(Red, RotateY(Block(V(c.X-2.4, 1.6, c.Y-1.2), V(c.X+2.4, 1.9, c.Y+1.2)), c, a))
	}
	// Sheep grazing on the lower slopes.
	g.Use(White)
	for i := 0; i < 70; i++ {
		x := (Hash01(i, 81)*2 - 1) * (steamTrainHalfX - 6)
		z := steamTrainBack + 5 + Hash01(i, 82)*(steamTrainFront-steamTrainBack-10)
		h := t.height(x, z)
		s, n := steamTrainLocal(x, z)
		if h < t.waterLevel(z)+1 || h > 16 || (math.Abs(n) < 6 && math.Abs(s) < steamTrainPortal+2) || t.nearBuildings(x, z, 2) {
			continue
		}
		a := Hash01(i, 83) * math.Pi
		d := Polar(P(0, 0), 0.45, a)
		g.Fill(Capsule{V(x-d.X, h+0.6, z-d.Y), V(x+d.X, h+0.6, z+d.Y), 0.5})
		g.With(Charcoal, func() { g.Fill(Sphere(V(x+d.X*2, h+0.8, z+d.Y*2), 0.3)) })
	}
}
