class_name OctaStructureBuilder
extends RefCounted
## Mutable authoring helper. Its temporary lookup is discarded after build().
## Use OctaStructure for storage and rendering.

var _occupied: Dictionary = {}

func add(cell: Vector3i) -> bool:
	if not OctaLattice.is_valid(cell) or _occupied.has(cell) or _occupied.size() >= OctaStructure.MAX_CELLS:
		return false
	_occupied[cell] = true
	return true

func remove(cell: Vector3i) -> bool:
	return _occupied.erase(cell)

func contains(cell: Vector3i) -> bool:
	return _occupied.has(cell)

func build(title: String = "Custom structure", span: float = 1.0) -> OctaStructure:
	var cells: Array = _occupied.keys()
	cells.sort_custom(func(a: Vector3i, b: Vector3i) -> bool:
		if a.y != b.y:
			return a.y < b.y
		if a.z != b.z:
			return a.z < b.z
		return a.x < b.x)
	var packed := PackedInt32Array()
	packed.resize(cells.size() * 3)
	for index in cells.size():
		var cell: Vector3i = cells[index]
		packed[index * 3] = cell.x
		packed[index * 3 + 1] = cell.y
		packed[index * 3 + 2] = cell.z
	return OctaStructure.from_cells(packed, title, span)
