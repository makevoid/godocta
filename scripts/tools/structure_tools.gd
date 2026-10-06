class_name StructureTools
extends Node3D
## Brush tools that edit the displayed structure in place.
##  · Brush in: empty slots under the brush that touch existing cells (or the ground)
##    fill with new cells, which fly in and snap to the cells that attract them. The
##    brush freezes its distance from the camera for the stroke, so a blob fills the
##    globe and stops instead of creeping towards the viewer; dragging paints sideways
##    at that depth. When nothing is left to fill, `brush_full` reports the stroke is
##    done: release and click again to lay the next layer.
##    Cells take the brush color, or their neighbors' color when it is auto (-1).
##  · Black hole: a globe that swallows every cell within its radius. Its depth stays
##    fixed for the stroke, so clearing the foreground never retargets the background.
##  · Scoop: a half globe, flat or tilted 45° around the screen's X and Y axes, that
##    shaves cells off the surface. Its cut plane is fixed for a stroke, so holding or
##    dragging it planes the build instead of digging deeper.
## Left click/drag applies a tool; right drag orbits, Alt/Shift + drag keep the camera's
## orbit and pan. Edits are pending (cells hidden or drawn by this node) until the stroke
## has been idle for COMMIT_DELAY, then one rebuilt structure replaces the old one.

signal committed(structure: OctaStructure)
## The brush in stroke has nothing left to fill at its current position.
signal brush_full

enum Tool { NAVIGATE, BRUSH_IN, BLACK_HOLE, SCOOP }
const NAMES: Array[String] = ["Orbit", "Brush in", "Black hole", "Scoop"]
const MIN_RADIUS := 0.6 # Brush radius limits, in cell spans.
const MAX_RADIUS := 12.0
const TICK := 0.05 # Seconds between applications during a stroke.
const ARRIVAL_TIME := 0.42
const SWALLOW_TIME := 0.55
const COMMIT_DELAY := 1.2
const MAX_SWALLOWED := 2500 # Larger bites vanish without the flight.

var tool: Tool = Tool.NAVIGATE
## Starting radius per tool, in cell spans. Each tool then keeps the last size it was
## given, so switching between them restores the size that suited each one.
const DEFAULT_RADII: Dictionary = {
	Tool.NAVIGATE: 3.0, Tool.BRUSH_IN: 3.0, Tool.BLACK_HOLE: 3.0, Tool.SCOOP: 2.0,
}
## Brush radius in cell spans, for the selected tool.
var radius: float = float(DEFAULT_RADII[Tool.BRUSH_IN])
var _radii: Dictionary = DEFAULT_RADII.duplicate()
## Palette index the brush paints, or -1 for auto: each new cell adopts its neighbors' color.
var brush_color: int = -1
## Scoop tilt in degrees around the screen X and Y axes: -45, 0 or 45 each.
var tilt: Vector2i = Vector2i.ZERO
var renderer: StructureRenderer
var camera: Camera3D
var editor: OctaEditor
var gizmo := BrushGizmo.new()
var _pending := MultiMeshInstance3D.new() # Added cells until the next commit.
var _vortex := MultiMeshInstance3D.new() # Removed cells falling into the black hole.
var _arrivals: Dictionary = {} # Vector3i → {from, to, born, custom}
var _swallowed: Array[Dictionary] = []
var _pending_dirty: bool = false
var _stroking: bool = false
var _stroke_axis: Vector3 = Vector3.UP
var _stroke_plane: Variant = null # Plane, set by the stroke's first hit.
var _saturated: bool = false # The brush found no free slot on its last tick.
var _tick: float = 0.0
var _commit_timer: float = 0.0
var _clock: float = 0.0
var _fallback_color: int = 0
var _random := RandomNumberGenerator.new()

func _ready() -> void:
	for layer: MultiMeshInstance3D in [_pending, _vortex]:
		var instances := MultiMesh.new()
		instances.transform_format = MultiMesh.TRANSFORM_3D
		instances.use_custom_data = true
		instances.mesh = TruncatedOctahedronMesh.shared_mesh()
		layer.multimesh = instances
		add_child(layer)
	add_child(gizmo)
	_random.seed = 7

func set_tool(value: Tool) -> void:
	tool = value
	radius = float(_radii[tool])
	if tool == Tool.NAVIGATE:
		_end_stroke()
		gizmo.visible = false
	Input.set_default_cursor_shape(Input.CURSOR_ARROW if tool == Tool.NAVIGATE else Input.CURSOR_CROSS)

func set_radius(cells: float) -> void:
	radius = clampf(cells, MIN_RADIUS, MAX_RADIUS)
	_radii[tool] = radius

## Pick the painted color: a palette index, or -1 to keep adopting neighbors' colors.
## Indices outside the displayed palette, and any color on an uncolored build, fall to auto.
func set_brush_color(index: int) -> void:
	var palette := renderer.structure.palette if renderer != null and renderer.structure != null else PackedColorArray()
	brush_color = index if index >= 0 and index < palette.size() else -1
	gizmo.brush_tint = palette[brush_color] if brush_color >= 0 else null

## Brush radius in structure-space units.
func world_radius() -> float:
	return radius * (editor.span() if editor != null else renderer.structure.cell_span)

## Discard pending edits, e.g. when another structure is presented.
func reset() -> void:
	_stroking = false
	_stroke_plane = null
	_saturated = false
	_commit_timer = 0.0
	_arrivals.clear()
	_swallowed.clear()
	_pending_dirty = true
	editor = null
	renderer.set_editing(false)

## Rebuild the structure from pending edits now: before saving or replaying.
func commit() -> void:
	_commit_timer = 0.0
	if editor == null or not editor.has_changes() or editor.structure != renderer.structure:
		return
	var data := editor.build()
	if data == null:
		return
	var layer := editor.ground_layer
	renderer.display(data, false, false)
	renderer.set_editing(false)
	editor = OctaEditor.new(data, layer)
	_arrivals.clear()
	_pending_dirty = true
	_rebuild_pending()
	committed.emit(data)

## Apply the current tool once at a structure-space point; axis orients the scoop.
## Mouse strokes call this every TICK; tests can call it directly.
func apply(center: Vector3, axis: Vector3 = Vector3.UP, toward: Vector3 = Vector3.UP) -> void:
	if not _ensure_editor():
		return
	renderer.set_editing(true)
	match tool:
		Tool.BRUSH_IN:
			_brush_in(center, toward)
		Tool.BLACK_HOLE:
			_swallow(editor.cells_within(center, world_radius()), center)
		Tool.SCOOP:
			_swallow(editor.cells_within(center, world_radius(), axis), center)
	_commit_timer = COMMIT_DELAY

## The scoop's dome direction: up, tipped by the tilt around the camera's screen axes.
func scoop_axis() -> Vector3:
	var right := camera.global_basis.x
	right.y = 0.0
	var toward := camera.global_basis.z
	toward.y = 0.0
	var axis := Vector3.UP
	if right.length_squared() > 1e-6:
		axis += right.normalized() * tan(deg_to_rad(tilt.x))
	if toward.length_squared() > 1e-6:
		axis += toward.normalized() * tan(deg_to_rad(tilt.y))
	return axis.normalized()

## How far below the hit the scoop's cut plane sits: about one layer of cells.
func scoop_depth() -> float:
	return editor.span() * 0.75

func _unhandled_input(event: InputEvent) -> void:
	if tool == Tool.NAVIGATE:
		return
	if event is InputEventMouseButton and event.button_index == MOUSE_BUTTON_LEFT:
		if event.pressed and not event.shift_pressed and not event.alt_pressed:
			_stroking = true
			_stroke_plane = null
			_saturated = false
			_stroke_axis = scoop_axis()
			_tick = 0.0
			get_viewport().set_input_as_handled()
		elif not event.pressed and _stroking:
			_end_stroke()
			get_viewport().set_input_as_handled()
	elif event is InputEventMouseMotion and _stroking:
		get_viewport().set_input_as_handled()

func _process(delta: float) -> void:
	_clock += delta
	position = renderer.position
	# Edit layers share the renderer's materials, including X-ray and their uniforms.
	for layer: MultiMeshInstance3D in [_pending, _vortex]:
		layer.material_override = renderer.material_override
		layer.cast_shadow = renderer.cast_shadow
	# Poll the release too: it may happen over the HUD.
	if _stroking and not Input.is_mouse_button_pressed(MOUSE_BUTTON_LEFT):
		_end_stroke()
	if tool != Tool.NAVIGATE:
		_update_brush(delta)
	_animate()
	if _commit_timer > 0.0 and not _stroking:
		_commit_timer -= delta
		if _commit_timer <= 0.0:
			commit()

func _end_stroke() -> void:
	_stroking = false
	# The next stroke picks its own depth and cut plane from its first hit.
	_stroke_plane = null
	_saturated = false

func _ensure_editor() -> bool:
	if renderer.structure == null:
		return false
	if editor == null or editor.structure != renderer.structure:
		# A new structure was displayed outside the tools: drop stale edit effects.
		var replaced := editor != null
		reset()
		editor = OctaEditor.new(renderer.structure)
		if replaced:
			_rebuild_pending()
	return true

func _update_brush(delta: float) -> void:
	if not _ensure_editor() or camera == null:
		return
	var viewport := get_viewport()
	if not _stroking and viewport.gui_get_hovered_control() != null:
		gizmo.visible = false
		return
	var mouse := viewport.get_mouse_position()
	var origin := to_local(camera.project_ray_origin(mouse))
	var direction := (global_basis.inverse() * camera.project_ray_normal(mouse)).normalized()
	_update_brush_ray(origin, direction, delta)

## Position and apply the brush along a mouse ray in structure space.
func _update_brush_ray(origin: Vector3, direction: Vector3, delta: float) -> void:
	var hit := editor.pick(origin, direction)
	if hit.get("hit", false):
		_fallback_color = maxi(editor.color_of(hit.cell), 0)
	var center: Vector3
	var axis := _stroke_axis if _stroking else scoop_axis()
	if tool == Tool.SCOOP:
		if _stroking and _stroke_plane == null and not hit.is_empty():
			_stroke_plane = Plane(axis, hit.point - axis * scoop_depth())
		# Glide along the stroke's cut plane, so the scoop never digs below it.
		var on_plane: Variant = (_stroke_plane as Plane).intersects_ray(origin, direction) if _stroke_plane != null else null
		if on_plane != null:
			center = on_plane
		elif not hit.is_empty():
			center = hit.point - axis * scoop_depth()
		else:
			gizmo.visible = false
			return
	elif tool == Tool.BRUSH_IN or tool == Tool.BLACK_HOLE:
		# Both globes stay at the first hit's depth: adding cannot creep towards the
		# viewer, and removing cannot jump to newly exposed cells in the background.
		if _stroking and _stroke_plane == null and not hit.is_empty():
			_stroke_plane = Plane(-direction, hit.point)
		var at_depth: Variant = (_stroke_plane as Plane).intersects_ray(origin, direction) if _stroke_plane != null else null
		if at_depth != null:
			center = at_depth
		elif _stroke_plane != null:
			# If the camera turns away from the plane, wait until the ray meets it again.
			gizmo.visible = false
			return
		elif not hit.is_empty():
			center = hit.point
		else:
			gizmo.visible = false
			return
	elif hit.is_empty():
		gizmo.visible = false
		return
	else:
		center = hit.point
	gizmo.show_brush(tool, center, world_radius(), axis, _stroking, delta, _stroking and _saturated)
	if _stroking:
		_tick -= delta
		if _tick <= 0.0:
			_tick = TICK
			apply(center, axis, -direction)

func _brush_in(center: Vector3, toward: Vector3) -> void:
	var span := editor.span()
	var count := maxi(1, roundi(radius * 1.5))
	var filled := 0
	for cell in editor.free_slots(center, world_radius(), count):
		var color := editor.color_choice(cell, brush_color, _fallback_color)
		if not editor.add(cell, color):
			continue
		var target := OctaLattice.to_world(cell, span)
		# Spawn in front of the surface, scattered around the brush, then get pulled in.
		var scatter := Vector3(_random.randf_range(-1, 1), _random.randf_range(-1, 1), _random.randf_range(-1, 1))
		var from := target + toward.normalized() * world_radius() * 1.3 + scatter * world_radius() * 0.6
		_arrivals[cell] = {"from": from, "to": target, "born": _clock, "custom": _custom(target, color)}
		filled += 1
	_pending_dirty = true
	# The globe is full: report it once, and again only after the brush moves on.
	if filled > 0:
		_saturated = false
	elif not _saturated:
		_saturated = true
		brush_full.emit()

func _swallow(cells: Array[Vector3i], center: Vector3) -> void:
	var span := editor.span()
	for cell in cells:
		var index := editor.original_index(cell)
		var color := editor.color_of(cell)
		if not editor.remove(cell):
			continue
		if index >= 0:
			renderer.hide_cell(index)
		elif _arrivals.erase(cell):
			_pending_dirty = true
		if _swallowed.size() < MAX_SWALLOWED:
			var start := OctaLattice.to_world(cell, span)
			_swallowed.append({"from": start, "to": center, "born": _clock,
				"custom": _custom(start, color), "spin": _random.randf_range(0.6, 1.4)})

func _custom(center: Vector3, color: int) -> PackedFloat32Array:
	# Rank 0 is always settled; seed only shapes the assembly flight.
	return PackedFloat32Array([0.0, 0.0, renderer.normalized_height(center.y), float(color)])

func _animate() -> void:
	var flying := false
	for arrival: Dictionary in _arrivals.values():
		if _clock - arrival.born < ARRIVAL_TIME:
			flying = true
			break
	if flying or _pending_dirty:
		_rebuild_pending()
	if not _swallowed.is_empty() or _vortex.multimesh.instance_count > 0:
		_rebuild_vortex()

## New cells accelerate towards their slot, spinning and growing as they are pulled in.
func _rebuild_pending() -> void:
	_pending_dirty = false
	var span := editor.span() if editor != null else 1.0
	var buffer := PackedFloat32Array()
	buffer.resize(_arrivals.size() * 16)
	var index := 0
	for arrival: Dictionary in _arrivals.values():
		var t := clampf((_clock - float(arrival.born)) / ARRIVAL_TIME, 0.0, 1.0)
		var pull := t * t * t
		var place: Vector3 = (arrival.from as Vector3).lerp(arrival.to, pull)
		var turn := Basis(Vector3(0.3, 1, 0.2).normalized(), (1.0 - pull) * 4.0).scaled(Vector3.ONE * span * lerpf(0.45, 1.0, sqrt(t)))
		_write(buffer, index, Transform3D(turn, place), arrival.custom)
		index += 1
	_upload(_pending.multimesh, buffer)

## Swallowed cells spiral into the hole, accelerating, shrinking and stretching.
func _rebuild_vortex() -> void:
	var span := editor.span() if editor != null else 1.0
	var alive: Array[Dictionary] = []
	for cell in _swallowed:
		if _clock - cell.born < SWALLOW_TIME:
			alive.append(cell)
	_swallowed = alive
	var buffer := PackedFloat32Array()
	buffer.resize(alive.size() * 16)
	for index in alive.size():
		var cell := alive[index]
		var t: float = (_clock - cell.born) / SWALLOW_TIME
		var fall := t * t
		var to: Vector3 = cell.to
		var offset: Vector3 = cell.from - to
		var swirl: Vector3 = offset.rotated(Vector3.UP, fall * TAU * 0.75 * cell.spin) * (1.0 - fall)
		var size := span * pow(1.0 - fall, 0.8)
		var spin := Basis(Vector3(0.4, 1, -0.3).normalized(), fall * 9.0 * float(cell.spin))
		_write(buffer, index, Transform3D(spin.scaled(Vector3.ONE * size), to + swirl), cell.custom)
	_upload(_vortex.multimesh, buffer)

func _upload(instances: MultiMesh, buffer: PackedFloat32Array) -> void:
	var count := buffer.size() / 16
	if instances.instance_count != count:
		instances.instance_count = count
	if count > 0:
		instances.buffer = buffer

## Write one instance in StructureRenderer's 16-float layout.
static func _write(buffer: PackedFloat32Array, index: int, xf: Transform3D, custom: PackedFloat32Array) -> void:
	var offset := index * 16
	for row in 3:
		buffer[offset + row * 4] = xf.basis.x[row]
		buffer[offset + row * 4 + 1] = xf.basis.y[row]
		buffer[offset + row * 4 + 2] = xf.basis.z[row]
		buffer[offset + row * 4 + 3] = xf.origin[row]
	for channel in 4:
		buffer[offset + 12 + channel] = custom[channel]
