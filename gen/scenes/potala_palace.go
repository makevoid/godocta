package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// PotalaPalace is the Potala in Lhasa climbing the south face of the Red Hill (Marpo
// Ri). The façade faces +Z. Massive white battered foundation walls carry the long
// zig-zag stairways; above them the White Palace wings (east, +X, and west) rise in
// layered storeys of small dark trapezoidal windows under a dark penbe frieze, with the
// Red Palace towering in the middle beneath its gold double-eaved roof pavilions and
// gilded victory banners. Round fortified bastions close both ends. The rocky hill
// wraps round the back and sides; the walled Shöl village lies at the foot with its
// gate tower, then the paved plaza with poplars, and the white chorten gate stands on
// the road to the west. The Lhasa valley floor eases down to a low rim.
type PotalaPalace struct{}

func init() { Register(PotalaPalace{}) }

func (PotalaPalace) Name() string  { return "potala_palace" }
func (PotalaPalace) Title() string { return "Potala Palace" }

const (
	potalaPlain = 1.6 // Valley floor level.
)

// Terrain kinds.
const (
	potalaPlainKind = iota
	potalaHillKind
	potalaVillageKind
	potalaPlazaKind
	potalaRoadKind
)

func (PotalaPalace) smooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func (PotalaPalace) noise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

func (m PotalaPalace) fbm(x, z float64, seed int) float64 {
	return (m.noise(x, z, seed)*4 + m.noise(x*2.1+5, z*2.1+3, seed+1)*2 + m.noise(x*4.3+1, z*4.3+7, seed+2)) / 7
}

// rim is the normalized site radius; above 1 is off the plot.
func (PotalaPalace) rim(x, z float64) float64 {
	ex, ez := math.Abs(x+2)/58, math.Abs(z-3)/42
	return math.Pow(ex*ex*ex*ex+ez*ez*ez*ez, 0.25)
}

// hill is the Red Hill's height above the plain.
func (m PotalaPalace) hill(x, z float64) float64 {
	dx, dz := math.Abs(x+2)/66, math.Abs(z+8)/22
	r := math.Cbrt(dx*dx*dx+dz*dz*dz) * (1 + 0.22*(m.fbm(x/7, z/7, 3)-0.5))
	if r >= 1 {
		return 0
	}
	h := 24 * math.Pow(1-r*r, 1.15)
	// Rocky ledges.
	h += 1.6 * (m.fbm(x/4, z/4, 5) - 0.5) * m.smooth(h/4)
	// Keep the foot of the façade and its stairways clear of rock.
	h *= 1 - m.smooth((z-7)/4)*m.smooth((41-math.Abs(x+2))/5)
	return math.Max(0, h)
}

// height is the ground height and kind of a column; ok is false off the plot.
func (m PotalaPalace) height(x, z float64) (float64, int, bool) {
	e := m.rim(x, z)
	if e > 1 {
		return 0, potalaPlainKind, false
	}
	base := 0.4 + (potalaPlain-0.4)*m.smooth((1-e)/0.1)
	h, kind := base, potalaPlainKind
	if hh := m.hill(x, z); hh > 0.3 {
		h, kind = base+hh, potalaHillKind
	}
	switch {
	case x > -47 && x < 45.5 && z > 7 && z < 29:
		if kind != potalaHillKind {
			kind = potalaVillageKind
		}
	case x > -38 && x < 36 && z > 31 && z < 38:
		kind = potalaPlazaKind
	case x < -38 && math.Abs(z-34.5) < 2.2:
		kind = potalaRoadKind
	}
	return h, kind, true
}

func (m PotalaPalace) ground(x, z float64) float64 {
	h, _, _ := m.height(x, z)
	return h
}

func (m PotalaPalace) Build(g *Grid) Order {
	m.terrain(g)
	blocks := m.blocks()
	for _, b := range blocks {
		g.Carve(b.outer(0.3, 0))
	}
	for _, b := range blocks {
		b.shell(g)
	}
	for _, b := range blocks {
		b.finish(g)
		if b.color == Red {
			b.ornaments(g, Gold, (b.bandA+b.bandB)/2, 2.4)
		}
		d := 0.9 + b.batter*(b.y1-b.y0)
		corners := []Vec2{P(b.x0+d, b.z1-d), P(b.x1-d, b.z1-d), P(b.x0+d, b.z0+d), P(b.x1-d, b.z0+d)}
		for _, c := range corners[:b.banners] {
			m.banner(g, c.At(b.y1))
		}
	}
	m.bastion(g, P(42.5, 4), 5.8)
	m.bastion(g, P(-45.5, 3), 5.8)
	m.stairs(g)
	m.roofs(g)
	m.shol(g)
	m.plaza(g)
	m.chorten(g, P(-50, 34.5))
	return Outward(P(-2, 0))
}

// terrain fills the heightfield as a hollow skin, column by column.
func (m PotalaPalace) terrain(g *Grid) {
	half := g.Span() / 2
	cx0, cx1 := int(math.Ceil(-61/half)), int(math.Floor(57/half))
	cz0, cz1 := int(math.Ceil(-40/half)), int(math.Floor(46/half))
	for cz := cz0; cz <= cz1; cz++ {
		for cx := cx0; cx <= cx1; cx++ {
			if (cx^cz)&1 != 0 {
				continue
			}
			x, z := float64(cx)*half, float64(cz)*half
			h, kind, ok := m.height(x, z)
			if !ok {
				continue
			}
			low := h
			for k := 0; k < 8; k++ {
				a := float64(k) * math.Pi / 4
				low = math.Min(low, m.ground(x+1.1*math.Cos(a), z+1.1*math.Sin(a)))
			}
			slope := (h - low) / 1.1
			y0 := math.Max(0, low-0.8)
			cy0 := int(math.Ceil(y0 / half))
			if (cy0^cx)&1 != 0 {
				cy0++
			}
			for cy := cy0; float64(cy)*half <= h; cy += 2 {
				y := float64(cy) * half
				var c Color
				switch kind {
				case potalaHillKind:
					n := m.noise(x/3+y/5, z/3-y/6, 31)
					switch {
					case slope > 1.3 || n > 0.66:
						c = DarkStone
						if m.noise(x/2+y/3, z/2-y/4, 33) > 0.55 {
							c = Stone
						}
					case m.fbm(x/5, z/5, 35) > 0.6 && slope < 1.0:
						c = Grass
					case m.fbm(x/6+3, z/6, 37) > 0.5:
						c = Sand
					default:
						c = Stone
					}
				case potalaVillageKind:
					c = Sand
					if Hash01(cx, cz, 39) < 0.15 {
						c = Stone
					}
				case potalaPlazaKind:
					c = Stone
					if math.Mod(math.Abs(x+0.5), 5) < 0.6 || math.Mod(math.Abs(z-31), 5) < 0.6 {
						c = DarkStone
					}
				case potalaRoadKind:
					c = Sand
				default:
					c = Grass
					if m.fbm(x/7, z/7, 41) > 0.55 {
						c = Sand
					}
					if y < h-0.5 {
						c = Sand
					}
				}
				g.Use(c)
				g.Add(Cell{int32(cx), int32(cy), int32(cz)})
			}
		}
	}
}

// potalaBlock is one hollow battered building mass with a flat roof and parapet.
type potalaBlock struct {
	x0, x1, z0, z1, y0, y1 float64
	batter, wall           float64
	color                  Color
	rows                   []float64 // Window sill heights.
	winW, winH, step       float64
	faces                  [4]bool // +Z, -Z, +X, -X.
	bandA, bandB           float64 // Dark frieze band; zero height disables it.
	band                   Color
	sparse                 float64 // Fraction of windows left out.
	seed                   int
	banners                int // Gilded banners on the front corners (2) or all four (4).
}

func (b *potalaBlock) roofY() float64 { return b.y1 - 0.8 }

// inset is the block's battered volume shrunk by d on every side, between ya and yb.
func (b *potalaBlock) inset(d, ya, yb float64) Solid {
	return Func{Box{V(b.x0+d, ya, b.z0+d), V(b.x1-d, yb, b.z1-d)}, func(p Vec3) bool {
		s := d + b.batter*(p.Y-b.y0)
		return p.X >= b.x0+s && p.X <= b.x1-s && p.Z >= b.z0+s && p.Z <= b.z1-s
	}}
}

func (b *potalaBlock) outer(lift, extra float64) Solid {
	return b.inset(0, b.y0+lift, b.y1+extra)
}

func (b *potalaBlock) shell(g *Grid) {
	t := b.wall
	g.Use(b.color).Fill(Subtract(b.outer(0, 0),
		b.inset(t, b.y0+0.7, b.roofY()-t),
		b.inset(t, b.roofY(), b.y1+1)))
}

// local maps a point to the along-face coordinate and outward distance from a face.
func (b *potalaBlock) local(side int, p Vec3) (u, out float64) {
	d := b.batter * (p.Y - b.y0)
	switch side {
	case 0:
		return p.X, p.Z - (b.z1 - d)
	case 1:
		return p.X, (b.z0 + d) - p.Z
	case 2:
		return p.Z, p.X - (b.x1 - d)
	}
	return p.Z, (b.x0 + d) - p.X
}

func (b *potalaBlock) point(side int, u, y, out float64) Vec3 {
	d := b.batter * (y - b.y0)
	switch side {
	case 0:
		return V(u, y, b.z1-d+out)
	case 1:
		return V(u, y, b.z0+d-out)
	case 2:
		return V(b.x1-d+out, y, u)
	}
	return V(b.x0+d-out, y, u)
}

// opening is a trapezoid on a face: wB wide at ya narrowing to wT at yb.
func (b *potalaBlock) opening(side int, uc, ya, yb, wB, wT, o0, o1 float64) Solid {
	box := EmptyBox()
	w := math.Max(wB, wT) / 2
	for _, u := range []float64{uc - w, uc + w} {
		for _, y := range []float64{ya, yb} {
			for _, o := range []float64{o0, o1} {
				q := b.point(side, u, y, o)
				box = box.Union(Box{q, q})
			}
		}
	}
	return Func{box, func(p Vec3) bool {
		u, out := b.local(side, p)
		if out < o0 || out > o1 || p.Y < ya || p.Y > yb {
			return false
		}
		t := (p.Y - ya) / (yb - ya)
		return math.Abs(u-uc) <= (wB+(wT-wB)*t)/2
	}}
}

// finish darkens the interior, paints the frieze and roof and cuts the windows.
func (b *potalaBlock) finish(g *Grid) {
	t := b.wall
	g.Paint(Charcoal, b.inset(t-0.75, b.y0+0.7-0.75, b.roofY()-t+0.75))
	g.Paint(DarkStone, b.inset(t, b.roofY()-0.45, b.roofY()+0.1))
	if b.bandB > b.bandA {
		g.Paint(b.band, Subtract(b.inset(0, b.bandA, b.bandB), b.inset(t+0.05, b.bandA-1, b.bandB+1)))
	}
	for side := 0; side < 4; side++ {
		if !b.faces[side] {
			continue
		}
		u0, u1 := b.x0, b.x1
		if side >= 2 {
			u0, u1 = b.z0, b.z1
		}
		for r, sill := range b.rows {
			d := b.batter*(sill+b.winH-b.y0) + 1.3 + b.winW/2
			n := int(math.Floor((u1 - u0 - 2*d) / b.step))
			if n < 0 {
				continue
			}
			start := (u0+u1)/2 - float64(n)*b.step/2
			for i := 0; i <= n; i++ {
				uc := start + float64(i)*b.step
				if Hash01(b.seed, side, r, i) < b.sparse {
					continue
				}
				if g.HasAt(b.point(side, uc, sill+b.winH/2, 0.9)) || g.HasAt(b.point(side, uc, sill+0.2, 0.9)) {
					continue
				}
				g.Paint(Charcoal, b.opening(side, uc, sill-0.3, sill+b.winH+0.25, b.winW+0.9, b.winW+0.4, -0.7, 0.4))
				g.Carve(b.opening(side, uc, sill, sill+b.winH, b.winW, b.winW, -(t + 0.3), 0.6))
			}
		}
	}
}

func (b *potalaBlock) ornaments(g *Grid, c Color, y, step float64) {
	for side := 0; side < 4; side++ {
		if !b.faces[side] {
			continue
		}
		u0, u1 := b.x0, b.x1
		if side >= 2 {
			u0, u1 = b.z0, b.z1
		}
		d := b.batter*(y-b.y0) + 1.2
		n := int(math.Floor((u1 - u0 - 2*d) / step))
		start := (u0+u1)/2 - float64(n)*step/2
		for i := 0; i <= n; i++ {
			uc := start + float64(i)*step
			g.Paint(c, b.opening(side, uc, y-0.4, y+0.4, 0.8, 0.8, -0.5, 0.3))
		}
	}
}

func (PotalaPalace) rows(from, step float64, n int) []float64 {
	r := make([]float64, n)
	for i := range r {
		r[i] = from + float64(i)*step
	}
	return r
}

// blocks lists the palace masses, lowest first.
func (m PotalaPalace) blocks() []*potalaBlock {
	front := [4]bool{true, false, false, false}
	all := [4]bool{true, true, true, true}
	white := func(x0, x1, z0, z1, y0, y1 float64, rows []float64, faces [4]bool, seed int) *potalaBlock {
		return &potalaBlock{x0: x0, x1: x1, z0: z0, z1: z1, y0: y0, y1: y1, batter: 0.07, wall: 1.3,
			color: White, rows: rows, winW: 1.3, winH: 1.0, step: 2.6, faces: faces,
			bandA: y1 - 2.5, bandB: y1 - 1.1, band: Wood, seed: seed}
	}
	var bs []*potalaBlock
	// Foundations: tall plain white walls with sparse small windows.
	f3 := white(-42, -18, -4, 10, 0.8, 10, m.rows(2.8, 2.6, 3), [4]bool{true, false, false, true}, 1)
	f2 := white(-18, 12, -6, 12, 0.8, 15, m.rows(3, 2.6, 4), front, 2)
	f1 := white(12, 39, -4, 11, 0.8, 12, m.rows(2.8, 2.6, 4), [4]bool{true, false, true, false}, 3)
	for _, f := range []*potalaBlock{f3, f2, f1} {
		f.winW, f.winH, f.step, f.sparse, f.bandB, f.batter = 0.8, 0.9, 3.0, 0.5, 0, 0.1
	}
	bs = append(bs, f3, f2, f1)
	// East White Palace stepping back in three ranges.
	e := white(12, 37, -6, 7.5, 12, 23.5, m.rows(13.2, 2.6, 3), [4]bool{true, true, true, false}, 4)
	eu := white(15, 35, -6, 3.5, 23.5, 27.5, []float64{24.3}, [4]bool{true, false, true, false}, 11)
	e2 := white(19, 32, -5, -0.5, 27.5, 31, []float64{28.2}, [4]bool{true, false, true, true}, 5)
	e2.winW, e2.winH, e2.step = 1.6, 1.4, 2.6
	// West White Palace (Namgyal monastery side) and the yellow hall.
	w := white(-40, -23, -5, 6.5, 10, 21, m.rows(11.2, 2.6, 3), [4]bool{true, true, false, true}, 6)
	wu := white(-38, -25, -5, 3, 21, 24.5, []float64{21.8}, [4]bool{true, false, false, true}, 12)
	w2 := white(-36, -27, -4, 0, 24.5, 27.8, []float64{25.2}, [4]bool{true, false, true, true}, 7)
	w2.winW, w2.winH, w2.step = 1.6, 1.4, 2.6
	for _, b := range []*potalaBlock{e, eu, e2, w, wu, w2} {
		b.banners = 2
	}
	y := white(-23, -18, -7, 6, 10, 29, m.rows(11.2, 2.6, 6), [4]bool{true, true, false, false}, 8)
	y.color = Sand
	y.step, y.banners = 2.0, 2
	// The Red Palace with its projecting central bay.
	r := white(-18, 12, -11, 9, 15, 36, m.rows(16.2, 2.6, 6), all, 9)
	r.color, r.bandA, r.bandB, r.banners = Red, 31.6, 34.4, 4
	rc := white(-7, 1, 3, 10.5, 15, 33.5, m.rows(16.2, 2.6, 5), [4]bool{true, false, true, true}, 10)
	rc.color, rc.bandA, rc.bandB = Red, 29.8, 32.4
	rc.winW, rc.winH, rc.step, rc.banners = 1.4, 1.6, 2.2, 2
	bs = append(bs, e, eu, e2, w, wu, w2, y, r, rc)
	return bs
}

// bastion is a round battered end tower.
func (m PotalaPalace) bastion(g *Grid, c Vec2, rad float64) {
	y0, y1 := 0.8, 15.0
	r1 := rad - 0.9
	at := func(y float64) float64 { return rad + (r1-rad)*(y-y0)/(y1-y0) }
	ring := func(d, ya, yb float64) Solid {
		return Func{Box{c.At(ya).Sub(V(rad, 0, rad)), c.At(yb).Add(V(rad, 0, rad))}, func(p Vec3) bool {
			return p.Y >= ya && p.Y <= yb && p.Plan().Sub(c).Length() <= at(p.Y)-d
		}}
	}
	roof := y1 - 0.8
	g.Carve(ring(0, y0+0.3, y1))
	g.Use(White).Fill(Subtract(ring(0, y0, y1), ring(1.3, y0+0.7, roof-1.3), ring(1.3, roof, y1+1)))
	g.Paint(Charcoal, ring(0.55, y0, roof-0.55))
	g.Paint(DarkStone, ring(1.3, roof-0.45, roof+0.1))
	g.Paint(Wood, Subtract(ring(0, y1-2.5, y1-1.1), ring(1.35, y0, y1+1)))
	// Slit windows in rows facing out.
	for r, sill := range []float64{3.5, 6.5, 9.5} {
		for k := 0; k < 16; k++ {
			a := (float64(k) + 0.5*float64(r%2)) * 2 * math.Pi / 16
			dir := P(math.Cos(a), math.Sin(a))
			rr := at(sill + 0.6)
			if g.HasAt(c.Add(dir.Scale(rr + 0.9)).At(sill + 0.6)) {
				continue
			}
			win := func(w, ya, yb, o0, o1 float64) Solid {
				q := c.Add(dir.Scale(rr))
				return Func{Box{V(q.X-2, ya, q.Y-2), V(q.X+2, yb, q.Y+2)}, func(p Vec3) bool {
					v := p.Plan().Sub(c)
					o := v.Length() - at(p.Y)
					along := v.Dot(dir)
					across := math.Abs(v.X*dir.Y - v.Y*dir.X)
					return along > 0 && o >= o0 && o <= o1 && across <= w/2 && p.Y >= ya && p.Y <= yb
				}}
			}
			g.Paint(Charcoal, win(1.6, sill-0.4, sill+1.6, -0.7, 0.4))
			g.Carve(win(0.8, sill, sill+1.2, -1.6, 0.6))
		}
	}
	// Victory banner on the parapet.
	m.banner(g, c.At(y1))
}

// banner is a gilded gyaltsen victory banner.
func (PotalaPalace) banner(g *Grid, base Vec3) {
	g.Use(Gold).Fill(
		Cylinder{Center: base.Plan(), Radius: 0.6, Y0: base.Y - 0.3, Y1: base.Y + 2.2},
		Cone(base.Plan(), 0.45, base.Y+2.2, 1.0),
	)
}

// stairs are the zig-zag ramps climbing the foundation walls.
func (m PotalaPalace) stairs(g *Grid) {
	type flight struct {
		xa, xb, ya, yb, zf float64 // Ramp from (xa, ya) to (xb, yb); zf is the wall face at the base.
	}
	flights := []flight{
		{14, 34, potalaPlain, 5.2, 11},
		{34, 16, 5.2, 8.6, 11},
		{16, 34, 8.6, 12, 11},
		{-20, -36, potalaPlain, 4.4, 10},
		{-36, -22, 4.4, 7.2, 10},
		{-22, -36, 7.2, 10, 10},
	}
	const batter = 0.1
	for _, f := range flights {
		lo, hi := math.Min(f.xa, f.xb), math.Max(f.xa, f.xb)
		ramp := func(x float64) float64 { return f.ya + (f.yb-f.ya)*(x-f.xa)/(f.xb-f.xa) }
		face := func(y float64) float64 { return f.zf - batter*(y-0.8) }
		top := math.Max(f.ya, f.yb)
		box := Box{V(lo, 0, f.zf-2), V(hi, top+1, f.zf+2.8)}
		body := Func{box, func(p Vec3) bool {
			r := ramp(p.X)
			dz := p.Z - face(r)
			return dz >= -0.4 && dz <= 2.6 && p.Y >= r-1.6 && p.Y <= r
		}}
		parapet := Func{box, func(p Vec3) bool {
			r := ramp(p.X)
			dz := p.Z - face(r)
			return dz >= 1.8 && dz <= 2.6 && p.Y >= r-1.6 && p.Y <= r+0.9
		}}
		coping := Func{box, func(p Vec3) bool {
			r := ramp(p.X)
			dz := p.Z - face(r)
			return dz >= 1.8 && dz <= 2.6 && p.Y >= r+0.5 && p.Y <= r+0.9
		}}
		tread := Func{box, func(p Vec3) bool {
			r := math.Round(ramp(p.X)/0.4) * 0.4
			dz := p.Z - face(r)
			return dz >= -0.4 && dz < 1.8 && p.Y >= r-0.45 && p.Y <= r+0.05
		}}
		g.Use(White).Fill(body, parapet)
		g.Paint(DarkStone, tread)
		g.Paint(Wood, coping)
	}
	// Landings at the turns.
	landing := func(x0, x1, y, zf float64) {
		z := zf - batter*(y-0.8)
		g.Use(White).Fill(Block(V(x0, y-1.6, z-0.4), V(x1, y, z+2.6)))
		g.Use(White).Fill(Block(V(x0, y-1.6, z+1.8), V(x1, y+0.9, z+2.6)))
		g.Paint(DarkStone, Block(V(x0, y-0.45, z-0.4), V(x1, y+0.05, z+1.75)))
		g.Paint(Wood, Block(V(x0, y+0.5, z+1.8), V(x1, y+0.9, z+2.6)))
	}
	landing(34, 37, 5.2, 11)
	landing(13, 16, 8.6, 11)
	landing(-39, -36, 4.4, 10)
	landing(-22, -19, 7.2, 10)
	g.Use(White).Fill(Block(V(34, 0.8, 10.5), V(37, 3.6, 13)), Block(V(-39, 0.8, 9.5), V(-36, 2.8, 12)))
}

// roofs adds the gold pavilions on the Red Palace and banners on the parapets.
func (m PotalaPalace) roofs(g *Grid) {
	roof := 36 - 0.8
	pav := func(cx, cz, w, d float64) {
		x0, x1, z0, z1 := cx-w/2, cx+w/2, cz-d/2, cz+d/2
		y := roof
		g.Use(Red).Fill(Block(V(x0, y, z0), V(x1, y+1.6, z1)))
		g.Paint(Wood, Block(V(x0, y+1.0, z0), V(x1, y+1.6, z1)))
		g.Carve(Block(V(cx-0.5, y, z1-0.4), V(cx+0.5, y+1.0, z1+0.4)))
		y += 1.6
		g.Use(Gold).Fill(Intersect(HipRoof{Min: P(x0-1, z0-1), Max: P(x1+1, z1+1), Base: y}, Block(V(x0-1, y, z0-1), V(x1+1, y+0.9, z1+1))))
		g.Use(Red).Fill(Block(V(x0+0.6, y+0.9, z0+0.6), V(x1-0.6, y+1.9, z1-0.6)))
		y += 1.9
		g.Use(Gold).Fill(HipRoof{Min: P(x0-0.2, z0-0.2), Max: P(x1+0.2, z1+0.2), Base: y})
		// Ridge finial: hip rises to min(w, d)/2 + 0.2.
		rise := math.Min(w, d)/2 + 0.2
		g.Use(Gold).Fill(Capsule{A: V(cx, y+rise-0.3, cz), B: V(cx, y+rise+1.3, cz), Radius: 0.5})
	}
	pav(-11.5, -5.5, 4.4, 4.4)
	pav(-3, -6, 5.8, 5.0)
	pav(6, -5.5, 4.4, 4.4)
	pav(-12, 3.2, 3.4, 3.4)
	pav(-3.0, 1.5, 3.6, 3.6)
	pav(6, 3.0, 3.4, 3.4)
}

// shol is the walled village at the foot of the palace.
func (m PotalaPalace) shol(g *Grid) {
	wall := func(a, b Vec3) {
		g.Use(White).Fill(Block(a, b))
		g.Paint(Wood, Block(V(a.X, b.Y-0.6, a.Z), b))
	}
	top := potalaPlain + 3.4
	// Front wall with a gap for the gate tower, and side walls up to the bastions.
	wall(V(-46.7, 0.8, 27.3), V(-7, top, 28.7))
	wall(V(1, 0.8, 27.3), V(44.7, top, 28.7))
	wall(V(-46.7, 0.8, 8), V(-45.3, top, 28.7))
	wall(V(43.3, 0.8, 9), V(44.7, top, 28.7))
	corner := func(cx, cz float64, seed int) *potalaBlock {
		return &potalaBlock{x0: cx - 2.8, x1: cx + 2.8, z0: cz - 2.8, z1: cz + 2.8, y0: 0.8, y1: 8,
			batter: 0.07, wall: 1.2, color: White, rows: []float64{4.4}, winW: 0.9, winH: 1.2, step: 2.4,
			faces: [4]bool{true, false, true, true}, bandA: 5.9, bandB: 7.0, band: Wood, seed: seed}
	}
	gate := &potalaBlock{x0: -7.5, x1: 1.5, z0: 25.5, z1: 30.5, y0: 0.8, y1: 10,
		batter: 0.06, wall: 1.2, color: White, rows: []float64{6.2}, winW: 0.9, winH: 1.3, step: 2.2,
		faces: [4]bool{true, true, false, false}, bandA: 7.6, bandB: 9.0, band: Wood, seed: 20}
	var houses []*potalaBlock
	houses = append(houses, corner(-46, 28, 21), corner(44, 28, 22), gate)
	// Village houses in two loose rows, leaving lanes and the central approach clear.
	k := 0
	for _, row := range []struct{ z0, z1 float64 }{{16, 19.5}, {21, 25.2}} {
		x := -44.0
		for x < 41 {
			w := 4.5 + 2.5*Hash01(k, 51)
			if x < -2 && x+w > -8 {
				x = 1
				continue
			}
			if x+w > 42.5 {
				break
			}
			d := row.z1 - row.z0 - Hash01(k, 52)*1.2
			storeys := 1
			if row.z0 > 20 && Hash01(k, 53) < 0.5 {
				storeys = 2
			}
			h := 1.8 + 2.4*float64(storeys)
			c := White
			if Hash01(k, 54) < 0.2 {
				c = Sand
			}
			hb := &potalaBlock{x0: x, x1: x + w, z0: row.z1 - d, z1: row.z1, y0: 1.0, y1: 1.0 + h,
				batter: 0.05, wall: 1.1, color: c, rows: m.rows(1.9, 2.4, storeys), winW: 0.9, winH: 1.1, step: 2.2,
				faces: [4]bool{true, false, true, true}, bandA: h + 1.0 - 1.4, bandB: h + 1.0 - 0.8, band: Wood, seed: 60 + k}
			if Hash01(k, 55) < 0.3 {
				hb.band = Brick
			}
			houses = append(houses, hb)
			x += w + 1.2 + 1.6*Hash01(k, 56)
			k++
		}
	}
	for _, b := range houses {
		g.Carve(b.outer(0.3, 0))
	}
	for _, b := range houses {
		b.shell(g)
	}
	for _, b := range houses {
		b.finish(g)
	}
	// Gate passage.
	g.Carve(Arch{Sill: V(-3, potalaPlain, 28), HalfWidth: 1.7, Spring: 2.4, Depth: 6, Through: AxisZ})
	m.banner(g, V(-6.4, 10, 29.4))
	m.banner(g, V(0.4, 10, 29.4))
	// A few trees in the village.
	for i, p := range []Vec2{P(-30, 19.5), P(-12, 19.8), P(14, 19.6), P(30, 19.4), P(-38, 26), P(36, 26)} {
		if g.HasAt(p.At(potalaPlain + 1)) {
			continue
		}
		m.poplar(g, p, 7+Hash01(i, 71)*2, i)
	}
}

func (PotalaPalace) poplar(g *Grid, p Vec2, h float64, i int) {
	g.Use(Wood).Fill(Cylinder{Center: p, Radius: 0.5, Y0: potalaPlain, Y1: potalaPlain + 2.4})
	c := Foliage
	if Hash01(i, 73) < 0.4 {
		c = Grass
	}
	g.Use(c).FillVolume(Ellipsoid{Center: p.At(potalaPlain + 1.6 + h/2), Radii: V(1.3, h/2, 1.3)})
}

// plaza lines the square with poplars.
func (m PotalaPalace) plaza(g *Grid) {
	i := 0
	for _, p := range []Vec2{P(-35, 32.5), P(-32, 36.5), P(-26, 37.5), P(31, 32.5), P(28, 36.5), P(22, 37.5), P(-40, 30), P(37, 29.5)} {
		m.poplar(g, p, 6+2*Hash01(i, 81), i+10)
		i++
	}
}

// chorten is the white stupa gate straddling the western road.
func (m PotalaPalace) chorten(g *Grid, c Vec2) {
	y := potalaPlain - 0.6
	g.Carve(BlockAt(c, 7.5, 7.5, potalaPlain-0.2, potalaPlain+20))
	g.Use(White).Fill(BlockAt(c, 7, 7, y, potalaPlain+4.2))
	g.Paint(Wood, BlockAt(c, 7, 7, potalaPlain+3.4, potalaPlain+4.2))
	g.Use(White).Fill(
		BlockAt(c, 5.6, 5.6, potalaPlain+4.2, potalaPlain+5.0),
		BlockAt(c, 4.6, 4.6, potalaPlain+5.0, potalaPlain+5.8),
		BlockAt(c, 3.8, 3.8, potalaPlain+5.8, potalaPlain+6.6),
	)
	g.Use(White).FillVolume(Intersect(
		Ellipsoid{Center: c.At(potalaPlain + 6.6), Radii: V(2.6, 3.0, 2.6)},
		Block(V(c.X-3, potalaPlain+6.6, c.Y-3), V(c.X+3, potalaPlain+10, c.Y+3)),
	))
	g.Use(White).Fill(BlockAt(c, 2, 2, potalaPlain+9.4, potalaPlain+10.4))
	g.Use(Gold).Fill(
		Cone(c, 1.3, potalaPlain+10.4, 4.2),
		Cylinder{Center: c, Radius: 1.0, Y0: potalaPlain + 13.6, Y1: potalaPlain + 14.2},
		Capsule{A: c.At(potalaPlain + 14.0), B: c.At(potalaPlain + 15.6), Radius: 0.45},
	)
	// Passage through along X.
	g.Carve(Arch{Sill: c.At(potalaPlain - 0.3), HalfWidth: 1.6, Spring: 1.6, Depth: 8, Through: AxisX})
	g.Paint(Charcoal, Arch{Sill: c.At(potalaPlain - 0.6), HalfWidth: 2.3, Spring: 1.6, Depth: 8, Through: AxisX})
}
