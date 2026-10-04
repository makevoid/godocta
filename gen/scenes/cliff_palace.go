package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// CliffPalace is the Ancestral Puebloan cliff dwelling at Mesa Verde. A long, shallow
// alcove is cut into a banded sandstone canyon wall facing +Z; its arched overhang roofs
// a village of stacked masonry rooms with T-shaped doorways, small windows and vigas,
// the four-storey Square Tower, two round towers and a plaza of sunken round kivas, some
// with ladders. The ledge drops through a short cliff band onto a juniper talus that runs
// down to the canyon floor and its dry wash; the mesa top above carries piñon-juniper
// woodland. The mesa ends and back step down in eroded ledges to a low rim.
type CliffPalace struct{}

func init() { Register(CliffPalace{}) }

func (CliffPalace) Name() string  { return "cliff_palace" }
func (CliffPalace) Title() string { return "Cliff Palace" }

const (
	cliffPalaceX0, cliffPalaceX1 = -38.0, 38.0
	cliffPalaceZ0, cliffPalaceZ1 = -22.0, 22.0
	cliffPalaceFloor             = 1.5  // Canyon floor.
	cliffPalaceLedge             = 15.0 // Alcove floor.
	cliffPalaceMesa              = 35.0
	cliffPalaceSkin              = 1.2
	cliffPalaceA                 = 27.5 // Alcove half-width.
	cliffPalaceD                 = 17.0 // Alcove depth behind the mouth.
	cliffPalaceH                 = 13.5 // Alcove height at the mouth.
	cliffPalaceMouth             = 2.0  // Plan Z of the alcove mouth.
	cliffPalaceStorey            = 2.8
	cliffPalaceWall              = 1.1
	cliffPalaceScale             = 1.2 // World units per model unit.
)

// cliffPalaceKiva is a sunken round ceremonial room.
type cliffPalaceKiva struct {
	x, z, r float64
	ladder  bool
}

var cliffPalaceKivas = []cliffPalaceKiva{
	{-24.5, 6.2, 1.9, false}, {-18.5, 5.8, 2.1, true}, {-12.5, 6.4, 2.0, false},
	{-6.5, 5.8, 2.2, true}, {0, 6.4, 2.0, false}, {6.2, 5.8, 2.1, true},
	{12.5, 6.4, 2.1, false}, {18.5, 5.8, 2.0, true}, {24.2, 6.0, 1.8, false},
	{-18.1, -2.0, 1.7, true}, {1.9, -2.0, 1.7, true}, {13.9, -2.0, 1.7, false},
}

// cliffPalaceSquareTower is the plan footprint of the four-storey Square Tower.
var cliffPalaceSquareTower = [4]float64{4.0, -7.6, 7.8, -4.0}

// cliffPalaceRound are the round towers: x, z, radius, storeys.
var cliffPalaceRound = [][4]float64{{-10.1, -2.0, 1.8, 3}, {21.9, -2.0, 1.7, 2}}

func (CliffPalace) smooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func (CliffPalace) noise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

func (s CliffPalace) fbm(x, z float64, seed int) float64 {
	return (s.noise(x, z, seed)*4 + s.noise(x*2.1+5, z*2.1+3, seed+1)*2 + s.noise(x*4.3+1, z*4.3+7, seed+2)) / 7
}

// steps is an eroded terrace profile: risers of the given height every bench units.
func (CliffPalace) steps(d, bench, riser, jitter float64) float64 {
	if d <= 0 {
		return 0
	}
	return riser * (1 + math.Floor(d/bench+jitter))
}

// faceZ is the plan Z of the canyon wall's face; it bows gently forward at the ends.
func (s CliffPalace) faceZ(x float64) float64 {
	return 1 + 0.004*x*x + 1.2*(s.noise(x/4, 0.5, 3)-0.5)
}

// ledgeZ is the front edge of the ledge under and around the alcove.
func (s CliffPalace) ledgeZ(x float64) float64 {
	w := 8.5 - 5.0*s.smooth((math.Abs(x)-25)/6)
	return s.faceZ(x) + w + 1.0*(s.noise(x/3, 2.5, 4)-0.5)
}

// washZ is the center of the dry wash winding along the canyon floor.
func (CliffPalace) washZ(x float64) float64 { return 18.6 + 2.2*math.Sin(x/8.5+0.6) }

// floorY is the alcove floor, rising in two terraces toward the back.
func (CliffPalace) floorY(x, z float64) float64 {
	y := cliffPalaceLedge
	if z < -3.9 {
		y++
	}
	if z < -7.7 {
		y++
	}
	return y
}

// planR is the alcove's normalized plan radius: below 1 lies under the overhang.
func (CliffPalace) planR(x, z float64) float64 {
	u := math.Abs(x) / cliffPalaceA
	v := math.Max(0, cliffPalaceMouth-z) / cliffPalaceD
	return math.Sqrt(math.Pow(u, 4) + math.Pow(v, 2.2))
}

// roofY is the underside of the overhang.
func (s CliffPalace) roofY(x, z float64) float64 {
	r := s.planR(x, z)
	if r >= 1 {
		return math.Inf(-1)
	}
	return cliffPalaceLedge + cliffPalaceH*math.Sqrt(1-r*r) + 1.4*(s.noise(x/3, z/3, 5)-0.5)
}

// cavity is the open air of the alcove.
func (s CliffPalace) cavity(p Vec3) bool {
	if p.Y <= s.floorY(p.X, p.Z) || p.Y > cliffPalaceLedge+cliffPalaceH+1 {
		return false
	}
	return p.Y < s.roofY(p.X, p.Z)
}

// height is the terrain surface, ignoring the alcove.
func (s CliffPalace) height(x, z float64) float64 {
	// Outer slopes step down to the rim on the ends and back.
	de := math.Min(math.Min(x-cliffPalaceX0, cliffPalaceX1-x), 1.8*(z-cliffPalaceZ0))
	de += 3*(s.fbm(x/7, z/7, 8)-0.5) + 3*(s.fbm(x/15, z/15, 15)-0.5)
	jt := s.noise(x/3, z/3, 10)
	outer := cliffPalaceFloor + s.steps(math.Min(de-1.2, 5), 2.0, 2.0, jt) + s.steps(de-6.5, 0.8+0.6*s.noise(x/6, z/6, 9), 6.5, jt)
	var h float64
	fz, lz := s.faceZ(x), s.ledgeZ(x)
	switch {
	case z < fz:
		h = cliffPalaceMesa + 3*(s.fbm(x/10, z/10, 1)-0.5) - 0.12*math.Max(0, -z-6)
		// The rim rounds off over the last unit.
		h -= 1.5 * s.smooth((z-fz+1.5)/1.5)
	case z < lz:
		h = cliffPalaceLedge
	default:
		t := z - lz
		h = cliffPalaceLedge - 3.2 - 1.05*t + 1.6*(s.fbm(x/4, z/4, 6)-0.5)
		if t < 0.6 {
			h = cliffPalaceLedge - 0.4 // Ledge lip.
		}
		// Dry wash.
		dw := math.Abs(z - s.washZ(x))
		wash := cliffPalaceFloor - 0.8*(1-s.smooth((dw-1.2)/1.2))
		h = math.Max(h, wash)
	}
	return math.Min(h, outer)
}

// cliffPalaceField is the sampled heightfield with bilinear lookup.
type cliffPalaceField struct {
	x0, z0, step float64
	nx, nz       int
	h            []float64
}

func (m *cliffPalaceField) at(x, z float64) float64 {
	fx, fz := (x-m.x0)/m.step, (z-m.z0)/m.step
	if fx < 0 || fz < 0 || fx >= float64(m.nx-1) || fz >= float64(m.nz-1) {
		return cliffPalaceFloor
	}
	i, j := int(fx), int(fz)
	tx, tz := fx-float64(i), fz-float64(j)
	k := j*m.nx + i
	a := m.h[k] + (m.h[k+1]-m.h[k])*tx
	b := m.h[k+m.nx] + (m.h[k+m.nx+1]-m.h[k+m.nx])*tx
	return a + (b-a)*tz
}

// cliffPalaceDirs are the 14 lattice neighbor directions, normalized.
var cliffPalaceDirs = func() []Vec3 {
	var d []Vec3
	for _, n := range Neighbors {
		v := V(float64(n.X), float64(n.Y), float64(n.Z))
		d = append(d, v.Scale(1/v.Length()))
	}
	return d
}()

func (s CliffPalace) Build(g *Grid) Order {
	mp := &cliffPalaceField{x0: cliffPalaceX0 - 2, z0: cliffPalaceZ0 - 2, step: 0.25}
	mp.nx, mp.nz = int((cliffPalaceX1-cliffPalaceX0+4)/mp.step)+2, int((cliffPalaceZ1-cliffPalaceZ0+4)/mp.step)+2
	mp.h = make([]float64, mp.nx*mp.nz)
	for j := 0; j < mp.nz; j++ {
		for i := 0; i < mp.nx; i++ {
			mp.h[j*mp.nx+i] = s.height(mp.x0+float64(i)*mp.step, mp.z0+float64(j)*mp.step)
		}
	}
	// The recipe is authored at model scale and enlarged by building it on a finer
	// lattice, then copying the cells: lattice coordinates scale world positions.
	t := NewGridWithSpan(g.Span() / cliffPalaceScale)
	s.terrain(t, mp)
	var interiors []Box
	s.village(t, &interiors)
	s.vegetation(t, mp)
	s.darkenInteriors(t, interiors)
	g.Merge(t)
	return Outward(P(0, -4*cliffPalaceScale))
}

// kivaFloor is the plaza level around a kiva.
func (s CliffPalace) kivaFloor(k cliffPalaceKiva) float64 { return s.floorY(k.x, k.z) }

// terrain fills the rock as a hollow skin and colors it.
func (s CliffPalace) terrain(g *Grid, mp *cliffPalaceField) {
	solid := func(p Vec3) bool {
		if p.X < cliffPalaceX0 || p.X > cliffPalaceX1 || p.Z < cliffPalaceZ0 || p.Z > cliffPalaceZ1 {
			return false
		}
		if p.Y < 0 {
			return true
		}
		if p.Y > mp.at(p.X, p.Z) {
			return false
		}
		if p.Y > cliffPalaceLedge-0.5 && s.cavity(p) {
			return false
		}
		for _, k := range cliffPalaceKivas {
			dx, dz := p.X-k.x, p.Z-k.z
			if dx*dx+dz*dz <= k.r*k.r {
				fy := s.kivaFloor(k)
				if p.Y > fy-1.9 {
					return false
				}
			}
		}
		return true
	}
	box := Box{V(cliffPalaceX0, 0, cliffPalaceZ0), V(cliffPalaceX1, cliffPalaceMesa+3, cliffPalaceZ1)}
	t := NewGridWithSpan(g.Span())
	t.Fill(Func{box, func(p Vec3) bool {
		if !solid(p) {
			return false
		}
		for _, d := range cliffPalaceDirs {
			if !solid(p.Add(d.Scale(cliffPalaceSkin))) {
				return true
			}
		}
		return false
	}})
	for _, c := range t.Cells() {
		p := t.World(c)
		g.Use(s.rockColor(p, mp.at(p.X, p.Z), c)).Add(c)
	}
}

// rockColor colors floor, talus, ledge, alcove, kiva pits, mesa top and banded cliffs.
func (s CliffPalace) rockColor(p Vec3, h float64, c Cell) Color {
	n := Hash01(int(c.X), int(c.Y), int(c.Z), 31)
	x, z := p.X, p.Z
	for _, k := range cliffPalaceKivas {
		dx, dz := x-k.x, z-k.z
		d := math.Sqrt(dx*dx + dz*dz)
		fy := s.kivaFloor(k)
		if d < k.r+0.9 && p.Y < fy+0.2 && p.Y > fy-3.2 && (d > k.r-0.4 || p.Y < fy-1.4) {
			if p.Y < fy-1.6 {
				return Charcoal
			}
			if d < k.r+0.6 {
				return DarkStone
			}
		}
	}
	top := p.Y > h-0.9
	fz := s.faceZ(x)
	// Canyon floor and wash.
	if h <= cliffPalaceFloor+0.01 {
		if math.Abs(z-s.washZ(x)) < 2.2 {
			if n < 0.25 {
				return Stone
			}
			return Sand
		}
		g := s.fbm(x/6, z/6, 17)
		switch {
		case g > 0.56 && n < 0.8:
			return Grass
		case n < 0.12:
			return Terracotta
		}
		return Sand
	}
	// Alcove: floor, smoke-blackened ceiling, back wall.
	if s.planR(x, z) < 1.08 && p.Y > cliffPalaceLedge-1.2 {
		if s.cavity(p.Add(V(0, 1.4, 0))) && p.Y < s.floorY(x, z)+0.5 {
			if n < 0.15 {
				return Terracotta
			}
			return Sand
		}
		if s.cavity(p.Sub(V(0, 1.6, 0))) {
			soot := s.noise(x/4, z/3, 19) + 0.25*(1-math.Abs(x)/cliffPalaceA)
			switch {
			case soot > 0.92:
				return Charcoal
			case soot > 0.72:
				return DarkStone
			case soot < 0.3:
				return Sand
			}
			return Stone
		}
		if s.cavity(p.Add(V(0, 0, 1.6))) || s.cavity(p.Add(V(0, 0.8, 1.2))) {
			if n < 0.3 {
				return Stone
			}
			return Sand
		}
	}
	// Ledge top in front of the alcove.
	if top && math.Abs(h-cliffPalaceLedge) < 0.3 && z > fz-1 {
		if n < 0.15 {
			return Terracotta
		}
		return Sand
	}
	// Mesa top: piñon-juniper woodland floor.
	if top && z < fz && h > cliffPalaceMesa-6 {
		g := s.fbm(x/5, z/5, 23)
		switch {
		case g > 0.6:
			return Foliage
		case g > 0.42 || n < 0.2:
			return Grass
		}
		return Sand
	}
	// Talus and outer slopes.
	if top && z > fz+1.5 && h < cliffPalaceLedge-2 {
		g := s.fbm(x/5, z/5, 29)
		switch {
		case g > 0.6:
			return Grass
		case n < 0.15:
			return Stone
		case g < 0.35:
			return Terracotta
		}
		return Sand
	}
	// Desert-varnish streaks running down the canyon wall.
	if p.Y > cliffPalaceLedge+1 && z > fz-3 {
		m := s.noise(x/1.4, p.Y/16, 41)
		f := (p.Y - cliffPalaceLedge) / (cliffPalaceMesa - cliffPalaceLedge)
		if m > 0.78-0.3*f {
			return DarkStone
		}
	}
	// Banded sandstone.
	b := (p.Y + 1.5*s.fbm(x/9, z/9, 21) + 0.4*math.Sin(x/5)) / 2.3
	k := Hash01(int(math.Floor(b)), 23)
	if n < 0.08 {
		k = 1 - k
	}
	switch {
	case k < 0.5:
		return Sand
	case k < 0.8:
		return Terracotta
	case k < 0.95:
		return Stone
	}
	return Brick
}

// cliffPalaceRoom is one stack of masonry rooms.
type cliffPalaceRoom struct {
	x0, z0, x1, z1 float64
	base           float64
	storeys        int
	color          Color
	setback        float64
	seed           int
}

// village builds the rooms, towers, kivas and ladders.
func (s CliffPalace) village(g *Grid, interiors *[]Box) {
	blocked := func(x0, z0, x1, z1 float64) bool {
		for _, k := range cliffPalaceKivas {
			cx, cz := math.Max(x0, math.Min(x1, k.x)), math.Max(z0, math.Min(z1, k.z))
			if math.Hypot(cx-k.x, cz-k.z) < k.r {
				return true
			}
		}
		t := cliffPalaceSquareTower
		if x1 > t[0] && x0 < t[2] && z1 > t[1]+0.3 && z0 < t[3]-0.3 {
			return true
		}
		for _, r := range cliffPalaceRound {
			cx, cz := math.Max(x0, math.Min(x1, r[0])), math.Max(z0, math.Min(z1, r[1]))
			if math.Hypot(cx-r[0], cz-r[1]) < r[2] {
				return true
			}
		}
		return false
	}
	palette := []Color{Sand, Sand, Sand, Terracotta, Terracotta, Stone, Sand, Brick}
	rows := [][3]float64{ // z0, z1, base storeys
		{-11.4, -7.8, 2}, {-7.6, -4.0, 2}, {-3.8, -0.2, 1}, {0.0, 3.4, 0},
	}
	var rooms []cliffPalaceRoom
	for j, row := range rows {
		for i := 0; i < 14; i++ {
			x0 := -28.0 + float64(i)*4.0 + 0.5*(Hash01(i, j, 51)-0.5)
			x1 := x0 + 3.7 + 0.3*Hash01(i, j, 52)
			z0 := row[0] + 0.4*(Hash01(i, j, 53)-0.5)
			z1 := row[1] + 0.4*(Hash01(i, j, 54)-0.5)
			if blocked(x0, z0, x1, z1) {
				continue
			}
			roof := math.Inf(1)
			base := 0.0
			inside := true
			for _, q := range [][2]float64{{x0, z0}, {x1, z0}, {x0, z1}, {x1, z1}, {(x0 + x1) / 2, z0}} {
				if s.planR(q[0], q[1]) > 0.96 {
					inside = false
				}
				roof = math.Min(roof, s.roofY(q[0], q[1]))
				base = math.Max(base, s.floorY(q[0], q[1]))
			}
			if !inside {
				continue
			}
			h := Hash01(i, j, 55)
			if h < 0.08 || j == 3 && h < 0.3 {
				continue // Open courts.
			}
			n := int(row[2])
			if Hash01(i, j, 56) < 0.6 {
				n++
			}
			if j == 1 && math.Abs(x0) < 18 && Hash01(i, j, 57) < 0.6 {
				n++
			}
			if n > 3 {
				n = 3 // The Square Tower stays tallest.
			}
			maxN := int((roof - base - 0.2) / cliffPalaceStorey)
			if n > maxN {
				n = maxN
			}
			if n < 1 {
				continue
			}
			sb := 0.0
			if Hash01(i, j, 58) < 0.45 {
				sb = 1.1
			}
			rooms = append(rooms, cliffPalaceRoom{x0, z0, x1, z1, base, n, palette[int(Hash01(i, j, 59)*float64(len(palette)))], sb, i*31 + j*7})
		}
	}
	for _, r := range rooms {
		s.room(g, r, interiors)
	}
	// Square Tower: four storeys, the tallest building, reaching up toward the overhang.
	t := cliffPalaceSquareTower
	base := math.Max(s.floorY(t[0], t[1]), s.floorY(t[2], t[3]))
	s.room(g, cliffPalaceRoom{t[0], t[1], t[2], t[3], base, 4, Sand, 0, 977}, interiors)
	g.Paint(Terracotta, Block(V(t[0]-0.1, base+3*cliffPalaceStorey, t[1]-0.1), V(t[2]+0.1, base+3*cliffPalaceStorey+0.8, t[3]+0.1)))
	// Round towers.
	for i, rt := range cliffPalaceRound {
		c := P(rt[0], rt[1])
		fy := s.floorY(rt[0], rt[1])
		top := fy + rt[3]*cliffPalaceStorey
		g.With(Sand, func() {
			g.Fill(Cylinder{Center: c, Radius: rt[2], Y0: fy - 0.8, Y1: top})
		})
		g.Carve(Cylinder{Center: c, Radius: rt[2] - cliffPalaceWall - 0.1, Y0: fy + 0.01, Y1: top - 0.7})
		*interiors = append(*interiors, Box{V(rt[0]-rt[2]+0.8, fy, rt[1]-rt[2]+0.8), V(rt[0]+rt[2]-0.8, top-0.7, rt[1]+rt[2]-0.8)})
		for k := 0; k < int(rt[3]); k++ {
			yb := fy + float64(k)*cliffPalaceStorey
			a := math.Pi/2 + 0.6*(Hash01(i, k, 61)-0.5)
			w := Polar(c, rt[2]-0.4, a)
			g.Carve(Block(V(w.X-0.5, yb+0.9, w.Y-0.8), V(w.X+0.5, yb+1.7, w.Y+0.8)))
			// Masonry courses.
			g.Paint(Terracotta, Cylinder{Center: c, Radius: rt[2] + 0.1, Inner: rt[2] - 0.6, Y0: yb + cliffPalaceStorey - 0.45, Y1: yb + cliffPalaceStorey - 0.05})
		}
	}
	// Kivas: masonry rims, ladders poking out.
	for i, k := range cliffPalaceKivas {
		fy := s.kivaFloor(k)
		c := P(k.x, k.z)
		g.With(Sand, func() {
			g.Fill(Cylinder{Center: c, Radius: k.r + 0.9, Inner: k.r, Y0: fy - 0.4, Y1: fy + 0.5})
		})
		g.Paint(Terracotta, Cylinder{Center: c, Radius: k.r + 0.9, Inner: k.r + 0.5, Y0: fy + 0.1, Y1: fy + 0.5})
		if k.ladder {
			a := Hash01(i, 71) * 2 * math.Pi
			d := P(math.Cos(a), math.Sin(a))
			side := P(-d.Y, d.X)
			foot := c.Add(d.Scale(-0.4))
			tip := c.Add(d.Scale(k.r - 0.1))
			g.With(Wood, func() {
				for _, o := range []float64{-0.6, 0.6} {
					a0 := foot.Add(side.Scale(o)).At(fy - 1.7)
					a1 := tip.Add(side.Scale(o)).At(fy + 2.8)
					g.Fill(Capsule{a0, a1, 0.38})
				}
			})
		}
	}
	// Front retaining wall along the ledge edge.
	g.With(Stone, func() {
		for x := -26.0; x <= 26; x += 0.5 {
			z := s.ledgeZ(x) - 0.7
			g.Fill(Block(V(x-0.3, cliffPalaceLedge-0.8, z-0.5), V(x+0.3, cliffPalaceLedge+0.7, z+0.5)))
		}
	})
	// Ladders leaning against a few upper storeys.
	for i, r := range rooms {
		if r.storeys < 2 || Hash01(i, 81) > 0.15 {
			continue
		}
		cx := (r.x0+r.x1)/2 + 0.6
		top := r.base + cliffPalaceStorey + 1.2
		g.With(Wood, func() {
			for _, o := range []float64{-0.6, 0.6} {
				g.Fill(Capsule{V(cx+o, r.base, r.z1+1.4), V(cx+o, top, r.z1+0.2), 0.36})
			}
		})
	}
}

// room builds one stack of hollow masonry storeys facing +Z.
func (s CliffPalace) room(g *Grid, r cliffPalaceRoom, interiors *[]Box) {
	w := cliffPalaceWall
	g.With(Stone, func() {
		g.Fill(Block(V(r.x0, r.base-1.0, r.z0), V(r.x1, r.base, r.z1)))
	})
	for k := 0; k < r.storeys; k++ {
		yb := r.base + float64(k)*cliffPalaceStorey
		yt := yb + cliffPalaceStorey
		z1 := r.z1 - float64(k)*r.setback
		if z1-r.z0 < 2.6 {
			z1 = r.z0 + 2.6
		}
		col := r.color
		if k > 0 && Hash01(r.seed, k, 91) < 0.3 {
			col = []Color{Sand, Terracotta, White, Stone}[int(Hash01(r.seed, k, 92)*4)]
		}
		g.With(col, func() { g.Fill(Block(V(r.x0, yb, r.z0), V(r.x1, yt, z1))) })
		in := Box{V(r.x0+w, yb+0.01, r.z0+w), V(r.x1-w, yt-0.7, z1-w)}
		g.Carve(Cuboid(in))
		*interiors = append(*interiors, in)
		// Front opening: a T-shaped doorway or a small square window.
		cx := (r.x0+r.x1)/2 + 0.8*(Hash01(r.seed, k, 93)-0.5)
		d0, d1 := z1-w-0.3, z1+0.3
		switch v := Hash01(r.seed, k, 94); {
		case v < 0.55:
			g.Carve(Block(V(cx-0.45, yb+0.2, d0), V(cx+0.45, yb+1.4, d1)))
			g.Carve(Block(V(cx-0.95, yb+1.4, d0), V(cx+0.95, yb+2.05, d1)))
		case v < 0.9:
			g.Carve(Block(V(cx-0.5, yb+0.9, d0), V(cx+0.5, yb+1.8, d1)))
		}
		// Side window on some rooms.
		if Hash01(r.seed, k, 95) < 0.3 {
			cz := (r.z0 + z1) / 2
			g.Carve(Block(V(r.x1-w-0.3, yb+0.9, cz-0.45), V(r.x1+0.3, yb+1.7, cz+0.45)))
		}
		// Vigas: roof-beam ends protruding below the top storey's roof.
		if k == r.storeys-1 && Hash01(r.seed, 96) < 0.45 {
			g.With(Wood, func() {
				for vx := r.x0 + 0.8; vx < r.x1-0.5; vx += 1.4 {
					g.Fill(Capsule{V(vx, yt-0.45, z1-0.3), V(vx, yt-0.45, z1+0.45), 0.36})
				}
			})
		}
	}
}

// darkenInteriors paints every wall cell facing a room interior dark, so doorways and
// windows read as openings.
func (s CliffPalace) darkenInteriors(g *Grid, interiors []Box) {
	for _, c := range g.Cells() {
		p := g.World(c)
		for _, b := range interiors {
			if !b.Grow(1.2).Contains(p) {
				continue
			}
			inner, outer := false, false
			for _, n := range Neighbors {
				q := c.Add(n)
				if g.Has(q) {
					continue
				}
				in := false
				for _, o := range interiors {
					if o.Contains(g.World(q)) {
						in = true
						break
					}
				}
				if in {
					inner = true
				} else {
					outer = true
				}
			}
			if inner && !outer {
				g.Use(Charcoal)
				g.Remove(c)
				g.Add(c)
				break
			}
		}
	}
}

// vegetation scatters junipers and piñons on the talus, canyon floor and mesa top, with
// boulders on the talus.
func (s CliffPalace) vegetation(g *Grid, mp *cliffPalaceField) {
	tree := func(x, z, size float64, seed int) {
		base := mp.at(x, z)
		g.With(Wood, func() {
			g.Fill(Capsule{V(x, base-0.6, z), V(x+0.3*(Hash01(seed, 1)-0.5), base+size*0.45, z), 0.5})
		})
		g.With(Foliage, func() {
			g.Fill(Ellipsoid{V(x, base+size*0.7, z), V(size*0.5, size*0.42, size*0.5)})
			ox, oz := (Hash01(seed, 2)-0.5)*size*0.6, (Hash01(seed, 3)-0.5)*size*0.6
			g.Fill(Ellipsoid{V(x+ox, base+size*0.55, z+oz), V(size*0.38, size*0.32, size*0.38)})
		})
		if Hash01(seed, 4) < 0.35 {
			g.Paint(Grass, Ellipsoid{V(x, base+size*0.95, z), V(size*0.35, size*0.25, size*0.35)})
		}
	}
	// Talus.
	for i := 0; i < 70; i++ {
		x := cliffPalaceX0 + 3 + Hash01(i, 101)*(cliffPalaceX1-cliffPalaceX0-6)
		lz := s.ledgeZ(x)
		z := lz + 1.5 + Hash01(i, 102)*11
		h := mp.at(x, z)
		if h < cliffPalaceFloor+0.5 || h > cliffPalaceLedge-3 {
			continue
		}
		if Hash01(i, 103) < 0.35 {
			r := 0.9 + Hash01(i, 104)*0.9
			g.With(Stone, func() { g.Fill(Ellipsoid{V(x, h, z), V(r*1.2, r*0.8, r)}) })
			continue
		}
		tree(x, z, 2.6+Hash01(i, 105)*1.6, i)
	}
	// Canyon floor beside the wash.
	for i := 0; i < 22; i++ {
		x := cliffPalaceX0 + 4 + Hash01(i, 111)*(cliffPalaceX1-cliffPalaceX0-8)
		z := s.washZ(x) + (Hash01(i, 112)-0.5)*7
		if math.Abs(z-s.washZ(x)) < 2.6 || z > cliffPalaceZ1-2 {
			continue
		}
		tree(x, z, 2.8+Hash01(i, 113)*1.6, 500+i)
	}
	// Mesa top woodland.
	for i := 0; i < 90; i++ {
		x := cliffPalaceX0 + 3 + Hash01(i, 121)*(cliffPalaceX1-cliffPalaceX0-6)
		z := cliffPalaceZ0 + 3 + Hash01(i, 122)*(s.faceZ(x)-2.5-cliffPalaceZ0-3)
		h := mp.at(x, z)
		if h < cliffPalaceMesa-4 {
			continue
		}
		if Hash01(i, 123) < 0.35 {
			r := 0.9 + Hash01(i, 124)*0.5
			g.With(Foliage, func() { g.Fill(Ellipsoid{V(x, h+0.2, z), V(r, r*0.7, r)}) })
			continue
		}
		tree(x, z, 2.8+Hash01(i, 125)*1.8, 900+i)
	}
}
