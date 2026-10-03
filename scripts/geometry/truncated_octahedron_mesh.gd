class_name TruncatedOctahedronMesh
extends RefCounted
## One regular truncated octahedron: 24 vertices, 36 edges, 6 squares, 8 hexagons.
## Vertex coordinates are permutations of (0, ±1, ±2), scaled by 1/4.
## Axis-aligned span = 1. Mesh triangles have flat normals and polygon-edge UVs.

static var _shared_mesh: ArrayMesh

static func vertices() -> PackedVector3Array:
	var result := PackedVector3Array()
	for zero_axis in 3:
		for swap in 2:
			for sign_a in [-1.0, 1.0]:
				for sign_b in [-1.0, 1.0]:
					var vertex := Vector3.ZERO
					vertex[(zero_axis + 1) % 3] = (1.0 + swap) * sign_a * 0.25
					vertex[(zero_axis + 2) % 3] = (2.0 - swap) * sign_b * 0.25
					result.append(vertex)
	return result

static func face_normals() -> PackedVector3Array:
	var result := PackedVector3Array([Vector3.RIGHT, Vector3.LEFT, Vector3.UP,
		Vector3.DOWN, Vector3.FORWARD, Vector3.BACK])
	for x in [-1.0, 1.0]:
		for y in [-1.0, 1.0]:
			for z in [-1.0, 1.0]:
				result.append(Vector3(x, y, z).normalized())
	return result

static func shared_mesh() -> ArrayMesh:
	if _shared_mesh != null:
		return _shared_mesh
	# Indexed: each face's corners are stored once (72 vertices) and shared by its
	# triangles, so the vertex shader runs ~45% less than with 132 unindexed corners.
	var positions := PackedVector3Array()
	var indices := PackedInt32Array()
	var normals := PackedVector3Array()
	var uvs := PackedVector2Array()
	var polygon_data := PackedVector2Array()
	var all_vertices := vertices()
	for normal in face_normals():
		var distance := 0.5 if normal.abs().dot(Vector3.ONE) < 1.01 else 0.75 / sqrt(3.0)
		var center := normal * distance
		var polygon: Array[Vector3] = []
		for vertex in all_vertices:
			if is_equal_approx(normal.dot(vertex), distance):
				polygon.append(vertex)
		var u := (polygon[0] - center).normalized()
		var v := normal.cross(u)
		polygon.sort_custom(func(a: Vector3, b: Vector3) -> bool:
			return atan2((a - center).dot(v), (a - center).dot(u)) < atan2((b - center).dot(v), (b - center).dot(u)))
		var apothem := (polygon[0] - center).length() * cos(PI / polygon.size())
		var first := positions.size()
		for vertex in polygon:
			positions.append(vertex)
			normals.append(normal)
			uvs.append(Vector2((vertex - center).dot(u), (vertex - center).dot(v)))
			polygon_data.append(Vector2(polygon.size(), apothem))
		for index in range(1, polygon.size() - 1):
			# Godot front faces use clockwise winding.
			indices.append_array([first, first + index + 1, first + index])
	var arrays: Array = []
	arrays.resize(Mesh.ARRAY_MAX)
	arrays[Mesh.ARRAY_VERTEX] = positions
	arrays[Mesh.ARRAY_NORMAL] = normals
	arrays[Mesh.ARRAY_TEX_UV] = uvs
	arrays[Mesh.ARRAY_TEX_UV2] = polygon_data
	arrays[Mesh.ARRAY_INDEX] = indices
	_shared_mesh = ArrayMesh.new()
	_shared_mesh.add_surface_from_arrays(Mesh.PRIMITIVE_TRIANGLES, arrays)
	return _shared_mesh
