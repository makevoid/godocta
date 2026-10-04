package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// KarnakTemple is the Precinct of Amun-Re at Karnak, Luxor, laid out along its
// east-west axis. From the Nile quay in the west, an avenue of ram-headed sphinxes
// leads to the unfinished First Pylon (with the remains of its mud-brick building
// ramp behind the south tower). The Great Court holds the single standing column of
// Taharqa's kiosk, the shrine of Seti II, the temple of Ramesses III and two
// red-granite colossi of Ramesses II before the Second Pylon. The Great Hypostyle Hall
// follows: a forest of closed-bud columns flanking a nave of twelve taller open-papyrus
// columns, architraves, clerestory grilles, surviving roof slabs and painted bands. East
// of the ruined Third Pylon stand the granite obelisks of Thutmose I and Hatshepsut,
// then the low later pylons, the granite barque shrine and the Akh-menu festival hall.
// The south axis runs past the Cachette court to the Seventh Pylon; the stepped Sacred
// Lake lies south-east with its scarab and the fallen obelisk tip. Palms, scattered
// blocks and fragments of the mud-brick enclosure sit on desert sand that slopes to a
// low rim.
//
// Recipes use (u, v): u runs east along the axis and v north. In world space east is
// -X and north is +Z, so the entrance faces +X.
type KarnakTemple struct{}

func init() { Register(KarnakTemple{}) }

func (KarnakTemple) Name() string  { return "karnak_temple" }
func (KarnakTemple) Title() string { return "Karnak Temple" }

// karnakTempleG is the paved ground level of the precinct.
const karnakTempleG = 1.6

// Lake extents in (u, v).
const (
	karnakTempleLakeU0, karnakTempleLakeU1 = 38.5, 57.0
	karnakTempleLakeV0, karnakTempleLakeV1 = -32.0, -19.0
)

func karnakTempleNoise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

func karnakTempleFloor(x float64) int { return int(math.Floor(x)) }

// at maps a plan point (u, v) to world plan coordinates.
func (KarnakTemple) at(u, v float64) Vec2 { return P(-u, v) }

// blk is an axis-aligned block given in (u, v, y).
func (KarnakTemple) blk(u0, u1, v0, v1, y0, y1 float64) Cuboid {
	return Block(V(-u0, y0, v0), V(-u1, y1, v1))
}

// fn wraps a predicate over (u, v, y) into a bounded world solid.
func (KarnakTemple) fn(u0, u1, v0, v1, y0, y1 float64, f func(u, v, y float64) bool) Func {
	return Func{BoxOf(V(-u0, y0, v0), V(-u1, y1, v1)), func(p Vec3) bool { return f(-p.X, p.Z, p.Y) }}
}

// shade fills a colored shape: f reports occupancy and color for each point.
func (k KarnakTemple) shade(g *Grid, u0, u1, v0, v1, y0, y1 float64, colors []Color, f func(u, v, y float64) (bool, Color)) {
	for _, c := range colors {
		c := c
		g.Use(c).Fill(k.fn(u0, u1, v0, v1, y0, y1, func(u, v, y float64) bool {
			in, col := f(u, v, y)
			return in && col == c
		}))
	}
}

// masonry picks a weathered sandstone block color.
func (KarnakTemple) masonry(seed, course, j, face int, low bool) Color {
	h := Hash01(seed, course, j, face)
	switch {
	case h < 0.03 || (low && h < 0.15):
		return DarkStone
	case h < 0.14:
		return Sand
	}
	return Stone
}

// height is the ground surface at (u, v); negative means no ground.
func (KarnakTemple) height(u, v float64) float64 {
	const g = karnakTempleG
	if u >= karnakTempleLakeU0 && u <= karnakTempleLakeU1 && v >= karnakTempleLakeV0 && v <= karnakTempleLakeV1 {
		e := math.Min(math.Min(u-karnakTempleLakeU0, karnakTempleLakeU1-u), math.Min(v-karnakTempleLakeV0, karnakTempleLakeV1-v))
		switch {
		case e < 1:
			return g - 0.4
		case e < 2:
			return g - 0.8
		}
		return 0.1
	}
	// The precinct is a union of rectangles hugging the monuments.
	d := math.Inf(1)
	for _, r := range [][4]float64{{-65, -38, -15, 15}, {-38, 27, -31.5, 30.5}, {27, 69, -35, 13}} {
		du := math.Max(r[0]-u, u-r[1])
		dv := math.Max(r[2]-v, v-r[3])
		var e float64
		if du > 0 || dv > 0 {
			e = math.Hypot(math.Max(du, 0), math.Max(dv, 0))
		} else {
			e = math.Max(du, dv)
		}
		d = math.Min(d, e)
	}
	d += 2.2 * (karnakTempleNoise(u/8, v/8, 3) - 0.5)
	const run = 5.5
	switch {
	case d <= 0:
		return g
	case d < run:
		t := d / run
		s := t * t * (3 - 2*t)
		return g - (g-0.35)*s + 0.25*(karnakTempleNoise(u/3, v/3, 5)-0.5)*(1-t)
	}
	return -1
}

func (k KarnakTemple) ground(g *Grid) {
	const u0, u1, v0, v1 = -76.0, 82.0, -46.0, 40.0
	low := func(u, v float64) float64 {
		lo := math.Inf(1)
		for _, o := range [][2]float64{{0, 0}, {0.8, 0}, {-0.8, 0}, {0, 0.8}, {0, -0.8}} {
			lo = math.Min(lo, math.Max(0, k.height(u+o[0], v+o[1])))
		}
		return lo
	}
	g.Use(Sand).Fill(k.fn(u0, u1, v0, v1, 0, karnakTempleG+0.1, func(u, v, y float64) bool {
		h := k.height(u, v)
		if h < 0 || y > h {
			return false
		}
		return y >= math.Max(0, low(u, v)-0.82)
	}))
	// Water in the lake basin.
	g.Use(Water).Fill(k.blk(karnakTempleLakeU0+2, karnakTempleLakeU1-2, karnakTempleLakeV0+2, karnakTempleLakeV1-2, 0, 0.6))
	g.Paint(Stone, k.fn(karnakTempleLakeU0-1.5, karnakTempleLakeU1+1.5, karnakTempleLakeV0-1.5, karnakTempleLakeV1+1.5, 0.15, karnakTempleG+0.1,
		func(u, v, y float64) bool {
			in := u >= karnakTempleLakeU0+2 && u <= karnakTempleLakeU1-2 && v >= karnakTempleLakeV0+2 && v <= karnakTempleLakeV1-2
			return !in
		}))
	// Desert speckle and sand drifts.
	g.Paint(Stone, k.fn(u0, u1, v0, v1, 0, karnakTempleG+0.1, func(u, v, y float64) bool {
		return Hash01(karnakTempleFloor(u*1.2), karnakTempleFloor(v*1.2), 17) < 0.05
	}))
	g.Paint(DarkStone, k.fn(u0, u1, v0, v1, 0, karnakTempleG+0.1, func(u, v, y float64) bool {
		return Hash01(karnakTempleFloor(u*1.2), karnakTempleFloor(v*1.2), 19) < 0.015
	}))
	// Processional paving along the axis and the south axis.
	paved := func(u, v float64) bool {
		switch {
		case u >= -64 && u <= 66 && math.Abs(v) <= 2.4:
			return true
		case u >= -63 && u <= -38 && math.Abs(v) <= 3.4:
			return true
		case u >= 28.5 && u <= 32.5 && v <= 0 && v >= -36:
			return true
		}
		return false
	}
	g.Paint(Stone, k.fn(-70, 74, -40, 4, karnakTempleG-0.5, karnakTempleG+0.1, func(u, v, y float64) bool {
		return paved(u, v) && k.height(u, v) >= karnakTempleG-0.01
	}))
	g.Paint(DarkStone, k.fn(-70, 74, -40, 4, karnakTempleG-0.5, karnakTempleG+0.1, func(u, v, y float64) bool {
		return paved(u, v) && Hash01(karnakTempleFloor(u/1.6), karnakTempleFloor(v/1.6), 23) < 0.12
	}))
	// The hypostyle hall floor.
	g.Paint(DarkStone, k.blk(-5, 23, -23.5, 23.5, karnakTempleG-0.5, karnakTempleG+0.1))
	g.Paint(Stone, k.fn(-5, 23, -23.5, 23.5, karnakTempleG-0.5, karnakTempleG+0.1, func(u, v, y float64) bool {
		return Hash01(karnakTempleFloor(u/1.6), karnakTempleFloor(v/1.6), 29) < 0.4
	}))
}

// karnakTemplePylon is a battered pylon: two hollow towers flanking a gateway. It is
// described in (a, b): a across its thickness, b along its width.
type karnakTemplePylon struct {
	a0, a1     float64 // Base extent across the thickness; the front faces a0.
	b, gw      float64 // Half width and gate half width.
	hNeg, hPos float64 // Tower heights on the b < 0 and b > 0 sides.
	gh, dw, dh float64 // Gate block height, doorway half width and height.
	batter     float64 // Inset per unit height.
	ruin       float64 // Depth of the broken top.
	cornice    bool
	slots      int // Flagpole niches per tower on the front face.
	seed       int
	south      bool    // Faces north-south: a = -v, b = u - center.
	center     float64 // Plan offset along b.
}

func (p karnakTemplePylon) uv(a, b float64) (float64, float64) {
	if p.south {
		return b + p.center, -a
	}
	return a, b + p.center
}

func (p karnakTemplePylon) ab(u, v float64) (float64, float64) {
	if p.south {
		return -v, u - p.center
	}
	return u, v - p.center
}

// shape reports occupancy and color at (a, b) and height yy above the ground.
func (p karnakTemplePylon) shape(k KarnakTemple, a, b, yy float64) (bool, Color) {
	if yy < 0 {
		return false, 0
	}
	if in, c := p.gate(k, a, b, yy); in {
		return true, c
	}
	ab := math.Abs(b)
	if ab < p.gw-0.2 {
		return false, 0
	}
	side := 1
	h := p.hPos
	if b < 0 {
		side, h = -1, p.hNeg
	}
	top := h
	intact := true
	if p.ruin > 0 {
		r := Hash01(p.seed, karnakTempleFloor(a/2.2), karnakTempleFloor(ab/2.2), side)
		// Ruins fall away toward the outer ends.
		edge := math.Max(0, (ab-p.gw)/(p.b-p.gw)-0.55) * 1.6
		top = h - p.ruin*(r*r+edge*r)
		intact = r < 0.25 && edge == 0
	}
	if yy > top+0.01 {
		return false, 0
	}
	ins := p.batter * yy
	flare := 0.0
	if p.cornice && intact && yy > top-1.4 {
		flare = 0.6 * (yy - (top - 1.4)) / 1.4
	}
	aW, aE := p.a0+ins-flare, p.a1-ins+flare
	bO, bI := p.b-ins+flare, p.gw+ins-flare
	if a < aW || a > aE || ab > bO || ab < bI {
		return false, 0
	}
	dW := a - aW
	inSlot := false
	for s := 0; s < p.slots; s++ {
		c := p.gw + 2.6 + float64(s)*4.2
		if math.Abs(ab-c) < 0.65 && yy > 0.4 {
			inSlot = true
		}
	}
	if inSlot {
		if dW < 0.8 {
			return false, 0
		}
		dW -= 0.8
	}
	d := math.Min(math.Min(dW, aE-a), math.Min(bO-ab, ab-bI))
	if d > 1.3 && yy < top-1.2 {
		return false, 0
	}
	if inSlot && dW < 0.7 {
		return true, DarkStone
	}
	// Torus moulding on the vertical corners.
	near := 0
	for _, e := range []float64{a - aW, aE - a, bO - ab} {
		if e < 0.7 {
			near++
		}
	}
	if near >= 2 && yy < top-0.4 {
		return true, Sand
	}
	course := karnakTempleFloor(yy / 0.8)
	face, along := 0, ab
	switch {
	case yy > top-1.0 && d > 0.9:
		face, along = 4, a
	case d == aE-a:
		face = 1
	case d == bO-ab || d == ab-bI:
		face, along = 2, a
	}
	j := karnakTempleFloor((along + float64(course%2)*1.2) / 2.4)
	return true, k.masonry(p.seed+side, course, j, face, yy < 1.8)
}

// gate is the portal block between the towers.
func (p karnakTemplePylon) gate(k KarnakTemple, a, b, yy float64) (bool, Color) {
	ab := math.Abs(b)
	if ab > p.gw+0.6 || yy > p.gh {
		return false, 0
	}
	flare := 0.0
	if p.cornice && yy > p.gh-1.4 {
		flare = 0.55 * (yy - (p.gh - 1.4)) / 1.4
	}
	aW, aE := p.a0+0.9-flare, p.a1-0.9+flare
	if a < aW || a > aE {
		return false, 0
	}
	if ab < p.dw && yy < p.dh {
		return false, 0
	}
	if ab < p.dw+0.7 && yy < p.dh+0.6 {
		return true, DarkStone
	}
	dFace := math.Min(a-aW, aE-a)
	if dFace > 1.3 && yy > p.dh+1.0 && yy < p.gh-1.2 && ab < p.gw-0.4 {
		return false, 0
	}
	if dFace < 0.6 {
		// Winged sun disk over the doorway.
		ys := p.dh + 1.5
		if ab*ab+(yy-ys)*(yy-ys) < 0.9*0.9 {
			return true, Gold
		}
		if ab < 3.4 && yy > ys-0.45 && yy < ys+0.4 {
			if ab > 2.5 {
				return true, Red
			}
			return true, Navy
		}
	}
	if p.cornice && yy > p.gh-1.0 && dFace < 0.6 {
		return true, Sand
	}
	course := karnakTempleFloor(yy / 0.8)
	j := karnakTempleFloor((ab + float64(course%2)*1.2) / 2.4)
	return true, k.masonry(p.seed+5, course, j, 0, yy < 1.8)
}

func (k KarnakTemple) pylon(g *Grid, p karnakTemplePylon) {
	h := math.Max(math.Max(p.hNeg, p.hPos), p.gh) + 1
	u0, v0 := p.uv(p.a0-1, -p.b-1)
	u1, v1 := p.uv(p.a1+1, p.b+1)
	colors := []Color{Stone, Sand, DarkStone, Gold, Navy, Red}
	k.shade(g, math.Min(u0, u1), math.Max(u0, u1), math.Min(v0, v1), math.Max(v0, v1), karnakTempleG, karnakTempleG+h, colors,
		func(u, v, y float64) (bool, Color) {
			a, b := p.ab(u, v)
			return p.shape(k, a, b, y-karnakTempleG)
		})
}

// column builds a column with base, drums, capital and abacus and returns its top.
func (k KarnakTemple) column(g *Grid, u, v, y0, r, hs float64, open bool, seed int, bands [][3]float64) float64 {
	c := k.at(u, v)
	g.Use(Stone).Fill(Cylinder{c, r + 0.3, 0, y0, y0 + 0.45})
	for y := y0 + 0.45; y < y0+hs; y += 0.8 {
		col := Sand
		if h := Hash01(seed, karnakTempleFloor(y*4)); h < 0.12 {
			col = Stone
		}
		g.Use(col).Fill(Cylinder{c, r, 0, y, math.Min(y+0.8, y0+hs)})
	}
	for _, b := range bands {
		g.Paint(Color(b[2]), Cylinder{c, r + 0.1, 0, y0 + b[0], y0 + b[1]})
	}
	y := y0 + hs
	if open {
		g.Use(Sand).Fill(Frustum{c, r, r * 1.65, y, y + 1.8})
		g.Paint(Foliage, Frustum{c, r + 0.2, r*1.2 + 0.2, y, y + 0.5})
		y += 1.8
		g.Use(Stone).Fill(BlockAt(c, r*1.7, r*1.7, y, y+0.6))
		return y + 0.6
	}
	g.Use(Sand).Fill(Frustum{c, r, r * 1.3, y, y + 0.5}, Frustum{c, r * 1.3, r * 0.75, y + 0.5, y + 1.4})
	y += 1.4
	g.Use(Stone).Fill(BlockAt(c, r*1.9, r*1.9, y, y+0.5))
	return y + 0.5
}

// sphinx is a ram-headed sphinx lying across the avenue, its head toward the axis.
func (k KarnakTemple) sphinx(g *Grid, u, side float64, seed int) {
	const G = karnakTempleG
	vv := func(d float64) float64 { return side * d }
	g.Use(Stone).Fill(k.blk(u-0.75, u+0.75, vv(3.1), vv(7.6), G, G+0.9))
	g.Paint(DarkStone, k.blk(u-0.75, u+0.75, vv(3.1), vv(7.6), G+0.6, G+0.9))
	g.Use(Sand).Fill(
		Ellipsoid{k.at(u, vv(6.0)).At(G + 1.55), V(0.68, 0.7, 1.6)},
		Ellipsoid{k.at(u, vv(4.8)).At(G + 1.8), V(0.7, 0.85, 0.9)},
		k.blk(u-0.6, u+0.6, vv(3.2), vv(4.6), G+0.9, G+1.35),
		Ellipsoid{k.at(u, vv(4.0)).At(G + 2.85), V(0.52, 0.6, 0.8)},
		k.blk(u-0.62, u+0.62, vv(4.4), vv(5.0), G+1.9, G+2.8),
	)
	g.Use(DarkStone).Fill(
		Capsule{k.at(u-0.5, vv(4.4)).At(G + 3.1), k.at(u-0.62, vv(4.0)).At(G + 2.5), 0.32},
		Capsule{k.at(u+0.5, vv(4.4)).At(G + 3.1), k.at(u+0.62, vv(4.0)).At(G + 2.5), 0.32},
	)
	if Hash01(seed, 3) < 0.25 {
		// Some heads are lost.
		g.Carve(k.blk(u-0.8, u+0.8, vv(3.1), vv(4.5), G+2.2, G+3.6))
	}
}

// colossus is a striding red-granite statue facing west (+X) with the double crown.
func (k KarnakTemple) colossus(g *Grid, u, v, s float64) {
	const G = karnakTempleG
	c := k.at(u, v)
	x, z := c.X, c.Y
	g.Use(Stone).Fill(Block(V(x-1.7, G, z-1.3), V(x+1.7, G+1.6, z+1.3)))
	g.Paint(Sand, Block(V(x-1.7, G+1.2, z-1.3), V(x+1.7, G+1.6, z+1.3)))
	y := G + 1.6
	at := func(dx, dy, dz float64) Vec3 { return V(x+dx*s, y+dy*s, z+dz*s) }
	g.Use(Brick).Fill(
		Block(at(-1.0, 0, -0.75), at(-0.4, 6.6, 0.75)),
		Block(at(0.15, 0, -0.9), at(1.1, 3.4, -0.05)),
		Block(at(-0.6, 0, 0.05), at(0.35, 3.4, 0.9)),
		Ellipsoid{at(0.05, 5.3, 0), V(0.75*s, 1.25*s, 1.2*s)},
		Capsule{at(-0.1, 6.1, -1.2), at(0.1, 3.8, -1.2), 0.45 * s},
		Capsule{at(-0.1, 6.1, 1.2), at(0.1, 3.8, 1.2), 0.45 * s},
		Ellipsoid{at(0.2, 6.95, 0), V(0.6*s, 0.65*s, 0.55*s)},
		Block(at(0.6, 5.9, -0.22), at(0.9, 6.5, 0.22)),
	)
	g.Use(White).Fill(Block(at(-0.6, 2.8, -1.05), at(0.75, 4.2, 1.05)))
	g.Use(Gold).Fill(Block(at(-0.6, 4.0, -1.1), at(0.75, 4.4, 1.1)))
	// Striped nemes headdress.
	g.Use(Gold).Fill(Block(at(-0.4, 6.0, -0.95), at(0.2, 7.3, 0.95)), Ellipsoid{at(-0.05, 7.15, 0), V(0.6*s, 0.55*s, 0.75*s)})
	for dy := 6.0; dy < 7.8; dy += 0.8 {
		g.Paint(Navy, Block(at(-0.7, dy, -1.0), at(0.3, dy+0.4, 1.0)))
	}
	// Double crown: red deshret and white hedjet.
	g.Use(Red).Fill(Cylinder{P(x-0.05*s, z), 0.6 * s, 0, y + 7.55*s, y + 8.15*s})
	g.Use(White).Fill(Frustum{P(x-0.05*s, z), 0.45 * s, 0.2 * s, y + 8.15*s, y + 9.4*s}, Sphere(at(-0.05, 9.4, 0), 0.32*s))
}

// obelisk is a tapering red-granite shaft on a pedestal with a gilded pyramidion.
func (k KarnakTemple) obelisk(g *Grid, u, v, h, w float64) {
	const G = karnakTempleG
	c := k.at(u, v)
	g.Use(Stone).Fill(BlockAt(c, w+1.6, w+1.6, G, G+1.0))
	y0 := G + 1.0
	half := func(y float64) float64 { return w / 2 * (1 - 0.32*(y-y0)/h) }
	box := Box{V(c.X-w, y0, c.Y-w), V(c.X+w, y0+h, c.Y+w)}
	g.Use(Brick).Fill(Func{box, func(p Vec3) bool {
		hw := half(p.Y)
		return math.Abs(p.X-c.X) <= hw && math.Abs(p.Z-c.Y) <= hw
	}})
	top := half(y0 + h)
	ph := top * 2.6
	g.Use(Gold).Fill(Func{Box{V(c.X-w, y0+h, c.Y-w), V(c.X+w, y0+h+ph, c.Y+w)}, func(p Vec3) bool {
		hw := top * (1 - (p.Y-y0-h)/ph)
		return math.Abs(p.X-c.X) <= hw+0.05 && math.Abs(p.Z-c.Y) <= hw+0.05
	}})
	// A column of inscription down the faces.
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		hw := half(p.Y)
		dx, dz := math.Abs(p.X-c.X), math.Abs(p.Z-c.Y)
		return p.Y < y0+h-0.8 && p.Y > y0+0.8 && math.Min(dx, dz) < 0.3 && math.Max(dx, dz) > hw-0.45 &&
			Hash01(karnakTempleFloor(p.Y*2.5), 41) < 0.5
	}})
}

// palm is a date palm with a leaning trunk and drooping fronds.
func (k KarnakTemple) palm(g *Grid, u, v, h float64, seed int) {
	base := k.at(u, v).At(karnakTempleG - 0.2)
	ang := Hash01(seed, 1) * 2 * math.Pi
	lean := P(math.Cos(ang), math.Sin(ang)).Scale(0.9 + 0.8*Hash01(seed, 2))
	pt := func(t float64) Vec3 {
		o := lean.Scale(t * t)
		return base.Add(V(o.X, t*h, o.Y))
	}
	g.Use(Wood)
	for i := 0; i < 4; i++ {
		t0, t1 := float64(i)/4, float64(i+1)/4
		g.Fill(Capsule{pt(t0), pt(t1), 0.5})
	}
	top := pt(1)
	g.Use(Foliage).Fill(Sphere(top.Add(V(0, 0.2, 0)), 0.7))
	n := 8
	for i := 0; i < n; i++ {
		th := float64(i)*2*math.Pi/float64(n) + Hash01(seed, 3)
		d := V(math.Cos(th), 0, math.Sin(th))
		m := top.Add(d.Scale(1.5)).Add(V(0, 0.55, 0))
		e := top.Add(d.Scale(3.0)).Add(V(0, -0.7-0.4*Hash01(seed, i, 4), 0))
		g.Fill(Capsule{top, m, 0.42}, Capsule{m, e, 0.4})
	}
	g.Use(Gold).Fill(Sphere(top.Add(V(0, -0.6, 0)), 0.45))
}

// blocks scatters fallen masonry blocks in a zone.
func (k KarnakTemple) blocks(g *Grid, u0, u1, v0, v1 float64, n, seed int) {
	const G = karnakTempleG
	for i := 0; i < n; i++ {
		u := u0 + (u1-u0)*Hash01(seed, i, 1)
		v := v0 + (v1-v0)*Hash01(seed, i, 2)
		l := 1.4 + 1.4*Hash01(seed, i, 3)
		w := 0.9 + 0.7*Hash01(seed, i, 4)
		h := 0.8
		if Hash01(seed, i, 5) < 0.3 {
			h = 1.6
		}
		col := Stone
		switch r := Hash01(seed, i, 6); {
		case r < 0.35:
			col = Sand
		case r < 0.5:
			col = DarkStone
		}
		b := Block(V(-l/2, G, -w/2), V(l/2, G+h, w/2))
		g.Use(col).Fill(Translate(RotateY(b, P(0, 0), Hash01(seed, i, 7)*math.Pi), k.at(u, v).At(0)))
	}
}

// ruinWall is a wall segment with a broken top.
func (k KarnakTemple) ruinWall(g *Grid, u0, u1, v0, v1, hMin, hMax float64, seed int, colors []Color) {
	const G = karnakTempleG
	alongU := math.Abs(u1-u0) >= math.Abs(v1-v0)
	k.shade(g, u0, u1, v0, v1, G, G+hMax, colors, func(u, v, y float64) (bool, Color) {
		t := v
		if alongU {
			t = u
		}
		j := karnakTempleFloor(t / 1.6)
		r := Hash01(seed, j)
		top := hMin + (hMax-hMin)*(1-r*r)
		if y-G > top {
			return false, 0
		}
		course := karnakTempleFloor((y - G) / 0.8)
		jj := karnakTempleFloor((t + float64(course%2)*0.8) / 1.6)
		return true, colors[int(Hash01(seed, course, jj)*float64(len(colors)))]
	})
}

func (k KarnakTemple) Build(g *Grid) Order {
	const G = karnakTempleG
	k.ground(g)

	// Quay on the old Nile channel with its small obelisks, then the ram avenue.
	g.Use(Stone).Fill(k.blk(-63.5, -59.3, -9, 9, G, G+0.8))
	g.Use(Stone).Fill(k.blk(-64.4, -63.5, -5, 5, G-0.4, G+0.4))
	g.Paint(DarkStone, k.blk(-63.5, -59.3, -9, 9, G+0.6, G+0.8))
	for _, v := range []float64{-6.5, 6.5} {
		c := k.at(-61.4, v)
		g.Use(Stone).Fill(BlockAt(c, 1.6, 1.6, G+0.8, G+1.6))
		g.Use(Brick).Fill(Func{Box{V(c.X-0.6, G+1.6, c.Y-0.6), V(c.X+0.6, G+6.2, c.Y+0.6)}, func(p Vec3) bool {
			hw := 0.55 - 0.1*(p.Y-G-1.6)/4.6
			if p.Y > G+5.6 {
				hw = 0.45 * (G + 6.2 - p.Y) / 0.6
			}
			return math.Abs(p.X-c.X) <= hw+0.05 && math.Abs(p.Z-c.Y) <= hw+0.05
		}})
		g.Paint(Gold, Block(V(c.X-0.6, G+5.6, c.Y-0.6), V(c.X+0.6, G+6.2, c.Y+0.6)))
	}
	for i := 0; i < 10; i++ {
		u := -57.5 + 1.9*float64(i)
		k.sphinx(g, u, 1, i)
		k.sphinx(g, u, -1, i+50)
	}

	// First Pylon: unfinished, rough topped, with flagpole niches.
	k.pylon(g, karnakTemplePylon{a0: -38, a1: -32, b: 28, gw: 4.8, hNeg: 18, hPos: 20.5, gh: 14, dw: 2.4, dh: 10.5,
		batter: 0.09, ruin: 1.6, slots: 2, seed: 100})
	// The mud-brick construction ramp left behind the south tower.
	k.shade(g, -32.5, -22, -22, -12, G, G+14, []Color{Wood, DarkStone, Terracotta}, func(u, v, y float64) (bool, Color) {
		t := (u + 32.5) / 10.5
		top := 13.5*(1-t) - 1.6*Hash01(karnakTempleFloor(u/1.6), karnakTempleFloor(v/1.6), 7)*t
		top = math.Floor(top/0.8) * 0.8
		yy := y - G
		if yy > top {
			return false, 0
		}
		edge := math.Min(v+22, -12-v)
		if yy < top-1.4 && edge > 1.3 {
			return false, 0
		}
		r := Hash01(karnakTempleFloor(u/1.2), karnakTempleFloor(yy/0.8), karnakTempleFloor(v/1.2), 9)
		switch {
		case r < 0.4:
			return true, DarkStone
		case r < 0.47:
			return true, Terracotta
		}
		return true, Wood
	})

	// Great Court walls.
	stone := []Color{Stone, Stone, Sand, DarkStone}
	k.ruinWall(g, -32, -10, 24.5, 25.7, 6.5, 8, 201, stone)
	k.ruinWall(g, -32, -21.5, -25.7, -24.5, 5, 8, 202, stone)
	// North colonnade with architrave and some roof slabs.
	for i := 0; i < 8; i++ {
		u := -29.5 + 2.6*float64(i)
		top := k.column(g, u, 21.5, G, 0.75, 5.4, false, 300+i, nil)
		_ = top
	}
	g.Use(Stone).Fill(k.blk(-30.4, -10.8, 20.9, 22.1, G+7.3, G+8.1))
	for i := 0; i < 7; i++ {
		if Hash01(i, 310) < 0.6 {
			u := -29.5 + 2.6*float64(i)
			g.Use(Stone).Fill(k.blk(u, u+2.6, 21.0, 24.5, G+8.1, G+8.7))
		}
	}
	// Taharqa's kiosk: one standing column among stumps, and the barque altar.
	for i := 0; i < 5; i++ {
		for _, v := range []float64{-2.8, 2.8} {
			u := -27 + 3.0*float64(i)
			if i == 3 && v > 0 {
				top := k.column(g, u, v, G, 1.1, 9.6, true, 400, [][3]float64{{4.0, 4.8, float64(Red)}, {6.4, 7.2, float64(Navy)}})
				_ = top
				continue
			}
			c := k.at(u, v)
			g.Use(Stone).Fill(Cylinder{c, 1.4, 0, G, G + 0.45})
			h := 0.4 + 1.6*Hash01(i, int(v), 401)
			g.Use(Sand).Fill(Cylinder{c, 1.1, 0, G + 0.45, G + 0.45 + h})
		}
	}
	g.Use(White).Fill(k.blk(-21, -17.5, -1.2, 1.2, G, G+1.2))
	// Shrine of Seti II: three chapels opening south onto the axis.
	k.shade(g, -31, -24.5, 10, 16.5, G, G+6, []Color{Stone, Sand, Charcoal, DarkStone}, func(u, v, y float64) (bool, Color) {
		yy := y - G
		in := u >= -30.6 && u <= -25 && v >= 10.5 && v <= 16 && yy <= 4.6
		cor := yy > 4.6 && yy <= 5.4 && u >= -31 && u <= -24.6 && v >= 10.1 && v <= 16.4
		if !in && !cor {
			return false, 0
		}
		if cor {
			return true, Sand
		}
		for _, du := range []float64{-29.2, -27.8, -26.4} {
			if math.Abs(u-du) < 0.5 && v < 11.3 && yy < 3.0 {
				return true, Charcoal
			}
		}
		return true, k.masonry(500, karnakTempleFloor(yy/0.8), karnakTempleFloor(u/1.6), 0, yy < 1.2)
	})
	g.Carve(
		k.blk(-29.6, -28.8, 10.3, 10.9, G+0.01, G+2.9),
		k.blk(-28.2, -27.4, 10.3, 10.9, G+0.01, G+2.9),
		k.blk(-26.8, -26.0, 10.3, 10.9, G+0.01, G+2.9),
	)

	// Temple of Ramesses III, entered from the court through its own pylon.
	k.pylon(g, karnakTemplePylon{a0: 12.5, a1: 14.6, b: 5.2, gw: 1.6, hNeg: 8, hPos: 8, gh: 6.5, dw: 1.0, dh: 4.5,
		batter: 0.06, cornice: true, seed: 600, south: true, center: -16})
	k.shade(g, -21, -11, -30.5, -14.6, G, G+6.2, []Color{Stone, Sand, DarkStone, Charcoal, Red, Navy}, func(u, v, y float64) (bool, Color) {
		yy := y - G
		dWall := math.Min(math.Min(u+21, -11-u), math.Min(-14.6-v, v+30.5))
		if dWall < 0 {
			return false, 0
		}
		roofed := v < -23
		if roofed {
			if yy > 5.2 && yy <= 6.0 {
				return true, Stone
			}
			if yy > 5.2 {
				return false, 0
			}
			if dWall > 1.0 {
				return true, Charcoal
			}
		} else {
			if yy > 5.2 {
				return false, 0
			}
			if dWall > 1.0 {
				// Osiride pillars line the open court.
				pu := math.Min(math.Abs(u+19.2), math.Abs(u+12.8))
				if pu < 0.5 && math.Mod(math.Abs(v+15.4), 2.2) < 1.0 && v < -15.5 {
					if yy > 3.0 && yy < 3.8 {
						return true, Red
					}
					if yy > 3.8 && yy < 4.6 {
						return true, Navy
					}
					return true, Stone
				}
				if yy > 4.4 && yy <= 5.2 && pu < 1.2 && v < -15.5 {
					return true, Stone
				}
				return false, 0
			}
		}
		if v > -24.2 && v < -22.8 && math.Abs(u+16) < 1.2 && yy < 3.8 {
			return true, Charcoal
		}
		return true, k.masonry(610, karnakTempleFloor(yy/0.8), karnakTempleFloor((u+v)/1.6), 0, yy < 1.2)
	})

	// Colossi of Ramesses II before the Second Pylon.
	k.colossus(g, -13.2, 7.2, 1.0)
	k.colossus(g, -13.2, -7.2, 0.85)

	// Second Pylon.
	k.pylon(g, karnakTemplePylon{a0: -10, a1: -5, b: 25, gw: 4.4, hNeg: 16, hPos: 15, gh: 14.5, dw: 2.2, dh: 9,
		batter: 0.08, ruin: 3, cornice: true, slots: 2, seed: 700})

	// Great Hypostyle Hall.
	k.hypostyle(g)

	// Third Pylon.
	k.pylon(g, karnakTemplePylon{a0: 23, a1: 27, b: 24.5, gw: 3.8, hNeg: 11, hPos: 11.5, gh: 9.5, dw: 1.8, dh: 6.5,
		batter: 0.08, ruin: 3.5, cornice: true, seed: 800})
	// Obelisks of Thutmose I and Hatshepsut, each beside the stump of its lost twin.
	k.obelisk(g, 29.1, -3.8, 14, 1.9)
	g.Use(Stone).Fill(BlockAt(k.at(29.1, 3.8), 3.4, 3.4, G, G+1.0))
	g.Use(Brick).Fill(BlockAt(k.at(29.1, 3.8), 1.9, 1.9, G+1.0, G+2.2))
	k.pylon(g, karnakTemplePylon{a0: 31.2, a1: 33.4, b: 14, gw: 2.6, hNeg: 6, hPos: 6.5, gh: 5.5, dw: 1.4, dh: 3.8,
		batter: 0.06, ruin: 2.5, seed: 900})
	k.obelisk(g, 35.1, 3.8, 18, 2.1)
	g.Use(Stone).Fill(BlockAt(k.at(35.1, -3.8), 3.6, 3.6, G, G+1.0))
	g.Use(Brick).Fill(BlockAt(k.at(35.1, -3.8), 2.1, 2.1, G+1.0, G+1.8))
	k.pylon(g, karnakTemplePylon{a0: 36.8, a1: 38.6, b: 12, gw: 2.4, hNeg: 4.5, hPos: 4, gh: 4, dw: 1.2, dh: 3,
		batter: 0.05, ruin: 2, seed: 1000})
	k.pylon(g, karnakTemplePylon{a0: 39.4, a1: 41, b: 10, gw: 2.2, hNeg: 3.5, hPos: 4, gh: 4, dw: 1.2, dh: 3,
		batter: 0.05, ruin: 2, seed: 1100})
	// Heraldic pillars and the granite barque shrine.
	for _, v := range []float64{-2.1, 2.1} {
		c := k.at(42, v)
		g.Use(Brick).Fill(BlockAt(c, 1.1, 1.1, G, G+5.6))
		g.Paint(Gold, BlockAt(c, 1.2, 1.2, G+4.8, G+5.6))
	}
	const su = 43.2 // West face of the barque shrine.
	k.shade(g, su-0.1, su+5.1, -2.8, 2.8, G, G+6, []Color{DarkStone, Brick, Charcoal, Navy, Gold}, func(u, v, y float64) (bool, Color) {
		yy := y - G
		if u < su || u > su+5 || math.Abs(v) > 2.7 || yy > 5.6 {
			return false, 0
		}
		d := math.Min(math.Min(u-su, su+5-u), 2.7-math.Abs(v))
		if yy > 4.8 {
			if math.Mod(u, 1.6) < 0.8 && d < 0.6 {
				return true, Gold
			}
			return true, Navy
		}
		if d > 0.9 || (u < su+1 && math.Abs(v) < 0.9 && yy < 3.4) {
			return true, Charcoal
		}
		if Hash01(karnakTempleFloor(u/1.2), karnakTempleFloor(yy/0.8), karnakTempleFloor(v/1.2), 1200) < 0.4 {
			return true, Brick
		}
		return true, DarkStone
	})
	g.Carve(k.blk(su-0.2, su+0.5, -0.8, 0.8, G+0.01, G+3.3))
	k.ruinWall(g, su-0.6, su+6.6, 3.4, 4.4, 1.0, 3.5, 1210, stone)
	k.ruinWall(g, su-0.6, su+6.6, -4.4, -3.4, 1.0, 3.5, 1211, stone)
	k.ruinWall(g, su-0.6, su+6.6, 8.2, 9.2, 0.8, 3.0, 1212, stone)
	k.ruinWall(g, su-0.6, su+6.6, -9.2, -8.2, 0.8, 3.0, 1213, stone)
	for i, u := range []float64{su + 1.8, su + 4.6} {
		k.ruinWall(g, u, u+1, 4.4, 8.2, 0.8, 2.8, 1214+i, stone)
		k.ruinWall(g, u, u+1, -8.2, -4.4, 0.8, 2.8, 1218+i, stone)
	}
	// Middle Kingdom court: foundation outlines.
	g.Use(Stone).Fill(
		k.blk(50.4, 53.6, 6.4, 7.4, G, G+0.8), k.blk(50.4, 53.6, -7.4, -6.4, G, G+0.8),
		k.blk(52.6, 53.6, -7.4, -2.4, G, G+0.8), k.blk(52.6, 53.6, 2.4, 7.4, G, G+0.8),
		k.blk(50.8, 52.2, -2, 2, G, G+0.6),
	)
	k.akhmenu(g)

	// South axis: Cachette court and the Seventh Pylon facing south.
	k.ruinWall(g, 24.6, 25.6, -30.5, -24.5, 3, 6, 1300, stone)
	k.ruinWall(g, 34.4, 35.4, -30.5, -14.2, 3, 6, 1301, stone)
	k.pylon(g, karnakTemplePylon{a0: 30.5, a1: 33.6, b: 8.5, gw: 2.4, hNeg: 8, hPos: 9.5, gh: 7.5, dw: 1.4, dh: 5.2,
		batter: 0.07, ruin: 2.5, seed: 1400, south: true, center: 30})

	// Sacred Lake details: scarab of Amenhotep III and Hatshepsut's fallen obelisk tip.
	sc := k.at(37.4, -17.2)
	g.Use(Stone).Fill(BlockAt(sc, 1.8, 1.8, G, G+1.2))
	g.Use(DarkStone).Fill(Ellipsoid{sc.At(G + 1.55), V(0.75, 0.55, 0.95)})
	g.Paint(Gold, Ellipsoid{sc.At(G + 1.95), V(0.4, 0.3, 0.5)})
	ob := k.at(41.5, -15.4)
	g.Use(Brick).Fill(Block(V(ob.X-4, G, ob.Y-0.8), V(ob.X+1, G+1.6, ob.Y+0.8)))
	g.Use(Gold).Fill(Func{Box{V(ob.X-5.8, G, ob.Y-0.9), V(ob.X-4, G+1.6, ob.Y+0.9)}, func(p Vec3) bool {
		hw := 0.8 * (p.X - (ob.X - 5.8)) / 1.8
		return math.Abs(p.Z-ob.Y) <= hw+0.05 && math.Abs(p.Y-(G+0.8)) <= hw+0.05
	}})

	// Fragments of the mud-brick enclosure.
	mud := []Color{Wood, DarkStone, Wood, Terracotta}
	for i := 0; i < 6; i++ {
		v0 := -22 + 8.0*float64(i)
		if Hash01(i, 1520) < 0.75 {
			k.ruinWall(g, 66.4, 67.8, v0, v0+4+3*Hash01(i, 1521), 0.8, 3.4, 1522+i, mud)
		}
	}

	// Fallen blocks and palms.
	k.blocks(g, -3, 22, 25.5, 27.5, 12, 1600)
	k.blocks(g, 0, 22, -29, -26, 10, 1601)
	k.blocks(g, 48, 63, -16.5, -12, 8, 1602)
	k.blocks(g, 36, 52, 10.5, 12.5, 5, 1603)
	palms := [][3]float64{
		{-62, 13, 7}, {-59, 17, 8}, {-55, 12.5, 6.5}, {-60, -13, 7.5}, {-56, -17, 6.8}, {-52, -13, 8},
		{60, -23, 7}, {61, -28, 8}, {63, -19, 6.5}, {56, 12, 7.5}, {62, 12.5, 6.8},
		{-16, 27.5, 7}, {12, -31, 7.2},
	}
	for i, p := range palms {
		k.palm(g, p[0], p[1], p[2], 1700+i)
	}
	return Along(P(-1, 0))
}

// hypostyle builds the Great Hypostyle Hall between the Second and Third Pylons.
func (k KarnakTemple) hypostyle(g *Grid) {
	const (
		G    = karnakTempleG
		wHs  = 7.0  // Wing shaft height (with base).
		wTop = 8.9  // Top of the wing abacus.
		nHs  = 10.6 // Nave shaft height.
		nTop = 13.0 // Top of the nave abacus.
	)
	wingRows := []float64{8.0, 11.6, 15.2, 18.8}
	cols := []float64{-2.6, 1.2, 5.0, 8.8, 12.6, 16.4, 20.2}
	naveCols := []float64{-2.4, 2.08, 6.56, 11.04, 15.52, 20.0}
	// Wing columns: closed papyrus bud capitals, some with painted bands.
	for s, side := range []float64{-1, 1} {
		for j, r := range wingRows {
			for i, u := range cols {
				var bands [][3]float64
				if j == 0 || Hash01(i, j, s, 2000) < 0.35 {
					bands = [][3]float64{{3.2, 4.0, float64(Red)}, {4.8, 5.6, float64(Navy)}, {5.6, 6.0, float64(Gold)}}
				}
				k.column(g, u, side*r, G, 0.95, wHs, false, 2000+s*100+j*10+i, bands)
			}
			// Architraves along the row; most of the outer ones have fallen.
			for i := 0; i+1 < len(cols); i++ {
				if j == 0 || Hash01(i, j, s, 2010) < 0.12 {
					g.Use(Sand).Fill(k.blk(cols[i]-0.6, cols[i+1]+0.6, side*r-0.6, side*r+0.6, G+wTop, G+wTop+0.8))
				}
			}
		}
	}
	// Nave: twelve taller open papyrus columns.
	for s, side := range []float64{-1, 1} {
		for i, u := range naveCols {
			k.column(g, u, side*3.6, G, 1.25, nHs, true, 2500+s*10+i, [][3]float64{
				{3.6, 4.4, float64(Red)}, {5.6, 6.8, float64(Navy)}, {7.6, 8.0, float64(Gold)}, {8.4, 9.2, float64(Red)},
			})
		}
		g.Use(Sand).Fill(k.blk(naveCols[0]-0.8, naveCols[len(naveCols)-1]+0.8, side*3.6-0.8, side*3.6+0.8, G+nTop, G+nTop+0.8))
	}
	// Clerestory walls with stone window grilles above the inner wing rows.
	r0 := wingRows[0]
	for _, side := range []float64{-1, 1} {
		k.shade(g, cols[0]-0.8, cols[len(cols)-1]+0.8, side*r0-0.6, side*r0+0.6, G+wTop+0.8, G+nTop+0.8, []Color{Stone, Sand, DarkStone},
			func(u, v, y float64) (bool, Color) {
				yy := y - G
				if yy < wTop+0.8 || yy > nTop+0.8 {
					return false, 0
				}
				if bay := karnakTempleFloor((u - cols[0]) / (cols[1] - cols[0])); side > 0 && Hash01(bay, 2610) < 0.5 {
					if yy > wTop+0.8+3.2*Hash01(bay, karnakTempleFloor(u/0.8), 2611) {
						return false, 0
					}
				}
				if yy > wTop+1.4 && yy < nTop+0.2 {
					for i := 0; i+1 < len(cols); i++ {
						m := (cols[i] + cols[i+1]) / 2
						d := math.Abs(u - m)
						if d < 1.45 && (d < 0.4 || d > 0.85) {
							return false, 0
						}
					}
				}
				course := karnakTempleFloor(yy / 0.8)
				if Hash01(course, karnakTempleFloor(u/1.6), 2600) < 0.25 {
					return true, Stone
				}
				return true, Sand
			})
	}
	// Roof slabs: part of the nave roof survives, and a few slabs in the wings.
	for i := 0; i+1 < len(naveCols); i++ {
		u0, u1 := naveCols[i]-0.8, naveCols[i+1]+0.8
		for j, strip := range [][2]float64{{-r0 - 0.6, -3.6}, {-3.6, 3.6}, {3.6, r0 + 0.6}} {
			keep := 0.35
			if j == 1 {
				keep = 0.22
			}
			if Hash01(i, j, 2700) < keep {
				g.Use(Sand).Fill(k.blk(u0, u1, strip[0], strip[1], G+nTop+0.8, G+nTop+1.5))
			}
		}
	}
	for s, side := range []float64{-1, 1} {
		for j := 0; j+1 < len(wingRows); j++ {
			for i := 0; i+1 < len(cols); i++ {
				if (j == 0 && Hash01(i, s, 2800) < 0.2) || Hash01(i, j, s, 2801) < 0.06 {
					a, b := side*wingRows[j], side*wingRows[j+1]
					g.Use(Sand).Fill(k.blk(cols[i]+0.05, cols[i+1]-0.05, math.Min(a, b), math.Max(a, b), G+wTop+0.8, G+wTop+1.5))
				}
			}
		}
	}
	// Outer walls with relief registers and breaches, and side doorways.
	const wallTop = 11.2
	for s, side := range []float64{-1, 1} {
		k.shade(g, -5, 23, side*23.1-0.6, side*23.1+0.6, G, G+wallTop, []Color{Stone, Sand, DarkStone}, func(u, v, y float64) (bool, Color) {
			yy := y - G
			j := karnakTempleFloor(u / 2.4)
			r := Hash01(j, s, 2900)
			top := wallTop
			switch {
			case side > 0 && r < 0.85:
				top = 1.6 + 2.8*Hash01(j, s, 2901)
			case side < 0 && r < 0.35:
				top = 4.5 + 5.7*Hash01(j, s, 2901)
			}
			if yy > top {
				return false, 0
			}
			if math.Abs(u-9.2) < 1.4 && yy < 5.2 {
				return false, 0
			}
			if yy > wallTop-0.8 {
				return true, Sand
			}
			outer := math.Abs(v) > 23.3
			if outer && (math.Abs(yy-3.2) < 0.25 || math.Abs(yy-6.4) < 0.25) {
				return true, DarkStone
			}
			course := karnakTempleFloor(yy / 0.8)
			if outer && yy > 1.2 && yy < 9.0 {
				if Hash01(karnakTempleFloor(u/0.8), course, s, 2902) < 0.3 {
					return true, Sand
				}
				return true, Stone
			}
			return true, k.masonry(2903+s, course, karnakTempleFloor(u/2.4), 0, yy < 1.2)
		})
	}
}

// akhmenu is Thutmose III's festival hall: tent-pole columns inside ruined walls.
func (k KarnakTemple) akhmenu(g *Grid) {
	const G = karnakTempleG
	stone := []Color{Stone, Stone, Sand, DarkStone}
	k.ruinWall(g, 53.6, 65.6, 9.2, 10.2, 1.2, 4.6, 3000, stone)
	k.ruinWall(g, 53.6, 65.6, -10.2, -9.2, 1.2, 4.6, 3001, stone)
	k.ruinWall(g, 64.6, 65.6, -9.2, 9.2, 1.4, 4.6, 3002, stone)
	k.ruinWall(g, 53.6, 54.6, 2.4, 9.2, 1.4, 4.6, 3003, stone)
	k.ruinWall(g, 53.6, 54.6, -9.2, -2.4, 1.4, 4.6, 3004, stone)
	for i := 0; i < 5; i++ {
		u := 56.2 + 2.0*float64(i)
		for _, v := range []float64{-2.4, 2.4} {
			c := k.at(u, v)
			g.Use(Stone).Fill(Cylinder{c, 0.9, 0, G, G + 0.4})
			g.Use(Sand).Fill(Frustum{c, 0.55, 0.85, G + 0.4, G + 4.4})
			g.Paint(Red, Frustum{c, 0.85, 0.95, G + 3.0, G + 3.6})
			g.Paint(Navy, Frustum{c, 0.85, 0.95, G + 3.6, G + 4.0})
			g.Use(Stone).Fill(Frustum{c, 0.85, 1.1, G + 4.4, G + 5.0}, BlockAt(c, 1.7, 1.7, G+5.0, G+5.4))
		}
		for _, v := range []float64{-6.4, 6.4} {
			if Hash01(i, int(v), 3010) < 0.75 {
				h := 2.5 + 2.4*Hash01(i, int(v), 3011)
				g.Use(Stone).Fill(BlockAt(k.at(u, v), 1.1, 1.1, G, G+h))
			}
		}
	}
	for _, v := range []float64{-2.4, 2.4} {
		g.Use(Stone).Fill(k.blk(55.4, 63, v-0.6, v+0.6, G+5.4, G+6.2))
	}
	g.Use(Sand).Fill(k.blk(58.2, 62.2, -2.4, 2.4, G+6.2, G+6.8))
}
