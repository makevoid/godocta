package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// Santorini is an Oia-style village cascading down a volcanic caldera cliff. The cliff
// is an arc around a caldera center far out to the south (+Z): banded lava strata
// (dark stone, brick, terracotta) at the foot, pale pumice under the village. Seven
// stepped terraces carry whitewashed cube and cave houses with flat parapet roofs or
// barrel vaults, dark doors and windows, blue doors and shutters, roof and terrace
// pools, and stairways threading between them. Blue-domed churches with bell gables
// stand among the houses; a large church with a belfry and two windmills crown the rim.
// A zigzag stair drops down the bare lava cliff to a little harbor with tavernas,
// a breakwater and moored caïques; an islet chapel and a gulet sit in the caldera sea,
// a water layer at y = 0.
type Santorini struct{}

func init() { Register(Santorini{}) }

func (Santorini) Name() string  { return "santorini" }
func (Santorini) Title() string { return "Santorini" }

// The caldera rim is an arc around santoriniCaldera. Plan positions use (t, u): t is
// the arc length along the shoreline (positive toward +X) and u the landward distance
// from the shoreline.
var santoriniCaldera = P(0, 150)

const (
	santoriniShore  = 140.0 // Shoreline radius around the caldera center.
	santoriniFoot   = 13.0  // u where the village terraces begin.
	santoriniDepth  = 4.0   // Terrace depth (u).
	santoriniRise   = 3.2   // Terrace riser height.
	santoriniFloor0 = 12.0  // Floor height of terrace 0.
	santoriniRows   = 7     // Terraces below the rim plateau.
	santoriniBack   = 57.0  // u of the back edge.
	santoriniHalfX  = 46.0  // Land half-width in X.
	santoriniSkin   = 1.1   // Ground skin thickness.
)

const (
	santoriniNone = iota
	santoriniRock
	santoriniTerrace
	santoriniPlateau
	santoriniPath
)

// santoriniFrame is a local building frame: a runs along the cliff (+t), b runs into
// the cliff (+u), so the front of a building at b = 0 faces the sea.
type santoriniFrame struct{ o, e, n Vec2 }

func (f santoriniFrame) local(p Vec3) (a, b float64) {
	d := p.Plan().Sub(f.o)
	return d.Dot(f.e), d.Dot(f.n)
}

func (f santoriniFrame) world(a, b float64) Vec2 { return f.o.Add(f.e.Scale(a)).Add(f.n.Scale(b)) }

// solid is the part of a local box accepted by in (nil accepts the whole box).
func (f santoriniFrame) solid(a0, a1, b0, b1, y0, y1 float64, in func(a, b, y float64) bool) Solid {
	box := EmptyBox()
	for _, a := range []float64{a0, a1} {
		for _, b := range []float64{b0, b1} {
			q := f.world(a, b)
			box = box.Union(Box{q.At(y0), q.At(y1)})
		}
	}
	return Func{box, func(p Vec3) bool {
		a, b := f.local(p)
		if a < a0 || a > a1 || b < b0 || b > b1 || p.Y < y0 || p.Y > y1 {
			return false
		}
		return in == nil || in(a, b, p.Y)
	}}
}

func (f santoriniFrame) block(a0, a1, b0, b1, y0, y1 float64) Solid {
	return f.solid(a0, a1, b0, b1, y0, y1, nil)
}

// arch is a round-headed opening through the frame's b axis.
func (f santoriniFrame) arch(ac, hw, sill, spring, b0, b1 float64) Solid {
	return f.solid(ac-hw, ac+hw, b0, b1, sill, sill+spring+hw, func(a, b, y float64) bool {
		r := y - sill - spring
		return r <= 0 || (a-ac)*(a-ac)+r*r <= hw*hw
	})
}

func (Santorini) frame(t, u float64) santoriniFrame {
	sn, cs := math.Sincos(t / santoriniShore)
	n := P(sn, -cs)
	return santoriniFrame{o: santoriniCaldera.Add(n.Scale(santoriniShore + u)), e: P(cs, sn), n: n}
}

func (Santorini) tu(x, z float64) (t, u float64) {
	d := P(x, z).Sub(santoriniCaldera)
	return math.Atan2(d.X, -d.Y) * santoriniShore, d.Length() - santoriniShore
}

func santoriniSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func santoriniNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return ((a+(b-a)*sx)*(1-sz)+(c+(d-c)*sx)*sz)*2 - 1
}

func santoriniFBM(x, z float64, seed int) float64 {
	return 0.65*santoriniNoise(x, z, seed) + 0.35*santoriniNoise(x*2.1+7, z*2.1+3, seed+1)
}

func (Santorini) rowFloor(k int) float64 { return santoriniFloor0 + santoriniRise*float64(k) }
func (Santorini) rowFront(k int) float64 { return santoriniFoot + santoriniDepth*float64(k) }

// rowEnds is the t extent of terrace k; the village widens toward the rim.
func (s Santorini) rowEnds(k int) (float64, float64) {
	half := 19 + 3.0*float64(k)
	limit := santoriniShore * math.Asin(42/(santoriniShore+s.rowFront(k)+santoriniDepth))
	return math.Max(-half-3*Hash01(k, 31), -limit), math.Min(half+3*Hash01(k, 37), limit)
}

// lower is the bare lava cliff's profile below the village.
func (Santorini) lower(u float64) float64 {
	return 12.4 * math.Pow(math.Max(u, 0)/santoriniFoot, 0.55)
}

// The zigzag stair from the harbor to terrace 0, as (t, u) points.
var santoriniStair = []Vec2{P(22, 0.4), P(32, 1.9), P(13, 4.7), P(32, 7.7), P(13, 10.6), P(15, 13.4)}

// stairLeg returns the distance from (t, u) to leg i and the centerline u there.
func (Santorini) stairLeg(i int, t, u float64) (float64, float64) {
	a, b := santoriniStair[i], santoriniStair[i+1]
	ab := b.Sub(a)
	k := math.Max(0, math.Min(1, P(t, u).Sub(a).Dot(ab)/ab.Dot(ab)))
	q := a.Add(ab.Scale(k))
	return P(t, u).Sub(q).Length(), q.Y
}

func (s Santorini) stairY(uc float64) float64 {
	y := math.Round(s.lower(uc)/0.4) * 0.4
	return math.Max(1.6, math.Min(y, santoriniFloor0))
}

// stair returns the distance to the nearest stair leg and its tread height.
func (s Santorini) stair(t, u float64) (float64, float64) {
	best, y := math.Inf(1), 0.0
	for i := 0; i+1 < len(santoriniStair); i++ {
		if d, uc := s.stairLeg(i, t, u); d < best {
			best, y = d, s.stairY(uc)
		}
	}
	return best, y
}

// ground returns the terrain height and kind at a plan point (height -1 off the land).
func (s Santorini) ground(x, z float64) (float64, int) {
	if math.Abs(x) > santoriniHalfX {
		return -1, santoriniNone
	}
	t, u := s.tu(x, z)
	if u < 0 || u > santoriniBack {
		return -1, santoriniNone
	}
	if u < santoriniFoot+1 {
		if d, y := s.stair(t, u); d < 1.05 {
			return y, santoriniPath
		}
	}
	rim := s.rowFront(santoriniRows)
	switch {
	case u < santoriniFoot:
		k := u / santoriniFoot
		h := s.lower(u) + 2.2*santoriniFBM(t*0.11, u*0.3, 5)*math.Min(1, 4*k*(1-k))
		return math.Max(h, 0.3), santoriniRock
	case u < rim:
		k := int((u - santoriniFoot) / santoriniDepth)
		tw, te := s.rowEnds(k)
		if t >= tw && t <= te {
			return s.rowFloor(k), santoriniTerrace
		}
		h := santoriniFloor0 + (u-santoriniFoot)*santoriniRise/santoriniDepth + 1.4*santoriniFBM(t*0.1, u*0.12, 9)
		return h, santoriniRock
	default:
		h := s.rowFloor(santoriniRows) + 1.6*santoriniSmooth((u-rim-2)/6) - 3.5*santoriniSmooth((u-rim-10)/7)
		h += 0.5 * santoriniFBM(t*0.15, u*0.15, 13) * santoriniSmooth((u-rim-3)/3)
		return h, santoriniPlateau
	}
}

// groundBox bounds the land.
func (Santorini) groundBox() Box {
	return Box{V(-santoriniHalfX, -0.3, -50), V(santoriniHalfX, 37, 21)}
}

// rockColor is the color of exposed rock: lava strata below the village, pumice above.
func (Santorini) rockColor(t, y float64) Color {
	wob := 0.9*math.Sin(t*0.07) + 0.5*math.Sin(t*0.19+1.3)
	band := int(math.Floor((y + wob) / 1.6))
	h := Hash01(band, 77)
	if y+wob < 12.5 {
		switch {
		case h < 0.32:
			return DarkStone
		case h < 0.58:
			return Brick
		case h < 0.8:
			return Terracotta
		case h < 0.9:
			return Charcoal
		default:
			return Sand
		}
	}
	switch {
	case h < 0.42:
		return Sand
	case h < 0.62:
		return Stone
	case h < 0.86:
		return Terracotta
	default:
		return Brick
	}
}

func (s Santorini) Build(g *Grid) Order {
	// The caldera sea around the land.
	g.Use(Water).Fill(Func{Box{V(-52, -0.2, -52), V(52, 0.2, 40)}, func(p Vec3) bool {
		dx, dz := math.Max(math.Abs(p.X)-44, 0), math.Max(math.Abs(p.Z+6)-38, 0)
		if dx*dx+dz*dz > 64 {
			return false
		}
		h, _ := s.ground(p.X, p.Z)
		return h < 0
	}})

	s.terrain(g)
	s.harbor(g)
	s.village(g)
	s.rim(g)
	s.islet(g)
	s.gulet(g, P(-14, 27), 0.25)
	s.yacht(g, P(-34, 24), -0.5)
	return Sweep(santoriniCaldera, -math.Pi/2-0.7)
}

// terrain builds the hollow ground skin, colors it, and walls the harbor stair.
func (s Santorini) terrain(g *Grid) {
	in := func(p Vec3) bool {
		h, _ := s.ground(p.X, p.Z)
		return h >= 0 && p.Y >= -0.2 && p.Y <= h
	}
	k := santoriniSkin
	box := s.groundBox()
	g.Use(DarkStone).Fill(Func{box, func(p Vec3) bool {
		if !in(p) {
			return false
		}
		return !(in(p.Add(V(k, 0, 0))) && in(p.Add(V(-k, 0, 0))) && in(p.Add(V(0, 0, k))) &&
			in(p.Add(V(0, 0, -k))) && in(p.Add(V(0, k, 0))))
	}})
	color := func(p Vec3) Color {
		h, kind := s.ground(p.X, p.Z)
		if h < 0 {
			return PaletteSize // Not ground (the sea).
		}
		t, u := s.tu(p.X, p.Z)
		top := p.Y > h-0.9
		switch {
		case kind == santoriniPath && top:
			return Stone
		case kind == santoriniTerrace && top:
			return Stone
		case kind == santoriniPlateau && top:
			if u < s.rowFront(santoriniRows)+3 || math.Abs(math.Mod(t+200, 16)-8) < 1.2 {
				return Stone // Rim promenade and lanes.
			}
			return Sand
		}
		return s.rockColor(t, p.Y)
	}
	for _, c := range []Color{Stone, Sand, Brick, Terracotta, Charcoal} {
		g.Paint(c, Func{box, func(p Vec3) bool { return color(p) == c }})
	}

	// Retaining walls on the downhill side of the stair, capped by white parapets.
	wall := func(p Vec3, parapet bool) bool {
		t, u := s.tu(p.X, p.Z)
		if u < 0 || u > santoriniFoot+1 {
			return false
		}
		if d, _ := s.stair(t, u); d < 1.05 {
			return false
		}
		for i := 0; i+1 < len(santoriniStair); i++ {
			d, uc := s.stairLeg(i, t, u)
			if d >= 1.05 && d <= 1.75 && u < uc {
				h, _ := s.ground(p.X, p.Z)
				y := s.stairY(uc)
				if p.Y <= y+0.9 && p.Y >= math.Min(h, y)-0.6 && (p.Y > y-0.2) == parapet {
					return true
				}
			}
		}
		return false
	}
	stairBox := Box{V(5, 0, -6), V(40, 13, 21)}
	g.Use(Stone).Fill(Func{stairBox, func(p Vec3) bool { return wall(p, false) }})
	g.Use(White).Fill(Func{stairBox, func(p Vec3) bool { return wall(p, true) }})
}

// house builds a whitewashed house in frame f: a ∈ [-w/2, w/2], b ∈ [b0, b1], floor y0,
// roof y1. Doors and windows face the sea; seed selects every variation.
func (s Santorini) house(g *Grid, f santoriniFrame, w, b0, b1, y0, y1 float64, seed int) {
	hw := w / 2
	wall := White
	switch r := Hash01(seed*7919+13, 1); {
	case y1-y0 > 5:
	case r < 0.07:
		wall = Sand
	case r < 0.13:
		wall = Terracotta
	case r < 0.17:
		wall = Stone
	}
	vault := w <= 5.6 && Hash01(seed, 2) < 0.3
	g.With(wall, func() {
		g.Fill(f.block(-hw, hw, b0, b1, y0-0.4, y1))
		if vault {
			hv := math.Min(hw, 1.7)
			g.FillVolume(f.solid(-hw, hw, b0, b1, y1-0.1, y1+hv, func(a, b, y float64) bool {
				return (a/hw)*(a/hw)+((y-y1)/hv)*((y-y1)/hv) <= 1
			}))
		}
	})
	if w > 3 && b1-b0 > 3 {
		g.Carve(f.block(-hw+1, hw-1, b0+1, b1-1, y0+0.5, y1-0.8))
	}
	if !vault {
		g.With(White, func() {
			g.Fill(f.solid(-hw, hw, b0, b1, y1, y1+0.5, func(a, b, y float64) bool {
				return a < -hw+0.6 || a > hw-0.6 || b < b0+0.6
			}))
		})
		if w >= 4.8 && Hash01(seed, 3) < 0.18 {
			g.Paint(Glass, f.block(-hw+1, hw-1, b0+1, b1-1.2, y1-0.45, y1+0.1))
		} else if Hash01(seed, 4) < 0.3 {
			a := (Hash01(seed, 5) - 0.5) * (w - 2)
			g.With(White, func() { g.Fill(f.block(a-0.45, a+0.45, b1-2.2, b1-1.3, y1, y1+1.6)) })
		}
	}
	// Bougainvillea spilling over a front corner.
	if Hash01(seed, 50) < 0.14 {
		side := 1.0
		if Hash01(seed, 51) < 0.5 {
			side = -1
		}
		c := f.world(side*(hw-0.3), b0-0.2).At(y1 - 0.6)
		g.With(Foliage, func() { g.Fill(Ellipsoid{c, V(1.0, 1.4, 1.0)}) })
		g.Paint(Red, Func{Box{c.Sub(V(1.1, 1.5, 1.1)), c.Add(V(1.1, 1.5, 1.1))}, func(p Vec3) bool {
			return Hash01(int(math.Floor(p.X*2)), int(math.Floor(p.Y*2)), int(math.Floor(p.Z*2)), seed) < 0.55
		}})
	}
	// Openings: a door and a window or two on each storey.
	storeys := int(math.Max(1, math.Round((y1-y0)/3.1)))
	for st := 0; st < storeys; st++ {
		fy := y0 + 3.1*float64(st)
		da := (Hash01(seed, 10+st) - 0.5) * math.Max(w-2.6, 0)
		door := Hash01(seed, 20+st)
		switch {
		case vault && st == storeys-1:
			g.Carve(f.arch(da, 0.65, fy, 1.4, b0-0.3, b0+1.1))
		case door < 0.4:
			g.Paint(Navy, f.block(da-0.6, da+0.6, b0-0.3, b0+0.6, fy, fy+2.2))
		case door < 0.55:
			g.Paint(Wood, f.block(da-0.6, da+0.6, b0-0.3, b0+0.6, fy, fy+2.2))
		default:
			g.Carve(f.block(da-0.6, da+0.6, b0-0.3, b0+1.1, fy+0.1, fy+2.2))
		}
		for _, side := range []float64{-1, 1} {
			wa := da + side*1.9
			if math.Abs(wa) > hw-0.9 || Hash01(seed, 30+st, int(side)) < 0.25 {
				continue
			}
			if Hash01(seed, 40+st, int(side)) < 0.35 {
				g.Paint(Navy, f.block(wa-0.5, wa+0.5, b0-0.3, b0+0.6, fy+1.0, fy+2.2))
			} else {
				g.Carve(f.block(wa-0.5, wa+0.5, b0-0.3, b0+1.1, fy+1.0, fy+2.2))
			}
		}
	}
}

// pool is a terrace with an infinity pool and a white cabana behind it.
func (s Santorini) pool(g *Grid, f santoriniFrame, w, b0, b1, y0 float64) {
	hw := w / 2
	g.With(White, func() {
		g.Fill(f.block(-hw, hw, b0, b1, y0-0.4, y0+1.0))
		g.Fill(f.block(-hw, hw, b1-1.6, b1, y0+1.0, y0+3.2))
	})
	g.Carve(f.block(-hw+0.8, hw-0.8, b0+0.8, b1-2.2, y0+0.5, y0+1.2))
	g.With(Glass, func() { g.Fill(f.block(-hw+0.8, hw-0.8, b0+0.8, b1-2.2, y0+0.2, y0+0.6)) })
	g.Carve(f.block(-hw+1, hw-1, b1-1.7, b1-0.8, y0+1.0, y0+2.6))
}

// stairs climbs from a terrace floor to the next one across the riser.
func (s Santorini) stairs(g *Grid, f santoriniFrame, hw, y0 float64) {
	g.With(Stone, func() {
		for i := 0; i < 8; i++ {
			g.Fill(f.block(-hw, hw, 0.45*float64(i)+0.6, santoriniDepth+0.8, y0-0.8, y0+0.4*float64(i+1)))
		}
	})
}

// church is a white chapel with a blue dome, a bell gable and gold crosses.
func (s Santorini) church(g *Grid, f santoriniFrame, w, b0, b1, y0, wallH, dome float64, domeColor Color) {
	hw := w / 2
	top := y0 + wallH
	g.With(White, func() {
		g.Fill(f.block(-hw, hw, b0, b1, y0-0.4, top))
		g.Fill(f.solid(-hw-0.3, hw+0.3, b0-0.3, b1, top, top+0.5, nil))
	})
	g.Carve(f.block(-hw+1, hw-1, b0+1, b1-1, y0+0.5, top-0.8))
	g.Carve(f.arch(0, 0.75, y0, 1.6, b0-0.4, b0+1.1))
	for _, side := range []float64{-1, 1} {
		g.Carve(f.arch(side*(hw-1.4), 0.45, y0+1.4, 0.9, b0-0.4, b0+1.1))
	}
	// Drum and dome.
	dc := f.world(0, (b0+b1)/2+0.4)
	drumTop := top + 0.5 + dome*0.6
	g.With(White, func() { g.FillVolume(Cylinder{dc, dome - 0.2, 0, top, drumTop}) })
	g.With(domeColor, func() {
		g.FillVolume(Intersect(Sphere(dc.At(drumTop), dome), Block(V(dc.X-dome, drumTop, dc.Y-dome), V(dc.X+dome, drumTop+dome, dc.Y+dome))))
	})
	s.cross(g, dc.At(drumTop+dome-0.2), f)
	// Bell gable over the façade: two arches with gold bells, curved top, cross.
	gt := top + 3.0
	g.With(White, func() {
		g.Fill(f.solid(-hw+0.6, hw-0.6, b0, b0+1.0, top, gt, func(a, b, y float64) bool {
			k := a / (hw - 0.6)
			return y <= gt-0.9+0.9*(1-k*k)
		}))
	})
	for _, side := range []float64{-1, 1} {
		ac := side * (hw - 0.6) / 2
		g.Carve(f.arch(ac, 0.6, top+0.6, 0.6, b0-0.4, b0+1.4))
		g.With(Gold, func() { g.Fill(f.block(ac-0.35, ac+0.35, b0+0.2, b0+0.8, top+1.0, top+1.6)) })
	}
	c := f.world(0, b0+0.5)
	s.cross(g, c.At(gt-0.1), f)
}

// cross is a small gold cross standing at base, facing the frame's front.
func (s Santorini) cross(g *Grid, base Vec3, f santoriniFrame) {
	g.With(Gold, func() {
		g.Fill(Capsule{base, base.Add(V(0, 1.8, 0)), 0.42})
		e := f.e.Scale(0.6)
		g.Fill(Capsule{base.Add(V(-e.X, 1.2, -e.Y)), base.Add(V(e.X, 1.2, e.Y)), 0.42})
	})
}

// village fills the terraces with houses, pools, churches and stairways.
func (s Santorini) village(g *Grid) {
	type churchAt struct {
		row   int
		t     float64
		color Color
	}
	churches := []churchAt{{1, -5, Navy}, {3, 12, Navy}, {4, -19, Navy}, {5, 27, Terracotta}, {2, 23, Navy}, {6, -3, Navy}, {0, 10, Navy}}
	stairs := map[int][]float64{}
	for k := 0; k < santoriniRows; k++ {
		front, y0 := s.rowFront(k), s.rowFloor(k)
		tw, te := s.rowEnds(k)
		// Terrace parapet at the front edge, open where the stairs below arrive.
		g.With(White, func() {
			g.Fill(Func{Box{V(-santoriniHalfX, y0-0.2, -40), V(santoriniHalfX, y0+1.0, 8)}, func(p Vec3) bool {
				t, u := s.tu(p.X, p.Z)
				if u < front || u > front+0.7 || t < tw || t > te {
					return false
				}
				for _, st := range stairs[k-1] {
					if math.Abs(t-st) < 1.3 {
						return false
					}
				}
				return true
			}})
		})
		var reserved [][2]float64
		for _, c := range churches {
			if c.row == k {
				reserved = append(reserved, [2]float64{c.t - 3.6, c.t + 3.6})
				f := s.frame(c.t, front)
				s.church(g, f, 6.0, 0.6, 5.8, y0, 4.2, 2.4, c.color)
			}
		}
		t := tw + 0.4
		for slot := 0; t < te-2.5; slot++ {
			seed := k*1000 + slot
			skip := false
			for _, r := range reserved {
				if t+1 > r[0] && t < r[1] {
					t, skip = r[1]+0.4, true
				}
			}
			if skip {
				continue
			}
			if Hash01(seed, 7) < 0.15 && t > tw+3 && t < te-4 && k < santoriniRows-1 {
				stairs[k] = append(stairs[k], t+1.1)
				s.stairs(g, s.frame(t+1.1, front), 1.1, y0)
				t += 2.2
				continue
			}
			w := 3.6 + 2.8*Hash01(seed, 8)
			for _, r := range reserved {
				if t+w > r[0] && t < r[0] {
					w = r[0] - t
				}
			}
			w = math.Min(w, te-t)
			if w < 2.8 {
				t += w + 0.4
				continue
			}
			f := s.frame(t+w/2, front)
			back := santoriniDepth + 1.2
			switch r := Hash01(seed, 9); {
			case r < 0.08:
				s.pool(g, f, w, 1.0, back, y0)
			case r < 0.28 && k < santoriniRows-1:
				s.house(g, f, w, 1.2+0.4*Hash01(seed, 11), back, y0, y0+6.2+0.4*Hash01(seed, 12), seed)
			default:
				s.house(g, f, w, 1.0+0.6*Hash01(seed, 11), back, y0, y0+2.9+0.7*Hash01(seed, 12), seed)
			}
			if Hash01(seed, 13) < 0.45 {
				t += w
			} else {
				t += w + 1.0
			}
		}
		// Outlying cave houses dug into the bare slope beyond the terrace ends.
		for _, side := range []float64{-1, 1} {
			edge := te
			if side < 0 {
				edge = tw
			}
			for j := 0; j < 2; j++ {
				seed := 7000 + k*10 + j*2 + int(side+1)/2
				if Hash01(seed, 1) < 0.3 {
					continue
				}
				tc := edge + side*(4+7*float64(j)+3*Hash01(seed, 2))
				uc := front + 1 + 2*Hash01(seed, 3)
				f := s.frame(tc, uc)
				q := f.world(0, 0)
				if math.Abs(q.X) > santoriniHalfX-4 {
					continue
				}
				h, _ := s.ground(q.X, q.Y)
				y := math.Round(h/0.4) * 0.4
				s.house(g, f, 3.6+1.2*Hash01(seed, 4), 0, 4.5, y, y+3.0+0.4*Hash01(seed, 5), seed)
			}
		}
	}
}

// rim builds the plateau: houses, the great church and belfry, and the windmills.
func (s Santorini) rim(g *Grid) {
	rim := s.rowFront(santoriniRows)
	floor := func(t, u float64) float64 {
		f := s.frame(t, u)
		q := f.world(0, 0)
		h, _ := s.ground(q.X, q.Y)
		return math.Round(h/0.4) * 0.4
	}
	// Rim parapet.
	g.With(White, func() {
		y0 := s.rowFloor(santoriniRows)
		g.Fill(Func{Box{V(-santoriniHalfX, y0-0.2, -45), V(santoriniHalfX, y0+1.0, 0)}, func(p Vec3) bool {
			_, u := s.tu(p.X, p.Z)
			return u >= rim && u <= rim+0.7
		}})
	})
	// Great church and belfry.
	gt := 2.0
	gf := s.frame(gt, rim+3)
	gy := floor(gt, rim+3)
	s.church(g, gf, 7.6, 0, 9.5, gy, 5.2, 3.0, Navy)
	s.belfry(g, s.frame(gt+7.2, rim+3.5), gy)
	// Houses in two rows behind the promenade.
	for row := 0; row < 2; row++ {
		front := rim + 3.2 + 6.5*float64(row)
		tw, te := -40.0+3*float64(row), 40.0-2*float64(row)
		t := tw
		for slot := 0; t < te; slot++ {
			seed := 9000 + row*100 + slot
			w := 4.0 + 3.0*Hash01(seed, 8)
			c := t + w/2
			skip := (row == 0 && c > gt-8 && c < gt+11) || (c > -32 && c < -22) || (row == 1 && c > -19 && c < -11)
			if skip || Hash01(seed, 6) < 0.12 {
				t += w + 1
				continue
			}
			f := s.frame(c, front)
			fl := floor(c, front)
			low := fl
			for _, q := range []Vec2{f.world(-w/2, 0), f.world(w/2, 0), f.world(-w/2, 5), f.world(w/2, 5)} {
				h, _ := s.ground(q.X, q.Y)
				if h < 0 {
					low = -1 // Off the land.
					break
				}
				fl, low = math.Max(fl, math.Round(h/0.4)*0.4), math.Min(low, h)
			}
			if low < 0 {
				t += w + 1
				continue
			}
			p := f.world(0, 0)
			if math.Abs(p.X) > santoriniHalfX-3 {
				t += w + 1
				continue
			}
			if fl-low > 0.8 {
				g.With(Stone, func() { g.Fill(f.block(-w/2, w/2, 0, 5, low-0.6, fl)) })
			}
			height := 3.0 + 0.8*Hash01(seed, 12)
			if row == 0 && Hash01(seed, 9) < 0.2 {
				height += 3.1
			}
			if Hash01(seed, 10) < 0.1 {
				s.pool(g, f, w, 0, 5, fl)
			} else {
				s.house(g, f, w, 0, 5.0, fl, fl+height, seed)
			}
			t += w + 0.4 + 1.2*Hash01(seed, 14)
		}
	}
	s.windmill(g, s.frame(-27, rim+7), floor(-27, rim+7), 1)
	s.windmill(g, s.frame(-15, rim+12.5), floor(-15, rim+12.5), 2)
}

// belfry is a tiered white bell tower with open arches, bells and a small blue dome.
func (s Santorini) belfry(g *Grid, f santoriniFrame, y0 float64) {
	const hw = 1.8
	tiers := []float64{y0 - 0.4, y0 + 5.6, y0 + 9.2, y0 + 12.0}
	g.With(White, func() {
		g.Fill(f.block(-hw, hw, -hw, hw, tiers[0], tiers[1]))
		g.Fill(f.block(-hw-0.3, hw+0.3, -hw-0.3, hw+0.3, tiers[1]-0.5, tiers[1]))
		g.Fill(f.block(-hw+0.2, hw-0.2, -hw+0.2, hw-0.2, tiers[1], tiers[2]))
		g.Fill(f.block(-hw, hw, -hw, hw, tiers[2]-0.5, tiers[2]))
		g.Fill(f.block(-hw+0.6, hw-0.6, -hw+0.6, hw-0.6, tiers[2], tiers[3]))
	})
	// Open belfry: arches through both axes.
	g.Carve(f.arch(0, 0.85, tiers[1]+0.4, 1.4, -hw-1, hw+1))
	g.Carve(f.solid(-hw-1, hw+1, -0.85, 0.85, tiers[1]+0.4, tiers[1]+2.7, func(a, b, y float64) bool {
		r := y - tiers[1] - 1.8
		return r <= 0 || b*b+r*r <= 0.85*0.85
	}))
	g.Carve(f.arch(0, 0.5, tiers[2]+0.4, 0.9, -hw-1, hw+1))
	g.With(Gold, func() {
		g.Fill(f.block(-0.5, 0.5, -0.5, 0.5, tiers[1]+1.4, tiers[1]+2.4))
		g.Fill(f.block(-0.4, 0.4, -0.4, 0.4, tiers[2]+0.6, tiers[2]+1.2))
	})
	c := f.world(0, 0)
	g.With(Navy, func() {
		g.FillVolume(Intersect(Sphere(c.At(tiers[3]), 1.4), Block(V(c.X-2, tiers[3], c.Y-2), V(c.X+2, tiers[3]+2, c.Y+2))))
	})
	s.cross(g, c.At(tiers[3]+1.2), f)
}

// windmill is a white tower with a thatched cone and six sails facing the sea.
func (s Santorini) windmill(g *Grid, f santoriniFrame, y0 float64, seed int) {
	c := f.world(0, 0)
	const r, h = 2.6, 7.5
	g.With(White, func() {
		g.FillVolume(Frustum{c, r + 0.3, r - 0.1, y0 - 0.6, y0 + h})
	})
	g.Carve(Cylinder{c, r - 1.1, 0, y0 + 0.5, y0 + h - 0.6})
	g.Carve(f.arch(0, 0.7, y0, 1.5, -r-0.5, -r+1.4))
	g.Paint(Navy, f.block(-0.5, 0.5, -r-0.5, -r+1, y0+4.2, y0+5.2))
	g.Carve(f.block(-r-0.5, -r+1.2, -0.5, 0.5, y0+3.6, y0+4.8))
	g.With(Wood, func() { g.FillVolume(Cone(c, r+0.5, y0+h, 3.2)) })
	// Hub, axle and sails in the plane b = -r - 1.4.
	hub := f.world(0, -r-1.4).At(y0 + h - 0.6)
	g.With(Wood, func() {
		g.Fill(Capsule{f.world(0, -r+0.5).At(y0 + h - 0.6), hub, 0.55})
	})
	spin := Hash01(seed, 3) * math.Pi / 3
	g.With(Wood, func() {
		for i := 0; i < 6; i++ {
			a := spin + float64(i)*math.Pi/3
			tip := f.world(6.2*math.Cos(a), -r-1.4).At(hub.Y + 6.2*math.Sin(a))
			g.Fill(Capsule{hub, tip, 0.45})
		}
	})
	g.With(White, func() {
		g.Fill(f.solid(-6.5, 6.5, -r-1.9, -r-0.9, hub.Y-6.5, hub.Y+6.5, func(a, b, y float64) bool {
			rr := math.Hypot(a, y-hub.Y)
			if rr < 1.6 || rr > 5.9 {
				return false
			}
			ang := math.Atan2(y-hub.Y, a) - spin
			ang = math.Mod(math.Mod(ang, math.Pi/3)+math.Pi/3, math.Pi/3)
			return ang > 0.12 && ang < 0.12+0.5*(rr-1.0)/5
		}))
	})
}

// caique is a small wooden fishing boat floating at c, bow toward angle.
func (s Santorini) caique(g *Grid, c Vec2, angle, length float64, band Color) {
	sn, cs := math.Sincos(angle)
	f := santoriniFrame{o: c, e: P(-sn, cs), n: P(cs, sn)} // b along the bow.
	beam := length * 0.36
	half := func(b, y float64) float64 {
		k := b / (length / 2)
		return beam / 2 * math.Sqrt(math.Max(1-k*k*k*k, 0)) * (0.7 + 0.3*math.Min(y/1.2, 1))
	}
	hull := f.solid(-beam, beam, -length/2, length/2, -0.2, 1.5, func(a, b, y float64) bool {
		return math.Abs(a) <= half(b, y) && y <= 1.0+0.35*math.Abs(b)/(length/2)
	})
	g.With(White, func() { g.FillVolume(hull) })
	g.Paint(band, Intersect(hull, f.solid(-beam, beam, -length/2, length/2, 0.6, 1.0, nil)))
	g.Carve(f.solid(-beam, beam, -length/2+1, length/2-1, 0.7, 2, func(a, b, y float64) bool {
		return math.Abs(a) <= half(b, y)-0.8
	}))
	g.With(Wood, func() {
		g.Fill(f.solid(-beam, beam, -length/2+1, length/2-1, 0.2, 0.6, func(a, b, y float64) bool {
			return math.Abs(a) <= half(b, y)-0.6
		}))
	})
}

// harbor builds the Ammoudi-style quay, tavernas, breakwater and boats.
func (s Santorini) harbor(g *Grid) {
	f := s.frame(23, 0)
	g.With(Stone, func() {
		g.Fill(f.block(-9, 9, -5.5, 2.5, -0.2, 1.6))
		g.Fill(f.block(-9, -5, -9, -5.5, -0.2, 1.2)) // Slipway pier.
	})
	g.Paint(DarkStone, f.block(-9, 9, -5.5, 2.5, -0.2, 0.6), f.block(-9, -5, -9, -5.5, -0.2, 0.6))
	// Tavernas against the cliff with awnings on the quay.
	for i, a := range []float64{-5.6, -1.2, 3.4} {
		s.house(g, s.frame(23+a, 0), 4.2, 0.4, 4.8, 1.6, 4.8+0.4*float64(i%2), 500+i)
		col := []Color{Navy, Terracotta, Navy}[i]
		fa := s.frame(23+a, 0)
		g.With(col, func() { g.Fill(fa.block(-1.9, 1.9, -3.0, 0.4, 4.0, 4.4)) })
		g.With(Wood, func() {
			for _, x := range []float64{-1.6, 1.6} {
				g.Fill(fa.block(x-0.4, x+0.4, -3.0, -2.2, 1.6, 4.0))
			}
		})
	}
	// Breakwater and harbor light.
	pts := []Vec2{P(8.5, -3.5), P(12, -10), P(8, -15.5)}
	g.With(DarkStone, func() {
		for i := 0; i+1 < len(pts); i++ {
			a, b := f.world(pts[i].X, pts[i].Y), f.world(pts[i+1].X, pts[i+1].Y)
			g.FillVolume(Intersect(Capsule{a.At(0), b.At(0), 2.0}, Block(V(-100, -0.2, -100), V(100, 1.5, 100))))
		}
	})
	lc := f.world(8, -15.5)
	g.With(White, func() { g.FillVolume(Cylinder{lc, 1.0, 0, 1.0, 4.4}) })
	g.With(Red, func() { g.FillVolume(Cylinder{lc, 1.1, 0, 4.4, 5.2}) })
	// Moored caïques.
	for i, b := range []struct {
		a, b, ang, l float64
		c            Color
	}{{-7, -8.5, 0, 5.5, Navy}, {-2.5, -9.5, 0, 6, Red}, {2, -9, 0.15, 5, Navy}, {6, -8.5, -0.1, 5.5, Wood}, {-14, -6, 0.6, 5, Red}} {
		_ = i
		c := f.world(b.a, b.b)
		ang := math.Atan2(f.n.Y, f.n.X) + math.Pi + b.ang
		s.caique(g, c, ang, b.l, b.c)
	}
}

// islet is a rocky islet with a tiny blue-domed chapel and steps from the water.
func (s Santorini) islet(g *Grid) {
	f := s.frame(43, -11)
	c := f.world(0, 0)
	in := func(p Vec3) bool {
		d := p.Plan().Sub(c)
		a := d.Angle()
		rr := 5.5 * (1 - 0.45*p.Y/6) * (0.85 + 0.25*santoriniNoise(math.Cos(a)*2+3, math.Sin(a)*2+p.Y*0.2, 41))
		return p.Y >= -0.2 && p.Y <= 5 && d.Length() <= rr
	}
	box := Box{V(c.X-7, -0.2, c.Y-7), V(c.X+7, 6, c.Y+7)}
	g.With(DarkStone, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			return in(p) && !(in(p.Add(V(1, 0, 0))) && in(p.Add(V(-1, 0, 0))) && in(p.Add(V(0, 0, 1))) &&
				in(p.Add(V(0, 0, -1))) && in(p.Add(V(0, 1, 0))))
		}})
	})
	g.Paint(Brick, Func{box, func(p Vec3) bool { return p.Y > 1.5 && p.Y < 3 }})
	g.With(White, func() { g.Fill(f.block(-3, 3, -3, 3, 4.4, 5.0)) })
	cf := s.frame(43, -10)
	s.church(g, cf, 4.2, -1.8, 2.4, 5.0, 3.2, 1.5, Navy)
}

// gulet is a two-masted wooden excursion schooner at anchor in the caldera.
func (s Santorini) gulet(g *Grid, c Vec2, angle float64) {
	sn, cs := math.Sincos(angle)
	f := santoriniFrame{o: c, e: P(-sn, cs), n: P(cs, sn)}
	const length, beam = 18.0, 5.2
	half := func(b, y float64) float64 {
		k := (b + 1) / (length / 2)
		if b > 0 {
			k = math.Pow(math.Abs(k), 1.6)
		} else {
			k = math.Pow(math.Abs(k), 3)
		}
		return beam / 2 * math.Sqrt(math.Max(1-k, 0)) * (0.65 + 0.35*math.Min(y/1.8, 1))
	}
	hull := f.solid(-beam, beam, -length/2, length/2+1, -0.2, 3, func(a, b, y float64) bool {
		return math.Abs(a) <= half(b, y) && y <= 2.0+0.06*b+0.004*b*b
	})
	g.With(Wood, func() { g.FillVolume(hull) })
	g.Paint(White, Intersect(hull, f.solid(-beam, beam, -length/2, length/2+1, -0.2, 0.8, nil)))
	g.Carve(f.solid(-beam, beam, -length/2+1, length/2-1, 1.6, 3, func(a, b, y float64) bool {
		return math.Abs(a) <= half(b, y)-0.8
	}))
	g.With(Wood, func() {
		g.Fill(f.solid(-beam, beam, -length/2+1, length/2-1, 1.0, 1.6, func(a, b, y float64) bool {
			return math.Abs(a) <= half(b, y)-0.6
		}))
		g.Fill(f.block(-1.6, 1.6, -6, -1.5, 1.6, 3.4)) // Deckhouse.
	})
	g.With(White, func() { g.Fill(f.block(-1.8, 1.8, -6.2, -1.3, 3.4, 3.8)) })
	for _, m := range []struct{ b, h float64 }{{2.5, 13}, {-4.5, 10}} {
		base := f.world(0, m.b).At(1.6)
		g.With(Wood, func() { g.Fill(Capsule{base, base.Add(V(0, m.h, 0)), 0.45}) })
		// Furled sails along a boom.
		boom := f.world(0, m.b-4.5).At(4.2)
		g.With(White, func() { g.Fill(Capsule{base.Add(V(0, 2.6, 0)), boom, 0.6}) })
	}
	bow := f.world(0, length/2+0.5).At(2.4)
	g.With(Wood, func() { g.Fill(Capsule{bow, f.world(0, length/2+4).At(3.2), 0.4}) })
}

// yacht is a small sloop under sail.
func (s Santorini) yacht(g *Grid, c Vec2, angle float64) {
	sn, cs := math.Sincos(angle)
	f := santoriniFrame{o: c, e: P(-sn, cs), n: P(cs, sn)}
	const length, beam = 9.0, 3.0
	half := func(b float64) float64 {
		k := (b + 0.5) / (length / 2)
		if b > 0 {
			k = k * k
		} else {
			k = k * k * k * k
		}
		return beam / 2 * math.Sqrt(math.Max(1-math.Abs(k), 0))
	}
	hull := f.solid(-beam, beam, -length/2, length/2+0.5, -0.2, 1.4, func(a, b, y float64) bool {
		return math.Abs(a) <= half(b)*(0.7+0.3*math.Min(y, 1)) && y <= 1.0+0.05*b
	})
	g.With(White, func() { g.FillVolume(hull) })
	g.Paint(Navy, Intersect(hull, f.solid(-beam, beam, -length/2, length/2+1, 0.5, 0.9, nil)))
	g.With(Wood, func() { g.Fill(f.block(-0.9, 0.9, -3.4, -1.2, 1.0, 1.6)) })
	mast := f.world(0, 0.8)
	g.With(Stone, func() { g.Fill(Capsule{mast.At(1), mast.At(11), 0.4}) })
	// Mainsail aft of the mast and a jib forward, as thin triangles in the a = 0 plane.
	g.With(White, func() {
		g.Fill(f.solid(-0.5, 0.5, -3.6, 4.4, 1.8, 10.6, func(a, b, y float64) bool {
			if b < 0.2 {
				return b >= 0.2-4.0*(10.6-y)/8.8
			}
			return b > 1.2 && b <= 1.2+3.2*(10.0-y)/8.2
		}))
	})
}
