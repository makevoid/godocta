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
	// Build composes the cells and chooses the assembly order.
	Build() (*octa.Grid, octa.Order)
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

// Compile builds a scene into an ordered structure.
func Compile(s Scene) (*octa.Structure, error) {
	grid, order := s.Build()
	return grid.Structure(s.Title(), order)
}
