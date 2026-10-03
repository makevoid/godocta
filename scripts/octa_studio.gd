class_name OctaStudio
extends Node3D
## Composition root. Geometry, storage, presets, animation, camera and UI are separate classes.

var renderer := StructureRenderer.new()
var camera := OrbitCamera.new()
var hud := StudioHUD.new()
var environment: StudioEnvironment
var _files := OctaFile.new()
var _save_dialog: FileDialog
var _load_dialog: FileDialog

func _ready() -> void:
	environment = StudioEnvironment.new()
	add_child(environment)
	add_child(renderer)
	add_child(camera)
	add_child(hud)
	camera.current = true
	hud.toolbar.shape_selected.connect(select_preset)
	hud.toolbar.file_selected.connect(_load_structure)
	hud.toolbar.replay_requested.connect(renderer.replay)
	hud.toolbar.save_requested.connect(_choose_save)
	hud.toolbar.load_requested.connect(_choose_load)
	hud.reset_view_requested.connect(camera.reset_view)
	hud.orbit_changed.connect(func(enabled: bool) -> void: camera.auto_orbit = enabled)
	hud.gap_changed.connect(renderer.set_gap)
	hud.xray_changed.connect(renderer.set_xray)
	renderer.assembly_progress.connect(hud.show_progress)
	_save_dialog = _make_dialog(FileDialog.FILE_MODE_SAVE_FILE)
	_save_dialog.file_selected.connect(_save_structure)
	_load_dialog = _make_dialog(FileDialog.FILE_MODE_OPEN_FILE)
	_load_dialog.file_selected.connect(_load_structure)
	_fit_window_to_screen()
	select_preset(StructurePresets.Shape.CUBE)

func select_preset(shape: StructurePresets.Shape) -> void:
	var structure := StructurePresets.generate(shape)
	_present(structure, shape)

func _present(structure: OctaStructure, preset: int = -1) -> void:
	hud.show_structure(structure, preset)
	renderer.display(structure)
	var box := structure.bounds()
	environment.apply_theme(WorldTheme.STUDIO if preset >= 0 else WorldTheme.name_for(structure.title))
	camera.default_pitch = 0.48 if environment.theme_name == WorldTheme.STUDIO else 0.24
	camera.frame_structure(box)
	environment.fit(box)

## The default 16:10 window is 1760 × 1100; shrink it to fit smaller displays unless an
## explicit --resolution was requested.
func _fit_window_to_screen() -> void:
	var window := get_window()
	if window.mode != Window.MODE_WINDOWED or OS.get_cmdline_args().has("--resolution"):
		return
	var usable := DisplayServer.screen_get_usable_rect(window.current_screen)
	if usable.size.x <= 0 or usable.size.y <= 0:
		return
	var scale := minf(usable.size.x * 0.94 / window.size.x, usable.size.y * 0.94 / window.size.y)
	if scale < 1.0:
		window.size = Vector2i(Vector2(window.size) * scale)
		window.position = usable.position + (usable.size - window.size) / 2

func toggle_xray() -> void:
	renderer.set_xray(not renderer.xray)
	hud.set_xray(renderer.xray)

func _unhandled_key_input(event: InputEvent) -> void:
	if not event is InputEventKey or not event.pressed or event.echo:
		return
	# Keep shortcuts out of modal file dialogs.
	if _save_dialog.visible or _load_dialog.visible:
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
			renderer.replay()
		KEY_F:
			camera.reset_view()

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
