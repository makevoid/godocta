package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// PetraTreasury is Al-Khazneh, the Treasury of Petra, carved into a rose-red sandstone
// cliff where the Siq gorge opens into a small plaza. The two-storey Hellenistic façade
// stands in a recess cut into the dressed cliff, facing +Z: a stepped podium, a portico
// of six Corinthian columns with a dark vestibule and doorways into the tomb chambers,
// the entablature and pediment; above, side pavilions under a broken pediment flank the
// round tholos with its conical roof and urn, with deep shadowed recesses between them.
// The cliff is a hollow banded rock mass (brick, terracotta, sand, red and dark strata)
// wrapping round the plaza on three sides, with eroded stepped outer faces that come
// down to a low sandy rim. The narrow winding Siq cuts through the eastern massif into
// the plaza; a sandy wadi opens to the front. Camels, people and a donkey cart wait in
// the plaza. East (+X) is to the right.
type PetraTreasury struct{}

func init() { Register(PetraTreasury{}) }

func (PetraTreasury) Name() string  { return "petra_treasury" }
func (PetraTreasury) Title() string { return "Petra Treasury" }

const (
	petraTreasuryX0, petraTreasuryX1 = -37.0, 37.0
	petraTreasuryZ0, petraTreasuryZ1 = -36.0, 20.0
	petraTreasuryFloor               = 1.0
	petraTreasurySkin                = 1.25
)

// petraTreasuryValley is the plaza and the wadi opening to the front, in plan. Its back
// edge, z = -8, is the dressed cliff face that carries the façade.
var petraTreasuryValley = []Vec2{
	P(-14.5, -8), P(14.5, -8), P(16, -5.5), P(16.5, 0), P(17, 6), P(18.5, 14), P(20, 23),
	P(-20, 23), P(-18.5, 14), P(-17, 6), P(-16.5, 0), P(-16, -5.5),
}

// petraTreasurySiq is the winding gorge from the east edge into the plaza.
var petraTreasurySiq = []Vec2{
	P(39, -23), P(30, -19.5), P(27.5, -15.5), P(25.5, -11), P(22.5, -9.5), P(21, -5.5), P(18.5, -2.5), P(15, -0.5),
}

func (PetraTreasury) smooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func (PetraTreasury) noise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

func (s PetraTreasury) fbm(x, z float64, seed int) float64 {
	return (s.noise(x, z, seed)*4 + s.noise(x*2.1+5, z*2.1+3, seed+1)*2 + s.noise(x*4.3+1, z*4.3+7, seed+2)) / 7
}

func (PetraTreasury) segDist(p, a, b Vec2) float64 {
	ab := b.Sub(a)
	t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
	return p.Sub(a.Add(ab.Scale(t))).Length()
}

// polyDist is the signed distance to a polygon: negative inside.
func (s PetraTreasury) polyDist(poly []Vec2, p Vec2) float64 {
	d := math.Inf(1)
	for i, j := 0, len(poly)-1; i < len(poly); j, i = i, i+1 {
		d = math.Min(d, s.segDist(p, poly[j], poly[i]))
	}
	if InsidePolygon(poly, p) {
		return -d
	}
	return d
}

func (s PetraTreasury) siqDist(p Vec2) float64 {
	d := math.Inf(1)
	for i := 0; i+1 < len(petraTreasurySiq); i++ {
		d = math.Min(d, s.segDist(p, petraTreasurySiq[i], petraTreasurySiq[i+1]))
	}
	return d
}

// steps is an eroded terrace profile: d units from the foot, risers of the given height
// every bench units, with bench edges wandering by the jitter in [0, 1).
func (PetraTreasury) steps(d, bench, riser, jitter float64) float64 {
	if d <= 0 {
		return 0
	}
	return riser * (1 + math.Floor(d/bench+jitter))
}

// height is the rock surface; the floor level is the plaza and rim.
func (s PetraTreasury) height(x, z float64) float64 {
	p := P(x, z)
	// Massif top: high behind the façade, sinking toward the front, with rounded domes.
	fz := s.smooth((math.Abs(x)-10)/6) + s.smooth((z+4)/4)
	wf := 0.3 + 0.7*math.Min(1, fz)
	top := 46 + 14*(s.fbm(x/11, z/11, 1)-0.5)*wf + 6*(s.noise(x/5, z/5, 16)-0.5)*wf - 0.65*math.Max(0, z+5)
	// Inner cliffs around the plaza, wadi and Siq.
	facade := s.smooth((math.Abs(x)-12)/4) + s.smooth((z+5)/4)
	facade = math.Min(1, facade)
	dv := s.polyDist(petraTreasuryValley, p) + facade*2.2*(s.fbm(x/4, z/4, 3)-0.5)
	w := 1.7 + 0.6*s.noise(x/3, z/3, 4)
	dv = math.Min(dv, s.siqDist(p)-w)
	inner := 0.0
	if dv > 0 {
		rough := s.steps(dv, 0.85+0.5*s.noise(x/5, z/5, 5), 7, s.noise(x/2.5, z/2.5, 6)) + 2*(s.noise(x/3, z/3, 7)-0.5)
		// Above the façade the dressed face ends and the cliff steps back in eroded ledges.
		dressed := 37 + s.steps(dv-0.6, 1.1+0.6*s.noise(x/3, z/3, 12), 3.2, s.noise(x/2, z/2, 13)) + 1.5*(s.noise(x/2.5, z/2.5, 14)-0.5)
		inner = facade*rough + (1-facade)*dressed
	}
	// Outer faces step down to the sandy rim.
	de := math.Min(math.Min(x-petraTreasuryX0, petraTreasuryX1-x), math.Min(z-petraTreasuryZ0, petraTreasuryZ1-z))
	de += 4*(s.fbm(x/7, z/7, 8)-0.5) + 7*(s.fbm(x/16, z/16, 15)-0.5)
	// A broad talus of low ledges, then steeper eroded cliffs.
	jt := s.noise(x/3, z/3, 10)
	outer := s.steps(math.Min(de-1.5, 7), 2.3, 2.3, jt) + s.steps(de-8.5, 1.2+0.9*s.noise(x/6, z/6, 9), 6.2, jt) + 1.5*(s.noise(x/2, z/2, 11)-0.5)
	h := petraTreasuryFloor + math.Max(0, math.Min(math.Min(top, inner), outer))
	if h < petraTreasuryFloor+1.6 {
		return petraTreasuryFloor
	}
	return h
}

// petraTreasuryField is the sampled heightfield with bilinear lookup.
type petraTreasuryField struct {
	x0, z0, step float64
	nx, nz       int
	h            []float64
}

func (m *petraTreasuryField) at(x, z float64) float64 {
	fx, fz := (x-m.x0)/m.step, (z-m.z0)/m.step
	if fx < 0 || fz < 0 || fx >= float64(m.nx-1) || fz >= float64(m.nz-1) {
		return petraTreasuryFloor
	}
	i, j := int(fx), int(fz)
	tx, tz := fx-float64(i), fz-float64(j)
	k := j*m.nx + i
	a := m.h[k] + (m.h[k+1]-m.h[k])*tx
	b := m.h[k+m.nx] + (m.h[k+m.nx+1]-m.h[k+m.nx])*tx
	return a + (b-a)*tz
}

// open removes peaks and slots narrower than r samples: erosion, then dilation.
func (m *petraTreasuryField) open(r int) {
	pass := func(src []float64, pick func(a, b float64) float64) []float64 {
		dst := make([]float64, len(src))
		for j := 0; j < m.nz; j++ {
			for i := 0; i < m.nx; i++ {
				v := src[j*m.nx+i]
				for dj := -r / 2; dj <= r/2; dj++ {
					for di := -r / 2; di <= r/2; di++ {
						ii, jj := i+di, j+dj
						if ii >= 0 && jj >= 0 && ii < m.nx && jj < m.nz {
							v = pick(v, src[jj*m.nx+ii])
						}
					}
				}
				dst[j*m.nx+i] = v
			}
		}
		return dst
	}
	m.h = pass(pass(m.h, math.Min), math.Max)
}

// petraTreasuryDirs are the 14 lattice neighbor directions, normalized.
var petraTreasuryDirs = func() []Vec3 {
	var d []Vec3
	for _, n := range Neighbors {
		v := V(float64(n.X), float64(n.Y), float64(n.Z))
		d = append(d, v.Scale(1/v.Length()))
	}
	return d
}()

// carves are the rooms cut into the cliff: the façade recess, vestibule, chambers and
// shadowed recesses of the upper storey.
func (PetraTreasury) carves() []Solid {
	var c []Solid
	c = append(c,
		Block(V(-12.5, 1.01, -11.5), V(12.5, 35.5, -7)),               // Façade recess.
		Block(V(-6.9, 2.2, -14.5), V(6.9, 12.5, -11)),                 // Vestibule.
		Block(V(-1.5, 2.2, -16.4), V(1.5, 9.2, -14)),                  // Central door.
		Block(V(-6, 2.2, -24), V(6, 10, -16)),                         // Tomb chamber.
		Block(V(8.3, 2.2, -14.5), V(9.9, 7.4, -11)),                   // East side door.
		Block(V(-9.9, 2.2, -14.5), V(-8.3, 7.4, -11)),                 // West side door.
		Block(V(7.4, 2.2, -19), V(11.4, 8, -14.5)),                    // East side chamber.
		Block(V(-11.4, 2.2, -19), V(-7.4, 8, -14.5)),                  // West side chamber.
		Block(V(3.6, 18.3, -13.6), V(5.3, 31, -11)),                   // Upper recesses.
		Block(V(-5.3, 18.3, -13.6), V(-3.6, 31, -11)),                 //
		Block(V(-12.5, 35.5, -9.5), V(12.5, 37.5, -7)),                // Cut-back lip above the recess.
		Block(V(-3.2, 18.3, -13), V(3.2, 34.5, -11)),                  // Tholos roof clearance.
		Frustum{P(0, -11.5), 3.2, 1.6, 34.5, 36.5},                    //
		Block(V(-13.2, petraTreasuryFloor+0.01, -8), V(13.2, 42, -6)), // Clean façade cliff line.
	)
	for _, t := range petraTreasuryTombs {
		c = append(c, t.local(Block(V(-t.w/2-0.8, petraTreasuryFloor+0.01, -2.5), V(t.w/2+0.8, t.h+1.2, 4))),
			t.local(Block(V(-1.8, petraTreasuryFloor+0.6, -6.5), V(1.8, 6.5, -2.4))))
	}
	return c
}

func (s PetraTreasury) Build(g *Grid) Order {
	mp := &petraTreasuryField{x0: petraTreasuryX0 - 2, z0: petraTreasuryZ0 - 2, step: 0.25}
	mp.nx, mp.nz = int((petraTreasuryX1-petraTreasuryX0+4)/mp.step)+2, int((petraTreasuryZ1-petraTreasuryZ0+4)/mp.step)+2
	mp.h = make([]float64, mp.nx*mp.nz)
	for j := 0; j < mp.nz; j++ {
		for i := 0; i < mp.nx; i++ {
			mp.h[j*mp.nx+i] = s.height(mp.x0+float64(i)*mp.step, mp.z0+float64(j)*mp.step)
		}
	}
	mp.open(11)
	s.cliff(g, mp)
	s.facade(g)
	for _, t := range petraTreasuryTombs {
		s.tomb(g, t)
	}
	s.figures(g)
	return Outward(P(0, -9))
}

// cliff fills the banded rock and floor as a hollow skin, then colors it.
func (s PetraTreasury) cliff(g *Grid, mp *petraTreasuryField) {
	carves := s.carves()
	solid := func(p Vec3) bool {
		if p.Y < 0 {
			return p.X >= petraTreasuryX0 && p.X <= petraTreasuryX1 && p.Z >= petraTreasuryZ0 && p.Z <= petraTreasuryZ1
		}
		if p.X < petraTreasuryX0 || p.X > petraTreasuryX1 || p.Z < petraTreasuryZ0 || p.Z > petraTreasuryZ1 {
			return false
		}
		if p.Y > mp.at(p.X, p.Z) {
			return false
		}
		if p.Y > petraTreasuryFloor {
			for _, c := range carves {
				if c.Bounds().Contains(p) && c.Contains(p) {
					return false
				}
			}
		}
		return true
	}
	box := Box{V(petraTreasuryX0, 0, petraTreasuryZ0), V(petraTreasuryX1, 60, petraTreasuryZ1)}
	t := NewGridWithSpan(g.Span())
	t.Fill(Func{box, func(p Vec3) bool {
		if !solid(p) {
			return false
		}
		for _, d := range petraTreasuryDirs {
			if !solid(p.Add(d.Scale(petraTreasurySkin))) {
				return true
			}
		}
		return false
	}})
	for _, c := range t.Cells() {
		p := t.World(c)
		g.Use(s.rockColor(p, mp.at(p.X, p.Z), c)).Add(c)
	}
}

// rockColor chooses floor sand, dressed rose stone, dark chamber interiors and banded strata.
func (s PetraTreasury) rockColor(p Vec3, h float64, c Cell) Color {
	n := Hash01(int(c.X), int(c.Y), int(c.Z), 31)
	if h <= petraTreasuryFloor+0.01 || p.Y < petraTreasuryFloor+0.2 {
		if n < 0.1 {
			return Terracotta
		}
		return Sand
	}
	// Chambers behind the doors.
	if p.Z < -14.9 && p.Z > -25.4 && math.Abs(p.X) < 7.4 && p.Y < 11.5 || p.Z < -14 && p.Z > -20.4 && math.Abs(p.X) > 6 && math.Abs(p.X) < 12.8 && p.Y < 9.4 {
		return Charcoal
	}
	if p.Z < -11.2 && p.Z > -15 && math.Abs(p.X) > 7.2 && math.Abs(p.X) < 11 && p.Y < 8.2 {
		return Charcoal
	}
	if p.Z < -11.2 && p.Z > -15 && math.Abs(p.X) > 3.3 && math.Abs(p.X) < 5.6 && p.Y > 18 && p.Y < 31.5 {
		return DarkStone
	}
	if p.Z < -11.2 && p.Z > -16 && math.Abs(p.X) < 7.2 && p.Y < 12.6 && p.Y > 2 {
		return Brick
	}
	// Dressed rock of the recess.
	if p.Z < -6.8 && p.Z > -14 && math.Abs(p.X) < 12.8 && p.Y < 37.8 {
		if n < 0.12 {
			return Brick
		}
		return Terracotta
	}
	b := (p.Y + 4*s.fbm(p.X/8, p.Z/8, 21) + 1.5*math.Sin((p.X+0.6*p.Z)/6) + 0.6*s.noise(p.X/2, p.Z/2, 22)) / 1.3
	if n < 0.12 {
		b += 1
	}
	k := Hash01(int(math.Floor(b)), 23)
	if p.Y > h-1 && k >= 0.88 {
		k = 0.7 // Weathered tops show no dark strata.
	}
	switch {
	case k < 0.33:
		return Brick
	case k < 0.62:
		return Terracotta
	case k < 0.78:
		return Sand
	case k < 0.88:
		return Red
	default:
		return Wood
	}
}

// column is a Corinthian column: base, shaft and flared sand capital.
func (PetraTreasury) column(g *Grid, c Vec2, r, y0, y1 float64) {
	g.With(Terracotta, func() {
		g.Fill(Cylinder{Center: c, Radius: r + 0.3, Y0: y0, Y1: y0 + 0.6})
		g.Fill(Cylinder{Center: c, Radius: r, Y0: y0, Y1: y1 - 1.3})
	})
	g.With(Sand, func() {
		g.Fill(Frustum{c, r + 0.05, r + 0.4, y1 - 1.3, y1})
	})
}

// facade builds the two storeys in the recess; the recess back is z = -11.5.
func (s PetraTreasury) facade(g *Grid) {
	const zb = -11.5
	g.Use(Terracotta)
	// Podium steps.
	g.Fill(Block(V(-12.5, 1, zb), V(12.5, 1.6, -6.6)))
	g.Fill(Block(V(-12.5, 1.6, zb), V(12.5, 2.2, -7.3)))
	// Antae walls of the outer bays, with the side doors.
	g.Fill(Block(V(6.9, 2.2, zb), V(12.5, 12.5, -9.6)), Block(V(-12.5, 2.2, zb), V(-6.9, 12.5, -9.6)))
	g.Carve(Block(V(8.3, 2.2, zb-0.5), V(9.9, 7.4, -9.4)), Block(V(-9.9, 2.2, zb-0.5), V(-8.3, 7.4, -9.4)))
	// Lintel frames over the side doors.
	g.With(Sand, func() {
		g.Fill(Block(V(7.9, 7.4, -9.9), V(10.3, 8.2, -9.3)), Block(V(-10.3, 7.4, -9.9), V(-7.9, 8.2, -9.3)))
		g.Fill(Block(V(-2.1, 9.2, -14.5), V(2.1, 10, -13.9)))
	})
	// Equestrian reliefs: rider figures in the outer bays.
	for _, sx := range []float64{-1, 1} {
		x := sx * 11.6
		g.With(Sand, func() {
			g.Fill(Ellipsoid{V(x, 4.2, -9.4), V(0.45, 0.9, 0.35)}, Sphere(V(x, 5.5, -9.4), 0.38))
		})
	}
	// Six columns.
	for _, x := range []float64{-10.6, -5.9, -2, 2, 5.9, 10.6} {
		s.column(g, P(x, -8.7), 0.72, 2.2, 12.5)
	}
	// Lower entablature: architrave, sand frieze, projecting cornice.
	g.Fill(Block(V(-12.5, 12.5, zb), V(12.5, 13.3, -7.9)))
	g.With(Sand, func() { g.Fill(Block(V(-12.5, 13.3, zb), V(12.5, 14.1, -8.1))) })
	g.Fill(Block(V(-12.5, 14.1, zb), V(12.5, 14.8, -7.4)))
	// Frieze ornaments: dark rosettes between triglyph blocks.
	for i := -6; i <= 6; i++ {
		g.Paint(Brick, Block(V(float64(i)*1.9-0.3, 13.3, -8.4), V(float64(i)*1.9+0.3, 14.1, -7.9)))
	}
	// Pediment over the central four columns, with a recessed tympanum.
	ped := func(y0, half, rise, inset float64) func(Vec3) bool {
		return func(p Vec3) bool {
			return math.Abs(p.X) <= half*(1-(p.Y-y0)/rise)-inset
		}
	}
	g.Fill(Func{Box{V(-7.6, 14.8, zb), V(7.6, 18.6, -7.5)}, ped(14.8, 7.6, 3.8, 0)})
	g.Paint(Sand, Func{Box{V(-7.6, 14.8, -8.4), V(7.6, 18.6, -7.5)}, ped(14.8, 7.6, 3.8, 0)})
	g.Carve(Func{Box{V(-7.6, 15.5, -8.4), V(7.6, 18.6, -7)}, ped(14.8, 7.6, 3.8, 1.3)})
	g.With(Sand, func() {
		g.Fill(Block(V(-0.5, 18.4, -9.4), V(0.5, 19.6, -8.4))) // Akroterion.
		g.Fill(Sphere(V(0, 16.2, -8.4), 0.6))                  // Tympanum boss.
		g.Fill(Block(V(-8, 14.8, -9), V(-6.9, 16, -8)), Block(V(6.9, 14.8, -9), V(8, 16, -8)))
	})
	// Attic between the storeys.
	g.Fill(Block(V(-12.5, 14.8, zb), V(12.5, 17.6, -9.4)))
	g.With(Sand, func() { g.Fill(Block(V(-12.5, 17.6, zb), V(12.5, 18.3, -8.6))) })

	// Upper storey: side pavilions.
	for _, sx := range []float64{-1, 1} {
		x0, x1 := sx*5.3, sx*12.5
		lo, hi := math.Min(x0, x1), math.Max(x0, x1)
		g.Fill(Block(V(lo, 18.3, zb), V(hi, 27, -9.6)))
		// Statue niches between the pavilion columns.
		nx := sx * 8.9
		g.Carve(Block(V(nx-1.2, 19.4, -10.4), V(nx+1.2, 25.4, -9)))
		g.Paint(DarkStone, Block(V(nx-1.2, 19.4, -10.9), V(nx+1.2, 25.4, -10.3)))
		g.With(Sand, func() {
			g.Fill(Capsule{V(nx, 19.8, -10), V(nx, 23.6, -10), 0.5}, Sphere(V(nx, 24.4, -10), 0.45))
		})
		for _, cx := range []float64{sx * 6.2, sx * 11.6} {
			s.column(g, P(cx, -9), 0.6, 18.3, 27)
		}
		// Entablature.
		g.Fill(Block(V(lo, 27, zb), V(hi, 27.7, -8.5)))
		g.With(Sand, func() { g.Fill(Block(V(lo, 27.7, zb), V(hi, 28.4, -8.2))) })
		// Broken pediment half, rising toward the tholos.
		sxx := sx
		g.Fill(Func{Box{V(lo, 28.4, zb), V(hi, 31.8, -8.5)}, func(p Vec3) bool {
			u := (12.5 - p.X*sxx) / 7.2
			return p.Y <= 28.4+3.4*u
		}})
		g.Paint(Sand, Func{Box{V(lo, 28.4, -9.2), V(hi, 31.8, -8.5)}, func(p Vec3) bool {
			u := (12.5 - p.X*sxx) / 7.2
			return p.Y >= 28.4+3.4*u-0.8
		}})
		g.With(Sand, func() { g.Fill(Block(V(sx*5.3-0.5, 31.8, -9.6), V(sx*5.3+0.5, 32.8, -8.6))) })
	}
	// Tholos: drum, ring of columns, entablature, conical roof, capital and urn.
	tc := P(0, -11)
	g.Fill(Cylinder{Center: tc, Radius: 2.3, Y0: 18.3, Y1: 27})
	g.Carve(Arch{Sill: V(0, 18.3, -8.9), HalfWidth: 0.85, Spring: 3.2, Depth: 1.4, Through: AxisZ})
	g.Paint(Charcoal, Arch{Sill: V(0, 18.3, -9.6), HalfWidth: 0.85, Spring: 3.2, Depth: 0.9, Through: AxisZ})
	for _, deg := range []float64{12, 48, 132, 168} {
		c := Polar(tc, 3.05, deg*math.Pi/180)
		s.column(g, c, 0.6, 18.3, 27)
	}
	ring := Subtract(Cylinder{Center: tc, Radius: 3.75, Y0: 27, Y1: 28.4}, Block(V(-4, 26, -16), V(4, 30, zb)))
	g.Fill(ring)
	g.Paint(Sand, Cylinder{Center: tc, Radius: 3.75, Inner: 3.2, Y0: 27.6, Y1: 28.4})
	g.FillVolume(Subtract(Frustum{tc, 3.5, 0.5, 28.4, 32}, Block(V(-4, 26, -16), V(4, 33, -12.2))))
	g.With(Sand, func() {
		g.FillVolume(Frustum{tc, 0.8, 1.3, 32, 32.8})
		g.FillVolume(Cylinder{Center: tc, Radius: 0.55, Y0: 32.8, Y1: 33.4})
		g.FillVolume(Ellipsoid{V(0, 34.4, -11), V(1.15, 1.2, 1.15)})
		g.FillVolume(Cylinder{Center: tc, Radius: 0.5, Y0: 35.3, Y1: 36.1})
	})
}

// place turns a local solid (heading 0 faces +X) to the heading and moves it to pos.
func (PetraTreasury) place(sol Solid, pos Vec2, heading float64) Solid {
	return RotateY(Translate(sol, V(pos.X, 0, pos.Y)), pos, heading)
}

func (s PetraTreasury) camel(g *Grid, pos Vec2, heading float64, kneel bool, body, cloth Color) {
	y := petraTreasuryFloor
	if kneel {
		y -= 0.9
	}
	g.With(body, func() {
		g.FillVolume(s.place(Union(
			Ellipsoid{V(0, y+2.2, 0), V(1.5, 0.75, 0.65)},
			Ellipsoid{V(-0.1, y+2.85, 0), V(0.75, 0.6, 0.5)},
			Capsule{V(1.2, y+2.3, 0), V(2, y+3.2, 0), 0.42},
			Ellipsoid{V(2.3, y+3.35, 0), V(0.6, 0.36, 0.33)},
		), pos, heading))
		if kneel {
			g.FillVolume(s.place(Ellipsoid{V(0, y+1.4, 0), V(1.6, 0.6, 0.85)}, pos, heading))
			return
		}
		for _, lx := range []float64{-1, 1} {
			for _, lz := range []float64{-0.35, 0.35} {
				g.Fill(s.place(Capsule{V(lx, y+1.9, lz), V(lx, y+0.2, lz), 0.34}, pos, heading))
			}
		}
	})
	g.Paint(cloth, s.place(Ellipsoid{V(-0.1, y+3.0, 0), V(0.9, 0.5, 0.75)}, pos, heading))
}

func (s PetraTreasury) person(g *Grid, pos Vec2, robe, head Color) {
	y := petraTreasuryFloor
	g.With(robe, func() { g.FillVolume(Frustum{pos, 0.55, 0.38, y, y + 1.9}) })
	g.With(head, func() { g.FillVolume(Sphere(pos.At(y+2.3), 0.42)) })
}

func (s PetraTreasury) figures(g *Grid) {
	// Camels resting in the plaza and a caravan arriving from the Siq.
	s.camel(g, P(-9, -2), 0.3, true, Wood, Red)
	s.camel(g, P(-11.5, 1.5), -0.2, true, Sand, Navy)
	s.camel(g, P(9.5, 3), math.Pi, false, Wood, Red)
	s.camel(g, P(13.5, 0.5), math.Pi*0.9, false, Wood, Navy)
	s.camel(g, P(4.5, 9), -2.4, false, Sand, Red)
	// People.
	for _, p := range []struct {
		pos        Vec2
		robe, head Color
	}{
		{P(-2.5, -4.5), White, Charcoal}, {P(-1, -3.8), Navy, Sand}, {P(3, -5), Red, Charcoal},
		{P(7.5, 3.5), White, Red}, {P(-6, 3), Navy, White}, {P(1, 2), Charcoal, Sand},
		{P(-4, 8), White, Charcoal}, {P(11.5, -1.5), Red, White}, {P(-13, -4.5), White, Red},
		{P(0, 12), Navy, Sand}, {P(2.5, 14), White, Charcoal},
	} {
		s.person(g, p.pos, p.robe, p.head)
	}
	// Donkey cart.
	cp, ch := P(-6, 11), -0.4
	g.With(Wood, func() {
		g.Fill(s.place(Block(V(-1.4, 2, -0.8), V(1.4, 2.6, 0.8)), cp, ch))
		g.Fill(s.place(Block(V(-1.4, 2.6, -0.8), V(1.4, 3.2, -0.5)), cp, ch), s.place(Block(V(-1.4, 2.6, 0.5), V(1.4, 3.2, 0.8)), cp, ch))
		g.Fill(s.place(Capsule{V(1.2, 2.1, -0.6), V(3.4, 2.4, -0.6), 0.2}, cp, ch), s.place(Capsule{V(1.2, 2.1, 0.6), V(3.4, 2.4, 0.6), 0.2}, cp, ch))
	})
	g.With(DarkStone, func() {
		for _, wz := range []float64{-1.05, 1.05} {
			wz := wz
			g.Fill(s.place(Func{Box{V(-1, 1, wz-0.2), V(1, 3, wz+0.2)}, func(p Vec3) bool {
				return math.Hypot(p.X, p.Y-2) <= 0.95
			}}, cp, ch))
		}
	})
	g.With(Sand, func() { g.Fill(s.place(Block(V(-1.2, 2.6, -0.4), V(0.6, 3.3, 0.4)), cp, ch)) })
	g.With(DarkStone, func() {
		g.FillVolume(s.place(Union(
			Ellipsoid{V(4.2, 2.5, 0), V(1, 0.55, 0.45)},
			Capsule{V(5, 2.6, 0), V(5.5, 3.3, 0), 0.3},
			Ellipsoid{V(5.8, 3.3, 0), V(0.45, 0.28, 0.25)},
		), cp, ch))
		for _, lx := range []float64{3.6, 4.8} {
			for _, lz := range []float64{-0.25, 0.25} {
				g.Fill(s.place(Capsule{V(lx, 2.2, lz), V(lx, 1.1, lz), 0.3}, cp, ch))
			}
		}
	})
	// Bedouin tent: goat-hair awning on poles over red rugs.
	tp := P(-12.5, 15)
	g.With(Wood, func() {
		for _, dx := range []float64{-2.6, 0, 2.6} {
			for _, dz := range []float64{-1.6, 1.6} {
				g.Fill(Capsule{V(tp.X+dx, 1, tp.Y+dz), V(tp.X+dx, 3.6, tp.Y+dz), 0.3})
			}
		}
	})
	g.With(Charcoal, func() {
		g.Fill(Func{Box{V(tp.X-3.2, 3.2, tp.Y-2.4), V(tp.X+3.2, 4.6, tp.Y+2.2)}, func(p Vec3) bool {
			return p.Y <= 4.4-0.5*math.Abs(p.Z-tp.Y) && p.Y >= 3.9-0.5*math.Abs(p.Z-tp.Y)
		}})
		g.Fill(Block(V(tp.X-3.2, 1, tp.Y-2.3), V(tp.X+3.2, 3.6, tp.Y-1.9)))
	})
	g.Paint(Red, Block(V(tp.X-3.3, 3, tp.Y-2.5), V(tp.X+3.3, 4.6, tp.Y-1.5)))
	g.With(Red, func() { g.Fill(Block(V(tp.X-2.2, 1, tp.Y-1.2), V(tp.X+2.2, 1.3, tp.Y+1.6))) })
	s.camel(g, P(-7.5, 18), 2.2, true, Sand, Navy)
	s.person(g, P(-10, 18), White, Red)
	// Fallen boulders at the cliff feet.
	for i, b := range []Vec2{P(-15.5, -1), P(-17.5, 12.5), P(16.8, 3), P(18.5, 18), P(-20, 19), P(13.5, -6.8), P(21, 21.5), P(-15.2, -7)} {
		r := 0.9 + 0.8*Hash01(i, 51)
		c := []Color{Brick, Terracotta, Sand}[i%3]
		g.With(c, func() {
			g.FillVolume(Ellipsoid{V(b.X, petraTreasuryFloor+r*0.5, b.Y), V(r*1.3, r*0.9, r)})
		})
	}
	// Oleander and tamarisk shrubs along the wadi.
	for i, sp := range []Vec2{P(-15.5, 4), P(-16.5, 9), P(16, 8), P(17.5, 16), P(-18, 20), P(15, -4), P(-14.5, -6.5), P(18.5, 22)} {
		r := 0.9 + 0.6*Hash01(i, 41)
		g.With(Foliage, func() { g.FillVolume(Ellipsoid{V(sp.X, petraTreasuryFloor+r*0.7, sp.Y), V(r*1.2, r, r*1.2)}) })
		if Hash01(i, 42) < 0.5 {
			g.Paint(Red, Sphere(V(sp.X, petraTreasuryFloor+r*1.5, sp.Y), r*0.6))
		}
	}
}

// petraTreasuryTomb is a smaller rock-cut tomb with a crow-step crown, cut into the
// plaza cliffs. Local +Z faces out of the cliff; the façade's back plane is local z = -2.5.
type petraTreasuryTomb struct {
	pos     Vec2
	heading float64
	w, h    float64
}

var petraTreasuryTombs = []petraTreasuryTomb{
	{P(-17.2, 1.5), -math.Pi / 2, 6.4, 13},
	{P(-18.8, 9.5), -math.Pi / 2, 4.6, 9},
	{P(17.8, 9.5), math.Pi / 2, 5.4, 11},
	{P(27.2, 4), -math.Pi / 2, 5.6, 12},
	{P(-27.2, -12), math.Pi / 2, 5.6, 12},
	{P(6, -28.2), math.Pi, 6, 13},
}

func (t petraTreasuryTomb) local(sol Solid) Solid {
	return RotateY(Translate(sol, V(t.pos.X, 0, t.pos.Y)), t.pos, t.heading)
}

// tomb fills a tomb façade: wall, pilasters, cornice, crow steps and a dark door.
func (PetraTreasury) tomb(g *Grid, t petraTreasuryTomb) {
	w2, f := t.w/2, petraTreasuryFloor
	g.Paint(Charcoal, t.local(Block(V(-3, f+0.2, -7.8), V(3, 7.8, -2.2))))
	g.With(Terracotta, func() {
		g.Fill(t.local(Block(V(-w2, f, -2.5), V(w2, t.h-2, -1.4))))
		g.Fill(t.local(Block(V(-w2-0.3, f, -2.5), V(w2+0.3, f+0.6, -0.9))))
		for _, x := range []float64{-w2 + 0.5, w2 - 0.5} {
			g.Fill(t.local(Block(V(x-0.5, f, -2.5), V(x+0.5, t.h-2.6, -0.8))))
		}
		g.Fill(t.local(Block(V(-w2-0.3, t.h-2.6, -2.5), V(w2+0.3, t.h-1.6, -0.6))))
		// Crow steps: two staggered rows of stepped merlons.
		n := int(t.w / 1.6)
		for i := 0; i < n; i++ {
			x := -w2 + 0.8 + float64(i)*(t.w-1.6)/math.Max(1, float64(n-1))
			g.Fill(t.local(Block(V(x-0.55, t.h-1.6, -2.5), V(x+0.55, t.h-0.4, -1.4))))
			g.Fill(t.local(Block(V(x-0.25, t.h-0.4, -2.5), V(x+0.25, t.h+0.6, -1.4))))
		}
	})
	g.With(Sand, func() {
		g.Fill(t.local(Block(V(-w2-0.3, t.h-2.2, -1.3), V(w2+0.3, t.h-1.6, -0.6))))
		g.Fill(t.local(Block(V(-1.6, 5.6, -1.6), V(1.6, 6.3, -1))))
	})
	g.Carve(t.local(Block(V(-1, f+0.6, -3), V(1, 5.6, -0.5))))
}
