class_name ToolPalette
extends PanelContainer
## Editing tools: orbit, brush cells in, or swallow them with a black hole or a tilted scoop.
## The brush color submenu lists auto plus the build's palette.
## Emits intent only; OctaStudio applies it to StructureTools and calls the show_* setters.

signal tool_selected(tool: StructureTools.Tool)
signal radius_changed(cells: float)
signal tilt_changed(tilt: Vector2i)
## A palette index, or -1 for auto: new cells adopt their neighbors' color.
signal color_selected(index: int)
const SHORTCUTS: Array[String] = ["V", "B", "E", "S"]
const TIPS: Array[String] = [
	"Left drag orbits the camera.",
	"Paint new cells: they are attracted to the cells under the brush, or rest on the ground.",
	"A globe that swallows every cell within its radius.",
	"A half globe that shaves the surface without digging deep. Tilt it 45° with the arrow keys.",
]
const TILTS: Array[int] = [-45, 0, 45]
const AUTO_COLOR_NAME := "Auto · neighbors"
const SWATCH := 18 # Side of a color swatch, in pixels.
var tool_buttons: Array[Button] = []
var color_menu: MenuButton
var _palette: PackedColorArray = OctaPalette.colors()
var _radius: HSlider
var _radius_label: Label
var _tilt_buttons: Array = [[], []] # [axis][step] → Button
var _tilt_box: VBoxContainer
var _color_box: VBoxContainer

func _ready() -> void:
	add_theme_stylebox_override("panel", StudioTheme.panel(Color(0.04, 0.06, 0.07, 0.72), Color(1, 1, 1, 0.06), 14))
	var box := VBoxContainer.new()
	box.add_theme_constant_override("separation", 8)
	add_child(box)
	box.add_child(StudioTheme.label("TOOLS", 10, StudioTheme.MUTED))
	var group := ButtonGroup.new()
	for index in StructureTools.NAMES.size():
		var button := StudioTheme.button("%s   %s" % [StructureTools.NAMES[index], SHORTCUTS[index]], 190)
		button.custom_minimum_size.y = 38
		button.toggle_mode = true
		button.button_group = group
		button.focus_mode = Control.FOCUS_NONE
		button.alignment = HORIZONTAL_ALIGNMENT_LEFT
		button.tooltip_text = TIPS[index]
		button.pressed.connect(func() -> void: tool_selected.emit(index as StructureTools.Tool))
		tool_buttons.append(button)
		box.add_child(button)
	tool_buttons[0].set_pressed_no_signal(true)
	var size_row := HBoxContainer.new()
	size_row.add_child(StudioTheme.label("BRUSH SIZE   [ ]", 10, StudioTheme.MUTED))
	_radius_label = StudioTheme.label("", 12, StudioTheme.ACCENT)
	_radius_label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_radius_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
	size_row.add_child(_radius_label)
	box.add_child(size_row)
	_radius = HSlider.new()
	_radius.min_value = StructureTools.MIN_RADIUS
	_radius.max_value = StructureTools.MAX_RADIUS
	_radius.step = 0.1
	_radius.focus_mode = Control.FOCUS_NONE
	_radius.tooltip_text = "Brush radius in cells."
	_radius.value_changed.connect(func(value: float) -> void:
		_radius_label.text = "%.1f cells" % value
		radius_changed.emit(value))
	box.add_child(_radius)
	var color_box := VBoxContainer.new()
	color_box.add_theme_constant_override("separation", 6)
	color_box.add_child(StudioTheme.label("BRUSH COLOR", 10, StudioTheme.MUTED))
	color_menu = MenuButton.new()
	color_menu.custom_minimum_size = Vector2(190, 34)
	color_menu.alignment = HORIZONTAL_ALIGNMENT_LEFT
	color_menu.flat = false
	color_menu.focus_mode = Control.FOCUS_NONE
	color_menu.mouse_default_cursor_shape = Control.CURSOR_POINTING_HAND
	color_menu.tooltip_text = "Color the brush paints. Auto copies the cells it grows from."
	var colors := color_menu.get_popup()
	colors.add_icon_radio_check_item(_swatch(StudioTheme.MUTED), AUTO_COLOR_NAME, 0)
	for index in OctaPalette.NAMES.size():
		colors.add_icon_radio_check_item(_swatch(OctaPalette.color_of(OctaPalette.HEX[index])), OctaPalette.NAMES[index], index + 1)
	# Item ids are the palette index plus one, so auto can keep the id 0 that popups need.
	colors.id_pressed.connect(func(id: int) -> void: color_selected.emit(id - 1))
	color_box.add_child(color_menu)
	box.add_child(color_box)
	_color_box = color_box
	_tilt_box = VBoxContainer.new()
	_tilt_box.add_theme_constant_override("separation", 6)
	_tilt_box.add_child(StudioTheme.label("SCOOP TILT   ← → ↑ ↓", 10, StudioTheme.MUTED))
	for axis in 2:
		var row := HBoxContainer.new()
		row.add_theme_constant_override("separation", 4)
		var axis_label := StudioTheme.label("X" if axis == 0 else "Y", 12, StudioTheme.MUTED)
		axis_label.custom_minimum_size.x = 22
		row.add_child(axis_label)
		var axis_group := ButtonGroup.new()
		for step in TILTS.size():
			var button := StudioTheme.button("%+d°" % TILTS[step] if TILTS[step] != 0 else "Flat", 54)
			button.custom_minimum_size.y = 30
			button.add_theme_font_size_override("font_size", 12)
			button.toggle_mode = true
			button.button_group = axis_group
			button.focus_mode = Control.FOCUS_NONE
			button.tooltip_text = ("Tip the scoop towards screen right (+) or left (−)." if axis == 0
				else "Tip the scoop towards the camera (+) or away (−).")
			button.pressed.connect(func() -> void: tilt_changed.emit(_selected_tilt()))
			_tilt_buttons[axis].append(button)
			row.add_child(button)
		_tilt_box.add_child(row)
	box.add_child(_tilt_box)
	show_radius(2.0)
	show_tilt(Vector2i.ZERO)
	show_color(-1)
	show_tool(StructureTools.Tool.NAVIGATE)

func show_tool(tool: StructureTools.Tool) -> void:
	for index in tool_buttons.size():
		tool_buttons[index].set_pressed_no_signal(index == tool)
	_tilt_box.modulate.a = 1.0 if tool == StructureTools.Tool.SCOOP else 0.45
	_color_box.modulate.a = 1.0 if tool == StructureTools.Tool.BRUSH_IN else 0.45

func show_radius(cells: float) -> void:
	_radius.set_value_no_signal(cells)
	_radius_label.text = "%.1f cells" % cells

## The colors the submenu offers: the displayed build's palette. Uncolored builds leave
## only auto selectable, because their cells carry no palette index.
func show_palette(palette: PackedColorArray) -> void:
	_palette = palette if not palette.is_empty() else OctaPalette.colors()
	var colors := color_menu.get_popup()
	for index in OctaPalette.NAMES.size():
		var known := index < palette.size()
		colors.set_item_icon(index + 1, _swatch(_palette[index] if index < _palette.size() else StudioTheme.MUTED))
		colors.set_item_disabled(index + 1, not known)
	color_menu.disabled = palette.is_empty()

## Show the selected color: a palette index, or -1 for auto.
func show_color(index: int) -> void:
	var colors := color_menu.get_popup()
	for item in colors.item_count:
		colors.set_item_checked(item, item == index + 1)
	color_menu.text = OctaPalette.name_of(index) if index >= 0 else AUTO_COLOR_NAME
	color_menu.icon = _swatch(_palette[index] if index >= 0 and index < _palette.size() else StudioTheme.MUTED)

## A flat square of one color, for the menu button and its items.
static func _swatch(color: Color) -> ImageTexture:
	var image := Image.create(SWATCH, SWATCH, false, Image.FORMAT_RGBA8)
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
