class_name StudioTheme
extends RefCounted

const TEXT := Color("e5ede9")
const MUTED := Color("839990")
const ACCENT := Color("b9e9cb")
const PANEL := Color("111c20")
const LINE := Color("2b3b3e")

static func create() -> Theme:
	var theme := Theme.new()
	var font := SystemFont.new()
	font.font_names = PackedStringArray(["Inter", "Avenir Next", "Helvetica Neue"])
	theme.default_font = font
	theme.default_font_size = 15
	theme.set_color("font_color", "Label", TEXT)
	theme.set_color("font_color", "Button", TEXT)
	theme.set_color("font_hover_color", "Button", ACCENT)
	theme.set_color("font_pressed_color", "Button", ACCENT)
	theme.set_stylebox("normal", "Button", panel(Color("142125"), LINE, 8))
	theme.set_stylebox("hover", "Button", panel(Color("233631"), Color("648c77"), 8))
	theme.set_stylebox("pressed", "Button", panel(Color("263d32"), ACCENT, 8))
	theme.set_stylebox("focus", "Button", panel(Color(0, 0, 0, 0), ACCENT, 8))
	theme.set_constant("h_separation", "Button", 10)
	return theme

static func panel(color: Color = PANEL, border: Color = LINE, radius: int = 12) -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = color
	style.border_color = border
	style.set_border_width_all(1)
	style.set_corner_radius_all(radius)
	style.content_margin_left = 16
	style.content_margin_right = 16
	style.content_margin_top = 12
	style.content_margin_bottom = 12
	return style

static func label(text: String, size: int = 15, color: Color = TEXT) -> Label:
	var result := Label.new()
	result.text = text
	result.add_theme_font_size_override("font_size", size)
	result.add_theme_color_override("font_color", color)
	result.mouse_filter = Control.MOUSE_FILTER_IGNORE
	return result

static func button(text: String, width: float = 80.0) -> Button:
	var result := Button.new()
	result.text = text
	result.custom_minimum_size = Vector2(width, 44)
	result.mouse_default_cursor_shape = Control.CURSOR_POINTING_HAND
	return result
