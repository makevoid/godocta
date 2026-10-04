package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// Caravanserai is a Safavid Silk Road inn in the desert, after the Abbasi and Tash Rabat
// caravanserais: a square fortified courtyard of sand and fired-brick walls with round
// corner towers and half-round buttress turrets. A monumental tiled pishtaq with a deep
// pointed iwan opens to +Z; inside, two storeys of arched alcoves with dark rooms ring
// the courtyard, four iwans face a central pool with a domed pavilion, and rows of small
// domes and walkways cover the roof. Outside, a camel caravan arrives along the track,
// camels rest at a hitching rail, merchants' carpets and bales are piled by the gate, and
// tents, date palms, a domed ab anbar cistern with wind catchers and a line of qanat
// mounds sit on rolling dunes and rocky desert sloping to a low rim.
type Caravanserai struct{}

func init() { Register(Caravanserai{}) }

func (Caravanserai) Name() string  { return "caravanserai" }
func (Caravanserai) Title() string { return "Caravanserai" }

const (
	caravanseraiFloor  = 1.5  // Ground level at the building.
	caravanseraiHalf   = 28.0 // Outer wall half-size.
	caravanseraiWall   = 1.4  // Outer wall thickness.
	caravanseraiCourt  = 19.5 // Courtyard half-size.
	caravanseraiFacade = 1.2  // Courtyard facade thickness.
	caravanseraiRoof   = caravanseraiFloor + 11.2
	caravanseraiBack   = 23.3 // Courtyard iwans' back wall (inner face).
	caravanseraiSiteX  = 46.0
	caravanseraiSiteZ  = 46.0
	caravanseraiSiteZ0 = 1.5
)

// caravanseraiDunes are smooth sand bumps: center x, z, radii x, z and height.
var caravanseraiDunes = [][5]float64{
	{-24, -37.5, 14, 6.5, 4.5},
	{2, -39, 12, 5.5, 3.4},
	{-40, -14, 5.5, 11, 2.6},
	{22, -40, 9, 5, 2.6},
	{33, -35, 10, 6.5, 3.6},
	{41, -20, 4.5, 9, 2.4},
	{-37, 6, 5, 9, 1.6},
}

// ground is the desert surface height, or -1 outside the site: a flat plateau at the
// building that drops smoothly to a low rim, with dunes rising away from the walls.
func (Caravanserai) ground(x, z float64) float64 {
	nx, nz := x/caravanseraiSiteX, (z-caravanseraiSiteZ0)/caravanseraiSiteZ
	n := math.Pow(nx*nx*nx*nx+nz*nz*nz*nz, 0.25)
	if n > 1 {
		return -1
	}
	h := caravanseraiFloor
	if n > 0.7 {
		t := (n - 0.7) / 0.3
		h -= 1.25 * t * t * (3 - 2*t)
	}
	out := math.Max(math.Abs(x), math.Abs(z)) - 33
	mask := math.Max(0, math.Min(1, out/7)) * math.Max(0, math.Min(1, (1-n)/0.16))
	if mask > 0 {
		for _, d := range caravanseraiDunes {
			dx, dz := (x-d[0])/d[2], (z-d[1])/d[3]
			if r := dx*dx + dz*dz; r < 1 {
				h += mask * d[4] * 0.5 * (1 + math.Cos(math.Pi*math.Sqrt(r)))
			}
		}
		h += mask * 0.35 * math.Sin(0.45*x+0.25*z) * math.Sin(0.2*z-0.1*x)
	}
	return h
}

// noise is smooth value noise in [0, 1) on a grid of the given cell size.
func (Caravanserai) noise(x, z, cell float64, seed int) float64 {
	fx, fz := x/cell, z/cell
	ix, iz := math.Floor(fx), math.Floor(fz)
	tx, tz := fx-ix, fz-iz
	tx, tz = tx*tx*(3-2*tx), tz*tz*(3-2*tz)
	h := func(a, b int) float64 { return Hash01(int(ix)+a, int(iz)+b, seed) }
	lo := h(0, 0) + (h(1, 0)-h(0, 0))*tx
	hi := h(0, 1) + (h(1, 1)-h(0, 1))*tx
	return lo + (hi-lo)*tz
}

// cellHash is a per-cell deterministic value for color variation.
func (Caravanserai) cellHash(p Vec3, seed int) float64 {
	return Hash01(int(math.Round(p.X*4)), int(math.Round(p.Y*4)), int(math.Round(p.Z*4)), seed)
}

// pointed is a two-centred Persian arch opening along Z, centred at x = cx, from z0
// to z1: jambs up to spring, then two arcs meeting at a pointed crown.
func (Caravanserai) pointed(cx, sill, spring, hw, z0, z1 float64) Func {
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
func (Caravanserai) apex(spring, hw float64) float64 {
	r := 1.6 * hw
	return spring + math.Sqrt(r*r-(r-hw)*(r-hw))
}

// side turns a solid authored on the +Z side to side k (quarter turns).
func (Caravanserai) side(s Solid, k int) Solid {
	if k == 0 {
		return s
	}
	return RotateY(s, P(0, 0), float64(k)*math.Pi/2)
}

// square is a centred square block.
func (Caravanserai) square(h, y0, y1 float64) Cuboid {
	return Block(V(-h, y0, -h), V(h, y1, h))
}

// place turns a local solid (authored around the origin) to a heading and moves it to pos.
func (Caravanserai) place(s Solid, pos Vec2, heading float64) Solid {
	return RotateY(Translate(s, V(pos.X, 0, pos.Y)), pos, heading)
}

func (c Caravanserai) Build(g *Grid) Order {
	c.terrain(g)
	c.shell(g)
	c.courtyardFacade(g)
	c.iwans(g)
	c.entrance(g)
	c.roof(g)
	c.courtyard(g)
	c.cistern(g, P(-34, 35))
	c.outside(g)
	return Sweep(P(0, 0), math.Pi/2)
}

// terrain lays a hollow desert surface, colors sand drifts and rocky patches, and adds
// eroded outcrops toward the back.
func (c Caravanserai) terrain(g *Grid) {
	box := Box{V(-caravanseraiSiteX, 0, caravanseraiSiteZ0-caravanseraiSiteZ), V(caravanseraiSiteX, 9, caravanseraiSiteZ0+caravanseraiSiteZ)}
	g.Use(Sand).Fill(Func{box, func(p Vec3) bool {
		h := c.ground(p.X, p.Z)
		if h < 0 || p.Y > h {
			return false
		}
		const e = 0.5
		sx := (c.ground(p.X+e, p.Z) - c.ground(p.X-e, p.Z)) / (2 * e)
		sz := (c.ground(p.X, p.Z+e) - c.ground(p.X, p.Z-e)) / (2 * e)
		t := 0.8 + 0.8*math.Hypot(sx, sz)
		return p.Y >= h-t || p.Y <= 0.05
	}})
	// Rocky desert patches and a gravel rim.
	g.Paint(Terracotta, Func{box, func(p Vec3) bool {
		if math.Max(math.Abs(p.X), math.Abs(p.Z)) < 31 {
			return false
		}
		n := c.noise(p.X, p.Z, 7, 1)
		nx, nz := p.X/caravanseraiSiteX, (p.Z-caravanseraiSiteZ0)/caravanseraiSiteZ
		rim := math.Pow(nx*nx*nx*nx+nz*nz*nz*nz, 0.25)
		return n > 0.72 || rim > 0.97 || (n > 0.62 && c.cellHash(p, 2) < 0.3)
	}})
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		return math.Max(math.Abs(p.X), math.Abs(p.Z)) > 30 && c.cellHash(p, 3) < 0.02
	}})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		return math.Max(math.Abs(p.X), math.Abs(p.Z)) > 30 && c.cellHash(p, 4) < 0.012
	}})

	// Eroded outcrops of layered rock.
	for i, o := range [][4]float64{{37, -27, 3.2, 4.0}, {41, -12, 2.4, 3.0}, {29, -38, 2.6, 2.8}, {-40, -31, 2.3, 2.5}, {-6, -40, 2.0, 2.0}} {
		at := P(o[0], o[1])
		gy := c.ground(at.X, at.Y)
		var parts []Solid
		for k := 0; k < 4; k++ {
			off := Polar(at, o[2]*0.7*Hash01(i, k, 1), Hash01(i, k, 2)*2*math.Pi)
			r := o[2] * (0.55 + 0.4*Hash01(i, k, 3))
			parts = append(parts, Ellipsoid{off.At(gy - 0.5), V(r, o[3]*(0.6+0.4*Hash01(i, k, 4)), r*0.9)})
		}
		g.Use(Terracotta).FillVolume(Union(parts...))
		top := gy + o[3]
		g.Paint(Brick, Func{Box{V(at.X-6, gy, at.Y-6), V(at.X+6, top+1, at.Y+6)}, func(p Vec3) bool {
			return math.Mod(p.Y-gy+10, 1.6) < 0.45
		}})
		g.Paint(DarkStone, Func{Box{V(at.X-6, gy, at.Y-6), V(at.X+6, top+1, at.Y+6)}, func(p Vec3) bool {
			return c.cellHash(p, 9) < 0.08
		}})
	}
}

// shell builds the outer walls, corner towers, buttress turrets, the two storeys of
// rooms around the courtyard and their dark interiors.
func (c Caravanserai) shell(g *Grid) {
	const (
		F, H, W, C = caravanseraiFloor, caravanseraiHalf, caravanseraiWall, caravanseraiCourt
		roof       = caravanseraiRoof
		towerTop   = roof + 3.6
	)
	rin, rout := C+caravanseraiFacade, H-W
	g.Use(Sand)
	g.Fill(Subtract(c.square(H, F, roof), c.square(C, F-1, roof+1)))
	// Outer parapet with stepped merlons.
	g.Fill(Subtract(c.square(H, roof, roof+0.9), c.square(H-0.8, roof-1, roof+2)))
	g.Fill(Func{Box{V(-H, roof+0.9, -H), V(H, roof+1.8, H)}, func(p Vec3) bool {
		ax, az := math.Abs(p.X), math.Abs(p.Z)
		if ax < H-0.8 && az < H-0.8 {
			return false
		}
		along := p.X
		if ax >= H-0.8 {
			along = p.Z
		}
		return math.Mod(along+100, 2.0) < 1.1
	}})

	// Round corner towers with a slight batter and a crenellated top.
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			ct := P(sx*H, sz*H)
			g.FillVolume(Subtract(Frustum{ct, 4.7, 4.1, F, towerTop}, Frustum{ct, 3.4, 2.8, F + 0.5, towerTop - 1}))
			g.FillVolume(Cylinder{Center: ct, Radius: 4.4, Inner: 3.5, Y0: towerTop, Y1: towerTop + 0.9})
			g.Fill(Func{Box{V(ct.X-4.4, towerTop+0.9, ct.Y-4.4), V(ct.X+4.4, towerTop+1.8, ct.Y+4.4)}, func(p Vec3) bool {
				d := p.Plan().Sub(ct)
				l := d.Length()
				return l <= 4.4 && l > 3.5 && math.Mod(d.Angle()/(2*math.Pi)*28+28, 2) < 1.1
			}})
		}
	}
	// Half-round buttress turrets along the walls.
	for k := 0; k < 4; k++ {
		ts := []float64{-19, -9.5, 0, 9.5, 19}
		if k == 0 {
			ts = []float64{-19, 19}
		}
		for _, t := range ts {
			ct := P(t, H).Rotate(float64(k) * math.Pi / 2)
			g.FillVolume(Subtract(Frustum{ct, 2.1, 1.75, F, roof + 1.4}, Cylinder{Center: ct, Radius: 0.8, Y0: F + 0.5, Y1: roof}))
			g.FillVolume(Cylinder{Center: ct, Radius: 2.0, Y0: roof + 1.4, Y1: roof + 2.0})
			g.FillVolume(Ellipsoid{ct.At(roof + 2.0), V(1.3, 1.0, 1.3)})
		}
	}

	// Rooms: two storeys around the courtyard, divided by a floor slab.
	rooms := func(y0, y1 float64) Solid { return Subtract(c.square(rout, y0, y1), c.square(rin, y0-1, y1+1)) }
	g.Carve(rooms(F+0.3, F+5), rooms(F+5.6, roof-1.0))

	// Front arcade of shops (two storeys) either side of the portal.
	for _, x := range []float64{-14.6, -10.9, 10.9, 14.6} {
		for _, sill := range []float64{F + 0.4, F + 6} {
			g.Carve(Block(V(x-1.75, sill-0.2, H-0.8), V(x+1.75, sill+4.6, H+0.5)))
			g.Carve(c.pointed(x, sill, sill+2.2, 1.25, H-2, H+1))
		}
	}
	// Blind pointed niches with small upper windows on the other three sides.
	for k := 1; k < 4; k++ {
		for _, t := range []float64{-14.25, -4.75, 4.75, 14.25} {
			g.Carve(c.side(c.pointed(t, F+1.2, F+6.4, 2.1, H-0.8, H+1), k))
			g.Carve(c.side(c.pointed(t, F+7.0, F+8.4, 0.8, H-2, H+1), k))
		}
	}

	// Fired-brick bands and mottled mud brick on every wall surface.
	box := Box{V(-H-5, F+0.1, -H-5), V(H+5, towerTop+2, H+5)}
	band := func(p Vec3) bool { return math.Mod(p.Y-F+0.2, 2.4) < 0.75 }
	g.Paint(Brick, Func{box, func(p Vec3) bool { return band(p) && c.cellHash(p, 11) > 0.12 }})
	g.Paint(Terracotta, Func{box, func(p Vec3) bool {
		h := c.cellHash(p, 11)
		return (band(p) && h <= 0.12) || (!band(p) && c.cellHash(p, 12) < 0.035)
	}})
	g.Paint(Brick, Cylinder{Center: P(0, 0), Radius: 100, Y0: towerTop + 0.85, Y1: towerTop + 2})
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			g.Paint(Sand, Cylinder{Center: P(sx*H, sz*H), Radius: 3.5, Y0: towerTop - 0.5, Y1: towerTop + 0.2})
		}
	}

	// Dark interiors: the inner skin of every room, its floor and ceiling.
	g.Paint(Charcoal, Subtract(c.square(rout+0.55, F-0.7, roof-0.55), c.square(rin-0.55, F-2, roof)))
	// Tiled spandrels in the front arcade panels.
	for _, x := range []float64{-14.6, -10.9, 10.9, 14.6} {
		for _, sill := range []float64{F + 0.4, F + 6} {
			panel := Block(V(x-1.75, sill-0.2, H-1.25), V(x+1.75, sill+4.6, H-0.6))
			g.Paint(Water, panel)
			g.Paint(Navy, Subtract(panel, Block(V(x-1.35, sill-1, H-2), V(x+1.35, sill+4.2, H))))
			g.Paint(White, Intersect(panel, Subtract(c.pointed(x, sill, sill+2.2, 1.75, H-2, H), c.pointed(x, sill-1, sill+2.2, 1.25, H-2, H))))
		}
	}
	// Blind niche backs in plain mud plaster.
	for k := 1; k < 4; k++ {
		for _, t := range []float64{-14.25, -4.75, 4.75, 14.25} {
			g.Paint(Terracotta, c.side(Subtract(c.pointed(t, F+1.2, F+6.4, 2.1, H-1.3, H-0.6), c.pointed(t, F+7.0, F+8.4, 0.8, H-2, H+1)), k))
		}
	}
}

// courtyardFacade carves the two storeys of framed alcoves on all four courtyard sides.
func (c Caravanserai) courtyardFacade(g *Grid) {
	const F, C = caravanseraiFloor, caravanseraiCourt
	for k := 0; k < 4; k++ {
		for _, t := range []float64{-16.75, -12.25, -7.75, 7.75, 12.25, 16.75} {
			for _, st := range [][2]float64{{F + 0.3, F + 0.9}, {F + 5.5, F + 5.7}} {
				panel, sill := st[0], st[1]
				g.Carve(c.side(Block(V(t-1.95, panel, C-0.6), V(t+1.95, panel+4.6, C+0.4)), k))
				g.Carve(c.side(c.pointed(t, sill, sill+1.8, 1.4, C-1, C+2), k))
				// Brick frame around each panel.
				g.Paint(Brick, c.side(Subtract(Block(V(t-2.35, panel-0.3, C-0.4), V(t+2.35, panel+5.0, C+0.45)),
					Block(V(t-1.95, panel, C-1), V(t+1.95, panel+4.6, C+1))), k))
				g.Paint(Terracotta, c.side(Block(V(t-1.95, panel, C+0.35), V(t+1.95, panel+4.6, C+0.45)), k))
			}
		}
		// String course at the floor slab and a brick cornice at the roof.
		g.Paint(Brick, c.side(Block(V(-C, F+4.9, C-0.4), V(C, F+5.5, C+0.3)), k))
	}
}

// iwans builds the four courtyard iwans: tall tiled pishtaq frames before deep vaults
// that run into the ring of rooms.
func (c Caravanserai) iwans(g *Grid) {
	const (
		F, C   = caravanseraiFloor, caravanseraiCourt
		back   = caravanseraiBack
		hw, sp = 3.4, F + 7.0
		fw     = 5.6
		ftop   = F + 15.0
	)
	for k := 0; k < 4; k++ {
		sd := func(s Solid) Solid { return c.side(s, k) }
		g.Use(Sand)
		g.Fill(sd(Subtract(c.pointed(0, F, sp, hw+1.2, C, back+1.2), c.pointed(0, F-1, sp, hw, C-3, back))))
		g.Fill(sd(Block(V(-fw, F, C-1.2), V(fw, ftop, C+1.2))))
		g.Carve(sd(c.pointed(0, F+0.3, sp, hw, C-2, back)))
		// Tiles on the frame face.
		face := func(x0, y0, x1, y1 float64) Cuboid { return Block(V(x0, y0, C-1.45), V(x1, y1, C-0.85)) }
		g.Paint(Navy, sd(face(-fw, F, fw, ftop)))
		g.Paint(Water, sd(face(-fw+1.0, F+1.6, fw-1.0, ftop-1.2)))
		g.Paint(Glass, sd(Func{face(-fw+1.0, F+1.6, fw-1.0, ftop-1.2).Bounds(), func(p Vec3) bool {
			q := p.Plan().Rotate(-float64(k) * math.Pi / 2)
			return (int(math.Floor((q.X+p.Y)/1.3+50))+int(math.Floor((q.X-p.Y)/1.3+50)))%2 == 0
		}}))
		g.Paint(White, sd(Subtract(face(-fw+0.7, F+1.3, fw-0.7, ftop-0.9), face(-fw+1.1, F+1.7, fw-1.1, ftop-1.3))))
		g.Paint(White, sd(Intersect(face(-fw, F, fw, ftop), Subtract(c.pointed(0, F, sp, hw+0.9, C-2, C), c.pointed(0, F-1, sp, hw, C-3, C+1)))))
		g.Paint(Stone, sd(face(-fw, F, fw, F+1.3)))
		g.Paint(Gold, sd(Block(V(-fw, ftop-0.45, C-1.45), V(fw, ftop, C+1.2))))
		// Dark vault with a tiled hood near the opening.
		g.Paint(Charcoal, sd(c.pointed(0, F-0.8, sp, hw+0.6, C-0.6, back+0.6)))
		g.Paint(Navy, sd(Intersect(Block(V(-hw-1, sp-0.2, C-0.7), V(hw+1, ftop, C+1.4)), c.pointed(0, F, sp, hw+0.6, C-0.7, C+1.4))))
		g.Paint(White, sd(Func{Intersect(Block(V(-hw-1, sp-0.2, C-0.7), V(hw+1, ftop, C+1.4)), c.pointed(0, F, sp, hw+0.6, C-0.7, C+1.4)).Bounds(), func(p Vec3) bool {
			return c.pointed(0, F, sp, hw+0.6, -1000, 1000).Contains(p.Plan().Rotate(-float64(k)*math.Pi/2).At(p.Y)) &&
				p.Y > sp && Hash01(int(math.Floor(p.X)), int(math.Floor(p.Y)), int(math.Floor(p.Z))) < 0.35
		}}))
		// Gold finials on the frame corners.
		for _, x := range []float64{-fw + 0.5, fw - 0.5} {
			ct := P(x, C-0.7).Rotate(float64(k) * math.Pi / 2)
			g.With(Gold, func() { g.Fill(Cylinder{Center: ct, Radius: 0.5, Y0: ftop, Y1: ftop + 1.2}) })
		}
	}
	// The +Z iwan's back wall opens into the entrance passage.
	g.Carve(c.pointed(0, F+0.3, F+4.2, 2.0, back-1, back+2.4))
}

// entrance builds the monumental portal on the outer +Z face: a raised pishtaq framed
// in tile bands, a deep pointed vault with a tiled hood and two tiled corner shafts.
func (c Caravanserai) entrance(g *Grid) {
	const (
		F, H   = caravanseraiFloor, caravanseraiHalf
		back   = caravanseraiBack
		hw, sp = 4.2, F + 8.5
		fw     = 8.5
		ftop   = F + 19.0
		front  = H + 2.5
	)
	g.Use(Sand)
	g.Fill(Block(V(-fw, F, H-1.5), V(fw, ftop, front)))
	g.Carve(Block(V(-fw+1.2, F+0.5, H-0.3), V(fw-1.2, ftop-1.2, front-1.2)))
	g.Fill(Subtract(c.pointed(0, F, sp, hw+1.2, back, front), c.pointed(0, F-1, sp, hw, back+1.2, front+2)))
	g.Carve(c.pointed(0, F+0.3, sp, hw, back+1.2, front+2))
	g.Carve(c.pointed(0, F+0.3, F+4.2, 2.0, back-1, back+2.4))
	// Upper window in the vault's back wall.
	g.Carve(c.pointed(0, F+7.6, F+9.2, 1.1, back, back+2.4))

	face := func(x0, y0, x1, y1 float64) Cuboid { return Block(V(x0, y0, front-0.35), V(x1, y1, front+0.3)) }
	g.Paint(Navy, face(-fw, F, fw, ftop))
	g.Paint(Water, face(-fw+1.2, F+2, fw-1.2, ftop-3))
	g.Paint(Glass, Func{face(-fw+1.2, F+2, fw-1.2, ftop-3).Bounds(), func(p Vec3) bool {
		u, v := (p.X+p.Y)/1.4, (p.X-p.Y)/1.4
		fu, fv := u-math.Floor(u), v-math.Floor(v)
		return math.Abs(fu-0.5) < 0.22 || math.Abs(fv-0.5) < 0.22
	}})
	g.Paint(Navy, Func{face(-fw+1.2, F+2, fw-1.2, ftop-3).Bounds(), func(p Vec3) bool {
		u, v := (p.X+p.Y)/1.4, (p.X-p.Y)/1.4
		fu, fv := u-math.Floor(u), v-math.Floor(v)
		return math.Abs(fu-0.5) < 0.22 && math.Abs(fv-0.5) < 0.22
	}})
	g.Paint(White, Subtract(face(-fw+0.8, F+1.6, fw-0.8, ftop-2.6), face(-fw+1.2, F+2, fw-1.2, ftop-3)))
	// Inscription band.
	g.Paint(White, face(-fw+1.2, ftop-2.4, fw-1.2, ftop-0.8))
	g.Paint(Navy, Func{face(-fw+1.6, ftop-2.1, fw-1.6, ftop-1.1).Bounds(), func(p Vec3) bool {
		return Hash01(int(math.Floor(p.X*1.3)), int(math.Floor(p.Y*1.6)), 5) < 0.5
	}})
	g.Paint(Gold, face(-fw, ftop-0.5, fw, ftop))
	// Arch outline bands.
	ring := func(a, b float64) Solid {
		return Intersect(face(-fw, F, fw, ftop), Subtract(c.pointed(0, F, sp, a, front-2, front+1), c.pointed(0, F-1, sp, b, front-3, front+2)))
	}
	g.Paint(White, ring(hw+1.0, hw))
	g.Paint(Gold, ring(hw+1.4, hw+1.0))
	g.Paint(Stone, face(-fw, F, fw, F+1.6))
	// Arch reveal tiles, dark vault and tiled hood.
	g.Paint(Charcoal, c.pointed(0, F-0.8, sp, hw+0.6, back+0.6, front-1.6))
	g.Paint(Water, Subtract(c.pointed(0, F+1.6, sp, hw+0.6, front-1.6, front-0.3), c.pointed(0, F, sp, hw-0.3, front-3, front+2)))
	hood := Intersect(Block(V(-hw-1, sp-0.6, front-3.4), V(hw+1, ftop, front-0.3)), c.pointed(0, F, sp, hw+0.6, front-3.4, front-0.3))
	g.Paint(Navy, hood)
	g.Paint(White, Func{hood.Bounds(), func(p Vec3) bool {
		return hood.Contains(p) && Hash01(int(math.Floor(p.X*1.2)), int(math.Floor(p.Y*1.2)), int(math.Floor(p.Z*1.2)), 7) < 0.3
	}})
	g.Paint(Gold, Func{hood.Bounds(), func(p Vec3) bool {
		return hood.Contains(p) && Hash01(int(math.Floor(p.X*1.2)), int(math.Floor(p.Y*1.2)), int(math.Floor(p.Z*1.2)), 8) < 0.12
	}})

	// Tiled shafts at the frame corners.
	for _, x := range []float64{-fw, fw} {
		ct := P(x, front-0.6)
		g.With(Sand, func() { g.FillVolume(Cylinder{Center: ct, Radius: 1.2, Y0: F, Y1: ftop + 2.4}) })
		g.Paint(Navy, Cylinder{Center: ct, Radius: 1.3, Y0: F + 11, Y1: ftop + 2.4})
		g.Paint(Water, Func{Box{V(ct.X-1.3, F+11, ct.Y-1.3), V(ct.X+1.3, ftop+2.4, ct.Y+1.3)}, func(p Vec3) bool {
			a := p.Plan().Sub(ct).Angle()
			return math.Mod(a/(2*math.Pi)*8+p.Y/1.2+100, 2) < 1
		}})
		g.Paint(Gold, Cylinder{Center: ct, Radius: 1.3, Y0: F + 10.6, Y1: F + 11})
		g.With(Glass, func() { g.FillVolume(Ellipsoid{ct.At(ftop + 2.4), V(1.3, 1.5, 1.3)}) })
		g.With(Gold, func() { g.Fill(Cylinder{Center: ct, Radius: 0.45, Y0: ftop + 3.6, Y1: ftop + 4.8}) })
	}
	// Doors folded open at the passage.
	g.With(Wood, func() {
		g.Fill(Block(V(-2.6, F+0.3, back+1.2), V(-1.9, F+4.4, back+2.6)))
		g.Fill(Block(V(1.9, F+0.3, back+1.2), V(2.6, F+4.4, back+2.6)))
	})
}

// roof adds the rows of small domes over the rooms, the inner parapet and walkways.
func (c Caravanserai) roof(g *Grid) {
	const (
		C, H = caravanseraiCourt, caravanseraiHalf
		roof = caravanseraiRoof
	)
	rin, rout := C+caravanseraiFacade, H-caravanseraiWall
	m := (rin + rout) / 2
	// Walkway paving along the outer parapet and a low inner parapet.
	g.Paint(Stone, Subtract(c.square(rout+0.1, roof-0.4, roof+0.1), c.square(rout-1.3, roof-1, roof+1)))
	g.Use(Sand).Fill(Subtract(c.square(rin, roof, roof+0.7), c.square(C, roof-1, roof+1)))
	g.Paint(Brick, Subtract(c.square(rin, roof+0.4, roof+0.7), c.square(C, roof-1, roof+1)))
	for k := 0; k < 4; k++ {
		for i, t := range []float64{-16.75, -12.25, -7.75, 7.75, 12.25, 16.75} {
			ct := P(t, m).Rotate(float64(k) * math.Pi / 2)
			tiled := math.Abs(t) < 8
			col := Sand
			if tiled {
				col = Navy
			}
			g.With(col, func() {
				g.FillVolume(Cylinder{Center: ct, Radius: 2.0, Y0: roof - 0.2, Y1: roof + 0.9})
				g.FillVolume(Ellipsoid{ct.At(roof + 0.9), V(1.95, 1.9, 1.95)})
			})
			if tiled {
				g.Paint(Water, Func{Box{V(ct.X-2, roof+0.9, ct.Y-2), V(ct.X+2, roof+2.9, ct.Y+2)}, func(p Vec3) bool {
					a := p.Plan().Sub(ct).Angle()
					return math.Mod(a/(2*math.Pi)*8+16, 1) < 0.35
				}})
				g.With(Gold, func() { g.Fill(Cylinder{Center: ct, Radius: 0.45, Y0: roof + 2.6, Y1: roof + 3.6}) })
			} else {
				g.Paint(Brick, Cylinder{Center: ct, Radius: 2.1, Y0: roof + 0.4, Y1: roof + 0.9})
				g.Paint(Terracotta, Func{Box{V(ct.X-2, roof+0.9, ct.Y-2), V(ct.X+2, roof+2.9, ct.Y+2)}, func(p Vec3) bool {
					return c.cellHash(p, 20+i) < 0.12
				}})
			}
		}
	}
}

// courtyard paves the court and adds the central pool with its domed pavilion,
// planted trees and resting camels with goods.
func (c Caravanserai) courtyard(g *Grid) {
	const (
		F, C = caravanseraiFloor, caravanseraiCourt
	)
	g.Paint(Sand, c.square(C, F-0.6, F+0.3))
	g.Paint(Terracotta, Func{c.square(C, F-0.6, F+0.3).Bounds(), func(p Vec3) bool {
		return math.Mod(p.X+60, 3.2) < 0.7 || math.Mod(p.Z+60, 3.2) < 0.7
	}})
	g.Paint(Stone, Func{c.square(C, F-0.6, F+0.3).Bounds(), func(p Vec3) bool {
		return math.Abs(p.X) < 1.8 || math.Abs(p.Z) < 1.8 || math.Abs(math.Abs(p.X)-C+0.8) < 0.8 || math.Abs(math.Abs(p.Z)-C+0.8) < 0.8
	}})
	g.Paint(DarkStone, Func{c.square(C, F-0.6, F+0.3).Bounds(), func(p Vec3) bool { return c.cellHash(p, 31) < 0.05 }})

	// Pool with a marble rim around an island pavilion reached by four causeways.
	const pool = 7.2
	g.Carve(c.square(pool, F-0.25, F+1))
	g.Paint(Water, c.square(pool, F-1.5, F-0.25))
	g.Use(White).Fill(Subtract(c.square(pool+0.8, F-0.5, F+0.7), c.square(pool, F-1, F+1)))
	g.Use(Stone).Fill(c.square(3.6, F-1, F+0.6))
	for k := 0; k < 4; k++ {
		g.Fill(c.side(Block(V(-1.2, F-1, 3.6), V(1.2, F+0.5, pool)), k))
	}
	// Pavilion: four piers under pointed arches and a turquoise dome.
	g.Use(Sand).Fill(Subtract(c.square(3.2, F+0.6, F+6.4), c.square(2.0, F, F+5.6)))
	g.Carve(c.pointed(0, F+0.6, F+3.2, 1.7, -4, 4), c.side(c.pointed(0, F+0.6, F+3.2, 1.7, -4, 4), 1))
	g.Paint(Brick, Func{c.square(3.3, F+0.6, F+6.4).Bounds(), func(p Vec3) bool { return math.Mod(p.Y-F, 1.8) < 0.4 }})
	g.Paint(Navy, Subtract(c.square(3.3, F+5.6, F+6.4), c.square(2.9, F, F+7)))
	g.Use(Sand).FillVolume(Cylinder{Center: P(0, 0), Radius: 2.6, Y0: F + 6.4, Y1: F + 7.4})
	g.Paint(White, Cylinder{Center: P(0, 0), Radius: 2.7, Y0: F + 7.0, Y1: F + 7.4})
	g.Use(Glass).FillVolume(Intersect(Ellipsoid{V(0, F+7.4, 0), V(2.9, 3.6, 2.9)}, Block(V(-3, F+7.4, -3), V(3, F+11.2, 3))))
	g.Paint(Navy, Func{Box{V(-3, F+7.4, -3), V(3, F+11.2, 3)}, func(p Vec3) bool {
		return math.Mod(p.Plan().Angle()/(2*math.Pi)*10+20, 1) < 0.28
	}})
	g.Use(Gold).Fill(Cylinder{Center: P(0, 0), Radius: 0.5, Y0: F + 10.6, Y1: F + 12.4})
	g.FillVolume(Sphere(V(0, F+11.6, 0), 0.7))

	// Planted trees in the four quarters.
	for i, at := range []Vec2{P(-12.5, -12.5), P(12.5, -12.5), P(-12.5, 12.5), P(12.5, 12.5)} {
		g.Use(Brick).Fill(Cylinder{Center: at, Radius: 1.7, Inner: 0.9, Y0: F, Y1: F + 0.8})
		g.Use(Wood).Fill(Capsule{at.At(F), at.At(F + 3.6), 0.5})
		r := 2.5 + 0.3*Hash01(i, 40)
		g.Use(Foliage).FillVolume(Subtract(Ellipsoid{at.At(F + 4.6), V(r, r*0.8, r)}, Ellipsoid{at.At(F + 4.6), V(r-1.2, r*0.8-1.2, r-1.2)}))
		g.Paint(Grass, Func{Box{V(at.X-r, F+3, at.Y-r), V(at.X+r, F+7, at.Y+r)}, func(p Vec3) bool { return c.cellHash(p, 41) < 0.3 }})
	}
	// Camels resting in the court with their loads.
	c.camel(g, P(-10, 4.5), 0.4, true, Sand, Red, Gold)
	c.camel(g, P(9.5, -5), math.Pi+0.3, true, Wood, Navy, Red)
	c.goods(g, P(-11, -4), 1)
	c.goods(g, P(11, 5.5), 2)
	c.person(g, P(-7.5, 7), Navy, White)
	c.person(g, P(8, -8), Red, White)
}

// camel places a dromedary (local heading 0 faces +X) on the ground, standing or
// kneeling, with a saddle blanket, side packs and a rolled carpet on top.
func (c Caravanserai) camel(g *Grid, pos Vec2, heading float64, kneel bool, body, blanket, pack Color) {
	const k = 1.3 // Model scale.
	y := c.ground(pos.X, pos.Y)
	if math.Max(math.Abs(pos.X), math.Abs(pos.Y)) < caravanseraiCourt {
		y = caravanseraiFloor
	}
	if kneel {
		y -= 1.25 * k
	}
	at := func(x, h, z float64) Vec3 { return V(x*k, y+h*k, z*k) }
	pl := func(s Solid) Solid { return c.place(s, pos, heading) }
	g.With(body, func() {
		g.FillVolume(pl(Union(
			Ellipsoid{at(0, 2.5, 0), V(1.8, 0.85, 0.8).Scale(k)},
			Ellipsoid{at(-0.1, 3.2, 0), V(0.85, 0.75, 0.6).Scale(k)},
			Capsule{at(1.4, 2.7, 0), at(2.3, 2.5, 0), 0.45 * k},
			Capsule{at(2.3, 2.5, 0), at(2.75, 3.55, 0), 0.4 * k},
			Ellipsoid{at(3.05, 3.6, 0), V(0.72, 0.36, 0.34).Scale(k)},
		)))
		g.Fill(pl(Capsule{at(-1.7, 2.6, 0), at(-2.0, 1.7, 0), 0.32 * k}))
		if kneel {
			g.FillVolume(pl(Ellipsoid{at(0, 1.75, 0), V(1.9, 0.6, 1.0).Scale(k)}))
			return
		}
		for _, lx := range []float64{-1.15, 1.15} {
			for _, lz := range []float64{-0.42, 0.42} {
				g.Fill(pl(Capsule{at(lx, 2.0, lz), at(lx+0.1, 0.1, lz), 0.34 * k}))
			}
		}
	})
	// Saddle blanket draped over the hump, with a gold fringe.
	drape := Intersect(Ellipsoid{at(-0.1, 2.75, 0), V(1.25, 1.35, 1.1).Scale(k)}, Block(at(-0.7, 1.9, -1.2), at(0.5, 4.2, 1.2)))
	g.Paint(blanket, pl(drape))
	g.Paint(Gold, pl(Intersect(drape, Block(at(-0.7, 1.9, -1.2), at(0.5, 2.15, 1.2)))))
	g.With(pack, func() {
		g.FillVolume(pl(Ellipsoid{at(-0.15, 2.3, 1.05), V(0.65, 0.6, 0.38).Scale(k)}))
		g.FillVolume(pl(Ellipsoid{at(-0.15, 2.3, -1.05), V(0.65, 0.6, 0.38).Scale(k)}))
	})
	roll := Red
	if blanket == Red {
		roll = Navy
	}
	g.With(roll, func() { g.Fill(pl(Capsule{at(-0.2, 4.05, -0.85), at(-0.2, 4.05, 0.85), 0.36 * k})) })
}

// person is a robed figure with a turban.
func (c Caravanserai) person(g *Grid, pos Vec2, robe, head Color) {
	y := c.ground(pos.X, pos.Y)
	if math.Max(math.Abs(pos.X), math.Abs(pos.Y)) < caravanseraiCourt {
		y = caravanseraiFloor
	}
	g.With(robe, func() { g.FillVolume(Frustum{pos, 0.75, 0.5, y - 0.2, y + 2.6}) })
	g.With(Terracotta, func() { g.FillVolume(Sphere(pos.At(y+3.0), 0.5)) })
	g.With(head, func() { g.FillVolume(Ellipsoid{pos.At(y + 3.4), V(0.6, 0.35, 0.6)}) })
}

// goods is a merchant's pile: stacked carpets with borders, bales and jars.
func (c Caravanserai) goods(g *Grid, at Vec2, seed int) {
	y := c.ground(at.X, at.Y)
	if math.Max(math.Abs(at.X), math.Abs(at.Y)) < caravanseraiCourt {
		y = caravanseraiFloor
	}
	cols := []Color{Red, Navy, Gold, Red, Brick}
	for i := 0; i < 3; i++ {
		col := cols[(i+seed)%len(cols)]
		border := Gold
		if col == Gold {
			border = Red
		}
		y0 := y + 0.6*float64(i)
		jx := 0.4 * (Hash01(seed, i, 1) - 0.5)
		rug := Block(V(at.X-1.7+jx, y0, at.Y-1.2), V(at.X+1.7+jx, y0+0.6, at.Y+1.2))
		g.Use(col).Fill(rug)
		g.Paint(border, Subtract(rug, Block(V(at.X-1.3+jx, y0-1, at.Y-0.8), V(at.X+1.3+jx, y0+2, at.Y+0.8))))
	}
	// Bales and a jar beside the stack.
	for i := 0; i < 3; i++ {
		b := at.Add(P(2.6, -1.3+1.3*float64(i)))
		col := Wood
		if Hash01(seed, i, 2) < 0.5 {
			col = Sand
		}
		g.Use(col).FillVolume(Ellipsoid{b.At(y + 0.6), V(0.8, 0.75, 0.65)})
	}
	g.Use(Wood).FillVolume(Ellipsoid{at.Add(P(2.6, 0)).At(y + 1.8), V(0.75, 0.6, 0.6)})
	jar := at.Add(P(-2.6, 0.6))
	g.Use(Terracotta).FillVolume(Ellipsoid{jar.At(y + 0.9), V(0.7, 0.9, 0.7)})
	g.Fill(Cylinder{Center: jar, Radius: 0.4, Y0: y + 1.6, Y1: y + 2.1})
}

// palm is a date palm: a leaning ringed trunk, arching fronds and date clusters.
func (c Caravanserai) palm(g *Grid, base Vec2, seed int) {
	y := c.ground(base.X, base.Y)
	h := 7 + 3*Hash01(seed, 1)
	lean := Polar(P(0, 0), 0.5+1.0*Hash01(seed, 2), Hash01(seed, 3)*2*math.Pi)
	mid := base.Add(lean.Scale(0.3)).At(y + h*0.55)
	top := base.Add(lean).At(y + h)
	g.Use(Wood).Fill(Capsule{base.At(y - 0.5), mid, 0.5}, Capsule{mid, top, 0.45})
	g.Paint(Brick, Func{Box{V(base.X-3, y, base.Y-3), V(base.X+3, y+h, base.Y+3)}, func(p Vec3) bool {
		return math.Mod(p.Y-y, 1.0) < 0.3
	}})
	g.Use(Foliage).FillVolume(Sphere(top, 0.8))
	for k := 0; k < 9; k++ {
		a := (float64(k) + 0.6*Hash01(seed, k, 4)) * 2 * math.Pi / 9
		m := Polar(top.Plan(), 1.7, a).At(top.Y + 0.6)
		tip := Polar(top.Plan(), 3.4, a).At(top.Y - 1.1)
		col := Foliage
		if k%3 == 0 {
			col = Grass
		}
		g.Use(col).Fill(Capsule{top, m, 0.45}, Capsule{m, tip, 0.4})
	}
	for k := 0; k < 3; k++ {
		at := Polar(top.Plan(), 0.9, float64(k)*2.1+Hash01(seed, 5))
		g.Use(Gold).FillVolume(Ellipsoid{at.At(top.Y - 0.9), V(0.45, 0.6, 0.45)})
	}
}

// tent is a black goat-hair tent with peaked ridge, red woven bands, open front
// (local +Z), side poles and a rug at the door.
func (c Caravanserai) tent(g *Grid, pos Vec2, heading float64, w, d float64) {
	y := c.ground(pos.X, pos.Y)
	top := func(x, z float64) float64 {
		return y + 2.2 + 0.7*(0.5-0.5*math.Cos(4*math.Pi*x/w)) - 0.7*math.Abs(z)/(d/2)
	}
	pl := func(s Solid) Solid { return c.place(s, pos, heading) }
	shape := Func{Box{V(-w/2, y-0.5, -d/2), V(w/2, y+3.2, d/2)}, func(p Vec3) bool {
		t := top(p.X, p.Z)
		if p.Y > t {
			return false
		}
		if p.Y >= t-0.7 {
			return true
		}
		return p.Z <= -d/2+0.6 || math.Abs(p.X) >= w/2-0.6
	}}
	g.Use(Charcoal).Fill(pl(shape))
	g.Paint(Red, pl(Func{shape.Box, func(p Vec3) bool {
		return math.Mod(p.X+w, 2.4) < 0.5
	}}))
	g.Paint(White, pl(Func{shape.Box, func(p Vec3) bool { return p.Z > d/2-0.6 && p.Y > top(p.X, p.Z)-0.7 }}))
	g.Use(Wood)
	for _, x := range []float64{-w / 2, -w / 4, w / 4, w / 2} {
		g.Fill(pl(Capsule{V(x*0.95, y-0.3, d/2-0.2), V(x*0.95, top(x, d/2)+0.2, d/2-0.2), 0.35}))
	}
	g.Use(Red).Fill(pl(Block(V(-w/4, y-0.4, d/2), V(w/4, y+0.15, d/2+2.4))))
	g.Paint(Gold, pl(Subtract(Block(V(-w/4, y-0.4, d/2), V(w/4, y+0.15, d/2+2.4)), Block(V(-w/4+0.6, y-1, d/2+0.6), V(w/4-0.6, y+1, d/2+1.8)))))
}

// well is a round stone well head with a wooden frame and pulley beam.
func (c Caravanserai) well(g *Grid, at Vec2) {
	y := c.ground(at.X, at.Y)
	g.Use(Stone).FillVolume(Cylinder{Center: at, Radius: 1.7, Inner: 0.8, Y0: y - 0.5, Y1: y + 1.2})
	g.Paint(Charcoal, Cylinder{Center: at, Radius: 1.2, Y0: y - 2, Y1: y + 0.1})
	g.Use(Wood).Fill(Capsule{at.Add(P(-1.5, 0)).At(y + 1), at.Add(P(-1.5, 0)).At(y + 3.6), 0.4},
		Capsule{at.Add(P(1.5, 0)).At(y + 1), at.Add(P(1.5, 0)).At(y + 3.6), 0.4},
		Capsule{at.Add(P(-1.7, 0)).At(y + 3.6), at.Add(P(1.7, 0)).At(y + 3.6), 0.4})
}

// cistern is an ab anbar: a stepped pointed dome over a sunken tank, an entrance
// portal toward +Z and two wind catchers (badgirs) cooling the water.
func (c Caravanserai) cistern(g *Grid, at Vec2) {
	y := c.ground(at.X, at.Y)
	const R, dh = 5.0, 7.5
	g.Use(Sand).FillVolume(Cylinder{Center: at, Radius: R + 0.8, Inner: R - 1.5, Y0: y - 1, Y1: y + 0.6})
	g.FillVolume(Cylinder{Center: at, Radius: R, Inner: R - 1.2, Y0: y + 0.6, Y1: y + 2.6})
	prof := func(u, inset float64) float64 { return R*math.Pow(math.Max(0, 1-u*u), 0.7) - inset }
	g.FillVolume(Func{Box{V(at.X-R, y+2.6, at.Y-R), V(at.X+R, y+2.6+dh, at.Y+R)}, func(p Vec3) bool {
		u := (p.Y - y - 2.6) / dh
		l := p.Plan().Sub(at).Length()
		return l <= prof(u, 0) && l > prof(u, 1.2)
	}})
	g.Paint(Brick, Func{Box{V(at.X-R-1, y, at.Y-R-1), V(at.X+R+1, y+dh+4, at.Y+R+1)}, func(p Vec3) bool {
		return math.Mod(p.Y-y, 1.3) < 0.4
	}})
	g.Use(Sand).Fill(Cylinder{Center: at, Radius: 0.6, Y0: y + dh + 2.2, Y1: y + dh + 3.4})
	// Entrance portal with a dark stair arch.
	pz := at.Y + R - 0.3
	g.Use(Sand).Fill(Block(V(at.X-2.6, y, pz-1), V(at.X+2.6, y+6.2, pz+1.6)))
	arch := Translate(c.pointed(0, y, y+2.6, 1.5, -2, 4), V(at.X, 0, pz-1.6))
	g.Carve(arch)
	g.Paint(Charcoal, Translate(c.pointed(0, y-1, y+2.6, 2.1, -3, 2.0), V(at.X, 0, pz-1.6)))
	g.Paint(Navy, Block(V(at.X-2.6, y+5.6, pz+1.1), V(at.X+2.6, y+6.2, pz+1.7)))
	g.Paint(Water, Subtract(Translate(c.pointed(0, y, y+2.6, 2.1, 1.2, 1.9), V(at.X, 0, pz)), Translate(c.pointed(0, y-1, y+2.6, 1.5, 0, 3), V(at.X, 0, pz))))

	// Badgirs: tall square towers with slatted vents near the top.
	for i, b := range []Vec2{at.Add(P(7.2, 0.5)), at.Add(P(-5.5, -5.5))} {
		by := c.ground(b.X, b.Y)
		top := y + 12.5 + float64(i)*1.2
		g.Use(Sand).Fill(Subtract(BlockAt(b, 2.8, 2.8, by-0.8, top), BlockAt(b, 1.2, 1.2, by, top-0.8)))
		for _, ax := range []float64{-0.7, 0.7} {
			g.Carve(BlockAt(b.Add(P(ax, 0)), 0.75, 4, top-4.4, top-1.2), BlockAt(b.Add(P(0, ax)), 4, 0.75, top-4.4, top-1.2))
		}
		g.Paint(Brick, Func{BlockAt(b, 3, 3, by, top).Bounds(), func(p Vec3) bool { return math.Mod(p.Y-by, 2.0) < 0.4 }})
		g.Paint(Charcoal, BlockAt(b, 1.9, 1.9, by, top-0.5))
		g.Use(Sand).Fill(BlockAt(b, 3.4, 3.4, top, top+0.6))
		g.Paint(Brick, BlockAt(b, 3.4, 3.4, top+0.3, top+0.6))
		g.Use(Sand).Fill(Capsule{b.Add(at.Sub(b).Scale(0.35)).At(y + 0.5), b.At(by + 0.5), 1.1})
	}
}

// outside places the arriving caravan and its track, tethered camels, goods by the
// gate, tents, palms and the qanat line.
func (c Caravanserai) outside(g *Grid) {
	const H = caravanseraiHalf
	track := []Vec2{P(0, H+2.5), P(2.5, 36), P(9, 41.5), P(19, 43), P(30, 44), P(41, 44.5)}
	dist := func(p Vec2) float64 {
		best := 1e9
		for i := 1; i < len(track); i++ {
			a, b := track[i-1], track[i]
			ab := b.Sub(a)
			t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
			best = math.Min(best, p.Sub(a.Add(ab.Scale(t))).Length())
		}
		return best
	}
	g.Paint(Stone, Func{Box{V(-6, 0, H), V(50, 4, 52)}, func(p Vec3) bool {
		d := dist(p.Plan())
		return d < 1.7 || (d < 2.4 && c.cellHash(p, 50) < 0.5)
	}})
	// The caravan along the track, heading for the gate.
	along := func(s float64) (Vec2, float64) {
		for i := 1; i < len(track); i++ {
			a, b := track[i-1], track[i]
			l := b.Sub(a).Length()
			if s <= l || i == len(track)-1 {
				p := a.Add(b.Sub(a).Scale(s / l))
				return p, a.Sub(b).Angle()
			}
			s -= l
		}
		return track[0], 0
	}
	type load struct{ body, blanket, pack Color }
	loads := []load{{Sand, Red, Gold}, {Wood, Navy, Red}, {Sand, Gold, Navy}, {Wood, Red, Sand}, {Sand, Navy, Red}, {Wood, Gold, Navy}}
	for i, l := range loads {
		p, h := along(11 + 5.6*float64(i))
		side := P(-math.Sin(h), math.Cos(h)).Scale(0.15 * float64(i%2))
		c.camel(g, p.Add(side), -h, false, l.body, l.blanket, l.pack)
	}
	lead, lh := along(8)
	c.person(g, lead.Add(Polar(P(0, 0), 1.4, lh+math.Pi/2)), Navy, White)
	mid, mh := along(25)
	c.person(g, mid.Add(Polar(P(0, 0), 1.9, mh-math.Pi/2)), Red, White)
	tail, th := along(43)
	c.person(g, tail.Add(Polar(P(0, 0), 1.6, th+math.Pi/2)), Wood, White)

	// Tethered camels at a hitching rail and trough along the east wall.
	g.Use(Wood)
	for z := 3.0; z <= 23; z += 5 {
		y := c.ground(37.5, z)
		g.Fill(Cylinder{Center: P(37.5, z), Radius: 0.4, Y0: y - 0.5, Y1: y + 2.2})
	}
	g.Fill(Capsule{V(37.5, c.ground(37.5, 3)+1.9, 3), V(37.5, c.ground(37.5, 23)+1.9, 23), 0.4})
	g.Use(Stone).Fill(Block(V(31.2, caravanseraiFloor-0.5, 9), V(32.8, caravanseraiFloor+0.9, 16)))
	g.Paint(Water, Block(V(31.6, caravanseraiFloor+0.5, 9.4), V(32.4, caravanseraiFloor+0.9, 15.6)))
	c.camel(g, P(34.5, 5.5), math.Pi/2+0.2, true, Sand, Red, Gold)
	c.camel(g, P(35, 12.5), -math.Pi/2-0.3, false, Wood, Gold, Navy)
	c.camel(g, P(34.5, 19.5), math.Pi/2-0.15, true, Sand, Navy, Red)
	c.person(g, P(33.5, 24), Navy, White)

	// Merchants' carpets and bales either side of the gate.
	c.goods(g, P(-13, 34.5), 0)
	c.goods(g, P(-12.5, 40), 3)
	c.goods(g, P(14, 33), 1)
	c.person(g, P(-9.5, 37.5), Red, White)
	c.person(g, P(-16.5, 37), Brick, White)
	c.person(g, P(10.5, 34.5), Navy, Gold)

	// Tents of the camel drivers.
	c.tent(g, P(37, -4), -math.Pi/2, 7, 4.5)
	c.tent(g, P(36, -15.5), -math.Pi/2, 6, 4)
	c.person(g, P(41.5, -6), Navy, White)
	c.tent(g, P(36.5, 35), 0.15, 6.5, 4.2)
	c.person(g, P(34, 40.5), Brick, White)
	c.well(g, P(27.5, 36))

	// Date palms around the cistern, the gate and the camp.
	for i, p := range []Vec2{P(-26, 42.5), P(-20.5, 39), P(-42, 23), P(-40, 16), P(-31, 44), P(-17, 44),
		P(-21, 30.5), P(15.5, 38.5), P(21, 37.5), P(40.5, 15), P(41, 4), P(40, 25), P(-44, 30)} {
		c.palm(g, p, i)
	}

	// Qanat: a line of shaft mounds bringing water down to the cistern.
	for z := -30.0; z <= 18; z += 6 {
		x := -39 + 1.2*math.Sin(z*0.15)
		y := c.ground(x, z)
		at := P(x, z)
		g.Use(Sand).FillVolume(Subtract(Frustum{at, 2.2, 0.9, y - 0.6, y + 1.1}, Cylinder{Center: at, Radius: 0.6, Y0: y - 0.3, Y1: y + 2}))
		g.Paint(Terracotta, Frustum{at, 2.3, 1.2, y - 0.6, y + 0.4})
		g.Paint(Charcoal, Cylinder{Center: at, Radius: 1.0, Y0: y - 1.5, Y1: y - 0.1})
	}
}
