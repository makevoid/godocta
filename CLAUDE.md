# Project guidance

This is the shared repository guide for Claude Code and Codex. `AGENTS.md` is a
relative symlink to `CLAUDE.md`; edit this file to keep both entry points in sync.

Octa is a Godot 4 structure viewer and assembly studio, backed by a standalone Go
generator. Cells are regular truncated octahedrons on a body-centered cubic (BCC)
lattice. The native app uses GDScript `class_name` classes and GL Compatibility;
there are no addons, C# assemblies, autoload singletons, or external Go packages.

## Documentation map

- [README.md](README.md): running the app, controls, examples, and current build sizes.
- [Architecture](docs/architecture.md): every application class, data flow, rendering,
  shader contracts, threading, and world themes.
- [Development](docs/development.md): setup, tests, capture options, export, and change recipes.
- [Generator](gen/README.md): CLI, density, Go types, components, and adding scenes.
- [OCTA format](docs/octa-format.md): binary layouts and interoperability constraints.

Read the relevant implementation when making a change; these docs describe the
current design, not a substitute API specification for Godot or Go.

## Commands

Run from the repository root unless stated otherwise. The project declares Godot
4.4 compatibility and has been tested with 4.7.2. `gen/go.mod` requires Go 1.22+.

```sh
bin/run        # launches the app (foreground window): ask the user to run it

# Refresh imports/global class registration, especially after adding or renaming scripts.
godot --headless --path . --editor --import --quit
godot --headless --path . --script tests/test_core.gd
godot --headless --path . --script tests/test_scene.gd

# Subshells keep subsequent commands at the repository root.
(cd gen && go vet -composites=false ./... && go test ./...)
(cd gen && go test ./octa -run TestRoundTripColored)
(cd gen && go test ./scenes -run TestScenesCompile/castle -v)
(cd gen && go run ./cmd/octagen -list)
(cd gen && go run ./cmd/octagen)          # Regenerate every registered build (~2 min).
(cd gen && go run ./cmd/octagen castle)   # Regenerate one build.
(cd gen && go run ./cmd/octapreview castle)  # Headless CPU preview sheet: captures/preview/castle.png

# macOS: real GPU capture, hidden and without taking focus.
tools/capture.sh --shapes=castle,cube --views=hero,front,side,top --modes=solid,xray --at=0.5
tools/capture.sh --shapes=castle,arena --views= --bench
```

Go vet disables only the composite-literal check because the geometric DSL uses
positional structs from `octa`. Godot tests are standalone `SceneTree` scripts with
explicit test calls, not an automatically discovered test framework. They have no
test filter. Success markers are `Core checks: N passed, 0 failed.` and
`Scene integration: PASS`; also inspect output for `SCRIPT ERROR` and parse errors.
The core test uses `/private/tmp/godocta-roundtrip.octa`, so do not run multiple
copies concurrently.

Never use `--headless` for images: it uses a dummy renderer. Preserve the user's
preference that captures do not bring the app to the foreground. The macOS wrapper
uses `open -g -j`, writes PNGs under `captures/`, and logs to `captures/capture.log`.
`GODOT_APP` overrides `/Applications/Godot.app`. Hidden-window FPS is throttled;
use `--bench` for forced draws plus GPU readback, rather than the HUD FPS counter.

## Architecture and ownership

- `project.godot` → `scenes/main.tscn` → `OctaStudio` is the entry path. The scene
  contains only the root; `OctaStudio._ready()` builds the environment, renderer,
  camera, HUD, and file dialogs and connects their signals.
- `scripts/data/`: `OctaLattice` owns coordinate rules; `OctaStructure` owns packed
  data; `OctaStructureBuilder` is mutable authoring; `StructurePresets` creates the
  four primitives; `OctaFile` reads/writes files.
- `scripts/geometry/`: `TruncatedOctahedronMesh` caches one indexed unit mesh with
  72 face vertices, 132 indices, and 44 triangles.
- `scripts/rendering/`: `StructureRenderer` derives instance buffers;
  `AssemblyTimeline` owns timing; `OrbitCamera` owns navigation; `WorldTheme`
  supplies dictionaries consumed by `StudioEnvironment`.
- `scripts/tools/`: `StructureTools` owns brush input, pending edits and commits through
  `OctaEditor` (`scripts/data/`); `BrushGizmo` draws the globe/dome and black hole.
- `scripts/ui/`: `StudioHUD` presents state and emits intent; `ToolPalette` selects
  tools; `StructureToolbar` owns preset buttons and the file gallery; `StudioTheme` builds control styles;
  `UnitGlyph` draws a diagram from the actual unit vertices.
- `gen/`: `octa` supplies geometry, colored grids, ordering, and codecs; `kit`
  supplies reusable components; `scenes` composes registered recipes;
  `cmd/octagen` writes them into `structures/`; `cmd/octapreview` renders `.octa`
  files to PNG on the CPU. Godot ignores this module through `gen/.gdignore`.

## Contracts to preserve

1. **Lattice/data:** all three integer coordinate components share parity, including
   negatives. World center = coordinate × span / 2; each cell has 14 face neighbors.
   Keep limits aligned between Go and GDScript (1,000,000 cells, coordinate magnitude
   ≤1,000,000, span 0.001–1000, palette ≤16, title ≤1024 UTF-8 bytes on disk).
2. **Ownership:** `OctaStructure` copies coordinates and color arrays at its API
   boundaries, but metadata and `set_colors()` remain mutable. Treat a structure as
   read-only while displayed or used by a worker. Call `display()` again after edits.
   `OctaStructureBuilder.build()` does not clear its occupancy dictionary; release
   the builder when done.
3. **File format:** both `scripts/data/octa_file.gd` and `gen/octa/structure.go`
   implement OCTA. Uncolored data writes v1; colored data writes v2. Preserve exact
   payload lengths, palette/index correspondence, original coordinates, and order.
   Storage order is assembly order; display recentering and cell gap are not saved.
4. **Rendering:** one `MultiMeshInstance3D` draws a shared mesh. Each instance has
   16 floats: 12 transform values plus rank, seed, normalized height, and palette
   index (`-1` for uncolored). Mesh UV/UV2 and this layout are shader interfaces.
   Convert palette sRGB values to linear on upload; array uniforms do not do it.
5. **Animation:** `AssemblyTimeline` and `shaders/cell_common.gdshaderinc` share the
   inverse-quadratic release and cubic ease-out formulas. Duration scales with the
   cube root of cell count, clamped to 4–24 seconds. Keep both materials' uniforms
   synchronized through `StructureRenderer._set_parameter()`.
6. **Surface optimization:** `display()` immediately schedules a worker for nonempty
   builds. It drops cells with all 14 neighbors occupied. Only settled solid builds
   with gap ≤0.1 use its result; assembly, X-ray, and wider gaps use the full buffer.
   Workers read supplied data/buffers, results return deferred to the main thread, generation
   IDs discard stale results, and `_exit_tree()` joins outstanding tasks.
7. **Worlds/gallery:** file titles select `WorldTheme` keys through `to_snake_case()`;
   unknown titles and all presets use `studio`. Gallery labels come from filenames,
   sorted alphabetically; only its first five entries get shortcuts 5–9. The current
   generated worlds share `LATE_AFTERNOON` settings plus per-world overrides.
8. **Generator:** scenes are authored in world units. Density is a multiplier over
   span 1: `span = cbrt(1 / density)`; default 2 means span ≈0.794. Finer `Fill`,
   `Carve`, and `Paint` operations also test half-unit-snapped centers to preserve
   authored seams. `FillVolume` uses sampled volume instead. Go assembly orders
   rank within a Y course, always preserving bottom-up layers.

## Making changes

Keep UI coordination in `OctaStudio`, geometry/storage independent of nodes, and
per-cell animation in the shader. Follow existing typed GDScript and `gofmt` Go
style. Do not hand-edit generated import caches; preserve tracked `.uid` files.

For data/format changes, update both codecs, the format document, and relevant
round-trip/corruption checks. For renderer/UI changes, run the Godot tests and use
hidden GPU captures when appearance matters. For generator changes, run Go vet and
tests, regenerate affected checked-in builds at the default density, review them with
`octapreview` (headless; do not launch Godot just to look at a build), then run the
Godot core checks against those files. A new scene also needs the explicit generated
build list in `tests/test_core.gd` updated; its world theme is optional.

Keep the docs linked above current when controls, commands, contracts, or public
types change. `structures/*.octa` are versioned fixtures/assets; `.godot/`,
`captures/`, and `preview/` are generated and ignored. `neighbors.octa` is a small
15-cell format fixture, not a registered Go scene.
