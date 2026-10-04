package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// GreatWall is a Ming section of the Great Wall in the Mutianyu/Jinshanling style,
// snaking in an S-curve along the knife-edge ridges of forested mountains. The wall
// path runs from the far west (-X) to the near east (+X): it climbs to the high
// back-left summit, plunges into a central valley where a pass fortress with a
// double-eaved gate pavilion straddles the road, then climbs the front-right ridge
// and drops toward the eastern rim. The grey brick wall has a crenellated outer
// parapet, a lower crenellated inner parapet and a stepped walkway on the climbs.
// Watchtowers of several designs stand on the summits and shoulders: plain
// crenellated towers, towers with a small tiled guardroom, two-storey towers with
// pavilion roofs and two ruined, roofless ones near the ends. A village lines the
// road outside the gate; autumn colors patch the forest.
type GreatWall struct{}

func init() { Register(GreatWall{}) }

func (GreatWall) Name() string  { return "great_wall" }
func (GreatWall) Title() string { return "Great Wall" }

const (
	greatWallDX     = 0.2  // Path sample spacing.
	greatWallHalf   = 2.4  // Half-width of the wall body.
	greatWallRim    = 1.0  // Parapet thickness.
	greatWallHeight = 2.8  // Walkway height above the ridge.
	greatWallStep   = 0.8  // Stair riser on the walkway.
	greatWallFloor  = 1.2  // Valley floor.
	greatWallSiteX  = 63.0 // Site half-extents (superellipse rim).
	greatWallSiteZ  = 46.0
	greatWallWorldX = 77.0 // World plan half-extents enclosing the turned site.
	greatWallWorldZ = 72.0
	greatWallTowerA = 4.0 // Tower half-size.
	greatWallGateU  = 8.0 // Gate fortress half-length along the wall.
	greatWallGateV  = 5.5 // Gate fortress half-depth across the wall.
	greatWallGateH  = 7.4 // Gate platform height above the valley floor.
)

// Tower kinds.
const (
	greatWallPlain = iota
	greatWallGuardroom
	greatWallTwoStorey
	greatWallRuin
)

func greatWallSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

// greatWallNoise is smooth value noise in [0, 1).
func greatWallNoise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

func greatWallFBM(x, z float64, seed int) float64 {
	return (greatWallNoise(x, z, seed)*4 + greatWallNoise(x*2.1+5, z*2.1+3, seed+1)*2 +
		greatWallNoise(x*4.3+1, z*4.3+7, seed+2)) / 7
}

// The site is authored in a local frame (a along the wall's main axis, b across it)
// turned so the S faces the hero camera.
const greatWallPhi = -0.6

func greatWallToWorld(a, b float64) (float64, float64) {
	s, c := math.Sincos(greatWallPhi)
	return a*c - b*s, a*s + b*c
}

func greatWallToLocal(x, z float64) (float64, float64) {
	s, c := math.Sincos(greatWallPhi)
	return x*c + z*s, -x*s + z*c
}

// greatWallRimE is the normalized site radius in the local frame; above 1 is off the plot.
func greatWallRimE(a, b float64) float64 {
	ea, eb := math.Abs(a)/greatWallSiteX, math.Abs(b)/greatWallSiteZ
	return math.Pow(ea*ea*ea*ea+eb*eb*eb*eb, 0.25)
}

// greatWallKnots are the wall path's control points in the local frame (a, b) with
// the ridge-top height under the wall: from the low western end up to the far
// summit, down into the pass, over the low near lobe and up to the eastern spur.
var greatWallKnots = [][3]float64{
	{-62, 12, 1}, {-56, 6, 3}, {-48, -5, 10}, {-39, -20, 19}, {-27, -31, 29}, {-15, -28, 22},
	{-6, -14, 9}, {0, 0, 1.2}, {6, 14, 6.5}, {15, 26, 9}, {27, 31, 10.5}, {39, 26, 14},
	{47, 13, 20.5}, {53, -2, 15}, {57, -14, 5}, {60, -22, 1},
}

// greatWallSpline evaluates the Catmull-Rom spline through the knots; t runs over
// knot intervals.
func greatWallSpline(t float64) [3]float64 {
	k := greatWallKnots
	n := len(k)
	i := int(math.Floor(t))
	i = max(0, min(n-2, i))
	u := t - float64(i)
	u2, u3 := u*u, u*u*u
	var r [3]float64
	for c := 0; c < 3; c++ {
		p0, p1, p2, p3 := k[max(i-1, 0)][c], k[i][c], k[i+1][c], k[min(i+2, n-1)][c]
		r[c] = 0.5 * (2*p1 + (p2-p0)*u + (2*p0-5*p1+4*p2-p3)*u2 + (3*p1-p0-3*p2+p3)*u3)
	}
	return r
}

// greatWallValley is the road valley that crosses the wall at the pass fortress,
// authored in the local frame and stored in world plan coordinates.
var greatWallValley = func() []Vec2 {
	var line []Vec2
	for _, q := range []Vec2{
		P(-68, 26), P(-52, 21), P(-36, 15), P(-22, 9.5), P(-11, 4.7), P(0, 0),
		P(11, -4.7), P(21, -11), P(24, -26), P(22, -50),
	} {
		x, z := greatWallToWorld(q.X, q.Y)
		line = append(line, P(x, z))
	}
	return line
}()

func greatWallSegDist(p, a, b Vec2) (float64, float64) {
	ab := b.Sub(a)
	t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
	return p.Sub(a.Add(ab.Scale(t))).Length(), t
}

func greatWallPolyDist(p Vec2, line []Vec2) float64 {
	best := math.Inf(1)
	for i := 0; i+1 < len(line); i++ {
		d, _ := greatWallSegDist(p, line[i], line[i+1])
		best = math.Min(best, d)
	}
	return best
}

// greatWallMap is a sampled heightfield with bilinear lookup.
type greatWallMap struct {
	x0, z0, step float64
	nx, nz       int
	h            []float64
}

func (m *greatWallMap) at(x, z float64) float64 {
	fx, fz := (x-m.x0)/m.step, (z-m.z0)/m.step
	if fx < 0 || fz < 0 || fx >= float64(m.nx-1) || fz >= float64(m.nz-1) {
		return 0.6
	}
	i, j := int(fx), int(fz)
	tx, tz := fx-float64(i), fz-float64(j)
	k := j*m.nx + i
	a := m.h[k] + (m.h[k+1]-m.h[k])*tx
	b := m.h[k+m.nx] + (m.h[k+m.nx+1]-m.h[k+m.nx])*tx
	return a + (b-a)*tz
}

func (m *greatWallMap) low(x, z float64) float64 {
	lo := m.at(x, z)
	for k := 0; k < 8; k++ {
		a := float64(k) * math.Pi / 4
		lo = math.Min(lo, m.at(x+0.8*math.Cos(a), z+0.8*math.Sin(a)))
	}
	return lo
}

// greatWallTower is one watchtower on the path.
type greatWallTower struct {
	a, b  float64 // Approximate local position; snapped to a path sample.
	kind  int
	i     int     // Path sample index.
	floor float64 // Walkway level inside the tower.
}

// greatWallSite holds the shared state while the scene is built.
type greatWallSite struct {
	g      *Grid
	hm     *greatWallMap
	n      int
	pa, pb []float64 // Path samples in the local frame.
	px, pz []float64 // Path samples in world plan coordinates.
	tx, tz []float64 // Unit tangents.
	s      []float64 // Arc length.
	top    []float64 // Quantized walkway height.
	crest  []float64 // Ridge-top knot height at each sample.
	ground []float64 // Smoothed ridge height under the wall.
	gate   int       // Path sample of the pass fortress.
	gateY  float64   // Valley floor at the gate.
	towers []greatWallTower
}

func (GreatWall) Build(g *Grid) Order {
	w := &greatWallSite{g: g}
	w.path()
	w.heights()
	w.towers = []greatWallTower{
		{a: -52, b: 0, kind: greatWallRuin}, {a: -40, b: -19, kind: greatWallGuardroom},
		{a: -27, b: -31, kind: greatWallTwoStorey}, {a: -13, b: -26, kind: greatWallPlain},
		{a: 15, b: 26, kind: greatWallGuardroom}, {a: 30, b: 31, kind: greatWallPlain},
		{a: 47, b: 13, kind: greatWallTwoStorey}, {a: 56, b: -11, kind: greatWallRuin},
	}
	w.walkway()
	w.wall()
	for _, t := range w.towers {
		w.tower(t)
	}
	w.gateFort()
	w.village()
	w.terrain()
	w.trees()
	return Along(P(1, 0.2))
}

// path samples the S-curve evenly by arc length.
func (w *greatWallSite) path() {
	type pt struct{ a, b, h, s float64 }
	var dense []pt
	end := float64(len(greatWallKnots) - 1)
	for t := 0.0; t <= end+1e-9; t += 0.01 {
		r := greatWallSpline(t)
		q := pt{r[0], r[1], r[2], 0}
		if len(dense) > 0 {
			l := dense[len(dense)-1]
			q.s = l.s + math.Hypot(q.a-l.a, q.b-l.b)
		}
		dense = append(dense, q)
	}
	total := dense[len(dense)-1].s
	j := 0
	for s := 0.0; s <= total; s += greatWallDX {
		for j+2 < len(dense) && dense[j+1].s < s {
			j++
		}
		p, q := dense[j], dense[j+1]
		f := (s - p.s) / (q.s - p.s)
		a, b := p.a+(q.a-p.a)*f, p.b+(q.b-p.b)*f
		l := math.Hypot(q.a-p.a, q.b-p.b)
		x, z := greatWallToWorld(a, b)
		tx, tz := greatWallToWorld((q.a-p.a)/l, (q.b-p.b)/l)
		w.pa, w.pb = append(w.pa, a), append(w.pb, b)
		w.px, w.pz = append(w.px, x), append(w.pz, z)
		w.tx, w.tz = append(w.tx, tx), append(w.tz, tz)
		w.s = append(w.s, s)
		w.crest = append(w.crest, p.h+(q.h-p.h)*f)
	}
	w.n = len(w.s)
	w.gate = w.sampleNear(0, 0)
}

// sampleNear returns the path sample closest to a local-frame point.
func (w *greatWallSite) sampleNear(a, b float64) int {
	best, bi := math.Inf(1), 0
	for i := range w.pa {
		if d := math.Hypot(w.pa[i]-a, w.pb[i]-b); d < best {
			best, bi = d, i
		}
	}
	return bi
}

// nearest finds the closest path sample to a plan point, its distance, signed lateral
// offset (positive toward the +normal side) and the arc length at the projection.
func (w *greatWallSite) nearest(x, z float64) (i int, dist, d, along float64) {
	best := math.Inf(1)
	for j := 0; j < w.n; j += 20 {
		dx, dz := x-w.px[j], z-w.pz[j]
		if q := dx*dx + dz*dz; q < best {
			best, i = q, j
		}
	}
	lo, hi := max(0, i-25), min(w.n-1, i+25)
	for j := lo; j <= hi; j++ {
		dx, dz := x-w.px[j], z-w.pz[j]
		if q := dx*dx + dz*dz; q < best {
			best, i = q, j
		}
	}
	dx, dz := x-w.px[i], z-w.pz[i]
	return i, math.Sqrt(best), dx*-w.tz[i] + dz*w.tx[i], w.s[i] + dx*w.tx[i] + dz*w.tz[i]
}

// raw is the natural ground height at a plan point; ok is false off the site.
func (w *greatWallSite) raw(x, z float64) (float64, bool) {
	la, lb := greatWallToLocal(x, z)
	e := greatWallRimE(la, lb)
	if e > 1 {
		return 0, false
	}
	// Ridge under the wall: a union of cones along the spine, with a narrow flat top.
	k := 1.5 + 0.8*(greatWallFBM(x/10, z/10, 3)-0.5)
	best, dmin := -100.0, math.Inf(1)
	for j := 0; j < w.n; j += 5 {
		d := math.Hypot(x-w.px[j], z-w.pz[j])
		dmin = math.Min(dmin, d)
		if v := w.crest[j] - k*math.Max(0, d-greatWallHalf-0.4); v > best {
			best = v
		}
	}
	h := best
	// Detached peaks: a high summit behind the right of the wall, lower hills elsewhere.
	for _, pk := range [][4]float64{
		{44, -30, 31, 18}, {-44, 38, 11, 12}, {-1, -36, 20, 13}, {52, 36, 11, 11},
		{-52, -30, 18, 13}, {2, 38, 9, 10}, {-22, 38, 9, 10}, {27, 11, 15, 11},
	} {
		r := math.Hypot(la-pk[0], lb-pk[1]) * (1 + 0.35*(greatWallFBM(x/6, z/6, 11)-0.5))
		h = math.Max(h, pk[2]*(1-r/pk[3]))
	}
	// Concave foot.
	h = 1 + 2.5*math.Log1p(math.Exp((h-1)/2.5))
	// Ridged detail away from the spine: spurs and gullies on the flanks.
	rg := 1 - math.Abs(2*greatWallFBM(x/6+3, z/6-2, 21)-1)
	h += 7 * (rg - 0.5) * greatWallSmooth((dmin-4)/7) * greatWallSmooth((h-3)/6)
	// The road valley through the pass.
	dv := greatWallPolyDist(P(x, z), greatWallValley)
	h = math.Min(h, greatWallFloor+0.25*greatWallFBM(x/5, z/5, 25)+0.85*math.Max(0, dv-3.8)+0.03*math.Pow(math.Max(0, dv-3.8), 2))
	// Slope down to a low rim.
	if e > 1e-6 {
		h = math.Min(h, 0.6+1.5*math.Hypot(la, lb)*(1/e-1))
	}
	return math.Max(h, 0.6), true
}

func (w *greatWallSite) heights() {
	m := &greatWallMap{x0: -greatWallWorldX - 1, z0: -greatWallWorldZ - 1, step: 0.4}
	m.nx = int(2*(greatWallWorldX+1)/m.step) + 2
	m.nz = int(2*(greatWallWorldZ+1)/m.step) + 2
	m.h = make([]float64, m.nx*m.nz)
	for j := 0; j < m.nz; j++ {
		for i := 0; i < m.nx; i++ {
			h, ok := w.raw(m.x0+float64(i)*m.step, m.z0+float64(j)*m.step)
			if !ok {
				h = 0.6
			}
			m.h[j*m.nx+i] = h
		}
	}
	w.hm = m
}

// walkway derives the stepped walkway height along the path.
func (w *greatWallSite) walkway() {
	base := make([]float64, w.n)
	for i := range base {
		base[i] = w.hm.at(w.px[i], w.pz[i])
	}
	w.ground = make([]float64, w.n)
	raw := make([]float64, w.n)
	for i := range base {
		sum, cnt := 0.0, 0.0
		for j := max(0, i-10); j <= min(w.n-1, i+10); j++ {
			sum += base[j]
			cnt++
		}
		w.ground[i] = sum / cnt
		raw[i] = w.ground[i] + greatWallHeight
	}
	// The wall rises to the gate platform on either side of the pass.
	w.gateY = w.hm.at(w.px[w.gate], w.pz[w.gate])
	gateTop := w.gateY + greatWallGateH
	for i := range raw {
		ds := math.Abs(w.s[i] - w.s[w.gate])
		raw[i] = math.Max(raw[i], gateTop-0.42*math.Max(0, ds-greatWallGateU))
	}
	// Level landings at each tower door.
	for k := range w.towers {
		t := &w.towers[k]
		t.i = w.sampleNear(t.a, t.b)
		t.floor = math.Round(raw[t.i]/greatWallStep) * greatWallStep
	}
	w.top = make([]float64, w.n)
	for i := range raw {
		v := raw[i]
		for _, t := range w.towers {
			ds := math.Abs(w.s[i] - w.s[t.i])
			v = t.floor + (v-t.floor)*greatWallSmooth((ds-greatWallTowerA-0.5)/5)
		}
		w.top[i] = math.Round(v/greatWallStep) * greatWallStep
	}
}

// eachColumn visits every lattice column center in a plan box.
func (w *greatWallSite) eachColumn(x0, z0, x1, z1 float64, f func(cx, cz int, x, z float64)) {
	half := w.g.Span() / 2
	for cz := int(math.Ceil(z0 / half)); cz <= int(math.Floor(z1/half)); cz++ {
		for cx := int(math.Ceil(x0 / half)); cx <= int(math.Floor(x1/half)); cx++ {
			if (cx^cz)&1 != 0 {
				continue
			}
			f(cx, cz, float64(cx)*half, float64(cz)*half)
		}
	}
}

// column adds the cells of a column between two heights, colored per cell.
func (w *greatWallSite) column(cx, cz int, y0, y1 float64, color func(cy int, y float64) Color) {
	half := w.g.Span() / 2
	cy := int(math.Ceil(y0 / half))
	if (cy^cx)&1 != 0 {
		cy++
	}
	for ; float64(cy)*half <= y1; cy += 2 {
		y := float64(cy) * half
		w.g.Use(color(cy, y))
		w.g.Add(Cell{X: int32(cx), Y: int32(cy), Z: int32(cz)})
	}
}

// wall rasterizes the wall body, walkway and both parapets column by column.
func (w *greatWallSite) wall() {
	w.eachColumn(-greatWallWorldX, -greatWallWorldZ, greatWallWorldX, greatWallWorldZ, func(cx, cz int, x, z float64) {
		i, dist, d, along := w.nearest(x, z)
		ad := math.Abs(d)
		if dist > greatWallHalf+1 || ad > greatWallHalf {
			return
		}
		if (i == 0 && along < w.s[0]) || (i == w.n-1 && along > w.s[w.n-1]) {
			return
		}
		top := w.top[i]
		y1 := top
		parapet := ad > greatWallHalf-greatWallRim
		if parapet {
			if d > 0 {
				// Outer parapet: tall merlons and crenels.
				if math.Mod(along, 2.4) < 1.4 {
					y1 = top + 2.2
				} else {
					y1 = top + 1.0
				}
			} else {
				// Inner parapet: lower, with wider merlons.
				if math.Mod(along+1.2, 3.2) < 1.8 {
					y1 = top + 1.7
				} else {
					y1 = top + 0.9
				}
			}
		}
		gl := w.hm.at(x, z)
		y0 := math.Min(gl, w.hm.low(x, z)) - 0.8
		// Ruined ends: the wall crumbles toward the edges of the site.
		if e := math.Min(w.s[i], w.s[w.n-1]-w.s[i]); e < 7 {
			f := greatWallSmooth(e/7) + 0.5*(Hash01(int(along/1.3), 71)-0.5)
			y1 = math.Min(y1, gl+(top-gl)*math.Max(0.15, f))
			if f < 0.8 {
				parapet = false
			}
		}
		w.column(cx, cz, y0, y1, func(cy int, y float64) Color {
			h := Hash01(cx, cy, cz, 73)
			switch {
			case y < gl+1.2:
				if h < 0.75 {
					return DarkStone
				}
				return Stone
			case parapet && y >= top-0.45 && y < top+0.35:
				return DarkStone
			case !parapet && y > top-0.7:
				if h < 0.55 {
					return DarkStone
				}
				return Stone
			case h < 0.24:
				return DarkStone
			}
			return Stone
		})
	})
}

// local maps a solid authored around the origin, X along the wall, into the frame of
// path sample i.
func (w *greatWallSite) local(i int, s Solid) Solid {
	c := P(w.px[i], w.pz[i])
	return RotateY(Translate(s, V(c.X, 0, c.Y)), c, math.Atan2(w.tz[i], w.tx[i]))
}

// footLow is the lowest ground under a square footprint around sample i.
func (w *greatWallSite) footLow(i int, hu, hv float64) float64 {
	lo := math.Inf(1)
	for a := -1.0; a <= 1; a += 0.25 {
		for b := -1.0; b <= 1; b += 0.25 {
			u, v := a*hu, b*hv
			x := w.px[i] + u*w.tx[i] - v*w.tz[i]
			z := w.pz[i] + u*w.tz[i] + v*w.tx[i]
			lo = math.Min(lo, w.hm.at(x, z))
		}
	}
	return lo
}

// greatWallCrenel is a crenellated parapet ring around a local rectangle.
func greatWallCrenel(hu, hv, base, low, high, rim, period, merlon float64) Solid {
	return Func{Box{V(-hu, base, -hv), V(hu, base+high, hv)}, func(p Vec3) bool {
		onU := math.Abs(p.X) >= hu-rim
		onV := math.Abs(p.Z) >= hv-rim
		if !onU && !onV {
			return false
		}
		if p.Y <= base+low {
			return true
		}
		if onV && math.Mod(p.X+hu+100*period, period) < merlon {
			return true
		}
		return onU && math.Mod(p.Z+hv+100*period, period) < merlon
	}}
}

// greatWallRoof is a hollow hip roof with a ridge along X, rising at the given pitch.
func greatWallRoof(hu, hv, base, pitch, thick float64) Solid {
	in := func(p Vec3, b float64) bool {
		r := (p.Y - b) / pitch
		return p.Y >= b && math.Abs(p.Z) <= hv-r && math.Abs(p.X) <= hu-r
	}
	return Func{Box{V(-hu, base, -hv), V(hu, base+hv*pitch, hv)}, func(p Vec3) bool {
		return in(p, base) && !in(p, base+thick)
	}}
}

// room fills a hollow rectangular storey with a dark inner face.
func (w *greatWallSite) room(i int, hu, hv, y0, y1, wall, slab float64, c Color) {
	g := w.g
	g.Use(c).Fill(w.local(i, Block(V(-hu, y0, -hv), V(hu, y1, hv))))
	g.Paint(Charcoal, w.local(i, Block(V(-hu+wall-0.5, y0-0.4, -hv+wall-0.5), V(hu-wall+0.5, y1-slab+0.5, hv-wall+0.5))))
	g.Carve(w.local(i, Block(V(-hu+wall, y0+0.3, -hv+wall), V(hu-wall, y1-slab, hv-wall))))
}

// tower builds one watchtower astride the wall.
func (w *greatWallSite) tower(t greatWallTower) {
	g := w.g
	i, a, f := t.i, greatWallTowerA, t.floor
	yb := w.footLow(i, a, a) - 1
	// Clear the wall inside the footprint, then build the plinth up to the floor.
	g.Carve(w.local(i, Block(V(-a-0.2, f, -a-0.2), V(a+0.2, f+20, a+0.2))))
	g.Use(Stone).Fill(w.local(i, Block(V(-a, yb, -a), V(a, f+0.3, a))))
	g.Paint(DarkStone, w.local(i, Block(V(-a, yb, -a), V(a, math.Min(f-1.5, yb+2.2), a))))
	h1 := 4.0
	w.room(i, a, a, f, f+h1, 1.0, 0.8, Stone)
	// Arched windows on the flanks, doors and windows on the wall ends.
	for _, u := range []float64{-2.2, 0, 2.2} {
		for _, v := range []float64{-a + 0.5, a - 0.5} {
			g.Carve(w.local(i, Arch{Sill: V(u, f+1.0, v), HalfWidth: 0.6, Spring: 0.9, Depth: 1.6, Through: AxisZ}))
		}
	}
	for _, u := range []float64{-a + 0.5, a - 0.5} {
		g.Carve(w.local(i, Arch{Sill: V(u, f+0.3, 0), HalfWidth: 0.9, Spring: 1.5, Depth: 1.6, Through: AxisX}))
		for _, v := range []float64{-2.2, 2.2} {
			g.Carve(w.local(i, Arch{Sill: V(u, f+1.0, v), HalfWidth: 0.6, Spring: 0.9, Depth: 1.6, Through: AxisX}))
		}
	}
	// Cornice band.
	g.Use(DarkStone).Fill(w.local(i, Subtract(Block(V(-a-0.35, f+h1-0.8, -a-0.35), V(a+0.35, f+h1, a+0.35)),
		Block(V(-a+0.6, f+h1-1, -a+0.6), V(a-0.6, f+h1+1, a-0.6)))))
	roof := f + h1
	switch t.kind {
	case greatWallTwoStorey:
		b := a - 0.4
		w.room(i, b, b, roof, roof+3.6, 0.9, 0.8, Stone)
		for _, u := range []float64{-1.6, 1.6} {
			for _, v := range []float64{-b + 0.45, b - 0.45} {
				g.Carve(w.local(i, Arch{Sill: V(u, roof+1.0, v), HalfWidth: 0.55, Spring: 0.8, Depth: 1.4, Through: AxisZ}))
				g.Carve(w.local(i, Arch{Sill: V(v, roof+1.0, u), HalfWidth: 0.55, Spring: 0.8, Depth: 1.4, Through: AxisX}))
			}
		}
		// Pavilion roof with flared eaves.
		g.Use(Charcoal).Fill(w.local(i, greatWallRoof(b+1.2, b+1.2, roof+3.0, 0.62, 1.0)))
		g.Use(Slate).Fill(w.local(i, Block(V(-b-1.2, roof+3.0, -b-1.2), V(b+1.2, roof+3.5, b+1.2))))
		g.Carve(w.local(i, Block(V(-b+0.9, roof+3.0, -b+0.9), V(b-0.9, roof+3.6, b-0.9))))
		top := roof + 3.0 + (b+1.2)*0.62
		g.Use(Gold).Fill(w.local(i, Block(V(-0.5, top-1.0, -0.5), V(0.5, top+0.6, 0.5))))
	case greatWallGuardroom:
		g.Use(Stone).Fill(w.local(i, greatWallCrenel(a, a, roof, 1.0, 2.0, 0.9, 2.0, 1.1)))
		w.room(i, 2.4, 1.8, roof, roof+2.6, 0.8, 0.6, Stone)
		g.Carve(w.local(i, Arch{Sill: V(0, roof+0.3, 1.8-0.4), HalfWidth: 0.7, Spring: 1.1, Depth: 1.2, Through: AxisZ}))
		g.Use(Charcoal).Fill(w.local(i, greatWallRoof(3.4, 2.8, roof+2.4, 0.6, 0.9)))
		g.Use(Slate).Fill(w.local(i, Subtract(Block(V(-3.4, roof+2.4, -2.8), V(3.4, roof+2.8, 2.8)),
			Block(V(-2.0, roof+2, -1.4), V(2.0, roof+3, 1.4)))))
	case greatWallPlain:
		g.Use(Stone).Fill(w.local(i, greatWallCrenel(a, a, roof, 1.0, 2.0, 0.9, 2.0, 1.1)))
	case greatWallRuin:
		g.Use(Stone).Fill(w.local(i, greatWallCrenel(a, a, roof, 1.0, 2.0, 0.9, 2.0, 1.1)))
		// Collapsed vault and broken walls stepping down to one standing corner.
		g.Carve(w.local(i, Block(V(-a+1, f+0.3, -a+1), V(a-1, f+20, a-1))))
		g.Carve(w.local(i, Func{Box{V(-a-1, f, -a-1), V(a+1, f+20, a+1)}, func(p Vec3) bool {
			k := Hash01(int(math.Floor((p.X+10)/1.6)), int(math.Floor((p.Z+10)/1.6)), t.i, 75)
			keep := f + 1.2 + 5.5*greatWallSmooth((p.X+p.Z)/(2*a)+0.3)*(0.4+0.6*k)
			return p.Y > keep
		}}))
		// Weathered inner faces are open to the sky now.
		g.Paint(Stone, w.local(i, Block(V(-a-0.5, f+0.6, -a-0.5), V(a+0.5, f+20, a+0.5))))
		// Weeds and rubble over the open floor.
		g.Paint(Grass, w.local(i, Func{Box{V(-a+0.5, f-0.5, -a+0.5), V(a-0.5, f+0.6, a-0.5)}, func(p Vec3) bool {
			return greatWallNoise(p.X/1.5+float64(t.i), p.Z/1.5, 79) > 0.4
		}}))
		g.Paint(DarkStone, w.local(i, Func{Box{V(-a+0.5, f-0.5, -a+0.5), V(a-0.5, f+0.6, a-0.5)}, func(p Vec3) bool {
			return greatWallNoise(p.X/1.5+float64(t.i), p.Z/1.5, 79) <= 0.4
		}}))
		// Rubble at the foot.
		for k := 0; k < 9; k++ {
			u := (Hash01(k, t.i, 76) - 0.5) * 2 * (a + 2)
			v := a + 0.8 + 1.5*Hash01(k, t.i, 77)
			if k%2 == 0 {
				v = -v
			}
			x := w.px[i] + u*w.tx[i] - v*w.tz[i]
			z := w.pz[i] + u*w.tz[i] + v*w.tx[i]
			r := 0.7 + 0.5*Hash01(k, t.i, 78)
			g.Use(DarkStone).FillVolume(Sphere(V(x, w.hm.at(x, z)+0.2, z), r))
		}
	}
}

// gateFort builds the pass fortress: a brick platform with an arched road tunnel,
// crenellations and a two-storey gate pavilion with double eaves.
func (w *greatWallSite) gateFort() {
	g, i := w.g, w.gate
	U, Vv := greatWallGateU, greatWallGateV
	yb := w.footLow(i, U, Vv)
	top := w.gateY + greatWallGateH
	g.Carve(w.local(i, Block(V(-U-0.2, yb-0.5, -Vv-0.2), V(U+0.2, top+30, Vv+0.2))))
	g.Use(Stone).Fill(w.local(i, Block(V(-U, yb-1.5, -Vv), V(U, top, Vv))))
	g.Paint(DarkStone, w.local(i, Block(V(-U, yb-1.5, -Vv), V(U, w.gateY+1.6, Vv))))
	// Road tunnel across the wall with a dark vault and a lighter voussoir ring.
	tun := func(grow, depth float64) Solid {
		return w.local(i, Arch{Sill: V(0, w.gateY+0.2-grow, 0), HalfWidth: 2.1 + grow, Spring: 2.9, Depth: depth, Through: AxisZ})
	}
	g.Paint(White, tun(0.8, 2*Vv+1))
	g.Paint(Charcoal, tun(0.6, 2*Vv-1.2))
	g.Carve(tun(0, 2*Vv+2))
	// Name plaque above each portal.
	for _, v := range []float64{-Vv - 0.1, Vv + 0.1} {
		g.Use(Gold).Fill(w.local(i, Block(V(-1.6, top-2.3, v-0.3), V(1.6, top-1.2, v+0.3))))
		g.Use(Charcoal).Fill(w.local(i, Block(V(-1.1, top-2.0, v-0.4), V(1.1, top-1.5, v+0.4))))
	}
	// Crenellated parapets on the front and back edges.
	g.Use(Stone).Fill(w.local(i, Func{Box{V(-U, top, -Vv), V(U, top+2.2, Vv)}, func(p Vec3) bool {
		if math.Abs(p.Z) < Vv-0.9 {
			return false
		}
		return p.Y <= top+1.0 || math.Mod(p.X+U+24, 2.4) < 1.4
	}}))
	// Pavilion: red-walled lower storey, slate skirt eave, upper storey, main roof.
	pu, pv := 5.4, 3.0
	g.Use(Red).Fill(w.local(i, Block(V(-pu, top, -pv), V(pu, top+3.8, pv))))
	g.Carve(w.local(i, Block(V(-pu+0.9, top+0.3, -pv+0.9), V(pu-0.9, top+3.4, pv-0.9))))
	// Lattice window bays between the red columns.
	for b := -2; b <= 2; b++ {
		u := float64(b) * 2.1
		g.Paint(Wood, w.local(i, Block(V(u-0.7, top+0.6, -pv-0.2), V(u+0.7, top+3.0, pv+0.2))))
		g.Paint(Charcoal, w.local(i, Block(V(u-0.35, top+1.2, -pv-0.2), V(u+0.35, top+2.5, pv+0.2))))
	}
	g.Use(Stone).Fill(w.local(i, Block(V(-pu-0.6, top, -pv-0.6), V(pu+0.6, top+0.4, pv+0.6))))
	g.Use(Slate).Fill(w.local(i, Intersect(greatWallRoof(pu+1.4, pv+1.4, top+3.6, 0.55, 0.9),
		Block(V(-20, top, -20), V(20, top+4.6, 20)))))
	qu, qv := 4.2, 2.0
	g.Use(Red).Fill(w.local(i, Block(V(-qu, top+4.0, -qv), V(qu, top+6.8, qv))))
	g.Carve(w.local(i, Block(V(-qu+0.9, top+4.3, -qv+0.9), V(qu-0.9, top+6.4, qv-0.9))))
	for b := -1; b <= 1; b++ {
		u := float64(b) * 2.3
		g.Paint(Charcoal, w.local(i, Block(V(u-0.5, top+4.8, -qv-0.2), V(u+0.5, top+6.0, qv+0.2))))
	}
	g.Use(Slate).Fill(w.local(i, greatWallRoof(qu+1.8, qv+1.8, top+6.5, 0.75, 1.0)))
	ridge := top + 6.5 + (qv+1.8)*0.75
	g.Paint(Charcoal, w.local(i, Block(V(-qu, ridge-1.0, -1), V(qu, ridge+1, 1))))
	g.Use(Charcoal).Fill(w.local(i, Block(V(-2.4, ridge-0.6, -0.5), V(2.4, ridge+0.4, 0.5))))
	for _, u := range []float64{-2.4, 2.4} {
		g.Use(Gold).Fill(w.local(i, Block(V(u-0.5, ridge-0.4, -0.5), V(u+0.5, ridge+1.0, 0.5))))
	}
	// Paved forecourt on the village side with two banner poles.
	g.Use(Stone).Fill(w.local(i, Block(V(-U+1, w.gateY-0.8, Vv), V(U-1, w.gateY+0.2, Vv+6))))
	g.Paint(DarkStone, w.local(i, Func{Box{V(-U+1, w.gateY-0.8, Vv), V(U-1, w.gateY+0.2, Vv+6)}, func(p Vec3) bool {
		return math.Mod(p.X+40, 2.0) < 0.6 || math.Mod(p.Z+40, 2.0) < 0.6
	}}))
	for _, u := range []float64{-5.2, 5.2} {
		g.Use(Wood).Fill(w.local(i, Block(V(u-0.5, w.gateY, Vv+3.5), V(u+0.5, w.gateY+8.5, Vv+4.5))))
		g.Use(Red).Fill(w.local(i, Block(V(u-0.5, w.gateY+5.2, Vv+4.5), V(u+0.5, w.gateY+8.0, Vv+6.6))))
		g.Use(Gold).Fill(w.local(i, Block(V(u-0.5, w.gateY+8.5, Vv+3.5), V(u+0.5, w.gateY+9.3, Vv+4.5))))
	}
}

// village lines the road outside the gate with grey brick courtyard houses.
func (w *greatWallSite) village() {
	g := w.g
	houses := []struct {
		t, off, hu, hv float64
	}{
		{0.25, 5.2, 2.8, 1.9}, {0.45, -5.0, 3.0, 2.0}, {0.62, 5.4, 2.4, 1.8},
		{0.8, -5.2, 2.8, 1.9}, {0.15, -5.4, 2.4, 1.8}, {0.92, 5.4, 2.6, 2.0},
	}
	a, b := greatWallValley[4], greatWallValley[3]
	c := greatWallValley[2]
	for k, h := range houses {
		p0, p1 := a, b
		t := h.t * 2
		if t > 1 {
			p0, p1, t = b, c, t-1
		}
		dir := p1.Sub(p0).Scale(1 / p1.Sub(p0).Length())
		n := P(-dir.Y, dir.X)
		ctr := p0.Add(p1.Sub(p0).Scale(t)).Add(n.Scale(h.off))
		ang := math.Atan2(dir.Y, dir.X)
		loc := func(s Solid) Solid { return RotateY(Translate(s, V(ctr.X, 0, ctr.Y)), ctr, ang) }
		y := w.hm.at(ctr.X, ctr.Y)
		lo := y
		for _, q := range []Vec2{P(-h.hu, -h.hv), P(h.hu, -h.hv), P(-h.hu, h.hv), P(h.hu, h.hv)} {
			q = q.Rotate(ang).Add(ctr)
			lo = math.Min(lo, w.hm.at(q.X, q.Y))
		}
		y = math.Round(y/0.4) * 0.4
		g.Use(DarkStone).Fill(loc(Block(V(-h.hu-0.3, lo-1, -h.hv-0.3), V(h.hu+0.3, y+0.3, h.hv+0.3))))
		g.Use(Stone).Fill(loc(Block(V(-h.hu, y, -h.hv), V(h.hu, y+2.6, h.hv))))
		g.Carve(loc(Block(V(-h.hu+0.9, y+0.3, -h.hv+0.9), V(h.hu-0.9, y+2.6, h.hv-0.9))))
		door := -h.hv
		if h.off < 0 {
			door = h.hv
		}
		g.Use(Red).Fill(loc(Block(V(-0.6, y+0.3, door-0.3), V(0.6, y+2.0, door+0.3))))
		for _, u := range []float64{-h.hu + 1.0, h.hu - 1.0} {
			g.Paint(Charcoal, loc(Block(V(u-0.4, y+1.0, door-0.3), V(u+0.4, y+1.9, door+0.3))))
		}
		roof := Charcoal
		if Hash01(k, 81) < 0.4 {
			roof = Slate
		}
		g.Use(roof).Fill(loc(Subtract(Gable{Min: P(-h.hu-0.6, -h.hv-0.7), Max: P(h.hu+0.6, h.hv+0.7), Base: y + 2.4, Ridge: AxisX},
			Gable{Min: P(-h.hu-0.6, -h.hv+0.3), Max: P(h.hu+0.6, h.hv-0.3), Base: y + 2.4, Ridge: AxisX})))
		g.Paint(Stone, loc(Block(V(-h.hu-0.6, y+2.4, -h.hv-0.7), V(-h.hu+0.2, y+6, h.hv+0.7))))
		g.Paint(Stone, loc(Block(V(h.hu-0.2, y+2.4, -h.hv-0.7), V(h.hu+0.6, y+6, h.hv+0.7))))
	}
}

// terrain fills the heightfield as a hollow skin, coloring road, meadow, rock and
// forest with patches of autumn color.
func (w *greatWallSite) terrain() {
	m := w.hm
	w.eachColumn(-greatWallWorldX, -greatWallWorldZ, greatWallWorldX, greatWallWorldZ, func(cx, cz int, x, z float64) {
		h, ok := w.raw(x, z)
		if !ok {
			return
		}
		h = m.at(x, z)
		low := math.Min(h, m.low(x, z))
		slope := (h - low) / 0.8
		dv := greatWallPolyDist(P(x, z), greatWallValley)
		autumn := greatWallFBM(x/9, z/9, 41)
		w.column(cx, cz, math.Max(0, low-0.6), h, func(cy int, y float64) Color {
			r := Hash01(cx, cy, cz, 43)
			switch {
			case dv < 1.4 && h < 2.6:
				return Sand
			case slope > 1.5+1.2*greatWallNoise(x/6, z/6, 45):
				if greatWallNoise(x/2.5+y/4, z/2.5-y/5, 47) > 0.6 {
					return Stone
				}
				return DarkStone
			case h < 2.4 || dv < 6:
				if r < 0.15 {
					return Foliage
				}
				return Grass
			case autumn > 0.62 && r < 0.45:
				if r < 0.15 {
					return Terracotta
				}
				return Gold
			case greatWallFBM(x/5, z/5, 49) > 0.6:
				return Grass
			}
			return Foliage
		})
	})
}

// trees scatters round canopies over the forested slopes, with autumn color.
func (w *greatWallSite) trees() {
	const step = 2.2
	m := w.hm
	for i := -30; i <= 30; i++ {
		for j := -22; j <= 22; j++ {
			la := float64(i)*step + (Hash01(i, j, 51)-0.5)*1.9
			lb := float64(j)*step + (Hash01(i, j, 52)-0.5)*1.9
			if greatWallRimE(la, lb) > 0.93 || Hash01(i, j, 53) < 0.3 {
				continue
			}
			x, z := greatWallToWorld(la, lb)
			h := m.at(x, z)
			if (h-m.low(x, z))/0.8 > 1.9 || h < 1.8 {
				continue
			}
			if greatWallPolyDist(P(x, z), greatWallValley) < 4.5 {
				continue
			}
			k, dist, _, _ := w.nearest(x, z)
			if dist < 5.2 {
				continue
			}
			near := false
			for _, t := range w.towers {
				if math.Hypot(x-w.px[t.i], z-w.pz[t.i]) < 8 {
					near = true
				}
			}
			if near || math.Hypot(x-w.px[w.gate], z-w.pz[w.gate]) < 16 || k < 0 {
				continue
			}
			r := 0.8 + 0.5*Hash01(i, j, 54)
			c := Foliage
			au := greatWallFBM(x/9, z/9, 41)
			q := Hash01(i, j, 55)
			switch {
			case au > 0.58 && q < 0.75:
				c = []Color{Gold, Terracotta, Red, Gold}[int(Hash01(i, j, 56)*4)]
			case q < 0.2:
				c = Grass
			}
			g := w.g
			g.Use(c).FillVolume(Ellipsoid{Center: V(x, h+r*0.6, z), Radii: V(r, r*1.1, r)})
		}
	}
}
