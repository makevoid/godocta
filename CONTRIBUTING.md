# Contributing to Octa

Bug fixes, documentation improvements, new structures, and focused tool improvements
are welcome. Use [issues](https://github.com/makevoid/godocta/issues) to discuss larger
changes before implementing them. Please keep feedback specific and respectful.

## Start here

1. Fork and clone the repository, then create a branch for your change.
2. Install standard Godot 4.4+ (4.7.2 is the locally tested version). Go 1.22+ is
   needed for generator changes; Python 3.9+ runs the combined check command.
3. Open `project.godot` and press F5, or use `bin/run` on macOS/Linux.
4. Read the relevant implementation and the [architecture](docs/architecture.md).

The viewer works offline with the checked-in builds and artwork. You do not need an
AI account, API key, Ruby, or Go to run it. AI artwork regeneration is optional and
submits paid requests; see [development](docs/development.md#ui-artwork).

## Validate your change

From the repository root:

```sh
python3 tools/check.py
python3 tools/check.py --only godot
python3 tools/check.py --only go
```

On Windows use `python` instead of `python3`. Set `GODOT` or pass
`--godot /path/to/Godot` if the engine is not on PATH. The check command imports
resources first, requires both Godot success summaries, rejects engine/script
errors, and runs Go formatting checks, vet, and tests. It does not regenerate assets.
Individual commands remain documented in [development](docs/development.md#validation).

For UI/rendering changes, also inspect real GPU captures. Headless Godot uses a
dummy renderer. On macOS, `tools/capture.sh` captures without taking focus.
For scene recipes, regenerate only affected builds at default density, inspect
them with `octapreview`, and run the Godot core suite against the resulting files.

## Keep changes reviewable

- Follow typed GDScript and `gofmt` Go style. Keep UI coordination in `OctaStudio`
  and geometry/storage independent of nodes.
- Preserve the [lattice and format contracts](docs/octa-format.md). Format changes
  require both codecs, their checks, and the specification to change together.
- Include `.uid` and asset `.import` metadata. Do not commit `.godot/`, captures,
  preview/export output, credentials, or local editor configuration.
- Include default-density `.octa` assets when changing their recipes. A new scene
  also needs the explicit build list in `tests/test_core.gd` updated.
- Update docs when controls, commands, or public APIs change. Include focused tests
  for behavior changes and regression fixes.
- Describe the problem, resulting behavior, and checks in the pull request.

Contributions are provided under the repository's [MIT license](LICENSE).
Only contribute code and assets you have permission to distribute. Document external
sources and their licenses in [asset and dependency credits](docs/credits.md);
retain existing notices. No contributor license agreement is required.
