package scenes

import (
	"math"

	"godocta/gen/kit"
	. "godocta/gen/octa"
)

// Manhattan is a stylised island city: a street grid of setback blocks whose heights
// peak in Midtown and Downtown, Central Park with its reservoir and trees, landmark
// towers, the Brooklyn Bridge and a small Liberty Island. North is -Z.
type Manhattan struct{}

func init() { Register(Manhattan{}) }

func (Manhattan) Name() string  { return "manhattan" }
func (Manhattan) Title() string { return "Manhattan" }

var island = []Vec2{
	{-4, 80}, {6, 78}, {14, 70}, {19, 60}, {20, 50}, {18, 30}, {19, 10}, {17, -10}, {14, -30},
	{11, -50}, {9, -65}, {5, -80}, {-6, -80}, {-10, -65}, {-14, -50}, {-17, -30}, {-19, -10},
	{-20, 10}, {-19, 30}, {-16, 50}, {-14, 60}, {-10, 70},
}

func (Manhattan) Build() (*Grid, Order) {
	const ground = 1.0
	g := NewGrid()
	g.Fill(Prism{island, 0, ground - 0.5})
	park := Block(V(-9, ground, -56), V(6, ground, -24))
	landmarks := []Box{
		BoxOf(V(-4, 0, -12), V(4, 0, -4)),  // Empire State
		BoxOf(V(5, 0, -18), V(11, 0, -12)), // Chrysler
		BoxOf(V(-10, 0, 60), V(-2, 0, 68)), // One World Trade Center
		BoxOf(V(-2, 0, 16), V(6, 0, 24)),   // Flatiron
	}
	reserved := func(lo, hi Vec2) bool {
		lot := Box{lo.At(0), hi.At(0)}
		if !lot.Intersect(Box(park).Grow(0.5)).Empty() {
			return true
		}
		for _, l := range landmarks {
			if !lot.Intersect(l.Grow(0.5)).Empty() {
				return true
			}
		}
		return false
	}
	// Street grid: avenues every 9 units along X, streets every 5 along Z.
	for x := -22.0; x < 22; x += 9 {
		for z := -82.0; z < 82; z += 5 {
			for half := 0; half < 2; half++ {
				lo := P(x+1+float64(half)*3.5, z+1)
				hi := P(lo.X+3, z+4)
				corners := []Vec2{lo, hi, P(lo.X, hi.Y), P(hi.X, lo.Y)}
				onIsland := true
				for _, c := range corners {
					onIsland = onIsland && InsidePolygon(island, c.Sub(c.Sub(lo.Add(hi).Scale(0.5)).Scale(0.15)))
				}
				if !onIsland || reserved(lo, hi) {
					continue
				}
				g.Build(cityBlock(lo, hi, ground))
			}
		}
	}
	// Central Park: lawn, reservoir and deterministic trees.
	g.Fill(Block(V(-9, ground-0.5, -56), V(6, ground, -24)))
	g.Carve(EllipticCylinder{P(-1.5, -44), 4.5, 3.5, ground - 0.5, ground})
	for i := 0; i < 60; i++ {
		at := P(-8+Hash01(i, 1)*13, -55+Hash01(i, 2)*30)
		if (at.X+1.5)*(at.X+1.5)/30+(at.Y+44)*(at.Y+44)/18 < 1 {
			continue
		}
		g.Build(kit.Tree{At: at, Base: ground, Height: 2.5 + Hash01(i, 3)*1.5})
	}
	// Landmarks.
	g.Build(
		kit.Building{Min: P(-4, -12), Max: P(4, -4), Base: ground, Crown: kit.CrownMast, CrownHeight: 12,
			Tiers: []kit.Tier{{Height: 6}, {Height: 10, Inset: 1}, {Height: 26, Inset: 2}, {Height: 4, Inset: 2.5}}},
		kit.Building{Min: P(5, -18), Max: P(11, -12), Base: ground, Crown: kit.CrownSpire, CrownHeight: 14,
			Tiers: []kit.Tier{{Height: 8}, {Height: 26, Inset: 0.5}, {Height: 4, Inset: 1}}},
		kit.TwistTower{Center: P(-6, 64), Half: 4, Base: ground, Top: ground + 50, SpireHeight: 12},
		kit.Wedge{Corners: []Vec2{{-2, 24}, {6, 24}, {1, 16}}, Base: ground, Top: ground + 12},
	)
	// Brooklyn Bridge out to a pier on the east, and Liberty Island to the south-west.
	g.Fill(Block(V(42, 0, 60), V(50, ground-0.5, 72)))
	g.Build(kit.SuspensionBridge{From: P(17, 66), To: P(43, 66), Deck: ground + 3, Width: 4, TowerHeight: 11})
	g.Build(liberty(P(-30, 92)))
	return g, Outward(P(0, -8))
}

// skyline gives the target height at a plan point: peaks in Midtown and Downtown.
func skyline(p Vec2) float64 {
	midtown := math.Exp(-math.Pow((p.Y+10)/16, 2))
	downtown := math.Exp(-math.Pow((p.Y-62)/9, 2))
	return 3 + 26*midtown + 22*downtown
}

// cityBlock places one lot: a podium with setbacks, sized by the skyline and varied
// deterministically per lot.
func cityBlock(lo, hi Vec2, ground float64) Component {
	center := lo.Add(hi).Scale(0.5)
	seed := int(center.X*7) + int(center.Y*13)*31
	height := math.Max(2, skyline(center)*(0.45+Hash01(seed)*0.9))
	height = math.Round(height*2) / 2
	b := kit.Building{Min: lo, Max: hi, Base: ground}
	switch {
	case height > 18:
		b.Tiers = []kit.Tier{{Height: height * 0.3}, {Height: height * 0.5, Inset: 0.5}, {Height: height * 0.2, Inset: 1}}
		if Hash01(seed, 9) > 0.7 {
			b.Crown = kit.CrownPyramid
		}
	case height > 9:
		b.Tiers = []kit.Tier{{Height: height * 0.6}, {Height: height * 0.4, Inset: 0.5}}
	default:
		b.Tiers = []kit.Tier{{Height: height}}
	}
	return b
}

// liberty is a star-fort island with a pedestal and a stylised statue.
func liberty(at Vec2) Component {
	return ComponentFunc(func(g *Grid) {
		g.Fill(EllipticCylinder{at, 7, 5, 0, 0.5})
		var star []Vec2
		for k := 0; k < 22; k++ {
			r := 4.5
			if k%2 == 1 {
				r = 2.8
			}
			star = append(star, Polar(at, r, float64(k)*2*math.Pi/22))
		}
		g.Fill(Prism{star, 1, 2})
		g.Fill(Frustum{at, 2, 1.5, 2, 6})
		g.Fill(Frustum{at, 1.2, 0.7, 6, 11}, Sphere(at.At(11.5), 0.9))
		g.Fill(Capsule{at.Add(P(0.6, 0)).At(10), at.Add(P(1, 0)).At(14), CellRadius})
		g.Fill(Sphere(at.Add(P(1, 0)).At(14.6), 0.8))
	})
}
