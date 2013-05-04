package nbt

import (
	"encoding/binary"
	"io"
	"math"
)

func Parse(reader io.Reader) Type {
	out := make(Type)
	r := Reader{R: reader}
	typeID := r.ReadByte()
	if typeID != 10 {
		panic("Not an NBT file")
	}
	r.ReadString()
	parseCompound(out, r)
	return out
}

func parseCompound(out Type, r Reader) {
compoundLoop:
	for {
		switch r.ReadByte() {
		case 0:
			break compoundLoop
		case 1:
			out[r.ReadString()] = r.ReadByte()
		case 2:
			out[r.ReadString()] = r.ReadShort()
		case 3:
			out[r.ReadString()] = r.ReadInt()
		case 4:
			out[r.ReadString()] = r.ReadLong()
		case 5:
			out[r.ReadString()] = r.ReadFloat()
		case 6:
			out[r.ReadString()] = r.ReadDouble()
		case 7:
			name := r.ReadString()
			data := make([]int8, r.ReadInt())
			for i, _ := range data {
				data[i] = r.ReadByte()
			}
			out[name] = data
		case 8:
			out[r.ReadString()] = r.ReadString()
		case 9:
			name := r.ReadString()
			out[name] = parseList(r)
		case 10:
			name := r.ReadString()
			com := make(Type)
			parseCompound(com, r)
			out[name] = com
		case 11:
			name := r.ReadString()
			data := make([]int32, r.ReadInt())
			for i, _ := range data {
				data[i] = r.ReadInt()
			}
			out[name] = data

		}
	}
}

func parseList(r Reader) []interface{} {
	typeID := r.ReadByte()
	list := make([]interface{}, r.ReadInt())
	switch typeID {
	case 1:
		for i, _ := range list {
			list[i] = r.ReadByte()
		}
	case 2:
		for i, _ := range list {
			list[i] = r.ReadShort()
		}
	case 3:
		for i, _ := range list {
			list[i] = r.ReadInt()
		}
	case 4:
		for i, _ := range list {
			list[i] = r.ReadLong()
		}
	case 5:
		for i, _ := range list {
			list[i] = r.ReadFloat()
		}
	case 6:
		for i, _ := range list {
			list[i] = r.ReadDouble()
		}
	case 7:
		for i, _ := range list {
			data := make([]int8, r.ReadInt())
			for j, _ := range data {
				data[j] = r.ReadByte()
			}
			list[i] = data
		}
	case 8:
		for i, _ := range list {
			list[i] = r.ReadString()
		}
	case 9:
		for i, _ := range list {
			list[i] = parseList(r)
		}
	case 10:
		for i, _ := range list {
			com := make(Type)
			parseCompound(com, r)
			list[i] = com
		}
	case 11:
		for i, _ := range list {
			data := make([]int32, r.ReadInt())
			for j, _ := range data {
				data[j] = r.ReadInt()
			}
			list[i] = data
		}
	}
	return list
}

type Reader struct {
	R io.Reader
}

func (r Reader) ReadUByte() byte {
	b := make([]byte, 1)
	r.R.Read(b)
	return b[0]
}

func (r Reader) ReadByte() int8 {
	return int8(r.ReadUByte())
}

func (r Reader) ReadUShort() uint16 {
	b := make([]byte, 2)
	r.R.Read(b)
	return binary.BigEndian.Uint16(b)
}

func (r Reader) ReadShort() int16 {
	return int16(r.ReadUShort())
}

func (r Reader) ReadInt() int32 {
	b := make([]byte, 4)
	r.R.Read(b)
	return int32(binary.BigEndian.Uint32(b))
}

func (r Reader) ReadLong() int64 {
	b := make([]byte, 8)
	r.R.Read(b)
	return int64(binary.BigEndian.Uint64(b))
}

func (r Reader) ReadFloat() float32 {
	b := make([]byte, 4)
	r.R.Read(b)
	return math.Float32frombits(binary.BigEndian.Uint32(b))
}

func (r Reader) ReadDouble() float64 {
	b := make([]byte, 8)
	r.R.Read(b)
	return math.Float64frombits(binary.BigEndian.Uint64(b))
}

func (r Reader) ReadString() string {
	l := int(r.ReadUShort())
	d := make([]byte, l)
	r.R.Read(d)
	return string(d)
}