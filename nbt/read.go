package nbt

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
)

func Parse(reader io.Reader) (t Type, e error) {
	defer func() {
		if err := recover(); err != nil {
			e = err.(error)
		}
	}()
	out := make(Type)
	r := Reader{R: reader}
	typeID := r.ReadByte()
	if typeID != 10 {
		return nil, errors.New("Not an NBT file")
	}
	r.ReadString()
	parseCompound(out, r)
	t = out
	return
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
			size := int(r.ReadInt())
			data := make([]int8, size)
			dataBytes := make([]byte, size)
			_, err := io.ReadFull(r.R, dataBytes)
			if err != nil {
				panic(err)
			}
			for i := 0; i < size; i++ {
				data[i] = int8(dataBytes[i])
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
			size := int(r.ReadInt())
			data := make([]int32, size)
			dataBytes := make([]byte, size*4)
			_, err := io.ReadFull(r.R, dataBytes)
			if err != nil {
				panic(err)
			}
			for i := 0; i < size; i++ {
				data[i] = int32(binary.BigEndian.Uint32(dataBytes[i*4:]))
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
			size := int(r.ReadInt())
			data := make([]int8, size)
			dataBytes := make([]byte, size)
			_, err := io.ReadFull(r.R, dataBytes)
			if err != nil {
				panic(err)
			}
			for i := 0; i < size; i++ {
				data[i] = int8(dataBytes[i])
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
			size := int(r.ReadInt())
			data := make([]int32, size)
			dataBytes := make([]byte, size*4)
			_, err := io.ReadFull(r.R, dataBytes)
			if err != nil {
				panic(err)
			}
			for i := 0; i < size; i++ {
				data[i] = int32(binary.BigEndian.Uint32(dataBytes[i*4:]))
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
	_, err := io.ReadFull(r.R, b)
	if err != nil {
		panic(err)
	}
	return b[0]
}

func (r Reader) ReadByte() int8 {
	return int8(r.ReadUByte())
}

func (r Reader) ReadUShort() uint16 {
	b := make([]byte, 2)
	_, err := io.ReadFull(r.R, b)
	if err != nil {
		panic(err)
	}
	return binary.BigEndian.Uint16(b)
}

func (r Reader) ReadShort() int16 {
	return int16(r.ReadUShort())
}

func (r Reader) ReadInt() int32 {
	b := make([]byte, 4)
	_, err := io.ReadFull(r.R, b)
	if err != nil {
		panic(err)
	}
	return int32(binary.BigEndian.Uint32(b))
}

func (r Reader) ReadLong() int64 {
	b := make([]byte, 8)
	_, err := io.ReadFull(r.R, b)
	if err != nil {
		panic(err)
	}
	return int64(binary.BigEndian.Uint64(b))
}

func (r Reader) ReadFloat() float32 {
	b := make([]byte, 4)
	_, err := io.ReadFull(r.R, b)
	if err != nil {
		panic(err)
	}
	return math.Float32frombits(binary.BigEndian.Uint32(b))
}

func (r Reader) ReadDouble() float64 {
	b := make([]byte, 8)
	_, err := io.ReadFull(r.R, b)
	if err != nil {
		panic(err)
	}
	return math.Float64frombits(binary.BigEndian.Uint64(b))
}

func (r Reader) ReadString() string {
	l := int(r.ReadUShort())
	d := make([]byte, l)
	_, err := io.ReadFull(r.R, d)
	if err != nil {
		panic(err)
	}
	return string(d)
}
