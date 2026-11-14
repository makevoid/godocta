package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// AlpineVillage is a Swiss mountain village beneath a Matterhorn-like horn. The horn
// rises at the back (-X/-Z): a hollow four-faced pyramid whose ridges are the corners of
// a pinched diamond, with one ridge pointing at the hero camera, a snow-banded face
// toward +Z, a darker rock face toward +X and a summit that hooks over to the left. A
// snow bench rings its foot; a glacier tongue with blue crevasses and dark moraines
// flows off the bench, and its brook crosses the upland, drops over a rock step as a
// waterfall, runs under a stone bridge and into a lake at the front. Scree, alpine
// meadow and a belt of dark spruce with a few golden larches cover the slopes between.
// The village fills the valley floor in front (+X/+Z): chalets with stone ground floors,
// timber upper storeys, flower-box balconies and low stone-weighted roofs, raised
// granaries, a white church with a pointed slate spire on a cobbled square, a sawmill
// with a water wheel, a barn with hay rows, haystacks and cows. A cable car climbs from
// the valley station over three pylons to a station on the shoulder of the front ridge,
// with two red cabins on its ropes. Terrain is a one-cell skin over a sampled height
// field on an elliptical plate. North is -Z.
type AlpineVillage struct{}

func init() { Register(AlpineVillage{}) }

func (AlpineVillage) Name() string  { return "alpine_village" }
func (AlpineVillage) Title() string { return "Alpine Village" }

const (
	alpineVillageA      = 64.0 // Plate semi-axes around alpineVillageCenter.
	alpineVillageB      = 60.0
	alpineVillageSummit = 85.0 // Top of the horn.
	alpineVillageBench  = 33.0 // Snow bench at the foot of its faces.
	alpineVillageHook   = 60.0 // Height above which the summit leans over.
	alpineVillageFloor  = 4.0  // Valley floor.
	alpineVillageLakeY  = 3.2  // Lake surface.
	alpineVillageIceLen = 15.1 // Length of the glacier tongue's centerline.
)

// Kinds of timber building.
const (
	alpineVillageChalet = iota
	alpineVillageBarn
	alpineVillageMazot
)

// Church footprint: nave, then the tower at its west end.
const (
	alpineVillageNaveX0, alpineVillageNaveX1   = 16.5, 30.5
	alpineVillageNaveZ0, alpineVillageNaveZ1   = 12.5, 21.0
	alpineVillageTowerX0, alpineVillageTowerX1 = 11.0, 16.5
	alpineVillageTowerZ0, alpineVillageTowerZ1 = 14.0, 19.5
)

var (
	alpineVillageCenter = P(-3, -4)          // Plate center.
	alpineVillagePeak   = P(-23, -27)        // Summit axis of the horn.
	alpineVillageU      = P(0.6294, 0.7771)  // Front ridge, toward the hero camera.
	alpineVillageV      = P(-0.7771, 0.6294) // Left ridge.
	alpineVillageLean   = P(-0.62, 0.78)     // Direction the summit hooks over.
	alpineVillageLake   = P(7, 48.5)
	alpineVillageFalls  = P(-5.9, 13.8)
	alpineVillageValley = P(41, 0)      // Valley station.
	alpineVillageSaddle = P(-3.8, -3.3) // Mountain station on the front ridge's shoulder.
	alpineVillageSquare = [2]Vec2{P(15, 21.5), P(30, 28.8)}
	alpineVillageHay    = [2]Vec2{P(-17, 44), P(-4.5, 51.5)}
	// The brook from the glacier snout over the falls and through the village to the lake.
	alpineVillageBrook = []Vec2{P(-25, 8.5), P(-20, 9.5), P(-13, 10.8), P(-7.2, 12.2), P(-4.6, 15.4),
		P(-1.5, 20.5), P(0.8, 27), P(1.6, 33.5), P(2.2, 39.5), P(4, 44.5), P(7, 48.5)}
	alpineVillageIce   = []Vec2{P(-24.5, -10), P(-25.5, -2), P(-26, 5)}
	alpineVillagePaths = [][]Vec2{
		{P(6.5, 32.5), P(9, 30.5), P(14, 30), P(44, 30)},
		{P(34.4, 30), P(34.4, 17), P(33, 14), P(32.6, 3.5), P(37, 1.5)},
		{P(-3.5, 32.5), P(-9, 33), P(-9.5, 40.5), P(-15.5, 42.5)},
		{P(22, 30), P(22, 28)},
	}
	alpineVillageHouses = []alpineVillageHouse{
		{P(16, 6), 7.5, 7, AxisZ, alpineVillageChalet},
		{P(27, 6.8), 8, 7, AxisZ, alpineVillageChalet},
		{P(37.5, 10.5), 8.5, 7, AxisX, alpineVillageChalet},
		{P(49, 10), 8, 7.5, AxisZ, alpineVillageChalet},
		{P(51.5, -3), 7.5, 6.5, AxisZ, alpineVillageChalet},
		{P(41, 22), 8.5, 7.5, AxisZ, alpineVillageChalet},
		{P(9, 24.5), 8, 7.5, AxisZ, alpineVillageChalet},
		{P(5, 13), 7.5, 7, AxisX, alpineVillageChalet},
		{P(12.5, 38), 8, 7, AxisZ, alpineVillageChalet},
		{P(22.5, 37), 8.5, 7, AxisX, alpineVillageChalet},
		{P(32.5, 36.5), 8, 7, AxisZ, alpineVillageChalet},
		{P(-15, 35), 8, 7, AxisZ, alpineVillageChalet},
		{P(-13, 26.5), 8, 7, AxisX, alpineVillageChalet},
		{P(25, -13.5), 7, 6.5, AxisZ, alpineVillageChalet},
		{P(-21, 43.5), 10, 8, AxisZ, alpineVillageBarn},
		{P(-6.5, 28), 4, 4, AxisZ, alpineVillageMazot},
		{P(21, 45), 4, 4, AxisZ, alpineVillageMazot},
		{P(31.5, 24.5), 4, 4, AxisZ, alpineVillageMazot},
	}
)

// alpineVillageHouse is one timber building. Its gable front faces +Z when the ridge
// runs along Z and +X when it runs along X.
type alpineVillageHouse struct {
	at    Vec2
	w, d  float64 // Width of the gable front and depth along the ridge.
	ridge Axis
	kind  int
}

func (h alpineVillageHouse) rect() (min, max Vec2) {
	hx, hz := h.w/2, h.d/2
	if h.ridge == AxisX {
		hx, hz = hz, hx
	}
	return P(h.at.X-hx, h.at.Y-hz), P(h.at.X+hx, h.at.Y+hz)
}

// alpineVillageFrame maps a building's local axes (a across the gable front, b toward
// the front) to the world.
type alpineVillageFrame struct {
	c     Vec2
	ridge Axis
}

func (f alpineVillageFrame) at(a, b, y float64) Vec3 {
	if f.ridge == AxisX {
		return V(f.c.X+b, y, f.c.Y-a)
	}
	return V(f.c.X+a, y, f.c.Y+b)
}

func (f alpineVillageFrame) local(p Vec3) (a, b float64) {
	if f.ridge == AxisX {
		return f.c.Y - p.Z, p.X - f.c.X
	}
	return p.X - f.c.X, p.Z - f.c.Y
}

func (f alpineVillageFrame) box(a0, b0, y0, a1, b1, y1 float64) Cuboid {
	return Block(f.at(a0, b0, y0), f.at(a1, b1, y1))
}

// roof is a pitched slab: its top is at eave over the walls (|a| = half), rising pitch
// per unit toward the ridge, overhanging the eaves by over and the gables by lap.
func (f alpineVillageFrame) roof(half, depth, eave, pitch, thick, over, lap float64) Func {
	box := BoxOf(f.at(-half-over, -depth-lap, eave-over*pitch-thick), f.at(half+over, depth+lap, eave+half*pitch))
	return Func{box, func(p Vec3) bool {
		a, _ := f.local(p)
		y := eave + (half-math.Abs(a))*pitch
		return p.Y <= y && p.Y >= y-thick
	}}
}

type alpineVillagePad struct {
	min, max Vec2
	y        float64
}

// alpineVillageLand is the surveyed site: level building pads and the height field
// memoized on the lattice's half-step grid.
type alpineVillageLand struct {
	step   float64
	x0, z0 int
	nx, nz int
	h      []float64
	pads   []alpineVillagePad
}

// col is the height of a lattice column, or -100 outside the plate.
func (l *alpineVillageLand) col(ix, iz int) float64 {
	i, j := ix-l.x0, iz-l.z0
	if i < 0 || j < 0 || i >= l.nx || j >= l.nz {
		return -100
	}
	return l.h[j*l.nx+i]
}

// at is the bilinear height at a world point.
func (l *alpineVillageLand) at(x, z float64) float64 {
	fx, fz := x/l.step-float64(l.x0), z/l.step-float64(l.z0)
	i, j := int(math.Floor(fx)), int(math.Floor(fz))
	if i < 0 || j < 0 || i >= l.nx-1 || j >= l.nz-1 {
		return -100
	}
	tx, tz := fx-float64(i), fz-float64(j)
	k := j*l.nx + i
	a := l.h[k] + (l.h[k+1]-l.h[k])*tx
	b := l.h[k+l.nx] + (l.h[k+l.nx+1]-l.h[k+l.nx])*tx
	return a + (b-a)*tz
}

func (l *alpineVillageLand) slope(x, z float64) float64 {
	return math.Hypot(l.at(x+0.8, z)-l.at(x-0.8, z), l.at(x, z+0.8)-l.at(x, z-0.8)) / 1.6
}

// alpineVillageSite is what the ground painter knows about one lattice column.
type alpineVillageSite struct {
	x, z, h        float64
	slope, aspect  float64 // Gradient size and downhill direction.
	base, delta    float64 // Valley bowl height and distance outside the floor.
	n, u           float64 // Horn norm and distance along the front ridge.
	brook, brookW  float64 // Distance to the brook and its half-width.
	run            float64 // Arc length along the brook.
	ice, iceW      float64
	iceRun         float64
	lake           float64
	path           float64
	wood           float64
	skirt, cobbles bool
	hay            bool
}

func alpineVillageClamp(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

func alpineVillageSmooth(t float64) float64 {
	t = alpineVillageClamp(t, 0, 1)
	return t * t * (3 - 2*t)
}

func alpineVillageLerp(a, b, t float64) float64 { return a + (b-a)*t }

// alpineVillageNoise is smooth value noise in [-1, 1].
func alpineVillageNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return alpineVillageLerp(alpineVillageLerp(a, b, sx), alpineVillageLerp(c, d, sx), sz)*2 - 1
}

func alpineVillageFBM(x, z float64, seed int) float64 {
	return 0.6*alpineVillageNoise(x, z, seed) + 0.3*alpineVillageNoise(x*2.03+17, z*2.03+5, seed+1) +
		0.1*alpineVillageNoise(x*4.1+3, z*4.1+29, seed+2)
}

// alpineVillagePolyline returns the plan distance to a polyline and the arc length at
// the nearest point.
func alpineVillagePolyline(p Vec2, pts []Vec2) (dist, along float64) {
	dist = math.Inf(1)
	run := 0.0
	for i := 0; i+1 < len(pts); i++ {
		a, b := pts[i], pts[i+1]
		ab := b.Sub(a)
		l := ab.Length()
		t := alpineVillageClamp(p.Sub(a).Dot(ab)/(l*l), 0, 1)
		if d := p.Sub(a.Add(ab.Scale(t))).Length(); d < dist {
			dist, along = d, run+t*l
		}
		run += l
	}
	return
}

// alpineVillageNorm measures a plan offset from the summit axis: 1 at the foot of the
// horn, whose four ridges are the corners of a pinched diamond of unequal arms.
func alpineVillageNorm(q Vec2) float64 {
	u, v := q.Dot(alpineVillageU), q.Dot(alpineVillageV)
	a, b := u/30, v/30
	if u < 0 {
		a = -u / 21
	}
	if v < 0 {
		b = -v / 24
	}
	return math.Pow(math.Pow(a, 0.9)+math.Pow(b, 0.9), 1/0.9)
}

// edge is the approximate distance inside the plate's rim; negative outside.
func (AlpineVillage) edge(x, z float64) float64 {
	return (1 - math.Hypot((x-alpineVillageCenter.X)/alpineVillageA, (z-alpineVillageCenter.Y)/alpineVillageB)) * 62
}

func (AlpineVillage) lakeNorm(x, z float64) float64 {
	return math.Hypot((x-alpineVillageLake.X)/11, (z-alpineVillageLake.Y)/4.6)
}

func (AlpineVillage) brookWidth(run, z float64) float64 {
	w := 1.0 + 0.5*alpineVillageSmooth((run-22)/8)
	if run > 42 {
		w += 0.8 * math.Exp(-(z-40.3)*(z-40.3)/5) // Mill pond.
	}
	return w
}

// bowl is the valley: a level floor, a meadow slope, a rock step and the upland behind.
// delta is the distance outside the floor's edge.
func (AlpineVillage) bowl(x, z float64) (y, delta float64) {
	p := P(x, z)
	r, d := -p.Dot(alpineVillageV), p.Dot(alpineVillageU)
	delta = (math.Hypot((r+3)/48, (d-40)/25) - 1) * 30
	floor := alpineVillageFloor - 0.016*alpineVillageClamp(d-18, 0, 40)
	if delta <= 0 {
		return floor, delta
	}
	fall := alpineVillageSmooth(1.4 - p.Sub(alpineVillageFalls).Length()/9)
	delta += 2.2 * alpineVillageFBM(x*0.07, z*0.07, 11) * alpineVillageSmooth(delta/5) * (1 - fall)
	if delta <= 0 {
		return floor, 0
	}
	crag := math.Max(fall, alpineVillageSmooth(0.5+1.6*alpineVillageNoise(x*0.045, z*0.045, 12)))
	m := 4 + 5*crag // Where the meadow slope meets the step.
	switch {
	case delta < m:
		y = floor + 4.5*math.Pow(delta/m, 1.2)
	case delta < 12.5:
		y = floor + 4.5 + 8*(delta-m)/(12.5-m)
	default:
		y = floor + 12.5 + math.Min(0.36*(delta-12.5), 11.5)
	}
	return y, delta
}

// horn is the pyramid's height at offset q from the summit axis, inside its foot.
func (AlpineVillage) horn(x, z float64, q Vec2) float64 {
	profile := func(q Vec2) float64 {
		n := alpineVillageNorm(q)
		if n >= 1 {
			return alpineVillageBench
		}
		// A shoulder steps the left ridge halfway up.
		u, v := q.Dot(alpineVillageU), q.Dot(alpineVillageV)
		shoulder := 6 * math.Exp(-(v-13)*(v-13)/18-u*u/30)
		return alpineVillageBench + (alpineVillageSummit-alpineVillageBench)*math.Pow(1-n, 1.3) + shoulder
	}
	lean := func(y float64) Vec2 {
		t := math.Max(0, (y-alpineVillageHook)/(alpineVillageSummit-alpineVillageHook))
		return alpineVillageLean.Scale(7.5 * t * t)
	}
	y := profile(q)
	if alpineVillageNorm(q) < 0.75 {
		// The summit hooks over: take the highest level whose leaned section holds q.
		for top := alpineVillageSummit; top > alpineVillageHook; top -= 0.5 {
			if profile(q.Sub(lean(top))) < top {
				continue
			}
			lo, hi := top, top+0.5
			for i := 0; i < 8; i++ {
				if mid := (lo + hi) / 2; profile(q.Sub(lean(mid))) >= mid {
					lo = mid
				} else {
					hi = mid
				}
			}
			y = lo
			break
		}
		if y > alpineVillageHook+0.5 && profile(q.Sub(lean(y))) < y-0.6 {
			y = alpineVillageHook
		}
	}
	// Tilted strata: ledges that hold snow between steeper bands.
	fade := alpineVillageSmooth((y-alpineVillageBench-1)/5) * alpineVillageSmooth((alpineVillageSummit-2-y)/5)
	phase := y + 0.3*q.Dot(alpineVillageU) + 1.5*alpineVillageNoise(x*0.07, z*0.07, 8)
	return y - 0.85*fade*math.Sin(phase*2*math.Pi/6.5)
}

// massif is the bowl with the horn, its snow bench and its scree apron.
func (s AlpineVillage) massif(x, z float64) float64 {
	base, delta := s.bowl(x, z)
	rough := alpineVillageSmooth(delta / 8)
	q := P(x, z).Sub(alpineVillagePeak)
	q = q.Add(P(alpineVillageNoise(x*0.12, z*0.12, 3), alpineVillageNoise(x*0.12, z*0.12, 4)).Scale(0.8))
	n := alpineVillageNorm(q)
	bumps := alpineVillageFBM(x*0.14, z*0.14, 7)
	switch {
	case n < 1:
		return s.horn(x, z, q)
	case n < 1.18:
		return alpineVillageBench - 3*(n-1)/0.18 + 0.3*bumps
	case n < 1.78:
		w := math.Pow(1-(n-1.18)/0.6, 1.6)
		return base + math.Max(0, alpineVillageBench-3-base)*w + (0.6*rough+0.3*w)*bumps
	}
	return base + 0.6*rough*bumps
}

// inner is the terrain before the plate's rim cuts it: the massif with the glacier
// tongue and its moraines, the brook's gully and channel, and the lake basin.
func (s AlpineVillage) inner(x, z float64) float64 {
	h := s.massif(x, z)
	p := P(x, z)
	if d, run := alpineVillagePolyline(p, alpineVillageIce); d < 10 {
		w := 6.5 - 2.5*run/alpineVillageIceLen
		h += 1.5 * alpineVillageSmooth((w-d)/2.5)
		if run > 4 {
			h += 0.9 * math.Exp(-(d-w-0.9)*(d-w-0.9)/0.5)
		}
	}
	if d, run := alpineVillagePolyline(p, alpineVillageBrook); d < 7 {
		w := s.brookWidth(run, z)
		h -= 1.4 * alpineVillageSmooth((6-d)/6) * alpineVillageSmooth(run/3) * alpineVillageSmooth((19.5-run)/3)
		h -= alpineVillageSmooth(w + 1 - d)
	}
	if q := s.lakeNorm(x, z); q < 1.3 {
		h = math.Min(h, alpineVillageLakeY-1.0+1.7*alpineVillageSmooth((q-0.87)/0.4))
	}
	return h
}

// rim is the height the plate's cut edge allows.
func (s AlpineVillage) rim(x, z float64) float64 {
	e := s.edge(x, z) + 0.6*alpineVillageNoise(x*0.15, z*0.15, 19)
	return 0.8 + 3.5*math.Max(0, e)
}

// natural is the terrain without building pads; -100 outside the plate.
func (s AlpineVillage) natural(x, z float64) float64 {
	if s.edge(x, z) < 0 {
		return -100
	}
	return math.Min(s.inner(x, z), s.rim(x, z))
}

// survey levels the building pads and samples the height field.
func (s AlpineVillage) survey(step float64) *alpineVillageLand {
	l := &alpineVillageLand{step: step}
	pad := func(min, max Vec2) {
		c := min.Add(max).Scale(0.5)
		l.pads = append(l.pads, alpineVillagePad{min, max, math.Round(s.natural(c.X, c.Y)*2) / 2})
	}
	for _, h := range alpineVillageHouses {
		min, max := h.rect()
		pad(min.Sub(P(1, 1)), max.Add(P(1, 1)))
	}
	pad(P(alpineVillageTowerX0-1, alpineVillageNaveZ0-1), P(alpineVillageNaveX1+4.5, alpineVillageNaveZ1+1))
	pad(alpineVillageValley.Sub(P(5, 4.75)), alpineVillageValley.Add(P(5, 4.75)))
	pad(alpineVillageSaddle.Sub(P(4, 4)), alpineVillageSaddle.Add(P(4, 4)))
	l.x0 = int(math.Floor((alpineVillageCenter.X-alpineVillageA)/step)) - 3
	l.z0 = int(math.Floor((alpineVillageCenter.Y-alpineVillageB)/step)) - 3
	l.nx = int(math.Ceil(2*alpineVillageA/step)) + 7
	l.nz = int(math.Ceil(2*alpineVillageB/step)) + 7
	l.h = make([]float64, l.nx*l.nz)
	for j := 0; j < l.nz; j++ {
		for i := 0; i < l.nx; i++ {
			x, z := float64(l.x0+i)*step, float64(l.z0+j)*step
			h := s.natural(x, z)
			if h >= 0 {
				for _, p := range l.pads {
					dx := math.Max(math.Max(p.min.X-x, x-p.max.X), 0)
					dz := math.Max(math.Max(p.min.Y-z, z-p.max.Y), 0)
					if d := math.Hypot(dx, dz); d < 2.5 {
						h = alpineVillageLerp(p.y, h, alpineVillageSmooth(d/2.5))
					}
				}
			}
			l.h[j*l.nx+i] = h
		}
	}
	return l
}

// busy reports whether a plan point is taken by a building, water, a path, the ice or
// the cable car's swath, with a margin.
func (s AlpineVillage) busy(l *alpineVillageLand, p Vec2, margin float64) bool {
	for _, q := range l.pads {
		if p.X > q.min.X-margin && p.X < q.max.X+margin && p.Y > q.min.Y-margin && p.Y < q.max.Y+margin {
			return true
		}
	}
	if d, run := alpineVillagePolyline(p, alpineVillageBrook); d < s.brookWidth(run, p.Y)+1+margin {
		return true
	}
	if d, _ := alpineVillagePolyline(p, alpineVillageIce); d < 7.5+margin {
		return true
	}
	if d, _ := alpineVillagePolyline(p, []Vec2{alpineVillageValley, alpineVillageSaddle}); d < 3.2+margin {
		return true
	}
	for _, path := range alpineVillagePaths {
		if d, _ := alpineVillagePolyline(p, path); d < 1+margin {
			return true
		}
	}
	sq := alpineVillageSquare
	if p.X > sq[0].X-margin && p.X < sq[1].X+margin && p.Y > sq[0].Y-margin && p.Y < sq[1].Y+margin {
		return true
	}
	if p.X > -9-margin && p.X < 1+margin && p.Y > 34-margin && p.Y < 44.5+margin {
		return true // Sawmill yard.
	}
	return s.lakeNorm(p.X, p.Y) < 1.15+margin*0.15
}

// wooded is the density of the spruce belt at a plan point, 0 to 1.
func (s AlpineVillage) wooded(l *alpineVillageLand, x, z float64) float64 {
	h := l.at(x, z)
	if h < 3.4 || h > 25.5 {
		return 0
	}
	_, delta := s.bowl(x, z)
	belt := alpineVillageSmooth((delta-2)/3) * alpineVillageSmooth((31-delta)/5)
	// The hay meadows on the left bank and the pasture of the upper alp stay open.
	belt *= alpineVillageSmooth((P(x, z).Sub(P(-15, 40)).Length() - 13) / 4)
	belt *= alpineVillageSmooth((P(x, z).Sub(P(24, -15)).Length() - 7) / 4)
	belt *= alpineVillageSmooth((alpineVillageNorm(P(x, z).Sub(alpineVillagePeak)) - 1.3) / 0.25)
	return belt * alpineVillageClamp(0.75+0.9*alpineVillageFBM(x*0.06, z*0.06, 31), 0, 1)
}

func (s AlpineVillage) Build(g *Grid) Order {
	land := s.survey(g.Span() / 2)
	s.terrain(g, land)
	s.lake(g, land)
	s.forest(g, land)
	s.boulders(g, land)
	s.church(g, land)
	for k, h := range alpineVillageHouses {
		switch h.kind {
		case alpineVillageBarn:
			s.barn(g, land, h)
		case alpineVillageMazot:
			s.mazot(g, land, h)
		default:
			s.chalet(g, land, h, k)
		}
	}
	s.sawmill(g, land)
	s.bridge(g, land)
	s.square(g, land)
	s.cableway(g, land)
	s.pasture(g, land)
	return Outward(alpineVillagePeak)
}

// terrain adds the height field as a skin: only cells with an empty face neighbor are
// kept (space below y = 0 counts as filled), so the mountain costs its surface.
func (s AlpineVillage) terrain(g *Grid, l *alpineVillageLand) {
	step := l.step
	occ := func(ix, iy, iz int) bool { return iy < 0 || float64(iy)*step <= l.col(ix, iz) }
	around := [][2]int{{2, 0}, {-2, 0}, {0, 2}, {0, -2}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	for iz := l.z0; iz < l.z0+l.nz; iz++ {
		for ix := l.x0; ix < l.x0+l.nx; ix++ {
			if (ix^iz)&1 != 0 {
				continue
			}
			h := l.col(ix, iz)
			if h < 0 {
				continue
			}
			lo := h
			for _, o := range around {
				lo = math.Min(lo, l.col(ix+o[0], iz+o[1]))
			}
			iy := int(math.Floor(lo/step)) - 3
			if iy < 0 {
				iy = 0
			}
			if (iy^ix)&1 != 0 {
				iy++
			}
			site := s.site(l, ix, iz)
			for ; float64(iy)*step <= h; iy += 2 {
				exposed := false
				for _, d := range Neighbors {
					if !occ(ix+int(d.X), iy+int(d.Y), iz+int(d.Z)) {
						exposed = true
						break
					}
				}
				if exposed {
					g.Use(s.ground(&site, float64(iy)*step, ix, iy, iz))
					g.Add(Cell{X: int32(ix), Y: int32(iy), Z: int32(iz)})
				}
			}
		}
	}
}

// site gathers what the painter needs to know about a column.
func (s AlpineVillage) site(l *alpineVillageLand, ix, iz int) alpineVillageSite {
	x, z := float64(ix)*l.step, float64(iz)*l.step
	p := P(x, z)
	c := alpineVillageSite{x: x, z: z, h: l.col(ix, iz)}
	near := func(dx, dz int) float64 {
		if h := l.col(ix+dx, iz+dz); h >= 0 {
			return h
		}
		return c.h
	}
	gx := (near(2, 0) - near(-2, 0)) / (4 * l.step)
	gz := (near(0, 2) - near(0, -2)) / (4 * l.step)
	c.slope, c.aspect = math.Hypot(gx, gz), math.Atan2(-gz, -gx)
	c.base, c.delta = s.bowl(x, z)
	q := p.Sub(alpineVillagePeak)
	c.n, c.u = alpineVillageNorm(q), q.Dot(alpineVillageU)
	c.brook, c.run = alpineVillagePolyline(p, alpineVillageBrook)
	c.brookW = s.brookWidth(c.run, z)
	c.ice, c.iceRun = alpineVillagePolyline(p, alpineVillageIce)
	c.iceW = 6.5 - 2.5*c.iceRun/alpineVillageIceLen
	c.lake = s.lakeNorm(x, z)
	rim := s.rim(x, z)
	c.skirt = c.h <= rim+0.3 && s.inner(x, z) > rim+0.5
	c.path = math.Inf(1)
	for _, path := range alpineVillagePaths {
		d, _ := alpineVillagePolyline(p, path)
		c.path = math.Min(c.path, d)
	}
	sq, hay := alpineVillageSquare, alpineVillageHay
	c.cobbles = x > sq[0].X && x < sq[1].X && z > sq[0].Y && z < sq[1].Y
	c.hay = x > hay[0].X && x < hay[1].X && z > hay[0].Y && z < hay[1].Y
	c.wood = s.wooded(l, x, z)
	return c
}

// ground picks the color of one exposed terrain cell.
func (s AlpineVillage) ground(c *alpineVillageSite, y float64, ix, iy, iz int) Color {
	r := Hash01(ix, iy, iz, 41)
	depth := c.h - y
	top := depth < 0.8
	pick := func(a Color, share float64, b Color) Color {
		if r < share {
			return a
		}
		return b
	}
	strata := func(period float64, seed int) int {
		return int(math.Floor((y + 0.3*c.u + 1.6*alpineVillageFBM(c.x*0.08, c.z*0.08, seed)) / period))
	}
	// Pale valley rock for the rock step, outcrops and the plate's cut edge.
	crag := func() Color {
		if Hash01(strata(2.2, 51), 52) < 0.5 {
			return pick(Stone, 0.9, DarkStone)
		}
		return pick(DarkStone, 0.92, Stone)
	}
	// Dark banded gneiss for the horn.
	rock := func(dark float64) Color {
		band := strata(2.1, 21)
		k := Hash01(band, 23)
		if r < 0.14 {
			k = Hash01(band+1, 23)
		}
		switch k -= dark; {
		case k < 0.14:
			return Charcoal
		case k < 0.5:
			return Slate
		}
		return DarkStone
	}
	if c.skirt || depth > 1.5+1.4*c.slope {
		return crag()
	}
	if c.lake < 1.15 && c.h < alpineVillageLakeY+0.3 {
		return pick(Sand, 0.7, Stone)
	}
	// The brook: white water where it runs steep, gravel banks.
	if c.run > 0.3 && c.brook < c.brookW+0.25 {
		if !top {
			return DarkStone
		}
		wild := alpineVillageSmooth((c.slope - 0.3) / 0.9)
		switch {
		case r < 0.1+0.8*wild:
			return White
		case r > 0.9:
			return Glass
		}
		return Water
	}
	if c.run > 0.3 && c.brook < c.brookW+1.2 && top && r < 0.7 {
		return pick(Stone, 0.42, DarkStone)
	}
	// Glacier tongue with transverse crevasses, a blue snout and dark moraines.
	if c.ice < c.iceW {
		if c.iceRun > alpineVillageIceLen-0.3 && c.ice > c.iceW-1.8 {
			return pick(Glass, 0.7, White)
		}
		k := c.iceRun + 0.8*alpineVillageNoise(c.x*0.25, c.z*0.25, 56)
		if c.ice < c.iceW-1 && math.Mod(k+50, 2.7) < 0.8 &&
			Hash01(int(math.Floor(k/2.7)), int(math.Floor(c.x/3.5)), 57) < 0.8 {
			return Glass
		}
		return White
	}
	if c.ice < c.iceW+2 && c.iceRun > 4 && c.n > 1.1 {
		return pick(DarkStone, 0.6, Stone)
	}
	// The horn: snow wherever a face lies back far enough to hold it.
	if c.n < 1.18 {
		if c.n > 0.97 {
			return White
		}
		face := math.Cos(c.aspect-1.68) - math.Cos(c.aspect-0.10) // 1 on the +Z face, -1 on the +X face.
		limit := 2.4 + 0.9*face + 0.014*(y-55) + 0.45*alpineVillageFBM(c.x*0.18, c.z*0.18, 61)
		// Rock ribs stand out of the snow and couloirs hold it on the rock, both
		// running down the fall line.
		around := math.Atan2(c.z-alpineVillagePeak.Y, c.x-alpineVillagePeak.X)
		rib := alpineVillageNoise(around*7, y*0.03, 66)
		if c.slope < limit {
			if rib > 0.45 && c.n < 0.9 {
				return rock(0.12 - 0.13*face)
			}
			return White
		}
		if rib < -0.62 {
			return White
		}
		return rock(0.12 - 0.13*face)
	}
	patch := alpineVillageFBM(c.x*0.16, c.z*0.16, 63)
	// Scree apron under the bench, with snow patches at its top.
	if c.n < 1.78 && c.h-c.base > 1 {
		if c.slope > 2.1 {
			return rock(0.1)
		}
		if y > 27.5+2.5*patch {
			return White
		}
		if (c.n-1.18)/0.6 < 0.55+0.35*patch {
			return pick(Stone, 0.7, DarkStone)
		}
	}
	if c.slope > 1.75 {
		return crag()
	}
	if y > 28+2*patch {
		return White
	}
	if y > 24.5+2*patch {
		return pick(Stone, 0.7, DarkStone)
	}
	if top && c.cobbles {
		if (int(math.Floor(c.x/1.6))+int(math.Floor(c.z/1.6)))%2 == 0 {
			return Stone
		}
		return DarkStone
	}
	if top && c.path < 1.25 {
		return Sand
	}
	if c.wood > 0.3 {
		return pick(Foliage, 0.82, Grass)
	}
	if top && c.hay && math.Mod(c.z+100, 2.4) < 0.8 {
		return Gold // Hay rows drying on the mown meadow.
	}
	if alpineVillageFBM(c.x*0.11, c.z*0.11, 71) > 0.3 {
		return Foliage
	}
	return Grass
}

// lake lays the water sheet and its ripples, a jetty and a rowing boat.
func (s AlpineVillage) lake(g *Grid, l *alpineVillageLand) {
	c := alpineVillageLake
	box := Box{V(c.X-13, alpineVillageLakeY-0.8, c.Y-6), V(c.X+13, alpineVillageLakeY, c.Y+6)}
	wet := func(p Vec3) bool { return s.lakeNorm(p.X, p.Z) < 1.2 && l.at(p.X, p.Z) < alpineVillageLakeY-0.05 }
	g.Use(Water).Fill(Func{box, wet})
	g.Paint(Glass, Func{box, func(p Vec3) bool {
		streak := alpineVillageNoise(p.X*0.18, p.Z*0.9, 75)
		return wet(p) && streak > 0.45 && s.lakeNorm(p.X, p.Z) < 0.95
	}})
	// Jetty from the village shore.
	top := alpineVillageLakeY + 0.7
	g.Use(Wood).Fill(Block(V(10.2, top-0.5, 42.6), V(11.8, top, 47.6)))
	for _, z := range []float64{44.6, 47.2} {
		g.Fill(Block(V(10.1, alpineVillageLakeY-1, z-0.4), V(10.9, top, z+0.4)), Block(V(11.1, alpineVillageLakeY-1, z-0.4), V(11.9, top, z+0.4)))
	}
	// Rowing boat.
	b := V(14.2, alpineVillageLakeY+0.3, 47.8)
	hull := Intersect(Ellipsoid{b, V(2.4, 1.1, 1.2)}, Block(V(b.X-3, b.Y-1, b.Z-2), V(b.X+3, b.Y+0.5, b.Z+2)))
	g.Use(Red).FillVolume(hull)
	g.Paint(White, Block(V(b.X-3, b.Y+0.1, b.Z-2), V(b.X+3, b.Y+0.6, b.Z+2)))
	g.Paint(Wood, Ellipsoid{b.Add(V(0, 0.5, 0)), V(1.5, 0.6, 0.6)})
}

// spruce is a trunk with three stacked cones.
func (AlpineVillage) spruce(g *Grid, at Vec2, base, h, r float64, c Color) {
	g.Use(Wood).Fill(Capsule{at.At(base - 0.8), at.At(base + 1.6), 0.45})
	g.Use(c)
	for k := 0; k < 3; k++ {
		fk := float64(k)
		g.FillVolume(Cone(at, r*(1-0.25*fk), base+1.1+fk*h*0.25, h*(0.58-0.09*fk)))
	}
}

// forest plants the spruce belt, with larches near the tree line and loose trees on
// the valley floor and the upper alp.
func (s AlpineVillage) forest(g *Grid, l *alpineVillageLand) {
	const pitch = 3.1
	for i := -23; i <= 21; i++ {
		for j := -22; j <= 19; j++ {
			at := P((float64(i)+0.15+0.7*Hash01(i, j, 81))*pitch, (float64(j)+0.15+0.7*Hash01(i, j, 82))*pitch)
			if s.edge(at.X, at.Y) < 4 {
				continue
			}
			h := l.at(at.X, at.Y)
			chance := s.wooded(l, at.X, at.Y)
			_, delta := s.bowl(at.X, at.Y)
			if chance < 0.1 && h > 3.4 && h < 24 {
				chance = 0.07
				if delta > 12 {
					chance = 0.1
				}
			}
			if Hash01(i, j, 83) >= chance || l.slope(at.X, at.Y) > 1.9 || s.busy(l, at, 2.2) {
				continue
			}
			if alpineVillageNorm(at.Sub(alpineVillagePeak)) < 1.35 {
				continue
			}
			c := Foliage
			if delta > 15 && Hash01(i, j, 85) < 0.12 {
				c = Gold
			} else if Hash01(i, j, 85) > 0.9 {
				c = Grass
			}
			s.spruce(g, at, h-0.3, 6.5+4*Hash01(i, j, 84), 1.7+0.5*Hash01(i, j, 86), c)
		}
	}
}

// boulders scatters erratic blocks over the scree and the upland.
func (s AlpineVillage) boulders(g *Grid, l *alpineVillageLand) {
	for i := -9; i <= 8; i++ {
		for j := -9; j <= 5; j++ {
			at := P((float64(i)+Hash01(i, j, 87))*6.5, (float64(j)+Hash01(i, j, 88))*6.5)
			n := alpineVillageNorm(at.Sub(alpineVillagePeak))
			_, delta := s.bowl(at.X, at.Y)
			if s.edge(at.X, at.Y) < 5 || n < 1.2 || delta < 13 || Hash01(i, j, 89) > 0.45 || s.busy(l, at, 1.5) {
				continue
			}
			h := l.at(at.X, at.Y)
			if h > 24 {
				continue // Nothing dots the snowfields.
			}
			size := 0.9 + 0.9*Hash01(i, j, 90)
			c := Stone
			if Hash01(i, j, 91) < 0.4 {
				c = DarkStone
			}
			g.Use(c).FillVolume(Ellipsoid{at.At(h + size*0.3), V(size*1.2, size, size)})
		}
	}
}

// chalet is a timber house: stone ground floor, dark timber storey with a flower-box
// balcony across the gable front, shutters, and a low roof weighted with stones.
func (s AlpineVillage) chalet(g *Grid, l *alpineVillageLand, h alpineVillageHouse, k int) {
	f := alpineVillageFrame{h.at, h.ridge}
	y0 := l.at(h.at.X, h.at.Y)
	hw, hd := h.w/2, h.d/2
	const stone, wood, pitch = 2.7, 5.7, 0.5
	m := NewGridWithSpan(g.Span())
	base, roofing, shutter := Stone, Slate, Red
	if Hash01(k, 101) < 0.3 {
		base = White
	}
	switch v := Hash01(k, 102); {
	case v < 0.18:
		roofing = DarkStone
	case v < 0.48:
		roofing = Charcoal
	}
	if Hash01(k, 103) < 0.4 {
		shutter = Foliage
	}
	inner := f.box(-hw+1, -hd+1, y0-3, hw-1, hd-1, y0+wood+0.4)
	m.Use(base).Fill(Subtract(f.box(-hw, -hd, y0-0.8, hw, hd, y0+stone), inner))
	m.Use(Wood).Fill(Subtract(f.box(-hw, -hd, y0+stone, hw, hd, y0+wood), inner))
	eave := y0 + wood + 0.5
	m.Fill(Func{Box(f.box(-hw, -hd, y0+wood, hw, hd, eave+hw*pitch)), func(p Vec3) bool {
		a, b := f.local(p)
		return math.Abs(b) >= hd-1 && p.Y <= eave+(hw-math.Abs(a))*pitch-0.4
	}})
	front := func(c Color, a0, y1, a1, y2 float64) { m.Paint(c, f.box(a0, hd-0.6, y1, a1, hd+0.1, y2)) }
	flank := func(c Color, sgn, b0, y1, b1, y2 float64) {
		m.Paint(c, f.box(sgn*(hw-0.6), b0, y1, sgn*(hw+0.1), b1, y2))
	}
	front(Charcoal, -0.6, y0, 0.6, y0+2.1)
	front(Glass, -0.5, y0+wood+0.3, 0.5, y0+wood+1.2)
	for _, sgn := range []float64{-1, 1} {
		a := sgn * hw * 0.52
		front(Glass, a-0.5, y0+1, a+0.5, y0+2)
		front(Glass, a-0.55, y0+3.6, a+0.55, y0+4.8)
		front(shutter, a-1.15, y0+3.6, a-0.6, y0+4.8)
		front(shutter, a+0.6, y0+3.6, a+1.15, y0+4.8)
		for _, side := range []float64{-1, 1} {
			b := side * hd * 0.42
			flank(Charcoal, sgn, b-0.5, y0+1, b+0.5, y0+2)
			flank(Glass, sgn, b-0.5, y0+3.6, b+0.5, y0+4.8)
			flank(shutter, sgn, b-1.1, y0+3.6, b-0.55, y0+4.8)
			flank(shutter, sgn, b+0.55, y0+3.6, b+1.1, y0+4.8)
		}
	}
	// Balcony with a board balustrade and geraniums along its rail.
	deck := y0 + stone
	m.Use(Wood).Fill(f.box(-hw-0.4, hd, deck-0.3, hw+0.4, hd+1.5, deck+0.2),
		f.box(-hw-0.4, hd+1, deck+0.2, hw+0.4, hd+1.5, deck+1.3),
		f.box(-hw-0.4, hd, deck+0.2, -hw+0.1, hd+1.5, deck+1.3),
		f.box(hw-0.1, hd, deck+0.2, hw+0.4, hd+1.5, deck+1.3),
		f.box(-hw-0.35, hd+0.8, y0-0.5, -hw+0.35, hd+1.5, deck),
		f.box(hw-0.35, hd+0.8, y0-0.5, hw+0.35, hd+1.5, deck))
	m.Paint(Red, f.box(-hw-0.5, hd+0.9, deck+0.85, hw+0.5, hd+1.6, deck+1.4))
	s.slabs(m, f, hw, hd, eave, roofing)
	// Chimney.
	ca, cb := hw*0.45, -hd*0.35
	top := eave + (hw-ca)*pitch + 1.6
	m.Use(Stone).Fill(f.box(ca-0.55, cb-0.55, eave, ca+0.55, cb+0.55, top))
	m.Paint(Charcoal, f.box(ca-0.6, cb-0.6, top-0.4, ca+0.6, cb+0.6, top+0.1))
	g.Merge(m)
}

// slabs is the low stone-slab roof with timber verges and rows of weighting stones.
func (AlpineVillage) slabs(m *Grid, f alpineVillageFrame, hw, hd, eave float64, roofing Color) {
	const pitch = 0.5
	roof := f.roof(hw, hd, eave, pitch, 0.9, 1.4, 1.3)
	m.Use(roofing).Fill(roof)
	m.Paint(Wood, Intersect(roof, Func{roof.Box, func(p Vec3) bool {
		_, b := f.local(p)
		return math.Abs(b) > hd+0.75
	}}))
	m.Use(Stone)
	for _, sgn := range []float64{-1, 1} {
		for row, a := range []float64{hw * 0.45, hw + 0.7} {
			for b := -hd + 1.3*float64(row); b <= hd+0.5; b += 2.6 {
				m.Fill(Sphere(f.at(sgn*a, b, eave+(hw-a)*pitch+0.35), 0.5))
			}
		}
	}
}

// barn is a big hay barn: stone footing, plank walls, wagon door and hay loft.
func (s AlpineVillage) barn(g *Grid, l *alpineVillageLand, h alpineVillageHouse) {
	f := alpineVillageFrame{h.at, h.ridge}
	y0 := l.at(h.at.X, h.at.Y)
	hw, hd := h.w/2, h.d/2
	const stone, wood, pitch = 1.4, 5.2, 0.5
	m := NewGridWithSpan(g.Span())
	inner := f.box(-hw+1, -hd+1, y0-3, hw-1, hd-1, y0+wood+0.4)
	m.Use(Stone).Fill(Subtract(f.box(-hw, -hd, y0-0.8, hw, hd, y0+stone), inner))
	m.Use(Wood).Fill(Subtract(f.box(-hw, -hd, y0+stone, hw, hd, y0+wood), inner))
	eave := y0 + wood + 0.5
	m.Fill(Func{Box(f.box(-hw, -hd, y0+wood, hw, hd, eave+hw*pitch)), func(p Vec3) bool {
		a, b := f.local(p)
		return math.Abs(b) >= hd-1 && p.Y <= eave+(hw-math.Abs(a))*pitch-0.4
	}})
	m.Paint(Charcoal, f.box(-1.5, hd-0.6, y0, 1.5, hd+0.1, y0+3.3))
	m.Paint(Gold, f.box(-1, hd-0.6, y0+4.2, 1, hd+0.1, y0+wood+0.9))
	for _, sgn := range []float64{-1, 1} {
		for _, b := range []float64{-hd * 0.45, hd * 0.45} {
			m.Paint(Charcoal, f.box(sgn*(hw-0.6), b-0.5, y0+2.6, sgn*(hw+0.1), b+0.5, y0+3.6))
		}
	}
	s.slabs(m, f, hw, hd, eave, Charcoal)
	g.Merge(m)
}

// mazot is a small granary raised on stilts capped with flat stones.
func (s AlpineVillage) mazot(g *Grid, l *alpineVillageLand, h alpineVillageHouse) {
	f := alpineVillageFrame{h.at, h.ridge}
	y0 := l.at(h.at.X, h.at.Y)
	hw, hd := h.w/2, h.d/2
	const pitch = 0.5
	m := NewGridWithSpan(g.Span())
	for _, sa := range []float64{-1, 1} {
		for _, sb := range []float64{-1, 1} {
			a, b := sa*(hw-0.6), sb*(hd-0.6)
			m.Use(Wood).Fill(f.box(a-0.45, b-0.45, y0-0.6, a+0.45, b+0.45, y0+1))
			m.Use(Stone).Fill(f.box(a-0.85, b-0.85, y0+1, a+0.85, b+0.85, y0+1.45))
		}
	}
	floor, top := y0+1.45, y0+4.2
	m.Use(Wood).Fill(Subtract(f.box(-hw, -hd, floor, hw, hd, top), f.box(-hw+1, -hd+1, floor+0.6, hw-1, hd-1, top+0.4)))
	m.Paint(Charcoal, f.box(-hw-0.1, -hd-0.1, floor+1.2, hw+0.1, hd+0.1, floor+1.6))
	m.Paint(Charcoal, f.box(-0.5, hd-0.6, floor+0.2, 0.5, hd+0.1, floor+1.9))
	eave := top + 0.5
	m.Use(Wood).Fill(Func{Box(f.box(-hw, -hd, top, hw, hd, eave+hw*pitch)), func(p Vec3) bool {
		a, b := f.local(p)
		return math.Abs(b) >= hd-1 && p.Y <= eave+(hw-math.Abs(a))*pitch-0.4
	}})
	roof := f.roof(hw, hd, eave, pitch, 0.9, 1, 0.9)
	m.Use(Slate).Fill(roof)
	m.Use(Stone)
	for _, sgn := range []float64{-1, 1} {
		for b := -hd + 0.6; b <= hd; b += 2.8 {
			m.Fill(Sphere(f.at(sgn*hw*0.6, b, eave+hw*0.4*pitch+0.35), 0.5))
		}
	}
	g.Merge(m)
}

// church is the white parish church: nave with a steep slate roof and round apse, and
// a west tower with clock faces, belfry and a tall pointed spire.
func (s AlpineVillage) church(g *Grid, l *alpineVillageLand) {
	const x0, x1, z0, z1 = alpineVillageNaveX0, alpineVillageNaveX1, alpineVillageNaveZ0, alpineVillageNaveZ1
	const tx0, tx1, tz0, tz1 = alpineVillageTowerX0, alpineVillageTowerX1, alpineVillageTowerZ0, alpineVillageTowerZ1
	const pitch = 1.15
	zc, half := (z0+z1)/2, (z1-z0)/2
	y0 := l.at((x0+x1)/2, zc)
	wallTop, towerTop := y0+7.5, y0+16
	m := NewGridWithSpan(g.Span())
	roofY := func(z float64) float64 { return wallTop + 0.5 + (half-math.Abs(z-zc))*pitch }
	// Nave walls, gables and apse.
	m.Use(White).Fill(Subtract(Block(V(x0, y0-0.8, z0), V(x1, wallTop, z1)), Block(V(x0+1, y0-3, z0+1), V(x1-1, wallTop+0.4, z1-1))))
	m.Fill(Func{Box{V(x0, wallTop, z0), V(x1, wallTop+half*pitch, z1)}, func(p Vec3) bool {
		return (p.X <= x0+1 || p.X >= x1-1) && p.Y <= roofY(p.Z)-0.6
	}})
	m.FillVolume(Subtract(Cylinder{Center: P(x1, zc), Radius: 3.4, Y0: y0 - 0.8, Y1: y0 + 6.5},
		Cylinder{Center: P(x1, zc), Radius: 2.4, Y0: y0 - 3, Y1: y0 + 7}, Block(V(x0, y0-3, z0), V(x1-0.05, y0+8, z1))))
	m.Use(Slate).Fill(Func{Box{V(x0-0.7, wallTop-1.6, z0-0.8), V(x1+0.4, wallTop+1+half*pitch, z1+0.8)}, func(p Vec3) bool {
		y := roofY(p.Z)
		return p.Y <= y && p.Y >= y-1.5
	}})
	m.FillVolume(Intersect(Cone(P(x1, zc), 4, y0+6.5, 4.4), Block(V(x1, y0+6, zc-5), V(x1+5, y0+12, zc+5))))
	// Tall windows and a stone plinth.
	for _, x := range []float64{x0 + 2.8, x0 + 7, x0 + 11.2} {
		m.Paint(Navy, Block(V(x-0.6, y0+2.4, z1-0.6), V(x+0.6, y0+5.8, z1+0.1)), Block(V(x-0.6, y0+2.4, z0-0.1), V(x+0.6, y0+5.8, z0+0.6)))
	}
	m.Paint(Navy, Block(V(x1+2.6, y0+2.6, zc-0.6), V(x1+3.5, y0+5, zc+0.6)))
	m.Paint(Stone, Block(V(tx0-1, y0-1, z0-1), V(x1+4, y0+0.5, z1+1)))
	// Tower.
	tc := P((tx0+tx1)/2, zc)
	m.Use(White).Fill(Subtract(Block(V(tx0, y0-0.8, tz0), V(tx1, towerTop, tz1)), Block(V(tx0+1, y0-3, tz0+1), V(tx1-1, towerTop-1, tz1-1))))
	m.Paint(Stone, Block(V(tx0-1, y0-1, tz0-1), V(tx1, y0+0.5, tz1+1)))
	m.Paint(Wood, Block(V(tx0-0.1, y0+0.5, zc-0.8), V(tx0+0.6, y0+3.2, zc+0.8)))
	face := func(c Color, w, y1, y2 float64) {
		m.Paint(c, Block(V(tc.X-w, y1, tz1-0.6), V(tc.X+w, y2, tz1+0.1)), Block(V(tc.X-w, y1, tz0-0.1), V(tc.X+w, y2, tz0+0.6)),
			Block(V(tx0-0.1, y1, zc-w), V(tx0+0.6, y2, zc+w)), Block(V(tx1-0.6, y1, zc-w), V(tx1+0.1, y2, zc+w)))
	}
	face(Charcoal, 1.1, y0+12.6, y0+14.9)
	face(Gold, 0.9, y0+9.4, y0+11.2)
	face(Charcoal, 0.3, y0+10, y0+10.6)
	m.Use(Stone).Fill(Block(V(tx0-0.4, towerTop, tz0-0.4), V(tx1+0.4, towerTop+0.6, tz1+0.4)))
	// Spire, hollow in its lower half, with a gilded ball and cross.
	const spire = 12.5
	base := towerTop + 0.6
	needle := func(inset float64) Func {
		return Func{Box{V(tc.X-3.1, base, tc.Y-3.1), V(tc.X+3.1, base+spire, tc.Y+3.1)}, func(p Vec3) bool {
			r := 3.0*math.Pow(1-(p.Y-base)/spire, 1.25) - inset
			if inset == 0 {
				r = math.Max(r, 0.5)
			}
			return math.Abs(p.X-tc.X) <= r && math.Abs(p.Z-tc.Y) <= r
		}}
	}
	m.Use(Slate).FillVolume(needle(0))
	m.Carve(needle(1.1))
	m.Use(Gold).FillVolume(Sphere(tc.At(base+spire+0.3), 0.8))
	m.Fill(Block(V(tc.X-0.45, base+spire+0.8, tc.Y-0.45), V(tc.X+0.45, base+spire+3.7, tc.Y+0.45)),
		Block(V(tc.X-1.3, base+spire+2.2, tc.Y-0.45), V(tc.X+1.3, base+spire+3, tc.Y+0.45)))
	g.Merge(m)
}

// square furnishes the cobbled church square: a fountain and a flag.
func (s AlpineVillage) square(g *Grid, l *alpineVillageLand) {
	c := P(22, 25)
	y := l.at(c.X, c.Y)
	g.Use(Stone).FillVolume(Cylinder{Center: c, Radius: 1.9, Inner: 1.0, Y0: y - 0.3, Y1: y + 1})
	g.Use(Water).Fill(Cylinder{Center: c, Radius: 1.1, Y0: y, Y1: y + 0.6})
	g.Use(Stone).Fill(Cylinder{Center: c, Radius: 0.5, Y0: y, Y1: y + 2.2})
	s.flag(g, V(27.5, l.at(27.5, 23.5), 23.5), 7)
}

// flag is a pole flying the red flag with its white cross.
func (AlpineVillage) flag(g *Grid, at Vec3, h float64) {
	g.Use(Stone).Fill(Capsule{at.Add(V(0, -0.5, 0)), at.Add(V(0, h, 0)), 0.5})
	g.Use(Red).Fill(Block(at.Add(V(0.4, h-2.2, -0.3)), at.Add(V(3.0, h-0.2, 0.3))))
	g.Paint(White, Block(at.Add(V(1.3, h-1.6, -0.5)), at.Add(V(2.1, h-0.8, 0.5))))
}

// bridge is the humpbacked stone bridge carrying the village street over the brook.
func (s AlpineVillage) bridge(g *Grid, l *alpineVillageLand) {
	const xc, zc, half, wide = 1.5, 32.5, 5.0, 2.0
	y0 := l.at(xc+half+1, zc)
	deck := func(x float64) float64 {
		t := (x - xc) / half
		return y0 + 0.5 + 2.2*(1-t*t)
	}
	m := NewGridWithSpan(g.Span())
	box := Box{V(xc-half, y0-1.6, zc-wide), V(xc+half, y0+4.2, zc+wide)}
	m.Use(Stone).Fill(Func{box, func(p Vec3) bool {
		if math.Abs(p.Z-zc) > wide-0.8 {
			return p.Y <= deck(p.X)+1.1 // Parapets.
		}
		return p.Y <= deck(p.X)
	}})
	m.Carve(Arch{Sill: V(xc, y0-1.7, zc), HalfWidth: 2.4, Spring: 1, Depth: 2*wide + 1, Through: AxisZ})
	m.Paint(DarkStone, Func{box, func(p Vec3) bool {
		return Hash01(int(math.Floor(p.X/1.1)), int(math.Floor(p.Y/0.8)), int(math.Floor(p.Z/1.1)), 93) < 0.3
	}})
	m.Paint(Sand, Func{box, func(p Vec3) bool { return math.Abs(p.Z-zc) <= wide-0.8 && p.Y > deck(p.X)-0.5 }})
	g.Merge(m)
}

// sawmill is an open-sided shed on the brook with an undershot water wheel, a log on
// the saw carriage and a log pile.
func (s AlpineVillage) sawmill(g *Grid, l *alpineVillageLand) {
	const x0, x1, z0, z1 = -7.5, -0.5, 37.0, 43.5
	c := P((x0+x1)/2, (z0+z1)/2)
	y0 := l.at(c.X, c.Y)
	f := alpineVillageFrame{c, AxisZ}
	hw, hd := (x1-x0)/2, (z1-z0)/2
	m := NewGridWithSpan(g.Span())
	m.Use(Stone).Fill(Subtract(Block(V(x0, y0-0.8, z0), V(x1, y0+1, z1)), Block(V(x0+1, y0-3, z0+1), V(x1-1, y0+2, z1-1))))
	m.Use(Wood)
	for _, x := range []float64{x0 + 0.5, c.X, x1 - 0.5} {
		for _, z := range []float64{z0 + 0.5, z1 - 0.5} {
			m.Fill(Block(V(x-0.5, y0+1, z-0.5), V(x+0.5, y0+4.4, z+0.5)))
		}
	}
	m.Fill(Block(V(x0, y0+1, z0), V(x0+1, y0+4.4, z1)), Block(V(x0, y0+1, z0), V(x1, y0+4.4, z0+1)))
	m.Fill(Block(V(x0, y0+3.6, z0), V(x1, y0+4.4, z1)))
	m.Carve(Block(V(x0+1, y0+3.5, z0+1), V(x1-1, y0+4.5, z1-1)))
	eave := y0 + 4.9
	m.Fill(Func{Box{V(x0, y0+4.4, z0), V(x1, eave+hw*0.5, z1)}, func(p Vec3) bool {
		return (p.Z <= z0+1 || p.Z >= z1-1) && p.Y <= eave+(hw-math.Abs(p.X-c.X))*0.5-0.4
	}})
	s.slabs(m, f, hw, hd, eave, DarkStone)
	// The log on the carriage and stacked planks.
	m.Use(Sand).Fill(Capsule{V(c.X+0.6, y0+1.9, z0+1.8), V(c.X+0.6, y0+1.9, z1-0.6), 0.7})
	m.Fill(Block(V(x0+1.2, y0+1, z1-3), V(x0+3, y0+2.2, z1-0.8)))
	m.Use(Charcoal).Fill(Block(V(c.X, y0+1, c.Y-0.4), V(c.X+1.2, y0+3.4, c.Y+0.4)))
	// Water wheel in the mill pond.
	wc := V(x1+1.2, y0+1.3, 40.3)
	m.Use(Wood).Fill(Func{Box{V(wc.X-0.5, wc.Y-2.6, wc.Z-2.6), V(wc.X+0.5, wc.Y+2.6, wc.Z+2.6)}, func(p Vec3) bool {
		dy, dz := p.Y-wc.Y, p.Z-wc.Z
		r := math.Hypot(dy, dz)
		return r <= 2.6 && (r >= 1.7 || math.Abs(dy) <= 0.45 || math.Abs(dz) <= 0.45)
	}})
	m.Use(Charcoal).Fill(Capsule{V(x1-0.5, wc.Y, wc.Z), V(wc.X+0.8, wc.Y, wc.Z), 0.55})
	g.Merge(m)
	// Log pile by the bridge road.
	py := l.at(-5, 35.4)
	g.Use(Wood).Fill(Capsule{V(-7.6, py+0.5, 34.8), V(-2.6, py+0.5, 34.8), 0.6}, Capsule{V(-7.6, py+0.5, 36), V(-2.6, py+0.5, 36), 0.6},
		Capsule{V(-7.4, py+1.5, 35.4), V(-2.8, py+1.5, 35.4), 0.6})
	g.Paint(Sand, Block(V(-8.3, py, 34), V(-7.5, py+2.3, 36.8)), Block(V(-2.7, py, 34), V(-1.9, py+2.3, 36.8)))
}

// cableway is the aerial tramway: valley station, three pylons, twin ropes with two red
// cabins, and the mountain station on the front ridge's shoulder.
func (s AlpineVillage) cableway(g *Grid, l *alpineVillageLand) {
	a, b := alpineVillageValley, alpineVillageSaddle
	span := b.Sub(a)
	length := span.Length()
	side := P(-span.Y, span.X).Scale(1 / length)
	ya, yb := l.at(a.X, a.Y), l.at(b.X, b.Y)

	// Valley station: stone and white walls under a roof rising toward the mountain.
	m := NewGridWithSpan(g.Span())
	roofY := func(x float64) float64 { return ya + 7 + 0.28*(a.X+4-x) }
	m.Use(Stone).Fill(Subtract(Block(V(a.X-4, ya-0.8, a.Y-3.75), V(a.X+4, ya+3, a.Y+3.75)),
		Block(V(a.X-3, ya-3, a.Y-2.75), V(a.X+3, ya+2.5, a.Y+2.75))))
	m.Use(White).Fill(Func{Box{V(a.X-4, ya+3, a.Y-3.75), V(a.X+4, ya+10, a.Y+3.75)}, func(p Vec3) bool {
		wall := math.Abs(p.Z-a.Y) >= 2.75 || p.X >= a.X+3
		return wall && p.Y <= roofY(p.X)-0.5
	}})
	m.Use(Slate).Fill(Func{Box{V(a.X-6, ya+6, a.Y-4.5), V(a.X+4.7, ya+11, a.Y+4.5)}, func(p Vec3) bool {
		y := roofY(p.X)
		return p.Y <= y && p.Y >= y-0.9
	}})
	m.Paint(Red, Block(V(a.X-3.2, ya+4.2, a.Y+3.2), V(a.X+3.2, ya+5.3, a.Y+3.8)))
	m.Paint(Charcoal, Block(V(a.X-0.7, ya, a.Y+3.2), V(a.X+0.7, ya+2.3, a.Y+3.8)))
	for _, x := range []float64{a.X - 2.4, a.X + 2.4} {
		m.Paint(Glass, Block(V(x-0.5, ya+1, a.Y+3.2), V(x+0.5, ya+2, a.Y+3.8)))
	}
	g.Merge(m)

	// Mountain station: a squat stone block with a snow-covered slab roof.
	m = NewGridWithSpan(g.Span())
	m.Use(Stone).Fill(Subtract(Block(V(b.X-3.5, yb-4, b.Y-3.5), V(b.X+3.5, yb+5.6, b.Y+3.5)),
		Block(V(b.X-2.5, yb-6, b.Y-2.5), V(b.X+2.5, yb+5.6, b.Y+2.5)),
		Block(V(b.X+2.4, yb+1.8, b.Y-2.5), V(b.X+3.7, yb+5.6, b.Y+2.5))))
	m.Paint(DarkStone, Func{Box{V(b.X-3.6, yb-4, b.Y-3.6), V(b.X+3.6, yb+5.6, b.Y+3.6)}, func(p Vec3) bool {
		return Hash01(int(math.Floor(p.X/1.2)), int(math.Floor(p.Y/0.8)), int(math.Floor(p.Z/1.2)), 95) < 0.3
	}})
	m.Paint(Glass, Block(V(b.X-2.4, yb+2.4, b.Y+2.9), V(b.X+1.6, yb+3.6, b.Y+3.6)))
	m.Paint(Red, Block(V(b.X-3.6, yb+4.4, b.Y+2.9), V(b.X+3.6, yb+5.1, b.Y+3.6)))
	m.Use(Slate).Fill(Block(V(b.X-4.2, yb+5.6, b.Y-4.2), V(b.X+5.6, yb+6.5, b.Y+4.2)))
	m.Paint(White, Block(V(b.X-4.3, yb+6.1, b.Y-4.3), V(b.X+5.7, yb+6.6, b.Y+4.3)))
	g.Merge(m)
	s.flag(g, V(b.X-2.6, yb+6.3, b.Y+2.6), 5)

	// Rope nodes: the stations' portals and the pylon heads.
	ts := []float64{3 / length, 0.27, 0.56, 0.83, 1 - 3/length}
	drop := []float64{0, 1.5, 1.6, 0.6, 0}
	y0, y1 := ya+5.4, yb+4.4
	node := func(i int) Vec3 { return a.Add(span.Scale(ts[i])).At(alpineVillageLerp(y0, y1, ts[i]) - drop[i]) }
	rope := func(t float64) Vec3 {
		i := 0
		for i+2 < len(ts) && t > ts[i+1] {
			i++
		}
		u := alpineVillageClamp((t-ts[i])/(ts[i+1]-ts[i]), 0, 1)
		p, q := node(i), node(i+1)
		at := p.Add(q.Sub(p).Scale(u))
		at.Y -= 0.6 * 4 * u * (1 - u)
		return at
	}
	off := func(p Vec3, k float64) Vec3 { return p.Add(V(side.X*k, 0, side.Y*k)) }
	g.Use(Charcoal)
	for _, sgn := range []float64{-1.6, 1.6} {
		const steps = 48
		for k := 0; k < steps; k++ {
			t0 := alpineVillageLerp(ts[0], ts[4], float64(k)/steps)
			t1 := alpineVillageLerp(ts[0], ts[4], float64(k+1)/steps)
			g.Fill(Capsule{off(rope(t0), sgn), off(rope(t1), sgn), 0.5})
		}
	}
	// Pylons: A-frames with a crossarm under the ropes.
	for i := 1; i <= 3; i++ {
		head := node(i)
		foot := l.at(head.X, head.Z)
		g.Use(DarkStone)
		for _, sgn := range []float64{-1, 1} {
			p := off(V(head.X, foot, head.Z), 2.4*sgn)
			g.Fill(Block(p.Add(V(-0.9, -1.5, -0.9)), p.Add(V(0.9, 0.5, 0.9))))
		}
		g.Use(White)
		for _, sgn := range []float64{-1, 1} {
			g.Fill(Capsule{off(V(head.X, foot, head.Z), 2.4*sgn), off(head.Add(V(0, -0.8, 0)), 0.8*sgn), 0.6})
		}
		mid := (foot + head.Y) / 2
		g.Fill(Capsule{off(V(head.X, mid, head.Z), -1.6), off(V(head.X, mid, head.Z), 1.6), 0.5})
		g.Fill(Capsule{off(head.Add(V(0, -0.8, 0)), -2.4), off(head.Add(V(0, -0.8, 0)), 2.4), 0.6})
		g.Use(Red).Fill(Sphere(head.Add(V(0, 0.1, 0)), 0.7))
	}
	// Cabins hanging from their carriages.
	cabin := func(t, sgn float64) {
		p := off(rope(t), sgn)
		g.Use(Charcoal).Fill(Block(p.Add(V(-1.1, -0.2, -0.5)), p.Add(V(1.1, 0.6, 0.5))), Block(p.Add(V(-0.4, -1.7, -0.4)), p.Add(V(0.4, 0, 0.4))))
		top := p.Y - 1.7
		g.Use(Red).Fill(Block(V(p.X-2, top-3.3, p.Z-1.5), V(p.X+2, top, p.Z+1.5)))
		g.Paint(Glass, Block(V(p.X-2.1, top-1.9, p.Z-1.6), V(p.X+2.1, top-0.8, p.Z+1.6)))
		g.Paint(Red, Block(V(p.X-0.3, top-1.9, p.Z-1.6), V(p.X+0.3, top-0.8, p.Z+1.6)))
		g.Paint(White, Block(V(p.X-2.1, top-0.4, p.Z-1.6), V(p.X+2.1, top+0.1, p.Z+1.6)))
	}
	cabin(0.415, 1.6)
	cabin(0.70, -1.6)
}

// pasture puts out the haystacks and the cows.
func (s AlpineVillage) pasture(g *Grid, l *alpineVillageLand) {
	for _, at := range []Vec2{P(-13.5, 45.3), P(-14, 49), P(-9.5, 47.2), P(-6.5, 50), P(42, 33)} {
		y := l.at(at.X, at.Y)
		g.Use(Gold).FillVolume(Cone(at, 1.7, y-0.3, 3.4))
		g.Use(Wood).Fill(Capsule{at.At(y + 2.4), at.At(y + 3.6), 0.45})
	}
	cows := []struct {
		at    Vec2
		along Axis
	}{
		{P(-19.5, 50.3), AxisX}, {P(-24, 49.5), AxisZ}, {P(-10.5, 43), AxisX}, {P(25.5, 44.5), AxisX},
		{P(30, 43), AxisZ}, {P(45, 27), AxisX}, {P(19, -16), AxisX}, {P(23, -20.5), AxisZ},
		{P(30.5, -15.5), AxisX}, {P(-22, 36.5), AxisZ}, {P(40, 36.5), AxisZ},
	}
	for k, c := range cows {
		s.cow(g, l, c.at, c.along, k)
	}
}

// cow is a small spotted cow, head toward its frame's front.
func (AlpineVillage) cow(g *Grid, l *alpineVillageLand, at Vec2, along Axis, k int) {
	f := alpineVillageFrame{at, along}
	y := l.at(at.X, at.Y)
	coat, spot := Brick, White
	if k%3 == 1 {
		coat, spot = White, Charcoal
	}
	g.Use(coat).Fill(f.box(-0.65, -1.3, y+0.9, 0.65, 1.3, y+2.2),
		f.box(-0.65, -1.3, y-0.3, 0.65, -0.6, y+0.9), f.box(-0.65, 0.6, y-0.3, 0.65, 1.3, y+0.9))
	g.Use(spot).Fill(f.box(-0.5, 1.3, y+1.2, 0.5, 2.1, y+2.3))
	g.Paint(spot, f.box(-0.8, -0.5, y+1.3, 0.8, 0.3, y+2.3))
}
