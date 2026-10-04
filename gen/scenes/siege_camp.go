package scenes

import (
	"math"

	"godocta/gen/kit"
	. "godocta/gen/octa"
)

// SiegeCamp is a medieval siege in progress. Along the back (-Z) a section of a
// besieged town wall faces the camp: a curtain with a gallery, two round towers, a big
// gatehouse whose drawbridge is raised over a water moat, a breach with a spill of
// rubble, scorch marks, defenders' red and navy banners, and the town's roofs, church
// spire and a burning house behind. In front (+Z) the besiegers have dug a trench and
// palisaded bank, set up mantlets, pushed a hide-covered wheeled siege tower toward the
// wall and rolled a covered battering ram over a fascine causeway to the gate. Behind
// the lines stand three big timber trebuchets (two cocked, one mid-throw with its stone
// in flight), wagons and stone piles, then the camp: rows of striped pavilions, a
// command pavilion with banners, campfires and horses at a picket line. Trampled
// fields with gentle hills slope down to a low rim.
type SiegeCamp struct{}

func init() { Register(SiegeCamp{}) }

func (SiegeCamp) Name() string  { return "siege_camp" }
func (SiegeCamp) Title() string { return "Siege Camp" }

const (
	siegeCampHX      = 56.0  // Site half width along X.
	siegeCampZ0      = -47.0 // Back (town) edge.
	siegeCampZ1      = 49.0  // Front (camp) edge.
	siegeCampTown    = 2.4   // Ground level along the wall.
	siegeCampWallTop = siegeCampTown + 13
	siegeCampFront   = -28.0 // Curtain front face.
	siegeCampBack    = -32.0 // Curtain back face.
	siegeCampGateX   = -8.0
	siegeCampBreachX = 20.0
	siegeCampMoatZ   = -20.0
	siegeCampWater   = 1.0
	siegeCampTrenchZ = -6.0
	siegeCampBelfryX = -26.0
	siegeCampBelfryZ = -9.5
)

// Plan positions of the larger camp features.
var (
	siegeCampCommand = P(12, 39)
	siegeCampPicket  = P(46.5, 36)
)

func (SiegeCamp) smooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

// noise is smooth 2D value noise in [0, 1).
func (s SiegeCamp) noise(x, z, scale float64, seed int) float64 {
	fx, fz := x/scale, z/scale
	ix, iz := math.Floor(fx), math.Floor(fz)
	tx, tz := s.smooth(fx-ix), s.smooth(fz-iz)
	i, k := int(ix), int(iz)
	a, b := Hash01(i, k, seed), Hash01(i+1, k, seed)
	c, d := Hash01(i, k+1, seed), Hash01(i+1, k+1, seed)
	return (a+(b-a)*tx)*(1-tz) + (c+(d-c)*tx)*tz
}

// lane reports whether x lies in an approach lane kept clear through the siege lines.
func (SiegeCamp) lane(x float64) bool {
	return math.Abs(x-siegeCampGateX) < 5 || math.Abs(x-siegeCampBelfryX) < 6.5
}

// height is the terrain surface: rolling camp fields, a glacis rising to the town, the
// moat (bridged by a fascine causeway at the gate), the besiegers' trench and bank,
// a knoll under the command pavilion, and a falloff to a low rim at the site edge.
func (s SiegeCamp) height(x, z float64) float64 {
	h := 1.2 + 0.5*math.Sin(x*0.10+0.7)*math.Cos(z*0.12-0.4) + 0.3*math.Sin(x*0.045-z*0.06+1.3)
	k := P(x, z).Sub(siegeCampCommand).Length()
	h += 1.0 * s.smooth(1-k/12)
	h += (siegeCampTown - h) * s.smooth((-z-9)/8)
	// Moat, faded out at its ends and beside the causeway.
	m := math.Abs(z - siegeCampMoatZ)
	moat := 0.2 + (h-0.2)*s.smooth((m-3)/2.5)
	w := s.smooth((49-math.Abs(x))/3) * s.smooth((math.Abs(x-siegeCampGateX)-4.5)/1.5)
	h += (moat - h) * w
	// Trench with a bank behind it, broken by the approach lanes.
	if !s.lane(x) {
		f := s.smooth((50 - math.Abs(x)) / 4)
		dt := (z - siegeCampTrenchZ) / 1.5
		db := (z - siegeCampTrenchZ - 2.8) / 1.9
		if math.Abs(dt) < 1 {
			h -= 1.1 * f * (1 - dt*dt) * (1 - dt*dt)
		}
		if math.Abs(db) < 1 {
			h += 1.2 * f * (1 - db*db) * (1 - db*db)
		}
	}
	e := math.Min(math.Min(x+siegeCampHX, siegeCampHX-x), math.Min(z-siegeCampZ0, siegeCampZ1-z))
	return 0.2 + (h-0.2)*s.smooth(e/8)
}

func (s SiegeCamp) slope(x, z float64) float64 {
	const d = 0.6
	h := s.height(x, z)
	m := 0.0
	for _, o := range []Vec2{P(d, 0), P(-d, 0), P(0, d), P(0, -d)} {
		m = math.Max(m, math.Abs(s.height(x+o.X, z+o.Y)-h)/d)
	}
	return m
}

func (s SiegeCamp) Build(g *Grid) Order {
	s.ground(g)
	s.town(g)
	s.curtain(g)
	s.gatehouse(g)
	s.breach(g)
	s.scorch(g)
	s.banners(g)
	s.works(g)
	s.ram(g)
	s.belfry(g)
	s.trebuchet(g, P(-38, 10), -0.63, 1.05)
	s.trebuchet(g, P(2, 9), 2.15, 1.12)
	s.trebuchet(g, P(36, 10), -0.63, 1.12)
	s.supplies(g)
	s.camp(g)
	return Along(P(0.6, 0.8))
}

// ground is a terrain skin about one cell layer deep, thicker on slopes so banks stay closed.
func (s SiegeCamp) ground(g *Grid) {
	box := Box{V(-siegeCampHX, -1.5, siegeCampZ0), V(siegeCampHX, 4.5, siegeCampZ1)}
	g.Use(Grass).Fill(Func{box, func(p Vec3) bool {
		h := s.height(p.X, p.Z)
		return p.Y <= h && p.Y > h-0.45-0.9*s.slope(p.X, p.Z)
	}})
	g.Paint(Foliage, Func{box, func(p Vec3) bool {
		return Hash01(int(math.Floor(p.X*1.3)), int(math.Floor(p.Z*1.3)), 11) < 0.06
	}})
	// Trampled earth: heaviest in the siege works, along the camp lanes and around the town.
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		x, z := p.X, p.Z
		n := s.noise(x, z, 5, 3)*0.7 + s.noise(x, z, 1.6, 4)*0.3
		limit := 0.78
		switch {
		case z < -14:
			limit = 0.5
		case z < 4:
			limit = 0.38
		case math.Abs(z-27) < 2.6 || math.Abs(x-siegeCampGateX) < 2.4 && z < 47:
			limit = 0.25
		case z > 24:
			limit = 0.62
		}
		return n > limit
	}})
	// Earth of the trench and bank, mud along the moat.
	g.Paint(Terracotta, Func{box, func(p Vec3) bool {
		if s.lane(p.X) || math.Abs(p.X) > 49 {
			return false
		}
		dz := p.Z - siegeCampTrenchZ
		return dz > -1.6 && dz < 4.7 && Hash01(int(math.Floor(p.X*2)), int(math.Floor(p.Z*2)), 5) < 0.45
	}})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		if !s.lane(p.X) && math.Abs(p.X) < 48 && math.Abs(p.Z-siegeCampTrenchZ) < 0.7 {
			return true
		}
		m := math.Abs(p.Z - siegeCampMoatZ)
		return m > 2.5 && m < 4.2 && math.Abs(p.X) < 49 && s.noise(p.X, p.Z, 1.5, 6) > 0.45
	}})
	// Fascine causeway across the moat: bundles laid along X.
	g.Paint(Wood, Func{Box{V(siegeCampGateX-4.6, 1, -26), V(siegeCampGateX+4.6, 3, -14)}, func(p Vec3) bool {
		return math.Mod(p.Z+100, 1.6) < 0.8
	}})
	// Moat water.
	g.Use(Water).Fill(Func{Box{V(-49, 0, siegeCampMoatZ-6), V(49, siegeCampWater, siegeCampMoatZ+6)}, func(p Vec3) bool {
		return p.Y > s.height(p.X, p.Z)
	}})
}

// house is a hollow gabled town house with a dark interior behind its windows.
func (s SiegeCamp) house(g *Grid, c Vec2, w, d, h float64, wall, roof Color, ridge Axis) {
	y0 := s.height(c.X, c.Y) - 0.6
	lo, hi := V(c.X-w/2, y0, c.Y-d/2), V(c.X+w/2, y0+h, c.Y+d/2)
	g.Use(wall).Fill(Block(lo, hi))
	g.Use(Charcoal).Fill(Block(lo.Add(V(0.9, 0.6, 0.9)), hi.Sub(V(0.9, 0.4, 0.9))))
	g.Carve(Block(lo.Add(V(1.3, 1, 1.3)), hi.Sub(V(1.3, 0.8, 1.3))))
	g.Use(roof).Fill(Gable{lo.Plan().Sub(P(0.6, 0.6)), hi.Plan().Add(P(0.6, 0.6)), hi.Y, ridge})
	for x := lo.X + 1.8; x < hi.X-1; x += 2.2 {
		g.Carve(Block(V(x-0.5, y0+h-2.6, lo.Z-0.5), V(x+0.5, y0+h-1.2, lo.Z+1.0)))
		g.Carve(Block(V(x-0.5, y0+h-2.6, hi.Z-1.0), V(x+0.5, y0+h-1.2, hi.Z+0.5)))
	}
	g.Paint(Wood, Block(V(lo.X, y0+h-0.4, lo.Z), V(hi.X, y0+h, hi.Z)))
}

// town is a strip of roofs, a church and a burning house behind the wall.
func (s SiegeCamp) town(g *Grid) {
	walls := []Color{White, Sand, White, Stone, Sand}
	roofs := []Color{Terracotta, Brick, Terracotta, Slate, Brick}
	i := 0
	for x := -40.0; x <= 36; x += 7.2 {
		if math.Abs(x-28) < 5 || math.Abs(x-siegeCampBreachX+2) < 4.5 {
			continue
		}
		h := 5.5 + 2*Hash01(i, 21)
		s.house(g, P(x, -37.6), 6, 5, h, walls[i%5], roofs[(i+2)%5], AxisX)
		if Hash01(i, 23) < 0.35 && x < 20 {
			s.house(g, P(x+2.5, -43), 5, 4.5, h-1.5, walls[(i+3)%5], roofs[i%5], AxisZ)
		}
		i++
	}
	// Church with a square tower and a slate spire.
	cx, cz := 28.0, -40.5
	ty := s.height(cx, cz)
	g.Use(Stone).Fill(BlockAt(P(cx, cz), 6, 6, ty-1, ty+19))
	g.Use(Charcoal).Fill(BlockAt(P(cx, cz), 4, 4, ty, ty+18))
	g.Carve(BlockAt(P(cx, cz), 3.4, 3.4, ty+0.5, ty+18.5))
	for _, y := range []float64{ty + 9, ty + 14.5} {
		g.Carve(Arch{V(cx, y, cz+3), 0.9, 1.8, 3, AxisZ}, Arch{V(cx+3, y, cz), 0.9, 1.8, 3, AxisX}, Arch{V(cx-3, y, cz), 0.9, 1.8, 3, AxisX})
	}
	g.Use(Slate).FillVolume(Func{Box{V(cx-3.6, ty+19, cz-3.6), V(cx+3.6, ty+28, cz+3.6)}, func(p Vec3) bool {
		r := 3.6 * (1 - (p.Y-ty-19)/9)
		return math.Abs(p.X-cx) <= r && math.Abs(p.Z-cz) <= r
	}})
	g.Use(Gold).Fill(Cylinder{Center: P(cx, cz), Radius: 0.5, Y0: ty + 27, Y1: ty + 29.5})
	g.Fill(BlockAt(P(cx, cz), 2.4, 0.8, ty+28, ty+28.8))
	// Burning house behind the breach: a broken roof, flames and a leaning smoke column.
	c := P(siegeCampBreachX-2, -37.6)
	s.house(g, c, 6.5, 5, 6.5, Sand, Terracotta, AxisX)
	y0 := s.height(c.X, c.Y) - 0.6
	g.Carve(Ellipsoid{V(c.X+1, y0+9, c.Y), V(2.8, 2.5, 3.2)})
	g.Paint(Charcoal, Ellipsoid{V(c.X+1, y0+8.5, c.Y), V(4, 3, 4)})
	for k, o := range []Vec2{P(0.5, 0.4), P(2.2, -0.8), P(-0.8, -0.4)} {
		at := c.Add(o)
		hgt := 3.5 + 1.2*Hash01(k, 31)
		g.Use(Red).FillVolume(Cone(at, 1.6, y0+6, hgt))
		g.Use(Gold).FillVolume(Cone(at, 0.9, y0+6, hgt+1.3))
	}
	for k := 0; k < 6; k++ {
		t := float64(k)
		r := 1.5 + 0.3*t
		at := V(c.X+1+0.9*t+(Hash01(k, 33)-0.5)*1.2, y0+10.5+t*3.1, c.Y+0.3*t+(Hash01(k, 34)-0.5)*1.2)
		col := Stone
		if k < 2 {
			col = DarkStone
		} else if k >= 4 {
			col = White
		}
		g.Use(col).FillVolume(Ellipsoid{at, V(r, r*0.8, r)})
	}
}

// curtain is the hollow crenellated wall between the two round towers.
func (s SiegeCamp) curtain(g *Grid) {
	t, top := siegeCampTown, siegeCampWallTop
	x0, x1 := -44.0, 40.0
	zb, zf := siegeCampBack, siegeCampFront
	g.Use(Stone).Fill(Block(V(x0, -0.2, zb), V(x1, top, zf)))
	// Battered plinth along the front.
	g.Fill(Func{Box{V(x0, -0.2, zf), V(x1, t+2.5, zf+1.6)}, func(p Vec3) bool { return p.Z-zf <= (t+2.5-p.Y)*0.55 }})
	g.Use(Charcoal).Fill(Block(V(x0+1, t+0.6, zb+0.6), V(x1-1, top-1.1, zf-0.6)))
	g.Carve(Block(V(x0+1, t+1, zb+1), V(x1-1, top-1.5, zf-1)))
	g.Use(Stone).Build(kit.Battlements{Min: P(x0, zb), Max: P(x1, zf), Top: top})
	for x := x0 + 5; x < x1-3; x += 5 {
		if math.Abs(x-siegeCampGateX) < 12 || math.Abs(x-siegeCampBreachX) < 7 {
			continue
		}
		g.Carve(Block(V(x-0.5, t+5.5, zf-1.5), V(x+0.5, t+8.5, zf+0.5)))
	}
	for _, x := range []float64{-44, 40} {
		c := P(x, -30)
		g.Use(Stone).Build(kit.Tower{Center: c, Radius: 6.5, Base: -0.2, Top: t + 18, Merlons: 14, Windows: true})
		g.Paint(Charcoal, Cylinder{Center: c, Radius: 5.2, Y0: 0.5, Y1: t + 16})
		// Flag on a pole.
		s.flag(g, c.At(t+19.5), 6.5, Red, Navy)
	}
	// Masonry texture.
	g.Paint(DarkStone, Func{Box{V(-52, 0, -38), V(48, t+22, -22)}, func(p Vec3) bool {
		if c, ok := g.ColorOf(Nearest(p.Scale(1 / g.Span()))); !ok || c != Stone {
			return false
		}
		return Hash01(int(math.Floor(p.X*1.25)), int(math.Floor(p.Y*1.25)), int(math.Floor(p.Z*1.25)), 41) < 0.14
	}})
}

// flag is a pole with a two-color flag streaming toward +X.
func (s SiegeCamp) flag(g *Grid, base Vec3, h float64, a, b Color) {
	g.Use(Wood).Fill(Cylinder{Center: base.Plan(), Radius: 0.45, Y0: base.Y - 1, Y1: base.Y + h})
	g.Use(Gold).Fill(Sphere(base.Add(V(0, h+0.3, 0)), 0.6))
	top := base.Y + h - 0.3
	g.Use(a).Fill(Func{Box{V(base.X, top-2.6, base.Z-0.45), V(base.X+4.4, top, base.Z+0.45)}, func(p Vec3) bool {
		u := p.X - base.X
		wave := 0.25 * math.Sin(u*1.6)
		return p.Y-wave >= top-2.6 && p.Y-wave <= top
	}})
	g.Paint(b, Func{Box{V(base.X+0.3, top-1.75, base.Z-0.5), V(base.X+4.6, top-0.85, base.Z+0.5)}, func(p Vec3) bool { return true }})
}

// gatehouse is a tall hollow block with two D-towers projecting toward the camp, a
// passage with a portcullis, and the drawbridge drawn up against the opening.
func (s SiegeCamp) gatehouse(g *Grid) {
	t, gx := siegeCampTown, siegeCampGateX
	top := t + 17
	x0, x1, z0, z1 := gx-10, gx+10, -35.0, -27.0
	g.Use(Stone).Fill(Block(V(x0, -0.2, z0), V(x1, top, z1)))
	g.Use(Charcoal).Fill(Block(V(x0+1, t+0.6, z0+0.6), V(x1-1, top-1.1, z1-0.6)))
	g.Carve(Block(V(x0+1.2, t+1, z0+1.2), V(x1-1.2, top-1.5, z1-1.2)))
	g.Use(Stone).Build(kit.Battlements{Min: P(x0, z0), Max: P(x1, z1), Top: top})
	for _, side := range []float64{-1, 1} {
		c := P(gx+side*7, z1)
		g.Use(Stone).Fill(Cylinder{Center: c, Radius: 4.3, Y0: -0.2, Y1: top + 2})
		g.Fill(Frustum{c, 4.3, 5.1, top + 0.2, top + 1.2}, Cylinder{Center: c, Radius: 5.1, Y0: top + 1.2, Y1: top + 2.5})
		g.Use(Charcoal).Fill(Cylinder{Center: c, Radius: 3.4, Y0: t + 1, Y1: top})
		g.Carve(Cylinder{Center: c, Radius: 3.0, Y0: t + 1.4, Y1: top - 0.5})
		g.Use(Stone).Build(kit.RingBattlements{Center: c, Radius: 5.1, Top: top + 2.5, Count: 12})
		for k, y := range []float64{t + 5, t + 10.5} {
			for _, a := range []float64{0.5, 1.25} {
				ang := math.Pi/2 + side*(a+0.3*float64(k)) - side*0.2
				dir := P(math.Cos(ang), math.Sin(ang))
				at := c.Add(dir.Scale(4.0))
				g.Carve(Func{Box{V(at.X-1.2, y, at.Y-1.2), V(at.X+1.2, y+2.6, at.Y+1.2)}, func(p Vec3) bool {
					d := p.Plan().Sub(at)
					return math.Abs(d.Dot(P(-dir.Y, dir.X))) < 0.5
				}})
			}
		}
		// Slate cone roofs with defenders' flags.
		g.Use(Slate).FillVolume(Cone(c, 4.4, top+2.5, 7.5))
		s.flag(g, c.At(top+9.5), 4, Red, Navy)
	}
	// Passage, dark inside, with a portcullis halfway.
	pass := Arch{Sill: V(gx, t, -30), HalfWidth: 2.8, Spring: 5.2, Depth: 16, Through: AxisZ}
	g.Paint(Charcoal, Arch{Sill: V(gx, t-0.6, -30), HalfWidth: 3.5, Spring: 5.4, Depth: 15, Through: AxisZ})
	g.Carve(pass)
	g.Use(Charcoal).Fill(Func{Box{V(gx-2.9, t, -31.4), V(gx+2.9, t+8.2, -30.6)}, func(p Vec3) bool {
		in := Arch{Sill: V(gx, t, -31), HalfWidth: 2.9, Spring: 5.2, Depth: 1, Through: AxisZ}
		return in.Contains(p) && (math.Mod(p.X-gx+30.6, 1.5) < 0.55 || math.Mod(p.Y+30, 1.8) < 0.6)
	}})
	// Machicolation band and window over the gate.
	g.Use(Stone).Fill(Block(V(gx-3.4, top-3, z1), V(gx+3.4, top, z1+1.3)))
	for x := gx - 2.5; x <= gx+2.6; x += 1.7 {
		g.Carve(Block(V(x-0.4, top-3.1, z1+0.3), V(x+0.4, top-2.2, z1+1.4)))
	}
	g.Carve(Arch{V(gx, t+10.5, z1), 1.0, 1.6, 2.5, AxisZ})
	// Raised drawbridge with iron bands and chains.
	g.Use(Wood).Fill(Block(V(gx-3.3, t-0.3, -25.6), V(gx+3.3, t+9.4, -24.8)))
	g.Paint(Charcoal, Block(V(gx-3.3, t+1.5, -25.6), V(gx+3.3, t+2.1, -24.8)), Block(V(gx-3.3, t+7, -25.6), V(gx+3.3, t+7.6, -24.8)))
	for _, side := range []float64{-1, 1} {
		g.Use(Charcoal).Fill(Capsule{V(gx+side*2.8, t+9.2, -25.2), V(gx+side*2.6, t+13, -27.1), 0.48})
	}
	s.flag(g, V(gx, top+0.5, z0+2), 6, Red, Navy)
}

// breach knocks a ragged gap in the curtain and spills rubble into the moat.
func (s SiegeCamp) breach(g *Grid) {
	t, bx := siegeCampTown, siegeCampBreachX
	g.Carve(Func{Box{V(bx-12, t+3, siegeCampBack-2), V(bx+12, siegeCampWallTop+3, siegeCampFront+2)}, func(p Vec3) bool {
		dy := p.Y - (t + 4.2)
		w := 1.8 + dy*0.62 + (Hash01(int(math.Floor(p.X)), int(math.Floor(p.Y*0.8)), 7)-0.5)*1.6
		return math.Abs(p.X-bx) < w
	}})
	pile := Box{V(bx-12, -0.5, -38), V(bx+12, t+6, -15)}
	g.Use(Stone).Fill(Func{pile, func(p Vec3) bool {
		dx, dz := (p.X-bx)/(10-0.2*math.Abs(p.Z+30)), (p.Z+29)/9
		if dz < 0 {
			dz *= 1.6
		}
		d := dx*dx + dz*dz
		if d >= 1 {
			return false
		}
		top := 0.4 + (t+3.8)*(1-d) + 0.9*s.noise(p.X, p.Z, 1.3, 9)
		return p.Y <= top && p.Y > top-1.6
	}})
	g.Paint(DarkStone, Func{pile, func(p Vec3) bool {
		dz := (p.Z + 29) / 9
		return Hash01(int(math.Floor(p.X*1.1)), int(math.Floor(p.Y*1.1)), int(math.Floor(p.Z*1.1)), 13) < 0.35 && math.Abs(p.X-bx) < 11 && dz > -1.2 && dz < 1.3 && p.Y < t+8
	}})
	// Scattered blocks.
	for k := 0; k < 14; k++ {
		x := bx + (Hash01(k, 51)-0.5)*22
		z := -23 + (Hash01(k, 52)-0.5)*10
		y := math.Max(s.height(x, z), siegeCampWater-0.2)
		col := Stone
		if k%3 == 0 {
			col = DarkStone
		}
		g.Use(col).Fill(BlockAt(P(x, z), 1.3, 1.1, y-0.3, y+0.9))
	}
}

// scorch blackens the masonry around the breach and other hits: a soot core fading
// through dark stone.
func (s SiegeCamp) scorch(g *Grid) {
	t := siegeCampTown
	blobs := []struct {
		c Vec3
		r float64
	}{
		{V(siegeCampBreachX, t+7, -28), 9},
		{V(siegeCampGateX+3.5, t+13, -27), 3.2},
		{V(36, t+12, -25), 3.5},
		{V(-41, t+10, -25), 3},
	}
	for i, b := range blobs {
		box := Box{b.c.Sub(V(b.r, b.r, b.r)), b.c.Add(V(b.r, b.r*1.6, b.r))}
		d := func(p Vec3) float64 {
			q := p.Sub(b.c)
			if q.Y > 0 {
				q.Y /= 1.6 // Soot streaks upward.
			}
			return q.Length()/b.r + 0.5*(s.noise(p.X+p.Z, p.Y, 1.2, 60+i)-0.5)
		}
		g.Paint(DarkStone, Func{box, func(p Vec3) bool { return d(p) < 0.85 }})
		g.Paint(Charcoal, Func{box, func(p Vec3) bool { return d(p) < 0.55 }})
	}
}

// banners hangs the defenders' long red and navy banners on the curtain.
func (s SiegeCamp) banners(g *Grid) {
	t, top := siegeCampTown, siegeCampWallTop
	for _, x := range []float64{-36, -23, 5, 31} {
		z := siegeCampFront
		g.Use(Red).Fill(Block(V(x-1.4, t+4.5, z+0.1), V(x+1.4, top-0.6, z+0.9)))
		g.Paint(Navy, Block(V(x-0.45, t+4.5, z), V(x+0.45, top-0.6, z+1)))
		g.Paint(Gold, Block(V(x-0.9, top-4.4, z), V(x+0.9, top-3.2, z+1)))
		g.Carve(Func{Box{V(x-1, t+4, z), V(x+1, t+6, z+1)}, func(p Vec3) bool {
			return p.Y-(t+4.5) < 1.4-math.Abs(p.X-x)*1.4
		}})
		g.Use(Wood).Fill(Block(V(x-1.9, top-0.8, z+0.1), V(x+1.9, top-0.1, z+0.9)))
	}
}

// works are the siege lines: a palisade on the bank behind the trench and mantlets ahead.
func (s SiegeCamp) works(g *Grid) {
	bank := siegeCampTrenchZ + 2.8
	for x := -48.5; x <= 48.5; x += 1.15 {
		if s.lane(x) {
			continue
		}
		y0 := s.height(x, bank)
		h := 3.0 + 0.9*Hash01(int(math.Round(x*4)), 71)
		lean := V(0, 0, -0.18)
		top := V(x, y0+h, bank).Add(lean.Scale(h))
		g.Use(Wood).Fill(Capsule{V(x, y0-0.8, bank), top, 0.52})
		g.FillVolume(Sheared{Cone(P(x, bank), 0.62, y0+h, 1.3), P(lean.X, lean.Z), y0})
	}
	// Mantlets: leaning plank shields with props, some painted.
	k := 0
	for x := -47.0; x <= 47; x += 6.5 {
		if s.lane(x) || math.Abs(x-siegeCampBreachX) < 2.5 {
			continue
		}
		z := -13 + 1.2*Hash01(k, 81)
		y0 := s.height(x, z)
		shield := Sheared{Block(V(x-1.8, y0-0.4, z-0.4), V(x+1.8, y0+3.4, z+0.4)), P(0, 0.35), y0}
		g.Use(Wood).Fill(shield)
		if k%2 == 0 {
			g.Paint(Navy, shield)
			g.Paint(Gold, Intersect(shield, Block(V(x-1.9, y0+1.3, z-1), V(x+1.9, y0+2.0, z+2))), Intersect(shield, Block(V(x-0.35, y0-1, z-1), V(x+0.35, y0+4, z+2))))
		}
		g.Use(Wood).Fill(Capsule{V(x, y0+2.6, z+1.2), V(x, y0, z+3), 0.45})
		k++
	}
}

// wheel is a disc standing in the plane normal to X (Axis X = axle along X).
func (SiegeCamp) wheel(center Vec3, radius, half float64, alongX bool) Solid {
	b := Box{center.Sub(V(radius, radius, radius)), center.Add(V(radius, radius, radius))}
	return Func{b, func(p Vec3) bool {
		d := p.Sub(center)
		if alongX {
			return math.Abs(d.X) <= half && d.Y*d.Y+d.Z*d.Z <= radius*radius
		}
		return math.Abs(d.Z) <= half && d.Y*d.Y+d.X*d.X <= radius*radius
	}}
}

// ram is a battering ram under a hide-covered tortoise roof on the fascine causeway.
func (s SiegeCamp) ram(g *Grid) {
	gx := siegeCampGateX
	z0, z1 := -23.6, -12.0
	y0 := s.height(gx, -18)
	// Chassis beams and wheels.
	for _, side := range []float64{-1, 1} {
		g.Use(Wood).Fill(Block(V(gx+side*2.6-0.5, y0+0.6, z0), V(gx+side*2.6+0.5, y0+1.4, z1)))
		for _, z := range []float64{z0 + 1.8, (z0 + z1) / 2, z1 - 1.8} {
			g.Use(Wood).Fill(s.wheel(V(gx+side*3.4, y0+0.9, z), 1.2, 0.4, true))
			g.Use(Charcoal).Fill(s.wheel(V(gx+side*3.6, y0+0.9, z), 0.45, 0.45, true))
		}
		for z := z0 + 0.5; z <= z1; z += 2.8 {
			g.Use(Wood).Fill(Cylinder{Center: P(gx+side*2.6, z), Radius: 0.5, Y0: y0 + 1.2, Y1: y0 + 4})
		}
	}
	// Gabled tortoise roof: hide patches on timber ribs, hollow beneath.
	roof := Func{Box{V(gx-3.6, y0+3.6, z0-0.4), V(gx+3.6, y0+7.4, z1+0.4)}, func(p Vec3) bool {
		return math.Abs(p.X-gx) <= 3.6-(p.Y-y0-3.6)
	}}
	g.Use(Sand).Fill(roof)
	g.Carve(Func{roof.Box, func(p Vec3) bool {
		return math.Abs(p.X-gx) <= 3.6-(p.Y-y0-3.6)-1.1 && p.Z > z0+0.4 && p.Z < z1-0.4
	}})
	g.Paint(Terracotta, Func{roof.Box, func(p Vec3) bool {
		return Hash01(int(math.Floor(p.Z/1.6)), int(math.Floor((p.Y-y0)/1.1)), int(math.Floor((p.X-gx)/8)), 91) < 0.5
	}})
	g.Paint(Wood, Func{roof.Box, func(p Vec3) bool { return math.Mod(p.Z+100, 2.8) < 0.6 }})
	// The ram: a heavy beam on rope slings with an iron head at the drawbridge.
	ry := y0 + 2.6
	g.Use(Wood).Fill(Capsule{V(gx, ry, z1-0.5), V(gx, ry, z0-0.2), 0.8})
	g.Use(Charcoal).Fill(Capsule{V(gx, ry, z0+0.4), V(gx, ry, -24.75), 1.05})
	for _, z := range []float64{z0 + 3, z1 - 3} {
		g.Use(Sand).Fill(Capsule{V(gx, ry+0.5, z), V(gx, y0+6, z), 0.48})
	}
}

// belfry is the wheeled siege tower: a tapering timber frame covered in hides, with a
// raised boarding bridge at the top facing the wall and a crenellated fighting deck.
func (s SiegeCamp) belfry(g *Grid) {
	cx, cz := siegeCampBelfryX, siegeCampBelfryZ
	y0 := s.height(cx, cz)
	yb, height := y0+1.8, 18.0
	half := func(y float64) float64 { return 4.2 - 0.9*(y-yb)/height }
	for _, side := range []float64{-1, 1} {
		for _, z := range []float64{cz - 2.6, cz + 2.6} {
			g.Use(Wood).Fill(s.wheel(V(cx+side*4.9, y0+1.5, z), 1.75, 0.45, true))
			g.Use(Charcoal).Fill(s.wheel(V(cx+side*5.2, y0+1.5, z), 0.5, 0.4, true))
		}
	}
	g.Use(Wood).Fill(Block(V(cx-4.5, y0+1.0, cz-4.5), V(cx+4.5, yb+0.6, cz+4.5)))
	body := Func{Box{V(cx-4.3, yb, cz-4.3), V(cx+4.3, yb+height, cz+4.3)}, func(p Vec3) bool {
		h := half(p.Y)
		return math.Max(math.Abs(p.X-cx), math.Abs(p.Z-cz)) <= h
	}}
	inner := Func{body.Box, func(p Vec3) bool {
		h := half(p.Y) - 1.0
		return math.Max(math.Abs(p.X-cx), math.Abs(p.Z-cz)) <= h && p.Y > yb+0.6
	}}
	g.Use(Sand).Fill(body)
	g.Use(Charcoal).Fill(Func{body.Box, func(p Vec3) bool {
		h := half(p.Y) - 0.6
		return math.Max(math.Abs(p.X-cx), math.Abs(p.Z-cz)) <= h
	}})
	g.Carve(inner)
	// Storey floors.
	for y := yb + 5.5; y < yb+height-1; y += 5.5 {
		g.Use(Wood).Fill(Intersect(body, Block(V(cx-5, y, cz-5), V(cx+5, y+0.6, cz+5))))
	}
	// Hides in overlapping patches, timber corner posts and storey rails.
	outer := Func{body.Box, func(p Vec3) bool {
		h := half(p.Y)
		return math.Max(math.Abs(p.X-cx), math.Abs(p.Z-cz)) > h-0.6
	}}
	g.Paint(Terracotta, Intersect(outer, Func{body.Box, func(p Vec3) bool {
		u := p.X + p.Z
		return Hash01(int(math.Floor(u/2.2)), int(math.Floor((p.Y-yb)/1.8)), 101) < 0.45
	}}))
	g.Paint(Wood, Intersect(outer, Func{body.Box, func(p Vec3) bool {
		h := half(p.Y)
		corner := math.Abs(p.X-cx) > h-1.1 && math.Abs(p.Z-cz) > h-1.1
		rail := math.Mod(p.Y-yb+100, 5.5) < 0.7
		return corner || rail
	}}))
	// Arrow loops on the sides, an open back with a ladder visible.
	for _, y := range []float64{yb + 7, yb + 12.5} {
		for _, dz := range []float64{-1.5, 1.5} {
			g.Carve(Block(V(cx-5, y, cz+dz-0.5), V(cx+5, y+1.8, cz+dz+0.5)))
		}
	}
	g.Carve(Block(V(cx-2.2, yb+0.6, cz+2), V(cx+2.2, yb+4.6, cz+5)))
	g.Use(Wood).Fill(Block(V(cx-1, yb+0.6, cz+1.6), V(cx+1, yb+5.5, cz+2.2)))
	g.Paint(Charcoal, Block(V(cx-0.6, yb+1, cz+1.4), V(cx+0.6, yb+5.5, cz+2.4)))
	// Fighting deck with a crenellated wooden parapet.
	top := yb + height
	ht := half(top)
	g.Use(Wood).Fill(Block(V(cx-ht-0.3, top-0.4, cz-ht-0.3), V(cx+ht+0.3, top+0.4, cz+ht+0.3)))
	g.Use(Wood).Build(kit.Battlements{Min: P(cx-ht-0.3, cz-ht-0.3), Max: P(cx+ht+0.3, cz+ht+0.3), Top: top - 0.2, Period: 2.2, Rim: 0.9})
	// Raised boarding bridge on the wall-facing side, held by chains.
	bz := cz - half(top-4) - 0.6
	g.Use(Wood).Fill(Block(V(cx-2.6, top-4.6, bz-0.4), V(cx+2.6, top+1.6, bz+0.4)))
	g.Paint(Charcoal, Block(V(cx-2.6, top-0.2, bz-0.5), V(cx+2.6, top+0.4, bz+0.5)))
	for _, side := range []float64{-1, 1} {
		g.Use(Charcoal).Fill(Capsule{V(cx+side*2.2, top+1.3, bz), V(cx+side*2.6, top+2.4, cz-1), 0.45})
	}
	s.flag(g, V(cx+2, top+0.5, cz+2), 4.5, Navy, Gold)
	// Push poles out the back.
	for _, dx := range []float64{-2.5, 2.5} {
		g.Use(Wood).Fill(Capsule{V(cx+dx, y0+2.4, cz+4.4), V(cx+dx, y0+1.4, cz+9), 0.5})
	}
}

// trebuchet builds a counterweight trebuchet throwing toward -Z. theta is the long
// arm's angle in the YZ plane (0 = pointing back to +Z, π/2 = straight up); a negative
// angle is the cocked pose with the sling laid in its trough.
func (s SiegeCamp) trebuchet(g *Grid, c Vec2, theta, sc float64) {
	y0 := s.height(c.X, c.Y) - 0.2
	H := 15.0 * sc
	L, S := 22.0*sc, 5.4*sc
	fx := 3.8 * sc // Frame half spacing.
	pivot := V(c.X, y0+H, c.Y)
	g.Use(Wood)
	for _, side := range []float64{-1, 1} {
		x := c.X + side*fx
		// Sill, then the A-frame: two raking legs, a king post, a cross brace and a side shore.
		g.Fill(Block(V(x-0.7, y0-0.3, c.Y-9*sc), V(x+0.7, y0+1.1, c.Y+9*sc)))
		for _, dz := range []float64{-7.8 * sc, 7.8 * sc} {
			g.Fill(Capsule{V(x, y0+1, c.Y+dz), V(x, pivot.Y-0.3, c.Y), 0.75})
		}
		g.Fill(Capsule{V(x, y0+1, c.Y), V(x, pivot.Y, c.Y), 0.65})
		g.Fill(Block(V(c.X+side*6.5*sc-0.6, y0-0.3, c.Y-1.6), V(c.X+side*6.5*sc+0.6, y0+0.9, c.Y+1.6)))
		g.Fill(Capsule{V(c.X+side*6.5*sc, y0+0.8, c.Y), V(x+side*0.4, y0+H*0.66, c.Y), 0.6})
		g.With(Charcoal, func() { g.Fill(BlockAt(P(x, c.Y), 1.8, 1.8, pivot.Y-0.9, pivot.Y+0.9)) })
	}
	for _, dz := range []float64{-8 * sc, 0, 8 * sc} {
		g.Fill(Block(V(c.X-6.5*sc-0.6, y0-0.3, c.Y+dz-0.6), V(c.X+6.5*sc+0.6, y0+0.8, c.Y+dz+0.6)))
	}
	// Windlass across the rear legs with capstan bars.
	wy, wz := y0+3.2*sc, c.Y+5.6*sc
	g.Fill(Capsule{V(c.X-fx-1.4, wy, wz), V(c.X+fx+1.4, wy, wz), 0.75})
	// Sling trough.
	g.Fill(Block(V(c.X-1.3, y0-0.2, c.Y-9*sc), V(c.X+1.3, y0+0.6, c.Y+1)))
	// Axle.
	g.With(Charcoal, func() { g.Fill(Capsule{pivot.Sub(V(fx+1, 0, 0)), pivot.Add(V(fx+1, 0, 0)), 0.65}) })
	// Throwing arm, tapering from the counterweight end, with iron bands.
	dir := V(0, math.Sin(theta), math.Cos(theta))
	long, short := pivot.Add(dir.Scale(L)), pivot.Sub(dir.Scale(S))
	g.Fill(Capsule{pivot, pivot.Add(dir.Scale(L * 0.5)), 0.95}, Capsule{pivot, long, 0.7}, Capsule{pivot, short, 1.15})
	g.With(Charcoal, func() {
		for _, t := range []float64{-2.5, 2.5, 8} {
			at := pivot.Add(dir.Scale(t * sc))
			g.Fill(Capsule{at.Sub(dir.Scale(0.35)), at.Add(dir.Scale(0.35)), 1.15})
		}
	})
	// Counterweight box hanging plumb under the short end, heaped with stones.
	bh, bw, bd := 4.2*sc, 2.5*sc, 2.3*sc
	boxTop := short.Y - 1.5
	g.Fill(Capsule{short.Sub(V(bw-0.3, 0, 0)), short.Add(V(bw-0.3, 0, 0)), 0.6})
	for _, side := range []float64{-1, 1} {
		g.Fill(Capsule{short.Add(V(side*(bw-0.5), 0, 0)), V(c.X+side*(bw-0.5), boxTop+0.3, short.Z), 0.5})
	}
	box := Block(V(c.X-bw, boxTop-bh, short.Z-bd), V(c.X+bw, boxTop, short.Z+bd))
	g.Fill(box)
	g.Paint(Charcoal, Intersect(box, Func{box.Bounds(), func(p Vec3) bool {
		return math.Abs(p.X-c.X) > bw-0.7 && math.Abs(p.Z-short.Z) > bd-0.7 || math.Abs(p.Y-(boxTop-bh/2)) < 0.35
	}}))
	g.Use(DarkStone).Fill(Func{Box{V(c.X-bw+0.5, boxTop-0.2, short.Z-bd+0.5), V(c.X+bw-0.5, boxTop+0.9, short.Z+bd-0.5)}, func(p Vec3) bool {
		return Hash01(int(math.Floor(p.X*1.2)), int(math.Floor(p.Z*1.2)), 111) < 0.65
	}})
	if theta < 0 {
		// Cocked: the sling runs from the arm tip into the trough, a stone in its pouch.
		pouch := V(c.X, y0+1.2, c.Y-3*sc)
		g.Use(Sand).Fill(Capsule{long, long.Add(V(0, 0, -3)), 0.5}, Capsule{long.Add(V(0, 0, -3)), pouch, 0.5})
		g.Use(Sand).Fill(Ellipsoid{pouch, V(1.4, 0.7, 1.7)})
		g.Use(Stone).FillVolume(Sphere(pouch.Add(V(0, 1.1, 0)), 1.2))
		return
	}
	// Throwing: the sling whips forward over the top and the stone flies at the wall.
	end := long.Add(V(0, 2.5, -6))
	g.Use(Sand).Fill(Capsule{long, end, 0.5})
	g.Use(Sand).Fill(Ellipsoid{end, V(1.0, 0.7, 1.3)})
	g.Use(DarkStone).FillVolume(Sphere(end.Add(V(0, 1.2, -8)), 1.3))
}

// pile is a pyramid of stone balls.
func (s SiegeCamp) pile(g *Grid, c Vec2) {
	y := s.height(c.X, c.Y) + 0.5
	r := 0.75
	for layer, n := 0, 3; n > 0; layer, n = layer+1, n-1 {
		off := float64(n-1) * r
		for i := 0; i < n; i++ {
			for k := 0; k < n; k++ {
				col := Stone
				if Hash01(int(c.X), int(c.Y), layer, i, k) < 0.35 {
					col = DarkStone
				}
				at := V(c.X-off+float64(i)*2*r, y+float64(layer)*1.1, c.Y-off+float64(k)*2*r)
				g.Use(col).FillVolume(Sphere(at, r+0.1))
			}
		}
	}
}

// wagon is a four-wheeled cart along Z with a load: 0 stones, 1 barrels, 2 canvas cover.
func (s SiegeCamp) wagon(g *Grid, c Vec2, load int) {
	y0 := s.height(c.X, c.Y)
	bed := Block(V(c.X-1.4, y0+1.3, c.Y-2.6), V(c.X+1.4, y0+2.1, c.Y+2.6))
	g.Use(Wood).Fill(bed)
	g.Fill(Block(V(c.X-1.4, y0+2.1, c.Y-2.6), V(c.X+1.4, y0+2.8, c.Y+2.6)))
	g.Carve(Block(V(c.X-0.9, y0+2.1, c.Y-2.1), V(c.X+0.9, y0+3, c.Y+2.1)))
	for _, sx := range []float64{-1, 1} {
		for _, dz := range []float64{-1.7, 1.7} {
			g.Use(Wood).Fill(s.wheel(V(c.X+sx*1.85, y0+1.15, c.Y+dz), 1.15, 0.38, true))
			g.Use(Charcoal).Fill(s.wheel(V(c.X+sx*2.0, y0+1.15, c.Y+dz), 0.4, 0.38, true))
		}
	}
	g.Use(Wood).Fill(Capsule{V(c.X, y0+1.6, c.Y-2.6), V(c.X, y0+0.8, c.Y-5.5), 0.45})
	switch load {
	case 0:
		for i := 0; i < 4; i++ {
			at := V(c.X+[]float64{-0.6, 0.6, 0, 0}[i], y0+2.7+[]float64{0, 0, 0, 0.9}[i], c.Y+[]float64{-1.2, 0.3, 1.5, 0}[i])
			g.Use(Stone).FillVolume(Sphere(at, 0.8))
		}
	case 1:
		for _, dz := range []float64{-1.2, 1.2} {
			g.Use(Wood).Fill(Cylinder{Center: P(c.X, c.Y+dz), Radius: 0.95, Y0: y0 + 2.1, Y1: y0 + 3.9})
			g.Paint(Charcoal, Cylinder{Center: P(c.X, c.Y+dz), Radius: 1, Y0: y0 + 3.0, Y1: y0 + 3.3})
		}
	default:
		cover := Func{Box{V(c.X-1.6, y0+2.1, c.Y-2.4), V(c.X+1.6, y0+4.6, c.Y+2.4)}, func(p Vec3) bool {
			d := math.Hypot(p.X-c.X, p.Y-(y0+2.4))
			return d <= 1.6 && d > 0.7
		}}
		g.Use(White).Fill(cover)
		g.Paint(Charcoal, Block(V(c.X-1.6, y0+2.1, c.Y+2.0), V(c.X+1.6, y0+4.6, c.Y+2.4)))
	}
}

// supplies are wagons and ammunition piles beside the engines.
func (s SiegeCamp) supplies(g *Grid) {
	for _, p := range []Vec2{P(-49, 4), P(-26, 2), P(13, 1), P(25, 3), P(48, 12)} {
		s.pile(g, p)
	}
	s.wagon(g, P(-50, 17), 0)
	s.wagon(g, P(-26, 16), 1)
	s.wagon(g, P(14, 14), 0)
	s.wagon(g, P(24, 17), 2)
	s.wagon(g, P(-12, 19), 2)
}

// pavilion is a round striped tent with a cone roof, a door, a pole and a pennant.
func (s SiegeCamp) pavilion(g *Grid, c Vec2, r float64, a, b Color, door float64, seed int) {
	y0 := s.height(c.X, c.Y)
	wallH := 2.6
	roofH := r*1.15 + 1.2
	stripe := func(p Vec3) bool {
		ang := p.Plan().Sub(c).Angle()
		k := int(math.Floor((ang + math.Pi) / (2 * math.Pi) * 10))
		return k%2 == 0
	}
	shell := Cylinder{Center: c, Radius: r, Inner: r - 0.9, Y0: y0 - 0.8, Y1: y0 + wallH}
	g.Use(a).Fill(shell)
	g.Paint(b, Func{shell.Bounds(), stripe})
	g.Use(Charcoal).Fill(Cylinder{Center: c, Radius: r - 0.85, Inner: r - 1.3, Y0: y0 - 0.2, Y1: y0 + wallH})
	roof := Frustum{c, r + 0.7, 0.3, y0 + wallH - 0.2, y0 + wallH + roofH}
	g.Use(a).FillVolume(roof)
	g.Paint(b, Func{roof.Bounds(), func(p Vec3) bool { return stripe(p) && p.Y > y0+wallH+0.6 && p.Y < y0+wallH+roofH*0.55 }})
	// Scalloped valance.
	g.Use(b).Fill(Func{Box{V(c.X-r-0.8, y0+wallH-0.9, c.Y-r-0.8), V(c.X+r+0.8, y0+wallH+0.4, c.Y+r+0.8)}, func(p Vec3) bool {
		d := p.Plan().Sub(c)
		if q := d.Length(); q > r+0.75 || q < r-0.1 {
			return false
		}
		f := math.Mod((d.Angle()+math.Pi)/(2*math.Pi)*20+20, 1)
		return p.Y > y0+wallH-0.9+0.9*math.Abs(f-0.5)*2 || p.Y > y0+wallH-0.2
	}})
	// Door with the flap pulled aside.
	dz := door * r
	g.Carve(Block(V(c.X-0.9, y0-0.2, c.Y+dz-1.2), V(c.X+0.9, y0+2.1, c.Y+dz+1.2)))
	// Pole and pennant.
	topY := y0 + wallH + roofH
	g.Use(Wood).Fill(Cylinder{Center: c, Radius: 0.45, Y0: topY - 1, Y1: topY + 2.4})
	pen := []Color{Red, Gold, Navy, White}[seed%4]
	if pen == a || pen == b {
		pen = []Color{Red, Gold, Navy, White}[(seed+1)%4]
	}
	g.Use(pen).Fill(Func{Box{V(c.X, topY+0.6, c.Y-0.45), V(c.X+3, topY+2.4, c.Y+0.45)}, func(p Vec3) bool {
		u := (p.X - c.X) / 3
		return math.Abs(p.Y-(topY+1.8)) <= 0.62*(1-u)+0.15
	}})
}

// ridgeTent is a small A-tent along X, open at its +X end.
func (s SiegeCamp) ridgeTent(g *Grid, c Vec2, col, trim Color) {
	y0 := s.height(c.X, c.Y)
	h, hw, hl := 3.4, 2.0, 2.5
	tent := Func{Box{V(c.X-hl, y0-0.6, c.Y-hw), V(c.X+hl, y0+h, c.Y+hw)}, func(p Vec3) bool {
		return math.Abs(p.Z-c.Y) <= hw*(1-(p.Y-y0)/h)
	}}
	g.Use(col).Fill(tent)
	g.Use(Charcoal).Fill(Func{tent.Box, func(p Vec3) bool {
		return math.Abs(p.Z-c.Y) <= hw*(1-(p.Y-y0)/h)-0.6 && p.X < c.X+hl-0.6
	}})
	g.Carve(Func{tent.Box, func(p Vec3) bool {
		return math.Abs(p.Z-c.Y) <= hw*(1-(p.Y-y0)/h)-1.0 && p.Y > y0-0.2
	}})
	g.Paint(trim, Func{tent.Box, func(p Vec3) bool { return p.Y > y0+h-1.1 || math.Abs(p.X-c.X) > hl-0.5 && p.Y > y0+0.3 }})
}

// banner is a tall pole with a crossbar and a hanging swallow-tailed banner.
func (s SiegeCamp) banner(g *Grid, c Vec2, h float64, a, b Color) {
	y0 := s.height(c.X, c.Y)
	top := y0 + h
	g.Use(Wood).Fill(Cylinder{Center: c, Radius: 0.45, Y0: y0 - 0.6, Y1: top})
	g.Fill(Block(V(c.X-1.6, top-0.9, c.Y-0.4), V(c.X+1.6, top-0.3, c.Y+0.4)))
	g.Use(Gold).Fill(Sphere(c.At(top+0.4), 0.6))
	cloth := Block(V(c.X-1.4, top-5, c.Y-0.45), V(c.X+1.4, top-0.9, c.Y+0.45))
	g.Use(a).Fill(cloth)
	g.Paint(b, Block(V(c.X-1.5, top-3.6, c.Y-0.5), V(c.X+1.5, top-2.8, c.Y+0.5)), Block(V(c.X-0.4, top-5, c.Y-0.5), V(c.X+0.4, top-0.9, c.Y+0.5)))
	g.Carve(Func{Box{V(c.X-1, top-5.2, c.Y-0.5), V(c.X+1, top-3.8, c.Y+0.5)}, func(p Vec3) bool {
		return p.Y-(top-5) < 1.2-math.Abs(p.X-c.X)*1.2
	}})
}

// campfire is a stone ring with logs and a red and gold flame.
func (s SiegeCamp) campfire(g *Grid, c Vec2) {
	y0 := s.height(c.X, c.Y)
	g.Use(DarkStone).Fill(Cylinder{Center: c, Radius: 1.6, Inner: 0.9, Y0: y0 - 0.3, Y1: y0 + 0.6})
	g.Use(Wood).Fill(Capsule{V(c.X-1.1, y0+0.4, c.Y-0.5), V(c.X+1.1, y0+0.4, c.Y+0.5), 0.45})
	g.Use(Red).FillVolume(Cone(c, 1.2, y0+0.4, 2.6))
	g.Use(Gold).FillVolume(Cone(c, 0.7, y0+0.6, 3.4))
	// A cooking tripod.
	for k := 0; k < 3; k++ {
		a := float64(k) * 2 * math.Pi / 3
		g.Use(Wood).Fill(Capsule{Polar(c, 1.8, a).At(y0), c.At(y0 + 3.6), 0.4})
	}
}

// horse is a simple horse facing along heading (plan direction) at a plan position.
func (s SiegeCamp) horse(g *Grid, c Vec2, heading Vec2, coat, blanket Color) {
	y0 := s.height(c.X, c.Y)
	f := heading.Scale(1 / heading.Length())
	side := P(-f.Y, f.X)
	at := func(along, up, across float64) Vec3 {
		p := c.Add(f.Scale(along)).Add(side.Scale(across))
		return p.At(y0 + up)
	}
	g.Use(coat)
	body := Capsule{at(-1.2, 2.75, 0), at(1.2, 2.9, 0), 0.95}
	g.FillVolume(body)
	for _, a := range []float64{-1.2, 1.25} {
		for _, b := range []float64{-0.5, 0.5} {
			g.Fill(Capsule{at(a, 2.4, b), at(a, -0.2, b), 0.47})
		}
	}
	g.Fill(Capsule{at(1.7, 3.2, 0), at(2.6, 4.7, 0), 0.65})
	g.Fill(Capsule{at(2.6, 4.7, 0), at(3.6, 3.9, 0), 0.55})
	g.With(Charcoal, func() {
		g.Fill(Capsule{at(1.75, 3.9, 0), at(2.4, 5.2, 0), 0.47})
		g.Fill(Capsule{at(-2.1, 3.0, 0), at(-2.5, 1.5, 0), 0.5})
	})
	if blanket != coat {
		g.Paint(blanket, Func{body.Bounds().Grow(0.6), func(p Vec3) bool {
			d := p.Plan().Sub(c).Dot(f)
			return math.Abs(d) < 0.8 && p.Y > y0+2.9
		}})
	}
}

// camp lays out the pavilions, command tent, fires, horses and a back row of A-tents.
func (s SiegeCamp) camp(g *Grid) {
	schemes := [][2]Color{{Red, White}, {Navy, Gold}, {Gold, White}, {White, Red}, {Gold, Navy}, {White, Gold}}
	k := 0
	for _, row := range []struct {
		z, door float64
		xs      []float64
	}{
		{32, 1, []float64{-48, -40, -32, -24, -16, -8, 25, 32}},
		{40.5, -1, []float64{-45, -37, -29, -21, -13, -5, 1, 28}},
	} {
		for i, x := range row.xs {
			sch := schemes[int(Hash01(k, 120)*float64(len(schemes)))]
			r := 2.7 + 0.7*Hash01(k, i, 121)
			z := row.z + (Hash01(k, i, 122)-0.5)*1.6
			s.pavilion(g, P(x, z), r, sch[0], sch[1], row.door, k*7+i)
			k++
		}
	}
	// Back row of soldiers' A-tents.
	for i, x := 0, -44.0; x <= 34; i, x = i+1, x+6.5 {
		if math.Abs(x-siegeCampCommand.X) < 8 {
			continue
		}
		col, trim := White, Red
		if i%3 == 1 {
			col, trim = Sand, Navy
		} else if i%3 == 2 {
			col, trim = White, Navy
		}
		s.ridgeTent(g, P(x, 45.6), col, trim)
	}
	s.command(g)
	for _, p := range []Vec2{P(-36, 36.5), P(-12, 36.5), P(21, 30), P(-20, 27.5), P(-2, 33)} {
		s.campfire(g, p)
	}
	// Horse picket line with a hay pile.
	pl := siegeCampPicket
	y := s.height(pl.X, pl.Y)
	for _, dz := range []float64{-8, 0, 8} {
		g.Use(Wood).Fill(Cylinder{Center: P(pl.X, pl.Y+dz), Radius: 0.5, Y0: y - 0.8, Y1: y + 3.4})
	}
	g.Use(Sand).Fill(Capsule{V(pl.X, y+3.0, pl.Y-8), V(pl.X, y+3.0, pl.Y+8), 0.45})
	coats := []Color{Wood, Charcoal, White, Sand, Wood, Charcoal, Wood, White}
	blankets := []Color{Red, Navy, Navy, Red, Gold, Red, Navy, Gold}
	for i, dz := range []float64{-5.5, 0, 5.5} {
		s.horse(g, P(pl.X-4.2, pl.Y+dz), P(1, 0), coats[i], blankets[i])
		s.horse(g, P(pl.X+4.2, pl.Y+dz+2), P(-1, 0), coats[i+4], blankets[i+4])
	}
	g.Use(Gold).FillVolume(Ellipsoid{V(pl.X-1, s.height(pl.X, pl.Y-11.5)+0.4, pl.Y-11.5), V(2.2, 1.5, 1.6)})
}

// command is the commander's large striped pavilion on its knoll, with banners.
func (s SiegeCamp) command(g *Grid) {
	c := siegeCampCommand
	y0 := s.height(c.X, c.Y)
	hw, hd, wallH := 6.5, 4.5, 3.0
	lo, hi := V(c.X-hw, y0-0.8, c.Y-hd), V(c.X+hw, y0+wallH, c.Y+hd)
	g.Use(Gold).Fill(Block(lo, hi))
	g.Paint(Navy, Func{Box{lo, hi}, func(p Vec3) bool {
		return math.Mod(p.X-lo.X+math.Mod(p.Z-lo.Z, 0.01), 2.6) < 1.3 && math.Abs(p.X-c.X) < hw-0.6 || math.Abs(p.X-c.X) >= hw-0.6 && math.Mod(p.Z-lo.Z, 2.6) < 1.3
	}})
	g.Use(Charcoal).Fill(Block(lo.Add(V(0.9, 0.6, 0.9)), hi.Sub(V(0.9, 0, 0.9))))
	g.Carve(Block(lo.Add(V(1.3, 0.8, 1.3)), hi.Add(V(-1.3, 1, -1.3))))
	// Steep tent roof with a ridge along X: hollow, Navy with gold edge.
	rh := 6.5
	roof := Func{Box{V(lo.X-0.8, hi.Y-0.2, lo.Z-0.8), V(hi.X+0.8, hi.Y+rh, hi.Z+0.8)}, func(p Vec3) bool {
		t := (p.Y - hi.Y + 0.2) / rh
		return math.Abs(p.Z-c.Y) <= (hd+0.8)*(1-t) && math.Abs(p.X-c.X) <= hw+0.8-3.2*t
	}}
	g.Use(Navy).FillVolume(roof)
	g.Carve(Func{roof.Box, func(p Vec3) bool {
		t := (p.Y - hi.Y + 0.2) / rh
		return math.Abs(p.Z-c.Y) <= (hd+0.8)*(1-t)-1.2 && math.Abs(p.X-c.X) <= hw+0.8-3.2*t-1.2
	}})
	g.Paint(Gold, Func{roof.Box, func(p Vec3) bool {
		return p.Y < hi.Y+0.7 || p.Y > hi.Y+rh-1.3 || math.Mod(p.X-lo.X+100, 2.6) < 1.0
	}})
	// Valance scallops.
	g.Use(Gold).Fill(Func{Box{V(lo.X-0.8, hi.Y-1, lo.Z-0.8), V(hi.X+0.8, hi.Y, hi.Z+0.8)}, func(p Vec3) bool {
		edge := math.Abs(p.Z-c.Y) > hd-0.1 || math.Abs(p.X-c.X) > hw-0.1
		u := p.X + p.Z
		f := math.Mod(u+100, 1.6) / 1.6
		return edge && p.Y > hi.Y-1+math.Abs(f-0.5)*2
	}})
	// Open front with the flaps pulled back.
	g.Carve(Block(V(c.X-2, y0-0.2, lo.Z-0.5), V(c.X+2, y0+3, lo.Z+1.4)))
	// Ridge poles with pennants and flanking banners.
	for _, dx := range []float64{-4, 4} {
		top := hi.Y + rh
		g.Use(Wood).Fill(Cylinder{Center: P(c.X+dx, c.Y), Radius: 0.45, Y0: top - 1.5, Y1: top + 2.8})
		g.Use(Red).Fill(Func{Box{V(c.X+dx, top+0.9, c.Y-0.45), V(c.X+dx+3.6, top+2.8, c.Y+0.45)}, func(p Vec3) bool {
			u := (p.X - c.X - dx) / 3.6
			return math.Abs(p.Y-(top+2.1)) <= 0.7*(1-u)+0.15
		}})
	}
	s.banner(g, P(c.X-4, c.Y-hd-3), 9, Navy, Gold)
	s.banner(g, P(c.X+4, c.Y-hd-3), 9, Gold, Navy)
	s.banner(g, P(c.X, c.Y-hd-6.5), 7, Red, Gold)
}
