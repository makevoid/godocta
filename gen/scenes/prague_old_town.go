package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// PragueOldTown is Prague's Old Town Square. North is -Z: the Gothic Týn Church rises
// behind the arcaded Týn school and the burgher houses of the north row, its two dark
// spires ringed by gold-tipped turrets. The Old Town Hall tower stands on the east
// side (+X) with the astronomical clock on its south face; the white Baroque church of
// St Nicholas with its green dome closes the north-west corner. Pastel houses with
// volute, stepped and Venetian gables, arcades, dormers and chimneys line the west and
// south sides; the Jan Hus monument and the Marian column stand on the cobbled square,
// with market stalls and trees in its open south-east corner.
type PragueOldTown struct{}

func init() { Register(PragueOldTown{}) }

func (PragueOldTown) Name() string  { return "prague_old_town" }
func (PragueOldTown) Title() string { return "Prague Old Town" }

// pragueOldTownFrame places local-frame solids: local +Z points along plan angle theta
// and the local origin lands on c. House facades lie on local z = 0, facing +Z.
type pragueOldTownFrame struct {
	c     Vec2
	theta float64
}

func (f pragueOldTownFrame) at(s Solid) Solid {
	return Translate(RotateY(s, P(0, 0), f.theta-math.Pi/2), f.c.At(0))
}

// Burgher house roof and gable styles.
const (
	pragueOldTownVolute = iota
	pragueOldTownDormers
	pragueOldTownVenetian
	pragueOldTownStepped
	pragueOldTownTowerHouse
)

// pragueOldTownHouse is one burgher house: a hollow shell w wide (local X) and d deep
// (local -Z) with its facade on local z = 0.
type pragueOldTownHouse struct {
	f                 pragueOldTownFrame
	w, d, eave        float64
	wall, trim, roof  Color
	style             int
	arcade, sgraffito bool
	sideWin           int // Bit 1: windows on the local -X side wall; bit 2: on the +X side.
	seed              int
}

// lancet is a pointed opening across local X centred on x, through local Z from z0 to z1.
func (PragueOldTown) lancet(x, hw, sill, spring, z0, z1 float64) Func {
	apex := spring + math.Sqrt(3)*hw
	return Func{Box{V(x-hw, sill, z0), V(x+hw, apex, z1)}, func(p Vec3) bool {
		ax := math.Abs(p.X - x)
		if ax > hw {
			return false
		}
		rise := p.Y - spring
		if rise <= 0 {
			return true
		}
		d := ax + hw
		return d*d+rise*rise <= 4*hw*hw
	}}
}

// disk is a vertical disk in the local XY plane between z0 and z1, with radius range.
func (PragueOldTown) disk(cx, cy, r0, r1, z0, z1 float64) Func {
	return Func{Box{V(cx-r1, cy-r1, z0), V(cx+r1, cy+r1, z1)}, func(p Vec3) bool {
		r := math.Hypot(p.X-cx, p.Y-cy)
		return r >= r0 && r <= r1
	}}
}

// window carves a recess into a facade on local z = face (outward dir ±1), backs it
// with glass color, and optionally adds a hood and sill in trim.
func (h pragueOldTownHouse) window(g *Grid, x, y0, y1, ww, face, dir float64, pane Color, trim, arched bool) {
	at := h.f.at
	zr := func(a, b float64) (float64, float64) {
		z0, z1 := face+dir*a, face+dir*b
		return math.Min(z0, z1), math.Max(z0, z1)
	}
	var hole, back Solid
	if arched {
		z0, z1 := zr(-0.55, 0.3)
		hole = Arch{V(x, y0, (z0+z1)/2), ww, y1 - y0 - ww, z1 - z0, AxisZ}
		z0, z1 = zr(-1.4, -0.5)
		back = Arch{V(x, y0, (z0+z1)/2), ww, y1 - y0 - ww, z1 - z0, AxisZ}
	} else {
		z0, z1 := zr(-0.55, 0.3)
		hole = Block(V(x-ww, y0, z0), V(x+ww, y1, z1))
		z0, z1 = zr(-1.4, -0.5)
		back = Block(V(x-ww, y0, z0), V(x+ww, y1, z1))
	}
	g.Carve(at(hole))
	g.Paint(pane, at(back))
	if trim {
		g.With(h.trim, func() {
			z0, z1 := zr(-0.2, 0.45)
			if !arched {
				g.Fill(at(Block(V(x-ww-0.3, y1, z0), V(x+ww+0.3, y1+0.5, z1))))
			}
			g.Fill(at(Block(V(x-ww-0.2, y0-0.45, z0), V(x+ww+0.2, y0, z1))))
		})
	}
}

func (h pragueOldTownHouse) build(g *Grid) {
	at := h.f.at
	hw, d, e := h.w/2, h.d, h.eave
	const (
		t  = 1.0  // Wall thickness.
		k  = 1.35 // Roof pitch.
		gf = 4.4  // Ground floor height.
		fh = 3.3  // Upper floor height.
	)
	th := 0.85 * math.Sqrt(1+k*k)

	// Shell.
	g.Use(h.wall)
	g.Fill(at(Block(V(-hw, 0.2, -d), V(hw, e, 0))))
	g.Carve(at(Block(V(-hw+t, 0.3, -d+t), V(hw-t, e+0.2, -t))))
	if h.sgraffito {
		g.Paint(Charcoal, at(Func{Box{V(-hw, gf, -0.5), V(hw, e-0.8, 0.1)}, func(p Vec3) bool {
			a := math.Mod(p.X+p.Y+100, 1.6)
			b := math.Mod(p.X-p.Y+100, 1.6)
			return a < 0.42 || b < 0.42
		}}))
	}

	// Trim: floor band, cornice and corner quoins.
	g.With(h.trim, func() {
		g.Fill(at(Block(V(-hw, gf-0.5, -0.3), V(hw, gf, 0.4))))
		g.Fill(at(Block(V(-hw, e-0.6, -0.3), V(hw, e, 0.5))))
		g.Fill(at(Block(V(-hw, 0.2, -0.3), V(-hw+0.6, e, 0.3))), at(Block(V(hw-0.6, 0.2, -0.3), V(hw, e, 0.3))))
	})

	// Upper-floor windows, front and back.
	n := int(math.Max(2, math.Round(h.w/2.2)))
	bay := h.w / float64(n)
	floors := int((e - gf - 0.4) / fh)
	gothic := h.style == pragueOldTownTowerHouse
	for i := 0; i < floors; i++ {
		y0 := gf + float64(i)*fh + 0.9
		y1 := y0 + 1.9
		for j := 0; j < n; j++ {
			x := -hw + bay*(float64(j)+0.5)
			h.window(g, x, y0, y1, 0.5, 0, 1, Charcoal, !gothic, gothic)
			h.window(g, x, y0, y1, 0.5, -d, -1, Charcoal, false, false)
		}
	}

	// Windows on exposed side walls.
	for _, sx := range []float64{-1, 1} {
		if (sx < 0 && h.sideWin&1 == 0) || (sx > 0 && h.sideWin&2 == 0) {
			continue
		}
		ns := int(math.Max(2, math.Round(d/2.4)))
		sb := d / float64(ns)
		xs := func(a, b float64) (float64, float64) {
			x0, x1 := sx*(hw+a), sx*(hw+b)
			return math.Min(x0, x1), math.Max(x0, x1)
		}
		for i := -1; i < floors; i++ {
			y0 := gf + float64(i)*fh + 0.9
			if i < 0 {
				y0 = 1.2
			}
			for j := 0; j < ns; j++ {
				z := -d + sb*(float64(j)+0.5)
				x0, x1 := xs(-0.55, 0.3)
				g.Carve(at(Block(V(x0, y0, z-0.5), V(x1, y0+1.9, z+0.5))))
				x0, x1 = xs(-1.4, -0.5)
				g.Paint(Charcoal, at(Block(V(x0, y0, z-0.5), V(x1, y0+1.9, z+0.5))))
			}
		}
	}

	// Ground floor: an arcade or a shop front.
	if h.arcade {
		na := int(math.Max(1, math.Round(h.w/3.3)))
		ab := h.w / float64(na)
		ahw := ab/2 - 0.55
		g.Fill(at(Block(V(-hw+t, 0.2, -4.6), V(hw-t, gf, -3.4))))
		g.Fill(at(Block(V(-hw, gf-0.8, -4.6), V(hw, gf, -0.6))))
		g.Carve(at(Block(V(-hw-0.1, 0.3, -3.4), V(hw+0.1, gf-0.9, -1.2))))
		for j := 0; j < na; j++ {
			x := -hw + ab*(float64(j)+0.5)
			g.Carve(at(Arch{V(x, 0.3, -0.6), ahw, gf - 1.2 - ahw, 1.8, AxisZ}))
			g.Paint(Wood, at(Block(V(x-0.6, 0.3, -3.8), V(x+0.6, 2.6, -3.3))))
		}
	} else {
		door := -hw + bay*0.5
		if Hash01(h.seed, 3) < 0.5 {
			door = hw - bay*0.5
		}
		g.Carve(at(Arch{V(door, 0.3, -0.2), 0.75, 2.0, 1.0, AxisZ}))
		g.Paint(Wood, at(Block(V(door-0.8, 0.3, -1.5), V(door+0.8, 3.0, -0.6))))
		for j := 0; j < n; j++ {
			x := -hw + bay*(float64(j)+0.5)
			if math.Abs(x-door) < 0.1 {
				continue
			}
			h.window(g, x, 1.1, 3.3, 0.6, 0, 1, Glass, false, false)
		}
	}

	// Roof and gables. top(x, z) gives the roof surface for chimneys.
	var top func(x, z float64) float64
	switch h.style {
	case pragueOldTownVolute, pragueOldTownStepped:
		top = func(x, z float64) float64 { return e + k*(hw-math.Abs(x)) }
		g.With(h.roof, func() {
			g.Fill(at(Func{Box{V(-hw, e-0.3, -d-0.4), V(hw, e+k*hw, -0.8)}, func(p Vec3) bool {
				y := top(p.X, p.Z)
				return p.Y <= y && p.Y >= y-th
			}}))
		})
		g.Fill(at(Func{Box{V(-hw, e, -d), V(hw, e+k*hw, -d+t)}, func(p Vec3) bool {
			return p.Y <= top(p.X, p.Z)
		}}))
		roofHalf := func(y float64) float64 { return hw - y/k + 0.15 }
		var half func(y float64) float64
		var H float64
		if h.style == pragueOldTownVolute {
			H = k*hw + 0.9
			s1, nar := H*0.5, hw*0.42
			half = func(y float64) float64 {
				var v float64
				switch {
				case y < 0 || y > H:
					return -1
				case y < s1:
					q := y / s1
					v = nar + (hw-nar)*math.Sqrt(math.Max(0, 1-q*q))
				case y < H*0.78:
					v = nar
				default:
					v = nar*(1-(y-H*0.78)/(H*0.22)) + 0.2
				}
				return math.Max(v, roofHalf(y))
			}
			g.With(h.trim, func() {
				g.Fill(at(Block(V(-nar-0.3, e+s1-0.1, -0.4), V(nar+0.3, e+s1+0.35, 0.45))))
				g.Fill(at(Block(V(-nar-0.3, e+H*0.78-0.4, -0.4), V(nar+0.3, e+H*0.78, 0.45))))
				for _, sx := range []float64{-1, 1} {
					g.Fill(at(Sphere(V(sx*(nar+0.35), e+s1+0.6, -0.3), 0.6)))
					g.Fill(at(Frustum{P(sx*(hw-0.45), -0.3), 0.45, 0.1, e, e + 1.4}))
				}
			})
		} else {
			H = k*hw + 1.2
			sh := H / 4
			half = func(y float64) float64 {
				if y < 0 || y > H {
					return -1
				}
				return hw * (1 - math.Floor(y/sh)/4)
			}
		}
		g.Fill(at(Func{Box{V(-hw, e, -t), V(hw, e+H, 0)}, func(p Vec3) bool {
			return math.Abs(p.X) <= half(p.Y-e)
		}}), at(Func{Box{V(-hw, e, -d), V(hw, e+H, -d+t)}, func(p Vec3) bool {
			return math.Abs(p.X) <= half(p.Y-e)
		}}))
		g.Carve(at(PragueOldTown{}.disk(0, e+H*0.6, 0, 0.65, -d-0.3, -d+0.55)))
		g.Paint(Charcoal, at(PragueOldTown{}.disk(0, e+H*0.6, 0, 0.65, -d+0.5, -d+1.4)))
		// Gable windows: a round light high up and a pair of attic windows.
		g.Carve(at(PragueOldTown{}.disk(0, e+H*0.6, 0, 0.65, -0.55, 0.3)))
		g.Paint(Charcoal, at(PragueOldTown{}.disk(0, e+H*0.6, 0, 0.65, -1.4, -0.5)))
		if hw*0.45+0.6 < half(2.4) {
			for _, sx := range []float64{-1, 1} {
				h.window(g, sx*hw*0.42, e+0.6, e+2.2, 0.45, 0, 1, Charcoal, false, false)
			}
		}
		g.With(Gold, func() { g.Fill(at(Sphere(V(0, e+H+0.4, -0.6), 0.45))) })
	case pragueOldTownDormers, pragueOldTownVenetian:
		rd, zc := d/2+0.4, -d/2
		top = func(x, z float64) float64 { return e + k*(rd-math.Abs(z-zc)) }
		g.With(h.roof, func() {
			g.Fill(at(Func{Box{V(-hw, e-0.3, -d-0.4), V(hw, e+k*rd, 0.4)}, func(p Vec3) bool {
				y := top(p.X, p.Z)
				return p.Y <= y && p.Y >= y-th
			}}))
		})
		g.Fill(at(Func{Box{V(-hw, e, -d), V(hw, e+k*rd, 0)}, func(p Vec3) bool {
			return math.Abs(p.X) >= hw-t && p.Y <= top(p.X, p.Z)-0.4
		}}))
		if h.style == pragueOldTownVenetian {
			m := math.Max(2, math.Round(h.w/1.8))
			p := h.w / m
			g.Fill(at(Block(V(-hw, e, -t), V(hw, e+1.6, 0))))
			g.Fill(at(Func{Box{V(-hw, e+1.6, -t), V(hw, e+1.6+p/2, 0)}, func(q Vec3) bool {
				j := math.Floor((q.X + hw) / p)
				cx := -hw + p*(j+0.5)
				return math.Hypot(q.X-cx, q.Y-e-1.6) <= p/2-0.12
			}}))
			g.Paint(h.trim, at(Block(V(-hw, e+1.2, -0.4), V(hw, e+1.6, 0.1))))
		} else {
			nd := int(math.Max(1, math.Round(h.w/3.2)))
			db := h.w / float64(nd)
			ye := e + 1.0
			zs := zc + rd - 1.0/k
			for j := 0; j < nd; j++ {
				x := -hw + db*(float64(j)+0.5)
				g.Fill(at(Block(V(x-0.95, ye, zs-2.6), V(x+0.95, ye+2.2, zs))))
				g.Paint(Charcoal, at(Block(V(x-0.45, ye+0.4, zs-0.45), V(x+0.45, ye+1.8, zs+0.1))))
				g.With(h.roof, func() {
					g.Fill(at(Func{Box{V(x-1.25, ye+2.0, zs-2.6), V(x+1.25, ye+3.6, zs+0.3)}, func(p Vec3) bool {
						return p.Y <= ye+2.1+1.2*(1.25-math.Abs(p.X-x))
					}}))
				})
			}
		}
	case pragueOldTownTowerHouse:
		top = func(x, z float64) float64 { return e + 1.4 }
		g.Fill(at(Block(V(-hw, e, -d), V(hw, e+1.4, 0))))
		g.Carve(at(Block(V(-hw+t, e-0.2, -d+t), V(hw-t, e+1.6, -t))))
		g.Carve(at(Func{Box{V(-hw, e+0.6, -d), V(hw, e+1.5, 0)}, func(p Vec3) bool {
			return math.Mod(p.X+p.Z+100, 1.6) < 0.7
		}}))
		g.With(h.roof, func() {
			g.Fill(at(Func{Box{V(-hw+t, e, -d+t), V(hw-t, e+hw, -t)}, func(p Vec3) bool {
				y := e + 0.6 + 1.1*math.Min(hw-t-math.Abs(p.X), d/2-t-math.Abs(p.Z+d/2))
				return p.Y <= y && p.Y >= y-1.6
			}}))
		})
	}

	// Chimneys.
	nc := 1 + int(Hash01(h.seed, 7)*2)
	colors := []Color{White, DarkStone, Brick}
	for c := 0; c < nc; c++ {
		cx := (Hash01(h.seed, 11, c) - 0.5) * (h.w - 2.6)
		cz := -d * (0.35 + 0.4*Hash01(h.seed, 13, c))
		y1 := top(cx, cz) + 1.4
		g.With(colors[int(Hash01(h.seed, 17, c)*3)], func() {
			g.Fill(at(Block(V(cx-0.5, e+0.5, cz-0.5), V(cx+0.5, y1, cz+0.5))))
		})
		g.With(Charcoal, func() {
			g.Fill(at(Block(V(cx-0.65, y1, cz-0.65), V(cx+0.65, y1+0.45, cz+0.65))))
		})
	}
}

func (t PragueOldTown) ground(g *Grid) {
	g.Use(DarkStone).Fill(Block(V(-40, -0.2, -60), V(35, 0.2, 33)))
	g.Paint(Stone, Func{Box{V(-40, -0.3, -60), V(35, 0.3, 33)}, func(p Vec3) bool {
		return Hash01(int(math.Round(p.X*4)), int(math.Round(p.Z*4)), 5) < 0.42
	}})
	// Light paving rings around the Hus monument.
	g.Paint(Sand, Func{Box{V(-20, -0.3, -16), V(10, 0.3, 10)}, func(p Vec3) bool {
		r := math.Hypot((p.X+5)/1.4, p.Z+3)
		return math.Mod(r, 4) < 0.45 && r > 5 && r < 14
	}})
	// Churchyard behind the Týn school.
	g.Paint(Stone, Block(V(-12, -0.3, -60), V(12, 0.3, -26)))
	// The former monastery garden behind St Nicholas.
	g.Paint(Grass, Block(V(-39, -0.3, -59), V(-22, 0.3, -39)))
	for i := 0; i < 7; i++ {
		q := P(-36+float64(i%3)*6+3*Hash01(i, 51), -43-float64(i/3)*6-2*Hash01(i, 53))
		t.tree(g, q, 6+2*Hash01(i, 55), i+20)
	}
}

func (PragueOldTown) tree(g *Grid, q Vec2, h float64, seed int) {
	g.With(Wood, func() { g.FillVolume(Cylinder{q, 0.55, 0, 0.2, h * 0.5}) })
	g.With(Foliage, func() { g.FillVolume(Ellipsoid{q.At(h * 0.68), V(2.4, 2.2, 2.4)}) })
	g.Paint(Grass, Func{Box{V(q.X-2.5, h*0.4, q.Y-2.5), V(q.X+2.5, h, q.Y+2.5)}, func(p Vec3) bool {
		return p.Y > h*0.7 && Hash01(int(math.Round(p.X*4)), int(math.Round(p.Y*4)), int(math.Round(p.Z*4)), seed) < 0.4
	}})
}

func (t PragueOldTown) tynChurch(g *Grid) {
	const (
		zf   = -28.0 // Facade face.
		zb   = -52.0 // Start of the apse.
		W    = 8.0   // Outer half-width.
		eave = 22.0
		k    = 1.8
		tw   = 1.4
	)
	rh := W + 0.5
	th := 0.85 * math.Sqrt(1+k*k)
	// Nave and polygonal apse.
	apse := func(r float64, y0, y1 float64) Prism {
		poly := []Vec2{}
		for i := 0; i <= 4; i++ {
			a := math.Pi + float64(i)*math.Pi/4
			poly = append(poly, P(r*math.Cos(a)/math.Cos(math.Pi/8), zb+r*math.Sin(a)/math.Cos(math.Pi/8)))
		}
		return Prism{poly, y0, y1}
	}
	g.Use(Stone)
	g.Fill(Block(V(-W, 0.2, zb), V(W, eave, zf)), apse(W, 0.2, eave))
	g.Carve(Block(V(-W+tw, 0.4, zb), V(W-tw, eave+0.5, zf-tw)), apse(W-tw, 0.4, eave+0.5))
	// Roof: a ridge over the nave that hips round the apse.
	dist := func(p Vec3) float64 {
		if p.Z >= zb {
			return math.Abs(p.X)
		}
		return math.Hypot(p.X, p.Z-zb)
	}
	g.With(Slate, func() {
		g.Fill(Func{Box{V(-rh, eave-0.3, zb-rh), V(rh, eave+k*rh, zf-0.5)}, func(p Vec3) bool {
			y := eave + k*(rh-dist(p))
			return p.Y <= y && p.Y >= y-th && p.Y >= eave-0.3
		}})
	})
	// Front gable.
	g.Fill(Func{Box{V(-W, eave, zf-tw), V(W, eave+k*W+0.3, zf)}, func(p Vec3) bool {
		return p.Y <= eave+k*(W-math.Abs(p.X))+0.3
	}})
	front := pragueOldTownFrame{P(0, zf), math.Pi / 2}
	g.Carve(front.at(t.lancet(0, 2.0, 7, 18, -0.7, 0.3)))
	g.Paint(Navy, front.at(t.lancet(0, 2.0, 7, 18, -1.6, -0.6)))
	g.With(Stone, func() { g.Fill(front.at(Block(V(-0.25, 7, -1.0), V(0.25, 21, -0.3)))) })
	// Madonna niche and gable pinnacles.
	g.Carve(front.at(t.lancet(0, 0.9, 29.5, 31.5, -0.6, 0.3)))
	g.With(Gold, func() {
		g.Fill(front.at(Ellipsoid{V(0, 31.0, -0.4), V(0.6, 1.4, 0.45)}))
		g.Fill(front.at(Capsule{V(0, eave+k*W, -0.7), V(0, eave+k*W+2.4, -0.7), 0.42}))
	})
	for _, sx := range []float64{-1, 1} {
		x := sx * 2.6
		y := eave + k*(W-2.6)
		g.With(Charcoal, func() { g.FillVolume(front.at(Frustum{P(x, -0.7), 0.6, 0, y - 0.5, y + 2.6})) })
		g.With(Gold, func() { g.Fill(front.at(Capsule{V(x, y+2.2, -0.7), V(x, y+3.2, -0.7), 0.4})) })
	}
	// Side buttresses and lancets.
	for _, sx := range []float64{-1, 1} {
		side := pragueOldTownFrame{P(sx*W, 0), 0}
		if sx < 0 {
			side.theta = math.Pi
		}
		for z := zb + 1.0; z <= zf-4; z += 5 {
			bz := z
			g.Use(Stone)
			g.Fill(Block(V(sx*W, 0.2, bz-0.7), V(sx*(W+1.8), 15, bz+0.7)))
			g.Fill(Block(V(sx*W, 15, bz-0.7), V(sx*(W+1.0), 19.5, bz+0.7)))
			g.With(Charcoal, func() { g.FillVolume(Frustum{P(sx*(W+0.6), bz), 0.6, 0, 19.5, 22.5}) })
			if bz+2.5 < zf-1.5 {
				// Local X along the wall maps to world -Z (+X side) or +Z (-X side).
				lx := -(bz + 2.5) * sx
				g.Carve(side.at(t.lancet(lx, 1.2, 6, 15, -0.7, 0.3)))
				g.Paint(Navy, side.at(t.lancet(lx, 1.2, 6, 15, -1.6, -0.6)))
			}
		}
	}
	// Apse buttresses and windows.
	for i := 0; i < 4; i++ {
		a := math.Pi + (float64(i)+0.5)*math.Pi/4
		r := W
		f := pragueOldTownFrame{Polar(P(0, zb), r, a), a}
		g.Carve(f.at(t.lancet(0, 1.1, 6, 14.5, -0.7, 0.3)))
		g.Paint(Navy, f.at(t.lancet(0, 1.1, 6, 14.5, -1.6, -0.6)))
		if i > 0 {
			b := math.Pi + float64(i)*math.Pi/4
			fb := pragueOldTownFrame{Polar(P(0, zb), W/math.Cos(math.Pi/8), b), b}
			g.Use(Stone).Fill(fb.at(Block(V(-0.7, 0.2, -0.5), V(0.7, 15, 1.6))))
			g.With(Charcoal, func() { g.FillVolume(fb.at(Frustum{P(0, 0.5), 0.6, 0, 15, 18})) })
		}
	}

	// Twin towers with dark spires ringed by gold-tipped turrets.
	for _, sx := range []float64{-1, 1} {
		c := P(sx*7.5, zf-3)
		g.Use(DarkStone)
		g.Fill(BlockAt(c, 6, 6, 0.2, 32))
		g.Carve(BlockAt(c, 3.4, 3.4, 0.4, 31.6))
		for _, dx := range []float64{-1, 1} {
			for _, dz := range []float64{-1, 1} {
				q := c.Add(P(dx*2.9, dz*2.9))
				g.Fill(BlockAt(q, 1.8, 1.8, 0.2, 12), BlockAt(q, 1.3, 1.3, 12, 24))
			}
		}
		g.With(Stone, func() {
			for _, y := range []float64{10, 20, 30.4} {
				g.Fill(Subtract(BlockAt(c, 6.7, 6.7, y, y+0.6), BlockAt(c, 5, 5, y-1, y+1)))
			}
		})
		for i := 0; i < 4; i++ {
			a := float64(i) * math.Pi / 2
			f := pragueOldTownFrame{Polar(c, 3, a), a}
			for _, x := range []float64{-1.0, 1.0} {
				g.Carve(f.at(t.lancet(x, 0.55, 22.5, 26.8, -1.6, 0.4)))
			}
			g.Carve(f.at(t.lancet(0, 0.5, 13, 16.5, -0.7, 0.3)))
			g.Paint(Charcoal, f.at(t.lancet(0, 0.5, 13, 16.5, -1.6, -0.6)))
		}
		// Gallery parapet.
		g.With(Stone, func() {
			g.Fill(Subtract(BlockAt(c, 7.0, 7.0, 31.4, 33.0), BlockAt(c, 5.8, 5.8, 31.8, 34)))
		})
		// Main spire.
		const sy0, sy1, sr = 32.0, 53.0, 2.9
		g.With(Charcoal, func() {
			g.Fill(Func{Box{V(c.X-sr, sy0, c.Y-sr), V(c.X+sr, sy1, c.Y+sr)}, func(p Vec3) bool {
				r := sr * (1 - (p.Y-sy0)/(sy1-sy0))
				m := math.Max(math.Abs(p.X-c.X), math.Abs(p.Z-c.Y))
				return m <= r && m >= r-1.1
			}})
		})
		turret := func(q Vec2, r, y0, y1, cone float64) {
			g.With(Slate, func() { g.FillVolume(Cylinder{q, r, 0, y0, y1}) })
			g.With(Charcoal, func() { g.FillVolume(Frustum{q, r + 0.25, 0, y1, y1 + cone}) })
			g.With(Gold, func() {
				g.Fill(Capsule{q.At(y1 + cone - 0.6), q.At(y1 + cone + 1.0), 0.42})
			})
		}
		for _, dx := range []float64{-1, 1} {
			for _, dz := range []float64{-1, 1} {
				turret(c.Add(P(dx*3.0, dz*3.0)), 1.05, 29, 36, 5)
			}
		}
		for i := 0; i < 4; i++ {
			a := float64(i) * math.Pi / 2
			turret(Polar(c, 3.1, a), 0.75, 31, 34.5, 3.6)
			turret(Polar(c, 1.75, a+math.Pi/4), 0.6, 39.5, 41, 3.2)
		}
		g.With(Gold, func() {
			g.Fill(Sphere(c.At(53.2), 0.7))
			g.Fill(Capsule{c.At(53.2), c.At(55.6), 0.4})
		})
	}
}

func (t PragueOldTown) stNicholas(g *Grid) {
	const cx, zf, zb = -30.0, -20.0, -37.0
	g.Use(White)
	g.Fill(Block(V(cx-6, 0.2, zb), V(cx+6, 14, zf)))
	g.Carve(Block(V(cx-4.8, 0.4, zb+1.2), V(cx+4.8, 14.5, zf-1.2)))
	g.With(Sand, func() {
		g.Fill(Block(V(cx-6.3, 13.3, zb-0.3), V(cx+6.3, 14.1, zf+0.3)))
		g.Fill(Block(V(cx-6.3, 4.6, zb-0.3), V(cx+6.3, 5.1, zf+0.3)))
	})
	// Pilasters and tall windows on the side walls.
	for _, sx := range []float64{-1, 1} {
		side := pragueOldTownFrame{P(cx+sx*6, 0), 0}
		if sx < 0 {
			side.theta = math.Pi
		}
		for z := zb + 1.5; z <= zf-1.5; z += 3.4 {
			g.With(Sand, func() { g.Fill(Block(V(cx+sx*6, 0.2, z-0.4), V(cx+sx*6.4, 13.3, z+0.4))) })
			wz := z + 1.7
			if wz < zf-1.5 {
				h := pragueOldTownHouse{f: side, trim: Sand}
				h.window(g, -wz*sx, 6.5, 11.5, 0.7, 0, 1, Charcoal, true, true)
				h.window(g, -wz*sx, 1.6, 3.6, 0.6, 0, 1, Charcoal, true, false)
			}
		}
	}
	// Roof.
	g.With(Slate, func() {
		g.Fill(Func{Box{V(cx-6.4, 13.8, zb-0.4), V(cx+6.4, 20, zf)}, func(p Vec3) bool {
			y := 14 + 1.0*(6.4-math.Abs(p.X-cx))
			return p.Y <= y && p.Y >= y-1.6
		}})
	})
	// Projecting central bay with a segmental pediment and portal.
	g.Fill(Block(V(cx-3.4, 0.2, zf), V(cx+3.4, 17, zf+1.0)))
	g.Fill(Func{Box{V(cx-3.4, 17, zf), V(cx+3.4, 19.4, zf+1.0)}, func(p Vec3) bool {
		x, y := (p.X-cx)/3.6, (p.Y-17)/2.4
		return x*x+y*y <= 1
	}})
	g.With(Sand, func() {
		g.Fill(Block(V(cx-3.7, 16.4, zf-0.2), V(cx+3.7, 17.0, zf+1.5)))
		g.Fill(Func{Box{V(cx-3.8, 17, zf+0.6), V(cx+3.8, 20, zf+1.4)}, func(p Vec3) bool {
			x, y := (p.X-cx)/3.8, (p.Y-17)/2.6
			r := x*x + y*y
			return r <= 1 && r >= 0.62
		}})
		for _, dx := range []float64{-3.4, -1.8, 1.8, 3.4} {
			g.Fill(Block(V(cx+dx-0.35, 0.2, zf+0.8), V(cx+dx+0.35, 16.4, zf+1.4)))
		}
	})
	fr := pragueOldTownFrame{P(cx, zf+1.0), math.Pi / 2}
	g.Carve(fr.at(Arch{V(0, 0.3, -0.5), 1.2, 2.6, 1.6, AxisZ}))
	g.Paint(Wood, fr.at(Block(V(-1.2, 0.3, -2.4), V(1.2, 4, -1.2))))
	g.Carve(fr.at(Arch{V(0, 6.5, -0.3), 1.0, 4.5, 1.2, AxisZ}))
	g.Paint(Charcoal, fr.at(Arch{V(0, 6.5, -1.3), 1.0, 4.5, 1.0, AxisZ}))
	g.Carve(fr.at(t.disk(0, 14.6, 0, 0.8, -0.6, 0.3)))
	g.Paint(Charcoal, fr.at(t.disk(0, 14.6, 0, 0.8, -1.4, -0.5)))
	// Statues on the pediment.
	g.With(White, func() {
		for _, dx := range []float64{-3.4, 0, 3.4} {
			y := 17.0
			if dx == 0 {
				y = 19.6
			}
			g.Fill(Capsule{V(cx+dx, y, zf+0.7), V(cx+dx, y+1.4, zf+0.7), 0.45})
		}
	})
	// Corner towers with green bulbous cupolas.
	for _, sx := range []float64{-1, 1} {
		c := P(cx+sx*6.6, zf-1)
		g.Use(White)
		g.Fill(BlockAt(c, 3.8, 3.8, 0.2, 21))
		g.Carve(BlockAt(c, 1.6, 1.6, 0.4, 20.6))
		g.With(Sand, func() {
			for _, dx := range []float64{-1, 1} {
				for _, dz := range []float64{-1, 1} {
					g.Fill(BlockAt(c.Add(P(dx*1.75, dz*1.75)), 0.8, 0.8, 0.2, 21))
				}
			}
			g.Fill(BlockAt(c, 4.4, 4.4, 13.3, 14.1), BlockAt(c, 4.4, 4.4, 20.6, 21.4))
		})
		for i := 0; i < 4; i++ {
			a := float64(i) * math.Pi / 2
			f := pragueOldTownFrame{Polar(c, 1.9, a), a}
			g.Carve(f.at(Arch{V(0, 15.5, -0.6), 0.6, 2.6, 2.0, AxisZ}))
		}
		g.With(Grass, func() {
			g.FillVolume(Intersect(Ellipsoid{c.At(22.6), V(2.0, 2.5, 2.0)}, BlockAt(c, 5, 5, 21.4, 26)))
			g.FillVolume(Frustum{c, 0.9, 0.3, 24.6, 26.0})
			g.FillVolume(Ellipsoid{c.At(26.6), V(0.85, 0.9, 0.85)})
		})
		g.With(Gold, func() { g.Fill(Capsule{c.At(27.2), c.At(29.4), 0.42}) })
	}
	// Dome on an octagonal drum.
	dc := P(cx, -29)
	g.Use(White)
	g.Fill(Cylinder{dc, 5, 3.8, 13.8, 20.2})
	for i := 0; i < 8; i++ {
		a := float64(i)*math.Pi/4 + math.Pi/8
		f := pragueOldTownFrame{Polar(dc, 5, a), a}
		g.Carve(f.at(Arch{V(0, 15.2, -0.6), 0.75, 2.4, 2.2, AxisZ}))
		g.With(Sand, func() {
			g.Fill(RotateY(BlockAt(Polar(dc, 5.1, a+math.Pi/8), 0.8, 0.8, 13.8, 20.2), Polar(dc, 5.1, a+math.Pi/8), a))
		})
	}
	g.With(Sand, func() { g.Fill(Cylinder{dc, 5.5, 4.0, 20.2, 20.9}) })
	g.With(Grass, func() {
		g.FillVolume(Subtract(Ellipsoid{dc.At(20.9), V(5.1, 5.6, 5.1)},
			Ellipsoid{dc.At(20.9), V(3.9, 4.4, 3.9)}, BlockAt(dc, 12, 12, 10, 20.9)))
	})
	g.Paint(Foliage, Func{Box{V(dc.X-5.3, 21, dc.Y-5.3), V(dc.X+5.3, 27, dc.Y+5.3)}, func(p Vec3) bool {
		a := math.Atan2(p.Z-dc.Y, p.X-dc.X)
		s := math.Mod(a+20*math.Pi, math.Pi/4)
		return math.Min(s, math.Pi/4-s)*math.Hypot(p.X-dc.X, p.Z-dc.Y) < 0.35
	}})
	g.With(White, func() { g.FillVolume(Cylinder{dc, 1.4, 0, 25.6, 28}) })
	for i := 0; i < 4; i++ {
		a := float64(i) * math.Pi / 2
		f := pragueOldTownFrame{Polar(dc, 1.4, a), a}
		g.Carve(f.at(Arch{V(0, 26.0, -0.5), 0.45, 1.1, 1.4, AxisZ}))
	}
	g.With(Grass, func() {
		g.FillVolume(Intersect(Ellipsoid{dc.At(28), V(1.7, 1.6, 1.7)}, BlockAt(dc, 4, 4, 28, 30)))
	})
	g.With(Gold, func() { g.Fill(Capsule{dc.At(29.4), dc.At(31.8), 0.42}) })
}

func (t PragueOldTown) townHall(g *Grid) {
	c := P(25.5, 6.5)
	const s = 6.5
	g.Use(Stone)
	g.Fill(BlockAt(c, s, s, 0.2, 32))
	g.Carve(BlockAt(c, s-2.6, s-2.6, 0.4, 31.6))
	g.With(DarkStone, func() {
		for _, dx := range []float64{-1, 1} {
			for _, dz := range []float64{-1, 1} {
				g.Fill(BlockAt(c.Add(P(dx*(s/2-0.3), dz*(s/2-0.3))), 0.8, 0.8, 0.2, 31))
			}
		}
		for _, y := range []float64{2.6, 19.6} {
			g.Fill(Subtract(BlockAt(c, s+0.5, s+0.5, y, y+0.6), BlockAt(c, s-1, s-1, y-1, y+1)))
		}
	})
	// Lancets on every face, and the chapel oriel facing the square (-X).
	for i := 0; i < 4; i++ {
		a := float64(i) * math.Pi / 2
		f := pragueOldTownFrame{Polar(c, s/2, a), a}
		for _, y := range []float64{21.5, 26} {
			g.Carve(f.at(t.lancet(0, 0.6, y, y+2, -1.4, 0.3)))
		}
		if i != 1 {
			g.Carve(f.at(t.lancet(0, 0.55, 14, 16.4, -0.7, 0.3)))
			g.Paint(Charcoal, f.at(t.lancet(0, 0.55, 14, 16.4, -1.6, -0.6)))
		}
	}
	oc := P(c.X-s/2-0.4, c.Y)
	g.With(Stone, func() {
		g.FillVolume(Frustum{oc, 0.4, 1.6, 4.5, 6.5})
		g.FillVolume(Cylinder{oc, 1.6, 0, 6.5, 12})
	})
	of := pragueOldTownFrame{oc.Add(P(-1.6, 0)), math.Pi}
	g.Carve(of.at(t.lancet(0, 0.55, 7.5, 10, -0.6, 0.3)))
	g.Paint(Navy, of.at(t.lancet(0, 0.55, 7.5, 10, -1.4, -0.5)))
	g.With(Charcoal, func() { g.FillVolume(Frustum{oc, 1.9, 0, 12, 16}) })
	g.With(Gold, func() { g.Fill(Capsule{oc.At(15.4), oc.At(16.8), 0.4}) })

	// The astronomical clock on the south face (+Z).
	zf := c.Y + s/2
	f := pragueOldTownFrame{P(c.X, zf), math.Pi / 2}
	g.With(Sand, func() { g.Fill(f.at(Block(V(-2.8, 2.6, 0), V(2.8, 18, 0.5)))) })
	g.With(Charcoal, func() {
		g.Fill(f.at(Func{Box{V(-3.3, 18, -0.2), V(3.3, 21.5, 1.6)}, func(p Vec3) bool {
			return p.Y <= 18.2+1.3*(3.3-math.Abs(p.X)) && p.Y >= 18.2+1.3*(3.3-math.Abs(p.X))-1.3
		}}))
	})
	const ay, cy = 12.7, 6.4
	g.With(Navy, func() { g.Fill(f.at(t.disk(0, ay, 0, 2.4, 0.4, 1.1))) })
	g.Paint(Glass, f.at(Func{Box{V(-2.4, ay, 0.3), V(2.4, ay+2.4, 1.2)}, func(p Vec3) bool {
		return p.Y > ay+0.5
	}}))
	g.Paint(Charcoal, f.at(Func{Box{V(-2.4, ay-2.4, 0.3), V(2.4, ay, 1.2)}, func(p Vec3) bool {
		return p.Y < ay-1.0
	}}))
	g.Paint(Gold, f.at(t.disk(0, ay, 1.95, 2.5, 0.3, 1.2)), f.at(t.disk(0, ay+0.55, 1.0, 1.35, 0.3, 1.2)),
		f.at(t.disk(0, ay, 0, 0.45, 0.3, 1.2)))
	g.With(Gold, func() {
		g.Fill(f.at(Capsule{V(0, ay, 1.3), V(1.3, ay+1.0, 1.3), 0.3}))
		g.Fill(f.at(Capsule{V(0, ay, 1.3), V(-0.5, ay-1.5, 1.3), 0.3}))
	})
	g.With(Gold, func() { g.Fill(f.at(t.disk(0, cy, 0, 2.1, 0.4, 1.0))) })
	g.Paint(White, f.at(t.disk(0, cy, 1.15, 1.65, 0.3, 1.1)))
	g.Paint(Navy, f.at(t.disk(0, cy, 0.5, 1.15, 0.3, 1.1)))
	g.Paint(Charcoal, f.at(Block(V(-1.6, 15.8, 0.3), V(-0.6, 17.2, 0.6))), f.at(Block(V(0.6, 15.8, 0.3), V(1.6, 17.2, 0.6))))
	g.With(Stone, func() {
		for _, sx := range []float64{-1, 1} {
			g.Fill(f.at(Capsule{V(sx*2.3, 9.6, 0.9), V(sx*2.3, 10.8, 0.9), 0.42}))
		}
	})

	// Gallery with corner oriels and a steep pointed roof.
	g.With(DarkStone, func() { g.Fill(BlockAt(c, s+0.8, s+0.8, 31, 32)) })
	g.With(Stone, func() {
		g.Fill(Subtract(BlockAt(c, s+0.8, s+0.8, 32, 33.2), BlockAt(c, s-0.4, s-0.4, 31, 34)))
		g.Carve(Func{Box{V(c.X-s, 32.4, c.Y-s), V(c.X+s, 33.0, c.Y+s)}, func(p Vec3) bool {
			return math.Mod(p.X+p.Z+100, 1.4) < 0.6
		}})
	})
	for _, dx := range []float64{-1, 1} {
		for _, dz := range []float64{-1, 1} {
			q := c.Add(P(dx*(s/2+0.2), dz*(s/2+0.2)))
			g.With(Stone, func() {
				g.FillVolume(Frustum{q, 0.3, 1.05, 28.5, 30.5})
				g.FillVolume(Cylinder{q, 1.05, 0, 30.5, 34})
			})
			g.With(Charcoal, func() { g.FillVolume(Frustum{q, 1.3, 0, 34, 38}) })
			g.With(Gold, func() { g.Fill(Capsule{q.At(37.4), q.At(38.8), 0.42}) })
		}
	}
	const ry0, ry1, rr = 32.0, 43.5, 3.0
	g.With(Charcoal, func() {
		g.Fill(Func{Box{V(c.X-rr, ry0, c.Y-rr), V(c.X+rr, ry1, c.Y+rr)}, func(p Vec3) bool {
			r := rr * (1 - (p.Y-ry0)/(ry1-ry0))
			m := math.Max(math.Abs(p.X-c.X), math.Abs(p.Z-c.Y))
			return m <= r && m >= r-1.1
		}})
	})
	g.With(Gold, func() {
		g.Fill(Sphere(c.At(43.8), 0.65))
		g.Fill(Capsule{c.At(43.8), c.At(45.6), 0.4})
	})
}

func (t PragueOldTown) rows(g *Grid) {
	type spec struct {
		w, eave          float64
		wall, trim, roof Color
		style            int
		arcade           bool
		side             int
	}
	place := func(c Vec2, theta, d float64, s spec, seed int, sgraffito bool) {
		h := pragueOldTownHouse{
			f: pragueOldTownFrame{c, theta}, w: s.w, d: d, eave: s.eave,
			wall: s.wall, trim: s.trim, roof: s.roof, style: s.style, arcade: s.arcade,
			sgraffito: sgraffito, sideWin: s.side, seed: seed,
		}
		h.build(g)
	}
	const (
		V0 = pragueOldTownVolute
		D0 = pragueOldTownDormers
		N0 = pragueOldTownVenetian
		S0 = pragueOldTownStepped
		T0 = pragueOldTownTowerHouse
	)
	// West row: facades face +X at x = -26, from z = -12 to 22. Local +X is world -Z.
	west := []spec{
		{6, 13.6, Gold, White, Terracotta, V0, true, 0},
		{5.5, 12.4, Glass, White, Brick, D0, false, 0},
		{6.5, 14.2, White, Sand, Terracotta, S0, true, 0},
		{5.5, 13.0, Red, White, Terracotta, V0, false, 0},
		{5, 12.0, Sand, White, Brick, N0, true, 0},
		{5.5, 13.4, White, Gold, Terracotta, D0, true, 2},
	}
	z := 22.0
	for i, s := range west {
		place(P(-26, z-s.w/2), 0, 8, s, 100+i, false)
		z -= s.w
	}
	// South row: facades face -Z at z = 22, from x = -34 to -2. Local +X is world -X.
	south := []spec{
		{8, 12.6, Sand, White, Terracotta, D0, false, 0},
		{6, 13.8, White, Sand, Brick, V0, true, 0},
		{5.5, 12.8, Gold, White, Terracotta, S0, false, 0},
		{6.5, 14.0, Glass, White, Terracotta, V0, true, 0},
		{6, 12.2, Red, White, Slate, D0, true, 1},
	}
	x := -34.0
	for i, s := range south {
		place(P(x+s.w/2, 22), -math.Pi/2, 8, s, 200+i, false)
		x += s.w
	}
	// North row: facades face +Z at z = -16, from x = -20 to 22. Local +X is world +X.
	north := []spec{
		{5, 13.2, Glass, White, Terracotta, V0, false, 1},
		{5.5, 16.5, Stone, DarkStone, Charcoal, T0, false, 0},
		{8.5, 13.8, White, Sand, Terracotta, N0, true, 0},
		{8, 14.4, Sand, White, Terracotta, V0, true, 0},
		{15, 15.0, Terracotta, White, Slate, D0, false, 0},
	}
	x = -20.0
	for i, s := range north {
		d := 8.0
		if i == 4 {
			d = 10
		}
		place(P(x+s.w/2, -16), math.Pi/2, d, s, 300+i, false)
		x += s.w
	}
	// The Stone Bell on its corner.
	g.With(Gold, func() { g.Fill(Sphere(V(-9.9, 6.0, -15.6), 0.75)) })
	// Kinsky palace pediments and statues.
	g.With(White, func() {
		for _, px := range []float64{10.5, 18.5} {
			g.Fill(Func{Box{V(px-2.6, 15, -16.8), V(px+2.6, 17.4, -15.6)}, func(p Vec3) bool {
				return p.Y <= 15+0.9*(2.6-math.Abs(p.X-px))
			}})
		}
		for _, px := range []float64{7.6, 12.0, 17.0, 21.4} {
			g.Fill(Capsule{V(px, 15.2, -16.0), V(px, 16.6, -16.0), 0.42})
		}
	})
	// East row (Old Town Hall houses): facades face -X at x = 22, z from 3.25 to -16.
	// Local +X is world +Z.
	east := []spec{
		{6.5, 13.0, White, Sand, Terracotta, N0, true, 0},
		{6.25, 14.6, Stone, Sand, Slate, S0, false, 0},
		{6.5, 13.6, Sand, White, Terracotta, D0, false, 0},
	}
	z = 3.25
	for i, s := range east {
		place(P(22, z-s.w/2), math.Pi, 10, s, 400+i, i == 0)
		z -= s.w
	}

	// Back streets: houses along the lanes beside the Týn church and the Ungelt.
	lane := []spec{
		{6, 11.6, Sand, White, Terracotta, D0, false, 0},
		{5.5, 12.4, White, Sand, Brick, V0, false, 0},
		{6.5, 11.2, Gold, White, Terracotta, D0, false, 0},
		{6, 12.8, Glass, White, Terracotta, S0, false, 0},
		{6.5, 11.8, White, Gold, Brick, D0, false, 0},
	}
	z = -27.0
	for i, s := range lane {
		place(P(13, z-s.w/2), math.Pi, 8, s, 500+i, false)
		z -= s.w
	}
	ungelt := []spec{
		{7, 12.0, White, Sand, Terracotta, D0, false, 0},
		{6, 13.0, Red, White, Terracotta, V0, false, 0},
		{6.5, 11.4, Sand, White, Brick, D0, false, 0},
		{7, 12.6, Gold, White, Terracotta, N0, false, 0},
		{6.5, 11.8, Glass, White, Terracotta, V0, false, 0},
		{7, 12.2, White, Sand, Slate, D0, false, 2},
	}
	z = -17.0
	for i, s := range ungelt {
		place(P(33, z-s.w/2), 0, 8, s, 600+i, false)
		z -= s.w
	}
	west2 := []spec{
		{6, 12.2, Gold, White, Terracotta, V0, false, 0},
		{6.5, 11.4, White, Sand, Terracotta, D0, false, 0},
		{5.5, 12.8, Sand, White, Brick, S0, false, 0},
		{6, 11.8, Glass, White, Terracotta, D0, false, 0},
		{6, 12.4, White, Gold, Terracotta, V0, false, 0},
	}
	z = -26.0
	for i, s := range west2 {
		place(P(-13, z-s.w/2), 0, 8, s, 700+i, false)
		z -= s.w
	}
}

func (t PragueOldTown) monuments(g *Grid) {
	// Jan Hus monument: stepped oval base, rock, and the bronze groups.
	hc := P(-5, -3)
	g.With(DarkStone, func() { g.Fill(EllipticCylinder{hc, 8.6, 5.6, 0.2, 0.9}) })
	g.With(Stone, func() { g.Fill(EllipticCylinder{hc, 7.8, 4.9, 0.9, 1.6}) })
	g.With(DarkStone, func() {
		g.FillVolume(Intersect(Ellipsoid{hc.At(1.4), V(6.8, 3.0, 4.0)}, Block(V(hc.X-7, 1.4, hc.Y-5), V(hc.X+7, 5, hc.Y+5))))
		g.FillVolume(Intersect(Ellipsoid{hc.Add(P(0, -1)).At(3.6), V(2.4, 1.6, 2.0)}, Block(V(hc.X-3, 3.6, hc.Y-4), V(hc.X+3, 5.4, hc.Y+2))))
	})
	g.With(Foliage, func() {
		hus := hc.Add(P(0, -1.0))
		g.FillVolume(Frustum{hus, 1.25, 0.75, 4.6, 10.4})
		g.FillVolume(Sphere(hus.At(11.0), 0.65))
		for i := 0; i < 12; i++ {
			side := -1.0
			if i%2 == 1 {
				side = 1
			}
			j := float64(i / 2)
			q := hc.Add(P(side*(2.2+j*0.95), 1.4+0.6*math.Sin(j*1.7)-0.25*j))
			h := 2.6 + 1.4*Hash01(i, 31) - 0.15*j
			base := 1.6 + (2.2-0.35*j)*Hash01(i, 37)
			g.FillVolume(Frustum{q, 0.7, 0.45, base, base + h})
			g.FillVolume(Sphere(q.At(base+h+0.35), 0.48))
		}
	})
	g.Paint(Grass, Func{Box{V(hc.X-8, 4.4, hc.Y-6), V(hc.X+8, 12, hc.Y+6)}, func(p Vec3) bool {
		c, _ := g.ColorOf(Nearest(p.Scale(1 / g.Span())))
		return c == Foliage && Hash01(int(math.Round(p.X*4)), int(math.Round(p.Y*4)), int(math.Round(p.Z*4))) < 0.3
	}})
	// Marian column.
	mc := P(6, 6)
	g.Use(Stone)
	g.Fill(BlockAt(mc, 2.6, 2.6, 0.2, 1.2), BlockAt(mc, 1.8, 1.8, 1.2, 2.6))
	g.FillVolume(Cylinder{mc, 0.6, 0, 2.6, 12})
	g.Fill(BlockAt(mc, 1.6, 1.6, 12, 12.6))
	g.With(Gold, func() {
		g.FillVolume(Frustum{mc, 0.6, 0.35, 12.6, 14.2})
		g.Fill(Sphere(mc.At(14.6), 0.42))
	})
}

func (t PragueOldTown) market(g *Grid) {
	// Wooden stalls with striped roofs in the open south-east of the square.
	for i := 0; i < 4; i++ {
		for j := 0; j < 2; j++ {
			c := P(3+float64(i)*4.2, 14+float64(j)*4.6)
			g.Use(Wood)
			g.Fill(BlockAt(c, 3.2, 2.4, 0.2, 2.4))
			g.Carve(BlockAt(c.Add(P(0, 0.2)), 2.4, 2.2, 1.2, 2.3))
			roof := Red
			if (i+j)%2 == 1 {
				roof = Navy
			}
			g.With(roof, func() {
				g.Fill(Func{Box{V(c.X-1.9, 2.4, c.Y-1.6), V(c.X+1.9, 3.6, c.Y+1.6)}, func(p Vec3) bool {
					return p.Y <= 2.4+0.9*(1.6-math.Abs(p.Z-c.Y))
				}})
			})
			g.Paint(White, Func{Box{V(c.X-1.9, 2.4, c.Y-1.6), V(c.X+1.9, 3.6, c.Y+1.6)}, func(p Vec3) bool {
				return math.Mod(p.X-c.X+10, 1.2) < 0.5
			}})
		}
	}
	// Trees.
	trees := []Vec2{P(24, 16), P(30, 18), P(26, 23), P(33, 26), P(21, 29), P(29, 30), P(32, 11)}
	for i, q := range trees {
		t.tree(g, q, 6+2*Hash01(i, 41), i)
	}
}

func (t PragueOldTown) Build(g *Grid) Order {
	t.ground(g)
	t.tynChurch(g)
	t.stNicholas(g)
	t.rows(g)
	t.townHall(g)
	t.monuments(g)
	t.market(g)
	return Outward(P(-4, 0))
}
