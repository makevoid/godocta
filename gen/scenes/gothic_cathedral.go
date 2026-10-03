package scenes

import (
	"math"

	"godocta/gen/kit"
	. "godocta/gen/octa"
)

// GothicCathedral is a Notre-Dame-style cathedral on its island. The west facade faces
// +X onto the paved parvis: two square towers with open belfry lancets, three deeply
// recessed portals, the gallery of kings, the great rose and the grand gallery. The
// nave and choir run to the rounded apse at -X with its radiating chapels; flying
// buttresses leap over the double aisles on both sides and around the chevet. The
// transept carries its own rose windows and the slender lead fleche rises over the
// crossing. The Seine runs along -Z below a tree-lined quay.
type GothicCathedral struct{}

func init() { Register(GothicCathedral{}) }

func (GothicCathedral) Name() string  { return "gothic_cathedral" }
func (GothicCathedral) Title() string { return "Gothic Cathedral" }

// lancet is an equilateral pointed opening in a local frame: across X centred on 0,
// passing through Z from z0 to z1, with jambs from sill to spring.
func (GothicCathedral) lancet(hw, sill, spring, z0, z1 float64) Func {
	apex := spring + math.Sqrt(3)*hw
	return Func{Box{V(-hw, sill, z0), V(hw, apex, z1)}, func(p Vec3) bool {
		ax := math.Abs(p.X)
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

// at places a local-frame solid: local +Z points along plan angle theta and the local
// origin lands on c.
func (GothicCathedral) at(s Solid, c Vec2, theta float64) Solid {
	return Translate(RotateY(s, P(0, 0), theta-math.Pi/2), c.At(0))
}

// rose builds a rose window in a local frame whose wall face is z = 0 (outside +Z):
// a carved recess with colored glass, stone a stone ring, alternating red and blue lights and a sand frame.
func (c GothicCathedral) rose(g *Grid, place func(Solid) Solid, cy, R float64, lights int) {
	polar := func(p Vec3) (float64, float64) { return math.Hypot(p.X, p.Y-cy), math.Atan2(p.Y-cy, p.X) }
	box := func(z0, z1, grow float64) Box {
		return Box{V(-R-grow, cy-R-grow, z0), V(R+grow, cy+R+grow, z1)}
	}
	step := 2 * math.Pi / float64(lights)
	g.Carve(place(Func{box(-1.0, 1.5, 0), func(p Vec3) bool { r, _ := polar(p); return r < R }}))
	glass := box(-1.9, -0.9, 0)
	g.With(Navy, func() {
		g.Fill(place(Func{glass, func(p Vec3) bool { r, _ := polar(p); return r < R }}))
	})
	g.Paint(Red, place(Func{glass, func(p Vec3) bool {
		r, a := polar(p)
		sector := int(math.Floor((a + 20*math.Pi) / step))
		return (r > 0.9 && r < 1.8) || (r > 2.9 && sector%2 == 0)
	}}))
	g.Paint(Glass, place(Func{glass, func(p Vec3) bool { r, _ := polar(p); return r > 1.9 && r < 2.5 }}))
	g.Paint(Gold, place(Func{glass, func(p Vec3) bool { r, _ := polar(p); return r <= 0.9 }}))
	// Radial mullions between the outer lights, set in the glass plane.
	g.Paint(Stone, place(Func{glass, func(p Vec3) bool {
		r, a := polar(p)
		d := math.Mod(a+20*math.Pi, step)
		return r > 2.9 && r < R && math.Min(d, step-d)*r < 0.22
	}}))
	// Tracery: a stone ring between the inner and outer lights.
	g.With(Stone, func() {
		g.Fill(place(Func{box(-0.9, -0.2, 0), func(p Vec3) bool {
			r, _ := polar(p)
			if r >= R-0.1 {
				return false
			}
			return r > 2.5 && r < 2.9
		}}))
	})
	g.With(Sand, func() {
		g.Fill(place(Func{box(-0.8, 0.7, 1), func(p Vec3) bool { r, _ := polar(p); return r >= R && r <= R+0.7 }}))
	})
}

// flyer is a flying buttress in a local frame running out along +Z from the clerestory
// (z = w) to its pier (z = pz): a straight rampant over an arched underside.
func (GothicCathedral) flyer(w, pz, top0, top1 float64) Func {
	return Func{Box{V(-0.75, top1-4, w-0.4), V(0.75, top0+0.2, pz)}, func(p Vec3) bool {
		if math.Abs(p.X) > 0.75 {
			return false
		}
		t := (p.Z - w) / (pz - w)
		if t < -0.05 || t > 1 {
			return false
		}
		top := top0 + (top1-top0)*t
		bottom := (top0 - 4.6) + ((top1-3.6)-(top0-4.6))*t + 3.0*4*t*(1-t)
		return p.Y <= top && p.Y >= bottom
	}}
}

// pier is a stepped buttress pier with a pinnacle, in the same local frame as flyer.
func (GothicCathedral) pier(g *Grid, place func(Solid) Solid, z0, z1, top float64) {
	g.Fill(place(Block(V(-0.95, 0, z0), V(0.95, top-6, z1+0.8))))
	g.Fill(place(Block(V(-0.95, 0, z0), V(0.95, top, z1))))
	mid := (z0 + z1) / 2
	g.FillVolume(place(Frustum{P(0, mid+0.2), 1.0, 0.35, top, top + 4.2}))
	g.Fill(place(Block(V(-1.1, top-0.3, z0-0.1), V(1.1, top+0.3, z1+0.1))))
}

func (c GothicCathedral) Build(g *Grid) Order {
	const (
		xf       = 42.0  // West facade face.
		xb       = 32.0  // Back of the tower block.
		xa       = -30.0 // Apse centre.
		xc       = -5.0  // Crossing centre.
		W        = 5.0   // Central vessel outer half-width.
		wt       = 1.2   // Wall thickness.
		A        = 13.5  // Aisle outer wall.
		wallTop  = 22.0  // Clerestory wall top.
		roofK    = 1.5   // Roof slope.
		eave     = 0.6
		zt       = 18.0 // Transept end.
		pierZ0   = 13.3
		pierZ1   = 15.6
		pierTop  = 17.8
		flyTop0  = 21.4
		flyTop1  = 17.6
		aisleIn  = 13.0
		aisleOut = 10.5
	)
	h := W + eave
	ridge := wallTop + roofK*h
	tv := 1.25 * math.Sqrt(1+roofK*roofK)

	axisA := func(p Vec3) float64 {
		if p.X >= xa {
			return math.Abs(p.Z)
		}
		return math.Hypot(p.X-xa, p.Z)
	}
	band := func(a0, a1, y0, y1, xmax float64) Func {
		return Func{Box{V(xa-a1, y0, -a1), V(xmax, y1, a1)}, func(p Vec3) bool {
			a := axisA(p)
			return a >= a0 && a <= a1
		}}
	}
	aisleTop := func(a float64) float64 { return aisleIn - (aisleIn-aisleOut)*(a-W)/(A-W) }

	const xc0 = -5.0 // Crossing centre, for the north transept path.
	// ---------------------------------------------------------------- Site.
	// Ground layer, the Seine to the south (-Z) below a quay, and the far bank.
	g.Use(Grass).Fill(Block(V(-58, -0.2, -27), V(70, 0.2, 36)))
	g.Use(Sand).Fill(Block(V(43.4, -0.2, -27), V(70, 0.2, 36)))
	g.Paint(Stone, Func{Box{V(43.4, -0.3, -27), V(70, 0.3, 36)}, func(p Vec3) bool {
		mx := math.Mod(p.X-43.4+100, 6)
		mz := math.Mod(p.Z+100, 6)
		return mx < 0.5 || mz < 0.5
	}})
	// South promenade and paths.
	g.Paint(Stone, Block(V(-58, -0.3, -27), V(43.4, 0.3, -18)))
	g.Paint(Sand, Block(V(-58, -0.3, -16.5), V(43.4, 0.3, 16.5)))
	g.Paint(Sand, Block(V(-58, -0.3, 25), V(43.4, 0.3, 27)))
	g.Paint(Sand, Block(V(-52, -0.3, -27), V(-49, 0.3, 36)))
	g.Paint(Sand, Block(V(xc0-2, -0.3, 16), V(xc0+2, 0.3, 36)))
	g.Use(Water).Fill(Block(V(-58, -1.35, -42), V(70, -1.05, -27.4)))
	g.Use(DarkStone)
	g.Fill(Block(V(-58, -1.35, -28.2), V(70, 0.2, -27.2)))
	g.Fill(Block(V(-58, -1.35, -42.6), V(70, 0.2, -41.6)))
	g.Use(Stone)
	g.Fill(Block(V(-58, 0.2, -28.2), V(70, 1.1, -27.4)))
	g.Fill(Block(V(-58, 0.2, -42.4), V(70, 1.1, -41.6)))
	g.Carve(Func{Box{V(-58, 0.5, -43), V(70, 1.2, -27)}, func(p Vec3) bool {
		return math.Mod(p.X+100, 3) < 1.2
	}})
	g.Use(Stone).Fill(Block(V(-58, -0.2, -48), V(70, 0.2, -42.4)))
	// Quay walls around the island and the far bank, down to the river.
	g.With(DarkStone, func() {
		g.Fill(Block(V(-59, -1.35, -27.4), V(-58, 0.2, 36)), Block(V(70, -1.35, -27.4), V(71, 0.2, 36)),
			Block(V(-59, -1.35, 36), V(71, 0.2, 37)),
			Block(V(-59, -1.35, -48), V(-58, 0.2, -41.6)), Block(V(70, -1.35, -48), V(71, 0.2, -41.6)),
			Block(V(-59, -1.35, -49), V(71, 0.2, -48)))
	})
	// Bridge to the left bank in front of the parvis: a segmental arch.
	g.Use(Stone).Fill(Subtract(Block(V(58, -1.35, -42), V(64.5, 0.4, -27.4)),
		Func{Box{V(57, -1.4, -42), V(66, -0.3, -27)}, func(p Vec3) bool {
			dz, dy := p.Z+34.7, p.Y+10
			return dz*dz+dy*dy <= 9.6*9.6
		}}))
	g.Fill(Block(V(58, 0.4, -42), V(58.7, 1.4, -27.4)), Block(V(63.8, 0.4, -42), V(64.5, 1.4, -27.4)))
	g.Paint(Sand, Block(V(58.7, 0.0, -42), V(63.8, 0.5, -27.4)))

	// Trees: north garden, square behind the apse, quays.
	trees := []Vec2{}
	for x := -46.0; x <= 36; x += 6.5 {
		trees = append(trees, P(x+3.2, 30.5))
	}
	for x := -56.0; x <= -48; x += 5 {
		for z := -14.0; z <= 22; z += 6 {
			trees = append(trees, P(x+2*Hash01(int(x), int(z)), z))
		}
	}
	for x := -54.0; x <= 40; x += 7 {
		trees = append(trees, P(x, -24.5))
	}
	for x := -52.0; x <= 74; x += 8 {
		if x < 54 || x > 68 {
			trees = append(trees, P(x+3, -45.5))
		}
	}
	for i, t := range trees {
		g.Build(kit.Tree{At: t, Base: 0.2, Height: 5.5 + 2*Hash01(i, 7)})
	}

	// ---------------------------------------------------------------- Cathedral body.
	b := NewGridWithSpan(g.Span())
	b.Use(Stone)

	// Central vessel walls (nave, choir, apse) and aisles with the ambulatory.
	b.Fill(band(W-wt, W, 0, wallTop, xb+0.5))
	b.Fill(band(A-wt, A, 0, aisleOut, xb+0.5))
	b.With(Slate, func() {
		b.Fill(Func{Box{V(xa-A-0.5, aisleOut-1.5, -A-0.5), V(xb+0.5, aisleIn+0.2, A+0.5)}, func(p Vec3) bool {
			a := axisA(p)
			if a < W-0.2 || a > A+0.4 {
				return false
			}
			top := aisleTop(math.Min(a, A))
			return p.Y <= top && p.Y >= top-1.25
		}})
	})
	// Cornice balustrade at the top of the clerestory and the aisles.
	b.With(Sand, func() {
		b.Fill(band(W-0.2, W+0.55, wallTop-0.9, wallTop, xb+0.5))
		b.Fill(band(A-0.2, A+0.45, aisleOut-0.8, aisleOut, xb+0.5))
	})

	// Radiating chapels around the ambulatory.
	chapelAngles := []float64{}
	for k := 0; k < 6; k++ {
		chapelAngles = append(chapelAngles, math.Pi/2+math.Pi/12+float64(k)*math.Pi/6)
	}
	for _, a := range chapelAngles {
		cc := Polar(P(xa, 0), 13.0, a)
		b.FillVolume(Cylinder{Center: cc, Radius: 2.6, Inner: 1.4, Y0: 0, Y1: 9})
		b.With(Slate, func() { b.FillVolume(Frustum{cc, 3.0, 0.4, 9, 12.5}) })
	}

	// Crossing: open the walls inside the transept, then build the transept arms.
	b.Carve(Block(V(xc-W+wt+0.05, 0.5, -zt+wt), V(xc+W-wt-0.05, ridge+2, zt-wt)))
	b.Fill(Func{Box{V(xc-W, 0, -zt), V(xc+W, wallTop, zt)}, func(p Vec3) bool {
		ax := math.Abs(p.X - xc)
		return math.Abs(p.Z) >= W-0.05 && ax >= W-wt && ax <= W
	}})
	b.Fill(Func{Box{V(xc-W, 0, -zt), V(xc+W, ridge+1.2, zt)}, func(p Vec3) bool {
		az, ax := math.Abs(p.Z), math.Abs(p.X-xc)
		return az >= zt-wt && p.Y <= wallTop+roofK*(h-ax)+1.0
	}})
	b.With(Sand, func() {
		b.Fill(Func{Box{V(xc-W-0.55, wallTop-0.9, -zt-0.55), V(xc+W+0.55, wallTop, zt+0.55)}, func(p Vec3) bool {
			ax, az := math.Abs(p.X-xc), math.Abs(p.Z)
			return (az >= W && ax >= W-0.2) || az >= zt-0.2
		}})
	})

	// Flying buttresses and piers along the nave and choir, and around the chevet.
	pierXs := []float64{}
	for k := 1; k <= 5; k++ {
		pierXs = append(pierXs, xb-float64(k)*(xb-xc-W)/6)
	}
	for k := 1; k <= 4; k++ {
		pierXs = append(pierXs, xc-W-float64(k)*(xc-W-xa)/4)
	}
	flyer := c.flyer(W-0.2, pierZ0+1, flyTop0, flyTop1)
	buttress := func(pivot Vec2, theta float64) {
		place := func(s Solid) Solid { return c.at(s, pivot, theta) }
		c.pier(b, place, pierZ0, pierZ1, pierTop)
		b.Fill(place(flyer))
	}
	for _, x := range pierXs {
		buttress(P(x, 0), math.Pi/2)
		buttress(P(x, 0), -math.Pi/2)
	}
	for k := 1; k <= 5; k++ {
		buttress(P(xa, 0), math.Pi/2+float64(k)*math.Pi/6)
	}

	// Windows: clerestory lancets with recessed glass, aisle lancets below.
	pane := func(place func(Solid) Solid, hw, sill, spring, face float64, color Color) {
		b.Carve(place(c.lancet(hw, sill, spring, face-wt-0.4, face+0.6)))
		b.With(color, func() { b.Fill(place(c.lancet(hw, sill, spring, face-wt-0.1, face-wt+0.45))) })
	}
	bays := []float64{}
	for i := 0; i+1 < len(pierXs); i++ {
		if pierXs[i] > xc && pierXs[i+1] < xc {
			continue
		}
		bays = append(bays, (pierXs[i]+pierXs[i+1])/2)
	}
	bays = append(bays, (xb+pierXs[0])/2, (pierXs[4]+xc+W)/2, (xc-W+pierXs[5])/2)
	for i, x := range bays {
		for _, s := range []float64{1, -1} {
			place := func(sol Solid) Solid { return c.at(sol, P(x, 0), s*math.Pi/2) }
			color := Navy
			if Hash01(i, int(s)) < 0.3 {
				color = Red
			}
			pane(place, 1.0, 14.2, 17.8, W, color)
			pane(place, 1.05, 2.4, 6.0, A, Navy)
		}
	}
	for k := 0; k < 6; k++ {
		theta := math.Pi/2 + math.Pi/12 + float64(k)*math.Pi/6
		place := func(sol Solid) Solid { return c.at(sol, P(xa, 0), theta) }
		pane(place, 0.95, 14.2, 17.8, W, Navy)
		cc := Polar(P(xa, 0), 13.0, theta)
		pane(func(sol Solid) Solid { return c.at(sol, cc, theta) }, 0.8, 2.4, 5.4, 2.6, Red)
	}
	// Transept: clerestory above the aisles, tall lancets at the ends, roses and portals.
	for _, s := range []float64{1, -1} {
		for _, side := range []float64{1, -1} {
			theta := 0.0
			if side < 0 {
				theta = math.Pi
			}
			for _, z := range []float64{9.2, 15.8} {
				pl := func(sol Solid) Solid { return c.at(sol, P(xc, s*z), theta) }
				if z < 10 {
					pane(pl, 1.0, 14.2, 17.8, W, Navy)
				} else {
					pane(pl, 1.0, 5.0, 15.5, W, Navy)
				}
			}
		}
		theta := s * math.Pi / 2
		end := func(sol Solid) Solid { return c.at(sol, P(xc, 0), theta) }
		endFace := func(sol Solid) Solid { return end(Translate(sol, V(0, 0, zt))) }
		c.rose(b, endFace, 14.6, 3.6, 12)
		b.Carve(endFace(c.lancet(1.7, 0.3, 4.6, -wt-0.3, 0.6)))
		b.With(Wood, func() { b.Fill(endFace(c.lancet(1.7, 0.3, 4.6, -wt-0.2, -wt+0.4))) })
		b.With(Sand, func() {
			b.Fill(endFace(Subtract(c.lancet(2.4, 0.3, 4.6, -0.2, 0.7), c.lancet(1.7, 0, 4.6, -1, 1))))
		})
		// Corner turrets with pinnacles.
		for _, side := range []float64{1, -1} {
			tc := P(xc+side*(W+0.2), s*(zt+0.2))
			b.FillVolume(Cylinder{Center: tc, Radius: 1.15, Y0: 0, Y1: wallTop + 4})
			b.FillVolume(Frustum{tc, 1.15, 0.3, wallTop + 4, wallTop + 9})
		}
		// Gable cross.
		b.With(Gold, func() {
			top := ridge + 1.0
			b.Fill(Block(V(xc-0.4, top, s*(zt-0.6)-0.4), V(xc+0.4, top+3.2, s*(zt-0.6)+0.4)))
			b.Fill(Block(V(xc-1.3, top+1.9, s*(zt-0.6)-0.4), V(xc+1.3, top+2.6, s*(zt-0.6)+0.4)))
		})
	}

	// Roofs: one shell over the nave, choir, apse and transept.
	b.With(Slate, func() {
		naveTop := func(p Vec3) (float64, bool) {
			if p.X > xf-1 || p.X < xa-h {
				return 0, false
			}
			a := axisA(p)
			if a > h {
				return 0, false
			}
			return wallTop + roofK*(h-a), true
		}
		transTop := func(p Vec3) (float64, bool) {
			ax := math.Abs(p.X - xc)
			if ax > h || math.Abs(p.Z) > zt-wt+0.3 {
				return 0, false
			}
			return wallTop + roofK*(h-ax), true
		}
		b.Fill(Func{Box{V(xa-h, wallTop, -zt-0.3), V(xf-1, ridge, zt+0.3)}, func(p Vec3) bool {
			outer, inner := false, false
			for _, f := range []func(Vec3) (float64, bool){naveTop, transTop} {
				if top, ok := f(p); ok {
					if p.Y <= top {
						outer = true
					}
					if p.Y < top-tv {
						inner = true
					}
				}
			}
			return outer && !inner
		}})
	})

	// Fleche over the crossing: an open octagonal lantern and a lead spire.
	fc := P(xc, 0)
	b.With(Slate, func() {
		b.FillVolume(Cylinder{Center: fc, Radius: 2.4, Inner: 1.2, Y0: ridge - 3, Y1: ridge + 6})
		for k := 0; k < 8; k++ {
			b.Carve(c.at(c.lancet(0.75, ridge+1.2, ridge+3.4, 0.8, 3), fc, float64(k)*math.Pi/4+math.Pi/8))
		}
		b.FillVolume(Cylinder{Center: fc, Radius: 2.9, Y0: ridge + 6, Y1: ridge + 7})
		b.FillVolume(Frustum{fc, 2.2, 0.5, ridge + 7, ridge + 30})
	})
	b.With(Gold, func() {
		b.FillVolume(Cylinder{Center: fc, Radius: 2.0, Y0: ridge + 13, Y1: ridge + 13.7})
		b.FillVolume(Cylinder{Center: fc, Radius: 1.5, Y0: ridge + 20, Y1: ridge + 20.7})
		top := ridge + 30
		b.Fill(Cylinder{Center: fc, Radius: 0.5, Y0: top - 0.5, Y1: top + 4})
		b.Fill(Block(V(xc-1.4, top+2.2, -0.4), V(xc+1.4, top+2.9, 0.4)))
	})
	for k := 0; k < 4; k++ {
		pc := Polar(fc, 3.4, float64(k)*math.Pi/2+math.Pi/4)
		b.FillVolume(Frustum{pc, 0.8, 0.3, ridge - 2, ridge + 5})
	}

	// ---------------------------------------------------------------- West front.
	tw := []float64{5.5, 14.5} // Tower span in |z|.
	b.Use(Stone)
	// Lower block up to the grand gallery, hollow, with a floor slab at its top.
	b.Fill(Subtract(Block(V(xb, 0, -tw[1]), V(xf, 27, tw[1])), Block(V(xb+1.3, 0.3, -tw[1]+1.3), V(xf-1.3, 26, tw[1]-1.3))))
	// Towers.
	for _, s := range []float64{1, -1} {
		z0, z1 := s*tw[0], s*tw[1]
		b.Fill(Subtract(Block(V(xb, 27, z0), V(xf, 47, z1)), Block(V(xb+1.3, 27, math.Min(z0, z1)+1.3), V(xf-1.3, 46, math.Max(z0, z1)-1.3))))
	}
	// Grand gallery screen between the towers and balustrade above.
	b.Fill(Block(V(xf-1.3, 27, -tw[0]), V(xf, 31.5, tw[0])))
	b.With(Sand, func() { b.Fill(Block(V(xf-1.3, 31.5, -tw[0]), V(xf+0.3, 32.3, tw[0]))) })
	// Nave gable behind the gallery.
	b.Fill(Func{Box{V(xb-0.3, wallTop, -h), V(xb+1.0, ridge+1, h)}, func(p Vec3) bool {
		return p.Y <= wallTop+roofK*(h-math.Abs(p.Z))+0.8
	}})
	// Porch zone and gallery of kings.
	b.Fill(Block(V(xf, 0, -tw[1]), V(xf+1.3, 12, tw[1])))
	b.With(Sand, func() {
		b.Fill(Block(V(xf, 12, -tw[1]), V(xf+1.0, 14.6, tw[1])))
		b.Fill(Block(V(xf, 14.6, -tw[1]), V(xf+1.5, 15.3, tw[1])))
	})
	for z := -13.6; z <= 13.7; z += 1.6 {
		b.Carve(Block(V(xf+0.3, 12.5, z-0.4), V(xf+1.2, 14.2, z+0.4)))
		b.Paint(DarkStone, Block(V(xf-0.4, 12.5, z-0.4), V(xf+0.3, 14.2, z+0.4)))
	}
	// Facade buttresses.
	for _, z := range []float64{-tw[1], -tw[0], tw[0], tw[1]} {
		b.Fill(BlockAt(P(xf+0.8, z), 1.6, 1.7, 0, 27))
		b.Fill(BlockAt(P(xf+0.4, z), 0.8, 1.5, 27, 46))
	}
	// Tower flanks: corner buttresses, string courses, blind lancets and the gallery.
	for _, s := range []float64{1, -1} {
		zo := s * tw[1]
		for _, x := range []float64{xb + 0.85, xf - 0.85} {
			b.Fill(BlockAt(P(x, zo+s*0.4), 1.7, 0.8, 0, 46))
		}
		b.With(Sand, func() {
			for _, y := range []float64{14.6, 31.5} {
				b.Fill(Block(V(xb-0.2, y, zo), V(xf, y+0.8, zo+s*1.0)))
			}
		})
		side := func(x float64) func(Solid) Solid {
			return func(sol Solid) Solid { return c.at(sol, P(x, zo), s*math.Pi/2) }
		}
		for _, x := range []float64{35.1, 38.9} {
			b.Carve(side(x)(c.lancet(1.05, 16.4, 22.0, -0.5, 1)))
			b.With(Navy, func() { b.Fill(side(x)(c.lancet(1.05, 16.4, 22.0, -1.1, -0.4))) })
		}
		b.With(Sand, func() {
			b.Fill(side(37)(Subtract(c.lancet(3.0, 16.0, 22.0, -0.2, 0.6), c.lancet(2.6, 15, 22, -1, 1))))
		})
		b.Carve(side(37)(c.lancet(1.1, 3.5, 8.0, -0.5, 1)))
		b.With(Navy, func() { b.Fill(side(37)(c.lancet(1.1, 3.5, 8.0, -1.1, -0.4))) })
		for _, x := range []float64{35, 37, 39} {
			b.Carve(side(x)(c.lancet(0.6, 27.4, 29.9, -1.6, 1.0)))
		}
	}
	front := func(z float64) func(Solid) Solid {
		return func(s Solid) Solid { return c.at(s, P(xf, z), 0) }
	}
	// Portals: stepped archivolts, wooden doors and sand tympana.
	for _, p := range []struct{ z, hw, spring float64 }{{0, 3.0, 5.6}, {-10, 2.5, 5.0}, {10, 2.5, 5.0}} {
		pl := front(p.z)
		b.Carve(pl(c.lancet(p.hw+0.5, 0.3, p.spring, 0.55, 2)))
		b.Paint(Sand, pl(c.lancet(p.hw+0.6, 0.3, p.spring, -1.0, 2)))
		for i := 0; i < 3; i++ {
			b.Carve(pl(c.lancet(p.hw-0.6*float64(i), 0.3, p.spring, -0.2-0.7*float64(i), 2)))
		}
		inner := p.hw - 1.2
		b.With(Wood, func() { b.Fill(pl(Block(V(-inner, 0.3, -1.9), V(inner, p.spring-0.6, -1.3)))) })
		b.With(Sand, func() {
			b.Fill(pl(Subtract(c.lancet(inner, p.spring-0.6, p.spring, -1.9, -1.3), Block(V(-inner, 0, -2), V(inner, p.spring-0.6, 0)))))
		})
		b.With(DarkStone, func() { b.Fill(pl(Block(V(-0.3, 0.3, -1.9), V(0.3, p.spring-0.6, -1.2)))) })
	}
	// Great rose.
	c.rose(b, front(0), 21, 4.4, 16)
	// Twin blind lancets on the towers at rose level.
	for _, z := range []float64{-11.9, -8.1, 8.1, 11.9} {
		b.Carve(front(z)(c.lancet(1.05, 16.4, 22.0, -0.5, 1)))
		b.With(Navy, func() { b.Fill(front(z)(c.lancet(1.05, 16.4, 22.0, -1.1, -0.4))) })
	}
	for _, s := range []float64{1, -1} {
		b.With(Sand, func() {
			b.Fill(front(s * 10)(Subtract(c.lancet(3.0, 16.0, 22.0, -0.2, 0.6), c.lancet(2.6, 15, 22, -1, 1))))
		})
	}
	// Grand gallery: open lancet arcade across the whole front.
	for z := -13.4; z <= 13.5; z += 2.0 {
		if math.Abs(math.Abs(z)-tw[0]) < 1.3 || math.Abs(math.Abs(z)-tw[1]) < 1.3 {
			continue
		}
		b.Carve(front(z)(c.lancet(0.6, 27.4, 29.9, -1.6, 1.0)))
	}
	b.With(DarkStone, func() { b.Fill(Block(V(xb+1.3, 26, -tw[1]+1.3), V(xf-1.3, 27, tw[1]-1.3))) })
	// Belfries: two tall open lancets per face, through both walls.
	for _, s := range []float64{1, -1} {
		zc := s * (tw[0] + tw[1]) / 2
		for _, d := range []float64{-2.0, 2.0} {
			b.Carve(c.at(c.lancet(1.15, 32.6, 41.6, -6, 6), P((xb+xf)/2, zc+d), 0))
			b.Carve(c.at(c.lancet(1.15, 32.6, 41.6, -6, 6), P((xb+xf)/2+d, zc), math.Pi/2))
		}
		// Platform, open parapet and corner pinnacles.
		zmin, zmax := math.Min(s*tw[0], s*tw[1]), math.Max(s*tw[0], s*tw[1])
		b.Fill(Block(V(xb, 46, zmin), V(xf, 47, zmax)))
		b.With(Sand, func() {
			b.Fill(Subtract(Block(V(xb-0.3, 47, zmin-0.3), V(xf+0.3, 48.4, zmax+0.3)), Block(V(xb+0.6, 47, zmin+0.6), V(xf-0.6, 49, zmax-0.6))))
		})
		b.Carve(Func{Box{V(xb-1, 47.6, zmin-1), V(xf+1, 48.5, zmax+1)}, func(p Vec3) bool {
			return math.Mod(p.X+p.Z+200, 2.0) < 0.9
		}})
		for _, x := range []float64{xb + 0.4, xf - 0.4} {
			for _, z := range []float64{zmin + 0.4, zmax - 0.4} {
				b.FillVolume(Frustum{P(x, z), 0.8, 0.3, 47, 50.6})
			}
		}
	}

	// Plinth course.
	b.Paint(DarkStone, Block(V(-60, -1, -40), V(50, 0.9, 40)))
	g.Merge(b)

	// Historically the choir rose first and the west front last.
	return Along(P(1, 0))
}
