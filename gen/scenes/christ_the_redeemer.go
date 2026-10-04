package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// ChristTheRedeemer is the Art Deco statue on the Corcovado summit above Rio de Janeiro.
// The soapstone figure stands with arms outstretched in a cross, its long robe falling
// in straight vertical pleats, a corded sash at the waist, sleeves hanging from the arms,
// open palms and a head bowed slightly over the city. It rises from a stepped, fluted
// pedestal with the chapel door, on the summit platform with railings; stairs lead down
// to a lower terrace on arcades, a panoramic lift and the cog-railway station, whose
// track winds down through Atlantic forest to the rim. The granite peak drops through
// forested slopes to lower hills and a bay with a small Sugarloaf. The statue faces +Z.
type ChristTheRedeemer struct{}

func init() { Register(ChristTheRedeemer{}) }

func (ChristTheRedeemer) Name() string  { return "christ_the_redeemer" }
func (ChristTheRedeemer) Title() string { return "Christ the Redeemer" }

const (
	christTheRedeemerSummit  = 24.0                          // Summit rock top under the platform.
	christTheRedeemerDeck    = christTheRedeemerSummit + 1.5 // Platform deck top.
	christTheRedeemerFeet    = christTheRedeemerDeck + 11.7  // Pedestal top: the statue's feet.
	christTheRedeemerK       = 1.47                          // World units per statue metre (30 m statue).
	christTheRedeemerRim     = 38.0                          // Site radius.
	christTheRedeemerSea     = 0.5                           // Bay water top.
	christTheRedeemerTerrace = christTheRedeemerDeck - 5.5   // Lower terrace top behind the summit.
	christTheRedeemerShell   = 1.0                           // Statue shell thickness in metres.
	christTheRedeemerScale   = 1.12                          // Whole-site scale over the authored units.
)

// christTheRedeemerDeckPoly is the summit platform plan: a chamfered rectangle.
var christTheRedeemerDeckPoly = []Vec2{
	P(-7.5, -14), P(7.5, -14), P(11.5, -10), P(11.5, 8), P(8, 11.5), P(-8, 11.5), P(-11.5, 8), P(-11.5, -10),
}

// christTheRedeemerRail is the cog railway centreline: x, track top y, z, with arc length.
type christTheRedeemerRail struct {
	pts []Vec3
	acc []float64
}

// near returns the plan distance to the track, the arc length and track top there.
func (r *christTheRedeemerRail) near(x, z float64) (d, s, y float64) {
	d = math.Inf(1)
	p := P(x, z)
	for i := 1; i < len(r.pts); i++ {
		a, b := r.pts[i-1].Plan(), r.pts[i].Plan()
		ab := b.Sub(a)
		t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
		e := p.Sub(a.Add(ab.Scale(t))).Length()
		if e < d {
			d = e
			s = r.acc[i-1] + (r.acc[i]-r.acc[i-1])*t
			y = r.pts[i-1].Y + (r.pts[i].Y-r.pts[i-1].Y)*t
		}
	}
	return
}

// at returns the centreline point and unit plan tangent at an arc length.
func (r *christTheRedeemerRail) at(s float64) (Vec3, Vec2) {
	for i := 1; i < len(r.pts); i++ {
		if s <= r.acc[i] || i == len(r.pts)-1 {
			t := (s - r.acc[i-1]) / (r.acc[i] - r.acc[i-1])
			a, b := r.pts[i-1], r.pts[i]
			dir := b.Plan().Sub(a.Plan())
			return a.Add(b.Sub(a).Scale(t)), dir.Scale(1 / dir.Length())
		}
	}
	return r.pts[0], P(1, 0)
}

func (ChristTheRedeemer) smooth(e0, e1, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-e0)/(e1-e0)))
	return t * t * (3 - 2*t)
}

func (ChristTheRedeemer) smin(a, b, k float64) float64 {
	h := math.Max(0, math.Min(1, 0.5+0.5*(b-a)/k))
	return b + (a-b)*h - k*h*(1-h)
}

func (c ChristTheRedeemer) smax(a, b, k float64) float64 { return -c.smin(-a, -b, k) }

func (ChristTheRedeemer) lerpKnots(ys, vs []float64, y float64) float64 {
	if y <= ys[0] {
		return vs[0]
	}
	for i := 1; i < len(ys); i++ {
		if y <= ys[i] {
			t := (y - ys[i-1]) / (ys[i] - ys[i-1])
			return vs[i-1] + (vs[i]-vs[i-1])*t
		}
	}
	return vs[len(vs)-1]
}

// polyEdge returns the signed distance to a polygon's boundary (positive inside) and
// the coordinate along the nearest edge.
func (ChristTheRedeemer) polyEdge(poly []Vec2, p Vec2) (float64, float64) {
	d, along := math.Inf(1), 0.0
	for i, j := 0, len(poly)-1; i < len(poly); j, i = i, i+1 {
		a, b := poly[j], poly[i]
		ab := b.Sub(a)
		t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
		e := p.Sub(a.Add(ab.Scale(t))).Length()
		if e < d {
			d, along = e, t*ab.Length()
		}
	}
	if !InsidePolygon(poly, p) {
		d = -d
	}
	return d, along
}

// siteQ is the normalized site radius (1 at the rim).
func (ChristTheRedeemer) siteQ(x, z float64) float64 {
	a := math.Atan2(z, x)
	w := 1 + 0.04*math.Sin(3*a+0.6) + 0.03*math.Sin(5*a+2.0) + 0.02*math.Sin(8*a+1.1)
	return math.Hypot(x, z) / (christTheRedeemerRim * w)
}

// bay weights the cove at the front-right that opens to the sea past the rim.
func (c ChristTheRedeemer) bay(x, z float64) float64 {
	return c.smooth(13, 10, math.Hypot(x-27, z-25)+1.0*math.Sin(math.Atan2(z-25, x-27)*3+1))
}

// bare reports where the granite shows through the forest: the sheer faces turned
// towards the bay and the city, and scattered outcrops elsewhere.
func (c ChristTheRedeemer) bare(x, z, slope float64) bool {
	a := math.Atan2(z+1, x)
	front := math.Cos(a - math.Pi/3)
	return slope > 1.25 && (front > -0.15 || c.noise(z*1.3, x*1.3) > 0.45) || slope > 2.2
}

// height is the natural terrain: the steep granite dome, its ridges and foothills.
func (c ChristTheRedeemer) height(x, z float64) float64 {
	const ys = christTheRedeemerSummit
	dz := z + 1
	a := math.Atan2(dz, x)
	// Sheer towards the bay (+X, +Z), a long shoulder down the back ridge (-Z, -X).
	sz := 1.0 - 0.45*c.smooth(0, -14, dz)
	sx := 1.0 - 0.15*c.smooth(0, -12, x)
	r := math.Hypot(x*sx, dz*sz) * (1 + 0.08*math.Sin(3*a+0.7) + 0.06*math.Sin(5*a+2.1))
	core := ys * math.Min(1, 1.25/(1+math.Pow(r/13, 3.2)))
	// The rock rises steeply right up under the platform's edge.
	if e, _ := c.polyEdge(christTheRedeemerDeckPoly, P(x, z)); e > -6 {
		core = math.Max(core, ys-1-math.Max(0, -e)*2.2+0.4*math.Sin(x*0.9+z*0.7))
	}
	ridge := 14.0 * math.Exp(-((x+4)*(x+4)/110 + (z+25)*(z+25)/160))
	west := 9.0 * math.Exp(-((x+24)*(x+24)/90 + (z-4)*(z-4)/120))
	east := 8.0 * math.Exp(-((x-22)*(x-22)/80 + (z+12)*(z+12)/70))
	south := 6.0 * math.Exp(-((x+16)*(x+16)/70 + (z-24)*(z-24)/60))
	east = math.Max(east, south)
	h := c.smax(core, math.Max(ridge, math.Max(west, east)), 2.5)
	h += 0.9*math.Sin(x*0.37+1.3)*math.Sin(z*0.31+0.4) + 0.45*math.Sin(x*0.83+z*0.61+2)
	t := c.smooth(1.0, 0.7, c.siteQ(x, z))
	h = 0.9 + (h-0.9)*t
	b := c.bay(x, z)
	return h*(1-b) - 0.6*b
}

// noise is a cheap smooth pattern in about [-1, 1].
func (ChristTheRedeemer) noise(x, z float64) float64 {
	return 0.5*math.Sin(x*0.21+z*0.13+0.7) + 0.3*math.Sin(x*0.47-z*0.39+2.2) + 0.2*math.Sin(z*0.71+x*0.05+4.1)
}

// built reports whether a plan point lies under the summit works (no trees there).
func (c ChristTheRedeemer) built(x, z float64) bool {
	if d, _ := c.polyEdge(christTheRedeemerDeckPoly, P(x, z)); d > -2.5 {
		return true
	}
	if x > -23 && x < 11 && z > -30 && z < -12 {
		return true
	}
	return false
}

func (c ChristTheRedeemer) Build(g *Grid) Order {
	// Author on a finer lattice, then reinterpret its cells at the grid's span: the
	// whole site comes out christTheRedeemerScale times larger than its authored units.
	outer := g
	g = NewGridWithSpan(outer.Span() / christTheRedeemerScale)
	rail := c.rail()
	c.terrain(g, rail)
	c.sea(g)
	c.sugarloaf(g)
	c.track(g, rail)
	c.city(g)
	c.forest(g, rail)
	c.platform(g)
	c.terrace(g)
	c.pedestal(g)
	c.statue(g)
	outer.Merge(g)
	return Sweep(P(0, 0), -math.Pi/2)
}

// rail lays out the cog railway from the summit station down to the front-left rim,
// easing its grade over the terrain.
func (c ChristTheRedeemer) rail() *christTheRedeemerRail {
	ctrl := []Vec2{P(-12, -24), P(-15, -24.5), P(-20, -24.5), P(-25.5, -20.5), P(-29, -12), P(-30, -1), P(-28.5, 10),
		P(-23, 19.5), P(-15, 26.5), P(-7, 32), P(-2, 36.5)}
	var pts []Vec3
	for i := 0; i+1 < len(ctrl); i++ {
		p0, p1, p2, p3 := ctrl[max(i-1, 0)], ctrl[i], ctrl[i+1], ctrl[min(i+2, len(ctrl)-1)]
		for k := 0; k < 10; k++ {
			t := float64(k) / 10
			t2, t3 := t*t, t*t*t
			f := func(a, b, c2, d float64) float64 {
				return 0.5 * (2*b + (-a+c2)*t + (2*a-5*b+4*c2-d)*t2 + (-a+3*b-3*c2+d)*t3)
			}
			pts = append(pts, V(f(p0.X, p1.X, p2.X, p3.X), 0, f(p0.Y, p1.Y, p2.Y, p3.Y)))
		}
	}
	last := ctrl[len(ctrl)-1]
	pts = append(pts, V(last.X, 0, last.Y))
	r := &christTheRedeemerRail{pts: pts, acc: make([]float64, len(pts))}
	for i := 1; i < len(pts); i++ {
		r.acc[i] = r.acc[i-1] + pts[i].Plan().Sub(pts[i-1].Plan()).Length()
	}
	// Smoothed terrain under the line, then monotone with a capped grade.
	raw := make([]float64, len(pts))
	for i, p := range pts {
		raw[i] = c.height(p.X, p.Z)
	}
	station := c.stationY()
	for i := range pts {
		sum, n := 0.0, 0.0
		for j := max(i-6, 0); j <= min(i+6, len(pts)-1); j++ {
			sum += raw[j]
			n++
		}
		y := sum / n
		if i == 0 {
			y = station
		} else {
			ds := r.acc[i] - r.acc[i-1]
			y = math.Max(math.Min(y, pts[i-1].Y), pts[i-1].Y-0.22*ds)
			if r.acc[i] < 8 {
				y = station
			}
		}
		pts[i].Y = math.Max(y, 1.0)
	}
	return r
}

// stationY is the track top at the summit station.
func (c ChristTheRedeemer) stationY() float64 { return math.Round(c.height(-16, -24.5)*2) / 2 }

// terrain fills the heightfield as a shell, coloring rock, forest floor and lawns,
// and cutting the railway bench.
func (c ChristTheRedeemer) terrain(g *Grid, rail *christTheRedeemerRail) {
	s := g.Span()
	n := int32(math.Ceil(christTheRedeemerRim * 1.12 * 2 / s))
	hAt := func(x, z float64) float64 {
		h := c.height(x, z)
		if d, _, y := rail.near(x, z); d < 4 {
			cut := y - 0.9 + math.Max(0, d-1.6)*1.4
			h = math.Min(h, cut)
		}
		return h
	}
	for X := -n; X <= n; X++ {
		for Z := -n; Z <= n; Z++ {
			if (X^Z)&1 != 0 {
				continue
			}
			x, z := float64(X)*s/2, float64(Z)*s/2
			if c.siteQ(x, z) > 1 {
				continue
			}
			h := hAt(x, z)
			lo := h
			for k := 0; k < 8; k++ {
				a := float64(k) * math.Pi / 4
				lo = math.Min(lo, hAt(x+0.9*math.Cos(a), z+0.9*math.Sin(a)))
			}
			slope := (h - lo) / 0.9
			lo = math.Min(lo, h) - 1.1
			if lo < -1.4 {
				lo = -1.4
			}
			y0 := int32(math.Ceil(lo * 2 / s))
			y1 := int32(math.Floor(h * 2 / s))
			if (y0^X)&1 != 0 {
				y0++
			}
			for Y := y0; Y <= y1; Y += 2 {
				y := float64(Y) * s / 2
				g.Use(c.groundColor(x, y, z, h, slope)).Add(Cell{X, Y, Z})
			}
		}
	}
}

func (c ChristTheRedeemer) groundColor(x, y, z, h, slope float64) Color {
	top := h-y < 0.7
	switch {
	case c.bay(x, z) > 0.2 && h < 1.3:
		return Sand
	case c.bare(x, z, slope):
		if top && slope < 1.9 && c.noise(x*2.1+5, z*2.1) > 0.4 {
			return Foliage
		}
		col := int(math.Floor(x/1.3)) + 7*int(math.Floor(z/1.3))
		if Hash01(col, int(math.Floor(y/5))) < 0.25 {
			return Stone
		}
		return DarkStone
	case !top:
		return DarkStone
	case h > christTheRedeemerSummit-1.5:
		return Stone
	case h < 2.0 || c.noise(x, z) > 0.6:
		return Grass
	default:
		return Foliage
	}
}

// sea fills the bay with water to the sea level.
func (c ChristTheRedeemer) sea(g *Grid) {
	g.Use(Water).Fill(Func{Box{V(-5, -1, -5), V(42, christTheRedeemerSea, 42)}, func(p Vec3) bool {
		return c.siteQ(p.X, p.Z) <= 1 && c.bay(p.X, p.Z) > 0.05 && c.height(p.X, p.Z) < p.Y
	}})
}

// sugarloaf is the small granite dome rising from the bay, with Urca hill beside it.
func (c ChristTheRedeemer) sugarloaf(g *Grid) {
	dome := func(cx, cz, rx, rz, top float64) Solid {
		return Func{Box{V(cx-rx, -0.6, cz-rz), V(cx+rx, top, cz+rz)}, func(p Vec3) bool {
			t := math.Max(0, p.Y/top)
			k := math.Sqrt(math.Max(0, 1-t*t*t))
			dx, dz := (p.X-cx)/rx, (p.Z-cz)/rz
			return dx*dx+dz*dz <= k*k
		}}
	}
	g.Use(DarkStone).FillVolume(dome(28.0, 26.0, 3.0, 3.6, 13.5), dome(22.5, 30.3, 3.4, 2.6, 6.0))
	// Forest skirts and pale granite streaks, on the domes' own cells only.
	box := Box{V(17, -1, 20), V(33, 14, 35)}
	for _, cell := range g.Cells() {
		p := g.World(cell)
		if col, _ := g.ColorOf(cell); col != DarkStone || !box.Contains(p) || c.siteQ(p.X, p.Z) < 0.75 {
			continue
		}
		next := DarkStone
		switch {
		case p.Y < 2.0+2.2*Hash01(int(math.Floor(p.X)), int(math.Floor(p.Z))):
			next = Foliage
		case p.Y > 7 && Hash01(int(math.Floor(p.X/1.2)), int(math.Floor(p.Z/1.2)), 3) < 0.3:
			next = Stone
		}
		if next != DarkStone {
			g.Remove(cell)
			g.Use(next).Add(cell)
		}
	}
	// Cable car line from Urca to the summit, a red car halfway.
	g.Use(Charcoal).Fill(Capsule{V(23.1, 6.4, 29.8), V(27.6, 13.9, 26.5), 0.45})
	g.Use(Red).Fill(BlockAt(P(25.35, 28.15), 1.2, 1.2, 8.6, 10.0))
}

// track lays ballast, rails and viaduct piers along the railway.
func (c ChristTheRedeemer) track(g *Grid, rail *christTheRedeemerRail) {
	box := Box{V(-34, 0, -30), V(2, 30, 40)}
	g.Use(DarkStone).Fill(Func{box, func(p Vec3) bool {
		d, _, y := rail.near(p.X, p.Z)
		return d <= 1.5 && p.Y <= y && p.Y > y-0.9
	}})
	g.Use(Stone).Fill(Func{box, func(p Vec3) bool {
		d, s, y := rail.near(p.X, p.Z)
		if d > 1.0 || p.Y > y-0.9 || math.Mod(s, 4.5) > 1.4 {
			return false
		}
		return p.Y > c.height(p.X, p.Z)-1
	}})
	g.Use(Charcoal).Fill(Func{box, func(p Vec3) bool {
		d, _, y := rail.near(p.X, p.Z)
		return d >= 0.4 && d <= 1.1 && p.Y > y && p.Y <= y+0.5
	}})
	// Two-car train climbing the middle of the line.
	for _, s0 := range []float64{30.5, 36} {
		p, dir := rail.at(s0)
		n := P(-dir.Y, dir.X)
		g.Use(Red).Fill(Func{Box{p.Sub(V(4, 1, 4)), p.Add(V(4, 4, 4))}, func(q Vec3) bool {
			d := q.Plan().Sub(p.Plan())
			u, v := d.Dot(dir), d.Dot(n)
			_, _, y := rail.near(q.X, q.Z)
			return math.Abs(u) <= 2.4 && math.Abs(v) <= 1.1 && q.Y > y+0.3 && q.Y <= y+2.6
		}})
		g.Paint(Glass, Func{Box{p.Sub(V(4, 1, 4)), p.Add(V(4, 4, 4))}, func(q Vec3) bool {
			d := q.Plan().Sub(p.Plan())
			u := d.Dot(dir)
			_, _, y := rail.near(q.X, q.Z)
			return math.Abs(u) <= 2.0 && math.Mod(u+10, 1.4) < 0.8 && q.Y > y+1.4 && q.Y <= y+2.1
		}})
		g.Paint(White, Func{Box{p.Sub(V(4, 1, 4)), p.Add(V(4, 4, 4))}, func(q Vec3) bool {
			_, _, y := rail.near(q.X, q.Z)
			return q.Y > y+2.15
		}})
	}
}

// town is the shore district around the cove.
func (c ChristTheRedeemer) town(x, z float64) bool {
	d := math.Hypot(x-27, z-25) + 1.0*math.Sin(math.Atan2(z-25, x-27)*3+1)
	return d > 13.7 && d < 19 && math.Hypot(x, z) < 33 && c.siteQ(x, z) < 0.96 && c.height(x, z) < 4.5
}

// city lines the cove with Rio's apartment blocks and terracotta-roofed houses.
func (c ChristTheRedeemer) city(g *Grid) {
	const step = 3.1
	for i := 0; i <= 14; i++ {
		for j := 0; j <= 14; j++ {
			x := 2 + float64(i)*step + (Hash01(i, j, 11)-0.5)*0.8
			z := 2 + float64(j)*step + (Hash01(i, j, 12)-0.5)*0.8
			if !c.town(x, z) || Hash01(i, j, 13) < 0.12 {
				continue
			}
			w, dp := 2.0+0.7*Hash01(i, j, 14), 2.0+0.7*Hash01(i, j, 15)
			base := math.Inf(1)
			for _, o := range [][2]float64{{-1, -1}, {1, -1}, {-1, 1}, {1, 1}} {
				base = math.Min(base, c.height(x+o[0]*w/2, z+o[1]*dp/2))
			}
			if c.bay(x, z) > 0.2 {
				base = math.Max(base, christTheRedeemerSea)
			}
			ht := 1.8 + 1.6*Hash01(i, j, 16)
			tall := Hash01(i, j, 17) < 0.3
			if tall {
				ht += 3.5
			}
			walls := []Color{White, White, Sand, Stone}[int(Hash01(i, j, 18)*4)]
			g.Use(walls).Fill(BlockAt(P(x, z), w, dp, base-0.6, base+ht))
			g.Paint(Charcoal, Func{BlockAt(P(x, z), w+0.2, dp+0.2, base+1.0, base+ht-0.4).Bounds(), func(p Vec3) bool {
				return tall && math.Mod(p.Y-base+10, 2.0) < 0.6 && (math.Abs(math.Abs(p.X-x)-w/2) < 0.3 || math.Abs(math.Abs(p.Z-z)-dp/2) < 0.3)
			}})
			if !tall {
				g.Use(Terracotta).Fill(HipRoof{P(x-w/2-0.2, z-dp/2-0.2), P(x+w/2+0.2, z+dp/2+0.2), base + ht})
			} else {
				g.Use(Stone).Fill(BlockAt(P(x, z), w-0.6, dp-0.6, base+ht, base+ht+0.6))
			}
		}
	}
}

// forest scatters tree crowns over the slopes: Atlantic forest with flowering ipês.
func (c ChristTheRedeemer) forest(g *Grid, rail *christTheRedeemerRail) {
	const step = 2.7
	for i := -15; i <= 15; i++ {
		for j := -15; j <= 15; j++ {
			x := float64(i)*step + (Hash01(i, j, 1)-0.5)*1.8
			z := float64(j)*step + (Hash01(i, j, 2)-0.5)*1.8
			if c.siteQ(x, z) > 0.92 || c.bay(x, z) > 0.1 || c.built(x, z) {
				continue
			}
			h := c.height(x, z)
			slope := 0.0
			for k := 0; k < 4; k++ {
				a := float64(k) * math.Pi / 2
				slope = math.Max(slope, math.Abs(c.height(x+math.Cos(a), z+math.Sin(a))-h))
			}
			if h < 1.0 || h > christTheRedeemerSummit-2 || c.noise(x, z) > 0.62 || c.town(x, z) {
				continue
			}
			r := 1.25 + 0.75*Hash01(i, j, 3)
			if c.bare(x, z, slope) || slope > 2.0 {
				// Only scrub clings to the granite.
				if slope > 2.6 || Hash01(i, j, 5) > 0.3 {
					continue
				}
				r = 0.9 + 0.3*Hash01(i, j, 3)
			}
			if d, _, _ := rail.near(x, z); d < 2.8 {
				continue
			}
			col := Foliage
			switch v := Hash01(i, j, 4); {
			case v < 0.05:
				col = Gold
			case v < 0.25:
				col = Grass
			}
			g.Use(col).Fill(Ellipsoid{V(x, h+r*0.45, z), V(r, r*0.8, r)})
		}
	}
}

// platform builds the summit deck on retaining walls with its railing.
func (c ChristTheRedeemer) platform(g *Grid) {
	const d = christTheRedeemerDeck
	poly := christTheRedeemerDeckPoly
	box := Box{V(-12, 0, -15), V(12, d+2, 12)}
	g.Use(Stone).Fill(Func{box, func(p Vec3) bool {
		e, _ := c.polyEdge(poly, p.Plan())
		if e < 0 {
			return false
		}
		if p.Y > d-1 && p.Y <= d {
			return true
		}
		return e < 1.3 && p.Y <= d && p.Y > c.height(p.X, p.Z)-1.5
	}})
	// Paving and a rusticated band on the retaining wall.
	g.Paint(Sand, Func{box, func(p Vec3) bool {
		e, _ := c.polyEdge(poly, p.Plan())
		return e > 1.2 && p.Y > d-0.5 && p.Y <= d
	}})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		e, _ := c.polyEdge(poly, p.Plan())
		return e >= 0 && e < 0.7 && p.Y < d-1.6 && math.Mod(p.Y+100, 2.4) < 0.7
	}})
	// Balustrade: plinth, posts and top rail.
	g.Use(White).Fill(Func{box, func(p Vec3) bool {
		e, along := c.polyEdge(poly, p.Plan())
		if e < 0.1 || e > 0.9 || p.Y <= d {
			return false
		}
		y := p.Y - d
		return y <= 0.5 || (y > 1.1 && y <= 1.7) || (y <= 1.7 && math.Mod(along+0.3, 2.2) < 0.9)
	}})
	// Back opening for the stairs.
	g.Carve(Block(V(-3, d+0.01, -14.5), V(3, d+2, -12.5)))
}

// terrace builds the back staircase, the arcaded lower terrace, the panoramic lift,
// the escalator hall and the railway station.
func (c ChristTheRedeemer) terrace(g *Grid) {
	const d, t = christTheRedeemerDeck, christTheRedeemerTerrace
	// Stair from the deck down to the terrace.
	g.Use(Stone).Fill(Func{Box{V(-3, 0, -21.5), V(3, d, -13)}, func(p Vec3) bool {
		run := -13 - p.Z
		step := math.Floor(run / 1.15)
		top := d - (step+1)*0.79
		return p.Y <= top+0.79 && p.Y > top-0.9
	}})
	g.Use(White).Fill(Func{Box{V(-4, 0, -21.5), V(4, d+1.4, -13.5)}, func(p Vec3) bool {
		if math.Abs(math.Abs(p.X)-3.5) > 0.5 {
			return false
		}
		run := -13.5 - p.Z
		top := d - run*0.69 + 1.4
		return p.Y <= top && p.Y > top-2.4
	}})
	// Terrace slab on arcades.
	tp := []Vec2{P(-10, -27), P(9, -27), P(9, -21.5), P(-10, -21.5)}
	box := Box{V(-11, 0, -28), V(10, t+2, -20)}
	g.Use(Stone).Fill(Func{box, func(p Vec3) bool {
		e, along := c.polyEdge(tp, p.Plan())
		if e < 0 || p.Y > t {
			return false
		}
		if p.Y > t-0.9 {
			return true
		}
		if e > 1.1 || p.Y < c.height(p.X, p.Z)-1.2 {
			return false
		}
		// Arches: openings 2 wide every 3.6 between springing 3 below the deck.
		u := math.Mod(along+100, 3.6) - 1.8
		top := t - 2.6 - math.Sqrt(math.Max(0, 1-u*u))*0.9
		return !(math.Abs(u) < 1.0 && p.Y < top)
	}})
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		e, _ := c.polyEdge(tp, p.Plan())
		return e > 0.7 && e < 1.3 && p.Y < t-0.9
	}})
	g.Use(Charcoal).Fill(Func{box, func(p Vec3) bool {
		e, _ := c.polyEdge(tp, p.Plan())
		return e >= 1.1 && e < 1.8 && p.Y < t-0.9 && p.Y > c.height(p.X, p.Z)-0.5
	}})
	g.Use(White).Fill(Func{box, func(p Vec3) bool {
		e, along := c.polyEdge(tp, p.Plan())
		if e < 0.1 || e > 0.9 || p.Y <= t || math.Abs(p.X) < 3.2 && p.Z > -22.5 {
			return false
		}
		y := p.Y - t
		return y <= 0.5 || (y > 1.1 && y <= 1.7) || (y <= 1.7 && math.Mod(along+0.3, 2.2) < 0.9)
	}})
	// Escalator hall: a glazed sloping gallery from the terrace up to the deck's east side.
	g.Use(White).Fill(Func{Box{V(4, 0, -21.5), V(9, d+2.5, -11)}, func(p Vec3) bool {
		run := (-p.Z - 11.5) / 10
		if run < 0 || run > 1 || p.X < 4.8 || p.X > 8.6 {
			return false
		}
		floor := d - run*(d-t)
		return p.Y > floor-0.8 && p.Y <= floor+3.0
	}})
	g.Paint(Glass, Func{Box{V(4, 0, -21.5), V(9, d+2.5, -11)}, func(p Vec3) bool {
		run := (-p.Z - 11.5) / 10
		floor := d - run*(d-t)
		return p.Y > floor+0.7 && p.Y <= floor+2.2 && math.Mod(p.Z+100, 2.0) < 1.4
	}})
	g.Carve(Func{Box{V(5.6, 0, -21.5), V(7.8, d+2.0, -11)}, func(p Vec3) bool {
		run := (-p.Z - 11.5) / 10
		floor := d - run*(d-t)
		return p.Y > floor && p.Y <= floor+2.2
	}})
	g.Paint(Charcoal, Func{Box{V(4.6, 0, -21.5), V(8.8, d+2.5, -11)}, func(p Vec3) bool {
		run := (-p.Z - 11.5) / 10
		floor := d - run*(d-t)
		return p.Y > floor-0.8 && p.Y <= floor+2.9 && p.X > 5.2 && p.X < 8.2
	}})
	// Station pad, platform shed and lift tower.
	sy := c.stationY()
	g.Use(Stone).Fill(Func{Box{V(-22, 0, -28.5), V(-10, sy, -20.5)}, func(p Vec3) bool {
		return p.Y > c.height(p.X, p.Z)-1.2 && p.Y > sy-1.6
	}})
	g.Use(Stone).Fill(Func{Box{V(-22, 0, -28.5), V(-10, sy, -20.5)}, func(p Vec3) bool {
		e := math.Min(math.Min(p.X+22, -10-p.X), math.Min(p.Z+28.5, -20.5-p.Z))
		return e < 1.2 && p.Y > c.height(p.X, p.Z)-1.2
	}})
	// Shed: white posts, terracotta roof, a waiting room on the north side.
	for _, x := range []float64{-21, -17.5, -14, -10.8} {
		for _, z := range []float64{-27.5, -21.5} {
			g.Use(White).Fill(Block(V(x-0.5, sy, z-0.5), V(x+0.5, sy+3.6, z+0.5)))
		}
	}
	g.Use(White).Fill(Block(V(-21.5, sy, -28.5), V(-11, sy+3.6, -27)))
	g.Paint(Charcoal, Block(V(-20, sy+0.5, -27.2), V(-12.5, sy+2.6, -26.9)))
	g.Use(Terracotta).Fill(Gable{P(-22.5, -29.2), P(-10, -20), sy + 3.6, AxisX})
	g.Carve(Block(V(-22.5, sy+3.6, -28.2), V(-10, sy+4.3, -21)))
	// Lift tower from the station up past the terrace, with a bridge across.
	g.Use(White).Fill(Block(V(-12.5, sy, -27), V(-9.5, t+4.2, -24)))
	g.Paint(Glass, Block(V(-12.5, sy+1, -24.1), V(-9.5, t+3.4, -23.9)), Block(V(-9.6, sy+1, -26.5), V(-9.4, t+3.4, -24.5)))
	g.Use(Charcoal).Fill(Block(V(-12.5, t+3.6, -27), V(-9.5, t+4.6, -24)))
	g.Use(Stone).Fill(Block(V(-10, t-0.9, -26.5), V(-9, t, -24.5)))
}

// pedestal builds the stepped Art Deco plinth: fluted hollow walls, cornices and the
// chapel door on the east (+X) side with a dark interior.
func (c ChristTheRedeemer) pedestal(g *Grid) {
	const d = christTheRedeemerDeck
	sq := func(hw, y0, y1 float64) Cuboid { return Block(V(-hw, y0, -hw), V(hw, y1, hw)) }
	g.Use(Stone).Fill(sq(7.8, d, d+0.8), sq(7.0, d+0.8, d+1.6))
	g.Use(White).Fill(sq(5.7, d+1.6, d+10))
	// Corner piers stepped forward, vertical flutes between them.
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			g.Fill(Block(V(sx*6.1, d+1.6, sz*6.1), V(sx*4.3, d+10, sz*4.3)))
		}
	}
	for _, u := range []float64{-2.6, -0.9, 0.9, 2.6} {
		g.Carve(Block(V(u-0.45, d+2.4, 5.2), V(u+0.45, d+9.3, 6)), Block(V(u-0.45, d+2.4, -6), V(u+0.45, d+9.3, -5.2)))
		g.Carve(Block(V(-6, d+2.4, u-0.45), V(-5.2, d+9.3, u+0.45)))
		if math.Abs(u) > 2 {
			g.Carve(Block(V(5.2, d+4.1, u-0.45), V(6, d+9.3, u+0.45)))
		}
	}
	g.Paint(Stone, Func{Box{V(-6.2, d+1.6, -6.2), V(6.2, d+10, 6.2)}, func(p Vec3) bool {
		return math.Max(math.Abs(p.X), math.Abs(p.Z)) < 5.65 && math.Max(math.Abs(p.X), math.Abs(p.Z)) > 5.1
	}})
	g.Use(Stone).Fill(sq(6.6, d+10, d+10.9))
	g.Use(White).Fill(sq(5.2, d+10.9, christTheRedeemerFeet))
	// Hollow chapel with its door.
	g.Carve(sq(4.2, d+1.6, d+9.5))
	g.Carve(Arch{V(5.5, d+1.6, 0), 1.7, 3.6, 3, AxisX})
	g.Paint(Charcoal, Func{Box{V(-5.2, d, -5.2), V(5.2, d+10.1, 5.2)}, func(p Vec3) bool {
		return math.Max(math.Abs(p.X), math.Abs(p.Z)) < 4.6 && p.Y < d+10.1
	}})
	g.Use(Stone).Fill(Block(V(5.2, d+1.6, -2.6), V(6.3, d+2.0, 2.6)))
	g.Use(Gold).Fill(Block(V(6.0, d+8.0, -0.4), V(6.6, d+10.4, 0.4)), Block(V(6.0, d+9.1, -1.1), V(6.6, d+9.7, 1.1)))
}

// Statue robe profile: half widths across (a) and front-to-back (b) by height in metres.
var (
	christTheRedeemerKY = []float64{0, 1.5, 8, 14, 17, 20, 22.5}
	christTheRedeemerKA = []float64{2.45, 2.25, 2.15, 2.25, 2.3, 2.75, 3.2}
	christTheRedeemerKB = []float64{1.8, 1.7, 1.62, 1.7, 1.75, 1.95, 2.0}
)

func (ChristTheRedeemer) sdEll(p, c, r Vec3) float64 {
	q := p.Sub(c)
	k0 := math.Sqrt(q.X*q.X/(r.X*r.X) + q.Y*q.Y/(r.Y*r.Y) + q.Z*q.Z/(r.Z*r.Z))
	k1 := math.Sqrt(q.X*q.X/(r.X*r.X*r.X*r.X) + q.Y*q.Y/(r.Y*r.Y*r.Y*r.Y) + q.Z*q.Z/(r.Z*r.Z*r.Z*r.Z))
	if k1 == 0 {
		return -math.Min(r.X, math.Min(r.Y, r.Z))
	}
	return k0 * (k0 - 1) / k1
}

func (ChristTheRedeemer) sdCap(p, a, b Vec3, r float64) float64 {
	pa, ba := p.Sub(a), b.Sub(a)
	h := math.Max(0, math.Min(1, pa.Dot(ba)/ba.Dot(ba)))
	return pa.Sub(ba.Scale(h)).Length() - r
}

func (ChristTheRedeemer) sdBox(p, c, half Vec3, round float64) float64 {
	q := p.Sub(c)
	dx, dy, dz := math.Abs(q.X)-half.X+round, math.Abs(q.Y)-half.Y+round, math.Abs(q.Z)-half.Z+round
	out := V(math.Max(dx, 0), math.Max(dy, 0), math.Max(dz, 0)).Length()
	return out + math.Min(math.Max(dx, math.Max(dy, dz)), 0) - round
}

// robe returns the robe's approximate signed distance and its fold value.
func (c ChristTheRedeemer) robe(p Vec3) (float64, float64) {
	a := c.lerpKnots(christTheRedeemerKY, christTheRedeemerKA, p.Y)
	b := c.lerpKnots(christTheRedeemerKY, christTheRedeemerKB, p.Y)
	theta := math.Atan2(p.Z/b, p.X/a)
	f := 2*math.Abs(math.Cos(4*theta)) - 1
	amp := 0.14 * c.smooth(21.5, 18, p.Y)
	limit := 1 + amp*f
	if p.Y < 1.0 {
		limit += 0.05 * (1 - p.Y)
	}
	q := math.Hypot(p.X/a, p.Z/b)
	d := (q - limit) * math.Min(a, b) * 0.95
	return math.Max(d, math.Max(-p.Y, p.Y-23.2)), f
}

// head maps a point into the bowed head frame (pivot at the neck).
func (ChristTheRedeemer) head(p Vec3) Vec3 {
	const tilt = 13 * math.Pi / 180
	q := p.Sub(V(0, 25.4, 0.1))
	cs, sn := math.Cos(tilt), math.Sin(tilt)
	return V(q.X, q.Y*cs+q.Z*sn, -q.Y*sn+q.Z*cs)
}

// statueParts returns signed distances for the body, the hair and beard, and the sash.
func (c ChristTheRedeemer) statueParts(p Vec3) (body, hair, sash, fold float64) {
	ax := math.Abs(p.X)
	m := V(ax, p.Y, p.Z) // Mirrored for the symmetric arms.
	body, fold = c.robe(p)
	body = c.smin(body, c.sdEll(p, V(0, 22.6, -0.1), V(3.4, 1.9, 2.0)), 0.8)
	// Arms: bulky sleeves, wrists and open hands with raised thumbs.
	arm := c.sdCap(m, V(2.6, 23.3, 0), V(10.5, 23.35, 0.2), 0.95)
	arm = c.smin(arm, c.sdCap(m, V(10.2, 23.3, 0.2), V(11.3, 23.25, 0.3), 0.65), 0.3)
	// Sleeve drape: a cloth wing from the forearm down to the hip.
	t := (ax - 2.4) / (10.3 - 2.4)
	lower := 18.6 + t*(22.5-18.6) - 0.5*math.Sin(math.Pi*math.Max(0, math.Min(1, t)))
	th := 0.55 + 0.45*(1-math.Max(0, math.Min(1, t)))
	drape := math.Max(math.Max(lower-p.Y, p.Y-23.3), math.Max(math.Max(ax-10.4, 2.0-ax), math.Abs(p.Z-0.05)-th))
	arm = c.smin(arm, drape, 0.4)
	hand := c.sdBox(m, V(11.95, 23.15, 0.3), V(0.75, 0.85, 0.4), 0.2)
	hand = c.smin(hand, c.sdBox(m, V(13.1, 22.95, 0.3), V(0.7, 0.62, 0.32), 0.18), 0.2)
	hand = math.Min(hand, c.sdCap(m, V(11.6, 23.7, 0.55), V(11.85, 24.55, 0.75), 0.32))
	arm = c.smin(arm, hand, 0.25)
	body = c.smin(body, arm, 0.6)
	// Neck and bowed head.
	body = c.smin(body, c.sdCap(p, V(0, 23.8, 0), V(0, 25.9, 0.2), 0.75), 0.4)
	h := c.head(p)
	face := c.sdEll(h, V(0, 1.95, 0.3), V(0.95, 1.5, 1.15))
	body = c.smin(body, face, 0.2)
	hair = c.sdEll(h, V(0, 2.15, -0.12), V(1.15, 1.6, 1.15))
	hair = c.smin(hair, c.sdCap(h, V(0, 1.6, -0.55), V(0, -0.5, -0.95), 0.95), 0.4)
	hair = math.Min(hair, c.sdCap(V(math.Abs(h.X), h.Y, h.Z), V(0.8, 1.6, -0.15), V(0.85, 0.2, -0.35), 0.45))
	hair = math.Min(hair, c.sdEll(h, V(0, 0.85, 0.75), V(0.72, 0.62, 0.48)))
	// Corded sash at the waist with the knot's hanging ends.
	a := c.lerpKnots(christTheRedeemerKY, christTheRedeemerKA, p.Y) + 0.28
	b := c.lerpKnots(christTheRedeemerKY, christTheRedeemerKB, p.Y) + 0.28
	sash = math.Max((math.Hypot(p.X/a, p.Z/b)-1)*b, math.Abs(p.Y-16.9)-0.38)
	sash = math.Min(sash, c.sdCap(p, V(0.25, 16.7, 2.05), V(0.45, 12.2, 1.95), 0.42))
	sash = math.Min(sash, c.sdCap(p, V(-0.25, 16.7, 2.05), V(-0.5, 13.2, 1.98), 0.42))
	// Feet under the hem and the heart on the chest.
	body = math.Min(body, c.sdEll(m, V(0.6, 0.3, 1.85), V(0.5, 0.38, 0.75)))
	body = math.Min(body, c.sdEll(p, V(0.9, 21.0, 2.0), V(0.42, 0.45, 0.3)))
	return
}

func (c ChristTheRedeemer) statueSD(p Vec3) float64 {
	body, hair, sash, _ := c.statueParts(p)
	return math.Min(body, math.Min(hair, sash))
}

func (ChristTheRedeemer) local(p Vec3) Vec3 {
	return V(p.X, p.Y-christTheRedeemerFeet, p.Z).Scale(1 / christTheRedeemerK)
}

// statue fills the figure as a soapstone shell, then colors pleats, hair and sash.
func (c ChristTheRedeemer) statue(g *Grid) {
	const k, f = christTheRedeemerK, christTheRedeemerFeet
	box := Box{V(-14.6*k, f-0.5, -3.2*k), V(14.6*k, f+31*k, 3.4*k)}
	g.Use(White).FillVolume(Func{box, func(p Vec3) bool {
		d := c.statueSD(c.local(p))
		return d <= 0 && d > -christTheRedeemerShell
	}})
	for _, cell := range g.Cells() {
		p := g.World(cell)
		if !box.Contains(p) {
			continue
		}
		body, hair, sash, fold := c.statueParts(c.local(p))
		col := White
		switch {
		case sash < body-0.05 && sash < hair:
			col = Stone
		case hair < body+0.12:
			col = Stone
		case fold < 0.1 && c.local(p).Y < 18.5 && c.local(p).Y > 0.8:
			col = Stone
		}
		if col != White {
			g.Remove(cell)
			g.Use(col).Add(cell)
		}
	}
}
