package kit

import (
	"math"

	. "godocta/gen/octa"
)

// Tier is one setback stage of a building: it rises Height above the previous tier
// and is inset by Inset on every side relative to the building footprint.
type Tier struct {
	Height, Inset float64
}

// Crown styles for building tops.
type Crown int

const (
	CrownFlat    Crown = iota
	CrownSpire         // Stacked shrinking frustums and a needle (Chrysler-like).
	CrownMast          // A slim stepped mast (Empire State-like).
	CrownPyramid       // A 45° hip roof.
)

// Building is a hollow, floor-slabbed block with setbacks and a crown.
// Small footprints are built solid; hollow interiors show up in the X-ray view.
type Building struct {
	Min, Max    Vec2 // Plan footprint.
	Base        float64
	Tiers       []Tier
	Crown       Crown
	CrownHeight float64
	Floor       float64 // Floor slab spacing for hollow buildings. Default 4.
	Windows     Color   // Glazing color for window bands between floors; 0 for none.
	CrownColor  Color   // Default: the facade (current brush).
}

// Plain is a single-tier building.
func Plain(min, max Vec2, base, height float64) Building {
	return Building{Min: min, Max: max, Base: base, Tiers: []Tier{{Height: height}}}
}

func (b Building) Build(g *Grid) {
	floor := or(b.Floor, 4)
	y := b.Base
	var lo, hi Vec2
	for _, t := range b.Tiers {
		lo, hi = b.Min.Add(P(t.Inset, t.Inset)), b.Max.Sub(P(t.Inset, t.Inset))
		if hi.X-lo.X < 1 || hi.Y-lo.Y < 1 {
			break
		}
		top := y + t.Height
		g.Fill(Block(lo.At(y), hi.At(top)))
		if b.Windows != 0 {
			g.Paint(b.Windows, windowBands(lo, hi, y, top, floor))
		}
		if hi.X-lo.X >= 5 && hi.Y-lo.Y >= 5 {
			for s := y + 0.5; s < top-1; s += floor {
				g.Carve(Block(lo.Add(P(1, 1)).At(s), hi.Sub(P(1, 1)).At(math.Min(s+floor-1, top-1))))
			}
		}
		y = top
	}
	center := lo.Add(hi).Scale(0.5)
	half := math.Min(hi.X-lo.X, hi.Y-lo.Y) / 2
	previous := g.Brush()
	defer g.Use(previous)
	g.Use(b.CrownColor.Or(previous))
	switch b.Crown {
	case CrownPyramid:
		g.Fill(HipRoof{lo, hi, y})
	case CrownSpire:
		height := or(b.CrownHeight, 4*half)
		steps := 5
		for i := 0; i < steps; i++ {
			r0 := half * (1 - float64(i)/float64(steps))
			g.Fill(Frustum{center, r0, r0 * 0.75, y + height*float64(i)/float64(steps)*0.7, y + height*float64(i+1)/float64(steps)*0.7})
		}
		g.Fill(Capsule{center.At(y), center.At(y + height), 0.6})
	case CrownMast:
		height := or(b.CrownHeight, 3*half)
		g.Fill(Cylinder{Center: center, Radius: math.Max(half*0.45, 1), Y0: y, Y1: y + height*0.4})
		g.Fill(Cylinder{Center: center, Radius: math.Max(half*0.25, 0.8), Y0: y, Y1: y + height*0.7})
		g.Fill(Capsule{center.At(y), center.At(y + height), 0.6})
	}
}

// windowBands selects the outer skin between floor lines, leaving corner piers.
func windowBands(lo, hi Vec2, y0, y1, floor float64) Solid {
	return Func{Box{lo.At(y0), hi.At(y1)}, func(p Vec3) bool {
		nearX := p.X <= lo.X+1 || p.X >= hi.X-1
		nearZ := p.Z <= lo.Y+1 || p.Z >= hi.Y-1
		band := math.Mod(p.Y-y0, floor) >= 1 && p.Y < y1-0.5
		return band && nearX != nearZ
	}}
}

// TwistTower is a square tower whose corners chamfer as it rises, ending as a square
// rotated 45° (One World Trade Center-like), with a hollow core and a spire.
type TwistTower struct {
	Center      Vec2
	Half        float64 // Half the base side.
	Base, Top   float64
	SpireHeight float64
	Windows     Color // Glazing between floor lines; 0 for none.
}

func (t TwistTower) contains(inset float64) func(Vec3) bool {
	return func(p Vec3) bool {
		d := p.Plan().Sub(t.Center)
		ax, az := math.Abs(d.X), math.Abs(d.Y)
		u := (p.Y - t.Base) / (t.Top - t.Base)
		return math.Max(ax, az) <= t.Half-inset && ax+az <= (t.Half-inset)*(2-u)
	}
}

func (t TwistTower) Build(g *Grid) {
	box := Box{V(t.Center.X-t.Half, t.Base, t.Center.Y-t.Half), V(t.Center.X+t.Half, t.Top, t.Center.Y+t.Half)}
	g.Fill(Func{box, t.contains(0)})
	if t.Windows != 0 {
		g.Paint(t.Windows, Func{box, func(p Vec3) bool { return math.Mod(p.Y-t.Base, 4) >= 1 }})
	}
	for y := t.Base + 0.5; y < t.Top-1.5; y += 4 {
		inner := box
		inner.Min.Y, inner.Max.Y = y, y+3
		g.Carve(Func{inner, t.contains(1)})
	}
	if t.SpireHeight > 0 {
		g.Fill(Capsule{t.Center.At(t.Top), t.Center.At(t.Top + t.SpireHeight), 0.6})
	}
}

// Wedge is a triangular-plan building (Flatiron-like) given by its plan corners.
type Wedge struct {
	Corners   []Vec2
	Base, Top float64
}

func (w Wedge) Build(g *Grid) {
	g.Fill(Prism{w.Corners, w.Base, w.Top})
	g.Fill(Prism{w.Corners, w.Top, w.Top + 1})
}

// Tree is a lollipop tree: a trunk and an ellipsoid crown.
type Tree struct {
	At     Vec2
	Base   float64
	Height float64 // Default 3.
}

func (t Tree) Build(g *Grid) {
	h := or(t.Height, 3)
	g.With(Wood, func() { g.Fill(Cylinder{Center: t.At, Radius: 0.6, Y0: t.Base, Y1: t.Base + h*0.5}) })
	g.With(Foliage, func() { g.Fill(Ellipsoid{t.At.At(t.Base + h*0.7), V(h*0.35, h*0.3, h*0.35)}) })
}

// SuspensionBridge spans between two plan points: a deck, two arched stone towers
// and catenary cables sampled into connected cell lines.
type SuspensionBridge struct {
	From, To    Vec2
	Deck        float64 // Deck height.
	Width       float64 // Default 4.
	TowerHeight float64 // Above the deck. Default 12.
}

func (s SuspensionBridge) Build(g *Grid) {
	width, towerHeight := or(s.Width, 4), or(s.TowerHeight, 12)
	axis := s.To.Sub(s.From)
	length := axis.Length()
	dir := axis.Scale(1 / length)
	side := P(-dir.Y, dir.X)
	angle := dir.Angle()
	at := func(t, offset float64) Vec2 { return s.From.Add(dir.Scale(t)).Add(side.Scale(offset)) }
	deck := Block(V(0, s.Deck-0.5, -width/2), V(length, s.Deck, width/2))
	g.With(Charcoal, func() { g.Fill(Translate(RotateY(deck, P(0, 0), angle), s.From.At(0))) })
	towers := []float64{length * 0.22, length * 0.78}
	for _, t := range towers {
		c := at(t, 0)
		tower := Block(V(-1.5, 0, -width/2-1.5), V(1.5, s.Deck+towerHeight, width/2+1.5))
		gate := Arch{V(0, s.Deck, 0), width/2 - 0.5, towerHeight * 0.45, 4, AxisX}
		g.Fill(Translate(RotateY(Subtract(tower, gate), P(0, 0), angle), c.At(0)))
	}
	// Main cables: a parabola between tower tops, sagging to the deck mid-span.
	top := s.Deck + towerHeight - 0.5
	for _, offset := range []float64{-width/2 - 0.5, width/2 + 0.5} {
		var points []Vec3
		for i := 0; i <= 80; i++ {
			t := float64(i) / 80 * length
			var y float64
			switch {
			case t < towers[0]:
				y = s.Deck + (top-s.Deck)*t/towers[0]
			case t > towers[1]:
				y = s.Deck + (top-s.Deck)*(length-t)/(length-towers[1])
			default:
				u := (t-towers[0])/(towers[1]-towers[0])*2 - 1
				y = s.Deck + 1 + (top-s.Deck-1)*u*u
			}
			points = append(points, at(t, offset).At(y))
		}
		for i := 1; i < len(points); i++ {
			g.With(Charcoal, func() { g.Fill(Capsule{points[i-1], points[i], CellRadius}) })
		}
	}
}
