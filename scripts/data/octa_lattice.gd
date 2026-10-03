class_name OctaLattice
extends RefCounted
## Integer coordinates on a body-centered cubic lattice.
## All three coordinates must have the same parity. World = coordinate * span / 2.

const MAX_COORDINATE: int = 1_000_000
const NEIGHBORS: Array[Vector3i] = [
	Vector3i(2, 0, 0), Vector3i(-2, 0, 0),
	Vector3i(0, 2, 0), Vector3i(0, -2, 0),
	Vector3i(0, 0, 2), Vector3i(0, 0, -2),
	Vector3i(1, 1, 1), Vector3i(1, 1, -1),
	Vector3i(1, -1, 1), Vector3i(1, -1, -1),
	Vector3i(-1, 1, 1), Vector3i(-1, 1, -1),
	Vector3i(-1, -1, 1), Vector3i(-1, -1, -1),
]

static func is_valid(cell: Vector3i) -> bool:
	return (cell.x & 1) == (cell.y & 1) and (cell.y & 1) == (cell.z & 1) \
		and absi(cell.x) <= MAX_COORDINATE and absi(cell.y) <= MAX_COORDINATE \
		and absi(cell.z) <= MAX_COORDINATE

static func to_world(cell: Vector3i, cell_span: float = 1.0) -> Vector3:
	return Vector3(cell) * (cell_span * 0.5)
