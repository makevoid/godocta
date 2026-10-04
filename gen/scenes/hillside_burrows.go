package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// HillsideBurrows is a cozy hillside village of homes dug into rolling green hills: round
// painted doors and windows in turf banks, chimneys poking out of the grass, fenced front
// gardens, vegetable plots, winding sand lanes with stone steps, a great oak above the
// largest home, a mill pond with a water wheel, a double-arched bridge and a party field.
type HillsideBurrows struct{}

func init() { Register(HillsideBurrows{}) }

func (HillsideBurrows) Name() string  { return "hillside_burrows" }
func (HillsideBurrows) Title() string { return "Hillside Burrows" }

// Site plan: the camera's front is +Z, so the homes face mostly +Z.
var (
	hillsideBurrowsSite   = P(0, 4)
	hillsideBurrowsSiteA  = 52.0
	hillsideBurrowsSiteB  = 44.0
	hillsideBurrowsPond   = P(30, 15)
	hillsideBurrowsParty  = P(-27, 29)
	hillsideBurrowsBridge = P(14, 26)
	hillsideBurrowsOak    = P(-1, -13.5)
	hillsideBurrowsLane   = []Vec2{P(-56, 11), P(-40, 11.5), P(-27, 16.5), P(-12, 19), P(4, 22.5), P(14, 26), P(24, 29.5), P(34, 27), P(41, 21.5)}
	hillsideBurrowsStream = []Vec2{P(23, 19.5), P(19, 22), P(14, 26), P(10.5, 33), P(8, 40), P(6, 52)}
)

const (
	hillsideBurrowsWater = 1.0 // Pond and stream surface.
	hillsideBurrowsField = 1.9 // Party field level.
)

type hillsideBurrowsHill struct {
	c    Vec2
	r, h float64
	seed int
}

var hillsideBurrowsHills = []hillsideBurrowsHill{
	{P(-2, -16), 28, 19, 21}, // The great hill.
	{P(-31, -8), 17, 12, 22},
	{P(27, -14), 18, 13, 23},
	{P(14, 4), 12, 9, 24},
	{P(-15, 6), 11, 8.5, 25},
	{P(-43, -22), 14, 9, 26},
	{P(41, -25), 13, 8, 27},
	{P(-4, 31), 10, 7.5, 28},
}

// hillsideBurrowsHome is one home dug into a hill: a round door in a vertical turf face,
// with a flat front garden in front of it along the facing direction f.
type hillsideBurrowsHome struct {
	door, f    Vec2    // Door point on the face and outward unit facing.
	y0         float64 // Garden level and threshold.
	rd         float64 // Door radius.
	w, d       float64 // Garden half-width and depth.
	gate       float64 // Gate offset across the front.
	door1      Color
	frame      Color
	fence      Color
	big        bool
	seed       int
	faceHeight float64
}

func (h *hillsideBurrowsHome) local(p Vec2) (u, v float64) {
	r := p.Sub(h.door)
	return r.Dot(h.f), r.X*-h.f.Y + r.Y*h.f.X
}

func (h *hillsideBurrowsHome) world(u, v float64) Vec2 {
	return h.door.Add(h.f.Scale(u)).Add(P(-h.f.Y, h.f.X).Scale(v))
}

// sd is the signed distance to the garden's rounded rectangle; the garden is its u ≥ 0 part.
func (h *hillsideBurrowsHome) sd(u, v float64) float64 {
	const rc = 2.0
	// The rectangle reaches behind the face so its back corners stay inside the bank.
	qx := math.Abs(u-(h.d-rc)/2) - ((h.d+rc)/2 - rc)
	qy := math.Abs(v) - (h.w - rc)
	out := math.Hypot(math.Max(qx, 0), math.Max(qy, 0))
	return out + math.Min(math.Max(qx, qy), 0) - rc
}

type hillsideBurrowsMap struct {
	x0, z0, step float64
	nx, nz       int
	h            []float64
	stair        []bool
}

func (m *hillsideBurrowsMap) at(x, z float64) float64 {
	fx, fz := (x-m.x0)/m.step, (z-m.z0)/m.step
	if fx < 0 || fz < 0 || fx >= float64(m.nx-1) || fz >= float64(m.nz-1) {
		return 0
	}
	i, j := int(fx), int(fz)
	tx, tz := fx-float64(i), fz-float64(j)
	k := j*m.nx + i
	a := m.h[k] + (m.h[k+1]-m.h[k])*tx
	b := m.h[k+m.nx] + (m.h[k+m.nx+1]-m.h[k+m.nx])*tx
	return a + (b-a)*tz
}

func (m *hillsideBurrowsMap) stairAt(x, z float64) bool {
	i, j := int(math.Round((x-m.x0)/m.step)), int(math.Round((z-m.z0)/m.step))
	if i < 0 || j < 0 || i >= m.nx || j >= m.nz {
		return false
	}
	return m.stair[j*m.nx+i]
}

// low is the lowest terrain around a point, used to close banks and cuts into shells.
func (m *hillsideBurrowsMap) low(x, z float64) float64 {
	lo := m.at(x, z)
	for k := 0; k < 8; k++ {
		a := float64(k) * math.Pi / 4
		lo = math.Min(lo, m.at(x+1.1*math.Cos(a), z+1.1*math.Sin(a)))
	}
	return lo
}

func (HillsideBurrows) noise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

func (hv HillsideBurrows) fbm(x, z float64, seed int) float64 {
	return (hv.noise(x, z, seed)*4 + hv.noise(x*2.1, z*2.1, seed+1)*2 + hv.noise(x*4.3, z*4.3, seed+2)) / 7
}

// segDist is the plan distance from p to a polyline.
func (HillsideBurrows) segDist(line []Vec2, p Vec2) float64 {
	best := math.Inf(1)
	for i := 0; i+1 < len(line); i++ {
		a, b := line[i], line[i+1]
		ab := b.Sub(a)
		t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
		best = math.Min(best, p.Sub(a.Add(ab.Scale(t))).Length())
	}
	return best
}

func (hv HillsideBurrows) siteE(x, z float64) float64 {
	ex, ez := (x-hillsideBurrowsSite.X)/hillsideBurrowsSiteA, (z-hillsideBurrowsSite.Y)/hillsideBurrowsSiteB
	return math.Sqrt(ex*ex+ez*ez) + 0.05*(hv.fbm(x/9, z/9, 3)-0.5)
}

// base is the meadow and hills before ponds, streams, homes and paths.
func (hv HillsideBurrows) base(x, z float64) float64 {
	h := 1.9 + 2.4*(hv.fbm(x/13, z/13, 1)-0.5)
	for _, hl := range hillsideBurrowsHills {
		r := math.Hypot(x-hl.c.X, z-hl.c.Y) / hl.r
		r *= 1 + 0.16*(hv.fbm(x/8, z/8, hl.seed)-0.5)
		if r < 1 {
			h += hl.h * math.Pow(0.5+0.5*math.Cos(math.Pi*r), 0.75)
		}
	}
	return h
}

// homes places the homes on the hills: each door sits where the hill rises a face
// height above its garden.
func (hv HillsideBurrows) homes() []*hillsideBurrowsHome {
	type spec struct {
		hill       int
		angle, r   float64
		door       Color
		frame      Color
		fence      Color
		gate, w, d float64
		big        bool
	}
	specs := []spec{
		{0, 90, 10, Grass, Wood, White, 2.5, 8, 8, true},
		{0, 42, 17, Red, White, Wood, 0, 5.4, 5, false},
		{0, 138, 17, Navy, White, Wood, 0, 5.4, 5, false},
		{1, 100, 8, Red, Wood, Wood, 1.2, 5.4, 5, false},
		{1, 25, 8, Wood, White, Wood, 0, 5.4, 5, false},
		{2, 80, 9, Navy, Wood, Wood, 1, 5.4, 5, false},
		{2, 155, 8, Red, White, Wood, 0, 5.4, 5, false},
		{2, 10, 9.5, Grass, White, Wood, 0, 5.4, 5, false},
		{3, 95, 5.5, Gold, Wood, Wood, -1, 5.4, 5, false},
		{3, -10, 5.5, Red, White, Wood, 0, 5.4, 5, false},
		{4, 85, 5, Navy, White, Wood, 1, 5.4, 5, false},
		{7, 95, 5, Gold, Wood, White, 1.5, 5.4, 4.5, false},
		{7, 185, 5, Red, White, Wood, 0, 5.4, 4.5, false},
	}
	var homes []*hillsideBurrowsHome
	for i, s := range specs {
		hl := hillsideBurrowsHills[s.hill]
		a := s.angle * math.Pi / 180
		f := P(math.Cos(a), math.Sin(a))
		door := hl.c.Add(f.Scale(s.r))
		home := &hillsideBurrowsHome{door: door, f: f, rd: 1.8, w: s.w, d: s.d, gate: s.gate,
			door1: s.door, frame: s.frame, fence: s.fence, big: s.big, seed: 100 + i, faceHeight: 5.6}
		if s.big {
			home.rd, home.faceHeight = 2.3, 6.9
		}
		home.y0 = math.Max(hv.base(door.X, door.Y)-home.faceHeight, 1.9)
		homes = append(homes, home)
	}
	return homes
}

// branches are footpaths from each garden gate down to the lane.
func (hv HillsideBurrows) branches(homes []*hillsideBurrowsHome) [][]Vec2 {
	var paths [][]Vec2
	for _, h := range homes {
		gate := h.world(h.d+0.4, h.gate)
		out := h.world(h.d+3, h.gate)
		// Nearest lane point.
		best, bp := math.Inf(1), gate
		for i := 0; i+1 < len(hillsideBurrowsLane); i++ {
			a, b := hillsideBurrowsLane[i], hillsideBurrowsLane[i+1]
			ab := b.Sub(a)
			t := math.Max(0, math.Min(1, out.Sub(a).Dot(ab)/ab.Dot(ab)))
			q := a.Add(ab.Scale(t))
			if d := out.Sub(q).Length(); d < best {
				best, bp = d, q
			}
		}
		mid := out.Add(bp).Scale(0.5)
		dir := bp.Sub(out)
		if dir.Length() > 1e-6 {
			n := P(-dir.Y, dir.X).Scale(1 / dir.Length())
			mid = mid.Add(n.Scale((Hash01(h.seed, 7) - 0.5) * 0.35 * dir.Length()))
		}
		paths = append(paths, []Vec2{gate, out, mid, bp})
	}
	return paths
}

// height is the final terrain: meadow, hills, rim, pond, stream, party field and gardens.
func (hv HillsideBurrows) height(x, z float64, homes []*hillsideBurrowsHome) float64 {
	e := hv.siteE(x, z)
	if e > 1 {
		return 0
	}
	h := hv.base(x, z)
	p := P(x, z)
	// Party field, flattened.
	if d := p.Sub(hillsideBurrowsParty).Length(); d < 14 {
		w := math.Min(1, math.Max(0, (14-d)/5))
		w = w * w * (3 - 2*w)
		h += (hillsideBurrowsField - h) * w
	}
	// Homes: raise a bank behind each face, then cut the flat garden.
	for _, hm := range homes {
		u, v := hm.local(p)
		if u < 1 && u > -6 && math.Abs(v) < hm.w+3 {
			k := math.Max(0, -u-2)*0.9 + math.Max(0, math.Abs(v)-(hm.w-1))*0.9
			h = math.Max(h, hm.y0+hm.faceHeight+0.4-k)
		}
	}
	for _, hm := range homes {
		u, v := hm.local(p)
		if u >= 0 && hm.sd(u, v) <= 0 {
			h = hm.y0
		}
	}
	// Everything slopes down to a low rim.
	h = math.Min(h, 0.5+12*(1-e)+math.Max(0, 1-e-0.12)*60)
	// Mill pond and race.
	pe := math.Hypot((x-hillsideBurrowsPond.X)/8.5, (z-hillsideBurrowsPond.Y)/6.5)
	pe += 0.08 * (hv.fbm(x/5, z/5, 9) - 0.5)
	h = math.Min(h, hillsideBurrowsWater+math.Max(0, pe-1)*9)
	if x > 33 && x < 39.6 && z > 11.5 && z < 16.5 {
		h = hillsideBurrowsWater
	}
	// Stream from the pond to the rim.
	ds := hv.segDist(hillsideBurrowsStream, p) - 2.2 - 0.6*(hv.fbm(x/5, z/5, 10)-0.5)
	h = math.Min(h, hillsideBurrowsWater+math.Max(0, ds)*1.1)
	return math.Max(h, 0.4)
}

func (hv HillsideBurrows) wet(x, z float64) bool {
	pe := math.Hypot((x-hillsideBurrowsPond.X)/8.5, (z-hillsideBurrowsPond.Y)/6.5)
	pe += 0.08 * (hv.fbm(x/5, z/5, 9) - 0.5)
	if pe < 1 || x > 33 && x < 39.6 && z > 11.5 && z < 16.5 {
		return true
	}
	return hv.segDist(hillsideBurrowsStream, P(x, z))-2.2-0.6*(hv.fbm(x/5, z/5, 10)-0.5) < 0
}

func (hv HillsideBurrows) Build(g *Grid) Order {
	homes := hv.homes()
	paths := hv.branches(homes)
	m := &hillsideBurrowsMap{x0: -56, z0: -44, step: 0.4}
	m.nx, m.nz = int(112/m.step)+2, int(100/m.step)+2
	m.h = make([]float64, m.nx*m.nz)
	m.stair = make([]bool, m.nx*m.nz)
	for j := 0; j < m.nz; j++ {
		for i := 0; i < m.nx; i++ {
			m.h[j*m.nx+i] = hv.height(m.x0+float64(i)*m.step, m.z0+float64(j)*m.step, homes)
		}
	}
	// Stone steps where footpaths climb.
	steps := make([]float64, len(m.h))
	copy(steps, m.h)
	for j := 1; j < m.nz-1; j++ {
		for i := 1; i < m.nx-1; i++ {
			k := j*m.nx + i
			x, z := m.x0+float64(i)*m.step, m.z0+float64(j)*m.step
			on := false
			for _, pth := range paths {
				if hv.segDist(pth[1:], P(x, z)) < 1.0 {
					on = true
					break
				}
			}
			if !on || m.h[k] <= 0 {
				continue
			}
			gx := (m.h[k+1] - m.h[k-1]) / (2 * m.step)
			gz := (m.h[k+m.nx] - m.h[k-m.nx]) / (2 * m.step)
			if math.Hypot(gx, gz) > 0.3 {
				steps[k] = math.Round(m.h[k]/0.7) * 0.7
				m.stair[k] = true
			}
		}
	}
	m.h = steps

	hv.terrain(g, m, homes, paths)
	for _, hm := range homes {
		hv.home(g, m, hm)
	}
	hv.oak(g, m)
	hv.bridge(g)
	hv.mill(g, m)
	hv.party(g, m)
	hv.plots(g, m)
	hv.hedges(g, m)
	hv.trees(g, m)
	return Outward(P(-2, -10))
}

// terrain fills the heightfield as a hollow turf shell and paints lanes, steps and water.
func (hv HillsideBurrows) terrain(g *Grid, m *hillsideBurrowsMap, homes []*hillsideBurrowsHome, paths [][]Vec2) {
	box := Box{V(m.x0, -1, m.z0), V(m.x0+112, 24, m.z0+100)}
	g.Use(Grass).Fill(Func{box, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		if h <= 0 {
			return false
		}
		return p.Y <= h && p.Y >= math.Min(h, m.low(p.X, p.Z))-1.0
	}})
	top := func(p Vec3) bool { return p.Y >= m.at(p.X, p.Z)-0.9 }
	// Darker turf on steep banks and in meadow patches.
	g.Paint(Foliage, Func{box, func(p Vec3) bool {
		h := m.at(p.X, p.Z)
		return p.Y < h-1.2 && h-m.low(p.X, p.Z) > 1.2
	}})
	// Lane and footpaths.
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		if !top(p) {
			return false
		}
		q := p.Plan()
		w := 1.3 + 0.25*(hv.fbm(q.X/4, q.Y/4, 30)-0.5)
		if hv.segDist(hillsideBurrowsLane, q) < w {
			return true
		}
		for _, pth := range paths {
			if hv.segDist(pth, q) < 0.85 {
				return true
			}
		}
		return false
	}})
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		return top(p) && m.stairAt(p.X, p.Z)
	}})
	// Water in the pond, race and stream.
	g.Paint(Water, Func{box, func(p Vec3) bool {
		return p.Y <= hillsideBurrowsWater+0.3 && top(p) && hv.wet(p.X, p.Z)
	}})
	// Wildflowers dotted over patches of meadow.
	span := g.Span()
	for ci, c := range []Color{Red, Gold, White} {
		g.With(c, func() {
			g.Fill(Func{box, func(p Vec3) bool {
				y := p.Y - m.at(p.X, p.Z)
				if y < 0.05 || y > 0.85 || hv.noise(p.X/6, p.Z/6, 50) < 0.7 {
					return false
				}
				r := Hash01(int(math.Round(p.X*8)), int(math.Round(p.Z*8)), 51)
				if int(r*20) != ci {
					return false
				}
				below, ok := g.ColorOf(Nearest(p.Sub(V(0, span, 0)).Scale(1 / span)))
				return ok && below == Grass
			}})
		})
	}
	// Reeds along the pond's western margin.
	g.With(Foliage, func() {
		g.Fill(Func{Box{V(18, 0.8, 6), V(33, 3, 24)}, func(p Vec3) bool {
			pe := math.Hypot((p.X-hillsideBurrowsPond.X)/8.5, (p.Z-hillsideBurrowsPond.Y)/6.5)
			if pe < 0.86 || pe > 1.05 || p.Y > hillsideBurrowsWater+1.8 || p.Y < hillsideBurrowsWater+0.3 {
				return false
			}
			return hv.noise(p.X/1.3, p.Z/1.3, 40) > 0.6 && Hash01(int(p.X*3), int(p.Z*3), 41) < 0.8
		}})
	})
}

// home details one dug-in home: retaining wall, garden beds and flowers, picket fence,
// round door with frame and knob, round windows and chimneys rising from the turf.
func (hv HillsideBurrows) home(g *Grid, m *hillsideBurrowsMap, h *hillsideBurrowsHome) {
	y0, rd := h.y0, h.rd
	reach := h.d + h.w + 6
	box := Box{V(h.door.X-reach, y0-4, h.door.Y-reach), V(h.door.X+reach, y0+h.faceHeight+3, h.door.Y+reach)}
	loc := func(p Vec3) (u, v, y float64) {
		u, v = h.local(p.Plan())
		return u, v, p.Y - y0
	}
	inYard := func(u, v float64) bool { return u >= 0 && h.sd(u, v) <= 0 }
	pathD := func(u, v float64) float64 {
		a, b := P(0.3, 0), P(h.d, h.gate)
		ab := b.Sub(a)
		q := P(u, v)
		t := math.Max(0, math.Min(1, q.Sub(a).Dot(ab)/ab.Dot(ab)))
		return q.Sub(a.Add(ab.Scale(t))).Length()
	}
	bed := func(u, v float64) bool {
		sd := h.sd(u, v)
		return u > 1.6 && sd > -2.4 && sd <= -0.9 && pathD(u, v) > 1.4
	}
	// Stone retaining walls where the garden stands above the meadow.
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		u, v, y := loc(p)
		sd := h.sd(u, v)
		return u > 0.3 && sd > -0.5 && sd < 1.5 && y < -0.6
	}})
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		u, v, y := loc(p)
		return inYard(u, v) && y > -0.9 && pathD(u, v) < 0.8
	}})
	g.Paint(Wood, Func{box, func(p Vec3) bool {
		u, v, y := loc(p)
		return inYard(u, v) && y > -0.9 && bed(u, v)
	}})
	for ci, c := range []Color{Red, Gold, White} {
		g.With(c, func() {
			g.Fill(Func{box, func(p Vec3) bool {
				u, v, y := loc(p)
				if y < 0.1 || y > 0.9 || !inYard(u, v) || !bed(u, v) {
					return false
				}
				r := Hash01(int(math.Round(p.X*8)), int(math.Round(p.Y*8)), int(math.Round(p.Z*8)), h.seed)
				return int(r*5) == ci
			}})
		})
	}
	// Picket fence along the open sides of the garden, with a gate gap.
	g.With(h.fence, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			u, v, y := loc(p)
			if u < 0.9 || y <= 0 || y > 1.3 {
				return false
			}
			sd := h.sd(u, v)
			if sd > 0 || sd < -0.85 {
				return false
			}
			if u > h.d-1.6 && math.Abs(v-h.gate) < 1.15 {
				return false
			}
			gu := (h.sd(u+0.05, v) - h.sd(u-0.05, v)) / 0.1
			gv := (h.sd(u, v+0.05) - h.sd(u, v-0.05)) / 0.1
			q := h.world(u+gu*1.6, v+gv*1.6)
			if m.at(q.X, q.Y) > y0+1.0 {
				return false
			}
			if y > 0.8 {
				return true
			}
			return math.Mod(math.Abs(u+v)+100, 2.4) < 0.8
		}})
	})
	yc := y0 + rd*0.92
	ring := func(u, v float64, p Vec3, cy float64) float64 { return math.Hypot(v, p.Y-cy) }
	// Clear the porch so the door reads as set into the face.
	g.Carve(Func{box, func(p Vec3) bool {
		u, v, y := loc(p)
		return u > -0.15 && u < 1.6 && y > 0.1 && ring(u, v, p, yc) < rd+1.0
	}})
	g.With(Stone, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			u, v, y := loc(p)
			d := ring(u, v, p, yc)
			return u > -1.0 && u < 0.2 && y > 0 && d > rd+0.7 && d < rd+1.6
		}})
	})
	g.With(h.frame, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			u, v, y := loc(p)
			d := ring(u, v, p, yc)
			return u > -1.0 && u < 0.6 && y > 0 && d > rd && d <= rd+0.8
		}})
	})
	g.With(h.door1, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			u, v, y := loc(p)
			return u > -1.0 && u < 0.15 && y > 0 && ring(u, v, p, yc) <= rd
		}})
	})
	g.With(Gold, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			u, v, _ := loc(p)
			return u > 0.1 && u < 0.8 && ring(u, v, p, yc) < 0.5
		}})
	})
	// Doorstep.
	g.With(Stone, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			u, v, y := loc(p)
			return u > 0 && u < 1.3 && math.Abs(v) < rd && y > 0 && y < 0.45
		}})
	})
	// Round windows either side of the door.
	offsets := []float64{-(rd + 2.3), rd + 2.3}
	if h.big {
		offsets = []float64{-(rd + 2.6), rd + 2.6, -(rd + 5.3), rd + 5.3}
	} else if Hash01(h.seed, 3) < 0.35 {
		offsets = offsets[1:]
	}
	for _, wv := range offsets {
		cy := yc + 0.3
		g.With(h.frame, func() {
			g.Fill(Func{box, func(p Vec3) bool {
				u, v, _ := loc(p)
				d := math.Hypot(v-wv, p.Y-cy)
				return u > -1.0 && u < 0.45 && d > 0.85 && d <= 1.4
			}})
		})
		g.With(Glass, func() {
			g.Fill(Func{box, func(p Vec3) bool {
				u, v, _ := loc(p)
				return u > -1.0 && u < 0.0 && math.Hypot(v-wv, p.Y-cy) <= 0.85
			}})
		})
	}
	// A bench under one window.
	if h.big || Hash01(h.seed, 11) < 0.5 {
		side := rd + 2.3
		if h.big {
			side = rd + 3.9
		}
		bench := h.world(1.1, -side)
		g.With(Wood, func() {
			g.Fill(Func{box, func(p Vec3) bool {
				u, v, y := loc(p)
				along := math.Abs(v + side)
				return p.Plan().Sub(bench).Length() < 2 && along < 1.2 && math.Abs(u-1.1) < 0.45 && y > 0 && y < 0.9 && (y > 0.5 || along > 0.8)
			}})
		})
	}
	// Letterbox on a post beside the gate.
	post := h.world(h.d+0.2, h.gate+1.9)
	g.With(Wood, func() { g.Fill(Cylinder{Center: post, Radius: 0.4, Y0: y0, Y1: y0 + 1.6}) })
	g.With(Red, func() { g.Fill(BlockAt(post, 0.9, 0.9, y0+1.6, y0+2.3)) })
	// Chimneys poking out of the turf behind the face.
	chims := []float64{(Hash01(h.seed, 5)*2 - 1) * h.w * 0.5}
	if h.big {
		chims = []float64{-h.w * 0.55, h.w * 0.4}
	}
	for i, cv := range chims {
		at := h.world(-4.8-float64(i), cv)
		b := m.at(at.X, at.Y)
		c := Brick
		if Hash01(h.seed, i, 9) < 0.4 {
			c = Stone
		}
		g.With(c, func() { g.Fill(Cylinder{Center: at, Radius: 0.7, Y0: b - 1.5, Y1: b + 2.0}) })
		g.With(Charcoal, func() { g.Fill(Cylinder{Center: at, Radius: 0.9, Y0: b + 2.0, Y1: b + 2.5}) })
	}
}

// crown is a hollow clump of ellipsoids: foliage shells dappled with lighter leaves.
func (hv HillsideBurrows) crown(g *Grid, clumps []Ellipsoid, shell float64, seed int) {
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
	g.With(Foliage, func() {
		g.FillVolume(Func{box, func(p Vec3) bool { return inside(p, 0) && !inside(p, -shell) }})
	})
	g.Paint(Grass, Func{box, func(p Vec3) bool {
		return inside(p, 0) && hv.noise(p.X/1.7+p.Y/3, p.Z/1.7, seed) > 0.62
	}})
}

// tree is a round deciduous tree of a few clumps on a trunk.
func (hv HillsideBurrows) tree(g *Grid, at Vec2, base, size float64, seed int) {
	g.With(Wood, func() {
		g.Fill(Cylinder{Center: at, Radius: 0.25 + size*0.07, Y0: base - 1, Y1: base + size*0.55})
	})
	c := at.At(base + size*0.72)
	clumps := []Ellipsoid{{c.Add(V(0, size*0.08, 0)), V(size*0.3, size*0.26, size*0.3)}}
	for k := 0; k < 5; k++ {
		a := float64(k)*2*math.Pi/5 + Hash01(seed, k)*1.2
		q := Polar(at, size*0.24, a)
		clumps = append(clumps, Ellipsoid{q.At(c.Y + size*(0.1*Hash01(seed, k, 2)-0.04)), V(size*0.22, size*0.19, size*0.22)})
	}
	hv.crown(g, clumps, 1.5, seed)
}

// oak is the great oak spreading over the hilltop above the largest home.
func (hv HillsideBurrows) oak(g *Grid, m *hillsideBurrowsMap) {
	c := hillsideBurrowsOak
	b := m.at(c.X, c.Y)
	g.With(Wood, func() {
		g.FillVolume(Frustum{c, 2.8, 1.8, b - 2, b + 7})
		for k := 0; k < 6; k++ {
			a := float64(k)*2*math.Pi/6 + 0.4
			q := Polar(c, 4.8, a)
			g.Fill(Capsule{c.At(b + 1.8), q.At(m.at(q.X, q.Y) - 0.2), 1.0})
		}
		for k := 0; k < 7; k++ {
			a := float64(k)*2*math.Pi/7 + 0.25
			mid := Polar(c, 4.5, a).At(b + 9 + Hash01(k, 60))
			end := Polar(c, 9.5, a+0.25).At(b + 11.5 + 2*Hash01(k, 61))
			g.Fill(Capsule{c.At(b + 6), mid, 1.0}, Capsule{mid, end, 0.8})
		}
		g.Fill(Capsule{c.At(b + 6), c.At(b + 15), 1.2})
	})
	var clumps []Ellipsoid
	for k := 0; k < 9; k++ {
		a := float64(k)*2*math.Pi/9 + 0.2
		q := Polar(c, 9+1.2*Hash01(k, 64), a)
		clumps = append(clumps, Ellipsoid{q.At(b + 13 + 1.5*Hash01(k, 62)), V(5, 3.8, 5)})
	}
	for k := 0; k < 5; k++ {
		q := Polar(c, 4.5, float64(k)*2*math.Pi/5+0.6)
		clumps = append(clumps, Ellipsoid{q.At(b + 16.5 + Hash01(k, 65)), V(5.5, 4.2, 5.5)})
	}
	clumps = append(clumps, Ellipsoid{c.At(b + 19), V(5, 3.8, 5)})
	hv.crown(g, clumps, 1.4, 63)
}

// bridge is the double-arched stone bridge carrying the lane over the stream.
func (hv HillsideBurrows) bridge(g *Grid) {
	c := hillsideBurrowsBridge
	dir := P(20, 7).Scale(1 / math.Hypot(20, 7))
	n := P(-dir.Y, dir.X)
	deck := func(s float64) float64 { return 4.3 - 2.2*(s/8.5)*(s/8.5) }
	box := Box{V(c.X-10, 0, c.Y-10), V(c.X+10, 6.5, c.Y+10)}
	local := func(p Vec3) (s, t float64) {
		r := p.Plan().Sub(c)
		return r.Dot(dir), r.Dot(n)
	}
	arch := func(s, y float64) float64 { // Distance-like measure to the arch intrados.
		best := math.Inf(1)
		for _, sk := range []float64{-2.55, 2.55} {
			ds := math.Abs(s - sk)
			d := ds
			if y > 1.2 {
				d = math.Hypot(ds, y-1.2)
			}
			best = math.Min(best, d)
		}
		return best
	}
	g.With(Stone, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			s, t := local(p)
			if math.Abs(s) > 8.5 || math.Abs(t) > 2.5 || p.Y < 0.2 {
				return false
			}
			top := deck(s)
			if math.Abs(t) > 1.7 {
				top += 1.0
			}
			return p.Y <= top && arch(s, p.Y) > 2.1
		}})
	})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		s, t := local(p)
		a := arch(s, p.Y)
		return math.Abs(s) <= 8.5 && math.Abs(t) > 1.6 && math.Abs(t) <= 2.5 && a > 2.05 && a < 2.9 && p.Y > 1.2
	}})
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		s, t := local(p)
		return math.Abs(s) <= 8.5 && math.Abs(t) <= 1.7 && p.Y > deck(s)-0.5
	}})
}

// mill is a stone and timber mill on the pond's east bank with an overshot wheel.
func (hv HillsideBurrows) mill(g *Grid, m *hillsideBurrowsMap) {
	x0, x1, z0, z1 := 39.8, 46.8, 10.5, 17.5
	const floor, mid, eave = 2.2, 5.8, 8.6
	g.With(Stone, func() {
		g.Fill(Block(V(x0-0.3, -0.5, z0-0.3), V(x1+0.3, floor, z1+0.3)))
		g.Fill(Subtract(Block(V(x0, floor, z0), V(x1, mid, z1)), Block(V(x0+1, floor, z0+1), V(x1-1, mid, z1-1))))
	})
	g.With(White, func() {
		g.Fill(Subtract(Block(V(x0, mid, z0), V(x1, eave, z1)), Block(V(x0+1, mid, z0+1), V(x1-1, eave, z1-1))))
		// Gable end walls.
		g.Fill(Intersect(Gable{P(x0, z0), P(x1, z1), eave, AxisX},
			Union(Block(V(x0, eave, z0), V(x0+1, 20, z1)), Block(V(x1-1, eave, z0), V(x1, 20, z1)))))
	})
	// Timber framing on the upper storey.
	g.Paint(Wood, Func{Box{V(x0-0.5, mid, z0-0.5), V(x1+0.5, 15, z1+0.5)}, func(p Vec3) bool {
		edge := p.X < x0+0.6 || p.X > x1-0.6 || p.Z < z0+0.6 || p.Z > z1-0.6
		if !edge {
			return false
		}
		along := p.X
		if p.X < x0+0.6 || p.X > x1-0.6 {
			along = p.Z
		}
		return p.Y < mid+0.5 || (p.Y > eave-0.5 && p.Y < eave+0.1) || math.Mod(along+100, 2.4) < 0.75
	}})
	g.With(Slate, func() {
		roof := Gable{P(x0-0.7, z0-0.7), P(x1+0.7, z1+0.7), eave - 0.2, AxisX}
		g.Fill(Subtract(roof, Translate(roof, V(0, -1.0, 0))))
	})
	g.With(Brick, func() { g.Fill(Cylinder{Center: P(x1-1.6, z0+2.2), Radius: 0.7, Y0: eave, Y1: eave + 5.6}) })
	g.With(Charcoal, func() { g.Fill(Cylinder{Center: P(x1-1.6, z0+2.2), Radius: 0.9, Y0: eave + 5.6, Y1: eave + 6.1}) })
	// Door and windows on the lane side (+Z) and the pond side (-X).
	g.Paint(Wood, Block(V(x1-4.2, floor, z1-0.6), V(x1-2.4, floor+3, z1+0.1)))
	g.Carve(Block(V(x1-4.2, floor, z1-0.3), V(x1-2.4, floor+2.8, z1+0.1)))
	g.With(Wood, func() { g.Fill(Block(V(x1-4.2, floor, z1-0.9), V(x1-2.4, floor+2.8, z1-0.4))) })
	g.Paint(Glass, Block(V(x0+1.4, floor+1.2, z1-0.6), V(x0+2.8, floor+2.8, z1+0.1)),
		Block(V(x0+1.4, mid+0.9, z1-0.6), V(x0+2.8, mid+2.4, z1+0.1)),
		Block(V(x1-3.8, mid+0.9, z1-0.6), V(x1-2.4, mid+2.4, z1+0.1)),
		Block(V(x0-0.1, mid+0.9, z0+2.0), V(x0+0.6, mid+2.4, z0+3.2)),
		Block(V(x0-0.1, mid+0.9, z1-3.2), V(x0+0.6, mid+2.4, z1-2.0)),
		Block(V(x0+0.5, eave+1.0, 13.3), V(x0+1.1, eave+2.4, 14.7)))
	// The wheel: rim, spokes, paddles and hub turning on an axle from the west wall.
	wc := V(38.6, 4.3, 14)
	g.With(Wood, func() {
		g.Fill(Func{Box{V(wc.X-0.6, wc.Y-4.2, wc.Z-4.2), V(wc.X+0.6, wc.Y+4.2, wc.Z+4.2)}, func(p Vec3) bool {
			dy, dz := p.Y-wc.Y, p.Z-wc.Z
			r := math.Hypot(dy, dz)
			if r <= 0.9 || (r >= 2.6 && r <= 3.25) {
				return true
			}
			a := math.Atan2(dy, dz)
			if r < 2.6 {
				for k := 0; k < 6; k++ {
					if s := math.Abs(r * math.Sin(a-float64(k)*math.Pi/3)); s < 0.4 && math.Cos(a-float64(k)*math.Pi/3) > 0 {
						return true
					}
				}
				return false
			}
			if r > 4.1 {
				return false
			}
			for k := 0; k < 12; k++ {
				if s := math.Abs(r * math.Sin(a-float64(k)*math.Pi/6)); s < 0.45 && math.Cos(a-float64(k)*math.Pi/6) > 0 {
					return true
				}
			}
			return false
		}})
	})
	g.With(Charcoal, func() { g.Fill(Capsule{V(wc.X-0.9, wc.Y, wc.Z), V(x0+0.2, wc.Y, wc.Z), 0.5}) })
}

// party is the party field: the great party tree, a striped marquee with tables and
// strings of bunting.
func (hv HillsideBurrows) party(g *Grid, m *hillsideBurrowsMap) {
	const f = hillsideBurrowsField
	treeAt := P(-31, 25)
	hv.tree(g, treeAt, m.at(treeAt.X, treeAt.Y), 15, 70)
	x0, x1, z0, z1 := -25.0, -15.0, 29.5, 36.5
	g.With(Wood, func() {
		for _, x := range []float64{x0, -20, x1} {
			for _, z := range []float64{z0, z1} {
				g.Fill(BlockAt(P(x, z), 0.8, 0.8, f-0.5, f+3.4))
			}
		}
		for _, z := range []float64{31.8, 34.2} {
			g.Fill(Block(V(x0+2, f+0.9, z-0.5), V(x1-2, f+1.3, z+0.5)))
			for _, x := range []float64{x0 + 2.3, x1 - 2.3} {
				g.Fill(BlockAt(P(x, z), 0.6, 0.6, f, f+1))
			}
		}
	})
	roof := Gable{P(x0-0.5, z0-0.5), P(x1+0.5, z1+0.5), f + 3.4, AxisX}
	g.With(White, func() {
		g.Fill(Subtract(roof, Translate(roof, V(0, -0.9, 0))))
		// Scalloped valance.
		g.Fill(Func{Box{V(x0-0.5, f+2.4, z0-0.5), V(x1+0.5, f+3.4, z1+0.5)}, func(p Vec3) bool {
			edge := p.X < x0+0.1 || p.X > x1-0.1 || p.Z < z0+0.1 || p.Z > z1-0.1
			along := p.X + p.Z
			return edge && (p.Y > f+3.0 || math.Mod(along+100, 1.6) < 1.0)
		}})
	})
	g.Paint(Red, Func{Box{V(x0-0.6, f+2.3, z0-0.6), V(x1+0.6, f+8, z1+0.6)}, func(p Vec3) bool {
		return int(math.Floor((p.X+100)/1.5))%2 == 0
	}})
	g.With(Gold, func() { g.Fill(Cylinder{Center: P(-20, 33), Radius: 0.5, Y0: f + 7, Y1: f + 8.6}) })
	// Bunting strung from the party tree to the marquee and to a pole.
	pole := P(-36, 37)
	g.With(Wood, func() { g.Fill(Cylinder{Center: pole, Radius: 0.5, Y0: f - 0.5, Y1: f + 6.5}) })
	strings := [][2]Vec3{
		{treeAt.At(f + 7), V(x0-0.5, f+6.6, 33)},
		{treeAt.At(f + 7), pole.At(f + 6.3)},
		{pole.At(f + 6.3), V(x0-0.5, f+5.8, z1+0.4)},
	}
	colors := []Color{Red, Gold, White, Navy, Grass}
	for si, st := range strings {
		n := int(st[1].Sub(st[0]).Length() / 0.8)
		for i := 1; i < n; i++ {
			t := float64(i) / float64(n)
			p := st[0].Add(st[1].Sub(st[0]).Scale(t)).Sub(V(0, 1.6*4*t*(1-t), 0))
			c := colors[(i+si)%len(colors)]
			g.With(c, func() {
				g.AddAt(p)
				if i%2 == 0 {
					g.AddAt(p.Sub(V(0, 0.8, 0)))
				}
			})
		}
	}
}

// plots are fenced vegetable gardens with rows of greens, pumpkins and tomatoes.
func (hv HillsideBurrows) plots(g *Grid, m *hillsideBurrowsMap) {
	type plot struct {
		c    Vec2
		w, d float64
	}
	for pi, pl := range []plot{{P(-43, -1), 8, 6}, {P(39, 3.5), 7, 5}, {P(3.5, 13), 6, 5}} {
		x0, x1 := pl.c.X-pl.w/2, pl.c.X+pl.w/2
		z0, z1 := pl.c.Y-pl.d/2, pl.c.Y+pl.d/2
		box := Box{V(x0-1, -1, z0-1), V(x1+1, 20, z1+1)}
		inner := func(p Vec3) bool { return p.X > x0+0.6 && p.X < x1-0.6 && p.Z > z0+0.6 && p.Z < z1-0.6 }
		g.Paint(Wood, Func{box, func(p Vec3) bool {
			return p.X > x0 && p.X < x1 && p.Z > z0 && p.Z < z1 && p.Y > m.at(p.X, p.Z)-0.9
		}})
		row := func(p Vec3) (int, bool) {
			r := (p.Z - z0 - 0.6) / 1.6
			return int(math.Floor(r)), r-math.Floor(r) < 0.55
		}
		crops := []Color{Foliage, Grass, Red, Foliage, Grass}
		for ci, c := range crops {
			g.With(c, func() {
				g.Fill(Func{box, func(p Vec3) bool {
					y := p.Y - m.at(p.X, p.Z)
					k, ok := row(p)
					if !inner(p) || !ok || y < 0.1 || y > 0.9 || (k+pi)%len(crops) != ci {
						return false
					}
					return Hash01(int(math.Round(p.X*8)), int(math.Round(p.Z*8)), 80) < 0.85
				}})
			})
		}
		// A row of pumpkins along the front.
		g.With(Gold, func() {
			for x := x0 + 1.3; x < x1-0.8; x += 1.9 {
				q := P(x, z1-1.1)
				g.Fill(Sphere(q.At(m.at(q.X, q.Y)+0.5), 0.6))
			}
		})
		g.With(Wood, func() {
			g.Fill(Func{box, func(p Vec3) bool {
				edge := p.X < x0+0.4 && p.X > x0-0.4 || p.X > x1-0.4 && p.X < x1+0.4 || p.Z < z0+0.4 && p.Z > z0-0.4 || p.Z > z1-0.4 && p.Z < z1+0.4
				if !edge || p.X < x0-0.4 || p.X > x1+0.4 || p.Z < z0-0.4 || p.Z > z1+0.4 {
					return false
				}
				y := p.Y - m.at(p.X, p.Z)
				if y < 0.1 || y > 1.5 {
					return false
				}
				return (y > 0.8 && y < 1.25) || math.Mod(p.X+p.Z+100, 1.6) < 0.8
			}})
		})
	}
}

// hedges are rows of bushy foliage along lanes and field edges.
func (hv HillsideBurrows) hedges(g *Grid, m *hillsideBurrowsMap) {
	lines := [][]Vec2{
		{P(-41, 21), P(-40.5, 32), P(-34, 40)},
		{P(-49, 8), P(-40, 8.4), P(-31, 12.5)},
		{P(19, 32.5), P(29, 33.5), P(37, 30.5)},
		{P(43, -10), P(45, -1), P(44, 6)},
		{P(-30, -24), P(-22, -31), P(-14, -38)},
		{P(8, -38), P(18, -29), P(30, -31), P(36, -36)},
		{P(-49, -8), P(-40, -9), P(-34, -21)},
	}
	// Hedgerows lining stretches of the lane, broken where footpaths and the bridge meet it.
	paths := hv.branches(hv.homes())
	laneBox := Box{V(-56, -1, -10), V(56, 14, 40)}
	g.With(Foliage, func() {
		g.Fill(Func{laneBox, func(p Vec3) bool {
			q := p.Plan()
			d := hv.segDist(hillsideBurrowsLane, q)
			if d < 2.0 || d > 3.2 || hv.noise(q.X/7, q.Y/7, 97) < 0.42 {
				return false
			}
			h := m.at(q.X, q.Y)
			y := p.Y - h
			if h < 0.45 || y < -0.5 || y > 1.4+0.7*hv.noise(q.X/1.5, q.Y/1.5, 98) || hv.wet(q.X, q.Y) {
				return false
			}
			if q.Sub(hillsideBurrowsBridge).Length() < 11 || q.Sub(P(43.3, 19)).Length() < 5 {
				return false
			}
			for _, pth := range paths {
				if hv.segDist(pth, q) < 2.2 {
					return false
				}
			}
			return hv.segDist(hillsideBurrowsStream, q) > 4
		}})
	})
	for li, line := range lines {
		box := EmptyBox()
		for _, q := range line {
			box = box.Union(Box{q.At(-1), q.At(12)})
		}
		box = box.Grow(1.5)
		g.With(Foliage, func() {
			g.Fill(Func{box, func(p Vec3) bool {
				q := p.Plan()
				d := hv.segDist(line, q)
				y := p.Y - m.at(q.X, q.Y)
				top := 1.5 + 0.8*hv.noise(q.X/1.5, q.Y/1.5, 90+li)
				return d < 0.9+0.3*hv.noise(q.X/2, q.Y/2, 95+li) && y > -0.5 && y < top && m.at(q.X, q.Y) > 0.45
			}})
		})
	}
}

// trees scatters round trees over the hilltops, the pond banks and the meadow.
func (hv HillsideBurrows) trees(g *Grid, m *hillsideBurrowsMap) {
	type tr struct {
		at   Vec2
		size float64
	}
	for i, t := range []tr{
		{P(33, 24), 8}, {P(-33, -11), 10}, {P(29, -19), 11},
		{P(-22, -30), 10}, {P(18, -32), 11}, {P(-42, -14), 9}, {P(42, -18), 9},
		{P(6, -38), 9}, {P(-12, -38), 9}, {P(22, 39), 9},
		{P(36, 36), 8}, {P(-44, 15), 9}, {P(-17, 3), 7},
		{P(-47, 6), 8}, {P(-28, -30), 9}, {P(30, -30), 9}, {P(47.5, -3), 8},
	} {
		hv.tree(g, t.at, m.at(t.at.X, t.at.Y), t.size, 200+i)
		if i%3 == 1 {
			// Apple trees.
			r := t.size * 0.55
			g.Paint(Red, Func{Box{V(t.at.X-r, 0, t.at.Y-r), V(t.at.X+r, 40, t.at.Y+r)}, func(p Vec3) bool {
				return Hash01(int(math.Round(p.X*8)), int(math.Round(p.Y*8)), int(math.Round(p.Z*8)), 77) < 0.07 &&
					p.Y > m.at(t.at.X, t.at.Y)+t.size*0.5
			}})
		}
	}
	// Rowing boat and jetty on the mill pond, stepping stones across the stream.
	g.With(Wood, func() {
		g.Fill(Block(V(28.6, hillsideBurrowsWater+0.3, 16.5), V(30.4, hillsideBurrowsWater+0.7, 22.5)))
		for _, z := range []float64{17, 19.5} {
			for _, x := range []float64{28.8, 30.2} {
				g.Fill(Cylinder{Center: P(x, z), Radius: 0.4, Y0: 0, Y1: hillsideBurrowsWater + 0.7})
			}
		}
	})
	boat := V(25, hillsideBurrowsWater+0.6, 13.5)
	g.With(Red, func() {
		g.FillVolume(Subtract(Intersect(Ellipsoid{boat, V(1.4, 1.2, 3)}, Block(boat.Sub(V(2, 1.2, 3.5)), boat.Add(V(2, 0.4, 3.5)))),
			Ellipsoid{boat.Add(V(0, 0.4, 0)), V(0.8, 1, 2.3)}))
	})
	g.With(Wood, func() { g.Fill(BlockAt(P(boat.X, boat.Z), 2.4, 0.7, boat.Y-0.3, boat.Y+0.1)) })
	g.With(Stone, func() {
		for i, z := range []float64{36, 38.2, 40.4, 42.6} {
			x := 9.8 - 0.2*float64(i) - 0.3*(z-36)
			g.Fill(Sphere(V(x, hillsideBurrowsWater+0.2, z), 0.75))
		}
	})
}
