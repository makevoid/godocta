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

// Grid is a mutable set of occupied lattice cells.
type Grid struct {
	cells map[Cell]struct{}
}

func NewGrid() *Grid { return &Grid{cells: make(map[Cell]struct{})} }

func (g *Grid) Len() int           { return len(g.cells) }
func (g *Grid) Has(c Cell) bool    { _, ok := g.cells[c]; return ok }
func (g *Grid) Remove(c Cell)      { delete(g.cells, c) }
func (g *Grid) HasAt(p Vec3) bool  { return g.Has(Nearest(p)) }
func (g *Grid) AddAt(p Vec3) *Grid { g.Add(Nearest(p)); return g }

// Add occupies a valid cell and reports whether it was newly added.
func (g *Grid) Add(c Cell) bool {
	if !c.Valid() || g.Has(c) {
		return false
	}
	g.cells[c] = struct{}{}
	return true
}

// Fill occupies every cell whose center lies inside any of the solids.
func (g *Grid) Fill(solids ...Solid) *Grid {
	for _, s := range solids {
		eachCell(s.Bounds(), func(c Cell, p Vec3) {
			if s.Contains(p) {
				g.cells[c] = struct{}{}
			}
		})
	}
	return g
}

// FillVolume occupies every cell with more than half of its volume inside a solid.
// Slower than Fill, but curved and sloped surfaces come out symmetric and spur-free.
func (g *Grid) FillVolume(solids ...Solid) *Grid {
	for _, s := range solids {
		eachCell(s.Bounds().Grow(0.5), func(c Cell, p Vec3) {
			inside := 0
			for _, q := range volumeSamples {
				if s.Contains(p.Add(q)) {
					inside++
				}
			}
			if inside*2 > len(volumeSamples) {
				g.cells[c] = struct{}{}
			}
		})
	}
	return g
}

// Carve removes every cell whose center lies inside any of the solids.
func (g *Grid) Carve(solids ...Solid) *Grid {
	for _, s := range solids {
		eachCell(s.Bounds(), func(c Cell, p Vec3) {
			if s.Contains(p) {
				delete(g.cells, c)
			}
		})
	}
	return g
}

// Keep removes every cell outside the solid.
func (g *Grid) Keep(s Solid) *Grid {
	for c := range g.cells {
		if !s.Contains(c.Center()) {
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

// Merge adds every cell of another grid.
func (g *Grid) Merge(o *Grid) *Grid {
	for c := range o.cells {
		g.cells[c] = struct{}{}
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
		p := c.Center()
		b = b.Union(Box{p, p})
	}
	return b.Grow(0.5)
}

// Order ranks cells within one horizontal course. Courses always assemble bottom to
// top; within a course, lower keys arrive first.
type Order func(c Cell) float64

// Sweep orders each course clockwise (seen from above) around center, starting at
// the given plan direction angle. It reads like masons working around a site.
func Sweep(center Vec2, start float64) Order {
	return func(c Cell) float64 {
		a := c.Center().Plan().Sub(center).Angle() - start
		return math.Mod(math.Mod(a, 2*math.Pi)+2*math.Pi, 2*math.Pi)
	}
}

// Along orders each course by distance along a plan direction.
func Along(direction Vec2) Order {
	return func(c Cell) float64 { return c.Center().Plan().Dot(direction) }
}

// Outward orders each course from center to the edge.
func Outward(center Vec2) Order {
	return func(c Cell) float64 { return c.Center().Plan().Sub(center).Length() }
}

// Structure compiles the grid into an ordered, validated structure.
func (g *Grid) Structure(title string, order Order) (*Structure, error) {
	type ranked struct {
		cell Cell
		key  float64
	}
	cells := make([]ranked, 0, len(g.cells))
	for c := range g.cells {
		key := 0.0
		if order != nil {
			key = order(c)
		}
		cells = append(cells, ranked{c, key})
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
	s := &Structure{Title: title, Span: 1, Cells: make([]Cell, len(cells))}
	for i, r := range cells {
		s.Cells[i] = r.cell
	}
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", title, err)
	}
	return s, nil
}
