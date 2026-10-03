shader_type spatial;
// X-ray: faint additive faces, glowing edges and vertex dots, front and back faces.
// Dense regions glow brighter, so interiors, floors and hollow storeys stay legible.
render_mode unshaded, blend_add, depth_draw_never, cull_disabled, shadows_disabled;

#include "res://shaders/cell_common.gdshaderinc"

uniform float face_alpha = 0.012;
uniform float edge_alpha = 0.16;
uniform float vertex_alpha = 0.9;
uniform vec4 vertex_color : source_color = vec4(0.95, 1.0, 0.86, 1.0);

void vertex() {
	float tint;
	float glow;
	assemble(VERTEX, NORMAL, INSTANCE_CUSTOM, tint, glow);
	height_tint = tint;
	arriving = glow;
	paint = paint_for(INSTANCE_CUSTOM);
}

void fragment() {
	float to_edge = edge_distance(UV, UV2);
	float to_vertex = vertex_distance(UV, UV2);
	// World units per pixel. Far away, cells span only a few pixels: fade lines and dots
	// so the build reads as a glowing volume instead of saturating to white.
	float texel = length(fwidth(UV));
	float detail = 1.0 - smoothstep(0.02, 0.09, texel);
	float line_width = max(0.006, texel * 0.9);
	float dot_radius = max(0.03, texel * 1.4);
	float line = 1.0 - smoothstep(line_width * 0.5, line_width, to_edge);
	float dot = 1.0 - smoothstep(dot_radius * 0.6, dot_radius, to_vertex);
	vec3 cell_color = shade(paint, height_tint, vec2(0.75, 1.2), vec2(0.9, 1.1));
	vec3 color = cell_color * (face_alpha + line * edge_alpha * mix(0.12, 1.0, detail))
		+ vertex_color.rgb * dot * vertex_alpha * mix(0.04, 1.0, detail);
	ALBEDO = color * (1.0 + arriving * 4.0);
	ALPHA = 1.0;
}
