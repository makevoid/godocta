package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// FootballStadium is a modern football ground on match day. A rounded-rectangle bowl
// wraps the pitch, its long axis on X: a red lower tier and a navy-and-white upper tier
// of thin stepped seating skins on raked frames, a band of glazed boxes between them,
// and a slate cantilever roof with a glass inner edge hung from white masts and trusses.
// The bowl is a horseshoe open toward the hero camera: the +X end and both of its
// corners keep only the lower tier under a terrace, and the roof stops short on the +Z
// side, so the striped pitch, its markings, goals, dugouts and the two teams show from
// +X/+Z. Scoreboards stand on the roof of the -X end and on the open terrace; lamps rim
// the roof edge and floodlight masts stand outside three corners. The fourth, at +X/+Z,
// is the entrance: fan stairs with turnstiles climb to the terrace from a plaza with a
// football monument, kiosks and flags. Around the ribbed façade run a walkway and a
// ring road with the team coaches and rows of parked cars on the +Z side, fans and
// trees, all on a rounded plate. The main stand with the dugouts is on -Z.
type FootballStadium struct{}

func init() { Register(FootballStadium{}) }

func (FootballStadium) Name() string  { return "football_stadium" }
func (FootballStadium) Title() string { return "Football Stadium" }

// Rows, ribs and pitch lines are one cell wide, so the plan is laid out in multiples of
// the default cell span to keep them crisp.
const (
	footballStadiumU   = 0.7937005259841         // Cell span at the default density.
	footballStadiumH   = footballStadiumU / 2    // One lattice layer.
	footballStadiumBay = 12 * footballStadiumU   // Structural bay on the straights.
	footballStadiumHX  = 30 * footballStadiumU   // Half length of the side straights.
	footballStadiumHZ  = 18 * footballStadiumU   // Half length of the end straights.
	footballStadiumR0  = 9.25 * footballStadiumU // Corner radius at the foot of the stands.
	footballStadiumPX  = 34 * footballStadiumU   // Pitch half length.
	footballStadiumPZ  = 22 * footballStadiumU   // Pitch half width.

	footballStadiumPlateX = 62.5 // Plate half extents and corner radius.
	footballStadiumPlateZ = 55.0
	footballStadiumPlateR = 22.0
	footballStadiumGround = 2 * footballStadiumH // Center height of the plate's top layer.
)

// footballStadiumPlate is the plate center: the bowl sits toward -X/-Z to leave room for
// the precinct in front of the camera.
var footballStadiumPlate = P(5.5, 5.5)

// footballStadiumRing maps a plan point to bowl coordinates. d is the distance outside
// the rounded rectangle at the foot of the stands (negative toward the pitch). b is the
// bay coordinate around it, 32 bays in all: 0 at the middle of the +X end, 8 the +Z
// side, 16 the -X end, 24 the -Z side, with structural frames on the half-integers. w is
// the plan width of one bay at that point.
func footballStadiumRing(p Vec2) (d, b, w float64) {
	ax, az := math.Abs(p.X), math.Abs(p.Y)
	qx, qz := ax-footballStadiumHX, az-footballStadiumHZ
	var t float64
	switch {
	case qx >= 0 && qz >= 0:
		r := math.Hypot(qx, qz)
		d, t, w = r-footballStadiumR0, 1.5+math.Atan2(qz, qx)*8/math.Pi, r*math.Pi/8
	case qx > qz:
		d, t, w = qx-footballStadiumR0, az/footballStadiumBay, footballStadiumBay
	default:
		d, t, w = qz-footballStadiumR0, 8-ax/footballStadiumBay, footballStadiumBay
	}
	switch {
	case p.X >= 0 && p.Y >= 0:
		b = t
	case p.Y >= 0:
		b = 16 - t
	case p.X < 0:
		b = 16 + t
	default:
		b = 32 - t
	}
	if b >= 32 {
		b -= 32
	}
	return
}

// footballStadiumK is the lattice layer (at the default density) nearest a height.
func footballStadiumK(y float64) int { return int(math.Floor(y/footballStadiumH + 0.5)) }

// footballStadiumRound is the signed distance outside a rounded rectangle at the origin.
func footballStadiumRound(q Vec2, hx, hz, r float64) float64 {
	dx, dz := math.Abs(q.X)-(hx-r), math.Abs(q.Y)-(hz-r)
	return math.Hypot(math.Max(dx, 0), math.Max(dz, 0)) + math.Min(math.Max(dx, dz), 0) - r
}

// footballStadiumFrame returns the plan distance along the bowl to the nearest
// structural frame and to the nearest bay center.
func footballStadiumFrame(b, w float64) (frame, center float64) {
	f := b - math.Floor(b)
	return math.Abs(f-0.5) * w, math.Min(f, 1-f) * w
}

// footballStadiumLower returns the top layer and row of the lower tier at depth d. Row
// -1 is the wall at the pitch side; rows 0–5 step back two cells and up two layers each.
func footballStadiumLower(d float64) (top, row int, ok bool) {
	du := d / footballStadiumU
	switch {
	case du < 0 || du >= 13:
		return 0, 0, false
	case du < 1:
		return 7, -1, true
	}
	row = int((du - 1) / 2)
	return 5 + 2*row, row, true
}

// footballStadiumUpper returns the top layer and row (0–6) of the upper tier at depth d.
func footballStadiumUpper(d float64) (top, row int, ok bool) {
	du := d / footballStadiumU
	if du < 14 || du >= 24.5 {
		return 0, 0, false
	}
	row = int((du - 14) / 1.5)
	return 26 + 2*row, row, true
}

// footballStadiumRoofY is the roof's top surface, rising gently toward the back.
func footballStadiumRoofY(d float64) float64 { return 21.4 + 0.1*(d-7*footballStadiumU) }

// footballStadiumLit reports whether a pixel is a stroke of text set in 3×5 glyphs one
// pixel apart; col runs from the left and row down from the top.
func footballStadiumLit(text string, col, row int) bool {
	glyphs := map[byte][5]string{
		'O': {"###", "#.#", "#.#", "#.#", "###"},
		'C': {"###", "#..", "#..", "#..", "###"},
		'T': {"###", ".#.", ".#.", ".#.", ".#."},
		'A': {"###", "#.#", "###", "#.#", "#.#"},
		'F': {"###", "#..", "##.", "#..", "#.."},
		'1': {".#.", "##.", ".#.", ".#.", "###"},
		'2': {"###", "..#", "###", "#..", "###"},
	}
	if col < 0 || row < 0 || row > 4 {
		return false
	}
	ch, c := col/4, col%4
	if ch >= len(text) || c == 3 {
		return false
	}
	glyph, ok := glyphs[text[ch]]
	return ok && glyph[row][c] == '#'
}

// exact wraps a predicate so that it only accepts true lattice centers. Fill, Paint and
// Carve also probe each center snapped to the half-unit authoring grid, which would blur
// the one-cell rows, ribs and lines laid out on the lattice here.
func (FootballStadium) exact(g *Grid, box Box, f func(Vec3) bool) Func {
	h := g.Span() / 2
	on := func(v float64) bool {
		q := v / h
		return math.Abs(q-math.Round(q)) < 1e-6
	}
	return Func{box, func(p Vec3) bool { return on(p.Y) && on(p.X) && on(p.Z) && f(p) }}
}

// slab is an exact cuboid between two corners.
func (s FootballStadium) slab(g *Grid, a, b Vec3) Func {
	return s.exact(g, BoxOf(a, b), func(Vec3) bool { return true })
}

// ring is a predicate over bowl coordinates, bounded to reach units beyond the foot of
// the stands and to heights y0–y1.
func (s FootballStadium) ring(g *Grid, reach, y0, y1 float64, f func(p Vec3, d, b, w float64) bool) Func {
	ex := footballStadiumHX + footballStadiumR0 + reach
	ez := footballStadiumHZ + footballStadiumR0 + reach
	return s.exact(g, Box{V(-ex, y0, -ez), V(ex, y1, ez)}, func(p Vec3) bool {
		d, b, w := footballStadiumRing(p.Plan())
		return f(p, d, b, w)
	})
}

// sector reports whether bay coordinate b lies outside the open stretch from lo to hi
// (bays either side of the +X end's middle), grown by pad plan units.
func (FootballStadium) sector(b, w, lo, hi, pad float64) bool {
	if b > 16 {
		b -= 32
	}
	return b <= lo+pad/w || b >= hi-pad/w
}

// tiered is where the bowl carries its upper tier; roofed is where the roof covers it.
func (s FootballStadium) tiered(b, w, pad float64) bool { return s.sector(b, w, -5.5, 6.5, pad) }
func (s FootballStadium) roofed(b, w, pad float64) bool { return s.sector(b, w, -5.5, 9.5, pad) }

// stand stacks a figure of single cells on the highest occupied cell under (x, z) below
// y. When want ≥ 0 the figure is only placed if that cell's center is at height want.
func (s FootballStadium) stand(g *Grid, x, z, y, want float64, colors ...Color) bool {
	sp := g.Span()
	best, found := Cell{}, false
	for parity := int32(0); parity < 2; parity++ {
		rnd := func(v float64) int32 {
			return int32(math.Floor((v*2/sp-float64(parity))/2+0.5))*2 + parity
		}
		c := Cell{rnd(x), rnd(y), rnd(z)}
		for n := 0; n < 80 && !g.Has(c); n++ {
			c.Y -= 2
		}
		if g.Has(c) && (!found || c.Y > best.Y) {
			best, found = c, true
		}
	}
	if !found || want >= 0 && math.Abs(float64(best.Y)*sp/2-want) > 0.3 {
		return false
	}
	for i := range colors {
		if g.Has(Cell{best.X, best.Y + int32(2*(i+1)), best.Z}) {
			return false
		}
	}
	for i, col := range colors {
		g.With(col, func() { g.Add(Cell{best.X, best.Y + int32(2*(i+1)), best.Z}) })
	}
	return true
}

func (s FootballStadium) Build(g *Grid) Order {
	s.ground(g)
	s.pitch(g)
	s.bowl(g)
	s.seats(g)
	s.facade(g)
	s.roof(g)
	s.fixtures(g)
	s.match(g)
	s.scoreboards(g)
	s.floodlights(g)
	s.precinct(g)
	return Sweep(P(0, 0), math.Pi/4)
}

// inset is the distance of a plan point inside the plate edge.
func (FootballStadium) inset(p Vec2) float64 {
	return -footballStadiumRound(p.Sub(footballStadiumPlate), footballStadiumPlateX, footballStadiumPlateZ, footballStadiumPlateR)
}

// Precinct rings, as distances outside the foot of the stands.
const (
	footballStadiumApron = 24.4 // Walkway around the whole bowl ends here.
	footballStadiumRoad0 = 28.4 // Ring road on the camera side.
	footballStadiumRoad1 = 32.0
	footballStadiumPark1 = 36.0                  // Outer parking row on the +Z side.
	footballStadiumLayBy = -3 * footballStadiumU // The coaches' lay-by starts at this x.
)

// zone is the ground color at a plan point of the plate.
func (s FootballStadium) zone(p Vec2) Color {
	const U = footballStadiumU
	d, b, _ := footballStadiumRing(p)
	if b > 16 {
		b -= 32
	}
	joint := d > 0 && int(math.Floor(d/U+0.5))%4 == 0
	paving := func(a, c Color) Color {
		if joint {
			return c
		}
		return a
	}
	near := b > -6.5 && b < 14.8
	side := b > 5.5 && b < 12.8 // The +Z side with its two parking rows.
	switch {
	case d <= 1.3*U:
		return Grass
	case near && d >= footballStadiumRoad0 && d < footballStadiumRoad1:
		return DarkStone // The road runs off the plate at both ends.
	case s.inset(p) < 2.4:
		return Grass
	case d < footballStadiumApron:
		return paving(Stone, Sand)
	case !near:
		return Grass
	case d < footballStadiumRoad0:
		if side || b > -1.5 && b < 1.5 {
			return DarkStone // Inner parking row and the coaches' lay-by.
		}
		return paving(Stone, Sand)
	case d < footballStadiumPark1 && side:
		return DarkStone
	case b >= 1.5 && b <= 5.5:
		return paving(Sand, Stone) // Corner plaza.
	}
	return Grass
}

// ground lays the two-layer plate, left open under the stands, and paints the precinct.
func (s FootballStadium) ground(g *Grid) {
	const U = footballStadiumU
	c := footballStadiumPlate
	box := Box{V(c.X-footballStadiumPlateX, 0.3, c.Y-footballStadiumPlateZ), V(c.X+footballStadiumPlateX, 0.9, c.Y+footballStadiumPlateZ)}
	for _, col := range []Color{Grass, Stone, Sand, DarkStone} {
		g.Use(col).Fill(s.exact(g, box, func(p Vec3) bool {
			q := p.Plan()
			if s.inset(q) < 0 {
				return false
			}
			if d, _, _ := footballStadiumRing(q); d > 1.3*U && d < 24.3*U {
				return false
			}
			return s.zone(q) == col
		}))
	}
	// Lawn mottle, road center dashes, the zebra crossing below the corner stairs and
	// parking bay lines.
	g.Paint(Foliage, s.exact(g, box, func(p Vec3) bool {
		q := p.Plan()
		d, _, _ := footballStadiumRing(q)
		return d > 2*U && s.zone(q) == Grass && Hash01(int(math.Floor(p.X/1.6)), int(math.Floor(p.Z/1.6)), 3) < 0.22
	}))
	g.Paint(White, s.exact(g, box, func(p Vec3) bool {
		q := p.Plan()
		if s.zone(q) != DarkStone {
			return false
		}
		d, b, w := footballStadiumRing(q)
		if b > 16 {
			b -= 32
		}
		mid := (footballStadiumRoad0 + footballStadiumRoad1) / 2
		switch {
		case d >= footballStadiumRoad0 && d < footballStadiumRoad1 && b > 3.15 && b < 3.85:
			return math.Mod(b*w+100, 1.6) < 0.8
		case math.Abs(d-mid) < 0.3:
			return math.Mod(b*w+100, 4.8) < 2.4
		case d >= footballStadiumRoad1+0.4 && b > 5.5: // Bay lines of the outer +Z row.
			return math.Abs(p.X-math.Round(p.X/(3*U))*3*U) < 0.3
		case d < footballStadiumRoad0-0.4 && b > 5.5: // Inner +Z row, up to the lay-by.
			return p.X < footballStadiumLayBy+0.3 && math.Abs(p.X-math.Round(p.X/(3*U))*3*U) < 0.3
		case d < footballStadiumRoad0-0.4: // Bay lines of the +X row.
			return math.Abs(p.Z-math.Round(p.Z/(3*U))*3*U) < 0.3
		}
		return false
	}))
}

// pitch mows the stripes and paints the markings.
func (s FootballStadium) pitch(g *Grid) {
	const U, PX, PZ = footballStadiumU, footballStadiumPX, footballStadiumPZ
	box := Box{V(-PX-1, 0.3, -PZ-1), V(PX+1, 0.9, PZ+1)}
	g.Paint(Foliage, s.exact(g, box, func(p Vec3) bool {
		if math.Abs(p.X) > PX+0.1 || math.Abs(p.Z) > PZ+0.1 {
			return false
		}
		return int(math.Floor((p.X+PX+0.25*U)/(4*U)))%2 == 1
	}))
	const t = 0.45
	g.Paint(White, s.exact(g, box, func(p Vec3) bool {
		ax, az := math.Abs(p.X), math.Abs(p.Z)
		if ax > PX+t || az > PZ+t {
			return false
		}
		u := PX - ax // Distance in from the goal line.
		spot := math.Hypot(u-7*U, az)
		switch {
		case ax > PX-t || az > PZ-t: // Touchlines and goal lines.
			return true
		case ax < t: // Halfway line.
			return true
		case math.Abs(math.Hypot(p.X, p.Z)-6*U) < t: // Centre circle.
			return true
		case math.Abs(u-11*U) < t && az < 13*U+t, math.Abs(az-13*U) < t && u < 11*U: // Penalty area.
			return true
		case math.Abs(u-4*U) < t && az < 6*U+t, math.Abs(az-6*U) < t && u < 4*U: // Goal area.
			return true
		case spot < 0.3: // Penalty spot.
			return true
		case u > 11*U && math.Abs(spot-6*U) < t: // Penalty arc.
			return true
		}
		return false
	}))
}

// vomitory returns the floor layer of the stair mouth cut into a tier at (d, b), if any.
func (s FootballStadium) vomitory(d, b, w float64) (floor int, ok bool) {
	_, center := footballStadiumFrame(b, w)
	du := d / footballStadiumU
	lettered := b > 5.5 && b < 10.5 || b > 21.5 && b < 26.5
	switch {
	case lettered && center <= 1.25 && du >= 11 && du < 13:
		return 13, true
	case !lettered && center <= 1.25 && du >= 9 && du < 13:
		return 11, true
	case center <= 1.0 && du >= 15.5 && du < 18.5 && s.tiered(b, w, 0.2):
		return 26, true
	}
	return 0, false
}

// bowl raises the seating: three-layer stepped skins on raked frames, the box band, the
// terrace of the open end, the gable walls where the upper tier stops, and the stair
// mouths.
func (s FootballStadium) bowl(g *Grid) {
	const U = footballStadiumU
	K := footballStadiumK
	// Lower tier, all the way round.
	g.Use(Red).Fill(s.ring(g, 14*U, 1, 6.3, func(p Vec3, d, b, w float64) bool {
		top, row, ok := footballStadiumLower(d)
		if !ok {
			return false
		}
		k := K(p.Y)
		if row < 0 {
			return k >= 3 && k <= top
		}
		return k >= top-2 && k <= top
	}))
	// Box fronts and the upper tier's fascia.
	g.Use(White).Fill(s.ring(g, 15*U, 5, 11.2, func(p Vec3, d, b, w float64) bool {
		k := K(p.Y)
		return d >= 13*U && d < 14*U && k >= 14 && k <= 27 && s.tiered(b, w, 0.2)
	}))
	// Upper tier.
	g.Use(Navy).Fill(s.ring(g, 25*U, 9, 15.5, func(p Vec3, d, b, w float64) bool {
		top, _, ok := footballStadiumUpper(d)
		k := K(p.Y)
		return ok && k >= top-2 && k <= top && s.tiered(b, w, 0.2)
	}))
	// Terrace behind the lower tier of the open end.
	g.Use(Stone).Fill(s.ring(g, 25*U, 5.5, 7.2, func(p Vec3, d, b, w float64) bool {
		if d < 13*U || d >= 24.5*U || s.tiered(b, w, 0.2) {
			return false
		}
		k := K(p.Y)
		return k == 16 || k == 17 || k == 15 && d < 14*U
	}))
	g.Paint(Sand, s.ring(g, 25*U, 5.5, 7.2, func(p Vec3, d, b, w float64) bool {
		return d >= 14*U && d < 24.5*U && !s.tiered(b, w, 0.2) && int(math.Floor(d/U+0.5))%4 == 0
	}))
	// Gable walls closing the two ends of the upper tier, stepped with the rake.
	g.Use(White).Fill(s.ring(g, 26*U, 5, 18, func(p Vec3, d, b, w float64) bool {
		if !s.tiered(b, w, 0.6) || s.tiered(b, w, -0.3) {
			return false
		}
		k := K(p.Y)
		if d >= 13*U && d < 14*U {
			return k >= 14 && k <= 28
		}
		top, _, ok := footballStadiumUpper(d)
		return ok && k >= 14 && k <= top+2
	}))
	// Raked frames on every bay line: a beam under each skin on two columns.
	g.Use(DarkStone).Fill(s.ring(g, 25*U, 1, 15, func(p Vec3, d, b, w float64) bool {
		if frame, _ := footballStadiumFrame(b, w); frame > 0.45 {
			return false
		}
		k, du := K(p.Y), d/U
		if k < 3 {
			return false
		}
		if top, row, ok := footballStadiumLower(d); ok && row >= 0 {
			column := du >= 5 && du < 6 || du >= 11.5 && du < 12.5
			return k <= top-3 && (k >= top-5 || column)
		}
		if top, _, ok := footballStadiumUpper(d); ok && s.tiered(b, w, 0.2) {
			column := du >= 15 && du < 16 || du >= 22.5 && du < 23.5
			return k <= top-3 && (k >= top-5 || column)
		}
		return false
	}))
	// Stair mouths: a notch in the rows with a dark floor.
	g.Carve(s.ring(g, 25*U, 1, 15.5, func(p Vec3, d, b, w float64) bool {
		floor, ok := s.vomitory(d, b, w)
		k := K(p.Y)
		return ok && k > floor && k <= floor+5
	}))
	g.Use(Charcoal).Fill(s.ring(g, 25*U, 1, 15.5, func(p Vec3, d, b, w float64) bool {
		floor, ok := s.vomitory(d, b, w)
		k := K(p.Y)
		return ok && k >= floor-2 && k <= floor
	}))
}

// seats paints the seat blocks: a red lower tier with white lettering on the sides and
// chevrons on the ends, a navy upper tier with white bays and red top rows.
func (s FootballStadium) seats(g *Grid) {
	const U = footballStadiumU
	K := footballStadiumK
	const text = "OCTA FC"
	seat := func(p Vec3, d, b, w float64) (Color, bool) {
		if _, ok := s.vomitory(d, b, w); ok {
			return 0, false
		}
		k := K(p.Y)
		frame, _ := footballStadiumFrame(b, w)
		if top, row, ok := footballStadiumLower(d); ok {
			switch {
			case row < 0:
				return Navy, k >= 3 && k <= top
			case k < top-2 || k > top:
				return 0, false
			case b > 5.5 && b < 10.5 || b > 21.5 && b < 26.5: // Side straights: lettering.
				x := p.X
				if p.Z > 0 {
					x = -x // Read from the pitch on either side.
				}
				col := int(math.Floor((x + 27.25*U) / (2 * U)))
				if footballStadiumLit(text, col, 4-row) {
					return White, true
				}
				return Red, true
			case frame <= 0.45:
				return Stone, true // Gangways.
			case b < 1.5 || b > 30.5 || b > 14.5 && b < 17.5: // End straights: chevrons.
				a := int(math.Floor((math.Abs(p.Z) + 0.75*U) / (2 * U)))
				switch (a + row) % 4 {
				case 0:
					return White, true
				case 1:
					return Navy, true
				}
			}
			return Red, true
		}
		if top, row, ok := footballStadiumUpper(d); ok && s.tiered(b, w, 0.2) {
			switch {
			case k < top-2 || k > top:
				return 0, false
			case frame <= 0.45:
				return Stone, true
			case row >= 5:
				return Red, true
			case int(math.Floor(b+0.5))%3 == 0:
				return White, true
			}
			return Navy, true
		}
		return 0, false
	}
	for _, col := range []Color{Navy, White, Stone, Red} {
		g.Paint(col, s.ring(g, 25*U, 1, 15.5, func(p Vec3, d, b, w float64) bool {
			c, ok := seat(p, d, b, w)
			return ok && c == col
		}))
	}
	// Box glazing between white mullions, and a navy stripe along the upper fascia.
	g.Paint(Glass, s.ring(g, 15*U, 5, 11.2, func(p Vec3, d, b, w float64) bool {
		k := K(p.Y)
		if d < 13*U || d >= 14*U || k < 17 || k > 22 || !s.tiered(b, w, -0.4) {
			return false
		}
		f := b - math.Floor(b)
		return math.Abs(math.Mod(f*4+0.5, 1)-0.5)*w/4 > 0.45
	}))
	g.Paint(Navy, s.ring(g, 15*U, 5, 11.2, func(p Vec3, d, b, w float64) bool {
		k := K(p.Y)
		return d >= 13*U && d < 14*U && k >= 24 && k <= 25 && s.tiered(b, w, -0.4)
	}))
}

// facade wraps the bowl in a wall banded in stone, glass, slate and red, with white
// ribs, taller piers on the bay lines and doorways under red canopies.
func (s FootballStadium) facade(g *Grid) {
	const U = footballStadiumU
	K := footballStadiumK
	stairs := func(b float64) bool { return b > 3.0 && b < 4.0 }
	top := func(b, w float64) int {
		switch {
		case stairs(b):
			return 17
		case s.tiered(b, w, 0.6):
			return 43
		}
		return 20
	}
	wall := func(f func(k int, d, b, w float64) bool) Func {
		return s.ring(g, 26*U, 1, 17.5, func(p Vec3, d, b, w float64) bool {
			k := K(p.Y)
			return d >= 24.5*U && d < 25.5*U && k >= 3 && k <= top(b, w) && f(k, d, b, w)
		})
	}
	g.Use(Slate).Fill(wall(func(k int, d, b, w float64) bool { return true }))
	g.Paint(Stone, wall(func(k int, d, b, w float64) bool { return k <= 11 }))
	g.Paint(Charcoal, wall(func(k int, d, b, w float64) bool {
		_, center := footballStadiumFrame(b, w)
		return k <= 9 && center <= 1.3
	}))
	g.Paint(White, wall(func(k int, d, b, w float64) bool { return k >= 12 && k <= 16 && top(b, w) == 43 }))
	g.Paint(Glass, wall(func(k int, d, b, w float64) bool { return k >= 17 && k <= 22 && top(b, w) == 43 }))
	g.Paint(Red, wall(func(k int, d, b, w float64) bool {
		t := top(b, w)
		return t == 43 && k >= 40 || t == 20 && k >= 18
	}))
	// Ribs and piers. The piers of the roofed stretch rise through the roof as masts.
	g.Use(White).Fill(s.ring(g, 28*U, 1, 31, func(p Vec3, d, b, w float64) bool {
		if d < 25.5*U || d >= 27.5*U || stairs(b) {
			return false
		}
		k := K(p.Y)
		sd := math.Abs(b-math.Floor(b)-0.5) * w
		t := top(b, w) + 2
		if sd <= 0.45 {
			if s.roofed(b, w, 0.6) {
				return k >= 3 && p.Y <= footballStadiumRoofY(26*U)+6
			}
			return k >= 3 && k <= t+1
		}
		return d < 26.5*U && math.Abs(sd-w/3) <= 0.45 && k >= 12 && k <= t-6
	}))
	g.Use(Red).Fill(s.ring(g, 27*U, 3, 5, func(p Vec3, d, b, w float64) bool {
		k := K(p.Y)
		_, center := footballStadiumFrame(b, w)
		return d >= 25.5*U && d < 26.5*U && !stairs(b) && center <= 1.2 && (k == 10 || k == 11)
	}))
}

// roof hangs the cantilever: a two-layer slate plate with a glass inner edge and white
// rims, and on every bay line a white truss tapering from the mast to the tip.
func (s FootballStadium) roof(g *Grid) {
	const U = footballStadiumU
	plate := func(f func(d, run float64) bool) Func {
		return s.ring(g, 28*U, 20, 24, func(p Vec3, d, b, w float64) bool {
			ry := footballStadiumRoofY(d)
			return d >= 7*U && d < 27.5*U && s.roofed(b, w, 0.3) && p.Y <= ry && p.Y > ry-2*footballStadiumH && f(d, b*w)
		})
	}
	g.Use(Slate).Fill(plate(func(d, run float64) bool { return true }))
	g.Paint(Glass, plate(func(d, run float64) bool { return d < 11*U }))
	g.Paint(White, plate(func(d, run float64) bool { return d < 8*U || d >= 26.5*U || d >= 17*U && d < 18*U }))
	// A rim of lamps along the leading edge.
	g.Paint(Gold, plate(func(d, run float64) bool { return d < 8*U && math.Mod(run, 6*U) < 2*U }))
	g.Use(White).Fill(s.ring(g, 28*U, 20, 31, func(p Vec3, d, b, w float64) bool {
		frame, _ := footballStadiumFrame(b, w)
		if frame > 0.45 || d < 7.5*U || d >= 26.5*U || !s.roofed(b, w, 0.6) {
			return false
		}
		t := math.Max(0, math.Min(1, (d-8*U)/(18*U)))
		chord := 0.9 + 4.3*t
		h := p.Y - footballStadiumRoofY(d)
		if h <= 0 || h > chord {
			return false
		}
		du := d / U
		post := du >= 25.3 || math.Abs(du-21) <= 0.55 || math.Abs(du-16) <= 0.55
		return h > chord-0.95 || h <= 0.45 || post
	}))
}

// goal builds one goal on the goal line at x = sx × the pitch half length.
func (s FootballStadium) goal(g *Grid, sx float64) {
	const U, H = footballStadiumU, footballStadiumH
	K := footballStadiumK
	x0 := sx * footballStadiumPX
	box := BoxOf(V(x0-0.3*sx, 1, -4.4*U), V(x0+sx*3*U, 3.7, 4.4*U))
	// Net: a sheet sloping from the crossbar to the ground, with side panels.
	g.Use(Stone).Fill(s.exact(g, box, func(p Vec3) bool {
		u, az, k := (p.X-x0)*sx, math.Abs(p.Z), K(p.Y)
		if u < 0.3*U || az > 4.2*U || k < 3 {
			return false
		}
		slope := 8 - int(math.Round(u/H))
		if az > 3.7*U {
			return k <= slope
		}
		return k == slope
	}))
	g.Use(White).Fill(s.exact(g, box, func(p Vec3) bool {
		u, az, k := (p.X-x0)*sx, math.Abs(p.Z), K(p.Y)
		if math.Abs(u) > 0.2 {
			return false
		}
		return math.Abs(az-4*U) < 0.2 && k >= 3 && k <= 8 || k == 8 && az < 4*U
	}))
}

// fixtures places what stands on the grass: advertising boards, goals, dugouts, the
// players' tunnel and corner flags.
func (s FootballStadium) fixtures(g *Grid) {
	const U = footballStadiumU
	K := footballStadiumK
	// Advertising boards in colored lengths around the pitch, open at the dugouts.
	ads := []Color{Red, White, Navy, Gold, White, Glass}
	box := Box{V(-39*U, 1, -26*U), V(39*U, 2.6, 26*U)}
	for i, col := range ads {
		g.Use(col).Fill(s.exact(g, box, func(p Vec3) bool {
			k := K(p.Y)
			if k != 4 && k != 6 || math.Abs(footballStadiumRound(p.Plan(), 38*U, 25*U, 4*U)) > 0.3 {
				return false
			}
			if p.Z < 0 && math.Abs(p.X) < 18.5*U {
				return false
			}
			seg := int(math.Floor((p.X+0.25*U)/(4*U))) + int(math.Floor((p.Z+0.25*U)/(4*U)))
			return ((seg%len(ads))+len(ads))%len(ads) == i
		}))
	}
	s.goal(g, 1)
	s.goal(g, -1)
	// Dugouts against the -Z stand: white back and ends, glass roof, team benches.
	for i, cx := range []float64{-11 * U, 11 * U} {
		x0, x1 := cx-5.2*U, cx+5.2*U
		g.Use(White).Fill(s.slab(g, V(x0, 1, -26.8*U), V(x1, 2.5, -25.9*U)))
		g.Fill(s.slab(g, V(x0, 1, -26.8*U), V(x0+0.9*U, 2.5, -23.9*U)), s.slab(g, V(x1-0.9*U, 1, -26.8*U), V(x1, 2.5, -23.9*U)))
		g.Use(Glass).Fill(s.slab(g, V(x0, 2.6, -26.8*U), V(x1, 3.3, -23.9*U)))
		g.Use([]Color{Red, Gold}[i]).Fill(s.slab(g, V(x0+U, 1, -25.6*U), V(x1-U, 1.7, -24.9*U)))
	}
	// Tunnel mouth in the pitch wall between the dugouts.
	g.Paint(Charcoal, s.slab(g, V(-1.8*U, 1, -28.4*U), V(1.8*U, 2.5, -27.2*U)))
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			s.stand(g, sx*footballStadiumPX, sz*footballStadiumPZ, 2.2, footballStadiumGround, White, Red)
		}
	}
}

// match sets out two teams, the referee and the assistants as two-cell figures.
func (s FootballStadium) match(g *Grid) {
	const U = footballStadiumU
	type player struct{ x, z float64 }
	home := []player{{-31, 0}, {-22, -14}, {-23, -5}, {-22, 5}, {-21, 15}, {-11, -9}, {-9, 2}, {-12, 12}, {2, -13}, {5, -3}, {3, 10}}
	away := []player{{31, 1}, {22, 13}, {24, 4}, {23, -6}, {21, -15}, {13, 9}, {10, -1}, {14, -11}, {8, 16}, {-3, 6}, {-5, -7}}
	for i, q := range home {
		kit := []Color{White, Red}
		if i == 0 {
			kit = []Color{Charcoal, Glass}
		}
		s.stand(g, q.x*U, q.z*U, 2.2, footballStadiumGround, kit...)
	}
	for i, q := range away {
		kit := []Color{Navy, Gold}
		if i == 0 {
			kit = []Color{Charcoal, Terracotta}
		}
		s.stand(g, q.x*U, q.z*U, 2.2, footballStadiumGround, kit...)
	}
	for _, q := range []player{{1, 4}, {-14, -23}, {16, 23}} {
		s.stand(g, q.x*U, q.z*U, 2.2, footballStadiumGround, Charcoal, Charcoal)
	}
}

// scoreboard is a screen between x0 and x1 on layers k0–k1, showing 2:1 between the team
// colors on each listed face: +1 faces +X, -1 faces -X.
func (s FootballStadium) scoreboard(g *Grid, x0, x1 float64, k0, k1 int, faces ...float64) {
	const U = footballStadiumU
	K := footballStadiumK
	y0, y1 := float64(k0)*footballStadiumH-0.1, float64(k1)*footballStadiumH+0.1
	g.Use(Charcoal).Fill(s.slab(g, V(x0, y0, -9.3*U), V(x1, y1, 9.8*U)))
	g.Paint(White, s.exact(g, Box{V(x0, y0, -9.3*U), V(x1, y1, 9.8*U)}, func(p Vec3) bool {
		k := K(p.Y)
		return k >= k1-1 || k <= k0 || p.Z > 8.7*U || p.Z < -8.3*U
	}))
	for _, face := range faces {
		pixel := func(p Vec3) (col, row int, ok bool) {
			if face > 0 && p.X < x1-0.8*U || face < 0 && p.X > x0+0.8*U {
				return 0, 0, false
			}
			return int(math.Floor((9.75*U - face*p.Z) / U)), (k1 - K(p.Y)) / 2, true
		}
		paint := func(c Color, f func(col, row int) bool) {
			g.Paint(c, s.exact(g, Box{V(x0, y0, -9.3*U), V(x1, y1, 9.8*U)}, func(p Vec3) bool {
				col, row, ok := pixel(p)
				return ok && row >= 1 && row <= 5 && f(col, row)
			}))
		}
		paint(Red, func(col, row int) bool { return col >= 1 && col <= 3 })
		paint(Gold, func(col, row int) bool { return col >= 15 && col <= 17 })
		paint(White, func(col, row int) bool {
			return footballStadiumLit("2", col-5, row-1) || footballStadiumLit("1", col-11, row-1) ||
				col == 9 && (row == 2 || row == 4)
		})
	}
}

// scoreboards stands one screen on the roof of the -X end and one, two-sided, on the
// terrace of the open +X end.
func (s FootballStadium) scoreboards(g *Grid) {
	const U = footballStadiumU
	// On the roof: legs on two trusses, braced back to the roof.
	xa := -(footballStadiumHX + footballStadiumR0 + 8.25*U)
	s.scoreboard(g, xa-1.7*U, xa, 58, 72, 1)
	g.Use(White)
	for _, z := range []float64{-6 * U, 6 * U} {
		g.Fill(s.slab(g, V(xa-1.3*U, 21.3, z-0.3), V(xa-0.2*U, 23, z+0.3)))
		g.Fill(Capsule{V(xa-1.2, 27.6, z), V(xa-8.5, footballStadiumRoofY(18*U)+0.4, z), 0.5})
	}
	// On the terrace: two legs from the deck.
	xb := footballStadiumHX + footballStadiumR0 + 15*U
	s.scoreboard(g, xb, xb+1.7*U, 24, 38, 1, -1)
	for _, z := range []float64{-6 * U, 6 * U} {
		g.Fill(s.slab(g, V(xb+0.2*U, 7, z-0.45), V(xb+1.5*U, 9.6, z+0.45)))
	}
}

// floodlights raises a mast outside three corners, each with a lamp bank turned toward
// the pitch. The entrance corner is left clear: a mast there would stand between the
// hero camera and the pitch.
func (s FootballStadium) floodlights(g *Grid) {
	reach := (footballStadiumR0 + 24.5) / math.Sqrt2
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			if sx > 0 && sz > 0 {
				continue
			}
			c := P(sx*(footballStadiumHX+reach), sz*(footballStadiumHZ+reach))
			g.Use(Stone).FillVolume(Frustum{c, 1.1, 0.75, 1, 35})
			turn := func(sol Solid) Solid {
				return Translate(RotateY(sol, P(0, 0), math.Atan2(c.X, -c.Y)), c.At(0))
			}
			g.Use(Slate).Fill(turn(Block(V(-4.6, 33.2, -0.9), V(4.6, 39, 0.9))))
			g.Paint(White, turn(Block(V(-4, 33.8, 0), V(4, 38.4, 1.4))))
			for _, y := range []float64{34.6, 36.2, 37.8} {
				g.Paint(Gold, turn(Block(V(-4, y-0.2, 0), V(4, y+0.2, 1.4))))
			}
		}
	}
}

// car parks one car with its nose along +Z or +X on the ground.
func (s FootballStadium) car(g *Grid, x, z float64, alongX bool, body Color) {
	const U = footballStadiumU
	w, l := 0.9*U, 1.9*U
	if alongX {
		w, l = l, w
	}
	g.Use(body).Fill(s.slab(g, V(x-w, 1, z-l), V(x+w, 2.1, z+l)))
	cw, cl := w, 0.9*U
	if alongX {
		cw, cl = 0.9*U, l
	}
	cabin := Glass
	if body == White || body == Stone {
		cabin = Charcoal
	}
	g.Use(cabin).Fill(s.slab(g, V(x-cw, 2.2, z-cl), V(x+cw, 2.6, z+cl)))
}

// bus parks a team coach along X with its near corner at (x, z).
func (s FootballStadium) bus(g *Grid, x, z float64, body Color) {
	const U = footballStadiumU
	K := footballStadiumK
	box := Box{V(x-0.1, 1, z-0.1), V(x+12*U+0.1, 4.2, z+3*U+0.1)}
	g.Use(body).Fill(s.exact(g, box, func(p Vec3) bool { return true }))
	g.Paint(Charcoal, s.exact(g, box, func(p Vec3) bool { return K(p.Y) == 3 }))
	g.Paint(Glass, s.exact(g, box, func(p Vec3) bool {
		k := K(p.Y)
		u := (p.X - x) / U
		return k >= 6 && k <= 8 && u > 0.6 && u < 11.4 && math.Mod(u+0.25, 3) > 0.9
	}))
	g.Paint(White, s.exact(g, box, func(p Vec3) bool { return K(p.Y) >= 10 && math.Abs(p.Z-z-1.5*U) < 0.5*U }))
}

// tree plants a round-crowned tree if it fits on the plate.
func (s FootballStadium) tree(g *Grid, x, z, r float64, seed int) {
	if s.inset(P(x, z)) < r*0.7 {
		return
	}
	g.Use(Wood).Fill(Cylinder{Center: P(x, z), Radius: 0.5, Y0: 1, Y1: 3.4})
	crown := Foliage
	if Hash01(seed, 7) < 0.35 {
		crown = Grass
	}
	g.Use(crown).FillVolume(Ellipsoid{V(x, 3.2+r, z), V(r, r*0.95, r)})
}

// precinct builds everything outside the bowl: the corner stairs and turnstiles, flags
// on the terrace rim, buses, cars, trees and the crowd.
func (s FootballStadium) precinct(g *Grid) {
	const U, H = footballStadiumU, footballStadiumH
	K := footballStadiumK
	foot := footballStadiumHZ + footballStadiumR0 // Foot of the +Z stand.
	// Fan stairs up the +X/+Z corner to the terrace, between white cheek walls.
	step := func(d float64) int { return 17 - int(math.Floor((d-25.5*U)/H)) }
	g.Use(Stone).Fill(s.ring(g, 34*U, 1, 7.6, func(p Vec3, d, b, w float64) bool {
		if b <= 3.0 || b >= 4.0 || d < 25.5*U {
			return false
		}
		k, top := K(p.Y), step(d)
		return k >= 3 && k <= top && k >= top-2
	}))
	g.Paint(White, s.ring(g, 34*U, 1, 7.6, func(p Vec3, d, b, w float64) bool {
		k := K(p.Y)
		return b > 3.0 && b < 4.0 && d >= 25.5*U && k >= 3 && k%4 < 2
	}))
	g.Use(Navy).Fill(s.ring(g, 34*U, 1, 8.6, func(p Vec3, d, b, w float64) bool {
		if d < 25.5*U || d > 25.5*U+16*H {
			return false
		}
		k, top := K(p.Y), step(d)
		if math.Abs(b-3.5)*w <= 0.45 {
			return k > top && k <= top+2 // Center handrail.
		}
		edge := math.Min(math.Abs(b-3.0), math.Abs(b-4.0)) * w
		return edge <= 0.5 && k >= 3 && k <= top+3
	}))
	// Turnstile booths under a red canopy at the head of the stairs.
	g.Use(Navy).Fill(s.ring(g, 25*U, 6.9, 8.6, func(p Vec3, d, b, w float64) bool {
		k := K(p.Y)
		return b > 3.06 && b < 3.94 && d >= 22.5*U && d < 24*U && k >= 18 && k <= 21 &&
			int(math.Floor((b-3)*w/1.6))%2 == 0
	}))
	g.Use(Red).Fill(s.ring(g, 25*U, 8.5, 9.4, func(p Vec3, d, b, w float64) bool {
		k := K(p.Y)
		return b > 3.03 && b < 3.97 && d >= 22*U && d < 24.5*U && (k == 22 || k == 23)
	}))
	// Flags on the piers of the terrace rim.
	flags := []Color{Red, White, Navy, Gold}
	for i := -5; i <= 5; i++ {
		if i == 3 {
			continue // The stairs take this pier's place.
		}
		// Pier i stands on bay line i + 0.5; find its plan position on the wall.
		c := s.point(float64(i)+0.5, 26.2*U)
		g.Use(White).Fill(Capsule{c.At(8.4), c.At(12.6), 0.5})
		g.Use(flags[(i+5)%len(flags)]).Fill(Block(V(c.X+0.3, 10.9, c.Y-0.3), V(c.X+3.1, 12.7, c.Y+0.3)))
	}
	// Team coaches in the lay-by on the +Z side.
	s.bus(g, 0, foot+footballStadiumApron+0.7, Red)
	s.bus(g, 15*U, foot+footballStadiumApron+0.7, Navy)
	// Parked cars: two rows either side of the road on +Z, one against the apron on +X.
	paint := []Color{Red, White, Navy, Slate, Gold, Charcoal, Water, Stone, Terracotta}
	inner := (footballStadiumApron + footballStadiumRoad0) / 2
	outer := (footballStadiumRoad1 + footballStadiumPark1) / 2
	for n := -19; n <= 9; n++ {
		x := (float64(3*n) + 1.5) * U
		if x < footballStadiumLayBy && s.zone(P(x, foot+inner)) == DarkStone && Hash01(n, 35) > 0.2 {
			s.car(g, x, foot+inner, false, paint[int(Hash01(n, 36)*float64(len(paint)))])
		}
		if s.zone(P(x, foot+outer+1.6)) == DarkStone && s.zone(P(x, foot+outer-1.6)) == DarkStone && Hash01(n, 31) > 0.14 {
			s.car(g, x, foot+outer, false, paint[int(Hash01(n, 32)*float64(len(paint)))])
		}
	}
	for n := -6; n <= 5; n++ {
		z := (float64(3*n) + 1.5) * U
		if Hash01(n, 33) > 0.14 {
			s.car(g, footballStadiumHX+footballStadiumR0+inner, z, true, paint[int(Hash01(n, 34)*float64(len(paint)))])
		}
	}
	// Trees along the back and end verges, bushes along the car park.
	c := footballStadiumPlate
	n := 0
	for x := c.X - 52; x <= c.X+52; x += 6.5 {
		s.tree(g, x+Hash01(n, 41)-0.5, c.Y-footballStadiumPlateZ+3, 1.9+0.5*Hash01(n, 42), n)
		n++
	}
	for z := c.Y - 44; z <= c.Y+44; z += 6.5 {
		s.tree(g, c.X-footballStadiumPlateX+2.2, z+Hash01(n, 43)-0.5, 1.5, n)
		n++
		if z < 16 {
			s.tree(g, c.X+footballStadiumPlateX-3, z+Hash01(n, 43)-0.5, 1.8+0.4*Hash01(n, 44), n)
		}
		n++
	}
	for x := -46.0; x <= 26; x += 4.8 {
		if q := P(x, c.Y+footballStadiumPlateZ-1.3); s.inset(q) > 1 {
			g.Use(Foliage).FillVolume(Sphere(q.At(1.5), 1.05))
		}
	}
	// The corner plaza: a giant football on a golden plinth in a roundel, kiosks, flags
	// and trees.
	ball := s.point(3.5, 38.6)
	g.Paint(Navy, Cylinder{Center: ball, Radius: 4.3, Inner: 3.2, Y0: 0.3, Y1: 1.2})
	g.Paint(White, Cylinder{Center: ball, Radius: 3.2, Y0: 0.3, Y1: 1.2})
	g.Use(Gold).Fill(BlockAt(ball, 2.4, 2.4, 1, 2.6))
	center := ball.At(4.5)
	g.Use(White).FillVolume(Sphere(center, 2.1))
	g.Paint(Charcoal, Func{BoxOf(center.Sub(V(2.5, 2.5, 2.5)), center.Add(V(2.5, 2.5, 2.5))), func(p Vec3) bool {
		n := p.Sub(center)
		return math.Max(math.Abs(n.X), math.Max(math.Abs(n.Y), math.Abs(n.Z))) > 0.86*n.Length() && n.Length() > 1.2
	}})
	for i, b := range []float64{2.3, 4.7} {
		q := s.point(b, 34.4)
		g.Use(White).Fill(BlockAt(q, 3.2, 2.4, 1, 3.3))
		g.Paint(Charcoal, BlockAt(q, 2.2, 3, 1.9, 2.9))
		g.Use([]Color{Red, Navy}[i]).Fill(BlockAt(q, 4.2, 3.4, 3.3, 4.2))
	}
	banners := []Color{Red, White, Navy, Gold}
	for i, at := range [][2]float64{{2.7, 36.5}, {3.05, 40}, {3.95, 41.5}, {4.3, 38.3}} {
		q := s.point(at[0], at[1])
		g.Use(White).Fill(Capsule{q.At(1), q.At(10.4), 0.5})
		g.Use(banners[i]).Fill(Block(V(q.X+0.4, 7.6, q.Y-0.3), V(q.X+3, 10.4, q.Y+0.3)))
	}
	s.tree(g, 30, 57.6, 1.9, 201)
	s.tree(g, 36.5, 57, 1.7, 202)
	s.tree(g, 65, 18.5, 1.8, 203)
	// The crowd: fans in club colors on the walkway, the plaza, the stairs and the terrace.
	shirts := []Color{Red, Red, Red, White, Navy}
	legs := []Color{Navy, Charcoal}
	crowd := func(count, seed int, b0, b1, d0, d1, y, want float64) {
		for i := 0; i < count; i++ {
			q := s.point(b0+(b1-b0)*Hash01(i, seed, 53), d0+(d1-d0)*Hash01(i, seed, 54))
			if ground := s.zone(q); want == footballStadiumGround && ground != Stone && ground != Sand {
				continue
			}
			s.stand(g, q.X, q.Y, y, want, legs[int(Hash01(i, seed, 51)*float64(len(legs)))], shirts[int(Hash01(i, seed, 52)*float64(len(shirts)))])
		}
	}
	crowd(60, 1, -5.4, 10.4, 21.9, footballStadiumApron, 2.2, footballStadiumGround)
	crowd(45, 2, 1.6, 5.4, footballStadiumRoad1+0.5, 45, 2.2, footballStadiumGround)
	crowd(20, 3, 1.6, 5.4, footballStadiumApron, footballStadiumRoad0, 2.2, footballStadiumGround)
	crowd(6, 4, 3.12, 3.44, 20.6, 26.4, 8.6, -1)
	crowd(6, 6, 3.56, 3.88, 20.6, 26.4, 8.6, -1)
	crowd(60, 5, -5.2, 6.2, 14.5*U, 24*U, 8.2, 17*H)
}

// point returns the plan position on bay coordinate b (within ±8 of the +X end) at
// distance d outside the foot of the stands.
func (FootballStadium) point(b, d float64) Vec2 {
	sz := 1.0
	if b < 0 {
		b, sz = -b, -1
	}
	r := footballStadiumR0 + d
	switch {
	case b <= 1.5:
		return P(footballStadiumHX+r, sz*b*footballStadiumBay)
	case b <= 5.5:
		a := (b - 1.5) * math.Pi / 8
		return P(footballStadiumHX+r*math.Cos(a), sz*(footballStadiumHZ+r*math.Sin(a)))
	}
	return P((8-b)*footballStadiumBay, sz*(footballStadiumHZ+r))
}
