# Octa · Structure Studio

A class-based Godot 4 project for assembling structures from regular truncated octahedrons. The top toolbar builds a cube, sphere, square pyramid, or cylinder. Every selection and replay runs a **4-second assembly**, with a decelerating release wave and cubic ease-out motion for each cell.

## Run

Open `project.godot` in Godot 4.4 or newer and press **F5**, or double-click **Launch Octa.command** on macOS. The window defaults to **1500 × 960**.

```sh
godot --path . --resolution 1500x960
```

Tested with Godot **4.7.2**. No addons, external packages, or C# runtime are required.

| Control | Action |
| --- | --- |
| Cube / Sphere / Pyramid / Cylinder, or **1–4** | Assemble a preset |
| Replay, **R**, or **Space** | Restart the current arrangement |
| Drag | Orbit |
| Wheel / pinch | Zoom |
| Shift + drag / middle drag | Pan |
| Reset view / **F** | Restore framing |
| Auto orbit | Slowly rotate the view |
| Cell separation | Reveal individual cells; zero gives exact face-to-face tiling |
| Save / Load | Write or open a compact `.octa` arrangement |

## The cell is the unit of data

An `OctaStructure` contains a `PackedInt32Array` of `(x, y, z)` triples, a title, and a cell span. Coordinates occupy **12 bytes per cell**. Meshes, scene nodes, transforms, animation state, and colors are not stored per cell in the structure.

Truncated octahedrons tile on a **body-centered cubic lattice**. Coordinate triples must be **all even or all odd**, including negative coordinates. Their world-space centers are:

```text
world_position = Vector3(x, y, z) * cell_span / 2
```

For `cell_span = 1`, the unit has vertices at the permutations of `(0, ±1, ±2) / 4`. Its axis-aligned span is one unit, its volume is 0.5, and its edge length is `sqrt(2) / 4`. `OctaLattice.NEIGHBORS` provides all 14 face-sharing offsets: six axial `(±2, 0, 0)` permutations and eight diagonal `(±1, ±1, ±1)` combinations.

**The cube's dimension of 10 means a 10 × 10 × 10 world-unit envelope**, measured with zero visual separation. Its lattice contains `10³ + 9³ = 1,729` truncated octahedrons, using 20,748 bytes of coordinates. The other presets select whole cells within their shape; their boundaries are cellular approximations, not clipped meshes.

| Preset | Cells | Coordinate bytes | Cell envelope |
| --- | ---: | ---: | --- |
| Cube | 1,729 | 20,748 | 10 × 10 × 10 |
| Sphere | 749 | 8,988 | 9 × 9 × 9 |
| Pyramid | 589 | 7,068 | 10 × 9.5 × 10 |
| Cylinder | 1,221 | 14,652 | 9 × 10 × 9 |

### Specify your own arrangement

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

Invalid parity, duplicate cells, invalid spans, and malformed payloads are rejected. `from_cells()` returns `null` for invalid input; `OctaStructure.validation_error()` explains why. `OctaFile.last_error` explains save/load failures. Discard the builder after compiling an arrangement to release its temporary occupancy dictionary. A structure can contain up to 200,000 cells. This cap bounds loading; it is not a performance guarantee on every GPU.

### Rendering and animation

`StructureRenderer` uploads one derived buffer to a `MultiMesh`, sharing one 44-triangle mesh. There is **one render node per arrangement**. Each instance has 12 transform floats and 4 custom floats (64 bytes); Godot and the GPU also require their own resource/storage overhead. The 12-byte figure describes the compact coordinate payload, not total process memory.

A single material uniform advances the animation clock. Instance rank determines its release time over 2.8 seconds; each cell settles over 1.2 seconds using cubic ease-out. The last cell settles at 4.0 seconds. Completed assemblies stop requesting per-frame render updates. Switching shapes replaces the arrangement and resets that clock; replay reuses the current buffer.

Geometry, storage, rendering, presets, UI, and camera are separate `class_name` classes under `scripts/`. The entry scene is `scenes/main.tscn`; `OctaStudio` connects their signals. Cell separation and display recentering do not change the saved coordinates.

For much larger worlds, partition structures into spatial chunks with one `MultiMesh` per chunk. A single `MultiMesh` is culled as a unit. The current implementation deliberately keeps the preset-scale renderer simple. See the official [MultiMesh API](https://docs.godotengine.org/en/stable/classes/class_multimesh.html).

## File format and checks

[OCTA v1 format](docs/octa-format.md) describes the small, versioned, little-endian binary format. It is already implemented by Save/Load, so future external tools can produce the same arrangements without constructing Godot resources.

```sh
godot --headless --path . --editor --import --quit
godot --headless --path . --script tests/test_core.gd
godot --headless --path . --script tests/test_scene.gd
```

Core checks cover topology, all 14 shared faces, mesh volume/winding, lattice validation, presets, exact timing/easing, file round trips, and corrupt files. Scene checks exercise toolbar wiring, replay/interruption, render buffers, camera framing, and empty arrangements.

With desktop access, capture the actual GPU rendering at 1500 × 960:

```sh
godot --path . --resolution 1500x960 --script tests/test_scene.gd -- --capture
```

This saves all four completed presets and an assembly frame under `captures/`. Headless tests use Godot's dummy rendering driver and do not validate GPU pixels. Native window launch and browser inspection were unavailable in the creation session; visual QA remains to be performed on a desktop.

An optional Web Preview export preset is included. Place the matching engine version's `web_nothreads_debug.zip` template in `.godot/`, then run `godot --headless --path . --export-debug "Web Preview" preview/index.html`. `preview/` is a generated, ignored folder. Use the native project for local filesystem Save/Load.
