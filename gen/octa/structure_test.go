package octa

import (
	"bytes"
	"testing"
)

func TestRoundTripColored(t *testing.T) {
	g := NewGrid()
	g.Use(Brick).Fill(Block(V(-2, 0, -2), V(2, 2, 2)))
	g.Paint(Gold, Block(V(-2, 2, -2), V(2, 2, 2)))
	s, err := g.Structure("Test", Along(P(1, 0)))
	if err != nil {
		t.Fatal(err)
	}
	data, err := s.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if v := data[4]; v != 2 {
		t.Fatalf("colored structure wrote version %d", v)
	}
	var back Structure
	if err := back.UnmarshalBinary(data); err != nil {
		t.Fatal(err)
	}
	again, _ := back.MarshalBinary()
	if !bytes.Equal(data, again) || len(back.Colors) != len(back.Cells) || len(back.Palette) != int(PaletteSize) {
		t.Fatal("round trip changed the structure")
	}
	seen := map[Color]bool{}
	for _, c := range back.Colors {
		seen[c] = true
	}
	if !seen[Brick] || !seen[Gold] || len(seen) != 2 {
		t.Fatalf("unexpected colors %v", seen)
	}
}

func TestUncoloredWritesV1(t *testing.T) {
	s := &Structure{Title: "Plain", Span: 1, Cells: []Cell{{0, 0, 0}, {1, 1, 1}}}
	data, err := s.MarshalBinary()
	if err != nil || data[4] != 1 || len(data) != headerV1+5+24 {
		t.Fatalf("v1 encoding: %v, %d bytes", err, len(data))
	}
}

func TestValidateRejectsBadColors(t *testing.T) {
	cells := []Cell{{0, 0, 0}, {2, 0, 0}}
	cases := []*Structure{
		{Span: 1, Cells: cells, Colors: []Color{0}, Palette: Palette[:]},
		{Span: 1, Cells: cells, Colors: []Color{0, 3}, Palette: Palette[:2]},
		{Span: 1, Cells: cells, Colors: []Color{0, 0}},
		{Span: 1, Cells: cells, Palette: make([]RGB, 17)},
	}
	for i, s := range cases {
		if s.Validate() == nil {
			t.Errorf("case %d validated", i)
		}
	}
}

func TestTruncatedPayloadFails(t *testing.T) {
	g := NewGrid()
	g.Fill(Sphere(V(0, 0, 0), 2))
	ball, _ := g.Structure("Ball", Along(P(1, 0)))
	data, _ := ball.MarshalBinary()
	var s Structure
	if s.UnmarshalBinary(data[:len(data)-1]) == nil {
		t.Fatal("truncated file decoded")
	}
}
