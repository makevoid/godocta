class_name StudioHUD
extends CanvasLayer
## Chrome over the 3D view: the top bar, the structure readout inside it, the assembly
## line under it, the contextual tool options, toasts and the help card. Everything the
## user can act on lives in `StructureToolbar`, `ToolPalette` or `ToolOptions`; this
## class only presents state and forwards their signals' owners to `OctaStudio`.

var toolbar: StructureToolbar
## The tool strip inside the top bar, and the options panel for the selected tool.
var tools: ToolPalette
var options: ToolOptions
var help: PanelContainer
var _title: Label
var _count: Label
var _memory: Label
var _progress: ProgressBar
var _state: Label
var _time: Label
var _clock: HBoxContainer
var _toast: Label
var _toast_remaining: float = 0.0
var _fps: Label
var _fps_elapsed: float = 0.0

## Rows of the help card: shortcut and what it does.
const HELP: Array[Array] = [
	["Drag", "Orbit the view (right drag or Alt + drag while a tool is active)"],
	["Wheel", "Zoom · Shift + drag or middle drag pans"],
	["1 – 4", "Primitives: cube, sphere, pyramid, cylinder"],
	["5 – 9", "The first five builds in the Builds menu"],
	["V / Esc", "Orbit only — no editing"],
	["B", "Brush in: paint cells onto the build in the brush color"],
	["E", "Black hole: swallow every cell inside the globe"],
	["S", "Scoop: shave the surface (enable it in Settings first)"],
	["[ ]", "Shrink or grow the brush, 0.6 – 12 cells"],
	["← → ↑ ↓", "Tilt the scoop 45° around the screen axes"],
	["R / Space", "Replay the assembly"],
	["F", "Reset the framing"],
	["X", "X-ray view, also in Settings"],
	["H", "This card"],
]

func _ready() -> void:
	var root := Control.new()
	root.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	root.mouse_filter = Control.MOUSE_FILTER_IGNORE
	root.theme = StudioTheme.create()
	add_child(root)
	toolbar = StructureToolbar.new()
	root.add_child(toolbar)
	tools = toolbar.tools
	_build_info()
	_build_assembly(root)
	options = ToolOptions.new()
	root.add_child(options)
	_build_help(root)
	_toast = StudioTheme.label("", 12, StudioTheme.ACCENT)
	_toast.set_anchors_and_offsets_preset(Control.PRESET_BOTTOM_LEFT)
	_toast.offset_left = 18
	_toast.offset_top = -44
	_toast.offset_right = 620
	root.add_child(_toast)

## The structure readout, at the right of the top bar.
func _build_info() -> void:
	_title = StudioTheme.label("Cube", 15)
	_title.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
	toolbar.info_slot.add_child(_title)
	var row := HBoxContainer.new()
	row.add_theme_constant_override("separation", 6)
	row.alignment = BoxContainer.ALIGNMENT_END
	row.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_count = StudioTheme.label("1,729", 10, StudioTheme.ACCENT)
	row.add_child(_count)
	row.add_child(StudioTheme.label("CELLS  ·", 9, StudioTheme.MUTED))
	_memory = StudioTheme.label("20.3 KiB", 10, StudioTheme.MUTED)
	row.add_child(_memory)
	row.add_child(StudioTheme.label("·", 9, StudioTheme.MUTED))
	_fps = StudioTheme.label("— FPS", 10, StudioTheme.MUTED)
	_fps.custom_minimum_size.x = 92
	row.add_child(_fps)
	toolbar.info_slot.add_child(row)

## A hairline progress bar under the top bar, with its clock, while cells still fly in.
func _build_assembly(root: Control) -> void:
	_progress = ProgressBar.new()
	_progress.set_anchors_and_offsets_preset(Control.PRESET_TOP_WIDE)
	_progress.offset_top = StructureToolbar.HEIGHT
	_progress.offset_bottom = StructureToolbar.HEIGHT + 3
	_progress.custom_minimum_size.y = 3
	_progress.max_value = 1.0
	_progress.show_percentage = false
	var track := StudioTheme.panel(Color("1b2a2c"), Color("1b2a2c"), 0, 0, 0)
	_progress.add_theme_stylebox_override("background", track)
	_progress.add_theme_stylebox_override("fill", StudioTheme.panel(StudioTheme.ACCENT, StudioTheme.ACCENT, 0, 0, 0))
	root.add_child(_progress)
	_clock = HBoxContainer.new()
	var clock := _clock
	clock.set_anchors_and_offsets_preset(Control.PRESET_TOP_LEFT)
	clock.offset_left = 18
	clock.offset_top = StructureToolbar.HEIGHT + 7
	clock.add_theme_constant_override("separation", 8)
	clock.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_state = StudioTheme.label("Assembling…", 10, StudioTheme.ACCENT)
	clock.add_child(_state)
	_time = StudioTheme.label("0.00 / 4.00 s", 9, StudioTheme.MUTED)
	clock.add_child(_time)
	root.add_child(clock)
	_progress.visible = false
	clock.visible = false

func _build_help(root: Control) -> void:
	help = PanelContainer.new()
	help.set_anchors_and_offsets_preset(Control.PRESET_CENTER)
	help.grow_horizontal = Control.GROW_DIRECTION_BOTH
	help.grow_vertical = Control.GROW_DIRECTION_BOTH
	help.add_theme_stylebox_override("panel", StudioTheme.panel(Color(0.03, 0.05, 0.06, 0.95), StudioTheme.LINE, 12, 22, 18))
	help.visible = false
	root.add_child(help)
	var column := VBoxContainer.new()
	column.add_theme_constant_override("separation", 12)
	help.add_child(column)
	var header := HBoxContainer.new()
	header.add_theme_constant_override("separation", 12)
	header.add_child(StudioTheme.label("How to use Octa", 20))
	var pad := Control.new()
	pad.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	header.add_child(pad)
	var close := StudioTheme.button("Close", 60)
	close.pressed.connect(func() -> void: set_help_visible(false))
	header.add_child(close)
	column.add_child(header)
	var body := HBoxContainer.new()
	body.add_theme_constant_override("separation", 26)
	var keys := GridContainer.new()
	keys.columns = 2
	keys.add_theme_constant_override("h_separation", 16)
	keys.add_theme_constant_override("v_separation", 5)
	for entry in HELP:
		var key := StudioTheme.label(entry[0], 11, StudioTheme.ACCENT)
		key.custom_minimum_size.x = 76
		key.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
		keys.add_child(key)
		keys.add_child(StudioTheme.label(entry[1], 12, StudioTheme.MUTED))
	body.add_child(keys)
	# The same geometry the renderer draws, as a reminder of what a cell is.
	var unit := VBoxContainer.new()
	unit.alignment = BoxContainer.ALIGNMENT_CENTER
	unit.add_theme_constant_override("separation", 4)
	unit.add_child(StudioTheme.label("THE BUILDING BLOCK", 9, StudioTheme.MUTED))
	var glyph := UnitGlyph.new()
	glyph.size_flags_horizontal = Control.SIZE_SHRINK_CENTER
	unit.add_child(glyph)
	unit.add_child(StudioTheme.label("Truncated octahedron", 12))
	unit.add_child(StudioTheme.label("6 squares  +  8 hexagons", 10, StudioTheme.MUTED))
	body.add_child(unit)
	column.add_child(body)
	column.add_child(StudioTheme.label("Builds load from structures/*.octa · Save writes the same compact format.", 11, StudioTheme.MUTED))

func toggle_help() -> void:
	set_help_visible(not help.visible)

func set_help_visible(shown: bool) -> void:
	help.visible = shown

func show_structure(structure: OctaStructure, preset: int = -1) -> void:
	_title.text = structure.title
	_count.text = _format_count(structure.cell_count())
	_memory.text = "%.1f KiB" % (structure.storage_bytes() / 1024.0)
	toolbar.select_shape(preset)

## Refresh the counters after an edit, keeping the title.
func show_counts(structure: OctaStructure) -> void:
	_count.text = _format_count(structure.cell_count())
	_memory.text = "%.1f KiB" % (structure.storage_bytes() / 1024.0)

func show_progress(fraction: float, duration: float) -> void:
	_progress.value = fraction
	_state.text = "Assembling…" if fraction < 1.0 else "Assembly complete"
	_time.text = "%.2f / %.2f s" % [fraction * duration, duration]
	var running := fraction < 1.0
	_progress.visible = running
	_clock.visible = running

func set_xray(enabled: bool) -> void:
	toolbar.set_setting_checked(StructureToolbar.Setting.XRAY, enabled)

func notify(message: String) -> void:
	_toast.text = message
	_toast_remaining = 5.0

func _process(delta: float) -> void:
	_fps_elapsed += delta
	if _fps_elapsed >= 0.5:
		_fps_elapsed = 0.0
		var fps := Engine.get_frames_per_second()
		_fps.text = "%d FPS · %.1f ms" % [fps, 1000.0 / maxf(fps, 1.0)]
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
