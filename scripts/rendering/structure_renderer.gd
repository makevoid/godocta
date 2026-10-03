class_name StructureRenderer
extends MultiMeshInstance3D
## One render node and one shared mesh, regardless of the cell count.
## The 64-byte per-instance GPU buffer is derived from the compact structure.
## Once a solid build has settled, only its surface cells are drawn: a background task
## drops every cell whose 14 neighbors are all occupied. Assembly and X-ray use all cells.

signal assembly_progress(fraction: float, duration: float)
signal assembly_finished

var structure: OctaStructure
var timeline := AssemblyTimeline.new()
var xray: bool = false
var _solid := ShaderMaterial.new()
var _xray := ShaderMaterial.new()
var _gap: float = 0.045
## Every cell; drawn while assembling and in X-ray.
var full_multimesh: MultiMesh
## Surface cells only, or null until the background task finishes.
var surface_multimesh: MultiMesh
var _generation: int = 0
var _tasks: Dictionary = {} # generation → WorkerThreadPool task id
const SURFACE_MAX_GAP := 0.1 # Wider gaps would let hidden interiors show through.

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
	_set_parameter("flight_scale", clampf(pow(box.get_longest_axis_size() / 10.0, 0.25), 1.0, 2.0))
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
		buffer[offset + 15] = float(data.color_index(index))
	instances.buffer = buffer
	# File colors are sRGB; shading happens in linear space.
	var palette := PackedColorArray()
	for color in data.palette:
		palette.append(color.srgb_to_linear())
	palette.resize(OctaStructure.MAX_PALETTE)
	_set_parameter("palette", palette)
	instances.custom_aabb = box.grow(data.cell_span * 9.0 * 2.0)
	full_multimesh = instances
	surface_multimesh = null
	multimesh = instances
	_generation += 1
	if data.cell_count() > 0:
		_tasks[_generation] = WorkerThreadPool.add_task(_compute_surface.bind(data, buffer, _generation), false, "Surface cells")
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
	_refresh_instances()

## Instances currently drawn: the surface once settled and solid, otherwise every cell.
func drawn_count() -> int:
	return multimesh.instance_count if multimesh != null else 0

func _refresh_instances() -> void:
	var settled := not timeline.is_running() and not xray and _gap <= SURFACE_MAX_GAP
	var target := surface_multimesh if settled and surface_multimesh != null else full_multimesh
	if multimesh != target:
		multimesh = target

## Runs on a worker thread: reads the immutable structure and buffer, never the scene.
func _compute_surface(data: OctaStructure, buffer: PackedFloat32Array, generation: int) -> void:
	var cells := data.packed_cells()
	var occupied: Dictionary = {}
	for offset in range(0, cells.size(), 3):
		occupied[Vector3i(cells[offset], cells[offset + 1], cells[offset + 2])] = true
	var surface := PackedFloat32Array()
	for index in cells.size() / 3:
		var cell := Vector3i(cells[index * 3], cells[index * 3 + 1], cells[index * 3 + 2])
		for neighbor in OctaLattice.NEIGHBORS:
			if not occupied.has(cell + neighbor):
				surface.append_array(buffer.slice(index * 16, index * 16 + 16))
				break
	_surface_ready.call_deferred(generation, surface)

func _surface_ready(generation: int, surface: PackedFloat32Array) -> void:
	WorkerThreadPool.wait_for_task_completion(_tasks[generation])
	_tasks.erase(generation)
	if generation != _generation:
		return # A newer structure replaced this one.
	var instances := MultiMesh.new()
	instances.transform_format = MultiMesh.TRANSFORM_3D
	instances.use_custom_data = true
	instances.mesh = full_multimesh.mesh
	instances.instance_count = surface.size() / 16
	instances.buffer = surface
	instances.custom_aabb = full_multimesh.custom_aabb
	surface_multimesh = instances
	_refresh_instances()

func _exit_tree() -> void:
	for task in _tasks.values():
		WorkerThreadPool.wait_for_task_completion(task)
	_tasks.clear()

## Jump the assembly clock, e.g. to inspect or capture a partial build.
func seek(fraction: float) -> void:
	timeline.elapsed = clampf(fraction, 0.0, 1.0) * timeline.duration
	_set_parameter("assembly_time", timeline.elapsed)
	assembly_progress.emit(timeline.fraction(), timeline.duration)
	set_process(timeline.is_running())
	_refresh_instances()

func set_gap(value: float) -> void:
	_gap = clampf(value, 0.0, 0.25)
	_set_parameter("cell_gap", _gap)
	_refresh_instances()

## Semi-transparent view: faint faces, bright edges and every vertex visible.
func set_xray(enabled: bool) -> void:
	xray = enabled
	material_override = _xray if enabled else _solid
	cast_shadow = GeometryInstance3D.SHADOW_CASTING_SETTING_OFF if enabled else GeometryInstance3D.SHADOW_CASTING_SETTING_ON
	_refresh_instances()

func _set_parameter(name: StringName, value: Variant) -> void:
	_solid.set_shader_parameter(name, value)
	_xray.set_shader_parameter(name, value)

func _process(delta: float) -> void:
	timeline.advance(delta)
	_set_parameter("assembly_time", timeline.elapsed)
	assembly_progress.emit(timeline.fraction(), timeline.duration)
	if not timeline.is_running():
		set_process(false)
		_refresh_instances()
		assembly_finished.emit()
