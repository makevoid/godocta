extends SceneTree
## Loads the real studio scene, assembles structures and saves PNG screenshots.
## Run it through tools/capture.sh, which keeps the window hidden in the background.
##   --shapes=cube,sphere,castle   Presets by name, or structures/<name>.octa stems. Default: all.
##   --views=hero,front,side,top   Camera angles for the finished build. Default: hero,front,top.
##   --modes=solid,xray            Render modes. Default: solid.
##   --at=0.25,0.6                 Also capture mid-assembly at these fractions (hero view).
##   --gap=0.045 --out=captures

const VIEWS := {
	"hero": Vector2(0.68, 0.48),
	"front": Vector2(0.0, 0.08),
	"side": Vector2(PI * 0.5, 0.08),
	"top": Vector2(0.0, 1.35),
}

var studio: OctaStudio
var out_dir: String = "res://captures"

func _initialize() -> void:
	call_deferred("_run")

func _run() -> void:
	var options := _parse_options()
	out_dir = options.get("out", out_dir)
	if not out_dir.begins_with("res://") and not out_dir.begins_with("/"):
		out_dir = "res://" + out_dir
	DirAccess.make_dir_recursive_absolute(ProjectSettings.globalize_path(out_dir))
	studio = load("res://scenes/main.tscn").instantiate() as OctaStudio
	root.add_child(studio)
	current_scene = studio
	await process_frame
	var all_names := Array(StructurePresets.TITLES).map(func(t: String) -> String: return t.to_lower())
	for path in studio.hud.toolbar.gallery_paths:
		all_names.append(path.get_file().get_basename())
	var names: PackedStringArray = options.get("shapes", ",".join(all_names)).split(",", false)
	var views: PackedStringArray = options.get("views", "hero,front,top").split(",", false)
	var modes: PackedStringArray = options.get("modes", "xray" if options.has("xray") else "solid").split(",", false)
	var stages: PackedStringArray = options.get("at", "").split(",", false)
	if options.has("gap"):
		studio.renderer.set_gap(float(options["gap"]))
	for name in names:
		if not _show(name.strip_edges().to_lower()):
			push_error("Unknown structure: " + name)
			continue
		print("Showing %s: %d cells, %.1f s assembly" % [name, studio.renderer.structure.cell_count(), studio.renderer.timeline.duration])
		for mode in modes:
			studio.renderer.set_xray(mode == "xray")
			var suffix := "-xray" if mode == "xray" else ""
			for stage in stages:
				studio.renderer.seek(float(stage))
				studio.renderer.set_process(false)
				_set_view("hero")
				await _capture("%s-build-%02d%s" % [name, roundi(float(stage) * 100), suffix])
			studio.renderer.seek(1.0)
			for view in views:
				if not VIEWS.has(view):
					push_warning("Unknown view: " + view)
					continue
				_set_view(view)
				await _capture("%s-%s%s" % [name, view, suffix])
	quit(0)

func _show(name: String) -> bool:
	var preset := Array(StructurePresets.TITLES).map(func(t: String) -> String: return t.to_lower()).find(name)
	if preset >= 0:
		studio.select_preset(preset)
		return true
	var path := StructureToolbar.STRUCTURES_DIR.path_join(name + ".octa")
	if not FileAccess.file_exists(path):
		return false
	studio._load_structure(path)
	return true

func _set_view(view: String) -> void:
	studio.camera.reset_view()
	studio.camera.yaw = VIEWS[view].x
	studio.camera.pitch = VIEWS[view].y
	studio.camera._update_transform()

func _capture(name: String) -> void:
	# Two frames: one to apply the camera and clock, one to finish drawing them.
	await process_frame
	await RenderingServer.frame_post_draw
	var path := "%s/%s.png" % [out_dir, name]
	if root.get_texture().get_image().save_png(path) == OK:
		print("Captured ", ProjectSettings.globalize_path(path))
	else:
		push_error("Failed to save " + path)

func _parse_options() -> Dictionary:
	var result := {}
	for arg in OS.get_cmdline_user_args():
		if not arg.begins_with("--"):
			continue
		var pair := arg.substr(2).split("=", true, 1)
		result[pair[0]] = pair[1] if pair.size() > 1 else "true"
	return result
