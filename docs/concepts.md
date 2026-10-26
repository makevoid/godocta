# How Octa works

A short tour of the data model, the scripting API and the renderer. The
[architecture reference](architecture.md) has the full class inventory, buffer
layouts and shader contracts.

## The cell is the unit of data

An `OctaStructure` contains a `PackedInt32Array` of `(x, y, z)` triples, a title, and a cell span. Coordinates occupy **12 bytes per cell**. Colored structures add a palette of up to 16 colors and **one palette-index byte per cell** (`set_colors()`, `color_index()`). Meshes, scene nodes, transforms and animation state are not stored per cell in the structure.

Truncated octahedrons tile on a **body-centered cubic lattice**. Coordinate triples must be **all even or all odd**, including negative coordinates. Their world-space centers are:

```text
world_position = Vector3(x, y, z) * cell_span / 2
```

For `cell_span = 1`, the unit has vertices at the permutations of `(0, ±1, ±2) / 4`. Its axis-aligned span is one unit, its volume is 0.5, and its edge length is `sqrt(2) / 4`. `OctaLattice.NEIGHBORS` provides all 14 face-sharing offsets: six axial `(±2, 0, 0)` permutations and eight diagonal `(±1, ±1, ±1)` combinations.

**The cube's dimension of 10 means a 10 × 10 × 10 world-unit envelope**, measured with zero visual separation. Its lattice contains `10³ + 9³ = 1,729` truncated octahedrons, using 20,748 bytes of coordinates. The sphere and cylinder keep a cell when more than half of its volume lies inside the shape, so silhouettes are symmetric and free of single-cell spurs. The pyramid steps in by half a cell per course, the lattice's natural 45° slope. Boundaries are cellular approximations, not clipped meshes.

| Preset | Cells | Coordinate bytes | Cell envelope |
| --- | ---: | ---: | --- |
| Cube | 1,729 | 20,748 | 10 × 10 × 10 |
| Sphere | 1,037 | 12,444 | 10 × 10 × 10 |
| Pyramid | 385 | 4,620 | 10 × 5.5 × 10 |
| Cylinder | 1,421 | 17,052 | 10 × 10 × 10 |

## Specify your own arrangement

Use packed triples directly for generated/imported data:

```gdscript
var arrangement := OctaStructure.from_cells(
    PackedInt32Array([
        0, 0, 0,
        2, 0, 0,  # Square-face neighbor.
        1, 1, 1,  # Hexagonal-face neighbor.
    ]),
    "Three connected cells",
    1.0
)
renderer.display(arrangement)  # Four-second assembly.
```

Or use the mutable authoring helper, which checks occupancy:

```gdscript
var builder := OctaStructureBuilder.new()
builder.add(Vector3i.ZERO)
for offset in OctaLattice.NEIGHBORS:
    builder.add(offset)
var arrangement := builder.build("Central cell and its 14 neighbors")

var files := OctaFile.new()
files.save_structure("user://neighbors.octa", arrangement)
var restored := files.load_structure("user://neighbors.octa")
renderer.display(restored)
```

Invalid parity, duplicate cells, invalid spans, and malformed payloads are rejected. `from_cells()` returns `null` for invalid input; `OctaStructure.validation_error()` explains why. `OctaFile.last_error` explains save/load failures. Discard the builder after compiling an arrangement to release its temporary occupancy dictionary; `build()` itself does not clear it. Finish metadata and color edits before display, and treat displayed structures as read-only while the renderer's worker uses them. A structure can contain up to 1,000,000 cells. This cap bounds accepted data; it is not a performance guarantee on every GPU.

## Rendering and animation

`StructureRenderer` uploads one derived buffer to a `MultiMesh`, sharing one indexed 44-triangle mesh (72 vertices). There is **one render node per arrangement**. Each instance has 12 transform floats and 4 custom floats (rank, seed, height, palette index; 64 bytes); Godot and the GPU also require their own resource/storage overhead. The 12-byte figure describes the compact coordinate payload, not total process memory.

A single material uniform advances the animation clock. `AssemblyTimeline.duration_for(cells)` is `clamp(4 × ∛(cells / 1729), 4, 24)` seconds. Instance rank determines its release time over the release window; each cell settles using cubic ease-out over about 1.2 seconds (slightly longer for big builds), and the last cell settles exactly at the end. Big builds drop cells from proportionally higher so the motion stays visible. Completed assemblies stop requesting per-frame render updates. Switching shapes replaces the arrangement and resets that clock; replay reuses the current buffer.

A background task starts when a structure is displayed and finds cells whose 14 neighbors are all occupied. Once an assembly settles in solid view with cell separation at or below 0.1 — zero in the app, so cells tile face to face — the renderer uses the resulting surface buffer when ready. Assembly, X-ray and wider separation draw every cell. The full dataset remains available for saving and replay.

`WorldTheme` defines each world (sky colors, sun, clouds, horizon silhouette, ground or water, light and ambient). The generated worlds share warm late-afternoon lighting with per-world overrides. `StudioEnvironment.apply_theme()` switches them; file titles select themes through snake case, unknown titles fall back to studio, and presets always use the dark studio. The sky is a static procedural shader (`shaders/sky.gdshader`) without a time dependency, so it does not animate its radiance every frame.

The solid and X-ray materials share `shaders/cell_common.gdshaderinc`. The X-ray material is additive and unshaded with no depth writes, so interiors show through; lines and vertex dots fade with screen-space density so distant views don't saturate.

## Editing

Brush edits are pending while you paint: swallowed cells are hidden in the full buffer and new cells are drawn by an overlay. About a second after a stroke (or before another stroke, switching tools, Save or Replay), `OctaEditor.build()` compacts them into a new structure that replaces the old one in place, without recentering or replaying. Survivors keep their assembly order and added cells follow. Undo/Redo keeps the last 10 strokes and restores exact coordinates, order and colors without moving the view. Switching builds discards pending edits and history.

Geometry, storage, rendering, presets, UI, and camera are separate `class_name` classes under `scripts/`. The entry scene is `scenes/main.tscn`; `OctaStudio` connects their signals. Cell separation and display recentering do not change the saved coordinates.

For much larger worlds, partition structures into spatial chunks with one `MultiMesh` per chunk. A single `MultiMesh` is culled as a unit. The current implementation uses one render node for each displayed build. See the [architecture reference](architecture.md) for buffer layout, shader contracts and worker lifecycle.

## File format

[OCTA format](octa-format.md) describes the small, versioned, little-endian binary format: v1 holds coordinates, v2 adds a palette and per-cell colors. Save/Load and the Go toolkit both implement it.
