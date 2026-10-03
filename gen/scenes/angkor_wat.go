package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// AngkorWat is the Khmer temple-mountain: three concentric galleried enclosures rising
// in terraces to the quincunx of lotus-bud towers, the central one tallest. The west
// entrance faces +Z: a naga causeway crosses the moat to the long west gopura, then the
// raised inner causeway passes the two libraries and reflecting ponds to the cruciform
// Terrace of Honour. A laterite wall and jungle grounds surround it.
type AngkorWat struct{}

func init() { Register(AngkorWat{}) }

func (AngkorWat) Name() string  { return "angkor_wat" }
func (AngkorWat) Title() string { return "Angkor Wat" }

const (
	angkorWatGround = 0.5  // Ground top.
	angkorWatL1     = 2.5  // Third-enclosure terrace top.
	angkorWatL2     = 7.5  // Second-level terrace top.
	angkorWatL3     = 17.0 // Bakan (upper level) top.
	angkorWatWallX  = 34.0 // Outer enclosure half-width.
	angkorWatWallE  = -31.0
	angkorWatWallW  = 56.0
	angkorWatMoat   = 8.0 // Moat width.
)

// angkorWatKind is what stands along one long side of a gallery segment.
type angkorWatKind int

const (
	angkorWatOpen angkorWatKind = iota
	angkorWatColonnade
	angkorWatWall
	angkorWatWindows
)

// angkorWatSeg is a straight vaulted gallery running along Axis from A0 to A1, centered
// across at C, D deep, with walls from Y0 to Y0+H and a corbelled vault above. Neg is
// the side at C-D/2; insets shorten a side at both ends so ring corners stay clear.
type angkorWatSeg struct {
	Axis               Axis
	A0, A1, C, D       float64
	Y0, H              float64
	Neg, Pos           angkorWatKind
	NegInset, PosInset float64
	Ends               bool // Closed pediment ends.
	Rise               float64
}

func (s angkorWatSeg) local(p Vec3) (float64, float64) {
	if s.Axis == AxisX {
		return p.X, p.Z - s.C
	}
	return p.Z, p.X - s.C
}

func (s angkorWatSeg) box(across, y0, y1 float64) Box {
	if s.Axis == AxisX {
		return Box{V(s.A0-1, y0, s.C-across), V(s.A1+1, y1, s.C+across)}
	}
	return Box{V(s.C-across, y0, s.A0-1), V(s.C+across, y1, s.A1+1)}
}

func (AngkorWat) mod(v, m float64) float64 { return math.Mod(math.Mod(v, m)+m, m) }

// segment builds one gallery: side walls, colonnades or windowed walls, and the vault.
func (a AngkorWat) segment(g *Grid, s angkorWatSeg) {
	w := s.D / 2
	eave := s.Y0 + s.H
	ow := w + 0.5
	rise := s.Rise
	if rise == 0 {
		rise = 0.5*s.D + 0.6
	}
	box := s.box(ow+0.5, s.Y0-0.5, eave+rise+0.5)
	side := func(kind angkorWatKind, sign, inset float64) {
		if kind == angkorWatOpen {
			return
		}
		g.Use(Sand).Fill(Func{box, func(p Vec3) bool {
			al, ac := s.local(p)
			if al < s.A0+inset || al > s.A1-inset || p.Y < s.Y0 || p.Y > eave {
				return false
			}
			x := sign * ac
			if x < w-1.0 || x > w {
				return false
			}
			switch kind {
			case angkorWatColonnade:
				return p.Y > eave-0.8 || a.mod(al-s.A0-inset, 2.6) < 1.0
			case angkorWatWindows:
				u := a.mod(al-s.A0-inset, 3.2)
				return !(u > 1.1 && u < 2.7 && p.Y > s.Y0+1.0 && p.Y < eave-0.9)
			}
			return true
		}})
	}
	relief := func(kind angkorWatKind, sign, inset float64) {
		if kind != angkorWatWall {
			return
		}
		// Bas-relief face inside the gallery, darkened by centuries of hands.
		g.Paint(DarkStone, Func{box, func(p Vec3) bool {
			al, ac := s.local(p)
			x := sign * ac
			return al >= s.A0+inset && al <= s.A1-inset && x >= w-1.0 && x <= w-0.6 && p.Y > s.Y0+0.3 && p.Y < eave-0.6
		}})
	}
	defer relief(s.Neg, -1, s.NegInset)
	defer relief(s.Pos, 1, s.PosInset)
	side(s.Neg, -1, s.NegInset)
	side(s.Pos, 1, s.PosInset)
	g.Use(DarkStone).FillVolume(Func{box, func(p Vec3) bool {
		al, ac := s.local(p)
		if al < s.A0-0.3 || al > s.A1+0.3 {
			return false
		}
		t := math.Abs(ac) / ow
		if t > 1 {
			return false
		}
		ho := rise * (1 - math.Pow(t, 1.7))
		y := p.Y - eave
		if y > ho || y < -0.7 {
			return false
		}
		end := s.Ends && (al < s.A0+0.6 || al > s.A1-0.6)
		return end || y > ho-1.0 || (y < 0 && math.Abs(ac) > w-1.0)
	}})
	// Corbel courses: pale stone bands every few rows.
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		al, ac := s.local(p)
		if al < s.A0-0.3 || al > s.A1+0.3 || math.Abs(ac) > ow {
			return false
		}
		y := p.Y - eave
		return y > 0.2 && a.mod(y, 1.6) < 0.4
	}})
}

// ring builds a rectangular galleried enclosure from four segments.
func (a AngkorWat) ring(g *Grid, x0, x1, z0, z1, d, y0, h float64, outer, inner angkorWatKind) {
	a.segment(g, angkorWatSeg{Axis: AxisX, A0: x0, A1: x1, C: z0 + d/2, D: d, Y0: y0, H: h, Neg: outer, Pos: inner, PosInset: d})
	a.segment(g, angkorWatSeg{Axis: AxisX, A0: x0, A1: x1, C: z1 - d/2, D: d, Y0: y0, H: h, Neg: inner, Pos: outer, NegInset: d})
	a.segment(g, angkorWatSeg{Axis: AxisZ, A0: z0, A1: z1, C: x0 + d/2, D: d, Y0: y0, H: h, Neg: outer, Pos: inner, PosInset: d})
	a.segment(g, angkorWatSeg{Axis: AxisZ, A0: z0, A1: z1, C: x1 - d/2, D: d, Y0: y0, H: h, Neg: inner, Pos: outer, NegInset: d})
}

// pavilion is a cruciform gopura or corner pavilion: arms reach arm from the center and
// are 2*half wide; doors pass on both axes. tower > 0 adds a (possibly truncated) prasat.
func (a AngkorWat) pavilion(g *Grid, c Vec2, arm, half, y0, h, tower, cut float64) {
	cross := func(dx, dz, in float64) bool {
		ax, az := math.Abs(dx), math.Abs(dz)
		return ax <= arm-in && az <= half-in || az <= arm-in && ax <= half-in
	}
	box := Box{V(c.X-arm-1, y0, c.Y-arm-1), V(c.X+arm+1, y0+h+half+3, c.Y+arm+1)}
	g.Carve(Func{box, func(p Vec3) bool { return p.Y > y0+0.05 && cross(p.X-c.X, p.Z-c.Y, 0) }})
	g.Use(Sand).Fill(Func{box, func(p Vec3) bool {
		return p.Y <= y0+h && cross(p.X-c.X, p.Z-c.Y, 0) && !cross(p.X-c.X, p.Z-c.Y, 1.0)
	}})
	rise := half + 1.0
	a.segment(g, angkorWatSeg{Axis: AxisX, A0: c.X - arm, A1: c.X + arm, C: c.Y, D: 2 * half, Y0: y0, H: h, Ends: true, Rise: rise})
	a.segment(g, angkorWatSeg{Axis: AxisZ, A0: c.Y - arm, A1: c.Y + arm, C: c.X, D: 2 * half, Y0: y0, H: h, Ends: true, Rise: rise})
	door := math.Min(1.1, half-1.0)
	g.Carve(
		Arch{Sill: c.At(y0 + 0.05), HalfWidth: door, Spring: h - door - 0.6, Depth: 2*arm + 2, Through: AxisX},
		Arch{Sill: c.At(y0 + 0.05), HalfWidth: door, Spring: h - door - 0.6, Depth: 2*arm + 2, Through: AxisZ},
	)
	if tower > 0 {
		a.prasat(g, c, y0+h+rise*0.45, half*0.85, 1.2, tower, 4, cut, false)
	}
}

// redent tests a plan offset against a square of half-size r with doubly redented corners.
func (AngkorWat) redent(dx, dz, r float64) bool {
	ax, az := math.Abs(dx), math.Abs(dz)
	return ax <= r && az <= 0.64*r || az <= r && ax <= 0.64*r || ax <= 0.84*r && az <= 0.84*r
}

// prasat is a Khmer tower: a redented cella with porches, then the tiered lotus-bud
// (corncob) spire. cut < 1 truncates the spire as a ruin.
func (a AngkorWat) prasat(g *Grid, c Vec2, y0, r, body, bud float64, tiers int, cut float64, porch bool) {
	yb := y0 + body
	top := yb + bud
	box := Box{V(c.X-r-2, y0, c.Y-r-2), V(c.X+r+2, top+1.5, c.Y+r+2)}
	shell := func(dx, dz, rr float64) bool {
		if !a.redent(dx, dz, rr) {
			return false
		}
		return rr < 2.0 || !a.redent(dx, dz, rr-1.2)
	}
	g.Use(Sand).Fill(Func{box, func(p Vec3) bool {
		return p.Y >= y0 && p.Y <= yb && shell(p.X-c.X, p.Z-c.Y, r)
	}})
	if porch {
		pw := math.Max(1.3, 0.42*r)
		ph := body * 0.85
		g.Fill(Func{box, func(p Vec3) bool {
			dx, dz := math.Abs(p.X-c.X), math.Abs(p.Z-c.Y)
			if p.Y < y0 || p.Y > y0+ph+pw {
				return false
			}
			arm := func(along, across float64) bool {
				if along > r+1.3 || across > pw {
					return false
				}
				return p.Y <= y0+ph || across <= pw-(p.Y-y0-ph) // Pediment.
			}
			return arm(dx, dz) || arm(dz, dx)
		}})
		door := math.Max(0.8, 0.22*r)
		g.Carve(
			Arch{Sill: c.At(y0 + 0.05), HalfWidth: door, Spring: ph - door - 0.8, Depth: 2*r + 3, Through: AxisX},
			Arch{Sill: c.At(y0 + 0.05), HalfWidth: door, Spring: ph - door - 0.8, Depth: 2*r + 3, Through: AxisZ},
		)
	}
	g.Use(DarkStone).Fill(Func{box, func(p Vec3) bool {
		return p.Y >= yb-0.7 && p.Y <= yb && shell(p.X-c.X, p.Z-c.Y, r+0.4)
	}})
	env := func(t float64) float64 { return math.Pow(1-math.Pow(t, 2.2), 0.6) }
	n := float64(tiers)
	radius := func(y float64) (float64, float64, bool) {
		t := (y - yb) / bud
		if t < 0 || t > 1 || (cut < 1 && t > cut) {
			return 0, 0, false
		}
		k := math.Floor(t * n)
		f := t*n - k
		return r * env(k/n) * (1 - 0.17*f), f, true
	}
	g.Use(Sand).FillVolume(Func{box, func(p Vec3) bool {
		rr, _, ok := radius(p.Y)
		return ok && shell(p.X-c.X, p.Z-c.Y, rr)
	}})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		_, f, ok := radius(p.Y)
		return ok && f < 0.16
	}})
	if cut >= 1 {
		g.Use(Stone).Fill(Capsule{c.At(yb + bud*0.8), c.At(top + 0.8), 0.6})
	}
}

// stair descends from top (at height y1) outward along out over run to height y0.
func (a AngkorWat) stair(g *Grid, top, out Vec2, run, y0, y1, w float64) {
	side := P(-out.Y, out.X)
	foot := top.Add(out.Scale(run))
	reach := w/2 + 1
	box := BoxOf(top.At(y0-0.5), foot.At(y1+1.2)).Grow(reach)
	box.Min.Y, box.Max.Y = y0-0.5, y1+1.2
	h := func(p Vec3) (float64, float64, float64) {
		d := p.Plan().Sub(top)
		s, l := d.Dot(out), d.Dot(side)
		hs := y1 - (y1-y0)*s/run
		return s, math.Abs(l), hs
	}
	g.Use(Sand).Fill(Func{box, func(p Vec3) bool {
		s, l, hs := h(p)
		hq := y0 + 0.8*math.Ceil((hs-y0)/0.8)
		return s >= -0.6 && s <= run && l <= w/2 && p.Y <= math.Min(hq, y1) && p.Y >= y0-0.5
	}})
	g.Use(Stone).Fill(Func{box, func(p Vec3) bool {
		s, l, hs := h(p)
		return s >= -0.6 && s <= run+0.4 && l > w/2 && l <= w/2+0.9 && p.Y <= math.Min(hs+1.0, y1+1.0) && p.Y >= y0-0.5
	}})
}

// naga is a serpent balustrade: posts and a rounded body; heads rears a fanned hood.
func (a AngkorWat) naga(g *Grid, from, to Vec2, y float64, heads bool) {
	g.Use(Stone).Fill(Capsule{from.At(y + 0.9), to.At(y + 0.9), 0.55})
	dir := to.Sub(from)
	n := int(dir.Length() / 1.8)
	for i := 0; i <= n; i++ {
		q := from.Add(dir.Scale(float64(i) / math.Max(float64(n), 1)))
		g.Fill(BlockAt(q, 0.8, 0.8, y, y+0.9))
	}
	if heads {
		u := dir.Scale(1 / dir.Length())
		side := P(-u.Y, u.X)
		c := from.At(y + 0.9)
		box := Box{c.Sub(V(2.2, 0.5, 2.2)), c.Add(V(2.2, 2.4, 2.2))}
		g.Use(DarkStone).Fill(Func{box, func(p Vec3) bool {
			d := p.Sub(c)
			along := d.X*u.X + d.Z*u.Y
			across := d.X*side.X + d.Z*side.Y
			return math.Abs(along) <= 0.5 && d.Y >= -0.3 && across*across+d.Y*d.Y*0.9 <= 2.0*2.0
		}})
		g.Fill(BlockAt(from, 1.2, 1.2, y, y+1.2))
	}
}

func (a AngkorWat) Build(g *Grid) Order {
	a.site(g)
	a.temple(g)
	a.approach(g)
	a.outerWall(g)
	a.trees(g)
	a.weather(g)
	return Outward(P(0, 0))
}

// plot is the site rectangle with rounded corners, inset by d.
func (AngkorWat) plot(x, z, d float64) bool {
	const r = 5.0
	hx, z0, z1 := angkorWatWallX+angkorWatMoat+4.5-d, angkorWatWallE-angkorWatMoat-4.5+d, angkorWatWallW+angkorWatMoat+4.5-d
	if math.Abs(x) > hx || z < z0 || z > z1 {
		return false
	}
	ex := math.Max(0, math.Abs(x)-(hx-r))
	ez := math.Max(0, math.Max(z0+r-z, z-(z1-r)))
	return ex*ex+ez*ez <= r*r
}

// moat reports whether a plan point lies in the moat, inset from its banks by d.
func (AngkorWat) moat(x, z, d float64) bool {
	in := math.Abs(x) <= angkorWatWallX+2-d && z >= angkorWatWallE-2+d && z <= angkorWatWallW+2-d
	out := math.Abs(x) <= angkorWatWallX+2+angkorWatMoat+d && z >= angkorWatWallE-2-angkorWatMoat-d && z <= angkorWatWallW+2+angkorWatMoat+d
	return out && !in
}

func (a AngkorWat) site(g *Grid) {
	box := Box{V(-60, -1, -60), V(60, 1, 80)}
	under := func(x, z float64) bool { return math.Abs(x) < 28 && math.Abs(z) < 25 }
	g.Use(Grass).Fill(Func{box, func(p Vec3) bool {
		return p.Y >= 0 && p.Y <= angkorWatGround && a.plot(p.X, p.Z, 0) && !a.moat(p.X, p.Z, 0) && !under(p.X, p.Z)
	}})
	g.Use(Water).Fill(Func{box, func(p Vec3) bool {
		return p.Y >= -0.45 && p.Y <= 0.05 && a.moat(p.X, p.Z, 0)
	}})
	// Laterite steps down into the moat on both banks.
	g.Use(DarkStone).Fill(Func{box, func(p Vec3) bool {
		return p.Y >= -0.45 && p.Y <= 0.45 && a.moat(p.X, p.Z, 0) && !a.moat(p.X, p.Z, -1.2) && !(p.Y > 0.05 && a.moat(p.X, p.Z, -0.6))
	}})
	// Sand paths around the third enclosure and from the gates.
	g.Paint(Sand, Func{Box{V(-60, angkorWatGround, -60), V(60, angkorWatGround, 80)}, func(p Vec3) bool {
		ring := math.Abs(p.X) < 32.5 && p.Z > -29.5 && p.Z < 29.5 && !(math.Abs(p.X) < 30 && p.Z > -27 && p.Z < 27)
		east := math.Abs(p.X) < 1.6 && p.Z < -26
		return ring || east
	}})
}

// temple builds the three enclosures, stairs and the five towers.
func (a AngkorWat) temple(g *Grid) {
	// Third-enclosure terrace: hollow plinth with moulded base and cornice.
	t3 := Box{V(-29, 0, -26), V(29, angkorWatL1, 26)}
	inRect := func(p Vec3, hx, hz, d float64) bool { return math.Abs(p.X) <= hx-d && math.Abs(p.Z) <= hz-d }
	g.Use(Sand).Fill(Func{t3.Grow(0.5), func(p Vec3) bool {
		if p.Y < 0 || p.Y > angkorWatL1 {
			return false
		}
		if p.Y >= angkorWatL1-0.8 {
			return inRect(p, 29, 26, 0) || p.Y >= angkorWatL1-0.4 && inRect(p, 29.4, 26.4, 0)
		}
		return inRect(p, 29, 26, 0) && !inRect(p, 29, 26, 1.2) || p.Y < 0.9 && inRect(p, 29.4, 26.4, 0) && !inRect(p, 29, 26, 1.2)
	}})
	g.Paint(DarkStone, Func{t3.Grow(0.5), func(p Vec3) bool { return p.Y < 0.9 || p.Y > angkorWatL1-0.4 && !inRect(p, 29, 26, 0) }})
	// Courtyard lawns and axis paths on the terrace.
	g.Paint(Grass, Func{Box{V(-25, angkorWatL1, -22), V(25, angkorWatL1, 22)}, func(p Vec3) bool {
		return math.Abs(p.X) > 1.6 && math.Abs(p.Z) > 1.6
	}})
	a.ring(g, -29, 29, -26, 26, 4, angkorWatL1, 3.6, angkorWatColonnade, angkorWatWall)

	// Cruciform cloister (Preah Poan) between the west gopura and the second level.
	cl := angkorWatSeg{Axis: AxisX, C: 19, D: 2.6, Y0: angkorWatL1, H: 3.0, Neg: angkorWatColonnade, Pos: angkorWatColonnade, Ends: true}
	cl.A0, cl.A1 = -16, -2.2
	a.segment(g, cl)
	cl.A0, cl.A1 = 2.2, 16
	a.segment(g, cl)
	for _, x := range []float64{-9, 9} {
		a.segment(g, angkorWatSeg{Axis: AxisZ, A0: 16, A1: 22, C: x, D: 2.6, Y0: angkorWatL1, H: 3.0, Neg: angkorWatColonnade, Pos: angkorWatColonnade})
	}
	// Basins in the four cloister courts.
	g.Paint(Water, Func{Box{V(-16, angkorWatL1, 16.5), V(16, angkorWatL1, 21.5)}, func(p Vec3) bool {
		ax := math.Abs(p.X)
		return (ax > 3.2 && ax < 7.6 || ax > 10.4 && ax < 15.5) && math.Abs(p.Z-19) > 1.5
	}})

	// Second level: hollow plinth, gallery, stairs.
	g.Use(Sand).Fill(Func{Box{V(-18.4, angkorWatL1-1, -16.4), V(18.4, angkorWatL2, 16.4)}, func(p Vec3) bool {
		if p.Y >= angkorWatL2-0.8 {
			return inRect(p, 18, 16, 0)
		}
		if p.Y < angkorWatL1+0.7 && inRect(p, 18.4, 16.4, 0) && !inRect(p, 18, 16, 1.3) {
			return true
		}
		return inRect(p, 18, 16, 0) && !inRect(p, 18, 16, 1.3)
	}})
	g.Paint(DarkStone, Func{Box{V(-18.4, angkorWatL1, -16.4), V(18.4, angkorWatL2, 16.4)}, func(p Vec3) bool {
		return p.Y < angkorWatL1+0.7 || math.Abs(p.Y-(angkorWatL2-1.2)) < 0.3
	}})
	a.ring(g, -18, 18, -16, 16, 3, angkorWatL2, 3.4, angkorWatWindows, angkorWatColonnade)

	// Bakan: steep stepped base of the upper level.
	half := func(y float64) float64 {
		h := 9 - 0.1*(y-angkorWatL2)
		if y > angkorWatL2+3.4 {
			h -= 0.35
		}
		if y > angkorWatL2+6.6 {
			h -= 0.35
		}
		return h
	}
	bk := Box{V(-9.5, angkorWatL2-1, -9.5), V(9.5, angkorWatL3, 9.5)}
	g.Use(Sand).Fill(Func{bk, func(p Vec3) bool {
		if p.Y < angkorWatL2-0.8 || p.Y > angkorWatL3 {
			return false
		}
		h := half(p.Y)
		if p.Y >= angkorWatL3-0.9 {
			return inRect(p, h, h, 0)
		}
		return inRect(p, h, h, 0) && !inRect(p, h, h, 1.3)
	}})
	g.Paint(DarkStone, Func{bk, func(p Vec3) bool {
		y := p.Y - angkorWatL2
		return y < 0.6 || math.Abs(y-3.6) < 0.35 || math.Abs(y-6.8) < 0.35 || p.Y > angkorWatL3-0.5
	}})

	// Stairs: third terrace (N, S, E), second level (N, S, E; west via cloister), Bakan.
	for _, d := range []Vec2{P(1, 0), P(-1, 0), P(0, -1)} {
		ex := 29.0
		if d.Y != 0 {
			ex = 26
		}
		a.stair(g, d.Scale(ex), d, 2.6, angkorWatGround, angkorWatL1, 3)
		ex2 := 18.0
		if d.Y != 0 {
			ex2 = 16
		}
		a.stair(g, d.Scale(ex2), d, 3.4, angkorWatL1, angkorWatL2, 2.8)
	}
	a.stair(g, P(0, 16), P(0, 1), 5, angkorWatL1, angkorWatL2, 3.2)
	for _, d := range []Vec2{P(1, 0), P(-1, 0), P(0, -1), P(0, 1)} {
		a.stair(g, d.Scale(8), d, 4.6, angkorWatL2, angkorWatL3, 2.4)
	}

	// Gopuras and corner pavilions of the third gallery.
	a.pavilion(g, P(0, 24), 4.6, 3.0, angkorWatL1, 5.0, 6, 0.55)
	a.pavilion(g, P(-9, 24), 3.0, 2.4, angkorWatL1, 4.4, 4, 0.45)
	a.pavilion(g, P(9, 24), 3.0, 2.4, angkorWatL1, 4.4, 4, 0.45)
	a.pavilion(g, P(0, -24), 3.6, 2.6, angkorWatL1, 4.6, 0, 0)
	a.pavilion(g, P(27, 0), 3.6, 2.6, angkorWatL1, 4.6, 0, 0)
	a.pavilion(g, P(-27, 0), 3.6, 2.6, angkorWatL1, 4.6, 0, 0)
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			a.pavilion(g, P(sx*27, sz*24), 3.2, 2.4, angkorWatL1, 4.3, 0, 0)
			a.pavilion(g, P(sx*16.5, sz*14.5), 2.6, 1.9, angkorWatL2, 3.8, 0, 0)
		}
	}
	for _, c := range []Vec2{P(0, 14.5), P(0, -14.5), P(16.5, 0), P(-16.5, 0)} {
		a.pavilion(g, c, 3.0, 2.2, angkorWatL2, 4.2, 0, 0)
	}

	// Upper level: gallery, cruciform galleries, gopuras, corner towers, central tower.
	a.ring(g, -8, 8, -8, 8, 2.4, angkorWatL3, 2.8, angkorWatWindows, angkorWatColonnade)
	a.segment(g, angkorWatSeg{Axis: AxisX, A0: -6, A1: 6, C: 0, D: 2.4, Y0: angkorWatL3, H: 2.8, Neg: angkorWatColonnade, Pos: angkorWatColonnade})
	a.segment(g, angkorWatSeg{Axis: AxisZ, A0: -6, A1: 6, C: 0, D: 2.4, Y0: angkorWatL3, H: 2.8, Neg: angkorWatColonnade, Pos: angkorWatColonnade})
	for _, c := range []Vec2{P(0, 6.8), P(0, -6.8), P(6.8, 0), P(-6.8, 0)} {
		a.pavilion(g, c, 2.4, 1.7, angkorWatL3, 3.4, 0, 0)
	}
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			c := P(sx*6.6, sz*6.6)
			g.Carve(BlockAt(c, 6, 6, angkorWatL3+0.05, angkorWatL3+8))
			g.Use(Sand).Fill(BlockAt(c, 5.4, 5.4, angkorWatL3-0.5, angkorWatL3+0.8))
			a.prasat(g, c, angkorWatL3+0.8, 2.3, 3.4, 11, 4, 1, true)
		}
	}
	g.Carve(BlockAt(P(0, 0), 9, 9, angkorWatL3+0.05, angkorWatL3+8))
	g.Use(Sand).Fill(BlockAt(P(0, 0), 8.6, 8.6, angkorWatL3-0.5, angkorWatL3+0.8))
	g.Use(Sand).Fill(BlockAt(P(0, 0), 7.6, 7.6, angkorWatL3+0.8, angkorWatL3+1.6))
	g.Paint(DarkStone, BlockAt(P(0, 0), 8.6, 8.6, angkorWatL3+0.4, angkorWatL3+0.8))
	a.prasat(g, P(0, 0), angkorWatL3+1.6, 3.1, 5.2, 16, 5, 1, true)
}

// approach builds the Terrace of Honour, inner causeway, libraries, ponds, the west
// gopura and the outer naga causeway across the moat.
func (a AngkorWat) approach(g *Grid) {
	// Terrace of Honour: cruciform deck on colonnettes.
	cross := func(p Vec3, d float64) bool {
		ax := math.Abs(p.X)
		return ax <= 3.5-d && p.Z >= 26 && p.Z <= 35-d || ax <= 8-d && p.Z >= 29+d && p.Z <= 33-d
	}
	tb := Box{V(-9, 0, 26), V(9, angkorWatL1+2, 36)}
	g.Use(Sand).Fill(Func{tb, func(p Vec3) bool {
		if p.Y < angkorWatGround-0.1 || p.Y > angkorWatL1 || !cross(p, 0) {
			return false
		}
		if p.Y >= angkorWatL1-0.8 {
			return true
		}
		edge := !cross(p, 0.9)
		return edge && (a.mod(p.X, 1.8) < 0.9 || a.mod(p.Z, 1.8) < 0.9) && (a.mod(p.X+p.Z, 1.8) < 0.9)
	}})
	g.Paint(DarkStone, Func{tb, func(p Vec3) bool { return p.Y > angkorWatL1-0.4 && !cross(p, 0.6) }})
	railY := angkorWatL1
	a.naga(g, P(3, 35), P(3, 33), railY, true)
	a.naga(g, P(-3, 35), P(-3, 33), railY, true)
	for _, s := range []float64{-1, 1} {
		a.naga(g, P(s*3, 33), P(s*7.6, 33), railY, false)
		a.naga(g, P(s*7.6, 29.4), P(s*3, 29.4), railY, false)
		a.naga(g, P(s*7.6, 33), P(s*7.6, 32.6), railY, false)
		a.stair(g, P(s*8, 31), P(s, 0), 2.6, angkorWatGround, angkorWatL1, 2.2)
	}
	a.stair(g, P(0, 35), P(0, 1), 2.6, angkorWatGround, angkorWatL1, 4.4)

	// Inner causeway, raised with naga balustrades, from the west gopura to the terrace.
	cy := 1.4
	ce := angkorWatWallW - 2
	g.Use(DarkStone).Fill(Block(V(-2.5, angkorWatGround-0.1, 37.6), V(2.5, 0.9, ce)))
	g.Use(Stone).Fill(Block(V(-2.2, 0.9, 37.6), V(2.2, cy, ce)))
	for _, s := range []float64{-1, 1} {
		a.naga(g, P(s*2.1, 38.2), P(s*2.1, 41.8), cy, true)
		a.naga(g, P(s*2.1, 44.2), P(s*2.1, 46.8), cy, true)
		a.naga(g, P(s*2.1, 49.2), P(s*2.1, ce-0.5), cy, true)
		// Side landings with steps.
		for _, z := range []float64{43, 48} {
			g.Use(Stone).Fill(Block(V(s*2.2, angkorWatGround-0.1, z-1), V(s*4.2, cy, z+1)))
		}
	}

	// Libraries north and south of the causeway, joined by raised walks.
	for _, s := range []float64{-1, 1} {
		c := P(s*12, 48)
		g.Use(Stone).Fill(Block(V(s*2.2, angkorWatGround-0.1, 47.2), V(s*7, cy-0.4, 48.8)))
		a.library(g, c)
	}

	// Reflecting ponds with laterite edges.
	for _, s := range []float64{-1, 1} {
		x0, x1 := s*5, s*15
		pond := Block(V(math.Min(x0, x1), -0.5, 37.2), V(math.Max(x0, x1), angkorWatGround+0.1, 42.2))
		inner := Block(V(math.Min(x0, x1)+0.9, -0.5, 38.1), V(math.Max(x0, x1)-0.9, angkorWatGround+0.1, 41.3))
		g.Carve(pond)
		g.Use(DarkStone).Fill(Subtract(Block(pond.Min, V(pond.Max.X, 0.45, pond.Max.Z)), inner))
		g.Use(Water).Fill(Block(inner.Min.Add(V(0, 0.05, 0)), V(inner.Max.X, 0.05, inner.Max.Z)))
		for i := 0; i < 7; i++ {
			q := P(s*(6.5+Hash01(i, 3, 1)*7), 38.6+Hash01(i, 4, 1)*2.4)
			g.Use(Foliage).Fill(BlockAt(q, 1.0, 1.0, 0, 0.45))
		}
	}

	// West gopura: long gallery with three truncated towers and elephant gates.
	gy := angkorWatGround + 0.5
	g.Use(DarkStone).Fill(Block(V(-20, angkorWatGround-0.1, angkorWatWallW-2.6), V(20, gy, angkorWatWallW+2.6)))
	a.segment(g, angkorWatSeg{Axis: AxisX, A0: -20, A1: 20, C: angkorWatWallW, D: 4, Y0: gy, H: 3.6, Neg: angkorWatColonnade, Pos: angkorWatWindows, Ends: true})
	a.pavilion(g, P(0, angkorWatWallW), 4.4, 2.8, gy, 5.0, 7, 0.5)
	a.pavilion(g, P(-10, angkorWatWallW), 3.2, 2.4, gy, 4.4, 5, 0.4)
	a.pavilion(g, P(10, angkorWatWallW), 3.2, 2.4, gy, 4.4, 5, 0.4)
	for _, s := range []float64{-1, 1} {
		a.pavilion(g, P(s*19, angkorWatWallW), 2.6, 2.2, gy, 4.2, 0, 0)
	}
	a.stair(g, P(0, angkorWatWallW+2.6), P(0, 1), 1.2, angkorWatGround, gy, 4.4)

	// Outer causeway across the moat with naga balustrades and rearing heads.
	z0, z1 := angkorWatWallW+2.6, angkorWatWallW+2+angkorWatMoat+2.5
	g.Use(DarkStone).Fill(Block(V(-2.8, -0.5, z0), V(2.8, 0.45, z1)))
	g.Use(Stone).Fill(Block(V(-2.4, 0.45, z0), V(2.4, 0.95, z1)))
	for _, s := range []float64{-1, 1} {
		a.naga(g, P(s*2.3, z1-0.5), P(s*2.3, z0+0.6), 0.95, true)
	}
}

// library is a cruciform hall on a moulded plinth with porches and vaulted roofs.
func (a AngkorWat) library(g *Grid, c Vec2) {
	base := angkorWatGround - 0.1
	py := 1.5
	g.Use(Sand).Fill(Union(
		Block(V(c.X-3.4, base, c.Y-5.0), V(c.X+3.4, py, c.Y+5.0)),
		Block(V(c.X-5.4, base, c.Y-1.8), V(c.X+5.4, py, c.Y+1.8)),
	))
	g.Paint(DarkStone, Block(V(c.X-5.4, base, c.Y-5.0), V(c.X+5.4, base+0.4, c.Y+5.0)))
	// Hall body and transept porches.
	g.Use(Sand).Fill(Func{Box{V(c.X-5, py, c.Y-4.8), V(c.X+5, py+3.2, c.Y+4.8)}, func(p Vec3) bool {
		dx, dz := math.Abs(p.X-c.X), math.Abs(p.Z-c.Y)
		hall := dx <= 2.4 && dz <= 4.2 && !(dx <= 1.4 && dz <= 3.2)
		porch := dx <= 4.6 && dz <= 1.6 && !(dx <= 3.6 && dz <= 0.6)
		return hall || porch
	}})
	a.segment(g, angkorWatSeg{Axis: AxisZ, A0: c.Y - 4.2, A1: c.Y + 4.2, C: c.X, D: 4.8, Y0: py, H: 3.0, Ends: true, Rise: 2.8})
	a.segment(g, angkorWatSeg{Axis: AxisX, A0: c.X - 4.6, A1: c.X + 4.6, C: c.Y, D: 3.2, Y0: py, H: 3.0, Ends: true, Rise: 2.0})
	g.Carve(
		Arch{Sill: c.At(py + 0.05), HalfWidth: 0.9, Spring: 1.6, Depth: 12, Through: AxisX},
		Arch{Sill: c.At(py + 0.05), HalfWidth: 0.9, Spring: 1.6, Depth: 12, Through: AxisZ},
	)
	for _, d := range []Vec2{P(0, 1), P(0, -1)} {
		a.stair(g, c.Add(d.Scale(5.0)), d, 1.6, angkorWatGround, py, 2.4)
	}
}

// outerWall builds the laterite enclosure wall with a stone coping and an east gate.
func (a AngkorWat) outerWall(g *Grid) {
	x, e, w := angkorWatWallX, angkorWatWallE, angkorWatWallW
	box := Box{V(-x-1, 0, e-1), V(x+1, 3.6, w+1)}
	on := func(p Vec3) bool {
		ax := math.Abs(p.X)
		side := ax >= x-0.6 && ax <= x+0.6 && p.Z >= e-0.6 && p.Z <= w+0.6
		ends := (math.Abs(p.Z-e) <= 0.6 || math.Abs(p.Z-w) <= 0.6) && ax <= x+0.6
		gap := math.Abs(p.Z-w) <= 0.6 && ax < 21 || math.Abs(p.Z-e) <= 0.6 && ax < 2.2
		return (side || ends) && !gap
	}
	g.Use(Brick).Fill(Func{box, func(p Vec3) bool { return p.Y >= 0 && p.Y <= 2.1 && on(p) }})
	g.Use(DarkStone).Fill(Func{box, func(p Vec3) bool {
		return p.Y > 2.1 && p.Y <= 2.6 && on(p)
	}})
	a.pavilion(g, P(0, e), 2.6, 2.0, angkorWatGround, 3.8, 0, 0)
}

// trees scatters sugar palms and jungle trees over the grounds, avoiding structures.
func (a AngkorWat) trees(g *Grid) {
	clear := func(x, z float64) bool {
		if math.Abs(x) < 33 && z > -30 && z < 30 { // Third enclosure and its path.
			return false
		}
		if math.Abs(x) < 17 && z > 25 && z < 61 || math.Abs(x) < 19.5 && z > 41.5 && z < 54.5 { // Approach, ponds, libraries.
			return false
		}
		if math.Abs(x) < 23 && z > angkorWatWallW-4.5 { // West gopura.
			return false
		}
		if math.Abs(x) < 4 && z < -26 {
			return false
		}
		return true
	}
	inside := func(x, z float64) bool {
		return math.Abs(x) < angkorWatWallX-1.6 && z > angkorWatWallE+1.6 && z < angkorWatWallW-1.6
	}
	for gx := -9; gx <= 9; gx++ {
		for gz := -9; gz <= 14; gz++ {
			i := (gx+20)*100 + gz + 20
			if Hash01(i, 10) > 0.85 {
				continue
			}
			x := (float64(gx) + (Hash01(i, 11)-0.5)*0.6) * 4.4
			z := (float64(gz) + (Hash01(i, 12)-0.5)*0.6) * 4.4
			if !inside(x, z) || !clear(x, z) {
				continue
			}
			at := P(x, z)
			if Hash01(i, 13) < 0.45 {
				// Sugar palm.
				h := 6 + Hash01(i, 14)*3
				g.Use(Wood).Fill(Cylinder{Center: at, Radius: 0.5, Y0: angkorWatGround, Y1: angkorWatGround + h})
				g.Use(Foliage).FillVolume(Ellipsoid{at.At(angkorWatGround + h + 0.3), V(1.9, 1.1, 1.9)})
				continue
			}
			h := 6 + Hash01(i, 15)*4
			g.Use(Wood).Fill(Cylinder{Center: at, Radius: 0.6, Y0: angkorWatGround, Y1: angkorWatGround + h*0.6})
			col := Foliage
			if Hash01(i, 16) < 0.3 {
				col = Grass
			}
			g.Use(col).FillVolume(
				Ellipsoid{at.At(angkorWatGround + h*0.75), V(h*0.33, h*0.22, h*0.33)},
				Ellipsoid{at.Add(P(h*0.18, h*0.1)).At(angkorWatGround + h*0.62), V(h*0.24, h*0.17, h*0.24)},
			)
		}
	}
	// Palms along the outer banks by the west causeway.
	for _, s := range []float64{-1, 1} {
		for k := 0; k < 3; k++ {
			at := P(s*(24+float64(k)*6.5), angkorWatWallW+2+angkorWatMoat+2.4)
			h := 5 + Hash01(k, int(s)+5, 7)*2
			g.Use(Wood).Fill(Cylinder{Center: at, Radius: 0.5, Y0: angkorWatGround, Y1: angkorWatGround + h})
			g.Use(Foliage).FillVolume(Ellipsoid{at.At(angkorWatGround + h + 0.3), V(1.8, 1.0, 1.8)})
		}
	}
}

// weather streaks the sandstone with darker patina and moss.
func (a AngkorWat) weather(g *Grid) {
	for _, c := range g.Cells() {
		col, _ := g.ColorOf(c)
		p := g.World(c)
		if p.Y < angkorWatGround+0.2 || col != Sand && col != Stone && col != DarkStone {
			continue
		}
		bx, by, bz := int(math.Floor(p.X/2.2)), int(math.Floor(p.Y/1.6)), int(math.Floor(p.Z/2.2))
		patch := Hash01(bx, by, bz, 41)
		fine := Hash01(int(c.X), int(c.Y), int(c.Z), 42)
		moss := 0.05 * math.Max(0, 1-p.Y/20)
		switch {
		case patch < moss && fine < 0.75:
			g.Remove(c)
			g.Use(Foliage).Add(c)
		case col == Sand && patch > 0.86 && fine < 0.55:
			g.Remove(c)
			g.Use(Stone).Add(c)
		case col == Stone && fine < 0.07:
			g.Remove(c)
			g.Use(DarkStone).Add(c)
		}
	}
}
