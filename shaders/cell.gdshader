shader_type spatial;
render_mode cull_back;

#include "res://shaders/cell_common.gdshaderinc"

void vertex() {
	float tint;
	float glow;
	assemble(VERTEX, NORMAL, INSTANCE_CUSTOM, tint, glow);
	height_tint = tint;
	arriving = glow;
}

void fragment() {
	float edge = smoothstep(0.001, 0.009, edge_distance(UV, UV2));
	vec3 cell_color = base_color.rgb * mix(0.66, 1.13, height_tint);
	ALBEDO = mix(cell_color * 0.42, cell_color, edge);
	ROUGHNESS = 0.6;
	METALLIC = 0.13;
	EMISSION = cell_color * (0.055 + arriving);
}
