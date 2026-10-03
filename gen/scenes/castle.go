package scenes

import (
	"math"

	"godocta/gen/kit"
	. "godocta/gen/octa"
)

// Castle is a concentric fortress on a moated platform: curtain walls with galleries,
// corner towers with conical roofs, flanking towers, a gatehouse, a keep with corner
// turrets, a great hall and a chapel.
type Castle struct {
	Wall   float64 // Curtain wall centerline distance from the center. Default 25.
	Ground float64 // Platform top. Default 2.
}

func init() { Register(Castle{}) }

func (Castle) Name() string  { return "castle" }
func (Castle) Title() string { return "Castle" }

func (c Castle) Build() (*Grid, Order) {
	wall, ground := orDefault(c.Wall, 25), orDefault(c.Ground, 2)
	plinth := wall + 13
	g := NewGrid()
	g.Fill(Block(V(-plinth, 0, -plinth), V(plinth, ground-0.5, plinth)))
	g.Build(kit.Moat{Inner: wall + 8.5, Outer: wall + 11.5, Bottom: 0.5, Top: ground, Bridge: 3})
	for _, side := range []float64{-1, 1} {
		g.Build(
			kit.Wall{Min: V(-wall, ground, side*wall-2), Max: V(wall, ground+12, side*wall+2), Gallery: true},
			kit.Wall{Min: V(side*wall-2, ground, -wall), Max: V(side*wall+2, ground+12, wall), Gallery: true},
		)
	}
	for _, x := range []float64{-wall, wall} {
		for _, z := range []float64{-wall, wall} {
			g.Build(kit.Tower{Center: P(x, z), Radius: 6, Base: ground, Top: ground + 19, Roof: kit.RoofCone, RoofHeight: 11, Merlons: 16, Windows: true})
		}
	}
	for _, center := range []Vec2{P(-wall, 0), P(wall, 0), P(0, -wall)} {
		g.Build(kit.Tower{Center: center, Radius: 4.5, Base: ground, Top: ground + 15, Merlons: 12, Windows: true})
	}
	g.Build(
		kit.Gatehouse{Center: P(0, wall+1), Base: ground, Top: ground + 15},
		kit.Keep{Min: V(-10, ground, -14), Max: V(10, ground+24, 6), Turret: 2.8},
		kit.Hall{Min: V(-21, ground, -21), Max: V(-13, ground+7, -7)},
		kit.Hall{Min: V(13, ground, -21), Max: V(21, ground+6, -9)},
	)
	// Courtyard well.
	g.Fill(Cylinder{Center: P(12, 12), Radius: 1.8, Inner: 0.9, Y0: ground, Y1: ground + 1.5})
	return g, Sweep(P(0, 0), math.Pi/2)
}

func orDefault(v, fallback float64) float64 {
	if v == 0 {
		return fallback
	}
	return v
}
