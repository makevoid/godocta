class_name StructureToolbar
extends PanelContainer

signal shape_selected(shape: StructurePresets.Shape)
signal replay_requested
signal save_requested
signal load_requested
var shape_buttons: Array[Button] = []

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
		var button := StudioTheme.button(title, 136)
		button.icon = icons[index]
		button.toggle_mode = true
		button.button_group = group
		button.tooltip_text = "Assemble %s in 4 seconds · shortcut %d" % [StructurePresets.TITLES[index].to_lower(), index + 1]
		button.pressed.connect(func() -> void: shape_selected.emit(index as StructurePresets.Shape))
		shape_buttons.append(button)
		row.add_child(button)
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

func select_shape(index: int) -> void:
	for i in shape_buttons.size():
		shape_buttons[i].set_pressed_no_signal(i == index)
