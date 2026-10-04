package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// Alhambra is the Nasrid fortress-palace on the Sabika ridge above Granada. The ridge
// runs along X with its west tip at +X and north toward +Z, so the hero camera looks
// across the Darro valley as from the Albaicín. Reddish rammed-earth walls with many
// square towers follow the plateau's edge. The Alcazaba holds the west tip with the
// Torre de la Vela and its bell; behind the Plaza de los Aljibes stand the square
// Palace of Charles V with its round courtyard and, on the north wall, the Nasrid
// Palaces: the massive Comares Tower over the Court of the Myrtles, then the Court of
// the Lions with its slender arcades, pavilions, fountain and four channels, and the
// Partal pool before the Torre de las Damas. The upper medina holds the church, the
// convent, ruins and gardens. Across the Cuesta de los Chinos ravine the Generalife's
// terraced gardens hold the Patio de la Acequia, long pools, jets and cypresses.
// Woods cover the south slope; the Darro runs along the northern foot.
type Alhambra struct{}

func init() { Register(Alhambra{}) }

func (Alhambra) Name() string  { return "alhambra" }
func (Alhambra) Title() string { return "Alhambra" }

const (
	alhambraH    = 12.0 // Plateau level.
	alhambraWall = alhambraH + 2.6
	alhambraGH   = 15.0 // Generalife upper terrace.
	alhambraGM   = 12.0 // Generalife lower terrace.
	alhambraX0   = -81.0
	alhambraX1   = 52.0
	alhambraZ0   = -28.0
	alhambraZ1   = 31.0
)

// Terrain kinds.
const (
	alhambraPlateau = iota
	alhambraTerrace
	alhambraSlope
	alhambraRiver
	alhambraBank
)

// alhambraRidge is the walled plateau outline; towers stand on its vertices.
var alhambraRidge = []Vec2{
	P(43, 4), P(39, 8), P(33, 9.5), P(26, 10.5), P(20, 12.5), P(-2, 16), P(-10, 16.2),
	P(-18, 16.5), P(-26, 15), P(-32, 12.5), P(-37, 7), P(-39, 0), P(-37, -7), P(-30, -11),
	P(-22, -13), P(-14, -14), P(-6, -15), P(2, -15.5), P(10, -15.5), P(20, -15),
	P(27, -12), P(33, -9), P(39, -6), P(43, -3),
}

// alhambraTerraces are the Generalife's garden platforms: x0, z0, x1, z1, level.
var alhambraTerraces = [][5]float64{
	{-71, -8, -56, 7, alhambraGH},
	{-56, -11, -45, 9, alhambraGM},
}

func alhambraSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

// alhambraNoise is smooth value noise in [0, 1).
func alhambraNoise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

func alhambraFBM(x, z float64, seed int) float64 {
	return (alhambraNoise(x, z, seed)*4 + alhambraNoise(x*2.1+5, z*2.1+3, seed+1)*2 +
		alhambraNoise(x*4.3+1, z*4.3+7, seed+2)) / 7
}

// alhambraEdge returns the distance from p to a closed outline and the arc length of
// the nearest outline point.
func alhambraEdge(poly []Vec2, p Vec2) (float64, float64) {
	best, arc, s := math.Inf(1), 0.0, 0.0
	for i := range poly {
		a, b := poly[i], poly[(i+1)%len(poly)]
		ab := b.Sub(a)
		l := ab.Length()
		t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/(l*l)))
		if d := p.Sub(a.Add(ab.Scale(t))).Length(); d < best {
			best, arc = d, s+t*l
		}
		s += l
	}
	return best, arc
}

// alhambraRiverZ is the Darro's centerline.
func alhambraRiverZ(x float64) float64 { return 23 + 1.6*math.Sin(x/14+0.5) }

// alhambraMap is a sampled heightfield with bilinear lookup.
type alhambraMap struct {
	x0, z0, step float64
	nx, nz       int
	h            []float64
}

func (m *alhambraMap) at(x, z float64) float64 {
	fx, fz := (x-m.x0)/m.step, (z-m.z0)/m.step
	if fx < 0 || fz < 0 || fx >= float64(m.nx-1) || fz >= float64(m.nz-1) {
		return 0
	}
	i, j := int(fx), int(fz)
	tx, tz := fx-float64(i), fz-float64(j)
	k := j*m.nx + i
	a := m.h[k] + (m.h[k+1]-m.h[k])*tx
	b := m.h[k+m.nx] + (m.h[k+m.nx+1]-m.h[k+m.nx])*tx
	return a + (b-a)*tz
}

func (m *alhambraMap) low(x, z float64) float64 {
	lo := m.at(x, z)
	for k := 0; k < 8; k++ {
		a := float64(k) * math.Pi / 4
		lo = math.Min(lo, m.at(x+0.9*math.Cos(a), z+0.9*math.Sin(a)))
	}
	return lo
}

// rim is the distance inside the site's rounded-rectangle boundary; negative is off
// the plot.
func (Alhambra) rim(x, z float64) float64 {
	const r = 10.0
	dx := (alhambraX1-alhambraX0)/2 - math.Abs(x-(alhambraX0+alhambraX1)/2)
	dz := (alhambraZ1-alhambraZ0)/2 - math.Abs(z-(alhambraZ0+alhambraZ1)/2)
	if dx < r && dz < r {
		return r - math.Hypot(r-dx, r-dz)
	}
	return math.Min(dx, dz)
}

// height is the ground height and kind of a column; ok is false off the plot.
func (a Alhambra) height(x, z float64) (float64, int, bool) {
	e := a.rim(x, z)
	if e < 0 {
		return 0, alhambraSlope, false
	}
	p := P(x, z)
	kind := alhambraSlope
	var h float64
	if InsidePolygon(alhambraRidge, p) {
		h, kind = alhambraH, alhambraPlateau
	} else {
		d, _ := alhambraEdge(alhambraRidge, p)
		n := (alhambraFBM(x/7, z/7, 3) - 0.5) * 4 * alhambraSmooth((d-2)/5)
		h = alhambraH - 3.6 - 0.48*d - 1.8*alhambraSmooth(d/8) + n
	}
	// Cerro del Sol under the Generalife.
	hill := 1 + 12*math.Exp(-(math.Pow((x+62)/13, 2)+math.Pow(z/15, 2)))
	hill += (alhambraFBM(x/6, z/6, 9) - 0.5) * 2.5
	if hill > h {
		h, kind = hill, alhambraSlope
	}
	for _, t := range alhambraTerraces {
		dx := math.Max(0, math.Max(t[0]-x, x-t[2]))
		dz := math.Max(0, math.Max(t[1]-z, z-t[3]))
		d := math.Hypot(dx, dz)
		th := t[4]
		if d > 0 {
			th = t[4] - 2.4 - 0.9*d + (alhambraFBM(x/5, z/5, 11)-0.5)*2*alhambraSmooth((d-1)/3)
		}
		if th >= h {
			h = th
			if d == 0 {
				kind = alhambraTerrace
			} else {
				kind = alhambraSlope
			}
		}
	}
	// The Darro and its north bank.
	rd := z - alhambraRiverZ(x)
	if rd > 2.4 {
		if bank := 0.9 + 0.32*(rd-2.4) + (alhambraFBM(x/6, z/6, 5)-0.5)*1.2; bank > h {
			h, kind = bank, alhambraBank
		}
	}
	if v := 0.5 + 0.7*math.Max(0, math.Abs(rd)-2.3); v < h {
		h = v
		if kind == alhambraPlateau || kind == alhambraTerrace {
			kind = alhambraSlope
		}
	}
	if math.Abs(rd) < 2.3 {
		kind = alhambraRiver
	}
	if c := 0.9 + 1.2*e; c < h {
		h = c
	}
	if h < 0.9 && kind != alhambraRiver {
		h = 0.9
	}
	return h, kind, true
}

func (a Alhambra) Build(g *Grid) Order {
	m := &alhambraMap{x0: alhambraX0 - 1, z0: alhambraZ0 - 1, step: 0.4}
	m.nx, m.nz = int((alhambraX1-alhambraX0+2)/m.step)+2, int((alhambraZ1-alhambraZ0+2)/m.step)+2
	m.h = make([]float64, m.nx*m.nz)
	for j := 0; j < m.nz; j++ {
		for i := 0; i < m.nx; i++ {
			h, _, _ := a.height(m.x0+float64(i)*m.step, m.z0+float64(j)*m.step)
			m.h[j*m.nx+i] = h
		}
	}
	a.terrain(g, m)
	a.walls(g)
	a.alcazaba(g)
	a.charlesV(g)
	a.myrtles(g)
	a.lions(g)
	a.partal(g)
	a.medina(g)
	a.generalife(g)
	a.darro(g, m)
	a.woods(g, m)
	return Along(P(-1, 0.15))
}

// terrain fills the heightfield as a hollow skin, column by column.
func (a Alhambra) terrain(g *Grid, m *alhambraMap) {
	half := g.Span() / 2
	cx0, cx1 := int(math.Ceil(alhambraX0/half)), int(math.Floor(alhambraX1/half))
	cz0, cz1 := int(math.Ceil(alhambraZ0/half)), int(math.Floor(alhambraZ1/half))
	for cz := cz0; cz <= cz1; cz++ {
		for cx := cx0; cx <= cx1; cx++ {
			if (cx^cz)&1 != 0 {
				continue
			}
			x, z := float64(cx)*half, float64(cz)*half
			h, kind, ok := a.height(x, z)
			if !ok {
				continue
			}
			low := math.Min(h, m.low(x, z))
			slope := (m.at(x, z) - m.low(x, z)) / 0.9
			y0 := math.Max(0, low-0.8)
			cy0 := int(math.Ceil(y0 / half))
			if (cy0^cx)&1 != 0 {
				cy0++
			}
			for cy := cy0; float64(cy)*half <= h; cy += 2 {
				y := float64(cy) * half
				top := y >= h-0.6
				var c Color
				switch kind {
				case alhambraPlateau:
					c = Sand
					if !top {
						c = Brick
					}
				case alhambraTerrace:
					c = Grass
					if !top {
						c = Terracotta
						if Hash01(cx, cy, cz, 7) < 0.25 {
							c = Brick
						}
					}
				case alhambraRiver:
					c = Water
					if !top {
						c = DarkStone
					}
				case alhambraBank:
					c = Grass
					if alhambraFBM(x/5, z/5, 21) > 0.6 {
						c = Foliage
					}
				default:
					switch {
					case slope > 1.3 && a.rim(x, z) > 3:
						c = Brick
						if alhambraNoise(x/2+y/3, z/2-y/3, 31) > 0.6 {
							c = DarkStone
						}
					case h < 1.6:
						c = Grass
					default:
						c = Foliage
						if alhambraFBM(x/6, z/6, 33) > 0.58 {
							c = Grass
						}
					}
				}
				g.Use(c)
				g.Add(Cell{X: int32(cx), Y: int32(cy), Z: int32(cz)})
			}
		}
	}
}

// rammed paints lighter Terracotta patches over Brick rammed-earth courses.
func (Alhambra) rammed(g *Grid, b Box) {
	var patch []Cell
	for _, c := range g.Cells() {
		p := g.World(c)
		if col, _ := g.ColorOf(c); col != Brick || !b.Contains(p) {
			continue
		}
		if alhambraNoise(p.X/2.6+p.Z/2.6, math.Floor(p.Y/0.8)*1.7, 41) > 0.66 {
			patch = append(patch, c)
		}
	}
	g.With(Terracotta, func() {
		for _, c := range patch {
			g.Remove(c)
			g.Add(c)
		}
	})
}

// darken paints the cells lining a void box Charcoal so openings read dark: every
// occupied cell with an empty face neighbor inside the box.
func (a Alhambra) darken(g *Grid, x0, z0, x1, z1, y0, y1 float64) {
	a.darkenIn(g, Block(V(x0, y0, z0), V(x1, y1, z1)))
}

// darkenIn paints Charcoal the occupied cells facing void. Cells with more than one
// empty neighbor outside it are outer faces of thin walls and keep their color.
func (Alhambra) darkenIn(g *Grid, void Solid) {
	near := void.Bounds().Grow(g.Span())
	var lining []Cell
	for _, c := range g.Cells() {
		if !near.Contains(g.World(c)) {
			continue
		}
		in, out := false, 0
		for _, n := range Neighbors {
			q := c.Add(n)
			if g.Has(q) {
				continue
			}
			if w := g.World(q); near.Contains(w) && void.Contains(w) {
				in = true
			} else {
				out++
			}
		}
		if in && out <= 1 {
			lining = append(lining, c)
		}
	}
	g.With(Charcoal, func() {
		for _, c := range lining {
			g.Remove(c)
			g.Add(c)
		}
	})
}

// shell fills a hollow box with walls t thick and a roof slab, darkened inside.
func (a Alhambra) shell(g *Grid, x0, z0, x1, z1, y0, y1, t float64) {
	g.Fill(Block(V(x0, y0, z0), V(x1, y1, z1)))
	g.Carve(Block(V(x0+t, y0+0.5, z0+t), V(x1-t, y1-0.7, z1-t)))
	if t >= 0.5 {
		a.darken(g, x0+t, z0+t, x1-t, z1-t, y0+0.5, y1-0.7)
	}
}

// lowHip is a shallow hipped tile roof over a rectangle.
func (Alhambra) lowHip(x0, z0, x1, z1, base, pitch float64) Solid {
	rise := math.Min(x1-x0, z1-z0) / 2 * pitch
	return Func{Box{V(x0, base, z0), V(x1, base+rise+0.4, z1)}, func(p Vec3) bool {
		d := math.Min(math.Min(p.X-x0, x1-p.X), math.Min(p.Z-z0, z1-p.Z))
		return p.Y <= base+0.4+d*pitch
	}}
}

// crenels adds a merlon parapet around a flat rectangular top at height y.
func (Alhambra) crenels(g *Grid, x0, z0, x1, z1, y, rim float64) {
	g.Fill(Func{Box{V(x0, y, z0), V(x1, y+1.6, z1)}, func(p Vec3) bool {
		onX := p.X <= x0+rim || p.X >= x1-rim
		onZ := p.Z <= z0+rim || p.Z >= z1-rim
		if !onX && !onZ {
			return false
		}
		if p.Y <= y+0.8 {
			return true
		}
		return (onZ && math.Mod(p.X-x0+0.4, 1.6) < 0.8) || (onX && math.Mod(p.Z-z0+0.4, 1.6) < 0.8)
	}})
}

// arches carves a row of n round-headed openings through a wall. The row is centered
// on c (plan center at sill height y) and runs across the Through axis.
func (Alhambra) arches(g *Grid, c Vec3, n int, pitch, halfWidth, spring, depth float64, through Axis) {
	for i := 0; i < n; i++ {
		off := (float64(i) - float64(n-1)/2) * pitch
		at := c.Add(V(off, 0, 0))
		if through == AxisX {
			at = c.Add(V(0, 0, off))
		}
		g.Carve(Arch{Sill: at, HalfWidth: halfWidth, Spring: spring, Depth: depth, Through: through})
	}
}

// lattice fills arched openings with a Charcoal or Glass screen at the wall's center.
func (Alhambra) lattice(g *Grid, c Vec3, n int, pitch, halfWidth, spring float64, through Axis, color Color) {
	g.With(color, func() {
		for i := 0; i < n; i++ {
			off := (float64(i) - float64(n-1)/2) * pitch
			at := c.Add(V(off, 0, 0))
			if through == AxisX {
				at = c.Add(V(0, 0, off))
			}
			arch := Arch{Sill: at, HalfWidth: halfWidth, Spring: spring, Depth: 0.5, Through: through}
			g.Fill(Func{arch.Bounds(), func(p Vec3) bool {
				if !arch.Contains(p) {
					return false
				}
				k := int(math.Floor(p.Y/0.8)) + int(math.Floor((p.X+p.Z)/0.8))
				return color == Glass || k%2 == 0
			}})
		}
	})
}

// walls raises the curtain wall and its square towers along the plateau's edge.
func (a Alhambra) walls(g *Grid) {
	b := EmptyBox()
	for _, v := range alhambraRidge {
		b = b.Union(Box{v.At(alhambraH - 5), v.At(alhambraWall + 2.5)})
	}
	b = b.Grow(1.5)
	wall := Func{b, func(p Vec3) bool {
		d, s := alhambraEdge(alhambraRidge, p.Plan())
		if InsidePolygon(alhambraRidge, p.Plan()) {
			d = -d
		}
		if d < -1.3 || d > 0.2 || p.Y < alhambraH-4.8 {
			return false
		}
		top := alhambraWall
		if p.X > 27.5 {
			top += 1
		}
		if p.Y <= top {
			return true
		}
		return d > -0.75 && p.Y <= top+0.8 && math.Mod(s, 1.6) < 0.8
	}}
	g.Use(Brick).Fill(wall)
	a.rammed(g, b)
	// Towers on the vertices; the big named towers are built separately.
	skip := map[int]bool{0: true, 7: true, 23: true}
	for i, v := range alhambraRidge {
		if skip[i] {
			continue
		}
		r := 1.6
		top := alhambraH + 4.4 + 1.4*Hash01(i, 3)
		if v.X > 27 {
			top += 1.5
		}
		if i == 9 || i == 13 || i == 19 { // Infantas, Siete Suelos, Justicia.
			r, top = 2.3, top+1.2
		}
		x0, z0, x1, z1 := v.X-r, v.Y-r, v.X+r, v.Y+r
		g.Use(Brick)
		a.shell(g, x0, z0, x1, z1, alhambraH-6.5, top, 0.9)
		// Arched windows high on every face.
		for _, ax := range []Axis{AxisX, AxisZ} {
			g.Carve(Arch{Sill: v.At(top - 2.6), HalfWidth: 0.45, Spring: 0.8, Depth: 2*r + 1, Through: ax})
		}
		a.rammed(g, Box{V(x0, alhambraH-6.5, z0), V(x1, top+2, z1)})
		g.Use(Brick)
		if Hash01(i, 5) < 0.4 {
			g.Use(Terracotta).Fill(Block(V(x0-0.3, top, z0-0.3), V(x1+0.3, top+0.5, z1+0.3)))
			g.Fill(a.lowHip(x0-0.3, z0-0.3, x1+0.3, z1+0.3, top+0.5, 0.55))
		} else {
			a.crenels(g, x0, z0, x1, z1, top, 0.7)
		}
	}
}

// alcazaba builds the fortress on the west tip: the Torre de la Vela with its bell,
// the keep, the cross wall and the Plaza de Armas foundations.
func (a Alhambra) alcazaba(g *Grid) {
	h := alhambraH
	// Cross wall with its gate, between the Alcazaba and the Plaza de los Aljibes.
	g.Use(Brick).Fill(Block(V(27.2, h-0.5, -11.8), V(28.6, h+4.2, 10.8)))
	g.Fill(Func{Box{V(27.2, h+4.2, -11.8), V(28.6, h+5.0, 10.8)}, func(p Vec3) bool {
		return math.Mod(p.Z+20, 1.6) < 0.8
	}})
	g.Carve(Arch{Sill: V(27.9, h, 0), HalfWidth: 1.1, Spring: 1.6, Depth: 2, Through: AxisX})
	a.rammed(g, Box{V(27, h-1, -12), V(29, h+5, 11)})
	// Plaza de Armas: grid of barracks foundations.
	g.Use(Brick)
	for x := 29.5; x <= 37.5; x += 2.6 {
		g.Fill(Block(V(x, h, -6), V(x+0.7, h+0.9, 6)))
	}
	for z := -6.0; z <= 6; z += 3 {
		g.Fill(Block(V(29.5, h, z), V(38.2, h+0.9, z+0.7)))
	}
	g.Paint(DarkStone, Func{Box{V(29, h, -7), V(39, h+1, 7)}, func(p Vec3) bool {
		return Hash01(int(p.X*3), int(p.Z*3), 17) < 0.3
	}})
	// Torre de la Vela at the tip.
	vx0, vz0, vx1, vz1, vt := 38.0, -2.6, 44.2, 3.6, h+12
	g.Use(Brick)
	a.shell(g, vx0, vz0, vx1, vz1, h-6.5, vt, 1.1)
	for _, y := range []float64{h + 3.2, h + 7.6} {
		a.arches(g, V(44.2, y, 0.5), 2, 2.4, 0.5, 1.2, 2.4, AxisX)
		a.arches(g, V(41.1, y, 3.6), 2, 2.4, 0.5, 1.2, 2.4, AxisZ)
		a.arches(g, V(41.1, y, -2.6), 2, 2.4, 0.5, 1.2, 2.4, AxisZ)
	}
	a.rammed(g, Box{V(vx0, h-7, vz0), V(vx1, vt, vz1)})
	g.Use(Brick)
	a.crenels(g, vx0, vz0, vx1, vz1, vt, 0.7)
	// Bell gable: a small arched belfry with the gold bell and a flag.
	bx, bz := 40.0, 0.5
	g.Use(Sand).Fill(Block(V(bx-0.6, vt, bz-1.6), V(bx+0.6, vt+4.2, bz+1.6)))
	g.Carve(Arch{Sill: V(bx, vt+1.2, bz), HalfWidth: 0.9, Spring: 1.0, Depth: 2, Through: AxisX})
	g.Use(Terracotta).Fill(Gable{Min: P(bx-0.9, bz-1.9), Max: P(bx+0.9, bz+1.9), Base: vt + 4.2, Ridge: AxisZ})
	g.Use(Gold).FillVolume(Ellipsoid{V(bx, vt+2.2, bz), V(0.55, 0.75, 0.75)})
	g.Fill(Block(V(bx-0.4, vt+2.8, bz-0.4), V(bx+0.4, vt+3.4, bz+0.4)))
	g.Use(Charcoal).Fill(Cylinder{Center: P(42.6, -1.2), Radius: 0.35, Y0: vt, Y1: vt + 6})
	g.Use(Red).Fill(Block(V(42.4, vt+4.1, -4.4), V(42.8, vt+5.9, -1.4)))
	g.Use(Gold).Fill(Block(V(42.4, vt+4.7, -4.4), V(42.8, vt+5.3, -1.4)))
	// Torre del Homenaje and Torre Quebrada.
	for _, t := range [][5]float64{{29.2, 6.2, 33.8, 11.2, h + 10}, {29.5, -10.6, 33.5, -6.4, h + 7.6}} {
		g.Use(Brick)
		a.shell(g, t[0], t[1], t[2], t[3], h-6.5, t[4], 1)
		cx, cz := (t[0]+t[2])/2, (t[1]+t[3])/2
		for _, ax := range []Axis{AxisX, AxisZ} {
			g.Carve(Arch{Sill: V(cx, t[4]-3, cz), HalfWidth: 0.45, Spring: 0.9, Depth: 6, Through: ax})
		}
		a.rammed(g, Box{V(t[0], h-7, t[1]), V(t[2], t[4], t[3])})
		g.Use(Brick)
		a.crenels(g, t[0], t[1], t[2], t[3], t[4], 0.7)
	}
	// Puerta del Vino in the Plaza de los Aljibes.
	g.Use(Sand)
	a.shell(g, 21.2, -4.2, 24.2, -0.6, h, h+6.2, 0.8)
	g.Carve(Arch{Sill: V(22.7, h, -2.4), HalfWidth: 0.9, Spring: 2, Depth: 4, Through: AxisX})
	g.Use(Terracotta).Fill(a.lowHip(20.9, -4.5, 24.5, -0.3, h+6.2, 0.5))
	a.lattice(g, V(24.2, h+3.6, -2.4), 2, 1.2, 0.35, 0.8, AxisX, Charcoal)
}

// charlesV builds the square Renaissance palace with its round two-storey courtyard.
func (a Alhambra) charlesV(g *Grid) {
	h := alhambraH
	cx, cz, half := 12.5, -6.8, 7.6
	x0, z0, x1, z1 := cx-half, cz-half, cx+half, cz+half
	c := P(cx, cz)
	g.Use(Sand).Fill(Block(V(x0, h, z0), V(x1, h+6.6, z1)))
	g.Use(Stone).Fill(Block(V(x0-0.2, h, z0-0.2), V(x1+0.2, h+3, z1+0.2)))
	// Rusticated lower storey.
	g.Paint(DarkStone, Func{Box{V(x0-0.3, h, z0-0.3), V(x1+0.3, h+3, z1+0.3)}, func(p Vec3) bool {
		return int(math.Floor(p.Y/0.8)+math.Floor((p.X+p.Z)/1.6))%2 == 0 && Hash01(int(p.X*2), int(p.Y*2), int(p.Z*2)) < 0.5
	}})
	// Hollow rooms between the outer walls and the ring wall.
	inner := Subtract(Block(V(x0+1.1, h+0.5, z0+1.1), V(x1-1.1, h+6, z1-1.1)), Cylinder{Center: c, Radius: 8.8, Y0: h - 1, Y1: h + 8})
	g.Carve(inner)
	a.darkenIn(g, inner)
	// Courtyard and galleries.
	g.Carve(Cylinder{Center: c, Radius: 7.7, Y0: h + 0.3, Y1: h + 8})
	g.Use(Sand).Fill(Cylinder{Center: c, Radius: 7.7, Inner: 5.5, Y0: h + 3, Y1: h + 3.6})
	g.Fill(Cylinder{Center: c, Radius: 7.7, Inner: 5.5, Y0: h + 6.1, Y1: h + 6.6})
	a.darkenIn(g, Cylinder{Center: c, Radius: 7.7, Inner: 6.4, Y0: h + 0.3, Y1: h + 6.1})
	g.Use(Stone).Fill(Cylinder{Center: c, Radius: 5.5, Y0: h - 0.3, Y1: h + 0.2})
	for i := 0; i < 20; i++ {
		q := Polar(c, 5.95, float64(i)*math.Pi/10)
		g.Use(White).Fill(Cylinder{Center: q, Radius: 0.42, Y0: h, Y1: h + 3})
		g.Fill(Cylinder{Center: q, Radius: 0.42, Y0: h + 3.6, Y1: h + 6.1})
	}
	// Outer windows in two rows, and doorways in the middle of each face.
	for k := -3; k <= 3; k++ {
		off := float64(k) * 2
		for _, y := range []float64{h + 1, h + 4} {
			g.Carve(Block(V(cx+off-0.4, y, z0-0.5), V(cx+off+0.4, y+1.5, z0+1.2)),
				Block(V(cx+off-0.4, y, z1-1.2), V(cx+off+0.4, y+1.5, z1+0.5)),
				Block(V(x0-0.5, y, cz+off-0.4), V(x0+1.2, y+1.5, cz+off+0.4)),
				Block(V(x1-1.2, y, cz+off-0.4), V(x1+0.5, y+1.5, cz+off+0.4)))
		}
	}
	g.Carve(Arch{Sill: V(x1, h, cz), HalfWidth: 0.9, Spring: 1.8, Depth: 2.6, Through: AxisX})
	g.Carve(Arch{Sill: V(cx, h, z0), HalfWidth: 0.9, Spring: 1.8, Depth: 2.6, Through: AxisZ})
	// Cornice and the low tiled roof around the courtyard.
	g.Use(White).Fill(Subtract(Block(V(x0-0.3, h+6.6, z0-0.3), V(x1+0.3, h+7, z1+0.3)), Cylinder{Center: c, Radius: 5.5, Y0: h, Y1: h + 9}))
	g.Use(Terracotta).Fill(Func{Box{V(x0, h+7, z0), V(x1, h+9.5, z1)}, func(p Vec3) bool {
		r := p.Plan().Sub(c).Length()
		dEdge := math.Min(math.Min(p.X-x0, x1-p.X), math.Min(p.Z-z0, z1-p.Z))
		d := math.Min(dEdge, r-5.5)
		return d > 0 && p.Y <= h+7.2+0.45*d
	}})
}

// myrtles builds the Comares Tower, the Court of the Myrtles with its long pool and
// hedges, the end porticos and the Mexuar.
func (a Alhambra) myrtles(g *Grid) {
	h := alhambraH
	cx0, cz0, cx1, cz1 := 7.0, 2.9, 15.0, 11.2 // Court.
	mx := (cx0 + cx1) / 2
	// Ranges around the court.
	g.Use(Sand)
	a.shell(g, 4.4, 0.6, cx0, cz1, h, h+4.6, 0.8)
	a.shell(g, cx1, 0.6, 17.6, cz1, h, h+4.6, 0.8)
	a.shell(g, cx0, 0.6, cx1, 2.2, h, h+4.2, 0.6)
	g.Use(Terracotta).Fill(a.lowHip(4.1, 0.3, cx0+0.3, cz1+0.3, h+4.6, 0.55), a.lowHip(cx1-0.3, 0.3, 17.9, cz1+0.3, h+4.6, 0.55))
	g.Fill(Gable{Min: P(cx0, 0.3), Max: P(cx1, 2.5), Base: h + 4.2, Ridge: AxisX})
	// Windows with lattices looking into the court.
	for _, x := range []float64{cx0, cx1} {
		a.arches(g, V(x, h+2.4, (cz0+cz1)/2), 4, 2.1, 0.35, 0.6, 1.8, AxisX)
		a.lattice(g, V(x, h+2.4, (cz0+cz1)/2), 4, 2.1, 0.35, 0.6, AxisX, Charcoal)
	}
	// South portico.
	g.Use(White).Fill(Block(V(cx0, h, 2.2), V(cx1, h+3.4, 2.9)))
	a.arches(g, V(mx, h, 2.55), 5, 1.55, 0.5, 1.5, 1.2, AxisZ)
	// The court: myrtle hedges flank the long pool.
	g.Use(White).Fill(Block(V(cx0, h-0.4, cz0), V(cx1, h, cz1)))
	g.Use(Water).Fill(Block(V(mx-1.3, h-0.4, cz0+0.7), V(mx+1.3, h, cz1-0.4)))
	g.Use(Foliage).Fill(Block(V(cx0+0.8, h, cz0+0.9), V(cx0+1.9, h+1, cz1-0.6)), Block(V(cx1-1.9, h, cz0+0.9), V(cx1-0.8, h+1, cz1-0.6)))
	// Comares Tower with the north portico and Sala de la Barca at its foot.
	tx0, tz0, tx1, tz1, tt := 6.4, 12.2, 15.6, 20.4, h+16
	g.Use(Brick)
	a.shell(g, tx0, tz0, tx1, tz1, h-8.5, tt, 1.2)
	g.Use(White).Fill(Block(V(cx0, h, cz1-0.2), V(cx1, h+3.8, 12.4)))
	g.Carve(Block(V(cx0+0.4, h+0.3, cz1+0.6), V(cx1-0.4, h+3.2, 13.4)))
	a.arches(g, V(mx, h, cz1+0.3), 7, 1.1, 0.4, 1.6, 1.4, AxisZ)
	a.darken(g, cx0+0.4, cz1+0.6, cx1-0.4, 13.4, h+0.3, h+3.2)
	// Tall arched windows of the Hall of the Ambassadors and small high windows.
	tm := (tz0 + tz1) / 2
	y := h + 6.8
	a.arches(g, V(mx, y, tz1), 3, 2.6, 0.7, 1.9, 2.6, AxisZ)
	a.arches(g, V(tx1, y, tm), 2, 3.2, 0.7, 1.9, 2.6, AxisX)
	a.arches(g, V(tx0, y, tm), 2, 3.2, 0.7, 1.9, 2.6, AxisX)
	a.arches(g, V(mx, h+12.4, tz1), 4, 1.9, 0.4, 0.7, 2.6, AxisZ)
	a.arches(g, V(tx1, h+12.4, tm), 3, 2.2, 0.4, 0.7, 2.6, AxisX)
	a.arches(g, V(tx0, h+12.4, tm), 3, 2.2, 0.4, 0.7, 2.6, AxisX)
	a.rammed(g, Box{V(tx0, h-9, tz0), V(tx1, tt, tz1)})
	g.Use(Brick)
	a.crenels(g, tx0, tz0, tx1, tz1, tt, 0.8)
	g.Use(Terracotta).Fill(a.lowHip(tx0+1.2, tz0+1.2, tx1-1.2, tz1-1.2, tt-0.4, 0.4))
	// Mexuar and the Cuarto Dorado.
	g.Use(Sand)
	a.shell(g, 17.6, 1.6, 21.6, 11.4, h, h+4.4, 0.8)
	g.Use(Terracotta).Fill(a.lowHip(17.3, 1.3, 21.9, 11.7, h+4.4, 0.55))
	a.arches(g, V(21.6, h+1.8, 6.5), 3, 2.2, 0.45, 0.8, 1.8, AxisX)
	a.lattice(g, V(21.6, h+1.8, 6.5), 3, 2.2, 0.45, 0.8, AxisX, Charcoal)
}

// lions builds the Court of the Lions: the arcade of slender columns, the end
// pavilions, the fountain with its lions, the four channels and the halls around.
func (a Alhambra) lions(g *Grid) {
	h := alhambraH
	px0, pz0, px1, pz1 := -11.6, 2.6, 2.0, 11.4 // Court.
	ex0, ex1 := -14.0, 4.4                      // Outer extent along X.
	cx, cz := (px0+px1)/2, (pz0+pz1)/2
	g.Use(Sand)
	// Halls: Two Sisters (north), Abencerrajes (south), Kings (east), west range.
	a.shell(g, ex0, pz1, ex1, 15.2, h, h+5, 0.8)
	a.shell(g, ex0, 0, ex1, pz0, h, h+4.6, 0.7)
	a.shell(g, ex0, pz0, px0, pz1, h, h+4.6, 0.7)
	a.shell(g, px1, pz0, ex1, pz1, h, h+4.6, 0.6)
	// Lanterns over the domed halls, with lattice windows.
	a.shell(g, cx-2.4, pz1+0.4, cx+2.4, 14.8, h+4, h+7.8, 0.7)
	a.shell(g, cx-2.2, 0.2, cx+2.2, pz0-0.2, h+4, h+6.8, 0.6)
	for _, z := range []float64{pz1 + 0.4, 14.8} {
		a.arches(g, V(cx, h+5.4, z), 3, 1.4, 0.35, 0.8, 1.2, AxisZ)
		a.lattice(g, V(cx, h+5.4, z), 3, 1.4, 0.35, 0.8, AxisZ, Charcoal)
	}
	a.arches(g, V(cx, h+5.0, 0.2), 3, 1.4, 0.35, 0.6, 1.2, AxisZ)
	a.lattice(g, V(cx, h+5.0, 0.2), 3, 1.4, 0.35, 0.6, AxisZ, Charcoal)
	g.Use(Terracotta)
	g.Fill(a.lowHip(ex0-0.3, pz1-0.3, ex1+0.3, 15.5, h+5, 0.5), a.lowHip(ex0-0.3, -0.3, ex1+0.3, pz0+0.3, h+4.6, 0.55),
		a.lowHip(ex0-0.3, pz0, px0+0.3, pz1, h+4.6, 0.55))
	g.Fill(a.lowHip(cx-2.7, pz1+0.1, cx+2.7, 15.1, h+7.8, 0.6), a.lowHip(cx-2.5, -0.1, cx+2.5, pz0+0.1, h+6.8, 0.6))
	// Mirador de Lindaraja projecting from the Two Sisters.
	g.Use(Sand)
	a.shell(g, cx-1.6, 15.2, cx+1.6, 17.2, h-4.5, h+4, 0.6)
	a.arches(g, V(cx, h+1.6, 17.2), 2, 1.4, 0.4, 0.9, 1.2, AxisZ)
	a.lattice(g, V(cx, h+1.6, 17.2), 2, 1.4, 0.4, 0.9, AxisZ, Glass)
	g.Use(Terracotta).Fill(a.lowHip(cx-1.9, 14.9, cx+1.9, 17.5, h+4, 0.6))
	// Court floor and the four channels.
	g.Use(White).Fill(Block(V(px0, h-0.4, pz0), V(px1, h, pz1)))
	g.Use(Water).Fill(Block(V(px0+1.1, h-0.4, cz-0.3), V(px1-1.1, h, cz+0.3)), Block(V(cx-0.3, h-0.4, pz0+1.1), V(cx+0.3, h, pz1-1.1)))
	// Gallery: slender columns on the inset rectangle under an arcaded band.
	gx0, gz0, gx1, gz1 := px0+1.1, pz0+1.1, px1-1.1, pz1-1.1
	var cols []Vec2
	nx, nz := 9, 5
	for i := 0; i <= nx; i++ {
		x := gx0 + (gx1-gx0)*float64(i)/float64(nx)
		cols = append(cols, P(x, gz0), P(x, gz1))
	}
	for j := 1; j < nz; j++ {
		z := gz0 + (gz1-gz0)*float64(j)/float64(nz)
		cols = append(cols, P(gx0, z), P(gx1, z))
	}
	g.Use(White)
	for _, q := range cols {
		g.Fill(Cylinder{Center: q, Radius: 0.3, Y0: h, Y1: h + 2.3})
	}
	inner := Block(V(gx0+0.35, h, gz0+0.35), V(gx1-0.35, h+4, gz1-0.35))
	g.Use(Sand).Fill(Subtract(Block(V(gx0-0.3, h+2.3, gz0-0.3), V(gx1+0.3, h+3.1, gz1+0.3)), inner))
	// Arches between columns.
	for i := 0; i < nx; i++ {
		x := gx0 + (gx1-gx0)*(float64(i)+0.5)/float64(nx)
		for _, z := range []float64{gz0, gz1} {
			g.Carve(Arch{Sill: V(x, h+1.9, z), HalfWidth: 0.45, Spring: 0.3, Depth: 0.8, Through: AxisZ})
		}
	}
	for j := 0; j < nz; j++ {
		z := gz0 + (gz1-gz0)*(float64(j)+0.5)/float64(nz)
		for _, x := range []float64{gx0, gx1} {
			g.Carve(Arch{Sill: V(x, h+1.9, z), HalfWidth: 0.45, Spring: 0.3, Depth: 0.8, Through: AxisX})
		}
	}
	// Pent roof over the gallery, sloping into the court.
	g.Use(Terracotta).Fill(Func{Box{V(px0, h+3.1, pz0), V(px1, h+4.4, pz1)}, func(p Vec3) bool {
		d := math.Min(math.Min(p.X-px0, px1-p.X), math.Min(p.Z-pz0, pz1-p.Z))
		in := math.Min(math.Min(p.X-gx0, gx1-p.X), math.Min(p.Z-gz0, gz1-p.Z))
		return in < 0.3 && p.Y <= h+3.1+0.7*(1.4-d)
	}})
	a.darkenIn(g, Subtract(Block(V(px0, h+0.3, pz0), V(px1, h+2.3, pz1)), Block(V(gx0+0.4, h-1, gz0+0.4), V(gx1-0.4, h+3, gz1-0.4))))
	// End pavilions projecting into the court.
	for _, x := range []float64{gx0 - 0.2, gx1 + 0.2} {
		dir := 1.0
		if x > cx {
			dir = -1
		}
		fx := x + dir*1.8
		g.Use(White)
		for _, q := range []Vec2{P(fx, cz-1.2), P(fx, cz+1.2), P(fx, cz), P(x, cz-1.2), P(x, cz+1.2)} {
			g.Fill(Cylinder{Center: q, Radius: 0.3, Y0: h, Y1: h + 2.6})
		}
		lo, hi := math.Min(x, fx)-0.4, math.Max(x, fx)+0.4
		g.Use(Sand).Fill(Block(V(lo, h+2.6, cz-1.6), V(hi, h+3.6, cz+1.6)))
		g.Carve(Arch{Sill: V(fx, h+2.2, cz-0.6), HalfWidth: 0.4, Spring: 0.3, Depth: 0.9, Through: AxisX},
			Arch{Sill: V(fx, h+2.2, cz+0.6), HalfWidth: 0.4, Spring: 0.3, Depth: 0.9, Through: AxisX})
		g.Use(Terracotta).Fill(a.lowHip(lo-0.2, cz-1.8, hi+0.2, cz+1.8, h+3.6, 1.0))
		g.Use(Gold).Fill(Block(V((lo+hi)/2-0.3, h+5.4, cz-0.3), V((lo+hi)/2+0.3, h+6.0, cz+0.3)))
	}
	// Fountain: twelve lions carry the basin.
	c := P(cx, cz)
	g.Use(Stone)
	for i := 0; i < 12; i++ {
		q := Polar(c, 1.5, float64(i)*math.Pi/6)
		g.FillVolume(Ellipsoid{q.At(h + 0.5), V(0.45, 0.5, 0.45)})
	}
	g.Use(White).Fill(Cylinder{Center: c, Radius: 2.0, Y0: h - 0.4, Y1: h})
	g.Fill(Cylinder{Center: c, Radius: 1.9, Inner: 1.3, Y0: h + 0.9, Y1: h + 1.4})
	g.Use(Water).Fill(Cylinder{Center: c, Radius: 1.3, Y0: h + 0.9, Y1: h + 1.2})
	g.Use(White).Fill(Cylinder{Center: c, Radius: 0.4, Y0: h + 1.2, Y1: h + 2.2}, Cylinder{Center: c, Radius: 0.8, Y0: h + 2.2, Y1: h + 2.5})
	g.Use(Glass).Fill(Cylinder{Center: c, Radius: 0.3, Y0: h + 2.5, Y1: h + 3.1})
}

// partal builds the Torre de las Damas with its portico over the long pool and the
// terraced gardens around it.
func (a Alhambra) partal(g *Grid) {
	h := alhambraH
	g.Use(Grass).Fill(Func{Box{V(-32, h-0.4, -1), V(-14.4, h, 16)}, func(p Vec3) bool {
		return InsidePolygon(alhambraRidge, p.Plan()) && math.Abs(p.Z-3.6) > 0.6
	}})
	// Pool with a white curb.
	g.Use(White).Fill(Block(V(-25.2, h-0.4, 4.8), V(-15.6, h+0.4, 9.6)))
	g.Use(Water).Fill(Block(V(-24.6, h-0.4, 5.4), V(-16.2, h+0.2, 9.0)))
	// The tower straddles the wall.
	x0, z0, x1, z1 := -23.4, 11.6, -17.6, 17.6
	mx := (x0 + x1) / 2
	g.Use(Sand)
	a.shell(g, x0, z0, x1, z1, h-6, h+6.4, 0.9)
	g.Use(White).Fill(Block(V(x0, h, 10.4), V(x1, h+3.4, 11.6)))
	g.Carve(Block(V(x0+0.4, h+0.3, 11.0), V(x1-0.4, h+3, 12.4)))
	a.arches(g, V(mx, h, 10.8), 5, 1.1, 0.4, 1.5, 1.4, AxisZ)
	a.darken(g, x0+0.4, 11.0, x1-0.4, 12.4, h+0.3, h+3)
	g.Use(Terracotta).Fill(Block(V(x0-0.1, h+3.4, 10.2), V(x1+0.1, h+3.8, 11.8)))
	a.arches(g, V(mx, h+4.2, z1), 3, 1.5, 0.4, 0.8, 2, AxisZ)
	a.lattice(g, V(mx, h+4.2, z1-0.4), 3, 1.5, 0.4, 0.8, AxisZ, Charcoal)
	a.arches(g, V(x0, h+4.2, (z0+z1)/2), 2, 2, 0.4, 0.8, 2, AxisX)
	g.Use(Terracotta).Fill(a.lowHip(x0-0.3, z0-0.3, x1+0.3, z1+0.3, h+6.4, 0.55))
	// Mirador room on the roof.
	g.Use(Sand)
	a.shell(g, mx-1.1, 13.4, mx+1.1, 15.8, h+6.4, h+8.6, 0.5)
	a.arches(g, V(mx, h+7, 15.8), 1, 1, 0.4, 0.6, 1, AxisZ)
	g.Use(Terracotta).Fill(a.lowHip(mx-1.4, 13.1, mx+1.4, 16.1, h+8.6, 0.8))
	// Garden hedges, a cypress row and shade trees.
	g.Use(Foliage)
	for _, z := range []float64{2.2, 12.4} {
		g.Fill(Block(V(-31, h, z), V(-26.5, h+0.8, z+0.6)))
	}
	for i := 0; i < 5; i++ {
		g.FillVolume(Ellipsoid{V(-30.6+float64(i)*1.2, h+2.2, 7), V(0.6, 2.2, 0.6)})
	}
	for i, q := range []Vec2{P(-15.6, 3.2), P(-26, 10.6), P(-25.6, 3.4), P(-15, 10.2)} {
		r := 1.2 + 0.3*Hash01(i, 77)
		g.Use(Wood).Fill(Cylinder{Center: q, Radius: 0.3, Y0: h, Y1: h + 1.2})
		g.Use(Foliage).FillVolume(Ellipsoid{q.At(h + 1.2 + r*0.8), V(r, r*0.85, r)})
	}
}

// medina builds the upper town: church, convent, ruins, houses, gardens and trees.
func (a Alhambra) medina(g *Grid) {
	h := alhambraH
	// Gardens on the south side of the Calle Real.
	g.Use(Grass).Fill(Func{Box{V(-37, h-0.4, -16), V(4, h, -5)}, func(p Vec3) bool {
		return InsidePolygon(alhambraRidge, p.Plan())
	}})
	// Santa María de la Alhambra with its bell tower.
	g.Use(White)
	a.shell(g, -8, -12, 0, -6.6, h, h+5, 0.8)
	g.Use(Terracotta).Fill(Gable{Min: P(-8.3, -12.3), Max: P(0.3, -6.3), Base: h + 5, Ridge: AxisX})
	a.arches(g, V(-4, h+2.6, -6.6), 3, 2.2, 0.35, 0.8, 1.6, AxisZ)
	g.Use(White)
	a.shell(g, 0, -9.8, 2.6, -7.2, h, h+9, 0.6)
	g.Carve(Arch{Sill: V(1.3, h+7, -8.5), HalfWidth: 0.5, Spring: 0.6, Depth: 4, Through: AxisX})
	g.Carve(Arch{Sill: V(1.3, h+7, -8.5), HalfWidth: 0.5, Spring: 0.6, Depth: 4, Through: AxisZ})
	g.Use(Gold).Fill(Block(V(1.0, h+7.4, -8.8), V(1.6, h+7.9, -8.2)))
	g.Use(Terracotta).Fill(HipRoof{Min: P(-0.3, -10.1), Max: P(2.9, -6.9), Base: h + 9})
	// Convent of San Francisco around its cloister.
	g.Use(White)
	cx0, cz0, cx1, cz1 := -31.0, -9.0, -22.0, 1.0
	a.shell(g, cx0, cz0, cx1, cz1, h, h+4.4, 0.7)
	g.Carve(Block(V(cx0+2.4, h+0.3, cz0+2.4), V(cx1-2.4, h+6, cz1-2.4)))
	g.Use(Terracotta).Fill(Func{Box{V(cx0-0.3, h+4.4, cz0-0.3), V(cx1+0.3, h+6.4, cz1+0.3)}, func(p Vec3) bool {
		d := math.Min(math.Min(p.X-cx0+0.3, cx1+0.3-p.X), math.Min(p.Z-cz0+0.3, cz1+0.3-p.Z))
		return d < 3 && p.Y <= h+4.6+0.6*math.Min(d, 3-d)
	}})
	g.Use(Grass).Fill(Block(V(cx0+2.4, h-0.4, cz0+2.4), V(cx1-2.4, h+0.1, cz1-2.4)))
	g.Use(Foliage).FillVolume(Ellipsoid{V((cx0+cx1)/2, h+1.3, (cz0+cz1)/2), V(1.3, 1.2, 1.3)})
	for i := 0; i < 4; i++ {
		x := cx0 + 1.6 + float64(i)*2
		a.arches(g, V(x, h+2, cz0), 1, 1, 0.35, 0.6, 1.6, AxisZ)
		a.arches(g, V(x, h+2, cz1), 1, 1, 0.35, 0.6, 1.6, AxisZ)
	}
	// Ruined houses of the medina: low walls on a grid.
	g.Use(Brick)
	for i := 0; i < 6; i++ {
		for j := 0; j < 4; j++ {
			x, z := -37.0+float64(i)*2.4, 2.0+float64(j)*2.6
			if x > -31 && z < 4 {
				continue
			}
			if !InsidePolygon(alhambraRidge, P(x+1.2, z+1.3)) || !InsidePolygon(alhambraRidge, P(x+2.4, z+2.6)) {
				continue
			}
			y := h + 0.4 + 0.8*Hash01(i, j, 61)
			g.Fill(Block(V(x, h, z), V(x+2.4, y, z+0.6)), Block(V(x, h, z), V(x+0.6, y, z+2.6)))
		}
	}
	g.Paint(DarkStone, Func{Box{V(-38, h, 1), V(-22, h+2, 13)}, func(p Vec3) bool {
		return Hash01(int(p.X*3), int(p.Y*3), int(p.Z*3), 63) < 0.3
	}})
	// Whitewashed houses with tiled roofs.
	houses := [][4]float64{{-20, -12.6, -16.6, -9.6}, {-15.6, -13.4, -12.2, -10.6}, {-20.4, -8.4, -17.4, -5.6},
		{-15.4, -9.2, -11.2, -6.4}, {-11.6, -13.6, -9, -10}, {-36, -6.4, -33, -3.4}, {-35.4, -2.4, -32.2, 0.6}}
	for i, b := range houses {
		top := h + 2.4 + 0.8*Hash01(i, 71)
		g.Use(White)
		a.shell(g, b[0], b[1], b[2], b[3], h, top, 0.5)
		g.Carve(Block(V((b[0]+b[2])/2-0.3, h+1, b[1]-0.3), V((b[0]+b[2])/2+0.3, h+1.9, b[1]+0.6)))
		g.Use(Terracotta).Fill(a.lowHip(b[0]-0.3, b[1]-0.3, b[2]+0.3, b[3]+0.3, top, 0.6))
	}
	// Trees in the gardens.
	for i := 0; i < 26; i++ {
		x := -36 + 40*Hash01(i, 81)
		z := -14 + 10*Hash01(i, 82)
		if x > -9 && x < 4 && z < -5 || x > -31.5 && x < -21.5 || x > -21 && x < -8 && z < -5 && Hash01(i, 83) < 0.5 {
			continue
		}
		if !InsidePolygon(alhambraRidge, P(x, z)) {
			continue
		}
		d, _ := alhambraEdge(alhambraRidge, P(x, z))
		if d < 1.8 {
			continue
		}
		g.Use(Foliage)
		if Hash01(i, 84) < 0.45 {
			g.FillVolume(Ellipsoid{V(x, h+2.3, z), V(0.6, 2.3, 0.6)})
		} else {
			r := 1.1 + 0.4*Hash01(i, 85)
			g.Use(Wood).Fill(Cylinder{Center: P(x, z), Radius: 0.3, Y0: h, Y1: h + 1.2})
			g.Use(Foliage).FillVolume(Ellipsoid{V(x, h+1.2+r*0.8, z), V(r, r*0.85, r)})
		}
	}
}

// generalife builds the summer palace: the Patio de la Acequia with its jets, the
// pavilions and arcade, and the lower gardens with long pools and cypresses.
func (a Alhambra) generalife(g *Grid) {
	y := alhambraGH
	// Patio de la Acequia: the long channel, jets and flower beds.
	px0, px1 := -68.5, -58.6
	g.Use(Stone).Fill(Block(V(px0, y-0.4, -2.6), V(px1, y, 2.6)))
	g.Use(Grass).Fill(Block(V(px0+0.4, y-0.4, -2.2), V(px1-0.4, y+0.3, -0.9)), Block(V(px0+0.4, y-0.4, 0.9), V(px1-0.4, y+0.3, 2.2)))
	g.Use(Water).Fill(Block(V(px0, y-0.4, -0.6), V(px1, y, 0.6)))
	g.Fill(Func{Box{V(px0+0.4, y, -2.4), V(px1-0.4, y+0.8, 2.4)}, func(p Vec3) bool {
		return math.Abs(p.Z) > 0.9 && Hash01(int(p.X*2.5), int(p.Z*2.5), 91) < 0.3
	}})
	g.Paint(Red, Func{Box{V(px0, y, -2.4), V(px1, y+0.8, 2.4)}, func(p Vec3) bool {
		return math.Abs(p.Z) > 0.9 && Hash01(int(p.X*2.5), int(p.Z*2.5), 92) < 0.5
	}})
	g.Use(Glass)
	for x := px0 + 0.8; x < px1-0.6; x += 1.4 {
		for _, s := range []float64{-1, 1} {
			g.Fill(Capsule{V(x, y+0.2, s*1.0), V(x, y+1.1, s*0.45), 0.28})
			g.Fill(Capsule{V(x, y+1.1, s*0.45), V(x, y+0.4, 0), 0.28})
		}
	}
	// North pavilion (the palace) at the east end, with a portico and mirador.
	g.Use(White)
	a.shell(g, -71.6, -5.2, -68.5, 5.2, y, y+5.4, 0.8)
	g.Carve(Block(V(-69.6, y+0.3, -3.6), V(-68.0, y+3.2, 3.6)))
	a.darken(g, -69.6, -3.6, -68.0, 3.6, y+0.3, y+3.2)
	g.Use(White).Fill(Block(V(-68.9, y, -3.8), V(-68.5, y+3.6, 3.8)))
	a.arches(g, V(-68.7, y, 0), 5, 1.4, 0.45, 1.6, 1.2, AxisX)
	a.lattice(g, V(-71.6, y+3.8, 0), 4, 1.8, 0.35, 0.6, AxisX, Charcoal)
	g.Use(Terracotta).Fill(a.lowHip(-71.9, -5.5, -68.2, 5.5, y+5.4, 0.6))
	g.Use(White)
	a.shell(g, -71.8, 2.2, -69.0, 5.2, y+5.4, y+9, 0.6)
	a.arches(g, V(-70.4, y+7, 5.2), 2, 1.1, 0.35, 0.7, 1.4, AxisZ)
	a.arches(g, V(-69.0, y+7, 3.7), 2, 1.1, 0.35, 0.7, 1.4, AxisX)
	g.Use(Terracotta).Fill(a.lowHip(-72.1, 1.9, -68.7, 5.5, y+9, 0.8))
	// South pavilion at the west end.
	g.Use(White)
	a.shell(g, -58.6, -4.2, -56.2, 4.2, y, y+4.4, 0.7)
	a.arches(g, V(-58.6, y, 0), 3, 1.6, 0.5, 1.6, 1.2, AxisX)
	a.arches(g, V(-56.2, y+1.4, 0), 3, 1.6, 0.4, 0.9, 1.2, AxisX)
	g.Use(Terracotta).Fill(a.lowHip(-58.9, -4.5, -55.9, 4.5, y+4.4, 0.6))
	// Arcaded gallery along the north side and the wing along the south side.
	g.Use(White).Fill(Block(V(px0, y, 2.6), V(px1, y+3.4, 4.0)))
	for i := 0; i < 11; i++ {
		x := px0 + 0.45 + (px1-px0-0.9)*(float64(i)+0.5)/11
		g.Carve(Arch{Sill: V(x, y, 3.3), HalfWidth: 0.38, Spring: 1.7, Depth: 2, Through: AxisZ})
	}
	g.Use(Terracotta).Fill(Gable{Min: P(px0, 2.4), Max: P(px1, 4.2), Base: y + 3.4, Ridge: AxisX})
	g.Use(White)
	a.shell(g, px0, -4.4, px1, -2.6, y, y+4.2, 0.5)
	a.arches(g, V((px0+px1)/2, y+1.8, -2.6), 6, 1.6, 0.35, 0.7, 1, AxisZ)
	a.arches(g, V((px0+px1)/2, y+1.8, -4.4), 6, 1.6, 0.35, 0.7, 1, AxisZ)
	g.Use(Terracotta).Fill(Gable{Min: P(px0, -4.7), Max: P(px1, -2.3), Base: y + 4.2, Ridge: AxisX})
	// Hedges and cypresses around the upper terrace.
	g.Use(Foliage)
	for _, z := range []float64{-7.2, 6.2} {
		g.Fill(Block(V(-70.5, y, z), V(-56.5, y+0.9, z+0.7)))
	}
	for i := 0; i < 7; i++ {
		x := -70.0 + float64(i)*2.2
		for _, z := range []float64{-5.8, 5.4} {
			g.FillVolume(Ellipsoid{V(x, y+2.2, z), V(0.6, 2.2, 0.6)})
		}
	}
	// Lower gardens: a long pool with jets between clipped hedges and cypresses.
	y = alhambraGM
	g.Use(Sand).Fill(Block(V(-55.6, y-0.4, -2.6), V(-45.4, y, 2.6)))
	g.Use(White).Fill(Block(V(-55, y-0.4, -1.1), V(-46, y+0.4, 1.1)))
	g.Use(Water).Fill(Block(V(-54.6, y-0.4, -0.7), V(-46.4, y+0.2, 0.7)))
	g.Use(Glass)
	for x := -53.6; x < -46.6; x += 1.6 {
		g.Fill(Cylinder{Center: P(x, 0), Radius: 0.3, Y0: y + 0.2, Y1: y + 1.4})
	}
	g.Use(Foliage)
	for _, z := range []float64{-9.6, -6.0, 3.0, 6.6} {
		for x := -55.0; x < -46; x += 3.2 {
			g.Fill(Block(V(x, y, z), V(x+2.4, y+0.8, z+2.4)))
			g.Use(Grass).Fill(Block(V(x+0.6, y-0.4, z+0.6), V(x+1.8, y+0.8, z+1.8)))
			g.Use(Foliage)
		}
	}
	for x := -55.0; x <= -46; x += 1.8 {
		for _, z := range []float64{-3.5, 3.5} {
			g.FillVolume(Ellipsoid{V(x, y+2.3, z), V(0.6, 2.3, 0.6)})
		}
	}
	// Retaining wall below the upper terrace.
	g.Use(Terracotta).Fill(Block(V(-56.6, alhambraGM, -11), V(-55.8, alhambraGH+0.8, 9)))
	g.Carve(Block(V(-56.8, alhambraGM, -1.2), V(-55.6, alhambraGH+1, 1.2)))
	for s := 0; s < 6; s++ {
		g.Fill(Block(V(-55.8+float64(s)*0.5, alhambraGM, -1.2), V(-55.3+float64(s)*0.5, alhambraGH-float64(s)*0.5, 1.2)))
	}
}

// darro adds the river's bridges and the houses of the Carrera del Darro.
func (a Alhambra) darro(g *Grid, m *alhambraMap) {
	for _, bx := range []float64{-26.0, 8.0} {
		z := alhambraRiverZ(bx)
		g.Use(Stone).Fill(Subtract(Block(V(bx-1.2, 0.2, z-4.2), V(bx+1.2, 2.6, z+4.2)),
			Arch{Sill: V(bx, 0, z), HalfWidth: 2.2, Spring: 0.1, Depth: 3, Through: AxisZ}))
		g.Use(DarkStone).Fill(Block(V(bx-1.2, 2.6, z-4.2), V(bx-0.8, 3.0, z+4.2)), Block(V(bx+0.8, 2.6, z-4.2), V(bx+1.2, 3.0, z+4.2)))
	}
	i := 0
	for row := 0; row < 2; row++ {
		for x := -44.0 + 1.8*float64(row); x < 40; x += 3.6 {
			i++
			if Hash01(i, 101) < 0.2+0.25*float64(row) || math.Abs(x+26) < 3 || math.Abs(x-8) < 3 {
				continue
			}
			z := alhambraRiverZ(x) + 3.4 + 2.9*float64(row) + 0.3*Hash01(i, 102)
			w, d := 2.4+0.8*Hash01(i, 103), 2.4
			if z+d > alhambraZ1-0.5 {
				continue
			}
			base := m.at(x, z+d/2) - 0.3
			top := base + 2.4 + 1.4*Hash01(i, 104)
			g.Use(White)
			if Hash01(i, 105) < 0.3 {
				g.Use(Sand)
			}
			a.shell(g, x-w/2, z, x+w/2, z+d, base-1.2, top, 0.5)
			g.Carve(Block(V(x-0.3, top-1.6, z-0.3), V(x+0.3, top-0.8, z+0.6)))
			g.Use(Terracotta).Fill(Gable{Min: P(x-w/2-0.3, z-0.3), Max: P(x+w/2+0.3, z+d+0.3), Base: top, Ridge: AxisX})
			if Hash01(i, 106) < 0.35 {
				cz := z + d + 0.9
				g.Use(Foliage).FillVolume(Ellipsoid{V(x+w/2+0.6, m.at(x+w/2+0.6, cz)+2, cz), V(0.6, 2.2, 0.6)})
			}
		}
	}
}

// woods scatters the Alhambra forest over the slopes, with cypresses here and there.
func (a Alhambra) woods(g *Grid, m *alhambraMap) {
	const step = 2.3
	for i := int(math.Floor(alhambraX0 / step)); i <= int(math.Floor(alhambraX1/step)); i++ {
		for j := int(math.Floor(alhambraZ0 / step)); j <= int(math.Floor(alhambraZ1/step)); j++ {
			x := float64(i)*step + (Hash01(i, j, 111)-0.5)*1.6
			z := float64(j)*step + (Hash01(i, j, 112)-0.5)*1.6
			hh, kind, ok := a.height(x, z)
			if !ok || kind != alhambraSlope || hh < 1.4 {
				continue
			}
			if (m.at(x, z)-m.low(x, z))/0.9 > 1.6 || a.rim(x, z) < 2 {
				continue
			}
			if d, _ := alhambraEdge(alhambraRidge, P(x, z)); d < 3 {
				continue
			}
			if math.Abs(z-alhambraRiverZ(x)) < 4 {
				continue
			}
			skip := 0.2
			if z > 0 && x > -42 {
				skip = 0.6
			}
			if Hash01(i, j, 113) < skip {
				continue
			}
			if Hash01(i, j, 114) < 0.15 {
				g.Use(Foliage).FillVolume(Ellipsoid{V(x, hh+2.2, z), V(0.6, 2.4, 0.6)})
				continue
			}
			r := 1.1 + 0.6*Hash01(i, j, 115)
			c := Foliage
			if Hash01(i, j, 116) < 0.25 {
				c = Grass
			}
			g.Use(c).FillVolume(Ellipsoid{V(x, hh+r*0.6, z), V(r, r*0.85, r)})
		}
	}
}
