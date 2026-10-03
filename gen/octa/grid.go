package octa

import (
	"fmt"
	"math"
	"sort"
)

// Component is a reusable piece of architecture that places itself into a grid.
type Component interface {
	Build(g *Grid)
}

// ComponentFunc adapts a function into a Component.
type ComponentFunc func(g *Grid)

func (f ComponentFunc) Build(g *Grid) { f(g) }

// Grid is a mutable set of occupied, colored lattice cells. Fill operations paint with
// the current brush (Stone until Use changes it).
//
// Scenes are authored in world units, with boundaries on half-unit steps. The cell span
// sets the resolution: span 1 puts cell centers exactly on those steps; a smaller span
// packs 1/span³ times as many cells into the same volume for finer surfaces.
type Grid struct {
	cells map[Cell]Color
	brush Color
	span  float64
	snap  bool
}

// NewGrid is a grid of span-1 cells: two cells per cubic world unit in a filled lattice.
func NewGrid() *Grid { return NewGridWithSpan(1) }

// NewGridWithSpan sets the cell span in world units without validating it.
// Structure validates the serialized span; spans below 1 enable half-unit snapping.
func NewGridWithSpan(span float64) *Grid {
	return &Grid{cells: make(map[Cell]Color), span: span, snap: span < 1}
}

// SpanForDensity is the span that packs density times as many cells as span 1.
func SpanForDensity(density float64) float64 { return math.Cbrt(1 / density) }

// Span is the cell span in world units.
func (g *Grid) Span() float64 { return g.span }

// World is a cell's center in world units.
func (g *Grid) World(c Cell) Vec3 { return c.Center().Scale(g.span) }

// each visits every cell whose world center lies inside the box, grown so that cells
// which snap into it (see inside) are visited too.
func (g *Grid) each(b Box, f func(Cell, Vec3)) {
	if g.snap {
		b = b.Grow(0.25)
	}
	inv := 1 / g.span
	eachCell(Box{b.Min.Scale(inv), b.Max.Scale(inv)}, func(c Cell, p Vec3) { f(c, p.Scale(g.span)) })
}

// inside tests a cell center against a solid. On finer lattices it also tests the center
// snapped to the half-unit authoring grid: components authored half a unit apart then
// meet without seams, and zero-thickness paint planes still select a layer, while curved
// surfaces keep the finer lattice's detail.
func (g *Grid) inside(s Solid, p Vec3) bool {
	if s.Contains(p) {
		return true
	}
	if !g.snap {
		return false
	}
	return s.Contains(Vec3{math.Round(p.X*2) / 2, math.Round(p.Y*2) / 2, math.Round(p.Z*2) / 2})
}

// Use sets the brush for subsequent fills.
func (g *Grid) Use(c Color) *Grid { g.brush = c; return g }

// Brush is the color subsequent fills use.
func (g *Grid) Brush() Color { return g.brush }

// With fills using a color, then restores the previous brush.
func (g *Grid) With(c Color, fill func()) *Grid {
	previous := g.brush
	g.brush = c
	fill()
	g.brush = previous
	return g
}

// ColorOf returns a cell's color and whether it is occupied.
func (g *Grid) ColorOf(c Cell) (Color, bool) { color, ok := g.cells[c]; return color, ok }

// Paint recolors occupied cells whose centers lie inside any of the solids.
func (g *Grid) Paint(c Color, solids ...Solid) *Grid {
	for _, s := range solids {
		g.each(s.Bounds(), func(cell Cell, p Vec3) {
			if _, ok := g.cells[cell]; ok && g.inside(s, p) {
				g.cells[cell] = c
			}
		})
	}
	return g
}

func (g *Grid) Len() int           { return len(g.cells) }
func (g *Grid) Has(c Cell) bool    { _, ok := g.cells[c]; return ok }
func (g *Grid) Remove(c Cell)      { delete(g.cells, c) }
func (g *Grid) HasAt(p Vec3) bool  { return g.Has(Nearest(p.Scale(1 / g.span))) }
func (g *Grid) AddAt(p Vec3) *Grid { g.Add(Nearest(p.Scale(1 / g.span))); return g }

// Add occupies a valid cell and reports whether it was newly added.
func (g *Grid) Add(c Cell) bool {
	if !c.Valid() || g.Has(c) {
		return false
	}
	g.cells[c] = g.brush
	return true
}

// Fill occupies every cell whose center lies inside any of the solids.
func (g *Grid) Fill(solids ...Solid) *Grid {
	for _, s := range solids {
		g.each(s.Bounds(), func(c Cell, p Vec3) {
			if g.inside(s, p) {
				g.cells[c] = g.brush
			}
		})
	}
	return g
}

// FillVolume occupies every cell with more than half of its volume inside a solid.
// Slower than Fill, but curved and sloped surfaces come out symmetric and spur-free.
func (g *Grid) FillVolume(solids ...Solid) *Grid {
	for _, s := range solids {
		g.each(s.Bounds().Grow(g.span/2), func(c Cell, p Vec3) {
			inside := 0
			for _, q := range volumeSamples {
				if s.Contains(p.Add(q.Scale(g.span))) {
					inside++
				}
			}
			if inside*2 > len(volumeSamples) {
				g.cells[c] = g.brush
			}
		})
	}
	return g
}

// Carve removes every cell whose center lies inside any of the solids.
func (g *Grid) Carve(solids ...Solid) *Grid {
	for _, s := range solids {
		g.each(s.Bounds(), func(c Cell, p Vec3) {
			if g.inside(s, p) {
				delete(g.cells, c)
			}
		})
	}
	return g
}

// Keep removes every cell outside the solid.
func (g *Grid) Keep(s Solid) *Grid {
	for c := range g.cells {
		if !g.inside(s, g.World(c)) {
			delete(g.cells, c)
		}
	}
	return g
}

// Build places components in order; later components may carve earlier ones.
func (g *Grid) Build(components ...Component) *Grid {
	for _, c := range components {
		c.Build(g)
	}
	return g
}

// Merge adds every cell of another grid, keeping its colors.
func (g *Grid) Merge(o *Grid) *Grid {
	for c, color := range o.cells {
		g.cells[c] = color
	}
	return g
}

// Cells returns the occupied cells in no particular order.
func (g *Grid) Cells() []Cell {
	result := make([]Cell, 0, len(g.cells))
	for c := range g.cells {
		result = append(result, c)
	}
	return result
}

// Bounds is the world-space envelope of the cell centers grown by half a cell.
func (g *Grid) Bounds() Box {
	b := EmptyBox()
	for c := range g.cells {
		p := g.World(c)
		b = b.Union(Box{p, p})
	}
	return b.Grow(g.span / 2)
}

// Order ranks cells, given their world centers, within one horizontal course. Courses
// always assemble bottom to top; within a course, lower keys arrive first.
type Order func(p Vec3) float64

// Sweep orders each course clockwise (seen from above) around center, starting at
// the given plan direction angle. It reads like masons working around a site.
func Sweep(center Vec2, start float64) Order {
	return func(p Vec3) float64 {
		a := p.Plan().Sub(center).Angle() - start
		return math.Mod(math.Mod(a, 2*math.Pi)+2*math.Pi, 2*math.Pi)
	}
}

// Along orders each course by distance along a plan direction.
func Along(direction Vec2) Order {
	return func(p Vec3) float64 { return p.Plan().Dot(direction) }
}

// Outward orders each course from center to the edge.
func Outward(center Vec2) Order {
	return func(p Vec3) float64 { return p.Plan().Sub(center).Length() }
}

// Structure compiles the grid into an ordered, validated structure.
func (g *Grid) Structure(title string, order Order) (*Structure, error) {
	type ranked struct {
		cell  Cell
		color Color
		key   float64
	}
	cells := make([]ranked, 0, len(g.cells))
	for c, color := range g.cells {
		key := 0.0
		if order != nil {
			key = order(g.World(c))
		}
		cells = append(cells, ranked{c, color, key})
	}
	sort.Slice(cells, func(i, j int) bool {
		a, b := cells[i], cells[j]
		if a.cell.Y != b.cell.Y {
			return a.cell.Y < b.cell.Y
		}
		if a.key != b.key {
			return a.key < b.key
		}
		if a.cell.Z != b.cell.Z {
			return a.cell.Z < b.cell.Z
		}
		return a.cell.X < b.cell.X
	})
	s := &Structure{Title: title, Span: float32(g.span), Cells: make([]Cell, len(cells)), Colors: make([]Color, len(cells)), Palette: Palette[:]}
	for i, r := range cells {
		s.Cells[i] = r.cell
		s.Colors[i] = r.color
	}
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", title, err)
	}
	return s, nil
}
