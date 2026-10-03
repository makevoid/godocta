shader_type spatial;
render_mode unshaded;
varying vec3 world_position;

void vertex() {
	world_position = (MODEL_MATRIX * vec4(VERTEX, 1.0)).xyz;
}

void fragment() {
	vec2 p = world_position.xz;
	vec2 grid = abs(fract(p - 0.5) - 0.5) / max(fwidth(p), vec2(0.001));
	float line = 1.0 - min(min(grid.x, grid.y), 1.0);
	float fade = 1.0 - smoothstep(9.0, 25.0, length(p));
	float ring = (1.0 - smoothstep(0.015, 0.06, abs(length(p) - 9.0))) * 0.3;
	vec3 base = vec3(0.039, 0.055, 0.065);
	ALBEDO = base + vec3(0.08, 0.13, 0.14) * (line * 0.32 * fade + ring);
}
