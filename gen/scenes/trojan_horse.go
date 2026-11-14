package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// TrojanHorse is the wooden horse before the gates of Troy. The horse stands in the
// middle of a sandy plain, close to profile for the hero camera and heading west (-X)
// for the Scaean Gate: a hollow barrel of staves and dark hoops on four straight legs,
// an arched neck with a cropped mane, a long head with ears and a bronze bridle, a
// hanging tail, and an open trapdoor in the belly with a rope ladder down to a heavy
// timber platform on four solid plank wheels. Tow ropes run from the platform and the
// collar to files of small figures hauling it, and its ruts curve back to the abandoned
// Greek camp. Troy's walls wrap the back (-Z) and the left (-X): a battered limestone
// base under a vertical mud-brick top with merlons, the gate with its porch, timber
// lintel and open bronze-banded leaves beside a flanking tower, a corner bastion and a
// second tower, watchers on the walk, and flat-roofed houses and a megaron palace on the
// citadel inside. Toward the hero camera (+X, +Z) lie the camp of striped tents, cold
// fires and a stake palisade, olive trees, and a diagonal strip of beach and sea with
// three black-hulled ships.
type TrojanHorse struct{}

func init() { Register(TrojanHorse{}) }

func (TrojanHorse) Name() string  { return "trojan_horse" }
func (TrojanHorse) Title() string { return "Trojan Horse" }

// Key dimensions. The plate is a rounded rectangle around (CX, CZ); the shoreline runs
// across the seaward diagonal (0.6, 0.8); wall positions are the outer faces of the
// vertical mud-brick tops. The walk, city and tower levels sit just above a cell layer
// so that their half-unit skins come out exactly two layers thick.
const (
	trojanHorseCX      = 0.0
	trojanHorseCZ      = -0.25
	trojanHorseA       = 58.0
	trojanHorseB       = 51.25
	trojanHorsePlain   = 1.1
	trojanHorseCoast   = 37.0
	trojanHorseWestX   = -42.0
	trojanHorseNorthZ  = -33.0
	trojanHorseThick   = 4.5
	trojanHorseBatter  = 9.5
	trojanHorseWalk    = 15.52
	trojanHorseCity    = 10.75
	trojanHorseGateZ   = 16.0
	trojanHorseGateW   = 4.5
	trojanHorseGateTop = 12.5
	trojanHorsePorchX  = -38.0
	trojanHorseDeck    = 6.6
)

var (
	// Center of the horse's platform and of the Greek camp.
	trojanHorseStand = P(2, 0)
	trojanHorseCamp  = P(36, 8)
)

// trojanHorseTower is a rectangular wall tower. Its listed sides carry the same batter
// as the walls below the mud brick.
type trojanHorseTower struct {
	Min, Max           Vec2
	Top                float64
	XLo, XHi, ZLo, ZHi bool
}

// The gate's flanking tower, the corner bastion and the north wall's tower.
var trojanHorseTowers = []trojanHorseTower{
	{Min: P(-46.5, -1.5), Max: P(-35.5, 9.3), Top: 22.25, XHi: true, ZLo: true},
	{Min: P(-49, -40), Max: P(-36, -27), Top: 23.85, XHi: true, ZHi: true},
	{Min: P(26, -37.5), Max: P(37, -26.5), Top: 21.45, XLo: true, XHi: true, ZHi: true},
}

// Houses inside the walls: plan corners, height and plaster color.
var trojanHorseHouses = []struct {
	Lo, Hi Vec2
	H      float64
	Wall   Color
}{
	{P(-8, -47), P(-2, -40.5), 5, White},
	{P(0, -49.5), P(7, -41.5), 8.5, Sand},
	{P(9, -46.5), P(15, -40.5), 5, White},
	{P(17, -49.5), P(24, -43), 8, White},
	{P(26, -47.5), P(32.5, -41), 5.5, Sand},
	{P(34.5, -47.5), P(41, -41), 8.5, White},
	{P(-53.5, -37), P(-49.5, -31.5), 6.5, White},
	{P(-55.8, -29), P(-49, -22.5), 5.5, Sand},
	{P(-56.5, -20.5), P(-48.5, -12.5), 8.5, White},
	{P(-56, -10.5), P(-49.5, -4), 5.5, Sand},
	{P(-56.5, -2), P(-49, 6.5), 8, White},
	{P(-55.5, 24), P(-49, 30.5), 6, White},
	{P(-52.5, 32.5), P(-47.8, 38), 8.5, Sand},
}

// trojanHorseTaper is a capsule whose radius varies linearly from A to B.
type trojanHorseTaper struct {
	A, B   Vec3
	R0, R1 float64
}

func (c trojanHorseTaper) Contains(p Vec3) bool {
	ab := c.B.Sub(c.A)
	t := math.Max(0, math.Min(1, p.Sub(c.A).Dot(ab)/math.Max(ab.Dot(ab), 1e-12)))
	return p.Sub(c.A.Add(ab.Scale(t))).Length() <= c.R0+(c.R1-c.R0)*t
}

func (c trojanHorseTaper) Bounds() Box { return BoxOf(c.A, c.B).Grow(math.Max(c.R0, c.R1)) }

// trojanHorseFrame is an oriented frame: local X runs along the plan angle, local Z to
// its left, and local Y up from Y0.
type trojanHorseFrame struct {
	C     Vec2
	Y0    float64
	Angle float64
}

func (f trojanHorseFrame) local(p Vec3) Vec3 {
	q := p.Plan().Sub(f.C).Rotate(-f.Angle)
	return V(q.X, p.Y-f.Y0, q.Y)
}

func (f trojanHorseFrame) world(l Vec3) Vec3 {
	return P(l.X, l.Z).Rotate(f.Angle).Add(f.C).At(l.Y + f.Y0)
}

// trojanHorsePlaced maps a solid authored in a frame's local coordinates into the world.
type trojanHorsePlaced struct {
	S Solid
	F trojanHorseFrame
}

func (s trojanHorsePlaced) Contains(p Vec3) bool { return s.S.Contains(s.F.local(p)) }

func (s trojanHorsePlaced) Bounds() Box {
	in := s.S.Bounds()
	b := EmptyBox()
	for _, x := range []float64{in.Min.X, in.Max.X} {
		for _, y := range []float64{in.Min.Y, in.Max.Y} {
			for _, z := range []float64{in.Min.Z, in.Max.Z} {
				w := s.F.world(V(x, y, z))
				b = b.Union(Box{w, w})
			}
		}
	}
	return b
}

func trojanHorseSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func trojanHorseLerp(a, b, t float64) float64 { return a + (b-a)*t }

// trojanHorseNoise is smooth value noise in [-1, 1].
func trojanHorseNoise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	ix, iz := int(x0), int(z0)
	sx, sz := trojanHorseSmooth(x-x0), trojanHorseSmooth(z-z0)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	return trojanHorseLerp(trojanHorseLerp(a, b, sx), trojanHorseLerp(c, d, sx), sz)*2 - 1
}

func trojanHorseFBM(x, z float64, seed int) float64 {
	return 0.65*trojanHorseNoise(x, z, seed) + 0.35*trojanHorseNoise(x*2.1+13, z*2.1+7, seed+1)
}

// trojanHorseCellHash is a deterministic value in [0, 1) for the cell under a world point.
func trojanHorseCellHash(g *Grid, p Vec3, seed int) float64 {
	c := Nearest(p.Scale(1 / g.Span()))
	return Hash01(int(c.X), int(c.Y), int(c.Z), seed)
}

// trojanHorsePolyline returns the plan distance to a polyline and the arc length at the
// nearest point.
func trojanHorsePolyline(p Vec2, pts []Vec2) (dist, along float64) {
	dist = math.Inf(1)
	run := 0.0
	for i := 0; i+1 < len(pts); i++ {
		a, b := pts[i], pts[i+1]
		ab := b.Sub(a)
		l := ab.Length()
		t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/(l*l)))
		if d := p.Sub(a.Add(ab.Scale(t))).Length(); d < dist {
			dist, along = d, run+t*l
		}
		run += l
	}
	return
}

// trojanHorseShell is the skin of a predicate: its points within ts (sideways) or tu
// (above, and below when under is set) of the outside, so large masses stay hollow.
func trojanHorseShell(box Box, ts, tu float64, under bool, in func(Vec3) bool) Func {
	return Func{box, func(p Vec3) bool {
		if !in(p) {
			return false
		}
		if !in(p.Add(V(ts, 0, 0))) || !in(p.Add(V(-ts, 0, 0))) || !in(p.Add(V(0, 0, ts))) ||
			!in(p.Add(V(0, 0, -ts))) || !in(p.Add(V(0, tu, 0))) {
			return true
		}
		return under && !in(p.Add(V(0, -tu, 0)))
	}}
}

// trojanHorseSpread is how far the battered limestone stands out from the face above it.
func trojanHorseSpread(y float64) float64 {
	if y >= trojanHorseBatter {
		return 0
	}
	return 0.5 + (trojanHorseBatter-y)*0.3
}

// contains reports whether the tower, grown by e, holds a point.
func (t trojanHorseTower) contains(p Vec3, e float64) bool {
	if p.Y > t.Top+e {
		return false
	}
	lo, hi := t.Min.Sub(P(e, e)), t.Max.Add(P(e, e))
	b := trojanHorseSpread(p.Y)
	if t.XLo {
		lo.X -= b
	}
	if t.XHi {
		hi.X += b
	}
	if t.ZLo {
		lo.Y -= b
	}
	if t.ZHi {
		hi.Y += b
	}
	return p.X >= lo.X && p.X <= hi.X && p.Z >= lo.Y && p.Z <= hi.Y
}

// recolor repaints the occupied cells inside box for which f returns a new color.
func (TrojanHorse) recolor(g *Grid, box Box, f func(p Vec3, cur Color) (Color, bool)) {
	prev := g.Brush()
	for _, cell := range g.Cells() {
		p := g.World(cell)
		if !box.Contains(p) {
			continue
		}
		cur, _ := g.ColorOf(cell)
		if next, ok := f(p, cur); ok && next != cur {
			g.Remove(cell)
			g.Use(next)
			g.Add(cell)
		}
	}
	g.Use(prev)
}

// inPlate is the rounded-rectangle site.
func (TrojanHorse) inPlate(x, z float64) bool {
	dx, dz := math.Abs(x-trojanHorseCX)/trojanHorseA, math.Abs(z-trojanHorseCZ)/trojanHorseB
	return dx*dx*dx*dx*math.Sqrt(dx)+dz*dz*dz*dz*math.Sqrt(dz) <= 1
}

// inland is the distance of a plan point from the shoreline (negative out at sea).
func (TrojanHorse) inland(x, z float64) float64 {
	return trojanHorseCoast + 1.6*trojanHorseNoise((0.8*x-0.6*z)*0.06, 0.5, 21) - (0.6*x + 0.8*z)
}

// ground is the terrain height: a gently rolling plain, level around the horse, a low
// dune line behind the beach and a strand sloping under the sea.
func (s TrojanHorse) ground(x, z float64) float64 {
	d := s.inland(x, z)
	h := trojanHorsePlain + 0.3*trojanHorseFBM(x*0.05, z*0.05, 23)
	h += 0.6 * math.Exp(-(d-5.5)*(d-5.5)/9) * (0.5 + 0.5*trojanHorseNoise((0.8*x-0.6*z)*0.1, 3.5, 25))
	k := P(x, z).Sub(trojanHorseStand).Length()
	h = trojanHorseLerp(h, trojanHorsePlain+0.02, trojanHorseSmooth((34-k)/10))
	return trojanHorseLerp(-0.8, h, trojanHorseSmooth((d+4)/9))
}

// frame is the horse's frame: it heads from its stand straight for the gate.
func (TrojanHorse) frame() trojanHorseFrame {
	gate := P(trojanHorseWestX+2, trojanHorseGateZ)
	return trojanHorseFrame{trojanHorseStand, trojanHorsePlain + 0.3, gate.Sub(trojanHorseStand).Angle()}
}

// track is the horse's path from the camp gateway, curving onto its axis and on to the
// gate, with the arc length up to the front axle (where the ruts end).
func (s TrojanHorse) track() (pts []Vec2, ruts float64) {
	f := s.frame()
	rear, knee := f.world(V(-14, 0, 0)).Plan(), f.world(V(-26, 0, 0)).Plan()
	out := knee.Sub(trojanHorseCamp)
	start := trojanHorseCamp.Add(out.Scale(11 / out.Length()))
	for i := 0; i <= 12; i++ {
		t := float64(i) / 12
		a, b := start.Add(knee.Sub(start).Scale(t)), knee.Add(rear.Sub(knee).Scale(t))
		pts = append(pts, a.Add(b.Sub(a).Scale(t)))
	}
	pts = append(pts, f.world(V(14, 0, 0)).Plan())
	for i := 0; i+1 < len(pts); i++ {
		ruts += pts[i+1].Sub(pts[i]).Length()
	}
	pts = append(pts, P(trojanHorsePorchX, trojanHorseGateZ), P(trojanHorseWestX-6, trojanHorseGateZ))
	return pts, ruts
}

// faceW is the X of the west wall's outer face at a height: alternate panels stand a
// little proud, and the limestone base slopes out below the mud brick.
func (TrojanHorse) faceW(y, z float64) float64 {
	x := trojanHorseWestX + trojanHorseSpread(y)
	if int(math.Floor((z+96)/14))%2 == 0 {
		x += 0.8
	}
	return x
}

// faceN is the Z of the north wall's outer face at a height.
func (TrojanHorse) faceN(y, x float64) float64 {
	z := trojanHorseNorthZ + trojanHorseSpread(y)
	if int(math.Floor((x+103)/14))%2 == 0 {
		z += 0.8
	}
	return z
}

// depth is how far a plan point lies behind the walls' outer faces (negative on the plain).
func (TrojanHorse) depth(x, z float64) float64 {
	return math.Max(trojanHorseWestX-x, trojanHorseNorthZ-z)
}

// ramp is the floor of the gate passage and of the street climbing from it into the city.
func (TrojanHorse) ramp(dW float64) float64 {
	return math.Min(trojanHorseCity, trojanHorsePlain+math.Max(0, dW-5))
}

// solid is the citadel as one mass: the battered walls with their walk, the towers, the
// gate porch and passage, and the raised city ground behind.
func (s TrojanHorse) solid(p Vec3) bool {
	if p.Y < 0 || !s.inPlate(p.X, p.Z) {
		return false
	}
	dW := trojanHorseWestX - p.X
	dz := math.Abs(p.Z - trojanHorseGateZ)
	if dz < trojanHorseGateW && dW > -6 && p.Y > s.ramp(dW) && (dW >= trojanHorseThick || p.Y < trojanHorseGateTop) {
		return false
	}
	for _, t := range trojanHorseTowers {
		if t.contains(p, 0) {
			return true
		}
	}
	if dz < trojanHorseGateW+2.2 && p.X <= trojanHorsePorchX && dW < trojanHorseThick && p.Y <= trojanHorseGateTop+1.8 {
		return true
	}
	if p.X > s.faceW(p.Y, p.Z) && p.Z > s.faceN(p.Y, p.X) {
		return false
	}
	if s.depth(p.X, p.Z) < trojanHorseThick {
		return p.Y <= trojanHorseWalk
	}
	return p.Y <= trojanHorseCity
}

func (s TrojanHorse) Build(g *Grid) Order {
	s.plain(g)
	s.citadel(g)
	s.watchers(g)
	s.grove(g)
	s.camp(g)
	s.ships(g)
	s.horse(g)
	s.haulers(g)
	return Along(P(-0.78, -0.63))
}

// plain lays the sea plate and a thin skin of sand over the terrain, then paints the
// surf, the scrub, the horse's ruts and the road to the gate.
func (s TrojanHorse) plain(g *Grid) {
	box := Box{V(-62, -1, -58), V(66, 3.2, 58)}
	open := func(x, z float64) bool { return s.inPlate(x, z) && !s.solid(V(x, 0.6, z)) }
	g.Use(Water).Fill(Func{Box{V(-62, -0.6, -58), V(66, 0.2, 58)}, func(p Vec3) bool {
		return open(p.X, p.Z) && s.ground(p.X, p.Z) < 0.8
	}})
	g.Use(Sand).Fill(Func{box, func(p Vec3) bool {
		if !open(p.X, p.Z) {
			return false
		}
		h := s.ground(p.X, p.Z)
		return h > 0.1 && p.Y <= h && p.Y > h-0.5
	}})
	track, ruts := s.track()
	s.recolor(g, box, func(p Vec3, cur Color) (Color, bool) {
		x, z := p.X, p.Z
		d := s.inland(x, z)
		if cur == Water {
			h := s.ground(x, z)
			switch {
			case h > -0.05 && trojanHorseCellHash(g, p, 3) < 0.6:
				return White, true
			case h > -0.5:
				return Glass, true
			case d < -15+4*trojanHorseFBM(x*0.07, z*0.07, 27):
				return Navy, true
			}
			return cur, false
		}
		dist, along := trojanHorsePolyline(p.Plan(), track)
		switch {
		case along > 1 && along < ruts && math.Abs(dist-10.2) < 0.8:
			return DarkStone, true
		case along >= ruts && dist < 4.2:
			return Stone, true
		case d < 5:
			return cur, false
		}
		worn := dist < 12.5 || p.Plan().Sub(trojanHorseCamp).Length() < 11.5
		if !worn && d > 9 && trojanHorseFBM(x*0.07, z*0.07, 33) > 0.25 {
			return Grass, true
		}
		if trojanHorseFBM(x*0.12+40, z*0.12, 35) > 0.32 {
			return Stone, true
		}
		return cur, false
	})
}

// citadel builds Troy in its own grid: the hollow skin of the walls, towers and city
// ground, its masonry colors, parapets, gate leaves, houses and the megaron.
func (s TrojanHorse) citadel(g *Grid) {
	c := NewGridWithSpan(g.Span())
	all := Box{V(-62, 0, -58), V(66, 28, 58)}
	c.Use(Stone)
	c.Fill(trojanHorseShell(Box{V(-62, 0, -58), V(-26, 28, 58)}, 0.8, 0.5, false, s.solid))
	c.Fill(trojanHorseShell(Box{V(-26, 0, -58), V(66, 28, -18)}, 0.8, 0.5, false, s.solid))
	s.recolor(c, all, func(p Vec3, _ Color) (Color, bool) { return s.stone(p), true })
	// The timber lintel closing the gate passage from below.
	c.Use(Wood).Fill(Block(V(trojanHorseWestX-trojanHorseThick, trojanHorseGateTop, trojanHorseGateZ-trojanHorseGateW),
		V(trojanHorsePorchX, trojanHorseGateTop+0.9, trojanHorseGateZ+trojanHorseGateW)))

	// Parapets with merlons along the walls and around the tower decks.
	c.Use(Sand)
	c.Fill(Func{Box{V(-62, trojanHorseWalk, -58), V(66, trojanHorseWalk+2.3, 58)}, func(p Vec3) bool {
		for _, t := range trojanHorseTowers {
			if t.contains(V(p.X, t.Top, p.Z), 0.2) {
				return false
			}
		}
		fw, fn := s.faceW(trojanHorseWalk, p.Z), s.faceN(trojanHorseWalk, p.X)
		along := p.Z
		switch {
		case p.Z > fn && p.X <= fw && p.X > fw-1:
		case p.X > fw && p.Z <= fn && p.Z > fn-1:
			along = p.X
		default:
			return false
		}
		return s.inPlate(p.X, p.Z) && (p.Y <= trojanHorseWalk+1.2 || math.Mod(along+200, 2.8) < 1.4)
	}})
	for _, t := range trojanHorseTowers {
		c.Fill(Func{Box{t.Min.At(t.Top), t.Max.At(t.Top + 2.3)}, func(p Vec3) bool {
			ex := math.Min(p.X-t.Min.X, t.Max.X-p.X)
			ez := math.Min(p.Z-t.Min.Y, t.Max.Y-p.Z)
			along := p.X
			if ex < ez {
				along = p.Z
			}
			return math.Min(ex, ez) <= 1 && (p.Y <= t.Top+1.2 || math.Mod(along+200, 2.8) < 1.4)
		}})
		// Dark slits high on every face.
		for _, k := range []float64{0.3, 0.7} {
			x, z := trojanHorseLerp(t.Min.X, t.Max.X, k), trojanHorseLerp(t.Min.Y, t.Max.Y, k)
			y := t.Top - 4.6
			c.Paint(Charcoal,
				Block(V(x-0.5, y, t.Max.Y-0.9), V(x+0.5, y+2, t.Max.Y+0.3)),
				Block(V(x-0.5, y, t.Min.Y-0.3), V(x+0.5, y+2, t.Min.Y+0.9)),
				Block(V(t.Max.X-0.9, y, z-0.5), V(t.Max.X+0.3, y+2, z+0.5)),
				Block(V(t.Min.X-0.3, y, z-0.5), V(t.Min.X+0.9, y+2, z+0.5)))
		}
	}

	// Open gate leaves, swung out against the plain, with bronze bands.
	for _, sd := range []float64{-1, 1} {
		z0 := trojanHorseGateZ + sd*(trojanHorseGateW+0.1)
		lo, hi := V(trojanHorsePorchX, trojanHorsePlain-0.3, z0), V(trojanHorsePorchX+4.8, trojanHorseGateTop-0.6, z0+sd)
		c.Use(Wood).Fill(Block(lo, hi))
		for _, y := range []float64{3, 7, 11} {
			c.Paint(Gold, Block(V(lo.X, y, lo.Z), V(hi.X, y+0.8, hi.Z)))
		}
	}
	s.city(c)
	g.Merge(c)
}

// stone colors one cell of the citadel's skin: coursed limestone below, a timber
// lacing course, banded mud brick above, plastered walks, paved streets, and earth
// strata where the plate cuts the mound.
func (s TrojanHorse) stone(p Vec3) Color {
	dep := s.depth(p.X, p.Z)
	dW := trojanHorseWestX - p.X
	dz := math.Abs(p.Z - trojanHorseGateZ)
	top, tower := trojanHorseWalk, false
	for _, t := range trojanHorseTowers {
		if t.contains(p, 0.3) {
			top, tower = t.Top, true
		}
	}
	lime := func() Color {
		course := int(math.Floor(p.Y / 1.2))
		h := Hash01(int(math.Floor((p.X+p.Z)/2.6+0.5*float64(course&1))), course, 41)
		switch {
		case h < 0.25:
			return White
		case h > 0.85:
			return Sand
		}
		return Stone
	}
	switch {
	case tower:
	case dz < trojanHorseGateW+2.4 && dW < trojanHorseThick+0.2 && p.Y < trojanHorseGateTop+2:
		// The porch: white piers, a timber lintel, a shadowed passage.
		switch {
		case p.Y > trojanHorseGateTop+1:
			return Stone
		case p.Y >= trojanHorseGateTop-0.2:
			return Wood
		case dz < trojanHorseGateW-0.2:
			return Stone
		case p.X > trojanHorsePorchX-2.2:
			return White
		}
		return DarkStone
	case dep >= trojanHorseThick+0.3 && dz < trojanHorseGateW+1.3:
		// The stepped street up from the gate between retaining walls.
		if dz < trojanHorseGateW && p.Y <= s.ramp(dW)+0.2 {
			if int(math.Floor(p.Y/0.8))%2 == 0 {
				return DarkStone
			}
			return Stone
		}
		return lime()
	case dep >= trojanHorseThick+0.3:
		if p.Y > trojanHorseCity-0.9 {
			if Hash01(int(math.Floor(p.X/2.2)), int(math.Floor(p.Z/2.2)), 43) < 0.3 {
				return Sand
			}
			return Stone
		}
		band := int(math.Floor((p.Y+0.6*trojanHorseNoise((p.X+p.Z)*0.08, 1.5, 47))/1.6)) + 8
		return []Color{DarkStone, Sand, Terracotta, Sand, Stone, Sand, DarkStone, Terracotta}[band%8]
	}
	switch {
	case p.Y > top-0.85:
		return Sand
	case p.Y >= trojanHorseBatter+0.8:
		course := int(math.Floor((p.Y - trojanHorseBatter) / 0.8))
		switch {
		case course%4 == 2:
			return Brick
		case Hash01(int(math.Floor((p.X+p.Z)/2.2+0.5*float64(course&1))), course, 45) < 0.22:
			return Sand
		}
		return Terracotta
	case p.Y >= trojanHorseBatter:
		return Wood
	}
	return lime()
}

// city places the flat-roofed houses and the megaron on the ground inside the walls.
func (s TrojanHorse) city(g *Grid) {
	for i, h := range trojanHorseHouses {
		ok := true
		for _, q := range []Vec2{h.Lo, h.Hi, P(h.Lo.X, h.Hi.Y), P(h.Hi.X, h.Lo.Y)} {
			e := q.Sub(h.Lo.Add(h.Hi).Scale(0.5))
			ok = ok && s.inPlate(q.X+0.8*e.X/math.Abs(e.X), q.Y+0.8*e.Y/math.Abs(e.Y))
		}
		if ok {
			s.house(g, h.Lo, h.Hi, h.H, h.Wall, i)
		}
	}
	s.megaron(g)
}

// house is a flat-roofed plastered house: hollow walls, a roof parapet over a beam
// course, dark windows, and sometimes a roof room or a cloth awning.
func (TrojanHorse) house(g *Grid, lo, hi Vec2, h float64, wall Color, seed int) {
	y0, top := trojanHorseCity-0.4, trojanHorseCity+h
	in := P(0.9, 0.9)
	roof := Sand
	if wall == Sand {
		roof = Stone
	}
	g.With(wall, func() {
		g.Fill(Subtract(Block(lo.At(y0), hi.At(top)), Block(lo.Add(in).At(y0-1), hi.Sub(in).At(top-0.9))))
		g.Fill(Subtract(Block(lo.At(top), hi.At(top+0.8)), Block(lo.Add(in).At(top-1), hi.Sub(in).At(top+2))))
	})
	g.Paint(roof, Block(lo.Add(in).At(top-0.9), hi.Sub(in).At(top+0.1)))
	g.Paint(Wood, Subtract(Block(lo.At(top-1.6), hi.At(top-0.9)), Block(lo.Add(in).At(top-3), hi.Sub(in).At(top))))
	for x := lo.X + 1.9; x < hi.X-1.4; x += 2.6 {
		g.Paint(Charcoal, Block(V(x-0.5, top-3.8, hi.Y-0.6), V(x+0.5, top-2.4, hi.Y+0.3)))
	}
	for z := lo.Y + 1.9; z < hi.Y-1.4; z += 2.6 {
		g.Paint(Charcoal, Block(V(hi.X-0.6, top-3.8, z-0.5), V(hi.X+0.3, top-2.4, z+0.5)))
	}
	c := lo.Add(hi).Scale(0.5)
	switch seed % 3 {
	case 0:
		r := c.Add(P(-0.7, -0.6))
		g.With(wall, func() {
			g.Fill(Subtract(BlockAt(r, 3.4, 3.2, top, top+2.6), BlockAt(r, 1.6, 1.4, top, top+1.8)))
		})
		g.Paint(Charcoal, BlockAt(r.Add(P(1.5, 0)), 0.8, 1.2, top, top+1.8))
	case 1:
		cloth := []Color{Red, Navy}[(seed/3)%2]
		g.With(Wood, func() {
			for _, d := range []Vec2{P(-1.3, -1.3), P(1.3, -1.3), P(-1.3, 1.3), P(1.3, 1.3)} {
				g.Fill(BlockAt(c.Add(d), 0.9, 0.9, top, top+2.3))
			}
		})
		g.With(cloth, func() { g.Fill(BlockAt(c, 4, 4, top+2.3, top+2.9)) })
	}
}

// megaron is the palace hall behind the north wall: a long white hall on a podium with
// a porch of two red columns facing east, a painted frieze and a roof lantern.
func (TrojanHorse) megaron(g *Grid) {
	const x0, x1, z0, z1 = -34.5, -14.5, trojanHorseNorthZ - 14.5, trojanHorseNorthZ - 6.5
	base, floor := trojanHorseCity-0.4, trojanHorseCity+3.5
	top, hall, mid := floor+8.5, x1-4.5, (z0+z1)/2
	g.With(Stone, func() {
		g.Fill(Subtract(Block(V(x0-1.2, base, z0-0.8), V(x1+2.5, floor, z1+1.2)),
			Block(V(x0-0.2, base-1, z0+0.2), V(x1+1.5, floor-0.9, z1+0.2))))
		g.Fill(Block(V(x1+2.5, base, z0+1.5), V(x1+3.7, floor-1.2, z1-1.5)),
			Block(V(x1+3.7, base, z0+1.5), V(x1+4.9, floor-2.4, z1-1.5)))
	})
	g.With(White, func() {
		g.Fill(Subtract(Block(V(x0, floor, z0), V(hall, top, z1)), Block(V(x0+1, floor-1, z0+1), V(hall-1, top-0.9, z1-1))))
		g.Fill(Block(V(hall, floor, z0), V(x1, top, z0+1)), Block(V(hall, floor, z1-1), V(x1, top, z1)),
			Block(V(hall, top-0.9, z0), V(x1+0.5, top, z1)))
	})
	g.Paint(Sand, Block(V(x0+1, top-0.9, z0+1), V(x1-0.5, top+0.1, z1-1)))
	ring := func(y0, y1 float64) Solid {
		return Subtract(Block(V(x0-0.1, y0, z0-0.1), V(x1+0.6, y1, z1+0.1)), Block(V(x0+1, y0-1, z0+1), V(hall-1, y1+1, z1-1)))
	}
	g.Paint(Stone, ring(floor, floor+1.2))
	g.Paint(Red, ring(top-3.4, top-2.7))
	g.Paint(Navy, ring(top-2.7, top-1.6))
	g.Paint(Charcoal, Block(V(hall-1.1, floor, mid-1.3), V(hall+0.1, floor+5, mid+1.3)))
	for _, z := range []float64{z0 + 2.8, z1 - 2.8} {
		g.With(Red, func() { g.Fill(Cylinder{Center: P(x1-0.6, z), Radius: 0.7, Y0: floor, Y1: top - 1.7}) })
		g.With(Charcoal, func() { g.Fill(Cylinder{Center: P(x1-0.6, z), Radius: 0.9, Y0: top - 1.7, Y1: top - 0.9}) })
	}
	lc := P((x0+hall)/2, mid)
	g.With(White, func() { g.Fill(Subtract(BlockAt(lc, 5, 4, top, top+2.2), BlockAt(lc, 3.2, 2.2, top, top+1.4))) })
	g.Paint(Charcoal, BlockAt(lc, 5.2, 4.2, top+0.6, top+1.4))
	g.With(Terracotta, func() { g.Fill(BlockAt(lc, 6, 5, top+2.2, top+2.8)) })
}

// figure is a small standing person: a tunic and a head, leaning by lean in plan.
func (TrojanHorse) figure(g *Grid, at Vec3, lean Vec2, tunic, head Color) {
	sh := at.Add(V(lean.X, 2.1, lean.Y))
	g.With(tunic, func() { g.Fill(Capsule{at.Add(V(0, 0.5, 0)), sh, 0.62}) })
	g.With(head, func() { g.Fill(Sphere(sh.Add(V(lean.X*0.4, 0.95, lean.Y*0.4)), 0.6)) })
}

// watchers line the wall walks, crowd over the gate and the tower decks, and a few
// Trojans wait outside the gate.
func (s TrojanHorse) watchers(g *Grid) {
	tunics := []Color{White, Red, Navy, White, Glass, Red}
	n := 0
	stand := func(x, y, z float64) {
		head := Terracotta
		if n%3 == 0 {
			head = Gold
		}
		s.figure(g, V(x, y, z), P(0, 0), tunics[n%len(tunics)], head)
		n++
	}
	for _, x := range []float64{-31, -27, -14, -10, -1, 3.5, 8, 14, 19, 41, 46} {
		stand(x, trojanHorseWalk, trojanHorseNorthZ-1.9)
	}
	for _, z := range []float64{-22, -17, -9, 12.5, 15.5, 18.5, 21.5, 28, 33, 40} {
		stand(trojanHorseWestX-1.9, trojanHorseWalk, z)
	}
	for i, spots := range [][]Vec2{{P(-37.5, 1.5), P(-37.5, 6), P(-40.5, 7.4)}, {P(-37.8, -29.5), P(-41.5, -28.9)},
		{P(31.5, -28.5), P(28, -28.5)}} {
		for _, q := range spots {
			stand(q.X, trojanHorseTowers[i].Top, q.Y)
		}
	}
	for _, q := range []Vec2{P(-30, 25.5), P(-28.3, 27.4), P(-31.6, 27.8)} {
		stand(q.X, s.ground(q.X, q.Y)+0.3, q.Y)
	}
}

// grove scatters olive trees: a grove toward the front left, a few by the walls and
// more inland of the camp.
func (s TrojanHorse) grove(g *Grid) {
	for i, c := range []Vec2{P(-27, 32), P(-20, 38), P(-32, 41), P(-13, 31), P(-9, 43), P(-22, 46.5), P(-2, 37),
		P(-33.5, 46), P(5, 31), P(-29, -13), P(-19, -20), P(40, -21), P(46, -14), P(51, -20), P(34, -17), P(54, -5)} {
		s.olive(g, c, i)
	}
}

// olive is a gnarled leaning trunk under a low, lumpy two-tone crown.
func (s TrojanHorse) olive(g *Grid, c Vec2, seed int) {
	y := s.ground(c.X, c.Y) - 0.3
	fork := c.Add(Polar(P(0, 0), 0.9, Hash01(seed, 51)*2*math.Pi)).At(y + 2.6 + Hash01(seed, 52))
	r := 2.6 + 0.8*Hash01(seed, 53)
	g.With(Wood, func() {
		g.Fill(trojanHorseTaper{c.At(y), fork, 0.9, 0.65})
		for k := 0; k < 2; k++ {
			a := Hash01(seed, 54)*math.Pi + float64(k)*math.Pi
			g.Fill(Capsule{fork, Polar(fork.Plan(), 1.5, a).At(fork.Y + 1.3), 0.5})
		}
	})
	off := Polar(P(0, 0), 0.5*r, Hash01(seed, 55)*2*math.Pi)
	g.With(Foliage, func() { g.FillVolume(Ellipsoid{fork.Add(V(0, 1.9, 0)), V(r, 0.62*r, r)}) })
	g.With(Grass, func() { g.FillVolume(Ellipsoid{fork.Add(V(off.X, 2.5, off.Y)), V(0.6*r, 0.45*r, 0.6*r)}) })
}

// camp is the abandoned Greek camp: a stake palisade open to the beach with a gateway
// toward the horse, striped tents, one of them fallen, cold fire rings, jars and a
// pile of leftover timber.
func (s TrojanHorse) camp(g *Grid) {
	c := trojanHorseCamp
	track, _ := s.track()
	gate := track[0].Sub(c).Angle()
	const r = 11.0
	for i, a := 0, gate-1.75; a <= gate+1.75; i, a = i+1, a+0.95/r {
		off := math.Abs(a - gate)
		if off < 0.36 {
			continue
		}
		q := Polar(c, r, a)
		y := s.ground(q.X, q.Y)
		top := y + 3.1 + 0.8*float64(i%2)
		if off < 0.45 {
			top = y + 5.4
		}
		g.With(Wood, func() { g.Fill(Cylinder{Center: q, Radius: 0.55, Y0: y - 0.5, Y1: top}) })
		g.With(Sand, func() { g.Fill(Cone(q, 0.55, top, 0.8)) })
	}
	s.coneTent(g, c.Add(P(0, 1.5)), 3.7, 6.2, Red, gate)
	s.coneTent(g, c.Add(P(-7.2, -0.5)), 2.8, 4.6, Navy, gate+0.6)
	s.ridgeTent(g, c.Add(P(-6, 5.5)), 0.9, 6.5, 4.6, Sand, Red)
	s.ridgeTent(g, c.Add(P(6, -3.6)), 2.3, 6, 4.4, White, Navy)
	// A fallen tent with its pole.
	fallen := c.Add(P(1.8, -7.2))
	fy := s.ground(fallen.X, fallen.Y)
	g.With(Sand, func() { g.FillVolume(Ellipsoid{fallen.At(fy), V(2.8, 1.1, 2)}) })
	g.With(Wood, func() {
		g.Fill(Capsule{fallen.Add(P(-2.6, 0.6)).At(fy + 0.9), fallen.Add(P(2.9, -0.4)).At(fy + 1.2), 0.45})
	})
	for _, q := range []Vec2{c.Add(P(-2.5, -4.5)), c.Add(P(5.5, 3)), c.Add(P(-9, 9))} {
		y := s.ground(q.X, q.Y)
		g.With(DarkStone, func() { g.Fill(Cylinder{Center: q, Radius: 1.7, Inner: 0.8, Y0: y - 0.3, Y1: y + 0.6}) })
		g.With(Charcoal, func() {
			g.Fill(Cylinder{Center: q, Radius: 0.9, Y0: y - 0.3, Y1: y + 0.5})
			g.Fill(Capsule{q.Add(P(-1, 0.3)).At(y + 0.7), q.Add(P(1, -0.3)).At(y + 0.8), 0.4})
		})
	}
	for i, d := range []Vec2{P(1.5, -3.2), P(2.5, -2.6), P(1.6, -2.1)} {
		q := c.Add(d)
		y := s.ground(q.X, q.Y)
		g.With(Terracotta, func() { g.Fill(Capsule{q.At(y + 0.5), q.At(y + 1.5 + 0.3*float64(i%2)), 0.6}) })
	}
	// Timber left over from building the horse.
	f := trojanHorseFrame{P(44, -8), s.ground(44, -8) + 0.3, 0.4}
	for i, col := range []Color{Wood, Terracotta, Sand, Wood, Terracotta} {
		v, y := -1.4+1.4*float64(i), 0.5
		if i >= 3 {
			v, y = -0.7+1.4*float64(i-3), 1.7
		}
		g.With(col, func() { g.Fill(Capsule{f.world(V(-3.2, y, v)), f.world(V(3.2, y, v)), 0.62}) })
	}
}

// coneTent is a round striped tent with a dark doorway and a gold finial.
func (s TrojanHorse) coneTent(g *Grid, c Vec2, r, h float64, stripe Color, door float64) {
	y := s.ground(c.X, c.Y) - 0.2
	g.With(White, func() { g.FillVolume(Frustum{c, r, 0.3, y, y + h}) })
	g.Carve(Frustum{c, r - 1, 0, y - 0.3, y + h - 1.5})
	box := Box{V(c.X-r, y+0.4, c.Y-r), V(c.X+r, y+h, c.Y+r)}
	g.Paint(stripe, Func{box, func(p Vec3) bool {
		d := p.Plan().Sub(c)
		return d.Length() <= r*(1-(p.Y-y)/h)+0.5 && int(math.Floor((d.Angle()-door)/(math.Pi/5)+100))%2 == 0
	}})
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		return p.Y < y+2.6 && p.Plan().Sub(Polar(c, r-0.4, door)).Length() < 1.1
	}})
	g.With(Wood, func() { g.Fill(Cylinder{Center: c, Radius: 0.45, Y0: y + h - 0.6, Y1: y + h + 1}) })
	g.With(Gold, func() { g.Fill(Sphere(c.At(y+h+1.3), 0.6)) })
}

// ridgeTent is a striped cloth ridge tent, open at one end, with a pole at each gable.
func (s TrojanHorse) ridgeTent(g *Grid, c Vec2, angle, l, w float64, cloth, stripe Color) {
	f := trojanHorseFrame{c, s.ground(c.X, c.Y) - 0.2, angle}
	hgt := 0.62 * w
	skin := Func{Box{V(-l/2, 0, -w/2), V(l/2, hgt, w/2)}, func(q Vec3) bool {
		e := w / 2 * (1 - q.Y/hgt)
		if math.Abs(q.Z) > e {
			return false
		}
		return math.Abs(q.Z) > e-1 || q.Y > hgt-1.1 || q.X < -l/2+0.9
	}}
	g.With(cloth, func() { g.Fill(trojanHorsePlaced{skin, f}) })
	g.Paint(stripe, trojanHorsePlaced{Func{skin.Box, func(q Vec3) bool {
		return skin.F(q) && int(math.Floor((q.X+l/2)/1.5))%2 == 1
	}}, f})
	g.With(Wood, func() {
		for _, u := range []float64{-l / 2, l / 2} {
			g.Fill(Capsule{f.world(V(u, 0, 0)), f.world(V(u, hgt+0.7, 0)), 0.45})
		}
	})
}

// ships beaches two black ships stern-first on the strand and floats a third offshore.
func (s TrojanHorse) ships(g *Grid) {
	sea, shore := P(0.6, 0.8), P(0.8, -0.6)
	at := func(t, a float64) Vec2 { return sea.Scale(t).Add(shore.Scale(a)) }
	s.ship(g, at(trojanHorseCoast+5.5, -8), sea.Angle(), 0.1, true)
	s.ship(g, at(trojanHorseCoast+5.5, 6), sea.Angle(), 0.1, false)
	s.ship(g, at(trojanHorseCoast+11, 20.5), shore.Angle(), -0.4, true)
}

// ship is a black-hulled galley with a red gunwale, a curled stern post, a stem post
// over a bronze ram, bow eyes, thwarts and, optionally, a mast with its sail furled.
func (TrojanHorse) ship(g *Grid, c Vec2, angle, keel float64, mast bool) {
	const L, B, D = 19.0, 4.8, 3.2
	f := trojanHorseFrame{c, keel, angle}
	t := NewGridWithSpan(g.Span())
	put := func(sol Solid) Solid { return trojanHorsePlaced{sol, f} }
	gun := func(u float64) float64 {
		k := u / (L / 2)
		return D + 1.2*k*k*k*k
	}
	halfw := func(u, y float64) float64 {
		k := math.Abs(u) / (L / 2)
		if k >= 1 || y < 0 {
			return -1
		}
		w := B / 2 * math.Pow(1-math.Pow(k, 2.4), 0.6)
		return w * (0.4 + 0.6*math.Sqrt(math.Min(y/(0.7*D), 1)))
	}
	box := Box{V(-L/2-1, 0, -B/2-0.5), V(L/2+1, D+1.6, B/2+0.5)}
	t.Use(Charcoal).FillVolume(put(Func{box, func(q Vec3) bool {
		w := halfw(q.X, q.Y)
		return q.Y <= gun(q.X) && w >= 0 && math.Abs(q.Z) <= w
	}}))
	t.Carve(put(Func{box, func(q Vec3) bool {
		a, b := halfw(q.X-0.9, q.Y), halfw(q.X+0.9, q.Y)
		return q.Y >= 0.9 && a >= 0 && b >= 0 && math.Abs(q.Z) <= math.Min(a, b)-0.9
	}}))
	// A red wale along the outside of the black hull, below the gunwale.
	t.Paint(Red, put(Func{box, func(q Vec3) bool {
		d := gun(q.X) - q.Y
		return d > 0.6 && d < 1.45 && math.Abs(q.Z) > halfw(q.X, q.Y)-0.6
	}}))
	t.Paint(White, put(Block(V(L/2-3.6, D-1.6, -B), V(L/2-2.3, D-0.4, B))))
	t.Use(Wood)
	for u := -L/2 + 3.4; u <= L/2-3.4; u += 2.5 {
		w := halfw(u, D-0.9)
		t.Fill(put(Block(V(u-0.4, D-1.3, -w), V(u+0.4, D-0.6, w))))
	}
	chain := func(col Color, r float64, pts ...Vec3) {
		t.Use(col)
		for i := 0; i+1 < len(pts); i++ {
			t.Fill(put(Capsule{pts[i], pts[i+1], r}))
		}
	}
	chain(Charcoal, 0.62, V(-L/2+1.2, D-0.3, 0), V(-L/2-0.5, D+2.6, 0), V(-L/2+0.4, D+4.8, 0), V(-L/2+2.4, D+5.4, 0))
	chain(Red, 0.62, V(-L/2+2.4, D+5.4, 0), V(-L/2+3.2, D+5.1, 0))
	chain(Charcoal, 0.62, V(L/2-1, D-0.3, 0), V(L/2+0.1, D+2.9, 0))
	chain(Gold, 0.6, V(L/2-1.5, 0.6, 0), V(L/2+1.6, 0.5, 0))
	if mast {
		chain(Wood, 0.5, V(0.5, 0.6, 0), V(0.5, D+8.5, 0))
		chain(Wood, 0.45, V(0.5, D+7.4, -3.6), V(0.5, D+7.4, 3.6))
		chain(Sand, 0.7, V(0.5, D+6.6, -3), V(0.5, D+6.6, 3))
	}
	g.Merge(t)
}

// horse builds the wooden horse on its wheeled platform, in its own grid so that its
// plank painting never touches the ground. Local X runs from tail to head.
func (s TrojanHorse) horse(g *Grid) {
	f := s.frame()
	t := NewGridWithSpan(g.Span())
	put := func(sol Solid) Solid { return trojanHorsePlaced{sol, f} }
	fill := func(c Color, parts ...Solid) {
		t.Use(c)
		for _, part := range parts {
			t.Fill(put(part))
		}
	}
	// tint repaints the cells holding a marker color, by local position.
	everywhere := Box{V(-200, -50, -200), V(200, 100, 200)}
	tint := func(marker Color, color func(q Vec3) Color) {
		s.recolor(t, everywhere, func(p Vec3, cur Color) (Color, bool) {
			if cur != marker {
				return cur, false
			}
			return color(f.local(p)), true
		})
	}
	// plank picks a board's color: dark and warm boards alternate, a few are pale.
	plank := func(i, j, seed int) Color {
		switch {
		case Hash01(i, j, seed) < 0.07:
			return Sand
		case i%2 == 0:
			return Wood
		case Hash01(i, seed+1) < 0.25:
			return Sand
		}
		return Terracotta
	}
	const deck = trojanHorseDeck

	// Platform: a planked deck framed in dark beams with iron straps.
	fill(Slate, Block(V(-22, deck-0.9, -8.5), V(22, deck, 8.5)))
	tint(Slate, func(q Vec3) Color {
		i := int(math.Floor((q.X + 22) / 2.2))
		switch {
		case math.Abs(q.X) > 20.8 || math.Abs(q.Z) > 7.3 || Hash01(i, 71) < 0.2:
			return Wood
		case i%2 == 0:
			return Terracotta
		}
		return Sand
	})
	fill(Wood,
		Block(V(-22, 3.8, -8.5), V(22, deck-0.9, -7.3)), Block(V(-22, 3.8, 7.3), V(22, deck-0.9, 8.5)),
		Block(V(-22, 3.8, -8.5), V(-20.8, deck-0.9, 8.5)), Block(V(20.8, 3.8, -8.5), V(22, deck-0.9, 8.5)),
		Block(V(-14.8, 3.8, -8.5), V(-13.2, deck-0.9, 8.5)), Block(V(13.2, 3.8, -8.5), V(14.8, deck-0.9, 8.5)))
	t.Paint(Charcoal, put(Func{Box{V(-22, 3.8, -8.7), V(22, deck-0.9, 8.7)}, func(q Vec3) bool {
		return math.Abs(q.Z) > 7.2 && math.Mod(q.X+22.6, 5.5) < 1
	}}))
	t.Paint(Gold, put(Func{Box{V(-22.2, 3.8, -8.7), V(22.2, deck-0.9, 8.7)}, func(q Vec3) bool {
		return math.Abs(q.Z) > 7.2 && math.Abs(q.X) > 20.7
	}}))

	// Axles and four solid plank wheels with iron tyres and bronze hubs.
	for _, u := range []float64{-14, 14} {
		fill(Charcoal, Capsule{V(u, 4.6, -11.9), V(u, 4.6, 11.9), 0.75})
		for _, v := range []float64{-10.2, 10.2} {
			t.Use(Navy).FillVolume(put(Func{Box{V(u-4.7, 0, v-1), V(u+4.7, 9.3, v+1)}, func(q Vec3) bool {
				return math.Hypot(q.X-u, q.Y-4.6) <= 4.6
			}}))
			fill(Gold, Capsule{V(u, 4.6, v), V(u, 4.6, v*1.16), 1.05})
		}
	}
	tint(Navy, func(q Vec3) Color {
		du := q.X - 14
		if q.X < 0 {
			du = q.X + 14
		}
		r := math.Hypot(du, q.Y-4.6)
		switch {
		case r > 3.7:
			return Charcoal
		case r < 1.3:
			return Gold
		case int(math.Floor((du+20)/1.7))%2 == 0:
			return Wood
		}
		return Terracotta
	})

	// The hollow hull: barrel body, arched neck and long head as one skin.
	const bodyY = 26.2
	neckPts := []Vec3{V(9.5, 28.5, 0), V(12.5, 35, 0), V(16, 40.5, 0), V(20, 43.5, 0)}
	neckR := []float64{5.6, 4.5, 3.6, 3.1}
	body := func(q Vec3, e float64) bool {
		k := (q.X + 0.5) / (14.8 + e)
		if math.Abs(k) >= 1 {
			return false
		}
		sc := math.Pow(1-k*k*k*k, 0.45)
		dy, dv := (q.Y-bodyY)/(6.6*sc+e), q.Z/(6.3*sc+e)
		return dy*dy+dv*dv <= 1
	}
	neck := func(q Vec3, e float64) bool {
		w := V(q.X, q.Y, q.Z*1.5)
		for i := 0; i+1 < len(neckPts); i++ {
			if (trojanHorseTaper{neckPts[i], neckPts[i+1], neckR[i] + e, neckR[i+1] + e}).Contains(w) {
				return true
			}
		}
		return false
	}
	// neckAt is the arc length along the neck and the offset toward its crest.
	neckAt := func(q Vec3) (arc, off float64) {
		best, run := math.Inf(1), 0.0
		for i := 0; i+1 < len(neckPts); i++ {
			a, b := neckPts[i], neckPts[i+1]
			du, dy := b.X-a.X, b.Y-a.Y
			l := math.Hypot(du, dy)
			k := math.Max(0, math.Min(1, ((q.X-a.X)*du+(q.Y-a.Y)*dy)/(l*l)))
			if d := math.Hypot(q.X-a.X-du*k, q.Y-a.Y-dy*k); d < best {
				best, arc, off = d, run+k*l, ((q.Y-a.Y)*du-(q.X-a.X)*dy)/l
			}
			run += l
		}
		return
	}
	// headAt is the distance along the head from the poll and the offset toward the brow.
	headAt := func(q Vec3) (w, h float64) {
		du, dy := q.X-19.3, q.Y-44.2
		return du*0.684 - dy*0.729, du*0.729 + dy*0.684
	}
	head := func(q Vec3, e float64) bool {
		w, h := headAt(q)
		if w < -1.5-e || w > 11.3+e {
			return false
		}
		k := 1.0
		if w > 10 {
			m := (w - 10) / (1.3 + e)
			k = math.Sqrt(math.Max(0, 1-m*m))
		}
		return h <= (2.5-0.06*w)*k+e && h >= -(3.3-0.17*w)*k-e && math.Abs(q.Z) <= (2.25-0.075*w)*k+e
	}
	fill(Glass, trojanHorseShell(Box{V(-16.5, 18.5, -7.5), V(29.5, 49, 7.5)}, 1.3, 1.3, true, func(q Vec3) bool {
		return body(q, 0) || neck(q, 0) || head(q, 0)
	}))
	// The open trapdoor in the belly.
	t.Carve(put(Block(V(-3.2, 18.5, -2.6), V(3.2, 21.6, 2.6))))
	hoops := []float64{-12.2, -6.6, -1, 4.6, 10.2}
	tint(Glass, func(q Vec3) Color {
		if head(q, 0.35) && !neck(q, -0.3) {
			w, h := headAt(q)
			side := math.Abs(q.Z) > 1.5-0.075*w
			switch {
			case w > 7.3 && w < 8.5:
				return Gold
			case side && w > 1.9 && w < 3.3 && h > 0.5 && h < 1.9:
				return White
			case side && w > 9.4 && w < 10.6 && h > -0.2 && h < 1:
				return Charcoal
			}
			return plank(int(math.Floor((h+6)/1.5)), 0, 85)
		}
		if neck(q, 0.35) && !body(q, 0.1) {
			arc, off := neckAt(q)
			switch {
			case arc > 3.3 && arc < 5:
				return Gold
			case math.Abs(arc-10.5) < 0.55 || math.Abs(arc-16.5) < 0.55:
				return Charcoal
			}
			return plank(int(math.Floor((off+9)/1.7)), int(arc/6), 81)
		}
		j := 0
		for _, hu := range hoops {
			if math.Abs(q.X-hu) < 0.62 {
				return Charcoal
			}
			if q.X > hu {
				j++
			}
		}
		phi := math.Atan2(q.Y-bodyY, q.Z)
		return plank(int(math.Floor((phi+math.Pi)/(2*math.Pi)*20))%20, j, 61)
	})
	t.Paint(Charcoal, put(Block(V(-4.3, 18, -3.7), V(4.3, 21.8, 3.7))))
	fill(Terracotta, Block(V(-4.1, 13.6, -2.6), V(-3.2, 20.2, 2.6)))
	fill(Gold, Block(V(-4.1, 19.3, -2.6), V(-3.2, 20.2, 2.6)))
	// The rope ladder from the trapdoor to the deck.
	fill(Sand, Capsule{V(2.4, 20.8, -1.6), V(2.4, deck, -1.6), 0.5}, Capsule{V(2.4, 20.8, 1.6), V(2.4, deck, 1.6), 0.5})
	for y := 8.6; y < 19.5; y += 2.2 {
		fill(Wood, Block(V(2, y-0.35, -1.6), V(2.8, y+0.35, 1.6)))
	}

	// Four straight legs: three-plank posts with a knee hoop, a bronze band and a hoof.
	legs := []Vec2{P(10.5, -3.9), P(10.5, 3.9), P(-11, -3.9), P(-11, 3.9)}
	for _, c := range legs {
		fill(Brick, Func{Box{V(c.X-2.6, deck, c.Y-2.4), V(c.X+2.6, 24, c.Y+2.4)}, func(q Vec3) bool {
			k := (q.Y - deck) / (24 - deck)
			hu, hv := 1.75+0.75*k, 1.6+0.55*k
			switch {
			case q.Y < deck+1.7:
				hu, hv = 2.2, 2.05
			case q.Y >= 13.6 && q.Y < 15:
				hu, hv = hu+0.35, hv+0.35
			}
			return math.Abs(q.X-c.X) <= hu && math.Abs(q.Z-c.Y) <= hv
		}})
	}
	tint(Brick, func(q Vec3) Color {
		c := legs[0]
		for _, l := range legs {
			if math.Abs(q.X-l.X)+math.Abs(q.Z-l.Y) < math.Abs(q.X-c.X)+math.Abs(q.Z-c.Y) {
				c = l
			}
		}
		switch {
		case q.Y < deck+1.7 || q.Y >= 13.6 && q.Y < 15:
			return Charcoal
		case q.Y < deck+2.6:
			return Gold
		case math.Abs(q.X-c.X) < 0.7 || math.Abs(q.Z-c.Y) < 0.6:
			return Terracotta
		}
		return Wood
	})

	// Tail, cropped mane, forelock and ears.
	tail := []Vec3{V(-12.5, 31, 0), V(-17, 32, 0), V(-20.5, 29.5, 0), V(-21.8, 23, 0), V(-21.3, 15.5, 0)}
	tailR := []float64{1.3, 1.35, 1.3, 1.15, 0.8}
	for i := 0; i+1 < len(tail); i++ {
		fill(Charcoal, trojanHorseTaper{tail[i], tail[i+1], tailR[i], tailR[i+1]})
	}
	fill(Gold, Capsule{V(-16, 31.8, 0), V(-17.2, 32, 0), 1.5})
	run := 0.0
	for i := 0; i+1 < len(neckPts); i++ {
		a, b := neckPts[i], neckPts[i+1]
		du, dy := b.X-a.X, b.Y-a.Y
		l := math.Hypot(du, dy)
		for d := 0.0; d < l; d += 0.45 {
			arc, k := run+d, d/l
			if arc < 4.6 {
				continue
			}
			r := neckR[i] + (neckR[i+1]-neckR[i])*k - 0.25
			if int(arc/1.8)%2 == 0 {
				r += 0.8
			}
			fill(Charcoal, Sphere(V(a.X+du*k-dy/l*r, a.Y+dy*k+du/l*r, 0), 1))
		}
		run += l
	}
	fill(Charcoal, Sphere(V(20.8, 46.6, 0), 1), Sphere(V(21.8, 45.6, 0), 0.9))
	fill(Terracotta, trojanHorseTaper{V(18.8, 46, -1.5), V(17.9, 49.6, -2), 1, 0.5},
		trojanHorseTaper{V(18.8, 46, 1.5), V(17.9, 49.6, 2), 1, 0.5})
	g.Merge(t)
}

// haulers strings tow ropes from the platform and the collar to files of small figures
// leaning into them on the road to the gate.
func (s TrojanHorse) haulers(g *Grid) {
	f := s.frame()
	// A rope is laid in short alternating strands so that it reads as twisted hemp.
	rope := func(pts ...Vec3) {
		n := 0
		for i := 0; i+1 < len(pts); i++ {
			a, b := f.world(pts[i]), f.world(pts[i+1])
			steps := int(math.Ceil(b.Sub(a).Length() / 1.3))
			for k := 0; k < steps; k++ {
				p0 := a.Add(b.Sub(a).Scale(float64(k) / float64(steps)))
				p1 := a.Add(b.Sub(a).Scale(float64(k+1) / float64(steps)))
				g.With([]Color{Sand, Wood}[n%2], func() { g.Fill(Capsule{p0, p1, 0.5}) })
				n++
			}
		}
	}
	tunics := []Color{White, Red, Navy, Glass, White, Brick}
	lean := Polar(P(0, 0), 0.8, f.Angle)
	n := 0
	file := func(v, u0, u1 float64) {
		for u := u0; u <= u1; u += 3.6 {
			q := f.world(V(u, 0, v))
			s.figure(g, V(q.X, s.ground(q.X, q.Z)+0.3, q.Z), lean, tunics[n%len(tunics)], Terracotta)
			n++
		}
	}
	for _, v := range []float64{-4.8, 4.8} {
		rope(V(21.5, 4.6, v), V(26.5, 2.2, v), V(38.4, 2.2, v))
		file(v, 27.4, 38.5)
	}
	for _, v := range []float64{-1.7, 1.7} {
		rope(V(13.4, 30, v*1.5), V(22.3, 15, v*1.25), V(31.2, 2.4, v), V(40, 2.2, v))
		file(v, 32.2, 40)
	}
}
