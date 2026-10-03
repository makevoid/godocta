class_name OctaFile
extends RefCounted
## OCTA: little-endian, versioned, bounded binary coordinate format.
## v1 stores coordinates; v2 adds an RGB palette (≤ 16) and one palette index per cell.
## Uncolored structures are written as v1. It never deserializes objects, scripts or
## executable resources.

const VERSION: int = 2
const HEADER_BYTES: int = 20
const HEADER_V2_BYTES: int = 24
const MAX_TITLE_BYTES: int = 1024
var last_error: String = ""

func save_structure(path: String, structure: OctaStructure) -> Error:
	last_error = ""
	if structure == null:
		return _fail("No structure to save.", ERR_INVALID_DATA)
	var cells := structure.packed_cells()
	var invalid := OctaStructure.validation_error(cells, structure.cell_span)
	if not invalid.is_empty():
		return _fail(invalid, ERR_INVALID_DATA)
	var title_bytes := structure.title.to_utf8_buffer()
	if title_bytes.size() > MAX_TITLE_BYTES:
		return _fail("Structure title is too long.", ERR_INVALID_DATA)
	var colors := structure.packed_colors()
	var colored := not colors.is_empty()
	var header := HEADER_V2_BYTES if colored else HEADER_BYTES
	var palette := structure.palette if colored else PackedColorArray()
	var bytes := PackedByteArray()
	bytes.resize(header + title_bytes.size() + palette.size() * 3 + structure.storage_bytes())
	bytes[0] = 79 # O
	bytes[1] = 67 # C
	bytes[2] = 84 # T
	bytes[3] = 65 # A
	bytes.encode_u32(4, 2 if colored else 1)
	bytes.encode_float(8, structure.cell_span)
	bytes.encode_u32(12, structure.cell_count())
	bytes.encode_u32(16, title_bytes.size())
	if colored:
		bytes.encode_u32(20, palette.size())
	for index in title_bytes.size():
		bytes[header + index] = title_bytes[index]
	var offset := header + title_bytes.size()
	for color in palette:
		bytes[offset] = color.r8
		bytes[offset + 1] = color.g8
		bytes[offset + 2] = color.b8
		offset += 3
	for index in cells.size():
		bytes.encode_s32(offset + index * 4, cells[index])
	offset += cells.size() * 4
	for index in colors.size():
		bytes[offset + index] = colors[index]
	# Write beside the destination, then rename, so a failed write preserves it.
	var temporary := path + ".tmp"
	var file := FileAccess.open(temporary, FileAccess.WRITE)
	if file == null:
		return _fail("Cannot open the destination.", FileAccess.get_open_error())
	file.store_buffer(bytes)
	file.flush()
	var result := file.get_error()
	file.close()
	if result != OK:
		DirAccess.remove_absolute(temporary)
		return _fail("Could not write the complete file.", result)
	result = DirAccess.rename_absolute(temporary, path)
	if result != OK:
		DirAccess.remove_absolute(temporary)
		return _fail("Could not replace the destination file.", result)
	return OK

func load_structure(path: String) -> OctaStructure:
	last_error = ""
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		last_error = "Cannot open this file."
		return null
	var length := file.get_length()
	if length < HEADER_BYTES or length > HEADER_V2_BYTES + MAX_TITLE_BYTES \
		+ OctaStructure.MAX_PALETTE * 3 + OctaStructure.MAX_CELLS * 13:
		last_error = "Invalid OCTA file length."
		return null
	var bytes := file.get_buffer(length)
	file.close()
	if bytes.size() != length or bytes.slice(0, 4).get_string_from_ascii() != "OCTA":
		last_error = "This is not an OCTA structure."
		return null
	var version := bytes.decode_u32(4)
	var header := HEADER_BYTES
	var palette_size := 0
	if version == 2 and length >= HEADER_V2_BYTES:
		header = HEADER_V2_BYTES
		palette_size = bytes.decode_u32(20)
	elif version != 1:
		last_error = "Unsupported OCTA file version."
		return null
	var span := bytes.decode_float(8)
	var count := bytes.decode_u32(12)
	var title_size := bytes.decode_u32(16)
	var color_bytes := count if palette_size > 0 else 0
	if count > OctaStructure.MAX_CELLS or title_size > MAX_TITLE_BYTES \
		or palette_size > OctaStructure.MAX_PALETTE \
		or header + title_size + palette_size * 3 + count * 12 + color_bytes != length:
		last_error = "Invalid or incomplete coordinate payload."
		return null
	var title := bytes.slice(header, header + title_size).get_string_from_utf8()
	var offset := header + title_size
	var palette := PackedColorArray()
	for index in palette_size:
		palette.append(Color8(bytes[offset], bytes[offset + 1], bytes[offset + 2]))
		offset += 3
	var cells := PackedInt32Array()
	cells.resize(count * 3)
	for index in cells.size():
		cells[index] = bytes.decode_s32(offset + index * 4)
	offset += cells.size() * 4
	last_error = OctaStructure.validation_error(cells, span)
	if not last_error.is_empty():
		return null
	var structure := OctaStructure.from_cells(cells, title, span)
	if color_bytes > 0:
		last_error = structure.set_colors(bytes.slice(offset, offset + color_bytes), palette)
		if not last_error.is_empty():
			return null
	return structure

func _fail(message: String, error: Error) -> Error:
	last_error = message
	return error
