package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// MachuPicchu is the Inca citadel on its saddle ridge high above the Urubamba. The
// ridge runs along X: the steep sugarloaf of Huayna Picchu rises at its northern (-X)
// end, the slope of Machu Picchu mountain climbs to the south (+X), and the river
// curves round the base of both in a deep, forested valley. On the saddle the
// citadel's granite houses stand roofless around the sunken main plaza, with
// trapezoidal doorways and windows, steep gable ends and a few rebuilt thatched
// roofs. The upper (west, -Z) sector holds the stepped Intihuatana hill, the Sacred
// Plaza with the Main Temple and Temple of the Three Windows, and the curved Temple
// of the Sun on its boulder above the Royal Tomb; the lower (east, +Z) sector holds
// rows of houses and the Temple of the Condor. A dry moat and staircases divide the
// town from the agricultural sector, whose grass terraces climb to the thatched
// Guardhouse. Llamas graze on the terraces; a switchback road climbs the east slope.
type MachuPicchu struct{}

func init() { Register(MachuPicchu{}) }

func (MachuPicchu) Name() string  { return "machu_picchu" }
func (MachuPicchu) Title() string { return "Machu Picchu" }

// The citadel is authored in design units on a finer grid and scaled up by
// machuPicchuK: a lattice of span s/K scaled by K is exactly the lattice of span s.
const (
	machuPicchuK     = 1.2
	machuPicchuBase  = 20.0                           // Citadel ground level, design units.
	machuPicchuCrest = machuPicchuBase * machuPicchuK // Saddle level, world units.
	machuPicchuStep  = 1.6                            // Agricultural terrace riser.
)

// Terrain kinds.
const (
	machuPicchuForest = iota
	machuPicchuTerrace
	machuPicchuCitadel
	machuPicchuStair
	machuPicchuCliff
)

func machuPicchuSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func machuPicchuClamp(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

// machuPicchuNoise is smooth value noise in [0, 1).
func machuPicchuNoise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

func machuPicchuFBM(x, z float64, seed int) float64 {
	return (machuPicchuNoise(x, z, seed)*4 + machuPicchuNoise(x*2.1+5, z*2.1+3, seed+1)*2 +
		machuPicchuNoise(x*4.3+1, z*4.3+7, seed+2)) / 7
}

func machuPicchuSegDist(p, a, b Vec2) float64 {
	ab := b.Sub(a)
	t := machuPicchuClamp(p.Sub(a).Dot(ab)/ab.Dot(ab), 0, 1)
	return p.Sub(a.Add(ab.Scale(t))).Length()
}

func machuPicchuPolyDist(p Vec2, line []Vec2) float64 {
	best := math.Inf(1)
	for i := 0; i+1 < len(line); i++ {
		best = math.Min(best, machuPicchuSegDist(p, line[i], line[i+1]))
	}
	return best
}

// machuPicchuRiver is the Urubamba's centerline: a horseshoe round the mountain,
// open toward +X where the ridge climbs out of the site.
var machuPicchuRiver = func() []Vec2 {
	var line []Vec2
	for x := 62.0; x > -36; x -= 2 {
		line = append(line, P(x, 33+1.5*math.Sin(x/11)))
	}
	for t := 0.0; t <= math.Pi+1e-9; t += math.Pi / 40 {
		line = append(line, P(-36-29*math.Sin(t), 33*math.Cos(t)))
	}
	for x := -34.0; x <= 62; x += 2 {
		line = append(line, P(x, -33-1.5*math.Sin(x/13+1)))
	}
	return line
}()

// machuPicchuMap is a sampled heightfield with bilinear lookup.
type machuPicchuMap struct {
	x0, z0, step float64
	nx, nz       int
	h            []float64
}

func (m *machuPicchuMap) at(x, z float64) float64 {
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

func (m *machuPicchuMap) low(x, z float64) float64 {
	lo := m.at(x, z)
	for k := 0; k < 8; k++ {
		a := float64(k) * math.Pi / 4
		lo = math.Min(lo, m.at(x+1.2*math.Cos(a), z+1.2*math.Sin(a)))
	}
	return lo
}

// rim is the normalized site radius; above 1 is off the plot.
func (MachuPicchu) rim(x, z float64) float64 {
	ex, ez := math.Abs(x+6.5)/62, math.Abs(z)/39
	return math.Pow(math.Pow(ex, 4)+math.Pow(ez, 4), 0.25)
}

// crest is the ridge-top height along the spine.
func (MachuPicchu) crest(x float64) float64 {
	switch {
	case x < -31:
		return machuPicchuCrest - 7*machuPicchuSmooth((-31-x)/8)
	case x > 20:
		return machuPicchuCrest + 6*machuPicchuSmooth((x-20)/10)
	}
	return machuPicchuCrest
}

// halfWidth is the half-width of the ridge's flat top along the spine.
func (MachuPicchu) halfWidth(x float64) float64 {
	return 3 + 11.4*machuPicchuSmooth((x+42)/12) - 7.4*machuPicchuSmooth((x-18)/12)
}

// huayna is the Huayna Picchu sugarloaf with a lower shoulder to its east.
func (MachuPicchu) huayna(x, z float64) float64 {
	dx, dz := (x+46)/1.05, z+2
	r := math.Hypot(dx, dz) * (1 + 0.2*(machuPicchuFBM(x/5, z/5, 7)-0.5))
	h := 1.5 + 49*math.Exp(-math.Pow(r/11.5, 2.8))
	s := math.Hypot((x+54)/1.2, z-9) * (1 + 0.2*(machuPicchuFBM(x/4, z/4, 9)-0.5))
	return math.Max(h, 1.5+21*math.Exp(-math.Pow(s/7, 2.2)))
}

// citadel is the designed ground level of the town: upper and lower sectors, the
// sunken main plaza, the stepped Intihuatana hill and the dry moat.
func (MachuPicchu) citadel(x, z float64) (float64, bool) {
	ex, ez := (x+7.75)/18.75, (z+0.25)/12.75
	if ex*ex*ex*ex*ex*ex+ez*ez*ez*ez*ez*ez > 1 {
		return 0, false
	}
	b := machuPicchuBase
	l := b
	if z < -5 {
		l = b + 1.5
	}
	if x < -21.5 && z > -5 {
		l = b - 0.8
	}
	// Intihuatana hill.
	dx, dz := math.Abs(x+17)/1.2, math.Abs(z+8.5)
	rr := math.Pow(dx*dx*dx*dx+dz*dz*dz*dz, 0.25)
	switch {
	case rr < 2.1:
		l = b + 6
	case rr < 3.3:
		l = b + 4.5
	case rr < 4.6:
		l = b + 3
	}
	// Main plaza with an intermediate step.
	if x > -19.5 && x < 5.5 && z > -4.5 && z < 5 {
		l = b - 1.2
	}
	if x > -17 && x < 3 && z > -2.8 && z < 3.4 {
		l = b - 2.4
	}
	// Dry moat between the town and the agricultural sector.
	if x > 9.4 {
		l = b - 1.6
	}
	return l, true
}

// height is the final ground height and kind of a column; ok is false off the plot.
func (m MachuPicchu) height(x, z float64) (float64, int, bool) {
	e := m.rim(x, z)
	if e > 1 {
		return 0, machuPicchuForest, false
	}
	xs := machuPicchuClamp(x, -40, 30)
	r := math.Hypot(x-xs, z)
	w := m.halfWidth(xs)
	// Terraced shoulders under the flat top, then steep cloud-forest slopes.
	t := math.Max(0, r-w)
	// Spurs and ravines vary where and how steeply the flanks fall away.
	n1, n2 := machuPicchuFBM(x/9, z/9, 13), machuPicchuFBM(x/11+7, z/11, 15)
	drop := 0.8*t + (1.9+1.4*(n1-0.5))*math.Log1p(math.Exp(t-7-6*(n2-0.5)))
	// A concave foot eases into the valley floor.
	nat := 1.5 + 3.5*math.Log1p(math.Exp((m.crest(xs)-drop-1.5)/3.5))
	nat += 3 * (machuPicchuFBM(x/8, z/8, 3) - 0.5) * machuPicchuSmooth((t-7)/5) * machuPicchuSmooth((nat-1.5)/3)
	hp := m.huayna(x, z)
	h, kind := math.Max(nat, hp), machuPicchuForest
	if hp <= nat {
		tw := 7.0
		if z < 0 && x < 2 {
			tw = 5
		}
		if l, ok := m.citadel(x/machuPicchuK, z/machuPicchuK); ok {
			h, kind = l*machuPicchuK, machuPicchuCitadel
		} else if xs > -29 && r < w+tw && x < 33 {
			step := machuPicchuStep
			kind = machuPicchuTerrace
			// Staircases down both flanks and up to the Guardhouse.
			if math.Abs(x-14.6) < 0.9 || (math.Abs(z+1.2) < 0.9 && x > 13 && x < 29) {
				step, kind = 0.8, machuPicchuStair
			}
			h = machuPicchuCrest + math.Floor((h-machuPicchuCrest)/step)*step
		}
	} else if hp > 39 && x > -48+3*(z+2)/8 {
		h = math.Min(39+math.Floor((hp-39)/1.5)*1.5, 48.5)
		kind = machuPicchuTerrace
	} else {
		kind = machuPicchuCliff
	}
	rd := machuPicchuPolyDist(P(x, z), machuPicchuRiver)
	if !InsidePolygon(machuPicchuRiver, P(x, z)) {
		if bank := 1.5 + 2.5*machuPicchuSmooth((rd-4)/5) + 1.5*(machuPicchuFBM(x/6, z/6, 5)-0.5); bank > h {
			h, kind = bank, machuPicchuForest
		}
	}
	if c := 0.6 + math.Max(0, rd-2.6)*3; c < h {
		h, kind = c, machuPicchuForest
	}
	if c := 0.6 + 110*(1-e); c < h {
		h, kind = c, machuPicchuForest
	}
	return h, kind, true
}

// ground is the ground height at a point.
func (m MachuPicchu) ground(x, z float64) float64 {
	h, _, _ := m.height(x, z)
	return h
}

func (m MachuPicchu) Build(g *Grid) Order {
	mp := &machuPicchuMap{x0: -76, z0: -45, step: 0.4}
	mp.nx, mp.nz = int(140/mp.step)+2, int(90/mp.step)+2
	mp.h = make([]float64, mp.nx*mp.nz)
	for j := 0; j < mp.nz; j++ {
		for i := 0; i < mp.nx; i++ {
			mp.h[j*mp.nx+i] = m.ground(mp.x0+float64(i)*mp.step, mp.z0+float64(j)*mp.step)
		}
	}
	m.terrain(g, mp)
	m.forest(g, mp)
	d := NewGridWithSpan(g.Span() / machuPicchuK)
	m.upperSector(d)
	m.lowerSector(d)
	m.northEnd(d)
	// Lichen-darkened stones.
	for _, c := range d.Cells() {
		if col, _ := d.ColorOf(c); col == Stone && Hash01(int(c.X), int(c.Y), int(c.Z), 61) < 0.14 {
			d.Remove(c)
			d.Use(DarkStone).Add(c)
		}
	}
	g.Merge(d)
	m.agriculture(g)
	m.summit(g)
	m.llamas(g)
	m.weather(g)
	return Outward(P(-4, 0))
}

// terrain fills the heightfield as a hollow skin, column by column, coloring
// terrace treads and retaining walls, forest, cliffs, road, banks and river.
func (m MachuPicchu) terrain(g *Grid, mp *machuPicchuMap) {
	span := g.Span()
	half := span / 2
	cx0, cx1 := int(math.Ceil(-76/half)), int(math.Floor(64/half))
	cz0, cz1 := int(math.Ceil(-45/half)), int(math.Floor(45/half))
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
			low := math.Min(h, mp.low(x, z))
			slope := (mp.at(x, z) - mp.low(x, z)) / 1.2
			rd := machuPicchuPolyDist(P(x, z), machuPicchuRiver)
			y0 := math.Max(0, low-0.8)
			cy0 := int(math.Ceil(y0 / half))
			if (cy0^cx)&1 != 0 {
				cy0++
			}
			for cy := cy0; float64(cy)*half <= h; cy += 2 {
				y := float64(cy) * half
				var c Color
				switch kind {
				case machuPicchuCitadel, machuPicchuTerrace:
					c = Stone
					if y >= h-0.6 {
						c = Grass
					} else if Hash01(cx, cy, cz, 21) < 0.18 {
						c = DarkStone
					}
				case machuPicchuStair:
					c = Stone
				default:
					switch {
					case rd < 2.7:
						c = Water
					case slope > 2.4+1.6*machuPicchuNoise(x/7, z/7, 35) && (kind != machuPicchuCliff || machuPicchuNoise(x/2.5+y/8, z/2.5-y/8, 37) > 0.5):
						c = DarkStone
						if machuPicchuNoise(x/2.5+y/3, z/2.5-y/4, 31) > 0.62 {
							c = Stone
						}
					case h < 2.4:
						c = Grass
					default:
						c = Foliage
						if machuPicchuFBM(x/6, z/6, 33) > 0.62 {
							c = Grass
						}
					}
				}
				g.Use(c)
				g.Add(Cell{int32(cx), int32(cy), int32(cz)})
			}
		}
	}
}

// forest scatters cloud-forest canopy over the natural slopes.
func (m MachuPicchu) forest(g *Grid, mp *machuPicchuMap) {
	const step = 2.4
	for i := -32; i <= 28; i++ {
		for j := -20; j <= 20; j++ {
			x := float64(i)*step + (Hash01(i, j, 41)-0.5)*1.8
			z := float64(j)*step + (Hash01(i, j, 42)-0.5)*1.8
			h, kind, ok := m.height(x, z)
			if !ok || (kind != machuPicchuForest && kind != machuPicchuCliff) || h < 2.6 {
				continue
			}
			if (mp.at(x, z)-mp.low(x, z))/1.2 > 3 || machuPicchuPolyDist(P(x, z), machuPicchuRiver) < 4.5 {
				continue
			}
			if Hash01(i, j, 43) < 0.25 {
				continue
			}
			r := 1.0 + 0.7*Hash01(i, j, 44)
			c := Foliage
			if Hash01(i, j, 45) < 0.2 {
				c = Grass
			}
			g.Use(c).FillVolume(Ellipsoid{Center: V(x, h+0.2, z), Radii: V(r, r*0.85, r)})
		}
	}
}

// footprint returns the floor level at a rectangle's center and the lowest ground
// under it.
func (m MachuPicchu) footprint(k, x0, z0, x1, z1 float64) (floor, low float64) {
	floor = m.ground((x0+x1)/2*k, (z0+z1)/2*k) / k
	low = floor
	for i := 0; i <= 4; i++ {
		for j := 0; j <= 4; j++ {
			low = math.Min(low, m.ground((x0+(x1-x0)*float64(i)/4)*k, (z0+(z1-z0)*float64(j)/4)*k)/k)
		}
	}
	return floor, low - 0.3
}

// trapezoid is an Inca opening through a wall: wider at the sill, narrowing upward.
// at is the sill center; through is the axis the opening passes along.
func (MachuPicchu) trapezoid(at Vec3, bottom, top, height, depth float64, through Axis) Solid {
	b := Box{at.Sub(V(bottom/2, 0, bottom/2)), at.Add(V(bottom/2, height, bottom/2))}
	if through == AxisX {
		b.Min.X, b.Max.X = at.X-depth/2, at.X+depth/2
	} else {
		b.Min.Z, b.Max.Z = at.Z-depth/2, at.Z+depth/2
	}
	return Func{b, func(p Vec3) bool {
		across := math.Abs(p.X - at.X)
		if through == AxisX {
			across = math.Abs(p.Z - at.Z)
		}
		t := (p.Y - at.Y) / height
		return across <= (bottom+(top-bottom)*t)/2
	}}
}

// machuPicchuHouse describes one rectangular granite house.
type machuPicchuHouse struct {
	x0, z0, x1, z1 float64
	ridge          Axis    // Gable ridge direction; gable ends are perpendicular.
	door           int     // Wall with the doorway: 0 -Z, 1 +Z, 2 -X, 3 +X, -1 none.
	open           int     // Wall left out entirely (wayronas), -1 none.
	wall           float64 // Wall height above the floor.
	roof           bool    // Rebuilt thatched roof.
	gables         bool
}

// house builds a roofless (or thatched) Inca house with gable ends, a trapezoidal
// doorway and windows.
// k is the grid's scale: machuPicchuK on the citadel's design grid, 1 in the world.
func (m MachuPicchu) house(g *Grid, k float64, hs machuPicchuHouse) {
	floor, low := m.footprint(k, hs.x0, hs.z0, hs.x1, hs.z1)
	if hs.wall == 0 {
		hs.wall = 2.3 + 0.9*Hash01(int(hs.x0*10), int(hs.z0*10), 81)
	}
	top := floor + hs.wall
	const t = 1.0
	shell := Subtract(Block(V(hs.x0, low, hs.z0), V(hs.x1, top, hs.z1)),
		Block(V(hs.x0+t, floor, hs.z0+t), V(hs.x1-t, top+1, hs.z1-t)))
	g.Use(Stone).Fill(shell)
	// Fill the floor under the house so it stands on a level platform.
	g.Use(Grass).Fill(Block(V(hs.x0+t, low, hs.z0+t), V(hs.x1-t, floor, hs.z1-t)))
	if hs.open >= 0 {
		var cut Cuboid
		switch hs.open {
		case 0:
			cut = Block(V(hs.x0+t, floor, hs.z0-0.1), V(hs.x1-t, top+5, hs.z0+t+0.1))
		case 1:
			cut = Block(V(hs.x0+t, floor, hs.z1-t-0.1), V(hs.x1-t, top+5, hs.z1+0.1))
		case 2:
			cut = Block(V(hs.x0-0.1, floor, hs.z0+t), V(hs.x0+t+0.1, top+5, hs.z1-t))
		case 3:
			cut = Block(V(hs.x1-t-0.1, floor, hs.z0+t), V(hs.x1+0.1, top+5, hs.z1-t))
		}
		g.Carve(cut)
	}
	cx, cz := (hs.x0+hs.x1)/2, (hs.z0+hs.z1)/2
	across := (hs.z1 - hs.z0) / 2
	if hs.ridge == AxisZ {
		across = (hs.x1 - hs.x0) / 2
	}
	rise := across * 1.15
	if hs.gables {
		g.Use(Stone).Fill(Func{Box{V(hs.x0, top, hs.z0), V(hs.x1, top+rise, hs.z1)}, func(p Vec3) bool {
			a := math.Abs(p.Z - cz)
			end := p.X < hs.x0+t || p.X > hs.x1-t
			if hs.ridge == AxisZ {
				a = math.Abs(p.X - cx)
				end = p.Z < hs.z0+t || p.Z > hs.z1-t
			}
			return end && a <= across*(1-(p.Y-top)/rise)
		}})
	}
	// Doorway.
	switch hs.door {
	case 0:
		g.Carve(m.trapezoid(V(cx, floor, hs.z0+t/2), 1.8, 1.2, 2.3, 1.6, AxisZ))
	case 1:
		g.Carve(m.trapezoid(V(cx, floor, hs.z1-t/2), 1.8, 1.2, 2.3, 1.6, AxisZ))
	case 2:
		g.Carve(m.trapezoid(V(hs.x0+t/2, floor, cz), 1.8, 1.2, 2.3, 1.6, AxisX))
	case 3:
		g.Carve(m.trapezoid(V(hs.x1-t/2, floor, cz), 1.8, 1.2, 2.3, 1.6, AxisX))
	}
	// Windows in the long walls away from the door.
	long := hs.x1 - hs.x0
	if hs.ridge == AxisZ {
		long = hs.z1 - hs.z0
	}
	if long >= 5.5 {
		for _, s := range []float64{-1, 1} {
			off := long/2 - 1.6
			if hs.ridge == AxisX {
				for _, w := range []struct {
					z    float64
					side int
				}{{hs.z0 + t/2, 0}, {hs.z1 - t/2, 1}} {
					if w.side != hs.door && w.side != hs.open {
						g.Carve(m.trapezoid(V(cx+s*off, floor+1.1, w.z), 1.2, 0.9, 1.2, 1.6, AxisZ))
					}
				}
			} else {
				for _, w := range []struct {
					x    float64
					side int
				}{{hs.x0 + t/2, 2}, {hs.x1 - t/2, 3}} {
					if w.side != hs.door && w.side != hs.open {
						g.Carve(m.trapezoid(V(w.x, floor+1.1, cz+s*off), 1.2, 0.9, 1.2, 1.6, AxisX))
					}
				}
			}
		}
	}
	if hs.roof {
		m.thatch(g, hs.x0, hs.z0, hs.x1, hs.z1, top, hs.ridge)
	}
}

// thatch is a steep hollow thatched roof with a dark ridge.
func (MachuPicchu) thatch(g *Grid, x0, z0, x1, z1, top float64, ridge Axis) {
	const ov, pitch = 0.7, 1.2
	across := (z1-z0)/2 + ov
	if ridge == AxisZ {
		across = (x1-x0)/2 + ov
	}
	peak := top - 0.3 + across*pitch
	cx, cz := (x0+x1)/2, (z0+z1)/2
	b := Box{V(x0-ov, top-0.4, z0-ov), V(x1+ov, peak, z1+ov)}
	if ridge == AxisX {
		b.Min.X, b.Max.X = x0-0.4, x1+0.4
	} else {
		b.Min.Z, b.Max.Z = z0-0.4, z1+0.4
	}
	roof := Func{b, func(p Vec3) bool {
		a := math.Abs(p.Z - cz)
		if ridge == AxisZ {
			a = math.Abs(p.X - cx)
		}
		s := peak - pitch*a
		return p.Y <= s && p.Y >= s-1.3
	}}
	g.Use(Sand).Fill(roof)
	g.Paint(Wood, Func{b, func(p Vec3) bool { return p.Y > peak-0.9 }})
}

// upperSector builds the western quarter: Intihuatana, Sacred Plaza, Temple of the
// Sun with the Royal Tomb, and the royal residence.
func (m MachuPicchu) upperSector(g *Grid) {
	b := machuPicchuBase
	// Intihuatana: stepped hill retaining walls are terrain; the hitching post stone.
	g.Use(Stone).Fill(Block(V(-18.4, b+6, -9.4), V(-15.6, b+6.8, -7.6)))
	g.Use(DarkStone).Fill(Block(V(-17.5, b+6.8, -9), V(-16.5, b+8.6, -8)))
	// Hill staircase down to the plaza.
	for k := 0; k < 8; k++ {
		y := b - 1.2 + float64(k)*0.9
		g.Use(Stone).Fill(Block(V(-14.2+float64(k)*0.25, b-1.2, -4.6+float64(k)*-0.55), V(-12.6, y+0.9, -4.1+float64(k)*-0.55)))
	}
	// Sacred Plaza: Main Temple (three walls open to the plaza) and the Temple of the
	// Three Windows facing the main plaza.
	m.house(g, machuPicchuK, machuPicchuHouse{x0: -11, z0: -13, x1: -4.5, z1: -8.5, ridge: AxisX, door: -1, open: 1, wall: 3.6, gables: true})
	m.house(g, machuPicchuK, machuPicchuHouse{x0: -10.5, z0: -7.2, x1: -5, z1: -4.6, ridge: AxisX, door: -1, open: 0, wall: 3.4, gables: true})
	floor := b + 1.5
	for _, x := range []float64{-9.2, -7.75, -6.3} {
		g.Carve(m.trapezoid(V(x, floor+1.0, -5.1), 1.1, 0.8, 1.5, 1.6, AxisZ))
	}
	// Priest's house beside the Sacred Plaza.
	m.house(g, machuPicchuK, machuPicchuHouse{x0: -14.6, z0: -13.2, x1: -11.6, z1: -8.5, ridge: AxisZ, door: 3, open: -1, gables: true})
	// Temple of the Sun: curved tower on its boulder above the Royal Tomb.
	c := P(0, -8)
	rockTop := b + 4
	g.Use(Stone).Fill(
		Intersect(Cylinder{Center: c, Radius: 3.7, Inner: 2.6, Y0: b + 1, Y1: rockTop + 4.2}, Block(V(-5, b, c.Y-0.6), V(5, rockTop+5, c.Y+4))),
		Block(V(c.X-3.7, b+1, c.Y-1.6), V(c.X+3.7, rockTop+4.2, c.Y-0.6)),
	)
	for _, a := range []float64{0.75, 2.4} {
		dir := P(math.Cos(a), math.Sin(a))
		at := c.Add(dir.Scale(3.15))
		w := Rotated{m.trapezoid(V(0, rockTop+1.6, 0), 1.2, 0.9, 1.4, 1.8, AxisZ), P(0, 0), -(a - math.Pi/2)}
		g.Carve(Translate(w, V(at.X, 0, at.Y)))
	}
	g.Carve(m.trapezoid(V(c.X, rockTop, c.Y-1.1), 1.8, 1.2, 2.4, 1.6, AxisZ))
	g.Use(DarkStone).FillVolume(Ellipsoid{Center: V(c.X, b+0.3, c.Y+1.0), Radii: V(4.4, rockTop-b-0.3, 3.4)})
	// Royal Tomb cave under the boulder, opening onto the plaza.
	g.Carve(m.trapezoid(V(c.X, b-1.2, c.Y+3.9), 2.2, 1.4, 2.0, 3.0, AxisZ))
	// Enclosure wall west of the tower.
	g.Use(Stone).Fill(Block(V(c.X+3.7, b+1.5, c.Y-1.6), V(c.X+4.7, b+4, c.Y+2.5)))
	// Royal residence and the stair of the fountains.
	m.house(g, machuPicchuK, machuPicchuHouse{x0: 5, z0: -12.8, x1: 9.2, z1: -8.6, ridge: AxisZ, door: 3, open: -1, gables: true, roof: true})
	m.house(g, machuPicchuK, machuPicchuHouse{x0: 5, z0: -7.8, x1: 9.2, z1: -5.2, ridge: AxisX, door: 1, open: -1, gables: true})
	m.house(g, machuPicchuK, machuPicchuHouse{x0: -3.6, z0: -13.2, x1: 2.6, z1: -10.6, ridge: AxisX, door: -1, open: -1, gables: true})
	for k := 0; k < 4; k++ {
		y := b - 1.2 + float64(k)*0.9
		g.Use(Stone).Fill(Block(V(3.8, b-1.2, -4.8-float64(k)*0.5), V(4.8, y+0.9, -4.3-float64(k)*0.5)))
	}
}

// lowerSector builds the eastern quarter: rows of houses and the Temple of the Condor.
func (m MachuPicchu) lowerSector(g *Grid) {
	b := machuPicchuBase
	x := -20.5
	widths := []float64{5.2, 6.4, 5.0, 6.0, 5.6}
	for i, w := range widths {
		// Front row facing the plaza, back row on the outer edge.
		m.house(g, machuPicchuK, machuPicchuHouse{x0: x, z0: 5.8, x1: x + w, z1: 8.6, ridge: AxisX, door: 0, open: -1,
			gables: true, roof: i == 1})
		bw := w - 0.6
		m.house(g, machuPicchuK, machuPicchuHouse{x0: x + 0.3, z0: 9.6, x1: x + 0.3 + bw, z1: 12.4, ridge: AxisX, door: 0, open: -1,
			gables: i%2 == 0, roof: i == 3})
		x += w + 1.0
	}
	// Plaza staircase into the lower sector.
	for k := 0; k < 3; k++ {
		g.Use(Stone).Fill(Block(V(-14+float64(k)*0, b-2.4, 3.4+float64(k)*0.5), V(-12.4, b-2.4+float64(k+1)*0.8, 3.9+float64(k)*0.5)))
	}
	// Temple of the Condor: dark rock wings behind a three-walled court.
	c := P(4.6, 9.2)
	m.house(g, machuPicchuK, machuPicchuHouse{x0: c.X - 3.3, z0: c.Y - 3.2, x1: c.X + 3.3, z1: c.Y + 3.2, ridge: AxisZ, door: -1, open: 0, wall: 2.2})
	for _, s := range []float64{-1, 1} {
		wing := Func{Box{V(c.X-3, b-1, c.Y-1), V(c.X+3, b+4, c.Y+2.4)}, func(p Vec3) bool {
			dx := (p.X - c.X) * s
			if dx < 0.2 {
				return false
			}
			top := b + 0.6 + dx*0.95
			return p.Y <= top && p.Y >= b-0.5 && p.Z > c.Y-0.2-0.1*dx && p.Z < c.Y+1.6-0.25*dx
		}}
		g.Use(DarkStone).Fill(wing)
	}
	g.Use(Stone).Fill(Block(V(c.X-0.8, b, c.Y-2.4), V(c.X+0.8, b+0.5, c.Y-1.2)))
}

// northEnd builds the Sacred Rock between two thatched wayronas, the quarry and the
// gate toward Huayna Picchu.
func (m MachuPicchu) northEnd(g *Grid) {
	b := machuPicchuBase
	base := b - 0.8
	g.Use(DarkStone).Fill(Func{Box{V(-24.8, base, -3.2), V(-23.2, base+4.5, 4.2)}, func(p Vec3) bool {
		top := base + 2.6 + 1.6*math.Exp(-math.Pow((p.Z-1.6)/1.6, 2)) + 0.8*math.Exp(-math.Pow((p.Z+1.4)/1.2, 2)) - 0.3*math.Abs(p.Z-0.5)
		return p.Y <= top
	}})
	m.house(g, machuPicchuK, machuPicchuHouse{x0: -26.6, z0: -6.8, x1: -21.6, z1: -3.6, ridge: AxisX, door: -1, open: 1, gables: true, roof: true})
	m.house(g, machuPicchuK, machuPicchuHouse{x0: -26.6, z0: 4.6, x1: -21.6, z1: 7.8, ridge: AxisX, door: -1, open: 0, gables: true, roof: true})
	// Quarry boulders.
	for i := 0; i < 14; i++ {
		x := -24.5 + 4*Hash01(i, 51)
		z := -12 + 5*Hash01(i, 52)
		r := 0.7 + 0.6*Hash01(i, 53)
		c := DarkStone
		if Hash01(i, 54) < 0.5 {
			c = Stone
		}
		h := m.ground(x*machuPicchuK, z*machuPicchuK) / machuPicchuK
		g.Use(c).FillVolume(Ellipsoid{Center: V(x, h+r*0.4, z), Radii: V(r*1.2, r, r)})
	}
}

// agriculture adds storehouses on the terraces and the Guardhouse at the top.
func (m MachuPicchu) agriculture(g *Grid) {
	for i, x := range []float64{17, 21.2, 25.4} {
		m.house(g, 1, machuPicchuHouse{x0: x, z0: -14.6, x1: x + 3.4, z1: -11.8, ridge: AxisX, door: 1, open: -1,
			wall: 2.4, gables: true, roof: i != 1})
	}
	m.house(g, 1, machuPicchuHouse{x0: 27, z0: -6.8, x1: 31.6, z1: -3, ridge: AxisX, door: 1, open: 2, gables: true, roof: true})
	m.house(g, 1, machuPicchuHouse{x0: 18, z0: 11.5, x1: 22, z1: 14.2, ridge: AxisX, door: 0, open: -1, wall: 2.4, gables: true})
}

// summit adds terraces' huts on top of Huayna Picchu.
func (m MachuPicchu) summit(g *Grid) {
	m.house(g, 1, machuPicchuHouse{x0: -49.5, z0: -4, x1: -46, z1: -1.4, ridge: AxisX, door: 3, open: -1, wall: 2.2, gables: true})
	m.house(g, 1, machuPicchuHouse{x0: -45.8, z0: 0, x1: -42.8, z1: 2.6, ridge: AxisZ, door: 3, open: -1, wall: 2, gables: true})
}

// llamas puts grazing llamas on the terraces and the plaza.
func (m MachuPicchu) llamas(g *Grid) {
	spots := []struct {
		x, z, a float64
		c       Color
	}{
		{19, 8, 0.3, White}, {23, 10, 2.5, Wood}, {21.5, -7, 1.2, White}, {26.5, 5, 4, White},
		{-9.6, 1, 0.9, White}, {-13.8, 1.9, 3.5, Wood}, {16.5, 18, 5.2, White}, {28, -9, 1.9, Wood},
		{9, 20, 0.6, Wood}, {-2, -19, 2.2, White},
	}
	const s = 1.25
	for _, sp := range spots {
		h := m.ground(sp.x, sp.z)
		body := Union(
			Block(V(-1.1, 1.0, -0.5).Scale(s), V(1.1, 1.9, 0.5).Scale(s)),
			Block(V(0.5, 0, -0.45).Scale(s), V(1.1, 1.1, 0.45).Scale(s)),
			Block(V(-1.1, 0, -0.45).Scale(s), V(-0.5, 1.1, 0.45).Scale(s)),
			Block(V(0.6, 1.8, -0.4).Scale(s), V(1.3, 3.0, 0.4).Scale(s)),
			Block(V(0.8, 2.8, -0.4).Scale(s), V(1.9, 3.5, 0.4).Scale(s)),
		)
		g.Use(sp.c).Fill(Translate(RotateY(body, P(0, 0), sp.a), V(sp.x, h, sp.z)))
	}
}

// weather adds wisps of cloud clinging to Huayna Picchu.
func (m MachuPicchu) weather(g *Grid) {
	for _, c := range []struct {
		at   Vec3
		size float64
	}{{V(-55, 27, 4), 1}, {V(-31, 38, -17), 0.8}} {
		for k := 0; k < 6; k++ {
			o := V((Hash01(k, 71)-0.5)*9, (Hash01(k, 72)-0.5)*1.5, (Hash01(k, 73)-0.5)*5).Scale(c.size)
			r := (1.6 + 1.2*Hash01(k, 74)) * c.size
			g.Use(White).FillVolume(Ellipsoid{Center: c.at.Add(o), Radii: V(r*1.6, r*0.7, r)})
		}
	}
}
