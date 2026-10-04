package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// TreehouseVillage is a fantasy forest village built in five giant trees around a campfire
// clearing: hollow buttressed trunks with doors, plank platforms ringing the trunks with
// railings and lanterns, round thatched and shingled huts, spiral stairs, ladders, rope
// bridges between the trees, a pulley lift, and layered canopies (one in autumn colours).
// The forest floor has a stream with a waterwheel mill, stepping stones, ferns and giant
// mushrooms. The clearing opens toward the hero camera (+X/+Z).
type TreehouseVillage struct{}

func init() { Register(TreehouseVillage{}) }

func (TreehouseVillage) Name() string  { return "treehouse_village" }
func (TreehouseVillage) Title() string { return "Treehouse Village" }

const (
	treehouseVillageWater     = 0.8  // Stream surface.
	treehouseVillageLiftAngle = -0.6 // Lift direction from the front-left tree, over the open ground.
	treehouseVillageA         = 44.0
	treehouseVillageB         = 42.0
)

var (
	treehouseVillageSite     = P(1, 1)
	treehouseVillageClearing = P(1, 4)
	treehouseVillageFire     = P(1.5, 4.5)
	treehouseVillageStream   = []Vec2{P(52, 0), P(36, 10), P(25, 18), P(14, 25.5), P(4, 32), P(-6, 40), P(-12, 50)}
	treehouseVillageWheel    = P(21.5, 20.4) // On the stream, beside the mill.
	treehouseVillageStones   = P(9, 28.75)   // Stepping stone crossing.
)

type treehouseVillageHut struct {
	level int
	angle float64 // Degrees around the trunk.
	roof  Color   // Sand thatch or Terracotta shingle.
	wall  Color
	door  float64 // +1 or -1: door faces along the ring.
}

type treehouseVillageTree struct {
	c      Vec2
	r      float64   // Trunk radius above the flare.
	top    float64   // Height where the crown branches out.
	levels []float64 // Platform deck heights.
	huts   []treehouseVillageHut
	leaf   int // 0 green, 1 turning, 2 autumn.
	seed   int
	g0     float64 // Ground at the trunk.
	door   float64 // Door angle (radians), toward the clearing.
	// Angles (radians) kept clear of roots and braces: ladders, stairs, door.
	keep []float64
}

func (t *treehouseVillageTree) rad(y float64) float64 {
	h := math.Max(0, math.Min(1, (y-t.g0)/(t.top-t.g0)))
	return t.r*(1-0.32*h) + 2.4*math.Exp(-math.Max(0, y-t.g0)/2.2)
}

func (t *treehouseVillageTree) deckR(level int) float64 { return t.rad(t.levels[level]) + 3.8 }

func (t *treehouseVillageTree) hutAt(h treehouseVillageHut) Vec2 {
	y := t.levels[h.level]
	return Polar(t.c, t.rad(y)+6.0, h.angle*math.Pi/180)
}

// treehouseVillagePlatform is one deck: a ring around the trunk plus round hut bays.
type treehouseVillagePlatform struct {
	tree  *treehouseVillageTree
	level int
	y, R  float64
	bays  []Vec2
	gaps  []Vec2 // Railing openings.
	hatch []Vec2 // Ladder holes.
	stair bool   // A spiral stair arrives: no braces.
}

const treehouseVillageBay = 4.1

func (pl *treehouseVillagePlatform) in(p Vec2, grow float64) bool {
	if p.Sub(pl.tree.c).Length() <= pl.R+grow {
		return true
	}
	for _, b := range pl.bays {
		if p.Sub(b).Length() <= treehouseVillageBay+grow {
			return true
		}
	}
	return false
}

func (pl *treehouseVillagePlatform) nearGap(p Vec2, d float64) bool {
	for _, q := range pl.gaps {
		if p.Sub(q).Length() < d {
			return true
		}
	}
	return false
}

type treehouseVillageBridge struct {
	a, la, b, lb int
}

type treehouseVillageMap struct {
	x0, z0, step float64
	nx, nz       int
	h            []float64
}

func (m *treehouseVillageMap) at(x, z float64) float64 {
	fx, fz := (x-m.x0)/m.step, (z-m.z0)/m.step
	if fx < 0 || fz < 0 || fx >= float64(m.nx-1) || fz >= float64(m.nz-1) {
		return -1
	}
	i, j := int(fx), int(fz)
	tx, tz := fx-float64(i), fz-float64(j)
	k := j*m.nx + i
	a := m.h[k] + (m.h[k+1]-m.h[k])*tx
	b := m.h[k+m.nx] + (m.h[k+m.nx+1]-m.h[k+m.nx])*tx
	return a + (b-a)*tz
}

func (m *treehouseVillageMap) low(x, z float64) float64 {
	lo := m.at(x, z)
	for k := 0; k < 8; k++ {
		a := float64(k) * math.Pi / 4
		lo = math.Min(lo, m.at(x+1.1*math.Cos(a), z+1.1*math.Sin(a)))
	}
	return lo
}

func (TreehouseVillage) noise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

func (tv TreehouseVillage) fbm(x, z float64, seed int) float64 {
	return (tv.noise(x, z, seed)*4 + tv.noise(x*2.1, z*2.1, seed+1)*2 + tv.noise(x*4.3, z*4.3, seed+2)) / 7
}

func (TreehouseVillage) segDist(line []Vec2, p Vec2) float64 {
	best := math.Inf(1)
	for i := 0; i+1 < len(line); i++ {
		a, b := line[i], line[i+1]
		ab := b.Sub(a)
		t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
		best = math.Min(best, p.Sub(a.Add(ab.Scale(t))).Length())
	}
	return best
}

func (TreehouseVillage) angDiff(a, b float64) float64 {
	d := math.Mod(a-b, 2*math.Pi)
	if d > math.Pi {
		d -= 2 * math.Pi
	}
	if d < -math.Pi {
		d += 2 * math.Pi
	}
	return math.Abs(d)
}

func (tv TreehouseVillage) siteE(x, z float64) float64 {
	ex, ez := (x-treehouseVillageSite.X)/treehouseVillageA, (z-treehouseVillageSite.Y)/treehouseVillageB
	return math.Sqrt(ex*ex+ez*ez) + 0.06*(tv.fbm(x/9, z/9, 3)-0.5)
}

func (tv TreehouseVillage) streamDist(p Vec2) float64 {
	return tv.segDist(treehouseVillageStream, p) - 2.4 - 0.7*(tv.fbm(p.X/5, p.Y/5, 10)-0.5)
}

func (tv TreehouseVillage) trees() []*treehouseVillageTree {
	at := func(deg, r float64) Vec2 { return Polar(P(0, 0), r, deg*math.Pi/180) }
	return []*treehouseVillageTree{
		{c: at(232, 26), r: 5.4, top: 38, levels: []float64{11, 21, 30}, leaf: 0, seed: 1,
			huts: []treehouseVillageHut{{0, 20, Sand, White, 1}, {1, 120, Terracotta, White, -1}, {2, 50, Sand, White, 1}}},
		{c: at(296, 27), r: 4.6, top: 33, levels: []float64{13, 24}, leaf: 1, seed: 2,
			huts: []treehouseVillageHut{{0, 30, Terracotta, White, -1}, {1, 85, Sand, White, 1}}},
		{c: at(352, 28), r: 4.2, top: 31, levels: []float64{9, 17.5, 25}, leaf: 2, seed: 3,
			huts: []treehouseVillageHut{{0, 100, Sand, White, 1}, {1, 10, Terracotta, White, 1}, {2, 120, Sand, White, -1}}},
		{c: at(176, 27), r: 4.6, top: 32, levels: []float64{10, 21}, leaf: 0, seed: 4,
			huts: []treehouseVillageHut{{0, 70, Terracotta, White, 1}, {1, 120, Sand, White, -1}}},
		{c: at(118, 27), r: 4.0, top: 27, levels: []float64{9, 18}, leaf: 1, seed: 5,
			huts: []treehouseVillageHut{{0, 20, Sand, White, 1}, {1, 160, Terracotta, White, 1}}},
	}
}

var treehouseVillageBridges = []treehouseVillageBridge{
	{0, 1, 1, 1}, // A–B high.
	{1, 0, 2, 1}, // B–C.
	{0, 0, 3, 0}, // A–D low.
	{3, 1, 4, 1}, // D–E.
	{1, 1, 2, 2}, // B top – C top.
}

// height is the forest floor: gentle undulation, mounds around the trees, a flat
// clearing, the stream channel and a low rim.
func (tv TreehouseVillage) height(x, z float64, trees []*treehouseVillageTree) float64 {
	e := tv.siteE(x, z)
	if e > 1 {
		return -1
	}
	p := P(x, z)
	h := 1.8 + 1.8*(tv.fbm(x/14, z/14, 1)-0.5)
	h += math.Max(0, -(z+5)/40) * 2.5 // Rising toward the back.
	for _, t := range trees {
		d := p.Sub(t.c).Length()
		if d < t.r+12 {
			k := 1 - d/(t.r+12)
			h += 1.6 * k * k
		}
	}
	if d := p.Sub(treehouseVillageClearing).Length(); d < 13 {
		w := math.Min(1, (13-d)/5)
		w = w * w * (3 - 2*w)
		h += (1.9 - h) * w
	}
	h = math.Min(h, 0.4+14*(1-e))
	ds := tv.streamDist(p)
	h = math.Min(h, treehouseVillageWater-0.1+math.Max(0, ds)*0.7)
	return math.Max(h, 0.4)
}

func (tv TreehouseVillage) Build(g *Grid) Order {
	trees := tv.trees()
	m := &treehouseVillageMap{x0: -52, z0: -48, step: 0.4}
	m.nx, m.nz = int(104/m.step)+2, int(102/m.step)+2
	m.h = make([]float64, m.nx*m.nz)
	for j := 0; j < m.nz; j++ {
		for i := 0; i < m.nx; i++ {
			m.h[j*m.nx+i] = tv.height(m.x0+float64(i)*m.step, m.z0+float64(j)*m.step, trees)
		}
	}
	for _, t := range trees {
		t.g0 = m.at(t.c.X, t.c.Y)
		t.door = treehouseVillageClearing.Sub(t.c).Angle()
	}
	// Platforms with hut bays.
	plats := make([][]*treehouseVillagePlatform, len(trees))
	for i, t := range trees {
		for l, y := range t.levels {
			pl := &treehouseVillagePlatform{tree: t, level: l, y: y, R: t.deckR(l)}
			for _, h := range t.huts {
				if h.level == l {
					pl.bays = append(pl.bays, t.hutAt(h))
				}
			}
			plats[i] = append(plats[i], pl)
		}
	}
	// Access: spiral stairs on the two biggest trees, ladders elsewhere.
	type ladder struct {
		t      *treehouseVillageTree
		a      float64
		y0, y1 float64
	}
	type stair struct {
		t          *treehouseVillageTree
		a0, turns  float64
		y0, y1     float64
		arrive     *treehouseVillagePlatform
		startFloor bool
	}
	var ladders []ladder
	var stairs []stair
	for i, t := range trees {
		switch i {
		case 0, 3:
			// Stair from the ground, starting beside the door.
			a0 := t.door + 0.5
			st := stair{t: t, a0: a0, turns: 1.25, y0: t.g0, y1: t.levels[0], arrive: plats[i][0], startFloor: true}
			stairs = append(stairs, st)
			plats[i][0].stair = true
			for k := 0; k < 6; k++ {
				t.keep = append(t.keep, a0+float64(k)*0.4)
			}
			// Ladders between the upper levels.
			for l := 0; l+1 < len(t.levels); l++ {
				a := t.door + math.Pi + float64(l)*0.9
				ladders = append(ladders, ladder{t, a, t.levels[l], t.levels[l+1]})
				t.keep = append(t.keep, a)
				plats[i][l+1].hatch = append(plats[i][l+1].hatch, Polar(t.c, t.rad(t.levels[l+1])+0.9, a))
			}
		default:
			a := t.door - 0.6
			ladders = append(ladders, ladder{t, a, t.g0, t.levels[0]})
			t.keep = append(t.keep, a)
			plats[i][0].hatch = append(plats[i][0].hatch, Polar(t.c, t.rad(t.levels[0])+0.9, a))
			for l := 0; l+1 < len(t.levels); l++ {
				a := t.door + math.Pi*0.75 + float64(l)*1.1
				ladders = append(ladders, ladder{t, a, t.levels[l], t.levels[l+1]})
				t.keep = append(t.keep, a)
				plats[i][l+1].hatch = append(plats[i][l+1].hatch, Polar(t.c, t.rad(t.levels[l+1])+0.9, a))
			}
		}
		t.keep = append(t.keep, t.door)
	}
	// Bridge ends open the railings.
	type bridgeEnds struct {
		p1, p2 Vec3
	}
	var bridges []bridgeEnds
	for _, b := range treehouseVillageBridges {
		ta, tb := trees[b.a], trees[b.b]
		pa, pb := plats[b.a][b.la], plats[b.b][b.lb]
		u := tb.c.Sub(ta.c)
		u = u.Scale(1 / u.Length())
		p1 := ta.c.Add(u.Scale(pa.R - 0.3))
		p2 := tb.c.Sub(u.Scale(pb.R - 0.3))
		pa.gaps = append(pa.gaps, p1)
		pb.gaps = append(pb.gaps, p2)
		bridges = append(bridges, bridgeEnds{p1.At(pa.y), p2.At(pb.y)})
	}

	plats[4][1].gaps = append(plats[4][1].gaps, Polar(trees[4].c, plats[4][1].R-0.3, treehouseVillageLiftAngle))
	tv.terrain(g, m, trees)
	tv.stream(g, m)
	for _, t := range trees {
		tv.trunk(g, t)
	}
	for _, s := range stairs {
		tv.spiral(g, s.t, s.a0, s.turns, s.y0, s.y1, s.arrive)
	}
	for i := range trees {
		for _, pl := range plats[i] {
			tv.platform(g, pl)
		}
	}
	for _, l := range ladders {
		tv.ladder(g, l.t, l.a, l.y0, l.y1)
	}
	for i, t := range trees {
		for _, h := range t.huts {
			tv.hut(g, plats[i][h.level], t.hutAt(h), h)
		}
	}
	for _, b := range bridges {
		tv.bridge(g, b.p1, b.p2)
	}
	tv.lift(g, trees[4], plats[4][1], m)
	for _, t := range trees {
		tv.crown(g, t)
	}
	tv.mill(g, m)
	tv.clearing(g, m)
	tv.floor(g, m, trees)
	return Outward(treehouseVillageClearing)
}

// terrain fills the forest floor as a turf shell with moss, paths and stream banks.
func (tv TreehouseVillage) terrain(g *Grid, m *treehouseVillageMap, trees []*treehouseVillageTree) {
	box := Box{V(m.x0, -1, m.z0), V(m.x0+104, 12, m.z0+102)}
	g.Use(Grass).Fill(Func{box, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		if h <= 0 {
			return false
		}
		return p.Y <= h && p.Y >= math.Min(h, m.low(p.X, p.Z))-0.75
	}})
	top := func(p Vec3) bool { return p.Y >= m.at(p.X, p.Z)-0.9 }
	// Moss under the trees and in patches.
	g.Paint(Foliage, Func{box, func(p Vec3) bool {
		if !top(p) {
			return false
		}
		q := p.Plan()
		n := tv.fbm(q.X/7, q.Y/7, 20)
		for _, t := range trees {
			if q.Sub(t.c).Length() < t.r+9+4*n {
				return n > 0.35
			}
		}
		return n > 0.62
	}})
	// Paths: clearing to each tree door and to the stepping stones.
	var paths [][]Vec2
	for _, t := range trees {
		end := Polar(t.c, t.rad(t.g0)+0.5, t.door)
		mid := treehouseVillageClearing.Add(end).Scale(0.5).Add(P(Hash01(t.seed, 9)*4-2, Hash01(t.seed, 10)*4-2))
		paths = append(paths, []Vec2{treehouseVillageClearing, mid, end})
	}
	paths = append(paths, []Vec2{treehouseVillageClearing, P(5.5, 15), P(7.5, 22), treehouseVillageStones},
		[]Vec2{P(5.5, 15), P(11, 13.5), tv.millHut()})
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		if !top(p) {
			return false
		}
		q := p.Plan()
		w := 1.2 + 0.35*(tv.fbm(q.X/4, q.Y/4, 30)-0.5)
		for _, pth := range paths {
			if tv.segDist(pth, q) < w {
				return true
			}
		}
		return q.Sub(treehouseVillageClearing).Length() < 8.5+1.5*(tv.fbm(q.X/3, q.Y/3, 31)-0.5)
	}})
	// Pebbly stream banks.
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		if !top(p) {
			return false
		}
		d := tv.streamDist(p.Plan())
		return d > -0.6 && d < 0.8 && Hash01(int(math.Round(p.X*3)), int(math.Round(p.Z*3)), 33) < 0.6
	}})
}

// stream paints the water surface in the channel.
func (tv TreehouseVillage) stream(g *Grid, m *treehouseVillageMap) {
	box := Box{V(-20, -1, -2), V(56, 3, 58)}
	g.Paint(Water, Func{box, func(p Vec3) bool {
		return p.Y >= m.at(p.X, p.Z)-0.9 && tv.streamDist(p.Plan()) < -0.4
	}})
	// A layer of water where the bed sits low.
	g.Use(Water).Fill(Func{box, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		return h > 0 && p.Y > h && p.Y <= treehouseVillageWater && tv.streamDist(p.Plan()) < 0
	}})
}

// trunk is the hollow buttressed trunk with bark furrows, a door and windows.
func (tv TreehouseVillage) trunk(g *Grid, t *treehouseVillageTree) {
	R0 := t.rad(t.g0)
	box := Box{V(t.c.X-R0-8, t.g0-1.5, t.c.Y-R0-8), V(t.c.X+R0+8, t.top+4, t.c.Y+R0+8)}
	g.Use(Wood).FillVolume(Func{box, func(p Vec3) bool {
		if p.Y < t.g0-1.2 || p.Y > t.top+3 {
			return false
		}
		d := p.Plan().Sub(t.c).Length()
		r := t.rad(p.Y)
		if p.Y > t.top+1 {
			return d <= r
		}
		return d <= r && d >= r-1.4
	}})
	// Buttress roots between the kept angles.
	for k := 0; k < 8; k++ {
		a := float64(k)*2*math.Pi/8 + Hash01(t.seed, k, 1)*0.4
		skip := false
		for _, kk := range t.keep {
			if tv.angDiff(a, kk) < 0.42 {
				skip = true
			}
		}
		if skip {
			continue
		}
		L := 5.5 + 2*Hash01(t.seed, k, 2)
		hgt := 6.5 + 3*Hash01(t.seed, k, 3)
		g.Fill(Func{box, func(p Vec3) bool {
			q := p.Plan().Sub(t.c)
			s, c := math.Sincos(a)
			u := q.X*c + q.Y*s - t.r + 0.5
			v := -q.X*s + q.Y*c
			if u < 0 || u > L || math.Abs(v) > 0.8+0.5*(1-u/L) || p.Y < t.g0-1.2 {
				return false
			}
			return p.Y <= t.g0+hgt*math.Pow(1-u/L, 1.6)
		}})
		end := Polar(t.c, t.r+L-0.5, a)
		far := Polar(t.c, t.r+L+2.5, a+0.3*(Hash01(t.seed, k, 4)-0.5))
		g.Fill(Capsule{end.At(t.g0 + 0.2), far.At(t.g0 - 0.5), 0.75})
	}
	// Bark: vertical furrows of charcoal.
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		q := p.Plan().Sub(t.c)
		d := q.Length()
		if d < t.rad(p.Y)-0.5 && p.Y < t.top+1 {
			return false
		}
		a := q.Angle()
		s := a*t.r + 0.8*math.Sin(p.Y/2.7+float64(t.seed))
		i := int(math.Floor(s / 1.3))
		j := int(math.Floor((p.Y + Hash01(i, t.seed)*5) / 5))
		f := s/1.3 - math.Floor(s/1.3)
		return f < 0.3 && Hash01(i, j, t.seed, 5) < 0.6
	}})
	// Door into the hollow trunk.
	sill := t.c.Add(P(t.rad(t.g0+2)-0.6, 0)).At(t.g0 + 0.2)
	g.Carve(RotateY(Arch{Sill: sill, HalfWidth: 1.25, Spring: 1.9, Depth: 6, Through: AxisX}, t.c, t.door))
	// Windows between the platforms.
	for k := 0; k < 4; k++ {
		y := t.g0 + 4.5 + float64(k)*(t.top-t.g0-6)/4 + Hash01(t.seed, k, 6)*2
		a := t.door + 0.9 + float64(k)*1.9 + Hash01(t.seed, k, 7)*0.4
		r := t.rad(y)
		g.Carve(RotateY(Arch{Sill: t.c.Add(P(r-0.7, 0)).At(y), HalfWidth: 0.7, Spring: 0.7, Depth: 2.4, Through: AxisX}, t.c, a))
	}
	// Dark interior.
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		return p.Y < t.top+0.5 && p.Plan().Sub(t.c).Length() < t.rad(p.Y)-0.7
	}})
}

// spiral is a helical stair hugging the trunk, with an outer handrail and a deck hatch.
func (tv TreehouseVillage) spiral(g *Grid, t *treehouseVillageTree, a0, turns, y0, y1 float64, arrive *treehouseVillagePlatform) {
	const perTurn = 18.0
	n := int(math.Round(turns * perTurn))
	rise := (y1 - y0) / float64(n)
	da := 2 * math.Pi / perTurn
	box := Box{V(t.c.X-t.rad(y0)-4, y0-1, t.c.Y-t.rad(y0)-4), V(t.c.X+t.rad(y0)+4, y1+2, t.c.Y+t.rad(y0)+4)}
	step := func(p Vec3, k0 int) (int, bool) {
		q := p.Plan().Sub(t.c)
		u := math.Mod(q.Angle()-a0, 2*math.Pi)
		if u < 0 {
			u += 2 * math.Pi
		}
		d := q.Length()
		for w := 0; w <= int(turns)+1; w++ {
			k := int(math.Floor((u + float64(w)*2*math.Pi) / da))
			if k < k0 || k >= n {
				continue
			}
			ytop := y0 + float64(k+1)*rise
			r := t.rad(ytop)
			if p.Y <= ytop+0.1 && p.Y >= ytop-0.75 && d >= r-0.4 && d <= r+2.3 {
				return k, true
			}
		}
		return 0, false
	}
	// Hatch through the deck for the last steps.
	g.Carve(Func{box, func(p Vec3) bool {
		if p.Y < arrive.y-1.2 || p.Y > arrive.y+0.2 {
			return false
		}
		q := p.Plan().Sub(t.c)
		u := math.Mod(q.Angle()-a0, 2*math.Pi)
		if u < 0 {
			u += 2 * math.Pi
		}
		d := q.Length()
		r := t.rad(arrive.y)
		for w := 0; w <= int(turns)+1; w++ {
			k := int(math.Floor((u + float64(w)*2*math.Pi) / da))
			if k >= n-7 && k < n && d >= r-0.2 && d <= r+2.5 {
				return true
			}
		}
		return false
	}})
	g.Use(Sand).Fill(Func{box, func(p Vec3) bool { _, ok := step(p, 0); return ok }})
	// Outer handrail and posts.
	g.Use(Wood).Fill(Func{box, func(p Vec3) bool {
		q := p.Plan().Sub(t.c)
		u := math.Mod(q.Angle()-a0, 2*math.Pi)
		if u < 0 {
			u += 2 * math.Pi
		}
		d := q.Length()
		for w := 0; w <= int(turns)+1; w++ {
			s := (u + float64(w)*2*math.Pi) / da
			if s < 0 || s > float64(n) {
				continue
			}
			yr := y0 + s*rise + 1.4
			r := t.rad(yr - 1.4)
			if math.Abs(d-(r+2.0)) > 0.45 {
				continue
			}
			if math.Abs(p.Y-yr) < 0.35 && s < float64(n)-4 {
				return true
			}
			if int(s)%3 == 0 && s-math.Floor(s) < 0.45 && p.Y < yr && p.Y > yr-1.9 {
				return true
			}
		}
		return false
	}})
}

// platform is a plank deck with a beam rim, railings, lanterns and braces.
func (tv TreehouseVillage) platform(g *Grid, pl *treehouseVillagePlatform) {
	t := pl.tree
	y := pl.y
	ext := pl.R + 2*treehouseVillageBay + 3
	box := Box{V(t.c.X-ext, y-6, t.c.Y-ext), V(t.c.X+ext, y+2.6, t.c.Y+ext)}
	rt := t.rad(y)
	g.Use(Sand).Fill(Func{box, func(p Vec3) bool {
		q := p.Plan()
		return p.Y >= y-0.8 && p.Y <= y && pl.in(q, 0) && q.Sub(t.c).Length() >= rt-0.3
	}})
	// Plank seams.
	g.Paint(Wood, Func{box, func(p Vec3) bool {
		q := p.Plan()
		if p.Y < y-0.8 || p.Y > y || !pl.in(q, 0) {
			return false
		}
		return !pl.in(q, -0.8) // Rim beam.
	}})
	for _, h := range pl.hatch {
		g.Carve(Cylinder{Center: h, Radius: 1.0, Y0: y - 1, Y1: y + 0.3})
	}
	// Railing posts along the edge, top rail, lanterns on every third post.
	var posts []Vec2
	n := int(math.Round(2 * math.Pi * pl.R / 2.3))
	for k := 0; k < n; k++ {
		q := Polar(t.c, pl.R-0.35, float64(k)*2*math.Pi/float64(n)+float64(t.seed))
		inBay := false
		for _, b := range pl.bays {
			if q.Sub(b).Length() < treehouseVillageBay-0.2 {
				inBay = true
			}
		}
		if !inBay {
			posts = append(posts, q)
		}
	}
	for _, b := range pl.bays {
		nb := int(math.Round(2 * math.Pi * treehouseVillageBay / 2.3))
		for k := 0; k < nb; k++ {
			q := Polar(b, treehouseVillageBay-0.35, float64(k)*2*math.Pi/float64(nb))
			if q.Sub(t.c).Length() < pl.R+0.1 {
				continue
			}
			clash := false
			for _, b2 := range pl.bays {
				if b2 != b && q.Sub(b2).Length() < treehouseVillageBay-0.2 {
					clash = true
				}
			}
			if !clash {
				posts = append(posts, q)
			}
		}
	}
	g.Use(Wood)
	for i, q := range posts {
		if pl.nearGap(q, 1.6) {
			continue
		}
		g.Fill(Cylinder{Center: q, Radius: 0.42, Y0: y - 0.4, Y1: y + 1.5})
		if i%3 == 0 {
			g.With(Gold, func() { g.Fill(Cylinder{Center: q, Radius: 0.5, Y0: y + 1.75, Y1: y + 2.4}) })
		}
	}
	g.Fill(Func{box, func(p Vec3) bool {
		q := p.Plan()
		return p.Y >= y+1.0 && p.Y <= y+1.6 && pl.in(q, 0) && !pl.in(q, -0.7) && !pl.nearGap(q, 1.5)
	}})
	// Diagonal braces under the deck.
	if !pl.stair {
		for k := 0; k < 7; k++ {
			a := float64(k)*2*math.Pi/7 + float64(t.seed)*0.7
			skip := false
			for _, kk := range t.keep {
				if tv.angDiff(a, kk) < 0.35 {
					skip = true
				}
			}
			if skip {
				continue
			}
			g.Fill(Capsule{Polar(t.c, t.rad(y-4.5)-0.4, a).At(y - 4.5), Polar(t.c, pl.R-1.0, a).At(y - 0.9), 0.55})
		}
	}
	for _, b := range pl.bays {
		a := b.Sub(t.c).Angle()
		g.Fill(Capsule{Polar(t.c, t.rad(y-5)-0.4, a).At(y - 5), b.At(y - 0.9), 0.6})
		for _, s := range []float64{-1, 1} {
			q := Polar(b, 2.6, a+s*1.2)
			g.Fill(Capsule{Polar(t.c, t.rad(y-4)-0.4, a+s*0.25).At(y - 4), q.At(y - 0.9), 0.5})
		}
	}
}

// ladder runs up the trunk through a deck hatch.
func (tv TreehouseVillage) ladder(g *Grid, t *treehouseVillageTree, a, y0, y1 float64) {
	tan := P(-math.Sin(a), math.Cos(a))
	b := Polar(t.c, t.rad(y0+1)+0.9, a)
	e := Polar(t.c, t.rad(y1)+0.9, a)
	g.Use(Wood)
	for _, s := range []float64{-0.7, 0.7} {
		g.Fill(Capsule{b.Add(tan.Scale(s)).At(y0 - 0.4), e.Add(tan.Scale(s)).At(y1 + 1.3), 0.42})
	}
	g.Use(Sand)
	for y := y0 + 0.9; y < y1+0.2; y += 1.15 {
		f := (y - y0) / (y1 - y0)
		c := b.Add(e.Sub(b).Scale(f))
		g.Fill(Capsule{c.Add(tan.Scale(-0.7)).At(y), c.Add(tan.Scale(0.7)).At(y), 0.4})
	}
}

// hut is a round hut on a bay with a conical thatched or shingled roof.
func (tv TreehouseVillage) hut(g *Grid, pl *treehouseVillagePlatform, c Vec2, h treehouseVillageHut) {
	const hr, wall = 2.6, 3.9
	y := pl.y
	box := Box{V(c.X-5, y-1, c.Y-5), V(c.X+5, y+10, c.Y+5)}
	g.Use(h.wall).Fill(Cylinder{Center: c, Radius: hr, Inner: hr - 0.85, Y0: y, Y1: y + wall})
	// Timber frame.
	g.Paint(Wood, Func{box, func(p Vec3) bool {
		q := p.Plan().Sub(c)
		return q.Length() > hr-0.4 && (p.Y > y+wall-0.45 || math.Mod(q.Angle()+7, 2*math.Pi/6) < 0.22)
	}})
	out := c.Sub(pl.tree.c).Angle()
	doorA := out + h.door*math.Pi/2
	g.Carve(RotateY(Arch{Sill: c.Add(P(hr-0.4, 0)).At(y + 0.1), HalfWidth: 0.8, Spring: 1.3, Depth: 2, Through: AxisX}, c, doorA))
	for _, wa := range []float64{out, out - h.door*0.9} {
		g.Carve(RotateY(Arch{Sill: c.Add(P(hr-0.4, 0)).At(y + 1.3), HalfWidth: 0.55, Spring: 0.6, Depth: 2, Through: AxisX}, c, wa))
	}
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		d := p.Plan().Sub(c).Length()
		return p.Y < y+wall && p.Y > y-0.5 && d < hr-0.45
	}})
	// Lantern beside the door.
	lp := Polar(c, hr+0.3, doorA+0.55)
	g.With(Gold, func() { g.Fill(Cylinder{Center: lp, Radius: 0.5, Y0: y + 2.0, Y1: y + 2.7}) })
	// Roof: hollow cone with an overhanging eave.
	rr, rh := hr+0.9, 4.4
	g.Use(h.roof).FillVolume(Subtract(Frustum{c, rr, 0, y + wall - 0.3, y + wall - 0.3 + rh}, Frustum{c, rr, 0, y + wall - 1.6, y + wall - 1.6 + rh}))
	band := Wood
	if h.roof == Terracotta {
		band = Brick
	}
	g.Paint(band, Func{box, func(p Vec3) bool {
		f := (p.Y - (y + wall - 0.3)) / 1.5
		return p.Y > y+wall-0.6 && f-math.Floor(f) < 0.24 && p.Plan().Sub(c).Length() > 0.6
	}})
	g.With(Gold, func() { g.Fill(Sphere(c.At(y+wall-0.3+rh+0.3), 0.55)) })
}

// bridge is a sagging plank rope bridge with rope handrails, hangers and end posts.
func (tv TreehouseVillage) bridge(g *Grid, p1, p2 Vec3) {
	d := p2.Plan().Sub(p1.Plan())
	L := d.Length()
	u := d.Scale(1 / L)
	n := P(-u.Y, u.X)
	sag := 0.06*L + 0.6
	deck := func(t float64) float64 { return p1.Y + (p2.Y-p1.Y)*t - sag*4*t*(1-t) }
	rope := func(t float64) float64 { return p1.Y + (p2.Y-p1.Y)*t + 1.5 - sag*0.75*4*t*(1-t) }
	box := BoxOf(p1, p2).Grow(2.5)
	box.Min.Y -= sag + 1
	local := func(p Vec3) (s, a float64) {
		r := p.Plan().Sub(p1.Plan())
		return r.Dot(u), r.Dot(n)
	}
	g.Use(Wood).Fill(Func{box, func(p Vec3) bool {
		s, a := local(p)
		if s < -0.8 || s > L+0.8 || math.Abs(a) > 1.05 {
			return false
		}
		t := math.Max(0, math.Min(1, s/L))
		y := deck(t)
		return p.Y <= y && p.Y >= y-0.7
	}})
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		s, a := local(p)
		return s > 0.5 && s < L-0.5 && math.Abs(a) <= 1.05 && math.Mod(s, 1.4) < 0.7
	}})
	nh := int(math.Round(L / 2.2))
	g.Use(Sand).Fill(Func{box, func(p Vec3) bool {
		s, a := local(p)
		if s < 0 || s > L {
			return false
		}
		t := s / L
		for _, side := range []float64{-1.15, 1.15} {
			da := a - side
			if math.Hypot(da, p.Y-rope(t)) <= 0.45 {
				return true
			}
			if math.Abs(da) < 0.4 && p.Y > deck(t) && p.Y < rope(t) {
				f := t * float64(nh)
				if math.Abs(f-math.Round(f))*L/float64(nh) < 0.35 {
					return true
				}
			}
		}
		return false
	}})
	// End posts with lanterns.
	for _, e := range []Vec3{p1, p2} {
		for _, side := range []float64{-1.15, 1.15} {
			q := e.Plan().Add(n.Scale(side))
			g.With(Wood, func() { g.Fill(Cylinder{Center: q, Radius: 0.45, Y0: e.Y - 0.6, Y1: e.Y + 2.0}) })
			g.With(Gold, func() { g.Fill(Cylinder{Center: q, Radius: 0.5, Y0: e.Y + 2.2, Y1: e.Y + 2.8}) })
		}
	}
}

// lift is a pulley beam with a hanging basket off one of the autumn tree's decks.
func (tv TreehouseVillage) lift(g *Grid, t *treehouseVillageTree, pl *treehouseVillagePlatform, m *treehouseVillageMap) {
	a := treehouseVillageLiftAngle
	y := pl.y
	beamY := y + 4.0
	tip := Polar(t.c, pl.R+2.8, a)
	g.Use(Wood)
	g.Fill(Capsule{Polar(t.c, t.rad(beamY)-0.4, a).At(beamY), tip.At(beamY), 0.6})
	g.Fill(Capsule{Polar(t.c, pl.R-0.4, a).At(y), Polar(t.c, pl.R+1.4, a).At(beamY - 0.2), 0.5})
	// Pulley wheel in the radial plane.
	wc := Polar(t.c, pl.R+2.8, a).At(beamY - 1.0)
	dir := P(math.Cos(a), math.Sin(a))
	tan := P(-dir.Y, dir.X)
	g.With(Charcoal, func() {
		g.Fill(Func{Box{wc.Sub(V(1.4, 1.4, 1.4)), wc.Add(V(1.4, 1.4, 1.4))}, func(p Vec3) bool {
			r := p.Sub(wc)
			along := r.Plan().Dot(tan)
			rad := r.Plan().Dot(dir)
			return math.Abs(along) <= 0.35 && math.Hypot(rad, r.Y) <= 1.0
		}})
	})
	g.With(Gold, func() { g.Fill(Sphere(wc, 0.45)) })
	// Rope down to the basket.
	rx := Polar(t.c, pl.R+3.8, a)
	ground := m.at(rx.X, rx.Y)
	bY := ground + (y-ground)*0.45
	g.With(Sand, func() { g.Fill(Capsule{rx.At(beamY - 1.0), rx.At(bY + 2.4), 0.42}) })
	bc := rx
	g.With(Wood, func() {
		g.Fill(Subtract(BlockAt(bc, 3.2, 3.2, bY, bY+2.2), BlockAt(bc, 1.8, 1.8, bY+0.6, bY+3)))
		for _, s := range []float64{-1, 1} {
			g.Fill(Capsule{bc.Add(tan.Scale(s * 1.2)).At(bY + 1.6), bc.At(bY + 2.6), 0.4})
		}
	})
	g.Paint(Gold, BlockAt(bc, 3.4, 3.4, bY+1.7, bY+2.3))
	// A second rope to a cleat on the deck.
	g.With(Sand, func() {
		g.Fill(Capsule{Polar(t.c, pl.R+1.8, a).At(beamY - 1.0), Polar(t.c, pl.R-0.6, a+0.12).At(y + 1.0), 0.42})
	})
}

// crown is the branching crown: main limbs and layered foliage clumps with gaps.
func (tv TreehouseVillage) crown(g *Grid, t *treehouseVillageTree) {
	s := t.r / 4.6
	nb := 6
	var clumps []Ellipsoid
	g.Use(Wood)
	g.Fill(Capsule{t.c.At(t.top), t.c.At(t.top + 9*s), 1.6 * s})
	for k := 0; k < nb; k++ {
		a := float64(k)*2*math.Pi/float64(nb) + Hash01(t.seed, k, 40)*0.5 + float64(t.seed)
		mid := Polar(t.c, 4.5*s, a).At(t.top + 2 + 1.5*Hash01(t.seed, k, 41))
		end := Polar(t.c, (9+2*Hash01(t.seed, k, 42))*s, a+0.2).At(t.top + 4.5 + 2.5*Hash01(t.seed, k, 43))
		g.Fill(Capsule{t.c.At(t.top - 1.5), mid, 1.25 * s}, Capsule{mid, end, 0.9 * s})
		clumps = append(clumps, Ellipsoid{end.Add(V(0, 1.6, 0)), V(4.8*s, 2.9*s, 4.8*s)})
	}
	for k := 0; k < 4; k++ {
		a := float64(k)*2*math.Pi/4 + Hash01(t.seed, k, 44) + float64(t.seed)*1.3
		q := Polar(t.c, 4.5*s, a)
		clumps = append(clumps, Ellipsoid{q.At(t.top + 9*s + Hash01(t.seed, k, 45)), V(4.6*s, 2.8*s, 4.6*s)})
	}
	clumps = append(clumps, Ellipsoid{t.c.At(t.top + 12.5*s), V(3.8*s, 2.8*s, 3.8*s)})
	box := EmptyBox()
	for _, c := range clumps {
		box = box.Union(c.Bounds())
	}
	inside := func(p Vec3, grow float64) bool {
		for _, c := range clumps {
			d := p.Sub(c.Center)
			r := c.Radii.Add(V(grow, grow, grow))
			x, y, z := d.X/r.X, d.Y/r.Y, d.Z/r.Z
			if x*x+y*y+z*z <= 1 {
				return true
			}
		}
		return false
	}
	base, patch, dot := Foliage, Grass, Grass
	switch t.leaf {
	case 1:
		base, patch, dot = Foliage, Gold, Grass
	case 2:
		base, patch, dot = Gold, Terracotta, Red
	}
	g.Use(base).FillVolume(Func{box, func(p Vec3) bool {
		// Lumpy surface.
		bump := 0.5 * (tv.noise(p.X/1.6+p.Y/2.3, p.Z/1.6, t.seed+70) - 0.5)
		return inside(p, bump) && !inside(p, bump-1.2)
	}})
	g.Paint(patch, Func{box, func(p Vec3) bool {
		return tv.noise(p.X/2.1+p.Y/3, p.Z/2.1, t.seed+60) > 0.6
	}})
	g.Paint(dot, Func{box, func(p Vec3) bool {
		return Hash01(int(math.Round(p.X*2)), int(math.Round(p.Y*2)), int(math.Round(p.Z*2)), t.seed) < 0.07
	}})
}

// mill is a small round mill hut on the stream bank with an undershot waterwheel.
func (tv TreehouseVillage) mill(g *Grid, m *treehouseVillageMap) {
	w := treehouseVillageWheel
	flow, ax := tv.millAxes()
	hc := tv.millHut()
	hy := m.at(hc.X, hc.Y)
	g.Use(Stone).Fill(Cylinder{Center: hc, Radius: 3.6, Y0: hy - 1.5, Y1: hy + 0.6})
	g.Use(White).Fill(Cylinder{Center: hc, Radius: 3.2, Inner: 2.3, Y0: hy + 0.6, Y1: hy + 4.6})
	g.Paint(Wood, Func{Box{V(hc.X-4, hy, hc.Y-4), V(hc.X+4, hy+5, hc.Y+4)}, func(p Vec3) bool {
		q := p.Plan().Sub(hc)
		return p.Y > hy+0.7 && (p.Y > hy+4.1 || p.Y < hy+1.1 || math.Mod(q.Angle()+7, 2*math.Pi/8) < 0.2)
	}})
	g.Carve(RotateY(Arch{Sill: hc.Add(P(3, 0)).At(hy + 0.7), HalfWidth: 0.8, Spring: 1.5, Depth: 2, Through: AxisX}, hc, ax.Scale(-1).Angle()+math.Pi/2+0.4))
	g.Carve(RotateY(Arch{Sill: hc.Add(P(3, 0)).At(hy + 2.2), HalfWidth: 0.6, Spring: 0.6, Depth: 2, Through: AxisX}, hc, ax.Angle()+math.Pi+1.2))
	g.Paint(Charcoal, Cylinder{Center: hc, Radius: 2.75, Y0: hy - 0.2, Y1: hy + 4.6})
	g.Use(Terracotta).FillVolume(Subtract(Frustum{hc, 4.3, 0, hy + 4.3, hy + 9.3}, Frustum{hc, 4.3, 0, hy + 3.0, hy + 8.0}))
	g.Paint(Brick, Func{Box{V(hc.X-5, hy+4, hc.Y-5), V(hc.X+5, hy+10, hc.Y+5)}, func(p Vec3) bool {
		f := (p.Y - hy - 4.3) / 1.25
		return f-math.Floor(f) < 0.3
	}})
	g.With(Gold, func() { g.Fill(Sphere(hc.At(hy+9.5), 0.55)) })
	// Wheel: rim, spokes, paddles; plane contains the flow direction.
	const wr = 3.8
	wy := treehouseVillageWater + 2.3
	wc := w.Add(ax.Scale(0.6)).At(wy)
	box := Box{wc.Sub(V(wr+1, wr+1, wr+1)), wc.Add(V(wr+1, wr+1, wr+1))}
	g.Use(Wood).Fill(Func{box, func(p Vec3) bool {
		r := p.Sub(wc)
		across := r.Plan().Dot(ax)
		along := r.Plan().Dot(flow)
		rad := math.Hypot(along, r.Y)
		ang := math.Atan2(r.Y, along)
		if math.Abs(across) > 0.9 {
			return false
		}
		if rad <= wr && rad >= wr-0.75 {
			return true // Rim.
		}
		if rad < wr {
			// Spokes.
			f := math.Mod(ang+8*math.Pi, math.Pi/3)
			return rad < 0.9 || math.Min(f, math.Pi/3-f)*rad < 0.3
		}
		if rad < wr+0.9 {
			f := math.Mod(ang+8*math.Pi+0.2, math.Pi/6)
			return math.Min(f, math.Pi/6-f)*rad < 0.45
		}
		return false
	}})
	g.With(Charcoal, func() {
		g.Fill(Capsule{wc, hc.Add(ax.Scale(-2.6)).At(wy), 0.5})
	})
	// Leat posts beside the wheel.
	for _, s := range []float64{-1, 1} {
		q := w.Add(ax.Scale(-1.5)).Add(flow.Scale(s * 1.4))
		g.With(Wood, func() { g.Fill(Cylinder{Center: q, Radius: 0.5, Y0: 0, Y1: wy + 0.4}) })
	}
}

// millAxes returns the stream flow direction at the wheel and the axle direction,
// which points from the wheel toward the mill on the clearing side.
func (TreehouseVillage) millAxes() (flow, ax Vec2) {
	flow = P(-11, 7.5)
	flow = flow.Scale(1 / flow.Length())
	ax = P(-flow.Y, flow.X)
	if ax.Dot(treehouseVillageClearing.Sub(treehouseVillageWheel)) < 0 {
		ax = ax.Scale(-1)
	}
	return flow, ax
}

func (tv TreehouseVillage) millHut() Vec2 {
	_, ax := tv.millAxes()
	return treehouseVillageWheel.Add(ax.Scale(6.6))
}

// clearing is the campfire ring with log benches and lantern posts.
func (tv TreehouseVillage) clearing(g *Grid, m *treehouseVillageMap) {
	f := treehouseVillageFire
	y := m.at(f.X, f.Y)
	for k := 0; k < 10; k++ {
		a := float64(k) * 2 * math.Pi / 10
		c := DarkStone
		if k%2 == 0 {
			c = Stone
		}
		g.With(c, func() { g.FillVolume(Ellipsoid{Polar(f, 2.2, a).At(y + 0.2), V(0.75, 0.6, 0.75)}) })
	}
	g.Paint(Charcoal, Cylinder{Center: f, Radius: 1.6, Y0: y - 0.6, Y1: y + 0.1})
	g.Use(Wood)
	for k := 0; k < 3; k++ {
		a := float64(k) * math.Pi / 3
		g.Fill(Capsule{Polar(f, 1.3, a).At(y + 0.4), Polar(f, 1.3, a+math.Pi).At(y + 0.4), 0.45})
	}
	g.With(Red, func() { g.FillVolume(Frustum{f, 1.1, 0, y + 0.5, y + 2.7}) })
	g.With(Terracotta, func() { g.Fill(Frustum{f, 0.6, 0, y + 1.2, y + 3.3}) })
	g.With(Gold, func() { g.Fill(Frustum{f, 0.5, 0, y + 0.6, y + 2.2}) })
	// Log benches.
	for k := 0; k < 4; k++ {
		a := float64(k)*math.Pi/2 + 0.5
		c := Polar(f, 5.4, a)
		tan := P(-math.Sin(a), math.Cos(a))
		by := m.at(c.X, c.Y)
		g.With(Wood, func() { g.Fill(Capsule{c.Add(tan.Scale(-1.8)).At(by + 0.5), c.Add(tan.Scale(1.8)).At(by + 0.5), 0.65}) })
	}
	// Lantern posts along the path to the stones and the mill.
	for _, q := range []Vec2{P(4.8, 11.5), P(7.8, 17.5), P(5.6, 24), P(10.5, 11.8)} {
		h := m.at(q.X, q.Y)
		g.With(Wood, func() {
			g.Fill(Cylinder{Center: q, Radius: 0.45, Y0: h - 0.5, Y1: h + 2.9})
		})
		g.With(Gold, func() { g.Fill(Cylinder{Center: q, Radius: 0.55, Y0: h + 3.1, Y1: h + 3.9}) })
		g.With(Wood, func() { g.Fill(Cylinder{Center: q, Radius: 0.75, Y0: h + 4.0, Y1: h + 4.4}) })
	}
}

// floor scatters stepping stones, mushrooms, ferns and bushes.
func (tv TreehouseVillage) floor(g *Grid, m *treehouseVillageMap, trees []*treehouseVillageTree) {
	// Stepping stones across the stream.
	dir := P(0.6, 0.8)
	for k := -3; k <= 3; k++ {
		q := treehouseVillageStones.Add(dir.Scale(float64(k) * 1.75)).Add(P(Hash01(k, 80)-0.5, Hash01(k, 81)-0.5).Scale(0.6))
		c := Stone
		if k%2 != 0 {
			c = DarkStone
		}
		g.With(c, func() { g.FillVolume(EllipticCylinder{q, 0.95, 0.85, -0.5, treehouseVillageWater + 0.6}) })
	}
	free := func(q Vec2, clear float64) bool {
		h := m.at(q.X, q.Y)
		if h < 0.6 || tv.streamDist(q) < 1.5 {
			return false
		}
		if q.Sub(treehouseVillageClearing).Length() < 10.5 {
			return false
		}
		if q.Sub(treehouseVillageWheel).Length() < 11 {
			return false
		}
		for _, t := range trees {
			d := q.Sub(t.c).Length()
			if d < t.rad(t.g0)+clear {
				return false
			}
			// Keep the door paths clear.
			if d < 16 && tv.angDiff(q.Sub(t.c).Angle(), t.door) < 0.35 {
				return false
			}
		}
		return true
	}
	mushroom := func(q Vec2, size float64, seed int) {
		h := m.at(q.X, q.Y)
		g.With(White, func() {
			g.Fill(Cylinder{Center: q, Radius: 0.45 + 0.22*size, Y0: h - 0.5, Y1: h + 1.6*size})
		})
		cap := Ellipsoid{q.At(h + 1.6*size), V(1.1*size+0.4, 0.75*size+0.3, 1.1*size+0.4)}
		g.With(Red, func() {
			g.FillVolume(Intersect(cap, Cuboid{V(q.X-9, h+1.6*size-0.25, q.Y-9), V(q.X+9, h+20, q.Y+9)}))
		})
		g.Paint(White, Func{cap.Bounds(), func(p Vec3) bool {
			return cap.Contains(p) && Hash01(int(math.Round(p.X*1.6)), int(math.Round(p.Y*1.6)), int(math.Round(p.Z*1.6)), seed) < 0.22
		}})
	}
	// Giant mushrooms clustered at tree roots.
	for i, t := range trees {
		for k := 0; k < 3; k++ {
			a := t.door + 1.2 + float64(k)*1.5 + Hash01(i, k, 82)*0.6
			q := Polar(t.c, t.rad(t.g0)+5+3*Hash01(i, k, 83), a)
			if !free(q, 3.5) {
				continue
			}
			mushroom(q, 1.4+1.4*Hash01(i, k, 84), 90+i*7+k)
			for j := 0; j < 2; j++ {
				q2 := Polar(q, 2.4+Hash01(i, k, j, 85), Hash01(i, k, j, 86)*6.3)
				if free(q2, 3.5) {
					mushroom(q2, 0.5+0.4*Hash01(i, k, j, 87), 95+i*7+k+j)
				}
			}
		}
	}
	// Ferns and bushes on a jittered grid.
	fern := func(q Vec2, seed int) {
		h := m.at(q.X, q.Y)
		c := Foliage
		if Hash01(seed, 1) < 0.4 {
			c = Grass
		}
		g.Use(c)
		nf := 6
		for k := 0; k < nf; k++ {
			a := float64(k)*2*math.Pi/float64(nf) + Hash01(seed, 2)
			mid := Polar(q, 1.2, a).At(h + 1.3)
			end := Polar(q, 2.4, a+0.15).At(h + 0.6)
			g.Fill(Capsule{q.At(h), mid, 0.45}, Capsule{mid, end, 0.42})
		}
	}
	for gx := -48.0; gx <= 48; gx += 4.2 {
		for gz := -46.0; gz <= 50; gz += 4.2 {
			ix, iz := int(gx*10), int(gz*10)
			q := P(gx+(Hash01(ix, iz, 88)-0.5)*3, gz+(Hash01(ix, iz, 89)-0.5)*3)
			if !free(q, 2.5) || tv.siteE(q.X, q.Y) > 0.9 {
				continue
			}
			r := Hash01(ix, iz, 90)
			n := tv.fbm(q.X/9, q.Y/9, 91)
			switch {
			case r < 0.28+0.3*n:
				fern(q, ix*7+iz)
			case r < 0.36+0.3*n:
				h := m.at(q.X, q.Y)
				g.With(Foliage, func() { g.FillVolume(Ellipsoid{q.At(h + 0.6), V(1.6, 1.3, 1.6)}) })
				g.Paint(Grass, Ellipsoid{q.At(h + 1.2), V(1.4, 0.8, 1.4)})
			case r < 0.41:
				mushroom(q, 0.45, ix+iz)
			case r > 0.95:
				h := m.at(q.X, q.Y)
				c := DarkStone
				if Hash01(ix, iz, 92) < 0.5 {
					c = Stone
				}
				rock := Ellipsoid{q.At(h), V(1.3+Hash01(ix, iz, 93), 1.0+0.6*Hash01(ix, iz, 94), 1.2+Hash01(ix, iz, 95))}
				g.With(c, func() { g.FillVolume(rock) })
				g.Paint(Foliage, Func{rock.Bounds(), func(p Vec3) bool { return p.Y > h+0.5*rock.Radii.Y && Hash01(int(p.X*3), int(p.Z*3), 96) < 0.5 }})
			}
		}
	}
}
