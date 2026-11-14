package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// SpaceStation is an ISS-style orbital station floating above the curve of the Earth.
// The integrated truss, an open four-chord lattice, runs along X; beyond a rotary joint
// at each end its outboard section is turned about the truss and carries four solar array
// wings (eight in all), navy blankets with a gold cell grid either side of a white mast,
// all feathered toward the sun so the +Z wings dip toward the Earth and the -Z wings
// rise. Pleated white radiators trail aft (-Z) from the inboard truss. The pressurised
// spine runs along Z under the middle of the truss: a crew capsule docked at the forward
// (+Z) port, a node with two lateral laboratories and a porch of exposed payloads, the
// main laboratory, a second node with an airlock and a cupola module, then the darker
// service modules with their own small arrays and, at the aft end, a cargo vehicle with
// two round solar fans. A robotic arm folds high over the truss with a spacewalker riding
// its tip, and a second spacewalker drifts on a gold tether below the porch. Nothing
// joins the station to the base: a broad shallow spherical cap of Earth, a hollow skin of
// ocean, continents and islands under a cyclone, frontal bands and fair-weather cloud,
// with a glass atmosphere rim at the limb.
type SpaceStation struct{}

func init() { Register(SpaceStation{}) }

func (SpaceStation) Name() string  { return "space_station" }
func (SpaceStation) Title() string { return "Space Station" }

const (
	spaceStationEarthR = 68.0  // Plan radius of the Earth cap.
	spaceStationSphere = 122.0 // Radius of the sphere the cap is cut from.
	spaceStationEarthZ = -8.0  // Plan Z of the cap's axis, under the middle of the spine.
	spaceStationSkin   = 0.95  // Thickness of the Earth's skin.
	spaceStationLimb   = 2.2   // Plan width of the atmosphere over the limb.
	spaceStationTrussY = 57.0  // Truss axis.
	spaceStationHabY   = 50.9  // Module axis.
	spaceStationHabR   = 3.5   // Radius of the large modules.
	spaceStationJoint  = 24.0  // Inboard face of the rotary joints.
	spaceStationInner  = 35.0  // Truss station of the inner array wings.
	spaceStationOuter  = 50.0  // Truss station of the outer array wings.
	spaceStationTip    = 53.6  // Truss ends.
	spaceStationRoot   = 6.4   // Distance from the truss axis to an array blanket.
	spaceStationWing   = 34.0  // Blanket length of an array wing.
	spaceStationChord  = 6.0   // Half-width of an array wing.
	spaceStationPanel  = 0.58  // Half-thickness of an array blanket.

	spaceStationAlpha = 25 * math.Pi / 180 // Rotation of the arrays about the truss.
	spaceStationBeta  = 20 * math.Pi / 180 // Feathering of each wing about its mast.
)

// spaceStationBlob is one lobe of land: its center, plan radii and strength.
type spaceStationBlob struct{ X, Z, A, B, S float64 }

// spaceStationFront is an arc of frontal cloud: the arc's center and radius, its angular
// range, half-width and raised height.
type spaceStationFront struct{ X, Z, R, A0, A1, W, H float64 }

var (
	spaceStationStorm = P(20, 26) // Eye of the cyclone.
	spaceStationLands = []spaceStationBlob{
		// The northern continent across the far limb, with a peninsula reaching south.
		{-38, -42, 24, 18, 1.0}, {-12, -60, 22, 12, 1.0}, {-56, -16, 12, 18, 0.9}, {-22, -20, 8, 12, 0.85},
		{-16, -5, 5, 7, 0.8},
		// The western landmass on the near limb.
		{-46, 30, 13, 15, 1.0}, {-34, 44, 12, 8, 0.9},
		// The north-eastern coast.
		{40, -54, 18, 11, 1.0}, {58, -32, 8, 12, 0.9},
		// The eastern island arc and two southern islets.
		{48, 4, 5, 7, 0.9}, {54, -10, 4, 5, 0.85}, {43, 19, 3.5, 4, 0.8}, {4, 50, 3, 2.5, 0.8}, {-8, 55, 2.5, 2, 0.8},
	}
	spaceStationFronts = []spaceStationFront{
		{-75, 22, 52, -0.5, 0.35, 2.4, 1.1},
		{96, -50, 62, math.Pi - 0.55, math.Pi + 0.2, 2.4, 1.1},
	}
)

func spaceStationCross(a, b Vec3) Vec3 {
	return V(a.Y*b.Z-a.Z*b.Y, a.Z*b.X-a.X*b.Z, a.X*b.Y-a.Y*b.X)
}

func spaceStationUnit(a Vec3) Vec3 { return a.Scale(1 / a.Length()) }

func spaceStationSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

// spaceStationNoise is smooth value noise in [-1, 1].
func spaceStationNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return (a+(b-a)*sx+(c+(d-c)*sx-a-(b-a)*sx)*sz)*2 - 1
}

func spaceStationFBM(x, z float64, seed int) float64 {
	return 0.6*spaceStationNoise(x, z, seed) + 0.3*spaceStationNoise(x*2.03+17, z*2.03+5, seed+1) +
		0.1*spaceStationNoise(x*4.1+3, z*4.1+29, seed+2)
}

// spaceStationTri is a triangle wave of period 1 in [-1, 1].
func spaceStationTri(t float64) float64 { return 4*math.Abs(t-math.Floor(t)-0.5) - 1 }

// spaceStationFrame is a local frame: U runs along a part, V across it and W is its normal.
type spaceStationFrame struct{ O, U, V, W Vec3 }

func (f spaceStationFrame) local(p Vec3) (u, v, w float64) {
	d := p.Sub(f.O)
	return d.Dot(f.U), d.Dot(f.V), d.Dot(f.W)
}

func (f spaceStationFrame) at(u, v, w float64) Vec3 {
	return f.O.Add(f.U.Scale(u)).Add(f.V.Scale(v)).Add(f.W.Scale(w))
}

// box bounds a local cuboid in world space.
func (f spaceStationFrame) box(u0, u1, v0, v1, w0, w1 float64) Box {
	b := EmptyBox()
	for _, u := range []float64{u0, u1} {
		for _, v := range []float64{v0, v1} {
			for _, w := range []float64{w0, w1} {
				p := f.at(u, v, w)
				b = b.Union(Box{p, p})
			}
		}
	}
	return b
}

// slab is a local cuboid.
func (f spaceStationFrame) slab(u0, u1, v0, v1, w0, w1 float64) Solid {
	return Func{f.box(u0, u1, v0, v1, w0, w1), func(p Vec3) bool {
		u, v, w := f.local(p)
		return u >= u0 && u <= u1 && v >= v0 && v <= v1 && w >= w0 && w <= w1
	}}
}

// can is a round body along U, from radius r0 at u0 to r1 at u1.
func (f spaceStationFrame) can(u0, u1, r0, r1 float64) Solid {
	rm := math.Max(r0, r1)
	return Func{f.box(u0, u1, -rm, rm, -rm, rm), func(p Vec3) bool {
		u, v, w := f.local(p)
		if u < u0 || u > u1 {
			return false
		}
		rr := r0 + (r1-r0)*(u-u0)/(u1-u0)
		return v*v+w*w <= rr*rr
	}}
}

// spaceStationRig carries the grid and the frame shared by all eight array wings: e runs
// along the +Z wings (dipping toward the Earth), n0 is the outboard truss's up, v runs
// across a blanket and n is its sunward normal.
type spaceStationRig struct {
	g           *Grid
	e, n0, v, n Vec3
}

func spaceStationNewRig(g *Grid) *spaceStationRig {
	sa, ca := math.Sincos(spaceStationAlpha)
	sb, cb := math.Sincos(spaceStationBeta)
	e, n0, x := V(0, -sa, ca), V(0, ca, sa), V(1, 0, 0)
	return &spaceStationRig{g: g, e: e, n0: n0, v: x.Scale(cb).Sub(n0.Scale(sb)), n: n0.Scale(cb).Add(x.Scale(sb))}
}

func (SpaceStation) Build(g *Grid) Order {
	r := spaceStationNewRig(g)
	r.earth()
	r.girder()
	r.modules()
	r.vehicles()
	r.arrays()
	r.radiators()
	r.arm()
	r.hollow(2)
	return Outward(P(0, 0))
}

func (r *spaceStationRig) put(c Cell, col Color) {
	r.g.Remove(c)
	r.g.Use(col).Add(c)
}

// each visits every lattice cell whose center lies in the box.
func (r *spaceStationRig) each(b Box, f func(c Cell, p Vec3)) {
	h := r.g.Span() / 2
	lo := func(v float64) int32 { return int32(math.Ceil(v / h)) }
	hi := func(v float64) int32 { return int32(math.Floor(v / h)) }
	for y := lo(b.Min.Y); y <= hi(b.Max.Y); y++ {
		z0, x0 := lo(b.Min.Z), lo(b.Min.X)
		if (z0^y)&1 != 0 {
			z0++
		}
		if (x0^y)&1 != 0 {
			x0++
		}
		for z := z0; z <= hi(b.Max.Z); z += 2 {
			for x := x0; x <= hi(b.Max.X); x += 2 {
				c := Cell{X: x, Y: y, Z: z}
				f(c, r.g.World(c))
			}
		}
	}
}

// set gives every lattice cell in the box for which f reports a color that color, adding
// missing cells. Unlike Grid.Fill it never snaps centers to the half-unit grid, so thin
// tilted panels and one-cell lines stay crisp.
func (r *spaceStationRig) set(b Box, f func(p Vec3) (Color, bool)) {
	r.each(b, func(c Cell, p Vec3) {
		if col, ok := f(p); ok {
			r.put(c, col)
		}
	})
}

// tint recolors the occupied cells in the box for which f reports a color.
func (r *spaceStationRig) tint(b Box, f func(p Vec3) (Color, bool)) {
	r.each(b, func(c Cell, p Vec3) {
		if !r.g.Has(c) {
			return
		}
		if col, ok := f(p); ok {
			r.put(c, col)
		}
	})
}

// fill occupies the cells whose centers lie inside the solids.
func (r *spaceStationRig) fill(col Color, solids ...Solid) {
	for _, s := range solids {
		r.set(s.Bounds(), func(p Vec3) (Color, bool) { return col, s.Contains(p) })
	}
}

// round occupies the cells mostly inside the solids, for clean round bodies.
func (r *spaceStationRig) round(col Color, solids ...Solid) {
	r.g.With(col, func() { r.g.FillVolume(solids...) })
}

// hollow removes cells buried deeper than keep layers so solid bodies become shells.
func (r *spaceStationRig) hollow(keep int) {
	g := r.g
	inner := map[Cell]bool{}
	for _, c := range g.Cells() {
		all := true
		for _, n := range Neighbors {
			if !g.Has(c.Add(n)) {
				all = false
				break
			}
		}
		if all {
			inner[c] = true
		}
	}
	for k := 1; k < keep; k++ {
		next := map[Cell]bool{}
		for c := range inner {
			all := true
			for _, n := range Neighbors {
				if !inner[c.Add(n)] {
					all = false
					break
				}
			}
			if all {
				next[c] = true
			}
		}
		inner = next
	}
	for c := range inner {
		g.Remove(c)
	}
}

// field is the land field at a plan point: positive on land, negative at sea.
func (r *spaceStationRig) field(x, z float64) float64 {
	f := 0.3*spaceStationFBM(x*0.06, z*0.06, 21) - 0.5
	for _, b := range spaceStationLands {
		dx, dz := (x-b.X)/b.A, (z-b.Z)/b.B
		f += b.S * math.Exp(-1.2*(dx*dx+dz*dz))
	}
	return f
}

// surface returns the color of the Earth's surface at a plan point and how far it
// stands above the ocean sphere.
func (r *spaceStationRig) surface(x, z float64) (Color, float64) {
	f := r.field(x, z)
	if f <= 0 {
		if f > -0.09 {
			return Glass, 0 // Shallow shelf seas.
		}
		return Water, 0
	}
	if f < 0.05 {
		return Sand, 0.3 // Beaches.
	}
	relief := 0.45 + 0.4*spaceStationSmooth(f/0.3)
	// Mountain chains follow the zero lines of a broad noise, with snow along the crests.
	if chain := math.Abs(spaceStationNoise(x*0.075+2, z*0.075+9, 63)); chain < 0.085 && f > 0.14 {
		if chain < 0.02 && f > 0.35 {
			return White, relief + 1.0
		}
		return Stone, relief + 0.6
	}
	arid := spaceStationFBM(x*0.04+5, z*0.04+2, 61) + 0.004*(z+10)
	lush := spaceStationFBM(x*0.07+1, z*0.07+8, 62)
	switch {
	case f > 0.12 && arid > 0.1:
		if spaceStationFBM(x*0.11+8, z*0.11+3, 64) > 0.2 {
			return Grass, relief // Scrub breaking up the deserts.
		}
		return Sand, relief
	case lush > 0.05:
		return Foliage, relief
	}
	return Grass, relief
}

// cloud returns the depth of cloud over a plan point: negative under clear sky, zero for
// thin cloud lying on the surface and positive for decks raised above it.
func (r *spaceStationRig) cloud(x, z float64) float64 {
	best := -1.0
	rag := spaceStationNoise(x*0.3, z*0.3, 41)
	// The cyclone: a clear eye, a dense central overcast and two spiral bands that thin
	// and sink as they trail outward.
	d := P(x, z).Sub(spaceStationStorm)
	rho := d.Length()
	switch {
	case rho < 1.3:
	case rho < 4.6:
		best = 1.9 - 0.22*rho
	case rho < 22:
		const twist = 5.0
		phase := math.Remainder(2*d.Angle()+twist*math.Log(rho)+1.0, 2*math.Pi)
		across := math.Abs(phase) * rho / math.Sqrt(twist*twist+4)
		width := (1.9 - 0.06*(rho-4.6)) * (0.8 + 0.4*rag) * spaceStationSmooth((22-rho)/4)
		if across < width {
			best = math.Max(0, (1.3-0.08*rho)*(1-across/width))
		}
	}
	// Frontal bands: long comma-shaped arcs, tallest along their spine.
	for _, f := range spaceStationFronts {
		q := P(x-f.X, z-f.Z)
		a := q.Angle()
		if a < f.A0-1 {
			a += 2 * math.Pi
		}
		t := (a - f.A0) / (f.A1 - f.A0)
		if t <= 0 || t >= 1 {
			continue
		}
		width := f.W * math.Sqrt(math.Sin(math.Pi*t)) * (0.45 + 0.9*t) * (0.8 + 0.45*rag)
		if k := math.Abs(q.Length()-f.R) / math.Max(width, 1e-6); k < 1 {
			best = math.Max(best, math.Max(0, f.H*(1.25-1.6*k*k)-0.2))
		}
	}
	if best >= 0 || rho < 26 || r.field(x, z) > 0.1 {
		return best
	}
	// Fair-weather cumulus over the sea and the coasts: small clumps gathered into broad
	// fields, the largest of them standing a little proud of the surface.
	a, b := 0.88*x-0.48*z, 0.48*x+0.88*z
	if spaceStationFBM(x*0.03+3, z*0.03+9, 44) > -0.08 {
		if puff := spaceStationFBM(a*0.2+7, b*0.27+1, 45) - 0.32; puff > 0 {
			return math.Min(0.6, 3*puff)
		}
	}
	return best
}

// earth builds the cap column by column: a skin under the sphere, land a little proud of
// the ocean, cloud on top and the glass atmosphere over and just beyond the limb.
func (r *spaceStationRig) earth() {
	const floor, air = -0.4, 0.9
	A, rs := spaceStationEarthR, spaceStationSphere
	base := math.Sqrt(rs*rs - A*A) // Height of the rim plane above the sphere's center.
	h := r.g.Span() / 2
	n := int32(math.Ceil((A + 3) / h))
	cz := int32(math.Round(spaceStationEarthZ / h))
	for ix := -n; ix <= n; ix++ {
		for iz := cz - n; iz <= cz+n; iz++ {
			if (ix^iz)&1 != 0 {
				continue
			}
			x, z := float64(ix)*h, float64(iz)*h
			rho := math.Hypot(x, z-spaceStationEarthZ)
			if rho > A+1.6 {
				continue
			}
			column := func(y0, y1 float64, col Color) {
				lo, hi := int32(math.Ceil(y0/h)), int32(math.Floor(y1/h))
				if (lo^ix)&1 != 0 {
					lo++
				}
				for iy := lo; iy <= hi; iy += 2 {
					r.put(Cell{X: ix, Y: iy, Z: iz}, col)
				}
			}
			if rho > A {
				column(floor, air, Glass)
				continue
			}
			col, relief := r.surface(x, z)
			top := math.Sqrt(rs*rs-rho*rho) - base + relief
			// The skin keeps its thickness along the normal as the dome steepens.
			skin := spaceStationSkin * rs / math.Sqrt(rs*rs-rho*rho)
			column(math.Max(top-skin, floor), top, col)
			if rho > A-spaceStationLimb {
				column(top, top+air, Glass)
				continue
			}
			if depth := r.cloud(x, z); depth >= 0 {
				column(top-0.8, top+depth, White)
			}
		}
	}
}

// truss fills an open four-chord lattice girder from a to b: chords at ±hu along up and
// ±hv across it, battens at every node and one diagonal per bay on each face. Opposite
// faces are laced alike, so the far side lines up behind the near side.
func (r *spaceStationRig) truss(a, b, up Vec3, hu, hv float64, n int, chord, brace float64, cc, bc Color) {
	d := spaceStationUnit(b.Sub(a))
	v := spaceStationUnit(spaceStationCross(d, up))
	u := spaceStationCross(v, d)
	at := func(i int, su, sv float64) Vec3 {
		return a.Add(b.Sub(a).Scale(float64(i) / float64(n))).Add(u.Scale(su * hu)).Add(v.Scale(sv * hv))
	}
	for i := 0; i <= n; i++ {
		r.fill(bc, Capsule{at(i, 1, 1), at(i, -1, 1), brace}, Capsule{at(i, 1, -1), at(i, -1, -1), brace},
			Capsule{at(i, 1, 1), at(i, 1, -1), brace}, Capsule{at(i, -1, 1), at(i, -1, -1), brace})
	}
	for i := 0; i < n; i++ {
		s := 1.0
		if i%2 == 1 {
			s = -1
		}
		for _, k := range []float64{1, -1} {
			r.fill(bc, Capsule{at(i, s, k), at(i+1, -s, k), brace}, Capsule{at(i, k, s), at(i+1, k, -s), brace})
		}
	}
	for _, su := range []float64{1, -1} {
		for _, sv := range []float64{1, -1} {
			r.fill(cc, Capsule{at(0, su, sv), at(n, su, sv), chord})
		}
	}
}

// girder builds the truss: the fixed inboard lattice with its equipment bays over the
// laboratory, a rotary joint at each end and the turned outboard lattices that carry the
// array wings.
func (r *spaceStationRig) girder() {
	T, J := spaceStationTrussY, spaceStationJoint
	r.truss(V(-J, T, 0), V(J, T, 0), V(0, 1, 0), 2.2, 2.4, 8, 0.6, 0.5, Stone, DarkStone)
	bay := func(x0, x1, hy, hz float64, col Color) {
		r.fill(col, Block(V(x0, T-hy, -hz), V(x1, T+hy, hz)))
	}
	bay(-23, -19, 1.3, 1.5, Gold)
	bay(-17, -13, 1.4, 1.6, White)
	bay(-6, 6, 1.9, 2.1, Stone)
	bay(13, 17, 1.4, 1.6, White)
	bay(19, 23, 1.3, 1.5, Gold)
	// Struts down to the laboratory.
	for _, sx := range []float64{1, -1} {
		for _, sz := range []float64{1, -1} {
			r.fill(DarkStone, Capsule{V(5.5*sx, T-2.2, 2.4*sz), V(1.6*sx, spaceStationHabY+2.2, 2.0*sz), 0.5})
		}
	}
	// Pallets of spares on top of the truss.
	for i, x := range []float64{-15, 15} {
		r.fill(Stone, Block(V(x-2.2, T+2.8, -2.6), V(x+2.2, T+3.5, 2.6)))
		r.fill([]Color{Gold, White}[i], Block(V(x-1.8, T+3.5, -2.2), V(x-0.2, T+5.0, -0.4)))
		r.fill([]Color{White, Gold}[i], Block(V(x+0.2, T+3.5, 0.3), V(x+1.9, T+4.7, 2.2)))
		r.fill(Charcoal, Block(V(x-1.6, T+3.5, 0.6), V(x-0.4, T+4.3, 1.9)))
	}
	// Communications mast and dish to starboard.
	mast, aim := V(20.5, T+7.6, -1.2), spaceStationUnit(V(0.45, 1, 0.5))
	r.fill(Stone, Capsule{V(mast.X, T+2, mast.Z), mast, 0.55})
	r.fill(White, r.dish(mast.Add(aim.Scale(0.6)), aim, 2.5, 1.0))
	r.fill(Charcoal, Sphere(mast.Add(aim.Scale(2.2)), 0.5))

	for _, side := range []float64{1, -1} {
		f := spaceStationFrame{V(0, T, 0), V(side, 0, 0), V(0, 0, 1), V(0, 1, 0)}
		// Rotary joint: a dark drum with a gold race ring.
		r.round(DarkStone, f.can(J, J+2.6, 3.4, 3.4))
		r.tint(f.box(J, J+2.6, -3.5, 3.5, -3.5, 3.5), func(p Vec3) (Color, bool) {
			u, _, _ := f.local(p)
			return Gold, u > J+0.9 && u < J+1.7
		})
		r.truss(f.at(J+2.6, 0, 0), f.at(spaceStationTip, 0, 0), r.n0, 2.0, 2.2, 5, 0.6, 0.5, Stone, DarkStone)
		// Equipment assemblies at the wing stations.
		for _, x := range []float64{spaceStationInner, spaceStationOuter} {
			fi := spaceStationFrame{f.at(x, 0, 0), f.U, r.e, r.n0}
			r.fill(White, fi.slab(-2.2, 2.2, -2.5, 2.5, -2.3, 2.3))
			r.tint(fi.box(-2.2, 2.2, -2.5, 2.5, -2.3, 2.3), func(p Vec3) (Color, bool) {
				u, _, _ := fi.local(p)
				return Gold, math.Abs(u) < 0.5
			})
		}
	}
}

// dish is a shallow paraboloid shell opening along axis.
func (r *spaceStationRig) dish(c, axis Vec3, radius, depth float64) Solid {
	axis = spaceStationUnit(axis)
	return Func{Box{c, c}.Grow(radius + depth + 1), func(p Vec3) bool {
		d := p.Sub(c)
		a := d.Dot(axis)
		rr := d.Sub(axis.Scale(a)).Length()
		if rr > radius {
			return false
		}
		cup := depth * rr * rr / (radius * radius)
		return a >= cup-0.5 && a <= cup+0.5
	}}
}

// wing builds one solar array wing from the truss station at x: dir +1 runs toward +Z
// and dips, -1 runs toward -Z and rises.
func (r *spaceStationRig) wing(x, dir float64) {
	const pitch = spaceStationWing / 8 // Spacing of the gold rows across a blanket.
	f := spaceStationFrame{V(x, spaceStationTrussY, 0), r.e.Scale(dir), r.v, r.n}
	u0, half := spaceStationRoot, spaceStationChord
	u1 := u0 + spaceStationWing
	// Mast canister, then the blanket boxes across the root and the tip.
	r.round(Stone, f.can(2.0, u0-0.8, 1.0, 1.0))
	r.fill(White, f.slab(u0-1.2, u0, -half-0.2, half+0.2, -0.75, 0.75),
		f.slab(u1, u1+0.9, -half-0.2, half+0.2, -0.75, 0.75))
	t := spaceStationPanel
	r.set(f.box(u0, u1, -half, half, -t, t), func(p Vec3) (Color, bool) {
		u, v, w := f.local(p)
		if u < u0 || u > u1 || math.Abs(v) > half || math.Abs(w) > t {
			return 0, false
		}
		switch {
		case math.Abs(v) < 0.7:
			return White, true // The mast between the two blankets.
		case w < -0.25:
			return Gold, true // The blankets' gold backs.
		}
		k := (u - u0) / pitch
		if math.Abs(k-math.Round(k))*pitch < 0.36 && k > 0.5 && k < 7.5 {
			return Gold, true
		}
		if math.Abs(math.Abs(v)-3.4) < 0.36 {
			return Gold, true
		}
		return Navy, true
	})
}

// arrays hangs the eight wings and a small pleated radiator between each pair of stations.
func (r *spaceStationRig) arrays() {
	for _, side := range []float64{1, -1} {
		for _, x := range []float64{spaceStationInner, spaceStationOuter} {
			r.wing(side*x, 1)
			r.wing(side*x, -1)
		}
		x := side * (spaceStationInner + spaceStationOuter) / 2
		f := spaceStationFrame{V(x, spaceStationTrussY, 0), r.n0.Scale(-1), V(1, 0, 0), r.e}
		r.pleated(f, 2.4, 14, 1.7, 2.4)
	}
}

// pleated fills a zigzag radiator panel in a frame: along U from u0 to u1, half-width
// half across V, folded along W with the given pleat period.
func (r *spaceStationRig) pleated(f spaceStationFrame, u0, u1, half, period float64) {
	amp := period * 0.16
	r.set(f.box(u0, u1, -half, half, -amp-0.6, amp+0.6), func(p Vec3) (Color, bool) {
		u, v, w := f.local(p)
		if u < u0 || u > u1 || math.Abs(v) > half {
			return 0, false
		}
		t := (u - u0) / period
		if math.Abs(w-amp*spaceStationTri(t)) > 0.6 {
			return 0, false
		}
		if t-math.Floor(t) < 0.5 {
			return White, true
		}
		return Stone, true
	})
}

// radiators trails two sets of three pleated panels aft from the inboard truss.
func (r *spaceStationRig) radiators() {
	T := spaceStationTrussY
	for _, side := range []float64{1, -1} {
		r.fill(DarkStone, Block(V(side*12.2, T-0.9, -4.4), V(side*23.8, T+0.9, -2.8)))
		for i := 0; i < 3; i++ {
			x := side * (14.1 + 3.9*float64(i))
			f := spaceStationFrame{V(x, T, -4.4), V(0, 0, -1), V(1, 0, 0), V(0, 1, 0)}
			r.pleated(f, 0, 23.4, 1.5, 5.2)
		}
	}
}

// module is one pressurised can along f.U: a darker, narrower berthing collar shows at
// each end, trim rings band the shell every pitch units and, when asked, a dashed dark
// window band runs along both flanks.
func (r *spaceStationRig) module(f spaceStationFrame, u0, u1, rad float64, body, trim Color, pitch float64, windows bool) {
	r.round(DarkStone, f.can(u0, u1, rad-0.9, rad-0.9))
	r.round(body, f.can(u0+0.8, u1-0.8, rad, rad))
	r.tint(f.box(u0, u1, -rad, rad, -rad, rad), func(p Vec3) (Color, bool) {
		u, v, w := f.local(p)
		if u < u0+0.8 || u > u1-0.8 || math.Hypot(v, w) < rad-1.2 {
			return 0, false
		}
		if windows && math.Abs(w) < 0.6 && u > u0+1.6 && u < u1-1.6 && math.Mod(u-u0, 2.4) < 1.3 {
			return Charcoal, true
		}
		if pitch > 0 && math.Mod(u-u0-0.8, pitch) > pitch-0.7 {
			return trim, true
		}
		return 0, false
	})
}

// port is a stub docking port on a node at u, pointing along axis.
func (r *spaceStationRig) port(f spaceStationFrame, u, rad float64, axis Vec3) {
	c := f.at(u, 0, 0)
	r.fill(DarkStone, Capsule{c.Add(axis.Scale(rad - 0.6)), c.Add(axis.Scale(rad + 0.5)), 1.5})
}

// blanket is a small flat solar array in a frame: navy cells, gold rows and a gold back.
func (r *spaceStationRig) blanket(f spaceStationFrame, u0, u1, half, pitch float64) {
	t := spaceStationPanel
	r.set(f.box(u0, u1, -half, half, -t, t), func(p Vec3) (Color, bool) {
		u, v, w := f.local(p)
		if u < u0 || u > u1 || math.Abs(v) > half || math.Abs(w) > t {
			return 0, false
		}
		k := (u - u0) / pitch
		if w < -0.25 || math.Abs(k-math.Round(k))*pitch < 0.36 {
			return Gold, true
		}
		return Navy, true
	})
}

// modules builds the pressurised spine along Z and its lateral modules.
func (r *spaceStationRig) modules() {
	M, R := spaceStationHabY, spaceStationHabR
	up := V(0, 1, 0)
	fwd := spaceStationFrame{V(0, M, 0), V(0, 0, 1), V(1, 0, 0), up}
	aft := spaceStationFrame{V(0, M, 0), V(0, 0, -1), V(-1, 0, 0), up}
	stbd := func(z float64) spaceStationFrame { return spaceStationFrame{V(0, M, z), V(1, 0, 0), V(0, 0, -1), up} }
	port := func(z float64) spaceStationFrame { return spaceStationFrame{V(0, M, z), V(-1, 0, 0), V(0, 0, 1), up} }

	// The laboratory under the truss and the forward node with its two lateral labs.
	r.module(fwd, -5, 9.5, R, White, Stone, 3.5, true)
	r.module(fwd, 9.5, 17.5, R, White, Stone, 0, false)
	r.port(fwd, 13.5, R, up)
	r.port(fwd, 13.5, R, up.Scale(-1))
	r.module(stbd(13.5), 3.0, 12.6, 3.1, White, Stone, 3.1, true)
	r.fill(Gold, Block(V(12.6, M-1.4, 11.9), V(14.2, M+1.2, 15.1))) // External payload rack.
	r.module(port(13.5), 3.0, 17.4, R, White, Stone, 3.5, true)
	// The logistics module on top of the long lab and its porch of exposed payloads.
	attic := spaceStationFrame{V(-9, M, 13.5), up, V(1, 0, 0), V(0, 0, 1)}
	r.module(attic, 2.6, 8.2, 2.5, White, Stone, 0, false)
	r.fill(Stone, Block(V(-24.6, M-2.0, 10.7), V(-17.4, M-1.0, 16.3)))
	for i, col := range []Color{Gold, White, Charcoal} {
		x := -18.9 - 2.3*float64(i)
		z := 12 + 1.6*float64(i%2)
		r.fill(col, Block(V(x-0.9, M-1.0, z-0.9), V(x+0.9, M+0.8, z+0.9)))
	}
	r.fill(White, Capsule{V(-18, M+1.8, 16), V(-21.5, M+4.2, 17.6), 0.5}, Capsule{V(-21.5, M+4.2, 17.6), V(-24, M+1.4, 15), 0.5})

	// The aft node with the airlock to starboard and the cupola module to port.
	r.module(aft, 5, 12, R, White, Stone, 0, false)
	r.port(aft, 8.5, R, up)
	air := stbd(-8.5)
	r.module(air, 3.0, 7.8, 3.0, White, Stone, 0, false)
	r.round(Stone, air.can(7.8, 11, 1.9, 1.9))
	r.round(DarkStone, air.can(11, 11.6, 1.3, 1.3))
	for _, dz := range []float64{-1.5, 1.5} {
		r.round(Stone, Sphere(V(5.4, M+3.2, -8.5+dz), 1.3))
	}
	r.module(port(-8.5), 3.0, 12.4, 3.2, White, Stone, 3.2, false)
	down := spaceStationFrame{V(-9.4, M, -8.5), up.Scale(-1), V(1, 0, 0), V(0, 0, 1)}
	r.round(DarkStone, down.can(2.8, 3.5, 2.5, 2.5))
	r.round(Glass, down.can(3.5, 5.5, 2.4, 1.5))
	r.tint(down.box(3.5, 5.5, -2.5, 2.5, -2.5, 2.5), func(p Vec3) (Color, bool) {
		_, v, w := down.local(p)
		return DarkStone, math.Abs(math.Sin(3*math.Atan2(w, v))) < 0.2 && math.Hypot(v, w) > 1.0
	})

	// The service modules: darker, stepped, with pale thermal panels and small arrays.
	r.round(Charcoal, aft.can(12, 14.8, 1.9, 2.6))
	r.module(aft, 14.6, 28.4, 3.1, Stone, DarkStone, 3.4, false)
	r.tint(aft.box(16, 27, -3.2, 3.2, -3.2, 3.2), func(p Vec3) (Color, bool) {
		u, _, w := aft.local(p)
		return White, w > 1.5 && math.Mod(u-16, 3.4) < 1.9
	})
	node := aft.at(30.6, 0, 0)
	r.round(Stone, Sphere(node, 2.9))
	r.round(White, Capsule{node.Add(V(0, 2.4, 0)), node.Add(V(0, 5.8, 0)), 1.5})
	r.round(DarkStone, Capsule{node.Add(V(0, -2.4, 0)), node.Add(V(0, -4.8, 0)), 1.5})
	r.module(aft, 32.4, 36.4, 2.5, Stone, DarkStone, 0, false)
	r.module(aft, 35.8, 44, 3.3, Stone, DarkStone, 3.6, true)
	for _, f := range []spaceStationFrame{stbd(-40), port(-40)} {
		r.fill(DarkStone, Capsule{f.at(3.0, 0, 0), f.at(5.2, 0, 0), 0.55})
		r.blanket(f, 5.2, 17.2, 2.1, 3.0)
	}
}

// vehicles docks the crew capsule at the forward port and the cargo vehicle at the aft.
func (r *spaceStationRig) vehicles() {
	const spokes = 6.0 // Ribs on a cargo vehicle fan.
	M := spaceStationHabY
	fwd := spaceStationFrame{V(0, M, 0), V(0, 0, 1), V(1, 0, 0), V(0, 1, 0)}
	aft := spaceStationFrame{V(0, M, 0), V(0, 0, -1), V(-1, 0, 0), V(0, 1, 0)}

	// Crew capsule: docking adapter, gumdrop cabin nose-on to the station, then its trunk
	// with a solar-cell back and two fins.
	r.round(Charcoal, fwd.can(17.5, 20.6, 2.4, 1.6))
	r.round(White, fwd.can(20.6, 25.2, 1.6, 3.0))
	r.round(DarkStone, fwd.can(25.2, 26, 3.0, 3.0))
	r.round(White, fwd.can(26, 31.2, 2.7, 2.7))
	r.tint(fwd.box(20.6, 31.2, -3.1, 3.1, -3.1, 3.1), func(p Vec3) (Color, bool) {
		u, v, w := fwd.local(p)
		switch {
		case u > 26.2 && w > 0.6:
			if math.Mod(u-26.2, 1.7) < 0.45 {
				return Gold, true
			}
			return Navy, true
		case u > 22.4 && u < 24.2 && math.Abs(w-0.5) < 0.6 && math.Abs(v) > 1.4:
			return Charcoal, true
		}
		return 0, false
	})
	for _, sv := range []float64{1, -1} {
		r.fill(Stone, fwd.slab(28, 31.2, math.Min(sv*2.5, sv*3.9), math.Max(sv*2.5, sv*3.9), -0.5, 0.5))
	}

	// Cargo vehicle: white pressurised can, gold service module, engine, and two round
	// solar fans on short arms.
	r.round(DarkStone, aft.can(44, 45.6, 1.7, 1.7))
	r.module(aft, 45.2, 52.6, 2.7, White, Stone, 2.4, false)
	r.round(Gold, aft.can(52.4, 55.4, 2.2, 2.2))
	r.round(Charcoal, aft.can(55.4, 56.4, 1.1, 1.6))
	for _, sv := range []float64{1, -1} {
		c := aft.at(54, sv*7.9, 0)
		r.fill(Stone, Capsule{aft.at(54, sv*1.8, 0), aft.at(54, sv*4.3, 0), 0.5})
		r.set(Box{c, c}.Grow(4), func(p Vec3) (Color, bool) {
			d := p.Sub(c)
			rho := math.Hypot(d.X, d.Z)
			if rho > 3.8 || math.Abs(d.Y) > spaceStationPanel {
				return 0, false
			}
			if rho < 0.9 || d.Y < -0.25 {
				return Gold, true
			}
			a := math.Atan2(d.Z, d.X) / (2 * math.Pi) * spokes
			if math.Abs(a-math.Round(a))*2*math.Pi/spokes*rho < 0.36 {
				return Gold, true
			}
			return Navy, true
		})
	}
}

// walker is a suited spacewalker: head toward up, visor toward face, with a backpack.
func (r *spaceStationRig) walker(at, up, face Vec3, stripe bool) {
	const figure = 1.25 // Scale of a spacewalker.
	up = spaceStationUnit(up)
	side := spaceStationUnit(spaceStationCross(up, face))
	f := spaceStationFrame{at, up.Scale(figure), side.Scale(figure), spaceStationCross(side, up).Scale(figure)}
	legs := White
	if stripe {
		legs = Red
	}
	r.fill(legs, Capsule{f.at(-1.7, -0.5, 0.15), f.at(-0.5, -0.35, 0), 0.45},
		Capsule{f.at(-1.7, 0.5, 0.15), f.at(-0.5, 0.35, 0), 0.45})
	r.fill(White, Capsule{f.at(-0.3, 0, 0), f.at(0.5, 0, 0), 0.8},
		Capsule{f.at(0.5, -0.8, 0), f.at(0.1, -1.3, 0.7), 0.42}, Capsule{f.at(0.5, 0.8, 0), f.at(0.1, 1.3, 0.7), 0.42})
	r.fill(Stone, Capsule{f.at(-0.2, 0, -0.75), f.at(0.8, 0, -0.75), 0.6})
	r.fill(White, Sphere(f.at(1.35, 0, 0), 0.68))
	r.fill(Gold, Sphere(f.at(1.35, 0, 0.5), 0.42))
}

// arm builds the robotic arm, folded high over the truss with a spacewalker riding its
// tip, and a second spacewalker drifting on a tether below the porch.
func (r *spaceStationRig) arm() {
	T, M := spaceStationTrussY, spaceStationHabY
	r.fill(Gold, Block(V(5.2, T+2.2, -1.8), V(9.2, T+3.6, 1.8)))
	shoulder, mid := V(7.2, T+3.6, 0), V(7.2, T+6.2, 0.4)
	elbow, wrist := V(1.5, T+13.5, 2.4), V(-7.5, T+12.4, 1)
	for _, s := range [][2]Vec3{{shoulder, mid}, {mid, elbow}, {elbow, wrist}} {
		r.fill(White, Capsule{s[0], s[1], 0.6})
	}
	for _, j := range []Vec3{mid, elbow, wrist} {
		r.fill(Charcoal, Sphere(j, 1.0))
	}
	r.fill(Gold, Sphere(mid.Add(elbow).Scale(0.5), 0.85), Sphere(elbow.Add(wrist).Scale(0.5), 0.85))
	r.walker(wrist.Add(V(-0.6, 3.0, 0.3)), V(-0.15, 1, 0.1), V(0.3, 0, 1), true)

	man := V(-15, M-8.5, 24)
	anchor, sag := V(-22, M-1.6, 15.6), V(-20.5, M-7.5, 20)
	r.fill(Gold, Capsule{anchor, sag, 0.45}, Capsule{sag, man.Add(V(-0.9, 0.2, -0.6)), 0.45})
	r.walker(man, V(0.35, 1, -0.2), V(0.5, 0.25, 1), false)
}
