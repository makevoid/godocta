class_name StructureRenderer
extends MultiMeshInstance3D
## One render node and one shared mesh, regardless of the cell count.
## The 64-byte per-instance GPU buffer is derived from the compact structure.

signal assembly_progress(fraction: float)
signal assembly_finished

var structure: OctaStructure
var timeline := AssemblyTimeline.new()
var _material := ShaderMaterial.new()
var _gap: float = 0.045

func _init() -> void:
	_material.shader = preload("res://shaders/cell.gdshader")
	_material.set_shader_parameter("release_window", AssemblyTimeline.RELEASE_WINDOW)
	_material.set_shader_parameter("settle_time", AssemblyTimeline.SETTLE_TIME)
	material_override = _material
	set_process(false)

func display(data: OctaStructure, animate: bool = true) -> void:
	structure = data
	var instances := MultiMesh.new()
	instances.transform_format = MultiMesh.TRANSFORM_3D
	instances.use_custom_data = true
	instances.mesh = TruncatedOctahedronMesh.shared_mesh()
	instances.instance_count = data.cell_count()
	var box := data.bounds()
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
	instances.custom_aabb = box.grow(data.cell_span * 9.0)
	multimesh = instances
	if animate and data.cell_count() > 0:
		replay()
	else:
		timeline.elapsed = AssemblyTimeline.DURATION
		_material.set_shader_parameter("assembly_time", timeline.elapsed)
		set_process(false)
		assembly_progress.emit(1.0)

func replay() -> void:
	if structure == null or structure.cell_count() == 0:
		return
	timeline.restart()
	_material.set_shader_parameter("assembly_time", 0.0)
	assembly_progress.emit(0.0)
	set_process(true)

func set_gap(value: float) -> void:
	_gap = clampf(value, 0.0, 0.25)
	_material.set_shader_parameter("cell_gap", _gap)

func _process(delta: float) -> void:
	timeline.advance(delta)
	_material.set_shader_parameter("assembly_time", timeline.elapsed)
	assembly_progress.emit(timeline.fraction())
	if not timeline.is_running():
		set_process(false)
		assembly_finished.emit()
