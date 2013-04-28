package protocol

import (
	"encoding/binary"
	"math"
)

//byteWriter
type byteWriter struct {
	data   []byte
	offset int
}

func NewByteWriter(size int) *byteWriter {
	return &byteWriter{
		data: make([]byte, size),
	}
}

func (bw *byteWriter) Bytes() []byte {
	return bw.data
}

func (bw *byteWriter) WriteUByte(b byte) {
	bw.data[bw.offset] = b
	bw.offset++
}

func (bw *byteWriter) WriteBool(b bool) {
	if b {
		bw.data[bw.offset] = 1
	}
	bw.offset++
}

func (bw *byteWriter) WriteByte(b int8) {
	bw.WriteUByte(byte(b))
}

func (bw *byteWriter) WriteUShort(short uint16) {
	binary.BigEndian.PutUint16(bw.data[bw.offset:], short)
	bw.offset += 2
}

func (bw *byteWriter) WriteShort(short int16) {
	bw.WriteUShort(uint16(short))
}

func (bw *byteWriter) WriteInt(i int32) {
	binary.BigEndian.PutUint32(bw.data[bw.offset:], uint32(i))
	bw.offset += 4
}

func (bw *byteWriter) WriteLong(long int64) {
	binary.BigEndian.PutUint64(bw.data[bw.offset:], uint64(long))
	bw.offset += 8
}

func (bw *byteWriter) WriteFloat(float float32) {
	binary.BigEndian.PutUint32(bw.data[bw.offset:], math.Float32bits(float))
	bw.offset += 4
}

func (bw *byteWriter) WriteDouble(double float64) {
	binary.BigEndian.PutUint64(bw.data[bw.offset:], math.Float64bits(double))
	bw.offset += 8
}

func (bw *byteWriter) WriteString(str []rune) {
	bw.WriteShort(int16(len(str)))
	for _, r := range str {
		bw.WriteUShort(uint16(r))
	}
}
