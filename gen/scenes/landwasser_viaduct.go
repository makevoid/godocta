package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// LandwasserViaduct is the Rhaetian Railway's curved single-track limestone viaduct:
// six tall, slender arches on tapering piers stride across the Landwasser gorge and
// run straight into a tunnel portal halfway up a sheer cliff. A red train crosses,
// its tail still in the tunnel, under catenary masts. Below, the river foams around
// boulders; a spruce forest climbs the far slope to the spur where the line enters a
// second tunnel, and a wooden hut stands in a clearing near the valley mouth.
//
// The track follows a circle bowing toward +Z (the hero side). Railway parts are
// authored in curve coordinates (X = s along the track, Y = y, Z = n outward) and
// mapped onto the curve by landwasserBent; the train heads toward +s, away from the
// cliff tunnel at -X.
type LandwasserViaduct struct{}

func init() { Register(LandwasserViaduct{}) }

func (LandwasserViaduct) Name() string  { return "landwasser_viaduct" }
func (LandwasserViaduct) Title() string { return "Landwasser Viaduct" }

const (
	landwasserR      = 51.0  // Track curve radius.
	landwasserCZ     = -44.0 // Curve center z; the track crosses x = 0 at z = 7.
	landwasserE      = 35.7  // Viaduct half length along s (abutment pier centers).
	landwasserBays   = 6
	landwasserBed    = 34.0 // Deck top under the ballast.
	landwasserRail   = landwasserBed + 1
	landwasserSpring = landwasserBed - 6.8 // Arch springing height.
	landwasserFace   = -landwasserE - 1    // Cliff face (tunnel portal) along s.
	landwasserPortal = landwasserE + 9.5   // Second tunnel portal in the forested spur.
	landwasserWire   = landwasserBed + 6.6 // Contact wire underside.
	landwasserX0     = -56.0
	landwasserX1     = 60.0
	landwasserZ0     = -42.0
	landwasserZ1     = 30.0
)

// landwasserHut is the chalet's plan position.
var landwasserHut = P(15, 19)

type landwasserBent struct{ S Solid }

func landwasserLocal(x, z float64) (s, n float64) {
	dx, dz := x, z-landwasserCZ
	return landwasserR * math.Atan2(dx, dz), math.Hypot(dx, dz) - landwasserR
}

func landwasserAt(s, n float64) Vec2 {
	th, r := s/landwasserR, landwasserR+n
	return P(r*math.Sin(th), landwasserCZ+r*math.Cos(th))
}

func (b landwasserBent) Contains(p Vec3) bool {
	s, n := landwasserLocal(p.X, p.Z)
	return b.S.Contains(V(s, p.Y, n))
}

func (b landwasserBent) Bounds() Box {
	lb := b.S.Bounds()
	out := EmptyBox()
	const steps = 24
	for i := 0; i <= steps; i++ {
		s := lb.Min.X + (lb.Max.X-lb.Min.X)*float64(i)/steps
		for _, n := range []float64{lb.Min.Z, lb.Max.Z} {
			q := landwasserAt(s, n)
			out = out.Union(Box{q.At(lb.Min.Y), q.At(lb.Max.Y)})
		}
	}
	return out.Grow(0.6)
}

func landwasserSmooth(a, b, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-a)/(b-a)))
	return t * t * (3 - 2*t)
}

func landwasserNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	lerp := func(a, b, t float64) float64 { return a + (b-a)*t }
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return lerp(lerp(a, b, sx), lerp(c, d, sx), sz)*2 - 1
}

func landwasserFBM(x, z float64, seed int) float64 {
	return 0.6*landwasserNoise(x, z, seed) + 0.3*landwasserNoise(x*2.03+17, z*2.03+5, seed+1) +
		0.1*landwasserNoise(x*4.1+3, z*4.1+29, seed+2)
}

func landwasserBox(s0, y0, n0, s1, y1, n1 float64) Solid { return Block(V(s0, y0, n0), V(s1, y1, n1)) }

// faceFrame returns the cliff face's origin, its outward normal (along the track) and
// the in-face direction pointing toward the front-left.
func (LandwasserViaduct) faceFrame() (Vec2, Vec2, Vec2) {
	th := landwasserFace / landwasserR
	o := landwasserAt(landwasserFace, 0)
	t := P(math.Cos(th), -math.Sin(th))
	return o, t, P(-t.Y, t.X)
}

// faceCoords gives f (distance in front of the cliff face plane) and a (along it).
func (l LandwasserViaduct) faceCoords(x, z float64) (f, a float64) {
	o, t, u := l.faceFrame()
	d := P(x, z).Sub(o)
	return d.Dot(t), d.Dot(u)
}

// riverF is the river's centerline offset in front of the cliff face plane.
func (LandwasserViaduct) riverF(a float64) float64 {
	return 6.5 + 1.2*math.Sin(a*0.17+0.4) + 0.01*math.Max(0, a-14)*math.Max(0, a-14)
}

// water is the river surface; the river runs from the back toward the front-left.
func (LandwasserViaduct) water(a float64) float64 { return 2.3 - 0.025*a }

// massif is the cliff block's height: a sheer face around the portal, steep rocky flanks.
func (l LandwasserViaduct) massif(x, z float64) float64 {
	f, a := l.faceCoords(x, z)
	h := landwasserBed + 16 + 2.5*landwasserFBM(a*0.08+3, f*0.08, 5) - 2.6*math.Max(0, math.Abs(a)-10) - 1.1*math.Max(0, -f-11)
	steep := 9 - 7.7*landwasserSmooth(10, 17, math.Abs(a))
	rough := 1.3 * landwasserFBM(a*0.35, 3, 9) * landwasserSmooth(3, 7, math.Abs(a))
	return h - steep*math.Max(0, f+rough)
}

func (l LandwasserViaduct) rawHeight(x, z float64) float64 {
	_, n := landwasserLocal(x, z)
	f, a := l.faceCoords(x, z)
	// The far side stays low under the middle arches and climbs to the right abutment.
	h := 3.5 + (landwasserBed-4)*math.Pow(landwasserSmooth(-7, 37, x), 2.2)
	h += 9 * landwasserSmooth(-4, -24, n)
	h *= 1 - 0.85*landwasserSmooth(2, 24, n)
	h += 1.4 * landwasserFBM(x*0.08, z*0.08, 11)
	// The forested spur carrying the line to the second tunnel.
	s, _ := landwasserLocal(x, z)
	sc := math.Max(landwasserE+0.5, math.Min(landwasserE+15.5, s))
	top := landwasserBed + 1 + 7.5*landwasserSmooth(landwasserE+1.5, landwasserE+9.5, sc)
	d := P(x, z).Sub(landwasserAt(sc, 0)).Length()
	h = math.Max(h, top+1.2*landwasserFBM(x*0.1, z*0.1, 13)-1.05*math.Max(0, d-2))
	h = math.Max(h, l.massif(x, z))
	// River channel: gravel banks on the far side, the cliff drops straight in.
	dr := math.Abs(f - l.riverF(a))
	k := 1.3
	if f < l.riverF(a) {
		k = 8
	}
	return math.Min(h, l.water(a)-0.8+math.Max(0, dr-3.0)*k)
}

// height is the terrain surface; it is negative outside the site.
func (l LandwasserViaduct) height(x, z float64) float64 {
	ex := math.Min(x-landwasserX0, landwasserX1-x)
	ez := math.Min(z-landwasserZ0, landwasserZ1-z)
	if ex < 0 || ez < 0 {
		return -100
	}
	h := l.rawHeight(x, z)
	// A level terrace for the hut.
	if d := P(x, z).Sub(landwasserHut).Length(); d < 7 {
		base := l.rawHeight(landwasserHut.X, landwasserHut.Y)
		h = base + (h-base)*landwasserSmooth(4.5, 7, d)
	}
	// Rim: everything eases down to a low edge.
	e := math.Min(ex, ez)
	h = 1.2 + (h-1.2)*landwasserSmooth(-2, 15, e)
	// Track formation: cutting and portal approach beyond the right abutment.
	s, n := landwasserLocal(x, z)
	if s >= landwasserE-1 && s <= landwasserPortal+0.5 {
		w := math.Abs(n) - 3.2
		if w < 0 {
			h = landwasserBed
		} else {
			h = math.Max(landwasserBed-w, math.Min(landwasserBed+w*1.2, h))
		}
	}
	return h
}

func (l LandwasserViaduct) Build(g *Grid) Order {
	l.terrain(g)
	l.river(g)
	l.boulders(g)
	l.viaduct(g)
	l.tunnels(g)
	l.track(g)
	l.catenary(g)
	l.trees(g)
	l.hut(g)
	l.train(g)
	return Along(P(1, 0.2))
}

// ---------------------------------------------------------------------------------
// Terrain.

// terrain fills the height field as a skin: only cells with an empty face neighbor are
// kept (space below y = 0 counts as filled), so hills cost their surface, not volume.
func (l LandwasserViaduct) terrain(g *Grid) {
	step := g.Span() / 2
	x0 := int(math.Floor(landwasserX0/step)) - 3
	x1 := int(math.Ceil(landwasserX1/step)) + 3
	z0 := int(math.Floor(landwasserZ0/step)) - 3
	z1 := int(math.Ceil(landwasserZ1/step)) + 3
	w := x1 - x0 + 1
	cols := make([]float64, w*(z1-z0+1))
	for iz := z0; iz <= z1; iz++ {
		for ix := x0; ix <= x1; ix++ {
			cols[(iz-z0)*w+ix-x0] = l.height(float64(ix)*step, float64(iz)*step)
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
			rock := l.massif(x, z) >= h-0.05
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
				g.Use(l.groundColor(x, y, z, h, slope, rock, ix, iy, iz))
				g.Add(Cell{int32(ix), int32(iy), int32(iz)})
			}
		}
	}
}

func (l LandwasserViaduct) groundColor(x, y, z, h, slope float64, rock bool, ix, iy, iz int) Color {
	r := Hash01(ix, iy, iz, 41)
	f, a := l.faceCoords(x, z)
	s, n := landwasserLocal(x, z)
	inCut := s >= landwasserE-1 && s <= landwasserPortal+0.5
	top := h-y < 0.6
	face := rock && (slope > 2.4 || (f > -2 && math.Abs(a) < 14))
	strata := func() Color {
		band := math.Floor(y/1.7 + 0.8*landwasserFBM(a*0.12, y*0.05, 3))
		if int(band)%3 == 0 {
			if r < 0.7 {
				return DarkStone
			}
			return Stone
		}
		if r < 0.25 {
			return DarkStone
		}
		return Stone
	}
	switch {
	case inCut && math.Abs(n) < 3.3:
		return DarkStone
	case top && y > landwasserBed+4 && slope < 1.4 && rock && landwasserFBM(x*0.15, z*0.15, 7) > -0.1:
		// Snow lying on the cliff top and its high shelves.
		return White
	case face || (rock && slope > 2.4) || slope > 4.5:
		if top && y > landwasserBed+6 && slope < 2.8 && r < 0.6 {
			return White
		}
		return strata()
	case h-y > 1.4+1.4*slope:
		if r < 0.4 {
			return Stone
		}
		return DarkStone
	case slope > 1.8 && r < 0.05+0.25*landwasserSmooth(1.8, 3.4, slope):
		return strata()
	case math.Abs(f-l.riverF(a)) < 5.5 && h < l.water(a)+1.2:
		if r < 0.45 {
			return Stone
		}
		if r < 0.6 {
			return Sand
		}
		return DarkStone
	}
	if landwasserFBM(x*0.1+40, z*0.1, 23) > 0.2 && n > 6 {
		if r < 0.8 {
			return Grass
		}
		return Foliage
	}
	if r < 0.7 {
		return Foliage
	}
	return Grass
}

// river fills the channel up to the water level and froths it into rapids.
func (l LandwasserViaduct) river(g *Grid) {
	g.Use(Water).Fill(Func{Box{V(landwasserX0, 0, landwasserZ0), V(landwasserX1, 5, landwasserZ1)}, func(p Vec3) bool {
		f, a := l.faceCoords(p.X, p.Z)
		if math.Abs(f-l.riverF(a)) > 5 {
			return false
		}
		h := l.height(p.X, p.Z)
		return h > -50 && p.Y > h && p.Y <= l.water(a)
	}})
	// Rapids: white streaks on the surface where the bed steps down.
	g.Paint(White, Func{Box{V(landwasserX0, 0, landwasserZ0), V(landwasserX1, 5, landwasserZ1)}, func(p Vec3) bool {
		f, a := l.faceCoords(p.X, p.Z)
		if p.Y < l.water(a)-0.45 || p.Y > l.water(a) || math.Abs(f-l.riverF(a)) > 5 || p.Y <= l.height(p.X, p.Z) {
			return false
		}
		k := landwasserFBM(a*0.45, (f-l.riverF(a))*0.6, 31)
		return k > 0.18 || (math.Abs(f-l.riverF(a)) > 2.0 && k > 0.0)
	}})
}

// boulders scatter limestone blocks along the river and the gorge floor.
func (l LandwasserViaduct) boulders(g *Grid) {
	o, t, u := l.faceFrame()
	for i := 0; i < 34; i++ {
		a := -50 + float64(i)*2.7 + 1.5*Hash01(i, 1)
		off := (Hash01(i, 2)*2 - 1) * 4.2
		f := l.riverF(a) + off
		c := o.Add(t.Scale(f)).Add(u.Scale(a))
		h := l.height(c.X, c.Y)
		if h < 0 || h > l.water(a)+1.5 {
			continue
		}
		r := 0.9 + 1.2*Hash01(i, 3)
		col := DarkStone
		if Hash01(i, 4) < 0.45 {
			col = Stone
		}
		base := math.Min(h, l.water(a)-0.6)
		g.Use(col).FillVolume(Ellipsoid{c.At(base + r*0.35), V(r*1.2, r*0.8, r)})
		if math.Abs(off) < 3 {
			g.Paint(White, Func{Box{c.At(0).Sub(V(r+1.6, 0, r+1.6)), c.At(6).Add(V(r+1.6, 0, r+1.6))}, func(p Vec3) bool {
				_, a2 := l.faceCoords(p.X, p.Z)
				return p.Y <= l.water(a2) && p.Y > l.water(a2)-0.5 && P(p.X, p.Z).Sub(c).Length() < r+1.5 && p.Y < base+r*0.35
			}})
		}
	}
}

// ---------------------------------------------------------------------------------
// Viaduct.

func (l LandwasserViaduct) groundUnder(s, n float64) float64 {
	q := landwasserAt(s, n)
	return l.height(q.X, q.Y)
}

func (l LandwasserViaduct) viaduct(g *Grid) {
	const sp = 2 * landwasserE / landwasserBays
	const pw = 1.45 // Pier half width along the track at the springing.
	const a0 = sp/2 - pw
	opening := func(s, y float64) (bool, float64, float64) {
		j := math.Floor((s + landwasserE) / sp)
		if j < 0 || j >= landwasserBays {
			return false, 0, 0
		}
		c := -landwasserE + (j+0.5)*sp
		bat := math.Max(0, landwasserSpring-y)
		a := a0 - 0.045*bat
		ds := s - c
		dy := y - landwasserSpring
		if math.Abs(ds) < a && (dy < 0 || ds*ds+dy*dy < a0*a0) {
			return true, ds, a
		}
		return false, ds, a
	}
	halfN := func(y float64) float64 { return 2.3 + 0.055*math.Max(0, landwasserSpring-y) }
	lo, hi := landwasserFace-0.6, landwasserE+2.5
	body := Func{Box{V(lo, 0, -3.8), V(hi, landwasserBed, 3.8)}, func(p Vec3) bool {
		s, y, n := p.X, p.Y, p.Z
		if math.Abs(n) > halfN(y) {
			return false
		}
		if y < l.groundUnder(s, 0)-1.5 {
			return false
		}
		in, _, _ := opening(s, y)
		return !in
	}}
	g.Use(Stone).Fill(landwasserBent{body})
	// Mottled limestone: darker blocks, damp feet.
	g.Paint(DarkStone, landwasserBent{Func{body.Box, func(p Vec3) bool {
		if !body.F(p) {
			return false
		}
		if p.Y < l.groundUnder(p.X, 0)+2.2 {
			return Hash01(int(math.Floor(p.X*1.3)), int(math.Floor(p.Y*1.6)), int(math.Floor(p.Z*1.3)), 5) < 0.6
		}
		return Hash01(int(math.Floor(p.X/1.1)), int(math.Floor(p.Y/0.8)), int(math.Floor(p.Z/1.1)), 7) < 0.3
	}}})
	// Arch rings and imposts.
	g.Paint(DarkStone, landwasserBent{Func{body.Box, func(p Vec3) bool {
		_, ds, a := opening(p.X, p.Y)
		if p.Y >= landwasserSpring {
			r := math.Hypot(ds, p.Y-landwasserSpring)
			return r >= a0 && r < a0+0.8
		}
		return p.Y > landwasserSpring-0.6 && math.Abs(ds) < a+1.2 && math.Abs(ds) >= a
	}}})
	// Cornice, parapets and copings.
	g.Use(DarkStone).Fill(landwasserBent{landwasserBox(lo, landwasserBed-1.1, -2.75, landwasserE+1.6, landwasserBed-0.5, 2.75)})
	g.Use(Stone).Fill(landwasserBent{landwasserBox(lo, landwasserBed, 1.9, landwasserE+1.6, landwasserBed+1, 2.5)},
		landwasserBent{landwasserBox(lo, landwasserBed, -2.5, landwasserE+1.6, landwasserBed+1, -1.9)})
	g.Use(DarkStone).Fill(landwasserBent{landwasserBox(lo, landwasserBed+1, 1.85, landwasserE+1.6, landwasserBed+1.4, 2.6)},
		landwasserBent{landwasserBox(lo, landwasserBed+1, -2.6, landwasserE+1.6, landwasserBed+1.4, -1.85)})
}

// tunnels cut the portal into the cliff and the second one into the forested spur.
func (l LandwasserViaduct) tunnels(g *Grid) {
	const hw, spring = 2.9, landwasserBed + 4.4
	bore := func(s0, s1 float64) Solid {
		return Func{Box{V(s0, landwasserBed, -hw), V(s1, spring+hw, hw)}, func(p Vec3) bool {
			dy := p.Y - spring
			return math.Abs(p.Z) <= hw && (dy <= 0 || p.Z*p.Z+dy*dy <= hw*hw)
		}}
	}
	ring := func(s0, s1, r0, r1 float64) Solid {
		return Func{Box{V(s0, landwasserBed-0.5, -r1), V(s1, spring+r1, r1)}, func(p Vec3) bool {
			dy := math.Max(p.Y-spring, 0)
			r := math.Hypot(p.Z, dy)
			return r >= r0 && r < r1
		}}
	}
	// Linings exist only inside the rock; the terrain is a skin.
	inHill := func(sol Solid) Solid {
		b := landwasserBent{sol}
		return Func{b.Bounds(), func(p Vec3) bool {
			return b.Contains(p) && p.Y < l.height(p.X, p.Z)-0.2
		}}
	}
	// Cliff portal: a lined bore, dark inside, with a masonry ring set in the rock.
	f := landwasserFace
	g.Use(Charcoal).Fill(inHill(ring(f-7, f, hw, hw+1.4)))
	g.Use(Charcoal).Fill(inHill(landwasserBox(f-8, landwasserBed, -hw-1, f-7, spring+hw+1, hw+1)))
	g.Carve(landwasserBent{bore(f-7, f+1)})
	g.Use(Stone).Fill(landwasserBent{ring(f-0.6, f+0.7, hw, hw+1.3)})
	g.Paint(DarkStone, landwasserBent{Func{Box{V(f-0.6, landwasserBed, -5), V(f+0.7, spring+5, 5)}, func(p Vec3) bool {
		dy := math.Max(p.Y-spring, 0)
		ang := math.Atan2(dy, p.Z)
		return dy > 0 && int(math.Floor(ang/(math.Pi/11)))%2 == 0
	}}})
	// Second portal: a masonry headwall in the spur.
	p2 := landwasserPortal
	g.Use(Charcoal).Fill(inHill(ring(p2, p2+6, hw, hw+1.4)))
	g.Use(Charcoal).Fill(inHill(landwasserBox(p2+5, landwasserBed, -hw-1, p2+6, spring+hw+1, hw+1)))
	wall := Func{Box{V(p2-1.2, landwasserBed-1, -7), V(p2, landwasserBed+9, 7)}, func(p Vec3) bool {
		return p.Y <= landwasserBed+8.6-0.3*math.Abs(p.Z)
	}}
	g.Use(Stone).Fill(landwasserBent{wall})
	g.Paint(DarkStone, landwasserBent{Func{wall.Box, func(p Vec3) bool {
		return wall.F(p) && (p.Y > landwasserBed+7.8-0.3*math.Abs(p.Z) || Hash01(int(math.Floor(p.Y/0.8)), int(math.Floor(p.Z/1.3)), 9) < 0.25)
	}}})
	g.Carve(landwasserBent{bore(p2-1.5, p2+6)})
	g.Paint(DarkStone, landwasserBent{ring(p2-1.3, p2+0.1, hw, hw+1.0)})
}

func (l LandwasserViaduct) track(g *Grid) {
	lo, hi := landwasserFace-7, landwasserPortal+6
	g.Use(DarkStone).Fill(landwasserBent{landwasserBox(lo, landwasserBed, -1.8, hi, landwasserBed+0.4, 1.8)})
	g.Use(Wood)
	for s := math.Floor(lo/1.6) * 1.6; s < hi; s += 1.6 {
		g.Fill(landwasserBent{landwasserBox(s, landwasserBed+0.4, -1.5, s+0.8, landwasserBed+0.6, 1.5)})
	}
	g.Use(Slate).Fill(landwasserBent{landwasserBox(lo, landwasserBed+0.5, -1.1, hi, landwasserRail, -0.6)},
		landwasserBent{landwasserBox(lo, landwasserBed+0.5, 0.6, hi, landwasserRail, 1.1)})
}

// catenary: masts on the inner parapet with cantilevers and the contact wire.
func (l LandwasserViaduct) catenary(g *Grid) {
	const sp = 2 * landwasserE / landwasserBays
	top := landwasserWire + 1.6
	var masts []float64
	for k := 0; k < landwasserBays; k++ {
		masts = append(masts, -landwasserE+(float64(k)+0.5)*sp)
	}
	masts = append(masts, landwasserE+6.5)
	g.Use(Slate)
	for _, s := range masts {
		g.Fill(landwasserBent{landwasserBox(s-0.4, landwasserBed-1.1, -3.4, s+0.4, top, -2.6)})
		g.Fill(landwasserBent{landwasserBox(s-0.3, top-0.9, -3.0, s+0.3, top-0.4, 0.5)})
		g.Fill(landwasserBent{Capsule{V(s, top-2.6, -2.8), V(s, top-0.6, -0.2), 0.3}})
		g.Fill(landwasserBent{landwasserBox(s-0.3, landwasserWire, -0.3, s+0.3, top-0.6, 0.3)})
	}
	g.Use(Charcoal).Fill(landwasserBent{landwasserBox(landwasserFace+0.8, landwasserWire, -0.4, landwasserPortal-1.6, landwasserWire+0.6, 0.4)})
}

// ---------------------------------------------------------------------------------
// Forest and hut.

func (l LandwasserViaduct) spruce(g *Grid, c Vec2, base, height float64, seed int) {
	g.Use(Wood).Fill(Capsule{c.At(base - 1), c.At(base + 1.8), 0.45})
	r := 1.7 + 0.5*Hash01(seed, 1)
	col := Foliage
	if Hash01(seed, 2) < 0.18 {
		col = Grass
	}
	g.Use(col)
	tiers := 3
	for k := 0; k < tiers; k++ {
		fk := float64(k)
		b := base + 1.2 + fk*height*0.24
		rr := r * (1 - 0.24*fk)
		hh := height*0.62 - fk*height*0.1
		if k == tiers-1 {
			hh = base + height - b
		}
		g.FillVolume(Cone(c, rr, b, hh))
	}
}

func (l LandwasserViaduct) trees(g *Grid) {
	const cell = 3.3
	const sp = 2 * landwasserE / landwasserBays
	for i := -19; i <= 19; i++ {
		for j := -14; j <= 10; j++ {
			jx, jz := Hash01(i, j, 61), Hash01(i, j, 62)
			x := (float64(i) + jx - 0.5) * cell
			z := (float64(j) + jz - 0.5) * cell
			if x < landwasserX0+5 || x > landwasserX1-5 || z < landwasserZ0+5 || z > landwasserZ1-5 {
				continue
			}
			h := l.height(x, z)
			if h < 0 {
				continue
			}
			if fx, ax := l.faceCoords(x, z); fx > -3 && math.Abs(ax) < 14 {
				continue
			}
			f, a := l.faceCoords(x, z)
			if h < l.water(a)+1.2 || math.Abs(f-l.riverF(a)) < 5 {
				continue
			}
			sl := math.Hypot(l.height(x+1, z)-l.height(x-1, z), l.height(x, z+1)-l.height(x, z-1)) / 2
			if sl > 3.8 {
				continue
			}
			s, n := landwasserLocal(x, z)
			ht := 7 + 4*Hash01(i, j, 64)
			// Keep the track, its cutting and the viaduct deck clear.
			if math.Abs(n) < 6 && s > landwasserFace-4 && s < landwasserPortal+3 {
				if s > landwasserE-2 || h+ht > landwasserSpring-1 {
					continue
				}
				k := math.Round((s + landwasserE) / sp)
				if math.Abs(s-(-landwasserE+k*sp)) < 4 {
					continue
				}
			}
			if P(x, z).Sub(landwasserHut).Length() < 8 {
				continue
			}
			// Forest thins toward the valley mouth meadows.
			dens := 0.6 + 0.4*landwasserSmooth(-0.2, 0.3, landwasserFBM(x*0.05+7, z*0.05, 71))
			dens *= 1 - 0.6*landwasserSmooth(8, 26, n)
			if n > 3 && s > landwasserFace && s < landwasserE+4 {
				dens *= 0.3
			}
			if Hash01(i, j, 63) > dens {
				continue
			}
			l.spruce(g, P(x, z), h, ht, i*31+j)
		}
	}
}

// hut is a small timber chalet with a stone base and a shingled gable roof.
func (l LandwasserViaduct) hut(g *Grid) {
	c := landwasserHut
	base := l.height(c.X, c.Y)
	const w, d = 6.0, 4.6
	g.Use(DarkStone).Fill(BlockAt(c, w+0.4, d+0.4, base-1.5, base+0.6))
	g.Use(Wood).Fill(BlockAt(c, w, d, base+0.6, base+4.4))
	g.Carve(BlockAt(c, w-1.4, d-1.4, base+0.6, base+4.4))
	g.Use(Charcoal).Fill(BlockAt(c, w-1.4, d-1.4, base+0.6, base+1.1))
	// A footpath winds from the hut down to the river bank.
	o, t, u := l.faceFrame()
	bank := o.Add(t.Scale(l.riverF(16) + 3.4)).Add(u.Scale(16))
	path := []Vec2{P(c.X+w/2+0.6, c.Y), P(c.X+4, c.Y+5), P(c.X-6, c.Y+5.5), P(-4, 23), P(-18, 20), P(-28, 15), bank}
	g.Paint(Sand, Func{Box{V(landwasserX0, 0, landwasserZ0), V(landwasserX1, 20, landwasserZ1)}, func(p Vec3) bool {
		q := P(p.X, p.Z)
		h := l.height(p.X, p.Z)
		if p.Y < h-0.8 {
			return false
		}
		for i := 0; i+1 < len(path); i++ {
			a, b := path[i], path[i+1]
			ab := b.Sub(a)
			k := math.Max(0, math.Min(1, q.Sub(a).Dot(ab)/ab.Dot(ab)))
			if q.Sub(a.Add(ab.Scale(k))).Length() < 0.8 {
				return true
			}
		}
		return false
	}})
	// Windows and door.
	for _, x := range []float64{-1.6, 1.4} {
		g.Paint(Charcoal, BlockAt(P(c.X+x, c.Y+d/2), 1.0, 0.8, base+2.4, base+3.4))
		g.Paint(Charcoal, BlockAt(P(c.X+x, c.Y-d/2), 1.0, 0.8, base+2.4, base+3.4))
	}
	g.Carve(BlockAt(P(c.X+w/2, c.Y), 0.8, 1.2, base+1.5, base+3.6))
	g.Paint(White, BlockAt(P(c.X+w/2, c.Y), 0.8, 2.0, base+3.6, base+4.0))
	// Gable roof with overhangs.
	g.Use(Slate).Fill(Gable{P(c.X-w/2-0.8, c.Y-d/2-0.8), P(c.X+w/2+0.8, c.Y+d/2+0.8), base + 4.4, AxisX})
	g.Use(Wood).Fill(Gable{P(c.X-w/2, c.Y-d/2+0.5), P(c.X+w/2, c.Y+d/2-0.5), base + 4.4, AxisX})
	g.Use(DarkStone).Fill(BlockAt(P(c.X+1.5, c.Y+0.8), 0.8, 0.8, base+5, base+8.4))
	// Woodpile and a fence.
	g.Use(Wood).Fill(BlockAt(P(c.X-w/2-1.2, c.Y), 1.0, 2.4, base, base+1.5))
	for k := -3; k <= 3; k++ {
		p := P(c.X+float64(k)*1.6, c.Y+d/2+3)
		h := l.height(p.X, p.Y)
		g.Use(Wood).Fill(BlockAt(p, 0.5, 0.5, h, h+1.6))
	}
	g.Use(Wood).Fill(Func{Box{V(c.X-5, base-3, c.Y+d/2+2.7), V(c.X+5, base+3, c.Y+d/2+3.3)}, func(p Vec3) bool {
		h := l.height(p.X, p.Z)
		return p.Y > h+1.0 && p.Y < h+1.5
	}})
}

// ---------------------------------------------------------------------------------
// Train.

func (l LandwasserViaduct) fill(g *Grid, c Color, solids ...Solid) {
	g.Use(c)
	for _, s := range solids {
		g.Fill(landwasserBent{s})
	}
}

func (l LandwasserViaduct) paint(g *Grid, c Color, solids ...Solid) {
	for _, s := range solids {
		g.Paint(c, landwasserBent{s})
	}
}

const (
	landwasserFloor = landwasserBed + 2   // Car body bottom.
	landwasserRoof  = landwasserBed + 4.6 // Car body top.
	landwasserHalf  = 1.3                 // Car half width.
	landwasserClip  = landwasserFace - 4
)

func (l LandwasserViaduct) clip(s float64) float64 { return math.Max(s, landwasserClip) }

func (l LandwasserViaduct) bogies(g *Grid, s0, s1 float64) {
	for _, c := range []float64{s0 + 2.2, s1 - 2.2} {
		if c+1.6 < landwasserClip {
			continue
		}
		l.fill(g, Charcoal, landwasserBox(l.clip(c-1.6), landwasserRail, -1.15, c+1.6, landwasserFloor, 1.15))
	}
	l.fill(g, Charcoal, landwasserBox(l.clip(s0+0.3), landwasserFloor-0.4, -1.2, s1-0.3, landwasserFloor, 1.2))
}

func (l LandwasserViaduct) carriage(g *Grid, s0, s1 float64) {
	if s1 < landwasserClip {
		return
	}
	a := l.clip(s0)
	l.bogies(g, s0, s1)
	l.fill(g, Red, landwasserBox(a, landwasserFloor, -landwasserHalf, s1, landwasserRoof, landwasserHalf))
	// Window band with a white stripe beneath.
	l.paint(g, White, landwasserBox(a, landwasserBed+2.5, -2, s1, landwasserBed+2.9, -1.0), landwasserBox(a, landwasserBed+2.5, 1.0, s1, landwasserBed+2.9, 2))
	for x := s0 + 1.6; x+1.2 < s1-1.4; x += 1.75 {
		if x < a {
			continue
		}
		l.paint(g, Glass, landwasserBox(x, landwasserBed+3.2, -2, x+1.2, landwasserBed+4.2, -1.0), landwasserBox(x, landwasserBed+3.2, 1.0, x+1.2, landwasserBed+4.2, 2))
	}
	for _, d := range []float64{s0 + 0.4, s1 - 1.3} {
		if d < a {
			continue
		}
		l.paint(g, Charcoal, landwasserBox(d, landwasserBed+2.1, -2, d+0.9, landwasserBed+4.3, -1.0), landwasserBox(d, landwasserBed+2.1, 1.0, d+0.9, landwasserBed+4.3, 2))
	}
	// Grey roof.
	l.fill(g, Stone, Func{Box{V(a, landwasserRoof, -landwasserHalf), V(s1, landwasserRoof+0.6, landwasserHalf)}, func(p Vec3) bool {
		q := p.Z / landwasserHalf
		return p.Y-landwasserRoof <= 0.6*math.Sqrt(math.Max(0, 1-q*q))
	}})
	// Gangway to the vehicle ahead.
	l.fill(g, Charcoal, landwasserBox(s1, landwasserFloor, -0.9, s1+0.6, landwasserBed+4.2, 0.9))
}

func (l LandwasserViaduct) pantograph(g *Grid, s float64, raised bool) {
	top := landwasserRoof + 0.5
	l.fill(g, DarkStone, landwasserBox(s-0.9, top, -0.9, s+0.9, top+0.4, 0.9))
	if !raised {
		l.fill(g, Charcoal, landwasserBox(s-1.4, top+0.4, -0.3, s+1.0, top+0.8, 0.3))
		return
	}
	knee := V(s+0.7, top+1.1, 0)
	l.fill(g, Charcoal, Capsule{V(s-0.7, top+0.6, 0), knee, 0.33}, Capsule{knee, V(s-0.1, landwasserWire-0.2, 0), 0.33})
	l.fill(g, Charcoal, landwasserBox(s-0.4, landwasserWire-0.45, -1.0, s+0.3, landwasserWire, 1.0))
}

func (l LandwasserViaduct) locomotive(g *Grid, s0, s1 float64) {
	l.bogies(g, s0, s1)
	l.fill(g, Red, landwasserBox(s0, landwasserFloor, -landwasserHalf, s1, landwasserRoof, landwasserHalf))
	// Raked cab ends with windscreens.
	for _, e := range []struct{ s, dir float64 }{{s1, 1}, {s0, -1}} {
		l.paint(g, Glass, landwasserBox(math.Min(e.s, e.s-e.dir*0.5), landwasserBed+3.4, -1.0, math.Max(e.s, e.s-e.dir*0.5), landwasserBed+4.4, 1.0))
		l.paint(g, Glass, landwasserBox(math.Min(e.s-e.dir*0.6, e.s-e.dir*1.6), landwasserBed+3.4, -2, math.Max(e.s-e.dir*0.6, e.s-e.dir*1.6), landwasserBed+4.3, -1.0),
			landwasserBox(math.Min(e.s-e.dir*0.6, e.s-e.dir*1.6), landwasserBed+3.4, 1.0, math.Max(e.s-e.dir*0.6, e.s-e.dir*1.6), landwasserBed+4.3, 2))
		l.fill(g, White, landwasserBox(math.Min(e.s, e.s+e.dir*0.3), landwasserBed+2.4, -1.0, math.Max(e.s, e.s+e.dir*0.3), landwasserBed+2.8, -0.5),
			landwasserBox(math.Min(e.s, e.s+e.dir*0.3), landwasserBed+2.4, 0.5, math.Max(e.s, e.s+e.dir*0.3), landwasserBed+2.8, 1.0))
		l.fill(g, Charcoal, landwasserBox(math.Min(e.s, e.s+e.dir*0.5), landwasserBed+1.8, -0.4, math.Max(e.s, e.s+e.dir*0.5), landwasserBed+2.3, 0.4))
	}
	// Machine-room louvres and a white band.
	for x := s0 + 2.6; x < s1-2.6; x += 0.9 {
		l.paint(g, DarkStone, landwasserBox(x, landwasserBed+3.3, -2, x+0.45, landwasserBed+4.2, -1.0), landwasserBox(x, landwasserBed+3.3, 1.0, x+0.45, landwasserBed+4.2, 2))
	}
	l.paint(g, White, landwasserBox(s0+0.4, landwasserBed+2.5, -2, s1-0.4, landwasserBed+2.9, -1.0), landwasserBox(s0+0.4, landwasserBed+2.5, 1.0, s1-0.4, landwasserBed+2.9, 2))
	l.fill(g, Stone, landwasserBox(s0+0.2, landwasserRoof, -1.1, s1-0.2, landwasserRoof+0.5, 1.1))
	l.fill(g, DarkStone, landwasserBox(s0+3.5, landwasserRoof+0.5, -0.7, s1-3.5, landwasserRoof+0.9, 0.7))
	l.pantograph(g, s1-2.2, true)
	l.pantograph(g, s0+2.2, false)
}

func (l LandwasserViaduct) train(g *Grid) {
	nose := 15.0
	const loco, car, gap = 10.0, 11.0, 0.6
	l.locomotive(g, nose-loco, nose)
	s1 := nose - loco - gap
	for i := 0; i < 4; i++ {
		l.carriage(g, s1-car, s1)
		s1 -= car + gap
	}
}
