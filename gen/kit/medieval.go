// Package kit holds reusable architectural Components built from octa Solids.
// Each component is a plain struct: set its fields, then grid.Build(component).
// Dimensions are world units, independent of the grid's cell span.
// Masonry uses the grid's current brush (grid.Use); accents such as roofs have their
// own Color fields, where the zero value selects the component's default.
package kit

import (
	"math"

	. "godocta/gen/octa"
)

// Battlements add a parapet rim around a flat rectangular top, with merlons.
type Battlements struct {
	Min, Max Vec2    // Plan footprint of the top.
	Top      float64 // Height of the surface the parapet stands on.
	Rim      float64 // Parapet thickness. Default 1.
	Period   float64 // Merlon + crenel length. Default 3.
	Merlon   float64 // Merlon length. Default Period / 2.
}

func (b Battlements) Build(g *Grid) {
	rim, period := or(b.Rim, 1), or(b.Period, 3)
	merlon := or(b.Merlon, period/2)
	g.Fill(Func{Box{b.Min.At(b.Top + 0.5), b.Max.At(b.Top + 2.5)}, func(p Vec3) bool {
		onX := p.X <= b.Min.X+rim || p.X >= b.Max.X-rim
		onZ := p.Z <= b.Min.Y+rim || p.Z >= b.Max.Y-rim
		switch {
		case !onX && !onZ:
			return false
		case p.Y <= b.Top+1.5:
			return true
		}
		return (onZ && math.Mod(p.X-b.Min.X, period) < merlon) || (onX && math.Mod(p.Z-b.Min.Y, period) < merlon)
	}})
}

// RingBattlements add a circular parapet with Count merlons.
type RingBattlements struct {
	Center Vec2
	Radius float64 // Outer radius.
	Top    float64
	Rim    float64 // Default 1.
	Count  int     // Default 12.
}

func (r RingBattlements) Build(g *Grid) {
	rim := or(r.Rim, 1)
	count := r.Count
	if count == 0 {
		count = 12
	}
	g.Fill(Cylinder{r.Center, r.Radius, r.Radius - rim, r.Top + 0.5, r.Top + 1.5})
	g.Fill(Func{Box{V(r.Center.X-r.Radius, r.Top+2, r.Center.Y-r.Radius), V(r.Center.X+r.Radius, r.Top+2.5, r.Center.Y+r.Radius)}, func(p Vec3) bool {
		d := p.Plan().Sub(r.Center)
		turn := math.Mod(d.Angle()/(2*math.Pi)*float64(count)*2+float64(count)*2, 2)
		return d.Length() <= r.Radius && d.Length() > r.Radius-rim && turn < 1
	}})
}

// Wall is an axis-aligned crenellated curtain wall with an optional hollow gallery.
type Wall struct {
	Min, Max Vec3
	Gallery  bool
}

func (w Wall) Build(g *Grid) {
	g.Fill(Block(w.Min, w.Max))
	if w.Gallery {
		g.Carve(Block(w.Min.Add(V(1, 1, 1)), w.Max.Sub(V(1, 1.5, 1))))
	}
	g.Build(Battlements{Min: w.Min.Plan(), Max: w.Max.Plan(), Top: w.Max.Y})
}

// Roof styles for towers.
type Roof int

const (
	RoofBattlements Roof = iota // Open fighting platform with merlons.
	RoofCone                    // Conical roof behind a parapet, with a finial.
)

// Tower is a round tower with hollow storeys, a corbelled parapet and a roof.
type Tower struct {
	Center     Vec2
	Radius     float64
	Base, Top  float64 // Top is the walkway level.
	Storey     float64 // Storey height. Default 6.
	Roof       Roof
	RoofHeight float64 // For RoofCone. Default 1.8 × Radius.
	Merlons    int
	Windows    bool  // Arrow slits on every storey, turning as they rise.
	RoofColor  Color // Default Slate; the finial is Gold.
}

func (t Tower) Build(g *Grid) {
	storey := or(t.Storey, 6)
	g.Fill(Cylinder{Center: t.Center, Radius: t.Radius, Y0: t.Base, Y1: t.Top})
	level := 0
	for y := t.Base + 0.5; y+storey-1 < t.Top-1; y += storey {
		g.Carve(Cylinder{Center: t.Center, Radius: t.Radius - 1.5, Y0: y, Y1: y + storey - 1})
		if t.Windows && level > 0 {
			for k := 0; k < 4; k++ {
				angle := float64(k)*math.Pi/2 + float64(level)*0.6
				g.Carve(radialSlit(t.Center, t.Radius, angle, y+1.5, 2.5))
			}
		}
		level++
	}
	rim := t.Radius + 0.8
	// Corbelled machicolation ring under the parapet.
	g.Fill(Frustum{t.Center, t.Radius, rim, t.Top - 2.5, t.Top - 1.5}, Cylinder{Center: t.Center, Radius: rim, Y0: t.Top - 1.5, Y1: t.Top})
	g.Build(RingBattlements{Center: t.Center, Radius: rim, Top: t.Top, Count: t.Merlons})
	if t.Roof == RoofCone {
		height := or(t.RoofHeight, 1.8*t.Radius)
		g.With(t.RoofColor.Or(Slate), func() { g.Fill(Cone(t.Center, rim-0.6, t.Top+0.5, height)) })
		g.With(Gold, func() { g.Fill(Cylinder{Center: t.Center, Radius: 0.75, Y0: t.Top + height, Y1: t.Top + height + 1.5}) })
	}
}

// radialSlit is a narrow window through a round wall at the given plan angle.
func radialSlit(center Vec2, radius, angle, sill, height float64) Solid {
	slit := Arch{Sill: V(radius-0.75, sill, 0), HalfWidth: 0.5, Spring: height - 0.5, Depth: 3, Through: AxisX}
	return Translate(RotateY(slit, P(0, 0), angle), center.At(0))
}

// Turret is a slender tower with a conical roof, optionally cantilevered from a wall
// on a corbel cone (Cantilever > 0 is the corbel height below Base).
type Turret struct {
	Center     Vec2
	Radius     float64
	Base, Top  float64
	RoofHeight float64 // Default 2.5 × Radius.
	Cantilever float64
	RoofColor  Color // Default Slate.
}

func (t Turret) Build(g *Grid) {
	if t.Cantilever > 0 {
		g.Fill(Frustum{t.Center, 0.6, t.Radius, t.Base - t.Cantilever, t.Base})
	}
	g.Fill(Cylinder{Center: t.Center, Radius: t.Radius, Y0: t.Base, Y1: t.Top})
	g.Fill(Cylinder{Center: t.Center, Radius: t.Radius + 0.6, Y0: t.Top - 1, Y1: t.Top})
	height := or(t.RoofHeight, 2.5*t.Radius)
	g.With(t.RoofColor.Or(Slate), func() { g.Fill(Cone(t.Center, t.Radius+0.6, t.Top, height)) })
}

// Keep is a rectangular great tower: storeys, windows, a door, battlements and
// corner turrets that rise above the parapet.
type Keep struct {
	Min, Max  Vec3
	Storeys   int     // Default 3.
	Door      Axis    // The door faces +Z (AxisZ) or +X (AxisX).
	Turret    float64 // Corner turret radius; 0 for none.
	RoofColor Color   // Turret roofs. Default Slate.
}

func (k Keep) Build(g *Grid) {
	storeys := k.Storeys
	if storeys == 0 {
		storeys = 3
	}
	g.Fill(Block(k.Min, k.Max))
	height := (k.Max.Y - k.Min.Y) / float64(storeys)
	for s := 0; s < storeys; s++ {
		y := k.Min.Y + float64(s)*height
		g.Carve(Block(k.Min.Add(V(2, 0.5, 2)).Add(V(0, float64(s)*height, 0)), V(k.Max.X-2, y+height-1, k.Max.Z-2)))
		if s == 0 {
			continue
		}
		sill := y + height*0.35
		for x := k.Min.X + 5; x <= k.Max.X-5; x += 5 {
			g.Carve(Arch{V(x, sill, k.Min.Z), 0.9, 2, 3, AxisZ}, Arch{V(x, sill, k.Max.Z), 0.9, 2, 3, AxisZ})
		}
		for z := k.Min.Z + 5; z <= k.Max.Z-5; z += 5 {
			g.Carve(Arch{V(k.Min.X, sill, z), 0.9, 2, 3, AxisX}, Arch{V(k.Max.X, sill, z), 0.9, 2, 3, AxisX})
		}
	}
	if k.Door == AxisX {
		g.Carve(Arch{V(k.Max.X, k.Min.Y, (k.Min.Z+k.Max.Z)/2), 1.8, 3.5, 3, AxisX})
	} else {
		g.Carve(Arch{V((k.Min.X+k.Max.X)/2, k.Min.Y, k.Max.Z), 1.8, 3.5, 3, AxisZ})
	}
	g.Build(Battlements{Min: k.Min.Plan(), Max: k.Max.Plan(), Top: k.Max.Y})
	if k.Turret > 0 {
		for _, x := range []float64{k.Min.X, k.Max.X} {
			for _, z := range []float64{k.Min.Z, k.Max.Z} {
				g.Build(Turret{Center: P(x, z), Radius: k.Turret, Base: k.Min.Y, Top: k.Max.Y + 5, RoofColor: k.RoofColor})
			}
		}
	}
}

// Hall is a hollow gabled building with a door at the positive end of its ridge axis.
type Hall struct {
	Min, Max  Vec3  // Walls; the roof rises above Max.Y.
	Ridge     Axis  // Zero value AxisX; set AxisZ explicitly for a Z-aligned ridge.
	RoofColor Color // Default Terracotta.
}

func (h Hall) Build(g *Grid) {
	g.Fill(Block(h.Min, h.Max))
	g.Carve(Block(h.Min.Add(V(1.5, 0.5, 1.5)), h.Max.Sub(V(1.5, 1, 1.5))))
	g.With(h.RoofColor.Or(Terracotta), func() {
		g.Fill(Gable{h.Min.Plan().Sub(P(0.5, 0.5)), h.Max.Plan().Add(P(0.5, 0.5)), h.Max.Y, h.Ridge})
	})
	center := h.Min.Add(h.Max).Scale(0.5)
	if h.Ridge == AxisX {
		g.Carve(Arch{V(h.Max.X, h.Min.Y, center.Z), 1.4, 3, 3, AxisX})
	} else {
		g.Carve(Arch{V(center.X, h.Min.Y, h.Max.Z), 1.4, 3, 3, AxisZ})
	}
}

// Gatehouse is a crenellated block astride a wall, flanked by two round towers,
// with an arched passage along Z.
type Gatehouse struct {
	Center      Vec2    // Plan center of the passage.
	Width       float64 // Along X. Default 14.
	Depth       float64 // Along Z. Default 10.
	Base, Top   float64
	TowerRadius float64 // Default 4.
}

func (gh Gatehouse) Build(g *Grid) {
	width, depth, radius := or(gh.Width, 14), or(gh.Depth, 10), or(gh.TowerRadius, 4)
	low := V(gh.Center.X-width/2, gh.Base, gh.Center.Y-depth/2)
	high := V(gh.Center.X+width/2, gh.Top, gh.Center.Y+depth/2)
	g.Fill(Block(low, high))
	g.Carve(Block(low.Add(V(1.5, 1, 1.5)), high.Sub(V(1.5, 1.5, 1.5))))
	g.Build(Battlements{Min: low.Plan(), Max: high.Plan(), Top: gh.Top})
	for _, side := range []float64{-1, 1} {
		g.Build(Tower{Center: P(gh.Center.X+side*width/2, high.Z-radius*0.5), Radius: radius, Base: gh.Base, Top: gh.Top + 3, Merlons: 12})
	}
	g.Carve(Arch{V(gh.Center.X, gh.Base, gh.Center.Y), 2.5, 5.5, depth + 6, AxisZ})
	for _, z := range []float64{low.Z, high.Z} {
		for _, x := range []float64{-width / 4, width / 4} {
			g.Carve(Arch{V(gh.Center.X+x, gh.Base+(gh.Top-gh.Base)*0.6, z), 0.8, 2, 2, AxisZ})
		}
	}
}

// Moat removes a square channel from a platform, leaving a bridge across +Z.
type Moat struct {
	Inner, Outer float64 // Max-norm distance range from Center.
	Center       Vec2
	Bottom, Top  float64
	Bridge       float64 // Bridge half-width; 0 for none.
	Water        bool    // Paint the channel floor Water.
}

func (m Moat) Build(g *Grid) {
	channel := func(bottom, top float64) Func {
		return Func{Box{V(m.Center.X-m.Outer, bottom, m.Center.Y-m.Outer), V(m.Center.X+m.Outer, top, m.Center.Y+m.Outer)}, func(p Vec3) bool {
			d := p.Plan().Sub(m.Center)
			reach := math.Max(math.Abs(d.X), math.Abs(d.Y))
			bridge := math.Abs(d.X) <= m.Bridge && d.Y > 0
			return reach >= m.Inner && reach <= m.Outer && !bridge
		}}
	}
	g.Carve(channel(m.Bottom, m.Top))
	if m.Water {
		g.Paint(Water, channel(m.Bottom-0.5, m.Bottom-0.5))
	}
}

func or(v, fallback float64) float64 {
	if v == 0 {
		return fallback
	}
	return v
}
