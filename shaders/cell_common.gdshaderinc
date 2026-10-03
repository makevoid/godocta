// Shared by the solid and X-ray cell materials: assembly motion and polygon geometry.
uniform float assembly_time = 4.0;
uniform float release_window = 2.8;
uniform float settle_time = 1.2;
uniform float flight_scale = 1.0;
uniform float cell_gap : hint_range(0.0, 0.25) = 0.045;
uniform vec4 base_color : source_color = vec4(0.49, 0.76, 0.65, 1.0);
varying float height_tint;
varying float arriving;

// Cells drop in on a spiral, scale up as they arrive and settle with cubic ease-out.
// INSTANCE_CUSTOM: r = assembly rank, g = random seed, b = normalized height.
// Outputs the height tint and arrival glow; callers store them in the varyings.
void assemble(inout vec3 vertex, inout vec3 normal, vec4 custom, out float tint, out float glow) {
	float delay = release_window * (1.0 - sqrt(1.0 - custom.r));
	float t = clamp((assembly_time - delay) / settle_time, 0.0, 1.0);
	float ease_out = 1.0 - pow(1.0 - t, 3.0);
	float remaining = 1.0 - ease_out;
	float seed = custom.g;
	float angle = (seed - 0.5) * 2.8 * remaining;
	mat3 turn = mat3(vec3(cos(angle), 0.0, -sin(angle)), vec3(0.0, 1.0, 0.0), vec3(sin(angle), 0.0, cos(angle)));
	vertex = turn * vertex * (1.0 - cell_gap) * smoothstep(0.0, 0.28, t);
	vertex += vec3(sin(seed * 6.283185) * 3.5, 5.0 + seed * 2.0, cos(seed * 6.283185) * 3.5) * remaining * flight_scale;
	normal = turn * normal;
	tint = custom.b;
	glow = sin(t * PI) * 0.2;
}

// Distance from a face-local point to the nearest polygon edge. polygon = (sides, apothem).
// Vertex i of the regular polygon lies at angle 2πi/sides; edge normals bisect them.
float edge_distance(vec2 uv, vec2 polygon) {
	float result = 1.0;
	for (int i = 0; i < 6; i++) {
		if (float(i) < polygon.x) {
			float angle = (2.0 * float(i) + 1.0) * PI / polygon.x;
			result = min(result, polygon.y - dot(uv, vec2(cos(angle), sin(angle))));
		}
	}
	return result;
}

// Distance from a face-local point to the nearest polygon vertex.
float vertex_distance(vec2 uv, vec2 polygon) {
	float radius = polygon.y / cos(PI / polygon.x);
	float result = 1.0;
	for (int i = 0; i < 6; i++) {
		if (float(i) < polygon.x) {
			float angle = 2.0 * float(i) * PI / polygon.x;
			result = min(result, length(uv - radius * vec2(cos(angle), sin(angle))));
		}
	}
	return result;
}
