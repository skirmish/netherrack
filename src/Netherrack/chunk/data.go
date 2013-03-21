package chunk

import (
	"encoding/binary"
	"math"
)

func writeBool(out []byte, b bool) {
	if b {
		out[0] = 1
	} else {
		out[0] = 0
	}
}

func writeByte(out []byte, b int8) {
	out[0] = byte(b)
}

func writeShort(out []byte, s int16) {
	binary.BigEndian.PutUint16(out, uint16(s))
}

func writeUShort(out []byte, s uint16) {
	binary.BigEndian.PutUint16(out, s)
}

func writeInt(out []byte, s int32) {
	binary.BigEndian.PutUint32(out, uint32(s))
}

func writeFloat(out []byte, s float32) {
	binary.BigEndian.PutUint32(out, math.Float32bits(s))
}

func writeDouble(out []byte, s float64) {
	binary.BigEndian.PutUint64(out, math.Float64bits(s))
}

func writeString(out []byte, r []rune) int {
	//r := []rune(str)
	writeShort(out[0:2], int16(len(r)))
	pos := 2
	for _, b := range r {
		binary.BigEndian.PutUint16(out[pos:pos+2], uint16(b))
		pos += 2
	}
	return 2 + len(r)*2
}
