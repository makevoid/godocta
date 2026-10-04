class_name OctaEditor
extends RefCounted
## Pending edits over one structure, for the brush tools. Removals hide original cells by
## index; additions are kept in the order they were made. build() compacts both into a new
## structure: surviving cells keep their assembly order and additions follow.
## Points and radii are in structure space, where a cell center is OctaLattice.to_world().

var structure: OctaStructure
## The Y coordinate of cells that rest on the ground; nothing is added below it.
var ground_layer: int
var _span: float
var _index: Dictionary = {} # Vector3i → original index
var _removed: Dictionary = {} # original index → true
var _added: Dictionary = {} # Vector3i → palette index (-1 uncolored), in insertion order
var _bounds: AABB

## Keep ground_layer across commits so the ground plane does not move under an edit.
func _init(data: OctaStructure, layer: Variant = null) -> void:
	structure = data
	_span = data.cell_span
	var cells := data.packed_cells()
	var lowest := 1 << 30
	for index in cells.size() / 3:
		var cell := Vector3i(cells[index * 3], cells[index * 3 + 1], cells[index * 3 + 2])
		_index[cell] = index
		lowest = mini(lowest, cell.y)
	# An empty structure gets an odd layer: its cells then stand on the renderer's floor.
	ground_layer = layer if layer != null else (lowest if not _index.is_empty() else 1)
	_bounds = data.bounds() if not _index.is_empty() else AABB(Vector3(0, ground_height(), 0), Vector3.ZERO)

func span() -> float:
	return _span

## Height of the ground plane: the bottom faces of the ground layer.
func ground_height() -> float:
	return (ground_layer - 1) * _span * 0.5

func contains(cell: Vector3i) -> bool:
	if _added.has(cell):
		return true
	var index: int = _index.get(cell, -1)
	return index >= 0 and not _removed.has(index)

## Index in the displayed structure, or -1 for added or absent cells.
func original_index(cell: Vector3i) -> int:
	var index: int = _index.get(cell, -1)
	return index if index >= 0 and not _removed.has(index) and not _added.has(cell) else -1

func color_of(cell: Vector3i) -> int:
	if _added.has(cell):
		return _added[cell]
	var index: int = _index.get(cell, -1)
	return structure.color_index(index) if index >= 0 else -1

func cell_count() -> int:
	return _index.size() - _removed.size() + _added.size()

func has_changes() -> bool:
	return not _removed.is_empty() or not _added.is_empty()

func add(cell: Vector3i, color: int = -1) -> bool:
	if not OctaLattice.is_valid(cell) or contains(cell) or cell_count() >= OctaStructure.MAX_CELLS:
		return false
	_added[cell] = color
	var center := OctaLattice.to_world(cell, _span)
	_bounds = _bounds.merge(AABB(center, Vector3.ZERO).grow(_span * 0.5))
	return true

func remove(cell: Vector3i) -> bool:
	if _added.erase(cell):
		return true
	var index: int = _index.get(cell, -1)
	if index < 0 or _removed.has(index):
		return false
	_removed[index] = true
	return true

## Palette index for a new cell: its neighbors' most common color, else fallback.
## Uncolored structures always return -1.
func color_for(cell: Vector3i, fallback: int = 0) -> int:
	if not structure.is_colored():
		return -1
	var votes: Dictionary = {}
	var best := -1
	for offset in OctaLattice.NEIGHBORS:
		if contains(cell + offset):
			var color := color_of(cell + offset)
			votes[color] = votes.get(color, 0) + 1
			if best < 0 or votes[color] > votes[best]:
				best = color
	return best if best >= 0 else clampi(fallback, 0, maxi(structure.palette.size() - 1, 0))

## The palette index a new cell takes: the chosen color when it is in the palette, else
## its neighbors' vote. Uncolored structures always return -1, whatever is chosen.
func color_choice(cell: Vector3i, chosen: int, fallback: int = 0) -> int:
	if not structure.is_colored():
		return -1
	if chosen >= 0:
		return mini(chosen, structure.palette.size() - 1)
	return color_for(cell, fallback)

## Occupied cells whose centers lie within radius. A nonzero axis keeps only the half on
## its side of the plane through center: the scoop's half sphere.
func cells_within(center: Vector3, radius: float, axis: Vector3 = Vector3.ZERO) -> Array[Vector3i]:
	var found: Array[Vector3i] = []
	for cell in _lattice_ball(center, radius):
		if contains(cell) and (axis == Vector3.ZERO or (OctaLattice.to_world(cell, _span) - center).dot(axis) >= 0.0):
			found.append(cell)
	return found

## Empty slots within radius that touch an occupied cell or rest on the ground layer,
## nearest first. New cells are attracted to the existing ones.
func free_slots(center: Vector3, radius: float, limit: int) -> Array[Vector3i]:
	var candidates: Dictionary = {}
	var limit_squared := radius * radius
	for cell in _lattice_ball(center, radius + _span):
		if not contains(cell):
			continue
		for offset in OctaLattice.NEIGHBORS:
			var slot: Vector3i = cell + offset
			if slot.y >= ground_layer and not contains(slot) and OctaLattice.is_valid(slot):
				var distance := OctaLattice.to_world(slot, _span).distance_squared_to(center)
				if distance <= limit_squared:
					candidates[slot] = distance
	# Ground slots under the brush need no neighbor: they rest on the floor.
	var floor_y := OctaLattice.to_world(Vector3i(0, ground_layer, 0), _span).y
	if absf(floor_y - center.y) <= radius:
		for slot in _lattice_ball(Vector3(center.x, floor_y, center.z), radius):
			if slot.y == ground_layer and not contains(slot):
				var distance := OctaLattice.to_world(slot, _span).distance_squared_to(center)
				if distance <= limit_squared:
					candidates[slot] = distance
	var slots: Array = candidates.keys()
	slots.sort_custom(func(a: Vector3i, b: Vector3i) -> bool: return candidates[a] < candidates[b])
	var result: Array[Vector3i] = []
	for slot: Vector3i in slots.slice(0, limit):
		result.append(slot)
	return result

## First occupied cell along a ray: {cell, point, hit = true}. Otherwise the ray's
## crossing of the ground plane {point, hit = false}, or {} when it misses both.
func pick(origin: Vector3, direction: Vector3) -> Dictionary:
	var along := _ray_range(origin, direction, _bounds.grow(_span))
	if along.x <= along.y:
		var step := _span * 0.18
		var last := Vector3i(1 << 30, 0, 0)
		var t := along.x
		while t <= along.y:
			var point := origin + direction * t
			var cell := OctaLattice.nearest(point, _span)
			if cell != last:
				last = cell
				if contains(cell):
					return {"cell": cell, "point": point, "hit": true}
			t += step
	if direction.y < -1e-6:
		var distance := (ground_height() - origin.y) / direction.y
		if distance > 0.0:
			return {"point": origin + direction * distance, "hit": false}
	return {}

## The edited structure: surviving cells in their original order, then additions.
func build() -> OctaStructure:
	var source := structure.packed_cells()
	var colored := structure.is_colored()
	var cells := PackedInt32Array()
	var colors := PackedByteArray()
	cells.resize((_index.size() - _removed.size() + _added.size()) * 3)
	if colored:
		colors.resize(cells.size() / 3)
	var next := 0
	for index in _index.size():
		if _removed.has(index):
			continue
		cells[next * 3] = source[index * 3]
		cells[next * 3 + 1] = source[index * 3 + 1]
		cells[next * 3 + 2] = source[index * 3 + 2]
		if colored:
			colors[next] = structure.color_index(index)
		next += 1
	for cell: Vector3i in _added:
		cells[next * 3] = cell.x
		cells[next * 3 + 1] = cell.y
		cells[next * 3 + 2] = cell.z
		if colored:
			colors[next] = _added[cell]
		next += 1
	var result := OctaStructure.from_cells(cells, structure.title, _span)
	if colored and result != null:
		result.set_colors(colors, structure.palette)
	return result

## Every lattice coordinate whose center lies within radius of center.
func _lattice_ball(center: Vector3, radius: float) -> Array[Vector3i]:
	var cells: Array[Vector3i] = []
	var q := center * (2.0 / _span)
	var reach := radius * 2.0 / _span
	var reach_squared := reach * reach
	for x in range(floori(q.x - reach), ceili(q.x + reach) + 1):
		for y in range(floori(q.y - reach), ceili(q.y + reach) + 1):
			if ((x - y) & 1) != 0:
				continue
			var z := floori(q.z - reach)
			if ((z - x) & 1) != 0:
				z += 1
			while z <= q.z + reach:
				if Vector3(x, y, z).distance_squared_to(q) <= reach_squared:
					cells.append(Vector3i(x, y, z))
				z += 2
	return cells

## Entry and exit distances of a ray through a box; x > y means a miss.
static func _ray_range(origin: Vector3, direction: Vector3, box: AABB) -> Vector2:
	var near := 0.0
	var far := INF
	for axis in 3:
		if absf(direction[axis]) < 1e-9:
			if origin[axis] < box.position[axis] or origin[axis] > box.end[axis]:
				return Vector2(1, 0)
			continue
		var a := (box.position[axis] - origin[axis]) / direction[axis]
		var b := (box.end[axis] - origin[axis]) / direction[axis]
		near = maxf(near, minf(a, b))
		far = minf(far, maxf(a, b))
	return Vector2(near, far)
