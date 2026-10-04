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
	_test_colors()
	_test_editor()
	_test_generated_builds()
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
	var indices: PackedInt32Array = arrays[Mesh.ARRAY_INDEX]
	check(positions.size() == 72, "Each face stores its corners once: 6 × 4 + 8 × 6")
	check(indices.size() == 132, "Mesh must contain exactly 44 triangles")
	var volume := 0.0
	for index in range(0, indices.size(), 3):
		var a := positions[indices[index]]
		var b := positions[indices[index + 1]]
		var c := positions[indices[index + 2]]
		check((b - a).cross(c - a).dot(normals[indices[index]]) < 0.0, "Clockwise outward winding")
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
		var members: Dictionary = {}
		for index in structure.cell_count():
			members[structure.coordinate(index)] = true
		for index in structure.cell_count():
			var cell := structure.coordinate(index)
			check(cell.y >= last_y, "Preset sorted bottom to top")
			last_y = cell.y
			# Cells are kept when most of their volume is inside: centers stay within one
			# circumradius of the surface, and the shapes stay mirror-symmetric.
			var center := Vector3(cell) * 0.5
			check(members.has(Vector3i(-cell.x, cell.y, cell.z)) and members.has(Vector3i(cell.z, cell.y, cell.x)), "Preset is symmetric")
			match shape:
				StructurePresets.Shape.SPHERE:
					check(center.length() <= 5.0 + StructurePresets.CELL_RADIUS, "Sphere membership")
				StructurePresets.Shape.CYLINDER:
					check(Vector2(center.x, center.z).length() <= 5.0 + StructurePresets.CELL_RADIUS, "Cylinder membership")
				StructurePresets.Shape.PYRAMID:
					check(cell.y <= 0 and maxi(absi(cell.x), absi(cell.z)) <= -cell.y, "Pyramid steps in every course")
		print("%s: %d cells, %d bytes, bounds %s" % [structure.title, structure.cell_count(), structure.storage_bytes(), structure.bounds().size])
	var cube := StructurePresets.generate(StructurePresets.Shape.CUBE)
	check(cube.cell_count() == 1729, "Ten-unit cube has 10^3 + 9^3 cells")
	check(cube.bounds().size.is_equal_approx(Vector3(10, 10, 10)), "Cube envelope is exactly 10 × 10 × 10")
	check(StructurePresets.generate(StructurePresets.Shape.CUBE, 1) == null, "Reject invalid dimensions")

func _test_timeline() -> void:
	var timeline := AssemblyTimeline.new()
	check(timeline.duration == 4.0, "Default build lasts four seconds")
	check(AssemblyTimeline.duration_for(1729) == 4.0, "The reference cube takes four seconds")
	check(AssemblyTimeline.duration_for(10) == 4.0, "Small builds never go below four seconds")
	var castle_time := AssemblyTimeline.duration_for(70_000)
	check(castle_time > 12.0 and castle_time < AssemblyTimeline.duration_for(1_000_000), "Bigger builds take longer")
	check(AssemblyTimeline.duration_for(1_000_000) <= 24.0, "Assembly time is capped")
	timeline.configure(1729)
	timeline.restart()
	check(timeline.elapsed == 0.0, "Replay resets clock")
	timeline.advance(3.99)
	check(timeline.is_running(), "Still assembling before four seconds")
	timeline.advance(0.01)
	check(not timeline.is_running() and timeline.elapsed == 4.0, "Finished at exactly four seconds")
	timeline.advance(99)
	check(timeline.elapsed == 4.0, "Clock clamps at completion")
	for rank in [0.0, 0.25, 0.5, 0.75, 1.0]:
		check(is_zero_approx(timeline.cell_progress(0, rank)), "No cells assembled at time zero")
		check(is_equal_approx(timeline.cell_progress(4, rank), 1.0), "All cells settled at four seconds")
		var delay := timeline.release_time(rank)
		var a := timeline.cell_progress(delay + 0.3, rank)
		var b := timeline.cell_progress(delay + 0.6, rank)
		var c := timeline.cell_progress(delay + 0.9, rank)
		check(a > 0.0 and a < b and b < c and c < 1.0, "Gradual monotonic arrival")
		check(a > b - a and b - a > c - b, "Arrival decelerates")
	check(timeline.cell_progress(3.9, 1.0) < 1.0, "Last cell moves through end of animation")
	timeline.configure(70_000)
	check(timeline.duration == castle_time and timeline.release_window + timeline.settle_time == timeline.duration, "Large build timing is consistent")
	check(is_equal_approx(timeline.cell_progress(timeline.duration, 1.0), 1.0), "Large build settles at the end")

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
			"version": bytes.encode_u32(4, 3)
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

func _test_colors() -> void:
	var codec := OctaFile.new()
	var structure := OctaStructure.from_cells(PackedInt32Array([-1, -1, -1, 1, 1, 1, 0, 2, 0]), "Painted")
	var palette := PackedColorArray([Color8(200, 40, 30), Color8(20, 60, 200)])
	check(structure.color_index(0) == -1 and not structure.is_colored(), "Structures start uncolored")
	check(not structure.set_colors(PackedByteArray([0, 1]), palette).is_empty(), "Reject a color count mismatch")
	check(not structure.set_colors(PackedByteArray([0, 1, 2]), palette).is_empty(), "Reject an index outside the palette")
	var too_many := PackedColorArray()
	too_many.resize(17)
	check(not structure.set_colors(PackedByteArray([0, 0, 0]), too_many).is_empty(), "Reject palettes over 16 colors")
	check(structure.set_colors(PackedByteArray([0, 1, 1]), palette).is_empty(), "Accept valid colors")
	check(structure.color_index(1) == 1 and structure.storage_bytes() == 39, "13 bytes per colored cell")
	check(codec.save_structure(TEMP_FILE, structure) == OK, "Save colored file")
	var bytes := FileAccess.get_file_as_bytes(TEMP_FILE)
	check(bytes.decode_u32(4) == 2 and bytes.size() == 24 + 7 + 6 + 39, "Colored files are OCTA v2")
	var loaded := codec.load_structure(TEMP_FILE)
	check(loaded != null and loaded.packed_colors() == structure.packed_colors(), "Colors survive round trip")
	if loaded != null:
		check(loaded.palette.size() == 2 and loaded.palette[0].r8 == 200 and loaded.palette[1].b8 == 200, "Palette survives round trip")
	for mutation in ["index", "palette", "short"]:
		var broken := bytes.duplicate()
		match mutation:
			"index": broken[broken.size() - 1] = 2
			"palette": broken.encode_u32(20, 17)
			"short": broken.resize(broken.size() - 1)
		var file := FileAccess.open(TEMP_FILE, FileAccess.WRITE)
		file.store_buffer(broken)
		file.close()
		check(codec.load_structure(TEMP_FILE) == null, "Reject colored file with bad %s" % mutation)
	DirAccess.remove_absolute(TEMP_FILE)

## Brush-tool editing: nearest-cell lookup, picking, erasing regions and attached growth.
func _test_editor() -> void:
	for cell in [Vector3i.ZERO, Vector3i(1, 1, 1), Vector3i(-3, 5, -1), Vector3i(4, -2, 6)]:
		for jitter in [Vector3.ZERO, Vector3(0.48, 0, 0), Vector3(0, -0.48, 0), Vector3(0.24, 0.24, -0.24), Vector3(-0.2, 0.15, 0.1)]:
			check(OctaLattice.nearest(OctaLattice.to_world(cell, 0.8) + jitter * 0.8, 0.8) == cell, "Point near %s lies in that cell" % cell)
	var cube := StructurePresets.generate(StructurePresets.Shape.CUBE)
	var editor := OctaEditor.new(cube)
	check(editor.cell_count() == 1729 and not editor.has_changes(), "Editor starts from the structure")
	check(is_equal_approx(editor.ground_height(), cube.bounds().position.y), "Ground plane under the bottom layer")
	var hit := editor.pick(Vector3(0.1, 30, 0.2), Vector3.DOWN)
	check(hit.get("hit", false) and hit.cell.y >= 8 and hit.point.y <= 5.01 and hit.point.y > 3.9, "Ray picks the top layer")
	var floor_hit := editor.pick(Vector3(30, 20, 0), Vector3(0, -1, 0))
	check(not floor_hit.get("hit", true) and is_equal_approx(floor_hit.point.y, editor.ground_height()), "Ray beside the build reaches the ground")
	check(editor.pick(Vector3(30, 20, 0), Vector3.UP).is_empty(), "Ray into the sky misses")
	var top: Vector3 = hit.point
	var ball := editor.cells_within(top, 2.0)
	var half := editor.cells_within(top - Vector3.UP * 0.75, 2.0, Vector3.UP)
	check(not half.is_empty() and half.size() < ball.size(), "Scoop takes a shallow bite of the globe")
	var shallow := true
	for cell in half:
		shallow = shallow and OctaLattice.to_world(cell).y >= top.y - 0.75
	check(shallow, "Scoop stays above its cut plane")
	var tilted := editor.cells_within(top - Vector3.UP * 0.75, 2.0, Vector3(1, 1, 0).normalized())
	check(not tilted.is_empty() and tilted != half, "Tilted scoop takes a different bite")
	for cell in ball:
		check(editor.remove(cell), "Remove %s" % cell)
	check(not editor.remove(ball[0]) and editor.original_index(ball[0]) == -1, "Removed cells stay removed")
	check(editor.cell_count() == 1729 - ball.size() and editor.has_changes(), "Removal count")
	var slots := editor.free_slots(top, 1.5, 6)
	check(slots.size() == 6, "Brush finds attached slots")
	var attached := true
	var last_distance := 0.0
	for slot in slots:
		var touching := false
		for offset in OctaLattice.NEIGHBORS:
			touching = touching or editor.contains(slot + offset)
		var distance := OctaLattice.to_world(slot).distance_to(top)
		attached = attached and touching and not editor.contains(slot) and distance >= last_distance - 1e-4
		last_distance = distance
	check(attached, "Slots touch existing cells, nearest first")
	check(editor.add(slots[0]) and editor.contains(slots[0]) and editor.original_index(slots[0]) == -1, "Add a cell")
	check(not editor.add(slots[0]), "No duplicate additions")
	check(editor.remove(slots[0]) and not editor.contains(slots[0]), "Remove an added cell")
	editor.add(slots[1])
	editor.add(slots[0])
	var edited := editor.build()
	check(edited != null and edited.cell_count() == 1729 - ball.size() + 2, "Edited structure count")
	if edited != null:
		check(edited.coordinate(0) == cube.coordinate(0), "Survivors keep their order")
		check(edited.coordinate(edited.cell_count() - 2) == slots[1] and edited.coordinate(edited.cell_count() - 1) == slots[0], "Additions follow in order")
		check(not edited.is_colored(), "Uncolored builds stay uncolored")
	var painted := OctaStructure.from_cells(PackedInt32Array([0, 0, 0, 2, 0, 0, 1, 1, 1]), "Painted")
	painted.set_colors(PackedByteArray([0, 1, 0]), PackedColorArray([Color.RED, Color.BLUE]))
	var paint_editor := OctaEditor.new(painted)
	check(paint_editor.color_for(Vector3i(4, 0, 0)) == 1, "New cells adopt their neighbor's color")
	check(paint_editor.color_for(Vector3i(20, 0, 0), 1) == 1, "Unattached cells use the fallback color")
	paint_editor.add(Vector3i(4, 0, 0), paint_editor.color_for(Vector3i(4, 0, 0)))
	paint_editor.remove(Vector3i(1, 1, 1))
	var repainted := paint_editor.build()
	check(repainted.is_colored() and repainted.cell_count() == 3 and repainted.color_index(2) == 1 \
		and repainted.coordinate(2) == Vector3i(4, 0, 0) and repainted.color_index(1) == 1, "Colors follow edits")
	var empty := OctaEditor.new(OctaStructure.from_cells(PackedInt32Array()))
	var ground := empty.free_slots(Vector3(0, empty.ground_height(), 0), 1.2, 50)
	var on_floor := not ground.is_empty()
	for slot in ground:
		on_floor = on_floor and slot.y == empty.ground_layer
	check(on_floor and is_zero_approx(empty.ground_height()), "Empty builds grow from the floor")
	var castle := OctaFile.new().load_structure("res://structures/castle.octa")
	if castle != null:
		var started := Time.get_ticks_msec()
		var castle_editor := OctaEditor.new(castle)
		var indexed := Time.get_ticks_msec()
		castle_editor.remove(castle.coordinate(castle.cell_count() - 1))
		var rebuilt := castle_editor.build()
		check(rebuilt != null and rebuilt.cell_count() == castle.cell_count() - 1 and rebuilt.is_colored(), "Large build edit")
		print("Castle editor: index %d ms, build %d ms" % [indexed - started, Time.get_ticks_msec() - indexed])

## The Go generator's output (gen/, `go run ./cmd/octagen`) must load and be colored.
func _test_generated_builds() -> void:
	var codec := OctaFile.new()
	for name in [
		"aircraft_carrier", "alhambra", "amsterdam_canals", "angkor_wat", "arena", "battleship",
		"big_ben", "blue_mosque", "borobudur", "castle", "chichen_itza", "christ_the_redeemer",
		"cinque_terre", "cliff_palace", "clipper_ship", "colosseum_ruins", "cruise_ship",
		"dubrovnik", "easter_island", "eiffel_tower", "eilean_donan", "floating_islands",
		"forbidden_city", "funfair", "giza_pyramids", "golden_gate_bridge", "golden_pavilion",
		"gothic_cathedral", "great_wall", "hagia_sophia", "haunted_manor", "hillside_burrows",
		"ishtar_gate", "japanese_castle", "karnak_temple", "kinderdijk", "kremlin",
		"lighthouse_cove", "lunar_landing", "machu_picchu", "manhattan", "mars_base",
		"medieval_tower", "meteora", "mont_saint_michel", "monument_valley", "neuschwanstein",
		"oil_rig", "parthenon", "petra_treasury", "piazza_dei_miracoli", "pirate_galleon",
		"pont_du_gard", "potala_palace", "prague_old_town", "rothenburg", "sagrada_familia",
		"santorini", "saturn_v_launch", "st_basils_cathedral", "statue_of_liberty", "stave_church",
		"steam_train", "steampunk_airship", "sydney_opera_house", "taj_mahal", "temple_of_heaven",
		"tigers_nest", "tower_bridge", "treehouse_village", "venice_grand_canal", "viking_harbor",
		"volcano_island", "wild_west_town", "wizard_academy"
	]:
		var structure := codec.load_structure("res://structures/%s.octa" % name)
		check(structure != null, "Generated %s loads: %s" % [name, codec.last_error])
		if structure == null:
			continue
		check(structure.is_colored() and structure.palette.size() <= 16, "%s is colored with at most 16 colors" % name)
		var last_y: int = -(1 << 30)
		var sorted := true
		for index in structure.cell_count():
			var y: int = structure.coordinate(index).y
			sorted = sorted and y >= last_y
			last_y = y
		check(sorted, "%s assembles bottom-up" % name)
		print("%s: %d cells, %.1f s assembly" % [structure.title, structure.cell_count(), AssemblyTimeline.duration_for(structure.cell_count())])
