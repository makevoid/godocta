package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// HagiaSophia is Istanbul's great domed basilica. The shallow lead central dome rides
// on a ring of forty arched windows over a square of great arches; the north and south
// arches are closed by windowed tympanum walls, the east and west by half-domes, each
// with two smaller exedra semi-domes cascading down to the gallery roofs (and the apse
// semi-dome to the east). Massive buttress towers flank the dome to north and south.
// Four Ottoman minarets with pencil-point caps stand at the corners. West of the
// narthexes is the courtyard with its ablution fountain; to the south lie the domed
// türbes and, to the south-west, a corner of Sultanahmet park with its fountain.
// The apse points to +X (east); north is -Z.
type HagiaSophia struct{}

func init() { Register(HagiaSophia{}) }

func (HagiaSophia) Name() string  { return "hagia_sophia" }
func (HagiaSophia) Title() string { return "Hagia Sophia" }

// Main dimensions shared by the helpers.
const (
	hagiaSophiaR      = 12.0 // Central dome radius and half-width of the central square.
	hagiaSophiaFloor  = 0.3  // Top of the ground plate.
	hagiaSophiaRoof   = 12.0 // Gallery roof level.
	hagiaSophiaSpring = 17.0 // Spring of the great arches and half-domes.
	hagiaSophiaBX     = 26.0 // Main block half-length (east-west).
	hagiaSophiaBZ     = 24.0 // Main block half-width (north-south).
)

// shellBox is a hollow rectangular box with walls of thickness t and a roof slab.
func (HagiaSophia) shellBox(x0, z0, x1, z1, y0, y1, t float64) Solid {
	return Subtract(Block(V(x0, y0, z0), V(x1, y1, z1)), Block(V(x0+t, y0, z0+t), V(x1-t, y1-t, z1-t)))
}

// spur is a sloped Ottoman buttress against a wall: foot is the wall face point, out
// the unit outward direction; the top slopes from top at the wall down to 2.5 at l.
func (HagiaSophia) spur(foot, out Vec2, w, l, top float64) Solid {
	side := P(-out.Y, out.X)
	b := EmptyBox()
	for _, a := range []float64{-0.6, l} {
		for _, s := range []float64{-w / 2, w / 2} {
			q := foot.Add(out.Scale(a)).Add(side.Scale(s))
			b = b.Union(Box{q.At(hagiaSophiaFloor), q.At(top)})
		}
	}
	return Func{b, func(p Vec3) bool {
		d := p.Plan().Sub(foot)
		a, s := d.Dot(out), d.Dot(side)
		if a < -0.6 || a > l || math.Abs(s) > w/2 || p.Y < hagiaSophiaFloor {
			return false
		}
		return p.Y <= top-(top-2.5)*math.Max(a, 0)/l
	}}
}

// radialArch is an arched opening through a curved wall around center, at radius r
// and plan angle a.
func (HagiaSophia) radialArch(center Vec2, r, a, sill, hw, spring, depth float64) Solid {
	arch := Arch{Sill: V(r, sill, 0), HalfWidth: hw, Spring: spring, Depth: depth, Through: AxisX}
	return Translate(RotateY(arch, P(0, 0), a), center.At(0))
}

// quarterShell is the outer half (dot(p-c, dir) >= 0, above c.Y) of a spherical shell.
func (HagiaSophia) quarterShell(c Vec3, dir Vec2, r, t float64) Solid {
	return Func{Box{c.Sub(V(r, 0, r)), c.Add(V(r, r, r))}, func(p Vec3) bool {
		d := p.Sub(c)
		if d.Y < 0 || d.Plan().Dot(dir) < 0 {
			return false
		}
		l := d.Length()
		return l <= r && l >= r-t
	}}
}

// hagiaSophiaTurbes lists the mausolea: Selim II, Murad III and Mehmed III.
var hagiaSophiaTurbes = []struct {
	c       Vec2
	n       int
	r, wall float64
}{{P(16.5, 41), 8, 4.8, 8.4}, {P(5, 42), 6, 4, 7.6}, {P(29, 42), 8, 4.1, 7.8}}

// interior reports whether a point lies in the enclosed hollow of the church, its
// narthexes or the türbes; shadeInterior darkens the faces around these spaces.
func (HagiaSophia) interior(p Vec3) bool {
	const (
		R, floor, roof, spring = hagiaSophiaR, hagiaSophiaFloor, hagiaSophiaRoof, hagiaSophiaSpring
		bx, bz                 = hagiaSophiaBX, hagiaSophiaBZ
		domeBase               = spring + R
		drumTop                = domeBase + 4.2
	)
	if p.Y <= floor {
		return false
	}
	ax, az := math.Abs(p.X), math.Abs(p.Z)
	if ax < bx-1.2 && az < bz-1.2 && p.Y < roof-0.8 {
		return true
	}
	// Inner narthex and exonarthex.
	if p.X > -bx-6.8 && p.X < -bx-0.7 && az < 20.8 && p.Y < 12.3 {
		return true
	}
	if p.X > -bx-10.5 && p.X < -bx-8 && az < 20 && p.Y < 7 {
		return true
	}
	// Apse.
	if p.X >= 2*R {
		d := math.Hypot(p.X-2*R, p.Z)
		if (p.Y < roof && d < 4.6) || math.Hypot(d, p.Y-roof) < 4.7 {
			return true
		}
	}
	if p.Y >= roof-0.8 {
		if ax < R-0.6 && az < R-0.6 && p.Y < domeBase-1 {
			return true
		}
		if math.Hypot(p.X, p.Z) < R-0.8 && p.Y < drumTop {
			return true
		}
		domeC := V(0, drumTop-0.3, 0)
		if p.Y >= domeC.Y && (Ellipsoid{domeC, V(R-0.6, 7.4, R-0.6)}).Contains(p) {
			return true
		}
		for _, s := range []float64{-1, 1} {
			dx := (p.X - s*R) * s
			d := math.Hypot(dx, p.Z)
			if dx >= 0 && d < R-0.2 && (p.Y < spring || math.Hypot(d, p.Y-spring) < R-0.8) {
				return true
			}
		}
	}
	for _, t := range hagiaSophiaTurbes {
		if p.Y > 1.2 && p.Y < t.wall-0.8 && p.Plan().Sub(t.c).Length() < t.r-1.4 {
			return true
		}
	}
	return false
}

// shadeInterior paints every face that borders an interior hollow Charcoal, so that
// windows and doors read as dark openings.
func (h HagiaSophia) shadeInterior(g *Grid) {
	var dark []Cell
	for _, c := range g.Cells() {
		for _, n := range Neighbors {
			m := c.Add(n)
			if !g.Has(m) && h.interior(g.World(m)) {
				dark = append(dark, c)
				break
			}
		}
	}
	g.With(Charcoal, func() {
		for _, c := range dark {
			g.Remove(c)
			g.Add(c)
		}
	})
}

func (h HagiaSophia) Build(g *Grid) Order {
	h.ground(g)
	h.mainBlock(g)
	h.domes(g)
	h.buttresses(g)
	h.narthex(g)
	h.minarets(g)
	h.courtyard(g)
	h.turbes(g)
	h.shadeInterior(g)
	h.park(g)
	return Outward(P(0, 0))
}

// ground is the paved plate with lawns and paths.
func (HagiaSophia) ground(g *Grid) {
	const x0, x1, z0, z1 = -58.0, 40.0, -34.0, 50.0
	g.Use(Stone).Fill(Block(V(x0, 0, z0), V(x1, hagiaSophiaFloor, z1)))
	lawn := func(a, b Vec2) { g.Paint(Grass, Block(a.At(0), b.At(hagiaSophiaFloor))) }
	// North lawn strip, Sultanahmet park, türbe garden, east lawn.
	lawn(P(x0+1, z0+1), P(x1-1, -29))
	lawn(P(x0+1, 31), P(-6, z1-1))
	lawn(P(0, 33), P(x1-1, z1-1))
	lawn(P(34, -28), P(x1-1, 30))
	// Paths: Sand walks across the lawns.
	g.Paint(Sand, Func{Box{V(x0, 0, z0), V(x1, hagiaSophiaFloor, z1)}, func(p Vec3) bool {
		x, z := p.X, p.Z
		park := x < -6 && z > 31 && (math.Abs(x+31) < 1.6 || math.Abs(z-41) < 1.6 ||
			math.Abs(math.Hypot(x+31, z-41)-8.5) < 1.2)
		north := z < -29 && math.Abs(z+31.5) < 1.2
		turbe := x > 0 && z > 33 && (math.Abs(z-34.6) < 1.1 || math.Abs(x-11.5) < 1 || math.Abs(x-22.5) < 1)
		return park || north || turbe
	}})
	// A dark curb around the site.
	g.Paint(Sand, Subtract(Block(V(x0, 0, z0), V(x1, hagiaSophiaFloor, z1)), Block(V(x0+0.8, -1, z0+0.8), V(x1-0.8, 1, z1-0.8))))
}

// mainBlock is the gallery block: hollow walls with two rows of arched windows, the
// gallery roofs, the apse and the sloped Ottoman buttresses.
func (h HagiaSophia) mainBlock(g *Grid) {
	const (
		R, floor, roof = hagiaSophiaR, hagiaSophiaFloor, hagiaSophiaRoof
		bx, bz         = hagiaSophiaBX, hagiaSophiaBZ
	)
	g.Use(Terracotta).Fill(h.shellBox(-bx, -bz, bx, bz, floor, roof, 1.2))
	// Open the roof over the nave (under the domes) to save hidden cells.
	g.Carve(Func{Box{V(-2*R, roof-1.5, -R), V(2*R, roof+0.1, R)}, func(p Vec3) bool {
		return math.Hypot(math.Max(math.Abs(p.X)-R, 0), p.Z) < R-1.4
	}})
	g.Paint(Slate, Block(V(-bx+0.7, roof-0.45, -bz+0.7), V(bx-0.7, roof+0.1, bz-0.7)))
	// Low lead lean-to roofs over the north and south aisles.
	g.Use(Slate).Fill(Func{Box{V(-bx+0.5, roof, -bz), V(bx-0.5, roof+2, bz)}, func(p Vec3) bool {
		az := math.Abs(p.Z)
		if az < R+1 || az > bz-0.5 {
			return false
		}
		top := roof + 1.6*(bz-az)/(bz-R-1)
		return p.Y <= top && p.Y >= top-0.8
	}})
	// Cornice and string course.
	band := func(y0, y1 float64) Solid {
		return Subtract(Block(V(-bx-0.25, y0, -bz-0.25), V(bx+0.25, y1, bz+0.25)), Block(V(-bx+1, y0-1, -bz+1), V(bx-1, y1+1, bz-1)))
	}
	g.Use(Sand).Fill(band(roof-0.6, roof), band(6.2, 6.6))
	// Windows: two rows on each face, with Sand frames.
	window := func(at Vec3, through Axis, hw, spring float64) {
		frame := Arch{Sill: at.Sub(V(0, 0.4, 0)), HalfWidth: hw + 0.45, Spring: spring + 0.4, Depth: 2.6, Through: through}
		g.Paint(Sand, frame)
		g.Carve(Arch{Sill: at, HalfWidth: hw, Spring: spring, Depth: 3.2, Through: through})
	}
	rows := [][2]float64{{1.4, 2.9}, {7.4, 2.5}}
	for _, zs := range []float64{-1, 1} {
		for _, x := range []float64{0, 3.3, -3.3, 6.6, -6.6, 18.4, -18.4, 22, -22} {
			for _, r := range rows {
				window(V(x, r[0], zs*(bz-0.6)), AxisZ, 0.9, r[1])
			}
		}
	}
	for _, xs := range []float64{-1, 1} {
		for _, z := range []float64{8, -8, 16.2, -16.2, 20, -20} {
			for _, r := range rows {
				window(V(xs*(bx-0.6), r[0], z), AxisX, 0.9, r[1])
			}
		}
	}
	// Apse: a semicircular wall with two rows of three windows and a semi-dome.
	apse := P(2*R, 0)
	g.Use(Terracotta).FillVolume(Func{Box{V(2*R, floor, -6), V(2*R+6, roof, 6)}, func(p Vec3) bool {
		d := p.Plan().Sub(apse).Length()
		return p.X >= 2*R && d <= 5.8 && d >= 4.6
	}})
	g.Paint(Sand, Func{Box{V(2*R, roof-0.6, -6.5), V(2*R+6.5, roof, 6.5)}, func(p Vec3) bool { return p.X >= bx-0.5 }})
	g.Paint(Sand, Func{Box{V(2*R, 6.2, -6.5), V(2*R+6.5, 6.6, 6.5)}, func(p Vec3) bool { return p.X >= bx-0.5 }})
	for _, a := range []float64{-0.75, 0, 0.75} {
		g.Carve(h.radialArch(apse, 5.2, a, 1.6, 0.9, 2.8, 3), h.radialArch(apse, 5.2, a, 7.4, 0.9, 2.5, 3))
	}
	// Sloped buttresses at the corners and beside the apse.
	g.Use(Terracotta)
	for _, s := range []float64{-1, 1} {
		g.Fill(h.spur(P(bx, s*12.2), P(1, 0), 2.4, 4.5, roof-0.5))
		g.Fill(h.spur(P(s*(bx-2), bz), P(0, 1), 2.4, 4, roof-0.5))
		g.Fill(h.spur(P(s*(bx-2), -bz), P(0, -1), 2.4, 4, roof-0.5))
		g.Fill(h.spur(P(s*1.5+2*R+s*4.6, s*4.6), P(0.6, s*0.8), 1.8, 3.5, roof-1.5))
	}
}

// domes builds the central square of great arches, the windowed tympana, the drum
// with forty windows, the central dome, the half-domes, exedrae and apse semi-dome.
func (h HagiaSophia) domes(g *Grid) {
	const (
		R, roof, spring = hagiaSophiaR, hagiaSophiaRoof, hagiaSophiaSpring
		domeBase        = spring + R
		drumTop         = domeBase + 4.2
	)
	o := P(0, 0)
	// Tympanum walls closing the north and south arches, with three rows of windows.
	for _, zs := range []float64{-1, 1} {
		zc := zs * (R + 0.6)
		g.Use(Terracotta).Fill(Func{Box{V(-R, roof-1, zc-0.6), V(R, spring+R, zc+0.6)}, func(p Vec3) bool {
			return p.Y <= spring || math.Hypot(p.X, p.Y-spring) <= R
		}})
		win := func(x, sill, spring float64) {
			g.Paint(Sand, Arch{Sill: V(x, sill-0.4, zs*(R+1)), HalfWidth: 1.2, Spring: spring + 0.4, Depth: 0.6, Through: AxisZ})
			g.Carve(Arch{Sill: V(x, sill, zc), HalfWidth: 0.8, Spring: spring, Depth: 3, Through: AxisZ})
		}
		for k := -3; k <= 3; k++ {
			win(float64(k)*2.7, roof+1.6, 2.6)
		}
		for k := -2; k <= 2; k++ {
			win(float64(k)*2.7, roof+6.8, 2.4)
		}
		for k := -1; k <= 1; k++ {
			win(float64(k)*2.7, roof+11.6, 1.6)
		}
	}
	// The four great arches: projecting bands, Sand-edged, on the corner piers.
	archNS := Func{Box{V(-R-2, roof, -R-2), V(R+2, domeBase, R+2)}, func(p Vec3) bool {
		az := math.Abs(p.Z)
		if az < R-0.6 || az > R+2 {
			return false
		}
		ax := math.Abs(p.X)
		if p.Y < spring {
			return ax >= R && ax <= R+2
		}
		r := math.Hypot(p.X, p.Y-spring)
		return r >= R && r <= R+2
	}}
	archEW := Func{Box{V(-R-1.4, roof, -R-2), V(R+1.4, domeBase, R+2)}, func(p Vec3) bool {
		ax := math.Abs(p.X)
		if ax < R-0.6 || ax > R+1.4 {
			return false
		}
		az := math.Abs(p.Z)
		if p.Y < spring {
			return az >= R && az <= R+2
		}
		r := math.Hypot(p.Z, p.Y-spring)
		return r >= R && r <= R+2
	}}
	// Spandrels between the arches up to the dome base.
	spandrel := Func{Box{V(-R-2, spring, -R-2), V(R+2, domeBase, R+2)}, func(p Vec3) bool {
		ax, az := math.Abs(p.X), math.Abs(p.Z)
		if math.Max(ax, az) < R-0.6 {
			return false
		}
		if az >= R-0.6 && math.Hypot(p.X, p.Y-spring) < R+2 {
			return false
		}
		if ax >= R-0.6 && math.Hypot(p.Z, p.Y-spring) < R+2 {
			return false
		}
		return true
	}}
	g.Use(Terracotta).Fill(archNS, archEW, spandrel)
	g.Paint(Sand, Func{Box{V(-R-2, spring, -R-2), V(R+2, domeBase-1, R+2)}, func(p Vec3) bool {
		ax, az := math.Abs(p.X), math.Abs(p.Z)
		rNS, rEW := math.Hypot(p.X, p.Y-spring), math.Hypot(p.Z, p.Y-spring)
		return (az > R+1.3 && rNS > R+1.4 && rNS <= R+2) || (ax > R+0.7 && rEW > R+1.4 && rEW <= R+2)
	}})
	// Square platform under the drum, leaded on top.
	g.Fill(Subtract(Block(V(-R-2, domeBase-1, -R-2), V(R+2, domeBase, R+2)), Cylinder{Center: o, Radius: R - 1, Y0: domeBase - 2, Y1: domeBase + 1}))
	g.Paint(Slate, Block(V(-R-2, domeBase-0.4, -R-2), V(R+2, domeBase, R+2)))
	// Corner turrets on the four great piers.
	for _, xs := range []float64{-1, 1} {
		for _, zs := range []float64{-1, 1} {
			c := P(xs*(R+0.6), zs*(R+0.6))
			g.Use(Terracotta).Fill(BlockAt(c, 3.4, 3.4, spring, domeBase+3))
			g.Carve(Arch{Sill: c.At(domeBase + 0.2), HalfWidth: 0.6, Spring: 1.4, Depth: 4, Through: AxisX},
				Arch{Sill: c.At(domeBase + 0.2), HalfWidth: 0.6, Spring: 1.4, Depth: 4, Through: AxisZ})
			g.Paint(Sand, BlockAt(c, 3.6, 3.6, domeBase+2.6, domeBase+3))
			g.Use(Slate).Fill(HipRoof{Min: c.Sub(P(1.9, 1.9)), Max: c.Add(P(1.9, 1.9)), Base: domeBase + 3})
			g.With(Gold, func() { g.Fill(Cylinder{Center: c, Radius: 0.4, Y0: domeBase + 4.5, Y1: domeBase + 5.6}) })
		}
	}
	// Drum: forty arched windows between forty little buttresses.
	g.Use(Terracotta).FillVolume(Cylinder{Center: o, Radius: R + 0.4, Inner: R - 0.8, Y0: domeBase, Y1: drumTop})
	for k := 0; k < 40; k++ {
		a := float64(k) * 2 * math.Pi / 40
		g.Carve(h.radialArch(o, R-0.2, a+math.Pi/40, domeBase+0.6, 0.55, 2.4, 3))
		g.Fill(RotateY(Func{Box{V(R, domeBase, -0.5), V(R+1.5, drumTop, 0.5)}, func(p Vec3) bool {
			return p.Y <= drumTop-(p.X-R)*1.1
		}}, o, a))
	}
	g.Paint(Sand, Cylinder{Center: o, Radius: R + 2, Y0: drumTop - 0.5, Y1: drumTop})
	g.Paint(Sand, Cylinder{Center: o, Radius: R + 2, Y0: domeBase, Y1: domeBase + 0.4})
	// The shallow lead dome.
	domeC := o.At(drumTop - 0.3)
	outer, inner := Ellipsoid{domeC, V(R+0.6, 8.6, R+0.6)}, Ellipsoid{domeC, V(R-0.6, 7.4, R-0.6)}
	g.Use(Slate).FillVolume(Func{outer.Bounds(), func(p Vec3) bool {
		return p.Y >= domeC.Y && outer.Contains(p) && !inner.Contains(p)
	}})
	top := domeC.Y + 8.6
	g.With(Gold, func() {
		g.Fill(Cylinder{Center: o, Radius: 0.5, Y0: top - 0.6, Y1: top + 1.6})
		g.FillVolume(Sphere(o.At(top+0.8), 0.75))
		// Crescent.
		g.Fill(Func{Box{V(-1.5, top+1.5, -0.45), V(1.5, top+4.6, 0.45)}, func(p Vec3) bool {
			d := math.Hypot(p.X, p.Y-(top+3))
			return d <= 1.45 && d >= 0.7 && math.Hypot(p.X-0.7, p.Y-(top+3.2)) > 1.05
		}})
	})

	// Half-domes east and west on windowed ring walls, each with two exedrae; the apse
	// semi-dome to the east.
	for _, s := range []float64{-1, 1} {
		hc := P(s*R, 0)
		base := 0.0
		if s < 0 {
			base = math.Pi
		}
		axis := P(s, 0)
		// Ring wall.
		g.Use(Terracotta).Fill(Func{Box{V(math.Min(s*R, s*(2*R+1.4)), roof-1, -R-1.4), V(math.Max(s*R, s*(2*R+1.4)), spring+0.6, R+1.4)}, func(p Vec3) bool {
			d := p.Plan().Sub(hc)
			l := d.Length()
			return d.Dot(axis) >= 0 && l >= R-0.2 && l <= R+1.2
		}})
		g.Paint(Sand, Func{Box{V(math.Min(s*R, s*(2*R+1.4)), spring, -R-1.4), V(math.Max(s*R, s*(2*R+1.4)), spring+0.6, R+1.4)}, func(p Vec3) bool {
			return p.Plan().Sub(hc).Dot(axis) >= 0
		}})
		angles := []float64{-1.18, -0.39, 0.39, 1.18}
		if s < 0 {
			angles = append(angles, 0)
		}
		for _, a := range angles {
			g.Carve(h.radialArch(hc, R+0.5, base+a, roof+0.9, 0.8, 2.4, 3))
		}
		// Half-dome.
		hd := hc.At(spring)
		g.Use(Slate).FillVolume(h.quarterShell(hd, axis, R+0.6, 1.4))
		// A row of five windows at its base, separated by little buttresses.
		for k := -2; k <= 2; k++ {
			g.Carve(h.radialArch(hc, R-0.1, base+float64(k)*0.5, spring+0.6, 0.7, 1.5, 3.2))
		}
		for k := -2; k < 2; k++ {
			rib := Func{Box{V(R+0.2, spring, -0.45), V(R+1.5, spring+3, 0.45)}, func(p Vec3) bool {
				return p.Y <= spring+3-(p.X-R-0.2)*1.3
			}}
			g.Use(Terracotta).Fill(Translate(RotateY(rib, P(0, 0), base+(float64(k)+0.5)*0.5), hc.At(0)))
		}
		// Exedrae at ±45°, springing from the gallery roof.
		for _, zs := range []float64{-1, 1} {
			dir := P(s, zs).Scale(1 / math.Sqrt2)
			ec := hc.Add(dir.Scale(R - 1))
			c := ec.At(roof)
			g.Carve(Func{Box{c.Sub(V(4.8, 0, 4.8)), c.Add(V(4.8, 4.8, 4.8))}, func(p Vec3) bool {
				return p.Sub(c).Length() < 4.8 && p.Plan().Sub(ec).Dot(dir) >= -1
			}})
			g.Use(Slate).FillVolume(h.quarterShell(c, dir, 4.8, 1.2))
			g.Paint(Sand, Func{Box{c.Sub(V(5, 0, 5)), c.Add(V(5, 0.5, 5))}, func(p Vec3) bool {
				return p.Plan().Sub(ec).Dot(dir) >= 0 && p.Sub(c).Length() <= 5
			}})
		}
		if s > 0 {
			c := P(2*R, 0).At(roof)
			g.Carve(Func{Box{c.Sub(V(5.8, 0, 5.8)), c.Add(V(5.8, 5.8, 5.8))}, func(p Vec3) bool {
				return p.Sub(c).Length() < 5.8 && p.X >= 2*R-1
			}})
			g.Use(Slate).FillVolume(h.quarterShell(c, axis, 6, 1.3))
		}
	}
}

// buttresses builds the four great north and south buttress towers, pierced by
// flying arches over the gallery roofs.
func (h HagiaSophia) buttresses(g *Grid) {
	const (
		R, floor, roof = hagiaSophiaR, hagiaSophiaFloor, hagiaSophiaRoof
		bz, hw         = hagiaSophiaBZ, 2.9
		topIn, topOut  = 26.0, 19.5
	)
	for _, xs := range []float64{-1, 1} {
		for _, zs := range []float64{-1, 1} {
			x := xs * R
			z0, z1 := zs*(R+1.2), zs*(bz+5)
			lo, hi := math.Min(z0, z1), math.Max(z0, z1)
			// The lead-covered top slopes down and outward.
			slope := func(z float64) float64 {
				return topIn - (topIn-topOut)*(math.Abs(z)-R-1.2)/(bz+5-R-1.2)
			}
			g.Use(Terracotta).Fill(h.shellBox(x-hw, lo, x+hw, hi, floor, topIn, 1.2))
			g.Carve(Func{Box{V(x-hw-1, topOut-1, lo-1), V(x+hw+1, topIn+1, hi+1)}, func(p Vec3) bool { return p.Y > slope(p.Z) }})
			g.Carve(Arch{Sill: V(x, roof+1, zs*18.4), HalfWidth: 2.6, Spring: 3, Depth: 9, Through: AxisX})
			// Windows on the outer face and sides.
			for _, y := range []float64{1.6, 7.4, 13.6} {
				g.Paint(Sand, Arch{Sill: V(x, y-0.4, zs*(bz+4.6)), HalfWidth: 1.35, Spring: 2.8, Depth: 1.2, Through: AxisZ})
				g.Carve(Arch{Sill: V(x, y, zs*(bz+4.4)), HalfWidth: 0.9, Spring: 2.4, Depth: 2.4, Through: AxisZ})
			}
			g.Carve(Arch{Sill: V(x, 16.4, zs*(bz+0.8)), HalfWidth: 0.8, Spring: 2, Depth: 9, Through: AxisX})
			for _, y := range []float64{1.6, 7.4} {
				for _, xf := range []float64{x - hw + 0.6, x + hw - 0.6} {
					g.Paint(Sand, Arch{Sill: V(xf, y-0.4, zs*(bz+2.6)), HalfWidth: 1.3, Spring: 2.8, Depth: 1.4, Through: AxisX})
				}
				g.Carve(Arch{Sill: V(x, y, zs*(bz+2.6)), HalfWidth: 0.9, Spring: 2.4, Depth: 9, Through: AxisX})
			}
			g.Paint(Sand, Func{Box{V(x-hw-0.1, topOut-1, lo-0.1), V(x+hw+0.1, topIn, hi+0.1)}, func(p Vec3) bool {
				return p.Y > slope(p.Z)-0.6
			}})
			g.Paint(Sand, h.shellBox(x-hw-0.1, lo-0.1, x+hw+0.1, hi+0.1, 6.2, 6.6, 0.6))
			g.Use(Slate).Fill(Func{Box{V(x-hw-0.4, topOut-0.5, lo-0.4), V(x+hw+0.4, topIn+1, hi+0.4)}, func(p Vec3) bool {
				s := slope(p.Z)
				return p.Y > s-0.1 && p.Y <= s+0.8
			}})
		}
	}
}

// narthex builds the inner narthex and the arcaded exonarthex on the west front.
func (h HagiaSophia) narthex(g *Grid) {
	const (
		floor, bx = hagiaSophiaFloor, hagiaSophiaBX
		nz        = 22.0
	)
	inner0, exo0 := -bx-8.0, -bx-11.5
	g.Use(Terracotta).Fill(h.shellBox(inner0, -nz, -bx+0.5, nz, floor, 13.5, 1.2))
	g.Use(Slate).Fill(Func{Box{V(inner0-0.3, 13.5, -nz-0.3), V(-bx+0.5, 15.5, nz+0.3)}, func(p Vec3) bool {
		top := 13.5 + 1.8*(p.X-inner0)/(-bx-inner0)
		return p.Y <= top && p.Y >= top-0.8
	}})
	g.Paint(Sand, h.shellBox(inner0-0.2, -nz-0.2, -bx+0.5, nz+0.2, 12.9, 13.5, 0.6))
	for z := -19.2; z <= 19.3; z += 3.2 {
		g.Paint(Sand, Arch{Sill: V(inner0, 8.6, z), HalfWidth: 1.25, Spring: 2.8, Depth: 0.6, Through: AxisX})
		g.Carve(Arch{Sill: V(inner0+0.6, 9, z), HalfWidth: 0.85, Spring: 2.4, Depth: 2.6, Through: AxisX})
	}
	for _, zs := range []float64{-1, 1} {
		g.Carve(Arch{Sill: V(inner0+4, floor, zs*nz), HalfWidth: 1.4, Spring: 3.4, Depth: 3, Through: AxisZ})
	}
	// Exonarthex: lower arcade with buttress piers.
	g.Use(Terracotta).Fill(h.shellBox(exo0, -nz+1, inner0+0.5, nz-1, floor, 8, 1))
	g.Use(Slate).Fill(Func{Box{V(exo0-0.3, 8, -nz+0.7), V(inner0+0.5, 9.4, nz-0.7)}, func(p Vec3) bool {
		top := 8 + 1.2*(p.X-exo0)/(inner0-exo0)
		return p.Y <= top && p.Y >= top-0.8
	}})
	g.Paint(Sand, h.shellBox(exo0-0.2, -nz+0.8, inner0+0.5, nz-0.8, 7.4, 8, 0.6))
	for k := -6; k <= 6; k++ {
		z := float64(k) * 3.2
		hw, sp := 1.1, 2.8
		if k == 0 {
			hw, sp = 1.5, 3.4
		}
		g.Carve(Arch{Sill: V(exo0+0.5, floor, z), HalfWidth: hw, Spring: sp, Depth: 2.4, Through: AxisX})
		if k < 6 {
			g.Use(Terracotta).Fill(h.spur(P(exo0, z+1.6), P(-1, 0), 1, 1.6, 7))
		}
	}
	// Bronze-dark doors at the back of the central bay.
	g.Use(Wood).Fill(Block(V(inner0-0.2, floor, -1.2), V(inner0+0.4, 4, 1.2)))
}

// minaret is a slender shaft on a square pedestal with corbelled balconies and a
// pencil-point lead cap.
func (HagiaSophia) minaret(g *Grid, c Vec2, r float64, balconies []float64, shaftTop float64, shaft Color) {
	const floor = hagiaSophiaFloor
	g.Use(shaft).Fill(BlockAt(c, 2*r+2.2, 2*r+2.2, floor, 7.5))
	g.Use(Sand).Fill(BlockAt(c, 2*r+2.5, 2*r+2.5, 7.1, 7.6))
	g.Use(shaft).FillVolume(Frustum{c, r + 0.6, r, 7.5, 9.5}, Cylinder{Center: c, Radius: r, Y0: 9, Y1: shaftTop})
	for i, b := range balconies {
		g.Use(White).FillVolume(Frustum{c, r, r + 1.1, b - 1.4, b})
		g.FillVolume(Cylinder{Center: c, Radius: r + 1.1, Inner: r + 0.35, Y0: b, Y1: b + 0.9})
		if i == len(balconies)-1 {
			// The upper shaft above the last balcony is slightly slimmer.
			g.Use(shaft).Carve(Cylinder{Center: c, Radius: r + 0.1, Inner: r - 0.3, Y0: b + 0.9, Y1: shaftTop})
		}
	}
	g.Use(Slate).FillVolume(Frustum{c, r + 0.25, 0.15, shaftTop, shaftTop + 8.5})
	g.Use(Gold).Fill(Cylinder{Center: c, Radius: 0.4, Y0: shaftTop + 7.5, Y1: shaftTop + 10})
	g.FillVolume(Sphere(c.At(shaftTop+8.6), 0.55))
}

// minarets places the four minarets: the two identical western ones with two
// balconies, the slender north-east one and the red brick south-east one.
func (h HagiaSophia) minarets(g *Grid) {
	h.minaret(g, P(-34.5, -27), 1.55, []float64{25, 31.5}, 36, White)
	h.minaret(g, P(-34.5, 27), 1.55, []float64{25, 31.5}, 36, White)
	h.minaret(g, P(30, -27.5), 1.25, []float64{29}, 34.5, White)
	h.minaret(g, P(31, 27), 1.45, []float64{27}, 33.5, Brick)
}

// tree is a plane tree: a trunk and a hollow rounded canopy.
func (HagiaSophia) tree(g *Grid, c Vec2, seed int) {
	const floor = hagiaSophiaFloor
	h := 4.5 + 1.5*Hash01(seed, 1)
	r := 2.6 + 0.8*Hash01(seed, 2)
	g.With(Wood, func() { g.Fill(Cylinder{Center: c, Radius: 0.5, Y0: floor, Y1: h}) })
	color := Foliage
	if Hash01(seed, 3) < 0.3 {
		color = Grass
	}
	g.With(color, func() {
		ctr := c.At(h + 0.6)
		g.FillVolume(Subtract(Ellipsoid{ctr, V(r, r*0.8, r)}, Ellipsoid{ctr, V(r-1.1, r*0.8-1.1, r-1.1)}))
	})
}

// palm is a date palm: a leaning trunk and drooping fronds.
func (HagiaSophia) palm(g *Grid, c Vec2, seed int) {
	const floor = hagiaSophiaFloor
	h := 6 + 1.5*Hash01(seed, 5)
	lean := Polar(P(0, 0), 0.8, Hash01(seed, 6)*2*math.Pi)
	top := c.Add(lean).At(floor + h)
	g.With(Wood, func() { g.Fill(Capsule{A: c.At(floor), B: top, Radius: 0.45}) })
	g.With(Grass, func() {
		g.FillVolume(Sphere(top, 0.9))
		for k := 0; k < 7; k++ {
			a := (float64(k) + Hash01(seed, k)) * 2 * math.Pi / 7
			mid := Polar(top.Plan(), 1.8, a).At(top.Y + 0.5)
			tip := Polar(top.Plan(), 3.2, a).At(top.Y - 1)
			g.Fill(Capsule{A: top, B: mid, Radius: 0.5}, Capsule{A: mid, B: tip, Radius: 0.45})
		}
	})
}

// cypress is a tall narrow evergreen.
func (HagiaSophia) cypress(g *Grid, c Vec2, seed int) {
	const floor = hagiaSophiaFloor
	h := 3.2 + 0.8*Hash01(seed, 9)
	g.With(Foliage, func() { g.FillVolume(Ellipsoid{c.At(floor + h), V(0.95, h, 0.95)}) })
}

// courtyard is the walled garden west of the exonarthex with the domed ablution
// fountain (şadırvan) under its broad eaves.
func (h HagiaSophia) courtyard(g *Grid) {
	const floor = hagiaSophiaFloor
	const x0, x1, z0, z1 = -56.0, -38.0, -17.0, 17.0
	g.Use(Terracotta).Fill(h.shellBox(x0, z0, x1, z1, floor, 2.8, 0.8))
	g.Carve(Block(V(x1-1, floor, -2.5), V(x1+1, 3, 2.5)))
	// Wall top: remove the roof slab of the shell, cap with Sand.
	g.Carve(Block(V(x0+0.8, 1.5, z0+0.8), V(x1-0.8, 3, z1-0.8)))
	g.Paint(Sand, Block(V(x0, 2.4, z0), V(x1, 2.8, z1)))
	g.Carve(Arch{Sill: V(x0+0.4, floor, 0), HalfWidth: 1.6, Spring: 2.2, Depth: 2, Through: AxisX})
	g.Use(Terracotta).Fill(BlockAt(P(x0+0.4, 0), 1.6, 5.6, floor, 5.2))
	g.Carve(Arch{Sill: V(x0+0.4, floor, 0), HalfWidth: 1.6, Spring: 2.2, Depth: 2, Through: AxisX})
	g.Use(Slate).Fill(Gable{Min: P(x0-0.6, -3), Max: P(x0+1.4, 3), Base: 5.2, Ridge: AxisZ})
	// Paving and garden beds.
	g.Paint(Sand, Block(V(x0+0.8, 0, z0+0.8), V(x1-0.8, floor, z1-0.8)))
	for _, zs := range []float64{-1, 1} {
		for _, xs := range []float64{x0 + 4.5, x1 - 4.5} {
			bed := BlockAt(P(xs, zs*11), 6, 7, 0, floor)
			g.Paint(Grass, bed)
			h.tree(g, P(xs, zs*11), int(xs*10+zs*3))
		}
	}
	// Şadırvan: octagonal basin, eight columns, broad leaded eaves and a small dome.
	c := P((x0+x1)/2, 0)
	octagon := func(r, y0, y1 float64) Func {
		return Func{Box{V(c.X-r, y0, -r), V(c.X+r, y1, r)}, func(p Vec3) bool {
			d := p.Plan().Sub(c)
			a := math.Mod(math.Atan2(d.Y, d.X)+2*math.Pi, math.Pi/4) - math.Pi/8
			return d.Length()*math.Cos(a) <= r
		}}
	}
	g.Use(White).Fill(Subtract(octagon(3.4, floor, 1.4), octagon(2.6, 0.9, 1.5)))
	g.Paint(Water, octagon(2.6, 0, 1.0))
	g.Fill(Cylinder{Center: c, Radius: 0.6, Y0: 1, Y1: 3})
	for k := 0; k < 8; k++ {
		at := Polar(c, 3.8, float64(k)*math.Pi/4+math.Pi/8)
		g.Use(White).Fill(Cylinder{Center: at, Radius: 0.42, Y0: floor, Y1: 5})
	}
	g.Use(Slate).FillVolume(Subtract(Frustum{c, 6.2, 4.2, 5, 6.3}, Frustum{c, 5.2, 3.2, 5, 5.9}))
	g.Use(Terracotta).Fill(Cylinder{Center: c, Radius: 4.4, Inner: 3.4, Y0: 5.6, Y1: 6.6})
	g.Use(Slate).FillVolume(Intersect(Subtract(Ellipsoid{c.At(6.6), V(3.6, 2.8, 3.6)}, Ellipsoid{c.At(6.6), V(2.6, 1.8, 2.6)}),
		Block(V(c.X-4, 6.6, -4), V(c.X+4, 9.6, 4))))
	g.Use(Gold).Fill(Cylinder{Center: c, Radius: 0.35, Y0: 9, Y1: 10.6})
}

// turbes are the domed Ottoman mausolea in a walled garden south of the building.
func (h HagiaSophia) turbes(g *Grid) {
	const floor = hagiaSophiaFloor
	polygon := func(c Vec2, n int, r, y0, y1 float64) Func {
		return Func{Box{V(c.X-r, y0, c.Y-r), V(c.X+r, y1, c.Y+r)}, func(p Vec3) bool {
			d := p.Plan().Sub(c)
			step := 2 * math.Pi / float64(n)
			a := math.Mod(math.Atan2(d.Y, d.X)+2*math.Pi, step) - step/2
			return d.Length()*math.Cos(a) <= r
		}}
	}
	turbe := func(c Vec2, n int, r, wall float64, seed int) {
		g.Use(Stone).Fill(polygon(c, n, r+0.6, floor, 1.2))
		g.Fill(Subtract(polygon(c, n, r, 1.2, wall), polygon(c, n, r-1.1, 1.2, wall-0.8)))
		g.Paint(Sand, polygon(c, n, r+0.1, wall-0.5, wall))
		g.Paint(Sand, polygon(c, n, r+0.1, 4.6, 5.0))
		// Two rows of windows on every face.
		for k := 0; k < n; k++ {
			a := (float64(k) + 0.5) * 2 * math.Pi / float64(n)
			a = math.Atan2(math.Sin(a), math.Cos(a))
			g.Carve(h.radialArch(c, r-0.5, a, 1.8, 0.6, 1.9, 2.4), h.radialArch(c, r-0.5, a, 5.6, 0.55, 1.4, 2.4))
		}
		// Portico on the north face.
		pc := c.Add(P(0, -r-1.6))
		g.Use(Stone).Fill(BlockAt(pc, 2*r, 3.4, floor, 1.2))
		for _, dx := range []float64{-r + 0.8, -r/3 + 0.2, r/3 - 0.2, r - 0.8} {
			g.Use(White).Fill(Cylinder{Center: pc.Add(P(dx, -1)), Radius: 0.4, Y0: 1.2, Y1: 4.4})
		}
		g.Use(Slate).Fill(Func{Box{V(pc.X-r-0.3, 4.4, pc.Y-2.2), V(pc.X+r+0.3, 5.6, pc.Y+1.8)}, func(p Vec3) bool {
			return p.Y <= 5.6-0.35*(pc.Y+1.8-p.Z)
		}})
		// Drum and lead dome with a gold finial.
		g.Use(Stone).FillVolume(Cylinder{Center: c, Radius: r - 0.6, Inner: r - 1.7, Y0: wall - 0.6, Y1: wall + 1})
		dc := c.At(wall + 0.8)
		dr := r - 0.4
		g.Use(Slate).FillVolume(Func{Box{dc.Sub(V(dr, 0, dr)), dc.Add(V(dr, dr, dr))}, func(p Vec3) bool {
			l := p.Sub(dc).Length()
			return p.Y >= dc.Y && l <= dr && l >= dr-1.1
		}})
		g.Use(Gold).Fill(Cylinder{Center: c, Radius: 0.35, Y0: dc.Y + dr - 0.4, Y1: dc.Y + dr + 1.4})
		g.FillVolume(Sphere(c.At(dc.Y+dr+0.7), 0.5))
		_ = seed
	}
	for i, t := range hagiaSophiaTurbes {
		turbe(t.c, t.n, t.r, t.wall, i)
	}
	// Garden wall with cypresses.
	const x0, x1, z0, z1 = 0.0, 39.0, 33.4, 49.0
	g.Use(Stone).Fill(h.shellBox(x0, z0, x1, z1, floor, 2.2, 0.7))
	g.Carve(Block(V(x0+0.7, 1, z0+0.7), V(x1-0.7, 3, z1-0.7)))
	g.Paint(Sand, Block(V(x0, 1.8, z0), V(x1, 2.2, z1)))
	for _, x := range []float64{11.5, 22.5} {
		g.Carve(Block(V(x-1.1, floor, z0-1), V(x+1.1, 3, z0+1)))
	}
	seed := 40
	for x := 2.5; x < x1-1; x += 3.4 {
		if math.Abs(x-11.5) < 2 || math.Abs(x-22.5) < 2 {
			continue
		}
		seed++
		h.cypress(g, P(x, z1-2.2), seed)
	}
	for _, z := range []float64{37, 41, 45} {
		seed++
		h.cypress(g, P(x0+2.2, z), seed)
		h.cypress(g, P(x1-2.2, z), seed+100)
	}
}

// park is the corner of Sultanahmet park: lawns, a round fountain with jets, plane
// trees and palms, plus street trees along the north side and around the site.
func (h HagiaSophia) park(g *Grid) {
	const floor = hagiaSophiaFloor
	fc := P(-31, 41)
	g.Use(White).FillVolume(Cylinder{Center: fc, Radius: 5.6, Inner: 4.8, Y0: floor, Y1: 1.3})
	g.Paint(Water, Cylinder{Center: fc, Radius: 4.8, Y0: 0, Y1: floor})
	g.Use(Water).Fill(Cylinder{Center: fc, Radius: 4.8, Y0: floor, Y1: 0.8})
	g.Use(White).Fill(Cylinder{Center: fc, Radius: 1.2, Y0: floor, Y1: 1.6}, Cylinder{Center: fc, Radius: 0.45, Y0: 1.6, Y1: 3})
	g.Use(White).Fill(Cylinder{Center: fc, Radius: 0.5, Y0: 3, Y1: 4.4})
	for k := 0; k < 8; k++ {
		at := Polar(fc, 3.4, float64(k)*math.Pi/4)
		g.Use(White).Fill(Capsule{A: at.At(0.8), B: Polar(fc, 2.6, float64(k)*math.Pi/4).At(2.2), Radius: 0.4})
	}
	// Plane trees and palms around the park.
	seed := 200
	for _, p := range []Vec2{P(-52, 35), P(-52, 46), P(-43, 35), P(-43, 46), P(-19, 35), P(-19, 46), P(-11, 36), P(-11, 46)} {
		seed++
		h.tree(g, p, seed)
	}
	for _, p := range []Vec2{P(-37, 35.5), P(-25, 35.5), P(-37, 46.5), P(-25, 46.5)} {
		seed++
		h.palm(g, p, seed)
	}
	// Street trees along the north side and on the east lawn.
	for x := -54.0; x <= 36; x += 6.5 {
		seed++
		h.tree(g, P(x+Hash01(seed, 4)-0.5, -33+2.5*Hash01(seed, 7)+0.5), seed)
	}
	for z := -22.0; z <= 24; z += 6.5 {
		seed++
		if math.Abs(z-27) < 4 || math.Abs(z+27) < 4 {
			continue
		}
		if Hash01(seed, 8) < 0.5 {
			h.palm(g, P(37, z), seed)
		} else {
			h.tree(g, P(37, z), seed)
		}
	}
	// Palms on the south plaza in front of the türbe garden and by the courtyard.
	for _, p := range []Vec2{P(-6, 31), P(-16, 29), P(-48, 22), P(-48, -22), P(36, 31)} {
		seed++
		h.palm(g, p, seed)
	}
}
