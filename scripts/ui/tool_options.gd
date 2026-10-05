class_name ToolOptions
extends PanelContainer
## Options for the selected tool, in the bottom right corner. Orbit shows nothing;
## every brush shows its size; brush in adds the paint color, and the scoop its tilt.
## The color grid offers the first eight entries of the displayed build's palette and
## reveals the remaining eight on demand. Emits intent only: `OctaStudio` applies it
## to `StructureTools` and calls the `show_*` setters back.

signal radius_changed(cells: float)
signal tilt_changed(tilt: Vector2i)
## A palette index, or -1 for auto: new cells adopt their neighbors' color.
signal color_selected(index: int)
const TILTS: Array[int] = [-45, 0, 45]
const AUTO_COLOR_NAME := "Auto · neighbors"
const SWATCH := Vector2i(20, 12) # Size of a color swatch, in pixels.
const COLUMNS := 4
const VISIBLE_COLORS := 8 # Entries shown before "More" reveals the whole palette.
## Index 0..15 of the palette, parallel to `OctaPalette.NAMES`.
var color_buttons: Array[Button] = []
var auto_button: Button
var more_button: Button
var more_expanded: bool = false
var _palette: PackedColorArray = OctaPalette.colors()
var _has_palette: bool = false
var _tool: StructureTools.Tool = StructureTools.Tool.NAVIGATE
var _radius: HSlider
var _radius_label: Label
var _tilt_buttons: Array = [[], []] # [axis][step] → Button
var _size_box: VBoxContainer
var _color_box: VBoxContainer
var _tilt_box: VBoxContainer
var _column: VBoxContainer

func _ready() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_BOTTOM_RIGHT)
	offset_left = -218
	offset_right = -16
	# Zero height plus an upward growth direction: the panel takes its height from
	# the sections currently shown and keeps its bottom right corner pinned.
	offset_top = -16
	offset_bottom = -16
	grow_vertical = Control.GROW_DIRECTION_BEGIN
	add_theme_stylebox_override("panel", StudioTheme.card())
	_column = VBoxContainer.new()
	_column.add_theme_constant_override("separation", 10)
	add_child(_column)
	_build_size()
	_build_colors()
	_build_tilt()
	show_radius(float(StructureTools.DEFAULT_RADII[StructureTools.Tool.BRUSH_IN]))
	show_tilt(Vector2i.ZERO)
	show_color(-1)
	show_tool(StructureTools.Tool.NAVIGATE)

func _build_size() -> void:
	_size_box = VBoxContainer.new()
	_size_box.add_theme_constant_override("separation", 4)
	var row := HBoxContainer.new()
	row.add_child(StudioTheme.label("BRUSH SIZE   [ ]", 9, StudioTheme.MUTED))
	_radius_label = StudioTheme.label("", 11, StudioTheme.ACCENT)
	_radius_label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_radius_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
	row.add_child(_radius_label)
	_size_box.add_child(row)
	_radius = HSlider.new()
	_radius.min_value = StructureTools.MIN_RADIUS
	_radius.max_value = StructureTools.MAX_RADIUS
	_radius.step = 0.1
	_radius.focus_mode = Control.FOCUS_NONE
	_radius.tooltip_text = "Brush radius in cells."
	_radius.value_changed.connect(func(value: float) -> void:
		_radius_label.text = "%.1f cells" % value
		radius_changed.emit(value))
	_size_box.add_child(_radius)
	_column.add_child(_size_box)

func _build_colors() -> void:
	_color_box = VBoxContainer.new()
	_color_box.add_theme_constant_override("separation", 5)
	_color_box.add_child(StudioTheme.label("BRUSH COLOR", 9, StudioTheme.MUTED))
	var group := ButtonGroup.new()
	auto_button = StudioTheme.button(AUTO_COLOR_NAME, 0, 26)
	auto_button.icon = _swatch(StudioTheme.MUTED)
	auto_button.alignment = HORIZONTAL_ALIGNMENT_LEFT
	auto_button.add_theme_font_size_override("font_size", 11)
	auto_button.toggle_mode = true
	auto_button.button_group = group
	auto_button.tooltip_text = "New cells copy the color of the cells they grow from."
	auto_button.pressed.connect(func() -> void: color_selected.emit(-1))
	_color_box.add_child(auto_button)
	var grid := GridContainer.new()
	grid.columns = COLUMNS
	grid.add_theme_constant_override("h_separation", 4)
	grid.add_theme_constant_override("v_separation", 4)
	for index in OctaPalette.NAMES.size():
		var button := StudioTheme.button("", 42, 24)
		button.icon = _swatch(OctaPalette.color_of(OctaPalette.HEX[index]))
		button.toggle_mode = true
		button.button_group = group
		button.tooltip_text = OctaPalette.NAMES[index]
		button.pressed.connect(func() -> void: color_selected.emit(index))
		color_buttons.append(button)
		grid.add_child(button)
	_color_box.add_child(grid)
	more_button = StudioTheme.button("More  ▾", 0, 22)
	more_button.add_theme_font_size_override("font_size", 10)
	more_button.pressed.connect(func() -> void: set_more_expanded(not more_expanded))
	_color_box.add_child(more_button)
	_column.add_child(_color_box)
	set_more_expanded(false)

func _build_tilt() -> void:
	_tilt_box = VBoxContainer.new()
	_tilt_box.add_theme_constant_override("separation", 4)
	_tilt_box.add_child(StudioTheme.label("SCOOP TILT   ← → ↑ ↓", 9, StudioTheme.MUTED))
	for axis in 2:
		var row := HBoxContainer.new()
		row.add_theme_constant_override("separation", 4)
		var axis_label := StudioTheme.label("X" if axis == 0 else "Y", 11, StudioTheme.MUTED)
		axis_label.custom_minimum_size.x = 14
		row.add_child(axis_label)
		var axis_group := ButtonGroup.new()
		for step in TILTS.size():
			var button := StudioTheme.button("%+d°" % TILTS[step] if TILTS[step] != 0 else "Flat", 52, 24)
			button.add_theme_font_size_override("font_size", 11)
			button.toggle_mode = true
			button.button_group = axis_group
			button.tooltip_text = ("Tip the scoop towards screen right (+) or left (−)." if axis == 0
				else "Tip the scoop towards the camera (+) or away (−).")
			button.pressed.connect(func() -> void: tilt_changed.emit(_selected_tilt()))
			_tilt_buttons[axis].append(button)
			row.add_child(button)
		_tilt_box.add_child(row)
	_column.add_child(_tilt_box)

## Only the sections the selected tool uses stay on screen; orbit hides the panel.
func show_tool(tool: StructureTools.Tool) -> void:
	_tool = tool
	visible = tool != StructureTools.Tool.NAVIGATE
	_size_box.visible = visible
	_color_box.visible = _has_palette and tool == StructureTools.Tool.BRUSH_IN
	_tilt_box.visible = tool == StructureTools.Tool.SCOOP

func show_radius(cells: float) -> void:
	_radius.set_value_no_signal(cells)
	_radius_label.text = "%.1f cells" % cells

## The colors the grid offers: the displayed build's palette. Uncolored builds carry no
## palette index, so their cells can only be painted in auto and the grid stays hidden.
func show_palette(palette: PackedColorArray) -> void:
	_palette = palette if not palette.is_empty() else OctaPalette.colors()
	_has_palette = not palette.is_empty()
	for index in color_buttons.size():
		var known := index < palette.size()
		color_buttons[index].icon = _swatch(_palette[index] if index < _palette.size() else StudioTheme.MUTED)
		color_buttons[index].disabled = not known
	show_tool(_tool)

## Show the selected color: a palette index, or -1 for auto.
func show_color(index: int) -> void:
	auto_button.set_pressed_no_signal(index < 0)
	for item in color_buttons.size():
		color_buttons[item].set_pressed_no_signal(item == index)
	if index >= VISIBLE_COLORS:
		set_more_expanded(true)

## Reveal the whole palette, or fold it back to the first eight entries.
func set_more_expanded(expanded: bool) -> void:
	more_expanded = expanded
	for index in color_buttons.size():
		color_buttons[index].visible = expanded or index < VISIBLE_COLORS
	more_button.text = "Fewer  ▴" if expanded else "More  ▾"

## A flat chip of one color, for the swatch buttons.
static func _swatch(color: Color) -> ImageTexture:
	var image := Image.create(SWATCH.x, SWATCH.y, false, Image.FORMAT_RGBA8)
	image.fill(color)
	return ImageTexture.create_from_image(image)

func show_tilt(tilt: Vector2i) -> void:
	for axis in 2:
		for step in TILTS.size():
			(_tilt_buttons[axis][step] as Button).set_pressed_no_signal(TILTS[step] == (tilt.x if axis == 0 else tilt.y))

func _selected_tilt() -> Vector2i:
	var result := Vector2i.ZERO
	for axis in 2:
		for step in TILTS.size():
			if (_tilt_buttons[axis][step] as Button).button_pressed:
				result[axis] = TILTS[step]
	return result
