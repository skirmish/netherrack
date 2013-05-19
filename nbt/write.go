package nbt

import (
	"encoding/binary"
	"io"
	"math"
)

func (t Type) WriteTo(w io.Writer, name string) {
	out := make([]byte, 1+2)
	out[0] = 10
	binary.BigEndian.PutUint16(out[1:], uint16(len(name)))
	w.Write(out)
	w.Write([]byte(name))
	t.writeTo(w)

}

func (t Type) writeTo(w io.Writer) {
	for k, v := range t {
		switch v := v.(type) {
		case int8:
			out := make([]byte, 1+2)
			out[0] = 1
			binary.BigEndian.PutUint16(out[1:], uint16(len(k)))
			w.Write(out)
			w.Write([]byte(k))
			w.Write([]byte{byte(v)})
		case int16:
			out := make([]byte, 1+2)
			out[0] = 2
			binary.BigEndian.PutUint16(out[1:], uint16(len(k)))
			w.Write(out)
			w.Write([]byte(k))
			out = make([]byte, 2)
			binary.BigEndian.PutUint16(out, uint16(v))
			w.Write(out)
		case int32:
			out := make([]byte, 1+2)
			out[0] = 3
			binary.BigEndian.PutUint16(out[1:], uint16(len(k)))
			w.Write(out)
			w.Write([]byte(k))
			out = make([]byte, 4)
			binary.BigEndian.PutUint32(out, uint32(v))
			w.Write(out)
		case int64:
			out := make([]byte, 1+2)
			out[0] = 4
			binary.BigEndian.PutUint16(out[1:], uint16(len(k)))
			w.Write(out)
			w.Write([]byte(k))
			out = make([]byte, 8)
			binary.BigEndian.PutUint64(out, uint64(v))
			w.Write(out)
		case float32:
			out := make([]byte, 1+2)
			out[0] = 5
			binary.BigEndian.PutUint16(out[1:], uint16(len(k)))
			w.Write(out)
			w.Write([]byte(k))
			out = make([]byte, 4)
			binary.BigEndian.PutUint32(out, math.Float32bits(v))
			w.Write(out)
		case float64:
			out := make([]byte, 1+2)
			out[0] = 6
			binary.BigEndian.PutUint16(out[1:], uint16(len(k)))
			w.Write(out)
			w.Write([]byte(k))
			out = make([]byte, 8)
			binary.BigEndian.PutUint64(out, math.Float64bits(v))
			w.Write(out)
		case []int8:
			out := make([]byte, 1+2)
			out[0] = 7
			binary.BigEndian.PutUint16(out[1:], uint16(len(k)))
			w.Write(out)
			w.Write([]byte(k))
			out = make([]byte, 4+len(v))
			binary.BigEndian.PutUint32(out, uint32(len(v)))
			for i := 0; i < len(v); i++ {
				out[4+i] = byte(v[i])
			}
			w.Write(out)
		case string:
			out := make([]byte, 1+2)
			out[0] = 8
			binary.BigEndian.PutUint16(out[1:], uint16(len(k)))
			w.Write(out)
			w.Write([]byte(k))
			out = make([]byte, 2)
			binary.BigEndian.PutUint16(out, uint16(len(v)))
			w.Write(out)
			w.Write([]byte(v))
		case []interface{}:
			out := make([]byte, 1+2)
			out[0] = 9
			binary.BigEndian.PutUint16(out[1:], uint16(len(k)))
			w.Write(out)
			w.Write([]byte(k))
			writeList(w, v)
		case Type:
			out := make([]byte, 1+2)
			out[0] = 10
			binary.BigEndian.PutUint16(out[1:], uint16(len(k)))
			w.Write(out)
			w.Write([]byte(k))
			v.writeTo(w)
		case []int32:
			out := make([]byte, 1+2)
			out[0] = 11
			binary.BigEndian.PutUint16(out[1:], uint16(len(k)))
			w.Write(out)
			w.Write([]byte(k))
			out = make([]byte, 4+4*len(v))
			binary.BigEndian.PutUint32(out, uint32(len(v)))
			for i, b := range v {
				binary.BigEndian.PutUint32(out[4+i*4:], uint32(b))
			}
			w.Write(out)
		}
	}
	w.Write([]byte{0})
}

func writeList(w io.Writer, list []interface{}) {
	var val interface{}
	if len(list) > 0 {
		val = list[0]
	} else {
		val = []int8{0}
	}
	switch val.(type) {
	case int8:
		w.Write([]byte{1})
		out := make([]byte, 4+len(list))
		binary.BigEndian.PutUint32(out, uint32(len(list)))
		for i, b := range list {
			out[4+i] = byte(b.(int8))
		}
		w.Write(out)
	case int16:
		w.Write([]byte{2})
		out := make([]byte, 4+len(list)*2)
		binary.BigEndian.PutUint32(out, uint32(len(list)))
		for i, s := range list {
			binary.BigEndian.PutUint16(out[4+i*2:], uint16(s.(int16)))
		}
		w.Write(out)
	case int32:
		w.Write([]byte{3})
		out := make([]byte, 4+len(list)*4)
		binary.BigEndian.PutUint32(out, uint32(len(list)))
		for i, in := range list {
			binary.BigEndian.PutUint32(out[4+i*4:], uint32(in.(int32)))
		}
		w.Write(out)
	case int64:
		w.Write([]byte{4})
		out := make([]byte, 4+len(list)*8)
		binary.BigEndian.PutUint32(out, uint32(len(list)))
		for i, l := range list {
			binary.BigEndian.PutUint64(out[4+i*8:], uint64(l.(int64)))
		}
		w.Write(out)
	case float32:
		w.Write([]byte{5})
		out := make([]byte, 4+len(list)*4)
		binary.BigEndian.PutUint32(out, uint32(len(list)))
		for i, f := range list {
			binary.BigEndian.PutUint32(out[4+i*4:], math.Float32bits(f.(float32)))
		}
		w.Write(out)
	case float64:
		w.Write([]byte{6})
		out := make([]byte, 4+len(list)*8)
		binary.BigEndian.PutUint32(out, uint32(len(list)))
		for i, f := range list {
			binary.BigEndian.PutUint64(out[4+i*8:], math.Float64bits(f.(float64)))
		}
		w.Write(out)
	case []byte:
		w.Write([]byte{7})
		out := make([]byte, 4)
		binary.BigEndian.PutUint32(out, uint32(len(list)))
		w.Write(out)
		for _, b := range list {
			by := b.([]byte)
			binary.BigEndian.PutUint32(out, uint32(len(by)))
			w.Write(out)
			w.Write(by)
		}
	case string:
		w.Write([]byte{8})
		out := make([]byte, 4)
		binary.BigEndian.PutUint32(out, uint32(len(list)))
		w.Write(out)
		out = make([]byte, 2)
		for _, s := range list {
			str := s.(string)
			binary.BigEndian.PutUint16(out, uint16(len(str)))
			w.Write(out)
			w.Write([]byte(str))
		}
	case []interface{}:
		w.Write([]byte{9})
		out := make([]byte, 4)
		binary.BigEndian.PutUint32(out, uint32(len(list)))
		w.Write(out)
		for _, v := range list {
			writeList(w, v.([]interface{}))
		}
	case Type:
		w.Write([]byte{10})
		out := make([]byte, 4)
		binary.BigEndian.PutUint32(out, uint32(len(list)))
		w.Write(out)
		for _, v := range list {
			t := v.(Type)
			t.writeTo(w)
		}
	case []int32:
		w.Write([]byte{11})
		out := make([]byte, 4)
		binary.BigEndian.PutUint32(out, uint32(len(list)))
		w.Write(out)
		for _, i := range list {
			ints := i.([]int)
			out = make([]byte, 4+len(ints)*4)
			binary.BigEndian.PutUint32(out, uint32(len(ints)))
			for i, in := range ints {
				binary.BigEndian.PutUint32(out[4+i*4:], uint32(in))
			}
			w.Write(out)
		}
	}
}
