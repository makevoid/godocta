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
	check(is_zero_approx(studio.renderer.material_override.get_shader_parameter("cell_gap")), "X-ray shares parameters")
	studio.toggle_xray()
	check(not studio.renderer.xray, "X-ray toggles off")
	# Generated builds load from the gallery with their palette.
	var gallery := studio.hud.toolbar.gallery_paths
	check(gallery.size() >= 5, "Builds gallery lists generated structures")
	studio.hud.toolbar.gallery.get_popup().id_pressed.emit(gallery.find("res://structures/castle.octa"))
	await process_frame
	var castle := studio.renderer.structure
	check(castle != null and castle.title == "Castle" and castle.is_colored(), "Castle loads with colors")
	check(studio.renderer.timeline.duration > 4.0, "Large build takes longer to assemble")
	var palette: PackedColorArray = studio.renderer.material_override.get_shader_parameter("palette")
	check(palette.size() == 16 and palette[0].is_equal_approx(castle.palette[0].srgb_to_linear()), "Palette reaches the shader")
	check(studio.renderer.multimesh.buffer[15] == float(castle.color_index(0)), "Color index stored per instance")
	studio.renderer.display(OctaStructure.from_cells(PackedInt32Array()))
	check(studio.renderer.multimesh.instance_count == 0 and not studio.renderer.is_processing(), "Empty file renders safely")
	print("Scene integration: %s" % ("PASS" if failures == 0 else "FAIL"))
	quit(1 if failures else 0)

func _capture(name: String) -> void:
	await RenderingServer.frame_post_draw
	var result := root.get_texture().get_image().save_png("res://captures/%s.png" % name)
	check(result == OK, "Capture " + name)
	print("Captured ", name)

func check(condition: bool, message: String) -> void:
	if not condition:
		failures += 1
		push_error(message)
