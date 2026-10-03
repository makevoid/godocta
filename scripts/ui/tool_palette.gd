class_name ToolPalette
extends PanelContainer
## Editing tools: orbit, brush cells in, or swallow them with a black hole or a tilted scoop.
## Emits intent only; OctaStudio applies it to StructureTools and calls the show_* setters.

signal tool_selected(tool: StructureTools.Tool)
signal radius_changed(cells: float)
signal tilt_changed(tilt: Vector2i)
const SHORTCUTS: Array[String] = ["V", "B", "E", "S"]
const TIPS: Array[String] = [
	"Left drag orbits the camera.",
	"Paint new cells: they are attracted to the cells under the brush, or rest on the ground.",
	"A globe that swallows every cell within its radius.",
	"A half globe that shaves the surface without digging deep. Tilt it 45° with the arrow keys.",
]
const TILTS: Array[int] = [-45, 0, 45]
var tool_buttons: Array[Button] = []
var _radius: HSlider
var _radius_label: Label
var _tilt_buttons: Array = [[], []] # [axis][step] → Button
var _tilt_box: VBoxContainer

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
	show_tool(StructureTools.Tool.NAVIGATE)

func show_tool(tool: StructureTools.Tool) -> void:
	for index in tool_buttons.size():
		tool_buttons[index].set_pressed_no_signal(index == tool)
	_tilt_box.modulate.a = 1.0 if tool == StructureTools.Tool.SCOOP else 0.45

func show_radius(cells: float) -> void:
	_radius.set_value_no_signal(cells)
	_radius_label.text = "%.1f cells" % cells

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
