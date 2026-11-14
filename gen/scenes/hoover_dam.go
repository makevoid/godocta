package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// HooverDam is the arch-gravity dam in Black Canyon, cut out of the desert as a block of
// dark, banded volcanic rock. The pale concrete wedge bows upstream into Lake Mead and
// flares toward its base; the crest carries the old highway between parapets, with the
// two elevator towers, two utility towers, cars and a tour bus. Behind it the lake
// stands high against a white bathtub ring, with the four fluted intake towers (two per
// side) joined to each other and to the canyon walls by footbridges, a tour boat, and
// the Nevada side spillway trough with its drum-gate weir. At the toe the U-shaped
// powerhouse sends two wings down the canyon walls around the tailrace; further down,
// valve houses on both walls throw white jets into the green Colorado, transmission
// pylons lean over the rim, and the concrete arch of the bypass bridge spans the canyon
// high above the river.
//
// Everything is authored in canyon coordinates (u across, toward the Arizona wall;
// t downstream) and turned by hooverDamYaw, so the downstream face looks toward
// +Z/+X: the hero camera looks up the canyon at the dam, the powerhouse and the river,
// with the lake behind. Rock, dam and water are one hollow skin over three height
// fields.
type HooverDam struct{}

func init() { Register(HooverDam{}) }

func (HooverDam) Name() string  { return "hoover_dam" }
func (HooverDam) Title() string { return "Hoover Dam" }

const (
	hooverDamYaw     = 0.44 // Downstream points at (sin, cos) of this angle in world X/Z.
	hooverDamCrest   = 56.0 // Roadway level on the dam.
	hooverDamLake    = 49.0 // Lake Mead surface.
	hooverDamRing    = 54.2 // Top of the bathtub ring.
	hooverDamRiver   = 2.0  // Colorado surface below the dam.
	hooverDamArcR    = 46.0 // Radius of the upstream face.
	hooverDamArcT    = 24.0 // Arc center, downstream of the crown.
	hooverDamRoad    = 5.5  // Crest width.
	hooverDamToe     = 30.0 // Extra thickness at the base.
	hooverDamFloor   = 13.0 // Canyon floor half width.
	hooverDamRun     = 23.0 // Plan run of a canyon wall from floor to rim.
	hooverDamHead    = -42.0
	hooverDamSpill0  = -44.0 // Spillway trough, upstream and downstream ends.
	hooverDamSpill1  = -26.0
	hooverDamBridgeT = 50.0 // Bypass bridge center line.
	hooverDamBridgeY = 61.5 // Rim and road level at the bridge.
	hooverDamIntakeU = 23.0 // Intake towers' distance from the canyon axis.
)

// Terrain zones.
const (
	hooverDamFloorZone = iota
	hooverDamWallZone
	hooverDamTopZone
	hooverDamSpillZone
	hooverDamCutZone
)

// hooverDamHills are low summits on the rim: u, t, height, radius.
var hooverDamHills = [][4]float64{
	{-43, 37, 5, 8}, {-30, -49, 4, 9}, {22, -49, 5, 10}, {41, -27, 4, 8}, {-43, -8, 3.5, 6},
}

// hooverDamStrata is the rock's banding from the river up, repeated.
var hooverDamStrata = []Color{
	Charcoal, DarkStone, Brick, DarkStone, DarkStone, Charcoal, Brick, Terracotta, DarkStone, Charcoal,
	DarkStone, Brick, Brick, DarkStone, Charcoal, DarkStone, Terracotta, Brick, DarkStone, Charcoal,
	DarkStone, Brick, DarkStone, Terracotta, Charcoal, Brick, DarkStone, DarkStone, Charcoal, Brick,
}

// hooverDamColumn is one lattice column of the massing.
type hooverDamColumn struct {
	u, t             float64
	rock, dam, water float64
	inner            float64 // Rock surface before the outer cut face trims it.
	top              float64
	zone             int
	lake             bool // Upstream of the dam.
}

func hooverDamSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

// hooverDamNoise is smooth value noise in [-1, 1].
func hooverDamNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	ab, cd := a+(b-a)*sx, c+(d-c)*sx
	return (ab+(cd-ab)*sz)*2 - 1
}

func hooverDamFBM(x, z float64, seed int) float64 {
	return 0.6*hooverDamNoise(x, z, seed) + 0.3*hooverDamNoise(x*2.03+17, z*2.03+5, seed+1) +
		0.1*hooverDamNoise(x*4.1+3, z*4.1+29, seed+2)
}

// hooverDamLocal maps a world plan point to canyon coordinates.
func hooverDamLocal(x, z float64) (u, t float64) {
	q := P(x, z).Rotate(hooverDamYaw)
	return q.X, q.Y
}

// hooverDamPlace turns a solid authored in canyon coordinates (X = u, Z = t) to the world.
func hooverDamPlace(s Solid) Solid { return RotateY(s, P(0, 0), -hooverDamYaw) }

func (HooverDam) fill(g *Grid, c Color, solids ...Solid) {
	g.Use(c)
	for _, s := range solids {
		g.Fill(hooverDamPlace(s))
	}
}

func (HooverDam) vol(g *Grid, c Color, solids ...Solid) {
	g.Use(c)
	for _, s := range solids {
		g.FillVolume(hooverDamPlace(s))
	}
}

func (HooverDam) paint(g *Grid, c Color, solids ...Solid) {
	for _, s := range solids {
		g.Paint(c, hooverDamPlace(s))
	}
}

// box is a canyon-coordinate block from (u0, y0, t0) to (u1, y1, t1).
func (HooverDam) box(u0, y0, t0, u1, y1, t1 float64) Solid {
	return Block(V(u0, y0, t0), V(u1, y1, t1))
}

// arc swings a solid authored at the crown of the crest (X along the road, Z = t)
// round the dam's axis; positive angles go toward the Arizona abutment.
func (HooverDam) arc(s Solid, theta float64) Solid { return RotateY(s, P(0, hooverDamArcT), theta) }

// ring is a sector of an annulus about the dam's axis, on its upstream side.
func (HooverDam) ring(r0, r1, y0, y1, span float64) Solid {
	return Func{Box{V(-r1, y0, hooverDamArcT-r1), V(r1, y1, hooverDamArcT+r1)}, func(p Vec3) bool {
		du, dt := p.X, hooverDamArcT-p.Z
		r := math.Hypot(du, dt)
		return r >= r0 && r <= r1 && math.Abs(math.Atan2(du, dt)) <= span
	}}
}

// proud keeps the part of a canyon-coordinate solid that stands clear of the massing,
// sunk into it by embed, so nothing is spent inside the rock or the concrete.
func (s HooverDam) proud(sol Solid, embed float64) Solid {
	return Func{sol.Bounds(), func(p Vec3) bool {
		return sol.Contains(p) && p.Y > s.column(p.X, p.Z).top-embed
	}}
}

// ---------------------------------------------------------------------------------
// Height fields.

// edge is the inward distance from the block's rounded outline, wobbled into buttresses.
func (HooverDam) edge(u, t float64) float64 {
	const cu, ct, a, b, r = -1.5, 1.25, 45.5, 54.25, 13.0
	qx, qy := math.Abs(u-cu)-(a-r), math.Abs(t-ct)-(b-r)
	e := r - (math.Hypot(math.Max(qx, 0), math.Max(qy, 0)) + math.Min(math.Max(qx, qy), 0))
	return e - 0.7 - 0.7*hooverDamFBM(u*0.13, t*0.13, 23)
}

// canyon returns the plan distance d into the canyon wall (negative on the floor), the
// wall's plan run up to the rim, and the rim height. Upstream of the dam the floor widens
// into the lake basin and closes in a rounded head.
func (HooverDam) canyon(u, t float64) (d, run, rim float64) {
	up := hooverDamSmooth((-6 - t) / 22)
	floor := hooverDamFloor + 16*up
	run = hooverDamRun - 16*up
	rim = hooverDamCrest + (hooverDamBridgeY-hooverDamCrest)*hooverDamSmooth((t-12)/26)
	const r = 14.0
	a := math.Abs(u) - (floor - r)
	b := (hooverDamHead + r) - t
	d = math.Hypot(math.Max(a, 0), math.Max(b, 0)) + math.Min(math.Max(a, b), 0) - r
	side := 7.3
	if u < 0 {
		side = 1.9
	}
	n := 1.5*hooverDamFBM(t*0.1, u*0.1, 3) + 1.0*hooverDamFBM(t*0.3, side, 5)
	d += n * hooverDamSmooth(d/5)
	return
}

// roadOff is the distance outside the crest highway's bench, which runs on round the
// rim beyond both abutments.
func (HooverDam) roadOff(u, t float64) float64 {
	if t > 21 {
		return 100
	}
	r := math.Hypot(u, t-hooverDamArcT)
	return math.Max(hooverDamArcR-hooverDamRoad-r, r-hooverDamArcR)
}

// site samples the rock: its surface, the surface before the outer cut, and the zone.
func (s HooverDam) site(u, t float64) (h, inner float64, zone int) {
	e := s.edge(u, t)
	if e <= 0 {
		return -100, -100, hooverDamCutZone
	}
	d, run, rim := s.canyon(u, t)
	h, zone = 0.4, hooverDamFloorZone
	switch {
	case d <= 0:
	case d < run:
		x := d / run
		h = rim * math.Pow(x, 0.77)
		// Ledges: the wall climbs in benches.
		const step = 7.0
		q := h / step
		f := q - math.Floor(q)
		st := (math.Floor(q) + hooverDamSmooth((f-0.2)/0.6)) * step
		h += (st - h) * 0.5 * (1 - hooverDamSmooth((x-0.8)/0.2))
		zone = hooverDamWallZone
	default:
		p := d - run
		k := hooverDamSmooth(p / 3)
		h = rim + 1.2*k + 1.3*k*hooverDamFBM(u*0.09, t*0.09, 11)
		for _, m := range hooverDamHills {
			h += m[2] * k * hooverDamSmooth(1-math.Hypot(u-m[0], t-m[1])/m[3])
		}
		// Road benches for the crest highway and the bypass.
		if off := s.roadOff(u, t); off < 4 {
			h = math.Min(h, hooverDamCrest+1.2*math.Max(off, 0))
		}
		if off := math.Abs(t-hooverDamBridgeT) - 2.6; off < 4 {
			h = math.Min(h, hooverDamBridgeY+1.2*math.Max(off, 0))
		}
		zone = hooverDamTopZone
	}
	// Spillway: a lake bay, the weir and the concrete trough behind it.
	if t >= hooverDamSpill0 && t <= hooverDamSpill1 {
		switch {
		case u > -34.9 && u < -28:
			h, zone = math.Min(h, 40), hooverDamFloorZone
		case u >= -36.7 && u <= -34.9:
			h, zone = 51.4, hooverDamSpillZone
		case u >= -42 && u < -36.7:
			h, zone = 47.5-4*(t-hooverDamSpill0)/(hooverDamSpill1-hooverDamSpill0), hooverDamSpillZone
		}
	}
	inner = h
	if cut := 20 * e; cut < h {
		h, zone = math.Max(cut, 0.3), hooverDamCutZone
	}
	return
}

// upstream is the t of the dam's upstream face at u.
func (HooverDam) upstream(u float64) float64 {
	a := math.Min(math.Abs(u), 40)
	return hooverDamArcT - math.Sqrt(hooverDamArcR*hooverDamArcR-a*a)
}

// face is the t of the downstream face at u, at a depth fraction below the crest. Each
// level is an arc: both its radius and its crown move as the dam thickens downward.
func (HooverDam) face(u, depth float64) float64 {
	r := hooverDamArcR - hooverDamRoad - 14*depth
	crown := hooverDamArcT - hooverDamArcR + hooverDamRoad + hooverDamToe*math.Pow(depth, 1.5)
	a := math.Min(math.Abs(u), 0.985*r)
	return crown + r - math.Sqrt(r*r-a*a)
}

// damTop is the concrete's height over a canyon point, or -100 where there is no dam.
func (s HooverDam) damTop(u, t float64) float64 {
	if math.Abs(u) > 40 || t < s.upstream(u) || t > s.face(u, 1) {
		return -100
	}
	if t <= s.face(u, 0) {
		return hooverDamCrest
	}
	lo, hi := 0.0, 1.0
	for i := 0; i < 24; i++ {
		mid := (lo + hi) / 2
		if s.face(u, mid) < t {
			lo = mid
		} else {
			hi = mid
		}
	}
	return hooverDamCrest * (1 - hi)
}

func (s HooverDam) column(u, t float64) hooverDamColumn {
	c := hooverDamColumn{u: u, t: t, dam: -100, water: -100}
	c.rock, c.inner, c.zone = s.site(u, t)
	c.top = c.rock
	if c.rock <= -100 {
		return c
	}
	c.lake = t < s.upstream(u)
	c.dam = s.damTop(u, t)
	switch {
	case c.zone == hooverDamSpillZone:
	case c.lake && c.inner < hooverDamLake:
		c.water = hooverDamLake
	case !c.lake && c.inner < hooverDamRiver:
		c.water = hooverDamRiver
	}
	c.top = math.Max(c.rock, math.Max(c.dam, c.water))
	return c
}

// massing builds rock, concrete and water as one skin: only cells with an empty face
// neighbor are kept (space below y = 0 counts as filled), so the canyon costs its
// surface, not its volume, and nothing is spent on the lake bed or the dam's core.
func (s HooverDam) massing(g *Grid) {
	step := g.Span() / 2
	b := hooverDamPlace(Block(V(-49, 0, -56), V(46, 1, 58))).Bounds()
	x0, x1 := int(math.Floor(b.Min.X/step))-2, int(math.Ceil(b.Max.X/step))+2
	z0, z1 := int(math.Floor(b.Min.Z/step))-2, int(math.Ceil(b.Max.Z/step))+2
	w := x1 - x0 + 1
	cols := make([]hooverDamColumn, w*(z1-z0+1))
	for iz := z0; iz <= z1; iz++ {
		for ix := x0; ix <= x1; ix++ {
			u, t := hooverDamLocal(float64(ix)*step, float64(iz)*step)
			cols[(iz-z0)*w+ix-x0] = s.column(u, t)
		}
	}
	occ := func(ix, iy, iz int) bool {
		if iy < 0 {
			return true
		}
		if ix < x0 || ix > x1 || iz < z0 || iz > z1 {
			return false
		}
		return float64(iy)*step <= cols[(iz-z0)*w+ix-x0].top
	}
	for iz := z0; iz <= z1; iz++ {
		for ix := x0; ix <= x1; ix++ {
			if (ix^iz)&1 != 0 {
				continue
			}
			c := &cols[(iz-z0)*w+ix-x0]
			if c.top < 0 {
				continue
			}
			for iy := ix & 1; float64(iy)*step <= c.top; iy += 2 {
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
				switch {
				case y <= c.rock:
					g.Use(s.rockColor(c, y))
				case y <= c.dam:
					g.Use(s.damColor(c.u, c.t, y))
				default:
					g.Use(s.waterColor(c))
				}
				g.Add(Cell{int32(ix), int32(iy), int32(iz)})
			}
		}
	}
}

// rockColor bands the canyon in warped strata and bleaches the bathtub ring.
func (s HooverDam) rockColor(c *hooverDamColumn, y float64) Color {
	u, t := c.u, c.t
	if c.zone == hooverDamSpillZone {
		if c.rock-y < 0.6 && c.rock < 50 {
			return Stone
		}
		return White
	}
	if c.lake && c.zone != hooverDamCutZone && y > hooverDamLake-0.8 &&
		y < hooverDamRing+0.5*hooverDamNoise(u*0.3, t*0.3, 91) {
		if hooverDamFBM(u*0.5, t*0.5, 93) > 0.3 {
			return Stone
		}
		return White
	}
	if c.zone == hooverDamTopZone && c.rock-y < 0.9 {
		switch n := hooverDamFBM(u*0.12, t*0.12, 81); {
		case n > 0.45:
			return Sand
		case n > 0.2:
			return Terracotta
		case n < -0.25:
			return DarkStone
		}
		return Brick
	}
	w := y + 3.5*hooverDamFBM(u*0.035, t*0.035, 31) + 0.7*hooverDamFBM(u*0.2, t*0.2, 33)
	n := len(hooverDamStrata)
	col := hooverDamStrata[((int(math.Floor(w/2.3))%n)+n)%n]
	// Desert varnish and pale scars in coherent patches rather than speckle.
	switch p := hooverDamNoise(u*0.45+y*0.13, t*0.45-y*0.11, 75); {
	case p > 0.62:
		col = Charcoal
	case p < -0.62:
		col = DarkStone
	}
	return col
}

// damColor is pale concrete with lift lines, block joints and faint weathering.
func (s HooverDam) damColor(u, t, y float64) Color {
	arc := math.Atan2(u, hooverDamArcT-t) * 40
	if math.Mod(y+0.2, 8) < 0.42 || math.Abs(math.Mod(arc+90, 9)-4.5) < 0.22 {
		return Stone
	}
	if y < hooverDamCrest-6 && hooverDamFBM(arc*0.3, y*0.03, 51) > 0.38 {
		return Stone
	}
	return White
}

// waterColor: deep blue lake, green river.
func (s HooverDam) waterColor(c *hooverDamColumn) Color {
	if c.lake {
		d, _, _ := s.canyon(c.u, c.t)
		if d < -3 && hooverDamFBM(c.u*0.06, c.t*0.06, 41) > -0.2 {
			return Navy
		}
		return Water
	}
	if hooverDamFBM(c.u*0.3, c.t*0.06, 43) > 0.22 {
		return Water
	}
	return Foliage
}

func (s HooverDam) Build(g *Grid) Order {
	s.massing(g)
	s.crest(g)
	s.intakes(g)
	s.spillway(g)
	s.powerhouse(g)
	s.outlets(g)
	s.bridge(g)
	s.pylons(g)
	s.visitor(g)
	s.boat(g)
	return Along(P(math.Sin(hooverDamYaw), math.Cos(hooverDamYaw)))
}

// ---------------------------------------------------------------------------------
// Crest.

// crest lays the highway between parapets and stands the towers and traffic on it.
func (s HooverDam) crest(g *Grid) {
	const in, out, top = hooverDamArcR - hooverDamRoad, hooverDamArcR, hooverDamCrest
	onDam := func(sol Solid) Solid {
		return Func{sol.Bounds(), func(p Vec3) bool {
			if !sol.Contains(p) {
				return false
			}
			c := s.column(p.X, p.Z)
			return c.dam >= top-0.01 && c.rock < top-0.3
		}}
	}
	s.paint(g, Charcoal, s.ring(in+0.9, out-0.9, top-0.9, top+0.3, 1.45))
	s.fill(g, White, onDam(s.ring(in, in+0.9, top, top+1.2, 1.3)), onDam(s.ring(out-0.9, out, top, top+1.2, 1.3)))

	// Elevator towers near the crown and utility towers further out, all standing
	// proud of the downstream face.
	tower := func(theta, half, r0, r1, y1 float64) {
		z0, z1 := hooverDamArcT-r1, hooverDamArcT-r0
		s.fill(g, White, s.proud(s.arc(s.box(-half, 40, z0, half, y1, z1), theta), 0.6))
		s.fill(g, Stone, s.arc(s.box(-half-0.4, y1, z0-0.4, half+0.4, y1+0.6, z1+0.4), theta))
		s.fill(g, White, s.arc(s.box(-half+0.6, y1+0.6, z0+0.5, half-0.6, y1+1.8, z1-0.5), theta))
		s.paint(g, Charcoal, s.arc(s.box(-0.45, 49, z1-0.3, 0.45, y1-1.2, z1+0.3), theta),
			s.arc(s.box(-0.45, top+1.6, z0-0.3, 0.45, y1-1.2, z0+0.3), theta))
	}
	for _, sgn := range []float64{-1, 1} {
		tower(sgn*0.2, 1.5, in-1.6, in+1.0, 63.5)
		tower(sgn*0.62, 1.3, in-1.2, in+1.0, 61)
	}

	// Traffic: cars in both lanes and a yellow tour bus.
	lanes := [2]float64{hooverDamArcT - (in + 1.85), hooverDamArcT - (out - 1.85)}
	colors := []Color{Red, White, Water, Gold, Sand, Red, Terracotta, White, Red, Gold}
	for i, theta := range []float64{-0.92, -0.7, -0.5, -0.33, 0.12, 0.36, 0.52, 0.8, 0.98, 1.16} {
		z := lanes[i%2]
		s.fill(g, colors[i], s.arc(s.box(-1.3, top+0.15, z-0.65, 1.3, top+1.1, z+0.65), theta))
		s.fill(g, Charcoal, s.arc(s.box(-0.6, top+1.1, z-0.55, 0.7, top+1.6, z+0.55), theta))
	}
	z := lanes[1]
	s.fill(g, Gold, s.arc(s.box(-2.8, top+0.15, z-0.75, 2.8, top+2.1, z+0.75), -0.1))
	s.paint(g, Charcoal, s.arc(s.box(-2.5, top+1.1, z-0.9, 2.5, top+1.6, z+0.9), -0.1))
}

// ---------------------------------------------------------------------------------
// Lake side.

// intake is one fluted intake tower with its gallery, lantern and copper-dark cap.
func (s HooverDam) intake(g *Grid, c Vec2) {
	const r = 2.7
	lo := hooverDamLake - 1
	box := Box{c.At(lo).Sub(V(r+1, 0, r+1)), c.At(68).Add(V(r+1, 0, r+1))}
	sector := func(p Vec3, n int) int {
		a := p.Plan().Sub(c).Angle() + math.Pi
		return int(math.Floor(a / (2 * math.Pi) * float64(n)))
	}
	s.vol(g, White, Cylinder{c, r, 1.5, lo, 60.6})
	// Trash-rack slots between the concrete fins.
	s.paint(g, Charcoal, Func{box, func(p Vec3) bool {
		return p.Y > hooverDamLake+0.4 && p.Y < 58.6 && sector(p, 12)%2 == 0
	}})
	s.vol(g, White, Cylinder{c, r + 0.7, 0, 60.6, 61.6})
	s.vol(g, Stone, Cylinder{c, 2.1, 0, 61.6, 64.4})
	s.paint(g, Glass, Func{box, func(p Vec3) bool { return p.Y > 62.2 && p.Y < 63.9 && sector(p, 8)%2 == 0 }})
	s.vol(g, Slate, Frustum{c, 2.7, 0.9, 64.4, 66.2})
	s.fill(g, Gold, Capsule{c.At(66.2), c.At(67.4), 0.5})
}

// intakes stands two towers off each wall of the lake and bridges them to the rim and
// to each other.
func (s HooverDam) intakes(g *Grid) {
	const t0, t1 = -25.5, -35.5
	for _, sgn := range []float64{-1, 1} {
		u := sgn * hooverDamIntakeU
		for _, t := range []float64{t0, t1} {
			s.fill(g, White, s.proud(s.box(u+sgn*2.4, 56.2, t-0.8, u+sgn*15, 57.4, t+0.8), 0.4))
			s.fill(g, Stone, s.proud(s.box(u+sgn*6.6, 44, t-0.6, u+sgn*7.8, 56.2, t+0.6), 0.5))
		}
		s.fill(g, White, s.box(u-0.8, 56.2, t1+2.4, u+0.8, 57.4, t0-2.4))
		s.intake(g, P(u, t0))
		s.intake(g, P(u, t1))
	}
}

// spillway dresses the Nevada weir with drum gates between piers and opens the tunnel
// mouth at the low end of the trough.
func (s HooverDam) spillway(g *Grid) {
	s.paint(g, Slate, s.box(-36.7, 50.8, hooverDamSpill0, -34.9, 51.7, hooverDamSpill1))
	for i := 0; i <= 4; i++ {
		t := hooverDamSpill0 + 0.6 + float64(i)*(hooverDamSpill1-hooverDamSpill0-1.2)/4
		s.fill(g, White, s.box(-37.3, 51.4, t-0.6, -34.3, 54.6, t+0.6))
	}
	s.paint(g, Charcoal, s.box(-41.4, 43.2, hooverDamSpill1-0.2, -37.3, 48.2, hooverDamSpill1+1.2))
}

// boat is a tour boat on the lake, trailing a wake.
func (s HooverDam) boat(g *Grid) {
	y := hooverDamLake + 0.45
	s.fill(g, White, Capsule{V(-8, y, -34), V(-3.6, y, -32.2), 0.8})
	s.fill(g, Red, Capsule{V(-6.4, y+0.9, -33.35), V(-5.2, y+0.9, -32.85), 0.5})
	for _, side := range []float64{-1, 1} {
		s.paint(g, White, Capsule{V(-9, y-0.5, -34.4+side*0.6), V(-14, y-0.5, -36.5+side*2.6), 0.55})
	}
}

// ---------------------------------------------------------------------------------
// Canyon floor.

// powerhouse is the U at the toe: a hall across the canyon whose back is buried in the
// dam, and two wings along the walls with transformer decks over the tailrace.
func (s HooverDam) powerhouse(g *Grid) {
	const bar, t1, roof, high = 21.0, 38.0, 12.5, 15.0
	hall := Subtract(s.box(-14, 0.5, 1, 14, high, bar), s.box(-13, 0.5, 1, 13, high-1, bar-1),
		s.box(-13, 0.5, bar-1, -8.5, roof-1, bar), s.box(8.5, 0.5, bar-1, 13, roof-1, bar))
	s.fill(g, White, s.proud(hall, 0.8))
	s.fill(g, Stone, s.proud(Intersect(hall, s.box(-14, 0.5, 1, 14, 4.4, bar)), 0.8))
	s.paint(g, Slate, s.box(-13, high-0.5, 3.5, 13, high+0.3, bar-1))
	for _, u := range []float64{-5.2, -2.6, 0, 2.6, 5.2} {
		s.paint(g, Glass, s.box(u-0.45, 5.6, bar-0.4, u+0.45, 13.2, bar+0.4))
	}
	// Stepped penthouse over the middle of the hall.
	s.fill(g, White, s.box(-3, high, 14, 3, high+2, 19))
	s.paint(g, Slate, s.box(-2.4, high+1.6, 14.6, 2.4, high+2.3, 18.4))

	for _, sgn := range []float64{-1, 1} {
		in, out := sgn*7.5, sgn*14.0
		wing := Subtract(s.box(in, 0.5, bar, out, roof, t1), s.box(in+sgn, 0.5, bar, out, roof-1, t1-1))
		s.fill(g, White, s.proud(wing, 0.8))
		s.fill(g, Stone, s.proud(Intersect(wing, s.box(in, 0.5, bar, out, 4.4, t1)), 0.8))
		s.paint(g, Slate, s.box(in+sgn*0.9, roof-0.5, bar, sgn*13, roof+0.3, t1-0.9))
		// Tall window strips between piers, on the tailrace side and the end wall.
		for t := bar + 1.5; t < t1-1.2; t += 2.6 {
			s.paint(g, Charcoal, s.box(in-sgn*0.4, 5.6, t, in+sgn*0.4, 10.8, t+0.9))
		}
		for _, du := range []float64{1.8, 4.0} {
			s.paint(g, Charcoal, s.box(in+sgn*du, 5.6, t1-0.4, in+sgn*(du+0.9), 10.8, t1+0.4))
		}
		for t := bar + 3; t < t1-3; t += 5 {
			s.fill(g, White, s.box(sgn*9.6, roof, t, sgn*11.4, roof+1, t+1.8))
		}
		// Transformer deck over the draft-tube outlets.
		dk := sgn * 5.2
		s.fill(g, Stone, s.box(dk, 3.4, bar, in, 4.4, t1-0.8), s.box(dk, 0.8, bar, dk+sgn*0.8, 3.4, t1-0.8))
		for t := bar + 1.2; t < t1-2.4; t += 3.2 {
			s.fill(g, Slate, s.box(dk+sgn*0.5, 4.4, t, dk+sgn*1.7, 5.6, t+1.4))
			s.fill(g, White, s.box(dk+sgn*0.8, 5.6, t+0.4, dk+sgn*1.4, 6.1, t+1.0))
			s.paint(g, Charcoal, s.box(dk-sgn*0.3, 2.0, t+1.9, dk+sgn*0.3, 3.2, t+2.9))
		}
		foam := s.box(dk-sgn*1.7, 1.4, bar, dk, 2.3, t1+2).Bounds()
		s.paint(g, White, Func{foam, func(p Vec3) bool { return hooverDamNoise(p.Z*0.8, sgn*3, 77) > 0.15 }})
	}
}

// wallAt finds where a canyon wall reaches height y at station t, searching outward.
func (s HooverDam) wallAt(sgn, t, y float64) float64 {
	u := hooverDamFloor
	for ; u < 40; u += 0.25 {
		if h, _, _ := s.site(sgn*u, t); h >= y {
			break
		}
	}
	return u
}

// outlets are the canyon-wall valve houses, each throwing three jets into the river.
func (s HooverDam) outlets(g *Grid) {
	const t0, t1 = 40.5, 46.5
	for _, sgn := range []float64{-1, 1} {
		w := s.wallAt(sgn, (t0+t1)/2, 17)
		in, out := sgn*(w-2.6), sgn*(w+3.4)
		s.fill(g, Stone, s.proud(s.box(in, 9, t0, out, 16.6, t1), 0.8))
		s.fill(g, White, s.proud(s.box(in, 16.6, t0, out, 23.5, t1), 0.8))
		s.paint(g, Slate, s.box(in+sgn*0.6, 23, t0+0.6, in+sgn*3.6, 23.8, t1-0.6))
		for _, t := range []float64{42.55, 44.45} {
			s.paint(g, Glass, s.box(in-sgn*0.3, 20.2, t-0.5, in+sgn*0.3, 22, t+0.5))
		}
		for _, t := range []float64{41.6, 43.5, 45.4} {
			s.paint(g, Charcoal, s.box(in-sgn*0.3, 17, t-0.6, in+sgn*0.3, 18.6, t+0.6))
			a, b := V(in-sgn*0.3, 17.8, t), V(sgn*3.2, hooverDamRiver, t+2.5)
			s.paint(g, Glass, Cylinder{b.Plan(), 3, 0, 1.2, 2.4})
			s.paint(g, White, Cylinder{b.Plan(), 1.9, 0, 1.2, 2.4})
			prev := a
			for k := 1; k <= 6; k++ {
				f := float64(k) / 6
				q := V(a.X+(b.X-a.X)*f, a.Y+(b.Y-a.Y)*f*f, a.Z+(b.Z-a.Z)*f)
				s.fill(g, White, Capsule{prev, q, 0.55})
				prev = q
			}
		}
	}
}

// ---------------------------------------------------------------------------------
// Rim.

// bridge is the bypass: a deck at rim level on spandrel columns over twin arch ribs.
func (s HooverDam) bridge(g *Grid) {
	const T, Y = hooverDamBridgeT, hooverDamBridgeY
	crown := func(u float64) float64 { return 58.4 - 15*(u/28)*(u/28) }
	ribs := Func{Box{V(-29, 36, T-2.1), V(29, 60, T+2.1)}, func(p Vec3) bool {
		w := math.Abs(p.Z - T)
		if w < 0.7 || w > 2.1 {
			return false
		}
		slope := 30 * p.X / (28 * 28)
		return math.Abs(p.Y-crown(p.X)) <= 0.8*math.Sqrt(1+slope*slope)
	}}
	s.fill(g, Stone, s.proud(ribs, 0.6))
	for _, u := range []float64{-34, -28, -21, -14, -7, 7, 14, 21, 28, 34} {
		y0 := 30.0
		if math.Abs(u) < 27 {
			y0 = crown(u)
			s.fill(g, Stone, s.box(u-0.5, y0-0.6, T-0.7, u+0.5, y0+0.6, T+0.7))
		}
		for _, side := range []float64{-1, 1} {
			s.fill(g, Stone, s.proud(s.box(u-0.6, y0, T+side*0.9, u+0.6, Y-1, T+side*2.1), 0.6))
		}
	}
	s.fill(g, Stone, s.proud(s.box(-41, Y-1, T-2.4, 41, Y, T+2.4), 0.3))
	s.paint(g, Charcoal, s.box(-47, Y-0.6, T-1.6, 45, Y+0.2, T+1.6))
	s.fill(g, White, s.box(-38, Y, T-2.4, 38, Y+1.1, T-1.6), s.box(-38, Y, T+1.6, 38, Y+1.1, T+2.4))
}

// pylon is a transmission tower at the canyon lip, leaning out over the powerhouse.
func (s HooverDam) pylon(g *Grid, sgn, t float64) {
	u := sgn * (s.wallAt(sgn, t, hooverDamCrest-1) + 2.2)
	base, _, _ := s.site(u, t)
	at := func(du, y, dt float64) Vec3 { return V(u+du-sgn*0.28*y, base+y, t+dt) }
	for _, a := range []float64{-1, 1} {
		for _, b := range []float64{-1, 1} {
			s.fill(g, Slate, Capsule{at(a*1.3, -1.5, b*1.3), at(0, 8.5, 0), 0.5})
		}
	}
	s.fill(g, Slate, Capsule{at(0, 8.5, 0), at(0, 12, 0), 0.5},
		Capsule{at(0, 8.2, -3.2), at(0, 8.2, 3.2), 0.5}, Capsule{at(0, 11.2, -2.2), at(0, 11.2, 2.2), 0.5})
	for _, dt := range []float64{-3.2, 3.2} {
		s.fill(g, White, Capsule{at(0, 7.1, dt), at(0, 7.5, dt), 0.45})
	}
}

func (s HooverDam) pylons(g *Grid) {
	for _, v := range [][2]float64{{-1, 33}, {-1, 42}, {1, 24}, {1, 33}, {1, 42}} {
		s.pylon(g, v[0], v[1])
	}
}

// visitor is the round visitor center on the Nevada rim, where the old highway ends.
func (s HooverDam) visitor(g *Grid) {
	c := P(-40.5, 25.5)
	base, _, _ := s.site(c.X, c.Y)
	s.fill(g, Sand, s.proud(Cylinder{c, 3, 2, base - 3, base + 3.8}, 0.6))
	s.paint(g, Glass, Cylinder{c, 3.3, 2.1, base + 1.6, base + 2.9})
	s.vol(g, Terracotta, Cylinder{c, 3.5, 0, base + 3.8, base + 4.6}, Frustum{c, 2.4, 1, base + 4.6, base + 5.6})
}
