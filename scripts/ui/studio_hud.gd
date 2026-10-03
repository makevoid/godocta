class_name StudioHUD
extends CanvasLayer

signal reset_view_requested
signal orbit_changed(enabled: bool)
signal gap_changed(value: float)
signal xray_changed(enabled: bool)
var toolbar: StructureToolbar
var _title: Label
var _subtitle: Label
var _number: Label
var _count: Label
var _memory: Label
var _progress: ProgressBar
var _state: Label
var _time: Label
var _toast: Label
var _toast_remaining: float = 0.0
var _xray: CheckButton

func _ready() -> void:
	var root := Control.new()
	root.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	root.mouse_filter = Control.MOUSE_FILTER_IGNORE
	root.theme = StudioTheme.create()
	add_child(root)
	toolbar = StructureToolbar.new()
	root.add_child(toolbar)
	_build_info(root)
	_build_unit_card(root)
	_build_view_controls(root)
	_build_footer(root)
	_toast = StudioTheme.label("", 14, StudioTheme.ACCENT)
	_toast.set_anchors_and_offsets_preset(Control.PRESET_BOTTOM_RIGHT)
	_toast.offset_left = -550
	_toast.offset_top = -150
	_toast.offset_right = -32
	_toast.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
	root.add_child(_toast)

func _build_info(root: Control) -> void:
	var info := VBoxContainer.new()
	info.position = Vector2(34, 138)
	info.add_theme_constant_override("separation", 10)
	info.mouse_filter = Control.MOUSE_FILTER_IGNORE
	root.add_child(info)
	_number = StudioTheme.label("01 / PRIMITIVE", 12, StudioTheme.ACCENT)
	info.add_child(_number)
	_title = StudioTheme.label("Cube", 52)
	_title.custom_minimum_size.x = 260
	_title.text_overrun_behavior = TextServer.OVERRUN_TRIM_ELLIPSIS
	info.add_child(_title)
	_subtitle = StudioTheme.label("10 × 10 × 10 units", 16, StudioTheme.MUTED)
	info.add_child(_subtitle)
	var spacer := Control.new()
	spacer.custom_minimum_size.y = 16
	info.add_child(spacer)
	info.add_child(StudioTheme.label("A single cell. Endless arrangements.", 14, StudioTheme.MUTED))
	var line := HSeparator.new()
	line.custom_minimum_size = Vector2(254, 20)
	info.add_child(line)
	var stats := HBoxContainer.new()
	stats.add_theme_constant_override("separation", 32)
	info.add_child(stats)
	var cells := VBoxContainer.new()
	_count = StudioTheme.label("1,729", 28)
	cells.add_child(_count)
	cells.add_child(StudioTheme.label("CELLS", 10, StudioTheme.MUTED))
	stats.add_child(cells)
	var memory := VBoxContainer.new()
	_memory = StudioTheme.label("20.3 KiB", 28)
	memory.add_child(_memory)
	memory.add_child(StudioTheme.label("COORDINATE DATA", 10, StudioTheme.MUTED))
	stats.add_child(memory)

func _build_unit_card(root: Control) -> void:
	var panel := PanelContainer.new()
	panel.set_anchors_and_offsets_preset(Control.PRESET_BOTTOM_LEFT)
	panel.offset_left = 34
	panel.offset_right = 305
	panel.offset_top = -344
	panel.offset_bottom = -132
	panel.add_theme_stylebox_override("panel", StudioTheme.panel())
	root.add_child(panel)
	var column := VBoxContainer.new()
	column.add_theme_constant_override("separation", 8)
	panel.add_child(column)
	column.add_child(StudioTheme.label("THE BUILDING BLOCK", 11, StudioTheme.MUTED))
	var row := HBoxContainer.new()
	row.add_child(UnitGlyph.new())
	var name_box := VBoxContainer.new()
	name_box.alignment = BoxContainer.ALIGNMENT_CENTER
	name_box.add_child(StudioTheme.label("Truncated\noctahedron", 17))
	name_box.add_child(StudioTheme.label("SPACE-FILLING UNIT", 9, StudioTheme.ACCENT))
	row.add_child(name_box)
	column.add_child(row)
	column.add_child(HSeparator.new())
	column.add_child(StudioTheme.label("6 squares  +  8 hexagons", 13, StudioTheme.MUTED))

func _build_view_controls(root: Control) -> void:
	var box := VBoxContainer.new()
	box.set_anchors_and_offsets_preset(Control.PRESET_TOP_RIGHT)
	box.offset_left = -234
	box.offset_right = -34
	box.offset_top = 124
	box.add_theme_constant_override("separation", 12)
	root.add_child(box)
	var reset := StudioTheme.button("Reset view   F", 190)
	reset.pressed.connect(func() -> void: reset_view_requested.emit())
	box.add_child(reset)
	var orbit := CheckButton.new()
	orbit.text = "Auto orbit"
	orbit.toggled.connect(func(value: bool) -> void: orbit_changed.emit(value))
	box.add_child(orbit)
	_xray = CheckButton.new()
	_xray.text = "X-ray view   X"
	_xray.tooltip_text = "Semi-transparent cells with every edge and vertex visible."
	_xray.toggled.connect(func(value: bool) -> void: xray_changed.emit(value))
	box.add_child(_xray)
	box.add_child(StudioTheme.label("CELL SEPARATION", 10, StudioTheme.MUTED))
	var gap := HSlider.new()
	gap.min_value = 0
	gap.max_value = 0.2
	gap.step = 0.005
	gap.value = 0.045
	gap.tooltip_text = "Visual spacing only. Set to zero for exact face-to-face tiling."
	gap.value_changed.connect(func(value: float) -> void: gap_changed.emit(value))
	box.add_child(gap)

func _build_footer(root: Control) -> void:
	var footer := PanelContainer.new()
	footer.set_anchors_and_offsets_preset(Control.PRESET_BOTTOM_WIDE)
	footer.offset_top = -104
	footer.offset_left = 30
	footer.offset_right = -30
	footer.offset_bottom = -26
	footer.add_theme_stylebox_override("panel", StudioTheme.panel(Color("101a1e")))
	root.add_child(footer)
	var row := HBoxContainer.new()
	row.add_theme_constant_override("separation", 24)
	footer.add_child(row)
	var status := VBoxContainer.new()
	status.custom_minimum_size.x = 170
	status.add_child(StudioTheme.label("ASSEMBLY", 10, StudioTheme.MUTED))
	_state = StudioTheme.label("Assembling…", 17, StudioTheme.ACCENT)
	status.add_child(_state)
	row.add_child(status)
	var timeline := VBoxContainer.new()
	timeline.custom_minimum_size.x = 310
	timeline.alignment = BoxContainer.ALIGNMENT_CENTER
	_progress = ProgressBar.new()
	_progress.custom_minimum_size.y = 4
	_progress.max_value = 1.0
	_progress.show_percentage = false
	var track := StudioTheme.panel(Color("2b393a"), Color("2b393a"), 2)
	track.content_margin_top = 0
	track.content_margin_bottom = 0
	_progress.add_theme_stylebox_override("background", track)
	_progress.add_theme_stylebox_override("fill", StudioTheme.panel(StudioTheme.ACCENT, StudioTheme.ACCENT, 2))
	timeline.add_child(_progress)
	_time = StudioTheme.label("0.00 / 4.00 s   ·   CUBIC EASE OUT", 10, StudioTheme.MUTED)
	timeline.add_child(_time)
	row.add_child(timeline)
	var spacer := Control.new()
	spacer.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(spacer)
	var help := VBoxContainer.new()
	help.alignment = BoxContainer.ALIGNMENT_CENTER
	help.add_child(StudioTheme.label("Drag to orbit   ·   Scroll to zoom   ·   Shift + drag to pan", 13, StudioTheme.MUTED))
	help.add_child(StudioTheme.label("1–4  Primitives     5–9  Builds     X  X-ray     R  Replay     F  Reset view", 11, StudioTheme.MUTED))
	row.add_child(help)

func show_structure(structure: OctaStructure, preset: int = -1) -> void:
	_title.text = structure.title
	_number.text = "%02d / PRIMITIVE" % (preset + 1) if preset >= 0 else "FILE / STRUCTURE"
	var size := structure.bounds().size
	_subtitle.text = StructurePresets.DESCRIPTIONS[preset] if preset >= 0 \
		else "%s × %s × %s units" % [_format_count(roundi(size.x)), _format_count(roundi(size.y)), _format_count(roundi(size.z))]
	_count.text = _format_count(structure.cell_count())
	_memory.text = "%.1f KiB" % (structure.storage_bytes() / 1024.0)
	toolbar.select_shape(preset)

func show_progress(fraction: float, duration: float) -> void:
	_progress.value = fraction
	_state.text = "Assembling…" if fraction < 1.0 else "Assembly complete"
	_time.text = "%.2f / %.2f s   ·   CUBIC EASE OUT" % [fraction * duration, duration]

func set_xray(enabled: bool) -> void:
	_xray.set_pressed_no_signal(enabled)

func notify(message: String) -> void:
	_toast.text = message
	_toast_remaining = 5.0

func _process(delta: float) -> void:
	if _toast_remaining > 0.0:
		_toast_remaining -= delta
		if _toast_remaining <= 0.0:
			_toast.text = ""

func _format_count(value: int) -> String:
	var digits := str(value)
	var formatted := ""
	for index in digits.length():
		if index > 0 and (digits.length() - index) % 3 == 0:
			formatted += ","
		formatted += digits[index]
	return formatted
