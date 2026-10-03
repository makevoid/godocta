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

## The cell containing a world point. Truncated octahedrons are the Voronoi cells of the
## BCC lattice, so this is the nearest all-even or all-odd coordinate triple.
static func nearest(point: Vector3, cell_span: float = 1.0) -> Vector3i:
	var q := point * (2.0 / cell_span)
	var even := Vector3i(roundi(q.x * 0.5) * 2, roundi(q.y * 0.5) * 2, roundi(q.z * 0.5) * 2)
	var odd := Vector3i(floori(q.x * 0.5) * 2 + 1, floori(q.y * 0.5) * 2 + 1, floori(q.z * 0.5) * 2 + 1)
	return even if q.distance_squared_to(Vector3(even)) <= q.distance_squared_to(Vector3(odd)) else odd
