shader_type spatial;
// The ground disc: a studio grid, or themed terrain/water that fades into the horizon.
uniform vec4 ground_color : source_color = vec4(0.039, 0.055, 0.065, 1.0);
uniform vec4 far_color : source_color = vec4(0.039, 0.055, 0.065, 1.0);
uniform vec4 line_color : source_color = vec4(0.08, 0.13, 0.14, 1.0);
uniform float grid_strength = 1.0;
uniform float lit = 0.3; // 0 = flat studio floor, 1 = fully lit terrain (receives shadows).
uniform float mottle = 0.0; // Low-frequency color variation for grass, sand and earth.
uniform float water = 0.0;
uniform float ring_radius = 9.0;
uniform float fade_start = 9.0;
uniform float fade_end = 25.0;
uniform float far_start = 120.0;
uniform float far_end = 900.0;
varying vec3 world_position;

float hash(vec2 p) {
	return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453);
}

float noise(vec2 p) {
	vec2 i = floor(p);
	vec2 f = fract(p);
	vec2 u = f * f * (3.0 - 2.0 * f);
	return mix(mix(hash(i), hash(i + vec2(1.0, 0.0)), u.x),
		mix(hash(i + vec2(0.0, 1.0)), hash(i + vec2(1.0, 1.0)), u.x), u.y);
}

void vertex() {
	world_position = (MODEL_MATRIX * vec4(VERTEX, 1.0)).xyz;
}

void fragment() {
	vec2 p = world_position.xz;
	float distance = length(p);
	float variation = 0.0;
	if (mottle > 0.0) {
		variation = (noise(p * 0.08) * 0.65 + noise(p * 0.5) * 0.35 - 0.5) * mottle;
	}
	vec3 near = ground_color.rgb * (1.0 + variation);
	// Theme uniforms make these branches coherent across the entire ground draw.
	if (grid_strength > 0.0) {
		vec2 grid = abs(fract(p - 0.5) - 0.5) / max(fwidth(p), vec2(0.001));
		float line = 1.0 - min(min(grid.x, grid.y), 1.0);
		float fade = 1.0 - smoothstep(fade_start, fade_end, distance);
		float ring = (1.0 - smoothstep(0.015, 0.06, abs(distance - ring_radius))) * 0.3;
		near += line_color.rgb * (line * 0.32 * fade + ring) * grid_strength;
	}
	float far = smoothstep(far_start, far_end, distance);
	ALBEDO = near * lit * (1.0 - far);
	EMISSION = near * (1.0 - lit) * (1.0 - far) + far_color.rgb * far;
	ROUGHNESS = mix(0.95, 0.3, water);
	SPECULAR = mix(0.25, 0.45, water);
	if (water > 0.0) {
		// Small static ripples so the sun and sky catch the surface.
		// Low frequency only: fine ripples alias into grain at a distance.
		vec2 wave = vec2(noise(p * 0.22) - noise(p * 0.22 + 3.7), noise(p * 0.17 + 9.1) - noise(p * 0.17 + 5.3));
		vec3 normal = normalize(vec3(wave.x * 0.12 * water, 1.0, wave.y * 0.12 * water));
		NORMAL = normalize((VIEW_MATRIX * vec4(normal, 0.0)).xyz);
	}
}
