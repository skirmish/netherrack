package metadata

import (
	"encoding/binary"
	"github.com/thinkofdeath/netherrack/items"
	"github.com/thinkofdeath/soulsand"
	"io"
	"math"
)

var _ soulsand.EntityMetadata = Type{}

type Type map[int]interface{}

func (t Type) Set(index int, data interface{}) {
	t[index] = data
}

func (t Type) Get(index int) interface{} {
	return t[index]
}

func (t Type) WriteTo(w io.Writer) {
	for index, value := range t {
		var key byte = byte(index) & 0x1F
		switch value := value.(type) {
		case int8:
			key |= (0 << 5) & 0xE0
			w.Write([]byte{key, byte(value)})
		case int16:
			key |= (1 << 5) & 0xE0
			data := make([]byte, 3)
			data[0] = key
			binary.BigEndian.PutUint16(data[1:], uint16(value))
			w.Write(data)
		case int32:
			key |= (2 << 5) & 0xE0
			data := make([]byte, 5)
			data[0] = key
			binary.BigEndian.PutUint32(data[1:], uint32(value))
			w.Write(data)
		case float32:
			key |= (3 << 5) & 0xE0
			data := make([]byte, 5)
			data[0] = key
			binary.BigEndian.PutUint32(data[1:], math.Float32bits(value))
			w.Write(data)
		case string:
			key |= (4 << 5) & 0xE0
			runes := []rune(value)
			data := make([]byte, 1+2+len(runes)*2)
			data[0] = key
			binary.BigEndian.PutUint16(data[1:], uint16(len(runes)))
			for i, r := range runes {
				binary.BigEndian.PutUint16(data[3+i*2:], uint16(r))
			}
			w.Write(data)
		case items.ItemStack:
			key |= (5 << 5) & 0xE0
			panic("Not supported")
			w.Write([]byte{key})
		case []int32:
			key |= (6 << 5) & 0xE0
			data := make([]byte, 1+4+4+4)
			data[0] = key
			binary.BigEndian.PutUint32(data[1:], uint32(value[0]))
			binary.BigEndian.PutUint32(data[5:], uint32(value[1]))
			binary.BigEndian.PutUint32(data[9:], uint32(value[2]))
			w.Write(data)
		}
	}
	w.Write([]byte{127})
}
