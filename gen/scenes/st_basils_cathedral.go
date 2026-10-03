package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// StBasilsCathedral is the Cathedral of Vasily the Blessed on Red Square: a brick-red
// cluster of nine churches on a high arcaded basement. The central church rises into a
// tall tented spire; four large octagonal and four smaller round chapels around it are
// crowned by drums and patterned onion domes. A tented bell tower stands at one corner,
// St. Basil's own chapel at another, and the whole group sits on a cobbled square.
type StBasilsCathedral struct{}

func init() { Register(StBasilsCathedral{}) }

func (StBasilsCathedral) Name() string  { return "st_basils_cathedral" }
func (StBasilsCathedral) Title() string { return "St Basils Cathedral" }

// stBasilsCathedralPattern selects the first of a dome's two colors from the plan angle
// around the dome's axis and its normalized height (0 at the neck, 1 at the tip).
type stBasilsCathedralPattern func(theta, t float64) bool

// stBasilsCathedralChapel describes one drum-and-dome chapel tower.
type stBasilsCathedralChapel struct {
	c          Vec2
	round      bool    // Round tower; otherwise octagonal.
	a          float64 // Apothem (octagon) or radius.
	top        float64 // Wall top.
	drumR      float64
	drumH      float64
	neck, R, H float64 // Onion profile.
	a1, b      Color   // Pattern color and base color.
	pat        stBasilsCathedralPattern
}

const stBasilsCathedralOct = 0.41421356237309503 // tan(π/8)

// norm is the octagonal (faces toward the axes and diagonals) or round distance from c.
func (StBasilsCathedral) norm(c Vec2, p Vec3, round bool) float64 {
	q := p.Plan().Sub(c)
	if round {
		return q.Length()
	}
	d := math.Max(math.Abs(q.X), math.Abs(q.Y))
	return math.Max(d, (math.Abs(q.X)+math.Abs(q.Y))/math.Sqrt2)
}

// shell is a vertical octagonal or round prism/frustum from apothem a0 at y0 to a1 at
// y1; thick > 0 keeps only the outer layer of that thickness.
func (s StBasilsCathedral) shell(c Vec2, round bool, a0, a1, y0, y1, thick float64) Solid {
	r := math.Max(a0, a1) / math.Cos(math.Pi/8)
	return Func{Box{V(c.X-r, y0, c.Y-r), V(c.X+r, y1, c.Y+r)}, func(p Vec3) bool {
		a := a0 + (a1-a0)*(p.Y-y0)/math.Max(y1-y0, 1e-9)
		n := s.norm(c, p, round)
		return n <= a && (thick <= 0 || n > a-thick)
	}}
}

// onion is a bulb dome: a quarter ellipse from the neck to the widest point, then an
// full shoulder that pinches into a concave spike.
func (StBasilsCathedral) onion(c Vec2, base, neck, R, H float64) Solid {
	const tm = 0.34
	return Func{Box{V(c.X-R, base, c.Y-R), V(c.X+R, base+H, c.Y+R)}, func(p Vec3) bool {
		t := (p.Y - base) / H
		var r float64
		if t < tm {
			u := (tm - t) / tm
			r = neck + (R-neck)*math.Sqrt(1-u*u)
		} else {
			v := (t - tm) / (1 - tm)
			r = R * math.Pow(1-v, 1.3) * (1 + 1.3*v)
		}
		return p.Plan().Sub(c).Length() <= r
	}}
}

// arch is an arched opening facing the plan angle, dist from c, offset along the face.
func (StBasilsCathedral) arch(c Vec2, dist, off, angle, sill, hw, spring, depth float64) Solid {
	return RotateY(Arch{V(c.X+dist, sill, c.Y+off), hw, spring, depth, AxisX}, c, angle)
}

// kokoshnik is a keel-arched gable plate standing in a wall plane, dist from c at the
// given angle; inset shrinks it to select the panel inside its rim.
func (StBasilsCathedral) kokoshnik(c Vec2, dist, angle, base, w, hs, depth, inset float64) Solid {
	R := w/2 - inset
	at := Polar(c, dist, angle)
	ext := math.Max(w/2, depth) + 0.5
	top := base + hs + 1.3*w/2
	return Func{Box{V(at.X-ext, base, at.Y-ext), V(at.X+ext, top, at.Y+ext)}, func(p Vec3) bool {
		q := p.Plan().Sub(c).Rotate(-angle)
		along, u, h := q.X-dist, q.Y, p.Y-base
		if math.Abs(along) > depth/2 || math.Abs(u) > R || h < inset*0.6 {
			return false
		}
		v := math.Abs(u) / R
		return h <= hs+R*(0.85*math.Sqrt(1-v*v)+0.55*math.Pow(1-v, 2.5))
	}}
}

// domePaint selects dome cells for a pattern.
func (StBasilsCathedral) domePaint(c Vec2, base, R, H float64, pat stBasilsCathedralPattern) Solid {
	return Func{Box{V(c.X-R, base+0.05, c.Y-R), V(c.X+R, base+H, c.Y+R)}, func(p Vec3) bool {
		q := p.Plan().Sub(c)
		return pat(math.Atan2(q.Y, q.X), (p.Y-base)/H)
	}}
}

func stBasilsCathedralSpiral(n, twist float64) stBasilsCathedralPattern {
	return func(theta, t float64) bool {
		return int(math.Floor(theta/(2*math.Pi)*n+t*twist))%2 == 0
	}
}

func stBasilsCathedralChevron(n, bands, amp float64) stBasilsCathedralPattern {
	return func(theta, t float64) bool {
		x := theta / (2 * math.Pi) * n
		tri := math.Abs(x-math.Floor(x)-0.5) * 2
		return int(math.Floor(t*bands+tri*amp))%2 == 0
	}
}

func stBasilsCathedralDiamond(n, rows float64) stBasilsCathedralPattern {
	return func(theta, t float64) bool {
		x := theta / (2 * math.Pi) * n
		y := t * rows
		return (int(math.Floor(x+y))+int(math.Floor(x-y)))%2 == 0
	}
}

func stBasilsCathedralStripes(n float64) stBasilsCathedralPattern {
	return func(theta, t float64) bool {
		return int(math.Floor(theta/(2*math.Pi)*n))%2 == 0
	}
}

// finial is a gold spike with a cross.
func (StBasilsCathedral) finial(g *Grid, c Vec2, y, height float64) {
	g.With(Gold, func() {
		g.Fill(Capsule{c.At(y - 0.6), c.At(y + height), 0.5})
		bar := y + height*0.7
		g.Fill(Capsule{V(c.X-1.1, bar, c.Y), V(c.X+1.1, bar, c.Y), 0.45})
	})
}

// lesenes selects narrow vertical strips at n plan angles (offset by phase strips).
func (StBasilsCathedral) lesenes(c Vec2, r, y0, y1, n, phase, width float64) Solid {
	return Func{Box{V(c.X-r, y0, c.Y-r), V(c.X+r, y1, c.Y+r)}, func(p Vec3) bool {
		q := p.Plan().Sub(c)
		x := math.Atan2(q.Y, q.X)/(2*math.Pi)*n + phase
		return math.Abs(x-math.Round(x)) < width
	}}
}

// chapel builds a tower from the deck up: walls, trim, windows, kokoshnik tiers, drum
// and the patterned onion dome.
func (s StBasilsCathedral) chapel(g *Grid, ch stBasilsCathedralChapel, base float64) {
	c, a, top := ch.c, ch.a, ch.top
	g.Use(Brick).Fill(s.shell(c, ch.round, a, a, base, top, 1.2))
	// White trim: a thin mid cornice and a top cornice.
	mid := base + (top-base)*0.5
	g.Use(White).Fill(
		s.shell(c, ch.round, a+0.35, a+0.35, mid, mid+0.5, 0.8),
		s.shell(c, ch.round, a+0.45, a+0.45, top-0.7, top, 1.0),
	)
	if !ch.round {
		for k := 0; k < 8; k++ {
			at := Polar(c, a/math.Cos(math.Pi/8)-0.25, math.Pi/8+float64(k)*math.Pi/4)
			g.Fill(Capsule{at.At(base), at.At(top), 0.45})
		}
	}
	// Windows: a lower and an upper row, alternating faces.
	for k := 0; k < 8; k++ {
		angle := float64(k) * math.Pi / 4
		for row, sill := range []float64{base + 2, mid + 1.6} {
			if (k+row)%2 == 1 {
				continue
			}
			if row == 0 {
				g.Paint(White, s.arch(c, a-0.6, 0, angle, sill-0.4, 1.15, 2.0, 1.6))
			}
			g.Carve(s.arch(c, a-0.6, 0, angle, sill, 0.75, 1.7, 3))
		}
	}
	// Sloped roof up to the drum, fronted by two staggered tiers of kokoshniks.
	roofTop := top + (a - ch.drumR)
	g.Use(Brick).Fill(s.shell(c, ch.round, a, ch.drumR, top, roofTop, 1.2))
	face := 2 * a * stBasilsCathedralOct
	if ch.round {
		face = 2 * math.Pi * a / 8
	}
	for k := 0; k < 8; k++ {
		angle := float64(k) * math.Pi / 4
		g.Use(White).Fill(s.kokoshnik(c, a-0.3, angle, top-0.3, face*0.98, 0.8, 1.2, 0))
		g.Paint(Brick, s.kokoshnik(c, a-0.3, angle, top-0.3, face*0.98, 0.8, 1.6, 0.5))
	}
	tier := (a + ch.drumR) / 2
	tierFace := 2 * tier * stBasilsCathedralOct
	for k := 0; k < 8; k++ {
		angle := math.Pi/8 + float64(k)*math.Pi/4
		b := top + (a-tier)*0.8 + 0.8
		g.Use(White).Fill(s.kokoshnik(c, tier, angle, b, tierFace, 0.5, 1.3, 0))
		g.Paint(Red, s.kokoshnik(c, tier, angle, b, tierFace, 0.5, 1.5, 0.6))
	}
	// Drum with white lesenes, cornices and slit windows.
	drumBase, drumTop := roofTop-1.5, roofTop+ch.drumH
	g.Use(Brick).Fill(Cylinder{c, ch.drumR, ch.drumR - 1.2, drumBase, drumTop})
	g.Use(White).Fill(Cylinder{c, ch.drumR + 0.45, ch.drumR - 0.8, drumTop - 0.7, drumTop})
	g.Paint(White, s.lesenes(c, ch.drumR, roofTop, drumTop, 8, 0.5, 0.07))
	for k := 0; k < 8; k += 2 {
		g.Carve(s.arch(c, ch.drumR-0.6, 0, float64(k)*math.Pi/4+math.Pi/8, roofTop+1.1, 0.55, ch.drumH-3, 3))
	}
	// Neck collar, dome and cross.
	g.Use(White).Fill(Cylinder{c, ch.neck + 0.3, 0, drumTop - 0.2, drumTop + 0.5})
	g.Use(ch.b).FillVolume(s.onion(c, drumTop, ch.neck, ch.R, ch.H))
	g.Paint(ch.a1, s.domePaint(c, drumTop+0.4, ch.R, ch.H*0.88, ch.pat))
	s.finial(g, c, drumTop+ch.H, 3)
}

// Build authors the cathedral on a finer grid and merges it, scaling it up by 1.15 so
// trim, kokoshniks and dome patterns get enough cells to read.
func (s StBasilsCathedral) Build(out *Grid) Order {
	g := NewGridWithSpan(out.Span() / 1.15)
	s.compose(g)
	out.Merge(g)
	return Sweep(P(0, 0), math.Pi/2)
}

func (s StBasilsCathedral) compose(g *Grid) {
	const (
		deck  = 7.0  // Top of the basement platform.
		half  = 22.0 // Platform half-size.
		cham  = 8.0  // Platform corner chamfer.
		plaza = 33.0 // Square half-size.
		gal   = 3.6  // Gallery depth.
	)
	center := P(0, 0)
	// inset is the distance inside the chamfered-square platform edge (negative outside).
	inset := func(p Vec3) float64 {
		x, z := math.Abs(p.X), math.Abs(p.Z)
		return math.Min(math.Min(half-x, half-z), (2*half-cham-x-z)/math.Sqrt2)
	}
	band := func(y0, y1, d0, d1 float64) Solid {
		return Func{Box{V(-half-1, y0, -half-1), V(half+1, y1, half+1)}, func(p Vec3) bool {
			d := inset(p)
			return d >= d0 && d <= d1
		}}
	}
	sides := []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2}
	diag := (2*half - cham) / math.Sqrt2

	// Cobbled square: dark setts with lighter stones and paving lines, a lawn around the
	// platform, and a dark floor inside the basement.
	ground := Box{V(-plaza, 0, -plaza), V(plaza, 0.4, plaza)}
	g.Use(DarkStone).Fill(Cuboid(ground))
	g.Paint(Stone, Func{ground, func(p Vec3) bool {
		i, j := int(math.Floor(p.X/1.6)), int(math.Floor(p.Z/1.6))
		lx, lz := math.Mod(math.Abs(p.X)+0.4, 8), math.Mod(math.Abs(p.Z)+0.4, 8)
		return Hash01(i, j, 7) < 0.15 || lx < 0.8 || lz < 0.8
	}})
	g.Paint(Grass, band(0, 0.4, -3.5, 0))
	g.Paint(Charcoal, band(0, 0.4, 1, half))

	// Basement: hollow brick walls, a white plinth and cornice, small arched windows.
	g.Use(Brick).Fill(band(0, deck, 0, 1.2))
	g.Use(Terracotta).Fill(band(deck-1, deck, 0, half))
	g.Paint(White, band(0, 0.9, -1, 1.3), band(deck-1.6, deck-0.9, -1, 1.3))
	for _, side := range sides {
		for off := -12.0; off <= 12; off += 4 {
			g.Paint(White, s.arch(center, half-0.6, off, side, 2.0, 1.2, 1.4, 1.6))
			g.Carve(s.arch(center, half-0.6, off, side, 2.6, 0.7, 1.0, 3))
		}
		for _, off := range []float64{-2.5, 2.5} {
			g.Paint(White, s.arch(center, diag-0.5, off, side+math.Pi/4, 2.0, 1.2, 1.4, 1.8))
			g.Carve(s.arch(center, diag-0.5, off, side+math.Pi/4, 2.6, 0.7, 1.0, 3.5))
		}
	}

	// The upper gallery: an arcaded corridor round the deck edge under a green roof.
	g.Use(Brick).Fill(band(deck-0.5, deck+4, 0, 1.1), band(deck, deck+4, gal-1.1, gal))
	g.Use(White).Fill(band(deck+3.3, deck+4, -0.3, 1.1))
	for _, side := range sides {
		for off := -14.0; off <= 14; off += 4 {
			g.Paint(White, s.arch(center, half-0.55, off, side, deck+0.5, 1.15, 1.9, 1.5))
			g.Carve(s.arch(center, half-0.55, off, side, deck+0.9, 0.75, 1.6, 1.5))
		}
		for _, off := range []float64{-3, 1} {
			at := off + 1
			g.Paint(White, s.arch(center, diag-0.5, at, side+math.Pi/4, deck+0.5, 1.15, 1.9, 1.7))
			g.Carve(s.arch(center, diag-0.5, at, side+math.Pi/4, deck+0.9, 0.75, 1.6, 1.7))
		}
	}
	g.Use(Foliage).Fill(Func{Box{V(-half-1, deck+3.5, -half-1), V(half+1, deck+6, half+1)}, func(p Vec3) bool {
		d := inset(p)
		if d < -0.6 || d > gal+0.4 {
			return false
		}
		roof := deck + 4 + 1.6 - math.Abs(d-gal/2)*0.75
		return p.Y <= roof && p.Y >= roof-1.1
	}})

	// Covered stairs climbing to the gallery, north (+Z) and south (-Z), with tented porches.
	for _, dir := range []float64{1, -1} {
		x0, x1 := 5.0*dir, 10.0*dir
		lo, hi := math.Min(x0, x1), math.Max(x0, x1)
		mx := (lo + hi) / 2
		zs := half * dir
		zLo, zHi := math.Min(zs, zs+deck*dir), math.Max(zs, zs+deck*dir)
		g.Use(Stone).Fill(Func{Box{V(lo, 0, zLo), V(hi, deck, zHi)}, func(p Vec3) bool {
			return p.Y <= deck-math.Floor((p.Z-zs)*dir)
		}})
		for _, x := range []float64{lo - 0.6, hi + 0.6} {
			g.Use(Brick).Fill(Func{Box{V(x-0.6, 0, zLo), V(x+0.6, deck+1.5, zHi)}, func(p Vec3) bool {
				return p.Y <= deck+1.2-(p.Z-zs)*dir
			}})
		}
		g.Paint(White, Func{Box{V(lo-1.3, 0, zLo), V(hi+1.3, deck+1.5, zHi)}, func(p Vec3) bool {
			return p.Y >= deck+0.5-(p.Z-zs)*dir && (p.X < lo || p.X > hi)
		}})
		// Doorway through the gallery.
		g.Carve(Arch{V(mx, deck, zs-dir*gal/2), 2.0, 1.8, gal + 1.5, AxisZ})
		porch := P(mx, zs+2.6*dir)
		for _, px := range []float64{lo - 0.6, hi + 0.6} {
			for _, pz := range []float64{zs + 0.6*dir, zs + 4.6*dir} {
				g.Use(Brick).Fill(Cylinder{P(px, pz), 0.6, 0, 0, deck + 4.5})
			}
		}
		g.Use(White).Fill(BlockAt(porch, hi-lo+2.8, 5.6, deck+4.3, deck+5))
		g.Use(Foliage).Fill(s.shell(porch, false, 3.4, 0.5, deck+5, deck+11.5, 1.1))
		g.Use(Gold).FillVolume(s.onion(porch, deck+11.2, 0.5, 1.0, 2.2))
	}

	// The central church: tall octagon, kokoshnik tiers, octagonal drum, tented spire.
	{
		a, top := 6.5, 25.0
		g.Use(Brick).Fill(s.shell(center, false, a, a, deck, top, 1.2))
		g.Use(White).Fill(
			s.shell(center, false, a+0.35, a+0.35, deck+9, deck+9.5, 0.8),
			s.shell(center, false, a+0.45, a+0.45, top-0.7, top, 1.0),
		)
		for k := 0; k < 8; k++ {
			at := Polar(center, a/math.Cos(math.Pi/8)-0.25, math.Pi/8+float64(k)*math.Pi/4)
			g.Fill(Capsule{at.At(deck), at.At(top), 0.55})
			angle := float64(k) * math.Pi / 4
			for _, sill := range []float64{deck + 4.5, deck + 11.5} {
				g.Paint(White, s.arch(center, a-0.6, 0, angle, sill-0.4, 1.2, 2.7, 1.6))
				g.Carve(s.arch(center, a-0.6, 0, angle, sill, 0.8, 2.4, 3))
			}
		}
		drum := 4.8
		g.Use(Brick).Fill(s.shell(center, false, a, drum, top, top+1.7, 1.2))
		face := 2 * a * stBasilsCathedralOct
		for k := 0; k < 8; k++ {
			angle := float64(k) * math.Pi / 4
			g.Use(White).Fill(s.kokoshnik(center, a-0.3, angle, top-0.3, face*0.98, 0.9, 1.2, 0))
			g.Paint(Brick, s.kokoshnik(center, a-0.3, angle, top-0.3, face*0.98, 0.9, 1.6, 0.5))
			angle += math.Pi / 8
			w := 2 * drum * stBasilsCathedralOct * 1.15
			g.Use(White).Fill(s.kokoshnik(center, drum+0.4, angle, top+1.4, w, 0.6, 1.3, 0))
			g.Paint(Red, s.kokoshnik(center, drum+0.4, angle, top+1.4, w, 0.6, 1.5, 0.6))
		}
		dTop := 33.0
		g.Use(Brick).Fill(s.shell(center, false, drum, drum, top, dTop, 1.2))
		g.Use(White).Fill(s.shell(center, false, drum+0.5, drum+0.5, dTop-0.8, dTop, 1.3))
		for k := 0; k < 8; k++ {
			at := Polar(center, drum/math.Cos(math.Pi/8)-0.25, math.Pi/8+float64(k)*math.Pi/4)
			g.Fill(Capsule{at.At(top + 2), at.At(dTop), 0.5})
			g.Carve(s.arch(center, drum-0.6, 0, float64(k)*math.Pi/4, top+4.3, 0.55, 2.4, 3))
		}
		// The tent: brick with white corner ribs, a few bands and gold studs.
		tTop := 53.0
		tb := Box{V(-drum-1, dTop, -drum-1), V(drum+1, tTop, drum+1)}
		g.Use(Brick).Fill(s.shell(center, false, drum+0.3, 1.1, dTop, tTop, 1.3))
		g.Paint(White, Func{tb, func(p Vec3) bool {
			t := (p.Y - dTop) / (tTop - dTop)
			q := p.Plan()
			x := math.Atan2(q.Y, q.X)/(2*math.Pi)*8 - 0.5 // Corners lie at π/8 offsets.
			return math.Mod(t*4, 1) < 0.07 || math.Abs(x-math.Round(x)) < 0.05+0.08*t
		}})
		g.Paint(Gold, Func{tb, func(p Vec3) bool {
			t := (p.Y - dTop) / (tTop - dTop)
			q := p.Plan()
			x := math.Atan2(q.Y, q.X) / (2 * math.Pi) * 8
			return math.Abs(math.Mod(t*8, 1)-0.5) < 0.12 && math.Abs(x-math.Round(x)) < 0.09 && t < 0.85
		}})
		g.Use(Brick).Fill(Cylinder{center, 1.4, 0, tTop - 0.5, tTop + 1.5})
		g.Use(White).Fill(Cylinder{center, 1.7, 0, tTop + 1.2, tTop + 1.9})
		g.Use(Gold).FillVolume(s.onion(center, tTop+1.9, 1.3, 2.5, 5.2))
		s.finial(g, center, tTop+7.1, 3.5)
	}

	// Four large octagonal chapels on the axes and four small round ones between.
	big := []stBasilsCathedralChapel{
		{c: P(0, 14), a1: Gold, b: Grass, pat: stBasilsCathedralSpiral(10, 3)},
		{c: P(14, 0), a1: Red, b: White, pat: stBasilsCathedralSpiral(10, -3)},
		{c: P(0, -14), a1: Foliage, b: Gold, pat: stBasilsCathedralDiamond(10, 4)},
		{c: P(-14, 0), a1: Navy, b: White, pat: stBasilsCathedralChevron(6, 4, 1.2)},
	}
	for _, ch := range big {
		ch.a, ch.top, ch.drumR, ch.drumH = 4.2, 21, 2.9, 5
		ch.neck, ch.R, ch.H = 2.5, 5.0, 10.5
		s.chapel(g, ch, deck)
	}
	small := []stBasilsCathedralChapel{
		{c: P(10.5, 10.5), a1: Red, b: Gold, pat: stBasilsCathedralStripes(8)},
		{c: P(10.5, -10.5), a1: Grass, b: Red, pat: stBasilsCathedralChevron(8, 4, 1)},
		{c: P(-10.5, -10.5), a1: Red, b: White, pat: stBasilsCathedralDiamond(8, 3)},
		{c: P(-10.5, 10.5), a1: Gold, b: Navy, pat: stBasilsCathedralSpiral(8, 2)},
	}
	for _, ch := range small {
		ch.round, ch.a, ch.top, ch.drumR, ch.drumH = true, 3.2, 17, 2.1, 3.6
		ch.neck, ch.R, ch.H = 1.8, 3.7, 8
		s.chapel(g, ch, deck)
	}

	// The bell tower at the south-east corner: square base, open octagonal belfry, tent.
	{
		bt, h := P(-21.5, -21.5), 13.5
		g.Use(Brick).Fill(BlockAt(bt, 7, 7, 0, h))
		g.Carve(BlockAt(bt, 4.6, 4.6, 0.5, h-1))
		g.Use(White).Fill(BlockAt(bt, 7.8, 7.8, h-0.8, h), BlockAt(bt, 7.6, 7.6, 0, 0.9))
		for _, dx := range []float64{-3.5, 3.5} {
			for _, dz := range []float64{-3.5, 3.5} {
				g.Fill(Capsule{V(bt.X+dx, 0, bt.Y+dz), V(bt.X+dx, h, bt.Y+dz), 0.5})
			}
		}
		for k := 0; k < 4; k++ {
			angle := float64(k) * math.Pi / 2
			for _, sill := range []float64{3, 8.5} {
				g.Paint(White, s.arch(bt, 3.0, 0, angle, sill-0.5, 1.25, 2.0, 1.6))
				g.Carve(s.arch(bt, 3.0, 0, angle, sill, 0.75, 1.7, 3))
			}
		}
		a := 3.2
		g.Use(Brick).Fill(s.shell(bt, false, a, a, h, h+7.5, 1.2))
		for k := 0; k < 8; k++ {
			g.Paint(White, s.arch(bt, a-0.6, 0, float64(k)*math.Pi/4, h+0.6, 1.25, 3.4, 1.6))
			g.Carve(s.arch(bt, a-0.6, 0, float64(k)*math.Pi/4, h+1, 0.85, 3.2, 3))
		}
		g.Use(Gold).Fill(Frustum{bt, 1.2, 0.5, h + 2.5, h + 4.5}, Cylinder{bt, 0.5, 0, h + 4.5, h + 7})
		g.Use(White).Fill(s.shell(bt, false, a+0.5, a+0.5, h+7.5, h+8.3, 1.4))
		t0, tTop := h+8.3, h+17.5
		g.Use(Grass).Fill(s.shell(bt, false, a+0.2, 0.8, t0, tTop, 1.2))
		g.Paint(White, s.lesenes(bt, 4, t0, tTop, 8, -0.5, 0.1))
		g.Use(Brick).Fill(Cylinder{bt, 0.9, 0, tTop - 0.5, tTop + 1})
		g.Use(Gold).FillVolume(s.onion(bt, tTop+1, 0.8, 1.7, 3.6))
		s.finial(g, bt, tTop+4.6, 2.5)
	}

	// St. Basil's chapel at the north-east corner: a cube with a dotted green dome.
	{
		sb, h := P(-21, 15), 13.0
		g.Use(Brick).Fill(BlockAt(sb, 7, 7, 0, h))
		g.Carve(BlockAt(sb, 4.6, 4.6, 0.5, h-1))
		g.Use(White).Fill(BlockAt(sb, 7.8, 7.8, h-0.8, h), BlockAt(sb, 7.6, 7.6, 0, 0.9))
		for k := 0; k < 4; k++ {
			angle := float64(k) * math.Pi / 2
			for _, sill := range []float64{2.5, 8} {
				g.Paint(White, s.arch(sb, 3.0, 0, angle, sill-0.5, 1.25, 2.2, 1.6))
				g.Carve(s.arch(sb, 3.0, 0, angle, sill, 0.75, 1.9, 3))
			}
			g.Use(White).Fill(s.kokoshnik(sb, 3.0, angle, h-0.2, 5.6, 0.5, 1.3, 0))
			g.Paint(Brick, s.kokoshnik(sb, 3.0, angle, h-0.2, 5.6, 0.5, 1.5, 0.6))
		}
		g.Use(Brick).Fill(HipRoof{P(sb.X-3.5, sb.Y-3.5), P(sb.X+3.5, sb.Y+3.5), h})
		g.Use(Brick).Fill(Cylinder{sb, 2.0, 0.8, h + 1, h + 4})
		g.Paint(White, s.lesenes(sb, 2.1, h+1, h+4, 8, 0.5, 0.08))
		g.Use(White).Fill(Cylinder{sb, 2.3, 1, h + 3.3, h + 4})
		g.Use(Grass).FillVolume(s.onion(sb, h+4, 1.8, 3.1, 6))
		g.Paint(Gold, s.domePaint(sb, h+4.4, 3.1, 5.3, stBasilsCathedralDiamond(8, 3)))
		s.finial(g, sb, h+10, 2.5)
	}

	// Monument to Minin and Pozharsky facing the cathedral from the north.
	{
		m := P(0, 30)
		g.Use(DarkStone).Fill(BlockAt(m, 6, 3.6, 0, 1))
		g.Use(Stone).Fill(BlockAt(m, 4.6, 2.6, 1, 4.5))
		g.Use(DarkStone).Fill(BlockAt(m, 5, 3, 4.5, 5))
		g.Use(Charcoal).Fill(
			Capsule{V(-0.8, 5.3, 30), V(-0.8, 8.6, 30), 0.65},
			Sphere(V(-0.8, 9.4, 30), 0.55),
			Capsule{V(-0.8, 8.2, 30), V(0.4, 10, 30.5), 0.4},
			Capsule{V(1.0, 5.3, 30), V(1.0, 7.3, 30), 0.65},
			Sphere(V(1.0, 8, 30), 0.5),
		)
	}
}
