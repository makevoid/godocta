package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// Kinderdijk is the windmill landscape of the Dutch polder. Five tall thatched octagonal
// polder mills stand in a row along a straight boezem canal, their four lattice sails
// (some clothed in white canvas) set at different angles and facing the wind; each has
// a brick plinth, a thatched cap with a white beard board and a tail pole braced down to
// a red winch wheel among its posts. Behind a second canal and dyke stand two round
// brick mills, one a ground-sailer and one with a stage, and a brick steam pumping
// station with a tall chimney. Grass dykes with sandy paths line the canals; the flat
// meadows are cut into parcels by ditches, with reed beds, pollard and weeping willows,
// black-and-white cows, a farmhouse with a steep thatched roof, a hay barrack and boats.
// The site is turned so the row runs across the hero view; the sails face roughly +X.
type Kinderdijk struct{}

func init() { Register(Kinderdijk{}) }

func (Kinderdijk) Name() string  { return "kinderdijk" }
func (Kinderdijk) Title() string { return "Kinderdijk" }

const (
	kinderdijkG     = 1.2                    // Polder ground top.
	kinderdijkWater = 0.45                   // Canal water surface.
	kinderdijkX     = 46.0                   // Site half width along X.
	kinderdijkZ0    = -40.0                  // Site back edge.
	kinderdijkZ1    = 26.0                   // Site front edge.
	kinderdijkFace  = 40 * math.Pi / 180     // Plan direction the sails face.
	kinderdijkTurn  = -24 * math.Pi / 180    // Turn from the local plan into the world.
	kinderdijkScale = 1.15                   // Scale from the local plan into the world.
	kinderdijkTilt  = 10 * math.Pi / 180     // Windshaft tilt.
	kinderdijkDyke  = kinderdijkG + 2.0      // Dyke crest.
	kinderdijkRowA  = -3.0                   // Thatched mill row.
	kinderdijkRowB  = -31.0                  // Brick mill row.
	kinderdijkSq2   = 0.70710678118654752440 // 1/√2.
)

// Canal strips along X (z ranges) and dyke center lines.
var (
	kinderdijkCanals = [][2]float64{{4, 11}, {-18.5, -12.5}}
	kinderdijkDykes  = []float64{14, -21.5}
)

// kinderdijkFrame is a plan frame: a along n, b along u (n turned +90°).
type kinderdijkFrame struct{ o, n, u Vec2 }

func kinderdijkFacing(o Vec2, angle float64) kinderdijkFrame {
	n := P(math.Cos(angle), math.Sin(angle))
	return kinderdijkFrame{o, n, P(-n.Y, n.X)}
}

func (f kinderdijkFrame) world(a, b float64) Vec2 { return f.o.Add(f.n.Scale(a)).Add(f.u.Scale(b)) }

func (f kinderdijkFrame) local(p Vec3) (float64, float64) {
	d := p.Plan().Sub(f.o)
	return d.Dot(f.n), d.Dot(f.u)
}

// solid is the predicate over local (a, b, y) within a local box.
func (f kinderdijkFrame) solid(a0, a1, b0, b1, y0, y1 float64, pred func(a, b, y float64) bool) Solid {
	box := EmptyBox()
	for _, a := range []float64{a0, a1} {
		for _, b := range []float64{b0, b1} {
			q := f.world(a, b)
			box = box.Union(Box{q.At(y0), q.At(y1)})
		}
	}
	return Func{box.Grow(0.05), func(p Vec3) bool {
		a, b := f.local(p)
		if a < a0 || a > a1 || b < b0 || b > b1 {
			return false
		}
		return pred == nil || pred(a, b, p.Y)
	}}
}

// kinderdijkOct is the octagonal "radius" (apothem) with flats facing a and b.
func kinderdijkOct(a, b float64) float64 {
	a, b = math.Abs(a), math.Abs(b)
	return math.Max(math.Max(a, b), (a+b)*kinderdijkSq2)
}

// kinderdijkInCanal reports whether a local z lies within margin of a canal.
func kinderdijkInCanal(z, margin float64) bool {
	for _, c := range kinderdijkCanals {
		if z >= c[0]-margin && z <= c[1]+margin {
			return true
		}
	}
	return false
}

// kinderdijkSite is the rounded-rectangle site outline, shrunk by inset.
func kinderdijkSite(x, z, inset float64) bool {
	const r = 9.0
	hx, z0, z1 := kinderdijkX-inset, kinderdijkZ0+inset, kinderdijkZ1-inset
	if x < -hx || x > hx || z < z0 || z > z1 {
		return false
	}
	cx := math.Max(math.Abs(x)-(hx-r), 0)
	cz := math.Max(math.Max(z0+r-z, z-(z1-r)), 0)
	return cx*cx+cz*cz <= r*r
}

func kinderdijkCellHash(p Vec3, seed int) float64 {
	return Hash01(int(math.Round(p.X*4)), int(math.Round(p.Y*4)), int(math.Round(p.Z*4)), seed)
}

// Build composes the site in a local plan with the canals along local X and the front
// meadow towards local +Z, then turns it so the mill row runs across the hero view.
func (k Kinderdijk) Build(g *Grid) Order {
	k.compose(kinderdijkGrid{g})
	sn, cs := math.Sincos(kinderdijkTurn)
	return Along(P(cs, sn))
}

// kinderdijkGrid turns and scales every solid from the local plan into the world.
type kinderdijkGrid struct{ *Grid }

// kinderdijkPlaced is a local-plan solid placed in the world.
type kinderdijkPlaced struct{ s Solid }

func kinderdijkToLocal(p Vec3) Vec3 {
	return p.Plan().Rotate(-kinderdijkTurn).Scale(1 / kinderdijkScale).At(p.Y / kinderdijkScale)
}

func kinderdijkToWorld(b Box) Box {
	out := EmptyBox()
	for _, x := range []float64{b.Min.X, b.Max.X} {
		for _, z := range []float64{b.Min.Z, b.Max.Z} {
			q := P(x, z).Scale(kinderdijkScale).Rotate(kinderdijkTurn)
			out = out.Union(Box{q.At(b.Min.Y * kinderdijkScale), q.At(b.Max.Y * kinderdijkScale)})
		}
	}
	return out
}

func (v kinderdijkPlaced) Contains(p Vec3) bool { return v.s.Contains(kinderdijkToLocal(p)) }
func (v kinderdijkPlaced) Bounds() Box          { return kinderdijkToWorld(v.s.Bounds()) }

func (g kinderdijkGrid) turn(solids []Solid) []Solid {
	out := make([]Solid, len(solids))
	for i, s := range solids {
		out[i] = kinderdijkPlaced{s}
	}
	return out
}

func (g kinderdijkGrid) Fill(solids ...Solid)       { g.Grid.Fill(g.turn(solids)...) }
func (g kinderdijkGrid) FillVolume(solids ...Solid) { g.Grid.FillVolume(g.turn(solids)...) }
func (g kinderdijkGrid) Carve(solids ...Solid)      { g.Grid.Carve(g.turn(solids)...) }
func (g kinderdijkGrid) Paint(c Color, solids ...Solid) {
	g.Grid.Paint(c, g.turn(solids)...)
}

func (k Kinderdijk) compose(g kinderdijkGrid) {
	k.terrain(g)
	spins := []float64{0.3, math.Pi / 4, 0.08, 0.55, 0}
	canvas := [][4]float64{{0, 0, 0, 0}, {1, 1, 1, 1}, {0, 0, 0, 0}, {1, 0, 1, 0}, {0, 0, 0, 0}}
	for i, x := range []float64{-36, -19, -2, 15, 32} {
		k.mill(g, kinderdijkMill{
			c: P(x, kinderdijkRowA), height: 14, r0: 5.3, r1: 2.9, sail: 14.3,
			spin: spins[i], canvas: canvas[i], seed: i + 1,
		})
	}
	k.mill(g, kinderdijkMill{
		c: P(-27, kinderdijkRowB), round: true, height: 13.5, r0: 5.2, r1: 2.9, sail: 13.8,
		spin: 0.4, seed: 11,
	})
	k.mill(g, kinderdijkMill{
		c: P(-8, kinderdijkRowB), round: true, height: 17, r0: 5.0, r1: 2.8, stage: 6, sail: 12.8,
		spin: 0.15, canvas: [4]float64{1, 1, 1, 1}, seed: 12,
	})
	k.mill(g, kinderdijkMill{
		c: P(11, kinderdijkRowB), height: 14, r0: 5.3, r1: 2.9, sail: 14.3,
		spin: 0.65, seed: 13,
	})
	k.pumpingStation(g, P(32, kinderdijkRowB+1))
	k.bridge(g, -10)
	k.farm(g)
	k.life(g)
}

// kinderdijkExact sets cells whose actual centers satisfy pred, without the half-unit
// snapping of Fill, so thin sail members stay one member thick.
func kinderdijkExact(g kinderdijkGrid, local Box, color func(Vec3) (Color, bool)) {
	b := kinderdijkToWorld(local)
	inv := 2 / g.Span()
	y0, y1 := int32(math.Ceil(b.Min.Y*inv)), int32(math.Floor(b.Max.Y*inv))
	z0, z1 := int32(math.Ceil(b.Min.Z*inv)), int32(math.Floor(b.Max.Z*inv))
	x0, x1 := int32(math.Ceil(b.Min.X*inv)), int32(math.Floor(b.Max.X*inv))
	for y := y0; y <= y1; y++ {
		p := y & 1
		for z := z0 + ((z0&1)^p)&1; z <= z1; z += 2 {
			for x := x0 + ((x0&1)^p)&1; x <= x1; x += 2 {
				c := Cell{X: x, Y: y, Z: z}
				if col, ok := color(kinderdijkToLocal(g.World(c))); ok {
					g.Remove(c)
					g.With(col, func() { g.Add(c) })
				}
			}
		}
	}
}

// terrain lays the polder plate, canals, dykes, ditches and meadow colors.
func (k Kinderdijk) terrain(g kinderdijkGrid) {
	const G = kinderdijkG
	site := Box{V(-kinderdijkX, 0, kinderdijkZ0), V(kinderdijkX, G, kinderdijkZ1)}
	g.With(Grass, func() {
		g.Fill(Func{site, func(p Vec3) bool {
			if !kinderdijkSite(p.X, p.Z, 0) {
				return false
			}
			return p.Y <= G-0.5 || kinderdijkSite(p.X, p.Z, 1.0)
		}})
	})
	// Meadow parcels: ditches every 9 units along Z, each parcel a slightly different mix.
	ditch := func(x float64) bool {
		m := math.Mod(x+kinderdijkX+4.5, 9)
		return m < 1.0
	}
	inMeadow := func(z float64) bool { return z > 17.5 || z < -25.5 }
	g.Paint(Foliage, Func{site.Grow(1), func(p Vec3) bool {
		parcel := int(math.Floor((p.X + kinderdijkX + 4.5) / 9))
		side := 0
		if p.Z < 0 {
			side = 1
		}
		mix := 0.08 + 0.25*Hash01(parcel, side, 7)
		if !inMeadow(p.Z) {
			mix = 0.1
		}
		return kinderdijkCellHash(p, 3) < mix
	}})
	// Canals and ditches.
	water := func(p Vec3) bool {
		for _, c := range kinderdijkCanals {
			if p.Z >= c[0] && p.Z <= c[1] {
				return true
			}
		}
		return inMeadow(p.Z) && ditch(p.X) && kinderdijkSite(p.X, p.Z, 2.5) &&
			!(p.X < -18 && p.Z > 17) // The farmyard has no ditch.
	}
	cut := Func{Box{V(-kinderdijkX, 0.6, kinderdijkZ0), V(kinderdijkX, G+0.5, kinderdijkZ1)}, water}
	g.Carve(cut)
	g.Paint(Water, Func{site, func(p Vec3) bool { return p.Y <= kinderdijkWater+0.1 && water(p) }})
	// Dykes: trapezoid embankments along both canals with a sandy path on the crest.
	for _, zc := range kinderdijkDykes {
		zc := zc
		dyke := Func{Box{V(-kinderdijkX, G-0.3, zc-3.2), V(kinderdijkX, kinderdijkDyke, zc+3.2)}, func(p Vec3) bool {
			top := kinderdijkDyke
			// Ramps down near the site ends.
			top = math.Min(top, G+(kinderdijkX-1.5-math.Abs(p.X))*0.6)
			return kinderdijkSite(p.X, p.Z, 0.8) && p.Y <= top && math.Abs(p.Z-zc) <= 1.3+(top-p.Y)
		}}
		g.With(Grass, func() { g.FillVolume(dyke) })
		g.Paint(Sand, Func{dyke.Box, func(p Vec3) bool {
			return math.Abs(p.Z-zc) <= 0.8 && p.Y >= kinderdijkDyke-0.45 && dyke.F(p)
		}})
	}
	// A gravel towpath along the mill bank.
	g.Paint(Sand, Block(V(-kinderdijkX+2, G-0.1, 1.6), V(kinderdijkX-2, G+0.2, 3.0)))
}

// kinderdijkMill describes one polder mill.
type kinderdijkMill struct {
	c      Vec2
	round  bool       // Round brick tower instead of an octagonal thatched smock.
	height float64    // Body top above the ground.
	r0, r1 float64    // Body apothem/radius at its foot and its top.
	stage  float64    // Stage height above the ground; 0 for a ground-sailer.
	sail   float64    // Sail radius.
	spin   float64    // Angle of the first arm.
	canvas [4]float64 // Fraction of each sail covered with cloth.
	seed   int
}

// kinderdijkSails classifies points of a four-armed sail cross in a tilted plane.
type kinderdijkSails struct {
	hub     Vec3
	N, U, W Vec3
	R, spin float64
	canvas  [4]float64
}

const (
	kinderdijkNone = iota
	kinderdijkStock
	kinderdijkLattice
	kinderdijkCloth
	kinderdijkTip
)

func (s kinderdijkSails) part(p Vec3) int {
	v := p.Sub(s.hub)
	if math.Abs(v.Dot(s.N)) > 0.5 {
		return kinderdijkNone
	}
	x, y := v.Dot(s.U), v.Dot(s.W)
	if x*x+y*y > (s.R+0.3)*(s.R+0.3) {
		return kinderdijkNone
	}
	const r0, width, bar, period = 2.4, 3.1, 0.95, 2.5
	for i := 0; i < 4; i++ {
		sn, cs := math.Sincos(s.spin + float64(i)*math.Pi/2)
		along, across := x*cs+y*sn, -x*sn+y*cs
		if along < -0.7 || along > s.R {
			continue
		}
		tip := along >= s.R-1.5
		if math.Abs(across) <= 0.6 {
			if tip {
				return kinderdijkTip
			}
			return kinderdijkStock
		}
		if along < r0 || across < 0 || across > width {
			continue
		}
		// Cloth is spread over the front of the lattice, leaving the outer rail bare.
		cloth := along <= r0+(s.R-0.85-r0)*s.canvas[i]+0.4 && s.canvas[i] > 0
		if cloth && across < width-0.9 {
			return kinderdijkCloth
		}
		if across >= width-0.9 || math.Mod(along-r0, period) < bar || along >= s.R-0.85 {
			if tip {
				return kinderdijkTip
			}
			return kinderdijkLattice
		}
	}
	return kinderdijkNone
}

func (k Kinderdijk) mill(g kinderdijkGrid, m kinderdijkMill) {
	y0 := kinderdijkG
	f := kinderdijkFacing(m.c, kinderdijkFace)
	yb := y0 + 1.8 // Thatch starts above the brick plinth.
	if m.round {
		yb = y0 - 0.4
	}
	yTop := y0 + m.height
	rad := func(y float64) float64 {
		t := math.Max(0, math.Min(1, (y-yb)/(yTop-yb)))
		return m.r0 + (m.r1-m.r0)*t
	}
	dist := func(a, b float64) float64 {
		if m.round {
			return math.Hypot(a, b)
		}
		return kinderdijkOct(a, b)
	}
	ext := m.r0 + 0.5
	whole := func(y0, y1 float64, pred func(a, b, y float64) bool) Solid {
		return f.solid(-ext, ext, -ext, ext, y0, y1, pred)
	}
	// Body shell.
	body := whole(yb, yTop, func(a, b, y float64) bool {
		d, r := dist(a, b), rad(y)
		return d <= r && d > r-1.25
	})
	if m.round {
		g.With(Brick, func() { g.FillVolume(body) })
		g.Paint(Terracotta, Func{body.Bounds(), func(p Vec3) bool { return body.Contains(p) && kinderdijkCellHash(p, m.seed) < 0.12 }})
		g.Paint(DarkStone, Func{body.Bounds(), func(p Vec3) bool {
			return body.Contains(p) && kinderdijkCellHash(p, m.seed+50) < 0.1
		}})
	} else {
		g.With(Brick, func() {
			g.FillVolume(whole(y0-0.4, yb, func(a, b, y float64) bool {
				d := kinderdijkOct(a, b)
				return d <= m.r0-0.2 && d > m.r0-1.5
			}))
		})
		g.With(Charcoal, func() { g.FillVolume(body) })
		// Weathered streaks in the thatch, running down the courses.
		g.Paint(DarkStone, Func{body.Bounds(), func(p Vec3) bool {
			a, b := f.local(p)
			streak := int(math.Floor((math.Atan2(b, a) + math.Pi) * 7))
			return body.Contains(p) && Hash01(streak, int(math.Floor(p.Y/2.5)), m.seed)+0.4*kinderdijkCellHash(p, m.seed) < 0.45
		}})
	}
	g.Paint(Charcoal, whole(yb, yTop, func(a, b, y float64) bool { return dist(a, b) <= rad(y)-0.6 }))
	// Curb ring under the cap.
	g.With(White, func() {
		g.FillVolume(whole(yTop-0.5, yTop+0.2, func(a, b, y float64) bool {
			d := dist(a, b)
			return d <= m.r1+0.4 && d > m.r1-0.9
		}))
	})
	// Windows and doors, framed in white over a dark interior.
	opening := func(phi, yc, w, h float64, door bool) {
		sn, cs := math.Sincos(phi)
		proj := func(a, b float64) (float64, float64) { return a*cs + b*sn, -a*sn + b*cs }
		r := func(y float64) float64 {
			if !m.round && y < yb {
				return m.r0 - 0.2
			}
			return rad(y)
		}
		g.Paint(White, whole(yc-h/2-0.6, yc+h/2+0.5, func(a, b, y float64) bool {
			d, l := proj(a, b)
			return math.Abs(l) <= w/2+0.5 && d >= r(y)-0.75
		}))
		hole := func(a, b, y float64) bool {
			d, l := proj(a, b)
			return math.Abs(l) <= w/2 && math.Abs(y-yc) <= h/2 && d >= r(y)-1.8
		}
		g.Carve(whole(yc-h/2, yc+h/2, hole))
		if door {
			g.With(Red, func() {
				g.Fill(whole(yc-h/2, yc+h/2, func(a, b, y float64) bool {
					d, _ := proj(a, b)
					return hole(a, b, y) && d <= r(y)-0.9 && d >= r(y)-1.4
				}))
			})
		}
	}
	for _, s := range []float64{1, -1} {
		opening(s*math.Pi/2, y0+1.15, 1.6, 2.3, true)
	}
	for i, phi := range []float64{math.Pi / 4, 3 * math.Pi / 4, -3 * math.Pi / 4, -math.Pi / 4, math.Pi} {
		opening(phi, y0+4.6+float64(i%2)*0.7, 0.9, 1.3, false)
	}
	for _, phi := range []float64{0, math.Pi / 2, -math.Pi / 2} {
		opening(phi, yTop-3.4, 0.9, 1.2, false)
	}
	// Stage on brackets, with a railing and an upper door.
	if m.stage > 0 {
		ys := y0 + m.stage
		const out = 5.9
		g.With(Wood, func() {
			g.Fill(whole(ys-0.5, ys, func(a, b, y float64) bool {
				d := math.Hypot(a, b)
				return d <= out && d > rad(y)-0.6
			}))
			g.Fill(whole(ys+0.75, ys+1.2, func(a, b, y float64) bool {
				d := math.Hypot(a, b)
				return d <= out && d > out-0.6
			}))
			for i := 0; i < 16; i++ {
				sn, cs := math.Sincos(float64(i) * math.Pi / 8)
				q := f.world((out-0.3)*cs, (out-0.3)*sn)
				g.Fill(Cylinder{Center: q, Radius: 0.35, Y0: ys, Y1: ys + 1.2})
			}
			for i := 0; i < 8; i++ {
				sn, cs := math.Sincos(float64(i)*math.Pi/4 + math.Pi/8)
				lo := f.world((rad(ys-3)-0.2)*cs, (rad(ys-3)-0.2)*sn).At(ys - 3)
				hi := f.world((out-0.6)*cs, (out-0.6)*sn).At(ys - 0.4)
				g.Fill(Capsule{lo, hi, 0.4})
			}
		})
		opening(math.Pi/2, ys+1.2, 1.4, 2.2, true)
		opening(-math.Pi/2, ys+1.2, 1.4, 2.2, true)
	}
	// Cap: a thatched hood, longer front to back, white curb and beard.
	cap := whole(yTop-0.2, yTop+3.3, func(a, b, y float64) bool {
		ra, rb, ry := m.r1+1.3, m.r1+0.6, 3.1
		e := func(s float64) float64 {
			x, z, h := (a+0.3)/(ra-s), b/(rb-s), (y-yTop)/(ry-s)
			return x*x + z*z + h*h
		}
		return e(0) <= 1 && e(1.1) > 1
	})
	g.With(DarkStone, func() { g.FillVolume(cap) })
	g.Paint(Charcoal, Func{cap.Bounds(), func(p Vec3) bool { return cap.Contains(p) && kinderdijkCellHash(p, m.seed+9) < 0.55 }})
	// Windshaft, hub and beard board.
	sθ, cθ := math.Sincos(kinderdijkTilt)
	hubA := m.r1 + 2.1
	yS := yTop + 1.3
	hub := f.world(hubA, 0).At(yS + hubA*sθ/cθ)
	g.With(Charcoal, func() {
		g.Fill(Capsule{f.world(0.5, 0).At(yS), hub, 0.65})
		g.FillVolume(Sphere(hub, 1.05))
	})
	g.With(White, func() {
		g.Fill(f.solid(hubA-1.5, hubA-0.8, -1.6, 1.6, yTop-0.9, yS-0.4, func(a, b, y float64) bool {
			return math.Abs(b) <= 0.6+(y-(yTop-0.9))*0.6
		}))
	})
	g.Paint(Red, f.solid(hubA-1.5, hubA-0.6, -0.5, 0.5, yTop-0.5, yS-0.7, nil))
	// Sails.
	s := kinderdijkSails{
		hub:    hub,
		N:      V(f.n.X*cθ, sθ, f.n.Y*cθ),
		U:      V(f.u.X, 0, f.u.Y),
		W:      V(-f.n.X*sθ, cθ, -f.n.Y*sθ),
		R:      m.sail,
		spin:   m.spin,
		canvas: m.canvas,
	}
	colors := [...]Color{kinderdijkStock: Wood, kinderdijkLattice: Wood, kinderdijkCloth: White, kinderdijkTip: White}
	kinderdijkExact(g, Box{hub, hub}.Grow(m.sail+0.8), func(p Vec3) (Color, bool) {
		part := s.part(p)
		return colors[part], part != kinderdijkNone
	})
	// Tail pole with braces, down to the winch wheel and its ring of posts.
	footA, footY := -(m.r0 + 6.4), y0+1.6
	if m.stage > 0 {
		footA, footY = -5.2, y0+m.stage+1.5
	}
	top := f.world(-(m.r1 + 1.0), 0).At(yTop + 1.0)
	foot := f.world(footA, 0).At(footY)
	g.With(Wood, func() {
		g.Fill(Capsule{top, foot, 0.5})
		mid := top.Add(foot.Sub(top).Scale(0.6))
		for _, sb := range []float64{1, -1} {
			g.Fill(Capsule{f.world(-0.4, sb*(m.r1+0.4)).At(yTop + 0.2), mid, 0.42})
		}
	})
	g.Paint(White, Func{BoxOf(top, foot).Grow(0.6), func(p Vec3) bool {
		return p.Y <= footY+2.6 && Capsule{top, foot, 0.6}.Contains(p)
	}})
	wa := footA + 0.9
	wy := footY + 0.2
	if m.stage == 0 {
		wy = y0 + 1.75
	}
	g.With(Red, func() {
		g.Fill(f.solid(wa-0.45, wa+0.45, -1.9, 1.9, wy-1.9, wy+1.9, func(a, b, y float64) bool {
			r := math.Hypot(b, y-wy)
			return r <= 1.9 && r > 1.2
		}))
	})
	g.With(Wood, func() {
		g.Fill(f.solid(wa-0.4, wa+0.4, -1.3, 1.3, wy-1.3, wy+1.3, func(a, b, y float64) bool {
			return math.Abs(b) <= 0.4 || math.Abs(y-wy) <= 0.4
		}))
	})
	if m.stage == 0 {
		g.With(Wood, func() {
			for _, da := range []float64{-1.5, -1.0, -0.5, 0.5, 1.0, 1.5} {
				sn, cs := math.Sincos(math.Pi + da)
				q := f.world(-footA*cs, -footA*sn)
				if kinderdijkSite(q.X, q.Y, 1.5) && !kinderdijkInCanal(q.Y, 1) {
					g.Fill(Cylinder{Center: q, Radius: 0.45, Y0: y0 - 0.6, Y1: y0 + 1.3})
				}
			}
		})
	}
}

// pumpingStation is the brick steam pumping station with its tall chimney.
func (k Kinderdijk) pumpingStation(g kinderdijkGrid, c Vec2) {
	const G = kinderdijkG
	const hx, hz, wall = 7.0, 4.0, 5.5
	min, max := P(c.X-hx, c.Y-hz), P(c.X+hx, c.Y+hz)
	g.With(Brick, func() {
		g.Fill(Subtract(Block(min.At(G-0.4), max.At(G+wall)), Block(V(min.X+1.1, G, min.Y+1.1), V(max.X-1.1, G+wall+0.5, max.Y-1.1))))
		// Gable ends.
		g.Fill(Subtract(Gable{min, max, G + wall, AxisX}, Gable{P(min.X, min.Y+1.1), P(max.X, max.Y-1.1), G + wall, AxisX}))
	})
	g.Paint(White, Block(V(min.X-0.1, G+wall-0.5, min.Y-0.1), V(max.X+0.1, G+wall, max.Y+0.1)))
	g.With(Slate, func() {
		g.Fill(Subtract(Gable{P(min.X-0.5, min.Y-0.6), P(max.X+0.5, max.Y+0.6), G + wall - 0.6, AxisX},
			Gable{P(min.X-0.5, min.Y+0.6), P(max.X+0.5, max.Y-0.6), G + wall - 0.6, AxisX}))
	})
	g.Paint(Charcoal, Block(V(min.X+0.5, G, min.Y+0.5), V(max.X-0.5, G+wall, max.Y-0.5)))
	// Tall arched windows on the long sides, a door on the east end.
	for _, z := range []float64{max.Y, min.Y} {
		for i := -2; i <= 2; i++ {
			x := c.X + float64(i)*2.6
			g.Paint(White, Arch{V(x, G+1.0, z), 1.2, 1.9, 1.2, AxisZ})
			g.Carve(Arch{V(x, G+1.3, z), 0.7, 1.8, 3, AxisZ})
		}
	}
	g.Paint(White, Arch{V(max.X, G-0.1, c.Y), 1.7, 2.2, 1.2, AxisX})
	g.Carve(Arch{V(max.X, G, c.Y), 1.1, 2.0, 3, AxisX})
	g.With(Navy, func() { g.Fill(Block(V(max.X-1.2, G, c.Y-1.1), V(max.X-0.8, G+2.9, c.Y+1.1))) })
	// Chimney.
	ch := P(min.X-1.6, c.Y+1.5)
	g.With(Brick, func() {
		g.FillVolume(Subtract(Frustum{ch, 1.6, 1.15, G - 0.4, G + 17}, Cylinder{Center: ch, Radius: 0.6, Y0: G + 1, Y1: G + 18}))
		g.Fill(Block(V(ch.X-2.2, G-0.4, ch.Y-2.2), V(ch.X+2.2, G+2.2, ch.Y+2.2)))
	})
	g.Paint(Charcoal, Cylinder{Center: ch, Radius: 2, Y0: G + 15.6, Y1: G + 17})
	g.Paint(White, Block(V(ch.X-2.3, G+1.8, ch.Y-2.3), V(ch.X+2.3, G+2.2, ch.Y+2.3)))
	g.With(Brick, func() {
		g.FillVolume(Subtract(Cylinder{Center: ch, Radius: 1.75, Y0: G + 14.8, Y1: G + 15.6}, Cylinder{Center: ch, Radius: 0.6, Y0: G, Y1: G + 18}))
	})
}

// farm builds the farmhouse with its steep thatched hip roof and the hay barrack.
func (k Kinderdijk) farm(g kinderdijkGrid) {
	const G = kinderdijkG
	c := P(-32, 21.6)
	const hx, hz, wall, slope = 6.5, 3.6, 3.2, 1.45
	min, max := P(c.X-hx, c.Y-hz), P(c.X+hx, c.Y+hz)
	g.With(Brick, func() {
		g.Fill(Subtract(Block(min.At(G-0.4), max.At(G+wall)), Block(V(min.X+1, G, min.Y+1), V(max.X-1, G+wall+0.5, max.Y-1))))
	})
	g.Paint(Charcoal, Block(V(min.X+0.5, G, min.Y+0.5), V(max.X-0.5, G+wall, max.Y-0.5)))
	// Windows and a red door on the long front, facing the canal.
	for _, x := range []float64{-4.6, -2.4, 2.4, 4.6} {
		g.Paint(White, Block(V(c.X+x-0.9, G+0.9, max.Y-0.6), V(c.X+x+0.9, G+2.8, max.Y+0.1)))
		g.Carve(Block(V(c.X+x-0.45, G+1.3, max.Y-1.5), V(c.X+x+0.45, G+2.4, max.Y+0.1)))
	}
	g.Paint(White, Block(V(c.X-1.0, G, max.Y-0.6), V(c.X+1.0, G+2.9, max.Y+0.1)))
	g.Carve(Block(V(c.X-0.55, G, max.Y-1.5), V(c.X+0.55, G+2.5, max.Y+0.1)))
	g.With(Red, func() { g.Fill(Block(V(c.X-0.55, G, max.Y-1.0), V(c.X+0.55, G+2.5, max.Y-0.6))) })
	for _, x := range []float64{min.X, max.X} {
		g.Paint(White, Block(V(x-0.1, G+0.9, c.Y-0.9), V(x+0.1, G+2.8, c.Y+0.9)))
		g.Carve(Block(V(x-0.8, G+1.3, c.Y-0.45), V(x+0.8, G+2.4, c.Y+0.45)))
	}
	// Steep thatched hip roof with overhanging eaves.
	base := G + wall - 0.4
	roof := func(inset float64) func(p Vec3) bool {
		return func(p Vec3) bool {
			d := (p.Y - base) / slope
			return math.Abs(p.X-c.X) <= hx+0.8-d-inset && math.Abs(p.Z-c.Y) <= hz+0.8-d-inset
		}
	}
	rb := Box{V(min.X-0.8, base, min.Y-0.8), V(max.X+0.8, base+(hz+0.8)*slope, max.Y+0.8)}
	outer, inner := roof(0), roof(0.9)
	g.With(Wood, func() {
		g.FillVolume(Func{rb, func(p Vec3) bool { return outer(p) && (!inner(p) || p.Y < base+0.5) }})
	})
	g.Paint(Charcoal, Func{rb, func(p Vec3) bool { return outer(p) && p.Y >= base+(hz+0.8)*slope-1.0 }})
	g.With(Brick, func() { g.Fill(Block(V(c.X+2.5, base+2, c.Y-0.5), V(c.X+3.5, base+6.2, c.Y+0.5))) })
	// Hay barrack: four poles, a hay stack and a pyramidal roof.
	h := P(-20.5, 22)
	g.With(Wood, func() {
		for _, dx := range []float64{-1.8, 1.8} {
			for _, dz := range []float64{-1.8, 1.8} {
				g.Fill(Block(V(h.X+dx-0.4, G, h.Y+dz-0.4), V(h.X+dx+0.4, G+7, h.Y+dz+0.4)))
			}
		}
	})
	g.With(Sand, func() { g.FillVolume(Cylinder{Center: h, Radius: 2.0, Y0: G, Y1: G + 3.8}) })
	g.With(Charcoal, func() { g.Fill(HipRoof{P(h.X-2.7, h.Y-2.7), P(h.X+2.7, h.Y+2.7), G + 5}) })
}

// life adds willows, reeds, cows and boats.
func (k Kinderdijk) life(g kinderdijkGrid) {
	const G = kinderdijkG
	// Pollard willows along the dykes' foot and the ditches.
	pollard := func(c Vec2, seed int) {
		h := 2.0 + Hash01(seed, 1)*0.8
		g.With(Wood, func() { g.Fill(Cylinder{Center: c, Radius: 0.55, Y0: G - 0.3, Y1: G + h}) })
		g.With(Foliage, func() {
			g.FillVolume(Ellipsoid{c.At(G + h + 1.0), V(1.7, 1.4, 1.7)})
		})
		g.Paint(Grass, Func{Box{c.At(G + h), c.At(G + h + 2.6)}.Grow(1.9), func(p Vec3) bool { return kinderdijkCellHash(p, seed) < 0.3 }})
	}
	weeping := func(c Vec2, seed int) {
		g.With(Wood, func() { g.Fill(Cylinder{Center: c, Radius: 0.75, Y0: G - 0.3, Y1: G + 4}) })
		crown := Subtract(Ellipsoid{c.At(G + 5), V(3.4, 3.4, 3.4)}, Cylinder{Center: c, Radius: 2.0, Y0: G, Y1: G + 3.4})
		g.With(Grass, func() { g.FillVolume(crown) })
		g.Paint(Foliage, Func{crown.Bounds(), func(p Vec3) bool { return crown.Contains(p) && kinderdijkCellHash(p, seed) < 0.35 }})
	}
	i := 0
	for x := -40.0; x <= 40; x += 6.5 {
		i++
		if Hash01(i, 5) < 0.55 && math.Abs(x+32) > 8 {
			pollard(P(x+Hash01(i, 6)*2, 17.6+Hash01(i, 7)*0.6), 100+i)
		}
		if Hash01(i, 8) < 0.5 && math.Abs(x-19) > 10 {
			pollard(P(x+Hash01(i, 9)*2, -25.8-Hash01(i, 10)*0.6), 200+i)
		}
	}
	weeping(P(-41, 20), 301)
	weeping(P(-23, 17.8), 302)
	weeping(P(36, -35), 303)
	weeping(P(-38, -35), 304)
	// Reeds along the canal banks and in beds at the canal ends.
	reed := func(c Vec2, seed int) {
		h := 0.9 + Hash01(seed, 2)*1.1
		g.With(Foliage, func() { g.Fill(Cylinder{Center: c, Radius: 0.55, Y0: kinderdijkWater - 0.1, Y1: G + h}) })
		g.Paint(Sand, Cylinder{Center: c, Radius: 0.6, Y0: G + h - 0.4, Y1: G + h})
	}
	n := 0
	for ci, cz := range kinderdijkCanals {
		for _, z := range []float64{cz[0] + 0.5, cz[1] - 0.5} {
			for x := -kinderdijkX + 3; x <= kinderdijkX-3; x += 1.3 {
				n++
				band := math.Sin(x*0.21+float64(ci)*2) > 0.25
				if band && Hash01(n, 1) < 0.8 {
					reed(P(x+Hash01(n, 3)*0.6, z+(Hash01(n, 4)-0.5)*0.8), n)
				}
			}
		}
	}
	for _, bed := range []Vec2{P(-41, 7), P(41, 6.5), P(41, -15.5), P(-41, -15)} {
		for j := 0; j < 22; j++ {
			n++
			q := bed.Add(P((Hash01(n, 5)-0.5)*5, (Hash01(n, 6)-0.5)*5))
			if kinderdijkSite(q.X, q.Y, 2) {
				reed(q, n)
			}
		}
	}
	// Cows grazing in the meadows.
	cow := func(c Vec2, angle float64, seed int) {
		f := kinderdijkFacing(c, angle)
		body := f.solid(-1.4, 1.4, -0.65, 0.65, G+0.8, G+2.0, nil)
		legs := f.solid(-1.3, 1.3, -0.6, 0.6, G, G+0.8, func(a, b, y float64) bool { return math.Abs(a) >= 0.55 })
		head := f.solid(1.4, 2.3, -0.45, 0.45, G+1.4, G+2.3, nil)
		g.With(White, func() { g.Fill(body, legs, head) })
		g.Paint(Charcoal, Func{body.Bounds().Grow(1.2), func(p Vec3) bool {
			return Hash01(int(math.Round(p.X*1.5)), int(math.Round(p.Y*1.5)), int(math.Round(p.Z*1.5)), seed) < 0.45
		}})
		g.Paint(Charcoal, head)
	}
	for j, c := range []Vec2{P(-9, 21), P(-4, 23.5), P(3, 20.5), P(12, 22.8), P(20, 20.2), P(27, 23), P(-12, -30), P(-30, -36.5), P(4, -36), P(9, -33)} {
		cow(c, Hash01(j, 11)*2*math.Pi, 400+j)
	}
	// Boats: a green-black punt on the front canal, a white rowing boat on the back.
	boat := func(c Vec2, angle, length, beam float64, hull, rim Color) {
		f := kinderdijkFacing(c, angle)
		half := func(a float64) float64 {
			t := a / (length / 2)
			return beam / 2 * math.Sqrt(math.Max(1-t*t*t*t, 0))
		}
		shape := f.solid(-length/2, length/2, -beam/2, beam/2, kinderdijkWater-0.2, kinderdijkWater+1.1, func(a, b, y float64) bool {
			return math.Abs(b) <= half(a)
		})
		g.With(hull, func() { g.FillVolume(shape) })
		g.Paint(rim, f.solid(-length/2, length/2, -beam/2, beam/2, kinderdijkWater+0.8, kinderdijkWater+1.2, nil))
		g.Carve(f.solid(-length/2+1, length/2-1, -beam/2, beam/2, kinderdijkWater+0.6, kinderdijkWater+1.3, func(a, b, y float64) bool {
			return math.Abs(b) <= half(a)-0.7
		}))
		g.With(Wood, func() {
			g.Fill(f.solid(-0.4, 0.4, -beam/2, beam/2, kinderdijkWater+0.5, kinderdijkWater+0.9, func(a, b, y float64) bool {
				return math.Abs(b) <= half(a)-0.3
			}))
		})
	}
	boat(P(7, 7.6), 0.08, 6.5, 2.4, Navy, Wood)
	boat(P(-14, -15.4), -0.1, 5, 2.0, White, Red)
}

// bridge is a flat wooden footbridge on piers over the front canal, with white
// railings, from the mill bank onto the dyke.
func (k Kinderdijk) bridge(g kinderdijkGrid, x float64) {
	const G = kinderdijkG
	z0, z1 := kinderdijkCanals[0][0]-1.2, kinderdijkCanals[0][1]+1.6
	deck := G + 0.9
	g.With(Wood, func() {
		g.Fill(Block(V(x-1.4, deck-0.5, z0), V(x+1.4, deck, z1)))
		for _, z := range []float64{kinderdijkCanals[0][0] + 1.8, (kinderdijkCanals[0][0] + kinderdijkCanals[0][1]) / 2, kinderdijkCanals[0][1] - 1.8} {
			g.Fill(Block(V(x-1.4, 0, z-0.45), V(x+1.4, deck, z+0.45)))
		}
	})
	g.With(White, func() {
		for _, sx := range []float64{-1, 1} {
			g.Fill(Block(V(x+sx*1.4-0.45, deck+0.8, z0+0.6), V(x+sx*1.4+0.45, deck+1.25, z1-0.6)))
			for z := z0 + 0.6; z <= z1-0.5; z += 2.4 {
				g.Fill(Block(V(x+sx*1.4-0.45, deck, z-0.4), V(x+sx*1.4+0.45, deck+1.25, z+0.4)))
			}
		}
	})
}
