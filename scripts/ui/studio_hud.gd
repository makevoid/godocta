class_name StudioHUD
extends CanvasLayer
## Chrome over the 3D view: top bar, floating toolkit, contextual options, welcome
## screen, build library, status line, toasts and the help card. Everything the
## user can act on lives in `StructureToolbar`, `ToolPalette` or `ToolOptions`; this
## class only presents state and forwards their signals' owners to `OctaStudio`.

signal overlay_changed
var toolbar: StructureToolbar
var welcome: StudioWelcome
var library: BuildLibrary
var _workspace: Control
var _help_dim: ColorRect
## Floating toolkit and the options panel for the selected tool.
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
	["Cmd/Ctrl + Z", "Undo a stroke (up to 10); add Shift to redo, or use Ctrl + Y"],
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
	_workspace = Control.new()
	_workspace.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	_workspace.mouse_filter = Control.MOUSE_FILTER_IGNORE
	root.add_child(_workspace)
	toolbar = StructureToolbar.new()
	_workspace.add_child(toolbar)
	var dock := PanelContainer.new()
	dock.position = Vector2(24, StructureToolbar.HEIGHT + 34)
	dock.add_theme_stylebox_override("panel", StudioTheme.card())
	_workspace.add_child(dock)
	tools = ToolPalette.new()
	dock.add_child(tools)
	_build_info()
	_build_assembly(_workspace)
	options = ToolOptions.new()
	_workspace.add_child(options)
	_build_status(_workspace)
	_build_help(_workspace)
	_toast = StudioTheme.label("", 12, StudioTheme.ACCENT)
	_toast.set_anchors_and_offsets_preset(Control.PRESET_BOTTOM_LEFT)
	_toast.offset_left = 24
	_toast.offset_top = -82
	_toast.offset_right = 620
	_workspace.add_child(_toast)
	welcome = StudioWelcome.new()
	root.add_child(welcome)
	library = BuildLibrary.new()
	root.add_child(library)
	library.close_requested.connect(func() -> void: set_library_visible(false))

## The structure readout, at the right of the top bar.
func _build_info() -> void:
	_title = StudioTheme.label("Cube", 18)
	_title.custom_minimum_size.x = 220
	_title.text_overrun_behavior = TextServer.OVERRUN_TRIM_ELLIPSIS
	_title.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
	toolbar.info_slot.add_child(_title)
	var row := HBoxContainer.new()
	row.add_theme_constant_override("separation", 6)
	row.alignment = BoxContainer.ALIGNMENT_END
	row.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_count = StudioTheme.label("1,729", 12, StudioTheme.ACCENT)
	row.add_child(_count)
	row.add_child(StudioTheme.label("CELLS  ·", 10, StudioTheme.MUTED))
	_memory = StudioTheme.label("20.3 KiB", 12, StudioTheme.MUTED)
	row.add_child(_memory)
	toolbar.info_slot.add_child(row)

func _build_status(root: Control) -> void:
	var footer := PanelContainer.new()
	footer.set_anchors_and_offsets_preset(Control.PRESET_BOTTOM_WIDE)
	footer.offset_top = -38
	footer.add_theme_stylebox_override("panel", StudioTheme.panel(StudioTheme.INK, StudioTheme.LINE, 0, 24, 8))
	root.add_child(footer)
	var row := HBoxContainer.new()
	row.add_theme_constant_override("separation", 16)
	footer.add_child(row)
	row.add_child(StudioTheme.label("●  SANDBOX", 10, StudioTheme.ACCENT))
	row.add_child(VSeparator.new())
	row.add_child(StudioTheme.label("DRAG  Orbit     /     SCROLL  Zoom     /     SHIFT + DRAG  Pan", 11, StudioTheme.MUTED))
	var space := Control.new()
	space.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(space)
	row.add_child(StudioTheme.label("H  Controls", 11, StudioTheme.MUTED))
	row.add_child(VSeparator.new())
	_fps = StudioTheme.label("— FPS", 11, StudioTheme.MUTED)
	_fps.custom_minimum_size.x = 130
	_fps.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
	row.add_child(_fps)

## A hairline progress bar under the top bar, with its clock, while cells still fly in.
func _build_assembly(root: Control) -> void:
	_progress = ProgressBar.new()
	_progress.set_anchors_and_offsets_preset(Control.PRESET_TOP_WIDE)
	_progress.offset_top = StructureToolbar.HEIGHT
	_progress.offset_bottom = StructureToolbar.HEIGHT + 3
	_progress.custom_minimum_size.y = 3
	_progress.max_value = 1.0
	_progress.show_percentage = false
	var track := StudioTheme.panel(StudioTheme.PANEL, StudioTheme.PANEL, 0, 0, 0)
	_progress.add_theme_stylebox_override("background", track)
	_progress.add_theme_stylebox_override("fill", StudioTheme.panel(StudioTheme.ACCENT, StudioTheme.ACCENT, 0, 0, 0))
	root.add_child(_progress)
	_clock = HBoxContainer.new()
	var clock := _clock
	clock.set_anchors_and_offsets_preset(Control.PRESET_TOP_LEFT)
	clock.offset_left = 24
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
	_help_dim = ColorRect.new()
	_help_dim.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	_help_dim.color = Color(0.01, 0.025, 0.045, 0.75)
	_help_dim.visible = false
	root.add_child(_help_dim)
	help = PanelContainer.new()
	help.set_anchors_and_offsets_preset(Control.PRESET_CENTER)
	help.grow_horizontal = Control.GROW_DIRECTION_BOTH
	help.grow_vertical = Control.GROW_DIRECTION_BOTH
	help.add_theme_stylebox_override("panel", StudioTheme.panel(StudioTheme.INK, Color("37677c"), 12, 30, 26))
	help.visible = false
	root.add_child(help)
	var column := VBoxContainer.new()
	column.add_theme_constant_override("separation", 12)
	help.add_child(column)
	var header := HBoxContainer.new()
	header.add_theme_constant_override("separation", 12)
	header.add_child(StudioTheme.label("Make yourself at home.", 28))
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
		var key := StudioTheme.label(entry[0], 13, StudioTheme.ACCENT)
		key.custom_minimum_size.x = 76
		key.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
		keys.add_child(key)
		keys.add_child(StudioTheme.label(entry[1], 14, StudioTheme.MUTED))
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
	_help_dim.visible = shown
	overlay_changed.emit()

func set_welcome_visible(shown: bool) -> void:
	welcome.visible = shown
	_workspace.visible = not shown
	if shown:
		set_help_visible(false)
		welcome.enter_button.grab_focus()
	else:
		get_viewport().gui_release_focus()
	overlay_changed.emit()

func set_library_visible(shown: bool) -> void:
	if shown:
		library.open(toolbar.gallery_paths)
	else:
		library.hide()
		get_viewport().gui_release_focus()
	overlay_changed.emit()

func has_overlay() -> bool:
	return welcome.visible or library.visible or help.visible

func show_structure(structure: OctaStructure, preset: int = -1) -> void:
	_title.text = structure.title
	_title.tooltip_text = structure.title
	_toast.text = ""
	_toast_remaining = 0.0
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
