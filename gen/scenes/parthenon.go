package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// Parthenon is the Acropolis of Athens crowned by the ruined Parthenon. A hollow
// limestone rock with sheer cliffs and scrubby talus slopes carries the fortification
// walls. On top stand the Parthenon (three-step krepis, a peristyle of 8 × 17 Doric
// columns with entasis and capitals, triglyph frieze, pediment fragments, a roof that
// survives only at the west end, ruined cella walls, fallen drums), the Erechtheion with
// its Caryatid porch, the Propylaea gateway above the monumental west stair, the Temple
// of Athena Nike on its bastion and the round Temple of Rome and Augustus. The arched
// Odeon of Herodes Atticus is set into the south-west slope; olive trees and paths
// cover the slopes. East is +X, south (the front) is +Z.
type Parthenon struct{}

func init() { Register(Parthenon{}) }

func (Parthenon) Name() string  { return "parthenon" }
func (Parthenon) Title() string { return "Parthenon" }

const (
	parthenonTop = 16.0 // Plateau level.
	parthenonSX  = 51.7 / 16
	parthenonSZ  = 21.7 / 7
	parthenonK   = 1.12 // Parthenon scale over the base proportions.
)

// parthenonC is the Parthenon's plan center; columns lines are ±25.85 × ±10.85 around it.
var parthenonC = P(12, 9)

// parthenonOdeon is the Odeon's orchestra center (the cavea opens to the south).
var parthenonOdeon = P(-26, 42)

// parthenonPlateau outlines the summit plateau.
var parthenonPlateau = []Vec2{
	P(-54, -6), P(-48, -18), P(-34, -27), P(-10, -30), P(20, -30), P(42, -27), P(53, -16),
	P(58, -2), P(54, 12), P(44, 23), P(20, 30.5), P(-8, 30.5), P(-30, 27.5), P(-46, 17), P(-54, 8),
}

func parthenonNoise2(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

func parthenonNoise3(x, y, z float64, seed int) float64 {
	ix, iy, iz := math.Floor(x), math.Floor(y), math.Floor(z)
	fx, fy, fz := x-ix, y-iy, z-iz
	sx, sy, sz := fx*fx*(3-2*fx), fy*fy*(3-2*fy), fz*fz*(3-2*fz)
	i, j, k := int(ix), int(iy), int(iz)
	lerp := func(a, b, t float64) float64 { return a + (b-a)*t }
	c := func(di, dj, dk int) float64 { return Hash01(i+di, j+dj, k+dk, seed) }
	x00 := lerp(c(0, 0, 0), c(1, 0, 0), sx)
	x10 := lerp(c(0, 1, 0), c(1, 1, 0), sx)
	x01 := lerp(c(0, 0, 1), c(1, 0, 1), sx)
	x11 := lerp(c(0, 1, 1), c(1, 1, 1), sx)
	return lerp(lerp(x00, x10, sy), lerp(x01, x11, sy), sz)
}

// edge is the signed plan distance outside the (noisy) plateau outline.
func (Parthenon) edge(x, z float64) float64 {
	q := P(x, z)
	best := math.Inf(1)
	n := len(parthenonPlateau)
	for i := 0; i < n; i++ {
		a, b := parthenonPlateau[i], parthenonPlateau[(i+1)%n]
		ab := b.Sub(a)
		t := math.Max(0, math.Min(1, q.Sub(a).Dot(ab)/ab.Dot(ab)))
		best = math.Min(best, q.Sub(a.Add(ab.Scale(t))).Length())
	}
	if InsidePolygon(parthenonPlateau, q) {
		best = -best
	}
	return best + 1.8*(parthenonNoise2(x/11, z/11, 3)-0.5)
}

// The west stair climbs from the foot of the hill to the Propylaea.
func (Parthenon) inStair(x, z float64) bool {
	return x >= parthenonStair-0.5 && x <= -52 && z >= -6.5 && z <= 4
}

// parthenonStair is where the west stair starts: 0.75 risers on 1.0 treads up to -52.
const parthenonStair = -52 - (parthenonTop-0.5)/0.75

func (Parthenon) stair(x float64) float64 {
	return math.Max(0.5, math.Min(parthenonTop, 0.5+math.Ceil(x-parthenonStair)*0.75))
}

// parthenonDionysus is the orchestra center of the ruined Theatre of Dionysus.
var parthenonDionysus = P(27, 41.5)

func (Parthenon) dseat(r float64) float64 {
	if r < 4 {
		return 0.8
	}
	return 1.3 + 0.6*math.Floor((r-4)/0.85)
}

// seat is the Odeon cavea surface at radius r: orchestra, then stepped rows.
func (Parthenon) seat(r float64) float64 {
	if r < 5 {
		return 0.6
	}
	row := math.Floor((r - 5) / 0.9)
	return 1.0 + 0.7*row
}

// height is the terrain surface; negative means no ground.
func (a Parthenon) height(x, z float64) float64 {
	d := a.edge(x, z)
	cliff := 3.0 + 1.6*parthenonNoise2(x/13, z/13, 5)
	foot := 7.0 + 3*parthenonNoise2(x/15, z/15, 7)
	run := 9.5 + 4*parthenonNoise2(x/17, z/17, 9)
	var h float64
	switch {
	case d <= 1.0:
		h = parthenonTop
	case d < 1.0+cliff:
		t := (d - 1.0) / cliff
		h = parthenonTop - (parthenonTop-foot)*math.Sqrt(t)
	case d < 1.0+cliff+run:
		t := (d - 1.0 - cliff) / run
		h = foot*math.Pow(1-t, 1.5) + 1.2*(parthenonNoise2(x/4, z/4, 11)-0.4)*(1-t)
		h = math.Max(0.45, h)
	default:
		h = -1
	}
	if a.inStair(x, z) {
		h = a.stair(x)
	}
	// The Nike bastion stands on its own foundations.
	if x >= -64.5 && x <= -51.5 && z >= 5 && z <= 17 && h >= 0 {
		h = math.Min(h, 2)
	}
	// The Odeon is cut into the slope.
	r := P(x, z).Sub(parthenonOdeon).Length()
	if z <= parthenonOdeon.Y {
		if r <= 13.5 {
			h = a.seat(r)
		} else if r <= 14.8 {
			h = math.Max(h, 8.6)
		}
	} else if x >= -45 && x <= -7 && z <= 48 && h > 0.5 {
		h = 0.5
	}
	// The Stoa of Eumenes terrace is cut into the slope between the theatres.
	if x >= -7.5 && x <= 13 && z >= 36.4 && z <= 43 && h > 1.2 {
		h = 1.2
	}
	// The Theatre of Dionysus is a ruined cavea worn into the south-east slope.
	if r := P(x, z).Sub(parthenonDionysus).Length(); h > 0 && r <= 12.5 && z <= parthenonDionysus.Y+0.5 {
		h = math.Min(h, a.dseat(r))
	}
	return h
}

// parthenonMap samples the terrain on a regular plan grid.
type parthenonMap struct {
	x0, z0, step float64
	nx, nz       int
	h            []float64
}

func (m *parthenonMap) at(x, z float64) float64 {
	fx, fz := (x-m.x0)/m.step, (z-m.z0)/m.step
	if fx < 0 || fz < 0 || fx >= float64(m.nx-1) || fz >= float64(m.nz-1) {
		return -1
	}
	i, j := int(fx), int(fz)
	tx, tz := fx-float64(i), fz-float64(j)
	k := j*m.nx + i
	h00, h10, h01, h11 := m.h[k], m.h[k+1], m.h[k+m.nx], m.h[k+m.nx+1]
	if h00 < 0 || h10 < 0 || h01 < 0 || h11 < 0 {
		// Nearest sample at the rim keeps the edge crisp.
		if tx >= 0.5 {
			i++
		}
		if tz >= 0.5 {
			j++
		}
		return m.h[j*m.nx+i]
	}
	a := h00 + (h10-h00)*tx
	b := h01 + (h11-h01)*tx
	return a + (b-a)*tz
}

// low is the lowest neighboring surface, so steep faces are closed.
func (m *parthenonMap) low(x, z float64) float64 {
	lo := math.Max(0, m.at(x, z))
	for k := 0; k < 8; k++ {
		a := float64(k) * math.Pi / 4
		lo = math.Min(lo, math.Max(0, m.at(x+0.9*math.Cos(a), z+0.9*math.Sin(a))))
	}
	return lo
}

// where selects occupied cells of a given color for which f holds. Only exact cell
// centers pass, so the grid's half-unit snapping never selects a neighboring cell.
func (Parthenon) where(g *Grid, box Box, from Color, f func(Vec3) bool) Func {
	return Func{box, func(p Vec3) bool {
		c := Nearest(p.Scale(1 / g.Span()))
		if g.World(c).Sub(p).Length() > 1e-6 {
			return false
		}
		col, ok := g.ColorOf(c)
		return ok && col == from && f(p)
	}}
}

func (Parthenon) disk(c Vec2, p Vec3) float64 { return p.Plan().Sub(c).Length() }

// doric places a Doric column: tapering shaft with entasis, echinus and square abacus.
func (a Parthenon) doric(g *Grid, c Vec2, y0, height, r float64) {
	top := y0 + height*0.885
	g.FillVolume(Func{Box{V(c.X-r-0.2, y0, c.Y-r-0.2), V(c.X+r+0.2, top, c.Y+r+0.2)}, func(p Vec3) bool {
		t := (p.Y - y0) / (top - y0)
		return a.disk(c, p) <= r*(1-0.18*t+0.06*math.Sin(math.Pi*t))
	}})
	g.Fill(Frustum{c, r * 0.85, r * 1.35, top, y0 + height*0.95})
	g.Fill(BlockAt(c, r*2.8, r*2.8, y0+height*0.95, y0+height))
}

// ionic places a slender Ionic column with a molded base and a volute capital whose
// long side runs along the colonnade axis.
func (a Parthenon) ionic(g *Grid, c Vec2, y0, height, r float64, along Axis) {
	g.Fill(Cylinder{Center: c, Radius: r * 1.4, Y0: y0, Y1: y0 + 0.4})
	g.FillVolume(Func{Box{V(c.X-r-0.2, y0, c.Y-r-0.2), V(c.X+r+0.2, y0+height-0.5, c.Y+r+0.2)}, func(p Vec3) bool {
		return a.disk(c, p) <= r*(1-0.1*(p.Y-y0)/height)
	}})
	w, d := r*3.2, r*2
	if along == AxisZ {
		w, d = d, w
	}
	g.Fill(BlockAt(c, w, d, y0+height-0.6, y0+height))
}

// gable is a low-pitched roof or tympanum over a rectangle, ridge along the axis.
// Thick > 0 keeps only a shell of that thickness under the slopes.
func (Parthenon) gable(min, max Vec2, base, pitch, thick float64, ridge Axis) Func {
	half := (max.Y - min.Y) / 2
	mid := (max.Y + min.Y) / 2
	if ridge == AxisZ {
		half, mid = (max.X-min.X)/2, (max.X+min.X)/2
	}
	return Func{Box{min.At(base), max.At(base + pitch*half)}, func(p Vec3) bool {
		across := math.Abs(p.Z - mid)
		if ridge == AxisZ {
			across = math.Abs(p.X - mid)
		}
		top := base + pitch*(half-across)
		return p.Y <= top && (thick <= 0 || p.Y >= top-thick)
	}}
}

func (a Parthenon) Build(g *Grid) Order {
	a.terrain(g)
	a.walls(g)
	a.parthenon(g)
	a.erechtheion(g)
	a.propylaea(g)
	a.nike(g)
	a.roma(g)
	a.odeon(g)
	a.stoa(g)
	a.flag(g)
	a.olives(g)
	return Sweep(parthenonC, math.Pi)
}

func (a Parthenon) terrain(g *Grid) {
	m := &parthenonMap{x0: -84, z0: -54, step: 0.25}
	m.nx, m.nz = int(160/m.step)+1, int(106/m.step)+1
	m.h = make([]float64, m.nx*m.nz)
	for j := 0; j < m.nz; j++ {
		for i := 0; i < m.nx; i++ {
			m.h[j*m.nx+i] = a.height(m.x0+float64(i)*m.step, m.z0+float64(j)*m.step)
		}
	}
	box := Box{V(-84, 0, -54), V(76, parthenonTop, 52)}
	g.Use(Stone).Fill(Func{box, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		if h < 0 || p.Y > h {
			return false
		}
		return p.Y >= math.Min(h-0.8, m.low(p.X, p.Z)-0.35)
	}})
	surface := func(p Vec3) bool { return p.Y > m.at(p.X, p.Z)-0.9 }
	slope := func(p Vec3) bool { return p.Y < parthenonTop-2.5 && surface(p) }
	// Slopes: pale rocky ground, ochre earth, dark maquis scrub and the Peripatos path.
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		return slope(p) && parthenonNoise2(p.X/7, p.Z/7, 24)+0.25*parthenonNoise2(p.X/2, p.Z/2, 26) > 0.68
	}})
	scrub := func(p Vec3) float64 {
		return parthenonNoise2(p.X/5, p.Z/5, 21) + 0.3*parthenonNoise2(p.X/1.6, p.Z/1.6, 22)
	}
	g.Paint(Foliage, Func{box, func(p Vec3) bool { return slope(p) && scrub(p) > 0.66 }})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		return slope(p) && parthenonNoise2(p.X/1.5, p.Z/1.5, 23) > 0.82
	}})
	// Theatre of Dionysus: marble front rows, overgrown upper rows, a paved orchestra.
	dio := Func{Box{V(14, 0, 28), V(40, 9, 42.5)}, func(p Vec3) bool {
		r := a.disk(parthenonDionysus, p)
		return r <= 12.5 && p.Z <= parthenonDionysus.Y+0.5 && p.Y > a.dseat(r)-1.0 && p.Y <= a.dseat(r)+0.1
	}}
	g.Paint(White, Intersect(dio, Func{dio.Box, func(p Vec3) bool {
		r := a.disk(parthenonDionysus, p)
		return r < 8+3*parthenonNoise2(p.X/2, p.Z/2, 28) || Hash01(int(p.X), int(p.Z), 29) > 0.6
	}}))
	g.Paint(Stone, Intersect(dio, Func{dio.Box, func(p Vec3) bool {
		r := a.disk(parthenonDionysus, p)
		return r < 4 || p.Y < a.dseat(r)-0.35
	}}))
	// The cliffs: banded gray limestone with dark run-off streaks and ochre strata.
	cliff := func(p Vec3) bool {
		d := a.edge(p.X, p.Z)
		return d > 0.3 && p.Y > 4 && p.Y <= parthenonTop && m.at(p.X, p.Z)-m.low(p.X, p.Z) > 1.2
	}
	g.Paint(Stone, Func{box, cliff})
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		return cliff(p) && math.Sin(p.Y*1.1+3*parthenonNoise2(p.X/6, p.Z/6, 25)) > 0.55
	}})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		return cliff(p) && parthenonNoise3(p.X/1.8, p.Y/7, p.Z/1.8, 27) > 0.66
	}})
	// Plateau: worn limestone with patches.
	top := Func{Box{V(-60, parthenonTop-1.5, -34), V(64, parthenonTop+0.5, 33)}, func(p Vec3) bool {
		return a.edge(p.X, p.Z) < 1.0
	}}
	g.Paint(Stone, top)
	g.Paint(Sand, Intersect(top, Func{top.Box, func(p Vec3) bool { return parthenonNoise2(p.X/7, p.Z/7, 31) > 0.62 }}))
	g.Paint(DarkStone, Intersect(top, Func{top.Box, func(p Vec3) bool { return parthenonNoise2(p.X/2.5, p.Z/2.5, 33) > 0.82 }}))
	// The Sacred Way from the Propylaea east between the temples.
	way := Func{top.Box, func(p Vec3) bool {
		x, z := p.X, p.Z
		center := -1.0
		if x > -36 {
			center = -1 - (x+36)*0.12
		}
		if x > -14 {
			center = -3.6 - (x+14)*0.0
		}
		return x > -50 && x < 54 && math.Abs(z-center) < 1.3+0.4*parthenonNoise2(x/3, 0, 35)
	}}
	g.Paint(Sand, Intersect(top, way))
	// Peripatos path around the slopes.
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		if !surface(p) || p.Y > parthenonTop-4 {
			return false
		}
		d := a.edge(p.X, p.Z)
		w := 7.5 + 2*parthenonNoise2(p.X/20, p.Z/20, 37)
		return math.Abs(d-w) < 0.75
	}})
	// West stair: marble treads between ashlar cheeks.
	stairs := Func{Box{V(-76, 0, -7), V(-51, parthenonTop, 5)}, func(p Vec3) bool {
		return a.inStair(p.X, p.Z)
	}}
	g.Paint(White, stairs)
	g.Paint(Stone, Intersect(stairs, Func{stairs.Box, func(p Vec3) bool {
		return p.Z < -5.4 || p.Z > 2.9 || p.Y < a.stair(p.X)-0.5
	}}))
	g.Paint(Sand, Intersect(stairs, Func{stairs.Box, func(p Vec3) bool {
		return math.Abs(p.Z+1.25) < 1.1 && p.Y > a.stair(p.X)-0.5
	}}))
	// Odeon: marble seats, stone stairways, a checkered orchestra.
	cavea := Func{Box{V(-41, 0, 26), V(-11, 9, parthenonOdeon.Y)}, func(p Vec3) bool {
		return a.disk(parthenonOdeon, p) <= 13.5 && p.Z <= parthenonOdeon.Y
	}}
	g.Paint(White, cavea)
	g.Paint(Stone, Intersect(cavea, Func{cavea.Box, func(p Vec3) bool {
		v := p.Plan().Sub(parthenonOdeon)
		ang := math.Atan2(-v.Y, v.X) / math.Pi * 6
		return math.Abs(ang-math.Round(ang)) < 0.06*14/math.Max(v.Length(), 5) || p.Y < a.seat(v.Length())-0.4
	}}))
	g.Paint(DarkStone, Intersect(cavea, Func{cavea.Box, func(p Vec3) bool {
		r := a.disk(parthenonOdeon, p)
		return r < 5 && (int(math.Floor(p.X/1.2))+int(math.Floor(p.Z/1.2)))%2 == 0
	}}))
	g.Paint(Sand, Func{Box{V(-42, 0, 25), V(-10, 9, parthenonOdeon.Y)}, func(p Vec3) bool {
		r := a.disk(parthenonOdeon, p)
		return r > 13.5 && r <= 14.8 && p.Z <= parthenonOdeon.Y && p.Y > 0.5
	}})
}

// walls are the fortification walls on the cliff edge, rising above the plateau.
func (a Parthenon) walls(g *Grid) {
	box := Box{V(-60, parthenonTop-5, -34), V(64, parthenonTop+1.6, 33)}
	open := func(p Vec3) bool { return p.X < -46 && p.Z > -11 && p.Z < 18 }
	band := func(p Vec3) bool {
		d := a.edge(p.X, p.Z)
		return d >= -1.2 && d <= 1.0 && !open(p)
	}
	g.Use(Stone).Fill(Func{box, func(p Vec3) bool { return p.Y > parthenonTop && band(p) }})
	masonry := Func{box, func(p Vec3) bool {
		d := a.edge(p.X, p.Z)
		return d >= -1.2 && d <= 2.2 && p.Y > parthenonTop-4.5 && !open(p) && (p.Y > parthenonTop-1.2 || d > 0.6)
	}}
	g.Paint(Sand, masonry)
	g.Paint(Stone, Intersect(masonry, Func{box, func(p Vec3) bool {
		return math.Mod(p.Y+20, 1.5) < 0.45 || parthenonNoise2(p.X/2, p.Z/2+p.Y/3, 41) > 0.68
	}}))
	g.Paint(DarkStone, Intersect(masonry, Func{box, func(p Vec3) bool { return parthenonNoise2(p.X/1.5, p.Z/1.5+p.Y, 43) > 0.8 }}))
}

func (a Parthenon) parthenon(g *Grid) {
	const k = parthenonK
	c := parthenonC
	top := parthenonTop
	s := top + 2.25*k // Stylobate.
	e0 := s + 8.4*k   // Entablature base.
	lx, lz := 25.85*k, 10.85*k
	sx, sz := parthenonSX*k, parthenonSZ*k
	colX := func(i int) float64 { return c.X - lx + float64(i)*sx }
	colZ := func(j int) float64 { return c.Y - lz + float64(j)*sz }
	// Krepis: three steps, hollow below the stylobate floor.
	g.Use(White)
	for n := 0; n < 3; n++ {
		grow := 0.9 * k * float64(2-n)
		g.Fill(BlockAt(c, 54*k+2*grow, 24*k+2*grow, top, top+0.75*k*float64(n+1)))
	}
	g.Carve(BlockAt(c, 54*k-3, 24*k-3, top-1.5, s-1.0))
	g.Paint(Stone, Func{BlockAt(c, 60*k, 30*k, top, s).Bounds(), func(p Vec3) bool {
		return math.Mod(p.Y-top, 0.75*k) < 0.3
	}})
	// Peristyle: 8 × 17 Doric columns; the explosion of 1687 left gaps on both flanks.
	stump := func(i, j int) float64 {
		if j == 0 && i >= 5 && i <= 7 || j == 7 && (i == 8 || i == 9) {
			return 0.9 + 2.8*Hash01(i, j, 51)
		}
		return 0
	}
	for i := 0; i <= 16; i++ {
		for j := 0; j <= 7; j++ {
			if i != 0 && i != 16 && j != 0 && j != 7 {
				continue
			}
			if h := stump(i, j); h > 0 {
				g.FillVolume(Cylinder{Center: P(colX(i), colZ(j)), Radius: 0.78 * k, Y0: s, Y1: s + h})
				continue
			}
			a.doric(g, P(colX(i), colZ(j)), s, 8.4*k, 0.8*k)
		}
	}
	// Entablature ring: architrave, triglyph frieze, cornice.
	ea, ef, ec := 1.1*k, 2.2*k, 2.8*k // Tops of architrave, frieze and cornice above e0.
	off := func(p Vec3) (float64, bool) {
		dx, dz := math.Abs(p.X-c.X)-lx, math.Abs(p.Z-c.Y)-lz
		return math.Max(dx, dz), dz >= dx
	}
	ring := Box{V(c.X-lx-2.5, e0, c.Y-lz-2.5), V(c.X+lx+2.5, e0+ec, c.Y+lz+2.5)}
	triglyph := func(p Vec3, long bool) bool {
		u, step := p.X-colX(0), sx/2
		if !long {
			u, step = p.Z-colZ(0), sz/2
		}
		t := math.Mod(u+100*step, step)
		return math.Min(t, step-t) < 0.55
	}
	g.Fill(Func{ring, func(p Vec3) bool {
		o, long := off(p)
		switch {
		case p.Y <= e0+ea:
			return o >= -1.1 && o <= 1.1
		case p.Y <= e0+ef:
			if o < -1.1 || o > 1.1 {
				return false
			}
			return o <= 0.65 || triglyph(p, long)
		default:
			return o >= -1.1 && o <= 1.8
		}
	}})
	g.Paint(Stone, Func{ring, func(p Vec3) bool {
		o, long := off(p)
		return p.Y > e0+ea && p.Y <= e0+ef && o > 0.25 && triglyph(p, long)
	}})
	// Gaps over the fallen columns, with ragged ends.
	rag := func(y float64, seed int) float64 { return 1.4 * Hash01(int(math.Floor(y*2.5)), seed) }
	g.Carve(Func{Box{V(colX(4), e0-0.5, c.Y-lz-2.5), V(colX(8), e0+ec+0.5, c.Y-lz+1.3)}, func(p Vec3) bool {
		return p.X > colX(4)+0.7+rag(p.Y, 1) && p.X < colX(8)-0.7-rag(p.Y, 2)
	}})
	g.Carve(Func{Box{V(colX(7), e0-0.5, c.Y+lz-1.3), V(colX(10), e0+ec+0.5, c.Y+lz+2.5)}, func(p Vec3) bool {
		return p.X > colX(7)+0.5+rag(p.Y, 3) && p.X < colX(10)-0.7-rag(p.Y, 4)
	}})
	// Pediments: low tympana with raking cornices, both broken in places.
	p0 := e0 + ec
	half := lz + 0.7
	pitch := 3.4 * k / half
	for _, west := range []bool{true, false} {
		x0, sign := c.X-lx, -1.0
		if !west {
			x0, sign = c.X+lx, 1.0
		}
		lo, hi := math.Min(x0-sign*1.0, x0-sign*0.1), math.Max(x0-sign*1.0, x0-sign*0.1)
		tymp := a.gable(P(lo, c.Y-half), P(hi, c.Y+half), p0, pitch, 0, AxisX)
		lo, hi = math.Min(x0+sign*1.8, x0-sign*1.5), math.Max(x0+sign*1.8, x0-sign*1.5)
		cornice := a.gable(P(lo, c.Y-half-1.1), P(hi, c.Y+half+1.1), p0, pitch, 0.8, AxisX)
		keep := func(p Vec3) bool {
			if west {
				return !(p.Z > c.Y+3.5 && p.Z < c.Y+7.5 && p.Y > p0+1.6+Hash01(int(p.Z*2), 61))
			}
			return !(math.Abs(p.Z-c.Y+0.8) < 1.8+1.2*Hash01(int(p.Y*2), 62) && p.Y > p0+2.2)
		}
		g.Use(White).Fill(Intersect(tymp, Func{tymp.Box, keep}))
		g.Fill(Intersect(cornice, Func{cornice.Box, keep}))
		// Sculpture fragments standing in the pediment corners.
		g.Use(White)
		for _, f := range []Vec3{V(0, 0.8, -8), V(0, 0.8, -5.5), V(0, 1.5, -3), V(0, 0.8, 9)} {
			if !west && f.Z > 0 {
				continue
			}
			g.Fill(Ellipsoid{V(x0+sign*1.1, p0+f.Y, c.Y+f.Z*k), V(0.6, 0.9, 0.7)})
		}
	}
	// What remains of the marble roof: the west bays, broken off to the east.
	roof := a.gable(P(c.X-lx-1.5, c.Y-half-1.1), P(c.X-lx+8.5, c.Y+half+1.1), p0, pitch, 0.75, AxisX)
	g.Use(Stone).Fill(Intersect(roof, Func{roof.Box, func(p Vec3) bool {
		return p.X < c.X-lx+4.5+4*Hash01(int(math.Floor(p.Z/1.6)), 71)
	}}))
	g.Paint(DarkStone, Intersect(roof, Func{roof.Box, func(p Vec3) bool { return math.Mod(p.Z+100, 1.6) < 0.4 }}))
	// Ceiling beams across the west ambulatory.
	for n := 0; n < 3; n++ {
		x := c.X - lx + 1.6 + float64(n)*2.1
		g.Fill(Block(V(x-0.45, e0+ef-0.7, c.Y-lz), V(x+0.45, e0+ef, c.Y+lz)))
	}
	// Cella: ruined long walls, a high west end with its great door, porches.
	hl, hw := 21.0*k, 7.4*k
	xw, xe, xd := -17*k, 17*k, -7*k // West and east end walls, dividing wall.
	cella := func(p Vec3) (bool, float64) {
		x, z := p.X-c.X, p.Z-c.Y
		long := math.Abs(x) <= hl && math.Abs(z) >= hw-1.2 && math.Abs(z) <= hw
		cross := math.Abs(z) <= hw && (math.Abs(x-xw) <= 0.6 || math.Abs(x-xe) <= 0.6 || math.Abs(x-xd) <= 0.55)
		if !long && !cross {
			return false, 0
		}
		h := 0.8 + 5.5*math.Pow(parthenonNoise2(p.X/4.5, p.Z, 81), 2.2) + 0.8*Hash01(int(p.X*1.5), int(p.Z), 82)
		if x < xw+4 {
			h = 10.5 * k / 1.12 // The opisthodomos stands nearly full height.
		} else if x < xw+8 {
			h = math.Max(h, 10.5-(x-xw-4)*2)
		}
		if cross && x > 0 {
			h = 1.5 + 1.5*Hash01(int(p.Z*2), 83)
		}
		return true, h
	}
	g.Use(Stone).Fill(Func{BlockAt(c, 2*hl, 2*hw, s, e0+1).Bounds(), func(p Vec3) bool {
		ok, h := cella(p)
		return ok && p.Y <= s+h
	}})
	g.Carve(Block(V(c.X+xw-0.7, s, c.Y-2.4), V(c.X+xw+0.7, s+7.8, c.Y+2.4)))
	g.Carve(Block(V(c.X+xd-0.7, s, c.Y-1.8), V(c.X+xd+0.7, s+5.5, c.Y+1.8)))
	g.Paint(Sand, a.where(g, BlockAt(c, 2*hl+1, 2*hw+1, s, e0+1).Bounds(), Stone, func(p Vec3) bool {
		return parthenonNoise3(p.X/3, p.Y/3, p.Z/3, 95) > 0.55
	}))
	// Hexastyle porches: the west one carries its architrave, the east one is reduced.
	px := 20.3 * k
	g.Use(White)
	for n := 0; n < 6; n++ {
		z := c.Y - hw + 1.1 + float64(n)*(2*hw-2.2)/5
		a.doric(g, P(c.X-px, z), s, 8.4*k, 0.68*k)
		if n == 0 || n == 5 {
			a.doric(g, P(c.X+px, z), s, 8.4*k, 0.68*k)
		} else {
			g.FillVolume(Cylinder{Center: P(c.X+px, z), Radius: 0.66 * k, Y0: s, Y1: s + 0.8 + 2.2*Hash01(n, 91)})
		}
	}
	g.Fill(Block(V(c.X-px-1.0, e0, c.Y-hw), V(c.X+xw, e0+ea, c.Y+hw)))
	// The Athena Parthenos base and worn floor.
	g.Use(DarkStone).Fill(BlockAt(P(c.X+6*k, c.Y), 5, 7, s, s+0.7))
	floor := BlockAt(c, 2*lx-1.6, 2*lz-1.6, s-0.5, s-0.01).Bounds()
	g.Paint(Stone, a.where(g, floor, White, func(p Vec3) bool { return true }))
	g.Paint(DarkStone, a.where(g, floor, Stone, func(p Vec3) bool {
		return parthenonNoise2(p.X/3, p.Z/3, 93) > 0.35
	}))
	// Fallen drums and blocks around the flanks.
	g.Use(White)
	for n := 0; n < 16; n++ {
		x := c.X - 22 + 44*Hash01(n, 101)
		side := 1.0
		if n%2 == 0 {
			side = -1
		}
		z := c.Y + side*(16.6+1.0*Hash01(n, 102))
		if Hash01(n, 103) < 0.6 {
			g.FillVolume(Capsule{V(x-0.5, top+0.85, z), V(x+0.5, top+0.85, z), 0.85})
		} else {
			g.Fill(BlockAt(P(x, z), 1.6, 1.2, top, top+1.0))
		}
	}
	// Warm patina on the marble, strongest low down and on the south side.
	box := BlockAt(c, 64*k, 34*k, top, p0+5).Bounds()
	g.Paint(Sand, a.where(g, box, White, func(p Vec3) bool {
		return parthenonNoise3(p.X/5, p.Y/4, p.Z/5, 111)+0.1*(p.Z-c.Y)/lz-0.2*(p.Y-s)/10 > 0.6
	}))
}

func (a Parthenon) erechtheion(g *Grid) {
	top := parthenonTop
	f := top + 0.6
	// Krepis and hollow main block.
	g.Use(White).Fill(Block(V(-24.6, top, -23.6), V(-3.4, f, -12.4)))
	g.Fill(Subtract(Block(V(-24, f, -23), V(-4, top+7.2, -13)), Block(V(-23, f, -22), V(-5, top+8, -14))))
	// West façade windows, east and north doors.
	for _, z := range []float64{-20.5, -18, -15.5} {
		g.Carve(Block(V(-24.5, top+4.2, z-0.6), V(-22.5, top+6, z+0.6)))
	}
	g.Carve(Block(V(-5, f, -19.3), V(-3.5, top+5, -16.7)))
	g.Carve(Block(V(-19, f, -23.5), V(-16.5, top+5.2, -22.5)))
	// Dark Eleusinian frieze band and the roof.
	g.Use(DarkStone).Fill(Block(V(-24.3, top+7.2, -23.3), V(-1.4, top+8.1, -12.7)))
	g.Use(Stone).Fill(a.gable(P(-24.6, -23.6), P(-1.0, -12.4), top+8.1, 0.32, 0.8, AxisX))
	g.Use(White).Fill(a.gable(P(-24.2, -23.2), P(-23.4, -12.8), top+8.1, 0.32, 0, AxisX))
	g.Fill(a.gable(P(-1.8, -23.2), P(-1.0, -12.8), top+8.1, 0.32, 0, AxisX))
	// East porch: six Ionic columns.
	for k := 0; k < 6; k++ {
		a.ionic(g, P(-2.2, -22.3+float64(k)*1.72), f, 6.6, 0.48, AxisZ)
	}
	// North porch, taller, with its own pediment.
	g.Fill(Block(V(-21.5, top, -27), V(-14, f, -23)))
	for k := 0; k < 4; k++ {
		x := -20.8 + float64(k)*2.03
		a.ionic(g, P(x, -26.3), f, 7.2, 0.52, AxisX)
		if k == 0 || k == 3 {
			a.ionic(g, P(x, -24.3), f, 7.2, 0.52, AxisX)
		}
	}
	g.Fill(Block(V(-21.4, f+7.2, -27), V(-14.1, f+8.0, -23)))
	g.Use(DarkStone).Fill(Block(V(-21.4, f+8.0, -27), V(-14.1, f+8.6, -23)))
	g.Use(Stone).Fill(a.gable(P(-21.6, -27.2), P(-13.9, -23), f+8.6, 0.32, 0, AxisZ))
	// Caryatid porch on its podium.
	g.Use(White).Fill(Block(V(-23.2, top, -13), V(-15.8, top+1.9, -9.6)))
	py := top + 1.9
	for _, q := range []Vec2{P(-22.4, -10.3), P(-20.6, -10.3), P(-18.4, -10.3), P(-16.6, -10.3), P(-22.4, -12.3), P(-16.6, -12.3)} {
		g.FillVolume(Frustum{q, 0.62, 0.48, py, py + 2.5})
		g.FillVolume(Sphere(q.At(py+2.95), 0.5))
		g.Fill(Cylinder{Center: q, Radius: 0.6, Y0: py + 3.3, Y1: py + 3.8})
	}
	g.Fill(Block(V(-23.4, py+3.8, -13), V(-15.6, py+4.5, -9.4)))
	g.Use(Stone).Fill(Block(V(-23.6, py+4.5, -13), V(-15.4, py+5.0, -9.2)))
	g.Paint(Sand, a.where(g, Box{V(-23.5, py, -13), V(-15.5, py+3.3, -9.5)}, White, func(p Vec3) bool {
		return p.Y > py+2.4 && p.Y < py+3.0 && parthenonNoise3(p.X, p.Y, p.Z, 121) > 0.4
	}))
	// Dark interior seen through the door and windows.
	g.Paint(DarkStone, Block(V(-23.1, f, -22.1), V(-4.9, top+7.1, -13.9)))
	// The sacred olive of the Pandroseion.
	a.olive(g, P(-28.5, -19), top, 0)
}

func (a Parthenon) propylaea(g *Grid) {
	top := parthenonTop
	zc := -1.0
	offs := []float64{-6.6, -4.2, -1.7, 1.7, 4.2, 6.6}
	// Platform and side walls of the central hall.
	g.Use(White).Fill(Block(V(-51, top, zc-8.4), V(-36.5, top+0.7, zc+8.4)))
	f := top + 0.7
	ruin := func(p Vec3) float64 { return 7.5 + 2.2*parthenonNoise2(p.X/3, p.Z, 131) }
	g.Fill(Func{Box{V(-49.5, f, zc-7.9), V(-37.5, top+10, zc+7.9)}, func(p Vec3) bool {
		return math.Abs(math.Abs(p.Z-zc)-7.3) <= 0.55 && p.Y <= f+ruin(p)
	}})
	// Gate wall with five doorways.
	g.Fill(Block(V(-41.6, f, zc-7.3), V(-40.4, f+8.6, zc+7.3)))
	g.Carve(Block(V(-42, f, zc-1.7), V(-40, f+6.2, zc+1.7)))
	for _, s := range []float64{-1, 1} {
		g.Carve(Block(V(-42, f, zc+s*4.0-0.9), V(-40, f+4.6, zc+s*4.0+0.9)))
		g.Carve(Block(V(-42, f, zc+s*6.0-0.6), V(-40, f+3.2, zc+s*6.0+0.6)))
	}
	// West and east Doric porticoes with entablature and pediments.
	for _, x := range []float64{-49.6, -37.6} {
		for _, o := range offs {
			a.doric(g, P(x, zc+o), f, 7.6, 0.68)
		}
		g.Fill(Block(V(x-1.0, f+7.6, zc-7.8), V(x+1.0, f+9.4, zc+7.8)))
		g.Paint(Stone, Func{Box{V(x-1.1, f+8.6, zc-7.9), V(x+1.1, f+9.4, zc+7.9)}, func(p Vec3) bool {
			return math.Mod(p.Z+100, 1.25) < 0.55
		}})
	}
	g.Use(Sand).Fill(a.gable(P(-50.6, zc-7.9), P(-49.0, zc+7.9), f+9.4, 0.3, 0, AxisX))
	g.Fill(Intersect(a.gable(P(-38.2, zc-7.9), P(-36.6, zc+7.9), f+9.4, 0.3, 0, AxisX),
		Func{Box{V(-39, f, zc-9), V(-36, f+12, zc+9)}, func(p Vec3) bool { return p.Z < zc-1.5 }}))
	// North wing (Pinakotheke) with a three-column porch toward the approach.
	g.Use(White).Fill(Block(V(-53.5, top, -17.5), V(-45.5, f, -8.6)))
	g.Fill(Subtract(Block(V(-53, f, -17), V(-46, f+6.4, -12)), Block(V(-52, f, -16), V(-47, f+7, -13))))
	g.Carve(Block(V(-50.4, f, -12.5), V(-48.6, f+4.2, -11.5)))
	g.Carve(Block(V(-53.5, f+2.4, -15.2), V(-52.5, f+4.2, -13.8)))
	for _, x := range []float64{-52.2, -49.5, -46.8} {
		a.doric(g, P(x, -9.6), f, 5.4, 0.55)
	}
	g.Fill(Block(V(-53.2, f+5.4, -17.2), V(-45.8, f+6.4, -8.8)))
	g.Use(Stone).Fill(HipRoof{P(-53.4, -17.4), P(-45.6, -8.6), f + 6.4})
	g.Carve(Func{Box{V(-54, f+6.4, -18), V(-45, f+12, -8)}, func(p Vec3) bool { return p.Y > f+7.6 }})
	g.Paint(DarkStone, Block(V(-52, f, -16), V(-47, f+6.4, -13)))
	// South wing.
	g.Use(White).Fill(Block(V(-51.5, top, 6.6), V(-45.5, f, 11)))
	g.Fill(Subtract(Block(V(-49, f, 7), V(-46, f+6, 10.6)), Block(V(-48.2, f, 7.8), V(-46.8, f+7, 9.8))))
	for _, z := range []float64{7.4, 10.2} {
		a.doric(g, P(-50.6, z), f, 5.0, 0.55)
	}
	g.Fill(Block(V(-51.2, f+5.0, 6.8), V(-46, f+6, 10.8)))
	g.Paint(Sand, a.where(g, Box{V(-55, top, -19), V(-35, top+14, 12)}, White, func(p Vec3) bool {
		return parthenonNoise3(p.X/2.5, p.Y/3, p.Z/2.5, 141) > 0.6
	}))
}

// nike is the Temple of Athena Nike on its bastion at the south-west corner.
func (a Parthenon) nike(g *Grid) {
	top := parthenonTop
	g.Use(Sand).Fill(Subtract(Block(V(-64, 0, 5.5), V(-52, top+0.4, 16.5)), Block(V(-63, 0, 6.5), V(-53, top-0.6, 15.5))))
	g.Paint(Stone, Func{Box{V(-64.5, 0, 5), V(-51.5, top+0.5, 17)}, func(p Vec3) bool {
		return math.Mod(p.Y, 1.6) < 0.45 || parthenonNoise3(p.X/2, p.Y/2, p.Z/2, 151) > 0.68
	}})
	g.Fill(Block(V(-64.3, top+0.4, 5.2), V(-51.7, top+1.3, 5.8)))
	f := top + 1.2
	g.Use(White).Fill(Block(V(-62.6, top+0.4, 7.0), V(-53.4, f, 15.0)))
	g.Fill(Subtract(Block(V(-60.4, f, 8.2), V(-55.6, f+4.6, 13.8)), Block(V(-59.4, f, 9.2), V(-56.6, f+5, 12.8))))
	for _, x := range []float64{-61.7, -54.3} {
		for k := 0; k < 4; k++ {
			a.ionic(g, P(x, 8.2+float64(k)*1.87), f, 4.6, 0.46, AxisZ)
		}
	}
	g.Fill(Block(V(-62.5, f+4.6, 7.4), V(-53.5, f+5.5, 14.6)))
	g.Use(Stone).Fill(a.gable(P(-62.7, 7.2), P(-53.3, 14.8), f+5.5, 0.32, 0.7, AxisX))
	g.Use(White).Fill(a.gable(P(-62.5, 7.4), P(-61.8, 14.6), f+5.5, 0.32, 0, AxisX))
	g.Fill(a.gable(P(-54.2, 7.4), P(-53.5, 14.6), f+5.5, 0.32, 0, AxisX))
	g.Paint(DarkStone, Block(V(-59.3, f, 9.3), V(-56.7, f+4.6, 12.7)))
}

// roma is the small round Temple of Rome and Augustus east of the Parthenon.
func (a Parthenon) roma(g *Grid) {
	c := P(43, -17)
	top := parthenonTop
	g.Use(White).Fill(Cylinder{Center: c, Radius: 4.4, Y0: top, Y1: top + 0.7})
	for k := 0; k < 9; k++ {
		q := Polar(c, 3.4, float64(k)*2*math.Pi/9)
		a.ionic(g, q, top+0.7, 5.2, 0.5, AxisX)
	}
	g.Fill(Cylinder{Center: c, Radius: 4.2, Inner: 2.6, Y0: top + 5.9, Y1: top + 6.8})
	g.Use(Stone).FillVolume(Frustum{c, 4.4, 0.8, top + 6.8, top + 8.6})
	g.Carve(Frustum{c, 3.4, 0.0, top + 6.8, top + 8.2})
}

func (a Parthenon) odeon(g *Grid) {
	o := parthenonOdeon
	x0, x1 := o.X-15, o.X+15
	zf := o.Y + 2.8 // Front of the stage; the scaenae frons wall stands behind it.
	// Stage platform.
	g.Use(Wood).Fill(Block(V(x0, 0, o.Y), V(x1, 1.5, zf)))
	g.Paint(Sand, Block(V(x0, 0, o.Y), V(x1, 1.0, zf)))
	// Scaenae frons: three tiers of arches, ragged top.
	g.Use(Sand).Fill(Func{Box{V(x0, 0, zf), V(x1, 13.5, zf+2)}, func(p Vec3) bool {
		return p.Y <= 12.6+0.9*Hash01(int(math.Floor(p.X)), 161)
	}})
	// The lower two tiers are deep niches with shadowed backs; the top tier opens to the sky.
	for k := 0; k < 7; k++ {
		x := x0 + 15.0/7 + float64(k)*30.0/7
		g.Carve(Arch{Sill: V(x, 0.3, zf+2), HalfWidth: 1.35, Spring: 3.4, Depth: 2.6, Through: AxisZ})
		g.Carve(Arch{Sill: V(x, 6.0, zf+2), HalfWidth: 1.1, Spring: 1.9, Depth: 2.6, Through: AxisZ})
		g.Paint(Charcoal, Arch{Sill: V(x, 0.3, zf+0.4), HalfWidth: 1.35, Spring: 3.4, Depth: 0.8, Through: AxisZ})
		g.Paint(Charcoal, Arch{Sill: V(x, 6.0, zf+0.4), HalfWidth: 1.1, Spring: 1.9, Depth: 0.8, Through: AxisZ})
		g.Carve(Arch{Sill: V(x, 9.9, zf+1), HalfWidth: 0.95, Spring: 1.1, Depth: 3, Through: AxisZ})
	}
	// Cornice bands between the tiers.
	g.Use(Stone).Fill(Block(V(x0, 5.2, zf+1.9), V(x1, 5.7, zf+2.4)), Block(V(x0, 9.2, zf+1.9), V(x1, 9.6, zf+2.4)))
	// Side wings (paraskenia) projecting toward the cavea.
	for _, side := range []float64{-1, 1} {
		xa, xb := x0-3.5, x0
		if side > 0 {
			xa, xb = x1, x1+3.5
		}
		g.Use(Sand).Fill(Subtract(Block(V(xa, 0, o.Y-3.5), V(xb, 11.5, zf+2)), Block(V(xa+1, 0.5, o.Y-2.5), V(xb-1, 12, zf+1))))
		mid := (xa + xb) / 2
		g.Carve(Arch{Sill: V(mid, 0.3, zf+1.5), HalfWidth: 0.9, Spring: 3.2, Depth: 2, Through: AxisZ})
		for _, z := range []float64{o.Y - 1, o.Y + 2} {
			g.Carve(Arch{Sill: V(mid, 6.2, z), HalfWidth: 0.8, Spring: 1.8, Depth: 8, Through: AxisX})
		}
		g.Paint(DarkStone, Block(V(xa+1, 0.5, o.Y-2.5), V(xb-1, 11.5, zf+1)))
	}
	// Weathered blocks and the dark stage interior.
	g.Paint(Stone, a.where(g, Box{V(x0-4, 0, o.Y-4), V(x1+4, 14, zf+2.5)}, Sand, func(p Vec3) bool {
		return parthenonNoise3(p.X/2, p.Y/2, p.Z/2, 163) > 0.66
	}))
}

// stoa is the ruined Stoa of Eumenes: the arcaded retaining wall of its terrace and
// the stumps of its front colonnade.
func (a Parthenon) stoa(g *Grid) {
	const x0, x1, zb = -7.5, 13.0, 36.4
	g.Use(Sand).Fill(Func{Box{V(x0, 0, zb-1.4), V(x1, 7.5, zb)}, func(p Vec3) bool {
		return p.Y <= 5.6+1.4*parthenonNoise2(p.X/2.5, 0, 191)
	}})
	for x := x0 + 1.6; x < x1-1; x += 2.9 {
		g.Carve(Arch{Sill: V(x, 1.2, zb), HalfWidth: 1.0, Spring: 2.0, Depth: 1.6, Through: AxisZ})
		g.Paint(Charcoal, Arch{Sill: V(x, 1.2, zb-1.2), HalfWidth: 1.0, Spring: 2.0, Depth: 0.6, Through: AxisZ})
	}
	g.Paint(Stone, a.where(g, Box{V(x0, 0, zb-1.5), V(x1, 7.5, zb+0.1)}, Sand, func(p Vec3) bool {
		return math.Mod(p.Y, 1.4) < 0.4 || parthenonNoise3(p.X/2, p.Y/2, p.Z/2, 193) > 0.68
	}))
	g.Use(Stone).Fill(Block(V(x0, 0.5, 41.3), V(x1, 1.5, 42.6)))
	g.Use(White)
	for k, x := 0, x0+0.8; x < x1; k, x = k+1, x+2.5 {
		if Hash01(k, 195) < 0.3 {
			continue
		}
		g.FillVolume(Cylinder{Center: P(x, 41.95), Radius: 0.55, Y0: 1.5, Y1: 2.3 + 2.5*Hash01(k, 197)})
	}
}

// flag is the Greek flag at the Belvedere on the east tip.
func (a Parthenon) flag(g *Grid) {
	base := P(52, -8)
	top := parthenonTop
	g.Use(Stone).Fill(BlockAt(base, 2.6, 2.6, top, top+1.0))
	g.Use(White).Fill(Cylinder{Center: base, Radius: 0.45, Y0: top + 1, Y1: top + 11})
	g.Use(Navy).Fill(Block(V(base.X+0.4, top+7.0, base.Y-0.4), V(base.X+6.2, top+10.8, base.Y+0.4)))
	g.Paint(White, Func{Box{V(base.X+0.4, top+7.0, base.Y-0.5), V(base.X+6.2, top+10.8, base.Y+0.5)}, func(p Vec3) bool {
		band := int(math.Floor((top + 10.8 - p.Y) / (3.8 / 9)))
		inCanton := p.X < base.X+2.6 && p.Y > top+10.8-3.8*5/9
		if inCanton {
			return math.Abs(p.X-(base.X+1.5)) < 0.42 || math.Abs(p.Y-(top+10.8-3.8*5/18)) < 0.25
		}
		return band%2 == 1
	}})
}

// olive places a gnarled olive tree.
func (a Parthenon) olive(g *Grid, c Vec2, y float64, seed int) {
	lean := P(Hash01(seed, 171)-0.5, Hash01(seed, 172)-0.5).Scale(1.2)
	h := 2.0 + 0.8*Hash01(seed, 173)
	g.Use(Wood).Fill(Capsule{c.At(y - 0.3), c.Add(lean).At(y + h), 0.5})
	crown := c.Add(lean).At(y + h + 0.9)
	r := 1.7 + 0.6*Hash01(seed, 174)
	g.Use(Foliage).FillVolume(Ellipsoid{crown, V(r, r*0.62, r*0.9)})
	g.FillVolume(Ellipsoid{crown.Add(V(r*0.5, 0.4, -r*0.3)), V(r*0.7, r*0.5, r*0.7)})
	// Silvery leaf undersides catch the light.
	g.Paint(Stone, a.where(g, Box{crown.Sub(V(r*2, r, r*2)), crown.Add(V(r*2, r, r*2))}, Foliage, func(p Vec3) bool {
		return p.Y > crown.Y-0.2 && Hash01(int(math.Floor(p.X*2)), int(math.Floor(p.Y*2)), int(math.Floor(p.Z*2))) > 0.8
	}))
}

// olives scatters olive groves over the slopes.
func (a Parthenon) olives(g *Grid) {
	for gx := -80.0; gx <= 74; gx += 4.6 {
		for gz := -52.0; gz <= 48; gz += 4.6 {
			k := int(gx*7 + gz*131)
			x := gx + 3*(Hash01(k, 181)-0.5)
			z := gz + 3*(Hash01(k, 182)-0.5)
			d := a.edge(x, z)
			if d < 6 || d > 15 || Hash01(k, 183) > 0.45 {
				continue
			}
			if x > -48 && x < -4 && z > 24 || x < -48 && z > -12 && z < 20 || a.inStair(x, z) ||
				x > -10 && x < 15 && z > 33 || P(x, z).Sub(parthenonDionysus).Length() < 13.5 {
				continue
			}
			h := a.height(x, z)
			if h < 0.4 {
				continue
			}
			a.olive(g, P(x, z), h, k)
		}
	}
}
