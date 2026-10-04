# Development workflow

[Project overview](../README.md) · [Architecture](architecture.md) · [Generator](../gen/README.md) · [File format](octa-format.md)

## Setup and launch

The project declares Godot 4.4/GL Compatibility in `project.godot` and has been
tested with Godot 4.7.2. Use the standard Godot build; .NET and addons are not
required. Go 1.22+ is optional unless changing the generator. The Go module uses
only the standard library, so there is no dependency installation step.

From the repository root:

```sh
godot --version
godot --headless --path . --editor --import --quit
godot --path .
```

Alternatively open `project.godot` in the editor and press F5. On macOS,
`Launch Octa.command` first uses `godot` on PATH, then
`/Applications/Godot.app/Contents/MacOS/Godot`. It changes into the project directory
before launching. The native window is 1900×1000 points × the display scale (3800×2000 px at 2×), minimum 1100×720, and
shrinks to fit the display unless an explicit `--resolution` is supplied.

After adding or renaming scripts/classes, re-run the import command before
headless tests so Godot's global class cache is current. Keep tracked `.uid` files
and SVG `.import` metadata with their resources. `.godot/` is an ignored cache;
`gen/.gdignore` prevents importing the Go module.

## Validation

```sh
# Repository root; execute after import finishes.
godot --headless --path . --script tests/test_core.gd
godot --headless --path . --script tests/test_scene.gd

# Subshell preserves the repository-root working directory.
(cd gen && go vet -composites=false ./... && go test ./...)
```

| Check | Coverage / expected result |
| --- | --- |
| [test_core.gd](../tests/test_core.gd) | Mesh topology/winding/volume, 14 shared faces, lattice/builder validation, deterministic presets and symmetry, timing, v1/v2 round trips and corrupt payloads, all five checked-in generated builds; prints `Core checks: N passed, 0 failed.` |
| [test_scene.gd](../tests/test_scene.gd) | Instantiates the actual scene, selects presets via buttons, waits for assembly, checks buffers and surface reduction, rapid switching/replay, palette/X-ray, gallery/theme selection, camera and empty data; prints `Scene integration: PASS` |
| [octa/structure_test.go](../gen/octa/structure_test.go) | Colored binary round trip, uncolored v1 output, invalid colors, truncated payload |
| [scenes/scenes_test.go](../gen/scenes/scenes_test.go) | Compiles each registered scene at default density; validates count, bottom-up ordering and ≥3 colors |
| `go vet -composites=false ./...` | Static checks, with positional geometric struct literals allowed |

The Godot scripts return nonzero on failed checks. Also inspect `SCRIPT ERROR` and
parse-error output: a failed dependency can invalidate a script before its summary.
Core tests use the fixed `/private/tmp/godocta-roundtrip.octa` path and remove it;
run one copy at a time. Scene tests wait through four real-time assemblies (roughly
16 seconds plus setup). There is no Godot test-filter argument; the explicit calls
in `_initialize()`/`_run()` define coverage. Go supports focused checks:

```sh
(cd gen && go test ./octa -run TestRoundTripColored -v)
(cd gen && go test ./scenes -run TestScenesCompile/castle -v)
```

Headless tests validate logic and stored buffers. Per-instance GPU getters are
skipped under the dummy renderer; visual/shader changes need real GPU inspection.
The scene script also has a `--capture` mode, but the background capture wrapper
below is the normal workflow for images.

Godot may report macOS certificate-store, log, or editor-settings errors inside a
filesystem sandbox even when imports/tests complete. Identify the reported path
and check the test summary separately; these messages do not imply script failures.
Do not dismiss `SCRIPT ERROR`, missing imports, or failed assertions as sandbox noise.

If Go's normal build cache is outside a sandbox's writable paths, use a temporary
cache for the command, for example
`(cd gen && GOCACHE=/tmp/godocta-go-cache go test ./...)`. This does not change the
module or your global Go configuration.

## Headless build previews

To look at a generated build without Godot, use the pure-Go CPU renderer. It draws
the exposed faces of every cell with flat shading and sun shadows over a neutral
sky and ground, in about a second, with no window or GPU:

```sh
(cd gen && go run ./cmd/octapreview castle)                       # captures/preview/castle.png
(cd gen && go run ./cmd/octapreview -views hero -size 1800x1200 arena)
(cd gen && go run ./cmd/octapreview -out ../captures/preview/tmp ../structures/castle.octa)
```

The default output is a 2×2 sheet: hero and side views on the top row, front and
top below, using the studio camera angles. Bare names resolve to
`structures/<name>.octa`. `-views` picks any of `hero,side,front,top`, `-size` sets
each view in pixels and `-ss` sets supersampling. Use it to review generator changes;
it ignores world themes, cell gap, X-ray and assembly animation, which need the GPU
captures below.

## Screenshots and benchmarks

On macOS, use the wrapper so rendering does not take focus:

```sh
tools/capture.sh --shapes=cube,castle --views=hero
tools/capture.sh --shapes=castle --views=hero,front,side,top --modes=solid,xray
tools/capture.sh --shapes=arena --at=0.25,0.5,0.75
tools/capture.sh --shapes=castle,arena --views= --bench
```

`tools/capture.sh` changes to the repo root, launches a new Godot app instance with
`open -g -j -n -W`, requests 1900×1000 at an offscreen position (3800×2000 px on a 2× display), and waits for exit.
It uses a real GPU; `--headless` cannot produce screenshots. Set `GODOT_APP` to a
different `.app` path if necessary. The wrapper itself requires macOS and zsh.

| Capture argument | Default / behavior |
| --- | --- |
| `--shapes=cube,castle` | All four presets plus every gallery filename stem, including `neighbors` |
| `--views=hero,front,top` | These three settled views; `side` is also supported; empty value disables settled screenshots |
| `--modes=solid,xray` | Default `solid`; `--xray` is a shorthand when `--modes` is absent |
| `--at=0.25,0.6` | Additional frozen mid-assembly frames at fractions of total duration, always hero view |
| `--gap=0` | Renderer gap, default 0.045; shader positions change, saved data does not |
| `--out=captures` | PNG destination; relative paths resolve under `res://`, absolute paths also work |
| `--bench` | Measure settled hero frames, waiting up to 20 seconds for surface extraction |

Use `--name=value` syntax; the capture parser does not consume a separate value
argument. Hero uses the application's default framing, including lower pitch for
themed builds. Front/side/top override the camera angles.

PNG names are `<shape>-<view>[-xray].png` and
`<shape>-build-<percent>[-xray].png`. Existing names are overwritten. The log always
stays at `captures/capture.log` even with a custom `--out`; it is reset per run.
The wrapper prints Showing/Captured/Bench/error/warning lines and reports failure
if its log contains `ERROR`. Unknown views are warnings and are skipped; unknown
structures are errors. Check requested PNGs as well as exit status.

Hidden windows throttle ordinary frames, so HUD FPS is unsuitable for benchmarks.
Compatibility rendering does not supply useful GPU timing here. The benchmark
forces 20 draws with texture readback, reporting mean milliseconds, approximate
FPS, total cells, and drawn cells. Readback overhead is included; compare runs at
the same resolution, mode, gap, and hardware. This is not a pure GPU timer.

## Typical changes

| Change | Files to consider | Verification |
| --- | --- | --- |
| Add a generated build | `gen/scenes/`, optional `gen/kit/`, output in `structures/`, explicit generated-build list in `tests/test_core.gd`; optional `WorldTheme` entry | Go vet/tests, regenerate, `octapreview`, Godot core |
| Add a primitive preset | `StructurePresets` enum/titles/descriptions/generation, toolbar icon/button loop, `OctaStudio` key handling, HUD help, tests | Update hard-coded four-preset loops, core + scene, capture |
| Change cell mesh or edges | `TruncatedOctahedronMesh`, `cell_common.gdshaderinc`, both cell materials; `UnitGlyph` derives geometry | Topology/volume/shared-face checks, scene, solid + X-ray captures |
| Change timing | `AssemblyTimeline` and shared shader formulas/uniforms; renderer | Timing and scene tests, partial captures via `--at` |
| Change environment | `WorldTheme`, `StudioEnvironment`, sky/ground shaders | Scene theme wiring, themed hero/side captures |
| Change data or format | Both codecs/validators, `docs/octa-format.md`, relevant GDScript/Go tests | Both language suites and checked-in generated files |
| Change brush tools | `OctaEditor`, `StructureTools`, `BrushGizmo`, `ToolPalette`, `brush`/`black_hole` shaders, `OctaStudio` keys | Core editor checks, scene `_test_tools()`, hidden GPU capture of the gizmo |
| Change controls | `StudioHUD`, `StructureToolbar`, `StudioTheme`, `OrbitCamera`, `OctaStudio` signal wiring | Scene tests, native input/dialog inspection, README controls |

For a new Go scene, follow the complete example in the [generator guide](../gen/README.md).
`go test ./...` compiles recipes but does not rewrite `.octa` assets. Regenerate
affected scenes explicitly at the default density, then run Godot core checks to
exercise the other language's reader. Update build counts in the README if output
changes. Use `-out /tmp/octa-builds` for experiments that should not replace assets.

For a new world, match the snake-cased embedded title in `WorldTheme.themes()`;
filenames do not choose the theme. Reuse `_late_afternoon()` if appropriate and
provide all required fields listed in the [architecture reference](architecture.md).
Any new theme parameter must also be wired through `StudioEnvironment.apply_theme()`
and its target shader. Unknown titles continue to use studio lighting.

## Optional Web Preview export

`export_presets.cfg` includes a Web Preview preset with thread support disabled.
Place the matching Godot version's `web_nothreads_debug.zip` template at
`.godot/web_nothreads_debug.zip`, then run:

```sh
mkdir -p preview
godot --headless --path . --export-debug "Web Preview" preview/index.html
python3 -m http.server 8000 --directory preview
```

Open `http://localhost:8000` while the local server is running. The export includes
`structures/*.octa` and excludes tests/docs/captures/preview. The template and
`preview/` output are ignored; do not commit them. Use the native application for
local filesystem Save/Load. Native headless tests do not establish browser behavior;
check an exported preview separately when changing web support.

## Documentation and agent instructions

[CLAUDE.md](../CLAUDE.md) is the canonical agent guide. The root
[AGENTS.md](../AGENTS.md) is a relative symlink to it, giving both tools one source
of instructions without changing a developer's global configuration. Codex
discovers `AGENTS.md` by default; alternative names require a configured fallback.
See the [official instruction-discovery documentation](https://learn.chatgpt.com/docs/agent-configuration/agents-md).

Keep the overview focused on users, the architecture reference on runtime contracts,
this guide on commands/maintenance, and the generator guide on authoring APIs.
Update the format specification whenever serialization changes. Document nonobvious
invariants beside code using GDScript `##` or Go doc comments; avoid duplicating
whole implementations in prose.
