package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// BlueMosque is the Sultan Ahmed Mosque in Istanbul, whose subject is its cascade of
// lead domes: the great central dome on a windowed drum, four semi-domes around it,
// two exedra half-domes beside each semi-dome and four corner domes, all held by
// four massive octagonal pier turrets. The grey stone prayer hall has three rows of
// pointed windows and arcaded side galleries. Six pencil minarets stand at the hall
// corners and the far corners of the forecourt, whose portico carries a ring of small
// domes around the hexagonal ablution fountain. In front of the gate lies a corner of
// Sultanahmet park with plane trees and the round fountain pool; the site slopes
// gently down to a low rim. The courtyard and gate face +Z; the qibla wall is at -Z.
type BlueMosque struct{}

func init() { Register(BlueMosque{}) }

func (BlueMosque) Name() string  { return "blue_mosque" }
func (BlueMosque) Title() string { return "Blue Mosque" }

// Main dimensions shared by the helpers.
const (
	blueMosqueR        = 9.0                             // Central dome radius and half-width of the central square.
	blueMosqueB        = 20.0                            // Prayer hall half-width.
	blueMosqueFloor    = 1.2                             // Top of the precinct ground.
	blueMosqueRoof     = 13.2                            // Hall roof level.
	blueMosqueSpring   = 16.2                            // Spring of the semi-domes and great arches.
	blueMosqueDomeBase = blueMosqueSpring + blueMosqueR  // Base of the central drum.
	blueMosqueDrumTop  = blueMosqueDomeBase + 2.6        // Top of the drum.
	blueMosqueTurret   = blueMosqueDomeBase + 3.6        // Top of the pier turrets' walls.
	blueMosqueExA      = 0.80                            // Exedra angle from each semi-dome axis.
	blueMosqueExR      = 3.6                             // Exedra radius.
	blueMosqueCornerR  = 4.2                             // Corner dome radius.
	blueMosqueM        = 3.65                            // Courtyard bay module.
	blueMosqueCX       = 5.5 * blueMosqueM               // Courtyard half-width.
	blueMosqueCZ0      = blueMosqueB                     // Courtyard, mosque side.
	blueMosqueCZ1      = blueMosqueCZ0 + 8*blueMosqueM   // Courtyard, gate side.
	blueMosquePH       = blueMosqueFloor + 5.6           // Courtyard portico roof top.
	blueMosqueSX       = 30.0                            // Site half-width.
	blueMosqueSZ0      = -29.5                           // Site back edge.
	blueMosqueSZ1      = 68.0                            // Site front edge.
	blueMosqueCornerC  = (blueMosqueR + blueMosqueB) / 2 // Corner dome centre offset.
)

// blueMosqueAxes are the four semi-dome directions.
var blueMosqueAxes = []Vec2{P(1, 0), P(-1, 0), P(0, 1), P(0, -1)}

// height is the ground surface: flat precinct sloping smoothly to a low rim.
func (BlueMosque) height(x, z float64) float64 {
	e := math.Min(math.Min(x+blueMosqueSX, blueMosqueSX-x), math.Min(z-blueMosqueSZ0, blueMosqueSZ1-z))
	t := math.Max(0, math.Min(1, e/5))
	t = t * t * (3 - 2*t)
	return 0.3 + (blueMosqueFloor-0.3)*t
}

// shellBox is a hollow rectangular box with walls of thickness t and a roof slab.
func (BlueMosque) shellBox(x0, z0, x1, z1, y0, y1, t float64) Solid {
	return Subtract(Block(V(x0, y0, z0), V(x1, y1, z1)), Block(V(x0+t, y0-1, z0+t), V(x1-t, y1-t, z1-t)))
}

// pointed is a two-centred Ottoman arch opening: jambs up to spring, then two arcs.
func (BlueMosque) pointed(sill Vec3, hw, spring, depth float64, through Axis) Solid {
	r := hw * 1.4
	c := r - hw
	apex := math.Sqrt(r*r - c*c)
	span := V(hw, 0, depth/2)
	if through == AxisX {
		span = V(depth/2, 0, hw)
	}
	return Func{Box{sill.Sub(span), sill.Add(span).Add(V(0, spring+apex, 0))}, func(p Vec3) bool {
		d := p.Sub(sill)
		across := d.X
		if through == AxisX {
			across = d.Z
		}
		if math.Abs(across) > hw || d.Y < 0 {
			return false
		}
		rise := d.Y - spring
		if rise <= 0 {
			return true
		}
		a := math.Abs(across) + c
		return a*a+rise*rise <= r*r
	}}
}

// radial is a pointed opening through a curved wall around center, at radius r and
// plan angle a.
func (m BlueMosque) radial(center Vec2, r, a, sill, hw, spring, depth float64) Solid {
	return Translate(RotateY(m.pointed(V(r, sill, 0), hw, spring, depth, AxisX), P(0, 0), a), center.At(0))
}

// polygon is a regular n-gon prism with apothem r.
func (BlueMosque) polygon(c Vec2, n int, r, y0, y1, phase float64) Func {
	R := r / math.Cos(math.Pi/float64(n))
	return Func{Box{V(c.X-R, y0, c.Y-R), V(c.X+R, y1, c.Y+R)}, func(p Vec3) bool {
		d := p.Plan().Sub(c)
		step := 2 * math.Pi / float64(n)
		a := math.Mod(math.Atan2(d.Y, d.X)-phase+4*math.Pi, step) - step/2
		return d.Length()*math.Cos(a) <= r
	}}
}

// cap is the upper half of a spherical (or squashed) shell centred at c.
func (BlueMosque) cap(c Vec3, r, h, t float64) Solid {
	outer := Ellipsoid{c, V(r, h, r)}
	inner := Ellipsoid{c, V(r-t, h-t, r-t)}
	return Func{outer.Bounds(), func(p Vec3) bool {
		return p.Y >= c.Y && outer.Contains(p) && (t <= 0 || !inner.Contains(p))
	}}
}

// quarterShell is the outer half (dot(p-c, dir) >= 0, above c.Y) of a spherical shell.
func (BlueMosque) quarterShell(c Vec3, dir Vec2, r, t float64) Solid {
	return Func{Box{c.Sub(V(r, 0, r)), c.Add(V(r, r, r))}, func(p Vec3) bool {
		d := p.Sub(c)
		if d.Y < 0 || d.Plan().Dot(dir) < 0 {
			return false
		}
		l := d.Length()
		return l <= r && l >= r-t
	}}
}

// exedrae returns the centre and direction of the two exedrae beside a semi-dome.
func (BlueMosque) exedrae(axis Vec2) [2][2]Vec2 {
	hc := axis.Scale(blueMosqueR)
	var out [2][2]Vec2
	for i, s := range []float64{-1, 1} {
		dir := axis.Rotate(s * blueMosqueExA)
		out[i] = [2]Vec2{hc.Add(dir.Scale(blueMosqueR)), dir}
	}
	return out
}

// corners lists the four corner dome centres.
func (BlueMosque) corners() []Vec2 {
	const c = blueMosqueCornerC
	return []Vec2{P(c, c), P(-c, c), P(c, -c), P(-c, -c)}
}

// finial is a gold alem: rod, stacked spheres and a crescent facing +Z.
func (BlueMosque) finial(g *Grid, c Vec2, top, scale float64) {
	g.With(Gold, func() {
		g.Fill(Cylinder{Center: c, Radius: 0.45, Y0: top - 0.6, Y1: top + 2.4*scale})
		g.FillVolume(Sphere(c.At(top+0.9*scale), 0.75*scale))
		if scale >= 1 {
			g.FillVolume(Sphere(c.At(top+2.1*scale), 0.55*scale))
			y := top + 3.6*scale
			g.Fill(Func{Box{V(c.X-1.5, y-1.5, c.Y-0.45), V(c.X+1.5, y+1.5, c.Y+0.45)}, func(p Vec3) bool {
				d := math.Hypot(p.X-c.X, p.Y-y)
				return d <= 1.45 && d >= 0.7 && math.Hypot(p.X-c.X-0.75, p.Y-y-0.2) > 1.05
			}})
		}
	})
}

// interior reports whether a point lies in the enclosed hollow of the prayer hall, its
// domes, galleries or turrets; shadeInterior darkens the faces around these spaces.
func (m BlueMosque) interior(p Vec3) bool {
	const (
		R, B, floor, roof, spring = blueMosqueR, blueMosqueB, blueMosqueFloor, blueMosqueRoof, blueMosqueSpring
		domeBase, drumTop         = blueMosqueDomeBase, blueMosqueDrumTop
	)
	if p.Y <= floor {
		return false
	}
	ax, az := math.Abs(p.X), math.Abs(p.Z)
	if ax < B-1.5 && az < B-1.5 && p.Y < roof-1.5 {
		return true
	}
	if m.turbeInterior(p) {
		return true
	}
	if p.Y < roof-1.6 {
		return false
	}
	pl := p.Plan()
	if ax < R-0.6 && az < R-0.6 && p.Y < domeBase {
		return true
	}
	r := pl.Length()
	if r < R-0.8 && p.Y < drumTop {
		return true
	}
	dc := V(0, drumTop-0.2, 0)
	if p.Y >= dc.Y && (Ellipsoid{dc, V(R-0.7, 7.3, R-0.7)}).Contains(p) {
		return true
	}
	for _, axis := range blueMosqueAxes {
		d := pl.Sub(axis.Scale(R))
		if d.Dot(axis) >= 0 {
			l := d.Length()
			if (p.Y < spring && l < R-0.2) || math.Hypot(l, p.Y-spring) < R-0.8 {
				return true
			}
		}
		for _, e := range m.exedrae(axis) {
			if p.Sub(e[0].At(roof)).Length() < blueMosqueExR-1 {
				return true
			}
		}
	}
	for _, c := range m.corners() {
		d := pl.Sub(c).Length()
		if (d < blueMosqueCornerR-1 && p.Y < roof+1.6) || p.Sub(c.At(roof+1.6)).Length() < blueMosqueCornerR-1 {
			return true
		}
	}
	for _, s := range [][2]float64{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}} {
		if pl.Sub(P(s[0]*R, s[1]*R)).Length() < 1 && p.Y < blueMosqueTurret-0.6 {
			return true
		}
	}
	return false
}

// shadeInterior paints every face that borders an interior hollow Charcoal, so that
// windows and doors read as dark openings.
func (m BlueMosque) shadeInterior(g *Grid) {
	var dark []Cell
	for _, c := range g.Cells() {
		for _, n := range Neighbors {
			k := c.Add(n)
			if !g.Has(k) && m.interior(g.World(k)) {
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

func (m BlueMosque) Build(g *Grid) Order {
	m.ground(g)
	m.hall(g)
	m.galleries(g)
	m.domes(g)
	m.courtyard(g)
	m.minarets(g)
	m.precinct(g)
	m.turbe(g)
	m.shadeInterior(g)
	m.garden(g)
	return Outward(P(0, 8))
}

// ground is the sloped site: paved precinct, white marble courtyard, park lawns and paths.
func (m BlueMosque) ground(g *Grid) {
	const floor = blueMosqueFloor
	site := Box{V(-blueMosqueSX, 0, blueMosqueSZ0), V(blueMosqueSX, floor, blueMosqueSZ1)}
	g.Use(Grass).Fill(Func{site, func(p Vec3) bool {
		h := m.height(p.X, p.Z)
		return p.Y <= h && p.Y >= h-0.9
	}})
	g.Paint(Sand, Block(V(-blueMosqueB-5, 0, -blueMosqueB-4.5), V(blueMosqueB+5, floor+0.1, blueMosqueCZ1+3.5)))
	g.Paint(Stone, Block(V(-blueMosqueCX+blueMosqueM, 0, blueMosqueCZ0+blueMosqueM), V(blueMosqueCX-blueMosqueM, floor+0.1, blueMosqueCZ1-blueMosqueM)))
	// White marble paving bands in the court.
	g.Paint(White, Func{Box{V(-17, 0, 23.5), V(17, floor+0.1, 46)}, func(p Vec3) bool {
		zc := (blueMosqueCZ0 + blueMosqueCZ1) / 2
		return math.Abs(p.X) < 1.2 || math.Abs(p.Z-zc) < 1.2 || math.Abs(math.Hypot(p.X, p.Z-zc)-6.5) < 0.7
	}})
	// Park paths: the gate axis, the ring round the pool, a cross walk and diagonals.
	fc := P(0, 59.5)
	g.Paint(Sand, Func{Box{V(-blueMosqueSX, 0, 49), site.Max}, func(p Vec3) bool {
		d := p.Plan().Sub(fc)
		l := d.Length()
		diag := math.Abs(math.Abs(d.X)-math.Abs(d.Y)) < 1.4 && l < 14
		return math.Abs(p.X) < 1.5 || math.Abs(l-6.2) < 1 || (math.Abs(d.Y) < 1.1 && math.Abs(p.X) < 25) || diag
	}})
}

// hall is the prayer hall: hollow stone walls with three rows of pointed windows,
// cornices, doors and a lead-covered roof opened beneath the domes.
func (m BlueMosque) hall(g *Grid) {
	const (
		R, B, floor, roof = blueMosqueR, blueMosqueB, blueMosqueFloor, blueMosqueRoof
	)
	g.Use(Stone).Fill(m.shellBox(-B, -B, B, B, floor, roof, 1.5))
	// Open the roof beneath the cascade.
	g.Carve(Func{Box{V(-B, roof-1.8, -B), V(B, roof+0.1, B)}, func(p Vec3) bool {
		pl := p.Plan()
		if math.Max(math.Abs(p.X), math.Abs(p.Z)) < R-0.6 {
			return true
		}
		for _, axis := range blueMosqueAxes {
			d := pl.Sub(axis.Scale(R))
			if d.Dot(axis) >= -0.6 && d.Length() < R-0.2 {
				return true
			}
			for _, e := range m.exedrae(axis) {
				if pl.Sub(e[0]).Length() < blueMosqueExR-1 {
					return true
				}
			}
		}
		for _, c := range m.corners() {
			if pl.Sub(c).Length() < blueMosqueCornerR-1 {
				return true
			}
		}
		return false
	}})
	g.Paint(Slate, Block(V(-B+0.6, roof-0.45, -B+0.6), V(B-0.6, roof+0.1, B-0.6)))
	band := func(y0, y1 float64) Solid {
		return Subtract(Block(V(-B-0.3, y0, -B-0.3), V(B+0.3, y1, B+0.3)), Block(V(-B+1, y0-1, -B+1), V(B-1, y1+1, B-1)))
	}
	g.Use(White).Fill(band(roof-0.6, roof), band(floor+4.6, floor+5.0))
	g.Use(Stone).Fill(band(floor, floor+0.5))
	// Windows: three rows on each face, the lower two framed in white.
	window := func(at Vec3, through Axis, hw, spring float64, frame bool) {
		if frame {
			g.Paint(White, m.pointed(at.Sub(V(0, 0.4, 0)), hw+0.4, spring+0.3, 2.4, through))
		}
		g.Carve(m.pointed(at, hw, spring, 3.2, through))
	}
	for k := -5; k <= 5; k++ {
		x := float64(k) * 3.6
		for _, s := range []float64{-1, 1} {
			for _, f := range []struct {
				x, z float64
				ax   Axis
				side bool
			}{{x, s * (B - 0.6), AxisZ, false}, {s * (B - 0.6), x, AxisX, true}} {
				front := !f.side && s > 0
				if front && k == 0 {
					continue
				}
				if !(f.side && k == 0) {
					window(V(f.x, floor+1.0, f.z), f.ax, 0.85, 2.2, true)
				}
				if !front {
					window(V(f.x, floor+5.5, f.z), f.ax, 0.8, 1.9, true)
				}
				window(V(f.x, floor+8.2, f.z), f.ax, 0.55, 0.6, false)
			}
		}
	}
	// The main portal from the courtyard and the side doors under the galleries.
	g.Paint(White, m.pointed(V(0, floor, B), 2.6, 5.2, 0.8, AxisZ))
	g.Carve(m.pointed(V(0, floor, B-0.3), 1.6, 4.2, 3, AxisZ))
	for _, s := range []float64{-1, 1} {
		g.Carve(m.pointed(V(s*(B-0.6), floor, 0), 1.2, 2.8, 3, AxisX))
	}
	// The mihrab bay projecting from the qibla wall.
	g.Use(Stone).Fill(Block(V(-3.2, floor, -B-2.2), V(3.2, roof-2, -B+0.5)))
	g.Use(White).Fill(Block(V(-3.5, roof-2.5, -B-2.5), V(3.5, roof-2, -B)))
	g.Use(Slate).Fill(HipRoof{Min: P(-3.4, -B-2.4), Max: P(3.4, -B+0.6), Base: roof - 2})
	for _, x := range []float64{-1.6, 1.6} {
		g.Carve(m.pointed(V(x, floor+4, -B-2.2), 0.6, 1.6, 1.2, AxisZ))
	}
}

// galleries are the two-storey arcaded loggias along the east and west flanks.
func (m BlueMosque) galleries(g *Grid) {
	const (
		B, floor = blueMosqueB, blueMosqueFloor
		w, l, h  = 3.6, 16.5, 7.6
	)
	for _, s := range []float64{-1, 1} {
		x0, x1 := math.Min(s*(B-0.5), s*(B+w)), math.Max(s*(B-0.5), s*(B+w))
		g.Use(Stone).Fill(m.shellBox(x0, -l, x1, l, floor, floor+h, 0.9))
		g.Use(White).Fill(Subtract(Block(V(x0-0.25, floor+h-0.5, -l-0.25), V(x1+0.25, floor+h, l+0.25)), Block(V(x0+0.8, 0, -l+0.8), V(x1-0.8, 20, l-0.8))))
		g.Use(White).Fill(Subtract(Block(V(x0-0.2, floor+3.9, -l-0.2), V(x1+0.2, floor+4.3, l+0.2)), Block(V(x0+0.8, 0, -l+0.8), V(x1-0.8, 20, l-0.8))))
		xf := s * (B + w - 0.45)
		for k := -4; k <= 4; k++ {
			z := float64(k) * 3.5
			g.Carve(m.pointed(V(xf, floor, z), 1.1, 1.6, 2, AxisX))
			g.Carve(m.pointed(V(xf, floor+4.5, z), 0.95, 1.2, 2, AxisX))
		}
		for _, zs := range []float64{-1, 1} {
			g.Carve(m.pointed(V(s*(B+w/2), floor, zs*(l-0.45)), 1.0, 1.6, 2, AxisZ))
		}
		// Lead roof with a row of small domes, one over each bay.
		g.Paint(Slate, Block(V(x0+0.5, floor+h-0.45, -l+0.5), V(x1-0.5, floor+h+0.1, l-0.5)))
		for k := -4; k <= 4; k++ {
			c := P(s*(B+w/2-0.2), float64(k)*3.5)
			g.Use(Stone).Fill(Cylinder{Center: c, Radius: 1.45, Y0: floor + h - 0.2, Y1: floor + h + 0.4})
			g.Use(Slate).FillVolume(m.cap(c.At(floor+h+0.4), 1.65, 1.5, 0))
			g.Use(Gold).Fill(Cylinder{Center: c, Radius: 0.4, Y0: floor + h + 1.7, Y1: floor + h + 2.6})
		}
	}
}

// domes builds the cascade: central square of great arches, pier turrets, the
// windowed drum and great dome, semi-domes on windowed ring walls, exedrae and the
// four corner domes.
func (m BlueMosque) domes(g *Grid) {
	const (
		R, roof, spring   = blueMosqueR, blueMosqueRoof, blueMosqueSpring
		domeBase, drumTop = blueMosqueDomeBase, blueMosqueDrumTop
		exR, cR           = blueMosqueExR, blueMosqueCornerR
	)
	o := P(0, 0)
	// Great arches and pendentives: a square wall pierced by the four arches.
	g.Use(Stone).Fill(Func{Box{V(-R-0.6, spring, -R-0.6), V(R+0.6, domeBase, R+0.6)}, func(p Vec3) bool {
		ax, az := math.Abs(p.X), math.Abs(p.Z)
		if math.Max(ax, az) < R-0.6 {
			return false
		}
		if ax >= R-0.6 && math.Hypot(p.Z, p.Y-spring) < R-0.4 {
			return false
		}
		if az >= R-0.6 && math.Hypot(p.X, p.Y-spring) < R-0.4 {
			return false
		}
		return true
	}})
	g.Paint(Slate, Block(V(-R-0.7, domeBase-0.45, -R-0.7), V(R+0.7, domeBase, R+0.7)))

	// Semi-domes with ring walls, exedrae and windows.
	for _, axis := range blueMosqueAxes {
		hc := axis.Scale(R)
		base := axis.Angle()
		lo, hi := hc.Sub(P(R+1.2, R+1.2)), hc.Add(P(R+1.2, R+1.2))
		g.Use(Stone).Fill(Func{Box{lo.At(roof - 1), hi.At(spring + 0.6)}, func(p Vec3) bool {
			d := p.Plan().Sub(hc)
			l := d.Length()
			return d.Dot(axis) >= 0 && l >= R-0.2 && l <= R+1.0
		}})
		g.Paint(White, Func{Box{lo.At(spring), hi.At(spring + 0.6)}, func(p Vec3) bool {
			d := p.Plan().Sub(hc)
			return d.Dot(axis) >= 0 && d.Length() >= R+0.3
		}})
		for _, a := range []float64{0, -0.42, 0.42} {
			g.Carve(m.radial(hc, R+0.4, base+a, roof+0.5, 0.7, 1.2, 3))
		}
		// Exedrae: carve the ring wall, then the half-domes with three windows.
		for _, e := range m.exedrae(axis) {
			ec, dir := e[0], e[1]
			c := ec.At(roof)
			g.Carve(Func{Box{c.Sub(V(exR, 0.2, exR)), c.Add(V(exR, exR, exR))}, func(p Vec3) bool {
				return p.Sub(c).Length() < exR && p.Plan().Sub(ec).Dot(dir) >= -1.2
			}})
			g.Use(Slate).FillVolume(m.quarterShell(c, dir, exR, 1.1))
			g.Use(Stone).FillVolume(Func{Box{c.Sub(V(exR+0.2, 0, exR+0.2)), c.Add(V(exR+0.2, 0.9, exR+0.2))}, func(p Vec3) bool {
				d := p.Plan().Sub(ec)
				l := d.Length()
				return d.Dot(dir) >= 0 && l <= exR+0.2 && l >= exR-1.0
			}})
			da := dir.Angle()
			for _, a := range []float64{-0.75, 0, 0.75} {
				g.Carve(m.radial(ec, exR-0.5, da+a, roof+0.4, 0.5, 0.6, 1.6))
			}
		}
		// The semi-dome shell with a ring of five windows and buttress ribs at its base.
		g.Use(Slate).FillVolume(m.quarterShell(hc.At(spring), axis, R+0.6, 1.3))
		for k := -2; k <= 2; k++ {
			g.Carve(m.radial(hc, R-0.1, base+float64(k)*0.46, spring+0.5, 0.65, 1.0, 3.2))
		}
		for k := -2; k < 2; k++ {
			rib := Func{Box{V(R+0.1, spring, -0.45), V(R+1.6, spring+3, 0.45)}, func(p Vec3) bool {
				return p.Y <= spring+3-(p.X-R-0.1)*1.3
			}}
			g.Use(Stone).Fill(Translate(RotateY(rib, P(0, 0), base+(float64(k)+0.5)*0.46), hc.At(0)))
		}
	}

	// Corner domes on short drums.
	for _, c := range m.corners() {
		g.Use(Stone).FillVolume(Cylinder{Center: c, Radius: cR, Inner: cR - 1, Y0: roof - 0.6, Y1: roof + 1.6})
		for k := 0; k < 8; k++ {
			g.Carve(m.radial(c, cR-0.5, float64(k)*math.Pi/4+math.Pi/8, roof+0.2, 0.5, 0.5, 1.6))
		}
		g.Paint(White, Cylinder{Center: c, Radius: cR + 0.2, Y0: roof + 1.2, Y1: roof + 1.6})
		g.Use(Slate).FillVolume(m.cap(c.At(roof+1.6), cR, cR-0.3, 1.0))
		m.finial(g, c, roof+1.6+cR-0.3, 0.7)
		// A small domed kiosk on the corner of the roof beyond.
		k := c.Scale((blueMosqueB - 1.1) / blueMosqueCornerC)
		g.Use(Stone).Fill(Subtract(m.polygon(k, 8, 1.25, roof-0.6, roof+2.6, math.Pi/8), m.polygon(k, 8, 0.4, roof, roof+2, math.Pi/8)))
		for a := 0; a < 4; a++ {
			g.Carve(m.radial(k, 1, float64(a)*math.Pi/2, roof+0.6, 0.4, 0.8, 1.2))
		}
		g.Use(White).Fill(m.polygon(k, 8, 1.4, roof+2.2, roof+2.6, math.Pi/8))
		g.Use(Slate).FillVolume(m.cap(k.At(roof+2.6), 1.3, 1.4, 0))
		g.Use(Gold).Fill(Cylinder{Center: k, Radius: 0.4, Y0: roof + 3.6, Y1: roof + 4.8})
	}

	// The four octagonal pier turrets, open at the top, with lead domes.
	for _, s := range [][2]float64{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}} {
		c := P(s[0]*R, s[1]*R)
		top := blueMosqueTurret
		g.Use(Stone).Fill(Subtract(m.polygon(c, 8, 2.1, roof-0.6, top, math.Pi/8), m.polygon(c, 8, 1.1, roof, top-0.6, math.Pi/8)))
		for k := 0; k < 8; k++ {
			a := float64(k) * math.Pi / 4
			g.Carve(m.radial(c, 1.7, a, top-3.4, 0.55, 1.4, 1.6))
		}
		g.Paint(White, m.polygon(c, 8, 2.3, top-0.5, top, math.Pi/8))
		g.Paint(White, m.polygon(c, 8, 2.3, domeBase-0.3, domeBase+0.1, math.Pi/8))
		g.Use(White).Fill(m.polygon(c, 8, 2.35, top-0.5, top, math.Pi/8))
		g.Use(Slate).FillVolume(m.cap(c.At(top), 2.0, 2.2, 0))
		m.finial(g, c, top+2.2, 0.8)
	}

	// Drum: twenty-four windows between little buttresses.
	g.Use(Stone).FillVolume(Cylinder{Center: o, Radius: R + 0.5, Inner: R - 0.8, Y0: domeBase - 0.4, Y1: drumTop})
	for k := 0; k < 24; k++ {
		a := float64(k) * 2 * math.Pi / 24
		g.Carve(m.radial(o, R-0.2, a+math.Pi/24, domeBase+0.4, 0.55, 0.9, 3))
		g.Use(Stone).Fill(RotateY(Func{Box{V(R, domeBase-0.4, -0.5), V(R+1.6, drumTop+0.4, 0.5)}, func(p Vec3) bool {
			return p.Y <= drumTop+0.4-(p.X-R)*1.2
		}}, o, a))
	}
	g.Paint(White, Cylinder{Center: o, Radius: R + 2, Y0: drumTop - 0.45, Y1: drumTop})
	// The great dome.
	dc := o.At(drumTop - 0.2)
	g.Use(Slate).FillVolume(m.cap(dc, R+0.5, 8.5, 1.2))
	m.finial(g, o, dc.Y+8.5, 1.2)
}

// courtyard is the forecourt: an outer wall, a portico ring of domed bays round the
// open court, the monumental gate, side gates and the hexagonal şadırvan.
func (m BlueMosque) courtyard(g *Grid) {
	const (
		floor, M, X1, Z0, Z1, ph = blueMosqueFloor, blueMosqueM, blueMosqueCX, blueMosqueCZ0, blueMosqueCZ1, blueMosquePH
	)
	// Outer wall and portico roof.
	g.Use(Stone).Fill(Subtract(Block(V(-X1, floor, Z0), V(X1, ph, Z1)), Block(V(-X1+0.9, floor-1, Z0+0.9), V(X1-0.9, ph+1, Z1-0.9))))
	g.Fill(Subtract(Block(V(-X1, ph-0.8, Z0), V(X1, ph, Z1)), Block(V(-X1+M-0.45, 0, Z0+M-0.45), V(X1-M+0.45, ph+1, Z1-M+0.45))))
	// Inner arcade.
	g.Fill(Subtract(Block(V(-X1+M-0.45, floor, Z0+M-0.45), V(X1-M+0.45, ph, Z1-M+0.45)), Block(V(-X1+M+0.45, 0, Z0+M+0.45), V(X1-M-0.45, ph+1, Z1-M-0.45))))
	g.Paint(White, Subtract(Block(V(-X1-0.2, ph-0.5, Z0), V(X1+0.2, ph, Z1+0.2)), Block(V(-X1+M+0.5, 0, Z0+M+0.5), V(X1-M-0.5, ph+1, Z1-M-0.5))))
	for i := 1; i <= 9; i++ {
		x := -X1 + M*(float64(i)+0.5)
		for _, z := range []float64{Z0 + M, Z1 - M} {
			g.Carve(m.pointed(V(x, floor, z), 1.2, 2.4, 1.6, AxisZ))
		}
	}
	for j := 1; j <= 6; j++ {
		z := Z0 + M*(float64(j)+0.5)
		for _, x := range []float64{-X1 + M, X1 - M} {
			g.Carve(m.pointed(V(x, floor, z), 1.2, 2.4, 1.6, AxisX))
		}
	}
	// Windows in the outer wall.
	for i := 0; i <= 10; i++ {
		x := -X1 + M*(float64(i)+0.5)
		if i < 4 || i > 6 {
			g.Paint(White, m.pointed(V(x, floor+0.8, Z1), 1.0, 1.9, 0.6, AxisZ))
			g.Carve(m.pointed(V(x, floor+1.1, Z1-0.45), 0.65, 1.5, 2, AxisZ))
		}
	}
	for j := 0; j <= 7; j++ {
		z := Z0 + M*(float64(j)+0.5)
		if j == 3 || j == 0 {
			continue
		}
		for _, s := range []float64{-1, 1} {
			g.Paint(White, m.pointed(V(s*X1, floor+0.8, z), 1.0, 1.9, 0.6, AxisX))
			g.Carve(m.pointed(V(s*(X1-0.45), floor+1.1, z), 0.65, 1.5, 2, AxisX))
		}
	}
	// A ring of small domes over the portico bays.
	dome := func(c Vec2, r, drum float64) {
		g.Use(Stone).Fill(m.polygon(c, 8, r-0.15, ph-0.2, ph+drum, math.Pi/8))
		if drum > 0.6 {
			g.Paint(White, m.polygon(c, 8, r, ph+drum-0.35, ph+drum, math.Pi/8))
		}
		g.Use(Slate).FillVolume(m.cap(c.At(ph+drum), r, r-0.15, 0))
		g.Use(Gold).Fill(Cylinder{Center: c, Radius: 0.4, Y0: ph + drum + r - 0.6, Y1: ph + drum + r + 0.5})
	}
	// The mosque-side bays (the last congregation portico) carry taller, larger domes.
	for i := 0; i <= 10; i++ {
		x := -X1 + M*(float64(i)+0.5)
		switch {
		case i == 5:
			dome(P(x, Z0+M/2), 1.85, 1.6)
		case i > 0 && i < 10:
			dome(P(x, Z0+M/2), 1.7, 1.0)
		default:
			dome(P(x, Z0+M/2), 1.6, 0.45)
		}
		if i != 5 {
			dome(P(x, Z1-M/2), 1.6, 0.45)
		}
	}
	for j := 1; j <= 6; j++ {
		z := Z0 + M*(float64(j)+0.5)
		dome(P(-X1+M/2, z), 1.6, 0.45)
		dome(P(X1-M/2, z), 1.6, 0.45)
	}

	// Main gate: a tall portal block with a deep pointed niche and a small dome.
	gc := P(0, Z1-0.6)
	gTop := floor + 10.5
	g.Use(Stone).Fill(BlockAt(gc, 7, 3.6, floor, gTop))
	g.Use(White).Fill(BlockAt(gc, 7.4, 4, gTop-0.6, gTop))
	g.Paint(White, m.pointed(V(0, floor, Z1+1.2), 2.7, 5.2, 0.6, AxisZ))
	g.Carve(m.pointed(V(0, floor, Z1+1.2), 2.2, 4.8, 1.6, AxisZ))
	g.Carve(m.pointed(V(0, floor, Z1-0.6), 1.4, 3.6, 5, AxisZ))
	g.Use(Stone).FillVolume(Cylinder{Center: gc, Radius: 2.4, Y0: gTop, Y1: gTop + 0.8})
	g.Use(Slate).FillVolume(m.cap(gc.At(gTop+0.8), 2.4, 2.3, 0))
	m.finial(g, gc, gTop+3.1, 0.8)
	// Side gates.
	for _, s := range []float64{-1, 1} {
		c := P(s*X1, Z0+M*3.5)
		g.Use(Stone).Fill(BlockAt(c, 2.6, 4.4, floor, ph+1.6))
		g.Use(White).Fill(BlockAt(c, 2.9, 4.7, ph+1.2, ph+1.6))
		g.Carve(m.pointed(V(c.X, floor, c.Y), 1.2, 2.8, 4, AxisX))
		g.Use(Slate).FillVolume(m.cap(c.At(ph+1.6), 1.5, 1.4, 0))
	}

	// Şadırvan: hexagonal basin, six columns, broad lead eaves and a small dome.
	sc := P(0, (Z0+Z1)/2)
	g.Use(White).Fill(Subtract(m.polygon(sc, 6, 2.6, floor, floor+1.1, 0), m.polygon(sc, 6, 1.9, floor+0.6, floor+1.2, 0)))
	g.Paint(Water, m.polygon(sc, 6, 1.9, floor, floor+0.7, 0))
	g.Use(White).Fill(Cylinder{Center: sc, Radius: 0.6, Y0: floor, Y1: floor + 2.4})
	for k := 0; k < 6; k++ {
		at := Polar(sc, 3.2, float64(k)*math.Pi/3+math.Pi/6)
		g.Use(White).Fill(Cylinder{Center: at, Radius: 0.42, Y0: floor, Y1: floor + 4})
	}
	g.Use(Slate).FillVolume(Subtract(Frustum{sc, 4.6, 2.8, floor + 4, floor + 5.2}, Frustum{sc, 3.8, 2.0, floor + 4, floor + 4.8}))
	g.Use(Stone).Fill(Cylinder{Center: sc, Radius: 2.6, Inner: 1.8, Y0: floor + 4.9, Y1: floor + 5.8})
	g.Use(Slate).FillVolume(m.cap(sc.At(floor+5.8), 2.4, 2.2, 0.9))
	m.finial(g, sc, floor+8, 0.7)
}

// minaret is a slender fluted shaft on a square pedestal with corbelled şerefe
// balconies and a pencil-point lead cap.
func (m BlueMosque) minaret(g *Grid, c Vec2, balconies []float64, shaftTop float64) {
	const floor, r = blueMosqueFloor, 1.45
	g.Use(Stone).Fill(BlockAt(c, 2*r+1.9, 2*r+1.9, floor, floor+7))
	g.Use(White).Fill(BlockAt(c, 2*r+2.2, 2*r+2.2, floor+6.6, floor+7.1))
	g.Carve(m.pointed(V(c.X, floor, c.Y), 0.6, 1.4, 6, AxisX))
	g.Use(White).FillVolume(m.polygon(c, 16, r+0.6, floor+7, floor+8.2, 0), m.polygon(c, 12, r, floor+8, shaftTop, 0))
	for i, b := range balconies {
		g.Use(Stone).FillVolume(Frustum{c, r, r + 1.1, b - 1.4, b})
		g.FillVolume(Cylinder{Center: c, Radius: r + 1.1, Inner: r + 0.3, Y0: b, Y1: b + 0.9})
		if i == len(balconies)-1 {
			g.Carve(Cylinder{Center: c, Radius: r + 0.1, Inner: r - 0.3, Y0: b + 0.9, Y1: shaftTop})
		}
	}
	g.Use(Slate).FillVolume(Frustum{c, r + 0.3, 0.15, shaftTop, shaftTop + 8.5})
	g.Use(Gold).Fill(Cylinder{Center: c, Radius: 0.4, Y0: shaftTop + 7.5, Y1: shaftTop + 10.2})
	g.FillVolume(Sphere(c.At(shaftTop+8.7), 0.6))
}

// minarets places the six minarets: four with three balconies at the hall corners and
// two with two balconies at the far corners of the forecourt.
func (m BlueMosque) minarets(g *Grid) {
	const d = 22.3
	for _, s := range []float64{-1, 1} {
		m.minaret(g, P(s*d, -d), []float64{27, 34, 40.5}, 44.5)
		m.minaret(g, P(s*d, d), []float64{27, 34, 40.5}, 44.5)
		m.minaret(g, P(s*d, blueMosqueCZ1+2.2), []float64{28, 35.5}, 41)
	}
}

// planeTree is a broad oriental plane: a trunk and a hollow rounded canopy.
func (m BlueMosque) planeTree(g *Grid, c Vec2, seed int) {
	base := m.height(c.X, c.Y) - 0.2
	h := 4.5 + 1.5*Hash01(seed, 1)
	r := 2.6 + 0.9*Hash01(seed, 2)
	g.With(Wood, func() { g.Fill(Cylinder{Center: c, Radius: 0.5, Y0: base, Y1: base + h}) })
	color := Foliage
	if Hash01(seed, 3) < 0.35 {
		color = Grass
	}
	g.With(color, func() {
		ctr := c.At(base + h + 0.6)
		g.FillVolume(Subtract(Ellipsoid{ctr, V(r, r*0.8, r)}, Ellipsoid{ctr, V(r-1.1, r*0.8-1.1, r-1.1)}))
	})
}

// garden is the corner of Sultanahmet park in front of the gate: the round fountain
// pool with its jets, plane trees round the park and along the precinct.
func (m BlueMosque) garden(g *Grid) {
	const floor = blueMosqueFloor
	fc := P(0, 59.5)
	g.Use(White).FillVolume(Cylinder{Center: fc, Radius: 5.4, Inner: 4.6, Y0: floor - 0.2, Y1: floor + 0.9})
	g.Paint(Water, Cylinder{Center: fc, Radius: 4.6, Y0: 0, Y1: floor})
	g.Use(Water).Fill(Cylinder{Center: fc, Radius: 4.6, Y0: floor - 0.2, Y1: floor + 0.5})
	g.Use(White).Fill(Cylinder{Center: fc, Radius: 1.2, Y0: floor, Y1: floor + 1.4}, Cylinder{Center: fc, Radius: 0.5, Y0: floor + 1.4, Y1: floor + 3.6})
	for k := 0; k < 8; k++ {
		a := float64(k) * math.Pi / 4
		g.Use(White).Fill(Capsule{A: Polar(fc, 3.4, a).At(floor + 0.6), B: Polar(fc, 2.4, a).At(floor + 2.2), Radius: 0.4})
	}
	seed := 300
	for _, p := range []Vec2{P(-12, 55), P(-17, 61.5), P(-9.5, 65.5), P(8.5, 66), P(26, 56), P(25.5, 64)} {
		seed++
		m.planeTree(g, p, seed)
	}
	// Along the flanks of the mosque and courtyard, and behind the qibla wall.
	// The hall's flanks stay open so the galleries show.
	for _, z := range []float64{-17, -9.5, 25, 32.5, 40, 58} {
		for _, s := range []float64{-1, 1} {
			seed++
			if s > 0 && z > 50 {
				continue
			}
			m.planeTree(g, P(s*27.2, z+Hash01(seed, 5)-0.5), seed)
		}
	}
	for x := -15.0; x <= 15; x += 7.5 {
		seed++
		m.planeTree(g, P(x, -26.8), seed)
	}
}

// precinct is the low outer wall of the mosque precinct with grilled windows and gates
// on the axes.
func (m BlueMosque) precinct(g *Grid) {
	const (
		floor         = blueMosqueFloor
		x1, z0, z1, h = blueMosqueB + 5, -blueMosqueB - 4.5, blueMosqueCZ1 + 3.5, 2.8
	)
	g.Use(Stone).Fill(Subtract(Block(V(-x1, floor, z0), V(x1, floor+h, z1)), Block(V(-x1+0.8, 0, z0+0.8), V(x1-0.8, 9, z1-0.8))))
	g.Use(White).Fill(Subtract(Block(V(-x1-0.2, floor+h-0.4, z0-0.2), V(x1+0.2, floor+h, z1+0.2)), Block(V(-x1+1, 0, z0+1), V(x1-1, 9, z1-1))))
	// Grilles: rows of small openings.
	g.Carve(Func{Box{V(-x1-1, floor+0.9, z0-1), V(x1+1, floor+h-0.8, z1+1)}, func(p Vec3) bool {
		u := p.X
		if math.Abs(math.Abs(p.X)-x1) < 1 {
			u = p.Z
		}
		return math.Mod(math.Abs(u)+0.4, 2.2) < 1.1
	}})
	// Gates on the axes and opposite the courtyard side gates.
	gate := func(c Vec2, through Axis) {
		w, d := 5.0, 1.6
		if through == AxisX {
			w, d = d, w
		}
		g.Use(Stone).Fill(BlockAt(c, w, d, floor, floor+5.4))
		g.Use(White).Fill(BlockAt(c, w+0.4, d+0.4, floor+5, floor+5.4))
		g.Carve(m.pointed(c.At(floor), 1.4, 2.6, 4, through))
	}
	gate(P(0, z1), AxisZ)
	gate(P(0, z0), AxisZ)
	for _, s := range []float64{-1, 1} {
		gate(P(s*x1, 0), AxisX)
		gate(P(s*x1, blueMosqueCZ0+blueMosqueM*3.5), AxisX)
	}
}

// blueMosqueTurbeC is the centre of Sultan Ahmed's tomb in the park.
var blueMosqueTurbeC = P(19.5, 59)

// turbeInterior is the hollow of the tomb.
func (BlueMosque) turbeInterior(p Vec3) bool {
	d := p.Plan().Sub(blueMosqueTurbeC)
	if p.Y <= blueMosqueFloor || math.Abs(d.X) > 3.6 || math.Abs(d.Y) > 3.6 {
		return false
	}
	return p.Y < blueMosqueFloor+5.4 || p.Sub(blueMosqueTurbeC.At(blueMosqueFloor+6.2)).Length() < 2.6
}

// turbe is Sultan Ahmed's square domed tomb with a portico of three small domes.
func (m BlueMosque) turbe(g *Grid) {
	const floor = blueMosqueFloor
	c := blueMosqueTurbeC
	g.Use(Stone).Fill(BlockAt(c, 9.6, 9.6, floor-0.2, floor+0.5))
	g.Fill(m.shellBox(c.X-4.4, c.Y-4.4, c.X+4.4, c.Y+4.4, floor, floor+6.2, 1.0))
	g.Use(White).Fill(Subtract(BlockAt(c, 9.2, 9.2, floor+5.7, floor+6.2), BlockAt(c, 7, 7, floor, floor+7)))
	g.Fill(Subtract(BlockAt(c, 9.0, 9.0, floor+2.8, floor+3.1), BlockAt(c, 7, 7, floor, floor+7)))
	for _, a := range []float64{-1.9, 1.9} {
		for _, f := range []struct {
			x, z float64
			ax   Axis
		}{{c.X + a, c.Y - 4.0, AxisZ}, {c.X + a, c.Y + 4.0, AxisZ}, {c.X + 4.0, c.Y + a, AxisX}} {
			g.Carve(m.pointed(V(f.x, floor+0.9, f.z), 0.6, 1.2, 2, f.ax), m.pointed(V(f.x, floor+3.7, f.z), 0.5, 0.8, 2, f.ax))
		}
	}
	g.Carve(m.pointed(V(c.X-4.0, floor+0.5, c.Y), 0.9, 2.2, 2, AxisX))
	// Octagonal drum and dome.
	g.Use(Stone).FillVolume(Subtract(m.polygon(c, 8, 3.7, floor+5.6, floor+7.0, math.Pi/8), m.polygon(c, 8, 2.7, floor+5, floor+7.2, math.Pi/8)))
	g.Paint(White, m.polygon(c, 8, 3.9, floor+6.7, floor+7.0, math.Pi/8))
	g.Use(Slate).FillVolume(m.cap(c.At(floor+7.0), 3.6, 3.4, 1.0))
	m.finial(g, c, floor+10.4, 0.8)
	// Portico on the west front: four columns, a lead roof and three small domes.
	for _, dz := range []float64{-3.9, -1.3, 1.3, 3.9} {
		g.Use(White).Fill(Cylinder{Center: P(c.X-6.6, c.Y+dz), Radius: 0.42, Y0: floor, Y1: floor + 4})
	}
	g.Use(Stone).Fill(Block(V(c.X-7.3, floor+4, c.Y-4.4), V(c.X-4.4, floor+4.8, c.Y+4.4)))
	g.Use(White).Fill(Block(V(c.X-7.5, floor+4.4, c.Y-4.6), V(c.X-4.4, floor+4.8, c.Y+4.6)))
	for _, dz := range []float64{-2.6, 0, 2.6} {
		at := P(c.X-5.85, c.Y+dz)
		g.Use(Slate).FillVolume(m.cap(at.At(floor+4.8), 1.3, 1.2, 0))
	}
	// Cypresses round the tomb garden.
	g.With(Foliage, func() {
		for i, q := range []Vec2{P(c.X+6.4, c.Y-4), P(c.X+6.4, c.Y), P(c.X+6.4, c.Y+4), P(c.X-2, c.Y+6.4), P(c.X+2, c.Y+6.4), P(c.X-2, c.Y-6.4), P(c.X+2, c.Y-6.4)} {
			h := 3 + 0.7*Hash01(i, 77)
			g.FillVolume(Ellipsoid{q.At(m.height(q.X, q.Y) + h), V(0.95, h, 0.95)})
		}
	})
}
