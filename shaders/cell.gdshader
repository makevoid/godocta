shader_type spatial;
render_mode cull_back;

#include "res://shaders/cell_common.gdshaderinc"

void vertex() {
	float tint;
	float glow;
	assemble(VERTEX, NORMAL, INSTANCE_CUSTOM, tint, glow);
	height_tint = tint;
	arriving = glow;
	paint = paint_for(INSTANCE_CUSTOM);
}

void fragment() {
	float edge = smoothstep(0.001, 0.009, edge_distance(UV, UV2));
	vec3 cell_color = shade(paint, height_tint, vec2(0.66, 1.13), vec2(0.86, 1.05));
	ALBEDO = mix(cell_color * mix(0.42, 0.6, paint.a), cell_color, edge);
	ROUGHNESS = 0.6;
	METALLIC = 0.13;
	EMISSION = cell_color * (0.055 + arriving);
}
