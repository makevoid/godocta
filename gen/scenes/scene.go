// Package scenes composes kit components into complete, named structures.
package scenes

import (
	"sort"

	"godocta/gen/octa"
)

// Scene is a complete structure recipe.
type Scene interface {
	// Name is the file stem and CLI argument, e.g. "castle".
	Name() string
	// Title is stored in the .octa file and shown by the studio.
	Title() string
	// Build composes the cells in world units and chooses the assembly order. The grid's
	// span sets the resolution; scenes don't depend on it.
	Build(g *octa.Grid) octa.Order
}

var registry = map[string]Scene{}

// Register adds a scene; scene files call it from init.
func Register(s Scene) { registry[s.Name()] = s }

// Get looks up a scene by name.
func Get(name string) (Scene, bool) { s, ok := registry[name]; return s, ok }

// All returns every registered scene sorted by name.
func All() []Scene {
	result := make([]Scene, 0, len(registry))
	for _, s := range registry {
		result = append(result, s)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name() < result[j].Name() })
	return result
}

// DefaultDensity packs twice as many cells as span 1 for finer surfaces.
const DefaultDensity = 2

// Compile builds a scene into an ordered structure at a cell density relative to span 1.
func Compile(s Scene, density float64) (*octa.Structure, error) {
	grid := octa.NewGridWithSpan(octa.SpanForDensity(density))
	order := s.Build(grid)
	return grid.Structure(s.Title(), order)
}
