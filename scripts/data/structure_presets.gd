class_name StructurePresets
extends RefCounted
## Select whole cells by their centers; boundary silhouettes are intentionally cellular.

enum Shape { CUBE, SPHERE, PYRAMID, CYLINDER }
const TITLES: PackedStringArray = ["Cube", "Sphere", "Pyramid", "Cylinder"]
const DESCRIPTIONS: PackedStringArray = [
	"10 × 10 × 10 units", "9 × 9 × 9 unit envelope", "10-unit base · 9.5-unit height", "9-unit diameter · 10-unit height"
]

static func generate(shape: Shape, dimension: int = 10) -> OctaStructure:
	if dimension < 2 or dimension > 40 or shape < Shape.CUBE or shape > Shape.CYLINDER:
		return null
	var cells := PackedInt32Array()
	var extent := dimension - 1
	var radius := float(extent)
	# Bottom-to-top, then row-by-row: deterministic storage and assembly order.
	for y in range(-extent, extent + 1):
		for z in range(-extent, extent + 1):
			if (z & 1) != (y & 1):
				continue
			for x in range(-extent, extent + 1):
				if (x & 1) != (y & 1):
					continue
				var include := true
				match shape:
					Shape.SPHERE:
						include = Vector3(x, y, z).length_squared() <= radius * radius
					Shape.CYLINDER:
						include = float(x * x + z * z) <= radius * radius
					Shape.PYRAMID:
						var width := (radius - float(y)) * 0.5
						include = absf(float(x)) <= width and absf(float(z)) <= width
				if include:
					cells.append(x)
					cells.append(y)
					cells.append(z)
	return OctaStructure.from_cells(cells, TITLES[shape])
