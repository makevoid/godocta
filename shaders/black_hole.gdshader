shader_type spatial;
// Camera-facing black hole: a dark event horizon, a bright photon ring and a swirling
// accretion disk. strength rises from a hover (≈0.3) to a full pull (1) while erasing.
render_mode unshaded, blend_mix, depth_draw_never, depth_test_disabled, cull_disabled, shadows_disabled;

uniform float strength = 0.3;
uniform vec4 inner : source_color = vec4(1.0, 0.66, 0.34, 1.0);
uniform vec4 outer : source_color = vec4(0.52, 0.42, 1.0, 1.0);

void vertex() {
	// Billboard that keeps the node's scale.
	MODELVIEW_MATRIX = VIEW_MATRIX * mat4(INV_VIEW_MATRIX[0], INV_VIEW_MATRIX[1], INV_VIEW_MATRIX[2], MODEL_MATRIX[3]);
	MODELVIEW_MATRIX = MODELVIEW_MATRIX * mat4(
		vec4(length(MODEL_MATRIX[0].xyz), 0.0, 0.0, 0.0),
		vec4(0.0, length(MODEL_MATRIX[1].xyz), 0.0, 0.0),
		vec4(0.0, 0.0, length(MODEL_MATRIX[2].xyz), 0.0),
		vec4(0.0, 0.0, 0.0, 1.0));
}

void fragment() {
	vec2 p = UV * 2.0 - 1.0;
	float r = length(p);
	float horizon = 0.26;
	float angle = atan(p.y, p.x);
	// Spiral arms wind tighter towards the horizon and spin faster with strength.
	float swirl = 0.5 + 0.5 * sin(angle * 3.0 + 7.0 / (r + 0.12) - TIME * (2.5 + 9.0 * strength));
	float disk = smoothstep(1.0, 0.3, r) * smoothstep(horizon, horizon + 0.08, r);
	float ring = exp(-pow((r - horizon - 0.025) / 0.03, 2.0));
	float core = 1.0 - smoothstep(horizon - 0.015, horizon + 0.005, r);
	vec3 color = mix(outer.rgb, inner.rgb, smoothstep(0.85, 0.3, r)) * disk * (0.35 + 0.65 * swirl) * (0.5 + strength)
		+ vec3(1.0, 0.93, 0.8) * ring * (0.7 + strength);
	ALBEDO = color * (1.0 - core);
	ALPHA = clamp(max(core, disk * (0.25 + 0.55 * swirl) * (0.45 + 0.55 * strength) + ring), 0.0, 1.0) * smoothstep(1.0, 0.8, r);
}
