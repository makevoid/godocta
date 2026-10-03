class_name UnitGlyph
extends Control
## Small vector diagram of the actual unit, derived from the same 24 vertices.

func _ready() -> void:
	custom_minimum_size = Vector2(98, 95)
	mouse_filter = Control.MOUSE_FILTER_IGNORE

func _draw() -> void:
	var vertices := TruncatedOctahedronMesh.vertices()
	var rotation := Basis(Vector3.RIGHT, -0.42) * Basis(Vector3.UP, 0.62)
	for i in vertices.size():
		for j in range(i + 1, vertices.size()):
			if not is_equal_approx(vertices[i].distance_squared_to(vertices[j]), 0.125):
				continue
			var a: Vector3 = rotation * vertices[i]
			var b: Vector3 = rotation * vertices[j]
			var tint := StudioTheme.ACCENT
			tint.a = 0.35 if a.z + b.z < -0.1 else 0.95
			draw_line(size * 0.5 + Vector2(a.x, -a.y) * 72.0, size * 0.5 + Vector2(b.x, -b.y) * 72.0, tint, 1.3, true)
