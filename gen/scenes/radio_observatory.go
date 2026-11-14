package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// RadioObservatory is a radio astronomy site on a high desert plateau. The hero is a
// fully steerable 64-unit dish standing left of center: a thin white paraboloid with a
// faint grid of panel seams, tipped 40° from the zenith and aimed south-south-east
// (toward +Z, slightly +X) so the hero and front views look into the bowl. A slate
// quadrupod carries the charcoal-and-gold focus cabin; behind the reflector a deep cone
// of radial rib trusses and ring trusses hangs from the elevation cradle, whose gold
// toothed gear arc and charcoal counterweights swing between two slate A-frame towers.
// The towers stand on a turntable that rides four bogies around a circular rail on a
// concrete ring. Along the eastern side (+X) five identical small dishes of an
// interferometer array stand on a railway, all pointing the same way. In front (+Z)
// are the control building with its glazed operations room, a car park with pickups, a
// water tower, a log-periodic antenna and a pair of gold helical antennas; an optical
// dome with an open slit crowns a low mesa in the south-west corner (-X, +Z), and a dry
// arroyo winds between the great dish and the array. A perimeter road, a power line to
// a substation on the northern rim, saguaros, yuccas and scrub finish the banded sand,
// terracotta and brick plateau. North is -Z.
type RadioObservatory struct{}

func init() { Register(RadioObservatory{}) }

func (RadioObservatory) Name() string  { return "radio_observatory" }
func (RadioObservatory) Title() string { return "Radio Observatory" }

// Key dimensions.
const (
	radioObservatoryHalfX   = 56.0 // Plateau half-extents and corner radius.
	radioObservatoryHalfZ   = 53.0
	radioObservatoryCorner  = 30.0
	radioObservatoryTop     = 3.5 // Plateau surface.
	radioObservatoryFoot    = 1.6 // Height of the rim cliff's foot.
	radioObservatoryRoad    = 5.8 // Perimeter road: centerline inset from the rim, half-width.
	radioObservatoryRoadW   = 1.6
	radioObservatoryRowX    = 39.5 // Array railway and its half-length.
	radioObservatoryRowZ    = 38.0
	radioObservatoryAxisY   = 34.0 // Elevation axis of the great dish.
	radioObservatoryDishR   = 32.0 // Reflector radius and depth.
	radioObservatoryDepth   = 11.0
	radioObservatoryVertex  = 9.0  // Vertex distance ahead of the elevation axis.
	radioObservatoryGear    = 21.0 // Elevation gear radius.
	radioObservatoryRail    = 21.2 // Azimuth rail radius.
	radioObservatoryMesaH   = 9.0  // Mesa height above the plateau.
	radioObservatoryStep    = 0.25 // Height-field sample spacing.
	radioObservatoryAzimuth = 20 * math.Pi / 180
	radioObservatoryZenith  = 40 * math.Pi / 180
)

var (
	radioObservatoryBig   = P(-13, -13)  // Great dish azimuth axis.
	radioObservatoryMesa  = P(-35, 28.8) // Mesa center.
	radioObservatoryTank  = P(-21, 40.5) // Water tower.
	radioObservatoryLog   = P(-17, 29)   // Log-periodic antenna.
	radioObservatoryHelix = P(15, 15)    // Helical antenna pair.
	// Control points of the arroyo, north to south.
	radioObservatoryArroyo = []Vec2{P(27, -58), P(23, -44), P(28, -31), P(21.5, -19), P(26.5, -6),
		P(20.5, 6), P(26, 18), P(21.5, 30), P(27, 42), P(24, 58)}
	// Rock strata by height, bottom up: each color runs up to Y.
	radioObservatoryStrata = []struct {
		Y float64
		C Color
	}{{0.9, Brick}, {1.7, Terracotta}, {2.2, Sand}, {3.1, Terracotta}, {3.8, Brick}, {5.3, Terracotta},
		{5.9, Sand}, {7.2, Brick}, {8.6, Terracotta}, {9.2, Sand}, {10.6, Terracotta}, {11.4, Brick},
		{12, Terracotta}}
)

func radioObservatorySmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func radioObservatoryLerp(a, b, t float64) float64 { return a + (b-a)*t }

func radioObservatoryMix(a, b Vec3, t float64) Vec3 { return a.Add(b.Sub(a).Scale(t)) }

// radioObservatoryNoise is smooth value noise in [-1, 1].
func radioObservatoryNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return radioObservatoryLerp(radioObservatoryLerp(a, b, sx), radioObservatoryLerp(c, d, sx), sz)*2 - 1
}

func radioObservatoryFBM(x, z float64, seed int) float64 {
	return 0.6*radioObservatoryNoise(x, z, seed) + 0.3*radioObservatoryNoise(x*2.03+17, z*2.03+5, seed+1) +
		0.1*radioObservatoryNoise(x*4.1+3, z*4.1+29, seed+2)
}

// radioObservatoryNear is the plan distance from p to a polyline.
func radioObservatoryNear(p Vec2, pts []Vec2) float64 {
	best := math.Inf(1)
	for i := 0; i+1 < len(pts); i++ {
		a, b := pts[i], pts[i+1]
		ab := b.Sub(a)
		t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
		best = math.Min(best, p.Sub(a.Add(ab.Scale(t))).Length())
	}
	return best
}

// radioObservatorySpline samples a Catmull-Rom curve through pts, n steps per span.
func radioObservatorySpline(pts []Vec2, n int) []Vec2 {
	at := func(i int) Vec2 { return pts[max(0, min(len(pts)-1, i))] }
	var out []Vec2
	for i := 0; i+1 < len(pts); i++ {
		p0, p1, p2, p3 := at(i-1), at(i), at(i+1), at(i+2)
		for k := 0; k < n; k++ {
			t := float64(k) / float64(n)
			q := p1.Scale(2).Add(p2.Sub(p0).Scale(t)).
				Add(p0.Scale(2).Sub(p1.Scale(5)).Add(p2.Scale(4)).Sub(p3).Scale(t * t)).
				Add(p1.Scale(3).Sub(p0).Sub(p2.Scale(3)).Add(p3).Scale(t * t * t))
			out = append(out, q.Scale(0.5))
		}
	}
	return append(out, pts[len(pts)-1])
}

// radioObservatoryFrame is the tipping frame of a dish. O lies on the elevation axis; L
// runs along the axle, T toward the upper rim and A along the beam. H is the horizontal
// pointing direction.
type radioObservatoryFrame struct {
	O, L, T, A, H Vec3
}

// radioObservatoryAim builds a frame pointing at azimuth (from +Z toward +X) and zenith angle.
func radioObservatoryAim(o Vec3, azimuth, zenith float64) radioObservatoryFrame {
	sa, ca := math.Sincos(azimuth)
	sz, cz := math.Sincos(zenith)
	return radioObservatoryFrame{O: o, L: V(ca, 0, -sa), T: V(-cz*sa, sz, -cz*ca), A: V(sz*sa, cz, sz*ca), H: V(sa, 0, ca)}
}

// at maps frame coordinates (u along the axle, v toward the upper rim, w along the beam)
// to world space.
func (f radioObservatoryFrame) at(u, v, w float64) Vec3 {
	return f.O.Add(f.L.Scale(u)).Add(f.T.Scale(v)).Add(f.A.Scale(w))
}

func (f radioObservatoryFrame) local(p Vec3) (u, v, w float64) {
	d := p.Sub(f.O)
	return d.Dot(f.L), d.Dot(f.T), d.Dot(f.A)
}

// ground is the world point u along the axle and v ahead of the azimuth axis, at height y.
func (f radioObservatoryFrame) ground(u, v, y float64) Vec3 {
	return V(f.O.X+f.L.X*u+f.H.X*v, y, f.O.Z+f.L.Z*u+f.H.Z*v)
}

// level places a solid authored around the origin (X along the axle, Z ahead) on the
// azimuth axis without tipping it.
func (f radioObservatoryFrame) level(s Solid, u, v float64) Solid {
	q := f.ground(u, v, 0)
	return Translate(RotateY(s, P(0, 0), -math.Atan2(f.H.X, f.H.Z)), q)
}

// solid is the frame-aligned box u0..u1 × v0..v1 × w0..w1, optionally cut by a predicate.
func (f radioObservatoryFrame) solid(u0, u1, v0, v1, w0, w1 float64, in func(u, v, w float64) bool) Solid {
	b := EmptyBox()
	for _, u := range []float64{u0, u1} {
		for _, v := range []float64{v0, v1} {
			for _, w := range []float64{w0, w1} {
				p := f.at(u, v, w)
				b = b.Union(Box{p, p})
			}
		}
	}
	return Func{b, func(p Vec3) bool {
		u, v, w := f.local(p)
		return u >= u0 && u <= u1 && v >= v0 && v <= v1 && w >= w0 && w <= w1 && (in == nil || in(u, v, w))
	}}
}

// bowl is a paraboloid shell of even thickness with its vertex w0 along the beam.
func (f radioObservatoryFrame) bowl(w0, radius, depth, thick float64) Solid {
	k := depth / (radius * radius)
	return f.solid(-radius, radius, -radius, radius, w0-thick, w0+depth+thick, func(u, v, w float64) bool {
		r2 := u*u + v*v
		return r2 <= radius*radius && math.Abs(w-w0-k*r2) <= thick/2*math.Sqrt(1+4*k*k*r2)
	})
}

// radioObservatorySite holds the grid and the sampled height field.
type radioObservatorySite struct {
	g      *Grid
	wash   []Vec2 // Arroyo centerline.
	nx, nz int
	h      []float64
}

func newRadioObservatorySite(g *Grid) *radioObservatorySite {
	s := &radioObservatorySite{g: g, wash: radioObservatorySpline(radioObservatoryArroyo, 5)}
	s.nx = int((2*radioObservatoryHalfX+4)/radioObservatoryStep) + 2
	s.nz = int((2*radioObservatoryHalfZ+4)/radioObservatoryStep) + 2
	s.h = make([]float64, s.nx*s.nz)
	for j := 0; j < s.nz; j++ {
		for i := 0; i < s.nx; i++ {
			s.h[j*s.nx+i] = s.raw(-radioObservatoryHalfX-2+float64(i)*radioObservatoryStep,
				-radioObservatoryHalfZ-2+float64(j)*radioObservatoryStep)
		}
	}
	return s
}

// cell is the lattice cell under a world point.
func (s *radioObservatorySite) cell(p Vec3) Cell { return Nearest(p.Scale(1 / s.g.Span())) }

// rim is the distance inside the plateau's clean rounded outline.
func (radioObservatorySite) rim(x, z float64) float64 {
	const r = radioObservatoryCorner
	qx, qz := math.Abs(x)-(radioObservatoryHalfX-r), math.Abs(z)-(radioObservatoryHalfZ-r)
	return r - math.Hypot(math.Max(qx, 0), math.Max(qz, 0)) - math.Min(math.Max(qx, qz), 0)
}

// edge is the distance inside the ragged cliff line.
func (s radioObservatorySite) edge(x, z float64) float64 {
	return s.rim(x, z) - 0.7 - 0.7*radioObservatoryNoise(x*0.09, z*0.09, 9)
}

// lap is a running length along the perimeter road, for its center dashes.
func (radioObservatorySite) lap(x, z float64) float64 {
	ax, az := math.Abs(x), math.Abs(z)
	sx, sz := radioObservatoryHalfX-radioObservatoryCorner, radioObservatoryHalfZ-radioObservatoryCorner
	switch {
	case ax <= sx:
		return ax
	case az <= sz:
		return az
	}
	return math.Atan2(az-sz, ax-sx) * (radioObservatoryCorner - radioObservatoryRoad)
}

// built is the distance from a plan point to the nearest built footprint (negative inside).
func (radioObservatorySite) built(x, z float64) float64 {
	p := P(x, z)
	rect := func(x0, z0, x1, z1 float64) float64 {
		dx, dz := math.Max(x0-x, x-x1), math.Max(z0-z, z-z1)
		return math.Hypot(math.Max(dx, 0), math.Max(dz, 0)) + math.Min(math.Max(dx, dz), 0)
	}
	d := p.Sub(radioObservatoryBig).Length() - 24.5
	d = math.Min(d, rect(-14.5, 31.5, 20.5, 44.5)) // Control building and car park.
	d = math.Min(d, rect(radioObservatoryRowX-3.8, -radioObservatoryRowZ-1, radioObservatoryRowX+3.8, radioObservatoryRowZ+1))
	d = math.Min(d, p.Sub(radioObservatoryTank).Length()-4.5)
	d = math.Min(d, p.Sub(radioObservatoryLog).Length()-2.5)
	d = math.Min(d, p.Sub(radioObservatoryHelix).Length()-4)
	return d
}

// mesa is the mesa's rise above the plateau and the distance inside its cliff line.
func (radioObservatorySite) mesa(x, z float64) (rise, q float64) {
	d := P(x, z).Sub(radioObservatoryMesa)
	a := d.Angle()
	r := 10.6 + 0.5*radioObservatoryNoise(math.Cos(a)*1.7+3, math.Sin(a)*1.7+7, 41) + 0.2*math.Cos(2*a+0.6)
	q = r - d.Length()
	switch {
	case q < -3.2:
		return 0, q
	case q < 0:
		return 2.0 * radioObservatorySmooth((q+3.2)/3.2), q
	}
	return 2.0 + (radioObservatoryMesaH-2.0)*radioObservatorySmooth(q/1.2), q
}

// raw is the terrain height at a plan point, or -2 off the plateau.
func (s *radioObservatorySite) raw(x, z float64) float64 {
	e := s.edge(x, z)
	if e < 0 {
		return -2
	}
	const top = radioObservatoryTop
	ce := s.rim(x, z)
	wild := radioObservatorySmooth(s.built(x, z) / 5)
	h := top + wild*0.55*radioObservatoryFBM(x*0.07, z*0.07, 3)
	if rise, _ := s.mesa(x, z); rise > 0 {
		h = math.Max(h, top+rise)
	}
	// The road stays level and bridges the arroyo on a culvert.
	road := radioObservatorySmooth((ce-radioObservatoryRoad+radioObservatoryRoadW+1.2)/1.0) *
		radioObservatorySmooth((radioObservatoryRoad+radioObservatoryRoadW+1.2-ce)/1.0)
	if d := radioObservatoryNear(P(x, z), s.wash); d < 3.6 {
		h -= 2.1 * radioObservatorySmooth((3.4-d)/1.8) * (1 - road)
	}
	h = radioObservatoryLerp(h, top, road)
	return radioObservatoryFoot + (h-radioObservatoryFoot)*radioObservatorySmooth(e/1.2)
}

// H is the sampled terrain height (bilinear); -2 off the plateau.
func (s *radioObservatorySite) H(x, z float64) float64 {
	fx := (x + radioObservatoryHalfX + 2) / radioObservatoryStep
	fz := (z + radioObservatoryHalfZ + 2) / radioObservatoryStep
	i, j := int(math.Floor(fx)), int(math.Floor(fz))
	if i < 0 || j < 0 || i >= s.nx-1 || j >= s.nz-1 {
		return -2
	}
	u, v := fx-float64(i), fz-float64(j)
	a, b := s.h[j*s.nx+i], s.h[j*s.nx+i+1]
	c, d := s.h[(j+1)*s.nx+i], s.h[(j+1)*s.nx+i+1]
	return radioObservatoryLerp(radioObservatoryLerp(a, b, u), radioObservatoryLerp(c, d, u), v)
}

func (s *radioObservatorySite) slope(x, z float64) float64 {
	return math.Hypot(s.H(x+0.5, z)-s.H(x-0.5, z), s.H(x, z+0.5)-s.H(x, z-0.5))
}

// terrain builds the plateau as a hollow skin over the height field and colors it.
func (s *radioObservatorySite) terrain() {
	g := s.g
	box := Box{V(-radioObservatoryHalfX-1, -0.3, -radioObservatoryHalfZ-1),
		V(radioObservatoryHalfX+1, radioObservatoryTop+radioObservatoryMesaH+1, radioObservatoryHalfZ+1)}
	in := func(p Vec3) bool {
		h := s.H(p.X, p.Z)
		return h > 0.5 && p.Y <= h && p.Y >= -0.3
	}
	const t = 0.8
	const k = t * math.Sqrt2 / 2
	g.Use(Sand).Fill(Func{box, func(p Vec3) bool {
		if !in(p) {
			return false
		}
		return !(in(p.Add(V(0, t, 0))) &&
			in(p.Add(V(t, 0, 0))) && in(p.Add(V(-t, 0, 0))) && in(p.Add(V(0, 0, t))) && in(p.Add(V(0, 0, -t))) &&
			in(p.Add(V(k, 0, k))) && in(p.Add(V(-k, 0, k))) && in(p.Add(V(k, 0, -k))) && in(p.Add(V(-k, 0, -k))))
	}})
	for _, c := range g.Cells() {
		col := s.soil(g.World(c), c)
		g.Remove(c)
		g.Use(col).Add(c)
	}
}

// soil picks the color of one terrain cell.
func (s *radioObservatorySite) soil(p Vec3, c Cell) Color {
	x, z := p.X, p.Z
	hash := func(salt int) float64 { return Hash01(int(c.X), int(c.Y), int(c.Z), salt) }
	const top = radioObservatoryTop
	h := s.H(x, z)
	ce := s.rim(x, z)
	surface := p.Y > h-0.85
	level := surface && h < top+0.4 && h > top-0.4
	strata := func() Color {
		yy := p.Y + 0.4*radioObservatoryNoise(x*0.12, z*0.12, 21)
		for _, b := range radioObservatoryStrata {
			if yy < b.Y {
				return b.C
			}
		}
		return Brick
	}

	// Perimeter road with center dashes.
	if d := math.Abs(ce - radioObservatoryRoad); d < radioObservatoryRoadW && level {
		if d < 0.35 && math.Mod(s.lap(x, z)+400, 4) < 2.2 {
			return Gold
		}
		return Charcoal
	}
	// Car park, its drive to the road, and the gravel path to the telescope.
	if level && x > 10.5 && x < 19.5 && z > 32.5 && z < 43 {
		for _, line := range []float64{10.5, 13.5, 16.5, 19.5} {
			if math.Abs(x-line) < 0.3 && z < 39 {
				return White
			}
		}
		return Charcoal
	}
	if level && x > 12 && x < 16 && z >= 43 && ce > radioObservatoryRoad {
		return Charcoal
	}
	if level && radioObservatoryNear(P(x, z), []Vec2{P(-2, 33), P(-7.3, 10.5)}) < 1.3 {
		if hash(1) < 0.3 {
			return DarkStone
		}
		return Stone
	}
	// Gravel inside the foundation ring.
	if dc := P(x, z).Sub(radioObservatoryBig).Length(); dc < 24.4 && surface {
		if radioObservatoryNoise(x*0.4, z*0.4, 23) > 0.2 {
			return DarkStone
		}
		return Stone
	}
	// Arroyo: pale sandy bed between banded banks.
	if d := radioObservatoryNear(P(x, z), s.wash); d < 3.7 && h < top-0.4 {
		switch {
		case h < top-1.75 && surface:
			if hash(2) < 0.06 {
				return Stone
			}
			return Sand
		case p.Y < top-1.2:
			return Brick
		}
		return Terracotta
	}
	if !surface || s.slope(x, z) > 1.0 || s.edge(x, z) < 1.2 {
		return strata()
	}
	// Mesa cap.
	if _, q := s.mesa(x, z); q > 0.8 {
		switch n := radioObservatoryFBM(x*0.16, z*0.16, 27); {
		case q < 2.0:
			return Brick
		case n > 0.2:
			return Sand
		}
		return Terracotta
	}
	// Open desert: broad contour bands of sand, terracotta and brick.
	n := radioObservatoryFBM(x*0.04, z*0.04, 31)*2.6 + 0.1*radioObservatoryNoise(x*0.3, z*0.3, 33)
	switch f := n - math.Floor(n); {
	case f < 0.52:
		return Sand
	case f < 0.87:
		return Terracotta
	}
	return Brick
}

// telescope is the great dish with its foundation, alidade, cradle, reflector and feed.
func (s *radioObservatorySite) telescope() {
	g := s.g
	c := radioObservatoryBig
	f := radioObservatoryAim(c.At(radioObservatoryAxisY), radioObservatoryAzimuth, radioObservatoryZenith)
	const (
		R  = radioObservatoryDishR
		D  = radioObservatoryDepth
		W0 = radioObservatoryVertex
		G  = radioObservatoryGear
		y0 = radioObservatoryTop
	)
	ringTop := y0 + 1.1
	deck := y0 + 4.6 // Turntable beam level.

	// Foundation: concrete ring with radial joints, the rail, and the central pintle.
	g.Use(Stone).Fill(Cylinder{c, 23.3, 19.2, y0 - 0.2, ringTop})
	g.Paint(DarkStone, Func{Box{V(c.X-24, ringTop-0.5, c.Y-24), V(c.X+24, ringTop, c.Y+24)}, func(p Vec3) bool {
		d := p.Plan().Sub(c)
		return math.Abs(math.Remainder(d.Angle(), math.Pi/8))*d.Length() < 0.3
	}})
	g.Use(Charcoal).Fill(Cylinder{c, radioObservatoryRail + 0.5, radioObservatoryRail - 0.5, ringTop, ringTop + 0.45})
	g.Use(Stone).FillVolume(Cylinder{c, 3.4, 0, y0 - 0.3, y0 + 2.2})
	g.Use(Slate).FillVolume(Cylinder{c, 2.2, 0, y0 + 2.2, y0 + 4.2})

	// Alidade: four bogies, the turntable frame, and two A-frame towers.
	foot := func(su, sv float64) Vec3 { return f.ground(su*15, sv*15, deck) }
	apex := func(su float64) Vec3 { return f.at(su*13.5, 0, 0) }
	signs := []float64{-1, 1}
	for _, su := range signs {
		for _, sv := range signs {
			q := foot(su, sv)
			tangent := q.Plan().Sub(c).Angle() + math.Pi/2
			place := func(sol Solid) Solid { return Translate(RotateY(sol, P(0, 0), tangent), V(q.X, 0, q.Z)) }
			g.Use(Charcoal).Fill(place(BlockAt(P(0, 0), 5.6, 2.8, ringTop+0.45, deck-0.9)))
			g.Use(Gold).Fill(place(BlockAt(P(0, 0), 3.2, 3.2, deck-0.9, deck+1.0)))
		}
	}
	hub := f.ground(0, 0, deck-0.4)
	g.Use(Slate)
	for _, sv := range signs {
		g.Fill(Capsule{foot(-1, sv), foot(1, sv), 0.9})
		g.Fill(Capsule{radioObservatoryMix(foot(-1, sv), apex(-1), 0.16), radioObservatoryMix(foot(1, sv), apex(1), 0.16), 0.7})
		g.Fill(Capsule{foot(-1, sv), hub, 0.7}, Capsule{foot(1, sv), hub, 0.7})
	}
	for _, su := range signs {
		g.Fill(Capsule{foot(su, -1), foot(su, 1), 0.9})
		g.Fill(Capsule{foot(su, -1), apex(su), 1.1}, Capsule{foot(su, 1), apex(su), 1.1})
		g.Fill(Capsule{radioObservatoryMix(foot(su, -1), apex(su), 0.42), radioObservatoryMix(foot(su, 1), apex(su), 0.42), 0.8})
	}
	// Elevation drive: a pylon on the turntable with the pinion housing under the gear.
	g.Use(Slate).Fill(f.level(BlockAt(P(0, 0), 3.4, 3.4, deck-0.4, y0+7.6), 0, 0))
	g.Use(Gold).Fill(f.level(BlockAt(P(0, 0), 4.6, 2.6, y0+7.6, y0+9.0), 0, 0))
	// Equipment cabin on the eastern tower's tie beam.
	tie := radioObservatoryMix(foot(1, 0), apex(1), 0.42)
	g.Use(White).Fill(f.level(BlockAt(P(1.4, 0), 3.4, 5.4, tie.Y+0.6, tie.Y+3.8), 14.4, 0))
	g.Paint(Glass, f.level(BlockAt(P(2.9, 0), 0.8, 3.6, tie.Y+1.8, tie.Y+3.0), 14.4, 0))
	g.Paint(Slate, f.level(BlockAt(P(1.4, 0), 3.4, 5.4, tie.Y+3.4, tie.Y+3.8), 14.4, 0))

	// Back-up structure: radial rib trusses on a cone of lower chords, tied by ring trusses.
	const ribs = 16
	radii := []float64{6, 12.5, 19, 25.5, 31}
	cup := func(r float64) float64 { return W0 + D*r*r/(R*R) }
	chord := func(r float64) float64 {
		return radioObservatoryLerp(W0-9.5, cup(31)-2.2, (r-radii[0])/(radii[4]-radii[0]))
	}
	node := func(k int, r, w float64) Vec3 {
		a := float64(k) * 2 * math.Pi / ribs
		return f.at(r*math.Cos(a), r*math.Sin(a), w)
	}
	g.Use(Stone)
	for k := 0; k < ribs; k++ {
		first := k % 2 // Every other rib starts at the first ring, not the hub.
		for i := first; i < len(radii); i++ {
			r0 := radii[i]
			g.Fill(Capsule{node(k, r0, chord(r0)), node(k, r0, cup(r0)-0.6), 0.5})
			if i+1 < len(radii) {
				r1 := radii[i+1]
				g.Fill(Capsule{node(k, r0, chord(r0)), node(k, r1, chord(r1)), 0.6})
			}
		}
	}
	g.Use(DarkStone)
	for i, r := range radii {
		step := 1
		if i == 0 {
			step = 2
		}
		for k := 0; k < ribs; k += step {
			g.Fill(Capsule{node(k, r, chord(r)), node(k+step, r, chord(r)), 0.55})
		}
	}

	// Elevation cradle: axle, struts up to the ring trusses, gear arc and counterweights.
	g.Use(Slate)
	for _, su := range signs {
		a := f.at(su*11, 0, 0)
		side := 0
		if su < 0 {
			side = ribs / 2
		}
		g.Fill(Capsule{a, node(side, radii[1], chord(radii[1])), 0.8})
		g.Fill(Capsule{a, node(side+2, radii[2], chord(radii[2])), 0.8}, Capsule{a, node(side-2, radii[2], chord(radii[2])), 0.8})
		g.Fill(Capsule{f.at(su*5, 0, 0), f.at(su*5, 0, -13), 0.8})
	}
	for _, psi := range []float64{-0.9, -0.3, 0.3, 0.9} {
		sp, cp := math.Sincos(psi)
		g.Fill(Capsule{f.at(0, 0, 0), f.at(0, (G-2)*sp, -(G-2)*cp), 0.75})
	}
	arc := func(lo float64) Solid {
		return f.solid(-1.1, 1.1, -G-2, G+2, -G-2, 2, func(u, v, w float64) bool {
			rho, psi := math.Hypot(v, w), math.Atan2(v, -w)
			if math.Abs(psi) > 0.98 || rho < lo {
				return false
			}
			tooth := 0.0
			if math.Mod(psi*G+300, 3.0) < 1.5 {
				tooth = 1.2
			}
			return rho <= G+tooth
		})
	}
	g.Use(Gold).FillVolume(arc(G - 2.6))
	g.Paint(Charcoal, arc(G-0.1))
	for _, su := range signs {
		u0, u1 := math.Min(su*2.6, su*7.2), math.Max(su*2.6, su*7.2)
		g.Use(Charcoal).Fill(f.solid(u0, u1, -3.2, 3.2, -18, -12.8, nil))
		g.Paint(DarkStone, f.solid(u0, u1, -3.2, 3.2, -16, -14.8, nil))
	}
	g.Use(Charcoal).Fill(Capsule{f.at(-13.5, 0, 0), f.at(13.5, 0, 0), 1.3})
	for _, su := range signs {
		g.Use(Gold).Fill(Sphere(apex(su), 2.2))
	}

	// Reflector: built apart so its seams can be drawn per cell, then laid over the ribs.
	dish := NewGridWithSpan(g.Span())
	dish.Use(White).FillVolume(f.bowl(W0, R, D, 1.5))
	rings := []float64{6.4, 12.8, 19.2, 25.6}
	sectors := []int{8, 16, 24, 32, 40}
	for _, cell := range dish.Cells() {
		u, v, _ := f.local(dish.World(cell))
		r, a := math.Hypot(u, v), math.Atan2(v, u)
		zone := 0
		for zone < len(rings) && r > rings[zone] {
			zone++
		}
		col := White
		switch {
		case r > R-0.8:
			col = Stone
		case r > 2 && math.Abs(math.Remainder(a, 2*math.Pi/float64(sectors[zone])))*r < 0.4:
			col = Stone
		}
		for i, ring := range rings {
			if math.Abs(r-ring) < 0.36 {
				col = []Color{Stone, Glass}[i%2]
			}
		}
		if col != White {
			dish.Remove(cell)
			dish.Use(col).Add(cell)
		}
	}
	g.Merge(dish)

	// Quadrupod and prime-focus cabin.
	F := R * R / (4 * D)
	for k := 0; k < 4; k++ {
		sa, ca := math.Sincos(math.Pi/4 + float64(k)*math.Pi/2)
		g.Use(Slate).Fill(Capsule{f.at(21*ca, 21*sa, cup(21)), f.at(1.8*ca, 1.8*sa, W0+F-1.5), 0.7})
	}
	g.Use(Stone).Fill(f.solid(-2.2, 2.2, -2.2, 2.2, W0-0.5, W0+1.6, func(u, v, w float64) bool { return math.Hypot(u, v) <= 2.2 }))
	g.Use(Charcoal).Fill(f.solid(-2.6, 2.6, -2.6, 2.6, W0+F-2.6, W0+F+2.6, nil))
	g.Use(Gold).Fill(f.solid(-2.8, 2.8, -2.8, 2.8, W0+F-0.5, W0+F+0.5, nil))
	g.Use(Gold).Fill(f.solid(-1.4, 1.4, -1.4, 1.4, W0+F-4.4, W0+F-2.6, nil))
	g.Use(Red).Fill(Sphere(f.at(0, 0, W0+F+3.4), 0.85), Sphere(f.at(0, R, cup(R)+1.0), 0.85))
}

// antenna is one array dish on its pad: pedestal, yoke, gear sector, reflector and feed.
func (s *radioObservatorySite) antenna(c Vec2) {
	g := s.g
	const (
		y0 = radioObservatoryTop
		R  = 8.25
		D  = 2.9
		W0 = 2.4
	)
	g.Use(Stone).Fill(BlockAt(c, 7, 7, y0-0.2, y0+0.7))
	g.Use(White).FillVolume(Frustum{c, 2.3, 1.6, y0 + 0.7, y0 + 7.2})
	g.Paint(Slate, Cylinder{c, 2.6, 0, y0 + 0.7, y0 + 1.5})
	f := radioObservatoryAim(c.At(y0+8.6), radioObservatoryAzimuth, radioObservatoryZenith)
	g.Use(Slate).Fill(f.level(BlockAt(P(0, 0), 4.2, 2.6, y0+6.9, y0+9.5), 0, 0))
	g.Use(Gold).FillVolume(f.solid(-0.6, 0.6, -6, 6, -6, 2, func(u, v, w float64) bool {
		rho, psi := math.Hypot(v, w), math.Atan2(v, -w)
		return rho <= 5.0 && rho >= 3.2 && psi > -0.35 && psi < 1.3
	}))
	cup := func(r float64) float64 { return W0 + D*r*r/(R*R) }
	g.Use(Stone)
	for k := 0; k < 8; k++ {
		sa, ca := math.Sincos(float64(k) * math.Pi / 4)
		g.Fill(Capsule{f.at(1.2*ca, 1.2*sa, 0.2), f.at(6.6*ca, 6.6*sa, cup(6.6)-0.7), 0.5})
	}
	g.Use(Slate).Fill(Capsule{f.at(0, 0, 0), f.at(0, 0, W0), 1.4})
	g.Use(White).FillVolume(f.bowl(W0, R, D, 1.3))
	g.Paint(Stone, f.solid(-R, R, -R, R, W0-1, W0+D+1, func(u, v, w float64) bool {
		r := math.Hypot(u, v)
		return r > R-0.7 || math.Abs(r-4.3) < 0.3
	}))
	F := R * R / (4 * D)
	for k := 0; k < 4; k++ {
		sa, ca := math.Sincos(math.Pi/4 + float64(k)*math.Pi/2)
		g.Use(Slate).Fill(Capsule{f.at(5.6*ca, 5.6*sa, cup(5.6)), f.at(0.5*ca, 0.5*sa, W0+F), 0.5})
	}
	g.Use(Charcoal).Fill(Sphere(f.at(0, 0, W0+F+0.3), 1.2))
	g.Use(Gold).Fill(Sphere(f.at(0, 0, W0+F-0.9), 0.7))
}

// railway is the array track: ballast, ties, rails, buffer stops and the five antennas.
func (s *radioObservatorySite) railway() {
	g := s.g
	const (
		x  = radioObservatoryRowX
		zr = radioObservatoryRowZ
		y0 = radioObservatoryTop
	)
	g.Use(DarkStone).Fill(Block(V(x-2.6, y0-0.2, -zr), V(x+2.6, y0+0.4, zr)))
	g.Paint(Wood, Func{Box{V(x-2.1, y0+0.1, -zr), V(x+2.1, y0+0.5, zr)}, func(p Vec3) bool {
		return math.Mod(p.Z+100, 2) < 0.8
	}})
	g.Use(Charcoal).Fill(Block(V(x-1.5, y0+0.4, -zr), V(x-1.5, y0+0.8, zr)), Block(V(x+1.5, y0+0.4, -zr), V(x+1.5, y0+0.8, zr)))
	for _, z := range []float64{-zr + 0.5, zr - 0.5} {
		g.Use(Red).Fill(Block(V(x-1.8, y0+0.4, z-0.5), V(x+1.8, y0+1.9, z+0.5)))
		g.Use(White).Fill(Block(V(x-1.8, y0+1.0, z-0.5), V(x+1.8, y0+1.4, z+0.5)))
	}
	for i := -2; i <= 2; i++ {
		s.antenna(P(x, float64(i)*17))
	}
}

// hall is the control building: a long white block with ribbon windows, the glazed
// operations room above its east end, and antennas on the roof.
func (s *radioObservatorySite) hall() {
	g := s.g
	const (
		y0   = radioObservatoryTop
		x0   = -13.0
		z0   = 33.0
		x1   = 9.0
		z1   = 41.5
		roof = y0 + 4.5
	)
	g.Use(White).Fill(Block(V(x0, y0-0.3, z0), V(x1, roof, z1)))
	g.Carve(Block(V(x0+1, y0-0.3, z0+1), V(x1-1, roof-0.9, z1-1)))
	g.Paint(Glass, Func{Box{V(x0, y0+1.6, z0), V(x1, y0+3.1, z1)}, func(p Vec3) bool {
		if p.X < x0+0.8 {
			return math.Mod(p.Z-z0, 4.25) > 1.2
		}
		return (p.Z > z1-0.8 || p.Z < z0+0.8) && math.Mod(p.X-x0, 4) > 1.2
	}})
	g.Paint(Slate, Block(V(x0, roof-0.7, z0), V(x1, roof, z1)))
	g.Paint(Stone, Block(V(x0+0.8, roof-0.2, z0+0.8), V(x1-0.8, roof, z1-0.8)))
	g.Paint(Charcoal, Block(V(-5.4, y0, z1-0.8), V(-3.6, y0+2.7, z1)))
	g.Use(Slate).Fill(Block(V(-6.4, y0+2.9, z1), V(-2.6, y0+3.4, z1+1.6)))

	// Operations room: glass on all four sides under an overhanging slab.
	const (
		ox0 = 1.5
		ox1 = 10.0
		oz0 = 32.5
		oz1 = 42.0
		lid = roof + 3.3
	)
	g.Use(Glass).Fill(Block(V(ox0, roof, oz0), V(ox1, lid, oz1)))
	g.Carve(Block(V(ox0+1, roof, oz0+1), V(ox1-1, lid, oz1-1)))
	g.Paint(White, Block(V(ox0, roof, oz0), V(ox1, roof+0.7, oz1)))
	for _, x := range []float64{ox0, ox1} {
		for _, z := range []float64{oz0, (oz0 + oz1) / 2, oz1} {
			g.Paint(White, Cylinder{P(x, z), 0.55, 0, roof, lid})
		}
	}
	g.Use(White).Fill(Block(V(ox0-0.8, lid, oz0-0.8), V(ox1+0.8, lid+0.8, oz1+0.8)))
	g.Paint(Slate, Block(V(ox0-0.8, lid, oz0-0.8), V(ox1+0.8, lid+0.35, oz1+0.8)))

	// Rooftop: air handlers, whip masts, a satellite dish and a beacon mast.
	g.Use(Slate).Fill(Block(V(-4.5, roof, 34.5), V(-2.3, roof+1.3, 36.3)), Block(V(-1.3, roof, 34.5), V(0.5, roof+1.3, 36.3)))
	for i, q := range []Vec2{P(-11.5, 34.5), P(-6.5, 40), P(-0.5, 40)} {
		top := roof + 6 - 1.2*float64(i)
		g.Use(Stone).Fill(Capsule{q.At(roof), q.At(top), 0.5})
		g.Use(Red).Fill(Sphere(q.At(top+0.4), 0.6))
	}
	sat := radioObservatoryAim(V(-9.5, roof+2.6, 37.5), 0.9, 0.85)
	g.Use(Slate).Fill(Capsule{V(-9.5, roof, 37.5), sat.O, 0.6})
	g.Use(White).Fill(sat.bowl(0.4, 2.6, 0.9, 1.0))
	g.Use(Charcoal).Fill(Sphere(sat.at(0, 0, 2.6), 0.6))
	mast := V(5.8, lid+0.8, 37.2)
	g.Use(White).Fill(Capsule{mast, mast.Add(V(0, 5.5, 0)), 0.5})
	g.Paint(Red, Cylinder{mast.Plan(), 1, 0, mast.Y + 1.8, mast.Y + 3.6})
	g.Use(Stone).Fill(Capsule{mast.Add(V(-1.6, 4.2, 0)), mast.Add(V(1.6, 4.2, 0)), 0.5})
	g.Use(Red).Fill(Sphere(mast.Add(V(0, 6, 0)), 0.7))
}

// dome is the optical telescope on the mesa: a white drum and hemisphere with a dark,
// open slit facing the hero camera.
func (s *radioObservatorySite) dome() {
	g := s.g
	c := radioObservatoryMesa.Add(P(0.4, -0.2))
	y0 := radioObservatoryTop + radioObservatoryMesaH
	wall := y0 + 3.6
	const R = 5.3
	g.Use(White).Fill(Cylinder{c, R - 0.2, R - 1.3, y0 - 0.4, wall})
	g.Paint(Stone, Cylinder{c, R, 0, y0 - 0.4, y0 + 0.5})
	g.Use(Slate).Fill(Cylinder{c, R + 0.6, R - 1.2, wall - 0.4, wall + 0.3})
	o := c.At(wall)
	box := Box{V(c.X-R-1, wall, c.Y-R-1), V(c.X+R+1, wall+R+1, c.Y+R+1)}
	shell := func(r0, r1 float64) Solid {
		return Func{box, func(p Vec3) bool {
			d := p.Sub(o).Length()
			return d <= r1 && d > r0
		}}
	}
	g.Use(Charcoal).Fill(shell(R-2.1, R-0.9))
	g.Use(White).FillVolume(shell(R-1.0, R))
	dir := P(0.66, 0.75)
	dir = dir.Scale(1 / dir.Length())
	side := P(-dir.Y, dir.X)
	g.Carve(Func{box, func(p Vec3) bool {
		d := p.Sub(o)
		return math.Abs(d.Plan().Dot(side)) < 1.3 && d.Plan().Dot(dir) > -1.6 && d.Y > 0.5 && d.Length() > R-1.0
	}})
	// Shutter rails either side of the slit.
	g.Use(Stone).Fill(Func{box, func(p Vec3) bool {
		d := p.Sub(o)
		across := math.Abs(d.Plan().Dot(side))
		return across >= 1.3 && across < 2.1 && d.Plan().Dot(dir) > -2.4 && d.Y > 0.3 && d.Length() > R-0.3 && d.Length() <= R+0.6
	}})
	// The telescope's objective peers out of the slit.
	aim := V(dir.X*0.62, 0.78, dir.Y*0.62)
	g.Use(Slate).Fill(Capsule{o.Add(aim.Scale(1.0)), o.Add(aim.Scale(R - 1.6)), 1.0})
	g.Use(Glass).Fill(Sphere(o.Add(aim.Scale(R-1.2)), 0.9))
	// Door and a short stair on the east side.
	door := Polar(c, R-0.7, 0.35)
	g.Paint(Charcoal, Cylinder{door, 0.9, 0, y0 + 0.5, y0 + 2.6})
}

// tank is the elevated water tower.
func (s *radioObservatorySite) tank(c Vec2) {
	g := s.g
	const y0 = radioObservatoryTop
	g.Use(Slate)
	var feet [4]Vec2
	for k := range feet {
		a := math.Pi/4 + float64(k)*math.Pi/2
		feet[k] = Polar(c, 3.5, a)
		g.Fill(Capsule{feet[k].At(y0 - 0.3), Polar(c, 2.7, a).At(y0 + 8), 0.55})
	}
	for k := range feet {
		a, b := Polar(c, 3.1, math.Pi/4+float64(k)*math.Pi/2), Polar(c, 3.1, math.Pi/4+float64(k+1)*math.Pi/2)
		g.Fill(Capsule{a.At(y0 + 4), b.At(y0 + 4), 0.5})
	}
	g.Use(Stone).Fill(Cylinder{c, 0.7, 0, y0 - 0.3, y0 + 8})
	g.Use(White).FillVolume(Cylinder{c, 3.8, 0, y0 + 8, y0 + 12.6})
	g.Paint(Water, Cylinder{c, 4.2, 0, y0 + 9.6, y0 + 10.9})
	g.Use(Slate).FillVolume(Frustum{c, 4.2, 0.5, y0 + 12.6, y0 + 14.8})
	g.Use(Red).Fill(Sphere(c.At(y0+15.1), 0.6))
}

// logPeriodic is a log-periodic dipole array on a mast, aimed like the dishes.
func (s *radioObservatorySite) logPeriodic(c Vec2) {
	g := s.g
	y0 := s.H(c.X, c.Y)
	top := c.At(y0 + 8.5)
	g.Use(Stone).Fill(Cylinder{c, 1.6, 0, y0 - 0.3, y0 + 0.5})
	g.Fill(Capsule{c.At(y0), top, 0.6})
	f := radioObservatoryAim(top, radioObservatoryAzimuth, radioObservatoryZenith)
	g.Use(Slate).Fill(Capsule{f.at(0, 0, -3.6), f.at(0, 0, 8.2), 0.55})
	for i := 0; i < 6; i++ {
		w := -3 + 2.1*float64(i)
		half := 5.2 * math.Pow(0.78, float64(i))
		g.Use(White).Fill(Capsule{f.at(-half, 0, w), f.at(half, 0, w), 0.5})
	}
}

// helix is a pair of helical antennas on one mount.
func (s *radioObservatorySite) helix(c Vec2) {
	g := s.g
	y0 := s.H(c.X, c.Y)
	top := c.At(y0 + 4.2)
	g.Use(Stone).Fill(Cylinder{c, 1.8, 0, y0 - 0.3, y0 + 0.5})
	g.Use(Slate).Fill(Capsule{c.At(y0), top, 0.7})
	f := radioObservatoryAim(top, radioObservatoryAzimuth, radioObservatoryZenith)
	g.Fill(Capsule{f.at(-3.2, 0, 0), f.at(3.2, 0, 0), 0.6})
	for _, su := range []float64{-1, 1} {
		u0 := su * 3.2
		g.Use(Slate).Fill(f.solid(u0-2.3, u0+2.3, -2.3, 2.3, 0.3, 1.2, func(u, v, w float64) bool {
			return math.Hypot(u-u0, v) <= 2.3
		}))
		g.Use(Charcoal).Fill(Capsule{f.at(u0, 0, 1), f.at(u0, 0, 8.6), 0.5})
		g.Use(Gold)
		for i := 0; i <= 48; i++ {
			t := float64(i) / 48
			sa, ca := math.Sincos(t * 3.5 * 2 * math.Pi)
			g.Fill(Sphere(f.at(u0+1.15*ca, 1.15*sa, 1.5+6.6*t), 0.5))
		}
	}
}

// pickup is a pickup truck facing the given plan angle.
func (s *radioObservatorySite) pickup(c Vec2, angle float64, paint Color) {
	g := s.g
	y := s.H(c.X, c.Y)
	put := func(col Color, a, b Vec3) {
		g.Use(col).Fill(Translate(RotateY(Block(a, b), P(0, 0), angle), c.At(y)))
	}
	for _, wx := range []float64{-1.7, 1.7} {
		put(Charcoal, V(wx-0.5, 0, -1.25), V(wx+0.5, 0.9, 1.25))
	}
	put(paint, V(-2.8, 0.5, -1.2), V(2.8, 1.4, 1.2))
	put(paint, V(-0.2, 1.4, -1.1), V(1.6, 2.4, 1.1))
	put(Glass, V(-0.3, 1.5, -1.2), V(1.7, 2.0, 1.2))
	g.Carve(Translate(RotateY(Block(V(-2.4, 1.0, -0.7), V(-0.7, 1.5, 0.7)), P(0, 0), angle), c.At(y)))
	put(Stone, V(2.8, 0.5, -1.2), V(3.0, 0.9, 1.2))
}

// powerLine brings wooden poles and a sagging wire along the northern rim to a small
// substation behind the great dish.
func (s *radioObservatorySite) powerLine() {
	g := s.g
	const inset = 3.0
	sx := radioObservatoryHalfX - radioObservatoryCorner
	sz := radioObservatoryHalfZ - radioObservatoryCorner
	r := radioObservatoryCorner - inset
	var poles []Vec2
	for _, a := range []float64{-18, -47, -74} {
		poles = append(poles, Polar(P(sx, -sz), r, a*math.Pi/180))
	}
	poles = append(poles, P(17, inset-radioObservatoryHalfZ), P(3, inset-radioObservatoryHalfZ))
	wire := func(a, b Vec3, sag float64) {
		const n = 6
		prev := a
		for k := 1; k <= n; k++ {
			t := float64(k) / n
			p := radioObservatoryMix(a, b, t).Add(V(0, -sag*4*t*(1-t), 0))
			g.Use(Charcoal).Fill(Capsule{prev, p, 0.5})
			prev = p
		}
	}
	var tops []Vec3
	for i, q := range poles {
		y := s.H(q.X, q.Y)
		along := poles[min(i+1, len(poles)-1)].Sub(poles[max(i-1, 0)])
		across := P(-along.Y, along.X).Scale(1.5 / along.Length())
		g.Use(Wood).Fill(Capsule{q.At(y - 0.3), q.At(y + 10), 0.55})
		g.Fill(Capsule{q.Sub(across).At(y + 8.8), q.Add(across).At(y + 8.8), 0.5})
		g.Use(White).Fill(Sphere(q.Sub(across).At(y+9.5), 0.5), Sphere(q.Add(across).At(y+9.5), 0.5))
		tops = append(tops, q.At(y+10.2))
	}
	for i := 0; i+1 < len(tops); i++ {
		wire(tops[i], tops[i+1], 1.6)
	}
	// Substation: two transformers on a pad, fed from the last pole.
	const y0 = radioObservatoryTop
	pad := P(3.5, -41)
	g.Use(Stone).Fill(BlockAt(pad, 8, 5.5, y0-0.2, y0+0.5))
	for i, dx := range []float64{-2, 2} {
		q := pad.Add(P(dx, 0))
		g.Use(Slate).Fill(BlockAt(q, 2.6, 3, y0+0.5, y0+3.2))
		g.Paint(Charcoal, BlockAt(q.Add(P(0, 1.5)), 2.6, 0.5, y0+0.9, y0+2.6))
		g.Use(Gold).Fill(BlockAt(q, 2.6, 3, y0+3.2, y0+3.5))
		g.Use(White).Fill(Capsule{q.At(y0 + 3.5), q.At(y0 + 4.6), 0.5})
		if i == 0 {
			wire(tops[len(tops)-1], q.At(y0+4.8), 0.9)
		}
	}
}

// figure is a standing person in a hard hat, for scale.
func (s *radioObservatorySite) figure(c Vec2, coat Color) {
	g := s.g
	y := s.H(c.X, c.Y)
	g.Use(coat).Fill(Capsule{c.At(y + 0.3), c.At(y + 1.3), 0.5})
	g.Use(White).Fill(Sphere(c.At(y+2.0), 0.45))
}

// cactus is a saguaro with one or two upturned arms.
func (s *radioObservatorySite) cactus(c Vec2, seed int) {
	g := s.g
	y := s.H(c.X, c.Y)
	h := 3.6 + 2.2*Hash01(seed, 1)
	g.Use(Foliage).Fill(Capsule{c.At(y - 0.3), c.At(y + h), 0.6})
	arms := 1 + int(2*Hash01(seed, 2))
	for k := 0; k < arms; k++ {
		a := Hash01(seed, 3)*math.Pi + float64(k)*math.Pi
		ay := y + 1.4 + 1.1*Hash01(seed, 4+k)
		elbow := Polar(c, 1.5, a)
		g.Fill(Capsule{c.At(ay), elbow.At(ay), 0.5}, Capsule{elbow.At(ay), elbow.At(ay + 1.3 + Hash01(seed, 6+k)), 0.5})
	}
}

// scrub scatters creosote dots, yuccas, saguaros and a few boulders over open ground.
func (s *radioObservatorySite) scrub() {
	g := s.g
	open := func(x, z, margin float64) bool {
		ce := s.rim(x, z)
		if ce < 2.6 || math.Abs(ce-radioObservatoryRoad) < radioObservatoryRoadW+margin || s.built(x, z) < margin {
			return false
		}
		if _, q := s.mesa(x, z); q > -4 && q < 2.5 || P(x, z).Sub(radioObservatoryMesa).Length() < 7.5 {
			return false
		}
		return radioObservatoryNear(P(x, z), s.wash) > 3.2+margin && s.slope(x, z) < 0.6
	}
	n := 0
	for i := -54; i <= 54; i += 4 {
		for k := -50; k <= 50; k += 4 {
			x, z := float64(i)+3.2*Hash01(i, k, 201), float64(k)+3.2*Hash01(i, k, 202)
			roll := Hash01(i, k, 203)
			if roll > 0.72 || !open(x, z, 1.2) {
				continue
			}
			y := s.H(x, z)
			switch {
			case roll < 0.07 && open(x, z, 3):
				s.cactus(P(x, z), n)
			case roll < 0.12:
				// Yucca: a short trunk, a spiky head and a pale bloom stalk.
				g.Use(Wood).Fill(Capsule{V(x, y-0.2, z), V(x, y+1.2, z), 0.5})
				g.Use(Foliage).Fill(Sphere(V(x, y+1.9, z), 1.1))
				g.Use(Sand).Fill(Capsule{V(x, y+2.6, z), V(x, y+4.0, z), 0.5})
				g.Use(White).Fill(Sphere(V(x, y+4.3, z), 0.55))
			case roll < 0.17:
				col := DarkStone
				if Hash01(i, k, 204) < 0.5 {
					col = Brick
				}
				r := 0.8 + 0.7*Hash01(i, k, 205)
				g.Use(col).FillVolume(Ellipsoid{V(x, y+0.2, z), V(r*1.3, r*0.9, r)})
			default:
				col := Foliage
				if Hash01(i, k, 206) < 0.25 {
					col = Grass
				}
				g.Use(col).Fill(Sphere(V(x, y+0.35, z), 0.55+0.35*Hash01(i, k, 207)))
			}
			n++
		}
	}
}

func (RadioObservatory) Build(g *Grid) Order {
	s := newRadioObservatorySite(g)
	s.terrain()
	s.telescope()
	s.railway()
	s.hall()
	s.dome()
	s.tank(radioObservatoryTank)
	s.logPeriodic(radioObservatoryLog)
	s.helix(radioObservatoryHelix)
	s.pickup(P(12, 36.5), -math.Pi/2, Red)
	s.pickup(P(15, 36.2), -math.Pi/2, Navy)
	s.pickup(P(35, 9), math.Pi/2, White)
	s.pickup(P(-3.5, 15.5), 1.9, Gold)
	s.pickup(P(-10, 47.2), 0, White)
	for i, q := range []Vec2{P(-2, 44), P(-7, 44.5), P(17.5, 41), P(-6.5, 12.5), P(-1, 13), P(-27.5, 31.2), P(33.5, 13)} {
		s.figure(q, []Color{Red, Navy, Gold}[i%3])
	}
	s.powerLine()
	s.scrub()
	return Outward(radioObservatoryBig)
}
