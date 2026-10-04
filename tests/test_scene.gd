extends SceneTree
## Real scene integration and optional GPU captures. Pass -- --capture to save PNGs.

var studio: OctaStudio
var failures: int = 0
var capture: bool = false

func _initialize() -> void:
	capture = "--capture" in OS.get_cmdline_user_args()
	call_deferred("_run")

func _run() -> void:
	studio = load("res://scenes/main.tscn").instantiate() as OctaStudio
	root.add_child(studio)
	current_scene = studio
	await process_frame
	check(studio.renderer.structure.cell_count() == 1729, "Scene starts with cube")
	check(studio.renderer.multimesh.instance_count == 1729, "All cells uploaded")
	check(studio.renderer.get_child_count() == 0, "No node allocated per cell")
	if capture:
		DirAccess.make_dir_recursive_absolute(ProjectSettings.globalize_path("res://captures"))
	for shape in 4:
		# Exercise the toolbar signal wiring by pressing the actual Button.
		studio.hud.toolbar.shape_buttons[shape].pressed.emit()
		check(studio.renderer.timeline.elapsed == 0.0, "Button resets animation")
		check(studio.renderer.structure.title == StructurePresets.TITLES[shape], "Button selects matching structure")
		check(is_equal_approx(studio.camera.focus.y, studio.renderer.structure.bounds().size.y * 0.46), "Camera frames selected height")
		check(studio.hud.toolbar.shape_buttons[shape].button_pressed, "Selected toolbar button is highlighted")
		var mesh := studio.renderer.multimesh
		var buffer := mesh.buffer
		var first := Vector3(buffer[3], buffer[7], buffer[11])
		check(first.is_equal_approx(OctaLattice.to_world(studio.renderer.structure.coordinate(0))), "GPU buffer transform layout")
		check(is_equal_approx(buffer[(mesh.instance_count - 1) * 16 + 12], 1.0), "Last cell gets final stagger slot")
		# Dummy/headless rendering stores the buffer, but its per-instance getters
		# are stubs. Only request actual GPU readback with a rendering device.
		if DisplayServer.get_name() != "headless":
			check(mesh.get_instance_transform(0).origin.is_equal_approx(first), "GPU transform readback")
			check(is_equal_approx(mesh.get_instance_custom_data(mesh.instance_count - 1).r, 1.0), "GPU stagger readback")
		await create_timer(1.0).timeout
		check(studio.renderer.timeline.is_running(), "Animation is gradual")
		if capture and shape == 0:
			await _capture("cube-assembling")
		await studio.renderer.assembly_finished
		check(studio.renderer.timeline.elapsed == studio.renderer.timeline.duration, "Assembly completes on schedule")
		check(not studio.renderer.is_processing(), "No per-frame render updates after assembly")
		var deadline := Time.get_ticks_msec() + 10000
		while studio.renderer.surface_multimesh == null and Time.get_ticks_msec() < deadline:
			await process_frame
		check(studio.renderer.drawn_count() < studio.renderer.structure.cell_count(), "Settled builds draw only surface cells")
		check(studio.hud._state.text == "Assembly complete" and studio.hud._progress.value == 1.0, "HUD reflects completed assembly")
		if capture:
			await _capture(StructurePresets.TITLES[shape].to_lower())
	# Interrupt an in-flight preset and replay repeatedly: no stale tweens/signals.
	studio.select_preset(StructurePresets.Shape.SPHERE)
	await process_frame
	studio.select_preset(StructurePresets.Shape.PYRAMID)
	studio.renderer.replay()
	studio.renderer.replay()
	check(studio.renderer.structure.title == "Pyramid", "Rapid switching retains latest structure")
	check(studio.renderer.timeline.elapsed == 0.0, "Repeated replay has one reset clock")
	studio.renderer.set_gap(0.0)
	check(is_zero_approx(studio.renderer.material_override.get_shader_parameter("cell_gap")), "Zero gap enables exact tiling")
	# Presets are uncolored: the shader falls back to the base color.
	check(studio.renderer.multimesh.buffer[15] == -1.0, "Uncolored cells use the base color")
	# X-ray swaps the material and keeps shared parameters in sync.
	studio.toggle_xray()
	check(studio.renderer.xray and studio.hud._xray.button_pressed, "X-ray toggles on, HUD in sync")
	check(studio.renderer.drawn_count() == studio.renderer.structure.cell_count(), "X-ray draws every cell")
	check(is_zero_approx(studio.renderer.material_override.get_shader_parameter("cell_gap")), "X-ray shares parameters")
	studio.toggle_xray()
	check(not studio.renderer.xray, "X-ray toggles off")
	await _test_tools()
	# Generated builds load from the gallery with their palette.
	var gallery := studio.hud.toolbar.gallery_paths
	check(gallery.size() >= 5, "Builds gallery lists generated structures")
	studio.hud.toolbar.gallery.get_popup().id_pressed.emit(gallery.find("res://structures/castle.octa"))
	await process_frame
	var castle := studio.renderer.structure
	check(castle != null and castle.title == "Castle" and castle.is_colored(), "Castle loads with colors")
	check(studio.environment.theme_name == "castle", "Builds switch to their world theme")
	studio.select_preset(StructurePresets.Shape.CUBE)
	check(studio.environment.theme_name == WorldTheme.STUDIO, "Presets use the studio")
	studio.hud.toolbar.gallery.get_popup().id_pressed.emit(gallery.find("res://structures/castle.octa"))
	check(studio.renderer.timeline.duration > 4.0, "Large build takes longer to assemble")
	var palette: PackedColorArray = studio.renderer.material_override.get_shader_parameter("palette")
	check(palette.size() == 16 and palette[0].is_equal_approx(castle.palette[0].srgb_to_linear()), "Palette reaches the shader")
	check(studio.renderer.multimesh.buffer[15] == float(castle.color_index(0)), "Color index stored per instance")
	studio.renderer.display(OctaStructure.from_cells(PackedInt32Array()))
	check(studio.renderer.multimesh.instance_count == 0 and not studio.renderer.is_processing(), "Empty file renders safely")
	print("Scene integration: %s" % ("PASS" if failures == 0 else "FAIL"))
	quit(1 if failures else 0)

## Brush tools: palette wiring, pending edits, in-place commits and HUD counts.
func _test_tools() -> void:
	studio.select_preset(StructurePresets.Shape.CUBE)
	studio.renderer.set_gap(0.045)
	var tools := studio.tools
	var origin := studio.renderer.position
	studio.hud.tools.tool_buttons[StructureTools.Tool.BLACK_HOLE].pressed.emit()
	check(tools.tool == StructureTools.Tool.BLACK_HOLE and studio.hud.tools.tool_buttons[2].button_pressed, "Palette selects the black hole")
	var top := Vector3(0, studio.renderer.structure.bounds().end.y, 0)
	tools.apply(top)
	check(not studio.renderer.timeline.is_running(), "Editing finishes the assembly")
	check(tools.editor.cell_count() < 1729, "Black hole swallows cells")
	check(studio.renderer.drawn_count() == 1729, "Pending edits draw the full buffer")
	check(studio.renderer.structure.cell_count() == 1729, "Edits stay pending until committed")
	for frame in 10:
		await process_frame
	tools.commit()
	var eaten := studio.renderer.structure.cell_count()
	check(eaten < 1729 and studio.hud._count.text == studio.hud._format_count(eaten), "Commit rebuilds and updates the HUD")
	check(studio.renderer.position == origin, "Edits keep the build in place")
	studio.set_scoop_tilt(Vector2i(45, 0))
	check(studio.hud.tools._tilt_buttons[0][2].button_pressed and studio.hud.tools._tilt_buttons[1][1].button_pressed, "Tilt shows in the palette")
	studio.select_tool(StructureTools.Tool.SCOOP)
	var axis := tools.scoop_axis()
	check(is_equal_approx(axis.angle_to(Vector3.UP), PI / 4.0), "Tilted scoop leans 45°")
	tools.apply(Vector3(3, top.y - 0.75, 3), axis)
	tools.commit()
	check(studio.renderer.structure.cell_count() < eaten, "Scoop shaves cells")
	var before := studio.renderer.structure.cell_count()
	studio.set_brush_radius(1.5)
	studio.hud.tools.tool_buttons[StructureTools.Tool.BRUSH_IN].pressed.emit()
	tools.apply(top + Vector3.DOWN * 2.0, Vector3.UP, Vector3.UP)
	check(tools.editor.cell_count() > before, "Brush adds attracted cells")
	# Pending edits commit by themselves once the stroke goes idle.
	await create_timer(StructureTools.COMMIT_DELAY + 0.4).timeout
	check(studio.renderer.structure.cell_count() > before, "Idle edits commit automatically")
	check(tools._pending.multimesh.instance_count == 0, "Committed cells leave the pending layer")
	# Holding the brush in one spot fills its globe and stops, then reports itself full.
	var full := [false]
	tools.brush_full.connect(func() -> void: full[0] = true)
	for tick in 40:
		tools.apply(top, Vector3.UP, Vector3.UP)
	check(full[0], "A brush with no free slot left reports itself full")
	var reach := tools.world_radius() + tools.editor.span()
	var strays := 0
	for cell: Vector3i in tools.editor._added:
		if OctaLattice.to_world(cell, tools.editor.span()).distance_to(top) > reach:
			strays += 1
	check(strays == 0, "Brush fills only inside its globe")
	tools.apply(top, Vector3.UP, Vector3.UP)
	studio.replay()
	check(not tools.editor.has_changes() and studio.renderer.timeline.is_running(), "Replay includes pending edits")
	tools.apply(top, Vector3.UP, Vector3.UP)
	studio.select_preset(StructurePresets.Shape.CUBE)
	check(studio.renderer.structure.cell_count() == 1729 and tools.editor == null, "Switching builds discards pending edits")
	_test_brush_colors()
	studio.select_tool(StructureTools.Tool.NAVIGATE)
	studio.set_scoop_tilt(Vector2i.ZERO)

## The brush color submenu: auto on uncolored builds, one palette color on colored ones.
func _test_brush_colors() -> void:
	var tools := studio.tools
	var menu := studio.hud.tools.color_menu
	menu.get_popup().id_pressed.emit(5)
	check(tools.brush_color == -1 and menu.disabled, "Uncolored builds keep the brush on auto")
	var swatch := OctaStructure.from_cells(PackedInt32Array([0, 0, 0, 2, 0, 0]), "Swatch")
	swatch.set_colors(PackedByteArray([0, 0]), OctaPalette.colors())
	studio._present(swatch)
	check(not menu.disabled and tools.brush_color == -1, "A colored build opens the submenu, still on auto")
	studio.select_tool(StructureTools.Tool.BRUSH_IN)
	studio.set_brush_radius(1.5)
	tools.apply(OctaLattice.to_world(Vector3i.ZERO, swatch.cell_span), Vector3.UP, Vector3.UP)
	var auto_cells: Array = tools.editor._added.keys()
	var voted := not auto_cells.is_empty()
	for cell: Vector3i in auto_cells:
		voted = voted and tools.editor.color_of(cell) == 0
	check(voted, "Auto strokes copy their neighbors' color")
	menu.get_popup().id_pressed.emit(13)
	check(tools.brush_color == 12, "The submenu picks a palette color")
	check(menu.text == "Gold" and tools.gizmo.brush_tint != null, "The button and the gizmo show it")
	tools.apply(OctaLattice.to_world(Vector3i(2, 0, 0), swatch.cell_span), Vector3.UP, Vector3.UP)
	var gold := 0
	for cell: Vector3i in tools.editor._added:
		gold += 1 if tools.editor.color_of(cell) == 12 else 0
	check(gold > 0 and gold == tools.editor._added.size() - auto_cells.size(), "Strokes paint the chosen color")
	menu.get_popup().id_pressed.emit(0)
	check(tools.brush_color == -1 and tools.gizmo.brush_tint == null, "Auto comes back from the submenu")
	studio.select_preset(StructurePresets.Shape.CUBE)
	check(tools.brush_color == -1 and menu.disabled, "An uncolored build falls back to auto")

func _capture(name: String) -> void:
	await RenderingServer.frame_post_draw
	var result := root.get_texture().get_image().save_png("res://captures/%s.png" % name)
	check(result == OK, "Capture " + name)
	print("Captured ", name)

func check(condition: bool, message: String) -> void:
	if not condition:
		failures += 1
		push_error(message)
