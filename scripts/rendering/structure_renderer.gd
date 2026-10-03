class_name StructureRenderer
extends MultiMeshInstance3D
## One render node and one shared mesh, regardless of the cell count.
## The 64-byte per-instance GPU buffer is derived from the compact structure.

signal assembly_progress(fraction: float, duration: float)
signal assembly_finished

var structure: OctaStructure
var timeline := AssemblyTimeline.new()
var xray: bool = false
var _solid := ShaderMaterial.new()
var _xray := ShaderMaterial.new()
var _gap: float = 0.045

func _init() -> void:
	_solid.shader = preload("res://shaders/cell.gdshader")
	_xray.shader = preload("res://shaders/cell_xray.gdshader")
	material_override = _solid
	set_process(false)

func display(data: OctaStructure, animate: bool = true) -> void:
	structure = data
	timeline.configure(data.cell_count())
	_set_parameter("release_window", timeline.release_window)
	_set_parameter("settle_time", timeline.settle_time)
	var instances := MultiMesh.new()
	instances.transform_format = MultiMesh.TRANSFORM_3D
	instances.use_custom_data = true
	instances.mesh = TruncatedOctahedronMesh.shared_mesh()
	instances.instance_count = data.cell_count()
	var box := data.bounds()
	# Big builds drop cells from proportionally higher so the motion stays visible.
	_set_parameter("flight_scale", clampf(sqrt(box.get_longest_axis_size() / 10.0), 1.0, 3.5))
	# Recenter for inspection while preserving original file coordinates.
	position = Vector3(-box.get_center().x, -box.position.y + 0.06, -box.get_center().z)
	var buffer := PackedFloat32Array()
	buffer.resize(data.cell_count() * 16)
	for index in data.cell_count():
		var center := OctaLattice.to_world(data.coordinate(index), data.cell_span)
		var offset := index * 16
		buffer[offset] = data.cell_span
		buffer[offset + 3] = center.x
		buffer[offset + 5] = data.cell_span
		buffer[offset + 7] = center.y
		buffer[offset + 10] = data.cell_span
		buffer[offset + 11] = center.z
		buffer[offset + 12] = float(index) / maxf(data.cell_count() - 1, 1.0)
		buffer[offset + 13] = fposmod(float(index) * 0.61803398875, 1.0)
		buffer[offset + 14] = (center.y - box.position.y) / maxf(box.size.y, 0.001)
		buffer[offset + 15] = 1.0
	instances.buffer = buffer
	instances.custom_aabb = box.grow(data.cell_span * 9.0 * 3.5)
	multimesh = instances
	if animate and data.cell_count() > 0:
		replay()
	else:
		timeline.elapsed = timeline.duration
		_set_parameter("assembly_time", timeline.elapsed)
		set_process(false)
		assembly_progress.emit(1.0, timeline.duration)

func replay() -> void:
	if structure == null or structure.cell_count() == 0:
		return
	timeline.restart()
	_set_parameter("assembly_time", 0.0)
	assembly_progress.emit(0.0, timeline.duration)
	set_process(true)

## Jump the assembly clock, e.g. to inspect or capture a partial build.
func seek(fraction: float) -> void:
	timeline.elapsed = clampf(fraction, 0.0, 1.0) * timeline.duration
	_set_parameter("assembly_time", timeline.elapsed)
	assembly_progress.emit(timeline.fraction(), timeline.duration)
	set_process(timeline.is_running())

func set_gap(value: float) -> void:
	_gap = clampf(value, 0.0, 0.25)
	_set_parameter("cell_gap", _gap)

## Semi-transparent view: faint faces, bright edges and every vertex visible.
func set_xray(enabled: bool) -> void:
	xray = enabled
	material_override = _xray if enabled else _solid
	cast_shadow = GeometryInstance3D.SHADOW_CASTING_SETTING_OFF if enabled else GeometryInstance3D.SHADOW_CASTING_SETTING_ON

func _set_parameter(name: StringName, value: Variant) -> void:
	_solid.set_shader_parameter(name, value)
	_xray.set_shader_parameter(name, value)

func _process(delta: float) -> void:
	timeline.advance(delta)
	_set_parameter("assembly_time", timeline.elapsed)
	assembly_progress.emit(timeline.fraction(), timeline.duration)
	if not timeline.is_running():
		set_process(false)
		assembly_finished.emit()
