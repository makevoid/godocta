package scenes

import (
	"math"

	"godocta/gen/kit"
	. "godocta/gen/octa"
)

// WizardAcademy is a romantic Gothic castle school on a cliff above a dark lake: a
// great hall with tall glowing lancets and flying buttresses over the cliff edge, ranges
// around a courtyard, many slender cone-roofed towers led by a tall astronomy tower, a
// covered wooden bridge across a ravine to a clock tower and greenhouse, a boathouse
// and lantern boats on the lake, a pine forest and gamekeeper's hut in the western glen,
// and a two-tier railway viaduct with a little red steam train running into a tunnel
// under the castle. The lake lies to the front (+Z). The recipe is authored with the
// glen to the west (-X), then mirrored so the glen, forest and train face the hero view.
type WizardAcademy struct{}

func init() { Register(WizardAcademy{}) }

func (WizardAcademy) Name() string  { return "wizard_academy" }
func (WizardAcademy) Title() string { return "Wizard Academy" }

const (
	wizardAcademyTop  = 18.0 // Castle plateau.
	wizardAcademyRail = 12.0 // Viaduct deck and railway.
)

// wizardAcademyMap is a sampled terrain heightfield with bilinear lookup.
type wizardAcademyMap struct {
	x0, z0, step float64
	nx, nz       int
	h            []float64
}

func (m *wizardAcademyMap) at(x, z float64) float64 {
	fx, fz := (x-m.x0)/m.step, (z-m.z0)/m.step
	if fx < 0 || fz < 0 || fx >= float64(m.nx-1) || fz >= float64(m.nz-1) {
		return -9
	}
	i, j := int(fx), int(fz)
	tx, tz := fx-float64(i), fz-float64(j)
	k := j*m.nx + i
	a := m.h[k] + (m.h[k+1]-m.h[k])*tx
	b := m.h[k+m.nx] + (m.h[k+m.nx+1]-m.h[k+m.nx])*tx
	return a + (b-a)*tz
}

// low is the lowest terrain around a point, used to close cliffs into shells.
func (m *wizardAcademyMap) low(x, z float64) float64 {
	lo := m.at(x, z)
	for k := 0; k < 8; k++ {
		a := float64(k) * math.Pi / 4
		lo = math.Min(lo, m.at(x+0.9*math.Cos(a), z+0.9*math.Sin(a)))
	}
	return lo
}

// noise is smooth value noise in [0, 1).
func (WizardAcademy) noise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

// fbm is three octaves of value noise in [0, 1).
func (w WizardAcademy) fbm(x, z float64, seed int) float64 {
	return (w.noise(x, z, seed)*4 + w.noise(x*2.1, z*2.1, seed+1)*2 + w.noise(x*4.3, z*4.3, seed+2)) / 7
}

// site is the normalized radius of the elliptical plot; land and water stop at 1.
func (WizardAcademy) site(x, z float64) float64 {
	if z < 0 {
		return math.Hypot(x/66, z/42)
	}
	return math.Hypot(x/66, z/40)
}

// lake is the normalized radius of the lake basin (below 1 is open water).
func (w WizardAcademy) lake(x, z float64) float64 {
	return math.Hypot(x/54, (z-31)/23) + 0.12*(w.fbm(x/12, z/12, 1)-0.5)
}

// ravine and glen are the plan centerlines of the two valleys.
func (WizardAcademy) ravine(z float64) float64 { return 22.5 + 0.8*math.Sin(z/7) }
func (WizardAcademy) glen(z float64) float64   { return -47 + 1.5*math.Sin(z/8) }

// height is the terrain: rolling forested hills, the lake basin, the castle crag with
// its cliff plunging into the lake, the ravine through the crag and the western glen.
func (w WizardAcademy) height(x, z float64) float64 {
	e := w.site(x, z)
	if e > 1 {
		return -9
	}
	h := 2 + 7*w.fbm(x/16, z/16, 2)
	// West hill carrying the railway.
	h = math.Max(h, 12.8-0.01*((x+62)*(x+62)+0.5*(z+12)*(z+12))+1.5*(w.fbm(x/8, z/8, 3)-0.5))
	// Lake basin with gently rising shores.
	h = math.Min(h, -2+(w.lake(x, z)-1)*28+2)
	// Castle crag: a flat plateau with ledged cliffs.
	dx := math.Max(math.Max(-37-x, x-51), 0)
	front := 10 + 1.2*math.Sin(x/7) + 3.5*w.fbm(x/5, 3, 6)
	dz := math.Max(math.Max(-34-z, z-front), 0)
	d := math.Hypot(dx, dz)
	if d > 0 {
		d += 3 * (w.fbm(x/6, z/6, 4) - 0.5) * math.Min(d, 1)
		d = math.Max(d, 0.01)
	}
	crag := wizardAcademyTop
	if d > 0 {
		ledge := 1.3 * math.Floor(d/2.2) * (0.5 + w.noise(x/7, z/7, 5))
		crag = wizardAcademyTop - (2.6+1.8*w.noise(x/5, z/5, 7))*d - ledge
	}
	h = math.Max(h, crag)
	// Ravine between the castle and the clock court, opening to the lake.
	bottom := 0.5 + math.Max(0, -z)*0.05
	h = math.Min(h, bottom+math.Max(0, math.Abs(x-w.ravine(z))-2.2)*5)
	// Western glen under the viaduct.
	gb := 0.5 + math.Max(0, -z-10)*0.08
	h = math.Min(h, gb+math.Max(0, math.Abs(x-w.glen(z))-2)*2.2)
	// Railway cutting on the west hill.
	if x < -53 && math.Abs(z+7.5) < 2.5 {
		h = math.Min(h, wizardAcademyRail-0.1)
	}
	// Everything slopes down to the plot's rim.
	return math.Min(h, 0.8+100*(1-e))
}

func (w WizardAcademy) Build(out *Grid) Order {
	g := NewGridWithSpan(out.Span())
	defer func() {
		// Mirror east-west so the glen, forest and railway face the hero camera.
		for _, c := range g.Cells() {
			color, _ := g.ColorOf(c)
			out.Use(color).Add(Cell{X: -c.X, Y: c.Y, Z: c.Z})
		}
	}()
	m := &wizardAcademyMap{x0: -67, z0: -51, step: 0.4}
	m.nx, m.nz = int(134/m.step)+2, int(98/m.step)+2
	m.h = make([]float64, m.nx*m.nz)
	for j := 0; j < m.nz; j++ {
		for i := 0; i < m.nx; i++ {
			m.h[j*m.nx+i] = w.height(m.x0+float64(i)*m.step, m.z0+float64(j)*m.step)
		}
	}
	w.terrain(g, m)
	w.viaduct(g, m)
	w.ranges(g)
	w.greatHall(g)
	w.towers(g)
	w.astronomy(g)
	w.clockCourt(g)
	w.bridge(g)
	w.lakeside(g)
	w.hut(g, m)
	w.trees(g, m)
	w.train(g)
	return Sweep(P(0, -10), math.Pi/2)
}

// terrain lays the lake layer and fills the heightfield as a hollow rock shell with
// strata, grass and moss on gentle ground, streams and a shingle shore.
func (w WizardAcademy) terrain(g *Grid, m *wizardAcademyMap) {
	box := Box{V(m.x0, -0.2, m.z0), V(m.x0+134, wizardAcademyTop+0.5, m.z0+98)}
	g.Use(DarkStone).Fill(Func{box, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		if h < 0.4 || p.Y > h {
			return false
		}
		return p.Y >= math.Min(h, m.low(p.X, p.Z))-0.45
	}})
	gentle := func(p Vec3, slack float64) bool {
		h := m.at(p.X, p.Z)
		return p.Y >= h-1.5 && h-m.low(p.X, p.Z) < slack
	}
	// Rock strata on the cliffs.
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		band := int(math.Floor(p.Y/1.3 + 2*w.noise(p.X/9, p.Z/9, 11)))
		return Hash01(band, int(math.Floor(p.X/4)), int(math.Floor(p.Z/4)), 12) < 0.35
	}})
	g.Paint(Grass, Func{box, func(p Vec3) bool { return gentle(p, 1.3) }})
	// Dark forest floor and moss on the ledges.
	g.Paint(Foliage, Func{box, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		if !gentle(p, 1.6) || h > wizardAcademyTop-1 {
			return false
		}
		return w.fbm(p.X/9, p.Z/9, 13) > 0.48 || p.X < -40
	}})
	// Shingle shore.
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		return h >= 0.4 && h < 1.4 && p.Y >= h-1 && w.lake(p.X, p.Z) < 1.12
	}})
	// Streams in the ravine and the glen.
	g.Paint(Water, Func{box, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		if p.Y < h-1 || h > 4 || h < 0.4 {
			return false
		}
		return math.Abs(p.X-w.ravine(p.Z)) < 1.6 || math.Abs(p.X-w.glen(p.Z)) < 1.5
	}})
	// The lake: one water layer, laid after the land paints.
	g.Use(Water).Fill(Func{Box{V(m.x0, -0.2, m.z0), V(m.x0+134, 0.2, m.z0+98)}, func(p Vec3) bool {
		return w.site(p.X, p.Z) <= 1 && m.at(p.X, p.Z) < 0.4
	}})
	// Railway on the hill west of the viaduct.
	rail := Func{Box{V(m.x0, 8, -10), V(-54, 13, -5)}, func(p Vec3) bool { return p.Y >= m.at(p.X, p.Z)-0.9 }}
	g.Paint(Wood, Intersect(rail, Block(V(m.x0, 8, -8.8), V(-54, 13, -6.2))))
	g.Paint(Charcoal, Intersect(rail, Union(Block(V(m.x0, 8, -8.4), V(-54, 13, -7.9)), Block(V(m.x0, 8, -7.1), V(-54, 13, -6.6)))))
}

// lancet is a pointed (equilateral) Gothic arch: jambs up to Spring, then two arcs.
func (WizardAcademy) lancet(sill Vec3, half, spring, depth float64, through Axis) Func {
	r := 2 * half
	span := V(half, 0, depth/2)
	if through == AxisX {
		span = V(depth/2, 0, half)
	}
	box := Box{sill.Sub(span), sill.Add(span).Add(V(0, spring+1.8*half, 0))}
	return Func{box, func(p Vec3) bool {
		d := p.Sub(sill)
		across, along := d.X, d.Z
		if through == AxisX {
			across, along = d.Z, d.X
		}
		if math.Abs(along) > depth/2 || math.Abs(across) > half || d.Y < 0 {
			return false
		}
		rise := d.Y - spring
		if rise <= 0 {
			return true
		}
		return (across+half)*(across+half)+rise*rise <= r*r && (across-half)*(across-half)+rise*rise <= r*r
	}}
}

// window is a lancet on a wall's outer face: a shallow recess with a pane behind it.
// out is +1 or -1, the face's outward direction along the axis.
func (w WizardAcademy) window(g *Grid, at Vec3, half, spring float64, through Axis, out float64, pane Color) {
	shift := func(s float64) Vec3 {
		if through == AxisX {
			return at.Add(V(out*s, 0, 0))
		}
		return at.Add(V(0, 0, out*s))
	}
	g.Carve(w.lancet(shift(0.3), half, spring, 1.5, through))
	g.Paint(pane, w.lancet(shift(-1.25), half, spring, 1.6, through))
}

// radialWindow is a lancet in a round wall at a plan angle.
func (w WizardAcademy) radialWindow(g *Grid, c Vec2, r, a, sill, half, spring float64, pane Color) {
	place := func(x, depth float64) Solid {
		return Translate(RotateY(w.lancet(V(x, sill, 0), half, spring, depth, AxisX), P(0, 0), a), c.At(0))
	}
	g.Carve(place(r+0.3, 1.5))
	g.Paint(pane, place(r-1.25, 1.6))
}

// pane picks a lit (Gold) or dark window by hash.
func (WizardAcademy) pane(lit float64, keys ...int) Color {
	if Hash01(keys...) < lit {
		return Gold
	}
	return Charcoal
}

// hollow fills a box with walls of thickness t whose inner faces are dark, so windows
// and openings read as dark interiors.
func (WizardAcademy) hollow(g *Grid, lo, hi Vec3, t float64, wall Color) {
	g.With(Charcoal, func() { g.Fill(Block(lo, hi)) })
	g.With(wall, func() {
		g.Fill(Subtract(Block(lo, hi), Block(lo.Add(V(0.55, -1, 0.55)), hi.Sub(V(0.55, 0.55, 0.55)))))
	})
	g.Carve(Block(lo.Add(V(t, 0.5, t)), hi.Sub(V(t, t, t))))
}

// drum is a hollow round wall with a dark inner face.
func (WizardAcademy) drum(g *Grid, c Vec2, r, t, y0, y1 float64, wall Color) {
	g.With(Charcoal, func() {
		g.Fill(Cylinder{Center: c, Radius: r - 0.55, Inner: r - t, Y0: y0, Y1: y1})
		g.Fill(Cylinder{Center: c, Radius: r - 0.55, Y0: y0, Y1: y0 + 0.5})
	})
	g.With(wall, func() { g.Fill(Cylinder{Center: c, Radius: r, Inner: r - 0.55, Y0: y0, Y1: y1}) })
}

// steep is a pitched roof with the given rise per unit across, ridge along an axis.
func (WizardAcademy) steep(min, max Vec2, base, pitch float64, ridge Axis) Func {
	half := (max.Y - min.Y) / 2
	if ridge == AxisZ {
		half = (max.X - min.X) / 2
	}
	return Func{Box{min.At(base), max.At(base + half*pitch)}, func(p Vec3) bool {
		across := math.Abs(p.Z - (min.Y+max.Y)/2)
		if ridge == AxisZ {
			across = math.Abs(p.X - (min.X+max.X)/2)
		}
		return across <= half-(p.Y-base)/pitch
	}}
}

// roof fills a hollow steep roof shell of horizontal thickness t.
func (w WizardAcademy) roof(g *Grid, min, max Vec2, base, pitch, t float64, ridge Axis, color Color) {
	outer := w.steep(min, max, base, pitch, ridge)
	var inner Func
	if ridge == AxisZ {
		inner = w.steep(P(min.X+t, min.Y-1), P(max.X-t, max.Y+1), base, pitch, ridge)
		inner.Box.Max.Z, inner.Box.Min.Z = max.Y-t, min.Y+t
	} else {
		inner = w.steep(P(min.X-1, min.Y+t), P(max.X+1, max.Y-t), base, pitch, ridge)
		inner.Box.Max.X, inner.Box.Min.X = max.X-t, min.X+t
	}
	g.With(color, func() { g.Fill(Subtract(outer, inner)) })
}

// cone is a slender conical roof, hollow when large, with a flared eave.
func (WizardAcademy) cone(g *Grid, c Vec2, r, base, h float64, color Color) {
	outer := Cone(c, r, base, h)
	g.With(color, func() {
		if r < 2.6 {
			g.FillVolume(outer)
			return
		}
		inner := Cone(c, r-1.3, base-0.5, (r-1.3)*h/r)
		g.FillVolume(Subtract(outer, inner))
	})
}

// finial is a gold spike with a ball.
func (WizardAcademy) finial(g *Grid, at Vec2, y, h float64) {
	g.With(Gold, func() {
		g.Fill(Cylinder{Center: at, Radius: 0.55, Y0: y - 0.8, Y1: y + h})
		g.Fill(Sphere(at.At(y+h*0.35), 0.85))
	})
}

// tower is a round Gothic tower: hollow storeys with lancet windows, string courses, a
// corbelled cornice and a tall slender cone roof.
func (w WizardAcademy) tower(g *Grid, c Vec2, r, base, top, roofH float64, roof Color, seed int) {
	w.drum(g, c, r, 1.0, base, top, Stone)
	g.Paint(DarkStone, Cylinder{Center: c, Radius: r + 0.5, Y0: base, Y1: wizardAcademyTop + 1.2})
	for y := base + 8; y < top-3; y += 7 {
		g.Paint(DarkStone, Cylinder{Center: c, Radius: r + 0.5, Y0: y, Y1: y + 0.4})
	}
	count := 3
	if r > 3 {
		count = 4
	}
	level := 0
	for y := wizardAcademyTop + 3.5; y+3.5 < top-2.5; y += 5 {
		for k := 0; k < count; k++ {
			a := float64(k)*2*math.Pi/float64(count) + float64(level)*0.8 + Hash01(seed, 1)*6
			w.radialWindow(g, c, r, a, y, 0.6, 1.9, w.pane(0.4, seed, level, k))
		}
		level++
	}
	g.With(DarkStone, func() {
		g.Fill(Frustum{c, r, r + 0.8, top - 2.2, top - 0.8})
	})
	g.With(Stone, func() { g.Fill(Cylinder{Center: c, Radius: r + 0.8, Y0: top - 0.8, Y1: top}) })
	w.cone(g, c, r+1.1, top-0.3, roofH, roof)
	w.finial(g, c, top-0.3+roofH, 2)
}

// wing is a hollow range with storeys of lancet windows, a plinth and string courses,
// and a steep roof between gable walls, with chimneys.
func (w WizardAcademy) wing(g *Grid, lo, hi Vec3, ridge Axis, roof Color, seed int) {
	w.hollow(g, lo, hi, 1.0, Stone)
	g.Paint(DarkStone, Block(V(lo.X-1, lo.Y, lo.Z-1), V(hi.X+1, wizardAcademyTop+1, hi.Z+1)))
	g.Paint(DarkStone, Block(V(lo.X-1, hi.Y-0.5, lo.Z-1), V(hi.X+1, hi.Y, hi.Z+1)))
	storey := 0
	for y := wizardAcademyTop + 2; y+3 < hi.Y; y += 4.5 {
		if storey > 0 {
			g.Paint(DarkStone, Block(V(lo.X-1, y-1.2, lo.Z-1), V(hi.X+1, y-0.9, hi.Z+1)))
		}
		k := 0
		for x := lo.X + 2; x <= hi.X-1.8; x += 2.8 {
			w.window(g, V(x, y, lo.Z), 0.6, 1.8, AxisZ, -1, w.pane(0.35, seed, storey, k, 1))
			w.window(g, V(x, y, hi.Z), 0.6, 1.8, AxisZ, 1, w.pane(0.35, seed, storey, k, 2))
			k++
		}
		for z := lo.Z + 2; z <= hi.Z-1.8; z += 2.8 {
			w.window(g, V(lo.X, y, z), 0.6, 1.8, AxisX, -1, w.pane(0.35, seed, storey, k, 3))
			w.window(g, V(hi.X, y, z), 0.6, 1.8, AxisX, 1, w.pane(0.35, seed, storey, k, 4))
			k++
		}
		storey++
	}
	const pitch = 1.6
	min, max := P(lo.X-0.6, lo.Z-0.6), P(hi.X+0.6, hi.Z+0.6)
	if ridge == AxisX {
		half := (hi.Z - lo.Z) / 2
		g.With(Stone, func() {
			for _, gx := range []float64{lo.X, hi.X - 1.2} {
				g.Fill(Intersect(Block(V(gx, hi.Y, lo.Z), V(gx+1.2, hi.Y+half*pitch, hi.Z)), w.steep(P(gx, lo.Z-0.6), P(gx+1.2, hi.Z+0.6), hi.Y, pitch, AxisX)))
			}
		})
		min.X, max.X = lo.X+1.2, hi.X-1.2
	} else {
		half := (hi.X - lo.X) / 2
		g.With(Stone, func() {
			for _, gz := range []float64{lo.Z, hi.Z - 1.2} {
				g.Fill(Intersect(Block(V(lo.X, hi.Y, gz), V(hi.X, hi.Y+half*pitch, gz+1.2)), w.steep(P(lo.X-0.6, gz), P(hi.X+0.6, gz+1.2), hi.Y, pitch, AxisZ)))
			}
		})
		min.Y, max.Y = lo.Z+1.2, hi.Z-1.2
	}
	w.roof(g, min, max, hi.Y-0.5, pitch, 1.5, ridge, roof)
	// Chimney stacks along the ridge.
	length := hi.X - lo.X
	if ridge == AxisZ {
		length = hi.Z - lo.Z
	}
	for t := 5.0; t < length-3; t += 9 {
		at := P(lo.X+t, (lo.Z+hi.Z)/2+1.2)
		if ridge == AxisZ {
			at = P((lo.X+hi.X)/2+1.2, lo.Z+t)
		}
		g.With(DarkStone, func() {
			g.Fill(BlockAt(at, 1.4, 1.4, hi.Y, hi.Y+2.4*pitch+2.5))
		})
	}
}

// ranges builds the west, north and east ranges around the courtyard, with an open
// cloister arcade on the courtyard side and a fountain on the lawn.
func (w WizardAcademy) ranges(g *Grid) {
	base := wizardAcademyTop - 0.5
	w.wing(g, V(-32, base, -32), V(15, 28, -26), AxisX, Slate, 1)
	w.wing(g, V(-32, base, -26), V(-24, 30.5, -1), AxisZ, Slate, 2)
	w.wing(g, V(9, base, -26), V(15, 30, 6), AxisZ, Slate, 3)
	// Cloister arcades on the courtyard faces.
	for x := -21.5; x <= 7; x += 2.6 {
		g.Carve(w.lancet(V(x, wizardAcademyTop, -26), 0.9, 1.8, 3, AxisZ))
	}
	for z := -23.0; z <= -4; z += 2.6 {
		g.Carve(w.lancet(V(-24, wizardAcademyTop, z), 0.9, 1.8, 3, AxisX))
	}
	// Lawn paths and fountain.
	top := wizardAcademyTop
	g.Paint(Sand, Block(V(-24, top-1, -16.5), V(9, top+0.2, -14.5)), Block(V(-8.5, top-1, -26), V(-6.5, top+0.2, -5)))
	c := P(-7.5, -15.5)
	g.With(Stone, func() {
		g.Fill(Cylinder{Center: c, Radius: 2.8, Inner: 2, Y0: top - 0.5, Y1: top + 1})
		g.Fill(Cylinder{Center: c, Radius: 0.7, Y0: top, Y1: top + 2.5})
		g.Fill(Cylinder{Center: c, Radius: 1.3, Y0: top + 2.2, Y1: top + 2.7})
	})
	g.With(Water, func() { g.Fill(Cylinder{Center: c, Radius: 2, Y0: top - 0.5, Y1: top + 0.4}) })
	w.finial(g, c, top+3, 1)
	// Courtyard trees.
	for _, at := range []Vec2{P(-18, -21), P(3, -21), P(-18, -9.5), P(3, -9.5)} {
		g.Build(kit.Tree{At: at, Base: top, Height: 4.5})
	}
}

// greatHall is the long hall at the cliff edge: tall glowing lancets between buttresses,
// flying buttresses on the lake side, a steep slate roof and a slender flèche.
func (w WizardAcademy) greatHall(g *Grid) {
	const x0, x1, z0, z1 = -16.0, 9.0, -5.0, 4.0
	const top = 32.0
	base := wizardAcademyTop - 0.5
	w.hollow(g, V(x0, base, z0), V(x1, top, z1), 1.1, Stone)
	g.Paint(DarkStone, Block(V(x0-1, base, z0-1), V(x1+1, wizardAcademyTop+1.5, z1+1)))
	g.Paint(DarkStone, Block(V(x0-1, top-0.6, z0-1), V(x1+1, top, z1+1)))
	g.Paint(DarkStone, Block(V(x0-1, 27.8, z0-1), V(x1+1, 28.2, z1+1)))
	for k := 0; k < 8; k++ {
		x := x0 + 2.5 + 3*float64(k)
		if x > x1-1.5 {
			break
		}
		w.window(g, V(x, 20.5, z1), 0.85, 6.2, AxisZ, 1, Gold)
		w.window(g, V(x, 20.5, z0), 0.85, 6.2, AxisZ, -1, Gold)
		w.window(g, V(x, 28.6, z1), 0.6, 1, AxisZ, 1, Gold)
	}
	// Buttresses: flying on the lake side, plain on the courtyard side.
	g.Use(Stone)
	for k := 1; k < 8; k++ {
		x := x0 + 1 + 3*float64(k)
		if x > x1-1 {
			break
		}
		g.Fill(Block(V(x-0.6, base, z1), V(x+0.6, top-1, z1+0.9)))
		g.Fill(Block(V(x-0.75, base, 6.4), V(x+0.75, 22, 8.9)), Block(V(x-0.6, 22, 6.8), V(x+0.6, 28.5, 8.5)))
		g.Fill(Capsule{V(x, 27.6, 7.6), V(x, 30.6, z1+0.5), 0.6}, Capsule{V(x, 22.4, 7.6), V(x, 25.4, z1+0.5), 0.55})
		g.With(DarkStone, func() { g.FillVolume(Cone(P(x, 7.65), 0.95, 28.5, 3.4)) })
		g.Fill(Block(V(x-0.6, base, z0-1.4), V(x+0.6, 24, z0)), Block(V(x-0.6, 24, z0-0.8), V(x+0.6, 28, z0)))
	}
	// Steep roof with a gable wall at the west end, dormers and the flèche.
	const pitch = 1.7
	g.With(Stone, func() {
		g.Fill(Intersect(Block(V(x0, top, z0), V(x0+1.3, top+12, z1)), w.steep(P(x0, z0-0.7), P(x0+1.3, z1+0.7), top, pitch, AxisX)))
	})
	w.roof(g, P(x0+1.3, z0-0.7), P(x1, z1+0.7), top-0.5, pitch, 1.5, AxisX, Slate)
	for _, x := range []float64{-10, -1.5, 5.5} {
		g.With(Stone, func() { g.Fill(Block(V(x-1.2, top, 1), V(x+1.2, top+3, 4.4))) })
		g.With(Slate, func() { g.Fill(w.steep(P(x-1.7, 1), P(x+1.7, 4.9), top+3, 1.4, AxisZ)) })
		w.window(g, V(x, top+0.6, 4.4), 0.55, 1.2, AxisZ, 1, Gold)
	}
	ridge := top - 0.5 + 5.2*pitch
	f := P(-4, -0.5)
	g.With(Stone, func() { g.Fill(Cylinder{Center: f, Radius: 1.4, Y0: ridge - 3, Y1: ridge + 4}) })
	for k := 0; k < 4; k++ {
		g.Carve(Translate(RotateY(w.lancet(V(0, ridge+0.6, 0), 0.5, 1.6, 4, AxisX), P(0, 0), float64(k)*math.Pi/4), f.At(0)))
	}
	g.With(DarkStone, func() { g.Fill(Cylinder{Center: f, Radius: 1.8, Y0: ridge + 3.5, Y1: ridge + 4.2}) })
	w.cone(g, f, 1.9, ridge+4, 7, Navy)
	w.finial(g, f, ridge+11, 1.8)
}

// towers raises the corner and range towers of varied heights, and corbelled turrets.
func (w WizardAcademy) towers(g *Grid) {
	base := wizardAcademyTop - 0.5
	w.tower(g, P(-32, -32), 3.4, base, 46, 12, Slate, 11)
	w.tower(g, P(-8, -32.5), 3.6, base, 53, 13, Navy, 12)
	w.tower(g, P(15, -32), 3.0, base, 41, 10.5, Navy, 13)
	w.tower(g, P(9.8, 6), 3.2, base, 47, 11, Slate, 14)
	w.tower(g, P(-32.5, -13.5), 2.7, base, 40, 9.5, Navy, 15)
	w.tower(g, P(15.5, -13), 2.4, base, 37, 8.5, Slate, 16)
	// Corbelled turrets on range corners.
	for i, t := range []struct {
		at        Vec2
		base, top float64
		roof      Color
	}{
		{P(-24, -26), 26, 36, Navy},
		{P(9, -26), 25, 35, Slate},
		{P(15.3, -1), 26, 37, Navy},
		{P(-32.3, -1.5), 25, 36, Slate},
		{P(-16, -5.3), 28, 38, Navy},
	} {
		g.With(Stone, func() {
			g.Fill(Frustum{t.at, 0.5, 1.6, t.base - 3, t.base})
			g.Fill(Cylinder{Center: t.at, Radius: 1.6, Y0: t.base, Y1: t.top})
		})
		g.With(DarkStone, func() { g.Fill(Cylinder{Center: t.at, Radius: 2, Y0: t.top - 0.8, Y1: t.top}) })
		w.radialWindow(g, t.at, 1.6, float64(i)*1.3, t.base+2.5, 0.5, 1.4, Gold)
		w.cone(g, t.at, 2.1, t.top, 6.5, t.roof)
		w.finial(g, t.at, t.top+6.5, 1.3)
	}
}

// astronomy is the tallest tower: a tall round shaft, a corbelled gallery with a
// crenellated rail, an open arcaded observatory drum with a brass telescope, and a navy
// cone crowned by a gold spire.
func (w WizardAcademy) astronomy(g *Grid) {
	c := P(-21, 1)
	const r, shaft = 4.4, 58.0
	base := wizardAcademyTop - 0.5
	w.drum(g, c, r, 1.1, base, shaft, Stone)
	g.Paint(DarkStone, Cylinder{Center: c, Radius: r + 0.5, Y0: base, Y1: wizardAcademyTop + 1.5})
	for y := 26.0; y < shaft-4; y += 8 {
		g.Paint(DarkStone, Cylinder{Center: c, Radius: r + 0.5, Y0: y, Y1: y + 0.4})
	}
	level := 0
	for y := wizardAcademyTop + 4.0; y+3 < shaft-4; y += 5.5 {
		for k := 0; k < 5; k++ {
			a := float64(k)*2*math.Pi/5 + float64(level)*0.63
			w.radialWindow(g, c, r, a, y, 0.65, 2, w.pane(0.45, 21, level, k))
		}
		level++
	}
	// Gallery.
	g.With(DarkStone, func() { g.Fill(Frustum{c, r, r + 1.2, shaft - 2.5, shaft - 0.6}) })
	g.With(Stone, func() { g.Fill(Cylinder{Center: c, Radius: r + 1.2, Y0: shaft - 0.6, Y1: shaft}) })
	g.Use(Stone).Build(kit.RingBattlements{Center: c, Radius: r + 1.2, Top: shaft - 0.5, Count: 14})
	// Observatory drum with open arches.
	const dr, d0, d1 = 3.4, shaft, shaft + 8
	w.drum(g, c, dr, 1.1, d0, d1, Stone)
	for k := 0; k < 6; k++ {
		a := float64(k)*math.Pi/3 + 0.25
		g.Carve(Translate(RotateY(w.lancet(V(dr-0.5, d0+1, 0), 0.9, 3.2, 2.4, AxisX), P(0, 0), a), c.At(0)))
	}
	g.With(Gold, func() {
		tip := Polar(c, dr+2.2, math.Pi/2-0.4)
		g.Fill(Capsule{c.At(d0 + 2), tip.At(d0 + 5.2), 0.55})
		g.Fill(Cylinder{Center: c, Radius: 0.6, Y0: d0, Y1: d0 + 2.4})
	})
	g.With(DarkStone, func() { g.Fill(Frustum{c, dr, dr + 0.8, d1 - 0.8, d1}) })
	w.cone(g, c, dr+1.2, d1, 15, Navy)
	w.finial(g, c, d1+15, 4.5)
	g.With(Gold, func() { g.Fill(Sphere(c.At(d1+15+4.6), 0.8)) })
}

// clockCourt builds the far side of the ravine: the clock tower with four faces and an
// open belfry, the east range and tower, the glass greenhouse and a parapet terrace.
func (w WizardAcademy) clockCourt(g *Grid) {
	base := wizardAcademyTop - 0.5
	top := wizardAcademyTop
	w.wing(g, V(38, base, -22), V(49, 27, -15), AxisX, Navy, 4)
	w.tower(g, P(49.5, -22), 2.6, base, 36, 9, Slate, 17)
	// Clock tower.
	const x0, x1, z0, z1 = 30.0, 37.0, -14.0, -7.0
	const shaft, belfry, crown = 34.0, 40.5, 46.5
	cx, cz := (x0+x1)/2, (z0+z1)/2
	w.hollow(g, V(x0, base, z0), V(x1, belfry, z1), 1.0, Stone)
	g.With(DarkStone, func() {
		for _, x := range []float64{x0 - 0.4, x1 - 1} {
			for _, z := range []float64{z0 - 0.4, z1 - 1} {
				g.Fill(Block(V(x, base, z), V(x+1.4, crown-1, z+1.4)))
			}
		}
	})
	g.Paint(DarkStone, Block(V(x0-1, base, z0-1), V(x1+1, top+1.5, z1+1)))
	for _, y := range []float64{22.5, 28} {
		for _, o := range []float64{-1.6, 1.6} {
			w.window(g, V(cx+o, y, z1), 0.55, 2, AxisZ, 1, w.pane(0.5, int(y), int(o), 1))
			w.window(g, V(cx+o, y, z0), 0.55, 2, AxisZ, -1, w.pane(0.5, int(y), int(o), 2))
			w.window(g, V(x1, y, cz+o), 0.55, 2, AxisX, 1, w.pane(0.5, int(y), int(o), 3))
			w.window(g, V(x0, y, cz+o), 0.55, 2, AxisX, -1, w.pane(0.5, int(y), int(o), 4))
		}
	}
	// Gate passage where the bridge arrives.
	g.Carve(w.lancet(V(cx, top, cz), 1.6, 2.2, 9, AxisX))
	// Clock stage: projecting band with four faces.
	g.With(Stone, func() {
		g.Fill(Subtract(Block(V(x0-0.5, shaft, z0-0.5), V(x1+0.5, belfry, z1+0.5)), Block(V(x0+1, shaft-1, z0+1), V(x1-1, belfry+1, z1-1))))
	})
	g.Paint(DarkStone, Block(V(x0-0.6, shaft, z0-0.6), V(x1+0.6, shaft+0.5, z1+0.6)), Block(V(x0-0.6, belfry-0.5, z0-0.6), V(x1+0.6, belfry, z1+0.6)))
	cy := (shaft + belfry) / 2
	for _, f := range []struct {
		axis Axis
		out  float64
	}{{AxisZ, 1}, {AxisZ, -1}, {AxisX, 1}, {AxisX, -1}} {
		face := z1 + 0.5
		if f.out < 0 {
			face = z0 - 0.5
		}
		if f.axis == AxisX {
			face = x1 + 0.5
			if f.out < 0 {
				face = x0 - 0.5
			}
		}
		disc := func(r, a, b float64) Func {
			lo, hi := math.Min(face+f.out*a, face+f.out*b), math.Max(face+f.out*a, face+f.out*b)
			box := Box{V(cx-r, cy-r, lo), V(cx+r, cy+r, hi)}
			if f.axis == AxisX {
				box = Box{V(lo, cy-r, cz-r), V(hi, cy+r, cz+r)}
			}
			return Func{box, func(p Vec3) bool {
				u := p.X - cx
				if f.axis == AxisX {
					u = p.Z - cz
				}
				return u*u+(p.Y-cy)*(p.Y-cy) <= r*r
			}}
		}
		g.With(Gold, func() { g.Fill(disc(2.7, -0.2, 0.4)) })
		g.With(White, func() { g.Fill(disc(2.15, -0.2, 0.6)) })
		hand := func(du, dy float64) Capsule {
			a := V(cx, cy, face+f.out*0.8)
			if f.axis == AxisX {
				a = V(face+f.out*0.8, cy, cz)
			}
			b := a.Add(V(du, dy, 0))
			if f.axis == AxisX {
				b = a.Add(V(0, dy, du))
			}
			return Capsule{a, b, 0.42}
		}
		g.With(Charcoal, func() { g.Fill(hand(0, 1.7), hand(1.2, -0.4)) })
	}
	// Open belfry with a gold bell, parapet and corner pinnacles.
	w.hollow(g, V(x0, belfry, z0), V(x1, crown, z1), 1.1, Stone)
	g.Carve(Block(V(x0+1.1, belfry-1, z0+1.1), V(x1-1.1, crown-1.1, z1-1.1)))
	for _, o := range []float64{-1.5, 1.5} {
		g.Carve(w.lancet(V(cx+o, belfry+1, cz), 0.65, 2.6, 9, AxisZ), w.lancet(V(cx, belfry+1, cz+o), 0.65, 2.6, 9, AxisX))
	}
	g.With(Gold, func() {
		g.Fill(Frustum{P(cx, cz), 1.5, 0.8, belfry + 0.8, belfry + 3})
		g.Fill(Sphere(V(cx, belfry+3.2, cz), 0.7))
	})
	g.Use(Stone).Build(kit.Battlements{Min: P(x0-0.4, z0-0.4), Max: P(x1+0.4, z1+0.4), Top: crown - 1, Period: 1.8})
	for _, x := range []float64{x0 + 0.3, x1 - 0.3} {
		for _, z := range []float64{z0 + 0.3, z1 - 0.3} {
			w.cone(g, P(x, z), 0.95, crown-0.5, 4, Slate)
		}
	}
	// Spire.
	spire := Func{Box{V(x0, crown-1, z0), V(x1, crown+15, z1)}, func(p Vec3) bool {
		k := 3.3 * (1 - (p.Y-(crown-1))/16)
		return math.Max(math.Abs(p.X-cx), math.Abs(p.Z-cz)) <= k
	}}
	g.With(Slate, func() { g.FillVolume(spire) })
	w.finial(g, P(cx, cz), crown+14.2, 3)

	// Greenhouse: glass barrel vault with white ribs on a stone kerb.
	const gx0, gx1, gz, gr, gh = 37.0, 49.0, -3.0, 3.4, 4.8
	g.With(Stone, func() { g.Fill(Block(V(gx0-0.5, base, gz-gr-0.5), V(gx1+0.5, top+0.8, gz+gr+0.5))) })
	vault := func(t float64) Func {
		return Func{Box{V(gx0, top, gz-gr), V(gx1, top+0.8+gh, gz+gr)}, func(p Vec3) bool {
			u, v := (p.Z-gz)/gr, (p.Y-top-0.8)/gh
			q := u*u + v*v
			if q > 1 || p.Y < top+0.8 {
				return false
			}
			if p.X < gx0+0.9 || p.X > gx1-0.9 {
				return true
			}
			ui, vi := (p.Z-gz)/(gr-t), (p.Y-top-0.8)/(gh-t)
			return ui*ui+vi*vi > 1
		}}
	}
	g.With(Glass, func() { g.Fill(vault(1)) })
	g.Carve(Func{Box{V(gx0+0.9, top+0.8, gz-gr), V(gx1-0.9, top+gh, gz+gr)}, func(p Vec3) bool {
		u, v := (p.Z-gz)/(gr-1), (p.Y-top-0.8)/(gh-1)
		return u*u+v*v <= 1
	}})
	g.Paint(White, Func{Box{V(gx0, top, gz-gr), V(gx1, top+gh+1, gz+gr)}, func(p Vec3) bool {
		return math.Mod(p.X-gx0, 2.4) < 0.5 || math.Abs(p.Z-gz) < 0.4
	}})
	g.With(White, func() { g.Fill(Block(V(gx0, top+gh+0.4, gz-0.5), V(gx1, top+gh+1.2, gz+0.5))) })
	// Terrace parapet along the cliff.
	g.Use(Stone).Fill(Block(V(30, base, 6.6), V(50, top+1.2, 7.8)))
	g.Build(kit.Battlements{Min: P(30, 6.6), Max: P(50, 7.8), Top: top + 0.6, Period: 2})
	g.Paint(Sand, Block(V(28.5, top-1, -12), V(38, top+0.2, -9)), Block(V(36, top-1, -9), V(39, top+0.2, 6)))
}

// bridge is the covered wooden bridge across the ravine on two tall stone piers.
func (w WizardAcademy) bridge(g *Grid) {
	const xa, xb, z0, z1 = 14.5, 30.5, -12.6, -8.4
	deck := wizardAcademyTop + 0.4
	g.Use(Stone)
	for _, x := range []float64{19.8, 25.2} {
		g.Fill(Block(V(x-1, 0, z0-0.4), V(x+1, deck-0.8, z1+0.4)))
		g.With(DarkStone, func() { g.Fill(Block(V(x-1.4, deck-1.6, z0-0.6), V(x+1.4, deck-0.8, z1+0.6))) })
		g.Carve(w.lancet(V(x, 4, (z0+z1)/2), 0.6, 3, 6, AxisZ))
	}
	g.Use(Wood)
	g.Fill(Block(V(xa, deck-0.8, z0), V(xb, deck, z1)))
	// Timber braces from piers to deck.
	for _, x := range []float64{19.8, 25.2} {
		for _, s := range []float64{-1, 1} {
			for _, z := range []float64{z0 + 0.5, z1 - 0.5} {
				g.Fill(Capsule{V(x+s*1, deck-5, z), V(x+s*3.2, deck-0.8, z), 0.45})
			}
		}
	}
	// Walls: a low boarded rail, posts, a window band and a top beam.
	for _, z := range []float64{z0, z1 - 0.9} {
		g.Fill(Block(V(xa, deck, z), V(xb, deck+1.2, z+0.9)))
		g.Fill(Block(V(xa, deck+3, z), V(xb, deck+3.8, z+0.9)))
		for x := xa; x <= xb; x += 2 {
			g.Fill(Block(V(x, deck, z), V(x+0.9, deck+3.8, z+0.9)))
		}
	}
	g.With(Slate, func() { g.Fill(w.steep(P(xa-0.3, z0-0.7), P(xb+0.3, z1+0.7), deck+3.8, 1.1, AxisX)) })
	g.Carve(w.steep(P(xa-1, z0+0.5), P(xb+1, z1-0.5), deck+3.4, 1.1, AxisX))
	g.With(Gold, func() {
		for _, x := range []float64{xa + 1, xb - 1} {
			g.Fill(Sphere(V(x, deck+2.6, z1+0.3), 0.55))
		}
	})
}

// viaduct carries the railway across the glen on two tiers of arches into a tunnel
// portal under the castle's west range.
func (w WizardAcademy) viaduct(g *Grid, m *wizardAcademyMap) {
	const xa, xb, zc = -55.0, -39.0, -7.5
	deck := wizardAcademyRail
	g.Use(Stone).Fill(Func{Box{V(xa, 0, zc-1.6), V(xb, deck, zc+1.6)}, func(p Vec3) bool {
		return p.Y >= math.Max(m.at(p.X, p.Z)-1.2, -0.2)
	}})
	for i := -3; i <= 3; i++ {
		x := w.glen(zc) + 3.6*float64(i)
		g.Carve(Arch{V(x, 0, zc), 1.25, 3.4, 5, AxisZ})
		g.Carve(Arch{V(x, 5.8, zc), 1.25, 3, 5, AxisZ})
	}
	g.Paint(DarkStone, Block(V(xa, 5.2, zc-2), V(xb, 5.7, zc+2)))
	g.With(DarkStone, func() {
		g.Fill(Block(V(xa, deck-0.6, zc-2), V(xb+1, deck, zc+2)))
		g.Fill(Block(V(xa, deck, zc-2), V(xb+1, deck+0.9, zc-1.3)), Block(V(xa, deck, zc+1.3), V(xb+1, deck+0.9, zc+2)))
	})
	g.With(Wood, func() { g.Fill(Block(V(xa-1, deck-0.4, zc-1.2), V(xb+1, deck+0.1, zc+1.2))) })
	g.With(Charcoal, func() {
		g.Fill(Block(V(xa-1, deck, zc-0.9), V(xb+1, deck+0.3, zc-0.4)), Block(V(xa-1, deck, zc+0.4), V(xb+1, deck+0.3, zc+0.9)))
	})
	// Tunnel portal in the crag.
	const px = -39.0
	g.Use(Stone).Fill(Block(V(px-1, deck-1, zc-4), V(px+1, deck+5.5, zc+4)))
	g.Paint(DarkStone, Block(V(px-1.2, deck+4.8, zc-4.2), V(px+1.2, deck+5.5, zc+4.2)))
	g.Carve(Arch{V(px, deck-0.5, zc), 1.7, 2.2, 4, AxisX})
	g.With(Charcoal, func() { g.Fill(Block(V(px+1, deck-0.5, zc-2), V(px+2.2, deck+4.2, zc+2))) })
}

// train is a little red steam engine with tender and two carriages heading for the
// tunnel, trailing puffs of steam.
func (w WizardAcademy) train(g *Grid) {
	const zc = -7.5
	y := wizardAcademyRail + 0.3
	axial := func(x0, x1, cy, r float64) Func {
		return Func{Box{V(x0, cy-r, zc-r), V(x1, cy+r, zc+r)}, func(p Vec3) bool {
			return (p.Y-cy)*(p.Y-cy)+(p.Z-zc)*(p.Z-zc) <= r*r
		}}
	}
	// Engine.
	const nose = -42.0
	g.With(Charcoal, func() {
		g.Fill(Block(V(nose-6, y, zc-0.9), V(nose, y+1, zc+0.9)))
		g.Fill(axial(nose-1, nose+0.1, y+2, 1.05))
		g.Fill(Cylinder{Center: P(nose-0.9, zc), Radius: 0.5, Y0: y + 2.5, Y1: y + 4.3})
		g.Fill(Cylinder{Center: P(nose-0.9, zc), Radius: 0.75, Y0: y + 3.8, Y1: y + 4.5})
	})
	g.With(Red, func() {
		g.Fill(axial(nose-4.5, nose-1, y+2, 1.05))
		g.Fill(Block(V(nose-6.2, y+1, zc-1.2), V(nose-4.3, y+3.8, zc+1.2)))
		g.Fill(Block(V(nose+0.1, y, zc-1.1), V(nose+0.6, y+0.8, zc+1.1)))
	})
	g.Carve(Block(V(nose-6.2, y+2.4, zc-1.4), V(nose-4.4, y+3.3, zc+1.4)))
	g.With(Charcoal, func() { g.Fill(Block(V(nose-6.5, y+3.8, zc-1.4), V(nose-4, y+4.3, zc+1.4))) })
	g.With(Gold, func() {
		g.Fill(Sphere(V(nose-2.8, y+3.1, zc), 0.6))
		g.Fill(axial(nose-3.6, nose-3.2, y+2, 1.15))
	})
	// Tender.
	g.With(Charcoal, func() { g.Fill(Block(V(nose-9, y, zc-0.9), V(nose-6.5, y+1, zc+0.9))) })
	g.With(Red, func() { g.Fill(Block(V(nose-9, y+1, zc-1.1), V(nose-6.5, y+2.6, zc+1.1))) })
	g.With(Charcoal, func() { g.Fill(Block(V(nose-8.6, y+2.6, zc-0.8), V(nose-6.9, y+3, zc+0.8))) })
	// Carriages.
	for i := 0; i < 2; i++ {
		x1 := nose - 9.6 - float64(i)*5.4
		x0 := x1 - 4.8
		g.With(Charcoal, func() { g.Fill(Block(V(x0+0.3, y, zc-0.9), V(x1-0.3, y+0.8, zc+0.9))) })
		g.With(Red, func() { g.Fill(Block(V(x0, y+0.8, zc-1.15), V(x1, y+3.2, zc+1.15))) })
		g.Paint(Gold, Func{Box{V(x0+0.5, y+1.9, zc-1.2), V(x1-0.5, y+2.6, zc+1.2)}, func(p Vec3) bool {
			return math.Mod(p.X-x0-0.5, 1.3) < 0.8 && math.Abs(p.Z-zc) > 0.7
		}})
		g.With(Charcoal, func() { g.Fill(Block(V(x0-0.2, y+3.2, zc-1.3), V(x1+0.2, y+3.7, zc+1.3))) })
	}
	// Steam.
	g.With(White, func() {
		for i, s := range [][4]float64{{-42.9, 18.6, 0.9, 0}, {-44.2, 19.8, 1.15, 0.2}, {-46.2, 20.7, 1.4, 0.4}, {-48.8, 21.2, 1.55, 0.5}, {-51.6, 21.4, 1.3, 0.6}} {
			_ = i
			g.Fill(Sphere(V(s[0], s[1], zc+s[3]), s[2]))
		}
	})
}

// lakeside builds the stone quay and wooden boathouse at the cliff foot, a jetty and
// lantern-lit rowing boats on the lake.
func (w WizardAcademy) lakeside(g *Grid) {
	g.Use(DarkStone).Fill(Block(V(-10, -0.2, 13), V(5, 1, 16.5)))
	g.Use(Stone).Fill(Block(V(-10, 0.6, 15.8), V(5, 1.2, 16.5)))
	// Boathouse.
	const x0, x1, z0, z1 = -6.0, 1.0, 14.0, 21.5
	w.hollow(g, V(x0, 0.6, z0), V(x1, 5, z1), 1, Wood)
	g.Carve(Arch{V((x0+x1)/2, -0.5, z1), 2, 2.4, 3, AxisZ})
	g.Carve(Block(V(x0+1, -0.5, z0+1), V(x1-1, 0.7, z1-1)))
	g.With(Water, func() { g.Fill(Block(V(x0+1, -0.2, z0+1), V(x1-1, 0.2, z1+0.5))) })
	for _, z := range []float64{16.5, 19} {
		w.window(g, V(x0, 2.4, z), 0.5, 1, AxisX, -1, Gold)
		w.window(g, V(x1, 2.4, z), 0.5, 1, AxisX, 1, Gold)
	}
	g.With(Slate, func() { g.Fill(Gable{P(x0-0.7, z0), P(x1+0.7, z1+0.8), 5, AxisZ}) })
	g.Carve(Gable{P(x0+0.5, z0+1), P(x1-0.5, z1-0.6), 4.6, AxisZ})
	// Jetty with lanterns.
	g.Use(Wood).Fill(Block(V(2, 0.4, 16), V(3.8, 0.9, 28)))
	for z := 18.0; z <= 28; z += 3.3 {
		g.Fill(Cylinder{Center: P(2.9, z), Radius: 0.5, Y0: -0.2, Y1: 0.4})
	}
	g.Fill(Cylinder{Center: P(3.5, 27.5), Radius: 0.45, Y0: 0.9, Y1: 3})
	g.With(Gold, func() { g.Fill(Sphere(V(3.5, 3.3, 27.5), 0.6)) })
	// Boats.
	for i, b := range [][3]float64{{5.6, 21.5, 1.57}, {-12, 24, 0.4}, {7, 31, -0.5}, {-3, 34, 0.15}, {16, 25, 0.9}, {-22, 30, -0.3}, {-34, 26, 0.6}} {
		at := P(b[0], b[1])
		hull := Intersect(Ellipsoid{V(0, 0.9, 0), V(2.1, 1.1, 0.95)}, Block(V(-3, 0.1, -2), V(3, 0.95, 2)))
		g.With(Wood, func() { g.Fill(Translate(RotateY(hull, P(0, 0), b[2]), at.At(0))) })
		g.With(Charcoal, func() {
			g.Fill(Translate(RotateY(Ellipsoid{V(0, 0.95, 0), V(1.4, 0.4, 0.4)}, P(0, 0), b[2]), at.At(0)))
		})
		if i%2 == 0 {
			stern := at.Add(P(-1.4, 0).Rotate(b[2]))
			g.With(Gold, func() { g.Fill(Sphere(stern.At(1.6), 0.55)) })
		}
	}
}

// hut is the gamekeeper's round stone hut by the glen stream, with a pumpkin patch.
func (w WizardAcademy) hut(g *Grid, m *wizardAcademyMap) {
	c := P(-40.5, 14)
	y := m.at(c.X, c.Y)
	g.With(DarkStone, func() { g.Fill(Cylinder{Center: c, Radius: 3.3, Y0: y - 1.5, Y1: y + 0.3}) })
	w.drum(g, c, 2.6, 1, y, y+3.6, Stone)
	g.Carve(Arch{V(c.X+2.4, y+0.3, c.Y), 0.7, 1.4, 1.5, AxisX})
	g.With(Wood, func() { g.Fill(Block(V(c.X+1.7, y+0.3, c.Y-0.6), V(c.X+1.9, y+2.2, c.Y+0.6))) })
	w.radialWindow(g, c, 2.6, math.Pi/2, y+1.2, 0.5, 0.9, Gold)
	w.radialWindow(g, c, 2.6, -0.4, y+1.2, 0.5, 0.9, Gold)
	g.With(Wood, func() { g.FillVolume(Cone(c, 3.6, y+3.4, 3.8)) })
	g.With(DarkStone, func() { g.Fill(Cylinder{Center: P(c.X-1.2, c.Y-1.2), Radius: 0.6, Y0: y + 3, Y1: y + 7}) })
	// Pumpkin patch.
	for i := 0; i < 5; i++ {
		at := P(c.X-2+float64(i%3)*1.5, c.Y-3.8-float64(i/3)*1.5)
		h := math.Max(m.low(at.X, at.Y), 0.4)
		g.With(Terracotta, func() { g.Fill(Sphere(at.At(h+0.3), 0.55+0.3*Hash01(i, 77))) })
	}
	// Woodpile.
	g.With(Wood, func() { g.Fill(Block(V(c.X-1, y, c.Y+2.7), V(c.X+1.5, y+1.2, c.Y+3.6))) })
}

// trees scatters dark pines: a dense forest over the western glen and hill, thinner
// stands behind and east of the castle.
func (w WizardAcademy) trees(g *Grid, m *wizardAcademyMap) {
	for i := -17; i <= 17; i++ {
		for j := -13; j <= 12; j++ {
			at := P(float64(i)*3.8+3*(Hash01(i, j, 31)-0.5), float64(j)*3.8+3*(Hash01(i, j, 32)-0.5))
			dense := at.X < -36 || at.Y < -36
			if Hash01(i, j, 33) < 0.3 || (!dense && Hash01(i, j, 34) < 0.45) {
				continue
			}
			h := m.at(at.X, at.Y)
			lo := m.low(at.X, at.Y)
			if h < 1 || lo < 0.8 || h-lo > 2.6 || w.site(at.X, at.Y) > 0.93 {
				continue
			}
			if at.X > -39 && at.X < 53 && at.Y > -36 && at.Y < 13 {
				continue // Castle crag.
			}
			if at.X < -36 && math.Abs(at.Y+7.5) < 3.5 || at.X > -58 && at.X < -36 && math.Abs(at.Y+5) < 6 {
				continue // Railway.
			}
			if at.Sub(P(-38, 16)).Length() < 6 {
				continue // Hut clearing.
			}
			if math.Abs(at.X-w.glen(at.Y)) < 2.5 || math.Abs(at.X-w.ravine(at.Y)) < 3 {
				continue // Streams.
			}
			w.pine(g, at, lo-0.3, 5.5+4.5*Hash01(i, j, 35))
		}
	}
}

// pine is a trunk with three stacked foliage cones.
func (WizardAcademy) pine(g *Grid, at Vec2, base, h float64) {
	g.With(Wood, func() { g.Fill(Cylinder{Center: at, Radius: 0.55, Y0: base - 1, Y1: base + h*0.3}) })
	g.With(Foliage, func() {
		g.Fill(Cone(at, h*0.26, base+h*0.15, h*0.42), Cone(at, h*0.2, base+h*0.4, h*0.36), Cone(at, h*0.13, base+h*0.64, h*0.36))
	})
}
