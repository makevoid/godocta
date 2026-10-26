class_name StudioTheme
extends RefCounted
## Shared blueprint palette and scalable controls. Text always stays native.

const TEXT := Color("e8f6ff")
const MUTED := Color("93acbd")
const ACCENT := Color("53f5de")
const PANEL := Color("0c1d2b")
const LINE := Color("294457")
const INK := Color("07131f")
const BUTTON_HEIGHT := 38.0

static func create() -> Theme:
	var theme := Theme.new()
	var font := SystemFont.new()
	font.font_names = PackedStringArray(["Inter", "Avenir Next", "Helvetica Neue"])
	theme.default_font = font
	theme.default_font_size = 14
	for type in ["Label", "Button", "PopupMenu", "CheckButton", "LineEdit", "Tree", "ItemList"]:
		theme.set_color("font_color", type, TEXT)
	for type in ["Button", "MenuButton", "OptionButton"]:
		theme.set_color("font_color", type, TEXT)
		theme.set_color("font_hover_color", type, ACCENT)
		theme.set_color("font_pressed_color", type, ACCENT)
		theme.set_color("font_hover_pressed_color", type, ACCENT)
		theme.set_color("font_disabled_color", type, Color("536c7c"))
		theme.set_stylebox("normal", type, panel(Color("102737"), LINE, 6, 14, 8))
		theme.set_stylebox("hover", type, panel(Color("173c4b"), Color("4a9eaa"), 6, 14, 8))
		theme.set_stylebox("pressed", type, panel(Color("17483f"), ACCENT, 6, 14, 8))
		theme.set_stylebox("hover_pressed", type, panel(Color("205d53"), ACCENT, 6, 14, 8))
		theme.set_stylebox("disabled", type, panel(Color("0b1b26"), Color("1c3040"), 6, 14, 8))
		theme.set_stylebox("focus", type, panel(Color(0, 0, 0, 0), ACCENT, 6, 0, 0))
		theme.set_constant("h_separation", type, 10)
	theme.set_stylebox("panel", "PopupMenu", panel(PANEL, Color("3b6577"), 10, 10, 10))
	theme.set_stylebox("hover", "PopupMenu", panel(Color("16413f"), Color("28665f"), 5, 10, 6))
	theme.set_color("font_hover_color", "PopupMenu", ACCENT)
	theme.set_color("font_separator_color", "PopupMenu", MUTED)
	theme.set_color("font_accelerator_color", "PopupMenu", MUTED)
	theme.set_constant("v_separation", "PopupMenu", 14)
	theme.set_constant("h_separation", "PopupMenu", 16)
	theme.set_icon("checked", "PopupMenu", preload("res://assets/ui/icons/checked.svg"))
	theme.set_icon("unchecked", "PopupMenu", preload("res://assets/ui/icons/unchecked.svg"))
	theme.set_icon("radio_checked", "PopupMenu", preload("res://assets/ui/icons/radio-checked.svg"))
	theme.set_icon("radio_unchecked", "PopupMenu", preload("res://assets/ui/icons/radio-unchecked.svg"))
	theme.set_stylebox("panel", "PanelContainer", card())
	theme.set_stylebox("panel", "AcceptDialog", card())
	theme.set_stylebox("panel", "Window", panel(INK, LINE, 10, 16, 16))
	theme.set_color("title_color", "Window", TEXT)
	theme.set_color("font_color", "TooltipLabel", TEXT)
	theme.set_stylebox("panel", "TooltipPanel", panel(PANEL, LINE, 6, 12, 8))
	for type in ["Tree", "ItemList", "LineEdit"]:
		theme.set_stylebox("panel" if type != "LineEdit" else "normal", type, panel(INK, LINE, 5, 10, 8))
		theme.set_stylebox("selected", type, panel(Color("1a4b4a"), ACCENT, 4, 4, 4))
		theme.set_stylebox("selected_focus", type, panel(Color("1a4b4a"), ACCENT, 4, 4, 4))
		theme.set_color("font_selected_color", type, TEXT)
	theme.set_stylebox("focus", "LineEdit", panel(Color(0, 0, 0, 0), ACCENT, 5, 10, 8))
	theme.set_color("caret_color", "LineEdit", ACCENT)
	theme.set_color("selection_color", "LineEdit", Color("24635e"))
	var track := panel(Color("173344"), Color("173344"), 3, 0, 2)
	var fill := panel(ACCENT, ACCENT, 3, 0, 2)
	theme.set_stylebox("slider", "HSlider", track)
	theme.set_stylebox("grabber_area", "HSlider", fill)
	theme.set_stylebox("grabber_area_highlight", "HSlider", fill)
	var dot := preload("res://assets/ui/icons/slider.svg")
	theme.set_icon("grabber", "HSlider", dot)
	theme.set_icon("grabber_highlight", "HSlider", dot)
	var separator := StyleBoxLine.new()
	separator.color = LINE
	separator.thickness = 1
	theme.set_stylebox("separator", "HSeparator", separator)
	var vertical := separator.duplicate() as StyleBoxLine
	vertical.vertical = true
	theme.set_stylebox("separator", "VSeparator", vertical)
	return theme

static func panel(color: Color = PANEL, border: Color = LINE, radius: int = 8,
		margin_x: int = 14, margin_y: int = 10) -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = color
	style.border_color = border
	style.set_border_width_all(1)
	style.set_corner_radius_all(radius)
	style.content_margin_left = margin_x
	style.content_margin_right = margin_x
	style.content_margin_top = margin_y
	style.content_margin_bottom = margin_y
	return style

static func card() -> StyleBoxFlat:
	var style := panel(Color(0.035, 0.085, 0.13, 0.97), LINE, 10, 18, 16)
	style.shadow_color = Color(0.0, 0.015, 0.025, 0.24)
	style.shadow_size = 8
	style.shadow_offset = Vector2(0, 4)
	return style

static func label(text: String, size: int = 14, color: Color = TEXT) -> Label:
	var result := Label.new()
	result.text = text
	result.add_theme_font_size_override("font_size", size)
	result.add_theme_color_override("font_color", color)
	result.mouse_filter = Control.MOUSE_FILTER_IGNORE
	return result

static func button(text: String, width: float = 0.0, height: float = BUTTON_HEIGHT) -> Button:
	var result := Button.new()
	result.text = text
	result.custom_minimum_size = Vector2(width, height)
	# World shortcuts (Space, arrows) must still work after clicking a tool.
	result.focus_mode = Control.FOCUS_NONE
	result.mouse_default_cursor_shape = Control.CURSOR_POINTING_HAND
	return result

## Sunburst skin uses a nine-slice so the same artwork fits large and small buttons.
static func chrome(tint: Color = Color.WHITE, margin_x: int = 16, margin_y: int = 8) -> StyleBoxTexture:
	var style := StyleBoxTexture.new()
	style.texture = preload("res://assets/ui/button-chrome-runtime.png")
	style.modulate_color = tint
	for side in [SIDE_LEFT, SIDE_TOP, SIDE_RIGHT, SIDE_BOTTOM]:
		style.set_texture_margin(side, 10)
	style.axis_stretch_horizontal = StyleBoxTexture.AXIS_STRETCH_MODE_STRETCH
	style.axis_stretch_vertical = StyleBoxTexture.AXIS_STRETCH_MODE_STRETCH
	style.set_expand_margin_all(0)
	style.content_margin_left = margin_x
	style.content_margin_right = margin_x
	style.content_margin_top = margin_y
	style.content_margin_bottom = margin_y
	return style

static func primary(button: Button) -> Button:
	button.add_theme_color_override("font_color", ACCENT)
	button.add_theme_stylebox_override("normal", chrome(Color(0.8, 0.9, 0.9)))
	button.add_theme_stylebox_override("hover", chrome(Color(1.2, 1.45, 1.4)))
	button.add_theme_stylebox_override("pressed", chrome(Color(0.7, 0.95, 0.9)))
	return button

static func hero_button(text: String, width: float = 320) -> Button:
	var result := button(text, width, 60)
	result.focus_mode = Control.FOCUS_ALL
	result.add_theme_font_size_override("font_size", 18)
	result.add_theme_color_override("font_color", ACCENT)
	result.add_theme_color_override("font_focus_color", ACCENT)
	result.add_theme_stylebox_override("normal", chrome(Color.WHITE, 24, 12))
	result.add_theme_stylebox_override("hover", chrome(Color(1.2, 1.45, 1.4), 24, 12))
	result.add_theme_stylebox_override("pressed", chrome(Color(0.7, 0.95, 0.9), 24, 12))
	return result

static func menu(text: String, width: float = 0.0) -> MenuButton:
	var result := MenuButton.new()
	result.text = text
	result.custom_minimum_size = Vector2(width, BUTTON_HEIGHT)
	result.flat = false
	result.focus_mode = Control.FOCUS_NONE
	result.mouse_default_cursor_shape = Control.CURSOR_POINTING_HAND
	return result
