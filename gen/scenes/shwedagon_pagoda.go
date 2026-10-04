package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// ShwedagonPagoda is the great gilded stupa of Yangon on Singuttara hill. A square
// white marble platform carries the octagonal stepped plinth ringed by 64 small gilded
// stupas, the bell, turban band, inverted alms bowl, mouldings, lotus petals, banana
// bud and jewelled hti crown. Pyatthat-roofed shrine halls stand at the cardinal
// points and round the platform edge with medium stupas, planetary posts and a Bodhi
// tree. Four covered stairways with stepped tiered roofs descend the wooded hill to
// gate towers guarded by pairs of white chinthe. Stairways face ±X/±Z.
type ShwedagonPagoda struct{}

func init() { Register(ShwedagonPagoda{}) }

func (ShwedagonPagoda) Name() string  { return "shwedagon_pagoda" }
func (ShwedagonPagoda) Title() string { return "Shwedagon Pagoda" }

const (
	shwedagonPY = 13.0 // Platform top.
	shwedagonPH = 28.0 // Platform half size.
	shwedagonHT = 10.8 // Hill top round the platform.
	shwedagonS  = 50.0 // Site superellipse radius.
	shwedagonUE = 42.0 // Stair foot (gate) distance from the center.
	shwedagonCW = 2.4  // Stair corridor inner half width.
)

var shwedagonDir = [4]Vec2{{1, 0}, {0, 1}, {-1, 0}, {0, -1}}

func (ShwedagonPagoda) smooth(a, b, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-a)/(b-a)))
	return t * t * (3 - 2*t)
}

// oct is the octagonal "radius" (apothem) of a plan offset; faces normal to the axes.
func (ShwedagonPagoda) oct(x, z float64) float64 {
	ax, az := math.Abs(x), math.Abs(z)
	return math.Max(math.Max(ax, az), (ax+az)/math.Sqrt2)
}

func (ShwedagonPagoda) cheb(x, z float64) float64 { return math.Max(math.Abs(x), math.Abs(z)) }

// local converts a world point to side coordinates: u outward, v along the side.
func (ShwedagonPagoda) local(side int, p Vec3) (float64, float64) {
	n := shwedagonDir[side]
	return p.X*n.X + p.Z*n.Y, p.X*n.Y - p.Z*n.X
}

// at converts side coordinates back to a world point.
func (ShwedagonPagoda) at(side int, u, v, y float64) Vec3 {
	n := shwedagonDir[side]
	return V(u*n.X+v*n.Y, y, u*n.Y-v*n.X)
}

func (s ShwedagonPagoda) box(side int, u0, u1, v0, v1, y0, y1 float64) Box {
	return BoxOf(s.at(side, u0, v0, y0), s.at(side, u1, v1, y1)).Union(BoxOf(s.at(side, u0, v1, y0), s.at(side, u1, v0, y1)))
}

// sideFunc is a solid over a side-local rectangle.
func (s ShwedagonPagoda) sideFunc(side int, u0, u1, v0, v1, y0, y1 float64, f func(u, v, y float64) bool) Solid {
	return Func{s.box(side, u0, u1, v0, v1, y0, y1).Grow(0.01), func(p Vec3) bool {
		u, v := s.local(side, p)
		return u >= u0 && u <= u1 && v >= v0 && v <= v1 && p.Y >= y0 && p.Y <= y1 && (f == nil || f(u, v, p.Y))
	}}
}

// ell is an ellipsoid given in side coordinates.
func (s ShwedagonPagoda) ell(side int, u, v, y, ru, ry, rv float64) Ellipsoid {
	r := V(ru, ry, rv)
	if shwedagonDir[side].X == 0 {
		r = V(rv, ry, ru)
	}
	return Ellipsoid{s.at(side, u, v, y), r}
}

// lathe fills a solid of revolution from y0 to y1 whose outer radius r(u, angle) is
// given over u in [0, 1]; thick > 0 makes it a shell.
func (ShwedagonPagoda) lathe(g *Grid, c Vec2, y0, y1, thick float64, r func(u, a float64) float64) {
	rmax := 0.0
	for i := 0; i <= 64; i++ {
		for k := 0; k < 16; k++ {
			rmax = math.Max(rmax, r(float64(i)/64, float64(k)*math.Pi/8))
		}
	}
	rmax += 0.6
	g.FillVolume(Func{Box{V(c.X-rmax, y0, c.Y-rmax), V(c.X+rmax, y1, c.Y+rmax)}, func(p Vec3) bool {
		if p.Y < y0 || p.Y > y1 {
			return false
		}
		u := (p.Y - y0) / (y1 - y0)
		dx, dz := p.X-c.X, p.Z-c.Y
		rr := r(u, math.Atan2(dz, dx))
		d := math.Hypot(dx, dz)
		return d <= rr && (thick <= 0 || d >= rr-thick)
	}})
}

// groundBase is the smooth hill height at a plan point, or false outside the site.
func (s ShwedagonPagoda) groundBase(x, z float64) (float64, bool) {
	S := shwedagonS
	rho := math.Pow(math.Pow(math.Abs(x)/S, 3.5)+math.Pow(math.Abs(z)/S, 3.5), 1/3.5)
	if rho > 1 {
		return 0, false
	}
	t := s.smooth(shwedagonPH+1, S, rho*S)
	return shwedagonHT - (shwedagonHT-0.5)*t, true
}

func (s ShwedagonPagoda) ground(x, z float64) (float64, bool) {
	h, ok := s.groundBase(x, z)
	if !ok {
		return 0, false
	}
	S := shwedagonS
	rho := math.Pow(math.Pow(math.Abs(x)/S, 3.5)+math.Pow(math.Abs(z)/S, 3.5), 1/3.5)
	t := s.smooth(shwedagonPH+1, S, rho*S)
	h += 1.6 * t * (1 - t) * math.Sin(x*0.17+1.1) * math.Cos(z*0.13-0.7)
	return h, true
}

// stairY is the stepped floor height of a covered stairway at distance u.
func (s ShwedagonPagoda) stairY(u float64) float64 {
	if u <= shwedagonPH {
		return shwedagonPY
	}
	g0, _ := s.groundBase(u, 0)
	t := math.Min(1, (u-shwedagonPH)/(shwedagonUE-shwedagonPH))
	y := g0 + 2.2 - 1.5*t
	return math.Ceil(y/0.4-1e-9) * 0.4
}

// terrain lays the wooded hill as a surface shell around the platform.
func (s ShwedagonPagoda) terrain(g *Grid) {
	sp := g.Span()
	lo := func(v float64) int32 { return int32(math.Floor(v * 2 / sp)) }
	S := shwedagonS
	for par := int32(0); par < 2; par++ {
		for X := lo(-S - 1); X <= lo(S)+1; X++ {
			if X&1 != par {
				continue
			}
			for Z := lo(-S - 1); Z <= lo(S)+1; Z++ {
				if Z&1 != par {
					continue
				}
				x, z := float64(X)*sp/2, float64(Z)*sp/2
				if s.cheb(x, z) < shwedagonPH-0.8 {
					continue
				}
				h, ok := s.ground(x, z)
				if !ok {
					continue
				}
				low := h
				for _, o := range [4]Vec2{{0.8, 0}, {-0.8, 0}, {0, 0.8}, {0, -0.8}} {
					nh, nok := s.ground(x+o.X, z+o.Y)
					if !nok {
						nh = 0
					}
					low = math.Min(low, nh)
				}
				col := Grass
				if math.Sin(x*0.21+2*math.Sin(z*0.11))+math.Sin(z*0.25+1.5*math.Sin(x*0.09))+0.5*Hash01(int(X), int(Z), 3) > 0.95 {
					col = Foliage
				}
				ax, az := math.Abs(x), math.Abs(z)
				if math.Min(ax, az) < 2.6 && math.Max(ax, az) > shwedagonUE+2 {
					col = Sand // Approach paths beyond the gates.
				}
				for Y := lo(low - 0.8); Y <= lo(h)+1; Y++ {
					y := float64(Y) * sp / 2
					if Y&1 != par || y < low-0.8 || y > h {
						continue
					}
					c := Cell{X, Y, Z}
					g.Remove(c)
					g.Use(col)
					g.Add(c)
				}
			}
		}
	}
}

// platform raises the square marble terrace with its paving and balustrade.
func (s ShwedagonPagoda) platform(g *Grid) {
	PH, PY := shwedagonPH, shwedagonPY
	big := Box{V(-PH-1, shwedagonHT-2, -PH-1), V(PH+1, PY+2, PH+1)}
	g.Use(White).Fill(Func{big, func(p Vec3) bool {
		m := s.cheb(p.X, p.Z)
		return m <= PH && m > PH-1.2 && p.Y >= shwedagonHT-1.4 && p.Y <= PY
	}})
	g.Paint(Stone, Func{big, func(p Vec3) bool {
		m := s.cheb(p.X, p.Z)
		return m > PH-0.3 && (p.Y < shwedagonHT+0.4 || math.Abs(p.Y-(PY-1.1)) < 0.25)
	}})
	g.Use(White).Fill(Func{big, func(p Vec3) bool {
		return s.cheb(p.X, p.Z) <= PH && p.Y >= PY-0.9 && p.Y <= PY && s.oct(p.X, p.Z) > 17.6
	}})
	// Marble tiles, the circumambulation walk and the axial approaches.
	g.Paint(Stone, Func{big, func(p Vec3) bool {
		return p.Y > PY-0.5 && p.Y <= PY && Hash01(int(math.Floor(p.X/2)), int(math.Floor(p.Z/2)), 7) < 0.22
	}})
	g.Paint(Sand, Func{big, func(p Vec3) bool {
		if p.Y <= PY-0.5 || p.Y > PY {
			return false
		}
		o := s.oct(p.X, p.Z)
		ax, az := math.Abs(p.X), math.Abs(p.Z)
		return (o > 19 && o < 21.8) || (math.Min(ax, az) < shwedagonCW && math.Max(ax, az) > 21.8)
	}})
	// Low balustrade, open at the four stairways.
	g.Use(White).Fill(Func{big, func(p Vec3) bool {
		m := s.cheb(p.X, p.Z)
		return m <= PH && m > PH-0.8 && p.Y > PY && p.Y <= PY+1.0 &&
			math.Min(math.Abs(p.X), math.Abs(p.Z)) > shwedagonCW+0.7
	}})
	g.Paint(Gold, Func{big, func(p Vec3) bool {
		m := s.cheb(p.X, p.Z)
		return m > PH-0.8 && p.Y > PY+0.6 && p.Y <= PY+1.0
	}})
}

// satellite is a small gilded stupa: pedestal, bell, spire and jewel tip.
func (s ShwedagonPagoda) satellite(g *Grid, c Vec2, y, k float64, pedestal Color) {
	g.Use(pedestal).Fill(BlockAt(c, 1.5*k, 1.5*k, y-0.6, y+0.6*k))
	g.Use(Gold).Fill(BlockAt(c, 1.1*k, 1.1*k, y+0.6*k, y+0.9*k))
	yb := y + 0.9*k
	s.lathe(g, c, yb, yb+1.5*k, 0, func(u, a float64) float64 {
		return 0.3*k + 0.5*k*math.Sqrt(math.Max(0, 1-u*u))
	})
	top := yb + 1.5*k
	s.lathe(g, c, top-0.1, top+2.2*k, 0, func(u, a float64) float64 {
		return math.Max(0.42, 0.45*k*(1-u))
	})
	g.Use(White).AddAt(c.At(top + 2.2*k + 0.3))
}

// stupa raises the great stupa from its octagonal plinth to the diamond bud.
func (s ShwedagonPagoda) stupa(g *Grid) {
	PY := shwedagonPY
	c := P(0, 0)
	big := Box{V(-20, PY-1, -20), V(20, PY+12, 20)}
	type tier struct{ a, y0, y1 float64 }
	tiers := []tier{
		{19.0, PY - 0.8, PY + 2.0}, {16.2, PY + 2.0, PY + 4.0}, {13.8, PY + 4.0, PY + 6.0},
		{12.8, PY + 6.0, PY + 6.8}, {12.1, PY + 6.8, PY + 7.6}, {11.4, PY + 7.6, PY + 8.4},
	}
	g.Use(Gold)
	for i, t := range tiers {
		next := 10.4
		if i+1 < len(tiers) {
			next = tiers[i+1].a
		}
		g.Fill(Func{big, func(p Vec3) bool {
			o := s.oct(p.X, p.Z)
			if p.Y < t.y0 || p.Y > t.y1 || o > t.a {
				return false
			}
			return o > t.a-1.2 || (p.Y > t.y1-0.9 && o > next-1.4)
		}})
		// Pale moulding at the foot of each terrace and a gilded lip.
		if i < 3 {
			g.Paint(White, Func{big, func(p Vec3) bool {
				o := s.oct(p.X, p.Z)
				return o > t.a-0.35 && o <= t.a && p.Y >= t.y0 && p.Y < t.y0+0.5
			}})
		}
	}
	// Three round bands under the bell.
	for i, r := range []float64{10.9, 10.4, 9.9} {
		y0 := PY + 8.4 + 0.8*float64(i)
		g.Use(Gold).FillVolume(Cylinder{c, r, r - 1.3, y0, y0 + 0.8})
		g.Use(Gold).FillVolume(Cylinder{c, r, r - 2.2, y0 + 0.1, y0 + 0.8})
	}
	// The bell.
	yb, yt := PY+10.8, PY+27.0
	const rb, rt = 9.6, 4.9
	s.lathe(g, c, yb, yt, 1.3, func(u, a float64) float64 {
		return rt + (rb-rt)*math.Pow(math.Cos(u*math.Pi/2), 0.8)
	})
	// Turban band with three rings.
	y := yt
	s.lathe(g, c, y-0.4, y+2.6, 1.3, func(u, a float64) float64 { return rt })
	for _, d := range []float64{0.3, 1.2, 2.1} {
		g.Use(Gold).FillVolume(Cylinder{c, rt + 0.85, rt - 0.8, y + d, y + d + 0.55})
	}
	// Inverted alms bowl.
	y += 2.6
	s.lathe(g, c, y, y+3.6, 1.3, func(u, a float64) float64 {
		return 3.7 + 1.7*math.Sqrt(math.Max(0, 1-u*u))
	})
	g.Use(Gold).Fill(Cylinder{c, 4.0, 0, y + 2.8, y + 3.6})
	// Mouldings.
	y += 3.6
	for i, r := range []float64{4.0, 3.4, 3.7, 3.1, 3.4} {
		y0 := y + 0.6*float64(i)
		g.Use(Gold).FillVolume(Cylinder{c, r, r - 1.4, y0, y0 + 0.6})
	}
	// Lotus petals, turned down then up.
	y += 3.0
	petal := func(a float64) float64 { return 0.3 + 0.7*math.Abs(math.Cos(8*a)) }
	s.lathe(g, c, y, y+2.0, 0, func(u, a float64) float64 { return 2.8 + 1.4*(1-u)*petal(a) })
	s.lathe(g, c, y+2.0, y+4.0, 0, func(u, a float64) float64 { return 2.8 + 1.4*u*petal(a+math.Pi/16) })
	g.Paint(White, Func{Box{V(-5, y, -5), V(5, y+4.0, 5)}, func(p Vec3) bool {
		d := math.Hypot(p.X, p.Z)
		return d > 3.6 && (p.Y < y+0.5 || p.Y > y+3.5)
	}})
	// Banana bud.
	y += 4.0
	s.lathe(g, c, y, y+12.0, 1.3, func(u, a float64) float64 {
		return math.Max(0.6, 2.9*math.Pow(1-u, 0.85)*(1+0.12*math.Sin(math.Pi*u)))
	})
	// Hti: seven jewelled umbrella rings round a mast, vane and diamond orb.
	y += 12.0
	hy := y
	g.Use(Gold).Fill(Cylinder{c, 0.7, 0, y - 0.5, y + 5.6})
	for i := 0; i < 7; i++ {
		f := float64(i)
		y0 := y + 0.75*f
		g.Use(Gold).FillVolume(Frustum{c, 2.5 - 0.24*f, 1.5 - 0.16*f, y0, y0 + 0.55})
	}
	g.Paint(Glass, Func{Box{V(-3, hy, -3), V(3, hy+5.6, 3)}, func(p Vec3) bool {
		return math.Hypot(p.X, p.Z) > 0.9 && Hash01(int(math.Round(p.X*4)), int(math.Round(p.Y*4)), int(math.Round(p.Z*4)), 61) < 0.18
	}})
	g.Paint(Red, Func{Box{V(-3, hy, -3), V(3, hy+5.6, 3)}, func(p Vec3) bool {
		return math.Hypot(p.X, p.Z) > 0.9 && Hash01(int(math.Round(p.X*4)), int(math.Round(p.Y*4)), int(math.Round(p.Z*4)), 62) < 0.12
	}})
	g.Paint(White, Func{Box{V(-3, hy, -3), V(3, hy+5.6, 3)}, func(p Vec3) bool {
		return math.Hypot(p.X, p.Z) > 0.9 && Hash01(int(math.Round(p.X*4)), int(math.Round(p.Y*4)), int(math.Round(p.Z*4)), 63) < 0.1
	}})
	y += 5.6
	// Vane: a flag plate on one side of the mast.
	g.Use(Gold).Fill(Cylinder{c, 0.45, 0, y, y + 1.8})
	g.Use(Gold).Fill(Block(V(0.3, y+0.3, -0.3), V(2.0, y+1.4, 0.3)))
	g.Paint(Red, Block(V(1.2, y+0.3, -0.3), V(2.0, y+1.4, 0.3)))
	// Diamond orb.
	g.Use(Glass).FillVolume(Sphere(c.At(y+2.4), 0.85))
	g.Use(White).Fill(Capsule{A: c.At(y + 3.0), B: c.At(y + 3.8), Radius: 0.42})
}

// ring places the 64 small stupas on the lowest terrace, eight on each octagon face.
func (s ShwedagonPagoda) ring(g *Grid) {
	const a = 17.5
	side := 2 * a * math.Tan(math.Pi/8)
	step := side / 8
	for k := 0; k < 8; k++ {
		phi := float64(k) * math.Pi / 4
		mid := Polar(P(0, 0), a, phi)
		tan := P(-math.Sin(phi), math.Cos(phi))
		for j := 0; j < 8; j++ {
			c := mid.Add(tan.Scale((float64(j) - 3.5) * step))
			s.satellite(g, c, shwedagonPY+2.0, 1.0, Gold)
		}
	}
}

// pyatthat stacks tiered roofs with flared eaves, gilded hips and necks, then a spire.
// ex, ez are the lowest eave half sizes along X and Z.
func (s ShwedagonPagoda) pyatthat(g *Grid, c Vec2, y, ex, ez float64, n int, spire float64) {
	for i := 0; i < n; i++ {
		f := 1.0
		if n > 1 {
			f = 1 - 0.58*float64(i)/float64(n-1)
		}
		eX, eZ := ex*f, ez*f
		L := math.Min(eX, eZ) * 0.5
		rh := 0.9 + 0.5*f
		y0 := y
		roof := Func{Box{V(c.X-eX, y0-1.2, c.Y-eZ), V(c.X+eX, y0+rh+0.1, c.Y+eZ)}, func(p Vec3) bool {
			e := math.Min(eX-math.Abs(p.X-c.X), eZ-math.Abs(p.Z-c.Y))
			if e < 0 {
				return false
			}
			t := math.Min(1, e/L)
			hs := y0 + rh*math.Pow(t, 1.4) + 0.3*(1-t)*(1-t)
			return p.Y <= hs && p.Y >= hs-1.05
		}}
		g.Use(Red).Fill(roof)
		g.Paint(Gold, Func{roof.Box, func(p Vec3) bool {
			ax, az := eX-math.Abs(p.X-c.X), eZ-math.Abs(p.Z-c.Y)
			return roof.F(p) && ((i == 0 && math.Min(ax, az) < 0.3) || math.Abs(ax-az) < 0.22)
		}})
		// Upswept gilded horns at the eave corners.
		for _, sx := range []float64{-1, 1} {
			for _, sz := range []float64{-1, 1} {
				b := V(c.X+sx*(eX-0.3), y0+0.2, c.Y+sz*(eZ-0.3))
				g.Use(Gold).Fill(Capsule{A: b, B: b.Add(V(sx*0.3, 0.9, sz*0.3)), Radius: 0.42})
			}
		}
		nX, nZ := eX-L, eZ-L
		g.Use(Gold).Fill(BlockAt(c, 2*nX, 2*nZ, y0+rh-1.0, y0+rh+0.6))
		y = y0 + rh + 0.6
	}
	if spire <= 0 {
		g.Use(Gold).Fill(Capsule{A: c.At(y), B: c.At(y + 1.0), Radius: 0.45})
		return
	}
	g.Use(Gold)
	s.lathe(g, c, y-0.2, y+1.2, 0, func(u, a float64) float64 { return 0.5 + 0.6*math.Sqrt(math.Max(0, 1-u*u)) })
	s.lathe(g, c, y+1.2, y+1.2+spire, 0, func(u, a float64) float64 {
		return math.Max(0.42, 0.75*(1-u)) + 0.25*math.Max(0, math.Cos(u*math.Pi*5))*(1-u)
	})
	g.Use(Glass).AddAt(c.At(y + 1.5 + spire))
}

// pavilion is a tazaung: an open hall of red posts on a white plinth under a pyatthat.
func (s ShwedagonPagoda) pavilion(g *Grid, c Vec2, y0, hx, hz float64, tiers int, spire float64) {
	g.Use(White).Fill(BlockAt(c, 2*hx+0.8, 2*hz+0.8, y0-0.9, y0+0.6))
	g.Paint(Charcoal, BlockAt(c, 2*hx-1.2, 2*hz-1.2, y0+0.2, y0+0.6))
	top := y0 + 3.8
	for _, side := range [2]bool{false, true} {
		l, w := hx, hz
		if side {
			l, w = hz, hx
		}
		n := int(math.Ceil(2 * l / 1.9))
		for i := 0; i <= n; i++ {
			t := -l + 0.35 + (2*l-0.7)*float64(i)/float64(n)
			for _, sg := range []float64{-1, 1} {
				p := P(c.X+t, c.Y+sg*(w-0.35))
				if side {
					p = P(c.X+sg*(w-0.35), c.Y+t)
				}
				g.Use(Red).Fill(BlockAt(p, 0.7, 0.7, y0+0.6, top))
				g.Use(White).Fill(BlockAt(p, 0.9, 0.9, y0+0.6, y0+1.0))
			}
		}
	}
	// Gilded frieze and a dark ceiling.
	g.Use(Gold).Fill(Func{Box{V(c.X-hx, top, c.Y-hz), V(c.X+hx, top+0.7, c.Y+hz)}, func(p Vec3) bool {
		return math.Abs(p.X-c.X) > hx-0.7 || math.Abs(p.Z-c.Y) > hz-0.7
	}})
	g.Use(Charcoal).Fill(BlockAt(c, 2*hx-1.4, 2*hz-1.4, top, top+0.7))
	s.pyatthat(g, c, top+0.6, hx+0.9, hz+0.9, tiers, spire)
}

// stairway builds one covered stairway descending from the platform to its gate.
func (s ShwedagonPagoda) stairway(g *Grid, side int) {
	PH, UE, CW := shwedagonPH, shwedagonUE, shwedagonCW
	n := shwedagonDir[side]
	u0, u1 := PH-1.0, UE+1.5
	// Side walls and a floor slab following the stepped stair.
	g.Use(White).Fill(s.sideFunc(side, u0, u1, -CW-0.7, CW+0.7, -2, shwedagonPY+1, func(u, v, y float64) bool {
		sy := s.stairY(u)
		if math.Abs(v) > CW {
			return y <= sy+0.8 && y >= sy-3.6
		}
		return y <= sy && y >= sy-1.3
	}))
	// Carve the corridor clear of any hill above the steps.
	g.Carve(s.sideFunc(side, PH+0.2, u1+0.1, -CW, CW, -2, shwedagonPY+6, func(u, v, y float64) bool {
		return y > s.stairY(u)+0.1 && y < s.stairY(u)+3.5
	}))
	// Treads in sand between white risers.
	g.Paint(Sand, s.sideFunc(side, PH, u1, -CW, CW, -2, shwedagonPY+1, func(u, v, y float64) bool {
		sy := s.stairY(u)
		return y > sy-0.45 && y <= sy && s.stairY(u+0.5) == sy
	}))
	// Roof segments stepping down the hill, each on red posts.
	const seg = 4.0
	for k := 0; k < 4; k++ {
		ua := PH + seg*float64(k)
		ub := ua + seg
		yr := s.stairY(ua) + 3.7
		for _, pu := range []float64{ua + 0.6, ua + 2.6} {
			for _, sg := range []float64{-1, 1} {
				p := s.at(side, pu, sg*(CW+0.35), 0).Plan()
				g.Use(Red).Fill(BlockAt(p, 0.7, 0.7, s.stairY(pu), yr+0.2))
			}
		}
		mid := s.at(side, (ua+ub)/2, 0, 0).Plan()
		hu, hv := seg/2+0.3, CW+1.5
		ex, ez := hu, hv
		if n.X == 0 {
			ex, ez = hv, hu
		}
		g.Use(Gold).Fill(Func{s.box(side, ua, ub, -CW-0.7, CW+0.7, yr-0.6, yr+0.2), func(p Vec3) bool {
			u, v := s.local(side, p)
			return u >= ua && u <= ub && math.Abs(v) <= CW+0.7 && math.Abs(v) > CW-0.1 && p.Y >= yr-0.6 && p.Y <= yr+0.2
		}})
		s.pyatthat(g, mid, yr, ex, ez, 2, 0)
	}
	// Gate tower at the foot.
	gy := s.stairY(UE)
	gTop := gy + 5.4
	g.Use(White).Fill(s.sideFunc(side, UE-0.5, UE+2.5, -CW-2.0, CW+2.0, gy-3.0, gTop, nil))
	through := AxisX
	if n.X == 0 {
		through = AxisZ
	}
	g.Carve(Arch{Sill: s.at(side, UE+1.0, 0, gy+0.05), HalfWidth: CW - 0.3, Spring: 2.6, Depth: 3.4, Through: through})
	g.Paint(Charcoal, s.sideFunc(side, UE-0.6, UE+2.6, -CW+0.1, CW-0.1, gy-0.5, gy+5, func(u, v, y float64) bool {
		return math.Abs(v) > CW-0.9 || y < gy+0.05
	}))
	// Gilded frame round the archway.
	g.Paint(Gold, Intersect(
		Arch{Sill: s.at(side, UE+1.0, 0, gy+0.05), HalfWidth: CW + 0.4, Spring: 2.6, Depth: 3.6, Through: through},
		s.sideFunc(side, UE+1.9, UE+2.7, -CW-1, CW+1, gy-0.1, gy+6, nil),
	))
	// Front steps down to the approach path.
	fg, _ := s.groundBase(UE+4.5, 0)
	g.Use(White).Fill(s.sideFunc(side, UE+2.5, UE+5.0, -CW+0.2, CW-0.2, fg-1.0, gy, func(u, v, y float64) bool {
		return y <= gy-(gy-fg)*math.Floor((u-UE-2.5)/0.5+1)/5
	}))
	g.Paint(Gold, s.sideFunc(side, UE-0.6, UE+2.6, -CW-2.1, CW+2.1, gTop-0.6, gTop, nil))
	g.Paint(Red, s.sideFunc(side, UE+2.2, UE+2.6, -CW-1.6, CW+1.6, gy+0.5, gTop-0.9, func(u, v, y float64) bool {
		return math.Abs(v) > CW+0.2 && math.Abs(v) < CW+1.3
	}))
	gc := s.at(side, UE+1.0, 0, 0).Plan()
	hu, hv := 2.4, CW+2.7
	ex, ez := hu, hv
	if n.X == 0 {
		ex, ez = hv, hu
	}
	s.pyatthat(g, gc, gTop, ex, ez, 5, 3.5)
	// Chinthe pair guarding the gate.
	for _, sg := range []float64{-1, 1} {
		s.chinthe(g, side, UE+5.0, sg*(CW+2.8))
	}
}

// chinthe is a seated white guardian lion on a pedestal, facing outward.
func (s ShwedagonPagoda) chinthe(g *Grid, side int, u, v float64) {
	gb, _ := s.groundBase(u, v)
	const k = 1.25
	y := gb + 1.2
	g.Use(Stone).Fill(s.sideFunc(side, u-2.0, u+2.0, v-1.6, v+1.6, gb-1.0, y, nil))
	g.Use(White).Fill(s.sideFunc(side, u-1.9, u+1.9, v-1.5, v+1.5, y-0.4, y, nil))
	g.Use(White).FillVolume(
		s.ell(side, u-0.6*k, v, y+1.1*k, 1.3*k, 1.1*k, 1.1*k),
		s.ell(side, u+0.4*k, v, y+2.3*k, 0.95*k, 1.5*k, 0.95*k),
	)
	g.Use(White).Fill(
		Capsule{A: s.at(side, u-0.4*k, v, y+1.3*k), B: s.at(side, u+0.3*k, v, y+2.6*k), Radius: 0.95 * k},
	)
	for _, sg := range []float64{-1, 1} {
		g.Use(White).Fill(Capsule{A: s.at(side, u+1.0*k, v+sg*0.5*k, y+0.3), B: s.at(side, u+0.7*k, v+sg*0.5*k, y+2.0*k), Radius: 0.45 * k})
		g.Use(White).Fill(Capsule{A: s.at(side, u+1.0*k, v+sg*0.5*k, y+0.3), B: s.at(side, u+1.5*k, v+sg*0.5*k, y+0.3), Radius: 0.42 * k})
	}
	// Mane and head.
	g.Use(White).FillVolume(s.ell(side, u+0.1*k, v, y+3.7*k, 0.9*k, 1.3*k, 1.3*k))
	g.Paint(Gold, s.sideFunc(side, u-1.0*k, u+0.6*k, v-1.5*k, v+1.5*k, y+2.6*k, y+5*k, func(uu, vv, yy float64) bool {
		return Hash01(int(math.Round(uu*3)), int(math.Round(vv*3)), int(math.Round(yy*3)), 71) < 0.3
	}))
	g.Use(White).FillVolume(Sphere(s.at(side, u+0.9*k, v, y+4.0*k), 0.95*k))
	g.Use(Red).Fill(s.sideFunc(side, u+1.4*k, u+2.1*k, v-0.5*k, v+0.5*k, y+3.4*k, y+3.8*k, nil))
	g.Use(White).Fill(s.sideFunc(side, u+1.4*k, u+2.0*k, v-0.6*k, v+0.6*k, y+3.85*k, y+4.3*k, nil))
	for _, sg := range []float64{-1, 1} {
		g.Paint(Charcoal, s.sideFunc(side, u+1.3*k, u+2.0*k, v+sg*0.45*k-0.25, v+sg*0.45*k+0.25, y+4.3*k, y+4.7*k, nil))
	}
	g.Use(Gold).Fill(Capsule{A: s.at(side, u+0.7*k, v, y+4.8*k), B: s.at(side, u+0.5*k, v, y+5.6*k), Radius: 0.45})
	// Tail curling up behind.
	g.Use(White).Fill(Capsule{A: s.at(side, u-1.6*k, v, y+0.7*k), B: s.at(side, u-2.0*k, v, y+2.4*k), Radius: 0.42})
	g.Use(Gold).Fill(Sphere(s.at(side, u-1.9*k, v, y+2.8*k), 0.55))
}

// bodhi is the sacred fig on a railed white terrace.
func (s ShwedagonPagoda) bodhi(g *Grid, c Vec2) {
	y := shwedagonPY
	g.Use(White).Fill(Cylinder{c, 4.2, 0, y - 0.5, y + 0.8})
	g.Use(White).Fill(Cylinder{c, 4.2, 3.6, y + 0.8, y + 1.6})
	g.Paint(Gold, Cylinder{c, 4.3, 3.5, y + 1.3, y + 1.6})
	g.Use(Wood).Fill(Capsule{A: c.At(y), B: c.At(y + 5.5), Radius: 1.0})
	for i := 0; i < 5; i++ {
		a := float64(i)*2*math.Pi/5 + 0.4
		tip := Polar(c, 3.0, a).At(y + 7.5 + 0.8*Hash01(i, 81))
		g.Use(Wood).Fill(Capsule{A: c.At(y + 4.5), B: tip, Radius: 0.5})
		col := Foliage
		if i%2 == 0 {
			col = Grass
		}
		g.Use(col).FillVolume(Ellipsoid{Polar(c, 3.4, a).At(y + 8.8), V(3.0, 2.2, 3.0)})
	}
	g.Use(Foliage).FillVolume(Ellipsoid{c.At(y + 10.0), V(3.6, 2.6, 3.6)})
}

// planetary is a white planetary post with its gilded cap at an octagon corner.
func (ShwedagonPagoda) planetary(g *Grid, c Vec2) {
	y := shwedagonPY
	g.Use(White).Fill(BlockAt(c, 1.6, 1.6, y-0.5, y+0.6))
	g.Use(White).Fill(BlockAt(c, 0.8, 0.8, y+0.6, y+3.0))
	g.Use(Gold).Fill(BlockAt(c, 1.4, 1.4, y+3.0, y+3.5))
	g.Use(Gold).Fill(Capsule{A: c.At(y + 3.5), B: c.At(y + 4.6), Radius: 0.45})
	g.Paint(Red, BlockAt(c, 0.9, 0.9, y+1.6, y+2.1))
}

// platformShrines places the cardinal halls, edge pavilions, medium stupas, corner
// halls, planetary posts and the Bodhi tree.
func (s ShwedagonPagoda) platformShrines(g *Grid) {
	PY := shwedagonPY
	for side := 0; side < 4; side++ {
		n := shwedagonDir[side]
		c := s.at(side, 23.4, 0, 0).Plan()
		hu, hv := 2.6, 3.4
		hx, hz := hu, hv
		if n.X == 0 {
			hx, hz = hv, hu
		}
		s.pavilion(g, c, PY, hx, hz, 5, 3.5)
		for _, sg := range []float64{-1, 1} {
			s.pavilion(g, s.at(side, 24.0, sg*9.6, 0).Plan(), PY, 2.0, 2.0, 3, 2.2)
			s.satellite(g, s.at(side, 23.8, sg*15.2, 0).Plan(), PY, 2.0, White)
		}
	}
	corners := []Vec2{{22.5, 22.5}, {-22.5, 22.5}, {-22.5, -22.5}, {22.5, -22.5}}
	for i, c := range corners {
		switch i {
		case 3:
			s.bodhi(g, c)
		case 1:
			s.pavilion(g, c, PY, 2.8, 2.8, 5, 3.0)
		default:
			// A cluster: one larger stupa amid four small ones.
			s.satellite(g, c, PY, 2.6, White)
			for k := 0; k < 4; k++ {
				s.satellite(g, Polar(c, 3.6, math.Pi/4+float64(k)*math.Pi/2), PY, 1.1, White)
			}
		}
	}
	for k := 0; k < 8; k++ {
		s.planetary(g, Polar(P(0, 0), 21.3, math.Pi/8+float64(k)*math.Pi/4))
	}
}

// grove scatters shade trees and palms over the hill, clear of the stairways.
func (s ShwedagonPagoda) grove(g *Grid) {
	const sp = 5.6
	n := int(math.Ceil(shwedagonS / sp))
	for i := -n; i <= n; i++ {
		for j := -n; j <= n; j++ {
			x := (float64(i) + 0.2 + 0.6*Hash01(i, j, 31)) * sp
			z := (float64(j) + 0.2 + 0.6*Hash01(i, j, 32)) * sp
			if s.cheb(x, z) < shwedagonPH+3.0 || math.Min(math.Abs(x), math.Abs(z)) < 7.5 {
				continue
			}
			h, ok := s.ground(x, z)
			if !ok || h < 1.0 || Hash01(i, j, 33) > 0.72 {
				continue
			}
			seed := (i+30)*101 + j + 30
			base := V(x, h, z)
			if Hash01(i, j, 34) < 0.3 {
				s.palm(g, base, seed)
				continue
			}
			th := 2.0 + 1.5*Hash01(seed, 7)
			r := 1.9 + 0.9*Hash01(seed, 8)
			g.Use(Wood).Fill(Capsule{A: base.Add(V(0, -0.5, 0)), B: base.Add(V(0, th, 0)), Radius: 0.45})
			col := Foliage
			if Hash01(seed, 9) < 0.35 {
				col = Grass
			}
			g.Use(col).FillVolume(Ellipsoid{base.Add(V(0, th+r*0.6, 0)), V(r, r*0.8, r)})
		}
	}
}

func (ShwedagonPagoda) palm(g *Grid, base Vec3, seed int) {
	h := 5.0 + 2.0*Hash01(seed, 1)
	lean := Polar(P(0, 0), 0.5+0.8*Hash01(seed, 2), Hash01(seed, 3)*2*math.Pi)
	top := base.Plan().Add(lean).At(base.Y + h)
	g.Use(Wood).Fill(Capsule{A: base.Add(V(0, -0.5, 0)), B: top, Radius: 0.42})
	g.Use(Foliage).FillVolume(Sphere(top, 0.75))
	for k := 0; k < 6; k++ {
		a := (float64(k) + 0.5*Hash01(seed, k, 4)) * 2 * math.Pi / 6
		m := Polar(top.Plan(), 1.5, a).At(top.Y + 0.4)
		tip := Polar(top.Plan(), 2.8, a).At(top.Y - 0.8)
		g.Use(Grass).Fill(Capsule{A: top, B: m, Radius: 0.42}, Capsule{A: m, B: tip, Radius: 0.4})
	}
}

func (s ShwedagonPagoda) Build(g *Grid) Order {
	s.terrain(g)
	s.platform(g)
	for side := 0; side < 4; side++ {
		s.stairway(g, side)
	}
	s.stupa(g)
	s.ring(g)
	s.platformShrines(g)
	s.grove(g)
	return Sweep(P(0, 0), math.Pi/4)
}
