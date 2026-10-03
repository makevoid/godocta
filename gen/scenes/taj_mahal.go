package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// TajMahal is the white marble mausoleum on its square plinth: a chamfered main block
// with deep pointed iwans on every face, the onion dome on its drum, four roof chhatris
// and four corner minarets. It stands on a red sandstone riverside terrace between the
// mosque and its twin, the jawab, above the charbagh garden: a long reflecting pool,
// a cross channel, a raised central tank, paths, cypress rows and the great gateway.
// The main facade looks down the garden to +Z.
type TajMahal struct{}

func init() { Register(TajMahal{}) }

func (TajMahal) Name() string  { return "taj_mahal" }
func (TajMahal) Title() string { return "Taj Mahal" }

// pointed is a two-centred (Mughal) arch opening along Z, centred at x = cx, from z0
// to z1: jambs up to spring, then two arcs of radius 1.6 × the half-width meeting at
// an apex about 1.48 half-widths above the spring.
func (TajMahal) pointed(cx, sill, spring, hw, z0, z1 float64) Func {
	r := 1.6 * hw
	apex := spring + math.Sqrt(r*r-(r-hw)*(r-hw))
	return Func{Box{V(cx-hw, sill, z0), V(cx+hw, apex, z1)}, func(p Vec3) bool {
		ax := math.Abs(p.X - cx)
		if ax > hw {
			return false
		}
		rise := p.Y - spring
		if rise <= 0 {
			return true
		}
		d := ax + r - hw
		return d*d+rise*rise <= r*r
	}}
}

// apex is the height of a pointed arch's crown.
func (TajMahal) apex(spring, hw float64) float64 {
	r := 1.6 * hw
	return spring + math.Sqrt(r*r-(r-hw)*(r-hw))
}

// onionRadius is the onion dome's profile: a bulb swelling just above its base,
// pinched by a concave taper into the finial. u runs from 0 (base) to 1 (tip).
func (TajMahal) onionRadius(u float64) float64 {
	if u < 0 || u > 1 {
		return -1
	}
	b := (u - 0.36) / 0.6
	bulb := math.Sqrt(math.Max(0, 1-b*b))
	tip := 7.2 * math.Pow(1-u, 1.8)
	return math.Min(bulb, tip)
}

// onion is an onion dome of maximum radius rm and height h from y0, shrunk by inset.
func (t TajMahal) onion(c Vec2, y0, h, rm, inset float64) Func {
	return Func{Box{V(c.X-rm, y0, c.Y-rm), V(c.X+rm, y0+h, c.Y+rm)}, func(p Vec3) bool {
		r := rm*t.onionRadius((p.Y-y0)/h) - inset
		return r > 0 && p.Plan().Sub(c).Length() <= r
	}}
}

// dome places a hollow (when large enough) onion dome with a gold finial.
func (t TajMahal) dome(g *Grid, c Vec2, y0, h, rm, spire float64) {
	shell := Solid(t.onion(c, y0, h, rm, 0))
	if rm > 3 {
		shell = Subtract(shell, t.onion(c, y0-0.01, h, rm, 1.2))
	}
	g.With(White, func() { g.FillVolume(shell) })
	top := y0 + h
	g.With(Gold, func() {
		g.Fill(Cylinder{Center: c, Radius: 0.5, Y0: top - 1, Y1: top + spire})
		if spire > 2 {
			g.FillVolume(Sphere(c.At(top+spire*0.25), 0.4+spire*0.12), Sphere(c.At(top+spire*0.55), 0.3+spire*0.09))
		}
	})
}

// chhatri is an open domed kiosk: a platform, a ring of n pillars, an entablature with
// a projecting chajja, and an onion dome.
func (t TajMahal) chhatri(g *Grid, c Vec2, base, r float64, n int, pillar Color) {
	g.With(White, func() { g.Fill(Cylinder{Center: c, Radius: r + 0.3, Y0: base, Y1: base + 0.8}) })
	top := base + 0.8 + 1.15*r
	g.With(pillar, func() {
		for k := 0; k < n; k++ {
			at := Polar(c, r-0.5, (float64(k)+0.5)*2*math.Pi/float64(n))
			g.Fill(Cylinder{Center: at, Radius: 0.55, Y0: base + 0.8, Y1: top})
		}
	})
	g.With(White, func() {
		g.FillVolume(Cylinder{Center: c, Radius: r + 0.3, Inner: math.Max(r-1.3, 0), Y0: top, Y1: top + 0.8})
		g.FillVolume(Cylinder{Center: c, Radius: r + 0.9, Inner: r - 0.3, Y0: top + 0.4, Y1: top + 0.8})
	})
	t.dome(g, c, top+0.8, 1.35*r, r-0.1, 0.6+0.45*r)
}

// minaret is a tapering hollow shaft with three corbelled balconies and a crowning chhatri.
func (t TajMahal) minaret(g *Grid, c Vec2, base, height float64) {
	const r0, r1 = 2.3, 1.7
	radius := func(y float64) float64 { return r0 + (r1-r0)*(y-base)/height }
	top := base + height
	g.Use(White)
	g.FillVolume(Cylinder{Center: c, Radius: 3, Y0: base, Y1: base + 1.6})
	g.FillVolume(Subtract(Frustum{c, r0, r1, base, top}, Frustum{c, r0 - 1.1, r1 - 1.1, base, top}))
	for k := 1; k <= 3; k++ {
		y := base + height*float64(k)/3
		r := radius(y)
		g.FillVolume(Frustum{c, r, r + 1.2, y - 1.4, y})
		g.FillVolume(Cylinder{Center: c, Radius: r + 1.2, Inner: r + 0.45, Y0: y, Y1: y + 0.9})
		g.Paint(Sand, Cylinder{Center: c, Radius: r + 1.3, Y0: y - 0.5, Y1: y})
	}
	t.chhatri(g, c, top+0.9, 1.9, 4, White)
}

// mosque is a red sandstone hall with a white-outlined central pishtaq, open side bays,
// three white domes and four corner kiosks. Its facade faces +Z at the origin; place
// rotates and moves every solid into the world.
func (t TajMahal) mosque(g *Grid, place func(Solid) Solid, at func(Vec2) Vec2) {
	const base, top = 1.5, 10.5
	g.Use(Brick)
	g.Fill(place(Subtract(Block(V(-14, base, -6), V(14, top, 6)), Block(V(-13, base, -5), V(13, top-0.6, 5)))))
	// Central pishtaq.
	g.Fill(place(Block(V(-4.5, base, 5), V(4.5, 13.2, 6.6))))
	g.Fill(place(Block(V(-3.8, base, 2), V(3.8, 12, 6))))
	g.Carve(place(t.pointed(0, base, 7, 3, 3, 7)))
	g.Paint(White, place(Subtract(Block(V(-3.7, base, 6.5), V(3.7, 12.2, 6.6)), t.pointed(0, base, 7, 3.2, 6, 7))))
	g.Paint(White, place(Block(V(-4.5, 13, 5), V(4.5, 13.2, 6.6))))
	// Side bays: open arches with a recessed back wall.
	for _, x := range []float64{-11, -7, 7, 11} {
		g.Fill(place(Subtract(Block(V(x-2, base, 2.8), V(x+2, 9.6, 6)), Block(V(x-1.4, base, 4), V(x+1.4, 9.6, 6)))))
		g.Carve(place(t.pointed(x, base, 5.2, 1.5, 4, 7)))
	}
	g.Paint(White, place(Subtract(Block(V(-14, top-0.6, -6), V(14, top, 6.6)), Block(V(-13.2, top-0.6, -5.2), V(13.2, top, 5.2)))))
	// Domes on drums.
	for _, d := range []struct{ x, r float64 }{{0, 4.3}, {-8.6, 3.1}, {8.6, 3.1}} {
		c := at(P(d.x, -0.5))
		g.With(Brick, func() { g.FillVolume(Cylinder{Center: c, Radius: 0.82 * d.r, Y0: top, Y1: top + 2.8}) })
		g.With(White, func() { g.FillVolume(Cylinder{Center: c, Radius: 0.82*d.r + 0.3, Y0: top + 2.8, Y1: top + 3.3}) })
		t.dome(g, c, top+3.3, 1.9*d.r, d.r, 1+0.5*d.r)
	}
	// Octagonal corner towers with white kiosks.
	for _, x := range []float64{-14, 14} {
		for _, z := range []float64{-6, 6} {
			c := at(P(x, z))
			g.With(Brick, func() { g.FillVolume(Cylinder{Center: c, Radius: 1.4, Y0: base, Y1: top + 2}) })
			t.chhatri(g, c, top+2, 1.5, 4, White)
		}
	}
}

func (t TajMahal) Build(g *Grid) Order {
	origin := P(0, 0)
	turn := func(s Solid, a float64) Solid { return RotateY(s, origin, a) }
	quarter := func(k int) float64 { return float64(k) * math.Pi / 2 }

	const (
		terrace    = 1.5  // Red sandstone terrace top.
		plinthTop  = 5.0  // Marble plinth top.
		half       = 15.0 // Main block half-width.
		chamfer    = 5.0  // Corner cut along each side.
		roof       = 23.0 // Main block parapet top.
		gardenZ0   = 27.0 // Terrace front edge.
		gardenZ1   = 80.0 // Garden end (gateway axis).
		gardenHalf = 43.0
	)
	tankZ := (gardenZ0 + gardenZ1) / 2

	// Ground: river behind, terrace, garden.
	g.Use(Water).Fill(Block(V(-gardenHalf-6, 0, -36), V(gardenHalf+6, 0.4, -27)))
	g.Use(Brick).Fill(Subtract(Block(V(-gardenHalf, 0, -27), V(gardenHalf, terrace, gardenZ0)),
		Block(V(-gardenHalf+1, 0, -26), V(gardenHalf-1, terrace-0.4, gardenZ0-1))))
	// Red and white geometric paving on the terrace.
	g.Paint(Sand, Func{Box{V(-gardenHalf+1, terrace-0.4, -26), V(gardenHalf-1, terrace, gardenZ0-1)}, func(p Vec3) bool {
		return (int(math.Floor(p.X/3))+int(math.Floor(p.Z/3)))%2 == 0
	}})
	g.Use(Grass).Fill(Block(V(-gardenHalf, 0, gardenZ0), V(gardenHalf, 0.4, gardenZ1)))

	// The plinth: a hollow marble platform with a blind arcade.
	g.Use(White)
	g.Fill(Subtract(Block(V(-25, terrace, -25), V(25, plinthTop, 25)), Block(V(-23.4, terrace, -23.4), V(23.4, plinthTop-0.5, 23.4))))
	for k := 0; k < 4; k++ {
		for x := -21.6; x <= 21.7; x += 3.6 {
			g.Paint(Sand, turn(Block(V(x-1.1, terrace+0.4, 23.9), V(x+1.1, plinthTop-0.6, 24.5)), quarter(k)))
			g.Carve(turn(t.pointed(x, terrace+0.6, 2.9, 1.1, 24.4, 26), quarter(k)))
		}
	}
	g.Paint(Sand, Block(V(-25, plinthTop-0.4, -25), V(25, plinthTop, 25)))
	g.Paint(White, Block(V(-24.6, plinthTop-0.4, -24.6), V(24.6, plinthTop, 24.6)))

	// Main block: a hollow chamfered square.
	octagon := func(h, c, y0, y1 float64) Func {
		return Func{Box{V(-h, y0, -h), V(h, y1, h)}, func(p Vec3) bool {
			return math.Abs(p.X)+math.Abs(p.Z) <= 2*h-c
		}}
	}
	const wall = 1.2
	g.Fill(Subtract(octagon(half, chamfer, plinthTop, roof), octagon(half-wall, chamfer-wall*(2-math.Sqrt2), plinthTop, roof-0.6)))
	// Each main face: a raised pishtaq with a deep iwan, door and window, and two
	// stacked niches on either side.
	iwanSpring, iwanHW := plinthTop+10.5, 4.6
	for k := 0; k < 4; k++ {
		a := quarter(k)
		g.Fill(turn(Block(V(-6.5, plinthTop, half-0.5), V(6.5, roof+3, half+0.6)), a))
		g.Fill(turn(Block(V(-5.7, plinthTop, half-4.5), V(5.7, roof, half)), a))
		g.Carve(turn(t.pointed(0, plinthTop+0.4, iwanSpring, iwanHW, half-3.4, half+1), a))
		// Calligraphy band framing the iwan.
		g.Paint(Charcoal, turn(Subtract(t.pointed(0, plinthTop+0.4, iwanSpring, iwanHW+0.8, half+0.3, half+0.6),
			t.pointed(0, plinthTop, iwanSpring, iwanHW+0.1, half, half+1)), a))
		g.Paint(Sand, turn(Block(V(-iwanHW, plinthTop, half-4.6), V(iwanHW, roof, half-3.4)), a))
		g.Carve(turn(t.pointed(0, plinthTop+0.4, plinthTop+3.6, 1.6, half-6, half-3), a))
		g.Carve(turn(t.pointed(0, plinthTop+9, plinthTop+10.8, 1.4, half-6, half-3), a))
		for _, x := range []float64{-8.3, 8.3} {
			for _, sill := range []float64{plinthTop + 1, plinthTop + 9.2} {
				spring := sill + 4.2
				g.Fill(turn(Block(V(x-1.7, sill-0.4, half-2.4), V(x+1.7, t.apex(spring, 1.2)+0.8, half)), a))
				g.Carve(turn(t.pointed(x, sill, spring, 1.2, half-1.2, half+1), a))
			}
		}
	}
	// Chamfered faces: two stacked niches each.
	diag := (2*half - chamfer) / math.Sqrt2
	for k := 0; k < 4; k++ {
		a := quarter(k) + math.Pi/4
		for _, sill := range []float64{plinthTop + 1, plinthTop + 9.2} {
			spring := sill + 4.6
			g.Fill(turn(Block(V(-2.6, sill-0.4, diag-2.4), V(2.6, t.apex(spring, 1.8)+0.8, diag)), a))
			g.Carve(turn(t.pointed(0, sill, spring, 1.8, diag-1.2, diag+1), a))
		}
	}
	// Parapet band and corner pinnacles (guldastas).
	g.Paint(Sand, Subtract(octagon(half+1, chamfer, roof-0.5, roof), octagon(half-0.6, chamfer, roof-0.5, roof)))
	for k := 0; k < 4; k++ {
		for _, v := range []Vec2{P(half-chamfer, half), P(half, half-chamfer), P(6.5, half+0.1), P(-6.5, half+0.1)} {
			c := v.Rotate(quarter(k))
			topY := roof + 2.6
			if math.Abs(v.X) == 6.5 {
				topY = roof + 4.2
			}
			g.With(White, func() { g.FillVolume(Cylinder{Center: c, Radius: 0.75, Y0: plinthTop, Y1: topY}) })
			g.With(Gold, func() { g.Fill(Cylinder{Center: c, Radius: 0.5, Y0: topY, Y1: topY + 1}) })
		}
	}

	// The great dome on its drum.
	const drumTop, domeH, domeR = roof + 6.5, 19.0, 9.4
	g.Use(White).FillVolume(Cylinder{Center: origin, Radius: 7.7, Inner: 6.4, Y0: roof - 1, Y1: drumTop})
	for k := 0; k < 16; k++ {
		a := (float64(k) + 0.5) * math.Pi / 8
		g.Carve(turn(t.pointed(0, roof+0.6, roof+2, 0.9, 7, 8.6), a))
	}
	g.Paint(Sand, Cylinder{Center: origin, Radius: 7.9, Y0: drumTop - 0.5, Y1: drumTop})
	g.Paint(Sand, Cylinder{Center: origin, Radius: 7.9, Y0: roof + 3.4, Y1: roof + 3.8})
	t.dome(g, origin, drumTop, domeH, domeR, 6)
	// Lotus crown: petals hanging from a band below the finial.
	low, band := drumTop+0.72*domeH, drumTop+0.84*domeH
	g.Paint(Sand, Func{Box{V(-domeR, low, -domeR), V(domeR, drumTop+domeH, domeR)}, func(p Vec3) bool {
		f := math.Mod(p.Plan().Angle()/(2*math.Pi)*16+32, 1)
		return p.Y >= band || p.Y >= low+(band-low)*2*math.Abs(f-0.5)
	}})

	// Four chhatris around the dome.
	for k := 0; k < 4; k++ {
		c := P(10.6, 10.6).Rotate(quarter(k))
		t.chhatri(g, c, roof-0.4, 3, 8, White)
	}
	// Four minarets at the plinth corners.
	for k := 0; k < 4; k++ {
		t.minaret(g, P(21.8, 21.8).Rotate(quarter(k)), plinthTop, 18)
	}

	// Mosque (west, -X) and jawab (east, +X), both facing the tomb.
	for _, side := range []float64{-1, 1} {
		angle := side * math.Pi / 2
		offset := V(side*35, 0, 0)
		place := func(s Solid) Solid { return Translate(RotateY(s, origin, angle), offset) }
		at := func(p Vec2) Vec2 { return p.Rotate(angle).Add(offset.Plan()) }
		t.mosque(g, place, at)
	}

	// Steps down from the terrace to the garden.
	g.Use(Sand).Fill(Block(V(-5, 0, gardenZ0), V(5, 1.0, gardenZ0+1.2)), Block(V(-5, 0, gardenZ0), V(5, 0.6, gardenZ0+2.4)))

	// The charbagh: channels, paths and parterres.
	inGarden := func(f func(x, z float64) bool) Func {
		return Func{Box{V(-gardenHalf, 0, gardenZ0), V(gardenHalf, 0.4, gardenZ1)}, func(p Vec3) bool { return f(p.X, p.Z) }}
	}
	g.Paint(Sand, inGarden(func(x, z float64) bool {
		ax, dz := math.Abs(x), math.Abs(z-tankZ)
		return ax <= 5.5 || dz <= 5.5 || ax >= gardenHalf-3 || z >= gardenZ1-3 ||
			math.Abs(ax-25) <= 1 || math.Abs(dz-15) <= 1
	}))
	g.Paint(Water, inGarden(func(x, z float64) bool {
		return math.Abs(x) <= 1.6 || math.Abs(z-tankZ) <= 1.6
	}))
	// Marble curbs along the channels.
	g.Use(White).Fill(Func{Box{V(-gardenHalf+3, 0.4, gardenZ0+2.5), V(gardenHalf-3, 0.8, gardenZ1-3)}, func(p Vec3) bool {
		ax, dz := math.Abs(p.X), math.Abs(p.Z-tankZ)
		curbX := ax > 1.6 && ax <= 2.4 && dz > 2.4
		curbZ := dz > 1.6 && dz <= 2.4 && ax > 2.4
		return curbX || curbZ
	}})
	// Raised central tank.
	g.Fill(Block(V(-6, 0, tankZ-6), V(6, 1.2, tankZ+6)))
	g.Paint(Water, Block(V(-4.2, 0.8, tankZ-4.2), V(4.2, 1.2, tankZ+4.2)))
	g.Fill(Cylinder{Center: P(0, tankZ), Radius: 0.6, Y0: 1.2, Y1: 2.2})

	// Cypress rows along the channels; broad trees around the edges.
	g.Use(Foliage)
	for _, x := range []float64{-6.8, 6.8} {
		for z := gardenZ0 + 4.5; z < gardenZ1-4; z += 3.5 {
			if math.Abs(z-tankZ) < 7.5 {
				continue
			}
			h := 2.8 + 0.6*Hash01(int(x), int(z*2))
			g.FillVolume(Ellipsoid{V(x, 0.4+h, z), V(1.05, h, 1.05)})
		}
	}
	for _, z := range []float64{tankZ - 6.8, tankZ + 6.8} {
		for x := 9.5; x < gardenHalf-4; x += 3.5 {
			for _, sx := range []float64{-1, 1} {
				h := 2.8 + 0.6*Hash01(int(sx*x*2), int(z))
				g.FillVolume(Ellipsoid{V(sx*x, 0.4+h, z), V(1.05, h, 1.05)})
			}
		}
	}
	broad := func(c Vec2, seed int) {
		h := 4 + 1.5*Hash01(seed, 7)
		r := 2.2 + 0.6*Hash01(seed, 11)
		g.With(Charcoal, func() { g.Fill(Cylinder{Center: c, Radius: 0.5, Y0: 0.4, Y1: h - 1}) })
		color := Foliage
		if Hash01(seed, 3) < 0.35 {
			color = Grass
		}
		g.With(color, func() {
			g.FillVolume(Subtract(Ellipsoid{c.At(h), V(r, r*0.85, r)}, Ellipsoid{c.At(h), V(r-1.1, r*0.85-1.1, r-1.1)}))
		})
	}
	seed := 0
	for z := gardenZ0 + 5; z < gardenZ1-6; z += 5.5 {
		for _, x := range []float64{-gardenHalf + 5.5, gardenHalf - 5.5} {
			seed++
			if math.Abs(z-tankZ) < 3 {
				continue
			}
			broad(P(x+Hash01(seed, 1)-0.5, z), seed)
		}
	}
	// Garden walls.
	g.Use(Brick).Fill(
		Block(V(-gardenHalf, 0, gardenZ0), V(-gardenHalf+1, 3, gardenZ1)),
		Block(V(gardenHalf-1, 0, gardenZ0), V(gardenHalf, 3, gardenZ1)),
		Block(V(-gardenHalf, 0, gardenZ1-1), V(gardenHalf, 3, gardenZ1)),
	)
	g.Paint(White, Block(V(-gardenHalf, 2.6, gardenZ0), V(gardenHalf, 3, gardenZ1)))

	// The great gateway (Darwaza), straddling the far wall.
	t.gateway(g, P(0, gardenZ1))

	return Outward(origin)
}

// gateway is the red sandstone Darwaza: a hollow block with a white-framed pishtaq
// and through-passage on both faces, stacked side niches, octagonal corner towers with
// kiosks and a row of small white domes over each pishtaq.
func (t TajMahal) gateway(g *Grid, c Vec2) {
	const hw, hd, top = 13.0, 5.5, 12.5
	at := func(s Solid) Solid { return Translate(s, c.At(0)) }
	g.Use(Brick)
	g.Fill(at(Subtract(Block(V(-hw, 0, -hd), V(hw, top, hd)), Block(V(-hw+1, 0, -hd+1), V(hw-1, top-0.6, hd-1)))))
	for _, s := range []float64{-1, 1} {
		face := func(b Solid) Solid { return at(RotateY(b, P(0, 0), map[bool]float64{true: 0, false: math.Pi}[s > 0])) }
		g.Fill(face(Block(V(-5.5, 0, hd-0.5), V(5.5, top+2.5, hd+0.6))))
		g.Fill(face(Block(V(-4.6, 0, hd-3.5), V(4.6, top, hd))))
		g.Carve(face(t.pointed(0, 0.4, 6.2, 3.4, -1, hd+1)))
		g.Paint(White, face(Subtract(t.pointed(0, 0, 6.2, 4.2, hd+0.3, hd+0.6), t.pointed(0, 0, 6.2, 3.5, hd, hd+1))))
		g.Paint(White, face(Block(V(-5.5, top+2.1, hd-0.5), V(5.5, top+2.5, hd+0.6))))
		for _, x := range []float64{-9, 9} {
			for _, sill := range []float64{1, 6.8} {
				spring := sill + 2.8
				g.Fill(face(Block(V(x-2, sill-0.4, hd-2.2), V(x+2, t.apex(spring, 1.4)+0.8, hd))))
				g.Carve(face(t.pointed(x, sill, spring, 1.4, hd-1.1, hd+1)))
			}
		}
		// Row of small domes above the pishtaq.
		for k := -2; k <= 2; k++ {
			d := c.Add(P(float64(k)*2.1, s*(hd-0.8)))
			g.With(White, func() { g.Fill(Cylinder{Center: d, Radius: 0.9, Y0: top + 2.5, Y1: top + 3.1}) })
			t.dome(g, d, top+3.1, 1.6, 1.0, 0.8)
		}
	}
	g.Paint(White, at(Subtract(Block(V(-hw, top-0.5, -hd), V(hw, top, hd)), Block(V(-hw+0.8, top-0.5, -hd+0.8), V(hw-0.8, top, hd-0.8)))))
	for _, x := range []float64{-hw, hw} {
		for _, z := range []float64{-hd, hd} {
			p := c.Add(P(x, z))
			g.With(Brick, func() { g.FillVolume(Cylinder{Center: p, Radius: 1.6, Y0: 0, Y1: top + 2}) })
			t.chhatri(g, p, top+2, 1.8, 4, White)
		}
	}
	g.Use(Brick)
}
