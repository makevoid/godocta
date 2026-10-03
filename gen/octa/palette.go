package octa

// Color is an index into the 16-color base palette, stored as one byte per cell.
type Color uint8

// The base palette. Index 0 (Stone) is the default brush.
const (
	Stone Color = iota
	DarkStone
	Sand
	Brick
	Terracotta
	Slate
	Wood
	Grass
	Foliage
	Water
	White
	Charcoal
	Gold
	Glass
	Navy
	Red
	PaletteSize
)

// RGB is an sRGB color.
type RGB struct{ R, G, B uint8 }

// Palette holds the sRGB value of every Color. Files embed it, so viewers need no table.
var Palette = [PaletteSize]RGB{
	Stone:      {0xb9, 0xb3, 0xa4},
	DarkStone:  {0x7b, 0x76, 0x6c},
	Sand:       {0xdb, 0xc5, 0x93},
	Brick:      {0xa9, 0x55, 0x40},
	Terracotta: {0xcc, 0x7a, 0x4a},
	Slate:      {0x4f, 0x63, 0x77},
	Wood:       {0x7c, 0x56, 0x35},
	Grass:      {0x6f, 0x9b, 0x4b},
	Foliage:    {0x3e, 0x6b, 0x3a},
	Water:      {0x3c, 0x7d, 0xa8},
	White:      {0xee, 0xee, 0xe6},
	Charcoal:   {0x35, 0x39, 0x3e},
	Gold:       {0xd9, 0xaa, 0x3c},
	Glass:      {0x7f, 0xbb, 0xc6},
	Navy:       {0x20, 0x36, 0x52},
	Red:        {0xc2, 0x3c, 0x33},
}

var colorNames = [PaletteSize]string{"stone", "dark stone", "sand", "brick", "terracotta", "slate", "wood",
	"grass", "foliage", "water", "white", "charcoal", "gold", "glass", "navy", "red"}

func (c Color) String() string {
	if c < PaletteSize {
		return colorNames[c]
	}
	return "invalid"
}

// Or returns c, or fallback when c is zero (Stone). Components use it so that a zero
// color field means "this component's default".
func (c Color) Or(fallback Color) Color {
	if c == 0 {
		return fallback
	}
	return c
}
