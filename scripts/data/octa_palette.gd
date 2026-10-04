class_name OctaPalette
extends RefCounted
## The 16-color base palette, mirroring `octa.Palette` in `gen/octa/palette.go`. Generated
## builds embed their own copy, so files stay self-describing and `OctaStructure.palette`
## is what the renderer and the brush use; this table only names the indices and supplies
## swatches for the UI. Keep the hex values and names in step with the Go source.

## sRGB hex per index, in the Go order. Constant color literals cannot cross class
## boundaries in GDScript, so the table stores plain integers and builds colors on demand.
const HEX: Array[int] = [
	0xb9b3a4, 0x7b766c, 0xdbc593, 0xa95540,
	0xcc7a4a, 0x4f6377, 0x7c5635, 0x6f9b4b,
	0x3e6b3a, 0x3c7da8, 0xeeeee6, 0x35393e,
	0xd9aa3c, 0x7fbbc6, 0x203652, 0xc23c33,
]
const NAMES: Array[String] = ["Stone", "Dark stone", "Sand", "Brick", "Terracotta", "Slate",
	"Wood", "Grass", "Foliage", "Water", "White", "Charcoal", "Gold", "Glass", "Navy", "Red"]

## The whole palette, in index order.
static func colors() -> PackedColorArray:
	var result := PackedColorArray()
	for hex in HEX:
		result.append(color_of(hex))
	return result

## The color of one hex entry.
static func color_of(hex: int) -> Color:
	return Color8((hex >> 16) & 0xff, (hex >> 8) & 0xff, hex & 0xff)

## The name of a palette index, for menus and tooltips.
static func name_of(index: int) -> String:
	return NAMES[index] if index >= 0 and index < NAMES.size() else "Auto"
