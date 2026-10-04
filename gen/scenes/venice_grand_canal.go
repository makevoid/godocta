package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// VeniceGrandCanal is a bend of the Grand Canal at the Rialto. Everything is authored
// in a straight "canal frame" (X runs along the canal, Z across it, the canal centered on
// Z = 0) and then bent along a parabolic centerline; the canal runs along world X and
// bows towards +Z at its ends. Continuous rows of Gothic and Renaissance palazzi rise
// straight out of a built water layer: pointed or round-arched loggias with balconies and
// tracery, a Doge's-palace-like showpiece, water gates with striped mooring poles, white
// cornices, hollow roofs with funnel chimneys and wooden altane. The far (San Polo, -Z)
// bank is taller, with houses behind and a fondamenta (Riva del Vin) with café parasols,
// a gondola station and a vaporetto at its floating stop; the near (+Z) bank is lower,
// with walled gardens behind so the hero view sees across the water. The white Rialto
// Bridge crosses in the middle with stepped ramps, two rows of arcaded shops and the
// gabled central portico. Side canals with fondamente and humped bridges cut through both
// rows; two campanili, the little church of San Giacomo with its clock, a market and a
// well stand on the campi; gondolas and water taxis are under way.
type VeniceGrandCanal struct{}

func init() { Register(VeniceGrandCanal{}) }

func (VeniceGrandCanal) Name() string  { return "venice_grand_canal" }
func (VeniceGrandCanal) Title() string { return "Venice Grand Canal" }

const (
	veniceGrandCanalPhi     = 0.0    // Site rotation: canal frame +X turns towards world +Z by this angle.
	veniceGrandCanalBendA   = 0.0035 // Centerline: z = a·u².
	veniceGrandCanalHalf    = 11.0   // Half width of the canal.
	veniceGrandCanalLand    = 1.25   // Top of the pavement.
	veniceGrandCanalEnd     = 50.0   // Half length of the site along the canal.
	veniceGrandCanalRim     = 30.5   // Distance of the site rim from the centerline (far bank).
	veniceGrandCanalRimNear = 26.0   // Rim on the near bank (Z > 0).
)

// veniceGrandCanalCurve maps canal-frame plan coordinates (u along, d across) to
// world plan coordinates, before the site rotation.
func veniceGrandCanalCurve(u, d float64) Vec2 {
	a := veniceGrandCanalBendA
	n := P(-2*a*u, 1).Scale(1 / math.Sqrt(1+4*a*a*u*u))
	return P(u, a*u*u).Add(n.Scale(d))
}

// veniceGrandCanalInverse maps a world plan point back to canal-frame (u, d).
func veniceGrandCanalInverse(p Vec2) (float64, float64) {
	a := veniceGrandCanalBendA
	q := p.Rotate(-veniceGrandCanalPhi)
	u := q.X
	for i := 0; i < 5; i++ {
		f := u - q.X + 2*a*u*(a*u*u-q.Y)
		df := 1 + 2*a*(3*a*u*u-q.Y)
		u -= f / df
	}
	c := P(u, a*u*u)
	d := q.Sub(c).Length()
	if q.Y < a*u*u {
		d = -d
	}
	return u, d
}

// veniceGrandCanalPlaced is a solid authored in a local frame: local X runs along a
// facade, local Z points outward. The local origin sits at canal-frame (uc, dc); flip
// turns the frame half round. The result is bent along the canal and rotated to the site.
type veniceGrandCanalPlaced struct {
	s      Solid
	uc, dc float64
	flip   bool
	box    Box
}

func (v veniceGrandCanalPlaced) Contains(p Vec3) bool {
	u, d := veniceGrandCanalInverse(p.Plan())
	x, z := u-v.uc, d-v.dc
	if v.flip {
		x, z = -x, -z
	}
	return v.s.Contains(V(x, p.Y, z))
}

func (v veniceGrandCanalPlaced) Bounds() Box { return v.box }

// veniceGrandCanalFrame places local solids; see veniceGrandCanalPlaced.
type veniceGrandCanalFrame struct {
	uc, dc float64
	flip   bool
}

func (f veniceGrandCanalFrame) at(s Solid) Solid {
	lb := s.Bounds()
	x0, x1, z0, z1 := lb.Min.X, lb.Max.X, lb.Min.Z, lb.Max.Z
	if f.flip {
		x0, x1, z0, z1 = -x1, -x0, -z1, -z0
	}
	u0, u1, d0, d1 := f.uc+x0, f.uc+x1, f.dc+z0, f.dc+z1
	b := EmptyBox()
	steps := int(math.Ceil((u1-u0)/2)) + 1
	for i := 0; i <= steps; i++ {
		u := u0 + (u1-u0)*float64(i)/float64(steps)
		for _, d := range []float64{d0, d1} {
			q := veniceGrandCanalCurve(u, d).Rotate(veniceGrandCanalPhi)
			b = b.Union(Box{q.At(lb.Min.Y), q.At(lb.Max.Y)})
		}
	}
	return veniceGrandCanalPlaced{s, f.uc, f.dc, f.flip, b.Grow(0.6)}
}

// veniceGrandCanalCanal is the frame of the canal itself.
var veniceGrandCanalCanal = veniceGrandCanalFrame{}

// veniceGrandCanalPointed is a pointed (equilateral) arch opening across local X at x,
// passing through local Z from z0 to z1.
func veniceGrandCanalPointed(x, hw, sill, spring, z0, z1 float64) Solid {
	apex := spring + math.Sqrt(3)*hw
	return Func{Box{V(x-hw, sill, z0), V(x+hw, apex, z1)}, func(p Vec3) bool {
		ax := math.Abs(p.X - x)
		if ax > hw {
			return false
		}
		rise := p.Y - spring
		if rise <= 0 {
			return true
		}
		d := ax + hw
		return d*d+rise*rise <= 4*hw*hw
	}}
}

// veniceGrandCanalOpening is a pointed or round arch through local Z.
func veniceGrandCanalOpening(x, hw, sill, spring, z0, z1 float64, pointed bool) Solid {
	if pointed {
		return veniceGrandCanalPointed(x, hw, sill, spring, z0, z1)
	}
	return Arch{V(x, sill, (z0+z1)/2), hw, spring - sill, z1 - z0, AxisZ}
}

// veniceGrandCanalHip is a hollow low-pitched hip roof shell over a rectangle.
func veniceGrandCanalHip(x0, x1, z0, z1, base, pitch, thick float64) Solid {
	rise := 0.5 + pitch*math.Min(x1-x0, z1-z0)/2
	return Func{Box{V(x0, base, z0), V(x1, base+rise, z1)}, func(p Vec3) bool {
		m := math.Min(math.Min(p.X-x0, x1-p.X), math.Min(p.Z-z0, z1-p.Z))
		top := base + 0.5 + pitch*m
		return p.Y <= top && p.Y >= top-thick
	}}
}

// veniceGrandCanalHipTop is the height of that roof's surface at (x, z).
func veniceGrandCanalHipTop(x0, x1, z0, z1, base, pitch, x, z float64) float64 {
	m := math.Min(math.Min(x-x0, x1-x), math.Min(z-z0, z1-z))
	return base + 0.5 + pitch*m
}

// Palazzo styles.
const (
	veniceGrandCanalGothic = iota
	veniceGrandCanalRenaissance
	veniceGrandCanalPlain
)

// veniceGrandCanalLot is one canal-front lot: a palazzo with its facade on local
// z = zf facing the canal (+Z), and a lower house behind it reaching the site rim.
type veniceGrandCanalLot struct {
	f        veniceGrandCanalFrame
	w        float64
	zf       float64 // Facade plane: 0 on the water, negative behind a fondamenta.
	style    int
	storeys  int
	wall     Color
	rearWall Color
	seed     int
	paline   Color // Stripe color of the mooring poles; 0 for none.
	altana   bool
	roof     Color
	showcase bool // A Doge's-palace-like facade with diaper pattern and full loggias.
	near     bool // Near bank: lower, with a garden behind.
	endL     bool // Local -X side wall is exposed.
	endR     bool // Local +X side wall is exposed.
}

func (VeniceGrandCanal) chimney(g *Grid, at func(Solid) Solid, x, z, roof float64, funnel Color) {
	g.With(Brick, func() {
		g.Fill(at(Block(V(x-0.5, roof-1.2, z-0.5), V(x+0.5, roof+1.8, z+0.5))))
	})
	g.With(funnel, func() {
		g.Fill(at(Frustum{P(x, z), 0.5, 1.1, roof + 1.8, roof + 3.1}))
		g.Fill(at(Cylinder{Center: P(x, z), Radius: 1.15, Y0: roof + 3.1, Y1: roof + 3.5}))
	})
}

func (VeniceGrandCanal) altana(g *Grid, at func(Solid) Solid, cx, cz, roof float64) {
	const hx, hz = 2.3, 1.8
	deck := roof + 1.6
	g.With(Wood, func() {
		for _, sx := range []float64{-1, 1} {
			for _, sz := range []float64{-1, 1} {
				x, z := cx+sx*(hx-0.45), cz+sz*(hz-0.45)
				g.Fill(at(Block(V(x-0.45, roof-1.5, z-0.45), V(x+0.45, deck+1.3, z+0.45))))
			}
		}
		g.Fill(at(Block(V(cx-hx, deck-0.45, cz-hz), V(cx+hx, deck, cz+hz))))
		rail := Subtract(Block(V(cx-hx, deck+0.9, cz-hz), V(cx+hx, deck+1.3, cz+hz)),
			Block(V(cx-hx+0.6, deck+0.5, cz-hz+0.6), V(cx+hx-0.6, deck+1.5, cz+hz-0.6)))
		g.Fill(at(rail))
	})
}

// pole is a striped mooring pole (paleta) standing in the water at local (x, z).
func (VeniceGrandCanal) pole(g *Grid, at func(Solid) Solid, x, z, top float64, stripe Color, seed int) {
	g.With(White, func() {
		g.Fill(at(Cylinder{Center: P(x, z), Radius: 0.5, Y0: -0.4, Y1: top}))
	})
	twist := Hash01(seed, 3) * 1.6
	g.Paint(stripe, at(Func{Box{V(x-0.6, -0.4, z-0.6), V(x+0.6, top, z+0.6)}, func(p Vec3) bool {
		return math.Mod(p.Y+twist+0.4*(p.X-x)+10, 1.6) < 0.8
	}}))
	g.With(DarkStone, func() {
		g.Fill(at(Cylinder{Center: P(x, z), Radius: 0.5, Y0: -0.4, Y1: 0.5}))
	})
	knob := Gold
	if Hash01(seed, 4) < 0.5 {
		knob = stripe
	}
	g.With(knob, func() {
		g.Fill(at(Sphere(V(x, top+0.2, z), 0.6)))
	})
}

// gondola is a black gondola along local X centered at (x, z), bow towards +X·dir.
func (VeniceGrandCanal) gondola(g *Grid, at func(Solid) Solid, x, z, dir float64, seed int) {
	const L = 8.6
	hull := Func{Box{V(x-L/2, 0.3, z-0.9), V(x+L/2, 2.4, z+0.9)}, func(p Vec3) bool {
		s := (p.X - x) / (L / 2)
		if math.Abs(s) > 1 {
			return false
		}
		hw := math.Max(0.45, 0.85*math.Sqrt(1-s*s))
		// Asymmetric gondola: the hull leans slightly to one side.
		off := 0.12 * (1 - s*s)
		if math.Abs(p.Z-z-off) > hw {
			return false
		}
		e := math.Max(0, (math.Abs(s)-0.55)/0.45)
		top := 0.95 + 1.3*e*e
		if math.Abs(s) > 0.6 {
			return p.Y <= top && p.Y >= top-0.9
		}
		return p.Y <= top
	}}
	g.With(Charcoal, func() { g.Fill(at(hull)) })
	// Ferro on the bow and red seat cushions.
	bx := x + dir*(L/2-0.2)
	g.With(White, func() {
		g.Fill(at(Block(V(bx-0.4, 1.7, z-0.35), V(bx+0.4, 2.7, z+0.35))))
	})
	if Hash01(seed, 7) < 0.6 {
		g.With(Red, func() {
			g.Fill(at(Block(V(x-1.4, 0.9, z-0.5), V(x+0.4, 1.35, z+0.5))))
		})
	}
	if Hash01(seed, 8) < 0.5 {
		// Gondolier standing on the stern in a striped shirt.
		sx := x - dir*(L/2-1.6)
		g.With(Navy, func() { g.Fill(at(Block(V(sx-0.4, 1.0, z-0.4), V(sx+0.4, 2.2, z+0.4)))) })
		g.With(White, func() { g.Fill(at(Block(V(sx-0.4, 2.2, z-0.4), V(sx+0.4, 3.0, z+0.4)))) })
		g.Paint(Navy, at(Block(V(sx-0.5, 2.5, z-0.5), V(sx+0.5, 2.6, z+0.5))))
		g.With(Sand, func() { g.Fill(at(Block(V(sx-0.35, 3.0, z-0.35), V(sx+0.35, 3.6, z+0.35)))) })
		g.With(Wood, func() {
			g.Fill(at(Capsule{V(sx+0.8*dir, 3.0, z+0.6), V(sx+2.6*dir, 0.3, z+1.2), 0.42}))
		})
	}
}

// palazzo builds one lot.
func (v VeniceGrandCanal) palazzo(g *Grid, l veniceGrandCanalLot) {
	at := l.f.at
	hw := l.w / 2
	zf := l.zf
	depth := 11.0
	if zf < 0 {
		depth = 10
	}
	zb := zf - depth
	pointed := l.style == veniceGrandCanalGothic
	trim := White
	if l.style == veniceGrandCanalPlain && l.wall != Stone {
		trim = Stone
	}
	if l.style == veniceGrandCanalRenaissance && l.wall == White {
		trim = Stone
	}
	base := -0.45
	if zf < 0 {
		base = veniceGrandCanalLand - 0.5
	}
	ground := veniceGrandCanalLand
	pn := 5.0 // Piano nobile floor.
	floors := []float64{pn, pn + 4.4, pn + 8.2, pn + 11.0}
	eave := floors[l.storeys-1]
	if l.storeys < 3 {
		eave = pn + 4.6
	}

	// Shell, rising from the water.
	g.Use(l.wall)
	g.Fill(at(Block(V(-hw, base, zb), V(hw, eave, zf))))
	g.Carve(at(Block(V(-hw+1, base, zb+1), V(hw-1, eave+0.1, zf-1))))
	g.With(DarkStone, func() {
		g.Fill(at(Block(V(-hw+1, ground-0.45, zb+1), V(hw-1, ground, zf-1))))
	})
	// Floors seen through the loggias.
	g.With(Wood, func() {
		for _, y := range floors[:l.storeys-1] {
			g.Fill(at(Block(V(-hw+1, y-0.45, zb+1), V(hw-1, y, zf-1))))
		}
	})
	// Istrian stone plinth and the green-black tide line.
	g.Paint(White, at(Block(V(-hw, base, zf-0.6), V(hw, ground+0.5, zf+0.1))))
	if zf == 0 {
		g.Paint(DarkStone, at(Block(V(-hw, base, zf-0.6), V(hw, 0.5, zf+0.1))))
	}

	// Renaissance pilasters and corner quoins, painted flush.
	if l.style == veniceGrandCanalRenaissance {
		for _, x := range []float64{-hw + 0.5, hw - 0.5} {
			g.Paint(trim, at(Block(V(x-0.5, ground, zf-0.6), V(x+0.5, eave, zf+0.1))))
		}
	} else {
		for _, x := range []float64{-hw + 0.4, hw - 0.4} {
			for y := ground + 0.5; y < eave-1; y += 1.6 {
				g.Paint(trim, at(Block(V(x-0.45, y, zf-0.6), V(x+0.45, y+0.8, zf+0.1))))
			}
		}
	}

	// String courses at each floor.
	g.With(trim, func() {
		for i, y := range floors {
			if y >= eave-0.5 || i > 2 {
				break
			}
			g.Fill(at(Block(V(-hw, y-0.45, zf-0.5), V(hw, y, zf+0.45))))
		}
	})

	zo0, zo1 := zf-1.3, zf+0.7
	opening := func(x, ohw, sill, spring float64, frame bool) {
		if frame {
			g.Paint(trim, at(veniceGrandCanalOpening(x, ohw+0.5, sill-0.3, spring, zf-0.6, zf+0.1, pointed)))
		}
		g.Carve(at(veniceGrandCanalOpening(x, ohw, sill, spring, zo0, zo1, pointed)))
	}
	window := func(x, ohw, y0, y1 float64) {
		g.Paint(trim, at(Block(V(x-ohw-0.45, y0-0.45, zf-0.6), V(x+ohw+0.45, y1+0.45, zf+0.1))))
		g.Carve(at(Block(V(x-ohw, y0, zo0), V(x+ohw, y1, zo1))))
	}
	balcony := func(x0, x1, y float64) {
		g.With(trim, func() {
			g.Fill(at(Block(V(x0, y-0.45, zf), V(x1, y, zf+1.1))))
			g.Fill(at(Block(V(x0, y, zf+0.6), V(x1, y+0.9, zf+1.1))))
			g.Fill(at(Block(V(x0, y, zf), V(x0+0.5, y+0.9, zf+1.1))))
			g.Fill(at(Block(V(x1-0.5, y, zf), V(x1, y+0.9, zf+1.1))))
		})
		for x := x0 + 1.0; x < x1-0.8; x += 1.4 {
			g.Carve(at(Block(V(x, y+0.3, zf+0.5), V(x+0.5, y+0.6, zf+1.2))))
		}
	}

	// Water gate (or street door behind a fondamenta) and ground-floor windows.
	gateHW := 1.4
	gateSill := 0.3
	if zf < 0 {
		gateHW, gateSill = 1.0, ground
	}
	opening(0, gateHW, gateSill, ground+2.0, true)
	if Hash01(l.seed, 1) < 0.5 {
		g.With(Wood, func() {
			g.Fill(at(Block(V(-gateHW, gateSill, zf-1.6), V(gateHW, ground+3.6, zf-1.0))))
		})
	}
	if hw > 5 {
		for _, s := range []float64{-1, 1} {
			window(s*(hw-2.1), 0.55, ground+1.4, ground+2.8)
		}
	}

	// Noble floors: a central polifora with a balcony, single windows to the sides.
	bays := int(math.Floor((l.w - 6.0) / 2.3))
	if bays > 5 {
		bays = 5
	}
	if l.showcase {
		bays = int(math.Floor((l.w - 1.6) / 2.3))
		// Lozenge diaper pattern above the arcades.
		g.Paint(White, at(Func{Box{V(-hw, floors[0]+0.5, zf-0.6), V(hw, eave, zf+0.1)}, func(p Vec3) bool {
			a := math.Mod(p.X+p.Y+100, 3.0)
			b := math.Mod(p.X-p.Y+100, 3.0)
			return a < 0.8 || b < 0.8
		}}))
	}
	if bays < 2 {
		bays = 2
	}
	for i, y := range floors {
		top := eave
		if i+1 < len(floors) {
			top = floors[i+1]
		}
		if y >= eave-0.5 {
			break
		}
		h := math.Min(top, eave) - y
		if i >= 2 || l.style == veniceGrandCanalPlain && i == 1 {
			// Upper floor: small windows.
			n := int(math.Floor(l.w / 3.0))
			for k := 0; k < n; k++ {
				x := -hw + l.w*(float64(k)+0.5)/float64(n)
				window(x, 0.55, y+0.7, y+math.Min(2.2, h-0.9))
			}
			continue
		}
		pitch := 2.3
		ohw := 0.65
		if i == 1 {
			ohw = 0.6
		}
		spring := y + h - 1.2 - math.Sqrt(3)*ohw*0.9
		if !pointed {
			spring = y + h - 1.1 - ohw
		}
		if l.style == veniceGrandCanalPlain {
			for k := 0; k < bays; k++ {
				x := (float64(k) - float64(bays-1)/2) * pitch
				opening(x, ohw, y+0.3, spring, true)
			}
		} else {
			span := float64(bays-1)*pitch + 2*ohw
			g.Paint(trim, at(veniceGrandCanalOpening(0, span/2+0.5, y-0.1, spring, zf-0.6, zf+0.1, pointed)))
			g.Paint(l.wall, at(Block(V(-span/2-0.6, y-0.1, zf-0.6), V(span/2+0.6, spring, zf+0.1))))
			for k := 0; k < bays; k++ {
				x := (float64(k) - float64(bays-1)/2) * pitch
				opening(x, ohw, y, spring, true)
			}
			if pointed && i == 0 {
				// Tracery: round openings between the arch heads.
				for k := 0; k <= bays; k++ {
					x := (float64(k) - float64(bays)/2) * pitch
					cy := spring + math.Sqrt(3)*ohw + 0.3
					g.Carve(at(Func{Box{V(x-0.7, cy-0.7, zo0), V(x+0.7, cy+0.7, zo1)}, func(p Vec3) bool {
						return math.Hypot(p.X-x, p.Y-cy) <= 0.6
					}}))
				}
			}
			if i == 0 || Hash01(l.seed, 20+i) < 0.5 {
				balcony(-span/2-0.4, span/2+0.4, y)
			}
		}
		if hw > 5.5 && !l.showcase {
			for _, s := range []float64{-1, 1} {
				x := s * (hw - 1.9)
				opening(x, 0.6, y+0.3, spring, true)
				if i == 0 && l.style != veniceGrandCanalPlain {
					balcony(x-1.0, x+1.0, y)
				}
			}
		}
	}

	// Cornice and Gothic crenellation.
	g.With(trim, func() {
		g.Fill(at(Subtract(Block(V(-hw, eave-0.5, zb), V(hw, eave, zf+0.6)),
			Block(V(-hw+1, eave-1, zb+1), V(hw-1, eave+1, zf-1)))))
		if pointed && Hash01(l.seed, 5) < 0.6 || l.showcase {
			for x := -hw + 0.5; x <= hw-0.4; x += 1.8 {
				g.Fill(at(Block(V(x-0.45, eave, zf-0.5), V(x+0.45, eave+1.2, zf+0.4))))
				if l.showcase {
					g.With(Gold, func() { g.Fill(at(Block(V(x-0.4, eave+1.2, zf-0.4), V(x+0.4, eave+1.7, zf+0.3)))) })
				}
			}
		}
	})

	// Roof.
	x0, x1, z0, z1 := -hw-0.4, hw+0.4, zb-0.4, zf+0.5
	pitch := 0.5
	g.With(l.roof, func() { g.Fill(at(veniceGrandCanalHip(x0, x1, z0, z1, eave, pitch, 1.1))) })
	roofAt := func(x, z float64) float64 { return veniceGrandCanalHipTop(x0, x1, z0, z1, eave, pitch, x, z) }
	funnel := White
	if Hash01(l.seed, 9) < 0.4 {
		funnel = Terracotta
	}
	for k := 0; k < 2; k++ {
		x := (hw - 1.6) * (2*Hash01(l.seed, 10+k) - 1)
		z := zb + 1.6 + (depth-3.2)*Hash01(l.seed, 12+k)
		v.chimney(g, at, x, z, roofAt(x, z), funnel)
	}
	if l.near && l.altana {
		cx := (hw - 3) * (2*Hash01(l.seed, 33) - 1)
		cz := zb + 3
		v.altana(g, at, cx, cz, roofAt(cx, cz))
	}

	// Side walls left exposed at the end of a row get windows with green shutters.
	shuttered := func(xs, za, zb2, y0, y1 float64) {
		n := int(math.Floor((zb2 - za) / 3.0))
		for y := y0; y+2 < y1; y += 3.4 {
			for k := 0; k < n; k++ {
				z := za + (zb2-za)*(float64(k)+0.5)/float64(n)
				g.Paint(Foliage, at(Block(V(xs*hw-0.4, y, z-1.15), V(xs*hw+0.2, y+1.8, z+1.15))))
				g.Carve(at(Block(V(xs*hw-1.3, y, z-0.5), V(xs*hw+0.5, y+1.8, z+0.5))))
			}
		}
	}
	for _, xs := range []float64{-1, 1} {
		if xs < 0 && l.endL || xs > 0 && l.endR {
			shuttered(xs, zb, zf, ground+1.6, eave)
		}
	}

	if l.near {
		// Back wall onto a walled garden.
		n := int(math.Floor(l.w / 3.0))
		for y := ground + 1.6; y+2 < eave; y += 3.6 {
			for k := 0; k < n; k++ {
				x := -hw + l.w*(float64(k)+0.5)/float64(n)
				g.Paint(Foliage, at(Block(V(x-1.15, y, zb-0.2), V(x+1.15, y+1.8, zb+0.4))))
				g.Carve(at(Block(V(x-0.5, y, zb-0.5), V(x+0.5, y+1.8, zb+1.2))))
				if Hash01(l.seed, k, int(y)) < 0.5 {
					g.With(Red, func() { g.Fill(at(Block(V(x-0.7, y-0.5, zb-0.7), V(x+0.7, y, zb)))) })
				}
			}
		}
		g.Carve(at(Arch{V(0, ground, zb+0.5), 0.9, 1.6, 2.2, AxisZ}))
		rear := -(veniceGrandCanalRimNear - veniceGrandCanalHalf)
		L := veniceGrandCanalLand
		g.Paint(Grass, at(Block(V(-hw+0.5, L-0.15, rear), V(hw-0.5, L, zb-0.1))))
		g.With(Brick, func() {
			g.Fill(at(Block(V(-hw, L, rear), V(hw, L+2.6, rear+0.9))))
			for _, xs := range []float64{-1, 1} {
				g.Fill(at(Block(V(xs*hw-0.45, L, rear), V(xs*hw+0.45, L+2.6, zb))))
			}
		})
		g.With(White, func() {
			g.Fill(at(Block(V(-hw, L+2.6, rear-0.1), V(hw, L+3.0, rear+1.0))))
		})
		trees := 1
		if l.w > 11 {
			trees = 2
		}
		for k := 0; k < trees; k++ {
			x := -hw + l.w*(float64(k)+0.5)/float64(trees) + (Hash01(l.seed, 50+k)-0.5)*1.5
			z := (rear+zb)/2 + 0.2
			r := 1.9 + 0.5*Hash01(l.seed, 52+k)
			g.With(Wood, func() { g.Fill(at(Block(V(x-0.45, L, z-0.45), V(x+0.45, L+3.6, z+0.45)))) })
			g.With(Foliage, func() { g.FillVolume(at(Ellipsoid{V(x, L+4.6, z), V(r, r*0.85, r)})) })
		}
	} else if rear := -(veniceGrandCanalRim - veniceGrandCanalHalf); rear < zb-3 {
		// The house behind.
		rh := 8.0 + 3*Hash01(l.seed, 30)
		if rh > eave-1.5 {
			rh = eave - 1.5
		}
		g.Use(l.rearWall)
		g.Fill(at(Block(V(-hw, ground-0.5, rear), V(hw, rh, zb))))
		g.Carve(at(Block(V(-hw+1, ground, rear+1), V(hw-1, rh+0.1, zb-0.5))))
		g.Paint(Stone, at(Block(V(-hw, ground-0.5, rear-0.1), V(hw, ground+0.6, zb))))
		n := int(math.Floor(l.w / 3.0))
		for y := ground + 1.5; y+2 < rh; y += 3.2 {
			for k := 0; k < n; k++ {
				x := -hw + l.w*(float64(k)+0.5)/float64(n)
				g.Paint(Foliage, at(Block(V(x-1.15, y, rear-0.2), V(x+1.15, y+1.8, rear+0.4))))
				g.Carve(at(Block(V(x-0.5, y, rear-0.5), V(x+0.5, y+1.8, rear+1.2))))
			}
		}
		for _, xs := range []float64{-1, 1} {
			if xs < 0 && l.endL || xs > 0 && l.endR {
				shuttered(xs, rear, zb, ground+1.5, rh)
			}
		}
		rx0, rx1, rz0, rz1 := -hw-0.3, hw+0.3, rear-0.4, zb+0.3
		rr := Terracotta
		if l.roof == Terracotta && Hash01(l.seed, 34) < 0.4 {
			rr = Brick
		}
		g.With(rr, func() { g.Fill(at(veniceGrandCanalHip(rx0, rx1, rz0, rz1, rh, 0.5, 1.1))) })
		x := (hw - 1.5) * (2*Hash01(l.seed, 31) - 1)
		z := rear + 1.5 + (zb-rear-3)*Hash01(l.seed, 32)
		v.chimney(g, at, x, z, veniceGrandCanalHipTop(rx0, rx1, rz0, rz1, rh, 0.5, x, z), funnel)
		if l.altana {
			cx := (hw - 3) * (2*Hash01(l.seed, 33) - 1)
			cz := (rear + zb) / 2
			v.altana(g, at, cx, cz, veniceGrandCanalHipTop(rx0, rx1, rz0, rz1, rh, 0.5, cx, cz))
		}
	}

	// Mooring poles and a moored gondola in front of the water gate.
	if l.paline != 0 && zf == 0 {
		for _, s := range []float64{-1, 1} {
			v.pole(g, at, s*2.7, 1.7, 6.0+Hash01(l.seed, 40)*1.5, l.paline, l.seed+int(s))
		}
		if Hash01(l.seed, 41) < 0.7 {
			v.gondola(g, at, 0.3, 2.9, 1, l.seed)
		}
	}
}

// lots lays a row of lots along a stretch of one bank, from u0 to u1. Rows on the near
// bank (side > 0) are lower and have walled gardens behind instead of houses.
func (v VeniceGrandCanal) lots(g *Grid, side float64, u0, u1, zf float64, seed int) {
	gothic := []Color{Terracotta, Brick, Sand, Red, Terracotta, Sand}
	ren := []Color{White, Stone, Sand, White}
	plain := []Color{Terracotta, Sand, Stone, Brick, Terracotta}
	rears := []Color{Sand, Terracotta, Stone, Brick, Terracotta, Sand}
	stripes := []Color{Red, Navy, Red, Navy, Wood}
	n := int(math.Round((u1 - u0) / 10.5))
	if n < 1 {
		n = 1
	}
	// Uneven widths that still fill the stretch exactly.
	weights := make([]float64, n)
	total := 0.0
	for i := range weights {
		weights[i] = 0.8 + 0.5*Hash01(seed, i, 1)
		total += weights[i]
	}
	near := side > 0
	u := u0
	for i := 0; i < n; i++ {
		w := (u1 - u0) * weights[i] / total
		h := func(k int) float64 { return Hash01(seed, i, k) }
		style := veniceGrandCanalGothic
		if r := h(2); r < 0.3 {
			style = veniceGrandCanalRenaissance
		} else if r > 0.78 {
			style = veniceGrandCanalPlain
		}
		var wall Color
		switch style {
		case veniceGrandCanalGothic:
			wall = gothic[int(h(3)*float64(len(gothic)))]
		case veniceGrandCanalRenaissance:
			wall = ren[int(h(3)*float64(len(ren)))]
		default:
			wall = plain[int(h(3)*float64(len(plain)))]
		}
		storeys := 3
		if near {
			if h(4) < 0.4 {
				storeys = 2
			}
		} else if h(4) < 0.55 {
			storeys = 4
		}
		var stripe Color
		if h(5) < 0.65 {
			stripe = stripes[int(h(6)*float64(len(stripes)))]
		}
		roof := Terracotta
		if h(9) < 0.3 || wall == Terracotta && h(9) < 0.8 {
			roof = Brick
		}
		showcase := seed == 13 && i == 1
		if showcase {
			style, wall, storeys = veniceGrandCanalGothic, Terracotta, 4
		}
		uc := u + w/2
		f := veniceGrandCanalFrame{uc, -veniceGrandCanalHalf, false}
		lo, hi := i == 0, i == n-1
		if near {
			f = veniceGrandCanalFrame{uc, veniceGrandCanalHalf, true}
			lo, hi = hi, lo
		}
		v.palazzo(g, veniceGrandCanalLot{
			f: f, w: w, zf: zf, style: style, storeys: storeys, wall: wall,
			rearWall: rears[int(h(7)*float64(len(rears)))], seed: seed*100 + i,
			paline: stripe, altana: h(8) < 0.45, roof: roof, near: near, endL: lo, endR: hi,
			showcase: showcase,
		})
		u += w
	}
}

// ground lays the water, quays and pavement.
func (v VeniceGrandCanal) ground(g *Grid) {
	at := veniceGrandCanalCanal.at
	E, R, H, L := veniceGrandCanalEnd, veniceGrandCanalRim, veniceGrandCanalHalf, veniceGrandCanalLand
	g.Use(Water).Fill(at(Block(V(-E, -0.45, -H-0.6), V(E, 0.05, H+0.6))))
	land := Union(Block(V(-E, L-0.5, -R), V(E, L, -H)), Block(V(-E, L-0.5, H), V(E, L, veniceGrandCanalRimNear)))
	g.Use(DarkStone).Fill(at(land))
	// Campo paving pattern: white Istrian lines.
	g.Paint(Stone, at(Func{Box{V(-E, L-0.5, -R), V(E, L, R)}, func(p Vec3) bool {
		return math.Mod(p.X+200, 4) < 0.7 || math.Mod(p.Z+200, 4) < 0.7
	}}))
	// Quay edges along the canal.
	g.Use(White).Fill(at(Union(
		Block(V(-E, -0.45, -H-0.6), V(E, L, -H+0.2)),
		Block(V(-E, -0.45, H-0.2), V(E, L, H+0.6)))))
}

// rio cuts a side canal of width 5 inland from the bank at u, with a fondamenta on
// each side and a humped bridge across.
func (v VeniceGrandCanal) rio(g *Grid, u, side float64, mouthBridge bool) {
	f := veniceGrandCanalFrame{u, -veniceGrandCanalHalf, false}
	if side > 0 {
		f = veniceGrandCanalFrame{u, veniceGrandCanalHalf, true}
	}
	at := f.at
	rear := -(veniceGrandCanalRim - veniceGrandCanalHalf)
	if side > 0 {
		rear = -(veniceGrandCanalRimNear - veniceGrandCanalHalf)
	}
	L := veniceGrandCanalLand
	g.Carve(at(Block(V(-5.2, -0.5, rear), V(5.2, 30, 0.6))))
	g.With(Water, func() { g.Fill(at(Block(V(-2.6, -0.45, rear), V(2.6, 0.05, 0.6)))) })
	g.With(DarkStone, func() {
		g.Fill(at(Block(V(-5.2, L-0.5, rear), V(-2.6, L, 0.6))), at(Block(V(2.6, L-0.5, rear), V(5.2, L, 0.6))))
	})
	g.With(White, func() {
		g.Fill(at(Block(V(-3.3, -0.45, rear), V(-2.6, L, 0.6))), at(Block(V(2.6, -0.45, rear), V(3.3, L, 0.6))))
	})
	bridge := func(zc float64) {
		// Stepped brick hump with white stone edges and parapets.
		deck := func(x float64) float64 {
			return L + math.Max(0, math.Floor((2.4-0.5*math.Max(0, math.Abs(x)-1.2))*2)/2)
		}
		body := Func{Box{V(-5.2, -0.4, zc-1.6), V(5.2, L+3.0, zc+1.6)}, func(p Vec3) bool {
			if p.Y > deck(p.X) {
				return false
			}
			// Round arch beneath.
			r := math.Hypot(p.X, p.Y-(-0.4))
			return math.Abs(p.X) > 2.6 || r > 2.9
		}}
		g.With(Brick, func() { g.Fill(at(body)) })
		g.Paint(White, at(Func{Box{V(-5.2, -0.4, zc-1.6), V(5.2, L+3.0, zc+1.6)}, func(p Vec3) bool {
			return math.Abs(p.Y-deck(p.X)) < 0.35 || math.Abs(math.Abs(p.Z-zc)-1.3) < 0.35
		}}))
		g.With(White, func() {
			for _, s := range []float64{-1, 1} {
				g.Fill(at(Func{Box{V(-5.2, L, zc+s*1.6-0.4), V(5.2, L+4.2, zc+s*1.6+0.4)}, func(p Vec3) bool {
					d := deck(p.X)
					return p.Y > d && p.Y <= d+1.0 && math.Abs(p.Z-(zc+s*1.3)) <= 0.35
				}}))
			}
		})
	}
	bridge(-12)
	if mouthBridge {
		bridge(-2)
	}
	// Moored gondola in the rio.
	v.gondola(g, func(s Solid) Solid { return at(RotateY(s, P(0, 0), math.Pi/2)) }, -7, -1.3, 1, int(u))
}

// rialto builds the bridge in canal coordinates at u = 0, spanning across Z.
func (v VeniceGrandCanal) rialto(g *Grid) {
	at := veniceGrandCanalCanal.at
	const (
		W     = 8.6  // Half width along the canal.
		crown = 8.6  // Walkway height at the portico.
		por   = 3.4  // Half length of the central portico.
		foot  = 17.8 // Where the ramps meet the campo.
	)
	R := 12.56
	yc := -6.06
	intra := func(z float64) float64 {
		if math.Abs(z) >= veniceGrandCanalHalf {
			return -1
		}
		return yc + math.Sqrt(R*R-z*z)
	}
	smooth := func(z float64) float64 { return crown - 0.5*math.Max(0, math.Abs(z)-por) }
	deck := func(z float64) float64 {
		return math.Max(veniceGrandCanalLand, math.Floor(smooth(z)*2)/2)
	}
	L := veniceGrandCanalLand

	// Arch barrel and abutments: hollow, with solid spandrel faces.
	g.Use(White).Fill(at(Func{Box{V(-W, -0.45, -foot-0.5), V(W, crown, foot+0.5)}, func(p Vec3) bool {
		d := deck(p.Z)
		if p.Y > d || math.Abs(p.Z) > foot+0.5 {
			return false
		}
		in := intra(p.Z)
		if math.Abs(p.Z) < veniceGrandCanalHalf && p.Y < in {
			return false
		}
		if math.Abs(p.X) > W-1.2 {
			return true
		}
		if p.Y > d-1.0 {
			return true
		}
		if math.Abs(p.Z) < veniceGrandCanalHalf {
			return p.Y < in+1.2
		}
		return p.Y < L
	}}))
	// Archivolt and a sloping cornice on both faces.
	g.Paint(Stone, at(Func{Box{V(-W, -0.45, -veniceGrandCanalHalf), V(W, crown, veniceGrandCanalHalf)}, func(p Vec3) bool {
		in := intra(p.Z)
		return math.Abs(p.X) > W-0.7 && p.Y >= in && p.Y < in+1.1
	}}))
	g.With(White, func() {
		g.Fill(at(Func{Box{V(-W-0.5, 0, -foot), V(W+0.5, crown, foot)}, func(p Vec3) bool {
			s := smooth(p.Z)
			return math.Abs(p.X) > W && p.Y <= s-0.5 && p.Y > s-1.1
		}}))
	})
	// Outer balustrades with openings.
	g.With(White, func() {
		g.Fill(at(Func{Box{V(-W, L, -foot), V(W, crown+1.4, foot)}, func(p Vec3) bool {
			d := deck(p.Z)
			return math.Abs(p.X) > W-0.6 && p.Y > d && p.Y <= d+1.1
		}}))
	})
	g.Carve(at(Func{Box{V(-W-0.2, L, -foot+1), V(W+0.2, crown+1.4, foot-1)}, func(p Vec3) bool {
		d := deck(p.Z)
		return math.Abs(p.X) > W-0.8 && p.Y > d+0.3 && p.Y < d+0.8 && math.Mod(p.Z+100.4, 1.8) < 0.9
	}}))

	// Two rows of shops on each ramp: arcaded blocks stepping down with the deck.
	for _, s := range []float64{-1, 1} {
		for k := 0; k < 6; k++ {
			za, zb := por+2.4*float64(k), por+2.4*float64(k+1)
			if s < 0 {
				za, zb = -zb, -za
			}
			zc := (za + zb) / 2
			b := deck(zc)
			top := b + 3.4
			for _, sx := range []float64{-1, 1} {
				xa, xb := 2.2, 5.4
				if sx < 0 {
					xa, xb = -5.4, -2.2
				}
				g.With(White, func() {
					g.Fill(at(Func{Box{V(xa, L, za), V(xb, top, zb)}, func(p Vec3) bool {
						return p.Y > deck(p.Z)-0.5
					}}))
				})
				g.Carve(at(Arch{V((xa+xb)/2, b+0.1, zc), 0.7, 1.6, 4.0, AxisX}))
				g.Carve(at(Block(V(xa+1.0, b+0.1, za+0.4), V(xb-1.0, top-1, zb-0.4))))
				g.With(Terracotta, func() {
					g.Fill(at(Gable{P(xa-0.3, za), P(xb+0.3, zb), top, AxisZ}))
				})
			}
		}
	}

	// Central portico with arches on all four sides and a pedimented gable.
	ptop := 14.6
	g.With(White, func() {
		g.Fill(at(Block(V(-W, crown-0.5, -por), V(W, ptop, por))))
		g.Fill(at(Block(V(-W-0.3, ptop-0.6, -por-0.3), V(W+0.3, ptop, por+0.3))))
	})
	g.Carve(at(Block(V(-W+1, crown, -por+1), V(W-1, ptop-1, por-1))))
	g.Carve(at(Arch{V(0, crown, 0), 2.4, 2.8, 2*W + 2, AxisX}))
	g.Carve(at(Arch{V(0, crown, 0), 1.8, 2.8, 2*por + 2, AxisZ}))
	for _, sx := range []float64{-1, 1} {
		g.Carve(at(Arch{V(sx*6.9, crown, 0), 1.1, 2.0, 2*por + 2, AxisZ}))
	}
	pediment := Gable{P(-W-0.3, -por-0.4), P(W+0.3, por+0.4), ptop, AxisX}
	g.With(White, func() { g.Fill(at(pediment)) })
	g.Paint(Terracotta, at(Func{pediment.Bounds(), func(p Vec3) bool {
		across := math.Abs(p.Z)
		return math.Abs(p.X) < W-0.3 && across >= por+0.4-(p.Y-ptop)-1.1
	}}))
	g.Paint(Stone, at(Func{pediment.Bounds(), func(p Vec3) bool {
		across := math.Abs(p.Z)
		return math.Abs(p.X) > W-0.3 && across >= por+0.4-(p.Y-ptop)-0.6
	}}))
}

// campanile is a brick bell tower with a white belfry and a green spire.
func (v VeniceGrandCanal) campanile(g *Grid, u, d, shaft float64, cone bool) {
	at := veniceGrandCanalFrame{u, d, false}.at
	const hw = 2.8
	L := veniceGrandCanalLand
	// Clear a small campo around the foot.
	g.Carve(at(Block(V(-hw-1.5, L, -hw-1.5), V(hw+1.5, 40, hw+1.5))))
	g.With(DarkStone, func() { g.Fill(at(Block(V(-hw-1.5, L-0.5, -hw-1.5), V(hw+1.5, L, hw+1.5)))) })
	g.Use(Brick).Fill(at(Block(V(-hw, L-0.5, -hw), V(hw, shaft, hw))))
	g.Carve(at(Block(V(-hw+1, L, -hw+1), V(hw-1, shaft+0.1, hw-1))))
	// Pilaster strips: recessed panels between corner piers.
	g.Paint(Terracotta, at(Subtract(Block(V(-hw, L+2, -hw), V(hw, shaft-1, hw)),
		Block(V(-hw+1.1, 0, -hw+1.1), V(hw-1.1, 40, hw-1.1)))))
	g.Paint(Brick, at(Union(
		Block(V(-hw, 0, -hw), V(-hw+1, 40, -hw+1)), Block(V(hw-1, 0, -hw), V(hw, 40, -hw+1)),
		Block(V(-hw, 0, hw-1), V(-hw+1, 40, hw)), Block(V(hw-1, 0, hw-1), V(hw, 40, hw)))))
	g.With(White, func() {
		g.Fill(at(Block(V(-hw, L-0.5, -hw), V(hw, L+1.5, hw))))
		g.Fill(at(Block(V(-hw-0.4, shaft-0.6, -hw-0.4), V(hw+0.4, shaft, hw+0.4))))
		// Belfry.
		bel := shaft + 4.6
		g.Fill(at(Block(V(-hw, shaft, -hw), V(hw, bel, hw))))
		g.Carve(at(Block(V(-hw+0.9, shaft, -hw+0.9), V(hw-0.9, bel-0.8, hw-0.9))))
		for _, x := range []float64{-1.1, 1.1} {
			g.Carve(at(Arch{V(x, shaft+0.6, 0), 0.65, 2.0, 7, AxisZ}))
			g.Carve(at(Arch{V(0, shaft+0.6, x), 0.65, 2.0, 7, AxisX}))
		}
		g.Fill(at(Block(V(-hw-0.4, bel-0.6, -hw-0.4), V(hw+0.4, bel, hw+0.4))))
		// Attic with round windows.
		g.Use(Brick).Fill(at(Block(V(-hw+0.3, bel, -hw+0.3), V(hw-0.3, bel+2.4, hw-0.3))))
		g.Carve(at(Block(V(-hw+1.3, bel, -hw+1.3), V(hw-1.3, bel+2.5, hw-1.3))))
		g.Use(White)
		// Spire.
		sp := bel + 2.4
		if cone {
			g.Use(Terracotta).FillVolume(at(Frustum{P(0, 0), hw - 0.4, 0.3, sp, sp + 8.0}))
		} else {
			g.Use(Glass).FillVolume(at(Func{Box{V(-hw, sp, -hw), V(hw, sp+8.5, hw)}, func(p Vec3) bool {
				r := (hw - 0.3) * (1 - (p.Y-sp)/8.5)
				return math.Abs(p.X) <= r && math.Abs(p.Z) <= r
			}}))
		}
		g.Use(Gold).Fill(at(Block(V(-0.45, sp+7.5, -0.45), V(0.45, sp+10.2, 0.45))))
		g.Fill(at(Block(V(-0.45, sp+9.0, -1.1), V(0.45, sp+9.5, 1.1))))
	})
}

// church is San Giacomo di Rialto: brick front with a wooden porch, a big clock and a
// small bell gable, facing the canal across the campo.
func (v VeniceGrandCanal) church(g *Grid, u, d float64) {
	at := veniceGrandCanalFrame{u, d, false}.at
	L := veniceGrandCanalLand
	const hw, dep = 5.0, 7.0
	g.Use(Brick).Fill(at(Block(V(-hw, L-0.5, -dep), V(hw, 9, 0))))
	g.Carve(at(Block(V(-hw+1, L, -dep+1), V(hw-1, 9.1, -1))))
	g.Fill(at(Gable{P(-hw, -1.2), P(hw, 0), 9, AxisZ}))
	g.With(Terracotta, func() {
		g.Fill(at(Func{Box{V(-hw-0.4, 9, -dep-0.4), V(hw+0.4, 9+hw+0.4, 0)}, func(p Vec3) bool {
			t := hw + 0.4 - math.Abs(p.X)
			return p.Y <= 9+t && p.Y >= 9+t-1.1
		}}))
	})
	g.Paint(White, at(Block(V(-hw, L-0.5, -0.6), V(hw, L+0.6, 0.1))))
	g.Carve(at(Arch{V(0, L, 0), 1.2, 2.4, 3, AxisZ}))
	// Clock.
	g.With(White, func() {
		g.Fill(at(Func{Box{V(-1.9, 8.6, 0), V(1.9, 12.4, 0.6)}, func(p Vec3) bool {
			return math.Hypot(p.X, p.Y-10.5) <= 1.9
		}}))
	})
	g.Paint(Navy, at(Func{Box{V(-1.9, 8.6, 0), V(1.9, 12.4, 0.7)}, func(p Vec3) bool {
		return math.Hypot(p.X, p.Y-10.5) <= 1.25
	}}))
	g.Paint(Gold, at(Func{Box{V(-1.9, 8.6, 0), V(1.9, 12.4, 0.7)}, func(p Vec3) bool {
		return math.Hypot(p.X, p.Y-10.5) <= 0.45 || (math.Abs(p.X) < 0.3 && p.Y > 10.5 && p.Y < 11.6)
	}}))
	// Bell gable on the ridge.
	g.With(White, func() {
		g.Fill(at(Block(V(-1.4, 13.5, -1.0), V(1.4, 17.0, 0))))
	})
	g.Carve(at(Arch{V(0, 14.0, -0.5), 0.6, 1.4, 3, AxisZ}))
	// Wooden porch on columns.
	g.With(Wood, func() {
		for _, x := range []float64{-4.3, -1.6, 1.6, 4.3} {
			g.Fill(at(Block(V(x-0.45, L, 3.0), V(x+0.45, L+3.6, 3.9))))
		}
		g.Fill(at(Block(V(-hw, L+3.6, 0), V(hw, L+4.1, 4.0))))
	})
	g.With(Terracotta, func() {
		g.Fill(at(Func{Box{V(-hw-0.3, L+4.1, 0), V(hw+0.3, L+6.2, 4.4)}, func(p Vec3) bool {
			return p.Y <= L+4.1+0.45*(4.4-p.Z)+0.5 && p.Y >= L+4.1+0.45*(4.4-p.Z)-0.5
		}}))
	})
}

// well is a white well-head (vera da pozzo) on a campo.
func (v VeniceGrandCanal) well(g *Grid, u, d float64) {
	at := veniceGrandCanalFrame{u, d, false}.at
	L := veniceGrandCanalLand
	g.With(White, func() {
		g.Fill(at(Cylinder{Center: P(0, 0), Radius: 1.6, Inner: 0.8, Y0: L, Y1: L + 1.6}))
		g.Fill(at(Cylinder{Center: P(0, 0), Radius: 1.9, Y0: L, Y1: L + 0.4}))
	})
	g.With(DarkStone, func() { g.Fill(at(Cylinder{Center: P(0, 0), Radius: 0.8, Y0: L - 0.4, Y1: L + 0.5})) })
}

// taxi is a varnished wooden motor launch with a white cabin, along local X.
func (v VeniceGrandCanal) taxi(g *Grid, f veniceGrandCanalFrame, seed int) {
	at := f.at
	const L, hw = 4.6, 1.3
	hull := Func{Box{V(-L, 0.3, -hw), V(L, 1.6, hw)}, func(p Vec3) bool {
		w := hw
		if p.X > 1 {
			t := (p.X - 1) / (L - 1)
			w = hw * math.Sqrt(math.Max(0, 1-t*t))
		}
		return math.Abs(p.Z) <= math.Max(w, 0.4) && p.X >= -L && p.X <= L
	}}
	g.With(Wood, func() { g.Fill(at(hull)) })
	g.With(White, func() {
		g.Fill(at(Block(V(-2.2, 1.6, -0.9), V(1.0, 2.8, 0.9))))
	})
	g.Paint(Glass, at(Block(V(0.6, 2.0, -1.0), V(1.1, 2.6, 1.0))))
	g.Paint(Navy, at(Block(V(-L, 0.3, -hw-0.1), V(L, 0.7, hw+0.1))))
	if Hash01(seed, 1) < 0.5 {
		g.With(White, func() { g.Fill(at(Block(V(-L-1.6, 0.0, -0.9), V(-L, 0.2, 0.9)))) })
	}
}

// umbrella is a café table under a parasol on the fondamenta.
func (v VeniceGrandCanal) umbrella(g *Grid, u, d float64, c Color) {
	at := veniceGrandCanalFrame{u, d, false}.at
	L := veniceGrandCanalLand
	g.With(Wood, func() {
		g.Fill(at(Block(V(-0.4, L, -0.4), V(0.4, L+3.0, 0.4))))
		g.Fill(at(Cylinder{Center: P(0, 0), Radius: 1.0, Y0: L + 0.8, Y1: L + 1.2}))
	})
	g.With(c, func() {
		g.Fill(at(Func{Box{V(-1.7, L+2.4, -1.7), V(1.7, L+3.6, 1.7)}, func(p Vec3) bool {
			t := L + 3.5 - 0.5*math.Hypot(p.X, p.Z)
			return p.Y <= t && p.Y >= t-0.5
		}}))
	})
}

// vaporetto is a water bus along the canal at (u, d), heading +u·dir.
func (v VeniceGrandCanal) vaporetto(g *Grid, u, d, dir float64) {
	at := veniceGrandCanalFrame{u, d, dir < 0}.at
	const L, hw = 9.0, 2.2
	hull := func(y0, y1, grow float64) Solid {
		return Func{Box{V(-L-grow, y0, -hw-grow), V(L+grow, y1, hw+grow)}, func(p Vec3) bool {
			x := p.X
			w := hw + grow
			if x > L-4 {
				t := (x - (L - 4)) / (4 + grow)
				w = (hw + grow) * math.Sqrt(math.Max(0, 1-t*t))
			}
			if x < -L+1 {
				t := (-L + 1 - x) / (1 + grow)
				w = (hw + grow) * math.Sqrt(math.Max(0, 1-t*t))
			}
			return math.Abs(p.Z) <= w
		}}
	}
	g.With(Charcoal, func() { g.Fill(at(hull(0.3, 1.2, 0))) })
	g.With(White, func() { g.Fill(at(hull(1.2, 2.0, 0))) })
	g.With(DarkStone, func() { g.Fill(at(hull(1.4, 1.8, 0.4))) })
	// Cabin with a band of windows.
	g.With(White, func() {
		g.Fill(at(Block(V(-L+1.5, 2.0, -hw+0.3), V(L-4.5, 4.2, hw-0.3))))
		g.Fill(at(Block(V(1.0, 4.2, -1.3), V(3.6, 5.6, 1.3))))
	})
	g.Carve(at(Block(V(-L+2.5, 2.0, -hw+1.3), V(L-5.5, 4.0, hw-1.3))))
	g.Paint(Glass, at(Func{Box{V(-L+1.5, 2.6, -hw), V(L-4.5, 3.6, hw)}, func(p Vec3) bool {
		return math.Mod(p.X+50, 1.6) < 1.1
	}}))
	g.Paint(Glass, at(Block(V(3.0, 4.6, -1.3), V(3.6, 5.3, 1.3))))
	g.With(Charcoal, func() {
		g.Fill(at(Block(V(-L+1.2, 4.2, -hw+0.1), V(L-4.2, 4.6, hw-0.1))))
	})
	g.With(Gold, func() { g.Fill(at(Block(V(-L+1.5, 3.7, -hw+0.25), V(-L+4.5, 4.2, hw-0.25)))) })
}

// pontoon is a floating vaporetto stop moored off a bank, local +Z towards the canal.
func (v VeniceGrandCanal) pontoon(g *Grid, f veniceGrandCanalFrame) {
	at := f.at
	g.With(Charcoal, func() { g.Fill(at(Block(V(-7, 0.3, 0.8), V(7, 1.2, 4.2)))) })
	g.With(Wood, func() { g.Fill(at(Block(V(-1.2, 0.8, -1.0), V(1.2, 1.4, 1.0)))) })
	g.With(White, func() {
		g.Fill(at(Block(V(-6.6, 1.2, 0.9), V(6.6, 4.4, 4.1))))
	})
	g.Carve(at(Block(V(-5.6, 1.2, 1.9), V(5.6, 4.0, 3.1))))
	g.Carve(at(Block(V(-1.6, 1.2, 0.5), V(1.6, 3.4, 4.6))))
	g.Paint(Glass, at(Func{Box{V(-6.6, 2.0, 0.9), V(6.6, 3.6, 4.1)}, func(p Vec3) bool {
		return math.Mod(p.X+50, 2.0) < 1.3
	}}))
	g.With(Gold, func() { g.Fill(at(Block(V(-6.8, 4.0, 0.7), V(6.8, 4.4, 4.3)))) })
	g.With(Charcoal, func() { g.Fill(at(Block(V(-7, 4.4, 0.6), V(7, 4.9, 4.4)))) })
}

// market adds stalls with striped awnings on a campo.
func (v VeniceGrandCanal) market(g *Grid, f veniceGrandCanalFrame, n int) {
	at := f.at
	L := veniceGrandCanalLand
	for i := 0; i < n; i++ {
		x := (float64(i) - float64(n-1)/2) * 3.6
		g.With(Wood, func() { g.Fill(at(Block(V(x-1.4, L, -1), V(x+1.4, L+1.6, 1)))) })
		g.With(Foliage, func() { g.Fill(at(Block(V(x-1.2, L+1.6, -0.8), V(x+1.2, L+2.0, 0.8)))) })
		awn := []Color{Red, Navy, Red, White}[i%4]
		g.With(White, func() {
			g.Fill(at(Func{Box{V(x-1.7, L+3.0, -1.6), V(x+1.7, L+4.4, 1.6)}, func(p Vec3) bool {
				t := L + 4.2 - 0.6*math.Abs(p.Z)
				return p.Y <= t && p.Y >= t-0.5
			}}))
		})
		g.Paint(awn, at(Func{Box{V(x-1.7, L+3.0, -1.6), V(x+1.7, L+4.4, 1.6)}, func(p Vec3) bool {
			return math.Mod(p.X-x+10.3, 1.4) < 0.7
		}}))
		g.With(Wood, func() {
			g.Fill(at(Block(V(x-1.5, L, -1.5), V(x-1.0, L+3.2, -1.0))), at(Block(V(x+1.0, L, 1.0), V(x+1.5, L+3.2, 1.5))))
		})
	}
}

func (v VeniceGrandCanal) Build(g *Grid) Order {
	E, H := veniceGrandCanalEnd, veniceGrandCanalHalf
	v.ground(g)

	// San Polo bank (Z < 0): fondamenta (Riva del Vin) from the end to the bridge.
	v.lots(g, -1, -E, -33.7, -4, 11)
	v.lots(g, -1, -23.3, -10, -4, 12)
	v.lots(g, -1, 10, E, 0, 13)
	// San Marco bank (Z > 0).
	v.lots(g, 1, -E, -10, 0, 21)
	v.lots(g, 1, 10, 20.3, 0, 22)
	v.lots(g, 1, 30.7, E, 0, 23)

	// The fondamenta along the San Polo bank, with a stone edge.
	at := veniceGrandCanalCanal.at
	g.With(White, func() {
		g.Fill(at(Block(V(-E, veniceGrandCanalLand-0.5, -H-0.6), V(-10, veniceGrandCanalLand, -H+0.4))))
	})
	v.rio(g, -28.5, -1, true)
	v.rio(g, 25.5, 1, false)

	v.rialto(g)
	v.campanile(g, -6.5, 22.0, 23, false)
	v.campanile(g, 24, -26.5, 20, true)
	v.church(g, 3.0, -23.5)
	v.market(g, veniceGrandCanalFrame{-6, -24, false}, 2)
	v.well(g, 4, 22)

	// Riva mooring poles, gondola station and the vaporetto stop.
	for i, u := range []float64{-22, -17, -13} {
		v.pole(g, at, u, -H+1.4, 6.0, []Color{Red, Navy}[i%2], 200+i)
	}
	for i, u := range []float64{-20, -15.5} {
		gf := veniceGrandCanalFrame{u, -H + 3.0, false}
		v.gondola(g, gf.at, 0, 0, 1, 300+i)
		gf = veniceGrandCanalFrame{u + 0.5, -H + 4.9, false}
		v.gondola(g, gf.at, 0, 0, 1, 310+i)
	}
	v.pontoon(g, veniceGrandCanalFrame{-41, -H + 0.2, false})
	v.vaporetto(g, -38, -H+7.2, 1)

	// Gondolas and water taxis under way.
	for i, p := range [][3]float64{{-24, 3, 1}, {18, -3.5, -1}, {34, 2.5, 1}, {-8, -6.5, -1}, {44, -5, -1}} {
		gf := veniceGrandCanalFrame{p[0], p[1], p[2] < 0}
		v.gondola(g, gf.at, 0, 0, 1, 400+i)
	}
	for i, p := range [][3]float64{{24, 5.5, -1}, {-31, -1.5, 1}, {41, 6.5, 1}} {
		v.taxi(g, veniceGrandCanalFrame{p[0], p[1], p[2] < 0}, 500+i)
	}
	// Café umbrellas on the Riva del Vin.
	for i, u := range []float64{-21, -17.5, -14} {
		v.umbrella(g, u, -H-2.2, []Color{Red, White, Red}[i])
	}
	// Light ripples on the water.
	g.Paint(Glass, at(Func{Box{V(-E, -0.1, -H-0.6), V(E, 0.1, H+0.6)}, func(p Vec3) bool {
		return Hash01(int(math.Floor(p.X*1.3)), int(math.Floor(p.Z*1.3)), 77) < 0.06
	}}))
	return Sweep(P(0, 0), 0)
}
