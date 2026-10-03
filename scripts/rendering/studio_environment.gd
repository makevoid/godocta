class_name StudioEnvironment
extends Node3D
## Lights, sky dome and ground. `apply_theme` switches between WorldTheme worlds.

var theme_name: String = ""
var _environment: Environment
var _sky: ShaderMaterial
var _key: DirectionalLight3D
var _fill: DirectionalLight3D
var _ground_mesh: MeshInstance3D
var _ground: ShaderMaterial

func _ready() -> void:
	var world := WorldEnvironment.new()
	_environment = Environment.new()
	_environment.background_mode = Environment.BG_SKY
	_sky = ShaderMaterial.new()
	_sky.shader = preload("res://shaders/sky.gdshader")
	var sky := Sky.new()
	sky.sky_material = _sky
	sky.radiance_size = Sky.RADIANCE_SIZE_64
	_environment.sky = sky
	_environment.ambient_light_source = Environment.AMBIENT_SOURCE_COLOR
	_environment.reflected_light_source = Environment.REFLECTION_SOURCE_SKY
	_environment.tonemap_mode = Environment.TONE_MAPPER_FILMIC
	world.environment = _environment
	add_child(world)
	_key = DirectionalLight3D.new()
	_key.shadow_enabled = true
	_key.directional_shadow_max_distance = 75.0
	# Two cascades instead of four: every cascade redraws all cells.
	_key.directional_shadow_mode = DirectionalLight3D.SHADOW_PARALLEL_2_SPLITS
	add_child(_key)
	_fill = DirectionalLight3D.new()
	_fill.rotation_degrees = Vector3(-20, 130, 0)
	_fill.sky_mode = DirectionalLight3D.SKY_MODE_LIGHT_ONLY # Only the key appears as the sun.
	add_child(_fill)
	_ground_mesh = MeshInstance3D.new()
	var plane := PlaneMesh.new()
	plane.size = Vector2(40000, 40000) # Its edge sits just under the true horizon.
	_ground_mesh.mesh = plane
	_ground = ShaderMaterial.new()
	_ground.shader = preload("res://shaders/ground.gdshader")
	_ground_mesh.material_override = _ground
	add_child(_ground_mesh)
	apply_theme(WorldTheme.STUDIO)

## Switch sky, ground and lighting. Unknown names fall back to the studio.
func apply_theme(name: String) -> void:
	var themes := WorldTheme.themes()
	theme_name = name if themes.has(name) else WorldTheme.STUDIO
	var t: Dictionary = themes[theme_name]
	_sky.set_shader_parameter("zenith_color", t.zenith)
	_sky.set_shader_parameter("horizon_color", t.horizon)
	_sky.set_shader_parameter("below_color", t.below)
	_sky.set_shader_parameter("glow_color", t.get("glow", Color.WHITE))
	_sky.set_shader_parameter("glow_strength", t.get("glow_strength", 0.0))
	_sky.set_shader_parameter("sun_size", t.get("sun_size", 0.03))
	_sky.set_shader_parameter("sun_strength", t.get("sun_strength", 0.0))
	_sky.set_shader_parameter("cloud_cover", t.get("clouds", 0.0))
	_sky.set_shader_parameter("cloud_color", t.get("cloud_color", Color.WHITE))
	_sky.set_shader_parameter("cloud_scale", t.get("cloud_scale", 1.0))
	_sky.set_shader_parameter("silhouette_color", t.get("silhouette", Color.BLACK))
	_sky.set_shader_parameter("silhouette_height", t.get("silhouette_height", 0.0))
	_sky.set_shader_parameter("silhouette_city", t.get("silhouette_city", 0.0))
	_sky.set_shader_parameter("silhouette_seed", t.get("silhouette_seed", 0.0))
	_sky.set_shader_parameter("haze", t.get("haze", 0.4))
	_sky.set_shader_parameter("stars", t.get("stars", 0.0))
	_sky.set_shader_parameter("city_lights", t.get("city_lights", 0.0))
	_ground.set_shader_parameter("ground_color", t.ground)
	_ground.set_shader_parameter("far_color", t.below)
	_ground.set_shader_parameter("grid_strength", t.get("grid", 0.0))
	_ground.set_shader_parameter("lit", t.get("lit", 1.0))
	_ground.set_shader_parameter("mottle", t.get("mottle", 0.0))
	_ground.set_shader_parameter("water", t.get("water", 0.0))
	_ground_mesh.position.y = t.get("water_level", 0.0)
	_key.light_color = t.sun
	_key.light_energy = t.sun_energy
	_key.shadow_opacity = t.get("shadow_opacity", 1.0)
	_key.shadow_blur = t.get("shadow_blur", 1.0)
	_key.rotation_degrees = Vector3(-t.elevation, t.azimuth, 0)
	_fill.light_color = t.fill
	_fill.light_energy = t.fill_energy
	_environment.ambient_light_color = t.ambient
	_environment.ambient_light_energy = t.ambient_energy

## Scale the shadow range, the studio grid and the horizon fade to a structure.
func fit(box: AABB) -> void:
	var reach := maxf(Vector2(box.size.x, box.size.z).length() * 0.5, 9.0)
	_key.directional_shadow_max_distance = maxf(75.0, box.size.length() * 2.6)
	_ground.set_shader_parameter("ring_radius", reach)
	_ground.set_shader_parameter("fade_start", reach)
	_ground.set_shader_parameter("fade_end", reach * 2.7)
	_ground.set_shader_parameter("far_start", reach * 2.0)
	_ground.set_shader_parameter("far_end", reach * 10.0)
