class_name OrbitCamera
extends Camera3D

var focus: Vector3 = Vector3(0, 4.0, 0)
var yaw: float = 0.68
var pitch: float = 0.48
var distance: float = 27.0
var auto_orbit: bool = false
var _dragging: bool = false
var _pan: bool = false
var _default_focus: Vector3 = Vector3(0, 4.0, 0)
var _default_distance: float = 27.0

func _ready() -> void:
	fov = 39.0
	near = 0.05
	far = 1000.0
	h_offset = -2.2
	_update_transform()

func frame_structure(box: AABB) -> void:
	_default_focus = Vector3(0, box.size.y * 0.46, 0)
	_default_distance = maxf(box.size.length() * 1.6, 8.0)
	far = maxf(_default_distance * 20.0, 1000.0)
	near = maxf(_default_distance * 0.0001, 0.01)
	reset_view()

func reset_view() -> void:
	focus = _default_focus
	distance = _default_distance
	yaw = 0.68
	pitch = 0.48
	_update_transform()

func _process(delta: float) -> void:
	# Poll releases too: releasing over the toolbar must not leave a stuck drag.
	if _dragging and not Input.is_mouse_button_pressed(MOUSE_BUTTON_LEFT) and not Input.is_mouse_button_pressed(MOUSE_BUTTON_MIDDLE) and not Input.is_mouse_button_pressed(MOUSE_BUTTON_RIGHT):
		_dragging = false
	if auto_orbit and not _dragging:
		yaw += delta * 0.13
		_update_transform()

func _unhandled_input(event: InputEvent) -> void:
	if event is InputEventMouseButton:
		if event.button_index in [MOUSE_BUTTON_LEFT, MOUSE_BUTTON_MIDDLE, MOUSE_BUTTON_RIGHT]:
			_dragging = event.pressed
			_pan = event.button_index == MOUSE_BUTTON_MIDDLE or event.shift_pressed
		if event.pressed and event.button_index == MOUSE_BUTTON_WHEEL_UP:
			distance = maxf(distance * 0.9, _default_distance * 0.2)
		if event.pressed and event.button_index == MOUSE_BUTTON_WHEEL_DOWN:
			distance = minf(distance * 1.1, _default_distance * 4.0)
		_update_transform()
	elif event is InputEventMouseMotion and _dragging:
		if _pan:
			focus += (-basis.x * event.relative.x + basis.y * event.relative.y) * distance * 0.001
		else:
			yaw -= event.relative.x * 0.006
			pitch = clampf(pitch + event.relative.y * 0.005, 0.08, 1.35)
		_update_transform()
	elif event is InputEventMagnifyGesture:
		distance = clampf(distance / event.factor, _default_distance * 0.2, _default_distance * 4.0)
		_update_transform()
	elif event is InputEventPanGesture:
		yaw -= event.delta.x * 0.035
		pitch = clampf(pitch + event.delta.y * 0.035, 0.08, 1.35)
		_update_transform()

func _update_transform() -> void:
	position = focus + Vector3(sin(yaw) * cos(pitch), sin(pitch), cos(yaw) * cos(pitch)) * distance
	look_at(focus)
