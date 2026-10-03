class_name StructurePresets
extends RefCounted
## Whole-cell approximations of ideal solids. A cell belongs to a curved solid when more than
## half of its volume lies inside it, which keeps silhouettes symmetric and free of single-cell
## spurs or pits. The pyramid uses the lattice's natural 45° step, so every course sits fully
## on the one below it. Larger compositions (castle, arena, …) come from the Go generator in
## gen/ as .octa files under structures/.

enum Shape { CUBE, SPHERE, PYRAMID, CYLINDER }
const TITLES: PackedStringArray = ["Cube", "Sphere", "Pyramid", "Cylinder"]
const DESCRIPTIONS: PackedStringArray = [
	"10 × 10 × 10 units", "10-unit diameter", "10-unit base · 5.5-unit height",
	"10-unit diameter · 10-unit height"
]
## Truncated octahedron circumradius for cell_span = 1: |(0, 1, 2) / 4|.
const CELL_RADIUS: float = 0.559

static var _volume_samples: PackedVector3Array

static func generate(shape: Shape, dimension: int = 10) -> OctaStructure:
	if dimension < 2 or dimension > 40 or shape < Shape.CUBE or shape > Shape.CYLINDER:
		return null
	var cells := PackedInt32Array()
	var extent := dimension - 1
	var radius := dimension * 0.5
	# Bottom-to-top, then row-by-row: deterministic storage and assembly order.
	for y in range(-extent, extent + 1):
		if shape == Shape.PYRAMID and y > 0:
			break
		for z in range(-extent, extent + 1):
			if (z & 1) != (y & 1):
				continue
			for x in range(-extent, extent + 1):
				if (x & 1) != (y & 1):
					continue
				var include := true
				var center := Vector3(x, y, z) * 0.5
				match shape:
					Shape.SPHERE:
						include = mostly_inside(center, center.length() - radius,
							func(p: Vector3) -> bool: return p.length_squared() <= radius * radius)
					Shape.CYLINDER:
						include = mostly_inside(center, Vector2(center.x, center.z).length() - radius,
							func(p: Vector3) -> bool: return p.x * p.x + p.z * p.z <= radius * radius)
					Shape.PYRAMID:
						# Each course steps in by half a cell: the profile shrinks monotonically.
						include = maxi(absi(x), absi(z)) <= -y
				if include:
					cells.append(x)
					cells.append(y)
					cells.append(z)
	return OctaStructure.from_cells(cells, TITLES[shape])

## True when more than half of the cell centred at `center` (cell_span = 1) satisfies `inside`.
## `distance` is a signed distance estimate to the surface; only boundary cells are sampled.
static func mostly_inside(center: Vector3, distance: float, inside: Callable) -> bool:
	if distance <= -CELL_RADIUS:
		return true
	if distance >= CELL_RADIUS:
		return false
	var samples := volume_samples()
	var count := 0
	for sample in samples:
		if inside.call(center + sample):
			count += 1
	return count * 2 > samples.size()

## Evenly spaced points filling one truncated octahedron: |x|,|y|,|z| ≤ 1/2 and |x|+|y|+|z| ≤ 3/4.
static func volume_samples() -> PackedVector3Array:
	if not _volume_samples.is_empty():
		return _volume_samples
	const STEPS := 8
	for i in STEPS:
		for j in STEPS:
			for k in STEPS:
				var p := (Vector3(i, j, k) + Vector3.ONE * 0.5) / STEPS - Vector3.ONE * 0.5
				if absf(p.x) + absf(p.y) + absf(p.z) <= 0.75:
					_volume_samples.append(p)
	return _volume_samples
