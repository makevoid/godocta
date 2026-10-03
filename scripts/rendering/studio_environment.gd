class_name StudioEnvironment
extends Node3D

func _ready() -> void:
	var world := WorldEnvironment.new()
	var environment := Environment.new()
	environment.background_mode = Environment.BG_COLOR
	environment.background_color = Color("0a0e11")
	environment.ambient_light_source = Environment.AMBIENT_SOURCE_COLOR
	environment.ambient_light_color = Color("b0ced7")
	environment.ambient_light_energy = 0.55
	environment.tonemap_mode = Environment.TONE_MAPPER_FILMIC
	world.environment = environment
	add_child(world)
	var key := DirectionalLight3D.new()
	key.rotation_degrees = Vector3(-48, -35, 0)
	key.light_color = Color("e7f4e9")
	key.light_energy = 1.6
	key.shadow_enabled = true
	key.directional_shadow_max_distance = 75.0
	add_child(key)
	var fill := DirectionalLight3D.new()
	fill.rotation_degrees = Vector3(-20, 130, 0)
	fill.light_color = Color("8cbdcd")
	fill.light_energy = 0.7
	add_child(fill)
	var ground := MeshInstance3D.new()
	var plane := PlaneMesh.new()
	plane.size = Vector2(2000, 2000)
	ground.mesh = plane
	var material := ShaderMaterial.new()
	material.shader = preload("res://shaders/ground.gdshader")
	ground.material_override = material
	add_child(ground)
