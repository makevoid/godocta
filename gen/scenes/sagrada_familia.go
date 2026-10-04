package scenes

import (
	"math"

	"godocta/gen/kit"
	. "godocta/gen/octa"
)

// SagradaFamilia is Gaudí's basilica in Barcelona as completed, at roughly two metres
// per unit. The nave runs along X from the Glory façade (-X) to the apse (+X); the
// transept carries the Nativity façade (+Z, facing the pond of Plaça de Gaudí) and the
// Passion façade (-Z). Each façade has four perforated parabolic bell towers with
// mosaic finials. Around the crossing stand the four Evangelist towers, the Virgin Mary
// tower with its star rises over the apse, and the central Jesus tower with its cross
// tops them all. The nave carries pinnacles with fruit baskets; the apse a ring of
// spires. The basilica fills a chamfered Eixample block with plane trees.
type SagradaFamilia struct{}

func init() { Register(SagradaFamilia{}) }

func (SagradaFamilia) Name() string  { return "sagrada_familia" }
func (SagradaFamilia) Title() string { return "Sagrada Familia" }

// noise is smooth deterministic value noise in [0, 1).
func (SagradaFamilia) noise(x, y float64, seed int) float64 {
	ix, iy := math.Floor(x), math.Floor(y)
	fx, fy := x-ix, y-iy
	fx, fy = fx*fx*(3-2*fx), fy*fy*(3-2*fy)
	h := func(a, b float64) float64 { return Hash01(int(a), int(b), seed) }
	a := h(ix, iy) + (h(ix+1, iy)-h(ix, iy))*fx
	b := h(ix, iy+1) + (h(ix+1, iy+1)-h(ix, iy+1))*fx
	return a + (b-a)*fy
}

// mosaic paints the occupied cells of a solid in a trencadís of the given colors,
// chosen per coarse patch.
func (SagradaFamilia) mosaic(g *Grid, s Solid, patch float64, seed int, colors ...Color) {
	b := s.Bounds()
	for i := range colors {
		i := i
		g.Paint(colors[i], Func{b, func(p Vec3) bool {
			if !s.Contains(p) {
				return false
			}
			h := Hash01(int(math.Floor(p.X/patch)), int(math.Floor(p.Y/patch)), int(math.Floor(p.Z/patch)), seed)
			return int(h*float64(len(colors))) == i
		}})
	}
}

// finial is a bell tower's mosaic pinnacle standing on y0: a red stem, a swollen gold
// and red mitre studded with white bobbles, and a cross of white balls.
func (s SagradaFamilia) finial(g *Grid, c Vec2, y0, k float64, seed int) {
	g.With(Red, func() { g.FillVolume(Frustum{c, 1.1 * k, 0.9 * k, y0 - 0.5, y0 + 2*k}) })
	body := Ellipsoid{c.At(y0 + 4.2*k), V(1.75*k, 2.7*k, 1.75*k)}
	g.With(Gold, func() { g.FillVolume(body) })
	s.mosaic(g, Ellipsoid{body.Center, body.Radii.Add(V(0.5, 0.5, 0.5))}, 0.9, seed, Gold, Gold, Red, Red, Grass)
	g.With(White, func() {
		for i := 0; i < 4; i++ {
			a := float64(i)/4*2*math.Pi + float64(seed)
			g.Fill(Sphere(Polar(c, 1.75*k, a).At(y0+3.6*k), 0.7))
		}
	})
	g.With(Gold, func() { g.Fill(Capsule{c.At(y0 + 6.5*k), c.At(y0 + 9.6*k), 0.6}) })
	arm := 1.5 * k
	y := y0 + 8.2*k
	g.With(Red, func() {
		g.Fill(Capsule{c.Add(P(-arm, 0)).At(y), c.Add(P(arm, 0)).At(y), 0.55})
		g.Fill(Capsule{c.Add(P(0, -arm)).At(y), c.Add(P(0, arm)).At(y), 0.55})
	})
	g.With(White, func() {
		for _, d := range []Vec2{P(1, 0), P(-1, 0), P(0, 1), P(0, -1)} {
			g.Fill(Sphere(c.Add(d.Scale(arm)).At(y), 0.75))
		}
		g.Fill(Sphere(c.At(y0+10*k), 0.8))
	})
}

// bellTower is one of the twelve apostle towers: a hollow cylindrical shaft to `shaft`,
// then a parabolic taper perforated by helical louvres, a Sanctus band, and a finial
// whose base stands at `top`.
func (s SagradaFamilia) bellTower(g *Grid, c Vec2, shaft, top, r0 float64, seed int, body, band Color) {
	const wall = 1.1
	tf := 1 - math.Pow(0.95/r0, 1/0.75)
	apex := shaft + (top-shaft)/tf
	radius := func(y float64) float64 {
		if y <= shaft {
			return r0
		}
		return r0 * math.Pow(math.Max(0, 1-(y-shaft)/(apex-shaft)), 0.75)
	}
	box := Box{V(c.X-r0-1, 0, c.Y-r0-1), V(c.X+r0+1, top+0.5, c.Y+r0+1)}
	dist := func(p Vec3) float64 { return p.Plan().Sub(c).Length() }
	g.Use(body)
	g.FillVolume(Func{box, func(p Vec3) bool {
		r, d := radius(p.Y), dist(p)
		return d <= r && d > r-wall
	}})
	// Tapered courses alternate stone and sand.
	g.Paint(band, Func{box, func(p Vec3) bool { return p.Y > shaft && math.Mod(p.Y-shaft, 6) < 1.2 }})
	// Helical louvres.
	g.Carve(Func{box, func(p Vec3) bool {
		r := radius(p.Y)
		t := (p.Y - shaft) / (top - shaft)
		if t < 0.04 || t > 0.72 || r < 1.7 {
			return false
		}
		a := p.Plan().Sub(c).Angle()
		u := a/(2*math.Pi)*6 + (p.Y-shaft)/14 + float64(seed)*0.37
		u -= math.Floor(u)
		return u < 0.4 && math.Mod(p.Y-shaft, 2.4) < 1.3
	}})
	// Tall slits in the shaft.
	if shaft > 4 {
		g.Carve(Func{box, func(p Vec3) bool {
			if p.Y < 3 || p.Y > shaft-1.5 {
				return false
			}
			a := p.Plan().Sub(c).Angle()
			u := a/(2*math.Pi)*8 + 0.25
			u -= math.Floor(u)
			return u < 0.22 && math.Mod(p.Y, 5) > 1
		}})
	}
	// Dark inner face behind the openings.
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		r := radius(p.Y)
		return r > 2 && dist(p) < r-wall+0.55
	}})
	// Sanctus band with its apostle.
	yb := shaft + (top-shaft)*0.62
	g.With(Sand, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			r, d := radius(p.Y), dist(p)
			return p.Y >= yb && p.Y <= yb+1 && d > r-0.6 && d <= r+0.6
		}})
	})
	g.With(Gold, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			r, d := radius(p.Y), dist(p)
			a := p.Plan().Sub(c).Angle()
			return p.Y >= yb+1.2 && p.Y <= yb+2.2 && d > r-0.3 && d <= r+0.5 && math.Mod(a*6/math.Pi+12, 1) < 0.5
		}})
	})
	s.finial(g, c, top, math.Min(1, r0/2.6), seed)
}

// centralTower is an Evangelist, Mary or Jesus tower: a ribbed parabolic shell with
// window slots, glazed white above whiteFrom (fraction of height). It returns the
// height of its crown base.
func (SagradaFamilia) centralTower(g *Grid, c Vec2, base, apex, r0 float64, ribs int, whiteFrom, endR float64) float64 {
	const wall = 1.2
	radius := func(y float64) float64 { return r0 * math.Pow(math.Max(0, 1-(y-base)/(apex-base)), 0.55) }
	end := base + (apex-base)*(1-math.Pow(endR/r0, 1/0.55))
	box := Box{V(c.X-r0-1, base, c.Y-r0-1), V(c.X+r0+1, end+0.5, c.Y+r0+1)}
	sector := 2 * math.Pi / float64(ribs)
	polar := func(p Vec3) (d, off, r, t float64) {
		v := p.Plan().Sub(c)
		a := math.Mod(v.Angle()+sector/2+20*math.Pi, sector) - sector/2
		r = radius(p.Y)
		return v.Length(), math.Abs(a) * r, r, (p.Y - base) / (end - base)
	}
	rib := func(off float64) bool { return off < 0.55 }
	g.Use(Stone)
	g.FillVolume(Func{box, func(p Vec3) bool {
		d, off, r, _ := polar(p)
		out := r
		if rib(off) {
			out += 0.8
		}
		return d <= out && d > r-wall
	}})
	window := func(p Vec3) bool {
		_, off, r, t := polar(p)
		mid := math.Abs(off - sector*r/2)
		return t < 0.9 && r > 2.2 && mid < sector*r*0.2 && math.Mod(p.Y-base, 4.5) > 1.2 && math.Mod(p.Y-base, 4.5) < 4
	}
	g.Paint(White, Func{box, func(p Vec3) bool { _, _, _, t := polar(p); return t > whiteFrom }})
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		_, off, _, t := polar(p)
		return t <= whiteFrom && rib(off)
	}})
	g.Carve(Func{box, func(p Vec3) bool {
		d, _, r, _ := polar(p)
		return d > r-wall-0.3 && window(p)
	}})
	g.With(Glass, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			d, _, r, t := polar(p)
			return t > whiteFrom && d > r-wall && d <= r-wall+0.6 && window(p)
		}})
	})
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		d, _, r, t := polar(p)
		return t <= whiteFrom && r > 2.4 && d < r-wall+0.55
	}})
	return end
}

// pinnacle is a slender cone with a fruit basket of colored mosaic fruit.
func (s SagradaFamilia) pinnacle(g *Grid, c Vec2, y0, h float64, seed int) {
	g.With(Stone, func() { g.FillVolume(Frustum{c, 1.0, 0.45, y0, y0 + h}) })
	y := y0 + h
	g.With(Gold, func() { g.Fill(Frustum{c, 0.6, 1.3, y, y + 0.9}) })
	fruits := []Color{Red, Gold, Terracotta, Grass, Navy, White, Red, Terracotta}
	for i := 0; i < 4; i++ {
		col := fruits[int(Hash01(seed, i)*float64(len(fruits)))]
		a := float64(i)*math.Pi/2 + Hash01(seed)*math.Pi
		g.With(col, func() { g.Fill(Sphere(Polar(c, 0.7, a).At(y+1.5), 0.75)) })
	}
	col := fruits[int(Hash01(seed, 9)*float64(len(fruits)))]
	g.With(col, func() { g.Fill(Sphere(c.At(y+2.5), 0.8)) })
}

func (s SagradaFamilia) Build(g *Grid) Order {
	const (
		xGlory  = -34.0 // Nave west end.
		xT      = 8.0   // Transept half-width (X).
		zT      = 16.5  // Transept ends.
		zN      = 11.5  // Nave outer half-width.
		zC      = 4.5   // Central vessel half-width.
		hAisle  = 15.0
		hNave   = 23.0
		hCross  = 28.0
		xApse   = 12.0 // Apse semicircle centre.
		wall    = 1.2
		zFacade = 16.5
	)

	// ------------------------------------------------------------------ Site.
	// Chamfered Eixample blocks, streets and the Plaça de Gaudí with its pond.
	const gx0, gx1, gz0, gz1 = -53.0, 37.0, -35.0, 57.0
	chamfer := func(x0, z0, x1, z1, k float64) []Vec2 {
		return []Vec2{P(x0+k, z0), P(x1-k, z0), P(x1, z0+k), P(x1, z1-k), P(x1-k, z1), P(x0+k, z1), P(x0, z1-k), P(x0, z0+k)}
	}
	g.Use(Charcoal).Fill(Block(V(gx0, -0.2, gz0), V(gx1, 0.2, gz1)))
	// Zebra crossings and lane marks.
	g.Paint(White, Func{Box{V(gx0, -0.3, gz0), V(gx1, 0.3, gz1)}, func(p Vec3) bool {
		return (math.Abs(p.Z-31) < 0.5 || math.Abs(p.Z+31.5) < 0.5) && math.Mod(p.X+200, 6) < 2.5
	}})
	block := chamfer(-49, -28, 33, 28, 7)
	plaza := chamfer(-49, 34, 33, 57.5, 7)
	g.Use(Sand).Fill(Prism{block, -0.3, 0.3}, Prism{plaza, -0.3, 0.3})
	g.Paint(Stone, Func{Box{V(gx0, -0.3, gz0), V(gx1, 0.3, gz1)}, func(p Vec3) bool {
		return math.Mod(p.X+200, 5) < 0.6 || math.Mod(p.Z+200, 5) < 0.6
	}})
	// Curbs.
	g.Use(Stone).Fill(Subtract(Prism{block, 0.2, 0.6}, Prism{chamfer(-48, -27, 32, 27, 6.6), 0, 1}))
	// Plaça de Gaudí: lawns around an oval pond.
	pond := EllipticCylinder{P(-6, 45.5), 16, 5.5, -0.7, 0.6}
	g.Paint(Grass, Subtract(Prism{chamfer(-46, 36, 30, 55, 6), -0.3, 0.3}, EllipticCylinder{P(-6, 45.5), 18, 7.3, -1, 1},
		Block(V(-49, -1, 39.5), V(-35, 1, 41.5)), Block(V(-2, -1, 34), V(1, 1, 39))))
	g.Carve(pond)
	g.Use(Water).Fill(EllipticCylinder{P(-6, 45.5), 16, 5.5, -0.6, -0.2})
	g.Use(DarkStone).Fill(Subtract(EllipticCylinder{P(-6, 45.5), 17, 6.4, -0.6, 0.6}, EllipticCylinder{P(-6, 45.5), 16, 5.5, -1, 1}))
	g.Use(Water).Fill(EllipticCylinder{P(-6, 45.5), 16, 5.5, -0.6, -0.2})
	// Plane trees along the block and around the pond.
	var trees []Vec2
	for x := -44.0; x <= 28; x += 6 {
		trees = append(trees, P(x, -30.5))
		if x < -16 || x > 16 {
			trees = append(trees, P(x, 30.2))
		}
	}
	for a := 0.0; a < 2*math.Pi; a += 2 * math.Pi / 14 {
		if math.Abs(math.Sin(a)+1) < 0.12 {
			continue
		}
		q := P(-6+20.5*math.Cos(a), 45.5+8.6*math.Sin(a))
		if q.Y < 53.5 {
			trees = append(trees, q)
		}
	}
	trees = append(trees, P(-40, 50), P(-40, 42), P(24, 50), P(24, 42), P(-44, -20), P(-44, 20), P(28, -20), P(28, 20))
	for i, t := range trees {
		g.Build(kit.Tree{At: t, Base: 0.2, Height: 5 + 1.5*Hash01(i, 77)})
	}
	// Benches by the pond.
	g.Use(Wood)
	for _, x := range []float64{-16, -6, 4} {
		g.Fill(Block(V(x-1.5, 0.3, 37.2), V(x+1.5, 0.9, 38)))
	}

	// -------------------------------------------------------------- Body.
	// Five-aisled nave, three-aisled transept, apse with ambulatory chapels.
	apseHalf := func(r, y0, y1 float64) Solid {
		return Intersect(Cylinder{Center: P(xApse, 0), Radius: r, Y0: y0, Y1: y1}, Block(V(xApse, y0, -r), V(xApse+r, y1, r)))
	}
	outer := []Solid{
		Block(V(xGlory, 0, -zN), V(-xT, hAisle, zN)),
		Block(V(xGlory, 0, -zC), V(xApse, hNave, zC)),
		Block(V(-xT, 0, -zT), V(xT, hAisle, zT)),
		Block(V(-zC, 0, -zT), V(zC, hNave, zT)),
		Block(V(-xT, 0, -xT), V(xT, hCross, xT)),
		Block(V(xT, 0, -zN), V(xApse, hAisle, zN)),
		apseHalf(zN, 0, hAisle),
		apseHalf(zC+0.5, 0, hNave),
	}
	inner := []Solid{
		Block(V(xGlory+wall, 0.3, -zN+wall), V(-xT+wall, hAisle-wall, zN-wall)),
		Block(V(xGlory+wall, 0.3, -zC+wall), V(xApse, hNave-wall, zC-wall)),
		Block(V(-xT+wall, 0.3, -zT+wall), V(xT-wall, hAisle-wall, zT-wall)),
		Block(V(-zC+wall, 0.3, -zT+wall), V(zC-wall, hNave-wall, zT-wall)),
		Block(V(-xT+wall, 0.3, -xT+wall), V(xT-wall, hCross-wall, xT-wall)),
		Block(V(xT-wall, 0.3, -zN+wall), V(xApse+0.1, hAisle-wall, zN-wall)),
		apseHalf(zN-wall, 0.3, hAisle-wall),
		apseHalf(zC+0.5-wall, 0.3, hNave-wall),
	}
	g.Use(Stone).Fill(outer...)
	// Roofs: gabled central vessels, a conical choir roof.
	roofs := []Solid{
		Gable{P(xGlory, -zC-0.6), P(-xT, zC+0.6), hNave, AxisX},
		Gable{P(-zC-0.6, -zT), P(zC+0.6, zT), hNave, AxisZ},
		Gable{P(xT, -zC-0.6), P(xApse, zC+0.6), hNave, AxisX},
		Intersect(Frustum{P(xApse, 0), zC + 1.1, 0, hNave, hNave + zC + 1.1}, Block(V(xApse, hNave, -6), V(xApse+6, hNave+6, 6))),
	}
	g.Use(DarkStone).Fill(roofs...)
	g.Carve(
		Gable{P(xGlory+1, -zC+0.2), P(-xT, zC-0.2), hNave + 1.2, AxisX},
		Gable{P(-zC+0.2, -zT+1), P(zC-0.2, zT-1), hNave + 1.2, AxisZ},
		Gable{P(xT, -zC+0.2), P(xApse, zC-0.2), hNave + 1.2, AxisX},
	)
	g.Paint(Slate, Func{Box{V(xGlory, hNave, -zT), V(xApse+6, hNave+7, zT)}, func(p Vec3) bool {
		return math.Mod(p.Y-hNave, 2.4) < 0.8
	}})
	g.Carve(inner...)
	// Cornices.
	g.With(Sand, func() {
		g.Fill(Subtract(Block(V(xGlory-0.4, hAisle-1, -zN-0.4), V(-xT+0.4, hAisle+0.4, zN+0.4)),
			Block(V(xGlory+0.6, hAisle-2, -zN+0.6), V(-xT-0.6, hAisle+1, zN-0.6))))
		g.Fill(Subtract(Block(V(-xT-0.4, hAisle-1, -zT), V(xT+0.4, hAisle+0.4, zT)),
			Block(V(-xT+0.6, hAisle-2, -zT-1), V(xT-0.6, hAisle+1, zT+1))))
		g.Fill(Subtract(Block(V(xGlory, hNave-1, -zC-0.4), V(xApse, hNave+0.4, zC+0.4)),
			Block(V(xGlory-1, hNave-2, -zC+0.6), V(xApse+1, hNave+1, zC-0.6))))
		g.Fill(Subtract(apseHalf(zN+0.4, hAisle-1, hAisle+0.4), apseHalf(zN-0.6, hAisle-2, hAisle+1)))
	})
	// Aisle windows: tall parabolic lights with a round oculus above, between buttresses.
	parabola := func(cx, y0, hw, h float64) func(across, y float64) bool {
		return func(across, y float64) bool {
			d := (across - cx) / hw
			return y >= y0 && math.Abs(d) <= 1 && y <= y0+h*(1-d*d*d*d)
		}
	}
	type win struct{ across, y0, hw, h float64 }
	var bays []float64
	for x := xGlory + 4.2; x < -xT-1; x += 5.2 {
		bays = append(bays, x)
	}
	for _, x := range bays {
		for _, w := range []win{{x, 3, 1.3, 8}, {x, 16, 1.0, 5}} {
			op := parabola(w.across, w.y0, w.hw, w.h)
			g.Carve(Func{Box{V(w.across-2, w.y0, -zN-1), V(w.across+2, w.y0+w.h, zN+1)}, func(p Vec3) bool {
				return op(p.X, p.Y) && (math.Abs(p.Z) > zN-wall-0.4 || (p.Y > hAisle && math.Abs(p.Z) > zC-wall-0.4))
			}})
		}
		for _, oc := range []struct{ y, z, r float64 }{{12.6, zN, 1.1}, {21.4, zC, 0.9}} {
			oc := oc
			g.Carve(Func{Box{V(x-2, oc.y-2, -zN-1), V(x+2, oc.y+2, zN+1)}, func(p Vec3) bool {
				return math.Abs(math.Abs(p.Z)-oc.z) < 1.5 && math.Hypot(p.X-x, p.Y-oc.y) < oc.r
			}})
		}
	}
	// Transept and apse windows.
	for _, z := range []float64{-12.5, 12.5} {
		for _, sx := range []float64{-1, 1} {
			op := parabola(z, 3, 1.3, 8)
			g.Carve(Func{Box{V(sx*xT-2, 3, z-2), V(sx*xT+2, 11, z+2)}, func(p Vec3) bool {
				return op(p.Z, p.Y) && math.Abs(math.Abs(p.X)-xT) < wall+0.4
			}})
		}
	}
	for i := 0; i < 6; i++ {
		a := -math.Pi/2 + float64(i+1)*math.Pi/7
		d := P(math.Cos(a), math.Sin(a))
		side := P(-d.Y, d.X)
		g.Carve(Func{Box{V(xApse-1, 3, -zN-1), V(xApse+zN+1, 12, zN+1)}, func(p Vec3) bool {
			v := p.Plan().Sub(P(xApse, 0))
			along := v.Dot(d)
			across := math.Abs(v.Dot(side))
			return along > zN-wall-0.5 && along < zN+0.5 && across < 1.2*(1-math.Pow((p.Y-3)/8, 4)) && p.Y > 3
		}})
	}
	// Interiors read dark through every opening.
	for _, s := range inner {
		b := s.Bounds().Grow(0.6)
		s := s
		g.Paint(Charcoal, Func{b, func(p Vec3) bool {
			for _, q := range []Vec3{V(0.6, 0, 0), V(-0.6, 0, 0), V(0, 0.6, 0), V(0, -0.6, 0), V(0, 0, 0.6), V(0, 0, -0.6)} {
				if s.Contains(p.Add(q)) {
					return true
				}
			}
			return false
		}})
	}
	// Aisle buttresses with stone pinnacles; clerestory pinnacles with fruit baskets.
	for i, x := range bays {
		bx := x + 2.6
		for _, sz := range []float64{-1, 1} {
			g.With(Stone, func() {
				g.Fill(Block(V(bx-0.6, 0, sz*zN-0.9), V(bx+0.6, hAisle+1.5, sz*zN+0.9)))
			})
			g.With(Sand, func() { g.FillVolume(Frustum{P(bx, sz*(zN+0.2)), 0.9, 0.3, hAisle + 1.5, hAisle + 5}) })
			g.With(Gold, func() { g.Fill(Sphere(V(bx, hAisle+5.4, sz*(zN+0.2)), 0.65)) })
			s.pinnacle(g, P(x, sz*(zC+0.3)), hNave, 4.5, i*2+int(sz+1))
		}
	}

	// ------------------------------------------------------- Nativity façade (+Z).
	// Three deep portals under dripping, sculpted gables; the cypress of life on top.
	{
		zf := zFacade + 6.0 // Front face.
		height := func(x float64) float64 {
			ax := math.Abs(x)
			h := 16.0
			h = math.Max(h, 30-9*(x/5.5)*(x/5.5))
			h = math.Max(h, 24.5-6*((ax-9.6)/3.2)*((ax-9.6)/3.2))
			if ax > 15 {
				return 0
			}
			return h
		}
		portal := func(p Vec3) bool {
			ax := math.Abs(p.X)
			if ax < 3.4 {
				return p.Y <= 1+14*(1-math.Pow(ax/3.4, 2))
			}
			d := (ax - 9.6) / 2.2
			return math.Abs(d) <= 1 && p.Y <= 1+10*(1-d*d)
		}
		box := Box{V(-15, 0, zFacade-0.5), V(15, 31, zf+2.2)}
		relief := func(p Vec3) float64 {
			n := s.noise(p.X/2.1, p.Y/2.1, 11)
			ix := int(math.Floor(p.X / 0.8))
			th := 0.2 + 1.3*n
			for b := 0; b < 6; b++ {
				band := 5.5 + float64(b)*4.4
				l := 0.8 + 3.2*Hash01(ix, b, 3)
				if p.Y <= band && p.Y >= band-l && Hash01(ix, b, 5) > 0.35 {
					th += 0.8 * (1 - (band-p.Y)/l*0.5)
				}
			}
			return th
		}
		g.Use(Sand).Fill(Func{box, func(p Vec3) bool {
			h := height(p.X)
			if p.Y > h || p.Z < zFacade-0.5 {
				return false
			}
			th := relief(p)
			edge := h - p.Y
			if edge < 1.5 {
				th = math.Min(th, 0.6+edge)
			}
			return p.Z <= zf+th
		}})
		g.Carve(Func{Box{V(-14, 0.3, zFacade+0.7), V(14, 30, zf-1.2)}, func(p Vec3) bool {
			return p.Y < height(p.X)-1.3 && math.Abs(p.X) < 14
		}})
		// Portals cut back to the transept and carry stepped archivolts.
		g.Carve(Func{Box{V(-12, 0.3, zFacade-2), V(12, 16, zf+3)}, func(p Vec3) bool {
			inset := (zf + 2.2 - p.Z) * 0.25
			q := V(p.X*(1+0.0*inset), p.Y+inset, p.Z)
			return portal(q) || (portal(p) && p.Z < zf-1)
		}})
		g.Paint(DarkStone, Func{box, func(p Vec3) bool { return s.noise(p.X/1.3, p.Y/1.3, 4) < 0.3 }})
		g.Paint(Stone, Func{box, func(p Vec3) bool { return s.noise(p.X/1.7, p.Y/1.7, 9) > 0.72 }})
		g.Paint(Charcoal, Func{Box{V(-12, 0, zFacade-2), V(12, 16, zf-0.6)}, func(p Vec3) bool {
			return portal(V(p.X, p.Y-0.6, p.Z)) || portal(V(p.X+0.6, p.Y, p.Z)) || portal(V(p.X-0.6, p.Y, p.Z))
		}})
		// The trumeau column and sculpted figure groups in each portal.
		g.With(Sand, func() {
			g.Fill(Cylinder{Center: P(0, zf-1.5), Radius: 0.7, Y0: 0, Y1: 9})
			g.Fill(Ellipsoid{V(0, 10.2, zf-1.5), V(2.0, 1.4, 1.0)})
			for _, x := range []float64{-9.6, 9.6} {
				g.Fill(Ellipsoid{V(x, 7.8, zf-1.5), V(1.1, 1.3, 0.8)})
			}
		})
		g.With(Navy, func() { g.Fill(Sphere(V(0, 13.2, zf-1.2), 1.0)) }) // Star of Bethlehem halo.
		g.With(Gold, func() { g.Fill(Sphere(V(0, 13.2, zf-0.4), 0.6)) })
		// The green cypress, the tree of life, with white doves.
		cy := P(0, zf-2.5)
		g.With(Foliage, func() { g.FillVolume(Frustum{cy, 2.6, 0.4, 27.5, 39}) })
		g.Paint(Grass, Func{Box{V(-3, 27, cy.Y-3), V(3, 40, cy.Y+3)}, func(p Vec3) bool { return math.Mod(p.Y, 2.2) < 0.9 }})
		g.With(White, func() {
			for i := 0; i < 7; i++ {
				a := float64(i) * 2.4
				y := 29 + float64(i)*1.4
				r := 2.6*(1-(y-27.5)/11.5) + 0.6
				g.Fill(Sphere(Polar(cy, r, a).At(y), 0.6))
			}
		})
		g.With(Red, func() { g.Fill(Capsule{cy.At(39), cy.At(41), 0.5}) })
		g.With(White, func() { g.Fill(Sphere(cy.At(41.3), 0.7)) })
	}

	// -------------------------------------------------------- Passion façade (-Z).
	// A stark porch on six leaning bone columns, a canopy, and a pediment of
	// bone-like shafts under a pointed crown.
	{
		zw := -zFacade - 1.5 // Façade wall face.
		zp := -zFacade - 8.0 // Porch front.
		g.Use(Stone)
		g.Fill(Block(V(-15, 0, -zFacade-1.5), V(15, 30, -zFacade+0.5)))
		g.Carve(Block(V(-14, 0.3, -zFacade-0.3), V(14, 28.8, -zFacade+0.5)))
		// Three doors.
		for _, x := range []float64{-7.5, 0, 7.5} {
			hw := 1.8
			if x == 0 {
				hw = 2.6
			}
			g.Carve(Block(V(x-hw, 0.3, zw-0.5), V(x+hw, 9, zw+2)))
			g.With(Wood, func() { g.Fill(Block(V(x-hw, 0.3, zw+0.4), V(x+hw, 9, zw+0.9))) })
			g.Paint(Gold, Func{Box{V(x-hw, 0.3, zw+0.3), V(x+hw, 9, zw+1)}, func(p Vec3) bool {
				return Hash01(int(math.Floor(p.X/0.8)), int(math.Floor(p.Y/0.8)), 21) < 0.3
			}})
		}
		// Porch floor and canopy.
		g.Fill(Block(V(-15, 0, zp), V(15, 0.6, zw)))
		canopy := Func{Box{V(-15.5, 15, zp-1), V(15.5, 18.5, zw)}, func(p Vec3) bool {
			t := (zw - p.Z) / (zw - zp + 1)
			return p.Y >= 15.5+1.6*t && p.Y <= 17+1.6*t
		}}
		g.Use(Sand).Fill(canopy)
		g.Paint(Stone, Func{canopy.Box, func(p Vec3) bool { return math.Mod(p.X+40, 2.5) < 0.9 }})
		// Six leaning bone columns, leaning out at the top.
		for _, x := range []float64{-12, -7.5, -2.6, 2.6, 7.5, 12} {
			lean := x * 0.07
			a, b := V(x-lean, 0.5, zp+1.6), V(x+lean, 16, zp+0.2)
			g.Use(Stone).FillVolume(Capsule{a, b, 0.75})
			for _, f := range []float64{0.04, 0.33, 0.66, 0.95} {
				q := a.Add(b.Sub(a).Scale(f))
				g.FillVolume(Ellipsoid{q, V(1.15, 0.9, 1.15)})
			}
		}
		// Sculpture groups in the porch.
		for i, x := range []float64{-10, -4.5, 4.5, 10} {
			g.With(DarkStone, func() {
				g.FillVolume(Ellipsoid{V(x, 3+float64(i%2)*7, zw-1.2), V(1.2, 2, 0.9)})
			})
		}
		g.With(DarkStone, func() { g.FillVolume(Capsule{V(0, 9.5, zw-1), V(0, 14.5, zw-1), 0.7}) })
		// Upper pediment: bone shafts rising to a pointed crown.
		zc := zw - 2.5
		crown := func(x float64) float64 { return 31 - math.Abs(x)*0.55 }
		for x := -13.0; x <= 13.01; x += 2.0 {
			top := crown(x) - 2.2
			g.Use(Stone).FillVolume(Capsule{V(x, 18, zc), V(x, top, zc), 0.65})
			g.FillVolume(Ellipsoid{V(x, (18+top)/2, zc), V(0.95, 0.8, 0.95)})
		}
		g.Use(Sand).FillVolume(Func{Box{V(-15, 22, zc-1), V(15, 32, zc+1)}, func(p Vec3) bool {
			c := crown(p.X)
			return math.Abs(p.X) <= 14.6 && p.Y <= c && p.Y >= c-2.2 && math.Abs(p.Z-zc) <= 0.8
		}})
		g.With(Stone, func() { g.Fill(Capsule{V(0, 30, zc), V(0, 35, zc), 0.6}) })
		g.With(Gold, func() {
			g.Fill(Capsule{V(-1.6, 33.6, zc), V(1.6, 33.6, zc), 0.5})
			g.Fill(Sphere(V(0, 35.4, zc), 0.6))
		})
		g.Paint(Charcoal, Block(V(-14, 0.3, zw-0.2), V(14, 30, zw+0.6)))
		for _, x := range []float64{-7.5, 0, 7.5} {
			g.Paint(Wood, Block(V(x-2.6, 0.3, zw+0.3), V(x+2.6, 9, zw+1)))
		}
	}

	// ---------------------------------------------------------- Glory façade (-X).
	{
		xw := xGlory
		xp := xGlory - 9.0
		g.Use(Stone)
		g.Fill(Block(V(xw-1, 0, -14), V(xw+0.5, 26, 14)))
		g.Carve(Block(V(xw-0.2, 0.3, -13), V(xw+0.5, 25, 13)))
		for _, z := range []float64{-9, -4.5, 0, 4.5, 9} {
			hw := 1.5
			g.Carve(Func{Box{V(xw-2, 0.3, z-hw), V(xw+1, 12, z+hw)}, func(p Vec3) bool {
				d := (p.Z - z) / hw
				return p.Y <= 1+10*(1-d*d)
			}})
		}
		g.Paint(Charcoal, Block(V(xw-0.4, 0.3, -13.4), V(xw+1, 25.5, 13.4)))
		// Porch of the Glory on columns, under a cloud-crested roof with lanterns.
		g.Fill(Block(V(xp, 0, -15), V(xw, 0.6, 15)))
		for z := -13.5; z <= 13.6; z += 3.0 {
			g.FillVolume(Cylinder{Center: P(xp+1.2, z), Radius: 0.8, Y0: 0.6, Y1: 13})
			g.FillVolume(Frustum{P(xp+1.2, z), 0.8, 1.4, 11.5, 13})
		}
		g.Use(Sand).Fill(Block(V(xp, 13, -15), V(xw, 14.5, 15)))
		g.FillVolume(Func{Box{V(xp, 14.5, -15), V(xw, 22, 15)}, func(p Vec3) bool {
			n := s.noise(p.Z/2.5, p.X/2.5, 31)
			h := 15 + 4*n*(1-math.Abs(p.Z)/16)
			return p.Y <= h && (p.Y < 15.5 || p.X > xp+3 || n > 0.6)
		}})
		for _, z := range []float64{-9, 0, 9} {
			g.With(White, func() { g.FillVolume(Frustum{P(xp+5, z), 1.6, 0.4, 14.5, 21}) })
			g.With(Gold, func() { g.Fill(Sphere(V(xp+5, 21.5, z), 0.8)) })
		}
		// Steps down to the street.
		for i := 0; i < 3; i++ {
			g.Use(Stone).Fill(Block(V(xp-1-float64(i)*1.2, 0, -12+float64(i)), V(xp, 0.6-float64(i)*0.4, 12-float64(i))))
		}
	}

	// -------------------------------------------------------------- Towers.
	// Twelve apostle towers: Nativity (+Z), Passion (-Z), Glory (-X).
	type spire struct {
		c          Vec2
		shaft, top float64
		body, band Color
	}
	var spires []spire
	for _, x := range []float64{-12.5, -6, 6, 12.5} {
		inner := math.Abs(x) < 10
		hN, hP := 41.0, 43.0
		if inner {
			hN, hP = 44.5, 46.5
		}
		spires = append(spires, spire{P(x, zFacade+2), 22, hN, Sand, DarkStone}, spire{P(x, -zFacade-1.5), 24, hP, Stone, Sand})
	}
	for _, z := range []float64{-12.5, -6, 6, 12.5} {
		h := 47.0
		if math.Abs(z) < 10 {
			h = 50.5
		}
		spires = append(spires, spire{P(xGlory-2, z), 22, h, Stone, DarkStone})
	}
	for i, t := range spires {
		s.bellTower(g, t.c, t.shaft, t.top, 2.6, i+1, t.body, t.band)
	}

	// Four Evangelists around the crossing.
	for i, c := range []Vec2{P(-8.6, -8.6), P(8.6, -8.6), P(8.6, 8.6), P(-8.6, 8.6)} {
		end := s.centralTower(g, c, hAisle-0.5, 64, 3.9, 10, 0.74, 1.3)
		g.With(White, func() { g.FillVolume(Ellipsoid{c.At(end + 1.8), V(1.6, 2.4, 1.6)}) })
		s.mosaic(g, Ellipsoid{c.At(end + 1.8), V(2, 2.9, 2)}, 0.8, 40+i, White, White, Gold, Glass)
		// Winged symbol: two white wings and a gold crown.
		g.With(White, func() {
			for _, sd := range []float64{-1, 1} {
				dir := Polar(P(0, 0), 1, float64(i)*math.Pi/2+math.Pi/4)
				side := P(-dir.Y, dir.X).Scale(sd)
				g.Fill(Capsule{c.Add(side.Scale(1.2)).At(end + 2.5), c.Add(side.Scale(2.6)).At(end + 4.3), 0.55})
			}
		})
		g.With(Gold, func() { g.Fill(Sphere(c.At(end+4.9), 0.8)) })
	}

	// The Virgin Mary tower over the apse, crowned with a twelve-pointed glass star.
	{
		c := P(xApse+2, 0)
		end := s.centralTower(g, c, hNave-1, 70, 4.6, 12, 0.72, 1.4)
		star := c.At(end + 3.2)
		g.With(Gold, func() { g.Fill(Sphere(star, 1.2)) })
		phi := (1 + math.Sqrt(5)) / 2
		var dirs []Vec3
		for _, sgn1 := range []float64{-1, 1} {
			for _, sgn2 := range []float64{-1, 1} {
				dirs = append(dirs, V(0, sgn1, sgn2*phi), V(sgn1, sgn2*phi, 0), V(sgn2*phi, 0, sgn1))
			}
		}
		g.With(Glass, func() {
			g.Fill(Func{Box{star.Sub(V(3.4, 3.4, 3.4)), star.Add(V(3.4, 3.4, 3.4))}, func(p Vec3) bool {
				v := p.Sub(star)
				for _, d := range dirs {
					d = d.Scale(1 / d.Length())
					t := v.Dot(d)
					if t < 0 || t > 3.2 {
						continue
					}
					if v.Sub(d.Scale(t)).Length() <= 1.05*(1-t/3.2)+0.2 {
						return true
					}
				}
				return false
			}})
		})
		g.Paint(Gold, Func{Box{star.Sub(V(4, 4, 4)), star.Add(V(4, 4, 4))}, func(p Vec3) bool {
			return p.Sub(star).Length() > 2.5
		}})
	}

	// The Jesus tower, tallest of all, with its four-armed white cross.
	{
		c := P(0, 0)
		end := s.centralTower(g, c, hCross-0.5, 84, 6.6, 16, 0.66, 1.6)
		g.With(White, func() {
			g.Fill(Block(V(-0.8, end-0.5, -0.8), V(0.8, end+7, 0.8)))
			y := end + 4.4
			g.Fill(Block(V(-2.8, y-0.75, -0.75), V(2.8, y+0.75, 0.75)))
			g.Fill(Block(V(-0.75, y-0.75, -2.8), V(0.75, y+0.75, 2.8)))
		})
		g.Paint(Glass, Func{Box{V(-3, end, -3), V(3, end+8, 3)}, func(p Vec3) bool {
			return math.Abs(p.X) > 2 || math.Abs(p.Z) > 2 || p.Y > end+6.3
		}})
		g.With(Gold, func() { g.Fill(Sphere(c.At(end+4.4), 1.15)) })
	}

	// Apse spires: seven slender towers ringing the chevet with golden wheat finials.
	for i := 0; i < 7; i++ {
		a := -math.Pi/2 + (float64(i)+0.5)*math.Pi/7
		c := Polar(P(xApse, 0), zN-0.4, a)
		g.With(Stone, func() {
			g.FillVolume(Frustum{c, 1.5, 1.2, 0, hAisle + 1})
			g.FillVolume(Frustum{c, 1.2, 0.35, hAisle + 1, hAisle + 10})
		})
		g.Paint(Sand, Func{Box{V(c.X-2, hAisle+1, c.Y-2), V(c.X+2, hAisle+11, c.Y+2)}, func(p Vec3) bool {
			return math.Mod(p.Y-hAisle, 2.5) < 0.8
		}})
		g.With(Gold, func() {
			g.Fill(Capsule{c.At(hAisle + 10), c.At(hAisle + 12), 0.5})
			g.Fill(Ellipsoid{c.At(hAisle + 12.5), V(0.7, 1.1, 0.7)})
		})
		if i%2 == 0 {
			g.With(Red, func() { g.Fill(Sphere(c.At(hAisle+9), 0.6)) })
		}
	}

	return Sweep(P(0, 0), math.Pi/2)
}
