package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// DinosaurValley is a lush prehistoric valley. A long-necked sauropod browses the crowns
// of tall araucarias, a big theropod with open jaws strides toward a horned ceratopsian
// that stands its ground, a plated stegosaur drinks from the river on top of a banded
// rock escarpment before it spills over a waterfall, an armored ankylosaur grazes in the
// foreground and a small raptor pack circles a nest of white eggs. A smoking volcano
// with a glowing crater and a lava flow rises at the back left. Giant ferns, cycads,
// tree ferns and conifers fill the valley, whose floor slopes to a low rim; pterosaurs
// glide overhead.
//
// The layout is authored in a view frame (u, d): u runs left to right and d toward the
// hero camera (yaw 0.68), so the animals stand in profile in the hero view.
type DinosaurValley struct{}

func init() { Register(DinosaurValley{}) }

func (DinosaurValley) Name() string  { return "dinosaur_valley" }
func (DinosaurValley) Title() string { return "Dinosaur Valley" }

// dinosaurValleyW converts view-frame (u, d) to a world plan point.
func dinosaurValleyW(u, d float64) Vec2 {
	s, c := math.Sincos(0.68)
	return P(u*c+d*s, -u*s+d*c)
}

// dinosaurValleyUV converts a world plan point to the view frame.
func dinosaurValleyUV(p Vec2) Vec2 {
	s, c := math.Sincos(0.68)
	return P(p.X*c-p.Y*s, p.X*s+p.Y*c)
}

// dinosaurValleyHeading is the world plan angle of a view-frame direction.
func dinosaurValleyHeading(u, d float64) float64 { return dinosaurValleyW(u, d).Angle() }

func dinosaurValleyPath(pts ...Vec2) []Vec2 {
	out := make([]Vec2, len(pts))
	for i, p := range pts {
		out[i] = dinosaurValleyW(p.X, p.Y)
	}
	return out
}

var (
	dinosaurValleyVolcano = dinosaurValleyW(-28, -29) // Cone axis.
	dinosaurValleyFallsUV = P(16, -21)                // Waterfall lip in the view frame.
	dinosaurValleyPoolUV  = P(16.5, -16)              // Plunge pool.
	// River from the plateau's back edge, over the falls, to the front-right rim.
	dinosaurValleyRiver = dinosaurValleyPath(P(19, -46), P(17, -34), P(16, -26), P(16, -21), P(16.5, -16),
		P(23, -11), P(30, -5), P(35, 3), P(38, 12), P(38, 20), P(36, 28), P(33, 40))
	dinosaurValleyLavaAngle = dinosaurValleyHeading(-0.5, 0.87) // Lava flow direction.
)

const (
	dinosaurValleySiteA   = 46.0 // Site ellipse semi-axes in the view frame.
	dinosaurValleySiteB   = 38.0
	dinosaurValleyRim     = 24.0 // Crater rim height.
	dinosaurValleyCrater  = 4.5  // Crater radius.
	dinosaurValleyFloor   = 20.5 // Lava lake level.
	dinosaurValleyCone    = 19.0 // Cone base radius.
	dinosaurValleyPlateau = 12.0 // Escarpment height.
	dinosaurValleyWater   = 1.0  // River surface below the falls.
)

// dinosaurValleyTaper is a capsule whose radius varies linearly from A to B.
type dinosaurValleyTaper struct {
	A, B   Vec3
	R0, R1 float64
}

func (c dinosaurValleyTaper) Contains(p Vec3) bool {
	ab := c.B.Sub(c.A)
	t := math.Max(0, math.Min(1, p.Sub(c.A).Dot(ab)/math.Max(ab.Dot(ab), 1e-12)))
	return p.Sub(c.A.Add(ab.Scale(t))).Length() <= c.R0+(c.R1-c.R0)*t
}

func (c dinosaurValleyTaper) Bounds() Box {
	return BoxOf(c.A, c.B).Grow(math.Max(c.R0, c.R1))
}

// dinosaurValleyPlaced maps a solid authored in local coordinates (facing +X, feet at
// Y 0) into the world: scaled, turned to Heading (plan angle) and moved to At.
type dinosaurValleyPlaced struct {
	S       Solid
	At      Vec3
	Heading float64
	Scale   float64
}

func (s dinosaurValleyPlaced) local(p Vec3) Vec3 {
	d := p.Sub(s.At)
	q := P(d.X, d.Z).Rotate(-s.Heading)
	return V(q.X/s.Scale, d.Y/s.Scale, q.Y/s.Scale)
}

func (s dinosaurValleyPlaced) world(p Vec3) Vec3 {
	q := P(p.X, p.Z).Scale(s.Scale).Rotate(s.Heading)
	return V(s.At.X+q.X, s.At.Y+p.Y*s.Scale, s.At.Z+q.Y)
}

func (s dinosaurValleyPlaced) Contains(p Vec3) bool { return s.S.Contains(s.local(p)) }

func (s dinosaurValleyPlaced) Bounds() Box {
	in := s.S.Bounds()
	b := EmptyBox()
	for _, x := range []float64{in.Min.X, in.Max.X} {
		for _, y := range []float64{in.Min.Y, in.Max.Y} {
			for _, z := range []float64{in.Min.Z, in.Max.Z} {
				w := s.world(V(x, y, z))
				b = b.Union(Box{w, w})
			}
		}
	}
	return b
}

// dinosaurValleyBeast places local-space parts of one animal into its own grid, so
// that its paints never touch the terrain or plants.
type dinosaurValleyBeast struct {
	g       *Grid
	at      Vec3
	heading float64
	scale   float64
}

func (b dinosaurValleyBeast) w(s Solid) Solid {
	return dinosaurValleyPlaced{s, b.at, b.heading, b.scale}
}

// vol fills curved parts by sampled volume.
func (b dinosaurValleyBeast) vol(c Color, parts ...Solid) {
	b.g.With(c, func() {
		for _, s := range parts {
			b.g.FillVolume(b.w(s))
		}
	})
}

// fine fills thin parts (teeth, horns, claws) by cell center so they stay connected.
func (b dinosaurValleyBeast) fine(c Color, parts ...Solid) {
	b.g.With(c, func() {
		for _, s := range parts {
			b.g.Fill(b.w(s))
		}
	})
}

// paint recolors the animal's cells inside a local box that match a local predicate.
func (b dinosaurValleyBeast) paint(c Color, box Box, f func(Vec3) bool) {
	b.g.Paint(c, b.w(Func{box, f}))
}

// cellHash is a deterministic per-cell value for a local point.
func (b dinosaurValleyBeast) cellHash(p Vec3, seed int) float64 {
	w := dinosaurValleyPlaced{nil, b.at, b.heading, b.scale}.world(p)
	c := Nearest(w.Scale(1 / b.g.Span()))
	return Hash01(int(c.X), int(c.Y), int(c.Z), seed)
}

// animal builds one creature standing on the terrain at a view-frame position.
func (s DinosaurValley) animal(g *Grid, H func(x, z float64) float64, uv Vec2, dir Vec2, scale float64,
	build func(b dinosaurValleyBeast)) {
	p := dinosaurValleyW(uv.X, uv.Y)
	tmp := NewGridWithSpan(g.Span())
	build(dinosaurValleyBeast{tmp, p.At(H(p.X, p.Y) - 0.35), dinosaurValleyHeading(dir.X, dir.Y), scale})
	g.Merge(tmp)
}

// chain is a tapered tube through points with per-point radii.
func (DinosaurValley) chain(pts []Vec3, radii []float64) Solid {
	u := make(UnionSolid, 0, len(pts)-1)
	for i := 0; i+1 < len(pts); i++ {
		u = append(u, dinosaurValleyTaper{pts[i], pts[i+1], radii[i], radii[i+1]})
	}
	return u
}

func dinosaurValleySmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func dinosaurValleyLerp(a, b, t float64) float64 { return a + (b-a)*t }

// dinosaurValleyNoise is smooth value noise in [-1, 1].
func dinosaurValleyNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return dinosaurValleyLerp(dinosaurValleyLerp(a, b, sx), dinosaurValleyLerp(c, d, sx), sz)*2 - 1
}

func dinosaurValleyFBM(x, z float64, seed int) float64 {
	return 0.6*dinosaurValleyNoise(x, z, seed) + 0.3*dinosaurValleyNoise(x*2.03+17, z*2.03+5, seed+1) +
		0.1*dinosaurValleyNoise(x*4.1+3, z*4.1+29, seed+2)
}

// polyline returns the plan distance to a polyline and the arc length at the nearest point.
func (DinosaurValley) polyline(p Vec2, pts []Vec2) (dist, along float64) {
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

// siteR is the normalized, slightly irregular site radius (1 at the rim) of a world point.
func (DinosaurValley) siteR(x, z float64) float64 {
	q := dinosaurValleyUV(P(x, z))
	r := math.Hypot(q.X/dinosaurValleySiteA, q.Y/dinosaurValleySiteB)
	return r + 0.03*dinosaurValleyNoise(math.Atan2(q.Y, q.X)*3+5, 1, 7)
}

// riverW is the river's half width at an arc length.
func (DinosaurValley) riverW(along float64) float64 { return 1.6 + along*0.018 }

// volcanoH is the cone's height (0 outside its base) with the crater and lava notch.
func (DinosaurValley) volcanoH(x, z float64) float64 {
	d := P(x, z).Sub(dinosaurValleyVolcano)
	r, a := d.Length(), d.Angle()
	if r >= dinosaurValleyCone {
		return 0
	}
	da := math.Remainder(a-dinosaurValleyLavaAngle, 2*math.Pi)
	notch := 1.6 * math.Exp(-da*da/0.08)
	rimY := dinosaurValleyRim + 0.5*dinosaurValleyNoise(math.Cos(a)*3, math.Sin(a)*3, 5) - notch
	if r <= dinosaurValleyCrater {
		k := dinosaurValleySmooth((dinosaurValleyCrater - r) / 2.0)
		return dinosaurValleyLerp(rimY, dinosaurValleyFloor, k)
	}
	u := (dinosaurValleyCone - r) / (dinosaurValleyCone - dinosaurValleyCrater)
	h := rimY * math.Pow(u, 1.35)
	ridge := math.Sin(11*a + 1.4*math.Sin(3*a))
	fade := dinosaurValleySmooth((r-dinosaurValleyCrater)/4) * dinosaurValleySmooth((dinosaurValleyCone-r)/5)
	h -= 1.2 * fade * math.Max(0, ridge) * math.Max(0, ridge)
	return h
}

// lavaPath is the flow's centerline down the flank.
func (DinosaurValley) lavaPath() []Vec2 {
	pts := []Vec2{}
	for i := 0; i <= 12; i++ {
		t := float64(i) / 12
		r := dinosaurValleyCrater - 0.5 + (dinosaurValleyCone+1.5-dinosaurValleyCrater)*t
		a := dinosaurValleyLavaAngle + 0.12*math.Sin(t*7)
		pts = append(pts, Polar(dinosaurValleyVolcano, r, a))
	}
	return pts
}

// raw is the uncached terrain height; below 0 means outside the site.
func (s DinosaurValley) raw(x, z float64) float64 {
	q := dinosaurValleyUV(P(x, z))
	u, d := q.X, q.Y
	sr := s.siteR(x, z)
	vh := s.volcanoH(x, z)
	if sr > 1 && vh <= 0 {
		return -1
	}
	edge := dinosaurValleySmooth((1 - sr) * dinosaurValleySiteB / 7)
	h := 0.3 + 1.5*edge + 0.7*edge*dinosaurValleyFBM(x*0.07, z*0.07, 11)
	// Rocky escarpment at the back right, sloping off behind to the rim.
	cliffD := dinosaurValleyFallsUV.Y + 1.6*dinosaurValleyNoise(u*0.18, 3, 21)
	if u < -6 {
		cliffD -= (-6 - u) * 0.9
	}
	back := dinosaurValleySmooth((1 - sr) * dinosaurValleySiteB / 9)
	plateau := dinosaurValleyPlateau + 0.8*dinosaurValleyFBM(x*0.1, z*0.1, 23)
	if pl := plateau * dinosaurValleySmooth((cliffD-d)/2.0) * back; pl > h {
		h = pl
	}
	h = math.Max(h, vh)
	// River: a channel cut into the plateau and the valley floor.
	if dist, along := s.polyline(P(x, z), dinosaurValleyRiver); dist < 6 {
		w := s.riverW(along)
		if d < dinosaurValleyFallsUV.Y {
			h = dinosaurValleyLerp(h, h-1.4, dinosaurValleySmooth((w+1.2-dist)/1.2))
		} else {
			bed := dinosaurValleyWater - 0.7
			k := dinosaurValleySmooth((w + 1.8 - dist) / 1.8)
			h = dinosaurValleyLerp(h, math.Min(h, bed), k)
		}
	}
	// Plunge pool under the falls.
	if dp := q.Sub(dinosaurValleyPoolUV).Length(); dp < 5.5 && d > dinosaurValleyFallsUV.Y+1.5 {
		h = dinosaurValleyLerp(h, math.Min(h, dinosaurValleyWater-0.7), dinosaurValleySmooth((5.5-dp)/1.5))
	}
	return h
}

// dinosaurValleyLand memoizes terrain heights.
type dinosaurValleyLand struct{ cache map[[2]int64]float64 }

func (s DinosaurValley) height(land *dinosaurValleyLand, x, z float64) float64 {
	key := [2]int64{int64(math.Round(x * 1000)), int64(math.Round(z * 1000))}
	if h, ok := land.cache[key]; ok {
		return h
	}
	h := s.raw(x, z)
	land.cache[key] = h
	return h
}

// shell fills the cells of a predicate within t of its outside (sideways or above).
func (DinosaurValley) shell(g *Grid, box Box, t float64, in func(Vec3) bool) {
	g.Fill(Func{box, func(p Vec3) bool {
		if !in(p) {
			return false
		}
		return !(in(p.Add(V(t, 0, 0))) && in(p.Add(V(-t, 0, 0))) && in(p.Add(V(0, 0, t))) &&
			in(p.Add(V(0, 0, -t))) && in(p.Add(V(0, t, 0))))
	}})
}

func (DinosaurValley) cell(g *Grid, p Vec3) Cell { return Nearest(p.Scale(1 / g.Span())) }

// Animal placements in the view frame: position and facing direction.
var (
	dinosaurValleyBrachioUV = P(-2, -12)
	dinosaurValleyRexUV     = P(-14, 8)
	dinosaurValleyTriUV     = P(19, 8)
	dinosaurValleyStegoUV   = P(1.5, -28.5)
	dinosaurValleyNestUV    = P(9, 27)
	dinosaurValleyAnkyUV    = P(-25, 21)
	dinosaurValleyRaptors   = []struct{ at, dir Vec2 }{
		{P(-1, 23), P(1, 0.2)}, {P(2, 31), P(1, -0.35)}, {P(20, 29), P(-1, -0.1)},
	}
)

func (s DinosaurValley) Build(g *Grid) Order {
	land := &dinosaurValleyLand{cache: make(map[[2]int64]float64)}
	H := func(x, z float64) float64 { return s.height(land, x, z) }
	site := Box{V(-52, -0.5, -52), V(52, dinosaurValleyRim+2, 52)}
	slope := func(x, z float64) float64 {
		dx := (H(x+0.6, z) - H(x-0.6, z)) / 1.2
		dz := (H(x, z+0.6) - H(x, z-0.6)) / 1.2
		return math.Hypot(dx, dz)
	}
	inLand := func(p Vec3) bool {
		h := H(p.X, p.Z)
		return h >= 0 && p.Y <= h && p.Y >= -0.3
	}
	g.Use(DarkStone)
	s.shell(g, site, 1.5, inLand)
	s.terrainColors(g, H, slope, site)
	s.river(g, H)
	s.volcano(g, H)
	s.rocks(g, H)

	s.animal(g, H, dinosaurValleyBrachioUV, P(-1, 0), 1.15, s.brachio)
	s.animal(g, H, dinosaurValleyRexUV, P(1, 0), 1.15, s.rex)
	s.animal(g, H, dinosaurValleyTriUV, P(-1, 0), 1.2, s.trike)
	s.animal(g, H, dinosaurValleyStegoUV, P(1, 0.08), 1.25, s.stego)
	s.animal(g, H, dinosaurValleyAnkyUV, P(1, 0.15), 1.05, s.anky)
	s.nest(g, H)
	for i, r := range dinosaurValleyRaptors {
		i := i
		s.animal(g, H, r.at, r.dir, 1.4, func(b dinosaurValleyBeast) { s.raptor(b, i) })
	}
	s.plants(g, H, slope)
	for i, pt := range []struct {
		uv, dir Vec2
		y       float64
	}{{P(-8, -22), P(1, 0.3), 32}, {P(6, -3), P(-1, 0.4), 27}, {P(27, -20), P(-0.6, -1), 24},
		{P(-30, 24), P(0.3, 1), 17}} {
		p := dinosaurValleyW(pt.uv.X, pt.uv.Y)
		s.pterosaur(g, p.At(pt.y), dinosaurValleyHeading(pt.dir.X, pt.dir.Y), i)
	}
	return Sweep(P(0, 0), dinosaurValleyHeading(0, 1))
}

// terrainColors paints the valley floor, banks and escarpment.
func (s DinosaurValley) terrainColors(g *Grid, H func(x, z float64) float64, slope func(x, z float64) float64, site Box) {
	top := func(p Vec3) bool { h := H(p.X, p.Z); return h >= 0 && p.Y > h-0.9 }
	g.Paint(Grass, Func{site, top})
	g.Paint(Foliage, Func{site, func(p Vec3) bool {
		return top(p) && dinosaurValleyFBM(p.X*0.11, p.Z*0.11, 31) > 0.15
	}})
	// Bare rock on steep slopes, banded escarpment.
	g.Paint(Stone, Func{site, func(p Vec3) bool {
		return p.Y > 1.2 && slope(p.X, p.Z) > 1.3 && s.volcanoH(p.X, p.Z) <= 0
	}})
	for i, c := range []Color{Sand, DarkStone, Sand, DarkStone} {
		y0 := 2.5 + float64(i)*2.4
		g.Paint(c, Func{site, func(p Vec3) bool {
			n := 0.5 * dinosaurValleyNoise(p.X*0.3, p.Z*0.3, 33+i)
			return p.Y > y0+n && p.Y < y0+1.0+n && slope(p.X, p.Z) > 1.3 && s.volcanoH(p.X, p.Z) <= 0
		}})
	}
	// Grassy cap on the plateau.
	g.Paint(Grass, Func{site, func(p Vec3) bool {
		return top(p) && H(p.X, p.Z) > dinosaurValleyPlateau-1.5 && slope(p.X, p.Z) < 1.0
	}})
	// Low outer rim in sandy soil.
	g.Paint(Sand, Func{site, func(p Vec3) bool {
		c := s.cell(g, p)
		return top(p) && H(p.X, p.Z) < 0.75 && s.volcanoH(p.X, p.Z) <= 0 && Hash01(int(c.X), int(c.Z), 37) < 0.7
	}})
}

// river fills the water, the falls and the foam.
func (s DinosaurValley) river(g *Grid, H func(x, z float64) float64) {
	box := Box{V(-52, -0.5, -52), V(52, dinosaurValleyPlateau+2, 52)}
	fallsD := dinosaurValleyFallsUV.Y
	// Water fills the channel up to a level: on the plateau it follows the bed.
	g.Use(Water).Fill(Func{box, func(p Vec3) bool {
		dist, along := s.polyline(p.Plan(), dinosaurValleyRiver)
		q := dinosaurValleyUV(p.Plan())
		pool := q.Sub(dinosaurValleyPoolUV).Length() < 5.2 && q.Y > fallsD+1
		if dist > 6 && !pool {
			return false
		}
		w := s.riverW(along)
		h := H(p.X, p.Z)
		if h < 0 {
			return false
		}
		if q.Y < fallsD {
			return dist < w+0.6 && p.Y > h && p.Y <= h+0.9
		}
		return (dist < w+1.2 || pool) && p.Y > h && p.Y <= dinosaurValleyWater
	}})
	// Banks: sand and pebbles along the water.
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		dist, along := s.polyline(p.Plan(), dinosaurValleyRiver)
		if dist > 6 {
			return false
		}
		c := s.cell(g, p)
		col, _ := g.ColorOf(c)
		return col != Water && dist < s.riverW(along)+3 && p.Y > H(p.X, p.Z)-0.9 && p.Y < 3 &&
			Hash01(int(c.X), int(c.Z), 41) < 0.75
	}})
	// The falls: a curtain of water in front of the cliff face.
	fu := dinosaurValleyFallsUV.X
	faceD := fallsD
	for d := fallsD - 4; d < fallsD+6; d += 0.25 {
		p := dinosaurValleyW(fu, d)
		if H(p.X, p.Y) < dinosaurValleyPlateau-3 {
			faceD = d
			break
		}
	}
	lipP := dinosaurValleyW(fu, faceD-2.5)
	lip := H(lipP.X, lipP.Y) + 0.8
	c := dinosaurValleyW(fu, faceD)
	fbox := Box{V(c.X-6, 0, c.Y-6), V(c.X+6, lip, c.Y+6)}
	curtain := func(p Vec3) bool {
		q := dinosaurValleyUV(p.Plan())
		drop := (lip - p.Y) / lip
		out := faceD - 1.6 + 1.4*math.Sqrt(math.Max(0, drop))
		half := 2.0 + 0.6*drop
		return math.Abs(q.X-fu) <= half && q.Y >= out && q.Y <= out+1.2
	}
	g.Use(Water).Fill(Func{fbox, curtain})
	g.Paint(White, Func{fbox, func(p Vec3) bool {
		cl := s.cell(g, p)
		return curtain(p) && (Hash01(int(cl.X), int(cl.Y), int(cl.Z), 43) < 0.18 || p.Y > lip-0.9)
	}})
	// Foam at the foot of the falls.
	foam := Union()
	for i := 0; i < 7; i++ {
		q := dinosaurValleyW(fu-2.4+0.8*float64(i), faceD+1.4+2.4*Hash01(i, 44))
		foam = append(foam, Sphere(q.At(dinosaurValleyWater+0.1), 0.9+0.5*Hash01(i, 45)))
	}
	g.Use(White).FillVolume(foam)
	// Ripples of white water along the river.
	g.Paint(White, Func{box, func(p Vec3) bool {
		dist, along := s.polyline(p.Plan(), dinosaurValleyRiver)
		if dist > 3 {
			return false
		}
		cl := s.cell(g, p)
		col, _ := g.ColorOf(cl)
		return col == Water && dist < 2.5 && Hash01(int(cl.X), int(cl.Z), 46) < 0.05+0.1*math.Max(0, math.Sin(along*0.9))
	}})
}

// volcano colors the cone, the crater lava lake and the flow, and adds the smoke column.
func (s DinosaurValley) volcano(g *Grid, H func(x, z float64) float64) {
	vc := dinosaurValleyVolcano
	box := Box{V(vc.X-dinosaurValleyCone-2, 0, vc.Y-dinosaurValleyCone-2),
		V(vc.X+dinosaurValleyCone+2, dinosaurValleyRim+2, vc.Y+dinosaurValleyCone+2)}
	on := func(p Vec3) bool { return s.volcanoH(p.X, p.Z) > 1.8 && p.Y > 1.4 }
	top := func(p Vec3) bool { return p.Y > H(p.X, p.Z)-0.9 }
	g.Paint(Charcoal, Func{box, on})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		a := p.Plan().Sub(vc).Angle()
		r := p.Plan().Sub(vc).Length()
		return on(p) && dinosaurValleyFBM(math.Cos(a)*6+r*0.06, math.Sin(a)*6+r*0.06, 51) > 0.05
	}})
	g.Paint(Brick, Func{box, func(p Vec3) bool {
		c := s.cell(g, p)
		return on(p) && p.Y > dinosaurValleyRim-5 && Hash01(int(c.X), int(c.Y), int(c.Z), 52) < 0.35
	}})
	// Vegetation creeping up the lower flanks.
	g.Paint(Foliage, Func{box, func(p Vec3) bool {
		c := s.cell(g, p)
		lim := 4 + 2.5*dinosaurValleyFBM(p.X*0.15, p.Z*0.15, 53)
		return on(p) && top(p) && p.Y < lim && Hash01(int(c.X), int(c.Z), 54) < 0.8
	}})
	// Crater: lava lake and glowing walls.
	g.Paint(Red, Func{box, func(p Vec3) bool {
		r := p.Plan().Sub(vc).Length()
		return r < dinosaurValleyCrater && p.Y < dinosaurValleyFloor+1.4
	}})
	g.Paint(Gold, Func{box, func(p Vec3) bool {
		r := p.Plan().Sub(vc).Length()
		return r < dinosaurValleyCrater-1.6 && p.Y < dinosaurValleyFloor+0.6 &&
			dinosaurValleyNoise(p.X*0.6, p.Z*0.6, 55) > -0.2
	}})
	g.Paint(Terracotta, Func{box, func(p Vec3) bool {
		r := p.Plan().Sub(vc).Length()
		c := s.cell(g, p)
		return r < dinosaurValleyCrater+0.5 && p.Y >= dinosaurValleyFloor+1.4 && p.Y < dinosaurValleyFloor+3 &&
			Hash01(int(c.X), int(c.Y), int(c.Z), 56) < 0.5
	}})
	// Lava flow down the flank: dark crust edges, red flow, gold core.
	path := s.lavaPath()
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		d, _ := s.polyline(p.Plan(), path)
		return top(p) && d < 2.6 && p.Y > 0.8
	}})
	g.Paint(Red, Func{box, func(p Vec3) bool {
		d, along := s.polyline(p.Plan(), path)
		c := s.cell(g, p)
		w := 1.0 + along*0.05
		return top(p) && p.Y > 0.8 && (d < w || d < 2.6 && Hash01(int(c.X), int(c.Y), int(c.Z), 57) < 0.15)
	}})
	g.Paint(Gold, Func{box, func(p Vec3) bool {
		d, along := s.polyline(p.Plan(), path)
		w := 1.0 + along*0.05
		return top(p) && p.Y > 0.8 && d < 0.5*w && Hash01(int(math.Floor(along/1.4)), 58) < 0.7
	}})
	// The flow bulges a little proud of the flank so it reads in profile.
	g.Use(Red).Fill(Func{box, func(p Vec3) bool {
		d, along := s.polyline(p.Plan(), path)
		w := 0.8 + along*0.04
		h := H(p.X, p.Z)
		return d < w && p.Y > h && p.Y <= h+0.7 && p.Y > 0.8 && along > 1.5
	}})
	// Smoke column rising and drifting to the right of the hero view.
	var puffs []Ellipsoid
	for i := 0; i < 10; i++ {
		t := float64(i) / 9
		q := dinosaurValleyUV(vc).Add(P(-5*t*t+1.4*(Hash01(i, 61)-0.5), -7*t+1.4*(Hash01(i, 62)-0.5)))
		c := dinosaurValleyW(q.X, q.Y)
		r := 1.8 + 1.6*t + 0.5*Hash01(i, 63)
		puffs = append(puffs, Ellipsoid{c.At(dinosaurValleyRim + 1.5 + 16*t), V(r*1.15, r*0.9, r*1.15)})
	}
	in := func(p Vec3) bool {
		for _, e := range puffs {
			if e.Contains(p) {
				return true
			}
		}
		return false
	}
	pbox := Box{V(vc.X-16, dinosaurValleyRim, vc.Y-16), V(vc.X+16, dinosaurValleyRim+24, vc.Y+16)}
	g.Use(Stone)
	s.shell(g, pbox, 1.2, in)
	g.Paint(DarkStone, Func{pbox, func(p Vec3) bool {
		c := s.cell(g, p)
		return in(p) && (p.Y < dinosaurValleyRim+5 || Hash01(int(c.X), int(c.Y), int(c.Z), 64) < 0.15)
	}})
	g.Paint(White, Func{pbox, func(p Vec3) bool {
		c := s.cell(g, p)
		return in(p) && p.Y > dinosaurValleyRim+12 && Hash01(int(c.X), int(c.Y), int(c.Z), 65) < 0.3
	}})
}

// rocks scatters boulders on the valley floor.
func (s DinosaurValley) rocks(g *Grid, H func(x, z float64) float64) {
	spots := []Vec2{P(-22, 22), P(-36, 14), P(24, 22), P(38, -10), P(4, -22), P(-40, 2), P(-12, 30),
		P(27, 6), P(10, -18), P(-30, -12), P(40, 4), P(-4, 16)}
	for i, q := range spots {
		p := dinosaurValleyW(q.X, q.Y)
		h := H(p.X, p.Y)
		r := 1.2 + 1.1*Hash01(i, 71)
		col := DarkStone
		if i%3 == 0 {
			col = Stone
		}
		g.Use(col).FillVolume(Ellipsoid{p.At(h + 0.1), V(r*1.3, r*0.8, r)},
			Ellipsoid{p.Add(P(r, 0.6)).At(h), V(r*0.7, r*0.55, r*0.7)})
	}
}

// brachio is the long-necked sauropod with its head in the araucaria crowns.
func (s DinosaurValley) brachio(b dinosaurValleyBeast) {
	body := Union(
		Ellipsoid{V(3, 12, 0), V(6, 5, 4.4)},
		Ellipsoid{V(-3.5, 10.5, 0), V(5.8, 4.4, 4.1)},
	)
	legs := Union()
	for _, side := range []float64{-1, 1} {
		z := 2.6 * side
		legs = append(legs,
			s.chain([]Vec3{V(4.5, 11, z), V(4.8, 5.5, z), V(5.2, -0.6, z)}, []float64{1.9, 1.35, 1.3}),
			s.chain([]Vec3{V(-5, 9.5, z*0.95), V(-3.8, 5, z*0.95), V(-4.6, -0.6, z*0.95)}, []float64{2.2, 1.5, 1.4}),
			Ellipsoid{V(5.5, 0, z), V(1.6, 0.8, 1.5)},
			Ellipsoid{V(-4.2, 0, z*0.95), V(1.8, 0.8, 1.6)},
		)
	}
	neck := s.chain([]Vec3{V(5.5, 13.5, 0), V(8.5, 18, 0), V(10.6, 22.5, 0), V(12, 27, 0), V(13, 30.5, 0)},
		[]float64{3.2, 2.2, 1.7, 1.35, 1.15})
	head := Union(
		Ellipsoid{V(14.2, 31.2, 0), V(2.1, 1.25, 1.15)},
		dinosaurValleyTaper{V(14.5, 31, 0), V(16.6, 30.3, 0), 0.95, 0.75},
		Sphere(V(13.8, 32.3, 0), 0.95),
	)
	tail := s.chain([]Vec3{V(-8, 10.5, 0), V(-13, 8.5, -1.5), V(-18, 6, -3), V(-22, 4, -5), V(-24.5, 3, -7.5)},
		[]float64{3.0, 2.1, 1.4, 0.9, 0.55})
	b.vol(Slate, body, legs, neck, head, tail)
	// Pale belly and throat, darker dappled back.
	all := Box{V(-30, -2, -16), V(20, 36, 8)}
	b.paint(Stone, all, func(p Vec3) bool {
		switch {
		case p.X > 5.5 && p.Y > 13:
			// Throat: the underside of the neck.
			t := (p.Y - 13.5) / 17
			cx := dinosaurValleyLerp(5.5, 13, t) + 0.6
			return p.X > cx-0.2 && p.Y < 30
		case p.Y < 9.5:
			return Ellipsoid{V(3, 12, 0), V(6.6, 5.6, 5)}.Contains(p) || Ellipsoid{V(-3.5, 10.5, 0), V(6.4, 5, 4.7)}.Contains(p)
		}
		return false
	})
	b.paint(DarkStone, all, func(p Vec3) bool {
		return p.Y > 14.5 && p.X < 7 && b.cellHash(p, 81) < 0.4
	})
	b.paint(Charcoal, all, func(p Vec3) bool {
		return math.Abs(p.Y-31.6) < 0.45 && math.Abs(p.X-15.1) < 0.45 && math.Abs(math.Abs(p.Z)-1.0) < 0.6
	})
	b.paint(Charcoal, all, func(p Vec3) bool { return p.Y < 0.6 })
}

// rex is the striding theropod with open jaws, tiny arms and a balancing tail.
func (s DinosaurValley) rex(b dinosaurValleyBeast) {
	body := Union(
		Ellipsoid{V(1.5, 9, 0), V(5.5, 3.3, 2.9)},
		Ellipsoid{V(5, 8.6, 0), V(3.3, 3.2, 2.7)},
	)
	neck := s.chain([]Vec3{V(6.5, 9.4, 0), V(8.5, 10.6, 0), V(9.8, 11.2, 0)}, []float64{2.3, 2.0, 1.8})
	skull := Union(
		Ellipsoid{V(11, 11.7, 0), V(2.4, 1.8, 1.7)},
		dinosaurValleyTaper{V(11, 12, 0), V(15.2, 12.2, 0), 1.55, 1.0},
	)
	jaw := Union(
		dinosaurValleyTaper{V(10.4, 10.3, 0), V(14.4, 8.6, 0), 1.1, 0.7},
		Ellipsoid{V(10.8, 10.3, 0), V(1.6, 1.0, 1.3)},
	)
	tail := s.chain([]Vec3{V(-3, 9.3, 0), V(-7.5, 9.0, 0.6), V(-12, 8.4, 1.8), V(-16, 7.6, 3.2), V(-19.5, 7, 4.8)},
		[]float64{2.7, 2.0, 1.35, 0.85, 0.55})
	// Striding legs: the left foot planted ahead, the right pushing off behind.
	legs := Union(
		s.chain([]Vec3{V(1.2, 8.4, -2.0), V(4.2, 5.0, -2.3), V(3.2, 1.6, -2.4), V(5.2, 0.3, -2.4)}, []float64{2.4, 1.4, 0.85, 0.75}),
		s.chain([]Vec3{V(0.8, 8.4, 2.0), V(1.6, 4.6, 2.3), V(-2.3, 2.0, 2.4), V(-1.6, 0.0, 2.4)}, []float64{2.4, 1.4, 0.85, 0.75}),
		Ellipsoid{V(6.3, 0.2, -2.4), V(1.6, 0.6, 1.0)},
		Ellipsoid{V(-0.6, 0.0, 2.4), V(1.4, 0.6, 1.0)},
	)
	arms := Union(
		s.chain([]Vec3{V(6.8, 7.4, -1.9), V(7.6, 6.2, -2.3), V(8.4, 6.5, -2.2)}, []float64{0.55, 0.5, 0.45}),
		s.chain([]Vec3{V(6.8, 7.4, 1.9), V(7.8, 6.4, 2.3), V(8.6, 6.8, 2.2)}, []float64{0.55, 0.5, 0.45}),
	)
	b.vol(Brick, body, neck, skull, tail, legs, jaw)
	b.fine(Brick, arms)
	// Mouth: carve the gape between the jaws and line it.
	b.g.Carve(b.w(Func{Box{V(11.5, 8, -2), V(16, 12, 2)}, func(p Vec3) bool {
		top := 11.4 + (p.X-11.5)*0.05
		bot := 10.6 - (p.X-11.5)*0.42
		return p.Y < top && p.Y > bot && math.Abs(p.Z) < 0.9
	}}))
	b.paint(Red, Box{V(10, 8, -2), V(16, 12, 2)}, func(p Vec3) bool {
		bot := 10.6 - (p.X-11.5)*0.42
		return p.X > 10.8 && p.Y < bot+0.9 && p.Y > bot-0.6 && math.Abs(p.Z) < 0.9
	})
	teeth := Union()
	for i := 0; i < 5; i++ {
		x := 12 + 0.75*float64(i)
		for _, z := range []float64{-1.0, 1.0} {
			topY := 11.4 + (x-11.5)*0.05
			botY := 10.6 - (x-11.5)*0.42
			teeth = append(teeth, Capsule{V(x, topY+0.1, z), V(x+0.1, topY-0.45, z), 0.3})
			if i < 4 {
				teeth = append(teeth, Capsule{V(x+0.3, botY-0.1, z*0.9), V(x+0.3, botY+0.45, z*0.9), 0.3})
			}
		}
	}
	b.fine(White, teeth)
	all := Box{V(-24, -1, -6), V(17, 15, 8)}
	// Belly and throat in terracotta, dark stripes over the back.
	b.paint(Terracotta, all, func(p Vec3) bool {
		return (p.Y < 7.6 && p.Y > 4.5 && p.X > -4) || (p.X > 8.5 && p.Y < 10.2 && p.X < 12)
	})
	b.paint(Wood, all, func(p Vec3) bool {
		return p.Y > 10.3 && p.X < 9 && math.Sin(p.X*1.3) > 0.45
	})
	b.paint(Wood, all, func(p Vec3) bool {
		return p.X < -3 && p.Y > 7.8 && math.Sin(p.X*1.3) > 0.45
	})
	// Eye and brow.
	b.paint(Gold, all, func(p Vec3) bool {
		return math.Abs(p.Y-12.8) < 0.45 && math.Abs(p.X-11.6) < 0.45 && math.Abs(p.Z) > 0.9
	})
	b.paint(Charcoal, all, func(p Vec3) bool { return p.Y < 0.7 && p.Y > -1 })
	claws := Union(
		Capsule{V(7.2, 0.3, -2.4), V(7.8, 0.1, -2.4), 0.35},
		Capsule{V(8.6, 6.4, -2.2), V(8.9, 5.9, -2.2), 0.3},
		Capsule{V(8.8, 6.7, 2.2), V(9.1, 6.2, 2.2), 0.3},
	)
	b.fine(White, claws)
}

// trike is the three-horned ceratopsian with its frill raised.
func (s DinosaurValley) trike(b dinosaurValleyBeast) {
	body := Union(
		Ellipsoid{V(0, 4.6, 0), V(5.6, 3.3, 3.3)},
		Ellipsoid{V(-2, 5.0, 0), V(3.6, 3.0, 3.0)},
	)
	legs := Union()
	for _, side := range []float64{-1, 1} {
		z := 2.3 * side
		legs = append(legs,
			s.chain([]Vec3{V(3.4, 3.8, z), V(3.6, 1.8, z*1.1), V(3.8, -0.5, z*1.1)}, []float64{1.4, 1.1, 1.1}),
			s.chain([]Vec3{V(-3, 4.5, z), V(-2.6, 2, z), V(-3.2, -0.5, z)}, []float64{1.7, 1.25, 1.2}),
		)
	}
	head := Union(
		Ellipsoid{V(7, 4.2, 0), V(2.6, 1.8, 1.7)},
		dinosaurValleyTaper{V(8, 3.8, 0), V(10.6, 2.6, 0), 1.25, 0.55},
		s.chain([]Vec3{V(5, 4.6, 0), V(6.2, 4.4, 0)}, []float64{2.2, 2.0}),
	)
	tail := s.chain([]Vec3{V(-5, 4.6, 0), V(-8.5, 3.6, 0.8), V(-11.5, 2.2, 2.8)}, []float64{1.8, 1.0, 0.55})
	b.vol(Sand, body, legs, head, tail)
	// The frill: a tilted shield rising behind the head.
	n := V(0.8, 0.6, 0)
	c := V(5.4, 6.8, 0)
	frill := Func{Box{V(1.5, 2.5, -4.6), V(8, 11, 4.6)}, func(p Vec3) bool {
		d := p.Sub(c)
		along := d.Dot(n)
		radial := d.Sub(n.Scale(along))
		rr := math.Hypot(radial.X, radial.Y)
		ry := math.Abs(radial.Z)
		e := math.Hypot(rr/4.0, ry/4.2)
		return math.Abs(along) <= 0.5 && e <= 1
	}}
	b.vol(Terracotta, frill)
	b.paint(Red, Box{V(1.5, 2.5, -4.8), V(8.5, 11.5, 4.8)}, func(p Vec3) bool {
		d := p.Sub(c)
		along := d.Dot(n)
		radial := d.Sub(n.Scale(along))
		e := math.Hypot(math.Hypot(radial.X, radial.Y)/4.0, radial.Z/4.2)
		return e > 0.4 && e < 0.62 && along > -0.2
	})
	// Knobs around the frill edge.
	knobs := Union()
	for i := 0; i < 11; i++ {
		a := -math.Pi*0.5 + math.Pi*float64(i)/10
		// Edge point in the frill plane: up direction perpendicular to n within XY, and Z.
		up := V(-0.6, 0.8, 0)
		q := c.Add(up.Scale(4.0 * math.Sin(a+math.Pi/2))).Add(V(0, 0, 4.2*math.Cos(a+math.Pi/2)))
		if q.Y < 5 {
			continue
		}
		knobs = append(knobs, Sphere(q, 0.55))
	}
	b.vol(Brick, knobs)
	// Horns: two long brow horns and a short nose horn.
	horns := Union(
		dinosaurValleyTaper{V(7.4, 5.3, -0.9), V(11.2, 7.3, -1.3), 0.6, 0.3},
		dinosaurValleyTaper{V(7.4, 5.3, 0.9), V(11.2, 7.3, 1.3), 0.6, 0.3},
		dinosaurValleyTaper{V(9.3, 3.9, 0), V(10.1, 5.2, 0), 0.5, 0.3},
	)
	b.fine(White, horns)
	all := Box{V(-14, -1, -6), V(12, 12, 6)}
	b.paint(Stone, all, func(p Vec3) bool { return p.Y < 3 && p.Y > 1.4 && math.Abs(p.X) < 5 })
	b.paint(Wood, all, func(p Vec3) bool {
		return p.Y > 6.5 && p.X < 3.5 && math.Sin(p.X*1.4+p.Z*0.3) > 0.55
	})
	b.paint(Charcoal, all, func(p Vec3) bool {
		return math.Abs(p.Y-5.0) < 0.45 && math.Abs(p.X-7.4) < 0.5 && math.Abs(p.Z) > 1.2 && math.Abs(p.Z) < 2.0
	})
	b.paint(Charcoal, all, func(p Vec3) bool { return p.Y < 0.4 })
	b.paint(DarkStone, all, func(p Vec3) bool { return p.X > 9.4 && p.Y < 3.2 })
}

// stego is the plated stegosaur with its spiked tail, head down at the water.
func (s DinosaurValley) stego(b dinosaurValleyBeast) {
	backY := func(x float64) float64 { // Top of the arched back.
		return 5.2 + 3.6*math.Exp(-(x+1)*(x+1)/30) - 0.05*x
	}
	body := Union(
		Ellipsoid{V(-0.5, 5.4, 0), V(6.2, 3.4, 2.9)},
		Ellipsoid{V(3, 4.3, 0), V(3.4, 2.5, 2.4)},
	)
	legs := Union()
	for _, side := range []float64{-1, 1} {
		z := 1.9 * side
		legs = append(legs,
			s.chain([]Vec3{V(-2.6, 5, z), V(-2.0, 2.5, z), V(-2.8, -0.5, z)}, []float64{1.6, 1.15, 1.1}),
			s.chain([]Vec3{V(3.6, 3.4, z), V(3.8, 1.5, z*1.05), V(4.0, -0.5, z*1.05)}, []float64{1.15, 0.9, 0.9}),
		)
	}
	neck := s.chain([]Vec3{V(5.5, 4.2, 0), V(7.6, 2.8, 0), V(9.0, 2.0, 0)}, []float64{1.5, 1.05, 0.85})
	head := Union(Ellipsoid{V(9.8, 1.7, 0), V(1.5, 0.8, 0.75)})
	tail := s.chain([]Vec3{V(-6, 5.6, 0), V(-9.5, 4.8, -0.6), V(-13, 3.6, -1.6), V(-16, 2.8, -2.8)},
		[]float64{2.1, 1.35, 0.85, 0.55})
	b.vol(Wood, body, legs, neck, head, tail)
	all := Box{V(-19, -1, -7), V(12, 14, 6)}
	b.paint(DarkStone, all, func(p Vec3) bool { return p.Y > 6 && b.cellHash(p, 91) < 0.3 })
	b.paint(Sand, all, func(p Vec3) bool { return p.Y < 3.6 && p.Y > 1.5 && math.Abs(p.X) < 6 })
	// Two staggered rows of kite-shaped plates along the spine.
	plates := Union()
	k := 0
	for x := -12.5; x <= 6.0; x += 1.5 {
		side := 0.5
		if k%2 == 1 {
			side = -0.5
		}
		size := 1.2 + 2.4*math.Exp(-(x+1.5)*(x+1.5)/26)
		base := backY(x) - 0.6
		if x < -7 {
			base = 4.6 + (x+7)*0.32 + 0.5
		}
		cx, cy := x, base+size*0.7
		h, w := size, size*0.62
		plates = append(plates, Func{Box{V(cx-w, base-1, side-0.5), V(cx+w, cy+h, side+0.5)}, func(p Vec3) bool {
			return math.Abs(p.X-cx)/w+math.Abs(p.Y-cy)/h <= 1 && math.Abs(p.Z-side) <= 0.55
		}})
		k++
	}
	for i, p := range plates {
		c := Terracotta
		if i%2 == 0 {
			c = Red
		}
		b.fine(c, p)
	}
	// Thagomizer: four tail spikes.
	spikes := Union(
		dinosaurValleyTaper{V(-14, 3.4, -2.1), V(-15.5, 5.4, -0.2), 0.45, 0.25},
		dinosaurValleyTaper{V(-14, 3.4, -2.1), V(-15.0, 4.2, -4.4), 0.45, 0.25},
		dinosaurValleyTaper{V(-15.3, 3.0, -2.5), V(-17.0, 4.8, -1.0), 0.45, 0.25},
		dinosaurValleyTaper{V(-15.3, 3.0, -2.5), V(-16.6, 3.8, -5.0), 0.45, 0.25},
	)
	b.fine(White, spikes)
	b.paint(Charcoal, all, func(p Vec3) bool {
		return math.Abs(p.Y-2.0) < 0.4 && math.Abs(p.X-10.0) < 0.45 && math.Abs(p.Z) > 0.4
	})
	b.paint(Charcoal, all, func(p Vec3) bool { return p.Y < 0.4 })
}

// anky is a low armored ankylosaur with osteoderm rows, side spikes and a tail club.
func (s DinosaurValley) anky(b dinosaurValleyBeast) {
	body := Union(
		Ellipsoid{V(0, 3.4, 0), V(5.2, 2.4, 3.4)},
		Ellipsoid{V(4.2, 3.0, 0), V(2.4, 1.8, 2.4)},
	)
	legs := Union()
	for _, side := range []float64{-1, 1} {
		z := 2.3 * side
		legs = append(legs,
			s.chain([]Vec3{V(3.2, 2.6, z), V(3.4, 1.0, z*1.1), V(3.5, -0.5, z*1.1)}, []float64{1.1, 0.9, 0.9}),
			s.chain([]Vec3{V(-3, 2.8, z), V(-2.8, 1.0, z*1.1), V(-3.1, -0.5, z*1.1)}, []float64{1.2, 1.0, 1.0}),
		)
	}
	head := Union(
		Ellipsoid{V(6.8, 2.6, 0), V(1.7, 1.1, 1.5)},
		dinosaurValleyTaper{V(7.5, 2.4, 0), V(8.6, 2.0, 0), 0.9, 0.6},
	)
	tail := s.chain([]Vec3{V(-4.5, 3.4, 0), V(-8.5, 2.8, -0.6), V(-12.5, 2.3, -1.6)}, []float64{1.5, 0.8, 0.55})
	b.vol(DarkStone, body, legs, head, tail)
	b.vol(Sand, Ellipsoid{V(-13.4, 2.3, -1.9), V(1.2, 0.8, 1.5)})
	// Osteoderm bumps in rows over the back.
	bumps := Union()
	for i := 0; i < 7; i++ {
		x := -4.2 + 1.4*float64(i)
		for _, a := range []float64{-0.75, -0.25, 0.25, 0.75} {
			ry := 2.4 * math.Sqrt(math.Max(0, 1-(x/5.2)*(x/5.2)))
			y := 3.4 + ry*math.Cos(a)
			z := 3.4 * math.Sqrt(math.Max(0, 1-(x/5.2)*(x/5.2))) * math.Sin(a)
			bumps = append(bumps, Sphere(V(x, y, z), 0.5))
		}
	}
	b.fine(Sand, bumps)
	// Side spikes and head horns.
	spikes := Union()
	for i := 0; i < 6; i++ {
		x := -3.8 + 1.5*float64(i)
		e := math.Sqrt(math.Max(0, 1-(x/5.2)*(x/5.2)))
		for _, side := range []float64{-1, 1} {
			root := V(x, 3.2, side*3.2*e)
			spikes = append(spikes, dinosaurValleyTaper{root, root.Add(V(-0.4, 0.2, side*1.3)), 0.45, 0.25})
		}
	}
	for _, side := range []float64{-1, 1} {
		spikes = append(spikes, dinosaurValleyTaper{V(5.8, 3.2, side*1.3), V(5.0, 3.6, side*2.2), 0.4, 0.25})
	}
	b.fine(Stone, spikes)
	all := Box{V(-16, -1, -6), V(10, 7, 6)}
	b.paint(Sand, all, func(p Vec3) bool { return p.Y > 1.0 && p.Y < 2.2 && math.Abs(p.X) < 4.5 })
	b.paint(Charcoal, all, func(p Vec3) bool {
		return math.Abs(p.Y-2.9) < 0.4 && math.Abs(p.X-7.2) < 0.45 && math.Abs(p.Z) > 0.8
	})
	b.paint(Charcoal, all, func(p Vec3) bool { return p.Y < 0.4 })
}

// raptor is a small feathered-looking hunter in mid-run.
func (s DinosaurValley) raptor(b dinosaurValleyBeast, seed int) {
	stride := 0.8 * (Hash01(seed, 1) - 0.5)
	body := Union(
		Ellipsoid{V(0, 3.0, 0), V(2.0, 1.1, 1.0)},
		s.chain([]Vec3{V(1.6, 3.3, 0), V(2.6, 4.3, 0)}, []float64{0.75, 0.6}),
		Ellipsoid{V(3.4, 4.5, 0), V(1.2, 0.6, 0.55)},
		s.chain([]Vec3{V(-1.6, 3.2, 0), V(-4, 3.5, 0), V(-6.5, 3.7, 0.4)}, []float64{0.85, 0.6, 0.45}),
		s.chain([]Vec3{V(-0.2, 2.7, -0.6), V(0.8+stride, 1.6, -0.7), V(0.2+stride, 0.4, -0.7), V(1.2+stride, -0.2, -0.7)}, []float64{0.8, 0.5, 0.45, 0.4}),
		s.chain([]Vec3{V(-0.2, 2.7, 0.6), V(0.4-stride, 1.6, 0.7), V(-0.8-stride, 0.6, 0.7), V(-0.2-stride, -0.2, 0.7)}, []float64{0.8, 0.5, 0.45, 0.4}),
	)
	col := Sand
	if seed%2 == 1 {
		col = Terracotta
	}
	b.vol(col, body)
	arms := Union(
		Capsule{V(1.4, 2.8, -0.7), V(2.2, 2.2, -0.9), 0.4},
		Capsule{V(1.4, 2.8, 0.7), V(2.2, 2.2, 0.9), 0.4},
	)
	b.fine(col, arms)
	all := Box{V(-8, -1, -2), V(5, 6, 2)}
	b.paint(Brick, all, func(p Vec3) bool { return p.Y > 3.4 && math.Sin(p.X*2.2) > 0.2 })
	b.paint(Charcoal, all, func(p Vec3) bool {
		return math.Abs(p.Y-4.7) < 0.35 && math.Abs(p.X-3.3) < 0.4
	})
}

// nest is a ring of sticks and soil holding white eggs.
func (s DinosaurValley) nest(g *Grid, H func(x, z float64) float64) {
	c := dinosaurValleyW(dinosaurValleyNestUV.X, dinosaurValleyNestUV.Y)
	h := H(c.X, c.Y)
	g.Use(Wood).FillVolume(Subtract(Cylinder{Center: c, Radius: 3.2, Y0: h - 0.4, Y1: h + 1.0},
		Cylinder{Center: c, Radius: 2.0, Y0: h + 0.2, Y1: h + 2}))
	g.Paint(Sand, Func{Box{V(c.X-4, h-1, c.Y-4), V(c.X+4, h+2, c.Y+4)}, func(p Vec3) bool {
		cell := s.cell(g, p)
		return Hash01(int(cell.X), int(cell.Y), int(cell.Z), 101) < 0.3
	}})
	eggs := Union()
	for i := 0; i < 6; i++ {
		a := float64(i) * 2 * math.Pi / 6
		q := Polar(c, 1.15, a)
		eggs = append(eggs, Ellipsoid{q.At(h + 0.9), V(0.65, 0.85, 0.65)})
	}
	eggs = append(eggs, Ellipsoid{c.At(h + 1.0), V(0.65, 0.85, 0.65)})
	g.Use(White).FillVolume(eggs)
}

// pterosaur is a crested flier gliding on broad, slightly raised wings.
func (s DinosaurValley) pterosaur(g *Grid, at Vec3, heading float64, seed int) {
	b := dinosaurValleyBeast{g, at, heading, 1}
	body := Union(
		dinosaurValleyTaper{V(-1.4, 0, 0), V(1.0, 0.1, 0), 0.6, 0.55},
		Ellipsoid{V(1.7, 0.3, 0), V(0.9, 0.5, 0.45)},
		dinosaurValleyTaper{V(2.2, 0.3, 0), V(3.8, 0.0, 0), 0.4, 0.3},
		dinosaurValleyTaper{V(1.5, 0.5, 0), V(-0.4, 1.4, 0), 0.4, 0.3},
	)
	b.fine(Wood, body)
	wing := Func{Box{V(-2, -1, -7), V(2, 3, 7)}, func(p Vec3) bool {
		az := math.Abs(p.Z)
		if az < 0.3 || az > 6.5 {
			return false
		}
		// Planform: broad at the root, swept point at the tip.
		lead := 0.9 + 0.15*az
		trail := -1.6 + 0.24*az
		if p.X > lead || p.X < trail {
			return false
		}
		y := 0.12*az + 0.04*az*az*(0.5+0.3*float64(seed%2)) - 0.6
		return math.Abs(p.Y-y) <= 0.45
	}}
	b.fine(Terracotta, wing)
	b.paint(Brick, Box{V(-2, -1, -7), V(2, 3, 7)}, func(p Vec3) bool { return math.Abs(p.Z) > 4.8 })
}

// plants places araucarias, conifers, tree ferns, cycads and fern clumps.

// plants places araucarias, conifers, tree ferns, cycads and fern clumps.
func (s DinosaurValley) plants(g *Grid, H func(x, z float64) float64, slope func(x, z float64) float64) {
	type zone struct {
		c Vec2 // View frame.
		r float64
	}
	// Keep plants out of the animals' footprints (view frame).
	keep := []zone{
		{dinosaurValleyBrachioUV, 9}, {dinosaurValleyBrachioUV.Add(P(-10, 0)), 5}, {dinosaurValleyBrachioUV.Add(P(16, 3)), 6},
		{dinosaurValleyBrachioUV.Add(P(25, 7)), 4},
		{dinosaurValleyRexUV, 9}, {dinosaurValleyRexUV.Add(P(12, 0)), 6}, {dinosaurValleyRexUV.Add(P(-14, 2)), 6},
		{dinosaurValleyRexUV.Add(P(-22, 5)), 4},
		{dinosaurValleyTriUV, 9}, {dinosaurValleyTriUV.Add(P(-11, 0)), 5}, {dinosaurValleyTriUV.Add(P(13, -3)), 4},
		{dinosaurValleyStegoUV, 8}, {dinosaurValleyStegoUV.Add(P(-10, 1)), 5}, {dinosaurValleyStegoUV.Add(P(13, -2)), 5},
		{dinosaurValleyNestUV, 5}, {dinosaurValleyPoolUV, 6},
		{dinosaurValleyAnkyUV, 7}, {dinosaurValleyAnkyUV.Add(P(-12, -1)), 4}, {dinosaurValleyAnkyUV.Add(P(7, 1)), 4},
	}
	for _, r := range dinosaurValleyRaptors {
		keep = append(keep, zone{r.at, 4}, zone{r.at.Sub(r.dir.Scale(4)), 3})
	}
	lava := s.lavaPath()
	free := func(p Vec2, margin float64) bool {
		q := dinosaurValleyUV(p)
		for _, z := range keep {
			if q.Sub(z.c).Length() < z.r+margin {
				return false
			}
		}
		if d, _ := s.polyline(p, dinosaurValleyRiver); d < 4.5 {
			return false
		}
		if d, _ := s.polyline(p, lava); d < 4 {
			return false
		}
		return s.siteR(p.X, p.Y) < 0.9 && s.volcanoH(p.X, p.Y) < 3
	}
	// Araucarias the sauropod browses: tall bare trunks with layered umbrella crowns.
	for i, t := range []struct {
		at Vec2
		h  float64
	}{
		{P(-22, -11), 32}, {P(-16, -19), 28}, {P(-26, -4), 22}, {P(-38, 8), 20}, {P(-6, -35), 22},
		{P(26, -29), 21}, {P(-38, -12), 17}, {P(38, -20), 19},
	} {
		p := dinosaurValleyW(t.at.X, t.at.Y)
		s.araucaria(g, p.At(H(p.X, p.Y)-0.5), t.h, i)
	}
	// Conifers on the plateau, along the back and around the sides.
	for i, q := range []Vec2{P(20, -31), P(32, -24), P(36, -12), P(13, -36), P(30, -34), P(42, -2),
		P(-44, -4), P(-42, 6), P(44, 6), P(4, -37), P(36, -27), P(24, -38), P(-40, 16)} {
		p := dinosaurValleyW(q.X, q.Y)
		h := H(p.X, p.Y)
		if h < 0 {
			continue
		}
		s.conifer(g, p.At(h-0.4), 9+5*Hash01(i, 111), i)
	}
	// A ring of forest around the back and sides of the valley.
	for i := 0; i < 40; i++ {
		a := float64(i)*2*math.Pi/40 + 0.08*Hash01(i, 112)
		rr := 0.8 + 0.08*Hash01(i, 113)
		q := P(rr*dinosaurValleySiteA*math.Cos(a), rr*dinosaurValleySiteB*math.Sin(a))
		if q.Y > 10 && math.Abs(q.X) < 34 {
			continue // Keep the front open.
		}
		p := dinosaurValleyW(q.X, q.Y)
		h := H(p.X, p.Y)
		if h < 0.4 || !free(p, 2) || s.volcanoH(p.X, p.Y) > 0.5 || slope(p.X, p.Y) > 1.2 {
			continue
		}
		if Hash01(i, 114) < 0.35 {
			s.treeFern(g, p.At(h-0.3), 5+2*Hash01(i, 115), 200+i)
			continue
		}
		s.conifer(g, p.At(h-0.4), 8+5*Hash01(i, 116), 100+i)
	}
	// Tree ferns, cycads and ground ferns: dense around the edges, sparse in the meadow.
	n := 0
	for i := -46; i <= 46; i += 3 {
		for k := -38; k <= 38; k += 3 {
			u, d := float64(i)+2.4*Hash01(i, k, 121), float64(k)+2.4*Hash01(i, k, 122)
			p := dinosaurValleyW(u, d)
			h := H(p.X, p.Y)
			meadow := math.Abs(u) < 30 && d > -8 && d < 26
			skip := 0.25
			if meadow {
				skip = 0.75
			}
			if h < 0.6 || slope(p.X, p.Y) > 0.9 || Hash01(i, k, 123) < skip {
				continue
			}
			r := Hash01(i, k, 124)
			if meadow {
				r = 0.55 + 0.45*r // Low plants only.
			}
			switch {
			case r < 0.25 && free(p, 2):
				s.treeFern(g, p.At(h-0.3), 4.5+2.5*Hash01(i, k, 125), n)
			case r < 0.55 && free(p, 1):
				s.cycad(g, p.At(h-0.3), 1.6+1.6*Hash01(i, k, 126), n)
			case r >= 0.55 && free(p, 0):
				s.fern(g, p.At(h-0.2), 1.6+1.0*Hash01(i, k, 127), n)
			}
			n++
		}
	}
}

func (s DinosaurValley) araucaria(g *Grid, base Vec3, h float64, seed int) {
	g.Use(Wood).Fill(Frustum{base.Plan(), 0.95, 0.6, base.Y, base.Y + h - 1})
	crown := base.Y + h*0.62
	for k := 0; k < 5; k++ {
		t := float64(k) / 4
		y := crown + t*(h*0.38-1.5)
		r := dinosaurValleyLerp(4.6, 1.8, t) * (0.9 + 0.2*Hash01(seed, k))
		col := Foliage
		if k%2 == 1 {
			col = Grass
		}
		g.Use(col).FillVolume(Frustum{base.Plan(), r, r * 0.45, y, y + 1.4})
	}
	g.Use(Foliage).FillVolume(Ellipsoid{base.Add(V(0, h-0.4, 0)), V(1.6, 1.2, 1.6)})
}

// conifer is a tiered pine-like tree.
func (s DinosaurValley) conifer(g *Grid, base Vec3, h float64, seed int) {
	g.Use(Wood).Fill(Cylinder{Center: base.Plan(), Radius: 0.6, Y0: base.Y, Y1: base.Y + h*0.5})
	for k := 0; k < 3; k++ {
		y := base.Y + h*(0.25+0.22*float64(k))
		r := h * (0.3 - 0.07*float64(k))
		g.Use(Foliage).FillVolume(Frustum{base.Plan(), r, 0.4, y, y + h*0.38})
	}
}

// treeFern has a slim trunk and a crown of arching fronds.
func (s DinosaurValley) treeFern(g *Grid, base Vec3, h float64, seed int) {
	lean := Polar(P(0, 0), 0.6*Hash01(seed, 131), Hash01(seed, 132)*2*math.Pi)
	top := base.Add(V(lean.X, h, lean.Y))
	g.Use(Wood).Fill(Capsule{base, top, 0.55})
	phase := Hash01(seed, 133) * math.Pi
	for k := 0; k < 8; k++ {
		a := phase + float64(k)*2*math.Pi/8
		mid := Polar(top.Plan(), 1.9, a).At(top.Y + 0.9)
		tip := Polar(top.Plan(), 3.6, a).At(top.Y - 0.4)
		col := Grass
		if k%2 == 1 {
			col = Foliage
		}
		g.Use(col).Fill(Capsule{top, mid, 0.5}, Capsule{mid, tip, 0.45})
	}
}

// cycad is a stout pineapple trunk with a stiff rosette of fronds and a gold cone.
func (s DinosaurValley) cycad(g *Grid, base Vec3, h float64, seed int) {
	top := base.Add(V(0, h, 0))
	g.Use(Wood).FillVolume(Frustum{base.Plan(), 1.0, 0.8, base.Y, top.Y})
	phase := Hash01(seed, 141) * math.Pi
	for k := 0; k < 9; k++ {
		a := phase + float64(k)*2*math.Pi/9
		tip := Polar(top.Plan(), 2.8, a).At(top.Y + 1.2 + 0.4*Hash01(seed, k))
		g.Use(Grass).Fill(Capsule{top, tip, 0.48})
	}
	g.Use(Gold).Fill(Sphere(top.Add(V(0, 0.5, 0)), 0.65))
}

// fern is a low clump of arching fronds.
func (s DinosaurValley) fern(g *Grid, base Vec3, size float64, seed int) {
	phase := Hash01(seed, 151) * math.Pi
	for k := 0; k < 6; k++ {
		a := phase + float64(k)*2*math.Pi/6
		mid := Polar(base.Plan(), size*0.5, a).At(base.Y + size*0.7)
		tip := Polar(base.Plan(), size, a).At(base.Y + size*0.25)
		col := Foliage
		if k%2 == 0 {
			col = Grass
		}
		g.Use(col).Fill(Capsule{base, mid, 0.45}, Capsule{mid, tip, 0.42})
	}
}
