class_name StudioWelcome
extends Control
## Interactive splash. Sunburst supplies the art; all text and actions are native UI.

signal enter_requested
signal library_requested
signal load_requested
var enter_button: Button

func _ready() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	mouse_filter = Control.MOUSE_FILTER_STOP
	var art := TextureRect.new()
	art.texture = preload("res://assets/ui/splash.png")
	art.expand_mode = TextureRect.EXPAND_IGNORE_SIZE
	art.stretch_mode = TextureRect.STRETCH_KEEP_ASPECT_COVERED
	art.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	art.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(art)
	var margin := MarginContainer.new()
	margin.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	for side in ["left", "right"]:
		margin.add_theme_constant_override("margin_" + side, 76)
	for side in ["top", "bottom"]:
		margin.add_theme_constant_override("margin_" + side, 48)
	margin.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(margin)
	var layout := VBoxContainer.new()
	layout.mouse_filter = Control.MOUSE_FILTER_IGNORE
	margin.add_child(layout)
	var header := HBoxContainer.new()
	header.add_theme_constant_override("separation", 16)
	var logo := TextureRect.new()
	logo.texture = preload("res://assets/octa.svg")
	logo.expand_mode = TextureRect.EXPAND_IGNORE_SIZE
	logo.stretch_mode = TextureRect.STRETCH_KEEP_ASPECT_CENTERED
	logo.custom_minimum_size = Vector2(46, 46)
	header.add_child(logo)
	header.add_child(StudioTheme.label("O C T A", 28))
	var space := Control.new()
	space.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	header.add_child(space)
	var badge := PanelContainer.new()
	badge.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	badge.add_theme_stylebox_override("panel", StudioTheme.panel(StudioTheme.INK, StudioTheme.LINE, 5, 16, 10))
	var edition := StudioTheme.label("STRUCTURE STUDIO   /   CREATIVE SANDBOX", 12, StudioTheme.MUTED)
	edition.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	badge.add_child(edition)
	header.add_child(badge)
	layout.add_child(header)
	var body := VBoxContainer.new()
	body.size_flags_vertical = Control.SIZE_EXPAND_FILL
	body.size_flags_horizontal = Control.SIZE_SHRINK_BEGIN
	body.alignment = BoxContainer.ALIGNMENT_CENTER
	body.add_theme_constant_override("separation", 22)
	layout.add_child(body)
	body.add_child(StudioTheme.label("I M A G I N A T I O N ,   A S S E M B L E D .", 12, StudioTheme.ACCENT))
	var title := StudioTheme.label("Small blocks.\nInfinite worlds.", 62)
	title.add_theme_constant_override("line_spacing", -5)
	body.add_child(title)
	var subtitle := StudioTheme.label("Explore remarkable structures. Shape something new.\nA whole world of possibilities, one cell at a time.", 18, StudioTheme.MUTED)
	subtitle.add_theme_constant_override("line_spacing", 5)
	body.add_child(subtitle)
	var gap := Control.new()
	gap.custom_minimum_size.y = 10
	body.add_child(gap)
	enter_button = StudioTheme.hero_button("Enter the studio     →", 356)
	enter_button.size_flags_horizontal = Control.SIZE_SHRINK_BEGIN
	enter_button.pressed.connect(func() -> void: enter_requested.emit())
	body.add_child(enter_button)
	var secondary := HBoxContainer.new()
	secondary.add_theme_constant_override("separation", 12)
	var explore := StudioTheme.button("Explore builds", 172, 44)
	explore.focus_mode = Control.FOCUS_ALL
	explore.pressed.connect(func() -> void: library_requested.emit())
	secondary.add_child(explore)
	var open := StudioTheme.button("Open .octa", 172, 44)
	open.focus_mode = Control.FOCUS_ALL
	open.pressed.connect(func() -> void: load_requested.emit())
	secondary.add_child(open)
	body.add_child(secondary)
	body.add_child(StudioTheme.label("Press Enter to jump in", 12, StudioTheme.MUTED))
	var footer := HBoxContainer.new()
	footer.add_theme_constant_override("separation", 18)
	footer.add_child(StudioTheme.label("01  /  EXPLORE", 11, StudioTheme.ACCENT))
	footer.add_child(StudioTheme.label("02  /  ASSEMBLE", 11, StudioTheme.MUTED))
	footer.add_child(StudioTheme.label("03  /  MAKE IT YOURS", 11, StudioTheme.MUTED))
	var footer_space := Control.new()
	footer_space.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	footer.add_child(footer_space)
	footer.add_child(StudioTheme.label("ONE CELL.  FOURTEEN FACES.  ENDLESS POSSIBILITIES.", 10, StudioTheme.MUTED))
	layout.add_child(footer)
