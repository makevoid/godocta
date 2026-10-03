package scenes

import (
	"testing"

	. "godocta/gen/octa"
)

func TestScenesCompile(t *testing.T) {
	for _, scene := range All() {
		t.Run(scene.Name(), func(t *testing.T) {
			s, err := Compile(scene)
			if err != nil {
				t.Fatal(err)
			}
			if len(s.Cells) == 0 || len(s.Cells) > MaxCells {
				t.Fatalf("%d cells", len(s.Cells))
			}
			// Layers assemble bottom-up.
			for i := 1; i < len(s.Cells); i++ {
				if s.Cells[i].Y < s.Cells[i-1].Y {
					t.Fatalf("cell %d is below its predecessor", i)
				}
			}
			used := map[Color]int{}
			for _, c := range s.Colors {
				used[c]++
			}
			if len(used) < 3 {
				t.Fatalf("only %d colors used", len(used))
			}
			t.Logf("%d cells, %d colors", len(s.Cells), len(used))
		})
	}
}
