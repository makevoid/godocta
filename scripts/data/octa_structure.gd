class_name OctaStructure
extends RefCounted
## Compact, render-independent structure. No per-cell Nodes, meshes or transforms.
## The authoritative arrangement occupies 12 bytes per cell, excluding metadata.

const MAX_CELLS: int = 1_000_000
const MAX_PALETTE: int = 16
var title: String = "Untitled"
var cell_span: float = 1.0
## Optional per-cell colors: indices into palette, parallel to the cells. Empty when uncolored.
var palette: PackedColorArray = PackedColorArray()
var _cells: PackedInt32Array = PackedInt32Array()
var _colors: PackedByteArray = PackedByteArray()

static func from_cells(cells: PackedInt32Array, structure_title: String = "Untitled", span: float = 1.0) -> OctaStructure:
	if not validation_error(cells, span).is_empty():
		return null
	var result := OctaStructure.new()
	result._cells = cells.duplicate()
	result.title = structure_title
	result.cell_span = span
	return result

static func validation_error(cells: PackedInt32Array, span: float) -> String:
	if not is_finite(span) or span < 0.001 or span > 1000.0:
		return "Cell span must be between 0.001 and 1000."
	if cells.size() % 3 != 0 or cells.size() / 3 > MAX_CELLS:
		return "Expected at most %d complete integer coordinate triples." % MAX_CELLS
	var seen: Dictionary = {}
	for offset in range(0, cells.size(), 3):
		var cell := Vector3i(cells[offset], cells[offset + 1], cells[offset + 2])
		if not OctaLattice.is_valid(cell):
			return "Invalid lattice coordinate: %s. Coordinates must share parity." % cell
		if seen.has(cell):
			return "Duplicate cell: %s." % cell
		seen[cell] = true
	return ""

static func color_error(cell_total: int, colors: PackedByteArray, colors_palette: PackedColorArray) -> String:
	if colors_palette.size() > MAX_PALETTE:
		return "A palette holds at most %d colors." % MAX_PALETTE
	if colors.is_empty():
		return ""
	if colors.size() != cell_total:
		return "Expected one color per cell."
	for index in colors:
		if index >= colors_palette.size():
			return "Color index %d is outside the palette." % index
	return ""

## Attaches colors; returns an error message, or "" on success.
func set_colors(colors: PackedByteArray, colors_palette: PackedColorArray) -> String:
	var invalid := color_error(cell_count(), colors, colors_palette)
	if invalid.is_empty():
		_colors = colors.duplicate()
		palette = colors_palette.duplicate()
	return invalid

func is_colored() -> bool:
	return not _colors.is_empty()

## Palette index of a cell, or -1 when the structure is uncolored.
func color_index(index: int) -> int:
	return _colors[index] if is_colored() else -1

func packed_colors() -> PackedByteArray:
	return _colors.duplicate()

func cell_count() -> int:
	return _cells.size() / 3

func coordinate(index: int) -> Vector3i:
	var offset := index * 3
	return Vector3i(_cells[offset], _cells[offset + 1], _cells[offset + 2])

func packed_cells() -> PackedInt32Array:
	return _cells.duplicate()

func storage_bytes() -> int:
	return _cells.size() * 4 + _colors.size()

func bounds() -> AABB:
	if cell_count() == 0:
		return AABB()
	var first := OctaLattice.to_world(coordinate(0), cell_span)
	var box := AABB(first, Vector3.ZERO)
	for index in range(1, cell_count()):
		box = box.expand(OctaLattice.to_world(coordinate(index), cell_span))
	return box.grow(cell_span * 0.5)
