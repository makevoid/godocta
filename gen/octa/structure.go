package octa

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// OCTA limits, matching docs/octa-format.md and the studio's reader.
// Version 1 stores coordinates only; version 2 adds a palette and one color byte per cell.
const (
	Version       = 2
	MaxCells      = 1_000_000
	MaxTitleBytes = 1024
	MaxPalette    = 16
	headerV1      = 20
	headerV2      = 24
)

// Structure is an ordered arrangement: the order is the assembly order.
// Colors is empty or parallel to Cells, indexing Palette.
type Structure struct {
	Title   string
	Span    float32
	Cells   []Cell
	Colors  []Color
	Palette []RGB
}

// Validate applies the same rules as the studio's OctaStructure.validation_error.
func (s *Structure) Validate() error {
	span := float64(s.Span)
	if math.IsNaN(span) || math.IsInf(span, 0) || span < 0.001 || span > 1000 {
		return errors.New("cell span must be between 0.001 and 1000")
	}
	if len(s.Cells) > MaxCells {
		return fmt.Errorf("%d cells exceeds the %d cell limit", len(s.Cells), MaxCells)
	}
	if len(s.Title) > MaxTitleBytes {
		return errors.New("title is too long")
	}
	if len(s.Palette) > MaxPalette {
		return fmt.Errorf("palette has %d colors; the limit is %d", len(s.Palette), MaxPalette)
	}
	if len(s.Colors) != 0 && len(s.Colors) != len(s.Cells) {
		return errors.New("colors must be empty or match the cell count")
	}
	if len(s.Colors) != 0 && len(s.Palette) == 0 {
		return errors.New("colored cells need a palette")
	}
	for _, c := range s.Colors {
		if int(c) >= len(s.Palette) {
			return fmt.Errorf("color index %d is outside the %d-color palette", c, len(s.Palette))
		}
	}
	seen := make(map[Cell]struct{}, len(s.Cells))
	for _, c := range s.Cells {
		if !c.Valid() {
			return fmt.Errorf("invalid lattice coordinate %v: coordinates must share parity", c)
		}
		if _, dup := seen[c]; dup {
			return fmt.Errorf("duplicate cell %v", c)
		}
		seen[c] = struct{}{}
	}
	return nil
}

// Bounds is the world-space envelope, including the cells' extent.
func (s *Structure) Bounds() Box {
	b := EmptyBox()
	for _, c := range s.Cells {
		p := c.Center()
		b = b.Union(Box{p, p})
	}
	return Box{b.Min.Scale(float64(s.Span)), b.Max.Scale(float64(s.Span))}.Grow(float64(s.Span) / 2)
}

// MarshalBinary encodes OCTA v2, or v1 when the structure has no colors.
func (s *Structure) MarshalBinary() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	colored := len(s.Colors) > 0
	header, version, palette := headerV1, uint32(1), 0
	if colored {
		header, version, palette = headerV2, 2, len(s.Palette)
	}
	size := header + len(s.Title) + 3*palette + 12*len(s.Cells)
	if colored {
		size += len(s.Cells)
	}
	out := make([]byte, size)
	copy(out, "OCTA")
	binary.LittleEndian.PutUint32(out[4:], version)
	binary.LittleEndian.PutUint32(out[8:], math.Float32bits(s.Span))
	binary.LittleEndian.PutUint32(out[12:], uint32(len(s.Cells)))
	binary.LittleEndian.PutUint32(out[16:], uint32(len(s.Title)))
	if colored {
		binary.LittleEndian.PutUint32(out[20:], uint32(palette))
	}
	offset := header + copy(out[header:], s.Title)
	for _, c := range s.Palette[:palette] {
		out[offset], out[offset+1], out[offset+2] = c.R, c.G, c.B
		offset += 3
	}
	for _, c := range s.Cells {
		binary.LittleEndian.PutUint32(out[offset:], uint32(c.X))
		binary.LittleEndian.PutUint32(out[offset+4:], uint32(c.Y))
		binary.LittleEndian.PutUint32(out[offset+8:], uint32(c.Z))
		offset += 12
	}
	if colored {
		for i, c := range s.Colors {
			out[offset+i] = byte(c)
		}
	}
	return out, nil
}

// UnmarshalBinary decodes and validates OCTA v1 or v2.
func (s *Structure) UnmarshalBinary(data []byte) error {
	if len(data) < headerV1 || string(data[:4]) != "OCTA" {
		return errors.New("not an OCTA structure")
	}
	version := binary.LittleEndian.Uint32(data[4:])
	header, palette := headerV1, 0
	switch {
	case version == 2 && len(data) >= headerV2:
		header, palette = headerV2, int(binary.LittleEndian.Uint32(data[20:]))
	case version != 1:
		return errors.New("unsupported OCTA version")
	}
	span := math.Float32frombits(binary.LittleEndian.Uint32(data[8:]))
	count := int(binary.LittleEndian.Uint32(data[12:]))
	titleBytes := int(binary.LittleEndian.Uint32(data[16:]))
	colorBytes := 0
	if palette > 0 {
		colorBytes = count
	}
	if count > MaxCells || titleBytes > MaxTitleBytes || palette > MaxPalette ||
		len(data) != header+titleBytes+3*palette+12*count+colorBytes {
		return errors.New("invalid or incomplete payload")
	}
	result := Structure{Title: string(data[header : header+titleBytes]), Span: span, Cells: make([]Cell, count)}
	offset := header + titleBytes
	for i := 0; i < palette; i++ {
		result.Palette = append(result.Palette, RGB{data[offset], data[offset+1], data[offset+2]})
		offset += 3
	}
	for i := range result.Cells {
		result.Cells[i] = Cell{
			int32(binary.LittleEndian.Uint32(data[offset:])),
			int32(binary.LittleEndian.Uint32(data[offset+4:])),
			int32(binary.LittleEndian.Uint32(data[offset+8:])),
		}
		offset += 12
	}
	if palette > 0 {
		result.Colors = make([]Color, count)
		for i := range result.Colors {
			result.Colors[i] = Color(data[offset+i])
		}
	}
	if err := result.Validate(); err != nil {
		return err
	}
	*s = result
	return nil
}

// WriteFile writes beside the destination, then renames, so failures keep the old file.
func (s *Structure) WriteFile(path string) error {
	data, err := s.MarshalBinary()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, data, 0o644); err != nil {
		os.Remove(temporary)
		return err
	}
	return os.Rename(temporary, path)
}

// ReadFile loads and validates an OCTA v1 or v2 file.
func ReadFile(path string) (*Structure, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Structure
	if err := s.UnmarshalBinary(data); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &s, nil
}
