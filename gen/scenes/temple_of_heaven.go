package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// TempleOfHeaven is Beijing's Hall of Prayer for Good Harvests: a round red hall with
// three flared, deep-blue conical roofs and a gilded finial on a three-tier white marble
// terrace with balustrades and stairs on four axes, inside a red-walled square courtyard
// with a blue-roofed gate and side halls. The raised Danbi walkway runs south to the
// Imperial Vault of Heaven inside its round Echo Wall, among rows of cypress on grass.
// North is -Z.
type TempleOfHeaven struct{}

func init() { Register(TempleOfHeaven{}) }

func (TempleOfHeaven) Name() string  { return "temple_of_heaven" }
func (TempleOfHeaven) Title() string { return "Temple Of Heaven" }

const (
	templeOfHeavenGround = 0.5   // Ground top.
	templeOfHeavenHallZ  = -20.0 // Hall of Prayer center.
	templeOfHeavenVaultZ = 54.0  // Imperial Vault center.
	templeOfHeavenEcho   = 15.0  // Echo Wall outer radius.
	templeOfHeavenYardX  = 32.0  // Courtyard wall half-width.
	templeOfHeavenYardN  = -48.0 // Courtyard north wall.
	templeOfHeavenYardS  = 14.0  // Courtyard south wall.
	templeOfHeavenBackX  = 17.0  // Huangqian Hall yard half-width.
	templeOfHeavenBackN  = -64.0 // Huangqian Hall yard north wall.
	templeOfHeavenPlotX  = 39.0
	templeOfHeavenPlotN  = -69.0
	templeOfHeavenPlotS  = 73.0
)

// templeOfHeavenRoof is one round flared roof: eave radius/height to top radius/height.
type templeOfHeavenRoof struct{ Re, Rt, Ye, Yt float64 }

func (TempleOfHeaven) clamp01(v float64) float64 { return math.Max(0, math.Min(1, v)) }

func (t TempleOfHeaven) Build(g *Grid) Order {
	t.site(g)
	hall := P(0, templeOfHeavenHallZ)
	top := t.terrace(g, hall, []float64{23, 19.5, 16}, 1.9, []float64{8, 5.5, 5.5, 5.5})
	t.hallOfPrayer(g, hall, top)
	t.courtyard(g)
	t.danbi(g)
	t.vault(g)
	t.cypresses(g)
	return Sweep(hall, math.Pi/2)
}

// plot reports whether a plan point lies on the site, inset by d: a rectangle whose
// north corners are rounded (heaven) and south corners square (earth).
func (TempleOfHeaven) plot(x, z, d float64) bool {
	const r = 16.0
	hx, n, s := templeOfHeavenPlotX-d, templeOfHeavenPlotN+d, templeOfHeavenPlotS-d
	if math.Abs(x) > hx || z < n || z > s {
		return false
	}
	cz := n + r
	if z < cz && math.Abs(x) > hx-r {
		dx, dz := math.Abs(x)-(hx-r), z-cz
		return dx*dx+dz*dz <= r*r
	}
	return true
}

// site lays the grass ground with a stone curb, paved courts and paths.
func (t TempleOfHeaven) site(g *Grid) {
	box := Box{V(-templeOfHeavenPlotX, 0, templeOfHeavenPlotN), V(templeOfHeavenPlotX, 1.2, templeOfHeavenPlotS)}
	g.Use(Grass).Fill(Func{box, func(p Vec3) bool { return p.Y <= templeOfHeavenGround && t.plot(p.X, p.Z, 0) }})
	g.Use(DarkStone).Fill(Func{box, func(p Vec3) bool {
		return p.Y <= 1.0 && t.plot(p.X, p.Z, 0) && !t.plot(p.X, p.Z, 0.9)
	}})
	surface := func(f func(x, z float64) bool) Solid {
		return Func{Box{V(box.Min.X, templeOfHeavenGround, box.Min.Z), V(box.Max.X, templeOfHeavenGround, box.Max.Z)}, func(p Vec3) bool {
			return t.plot(p.X, p.Z, 0.9) && f(p.X, p.Z)
		}}
	}
	// Paved courtyard and the paved disc inside the Echo Wall.
	g.Paint(Stone, surface(func(x, z float64) bool {
		return math.Abs(x) < templeOfHeavenYardX && z > templeOfHeavenYardN && z < templeOfHeavenYardS ||
			math.Abs(x) < templeOfHeavenBackX && z > templeOfHeavenBackN && z <= templeOfHeavenYardN
	}))
	g.Paint(Stone, surface(func(x, z float64) bool {
		return math.Hypot(x, z-templeOfHeavenVaultZ) < templeOfHeavenEcho
	}))
	// Paths: south from the Echo Wall, round the courtyard and to the side gates.
	g.Paint(Sand, surface(func(x, z float64) bool {
		south := math.Abs(x) < 2.5 && z > templeOfHeavenVaultZ
		ring := math.Abs(x) < templeOfHeavenYardX+4 && z > templeOfHeavenYardN-4 && z < templeOfHeavenYardS+4 &&
			!(math.Abs(x) < templeOfHeavenYardX+2 && z > templeOfHeavenYardN-2 && z < templeOfHeavenYardS+2)
		ring = ring && !(math.Abs(x) < templeOfHeavenBackX+2 && z < templeOfHeavenYardN)
		back := math.Abs(x) < templeOfHeavenBackX+4 && z > templeOfHeavenBackN-4 && z < templeOfHeavenYardN &&
			!(math.Abs(x) < templeOfHeavenBackX+2 && z > templeOfHeavenBackN-2)
		ring = ring || back
		side := math.Abs(z-templeOfHeavenHallZ) < 2 && math.Abs(x) > templeOfHeavenYardX
		cross := math.Abs(z-(templeOfHeavenYardS+3)) < 1.5
		return south || ring || side || cross
	}))
	// Light marble axis paths inside the courts.
	g.Paint(White, surface(func(x, z float64) bool {
		return math.Abs(x) < 1.5 && (z > templeOfHeavenHallZ && z < templeOfHeavenYardS ||
			z > templeOfHeavenVaultZ-templeOfHeavenEcho && z < templeOfHeavenVaultZ)
	}))
}

// polar returns a point's radius and angle around plan center c.
func (TempleOfHeaven) polar(c Vec2, p Vec3) (float64, float64) {
	d := p.Plan().Sub(c)
	return d.Length(), d.Angle()
}

// terrace stacks round marble tiers (outer radii, equal heights) with balustrades,
// gargoyle spouts and stairs on the four axes (widths: +Z, +X, -Z, -X). It returns
// the top deck height.
func (t TempleOfHeaven) terrace(g *Grid, c Vec2, radii []float64, h float64, stairs []float64) float64 {
	y := templeOfHeavenGround
	ys := []float64{y}
	for range radii {
		y += h
		ys = append(ys, y)
	}
	for i, r := range radii {
		y0, y1 := ys[i], ys[i+1]
		g.Use(White).Fill(Cylinder{c, r, r - 1.2, y0 - 0.5, y1})
		g.Use(Stone).Fill(Cylinder{c, r - 1.2, 0, y1 - 0.5, y1})
		g.Paint(White, Cylinder{c, r, r - 1.4, y1 - 0.5, y1})
		g.Paint(DarkStone, Cylinder{c, r + 0.1, r - 1.3, y0 - 0.5, y0 + 0.3})
		t.balustrade(g, c, r-0.5, y1, true)
	}
	// Stairs: one continuous flight per axis down every tier.
	for k, w := range stairs {
		if w <= 0 {
			continue
		}
		dir := Polar(P(0, 0), 1, math.Pi/2-float64(k)*math.Pi/2)
		side := P(-dir.Y, dir.X)
		local := func(p Vec3) (float64, float64) {
			d := p.Plan().Sub(c)
			return d.Dot(dir), d.Dot(side)
		}
		reach := radii[0] + 3
		box := Box{c.At(0).Sub(V(reach, 0, reach)), c.At(ys[len(ys)-1] + 2).Add(V(reach, 0, reach))}
		surf := func(s float64) (float64, bool) {
			for i := len(radii) - 1; i >= 0; i-- {
				run := 3.0
				if i > 0 {
					run = radii[i-1] - radii[i]
				}
				a := radii[i] - 0.8
				if s >= a && s <= radii[i]+run {
					return ys[i+1] - (s-a)/(run+0.8)*(ys[i+1]-ys[i]), true
				}
			}
			return 0, false
		}
		g.Carve(Func{box, func(p Vec3) bool {
			s, l := local(p)
			if math.Abs(l) > w/2 {
				return false
			}
			for i, r := range radii {
				if s > r-1.4 && s < r+0.6 && p.Y > ys[i+1]+0.05 && p.Y < ys[i+1]+2 {
					return true
				}
			}
			return false
		}})
		g.Use(White).Fill(Func{box, func(p Vec3) bool {
			s, l := local(p)
			hs, ok := surf(s)
			return ok && math.Abs(l) <= w/2 && p.Y <= hs && p.Y >= hs-2.4
		}})
		// Stair cheeks with a rail.
		g.Fill(Func{box, func(p Vec3) bool {
			s, l := local(p)
			hs, ok := surf(s)
			al := math.Abs(l)
			return ok && al > w/2 && al <= w/2+0.9 && p.Y <= hs+1.2 && p.Y >= hs-2.4
		}})
		if k == 0 {
			// Carved dragon-and-phoenix ramp up the middle of the south stair.
			g.Paint(Stone, Func{box, func(p Vec3) bool {
				s, l := local(p)
				hs, ok := surf(s)
				return ok && math.Abs(l) <= 1.3 && p.Y >= hs-0.6
			}})
			g.Paint(Gold, Func{box, func(p Vec3) bool {
				s, l := local(p)
				hs, ok := surf(s)
				return ok && math.Abs(l) <= 0.5 && p.Y >= hs-0.5 && math.Mod(s+10, 2.2) < 0.9
			}})
		}
	}
	return ys[len(ys)-1]
}

// balustrade rings a deck edge with marble posts, a rail and (optionally) spouts.
func (t TempleOfHeaven) balustrade(g *Grid, c Vec2, r, y float64, spouts bool) {
	n := int(math.Round(2 * math.Pi * r / 1.7))
	g.Use(White).Fill(Cylinder{c, r + 0.45, r - 0.45, y + 0.7, y + 1.1})
	for k := 0; k < n; k++ {
		a := 2 * math.Pi * float64(k) / float64(n)
		at := Polar(c, r, a)
		g.Fill(Cylinder{Center: at, Radius: 0.45, Y0: y, Y1: y + 1.5})
		if spouts && k%2 == 0 {
			g.Fill(Capsule{Polar(c, r+0.3, a).At(y - 0.45), Polar(c, r+1.3, a).At(y - 0.35), 0.42})
		}
	}
}

// roofY is the flared roof surface height at radius r: shallow at the eave, steepening
// up to the top, with a slight upturn at the lip.
func (t TempleOfHeaven) roofY(f templeOfHeavenRoof, r float64) (float64, float64) {
	u := t.clamp01((f.Re - r) / (f.Re - f.Rt))
	const p = 1.7
	y := f.Ye + (f.Yt-f.Ye)*math.Pow(u, p)
	slope := (f.Yt - f.Ye) * p * math.Pow(math.Max(u, 1e-6), p-1) / (f.Re - f.Rt)
	if u < 0.12 {
		y += 0.45 * math.Pow(1-u/0.12, 2)
	}
	return y, slope
}

// roundRoof fills a flared conical shell of blue glazed tiles with slate rib rows and a
// painted (green) soffit at the eave.
func (t TempleOfHeaven) roundRoof(g *Grid, c Vec2, f templeOfHeavenRoof, th float64) {
	box := Box{c.At(f.Ye - th - 1).Sub(V(f.Re+0.5, 0, f.Re+0.5)), c.At(f.Yt + 0.5).Add(V(f.Re+0.5, 0, f.Re+0.5))}
	in := func(p Vec3, depth float64) bool {
		r, _ := t.polar(c, p)
		if r > f.Re || r < f.Rt-1 {
			return false
		}
		s, slope := t.roofY(f, r)
		v := th * math.Sqrt(1+slope*slope)
		return p.Y <= s && p.Y >= s-math.Min(v, depth)
	}
	g.Use(Navy).FillVolume(Func{box, func(p Vec3) bool { return in(p, 4) }})
	if f.Rt <= 0 {
		// A steep crown: fill solid so the shell has no gaps.
		g.Fill(Func{box, func(p Vec3) bool {
			r, _ := t.polar(c, p)
			s, _ := t.roofY(f, r)
			return r < f.Re-1.5 && p.Y <= s-0.6 && p.Y >= f.Ye
		}})
	}
	g.Paint(Slate, Func{box, func(p Vec3) bool {
		r, _ := t.polar(c, p)
		return in(p, 4) && r > f.Re-0.7
	}})
	// Blue-green painted rafters under the eave; a dark tile lip.
	g.Paint(Grass, Func{box, func(p Vec3) bool {
		r, _ := t.polar(c, p)
		s, slope := t.roofY(f, r)
		v := th * math.Sqrt(1+slope*slope)
		return in(p, 4) && p.Y < s-v+0.45 && r < f.Re-0.5
	}})
}

// band builds a painted bracket (dougong) frieze on a drum just under a roof: it
// corbels out from radius r, with gold lines and alternating blue and green panels.
func (t TempleOfHeaven) band(g *Grid, c Vec2, r float64, f templeOfHeavenRoof, y0 float64) {
	box := Box{c.At(y0).Sub(V(r+2, 0, r+2)), c.At(f.Yt).Add(V(r+2, 0, r+2))}
	panels := math.Round(2 * math.Pi * r / 2.6)
	in := func(p Vec3) bool {
		pr, _ := t.polar(c, p)
		s, _ := t.roofY(f, pr)
		reach := r + 0.5 + 1.0*t.clamp01((p.Y-y0)/2.2)
		return pr <= reach && pr >= r-1.2 && p.Y >= y0 && p.Y <= s-0.3
	}
	g.Use(Navy).Fill(Func{box, in})
	g.Paint(Grass, Func{box, func(p Vec3) bool {
		_, a := t.polar(c, p)
		frac := a / (2 * math.Pi) * panels
		return in(p) && p.Y > y0+0.45 && int(math.Floor(frac+100))%2 == 0
	}})
	g.Paint(Gold, Func{box, func(p Vec3) bool {
		return in(p) && p.Y <= y0+0.45
	}})
}

// drum is a red round wall with columns and lattice openings in the given bays.
func (t TempleOfHeaven) drum(g *Grid, c Vec2, r, y0, y1 float64, bays int, columns bool, ly0, ly1, door float64, frame bool) {
	g.Use(Red).Fill(Cylinder{c, r, r - 1.2, y0, y1})
	if ly1 > ly0 {
		t.lattice(g, c, r, bays, ly0, ly1, door, frame)
	}
	if columns {
		for k := 0; k < bays; k++ {
			a := 2*math.Pi*(float64(k)+0.5)/float64(bays) + math.Pi/2
			g.Fill(Cylinder{Center: Polar(c, r+0.2, a), Radius: 0.6, Y0: y0, Y1: ly1 + 0.8})
		}
	}
}

// lattice carves checkered openings (latticed doors/windows) into the bays of a drum
// and frames each panel with gold.
func (t TempleOfHeaven) lattice(g *Grid, c Vec2, r float64, bays int, y0, y1, width float64, frame bool) {
	bay := 2 * math.Pi / float64(bays)
	box := Box{c.At(y0).Sub(V(r+1, 0, r+1)), c.At(y1).Add(V(r+1, 0, r+1))}
	panel := func(p Vec3, grow float64) (float64, bool) {
		pr, a := t.polar(c, p)
		if pr < r-1.6 || pr > r+0.6 {
			return 0, false
		}
		k := math.Round((a - math.Pi/2) / bay)
		arc := (a - math.Pi/2 - k*bay) * r
		return arc, math.Abs(arc) <= width/2+grow && p.Y >= y0-grow && p.Y <= y1+grow
	}
	if frame {
		g.Paint(Gold, Func{box.Grow(0.6), func(p Vec3) bool {
			_, ok := panel(p, 0.5)
			return ok
		}})
	}
	g.Carve(Func{box, func(p Vec3) bool {
		arc, ok := panel(p, 0)
		if !ok {
			return false
		}
		return int(math.Floor((arc+20)/0.8))%2 == 0 && int(math.Floor((p.Y-y0)/0.8))%2 == 0
	}})
	g.Paint(Red, Func{box, func(p Vec3) bool {
		_, ok := panel(p, 0)
		return ok
	}})
}

// hallOfPrayer builds the triple-roofed round hall on the terrace top at y.
func (t TempleOfHeaven) hallOfPrayer(g *Grid, c Vec2, y float64) {
	g.Use(White).Fill(Cylinder{Center: c, Radius: 12, Y0: y - 0.5, Y1: y + 0.8})
	g.Use(Stone).Fill(Cylinder{Center: c, Radius: 9, Y0: y + 0.3, Y1: y + 0.8})
	base := y + 0.8
	r1 := templeOfHeavenRoof{13.6, 7.2, base + 7.4, base + 10}
	r2 := templeOfHeavenRoof{11.2, 5.4, r1.Ye + 6.9, r1.Ye + 9.3}
	r3 := templeOfHeavenRoof{8.8, 0, r2.Ye + 6.4, r2.Ye + 13.2}
	roofs := []templeOfHeavenRoof{r1, r2, r3}
	drums := []float64{10, 8, 6.2}
	for i, r := range drums {
		s, _ := t.roofY(roofs[i], r)
		s -= 0.3
		bandY := roofs[i].Ye - 2.2
		switch i {
		case 0:
			t.drum(g, c, r, base, s, 12, true, base+0.6, bandY-0.6, 3.2, true)
		default:
			wy, _ := t.roofY(roofs[i-1], r)
			t.drum(g, c, r, roofs[i-1].Ye-1, s, 12, false, wy+0.6, bandY-0.4, 2.6*r/8, false)
		}
		t.band(g, c, r, roofs[i], bandY)
	}
	for _, f := range roofs {
		t.roundRoof(g, c, f, 1.2)
	}
	// Gilded finial: stem, collar, ball and spike.
	apex := roofs[2].Yt
	g.Use(Gold).Fill(
		Cylinder{Center: c, Radius: 0.8, Y0: apex - 1, Y1: apex + 1.8},
		Cylinder{Center: c, Radius: 1.3, Y0: apex + 0.4, Y1: apex + 0.9},
		Sphere(c.At(apex+2.6), 1.3),
		Cylinder{Center: c, Radius: 0.5, Y0: apex + 3.5, Y1: apex + 4.6},
	)
}

// hipRoof is a concave rectangular hip roof shell over half-sizes hx, hz.
func (t TempleOfHeaven) hipRoof(g *Grid, c Vec2, hx, hz, ye, rise float64) {
	m := math.Min(hx, hz)
	surf := func(x, z float64) (float64, float64, bool) {
		d := math.Min(hx-math.Abs(x-c.X), hz-math.Abs(z-c.Y))
		if d < 0 {
			return 0, 0, false
		}
		u := t.clamp01(d / m)
		y := ye + rise*math.Pow(u, 1.6)
		if u < 0.15 {
			y += 0.35 * math.Pow(1-u/0.15, 2)
		}
		return y, rise * 1.6 * math.Pow(math.Max(u, 1e-6), 0.6) / m, true
	}
	box := Box{V(c.X-hx-0.5, ye-2, c.Y-hz-0.5), V(c.X+hx+0.5, ye+rise+0.6, c.Y+hz+0.5)}
	in := func(p Vec3) bool {
		s, slope, ok := surf(p.X, p.Z)
		return ok && p.Y <= s && p.Y >= s-1.2*math.Sqrt(1+slope*slope)
	}
	g.Use(Navy).FillVolume(Func{box, in})
	g.Paint(Slate, Func{box, func(p Vec3) bool {
		u, v := math.Abs(p.X-c.X)-hx, math.Abs(p.Z-c.Y)-hz
		along := p.X
		if u > v {
			along = p.Z
		}
		return in(p) && math.Mod(along+100, 1.6) < 0.55
	}})
	// Ridge and hips in dark tile; gold ridge ends.
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		u, v := hx-math.Abs(p.X-c.X), hz-math.Abs(p.Z-c.Y)
		return in(p) && (math.Abs(u-v) < 0.5 || math.Min(u, v) > m-0.6)
	}})
}

// rectHall is a red rectangular hall on a marble plinth with lattice doors on its long
// faces, a painted frieze and a blue hip roof. Long axis along X unless alongZ.
func (t TempleOfHeaven) rectHall(g *Grid, c Vec2, long, deep, plinth, wall float64, alongZ bool) {
	hx, hz := long/2, deep/2
	if alongZ {
		hx, hz = hz, hx
	}
	y0 := templeOfHeavenGround
	g.Use(White).Fill(BlockAt(c, 2*hx+2.4, 2*hz+2.4, y0-0.5, y0+plinth))
	g.Paint(Stone, BlockAt(c, 2*hx+2.4, 2*hz+2.4, y0+plinth-0.4, y0+plinth))
	base := y0 + plinth
	top := base + wall
	g.Use(Red).Fill(Subtract(BlockAt(c, 2*hx, 2*hz, base, top), BlockAt(c, 2*hx-2.4, 2*hz-2.4, base-1, top+1)))
	// Lattice doors along the long faces.
	n := int(math.Floor((2*math.Max(hx, hz) - 2) / 3.2))
	for k := 0; k < n; k++ {
		pos := (float64(k) - float64(n-1)/2) * 3.2
		for _, sd := range []float64{-1, 1} {
			var at Vec2
			var w, d float64
			if alongZ {
				at, w, d = P(c.X+sd*hx, c.Y+pos), 2.4, 2
			} else {
				at, w, d = P(c.X+pos, c.Y+sd*hz), 2, 2.4
			}
			fw, fd := w+0.8, d
			if alongZ {
				fw, fd = w, d+0.8
			}
			g.Paint(Gold, BlockAt(at, fw, fd, base+0.4, top-1.2))
			g.Carve(Func{BlockAt(at, w, d, base+0.8, top-1.6).Bounds(), func(p Vec3) bool {
				a := p.X
				if alongZ {
					a = p.Z
				}
				return int(math.Floor((a+100)/0.8))%2 == 0 && int(math.Floor((p.Y-base)/0.8))%2 == 1
			}})
		}
	}
	// Painted frieze under the eaves.
	g.Use(Navy).Fill(Subtract(BlockAt(c, 2*hx+1, 2*hz+1, top-1, top+0.6), BlockAt(c, 2*hx-2.4, 2*hz-2.4, top-2, top+2)))
	g.Paint(Grass, Func{BlockAt(c, 2*hx+1, 2*hz+1, top-0.5, top+0.6).Bounds(), func(p Vec3) bool {
		return int(math.Floor((p.X+p.Z+100)/1.8))%2 == 0
	}})
	g.Paint(Gold, BlockAt(c, 2*hx+1, 2*hz+1, top-1, top-0.6))
	t.hipRoof(g, c, hx+1.8, hz+1.8, top+0.2, math.Min(hx, hz)+0.8)
}

// courtyard builds the red enclosure walls with blue copings, the Qinian Gate on the
// south, side gates and the east and west annex halls.
func (t TempleOfHeaven) courtyard(g *Grid) {
	y0 := templeOfHeavenGround
	const h = 3.6
	x, n, s := templeOfHeavenYardX, templeOfHeavenYardN, templeOfHeavenYardS
	hz := templeOfHeavenHallZ
	wall := func(a, b Vec2, alongZ bool) {
		body := Block(a.At(y0-0.5), b.At(y0+h))
		g.Use(Red).Fill(body)
		g.Paint(Stone, Block(a.At(y0-0.5), b.At(y0+0.6)))
		lo, hi := P(a.X-0.8, a.Y-0.8), P(b.X+0.8, b.Y+0.8)
		ridge := AxisX
		if alongZ {
			ridge = AxisZ
		}
		roof := Gable{lo, hi, y0 + h - 0.4, ridge}
		g.Use(Navy).Fill(Subtract(roof, Translate(roof, V(0, -1, 0))))
		g.Paint(Gold, Block(P(lo.X, lo.Y).At(y0+h-0.4), P(hi.X, hi.Y).At(y0+h-0.4)))
	}
	const t2 = 0.6
	// North and side walls, the south wall split by the gate, side gaps for side gates.
	wall(P(-x, n-t2), P(-5, n+t2), false)
	wall(P(5, n-t2), P(x, n+t2), false)
	t.gate(g, P(0, n), 8, 2.6, false)
	// Huangqian Hall's walled yard behind (north of) the main court.
	bx, bn := templeOfHeavenBackX, templeOfHeavenBackN
	wall(P(-bx, bn-t2), P(bx, bn+t2), false)
	for _, sx := range []float64{-1, 1} {
		wall(P(sx*bx-t2, bn), P(sx*bx+t2, n), true)
	}
	t.rectHall(g, P(0, (bn+n)/2-0.5), 19, 7, 1.4, 4.4, false)
	wall(P(-x, s-t2), P(-8, s+t2), false)
	wall(P(8, s-t2), P(x, s+t2), false)
	for _, sx := range []float64{-1, 1} {
		wall(P(sx*x-t2, n), P(sx*x+t2, hz-3.5), true)
		wall(P(sx*x-t2, hz+3.5), P(sx*x+t2, s), true)
		// Side gate pavilions.
		t.gate(g, P(sx*x, hz), 6, 2.6, true)
	}
	// Corner posts.
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{n, s} {
			g.Use(Red).Fill(BlockAt(P(sx*x, sz), 2.2, 2.2, y0-0.5, y0+h+0.4))
			t.hipRoof(g, P(sx*x, sz), 1.8, 1.8, y0+h+0.4, 1.2)
		}
	}
	// Qinian Gate: a five-bay hall straddling the south wall, with open passages.
	gc := P(0, s)
	t.rectHall(g, gc, 15, 6, 1.4, 4.6, false)
	for _, px := range []float64{-4.8, 0, 4.8} {
		g.Carve(Arch{V(px, y0+1.4, s), 1.2, 2.2, 8, AxisZ})
	}
	// Steps up to the gate on both faces.
	for _, sd := range []float64{-1, 1} {
		for k := 0; k < 3; k++ {
			z := s + sd*(4.2+float64(k)*0.8)
			g.Use(White).Fill(BlockAt(P(0, z), 12, 0.8, y0-0.5, y0+1.4-float64(k+1)*0.45))
		}
	}
	// East and west annex halls.
	for _, sx := range []float64{-1, 1} {
		t.rectHall(g, P(sx*25.5, 2), 15, 5.5, 0.9, 4, true)
	}
	// Bronze-and-stone incense burners flanking the axis.
	for _, sx := range []float64{-1, 1} {
		at := P(sx*8.5, 8.5)
		g.Use(DarkStone).Fill(Cylinder{Center: at, Radius: 0.9, Y0: y0, Y1: y0 + 1.6})
		g.Use(Gold).Fill(Cylinder{Center: at, Radius: 0.5, Y0: y0 + 1.6, Y1: y0 + 2.2})
	}
}

// gate is a small red gate pavilion with an arched passage and a blue hip roof.
func (t TempleOfHeaven) gate(g *Grid, c Vec2, long, deep float64, alongZ bool) {
	y0 := templeOfHeavenGround
	w, d := long, deep
	through := AxisZ
	if alongZ {
		w, d = deep, long
		through = AxisX
	}
	g.Use(Red).Fill(BlockAt(c, w, d, y0-0.5, y0+5))
	g.Paint(Stone, BlockAt(c, w, d, y0-0.5, y0+0.6))
	g.Carve(Arch{c.At(y0), 1.4, 1.8, 10, through})
	g.Use(Navy).Fill(Subtract(BlockAt(c, w+1, d+1, y0+4.2, y0+5.2), BlockAt(c, w-1.2, d-1.2, y0+3, y0+6)))
	g.Paint(Gold, BlockAt(c, w+1, d+1, y0+4.2, y0+4.6))
	t.hipRoof(g, c, w/2+1.4, d/2+1.4, y0+5.2, math.Min(w, d)/2+0.6)
}

// danbi builds the raised Danbi walkway from the Qinian Gate south to the Echo Wall,
// rising towards the north, with three paved lanes and the arched Ghost Gate passage.
func (t TempleOfHeaven) danbi(g *Grid) {
	z0, z1 := templeOfHeavenYardS+6.5, templeOfHeavenVaultZ-templeOfHeavenEcho-0.5
	hTop := func(z float64) float64 { return 2.6 - 1.0*t.clamp01((z-z0)/(z1-z0)) }
	const hw = 4.5
	box := Box{V(-hw-1, 0, z0-1), V(hw+1, 3.2, z1)}
	g.Use(Stone).Fill(Func{box, func(p Vec3) bool {
		return math.Abs(p.X) <= hw && p.Z >= z0 && p.Z <= z1 && p.Y <= hTop(p.Z)
	}})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		return math.Abs(p.X) > hw-0.8 && p.Y > hTop(p.Z)-0.5
	}})
	g.Paint(White, Func{box, func(p Vec3) bool {
		return math.Abs(p.X) <= 1.3 && p.Y > hTop(p.Z)-0.5
	}})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		return math.Abs(p.X) <= hw && p.Y < templeOfHeavenGround+0.5
	}})
	// Steps from the Danbi down to the gate court at the north end.
	for k := 0; k < 4; k++ {
		z := z0 - 0.7*float64(k+1)
		g.Use(White).Fill(BlockAt(P(0, z), 2*hw-1, 0.8, 0, 2.6-float64(k+1)*0.5))
	}
	// Ghost Gate: an arched passage under the walkway.
	g.Carve(Arch{V(0, templeOfHeavenGround, (z0+z1)/2), 1.2, 0.3, 2*hw + 2, AxisX})
	g.Paint(DarkStone, Func{Box{V(-hw, 0, (z0+z1)/2-2.2), V(hw, 3, (z0+z1)/2+2.2)}, func(p Vec3) bool {
		d := math.Hypot(p.Z-(z0+z1)/2, p.Y-templeOfHeavenGround-0.3)
		return d < 2.2 && math.Abs(p.X) > hw-0.8
	}})
}

// vault builds the Imperial Vault of Heaven: the Echo Wall with its gates, two annex
// halls and the single-roofed round vault on a marble terrace.
func (t TempleOfHeaven) vault(g *Grid) {
	c := P(0, templeOfHeavenVaultZ)
	y0 := templeOfHeavenGround
	r := templeOfHeavenEcho
	// Echo Wall: red, on a stone foot, with a blue coping ring.
	g.Use(Red).Fill(Cylinder{c, r, r - 1, y0 - 0.5, y0 + 4})
	g.Paint(Stone, Cylinder{c, r + 0.5, r - 1.5, y0 - 0.5, y0 + 0.6})
	box := Box{c.At(y0 + 3).Sub(V(r+1.5, 0, r+1.5)), c.At(y0 + 6).Add(V(r+1.5, 0, r+1.5))}
	g.Use(Navy).Fill(Func{box, func(p Vec3) bool {
		pr, _ := t.polar(c, p)
		peak := y0 + 4 + 0.9 - 1.2*math.Abs(pr-(r-0.5))
		return pr <= r+0.5 && pr >= r-1.5 && p.Y <= peak && p.Y >= peak-1
	}})
	// Glazed triple gates on north (facing the Danbi) and south.
	for _, sd := range []float64{-1, 1} {
		gz := c.Y + sd*(r-0.5)
		g.Use(Red).Fill(BlockAt(P(0, gz), 10, 2.4, y0-0.5, y0+5))
		g.Paint(White, BlockAt(P(0, gz), 10, 2.4, y0-0.5, y0+0.8))
		g.Use(Navy).Fill(Subtract(BlockAt(P(0, gz), 11, 3.4, y0+4.3, y0+5.3), BlockAt(P(0, gz), 8.6, 1, y0+3, y0+6)))
		g.Paint(Grass, BlockAt(P(0, gz), 11, 3.4, y0+4.3, y0+4.7))
		t.hipRoof(g, P(0, gz), 6.2, 2.4, y0+5.3, 1.6)
		for _, px := range []float64{-3.2, 0, 3.2} {
			g.Carve(Arch{V(px, y0+0.3, gz), 0.9, 1.6, 6, AxisZ})
		}
	}
	// Annex halls.
	for _, sx := range []float64{-1, 1} {
		t.rectHall(g, P(sx*9.9, c.Y+0.5), 5, 2.6, 0.7, 3, true)
	}
	// Vault terrace and hall.
	ty := y0 + 1.8
	const tr, dr = 6.6, 4.4
	g.Use(White).Fill(Cylinder{c, tr, tr - 1.2, y0 - 0.5, ty})
	g.Use(Stone).Fill(Cylinder{c, tr - 1.2, 0, ty - 0.5, ty})
	g.Paint(DarkStone, Cylinder{c, tr + 0.1, tr - 1.3, y0 - 0.5, y0 + 0.3})
	t.balustrade(g, c, tr-0.5, ty, true)
	for _, sd := range []float64{-1, 1} {
		g.Carve(Block(V(-1.8, ty+0.05, c.Y+sd*(tr-1.4)), V(1.8, ty+2, c.Y+sd*(tr+0.6))))
		for k := 0; k < 4; k++ {
			z := c.Y + sd*(tr-0.6+float64(k)*0.75)
			g.Use(White).Fill(BlockAt(P(0, z), 3.6, 0.8, y0-0.5, ty-float64(k)*0.42))
		}
	}
	f := templeOfHeavenRoof{7, 0, ty + 6.6, ty + 13.4}
	s, _ := t.roofY(f, dr)
	s -= 0.3
	g.Use(White).Fill(Cylinder{Center: c, Radius: dr + 0.9, Y0: ty - 0.5, Y1: ty + 0.6})
	t.drum(g, c, dr, ty+0.6, s, 8, true, ty+1.0, ty+3.8, 1.8, true)
	t.band(g, c, dr, f, f.Ye-2.2)
	t.roundRoof(g, c, f, 1.2)
	g.Use(Gold).Fill(
		Cylinder{Center: c, Radius: 0.7, Y0: f.Yt - 1, Y1: f.Yt + 1.2},
		Sphere(c.At(f.Yt+1.9), 1.1),
		Cylinder{Center: c, Radius: 0.45, Y0: f.Yt + 2.6, Y1: f.Yt + 3.4},
	)
}

// cypresses plants rows of tall conical cypresses on the lawns around the courts.
func (t TempleOfHeaven) cypresses(g *Grid) {
	const step = 4.6
	for z := templeOfHeavenPlotN + 3; z < templeOfHeavenPlotS-2; z += step {
		for x := -templeOfHeavenPlotX + 3; x < templeOfHeavenPlotX-2; x += step {
			ix, iz := int(math.Round(x*10)), int(math.Round(z*10))
			at := P(x+(Hash01(ix, iz, 1)-0.5)*1.4, z+(Hash01(ix, iz, 2)-0.5)*1.4)
			if !t.plot(at.X, at.Y, 2.4) || Hash01(ix, iz, 3) < 0.12 {
				continue
			}
			yard := math.Abs(at.X) < templeOfHeavenYardX+5 && at.Y > templeOfHeavenYardN-5 && at.Y < templeOfHeavenYardS+5.5
			walk := math.Abs(at.X) < 8.5 && at.Y > templeOfHeavenYardS && at.Y < templeOfHeavenVaultZ
			echo := math.Hypot(at.X, at.Y-templeOfHeavenVaultZ) < templeOfHeavenEcho+3
			south := math.Abs(at.X) < 5 && at.Y > templeOfHeavenVaultZ
			side := math.Abs(at.Y-templeOfHeavenHallZ) < 4 && math.Abs(at.X) > templeOfHeavenYardX
			back := math.Abs(at.X) < templeOfHeavenBackX+5 && at.Y > templeOfHeavenBackN-5 && at.Y < templeOfHeavenYardN
			if yard || walk || echo || south || side || back {
				continue
			}
			h := 4.8 + 2.6*Hash01(ix, iz, 4)
			base := templeOfHeavenGround
			g.Use(Wood).Fill(Cylinder{Center: at, Radius: 0.5, Y0: base, Y1: base + 1.6})
			g.Use(Foliage).FillVolume(
				Frustum{at, 1.8, 0.3, base + 1.2, base + h},
				Ellipsoid{at.At(base + 1.9), V(1.8, 1.0, 1.8)},
			)
		}
	}
}
