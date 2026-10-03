package scenes

import (
	"math"

	"godocta/gen/kit"
	. "godocta/gen/octa"
)

// MedievalTower is a tall donjon on a motte: a hollow earthen mound ringed by a shell
// wall, the tower with spiralling arrow slits, a stair turret and four bartizans.
type MedievalTower struct{}

func init() { Register(MedievalTower{}) }

func (MedievalTower) Name() string  { return "medieval_tower" }
func (MedievalTower) Title() string { return "Medieval Tower" }

func (MedievalTower) Build() (*Grid, Order) {
	const hill, top = 8.0, 46.0
	center := P(0, 0)
	g := NewGrid()
	// The motte: a smooth grassy frustum shell, so its slope reads cleanly.
	g.Use(Grass).FillVolume(Subtract(Frustum{center, 24, 14, 0, hill}, Frustum{center, 22, 12.5, 0, hill - 1.5}))
	g.Use(Stone)
	// Shell wall around the summit, with a gate facing +Z and a ramp down the mound.
	g.Fill(Cylinder{Center: center, Radius: 13.5, Inner: 12, Y0: hill, Y1: hill + 5})
	g.Build(kit.RingBattlements{Center: center, Radius: 13.5, Top: hill + 5, Rim: 1.5, Count: 28})
	g.Carve(Arch{V(0, hill, 12.75), 2, 2.5, 4, AxisZ})
	g.With(Sand, func() {
		g.Fill(Func{Box{V(-2, 0, 13), V(2, hill, 26)}, func(p Vec3) bool { return p.Y <= hill-(p.Z-13)*hill/13 }})
	})
	// The donjon itself.
	g.Build(kit.Tower{Center: center, Radius: 6.5, Base: hill, Top: top, Storey: 6.5, Roof: kit.RoofCone, RoofHeight: 15, Merlons: 18, Windows: true})
	g.Carve(Arch{V(0, hill, 6), 1.5, 3, 3, AxisZ})
	// Stair turret attached to the north-west, rising above the parapet.
	stair := Polar(center, 7, -3*math.Pi/4)
	g.Build(kit.Turret{Center: stair, Radius: 2.5, Base: hill, Top: top + 6, RoofHeight: 6, RoofColor: Terracotta})
	// Bartizans cantilevered at the parapet's diagonals.
	for k := 0; k < 4; k++ {
		at := Polar(center, 7.6, math.Pi/4+float64(k)*math.Pi/2)
		if k == 2 {
			continue // The stair turret occupies this side.
		}
		g.Build(kit.Turret{Center: at, Radius: 1.8, Base: top - 2, Top: top + 3, RoofHeight: 4.5, Cantilever: 3, RoofColor: Red})
	}
	return g, Sweep(center, math.Pi/2)
}
