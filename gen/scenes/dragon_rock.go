package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// DragonRock is a great red dragon coiled around a basalt crag. The spire of stepped
// basalt columns stands a little behind the middle of an elliptical heath. The dragon's
// hollow body climbs it in one and a half turns, crossing the camera side at mid-height
// where its hind legs dig into the rock, and rears up the back of the shattered round
// tower on the summit. Its forelegs grip the broken wall head, both bat wings are half
// spread behind the tower facing the hero camera (+X/+Z), and the horned head swings
// out to the east (+X) on an arched neck to breathe a jet of flame over the valley. The
// tail runs out over the scree to a spade tip. A cave at the foot of the south face
// spills a gold hoard down a scree fan climbed by a switchback path, which loops across
// the scorched heath past a knight on horseback to a walled village in the east with
// one roof alight. Burnt stumps ring the crag, a tarn lies in the south-west, and green
// hills and woods close the plate's edge. North is -Z.
type DragonRock struct{}

func init() { Register(DragonRock{}) }

func (DragonRock) Name() string  { return "dragon_rock" }
func (DragonRock) Title() string { return "Dragon Rock" }

// Key positions. dragonRockNear is the plan direction toward the hero camera and
// dragonRockRight the direction to its right: the dragon is posed in that frame.
var (
	dragonRockAxis    = P(-6, -6) // Spire axis.
	dragonRockNear    = P(0.6288, 0.7776)
	dragonRockRight   = P(0.7776, -0.6288)
	dragonRockKeep    = P(-6+2.5*0.6288, -6+2.5*0.7776) // Tower axis on the summit.
	dragonRockVillage = P(37, -12)
	dragonRockTarn    = P(-31, 21)
	dragonRockRider   = P(24, 31.8) // The knight, on the path.
)

const (
	dragonRockSiteA  = 54.0 // Plate semi-axes.
	dragonRockSiteB  = 48.0
	dragonRockSummit = 45.0  // Height of the spire's top.
	dragonRockBreach = 152.0 // Bearing of the tower's breach, degrees from the hero's right.
	dragonRockHide   = 1.25  // Thickness of the dragon's hollow hide.
	dragonRockRun    = 8     // Centerline segments per bounding run.
)

func dragonRockSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func dragonRockLerp(a, b, t float64) float64 { return a + (b-a)*t }

func dragonRockDeg(d float64) float64 { return d * math.Pi / 180 }

// dragonRockNoise is smooth value noise in [-1, 1].
func dragonRockNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return dragonRockLerp(dragonRockLerp(a, b, sx), dragonRockLerp(c, d, sx), sz)*2 - 1
}

func dragonRockFBM(x, z float64, seed int) float64 {
	return 0.6*dragonRockNoise(x, z, seed) + 0.3*dragonRockNoise(x*2.03+17, z*2.03+5, seed+1) +
		0.1*dragonRockNoise(x*4.1+3, z*4.1+29, seed+2)
}

func dragonRockCross(a, b Vec3) Vec3 {
	return V(a.Y*b.Z-a.Z*b.Y, a.Z*b.X-a.X*b.Z, a.X*b.Y-a.Y*b.X)
}

func dragonRockUnit(a Vec3) Vec3 { return a.Scale(1 / math.Max(a.Length(), 1e-9)) }

// dragonRockPolar is the plan point at an angle in degrees and a radius from the spire axis.
func dragonRockPolar(deg, r float64) Vec2 { return Polar(dragonRockAxis, r, dragonRockDeg(deg)) }

// dragonRockAt places a point in the hero frame: near units toward the camera and right
// units to its right of the spire axis, at height y.
func dragonRockAt(near, right, y float64) Vec3 {
	return dragonRockAxis.Add(dragonRockNear.Scale(near)).Add(dragonRockRight.Scale(right)).At(y)
}

func dragonRockSegDist(p, a, b Vec2) float64 {
	ab := b.Sub(a)
	t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/math.Max(ab.Dot(ab), 1e-12)))
	return p.Sub(a.Add(ab.Scale(t))).Length()
}

// dragonRockTaper is a capsule whose radius varies linearly from A to B.
type dragonRockTaper struct {
	A, B   Vec3
	R0, R1 float64
}

func (c dragonRockTaper) Contains(p Vec3) bool {
	ab := c.B.Sub(c.A)
	t := math.Max(0, math.Min(1, p.Sub(c.A).Dot(ab)/math.Max(ab.Dot(ab), 1e-12)))
	return p.Sub(c.A.Add(ab.Scale(t))).Length() <= c.R0+(c.R1-c.R0)*t
}

func (c dragonRockTaper) Bounds() Box { return BoxOf(c.A, c.B).Grow(math.Max(c.R0, c.R1)) }

// dragonRockChain is a tapered tube through points with per-point radii.
func dragonRockChain(pts []Vec3, radii []float64) []Solid {
	var out []Solid
	for i := 0; i+1 < len(pts); i++ {
		out = append(out, dragonRockTaper{pts[i], pts[i+1], radii[i], radii[i+1]})
	}
	return out
}

// dragonRockFrame is an orthonormal local frame: parts authored along X (forward),
// Y (up) and Z (side) are placed at O.
type dragonRockFrame struct {
	O, X, Y, Z Vec3
	K          float64 // World units per local unit.
}

func (f dragonRockFrame) local(p Vec3) Vec3 {
	d := p.Sub(f.O).Scale(1 / f.K)
	return V(d.Dot(f.X), d.Dot(f.Y), d.Dot(f.Z))
}

func (f dragonRockFrame) world(p Vec3) Vec3 {
	return f.O.Add(f.X.Scale(p.X * f.K)).Add(f.Y.Scale(p.Y * f.K)).Add(f.Z.Scale(p.Z * f.K))
}

func (f dragonRockFrame) put(s Solid) Solid { return dragonRockFramed{s, f} }

// dragonRockFramed is a solid authored in a frame's local coordinates.
type dragonRockFramed struct {
	S Solid
	F dragonRockFrame
}

func (s dragonRockFramed) Contains(p Vec3) bool { return s.S.Contains(s.F.local(p)) }

func (s dragonRockFramed) Bounds() Box {
	in := s.S.Bounds()
	b := EmptyBox()
	for _, x := range []float64{in.Min.X, in.Max.X} {
		for _, y := range []float64{in.Min.Y, in.Max.Y} {
			for _, z := range []float64{in.Min.Z, in.Max.Z} {
				w := s.F.world(V(x, y, z))
				b = b.Union(Box{w, w})
			}
		}
	}
	return b
}

// dragonRockEach visits every lattice cell whose center lies in the box. Unlike
// Grid.Fill it never snaps centers to the half-unit authoring grid, which keeps the
// dragon's curved hide and thin members smooth.
func dragonRockEach(g *Grid, b Box, f func(c Cell, p Vec3)) {
	if b.Empty() {
		return
	}
	half := g.Span() / 2
	lo := func(v float64) int32 { return int32(math.Ceil(v / half)) }
	hi := func(v float64) int32 { return int32(math.Floor(v / half)) }
	x0, x1 := lo(b.Min.X), hi(b.Max.X)
	z0, z1 := lo(b.Min.Z), hi(b.Max.Z)
	for y := lo(b.Min.Y); y <= hi(b.Max.Y); y++ {
		par := y & 1
		for z := z0 + ((z0&1)^par)&1; z <= z1; z += 2 {
			for x := x0 + ((x0&1)^par)&1; x <= x1; x += 2 {
				f(Cell{x, y, z}, V(float64(x)*half, float64(y)*half, float64(z)*half))
			}
		}
	}
}

// dragonRockSet occupies a cell with a color, replacing whatever was there.
func dragonRockSet(g *Grid, c Cell, color Color) {
	prev := g.Brush()
	g.Remove(c)
	g.Use(color).Add(c)
	g.Use(prev)
}

// dragonRockFill occupies the cells whose centers lie inside the solids.
func dragonRockFill(g *Grid, color Color, solids ...Solid) {
	for _, s := range solids {
		dragonRockEach(g, s.Bounds(), func(c Cell, p Vec3) {
			if s.Contains(p) {
				dragonRockSet(g, c, color)
			}
		})
	}
}

// dragonRockCarve removes the cells whose centers lie inside the solids.
func dragonRockCarve(g *Grid, solids ...Solid) {
	for _, s := range solids {
		dragonRockEach(g, s.Bounds(), func(c Cell, p Vec3) {
			if s.Contains(p) {
				g.Remove(c)
			}
		})
	}
}

// dragonRockPaint recolors the occupied cells whose centers lie inside the solids.
func dragonRockPaint(g *Grid, color Color, solids ...Solid) {
	for _, s := range solids {
		dragonRockEach(g, s.Bounds(), func(c Cell, p Vec3) {
			if g.Has(c) && s.Contains(p) {
				dragonRockSet(g, c, color)
			}
		})
	}
}

// dragonRockKnot is one control point of the dragon's centerline: position, body radius
// and the direction its back faces.
type dragonRockKnot struct {
	P    Vec3
	R    float64
	Back Vec3
}

// dragonRockWyrm is the dragon's body: a smooth curve from the head joint to the tail
// tip sampled into short segments, with a radius and a back direction at every sample.
type dragonRockWyrm struct {
	pts   []Vec3
	rad   []float64
	back  []Vec3
	arc   []float64
	boxes []Box // Bounds of each run of dragonRockRun segments.
}

// dragonRockSpline threads a Catmull-Rom curve through the knots.
func dragonRockSpline(knots []dragonRockKnot, steps int) *dragonRockWyrm {
	w := &dragonRockWyrm{}
	n := len(knots)
	at := func(i int) dragonRockKnot { return knots[max(0, min(n-1, i))] }
	for i := 0; i+1 < n; i++ {
		p0, p1, p2, p3 := at(i-1).P, at(i).P, at(i+1).P, at(i+2).P
		for j := 0; j < steps; j++ {
			t := float64(j) / float64(steps)
			q := p1.Scale(2).
				Add(p2.Sub(p0).Scale(t)).
				Add(p0.Scale(2).Sub(p1.Scale(5)).Add(p2.Scale(4)).Sub(p3).Scale(t * t)).
				Add(p1.Scale(3).Sub(p0).Sub(p2.Scale(3)).Add(p3).Scale(t * t * t)).Scale(0.5)
			w.pts = append(w.pts, q)
			w.rad = append(w.rad, dragonRockLerp(at(i).R, at(i+1).R, t))
			w.back = append(w.back, at(i).Back.Add(at(i+1).Back.Sub(at(i).Back).Scale(t)))
		}
	}
	last := knots[n-1]
	w.pts, w.rad, w.back = append(w.pts, last.P), append(w.rad, last.R), append(w.back, last.Back)
	m := len(w.pts)
	w.arc = make([]float64, m)
	for i := 1; i < m; i++ {
		w.arc[i] = w.arc[i-1] + w.pts[i].Sub(w.pts[i-1]).Length()
	}
	// Square each back direction to the curve.
	for i := range w.pts {
		t := dragonRockUnit(w.pts[min(m-1, i+1)].Sub(w.pts[max(0, i-1)]))
		w.back[i] = dragonRockUnit(w.back[i].Sub(t.Scale(w.back[i].Dot(t))))
	}
	for lo := 0; lo < m-1; lo += dragonRockRun {
		b := EmptyBox()
		for i := lo; i <= min(m-1, lo+dragonRockRun); i++ {
			b = b.Union(Box{w.pts[i], w.pts[i]}.Grow(w.rad[i] + 0.6))
		}
		w.boxes = append(w.boxes, b)
	}
	return w
}

// probe finds the centerline point nearest p: its segment, the position along it and
// the distance. ok is false when p is outside every run's bounds.
func (w *dragonRockWyrm) probe(p Vec3) (seg int, t, dist float64, ok bool) {
	dist = math.Inf(1)
	for k, b := range w.boxes {
		if !b.Contains(p) {
			continue
		}
		ok = true
		for i := k * dragonRockRun; i < min(len(w.pts)-1, (k+1)*dragonRockRun); i++ {
			a := w.pts[i]
			ab := w.pts[i+1].Sub(a)
			u := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/math.Max(ab.Dot(ab), 1e-12)))
			if d := p.Sub(a.Add(ab.Scale(u))).Length(); d < dist {
				seg, t, dist = i, u, d
			}
		}
	}
	return
}

func (w *dragonRockWyrm) radius(seg int, t float64) float64 {
	return dragonRockLerp(w.rad[seg], w.rad[seg+1], t)
}

// at returns the centerline point, radius, back and tangent at an arc length from the head.
func (w *dragonRockWyrm) at(arc float64) (p Vec3, r float64, back, tan Vec3) {
	i := 0
	for i+2 < len(w.pts) && w.arc[i+1] < arc {
		i++
	}
	t := math.Max(0, math.Min(1, (arc-w.arc[i])/math.Max(w.arc[i+1]-w.arc[i], 1e-9)))
	p = w.pts[i].Add(w.pts[i+1].Sub(w.pts[i]).Scale(t))
	r = w.radius(i, t)
	back = dragonRockUnit(w.back[i].Add(w.back[i+1].Sub(w.back[i]).Scale(t)))
	tan = dragonRockUnit(w.pts[i+1].Sub(w.pts[i]))
	return
}

// skin is the hide color at a body cell: a red back with brick scale patches over a
// gold belly crossed by pale scutes.
func (w *dragonRockWyrm) skin(p Vec3, seg int, t float64) Color {
	c := w.pts[seg].Add(w.pts[seg+1].Sub(w.pts[seg]).Scale(t))
	v := dragonRockUnit(p.Sub(c))
	back := w.back[seg].Add(w.back[seg+1].Sub(w.back[seg]).Scale(t))
	tan := dragonRockUnit(w.pts[seg+1].Sub(w.pts[seg]))
	side := dragonRockCross(tan, back)
	arc := dragonRockLerp(w.arc[seg], w.arc[seg+1], t)
	cb := v.Dot(back)
	if cb < -0.08 {
		if math.Mod(arc, 3.2) < 1.0 {
			return Sand
		}
		return Gold
	}
	turn := math.Atan2(v.Dot(side), cb) * w.radius(seg, t)
	if Hash01(int(math.Floor(arc/1.9)), int(math.Floor(turn/1.9)), 31) < 0.24 {
		return Brick
	}
	return Red
}

// Spire profile: mean radius at a few heights.
var (
	dragonRockProfileY = []float64{0, 4, 10, 20, 32, 40, 45}
	dragonRockProfileR = []float64{16.5, 14, 12.2, 11, 10, 9.2, 8.7}
)

func dragonRockProfile(y float64) float64 {
	ys, rs := dragonRockProfileY, dragonRockProfileR
	if y <= ys[0] {
		return rs[0]
	}
	for i := 1; i < len(ys); i++ {
		if y <= ys[i] {
			return dragonRockLerp(rs[i-1], rs[i], (y-ys[i-1])/(ys[i]-ys[i-1]))
		}
	}
	return rs[len(rs)-1]
}

// dragonRockGirth is the spire's smooth radius toward plan angle phi at height y: the
// profile with broad lumps, which the dragon's coils follow.
func dragonRockGirth(phi, y float64) float64 {
	sn, cs := math.Sincos(phi)
	return dragonRockProfile(y) + 1.1*dragonRockNoise(cs*1.6+y*0.05, sn*1.6-y*0.03, 7)
}

// dragonRockLedge is an up-facing shelf on the spire: plan angle and half-width in
// degrees, the height of its top and how far it stands out.
type dragonRockLedge struct{ Phi, Half, Y, Out float64 }

var dragonRockLedges = []dragonRockLedge{
	{100, 26, 37, 2.6},
	{5, 22, 17.5, 2.4},
	{172, 22, 33, 2.4},
	{232, 24, 27, 2.4},
}

// dragonRockColumn names the basalt column toward phi and its block at height y.
func dragonRockColumn(phi, y float64) (col, block int) {
	col = int(math.Floor((phi/(2*math.Pi) + 1) * 30))
	block = int(math.Floor((y + 11*Hash01(col, 21)) / (6 + 6*Hash01(col, 22))))
	return
}

// dragonRockLimit is the radius of the spire's surface: columns that step in and out
// block by block, plus the ledges.
func dragonRockLimit(phi, y float64) float64 {
	r := dragonRockGirth(phi, y)
	col, block := dragonRockColumn(phi, y)
	r += 1.2 * (Hash01(col, block, 23) - 0.5)
	for _, l := range dragonRockLedges {
		if y > l.Y || y < l.Y-5 {
			continue
		}
		da := math.Abs(math.Remainder(phi-dragonRockDeg(l.Phi), 2*math.Pi))
		k := dragonRockSmooth((dragonRockDeg(l.Half) - da) / dragonRockDeg(7))
		r += l.Out * k * math.Pow((y-(l.Y-5))/5, 1.4)
	}
	return r
}

// dragonRockInCrag reports whether p is inside the spire. The outer columns stop at
// different heights, leaving a broken rim around the summit.
func dragonRockInCrag(p Vec3) bool {
	if p.Y > dragonRockSummit || p.Y < -0.5 {
		return false
	}
	d := p.Plan().Sub(dragonRockAxis)
	rho := d.Length()
	if rho > 21 {
		return false
	}
	phi := d.Angle()
	lim := dragonRockLimit(phi, p.Y)
	if rho > lim {
		return false
	}
	if rho > lim-2.6 {
		col, _ := dragonRockColumn(phi, p.Y)
		k := Hash01(col, 24)
		return p.Y <= dragonRockSummit-3.4*k*k
	}
	return true
}

// dragonRockDeep reports whether p lies well inside the spire, where nothing shows.
func dragonRockDeep(p Vec3) bool {
	if p.Y > dragonRockSummit-1.6 {
		return false
	}
	d := p.Plan().Sub(dragonRockAxis)
	return d.Length() < dragonRockLimit(d.Angle(), p.Y)-1.7
}

// dragonRockOn is the point off units outside the spire's smooth surface toward an
// angle in degrees at height y.
func dragonRockOn(deg, off, y float64) Vec3 {
	phi := dragonRockDeg(deg)
	return Polar(dragonRockAxis, dragonRockGirth(phi, y)+off, phi).At(y)
}

// dragonRockSite carries the state shared by the build steps.
type dragonRockSite struct {
	g     *Grid
	cache map[[2]int64]float64
	path  []Vec2
	wyrm  *dragonRockWyrm
}

func (DragonRock) Build(g *Grid) Order {
	s := &dragonRockSite{g: g, cache: make(map[[2]int64]float64)}
	// The path: switchbacks up the scree fan to the cave, then a loop across the heath
	// to the village gate.
	s.path = []Vec2{
		dragonRockPolar(91, 13.4), dragonRockPolar(96, 16.3), dragonRockPolar(120, 19), dragonRockPolar(96, 22),
		dragonRockPolar(122, 25), dragonRockPolar(104, 28.5),
		P(-9, 31), P(3, 38), P(17, 37), P(29, 28), P(37, 15), P(40, 3), P(38.5, -3.5),
	}
	s.wyrm = s.coil()
	s.terrain()
	s.crag()
	s.hoard()
	s.ruins()
	s.heath()
	s.village()
	s.knight()
	s.dragon()
	return Sweep(dragonRockAxis, math.Pi*0.75)
}

// fan is the height of the scree fan below the cave.
func (s *dragonRockSite) fan(rc, phi float64) float64 {
	da := math.Abs(math.Remainder(phi-dragonRockDeg(100), 2*math.Pi))
	k := dragonRockSmooth((dragonRockDeg(40) - da) / dragonRockDeg(22))
	return 8.2 * k * math.Max(0, math.Min(1, (28-rc)/14.5))
}

// tarn is the normalized radius of the tarn; below 1 is open water.
func (s *dragonRockSite) tarn(x, z float64) float64 {
	return math.Hypot((x-dragonRockTarn.X)/8, (z-dragonRockTarn.Y)/6)
}

func (s *dragonRockSite) raw(x, z float64) float64 {
	e := math.Hypot(x/dragonRockSiteA, z/dragonRockSiteB)
	if e > 1 {
		return -1
	}
	d := P(x, z).Sub(dragonRockAxis)
	rc, phi := d.Length(), d.Angle()
	h := 1.3 + 0.45*dragonRockFBM(x*0.08, z*0.08, 3)
	// Green hills along the back and left edges of the plate.
	lee := dragonRockSmooth((-(0.5*x + 0.87*z) - 4) / 34)
	h += dragonRockSmooth((e-0.58)/0.32) * lee * (3.6 + 3.2*dragonRockFBM(x*0.06+9, z*0.06+4, 11))
	h += 0.8 * dragonRockSmooth((e-0.7)/0.25)
	// Scree skirt around the spire and the fan below the cave.
	h += 2.8 * dragonRockSmooth((23-rc)/9)
	h = math.Max(h, 1.2+s.fan(rc, phi))
	// The tarn's hollow.
	if t := s.tarn(x, z); t < 1.35 {
		h = dragonRockLerp(0.3, h, dragonRockSmooth((t-0.85)/0.5))
	}
	return math.Min(h, 0.7+22*(1-e))
}

// H is the memoized terrain height; negative is off the plate.
func (s *dragonRockSite) H(x, z float64) float64 {
	key := [2]int64{int64(math.Round(x * 1000)), int64(math.Round(z * 1000))}
	if h, ok := s.cache[key]; ok {
		return h
	}
	h := s.raw(x, z)
	s.cache[key] = h
	return h
}

func (s *dragonRockSite) pathDist(q Vec2) float64 {
	d := math.Inf(1)
	for i := 0; i+1 < len(s.path); i++ {
		d = math.Min(d, dragonRockSegDist(q, s.path[i], s.path[i+1]))
	}
	return d
}

// burn is how scorched the ground is: 1 at the spire's foot, 0 in the green country,
// with a swath burnt toward the village.
func (s *dragonRockSite) burn(x, z float64) float64 {
	q := P(x, z)
	rc := q.Sub(dragonRockAxis).Length()
	b := 1 - (rc-15)/24 + 0.3*dragonRockFBM(x*0.09, z*0.09, 31)
	d := dragonRockSegDist(q, dragonRockAxis.Add(P(14, -1)), dragonRockVillage.Add(P(-7, 1)))
	return math.Max(b, 0.95-d/7+0.25*dragonRockFBM(x*0.15, z*0.15, 33))
}

// field is the color of the village yard and its strip fields, if q lies in them.
func (s *dragonRockSite) field(q Vec2) (Color, bool) {
	d := q.Sub(dragonRockVillage)
	if math.Hypot(d.X/10.5, d.Y/9) < 1 {
		return Sand, true
	}
	south := d.X > -17 && d.X < -1 && d.Y > 10 && d.Y < 24
	north := d.X > -13 && d.X < 6 && d.Y < -10 && d.Y > -19
	if !south && !north {
		return Grass, false
	}
	switch k := Hash01(int(math.Floor((q.X+100)/2.4)), 51); {
	case k < 0.45:
		return Gold, true
	case k < 0.7:
		return Wood, true
	}
	return Grass, true
}

// ground is the color of a terrain cell.
func (s *dragonRockSite) ground(p Vec3) Color {
	x, z := p.X, p.Z
	h := s.H(x, z)
	t := s.tarn(x, z)
	if h < 0.85 && t < 1.35 {
		if t < 0.5 {
			return Navy
		}
		return Water
	}
	if t < 1.2 {
		if dragonRockNoise(x*0.5, z*0.5, 47) > 0 {
			return Stone
		}
		return DarkStone
	}
	q := P(x, z)
	if s.pathDist(q) < 0.95 && p.Y > h-1.1 {
		return Sand
	}
	d := q.Sub(dragonRockAxis)
	rc := d.Length()
	n := dragonRockFBM(x*0.16, z*0.16, 41)
	if rc < 17.5 || s.fan(rc, d.Angle()) > 0.5 {
		if n > 0.2 {
			return Charcoal
		}
		return DarkStone
	}
	burn := s.burn(x, z)
	switch {
	case burn > 0.62:
		if n > 0.28 {
			return DarkStone
		}
		return Charcoal
	case burn > 0.36:
		if n > 0.2 {
			return Charcoal
		}
		return Wood
	case burn > 0.16:
		switch m := dragonRockFBM(x*0.2+7, z*0.2+3, 43); {
		case m > 0.2:
			return Foliage
		case m < -0.22:
			return Wood
		}
		return Grass
	}
	if c, ok := s.field(q); ok {
		return c
	}
	if dragonRockFBM(x*0.1+3, z*0.1+8, 45) > 0.26 || h > 5.2 {
		return Foliage
	}
	return Grass
}

// terrain fills the heath as a hollow skin over the height field and colors it.
func (s *dragonRockSite) terrain() {
	site := Box{V(-dragonRockSiteA-1, 0, -dragonRockSiteB-1), V(dragonRockSiteA+1, 13, dragonRockSiteB+1)}
	in := func(p Vec3) bool {
		if p.Y < 0 {
			return false
		}
		h := s.H(p.X, p.Z)
		return h >= 0 && p.Y <= h
	}
	land := NewGridWithSpan(s.g.Span())
	const t = 1.2
	land.Use(Grass).Fill(Func{site, func(p Vec3) bool {
		if !in(p) {
			return false
		}
		if in(p.Add(V(t, 0, 0))) && in(p.Add(V(-t, 0, 0))) && in(p.Add(V(0, 0, t))) &&
			in(p.Add(V(0, 0, -t))) && in(p.Add(V(0, t, 0))) {
			return false
		}
		// Nothing shows under the spire.
		return p.Plan().Sub(dragonRockAxis).Length() > dragonRockProfile(p.Y)-2.6
	}})
	c := dragonRockTarn
	land.Use(Water).Fill(Func{Box{V(c.X-12, 0.2, c.Y-9), V(c.X+12, 0.85, c.Y+9)}, func(p Vec3) bool {
		h := s.H(p.X, p.Z)
		return h >= 0 && h < 0.85 && s.tarn(p.X, p.Z) < 1.35
	}})
	for _, cell := range land.Cells() {
		dragonRockSet(land, cell, s.ground(land.World(cell)))
	}
	s.g.Merge(land)
}

// cavern is the cave hollowed into the foot of the south face.
func (s *dragonRockSite) cavern() Ellipsoid {
	return Ellipsoid{dragonRockPolar(86, 9.5).At(11.8), V(3.8, 4.4, 5.5)}
}

// crag builds the spire as a hollow skin of basalt columns.
func (s *dragonRockSite) crag() {
	ax := dragonRockAxis
	box := Box{V(ax.X-21, 0, ax.Y-21), V(ax.X+21, dragonRockSummit, ax.Y+21)}
	rock := NewGridWithSpan(s.g.Span())
	const t = 1.25
	dirs := []Vec3{V(t, 0, 0), V(-t, 0, 0), V(0, 0, t), V(0, 0, -t), V(0, t, 0),
		V(0.7*t, 0, 0.7*t), V(-0.7*t, 0, 0.7*t), V(0.7*t, 0, -0.7*t), V(-0.7*t, 0, -0.7*t)}
	cave := s.cavern()
	dragonRockFill(rock, Charcoal, Func{box, func(p Vec3) bool {
		if !dragonRockInCrag(p) || p.Y < s.H(p.X, p.Z)-1.6 || cave.Contains(p) {
			return false
		}
		exposed := false
		for _, d := range dirs {
			if !dragonRockInCrag(p.Add(d)) {
				exposed = true
				break
			}
		}
		if !exposed {
			return false
		}
		// Hidden inside the dragon's coils.
		if seg, u, dist, ok := s.wyrm.probe(p); ok && dist < s.wyrm.radius(seg, u)-1.5 {
			return false
		}
		return true
	}})
	// Column by column: charcoal basalt with paler and blue-grey blocks, worn grey on
	// the tops of the steps, and a pale surround of dressed stone at the cave mouth.
	lining := Ellipsoid{cave.Center, cave.Radii.Add(V(1.2, 1.2, 1.2))}
	surround := Ellipsoid{cave.Center, cave.Radii.Add(V(3.5, 3.5, 3.5))}
	for _, c := range rock.Cells() {
		p := rock.World(c)
		d := p.Plan().Sub(ax)
		col, block := dragonRockColumn(d.Angle(), p.Y)
		// The head of the spire stays dark under the pale tower.
		high := p.Y > 38
		color := Charcoal
		switch h := Hash01(col, block, 26); {
		case h < 0.22:
			if !high {
				color = DarkStone
			}
		case h < 0.32:
			color = Slate
		}
		in := d.Scale(-0.9 / math.Max(d.Length(), 1e-9))
		if !high && !dragonRockInCrag(p.Add(V(0, 0.9, 0))) && !dragonRockInCrag(p.Add(V(in.X, 0.9, in.Y))) {
			color = DarkStone
		}
		if surround.Contains(p) && !lining.Contains(p) && p.Y > 9 {
			color = Stone
			if Hash01(int(math.Floor(p.Y/1.5)), int(math.Floor(d.Angle()*8)), 28) < 0.35 {
				color = DarkStone
			}
		}
		dragonRockSet(rock, c, color)
	}
	// The cave's dark lining behind the mouth.
	dragonRockFill(rock, Charcoal, Func{lining.Bounds(), func(p Vec3) bool {
		d := p.Plan().Sub(ax)
		return lining.Contains(p) && !cave.Contains(p) && d.Length() < dragonRockLimit(d.Angle(), p.Y)-0.4
	}})
	s.g.Merge(rock)
}

// hoard is the gold spilling from the cave down the scree, with a few gems and a chest.
func (s *dragonRockSite) hoard() {
	g := s.g
	gold := []Ellipsoid{{dragonRockPolar(86, 12).At(9.1), V(3.3, 1.7, 3.6)}}
	for _, l := range [][3]float64{{83, 14.6, 2.5}, {80, 17.2, 2.1}, {78, 19.8, 1.7}, {81, 22.4, 1.3}} {
		c := dragonRockPolar(l[0], l[1])
		gold = append(gold, Ellipsoid{c.At(s.H(c.X, c.Y) + 0.1), V(l[2], 0.55*l[2], l[2])})
	}
	for _, e := range gold {
		dragonRockFill(g, Gold, e)
		dragonRockEach(g, e.Bounds(), func(c Cell, p Vec3) {
			if !e.Contains(p) {
				return
			}
			switch k := Hash01(int(c.X), int(c.Y), int(c.Z), 55); {
			case k < 0.05:
				dragonRockSet(g, c, Red)
			case k < 0.1:
				dragonRockSet(g, c, Glass)
			case k < 0.14:
				dragonRockSet(g, c, White)
			}
		})
	}
	chest := dragonRockPolar(101, 13.8)
	y := s.H(chest.X, chest.Y)
	g.With(Wood, func() { g.Fill(BlockAt(chest, 2.2, 1.6, y-0.3, y+1.2)) })
	g.With(Gold, func() { g.Fill(BlockAt(chest, 2.2, 1.6, y+1.2, y+1.7)) })
}

// rim is the height of the tower's broken wall head above a plan point.
func (s *dragonRockSite) rim(p Vec2) float64 {
	d := p.Sub(dragonRockKeep)
	th := math.Atan2(d.Dot(dragonRockNear), d.Dot(dragonRockRight))
	top := 56.7 + 0.5*math.Sin(3*th+0.7)
	// The breach torn out of the front-left; a few merlons survive away from it.
	da := math.Remainder(th-dragonRockDeg(dragonRockBreach), 2*math.Pi)
	top -= 7 * math.Exp(-da*da/0.16)
	if math.Abs(da) > 1.2 && int(math.Floor((th+7)*5.5/1.6))%2 == 0 {
		top += 1.5
	}
	k := int(math.Floor((th/(2*math.Pi) + 1) * 16))
	return top - 1.2*Hash01(k, 61)*Hash01(k, 62)
}

// ruins builds the shattered round tower, wall stubs and fallen masonry.
func (s *dragonRockSite) ruins() {
	g := s.g
	c := dragonRockKeep
	const outer, inner = 5.5, 4.2
	box := Box{V(c.X-outer-0.6, 41.5, c.Y-outer-0.6), V(c.X+outer+0.6, 59.5, c.Y+outer+0.6)}
	polar := func(p Vec3) (rho, th float64) {
		d := p.Plan().Sub(c)
		return d.Length(), math.Atan2(d.Dot(dragonRockNear), d.Dot(dragonRockRight))
	}
	// The drum, on a slightly wider plinth.
	g.With(Stone, func() {
		g.FillVolume(Func{box, func(p Vec3) bool {
			rho, _ := polar(p)
			if p.Y < 46.2 {
				return rho <= outer+0.5 && rho > inner
			}
			return rho <= outer && rho > inner && p.Y <= s.rim(p.Plan())
		}})
	})
	g.With(Wood, func() { g.Fill(Cylinder{Center: c, Radius: inner, Y0: 50.4, Y1: 51.2}) })
	masonry := func(p Vec3) (rho, th float64, ok bool) {
		rho, th = polar(p)
		return rho, th, rho > inner && rho <= outer+0.7 && p.Y > dragonRockSummit+0.3
	}
	// Weathered blocks, two string courses and soot around the breach.
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		_, th, ok := masonry(p)
		return ok && Hash01(int(math.Floor(th*3.2)), int(math.Floor(p.Y/1.2)), 63) < 0.16
	}})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		_, _, ok := masonry(p)
		return ok && (math.Abs(p.Y-49.6) < 0.35 || math.Abs(p.Y-54) < 0.35)
	}})
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		_, th, ok := masonry(p)
		da := math.Abs(math.Remainder(th-dragonRockDeg(dragonRockBreach), 2*math.Pi))
		return ok && da < 0.9 && p.Y > s.rim(p.Plan())-1.3
	}})
	// Dark arrow slits and the doorway, which faces the camera.
	opening := func(at, width, y0, y1 float64) {
		g.Paint(Charcoal, Func{box, func(p Vec3) bool {
			rho, th := polar(p)
			da := math.Abs(math.Remainder(th-dragonRockDeg(at), 2*math.Pi))
			return rho > inner && rho <= outer+0.7 && p.Y >= y0 && p.Y <= y1 && da*outer < width/2
		}})
	}
	opening(48, 1.3, 50.6, 53.4)
	opening(112, 1.3, 50.6, 53.4)
	opening(250, 1.3, 50.6, 53.4)
	opening(80, 2.3, 45.4, 48.6)

	// Stubs of the curtain wall along the summit's rim, either side of the tower: each
	// spans half degrees about a bearing and falls from y0 to y1 across it.
	stub := func(center, half, y0, y1 float64, seed int) {
		ax := dragonRockAxis
		g.With(Stone, func() {
			g.Fill(Func{Box{V(ax.X-9.5, 42, ax.Y-9.5), V(ax.X+9.5, 50.5, ax.Y+9.5)}, func(p Vec3) bool {
				d := p.Plan().Sub(ax)
				rho := d.Length()
				da := math.Remainder(d.Angle()-dragonRockDeg(center), 2*math.Pi) * 180 / math.Pi
				if rho < 7.3 || rho > 8.6 || math.Abs(da) > half {
					return false
				}
				top := dragonRockLerp(y0, y1, (da+half)/(2*half))
				arc := dragonRockDeg(da+half) * 8
				if top > 47.2 && math.Mod(arc, 3.2) < 1.6 {
					top += 1.2
				}
				return p.Y <= top-0.9*Hash01(int(math.Floor(arc/1.6)), seed)
			}})
		})
	}
	stub(137, 21, 48.8, 46.2, 71)
	stub(-23, 22, 46.2, 48.8, 72)

	block := func(q Vec2, y, w, dp, h, turn float64, col Color) {
		g.With(col, func() { g.Fill(RotateY(BlockAt(q, w, dp, y, y+h), q, turn)) })
	}
	// Fallen masonry below the breach, on the upper ledge and around the foot.
	block(dragonRockAt(5, -6, 0).Plan(), 45, 1.8, 1.4, 1.2, 0.5, Stone)
	shelf := dragonRockOn(100, 1.4, 37)
	block(shelf.Plan(), 37, 3.4, 1.3, 2.4, dragonRockDeg(190), Stone)
	block(dragonRockOn(88, 1.5, 37).Plan(), 37, 1.6, 1.4, 1.0, 0.4, Stone)
	for i, a := range []float64{18, 138, 160, 204, 250, 298, 336, 352} {
		q := dragonRockPolar(a, 17+4*Hash01(i, 73))
		col := Stone
		if i%3 == 2 {
			col = DarkStone
		}
		block(q, s.H(q.X, q.Y)-0.4, 1.6+Hash01(i, 74), 1.4+Hash01(i, 75), 1.3+0.6*Hash01(i, 76), 3*Hash01(i, 77), col)
	}
	// A dead tree on the upper ledge.
	tree := dragonRockOn(113, 1.3, 37)
	g.With(Charcoal, func() {
		g.Fill(Cylinder{Center: tree.Plan(), Radius: 0.55, Y0: 36.5, Y1: 41})
		g.Fill(Capsule{tree.Add(V(0, 2.6, 0)), tree.Add(V(-1.6, 4.4, 0.8)), 0.5})
		g.Fill(Capsule{tree.Add(V(0, 3.4, 0)), tree.Add(V(1.2, 5, 1.2)), 0.5})
	})
}

// stump is a burnt tree trunk, the taller ones with a broken limb.
func (s *dragonRockSite) stump(base Vec3, h float64, seed int) {
	g := s.g
	g.With(Charcoal, func() {
		g.Fill(Cylinder{Center: base.Plan(), Radius: 0.7, Y0: base.Y, Y1: base.Y + h})
		if h > 3.2 {
			from := base.Add(V(0, h*0.6, 0))
			tip := Polar(base.Plan(), 1.9, 2*math.Pi*Hash01(seed, 1)).At(base.Y + h*0.6 + 1.5)
			g.Fill(Capsule{from, tip, 0.5})
		}
	})
}

// tree is a round-crowned broadleaf of the green country.
func (s *dragonRockSite) tree(base Vec3, r float64, seed int) {
	g := s.g
	c := base.Plan()
	g.With(Wood, func() { g.Fill(Cylinder{Center: c, Radius: 0.6, Y0: base.Y, Y1: base.Y + 2.6}) })
	col, other := Foliage, Grass
	if Hash01(seed, 2) < 0.3 {
		col, other = Grass, Foliage
	}
	g.With(col, func() { g.FillVolume(Ellipsoid{V(c.X, base.Y+2.6+r*0.8, c.Y), V(r, r*0.85, r)}) })
	g.With(other, func() {
		g.FillVolume(Ellipsoid{V(c.X+0.5*r, base.Y+2.6+r*1.25, c.Y-0.4*r), V(r*0.6, r*0.5, r*0.6)})
	})
}

// heath scatters burnt stumps over the scorched ground and trees over the green edge.
func (s *dragonRockSite) heath() {
	clear := func(q Vec2) bool {
		if s.pathDist(q) < 2.5 || s.tarn(q.X, q.Y) < 1.4 || q.Sub(dragonRockRider).Length() < 8.5 {
			return false
		}
		d := q.Sub(dragonRockAxis)
		rc, a := d.Length(), d.Angle()
		if rc < 19 || s.fan(rc, a) > 0.2 {
			return false
		}
		v := q.Sub(dragonRockVillage)
		if math.Hypot(v.X/12.5, v.Y/11) < 1 {
			return false
		}
		if _, ok := s.field(q); ok {
			return false
		}
		// Keep the tail's sweep clear.
		if deg := a * 180 / math.Pi; deg > 30 && deg < 76 && rc < 41 {
			return false
		}
		return math.Hypot(q.X/dragonRockSiteA, q.Y/dragonRockSiteB) < 0.94
	}
	for i := -52; i <= 52; i += 5 {
		for k := -46; k <= 46; k += 5 {
			x, z := float64(i)+3.6*Hash01(i, k, 81), float64(k)+3.6*Hash01(i, k, 82)
			q := P(x, z)
			if !clear(q) {
				continue
			}
			h := s.H(x, z)
			switch burn := s.burn(x, z); {
			case burn > 0.3:
				if Hash01(i, k, 83) < 0.55 {
					s.stump(q.At(h-0.4), 2.2+3.2*Hash01(i, k, 84), i*131+k)
				}
			case burn < 0.12:
				if Hash01(i, k, 85) < 0.7 {
					s.tree(q.At(h-0.4), 2.0+0.9*Hash01(i, k, 86), i*131+k)
				}
			}
		}
	}
}

// village is the walled hamlet east of the crag, one house burning.
func (s *dragonRockSite) village() {
	g := s.g
	c := dragonRockVillage
	// Ring wall with a gate toward the path.
	g.With(Stone, func() {
		g.Fill(Subtract(EllipticCylinder{c, 10.5, 9, 0.6, 3.9}, EllipticCylinder{c, 9.4, 7.9, 0, 5}))
	})
	g.Paint(DarkStone, Subtract(EllipticCylinder{c, 10.6, 9.1, 3.4, 4}, EllipticCylinder{c, 9.3, 7.8, 0, 5}))
	g.Carve(BlockAt(P(38.5, -3.2), 3, 3.6, 1, 5))
	for _, x := range []float64{36.4, 40.6} {
		g.With(Stone, func() { g.Fill(Cylinder{Center: P(x, -3.2), Radius: 0.9, Y0: 0.6, Y1: 4.9}) })
		g.With(DarkStone, func() { g.Fill(Cylinder{Center: P(x, -3.2), Radius: 1.1, Y0: 4.9, Y1: 5.5}) })
	}
	house := func(q Vec2, w, dp, h float64, ridge Axis, wall, roof Color) float64 {
		y := math.Round(s.H(q.X, q.Y)*2) / 2
		lo, hi := q.Sub(P(w/2, dp/2)), q.Add(P(w/2, dp/2))
		g.With(wall, func() { g.Fill(Block(lo.At(y-0.8), hi.At(y+h))) })
		g.Carve(Block(lo.Add(P(1, 1)).At(y+0.5), hi.Sub(P(1, 1)).At(y+h)))
		g.With(roof, func() { g.Fill(Gable{lo.Sub(P(0.5, 0.5)), hi.Add(P(0.5, 0.5)), y + h, ridge}) })
		g.Carve(Gable{lo.Add(P(0.6, 0.6)), hi.Sub(P(0.6, 0.6)), y + h - 0.01, ridge})
		g.Paint(Wood, BlockAt(q.Add(P(-0.8, dp/2)), 1.2, 0.6, y, y+2.1))
		g.Paint(Charcoal, BlockAt(q.Add(P(1.3, dp/2)), 1.0, 0.6, y+1.2, y+2.2))
		g.Paint(Charcoal, BlockAt(q.Add(P(w/2, 0)), 0.6, 1.0, y+1.2, y+2.2))
		return y + h
	}
	house(c.Add(P(-4.6, -3.4)), 6, 4.6, 3, AxisX, White, Wood)
	house(c.Add(P(2.8, -4.2)), 5, 5.4, 3, AxisZ, Sand, Slate)
	house(c.Add(P(4.8, 2.6)), 5.4, 4.4, 3, AxisX, White, Wood)
	// The burning house, nearest the crag.
	fire := c.Add(P(-3.8, 3.2))
	top := house(fire, 5.4, 4.4, 2.8, AxisX, Sand, Charcoal)
	for i := 0; i < 5; i++ {
		base := V(fire.X-2.2+1.1*float64(i), top+1.2, fire.Y+0.8*(Hash01(i, 92)-0.5))
		h := 3.2 + 3*Hash01(i, 91)
		mid, tip := base.Add(V(0.3, h*0.55, 0)), base.Add(V(0.9, h, 0.2))
		g.With(Gold, func() { g.Fill(dragonRockTaper{base, mid, 1.0, 0.75}) })
		g.With(Red, func() { g.Fill(dragonRockTaper{mid, tip, 0.75, 0.5}) })
	}
	g.Paint(Red, Func{Box{V(fire.X-3.3, top, fire.Y-2.8), V(fire.X+3.3, top+2.2, fire.Y+2.8)}, func(p Vec3) bool {
		return Hash01(int(math.Floor(p.X/1.2)), int(math.Floor(p.Z/1.2)), 93) < 0.3
	}})
	for i := 0; i < 6; i++ {
		t := float64(i) / 5
		pc := fire.Add(P(1.5+6*t*t+1.1*math.Sin(2.3*float64(i)), -3*t)).At(top + 5 + 11*t)
		r := 1.3 + 1.5*t
		col := Charcoal
		if i%2 == 1 {
			col = DarkStone
		}
		g.With(col, func() { g.FillVolume(Ellipsoid{pc, V(r*1.15, r*0.9, r)}) })
	}
	// Haystacks and a well in the yard.
	for _, d := range []Vec2{P(0.4, 0.6), P(-0.6, -7), P(7.4, -1.2)} {
		q := c.Add(d)
		y := s.H(q.X, q.Y) - 0.3
		g.With(Gold, func() { g.FillVolume(Frustum{q, 1.5, 0.3, y, y + 2.6}) })
	}
	well := c.Add(P(0.6, 4.8))
	y := s.H(well.X, well.Y) - 0.3
	g.With(Stone, func() { g.Fill(Cylinder{Center: well, Radius: 1.3, Y0: y, Y1: y + 1.3}) })
	g.With(Water, func() { g.Fill(Cylinder{Center: well, Radius: 0.6, Y0: y + 0.8, Y1: y + 1.3}) })
}

// knight rides a white horse along the path toward the crag, lance raised.
func (s *dragonRockSite) knight() {
	g := s.g
	at := dragonRockRider
	dir := dragonRockUnit(V(-12, 0, 9)) // Along the path, side-on to the hero camera.
	up := V(0, 1, 0)
	f := dragonRockFrame{at.At(s.H(at.X, at.Y) - 0.2), dir, up, dragonRockCross(dir, up), 1.55}
	part := func(col Color, solids ...Solid) {
		for _, sol := range solids {
			dragonRockFill(g, col, f.put(sol))
		}
	}
	// The horse.
	part(White, Ellipsoid{V(0, 2.9, 0), V(2.4, 1.2, 1.05)},
		dragonRockTaper{V(1.7, 3.3, 0), V(3.0, 4.9, 0), 0.9, 0.62},
		dragonRockTaper{V(2.9, 5.0, 0), V(4.2, 4.2, 0), 0.68, 0.45})
	for _, leg := range [][2]float64{{1.7, 0.55}, {1.4, -0.55}, {-1.6, 0.55}, {-1.9, -0.55}} {
		part(White, Capsule{V(leg[0], 2.3, leg[1]), V(leg[0]+0.2, 0.2, leg[1]), 0.42})
	}
	part(Charcoal, Capsule{V(-2.3, 3.2, 0), V(-3.0, 1.5, 0), 0.4}, Capsule{V(2.0, 4.7, 0), V(2.7, 5.4, 0), 0.4})
	part(Red, Ellipsoid{V(-0.2, 3.1, 0), V(1.35, 1.0, 1.2)})
	// The rider: armor, helm and plume, a shield on the camera side and a couched lance.
	part(Glass, Ellipsoid{V(0.1, 4.9, 0), V(0.7, 1.15, 0.8)},
		Capsule{V(0.1, 4.1, 0.95), V(0.6, 2.6, 1.2), 0.38}, Capsule{V(0.1, 4.1, -0.95), V(0.6, 2.6, -1.2), 0.38})
	part(Stone, Sphere(V(0.2, 6.45, 0), 0.72))
	part(Red, Capsule{V(0.1, 7.15, 0), V(-0.7, 7.6, 0), 0.4})
	part(White, Ellipsoid{V(0.5, 4.8, -1.2), V(0.7, 0.95, 0.42)})
	part(Wood, Capsule{V(-1.2, 4.7, 1.25), V(5.6, 7.6, 1.25), 0.36})
	part(Red, Ellipsoid{V(4.7, 8.0, 1.25), V(0.9, 0.55, 0.4)})
}

// coil lays out the dragon's centerline from the head joint to the tail tip.
func (s *dragonRockSite) coil() *dragonRockWyrm {
	up := V(0, 1, 0)
	away := dragonRockNear.Scale(-1).At(0)
	var k []dragonRockKnot
	add := func(p Vec3, r float64, back Vec3) { k = append(k, dragonRockKnot{p, r, back}) }
	// The neck arches forward and to the right from the shoulders above the tower.
	add(dragonRockAt(5.5, 16, 61.5), 3.0, up)
	add(dragonRockAt(4.6, 11.5, 65.3), 3.0, up)
	add(dragonRockAt(2.6, 7, 67.2), 3.1, up)
	add(dragonRockAt(0, 3, 66.2), 3.4, up.Add(away.Scale(0.6)))
	add(dragonRockAt(-3, 0.8, 63.2), 3.9, away.Add(up.Scale(0.5)))
	// The forebody rears up the back of the tower.
	add(dragonRockAt(-5.2, 0, 59), 4.5, away)
	add(dragonRockAt(-6.6, 0, 53), 4.6, away)
	add(dragonRockAt(-7.3, -0.8, 47.5), 4.6, away)
	// One and a half turns down around the spire; the angle grows on the way down.
	turns := []float64{240, 258, 280, 300, 321}
	for a := 341.0; a <= 761; a += 20 {
		turns = append(turns, a)
	}
	for _, a := range turns {
		y, r := 33.5+0.105*(321-a), 4.6
		if a > 321 {
			y = 33.5 - 24*(a-321)/360
		}
		if a > 380 {
			r = 4.6 - 2.7*(a-380)/391
		}
		phi := dragonRockDeg(a)
		q := Polar(dragonRockAxis, dragonRockGirth(phi, y)+0.5*r, phi)
		y = math.Max(y, s.H(q.X, q.Y)+0.8*r)
		out := dragonRockUnit(q.Sub(dragonRockAxis).At(0))
		add(q.At(y), r, up.Add(out.Scale(0.4)))
	}
	// The tail trails out over the heath.
	for _, t := range [][3]float64{{60, 20.5, 1.75}, {66, 25, 1.6}, {66, 30, 1.45}, {58, 34, 1.25}, {49, 35.5, 1.05}, {42, 34.5, 0.85}} {
		q := dragonRockPolar(t[0], t[1])
		add(q.At(s.H(q.X, q.Y)+0.8*t[2]), t[2], up)
	}
	return dragonRockSpline(k, 6)
}

// dragon builds the beast in its own grid, so that its colors never touch the scenery.
func (s *dragonRockSite) dragon() {
	d := NewGridWithSpan(s.g.Span())
	s.body(d)
	s.spines(d)
	s.legs(d)
	s.spade(d)
	s.wing(d, 1)
	s.wing(d, -1)
	s.head(d)
	s.g.Merge(d)
}

// body fills the hollow hide along the centerline, leaving out what the rock hides.
func (s *dragonRockSite) body(d *Grid) {
	w := s.wyrm
	for _, box := range w.boxes {
		dragonRockEach(d, box, func(c Cell, p Vec3) {
			if d.Has(c) {
				return
			}
			seg, t, dist, ok := w.probe(p)
			if !ok {
				return
			}
			r := w.radius(seg, t)
			if dist > r || dist <= r-dragonRockHide || dragonRockDeep(p) {
				return
			}
			d.Use(w.skin(p, seg, t)).Add(c)
		})
	}
}

// spines is the row of charcoal dorsal spines from the neck to the tail.
func (s *dragonRockSite) spines(d *Grid) {
	w := s.wyrm
	total := w.arc[len(w.arc)-1]
	for a := 2.5; a < total-5; {
		p, r, back, tan := w.at(a)
		h := 1.2 + 0.6*r
		base := p.Add(back.Scale(r - 0.5))
		tip := p.Add(back.Scale(r + h)).Add(tan.Scale(0.35 * h))
		dragonRockFill(d, Charcoal, dragonRockTaper{base, tip, 0.45 + 0.14*r, 0.45})
		a += 2.2 + 0.45*r
	}
}

// leg is one limb: thigh, shin and a foot with white claws.
func (s *dragonRockSite) leg(d *Grid, hip, knee, ankle, foot Vec3, claws ...Vec3) {
	dragonRockFill(d, Red, dragonRockChain([]Vec3{hip, knee, ankle, foot}, []float64{2.3, 1.6, 1.15, 1.2})...)
	dragonRockFill(d, Red, Sphere(foot, 1.45))
	for _, c := range claws {
		dragonRockFill(d, White, Capsule{foot.Add(c.Sub(foot).Scale(0.35)), c, 0.5})
	}
}

// legs places the hind legs on the rock face and the forelegs on the tower's wall head.
func (s *dragonRockSite) legs(d *Grid) {
	on := dragonRockOn
	// Hind legs, where the coil crosses the camera side: one braced below the body,
	// one reaching up the face above it.
	s.leg(d, on(50, 4.6, 26), on(57, 7.4, 23.2), on(49, 2.4, 19), on(46, 1.0, 17.8),
		on(41, 0.9, 15.7), on(46, 0.9, 15.5), on(51, 0.9, 15.7))
	s.leg(d, on(40, 2.6, 31), on(30, 5.0, 34.4), on(37, 2.0, 37.4), on(39, 1.0, 38.6),
		on(34, 0.9, 40.6), on(39, 0.9, 40.9), on(44, 0.9, 40.6))
	// Forelegs: elbows out, claws hooked over the tower wall.
	for _, side := range []float64{1, -1} {
		grip := dragonRockAt(2.0, 4.8*side, 0)
		top := s.rim(grip.Plan()) + 0.6
		grip.Y = top
		s.leg(d, dragonRockAt(-4.6, 3.6*side, 60), dragonRockAt(-2, 8.2*side, 61.6),
			dragonRockAt(1.2, 6.6*side, top+1.5), grip,
			dragonRockAt(0.8, 6.5*side, top-2.3), dragonRockAt(2.1, 6.5*side, top-2.5), dragonRockAt(3.4, 6.1*side, top-2.3))
	}
}

// spade is the flat diamond tip of the tail.
func (s *dragonRockSite) spade(d *Grid) {
	w := s.wyrm
	n := len(w.pts)
	dir := w.pts[n-1].Sub(w.pts[n-4])
	dir = dragonRockUnit(V(dir.X, 0, dir.Z))
	up := V(0, 1, 0)
	f := dragonRockFrame{w.pts[n-1], dir, up, dragonRockCross(dir, up), 1}
	blade := func(grow float64) Solid {
		return f.put(Func{Box{V(-1.5, -0.8, -3.2), V(7.5, 0.9, 3.2)}, func(p Vec3) bool {
			return math.Abs(p.X-2.6)/(3.6+grow)+math.Abs(p.Z)/(2.6+grow) <= 1
		}})
	}
	dragonRockFill(d, Charcoal, blade(0))
	dragonRockFill(d, Red, blade(-0.95))
}

// wing builds one half-spread bat wing behind the shoulders; side is +1 for the wing
// on the hero view's right. The wing lies in a plane leaning back from the camera,
// cupped forward toward its tips: an arm to the wrist, four finger spars fanning from
// it and a membrane with a scalloped trailing edge.
func (s *dragonRockSite) wing(d *Grid, side float64) {
	up := V(0, 1, 0)
	near := dragonRockNear.At(0)
	right := dragonRockRight.At(0).Scale(side)
	const tilt = 0.35
	rise := up.Scale(math.Cos(tilt)).Sub(near.Scale(math.Sin(tilt)))
	face := near.Scale(math.Cos(tilt)).Add(up.Scale(math.Sin(tilt)))
	root := dragonRockAt(-7.4, 2.6*side, 61.2)
	cup := func(u, v float64) float64 { return 0.0055 * (u*u + v*v) }
	at := func(q Vec2) Vec3 {
		return root.Add(right.Scale(q.X)).Add(rise.Scale(q.Y)).Add(face.Scale(cup(q.X, q.Y)))
	}
	wrist := P(9, 7)
	var tips []Vec2
	lengths := []float64{20.5, 22.5, 21, 17.5}
	for i, elev := range []float64{53, 32, 11, -11} {
		tips = append(tips, Polar(wrist, lengths[i], dragonRockDeg(elev)))
	}
	attach := P(0.5, -6.5)
	outline := []Vec2{P(0, 0), wrist}
	scallop := func(a, b Vec2, k float64) {
		for _, t := range []float64{0.25, 0.5, 0.75} {
			m := a.Add(b.Sub(a).Scale(t))
			outline = append(outline, wrist.Add(m.Sub(wrist).Scale(1-k*4*t*(1-t))))
		}
	}
	for i, tip := range tips {
		outline = append(outline, tip)
		if i+1 < len(tips) {
			scallop(tip, tips[i+1], 0.2)
		}
	}
	scallop(tips[len(tips)-1], attach, 0.33)
	outline = append(outline, attach)
	// A slightly smaller outline marks off the darker hem.
	heart := wrist.Add(P(7, 2))
	inner := make([]Vec2, len(outline))
	box := EmptyBox()
	for i, q := range outline {
		inner[i] = heart.Add(q.Sub(heart).Scale(0.86))
		w := at(q)
		box = box.Union(Box{w, w})
	}
	box = box.Grow(1.5)
	plane := func(p Vec3) (q Vec2, off float64) {
		r := p.Sub(root)
		q = P(r.Dot(right), r.Dot(rise))
		return q, r.Dot(face) - cup(q.X, q.Y)
	}
	dragonRockFill(d, Terracotta, Func{box, func(p Vec3) bool {
		q, off := plane(p)
		return math.Abs(off) <= 0.8 && InsidePolygon(outline, q)
	}})
	dragonRockPaint(d, Brick, Func{box, func(p Vec3) bool {
		q, off := plane(p)
		return math.Abs(off) <= 0.8 && InsidePolygon(outline, q) && !InsidePolygon(inner, q)
	}})
	// Arm and finger spars, following the cupped surface.
	spar := func(a, b Vec2, r0, r1 float64) {
		const n = 5
		pts, radii := make([]Vec3, n+1), make([]float64, n+1)
		for i := range pts {
			t := float64(i) / n
			pts[i], radii[i] = at(a.Add(b.Sub(a).Scale(t))), dragonRockLerp(r0, r1, t)
		}
		dragonRockFill(d, Red, dragonRockChain(pts, radii)...)
	}
	spar(P(0, 0), wrist, 1.6, 1.3)
	for _, tip := range tips {
		spar(wrist, tip, 1.25, 0.6)
	}
	dragonRockFill(d, Red, Sphere(at(wrist), 1.6))
	dragonRockFill(d, Charcoal, dragonRockTaper{at(wrist), at(wrist.Add(P(-1.6, 3.8))), 0.8, 0.5})
}

// head builds the horned head with open jaws on the end of the neck, then the flame.
func (s *dragonRockSite) head(d *Grid) {
	up := V(0, 1, 0)
	fwd := dragonRockUnit(dragonRockNear.At(0).Scale(0.3).Add(dragonRockRight.At(0).Scale(0.9)).Add(up.Scale(-0.38)))
	side := dragonRockUnit(dragonRockCross(fwd, up))
	f := dragonRockFrame{s.wyrm.pts[0], fwd, dragonRockCross(side, fwd), side, 1.3}
	fill := func(col Color, solids ...Solid) {
		for _, sol := range solids {
			dragonRockFill(d, col, f.put(sol))
		}
	}
	paint := func(col Color, box Box, in func(Vec3) bool) { dragonRockPaint(d, col, f.put(Func{box, in})) }
	// The lower jaw hangs open below the snout.
	jaw := func(x float64) float64 { return -1.8 - 0.55*(x-1.8) }
	fill(Red,
		Ellipsoid{V(2.2, 0.5, 0), V(3.6, 2.7, 2.9)},
		dragonRockTaper{V(3.5, 0.5, 0), V(10.4, -0.2, 0), 2.35, 1.45},
		Ellipsoid{V(10.6, 0.2, 0), V(1.5, 1.35, 1.6)},
		dragonRockTaper{V(1.8, -1.8, 0), V(9.2, jaw(9.2), 0), 1.9, 1.05})
	all := Box{V(-8, -9, -5), V(14, 7, 5)}
	// Pale throat and chin.
	paint(Gold, all, func(p Vec3) bool {
		return (p.X > 1 && p.Y < jaw(p.X)-0.15) || (p.X <= 3 && p.Y < -1.7)
	})
	// Hollow the mouth and darken it.
	dragonRockCarve(d, f.put(Func{Box{V(3.2, -8, -1.3), V(13, -0.9, 1.3)}, func(p Vec3) bool {
		return p.Y > jaw(p.X)+0.1
	}}))
	paint(Brick, Box{V(2.5, -8, -1.75), V(12.5, -0.8, 1.75)}, func(p Vec3) bool { return p.Y > jaw(p.X)-0.5 })
	paint(Charcoal, Box{V(2.5, -8, -1.75), V(4.8, -0.8, 1.75)}, func(p Vec3) bool { return p.Y > jaw(p.X)-0.5 })
	// Teeth along both jaws, fangs at the front.
	for _, z := range []float64{-1.45, 1.45} {
		for i := 0; i < 5; i++ {
			x := 5 + 1.3*float64(i)
			fill(White, Capsule{V(x, -1.5, z), V(x+0.1, -2.7, z), 0.45})
		}
		fill(White, Capsule{V(10.7, -1.1, z*0.8), V(10.9, -3.1, z*0.8), 0.5})
		for i := 0; i < 4; i++ {
			x := 5.6 + 1.25*float64(i)
			r := dragonRockLerp(1.9, 1.05, (x-1.8)/7.4)
			fill(White, Capsule{V(x, jaw(x)+r-0.3, z*0.8), V(x+0.1, jaw(x)+r+0.8, z*0.8), 0.45})
		}
	}
	// Eyes under dark brows.
	for _, z := range []float64{-2, 2} {
		fill(Gold, Sphere(V(4.7, 1.3, z), 1.0))
		fill(Charcoal, Capsule{V(3.2, 2.6, z), V(6.0, 2.2, z*0.9), 0.5})
	}
	// Horns sweeping back, cheek spikes, a nose horn and the crest.
	for _, z := range []float64{-1, 1} {
		fill(Sand, dragonRockChain([]Vec3{V(0.8, 2.2, 1.7*z), V(-3.2, 4.6, 2.6*z), V(-6.6, 5.6, 3.1*z)},
			[]float64{1.05, 0.8, 0.5})...)
		fill(Sand, dragonRockTaper{V(0.5, -0.6, 2.6*z), V(-2.8, -1.0, 4.1*z), 0.75, 0.5})
	}
	fill(Sand, dragonRockTaper{V(10.6, 1.2, 0), V(10.0, 2.9, 0), 0.6, 0.45})
	fill(Charcoal, dragonRockTaper{V(2.6, 2.9, 0), V(1.4, 4.8, 0), 0.7, 0.45},
		dragonRockTaper{V(0, 2.8, 0), V(-1.4, 4.5, 0), 0.7, 0.45})
	s.flame(d, f)
}

// flame is the jet from the open jaws: a hollow, slowly drooping plume that flares out
// of the mouth and tapers to a ragged point, gold where it is hottest and cooling
// through terracotta to red, with loose tongues licking off it.
func (s *dragonRockSite) flame(d *Grid, head dragonRockFrame) {
	dir := dragonRockUnit(head.X.Add(head.Y.Scale(-0.3)))
	f := dragonRockFrame{head.world(V(8, -2.8, 0)), dir, dragonRockCross(head.Z, dir), head.Z, 1}
	const length, droop = 27.0, 0.004
	girth := func(u float64) float64 {
		if u < 0.45 {
			return 1.3 + 3.1*math.Sin(u/0.45*math.Pi/2)
		}
		return 4.4 * (1 - math.Pow((u-0.45)/0.55, 1.5))
	}
	// shape returns the distance from the jet's axis, the local radius, the position
	// along the jet and how far the fire has cooled there.
	shape := func(p Vec3) (dist, r, u, cool float64) {
		u = p.X / length
		dy := p.Y + droop*p.X*p.X
		dist = math.Hypot(dy, p.Z)
		sn, cs := p.Z/math.Max(dist, 1e-9), dy/math.Max(dist, 1e-9)
		lobe := dragonRockNoise(cs*1.4+p.X*0.22, sn*1.4-p.X*0.1, 95)
		cool = u + 0.16*dragonRockNoise(cs*1.2+p.X*0.35+4, sn*1.2+8, 97)
		return dist, girth(u) * (1 + (0.1+0.3*u)*lobe), u, cool
	}
	box := Box{V(0, -10, -6.5), V(length, 6.5, 6.5)}
	dragonRockEach(d, f.put(Func{Box: box}).Bounds(), func(c Cell, w Vec3) {
		p := f.local(w)
		if !box.Contains(p) {
			return
		}
		dist, r, _, cool := shape(p)
		if dist > r || dist <= r-1.3 {
			return
		}
		switch {
		case cool < 0.36:
			dragonRockSet(d, c, Gold)
		case cool < 0.66:
			dragonRockSet(d, c, Terracotta)
		default:
			dragonRockSet(d, c, Red)
		}
	})
	// Loose tongues: along the jet, bearing around it.
	for _, t := range [][2]float64{{9, 0.4}, {12.5, 3.4}, {15, 1.5}, {17.5, -1.2}, {20, 2.6}, {22, 0.2}} {
		x := t[0]
		r := girth(x / length)
		sn, cs := math.Sincos(t[1])
		a := V(x, cs*0.7*r-droop*x*x, sn*0.7*r)
		b := V(x+4.2, cs*(r+2.2)-droop*(x+4.2)*(x+4.2), sn*(r+2.2))
		col := Terracotta
		if x > 16 {
			col = Red
		}
		dragonRockFill(d, col, f.put(dragonRockTaper{a, b, 1.15, 0.5}))
	}
}
