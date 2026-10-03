class_name BrushGizmo
extends Node3D
## The brush cursor: a holographic globe (brush in, black hole), the globe clipped to a dome
## with its cut plane outlined (scoop) and, for erasing tools, a black hole at the brush center.

const BRUSH_IN_COLOR := Color("b9e9cb")
const BLACK_HOLE_COLOR := Color("b7a6ff")
const SCOOP_COLOR := Color("ffc58a")
var _globe := MeshInstance3D.new()
var _rim := MeshInstance3D.new()
var _hole := MeshInstance3D.new()
var _shell_material := ShaderMaterial.new()
var _rim_material := ShaderMaterial.new()
var _hole_material := ShaderMaterial.new()
var _strength: float = 0.3

func _init() -> void:
	var brush := preload("res://shaders/brush.gdshader")
	_shell_material.shader = brush
	_rim_material.shader = brush
	_rim_material.set_shader_parameter("fill", 0.7)
	_rim_material.set_shader_parameter("rim", 0.3)
	_rim_material.set_shader_parameter("lines", 0.0)
	_hole_material.shader = preload("res://shaders/black_hole.gdshader")
	_hole_material.render_priority = 1
	var sphere := SphereMesh.new()
	sphere.radial_segments = 48
	sphere.rings = 24
	_globe.mesh = sphere
	var ring := TorusMesh.new()
	ring.inner_radius = 0.975
	ring.outer_radius = 1.0
	ring.rings = 64
	_rim.mesh = ring
	var quad := QuadMesh.new()
	quad.size = Vector2(2, 2)
	_hole.mesh = quad
	for part: MeshInstance3D in [_globe, _rim, _hole]:
		part.cast_shadow = GeometryInstance3D.SHADOW_CASTING_SETTING_OFF
		add_child(part)
	_globe.material_override = _shell_material
	_rim.material_override = _rim_material
	_hole.material_override = _hole_material
	visible = false

## Place the brush. axis is the scoop's dome direction (zero for a full globe); center is
## the globe's center or the scoop's cut-plane center. active brightens it during a stroke.
func show_brush(tool: StructureTools.Tool, center: Vector3, radius: float, axis: Vector3, active: bool, delta: float) -> void:
	visible = tool != StructureTools.Tool.NAVIGATE
	if not visible:
		return
	var scoop := tool == StructureTools.Tool.SCOOP
	var color := SCOOP_COLOR if scoop else (BLACK_HOLE_COLOR if tool == StructureTools.Tool.BLACK_HOLE else BRUSH_IN_COLOR)
	_shell_material.set_shader_parameter("tint", color)
	_shell_material.set_shader_parameter("intensity", 1.6 if active else 1.0)
	_rim_material.set_shader_parameter("tint", color)
	var up := axis if scoop else Vector3.UP
	transform = Transform3D(Basis(Quaternion(Vector3.UP, up.normalized())).scaled(Vector3.ONE * radius), center)
	_shell_material.set_shader_parameter("dome", scoop)
	_rim.visible = scoop
	_hole.visible = tool != StructureTools.Tool.BRUSH_IN
	_strength = move_toward(_strength, 1.0 if active else 0.3, delta * 3.0)
	_hole_material.set_shader_parameter("strength", _strength)
	_hole.scale = Vector3.ONE * lerpf(0.42, 0.58, _strength)
