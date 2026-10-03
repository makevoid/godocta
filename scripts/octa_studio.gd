class_name OctaStudio
extends Node3D
## Composition root. Geometry, storage, presets, animation, camera and UI are separate classes.

var renderer := StructureRenderer.new()
var camera := OrbitCamera.new()
var hud := StudioHUD.new()
var _files := OctaFile.new()
var _save_dialog: FileDialog
var _load_dialog: FileDialog

func _ready() -> void:
	add_child(StudioEnvironment.new())
	add_child(renderer)
	add_child(camera)
	add_child(hud)
	camera.current = true
	hud.toolbar.shape_selected.connect(select_preset)
	hud.toolbar.replay_requested.connect(renderer.replay)
	hud.toolbar.save_requested.connect(_choose_save)
	hud.toolbar.load_requested.connect(_choose_load)
	hud.reset_view_requested.connect(camera.reset_view)
	hud.orbit_changed.connect(func(enabled: bool) -> void: camera.auto_orbit = enabled)
	hud.gap_changed.connect(renderer.set_gap)
	renderer.assembly_progress.connect(hud.show_progress)
	_save_dialog = _make_dialog(FileDialog.FILE_MODE_SAVE_FILE)
	_save_dialog.file_selected.connect(_save_structure)
	_load_dialog = _make_dialog(FileDialog.FILE_MODE_OPEN_FILE)
	_load_dialog.file_selected.connect(_load_structure)
	select_preset(StructurePresets.Shape.CUBE)

func select_preset(shape: StructurePresets.Shape) -> void:
	var structure := StructurePresets.generate(shape)
	hud.show_structure(structure, shape)
	renderer.display(structure)
	camera.frame_structure(structure.bounds())

func _unhandled_key_input(event: InputEvent) -> void:
	if not event is InputEventKey or not event.pressed or event.echo:
		return
	# Keep shortcuts out of modal file dialogs.
	if _save_dialog.visible or _load_dialog.visible:
		return
	match event.physical_keycode:
		KEY_1, KEY_2, KEY_3, KEY_4:
			select_preset((event.physical_keycode - KEY_1) as StructurePresets.Shape)
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
	else:
		_show_error(_files.last_error)

func _load_structure(path: String) -> void:
	var structure := _files.load_structure(path)
	if structure == null:
		_show_error(_files.last_error)
		return
	hud.show_structure(structure)
	renderer.display(structure)
	camera.frame_structure(structure.bounds())
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
