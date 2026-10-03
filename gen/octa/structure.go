package octa

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// OCTA v1 limits, matching docs/octa-format.md and the studio's reader.
const (
	Version       = 1
	MaxCells      = 200_000
	MaxTitleBytes = 1024
	headerBytes   = 20
)

// Structure is an ordered arrangement: the order is the assembly order.
type Structure struct {
	Title string
	Span  float32
	Cells []Cell
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

// MarshalBinary encodes OCTA v1.
func (s *Structure) MarshalBinary() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	out := make([]byte, headerBytes+len(s.Title)+12*len(s.Cells))
	copy(out, "OCTA")
	binary.LittleEndian.PutUint32(out[4:], Version)
	binary.LittleEndian.PutUint32(out[8:], math.Float32bits(s.Span))
	binary.LittleEndian.PutUint32(out[12:], uint32(len(s.Cells)))
	binary.LittleEndian.PutUint32(out[16:], uint32(len(s.Title)))
	copy(out[headerBytes:], s.Title)
	offset := headerBytes + len(s.Title)
	for _, c := range s.Cells {
		binary.LittleEndian.PutUint32(out[offset:], uint32(c.X))
		binary.LittleEndian.PutUint32(out[offset+4:], uint32(c.Y))
		binary.LittleEndian.PutUint32(out[offset+8:], uint32(c.Z))
		offset += 12
	}
	return out, nil
}

// UnmarshalBinary decodes and validates OCTA v1.
func (s *Structure) UnmarshalBinary(data []byte) error {
	if len(data) < headerBytes || string(data[:4]) != "OCTA" {
		return errors.New("not an OCTA structure")
	}
	if binary.LittleEndian.Uint32(data[4:]) != Version {
		return errors.New("unsupported OCTA version")
	}
	span := math.Float32frombits(binary.LittleEndian.Uint32(data[8:]))
	count := int(binary.LittleEndian.Uint32(data[12:]))
	titleBytes := int(binary.LittleEndian.Uint32(data[16:]))
	if count > MaxCells || titleBytes > MaxTitleBytes || len(data) != headerBytes+titleBytes+12*count {
		return errors.New("invalid or incomplete coordinate payload")
	}
	result := Structure{Title: string(data[headerBytes : headerBytes+titleBytes]), Span: span, Cells: make([]Cell, count)}
	offset := headerBytes + titleBytes
	for i := range result.Cells {
		result.Cells[i] = Cell{
			int32(binary.LittleEndian.Uint32(data[offset:])),
			int32(binary.LittleEndian.Uint32(data[offset+4:])),
			int32(binary.LittleEndian.Uint32(data[offset+8:])),
		}
		offset += 12
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

// ReadFile loads and validates an OCTA v1 file.
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
