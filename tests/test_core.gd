extends SceneTree
## Run: godot --headless --path . --script tests/test_core.gd

var checks: int = 0
var failures: int = 0
const TEMP_FILE := "/private/tmp/godocta-roundtrip.octa"

func _initialize() -> void:
	_test_geometry()
	_test_lattice_and_builder()
	_test_presets()
	_test_timeline()
	_test_files()
	print("Core checks: %d passed, %d failed." % [checks - failures, failures])
	quit(1 if failures else 0)

func check(condition: bool, message: String) -> void:
	checks += 1
	if not condition:
		failures += 1
		push_error(message)

func _test_geometry() -> void:
	var vertices := TruncatedOctahedronMesh.vertices()
	check(vertices.size() == 24, "Unit must have 24 vertices")
	var edges := 0
	for a in vertices.size():
		for b in range(a + 1, vertices.size()):
			if is_equal_approx(vertices[a].distance_squared_to(vertices[b]), 0.125):
				edges += 1
	check(edges == 36, "Unit must have 36 equal edges")
	var squares := 0
	var hexagons := 0
	for normal in TruncatedOctahedronMesh.face_normals():
		var distance := 0.5 if normal.abs().dot(Vector3.ONE) < 1.01 else 0.75 / sqrt(3.0)
		var count := 0
		for vertex in vertices:
			if is_equal_approx(vertex.dot(normal), distance):
				count += 1
		if count == 4:
			squares += 1
		elif count == 6:
			hexagons += 1
	check(squares == 6 and hexagons == 8, "Unit must have 6 square and 8 hexagonal faces")
	# Every one of the 14 neighbors meets precisely at its corresponding face.
	for offset in OctaLattice.NEIGHBORS:
		var translation := OctaLattice.to_world(offset)
		var shared := 0
		for a in vertices:
			for b in vertices:
				if a.is_equal_approx(b + translation):
					shared += 1
		check(shared == (4 if offset.length_squared() == 4 else 6), "Neighbor %s must share a full face" % offset)
	var mesh := TruncatedOctahedronMesh.shared_mesh()
	check(mesh == TruncatedOctahedronMesh.shared_mesh(), "Mesh must be cached and shared")
	var arrays := mesh.surface_get_arrays(0)
	var positions: PackedVector3Array = arrays[Mesh.ARRAY_VERTEX]
	var normals: PackedVector3Array = arrays[Mesh.ARRAY_NORMAL]
	check(positions.size() == 132, "Mesh must contain exactly 44 triangles")
	var volume := 0.0
	for index in range(0, positions.size(), 3):
		var a := positions[index]
		var b := positions[index + 1]
		var c := positions[index + 2]
		check((b - a).cross(c - a).dot(normals[index]) < 0.0, "Clockwise outward winding")
		volume += a.dot(b.cross(c)) / 6.0
	check(is_equal_approx(absf(volume), 0.5), "BCC cell volume must be 1/2 cubic unit")

func _test_lattice_and_builder() -> void:
	check(OctaLattice.is_valid(Vector3i(-1, -3, 5)), "Negative odd lattice cells are valid")
	check(not OctaLattice.is_valid(Vector3i(0, 1, 0)), "Mixed parity must be rejected")
	check(OctaStructure.from_cells(PackedInt32Array([0, 0])) == null, "Incomplete coordinates rejected")
	check(OctaStructure.from_cells(PackedInt32Array([0, 0, 0, 0, 0, 0])) == null, "Duplicate cells rejected")
	check(OctaStructure.from_cells(PackedInt32Array([0, 1, 0])) == null, "Off-lattice cells rejected")
	var builder := OctaStructureBuilder.new()
	check(builder.add(Vector3i(1, 1, 1)), "Add valid cell")
	check(not builder.add(Vector3i(1, 1, 1)), "Do not add duplicate cell")
	check(builder.add(Vector3i.ZERO), "Add origin")
	check(builder.remove(Vector3i(1, 1, 1)), "Remove cell")
	check(builder.build().cell_count() == 1, "Builder creates compact arrangement")
	var structure := builder.build()
	var copy := structure.packed_cells()
	copy[0] = 100
	check(structure.coordinate(0) == Vector3i.ZERO, "External mutations cannot change packed coordinates")

func _test_presets() -> void:
	for shape in 4:
		var structure := StructurePresets.generate(shape as StructurePresets.Shape)
		check(structure != null and structure.cell_count() > 0, "Preset must produce cells")
		check(structure.storage_bytes() == 12 * structure.cell_count(), "12 bytes per cell")
		check(OctaStructure.validation_error(structure.packed_cells(), structure.cell_span).is_empty(), "Valid unique lattice cells")
		check(structure.packed_cells() == StructurePresets.generate(shape as StructurePresets.Shape).packed_cells(), "Deterministic generation")
		var last_y := -100
		for index in structure.cell_count():
			var cell := structure.coordinate(index)
			check(cell.y >= last_y, "Preset sorted bottom to top")
			last_y = cell.y
			match shape:
				StructurePresets.Shape.SPHERE:
					check(Vector3(cell).length_squared() <= 81.0, "Sphere center membership")
				StructurePresets.Shape.CYLINDER:
					check(cell.x * cell.x + cell.z * cell.z <= 81, "Cylinder center membership")
				StructurePresets.Shape.PYRAMID:
					check(maxi(absi(cell.x), absi(cell.z)) <= (9.0 - cell.y) / 2.0, "Pyramid center membership")
		print("%s: %d cells, %d bytes, bounds %s" % [structure.title, structure.cell_count(), structure.storage_bytes(), structure.bounds().size])
	var cube := StructurePresets.generate(StructurePresets.Shape.CUBE)
	check(cube.cell_count() == 1729, "Ten-unit cube has 10^3 + 9^3 cells")
	check(cube.bounds().size.is_equal_approx(Vector3(10, 10, 10)), "Cube envelope is exactly 10 × 10 × 10")
	check(StructurePresets.generate(StructurePresets.Shape.CUBE, 1) == null, "Reject invalid dimensions")

func _test_timeline() -> void:
	var timeline := AssemblyTimeline.new()
	timeline.restart()
	check(timeline.elapsed == 0.0, "Replay resets clock")
	timeline.advance(3.99)
	check(timeline.is_running(), "Still assembling before four seconds")
	timeline.advance(0.01)
	check(not timeline.is_running() and timeline.elapsed == 4.0, "Finished at exactly four seconds")
	timeline.advance(99)
	check(timeline.elapsed == 4.0, "Clock clamps at completion")
	for rank in [0.0, 0.25, 0.5, 0.75, 1.0]:
		check(is_zero_approx(AssemblyTimeline.cell_progress(0, rank)), "No cells assembled at time zero")
		check(is_equal_approx(AssemblyTimeline.cell_progress(4, rank), 1.0), "All cells settled at four seconds")
		var delay := AssemblyTimeline.release_time(rank)
		var a := AssemblyTimeline.cell_progress(delay + 0.3, rank)
		var b := AssemblyTimeline.cell_progress(delay + 0.6, rank)
		var c := AssemblyTimeline.cell_progress(delay + 0.9, rank)
		check(a > 0.0 and a < b and b < c and c < 1.0, "Gradual monotonic arrival")
		check(a > b - a and b - a > c - b, "Arrival decelerates")
	check(AssemblyTimeline.cell_progress(3.9, 1.0) < 1.0, "Last cell moves through end of animation")

func _test_files() -> void:
	var codec := OctaFile.new()
	var example := codec.load_structure("res://structures/neighbors.octa")
	check(example != null and example.cell_count() == 15, "Cross-language example file loads")
	for shape in 4:
		var source := StructurePresets.generate(shape as StructurePresets.Shape)
		source.title += " · round trip"
		source.cell_span = 1.25
		check(codec.save_structure(TEMP_FILE, source) == OK, "Save compact file")
		var loaded := codec.load_structure(TEMP_FILE)
		check(loaded != null, "Load compact file")
		if loaded != null:
			check(loaded.packed_cells() == source.packed_cells(), "Coordinates survive round trip, including negatives")
			check(loaded.title == source.title and loaded.cell_span == source.cell_span, "Metadata survives round trip")
		check(FileAccess.get_file_as_bytes(TEMP_FILE).size() == 20 + source.title.to_utf8_buffer().size() + source.storage_bytes(), "Exact binary file size")
	var empty := OctaStructure.from_cells(PackedInt32Array())
	check(codec.save_structure(TEMP_FILE, empty) == OK, "Empty arrangement can be saved")
	check(codec.load_structure(TEMP_FILE).cell_count() == 0, "Empty arrangement round trip")
	var small := OctaStructure.from_cells(PackedInt32Array([-1, -1, -1, 1, 1, 1]), "x")
	codec.save_structure(TEMP_FILE, small)
	var valid := FileAccess.get_file_as_bytes(TEMP_FILE)
	for mutation in ["magic", "version", "count", "title", "truncated", "trailing", "parity", "duplicate", "nan"]:
		var bytes := valid.duplicate()
		match mutation:
			"magic": bytes[0] = 0
			"version": bytes.encode_u32(4, 2)
			"count": bytes.encode_u32(12, 0xFFFFFFFF)
			"title": bytes.encode_u32(16, 99999)
			"truncated": bytes.resize(bytes.size() - 1)
			"trailing": bytes.append(0)
			"parity": bytes.encode_s32(21, 0)
			"duplicate":
				for offset in [33, 37, 41]:
					bytes.encode_s32(offset, -1)
			"nan": bytes.encode_float(8, NAN)
		var file := FileAccess.open(TEMP_FILE, FileAccess.WRITE)
		file.store_buffer(bytes)
		file.close()
		check(codec.load_structure(TEMP_FILE) == null and not codec.last_error.is_empty(), "Reject %s file" % mutation)
	DirAccess.remove_absolute(TEMP_FILE)
