package scenes

import (
	"math"

	"godocta/gen/kit"
	. "godocta/gen/octa"
)

// EiffelTower is the Paris landmark on the Champ de Mars: four splayed, open-lattice
// iron legs with X-braced panels, decorative arches under the first platform, three
// platforms with galleries, the tapering shaft, the campanile, cupola and antenna.
// Around it: the paved esplanade, the Champ de Mars lawns and tree-lined alleys (+X),
// the Seine with the Pont d'Iéna (-X), and the Trocadéro gardens with the fountain
// basin and the curved wings of the Palais de Chaillot beyond.
type EiffelTower struct{}

func init() { Register(EiffelTower{}) }

func (EiffelTower) Name() string  { return "eiffel_tower" }
func (EiffelTower) Title() string { return "Eiffel Tower" }

const (
	eiffelScale = 120.0 / 330 // World units per meter: 330 m to the antenna tip.
	eiffelBase  = 0.6         // Ground level the tower stands on.
	eiffelDepth = 1.1         // Horizontal depth of each latticed truss wall.
	eiffelChord = 1.2         // Width of the corner chords along a face.
	eiffelBrace = 0.55        // Half thickness of struts and diagonals.
)

// outer is the tower's outer half-width at world height y: an exponential taper.
func (EiffelTower) outer(y float64) float64 {
	h := (y - eiffelBase) / eiffelScale
	return (6.5 + 56*math.Exp(-h/85)) * eiffelScale
}

// inner is the half-gap between the legs at height y; it reaches zero where they merge.
func (t EiffelTower) inner(y float64) float64 {
	h := (y - eiffelBase) / eiffelScale
	return t.outer(y) - (11+15*math.Exp(-h/60))*eiffelScale
}

// level is the world height of a real height in meters.
func (EiffelTower) level(meters float64) float64 { return eiffelBase + meters*eiffelScale }

// raster sets every cell whose center lies in the box and for which f reports a color,
// without the half-unit snapping of Fill, so slanted lattices stay crisp.
func (EiffelTower) raster(g *Grid, b Box, f func(Vec3) (Color, bool)) {
	s := g.Span()
	lo := func(v float64) int32 { return int32(math.Ceil(v * 2 / s)) }
	hi := func(v float64) int32 { return int32(math.Floor(v * 2 / s)) }
	previous := g.Brush()
	for y := lo(b.Min.Y); y <= hi(b.Max.Y); y++ {
		z0, x0 := lo(b.Min.Z), lo(b.Min.X)
		if (z0^y)&1 != 0 {
			z0++
		}
		if (x0^y)&1 != 0 {
			x0++
		}
		for z := z0; z <= hi(b.Max.Z); z += 2 {
			for x := x0; x <= hi(b.Max.X); x += 2 {
				c := Cell{X: x, Y: y, Z: z}
				if color, ok := f(g.World(c)); ok {
					g.Remove(c)
					g.Use(color).Add(c)
				}
			}
		}
	}
	g.Use(previous)
}

// member reports whether face coordinates (u across a face of width w, v up a panel
// of height p) lie on a chord, a strut or a diagonal. Wide panels get an X brace;
// narrow ones a single diagonal whose direction alternates with odd (a Warren truss),
// so the slender upper faces stay open.
func (EiffelTower) member(u, w, v, p float64, odd bool) bool {
	chord := eiffelChord
	if w < 6 {
		chord = 1.0
	}
	if u < chord || u > w-chord || v < eiffelBrace || v > p-eiffelBrace {
		return true
	}
	l := math.Hypot(p, w)
	rising := math.Abs(u*p-v*w)/l < eiffelBrace
	falling := math.Abs((w-u)*p-v*w)/l < eiffelBrace
	switch {
	case w >= 6:
		return rising || falling
	case odd:
		return falling
	}
	return rising
}

// panels splits [y0, y1] into n equal panels, appending boundaries after y0.
func (EiffelTower) panels(bounds []float64, y0, y1 float64, n int) []float64 {
	for i := 1; i <= n; i++ {
		bounds = append(bounds, y0+(y1-y0)*float64(i)/float64(n))
	}
	return bounds
}

// panel returns the position within, height and parity of the panel containing y.
func (EiffelTower) panel(bounds []float64, y float64) (float64, float64, bool) {
	for i := 1; i < len(bounds); i++ {
		if y <= bounds[i] {
			return y - bounds[i-1], bounds[i] - bounds[i-1], i%2 == 1
		}
	}
	return 0, 1, false
}

func (t EiffelTower) Build(g *Grid) Order {
	t.ground(g)
	t.trocadero(g)
	t.trees(g)
	t.academy(g)
	t.tower(g)
	return Outward(P(0, 0))
}

// tower builds the legs, shaft, arches, platforms and the crown.
func (t EiffelTower) tower(g *Grid) {
	y1, y2, y3 := t.level(57), t.level(115), t.level(276)
	// Where the legs merge into one shaft.
	lo, hi := y2, y3
	for i := 0; i < 50; i++ {
		mid := (lo + hi) / 2
		if t.inner(mid) > 0 {
			lo = mid
		} else {
			hi = mid
		}
	}
	merge := lo
	bounds := []float64{eiffelBase}
	bounds = t.panels(bounds, eiffelBase, y1, 3)
	bounds = t.panels(bounds, y1, y2, 4)
	bounds = t.panels(bounds, y2, merge, 6)
	bounds = t.panels(bounds, merge, y3+3.2, 4)

	// face tests one truss wall: u across the face of width w.
	face := func(u, w, y float64) bool {
		v, p, odd := t.panel(bounds, y)
		return t.member(u, w, v, p, odd)
	}
	r0 := t.outer(eiffelBase) + 1
	g.Use(Wood)
	t.raster(g, Box{V(-r0, eiffelBase, -r0), V(r0, y3+3.2, r0)}, func(p Vec3) (Color, bool) {
		w, in := t.outer(p.Y), t.inner(p.Y)
		if p.Y > y3 {
			w = t.outer(y3)
		}
		ax, az := math.Abs(p.X), math.Abs(p.Z)
		if ax > w || az > w {
			return 0, false
		}
		if in > 0 { // Four separate legs, each a square truss tube.
			if ax < in || az < in {
				return 0, false
			}
			span := w - in
			// Above the first level the inner faces, which nearly touch, keep only
			// their struts so the legs read as open from outside.
			inside := func(u float64) bool {
				if p.Y < y1 {
					return face(u, span, p.Y)
				}
				v, h, _ := t.panel(bounds, p.Y)
				return v < eiffelBrace || v > h-eiffelBrace
			}
			if ax >= w-eiffelDepth && face(az-in, span, p.Y) || ax <= in+eiffelDepth && inside(az-in) {
				return Wood, true
			}
			if az >= w-eiffelDepth && face(ax-in, span, p.Y) || az <= in+eiffelDepth && inside(ax-in) {
				return Wood, true
			}
			return 0, false
		}
		// The merged shaft: one square truss tube.
		if ax >= w-eiffelDepth && face(p.Z+w, 2*w, p.Y) {
			return Wood, true
		}
		if az >= w-eiffelDepth && face(p.X+w, 2*w, p.Y) {
			return Wood, true
		}
		return 0, false
	})

	// Masonry piers under each leg.
	base, gap := t.outer(eiffelBase), t.inner(eiffelBase)
	g.With(Stone, func() {
		for _, sx := range []float64{-1, 1} {
			for _, sz := range []float64{-1, 1} {
				g.Fill(Block(V(sx*(gap-0.6), eiffelBase, sz*(gap-0.6)), V(sx*(base+0.6), eiffelBase+1.2, sz*(base+0.6))))
			}
		}
	})

	t.arches(g, y1)
	t.platforms(g, y1, y2)
	t.summit(g, y3)
}

// arches adds the four great decorative arches between the legs below the first
// platform. Each lies on the slanted outer face, with an openwork spandrel above.
func (t EiffelTower) arches(g *Grid, y1 float64) {
	top := y1 - 3                 // Crown, under the first platform frieze.
	spring := eiffelBase + 8      // Where the arch meets the legs' inner faces.
	half := t.inner(spring) + 0.3 // Half span at the springing.
	yc := (top*top - spring*spring - half*half) / (2 * (top - spring))
	radius := top - yc
	r0 := t.outer(eiffelBase) + 1
	t.raster(g, Box{V(-r0, spring-3, -r0), V(r0, y1, r0)}, func(p Vec3) (Color, bool) {
		for _, axes := range [2][2]float64{{p.X, p.Z}, {p.Z, p.X}} {
			across, depth := math.Abs(axes[0]), math.Abs(axes[1])
			w := t.outer(p.Y)
			if math.Abs(depth-(w-0.6)) > 0.6 || across > t.inner(p.Y)+0.6 {
				continue
			}
			r := math.Hypot(across, p.Y-yc)
			if r <= radius && r >= radius-1.6 && p.Y >= spring-3 {
				return Wood, true
			}
			// Spandrel: vertical ribs between the arch and the platform girder.
			if r > radius && p.Y <= top && math.Mod(across+0.5, 2.6) < 1.0 {
				return Wood, true
			}
		}
		return 0, false
	})
}

// platforms adds the first and second levels: decks with central voids, decorative
// friezes with arcades, glazed galleries, an upper deck and railings.
func (t EiffelTower) platforms(g *Grid, y1, y2 float64) {
	ring := func(outer, inner, y0, y1 float64) Solid {
		return Subtract(BlockAt(P(0, 0), 2*outer, 2*outer, y0, y1), BlockAt(P(0, 0), 2*inner, 2*inner, y0-1, y1+1))
	}
	// arcade carves small round-headed openings along all four sides of a square band.
	arcade := func(half, sill, rise, step float64) {
		for a := -half + step; a <= half-step+0.01; a += step {
			for _, s := range []float64{-1, 1} {
				g.Carve(Arch{V(a, sill, s*half), 0.55, rise, 2.4, AxisZ}, Arch{V(s*half, sill, a), 0.55, rise, 2.4, AxisX})
			}
		}
	}
	// First level.
	o1, i1 := t.outer(y1)+1.2, t.inner(y1)-0.4
	g.Use(Wood).Fill(ring(o1, i1, y1-0.6, y1))
	g.Fill(ring(o1, o1-1, y1-3, y1), ring(i1+1, i1, y1-2, y1))
	arcade(o1-0.5, y1-2.4, 0.5, 1.9)
	g.Paint(Gold, ring(o1+0.5, o1-0.4, y1-3, y1-3))
	g.Paint(Gold, ring(o1+0.5, o1-0.4, y1-0.4, y1))
	// Gallery pavilions with glazed walls, mullions and dark roofs, open at mid-sides.
	gallery := func(outer, inner, y0, height float64, opening float64) {
		box := Box{V(-outer, y0, -outer), V(outer, y0+height+0.6, outer)}
		g.Use(Glass).Fill(Func{box, func(p Vec3) bool {
			ax, az := math.Abs(p.X), math.Abs(p.Z)
			m := math.Max(ax, az)
			return m <= outer && m >= inner && p.Y <= y0+height && math.Min(ax, az) >= opening
		}})
		g.Use(Wood).Fill(Func{box, func(p Vec3) bool {
			ax, az := math.Abs(p.X), math.Abs(p.Z)
			m := math.Max(ax, az)
			post := math.Mod(math.Min(ax, az)+0.5, 2.5) < 0.9 || math.Abs(ax-az) < 0.8
			return m <= outer && m >= outer-1 && p.Y <= y0+height && math.Min(ax, az) >= opening && post
		}})
		g.Use(Charcoal).Fill(Func{box, func(p Vec3) bool {
			ax, az := math.Abs(p.X), math.Abs(p.Z)
			m := math.Max(ax, az)
			return m <= outer+0.5 && m >= inner-0.5 && p.Y > y0+height && math.Min(ax, az) >= opening-0.5
		}})
	}
	gallery(o1-2, o1-5, y1, 2.6, 2.5)
	railing := func(half, y float64) {
		g.Use(Charcoal).Fill(Subtract(ring(half, half-0.8, y, y+1.1), Func{Box{V(-half, y, -half), V(half, y+0.7, half)}, func(p Vec3) bool {
			return math.Mod(math.Min(math.Abs(p.X), math.Abs(p.Z))+0.75, 1.8) > 1.0
		}}))
	}
	railing(o1, y1)
	// Second level: deck, frieze, glazed lower gallery, upper deck.
	o2, i2 := t.outer(y2)+1.0, math.Max(t.inner(y2)-0.4, 1.5)
	g.Use(Wood).Fill(ring(o2, i2, y2-0.6, y2))
	g.Fill(ring(o2, o2-1, y2-2.2, y2))
	arcade(o2-0.5, y2-1.8, 0.4, 1.9)
	g.Paint(Gold, ring(o2+0.5, o2-0.4, y2-2.2, y2-2.2))
	gallery(o2-0.8, o2-3.2, y2, 2.4, 0)
	g.Use(Wood).Fill(ring(o2-0.3, i2, y2+3.0, y2+3.5))
	railing(o2, y2)
	railing(o2-0.3, y2+3.5)
	g.Paint(Gold, ring(o2+0.5, o2-0.6, y2-0.4, y2-0.4))
}

// summit adds the enclosed third level, its roof terrace, the campanile, the
// lighthouse cupola and the antenna.
func (t EiffelTower) summit(g *Grid, y3 float64) {
	c := P(0, 0)
	g.Use(Charcoal).Fill(BlockAt(c, 9.4, 9.4, y3-0.8, y3))
	g.Use(Glass).Fill(BlockAt(c, 8.8, 8.8, y3, y3+2.6))
	g.Carve(BlockAt(c, 6.8, 6.8, y3+0.4, y3+2.6))
	g.Use(Wood).Fill(Func{Box{V(-4.4, y3, -4.4), V(4.4, y3+2.6, 4.4)}, func(p Vec3) bool {
		ax, az := math.Abs(p.X), math.Abs(p.Z)
		return math.Max(ax, az) >= 3.4 && (math.Mod(math.Min(ax, az)+0.5, 2.4) < 0.9 || math.Abs(ax-az) < 0.8)
	}})
	g.Use(Charcoal).Fill(BlockAt(c, 10, 10, y3+2.6, y3+3.2))
	g.Fill(Subtract(BlockAt(c, 10, 10, y3+3.2, y3+4.2), BlockAt(c, 8.4, 8.4, y3+3, y3+5)))
	g.Carve(Func{Box{V(-5, y3+3.2, -5), V(5, y3+3.8, 5)}, func(p Vec3) bool {
		return math.Mod(math.Min(math.Abs(p.X), math.Abs(p.Z))+0.75, 1.8) > 1.0
	}})
	// Campanile: a small latticed box.
	const half = 2.6
	cy0, cy1 := y3+3.2, y3+7.4
	g.Use(Wood)
	t.raster(g, Box{V(-half, cy0, -half), V(half, cy1, half)}, func(p Vec3) (Color, bool) {
		ax, az := math.Abs(p.X), math.Abs(p.Z)
		v := p.Y - cy0
		if ax >= half-1 && t.member(p.Z+half, 2*half, v, cy1-cy0, false) {
			return Wood, true
		}
		if az >= half-1 && t.member(p.X+half, 2*half, v, cy1-cy0, true) {
			return Wood, true
		}
		return 0, false
	})
	g.Use(Charcoal).Fill(BlockAt(c, 6, 6, cy1, cy1+0.6))
	g.FillVolume(Frustum{c, 2.6, 1.3, cy1 + 0.6, cy1 + 2.2})
	g.Use(Gold).Fill(Cylinder{Center: c, Radius: 1.2, Y0: cy1 + 2.2, Y1: cy1 + 3.6})
	g.Use(Charcoal).FillVolume(Frustum{c, 1.6, 0.6, cy1 + 3.6, cy1 + 4.6})
	tip := eiffelBase + 120
	g.Use(White).Fill(Cylinder{Center: c, Radius: 0.6, Y0: cy1 + 4.6, Y1: tip})
	for y := cy1 + 6; y < tip-1; y += 3 {
		g.Paint(Red, Cylinder{Center: c, Radius: 1, Y0: y, Y1: y + 1})
	}
	g.Use(Red).Fill(Sphere(c.At(tip), 0.8))
}

// ground lays the site: esplanade, Champ de Mars, quays, the Seine and the bridge.
func (t EiffelTower) ground(g *Grid) {
	const x0, x1, zr = -76.0, 98.0, 38.0
	t.raster(g, Box{V(x0, -0.1, -zr), V(x1, 0.5, zr)}, func(p Vec3) (Color, bool) {
		x, z := p.X, p.Z
		az := math.Abs(z)
		switch {
		case x >= -44 && x <= -32: // The Seine, sunk below the quays.
			return 0, false
		case x < -44: // Trocadéro side.
			return t.trocaderoColor(x, z, p.Y)
		case x < -24: // Quai Branly: road and pavements.
			if x > -30.5 && x < -25.5 {
				return Charcoal, true
			}
			return Sand, true
		case x <= 24: // Esplanade and the gardens beside it.
			if az <= 24 {
				if math.Mod(math.Abs(x)+math.Abs(z), 8) < 0.8 {
					return Stone, true
				}
				return Sand, true
			}
			for _, sx := range []float64{-11, 11} {
				dx, dz := (x-sx)/7, (az-31)/3.5
				if dx*dx+dz*dz <= 1 {
					return Water, true
				}
			}
			if az < 26 || math.Abs(x) < 1.5 {
				return Sand, true
			}
			return Grass, true
		}
		// Champ de Mars, ending at the forecourt of the École Militaire.
		switch {
		case x > 86:
			return Sand, true
		case az > 34:
			return Sand, true
		case az > 30:
			return Charcoal, true
		case az > 26:
			return Grass, true
		case az > 15:
			return Sand, true
		case az > 13:
			return Foliage, true // Low clipped hedges framing the lawns.
		case x < 27 || math.Mod(x-27, 19) < 2 || az < 1:
			return Sand, true
		}
		return Grass, true
	})
	// Hedges stand one layer proud of the lawns.
	g.Use(Foliage).Fill(Func{Box{V(27, 0.6, -15), V(84, 1.0, 15)}, func(p Vec3) bool {
		return math.Abs(p.Z) > 13 && math.Mod(p.X-27, 19) >= 2
	}})
	// The Seine: water, quay walls and the Pont d'Iéna with three arches.
	g.Use(Water).Fill(Block(V(-44, -2.2, -zr), V(-32, -1.8, zr)))
	g.Use(Stone).Fill(Block(V(-44.5, -2.2, -zr), V(-43.5, 0.5, zr)), Block(V(-32.5, -2.2, -zr), V(-31.5, 0.5, zr)))
	g.Fill(Block(V(-44.5, -1.6, -5.5), V(-31.5, 0.5, 5.5)))
	for _, x := range []float64{-41.6, -38, -34.4} {
		g.Carve(Arch{V(x, -1.7, 0), 1.45, 0.4, 12, AxisZ})
	}
	g.Fill(Block(V(-44.5, 0.6, -5.5), V(-31.5, 1.4, -4.7)), Block(V(-44.5, 0.6, 4.7), V(-31.5, 1.4, 5.5)))
	g.Use(Sand).Fill(Block(V(-44, 0.4, -4.5), V(-32, 0.5, 4.5)))
	for _, x := range []float64{-40, -36} {
		for _, z := range []float64{-5.1, 5.1} {
			g.Use(Gold).Fill(BlockAt(P(x, z), 1, 1, 1.4, 2.6))
		}
	}
}

// trocaderoColor colors the gardens across the river: quay, lawns, paths, the basin.
func (EiffelTower) trocaderoColor(x, z, y float64) (Color, bool) {
	az := math.Abs(z)
	switch {
	case x > -49:
		if x > -47.5 && x < -45 {
			return Charcoal, true
		}
		return Sand, true
	case x >= -70 && x <= -50 && az <= 6.5: // Fountain basin, sunk one layer.
		if x < -69 || x > -51 || az > 5.5 {
			return Stone, true
		}
		if y < 0.2 {
			return Water, true
		}
		return 0, false
	case x < -70 || az < 9:
		return Sand, true
	case math.Mod(-x, 7) < 1.2:
		return Sand, true
	}
	return Grass, true
}

// trocadero adds the basin rim, fountain jets and the Palais de Chaillot wings.
func (t EiffelTower) trocadero(g *Grid) {
	g.Use(Stone).Fill(Subtract(BlockAt(P(-60, 0), 21, 13, 0.6, 1.0), BlockAt(P(-60, 0), 19, 11, 0, 2)))
	g.Use(White)
	for x := -68.0; x <= -52; x += 2 {
		h := 1.6 + 2.4*Hash01(int(x), 7)
		g.Fill(Cylinder{Center: P(x, 0), Radius: 0.6, Y0: 0, Y1: h})
	}
	// Arcing water cannons from both long sides into the axis.
	for x := -67.0; x <= -53; x += 4 {
		for _, s := range []float64{-1, 1} {
			var prev Vec3
			for i := 0; i <= 12; i++ {
				u := float64(i) / 12
				p := V(x+1.5*u, 0.8+4*u*(1-u)*1.2, s*5*(1-u))
				if i > 0 {
					g.Fill(Capsule{prev, p, 0.55})
				}
				prev = p
			}
		}
	}
	// Palais de Chaillot: two curved hollow wings embracing the gardens, with
	// arcaded galleries, and the end pavilions flanking the central esplanade.
	center, rOut, rIn := P(-48, 0), 27.0, 24.0
	wing := Func{Box{V(-76, 0.5, -30), V(-48, 6, 30)}, func(p Vec3) bool {
		d := p.Plan().Sub(center)
		r := d.Length()
		return p.X < center.X-8 && r <= rOut && r >= rIn && math.Abs(p.Z) >= 9
	}}
	g.Use(Stone).Fill(wing)
	g.Carve(Func{Box{V(-76, 0.6, -30), V(-48, 5, 30)}, func(p Vec3) bool {
		d := p.Plan().Sub(center)
		r := d.Length()
		return p.X < center.X-8 && r <= rOut-1 && r >= rIn+1 && math.Abs(p.Z) >= 10
	}})
	g.Carve(Func{Box{V(-76, 1.4, -30), V(-48, 4.2, 30)}, func(p Vec3) bool {
		d := p.Plan().Sub(center)
		r := d.Length()
		a := math.Atan2(math.Abs(d.Y), -d.X)
		return p.X < center.X-8 && r <= rIn+1 && r >= rIn-0.5 && math.Abs(p.Z) >= 10.5 && math.Mod(a*r, 2.2) < 1.1
	}})
	g.Use(Sand).Fill(Func{Box{V(-76, 6, -30), V(-48, 6.5, 30)}, func(p Vec3) bool {
		r := p.Plan().Sub(center).Length()
		return p.X < center.X-8 && r <= rOut+0.4 && r >= rIn-0.4 && math.Abs(p.Z) >= 9
	}})
	for _, s := range []float64{-1, 1} {
		t.hall(g, Block(V(-76, eiffelBase, math.Min(s*9, s*15)), V(-68, eiffelBase+7.5, math.Max(s*9, s*15))), 3.5)
		g.Use(Sand).Fill(BlockAt(P(-72, s*12), 8.6, 6.6, eiffelBase+7.5, eiffelBase+8))
	}
}

// trees plants alleys along the Champ de Mars, the quay and the Trocadéro gardens.
func (t EiffelTower) trees(g *Grid) {
	plant := func(x, z, h float64) {
		g.Build(kit.Tree{At: P(x, z), Base: eiffelBase, Height: h})
	}
	for x := 30.0; x <= 84; x += 4.5 {
		for _, z := range []float64{18.5, 23, -18.5, -23, 36.5, -36.5} {
			plant(x, z, 5+Hash01(int(x*2), int(z))*0.8)
		}
	}
	for z := -35.0; z <= 35; z += 4.4 {
		if math.Abs(z) > 7 {
			plant(-23, z, 5)
		}
	}
	for _, z := range []float64{-13, 13} {
		for x := -66.0; x <= -52; x += 4.5 {
			plant(x, z, 4.5)
		}
	}
	for i := 0; i < 16; i++ {
		x := -22 + 44*Hash01(i, 1)
		z := 27.5 + 9*Hash01(i, 2)
		if i%2 == 1 {
			z = -z
		}
		if math.Abs(math.Abs(x)-11) < 8 && math.Abs(math.Abs(z)-31) < 4.5 || math.Abs(x) < 2.5 {
			continue
		}
		plant(x, z, 4.5+1.5*Hash01(i, 3))
	}
}

// academy adds the École Militaire closing the Champ de Mars: a long two-storey
// range with a gray mansard roof and a domed, arcaded central pavilion.
func (t EiffelTower) academy(g *Grid) {
	const front, back = 88.0, 96.0
	for _, s := range []float64{-1, 1} {
		t.hall(g, Block(V(front, eiffelBase, math.Min(s*5, s*32)), V(back, eiffelBase+6, math.Max(s*5, s*32))), 3)
		g.Use(Slate).Fill(Gable{P(front-0.5, math.Min(s*5, s*32.5)), P(back+0.5, math.Max(s*5, s*32.5)), eiffelBase + 6, AxisZ})
	}
	// Central pavilion with an arcaded portico facing the tower.
	g.Use(Stone).Fill(Block(V(front-2, eiffelBase, -6), V(back, eiffelBase+9, 6)))
	g.Carve(Block(V(front-1, eiffelBase+0.5, -5), V(back-1, eiffelBase+8, 5)))
	for z := -4.0; z <= 4; z += 2.6 {
		g.Carve(Arch{V(front-2, eiffelBase, z), 0.8, 2.2, 3, AxisX})
		g.Paint(Glass, Arch{V(front-2, eiffelBase+5, z), 0.6, 1.2, 1, AxisX})
	}
	g.Use(Gold).Fill(Block(V(front-2.2, eiffelBase+8.4, -6.2), V(front-1.8, eiffelBase+9, 6.2)))
	g.Use(Slate).FillVolume(Subtract(Ellipsoid{V(front+4, eiffelBase+9, 0), V(4.5, 5, 4.5)}, Ellipsoid{V(front+4, eiffelBase+9, 0), V(3.3, 3.8, 3.3)}, Block(V(front-2, 0, -7), V(back+1, eiffelBase+9, 7))))
	g.Use(Gold).Fill(Cylinder{Center: P(front+4, 0), Radius: 0.6, Y0: eiffelBase + 13.8, Y1: eiffelBase + 15.5})
}

// hall is a hollow Stone block with rows of tall glazed windows on every face,
// one row per storey of the given height.
func (EiffelTower) hall(g *Grid, b Cuboid, storey float64) {
	g.Use(Stone).Fill(b)
	g.Carve(Block(b.Min.Add(V(1, 0.5, 1)), b.Max.Sub(V(1, 1, 1))))
	g.Paint(Glass, Func{Box(b), func(p Vec3) bool {
		edgeX := p.X <= b.Min.X+0.5 || p.X >= b.Max.X-0.5
		edgeZ := p.Z <= b.Min.Z+0.5 || p.Z >= b.Max.Z-0.5
		v := math.Mod(p.Y-b.Min.Y, storey)
		if edgeX == edgeZ || v < 1 || v > storey-0.8 || p.Y > b.Max.Y-1 {
			return false
		}
		along := p.Z - b.Min.Z
		if edgeZ {
			along = p.X - b.Min.X
		}
		return along > 1.5 && math.Mod(along, 2.2) < 1.0
	}})
}
