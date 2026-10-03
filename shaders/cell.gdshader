shader_type spatial;
render_mode cull_back;

uniform float assembly_time = 4.0;
uniform float release_window = 2.8;
uniform float settle_time = 1.2;
uniform float cell_gap : hint_range(0.0, 0.25) = 0.045;
uniform vec4 base_color : source_color = vec4(0.49, 0.76, 0.65, 1.0);
varying float height_tint;
varying float arriving;

void vertex() {
	float delay = release_window * (1.0 - sqrt(1.0 - INSTANCE_CUSTOM.r));
	float t = clamp((assembly_time - delay) / settle_time, 0.0, 1.0);
	float ease_out = 1.0 - pow(1.0 - t, 3.0);
	float remaining = 1.0 - ease_out;
	float seed = INSTANCE_CUSTOM.g;
	float angle = (seed - 0.5) * 2.8 * remaining;
	mat3 turn = mat3(vec3(cos(angle), 0.0, -sin(angle)), vec3(0.0, 1.0, 0.0), vec3(sin(angle), 0.0, cos(angle)));
	VERTEX = turn * VERTEX * (1.0 - cell_gap) * smoothstep(0.0, 0.28, t);
	VERTEX += vec3(sin(seed * 6.283185) * 3.5, 5.0 + seed * 2.0, cos(seed * 6.283185) * 3.5) * remaining;
	NORMAL = turn * NORMAL;
	height_tint = INSTANCE_CUSTOM.b;
	arriving = sin(t * 3.14159265) * 0.2;
}

void fragment() {
	float sides = UV2.x;
	float edge_distance = 1.0;
	for (int i = 0; i < 6; i++) {
		if (float(i) < sides) {
			float angle = (2.0 * float(i) + 1.0) * 3.14159265 / sides;
			edge_distance = min(edge_distance, UV2.y - dot(UV, vec2(cos(angle), sin(angle))));
		}
	}
	float edge = smoothstep(0.001, 0.009, edge_distance);
	vec3 cell_color = base_color.rgb * mix(0.66, 1.13, height_tint);
	ALBEDO = mix(cell_color * 0.42, cell_color, edge);
	ROUGHNESS = 0.6;
	METALLIC = 0.13;
	EMISSION = cell_color * (0.055 + arriving);
}
