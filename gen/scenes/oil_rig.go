package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// OilRig is a North Sea fixed production platform: a yellow steel jacket of four
// cross-braced legs and a conductor bundle rising from a choppy sea; a three-level
// topside (cellar, main and weather decks) with separators, pipe racks, colored
// process modules, stair flights and handrails; the drilling derrick on its
// substructure with a red windwall; an accommodation block with rows of windows,
// lifeboats and the octagonal helideck cantilevered off the (+X, +Z) corner; two
// pedestal cranes with lattice booms (one lifting a container off the supply vessel);
// a raked lattice flare boom with its flame; a bridge-linked wellhead platform; and an
// anchor-handling supply vessel alongside. The sea surface is the layer at y = 0.
type OilRig struct{}

func init() { Register(OilRig{}) }

func (OilRig) Name() string  { return "oil_rig" }
func (OilRig) Title() string { return "Oil Rig" }

const (
	oilRigCellar = 14.0 // Underside of each topside deck plate.
	oilRigMain   = 20.5
	oilRigTop    = 27.0
	oilRigPlate  = 0.8
	oilRigHX     = 24.0 // Topside half-length (X) and half-width (Z).
	oilRigHZ     = 15.0
	oilRigShipZ  = 26.0 // Supply vessel centerline.
)

var oilRigSea = Box{V(-50, -0.2, -20), V(56, 0.2, 33)}

func oilRigCross(a, b Vec3) Vec3 {
	return V(a.Y*b.Z-a.Z*b.Y, a.Z*b.X-a.X*b.Z, a.X*b.Y-a.Y*b.X)
}

func oilRigNorm(a Vec3) Vec3 { return a.Scale(1 / a.Length()) }

// cell is the lattice cell nearest a world point.
func (OilRig) cell(g *Grid, p Vec3) Cell { return Nearest(p.Scale(1 / g.Span())) }

// truss fills a four-chord lattice from a to b. up picks the first cross direction (u);
// hu/hv are the half-widths along u and v at a (0) and b (1). Panels are X-braced or
// Warren (zigzag) laced.
func (OilRig) truss(g *Grid, a, b, up Vec3, hu0, hv0, hu1, hv1 float64, n int, chord, brace float64, xBrace bool, chordColor, braceColor Color) {
	d := oilRigNorm(b.Sub(a))
	v := oilRigNorm(oilRigCross(d, up))
	u := oilRigCross(v, d)
	su := [4]float64{1, -1, -1, 1}
	sv := [4]float64{1, 1, -1, -1}
	corner := func(t float64, k int) Vec3 {
		hu := hu0 + (hu1-hu0)*t
		hv := hv0 + (hv1-hv0)*t
		return a.Add(b.Sub(a).Scale(t)).Add(u.Scale(su[k] * hu)).Add(v.Scale(sv[k] * hv))
	}
	g.With(braceColor, func() {
		for i := 0; i <= n; i++ {
			t := float64(i) / float64(n)
			for k := 0; k < 4; k++ {
				g.Fill(Capsule{corner(t, k), corner(t, (k+1)%4), brace})
			}
		}
		for i := 0; i < n; i++ {
			t0, t1 := float64(i)/float64(n), float64(i+1)/float64(n)
			for k := 0; k < 4; k++ {
				k2 := (k + 1) % 4
				if xBrace || (i+k)%2 == 0 {
					g.Fill(Capsule{corner(t0, k), corner(t1, k2), brace})
				}
				if xBrace || (i+k)%2 == 1 {
					g.Fill(Capsule{corner(t0, k2), corner(t1, k), brace})
				}
			}
		}
	})
	g.With(chordColor, func() {
		for k := 0; k < 4; k++ {
			g.Fill(Capsule{corner(0, k), corner(1, k), chord})
		}
	})
}

// module is a box between y0 and y1 with louvre panels and a door; big ones are hollow.
func (r OilRig) module(g *Grid, x0, z0, x1, z1, y0, y1 float64, wall, accent Color, seed int) {
	g.With(wall, func() { g.Fill(Block(V(x0, y0, z0), V(x1, y1, z1))) })
	box := Box{V(x0, y0, z0), V(x1, y1, z1)}
	// Louvre panels: dark vertical bands on the long faces, a colored trim at the top.
	g.Paint(accent, Block(V(x0, y1-0.6, z0), V(x1, y1, z1)))
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		if p.Y < y0+1.2 || p.Y > y1-1.4 {
			return false
		}
		along := p.X - x0
		if math.Abs(p.X-x0) < 0.6 || math.Abs(p.X-x1) < 0.6 {
			along = p.Z - z0
		}
		k := int(math.Floor(along / 3.2))
		f := along - float64(k)*3.2
		return f > 1.0 && f < 2.4 && Hash01(seed, k) < 0.6
	}})
	// A door at floor level on the +Z face.
	dx := x0 + (x1-x0)*(0.2+0.6*Hash01(seed, 9))
	g.Paint(Charcoal, Block(V(dx-0.6, y0, z1-0.5), V(dx+0.6, y0+2.2, z1)))
	if x1-x0 > 3 && z1-z0 > 3 && y1-y0 > 3 {
		g.Carve(Block(V(x0+1.1, y0+0.6, z0+1.1), V(x1-1.1, y1-1.1, z1-1.1)))
	}
}

// container is a 20-foot box with darker doors on its -X end.
func (OilRig) container(g *Grid, c Vec3, alongX bool, color Color) {
	hx, hz := 2.0, 0.85
	if !alongX {
		hx, hz = hz, hx
	}
	g.With(color, func() { g.Fill(Block(V(c.X-hx, c.Y, c.Z-hz), V(c.X+hx, c.Y+1.7, c.Z+hz))) })
	if alongX {
		g.Paint(Charcoal, Block(V(c.X-hx, c.Y, c.Z-hz), V(c.X-hx+0.3, c.Y+1.7, c.Z+hz)))
	} else {
		g.Paint(Charcoal, Block(V(c.X-hx, c.Y, c.Z-hz), V(c.X+hx, c.Y+1.7, c.Z-hz+0.3)))
	}
}

func (OilRig) sea(g *Grid) {
	g.Use(Water).Fill(Block(oilRigSea.Min, oilRigSea.Max))
	// Long swell: wave phase across a gently meandering front, darker troughs and
	// sparse whitecaps on the crests.
	phase := func(p Vec3) (float64, float64) {
		u := 0.6*p.X + 0.8*p.Z
		v := -0.8*p.X + 0.6*p.Z
		return 0.75*u + 0.8*math.Sin(0.11*v) + 0.4*math.Sin(0.27*v+1), v
	}
	g.Paint(Navy, Func{oilRigSea, func(p Vec3) bool {
		ph, v := phase(p)
		env := math.Sin(0.07*v+0.5) + 0.8*math.Sin(0.05*(p.X+p.Z)-0.3)
		return math.Sin(ph) > 0.78-0.12*env
	}})
	g.Paint(White, Func{oilRigSea, func(p Vec3) bool {
		ph, v := phase(p)
		c := Nearest(p.Scale(1 / g.Span()))
		k := int(math.Floor(v / 3))
		return math.Sin(ph) < -0.86 && Hash01(k, int(math.Floor(ph/(2*math.Pi))), 31) < 0.22 && Hash01(int(c.X), int(c.Z), 32) < 0.75
	}})
}

// jacket is a four-legged cross-braced steel jacket from the sea to the cellar deck.
func (r OilRig) jacket(g *Grid, c Vec2, tx, tz, bx, bz float64, panels int, leg float64) {
	r.truss(g, c.At(-0.3), c.At(oilRigCellar+0.2), V(1, 0, 0), bx, bz, tx, tz, panels, leg, 0.6, true, Gold, Gold)
	// Plan bracing: horizontal diagonals across each frame level.
	g.With(Gold, func() {
		for i := 1; i < panels; i++ {
			t := float64(i) / float64(panels)
			y := -0.3 + (oilRigCellar+0.5)*t
			hx, hz := bx+(tx-bx)*t, bz+(tz-bz)*t
			g.Fill(Capsule{V(c.X-hx, y, c.Y-hz), V(c.X+hx, y, c.Y+hz), 0.55})
			g.Fill(Capsule{V(c.X-hx, y, c.Y+hz), V(c.X+hx, y, c.Y-hz), 0.55})
		}
	})
	// Dark splash zone around the waterline.
	g.Paint(Charcoal, Func{Box{V(c.X-bx-2, 0.1, c.Y-bz-2), V(c.X+bx+2, 1.4, c.Y+bz+2)}, func(p Vec3) bool {
		col, _ := g.ColorOf(r.cell(g, p))
		return col == Gold
	}})
}

// deck is a plate with a yellow handrail around its edge.
func (OilRig) deck(g *Grid, x0, z0, x1, z1, y float64, rails bool) {
	g.With(Charcoal, func() { g.Fill(Block(V(x0, y, z0), V(x1, y+oilRigPlate, z1))) })
	g.Paint(Slate, Block(V(x0, y+oilRigPlate-0.3, z0), V(x1, y+oilRigPlate, z1)))
	if !rails {
		return
	}
	ry := y + oilRigPlate + 1.2
	g.With(Gold, func() {
		corners := []Vec3{V(x0+0.3, ry, z0+0.3), V(x1-0.3, ry, z0+0.3), V(x1-0.3, ry, z1-0.3), V(x0+0.3, ry, z1-0.3)}
		for k := range corners {
			a, b := corners[k], corners[(k+1)%4]
			g.Fill(Capsule{a, b, 0.45})
			n := int(b.Sub(a).Length() / 3)
			for i := 0; i <= n; i++ {
				q := a.Add(b.Sub(a).Scale(float64(i) / float64(n)))
				g.Fill(Capsule{q, V(q.X, y+oilRigPlate, q.Z), 0.4})
			}
		}
	})
}

// stairs is a switchback of two flights along X just outside a deck edge at z.
func (OilRig) stairs(g *Grid, x0, x1, y0, y1, z, dz float64) {
	mid := (y0 + y1) / 2
	g.With(Gold, func() {
		g.Fill(Capsule{V(x0, y0+0.4, z), V(x1, mid, z), 0.6})
		g.Fill(Block(V(x1, mid-0.4, z-0.7), V(x1+1.6, mid, z+dz+0.7)))
		g.Fill(Capsule{V(x1, mid, z+dz), V(x0, y1+0.4, z+dz), 0.6})
		g.Fill(Capsule{V(x1+1.5, oilRigCellar, z+dz/2), V(x1+1.5, mid, z+dz/2), 0.45})
	})
}

func (r OilRig) topside(g *Grid) {
	X, Z := oilRigHX, oilRigHZ
	r.deck(g, -X, -Z, X, Z, oilRigCellar, true)
	r.deck(g, -X, -Z, X, Z, oilRigMain, true)
	r.deck(g, -X, -Z, X, Z, oilRigTop, true)

	// Columns: the leg lines continue up through the topside; perimeter posts and
	// X-braces on the open cellar faces.
	g.With(White, func() {
		for _, x := range []float64{-X + 0.5, -12, 0, 12, X - 0.5} {
			for _, z := range []float64{-Z + 0.5, Z - 0.5} {
				g.Fill(Cylinder{Center: P(x, z), Radius: 0.6, Y0: oilRigCellar, Y1: oilRigTop + oilRigPlate})
			}
		}
		for _, x := range []float64{-17, 17} {
			for _, z := range []float64{-10.5, 10.5} {
				g.Fill(Cylinder{Center: P(x, z), Radius: 1.1, Y0: oilRigCellar, Y1: oilRigTop + oilRigPlate})
			}
		}
		for _, z := range []float64{-Z + 0.5, Z - 0.5} {
			for _, x := range []float64{-24, -12, 0, 12} {
				g.Fill(Capsule{V(x+0.5, oilRigCellar+0.8, z), V(x+11.5, oilRigMain, z), 0.45})
				g.Fill(Capsule{V(x+11.5, oilRigCellar+0.8, z), V(x+0.5, oilRigMain, z), 0.45})
			}
		}
	})

	// Cellar deck: separators, a pipe rack and pumps.
	cy := oilRigCellar + oilRigPlate
	tanks := []struct {
		x0, x1, z, rad float64
		col            Color
	}{
		{-21, -9, -5.5, 2.0, White},
		{-21, -10, 4.5, 2.0, Gold},
		{-4, 7, -8, 1.6, Slate},
		{-3, 8, 7, 1.6, White},
		{12, 21, -1.5, 1.5, Red},
		{12, 21, -9.5, 1.3, Gold},
	}
	for _, t := range tanks {
		y := cy + t.rad + 0.2
		g.With(t.col, func() { g.Fill(Capsule{V(t.x0, y, t.z), V(t.x1, y, t.z), t.rad}) })
		g.With(Charcoal, func() {
			for _, f := range []float64{0.25, 0.75} {
				x := t.x0 + (t.x1-t.x0)*f
				g.Fill(Block(V(x-0.4, cy, t.z-t.rad*0.7), V(x+0.4, y, t.z+t.rad*0.7)))
			}
		})
	}
	pipeCols := []Color{Red, Gold, White, Slate, Gold}
	for i, c := range pipeCols {
		z := -2 + float64(i)*1.0
		g.With(c, func() { g.Fill(Capsule{V(-23, oilRigMain-1.0, z), V(23, oilRigMain-1.0, z), 0.45}) })
	}
	for i := 0; i < 4; i++ {
		x := 11 + float64(i)*2.8
		g.With(Navy, func() { g.Fill(Block(V(x, cy, 8.5), V(x+1.6, cy+1.6, 10.5))) })
		g.With(Red, func() { g.Fill(Cylinder{Center: P(x+0.8, 12), Radius: 0.6, Y0: cy, Y1: cy + 1.2}) })
	}

	// Main deck: enclosed modules along the faces, open bays between them.
	my0, my1 := oilRigMain+oilRigPlate, oilRigTop
	r.module(g, -X, 7, -14, Z, my0, my1, Red, White, 1)
	r.module(g, -12.5, 8, -4, Z, my0, my1, White, Red, 2)
	r.module(g, -2.5, 8.5, 9.5, Z, my0, my1, Gold, Charcoal, 3)
	r.module(g, -X, -Z, -11, -6, my0, my1, White, Gold, 4)
	r.module(g, -9, -Z, 0.5, -7, my0, my1, Slate, White, 5)
	r.module(g, 2.5, -Z, 10, -7.5, my0, my1, Red, Gold, 6)
	r.module(g, 13, -Z, X, -10, my0, my1, White, Navy, 12)
	// Open central bay: compressor skid, a vertical scrubber and a cooler.
	g.With(Navy, func() { g.Fill(Block(V(-9, my0, -3), V(-2, my0+2.6, 2.5))) })
	g.With(White, func() { g.Fill(Cylinder{Center: P(3, -1), Radius: 1.5, Y0: my0, Y1: my1}) })
	g.With(Gold, func() { g.Fill(Capsule{V(-9, my0+1.6, 4.8), V(7, my0+1.6, 4.8), 0.9}) })

	// External stairs up the +Z face.
	r.stairs(g, -10, -4, oilRigCellar, oilRigMain, Z+0.9, 1.4)
	r.stairs(g, -10, -4, oilRigMain, oilRigTop, Z+0.9, 1.4)

	// Weather deck: pipe deck, mud module, generator package with exhausts and a
	// telecom mast, a knock-out drum and containers in a painted laydown area.
	ty := oilRigTop + oilRigPlate
	r.module(g, -X, -Z, -14.5, -6.5, ty, ty+5, Red, White, 7)
	g.With(Slate, func() {
		for i := 0; i < 3; i++ {
			for j := 0; j < 5; j++ {
				z := -5 + float64(j)*1.0
				y := ty + 0.5 + float64(i)*0.85
				g.Fill(Capsule{V(-23, y, z), V(-15.5, y, z), 0.45})
			}
		}
	})
	r.module(g, 0, -Z, 10, -6, ty, ty+4.8, White, Navy, 8)
	g.With(Charcoal, func() {
		g.Fill(Cylinder{Center: P(3, -10.5), Radius: 1.1, Y0: ty + 4.5, Y1: ty + 13})
		g.Fill(Cylinder{Center: P(7, -10.5), Radius: 1.1, Y0: ty + 4.5, Y1: ty + 12})
	})
	g.Paint(Red, Block(V(1.5, ty+11.2, -12), V(8.5, ty+12, -9)))
	r.truss(g, V(9, ty+4.8, -14), V(9, ty+20, -14), V(1, 0, 0), 1.0, 1.0, 0.6, 0.6, 5, 0.45, 0.4, true, White, Red)
	g.With(Red, func() { g.Fill(Sphere(V(9, ty+20.6, -14), 0.6)) })
	g.With(Gold, func() { g.Fill(Cylinder{Center: P(-11, 11), Radius: 1.5, Y0: ty, Y1: ty + 7}) })
	g.With(White, func() { g.Fill(Sphere(V(-11, ty+7, 11), 1.5)) })
	cols := []Color{Red, Navy, Gold, White, Slate, Red, Gold, Navy}
	k := 0
	for i := 0; i < 3; i++ {
		for j := 0; j < 2; j++ {
			c := V(1.5+float64(j)*4.4, ty, 4.5+float64(i)*2.0)
			r.container(g, c, true, cols[k%len(cols)])
			if Hash01(i, j, 5) < 0.5 {
				r.container(g, c.Add(V(0, 1.7, 0)), true, cols[(k+3)%len(cols)])
			}
			k++
		}
	}
	// Turbine compressor package between the mud module and the generators.
	r.module(g, -13, -Z, -1.5, -7.5, ty, ty+5.5, Navy, Red, 14)
	g.With(White, func() { g.Fill(Block(V(-11, ty+5.5, -13.5), V(-4, ty+7.5, -9))) })
	g.With(Charcoal, func() { g.Fill(Cylinder{Center: P(-9, -11.5), Radius: 0.9, Y0: ty + 7.5, Y1: ty + 11}) })
	// Free-fall lifeboats on launch ramps over the +Z edge.
	for _, x := range []float64{-7, -3} {
		g.With(Charcoal, func() { g.Fill(Capsule{V(x, ty+0.6, 11.5), V(x, ty+0.6, Z+1.2), 0.45}) })
		g.With(Gold, func() { g.Fill(Capsule{V(x, ty+2.6, 11.8), V(x, ty+1.0, Z+1.0), 1.1}) })
		g.Paint(Red, Func{Box{V(x-1.2, ty, 10.5), V(x+1.2, ty+4, Z+2.5)}, func(p Vec3) bool {
			return p.Y > ty+2.4-0.4*(p.Z-11.8)
		}})
	}
	// Central process module with a fin-fan cooler on its roof.
	r.module(g, -0.5, -5, 10, 1.5, ty, ty+6, Gold, White, 13)
	g.With(Charcoal, func() { g.Fill(Block(V(0, ty+6, -4.5), V(9.5, ty+6.8, 1))) })
	g.With(Slate, func() {
		for i := 0; i < 3; i++ {
			for j := 0; j < 2; j++ {
				g.Fill(Cylinder{Center: P(1.8+float64(i)*3, -2.8+float64(j)*2.8), Radius: 1.1, Y0: ty + 6.8, Y1: ty + 7.4})
			}
		}
	})
	g.Paint(Gold, Func{Box{V(-1.5, ty-0.3, 2.5), V(10.5, ty, 13.5)}, func(p Vec3) bool {
		return p.X < -0.9 || p.X > 9.9 || p.Z < 3.1 || p.Z > 12.9
	}})
}

// quarters is the accommodation block: six white storeys with rows of dark windows,
// lifeboats on davits along the +Z face.
func (r OilRig) quarters(g *Grid) {
	x0, x1, z0, z1 := 11.0, 24.6, -3.5, 15.6
	y0 := oilRigMain + oilRigPlate
	storey := 2.9
	y1 := y0 + 6*storey
	g.With(White, func() { g.Fill(Block(V(x0, y0, z0), V(x1, y1, z1))) })
	g.Carve(Block(V(x0+1.1, y0+0.5, z0+1.1), V(x1-1.1, y1-1.1, z1-1.1)))
	box := Box{V(x0, y0, z0), V(x1, y1, z1)}
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		s := (p.Y - y0) / storey
		f := s - math.Floor(s)
		if s >= 6 || f < 0.38 || f > 0.75 {
			return false
		}
		dx := math.Min(math.Abs(p.X-x0), math.Abs(p.X-x1))
		dz := math.Min(math.Abs(p.Z-z0), math.Abs(p.Z-z1))
		if math.Max(dx, dz) < 1.2 {
			return false // Keep the corners solid.
		}
		along := p.X
		if dx < dz {
			along = p.Z
		}
		w := math.Mod(along+100, 2.4)
		return w > 0.6 && w < 2.0
	}})
	g.Paint(Red, Block(V(x0, y1-0.6, z0), V(x1, y1, z1)))
	for _, x := range []float64{13.5, 17.5, 21.5} {
		zc := z1 + 1.3
		yc := y0 + 2.6
		g.With(Gold, func() { g.Fill(Capsule{V(x-1.4, yc, zc), V(x+1.4, yc, zc), 1.0}) })
		g.Paint(Red, Block(V(x-2.5, yc-1.2, zc-1.2), V(x+2.5, yc-0.2, zc+1.2)))
		g.Paint(White, Block(V(x-0.3, yc+0.3, zc-1.2), V(x+0.3, yc+1.1, zc+1.2)))
		g.With(Charcoal, func() {
			for _, dx := range []float64{-1.3, 1.3} {
				g.Fill(Capsule{V(x+dx, yc+2.6, z1), V(x+dx, yc+2.6, zc), 0.4})
				g.Fill(Capsule{V(x+dx, yc+2.6, zc), V(x+dx, yc+1.0, zc), 0.35})
			}
		})
	}
}

// helideck is the octagonal landing deck with its markings and support struts.
func (r OilRig) helideck(g *Grid, c Vec2, y float64) {
	const R = 10.2
	poly := make([]Vec2, 8)
	for k := range poly {
		poly[k] = Polar(c, R, math.Pi/8+float64(k)*math.Pi/4)
	}
	g.With(Navy, func() { g.Fill(Prism{poly, y, y + 0.8}) })
	apothem := func(p Vec3) float64 {
		d := p.Plan().Sub(c)
		m := 0.0
		for k := 0; k < 8; k++ {
			m = math.Max(m, d.Dot(Polar(P(0, 0), 1, float64(k)*math.Pi/4)))
		}
		return m
	}
	b := Box{V(c.X-R, y-1, c.Y-R), V(c.X+R, y+1, c.Y+R)}
	g.Paint(White, Func{b, func(p Vec3) bool { a := apothem(p); return a > 8.0 && a < 8.8 }})
	g.Paint(Gold, Func{b, func(p Vec3) bool {
		d := p.Plan().Sub(c).Length()
		return d > 3.7 && d < 4.9
	}})
	g.Paint(White, Func{b, func(p Vec3) bool {
		d := p.Plan().Sub(c)
		if math.Abs(d.Y) > 2.5 {
			return false
		}
		return (math.Abs(d.X) > 0.9 && math.Abs(d.X) < 2.0) || (math.Abs(d.X) <= 0.9 && math.Abs(d.Y) < 0.5)
	}})
	// Perimeter safety net shelf just below the deck edge.
	g.With(Charcoal, func() {
		g.Fill(Func{b, func(p Vec3) bool {
			a := apothem(p)
			return p.Y >= y-0.5 && p.Y <= y && a < R*math.Cos(math.Pi/8)+0.9 && !InsidePolygon(poly, p.Plan())
		}})
	})
	// Struts from the accommodation walls out to the cantilevered edge.
	g.With(Slate, func() {
		for _, z := range []float64{3, 10, 15} {
			g.Fill(Capsule{V(24.4, 32, z), V(c.X+8, y, z), 0.5})
		}
		for _, x := range []float64{14, 19, 24} {
			g.Fill(Capsule{V(x, 32, 15.4), V(x, y, c.Y+8), 0.5})
		}
		g.Fill(Cylinder{Center: P(c.X-4, c.Y-4), Radius: 0.6, Y0: y - 1.2, Y1: y})
	})
}

// derrick is the drilling substructure, drill floor and lattice derrick.
func (r OilRig) derrick(g *Grid, c Vec2) {
	ty := oilRigTop + oilRigPlate
	floor := ty + 5.5
	r.module(g, c.X-6, c.Y-6, c.X+6, c.Y+6, ty, floor, Red, White, 11)
	g.With(Charcoal, func() { g.Fill(BlockAt(c, 13.4, 13.4, floor, floor+0.8)) })
	fy := floor + 0.8
	top := fy + 30.0
	r.truss(g, c.At(fy), c.At(top), V(1, 0, 0), 4.6, 4.6, 1.5, 1.5, 6, 0.6, 0.45, true, White, White)
	// Red windwall cladding on the lower derrick, open on the V-door (-X) side.
	w := func(y float64) float64 { return 4.6 + (1.5-4.6)*(y-fy)/(top-fy) }
	g.With(Red, func() {
		g.Fill(Func{Box{V(c.X-5.5, fy, c.Y-5.5), V(c.X+5.5, fy+9, c.Y+5.5)}, func(p Vec3) bool {
			dx, dz := p.X-c.X, p.Z-c.Y
			m := math.Max(math.Abs(dx), math.Abs(dz))
			ww := w(p.Y)
			if m > ww+0.3 || m < ww-0.6 {
				return false
			}
			return !(dx < -ww+0.7 && math.Abs(dz) < ww-0.8)
		}})
	})
	// Crown block, monkey board, travelling block and the doghouse.
	g.With(Red, func() { g.Fill(BlockAt(c, 3.6, 3.6, top, top+1.6)) })
	g.With(Gold, func() {
		g.Fill(Sphere(c.At(top+2.3), 0.7))
		g.Fill(Block(V(c.X-w(fy+19)-1.8, fy+19, c.Y-1.4), V(c.X-w(fy+19)+0.3, fy+19.6, c.Y+1.4)))
		g.Fill(BlockAt(c, 1.6, 1.6, fy+11, fy+13.5))
	})
	g.With(Charcoal, func() {
		g.Fill(Capsule{c.At(fy + 13.5), c.At(top), 0.35})
		g.Fill(Capsule{c.At(fy), c.At(fy + 11), 0.4})
	})
	g.With(White, func() { g.Fill(Block(V(c.X+2.5, fy, c.Y+3.2), V(c.X+6, fy+3, c.Y+6.2))) })
	g.Paint(Charcoal, Block(V(c.X+2.5, fy+1.4, c.Y+6.0), V(c.X+6, fy+2.3, c.Y+6.2)))
	// Catwalk ramp from the V-door down to the pipe deck.
	g.With(Gold, func() { g.Fill(Capsule{V(c.X-6, fy-0.2, c.Y), V(-21, ty+0.5, c.Y), 0.7}) })
	// Stands of drill pipe racked inside the derrick.
	g.With(Slate, func() { g.Fill(Block(V(c.X+1.0, fy, c.Y-2.5), V(c.X+2.4, fy+13, c.Y+0.5))) })
}

// crane is a pedestal crane with an A-frame and a lattice boom; it returns the tip.
func (r OilRig) crane(g *Grid, base Vec2, y0, height, heading, elevation, length float64, cable float64) Vec3 {
	dir := Polar(P(0, 0), 1, heading)
	top := y0 + height
	g.With(Gold, func() {
		g.Fill(Cylinder{Center: base, Radius: 1.4, Y0: y0, Y1: top})
		g.Fill(Cylinder{Center: base, Radius: 2.2, Y0: top, Y1: top + 0.8})
	})
	side := P(-dir.Y, dir.X)
	cab := base.Add(side.Scale(1.6)).Add(dir.Scale(0.6))
	g.With(White, func() {
		g.Fill(RotateY(BlockAt(cab, 2.6, 1.8, top+0.8, top+3.2), cab, heading))
	})
	win := cab.Add(dir.Scale(1.2))
	g.Paint(Charcoal, Block(V(win.X-0.7, top+1.9, win.Y-0.7), V(win.X+0.7, top+3.0, win.Y+0.7)))
	back := base.Sub(dir.Scale(1.5))
	apex := back.At(top + 5.5)
	g.With(Gold, func() {
		g.Fill(Capsule{back.Add(side.Scale(1.2)).At(top + 0.8), apex, 0.5})
		g.Fill(Capsule{back.Sub(side.Scale(1.2)).At(top + 0.8), apex, 0.5})
	})
	foot := base.Add(dir.Scale(1.2)).At(top + 1.6)
	reach := V(dir.X*math.Cos(elevation), math.Sin(elevation), dir.Y*math.Cos(elevation))
	tip := foot.Add(reach.Scale(length))
	n := int(length / 2.4)
	r.truss(g, foot, tip, V(0, 1, 0), 0.9, 1.1, 0.5, 0.5, n, 0.5, 0.45, false, Gold, Gold)
	g.Paint(Red, Func{Box{tip.Sub(V(4, 4, 4)), tip.Add(V(4, 4, 4))}, func(p Vec3) bool {
		return p.Sub(tip).Length() < 3.0
	}})
	g.With(Red, func() { g.Fill(Sphere(tip, 0.8)) })
	g.With(Charcoal, func() {
		g.Fill(Capsule{apex, tip.Sub(reach.Scale(1.5)), 0.4})
		g.Fill(Capsule{tip, V(tip.X, cable, tip.Z), 0.4})
	})
	g.With(Gold, func() { g.Fill(BlockAt(tip.Plan(), 1.0, 1.0, cable-0.8, cable+0.2)) })
	return tip
}

// flare is the raked lattice flare boom off the -X end with its flame and smoke.
func (r OilRig) flare(g *Grid) {
	base := V(-oilRigHX+0.5, oilRigTop+0.8, 7)
	dir := oilRigNorm(V(-1, 0, 0.35))
	el := 32 * math.Pi / 180
	L := 28.0
	reach := V(dir.X*math.Cos(el), math.Sin(el), dir.Z*math.Cos(el))
	end := base.Add(reach.Scale(L))
	r.truss(g, base, end, V(0, 1, 0), 1.5, 1.6, 0.8, 0.8, 11, 0.55, 0.45, false, White, White)
	g.Paint(Red, Func{Box{end.Sub(V(6, 6, 6)), end.Add(V(6, 6, 6))}, func(p Vec3) bool {
		d := p.Sub(end).Length()
		return d < 5 && int(math.Floor(d/1.6))%2 == 0
	}})
	// Supporting strut from the main deck.
	g.With(White, func() {
		g.Fill(Capsule{V(-oilRigHX+0.5, oilRigMain+0.8, 7), base.Add(reach.Scale(L * 0.35)), 0.55})
	})
	// Flare tip and flame, leaning away with the wind.
	g.With(Charcoal, func() { g.Fill(Capsule{end, end.Add(V(0, 2.2, 0)), 0.75}) })
	f := end.Add(V(0, 2.4, 0))
	H := 13.0
	lean := P(dir.X, dir.Z)
	center := func(t float64) Vec2 { return f.Plan().Add(lean.Scale(4.0 * math.Pow(t, 1.5))) }
	radius := func(t float64) float64 { return 3.4 * math.Sin(math.Pi*math.Pow(t, 0.6)) }
	fb := Box{V(f.X-7, f.Y, f.Z-7), V(f.X+7, f.Y+H, f.Z+7)}
	inFlame := func(p Vec3) (float64, bool) {
		t := (p.Y - f.Y) / H
		if t < 0 || t > 1 {
			return 0, false
		}
		ang := math.Atan2(p.Z-f.Z, p.X-f.X)
		wob := 1 + 0.18*math.Sin(3*ang+7*t) + 0.12*math.Sin(5*ang-11*t)
		return t, p.Plan().Sub(center(t)).Length() <= radius(t)*wob
	}
	g.With(Gold, func() {
		g.Fill(Func{fb, func(p Vec3) bool { _, ok := inFlame(p); return ok }})
	})
	g.Paint(White, Func{fb, func(p Vec3) bool { t, ok := inFlame(p); return ok && t < 0.22 }})
	g.Paint(Red, Func{fb, func(p Vec3) bool {
		t, ok := inFlame(p)
		c := r.cell(g, p)
		return ok && t > 0.5+0.2*Hash01(int(c.X), int(c.Y), int(c.Z), 41)
	}})
}

// wellhead is the small bridge-linked wellhead platform and its bridge.
func (r OilRig) wellhead(g *Grid) {
	c := P(45, -8)
	r.jacket(g, c, 5, 5, 7, 7, 2, 1.0)
	r.deck(g, c.X-6.5, c.Y-6.5, c.X+6.5, c.Y+6.5, oilRigCellar, true)
	r.deck(g, c.X-6.5, c.Y-6.5, c.X+6.5, c.Y+6.5, oilRigMain, true)
	g.With(White, func() {
		for _, dx := range []float64{-5, 5} {
			for _, dz := range []float64{-5, 5} {
				g.Fill(Cylinder{Center: c.Add(P(dx, dz)), Radius: 0.8, Y0: oilRigCellar, Y1: oilRigMain})
			}
		}
	})
	// Christmas trees on the cellar deck, conductors below.
	g.With(Red, func() {
		for i := 0; i < 3; i++ {
			for j := 0; j < 2; j++ {
				q := c.Add(P(-2+float64(i)*2, -1+float64(j)*2.5))
				g.Fill(Cylinder{Center: q, Radius: 0.55, Y0: oilRigCellar + 0.8, Y1: oilRigCellar + 3.2})
			}
		}
	})
	g.With(Slate, func() {
		for i := 0; i < 3; i++ {
			for j := 0; j < 2; j++ {
				q := c.Add(P(-2+float64(i)*2, -1+float64(j)*2.5))
				g.Fill(Cylinder{Center: q, Radius: 0.55, Y0: -0.3, Y1: oilRigCellar})
			}
		}
	})
	ty := oilRigMain + oilRigPlate
	r.module(g, c.X-6.5, c.Y-6.5, c.X+1, c.Y-1, ty, ty+4, White, Red, 21)
	// Vent stack with red and white bands and a beacon.
	vs := c.Add(P(4, -4))
	g.With(Slate, func() { g.Fill(Cylinder{Center: vs, Radius: 0.8, Y0: ty, Y1: 40}) })
	for i := 0; i < 3; i++ {
		y := 33.0 + float64(i)*2.4
		g.Paint(Red, Block(V(vs.X-1, y, vs.Y-1), V(vs.X+1, y+1.2, vs.Y+1)))
		g.Paint(White, Block(V(vs.X-1, y+1.2, vs.Y-1), V(vs.X+1, y+2.4, vs.Y+1)))
	}
	g.With(Gold, func() { g.Fill(Sphere(vs.At(40.6), 0.8)) })
	r.crane(g, c.Add(P(4, 4)), ty, 3, math.Pi/2-0.3, 50*math.Pi/180, 11, 26)
	// Bridge from the main deck: walkway with two Warren side trusses.
	x0, x1, z := oilRigHX, c.X-6.5, c.Y
	g.With(Charcoal, func() { g.Fill(Block(V(x0, oilRigMain, z-1.4), V(x1, oilRigMain+oilRigPlate, z+1.4))) })
	g.With(Gold, func() {
		for _, s := range []float64{-1.4, 1.4} {
			lo, hi := oilRigMain+0.4, oilRigMain+3.6
			g.Fill(Capsule{V(x0, lo, z+s), V(x1, lo, z+s), 0.45})
			g.Fill(Capsule{V(x0, hi, z+s), V(x1, hi, z+s), 0.45})
			n := 7
			for i := 0; i < n; i++ {
				xa := x0 + (x1-x0)*float64(i)/float64(n)
				xb := x0 + (x1-x0)*float64(i+1)/float64(n)
				if i%2 == 0 {
					g.Fill(Capsule{V(xa, lo, z+s), V(xb, hi, z+s), 0.45})
				} else {
					g.Fill(Capsule{V(xa, hi, z+s), V(xb, lo, z+s), 0.45})
				}
			}
		}
	})
}

// conductors is the bundle of well conductors under the derrick and riser pipes.
func (r OilRig) conductors(g *Grid, c Vec2) {
	g.With(Slate, func() {
		for i := 0; i < 4; i++ {
			for j := 0; j < 3; j++ {
				q := c.Add(P(-3.3+float64(i)*2.2, -2.2+float64(j)*2.2))
				g.Fill(Cylinder{Center: q, Radius: 0.6, Y0: -0.3, Y1: oilRigCellar})
			}
		}
	})
	// Conductor guide frames.
	g.With(Gold, func() {
		for _, y := range []float64{5, 10} {
			g.Fill(BlockAt(c, 9.6, 7.2, y, y+0.7))
		}
	})
	g.Carve(BlockAt(c, 7.6, 5.2, 4.4, 11.3))
	g.With(Slate, func() {
		for i := 0; i < 4; i++ {
			for j := 0; j < 3; j++ {
				q := c.Add(P(-3.3+float64(i)*2.2, -2.2+float64(j)*2.2))
				g.Fill(Cylinder{Center: q, Radius: 0.6, Y0: 4.4, Y1: 11.4})
			}
		}
	})
	// Risers and a caisson clamped to the legs.
	g.With(Charcoal, func() {
		for _, q := range []Vec2{P(21.1, 8.5), P(21.1, 6.5), P(-21.3, -7)} {
			g.Fill(Capsule{V(q.X, -0.3, q.Y), V(q.X*17.6/20.5, oilRigCellar, q.Y), 0.5})
		}
	})
}

// vesselHalf is the supply vessel's half-beam at (x, y): a plumb-ish raked bow and a
// square transom; negative outside.
func (OilRig) vesselHalf(x, y float64) float64 {
	stern, bow := -16.5, 14.5+0.45*math.Max(y, 0)
	if x < stern || x > bow {
		return -1
	}
	w := 4.2
	if x > 6 {
		t := (x - 6) / (bow - 6)
		w *= math.Sqrt(math.Max(1-t*t, 0))
	}
	return w
}

func (OilRig) vesselDeck(x float64) float64 {
	if x > 5 {
		return 5.4
	}
	return 3.4
}

func (r OilRig) vessel(g *Grid) {
	zc := oilRigShipZ
	b := Box{V(-17, -0.5, zc-5), V(22, 7, zc+5)}
	g.With(Red, func() {
		g.Fill(Func{b, func(p Vec3) bool {
			w := r.vesselHalf(p.X, p.Y)
			return w >= 0 && p.Y >= -0.5 && p.Y <= r.vesselDeck(p.X) && math.Abs(p.Z-zc) <= w
		}})
	})
	g.Carve(Func{b, func(p Vec3) bool {
		w := r.vesselHalf(p.X, p.Y)
		return w >= 0 && p.X > -15.4 && p.Y < r.vesselDeck(p.X)-0.9 && math.Abs(p.Z-zc) <= w-1.1
	}})
	hull := func(p Vec3) bool {
		w := r.vesselHalf(p.X, p.Y)
		return w >= 0 && math.Abs(p.Z-zc) <= w+0.3
	}
	g.Paint(Charcoal, Func{b, func(p Vec3) bool { return p.Y < 1.0 && hull(p) }})
	g.Paint(White, Func{b, func(p Vec3) bool { return p.Y > 2.4 && p.Y < 2.9 && hull(p) }})
	g.Paint(Slate, Func{b, func(p Vec3) bool {
		w := r.vesselHalf(p.X, p.Y)
		return p.Y > r.vesselDeck(p.X)-0.5 && math.Abs(p.Z-zc) < w-0.9 && p.X > -16
	}})
	// Bulwarks around the work deck, a stern roller and the A-frame.
	g.With(Red, func() {
		g.Fill(Func{Box{V(-17, 3.4, zc-5), V(6, 4.6, zc+5)}, func(p Vec3) bool {
			w := r.vesselHalf(p.X, 3.4)
			d := math.Abs(p.Z - zc)
			return w >= 0 && d <= w && d > w-0.7
		}})
	})
	g.With(Gold, func() {
		g.Fill(Capsule{V(-16.2, 3.6, zc-1.8), V(-16.2, 3.6, zc+1.8), 0.6})
		for _, s := range []float64{-3.2, 3.2} {
			g.Fill(Capsule{V(-15.8, 3.8, zc+s), V(-15, 9.5, zc+s*0.7), 0.5})
		}
		g.Fill(Capsule{V(-15, 9.5, zc-2.3), V(-15, 9.5, zc+2.3), 0.5})
	})
	// Superstructure forward: two decks of cabins and a wheelhouse with bridge wings.
	g.With(White, func() {
		g.Fill(Block(V(6, 5.4, zc-3.2), V(12.5, 8.2, zc+3.2)))
		g.Fill(Block(V(6.5, 8.2, zc-3.0), V(12.2, 10.8, zc+3.0)))
		g.Fill(Block(V(8.5, 10.8, zc-4.2), V(12.5, 13.2, zc+4.2)))
	})
	g.Paint(Charcoal, Func{Box{V(6, 6.4, zc-4.3), V(12.6, 7.6, zc+4.3)}, func(p Vec3) bool {
		return math.Mod(p.X+100, 1.6) < 0.9 || p.X > 12
	}})
	g.Paint(Charcoal, Func{Box{V(6.5, 9.1, zc-4.3), V(12.3, 10.2, zc+4.3)}, func(p Vec3) bool {
		return math.Mod(p.X+100, 1.6) < 0.9 || p.X > 11.8
	}})
	g.Paint(Charcoal, Block(V(8.5, 11.8, zc-4.2), V(12.5, 12.8, zc+4.2)))
	g.Paint(Red, Block(V(8.5, 12.8, zc-4.2), V(12.5, 13.2, zc+4.2)))
	g.With(Gold, func() {
		g.Fill(Capsule{V(10, 13.2, zc), V(10, 17.2, zc), 0.45})
		g.Fill(Capsule{V(10, 16, zc-1.6), V(10, 16, zc+1.6), 0.4})
	})
	g.With(White, func() {
		for _, s := range []float64{-2.2, 2.2} {
			g.Fill(Cylinder{Center: P(5.0, zc+s), Radius: 0.7, Y0: 3.4, Y1: 9})
		}
	})
	g.Paint(Charcoal, Block(V(4, 8.3, zc-3), V(6, 9, zc+3)))
	// Deck cargo: containers, a pipe bundle and a skid.
	r.container(g, V(-11, 3.4, zc-2.0), true, Navy)
	r.container(g, V(-11, 3.4, zc+0.0), true, White)
	r.container(g, V(-11, 5.1, zc-1.0), true, Red)
	g.With(Slate, func() {
		for i := 0; i < 3; i++ {
			for j := 0; j < 2; j++ {
				z := zc + 0.8 + float64(i)*0.9
				g.Fill(Capsule{V(-3, 3.9+float64(j)*0.8, z), V(4, 3.9+float64(j)*0.8, z), 0.45})
			}
		}
	})
	g.With(Gold, func() { g.Fill(Block(V(-2, 3.4, zc-3.0), V(2, 4.8, zc-1.2))) })
	r.container(g, V(-6.5, 3.4, zc-2.0), true, Gold)
	r.container(g, V(-6.5, 3.4, zc+2.0), true, Red)
	// Knuckle-boom deck crane on the port side.
	g.With(Gold, func() {
		g.Fill(Cylinder{Center: P(3.5, zc-3), Radius: 0.8, Y0: 3.4, Y1: 6.5})
		g.Fill(Capsule{V(3.5, 6.5, zc-3), V(0, 9.5, zc-3), 0.5})
		g.Fill(Capsule{V(0, 9.5, zc-3), V(-4, 6.5, zc-3), 0.45})
	})
}

// foam paints whitewater on the sea around legs, conductors and the vessel.
func (r OilRig) foam(g *Grid) {
	legs := []Vec2{P(20.5, 13.5), P(-20.5, 13.5), P(20.5, -13.5), P(-20.5, -13.5),
		P(52, -1), P(38, -1), P(52, -15), P(38, -15), P(-7, -1), P(45, -8)}
	g.Paint(White, Func{oilRigSea, func(p Vec3) bool {
		col, _ := g.ColorOf(r.cell(g, p))
		if col != Water && col != Navy {
			return false
		}
		c := r.cell(g, p)
		h := Hash01(int(c.X), int(c.Z), 51)
		for i, l := range legs {
			rad := 3.0
			if i >= 8 {
				rad = 5.2
			}
			if p.Plan().Sub(l).Length() < rad {
				return h < 0.55
			}
		}
		// Around the hull and churned water astern from the thrusters.
		dz := math.Abs(p.Z - oilRigShipZ)
		w := r.vesselHalf(p.X, 0)
		if w >= 0 && dz < w+1.5 {
			return h < 0.6
		}
		if p.X < -16 && p.X > -23 && dz < 3.5-(-16-p.X)*0.2 {
			return h < 0.7
		}
		return false
	}})
}

func (r OilRig) Build(g *Grid) Order {
	r.sea(g)
	r.jacket(g, P(0, 0), 17, 10.5, 20.5, 13.5, 3, 1.3)
	r.conductors(g, P(-7, -1))
	r.topside(g)
	r.quarters(g)
	r.helideck(g, P(24, 10.5), 21.3+6*2.9+1.0)
	r.derrick(g, P(-7, -1))
	ty := oilRigTop + oilRigPlate
	// Crane 1 lifts a container off the supply vessel; crane 2 works the -Z side.
	h1 := math.Atan2(26-11.5, -8+20)
	tip := r.crane(g, P(-20, 11.5), ty, 6, h1, 43*math.Pi/180, 24, 14.5)
	r.container(g, V(tip.X, 11.8, tip.Z), true, Gold)
	r.crane(g, P(17, -11), ty, 4.5, math.Atan2(-1, 0.45), 55*math.Pi/180, 20, 26)
	r.flare(g)
	r.wellhead(g)
	r.vessel(g)
	// Boat landing and access ladder on the (+X, +Z) leg.
	g.With(Gold, func() {
		for _, dx := range []float64{-1.4, 1.4} {
			g.Fill(Capsule{V(19.5+dx, 0.2, 15.6), V(19.5+dx, 3.4, 15.6), 0.5})
		}
		for _, y := range []float64{0.6, 2.0, 3.4} {
			g.Fill(Capsule{V(18.1, y, 15.6), V(20.9, y, 15.6), 0.45})
			g.Fill(Capsule{V(19.5, y, 14.2), V(19.5, y, 15.6), 0.45})
		}
		g.Fill(Capsule{V(19.2, 3.4, 15.0), V(17.2, oilRigCellar, 11.8), 0.5})
	})
	// Trim jacket members that dip below the sea surface layer.
	g.Carve(Block(V(-100, -5, -100), V(100, -0.3, 100)))
	r.foam(g)
	return Sweep(P(0, 0), 0)
}
