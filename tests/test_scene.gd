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
	_test_interface()
	if capture:
		DirAccess.make_dir_recursive_absolute(ProjectSettings.globalize_path("res://captures"))
	for shape in 4:
		# Exercise the toolbar signal wiring through the actual Primitives menu.
		studio.hud.toolbar.shapes.get_popup().id_pressed.emit(shape)
		check(studio.renderer.timeline.elapsed == 0.0, "Button resets animation")
		check(studio.renderer.structure.title == StructurePresets.TITLES[shape], "Button selects matching structure")
		check(is_equal_approx(studio.camera.focus.y, studio.renderer.structure.bounds().size.y * 0.46), "Camera frames selected height")
		check(studio.hud.toolbar.shapes.get_popup().is_item_checked(shape), "Selected primitive is checked in the menu")
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
		check(studio.hud._state.text == "Assembly complete" and studio.hud._progress.value == 1.0
			and not studio.hud._progress.visible, "HUD reflects completed assembly and hides the bar")
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
	check(studio.renderer.xray and studio.hud.toolbar.is_setting_checked(StructureToolbar.Setting.XRAY),
		"X-ray toggles on, settings menu in sync")
	check(studio.renderer.drawn_count() == studio.renderer.structure.cell_count(), "X-ray draws every cell")
	check(is_zero_approx(studio.renderer.material_override.get_shader_parameter("cell_gap")), "X-ray shares parameters")
	studio.toggle_xray()
	check(not studio.renderer.xray, "X-ray toggles off")
	await _test_tools()
	_test_black_hole_depth()
	_test_undo_redo()
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

## Menu actions, search, and overlays must not leak input into the sandbox.
func _test_interface() -> void:
	check(studio.hud.welcome.visible and not studio.camera.is_processing_unhandled_input(),
		"Welcome owns input on launch")
	studio.hud.welcome.enter_button.pressed.emit()
	check(not studio.hud.has_overlay() and studio.camera.is_processing_unhandled_input(),
		"Enter opens the studio and restores navigation")
	studio.hud.toolbar.library_requested.emit()
	var library := studio.hud.library
	check(library.visible and not studio.tools.is_processing(), "Build library suspends brush processing")
	library.search.text_changed.emit("castle")
	var matches := library._buttons.filter(func(button: Button) -> bool: return button.visible)
	check(not matches.is_empty(), "Build search finds castle filenames")
	for button: Button in matches:
		check(button.text.to_lower().contains("castle"), "Search hides nonmatching filenames")
	library.search.text_changed.emit("no-such-octa-world-123")
	check(library._empty.visible, "An empty search has an explanation")
	library.search.text_changed.emit("")
	var castle := studio.hud.toolbar.gallery_paths.find("res://structures/castle.octa")
	library._buttons[castle].pressed.emit()
	check(studio.renderer.structure.title == "Castle" and not studio.hud.has_overlay(),
		"Library cards load their file and dismiss the menu")
	studio.hud.toolbar.home_requested.emit()
	check(studio.hud.welcome.visible and studio.renderer.structure.title == "Castle",
		"Home preserves the current structure")
	var key := InputEventKey.new()
	key.pressed = true
	key.physical_keycode = KEY_X
	studio._unhandled_key_input(key)
	check(not studio.renderer.xray, "Welcome suppresses world shortcuts")
	key.physical_keycode = KEY_ENTER
	studio._unhandled_key_input(key)
	check(not studio.hud.welcome.visible, "Enter also opens the studio from the keyboard")
	studio.hud.toggle_help()
	key.physical_keycode = KEY_B
	studio._unhandled_key_input(key)
	check(studio.tools.tool == StructureTools.Tool.NAVIGATE, "Help suppresses tool shortcuts")
	key.physical_keycode = KEY_ESCAPE
	studio._unhandled_key_input(key)
	check(not studio.hud.has_overlay() and studio.tools.is_processing(), "Escape closes help and restores tools")
	studio.select_preset(StructurePresets.Shape.CUBE)

## Brush tools: palette wiring, pending edits, in-place commits and HUD counts.
func _test_tools() -> void:
	studio.select_preset(StructurePresets.Shape.CUBE)
	studio.renderer.set_gap(0.045)
	var tools := studio.tools
	var origin := studio.renderer.position
	var options := studio.hud.options
	check(not studio.hud.tools.tool_buttons[StructureTools.Tool.SCOOP].visible, "The scoop is off by default")
	studio.hud.toolbar.settings.get_popup().id_pressed.emit(StructureToolbar.Setting.SCOOP)
	check(studio.scoop_enabled and studio.hud.tools.tool_buttons[StructureTools.Tool.SCOOP].visible,
		"The settings menu enables the scoop")
	studio.hud.tools.tool_buttons[StructureTools.Tool.BLACK_HOLE].pressed.emit()
	check(tools.tool == StructureTools.Tool.BLACK_HOLE and studio.hud.tools.tool_buttons[2].button_pressed, "Palette selects the black hole")
	check(options.visible and not options._color_box.visible, "The black hole shows its size, not a color")
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
	check(options._tilt_buttons[0][2].button_pressed and options._tilt_buttons[1][1].button_pressed, "Tilt shows in the options")
	studio.select_tool(StructureTools.Tool.SCOOP)
	check(options._tilt_box.visible, "The scoop reveals its tilt")
	var axis := tools.scoop_axis()
	check(is_equal_approx(axis.angle_to(Vector3.UP), PI / 4.0), "Tilted scoop leans 45°")
	tools.apply(Vector3(3, top.y - 0.75, 3), axis)
	tools.commit()
	check(studio.renderer.structure.cell_count() < eaten, "Scoop shaves cells")
	var before := studio.renderer.structure.cell_count()
	studio.hud.tools.tool_buttons[StructureTools.Tool.BRUSH_IN].pressed.emit()
	check(is_equal_approx(tools.radius, 3.0), "The brush starts at three cells")
	studio.set_brush_radius(1.5)
	studio.select_tool(StructureTools.Tool.BLACK_HOLE)
	check(is_equal_approx(tools.radius, 3.0), "Each tool keeps its own size")
	studio.select_tool(StructureTools.Tool.BRUSH_IN)
	check(is_equal_approx(tools.radius, 1.5) and is_equal_approx(studio.hud.options._radius.value, 1.5),
		"Switching back restores the last size, in the options too")
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
	studio.select_tool(StructureTools.Tool.SCOOP)
	studio.set_scoop_enabled(false)
	check(tools.tool == StructureTools.Tool.NAVIGATE and not options.visible,
		"Withdrawing the scoop falls back to orbit and hides the options")
	studio.set_scoop_tilt(Vector2i.ZERO)

## A held black hole stays local after its first bite, including while dragging.
func _test_black_hole_depth() -> void:
	var tools := studio.tools
	var foreground := Vector3i(0, 4, 8)
	var beside := Vector3i(8, 4, 8)
	var background := Vector3i(0, 4, -8)
	var behind_beside := Vector3i(8, 4, -8)
	studio._present(OctaStructure.from_cells(PackedInt32Array([
		0, 4, 8, 8, 4, 8, 0, 4, -8, 8, 4, -8,
	]), "Depth test"))
	studio.select_tool(StructureTools.Tool.BLACK_HOLE)
	studio.set_brush_radius(1.5)
	tools._ensure_editor()
	var origin := Vector3(0, 2, 12)
	var direction := Vector3.FORWARD
	var press := InputEventMouseButton.new()
	press.button_index = MOUSE_BUTTON_LEFT
	press.pressed = true
	var release := InputEventMouseButton.new()
	release.button_index = MOUSE_BUTTON_LEFT
	tools._unhandled_input(press)
	tools._update_brush_ray(origin, direction, StructureTools.TICK)
	var center := tools.gizmo.position
	check(not tools.editor.contains(foreground) and tools.editor.contains(background),
		"The black hole's first bite removes only the foreground")
	for tick in 20:
		tools._update_brush_ray(origin, direction, StructureTools.TICK)
	check(tools.editor.contains(background) and tools.gizmo.position.is_equal_approx(center),
		"Holding the black hole never retargets the exposed background")
	tools._update_brush_ray(origin, Vector3(4, 0, -8).normalized(), StructureTools.TICK)
	check(not tools.editor.contains(beside) and tools.editor.contains(behind_beside),
		"Dragging the black hole removes nearby foreground cells at the original depth")
	for tick in 20:
		tools._update_brush_ray(origin + Vector3.RIGHT * 4, direction, StructureTools.TICK)
	check(tools.editor.contains(behind_beside), "Dragging over a cleared area preserves the background")
	tools._update_brush_ray(Vector3(-8, 2, -4), Vector3.RIGHT, StructureTools.TICK)
	check(not tools.gizmo.visible and tools.editor.contains(background),
		"A ray parallel to the stroke plane cannot fall back to background cells")
	tools._unhandled_input(release)
	tools.commit()
	check(studio.renderer.structure.cell_count() == 2, "Committing the stroke preserves both background cells")
	tools._unhandled_input(press)
	tools._update_brush_ray(origin, direction, StructureTools.TICK)
	check(not tools.editor.contains(background) and tools.editor.contains(behind_beside),
		"A fresh click can deliberately target the background")
	tools._unhandled_input(release)
	studio.set_brush_radius(3.0)
	studio.select_preset(StructurePresets.Shape.CUBE)
	studio.select_tool(StructureTools.Tool.NAVIGATE)

## The brush colors: auto on uncolored builds, one palette color on colored ones.
func _test_brush_colors() -> void:
	var tools := studio.tools
	var options := studio.hud.options
	options.color_buttons[4].pressed.emit()
	check(tools.brush_color == -1 and options.color_buttons[4].disabled, "Uncolored builds keep the brush on auto")
	var swatch := OctaStructure.from_cells(PackedInt32Array([0, 0, 0, 2, 0, 0]), "Swatch")
	swatch.set_colors(PackedByteArray([0, 0]), OctaPalette.colors())
	studio._present(swatch)
	check(not options.color_buttons[4].disabled and tools.brush_color == -1, "A colored build offers the grid, still on auto")
	studio.select_tool(StructureTools.Tool.BRUSH_IN)
	check(options._color_box.visible and not options.more_expanded, "The brush shows eight colors behind More")
	studio.set_brush_radius(1.5)
	tools.apply(OctaLattice.to_world(Vector3i.ZERO, swatch.cell_span), Vector3.UP, Vector3.UP)
	var auto_cells: Array = tools.editor._added.keys()
	var voted := not auto_cells.is_empty()
	for cell: Vector3i in auto_cells:
		voted = voted and tools.editor.color_of(cell) == 0
	check(voted, "Auto strokes copy their neighbors' color")
	options.color_buttons[12].pressed.emit()
	check(tools.brush_color == 12, "The grid picks a palette color")
	check(options.color_buttons[12].button_pressed and options.more_expanded and not options.auto_button.button_pressed,
		"A color past the first eight unfolds the rest of the palette")
	check(tools.gizmo.brush_tint != null, "The gizmo shows the painted color")
	tools.apply(OctaLattice.to_world(Vector3i(2, 0, 0), swatch.cell_span), Vector3.UP, Vector3.UP)
	var gold := 0
	for cell: Vector3i in tools.editor._added:
		gold += 1 if tools.editor.color_of(cell) == 12 else 0
	check(gold > 0 and gold == tools.editor._added.size() - auto_cells.size(), "Strokes paint the chosen color")
	options.auto_button.pressed.emit()
	check(tools.brush_color == -1 and tools.gizmo.brush_tint == null, "Auto comes back from the grid")
	studio.select_preset(StructurePresets.Shape.CUBE)
	check(tools.brush_color == -1 and options.color_buttons[4].disabled, "An uncolored build falls back to auto")
	check(not options._color_box.visible, "Uncolored builds hide the color grid")

## Each mouse stroke is an action, including pending effects and an entirely erased build.
func _test_undo_redo() -> void:
	var tools := studio.tools
	var dock := studio.hud.tools
	var source := OctaStructure.from_cells(PackedInt32Array([0, -2, 0, 2, -2, 0]), "History", 0.5)
	source.set_colors(PackedByteArray([3, 12]), OctaPalette.colors())
	studio._present(source)
	studio.select_tool(StructureTools.Tool.BRUSH_IN)
	studio.set_brush_radius(1.5)
	studio.set_brush_color(7)
	var origin := studio.renderer.position
	var focus := studio.camera.focus
	var center := Vector3(0, -0.5, 0)
	var press := InputEventMouseButton.new()
	press.button_index = MOUSE_BUTTON_LEFT
	press.pressed = true
	var release := InputEventMouseButton.new()
	release.button_index = MOUSE_BUTTON_LEFT
	check(dock.undo_button.disabled and dock.redo_button.disabled, "History starts disabled")
	tools._unhandled_input(press)
	tools.apply(center)
	tools.apply(center)
	tools._animate()
	var first := tools.editor.build()
	check(first.cell_count() > source.cell_count() and not dock.undo_button.disabled,
		"A pending brush stroke enables Undo")
	tools._unhandled_input(release)
	tools._unhandled_input(press)
	_check_snapshot(studio.renderer.structure, first, "Next click commits the previous stroke")
	tools.apply(Vector3(2, -0.5, 0))
	tools._animate()
	var second := tools.editor.build()
	check(second.cell_count() > first.cell_count(), "Second rapid stroke also adds cells")
	dock.undo_button.pressed.emit()
	_check_snapshot(studio.renderer.structure, first, "Undo affects only the last rapid stroke")
	check(not tools._stroking and tools._pending.multimesh.instance_count == 0,
		"Undo ends the stroke and clears flying additions immediately")
	dock.undo_button.pressed.emit()
	_check_snapshot(studio.renderer.structure, source, "Undo restores the original colored build")
	check(dock.undo_button.disabled and not dock.redo_button.disabled, "History buttons follow availability")
	var key := InputEventKey.new()
	key.pressed = true
	key.physical_keycode = KEY_Z
	key.meta_pressed = true
	key.shift_pressed = true
	studio._unhandled_key_input(key)
	_check_snapshot(studio.renderer.structure, first, "Cmd+Shift+Z redoes exact colors and order")
	key.meta_pressed = false
	key.ctrl_pressed = true
	key.shift_pressed = false
	key.physical_keycode = KEY_Y
	studio._unhandled_key_input(key)
	_check_snapshot(studio.renderer.structure, second, "Ctrl+Y redoes the second stroke")
	studio.select_tool(StructureTools.Tool.BLACK_HOLE)
	studio.set_brush_radius(12.0)
	tools.apply(center)
	tools._animate()
	check(tools.editor.cell_count() == 0 and tools._vortex.multimesh.instance_count > 0,
		"Black hole can swallow the whole build")
	key.physical_keycode = KEY_Z
	studio._unhandled_key_input(key)
	_check_snapshot(studio.renderer.structure, second, "Ctrl+Z restores cells during the swallow animation")
	check(tools._vortex.multimesh.instance_count == 0 and tools._swallowed.is_empty(),
		"Undo clears swallowed effects immediately")
	tools.apply(Vector3(100, 100, 100))
	tools.commit()
	check(tools.can_redo(), "An empty stroke preserves redo")
	dock.redo_button.pressed.emit()
	check(studio.renderer.structure.cell_count() == 0 and tools.editor.ground_layer == -2,
		"Redo restores an empty build without moving its ground")
	check(studio.renderer.position == origin and studio.camera.focus == focus,
		"History preserves renderer origin and camera framing")
	studio.select_tool(StructureTools.Tool.BRUSH_IN)
	studio.set_brush_radius(1.5)
	tools.apply(center)
	tools.commit()
	check(studio.renderer.structure.is_colored() and studio.renderer.structure.color_index(0) == 7,
		"Painting after a complete erase retains the palette and chosen color")
	check(tools.undo() and tools.undo(), "Multiple actions remain undoable")
	studio.hud.set_help_visible(true)
	var blocked := studio.renderer.structure
	studio._unhandled_key_input(key)
	check(studio.renderer.structure == blocked, "Modal help suppresses undo shortcuts")
	studio.hud.set_help_visible(false)
	tools.apply(Vector3(3, -0.5, 0))
	check(not tools.can_redo() and dock.redo_button.disabled, "A new edit immediately clears redo")
	studio.select_tool(StructureTools.Tool.BLACK_HOLE)
	check(not tools.editor.has_changes(), "Switching tools commits the previous action")
	studio.replay()
	check(tools._undo.size() == 3, "Replay does not add an empty history entry")
	_test_history_limit()
	studio.select_preset(StructurePresets.Shape.CUBE)
	check(not tools.can_undo() and not tools.can_redo() and dock.undo_button.disabled and dock.redo_button.disabled,
		"Switching builds clears both history stacks")
	studio.select_tool(StructureTools.Tool.NAVIGATE)

func _test_history_limit() -> void:
	var cells := PackedInt32Array()
	for index in 12:
		cells.append_array(PackedInt32Array([index * 4, -2, 0]))
	var data := OctaStructure.from_cells(cells, "History limit", 0.5)
	studio._present(data)
	studio.select_tool(StructureTools.Tool.BLACK_HOLE)
	studio.set_brush_radius(0.6)
	var states: Array[OctaStructure] = [data]
	for index in 12:
		studio.tools.apply(Vector3(index, -0.5, 0))
		studio.tools.commit()
		states.append(studio.renderer.structure)
		check(studio.renderer.structure.cell_count() == 11 - index, "Limit fixture removes one cell per stroke")
	for index in 10:
		check(studio.tools.undo(), "The last ten strokes can be undone")
		_check_snapshot(studio.renderer.structure, states[11 - index], "Undo restores the matching state")
	check(not studio.tools.undo(), "The eleventh Undo stops at the history limit")
	for index in 10:
		check(studio.tools.redo(), "All ten undone strokes can be redone")
		_check_snapshot(studio.renderer.structure, states[3 + index], "Redo restores the matching state")
	check(not studio.tools.redo(), "Redo stops after ten restored strokes")
	studio.tools.undo()
	# A successful file load also uses _present(), and must clear redo as well as undo.
	studio._present(data)
	check(not studio.tools.can_undo() and not studio.tools.can_redo(), "Presenting a build clears both stacks")
	studio.tools.apply(Vector3(0, -0.5, 0))
	studio.tools.commit()

func _check_snapshot(actual: OctaStructure, expected: OctaStructure, message: String) -> void:
	check(actual.packed_cells() == expected.packed_cells() and actual.packed_colors() == expected.packed_colors()
		and actual.palette == expected.palette and actual.title == expected.title
		and actual.cell_span == expected.cell_span, message)

func _capture(name: String) -> void:
	await RenderingServer.frame_post_draw
	var result := root.get_texture().get_image().save_png("res://captures/%s.png" % name)
	check(result == OK, "Capture " + name)
	print("Captured ", name)

func check(condition: bool, message: String) -> void:
	if not condition:
		failures += 1
		push_error(message)
