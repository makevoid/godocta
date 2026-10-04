package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// Lalibela is the rock-hewn church of Bete Giyorgis (St George) at Lalibela, Ethiopia:
// a monolithic Greek-cross church carved downward out of the reddish volcanic tuff,
// standing on a stepped plinth in a deep square pit. Its flat roof carries three nested
// cross reliefs; the walls have a row of blind lower windows, a string course, ornate
// upper windows with dark interiors and three doors. A trench from the plateau ends in a
// tunnel into the pit, and hermit cells are cut into the sheer pit walls. To the west,
// a tunnel leads to a second excavated church in its own courtyard, a rectangular hall
// ringed by square pillars under a low gabled roof (after Bete Medhane Alem). The pits
// are carved into a gentle rise that falls toward the front (+Z) and slopes down to a
// low rim; the plateau has dry grass, eucalyptus and acacia trees, round thatched
// tukuls and white-robed pilgrims.
type Lalibela struct{}

func init() { Register(Lalibela{}) }

func (Lalibela) Name() string  { return "lalibela" }
func (Lalibela) Title() string { return "Lalibela" }

const (
	lalibelaX0, lalibelaX1 = -58.0, 46.0
	lalibelaZ0, lalibelaZ1 = -36.0, 38.0
	lalibelaRim            = 1.0
	lalibelaFloor          = 2.0
	lalibelaSkin           = 1.0
	lalibelaHalf           = 7.0  // Half width of a church cross arm.
	lalibelaReach          = 19.0 // Arm length from the church center.
)

var (
	lalibelaG = P(12, 0)   // Bete Giyorgis.
	lalibelaM = P(-33, -8) // The pillared church.
	// Pits as plan rectangles: min and max corners.
	lalibelaPitG = [2]Vec2{P(-13, -25), P(37, 25)}
	lalibelaPitM = [2]Vec2{P(-47, -20), P(-19, 13)}
	// The trench path, from the tunnel mouth in the pit wall out to the plateau.
	lalibelaPath = []Vec2{P(29, 25), P(29, 30.5), P(-4, 30.5)}
)

// lalibelaCave is a carved opening with a darkened interior.
type lalibelaCave struct {
	carve, dark Solid
}

func (Lalibela) smooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func (Lalibela) noise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

func (s Lalibela) fbm(x, z float64, seed int) float64 {
	return (s.noise(x, z, seed)*4 + s.noise(x*2.1+5, z*2.1+3, seed+1)*2 + s.noise(x*4.3+1, z*4.3+7, seed+2)) / 7
}

func (Lalibela) inRect(r [2]Vec2, p Vec2) bool {
	return p.X >= r[0].X && p.X <= r[1].X && p.Y >= r[0].Y && p.Y <= r[1].Y
}

// path returns the distance to the trench path and the arc length from its pit end.
func (Lalibela) path(p Vec2) (float64, float64) {
	best, arc, run := math.Inf(1), 0.0, 0.0
	for i := 0; i+1 < len(lalibelaPath); i++ {
		a, b := lalibelaPath[i], lalibelaPath[i+1]
		ab := b.Sub(a)
		l := ab.Length()
		t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/(l*l)))
		d := p.Sub(a.Add(ab.Scale(t))).Length()
		if d < best {
			best, arc = d, run+t*l
		}
		run += l
	}
	return best, arc
}

// height is the ground surface: a rise falling toward the front and down to the rim,
// with the two pits and the open part of the trench cut into it.
func (s Lalibela) height(x, z float64) float64 {
	p := P(x, z)
	if s.inRect(lalibelaPitG, p) || s.inRect(lalibelaPitM, p) {
		return lalibelaFloor
	}
	h := s.natural(x, z)
	if d, arc := s.path(p); d <= 1.6 && arc >= 4 {
		h = math.Min(h, lalibelaFloor+(arc-4)*0.3)
	}
	return math.Max(h, lalibelaRim)
}

// natural is the rise before excavation: higher at the back, sloping to the rim.
func (s Lalibela) natural(x, z float64) float64 {
	top := 20.5 - 0.16*z - 0.08*(x-12) + 2.4*(s.fbm(x/14, z/14, 1)-0.5)
	de := math.Min(math.Min(x-lalibelaX0, lalibelaX1-x), math.Min(z-lalibelaZ0, lalibelaZ1-z))
	de += 5*(s.fbm(x/9, z/9, 2)-0.5) - 1.5
	f := s.smooth(de / 12)
	h := lalibelaRim + (top-lalibelaRim)*f + 0.8*(s.noise(x/3, z/3, 3)-0.5)*f
	return math.Max(h, lalibelaRim)
}

// lalibelaHuts are the tukuls: center, wall radius, door direction, stone walls.
var lalibelaHuts = []struct {
	c       Vec2
	r, face float64
	stone   bool
}{
	{P(-40, 20), 2.4, 1.6, false},
	{P(-30, 23), 2.2, 1.2, true},
	{P(-50, 21), 2.2, 1.8, false},
	{P(-20, 19.5), 2.0, 2.2, false},
	{P(-43, 28.5), 2.0, 1.5, true},
	{P(41, -30), 2.2, 2.0, false},
	{P(-50, -28), 2.0, 1.0, true},
}

// lalibelaField is the sampled heightfield with bilinear lookup.
type lalibelaField struct {
	x0, z0, step float64
	nx, nz       int
	h            []float64
}

func (m *lalibelaField) at(x, z float64) float64 {
	fx, fz := (x-m.x0)/m.step, (z-m.z0)/m.step
	if fx < 0 || fz < 0 || fx >= float64(m.nx-1) || fz >= float64(m.nz-1) {
		return lalibelaRim
	}
	i, j := int(fx), int(fz)
	tx, tz := fx-float64(i), fz-float64(j)
	k := j*m.nx + i
	a := m.h[k] + (m.h[k+1]-m.h[k])*tx
	b := m.h[k+m.nx] + (m.h[k+m.nx+1]-m.h[k+m.nx])*tx
	return a + (b-a)*tz
}

// nearest is the height at the nearest sample, so pit walls stay sheer.
func (m *lalibelaField) nearest(x, z float64) float64 {
	i := int(math.Round((x - m.x0) / m.step))
	j := int(math.Round((z - m.z0) / m.step))
	if i < 0 || j < 0 || i >= m.nx || j >= m.nz {
		return lalibelaRim
	}
	return m.h[j*m.nx+i]
}

// lalibelaDirs are the 14 lattice neighbor directions, normalized.
var lalibelaDirs = func() []Vec3 {
	var d []Vec3
	for _, n := range Neighbors {
		v := V(float64(n.X), float64(n.Y), float64(n.Z))
		d = append(d, v.Scale(1/v.Length()))
	}
	return d
}()

// caves lists the hermit cells, the tunnel between the pits and the trench tunnel.
func (s Lalibela) caves() []lalibelaCave {
	var c []lalibelaCave
	// Cells in a wall facing +Z at z = wz.
	back := func(wz, x, y, w, h float64) {
		c = append(c, lalibelaCave{
			Block(V(x-w/2, y, wz-3), V(x+w/2, y+h, wz+0.5)),
			Block(V(x-w/2-1.3, y-1.3, wz-4.5), V(x+w/2+1.3, y+h+1.3, wz-0.6)),
		})
	}
	// Cells in a wall facing +X at x = wx.
	west := func(wx, z, y, w, h float64) {
		c = append(c, lalibelaCave{
			Block(V(wx-3, y, z-w/2), V(wx+0.5, y+h, z+w/2)),
			Block(V(wx-4.5, y-1.3, z-w/2-1.3), V(wx-0.6, y+h+1.3, z+w/2+1.3)),
		})
	}
	// Cells in a wall facing -X at x = wx.
	east := func(wx, z, y, w, h float64) {
		c = append(c, lalibelaCave{
			Block(V(wx-0.5, y, z-w/2), V(wx+3, y+h, z+w/2)),
			Block(V(wx+0.6, y-1.3, z-w/2-1.3), V(wx+4.5, y+h+1.3, z+w/2+1.3)),
		})
	}
	back(-25, -7, 2, 2.2, 2.8)
	back(-25, -1, 10, 1.8, 2.2)
	back(-25, 28.5, 2, 2.4, 3)
	back(-25, 33, 13, 1.8, 2.2)
	back(-25, 22, 17, 1.8, 2)
	west(-13, 12, 2, 2.2, 2.8)
	west(-13, 20.5, 8, 1.8, 2.2)
	west(-13, -17, 11, 1.8, 2.2)
	east(37, -14, 2, 2.2, 2.8)
	east(37, 8, 9, 1.8, 2.2)
	back(-20, -40, 2, 2.2, 2.6)
	back(-20, -25, 8, 1.8, 2.2)
	back(-20, -44, 4, 1.8, 2.2)
	east(-19, -16, 2, 2.0, 2.6)
	// Tunnel through the rock between the two pits.
	c = append(c, lalibelaCave{
		Block(V(-19.5, lalibelaFloor, -11), V(-12.5, 5.8, -7.5)),
		Block(V(-18.4, 0, -12.5), V(-13.6, 7.2, -6)),
	})
	// Tunnel from the trench into the pit.
	c = append(c, lalibelaCave{
		Func{Box{V(26, lalibelaFloor, 24), V(32, 6.2, 32)}, func(p Vec3) bool {
			d, arc := s.path(p.Plan())
			return d <= 1.5 && arc < 4.4 && p.Y > lalibelaFloor && p.Y < 6
		}},
		Func{Box{V(25, 0, 25.5), V(33, 7.4, 31)}, func(p Vec3) bool {
			d, arc := s.path(p.Plan())
			return d <= 2.8 && arc > 0.8 && arc < 3.6 && p.Y < 7.4
		}},
	})
	return c
}

func (s Lalibela) Build(g *Grid) Order {
	mp := &lalibelaField{x0: lalibelaX0 - 1, z0: lalibelaZ0 - 1, step: 0.25}
	mp.nx = int((lalibelaX1-lalibelaX0+2)/mp.step) + 2
	mp.nz = int((lalibelaZ1-lalibelaZ0+2)/mp.step) + 2
	mp.h = make([]float64, mp.nx*mp.nz)
	for j := 0; j < mp.nz; j++ {
		for i := 0; i < mp.nx; i++ {
			mp.h[j*mp.nx+i] = s.height(mp.x0+float64(i)*mp.step, mp.z0+float64(j)*mp.step)
		}
	}
	caves := s.caves()
	s.terrain(g, mp, caves)
	s.giyorgis(g)
	s.medhane(g)
	s.plateau(g, mp)
	return Outward(lalibelaG)
}

// terrain fills the tuff as a hollow skin and colors it.
func (s Lalibela) terrain(g *Grid, mp *lalibelaField, caves []lalibelaCave) {
	solid := func(p Vec3) bool {
		if p.X < lalibelaX0 || p.X > lalibelaX1 || p.Z < lalibelaZ0 || p.Z > lalibelaZ1 {
			return false
		}
		if p.Y < 0 {
			return true
		}
		if p.Y > mp.nearest(p.X, p.Z) {
			return false
		}
		for _, c := range caves {
			if c.carve.Bounds().Contains(p) && c.carve.Contains(p) {
				return false
			}
		}
		return true
	}
	box := Box{V(lalibelaX0, 0, lalibelaZ0), V(lalibelaX1, 34, lalibelaZ1)}
	t := NewGridWithSpan(g.Span())
	t.Fill(Func{box, func(p Vec3) bool {
		if !solid(p) {
			return false
		}
		for _, d := range lalibelaDirs {
			if !solid(p.Add(d.Scale(lalibelaSkin))) {
				return true
			}
		}
		return false
	}})
	for _, c := range t.Cells() {
		p := t.World(c)
		g.Use(s.rockColor(p, mp, c, caves)).Add(c)
	}
}

// rockColor picks dry grass, pit floor dust, dark cave interiors and banded tuff walls.
func (s Lalibela) rockColor(p Vec3, mp *lalibelaField, c Cell, caves []lalibelaCave) Color {
	for _, cv := range caves {
		if cv.dark.Bounds().Contains(p) && cv.dark.Contains(p) {
			return Charcoal
		}
	}
	n := Hash01(int(c.X), int(c.Y), int(c.Z), 41)
	h := mp.nearest(p.X, p.Z)
	steep := 0.0
	for _, o := range []Vec2{P(0.6, 0), P(-0.6, 0), P(0, 0.6), P(0, -0.6)} {
		steep = math.Max(steep, h-mp.nearest(p.X+o.X, p.Z+o.Y))
	}
	if h <= lalibelaFloor+0.01 && p.Y < lalibelaFloor+0.3 {
		if h <= lalibelaRim+0.01 {
			goto grass
		}
		// Pit floor: trodden red dust, darker toward the walls.
		switch {
		case n < 0.4:
			return Brick
		case n < 0.62:
			return DarkStone
		case n < 0.68:
			return Sand
		}
		return Terracotta
	}
	if p.Y > h-0.9 && steep < 1.5 {
		goto grass
	}
	{
		// Banded tuff walls with dark weathering streaks and lichen.
		if n < 0.025 {
			return Gold
		}
		if n < 0.04 {
			return Grass
		}
		pit := steep > 3
		col := int(math.Floor(p.X*0.9)) + 131*int(math.Floor(p.Z*0.9))
		if pit && Hash01(col, 7) < 0.16 && p.Y > h-2-12*Hash01(col, 8) {
			return DarkStone
		}
		// The pit walls darken toward the floor.
		if pit && p.Y < 7.5 && n < 0.55-(p.Y-2)*0.07 {
			if n < 0.12 {
				return Charcoal
			}
			return DarkStone
		}
		b := (p.Y + 1.4*s.fbm(p.X/6, p.Z/6, 9) + 0.5*math.Sin((p.X-p.Z)/5)) / 1.5
		k := Hash01(int(math.Floor(b)), 11)
		if n < 0.1 {
			k = 1 - k
		}
		switch {
		case k < 0.42:
			return Terracotta
		case k < 0.72:
			return Brick
		case k < 0.88:
			return Sand
		}
		if pit {
			return DarkStone
		}
		return Brick
	}
grass:
	m := s.fbm(p.X/7, p.Z/7, 21)
	switch {
	case m > 0.6 && n < 0.75:
		return Grass
	case n < 0.06:
		return Grass
	case n < 0.08:
		return Terracotta
	case m < 0.34 && n < 0.4:
		return Terracotta
	}
	return Sand
}

// cross is a Greek cross centered on c.
func (Lalibela) cross(c Vec2, half, reach, y0, y1 float64) Solid {
	return Union(BlockAt(c, 2*reach, 2*half, y0, y1), BlockAt(c, 2*half, 2*reach, y0, y1))
}

// lalibelaFace maps face-local coordinates: u along the outward normal from the face
// plane, t along the face, y up.
type lalibelaFace struct{ o, n, t Vec2 }

func (f lalibelaFace) at(u, t, y float64) Vec3 { return f.o.Add(f.n.Scale(u)).Add(f.t.Scale(t)).At(y) }

func (f lalibelaFace) block(u0, u1, t0, t1, y0, y1 float64) Cuboid {
	return Block(f.at(u0, t0, y0), f.at(u1, t1, y1))
}

func (f lalibelaFace) through() Axis {
	if f.n.X != 0 {
		return AxisX
	}
	return AxisZ
}

// giyorgis builds the cross-shaped church on its stepped plinth.
func (s Lalibela) giyorgis(g *Grid) {
	c := lalibelaG
	const half, reach = lalibelaHalf, lalibelaReach
	const wallTop = 21.0
	g.Use(Terracotta)
	g.Fill(s.cross(c, half+2.4, reach+2.4, lalibelaFloor, 3))
	g.Fill(s.cross(c, half+1.6, reach+1.6, 3, 4))
	g.Fill(s.cross(c, half+0.8, reach+0.8, 4, 5))
	g.Use(Brick)
	g.Fill(s.cross(c, half, reach, 5, wallTop))
	g.With(Terracotta, func() {
		g.Fill(s.cross(c, half+0.6, reach+0.6, 11.4, 12.2))                           // String course.
		g.Fill(s.cross(c, half+0.6, reach+0.6, 19.6, 20.4))                           // Cornice.
		g.With(Brick, func() { g.Fill(s.cross(c, half+0.3, reach+0.3, 20.4, 21.4)) }) // Roof parapet.
	})
	g.Carve(s.cross(c, half-1.25, reach-1.25, 3, wallTop-1.2))

	var faces []lalibelaFace
	var ends []lalibelaFace
	for _, d := range []Vec2{P(1, 0), P(-1, 0), P(0, 1), P(0, -1)} {
		perp := P(-d.Y, d.X)
		ends = append(ends, lalibelaFace{c.Add(d.Scale(reach)), d, perp})
		for _, sg := range []float64{-1, 1} {
			faces = append(faces, lalibelaFace{c.Add(d.Scale((half + reach) / 2)).Add(perp.Scale(half * sg)), perp.Scale(sg), d})
		}
	}
	door := func(f lalibelaFace) bool { return !(f.n.X == 0 && f.n.Y < 0) }
	type spot struct {
		f lalibelaFace
		t float64
	}
	var lower, upper []spot
	for _, f := range ends {
		off := 3.4
		if door(f) {
			off = 4.0
		}
		lower = append(lower, spot{f, -off}, spot{f, off})
		upper = append(upper, spot{f, -3.6}, spot{f, 3.6})
		if !door(f) {
			upper = append(upper, spot{f, 0})
		}
	}
	for _, f := range faces {
		lower = append(lower, spot{f, -2.8}, spot{f, 2.8})
		upper = append(upper, spot{f, 0})
	}
	// Frames first, then openings.
	g.Use(Terracotta)
	for _, w := range upper {
		f := w.f
		g.Fill(Arch{Sill: f.at(0.3, w.t, 12.6), HalfWidth: 1.9, Spring: 3.3, Depth: 0.6, Through: f.through()})
		g.Fill(f.block(0, 0.9, w.t-1.9, w.t+1.9, 12.2, 12.8)) // Sill.
	}
	for _, w := range lower {
		f := w.f
		g.Fill(f.block(0, 0.6, w.t-1.5, w.t+1.5, 5.6, 6.2))
		g.Fill(f.block(0, 0.6, w.t-1.5, w.t+1.5, 9.8, 10.4))
	}
	for _, f := range ends {
		if door(f) {
			g.Fill(f.block(0, 0.6, -2.3, 2.3, 5, 11.0))
		}
	}
	g.Use(Brick)
	for _, w := range upper {
		f := w.f
		g.Carve(Arch{Sill: f.at(0, w.t, 12.9), HalfWidth: 1.25, Spring: 3.0, Depth: 5, Through: f.through()})
		// Pointed ogee tip over the round head.
		g.Carve(f.block(-2.5, 2.5, w.t-0.4, w.t+0.4, 17.0, 18.0))
	}
	for _, w := range lower {
		f := w.f
		g.Carve(f.block(-0.75, 1, w.t-1.0, w.t+1.0, 6.2, 9.8))
		g.Paint(DarkStone, f.block(-1.3, -0.5, w.t-1.0, w.t+1.0, 6.2, 9.8))
	}
	for _, f := range ends {
		if door(f) {
			g.Carve(f.block(-2.5, 1, -1.5, 1.5, 5, 10.4))
		}
	}
	// Dark interior faces seen through the openings.
	g.Paint(Charcoal, s.cross(c, half-0.7, reach-0.7, 2.4, wallTop-1.05))

	// Roof: three nested relief crosses.
	g.With(Terracotta, func() { g.Fill(s.cross(c, half-1.5, reach-1.5, wallTop, wallTop+0.8)) })
	g.With(Brick, func() { g.Fill(s.cross(c, half-3, reach-3, wallTop+0.8, wallTop+1.6)) })
	g.With(Sand, func() { g.Fill(s.cross(c, half-4.5, reach-4.5, wallTop+1.6, wallTop+2.4)) })

	// Lichen and weathering on the church rock.
	for _, cell := range g.Cells() {
		p := g.World(cell)
		if p.Y < 2.3 || math.Abs(p.X-c.X) > reach+3 || math.Abs(p.Z-c.Y) > reach+3 {
			continue
		}
		col, _ := g.ColorOf(cell)
		if col != Brick && col != Terracotta {
			continue
		}
		n := Hash01(int(cell.X), int(cell.Y), int(cell.Z), 51)
		switch {
		case n < 0.035:
			g.Use(Gold)
		case n < 0.05 && p.Y > 14:
			g.Use(Grass)
		case n < 0.12 && col == Brick:
			g.Use(Terracotta)
		default:
			continue
		}
		g.Remove(cell)
		g.Add(cell)
	}
}

// medhane builds the rectangular pillared church in the western courtyard.
func (s Lalibela) medhane(g *Grid) {
	c := lalibelaM
	const bw, bd = 18.0, 11.0 // Body footprint.
	const base, eave = 3.6, 13.0
	g.Use(Terracotta)
	g.Fill(BlockAt(c, 25.5, 18.5, lalibelaFloor, 2.8))
	g.Fill(BlockAt(c, 24.2, 17.2, 2.8, base))
	g.Use(Brick)
	g.Fill(BlockAt(c, bw, bd, base, eave))
	g.Carve(BlockAt(c, bw-2.5, bd-2.5, 2.8, eave-1))
	// Pillars all round, carrying the overhanging roof.
	const pw, pd = 22.0, 15.0
	for i := 0; i <= 8; i++ {
		x := c.X - pw/2 + pw*float64(i)/8
		for _, z := range []float64{c.Y - pd/2, c.Y + pd/2} {
			g.Fill(BlockAt(P(x, z), 1.3, 1.3, base, eave))
		}
	}
	for i := 1; i < 5; i++ {
		z := c.Y - pd/2 + pd*float64(i)/5
		for _, x := range []float64{c.X - pw/2, c.X + pw/2} {
			g.Fill(BlockAt(P(x, z), 1.3, 1.3, base, eave))
		}
	}
	// Windows: small arched upper lights and square lower ones on the long sides.
	faces := []lalibelaFace{
		{c.Add(P(0, bd/2)), P(0, 1), P(1, 0)},
		{c.Add(P(0, -bd/2)), P(0, -1), P(1, 0)},
		{c.Add(P(bw/2, 0)), P(1, 0), P(0, 1)},
		{c.Add(P(-bw/2, 0)), P(-1, 0), P(0, 1)},
	}
	for k, f := range faces {
		var ts []float64
		if k < 2 {
			ts = []float64{-6.75, -4.05, -1.35, 1.35, 4.05, 6.75}
		} else {
			ts = []float64{-3.2, 3.2}
		}
		for _, t := range ts {
			g.Carve(Arch{Sill: f.at(0, t, 8.4), HalfWidth: 0.8, Spring: 1.6, Depth: 4, Through: f.through()})
			g.Carve(f.block(-2, 2, t-0.7, t+0.7, 4.8, 6.4))
		}
		if k >= 2 || k == 0 {
			g.Carve(f.block(-2, 2, -1.4, 1.4, base, 7.8))
		}
	}
	g.Paint(Charcoal, BlockAt(c, bw-1.4, bd-1.4, 3, eave-0.6))
	// Roof: slab over the colonnade and a low gable, ridge along X.
	g.Use(Terracotta)
	g.Fill(BlockAt(c, pw+2, pd+2, eave, eave+0.9))
	g.Use(Brick)
	g.Fill(Func{Box{V(c.X-pw/2-1, eave+0.9, c.Y-pd/2-1), V(c.X+pw/2+1, eave+3.4, c.Y+pd/2+1)}, func(p Vec3) bool {
		return p.Y <= eave+0.9+2.5*(1-math.Abs(p.Z-c.Y)/(pd/2+1))
	}})
	g.Carve(Func{Box{V(c.X-pw/2+0.5, eave+0.9, c.Y-pd/2+0.5), V(c.X+pw/2-0.5, eave+2.4, c.Y+pd/2-0.5)}, func(p Vec3) bool {
		return p.Y <= eave+0.9+2.5*(1-math.Abs(p.Z-c.Y)/(pd/2+1))-1.2
	}})
	g.With(Terracotta, func() {
		g.Fill(Block(V(c.X-pw/2-1, eave+3, c.Y-0.6), V(c.X+pw/2+1, eave+3.6, c.Y+0.6)))
	})
	// Raised ribs across the roof, one over each pillar bay.
	for i := 0; i <= 8; i++ {
		x := c.X - pw/2 + pw*float64(i)/8
		g.With(Terracotta, func() {
			g.Fill(Func{Box{V(x-0.4, eave+0.9, c.Y-pd/2-1), V(x+0.4, eave+4, c.Y+pd/2+1)}, func(p Vec3) bool {
				return p.Y <= eave+1.6+2.5*(1-math.Abs(p.Z-c.Y)/(pd/2+1))
			}})
		})
	}
}

// plateau places tukuls, trees and pilgrims on the ground surface.
func (s Lalibela) plateau(g *Grid, mp *lalibelaField) {
	ground := func(x, z float64) float64 {
		h := mp.nearest(x, z)
		for _, o := range []Vec2{P(1.5, 0), P(-1.5, 0), P(0, 1.5), P(0, -1.5)} {
			h = math.Min(h, mp.nearest(x+o.X, z+o.Y))
		}
		return h
	}
	tukul := func(c Vec2, r float64, face float64, stone bool) {
		b, hi := math.Inf(1), math.Inf(-1)
		for k := 0; k < 12; k++ {
			o := Polar(c, r+0.6, float64(k)*math.Pi/6)
			h := mp.nearest(o.X, o.Y)
			b, hi = math.Min(b, h), math.Max(hi, h)
		}
		b = math.Max(b, hi-2) - 0.3
		// A stone foundation ring levels the slope.
		g.With(Brick, func() { g.Fill(Cylinder{Center: c, Radius: r + 0.5, Inner: r - 0.8, Y0: b - 2.2, Y1: b + 0.6}) })
		wall := Wood
		if stone {
			wall = Terracotta
		}
		g.With(wall, func() { g.Fill(Cylinder{Center: c, Radius: r, Inner: r - 1.1, Y0: b, Y1: b + 3}) })
		if stone {
			g.Paint(Brick, Func{Box{V(c.X-r, b, c.Y-r), V(c.X+r, b+3, c.Y+r)}, func(p Vec3) bool {
				return Hash01(int(math.Floor(p.X*1.4)), int(math.Floor(p.Y*1.8)), int(math.Floor(p.Z*1.4)), 61) < 0.35
			}})
		}
		g.With(Charcoal, func() { g.Fill(Cylinder{Center: c, Radius: r - 1, Y0: b, Y1: b + 0.6}) })
		door := Polar(c, r-0.5, face)
		g.Carve(Cylinder{Center: door, Radius: 0.8, Y0: b + 0.6, Y1: b + 2.4})
		g.With(Sand, func() { g.Fill(Frustum{c, r + 1.1, 0.2, b + 2.8, b + 2.8 + r*1.6}) })
		g.Carve(Frustum{c, r - 0.2, 0, b + 2.8, b + 2.8 + r*1.6 - 1.4})
		g.With(Wood, func() { g.Fill(Cylinder{Center: c, Radius: 0.45, Y0: b + 2.8 + r*1.6 - 0.6, Y1: b + 3.6 + r*1.6}) })
	}
	eucalyptus := func(c Vec2, h float64, k int) {
		b := ground(c.X, c.Y) - 0.4
		g.With(Sand, func() { g.Fill(Cylinder{Center: c, Radius: 0.55, Y0: b, Y1: b + h*0.75}) })
		g.With(Foliage, func() {
			g.FillVolume(Ellipsoid{c.At(b + h*0.7), V(2.1, h*0.28, 2.1)})
			o := Polar(c, 1.2, Hash01(k, 1)*6.28)
			g.FillVolume(Ellipsoid{o.At(b + h*0.55), V(1.6, h*0.18, 1.6)})
		})
		g.Paint(Grass, Func{Box{V(c.X-3, b+h*0.3, c.Y-3), V(c.X+3, b+h, c.Y+3)}, func(p Vec3) bool {
			return p.Y > b+h*0.75 && Hash01(int(math.Floor(p.X*1.3)), int(math.Floor(p.Y*1.3)), int(math.Floor(p.Z*1.3)), 71) < 0.4
		}})
	}
	acacia := func(c Vec2, k int) {
		b := ground(c.X, c.Y) - 0.4
		a := Hash01(k, 3) * 6.28
		g.With(Wood, func() {
			g.Fill(Capsule{c.At(b), c.At(b + 3), 0.55})
			for i := 0; i < 3; i++ {
				o := Polar(c, 2.4, a+float64(i)*2.1)
				g.Fill(Capsule{c.At(b + 2.6), o.At(b + 4.6), 0.5})
			}
		})
		g.With(Foliage, func() {
			g.FillVolume(EllipticCylinder{c, 4.2, 3.6, b + 4.6, b + 5.6})
			g.FillVolume(Ellipsoid{c.At(b + 5.6), V(3.4, 0.9, 3)})
		})
		g.Paint(Grass, Func{Box{V(c.X-5, b+5.4, c.Y-5), V(c.X+5, b+7, c.Y+5)}, func(p Vec3) bool {
			return Hash01(int(math.Floor(p.X*1.2)), int(math.Floor(p.Z*1.2)), 81) < 0.45
		}})
	}
	pilgrim := func(x, z float64) {
		b := mp.nearest(x, z)
		g.With(White, func() {
			g.Fill(Capsule{V(x, b+0.5, z), V(x, b+1.9, z), 0.5})
		})
	}

	// A hamlet of tukuls on the western slope and a few by the back rim.
	for _, t := range lalibelaHuts {
		tukul(t.c, t.r, t.face, t.stone)
	}

	for i, t := range []struct {
		p Vec2
		h float64
	}{
		{P(-24, -30), 12}, {P(-15, -28), 10}, {P(-3, -31), 13}, {P(16, -31.5), 11},
		{P(-53, 8), 10}, {P(-53, 31), 9}, {P(41.5, 6), 11}, {P(41.5, -14), 9}, {P(40, 33), 10},
	} {
		eucalyptus(t.p, t.h, i)
	}
	for i, p := range []Vec2{P(28, -31), P(-40, -28), P(-16, 31), P(41.5, 19), P(-8, 34), P(-53, -14), P(34, 35)} {
		acacia(p, i)
	}
	// Pilgrims: around the church, in the trench and at the pit rim.
	for _, p := range []Vec2{
		P(15, 23), P(17.5, 23.6), P(1, 22.5), P(33, 15), P(34.5, 16.5), P(33, -17), P(-8, -21),
		P(29.5, 23), P(14, 30.5), P(18, 30.5), P(0, 30.5),
		P(-8, 28), P(-10, 27.5), P(6, 27.6), P(41, -6), P(-16, 4),
		P(-46.3, -10), P(-19.7, -15), P(-30, 6), P(-27.5, 7), P(-38, 9),
	} {
		pilgrim(p.X, p.Y)
	}
}
