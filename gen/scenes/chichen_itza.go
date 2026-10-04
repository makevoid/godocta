package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// ChichenItza is the Maya-Toltec city in the Yucatán jungle, centered on El Castillo
// (the Temple of Kukulcán): nine sloped terraces with recessed panels, four 45°
// staircases with balustrades, feathered-serpent heads at the foot of the north stair
// and the temple on top with its serpent-column entrance. East of it stands the Temple
// of the Warriors with its colonnade and Chac Mool, with the Group of a Thousand Columns
// to the south; north-west lies the Great Ball Court with its rings and the Temple of
// the Jaguars; south-west the round El Caracol observatory. A white sacbe runs north to
// the Sacred Cenote, a round sinkhole in the jungle that rings the cleared plaza and
// slopes down to a low rim. North (the serpent stair) faces +Z, so east is -X.
type ChichenItza struct{}

func init() { Register(ChichenItza{}) }

func (ChichenItza) Name() string  { return "chichen_itza" }
func (ChichenItza) Title() string { return "Chichen Itza" }

const (
	chichenItzaG    = 3.0  // Plaza ground top.
	chichenItzaH    = 2.1  // El Castillo terrace height.
	chichenItzaR0   = 22.0 // El Castillo base half-size.
	chichenItzaSet  = 1.75 // El Castillo setback per terrace.
	chichenItzaTop  = chichenItzaG + 9*chichenItzaH
	chichenItzaCenZ = 57.0 // Sacred Cenote center (x = 0).
	chichenItzaCenR = 9.5  // Cenote radius.
	chichenItzaSX   = -6.0 // Site center and superelliptic semi-axes.
	chichenItzaSZ   = 14.0
	chichenItzaSA   = 60.0
	chichenItzaSB   = 56.0
)

func (ChichenItza) clamp01(v float64) float64 { return math.Max(0, math.Min(1, v)) }

func (c ChichenItza) smooth(a, b, x float64) float64 {
	t := c.clamp01((x - a) / (b - a))
	return t * t * (3 - 2*t)
}

func (ChichenItza) mod(v, m float64) float64 { return math.Mod(math.Mod(v, m)+m, m) }

// set paints one lattice cell, replacing whatever was there.
func (ChichenItza) set(g *Grid, cell Cell, col Color) {
	g.Remove(cell)
	g.Use(col)
	g.Add(cell)
}

// rho is the site's superelliptic radius: 1 at the outer rim.
func (ChichenItza) rho(x, z float64) float64 {
	u, v := math.Pow((x-chichenItzaSX)/chichenItzaSA, 2), math.Pow((z-chichenItzaSZ)/chichenItzaSB, 2)
	return math.Pow(u*u*u+v*v*v, 1.0/6)
}

// crown is the jungle canopy bump at a plan point and a tone for the dominant tree.
func (ChichenItza) crown(x, z float64) (float64, float64) {
	const s = 3.4
	i0, j0 := int(math.Floor(x/s)), int(math.Floor(z/s))
	best, tone := 0.0, 0.5
	for i := i0 - 1; i <= i0+1; i++ {
		for j := j0 - 1; j <= j0+1; j++ {
			cx := (float64(i) + 0.2 + 0.6*Hash01(i, j, 1)) * s
			cz := (float64(j) + 0.2 + 0.6*Hash01(i, j, 2)) * s
			r := 2.0 + 1.6*Hash01(i, j, 3)
			d := math.Hypot(x-cx, z-cz) / r
			if d < 1 {
				v := (0.6 + 1.8*Hash01(i, j, 4)) * math.Sqrt(1-d*d)
				if v > best {
					best, tone = v, Hash01(i, j, 5)
				}
			}
		}
	}
	return best, tone
}

// edge is the plaza's radius (in rho) towards plan angle th from the site center.
func (ChichenItza) edge(th float64) float64 {
	return 0.81 + 0.02*math.Sin(5*th+1) + 0.012*math.Sin(11*th+2)
}

// chichenItzaSites are the monument footprints (x0, z0, x1, z1) kept clear of jungle.
var chichenItzaSites = [][4]float64{
	{-27, -27, 27, 27},  // El Castillo.
	{-57, -17, -30, 16}, // Temple of the Warriors and the Thousand Columns.
	{26, -3, 46, 41},    // Great Ball Court.
	{22, -33, 37, -22},  // El Caracol.
	{-21, 32, -9, 44},   // Platform of Venus.
	{10, 26, 24, 36},    // Tzompantli.
	{-4.6, 15, 4.6, 50}, // Sacbe.
}

// site reports whether a plan point lies within margin of a monument footprint.
func (ChichenItza) site(x, z, margin float64) bool {
	for _, r := range chichenItzaSites {
		if x > r[0]-margin && x < r[2]+margin && z > r[1]-margin && z < r[3]+margin {
			return true
		}
	}
	return false
}

// clearing is cleared ground inside the jungle: monuments, the sacbe and the cenote.
func (c ChichenItza) clearing(x, z float64) bool {
	return c.site(x, z, 1.5) || math.Hypot(x, z-chichenItzaCenZ) < chichenItzaCenR+3.5
}

// ground is the terrain top at a plan point: the flat plaza or the jungle canopy.
func (c ChichenItza) ground(x, z float64) (h float64, jungle bool, tone float64, ok bool) {
	r := c.rho(x, z)
	if r > 1 {
		return 0, false, 0, false
	}
	th := math.Atan2(z-chichenItzaSZ, x-chichenItzaSX)
	thr := c.edge(th)
	if r <= thr || c.clearing(x, z) {
		return chichenItzaG, false, 0, true
	}
	t := c.clamp01((r - thr) / (1 - thr))
	base := 0.5 + (chichenItzaG+2.0-0.5)*(1-c.smooth(0.3, 1, t))
	b, tone := c.crown(x, z)
	return base + b*(1-0.5*t), true, tone, true
}

// terrain lays the plaza shell, the jungle canopy sloping to the rim and the cenote.
func (c ChichenItza) terrain(g *Grid) {
	s := g.Span()
	G := chichenItzaG
	lo := func(v float64) int32 { return int32(math.Floor(v * 2 / s)) }
	for par := int32(0); par < 2; par++ {
		for X := lo(chichenItzaSX - chichenItzaSA - 1); X <= lo(chichenItzaSX+chichenItzaSA)+1; X++ {
			if X&1 != par {
				continue
			}
			for Z := lo(chichenItzaSZ - chichenItzaSB - 1); Z <= lo(chichenItzaSZ+chichenItzaSB)+1; Z++ {
				if Z&1 != par {
					continue
				}
				x, z := float64(X)*s/2, float64(Z)*s/2
				column := func(bot, top float64, color func(y float64) Color) {
					for Y := lo(bot); Y <= lo(top)+1; Y++ {
						y := float64(Y) * s / 2
						if Y&1 != par || y < bot || y > top {
							continue
						}
						c.set(g, Cell{X, Y, Z}, color(y))
					}
				}
				dc := math.Hypot(x, z-chichenItzaCenZ)
				if dc < chichenItzaCenR {
					column(0, 0.55, func(float64) Color { return Water })
					continue
				}
				if dc < chichenItzaCenR+1.4 {
					a := math.Atan2(z-chichenItzaCenZ, x)
					k := int(math.Floor(a * 7))
					vine := Hash01(k, 7) < 0.3
					vineEnd := G - 0.6 - 1.4*Hash01(k, 8)
					column(0, G, func(y float64) Color {
						if vine && y > vineEnd {
							return Foliage
						}
						switch v := Hash01(int(math.Floor(y*1.6)), k/2, 9); {
						case v < 0.25:
							return DarkStone
						case v < 0.45:
							return Sand
						}
						return Stone
					})
					continue
				}
				h, jungle, tone, ok := c.ground(x, z)
				if !ok {
					continue
				}
				low := h
				for _, o := range [4]Vec2{{0.8, 0}, {-0.8, 0}, {0, 0.8}, {0, -0.8}} {
					nh, _, _, nok := c.ground(x+o.X, z+o.Y)
					if !nok {
						nh = 0.2
					}
					low = math.Min(low, nh)
				}
				color := Grass
				if jungle {
					color = Foliage
					switch {
					case tone < 0.2:
						color = Grass
					case tone > 0.985:
						color = Gold
					}
				}
				column(low-0.8, h, func(float64) Color { return color })
			}
		}
	}
	// Worn sand apron around El Castillo and paths between the monuments.
	G2 := G + 0.1
	g.Paint(Sand,
		Subtract(BlockAt(P(0, 0), 49, 49, G-0.5, G2), BlockAt(P(0, 0), 44, 44, G-0.5, G2)),
		Capsule{V(-26, G, 0), V(-37, G, 6), 1.3},
		Capsule{V(26, G, 0), V(27, G, 8), 1.3},
		Capsule{V(0, G, -26), V(22, G, -27), 1.1},
		Capsule{V(-3, G, 38), V(-11, G, 38), 1.0},
		Capsule{V(3, G, 31), V(10, G, 31), 1.0},
		Capsule{V(24, G, 32), V(29, G, 42), 1.0},
		Capsule{V(-22, G, -22), V(-30, G, -12), 1.0},
		Capsule{V(36, G, -27), V(38, G, -5), 1.0},
		Capsule{V(-26, G, 0), V(-22, G, 22), 1.0},
		Capsule{V(-22, G, 22), V(-15, G, 33), 1.0},
	)
}

// stepped is the outer half-size at height y of a terraced pyramid: n terraces of
// height h, each set back by set, with a sloped talud and a projecting cornice.
// It returns -1 above the top or below the foundation.
func (ChichenItza) stepped(y, base, h, r0, set float64, n int) float64 {
	if y < base-0.8 {
		return -1
	}
	f := (y - base) / h
	if f < 0 {
		return r0
	}
	k := math.Floor(f)
	t := f - k
	if int(k) >= n {
		if f-float64(n) > 0.05 {
			return -1
		}
		k, t = float64(n-1), 1
	}
	r := r0 - k*set
	if t < 0.72 {
		return r - 0.5*t/0.72
	}
	return r - 0.15
}

// pyramid fills a hollow terraced pyramid shell centered on c.
func (cz ChichenItza) pyramid(g *Grid, c Vec2, base, h, r0, set float64, n int, shell float64) {
	top := base + float64(n)*h
	box := Box{V(c.X-r0-0.5, base-0.8, c.Y-r0-0.5), V(c.X+r0+0.5, top+0.2, c.Y+r0+0.5)}
	g.Use(Stone).Fill(Func{box, func(p Vec3) bool {
		r := cz.stepped(p.Y, base, h, r0, set, n)
		m := math.Max(math.Abs(p.X-c.X), math.Abs(p.Z-c.Y))
		if r < 0 || m > r {
			return false
		}
		ri := cz.stepped(p.Y+shell, base, h, r0, set, n)
		return !(p.Y > base && ri > 0 && m <= ri-shell)
	}})
}

// stair builds a 45° staircase descending from y1 at distance uTop from center along
// dir down to y0, with balustrades; it returns the distance of its foot.
func (ChichenItza) stair(g *Grid, center, dir Vec2, half, bal, y0, y1, uTop, step float64, tread Color) float64 {
	perp := P(-dir.Y, dir.X)
	uEnd := uTop + (y1 - y0)
	w := half + bal
	box := EmptyBox()
	for _, u := range []float64{uTop - 1, uEnd + 0.5} {
		for _, a := range []float64{-w - 0.5, w + 0.5} {
			q := center.Add(dir.Scale(u)).Add(perp.Scale(a))
			box = box.Union(Box{q.At(y0 - 0.8), q.At(y1 + 1.2)})
		}
	}
	loc := func(p Vec3) (float64, float64) {
		d := p.Plan().Sub(center)
		return d.Dot(dir), math.Abs(d.Dot(perp))
	}
	stepTop := func(u float64) float64 {
		if u <= uTop {
			return y1
		}
		return y1 - math.Ceil((u-uTop)/step-1e-9)*step
	}
	g.Use(tread).Fill(Func{box, func(p Vec3) bool {
		u, a := loc(p)
		if u < uTop-1 || u > uEnd || a > half || p.Y < y0-0.8 {
			return false
		}
		t := stepTop(u)
		return p.Y <= t && p.Y >= t-1.5
	}})
	g.Use(Stone).Fill(Func{box, func(p Vec3) bool {
		u, a := loc(p)
		if u < uTop-1 || u > uEnd+0.4 || a <= half || a > w || p.Y < y0-0.8 {
			return false
		}
		return p.Y <= math.Min(y1, y1-(u-uTop))+0.9
	}})
	return uEnd
}

// serpentHead is a feathered-serpent head lying at the foot of a balustrade: at plan
// point at, its open jaws facing dir, resting on y, scaled by k.
func (ChichenItza) serpentHead(g *Grid, at, dir Vec2, y, k float64) {
	perp := P(-dir.Y, dir.X)
	part := func(u0, u1, a0, a1, y0, y1 float64) Solid {
		u0, u1, a0, a1, y0, y1 = u0*k, u1*k, a0*k, a1*k, y0*k, y1*k
		box := EmptyBox()
		for _, u := range []float64{u0, u1} {
			for _, a := range []float64{a0, a1} {
				q := at.Add(dir.Scale(u)).Add(perp.Scale(a))
				box = box.Union(Box{q.At(y + y0), q.At(y + y1)})
			}
		}
		return Func{box, func(p Vec3) bool {
			d := p.Plan().Sub(at)
			u, a := d.Dot(dir), d.Dot(perp)
			return u >= u0 && u <= u1 && a >= a0 && a <= a1
		}}
	}
	g.Use(Sand).Fill(
		part(-0.6, 2.2, -1.0, 1.0, 1.0, 2.3), // Upper jaw and brow.
		part(1.4, 2.5, -0.8, 0.8, 2.3, 2.9),  // Curled snout.
		part(-0.6, 2.0, -1.0, 1.0, 0, 0.55),  // Lower jaw.
	)
	g.Use(Charcoal).Fill(part(-0.6, 0.6, -0.9, 0.9, 0.55, 1.0))
	g.Use(Red).Fill(part(0.6, 2.9, -0.35, 0.35, 0.3, 0.75)) // Forked tongue.
	g.Use(White).Fill(
		part(1.5, 2.2, -1.0, -0.55, 0.55, 1.0), // Fangs.
		part(1.5, 2.2, 0.55, 1.0, 0.55, 1.0),
	)
	g.Paint(DarkStone,
		part(0.1, 0.9, -1.0, -0.6, 1.5, 2.1), // Eyes.
		part(0.1, 0.9, 0.6, 1.0, 1.5, 2.1),
		part(-0.6, 2.2, -1.0, 1.0, 2.0, 2.3), // Feather crest.
	)
}

func (c ChichenItza) castillo(g *Grid) {
	G, h, top := chichenItzaG, chichenItzaH, chichenItzaTop
	R0, set := chichenItzaR0, chichenItzaSet
	o := P(0, 0)
	c.pyramid(g, o, G, h, R0, set, 9, 1.5)
	box := Box{V(-R0-0.5, G-0.8, -R0-0.5), V(R0+0.5, top+0.2, R0+0.5)}
	rOut := func(y float64) float64 { return c.stepped(y, G, h, R0, set, 9) }
	// Weathering: dark lichen patches, heavier on the lower courses.
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		v := Hash01(int(math.Floor(p.X/1.7)), int(math.Floor(p.Y/1.3)), int(math.Floor(p.Z/1.7)), 21)
		return p.Y > G && v < 0.06*(1-(p.Y-G)/(top-G)*0.7)
	}})
	// Recessed panels along every terrace.
	panel := func(p Vec3) (bool, float64) {
		r := rOut(p.Y)
		f := (p.Y - G) / h
		k := math.Floor(f)
		t := f - k
		if r < 0 || f < 0 || k > 8 || t < 0.14 || t > 0.64 {
			return false, 0
		}
		ax, az := math.Abs(p.X), math.Abs(p.Z)
		a, m := ax, az
		if ax > az {
			a, m = az, ax
		}
		if a > R0-k*set-1.3 || c.mod(a+0.6, 2.4) < 0.8 {
			return false, 0
		}
		return true, r - m
	}
	g.Carve(Func{box, func(p Vec3) bool { ok, d := panel(p); return ok && d < 0.45 }})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool { ok, d := panel(p); return ok && d >= 0.45 && d < 1.3 }})

	// Four staircases; the north one (+Z) ends in serpent heads.
	uTop := rOut(top) - 0.5
	for _, dir := range []Vec2{{0, 1}, {0, -1}, {1, 0}, {-1, 0}} {
		uEnd := c.stair(g, o, dir, 3.4, 1.3, G, top, uTop, 0.9, White)
		if dir.Y == 1 {
			perp := P(-dir.Y, dir.X)
			for _, s := range []float64{-1, 1} {
				c.serpentHead(g, dir.Scale(uEnd-0.8).Add(perp.Scale(s*4.05)), dir, G, 1.6)
			}
		}
	}
	c.castilloTemple(g, top)
}

// castilloTemple is the temple crowning El Castillo: thick walls, a projecting frieze,
// a dark sanctuary, single doors east, south and west and the serpent-column portico
// north (+Z).
func (c ChichenItza) castilloTemple(g *Grid, y float64) {
	const hw, ci = 5.2, 4.1
	o := P(0, 0)
	wallTop, frieze := y+3.4, y+5.0
	g.Use(Stone).Fill(BlockAt(o, 2*hw, 2*hw, y-0.5, wallTop))
	g.Fill(BlockAt(o, 2*hw+0.8, 2*hw+0.8, wallTop, frieze))
	box := Box{V(-hw-1, y, -hw-1), V(hw+1, frieze+1.5, hw+1)}
	m := func(p Vec3) float64 { return math.Max(math.Abs(p.X), math.Abs(p.Z)) }
	// Frieze: pale mouldings framing dark recessed panels.
	g.Paint(White, Func{box, func(p Vec3) bool {
		return p.Y >= wallTop && (p.Y <= wallTop+0.45 || p.Y >= frieze-0.45) && m(p) > hw-0.2
	}})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		a := math.Min(math.Abs(p.X), math.Abs(p.Z))
		return p.Y > wallTop+0.6 && p.Y < frieze-0.6 && m(p) > hw-0.1 && c.mod(a+0.5, 2.0) > 1.0
	}})
	// Roof crest of small merlons.
	g.Use(Stone).Fill(Func{Box{V(-hw-0.4, frieze, -hw-0.4), V(hw+0.4, frieze+1.0, hw+0.4)}, func(p Vec3) bool {
		a := math.Min(math.Abs(p.X), math.Abs(p.Z))
		return m(p) > hw-0.5 && c.mod(a, 1.8) < 0.9
	}})
	// Sanctuary and doors.
	g.Carve(BlockAt(o, 2*ci, 2*ci, y+0.05, frieze-0.8))
	for _, x := range []float64{-2.9, 0, 2.9} {
		g.Carve(BlockAt(P(x, hw-0.5), 1.7, 3, y+0.05, y+2.8))
	}
	g.Carve(BlockAt(P(0, -hw+0.5), 1.9, 3, y+0.05, y+2.8), BlockAt(P(hw-0.5, 0), 3, 1.9, y+0.05, y+2.8), BlockAt(P(-hw+0.5, 0), 3, 1.9, y+0.05, y+2.8))
	g.Paint(Charcoal, BlockAt(o, 2*(ci+0.3), 2*(ci+0.3), y-0.6, frieze-0.55))
	// Serpent columns between the north openings: heads at the floor, rattles under the lintel.
	for _, s := range []float64{-1, 1} {
		x := s * 1.45
		g.Use(Sand).Fill(BlockAt(P(x, hw-0.4), 1.1, 1.0, y, wallTop))
		g.Fill(BlockAt(P(x, hw+0.6), 1.1, 1.4, y, y+0.9))
		g.Fill(BlockAt(P(x, hw+0.3), 1.1, 0.8, wallTop-0.7, wallTop))
		g.Paint(Red, BlockAt(P(x, hw+1.1), 0.6, 0.6, y, y+0.5))
	}
}

// warriors builds the Temple of the Warriors east (-X) of El Castillo, its colonnade
// facing the plaza and the Group of a Thousand Columns to the south.
func (c ChichenItza) warriors(g *Grid) {
	G := chichenItzaG
	w := P(-48, 6)
	const h, r0, set = 1.4, 8.5, 1.0
	top := G + 4*h
	c.pyramid(g, w, G, h, r0, set, 4, 1.4)
	box := Box{V(w.X-r0, G, w.Y-r0), V(w.X+r0, top, w.Y+r0)}
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		v := Hash01(int(math.Floor(p.X/1.5)), int(math.Floor(p.Y/1.2)), int(math.Floor(p.Z/1.5)), 31)
		return v < 0.14
	}})
	g.Paint(White, Func{box, func(p Vec3) bool { // Pale cornices.
		f := (p.Y - G) / h
		r := c.stepped(p.Y, G, h, r0, set, 4)
		return r > 0 && f-math.Floor(f) >= 0.74 && math.Max(math.Abs(p.X-w.X), math.Abs(p.Z-w.Y)) > r-0.6
	}})
	rTop := c.stepped(top, G, h, r0, set, 4)
	c.stair(g, w, P(1, 0), 2.0, 0.8, G, top, rTop-0.3, 0.7, White)

	// Roofless temple on the platform's rear half, its walls broken to varying heights.
	x0, x1, z0, z1 := w.X-5.6, w.X+2.2, w.Y-5.0, w.Y+5.0
	tb := Box{V(x0, top, z0), V(x1, top+3.2, z1)}
	g.Use(Stone).Fill(Func{tb, func(p Vec3) bool {
		in := p.X > x0+1 && p.X < x1-1 && p.Z > z0+1 && p.Z < z1-1
		if in {
			return false
		}
		if p.X >= x1-1 && math.Abs(p.Z-w.Y) < 2.6 {
			return false // Entrance.
		}
		k := int(math.Floor((p.X + p.Z) / 1.6))
		return p.Y <= top+1.0+1.9*Hash01(k, 32)
	}})
	for i, x := range []float64{w.X - 3.8, w.X - 1.4} {
		for j, z := range []float64{-3.0, -1.0, 1.0, 3.0} {
			ht := 2.6
			if Hash01(i, j, 33) < 0.3 {
				ht = 1.2
			}
			g.Use(Stone).Fill(BlockAt(P(x, w.Y+z), 0.9, 0.9, top, top+ht))
		}
	}
	// Serpent columns flanking the entrance.
	for _, s := range []float64{-1, 1} {
		cz := w.Y + s*1.6
		g.Use(Sand).Fill(BlockAt(P(x1-0.5, cz), 1.0, 1.0, top, top+3.4))
		g.Fill(BlockAt(P(x1+0.4, cz), 1.6, 1.0, top, top+1.0))
		g.Fill(BlockAt(P(x1+0.2, cz), 1.2, 1.0, top+2.8, top+3.6))
		g.Paint(Red, BlockAt(P(x1+1.0, cz), 0.5, 0.6, top, top+0.5))
	}
	// Chac Mool reclining before the entrance, head turned to the plaza.
	cm := P(x1+2.1, w.Y)
	g.Use(Sand).Fill(
		BlockAt(cm, 1.0, 2.6, top, top+0.5),
		BlockAt(cm.Add(P(0, 0.8)), 1.0, 0.9, top, top+1.3),  // Raised knees.
		BlockAt(cm.Add(P(0, -0.9)), 1.0, 0.9, top, top+1.0), // Propped torso.
		BlockAt(cm.Add(P(0.3, -1.2)), 1.1, 0.8, top+0.9, top+1.7),
	)
	g.Paint(Gold, BlockAt(cm, 0.8, 0.6, top+0.5, top+0.9))

	// Hall of the columns in front: square pillars on a dais, the stair passing through.
	dx0, dx1 := w.X+r0, w.X+r0+8
	g.Use(Sand).Fill(Block(V(dx0, G-0.5, -2), V(dx1, G+0.4, 16)))
	for i, x := range []float64{dx0 + 1.3, dx0 + 4.0, dx0 + 6.7} {
		for j := 0; j < 7; j++ {
			z := -0.8 + 2.6*float64(j)
			if math.Abs(z-w.Y) < 3.4 {
				continue
			}
			c.pillar(g, P(x, z), G+0.4, 3.0, false, i, j, 34)
		}
	}
	// Group of a Thousand Columns: round columns south of the temple, a bench wall behind.
	g.Use(Sand).Fill(Block(V(-55, G-0.5, -14), V(-30, G+0.35, -5.5)))
	for i := 0; i < 9; i++ {
		for j, z := range []float64{-7, -9.8, -12.6} {
			c.pillar(g, P(-53.6+2.8*float64(i), z), G+0.35, 2.9, true, i, j, 35)
		}
	}
	g.Use(Stone).Fill(Block(V(-56, G-0.5, -16.4), V(-29, G+2.0, -15.0)), Block(V(-56, G-0.5, -15.0), V(-29, G+0.9, -14.0)))
}

// pillar is a square or round column with a dark base and a pale capital; some are
// broken off.
func (ChichenItza) pillar(g *Grid, at Vec2, y, ht float64, round bool, i, j, seed int) {
	broken := Hash01(i, j, seed) < 0.22
	if broken {
		ht *= 0.35 + 0.35*Hash01(i, j, seed+1)
	}
	if round {
		g.Use(Stone).Fill(Cylinder{Center: at, Radius: 0.65, Y0: y, Y1: y + ht})
	} else {
		g.Use(Stone).Fill(BlockAt(at, 1.1, 1.1, y, y+ht))
	}
	g.Paint(DarkStone, BlockAt(at, 1.4, 1.4, y, y+0.5))
	if !broken {
		g.Use(White).Fill(BlockAt(at, 1.5, 1.5, y+ht, y+ht+0.4))
	}
}

// ballCourt is the Great Ball Court north-west of El Castillo: two long walls with
// sloped benches and stone rings, the Temple of the Jaguars on the east wall and
// temples closing the north and south ends.
func (c ChichenItza) ballCourt(g *Grid) {
	G := chichenItzaG
	const z0, z1 = 4.0, 34.0
	wallTop := G + 5.0
	for _, side := range []struct{ outer, inner, sign float64 }{{31, 34, 1}, {45, 42, -1}} {
		xa, xb := math.Min(side.outer, side.inner), math.Max(side.outer, side.inner)
		g.Use(Stone).Fill(Block(V(xa, G-0.8, z0), V(xb, wallTop, z1)))
		g.Carve(Block(V(xa+1, G, z0+1), V(xb-1, wallTop-0.9, z1-1)))
		g.Paint(DarkStone, Func{Box{V(xa, G, z0), V(xb, wallTop, z1)}, func(p Vec3) bool {
			return Hash01(int(math.Floor(p.Y/1.1)), int(math.Floor(p.Z/1.6)), int(xa), 38) < 0.15
		}})
		// Sloped talud along the outer foot and a pale cornice.
		ox := side.outer
		g.Use(Stone).Fill(Func{Box{V(ox-2, G-0.5, z0), V(ox+2, G+1.6, z1)}, func(p Vec3) bool {
			d := (ox - p.X) * side.sign // Distance outward from the wall.
			return d >= -0.5 && d <= 1.2 && p.Y <= G+1.5-1.0*d
		}})
		g.Paint(White, Block(V(xa-0.1, wallTop-0.4, z0), V(xa+0.4, wallTop, z1)), Block(V(xb-0.4, wallTop-0.4, z0), V(xb+0.1, wallTop, z1)))
		g.Paint(DarkStone, Func{Box{V(xa, wallTop-1.3, z0), V(xb, wallTop-0.5, z1)}, func(p Vec3) bool {
			return c.mod(p.Z, 2.4) < 1.2
		}})
		// Sloped bench along the playing alley.
		fx := side.inner
		g.Use(Stone).Fill(Func{Box{V(fx-2, G-0.5, z0), V(fx+2, G+1.8, z1)}, func(p Vec3) bool {
			d := (p.X - fx) * side.sign // Distance into the alley.
			return d >= -0.5 && d <= 1.8 && p.Y <= G+1.7-0.4*d
		}})
		g.Paint(DarkStone, Func{Box{V(fx-2, G, z0+1), V(fx+2, G+1.8, z1-1)}, func(p Vec3) bool {
			d := (p.X - fx) * side.sign
			return d > 1.2 && c.mod(p.Z, 3.0) < 1.6 && p.Y > G+0.3
		}})
		// Stone ring high on the wall, its hole facing across the court.
		ring := V(fx, wallTop-1.3, (z0+z1)/2)
		g.Use(White).Fill(Func{Box{ring.Sub(V(1, 1.6, 1.6)), ring.Add(V(1, 1.6, 1.6))}, func(p Vec3) bool {
			d := (p.X - fx) * side.sign
			r := math.Hypot(p.Y-ring.Y, p.Z-ring.Z)
			return d >= 0 && d <= 0.8 && r >= 0.7 && r <= 1.55
		}})
	}
	// Playing alley of packed sand.
	g.Paint(Sand, Block(V(34, G-0.5, z0-1), V(42, G+0.1, z1+1)))

	// Temple of the Jaguars on the south end of the east wall, facing the court.
	jx0, jx1, jz0, jz1 := 30.1, 35.5, z0+0.6, z0+7.6
	jy := wallTop
	g.Use(Stone).Fill(Block(V(jx0, jy, jz0), V(jx1, jy+3.0, jz1)))
	g.Fill(Block(V(jx0-0.3, jy+3.0, jz0-0.3), V(jx1+0.3, jy+4.2, jz1+0.3)))
	g.Paint(White, Block(V(jx0-0.3, jy+3.0, jz0-0.3), V(jx1+0.3, jy+3.35, jz1+0.3)))
	g.Paint(DarkStone, Func{Box{V(jx0-0.3, jy+3.5, jz0-0.3), V(jx1+0.3, jy+3.9, jz1+0.3)}, func(p Vec3) bool {
		return c.mod(p.X+p.Z, 1.6) < 0.8
	}})
	g.Carve(Block(V(jx0+1, jy+0.05, jz0+1), V(jx1-1, jy+2.6, jz1-1)))
	g.Carve(Block(V(jx1-1.2, jy+0.05, z0+2.0), V(jx1+0.2, jy+2.4, z0+6.2)))
	g.Paint(Charcoal, Block(V(jx0+0.7, jy-0.5, jz0+0.7), V(jx1-0.7, jy+2.75, jz1-0.7)))
	for _, z := range []float64{z0 + 3.3, z0 + 4.9} {
		g.Use(Sand).Fill(BlockAt(P(jx1-0.5, z), 1.0, 0.8, jy, jy+3.0))
		g.Fill(BlockAt(P(jx1+0.3, z), 1.0, 0.8, jy, jy+0.8))
	}
	// Stair up the outer face to the temple.
	for i := 0; i < 7; i++ {
		y := G + float64(i+1)*5.0/7
		g.Use(White).Fill(Block(V(26.4+float64(i)*0.65, G-0.5, z0+2.2), V(31, y, z0+6.0)))
	}

	// North and south temples closing the ends of the court.
	for _, e := range []struct{ z, dir float64 }{{z1 + 4.2, -1}, {z0 - 4.2, 1}} {
		g.Use(Stone).Fill(BlockAt(P(38, e.z), 9, 5, G-0.8, G+2.2))
		g.Carve(BlockAt(P(38, e.z), 7, 3, G, G+1.4))
		c.rim(g, P(38, e.z), 9, 5, G+1.85, G+2.2)
		for i := 0; i < 4; i++ {
			g.Use(White).Fill(BlockAt(P(38, e.z+e.dir*(2.5+0.6*float64(i))), 3.0, 0.7, G-0.5, G+2.2-0.55*float64(i+1)))
		}
		for i, x := range []float64{34.6, 36.8, 39.2, 41.4} {
			c.pillar(g, P(x, e.z+e.dir*1.2), G+2.2, 2.6, true, i, int(e.z), 36)
		}
		g.Use(Stone).Fill(BlockAt(P(38, e.z-e.dir*1.6), 9, 1.2, G+2.2, G+3.4+1.2*Hash01(int(e.z), 37)))
	}
}

// rim paints a pale moulding around the top edge of a rectangular platform.
func (ChichenItza) rim(g *Grid, c Vec2, w, d, y0, y1 float64) {
	g.Paint(White, Subtract(BlockAt(c, w, d, y0, y1), BlockAt(c, w-1.4, d-1.4, y0, y1)))
}

// caracol is the round observatory on its two platforms, the upper drum half fallen.
func (c ChichenItza) caracol(g *Grid) {
	G := chichenItzaG
	o := P(29, -28)
	y1, y2 := G+1.6, G+3.4
	g.Use(Stone).Fill(BlockAt(o, 13, 11, G-0.8, y1))
	g.Carve(BlockAt(o, 11, 9, G, y1-0.9))
	c.rim(g, o, 13, 11, y1-0.35, y1)
	u := o.Add(P(-0.6, 0))
	g.Use(Stone).Fill(BlockAt(u, 8.5, 8.5, y1, y2))
	g.Carve(BlockAt(u, 6.5, 6.5, y1, y2-0.9))
	c.rim(g, u, 8.5, 8.5, y2-0.35, y2)
	// Stairs on the west (+X) faces.
	for i := 0; i < 3; i++ {
		g.Use(White).Fill(Block(V(o.X+6.5, G-0.5, o.Y-2.2), V(o.X+6.5+0.7*float64(3-i), G+float64(i+1)*1.6/3, o.Y+2.2)))
		g.Fill(Block(V(u.X+4.25, y1-0.5, o.Y-1.6), V(u.X+4.25+0.6*float64(3-i), y1+float64(i+1)*1.8/3, o.Y+1.6)))
	}
	// Lower drum with four doors and a central core; cornice; upper drum with slits.
	d0, d1 := y2+3.4, y2+6.4
	g.Use(Stone).Fill(Cylinder{Center: u, Radius: 3.6, Inner: 2.6, Y0: y2, Y1: d0})
	g.Fill(Cylinder{Center: u, Radius: 1.3, Y0: y2, Y1: d0})
	g.Use(White).Fill(Cylinder{Center: u, Radius: 4.0, Y0: d0, Y1: d0 + 0.5})
	g.Use(Stone).Fill(Cylinder{Center: u, Radius: 2.6, Inner: 1.6, Y0: d0 + 0.5, Y1: d1})
	g.Fill(Frustum{Center: u, R0: 2.6, R1: 1.1, Y0: d1, Y1: d1 + 1.3})
	for _, a := range []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2} {
		q := Polar(u, 3.1, a)
		g.Carve(Func{Box{V(q.X-1.5, y2+0.05, q.Y-1.5), V(q.X+1.5, y2+2.5, q.Y+1.5)}, func(p Vec3) bool {
			d := p.Plan().Sub(u)
			along := d.Dot(P(math.Cos(a), math.Sin(a)))
			across := math.Abs(d.Dot(P(-math.Sin(a), math.Cos(a))))
			return along > 1.9 && across < 0.85
		}})
	}
	for _, a := range []float64{0.3, 1.9, 3.6} {
		q := Polar(u, 2.1, a)
		g.Carve(Func{Box{V(q.X-1, d0+1.4, q.Y-1), V(q.X+1, d0+3.0, q.Y+1)}, func(p Vec3) bool {
			d := p.Plan().Sub(u)
			return d.Length() > 1.3 && math.Abs(d.Dot(P(-math.Sin(a), math.Cos(a)))) < 0.45 && d.Dot(P(math.Cos(a), math.Sin(a))) > 0
		}})
	}
	// The fallen south-east quarter of the upper drum.
	g.Carve(Func{Box{V(u.X-4, d0+1.2, u.Y-4), V(u.X+4, d1+2, u.Y+4)}, func(p Vec3) bool {
		a := p.Plan().Sub(u).Angle()
		k := int(math.Floor(a * 3))
		return a > -2.6 && a < -0.9 && p.Y > d0+1.0+1.6*Hash01(k, 41)
	}})
	g.Paint(Charcoal, Cylinder{Center: u, Radius: 2.95, Inner: 1.25, Y0: y2 - 0.5, Y1: d0 + 0.4})
	g.Paint(Charcoal, Cylinder{Center: u, Radius: 1.85, Y0: d0 + 0.4, Y1: d1})
}

// plazaPlatforms adds the Platform of Venus and the Tzompantli skull rack.
func (c ChichenItza) plazaPlatforms(g *Grid) {
	G := chichenItzaG
	v := P(-15, 38)
	g.Use(Stone).Fill(BlockAt(v, 8, 8, G-0.8, G+2.0))
	g.Carve(BlockAt(v, 6, 6, G, G+1.1))
	c.rim(g, v, 8, 8, G+1.65, G+2.0)
	g.Paint(DarkStone, Func{Box{V(v.X-4, G+0.4, v.Y-4), V(v.X+4, G+1.4, v.Y+4)}, func(p Vec3) bool {
		a := math.Min(math.Abs(p.X-v.X), math.Abs(p.Z-v.Y))
		return a > 1.9 && a < 3.2
	}})
	for _, d := range []Vec2{{0, 1}, {0, -1}, {1, 0}, {-1, 0}} {
		for i := 0; i < 3; i++ {
			q := v.Add(d.Scale(4.4 + 0.6*float64(i)))
			wx, wz := 3.0, 0.7
			if d.X != 0 {
				wx, wz = 0.7, 3.0
			}
			g.Use(White).Fill(BlockAt(q, wx, wz, G-0.5, G+2.0-0.6*float64(i+1)))
		}
	}
	// Tzompantli: a long low platform faced with rows of carved skulls.
	t := P(17, 31)
	g.Use(Stone).Fill(BlockAt(t, 13, 5, G-0.8, G+1.8))
	g.Carve(BlockAt(t, 11, 3, G, G+0.9))
	c.rim(g, t, 13, 5, G+1.45, G+1.8)
	g.Paint(DarkStone, BlockAt(t, 13, 5, G+0.2, G+1.3))
	g.Paint(White, Func{Box{V(t.X-6.5, G+0.2, t.Y-2.5), V(t.X+6.5, G+1.3, t.Y+2.5)}, func(p Vec3) bool {
		return int(math.Floor(p.X/0.8)+math.Floor(p.Z/0.8)+math.Floor((p.Y-G)/0.75))%2 == 0
	}})
	for i := 0; i < 3; i++ {
		g.Use(White).Fill(BlockAt(t.Add(P(0, -2.9-0.6*float64(i))), 3.0, 0.7, G-0.5, G+1.8-0.6*float64(i+1)))
	}
}

// sacbe is the raised white causeway north to the cenote, with a shrine at its end.
func (c ChichenItza) sacbe(g *Grid) {
	G := chichenItzaG
	z0, z1 := 26.0, chichenItzaCenZ-chichenItzaCenR-2.6
	g.Use(White).Fill(Block(V(-3, G-0.8, z0), V(3, G+0.45, z1)))
	g.Paint(Stone, Block(V(-3, G-0.8, z0), V(-2.3, G+0.45, z1)), Block(V(2.3, G-0.8, z0), V(3, G+0.45, z1)))
	sz := z1 + 1.4
	g.Use(Stone).Fill(BlockAt(P(0, sz), 9, 3.6, G-0.8, G+0.8))
	g.Paint(White, BlockAt(P(0, sz), 9, 3.6, G+0.5, G+0.8))
	g.Use(Stone).Fill(Subtract(BlockAt(P(2.6, sz), 3.6, 2.6, G+0.8, G+2.6), BlockAt(P(2.6, sz), 2.0, 1.0, G+0.8, G+2.6), BlockAt(P(2.6, sz-1.0), 1.2, 1.4, G+0.8, G+2.0)))
}

// tree is a shade tree: a wood trunk under a flattened leafy crown.
func (ChichenItza) tree(g *Grid, q Vec2, i int) {
	G := chichenItzaG
	ht := 2.0 + 1.4*Hash01(i, 51)
	g.Use(Wood).Fill(Cylinder{Center: q, Radius: 0.5, Y0: G, Y1: G + ht + 0.5})
	col := Foliage
	if Hash01(i, 52) < 0.35 {
		col = Grass
	}
	r := 1.8 + 1.0*Hash01(i, 53)
	g.Use(col).FillVolume(Ellipsoid{q.At(G + ht + 1.0), V(r, r*0.7, r)})
}

// trees fringes the plaza with shade trees just inside the jungle edge, and dots a few
// across the open lawn.
func (c ChichenItza) trees(g *Grid) {
	const n = 64
	for i := 0; i < n; i++ {
		th := 2*math.Pi*float64(i)/n + 0.06*Hash01(i, 54)
		dir := P(math.Cos(th), math.Sin(th))
		r1 := c.rho(chichenItzaSX+dir.X, chichenItzaSZ+dir.Y)
		d := (c.edge(th) - 0.025 - 0.035*Hash01(i, 55)) / r1
		q := P(chichenItzaSX, chichenItzaSZ).Add(dir.Scale(d))
		if Hash01(i, 56) > 0.6 || c.site(q.X, q.Y, 3) || math.Hypot(q.X, q.Y-chichenItzaCenZ) < chichenItzaCenR+4 {
			continue
		}
		c.tree(g, q, i)
	}
	for i, q := range []Vec2{{-24, 30}, {-30, 24}, {-12, -36}, {8, -38}, {45, -14}, {-8, 48}, {-26, 46}, {-50, 24}, {-30, -24}, {6, 42}} {
		c.tree(g, q, 100+i)
	}
}

func (c ChichenItza) Build(g *Grid) Order {
	c.terrain(g)
	c.sacbe(g)
	c.castillo(g)
	c.warriors(g)
	c.ballCourt(g)
	c.caracol(g)
	c.plazaPlatforms(g)
	c.trees(g)
	return Outward(P(0, 0))
}
