shader_type spatial;
// Brush outline: a holographic globe drawn over the build. Glowing rim plus faint
// latitude/longitude lines from the sphere's UVs, always visible (no depth test).
render_mode unshaded, blend_add, depth_draw_never, depth_test_disabled, cull_disabled, shadows_disabled;

uniform vec4 tint : source_color = vec4(0.72, 0.91, 0.8, 1.0);
uniform float intensity = 1.0;
uniform float fill = 0.03;
uniform float rim = 0.55;
uniform float lines = 0.16;
uniform vec2 grid = vec2(24.0, 12.0);
// Clip to the upper half (local y ≥ 0): the scoop's dome.
uniform bool dome = false;
varying float local_y;

void vertex() {
	local_y = VERTEX.y;
}

void fragment() {
	if (dome && local_y < 0.0) {
		discard;
	}
	float facing = abs(dot(normalize(NORMAL), normalize(VIEW)));
	float edge = pow(1.0 - facing, 2.5);
	vec2 cells = UV * grid;
	vec2 distance = abs(fract(cells + 0.5) - 0.5);
	vec2 width = max(fwidth(cells) * 1.2, vec2(1e-4));
	vec2 line = 1.0 - smoothstep(vec2(0.0), width, distance);
	float side = FRONT_FACING ? 1.0 : 0.45;
	ALBEDO = tint.rgb * (fill + edge * rim + max(line.x, line.y) * lines) * intensity * side;
}
