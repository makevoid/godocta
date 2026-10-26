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
`Launch Octa.command` delegates to `bin/run`, which first uses `godot` on PATH,
then `$GODOT_APP/Contents/MacOS/Godot` (default `/Applications/Godot.app`). The
launcher uses POSIX sh, works on Linux with Godot on PATH, and imports resources
on first launch. On Windows, use the editor or `godot --path .`. The native window is 1900×1000 points × the display scale (3800×2000 px at 2×), minimum 1100×720, and
shrinks to fit the display unless an explicit `--resolution` is supplied.

After adding or renaming scripts/classes, re-run the import command before
headless tests so Godot's global class cache is current. Keep tracked `.uid` files
and SVG `.import` metadata with their resources. `.godot/` is an ignored cache;
`gen/.gdignore` prevents importing the Go module.

## Validation

With Python 3.9+, `python3 tools/check.py` runs import, both Godot suites, Go
formatting checks, vet, and tests. Use `--only godot` or `--only go` for one
toolchain and `--godot /path/to/Godot` for a custom executable. On Windows use
`python`. A missing success summary, engine/script error, timeout, or nonzero
exit fails the command. Logs use a temporary directory. The commands below can
also be run separately.

GitHub Actions is **disabled**. The prepared workflow at
`.github/workflows/ci.yml.disabled` is not discovered by GitHub. It defines Linux
checks with Godot 4.4.1 and 4.7.2, plus Go checks on Linux/macOS/Windows and the
minimum Go 1.22 toolchain. It uses read-only permissions, pinned actions, and
official Godot archives verified against their release checksums. See
[releasing](releasing.md#enabling-ci-later) before activating it.

```sh
# Repository root; execute after import finishes.
godot --headless --path . --script tests/test_core.gd
godot --headless --path . --script tests/test_scene.gd

# Subshell preserves the repository-root working directory.
(cd gen && go vet -composites=false ./... && go test ./...)
```

| Check | Coverage / expected result |
| --- | --- |
| [test_core.gd](../tests/test_core.gd) | Mesh topology/winding/volume, 14 shared faces, lattice/builder validation, deterministic presets and symmetry, timing, v1/v2 round trips and corrupt payloads, all 85 checked-in generated builds; prints `Core checks: N passed, 0 failed.` |
| [test_scene.gd](../tests/test_scene.gd) | Instantiates the actual scene, checks welcome/collection actions, search and overlay input isolation, selects presets via buttons, waits for assembly, checks buffers and surface reduction, rapid switching/replay, palette/X-ray, gallery/theme selection, camera and empty data; prints `Scene integration: PASS` |
| [octa/structure_test.go](../gen/octa/structure_test.go) | Colored binary round trip, uncolored v1 output, invalid colors, truncated payload |
| [scenes/scenes_test.go](../gen/scenes/scenes_test.go) | Compiles each registered scene at default density; validates count, bottom-up ordering and ≥3 colors |
| `go vet -composites=false ./...` | Static checks, with positional geometric struct literals allowed |

The Godot scripts return nonzero on failed checks. Also inspect `SCRIPT ERROR` and
parse-error output: a failed dependency can invalidate a script before its summary.
Core tests use a process-specific file under `OS.get_temp_dir()` and remove it;
they work across desktop platforms without a macOS-specific path. Run imports
and suites sequentially within a checkout. Scene tests wait through four real-time assemblies (roughly
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

`-video` renders a showcase MP4 through `ffmpeg` instead (1920×1080 at 30 fps by
default, about 2.5 minutes for three 100k-cell builds on ten cores). Each build gets a
4-second, 150° low orbit around its center of mass. The orbit bursts fast around each
cut and drifts slowly mid-shot while the final second of assembly lands. Shots meet at
the same spin speed, so each build swaps for the next under rotational and sideways
motion blur while the camera keeps turning; a circular mask keeps the screen edges sharp. Cells use the shader's flight motion, outlines and the world
theme's sky and ground:

```sh
(cd gen && go run ./cmd/octapreview -video ../captures/showcase.mp4 shwedagon_pagoda rothenburg blue_mosque)
```

`-shot`, `-turn`, `-swap`, `-left`, `-fps` and `-samples` (minimum sub-frames per frame;
fast orbits take more) tune it.
Video renders lower CPU priority on macOS/Linux and run on half the cores (`-workers`) under a soft
memory cap (`-mem`, GiB), and load each build only around its shot, so a 30-build
showcase takes about an hour in under 1 GB without blocking the machine. Each shot is
encoded to `<output>.parts/` first: after an interruption, rerun the same command to
resume from the last finished shot; the parts are joined and removed at the end.

## UI walkthrough video

`tools/walkthrough.sh` records a scripted tour of about 80 seconds of the real app to
`captures/walkthrough.mp4` (1920×1080, 30 fps, H.264 through Homebrew `ffmpeg`): welcome
screen, a primitive from the menu, Brush in with the radius slider, Black hole, Undo and
Redo, orbit and zoom, Load through the file dialog, a large black hole biting into a
castle tower, the build library search and a red Brush in dot on the Taj Mahal. Key
clicks are checked against the UI and retried, with a warning in the log, if one is lost. Like the
captures below, Godot runs hidden in the background and never takes focus; a run takes
about a minute.

```sh
tools/walkthrough.sh
tools/walkthrough.sh --out=captures/tour.mp4 --fast=1   # Loaded builds assemble in real time.
tools/walkthrough.sh --stills=captures/readme           # Also save clean PNG stills.
```

`--stills` saves 1920×1080 PNGs of the studio alone, without the pointer or captions,
at key moments of the tour (welcome, primitives, brush in, black hole, castle, library,
Taj Mahal, paint). The README screenshots come from these stills and from
`tools/capture.sh --views=hero` gallery shots. The README hosts its images and a
720p cut of the tour (GitHub's limit is 10 MB) as user-attachment uploads made in
GitHub's file editor, so they are not committed. To replace one, drop the new file
into the README editor on GitHub and swap the URL.

`tools/walkthrough.gd` hosts the studio in an offscreen `SubViewport` at 2× and pushes
mouse and key events into it (a root window would read the real OS pointer), so menus,
tooltips, brushes and dialogs react exactly as they do for a user. It composites a drawn
pointer, click ripples and captions on top and streams frames to `ffmpeg`. With
`--fixed-fps` each frame advances exactly 1/fps seconds regardless of render speed.
`--fast` (default 2) speeds up the assembly of the loaded builds. The
file dialog shows the absolute path of `structures/`. Edit the scene functions in the
script to change the tour; controls are located from the live UI, not fixed coordinates.

## Screenshots and benchmarks

On macOS, use the wrapper so rendering does not take focus:

```sh
tools/capture.sh --shapes=cube,castle --views=hero
tools/capture.sh --shapes=castle --views=hero,front,side,top --modes=solid,xray
tools/capture.sh --shapes=arena --at=0.25,0.5,0.75
tools/capture.sh --shapes=castle,arena --views= --bench
tools/capture.sh --ui --shapes=cube --views=hero --out=captures/aqua-ui
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
| `--gap=0.045` | Renderer gap, default 0 (face-to-face); shader positions change, saved data does not |
| `--out=captures` | PNG destination; relative paths resolve under `res://`, absolute paths also work |
| `--bench` | Measure settled hero frames, waiting up to 20 seconds for surface extraction |
| `--ui` | Also capture welcome, collection, brush options, help, settings and an edit/undo/redo sequence; repeat welcome, collection and expanded options at the 1100×720 minimum window size |

Use `--name=value` syntax; the capture parser does not consume a separate value
argument. Hero uses the application's default framing, including lower pitch for
themed builds. Front/side/top override the camera angles.

PNG names are `<shape>-<view>[-xray].png` and
`<shape>-build-<percent>[-xray].png`. Existing names are overwritten. The log always
stays at `captures/capture.log` even with a custom `--out`; it is reset per run.
The wrapper prints Showing/Captured/Bench/error/warning lines and reports failure
if launching Godot fails or its log contains `ERROR`. Unknown views are warnings and are skipped; unknown
structures are errors. Check requested PNGs as well as exit status.

Hidden windows throttle ordinary frames, so HUD FPS is unsuitable for benchmarks.
Compatibility rendering does not supply useful GPU timing here. The benchmark
forces 20 draws with texture readback, reporting mean milliseconds, approximate
FPS, total cells, and drawn cells. Readback overhead is included; compare runs at
the same resolution, mode, gap, and hardware. This is not a pure GPU timer.

## UI artwork

The checked-in splash and button frame were generated with the project-local
[sunburst-images skill](../.agents/skills/sunburst-images/SKILL.md). Runtime use
requires only Godot; Ruby, `excon` (`gem install excon`), and a fal.ai credential are needed only when
regenerating artwork:

```sh
ruby tools/generate_ui_art.rb
godot --headless --path . --editor --import --quit
```

This submits two paid image requests. Set `FAL_KEY` in your environment or use the
home-directory credential files documented by the optional tooling; never commit
credentials. Edit `assets/ui/prompts/*.txt` first; the exact
prompt, settings and request ID are also retained in the generated JSON sidecars.
The regeneration script makes sidecar image paths repository-relative. See
[credits](credits.md) for the asset provenance inventory.
On macOS the script uses `sips` to resize the button master to the 192×64 runtime
skin; other platforms need that resize separately. Keep the ten-pixel nine-slice
corners intact when changing `StudioTheme.chrome()`. Button labels and states are
native Godot controls, and small SVG icons stay sharp at different display scales.

## Typical changes

| Change | Files to consider | Verification |
| --- | --- | --- |
| Add a generated build | `gen/scenes/`, optional `gen/kit/`, output in `structures/`, explicit generated-build list in `tests/test_core.gd`; optional `WorldTheme` entry | Go vet/tests, regenerate, `octapreview`, Godot core |
| Add a primitive preset | `StructurePresets` enum/titles/descriptions/generation, the toolbar's Primitives menu loop, `OctaStudio` key handling, HUD help card, tests | Update hard-coded four-preset loops, core + scene, capture |
| Change cell mesh or edges | `TruncatedOctahedronMesh`, `cell_common.gdshaderinc`, both cell materials; `UnitGlyph` derives geometry | Topology/volume/shared-face checks, scene, solid + X-ray captures |
| Change timing | `AssemblyTimeline` and shared shader formulas/uniforms; renderer | Timing and scene tests, partial captures via `--at` |
| Change environment | `WorldTheme`, `StudioEnvironment`, sky/ground shaders | Scene theme wiring, themed hero/side captures |
| Change data or format | Both codecs/validators, `docs/octa-format.md`, relevant GDScript/Go tests | Both language suites and checked-in generated files |
| Change brush tools | `OctaEditor`, `StructureTools`, `BrushGizmo`, `ToolPalette`, `ToolOptions`, `OctaPalette`, `brush`/`black_hole` shaders, `OctaStudio` keys | Core editor checks, scene `_test_tools()` and `_test_undo_redo()` (including the 10-stroke limit), hidden GPU capture of the gizmo and history controls |
| Change controls | `StudioHUD` (readout, assembly line, help card), `StructureToolbar` (bar and settings menu), `ToolOptions`, `StudioTheme`, `OrbitCamera`, `OctaStudio` signal wiring | Scene tests, native input/dialog inspection, [controls](controls.md) |

For a new Go scene, follow the complete example in the [generator guide](../gen/README.md).
`go test ./...` compiles recipes but does not rewrite `.octa` assets. Regenerate
affected scenes explicitly at the default density, then run Godot core checks to
exercise the other language's reader. Update cell counts in [builds](builds.md) if output
changes. Use `-out /tmp/octa-builds` for experiments that should not replace assets.

For a new world, match the snake-cased embedded title in `WorldTheme.themes()`;
filenames do not choose the theme. Reuse `_late_afternoon()` if appropriate and
provide all required fields listed in the [architecture reference](architecture.md).
Any new theme parameter must also be wired through `StudioEnvironment.apply_theme()`
and its target shader. Unknown titles continue to use studio lighting.

## Optional Web Preview export

`export_presets.cfg` includes an experimental Web Preview preset with thread support
disabled. Install the official export templates matching your engine version using
**Editor → Manage Export Templates**. The preset uses the standard installed
templates, with no local custom-template path. Then run:

```sh
mkdir -p preview
godot --headless --path . --export-debug "Web Preview" preview/index.html
python3 -m http.server 8000 --directory preview
```

Open `http://localhost:8000` while the local server is running. The export includes
`structures/*.octa` and excludes development tools, tests, docs, captures, export
output, artwork prompts/sidecars, and the unused button master. The templates and
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
