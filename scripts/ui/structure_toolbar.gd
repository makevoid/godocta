class_name StructureToolbar
extends PanelContainer
## The top bar: home, the Primitives and Builds menus, the structure
## readout filled by `StudioHUD`, file actions, help and the settings menu. Primitives
## sit in a menu because builds are the main entry point. Infrequent switches — X-ray,
## auto orbit and the optional scoop — live in the settings menu instead of the bar.

signal shape_selected(shape: StructurePresets.Shape)
signal file_selected(path: String)
signal replay_requested
signal save_requested
signal load_requested
signal reset_view_requested
signal help_requested
signal home_requested
signal library_requested
signal orbit_changed(enabled: bool)
signal xray_changed(enabled: bool)
signal scoop_enabled_changed(enabled: bool)
const STRUCTURES_DIR := "res://structures"
const HEIGHT := 76.0
## Settings popup item ids.
enum Setting { XRAY, ORBIT, SCOOP }
var shapes: MenuButton
var gallery: MenuButton
var settings: MenuButton
var gallery_paths: PackedStringArray = []
## Where `StudioHUD` adds the title and counters.
var info_slot: VBoxContainer

func _ready() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_TOP_WIDE)
	offset_bottom = HEIGHT
	var background := StudioTheme.panel(StudioTheme.INK, StudioTheme.LINE, 0, 24, 16)
	add_theme_stylebox_override("panel", background)
	var row := HBoxContainer.new()
	row.add_theme_constant_override("separation", 12)
	add_child(row)
	var home := StudioTheme.button("O C T A", 150, 42)
	home.icon = preload("res://assets/octa.svg")
	home.add_theme_constant_override("icon_max_width", 30)
	home.add_theme_font_size_override("font_size", 20)
	home.tooltip_text = "Back to the welcome screen"
	home.pressed.connect(func() -> void: home_requested.emit())
	row.add_child(home)
	row.add_child(VSeparator.new())
	_build_shapes(row)
	_build_gallery(row)
	var library := StudioTheme.button("Explore builds", 132)
	StudioTheme.primary(library)
	library.tooltip_text = "Search the full collection of worlds"
	library.pressed.connect(func() -> void: library_requested.emit())
	row.add_child(library)
	var spacer := Control.new()
	spacer.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(spacer)
	info_slot = VBoxContainer.new()
	info_slot.add_theme_constant_override("separation", 0)
	info_slot.alignment = BoxContainer.ALIGNMENT_CENTER
	info_slot.mouse_filter = Control.MOUSE_FILTER_IGNORE
	row.add_child(info_slot)
	row.add_child(VSeparator.new())
	_build_actions(row)

## The four primitives, behind one menu: generated builds are the usual starting point.
func _build_shapes(row: HBoxContainer) -> void:
	shapes = StudioTheme.menu("Primitives  ▾", 108)
	shapes.icon = preload("res://assets/cube.svg")
	shapes.tooltip_text = "Assemble a primitive · shortcuts 1–4"
	var popup := shapes.get_popup()
	for index in 4:
		popup.add_radio_check_item("%s    %d" % [StructurePresets.TITLES[index], index + 1], index)
	popup.id_pressed.connect(func(id: int) -> void: shape_selected.emit(id as StructurePresets.Shape))
	row.add_child(shapes)

## Every .octa file in structures/: generated builds (castle, arena, …) and saved ones.
func _build_gallery(row: HBoxContainer) -> void:
	gallery = StudioTheme.menu("Builds  ▾", 118)
	gallery.icon = preload("res://assets/castle.svg")
	gallery.tooltip_text = "Large structures from structures/*.octa · shortcuts 5–9"
	gallery.get_popup().id_pressed.connect(func(id: int) -> void: file_selected.emit(gallery_paths[id]))
	row.add_child(gallery)
	refresh_gallery()

func _build_actions(row: HBoxContainer) -> void:
	var load_button := StudioTheme.button("Load", 58)
	load_button.tooltip_text = "Open an .octa structure"
	load_button.pressed.connect(func() -> void: load_requested.emit())
	row.add_child(load_button)
	var save_button := StudioTheme.button("Save", 58)
	save_button.tooltip_text = "Save the arrangement as a compact .octa file"
	save_button.pressed.connect(func() -> void: save_requested.emit())
	row.add_child(save_button)
	var reset := StudioTheme.button("Frame  F", 88)
	reset.tooltip_text = "Restore the default framing"
	reset.pressed.connect(func() -> void: reset_view_requested.emit())
	row.add_child(reset)
	var replay := StudioTheme.button("Replay  R", 94)
	replay.add_theme_color_override("font_color", StudioTheme.ACCENT)
	replay.pressed.connect(func() -> void: replay_requested.emit())
	row.add_child(replay)
	var help := StudioTheme.button("?", 32)
	help.tooltip_text = "Controls and shortcuts · H"
	help.pressed.connect(func() -> void: help_requested.emit())
	row.add_child(help)
	settings = StudioTheme.menu("Settings  ▾", 92)
	settings.tooltip_text = "X-ray, auto orbit and the scoop tool"
	var popup := settings.get_popup()
	popup.add_check_item("X-ray view    X", Setting.XRAY)
	popup.set_item_tooltip(Setting.XRAY, "Semi-transparent cells with every edge and vertex visible.")
	popup.add_check_item("Auto orbit", Setting.ORBIT)
	popup.add_check_item("Enable scoop    S", Setting.SCOOP)
	popup.set_item_tooltip(Setting.SCOOP, "Add the scoop tool, a half globe that shaves the surface.")
	popup.hide_on_checkable_item_selection = false
	popup.id_pressed.connect(_toggle_setting)
	row.add_child(settings)

## Settings are checkable: flip the item, then report it.
func _toggle_setting(id: int) -> void:
	var popup := settings.get_popup()
	var index := popup.get_item_index(id)
	var enabled := not popup.is_item_checked(index)
	popup.set_item_checked(index, enabled)
	match id:
		Setting.XRAY:
			xray_changed.emit(enabled)
		Setting.ORBIT:
			orbit_changed.emit(enabled)
		Setting.SCOOP:
			scoop_enabled_changed.emit(enabled)

## Reflect a setting changed elsewhere, e.g. by a keyboard shortcut.
func set_setting_checked(id: Setting, enabled: bool) -> void:
	var popup := settings.get_popup()
	popup.set_item_checked(popup.get_item_index(id), enabled)

func is_setting_checked(id: Setting) -> bool:
	var popup := settings.get_popup()
	return popup.is_item_checked(popup.get_item_index(id))

func refresh_gallery() -> void:
	var files := Array(DirAccess.get_files_at(STRUCTURES_DIR)).filter(
		func(name: String) -> bool: return name.get_extension().to_lower() == "octa")
	files.sort()
	gallery_paths.clear()
	var popup := gallery.get_popup()
	popup.clear()
	for name: String in files:
		gallery_paths.append(STRUCTURES_DIR.path_join(name))
		var shortcut := "    %d" % (gallery_paths.size() + 4) if gallery_paths.size() <= 5 else ""
		popup.add_item(name.get_basename().capitalize() + shortcut, gallery_paths.size() - 1)

## Check the displayed primitive, or clear every entry for a loaded build.
func select_shape(index: int) -> void:
	var popup := shapes.get_popup()
	for item in popup.item_count:
		popup.set_item_checked(item, item == index)
