package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// CruiseShip is a modern liner: a hollow hull with a raked bow, rounded bilges and
// porthole rows; terraced superstructure decks with balcony bays; a recessed promenade
// with lifeboats; a bridge with wings; pools; a raked funnel and a radar mast.
// Bow points to +X.
type CruiseShip struct {
	Length, Beam, Depth float64 // Defaults 120 × 18 × 10.
}

func init() { Register(CruiseShip{}) }

func (CruiseShip) Name() string  { return "cruise_ship" }
func (CruiseShip) Title() string { return "Cruise Ship" }

// hullHalfWidth is the plan half-width at x and height y, with an inset for hollowing.
func (s CruiseShip) hullHalfWidth(x, y, inset float64) float64 {
	half := s.Beam / 2
	stern, bowStart := -s.Length/2, s.Length/6
	bowEnd := s.Length/2 - (s.Depth-y)*0.8 // Raked stem: the keel is shorter than the deck.
	var w float64
	switch {
	case x < stern || x > bowEnd:
		return -1
	case x > bowStart:
		t := (x - bowStart) / (bowEnd - bowStart)
		w = half * (1 - t*t)
	case x < stern+15:
		t := (stern + 15 - x) / 15
		w = half - 1.5*t*t
	default:
		w = half
	}
	if y < 4 { // Rounded bilge.
		t := (4 - y) / 4
		w *= 1 - 0.55*t*t
	}
	return w - inset
}

func (s CruiseShip) Build() (*Grid, Order) {
	if s.Length == 0 {
		s.Length, s.Beam, s.Depth = 120, 18, 10
	}
	L, B, D := s.Length, s.Beam, s.Depth
	hull := func(inset, floor, top float64) Func {
		return Func{Box{V(-L/2, floor, -B/2), V(L/2, top, B/2)}, func(p Vec3) bool {
			return math.Abs(p.Z) <= s.hullHalfWidth(p.X, p.Y, inset)
		}}
	}
	g := NewGrid()
	g.FillVolume(hull(0, 0, D))
	g.Carve(hull(1.2, 1.2, D-0.6))
	// Internal decks, clipped to the hull.
	for y := 4.0; y < D-1; y += 3 {
		g.Fill(Intersect(hull(0, y, y+0.5)))
	}
	// Porthole rows: recesses in the outer skin.
	g.Carve(Func{Box{V(-L/2, 5, -B/2), V(L/2, D-1, B/2)}, func(p Vec3) bool {
		w := s.hullHalfWidth(p.X, p.Y, 0)
		row := math.Mod(p.Y, 2.5) < 0.6
		return row && math.Abs(p.Z) > w-0.6 && math.Mod(p.X+L, 2) < 0.6
	}})
	// Superstructure: seven decks, terraced aft and stepped back at the front.
	const deck = 2.5
	var top float64
	for k := 0; k < 7; k++ {
		y0 := D + float64(k)*deck
		aft, fore := -L/2+8+float64(k)*2.5, L/2-22-float64(k)*1.8
		half := B/2 - 0.5
		if k == 1 || k == 2 {
			half = B/2 - 2.5 // Promenade decks, recessed behind the lifeboats.
		}
		front := Vec2{fore, 0}
		g.Fill(Func{Box{V(aft, y0, -half), V(fore+half, y0+deck, half)}, func(p Vec3) bool {
			if p.X <= fore {
				return true
			}
			return p.Plan().Sub(front).Length() <= half // Rounded front.
		}})
		g.Carve(Block(V(aft+1, y0+0.5, -half+1), V(fore-1, y0+deck-0.5, half-1)))
		// Balcony bays: recesses between thin dividers, under each deck slab.
		if k != 1 && k != 2 {
			g.Carve(Func{Box{V(aft+2, y0+0.5, -half), V(fore-2, y0+deck-0.5, half)}, func(p Vec3) bool {
				return math.Abs(p.Z) > half-1 && math.Mod(p.X+L, 3) > 0.6
			}})
		}
		top = y0 + deck
	}
	bridgeY := D + 6*deck
	foreTop := L/2 - 22 - 6*1.8
	// Bridge with wings out to the full beam, and a radar mast.
	g.Fill(Block(V(foreTop-4, bridgeY, -B/2-0.5), V(foreTop-1, bridgeY+deck, B/2+0.5)))
	g.Fill(Cylinder{Center: P(foreTop-8, 0), Radius: 0.8, Y0: top, Y1: top + 7})
	g.Fill(Block(V(foreTop-8.5, top+5, -3), V(foreTop-7.5, top+5.5, 3)))
	// Pools on the top deck.
	g.Carve(Block(V(-6, top-1, -4), V(6, top, 4)), Cylinder{Center: P(12, 0), Radius: 3, Y0: top - 1, Y1: top})
	// Lifeboats hanging along the promenade.
	for x := -L/2 + 18; x <= L/2-34; x += 6.5 {
		for _, side := range []float64{-1, 1} {
			g.Fill(Ellipsoid{V(x, D+deck+1.4, side*(B/2-1.2)), V(2.6, 1.0, 1.0)})
		}
	}
	// Raked funnel with a hollow top.
	funnel := Sheared{EllipticCylinder{P(-L/2+30, 0), 5, 2.8, top, top + 9}, P(-0.35, 0), top}
	g.Fill(funnel)
	g.Carve(Sheared{EllipticCylinder{P(-L/2+30, 0), 3.8, 1.6, top + 7.5, top + 9}, P(-0.35, 0), top})
	return g, Along(P(-1, 0))
}
