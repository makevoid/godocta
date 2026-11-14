package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// Stonehenge is the monument as it stands today on Salisbury Plain, with every stone
// in its surveyed place and state. Seventeen sarsens of the outer circle still stand:
// the run on the north-east side carries three lintels end to end over the entrance,
// two more pairs keep theirs on the east and one on the west, and the gaps between are
// strewn with fallen uprights and lintels. Inside rises the horseshoe of great
// trilithons: three complete, the tallest stone of all standing alone at the centre
// back with its tenon bare beside its fallen partner and lintel, and one upright of the
// northern pair left above its broken neighbour. Blue-grey bluestones form a broken
// ring between the two and a horseshoe of slim pillars within; the Altar Stone lies
// half hidden under the fallen upright. Outside are the two Station Stones and the
// ring ditches of the others, the recumbent Slaughter Stone in the entrance, the white
// spots of the excavated Aubrey holes, the bank and chalk-white ditch, and the leaning
// Heel Stone in the Avenue whose twin banks run out to the edge. A worn visitor path
// loops round, close to the stones on the far side and across the Avenue in front, with
// a spur to the edge and a handful of visitors; sheep graze the rolling downland, and
// round barrows crown the ridge behind. The solstice axis and the Avenue (real
// north-east) point to +X/+Z, so real north is roughly +X and real east roughly +Z.
type Stonehenge struct{}

func init() { Register(Stonehenge{}) }

func (Stonehenge) Name() string  { return "stonehenge" }
func (Stonehenge) Title() string { return "Stonehenge" }

const (
	stonehengeAxis    = math.Pi / 4          // Plan angle of the solstice axis: the real north-east.
	stonehengeSite    = 2.6                  // Level turf inside the earthwork.
	stonehengeRimY    = 1.6                  // Turf level at the edge of the plate.
	stonehengeSkin    = 1.25                 // Thickness of the hollow turf skin.
	stonehengeSarsenR = 28.0                 // Sarsen circle, to the middle of the uprights.
	stonehengeBlueR   = 21.2                 // Bluestone circle.
	stonehengeTriR    = 14.0                 // Trilithon horseshoe.
	stonehengeOvalR   = 9.6                  // Bluestone horseshoe.
	stonehengeAubreyR = 37.5                 // Aubrey holes and Station Stones.
	stonehengeBankR   = 40.5                 // Crest of the bank.
	stonehengeDitchR  = 44.7                 // Middle of the ditch.
	stonehengeLintelY = stonehengeSite + 9.5 // Underside of the circle's lintels.
	stonehengePlateA  = 72.0                 // Plate semi-axes along and across the axis.
	stonehengePlateB  = 60.0
	stonehengePlateAt = 1.0 // Plate center along the axis.
	stonehengeHeelA   = 57.0
	stonehengeHeelB   = 3.5
)

var (
	stonehengeAlong  = P(math.Cos(stonehengeAxis), math.Sin(stonehengeAxis))
	stonehengeAcross = P(-math.Sin(stonehengeAxis), math.Cos(stonehengeAxis))
	// stonehengeBarrows are the round barrows: axis coordinates, radius and height.
	stonehengeBarrows = [][4]float64{{-59, -17, 6, 3.6}, {-61, 2, 7.2, 4.4}, {-58, 20, 5.6, 3.2}, {-22, -50, 5, 2.8}}
	// stonehengeRings are the small ring ditches: the Heel Stone's and the two Station
	// Stone "barrows". Axis coordinates and radius.
	stonehengeRings = [][3]float64{
		{stonehengeHeelA, stonehengeHeelB, 4.8},
		{stonehengeAubreyR * math.Cos(-62*math.Pi/180), stonehengeAubreyR * math.Sin(-62*math.Pi/180), 4.2},
		{stonehengeAubreyR * math.Cos(118*math.Pi/180), stonehengeAubreyR * math.Sin(118*math.Pi/180), 4.2},
	}
)

// ---------------------------------------------------------------------------------
// Helpers.

func stonehengeSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

// stonehengeBump is 1 at 0 and falls smoothly to 0 at ±1.
func stonehengeBump(x float64) float64 { return stonehengeSmooth(1 - math.Abs(x)) }

func stonehengeLerp(a, b, t float64) float64 { return a + (b-a)*t }

// stonehengeNoise is smooth value noise in [-1, 1].
func stonehengeNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return stonehengeLerp(stonehengeLerp(a, b, sx), stonehengeLerp(c, d, sx), sz)*2 - 1
}

func stonehengeFBM(x, z float64, seed int) float64 {
	return 0.6*stonehengeNoise(x, z, seed) + 0.3*stonehengeNoise(x*2.03+17, z*2.03+5, seed+1) +
		0.1*stonehengeNoise(x*4.1+3, z*4.1+29, seed+2)
}

// stonehengeNoise3 is smooth value noise in [0, 1] over space.
func stonehengeNoise3(x, y, z float64, seed int) float64 {
	x0, y0, z0 := math.Floor(x), math.Floor(y), math.Floor(z)
	fx, fy, fz := x-x0, y-y0, z-z0
	ix, iy, iz := int(x0), int(y0), int(z0)
	sx, sy, sz := fx*fx*(3-2*fx), fy*fy*(3-2*fy), fz*fz*(3-2*fz)
	h := func(i, j, k int) float64 { return Hash01(ix+i, iy+j, iz+k, seed) }
	lo := stonehengeLerp(stonehengeLerp(h(0, 0, 0), h(1, 0, 0), sx), stonehengeLerp(h(0, 0, 1), h(1, 0, 1), sx), sz)
	hi := stonehengeLerp(stonehengeLerp(h(0, 1, 0), h(1, 1, 0), sx), stonehengeLerp(h(0, 1, 1), h(1, 1, 1), sx), sz)
	return stonehengeLerp(lo, hi, sy)
}

func stonehengeCross(a, b Vec3) Vec3 {
	return V(a.Y*b.Z-a.Z*b.Y, a.Z*b.X-a.X*b.Z, a.X*b.Y-a.Y*b.X)
}

func stonehengeUnit(a Vec3) Vec3 { return a.Scale(1 / a.Length()) }

// stonehengeAt maps axis coordinates to plan: a runs from the center toward the
// north-east entrance, b across it toward the clockwise (real south-east) side.
func stonehengeAt(a, b float64) Vec2 {
	return stonehengeAlong.Scale(a).Add(stonehengeAcross.Scale(b))
}

// stonehengeLocal is the inverse of stonehengeAt.
func stonehengeLocal(p Vec2) (a, b float64) { return p.Dot(stonehengeAlong), p.Dot(stonehengeAcross) }

// stonehengeRing is the plan point at radius r and bearing deg, in degrees clockwise
// (seen from above) from the axis.
func stonehengeRing(r, deg float64) Vec2 {
	return Polar(P(0, 0), r, stonehengeAxis+deg*math.Pi/180)
}

// ---------------------------------------------------------------------------------
// Megaliths.

// stonehengeRock is one megalith in its own frame: a tapering block with rounded
// arrises whose faces wander in and out along its length. ax is the direction of its
// width, ay of its length (up, for a standing stone) and az of its thickness.
type stonehengeRock struct {
	o          Vec3 // Middle of the foot.
	ax, ay, az Vec3
	w, h, t    float64
	taper      float64 // Share of the width lost toward the top.
	crown      float64 // How far the shoulders of the top fall away.
	foot       float64 // The same for the foot of a fallen stone.
	rough      float64 // Amplitude of the wandering faces.
	bury       float64 // Length continued below the foot.
	level      bool    // Dressed flat on top to carry a lintel.
	seed       int
}

func (k *stonehengeRock) Contains(p Vec3) bool {
	d := p.Sub(k.o)
	y := d.Dot(k.ay)
	if y < -k.bury || y > k.h+0.6 {
		return false
	}
	u, v := d.Dot(k.ax), d.Dot(k.az)
	if math.Abs(u) > k.w/2+k.rough+0.05 || math.Abs(v) > k.t/2+k.rough+0.05 {
		return false
	}
	f := math.Max(0, math.Min(1, y/k.h))
	su, sv := 13.0, 29.0
	if u < 0 {
		su = -su
	}
	if v < 0 {
		sv = -sv
	}
	hw := k.w/2*(1-k.taper*f*f) + k.rough*stonehengeNoise(y*0.33+su, v*0.4, k.seed)
	ht := k.t/2*(1-0.5*k.taper*f) + 0.7*k.rough*stonehengeNoise(y*0.33+sv, u*0.3, k.seed+1)
	if hw < 0.2 || ht < 0.2 {
		return false
	}
	qa, qb := u/hw, v/ht
	qa, qb = qa*qa, qb*qb
	if qa*qa+qb*qb > 1 {
		return false
	}
	round := 0.8*qa*qa + 0.3*qb
	top := k.h - k.crown*round
	if !k.level {
		top += 1.6 * k.rough * stonehengeNoise(u*0.45, v*0.45, k.seed+2)
	}
	return y <= top && y >= k.foot*round-k.bury
}

func (k *stonehengeRock) Bounds() Box {
	b := EmptyBox()
	hu, hv := k.w/2+k.rough+0.2, k.t/2+k.rough+0.2
	for _, y := range []float64{-k.bury, k.h + 0.6} {
		for _, u := range []float64{-hu, hu} {
			for _, v := range []float64{-hv, hv} {
				q := k.o.Add(k.ay.Scale(y)).Add(k.ax.Scale(u)).Add(k.az.Scale(v))
				b = b.Union(Box{q, q})
			}
		}
	}
	return b
}

// stonehengeStanding sets a stone upright at a plan point. yaw is the plan angle of its
// broad face; lean tips it over (plan offset per unit of height).
func stonehengeStanding(at Vec2, base, yaw float64, lean Vec2, w, t, h float64, seed int) *stonehengeRock {
	up := stonehengeUnit(V(lean.X, 1, lean.Y))
	ax := V(math.Cos(yaw), 0, math.Sin(yaw))
	ax = stonehengeUnit(ax.Sub(up.Scale(ax.Dot(up))))
	return &stonehengeRock{o: at.At(base), ax: ax, ay: up, az: stonehengeCross(ax, up),
		w: w, h: h, t: t, taper: 0.1, crown: 0.9, rough: 0.2, bury: 1.3, seed: seed}
}

// stonehengeFallen lays a stone flat with its length along the plan angle dir. bottom
// is the height of its underside.
func stonehengeFallen(at Vec2, bottom, dir, length, w, t float64, seed int) *stonehengeRock {
	ay := V(math.Cos(dir), 0, math.Sin(dir))
	ax := V(-math.Sin(dir), 0, math.Cos(dir))
	return &stonehengeRock{o: at.At(bottom + t/2).Sub(ay.Scale(length / 2)), ax: ax, ay: ay, az: V(0, 1, 0),
		w: w, h: length, t: t, taper: 0.08, crown: 0.7, foot: 0.5, rough: 0.2, seed: seed}
}

// raise rasterizes one megalith and colors each of its cells with tint.
func (Stonehenge) raise(g *Grid, rock Solid, tint func(Vec3, Cell) Color) {
	tmp := NewGridWithSpan(g.Span())
	tmp.FillVolume(rock)
	brush := g.Brush()
	for _, c := range tmp.Cells() {
		g.Remove(c)
		g.Use(tint(tmp.World(c), c)).Add(c)
	}
	g.Use(brush)
}

// sarsen is the weathering of one sarsen: pale stone with darker stained patches,
// blooms of yellow lichen, a few blue-grey hollows and a damp foot above ground.
func (Stonehenge) sarsen(seed int, ground, dark float64) func(Vec3, Cell) Color {
	return func(p Vec3, c Cell) Color {
		n := stonehengeNoise3(p.X*0.4, p.Y*0.28, p.Z*0.4, seed)
		m := stonehengeNoise3(p.X*0.85+9, p.Y*0.85, p.Z*0.85+4, seed+7)
		hash := Hash01(int(c.X), int(c.Y), int(c.Z), 5)
		switch {
		case p.Y < ground+0.8 && hash < 0.6:
			return DarkStone
		case n > 0.62-dark:
			return DarkStone
		case m > 0.69:
			return Sand
		case n < 0.28:
			return Slate
		case n > 0.56-dark && hash < 0.4:
			return DarkStone
		}
		return Stone
	}
}

// bluestone is the darker, blue-grey dolerite of the smaller stones.
func (Stonehenge) bluestone(seed int) func(Vec3, Cell) Color {
	return func(p Vec3, c Cell) Color {
		n := stonehengeNoise3(p.X*0.7, p.Y*0.5, p.Z*0.7, seed)
		switch {
		case n > 0.68:
			return Stone
		case n > 0.57:
			return DarkStone
		}
		return Slate
	}
}

// lintel lays a curved lintel on sarsen n and its clockwise neighbour.
func (s Stonehenge) lintel(g *Grid, n int) {
	mid := stonehengeAxis + float64(n)*12*math.Pi/180
	half := 6 * math.Pi / 180
	c := Polar(P(0, 0), stonehengeSarsenR, mid)
	y0 := stonehengeLintelY
	seed := 200 + n
	block := Func{Box{V(c.X-4.6, y0-0.3, c.Y-4.6), V(c.X+4.6, y0+2.4, c.Y+4.6)}, func(p Vec3) bool {
		q := p.Plan()
		da := math.Remainder(q.Angle()-mid, 2*math.Pi)
		if math.Abs(da) > half {
			return false
		}
		along := da * stonehengeSarsenR
		if math.Abs(q.Length()-stonehengeSarsenR) > 1.42+0.1*stonehengeNoise(along*0.4, p.Y*0.5, seed) {
			return false
		}
		y := p.Y - y0
		return y >= -0.05 && y <= 1.85+0.2*stonehengeNoise(along*0.35, 1.7, seed+1)
	}}
	tint := s.sarsen(seed, -100, 0)
	s.raise(g, block, func(p Vec3, cell Cell) Color {
		// A dark joint where the lintel butts against the next.
		da := math.Remainder(p.Plan().Angle()-mid, 2*math.Pi)
		if da > half-0.45/stonehengeSarsenR {
			return DarkStone
		}
		return tint(p, cell)
	})
}

// trilithon raises a pair of great uprights at a bearing on the horseshoe. first and
// second say whether the anticlockwise and the clockwise upright still stand; a lintel
// caps the pair when both do, and a lone survivor shows its tenon.
func (s Stonehenge) trilithon(g *Grid, deg, h float64, first, second bool, seed int) {
	ang := stonehengeAxis + deg*math.Pi/180
	c := Polar(P(0, 0), stonehengeTriR, ang)
	tan := P(-math.Sin(ang), math.Cos(ang))
	base := stonehengeSite
	whole := first && second
	for i, stands := range []bool{first, second} {
		if !stands {
			continue
		}
		at := c.Add(tan.Scale(float64(2*i-1) * 3.0))
		rock := stonehengeStanding(at, base, ang+math.Pi/2, P(0, 0),
			4.3+0.3*Hash01(seed, i, 1), 2.9+0.3*Hash01(seed, i, 2), h, seed+i)
		rock.taper, rock.rough, rock.crown = 0.13, 0.15, 0.35
		if whole {
			rock.crown, rock.level = 0, true
		}
		s.raise(g, rock, s.sarsen(seed+i, base, 0))
		if !whole {
			tenon := stonehengeStanding(at, base+h-0.4, ang+math.Pi/2, P(0, 0), 1.6, 1.4, 1.3, seed+9)
			tenon.rough, tenon.crown, tenon.bury = 0.05, 0.3, 0.2
			s.raise(g, tenon, s.sarsen(seed+i, base, 0))
		}
	}
	if whole {
		beam := stonehengeFallen(c, base+h-0.1, ang+math.Pi/2, 11, 3.5, 2.2, seed+5)
		beam.taper, beam.crown, beam.foot, beam.rough = 0, 0.3, 0.3, 0.12
		s.raise(g, beam, s.sarsen(seed+5, -100, 0))
	}
}

// megaliths builds the stone settings inside the earthwork.
func (s Stonehenge) megaliths(g *Grid) {
	base := stonehengeSite
	rad := math.Pi / 180

	// The sarsen circle. Stones are numbered clockwise from the axis; lintel n rests on
	// sarsen n and the next one round.
	var up, bears [31]bool
	for _, n := range []int{1, 2, 3, 4, 5, 6, 7, 10, 11, 16, 21, 22, 23, 27, 28, 29, 30} {
		up[n] = true
	}
	lintels := []int{29, 30, 1, 4, 6, 21}
	for _, n := range lintels {
		bears[n], bears[n%30+1] = true, true
	}
	for n := 1; n <= 30; n++ {
		if !up[n] {
			continue
		}
		deg := (float64(n) - 0.5) * 12
		rock := stonehengeStanding(stonehengeRing(stonehengeSarsenR, deg), base, stonehengeAxis+deg*rad+math.Pi/2,
			P((Hash01(n, 5)-0.5)*0.05, (Hash01(n, 6)-0.5)*0.05),
			3.9+0.45*Hash01(n, 1), 2.4+0.35*Hash01(n, 2), 8.8+1.1*Hash01(n, 3), 100+n)
		rock.crown = 0.7 + 0.7*Hash01(n, 4)
		switch {
		case bears[n]:
			rock.ay, rock.ax = V(0, 1, 0), V(math.Cos(stonehengeAxis+deg*rad+math.Pi/2), 0, math.Sin(stonehengeAxis+deg*rad+math.Pi/2))
			rock.az = stonehengeCross(rock.ax, rock.ay)
			rock.h, rock.crown, rock.level = stonehengeLintelY+0.15-base, 0, true
		case n == 11:
			rock.w, rock.t, rock.h = 2.7, 1.9, 5.6 // The odd half-size stone.
		case n == 16:
			rock.h, rock.crown = 9.9, 0.5
		}
		s.raise(g, rock, s.sarsen(100+n, base, 0))
	}
	for _, n := range lintels {
		s.lintel(g, n)
	}

	// lie drops a fallen sarsen at a ring point, turned from the radial direction.
	lie := func(r, deg, turn, length, w, t, sink float64, seed int) {
		at := stonehengeRing(r, deg)
		rock := stonehengeFallen(at, base-sink, stonehengeAxis+(deg+turn)*rad, length, w, t, seed)
		s.raise(g, rock, s.sarsen(seed, base, 0.05))
	}
	// Fallen uprights of the circle (stones 8, 9, 12, 14, 15, 19, 25, 26).
	lie(29.5, 90, 80, 4.8, 3.6, 2.2, 0.5, 408)
	lie(27.0, 101, 100, 5.2, 3.8, 2.3, 0.5, 409)
	lie(24.2, 106, 30, 4.2, 3.4, 2.2, 0.6, 410)
	lie(26.3, 138, 62, 8.8, 4.0, 2.4, 0.5, 412)
	lie(24.6, 162, 12, 8.6, 4.0, 2.4, 0.5, 414)
	lie(27.8, 175, 95, 4.6, 3.8, 2.3, 0.5, 415)
	lie(30.4, 171, 40, 3.6, 3.0, 2.0, 0.6, 416)
	lie(26.4, 222, 70, 8.2, 4.0, 2.4, 0.5, 419)
	lie(24.8, 294, 10, 8.4, 4.0, 2.4, 0.5, 425)
	lie(28.6, 306, 85, 3.8, 3.4, 2.2, 0.6, 426)
	// Fallen lintels.
	lie(31.3, 150, 90, 5.6, 2.7, 1.9, 0.4, 430)
	lie(24.3, 246, 80, 5.6, 2.7, 1.9, 0.4, 431)
	lie(30.6, 276, 100, 5.6, 2.7, 1.9, 0.4, 432)
	lie(24.4, 122, 85, 5.4, 2.7, 1.9, 0.4, 433)

	// The trilithon horseshoe, open to the north-east.
	s.trilithon(g, 52, 12.3, true, true, 500)
	s.trilithon(g, 112, 13.4, true, true, 510)
	s.trilithon(g, 180, 16.6, false, true, 520) // Stone 56, the tallest, stands alone.
	s.trilithon(g, 248, 13.4, true, true, 530)
	s.trilithon(g, 308, 12.3, false, true, 540) // Stone 60.

	// drop lays a fallen piece at axis coordinates, turned from the axis.
	drop := func(a, b, turn, length, w, t, sink float64, seed int, tint func(Vec3, Cell) Color) {
		rock := stonehengeFallen(stonehengeAt(a, b), base-sink, stonehengeAxis+turn*rad, length, w, t, seed)
		s.raise(g, rock, tint)
	}
	// The Altar Stone, then stone 55 broken in two across it and the great lintel.
	drop(-5.4, 0.4, 90, 10.4, 2.3, 1.6, 0.7, 580, func(p Vec3, c Cell) Color {
		if stonehengeNoise3(p.X*0.6, p.Y, p.Z*0.6, 580) > 0.6 {
			return Sand
		}
		return Slate
	})
	drop(-9.7, 3.0, 0, 5.4, 4.4, 2.7, 0.6, 555, s.sarsen(555, base, 0.05))
	drop(-3.4, 3.6, 8, 5.6, 4.3, 2.6, -0.5, 556, s.sarsen(556, base, 0.05))
	drop(2.3, 1.5, 100, 9.6, 3.4, 2.2, 0.4, 557, s.sarsen(557, base, 0.05))
	// Stone 59 in pieces, with a fragment of its lintel.
	lie(11.7, -64, 0, 4.4, 4.2, 2.7, 0.5, 559)
	lie(7.0, -60, 12, 4.2, 4.0, 2.6, 0.5, 560)
	lie(9.2, -30, 80, 4.0, 3.2, 2.1, 0.5, 561)

	// blue sets one bluestone: S standing, L leaning, s a stump, F fallen.
	blue := func(r, deg float64, state byte, h float64, seed int) {
		at := stonehengeRing(r, deg)
		yaw := stonehengeAxis + deg*rad + math.Pi/2 + (Hash01(seed, 1)-0.5)*0.6
		w, t := 1.7+0.5*Hash01(seed, 2), 1.3+0.25*Hash01(seed, 3)
		lean := P(0, 0)
		switch state {
		case 'L':
			lean = Polar(P(0, 0), 0.3, Hash01(seed, 4)*2*math.Pi)
		case 's':
			h = 1.2 + 0.5*Hash01(seed, 5)
		case 'F':
			rock := stonehengeFallen(at, base-0.3, Hash01(seed, 6)*math.Pi, h, w, t, seed)
			rock.rough = 0.1
			s.raise(g, rock, s.bluestone(seed))
			return
		case '.':
			return
		}
		rock := stonehengeStanding(at, base, yaw, lean, w, t, h, seed)
		rock.taper, rock.crown, rock.rough, rock.bury = 0.16, 0.5, 0.1, 1.0
		s.raise(g, rock, s.bluestone(seed))
	}
	// The bluestone circle, 6° to a place, clockwise from the axis.
	circle := "S.SS.sS.SLS.S.F" + "S.sS..F.S.L...s" + "..S..F...s.S..." + ".S....SF.S.sSS."
	for k := 0; k < len(circle); k++ {
		blue(stonehengeBlueR+(Hash01(k, 40)-0.5)*0.8, 3+6*float64(k), circle[k], 2.9+1.6*Hash01(k, 41), 300+k)
	}
	// The bluestone horseshoe of taller pillars, counted either way from the apex.
	oval := "S.SSLS" + "SSLSs."
	for k := 0; k < len(oval); k++ {
		deg := 180 - 17*float64(k+1)
		if k >= 6 {
			deg = 180 + 17*float64(k-5)
		}
		blue(stonehengeOvalR, deg, oval[k], 4.6+1.5*Hash01(k, 42), 370+k)
	}
}

// outliers builds the stones beyond the circle.
func (s Stonehenge) outliers(g *Grid) {
	rad := math.Pi / 180
	// The Heel Stone: a rough unworked sarsen leaning toward the circle.
	at := stonehengeAt(stonehengeHeelA, stonehengeHeelB)
	y := s.height(at.X, at.Y)
	heel := stonehengeStanding(at, y, stonehengeAxis+math.Pi/2+0.2, stonehengeAlong.Scale(-0.3), 4.8, 3.6, 10.8, 600)
	heel.taper, heel.crown, heel.rough, heel.bury = 0.38, 2.2, 0.38, 1.8
	s.raise(g, heel, s.sarsen(600, y, 0.06))

	// The Slaughter Stone, sunk in the turf of the entrance.
	at = stonehengeAt(40.3, 4.8)
	y = s.height(at.X, at.Y)
	s.raise(g, stonehengeFallen(at, y-1.1, stonehengeAxis+22*rad, 10, 3.8, 2.0, 601), s.sarsen(601, y, 0.04))

	// Station Stones 91 (fallen) and 93 (a stump still upright).
	at = stonehengeRing(stonehengeAubreyR, 62)
	y = s.height(at.X, at.Y)
	s.raise(g, stonehengeFallen(at, y-0.4, stonehengeAxis+100*rad, 4.8, 2.3, 1.9, 602), s.sarsen(602, y, 0.04))
	at = stonehengeRing(stonehengeAubreyR, -118)
	y = s.height(at.X, at.Y)
	stump := stonehengeStanding(at, y, stonehengeAxis-28*rad, P(0, 0), 2.3, 1.8, 2.7, 603)
	stump.crown = 0.6
	s.raise(g, stump, s.sarsen(603, y, 0.04))
}

// ---------------------------------------------------------------------------------
// Ground.

// stonehengePath is the radius of the visitor path at a bearing in radians: far out in
// front, where it crosses the Avenue, and close to the stones behind.
func stonehengePath(phi float64) (rho, slope float64) {
	d := phi - 35*math.Pi/180
	rho = 43.5 + 9*math.Cos(d) - 2.3*math.Cos(2*d)
	return rho, (-9*math.Sin(d) + 4.6*math.Sin(2*d)) / rho
}

// pathDist is the plan distance from a point (in axis coordinates) to the middle of
// the visitor path, including the spur to the edge of the plate.
func (Stonehenge) pathDist(a, b float64) float64 {
	rho, slope := stonehengePath(math.Atan2(b, a))
	d := math.Abs(math.Hypot(a, b)-rho) / math.Sqrt(1+slope*slope)
	// The spur.
	r0, _ := stonehengePath(-50 * math.Pi / 180)
	p0 := Polar(P(0, 0), r0, -50*math.Pi/180)
	p1 := Polar(P(0, 0), 74, -57*math.Pi/180)
	ab := p1.Sub(p0)
	t := math.Max(0, math.Min(1, P(a, b).Sub(p0).Dot(ab)/ab.Dot(ab)))
	return math.Min(d, P(a, b).Sub(p0.Add(ab.Scale(t))).Length())
}

// stonehengeWorks are the strengths (0 to 1) of the earthworks at a point.
type stonehengeWorks struct {
	bank, ditch     float64 // The henge.
	avBank, avDitch float64 // The Avenue.
	ring            float64 // Small ring ditches.
}

func (s Stonehenge) works(a, b float64) stonehengeWorks {
	var w stonehengeWorks
	r := math.Hypot(a, b)
	if r < stonehengeAubreyR-6 {
		return w
	}
	// The path crosses everything on a causeway.
	keep := stonehengeSmooth((s.pathDist(a, b) - 1.5) / 2.2)
	gate := 1.0
	if a > 0 {
		gate = stonehengeSmooth((math.Abs(b) - 6.5) / 3.5)
	}
	w.bank = keep * gate * stonehengeBump((r-stonehengeBankR)/2.5)
	w.ditch = keep * gate * stonehengeBump((r-stonehengeDitchR)/2.7)
	if a > 0 && r > stonehengeDitchR {
		fade := keep * stonehengeSmooth((r-stonehengeDitchR-1.5)/3)
		w.avBank = fade * stonehengeBump((math.Abs(b)-10.5)/1.5)
		w.avDitch = fade * stonehengeBump((math.Abs(b)-13.4)/1.4)
	}
	for _, c := range stonehengeRings {
		d := math.Hypot(a-c[0], b-c[1])
		w.ring = math.Max(w.ring, keep*stonehengeBump((d-c[2])/1.1))
	}
	return w
}

// barrows returns the mound height and the strength of the ring ditch at a point.
func (Stonehenge) barrows(a, b float64) (mound, moat float64) {
	for _, bw := range stonehengeBarrows {
		d := math.Hypot(a-bw[0], b-bw[1])
		if d < bw[2] {
			mound = math.Max(mound, bw[3]*math.Pow(0.5+0.5*math.Cos(math.Pi*d/bw[2]), 0.7))
		}
		moat = math.Max(moat, stonehengeBump((d-bw[2]-0.8)/1.2))
	}
	return
}

// aubrey reports whether a point lies on one of the concrete spots that mark the
// excavated Aubrey holes: the 32 clockwise from the entrance.
func (Stonehenge) aubrey(a, b float64) bool {
	if math.Abs(math.Hypot(a, b)-stonehengeAubreyR) > 0.7 {
		return false
	}
	const step = 2 * math.Pi / 56
	phi := math.Atan2(b, a)
	if phi < 0 {
		phi += 2 * math.Pi
	}
	k := math.Floor(phi / step)
	if k >= 32 {
		return false
	}
	c := Polar(P(0, 0), stonehengeAubreyR, (k+0.5)*step)
	return P(a, b).Sub(c).Length() < 0.62
}

// height is the turf level, or -1 beyond the plate.
func (s Stonehenge) height(x, z float64) float64 {
	a, b := stonehengeLocal(P(x, z))
	e := math.Hypot((a-stonehengePlateAt)/stonehengePlateA, b/stonehengePlateB)
	if e >= 1 {
		return -1
	}
	r := math.Hypot(a, b)
	// Level inside the earthwork, rolling downland beyond, a ridge behind.
	h := stonehengeSite + 1.2*stonehengeSmooth((r-48)/12)*stonehengeFBM(x*0.03, z*0.03, 11)
	h += 3.4 * stonehengeSmooth((-45-a)/15) * (0.85 + 0.15*stonehengeNoise(b*0.06, 3.3, 17))
	mound, moat := s.barrows(a, b)
	h += mound - 0.6*moat
	w := s.works(a, b)
	h += 1.5*w.bank - 1.9*w.ditch + 0.7*w.avBank - 0.6*w.avDitch - 0.55*w.ring
	// Everything eases down to the rim.
	rim := (1 - e) * math.Hypot(a-stonehengePlateAt, b) / math.Max(e, 1e-6)
	return stonehengeRimY + (h-stonehengeRimY)*stonehengeSmooth(rim/9)
}

// turf colors one column of ground.
func (s Stonehenge) turf(c Cell, x, z float64) Color {
	a, b := stonehengeLocal(P(x, z))
	r := math.Hypot(a, b)
	hash := Hash01(int(c.X), int(c.Z), 3)
	// The worn visitor path.
	if s.pathDist(a, b) < 1.2+0.3*stonehengeNoise(x*0.6, z*0.6, 31) {
		if hash < 0.08 {
			return Stone
		}
		return Sand
	}
	// Chalk shows white in the bottom of the ditch.
	w := s.works(a, b)
	switch {
	case w.ditch > 0.55:
		if hash < 0.14 {
			return Stone
		}
		if hash > 0.94 {
			return Grass
		}
		return White
	case w.ditch > 0.3 && hash < 0.4:
		return White
	case s.aubrey(a, b):
		return White
	}
	_, moat := s.barrows(a, b)
	if moat > 0.5 || w.ring > 0.5 || w.avDitch > 0.5 {
		return Foliage
	}
	if w.avBank > 0.55 && hash < 0.45 {
		return Sand
	}
	// Bare, trodden earth among the stones.
	if r < 26 && stonehengeFBM(x*0.13, z*0.13, 47) > 0.4 && hash < 0.85 {
		return Sand
	}
	patch := stonehengeFBM(x*0.07, z*0.07, 41)
	switch {
	case patch > 0.26 && hash < 0.85:
		return Foliage
	case r > 48 && stonehengeFBM(x*0.1+40, z*0.1, 43) > 0.45 && hash < 0.7:
		return Sand
	case hash < 0.05:
		return Foliage
	}
	return Grass
}

// terrain fills the downland as a hollow skin over the height field, closed by a turf
// wall round the edge of the plate.
func (s Stonehenge) terrain(g *Grid) {
	span := g.Span()
	inv := 2 / span
	n := int32(math.Ceil(70 * inv))
	for cz := -n; cz <= n; cz++ {
		for cx := -n; cx <= n; cx++ {
			if (cx^cz)&1 != 0 {
				continue
			}
			x, z := float64(cx)*span/2, float64(cz)*span/2
			h := s.height(x, z)
			if h < 0 {
				continue
			}
			lo := h
			for k := 0; k < 8; k++ {
				q := Polar(P(x, z), 1.0, float64(k)*math.Pi/4)
				lo = math.Min(lo, s.height(q.X, q.Y))
			}
			if lo < 0 {
				lo = 0 // The edge of the plate.
			} else {
				lo = math.Max(lo-stonehengeSkin, 0)
			}
			cy := int32(math.Ceil(lo * inv))
			if (cy^cx)&1 != 0 {
				cy++
			}
			col := s.turf(Cell{cx, 0, cz}, x, z)
			for ; float64(cy)*span/2 <= h; cy += 2 {
				c := Cell{cx, cy, cz}
				use := col
				if float64(cy)*span/2 < h-1.3 {
					use = Grass
				}
				g.Use(use).Add(c)
			}
		}
	}
}

// ---------------------------------------------------------------------------------
// Life.

// person stands a figure five cells tall on the ground.
func (s Stonehenge) person(g *Grid, at Vec2, shirt, legs Color) {
	c := Nearest(at.At(s.height(at.X, at.Y)).Scale(1 / g.Span()))
	for g.Has(c) {
		c.Y += 2
	}
	for i, col := range []Color{legs, legs, shirt, shirt, Sand} {
		g.Use(col).Add(Cell{c.X, c.Y + int32(2*i), c.Z})
	}
}

// sheep is a woolly body with a dark head, grazing along the plan angle dir.
func (s Stonehenge) sheep(g *Grid, at Vec2, dir float64) {
	y := s.height(at.X, at.Y)
	f := P(math.Cos(dir), math.Sin(dir))
	g.Use(White).Fill(Func{Box{V(at.X-1.4, y, at.Y-1.4), V(at.X+1.4, y+2.2, at.Y+1.4)}, func(p Vec3) bool {
		d := p.Plan().Sub(at)
		u, v := d.Dot(f), d.X*-f.Y+d.Y*f.X
		q, r, t := u/1.15, (p.Y-y-1.0)/0.72, v/0.72
		return q*q+r*r+t*t <= 1
	}})
	g.Use(Charcoal).Fill(Sphere(at.Add(f.Scale(1.4)).At(y+0.8), 0.45))
}

// life adds the visitors on the path and the flocks on the down.
func (s Stonehenge) life(g *Grid) {
	shirts := []Color{Red, Navy, Gold, White, Terracotta, Glass, Brick}
	legs := []Color{Navy, Charcoal, Wood}
	// Visitors: bearing in degrees and offset across the path.
	for i, v := range [][2]float64{
		{-6, 0.5}, {-2, -0.6}, {24, 0}, {72, 0.3}, {128, 0},
		{203, -0.8}, {208, 0.5}, {214, -0.9}, {221, -0.2},
		{260, 0.3}, {300, 0}, {330, 0.5}, {334, -0.5},
	} {
		rho, _ := stonehengePath(v[0] * math.Pi / 180)
		s.person(g, stonehengeRing(rho+v[1], v[0]), shirts[(i*3+1)%len(shirts)], legs[i%len(legs)])
	}
	r0, _ := stonehengePath(-50 * math.Pi / 180)
	p0, p1 := stonehengeRing(r0, -50), stonehengeRing(74, -57)
	for i, t := range []float64{0.3, 0.36, 0.62} {
		at := p0.Add(p1.Sub(p0).Scale(t)).Add(P(0.5-float64(i%2), 0))
		s.person(g, at, shirts[(i*2+4)%len(shirts)], legs[(i+1)%len(legs)])
	}

	// Flocks: axis coordinates, spread and head count.
	for fi, f := range [][4]float64{{45, 38, 8, 9}, {-41, 41, 7, 7}, {-52, -34, 6, 5}} {
		var placed []Vec2
		for k := 0; len(placed) < int(f[3]) && k < 200; k++ {
			at := stonehengeAt(f[0]+(Hash01(fi, k, 71)-0.5)*2*f[2], f[1]+(Hash01(fi, k, 72)-0.5)*2*f[2])
			a, b := stonehengeLocal(at)
			if math.Hypot((a-stonehengePlateAt)/stonehengePlateA, b/stonehengePlateB) > 0.9 || s.pathDist(a, b) < 3.5 {
				continue
			}
			near := false
			for _, q := range placed {
				near = near || q.Sub(at).Length() < 3.4
			}
			if near {
				continue
			}
			placed = append(placed, at)
			s.sheep(g, at, Hash01(fi, k, 73)*2*math.Pi)
		}
	}
}

func (s Stonehenge) Build(g *Grid) Order {
	s.terrain(g)
	s.megaliths(g)
	s.outliers(g)
	s.life(g)
	return Sweep(P(0, 0), stonehengeAxis)
}
