package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// VikingHarbor is a Norse settlement at the head of a fjord. Steep terraced rock walls
// with pines, snow and three waterfalls rise from the water on the west and north; a
// pebble beach and two wooden jetties front the village plain on the east shore. Four
// clinker-built longships with dragon prows and rows of painted shields ride the fjord:
// one under a striped sail, one rowing, two moored. Behind a log palisade with a gate
// stand turf-roofed longhouses, a great hall with crossed dragon gables, a stave church
// with stacked shingled roofs and a smithy, fish racks and sheep outside. Water is a
// single layer at y = 0. North is -Z; the fjord opens south (+Z).
type VikingHarbor struct{}

func init() { Register(VikingHarbor{}) }

func (VikingHarbor) Name() string  { return "viking_harbor" }
func (VikingHarbor) Title() string { return "Viking Harbor" }

const (
	vikingHarborX0, vikingHarborX1 = -58.0, 64.0
	vikingHarborZ0, vikingHarborZ1 = -55.0, 54.0
	vikingHarborPlain              = 2.4  // Village ground level.
	vikingHarborStep               = 0.25 // Terrain raster resolution.
)

var (
	vikingHarborHall      = P(34, -8)
	vikingHarborChurch    = P(44, -32)
	vikingHarborSmithy    = P(21, 44)
	vikingHarborGate      = P(32, 35)
	vikingHarborWestFalls = []float64{-6, 40}
	vikingHarborBackFall  = -14.0
	vikingHarborPalisade  = []Vec2{P(12, -36), P(22, -39.5), P(38, -40.5), P(52, -38.5), P(59, -28), P(60, -10),
		P(59, 10), P(54, 24), P(44, 33), P(32, 35), P(20, 34), P(12, 31)}
)

func (VikingHarbor) smooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func (VikingHarbor) lerp(a, b, t float64) float64 { return a + (b-a)*t }

// noise is smooth value noise in [-1, 1].
func (s VikingHarbor) noise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return s.lerp(s.lerp(a, b, sx), s.lerp(c, d, sx), sz)*2 - 1
}

func (s VikingHarbor) fbm(x, z float64, seed int) float64 {
	return 0.6*s.noise(x, z, seed) + 0.3*s.noise(x*2.03+17, z*2.03+5, seed+1) +
		0.1*s.noise(x*4.1+3, z*4.1+29, seed+2)
}

// ridge is ridged noise in [0, 1]: sharp crests where the noise crosses zero.
func (s VikingHarbor) ridge(a, b float64, seed int) float64 {
	return 1 - math.Abs(s.fbm(a, b, seed))
}

// Shorelines: the west wall's foot (x at z), the north wall's foot (z at x; it steps
// back behind the village) and the village shore (x at z).
func (s VikingHarbor) westShore(z float64) float64 {
	return -33 + 4.5*s.fbm(z*0.04, 3.1, 1) - 4*s.smooth((z-10)/40)
}

func (s VikingHarbor) backShore(x float64) float64 {
	return -27 + 1.5*s.fbm(x*0.07, 7.7, 2) - 14*s.smooth((x-8)/14)
}

func (s VikingHarbor) eastShore(z float64) float64 { return 8 + 1.6*s.fbm(z*0.08, 1.3, 3) }

// backBeach is the width of the pebble beach at the fjord head.
func (s VikingHarbor) backBeach(x float64) float64 { return 5 * (1 - s.smooth(x/8)) }

// wall is a fjord-wall profile: a steep foot rising to a crest 16 units inland.
func (VikingHarbor) wall(d, crest, k float64) float64 {
	if d <= 0 {
		return 0
	}
	cd := 13 + 8*k
	if d >= cd {
		return crest
	}
	return crest * math.Pow(math.Sin(d/cd*math.Pi/2), 0.5+0.7*k)
}

// gully is the depth of a waterfall's channel at lateral offset lat and inland distance d.
func (s VikingHarbor) gully(lat, d float64) float64 {
	if d < 0.5 || math.Abs(lat) >= 3 {
		return 0
	}
	t := lat / 3
	return 2.6 * (1 - t*t) * s.smooth((d-0.5)/2) * (1 - s.smooth((d-17)/4))
}

// rock adds crags and ledges to a wall height.
func (s VikingHarbor) rock(x, z, h float64) float64 {
	if h <= 2 {
		return h
	}
	k := s.smooth((h - 2) / 5)
	h += (2.2*s.fbm(x*0.11, z*0.11, 5) + 3*s.fbm(x*0.04, z*0.04, 6)) * k
	const step = 3.4
	t := h / step
	f := t - math.Floor(t)
	st := (math.Floor(t) + s.smooth((f-0.35)/0.3)) * step
	return s.lerp(h, st, 0.6*k)
}

func (s VikingHarbor) raw(x, z float64) float64 {
	edge := math.Min(math.Min(x-vikingHarborX0, vikingHarborX1-x), math.Min(z-vikingHarborZ0, vikingHarborZ1-z))
	if edge <= 0 {
		return 0
	}
	fall := 0.08 + 0.92*s.smooth(edge/12)
	h := 0.0
	// West wall, lower toward the fjord mouth.
	if d := s.westShore(z) - x; d > 0 {
		crest := s.lerp(38, 18, s.smooth((z+15)/65)) * (0.72 + 0.45*s.ridge(z*0.055, x*0.04, 12))
		w := s.wall(d, crest, 0.5+0.5*s.fbm(z*0.05, 2.7, 14))
		for _, wz := range vikingHarborWestFalls {
			w -= s.gully(z-wz, d)
		}
		h = math.Max(h, s.rock(x, z, w)*fall)
	}
	// North wall behind the fjord head, falling to a hill behind the village.
	if d := s.backShore(x) - z; d > 0 {
		crest := s.lerp(38, 15, s.smooth((x+2)/26)) * (0.72 + 0.45*s.ridge(x*0.055, z*0.04, 13))
		beach := s.backBeach(x)
		w := 0.0
		if beach > 0.05 {
			w = 1.7 * s.smooth(beach/1.5) * math.Min(d/beach, 1)
		}
		w = math.Max(w, s.wall(d-beach, crest, 0.5+0.5*s.fbm(x*0.05, 5.3, 15)))
		w -= s.gully(x-vikingHarborBackFall, d-beach)
		h = math.Max(h, s.rock(x, z, w)*fall)
	}
	// Village plain behind a pebble beach, with a knoll for the church.
	if d := x - s.eastShore(z); d > 0 {
		w := 0.5 + (vikingHarborPlain-0.5)*s.smooth(d/6)
		w += 0.3 * s.fbm(x*0.09, z*0.09, 9) * s.smooth((d-6)/4)
		if k := P(x, z).Sub(vikingHarborChurch).Length(); k < 13 {
			w += 1.6 * s.smooth((13-k)/6)
		}
		w *= 0.45 + 0.55*s.smooth(edge/5)
		h = math.Max(h, w)
	}
	return h
}

// vikingHarborLand is the terrain height field rasterized for fast bilinear lookups.
type vikingHarborLand struct {
	h      []float64
	nx, nz int
}

func (s VikingHarbor) land() *vikingHarborLand {
	nx := int((vikingHarborX1-vikingHarborX0)/vikingHarborStep) + 2
	nz := int((vikingHarborZ1-vikingHarborZ0)/vikingHarborStep) + 2
	l := &vikingHarborLand{make([]float64, nx*nz), nx, nz}
	for j := 0; j < nz; j++ {
		for i := 0; i < nx; i++ {
			l.h[j*nx+i] = s.raw(vikingHarborX0+float64(i)*vikingHarborStep, vikingHarborZ0+float64(j)*vikingHarborStep)
		}
	}
	return l
}

func (l *vikingHarborLand) at(x, z float64) float64 {
	fx := (x - vikingHarborX0) / vikingHarborStep
	fz := (z - vikingHarborZ0) / vikingHarborStep
	if fx < 0 || fz < 0 || fx >= float64(l.nx-1) || fz >= float64(l.nz-1) {
		return 0
	}
	i, j := int(fx), int(fz)
	tx, tz := fx-float64(i), fz-float64(j)
	a, b := l.h[j*l.nx+i], l.h[j*l.nx+i+1]
	c, d := l.h[(j+1)*l.nx+i], l.h[(j+1)*l.nx+i+1]
	return (a*(1-tx)+b*tx)*(1-tz) + (c*(1-tx)+d*tx)*tz
}

func (l *vikingHarborLand) slope(x, z float64) float64 {
	dx := (l.at(x+0.6, z) - l.at(x-0.6, z)) / 1.2
	dz := (l.at(x, z+0.6) - l.at(x, z-0.6)) / 1.2
	return math.Hypot(dx, dz)
}

// shell fills the cells of a predicate that lie within t of its outside (above or
// sideways): a hollow skin, open underneath.
func (VikingHarbor) shell(g *Grid, box Box, t float64, in func(Vec3) bool) {
	g.Fill(Func{box, func(p Vec3) bool {
		if !in(p) {
			return false
		}
		return !(in(p.Add(V(t, 0, 0))) && in(p.Add(V(-t, 0, 0))) && in(p.Add(V(0, 0, t))) &&
			in(p.Add(V(0, 0, -t))) && in(p.Add(V(0, t, 0))))
	}})
}

// recolor visits every occupied cell and repaints those for which f returns true.
func (VikingHarbor) recolor(g *Grid, f func(c Cell, p Vec3, cur Color) (Color, bool)) {
	prev := g.Brush()
	for _, cell := range g.Cells() {
		cur, _ := g.ColorOf(cell)
		if next, ok := f(cell, g.World(cell), cur); ok && next != cur {
			g.Remove(cell)
			g.Use(next)
			g.Add(cell)
		}
	}
	g.Use(prev)
}

// vikingHarborFrame is an oriented plan frame: u along the heading, v to its left.
type vikingHarborFrame struct {
	C     Vec2
	Angle float64
}

func (f vikingHarborFrame) local(p Vec3) (u, v float64) {
	q := p.Plan().Sub(f.C).Rotate(-f.Angle)
	return q.X, q.Y
}

func (f vikingHarborFrame) at(u, v float64) Vec2    { return f.C.Add(P(u, v).Rotate(f.Angle)) }
func (f vikingHarborFrame) pt(u, y, v float64) Vec3 { return f.at(u, v).At(y) }

func (f vikingHarborFrame) box(r, y0, y1 float64) Box {
	return Box{V(f.C.X-r, y0, f.C.Y-r), V(f.C.X+r, y1, f.C.Y+r)}
}

// vikingHarborShip configures a longship (or, without heads and mast, a small boat).
type vikingHarborShip struct {
	C            Vec2
	Angle        float64 // Heading of the bow in plan.
	L, B         float64
	Strake       Color // Painted top strake.
	Mast, Sail   bool  // Mast with a furled sail on its yard, or a full sail.
	Oars         int   // 0 none, 1 rowing on both sides, 2 resting on the +v side.
	Shields      bool
	Heads        bool // Dragon prow and curled tail.
	Wake         bool
	Gangway      float64 // Side (+1/-1 in v) of a plank to a jetty; 0 none.
	GangwayReach float64
}

func (s VikingHarbor) ship(g *Grid, sh vikingHarborShip) {
	f := vikingHarborFrame{sh.C, sh.Angle}
	L, B := sh.L, sh.B
	D := 0.13 * L
	if D < 1.6 {
		D = 1.6
	}
	gun := func(u float64) float64 {
		k := u / (L / 2)
		return D + 0.09*L*k*k*k*k
	}
	strakes := 5.0
	halfw := func(u, y float64) float64 {
		k := u / (L / 2)
		if k <= -1 || k >= 1 || y < -0.4 {
			return -1
		}
		w := B / 2 * math.Pow(1-k*k, 0.55)
		t := math.Min(math.Max(y/(0.75*D), 0), 1)
		w *= 0.3 + 0.7*math.Sqrt(t)
		// Clinker: each strake's lower edge laps outside the one below.
		q := strakes * y / gun(u)
		w += 0.2 * (1 - (q - math.Floor(q)))
		return w
	}
	hullBox := f.box(L/2+1, -0.4, D+0.09*L+1)
	g.With(Wood, func() {
		g.FillVolume(Func{hullBox, func(p Vec3) bool {
			u, v := f.local(p)
			if p.Y > gun(u) {
				return false
			}
			w := halfw(u, p.Y)
			return w >= 0 && math.Abs(v) <= w
		}})
	})
	// near selects hull cells (grown slightly to catch volume-filled edge cells).
	near := func(p Vec3) bool {
		u, v := f.local(p)
		if p.Y > gun(u)+0.45 {
			return false
		}
		w := math.Max(halfw(math.Max(-L/2+0.01, math.Min(L/2-0.01, u)), math.Max(p.Y, -0.3)), 0)
		return math.Abs(u) <= L/2+0.4 && math.Abs(v) <= w+0.45
	}
	inner := func(u, y float64) float64 {
		a, b := halfw(u-1, y), halfw(u+1, y)
		if a < 0 || b < 0 {
			return -1
		}
		return math.Min(a, b) - 0.85
	}
	g.Carve(Func{hullBox, func(p Vec3) bool {
		u, v := f.local(p)
		return p.Y >= 0.9 && math.Abs(v) <= inner(u, math.Max(p.Y, 0.9))
	}})
	// Tarred bottom, a dark wale and the painted top strake.
	g.Paint(Charcoal, Func{hullBox, func(p Vec3) bool {
		u, v := f.local(p)
		return p.Y < 0.5 && near(p) && math.Abs(v) > inner(u, 0.9)+0.1
	}})
	g.Paint(Charcoal, Func{hullBox, func(p Vec3) bool {
		u, v := f.local(p)
		q := strakes * p.Y / gun(u)
		return near(p) && q >= 2 && q < 2.75 && math.Abs(v) > inner(u, p.Y)+0.3
	}})
	g.Paint(sh.Strake, Func{hullBox, func(p Vec3) bool {
		u, _ := f.local(p)
		return near(p) && p.Y > gun(u)-0.6
	}})
	// Rowing benches.
	for u := -0.32 * L; u <= 0.32*L; u += 2.2 {
		uu := u
		g.With(Wood, func() {
			g.Fill(Func{f.box(L/2, 0.8, D), func(p Vec3) bool {
				cu, v := f.local(p)
				return math.Abs(cu-uu) < 0.45 && p.Y >= 1.0 && p.Y <= 1.6 && math.Abs(v) <= inner(cu, 1.2)+0.3
			}})
		})
	}
	// Stem and stern posts.
	stem := func(pts []Vec2, r float64, gold int) {
		for i := 0; i+1 < len(pts); i++ {
			c := Wood
			if i >= len(pts)-1-gold {
				c = Gold
			}
			a, b := pts[i], pts[i+1]
			g.With(c, func() { g.Fill(Capsule{f.pt(a.X, a.Y, 0), f.pt(b.X, b.Y, 0), r}) })
		}
	}
	H := L / 2
	if sh.Heads {
		stem([]Vec2{P(H-1.5, 0.3), P(H-0.2, 2.2), P(H+0.6, 4.4), P(H+1.0, 6.4), P(H+0.9, 8.0), P(H+1.5, 9.3)}, 0.55, 2)
		stem([]Vec2{P(-H+1.5, 0.3), P(-H+0.2, 2.2), P(-H-0.6, 4.4), P(-H-1.0, 6.4), P(-H-0.8, 8.0),
			P(-H-0.1, 9.0), P(-H+0.6, 8.6), P(-H+0.7, 7.7)}, 0.5, 3)
		// Dragon head: gold skull and snout, red jaw and crest.
		hu, hy := H+2.5, 9.7
		g.With(Gold, func() {
			g.Fill(Func{f.box(L/2+5, hy-1.5, hy+1.5), func(p Vec3) bool {
				u, v := f.local(p)
				a, b, c := (u-hu)/1.35, (p.Y-hy)/0.8, v/0.75
				return a*a+b*b+c*c <= 1
			}})
			g.Fill(Capsule{f.pt(hu, hy-0.1, 0), f.pt(H+3.9, 9.4, 0), 0.5})
		})
		g.With(Red, func() {
			g.Fill(Capsule{f.pt(H+2.1, 9.0, 0), f.pt(H+3.7, 8.6, 0), 0.42})
			g.Fill(Capsule{f.pt(H+1.8, 10.2, 0), f.pt(H+1.0, 11.1, 0), 0.42})
		})
	} else {
		stem([]Vec2{P(H-1, 0.3), P(H+0.2, D+0.09*L+0.6)}, 0.45, 0)
		stem([]Vec2{P(-H+1, 0.3), P(-H-0.2, D+0.09*L+0.6)}, 0.45, 0)
	}
	// Shields hung along the top strake.
	if sh.Shields {
		cols := []Color{Red, Gold, Navy, White}
		n := 0
		for ui := -0.33 * L; ui <= 0.33*L+0.01; ui += 1.6 {
			y := gun(ui) - 0.55
			w := halfw(ui, y) + 0.2
			for side := -1.0; side <= 1; side += 2 {
				col := cols[(n+int(side+1))%4]
				sd, u0 := side, ui
				g.With(col, func() {
					g.Fill(Func{Box{f.pt(u0, y, sd*w).Sub(V(1.4, 1.2, 1.4)), f.pt(u0, y, sd*w).Add(V(1.4, 1.2, 1.4))},
						func(p Vec3) bool {
							u, v := f.local(p)
							return math.Abs(v*sd-w) < 0.4 && math.Hypot(u-u0, p.Y-y) < 0.76
						}})
				})
			}
			n++
		}
	}
	// Oars.
	if sh.Oars > 0 {
		for ui := -0.33*L + 0.8; ui <= 0.33*L; ui += 1.6 {
			for side := -1.0; side <= 1; side += 2 {
				if sh.Oars == 2 && side < 0 {
					continue
				}
				y0 := gun(ui) - 0.9
				w := halfw(ui, y0)
				a := f.pt(ui, y0+0.3, side*(w-0.9))
				var b Vec3
				if sh.Oars == 1 {
					b = f.pt(ui-1.6, 0.25, side*(w+6.5))
				} else {
					b = f.pt(ui+0.4, 0.3, side*(w+5.5))
				}
				g.With(Wood, func() {
					g.Fill(Capsule{a, b, 0.42})
					g.Fill(Capsule{a.Add(b.Sub(a).Scale(0.8)), b, 0.6})
				})
			}
		}
	}
	// Steering oar on the starboard quarter.
	if sh.Heads {
		u := -0.36 * L
		w := halfw(u, gun(u)-0.5)
		top, bot := f.pt(u, gun(u)+1.4, -(w+0.2)), f.pt(u-0.6, -0.1, -(w+0.9))
		g.With(Wood, func() {
			g.Fill(Capsule{top, bot, 0.45})
			g.Fill(Capsule{top.Add(bot.Sub(top).Scale(0.6)), bot, 0.65})
		})
	}
	// Mast and sail.
	if sh.Mast {
		mu := 0.04 * L
		mc := f.at(mu, 0)
		if sh.Sail {
			yard, foot, half := 14.2, 5.2, 5.6
			g.With(Wood, func() {
				g.Fill(Cylinder{Center: mc, Radius: 0.45, Y0: 0.5, Y1: 15.6})
				g.Fill(Capsule{f.pt(mu+0.5, yard+0.3, -half-0.6), f.pt(mu+0.5, yard+0.3, half+0.6), 0.42})
			})
			bulge := func(v, y float64) float64 {
				k := v / half
				return mu + 0.8 + 1.9*(1-k*k)*(0.25+0.75*math.Sin(math.Pi*(yard-y)/(yard-foot+1.5)))
			}
			sail := Func{f.box(half+3, foot-0.5, yard+0.5), func(p Vec3) bool {
				u, v := f.local(p)
				return math.Abs(v) <= half && p.Y >= foot && p.Y <= yard && math.Abs(u-bulge(v, p.Y)) < 0.45
			}}
			g.With(White, func() { g.Fill(sail) })
			g.Paint(Red, Func{sail.Box, func(p Vec3) bool {
				_, v := f.local(p)
				return sail.F(p) && int(math.Floor((v+half)/1.4))%2 == 0
			}})
			// Sheets from the sail's foot to the hull.
			for _, sd := range []float64{-1, 1} {
				g.With(Wood, func() {
					g.Fill(Capsule{f.pt(bulge(sd*half, foot), foot, sd*half), f.pt(-0.18*L, gun(-0.18*L), sd*(B/2-0.3)), 0.32})
				})
			}
		} else {
			yard := 10.0
			g.With(Wood, func() {
				g.Fill(Cylinder{Center: mc, Radius: 0.45, Y0: 0.5, Y1: 11.4})
				g.Fill(Capsule{f.pt(mu+0.5, yard, -5.4), f.pt(mu+0.5, yard, 5.4), 0.4})
			})
			bundle := Capsule{f.pt(mu+0.6, yard-0.6, -4.6), f.pt(mu+0.6, yard-0.6, 4.6), 0.75}
			g.With(White, func() { g.Fill(bundle) })
			g.Paint(Red, Func{bundle.Bounds(), func(p Vec3) bool {
				_, v := f.local(p)
				return bundle.Contains(p) && int(math.Floor((v+5.6)/1.4))%2 == 0
			}})
		}
		// Fore and back stays.
		g.With(Wood, func() {
			top := 11.4
			if sh.Sail {
				top = 15.6
			}
			g.Fill(Capsule{mc.At(top - 0.4), f.pt(H-0.6, gun(H-0.6), 0), 0.32})
			g.Fill(Capsule{mc.At(top - 0.4), f.pt(-H+0.6, gun(-H+0.6), 0), 0.32})
		})
	}
	// Gangway to a jetty.
	if sh.Gangway != 0 {
		a := f.pt(-0.1*L, gun(0)-0.2, sh.Gangway*(B/2-0.5))
		b := f.pt(-0.1*L, 1.8, sh.Gangway*(B/2+sh.GangwayReach))
		g.With(Wood, func() { g.Fill(Capsule{a, b, 0.5}) })
	}
	// Foam: bow wave and wake on the water layer.
	water := f.box(L/2+22, -0.25, 0.25)
	if sh.Wake {
		g.Paint(White, Func{water, func(p Vec3) bool {
			u, v := f.local(p)
			c := Nearest(p.Scale(1 / g.Span()))
			r := Hash01(int(c.X), int(c.Z), 41)
			if u < -H+1 && u > -H-15 {
				t := -H + 1 - u
				return math.Abs(math.Abs(v)-(0.32*t+1.0)) < 0.5+0.03*t && r < 0.7-0.04*t
			}
			if u > 0.15*L && u < H+1.5 {
				w := halfw(math.Min(u, H-0.5), 0.3)
				return math.Abs(v) < w+1.1 && math.Abs(v) > w-0.1 && r < 0.7
			}
			return false
		}})
	}
	if sh.Oars == 1 {
		g.Paint(White, Func{water, func(p Vec3) bool {
			u, v := f.local(p)
			c := Nearest(p.Scale(1 / g.Span()))
			if Hash01(int(c.X), int(c.Z), 43) > 0.6 {
				return false
			}
			for ui := -0.33*L + 0.8; ui <= 0.33*L; ui += 1.6 {
				w := halfw(ui, gun(ui)-0.9)
				if math.Hypot(u-(ui-1.6), math.Abs(v)-(w+6.5)) < 1.4 {
					return true
				}
			}
			return false
		}})
	}
}

// vikingHarborHouse configures a bow-walled longhouse with a turf roof.
type vikingHarborHouse struct {
	C             Vec2
	Angle         float64 // Direction of the ridge.
	Length, Width float64
	Wall, Roof    float64 // Wall height and ridge rise above the eaves.
	Walls, Gables Color
	Dragons       bool    // Gold dragon heads on the crossed gable boards.
	Door          float64 // Door position along the ridge on the +v side.
	Posts         bool    // Sloping buttress posts along the walls.
	Smoke         int     // Puffs from the smoke hole.
	Stone         bool    // Stone lower walls.
}

func (s VikingHarbor) house(g *Grid, land *vikingHarborLand, h vikingHarborHouse) {
	f := vikingHarborFrame{h.C, h.Angle}
	L, W := h.Length, h.Width
	base, low := 0.0, math.Inf(1)
	for i := -1.0; i <= 1; i++ {
		for j := -1.0; j <= 1; j++ {
			q := f.at(i*L/2, j*W/2)
			t := land.at(q.X, q.Y)
			base, low = math.Max(base, t), math.Min(low, t)
		}
	}
	base = math.Round(base*2) / 2
	hw := func(u float64) float64 {
		k := 2 * math.Max(-L/2, math.Min(L/2, u)) / L
		return W / 2 * (1 - 0.14*k*k)
	}
	const ov = 0.9
	eave := base + h.Wall
	ridge := func(u float64) float64 {
		k := 2 * u / (L + 2*ov)
		return eave + h.Roof + 1.1*k*k
	}
	roofTop := func(u, v float64) float64 {
		e := hw(u) + ov
		return ridge(u) - (ridge(u)-(eave-0.5))*math.Abs(v)/e
	}
	box := f.box(math.Hypot(L/2+ov+3, W/2+ov+3), low-1.6, eave+h.Roof+4)
	g.With(Stone, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			u, v := f.local(p)
			return math.Abs(u) <= L/2+0.4 && math.Abs(v) <= hw(u)+0.4 && p.Y >= low-1.2 && p.Y <= base+0.35
		}})
	})
	const t = 1.25
	g.With(h.Walls, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			u, v := f.local(p)
			if math.Abs(u) > L/2 || math.Abs(v) > hw(u) || p.Y < base || p.Y > eave+0.2 {
				return false
			}
			return !(math.Abs(u) <= L/2-t && math.Abs(v) <= hw(u)-t && p.Y > base+0.4)
		}})
		// Gable-end walls.
		g.Fill(Func{box, func(p Vec3) bool {
			u, v := f.local(p)
			return math.Abs(u) <= L/2 && math.Abs(u) >= L/2-t && math.Abs(v) <= hw(u) &&
				p.Y >= eave && p.Y <= roofTop(u, v)-0.4
		}})
	})
	if h.Stone {
		lower := func(p Vec3) bool {
			u, v := f.local(p)
			return p.Y >= base && p.Y <= base+1.3 && math.Abs(u) <= L/2+0.1 && math.Abs(v) <= hw(u)+0.1
		}
		g.Paint(DarkStone, Func{box, lower})
		g.Paint(Stone, Func{box, func(p Vec3) bool {
			c := Nearest(p.Scale(1 / g.Span()))
			return lower(p) && Hash01(int(c.X), int(c.Y), int(c.Z), 51) < 0.4
		}})
	}
	// Turf roof.
	inRoof := func(p Vec3) bool {
		u, v := f.local(p)
		if math.Abs(u) > L/2+ov || math.Abs(v) > hw(u)+ov {
			return false
		}
		top := roofTop(u, v)
		return p.Y <= top && p.Y >= top-1.5 && p.Y >= eave-0.6
	}
	g.With(Grass, func() { g.Fill(Func{box, inRoof}) })
	g.Paint(Foliage, Func{box, func(p Vec3) bool {
		u, v := f.local(p)
		return inRoof(p) && s.noise(u*0.45, v*0.45, 61+int(h.C.X)) > 0.45
	}})
	g.Paint(Wood, Func{box, func(p Vec3) bool {
		u, v := f.local(p)
		return inRoof(p) && (math.Abs(v) > hw(u)+ov-0.55 || math.Abs(u) > L/2+ov-0.55)
	}})
	// Dark interior seen through the doors.
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		u, v := f.local(p)
		return math.Abs(u) <= L/2-0.85 && math.Abs(v) <= hw(u)-0.85 && p.Y >= base-0.3 && p.Y <= roofTop(u, v)-1.1
	}})
	// Doors: one in the +v long wall, one in the -u gable for long halls.
	g.Carve(Func{box, func(p Vec3) bool {
		u, v := f.local(p)
		return math.Abs(u-h.Door) < 1.0 && v > hw(u)-1.8 && v < hw(u)+0.6 && p.Y >= base+0.4 && p.Y <= base+2.5
	}})
	g.With(h.Gables, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			u, v := f.local(p)
			d := math.Abs(u - h.Door)
			return d >= 1.0 && d < 1.6 && v > hw(u)-0.6 && v < hw(u)+0.4 && p.Y >= base+0.4 && p.Y <= base+2.9
		}})
	})
	if L > 24 {
		g.Carve(Func{box, func(p Vec3) bool {
			u, v := f.local(p)
			return u < -L/2+1.8 && u > -L/2-0.6 && math.Abs(v) < 1.2 && p.Y >= base+0.4 && p.Y <= base+2.8
		}})
	}
	// Crossed gable boards with optional dragon heads.
	for _, sd := range []float64{-1, 1} {
		ue := sd * (L/2 + ov - 0.2)
		e := hw(ue) + ov
		peak := ridge(ue) + 0.3
		for _, side := range []float64{-1, 1} {
			a := V(0, eave-0.3, side*e)
			b := V(0, peak, 0)
			dir := b.Sub(a)
			dir = dir.Scale(1 / dir.Length())
			end := b.Add(dir.Scale(2.2))
			A, E := f.pt(ue, a.Y, a.Z), f.pt(ue, end.Y, end.Z)
			g.With(h.Gables, func() { g.Fill(Capsule{A, E, 0.45}) })
			if h.Dragons {
				tip := f.pt(ue+sd*0.6, end.Y+0.4, end.Z+dir.Z*0.6)
				g.With(Gold, func() {
					g.Fill(Sphere(tip, 0.75))
					g.Fill(Capsule{tip, f.pt(ue+sd*1.6, end.Y+0.2, end.Z+dir.Z*1.2), 0.45})
				})
			}
		}
	}
	// Sloping posts braced against the walls.
	if h.Posts {
		for u := -L/2 + 2.2; u <= L/2-2.2; u += 2.6 {
			for _, side := range []float64{-1, 1} {
				if side > 0 && math.Abs(u-h.Door) < 2 {
					continue
				}
				g.With(Wood, func() {
					g.Fill(Capsule{f.pt(u, base-0.2, side*(hw(u)+1.9)), f.pt(u, eave-0.2, side*(hw(u)+0.3)), 0.45})
				})
			}
		}
	}
	// Smoke hole and smoke.
	su := 0.15 * L
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		u, v := f.local(p)
		return math.Abs(u-su) < 0.9 && math.Abs(v) < 0.9 && p.Y > roofTop(u, v)-0.8
	}})
	s.smoke(g, f.pt(su, ridge(su)+0.6, 0), h.Smoke, int(h.C.X*7+h.C.Y))
}

// smoke stacks n drifting puffs from a vent, dark near the source and pale above.
func (VikingHarbor) smoke(g *Grid, at Vec3, n, seed int) {
	p := at
	for i := 0; i < n; i++ {
		r := 0.6 + 0.14*float64(i)
		c := Charcoal
		switch {
		case i >= 5:
			c = White
		case i >= 3:
			c = Stone
		case i >= 1:
			c = DarkStone
		}
		g.With(c, func() { g.Fill(Sphere(p, r)) })
		p = p.Add(V(0.8+0.6*Hash01(i, seed), 0.9+0.12*float64(i), 0.6*(Hash01(i, seed, 1)-0.5)))
	}
}

// pine is a conifer of stacked cones.
func (VikingHarbor) pine(g *Grid, c Vec2, y0, height float64, seed int) {
	r := 0.32 * height
	g.With(Wood, func() { g.Fill(Cylinder{Center: c, Radius: 0.45, Y0: y0 - 0.8, Y1: y0 + 1.6}) })
	g.With(Foliage, func() {
		tiers := 3
		for i := 0; i < tiers; i++ {
			k := float64(i) / float64(tiers)
			b := y0 + 1.0 + k*(height-1.6)
			rr := r * (1 - 0.28*float64(i))
			g.Fill(Cone(c, rr, b, height*0.48))
		}
	})
}

// church is a stave church: arcaded gallery with a lean-to roof, a steep-roofed nave,
// chancel and apse, and a central tower with a skirt roof, belfry and spire.
func (s VikingHarbor) church(g *Grid, land *vikingHarborLand) {
	c := vikingHarborChurch
	base := math.Round(land.at(c.X, c.Y)*2) / 2
	at := func(x, y, z float64) Vec3 { return V(c.X+x, base+y, c.Y+z) }
	box := Box{at(-9, -2, -8), at(11, 30, 8)}
	rel := func(p Vec3) (float64, float64, float64) { return p.X - c.X, p.Y - base, p.Z - c.Y }
	g.With(Stone, func() { g.Fill(Block(at(-6.8, -2, -6.3), at(10, 0.4, 6.3))) })
	// Gallery walls and lean-to roof.
	g.With(Wood, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			x, y, z := rel(p)
			if y < 0.4 || y > 2.6 || math.Abs(x) > 6 || math.Abs(z) > 5.5 {
				return false
			}
			return math.Abs(x) > 5.1 || math.Abs(z) > 4.6
		}})
	})
	// Arcade openings.
	g.Carve(Func{box, func(p Vec3) bool {
		x, y, z := rel(p)
		if y < 1.0 || y > 2.1 {
			return false
		}
		if math.Abs(x) > 5.0 && math.Abs(z) < 4.5 {
			k := z/1.5 - math.Round(z/1.5)
			return math.Abs(k) < 0.28 && !(x < 0 && math.Abs(z) < 1.2)
		}
		if math.Abs(z) > 4.5 && math.Abs(x) < 5.0 {
			k := x/1.5 - math.Round(x/1.5)
			return math.Abs(k) < 0.28
		}
		return false
	}})
	cheb := func(x, z, hx, hz float64) float64 { return math.Max(math.Abs(x)-hx, math.Abs(z)-hz) }
	g.With(Charcoal, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			x, y, z := rel(p)
			d := cheb(x, z, 3.5, 3.0)
			if d < 0 || d > 3.3 {
				return false
			}
			top := 4.8 - 2.4*d/3.3
			return y <= top && y >= top-0.9
		}})
	})
	// Nave.
	g.With(Wood, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			x, y, z := rel(p)
			if y < 0.4 || y > 7.4 || math.Abs(x) > 3.5 || math.Abs(z) > 3 {
				return false
			}
			return math.Abs(x) > 2.4 || math.Abs(z) > 1.9 || y < 0.8
		}})
		g.Fill(Func{box, func(p Vec3) bool { // Gable triangles.
			x, y, z := rel(p)
			return math.Abs(x) <= 3.5 && math.Abs(x) >= 2.6 && y >= 7.2 && y <= 7.2+1.8*(3-math.Abs(z))
		}})
	})
	steep := func(hx, hz, y0, slope float64) func(x, y, z float64) bool {
		return func(x, y, z float64) bool {
			if math.Abs(x) > hx {
				return false
			}
			top := y0 + slope*(hz-math.Abs(z))
			return y <= top && y >= top-1.0 && y >= y0-0.4
		}
	}
	nave := steep(4.2, 3.8, 7.0, 1.8)
	g.With(Charcoal, func() { g.Fill(Func{box, func(p Vec3) bool { x, y, z := rel(p); return nave(x, y, z) }}) })
	// Chancel with its own steep roof, and a round apse.
	g.With(Wood, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			x, y, z := rel(p)
			return x >= 3.5 && x <= 7.2 && math.Abs(z) <= 2.4 && y >= 0.4 && y <= 5.4+1.8*(2.4-math.Abs(z))-0.6
		}})
		g.Fill(Cylinder{Center: P(c.X+7.2, c.Y), Radius: 2.1, Y0: base + 0.4, Y1: base + 4.6})
	})
	chancel := steep(3.9, 3.0, 5.2, 1.8)
	g.With(Charcoal, func() {
		g.Fill(Func{box, func(p Vec3) bool { x, y, z := rel(p); return x >= 3.3 && x <= 7.6 && chancel(x, y, z) }})
		g.Fill(Cone(P(c.X+7.2, c.Y), 2.8, base+4.4, 5.2))
		g.Fill(Cone(P(c.X+5.4, c.Y), 1.3, base+9.6, 5))
	})
	g.With(Wood, func() { g.Fill(BlockAt(P(c.X+5.4, c.Y), 1.8, 1.8, base+8, base+10)) })
	// Central tower: shaft, skirt roof, belfry and spire.
	g.With(Wood, func() { g.Fill(BlockAt(c, 3.2, 3.2, base+11, base+17.6)) })
	g.With(Charcoal, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			x, y, z := rel(p)
			d := cheb(x, z, 1.6, 1.6)
			if d < 0 || d > 1.8 {
				return false
			}
			top := 15.6 - 1.3*d
			return y <= top && y >= top-0.9
		}})
		g.Fill(Func{box, func(p Vec3) bool {
			x, y, z := rel(p)
			d := math.Max(math.Abs(x), math.Abs(z))
			top := 17.6 + 3.6*(2.3-d)
			return y >= 17.4 && y <= top && d <= 2.3
		}})
	})
	g.Carve(Func{box, func(p Vec3) bool {
		x, y, z := rel(p)
		return y >= 15.9 && y <= 17.2 && ((math.Abs(x) < 0.55 && math.Abs(z) < 2) || (math.Abs(z) < 0.55 && math.Abs(x) < 2))
	}})
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		x, y, z := rel(p)
		return y >= 11 && y <= 17.4 && math.Abs(x) < 1.0 && math.Abs(z) < 1.0
	}})
	g.With(Gold, func() {
		g.Fill(Sphere(at(0, 26.3, 0), 0.6))
		g.Fill(Capsule{at(0, 26.3, 0), at(0, 28.2, 0), 0.35})
		g.Fill(Capsule{at(0, 27.5, -0.8), at(0, 27.5, 0.8), 0.32})
		// Dragon heads on the gable peaks.
		for _, sx := range []float64{-1, 1} {
			g.Fill(Capsule{at(sx*3.8, 13.3, 0), at(sx*5.7, 14.8, 0), 0.45})
			g.Fill(Sphere(at(sx*5.9, 15.0, 0), 0.6))
		}
		g.Fill(Capsule{at(7.4, 9.0, 0), at(9.4, 10.3, 0), 0.4})
		for _, sx := range []float64{-1, 1} {
			for _, sz := range []float64{-1, 1} {
				g.Fill(Capsule{at(sx*3.2, 15.0, sz*3.2), at(sx*4.0, 15.9, sz*4.0), 0.38})
			}
		}
	})
	// Dark interior and the west portal.
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		x, y, z := rel(p)
		inGallery := math.Abs(x) < 5.2 && math.Abs(z) < 4.7 && y < 3.6
		inNave := math.Abs(x) < 2.5 && math.Abs(z) < 2.0 && y < 11
		return y > 0 && (inGallery || inNave)
	}})
	g.Carve(Func{box, func(p Vec3) bool {
		x, y, z := rel(p)
		return x < -2.3 && x > -6.5 && math.Abs(z) < 0.8 && y >= 0.5 && y <= 2.4+0.4*(1-z*z/0.64)
	}})
	g.With(Red, func() { g.Fill(Block(at(-2.6, 0.5, -0.8), at(-2.2, 2.6, 0.8))) })
}

// smithy is a low stone forge with an open front, a glowing hearth and smoke.
func (s VikingHarbor) smithy(g *Grid, land *vikingHarborLand) {
	c := vikingHarborSmithy
	base := math.Round(land.at(c.X, c.Y)*2) / 2
	f := vikingHarborFrame{c, 0}
	box := f.box(8, base-2, base+9)
	g.With(DarkStone, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			u, v := f.local(p)
			if math.Abs(u) > 4 || math.Abs(v) > 3 || p.Y < base-1.5 || p.Y > base+2.6 {
				return false
			}
			return math.Abs(u) > 2.9 || v < -1.9 || p.Y < base+0.4
		}})
	})
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		cc := Nearest(p.Scale(1 / g.Span()))
		u, v := f.local(p)
		return math.Abs(u) <= 4 && math.Abs(v) <= 3 && p.Y >= base && Hash01(int(cc.X), int(cc.Y), int(cc.Z), 71) < 0.4
	}})
	// Lean-to turf roof sloping to the open +Z front, on two posts.
	g.With(Wood, func() {
		for _, x := range []float64{-3.4, 3.4} {
			g.Fill(Cylinder{Center: P(c.X+x, c.Y+3.2), Radius: 0.45, Y0: base, Y1: base + 3})
		}
	})
	g.With(Grass, func() {
		g.Fill(Func{box, func(p Vec3) bool {
			u, v := f.local(p)
			top := base + 4.4 - 0.3*(v+3.6)
			return math.Abs(u) <= 4.8 && math.Abs(v) <= 4.0 && p.Y <= top && p.Y >= top-1.0
		}})
	})
	// Hearth, glow, anvil and chimney.
	g.With(Stone, func() { g.Fill(BlockAt(P(c.X-1.4, c.Y-1.2), 2.4, 1.6, base+0.4, base+1.6)) })
	g.With(Red, func() { g.Fill(BlockAt(P(c.X-1.4, c.Y-1.0), 1.4, 0.9, base+1.4, base+1.9)) })
	g.With(Gold, func() { g.Fill(Sphere(V(c.X-1.4, base+2.0, c.Y-0.9), 0.5)) })
	g.With(Charcoal, func() { g.Fill(BlockAt(P(c.X+1.4, c.Y+0.4), 1.6, 0.8, base+0.4, base+1.6)) })
	g.With(DarkStone, func() { g.Fill(BlockAt(P(c.X-1.4, c.Y-2.2), 1.6, 1.6, base+1.6, base+6.2)) })
	s.smoke(g, V(c.X-1.2, base+6.9, c.Y-2.2), 7, 5)
	// Anvil-side wood pile.
	g.With(Wood, func() {
		for i := 0; i < 3; i++ {
			y := base + 0.5 + 0.7*float64(i)
			g.Fill(Capsule{V(c.X+4.6, y, c.Y-2.0+0.3*float64(i)), V(c.X+4.6, y, c.Y+2.0-0.3*float64(i)), 0.4})
		}
	})
}

// fishRack is a drying rack (hjell) of posts and rails hung with stockfish.
func (s VikingHarbor) fishRack(g *Grid, land *vikingHarborLand, a, b Vec2) {
	dir := b.Sub(a)
	n := dir.Scale(1 / dir.Length())
	side := P(-n.Y, n.X)
	l := dir.Length()
	for t := 0.0; t <= l+0.01; t += l / math.Ceil(l/3.2) {
		for _, sd := range []float64{-1, 1} {
			q := a.Add(n.Scale(t)).Add(side.Scale(sd * 0.9))
			y := land.at(q.X, q.Y)
			g.With(Wood, func() { g.Fill(Cylinder{Center: q, Radius: 0.42, Y0: y - 0.6, Y1: y + 3.6}) })
		}
	}
	y := land.at(a.X, a.Y) + 3.6
	y2 := land.at(b.X, b.Y) + 3.6
	for _, sd := range []float64{-1, 1} {
		o := side.Scale(sd * 0.9)
		g.With(Wood, func() { g.Fill(Capsule{a.Add(o).At(y), b.Add(o).At(y2), 0.4}) })
		for t := 0.6; t < l-0.3; t += 0.85 {
			q := a.Add(n.Scale(t)).Add(o)
			yy := y + (y2-y)*t/l
			c := []Color{Stone, Sand, White}[int(Hash01(int(t*10), int(sd), int(a.X))*3)]
			g.With(c, func() { g.Fill(Capsule{q.At(yy - 0.4), q.At(yy - 1.7), 0.4}) })
		}
	}
}

// palisade is a ring of pointed logs with a gatehouse of two towers.
func (s VikingHarbor) palisade(g *Grid, land *vikingHarborLand) {
	pts := vikingHarborPalisade
	gate := vikingHarborGate
	for i := 0; i+1 < len(pts); i++ {
		a, b := pts[i], pts[i+1]
		d := b.Sub(a)
		l := d.Length()
		for t := 0.0; t < l; t += 0.85 {
			q := a.Add(d.Scale(t / l))
			if q.Sub(gate).Length() < 5.2 {
				continue
			}
			y := land.at(q.X, q.Y)
			k := Nearest(q.At(0).Scale(1 / g.Span()))
			top := y + 3.0 + 0.6*Hash01(int(k.X), int(k.Z), 81)
			g.With(Wood, func() {
				g.Fill(Cylinder{Center: q, Radius: 0.55, Y0: y - 1.0, Y1: top})
				g.Fill(Cone(q, 0.55, top, 0.9))
			})
		}
	}
	// Gate towers either side of the opening, with a walkway over the gate.
	base := math.Round(land.at(gate.X, gate.Y)*2) / 2
	dir := P(1, 0)
	for _, sd := range []float64{-1, 1} {
		c := gate.Add(dir.Scale(sd * 4.2))
		g.With(Wood, func() {
			g.Fill(Subtract(BlockAt(c, 3.4, 3.4, base-1, base+7.5), BlockAt(c, 1.6, 1.6, base+0.4, base+7)))
			g.Fill(BlockAt(c, 4.2, 4.2, base+7.0, base+7.6))
		})
		g.With(Charcoal, func() {
			g.Fill(Func{Box{V(c.X-2.4, base+7.5, c.Y-2.4), V(c.X+2.4, base+11, c.Y+2.4)}, func(p Vec3) bool {
				d := math.Max(math.Abs(p.X-c.X), math.Abs(p.Z-c.Y))
				return d <= 2.3 && p.Y <= base+7.6+1.4*(2.3-d)
			}})
		})
		g.With(Red, func() { g.Fill(Sphere(V(c.X, base+11.0, c.Y), 0.55)) })
		g.Carve(BlockAt(P(c.X, c.Y+1.7), 1.2, 1.0, base+4.5, base+5.8))
	}
	g.With(Wood, func() {
		g.Fill(Block(V(gate.X-2.6, base+5.2, gate.Y-1.2), V(gate.X+2.6, base+6.0, gate.Y+1.2)))
		for x := gate.X - 2.4; x <= gate.X+2.41; x += 0.8 {
			g.Fill(Cylinder{Center: P(x, gate.Y+1.0), Radius: 0.4, Y0: base + 6, Y1: base + 7.3})
		}
	})
	// Open gate leaves.
	g.With(Wood, func() {
		g.Fill(Block(V(gate.X-2.5, base, gate.Y+0.2), V(gate.X-2.0, base+4.6, gate.Y+2.8)))
		g.Fill(Block(V(gate.X+2.0, base, gate.Y+0.2), V(gate.X+2.5, base+4.6, gate.Y+2.8)))
	})
	g.With(Red, func() {
		g.Fill(Block(V(gate.X-2.5, base+1.5, gate.Y+0.2), V(gate.X-2.0, base+2.1, gate.Y+2.8)))
		g.Fill(Block(V(gate.X+2.0, base+1.5, gate.Y+0.2), V(gate.X+2.5, base+2.1, gate.Y+2.8)))
	})
}

// jetty is a plank deck on posts running from the shore out along -X.
func (s VikingHarbor) jetty(g *Grid, z, x0, x1 float64) {
	g.With(Wood, func() {
		g.Fill(Block(V(x0, 1.2, z-1.6), V(x1, 1.8, z+1.6)))
		for x := x0 + 0.5; x <= x1; x += 3 {
			for _, sz := range []float64{-1.6, 1.6} {
				g.With(DarkStone, func() { g.Fill(Cylinder{Center: P(x, z+sz), Radius: 0.5, Y0: -0.2, Y1: 2.7}) })
			}
		}
	})
}

// sheep is a woolly body with a dark head facing angle a.
func (VikingHarbor) sheep(g *Grid, c Vec2, y, a float64) {
	f := vikingHarborFrame{c, a}
	g.With(White, func() {
		g.Fill(Func{f.box(2, y, y+2), func(p Vec3) bool {
			u, v := f.local(p)
			q, r, t := u/1.0, (p.Y-y-1.0)/0.62, v/0.62
			return q*q+r*r+t*t <= 1
		}})
	})
	g.With(Charcoal, func() {
		g.Fill(Sphere(f.pt(1.15, y+1.3, 0), 0.42))
		for _, u := range []float64{-0.6, 0.6} {
			g.Fill(Block(f.pt(u, y-0.2, 0).Sub(V(0.2, 0, 0.2)), f.pt(u, y+0.5, 0).Add(V(0.2, 0, 0.2))))
		}
	})
}

func (s VikingHarbor) Build(g *Grid) Order {
	land := s.land()
	H := land.at
	site := Box{V(vikingHarborX0, -0.4, vikingHarborZ0), V(vikingHarborX1, 40, vikingHarborZ1)}
	water := Box{V(vikingHarborX0, -0.2, vikingHarborZ0), V(vikingHarborX1, 0.2, vikingHarborZ1)}

	// Fjord water: one layer.
	g.Use(Water).Fill(Func{water, func(p Vec3) bool { return H(p.X, p.Z) < 0.45 }})
	// Land: hollow skin over the height field.
	g.Use(DarkStone)
	s.shell(g, site, 0.9, func(p Vec3) bool {
		h := H(p.X, p.Z)
		return h >= 0.45 && p.Y >= -0.3 && p.Y <= h
	})

	// Waterfall geometry: lateral offset and inland distance for each fall.
	falls := func(x, z float64) (lat, d float64, ok bool) {
		for _, wz := range vikingHarborWestFalls {
			if l := z - wz; math.Abs(l) < 1.6 {
				return l, s.westShore(z) - x, true
			}
		}
		if l := x - vikingHarborBackFall; math.Abs(l) < 1.6 {
			return l, s.backShore(x) - z, true
		}
		return 0, 0, false
	}

	// Terrain colors.
	s.recolor(g, func(c Cell, p Vec3, cur Color) (Color, bool) {
		r := Hash01(int(c.X), int(c.Y), int(c.Z), 17)
		if cur == Water {
			// Pale shallows along every shore.
			for k := 0; k < 8; k++ {
				q := Polar(p.Plan(), 2.2, float64(k)*math.Pi/4)
				if H(q.X, q.Y) >= 0.45 {
					if r < 0.6 {
						return Glass, true
					}
					return Water, false
				}
			}
			if r < 0.025 {
				return Glass, true
			}
			return Water, false
		}
		h := H(p.X, p.Z)
		sl := land.slope(p.X, p.Z)
		top := p.Y > h-0.95
		east := p.X > s.eastShore(p.Z)-0.5 && p.X < s.eastShore(p.Z)+7
		switch {
		case top && h < 2.1 && sl < 1.2 && (east || (p.X < 10 && p.Z > s.backShore(p.X)-s.backBeach(p.X)-0.5 && p.Z < s.backShore(p.X)+2)):
			// Pebble beach.
			switch {
			case r < 0.35:
				return Stone, true
			case r < 0.6:
				return DarkStone, true
			case r < 0.78:
				return Slate, true
			}
			return Sand, true
		case top && h > 32+3*s.fbm(p.X*0.08, p.Z*0.08, 37) && sl < 1.5:
			if r < 0.85 {
				return White, true
			}
			return Stone, true
		case top && sl < 0.75:
			if r < 0.1 {
				return Foliage, true
			}
			return Grass, true
		case top && sl < 1.25:
			if r < 0.55 {
				return Foliage, true
			}
			return Grass, true
		}
		// Rock strata: dark gneiss with lighter and bluish bands.
		band := int(math.Floor((p.Y + 0.12*p.X - 0.08*p.Z + 2.2*s.fbm(p.X*0.09, p.Z*0.09, 31)) / 1.5))
		k := Hash01(band, 3)
		switch {
		case k < 0.3:
			return Stone, true
		case k < 0.5:
			return Slate, true
		}
		return DarkStone, true
	})

	// Waterfalls: a water skin over each gully, white where it falls steeply, with
	// foam where it meets the fjord.
	fallBox := site
	inFall := func(p Vec3) bool {
		lat, d, ok := falls(p.X, p.Z)
		if !ok || d < -0.5 || d > 19 {
			return false
		}
		h := H(p.X, p.Z)
		w := 1.25 + 0.15*math.Sin(d)
		return math.Abs(lat) < w && p.Y >= h-0.5 && p.Y <= h+0.7 && h > 0.3
	}
	g.With(Water, func() { g.Fill(Func{fallBox, inFall}) })
	g.Paint(White, Func{fallBox, func(p Vec3) bool {
		if !inFall(p) {
			return false
		}
		c := Nearest(p.Scale(1 / g.Span()))
		sl := land.slope(p.X, p.Z)
		r := Hash01(int(c.X), int(c.Y), int(c.Z), 19)
		return (sl > 1.4 && r < 0.75) || r < 0.15
	}})
	foam := []Vec2{P(vikingHarborBackFall, s.backShore(vikingHarborBackFall)+0.5)}
	for _, wz := range vikingHarborWestFalls {
		foam = append(foam, P(s.westShore(wz)+0.5, wz))
	}
	g.Paint(White, Func{water, func(p Vec3) bool {
		c := Nearest(p.Scale(1 / g.Span()))
		r := Hash01(int(c.X), int(c.Z), 23)
		for _, q := range foam {
			if d := p.Plan().Sub(q).Length(); d < 4.5 && r < 1-d/5 {
				return true
			}
		}
		return false
	}})

	// Village paths of trodden earth.
	paths := [][2]Vec2{
		{vikingHarborGate, P(32, 54)}, {vikingHarborGate, P(30, 16)}, {P(30, 16), P(14, 4)},
		{P(30, 16), P(30, 10)}, {P(26, -6), P(12, -14)}, {P(30, -26), P(40, -27)},
		{P(26, -6), P(26, -28)}, {P(32, 40), vikingHarborSmithy},
	}
	g.Paint(Sand, Func{Box{V(8, 0, -48), V(64, 6, 54)}, func(p Vec3) bool {
		h := H(p.X, p.Z)
		if p.Y < h-0.95 {
			return false
		}
		c := Nearest(p.Scale(1 / g.Span()))
		for _, ab := range paths {
			a, b := ab[0], ab[1]
			d := b.Sub(a)
			t := math.Max(0, math.Min(1, p.Plan().Sub(a).Dot(d)/d.Dot(d)))
			if p.Plan().Sub(a.Add(d.Scale(t))).Length() < 1.3 && Hash01(int(c.X), int(c.Z), 29) < 0.85 {
				return true
			}
		}
		return false
	}})

	// Pines on the fjord walls and the hill behind the village.
	for gx := vikingHarborX0 + 2; gx < vikingHarborX1-2; gx += 3.1 {
		for gz := vikingHarborZ0 + 2; gz < vikingHarborZ1-2; gz += 3.1 {
			i, j := int(gx*10), int(gz*10)
			x := gx + 2.4*(Hash01(i, j, 91)-0.5)
			z := gz + 2.4*(Hash01(i, j, 92)-0.5)
			h := H(x, z)
			if h < 3.2 || h > 30 || land.slope(x, z) > 1.5 {
				continue
			}
			if _, _, ok := falls(x, z); ok {
				continue
			}
			if x > 9 && z > -44 {
				continue // Village plain.
			}
			if lat, _, ok := falls(x+2, z+2); ok && math.Abs(lat) < 1 {
				continue
			}
			if Hash01(i, j, 93) > 0.7 {
				continue
			}
			s.pine(g, P(x, z), h, 4.5+3.5*Hash01(i, j, 94), i+j)
		}
	}
	// A few pines just outside the palisade.
	for k, q := range []Vec2{P(62, -24), P(62, 2), P(61, 22), P(58, 40), P(50, 47), P(10, 47), P(62, 32)} {
		s.pine(g, q, H(q.X, q.Y), 5.5+1.5*Hash01(k, 95), k)
	}

	// Jetties.
	s.jetty(g, -14, -9, 14)
	s.jetty(g, 4, -9, 14)

	// Village.
	s.palisade(g, land)
	s.house(g, land, vikingHarborHouse{C: vikingHarborHall, Angle: math.Pi / 2, Length: 32, Width: 12,
		Wall: 3.4, Roof: 5.6, Walls: Wood, Gables: Red, Dragons: true, Door: 0, Posts: true, Smoke: 4})
	s.house(g, land, vikingHarborHouse{C: P(22, -30), Angle: 0.05, Length: 18, Width: 7.4,
		Wall: 2.4, Roof: 3.6, Walls: Wood, Gables: Wood, Door: -2, Smoke: 3, Stone: true})
	s.house(g, land, vikingHarborHouse{C: P(19, -5), Angle: -math.Pi/2 + 0.08, Length: 16, Width: 7,
		Wall: 2.4, Roof: 3.4, Walls: Wood, Gables: Wood, Door: 1})
	s.house(g, land, vikingHarborHouse{C: P(21, 21), Angle: 0.12, Length: 18, Width: 7.6,
		Wall: 2.4, Roof: 3.6, Walls: Wood, Gables: Red, Door: 1, Smoke: 3})
	s.house(g, land, vikingHarborHouse{C: P(46, 18), Angle: math.Pi - 0.25, Length: 15, Width: 7,
		Wall: 2.3, Roof: 3.4, Walls: Wood, Gables: Wood, Door: 0, Stone: true})
	s.house(g, land, vikingHarborHouse{C: P(52, -6), Angle: math.Pi/2 + 0.06, Length: 16, Width: 7,
		Wall: 2.4, Roof: 3.4, Walls: Wood, Gables: Wood, Door: 2, Smoke: 2})
	s.church(g, land)
	s.smithy(g, land)
	// Storehouse on stilts.
	{
		c := P(46, -20)
		y := H(c.X, c.Y)
		g.With(Wood, func() {
			for _, dx := range []float64{-1.4, 1.4} {
				for _, dz := range []float64{-1.4, 1.4} {
					g.Fill(Cylinder{Center: P(c.X+dx, c.Y+dz), Radius: 0.45, Y0: y - 0.5, Y1: y + 1.6})
				}
			}
			g.Fill(Subtract(BlockAt(c, 4.0, 4.0, y+1.6, y+4.4), BlockAt(c, 2.2, 2.2, y+2.0, y+4.0)))
			g.Fill(BlockAt(c, 5.0, 5.0, y+4.4, y+5.6))
		})
		g.With(Grass, func() { g.Fill(Gable{P(c.X-3.0, c.Y-3.0), P(c.X+3.0, c.Y+3.0), y + 5.6, AxisX}) })
		g.Carve(BlockAt(P(c.X, c.Y+2), 1.2, 1.2, y+2.0, y+3.6))
	}
	// Fish-drying racks on the shore outside the palisade.
	s.fishRack(g, land, P(13.5, 37), P(13.5, 50))
	s.fishRack(g, land, P(17.5, 49), P(17.5, 38))
	s.fishRack(g, land, P(14, -24), P(14, -34))
	// Barley and hay fields in the meadow, with haystacks.
	g.Paint(Gold, Func{Box{V(43, 0, 39), V(61, 6, 52)}, func(p Vec3) bool {
		if p.Y < H(p.X, p.Z)-0.95 {
			return false
		}
		if p.X > 52 {
			return int(math.Floor(p.Z/1.3))%3 != 0
		}
		return false
	}})
	g.Paint(Sand, Func{Box{V(43, 0, 39), V(61, 6, 52)}, func(p Vec3) bool {
		if p.Y < H(p.X, p.Z)-0.95 {
			return false
		}
		return p.X > 52 && int(math.Floor(p.Z/1.3))%3 == 0
	}})
	g.Paint(Foliage, Func{Box{V(42.5, 0, 38.5), V(61.5, 6, 52.5)}, func(p Vec3) bool {
		return p.Y >= H(p.X, p.Z)-0.95 && (math.Abs(p.X-42.8) < 0.5 || math.Abs(p.X-52) < 0.5 ||
			math.Abs(p.X-61.2) < 0.5 || math.Abs(p.Z-38.8) < 0.5 || math.Abs(p.Z-52.2) < 0.5)
	}})
	for k, q := range []Vec2{P(45.5, 42), P(49, 46.5), P(46, 50), P(50, 41)} {
		y := H(q.X, q.Y)
		g.With(Sand, func() { g.FillVolume(Ellipsoid{V(q.X, y+0.3, q.Y), V(1.5, 2.0+0.4*Hash01(k, 7), 1.5)}) })
		g.With(Gold, func() { g.Fill(Cone(q, 0.6, y+2.0, 1.0)) })
	}
	// Sheep in the meadow and on the hill.
	for k := 0; k < 20; k++ {
		var q Vec2
		if k < 10 {
			q = P(34.5+6*Hash01(k, 101), 39+13*Hash01(k, 102))
		} else {
			q = P(16+42*Hash01(k, 101), -51.5+7*Hash01(k, 102))
		}
		s.sheep(g, q, H(q.X, q.Y), Hash01(k, 103)*2*math.Pi)
	}
	// Runestone by the gate.
	{
		q := P(38, 40)
		y := H(q.X, q.Y)
		g.With(Stone, func() { g.FillVolume(Ellipsoid{V(q.X, y+1.4, q.Y), V(0.8, 2.4, 1.4)}) })
		g.Paint(Red, Func{Box{V(q.X-1, y+1.2, q.Y-1.5), V(q.X+1, y+2.4, q.Y+1.5)}, func(p Vec3) bool {
			return math.Abs(p.Y-(y+1.6+0.5*math.Sin((p.Z-q.Y)*2))) < 0.35
		}})
	}

	// Longships.
	s.ship(g, vikingHarborShip{C: P(-24, 20), Angle: -math.Pi / 2, L: 26, B: 5.6, Strake: Red,
		Mast: true, Sail: true, Shields: true, Heads: true, Wake: true})
	s.ship(g, vikingHarborShip{C: P(-6, 38), Angle: -math.Pi/2 - 0.2, L: 24, B: 5.4, Strake: Navy,
		Mast: true, Oars: 1, Shields: true, Heads: true, Wake: true})
	s.ship(g, vikingHarborShip{C: P(-6, -19.3), Angle: math.Pi, L: 26, B: 5.6, Strake: Gold,
		Mast: true, Shields: true, Heads: true, Gangway: -1, GangwayReach: 2.6})
	s.ship(g, vikingHarborShip{C: P(-6, -1.2), Angle: math.Pi, L: 24, B: 5.4, Strake: Red,
		Mast: true, Oars: 2, Shields: true, Heads: true, Gangway: -1, GangwayReach: 2.6})
	// Small boats beached at the fjord head and tied at a jetty.
	s.ship(g, vikingHarborShip{C: P(-22, -25.5), Angle: 0.35, L: 9, B: 2.8, Strake: Red})
	s.ship(g, vikingHarborShip{C: P(-2, 8.6), Angle: math.Pi, L: 8, B: 2.6, Strake: Navy})

	return Outward(P(10, -6))
}
