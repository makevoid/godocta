package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// HauntedManor is a spooky Victorian mansion crowning a grassy hill: a Second Empire
// main block with a banded mansard and iron cresting, a tall square tower with its own
// mansard and a crooked weathervane, a Queen Anne wing with a steep gable, bay window
// and a witch-hat turret, a wraparound porch, tall chimneys, and peeling siding with
// boarded, broken and a few lit windows. Down the slope lie a wrought-iron graveyard with
// tombstones, crosses and a mausoleum, gnarled dead trees, a pumpkin patch, a swampy
// pond, and a crumbling stone wall whose gate starts the path winding up to the porch.
// The house faces +Z.
type HauntedManor struct{}

func init() { Register(HauntedManor{}) }

func (HauntedManor) Name() string  { return "haunted_manor" }
func (HauntedManor) Title() string { return "Haunted Manor" }

const (
	hauntedManorTop   = 9.0  // Hilltop plateau.
	hauntedManorRim   = 0.8  // Site rim.
	hauntedManorFloor = 10.5 // Ground floor of the house.
)

var (
	hauntedManorGrave     = P(24, 20) // Graveyard terrace center.
	hauntedManorGraveHalf = P(9.5, 7.5)
	hauntedManorPond      = P(-24, 22)
	hauntedManorPath      = []Vec2{P(-5.8, 10.4), P(-5, 14), P(1, 18), P(4.5, 23), P(1.5, 29), P(-2, 33.5), P(0, 40)}
	hauntedManorSidePath  = []Vec2{P(3.8, 21), P(10, 20.5), P(15.8, 20)}
)

// hill is the bare hill: a plateau easing down to a low elliptical rim.
func (HauntedManor) hill(x, z float64) float64 {
	e1 := math.Hypot((x-1)/36, (z+11)/24)
	e2 := math.Hypot(x/50, (z+4)/44)
	if e2 >= 1 {
		return hauntedManorRim
	}
	if e1 <= 1 {
		return hauntedManorTop
	}
	t := (e1 - 1) / ((e1 - 1) + (1 - e2))
	s := 1 - t
	h := hauntedManorRim + (hauntedManorTop-hauntedManorRim)*s*s*(3-2*s)
	return h + 2.4*s*(1-s)*math.Sin(x*0.23+z*0.11)*math.Sin(z*0.19-x*0.08)
}

func hauntedManorSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

// graveSD is the signed distance to the graveyard terrace's rounded rectangle.
func (HauntedManor) graveSD(x, z float64) float64 {
	const rc = 2.0
	qx := math.Abs(x-hauntedManorGrave.X) - (hauntedManorGraveHalf.X - rc)
	qz := math.Abs(z-hauntedManorGrave.Y) - (hauntedManorGraveHalf.Y - rc)
	return math.Hypot(math.Max(qx, 0), math.Max(qz, 0)) + math.Min(math.Max(qx, qz), 0) - rc
}

func (m HauntedManor) graveLevel() float64 { return m.hill(hauntedManorGrave.X, hauntedManorGrave.Y) }
func (m HauntedManor) pondLevel() float64 {
	return m.hill(hauntedManorPond.X, hauntedManorPond.Y) - 0.4
}

func (HauntedManor) pondNorm(x, z float64) float64 {
	return math.Hypot((x-hauntedManorPond.X)/8.5, (z-hauntedManorPond.Y)/5.8)
}

// ground is the finished terrain height with the graveyard terrace and the pond basin.
func (m HauntedManor) ground(x, z float64) float64 {
	h := m.hill(x, z)
	if w := hauntedManorSmooth(1 - m.graveSD(x, z)/4.5); w > 0 {
		h += (m.graveLevel() - h) * w
	}
	if w := hauntedManorSmooth((1.7 - m.pondNorm(x, z)) / 0.7); w > 0 {
		h += (m.pondLevel() - h) * w
	}
	return h
}

// The house is authored at its own scale, then enlarged about a pivot on the plateau.
const hauntedManorScale = 1.3

var hauntedManorPivot = V(2, hauntedManorTop, -9)

func hauntedManorLocal(p Vec3) Vec3 {
	return hauntedManorPivot.Add(p.Sub(hauntedManorPivot).Scale(1 / hauntedManorScale))
}

func hauntedManorWorld(p Vec3) Vec3 {
	return hauntedManorPivot.Add(p.Sub(hauntedManorPivot).Scale(hauntedManorScale))
}

func hauntedManorWorldBox(b Box) Box { return Box{hauntedManorWorld(b.Min), hauntedManorWorld(b.Max)} }

// hauntedManorScaled maps a house-local solid into the world.
type hauntedManorScaled struct{ S Solid }

func (s hauntedManorScaled) Contains(p Vec3) bool { return s.S.Contains(hauntedManorLocal(p)) }
func (s hauntedManorScaled) Bounds() Box          { return hauntedManorWorldBox(s.S.Bounds()) }

// hauntedManorPen draws house-local solids into the grid.
type hauntedManorPen struct{ g *Grid }

func (p hauntedManorPen) wrap(ss []Solid) []Solid {
	out := make([]Solid, len(ss))
	for i, s := range ss {
		out[i] = hauntedManorScaled{s}
	}
	return out
}

func (p hauntedManorPen) Use(c Color) hauntedManorPen      { p.g.Use(c); return p }
func (p hauntedManorPen) With(c Color, f func())           { p.g.With(c, f) }
func (p hauntedManorPen) Fill(ss ...Solid) hauntedManorPen { p.g.Fill(p.wrap(ss)...); return p }
func (p hauntedManorPen) FillVolume(ss ...Solid) hauntedManorPen {
	p.g.FillVolume(p.wrap(ss)...)
	return p
}
func (p hauntedManorPen) Carve(ss ...Solid) hauntedManorPen { p.g.Carve(p.wrap(ss)...); return p }
func (p hauntedManorPen) Paint(c Color, ss ...Solid) hauntedManorPen {
	p.g.Paint(c, p.wrap(ss)...)
	return p
}

// pathDist is the plan distance to a polyline.
func hauntedManorPathDist(line []Vec2, p Vec2) float64 {
	best := math.Inf(1)
	for i := 0; i+1 < len(line); i++ {
		a, b := line[i], line[i+1]
		ab := b.Sub(a)
		t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
		best = math.Min(best, p.Sub(a.Add(ab.Scale(t))).Length())
	}
	return best
}

func (m HauntedManor) Build(g *Grid) Order {
	m.terrain(g)
	m.house(g)
	m.garden(g)
	m.graveyard(g)
	m.pond(g)
	m.boundary(g)
	m.trees(g)
	m.spirits(g)
	return Sweep(P(0, -9), math.Pi/2)
}

// recolor sets a new color on cells chosen by f.
func (HauntedManor) recolor(g *Grid, region Box, f func(c Cell, p Vec3, col Color) (Color, bool)) {
	brush := g.Brush()
	for _, c := range g.Cells() {
		p := g.World(c)
		if !region.Contains(p) {
			continue
		}
		col, _ := g.ColorOf(c)
		if nc, ok := f(c, p, col); ok && nc != col {
			g.Remove(c)
			g.Use(nc)
			g.Add(c)
		}
	}
	g.Use(brush)
}

func (m HauntedManor) terrain(g *Grid) {
	box := Box{V(-50.5, 0, -48.5), V(50.5, 10, 40.5)}
	g.Use(Foliage).FillVolume(Func{box, func(p Vec3) bool {
		if math.Hypot(p.X/50, (p.Z+4)/44) > 1 {
			return false
		}
		h := m.ground(p.X, p.Z)
		return p.Y <= h && p.Y >= math.Max(0, h-1.3)
	}})
	// Patchy grass, bare earth and the gravel path.
	m.recolor(g, box, func(c Cell, p Vec3, col Color) (Color, bool) {
		n := math.Sin(p.X*0.29+math.Sin(p.Z*0.21)*2) * math.Sin(p.Z*0.26+p.X*0.07+math.Sin(p.X*0.13))
		if d := hauntedManorPathDist(hauntedManorPath, p.Plan()); d < 1.7 || hauntedManorPathDist(hauntedManorSidePath, p.Plan()) < 1.2 {
			if Hash01(int(math.Floor(p.X/1.3)), int(math.Floor(p.Z/1.3)), 5) < 0.3 {
				return DarkStone, true
			}
			return Sand, true
		}
		switch {
		case n > 0.7:
			return Grass, true
		case n < -0.75:
			return Wood, true
		}
		return col, false
	})
	// Boulders on the slopes.
	for i := 0; i < 16; i++ {
		a := Hash01(i, 11) * 2 * math.Pi
		r := 0.55 + 0.35*Hash01(i, 12)
		x, z := math.Sin(a)*50*r, -4+math.Cos(a)*44*r
		if math.Hypot((x-1)/36, (z+11)/24) < 1.1 || m.graveSD(x, z) < 2 || m.pondNorm(x, z) < 1.4 ||
			hauntedManorPathDist(hauntedManorPath, P(x, z)) < 4 {
			continue
		}
		s := 1.0 + 1.2*Hash01(i, 13)
		g.Use(DarkStone).FillVolume(Ellipsoid{V(x, m.ground(x, z)-0.2, z), V(s*1.3, s*0.8, s)})
	}
}

// window cuts a trimmed window into a wall whose outer plane passes through origin and
// faces the plan angle. kind: -1 picks one, 0 dark, 1 boarded, 2 lit.
func (m HauntedManor) window(g hauntedManorPen, n *int, origin Vec2, angle, sill, hw, spring float64, kind int) {
	*n++
	if kind < 0 {
		h := Hash01(*n, 71)
		switch {
		case h < 0.3:
			kind = 1
		case h < 0.42:
			kind = 2
		default:
			kind = 0
		}
	}
	place := func(s Solid) Solid { return Translate(RotateY(s, P(0, 0), angle), origin.At(0)) }
	g.With(Stone, func() {
		g.Fill(place(Intersect(Arch{Sill: V(0.1, sill-0.4, 0), HalfWidth: hw + 0.45, Spring: spring + 0.4, Depth: 0.6, Through: AxisX},
			Block(V(-0.5, sill+spring-0.3, -hw-1), V(0.5, sill+spring+hw+1, hw+1)))))
		g.Fill(place(Block(V(-0.2, sill-0.8, -hw-0.7), V(0.7, sill-0.2, hw+0.7))))
	})
	g.Carve(place(Arch{Sill: V(-0.3, sill, 0), HalfWidth: hw, Spring: spring, Depth: 2.6, Through: AxisX}))
	switch kind {
	case 1:
		seed := float64(*n)
		g.With(Wood, func() {
			g.Fill(place(Func{Box{V(0.1, sill, -hw-0.4), V(0.7, sill+spring+hw, hw+0.4)}, func(p Vec3) bool {
				r := math.Mod(p.Y-sill+0.3*math.Sin(seed)*p.Z+5, 1.35)
				return r < 0.7
			}}))
		})
	case 2:
		g.With(Gold, func() {
			g.Fill(place(Arch{Sill: V(-0.75, sill, 0), HalfWidth: hw, Spring: spring, Depth: 0.5, Through: AxisX}))
		})
	}
}

// gable is a steep pitched roof volume rising from base with the given pitch.
func (HauntedManor) gable(x0, x1, z0, z1, base, pitch float64, ridge Axis) Func {
	half, mid := (x1-x0)/2, (x0+x1)/2
	if ridge == AxisX {
		half, mid = (z1-z0)/2, (z0+z1)/2
	}
	return Func{Box{V(x0, base, z0), V(x1, base+half*pitch, z1)}, func(p Vec3) bool {
		across := math.Abs(p.X - mid)
		if ridge == AxisX {
			across = math.Abs(p.Z - mid)
		}
		return p.Y >= base && across <= half-(p.Y-base)/pitch
	}}
}

// mansard is a flared steep lower slope of height h, then a hipped top of the given pitch.
func (HauntedManor) mansard(x0, x1, z0, z1, base, h, flare, pitch float64) Func {
	top := base + h + (math.Min(x1-x0, z1-z0)/2-flare)*pitch
	return Func{Box{V(x0, base, z0), V(x1, top, z1)}, func(p Vec3) bool {
		var inset float64
		if p.Y < base {
			return false
		} else if p.Y <= base+h {
			t := 1 - (p.Y-base)/h
			inset = flare * (1 - t*t)
		} else {
			inset = flare + (p.Y-base-h)/pitch
		}
		return p.X >= x0+inset && p.X <= x1-inset && p.Z >= z0+inset && p.Z <= z1-inset
	}}
}

// cresting is an iron rail with spikes around a rectangle at the curb of a mansard.
func (HauntedManor) cresting(g hauntedManorPen, x0, x1, z0, z1, y float64) {
	g.Use(Charcoal).Fill(Func{Box{V(x0, y, z0), V(x1, y+1.8, z1)}, func(p Vec3) bool {
		edge := p.X <= x0+0.8 || p.X >= x1-0.8 || p.Z <= z0+0.8 || p.Z >= z1-0.8
		if !edge {
			return false
		}
		if p.Y <= y+0.6 {
			return true
		}
		along := p.X + p.Z
		return p.Y <= y+1.5 && math.Mod(along+100, 2.0) < 0.8
	}})
}

func (m HauntedManor) chimney(g hauntedManorPen, x0, z0, w, d, y0, top float64) {
	g.Use(DarkStone).Fill(Block(V(x0, y0, z0), V(x0+w, top, z0+d)))
	g.Use(Stone).Fill(Block(V(x0-0.4, top-1.6, z0-0.4), V(x0+w+0.4, top-0.8, z0+d+0.4)))
	g.Use(Charcoal).Fill(
		Block(V(x0+0.1, top, z0+0.2), V(x0+0.9, top+1.3, z0+d-0.2)),
		Block(V(x0+w-0.9, top, z0+0.2), V(x0+w-0.1, top+1.0, z0+d-0.2)))
}

func (m HauntedManor) house(grid *Grid) {
	const (
		F = hauntedManorFloor
		t = 1.2
	)
	g := hauntedManorPen{grid}
	var rooms []Solid
	room := func(s Solid) { g.Carve(s); rooms = append(rooms, s) }
	wins := 0
	win := func(origin Vec2, angle, sill, hw, spring float64, kind int) {
		m.window(g, &wins, origin, angle, sill, hw, spring, kind)
	}
	front, back, east, west := math.Pi/2, -math.Pi/2, 0.0, math.Pi

	// Foundations under every part of the house, hollowed below the floor.
	feet := []Solid{
		Block(V(-16.3, 6, -18.3), V(6.3, F, -3.7)),
		Block(V(-8.3, 6, -6.3), V(-0.7, F, 1.3)),
		Block(V(5.7, 6, -16.3), V(20.3, F, 0.3)),
		Block(V(-15.3, 6, -27.3), V(-4.7, F, -17.7)),
		Prism{[]Vec2{P(9.2, 0), P(10.6, 2.9), P(15.4, 2.9), P(16.8, 0)}, 6, F},
		Cylinder{Center: P(20.5, 0.5), Radius: 3.6, Y0: 6, Y1: F},
	}
	g.Use(DarkStone).Fill(feet...)
	g.Carve(Block(V(-15, 6, -17), V(5, F-1, -5)), Block(V(7, 6, -15), V(19, F-1, -1)), Block(V(-14, 6, -26), V(-6, F-1, -19)))
	g.Use(Stone).Fill(
		Block(V(-16.5, F-0.7, -18.5), V(6.5, F-0.1, -3.5)),
		Block(V(5.5, F-0.7, -16.5), V(20.5, F-0.1, 0.5)),
		Block(V(-15.5, F-0.7, -27.5), V(-4.5, F-0.1, -17.5)),
	)

	// Main block: two tall storeys of dark siding under a banded mansard.
	const mainTop = F + 12
	g.Use(DarkStone).Fill(Block(V(-16, F, -18), V(6, mainTop, -4)))
	g.Use(Stone).Fill(
		Block(V(-16.35, F+5.6, -18.35), V(6, F+6.4, -3.65)),
		Block(V(-16.6, mainTop-0.8, -18.6), V(6, mainTop, -3.4)),
		Func{Box{V(-16.5, mainTop-1.6, -18.5), V(6, mainTop-0.8, -3.5)}, func(p Vec3) bool {
			return math.Mod(p.X+p.Z+100, 2.2) < 0.8
		}},
	)
	main := m.mansard(-16.7, 6.7, -18.7, -3.3, mainTop, 5.5, 1.8, 0.45)
	g.Use(Slate).Fill(main)
	g.Paint(Charcoal, Func{Box{V(-17, mainTop+2.4, -19), V(7, mainTop+3.3, -3)}, func(p Vec3) bool { return true }})
	room(Block(V(-16+t, F, -18+t), V(6-t, mainTop, -4-t)))
	room(Func{main.Box, func(p Vec3) bool {
		q := V(p.X, p.Y+1.0, p.Z)
		return p.X > -16.7+t && p.X < 6.7-t && p.Z > -18.7+t && p.Z < -3.3-t && main.F(q) &&
			m.mansard(-16.7+t, 6.7-t, -18.7+t, -3.3-t, mainTop, 5.5, 1.8, 0.45).F(q)
	}})
	m.cresting(g, -14.9, 4.9, -16.9, -5.1, mainTop+5.5)

	// Queen Anne wing: navy siding, steep front gable, bay window and witch-hat turret.
	const wingTop, wingPitch = F + 11, 1.35
	g.Use(Navy).Fill(Block(V(6, F, -16), V(20, wingTop, 0)))
	g.Use(Stone).Fill(
		Block(V(6, F+5.6, -16.35), V(20.35, F+6.4, 0.35)),
		Block(V(6, wingTop-0.7, -16.5), V(20.5, wingTop, 0.5)),
	)
	wingRoof := m.gable(5.3, 20.7, -16.9, 0.9, wingTop-0.7*wingPitch, wingPitch, AxisZ)
	wingInner := m.gable(5.3+1.5, 20.7-1.5, -17, 1, wingTop-0.7*wingPitch, wingPitch, AxisZ)
	g.Use(Slate).Fill(wingRoof)
	g.Carve(Intersect(wingInner, Block(V(5, wingTop, -17), V(21, 40, 1))))
	g.Use(Navy).Fill(Intersect(wingInner, Block(V(5, wingTop, -1.2), V(21, 40, 0))),
		Intersect(wingInner, Block(V(5, wingTop, -16), V(21, 40, -14.8))))
	room(Block(V(6+t, F, -16+t), V(20-t, wingTop+0.01, -t)))
	room(Intersect(wingInner, Block(V(6+t, wingTop, -16+t), V(20-t, 40, -t))))
	// Half-timber truss in the front gable.
	g.Use(Wood).Fill(
		Block(V(12.6, wingTop+5.5, -0.2), V(13.4, wingTop+9.6, 0.5)),
		Block(V(9.0, wingTop+5.0, -0.2), V(17.0, wingTop+5.7, 0.5)),
		Capsule{V(13, wingTop+9.2, 0.15), V(10.5, wingTop+5.5, 0.15), 0.45},
		Capsule{V(13, wingTop+9.2, 0.15), V(15.5, wingTop+5.5, 0.15), 0.45},
	)
	// Bay window, two storeys with a hipped cap.
	bay := []Vec2{P(9.5, 0), P(10.8, 2.6), P(15.2, 2.6), P(16.5, 0)}
	g.Use(Navy).Fill(Prism{bay, F, F + 10})
	g.Use(Stone).Fill(Prism{[]Vec2{P(9.1, 0), P(10.5, 3.0), P(15.5, 3.0), P(16.9, 0)}, F + 5.6, F + 6.4})
	room(Prism{[]Vec2{P(10.9, -1.3), P(11.6, 1.4), P(14.4, 1.4), P(15.1, -1.3)}, F, F + 9})
	bayCap := []Vec2{P(8.9, -0.2), P(10.4, 3.2), P(15.6, 3.2), P(17.1, -0.2)}
	g.Use(Slate).Fill(Func{Box{V(8.9, F+10, -0.2), V(17.1, F+12.2, 3.2)}, func(p Vec3) bool {
		k := 1 - (p.Y-F-10)/2.2
		q := P(13, -0.5).Add(p.Plan().Sub(P(13, -0.5)).Scale(1 / math.Max(k, 1e-6)))
		return InsidePolygon(bayCap, q)
	}})

	// Back kitchen ell with a charcoal roof.
	const ellTop, ellPitch = F + 7.5, 1.3
	g.Use(DarkStone).Fill(Block(V(-15, F, -27), V(-5, ellTop, -18)))
	g.Use(Stone).Fill(Block(V(-15.4, ellTop-0.7, -27.4), V(-4.6, ellTop, -18)))
	ellInner := m.gable(-15.7, -4.3, -27.7+1.5, -17.3-1.5, ellTop-0.7*ellPitch, ellPitch, AxisX)
	g.Use(Charcoal).Fill(m.gable(-15.7, -4.3, -27.7, -17.3, ellTop-0.7*ellPitch, ellPitch, AxisX))
	g.Carve(Intersect(ellInner, Block(V(-16, ellTop, -28), V(-4, 40, -17))))
	g.Use(DarkStone).Fill(Intersect(ellInner, Block(V(-15, ellTop, -28), V(-13.8, 40, -18))),
		Intersect(ellInner, Block(V(-6.2, ellTop, -28), V(-5, 40, -18))))
	room(Block(V(-15+t, F, -27+t), V(-5-t, ellTop+0.01, -18)))
	room(Intersect(ellInner, Block(V(-15+t, ellTop, -27+t), V(-5-t, 40, -18))))

	// Wraparound porch on the front and west sides.
	g.Use(DarkStone).Fill(Block(V(-20.5, 8, -4), V(6, F-0.8, 3.5)), Block(V(-20.5, 8, -16), V(-16, F-0.8, 3.5)))
	g.Use(Wood).Fill(Block(V(-20.5, F-0.8, -4), V(6, F, 3.5)), Block(V(-20.5, F-0.8, -16), V(-16, F, 3.5)))
	g.Use(Stone)
	var posts []Vec2
	for x := -20.0; x <= 5.6; x += 3.2 {
		if x < -7.6 || x > -1.4 {
			posts = append(posts, P(x, 3.0))
		}
	}
	posts = append(posts, P(-7.4, 3.0), P(-1.6, 3.0), P(5.5, 3.0))
	for z := -15.5; z < 2; z += 3.1 {
		posts = append(posts, P(-20, z))
	}
	for _, q := range posts {
		g.Fill(BlockAt(q, 0.9, 0.9, F, F+4.9))
	}
	// Railings with spindles, open at the steps.
	rail := func(x0, z0, x1, z1 float64) {
		g.Fill(Block(V(x0, F+1.6, z0), V(x1, F+2.2, z1)))
		g.With(Wood, func() {
			g.Fill(Func{Box{V(x0, F, z0), V(x1, F+1.6, z1)}, func(p Vec3) bool {
				return math.Mod(p.X+p.Z+100, 2.2) < 0.7
			}})
		})
	}
	rail(-20.4, 2.6, -7.4, 3.4)
	rail(-1.6, 2.6, 5.9, 3.4)
	rail(-20.4, -15.9, -19.6, 3.4)
	// Frieze, brackets and the gently sloped porch roof.
	g.Fill(Block(V(-20.5, F+4.2, 2.6), V(6, F+4.9, 3.4)), Block(V(-20.5, F+4.2, -16), V(-19.6, F+4.9, 3.4)))
	g.Use(Slate).Fill(
		Func{Box{V(-20.9, F+4.8, -4), V(6, F+6, 3.9)}, func(p Vec3) bool {
			top := F + 5.9 - (p.Z+4)*0.12
			return p.Y <= top && p.Y >= top-0.8
		}},
		Func{Box{V(-20.9, F+4.8, -16.4), V(-16, F+6, 3.9)}, func(p Vec3) bool {
			top := F + 5.9 - (-16-p.X)*0.12
			return p.Y <= top && p.Y >= top-0.8
		}},
	)
	// Steps down to the path.
	g.Use(Stone).Fill(Block(V(-7, 8.4, 3.5), V(-2, F-0.5, 4.7)), Block(V(-7, 8.4, 4.7), V(-2, F-1.0, 5.9)))

	// Square tower over the front door, with its own mansard and cresting.
	const towerTop, towerRoof, towerPitch = F + 27, 6.0, 1.9
	g.Use(Navy).Fill(Block(V(-8, F, -6), V(-1, towerTop, 1)))
	g.Use(Stone).Fill(
		Block(V(-8.35, F+5.6, -6), V(-0.65, F+6.4, 1.35)),
		Block(V(-8.35, F+11.6, -6.35), V(-0.65, F+12.4, 1.35)),
		Block(V(-8.35, F+17.6, -6.35), V(-0.65, F+18.4, 1.35)),
		Block(V(-8.7, towerTop-0.8, -6.7), V(-0.3, towerTop, 1.7)),
		Func{Box{V(-8.6, towerTop-1.6, -6.6), V(-0.4, towerTop-0.8, 1.6)}, func(p Vec3) bool {
			return math.Mod(p.X+p.Z+100, 2.0) < 0.8
		}},
	)
	tower := m.mansard(-8.8, -0.2, -6.8, 1.8, towerTop, towerRoof, 1.5, towerPitch)
	g.Use(Charcoal).Fill(tower)
	g.Paint(Slate, Func{Box{V(-9, towerTop+2.2, -7), V(0, towerTop+3.1, 2)}, func(p Vec3) bool { return true }})
	room(Block(V(-8+t, F, -5.5), V(-1-t, towerTop, 1-t)))
	room(Func{tower.Box, func(p Vec3) bool {
		q := V(p.X, p.Y+1.0, p.Z)
		return m.mansard(-8.8+t, -0.2-t, -6.8+t, 1.8-t, towerTop, towerRoof, 1.5, towerPitch).F(q)
	}})
	m.cresting(g, -7.3, -1.7, -5.3, 0.3, towerTop+towerRoof)
	// Round attic windows in the tower mansard.
	for _, o := range []Vec3{V(-4.5, towerTop+3.8, 1.0), V(-4.5, towerTop+3.8, -6.0), V(-8.0, towerTop+3.8, -2.5), V(-1.0, towerTop+3.8, -2.5)} {
		g.Carve(Sphere(o, 1.0))
	}
	// Crooked weathervane.
	apex := V(-4.5, towerTop+towerRoof+2.8*towerPitch, -2.5)
	g.Use(Charcoal).Fill(
		Capsule{apex.Sub(V(0, 0.8, 0)), apex.Add(V(0.6, 3.7, 0.3)), 0.45},
		Capsule{apex.Add(V(0.3, 1.7, -1.3)), apex.Add(V(0.35, 1.6, 1.5)), 0.4},
		Capsule{apex.Add(V(-1.8, 3.9, 0.1)), apex.Add(V(2.4, 3.5, 0.5)), 0.4},
		Block(apex.Add(V(2.0, 3.0, 0.0)), apex.Add(V(2.9, 4.2, 0.9))),
		Block(apex.Add(V(-2.3, 3.6, -0.2)), apex.Add(V(-1.5, 5.0, 0.5))),
	)
	g.Use(Gold).Fill(Sphere(apex.Add(V(0.3, 2.5, 0.15)), 0.6))

	// Round turret with a bent witch-hat roof.
	const turretTop, hatHeight = F + 16.5, 16.0
	tc := P(20.5, 0.5)
	lean := P(-2.4, 1.6)
	g.Use(Navy).Fill(Cylinder{Center: tc, Radius: 3.3, Y0: F, Y1: turretTop})
	g.Use(Stone).Fill(
		Cylinder{Center: tc, Radius: 3.65, Y0: F + 5.6, Y1: F + 6.4},
		Cylinder{Center: tc, Radius: 3.65, Y0: F + 11.6, Y1: F + 12.4},
		Cylinder{Center: tc, Radius: 4.0, Y0: turretTop - 0.8, Y1: turretTop},
	)
	room(Cylinder{Center: tc, Radius: 3.3 - t, Y0: F, Y1: turretTop})
	hat := func(r0, shrink float64) Func {
		return Func{Box{V(tc.X-r0-3, turretTop, tc.Y-r0-3), V(tc.X+r0+3, turretTop+hatHeight, tc.Y+r0+3)}, func(p Vec3) bool {
			s := (p.Y - turretTop) / hatHeight
			c := tc.Add(lean.Scale(math.Pow(s, 2.5)))
			r := r0*math.Pow(1-s, 1.3) - shrink
			return r > 0 && p.Plan().Sub(c).Length() <= math.Max(r, 0.5)
		}}
	}
	g.Use(Slate).FillVolume(hat(4.6, 0))
	g.Paint(Charcoal, Func{Box{V(15, turretTop+2.6, -5), V(26, turretTop+3.4, 6)}, func(p Vec3) bool { return true }})
	room(Func{hat(4.6, 0).Box, func(p Vec3) bool { return hat(4.6, 1.3).F(V(p.X, p.Y+0.8, p.Z)) }})
	tip := tc.Add(lean.Scale(math.Pow(0.97, 2.5)))
	g.Use(Charcoal).Fill(Capsule{tc.Add(lean.Scale(math.Pow(0.85, 2.5))).At(turretTop + 0.85*hatHeight), tip.Add(P(-0.6, 0.3)).At(turretTop + hatHeight + 0.6), 0.45})
	_ = tip

	// Chimneys.
	m.chimney(g, -17.6, -12, 1.8, 2.4, 7, mainTop+10.5)
	m.chimney(g, 0.8, -13, 2.2, 2.2, F+8, mainTop+9.5)
	m.chimney(g, 12.0, -10, 2.2, 2.2, F+8, wingTop+12.5)
	m.chimney(g, -16.6, -23.5, 1.6, 2.2, 7, ellTop+9)
	m.chimney(g, -12, -15.5, 2.0, 2.0, F+8, mainTop+8)

	// Dormers on the main mansard: outward is local +X.
	dormer := func(origin Vec2, angle float64) {
		place := func(s Solid) Solid { return Translate(RotateY(s, P(0, 0), angle), origin.At(0)) }
		E := mainTop
		g.Use(DarkStone).Fill(place(Block(V(-4.6, E, -1.5), V(0.4, E+4, 1.5))))
		g.Use(Slate).Fill(place(Func{Box{V(-4.6, E+3.8, -2.1), V(0.9, E+6.6, 2.1)}, func(p Vec3) bool {
			return p.Y >= E+3.8 && math.Abs(p.Z) <= 2.1-(p.Y-E-3.8)/1.3
		}}))
		room(place(Block(V(-5.6, E, -0.8), V(-0.8, E+3.3, 0.8))))
		win(origin.Add(P(0.4, 0).Rotate(angle)), angle, E+0.7, 0.75, 1.7, -1)
	}
	for _, x := range []float64{-12.5, 2.5} {
		dormer(P(x, -3.7), front)
	}
	for _, x := range []float64{-10.5, -3.5, 2.5} {
		dormer(P(x, -18.3), back)
	}
	dormer(P(-16.3, -7.0), west)
	dormer(P(-16.3, -15.0), west)

	// Windows: ground floor taller, upper floor arched.
	for i, x := range []float64{-13.5, -10.5, 1.5, 4.3} {
		win(P(x, -4), front, F+1.0, 1.0, 2.6, []int{0, 2, 0, 1}[i])
		win(P(x, -4), front, F+7.2, 1.0, 2.6, -1)
	}
	for _, z := range []float64{-6.5, -14.5} {
		win(P(-16, z), west, F+1.0, 1.0, 2.6, -1)
		win(P(-16, z), west, F+7.2, 1.0, 2.6, -1)
	}
	for _, x := range []float64{0, 4} {
		win(P(x, -18), back, F+1.0, 1.0, 2.6, -1)
		win(P(x, -18), back, F+7.2, 1.0, 2.6, -1)
	}
	for _, x := range []float64{-12, -8} {
		win(P(x, -18), back, F+7.8, 0.9, 2.2, -1)
	}
	// Wing.
	for _, z := range []float64{-4.5, -9, -13.5} {
		win(P(20, z), east, F+1.0, 1.0, 2.6, -1)
		win(P(20, z), east, F+7.0, 1.0, 2.4, -1)
	}
	for _, x := range []float64{9, 13, 17} {
		win(P(x, -16), back, F+1.0, 1.0, 2.6, -1)
		win(P(x, -16), back, F+7.0, 1.0, 2.4, -1)
	}
	win(P(13, -16), back, wingTop+2.0, 0.9, 1.8, 2)
	win(P(13, 0), front, wingTop+1.0, 0.9, 1.8, 2)
	win(P(13, 2.6), front, F+1.2, 1.1, 2.4, 2)
	win(P(13, 2.6), front, F+6.9, 1.1, 1.8, 1)
	for _, side := range []struct {
		c Vec2
		a float64
	}{{P(10.15, 1.3), 2.645}, {P(15.85, 1.3), 0.497}} {
		win(side.c, side.a, F+1.2, 0.6, 2.4, 0)
		win(side.c, side.a, F+6.9, 0.6, 1.8, -1)
	}
	// Ell.
	for _, x := range []float64{-12, -8} {
		win(P(x, -27), back, F+1.0, 0.9, 2.4, -1)
	}
	win(P(-5, -22.5), east, F+1.0, 0.9, 2.4, -1)
	// Tower: door, then windows on each storey.
	g.Carve(Arch{V(-4.5, F, 0.7), 1.3, 2.8, 2.6, AxisZ})
	g.Use(Wood).Fill(Block(V(-5.8, F, -0.5), V(-4.6, F+3.8, -0.1)))
	g.Use(Stone).Fill(Subtract(Arch{V(-4.5, F, 1.2), 1.9, 2.8, 0.6, AxisZ}, Arch{V(-4.5, F, 1.2), 1.3, 2.8, 1, AxisZ}))
	win(P(-4.5, 1), front, F+7.2, 1.0, 2.6, 2)
	win(P(-4.5, 1), front, F+13.2, 1.0, 2.6, 0)
	win(P(-5.8, 1), front, F+18.8, 0.7, 2.0, 1)
	win(P(-3.2, 1), front, F+18.8, 0.7, 2.0, 0)
	for _, s := range []struct {
		o Vec2
		a float64
	}{{P(-8, -1.5), west}, {P(-1, -1.5), east}} {
		win(s.o, s.a, F+7.2, 0.9, 2.4, -1)
		win(s.o.Add(P(0, -1)), s.a, F+13.2, 0.9, 2.4, -1)
		win(s.o.Add(P(0, -1)), s.a, F+18.8, 0.9, 2.2, -1)
	}
	win(P(-4.5, -6), back, F+18.8, 0.9, 2.2, 0)
	win(P(-4.5, 1), front, F+23.0, 1.0, 2.0, 2)
	win(P(-8, -2.5), west, F+23.0, 1.0, 2.0, 0)
	win(P(-1, -2.5), east, F+23.0, 1.0, 2.0, 1)
	win(P(-4.5, -6), back, F+23.0, 1.0, 2.0, 0)
	// Turret windows.
	for i, a := range []float64{-0.6, 0.45, 1.5, 2.4} {
		for j, sill := range []float64{F + 1.2, F + 7.2, F + 13.0} {
			o := Polar(tc, 3.3, a)
			k := -1
			if i == 2 && j == 2 {
				k = 2
			}
			win(o, a, sill, 0.75, 2.2, k)
		}
	}

	// Dark interiors behind every opening.
	inRoom := func(p Vec3) bool {
		for _, s := range rooms {
			if s.Bounds().Contains(p) && s.Contains(p) {
				return true
			}
		}
		return false
	}
	region := hauntedManorWorldBox(Box{V(-22, F-1, -30), V(26, 50, 6)})
	m.recolor(grid, region, func(c Cell, p Vec3, col Color) (Color, bool) {
		if col == Gold || col == Wood {
			return col, false
		}
		for _, n := range Neighbors {
			nc := c.Add(n)
			if !grid.Has(nc) && inRoom(hauntedManorLocal(grid.World(nc))) {
				return Charcoal, true
			}
		}
		return col, false
	})
	// Peeling paint and creeping ivy on the siding.
	m.recolor(grid, region, func(c Cell, p Vec3, col Color) (Color, bool) {
		p = hauntedManorLocal(p)
		if col != DarkStone && col != Navy {
			return col, false
		}
		peel := math.Sin(p.X*0.9+p.Y*0.35) * math.Sin(p.Z*0.8-p.Y*0.5) * math.Sin(p.Y*0.7+p.X*0.3)
		ivy := math.Sin(p.X*0.45+p.Z*0.3) * math.Sin(p.Y*0.35+p.Z*0.4+p.X*0.2)
		switch {
		case ivy > 0.45 && p.Y < F+4+9*(ivy-0.45):
			return Foliage, true
		case peel > 0.42 && col == Navy:
			return Slate, true
		case peel > 0.42:
			return Stone, true
		}
		return col, false
	})
	// Crows on the ridges.
	for _, q := range []Vec3{V(13, wingTop+9.7, -6), V(13.2, wingTop+9.7, -10.5), V(-9, mainTop+7.9, -11), V(-1, mainTop+6.9, -7)} {
		g.Use(Charcoal).Fill(Ellipsoid{q, V(0.75, 0.55, 0.5)}, Sphere(q.Add(V(0.6, 0.5, 0)), 0.4))
	}
}

// pumpkin is a ribbed gourd; faced pumpkins get a dark carved face toward +Z.
func (HauntedManor) pumpkin(g *Grid, c Vec3, r float64, col Color, face bool) {
	g.Use(col).FillVolume(Ellipsoid{c, V(r*1.15, r*0.85, r)})
	g.Use(Foliage).Fill(Block(c.Add(V(-0.3, r*0.7, -0.3)), c.Add(V(0.3, r*0.85+0.6, 0.3))))
	if face {
		g.Paint(Charcoal, Func{Box{c.Add(V(-r, -r, 0)), c.Add(V(r, r, r+0.5))}, func(p Vec3) bool {
			d := p.Sub(c)
			if d.Z < r*0.45 {
				return false
			}
			eye := d.Y > 0.05 && d.Y < r*0.55 && math.Abs(math.Abs(d.X)-r*0.45) < 0.32
			mouth := d.Y < -0.1 && d.Y > -r*0.6 && math.Abs(d.X) < r*0.6
			return eye || mouth
		}})
	}
}

func (m HauntedManor) garden(g *Grid) {
	// Overgrown shrubs along the porch and the yard, avoiding the path.
	for i := 0; i < 40; i++ {
		x := -28 + 58*Hash01(i, 31)
		z := 9 + 5*Hash01(i, 32)
		if i%3 == 0 {
			x = -33 + 5*Hash01(i, 33)
			z = -26 + 32*Hash01(i, 34)
		}
		if hauntedManorPathDist(hauntedManorPath, P(x, z)) < 2.8 || (x > 9.5 && x < 23 && z > 8) || math.Hypot((x-1)/36, (z+11)/24) > 1.08 {
			continue
		}
		r := 1.0 + 0.9*Hash01(i, 35)
		y := m.ground(x, z)
		col := Foliage
		if Hash01(i, 36) < 0.3 {
			col = Grass
		}
		g.Use(col).FillVolume(Ellipsoid{V(x, y+r*0.5, z), V(r*1.2, r*0.9, r)})
	}
	// Brambles scattered over the slopes.
	for i := 0; i < 60; i++ {
		a := Hash01(i, 37) * 2 * math.Pi
		rr := 0.45 + 0.5*Hash01(i, 38)
		x, z := math.Sin(a)*50*rr, -4+math.Cos(a)*44*rr
		if math.Hypot((x-1)/36, (z+11)/24) < 1.05 || m.graveSD(x, z) < 1.5 || m.pondNorm(x, z) < 1.3 ||
			hauntedManorPathDist(hauntedManorPath, P(x, z)) < 2.5 || hauntedManorPathDist(hauntedManorSidePath, P(x, z)) < 2 {
			continue
		}
		r := 0.8 + 0.8*Hash01(i, 39)
		col := Foliage
		switch h := Hash01(i, 40); {
		case h < 0.25:
			col = Wood
		case h < 0.45:
			col = Grass
		}
		g.Use(col).FillVolume(Ellipsoid{V(x, m.ground(x, z)+r*0.3, z), V(r*1.3, r*0.8, r)})
	}
	// Pumpkin patch in front of the wing.
	g.Use(Wood).Fill(Func{Box{V(10, 5, 8), V(22, 9.2, 13.5)}, func(p Vec3) bool {
		return p.Y <= m.ground(p.X, p.Z) && p.Y >= m.ground(p.X, p.Z)-0.6
	}})
	k := 0
	for x := 11.3; x < 21.5; x += 2.7 {
		for z := 9.2; z < 13.2; z += 2.4 {
			k++
			jx, jz := Hash01(k, 41)-0.5, Hash01(k, 42)-0.5
			p := P(x+jx, z+jz)
			r := 0.75 + 0.45*Hash01(k, 43)
			col := Terracotta
			if Hash01(k, 44) < 0.25 {
				col = Gold
			}
			m.pumpkin(g, p.At(m.ground(p.X, p.Y)+r*0.7), r, col, Hash01(k, 45) < 0.35)
			g.Use(Foliage).Fill(Capsule{p.At(m.ground(p.X, p.Y) + 0.2), P(x+1.3, z+0.9).At(m.ground(x+1.3, z+0.9) + 0.2), 0.45})
		}
	}
	// Jack-o'-lanterns flanking the steps and on the porch.
	deck := hauntedManorWorld(V(0, hauntedManorFloor, 0)).Y
	for i, q := range []Vec3{V(-8.2, 0, 5.4), V(-0.8, 0, 5.6)} {
		w := hauntedManorWorld(q)
		r := 1.4 - 0.1*float64(i)
		m.pumpkin(g, V(w.X, m.ground(w.X, w.Z)+r*0.75, w.Z), r, Terracotta, true)
	}
	for i, q := range []Vec3{V(-10.5, 0, 2.3), V(0.6, 0, 2.4), V(-15, 0, 2.5)} {
		w := hauntedManorWorld(q)
		r := 1.1 - 0.1*float64(i)
		col := Terracotta
		if i == 0 {
			col = Gold
		}
		m.pumpkin(g, V(w.X, deck+r*0.75, w.Z), r, col, true)
	}
}

func (m HauntedManor) graveyard(g *Grid) {
	lg := m.graveLevel()
	cx, cz := hauntedManorGrave.X, hauntedManorGrave.Y
	x0, x1, z0, z1 := cx-8.8, cx+8.8, cz-6.8, cz+6.8
	gz := cz
	// Wrought-iron fence with spear tips, a gate on the west side toward the path.
	g.Use(Charcoal).Fill(Func{Box{V(x0-0.4, lg-0.5, z0-0.4), V(x1+0.4, lg+3.2, z1+0.4)}, func(p Vec3) bool {
		onX := math.Abs(p.X-x0) <= 0.4 || math.Abs(p.X-x1) <= 0.4
		onZ := math.Abs(p.Z-z0) <= 0.4 || math.Abs(p.Z-z1) <= 0.4
		if !(onX && p.Z >= z0-0.4 && p.Z <= z1+0.4) && !(onZ && p.X >= x0-0.4 && p.X <= x1+0.4) {
			return false
		}
		if math.Abs(p.X-x0) <= 0.4 && p.Z > gz-1.4 && p.Z < gz+1.4 {
			return false
		}
		if (p.Y > lg+0.5 && p.Y < lg+1.0) || (p.Y > lg+2.2 && p.Y < lg+2.7) {
			return true
		}
		along := p.X + p.Z
		bar := math.Mod(along+100, 2.1) < 0.8
		return bar && p.Y <= lg+2.9+0.4*math.Cos(along)
	}})
	for _, q := range []Vec2{P(x0, z0), P(x0, z1), P(x1, z0), P(x1, z1), P(x0, gz-1.6), P(x0, gz+1.6)} {
		g.Fill(BlockAt(q, 1.2, 1.2, lg-0.5, lg+3.4))
		g.Fill(Sphere(q.At(lg+3.8), 0.6))
	}
	// An iron gate leaf swung open.
	g.Fill(Func{Box{V(x0-2.6, lg-0.2, gz-1.6), V(x0+0.4, lg+3, gz-0.6)}, func(p Vec3) bool {
		return p.Y < lg+0.6 || p.Y > lg+2.3 || math.Mod(p.X+100, 1.3) < 0.7
	}})

	// Mausoleum at the back of the plot.
	mx0, mx1, mz0, mz1 := cx-3.2, cx+3.2, cz-6.1, cz-1.4
	g.Use(Stone).Fill(Block(V(mx0-0.4, lg-0.5, mz0-0.4), V(mx1+0.4, lg+0.6, mz1+2.4)))
	g.Use(Stone).Fill(Block(V(mx0, lg, mz0), V(mx1, lg+5, mz1)))
	g.Carve(Block(V(mx0+1, lg+0.6, mz0+1), V(mx1-1, lg+4.4, mz1-1)))
	g.Use(DarkStone).Fill(m.gable(mx0-0.6, mx1+0.6, mz0-0.6, mz1+2.6, lg+5, 0.9, AxisZ))
	g.Use(Stone).Fill(Intersect(m.gable(mx0+0.4, mx1-0.4, mz0-1, mz1+2.7, lg+5, 0.9, AxisZ), Block(V(mx0, lg+5, mz1+1.6), V(mx1, lg+9, mz1+2.6))))
	for _, x := range []float64{mx0 + 0.9, mx1 - 0.9} {
		g.Fill(Cylinder{Center: P(x, mz1+1.6), Radius: 0.6, Y0: lg, Y1: lg + 5})
		g.Fill(BlockAt(P(x, mz1+1.6), 1.6, 1.6, lg+0.5, lg+1.0), BlockAt(P(x, mz1+1.6), 1.6, 1.6, lg+4.4, lg+5.0))
	}
	g.Carve(Arch{V(cx, lg+0.6, mz1), 1.1, 2.0, 2.6, AxisZ})
	g.Paint(Charcoal, Block(V(mx0+0.6, lg, mz0+0.6), V(mx1-0.6, lg+4.8, mz1-0.6)))
	g.Use(Stone).Fill(Block(V(cx-0.4, lg+7.7, mz1+1.6), V(cx+0.4, lg+10.4, mz1+2.4)), Block(V(cx-1.2, lg+9.0, mz1+1.6), V(cx+1.2, lg+9.6, mz1+2.4)))

	// Headstones, crosses and obelisks with a little tilt.
	k := 0
	stone := func(x, z float64) {
		k++
		h := Hash01(k, 51)
		col := Stone
		if Hash01(k, 52) < 0.4 {
			col = DarkStone
		}
		base := m.ground(x, z) - 0.4
		var s Solid
		switch {
		case h < 0.45:
			s = Union(Block(V(x-0.8, base, z-0.45), V(x+0.8, base+2.0, z+0.45)),
				Arch{V(x, base+1.9, z), 0.8, 0, 0.9, AxisZ})
		case h < 0.75:
			s = Union(Block(V(x-0.45, base, z-0.45), V(x+0.45, base+3.2, z+0.45)),
				Block(V(x-1.2, base+2.0, z-0.45), V(x+1.2, base+2.7, z+0.45)))
		case h < 0.9:
			s = Union(Block(V(x-0.7, base, z-0.7), V(x+0.7, base+0.8, z+0.7)),
				Frustum{P(x, z), 0.65, 0.35, base + 0.8, base + 3.4}, Cone(P(x, z), 0.45, base+3.4, 0.9))
		default:
			s = Union(Block(V(x-0.9, base, z-0.5), V(x+0.9, base+1.4, z+0.5)),
				Arch{V(x, base+1.3, z), 0.9, 0, 1.0, AxisZ})
		}
		lean := P((Hash01(k, 53)-0.5)*0.35, (Hash01(k, 54)-0.5)*0.3)
		g.Use(col).Fill(Sheared{s, lean, base})
		if Hash01(k, 55) < 0.5 {
			mound := Wood
			if Hash01(k, 56) < 0.5 {
				mound = Grass
			}
			g.Use(mound).FillVolume(Ellipsoid{V(x, m.ground(x, z+1.8), z+1.9), V(0.9, 0.55, 1.3)})
		}
	}
	for _, z := range []float64{cz + 2.6, cz + 5.3} {
		for x := x0 + 1.8; x < x1-1; x += 2.45 {
			stone(x+0.3*(Hash01(int(x*10), int(z))-0.5), z)
		}
	}
	for _, z := range []float64{cz - 4.7, cz - 2.0} {
		for _, x := range []float64{cx - 7.2, cx - 5.0, cx + 5.2, cx + 7.4} {
			stone(x, z)
		}
	}
	// Crows on fence posts.
	for _, q := range []Vec2{P(x1, z1), P(x0, z0)} {
		c := q.At(lg + 4.7)
		g.Use(Charcoal).Fill(Ellipsoid{c, V(0.7, 0.5, 0.5)}, Sphere(c.Add(V(0.5, 0.5, 0)), 0.4))
	}
}

func (m HauntedManor) pond(g *Grid) {
	lp := m.pondLevel()
	box := Box{V(hauntedManorPond.X-10, lp-2, hauntedManorPond.Y-7), V(hauntedManorPond.X+10, lp+1, hauntedManorPond.Y+7)}
	m.recolor(g, box, func(c Cell, p Vec3, col Color) (Color, bool) {
		n := m.pondNorm(p.X, p.Z)
		if n > 1 {
			return col, false
		}
		if Hash01(int(math.Floor(p.X/1.2)), int(math.Floor(p.Z/1.2)), 61) < 0.12 && n < 0.85 {
			return Foliage, true
		}
		return Water, true
	})
	// Reeds and stumps around the edge.
	for i := 0; i < 26; i++ {
		a := Hash01(i, 62) * 2 * math.Pi
		r := 1.0 + 0.15*Hash01(i, 63)
		x := hauntedManorPond.X + math.Cos(a)*8.5*r
		z := hauntedManorPond.Y + math.Sin(a)*5.8*r
		y := m.ground(x, z)
		h := 1.2 + 1.4*Hash01(i, 64)
		col := Foliage
		if i%4 == 0 {
			col = Wood
		}
		g.Use(col).Fill(BlockAt(P(x, z), 0.8, 0.8, y, y+h))
	}
}

// boundary is the crumbling stone wall around the foot of the hill with the iron gate.
func (m HauntedManor) boundary(g *Grid) {
	const ra, rb, cz = 50 * 0.935, 44 * 0.935, -4.0
	box := Box{V(-ra-1, 0, cz-rb-1), V(ra+1, 5, cz+rb+1)}
	wall := func(p Vec3) (bool, float64) {
		u, v := p.X/ra, (p.Z-cz)/rb
		e := math.Hypot(u, v)
		if math.Abs(e-1)*45 > 0.65 {
			return false, 0
		}
		phi := math.Atan2(u, v)
		if math.Abs(phi) > 2.2 || (math.Abs(p.X) < 2.0 && p.Z > 0) {
			return false, 0
		}
		seg := int(math.Floor(phi * 45 / 1.6))
		top := 2.0 + 0.8*Hash01(seg, 81)
		if Hash01(seg, 82) < 0.18 {
			top = 0.6 + 0.4*Hash01(seg, 83)
		}
		ground := m.ground(p.X, p.Z)
		return p.Y >= ground-0.8 && p.Y <= ground+top, phi
	}
	g.Use(DarkStone).Fill(Func{box, func(p Vec3) bool { ok, _ := wall(p); return ok }})
	m.recolor(g, box, func(c Cell, p Vec3, col Color) (Color, bool) {
		ok, phi := wall(p)
		if !ok || col != DarkStone {
			return col, false
		}
		if Hash01(int(math.Floor(phi*45/1.4+math.Floor(p.Y/0.8)*0.5)), int(math.Floor(p.Y/0.8)), 84) < 0.3 {
			return Stone, true
		}
		if Hash01(int(math.Floor(phi*45/1.1)), int(math.Floor(p.Y/0.8)), 85) < 0.1 {
			return Foliage, true
		}
		return col, false
	})
	// Gate pillars with lanterns and iron gates, one ajar.
	gz := cz + rb
	for _, side := range []float64{-1, 1} {
		c := P(side*2.9, gz)
		y := m.ground(c.X, c.Y)
		g.Use(Stone).Fill(BlockAt(c, 1.9, 1.9, y-0.8, y+4.4), BlockAt(c, 2.5, 2.5, y+4.4, y+5.0))
		g.Use(DarkStone).Fill(BlockAt(c, 2.3, 2.3, y-0.8, y+0.6))
		g.Use(Gold).Fill(BlockAt(c, 0.8, 0.8, y+5.0, y+6.0))
		g.Use(Charcoal).Fill(BlockAt(c, 1.2, 1.2, y+6.0, y+6.5))
	}
	y := m.ground(0, gz)
	g.Use(Charcoal).Fill(Func{Box{V(-2.0, y-0.3, gz-0.4), V(0, y+3.6, gz+0.4)}, func(p Vec3) bool {
		top := y + 3.0 + 0.5*math.Sin((p.X+2)/2*math.Pi)
		return p.Y <= top && (p.Y < y+0.5 || (p.Y > y+2.2 && p.Y < y+2.8) || math.Mod(p.X+100, 1.2) < 0.6)
	}})
	g.Fill(RotateY(Func{Box{V(0, y-0.3, gz-0.4), V(2.0, y+3.6, gz+0.4)}, func(p Vec3) bool {
		top := y + 3.0 + 0.5*math.Sin(p.X/2*math.Pi)
		return p.Y <= top && (p.Y < y+0.5 || (p.Y > y+2.2 && p.Y < y+2.8) || math.Mod(p.X+100, 1.2) < 0.6)
	}}, P(2.0, gz), 1.1))
}

// spirits adds a few ghosts drifting over the grounds and bats around the tower.
func (m HauntedManor) spirits(g *Grid) {
	ghost := func(c Vec3, r float64, facing float64) {
		g.Use(White).FillVolume(Func{Box{c.Sub(V(r+1, 2.6*r, r+1)), c.Add(V(r+1, r+0.2, r+1))}, func(p Vec3) bool {
			d := p.Sub(c)
			if d.Y > 0 {
				return d.Length() <= r
			}
			// A tapering, wavy sheet below the head.
			k := 1 + d.Y/(2.6*r)
			wave := 0.4 * math.Sin(math.Atan2(d.Z, d.X)*4)
			return math.Hypot(d.X, d.Z) <= r*(0.6+0.4*k) && d.Y >= -2.5*r+wave
		}})
		dir := P(math.Cos(facing), math.Sin(facing))
		side := P(-dir.Y, dir.X)
		for _, s := range []float64{-0.42, 0.42} {
			e := c.Plan().Add(dir.Scale(r * 0.8)).Add(side.Scale(s * r))
			g.Use(Charcoal).Fill(Sphere(e.At(c.Y+0.2*r), 0.38))
		}
		mouth := c.Plan().Add(dir.Scale(r * 0.85))
		g.Use(Charcoal).Fill(Sphere(mouth.At(c.Y-0.45*r), 0.4))
	}
	lg := m.graveLevel()
	ghost(V(hauntedManorGrave.X-1, lg+8, hauntedManorGrave.Y+3.5), 1.25, 2.2)
	ghost(V(-17, 23, 9.5), 1.15, 1.2)
	top := hauntedManorWorld(V(-4.5, hauntedManorFloor+27, -2.5))
	for i := 0; i < 6; i++ {
		a := float64(i)*1.13 + 0.4
		c := Polar(top.Plan(), 9+3*Hash01(i, 111), a).At(top.Y + 2 + 9*Hash01(i, 112))
		w := P(math.Cos(a+1.6), math.Sin(a+1.6)).Scale(1.6)
		g.Use(Charcoal).Fill(
			Capsule{c, c.Add(w.At(0.7)), 0.45},
			Capsule{c, c.Sub(w.At(-0.7)), 0.45},
		)
	}
}

// trees grows gnarled leafless trees.
func (m HauntedManor) trees(g *Grid) {
	var branch func(a, dir Vec3, length, radius float64, depth, seed int)
	branch = func(a, dir Vec3, length, radius float64, depth, seed int) {
		jit := V(Hash01(seed, 91)-0.5, Hash01(seed, 92)-0.5, Hash01(seed, 93)-0.5).Scale(length * 0.35)
		b := a.Add(dir.Scale(length))
		mid := a.Add(b).Scale(0.5).Add(jit)
		col := Charcoal
		if depth >= 2 {
			col = Wood
		}
		g.Use(col).Fill(Capsule{a, mid, radius}, Capsule{mid, b, math.Max(radius*0.85, 0.5)})
		if depth == 0 {
			return
		}
		n := 2
		if Hash01(seed, 94) < 0.25 && depth >= 2 {
			n = 3
		}
		turn := Hash01(seed, 95) * 2 * math.Pi
		for i := 0; i < n; i++ {
			az := turn + float64(i)*2*math.Pi/float64(n)
			spread := 0.9 + 0.6*Hash01(seed, i, 96)
			d := dir.Add(V(math.Cos(az)*spread, -0.05+0.3*Hash01(seed, i, 97), math.Sin(az)*spread))
			d = d.Scale(1 / d.Length())
			branch(b, d, length*(0.68+0.15*Hash01(seed, i, 98)), math.Max(radius*0.72, 0.5), depth-1, seed*7+i+1)
		}
	}
	tree := func(x, z, height float64, seed int) {
		y := m.ground(x, z)
		g.Use(Wood).Fill(Frustum{P(x, z), 1.9, 1.0, y - 0.5, y + 1.2})
		lean := V(Hash01(seed, 99)-0.5, 2.5, Hash01(seed, 100)-0.5)
		branch(V(x, y, z), lean.Scale(1/lean.Length()), height*0.36, 1.1, 3, seed)
	}
	tree(hauntedManorGrave.X-6.8, hauntedManorGrave.Y-4.2, 15, 3)
	tree(-37, -8, 14, 5)
	tree(-30, -30, 13, 7)
	tree(37, -14, 14, 9)
	tree(30, -32, 12, 11)
	tree(-14, 19, 12, 13)
	tree(10, 32, 11, 15)
	tree(-8, -41, 12, 17)
	tree(-41, 12, 11, 19)
	tree(20, -38, 12, 21)
	tree(-21, -40, 11, 23)
	tree(41, 7, 11, 25)
	tree(-45, -10, 10, 27)
	tree(37.5, 20, 9, 29)
}
