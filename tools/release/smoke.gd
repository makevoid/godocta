extends SceneTree
## Use the matching Godot editor with --main-pack exported.pck --script this-file.gd.
## Official release templates disable external scripts; smoke their startup separately.
## Pass -- --capture=/absolute/path.png with a real rendering device for a screenshot.

var failures: int = 0

func _initialize() -> void:
	call_deferred("_run")

func check(condition: bool, message: String) -> void:
	if not condition:
		failures += 1
		push_error(message)

func _run() -> void:
	var studio = load("res://scenes/main.tscn").instantiate()
	root.add_child(studio)
	current_scene = studio
	await process_frame
	var expected_version := ""
	for argument in OS.get_cmdline_user_args():
		if argument.begins_with("--expect-version="):
			expected_version = argument.trim_prefix("--expect-version=")
	check(not expected_version.is_empty() and ProjectSettings.get_setting("application/config/version") == expected_version, "Release version")
	check(studio.hud.welcome.visible, "Welcome screen opens")
	check(studio.hud.toolbar.gallery_paths.size() == 101, "100 builds plus fixture are available")
	studio.hud.welcome.enter_button.pressed.emit()
	check(not studio.hud.has_overlay(), "Welcome enters studio")
	for shape in 4:
		studio.hud.toolbar.shapes.get_popup().id_pressed.emit(shape)
		check(studio.renderer.structure.cell_count() > 0, "Primitive loads")
	var gallery = studio.hud.toolbar.gallery_paths
	studio.hud.toolbar.gallery.get_popup().id_pressed.emit(gallery.find("res://structures/balloon_festival.octa"))
	await process_frame
	check(studio.renderer.structure.cell_count() == 105955, "Packaged Fable build loads")
	check(studio.environment.theme_name == "balloon_festival", "World theme loads")
	var codec = load("res://scripts/data/octa_file.gd").new()
	var temporary := OS.get_temp_dir().path_join("octa-release-%d.octa" % OS.get_process_id())
	check(codec.save_structure(temporary, studio.renderer.structure) == OK, "Save works")
	var restored = codec.load_structure(temporary)
	check(restored != null and restored.cell_count() == 105955, "Load round trip works")
	DirAccess.remove_absolute(temporary)
	studio.toggle_xray()
	check(studio.renderer.xray, "X-ray turns on")
	studio.toggle_xray()
	check(not studio.renderer.xray, "Solid mode returns")
	studio.renderer.seek(1.0)
	for frame in 10:
		await process_frame
	for argument in OS.get_cmdline_user_args():
		if argument.begins_with("--capture="):
			check(DisplayServer.get_name() != "headless", "Screenshot needs a real renderer")
			if DisplayServer.get_name() != "headless":
				RenderingServer.force_draw(false)
				var screenshot := root.get_texture().get_image()
				check(screenshot.save_png(argument.trim_prefix("--capture=")) == OK, "Screenshot saved")
	print("Export smoke: %s" % ("PASS" if failures == 0 else "FAIL"))
	quit(1 if failures else 0)
