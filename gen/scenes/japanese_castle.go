package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// JapaneseCastle is a Himeji-style hilltop castle: a five-tier white main keep with
// flared, upturned gray roofs, chidori and karahafu gables and gold shachihoko,
// linked to three smaller keeps by corridor turrets on battered ishigaki stone bases,
// inside a walled island compound with corner turrets, a gatehouse, a red bridge,
// a moat and pine, cherry and maple trees.
type JapaneseCastle struct{}

func init() { Register(JapaneseCastle{}) }

func (JapaneseCastle) Name() string  { return "japanese_castle" }
func (JapaneseCastle) Title() string { return "Japanese Castle" }

// japaneseCastleTier is one storey of a keep: plan half-sizes and wall height.
type japaneseCastleTier struct{ HX, HZ, H float64 }

// japaneseCastleBase is one rectangular piece of battered ishigaki masonry.
type japaneseCastleBase struct {
	C              Vec2
	HX, HZ         float64 // Top half-sizes.
	Y0, Y1, Batter float64 // Batter: extra half-size at the foot.
}

const (
	japaneseCastleIsland = 4.0  // Compound ground level.
	japaneseCastleEdge   = 33.0 // Island half-size (max-norm) at the top.
	japaneseCastleMoat   = 40.0 // Outer moat edge.
	japaneseCastleSite   = 47.0 // Outer ground half-size.
)

func (j JapaneseCastle) Build(g *Grid) Order {
	j.site(g)

	// Ishigaki: one masonry mass for the four keeps and their corridors.
	main := P(7, 7)
	west, northWest, north := P(-15, 5), P(-15, -15), P(5, -15)
	bases := []japaneseCastleBase{
		{main, 12.5, 10.5, japaneseCastleIsland, 16, 4.5},
		{west, 7, 6.5, japaneseCastleIsland, 12, 3},
		{northWest, 7, 6.5, japaneseCastleIsland, 12, 3},
		{north, 6.5, 6.5, japaneseCastleIsland, 12, 3},
		{P(-15, -5), 4, 5, japaneseCastleIsland, 12, 3},
		{P(-5, -15), 5, 4, japaneseCastleIsland, 12, 3},
		{P(5, -6), 4, 4, japaneseCastleIsland, 12, 3},
		{P(-7, 4), 4, 4, japaneseCastleIsland, 12, 3},
	}
	j.ishigaki(g, bases)

	// Main keep (tenshu): five tiers with an irimoya crown.
	tiers := []japaneseCastleTier{{11.5, 9.5, 7.5}, {10.4, 8.4, 6.5}, {9.2, 7.2, 6}, {7.9, 5.9, 5.5}, {6.4, 4.6, 5.5}}
	ys := j.keep(g, main, 16, tiers, 2.3, 2.0, AxisX)
	// Gables on the main keep's roofs (faces: 0 +Z, 1 +X, 2 -Z, 3 -X).
	for _, f := range []int{0, 2} {
		j.chidori(g, main, f, 0, 4.2, ys[1]-0.3, tiers[1].HZ-1, tiers[0].HZ+2)
	}
	for _, f := range []int{1, 3} {
		j.chidori(g, main, f, 0, 3.2, ys[1]-0.3, tiers[1].HX-1, tiers[0].HX+2)
	}
	j.karahafu(g, main, 0, 0, 3.4, ys[2]-0.3, tiers[2].HZ-1, tiers[1].HZ+2, 2.2)
	for _, f := range []int{1, 3} {
		j.chidori(g, main, f, 0, 2.8, ys[2]-0.3, tiers[2].HX-1, tiers[1].HX+2)
	}
	for _, off := range []float64{-3.6, 3.6} {
		j.chidori(g, main, 0, off, 2.3, ys[3]-0.3, tiers[3].HZ-1, tiers[2].HZ+2)
		j.chidori(g, main, 2, off, 2.3, ys[3]-0.3, tiers[3].HZ-1, tiers[2].HZ+2)
	}
	j.karahafu(g, main, 1, 0, 2.6, ys[4]-0.3, tiers[4].HX-1, tiers[3].HX+2, 1.8)
	j.chidori(g, main, 0, 0, 2.4, ys[4]-0.3, tiers[4].HZ-1, tiers[3].HZ+2)

	// Smaller keeps.
	small := []japaneseCastleTier{{6, 5.5, 6}, {4.8, 4.3, 5}, {3.6, 3.1, 4.8}}
	ws := j.keep(g, west, 12, small, 1.9, 1.6, AxisZ)
	j.chidori(g, west, 0, 0, 2.6, ws[1]-0.3, small[1].HZ-1, small[0].HZ+1.5)
	j.chidori(g, west, 3, 0, 2.6, ws[1]-0.3, small[1].HX-1, small[0].HX+1.5)
	ns := j.keep(g, northWest, 12, small, 1.9, 1.6, AxisX)
	j.chidori(g, northWest, 3, 0, 2.6, ns[1]-0.3, small[1].HX-1, small[0].HX+1.5)
	j.karahafu(g, northWest, 2, 0, 2.4, ns[1]-0.3, small[1].HZ-1, small[0].HZ+1.5, 1.6)
	little := []japaneseCastleTier{{5.5, 5.5, 5.5}, {4.3, 4.3, 4.8}, {3.2, 3.2, 4.6}}
	es := j.keep(g, north, 12, little, 1.9, 1.6, AxisZ)
	j.chidori(g, north, 1, 0, 2.4, es[1]-0.3, little[1].HX-1, little[0].HX+1.5)
	j.chidori(g, north, 2, 0, 2.4, es[1]-0.3, little[1].HZ-1, little[0].HZ+1.5)

	// Watari-yagura corridors between the keeps.
	j.keep(g, P(-15, -5), 12, []japaneseCastleTier{{3, 5, 4.8}}, 1.5, 1.3, AxisZ)
	j.keep(g, P(-5, -15), 12, []japaneseCastleTier{{5, 3, 4.8}}, 1.5, 1.3, AxisX)
	j.keep(g, P(5, -6), 12, []japaneseCastleTier{{3, 4, 4.8}}, 1.5, 1.3, AxisZ)
	j.keep(g, P(-7, 4), 12, []japaneseCastleTier{{4, 3, 4.8}}, 1.5, 1.3, AxisX)

	j.compound(g)
	j.bridge(g)
	j.trees(g)
	return Sweep(P(0, 0), math.Pi/2)
}

// cheb is the max-norm outward distance from a rectangle, with its per-axis parts.
func (JapaneseCastle) cheb(c Vec2, hx, hz float64, x, z float64) (u, v, d float64) {
	u, v = math.Abs(x-c.X)-hx, math.Abs(z-c.Y)-hz
	return u, v, math.Max(u, v)
}

func (JapaneseCastle) clamp01(v float64) float64 { return math.Max(0, math.Min(1, v)) }

// site lays the outer grass bank, the moat water and the island with its stone
// revetment and gravel surface.
func (j JapaneseCastle) site(g *Grid) {
	o := P(0, 0)
	ring := func(lo, hi, y0, y1 float64) Func {
		return Func{Box{V(-hi, y0, -hi), V(hi, y1, hi)}, func(p Vec3) bool {
			_, _, d := j.cheb(o, 0, 0, p.X, p.Z)
			return d >= lo && d <= hi
		}}
	}
	g.Use(Grass).Fill(ring(japaneseCastleMoat, japaneseCastleSite, 0, 1))
	// Outer moat revetment.
	g.Use(DarkStone).Fill(ring(japaneseCastleMoat-0.8, japaneseCastleMoat+0.4, 0, 1.4))
	g.Use(Water).Fill(ring(japaneseCastleEdge, japaneseCastleMoat-0.8, 0, 0.5))
	// Island revetment: battered masonry.
	edge := Func{Box{V(-japaneseCastleEdge-2, 0, -japaneseCastleEdge-2), V(japaneseCastleEdge+2, japaneseCastleIsland, japaneseCastleEdge+2)}, func(p Vec3) bool {
		_, _, d := j.cheb(o, japaneseCastleEdge, japaneseCastleEdge, p.X, p.Z)
		b := 1.6 * math.Pow((japaneseCastleIsland-p.Y)/japaneseCastleIsland, 1.5)
		return d <= b && d >= b-1.5
	}}
	g.Use(Stone).FillVolume(edge)
	j.masonry(g, edge)
	// Island surface: gravel with lawns.
	inner := japaneseCastleEdge - 1
	g.Use(Grass).Fill(Block(V(-inner, japaneseCastleIsland-0.5, -inner), V(inner, japaneseCastleIsland-0.5, inner)))
	g.Paint(Sand, Block(V(-26, japaneseCastleIsland-0.5, -26), V(23, japaneseCastleIsland-0.5, 23)))
	g.Paint(Sand, Block(V(-3, japaneseCastleIsland-0.5, 20), V(3, japaneseCastleIsland-0.5, inner)))
}

// masonry speckles stone courses with darker blocks in a running bond.
func (JapaneseCastle) masonry(g *Grid, s Solid) {
	g.Paint(DarkStone, Func{s.Bounds().Grow(0.5), func(p Vec3) bool {
		course := int(math.Floor(p.Y / 0.9))
		shift := float64(course%2) * 0.9
		ix, iz := int(math.Floor((p.X+shift)/1.8)), int(math.Floor((p.Z+shift)/1.8))
		return s.Contains(p) && Hash01(ix, course, iz) < 0.32
	}})
}

// ishigaki builds the union of battered stone pieces as a hollow shell whose tops are
// capped, with a concave (fan-shaped) slope: vertical at the top, flaring at the foot.
func (j JapaneseCastle) ishigaki(g *Grid, bases []japaneseCastleBase) {
	box := EmptyBox()
	for _, b := range bases {
		r := b.Batter + 0.5
		box = box.Union(Box{V(b.C.X-b.HX-r, b.Y0, b.C.Y-b.HZ-r), V(b.C.X+b.HX+r, b.Y1, b.C.Y+b.HZ+r)})
	}
	in := func(p Vec3, inset, cap float64) bool {
		for _, b := range bases {
			if p.Y < b.Y0 || p.Y > b.Y1-cap {
				continue
			}
			_, _, d := j.cheb(b.C, b.HX, b.HZ, p.X, p.Z)
			if d <= b.Batter*math.Pow((b.Y1-p.Y)/(b.Y1-b.Y0), 1.7)-inset {
				return true
			}
		}
		return false
	}
	shell := Func{box, func(p Vec3) bool { return in(p, 0, 0) && !in(p, 1.6, 1.2) }}
	g.Use(Stone).FillVolume(shell)
	j.masonry(g, shell)
	// Dark capstones along each top edge.
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		for _, b := range bases {
			_, _, d := j.cheb(b.C, b.HX, b.HZ, p.X, p.Z)
			if p.Y > b.Y1-0.9 && p.Y <= b.Y1 && d > -1.2 {
				return true
			}
		}
		return false
	}})
}

// eaveSurface is the flared roof skirt height above a plan point: from yRidge at
// `inner` units inside the rectangle down a concave curve to yEave at `over` units
// outside, with corners swept up by `up`.
func (j JapaneseCastle) eaveSurface(c Vec2, hx, hz, inner, over, yEave, yRidge, up, x, z float64) (s, u, v, d float64, ok bool) {
	u, v, d = j.cheb(c, hx, hz, x, z)
	if d < -inner || d > over {
		return 0, u, v, d, false
	}
	t := (d + inner) / (inner + over)
	s = yEave + (yRidge-yEave)*math.Pow(1-t, 1.8)
	ext := hx + over
	if u > v {
		ext = hz + over
	}
	near := j.clamp01(1 - (d-math.Min(u, v))/ext)
	s += up * math.Pow(near, 3) * (0.3 + 0.7*t)
	return s, u, v, d, true
}

// keep stacks white tiers with flared eaves and crowns them with an irimoya roof whose
// ridge runs along the given axis. It returns the tier base heights plus the top.
func (j JapaneseCastle) keep(g *Grid, c Vec2, base float64, tiers []japaneseCastleTier, over, rise float64, ridge Axis) []float64 {
	ys := []float64{base}
	for _, t := range tiers {
		ys = append(ys, ys[len(ys)-1]+t.H)
	}
	for i, t := range tiers {
		g.Use(White).Fill(Subtract(BlockAt(c, 2*t.HX, 2*t.HZ, ys[i], ys[i+1]), BlockAt(c, 2*t.HX-2, 2*t.HZ-2, ys[i]-1, ys[i+1]+1)))
		lo := ys[i] + 1.4
		if i > 0 {
			lo = ys[i] + rise + 0.4
		}
		hi := math.Min(ys[i+1]-0.9, lo+1.7)
		if hi-lo >= 0.9 {
			j.windows(g, c, t.HX, t.HZ, lo, hi)
		}
		if ys[i+1]-1.2-(hi+1.3) >= 1.3 {
			j.windows(g, c, t.HX, t.HZ, hi+1.3, math.Min(ys[i+1]-1.2, hi+3))
		}
	}
	up := over * 0.6
	for i, t := range tiers {
		top := ys[i+1]
		inner, r := math.Min(t.HX, t.HZ)*0.45, rise*0.8
		if i+1 < len(tiers) {
			inner, r = math.Max(t.HX-tiers[i+1].HX, t.HZ-tiers[i+1].HZ)+0.4, rise
		}
		j.eave(g, c, t.HX, t.HZ, inner, over, top-0.4, top+r, up)
	}
	// Irimoya crown: a gable roof over the top storey with recessed white gable ends.
	t, top := tiers[len(tiers)-1], ys[len(ys)-1]
	hx, hz := t.HX-1, t.HZ+0.3
	if ridge == AxisZ {
		hx, hz = t.HX+0.3, t.HZ-1
	}
	lo, hi := P(c.X-hx, c.Y-hz), P(c.X+hx, c.Y+hz)
	roof := Gable{lo, hi, top + 0.4, ridge}
	body := roof
	if ridge == AxisX {
		body.Min.X, body.Max.X = lo.X+0.8, hi.X-0.8
	} else {
		body.Min.Y, body.Max.Y = lo.Y+0.8, hi.Y-0.8
	}
	g.Use(White).Fill(body)
	g.Use(Slate).Fill(Subtract(roof, Translate(roof, V(0, -1.3, 0))))
	half := roof.Bounds().Max.Y
	var a, b Vec2
	if ridge == AxisX {
		a, b = P(lo.X+0.3, c.Y), P(hi.X-0.3, c.Y)
	} else {
		a, b = P(c.X, lo.Y+0.3), P(c.X, hi.Y-0.3)
	}
	g.Use(Charcoal).Fill(Capsule{a.At(half - 0.2), b.At(half - 0.2), 0.7})
	// Gold shachihoko: tail-up dolphins at each end of the ridge.
	size := 0.7 + 0.1*float64(len(tiers))
	for _, end := range []Vec2{a, b} {
		out := end.Sub(c)
		out = out.Scale(1 / out.Length())
		foot := end.Sub(out.Scale(0.4))
		g.Use(Gold).Fill(
			Capsule{foot.At(half + 0.2), foot.Add(out.Scale(0.5 * size)).At(half + 1.6*size), 0.6},
			Capsule{foot.Add(out.Scale(0.5 * size)).At(half + 1.6*size), foot.Sub(out.Scale(0.6 * size)).At(half + 2.4*size), 0.55},
		)
		// Gegyo pendant on the gable end below.
		gable := end.Add(out.Scale(-0.6))
		g.Fill(Sphere(gable.At(half-1.6), 0.6))
	}
	return ys
}

// eave fills one flared hip-roof skirt with dark hip ridges and eave lip.
func (j JapaneseCastle) eave(g *Grid, c Vec2, hx, hz, inner, over, yEave, yRidge, up float64) {
	r := over + 0.5
	box := Box{V(c.X-hx-r, yEave-2, c.Y-hz-r), V(c.X+hx+r, yRidge+up+1, c.Y+hz+r)}
	g.Use(Slate).FillVolume(Func{box, func(p Vec3) bool {
		s, _, _, _, ok := j.eaveSurface(c, hx, hz, inner, over, yEave, yRidge, up, p.X, p.Z)
		return ok && p.Y <= s && p.Y >= s-1.4
	}})
	// Plastered (white) hip ridges and a dark tile lip along the eave.
	g.Use(White).Fill(Func{box, func(p Vec3) bool {
		s, u, v, d, ok := j.eaveSurface(c, hx, hz, inner, over, yEave, yRidge, up, p.X, p.Z)
		return ok && p.Y >= s-0.6 && p.Y <= s+0.5 && math.Abs(u-v) < 0.45 && d > -inner+0.6 && d < over-1.1
	}})
	// Dark end tiles along the eave and onigawara at the swept-up corner tips.
	g.Use(Charcoal).Fill(Func{box, func(p Vec3) bool {
		s, u, v, d, ok := j.eaveSurface(c, hx, hz, inner, over, yEave, yRidge, up, p.X, p.Z)
		lip := d > over-0.4 && p.Y >= s-1.4 && p.Y <= s+0.2
		tip := d >= over-1.1 && math.Abs(u-v) < 0.7 && p.Y >= s-1.0 && p.Y <= s+0.6
		return ok && (lip || tip)
	}})
}

// windows paints dark latticed window slots along all four walls.
func (JapaneseCastle) windows(g *Grid, c Vec2, hx, hz, y0, y1 float64) {
	const pitch, w = 3.0, 1.4
	for _, axis := range []Axis{AxisX, AxisZ} {
		length, depth := hx, hz
		if axis == AxisZ {
			length, depth = hz, hx
		}
		n := int((2*length - 2) / pitch)
		if n < 1 {
			n = 1
		}
		for k := 0; k < n; k++ {
			pos := (float64(k) - float64(n-1)/2) * pitch
			for _, side := range []float64{-1, 1} {
				var box Cuboid
				if axis == AxisX {
					box = Block(V(c.X+pos-w/2, y0, c.Y+side*(depth-1.1)), V(c.X+pos+w/2, y1, c.Y+side*(depth+0.2)))
				} else {
					box = Block(V(c.X+side*(depth-1.1), y0, c.Y+pos-w/2), V(c.X+side*(depth+0.2), y1, c.Y+pos+w/2))
				}
				g.Paint(Charcoal, box)
			}
		}
	}
}

// face turns a solid authored on the +Z face of center c to face f (0 +Z, 1 +X, 2 -Z, 3 -X).
func (JapaneseCastle) face(s Solid, c Vec2, f int) Solid {
	return RotateY(s, c, []float64{0, -math.Pi / 2, math.Pi, math.Pi / 2}[f])
}

// chidori adds a triangular dormer gable on a roof skirt: it runs out from inside the
// upper wall (in) to the eave (out), with a recessed white face and a gold pendant.
func (j JapaneseCastle) chidori(g *Grid, c Vec2, f int, off, w, base, in, out float64) {
	x := c.X + off
	roof := Gable{P(x-w-0.5, c.Y+in), P(x+w+0.5, c.Y+out), base - 0.5, AxisZ}
	body := Gable{P(x-w, c.Y+in), P(x+w, c.Y+out-0.8), base, AxisZ}
	g.Use(White).Fill(j.face(body, c, f))
	g.Use(Slate).Fill(j.face(Subtract(roof, Translate(roof, V(0, -1.2, 0))), c, f))
	ridge := base + w
	g.Use(White).Fill(j.face(Capsule{V(x, ridge-0.1, c.Y+in), V(x, ridge-0.1, c.Y+out-0.3), 0.55}, c, f))
	g.Use(Gold).Fill(j.face(Sphere(V(x, ridge-1.3, c.Y+out-0.9), 0.55), c, f))
}

// karahafu adds a bell-curved gable with swept-up ends on a roof skirt.
func (j JapaneseCastle) karahafu(g *Grid, c Vec2, f int, off, w, base, in, out, h float64) {
	x := c.X + off
	curve := func(px float64) float64 {
		a := math.Abs(px-x) / w
		return base + h*(0.5+0.5*math.Cos(math.Pi*math.Min(a, 1))) + 0.9*math.Pow(math.Max(0, a-0.7)/0.3, 2)
	}
	box := Box{V(x-w-0.5, base-1.5, c.Y+in), V(x+w+0.5, base+h+1, c.Y+out)}
	body := Func{box, func(p Vec3) bool {
		return math.Abs(p.X-x) <= w-0.6 && p.Y >= base-1 && p.Y <= curve(p.X)-1 && p.Z <= c.Y+out-0.8
	}}
	shell := Func{box, func(p Vec3) bool {
		if math.Abs(p.X-x) > w+0.4 {
			return false
		}
		s := curve(math.Max(x-w, math.Min(x+w, p.X)))
		return p.Y <= s && p.Y >= s-1.2
	}}
	g.Use(White).Fill(j.face(body, c, f))
	g.Use(Slate).FillVolume(j.face(shell, c, f))
	g.Use(Charcoal).Fill(j.face(Func{box, func(p Vec3) bool {
		s := curve(math.Max(x-w, math.Min(x+w, p.X)))
		return math.Abs(p.X-x) <= w+0.4 && p.Z >= c.Y+out-0.7 && p.Y <= s+0.2 && p.Y >= s-1.2
	}}, c, f))
	g.Use(Gold).Fill(j.face(Sphere(V(x, base+h-1.6, c.Y+out-0.9), 0.55), c, f))
}

// compound builds the plastered dobei walls on the island edge with corner turrets
// and the main gatehouse facing the bridge on +Z.
func (j JapaneseCastle) compound(g *Grid) {
	const y0, y1 = japaneseCastleIsland, japaneseCastleIsland + 3.2
	r := japaneseCastleEdge - 2
	gate := 7.0
	for _, side := range []float64{-1, 1} {
		for _, axis := range []Axis{AxisX, AxisZ} {
			segs := [][2]float64{{-r, r}}
			if axis == AxisX && side > 0 {
				segs = [][2]float64{{-r, -gate}, {gate, r}}
			}
			for _, s := range segs {
				var wall Cuboid
				var cap Gable
				if axis == AxisX {
					wall = Block(V(s[0], y0, side*r-0.5), V(s[1], y1, side*r+0.5))
					cap = Gable{P(s[0]-0.5, side*r-1.3), P(s[1]+0.5, side*r+1.3), y1 - 0.4, AxisX}
				} else {
					wall = Block(V(side*r-0.5, y0, s[0]), V(side*r+0.5, y1, s[1]))
					cap = Gable{P(side*r-1.3, s[0]-0.5), P(side*r+1.3, s[1]+0.5), y1 - 0.4, AxisZ}
				}
				g.Use(White).Fill(wall)
				g.Paint(DarkStone, Block(wall.Min, V(wall.Max.X, y0+0.6, wall.Max.Z)))
				g.Use(Slate).Fill(cap)
				// Loopholes (sama): alternating square and triangular openings.
				for k := s[0] + 2; k < s[1]-1; k += 2.6 {
					size := 0.5
					if int(math.Round(k))%2 == 0 {
						size = 0.7
					}
					var hole Cuboid
					if axis == AxisX {
						hole = Block(V(k-size, y0+1.4, side*r-1), V(k+size, y0+2.3, side*r+1))
					} else {
						hole = Block(V(side*r-1, y0+1.4, k-size), V(side*r+1, y0+2.3, k+size))
					}
					g.Paint(Charcoal, hole)
				}
			}
		}
	}
	// Corner turrets (sumi-yagura), two storeys flush with the island edge.
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			at := P(sx*(japaneseCastleEdge-4.2), sz*(japaneseCastleEdge-4.2))
			axis := AxisX
			if sx*sz < 0 {
				axis = AxisZ
			}
			j.keep(g, at, japaneseCastleIsland, []japaneseCastleTier{{4, 4, 5.5}, {3, 3, 4.5}}, 1.5, 1.3, axis)
		}
	}
	// Gatehouse (yagura-mon): stone piers carrying a white storey over the passage.
	gz := r
	g.Use(Stone)
	for _, sx := range []float64{-1, 1} {
		pier := Block(V(sx*2.6, y0, gz-2.6), V(sx*7.5, y0+4.5, gz+2.6))
		g.Fill(pier)
		j.masonry(g, pier)
	}
	g.Use(Wood).Fill(Block(V(-2.6, y0+4, gz-2.6), V(2.6, y0+4.5, gz+2.6)))
	ys := j.keep(g, P(0, gz), y0+4.5, []japaneseCastleTier{{7.5, 3.2, 4.4}}, 1.8, 1.5, AxisX)
	j.chidori(g, P(0, gz), 0, 0, 2.4, ys[1]-0.3, 1.2, 3.2+1.4)
	// Open gate leaves.
	g.Use(Wood).Fill(Block(V(-2.6, y0, gz+1.5), V(-1.9, y0+3.8, gz+2.4)), Block(V(1.9, y0, gz+1.5), V(2.6, y0+3.8, gz+2.4)))
}

// bridge spans the moat on +Z: a gently arched wooden deck with red railings and piers.
func (j JapaneseCastle) bridge(g *Grid) {
	z0, z1 := japaneseCastleEdge-1, japaneseCastleMoat+1.5
	deck := func(z float64) float64 {
		t := j.clamp01((z - z0) / (z1 - z0))
		return japaneseCastleIsland - 0.2 + (1-japaneseCastleIsland)*t + 1.2*math.Sin(math.Pi*t)
	}
	box := Box{V(-3.2, 0, z0), V(3.2, japaneseCastleIsland+3, z1)}
	g.Use(Wood).Fill(Func{box, func(p Vec3) bool {
		s := deck(p.Z)
		return math.Abs(p.X) <= 2.6 && p.Y <= s && p.Y >= s-0.9
	}})
	for _, z := range []float64{z0 + 3.5, z0 + 6.5} {
		for _, x := range []float64{-2, 2} {
			g.Fill(Cylinder{Center: P(x, z), Radius: 0.6, Y0: 0, Y1: deck(z)})
		}
	}
	g.Use(Red).Fill(Func{box, func(p Vec3) bool {
		s := deck(p.Z)
		if math.Abs(p.X) < 2.2 || math.Abs(p.X) > 3.2 || p.Y < s || p.Z > z1-0.6 || p.Z < z0+1.2 {
			return false
		}
		post := math.Mod(p.Z-z0, 1.8) < 0.7
		return p.Y >= s+0.9 && p.Y <= s+1.5 || post && p.Y <= s+1.5
	}})
}

// trees scatters cloud-pruned pines, cherries and maples inside and outside the moat.
func (j JapaneseCastle) trees(g *Grid) {
	ground := japaneseCastleIsland - 0.3
	type spot struct {
		at   Vec2
		base float64
	}
	var spots []spot
	for _, p := range []Vec2{P(-7, 25.5), P(9, 26), P(17, 25.5), P(26.5, 17), P(27, 8), P(26.5, -2), P(27, -12), P(25, -24),
		P(-27, 14), P(-27.5, -3), P(-22, 25.5), P(-14, 26), P(14, -26), P(-5, -27), P(-21, -26)} {
		spots = append(spots, spot{p, ground})
	}
	outer := (japaneseCastleMoat + japaneseCastleSite) / 2
	for k := -4; k <= 4; k++ {
		t := float64(k) * 10
		for _, p := range []Vec2{P(t, outer), P(t, -outer), P(outer, t), P(-outer, t)} {
			if math.Abs(p.X) < 6 && p.Y > 0 {
				continue // Keep the bridge approach clear.
			}
			if Hash01(int(p.X), int(p.Y), 7) < 0.25 {
				continue
			}
			spots = append(spots, spot{p.Add(P(Hash01(int(p.X), 1)*3-1.5, Hash01(int(p.Y), 2)*3-1.5)), 0.7})
		}
	}
	for i, s := range spots {
		kind := Hash01(i, 11)
		h := 5 + 2.5*Hash01(i, 12)
		switch {
		case kind < 0.5:
			j.pine(g, s.at, s.base, h, i)
		case kind < 0.8:
			j.blossom(g, s.at, s.base, h*0.85, White, Red, 0.15)
		default:
			j.blossom(g, s.at, s.base, h*0.85, Red, Gold, 0.2)
		}
	}
}

// pine is a leaning trunk carrying flat, cloud-pruned foliage pads.
func (JapaneseCastle) pine(g *Grid, at Vec2, base, h float64, seed int) {
	lean := Polar(P(0, 0), h*0.18, Hash01(seed, 3)*2*math.Pi)
	top := at.Add(lean)
	g.Use(Wood).Fill(Capsule{at.At(base), top.At(base + h*0.85), 0.65})
	for k, f := range []float64{0.5, 0.72, 0.95} {
		side := Polar(P(0, 0), 1.2, Hash01(seed, k, 5)*2*math.Pi)
		c := at.Add(lean.Scale(f)).Add(side.Scale(1 - f))
		r := h * (0.5 - 0.22*f)
		g.Use(Foliage).Fill(Ellipsoid{c.At(base + h*f), V(r, 0.9, r)})
	}
}

// blossom is a round-crowned tree speckled with a second color.
func (JapaneseCastle) blossom(g *Grid, at Vec2, base, h float64, main, speckle Color, share float64) {
	g.Use(Wood).Fill(Capsule{at.At(base), at.At(base + h*0.55), 0.6})
	crown := Ellipsoid{at.At(base + h*0.7), V(h*0.42, h*0.3, h*0.42)}
	g.Use(main).Fill(crown)
	g.Paint(speckle, Func{crown.Bounds(), func(p Vec3) bool {
		return crown.Contains(p) && Hash01(int(math.Round(p.X*3)), int(math.Round(p.Y*3)), int(math.Round(p.Z*3))) < share
	}})
}
