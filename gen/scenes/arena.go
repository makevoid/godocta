package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// Arena is an amphitheatre inspired by the Colosseum: an elliptical shell of three
// arcaded storeys and an attic, a raked cavea carried on radial walls, an inner ring
// gallery, a partly exposed hypogeum, and the ruined southern stretch of outer wall.
type Arena struct {
	A, B  float64 // Arena floor semi-axes. Defaults 21 × 13.
	Depth float64 // From the arena edge to the outer face. Default 26.
	Bays  int     // Arches per ring. Default 80.
}

func init() { Register(Arena{}) }

func (Arena) Name() string  { return "arena" }
func (Arena) Title() string { return "Arena" }

// ring maps a plan point to (d, φ): d is the offset from the arena ellipse within a
// family of ellipses with semi-axes (A + d, B + d); φ is the elliptic angle.
func (a Arena) ring(p Vec2) (float64, float64) {
	lo, hi := -math.Min(a.A, a.B)+1e-6, a.Depth+40
	for i := 0; i < 40; i++ {
		mid := (lo + hi) / 2
		x, z := p.X/(a.A+mid), p.Y/(a.B+mid)
		if x*x+z*z > 1 {
			lo = mid
		} else {
			hi = mid
		}
	}
	d := (lo + hi) / 2
	return d, math.Atan2(p.Y/(a.B+d), p.X/(a.A+d))
}

// bay returns the signed arc distance from the nearest bay center (arch axis) and
// from the nearest pier center, measured along the ellipse at offset d.
func (a Arena) bay(d, phi float64) (fromArch, fromPier float64) {
	speed := math.Hypot((a.A+d)*math.Sin(phi), (a.B+d)*math.Cos(phi))
	step := 2 * math.Pi / float64(a.Bays)
	u := math.Mod(phi+2*math.Pi, step)
	fromPier = math.Min(u, step-u) * speed
	fromArch = math.Abs(u-step/2) * speed
	return
}

func (a Arena) Build(g *Grid) Order {
	if a.A == 0 {
		a.A, a.B = 21, 13
	}
	a.Depth = orDefault(a.Depth, 26)
	if a.Bays == 0 {
		a.Bays = 80
	}
	const storey, attic = 6.5, 5.0
	outerTop := 3*storey + attic
	outer := a.Depth
	extent := Box{V(-(a.A + outer + 3), 0, -(a.B + outer + 3)), V(a.A+outer+3, outerTop+1, a.B+outer+3)}
	field := func(f func(p Vec3, d, phi float64) bool) Func {
		return Func{extent, func(p Vec3) bool {
			d, phi := a.ring(p.Plan())
			return f(p, d, phi)
		}}
	}
	// Ruin: the southern outer ring stands only one storey, with a ragged edge.
	ruined := func(p Vec3, phi float64) bool {
		if phi < 0.35 || phi > 1.45 {
			return false
		}
		edge := math.Min(phi-0.35, 1.45-phi) * 30
		rag := Hash01(int(phi*60), int(p.Y)) * 3
		return p.Y > storey+math.Min(edge, outerTop)+rag-3
	}
	arcade := func(p Vec3, d, phi, top float64) bool {
		level := math.Floor(p.Y / storey)
		if p.Y >= top || level >= 3 || p.Y < 1 {
			return false
		}
		local := p.Y - level*storey - 0.5
		fromArch, _ := a.bay(d, phi)
		rise := local - 3.9
		return fromArch <= 1.1 && (rise <= 0 || fromArch*fromArch+rise*rise <= 1.21) && local >= 0
	}
	seat := func(d float64) float64 { return 4 + math.Floor(d) }
	// Ground slab and a stepped podium around the outer ring.
	g.Use(Grass)
	g.Fill(field(func(p Vec3, d, _ float64) bool {
		return p.Y <= 0.5 && d <= outer+2 || p.Y <= 1 && d >= outer && d <= outer+1.5
	}))
	g.Paint(Sand, field(func(p Vec3, d, _ float64) bool { return d <= 1 }))
	g.Use(Sand)
	// Travertine outer wall: three arcaded storeys and an attic with small windows, cornice bands.
	g.Fill(field(func(p Vec3, d, phi float64) bool {
		if p.Y > outerTop || ruined(p, phi) {
			return false
		}
		cornice := d > outer && d <= outer+0.6 && math.Mod(p.Y, storey) < 0.6 && p.Y > 1 && p.Y < outerTop
		if cornice {
			return true
		}
		if d < outer-2.5 || d > outer {
			return false
		}
		if arcade(p, d, phi, 3*storey) {
			return false
		}
		if p.Y > 3*storey+1.5 && p.Y < 3*storey+3 {
			// Attic windows in every other bay.
			fromArch, _ := a.bay(d, phi)
			bayIndex := int(math.Floor((phi + math.Pi) / (2 * math.Pi) * float64(a.Bays)))
			return !(fromArch < 0.6 && bayIndex%2 == 0)
		}
		return true
	}))
	g.Paint(Stone, field(func(p Vec3, d, _ float64) bool { return p.Y <= 1 && d >= outer }))
	// Inner ring gallery: two arcaded storeys.
	g.Use(Stone)
	g.Fill(field(func(p Vec3, d, phi float64) bool {
		return d >= 19.5 && d <= 20.5 && p.Y <= 2*storey && !arcade(p, d, phi, 2*storey)
	}))
	// Podium wall around the arena floor.
	g.Fill(field(func(p Vec3, d, _ float64) bool { return d >= 0 && d <= 1 && p.Y <= 5 }))
	// Brick radial walls carrying the cavea, aligned with every other pier.
	g.Use(Brick)
	g.Fill(field(func(p Vec3, d, phi float64) bool {
		if d < 1 || d > outer-2.5 || p.Y < 0.5 || ruined(p, phi) && d > outer-6 {
			return false
		}
		_, fromPier := a.bay(d, phi)
		index := int(math.Round((phi + math.Pi) / (2 * math.Pi) * float64(a.Bays)))
		limit := seat(math.Min(d, 20))
		if d > 20 {
			limit = 2*storey + 0.5
		}
		return index%2 == 0 && fromPier <= 0.5 && p.Y <= limit
	}))
	// Marble cavea: a stepped seating shell one unit thick, rising 1:1 from the podium wall.
	g.Use(White)
	g.Fill(field(func(p Vec3, d, phi float64) bool {
		if d < 1 || d > outer-2.5 || ruined(p, phi) && d > outer-6 {
			return false
		}
		h := seat(math.Min(d, 20))
		if d > 20 { // Upper gallery floor behind the seats, level with the top row.
			h = 2*storey + 0.5
		}
		return p.Y <= h && p.Y > h-1.5
	}))
	// Hypogeum: brick corridors under the arena; a wooden floor covers only the west half.
	g.Use(Brick)
	g.Fill(field(func(p Vec3, d, _ float64) bool {
		if d > 0 || p.Y < 0.5 {
			return false
		}
		walls := math.Mod(math.Abs(p.Z)+0.5, 3) < 1 && p.Y <= 3
		deck := p.X < 0 && p.Y > 2.5 && p.Y <= 3.5
		return walls || deck
	}))
	g.Paint(Wood, field(func(p Vec3, d, _ float64) bool { return d <= 0 && p.Y > 2.5 && p.Y <= 3.5 }))
	return Sweep(P(0, 0), math.Pi/2)
}
