package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// AmsterdamCanals is a block of Amsterdam's Golden Age canal ring. Two parallel canals
// curve around a far-away center (concentric arcs bowing towards +Z) with brick quay
// walls, cobbled quays and elms. Continuous rows of tall narrow canal houses line the
// quays: step, spout, neck, bell and cornice gables with white trim and hoist beams,
// many-paned white-framed windows over dark interiors, brick, grey, black, navy, green
// and white facades, and a slight forward lean. A Westerkerk-style church stands in the
// middle block, its brick and stone tower crowned with a blue and gold imperial crown.
// Brick arch bridges carry a cross street over both canals and a white double-leaf
// drawbridge in the style of the Magere Brug crosses the front canal; houseboats are
// moored along the quays, tour boats and a sloop are under way, bicycles lean by the
// water. The front quay is a low street-level plate.
type AmsterdamCanals struct{}

func init() { Register(AmsterdamCanals{}) }

func (AmsterdamCanals) Name() string  { return "amsterdam_canals" }
func (AmsterdamCanals) Title() string { return "Amsterdam Canals" }

const (
	amsterdamCanalsRc   = 200.0                    // Radius of the reference arc d = 0; the ring's center is (0, -Rc).
	amsterdamCanalsHalf = 42.0 / amsterdamCanalsRc // Half angle of the site.
	amsterdamCanalsQuay = 1.5                      // Quay (street) level; water is at 0.
	amsterdamCanalsBase = 1.0                      // Bottom of the house shells.
)

// Radial layout (d = r - Rc, +d towards world +Z).
var (
	amsterdamCanalsCanals = [][2]float64{{-18.5, -7.5}, {20, 31}}
	amsterdamCanalsLand   = [][2]float64{{-31, -18.5}, {-7.5, 20}, {31, 35}}
)

// Cross streets and landmarks, as plan angles.
const (
	amsterdamCanalsStreet  = 14.0 / amsterdamCanalsRc  // Cross street with brick bridges over both canals.
	amsterdamCanalsDraw    = -6.5 / amsterdamCanalsRc  // Drawbridge over the front canal, with its lane.
	amsterdamCanalsChurch  = -23.5 / amsterdamCanalsRc // Church axis.
	amsterdamCanalsChurchW = 7.4 / amsterdamCanalsRc   // Half angle of the church plot.
)

// amsterdamCanalsPolar returns the plan angle (0 along +Z, increasing towards +X) and
// radius of a world plan point about the ring center.
func amsterdamCanalsPolar(p Vec2) (float64, float64) {
	q := P(p.X, p.Y+amsterdamCanalsRc)
	return math.Atan2(q.X, q.Y), q.Length()
}

func amsterdamCanalsAt(th, r float64) Vec2 {
	s, c := math.Sincos(th)
	return P(r*s, r*c-amsterdamCanalsRc)
}

// amsterdamCanalsFrame is a curved local frame: local X is arc length at radius r0
// from angle th, local Z is r - r0 (outward). flip turns it half round.
type amsterdamCanalsFrame struct {
	th, r0 float64
	flip   bool
}

type amsterdamCanalsPlaced struct {
	s   Solid
	f   amsterdamCanalsFrame
	box Box
}

func (v amsterdamCanalsPlaced) Contains(p Vec3) bool {
	th, r := amsterdamCanalsPolar(p.Plan())
	x, z := v.f.r0*(th-v.f.th), r-v.f.r0
	if v.f.flip {
		x, z = -x, -z
	}
	return v.s.Contains(V(x, p.Y, z))
}

func (v amsterdamCanalsPlaced) Bounds() Box { return v.box }

func (f amsterdamCanalsFrame) at(s Solid) Solid {
	lb := s.Bounds()
	x0, x1, z0, z1 := lb.Min.X, lb.Max.X, lb.Min.Z, lb.Max.Z
	if f.flip {
		x0, x1, z0, z1 = -x1, -x0, -z1, -z0
	}
	b := EmptyBox()
	steps := int(math.Ceil((x1-x0)/2)) + 1
	for i := 0; i <= steps; i++ {
		th := f.th + (x0+(x1-x0)*float64(i)/float64(steps))/f.r0
		for _, z := range []float64{z0, z1} {
			q := amsterdamCanalsAt(th, f.r0+z)
			b = b.Union(Box{q.At(lb.Min.Y), q.At(lb.Max.Y)})
		}
	}
	return amsterdamCanalsPlaced{s, f, b.Grow(0.6)}
}

// amsterdamCanalsSector is the ring sector between two angles and two radial offsets.
func amsterdamCanalsSector(th0, th1, d0, d1, y0, y1 float64) Solid {
	f := amsterdamCanalsFrame{th: (th0 + th1) / 2, r0: amsterdamCanalsRc}
	hw := (th1 - th0) / 2 * amsterdamCanalsRc
	return f.at(Block(V(-hw, y0, d0), V(hw, y1, d1)))
}

// Gable styles.
const (
	amsterdamCanalsSpout = iota
	amsterdamCanalsStep
	amsterdamCanalsNeck
	amsterdamCanalsBell
	amsterdamCanalsCornice
)

// amsterdamCanalsHouse is one canal house in a frame whose local origin is the facade
// center at street level: x across the facade, the facade on z = 0 facing +Z.
type amsterdamCanalsHouse struct {
	f                  amsterdamCanalsFrame
	w, depth           float64
	storeys            int
	gable              int
	facade, trim, roof Color
	lean               Vec2
	seed               int
	sides              []float64 // Exposed side walls (local -1 / +1) that get windows.
}

func (h amsterdamCanalsHouse) top() float64 {
	return amsterdamCanalsBase + 3.4 + float64(h.storeys-1)*2.5
}

// gableHeight is the height of the front gable above the cornice line.
func (h amsterdamCanalsHouse) gableHeight() float64 {
	hw := h.w / 2
	switch h.gable {
	case amsterdamCanalsSpout:
		return hw + 1.4
	case amsterdamCanalsStep:
		return float64(h.steps()+1) * 1.15
	case amsterdamCanalsNeck:
		return 5.8
	case amsterdamCanalsBell:
		return 5.2
	}
	return 2.4
}

func (h amsterdamCanalsHouse) steps() int {
	if h.w < 5.2 {
		return 3
	}
	return 4
}

// inGable tests a point of the front gable: gx = |x|, y above the cornice line.
func (h amsterdamCanalsHouse) inGable(gx, y float64) bool {
	hw := h.w / 2
	if y < 0 || gx > hw {
		return false
	}
	switch h.gable {
	case amsterdamCanalsSpout:
		rise := hw
		if y <= rise {
			return gx <= hw-y*(hw-0.8)/rise
		}
		return y <= rise+1.4 && gx <= 0.8
	case amsterdamCanalsStep:
		n := h.steps()
		k := int(y / 1.15)
		if k > n {
			return false
		}
		return gx <= hw-float64(k)*(hw-0.8)/float64(n)
	case amsterdamCanalsNeck:
		nh := 1.1
		if hw > 2.8 {
			nh = 1.4
		}
		switch {
		case y <= 1.0:
			return true
		case y <= 2.8:
			return gx <= nh+(hw-nh)*(2.8-y)/1.8
		case y <= 4.6:
			return gx <= nh
		}
		ex, ey := gx/(nh+0.35), (y-4.6)/1.2
		return ex*ex+ey*ey <= 1
	case amsterdamCanalsBell:
		nh := 1.15
		if y <= 3.6 {
			return gx <= nh+(hw-nh)*0.5*(1+math.Cos(math.Pi*y/3.6))
		}
		ex, ey := gx/(nh+0.2), (y-3.6)/1.6
		return ex*ex+ey*ey <= 1
	}
	// Cornice house: attic parapet and a raised central crest.
	if y <= 0.9 {
		return true
	}
	ex, ey := gx/1.5, (y-0.9)/1.5
	return ex*ex+ey*ey <= 1
}

// roofRise is the ridge height of the roof behind the gable, above the cornice line.
func (h amsterdamCanalsHouse) roofRise() float64 {
	if h.gable == amsterdamCanalsCornice {
		return math.Min(h.w/2, 3.2)
	}
	return h.gableHeight() - 0.9
}

func (h amsterdamCanalsHouse) put(s Solid) Solid {
	return h.f.at(Sheared{s, h.lean, amsterdamCanalsBase})
}

// window carves a dark recessed window. Dark painted fronts get white frames.
func (h amsterdamCanalsHouse) window(g *Grid, x, y0, y1, ww float64, lintel bool) {
	if h.dark() {
		g.Paint(h.trim, h.put(Block(V(x-ww/2-0.4, y0-0.4, -0.5), V(x+ww/2+0.4, y1+0.4, 0.5))))
	} else if lintel {
		g.With(h.trim, func() {
			g.Fill(h.put(Block(V(x-ww/2-0.25, y1, -0.3), V(x+ww/2+0.25, y1+0.4, 0.4))))
		})
	}
	g.Carve(h.put(Block(V(x-ww/2, y0, -1.05), V(x+ww/2, y1, 0.6))))
}

func (h amsterdamCanalsHouse) dark() bool {
	return h.facade == Charcoal || h.facade == Navy || h.facade == Foliage || h.facade == DarkStone
}

// sideWindow carves a window in the exposed side wall sx (±1) at depth z, backed dark.
func (h amsterdamCanalsHouse) sideWindow(g *Grid, sx, z, y0, y1 float64) {
	hw := h.w / 2
	xa, xb := sx*(hw-1.05), sx*(hw+0.6)
	g.Carve(h.put(Block(V(math.Min(xa, xb), y0, z-0.55), V(math.Max(xa, xb), y1, z+0.55))))
	xa, xb = sx*(hw-1.0), sx*(hw-2.0)
	g.With(Charcoal, func() {
		g.Fill(h.put(Block(V(math.Min(xa, xb), y0, z-0.6), V(math.Max(xa, xb), y1, z+0.6))))
	})
	xa, xb = sx*(hw-0.3), sx*(hw+0.4)
	g.With(h.trim, func() {
		g.Fill(h.put(Block(V(math.Min(xa, xb), y0-0.4, z-0.75), V(math.Max(xa, xb), y0, z+0.75))))
	})
}

func (h amsterdamCanalsHouse) build(g *Grid) {
	hw, D, B := h.w/2, h.depth, amsterdamCanalsBase
	top := h.top()
	hg := h.gableHeight()
	R := h.roofRise()

	// Shell: front wall 2.2 deep (facade skin plus dark backing), other walls 1.
	g.Use(h.facade)
	g.Fill(h.put(Block(V(-hw, B, -D), V(hw, top, 0))))
	g.Carve(h.put(Block(V(-hw+1, B+0.6, -D+1), V(hw-1, top+0.1, -2.2))))
	gable := Func{Box{V(-hw, top, -2.2), V(hw, top+hg, 0)}, func(p Vec3) bool {
		gx, y := math.Abs(p.X), p.Y-top
		if p.Z < -1.2 {
			return h.inGable(gx+0.8, y+0.8)
		}
		return h.inGable(gx, y)
	}}
	g.Fill(h.put(gable))
	// Painted fronts have brick party and rear walls.
	if h.facade != Brick {
		g.Paint(Brick, h.put(Block(V(-hw, B, -D), V(hw, top+hg, -1.15))))
	}
	// Dark backing behind the windows.
	g.Paint(Charcoal, h.put(Func{Box{V(-hw+0.9, B+0.5, -2.3), V(hw-0.9, top+hg, -1.15)}, func(p Vec3) bool {
		y := p.Y - top
		return y <= 0 || h.inGable(math.Abs(p.X)+0.8, y+0.8)
	}}))
	// Steep roof behind the gable and the plain rear gable.
	tv := 1.1 * math.Sqrt(1+(R/hw)*(R/hw))
	g.With(h.roof, func() {
		g.Fill(h.put(Func{Box{V(-hw, top-0.2, -D), V(hw, top+R, -1)}, func(p Vec3) bool {
			s := R * (1 - math.Abs(p.X)/hw)
			y := p.Y - top
			return y <= s && y >= s-tv
		}}))
	})
	if Hash01(h.seed, 12) < 0.6 {
		x := (hw - 1.0) * math.Copysign(1, Hash01(h.seed, 13)-0.5)
		yr := top + R*(1-math.Abs(x)/hw)
		g.With(Brick, func() {
			g.Fill(h.put(Block(V(x-0.45, yr-1.2, -D+1.6), V(x+0.45, yr+1.3, -D+2.5))))
		})
		g.With(Charcoal, func() {
			g.Fill(h.put(Block(V(x-0.6, yr+1.3, -D+1.45), V(x+0.6, yr+1.7, -D+2.65))))
		})
	}
	rear := h.facade
	if h.facade != Brick {
		rear = Brick
	}
	g.With(rear, func() {
		g.Fill(h.put(Func{Box{V(-hw, top, -D), V(hw, top+R, -D+1)}, func(p Vec3) bool {
			return p.Y-top <= R*(1-math.Abs(p.X)/hw)-0.4
		}}))
	})

	// Cornice band and gable outline.
	if h.gable == amsterdamCanalsCornice {
		g.With(h.trim, func() {
			g.Fill(h.put(Block(V(-hw, top-0.7, -1), V(hw, top+0.1, 0.7))))
		})
	} else {
		g.Paint(h.trim, h.put(Block(V(-hw, top-0.4, -1), V(hw, top+0.2, 0.5))))
	}
	g.Paint(h.trim, h.put(Func{Box{V(-hw, top, -0.5), V(hw, top+hg, 0.3)}, func(p Vec3) bool {
		gx, y := math.Abs(p.X), p.Y-top
		if !h.inGable(gx, y) {
			return false
		}
		if h.gable == amsterdamCanalsStep || h.gable == amsterdamCanalsSpout {
			return !h.inGable(gx, y+0.75)
		}
		return !h.inGable(gx+0.75, y) || !h.inGable(gx, y+0.75)
	}}))

	// Windows: k per storey, white sills and lintels, dark panes.
	k := int(math.Max(2, math.Round(h.w/2.7)))
	const ww = 1.4
	xs := make([]float64, k)
	for i := range xs {
		xs[i] = -hw + (float64(i)+0.5)*h.w/float64(k)
	}
	door := 0
	if Hash01(h.seed, 1) < 0.5 {
		door = k - 1
	}
	for i, x := range xs {
		if i == door {
			g.Paint(h.trim, h.put(Block(V(x-0.9, B+0.5, -0.3), V(x+0.9, B+3.3, 0.5))))
			g.Carve(h.put(Block(V(x-0.6, B+0.55, -1.05), V(x+0.6, B+2.95, 0.6))))
			g.With(Stone, func() {
				g.Fill(h.put(Block(V(x-0.9, amsterdamCanalsQuay, 0.1), V(x+0.9, amsterdamCanalsQuay+0.5, 1.2))))
			})
			continue
		}
		h.window(g, x, B+1.1, B+3.0, ww, false)
	}
	for s := 1; s < h.storeys; s++ {
		yb := B + 3.4 + float64(s-1)*2.5
		for _, x := range xs {
			h.window(g, x, yb+0.4, yb+2.2, ww, false)
		}
	}
	// Windows in exposed side walls.
	for _, sx := range h.sides {
		for _, z := range []float64{-3.2, -h.depth + 2.4} {
			for s := 0; s < h.storeys; s++ {
				yb := B + 3.4 + float64(s-1)*2.5
				if s == 0 {
					yb = B + 0.5
				}
				h.sideWindow(g, sx, z, yb+0.6, yb+2.2)
			}
		}
	}
	// Gable windows and the hoist beam.
	if h.gable != amsterdamCanalsCornice {
		h.window(g, 0, top+0.7, top+2.2, 1.2, false)
		if hg > 4.8 {
			h.window(g, 0, top+3.0, top+3.9, 0.85, false)
		}
		beam := Wood
		if h.facade == Wood || h.facade == Charcoal {
			beam = White
		}
		y := top + hg - 1.0
		if h.gable == amsterdamCanalsSpout {
			y = top + hg - 0.9
		}
		g.With(beam, func() {
			g.Fill(h.put(Block(V(-0.45, y-0.45, -1), V(0.45, y+0.45, 1.5))))
		})
	}
}

// amsterdamCanalsFacades are weighted facade colors.
var amsterdamCanalsFacades = []Color{Brick, Brick, Brick, Brick, DarkStone, Charcoal, Charcoal, Navy, Foliage, White, White, Sand, Brick, Terracotta}

// row lines one quay with houses. dF is the facade line; flip faces the row towards -d.
func (a AmsterdamCanals) row(g *Grid, idx int, dF, depth float64, flip bool, gaps [][2]float64) {
	rF := amsterdamCanalsRc + dF
	// Free angular intervals.
	free := [][2]float64{{-amsterdamCanalsHalf, amsterdamCanalsHalf}}
	for _, gp := range gaps {
		var next [][2]float64
		for _, iv := range free {
			if gp[1] <= iv[0] || gp[0] >= iv[1] {
				next = append(next, iv)
				continue
			}
			if gp[0] > iv[0] {
				next = append(next, [2]float64{iv[0], gp[0]})
			}
			if gp[1] < iv[1] {
				next = append(next, [2]float64{gp[1], iv[1]})
			}
		}
		free = next
	}
	seed := idx * 1000
	prev := Color(255)
	for _, iv := range free {
		L := (iv[1] - iv[0]) * rF
		var ws []float64
		sum := 0.0
		for L-sum > 6.5 {
			seed++
			w := 4.4 + 2.2*Hash01(seed, 7)
			if Hash01(seed, 8) < 0.08 {
				w = 8.2 + Hash01(seed, 9)
			}
			if L-sum-w < 4.4 {
				w = L - sum
			}
			ws = append(ws, w)
			sum += w
		}
		if L-sum >= 3.6 {
			ws = append(ws, L-sum)
			sum = L
		} else if len(ws) > 0 {
			ws[len(ws)-1] += L - sum
		}
		s := 0.0
		low := -1.0
		if flip {
			low = 1
		}
		for wi, w := range ws {
			seed++
			var sides []float64
			if wi == 0 {
				sides = append(sides, low)
			}
			if wi == len(ws)-1 {
				sides = append(sides, -low)
			}
			th := iv[0] + (s+w/2)/rF
			s += w
			facade := amsterdamCanalsFacades[int(Hash01(seed, 1)*float64(len(amsterdamCanalsFacades)))]
			for facade == prev {
				seed += 7
				facade = amsterdamCanalsFacades[int(Hash01(seed, 1)*float64(len(amsterdamCanalsFacades)))]
			}
			prev = facade
			trim := White
			if facade == White || facade == Sand {
				trim = Stone
			}
			roof := Slate
			switch r := Hash01(seed, 3); {
			case r < 0.3:
				roof = Charcoal
			case r < 0.42:
				roof = Terracotta
			}
			gable := int(Hash01(seed, 4) * 4)
			if w > 8 {
				gable = amsterdamCanalsCornice
			} else if Hash01(seed, 5) < 0.12 {
				gable = amsterdamCanalsCornice
			}
			storeys := 4
			if Hash01(seed, 6) < 0.45 {
				storeys = 5
			}
			if w > 8 {
				storeys = 4
			}
			lean := P((Hash01(seed, 10)-0.5)*0.024, 0.008+0.026*Hash01(seed, 11))
			h := amsterdamCanalsHouse{
				f: amsterdamCanalsFrame{th: th, r0: rF, flip: flip}, w: w, depth: depth,
				storeys: storeys, gable: gable, facade: facade, trim: trim, roof: roof,
				lean: lean, seed: seed, sides: sides,
			}
			h.build(g)
		}
	}
}

// elm is a quay tree: wood trunk and a lumpy foliage crown.
func (AmsterdamCanals) elm(g *Grid, c Vec2, s float64, seed int) {
	y := amsterdamCanalsQuay
	g.With(Wood, func() {
		g.Fill(Cylinder{Center: c, Radius: 0.55, Y0: y - 0.2, Y1: y + 5.5*s})
	})
	crown := UnionSolid{Ellipsoid{c.At(y + 7.4*s), V(2.5*s, 2.3*s, 2.5*s)}}
	for i := 0; i < 3; i++ {
		a := Hash01(seed, i) * 2 * math.Pi
		q := Polar(c, 1.6*s, a)
		crown = append(crown, Sphere(q.At(y+6.8*s+1.4*Hash01(seed, i+5)*s), 1.6*s))
	}
	g.With(Foliage, func() { g.Fill(crown) })
	g.Paint(Grass, Intersect(crown, Func{crown.Bounds(), func(p Vec3) bool {
		return p.Y > y+7.2*s && Hash01(int(p.X*3), int(p.Y*3), int(p.Z*3), seed) < 0.35
	}}))
}

// archBridge is a humped brick bridge with a stone-ringed arch across the band dA..dB.
func (AmsterdamCanals) archBridge(g *Grid, th, dA, dB float64) {
	rm := amsterdamCanalsRc + (dA+dB)/2
	f := amsterdamCanalsFrame{th: th, r0: rm}
	L := dB - dA
	hl := L/2 + 1.8
	const hwB = 2.7
	yd := func(z float64) float64 {
		t := z / hl
		return amsterdamCanalsQuay + 1.3*math.Max(0, 1-t*t)
	}
	ha, hb := L/2-1.1, 2.3
	inArch := func(p Vec3, k float64) bool {
		ez, ey := p.Z/(ha*k), (p.Y+0.4)/(hb*k)
		return ez*ez+ey*ey <= 1
	}
	g.With(Brick, func() {
		g.Fill(f.at(Func{Box{V(-hwB, 0.1, -hl), V(hwB, 3.0, hl)}, func(p Vec3) bool {
			return p.Y <= yd(p.Z) && !inArch(p, 1)
		}}))
		// Parapets.
		g.Fill(f.at(Func{Box{V(-hwB, 1.5, -hl+0.3), V(hwB, 4.2, hl-0.3)}, func(p Vec3) bool {
			return math.Abs(p.X) >= hwB-0.9 && p.Y >= yd(p.Z) && p.Y <= yd(p.Z)+1.0
		}}))
	})
	g.Paint(White, f.at(Func{Box{V(-hwB, 0.1, -hl), V(hwB, 4.2, hl)}, func(p Vec3) bool {
		return inArch(p, 1.28) || (math.Abs(p.X) >= hwB-0.9 && p.Y >= yd(p.Z)+0.6)
	}}))
	// Lanterns at the four corners.
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			x, z := sx*(hwB-0.45), sz*(hl-0.7)
			y := yd(z) + 1.0
			g.With(Charcoal, func() {
				g.Fill(f.at(Block(V(x-0.4, y-0.5, z-0.4), V(x+0.4, y+2.6, z+0.4))))
			})
			g.With(Gold, func() {
				g.Fill(f.at(Block(V(x-0.5, y+2.6, z-0.5), V(x+0.5, y+3.4, z+0.5))))
			})
		}
	}
}

// drawbridge is a white double-leaf wooden drawbridge with overhead balance beams.
func (AmsterdamCanals) drawbridge(g *Grid, th, dA, dB float64) {
	rm := amsterdamCanalsRc + (dA+dB)/2
	f := amsterdamCanalsFrame{th: th, r0: rm}
	L := dB - dA
	hl := L/2 + 1.8
	const hwB = 2.7
	yd := func(z float64) float64 {
		return amsterdamCanalsQuay + 1.1*amsterdamCanalsClamp01((hl-math.Abs(z))/2.6)
	}
	// Deck and railings.
	g.With(White, func() {
		g.Fill(f.at(Func{Box{V(-hwB, 1.0, -hl), V(hwB, 2.7, hl)}, func(p Vec3) bool {
			return p.Y <= yd(p.Z) && p.Y >= yd(p.Z)-0.7
		}}))
		g.Fill(f.at(Func{Box{V(-hwB, 1.5, -hl+0.4), V(hwB, 3.8, hl-0.4)}, func(p Vec3) bool {
			if math.Abs(p.X) < hwB-0.6 || p.Y < yd(p.Z) {
				return false
			}
			y := p.Y - yd(p.Z)
			return y >= 0.7 && y <= 1.1 || (y < 0.7 && math.Mod(p.Z+40, 1.6) < 0.6)
		}}))
	})
	// Piers in the water.
	for _, zp := range []float64{-L/2 + 1.6, -2.0, 2.0, L/2 - 1.6} {
		g.With(White, func() {
			g.Fill(f.at(Block(V(-hwB+0.2, 0.1, zp-0.5), V(hwB-0.2, yd(zp)-0.6, zp+0.5))))
		})
		g.Carve(f.at(Arch{V(0, 0.1, zp), 1.0, 0.6, 1.2, AxisZ}))
		g.Paint(Charcoal, f.at(Block(V(-hwB, 0.0, zp-0.6), V(hwB, 0.6, zp+0.6))))
	}
	// Portals and balance beams.
	deck := amsterdamCanalsQuay + 1.1
	for _, sz := range []float64{-1, 1} {
		zp := 2.2 * sz
		for _, sx := range []float64{-1, 1} {
			x := sx * (hwB - 0.35)
			g.With(White, func() {
				g.Fill(f.at(Block(V(x-0.4, deck, zp-0.4), V(x+0.4, deck+5.6, zp+0.4))))
				// Balance arm from behind the portal to the middle of the span.
				g.Fill(f.at(Capsule{V(x, deck+6.1, zp+sz*2.4), V(x, deck+5.5, -sz*0.7), 0.45}))
			})
			g.With(Charcoal, func() {
				g.Fill(f.at(Block(V(x-0.5, deck+5.0, zp+sz*1.9), V(x+0.5, deck+6.0, zp+sz*2.8))))
				g.Fill(f.at(Block(V(x-0.2, deck+0.2, -sz*0.9-0.2), V(x+0.2, deck+5.2, -sz*0.9+0.2))))
			})
		}
		g.With(White, func() {
			g.Fill(f.at(Block(V(-hwB+0.3, deck+5.3, zp-0.4), V(hwB-0.3, deck+6.1, zp+0.4))))
		})
		g.With(Gold, func() {
			g.Fill(f.at(Block(V(-0.45, deck+6.1, zp-0.45), V(0.45, deck+6.9, zp+0.45))))
		})
	}
	// Wooden deck planking between the railings.
	g.Paint(Wood, f.at(Func{Box{V(-hwB+0.6, 1.0, -hl), V(hwB-0.6, 2.7, hl)}, func(p Vec3) bool {
		return p.Y > yd(p.Z)-0.45
	}}))
}

func amsterdamCanalsClamp01(v float64) float64 { return math.Max(0, math.Min(1, v)) }

// houseboat is a moored boat house along the canal at plan angle th, offset d.
func (AmsterdamCanals) houseboat(g *Grid, th, d, L float64, seed int) {
	f := amsterdamCanalsFrame{th: th, r0: amsterdamCanalsRc + d}
	hull := []Color{Charcoal, Navy, Charcoal, DarkStone}[int(Hash01(seed, 1)*4)]
	cabin := []Color{Wood, Foliage, White, Navy, Red, Grass}[int(Hash01(seed, 2)*6)]
	g.With(hull, func() {
		g.Fill(f.at(Block(V(-L/2, -0.4, -1.7), V(L/2, 0.9, 1.7))))
	})
	cl := L/2 - 1.0
	g.With(cabin, func() {
		g.Fill(f.at(Block(V(-cl, 0.9, -1.35), V(cl, 3.0, 1.35))))
	})
	g.With(Charcoal, func() {
		g.Carve(f.at(Block(V(-cl+1.2, 0.95, -0.4), V(cl-1.2, 2.9, 0.4))))
	})
	// Windows along both sides.
	for x := -cl + 1.4; x < cl-0.8; x += 2.2 {
		g.Paint(White, f.at(Block(V(x-0.75, 1.5, -1.5), V(x+0.75, 2.6, 1.5))))
		g.Carve(f.at(Func{Box{V(x-0.4, 1.8, -1.5), V(x+0.4, 2.4, 1.5)}, func(p Vec3) bool {
			return math.Abs(p.Z) > 0.9
		}}))
	}
	g.Paint(Charcoal, f.at(Block(V(-cl+1.0, 1.0, -0.95), V(cl-1.0, 2.9, 0.95))))
	if Hash01(seed, 3) < 0.5 {
		g.With(White, func() {
			g.Fill(f.at(Block(V(-cl-0.3, 3.0, -1.6), V(cl+0.3, 3.4, 1.6))))
		})
		// Roof garden.
		for i := 0; i < 3; i++ {
			x := -cl + 1.2 + float64(i)*(2*cl-2.4)/2
			c := Foliage
			if Hash01(seed, i+10) < 0.35 {
				c = Red
			}
			g.With(c, func() { g.Fill(f.at(Sphere(V(x, 3.8, 0.4*float64(i%2*2-1)), 0.8))) })
		}
	} else {
		g.With(Slate, func() {
			g.Fill(f.at(Func{Box{V(-cl-0.3, 3.0, -1.7), V(cl+0.3, 4.7, 1.7)}, func(p Vec3) bool {
				return p.Y-3.0 <= 1.7-math.Abs(p.Z)
			}}))
		})
	}
	// Deck plants and a gangway to the quay.
	g.With(Grass, func() { g.Fill(f.at(Sphere(V(L/2-0.6, 1.4, 0), 0.7))) })
}

// tourBoat is a long low glass-roofed canal boat heading along local +X·dir.
func (AmsterdamCanals) tourBoat(g *Grid, th, d float64, seed int) {
	f := amsterdamCanalsFrame{th: th, r0: amsterdamCanalsRc + d}
	const L, W = 15.0, 1.6
	plan := func(x float64) float64 {
		s := math.Abs(x) / (L / 2)
		if s > 1 {
			return -1
		}
		if s < 0.7 {
			return W
		}
		t := (s - 0.7) / 0.3
		return W * math.Sqrt(math.Max(0, 1-t*t))
	}
	hull := Navy
	if Hash01(seed, 1) < 0.5 {
		hull = Charcoal
	}
	g.With(hull, func() {
		g.Fill(f.at(Func{Box{V(-L/2, -0.4, -W), V(L/2, 0.9, W)}, func(p Vec3) bool {
			return math.Abs(p.Z) <= plan(p.X)
		}}))
	})
	g.Paint(White, f.at(Func{Box{V(-L/2, 0.45, -W), V(L/2, 0.9, W)}, func(p Vec3) bool { return true }}))
	g.With(Glass, func() {
		g.Fill(f.at(Func{Box{V(-L/2+2, 0.9, -W), V(L/2-2.6, 2.4, W)}, func(p Vec3) bool {
			hz := plan(p.X) - 0.2
			e := (p.Y - 0.9) / 1.5
			return math.Abs(p.Z) <= hz*math.Sqrt(math.Max(0, 1-e*e*e*e))
		}}))
	})
	for x := -L/2 + 2.4; x < L/2-2.6; x += 1.8 {
		g.Paint(White, f.at(Block(V(x-0.2, 0.9, -W), V(x+0.2, 2.6, W))))
	}
}

// sloep is a small open motor boat with people aboard.
func (AmsterdamCanals) sloep(g *Grid, th, d float64, seed int) {
	f := amsterdamCanalsFrame{th: th, r0: amsterdamCanalsRc + d}
	g.With(Wood, func() {
		g.Fill(f.at(Func{Box{V(-3, -0.3, -1.2), V(3, 0.9, 1.2)}, func(p Vec3) bool {
			s := (p.X + 3) / 6
			return math.Abs(p.Z) <= 1.2*math.Min(1, (1-s)*3)
		}}))
	})
	g.With(White, func() {
		g.Carve(f.at(Block(V(-2.2, 0.5, -0.6), V(1.6, 1.0, 0.6))))
	})
	for i, x := range []float64{-1.8, -0.3, 1.0} {
		c := []Color{Red, White, Navy, Gold}[int(Hash01(seed, i)*4)]
		g.With(c, func() { g.Fill(f.at(Sphere(V(x, 1.3, 0.35*float64(i%2*2-1)), 0.55))) })
	}
}

// bikes parks a row of bicycles across the quay at radial offset d.
func (AmsterdamCanals) bikes(g *Grid, th, d float64, n int, seed int) {
	f := amsterdamCanalsFrame{th: th, r0: amsterdamCanalsRc + d}
	for i := 0; i < n; i++ {
		x := (float64(i) - float64(n-1)/2) * 1.25
		c := []Color{Charcoal, Navy, Red, Charcoal, White, Foliage}[int(Hash01(seed, i)*6)]
		g.With(c, func() {
			g.Fill(f.at(Block(V(x-0.25, amsterdamCanalsQuay, -0.85), V(x+0.25, amsterdamCanalsQuay+1.1, 0.85))))
		})
	}
}

// church is a Westerkerk-like brick church in the middle block with its tower on the
// front canal.
func (AmsterdamCanals) church(g *Grid) {
	f := amsterdamCanalsFrame{th: amsterdamCanalsChurch, r0: amsterdamCanalsRc + 6}
	B := amsterdamCanalsBase
	const nx, z0, z1, eave = 6.0, -8.6, 3.6, 13.0
	// Nave shell.
	g.Use(Brick)
	g.Fill(f.at(Block(V(-nx, B, z0), V(nx, eave, z1))))
	g.Carve(f.at(Block(V(-nx+1.2, B+0.6, z0+1.2), V(nx-1.2, eave+0.1, z1-1.2))))
	g.Paint(White, f.at(Block(V(-nx, eave-0.6, z0), V(nx, eave, z1))),
		f.at(Block(V(-nx, 5.6, z0), V(nx, 6.0, z1))))
	// Tall arched windows on the long sides and the rear end.
	for _, zc := range []float64{-6.5, -3.0, 0.5} {
		for _, sx := range []float64{-1, 1} {
			g.Paint(White, f.at(Arch{V(sx*nx, 3.4, zc), 1.5, 5.6, 1.0, AxisX}))
			g.Carve(f.at(Arch{V(sx*nx, 3.8, zc), 1.05, 5.4, 3.0, AxisX}))
		}
	}
	for _, x := range []float64{-3.2, 0, 3.2} {
		g.Paint(White, f.at(Arch{V(x, 3.4, z0), 1.5, 5.6, 1.0, AxisZ}))
		g.Carve(f.at(Arch{V(x, 3.8, z0), 1.05, 5.4, 3.0, AxisZ}))
	}
	g.Paint(Charcoal, f.at(Subtract(Block(V(-nx+0.7, B+0.6, z0+0.7), V(nx-0.7, eave, z1-0.7)),
		Block(V(-nx+1.2, B, z0+1.2), V(nx-1.2, eave+1, z1-1.2)))))
	// Steep main roof with the ridge along the nave, and two cross gables per side.
	pitch := func(x0, x1, za, zb, base, k float64, alongX bool, c Color) {
		half := (x1 - x0) / 2
		if alongX {
			half = (zb - za) / 2
		}
		rise := half * k
		tv := 1.2 * math.Sqrt(1+k*k)
		g.With(c, func() {
			g.Fill(f.at(Func{Box{V(x0, base-0.3, za), V(x1, base+rise, zb)}, func(p Vec3) bool {
				across := math.Abs(p.X - (x0+x1)/2)
				if alongX {
					across = math.Abs(p.Z - (za+zb)/2)
				}
				s := k * (half - across)
				y := p.Y - base
				return y <= s && y >= s-tv
			}}))
		})
	}
	pitch(-nx-0.6, nx+0.6, z0-0.4, z1, eave, 1.15, false, Slate)
	for _, zc := range []float64{-5.0, 0.0} {
		// Gable end walls rising above the eaves.
		g.With(Brick, func() {
			g.Fill(f.at(Func{Box{V(-nx, eave, zc-2.7), V(nx, eave+4.3, zc+2.7)}, func(p Vec3) bool {
				return math.Abs(p.X) >= nx-1.2 && p.Y-eave <= 1.6*(2.7-math.Abs(p.Z-zc))
			}}))
		})
		g.Paint(White, f.at(Func{Box{V(-nx, eave, zc-2.7), V(nx, eave+4.4, zc+2.7)}, func(p Vec3) bool {
			s := 1.6 * (2.7 - math.Abs(p.Z-zc))
			return math.Abs(p.X) >= nx-1.2 && p.Y-eave > s-0.9
		}}))
		pitch(-nx+0.4, nx-0.4, zc-2.7, zc+2.7, eave, 1.6, true, Slate)
		for _, sx := range []float64{-1, 1} {
			g.Carve(f.at(Func{Box{V(sx*nx-1.5, eave+0.6, zc-0.8), V(sx*nx+1.5, eave+2.2, zc+0.8)}, func(p Vec3) bool {
				return math.Hypot(p.Z-zc, p.Y-eave-1.4) <= 0.75
			}}))
		}
	}
	// Rear gable end.
	g.With(Brick, func() {
		g.Fill(f.at(Func{Box{V(-nx, eave, z0), V(nx, eave+7, z0+1.2)}, func(p Vec3) bool {
			return p.Y-eave <= 1.15*(nx-math.Abs(p.X))-0.3
		}}))
	})

	// Tower: brick shaft, stone belfry, two octagons and the imperial crown.
	tz := 6.75
	c := P(0, tz)
	const tw = 3.3
	g.Use(Brick)
	g.Fill(f.at(BlockAt(c, 2*tw, 2*tw, B, 22)))
	g.Carve(f.at(BlockAt(c, 2*tw-2.4, 2*tw-2.4, B+0.6, 22.1)))
	g.Paint(White, f.at(BlockAt(c, 2*tw, 2*tw, 8.0, 8.5)), f.at(BlockAt(c, 2*tw, 2*tw, 15.0, 15.5)),
		f.at(BlockAt(c, 2*tw, 2*tw, 21.2, 22)))
	g.Paint(White, f.at(Func{Box{V(-tw, B, tz-tw), V(tw, 22, tz+tw)}, func(p Vec3) bool {
		return math.Abs(p.X) > tw-1.0 && math.Abs(p.Z-tz) > tw-1.0 && math.Mod(p.Y, 1.6) < 0.8
	}}))
	// Door and windows.
	g.Paint(White, f.at(Arch{V(0, B+0.4, tz+tw), 1.6, 3.0, 1.0, AxisZ}))
	g.Carve(f.at(Arch{V(0, B+0.6, tz+tw), 1.1, 2.6, 3.0, AxisZ}))
	for _, y := range []float64{9.6, 16.2} {
		g.Paint(White, f.at(Arch{V(0, y-0.4, tz+tw), 1.4, 2.8, 1.0, AxisZ}), f.at(Arch{V(tw, y-0.4, tz), 1.4, 2.8, 1.0, AxisX}),
			f.at(Arch{V(-tw, y-0.4, tz), 1.4, 2.8, 1.0, AxisX}))
		g.Carve(f.at(Arch{V(0, y, tz+tw), 0.95, 2.6, 3.0, AxisZ}), f.at(Arch{V(tw, y, tz), 0.95, 2.6, 3.0, AxisX}),
			f.at(Arch{V(-tw, y, tz), 0.95, 2.6, 3.0, AxisX}))
	}
	g.Paint(Charcoal, f.at(Subtract(BlockAt(c, 2*tw-1.6, 2*tw-1.6, B+0.6, 21), BlockAt(c, 2*tw-2.4, 2*tw-2.4, B, 23))))
	// Balustrade.
	g.With(White, func() {
		g.Fill(f.at(Subtract(BlockAt(c, 2*tw+0.6, 2*tw+0.6, 22, 23.2), BlockAt(c, 2*tw-1.0, 2*tw-1.0, 21, 24))))
	})
	// Stone belfry with corner pinnacles.
	const bw = 2.4
	g.Use(Stone)
	g.Fill(f.at(BlockAt(c, 2*bw, 2*bw, 22, 28.5)))
	g.Carve(f.at(BlockAt(c, 2*bw-2, 2*bw-2, 22.6, 28.6)))
	g.Carve(f.at(Arch{V(0, 23.2, tz), 1.0, 3.0, 2*bw + 1, AxisZ}), f.at(Arch{V(0, 23.2, tz), 1.0, 3.0, 2*bw + 1, AxisX}))
	g.Paint(Charcoal, f.at(Subtract(BlockAt(c, 2*bw-1.2, 2*bw-1.2, 22.2, 28.5), BlockAt(c, 2*bw-2, 2*bw-2, 22, 29))))
	g.Paint(White, f.at(BlockAt(c, 2*bw, 2*bw, 28.0, 28.5)))
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			q := P(sx*(bw+0.1), tz+sz*(bw+0.1))
			g.With(Stone, func() { g.Fill(f.at(Cylinder{Center: q, Radius: 0.55, Y0: 26, Y1: 29.6})) })
			g.With(Gold, func() { g.Fill(f.at(Sphere(q.At(30.0), 0.6))) })
		}
	}
	// Octagons.
	oct := func(r, y0, y1 float64) Solid {
		poly := make([]Vec2, 8)
		for i := range poly {
			poly[i] = Polar(c, r, math.Pi/8+float64(i)*math.Pi/4)
		}
		return Prism{poly, y0, y1}
	}
	g.With(Stone, func() {
		g.Fill(f.at(Subtract(oct(2.2, 28.5, 33.5), oct(1.2, 28.6, 33.4))))
	})
	g.Carve(f.at(Arch{V(0, 29.4, tz), 0.7, 2.0, 5, AxisZ}), f.at(Arch{V(0, 29.4, tz), 0.7, 2.0, 5, AxisX}))
	g.Paint(Charcoal, f.at(Subtract(oct(1.8, 28.6, 33.4), oct(1.2, 28, 34))))
	g.With(White, func() { g.Fill(f.at(oct(2.5, 33.3, 33.9))) })
	g.With(Stone, func() {
		g.Fill(f.at(Subtract(oct(1.7, 33.9, 37.4), oct(0.8, 34, 37.3))))
	})
	// Clock dials.
	for _, d := range []Vec2{P(0, 1), P(0, -1), P(1, 0), P(-1, 0)} {
		q := c.Add(d.Scale(1.6))
		g.With(Navy, func() { g.Fill(f.at(Sphere(q.At(35.6), 0.9))) })
		g.With(Gold, func() { g.Fill(f.at(Sphere(q.Add(d.Scale(0.5)).At(35.6), 0.45))) })
	}
	g.With(Gold, func() { g.Fill(f.at(oct(2.0, 37.2, 37.9))) })
	// Imperial crown: blue bulb with gold ribs, orb and cross.
	g.With(Navy, func() {
		g.Fill(f.at(Ellipsoid{c.At(39.6), V(1.9, 2.2, 1.9)}))
	})
	g.Paint(Gold, f.at(Func{Box{V(-2.2, 37.4, tz-2.2), V(2.2, 42, tz+2.2)}, func(p Vec3) bool {
		a := math.Atan2(p.Z-tz, p.X) / (math.Pi / 4)
		return math.Abs(a-math.Round(a)) < 0.18 || p.Y < 38.2
	}}))
	g.With(Gold, func() {
		g.Fill(f.at(Sphere(c.At(42.3), 0.8)))
		g.Fill(f.at(Capsule{c.At(42.8), c.At(45.6), 0.42}))
		g.Fill(f.at(Capsule{V(-0.9, 44.6, tz), V(0.9, 44.6, tz), 0.42}))
	})
}

func (a AmsterdamCanals) Build(g *Grid) Order {
	T := amsterdamCanalsHalf
	sector := amsterdamCanalsSector
	// Water, quay walls, street plate with a closed skirt along the site edges.
	g.Use(Water)
	for _, c := range amsterdamCanalsCanals {
		g.Fill(sector(-T, T, c[0], c[1], -0.5, 0))
	}
	g.Use(DarkStone)
	for _, l := range amsterdamCanalsLand {
		g.Fill(sector(-T, T, l[0], l[1], 1.0, amsterdamCanalsQuay))
		g.Fill(sector(-T, -T+1.0/(amsterdamCanalsRc+l[0]), l[0], l[1], -0.5, amsterdamCanalsQuay),
			sector(T-1.0/(amsterdamCanalsRc+l[0]), T, l[0], l[1], -0.5, amsterdamCanalsQuay))
	}
	g.Fill(sector(-T, T, 34, 35, -0.5, amsterdamCanalsQuay), sector(-T, T, -31, -30, -0.5, amsterdamCanalsQuay))
	g.Use(Brick)
	for _, c := range amsterdamCanalsCanals {
		g.Fill(sector(-T, T, c[0]-1, c[0], -0.5, amsterdamCanalsQuay), sector(-T, T, c[1], c[1]+1, -0.5, amsterdamCanalsQuay))
	}
	// Gardens between the back-to-back rows of the middle block.
	g.Paint(Grass, sector(-T, T, 4.5, 7.5, 1.2, amsterdamCanalsQuay))

	// House rows.
	gap := func(th, half, d float64) [2]float64 {
		r := amsterdamCanalsRc + d
		return [2]float64{th - half/r, th + half/r}
	}
	church := [2]float64{amsterdamCanalsChurch - amsterdamCanalsChurchW, amsterdamCanalsChurch + amsterdamCanalsChurchW}
	a.row(g, 1, -23, 8, false, [][2]float64{gap(amsterdamCanalsStreet, 2.4, -23)})
	a.row(g, 2, -3, 7.5, true, [][2]float64{gap(amsterdamCanalsStreet, 2.4, -3), gap(amsterdamCanalsDraw, 2.4, -3), church})
	a.row(g, 3, 15.5, 8, false, [][2]float64{gap(amsterdamCanalsStreet, 2.4, 15.5), gap(amsterdamCanalsDraw, 2.4, 15.5), church})
	a.church(g)

	// Bridges.
	a.archBridge(g, amsterdamCanalsStreet, amsterdamCanalsCanals[0][0], amsterdamCanalsCanals[0][1])
	a.archBridge(g, amsterdamCanalsStreet, amsterdamCanalsCanals[1][0], amsterdamCanalsCanals[1][1])
	a.drawbridge(g, amsterdamCanalsDraw, amsterdamCanalsCanals[1][0], amsterdamCanalsCanals[1][1])

	// Small garden trees between the back-to-back rows.
	for i := 0; i < 9; i++ {
		th := -T + (float64(i)+0.5+0.3*Hash01(i, 77))*2*T/9
		if math.Abs(th-amsterdamCanalsChurch) < amsterdamCanalsChurchW || math.Abs(th-amsterdamCanalsStreet)*106 < 4 || math.Abs(th-amsterdamCanalsDraw)*106 < 4 {
			continue
		}
		a.elm(g, amsterdamCanalsAt(th, amsterdamCanalsRc+6), 0.55, 500+i)
	}

	// Elms along the quays, clear of bridges and the church tower.
	quays := []struct {
		d, step, s float64
	}{{-20.2, 8, 0.95}, {-5.9, 8.5, 1.0}, {18.4, 9, 1.0}, {32.6, 11, 0.9}}
	for qi, q := range quays {
		r := amsterdamCanalsRc + q.d
		for th := -T + 3/r; th < T-2/r; th += q.step / r {
			clear := func(c, half float64) bool { return math.Abs(th-c)*r > half }
			if !clear(amsterdamCanalsStreet, 4.5) || !clear(amsterdamCanalsDraw, 4.5) {
				continue
			}
			if qi == 2 && !clear(amsterdamCanalsChurch, 5) {
				continue
			}
			// The front quay keeps only its end trees, leaving the canal in view.
			if qi == 3 && math.Abs(th)*r < 26 {
				continue
			}
			a.elm(g, amsterdamCanalsAt(th, r), q.s, qi*100+int(th*1000))
		}
	}

	// Boats.
	u := func(x float64) float64 { return x / amsterdamCanalsRc }
	a.houseboat(g, u(-25), -9.6, 11, 1)
	a.houseboat(g, u(-7), -9.6, 10, 2)
	a.houseboat(g, u(27), -16.4, 10, 3)
	a.houseboat(g, u(25), 28.9, 11, 4)
	a.houseboat(g, u(3), 28.9, 10, 5)
	a.houseboat(g, u(-25), 28.9, 12, 6)
	a.tourBoat(g, u(3), -13.2, 1)
	a.tourBoat(g, u(-18), 23.8, 2)
	a.sloep(g, u(20), 24.5, 3)

	// Bicycles by the water and on the bridge approaches.
	a.bikes(g, amsterdamCanalsStreet+u(3.5), 17.8, 5, 1)
	a.bikes(g, amsterdamCanalsStreet-u(4), 17.8, 4, 2)
	a.bikes(g, amsterdamCanalsDraw+u(4), 33.0, 6, 3)
	a.bikes(g, u(24), 17.8, 4, 4)
	a.bikes(g, u(-30), 33.0, 5, 5)
	a.bikes(g, amsterdamCanalsStreet+u(4), -5.6, 5, 6)
	a.bikes(g, u(-15), -5.6, 4, 7)

	return func(p Vec3) float64 {
		th, _ := amsterdamCanalsPolar(p.Plan())
		return th
	}
}
