# OCTA binary format, versions 1 and 2

[Project overview](../README.md) · [Architecture](architecture.md) · [Generator](../gen/README.md)

All numeric values are little-endian. Files are uncompressed and have no implicit padding. A reader must check the complete length and validate the lattice before rendering.

The two implementations are [OctaFile](../scripts/data/octa_file.gd) with
[OctaStructure](../scripts/data/octa_structure.gd) in GDScript, and
[octa.Structure](../gen/octa/structure.go) in Go. Keep both codecs, their limits,
and their tests aligned when changing the format.

## Version 1: coordinates

| Offset | Size | Value |
| ---: | ---: | --- |
| 0 | 4 bytes | ASCII magic `OCTA` |
| 4 | 4 bytes | Unsigned 32-bit version `1` |
| 8 | 4 bytes | IEEE 754 float32 cell span |
| 12 | 4 bytes | Unsigned 32-bit cell count `N` |
| 16 | 4 bytes | Unsigned 32-bit UTF-8 title byte count `T` |
| 20 | `T` bytes | Title, no terminator |
| `20 + T` | `12 × N` bytes | Signed int32 triples `x, y, z`, in assembly order |

Exact file size: `20 + T + 12 × N` bytes. No trailing bytes are accepted. An empty structure is valid. There is no executable content or Godot Variant/object deserialization.

Constraints used by this implementation:

- `N ≤ 1,000,000`, `T ≤ 1,024`.
- Cell span is finite and between `0.001` and `1000` inclusive.
- Coordinates are integers in `[-1,000,000, 1,000,000]`.
- All three components of each coordinate have the same parity.
- No duplicate triples.

The coordinates are authoritative, including their origin and order. World centers
are `(x, y, z) × span / 2`; the renderer animates cells in file order. Files need not
be sorted bottom-up to be valid, although both preset and Go generators use that
ordering. The viewer recenters and grounds them without modifying the file.
The model contains no mesh, separation setting, animation clock, camera, or render
transforms. Loading a title never regenerates geometry. The title does affect
presentation: its snake case selects a matching `WorldTheme`, with studio as the
fallback. The filename is used for gallery labels, not theme selection.

## Version 2: palette colors

Version 2 adds an optional palette of up to 16 colors and one palette index per cell. Writers emit v2 only when a structure is colored; uncolored structures stay v1, byte for byte.

| Offset | Size | Value |
| ---: | ---: | --- |
| 0 | 4 bytes | ASCII magic `OCTA` |
| 4 | 4 bytes | Unsigned 32-bit version `2` |
| 8 | 4 bytes | IEEE 754 float32 cell span |
| 12 | 4 bytes | Unsigned 32-bit cell count `N` |
| 16 | 4 bytes | Unsigned 32-bit UTF-8 title byte count `T` |
| 20 | 4 bytes | Unsigned 32-bit palette size `P`, `0 ≤ P ≤ 16` |
| 24 | `T` bytes | Title, no terminator |
| `24 + T` | `3 × P` bytes | Palette: sRGB `r, g, b` bytes per color |
| `24 + T + 3P` | `12 × N` bytes | Signed int32 triples `x, y, z`, in assembly order |
| `24 + T + 3P + 12N` | `N` bytes | Palette index per cell, parallel to the coordinates; present only when `P > 0` |

Exact file size: `24 + T + 3 × P + 13 × N` bytes, or `24 + T + 12 × N` when `P = 0`. Every index must be below `P`. Colors are presentation data: they never change the lattice, and readers that only need geometry can skip them. The studio shades palette colors lightly by height; uncolored structures use the studio's base color.

Palette entries contain RGB only: no alpha channel or material/lighting settings
are serialized. Palette indices stay parallel to the cell order. Both writers
select v2 only when the per-cell color array is nonempty; an unused in-memory
palette alone does not force v2. A v2 file with `P = 0` is accepted but rewrites as
v1. Empty data is valid; palette-only empty files are not guaranteed to retain
their palette when loaded and saved through the studio.

The generator in `gen/` (`octa.Palette`) uses one shared 16-color palette: stone, dark stone, sand, brick, terracotta, slate, wood, grass, foliage, water, white, charcoal, gold, glass, navy and red.

## Validation and IO behavior

Readers reject unknown magic/version, oversized counts/palettes/titles, incomplete
or trailing payloads, mixed-parity or out-of-range coordinates, duplicates,
nonfinite/out-of-range spans, and invalid palette indices. The maximum accepted
v2 file size is `24 + 1024 + 48 + 13 × 1,000,000 = 13,001,096` bytes. Title size
is a byte count, not a character count; writers should provide valid UTF-8.
The current implementations do not impose a separate strict UTF-8 validation pass.

Godot's `load_structure()` checks the file length before reading the payload and
returns `null` with `last_error` on failure. Go's `ReadFile()` uses `os.ReadFile()`
before `UnmarshalBinary()` validates, so its initial read is not bounded by the
format cap. `UnmarshalBinary()` replaces its receiver only after successful
validation. GDScript geometry validation and `set_colors()` validate authoring
input; finish color edits through that API before saving rather than mutating the
public palette independently.

Both writers validate geometry/metadata and write a sibling `<path>.tmp`, then
rename it to the destination. Go creates missing parent directories; GDScript
requires an existing destination directory. Godot save failures return an `Error`
and set `last_error`; Go methods return `error`. Because the temporary name is
fixed, writes to one destination should be serialized. These are replacement
writes, not an appendable stream or a crash-durability guarantee. The explicit
version leaves room for future chunking/compression without changing earlier
versions' meaning.

## Minimal writer

Example writer in Python, using only the standard library:

```python
import struct

title = "Two face-sharing cells".encode("utf-8")
cells = [(0, 0, 0), (1, 1, 1)]
with open("two_cells.octa", "wb") as output:
    output.write(struct.pack("<4sIfII", b"OCTA", 1, 1.0, len(cells), len(title)))
    output.write(title)
    for cell in cells:
        output.write(struct.pack("<iii", *cell))
```

The example uses only standard-library Python and writes v1. It uses span 1 and
two hexagonal-face neighbors. The checked-in `structures/neighbors.octa` fixture
contains a central cell and all 14 face neighbors.

## Compatibility checks

From the repository root:

```sh
(cd gen && go test ./octa)
godot --headless --path . --editor --import --quit
godot --headless --path . --script tests/test_core.gd
```

The Go tests check colored round trips, v1 output, bad colors, and truncation.
Godot core checks also mutate headers/coordinates, reject trailing bytes and NaN,
round-trip both versions, and load the checked-in Go-generated colored structures.
After changing Go output, regenerate affected assets before running those Godot
checks; scene compilation tests alone do not update the fixtures.
