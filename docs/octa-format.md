# OCTA binary format, version 1

All numeric values are little-endian. Files are uncompressed and have no implicit padding. A reader must check the complete length and validate the lattice before rendering.

| Offset | Size | Value |
| ---: | ---: | --- |
| 0 | 4 bytes | ASCII magic `OCTA` |
| 4 | 4 bytes | Unsigned 32-bit version, currently `1` |
| 8 | 4 bytes | IEEE 754 float32 cell span |
| 12 | 4 bytes | Unsigned 32-bit cell count `N` |
| 16 | 4 bytes | Unsigned 32-bit UTF-8 title byte count `T` |
| 20 | `T` bytes | Title, no terminator |
| `20 + T` | `12 × N` bytes | Signed int32 triples `x, y, z`, in assembly order |

Exact file size: `20 + T + 12 × N` bytes. No trailing bytes are accepted in v1. An empty structure is valid. There is no executable content or Godot Variant/object deserialization.

Constraints used by this implementation:

- `N ≤ 200,000`, `T ≤ 1,024`.
- Cell span is finite and between `0.001` and `1000` inclusive.
- Coordinates are integers in `[-1,000,000, 1,000,000]`.
- All three components of each coordinate have the same parity.
- No duplicate triples.

The coordinates are authoritative, including their origin and order. The viewer recenters and grounds them for display without modifying the file. The model contains no mesh, color, separation setting, animation clock, camera, or render transforms. Preset names have no semantic effect after loading.

`OctaFile` writes a sibling temporary file and renames it after successful writing. It rejects truncated files, unknown versions, inconsistent counts, invalid parity, duplicate occupancy, and nonfinite spans. The header leaves versioning explicit so a future version can add chunking, palette data, or compression without changing v1's meaning.

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
