class_name OctaStudio
extends Node3D
## Composition root. Geometry, storage, presets, animation, camera and UI are separate classes.

var renderer := StructureRenderer.new()
var camera := OrbitCamera.new()
var hud := StudioHUD.new()
var tools := StructureTools.new()
var environment: StudioEnvironment
var _files := OctaFile.new()
var _save_dialog: FileDialog
var _load_dialog: FileDialog
## The scoop is opt-in: the settings menu adds it to the tool strip.
var scoop_enabled: bool = false

func _ready() -> void:
	environment = StudioEnvironment.new()
	add_child(environment)
	add_child(renderer)
	add_child(camera)
	# After the camera: unhandled input reaches later nodes first, so a brush stroke
	# consumes the left drag before the camera can orbit with it.
	tools.renderer = renderer
	tools.camera = camera
	add_child(tools)
	add_child(hud)
	camera.current = true
	hud.toolbar.shape_selected.connect(select_preset)
	hud.toolbar.file_selected.connect(_load_structure)
	hud.toolbar.replay_requested.connect(replay)
	hud.toolbar.save_requested.connect(_choose_save)
	hud.toolbar.load_requested.connect(_choose_load)
	hud.toolbar.reset_view_requested.connect(camera.reset_view)
	hud.toolbar.help_requested.connect(hud.toggle_help)
	hud.toolbar.home_requested.connect(func() -> void: hud.set_welcome_visible(true))
	hud.toolbar.library_requested.connect(func() -> void: hud.set_library_visible(true))
	hud.welcome.enter_requested.connect(enter_studio)
	hud.welcome.library_requested.connect(func() -> void: hud.set_library_visible(true))
	hud.welcome.load_requested.connect(_choose_load)
	hud.library.file_selected.connect(_load_structure)
	hud.overlay_changed.connect(_update_overlay_input)
	hud.toolbar.orbit_changed.connect(func(enabled: bool) -> void: camera.auto_orbit = enabled)
	hud.toolbar.xray_changed.connect(renderer.set_xray)
	hud.toolbar.scoop_enabled_changed.connect(set_scoop_enabled)
	renderer.assembly_progress.connect(hud.show_progress)
	hud.tools.tool_selected.connect(select_tool)
	hud.tools.undo_requested.connect(undo)
	hud.tools.redo_requested.connect(redo)
	tools.history_changed.connect(hud.tools.show_history)
	hud.options.radius_changed.connect(set_brush_radius)
	hud.options.tilt_changed.connect(set_scoop_tilt)
	hud.options.color_selected.connect(set_brush_color)
	tools.committed.connect(hud.show_counts)
	tools.brush_full.connect(func() -> void: hud.notify("Brush full here — release, then click again for another layer"))
	_save_dialog = _make_dialog(FileDialog.FILE_MODE_SAVE_FILE)
	_save_dialog.file_selected.connect(_save_structure)
	_load_dialog = _make_dialog(FileDialog.FILE_MODE_OPEN_FILE)
	_load_dialog.file_selected.connect(_load_structure)
	_fit_window_to_screen()
	select_preset(StructurePresets.Shape.CUBE)
	hud.set_welcome_visible(true)

## Enter resumes the current arrangement, including edits made before visiting Home.
func enter_studio() -> void:
	hud.set_welcome_visible(false)
	replay()

## Full-screen menus own input. A brush cannot keep editing behind an overlay.
func _update_overlay_input() -> void:
	var blocked := hud.has_overlay()
	if blocked:
		tools._end_stroke()
		tools.commit()
		tools.gizmo.hide()
		camera._dragging = false
		Input.set_default_cursor_shape(Input.CURSOR_ARROW)
	else:
		tools.set_tool(tools.tool)
	tools.set_process(not blocked)
	tools.set_process_unhandled_input(not blocked)
	camera.set_process(not blocked)
	camera.set_process_unhandled_input(not blocked)

func select_preset(shape: StructurePresets.Shape) -> void:
	var structure := StructurePresets.generate(shape)
	_present(structure, shape)
	hud.set_welcome_visible(false)

func _present(structure: OctaStructure, preset: int = -1) -> void:
	tools.reset()
	hud.show_structure(structure, preset)
	renderer.display(structure)
	# Palettes differ between builds: keep the chosen color only while it still exists.
	hud.options.show_palette(structure.palette)
	set_brush_color(tools.brush_color)
	var box := structure.bounds()
	environment.apply_theme(WorldTheme.STUDIO if preset >= 0 else WorldTheme.name_for(structure.title))
	camera.default_pitch = 0.48 if environment.theme_name == WorldTheme.STUDIO else 0.24
	camera.frame_structure(box)
	environment.fit(box)

## Window sizes, screen sizes and usable rects are all in pixels, while the UI is laid
## out in the 1900 × 1000 points of WINDOW_POINTS. `screen_get_scale()` reports the
## pixels per point: 2 on a Retina/HiDPI screen, 1 on an ordinary one. The window asks
## for 3800 × 2000 pixels on HiDPI and half that, 1900 × 1000, without it, so it covers
## the same physical area either way and the canvas_items stretch scales the UI to match.
## It shrinks to fit smaller displays; an explicit --resolution (pixels) is left alone.
const WINDOW_POINTS := Vector2(1900, 1000)
const TITLE_BAR_POINTS := 32.0

func _fit_window_to_screen() -> void:
	var window := get_window()
	if window.mode != Window.MODE_WINDOWED or OS.get_cmdline_args().has("--resolution"):
		return
	var screen := window.current_screen
	var usable := DisplayServer.screen_get_usable_rect(screen)
	if usable.size.x <= 0 or usable.size.y <= 0:
		return
	# Pixels per layout point: 2 on HiDPI, 1 elsewhere, which halves the pixel count.
	var scale := maxf(DisplayServer.screen_get_scale(screen), 1.0)
	var target := WINDOW_POINTS * scale
	var room := Vector2(usable.size) - Vector2(0.0, TITLE_BAR_POINTS * scale)
	var fit := minf(1.0, minf(room.x / target.x, room.y / target.y))
	window.size = Vector2i(target * fit)
	window.position = usable.position + (usable.size - window.size) / 2

## Replays the current arrangement, including pending brush edits.
func replay() -> void:
	tools.commit()
	renderer.replay()

func undo() -> void:
	if tools.undo():
		hud.notify("Undid last stroke")

func redo() -> void:
	if tools.redo():
		hud.notify("Redid stroke")

func select_tool(tool: StructureTools.Tool) -> void:
	tools.set_tool(tool)
	hud.tools.show_tool(tool)
	hud.options.show_tool(tool)
	# Each tool comes back with the size it was last used at.
	hud.options.show_radius(tools.radius)

## Add or withdraw the scoop. Withdrawing it while it is in use falls back to orbit.
func set_scoop_enabled(enabled: bool) -> void:
	scoop_enabled = enabled
	hud.tools.set_scoop_available(enabled)
	hud.toolbar.set_setting_checked(StructureToolbar.Setting.SCOOP, enabled)
	if not enabled and tools.tool == StructureTools.Tool.SCOOP:
		select_tool(StructureTools.Tool.NAVIGATE)

func set_brush_radius(cells: float) -> void:
	tools.set_radius(cells)
	hud.options.show_radius(tools.radius)

## -1 lets the brush adopt neighbors' colors; 0..15 paints one color of the build's palette.
func set_brush_color(index: int) -> void:
	tools.set_brush_color(index)
	hud.options.show_color(tools.brush_color)

func set_scoop_tilt(tilt: Vector2i) -> void:
	tools.tilt = tilt.clamp(Vector2i(-45, -45), Vector2i(45, 45))
	hud.options.show_tilt(tools.tilt)

func toggle_xray() -> void:
	renderer.set_xray(not renderer.xray)
	hud.set_xray(renderer.xray)

func _unhandled_key_input(event: InputEvent) -> void:
	if not event is InputEventKey or not event.pressed or event.echo:
		return
	# Keep shortcuts out of modal file dialogs.
	if _save_dialog.visible or _load_dialog.visible:
		return
	if hud.library.visible:
		if event.physical_keycode == KEY_ESCAPE:
			hud.set_library_visible(false)
		get_viewport().set_input_as_handled()
		return
	if hud.welcome.visible:
		if event.physical_keycode in [KEY_ENTER, KEY_KP_ENTER, KEY_ESCAPE]:
			enter_studio()
		get_viewport().set_input_as_handled()
		return
	if hud.help.visible:
		if event.physical_keycode in [KEY_ESCAPE, KEY_H, KEY_F1]:
			hud.set_help_visible(false)
		get_viewport().set_input_as_handled()
		return
	if (event.ctrl_pressed or event.meta_pressed) and not event.alt_pressed:
		if event.physical_keycode == KEY_Z:
			if event.shift_pressed:
				redo()
			else:
				undo()
			get_viewport().set_input_as_handled()
		elif event.physical_keycode == KEY_Y:
			redo()
			get_viewport().set_input_as_handled()
		return
	match event.physical_keycode:
		KEY_1, KEY_2, KEY_3, KEY_4:
			select_preset((event.physical_keycode - KEY_1) as StructurePresets.Shape)
		KEY_5, KEY_6, KEY_7, KEY_8, KEY_9:
			var index: int = event.physical_keycode - KEY_5
			if index < hud.toolbar.gallery_paths.size():
				_load_structure(hud.toolbar.gallery_paths[index])
		KEY_X:
			toggle_xray()
		KEY_R, KEY_SPACE:
			replay()
		KEY_F:
			camera.reset_view()
		KEY_V, KEY_ESCAPE:
			select_tool(StructureTools.Tool.NAVIGATE)
		KEY_B:
			select_tool(StructureTools.Tool.BRUSH_IN)
		KEY_E:
			select_tool(StructureTools.Tool.BLACK_HOLE)
		KEY_S:
			if scoop_enabled:
				select_tool(StructureTools.Tool.SCOOP)
		KEY_H, KEY_F1:
			hud.toggle_help()
		KEY_BRACKETLEFT, KEY_BRACKETRIGHT:
			set_brush_radius(tools.radius * (0.8 if event.physical_keycode == KEY_BRACKETLEFT else 1.25))
		KEY_LEFT, KEY_RIGHT:
			set_scoop_tilt(tools.tilt + Vector2i(-45 if event.physical_keycode == KEY_LEFT else 45, 0))
		KEY_UP, KEY_DOWN:
			set_scoop_tilt(tools.tilt + Vector2i(0, 45 if event.physical_keycode == KEY_DOWN else -45))

func _make_dialog(mode: FileDialog.FileMode) -> FileDialog:
	var dialog := FileDialog.new()
	dialog.file_mode = mode
	dialog.access = FileDialog.ACCESS_FILESYSTEM
	dialog.filters = PackedStringArray(["*.octa ; Octa structure"])
	dialog.current_dir = ProjectSettings.globalize_path("res://structures")
	dialog.title = "Save arrangement" if mode == FileDialog.FILE_MODE_SAVE_FILE else "Load arrangement"
	dialog.theme = StudioTheme.create()
	add_child(dialog)
	return dialog

func _choose_save() -> void:
	tools.commit()
	_save_dialog.current_file = renderer.structure.title.to_snake_case() + ".octa"
	_save_dialog.popup_centered(Vector2i(850, 560))

func _choose_load() -> void:
	_load_dialog.popup_centered(Vector2i(850, 560))

func _save_structure(path: String) -> void:
	if path.get_extension().to_lower() != "octa":
		path += ".octa"
	if _files.save_structure(path, renderer.structure) == OK:
		hud.notify("Saved " + path.get_file())
		hud.toolbar.refresh_gallery()
	else:
		_show_error(_files.last_error)

func _load_structure(path: String) -> void:
	var structure := _files.load_structure(path)
	if structure == null:
		_show_error(_files.last_error)
		return
	_present(structure)
	hud.set_library_visible(false)
	hud.set_welcome_visible(false)
	hud.notify("Loaded " + path.get_file())

func _show_error(message: String) -> void:
	var dialog := AcceptDialog.new()
	dialog.title = "Structure file"
	dialog.dialog_text = message
	dialog.theme = StudioTheme.create()
	dialog.confirmed.connect(dialog.queue_free)
	dialog.canceled.connect(dialog.queue_free)
	add_child(dialog)
	dialog.popup_centered(Vector2i(480, 160))
