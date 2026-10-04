package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// Kremlin is the Moscow Kremlin seen together with Red Square. A triangle of red brick
// walls with swallowtail merlons runs along the Moskva (south, +Z), Red Square (east,
// +X) and the Alexander Garden (north-west). Spasskaya Tower rises over its gate with
// clock faces, a green tent and a red star; Nikolskaya, Senate, Troitskaya, Borovitskaya
// and the round corner towers carry hipped tents. Inside stand the white cathedrals with
// gold onion domes, the Ivan the Great Bell Tower, the Grand Kremlin Palace and the
// Senate. Outside are the cobbled square, the Lenin Mausoleum with its spruces, a small
// St. Basil's, GUM, the Historical Museum and the river embankment.
type Kremlin struct{}

func init() { Register(Kremlin{}) }

func (Kremlin) Name() string  { return "kremlin" }
func (Kremlin) Title() string { return "Kremlin" }

// Plan shapes for kremlinPrism.
const (
	kremlinSquare = iota
	kremlinOctagon
	kremlinRound
)

// Wall corners: south-west (Vodovzvodnaya), south-east (Beklemishevskaya) and north
// (Corner Arsenal). The river is south (+Z); Red Square is east (+X).
var (
	kremlinSW = P(-40, 16)
	kremlinSE = P(20, 16)
	kremlinN  = P(20, -32)
)

const (
	kremlinWallHalf = 1.2
)

// norm is the plan distance from c in a frame turned by angle, for the given shape.
func (Kremlin) norm(c Vec2, angle float64, shape int, p Vec3) float64 {
	q := p.Plan().Sub(c).Rotate(-angle)
	x, z := math.Abs(q.X), math.Abs(q.Y)
	switch shape {
	case kremlinSquare:
		return math.Max(x, z)
	case kremlinOctagon:
		return math.Max(math.Max(x, z), (x+z)/math.Sqrt2)
	}
	return math.Hypot(x, z)
}

// prism is a vertical square, octagonal or round prism/frustum from apothem a0 at y0 to
// a1 at y1, turned by angle; thick > 0 keeps only the outer layer.
func (k Kremlin) prism(c Vec2, angle float64, shape int, a0, a1, y0, y1, thick float64) Solid {
	r := math.Max(a0, a1)*1.45 + 0.5
	return Func{Box{V(c.X-r, y0, c.Y-r), V(c.X+r, y1, c.Y+r)}, func(p Vec3) bool {
		a := a0 + (a1-a0)*(p.Y-y0)/math.Max(y1-y0, 1e-9)
		n := k.norm(c, angle, shape, p)
		return n <= a && (thick <= 0 || n > a-thick)
	}}
}

// local places a solid authored around the origin at c, turned by angle.
func (Kremlin) local(c Vec2, angle float64, s Solid) Solid {
	return RotateY(Translate(s, c.At(0)), c, angle)
}

// ribs selects narrow vertical strips at the octagon's corners (or n plan angles).
func (Kremlin) ribs(c Vec2, angle, r, y0, y1, n, phase, width float64) Solid {
	return Func{Box{V(c.X-r, y0, c.Y-r), V(c.X+r, y1, c.Y+r)}, func(p Vec3) bool {
		q := p.Plan().Sub(c).Rotate(-angle)
		x := math.Atan2(q.Y, q.X)/(2*math.Pi)*n + phase
		return math.Abs(x-math.Round(x)) < width
	}}
}

// swallow reports whether (u, y) lies in a swallowtail merlon: u is the offset from the
// merlon center along the wall, y the height above the wall top.
func (Kremlin) swallow(u, y float64) bool {
	const half, h, notch, nw = 1.25, 2.8, 1.3, 0.7
	au := math.Abs(u)
	if au > half || y < 0 {
		return false
	}
	top := h
	if au < nw {
		top = h - notch*(1-au/nw)
	}
	return y <= top
}

// onion is a bulb dome: a quarter ellipse from the neck to the widest point, then a
// shoulder that pinches into a concave spike.
func (Kremlin) onion(c Vec2, base, neck, R, H float64) Solid {
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

// star is a ruby five-pointed star: two crossed plates so it reads from every side.
func (Kremlin) star(c Vec3, R float64) Solid {
	ri := R * 0.42
	alpha := math.Pi / 5
	in := func(u, v float64) bool {
		r := math.Hypot(u, v)
		if r > R {
			return false
		}
		th := math.Atan2(u, v)
		phi := math.Abs(math.Mod(th+alpha+4*math.Pi, 2*alpha) - alpha)
		return r <= R*ri*math.Sin(alpha)/(ri*math.Sin(alpha-phi)+R*math.Sin(phi))
	}
	return Func{Box{c.Sub(V(R, R, R)), c.Add(V(R, R, R))}, func(p Vec3) bool {
		d := p.Sub(c)
		const t = 0.45
		return (math.Abs(d.Z) <= t && in(d.X, d.Y)) || (math.Abs(d.X) <= t && in(d.Z, d.Y))
	}}
}

// finial tops a tent: a gold spire with a red star or a gold vane.
func (k Kremlin) finial(g *Grid, c Vec2, y float64, star bool) {
	if star {
		g.Use(Gold).Fill(Capsule{c.At(y - 0.8), c.At(y + 1.6), 0.45}, Sphere(c.At(y+0.6), 0.7))
		g.Use(Red).Fill(k.star(c.At(y+3.2), 1.8))
		return
	}
	g.Use(Gold).Fill(
		Capsule{c.At(y - 0.8), c.At(y + 3), 0.45},
		Sphere(c.At(y+0.8), 0.65),
		Capsule{V(c.X, y+2.4, c.Y), V(c.X+1.3, y+2.4, c.Y), 0.4},
	)
}

// wall builds a curtain wall from a to b; out (+1/-1) picks the left or right side as
// the outside, where the swallowtail merlons stand.
func (k Kremlin) wall(g *Grid, a, b Vec2, out, top float64) {
	d := b.Sub(a)
	L := d.Length()
	u := d.Scale(1 / L)
	n := P(-u.Y, u.X).Scale(out)
	box := BoxOf(a.At(-1), b.At(top+3)).Grow(3)
	local := func(p Vec3) (float64, float64) {
		q := p.Plan().Sub(a)
		return q.Dot(u), q.Dot(n)
	}
	const half = kremlinWallHalf
	g.Use(Brick).Fill(Func{box, func(p Vec3) bool {
		s, w := local(p)
		if s < 0 || s > L || p.Y < 0 {
			return false
		}
		switch {
		case math.Abs(w) <= half && p.Y <= top:
			return true
		case w >= -half && w <= half+0.5 && p.Y <= 1.3: // Plinth.
			return true
		case w >= -half && w <= -half+0.7 && p.Y <= top+0.9: // Inner parapet.
			return true
		case w >= half-1.0 && w <= half && p.Y > top: // Merlons.
			const period = 3.8
			m := math.Mod(s, period) - period/2
			return k.swallow(m, p.Y-top)
		}
		return false
	}})
	// Arched niches on the inner face and a darker walkway.
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		s, w := local(p)
		return s >= 0 && s <= L && w > -half+0.6 && w < half-0.9 && p.Y > top-0.6 && p.Y <= top
	}})
	g.Carve(Func{box, func(p Vec3) bool {
		s, w := local(p)
		if s < 2 || s > L-2 || w > -half+0.6 || w < -half-0.5 {
			return false
		}
		m := math.Mod(s, 4.0) - 2
		y := p.Y - 1.3
		if math.Abs(m) > 1.2 || y < 0 {
			return false
		}
		rise := y - 3.0
		return rise <= 0 || m*m+rise*rise <= 1.44
	}})
}

// kremlinTower describes a wall tower: a square or round body with machicolations and
// merlons, an optional narrower stage, a hipped tent, optional lantern and finial.
type kremlinTower struct {
	c       Vec2
	angle   float64 // Turn of the local frame; the wall runs along local X.
	round   bool
	a       float64 // Body half-width or radius.
	h       float64 // Body top.
	stage   float64 // Height of the stage above the body (0: none).
	tent    float64 // Tent height.
	lantern bool
	star    bool
	gate    bool
	tentCol Color
}

func (k Kremlin) tower(g *Grid, t kremlinTower) {
	c, ang, a, h := t.c, t.angle, t.a, t.h
	shape := kremlinSquare
	if t.round {
		shape = kremlinRound
	}
	g.Carve(k.prism(c, ang, shape, a+0.4, a+0.4, 0.6, h+4, 0))
	g.Use(Brick).Fill(
		k.prism(c, ang, shape, a, a, 0, h, 1.3),
		k.prism(c, ang, shape, a+0.4, a+0.4, 0, 1.4, 1.0),
		k.prism(c, ang, shape, a, a, h-1.2, h, 0),
		// Machicolation corbel.
		k.prism(c, ang, shape, a+0.6, a+0.6, h-1.3, h, 1.1),
	)
	g.Use(White).Fill(k.prism(c, ang, shape, a+0.3, a+0.3, h-2.4, h-1.9, 0.6))
	g.Paint(Charcoal, k.prism(c, ang, shape, a-0.65, a-0.65, 0.5, h-1.3, 0.7))
	// Swallowtail merlons around the parapet.
	r := a + 0.6
	perim := 8 * r
	if t.round {
		perim = 2 * math.Pi * r
	}
	count := math.Max(4, math.Round(perim/3.8))
	period := perim / count
	g.Use(Brick).Fill(Func{Box{V(c.X-r*1.5, h, c.Y-r*1.5), V(c.X+r*1.5, h+3, c.Y+r*1.5)}, func(p Vec3) bool {
		n := k.norm(c, ang, shape, p)
		if n > r || n < r-1.0 {
			return false
		}
		q := p.Plan().Sub(c).Rotate(-ang)
		var s float64
		if t.round {
			s = (math.Atan2(q.Y, q.X) + math.Pi) * r
		} else if math.Abs(q.X) > math.Abs(q.Y) {
			s = q.Y
		} else {
			s = q.X
		}
		m := math.Mod(s+100*period, period) - period/2
		return k.swallow(m, p.Y-h)
	}})
	// Loopholes and arched windows in two rows.
	faces := 4
	if t.round {
		faces = 6
	}
	for i := 0; i < faces; i++ {
		an := ang + float64(i)*2*math.Pi/float64(faces) + math.Pi/float64(faces)*0.0
		for row, y := range []float64{h * 0.3, h * 0.62} {
			if (i+row)%2 == 1 && t.round {
				continue
			}
			g.Carve(RotateY(Arch{V(c.X+a-0.5, y, c.Y), 0.5, 1.1, 3, AxisX}, c, an))
		}
	}
	if t.gate {
		g.Paint(White, k.local(c, ang, Arch{V(0, 0.4, 0), 2.3, 3.0, 2*a + 1.2, AxisZ}))
		g.Carve(k.local(c, ang, Arch{V(0, 0.4, 0), 1.7, 2.8, 2*a + 8, AxisZ}))
	}
	// Stage: a smaller square on square towers, an octagon on round ones.
	y := h
	sa := a * 0.72
	sShape := kremlinSquare
	if t.round {
		sa, sShape = a*0.62, kremlinOctagon
	}
	if t.stage > 0 {
		y1 := h + t.stage
		g.Use(Brick).Fill(k.prism(c, ang, sShape, sa, sa, h-0.5, y1, 1.2), k.prism(c, ang, sShape, sa, sa, y1-1, y1, 0))
		g.Use(White).Fill(k.prism(c, ang, sShape, sa+0.35, sa+0.35, y1-0.7, y1, 0.9))
		g.Paint(Charcoal, k.prism(c, ang, sShape, sa-0.6, sa-0.6, h, y1-1, 0.6))
		n := 4
		if sShape == kremlinOctagon {
			n = 8
		}
		for i := 0; i < n; i += 1 + n/8 {
			an := ang + float64(i)*2*math.Pi/float64(n)
			g.Paint(White, RotateY(Arch{V(c.X+sa-0.5, h+0.6, c.Y), 0.9, t.stage - 2.6, 1.4, AxisX}, c, an))
			g.Carve(RotateY(Arch{V(c.X+sa-0.5, h+1.0, c.Y), 0.55, t.stage - 2.9, 3, AxisX}, c, an))
		}
		y = y1
	}
	// Tent, optionally with an open lantern and a second small tent.
	col := t.tentCol.Or(Foliage)
	base := sa + 0.4
	if t.stage == 0 {
		base = a + 0.2
	}
	if t.lantern {
		mid := y + t.tent*0.55
		la := 1.5
		g.Use(col).FillVolume(k.prism(c, ang+math.Pi/8*0, kremlinOctagon, base, la, y, mid, 1.2))
		g.Use(Brick).Fill(k.prism(c, ang, kremlinOctagon, la, la, mid-0.3, mid+2.8, 1.0), k.prism(c, ang, kremlinOctagon, la, la, mid+2.2, mid+2.8, 0))
		for i := 0; i < 8; i += 2 {
			an := ang + float64(i)*math.Pi/4
			g.Carve(RotateY(Arch{V(c.X+la-0.5, mid+0.3, c.Y), 0.45, 1.3, 2, AxisX}, c, an))
		}
		g.Use(White).Fill(k.prism(c, ang, kremlinOctagon, la+0.3, la+0.3, mid+2.2, mid+2.8, 0.8))
		top := y + t.tent
		g.Use(col).FillVolume(k.prism(c, ang, kremlinOctagon, la+0.3, 0.45, mid+2.8, top, 1.2))
		k.finial(g, c, top, t.star)
		return
	}
	top := y + t.tent
	g.Use(col).FillVolume(k.prism(c, ang, kremlinOctagon, base, 0.45, y, top, 1.2))
	k.finial(g, c, top, t.star)
}

// spasskaya builds the Saviour Tower over the main gate on the Red Square wall.
func (k Kremlin) spasskaya(g *Grid, c Vec2) {
	sq := func(a, y0, y1, thick float64) Solid { return k.prism(c, 0, kremlinSquare, a, a, y0, y1, thick) }
	g.Carve(sq(4.6, 0.6, 20, 0))
	// Gate body.
	g.Use(Brick).Fill(sq(4, 0, 14, 1.3), sq(4.4, 0, 1.4, 1.0), sq(4, 13, 14, 0))
	g.Use(White).Fill(sq(4.35, 12.9, 13.8, 0.8), sq(4.3, 6.4, 6.9, 0.6))
	g.Paint(Charcoal, sq(3.35, 0.5, 13, 0.7))
	g.Paint(White, Arch{V(c.X, 0.4, c.Y), 2.5, 3.4, 9.2, AxisX})
	g.Carve(Arch{V(c.X, 0.4, c.Y), 1.8, 3.2, 14, AxisX})
	for i := 0; i < 4; i++ {
		an := float64(i) * math.Pi / 2
		for _, off := range []float64{-2, 2} {
			g.Paint(White, RotateY(Arch{V(c.X+3.5, 8.6, c.Y+off), 0.85, 1.6, 1.4, AxisX}, c, an))
			g.Carve(RotateY(Arch{V(c.X+3.5, 9.0, c.Y+off), 0.5, 1.4, 3, AxisX}, c, an))
		}
	}
	// Second stage with a white arcade frieze and Gothic pinnacles.
	g.Use(Brick).Fill(sq(3.6, 13.5, 19.5, 1.2), sq(3.6, 18.5, 19.5, 0))
	g.Use(White).Fill(sq(3.95, 18.8, 19.5, 0.8))
	for i := 0; i < 4; i++ {
		an := float64(i) * math.Pi / 2
		for _, off := range []float64{-1.6, 0, 1.6} {
			g.Carve(RotateY(Arch{V(c.X+3.3, 15.4, c.Y+off), 0.45, 1.5, 3, AxisX}, c, an))
		}
	}
	for _, dx := range []float64{-1, 1} {
		for _, dz := range []float64{-1, 1} {
			at := P(c.X+dx*3.8, c.Y+dz*3.8)
			g.Use(White).Fill(Capsule{at.At(17.5), at.At(19.5), 0.6}, Frustum{at, 0.7, 0.2, 19.5, 21.5})
			g.Use(Gold).Fill(Sphere(at.At(21.8), 0.5))
			at = P(c.X+dx*3.6, c.Y+dz*3.6)
			g.Use(White).Fill(Frustum{at, 0.6, 0.2, 26, 28})
			g.Use(Gold).Fill(Sphere(at.At(28.3), 0.45))
		}
	}
	for i := 0; i < 4; i++ {
		at := Polar(c, 4.1, float64(i)*math.Pi/2)
		g.Use(White).Fill(Frustum{at, 0.6, 0.2, 19.3, 21.0})
	}
	// Clock stage: four clock faces with gold rims and hands.
	g.Use(Brick).Fill(sq(3.3, 19.5, 26, 1.2), sq(3.3, 25, 26, 0))
	g.Use(White).Fill(sq(3.65, 25.3, 26, 0.8))
	g.Paint(Charcoal, sq(2.7, 19.5, 25, 0.6))
	cy := 22.6
	for i := 0; i < 4; i++ {
		an := float64(i) * math.Pi / 2
		face := func(r0, r1, d0, d1 float64) Solid {
			return RotateY(Func{Box{V(c.X+3.3+d0, cy-r1, c.Y-r1), V(c.X+3.3+d1, cy+r1, c.Y+r1)}, func(p Vec3) bool {
				r := math.Hypot(p.Z-c.Y, p.Y-cy)
				return r >= r0 && r <= r1
			}}, c, an)
		}
		g.Use(Gold).Fill(face(2.15, 2.85, -0.4, 0.5))
		g.Use(Charcoal).Fill(face(0, 2.15, -0.4, 0.2))
		g.Use(White).Fill(face(0, 0.5, 0.2, 0.6))
		hand := func(len, th float64) Solid {
			e := P(math.Sin(th)*len, math.Cos(th)*len)
			return RotateY(Capsule{V(c.X+3.5, cy, c.Y), V(c.X+3.5, cy+e.Y, c.Y+e.X), 0.4}, c, an)
		}
		g.Use(Gold).Fill(hand(1.9, 0.5), hand(1.3, 2.4))
	}
	// Open octagonal belfry, white-columned.
	oct := func(a0, a1, y0, y1, thick float64) Solid { return k.prism(c, 0, kremlinOctagon, a0, a1, y0, y1, thick) }
	g.Use(Brick).Fill(oct(2.9, 2.9, 25.5, 31.5, 1.1), oct(2.9, 2.9, 30.8, 31.5, 0))
	for i := 0; i < 8; i++ {
		an := float64(i) * math.Pi / 4
		g.Carve(RotateY(Arch{V(c.X+2.4, 26.6, c.Y), 0.85, 2.3, 2.5, AxisX}, c, an))
		at := Polar(c, 2.9/math.Cos(math.Pi/8)-0.2, an+math.Pi/8)
		g.Use(White).Fill(Capsule{at.At(26), at.At(31), 0.45})
	}
	g.Use(Gold).Fill(Frustum{c, 1.3, 0.6, 26.4, 28.6})
	g.Use(White).Fill(oct(3.25, 3.25, 30.9, 31.6, 0.9))
	// Tall green tent, white ribs, gold spire and ruby star.
	tb, tt := 31.6, 44.0
	g.Use(Grass).FillVolume(oct(3.0, 0.5, tb, tt, 1.2))
	g.Paint(White, k.ribs(c, 0, 4, tb, tt-0.5, 8, 0.5, 0.06))
	g.Paint(Gold, Func{Box{V(c.X-4, tb, c.Y-4), V(c.X+4, tt, c.Y+4)}, func(p Vec3) bool {
		t := (p.Y - tb) / (tt - tb)
		return math.Abs(t-0.3) < 0.025 || math.Abs(t-0.62) < 0.025
	}})
	k.finial(g, c, tt, true)
}

// dome places a drum with slit windows, a gold onion and a cross at height y.
func (k Kremlin) dome(g *Grid, c Vec2, y, dr, dh, R, H float64, col Color) {
	g.Use(White).Fill(Cylinder{c, dr, math.Max(dr-1.1, 0), y - 0.6, y + dh})
	g.Use(White).Fill(Cylinder{c, dr - 0.9, 0, y + dh - 0.8, y + dh})
	g.Use(White).Fill(Cylinder{c, dr + 0.4, dr - 0.6, y + dh - 0.6, y + dh})
	if dr >= 1.8 {
		g.Paint(Charcoal, Cylinder{c, dr - 0.55, 0, y, y + dh - 0.8})
		for i := 0; i < 8; i += 2 {
			g.Carve(RotateY(Arch{V(c.X+dr-0.5, y+0.6, c.Y), 0.4, dh - 2.4, 2.4, AxisX}, c, float64(i)*math.Pi/4+math.Pi/8))
		}
	}
	g.Use(col).FillVolume(k.onion(c, y+dh, dr*0.8, R, H))
	top := y + dh + H
	g.Use(Gold).Fill(
		Capsule{c.At(top - 0.6), c.At(top + 2.4), 0.42},
		Capsule{V(c.X-0.9, top+1.6, c.Y), V(c.X+0.9, top+1.6, c.Y), 0.4},
	)
}

// kremlinDome is one drum and onion on a cathedral roof.
type kremlinDome struct {
	c            Vec2
	dr, dh, R, H float64
	col          Color
}

// cathedral is a white cross-domed church: zakomara gables, pilasters, apses on the
// east (+X), a gilded portal on the south (+Z) and domes on drums.
func (k Kremlin) cathedral(g *Grid, c Vec2, w, d, h float64, nx, nz, apses int, roof Color, domes []kremlinDome) {
	x0, x1, z0, z1 := c.X-w/2, c.X+w/2, c.Y-d/2, c.Y+d/2
	g.Use(Stone).Fill(BlockAt(c, w+1.0, d+1.0, 0, 1.0))
	bx, bz := w/float64(nx), d/float64(nz)
	// Apses on the east.
	for i := 0; i < apses; i++ {
		z := c.Y + (float64(i)-float64(apses-1)/2)*bz
		ar := math.Min(bz/2, 2.4)
		at := P(x1, z)
		g.Use(White).Fill(Cylinder{at, ar, ar - 1.2, 0, h * 0.72})
		g.Use(roof).Fill(Ellipsoid{at.At(h * 0.72), V(ar+0.2, ar*0.7, ar+0.2)})
		g.Carve(Arch{V(x1+ar-0.5, h*0.3, z), 0.45, 1.6, 2.4, AxisX})
	}
	g.Use(White).Fill(BlockAt(c, w, d, 0, h))
	g.Carve(BlockAt(c, w-2.4, d-2.4, 1.0, h-1.2))
	g.Paint(Charcoal, BlockAt(c, w-1.2, d-1.2, 1.0, h-1.2))
	// Zakomaras: semicircular gables over each bay, and the vaulted roof behind them.
	bump := func(x, z float64) float64 {
		ux := math.Mod(x-x0, bx) - bx/2
		uz := math.Mod(z-z0, bz) - bz/2
		vx := math.Sqrt(math.Max(0, bx*bx/4-ux*ux))
		vz := math.Sqrt(math.Max(0, bz*bz/4-uz*uz))
		return math.Max(vx, vz) * 0.8
	}
	top := h + math.Max(bx, bz)/2
	g.Use(roof).Fill(Func{Box{V(x0, h-1.2, z0), V(x1, top, z1)}, func(p Vec3) bool {
		return p.Y <= h+bump(p.X, p.Z)
	}})
	g.Use(White).Fill(Func{Box{V(x0-0.3, h-1.2, z0-0.3), V(x1+0.3, top, z1+0.3)}, func(p Vec3) bool {
		edgeZ := math.Abs(p.Z-c.Y) >= d/2-0.9
		edgeX := math.Abs(p.X-c.X) >= w/2-0.9
		if edgeZ {
			ux := math.Mod(p.X-x0, bx) - bx/2
			if (p.Y-h)*(p.Y-h)+ux*ux <= bx*bx/4*0.85 || p.Y <= h {
				return true
			}
		}
		if edgeX {
			uz := math.Mod(p.Z-z0, bz) - bz/2
			if (p.Y-h)*(p.Y-h)+uz*uz <= bz*bz/4*0.85 || p.Y <= h {
				return true
			}
		}
		return false
	}})
	// Pilasters between bays and a cornice band.
	for i := 0; i <= nx; i++ {
		x := x0 + float64(i)*bx
		g.Use(White).Fill(Block(V(x-0.4, 1, z0-0.4), V(x+0.4, h, z0)), Block(V(x-0.4, 1, z1), V(x+0.4, h, z1+0.4)))
	}
	for i := 0; i <= nz; i++ {
		z := z0 + float64(i)*bz
		g.Use(White).Fill(Block(V(x0-0.4, 1, z-0.4), V(x0, h, z+0.4)), Block(V(x1, 1, z-0.4), V(x1+0.4, h, z+0.4)))
	}
	g.Use(White).Fill(Subtract(BlockAt(c, w+0.7, d+0.7, h*0.48, h*0.48+0.5), BlockAt(c, w-1, d-1, 0, h)))
	// Windows: two rows per bay on every face.
	for i := 0; i < nx; i++ {
		x := x0 + (float64(i)+0.5)*bx
		for _, y := range []float64{h * 0.22, h * 0.62} {
			g.Carve(Arch{V(x, y, z0+0.6), 0.5, 1.4, 2.6, AxisZ}, Arch{V(x, y, z1-0.6), 0.5, 1.4, 2.6, AxisZ})
		}
	}
	for i := 0; i < nz; i++ {
		z := z0 + (float64(i)+0.5)*bz
		for _, y := range []float64{h * 0.22, h * 0.62} {
			g.Carve(Arch{V(x0+0.6, y, z), 0.5, 1.4, 2.6, AxisX})
		}
	}
	// Gilded south portal.
	g.Use(Gold).Fill(Subtract(Arch{V(c.X, 1.0, z1+0.3), 1.9, 2.4, 0.8, AxisZ}, Arch{V(c.X, 1.0, z1+0.3), 1.3, 2.2, 2, AxisZ}))
	g.Carve(Arch{V(c.X, 1.0, z1), 1.25, 2.1, 2.6, AxisZ})
	for _, dm := range domes {
		k.dome(g, dm.c, h+bump(dm.c.X, dm.c.Y)*0.5, dm.dr, dm.dh, dm.R, dm.H, dm.col)
	}
}

// ivan builds the Ivan the Great Bell Tower: stacked white octagons with open belfries,
// bells, a drum with a gold band and a great gold onion.
func (k Kremlin) ivan(g *Grid, c Vec2) {
	oct := func(a0, a1, y0, y1, thick float64) Solid { return k.prism(c, 0, kremlinOctagon, a0, a1, y0, y1, thick) }
	g.Use(Stone).Fill(oct(5.0, 5.0, 0, 1.2, 0))
	tiers := []struct{ a, y0, y1 float64 }{{4.4, 0, 14}, {3.7, 14, 24}, {3.1, 24, 32}}
	for i, t := range tiers {
		g.Use(White).Fill(oct(t.a, t.a, t.y0, t.y1, 1.3), oct(t.a, t.a, t.y1-1, t.y1, 0))
		g.Use(White).Fill(oct(t.a+0.4, t.a+0.4, t.y1-0.8, t.y1, 1.0))
		g.Paint(Stone, oct(t.a+0.45, t.a+0.45, t.y1-0.8, t.y1-0.4, 1.0))
		g.Paint(Charcoal, oct(t.a-0.65, t.a-0.65, t.y0, t.y1-1, 0.7))
		// Corner pilasters.
		for j := 0; j < 8; j++ {
			at := Polar(c, t.a/math.Cos(math.Pi/8)-0.2, math.Pi/8+float64(j)*math.Pi/4)
			g.Use(White).Fill(Capsule{at.At(t.y0 + 0.5), at.At(t.y1 - 0.8), 0.45})
		}
		// Belfry arches in the upper part of each tier, a bell inside.
		sill := t.y1 - 6.5
		if i == 0 {
			g.Carve(Arch{V(c.X, 1.0, c.Y+t.a-0.5), 1.0, 2.0, 2.6, AxisZ})
			for j := 0; j < 8; j += 2 {
				g.Carve(RotateY(Arch{V(c.X+t.a-0.5, 4.5, c.Y), 0.5, 1.3, 2.6, AxisX}, c, float64(j)*math.Pi/4+math.Pi/4))
			}
		}
		for j := 0; j < 8; j++ {
			hw := 1.05 - 0.1*float64(i)
			g.Carve(RotateY(Arch{V(c.X+t.a-0.5, sill, c.Y), hw, 3.6 - 0.4*float64(i), 2.6, AxisX}, c, float64(j)*math.Pi/4))
		}
		g.Use(Gold).Fill(Frustum{c, 1.6 - 0.2*float64(i), 0.7, sill + 0.4, sill + 3.0}, Sphere(c.At(sill+3.0), 0.7))
		g.Use(Charcoal).Fill(oct(t.a-1.3, t.a-1.3, sill-0.4, sill, 0))
	}
	// Drum with its gold inscription band, then the great onion.
	g.Use(White).Fill(Cylinder{c, 2.7, 1.5, 31.5, 37.5}, Cylinder{c, 1.6, 0, 36.6, 37.5})
	g.Paint(Charcoal, Cylinder{c, 2.1, 1.4, 31.5, 36.6})
	for j := 0; j < 8; j++ {
		g.Carve(RotateY(Arch{V(c.X+2.3, 32.6, c.Y), 0.45, 1.8, 2, AxisX}, c, float64(j)*math.Pi/4+math.Pi/8))
	}
	g.Use(Gold).Fill(Cylinder{c, 3.0, 1.8, 36.7, 37.6})
	g.Use(Gold).FillVolume(k.onion(c, 37.5, 2.3, 3.4, 7.6))
	g.Use(Gold).Fill(
		Capsule{c.At(44.6), c.At(48.4), 0.45},
		Capsule{V(c.X-1.2, 47.1, c.Y), V(c.X+1.2, 47.1, c.Y), 0.42},
	)
}

// palace builds the Grand Kremlin Palace facing the river: yellow walls with white
// pilasters, cornices and window frames, a green hipped roof and a central belvedere.
func (k Kremlin) palace(g *Grid, x0, x1, z0, z1, b, h float64) {
	cx := (x0 + x1) / 2
	// Terrace on the hill, then the yellow storeys.
	g.Use(Stone).Fill(Block(V(x0-0.6, 0, z0-0.6), V(x1+0.6, b, z1+0.6)))
	g.Use(White).Fill(Block(V(x0-0.4, b-0.4, z0-0.4), V(x1+0.4, b+1.0, z1+0.4)))
	g.Use(Sand).Fill(Block(V(x0, b+1.0, z0), V(x1, h, z1)))
	g.Carve(Block(V(x0+1.2, b+1.0, z0+1.2), V(x1-1.2, h-1.2, z1-1.2)))
	g.Paint(Charcoal, Block(V(x0+0.6, b+1.0, z0+0.6), V(x1-0.6, h-1.2, z1-0.6)))
	// Central risalit.
	g.Use(Sand).Fill(Block(V(cx-3.4, b+1.0, z1), V(cx+3.4, h+1.4, z1+0.8)))
	mid := b + 5.2
	g.Use(White).Fill(
		Block(V(x0-0.3, h-0.8, z0-0.3), V(x1+0.3, h, z1+0.3)),
		Block(V(x0-0.3, mid, z0-0.3), V(x1+0.3, mid+0.5, z1+0.3)),
		Block(V(cx-3.7, h+0.8, z1-0.2), V(cx+3.7, h+1.4, z1+1.1)),
	)
	// Pilasters and windows on the river façade.
	for x := x0 + 0.4; x <= x1+0.01; x += 2.4 {
		zf := z1
		if math.Abs(x-cx) <= 3.5 {
			zf = z1 + 0.8
		}
		g.Use(White).Fill(Block(V(x-0.35, b+1.0, zf), V(x+0.35, h-0.8, zf+0.4)))
		if x+1.2 > x1 {
			continue
		}
		wx := x + 1.2
		g.Paint(White, Block(V(wx-0.85, b+1.6, zf-0.4), V(wx+0.85, b+4.5, zf+0.4)), Arch{V(wx, mid+1.0, zf), 0.85, 2.2, 0.8, AxisZ})
		g.Carve(Block(V(wx-0.5, b+2.0, zf-1.6), V(wx+0.5, b+4.1, zf+0.5)), Arch{V(wx, mid+1.4, zf-0.5), 0.5, 2.1, 3, AxisZ})
	}
	for z := z0 + 1.6; z <= z1-1.2; z += 2.4 {
		for _, xf := range []float64{x0, x1} {
			g.Carve(Block(V(xf-0.7, b+2.0, z-0.5), V(xf+0.7, b+4.1, z+0.5)), Arch{V(xf, mid+1.4, z), 0.5, 2.1, 2.6, AxisX})
		}
	}
	g.Use(Foliage).Fill(HipRoof{P(x0-0.2, z0-0.2), P(x1+0.2, z1+0.2), h})
	// Belvedere with a gilded flagpole.
	bc := P(cx, (z0+z1)/2)
	g.Use(Sand).Fill(BlockAt(bc, 4.4, 4.4, h, h+5.5))
	g.Use(White).Fill(Subtract(BlockAt(bc, 4.8, 4.8, h+4.8, h+5.5), BlockAt(bc, 3, 3, h, h+6)))
	for i := 0; i < 4; i++ {
		g.Carve(RotateY(Arch{V(bc.X+2.2, h+1.8, bc.Y), 0.6, 1.5, 1.6, AxisX}, bc, float64(i)*math.Pi/2))
	}
	g.Use(Foliage).Fill(HipRoof{P(bc.X-2.4, bc.Y-2.4), P(bc.X+2.4, bc.Y+2.4), h + 5.5})
	g.Use(Gold).Fill(Capsule{bc.At(h + 7), bc.At(h + 11), 0.4})
	g.Use(Red).Fill(Block(V(bc.X+0.3, h+9.2, bc.Y-0.3), V(bc.X+3, h+10.9, bc.Y+0.3)))
}

// terem is the Terem Palace block crowned by the cluster of small gold domes of the
// Upper Saviour Cathedral behind the Grand Kremlin Palace.
func (k Kremlin) terem(g *Grid, x0, x1, z0, z1, h float64) {
	g.Use(White).Fill(Block(V(x0, 0, z0), V(x1, h, z1)))
	g.Carve(Block(V(x0+1.2, 1, z0+1.2), V(x1-1.2, h-1, z1-1.2)))
	g.Use(Red).Fill(Block(V(x0-0.3, h-1.6, z0-0.3), V(x1+0.3, h-0.8, z1+0.3)))
	g.Use(White).Fill(Block(V(x0-0.3, h-0.8, z0-0.3), V(x1+0.3, h, z1+0.3)))
	for x := x0 + 1.4; x <= x1-1; x += 2.0 {
		g.Carve(Arch{V(x, h*0.45, z1), 0.45, 1.3, 2.6, AxisZ}, Arch{V(x, h*0.45, z0), 0.45, 1.3, 2.6, AxisZ})
	}
	i := 0
	for x := x0 + 1.3; x <= x1-1; x += 1.7 {
		for _, z := range []float64{z0 + 1.1, z1 - 1.1} {
			i++
			if i%3 == 0 && z > z0+2 {
				continue
			}
			at := P(x, z)
			g.Use(White).Fill(Cylinder{at, 0.65, 0, h, h + 1.2})
			g.Use(Gold).FillVolume(k.onion(at, h+1.2, 0.55, 0.95, 2.2))
			g.Use(Gold).Fill(Capsule{at.At(h + 3.2), at.At(h + 4.2), 0.35})
		}
	}
}

// senate is the triangular Senate with its green rotunda dome.
func (k Kremlin) senate(g *Grid, tri []Vec2, dome Vec2, h float64) {
	g.Use(Sand).Fill(Prism{tri, 0, h})
	inner := make([]Vec2, len(tri))
	cen := P(0, 0)
	for _, v := range tri {
		cen = cen.Add(v.Scale(1 / float64(len(tri))))
	}
	for i, v := range tri {
		inner[i] = cen.Add(v.Sub(cen).Scale(0.55))
	}
	g.Carve(Prism{inner, 1, h + 1})
	g.Use(White).Fill(Subtract(Prism{tri, h - 0.7, h}, Prism{inner, 0, h + 1}))
	g.Use(White).Fill(Subtract(Prism{tri, 0, 1.2}, Prism{inner, 0, h + 1}))
	// Windows: a ring band of dark openings.
	tb := Prism{tri, 2.5, 6.2}.Bounds()
	g.Paint(Charcoal, Func{tb, func(p Vec3) bool {
		return InsidePolygon(tri, p.Plan()) && math.Mod(math.Abs(p.X+p.Z)+0.3, 2.2) < 0.9 && (p.Y < 4.0 || p.Y > 4.8)
	}})
	g.Paint(Sand, Prism{inner, 0, h + 1})
	g.Use(Foliage).Fill(Subtract(Prism{tri, h, h + 1.0}, Prism{inner, h - 1, h + 2}))
	g.Use(Sand).Fill(Cylinder{dome, 3.2, 2.0, 0, h + 2.5})
	g.Use(White).Fill(Cylinder{dome, 3.5, 2.4, h + 2, h + 2.6})
	for i := 0; i < 12; i++ {
		at := Polar(dome, 3.25, float64(i)*math.Pi/6)
		g.Use(White).Fill(Capsule{at.At(h - 0.5), at.At(h + 2), 0.4})
	}
	g.Use(Foliage).FillVolume(Ellipsoid{dome.At(h + 2.5), V(3.3, 2.8, 3.3)})
	g.Use(Sand).Fill(Cylinder{dome, 0.8, 0, h + 5.0, h + 6.2})
	g.Use(Gold).Fill(Capsule{dome.At(h + 6), dome.At(h + 10), 0.4})
	g.Use(Red).Fill(Block(V(dome.X+0.3, h+8.4, dome.Y-0.3), V(dome.X+2.8, h+9.9, dome.Y+0.3)))
}

// mausoleum is Lenin's stepped red and black granite tomb with tribunes and spruces.
func (k Kremlin) mausoleum(g *Grid, c Vec2) {
	levels := []struct {
		wx, wz, y0, y1 float64
		col            Color
	}{
		{6.6, 11, 0, 2.4, Red}, {5.4, 9.2, 2.4, 3.2, Charcoal}, {4.6, 7.6, 3.2, 4.4, Red},
		{3.8, 6.0, 4.4, 5.1, Charcoal}, {3.0, 4.4, 5.1, 6.5, Red}, {3.4, 4.8, 6.5, 7.2, Charcoal},
	}
	for _, l := range levels {
		g.Use(l.col).Fill(BlockAt(c, l.wx, l.wz, l.y0, l.y1))
	}
	g.Paint(Charcoal, Subtract(BlockAt(c, 7, 12, 1.2, 1.6), BlockAt(c, 6, 10, 0, 3)))
	// Portico columns on top.
	g.Carve(BlockAt(c, 2.2, 3.6, 5.3, 6.4))
	for _, dz := range []float64{-1.8, 0, 1.8} {
		for _, dx := range []float64{-1.2, 1.2} {
			g.Use(Charcoal).Fill(Cylinder{P(c.X+dx, c.Y+dz), 0.45, 0, 5.1, 6.5})
		}
	}
	// Entrance facing the square.
	g.Carve(Block(V(c.X+2.5, 0.4, c.Y-1), V(c.X+3.4, 2.0, c.Y+1)))
	g.Use(Charcoal).Fill(Block(V(c.X+2.3, 0.4, c.Y-1.2), V(c.X+2.5, 2.2, c.Y+1.2)))
	// Tribunes.
	for _, s := range []float64{-1, 1} {
		z0, z1 := c.Y+s*6.2, c.Y+s*11.5
		g.Use(Stone).Fill(Block(V(c.X-3, 0, z0), V(c.X+1.2, 1.4, z1)), Block(V(c.X-3, 1.4, z0), V(c.X-0.4, 2.4, z1)))
	}
}

// basils is a simplified St. Basil's at the south end of the square.
func (k Kremlin) basils(g *Grid, c Vec2) {
	g.Use(Brick).Fill(BlockAt(c, 11, 11, 0, 2.4))
	g.Carve(BlockAt(c, 8.6, 8.6, 1.0, 1.6))
	g.Paint(Terracotta, BlockAt(c, 11, 11, 2.0, 2.4))
	g.Use(White).Fill(Subtract(BlockAt(c, 11.6, 11.6, 1.9, 2.4), BlockAt(c, 10, 10, 0, 4)))
	for i := 0; i < 4; i++ {
		for _, off := range []float64{-3.4, 0, 3.4} {
			g.Carve(RotateY(Arch{V(c.X+5.1, 0.4, c.Y+off), 0.55, 0.6, 2, AxisX}, c, float64(i)*math.Pi/2))
		}
	}
	oct := func(cc Vec2, a0, a1, y0, y1, thick float64) Solid {
		return k.prism(cc, 0, kremlinOctagon, a0, a1, y0, y1, thick)
	}
	// Central tent.
	g.Use(Brick).Fill(oct(c, 2.5, 2.5, 2.4, 12.5, 1.2))
	g.Use(White).Fill(oct(c, 2.85, 2.85, 11.8, 12.5, 0.8))
	g.Use(Brick).FillVolume(oct(c, 2.7, 0.6, 12.5, 21.5, 1.2))
	g.Paint(White, k.ribs(c, 0, 3.5, 12.5, 21, 8, 0.5, 0.07))
	g.Use(Gold).FillVolume(k.onion(c, 21.3, 0.7, 1.3, 2.8))
	g.Use(Gold).Fill(Capsule{c.At(23.8), c.At(25.5), 0.4})
	type ch struct {
		at        Vec2
		big       bool
		base, pat Color
	}
	chs := []ch{
		{P(c.X, c.Y+4.1), true, Grass, Gold}, {P(c.X+4.1, c.Y), true, White, Red},
		{P(c.X, c.Y-4.1), true, Gold, Foliage}, {P(c.X-4.1, c.Y), true, White, Navy},
		{P(c.X+3.6, c.Y+3.6), false, Red, Gold}, {P(c.X+3.6, c.Y-3.6), false, Grass, Red},
		{P(c.X-3.6, c.Y-3.6), false, Gold, Grass}, {P(c.X-3.6, c.Y+3.6), false, Navy, White},
	}
	for i, h := range chs {
		a, top, R, H := 1.8, 10.5, 2.3, 5.0
		if !h.big {
			a, top, R, H = 1.3, 8.0, 1.7, 3.8
		}
		g.Use(Brick).Fill(oct(h.at, a, a, 2.4, top, 0))
		g.Carve(RotateY(Arch{V(h.at.X, top-4.2, h.at.Y), 0.45, 1.6, 2*a + 1, AxisX}, h.at, math.Atan2(h.at.Y-c.Y, h.at.X-c.X)))
		g.Use(White).Fill(oct(h.at, a+0.3, a+0.3, top-0.6, top, 0.6))
		g.Use(Brick).Fill(Cylinder{h.at, a * 0.65, 0, top, top + 1.4})
		g.Use(h.base).FillVolume(k.onion(h.at, top+1.4, a*0.6, R, H))
		n := 6.0
		twist := 2.0
		if i%2 == 1 {
			twist = -2
		}
		at := h.at
		g.Paint(h.pat, Func{Box{V(at.X-R, top+1.9, at.Y-R), V(at.X+R, top+1.4+H*0.85, at.Y+R)}, func(p Vec3) bool {
			q := p.Plan().Sub(at)
			t := (p.Y - top - 1.4) / H
			return int(math.Floor(math.Atan2(q.Y, q.X)/(2*math.Pi)*n+t*twist+10))%2 == 0
		}})
		g.Use(Gold).Fill(Capsule{at.At(top + 1.4 + H - 0.4), at.At(top + 1.4 + H + 1.3), 0.4})
	}
}

// gum is the long arcaded trade rows facing the Kremlin across the square.
func (k Kremlin) gum(g *Grid, x0, x1, z0, z1, h float64) {
	g.Use(Stone).Fill(Block(V(x0, 0, z0), V(x1, h, z1)))
	g.Carve(Block(V(x0+1.2, 1, z0+1.2), V(x1-1.2, h-1, z1-1.2)))
	g.Paint(Charcoal, Block(V(x0+0.6, 1, z0+0.6), V(x1-0.6, h-1, z1-0.6)))
	g.Use(White).Fill(Block(V(x0-0.3, h-0.6, z0-0.3), V(x1+0.3, h, z1+0.3)), Block(V(x0-0.3, 3.0, z0-0.3), V(x1+0.3, 3.5, z1+0.3)))
	for z := z0 + 1.6; z <= z1-1.2; z += 2.2 {
		for _, x := range []float64{x0, x1} {
			g.Carve(Arch{V(x, 0.9, z), 0.6, 1.0, 2.6, AxisX}, Arch{V(x, 3.8, z), 0.5, 0.9, 2.6, AxisX})
		}
	}
	g.Use(Slate).Fill(Gable{P(x0, z0), P(x1, z1), h, AxisZ})
	for _, z := range []float64{z0 + 3, (z0 + z1) / 2, z1 - 3} {
		at := P(x0+1.2, z)
		g.Use(Stone).Fill(BlockAt(at, 3, 3.4, 0, h+3.5))
		g.Use(White).Fill(BlockAt(at, 3.4, 3.8, h+3, h+3.5))
		g.Carve(Arch{V(at.X-1.5, 0.4, z), 1.0, 2.0, 2, AxisX})
		g.Use(Foliage).FillVolume(k.prism(at, 0, kremlinSquare, 1.7, 0.3, h+3.5, h+7, 0))
		g.Use(Gold).Fill(Capsule{at.At(h + 6.6), at.At(h + 8), 0.35})
	}
}

// museum is the red Historical Museum with its spired towers at the square's north end.
func (k Kremlin) museum(g *Grid, x0, x1, z0, z1, h float64) {
	g.Use(Brick).Fill(Block(V(x0, 0, z0), V(x1, h, z1)))
	g.Carve(Block(V(x0+1.2, 1, z0+1.2), V(x1-1.2, h-1, z1-1.2)))
	g.Paint(Charcoal, Block(V(x0+0.6, 1, z0+0.6), V(x1-0.6, h-1, z1-0.6)))
	g.Use(White).Fill(Block(V(x0-0.3, h-0.6, z0-0.3), V(x1+0.3, h, z1+0.3)))
	for x := x0 + 1.5; x <= x1-1; x += 2.2 {
		g.Carve(Arch{V(x, 1.6, z1), 0.5, 1.2, 2.6, AxisZ}, Arch{V(x, 5.0, z1), 0.5, 1.4, 2.6, AxisZ})
	}
	g.Use(Slate).Fill(Gable{P(x0, z0), P(x1, z1), h, AxisX})
	cx := (x0 + x1) / 2
	for _, dx := range []float64{-2.4, 2.4} {
		at := P(cx+dx, z1-1)
		g.Use(Brick).Fill(BlockAt(at, 2.8, 2.8, 0, h+6))
		g.Use(White).Fill(BlockAt(at, 3.2, 3.2, h+5.4, h+6))
		g.Carve(Arch{V(at.X, h+1.6, at.Y), 0.5, 1.6, 4, AxisZ})
		g.Use(Stone).FillVolume(k.prism(at, 0, kremlinOctagon, 1.5, 0.35, h+6, h+13, 0))
		g.Use(Gold).Fill(Capsule{at.At(h + 12.6), at.At(h + 14.4), 0.4})
	}
	for _, at := range []Vec2{P(x0+0.8, z1-0.8), P(x1-0.8, z1-0.8), P(x0+0.8, z0+0.8), P(x1-0.8, z0+0.8), P(cx, z0+0.8)} {
		g.Use(Brick).Fill(BlockAt(at, 2, 2, 0, h+2))
		g.Use(Stone).FillVolume(k.prism(at, 0, kremlinOctagon, 1.1, 0.3, h+2, h+6, 0))
	}
}

// tree is a broadleaf (crown) or a blue spruce (cone).
func (k Kremlin) tree(g *Grid, at Vec2, h float64, spruce bool) {
	if spruce {
		g.Use(Wood).Fill(Cylinder{at, 0.45, 0, 0, 1.2})
		g.Use(Foliage).FillVolume(Cone(at, 1.3, 0.8, h))
		return
	}
	g.Use(Wood).Fill(Cylinder{at, 0.5, 0, 0, h * 0.5})
	g.Use(Foliage).FillVolume(Ellipsoid{at.At(h * 0.68), V(h*0.34, h*0.3, h*0.34)})
}

// outside is the signed distance outside the north-west (Alexander Garden) wall line.
func (Kremlin) outside(p Vec2) float64 {
	d := kremlinN.Sub(kremlinSW)
	d = d.Scale(1 / d.Length())
	n := P(d.Y, -d.X) // Points north-west, away from the Kremlin.
	return p.Sub(kremlinSW).Dot(n)
}

func (k Kremlin) Build(g *Grid) Order {
	const (
		x0, x1 = -46.0, 45.0
		zN     = -44.0
		road   = 19.5
		bank   = 21.0
		river  = 28.0
		far    = 30.0
	)
	inside := func(p Vec2) bool {
		return k.outside(p) < 0 && p.X < kremlinSE.X && p.Y < kremlinSW.Y
	}
	// Ground: pavement and gardens on a thin plate, the river a step lower.
	land := Func{Box{V(x0, 0, zN), V(x1, 0.4, bank)}, func(p Vec3) bool { return k.outside(p.Plan()) <= 7.5 }}
	g.Use(DarkStone).Fill(land)
	g.Paint(Grass, Func{land.Box, func(p Vec3) bool {
		q := p.Plan()
		if inside(q) {
			return true
		}
		o := k.outside(q)
		return o > 1.8 && o < 6.8 && q.X < kremlinSE.X && math.Mod(o+20, 4.4) > 0.8
	}})
	// Paved Cathedral Square and lanes inside the walls.
	g.Paint(Stone, BlockAt(P(2, 1), 24, 26, 0, 0.4), Block(V(-34, 0, 13.5), V(18, 0.4, 14.5)))
	g.Paint(Stone, Capsule{V(18, 0.3, -12), V(2, 0.3, -12), 1.2}, Capsule{V(18, 0.3, -24), V(8, 0.3, -18), 1.2})
	// Red Square: dark setts with lighter stones and paving lines.
	sqr := Box{V(21.4, 0, -40), V(x1, 0.4, road)}
	g.Paint(Stone, Func{sqr, func(p Vec3) bool {
		i, j := int(math.Floor(p.X/1.6)), int(math.Floor(p.Z/1.6))
		return Hash01(i, j, 11) < 0.12
	}})
	g.Paint(DarkStone, Block(V(x0, 0, kremlinSW.Y+1.5), V(x1, 0.4, bank)))
	// Embankment parapet, the river, the far bank.
	g.Use(Stone).Fill(Block(V(x0, -1.2, bank-1.2), V(x1, 1.2, bank)))
	g.Use(Water).Fill(Block(V(x0, -1.2, bank), V(x1, -0.6, river)))
	g.Use(Stone).Fill(Block(V(x0, -1.2, river), V(x1, 0.8, river+1.2)))
	g.Use(DarkStone).Fill(Block(V(x0, -1.2, river+1.2), V(x1, 0, far)))
	// Bolshoy Moskvoretsky bridge from the square's south end across the river.
	{
		bx0, bx1 := 30.0, 36.5
		g.Use(Stone).Fill(Block(V(bx0, -1.2, road), V(bx1, 0.9, far)))
		for _, z := range []float64{23.2, 26.8} {
			g.Carve(Arch{V((bx0+bx1)/2, -1.2, z), 1.4, 0.7, bx1 - bx0 + 2, AxisX})
		}
		g.Paint(Water, Block(V(bx0, -1.2, bank), V(bx1, -0.6, river)))
		g.Use(Water).Fill(Block(V(bx0, -1.2, bank+0.5), V(bx1, -0.6, river-0.5)))
		g.Use(DarkStone).Fill(Block(V(bx0+0.6, 0.9, road), V(bx1-0.6, 1.1, far)))
		g.Use(Stone).Fill(Block(V(bx0, 0.9, bank-1.5), V(bx0+0.6, 1.9, far)), Block(V(bx1-0.6, 0.9, bank-1.5), V(bx1, 1.9, far)))
		g.Carve(Block(V(bx0+0.6, 0.6, bank-1.2), V(bx1-0.6, 1.3, bank)))
		g.Use(DarkStone).Fill(Block(V(bx0+0.6, 0.6, bank-1.2), V(bx1-0.6, 0.9, bank)))
	}

	// Walls: south along the river, east along Red Square, north-west along the garden.
	k.wall(g, kremlinSW, kremlinSE, 1, 6.5)
	k.wall(g, kremlinSE, kremlinN, 1, 8)
	k.wall(g, kremlinN, kremlinSW, 1, 8)

	// Towers.
	hyp := kremlinN.Sub(kremlinSW)
	hypAngle := math.Atan2(hyp.Y, hyp.X)
	along := func(t float64) Vec2 { return kremlinSW.Add(hyp.Scale(t)) }
	k.tower(g, kremlinTower{c: kremlinSE, round: true, a: 3.8, h: 12, stage: 3.2, tent: 13, tentCol: Foliage})
	k.tower(g, kremlinTower{c: kremlinSW, round: true, a: 4.4, h: 13, stage: 3.4, tent: 13, lantern: true, star: true})
	k.tower(g, kremlinTower{c: kremlinN, round: true, a: 4.8, h: 13, stage: 2.6, tent: 7})
	k.tower(g, kremlinTower{c: P(-8, 16), a: 3.2, h: 10, stage: 3.2, tent: 9})
	k.tower(g, kremlinTower{c: P(6, 16), a: 2.6, h: 10, tent: 8})
	k.tower(g, kremlinTower{c: P(-24, 16), a: 2.6, h: 10, tent: 8})
	k.tower(g, kremlinTower{c: P(20, -12), angle: math.Pi / 2, a: 3.0, h: 11, stage: 2.6, tent: 9})
	k.tower(g, kremlinTower{c: P(20, -24), angle: math.Pi / 2, a: 3.3, h: 13, stage: 4.0, tent: 12, lantern: true, star: true, gate: true})
	k.tower(g, kremlinTower{c: along(0.13), angle: hypAngle, a: 3.2, h: 11, stage: 3.6, tent: 11, lantern: true, star: true, gate: true})
	k.tower(g, kremlinTower{c: along(0.62), angle: hypAngle, a: 3.5, h: 14, stage: 4.4, tent: 14, lantern: true, star: true, gate: true})
	k.tower(g, kremlinTower{c: along(0.4), angle: hypAngle, a: 2.6, h: 10, tent: 8})
	k.spasskaya(g, P(20, 5))

	// Cathedral Square.
	k.cathedral(g, P(0, -5), 11, 8.4, 12, 4, 3, 5, DarkStone, []kremlinDome{
		{P(0, -5), 2.0, 3.6, 2.7, 5.4, Gold},
		{P(-3, -7.2), 1.5, 2.8, 2.0, 4.0, Gold}, {P(3, -7.2), 1.5, 2.8, 2.0, 4.0, Gold},
		{P(-3, -2.8), 1.5, 2.8, 2.0, 4.0, Gold}, {P(3, -2.8), 1.5, 2.8, 2.0, 4.0, Gold},
	})
	k.cathedral(g, P(5.5, 9.6), 8.6, 7, 10.5, 3, 3, 3, DarkStone, []kremlinDome{
		{P(5.5, 9.6), 1.8, 3.2, 2.4, 4.8, Gold},
		{P(3.2, 7.8), 1.3, 2.4, 1.7, 3.4, Stone}, {P(7.8, 7.8), 1.3, 2.4, 1.7, 3.4, Stone},
		{P(3.2, 11.4), 1.3, 2.4, 1.7, 3.4, Stone}, {P(7.8, 11.4), 1.3, 2.4, 1.7, 3.4, Stone},
	})
	k.cathedral(g, P(-4.6, 9.4), 6.6, 6, 9.5, 2, 2, 3, Gold, []kremlinDome{
		{P(-4.6, 9.4), 1.5, 3.0, 2.0, 4.2, Gold},
		{P(-6.3, 8.0), 1.1, 2.0, 1.4, 3.0, Gold}, {P(-2.9, 8.0), 1.1, 2.0, 1.4, 3.0, Gold},
		{P(-6.3, 10.8), 1.1, 2.0, 1.4, 3.0, Gold}, {P(-2.9, 10.8), 1.1, 2.0, 1.4, 3.0, Gold},
	})
	k.ivan(g, P(11, -1.5))
	k.terem(g, -20.5, -12, 2.4, 5.8, 10)
	k.palace(g, -27, -10, 6.4, 14.2, 4.5, 16)
	k.senate(g, []Vec2{P(7.5, -15.5), P(18, -15.5), P(16.5, -27)}, P(14.6, -18.4), 7.5)

	// Red Square.
	k.mausoleum(g, P(26, -12))
	for z := -30.0; z <= 0; z += 2.8 {
		if math.Abs(z+12) < 7 || math.Abs(z+24) < 3 {
			continue
		}
		k.tree(g, P(22.4, z), 4+Hash01(int(z*10), 3)*1.2, true)
	}
	k.basils(g, P(34, 9.5))
	k.gum(g, 39, 44, -29, 0, 6.5)
	k.museum(g, 24, 38, -43.5, -38, 9)

	// Gardens: the Alexander Garden outside the north-west wall, trees inside.
	for t := 0.06; t <= 0.95; t += 0.075 {
		at := along(t)
		d := P(-hyp.Y, hyp.X).Scale(1 / hyp.Length())
		if k.outside(at.Add(d)) < 0 {
			d = d.Scale(-1)
		}
		p := at.Add(d.Scale(4.5))
		if p.X > 17 {
			continue
		}
		k.tree(g, p, 5+Hash01(int(t*1000), 1)*1.5, false)
	}
	for i, at := range []Vec2{P(14, 12.5), P(17, 13.5), P(-31, 12.5), P(-34, 13.6), P(-20, 3.5), P(-14, 4), P(-24, 3.8),
		P(12, -21), P(4, -14.5), P(-8, -10), P(-14, -4.5)} {
		if !inside(at) {
			continue
		}
		k.tree(g, at, 4.5+Hash01(i, 5)*1.5, false)
	}
	return Sweep(P(0, 0), math.Pi/2)
}
