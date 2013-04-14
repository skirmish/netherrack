package items

import (
	"os"
)

var fontWidths []byte

func init() {
	f, err := os.Open("data/font.bin")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	fontWidths := make([]byte, 0xFFFF)
	f.Read(fontWidths)
}

//\u25CF = Large Dot
