// Shared by the solid and X-ray cell materials: assembly motion and polygon geometry.
uniform float assembly_time = 4.0;
uniform float release_window = 2.8;
uniform float settle_time = 1.2;
uniform float flight_scale = 1.0;
uniform float cell_gap : hint_range(0.0, 0.25) = 0.0;
uniform vec4 base_color : source_color = vec4(0.49, 0.76, 0.65, 1.0);
// Up to 16 structure colors, already linear (array uniforms skip source_color conversion).
// INSTANCE_CUSTOM.a picks one, or is negative for base_color.
uniform vec4 palette[16];
varying float height_tint;
varying float arriving;
varying vec4 paint; // rgb = cell color, a = 1 when it comes from the palette.

vec4 paint_for(vec4 custom) {
	if (custom.a < 0.0) {
		return vec4(base_color.rgb, 0.0);
	}
	vec3 color = palette[clamp(int(custom.a + 0.5), 0, 15)].rgb;
	// Filmic tonemapping boosts mid-tone saturation: pull back so file colors read true.
	float luma = dot(color, vec3(0.2126, 0.7152, 0.0722));
	return vec4(mix(vec3(luma), color, 0.72), 1.0);
}

// Height shading: strong on the single-color look, subtle so palette colors stay true.
vec3 shade(vec4 cell_paint, float tint, vec2 plain, vec2 painted) {
	vec2 range = mix(plain, painted, cell_paint.a);
	return cell_paint.rgb * mix(range.x, range.y, tint);
}

// Cells drop in on a spiral, scale up as they arrive and settle with cubic ease-out.
// INSTANCE_CUSTOM: r = assembly rank, g = random seed, b = normalized height, a = color.
// Outputs the height tint and arrival glow; callers store them in the varyings.
void assemble(inout vec3 vertex, inout vec3 normal, vec4 custom, out float tint, out float glow) {
	tint = custom.b;
	glow = 0.0;
	// This uniform branch also skips the rank/seed work in every shadow pass.
	if (assembly_time >= release_window + settle_time) {
		vertex *= 1.0 - cell_gap;
		return;
	}
	float delay = release_window * (1.0 - sqrt(1.0 - custom.r));
	float t = clamp((assembly_time - delay) / settle_time, 0.0, 1.0);
	if (t >= 1.0) {
		vertex *= 1.0 - cell_gap;
		return;
	}
	float ease_out = 1.0 - pow(1.0 - t, 3.0);
	float remaining = 1.0 - ease_out;
	float seed = custom.g;
	float angle = (seed - 0.5) * 2.8 * remaining;
	mat3 turn = mat3(vec3(cos(angle), 0.0, -sin(angle)), vec3(0.0, 1.0, 0.0), vec3(sin(angle), 0.0, cos(angle)));
	vertex = turn * vertex * (1.0 - cell_gap) * smoothstep(0.0, 0.28, t);
	vertex += vec3(sin(seed * 6.283185) * 3.5, 5.0 + seed * 2.0, cos(seed * 6.283185) * 3.5) * remaining * flight_scale;
	normal = turn * normal;
	glow = sin(t * PI) * 0.2;
}

// Distance from a face-local point to the nearest polygon edge. polygon = (sides, apothem).
// The shared mesh has only squares and hexagons, with a vertex on the +U axis.
// Fold into the first quadrant to replace six trigonometric half-plane tests.
float edge_distance(vec2 uv, vec2 polygon) {
	vec2 p = abs(uv);
	if (polygon.x < 5.0) {
		return polygon.y - (p.x + p.y) * 0.707106781187;
	}
	return polygon.y - max(p.y, p.x * 0.866025403784 + p.y * 0.5);
}

// Distance from a face-local point to the nearest polygon vertex.
// Symmetry leaves two candidates; compare their projections before taking one length.
float vertex_distance(vec2 uv, vec2 polygon) {
	vec2 p = abs(uv);
	if (polygon.x < 5.0) {
		float radius = polygon.y * 1.414213562373;
		return length(vec2(max(p.x, p.y) - radius, min(p.x, p.y)));
	}
	float radius = polygon.y * 1.154700538379;
	vec2 corner = p.x > p.y * 1.732050807569 ? vec2(radius, 0.0) : vec2(radius * 0.5, polygon.y);
	return length(p - corner);
}
