class_name ToolPalette
extends HBoxContainer
## The tool buttons at the left of the top bar: orbit, brush in, black hole and the
## optional scoop. The scoop stays hidden until it is enabled in the settings menu.
## Per-tool options live in `ToolOptions`; this strip emits intent only, and
## `OctaStudio` applies it to `StructureTools` and calls `show_tool()`.

signal tool_selected(tool: StructureTools.Tool)
const SHORTCUTS: Array[String] = ["V", "B", "E", "S"]
const TIPS: Array[String] = [
	"Left drag orbits the camera.",
	"Paint new cells in the brush color: they are attracted to the cells under the brush, or rest on the ground.\nA stroke holds its distance from the camera, so it fills the globe once; click again for the next layer.",
	"A globe that swallows every cell within its radius.",
	"A half globe that shaves the surface without digging deep. Tilt it 45° with the arrow keys.",
]
var tool_buttons: Array[Button] = []

func _ready() -> void:
	add_theme_constant_override("separation", 6)
	var group := ButtonGroup.new()
	for index in StructureTools.NAMES.size():
		var button := StudioTheme.button("%s  %s" % [StructureTools.NAMES[index], SHORTCUTS[index]])
		button.toggle_mode = true
		button.button_group = group
		button.tooltip_text = TIPS[index]
		button.pressed.connect(func() -> void: tool_selected.emit(index as StructureTools.Tool))
		tool_buttons.append(button)
		add_child(button)
	tool_buttons[StructureTools.Tool.SCOOP].visible = false
	show_tool(StructureTools.Tool.NAVIGATE)

## The scoop is an opt-in tool: the settings menu shows and hides its button.
func set_scoop_available(enabled: bool) -> void:
	tool_buttons[StructureTools.Tool.SCOOP].visible = enabled

func show_tool(tool: StructureTools.Tool) -> void:
	for index in tool_buttons.size():
		tool_buttons[index].set_pressed_no_signal(index == tool)
