extends SceneTree
## Records a scripted one-minute UI walkthrough of the real studio to an MP4: welcome,
## primitives, brush in, black hole, undo/redo, orbit, Load and the build library.
## Run it through tools/walkthrough.sh, which keeps the window hidden in the background.
##   --out=captures/walkthrough.mp4   Output video (1920×1080 H.264).
##   --fps=30                         Also pass the same value as Godot's --fixed-fps.
##   --fast=2                         Assembly speed-up for loaded builds (1 = real time).
## The studio runs inside an offscreen SubViewport rendered at 2× and downscaled by
## ffmpeg. Scripted mouse and key events are pushed into that viewport (a root window
## would read the real OS pointer), and a drawn cursor, click ripples and captions are
## composited on top. With --fixed-fps every frame advances exactly 1/fps seconds, so
## the video plays smoothly however long each frame takes to render and encode.

const SIZE := Vector2i(1920, 1080) # Layout points; rendered at 2× for crisp text and edges.
const SCALE := 2
const FFMPEG := ["/opt/homebrew/bin/ffmpeg", "/usr/local/bin/ffmpeg", "/usr/bin/ffmpeg"]

var studio: OctaStudio
var outer: SubViewport # Studio picture plus the cursor and caption overlay: what is recorded.
var inner: SubViewport # The studio itself, receiving the scripted input.
var overlay: WalkthroughOverlay
var fps: int = 30
var fast: float = 2.0
var out_path: String = "res://captures/walkthrough.mp4"
var _pointer := Vector2(1340, 900)
var _pipe: Dictionary = {}
var _frames: int = 0
var _last_drawn: int = -1
var _skipped: int = 0

func _initialize() -> void:
	var options := _parse_options()
	fps = int(options.get("fps", fps))
	fast = float(options.get("fast", fast))
	out_path = options.get("out", out_path)
	if not out_path.begins_with("res://") and not out_path.begins_with("/"):
		out_path = "res://" + out_path
	# Deliver scripted events immediately, so polled button state matches the events.
	Input.use_accumulated_input = false
	outer = _viewport()
	root.add_child(outer)
	inner = _viewport()
	inner.own_world_3d = true
	inner.gui_embed_subwindows = true # Menus, tooltips and dialogs draw into the recording.
	inner.msaa_3d = Viewport.MSAA_2X
	# A child viewport renders before its parent, so the overlay never lags the studio.
	outer.add_child(inner)
	var picture := TextureRect.new()
	picture.texture = inner.get_texture()
	picture.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	picture.expand_mode = TextureRect.EXPAND_IGNORE_SIZE
	outer.add_child(picture)
	overlay = WalkthroughOverlay.new()
	outer.add_child(overlay)
	studio = load("res://scenes/main.tscn").instantiate() as OctaStudio
	inner.add_child(studio)
	call_deferred("_run")

func _viewport() -> SubViewport:
	var viewport := SubViewport.new()
	viewport.size = SIZE * SCALE
	viewport.size_2d_override = SIZE
	viewport.size_2d_override_stretch = true
	viewport.render_target_update_mode = SubViewport.UPDATE_ALWAYS
	return viewport

func _run() -> void:
	# Unrecorded warm-up: a hidden app still gets window focus events shortly after launch,
	# and one landing between a press and its release would drop the click.
	var ready_at := Time.get_ticks_msec() + 3000
	while Time.get_ticks_msec() < ready_at:
		await process_frame
	inner.notification(Node.NOTIFICATION_VP_MOUSE_ENTER)
	# Hidden windows are throttled, so the live counter would only show the recorder's pace.
	studio.hud._fps.visible = false
	overlay.pointer = _pointer
	overlay.fade(1.0, 0.0, 0.7)
	_move_event(_pointer)
	_widen()
	await _wait(0.3)
	await _welcome()
	await _primitives()
	await _brush()
	await _black_hole()
	await _history()
	await _orbit()
	await _load_castle()
	await _carve_castle()
	await _explore()
	await _paint_taj()
	await _finish()

# --- Scenes -------------------------------------------------------------------------

func _welcome() -> void:
	overlay.caption("Octa — a structure studio built from truncated octahedra")
	await _wait(0.5)
	await _move(_center(studio.hud.welcome.enter_button), 1.2, 60.0)
	await _wait(0.4)
	await _click_until(func() -> bool: return not studio.hud.welcome.visible)
	overlay.caption("Every build assembles cell by cell")
	await _wait(0.6)
	await _move(Vector2(1080, 600), 0.8, -40.0)
	await _drag(Vector2(-260, 30), 2.2)
	await _wait(0.3)

func _primitives() -> void:
	overlay.caption("Primitives and builds live in the top bar")
	await _move(_center(studio.hud.toolbar.shapes), 0.9, 40.0)
	await _click_until(func() -> bool: return studio.hud.toolbar.shapes.get_popup().visible)
	await _wait(0.4)
	await _move(_popup_item(studio.hud.toolbar.shapes.get_popup(), StructurePresets.Shape.SPHERE), 0.6)
	await _wait(0.2)
	await _click_until(func() -> bool: return studio.renderer.structure.title == "Sphere")
	_widen()
	await _wait(0.5)
	overlay.caption("Brush in — paint new cells onto the surface")
	await _move(_center(studio.hud.tools.tool_buttons[StructureTools.Tool.BRUSH_IN]), 1.0, 50.0)
	await _click_until(func() -> bool: return studio.tools.tool == StructureTools.Tool.BRUSH_IN)
	await _wait_assembled()

func _brush() -> void:
	var center := _screen(studio.camera.focus)
	var r := _screen_radius(4.8)
	await _move(center + Vector2(-0.55, -0.62) * r, 0.9, 40.0)
	await _press()
	await _move(center + Vector2(0.15, -0.8) * r, 1.4, -20.0)
	await _release()
	await _wait(0.4)
	# Grow the brush with the radius slider in the options panel.
	var slider: HSlider = studio.hud.options._radius
	await _move(_slider_point(slider, slider.value), 1.0, 40.0)
	await _press()
	await _move(_slider_point(slider, 5.0), 0.7)
	await _release()
	await _wait(0.3)
	await _move(center + Vector2(0.55, -0.35) * r, 1.0, -40.0)
	await _press()
	await _move(center + Vector2(0.75, 0.1) * r, 1.2, 20.0)
	await _release()
	await _wait(1.4) # Strokes commit once idle; the counter in the top bar updates.

func _black_hole() -> void:
	overlay.caption("Black hole — swallow every cell inside the globe")
	await _move(_center(studio.hud.tools.tool_buttons[StructureTools.Tool.BLACK_HOLE]), 1.0, -50.0)
	await _click_until(func() -> bool: return studio.tools.tool == StructureTools.Tool.BLACK_HOLE)
	await _wait(0.2)
	var center := _screen(studio.camera.focus)
	var r := _screen_radius(4.8)
	await _move(center + Vector2(-0.7, 0.25) * r, 1.0, 40.0)
	await _press()
	await _move(center + Vector2(0.45, 0.05) * r, 1.6, 30.0)
	await _release()
	await _wait(1.4)

func _history() -> void:
	overlay.caption("Undo and redo — up to ten strokes")
	await _move(_center(studio.hud.tools.undo_button), 1.0, 60.0)
	await _click()
	await _wait(0.9)
	await _click()
	await _wait(0.9)
	await _move(_center(studio.hud.tools.redo_button), 0.4)
	await _click()
	await _wait(0.9)

func _orbit() -> void:
	overlay.caption("Drag to orbit · scroll to zoom")
	await _move(_center(studio.hud.tools.tool_buttons[StructureTools.Tool.NAVIGATE]), 0.7, 30.0)
	await _click_until(func() -> bool: return studio.tools.tool == StructureTools.Tool.NAVIGATE)
	await _move(Vector2(1000, 640), 0.8, -40.0)
	await _drag(Vector2(340, -60), 2.0)
	for notch in 2:
		await _wheel(MOUSE_BUTTON_WHEEL_UP)
		await _wait(0.15)
	await _wait(0.4)

func _load_castle() -> void:
	overlay.caption("Load any .octa file")
	var dialog: FileDialog = studio._load_dialog
	await _move(_center(_button(studio.hud.toolbar, "Load")), 1.0, 50.0)
	await _click_until(func() -> bool: return dialog.visible)
	await _wait(0.6)
	await _move(_file_item(dialog, "castle.octa"), 0.9, -30.0)
	await _click()
	await _wait(0.4)
	await _move(_center(dialog.get_ok_button(), dialog), 0.6)
	await _click_until(func() -> bool: return not dialog.visible)
	_speed_up()
	overlay.caption("Each build arrives in its own world")
	await _wait(0.8)
	await _move(Vector2(980, 620), 0.8, 40.0)
	await _drag(Vector2(-300, 20), 3.2)
	await _wait_assembled(0.4)

func _carve_castle() -> void:
	overlay.caption("A big black hole takes a bite out of a tower")
	await _move(_center(studio.hud.tools.tool_buttons[StructureTools.Tool.BLACK_HOLE]), 1.0, 50.0)
	await _click_until(func() -> bool: return studio.tools.tool == StructureTools.Tool.BLACK_HOLE)
	await _wait(0.2)
	var slider: HSlider = studio.hud.options._radius
	await _move(_slider_point(slider, slider.value), 1.0, -40.0)
	await _press()
	await _move(_slider_point(slider, 9.0), 0.8)
	await _release()
	await _wait(0.2)
	await _zoom_in(3, Vector2(960, 560))
	var tower := _on_build(_screen(_front_top() + Vector3.DOWN * 3.0))
	await _move(tower, 1.0, 50.0)
	await _press()
	await _move(tower + Vector2(0, 50), 1.1, 10.0)
	await _release()
	await _wait(1.6)

func _explore() -> void:
	overlay.caption("Explore the whole collection")
	await _move(_center(_button(studio.hud.toolbar, "Explore builds")), 1.0, 50.0)
	await _click_until(func() -> bool: return studio.hud.library.visible)
	await _wait(0.6)
	await _type("taj")
	await _wait(0.5)
	await _move(_center(_button(studio.hud.library, "Taj Mahal")), 0.8, -40.0)
	await _click_until(func() -> bool: return not studio.hud.library.visible)
	_speed_up()
	overlay.caption("")
	await _key(KEY_V) # Back to plain orbiting: the black hole is still selected.
	await _wait(0.5)
	await _move(Vector2(1040, 640), 0.8, 30.0)
	await _drag(Vector2(240, -20), 3.0)
	for notch in 2:
		await _wheel(MOUSE_BUTTON_WHEEL_UP)
		await _wait(0.2)
	await _wait_assembled(0.4)

func _paint_taj() -> void:
	overlay.caption("Brush in with a color — paint a red dot")
	await _move(_center(studio.hud.tools.tool_buttons[StructureTools.Tool.BRUSH_IN]), 1.0, 40.0)
	await _click_until(func() -> bool: return studio.tools.tool == StructureTools.Tool.BRUSH_IN)
	var options := studio.hud.options
	var red := _reddest(studio.renderer.structure.palette)
	if red >= ToolOptions.VISIBLE_COLORS and not options.more_expanded:
		await _move(_center(options.more_button), 0.8, -30.0)
		await _click()
		await _wait(0.2)
	await _move(_center(options.color_buttons[red]), 0.9, -40.0)
	await _click()
	await _wait(0.2)
	await _zoom_in(2, Vector2(960, 560))
	var spot := _on_build(_screen(_dome_spot()))
	await _move(spot, 1.1, 50.0)
	# Hold still: the brush fills its globe once, leaving a round dot.
	await _press()
	await _wait(1.2)
	await _release()
	await _move(Vector2(1500, 860), 1.0, 30.0)
	await _wait(1.6)

func _finish() -> void:
	overlay.fade(0.0, 1.0, 0.8)
	await _wait(0.9)
	_pipe.stdio.close()
	while OS.is_process_running(_pipe.pid):
		await create_timer(0.1).timeout
	var seconds := _frames / float(fps)
	print("Walkthrough: %d frames, %.1f s, %d skipped draws → %s" % [_frames, seconds, _skipped, ProjectSettings.globalize_path(out_path)])
	quit(0)

# --- Pointer, keys and timing ------------------------------------------------------

func _send(event: InputEvent) -> void:
	# Input keeps the polled button state; the viewport delivers the event to the studio.
	Input.parse_input_event(event)
	inner.push_input(event, true)

func _move_event(at: Vector2, relative: Vector2 = Vector2.ZERO) -> void:
	var event := InputEventMouseMotion.new()
	event.position = at
	event.global_position = at
	event.relative = relative
	event.velocity = relative * fps
	event.button_mask = Input.get_mouse_button_mask()
	_send(event)
	overlay.cross = studio.tools.tool != StructureTools.Tool.NAVIGATE and inner.gui_get_hovered_control() == null
	overlay.pointer = at

## Glide to a point with eased speed, bowing sideways by `arc` points like a hand would.
func _move(to: Vector2, seconds: float, arc: float = 0.0) -> void:
	var from := _pointer
	var steps := maxi(1, roundi(seconds * fps))
	var side := (to - from).orthogonal().normalized()
	for step in range(1, steps + 1):
		var t := float(step) / steps
		var eased := t * t * (3.0 - 2.0 * t)
		var at := from.lerp(to, eased) + side * arc * sin(PI * eased)
		_move_event(at, at - _pointer)
		_pointer = at
		await _frame()

func _button_event(index: MouseButton, pressed: bool) -> void:
	var event := InputEventMouseButton.new()
	event.position = _pointer
	event.global_position = _pointer
	event.button_index = index
	event.pressed = pressed
	event.factor = 1.0
	var mask := Input.get_mouse_button_mask()
	if index <= MOUSE_BUTTON_MIDDLE:
		var bit := 1 << (index - 1)
		mask = (mask | bit) if pressed else (mask & ~bit)
	event.button_mask = mask
	_send(event)

func _press() -> void:
	_button_event(MOUSE_BUTTON_LEFT, true)
	overlay.ripple(_pointer)
	await _frame()

func _release() -> void:
	_button_event(MOUSE_BUTTON_LEFT, false)
	await _frame()

func _click() -> void:
	await _press()
	await _frame()
	await _release()

## Click until the UI shows the click's effect, so one dropped click cannot derail the tour.
func _click_until(done: Callable, tries: int = 3) -> void:
	for attempt in tries:
		await _click()
		await _frame()
		if done.call():
			return
		push_warning("Click at %s had no effect; retrying" % _pointer)
		await _wait(0.3)
	push_error("Click at %s failed" % _pointer)

func _drag(by: Vector2, seconds: float) -> void:
	await _press()
	await _move(_pointer + by, seconds, by.length() * 0.06)
	await _release()

func _wheel(index: MouseButton) -> void:
	_button_event(index, true)
	_button_event(index, false)
	await _frame()

func _type(text: String) -> void:
	for character in text:
		for pressed in [true, false]:
			var event := InputEventKey.new()
			event.keycode = OS.find_keycode_from_string(character.to_upper())
			event.physical_keycode = event.keycode
			event.unicode = character.unicode_at(0)
			event.pressed = pressed
			_send(event)
			await _frame()
		await _wait(0.12)

func _wait(seconds: float) -> void:
	for step in roundi(seconds * fps):
		await _frame()

## Wait for the current assembly to finish, plus a short hold.
func _wait_assembled(hold: float = 0.3) -> void:
	while studio.renderer.timeline.is_running():
		await _frame()
	await _wait(hold)

## Primitives fill the default framing; pull the camera back to show them in their setting.
func _widen(factor: float = 1.3) -> void:
	studio.camera.distance *= factor
	studio.camera._update_transform()

## Scroll the wheel with the pointer over the view; the camera zooms towards its focus.
func _zoom_in(notches: int, over: Vector2) -> void:
	await _move(over, 0.7, 30.0)
	for notch in notches:
		await _wheel(MOUSE_BUTTON_WHEEL_UP)
		await _wait(0.18)

func _key(code: Key) -> void:
	for pressed in [true, false]:
		var event := InputEventKey.new()
		event.keycode = code
		event.physical_keycode = code
		event.pressed = pressed
		_send(event)
		await _frame()

## Large builds take up to 24 s to assemble; play them faster to fit the minute.
func _speed_up() -> void:
	var renderer := studio.renderer
	if fast <= 1.0 or renderer.has_meta("walkthrough_fast"):
		return
	renderer.set_meta("walkthrough_fast", true)
	renderer.assembly_progress.connect(func(_fraction: float, _duration: float) -> void:
		if renderer.timeline.is_running():
			renderer.timeline.advance((fast - 1.0) / fps))

## One recorded frame: wait for the draw, then stream the overlay viewport to ffmpeg.
func _frame() -> void:
	await RenderingServer.frame_post_draw
	var drawn := Engine.get_frames_drawn()
	if _last_drawn >= 0 and drawn > _last_drawn + 1:
		_skipped += drawn - _last_drawn - 1
	_last_drawn = drawn
	var image := outer.get_texture().get_image()
	if _pipe.is_empty():
		_open_encoder(image)
	_pipe.stdio.store_buffer(image.get_data())
	_frames += 1
	if _frames % (fps * 5) == 0:
		print("Walkthrough: %d s recorded" % (_frames / fps))

func _open_encoder(image: Image) -> void:
	var format := "rgba" if image.get_format() == Image.FORMAT_RGBA8 else "rgb24"
	if image.get_format() not in [Image.FORMAT_RGBA8, Image.FORMAT_RGB8]:
		push_error("Unexpected viewport image format %d" % image.get_format())
		quit(1)
	var ffmpeg: String = ""
	for candidate: String in FFMPEG:
		if FileAccess.file_exists(candidate):
			ffmpeg = candidate
			break
	var path := ProjectSettings.globalize_path(out_path)
	DirAccess.make_dir_recursive_absolute(path.get_base_dir())
	# Low priority and a few threads: the recording should not take over the machine.
	_pipe = OS.execute_with_pipe("/usr/bin/nice", ["-n", "15", ffmpeg if ffmpeg != "" else "ffmpeg",
		"-y", "-loglevel", "error", "-f", "rawvideo", "-pix_fmt", format,
		"-s", "%dx%d" % [image.get_width(), image.get_height()], "-r", str(fps), "-i", "-",
		"-vf", "scale=%d:%d:flags=lanczos" % [SIZE.x, SIZE.y], "-c:v", "libx264",
		"-preset", "slow", "-crf", "17", "-pix_fmt", "yuv420p", "-threads", "4",
		"-movflags", "+faststart", path])
	if _pipe.is_empty():
		push_error("Could not start ffmpeg")
		quit(1)

# --- Locating controls --------------------------------------------------------------

## Center of a control in the studio viewport's layout points, including one inside an
## embedded window such as the file dialog.
func _center(control: Control, window: Window = null) -> Vector2:
	var center := control.get_global_rect().get_center()
	return center + (Vector2(window.position) if window != null else Vector2.ZERO)

func _button(parent: Node, text: String) -> Button:
	for child in parent.find_children("*", "Button", true, false):
		var button := child as Button
		if button.is_visible_in_tree() and button.text.strip_edges() == text:
			return button
	push_error("No button: " + text)
	return null

## PopupMenu has no item rect API: split its content height evenly between the items.
func _popup_item(popup: PopupMenu, index: int) -> Vector2:
	var panel := popup.get_theme_stylebox("panel")
	var top := panel.get_margin(SIDE_TOP)
	var height := (popup.size.y - top - panel.get_margin(SIDE_BOTTOM)) / popup.item_count
	return Vector2(popup.position) + Vector2(popup.size.x * 0.35, top + height * (index + 0.5))

func _file_item(dialog: FileDialog, file: String) -> Vector2:
	for node in _descendants(dialog):
		if node is Tree and node.is_visible_in_tree():
			var tree := node as Tree
			var item := tree.get_root().get_first_child() if tree.get_root() != null else null
			while item != null:
				if item.get_text(0) == file:
					tree.scroll_to_item(item)
					var rect := tree.get_item_area_rect(item)
					return _center(tree, dialog) - tree.size * 0.5 + rect.position + Vector2(rect.size.x * 0.25, rect.size.y * 0.5) - tree.get_scroll()
				item = item.get_next()
		elif node is ItemList and node.is_visible_in_tree():
			var list := node as ItemList
			for index in list.item_count:
				if list.get_item_text(index) == file:
					list.ensure_current_is_visible()
					var rect := list.get_item_rect(index)
					return _center(list, dialog) - list.size * 0.5 + rect.get_center() - Vector2(0, list.get_v_scroll_bar().value)
	push_error("File not listed: " + file)
	return _pointer

func _descendants(node: Node) -> Array[Node]:
	var found: Array[Node] = []
	for child in node.get_children(true):
		found.append(child)
		found.append_array(_descendants(child))
	return found

func _slider_point(slider: HSlider, value: float) -> Vector2:
	var rect := slider.get_global_rect()
	var fraction := inverse_lerp(slider.min_value, slider.max_value, value)
	var grabber := slider.get_theme_icon("grabber").get_width()
	return Vector2(rect.position.x + grabber * 0.5 + fraction * (rect.size.x - grabber), rect.get_center().y)

## A structure-space point on the camera's side of the build, shown in world space.
func _world(local: Vector3) -> Vector3:
	return studio.renderer.global_transform * local

## The highest cell in the part of the build facing the camera: a front tower's tip.
func _front_top() -> Vector3:
	var data := studio.renderer.structure
	var center := data.bounds().get_center()
	var toward := studio.renderer.global_transform.affine_inverse() * studio.camera.global_position - center
	toward.y = 0.0
	toward = toward.normalized()
	var points := PackedVector3Array()
	var reach := 0.0
	for index in data.cell_count():
		var point := OctaLattice.to_world(data.coordinate(index), data.cell_span)
		points.append(point)
		reach = maxf(reach, (point - center).dot(toward))
	var best := Vector3(0, -INF, 0)
	for point in points:
		if (point - center).dot(toward) > reach * 0.45 and point.y > best.y:
			best = point
	return _world(best)

## A point on the side of the highest dome, facing the camera.
func _dome_spot() -> Vector3:
	var data := studio.renderer.structure
	var top := Vector3(0, -INF, 0)
	for index in data.cell_count():
		var point := OctaLattice.to_world(data.coordinate(index), data.cell_span)
		if point.y > top.y:
			top = point
	var world := _world(top)
	var toward := studio.camera.global_position - world
	toward.y = 0.0
	return world + Vector3.DOWN * data.bounds().size.y * 0.3

## Nudge a screen point down until the pointer ray hits a cell, so a stroke starts on the build.
func _on_build(at: Vector2) -> Vector2:
	var tools := studio.tools
	tools._ensure_editor()
	var camera := studio.camera
	for attempt in 40:
		var origin := tools.to_local(camera.project_ray_origin(at))
		var direction := (tools.global_basis.inverse() * camera.project_ray_normal(at)).normalized()
		if tools.editor.pick(origin, direction).get("hit", false):
			return at
		at.y += 6.0
	push_warning("No cell under the pointer near %s" % at)
	return at

## Palette index of the most saturated red, or 0.
func _reddest(palette: PackedColorArray) -> int:
	var best := 0
	var best_score := -INF
	for index in mini(palette.size(), ToolOptions.COLUMNS * 4):
		var color := palette[index]
		var score := color.r - maxf(color.g, color.b)
		if score > best_score:
			best_score = score
			best = index
	return best

func _screen(world: Vector3) -> Vector2:
	return studio.camera.unproject_position(world)

## On-screen size of a structure-space length around the camera focus.
func _screen_radius(world_length: float) -> float:
	var camera := studio.camera
	return _screen(camera.focus).distance_to(_screen(camera.focus + camera.global_basis.x * world_length))

func _parse_options() -> Dictionary:
	var result := {}
	for arg in OS.get_cmdline_user_args():
		if not arg.begins_with("--"):
			continue
		var pair := arg.substr(2).split("=", true, 1)
		result[pair[0]] = pair[1] if pair.size() > 1 else "true"
	return result

## Drawn pointer, click ripples, a caption pill and fades, over the recorded studio.
class WalkthroughOverlay extends Control:
	const RIPPLE_TIME := 0.45
	var pointer := Vector2.ZERO:
		set(value):
			pointer = value
			queue_redraw()
	var cross: bool = false # Crosshair while a brush tool is over the 3D view.
	var _ripples: Array[Dictionary] = []
	var _clock: float = 0.0
	var _caption: PanelContainer
	var _text: Label
	var _shade: ColorRect

	func _ready() -> void:
		set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
		mouse_filter = Control.MOUSE_FILTER_IGNORE
		_caption = PanelContainer.new()
		_caption.add_theme_stylebox_override("panel", StudioTheme.panel(Color(StudioTheme.INK, 0.9), Color("37677c"), 22, 26, 11))
		_caption.set_anchors_and_offsets_preset(Control.PRESET_CENTER_BOTTOM)
		_caption.grow_horizontal = Control.GROW_DIRECTION_BOTH
		_caption.grow_vertical = Control.GROW_DIRECTION_BEGIN
		_caption.offset_top = -64
		_caption.offset_bottom = -64
		_caption.modulate.a = 0.0
		_caption.mouse_filter = Control.MOUSE_FILTER_IGNORE
		_text = StudioTheme.label("", 19)
		_caption.add_child(_text)
		add_child(_caption)
		_shade = ColorRect.new()
		_shade.color = Color.BLACK
		_shade.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
		_shade.mouse_filter = Control.MOUSE_FILTER_IGNORE
		add_child(_shade)

	func caption(text: String) -> void:
		var tween := create_tween()
		if _caption.modulate.a > 0.0:
			tween.tween_property(_caption, "modulate:a", 0.0, 0.18)
		if text.is_empty():
			return
		tween.tween_callback(func() -> void:
			_text.text = text
			_caption.reset_size()
			_caption.offset_left = -_caption.get_combined_minimum_size().x * 0.5
			_caption.offset_right = -_caption.offset_left)
		tween.tween_property(_caption, "modulate:a", 1.0, 0.25)

	func fade(from: float, to: float, seconds: float) -> void:
		_shade.color.a = from
		create_tween().tween_property(_shade, "color:a", to, seconds)

	func ripple(at: Vector2) -> void:
		_ripples.append({"at": at, "born": _clock})

	func _process(delta: float) -> void:
		_clock += delta
		_ripples = _ripples.filter(func(r: Dictionary) -> bool: return _clock - r.born < RIPPLE_TIME)
		queue_redraw()

	func _draw() -> void:
		for r: Dictionary in _ripples:
			var t: float = (_clock - r.born) / RIPPLE_TIME
			var color := Color(StudioTheme.ACCENT, 0.8 * (1.0 - t))
			draw_arc(r.at, lerpf(6.0, 30.0, sqrt(t)), 0.0, TAU, 48, color, 2.5, true)
			draw_circle(r.at, lerpf(9.0, 4.0, t), Color(StudioTheme.ACCENT, 0.35 * (1.0 - t)))
		if cross:
			for d: Vector2 in [Vector2.RIGHT, Vector2.LEFT, Vector2.UP, Vector2.DOWN]:
				draw_line(pointer + d * 4.0, pointer + d * 13.0, Color(0, 0, 0, 0.6), 4.0, true)
				draw_line(pointer + d * 4.0, pointer + d * 12.0, Color.WHITE, 2.0, true)
			return
		# A classic arrow pointer, 24 points tall, tip at the pointer position.
		var arrow := PackedVector2Array([Vector2(0, 0), Vector2(0, 21), Vector2(5, 16.5),
			Vector2(8.6, 24.5), Vector2(12, 23), Vector2(8.5, 15.2), Vector2(15, 15)])
		var shadow := PackedVector2Array()
		var shape := PackedVector2Array()
		for point in arrow:
			shadow.append(pointer + point + Vector2(1.5, 2.0))
			shape.append(pointer + point)
		draw_colored_polygon(shadow, Color(0, 0, 0, 0.3))
		draw_colored_polygon(shape, Color.WHITE)
		shape.append(shape[0])
		draw_polyline(shape, Color.BLACK, 1.4, true)
