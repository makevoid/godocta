package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// ContainerPort is a container terminal with a feeder ship at berth. The quay runs along
// X with the basin on the +Z side: a navy container ship (bow +X, port side to the quay)
// with red boot-topping, a bulbous bow, a white accommodation block with bridge wings, a
// banded funnel and deck stacks of mixed boxes, one bay half discharged down to its hatch
// covers. Four red ship-to-shore gantry cranes stand on rails along the quay: three with
// their booms lowered over the ship, each with a trolley, spreader and hanging box, and
// one at the stern with its boom raised to 70°. Behind the apron lies the yard: two long
// blocks under yellow rubber-tyred gantries, an empties stack, a reefer rack, a warehouse,
// fuel tanks, a glazed control tower, light masts and, along the back, a rail siding with
// a locomotive and container flats. A tug pushes on the starboard bow, a pilot boat crosses
// the basin and mooring dolphins take the head and stern lines. Boxes are laid out on the
// lattice itself, so every one is a crisp 3 × 3 × 15 (or 7) cell block with dark seams.
type ContainerPort struct{}

func init() { Register(ContainerPort{}) }

func (ContainerPort) Name() string  { return "container_port" }
func (ContainerPort) Title() string { return "Container Port" }

const (
	containerPortGround   = 3.17 // Height of the terminal's paving course.
	containerPortQuay     = 13.0 // Quay face (z); the basin lies beyond it.
	containerPortKeel     = 23.0 // Ship centerline (z).
	containerPortBeam     = 8.5  // Ship half-beam.
	containerPortDeck     = 7.5  // Ship main deck.
	containerPortSeaRail  = 10.5 // Waterside crane rail (z).
	containerPortLandRail = -7.5 // Landside crane rail (z).
	containerPortGirder   = 32.4 // Underside of the crane girders.
	containerPortRTGBack  = -32.3
	containerPortRTGFront = -12.4
	containerPortLane     = -15.9 // Truck lane under the yard gantries (z).
	containerPortTrack    = -49.5 // Rail siding centerline (z).
)

// containerPortLiveries are the box colors, weighted toward the common shipping lines.
var containerPortLiveries = []Color{Red, Navy, Terracotta, Gold, White, Slate, Foliage, Brick, Red, Terracotta, White, Gold, Navy}

func containerPortLivery(seed ...int) Color {
	return containerPortLiveries[int(Hash01(seed...)*float64(len(containerPortLiveries)))]
}

func containerPortClamp(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

func containerPortFrac(v float64) float64 { return v - math.Floor(v) }

// containerPortNoise is smooth value noise in [0, 1).
func containerPortNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return (a+(b-a)*sx)*(1-sz) + (c+(d-c)*sx)*sz
}

// containerPortRound is the signed distance to a rounded rectangle, negative inside.
func containerPortRound(x, z, x0, z0, x1, z1, r float64) float64 {
	dx := math.Abs(x-(x0+x1)/2) - (x1-x0)/2 + r
	dz := math.Abs(z-(z0+z1)/2) - (z1-z0)/2 + r
	return math.Hypot(math.Max(dx, 0), math.Max(dz, 0)) + math.Min(math.Max(dx, dz), 0) - r
}

// containerPortPlate is the signed distance to the edge of the sea plate.
func containerPortPlate(x, z float64) float64 {
	return containerPortRound(x, z, -66, -57.5, 66, 48.5, 14)
}

// containerPortLand is the signed distance to the edge of the terminal: rounded at the
// back, cut straight along the quay.
func containerPortLand(x, z float64) float64 {
	return math.Max(containerPortRound(x, z, -58.5, -53, 53.5, 30, 8), z-containerPortQuay)
}

// containerPortLift is what hangs under a crane trolley.
type containerPortLift struct {
	trolley float64 // Trolley position along the girder (z).
	hoist   float64 // Height of the load's underside.
	cells   int     // Box length in cells: 15, 7, or 0 for an empty spreader.
	color   Color
}

// containerPortStack is a block of stacked containers laid out on the lattice. Boxes are
// 3 × 3 cells in section on a 4-cell pitch; the fourth cell is the dark seam.
type containerPortStack struct {
	at         Cell  // Low corner cell of the first box.
	alongX     bool  // Long axis of the boxes.
	bays, rows int   // Slots along and across.
	length     int32 // Slot length in lattice steps: 28 is a forty-foot box, 12 a twenty.
	pitch      int32 // Bay pitch in steps: length+4 butts bays on a seam, more leaves an air gap.
	tiers      func(bay, row int) int
	paint      func(bay, row, tier, half int) Color
	twin       func(bay, row, tier int) bool // Two twenties in a forty-foot slot.
}

func (s ContainerPort) Build(g *Grid) Order {
	s.ship(g)
	s.boats(g)
	s.land(g)
	s.sea(g)
	s.cranes(g)
	s.yard(g)
	s.depot(g)
	s.rail(g)
	s.traffic(g)
	s.masts(g)
	return Along(P(1, 0.25))
}

// put fills solids in a color.
func (ContainerPort) put(g *Grid, color Color, solids ...Solid) {
	g.With(color, func() { g.Fill(solids...) })
}

// set occupies a cell with a color, replacing whatever was there.
func (ContainerPort) set(g *Grid, c Cell, color Color) {
	previous := g.Brush()
	g.Remove(c)
	g.Use(color).Add(c)
	g.Use(previous)
}

// even snaps a world point to the sublattice of even coordinates, on which every
// lattice-exact part of the scene is laid out.
func (ContainerPort) even(g *Grid, p Vec3) Cell {
	r := func(v float64) int32 { return int32(math.Round(v/g.Span())) * 2 }
	return Cell{r(p.X), r(p.Y), r(p.Z)}
}

// each visits every lattice cell whose true center lies inside the box.
func (ContainerPort) each(g *Grid, b Box, f func(Cell, Vec3)) {
	u := g.Span() / 2
	lo := func(v float64) int32 { return int32(math.Ceil(v/u - 1e-9)) }
	hi := func(v float64) int32 { return int32(math.Floor(v/u + 1e-9)) }
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
				c := Cell{x, y, z}
				f(c, g.World(c))
			}
		}
	}
}

// tint recolors the occupied cells of a box for which f picks a color. Unlike Paint it
// tests only true cell centers, so one-cell bands and window rows stay one cell wide.
func (s ContainerPort) tint(g *Grid, b Box, f func(Vec3) (Color, bool)) {
	s.each(g, b, func(c Cell, p Vec3) {
		if !g.Has(c) {
			return
		}
		if color, ok := f(p); ok {
			s.set(g, c, color)
		}
	})
}

// box fills a crisp cuboid of nx × ny × nz cells whose low corner cell is c: the cells of
// c's sublattice plus the other sublattice's cells between them.
func (s ContainerPort) box(g *Grid, c Cell, nx, ny, nz int, color Color) {
	for i := 0; i < 2*nx-1; i++ {
		for j := 0; j < 2*ny-1; j++ {
			for k := 0; k < 2*nz-1; k++ {
				if i&1 == j&1 && j&1 == k&1 {
					s.set(g, Cell{c.X + int32(i), c.Y + int32(j), c.Z + int32(k)}, color)
				}
			}
		}
	}
}

// windows paints a grid of one-cell windows on the walls of a box: one row per storey of
// rise cells above y0, one window every `every` cells along the wall.
func (s ContainerPort) windows(g *Grid, b Box, y0 float64, rise, every int, color Color) {
	span := g.Span()
	s.tint(g, b.Grow(0.3), func(p Vec3) (Color, bool) {
		if containerPortFrac((p.Y-y0)/(float64(rise)*span)-0.4) >= 1/float64(rise) {
			return 0, false
		}
		along := p.X
		if math.Min(p.X-b.Min.X, b.Max.X-p.X) < math.Min(p.Z-b.Min.Z, b.Max.Z-p.Z) {
			along = p.Z
		}
		return color, containerPortFrac(along/(float64(every)*span)) < 1/float64(every)
	})
}

// stack rasterizes a block of containers as a hollow shell: only the cells within one
// cell of an outer face are placed.
func (s ContainerPort) stack(g *Grid, k containerPortStack) {
	const rowPitch, rowWide, tierPitch = 8, 4, 8
	slot := func(d, pitch, wide int32, n int) (i int, joint, ok bool) {
		if d < 0 || int(d/pitch) >= n {
			return 0, false, false
		}
		i = int(d / pitch)
		if d%pitch > wide {
			return i, true, i+1 < n
		}
		return i, false, true
	}
	tiers := func(dl, dw int32) int {
		b, jb, okb := slot(dl, k.pitch, k.length, k.bays)
		r, jr, okr := slot(dw, rowPitch, rowWide, k.rows)
		if !okb || !okr || jb && k.pitch > k.length+4 {
			return 0
		}
		t := k.tiers(b, r)
		if jb {
			t = min(t, k.tiers(b+1, r))
		}
		if jr {
			t = min(t, k.tiers(b, r+1))
		}
		if jb && jr {
			t = min(t, k.tiers(b+1, r+1))
		}
		return t
	}
	inside := func(dl, dw, dy int32) bool {
		t := tiers(dl, dw)
		return t > 0 && dy >= -1 && dy <= int32(tierPitch*t-4)
	}
	most := 0
	for b := 0; b < k.bays; b++ {
		for r := 0; r < k.rows; r++ {
			most = max(most, k.tiers(b, r))
		}
	}
	for dy := int32(-1); dy <= int32(tierPitch*most-4); dy++ {
		for dl := int32(dy & 1); dl <= int32(k.bays-1)*k.pitch+k.length; dl += 2 {
			for dw := int32(dy & 1); dw <= int32(k.rows-1)*rowPitch+rowWide; dw += 2 {
				if !inside(dl, dw, dy) {
					continue
				}
				if inside(dl-2, dw, dy) && inside(dl+2, dw, dy) && inside(dl, dw-2, dy) &&
					inside(dl, dw+2, dy) && inside(dl, dw, dy+2) {
					continue
				}
				b, jb, _ := slot(dl, k.pitch, k.length, k.bays)
				r, jr, _ := slot(dw, rowPitch, rowWide, k.rows)
				t, lt := int(dy+1)/tierPitch, (dy+1)%tierPitch
				color := Charcoal
				if !jb && !jr && lt <= 5 {
					l, half, seam := dl%k.pitch, 0, false
					if k.length == 28 && k.twin != nil && k.twin(b, r, t) {
						seam, half = l > 12 && l < 16, int(l/16)
					}
					if !seam {
						color = k.paint(b, r, t, half)
					}
				}
				c := Cell{k.at.X + dw, k.at.Y + dy, k.at.Z + dl}
				if k.alongX {
					c = Cell{k.at.X + dl, k.at.Y + dy, k.at.Z + dw}
				}
				s.set(g, c, color)
			}
		}
	}
}

// hullW is the ship's outer half-beam at (x, y), negative beyond the stem and stern.
func (ContainerPort) hullW(x, y float64) float64 {
	rise := containerPortClamp(y/9, 0, 1)
	stem := 49.5 + 2.6*rise
	stern := -50 - 1.5*containerPortClamp((y-2.5)/3, 0, 1)
	if x > stem || x < stern {
		return -1
	}
	w := containerPortBeam
	if x > 33 { // Entry, flaring fuller toward the forecastle.
		t := (x - 33) / (stem - 33)
		w *= 1 - math.Pow(t, 1.8+1.6*rise)
	}
	if x < -40 { // Run, drawn in low down under the transom.
		t := (-40 - x) / 11.5
		w *= 1 - (0.12+0.3*containerPortClamp((5-y)/5, 0, 1))*t*t
	}
	return w
}

// ship builds the container ship: hull, hatches, deck stacks, accommodation and funnel.
func (s ContainerPort) ship(g *Grid) {
	const zc, main, fore = containerPortKeel, containerPortDeck, 10.2
	span := g.Span()
	hullBox := Box{V(-52, -0.2, zc-9.6), V(55.5, 11.6, zc+9.6)}
	deck := func(x float64) float64 {
		if x >= 35 {
			return fore
		}
		return main
	}
	g.Use(Navy).FillVolume(Func{hullBox, func(p Vec3) bool {
		return p.Y <= deck(p.X) && math.Abs(p.Z-zc) <= s.hullW(p.X, p.Y)
	}})
	g.FillVolume(Intersect(Ellipsoid{V(50.6, -0.2, zc), V(4.3, 1.9, 2.3)}, Cuboid(hullBox)))
	inner := func(p Vec3, inset float64) bool {
		a, b := s.hullW(p.X-inset, p.Y), s.hullW(p.X+inset, p.Y)
		return a >= 0 && b >= 0 && math.Abs(p.Z-zc) < math.Min(a, b)-inset
	}
	g.Carve(Func{hullBox, func(p Vec3) bool {
		top := deck(p.X)
		if p.X >= 35 && p.X < 36.2 { // The bulkhead at the break of the forecastle.
			top = main
		}
		return p.Y < top-0.8 && inner(p, 1.05)
	}})
	// Livery: red boot-topping, a brick-red deck inside the sheer strake.
	g.Paint(Red, Block(V(-53, -1, zc-10), V(56, 1.45, zc+10)))
	s.tint(g, hullBox, func(p Vec3) (Color, bool) { return Brick, p.Y > deck(p.X)-0.9 && inner(p, 0.7) })
	// Bulwarks around the forecastle and the poop.
	rim := func(p Vec3, y, toward float64) bool {
		o := s.hullW(p.X, y) - math.Abs(p.Z-zc)
		return p.Y > y && p.Y <= y+1.2 && o >= 0 && (o < 1 || s.hullW(p.X+toward, y) < 0)
	}
	s.put(g, Navy, Func{hullBox, func(p Vec3) bool {
		return p.X >= 35 && rim(p, fore, 1) || p.X <= -46 && rim(p, main, -1)
	}})
	// Anchors in their hawse pockets and a white draft mark at the stem.
	s.tint(g, Box{V(42.5, 6.2, zc-9.6), V(44.5, 8.2, zc+9.6)}, func(p Vec3) (Color, bool) { return Charcoal, true })

	// Hatch covers on their coamings, with dark joints between the panels.
	s.put(g, DarkStone, Block(V(-28.4, main, zc-7.9), V(33.2, 8.9, zc+7.9)))
	g.Carve(Block(V(-27.4, main-1.5, zc-6.9), V(32.2, 8.2, zc+6.9)))
	first := s.even(g, V(-27, 9.5, zc-7.14))
	s.tint(g, Box{V(-28.4, 8.25, zc-7.9), V(33.2, 9, zc+7.9)}, func(p Vec3) (Color, bool) {
		if math.Abs(p.Z-zc) < span/2 || containerPortFrac((p.X-float64(first.X)*span/2+1.5*span)/(8.5*span)) < 2/17.0 {
			return Charcoal, true
		}
		return Stone, true
	})

	// Deck stacks: four forty-foot bays and a twenty-foot bay forward. Bay 3 is being
	// discharged from the seaward side, bay 1 has just lost a box from its second row.
	table := [][]int{{4, 4, 4, 3, 2}, {4, 2, 4, 4, 3}, {3, 4, 4, 4, 3}, {3, 2, 1, 0, 0}}
	s.stack(g, containerPortStack{
		at: first, alongX: true, bays: 4, rows: 5, length: 28, pitch: 34,
		tiers: func(b, r int) int { return table[b][r] },
		paint: func(b, r, t, h int) Color { return containerPortLivery(1, b, r, t, h) },
		twin:  func(b, r, t int) bool { return Hash01(2, b, r, t) < 0.4 },
	})
	bow := []int{2, 3, 3, 3, 2}
	s.stack(g, containerPortStack{
		at: Cell{first.X + 4*34, first.Y, first.Z}, alongX: true, bays: 1, rows: 5, length: 12, pitch: 16,
		tiers: func(b, r int) int { return bow[r] },
		paint: func(b, r, t, h int) Color { return containerPortLivery(3, b, r, t, h) },
	})

	// Accommodation block: five decks of cabins, then the wheelhouse and its wings.
	storey := 4 * span
	top := main + 5*storey
	house := Box{V(-41, main, zc-7), V(-30.5, top, zc+7)}
	s.put(g, White, Cuboid(house))
	g.Carve(Block(V(-39.9, main+0.1, zc-5.9), V(-31.6, top-0.9, zc+5.9)))
	s.windows(g, house, main, 4, 3, Charcoal)
	bridge := Box{V(-40.2, top, zc-6), V(-32.4, top+storey, zc+6)}
	s.put(g, White, Cuboid(bridge),
		Block(V(-39.4, top, zc-10.2), V(-34.2, top+0.7, zc+10.2)),
		Block(V(-40.6, top+storey, zc-6.5), V(-32, top+storey+0.5, zc+6.5)))
	g.Carve(Block(V(-39.1, top+0.1, zc-4.9), V(-33.5, top+storey-0.6, zc+4.9)))
	for _, side := range []float64{-1, 1} {
		s.put(g, White, Block(V(-39.4, top, zc+side*9.4), V(-34.2, top+1.9, zc+side*10.2)))
	}
	s.tint(g, bridge.Grow(0.2), func(p Vec3) (Color, bool) {
		return Glass, p.Y >= top+1.1 && p.Y < top+1.1+1.5*span
	})
	// Monkey island: radar mast, scanner and a satellite dome.
	s.put(g, White, Capsule{V(-37, top+storey, zc), V(-37, top+storey+5.5, zc), 0.5},
		Sphere(V(-34.5, top+storey+1.4, zc+3.4), 1.1))
	s.put(g, Charcoal, Capsule{V(-37, top+storey+3.4, zc-2.2), V(-37, top+storey+3.4, zc+2.2), 0.45})
	s.put(g, Red, Sphere(V(-37, top+storey+6, zc), 0.5))

	// Engine casing and the raked funnel with its gold band.
	s.put(g, White, Block(V(-47.6, main, zc-4.6), V(-41, 13.5, zc+4.6)))
	g.Carve(Block(V(-46.5, main+0.1, zc-3.5), V(-41, 12.7, zc+3.5)))
	s.windows(g, Box{V(-47.6, main, zc-4.6), V(-41.4, 13.5, zc+4.6)}, main, 4, 3, Charcoal)
	rake := func(b Cuboid) Solid { return Sheared{b, P(-0.07, 0), 13.5} }
	s.put(g, Navy, rake(Block(V(-46.6, 13.5, zc-2.6), V(-42.3, 28, zc+2.6))))
	g.Carve(rake(Block(V(-45.5, 13.4, zc-1.5), V(-43.4, 27.2, zc+1.5))))
	s.tint(g, Box{V(-48.2, 13.6, zc-2.8), V(-42, 28.4, zc+2.8)}, func(p Vec3) (Color, bool) {
		switch {
		case p.Y >= 26.3:
			return Charcoal, true
		case p.Y >= 20.6 && p.Y < 23.8:
			return Gold, true
		}
		return 0, false
	})
	// Free-fall lifeboat on its ramp over the stern.
	for _, side := range []float64{-1.4, 1.4} {
		s.put(g, Charcoal, Capsule{V(-46.8, 14.1, zc+side), V(-51.3, 11.9, zc+side), 0.45},
			Capsule{V(-50.6, main, zc+side), V(-50.6, 12, zc+side), 0.45})
	}
	s.put(g, Terracotta, Capsule{V(-47.2, 15.5, zc), V(-51, 13.6, zc), 1.25})

	// Forecastle: windlasses, bollards and the foremast.
	for _, side := range []float64{-1, 1} {
		s.put(g, Charcoal, Block(V(41.6, fore, zc+side*1.4), V(43.6, fore+1.4, zc+side*3.4)))
		s.put(g, Gold, Block(V(38, fore, zc+side*4.2), V(38.8, fore+0.9, zc+side*5)))
	}
	s.put(g, White, Capsule{V(46.4, fore, zc), V(46.4, fore+7.5, zc), 0.5},
		Capsule{V(46.4, fore+5.4, zc-1.7), V(46.4, fore+5.4, zc+1.7), 0.45})
	s.put(g, Gold, Sphere(V(46.4, fore+8, zc), 0.5))
	for _, x := range []float64{-49.5, -48} {
		s.put(g, Gold, Block(V(x, main, zc-5.4), V(x+0.7, main+0.9, zc-4.6)), Block(V(x, main, zc+4.6), V(x+0.7, main+0.9, zc+5.4)))
	}
}

// boats floats the harbor tug on the ship's starboard bow and the pilot boat.
func (s ContainerPort) boats(g *Grid) {
	span := g.Span()
	// Tug, bow toward the ship (-Z).
	const tx, bow, stern = 41.5, 32.0, 45.2
	half := func(z float64) float64 {
		if z < bow || z > stern {
			return -1
		}
		w := 2.9
		if z < bow+5 {
			t := (bow + 5 - z) / 5
			w *= math.Sqrt(1 - 0.8*t*t)
		}
		return w
	}
	sheer := func(z float64) float64 {
		t := containerPortClamp((bow+6-z)/6, 0, 1)
		return 2.3 + 0.9*t*t
	}
	tug := Box{V(tx-3.2, -0.2, bow-0.2), V(tx+3.2, 3.6, stern+0.2)}
	g.Use(Charcoal).FillVolume(Func{tug, func(p Vec3) bool { return p.Y <= sheer(p.Z) && math.Abs(p.X-tx) <= half(p.Z) }})
	well := func(p Vec3, y float64) bool {
		return p.Y > y && math.Abs(p.X-tx) < half(p.Z)-0.9 && p.Z > bow+1.2 && p.Z < stern-0.9
	}
	g.Carve(Func{tug, func(p Vec3) bool { return well(p, 1.5) }})
	g.Paint(Red, Block(V(tx-4, -1, bow-1), V(tx+4, 0.6, stern+1)))
	s.tint(g, tug, func(p Vec3) (Color, bool) { return Brick, well(p, 0.8) })
	s.put(g, White, Block(V(tx-1.9, 1.5, 35.4), V(tx+1.9, 4.5, 40.6)), Block(V(tx-1.6, 4.5, 35.8), V(tx+1.6, 7, 39)))
	s.tint(g, Box{V(tx-1.7, 5.2, 35.7), V(tx+1.7, 5.2+1.2*span, 39.1)}, func(p Vec3) (Color, bool) { return Glass, true })
	s.put(g, Charcoal, Block(V(tx-1.9, 7, 35.5), V(tx+1.9, 7.5, 39.3)))
	for _, side := range []float64{-1.2, 1.2} {
		s.put(g, Red, Block(V(tx+side-0.5, 4.5, 39.6), V(tx+side+0.5, 7.4, 40.6)))
		s.put(g, Charcoal, Block(V(tx+side-0.5, 7.4, 39.6), V(tx+side+0.5, 8, 40.6)))
	}
	s.put(g, White, Capsule{V(tx, 7.5, 37.4), V(tx, 11, 37.4), 0.45})
	s.put(g, Gold, Block(V(tx-1.1, 1.5, 41.8), V(tx+1.1, 2.9, 43.2)), Sphere(V(tx, 11.3, 37.4), 0.5))

	// Pilot boat heading +X across the basin.
	const pz, aft, fwd = 41.5, -29.0, -19.5
	beam := func(x float64) float64 {
		if x < aft || x > fwd {
			return -1
		}
		w := 1.8
		if x > fwd-4 {
			t := (x - fwd + 4) / 4
			w *= math.Sqrt(1 - 0.9*t*t)
		}
		return w
	}
	pilot := Box{V(aft-0.2, -0.2, pz-2), V(fwd+0.2, 1.7, pz+2)}
	g.Use(Terracotta).FillVolume(Func{pilot, func(p Vec3) bool { return math.Abs(p.Z-pz) <= beam(p.X) }})
	g.Paint(Charcoal, Block(V(aft-1, -1, pz-3), V(fwd+1, 0.2, pz+3)))
	s.put(g, White, Block(V(-26.4, 1.7, pz-1.3), V(-22.6, 3.9, pz+1.3)))
	s.tint(g, Box{V(-26.5, 2.5, pz-1.4), V(-22.5, 2.5+1.1*span, pz+1.4)}, func(p Vec3) (Color, bool) { return Glass, true })
	s.put(g, Charcoal, Capsule{V(-25.4, 3.9, pz), V(-25.4, 6.4, pz), 0.45})
	s.put(g, Gold, Sphere(V(-25.4, 6.6, pz), 0.5))
}

// land lays the terminal platform: paving, the quay wall with its fenders and bollards,
// and rock armor around the other three sides.
func (s ContainerPort) land(g *Grid) {
	span := g.Span()
	u := span / 2
	site := Box{V(-62, -0.2, -57), V(57, containerPortGround+0.1, containerPortQuay+1)}
	on := func(v, at float64) bool { return math.Abs(v-at) < span/2 }
	track := float64(s.even(g, V(0, 0, containerPortTrack-span)).Z) * u
	paving := func(x, z float64) Color {
		switch {
		case z > containerPortQuay-span:
			return White
		case on(z, containerPortSeaRail), on(z, containerPortLandRail):
			return Charcoal
		case on(z, 11.9):
			return Gold
		case z > -10.6: // Apron: concrete bays with truck lane lines.
			if on(z, 6.4) || on(z, -3.2) || on(z, 1.6) && containerPortFrac(x/(6*span)) < 0.5 {
				return White
			}
			if containerPortFrac(x/(8*span)) < 0.125 {
				return DarkStone
			}
			return Stone
		case z > track-2.5*span && z < track+4.5*span: // Rail siding on ballast.
			if on(z, track) || on(z, track+2*span) {
				return Charcoal
			}
			if containerPortFrac(x/(2*span)) < 0.5 {
				return Wood
			}
			return Stone
		case math.Abs(z-containerPortRTGFront) < 1.1, math.Abs(z-containerPortRTGBack) < 1.1:
			return Stone // Gantry runways.
		case on(z, -33.8) && containerPortFrac(x/(6*span)) < 0.5:
			return White
		case on(z, -18.3), on(x, -0.6) && z < -13.5 && z > -33 && containerPortFrac(z/(4*span)) < 0.5:
			return White
		}
		// Asphalt, patched with paler concrete slabs.
		if Hash01(int(math.Floor(x/(8*span))), int(math.Floor(z/(4*span))), 7) < 0.14 {
			return Stone
		}
		return DarkStone
	}
	y := containerPortGround
	s.each(g, Box{V(site.Min.X, y-0.1, site.Min.Z), V(site.Max.X, y+0.1, site.Max.Z)}, func(c Cell, p Vec3) {
		if containerPortLand(p.X, p.Z) <= 0 {
			s.set(g, c, paving(p.X, p.Z))
		}
	})
	// Quay wall, darker in the tidal zone, with rubber fenders and yellow bollards.
	s.put(g, Stone, Func{site, func(p Vec3) bool {
		return p.Z > containerPortQuay-0.9 && p.Y < y-0.2 && containerPortLand(p.X, p.Z) <= 0
	}})
	g.Paint(DarkStone, Block(V(site.Min.X, -1, containerPortQuay-1), V(site.Max.X, 1.0, containerPortQuay+0.1)))
	for x := -54.0; x <= 50; x += 6.5 {
		s.put(g, Charcoal, Block(V(x-0.7, 0.8, containerPortQuay), V(x+0.7, 3.0, containerPortQuay+0.8)))
	}
	for x := -57.0; x <= 52; x += 9 {
		c := s.even(g, V(x, y+span, containerPortQuay-1.5*span))
		s.box(g, c, 1, 1, 1, Gold)
	}
	// Rock armor sloping into the water.
	slope := func(p Vec3) float64 { return y + 0.1 - 1.25*containerPortLand(p.X, p.Z) }
	armor := Func{site, func(p Vec3) bool {
		d := containerPortLand(p.X, p.Z)
		h := slope(p)
		return p.Z <= containerPortQuay && d > 0 && d <= 2.9 && p.Y <= h && p.Y > h-1.1
	}}
	s.put(g, DarkStone, armor)
	s.tint(g, site, func(p Vec3) (Color, bool) {
		if !armor.Contains(p) {
			return 0, false
		}
		k := Hash01(int(math.Floor(p.X/1.7)), int(math.Floor(p.Y/1.2)), int(math.Floor(p.Z/1.7)), 9)
		return Stone, k < 0.4
	})
}

// sea floods the plate around everything already afloat, then adds wakes and wash.
func (s ContainerPort) sea(g *Grid) {
	const zc = containerPortKeel
	plate := Box{V(-67, -0.1, -58), V(67, 0.1, 49)}
	s.each(g, plate, func(c Cell, p Vec3) {
		if g.Has(c) || containerPortPlate(p.X, p.Z) > 0 {
			return
		}
		if d := containerPortLand(p.X, p.Z); d < 0 || d < 1.6 && p.Z <= containerPortQuay {
			return
		}
		if w := s.hullW(p.X, 0); w > 1.6 && math.Abs(p.Z-zc) < w-1 {
			return
		}
		color := Water
		if containerPortNoise(p.X*0.06, p.Z*0.32, 3) > 0.66 {
			color = Glass
		}
		s.set(g, c, color)
	})
	foam := func(in func(p Vec3) float64) {
		s.tint(g, plate, func(p Vec3) (Color, bool) {
			c := s.even(g, p)
			if col, _ := g.ColorOf(c); col != Water && col != Glass {
				return 0, false
			}
			return White, Hash01(int(c.X), int(c.Z), 5) < in(p)
		})
	}
	// Propeller wash astern of the tug, and the pilot boat's wake.
	foam(func(p Vec3) float64 {
		d := math.Hypot((p.X-41.5)/2.6, (p.Z-47.2)/2.6)
		return 1.1 - d
	})
	foam(func(p Vec3) float64 {
		along := -29 - p.X
		if along < -9.5 || along > 16 {
			return 0
		}
		off := math.Abs(p.Z - 41.5)
		if along < 0 { // Bow wave hugging the hull.
			return containerPortClamp(1.2-math.Abs(off-2.4), 0, 0.9)
		}
		arm := 1.2 + 0.28*along
		return containerPortClamp(1.3-math.Abs(off-arm), 0, 1) * (1 - along/18)
	})
	// Mooring dolphins off the bow and stern, channel buoys, and the ship's lines.
	for _, q := range []Vec2{P(58.6, 16.2), P(-62.4, 16.6)} {
		for k := 0; k < 3; k++ {
			pile := Polar(q, 1.1, float64(k)*2*math.Pi/3+0.5)
			s.put(g, DarkStone, Cylinder{Center: pile, Radius: 0.6, Y0: -0.2, Y1: 2.6})
		}
		s.put(g, Stone, Cylinder{Center: q, Radius: 1.9, Y0: 2.6, Y1: 3.7})
		s.put(g, Gold, Cylinder{Center: q, Radius: 0.6, Y0: 3.7, Y1: 4.7})
	}
	s.put(g, Red, Frustum{P(-48, 43), 1.2, 0.5, -0.2, 2.4})
	s.put(g, White, Sphere(V(-48, 2.8, 43), 0.5))
	s.put(g, Foliage, Cylinder{Center: P(22, 44.5), Radius: 1.1, Y0: -0.2, Y1: 1.8}, Frustum{P(22, 44.5), 0.9, 0.3, 1.8, 3})
	rope := func(a, b Vec3) { s.put(g, Sand, Capsule{a, b, 0.45}) }
	rope(V(48.2, 11.2, zc-3.6), V(58.6, 4.7, 16.2))
	rope(V(40.5, 11.2, zc-6.6), V(44.5, 3.8, containerPortQuay-0.6))
	rope(V(-50.8, 8.5, zc-4.8), V(-62.4, 4.7, 16.6))
	rope(V(-47.8, 8.5, zc-6.6), V(-43.5, 3.8, containerPortQuay-0.6))
}

// hang draws a spreader on its cables below a trolley, with the box it carries.
func (s ContainerPort) hang(g *Grid, x, z, top float64, lift containerPortLift) {
	span := g.Span()
	n := lift.cells
	if n == 0 {
		n = 15
	}
	c := s.even(g, V(x-float64(n-1)/2*span, lift.hoist, z-span))
	if lift.cells > 0 {
		s.box(g, c, n, 3, 3, lift.color)
		c.Y += 6
	}
	s.box(g, c, n, 1, 3, Gold)
	y := float64(c.Y) * span / 2
	s.put(g, Charcoal, Block(V(x-1.3, y+0.4, z-1), V(x+1.3, y+1.5, z+1)))
	for _, dx := range []float64{-1.9, 1.9} {
		s.put(g, Charcoal, Capsule{V(x+dx, y+1.2, z), V(x+dx, top, z), 0.5})
	}
}

// crane is a ship-to-shore gantry centered on x: a portal on rail bogies, twin girders
// with a machinery house over the backreach, an A-frame with stays and the boom, either
// lowered over the ship with a working trolley or raised to 70°.
func (s ContainerPort) crane(g *Grid, cx float64, raised bool, lift containerPortLift) {
	const zw, zl = containerPortSeaRail, containerPortLandRail
	const half, gy, gt, rail = 7.0, containerPortGirder, containerPortGirder + 2, 2.6
	const reach, back = 30.0, 12.5
	for _, z := range []float64{zw, zl} {
		for _, x := range []float64{cx - half, cx + half} {
			s.put(g, Charcoal, Block(V(x-2.6, 3.6, z-0.9), V(x+2.6, 5, z+0.9)))
			s.put(g, Red, BlockAt(P(x, z), 1.6, 1.6, 5, gt+1.8))
		}
		s.put(g, Red,
			Block(V(cx-half, 5, z-0.8), V(cx+half, 6.6, z+0.8)),
			Block(V(cx-half, 15, z-0.7), V(cx+half, 16.6, z+0.7)),
			Block(V(cx-half, gt, z-0.8), V(cx+half, gt+1.8, z+0.8)))
	}
	for _, x := range []float64{cx - half, cx + half} {
		s.put(g, Red, Block(V(x-0.7, 15, zl), V(x+0.7, 16.6, zw)), Capsule{V(x, 16.6, zl), V(x, gt, zw), 0.6})
	}
	// Fixed girders over the portal and backreach, tied at both ends.
	hinge := zw + 1.5
	for _, x := range []float64{cx - rail, cx + rail} {
		s.put(g, Red, Block(V(x-0.6, gy, zl-back), V(x+0.6, gt, hinge)))
	}
	s.put(g, Red, Block(V(cx-rail, gy, zl-back), V(cx+rail, gt, zl-back+1.2)))
	// Machinery house over the landside legs.
	house := Box{V(cx-4.6, gt, zl-10.5), V(cx+4.6, gt+5.2, zl+1.2)}
	s.put(g, White, Cuboid(house))
	g.Carve(Block(V(cx-3.5, gt+0.9, zl-9.4), V(cx+3.5, gt+4.3, zl+0.1)))
	s.windows(g, house, gt+0.6, 6, 3, Charcoal)
	s.tint(g, house.Grow(0.2), func(p Vec3) (Color, bool) { return Red, p.Y > gt+4.5 && math.Abs(p.X-cx) > 3.6 })
	s.put(g, Charcoal, Block(V(cx-1.4, gt+5.2, zl-8.6), V(cx+1.4, gt+6.3, zl-5.8)))

	// A-frame over the waterside legs, its apex tied back to the end of the girders.
	apex := func(x float64) Vec3 { return V(cx+x, gt+18, zw-1.5) }
	for _, side := range []float64{-1, 1} {
		a := apex(side * rail)
		s.put(g, Red, Capsule{V(cx+side*half, gt+1.8, zw), a, 0.75},
			Capsule{a, V(cx+side*rail, gt, zl-back+1), 0.5})
	}
	s.put(g, Red, Capsule{apex(-rail), apex(rail), 0.8})
	s.put(g, White, Sphere(apex(0).Add(V(0, 1.1, 0)), 0.7))

	if raised {
		// Boom hoisted about its hinge; the forestays fold toward the apex.
		sin, cos := math.Sincos(70 * math.Pi / 180)
		at := func(x, along float64) Vec3 { return V(x, gy+1+along*sin, hinge+along*cos) }
		for _, side := range []float64{-1, 1} {
			x := cx + side*rail
			boom := Func{Box{V(x-0.6, gy, hinge-2), V(x+0.6, gy+reach, hinge+reach*cos+2)}, func(p Vec3) bool {
				dz, dy := p.Z-hinge, p.Y-gy-1
				along, across := dz*cos+dy*sin, dy*cos-dz*sin
				return along >= -0.8 && along <= reach-1.5 && math.Abs(across) <= 1
			}}
			s.put(g, Red, boom)
			s.tint(g, boom.Box, func(p Vec3) (Color, bool) {
				along := (p.Z-hinge)*cos + (p.Y-gy-1)*sin
				return White, boom.Contains(p) && along > reach-9 && int(math.Floor(along/1.9))%2 == 0
			})
			s.put(g, Red, Capsule{apex(side * rail), at(x, 13), 0.5}, Capsule{apex(side * rail), at(x, 25), 0.5})
		}
		s.put(g, Red, Capsule{at(cx-rail, reach-2), at(cx+rail, reach-2), 0.8})
		lift = containerPortLift{trolley: zl + 4, hoist: gy - 5.5}
	} else {
		tip := hinge + reach
		for _, side := range []float64{-1, 1} {
			x := cx + side*rail
			s.put(g, Red, Block(V(x-0.6, gy, hinge), V(x+0.6, gt, tip)))
			s.put(g, Red, Capsule{apex(side * rail), V(x, gt, hinge+13), 0.5}, Capsule{apex(side * rail), V(x, gt, tip-3), 0.5})
		}
		s.put(g, Red, Block(V(cx-rail, gy, tip-1.2), V(cx+rail, gt, tip)))
		s.tint(g, Box{V(cx-rail-0.7, gy-0.1, tip-8), V(cx+rail+0.7, gt+0.1, tip+0.1)}, func(p Vec3) (Color, bool) {
			return White, int(math.Floor((tip-p.Z)/1.9))%2 == 1
		})
	}
	// Trolley under the girders with the operator's cab, and the load on its cables.
	zt := lift.trolley
	s.put(g, Charcoal, Block(V(cx-3.4, gy-1.2, zt-2.4), V(cx+3.4, gy, zt+2.4)))
	s.put(g, White, Block(V(cx+0.9, gy-3.8, zt+2.5), V(cx+3.3, gy-1.2, zt+4.9)))
	s.tint(g, Box{V(cx+0.8, gy-3.9, zt+2.4), V(cx+3.4, gy-2.6, zt+5)}, func(p Vec3) (Color, bool) { return Glass, true })
	s.hang(g, cx, zt, gy-1.2, lift)
}

// cranes stands the four gantries along the quay.
func (s ContainerPort) cranes(g *Grid) {
	const zc = containerPortKeel
	s.crane(g, -38, true, containerPortLift{})
	s.crane(g, -7.94, false, containerPortLift{trolley: 4, hoist: 9.5, cells: 15, color: Gold})
	s.crane(g, 19.05, false, containerPortLift{trolley: zc + 3.2, hoist: 16.5, cells: 15, color: Terracotta})
}

// rtg is a rubber-tyred yard gantry straddling a block and its truck lane.
func (s ContainerPort) rtg(g *Grid, cx float64, lift containerPortLift) {
	const z0, z1, top, wide = containerPortRTGBack, containerPortRTGFront, 22.6, 3.6
	for _, z := range []float64{z0, z1} {
		for _, x := range []float64{cx - wide, cx + wide} {
			s.put(g, Charcoal, Block(V(x-1.7, 3.6, z-0.8), V(x+1.7, 5, z+0.8)))
			s.put(g, Gold, BlockAt(P(x, z), 1.4, 1.4, 5, top))
		}
		s.put(g, Gold, Block(V(cx-wide-1.7, 5, z-0.7), V(cx+wide+1.7, 6.2, z+0.7)),
			Block(V(cx-wide, top-1.6, z-0.7), V(cx+wide, top, z+0.7)))
	}
	for _, x := range []float64{cx - wide, cx + wide} {
		s.put(g, Gold, Block(V(x-0.7, top-1.6, z0), V(x+0.7, top, z1)))
	}
	s.put(g, White, Block(V(cx-2.2, 6.2, z0-1.2), V(cx+2.2, 8.8, z0+1.2)))
	s.put(g, Charcoal, Block(V(cx-2.2, 8.8, z0-1.2), V(cx+2.2, 9.3, z0+1.2)))
	zt := lift.trolley
	s.put(g, Charcoal, Block(V(cx-wide, top, zt-2.2), V(cx+wide, top+1.2, zt+2.2)))
	s.put(g, White, Block(V(cx-1.2, top-4.6, zt-4.6), V(cx+1.2, top-1.7, zt-2.4)))
	s.tint(g, Box{V(cx-1.3, top-4.7, zt-4.7), V(cx+1.3, top-3.4, zt-2.3)}, func(p Vec3) (Color, bool) { return Glass, true })
	s.hang(g, cx, zt, top, lift)
}

// yard lays the two gantry blocks behind the apron and their three gantries.
func (s ContainerPort) yard(g *Grid) {
	span := g.Span()
	u := span / 2
	y := containerPortGround + span
	row := func(c Cell, r int) float64 { return float64(c.Z+int32(8*r)+2) * u }
	bay := func(c Cell, b int) float64 { return float64(c.X+int32(32*b)+14) * u }

	// Block A: four bays, worked by two gantries.
	a := s.even(g, V(-54, y, -30.2))
	tiersA := func(b, r int) int {
		switch {
		case b == 1 && r == 3:
			return 2
		case b == 3 && r == 2:
			return 2
		}
		return 2 + int(Hash01(11, b, r)*2.6)
	}
	s.stack(g, containerPortStack{
		at: a, alongX: true, bays: 4, rows: 4, length: 28, pitch: 32, tiers: tiersA,
		paint: func(b, r, t, h int) Color { return containerPortLivery(12, b, r, t, h) },
		twin:  func(b, r, t int) bool { return Hash01(13, b, r, t) < 0.35 },
	})
	s.rtg(g, bay(a, 1), containerPortLift{trolley: containerPortLane, hoist: 8.2, cells: 15, color: Navy})
	s.rtg(g, bay(a, 3), containerPortLift{trolley: row(a, 2), hoist: 16.8, cells: 15, color: Red})

	// Block B: three bays and one gantry, its spreader coming down empty.
	b := s.even(g, V(3, y, -30.2))
	tiersB := func(bay, r int) int {
		if bay == 1 && r == 1 {
			return 2
		}
		return 2 + int(Hash01(14, bay, r)*2.6)
	}
	s.stack(g, containerPortStack{
		at: b, alongX: true, bays: 3, rows: 4, length: 28, pitch: 32, tiers: tiersB,
		paint: func(b, r, t, h int) Color { return containerPortLivery(15, b, r, t, h) },
		twin:  func(b, r, t int) bool { return Hash01(16, b, r, t) < 0.35 },
	})
	s.rtg(g, bay(b, 1), containerPortLift{trolley: row(b, 1), hoist: 13.4})
}

// depot fills the back of the terminal: warehouse, empties stack, reefer rack, fuel
// tanks and the control tower.
func (s ContainerPort) depot(g *Grid) {
	span := g.Span()
	u := span / 2
	floor := containerPortGround + span/2
	y := containerPortGround + span

	// Warehouse: white shed with a slate roof, skylights and a row of loading doors.
	shed := Box{V(-56, floor, -46.5), V(-33.5, floor+6.4, -36)}
	s.put(g, White, Cuboid(shed))
	g.Carve(Block(V(-54.9, floor, -45.4), V(-34.6, floor+7, -37.1)))
	eave := shed.Max.Y
	s.put(g, Slate, Gable{P(-56.7, -47.2), P(-32.8, -35.3), eave, AxisX})
	g.Carve(Gable{P(-55.7, -46.5), P(-33.8, -36), eave - 0.6, AxisX})
	s.put(g, White, Intersect(Gable{P(-56, -46.5), P(-33.5, -36), eave, AxisX},
		Union(Block(V(-56, eave, -47), V(-55, eave+7, -35)), Block(V(-34.5, eave, -47), V(-33.5, eave+7, -35)))))
	s.tint(g, Box{V(-57, eave+0.5, -48), V(-32, eave+7, -35)}, func(p Vec3) (Color, bool) {
		ridge := math.Abs(p.Z + 41.25)
		if p.X < -55.2 || p.X > -34.3 {
			return 0, false
		}
		return Glass, ridge > 1.6 && ridge < 3.4 && containerPortFrac((p.X+56)/(7*span)) < 3/7.0
	})
	s.tint(g, shed.Grow(0.2), func(p Vec3) (Color, bool) {
		switch {
		case p.Y > eave-1.0:
			return Navy, true
		case p.Z > -36.9 && p.Y < floor+4.2 && containerPortFrac((p.X+55)/(6*span)) < 0.5:
			return Charcoal, true
		}
		return 0, false
	})

	// Empties stack: tall single-line columns waiting to be repositioned.
	e := s.even(g, V(-28.5, y, -45.6))
	lines := []Color{Slate, Terracotta, Navy, Red, Foliage, White}
	s.stack(g, containerPortStack{
		at: e, alongX: true, bays: 2, rows: 3, length: 28, pitch: 32,
		tiers: func(b, r int) int { return 3 + int(Hash01(21, b, r)*2.9) },
		paint: func(b, r, t, h int) Color { return lines[int(Hash01(22, b, r, t/2)*float64(len(lines)))] },
	})

	// Reefer rack: white refrigerated twenties between steel service gantries.
	r := s.even(g, V(3.2, y, -45.6))
	s.stack(g, containerPortStack{
		at: r, alongX: true, bays: 3, rows: 3, length: 12, pitch: 22,
		tiers: func(b, r int) int { return 3 + int(Hash01(23, b, r)*1.6) },
		paint: func(b, r, t, h int) Color {
			if Hash01(24, b, r, t) < 0.18 {
				return Glass
			}
			return White
		},
	})
	for k := 0; k < 2; k++ {
		x0 := float64(r.X+int32(22*k)+14) * u
		z0, z1 := float64(r.Z-2)*u, float64(r.Z+22)*u
		for t := 0; t < 4; t++ {
			top := float64(r.Y+int32(8*t)-2) * u
			s.put(g, Charcoal, Block(V(x0+0.2, top, z0), V(x0+2.8, top+0.5, z1)))
		}
		for _, z := range []float64{z0, z1} {
			s.put(g, Gold, BlockAt(P(x0+1.5, z), 1, 1, floor, float64(r.Y+28)*u))
		}
	}

	// Fuel tanks inside a bund, with a pipe run toward the quay.
	for i, c := range []Vec2{P(35.4, -41), P(46.2, -41)} {
		h := 8.6 - 1.4*float64(i)
		s.put(g, White, Cylinder{Center: c, Radius: 4.4, Inner: 3.3, Y0: floor, Y1: floor + h})
		g.With(White, func() { g.FillVolume(Frustum{c, 4.4, 0.6, floor + h, floor + h + 1.6}) })
		g.Carve(Frustum{c, 3.4, 0, floor + h - 0.2, floor + h + 0.7})
		s.tint(g, Box{V(c.X-5, floor+h-2.4, c.Y-5), V(c.X+5, floor+h-2.4+span, c.Y+5)}, func(p Vec3) (Color, bool) { return Red, true })
		s.put(g, Charcoal, Capsule{V(c.X-4.6, floor, c.Y+1.5), V(c.X-4.6, floor+h+0.6, c.Y+1.5), 0.45})
	}
	s.put(g, Stone, Subtract(Block(V(29.6, floor, -46.9), V(51.8, floor+1.1, -35.1)), Block(V(30.5, floor-1, -46), V(50.9, floor+2, -36))))
	s.put(g, Gold, Capsule{V(40.8, floor+0.5, -41), V(40.8, floor+0.5, -34), 0.5}, Capsule{V(35.4, floor+0.5, -41), V(46.2, floor+0.5, -41), 0.5})

	// Control tower: a white shaft, a glazed cab and an operations wing.
	c := P(48.2, -22.5)
	s.put(g, White, BlockAt(c, 4.8, 4.8, floor, floor+15))
	g.Carve(BlockAt(c, 2.6, 2.6, floor, floor+14))
	s.windows(g, Box{V(c.X-2.4, floor, c.Y-2.4), V(c.X+2.4, floor+12.5, c.Y+2.4)}, floor, 4, 3, Slate)
	cab := Box{V(c.X-3.8, floor+15, c.Y-3.8), V(c.X+3.8, floor+18.6, c.Y+3.8)}
	s.put(g, Glass, Cuboid(cab))
	g.Carve(BlockAt(c, 5.4, 5.4, floor+15.8, floor+17.8))
	s.put(g, Navy, BlockAt(c, 7.6, 7.6, floor+14.2, floor+15.2))
	s.put(g, Charcoal, BlockAt(c, 8.6, 8.6, floor+18.6, floor+19.3))
	s.tint(g, cab.Grow(0.2), func(p Vec3) (Color, bool) {
		corner := math.Abs(math.Abs(p.X-c.X)-3.6) < 0.5 && math.Abs(math.Abs(p.Z-c.Y)-3.6) < 0.5
		return Charcoal, corner && p.Y < floor+18.6
	})
	s.put(g, White, Capsule{c.At(floor + 19.3), c.At(floor + 24.5), 0.45})
	s.put(g, Red, Sphere(c.At(floor+24.9), 0.5))
	wing := Box{V(43.2, floor, -31), V(52, floor+4.8, -26.4)}
	s.put(g, White, Cuboid(wing))
	g.Carve(Block(V(44.3, floor, -29.9), V(50.9, floor+4, -27.5)))
	s.tint(g, wing.Grow(0.2), func(p Vec3) (Color, bool) {
		switch {
		case p.Y > floor+4.2:
			return Slate, true
		case p.Y > floor+1.8 && p.Y < floor+1.8+1.2*span:
			return Glass, containerPortFrac(p.X/(4*span)) < 0.75
		}
		return 0, false
	})
}

// rail runs the siding along the back: a diesel locomotive and five container flats.
func (s ContainerPort) rail(g *Grid) {
	span := g.Span()
	base := s.even(g, V(0, containerPortGround+span, containerPortTrack-span))
	at := func(x float64) Cell { return Cell{s.even(g, V(x, 0, 0)).X, base.Y, base.Z} }
	// Flats: a dark deck on bogies, loaded with one forty or two twenties.
	for i := 0; i < 5; i++ {
		c := at(-54 + 14.3*float64(i))
		s.box(g, c, 17, 1, 3, Charcoal)
		c.X, c.Y = c.X+2, c.Y+2
		switch {
		case i == 3: // An empty wagon.
		case Hash01(31, i) < 0.5:
			s.box(g, c, 15, 3, 3, containerPortLivery(32, i))
		default:
			s.box(g, c, 7, 3, 3, containerPortLivery(33, i))
			s.box(g, Cell{c.X + 16, c.Y, c.Z}, 7, 3, 3, containerPortLivery(34, i))
		}
	}
	// Locomotive at the head, cab forward.
	c := at(18.5)
	s.box(g, c, 18, 1, 3, Charcoal)
	c.Y += 2
	s.box(g, Cell{c.X + 2, c.Y, c.Z}, 12, 3, 3, Terracotta)
	s.box(g, Cell{c.X + 26, c.Y, c.Z}, 4, 4, 3, Terracotta)
	s.box(g, Cell{c.X + 26, c.Y + 8, c.Z}, 4, 1, 3, Charcoal)
	s.box(g, Cell{c.X + 4, c.Y + 6, c.Z}, 10, 1, 3, Charcoal)
	for k := int32(0); k < 3; k++ {
		for _, x := range []int32{26, 32} {
			s.set(g, Cell{c.X + x, c.Y + 6, c.Z + 2*k}, Glass)
		}
	}
	s.box(g, Cell{c.X + 2, c.Y, c.Z}, 16, 1, 1, Gold)
	s.box(g, Cell{c.X + 2, c.Y, c.Z + 4}, 16, 1, 1, Gold)
	// Buffer stop.
	s.put(g, Red, Block(V(47.6, containerPortGround+0.3, containerPortTrack-1.4), V(48.6, containerPortGround+2.2, containerPortTrack+1.4)))
}

// truck is a tractor and skeletal chassis heading along X (dir ±1), with or without a box.
func (s ContainerPort) truck(g *Grid, x, z float64, dir int32, cab Color, tall bool, cells int, load Color) {
	span := g.Span()
	c := s.even(g, V(x-7*span, containerPortGround+span, z-span))
	front, nose := c.X+32, c.X+36
	if dir < 0 {
		front, nose = c.X-8, c.X-8
	}
	s.box(g, Cell{min(c.X, front), c.Y, c.Z}, 19, 1, 3, Charcoal)
	if cells > 0 {
		rear := c
		if dir < 0 {
			rear.X += int32(2 * (15 - cells))
		}
		s.box(g, Cell{rear.X, c.Y + 2, c.Z}, cells, 3, 3, load)
	}
	h := 2
	if tall {
		h = 3
	}
	s.box(g, Cell{front, c.Y + 2, c.Z}, 3, h, 3, cab)
	for k := int32(0); k < 3; k++ {
		s.set(g, Cell{nose, c.Y + int32(2*h), c.Z + 2*k}, Glass)
	}
}

// straddle is a straddle carrier: two wheeled side frames, a top deck and a cab, with a
// twenty-foot box slung between the frames.
func (s ContainerPort) straddle(g *Grid, x, z float64, load Color) {
	span := g.Span()
	c := s.even(g, V(x-4*span, containerPortGround+span, z-2*span))
	for _, dz := range []int32{0, 8} {
		s.box(g, Cell{c.X, c.Y, c.Z + dz}, 9, 1, 1, Charcoal)
		for _, dx := range []int32{2, 14} {
			s.box(g, Cell{c.X + dx, c.Y + 2, c.Z + dz}, 1, 7, 1, Gold)
		}
		s.box(g, Cell{c.X, c.Y + 16, c.Z + dz}, 9, 1, 1, Gold)
	}
	for _, dx := range []int32{2, 8, 14} {
		s.box(g, Cell{c.X + dx, c.Y + 16, c.Z}, 1, 1, 5, Gold)
	}
	s.box(g, Cell{c.X + 14, c.Y + 18, c.Z}, 2, 2, 2, White)
	s.set(g, Cell{c.X + 16, c.Y + 20, c.Z}, Glass)
	s.set(g, Cell{c.X + 16, c.Y + 20, c.Z + 2}, Glass)
	s.box(g, Cell{c.X + 2, c.Y + 6, c.Z + 2}, 7, 3, 3, load)
	s.box(g, Cell{c.X + 6, c.Y + 12, c.Z + 4}, 3, 2, 1, Charcoal)
}

// traffic parks tractors, road trucks and straddle carriers around the terminal.
func (s ContainerPort) traffic(g *Grid) {
	// Under the quay cranes.
	s.truck(g, -7.94, 4, 1, White, false, 0, 0)
	s.truck(g, 19.05, 4, 1, Gold, false, 15, Navy)
	s.truck(g, -30, 4, 1, White, false, 15, Foliage)
	s.truck(g, -40, -0.8, -1, Gold, false, 15, Red)
	s.truck(g, 8, -0.8, -1, White, false, 0, 0)
	s.truck(g, 42, 4, 1, Gold, false, 15, Slate)
	// In the gantry lane and on the yard roads.
	s.truck(g, -35.6, containerPortLane, 1, White, false, 0, 0)
	s.truck(g, -10.2, containerPortLane, -1, Red, true, 0, 0)
	s.truck(g, 21.5, containerPortLane, 1, Navy, true, 15, Gold)
	s.truck(g, 44, -9.2, -1, Foliage, true, 15, Brick)
	s.truck(g, -45, -34.2, 1, Red, true, 15, White)
	s.truck(g, 14, -34.2, -1, White, true, 7, Terracotta)
	s.straddle(g, -23, -3.8, Red)
	s.straddle(g, 38, -3.8, Navy)
	s.straddle(g, -1, -40.6, Gold)
}

// masts raises the floodlight towers along the apron and the yard road.
func (s ContainerPort) masts(g *Grid) {
	floor := containerPortGround + g.Span()/2
	for _, q := range []Vec2{P(-56, -9.9), P(-23, -9.9), P(5.5, -9.9), P(34, -9.9), P(50.5, -9.9),
		P(-30.6, -34.4), P(0.8, -34.4), P(28.2, -34.4)} {
		s.put(g, Stone, Capsule{q.At(floor), q.At(floor + 26), 0.5})
		s.put(g, Charcoal, BlockAt(q, 1.4, 3.6, floor+26, floor+26.8))
		s.put(g, White, BlockAt(q, 1.4, 3.6, floor+25.2, floor+26))
		s.put(g, Stone, BlockAt(q, 1.2, 1.2, floor, floor+1.2))
	}
}
