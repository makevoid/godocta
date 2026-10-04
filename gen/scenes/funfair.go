package scenes

import (
	"math"

	"godocta/gen/kit"
	. "godocta/gen/octa"
)

// Funfair is a colorful amusement park on a lawn plate with a low curb. A tall Ferris
// wheel with a trussed rim, spokes, alternating gondolas and A-frame legs is the
// landmark (north-west); a steel roller coaster with a lift hill, first drop, vertical
// loop and camelback hill circles the north-east behind its station; a carousel with a
// striped conical canopy, poles and horses stands in the central plaza; a chair-o-plane
// swing ride, a helter-skelter spiral slide, striped big-top tents, food stalls, a
// bumper-car pavilion and a pond with swan boats fill the rest. Visitors enter through
// a striped arch lined with gold bulbs on the south (+Z) edge.
type Funfair struct{}

func init() { Register(Funfair{}) }

func (Funfair) Name() string  { return "funfair" }
func (Funfair) Title() string { return "Funfair" }

const (
	funfairWheelAngle = -0.32 // Plan rotation of the wheel's plane.
	funfairHalfX      = 48.0
	funfairHalfZ      = 36.0
)

// Key positions.
var (
	funfairWheel    = P(-28, -12) // Ferris wheel hub in plan.
	funfairCarousel = P(0, 5)
	funfairSwings   = P(-1, -19.5)
	funfairShip     = P(-24, 27.5)
	funfairFunHouse = P(40, 6)
	funfairHelter   = P(-38, 12)
	funfairBigTop   = P(-20, 14)
	funfairBumper   = P(18, 15)
	funfairPond     = P(35, 23)
)

func funfairIdx(v float64) int { return int(math.Floor(v * 2)) }

// funfairStripe reports the parity of the angular sector of p around center.
func funfairStripe(center Vec2, p Vec3, n int, phase float64) bool {
	a := p.Plan().Sub(center).Angle() + phase
	k := int(math.Floor(a / (2 * math.Pi) * float64(n)))
	return ((k%2)+2)%2 == 0
}

// funfairOBox is an oriented box: half extents along the unit axes t, u, s.
func funfairOBox(center, t, u, s Vec3, ht, hu, hs float64) Solid {
	ext := V(
		math.Abs(t.X)*ht+math.Abs(u.X)*hu+math.Abs(s.X)*hs,
		math.Abs(t.Y)*ht+math.Abs(u.Y)*hu+math.Abs(s.Y)*hs,
		math.Abs(t.Z)*ht+math.Abs(u.Z)*hu+math.Abs(s.Z)*hs,
	)
	return Func{Box{center.Sub(ext), center.Add(ext)}, func(p Vec3) bool {
		d := p.Sub(center)
		return math.Abs(d.Dot(t)) <= ht && math.Abs(d.Dot(u)) <= hu && math.Abs(d.Dot(s)) <= hs
	}}
}

func funfairNorm(v Vec3) Vec3 {
	l := v.Length()
	if l < 1e-9 {
		return V(0, 1, 0)
	}
	return v.Scale(1 / l)
}

func funfairCross(a, b Vec3) Vec3 {
	return V(a.Y*b.Z-a.Z*b.Y, a.Z*b.X-a.X*b.Z, a.X*b.Y-a.Y*b.X)
}

func (f Funfair) Build(g *Grid) Order {
	f.ground(g)
	f.pond(g)
	f.entrance(g)
	f.ferrisWheel(g)
	f.coaster(g)
	f.carousel(g)
	f.swingRide(g)
	f.helterSkelter(g)
	f.tents(g)
	f.stalls(g)
	f.bumperCars(g)
	f.pirateShip(g)
	f.funHouse(g)
	f.avenue(g)
	f.trees(g)
	f.people(g)
	return Sweep(P(0, 4), math.Pi/2)
}

// ground is a two-layer lawn plate with a stone curb, sand paths and ride pads.
func (f Funfair) ground(g *Grid) {
	hx, hz := funfairHalfX, funfairHalfZ
	g.Use(Grass).Fill(Block(V(-hx, -1.0, -hz), V(hx, 0, hz)))
	// Lawn mottle.
	box := Box{V(-hx, -0.1, -hz), V(hx, 0.1, hz)}
	g.Paint(Foliage, Func{box, func(p Vec3) bool {
		return Hash01(funfairIdx(p.X), funfairIdx(p.Z), 3) < 0.07
	}})
	// Paths: main avenue, plaza, branches to every ride.
	path := Func{box, func(p Vec3) bool {
		x, z := p.X, p.Z
		switch {
		case math.Abs(x) < 4 && z > 8 && z < hz-1: // Main avenue from the gate.
			return true
		case P(x, z).Sub(funfairCarousel).Length() < 13.5: // Carousel plaza.
			return true
		case math.Abs(z) < 2.6 && x > -44 && x < 0: // West branch.
			return true
		case math.Abs(x+11.5) < 2.2 && z > 0 && z < 30: // South-west spur.
			return true
		case math.Abs(z-7) < 2.5 && x > 0 && x < 35: // East branch.
			return true
		case math.Abs(x-funfairSwings.X) < 2.5 && z < 0 && z > -31: // North branch to the swings.
			return true
		case P(x, z).Sub(funfairSwings).Length() < 7: // Swing ride pad.
			return true
		case P(x, z).Sub(funfairHelter).Length() < 8: // Helter-skelter pad.
			return true
		case f.underWheel(P(x, z), 11.5, 8.5): // Wheel apron.
			return true
		case math.Abs(x+17) < 2 && z > -8 && z < 0: // Wheel path.
			return true
		case x > 10 && x < 30 && math.Abs(z) < 3.5: // Coaster station.
			return true
		case x > 30 && x < 45 && z > 9 && z < 14: // Fun house forecourt.
			return true
		}
		return false
	}}
	g.Paint(Sand, path)
	g.Paint(Terracotta, Func{box, func(p Vec3) bool {
		return path.Contains(p) && Hash01(funfairIdx(p.X), funfairIdx(p.Z), 5) < 0.05
	}})
	// Plaza mosaic ring.
	g.Paint(Brick, Func{box, func(p Vec3) bool {
		d := P(p.X, p.Z).Sub(funfairCarousel).Length()
		return math.Abs(d-11.5) < 0.6
	}})
	// Low stone curb with a hedge inside, open at the gate.
	g.Use(Stone).Fill(Func{Box{V(-hx, 0, -hz), V(hx, 0.9, hz)}, func(p Vec3) bool {
		return math.Abs(p.X) > hx-1 || math.Abs(p.Z) > hz-1
	}})
	g.Use(Foliage).Fill(Func{Box{V(-hx, 0, -hz), V(hx, 1.6, hz)}, func(p Vec3) bool {
		in := math.Abs(p.X) > hx-2.2 && math.Abs(p.X) <= hx-1 || math.Abs(p.Z) > hz-2.2 && math.Abs(p.Z) <= hz-1
		if !in || math.Abs(p.X) > hx-1 || math.Abs(p.Z) > hz-1 {
			return false
		}
		return !(p.Z > 0 && math.Abs(p.X) < 9)
	}})
	g.Carve(Block(V(-4, 0.2, hz-1.2), V(4, 2, hz+1)))
	g.Paint(DarkStone, Func{Box{V(-hx, -1.1, -hz), V(hx, -0.2, hz)}, func(p Vec3) bool {
		return math.Abs(p.X) > hx-1 || math.Abs(p.Z) > hz-1
	}})
}

// underWheel reports whether a plan point lies within a rectangle aligned with the
// Ferris wheel, of half extents hx in its plane and hz along its axle.
func (Funfair) underWheel(p Vec2, hx, hz float64) bool {
	d := p.Sub(funfairWheel)
	sn, cs := math.Sincos(funfairWheelAngle)
	lx, lz := d.X*cs+d.Y*sn, -d.X*sn+d.Y*cs
	return math.Abs(lx) < hx && math.Abs(lz) < hz
}

// pond is a shallow basin with a stone rim, a fountain islet and swan boats.
func (f Funfair) pond(g *Grid) {
	c := funfairPond
	ax, az := 9.0, 7.0
	inside := func(p Vec3, grow float64) bool {
		x, z := (p.X-c.X)/(ax+grow), (p.Z-c.Y)/(az+grow)
		return x*x+z*z <= 1
	}
	box := Box{V(c.X-ax-2, -1, c.Y-az-2), V(c.X+ax+2, 1, c.Y+az+2)}
	g.Carve(Func{box, func(p Vec3) bool { return p.Y > -0.2 && inside(p, 0) }})
	g.Paint(Water, Func{box, func(p Vec3) bool { return inside(p, 0) }})
	g.Use(Stone).Fill(Func{Box{V(box.Min.X, 0, box.Min.Z), V(box.Max.X, 0.6, box.Max.Z)}, func(p Vec3) bool {
		return inside(p, 1.1) && !inside(p, 0)
	}})
	// Fountain islet.
	g.Use(Stone).Fill(Cylinder{Center: c.Add(P(1.5, 0)), Radius: 1.8, Y0: -0.4, Y1: 0.8})
	g.Use(White).Fill(Cylinder{Center: c.Add(P(1.5, 0)), Radius: 0.6, Y0: 0.8, Y1: 2.6})
	g.Use(Water).Fill(Ellipsoid{c.Add(P(1.5, 0)).At(3.2), V(1.4, 0.7, 1.4)})
	// Swan boats drift around the islet.
	for i := 0; i < 5; i++ {
		a := float64(i)*2*math.Pi/5 + 0.4
		at := Polar(c.Add(P(1.5, 0)), 5.0, a)
		heading := a + math.Pi/2
		fw := V(math.Cos(heading), 0, math.Sin(heading))
		up := V(0, 1, 0)
		sd := funfairCross(fw, up)
		hull := Ellipsoid{at.At(0.3), V(1.6, 0.7, 1.6)}
		g.Use(White).Fill(Intersect(Func{hull.Bounds(), func(p Vec3) bool {
			d := p.Sub(at.At(0.3))
			x, y, z := d.Dot(fw)/1.7, d.Dot(up)/0.75, d.Dot(sd)/1.05
			return x*x+y*y+z*z <= 1
		}}, Block(V(at.X-2, 0.2, at.Y-2), V(at.X+2, 1.2, at.Y+2))))
		g.Carve(funfairOBox(at.At(1.0).Add(fw.Scale(-0.2)), fw, up, sd, 0.7, 0.3, 0.45))
		neck := at.At(0.8).Add(fw.Scale(1.3))
		head := neck.Add(V(0, 1.6, 0)).Add(fw.Scale(0.2))
		g.Use(White).Fill(Capsule{neck, head, 0.5})
		g.Use(Red).Fill(Capsule{head, head.Add(fw.Scale(0.8)), 0.4})
		if i%2 == 0 {
			g.Paint(Gold, Capsule{at.At(1.0).Sub(fw.Scale(1.4)), at.At(1.0).Sub(fw.Scale(1.0)), 0.5})
		}
	}
}

// entrance is the gate arch over the avenue: striped piers, a semicircular arch with
// gold bulbs, a navy sign board and ticket booths.
func (f Funfair) entrance(g *Grid) {
	z := funfairHalfZ - 2.5
	for _, sx := range []float64{-1, 1} {
		x := sx * 6
		pier := BlockAt(P(x, z), 2.6, 2.6, 0, 10)
		g.Use(White).Fill(pier)
		g.Paint(Red, Func{pier.Bounds(), func(p Vec3) bool { return math.Mod(p.Y, 2) >= 1 }})
		g.Use(Gold).Fill(Block(V(x-1.8, 10, z-1.8), V(x+1.8, 10.8, z+1.8)))
		g.Use(Red).Fill(Cone(P(x, z), 1.9, 10.8, 3.5))
		g.Use(Gold).Fill(Sphere(V(x, 14.8, z), 0.7))
		// Ticket booths beside the piers.
		bx := sx * 10.5
		g.Use(Navy).Fill(BlockAt(P(bx, z), 3.2, 3, 0, 3.2))
		g.Carve(BlockAt(P(bx, z+1.2), 2.2, 1.2, 1.4, 2.6))
		g.Use(Red).Fill(HipRoof{P(bx-2, z-1.9), P(bx+2, z+1.9), 3.2})
		g.Paint(White, Func{Box{V(bx-2, 3.2, z-2), V(bx+2, 6, z+2)}, func(p Vec3) bool {
			return math.Mod(p.X+20, 1.6) < 0.8
		}})
	}
	// Arch spanning the piers.
	center := V(0, 9, z)
	arch := Func{Box{V(-8, 9, z-0.8), V(8, 17, z+0.8)}, func(p Vec3) bool {
		r := math.Hypot(p.X-center.X, p.Y-center.Y)
		return p.Y >= center.Y && r >= 5.8 && r <= 7.2 && math.Abs(p.Z-z) <= 0.7
	}}
	g.Use(Red).Fill(arch)
	g.Paint(White, Func{arch.Bounds(), func(p Vec3) bool {
		return math.Hypot(p.X-center.X, p.Y-center.Y) < 6.5
	}})
	// Sign board under the arch crown.
	g.Use(Navy).Fill(Block(V(-4.4, 10.2, z-0.5), V(4.4, 12.8, z+0.5)))
	g.Use(Gold).Fill(Func{Box{V(-4.8, 9.8, z-0.6), V(4.8, 13.2, z+0.6)}, func(p Vec3) bool {
		return math.Abs(p.X) > 4.2 || p.Y < 10.4 || p.Y > 12.6
	}})
	g.Use(Navy).Fill(Block(V(-6, 9, z-0.5), V(6, 9.8, z+0.5)))
	// Gold bulbs along the arch and the piers.
	g.Use(Gold)
	for i := 0; i <= 16; i++ {
		a := math.Pi * float64(i) / 16
		g.Fill(Sphere(V(7.7*math.Cos(a), 9+7.7*math.Sin(a), z+0.4), 0.5))
	}
	for _, sx := range []float64{-1, 1} {
		for y := 1.5; y < 10; y += 2 {
			g.Fill(Sphere(V(sx*6-sx*1.6, y, z+1.6), 0.45))
			g.Fill(Sphere(V(sx*6+sx*1.6, y, z+1.6), 0.45))
		}
	}
	// Flags on the arch crown.
	g.Use(White).Fill(Capsule{V(0, 16, z), V(0, 19.5, z), 0.45})
	g.Use(Red).Fill(Block(V(0.4, 18.2, z-0.3), V(2.6, 19.5, z+0.3)))
}

// ferrisWheel is the landmark: a double trussed rim with spokes, cross ties, sixteen
// gondolas in alternating colors and A-frame legs on a boarding platform.
func (f Funfair) ferrisWheel(g *Grid) {
	const (
		R     = 18.0 // Outer rim radius.
		H     = 24.0 // Hub height.
		side  = 2.9  // Half distance between the two rims.
		count = 16
	)
	angle := funfairWheelAngle
	sn, cs := math.Sincos(angle)
	hub := funfairWheel
	world := func(v Vec3) Vec3 { // Local (x in wheel plane, y up, z along axle) to world.
		return V(hub.X+v.X*cs-v.Z*sn, v.Y, hub.Y+v.X*sn+v.Z*cs)
	}
	local := func(p Vec3) Vec3 {
		dx, dz := p.X-hub.X, p.Z-hub.Y
		return V(dx*cs+dz*sn, p.Y, -dx*sn+dz*cs)
	}
	wbox := Box{V(hub.X-R-3, 0, hub.Y-R-3), V(hub.X+R+3, H+R+2, hub.Y+R+3)}
	// Rims: outer and inner rings on both sides.
	rings := Func{wbox, func(p Vec3) bool {
		q := local(p)
		r := math.Hypot(q.X, q.Y-H)
		for _, zs := range []float64{-side, side} {
			dz := q.Z - zs
			if math.Hypot(r-R, dz) <= 0.6 || math.Hypot(r-(R-2.4), dz) <= 0.5 {
				return true
			}
		}
		return false
	}}
	g.Use(White).Fill(rings)
	// Truss zigzag between the rings, spokes and cross ties.
	at := func(r, a, z float64) Vec3 { return world(V(r*math.Cos(a), H+r*math.Sin(a), z)) }
	for i := 0; i < count*2; i++ {
		a0 := 2 * math.Pi * float64(i) / (count * 2)
		a1 := 2 * math.Pi * float64(i+1) / (count * 2)
		for _, zs := range []float64{-side, side} {
			if i%2 == 0 {
				g.Fill(Capsule{at(R, a0, zs), at(R-2.4, a1, zs), 0.45})
			} else {
				g.Fill(Capsule{at(R-2.4, a0, zs), at(R, a1, zs), 0.45})
			}
		}
	}
	for i := 0; i < count; i++ {
		a := 2*math.Pi*float64(i)/count + math.Pi/count
		for _, zs := range []float64{-side, side} {
			g.Fill(Capsule{at(1.6, a, zs*0.7), at(R-2.4, a, zs), 0.5})
		}
		g.Fill(Capsule{at(R-2.4, a, -side), at(R-2.4, a, side), 0.45})
	}
	// Gold bulbs on the outer rim.
	g.Paint(Gold, Func{wbox, func(p Vec3) bool {
		q := local(p)
		r := math.Hypot(q.X, q.Y-H)
		if r < R-0.2 || math.Hypot(r-R, math.Abs(q.Z)-side) > 0.9 {
			return false
		}
		a := math.Atan2(q.Y-H, q.X)
		k := a / (2 * math.Pi) * 64
		return math.Abs(k-math.Round(k)) < 0.18
	}})
	// Hub and axle.
	g.Use(Gold).Fill(Capsule{at(0, 0, -side-1.4), at(0, 0, side+1.4), 1.5})
	g.Use(Red).Fill(Capsule{at(0, 0, -0.8), at(0, 0, 0.8), 2.0})
	// Gondolas hang from pivots on the outer rim between the sides.
	colors := []Color{Red, Gold, Navy, Water}
	for i := 0; i < count; i++ {
		a := 2 * math.Pi * float64(i) / count
		pivot := V(R*math.Cos(a), H+R*math.Sin(a), 0)
		g.Use(Charcoal).Fill(Capsule{world(pivot.Add(V(0, 0, -side))), world(pivot.Add(V(0, 0, side))), 0.45})
		g.Fill(Capsule{world(pivot), world(pivot.Add(V(0, -1.4, 0))), 0.45})
		top := pivot.Y - 1.2
		body := Func{Box{world(pivot).Sub(V(4, 6, 4)), world(pivot).Add(V(4, 1, 4))}, func(p Vec3) bool {
			q := local(p)
			dx, dz := q.X-pivot.X, q.Z
			return math.Abs(dx) <= 1.8 && math.Abs(dz) <= 2.0 && q.Y <= top && q.Y >= top-3.6
		}}
		g.Use(colors[i%len(colors)]).Fill(body)
		// Window band through the cabin and a white roof cap.
		g.Carve(Func{body.Box, func(p Vec3) bool {
			q := local(p)
			dx, dz := q.X-pivot.X, q.Z
			return q.Y < top-1.0 && q.Y > top-2.6 && (math.Abs(dx) <= 1.0 || math.Abs(dz) <= 1.2)
		}})
		g.Paint(White, Func{body.Box, func(p Vec3) bool { return local(p).Y > top-0.6 }})
	}
	// A-frame legs on both sides, braced, standing on concrete pads.
	for _, zs := range []float64{-side - 1.6, side + 1.6} {
		foot := func(x float64) Vec3 { return world(V(x, 0.3, zs*1.6)) }
		apex := world(V(0, H, zs))
		g.Use(Navy).Fill(Capsule{foot(-9.5), apex, 0.8}, Capsule{foot(9.5), apex, 0.8})
		for _, y := range []float64{6.5, 14} {
			t := y / H
			lx := 9.5 * (1 - t)
			lz := zs*1.6 + (zs-zs*1.6)*t
			g.Fill(Capsule{world(V(-lx, y, lz)), world(V(lx, y, lz)), 0.55})
		}
		g.Use(Stone).Fill(Cylinder{Center: foot(-9.5).Plan(), Radius: 1.6, Y0: 0, Y1: 0.8})
		g.Fill(Cylinder{Center: foot(9.5).Plan(), Radius: 1.6, Y0: 0, Y1: 0.8})
	}
	// Boarding platform with steps and a ticket kiosk.
	plat := Func{Box{world(V(0, 0, 0)).Sub(V(6, 0, 6)), world(V(0, 0, 0)).Add(V(6, 1.6, 6))}, func(p Vec3) bool {
		q := local(p)
		return math.Abs(q.X) <= 3.5 && math.Abs(q.Z) <= 2.4 && q.Y <= 1.2 && q.Y >= 0
	}}
	g.Use(Wood).Fill(plat)
	g.Use(Red).Fill(Func{Box{world(V(0, 0, 0)).Sub(V(8, 0, 8)), world(V(0, 0, 0)).Add(V(8, 4, 8))}, func(p Vec3) bool {
		q := local(p)
		return q.X >= 4.5 && q.X <= 7 && math.Abs(q.Z-4) <= 1.2 && q.Y <= 3
	}})
}

// funfairTrackPoint is one sample along the coaster: position, sideways and up axes.
type funfairTrackPoint struct {
	pos, side, up Vec3
	loop          bool
}

// coasterPath builds the closed circuit: station, lift hill, first drop, vertical loop,
// turnaround, camelback hill and the return to the station.
func (f Funfair) coasterPath() []funfairTrackPoint {
	before := []Vec3{
		V(12, 2, 2), V(18, 2, 2), V(24, 2, 2), V(29, 2.2, 1.5), V(34, 2.8, -1),
		V(37.5, 3.6, -5), V(39, 6, -10), V(39.5, 12, -16), V(39.5, 18, -22), V(39, 23, -27.5),
		V(36.5, 24.5, -31), V(32.5, 24, -32.5), V(29.5, 16, -32.5), V(27.5, 7, -32.5), V(25.5, 2.8, -32.5),
		V(23, 2.4, -32.5),
	}
	const (
		loopR   = 6.0
		lateral = 4.0
	)
	entry := V(23, 2.4, -32.5)
	dir := V(-1, 0, 0)
	lat := V(0, 0, 1)
	after := []Vec3{
		entry.Add(lat.Scale(lateral)), V(19, 2.5, -28.5), V(15.5, 3.2, -27.5), V(13, 4, -25),
		V(12.5, 5, -21.5), V(14.5, 6, -18.5), V(18, 7.5, -16.5), V(22, 10.5, -16.5), V(26, 12.5, -16.5),
		V(30, 9, -16.5), V(33, 4, -15.5), V(34.8, 3, -12), V(33, 3, -7.5), V(28, 3.6, -6.5),
		V(22, 4.5, -6.5), V(16, 3.6, -6.5), V(11, 2.8, -5.5), V(8.5, 2.2, -2), V(10, 2, 1.4),
		V(12, 2, 2),
	}
	spline := func(pts []Vec3, step float64) []Vec3 {
		var out []Vec3
		n := len(pts)
		for i := 0; i+1 < n; i++ {
			p0, p1, p2, p3 := pts[max(i-1, 0)], pts[i], pts[i+1], pts[min(i+2, n-1)]
			steps := int(math.Ceil(p2.Sub(p1).Length() / step))
			for k := 0; k < steps; k++ {
				t := float64(k) / float64(steps)
				t2, t3 := t*t, t*t*t
				v := p1.Scale(2).
					Add(p2.Sub(p0).Scale(t)).
					Add(p0.Scale(2).Sub(p1.Scale(5)).Add(p2.Scale(4)).Sub(p3).Scale(t2)).
					Add(p1.Scale(3).Sub(p0).Sub(p2.Scale(3)).Add(p3).Scale(t3)).
					Scale(0.5)
				out = append(out, v)
			}
		}
		return out
	}
	var track []funfairTrackPoint
	flat := func(pts []Vec3) {
		for i, p := range pts {
			a, b := pts[max(i-1, 0)], pts[min(i+1, len(pts)-1)]
			t := funfairNorm(b.Sub(a))
			side := funfairNorm(V(-t.Z, 0, t.X))
			up := funfairNorm(funfairCross(t, side))
			if up.Y < 0 {
				up = up.Scale(-1)
			}
			track = append(track, funfairTrackPoint{pos: p, side: side, up: up})
		}
	}
	flat(spline(before, 0.5))
	// Vertical loop: forward, up and over, drifting sideways to clear the entry.
	center := entry.Add(V(0, loopR, 0))
	steps := int(math.Ceil(2 * math.Pi * loopR / 0.45))
	for k := 0; k <= steps; k++ {
		th := 2 * math.Pi * float64(k) / float64(steps)
		s := th / (2 * math.Pi)
		s = s * s * (3 - 2*s)
		p := entry.Add(dir.Scale(loopR * math.Sin(th))).Add(V(0, loopR*(1-math.Cos(th)), 0)).Add(lat.Scale(lateral * s))
		c := center.Add(lat.Scale(lateral * s))
		track = append(track, funfairTrackPoint{pos: p, side: lat, up: funfairNorm(c.Sub(p)), loop: true})
	}
	flat(spline(after, 0.5))
	return track
}

// coaster draws the track (two red rails, a gold spine and ties), its supports, the
// station and a train on the lift hill.
func (f Funfair) coaster(g *Grid) {
	track := f.coasterPath()
	n := len(track)
	rail := func(tp funfairTrackPoint, s float64) Vec3 { return tp.pos.Add(tp.side.Scale(s)) }
	spine := func(tp funfairTrackPoint) Vec3 { return tp.pos.Sub(tp.up.Scale(1.0)) }
	dist := 0.0
	nextTie := 0.0
	nextSupport := 2.0
	for i := 0; i+1 < n; i++ {
		a, b := track[i], track[i+1]
		g.Use(Red).Fill(Capsule{rail(a, -1.05), rail(b, -1.05), 0.48}, Capsule{rail(a, 1.05), rail(b, 1.05), 0.48})
		spineColor := Gold
		if !a.loop && b.pos.Y > a.pos.Y+0.12 && a.pos.Y > 3 && a.pos.X > 33 {
			spineColor = Charcoal // Lift chain.
		}
		g.Use(spineColor).Fill(Capsule{spine(a), spine(b), 0.55})
		dist += b.pos.Sub(a.pos).Length()
		if dist >= nextTie {
			nextTie = dist + 1.6
			g.Use(Navy).Fill(Capsule{rail(a, -1.05), spine(a), 0.42}, Capsule{rail(a, 1.05), spine(a), 0.42})
		}
		if dist >= nextSupport && !a.loop {
			nextSupport = dist + 5.5 + math.Min(a.pos.Y*0.12, 2.5)
			f.coasterSupport(g, track, i)
		}
	}
	// Station: deck, striped canopy on posts.
	g.Use(Wood).Fill(Block(V(12, 0, -1.5), V(26, 0.9, 5)))
	g.Use(Stone).Fill(Block(V(12, 0, 3.2), V(26, 1.2, 5)))
	for x := 12.5; x <= 26; x += 4.5 {
		for _, z := range []float64{-1.2, 4.6} {
			g.Use(White).Fill(Cylinder{Center: P(x, z), Radius: 0.5, Y0: 0, Y1: 5.5})
		}
	}
	roof := Gable{P(11.5, -2.2), P(26.5, 5.6), 5.5, AxisX}
	g.Use(Red).Fill(Subtract(roof, Translate(roof, V(0, -1.0, 0))))
	g.Paint(White, Func{roof.Bounds(), func(p Vec3) bool { return math.Mod(p.X+40, 2) < 1 }})
	g.Use(Navy).Fill(Block(V(15, 5.5, 5.4), V(23, 7.5, 5.9)))
	// Train cresting the lift hill.
	colors := []Color{Red, Gold, Navy, Gold, Navy}
	car := 0
	for i := 0; i < n && car < len(colors); i++ {
		tp := track[i]
		if tp.loop || tp.pos.X < 38 || tp.pos.Z > -12 {
			continue
		}
		if car == 0 && tp.pos.Y < 9 {
			continue
		}
		t := funfairNorm(track[min(i+1, n-1)].pos.Sub(track[max(i-1, 0)].pos))
		center := tp.pos.Add(tp.up.Scale(1.1))
		g.Use(colors[car]).Fill(funfairOBox(center, t, tp.up, tp.side, 1.0, 0.7, 1.3))
		g.Carve(funfairOBox(center.Add(tp.up.Scale(0.6)), t, tp.up, tp.side, 0.5, 0.5, 0.7))
		car++
		// Skip ahead about 2.4 units for the next car.
		d := 0.0
		for i+1 < n && d < 2.4 {
			d += track[i+1].pos.Sub(track[i].pos).Length()
			i++
		}
		i--
	}
}

// coasterSupport drops a column (or an A-frame for tall track) from the spine to the
// ground, unless another stretch of track lies directly below.
func (f Funfair) coasterSupport(g *Grid, track []funfairTrackPoint, i int) {
	tp := track[i]
	top := tp.pos.Sub(tp.up.Scale(1.4))
	if top.Y < 1.2 {
		return
	}
	for j, o := range track {
		if j > i-12 && j < i+12 {
			continue
		}
		if o.pos.Y < tp.pos.Y-1 && P(o.pos.X, o.pos.Z).Sub(top.Plan()).Length() < 2.6 {
			return
		}
	}
	g.Use(White)
	if top.Y < 7 {
		g.Fill(Capsule{V(top.X, 0, top.Z), top, 0.55})
	} else {
		spread := math.Min(top.Y*0.18, 3.2)
		l := V(top.X, 0, top.Z).Add(tp.side.Scale(-spread))
		r := V(top.X, 0, top.Z).Add(tp.side.Scale(spread))
		g.Fill(Capsule{l, top, 0.55}, Capsule{r, top, 0.55})
		for y := 5.0; y < top.Y-1; y += 6 {
			t := y / top.Y
			g.Fill(Capsule{l.Add(top.Sub(l).Scale(t)), r.Add(top.Sub(r).Scale(t)), 0.45})
		}
	}
	g.Use(Stone).Fill(Cylinder{Center: top.Plan(), Radius: 1.0, Y0: 0, Y1: 0.6})
}

// carousel is a round platform with a mirrored center column, gold poles carrying
// horses and a striped conical canopy with a scalloped valance and a flag.
func (f Funfair) carousel(g *Grid) {
	c := funfairCarousel
	g.Use(Stone).Fill(Cylinder{Center: c, Radius: 9.2, Y0: 0, Y1: 0.6})
	g.Use(Wood).Fill(Cylinder{Center: c, Radius: 8.4, Y0: 0.6, Y1: 1.3})
	g.Paint(Gold, Cylinder{Center: c, Radius: 8.4, Inner: 7.7, Y0: 0.6, Y1: 1.3})
	// Center column with panels.
	g.Use(White).Fill(Cylinder{Center: c, Radius: 2.4, Y0: 1.3, Y1: 7.6})
	g.Paint(Glass, Func{Box{V(c.X-3, 2.3, c.Y-3), V(c.X+3, 6.4, c.Y+3)}, func(p Vec3) bool {
		return funfairStripe(c, p, 8, 0)
	}})
	g.Use(Gold).Fill(Cylinder{Center: c, Radius: 2.9, Y0: 1.3, Y1: 2.0}, Cylinder{Center: c, Radius: 2.9, Y0: 6.8, Y1: 7.6})
	// Canopy: striped cone shell with a scalloped valance.
	outer := Frustum{c, 9.4, 0, 7.6, 13.6}
	inner := Frustum{c, 9.4, 0, 6.6, 12.6}
	g.Use(Red).FillVolume(Subtract(outer, inner))
	g.Paint(White, Func{outer.Bounds(), func(p Vec3) bool { return funfairStripe(c, p, 20, 0) }})
	val := Func{Box{V(c.X-9.6, 5.8, c.Y-9.6), V(c.X+9.6, 7.8, c.Y+9.6)}, func(p Vec3) bool {
		d := p.Plan().Sub(c)
		r := d.Length()
		if r > 9.5 || r < 8.6 {
			return false
		}
		a := d.Angle() / (2 * math.Pi) * 40
		scallop := 6.4 - 0.6*math.Abs(math.Sin(a*math.Pi))
		return p.Y >= scallop
	}}
	g.Use(Gold).Fill(val)
	g.Use(Gold).Fill(Cylinder{Center: c, Radius: 0.6, Y0: 13, Y1: 15.2})
	g.Fill(Sphere(c.At(15.4), 0.7))
	g.Use(White).Fill(Capsule{c.At(15.6), c.At(17.6), 0.4})
	g.Use(Red).Fill(Block(V(c.X+0.4, 16.4, c.Y-0.3), V(c.X+2.6, 17.6, c.Y+0.3)))
	// Poles and horses on two rings.
	horseColors := []Color{White, White, Navy, White, Gold, White, Red}
	for ring, rr := range []float64{6.6, 4.4} {
		count := 16 - ring*4
		for i := 0; i < count; i++ {
			a := 2*math.Pi*float64(i)/float64(count) + float64(ring)*0.2
			at := Polar(c, rr, a)
			g.Use(Gold).Fill(Capsule{at.At(1.3), at.At(7.6), 0.45})
			lift := 0.7 * math.Sin(float64(i)*1.7+float64(ring))
			y := 3.6 + lift
			tang := V(-math.Sin(a), 0, math.Cos(a))
			rad := V(math.Cos(a), 0, math.Sin(a))
			body := at.At(y)
			hc := horseColors[(i+ring*3)%len(horseColors)]
			g.Use(hc).Fill(funfairOBox(body, tang, V(0, 1, 0), rad, 1.3, 0.6, 0.45))
			head := body.Add(tang.Scale(1.4)).Add(V(0, 1.0, 0))
			g.Fill(Capsule{body.Add(tang.Scale(0.9)).Add(V(0, 0.3, 0)), head, 0.5})
			g.Fill(Capsule{head, head.Add(tang.Scale(0.8)).Add(V(0, -0.2, 0)), 0.42})
			for _, lx := range []float64{-0.9, 0.9} {
				g.Fill(Capsule{body.Add(tang.Scale(lx)), body.Add(tang.Scale(lx * 1.3)).Add(V(0, -1.4, 0)), 0.4})
			}
			saddle := Red
			if hc == Red || hc == Navy {
				saddle = Gold
			}
			g.Paint(saddle, funfairOBox(body.Add(V(0, 0.4, 0)), tang, V(0, 1, 0), rad, 0.5, 0.3, 0.6))
		}
	}
}

// swingRide is a chair-o-plane: a tall pole, a striped umbrella canopy and chairs
// flying out on long chains.
func (f Funfair) swingRide(g *Grid) {
	c := funfairSwings
	g.Use(Stone).Fill(Cylinder{Center: c, Radius: 4, Y0: 0, Y1: 0.8})
	g.Use(Navy).Fill(Cylinder{Center: c, Radius: 2.6, Y0: 0.8, Y1: 2.2})
	g.Use(White).Fill(Cylinder{Center: c, Radius: 1.2, Y0: 2.2, Y1: 22})
	g.Paint(Red, Func{Box{V(c.X-2, 2, c.Y-2), V(c.X+2, 22, c.Y+2)}, func(p Vec3) bool {
		return math.Mod(p.Y, 4) < 1.2
	}})
	// Canopy: an umbrella of 16 striped panels.
	outer := Frustum{c, 7, 1.4, 17.5, 22}
	inner := Frustum{c, 7, 1.4, 16.5, 21}
	g.Use(Navy).FillVolume(Subtract(outer, inner))
	g.Paint(Gold, Func{outer.Bounds(), func(p Vec3) bool { return funfairStripe(c, p, 16, 0) }})
	g.Use(Red).Fill(Cylinder{Center: c, Radius: 7.3, Inner: 6.4, Y0: 16.4, Y1: 17.6})
	g.Use(Gold).Fill(Cylinder{Center: c, Radius: 1.6, Y0: 22, Y1: 23.2})
	g.Fill(Sphere(c.At(23.8), 0.8))
	// Chains and chairs swung out.
	seats := []Color{Red, Gold, White, Water}
	n := 16
	for i := 0; i < n; i++ {
		a := 2*math.Pi*float64(i)/float64(n) + math.Pi/float64(n)
		top := Polar(c, 6.2, a).At(16.6)
		seat := Polar(c, 10.2, a).At(7.2)
		g.Use(Charcoal).Fill(Capsule{top, seat.Add(V(0, 1.2, 0)), 0.42})
		tang := V(-math.Sin(a), 0, math.Cos(a))
		rad := V(math.Cos(a), 0, math.Sin(a))
		g.Use(seats[i%len(seats)]).Fill(funfairOBox(seat, tang, V(0, 1, 0), rad, 0.8, 0.6, 0.7))
		g.Fill(funfairOBox(seat.Add(rad.Scale(0.6)).Add(V(0, 0.8, 0)), tang, V(0, 1, 0), rad, 0.8, 0.6, 0.35))
	}
	// Low fence ring around the ride.
	g.Use(White).Fill(Func{Box{V(c.X-12, 0, c.Y-12), V(c.X+12, 1.4, c.Y+12)}, func(p Vec3) bool {
		d := p.Plan().Sub(c).Length()
		if math.Abs(d-11.8) > 0.45 {
			return false
		}
		a := p.Plan().Sub(c).Angle()
		if math.Abs(math.Remainder(a-math.Pi/2, 2*math.Pi)) < 0.2 {
			return false // Gate toward the plaza.
		}
		k := a / (2 * math.Pi) * 48
		return p.Y > 0.9 || math.Abs(k-math.Round(k)) < 0.12
	}})
}

// helterSkelter is a striped tower with a gallery and cone roof, ringed by a spiral
// slide with a raised lip and a run-out at the bottom.
func (f Funfair) helterSkelter(g *Grid) {
	c := funfairHelter
	const (
		top   = 17.0
		pitch = 5.2
		rIn   = 3.4
		rOut  = 6.6
	)
	g.Use(Stone).Fill(Cylinder{Center: c, Radius: 4, Y0: 0, Y1: 0.6})
	tower := Cylinder{Center: c, Radius: rIn, Inner: rIn - 1.1, Y0: 0.6, Y1: top + 2}
	g.Use(White).Fill(tower)
	g.Paint(Red, Func{tower.Bounds(), func(p Vec3) bool { return funfairStripe(c, p, 12, 0) }})
	g.Carve(Arch{c.Add(P(0, rIn-0.5)).At(0.6), 1.0, 1.6, 2, AxisZ})
	// Gallery and roof.
	g.Use(Navy).Fill(Cylinder{Center: c, Radius: rIn + 1.2, Y0: top + 2, Y1: top + 2.8})
	g.Use(Gold).Fill(Cylinder{Center: c, Radius: rIn + 1.2, Inner: rIn + 0.4, Y0: top + 2.8, Y1: top + 3.6})
	cone := Subtract(Frustum{c, rIn + 0.9, 0, top + 3.6, top + 10}, Frustum{c, rIn + 0.9, 0, top + 2.6, top + 9})
	g.Use(Red).FillVolume(cone)
	g.Paint(White, Func{cone.Bounds(), func(p Vec3) bool { return funfairStripe(c, p, 12, 0) }})
	g.Use(White).Fill(Capsule{c.At(top + 9.5), c.At(top + 12.5), 0.4})
	g.Use(Gold).Fill(Block(V(c.X+0.4, top+11.2, c.Y-0.3), V(c.X+2.4, top+12.5, c.Y+0.3)))
	// Spiral slide.
	box := Box{V(c.X-rOut-0.5, 0.5, c.Y-rOut-0.5), V(c.X+rOut+0.5, top+1, c.Y+rOut+0.5)}
	surface := func(p Vec3) (float64, bool) {
		d := p.Plan().Sub(c)
		r := d.Length()
		if r < rIn-0.2 || r > rOut {
			return 0, false
		}
		a := math.Mod(d.Angle()+4*math.Pi, 2*math.Pi) / (2 * math.Pi)
		best, ok := 0.0, false
		for k := 0; k < 4; k++ {
			ys := top - pitch*(float64(k)+a)
			if ys < 1.0 {
				break
			}
			if p.Y <= ys+1.4 && p.Y >= ys-0.75 {
				best, ok = ys, true
			}
		}
		return best, ok
	}
	g.Use(Gold).Fill(Func{box, func(p Vec3) bool {
		ys, ok := surface(p)
		return ok && p.Y <= ys
	}})
	g.Use(Navy).Fill(Func{box, func(p Vec3) bool {
		ys, ok := surface(p)
		r := p.Plan().Sub(c).Length()
		return ok && p.Y > ys && r > rOut-0.9
	}})
	// Run-out and mats.
	g.Use(Gold).Fill(Block(V(c.X+rIn, 0.4, c.Y-1.2), V(c.X+rOut+4, 1.0, c.Y+1.2)))
	g.Use(Wood).Fill(Block(V(c.X+rOut+1, 0.4, c.Y-1.8), V(c.X+rOut+4.5, 0.8, c.Y+1.8)))
}

// tents are striped big tops: a red/white circus tent and smaller navy/gold ones.
func (f Funfair) tents(g *Grid) {
	type tent struct {
		at        Vec2
		r, wall   float64
		roof      float64
		a, b      Color
		stripes   int
		doorAngle float64
	}
	list := []tent{
		{funfairBigTop, 7.5, 5, 7, Red, White, 16, 0},
		{P(17.5, 28.5), 4, 3.2, 4.5, Navy, Gold, 12, math.Pi},
		{P(-16, -31), 4.5, 3.4, 5, Navy, Gold, 12, math.Pi / 2},
	}
	for _, t := range list {
		c := t.at
		wall := Cylinder{Center: c, Radius: t.r, Inner: t.r - 1, Y0: 0, Y1: t.wall}
		g.Use(t.a).Fill(wall)
		g.Paint(t.b, Func{wall.Bounds(), func(p Vec3) bool { return funfairStripe(c, p, t.stripes, 0) }})
		roof := Subtract(Frustum{c, t.r + 1, 0, t.wall, t.wall + t.roof}, Frustum{c, t.r + 1, 0, t.wall - 1, t.wall + t.roof - 1})
		g.Use(t.a).FillVolume(roof)
		g.Paint(t.b, Func{roof.Bounds(), func(p Vec3) bool { return funfairStripe(c, p, t.stripes, 0) }})
		// Scalloped edge.
		g.Use(Gold).Fill(Func{Box{V(c.X-t.r-1.2, t.wall-1, c.Y-t.r-1.2), V(c.X+t.r+1.2, t.wall+0.6, c.Y+t.r+1.2)}, func(p Vec3) bool {
			d := p.Plan().Sub(c)
			r := d.Length()
			return r <= t.r+1.1 && r >= t.r+0.2 && p.Y < t.wall+0.4 && p.Y > t.wall-0.6-0.4*math.Abs(math.Sin(d.Angle()*float64(t.stripes)))
		}})
		// Door.
		door := Polar(c, t.r-0.5, t.doorAngle)
		g.Carve(funfairOBox(door.At(1.6), V(-math.Sin(t.doorAngle), 0, math.Cos(t.doorAngle)), V(0, 1, 0), V(math.Cos(t.doorAngle), 0, math.Sin(t.doorAngle)), 1.2, 1.6, 1.2))
		apex := t.wall + t.roof
		g.Use(White).Fill(Capsule{c.At(apex - 1), c.At(apex + 2.5), 0.4})
		g.Use(t.b).Fill(Block(V(c.X+0.4, apex+1.4, c.Y-0.3), V(c.X+2.2, apex+2.5, c.Y+0.3)))
	}
}

// stalls line the avenue: open-fronted booths with striped gable awnings.
func (f Funfair) stalls(g *Grid) {
	type stall struct {
		at     Vec2
		facing float64 // +1 faces +X, -1 faces -X.
		a, b   Color
	}
	var list []stall
	k := 0
	for _, z := range []float64{17, 22.5, 28} {
		for _, sx := range []float64{-1, 1} {
			a, b := Red, White
			if k%2 == 1 {
				a, b = Navy, Gold
			}
			list = append(list, stall{P(sx*8, z), -sx, a, b})
			k++
		}
	}
	list = append(list, stall{P(30, 11), -1, Red, White}, stall{P(-8, -9), 1, Navy, Gold})
	for i, s := range list {
		x0, x1 := s.at.X-1.8, s.at.X+1.8
		z0, z1 := s.at.Y-2.2, s.at.Y+2.2
		body := []Color{White, Wood, Gold, Water}[i%4]
		g.Use(body).Fill(Block(V(x0, 0, z0), V(x1, 3, z1)))
		// Open serving front with a counter.
		front := x1
		if s.facing < 0 {
			front = x0
		}
		g.Carve(Block(V(front-1.2, 1.4, z0+0.8), V(front+1.2, 3, z1-0.8)))
		g.Carve(Block(V(x0+0.8, 0.5, z0+0.8), V(x1-0.8, 3, z1-0.8)))
		g.Use(s.a).Fill(Block(V(front-0.4, 0.4, z0+0.8), V(front+0.4, 1.3, z1-0.8)))
		// Awning roof striped across the ridge.
		roof := Gable{P(x0-0.8, z0-0.4), P(x1+0.8, z1+0.4), 3, AxisZ}
		g.Use(s.a).Fill(roof)
		g.Paint(s.b, Func{roof.Bounds(), func(p Vec3) bool { return math.Mod(p.Z+60, 1.6) < 0.8 }})
		// Sign board on top.
		g.Use(s.b).Fill(Block(V(s.at.X-0.3, 4.8, z0+0.4), V(s.at.X+0.3, 6, z1-0.4)))
	}
}

// bumperCars is an open pavilion: charcoal floor, gold posts, navy roof with gold
// cresting, a barrier rail and colorful cars with antenna poles.
func (f Funfair) bumperCars(g *Grid) {
	c := funfairBumper
	hw, hd := 8.0, 6.0
	x0, x1, z0, z1 := c.X-hw, c.X+hw, c.Y-hd, c.Y+hd
	g.Use(Charcoal).Fill(Block(V(x0, 0, z0), V(x1, 0.6, z1)))
	g.Use(Stone).Fill(Func{Box{V(x0, 0, z0), V(x1, 0.7, z1)}, func(p Vec3) bool {
		return p.X < x0+0.8 || p.X > x1-0.8 || p.Z < z0+0.8 || p.Z > z1-0.8
	}})
	// Barrier rail with an entry gap.
	g.Use(Red).Fill(Func{Box{V(x0+0.8, 0.6, z0+0.8), V(x1-0.8, 1.6, z1-0.8)}, func(p Vec3) bool {
		edge := p.X < x0+1.6 || p.X > x1-1.6 || p.Z < z0+1.6 || p.Z > z1-1.6
		gap := p.Z < z0+2 && math.Abs(p.X-c.X) < 1.5
		return edge && !gap
	}})
	for x := x0 + 0.8; x <= x1; x += (2*hw - 1.6) / 4 {
		for _, z := range []float64{z0 + 0.8, z1 - 0.8} {
			g.Use(Gold).Fill(Cylinder{Center: P(x, z), Radius: 0.55, Y0: 0.6, Y1: 6})
		}
	}
	for _, x := range []float64{x0 + 0.8, x1 - 0.8} {
		g.Use(Gold).Fill(Cylinder{Center: P(x, c.Y), Radius: 0.55, Y0: 0.6, Y1: 6})
	}
	// Roof: a low navy hip with a gold fascia and a sign board.
	g.Use(Gold).Fill(Block(V(x0-0.4, 6, z0-0.4), V(x1+0.4, 6.8, z1+0.4)))
	g.Carve(Block(V(x0+0.6, 6, z0+0.6), V(x1-0.6, 6.4, z1-0.6)))
	roof := HipRoof{P(x0-0.4, z0-0.4), P(x1+0.4, z1+0.4), 6.8}
	g.Use(Navy).Fill(Intersect(Subtract(roof, Translate(roof, V(0, -1, 0))), Block(V(x0-1, 6.8, z0-1), V(x1+1, 10.5, z1+1))))
	g.Use(Navy).Fill(Block(V(x0+5.4, 10.5, z0+5.4), V(x1-5.4, 11, z1-5.4)))
	g.Use(Red).Fill(Block(V(c.X-4, 6.8, z1+0.1), V(c.X+4, 8.8, z1+0.7)))
	g.Paint(Gold, Block(V(c.X-3.2, 7.4, z1+0.1), V(c.X+3.2, 8.2, z1+0.7)))
	// Cars.
	colors := []Color{Red, Gold, Water, White, Grass, Red, Navy, Gold}
	for i := 0; i < 8; i++ {
		px := x0 + 3 + Hash01(i, 21)*(2*hw-6)
		pz := z0 + 3 + Hash01(i, 22)*(2*hd-6)
		a := Hash01(i, 23) * 2 * math.Pi
		t := V(math.Cos(a), 0, math.Sin(a))
		s := V(-t.Z, 0, t.X)
		at := V(px, 1.2, pz)
		g.Use(Charcoal).Fill(funfairOBox(at, t, V(0, 1, 0), s, 1.3, 0.5, 1.0))
		g.Use(colors[i]).Fill(funfairOBox(at.Add(V(0, 0.7, 0)), t, V(0, 1, 0), s, 1.0, 0.45, 0.8))
		g.Carve(funfairOBox(at.Add(V(0, 1.1, 0)).Sub(t.Scale(0.1)), t, V(0, 1, 0), s, 0.5, 0.4, 0.5))
		g.Use(Charcoal).Fill(Capsule{at.Add(V(0, 1, 0)).Sub(t.Scale(0.8)), at.Add(V(0, 4.8, 0)).Sub(t.Scale(0.8)), 0.35})
	}
}

// trees fill the lawns on a jittered grid, avoiding rides and paths.
func (f Funfair) trees(g *Grid) {
	clear := func(x, z float64) bool {
		if math.Abs(x) > funfairHalfX-4 || math.Abs(z) > funfairHalfZ-4 {
			return false
		}
		p := P(x, z)
		zones := []struct {
			at Vec2
			r  float64
		}{
			{funfairCarousel, 15}, {funfairSwings, 14}, {funfairHelter, 10}, {funfairBigTop, 11},
			{funfairPond, 11.5}, {P(17.5, 28.5), 6.5}, {P(-16, -31), 7}, {funfairShip, 11}, {funfairFunHouse, 8},
		}
		for _, zn := range zones {
			if p.Sub(zn.at).Length() < zn.r {
				return false
			}
		}
		if f.underWheel(p, 19, 9.5) {
			return false
		}
		switch {
		case math.Abs(x) < 13 && z > 12: // Avenue and stalls.
			return false
		case x > 6 && z < 9 && z > -36: // Coaster area.
			return false
		case math.Abs(z) < 4.5 && x < 0: // West branch.
			return false
		case math.Abs(x+11.5) < 4 && z > 0: // Spur.
			return false
		case math.Abs(z-7) < 4.5 && x > 0 && x < 34: // East branch.
			return false
		case math.Abs(x-funfairSwings.X) < 4.5 && z < 0: // North branch.
			return false
		case x > 8 && x < 28 && z > 7 && z < 23: // Bumper cars.
			return false
		}
		return true
	}
	for gx := -46.0; gx <= 46; gx += 5 {
		for gz := -34.0; gz <= 34; gz += 5 {
			i, j := int(gx), int(gz)
			x := gx + (Hash01(i, j, 31)-0.5)*3
			z := gz + (Hash01(i, j, 32)-0.5)*3
			if !clear(x, z) || Hash01(i, j, 33) < 0.25 {
				continue
			}
			g.Build(kit.Tree{At: P(x, z), Base: 0.3, Height: 5.5 + 3*Hash01(i, j, 34)})
		}
	}
	// Coaster infield trees.
	for _, p := range []Vec2{P(24, -24), P(30, -24), P(18, -11), P(28, -11), P(44, -12), P(44, -24)} {
		g.Build(kit.Tree{At: p, Base: 0.3, Height: 5})
	}
}

// pirateShip is a swinging boat ride: an open wooden hull with benches and gold
// figureheads hangs from a gold axle on red A-frames, caught mid-swing.
func (f Funfair) pirateShip(g *Grid) {
	c := funfairShip
	const (
		axle = 12.5
		deck = 4.2
		L    = 7.5 // Half length.
		B    = 2.6 // Half beam.
		dep  = 2.8
	)
	sn, cs := math.Sincos(0.38)
	toRest := func(p Vec3) Vec3 {
		x, y := p.X-c.X, p.Y-axle
		return V(x*cs+y*sn, -x*sn+y*cs+axle, p.Z-c.Y)
	}
	fromRest := func(v Vec3) Vec3 {
		y := v.Y - axle
		return V(c.X+v.X*cs-y*sn, v.X*sn+y*cs+axle, c.Y+v.Z)
	}
	inside := func(q Vec3, inset float64) bool {
		if math.Abs(q.X) > L-inset {
			return false
		}
		t := q.X / L
		top := deck + 1.8*t*t*t*t
		keel := deck - dep + 2.0*t*t + inset
		if q.Y < keel || q.Y > top {
			return false
		}
		hb := (B - inset) * math.Sqrt(math.Max(0, math.Min(1, (q.Y-keel+0.3)/(deck-keel+0.3)))) * math.Sqrt(1-0.55*t*t)
		return math.Abs(q.Z) <= hb
	}
	box := Box{V(c.X-11, 0.5, c.Y-B-1), V(c.X+11, 16, c.Y+B+1)}
	hull := Func{box, func(p Vec3) bool {
		q := toRest(p)
		return inside(q, 0) && !inside(q, 1)
	}}
	g.Use(Wood).FillVolume(hull)
	g.Paint(Red, Func{box, func(p Vec3) bool {
		q := toRest(p)
		y := q.Y - 1.8*math.Pow(q.X/L, 4)
		return y > deck-1.5 && y < deck-0.8
	}})
	g.Paint(Gold, Func{box, func(p Vec3) bool {
		q := toRest(p)
		return q.Y > deck-0.4+1.8*math.Pow(q.X/L, 4)
	}})
	// Benches.
	g.Use(Navy)
	for xs := -4.5; xs <= 4.5; xs += 1.8 {
		xs := xs
		g.Fill(Func{box, func(p Vec3) bool {
			q := toRest(p)
			return math.Abs(q.X-xs) < 0.45 && q.Y < deck-0.5 && inside(q, 0)
		}})
	}
	// Figureheads.
	g.Use(Gold)
	for _, sx := range []float64{-1, 1} {
		a := fromRest(V(sx*(L-0.4), deck+1.4, 0))
		b := fromRest(V(sx*(L+0.6), deck+3.6, 0))
		g.Fill(Capsule{a, b, 0.65})
		g.Fill(Sphere(b, 0.9))
	}
	// Swing arms and axle.
	for _, sz := range []float64{-1, 1} {
		z := sz * (B + 0.5)
		hub := fromRest(V(0, axle, z))
		for _, x := range []float64{-4, 0, 4} {
			g.Fill(Capsule{hub, fromRest(V(x, deck-0.6, z)), 0.55})
		}
		g.Fill(Capsule{fromRest(V(-4, deck-0.6, z)), fromRest(V(4, deck-0.6, z)), 0.5})
	}
	g.Fill(Capsule{V(c.X, axle, c.Y-B-2.6), V(c.X, axle, c.Y+B+2.6), 0.8})
	// A-frames.
	g.Use(Red)
	for _, sz := range []float64{-1, 1} {
		z := c.Y + sz*(B+1.8)
		apex := V(c.X, axle+0.6, z)
		for _, sx := range []float64{-1, 1} {
			foot := V(c.X+sx*7.5, 0.3, z+sz*1.0)
			g.Fill(Capsule{foot, apex, 0.7})
		}
		g.Fill(Capsule{V(c.X-4.6, 5, z+sz*0.6), V(c.X+4.6, 5, z+sz*0.6), 0.5})
	}
	// Loading deck.
	g.Use(Wood).Fill(Block(V(c.X-4, 0, c.Y-B-2.4), V(c.X+4, 1.2, c.Y-B-0.4)))
	g.Use(Stone).Fill(Block(V(c.X-9, 0, c.Y-B-3.2), V(c.X+9, 0.5, c.Y+B+2.4)))
}

// funHouse is a fun house with a gold false front, a big arched door, porthole
// windows and a striped gable roof, facing the pond path (+Z).
func (f Funfair) funHouse(g *Grid) {
	c := funfairFunHouse
	x0, x1, z0, z1 := c.X-4.5, c.X+4.5, c.Y-4.5, c.Y+3.5
	g.Use(Navy).Fill(Block(V(x0, 0, z0), V(x1, 6.5, z1)))
	g.Carve(Block(V(x0+1, 0.5, z0+1), V(x1-1, 6.5, z1-1)))
	roof := Gable{P(x0-0.6, z0-0.5), P(x1+0.6, z1+0.6), 6.5, AxisZ}
	g.Use(Red).Fill(roof)
	g.Paint(White, Func{roof.Bounds(), func(p Vec3) bool { return math.Mod(p.Z+60, 2) < 1 }})
	// Side and back windows, gold stars between them.
	for z := z0 + 1.5; z < z1-1; z += 2.5 {
		g.Carve(Block(V(x0-0.2, 2.5, z), V(x1+0.2, 4.5, z+1)))
	}
	g.Carve(Block(V(c.X-2.5, 2.5, z0-0.2), V(c.X-1.5, 4.5, z0+0.2)), Block(V(c.X+1.5, 2.5, z0-0.2), V(c.X+2.5, 4.5, z0+0.2)))
	g.Paint(Gold, Func{Box{V(x0-0.2, 0.5, z0-0.2), V(x1+0.2, 6.5, z1)}, func(p Vec3) bool {
		return p.Y > 5 && p.Y < 5.8 && Hash01(funfairIdx(p.X), funfairIdx(p.Z), 41) < 0.35
	}})
	// False front with a rounded crest.
	fz := z1 + 1
	half := (x1 - x0 + 2.4) / 2
	front := Func{Box{V(x0-1.2, 0, z1), V(x1+1.2, 13, fz)}, func(p Vec3) bool {
		u := (p.X - c.X) / half
		return math.Abs(u) <= 1 && p.Y <= 8.5+3.5*math.Sqrt(math.Max(0, 1-u*u))
	}}
	g.Use(Gold).Fill(front)
	g.Paint(Red, Func{front.Box, func(p Vec3) bool {
		u := (p.X - c.X) / half
		inner := math.Abs(u) <= 0.86 && p.Y <= 7.9+3.0*math.Sqrt(math.Max(0, 1-u*u))
		return !inner
	}})
	g.Carve(Arch{V(c.X, 0.5, fz-0.5), 1.8, 2.4, 3, AxisZ})
	g.Use(Navy).Fill(Block(V(c.X-2, 0.5, z1-0.8), V(c.X+2, 4.5, z1-0.2)))
	for _, dx := range []float64{-3, 3} {
		g.Carve(Func{Box{V(c.X+dx-1.2, 4, z1-1), V(c.X+dx+1.2, 8, fz+0.5)}, func(p Vec3) bool {
			return math.Hypot(p.Y-6, p.X-c.X-dx) <= 1.0
		}})
	}
	g.Use(White).Fill(Func{Box{V(x0, 7, fz), V(x1, 12, fz+0.4)}, func(p Vec3) bool {
		return math.Hypot(p.Y-9.3, p.X-c.X) <= 1.5
	}})
	g.Use(Water).Fill(Func{Box{V(x0, 7, fz+0.2), V(x1, 12, fz+0.8)}, func(p Vec3) bool {
		return math.Hypot(p.Y-9.3, p.X-c.X) <= 0.7
	}})
	g.Use(Gold)
	for i := 0; i <= 10; i++ {
		a := math.Pi * float64(i) / 10
		g.Fill(Sphere(V(c.X+half*1.05*math.Cos(a), 8.5+3.9*math.Sin(a), fz+0.3), 0.45))
	}
}

// avenue adds lamp posts, bunting across the avenue and balloon sellers.
func (f Funfair) avenue(g *Grid) {
	lamp := func(at Vec2) {
		g.Use(Charcoal).Fill(Capsule{at.At(0.3), at.At(4.3), 0.42})
		g.Use(Gold).Fill(Sphere(at.At(4.8), 0.65))
	}
	for z := 12.0; z <= 32; z += 5 {
		lamp(P(-4.6, z))
		lamp(P(4.6, z))
	}
	for i := 0; i < 10; i++ {
		a := 2*math.Pi*float64(i)/10 + 0.3
		at := Polar(funfairCarousel, 12.6, a)
		if math.Abs(at.X) < 5 && at.Y > 8 {
			continue
		}
		lamp(at)
	}
	// Bunting strings with alternating flags.
	flags := []Color{Red, Gold, Navy, White, Water}
	for _, z := range []float64{17, 22.5, 28} {
		var prev Vec3
		for i := 0; i <= 36; i++ {
			t := float64(i) / 36
			x := -6.8 + 13.6*t
			y := 5.9 - 1.4*(1-(2*t-1)*(2*t-1))
			p := V(x, y, z)
			if i > 0 {
				g.Use(White).Fill(Capsule{prev, p, 0.36})
			}
			if i%3 == 1 {
				g.Use(flags[(i/3)%len(flags)]).Fill(Capsule{p.Add(V(0, -0.4, 0)), p.Add(V(0, -0.9, 0)), 0.4})
			}
			prev = p
		}
	}
	// Balloon sellers.
	for k, at := range []Vec2{P(-2.6, 25.5), P(6, 1), P(-9, -6)} {
		g.Use(Charcoal).Fill(Capsule{at.At(0.3), at.At(1.6), 0.4})
		top := at.At(4.6)
		colors := []Color{Red, Gold, Water, White, Grass, Navy}
		for i := 0; i < 6; i++ {
			a := float64(i)*2*math.Pi/6 + float64(k)
			b := top.Add(V(0.9*math.Cos(a), 0.6*math.Sin(float64(i)*2.1), 0.9*math.Sin(a)))
			g.Use(colors[(i+k)%len(colors)]).Fill(Sphere(b, 0.55))
		}
		g.Use(White).Fill(Capsule{at.At(1.6), top.Sub(V(0, 0.6, 0)), 0.3})
	}
}

// people stroll the paths: tiny figures of three or four stacked cells.
func (f Funfair) people(g *Grid) {
	span := g.Span()
	shirts := []Color{Red, Navy, Gold, White, Water, Grass, Brick, Terracotta}
	for x := -45.0; x <= 45; x += 1.6 {
		for z := -33.0; z <= 33; z += 1.6 {
			i, j := funfairIdx(x), funfairIdx(z)
			jx := x + (Hash01(i, j, 61)-0.5)*1.0
			jz := z + (Hash01(i, j, 62)-0.5)*1.0
			if Hash01(i, j, 63) > 0.24 {
				continue
			}
			ground, ok := g.ColorOf(Nearest(V(jx, 0, jz).Scale(1 / span)))
			if !ok || (ground != Sand && ground != Terracotta && ground != Brick) {
				continue
			}
			blocked := false
			for _, dx := range []float64{-0.8, 0, 0.8} {
				for _, dz := range []float64{-0.8, 0, 0.8} {
					for _, y := range []float64{0.4, 0.8, 1.2, 1.6, 2.0, 2.4} {
						if g.HasAt(V(jx+dx, y, jz+dz)) {
							blocked = true
						}
					}
				}
			}
			if blocked {
				continue
			}
			tall := Hash01(i, j, 64) < 0.7
			g.Use(Charcoal).AddAt(V(jx, 0.4, jz))
			g.Use(shirts[int(Hash01(i, j, 65)*float64(len(shirts)))]).AddAt(V(jx, 0.8, jz))
			head := 1.2
			if tall {
				g.AddAt(V(jx, 1.2, jz))
				head = 1.6
			}
			g.Use(Sand).AddAt(V(jx, head, jz))
		}
	}
}
