class_name BuildLibrary
extends Control
## Searchable build menu. Paths and ordering come from the toolbar's gallery.

signal file_selected(path: String)
signal close_requested
var search: LineEdit
var _grid: GridContainer
var _count: Label
var _empty: Label
var _paths: PackedStringArray = []
var _buttons: Array[Button] = []

func _ready() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	mouse_filter = Control.MOUSE_FILTER_STOP
	var dim := ColorRect.new()
	dim.color = Color(0.01, 0.035, 0.065, 0.84)
	dim.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	dim.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(dim)
	var panel := PanelContainer.new()
	panel.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	panel.anchor_left = 0.16
	panel.anchor_right = 0.84
	panel.anchor_top = 0.12
	panel.anchor_bottom = 0.88
	panel.add_theme_stylebox_override("panel", StudioTheme.panel(StudioTheme.INK, Color("37677c"), 12, 30, 26))
	add_child(panel)
	var column := VBoxContainer.new()
	column.add_theme_constant_override("separation", 18)
	panel.add_child(column)
	var top := HBoxContainer.new()
	var titles := VBoxContainer.new()
	titles.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	titles.add_child(StudioTheme.label("THE OCTA COLLECTION", 11, StudioTheme.ACCENT))
	titles.add_child(StudioTheme.label("Find your next world.", 32))
	top.add_child(titles)
	var close := StudioTheme.button("Close  Esc", 110)
	close.focus_mode = Control.FOCUS_ALL
	close.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	close.pressed.connect(func() -> void: close_requested.emit())
	top.add_child(close)
	column.add_child(top)
	column.add_child(StudioTheme.label("Landmarks, landscapes and impossible places. Pick a build and watch it come together.", 15, StudioTheme.MUTED))
	search = LineEdit.new()
	search.placeholder_text = "Search builds…"
	search.custom_minimum_size.y = 46
	search.clear_button_enabled = true
	search.text_changed.connect(_filter)
	column.add_child(search)
	_count = StudioTheme.label("", 11, StudioTheme.ACCENT)
	column.add_child(_count)
	var scroll := ScrollContainer.new()
	scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	scroll.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	column.add_child(scroll)
	_grid = GridContainer.new()
	_grid.columns = 3
	_grid.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_grid.add_theme_constant_override("h_separation", 12)
	_grid.add_theme_constant_override("v_separation", 12)
	scroll.add_child(_grid)
	_empty = StudioTheme.label("No builds match. Try a different name.", 16, StudioTheme.MUTED)
	_empty.visible = false
	column.add_child(_empty)
	column.add_child(StudioTheme.label("Every world is editable. Use Brush in to grow it, or Black hole to carve your own path.", 12, StudioTheme.MUTED))
	visible = false

func open(paths: PackedStringArray) -> void:
	_paths = paths.duplicate()
	for button in _buttons:
		_grid.remove_child(button)
		button.queue_free()
	_buttons.clear()
	for index in paths.size():
		var path := paths[index]
		var button := StudioTheme.button(path.get_file().get_basename().capitalize(), 0, 66)
		button.focus_mode = Control.FOCUS_ALL
		button.icon = preload("res://assets/castle.svg")
		button.alignment = HORIZONTAL_ALIGNMENT_LEFT
		button.text_overrun_behavior = TextServer.OVERRUN_TRIM_ELLIPSIS
		button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		button.tooltip_text = button.text + " · open build"
		button.pressed.connect(func() -> void: file_selected.emit(path))
		_grid.add_child(button)
		_buttons.append(button)
	search.text = ""
	_filter("")
	show()
	search.grab_focus()

func _filter(query: String) -> void:
	var found := 0
	for index in _buttons.size():
		var matches := query.strip_edges().is_empty() or _buttons[index].text.to_lower().contains(query.strip_edges().to_lower())
		_buttons[index].visible = matches
		found += 1 if matches else 0
	_count.text = "%02d / %02d BUILDS" % [found, _paths.size()]
	_empty.visible = found == 0
