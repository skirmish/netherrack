package items

import (
	"os"
)

var fontWidths []byte

func init() {
	f, err := os.Open("data/glyph_sizes.bin")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	def := []byte{
		0, 8, 8, 7, 7, 7, 7, 6, 8, 7, 8, 8, 7, 8, 8, 8,
		7, 7, 7, 7, 8, 8, 7, 8, 7, 7, 7, 7, 7, 8, 8, 8,
		3, 1, 4, 5, 5, 5, 5, 2, 4, 4, 4, 5, 1, 5, 1, 5,
		5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 1, 1, 4, 5, 4, 5,
		6, 5, 5, 5, 5, 5, 5, 5, 5, 3, 5, 5, 5, 5, 5, 5,
		5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 3, 5, 3, 5, 5,
		2, 5, 5, 5, 5, 5, 4, 5, 5, 1, 5, 4, 2, 5, 5, 5,
		5, 5, 5, 5, 3, 5, 5, 5, 5, 5, 5, 4, 1, 4, 6, 5,
		5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 3, 5, 2, 5, 5,
		5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 3, 5,
		5, 2, 5, 5, 5, 5, 5, 5, 5, 6, 5, 5, 5, 1, 5, 5,
		7, 8, 8, 5, 5, 5, 7, 7, 5, 7, 7, 7, 7, 7, 5, 5,
		8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
		8, 8, 8, 8, 8, 8, 8, 8, 8, 5, 8, 8, 8, 4, 8, 8,
		7, 6, 6, 7, 6, 7, 7, 7, 6, 7, 7, 6, 8, 8, 5, 6,
		6, 6, 6, 6, 8, 5, 6, 7, 6, 5, 5, 8, 6, 5, 6, 0,
	}
	fontWidths := make([]byte, 0xFFFF)
	f.Read(fontWidths)
	copy(fontWidths, def)
	for i := 256; i < len(fontWidths); i++ {
		fontWidths[i] = (fontWidths[i]&0x0F - fontWidths[i]>>4)
	}
}

//\u25CF = Large Dot
