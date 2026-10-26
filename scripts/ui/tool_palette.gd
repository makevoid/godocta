class_name ToolPalette
extends VBoxContainer
## The tool buttons in a floating left dock: orbit, brush in, black hole and the
## optional scoop. The scoop stays hidden until it is enabled in the settings menu.
## Per-tool options live in `ToolOptions`; this strip emits intent only, and
## `OctaStudio` applies it to `StructureTools` and calls `show_tool()`.

signal tool_selected(tool: StructureTools.Tool)
signal undo_requested
signal redo_requested
var undo_button: Button
var redo_button: Button
const SHORTCUTS: Array[String] = ["V", "B", "E", "S"]
const TIPS: Array[String] = [
	"Left drag orbits the camera.",
	"Paint new cells in the brush color: they are attracted to the cells under the brush, or rest on the ground.\nA stroke holds its distance from the camera, so it fills the globe once; click again for the next layer.",
	"A globe that swallows every cell within its radius.",
	"A half globe that shaves the surface without digging deep. Tilt it 45° with the arrow keys.",
]
var tool_buttons: Array[Button] = []
const ICONS: Array[Texture2D] = [
	preload("res://assets/ui/icons/orbit.svg"), preload("res://assets/ui/icons/brush.svg"),
	preload("res://assets/ui/icons/erase.svg"), preload("res://assets/ui/icons/scoop.svg"),
]

func _ready() -> void:
	add_theme_constant_override("separation", 8)
	add_child(StudioTheme.label("TOOLKIT", 10, StudioTheme.MUTED))
	var group := ButtonGroup.new()
	for index in StructureTools.NAMES.size():
		var button := StudioTheme.button("%s   %s" % [StructureTools.NAMES[index], SHORTCUTS[index]], 158, 46)
		button.icon = ICONS[index]
		button.alignment = HORIZONTAL_ALIGNMENT_LEFT
		button.toggle_mode = true
		button.add_theme_stylebox_override("pressed", StudioTheme.chrome())
		button.add_theme_stylebox_override("hover_pressed", StudioTheme.chrome(Color(1.15, 1.3, 1.25)))
		button.button_group = group
		button.tooltip_text = TIPS[index]
		button.pressed.connect(func() -> void: tool_selected.emit(index as StructureTools.Tool))
		tool_buttons.append(button)
		add_child(button)
	tool_buttons[StructureTools.Tool.SCOOP].visible = false
	add_child(HSeparator.new())
	var history := HBoxContainer.new()
	history.add_theme_constant_override("separation", 6)
	add_child(history)
	undo_button = StudioTheme.button("Undo", 76, 34)
	redo_button = StudioTheme.button("Redo", 76, 34)
	var modifier := "Cmd" if OS.get_name() == "macOS" else "Ctrl"
	undo_button.tooltip_text = "Undo the last stroke (%s+Z). Keeps up to 10 strokes." % modifier
	redo_button.tooltip_text = "Redo a stroke (%s+Shift+Z or Ctrl+Y)." % modifier
	undo_button.pressed.connect(func() -> void: undo_requested.emit())
	redo_button.pressed.connect(func() -> void: redo_requested.emit())
	history.add_child(undo_button)
	history.add_child(redo_button)
	show_history(false, false)
	show_tool(StructureTools.Tool.NAVIGATE)

func show_history(undo_available: bool, redo_available: bool) -> void:
	undo_button.disabled = not undo_available
	redo_button.disabled = not redo_available

## The scoop is an opt-in tool: the settings menu shows and hides its button.
func set_scoop_available(enabled: bool) -> void:
	tool_buttons[StructureTools.Tool.SCOOP].visible = enabled

func show_tool(tool: StructureTools.Tool) -> void:
	for index in tool_buttons.size():
		tool_buttons[index].set_pressed_no_signal(index == tool)
