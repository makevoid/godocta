class_name StudioTheme
extends RefCounted

const TEXT := Color("e5ede9")
const MUTED := Color("839990")
const ACCENT := Color("b9e9cb")
const PANEL := Color("111c20")
const LINE := Color("2b3b3e")
## Height of a top bar or panel button, in layout points.
const BUTTON_HEIGHT := 30.0

static func create() -> Theme:
	var theme := Theme.new()
	var font := SystemFont.new()
	font.font_names = PackedStringArray(["Inter", "Avenir Next", "Helvetica Neue"])
	theme.default_font = font
	theme.default_font_size = 13
	theme.set_color("font_color", "Label", TEXT)
	theme.set_color("font_color", "Button", TEXT)
	theme.set_color("font_hover_color", "Button", ACCENT)
	theme.set_color("font_pressed_color", "Button", ACCENT)
	theme.set_stylebox("normal", "Button", panel(Color("142125"), LINE, 6, 9, 5))
	theme.set_stylebox("hover", "Button", panel(Color("233631"), Color("648c77"), 6, 9, 5))
	theme.set_stylebox("pressed", "Button", panel(Color("263d32"), ACCENT, 6, 9, 5))
	theme.set_stylebox("focus", "Button", panel(Color(0, 0, 0, 0), ACCENT, 6, 9, 5))
	theme.set_constant("h_separation", "Button", 8)
	return theme

static func panel(color: Color = PANEL, border: Color = LINE, radius: int = 8,
		margin_x: int = 11, margin_y: int = 8) -> StyleBoxFlat:
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

## A translucent card, for the panels that float over the 3D view.
static func card() -> StyleBoxFlat:
	return panel(Color(0.04, 0.06, 0.07, 0.78), Color(1, 1, 1, 0.06), 10)

static func label(text: String, size: int = 13, color: Color = TEXT) -> Label:
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
	result.focus_mode = Control.FOCUS_NONE
	result.mouse_default_cursor_shape = Control.CURSOR_POINTING_HAND
	return result

## Mark the main entry point: an accent-tinted fill and border.
static func primary(button: Button) -> Button:
	button.add_theme_color_override("font_color", ACCENT)
	button.add_theme_stylebox_override("normal", panel(Color("1b3029"), Color("4f7a63"), 6, 9, 5))
	button.add_theme_stylebox_override("hover", panel(Color("27443a"), ACCENT, 6, 9, 5))
	button.add_theme_stylebox_override("pressed", panel(Color("27443a"), ACCENT, 6, 9, 5))
	return button

## A MenuButton styled like the regular buttons; its popup carries the entries.
static func menu(text: String, width: float = 0.0) -> MenuButton:
	var result := MenuButton.new()
	result.text = text
	result.custom_minimum_size = Vector2(width, BUTTON_HEIGHT)
	result.flat = false # MenuButton inherits the theme's Button styles once unflattened.
	result.focus_mode = Control.FOCUS_NONE
	result.mouse_default_cursor_shape = Control.CURSOR_POINTING_HAND
	return result
