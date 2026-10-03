class_name StructureToolbar
extends PanelContainer

signal shape_selected(shape: StructurePresets.Shape)
signal file_selected(path: String)
signal replay_requested
signal save_requested
signal load_requested
const STRUCTURES_DIR := "res://structures"
var shape_buttons: Array[Button] = []
var gallery: MenuButton
var gallery_paths: PackedStringArray = []

func _ready() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_TOP_WIDE)
	offset_bottom = 88
	var background := StudioTheme.panel(Color("10191d"), StudioTheme.LINE, 0)
	background.content_margin_left = 30
	background.content_margin_right = 30
	background.content_margin_top = 20
	background.content_margin_bottom = 20
	add_theme_stylebox_override("panel", background)
	var row := HBoxContainer.new()
	row.add_theme_constant_override("separation", 12)
	add_child(row)
	var logo := TextureRect.new()
	logo.texture = preload("res://assets/octa.svg")
	logo.custom_minimum_size = Vector2(37, 37)
	logo.expand_mode = TextureRect.EXPAND_IGNORE_SIZE
	logo.stretch_mode = TextureRect.STRETCH_KEEP_ASPECT_CENTERED
	row.add_child(logo)
	var identity := VBoxContainer.new()
	identity.custom_minimum_size.x = 162
	identity.add_theme_constant_override("separation", 0)
	identity.add_child(StudioTheme.label("O C T A", 23))
	identity.add_child(StudioTheme.label("STRUCTURE STUDIO", 10, StudioTheme.MUTED))
	row.add_child(identity)
	var separator := VSeparator.new()
	row.add_child(separator)
	var icons: Array[Texture2D] = [preload("res://assets/cube.svg"), preload("res://assets/sphere.svg"),
		preload("res://assets/pyramid.svg"), preload("res://assets/cylinder.svg")]
	var group := ButtonGroup.new()
	for index in 4:
		var title := "Cube 10³" if index == 0 else StructurePresets.TITLES[index]
		var button := StudioTheme.button(title, 124)
		button.icon = icons[index]
		button.toggle_mode = true
		button.button_group = group
		button.tooltip_text = "Assemble %s · shortcut %d" % [StructurePresets.TITLES[index].to_lower(), index + 1]
		button.pressed.connect(func() -> void: shape_selected.emit(index as StructurePresets.Shape))
		shape_buttons.append(button)
		row.add_child(button)
	_build_gallery(row)
	var spacer := Control.new()
	spacer.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(spacer)
	var load_button := StudioTheme.button("Load", 70)
	load_button.tooltip_text = "Open an .octa structure"
	load_button.pressed.connect(func() -> void: load_requested.emit())
	row.add_child(load_button)
	var save_button := StudioTheme.button("Save", 70)
	save_button.tooltip_text = "Save the arrangement as a compact .octa file"
	save_button.pressed.connect(func() -> void: save_requested.emit())
	row.add_child(save_button)
	var replay := StudioTheme.button("Replay   R", 115)
	replay.add_theme_color_override("font_color", StudioTheme.ACCENT)
	replay.pressed.connect(func() -> void: replay_requested.emit())
	row.add_child(replay)

## Every .octa file in structures/: generated builds (castle, arena, …) and saved ones.
func _build_gallery(row: HBoxContainer) -> void:
	gallery = MenuButton.new()
	gallery.text = "Builds  ▾"
	gallery.icon = preload("res://assets/castle.svg")
	gallery.custom_minimum_size = Vector2(124, 44)
	gallery.flat = false # MenuButton inherits the theme's Button styles once unflattened.
	gallery.tooltip_text = "Large structures from structures/*.octa · shortcuts 5–9"
	gallery.mouse_default_cursor_shape = Control.CURSOR_POINTING_HAND
	gallery.get_popup().id_pressed.connect(func(id: int) -> void: file_selected.emit(gallery_paths[id]))
	row.add_child(gallery)
	refresh_gallery()

func refresh_gallery() -> void:
	var files := Array(DirAccess.get_files_at(STRUCTURES_DIR)).filter(
		func(name: String) -> bool: return name.get_extension().to_lower() == "octa")
	files.sort()
	gallery_paths.clear()
	var popup := gallery.get_popup()
	popup.clear()
	for name: String in files:
		gallery_paths.append(STRUCTURES_DIR.path_join(name))
		var shortcut := "    %d" % (gallery_paths.size() + 4) if gallery_paths.size() <= 5 else ""
		popup.add_item(name.get_basename().capitalize() + shortcut, gallery_paths.size() - 1)

func select_shape(index: int) -> void:
	for i in shape_buttons.size():
		shape_buttons[i].set_pressed_no_signal(i == index)
