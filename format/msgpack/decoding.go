/*
   Copyright 2013 Matthew Collins (purggames@gmail.com)

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package msgpack

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

var decodeGenerators map[reflect.Kind]func(t reflect.Type) interface{}

func init() {
	decodeGenerators = map[reflect.Kind]func(t reflect.Type) interface{}{
		reflect.Struct:    decodeStruct,
		reflect.String:    decodeString,
		reflect.Bool:      decodeBool,
		reflect.Ptr:       decodePtr,
		reflect.Int:       decodeInt,
		reflect.Int8:      decodeInt8,
		reflect.Int16:     decodeInt16,
		reflect.Int32:     decodeInt32,
		reflect.Int64:     decodeInt64,
		reflect.Uint:      decodeUint,
		reflect.Uint8:     decodeUint8,
		reflect.Uint16:    decodeUint16,
		reflect.Uint32:    decodeUint32,
		reflect.Uint64:    decodeUint64,
		reflect.Float32:   decodeFloat32,
		reflect.Float64:   decodeFloat64,
		reflect.Slice:     decodeSlice,
		reflect.Array:     decodeArray,
		reflect.Map:       decodeMap,
		reflect.Interface: decodeInterface,
	}
}

func getTypeDecoder(t reflect.Type) interface{} {
	gen, ok := decodeGenerators[t.Kind()]
	if !ok {
		panic(fmt.Sprintf("Unhandled kind %s", t.Kind()))
	}
	return gen(t)
}

type dField struct {
	offset       uintptr
	decode       interface{}
	index        []int
	needsReflect bool
}

func decodeStruct(t reflect.Type) interface{} {
	fields := map[string]dField{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if len(f.Name) > 0 {
			r, _ := utf8.DecodeRuneInString(f.Name)
			if unicode.IsLower(r) {
				continue
			}
		}
		if f.Tag.Get("msgpack") == "ignore" {
			continue
		}
		if f.Anonymous {
			f.Name = f.Type.Name()
		}
		d := dField{
			offset: f.Offset,
			decode: getTypeDecoder(f.Type),
			index:  f.Index,
		}
		_, d.needsReflect = d.decode.(decodeReflectFunc)
		fields[f.Name] = d
	}
	return decodeReflectFunc(func(dec *Decoder, v reflect.Value) error {
		p := v.UnsafeAddr()
		b, err := dec.r.ReadByte()
		if err != nil {
			return err
		}
		var l int
		switch {
		case b&0xF0 == 0x80:
			l = int(b & 0xF)
		case b == 0xDE:
			bs := dec.b[:2]
			_, err := dec.fr.Read(bs)
			if err != nil {
				return err
			}
			l = int(binary.BigEndian.Uint16(bs))
		case b == 0xDF:
			bs := dec.b[:4]
			_, err := dec.fr.Read(bs)
			if err != nil {
				return err
			}
			l = int(binary.BigEndian.Uint32(bs))
		default:
			return ErrorIncorrectType{"map", b}
		}
		for i := 0; i < l; i++ {
			name := new(string)
			_decodeString(dec, unsafe.Pointer(name))
			d, ok := fields[*name]
			if !ok {
				panic("Skipping NYI " + *name)
			}
			if !d.needsReflect {
				err := (d.decode.(decodeFunc))(dec, unsafe.Pointer(p+d.offset))
				if err != nil {
					return err
				}
			} else {
				err := (d.decode.(decodeReflectFunc))(dec, v.FieldByIndex(d.index))
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func decodeString(t reflect.Type) interface{} {
	return decodeFunc(_decodeString)
}

func _decodeString(dec *Decoder, p unsafe.Pointer) error {
	b, err := dec.r.ReadByte()
	if err != nil {
		return err
	}
	var l int
	switch {
	case b&0xE0 == 0xA0:
		l = int(b & 0x1F)
	case b == 0xD9:
		c, err := dec.r.ReadByte()
		if err != nil {
			return err
		}
		l = int(c)
	case b == 0xDA:
		bs := dec.b[:2]
		_, err := dec.fr.Read(bs)
		if err != nil {
			return err
		}
		l = int(binary.BigEndian.Uint16(bs))
	case b == 0xDB:
		bs := dec.b[:4]
		_, err := dec.fr.Read(bs)
		if err != nil {
			return err
		}
		l = int(binary.BigEndian.Uint32(bs))
	default:
		return ErrorIncorrectType{"string", b}
	}
	by := make([]byte, l)
	_, err = dec.fr.Read(by)
	if err != nil {
		return err
	}
	*(*string)(p) = string(by)
	return nil
}

func decodeBool(t reflect.Type) interface{} {
	return decodeFunc(_decodeBool)
}

func _decodeBool(dec *Decoder, p unsafe.Pointer) error {
	b, err := dec.r.ReadByte()
	if err != nil {
		return err
	}
	switch b {
	case 0xC3:
		*(*bool)(p) = true
	case 0xC2:
		*(*bool)(p) = false
	default:
		return ErrorIncorrectType{"bool", b}
	}
	return nil
}

func decodePtr(t reflect.Type) interface{} {
	elem := t.Elem()
	elemDec := getTypeDecoder(elem)
	_, ok := elemDec.(decodeFunc)
	return decodeReflectFunc(func(dec *Decoder, v reflect.Value) error {
		p := unsafe.Pointer(v.UnsafeAddr())
		b, err := dec.r.ReadByte()
		if err != nil {
			return err
		}
		if b == 0xC0 {
			*(*uintptr)(p) = 0
			return nil
		}
		dec.r.UnreadByte()
		val := reflect.New(elem)
		if ok {
			err = (elemDec.(decodeFunc))(dec, unsafe.Pointer(val.Elem().UnsafeAddr()))
			if err != nil {
				return err
			}
		} else {
			err = (elemDec.(decodeReflectFunc))(dec, val.Elem())
			if err != nil {
				return err
			}
		}
		*(*uintptr)(p) = val.Pointer()
		return nil
	})
}

func readInt(dec *Decoder) (int64, error) {
	b, err := dec.r.ReadByte()
	if err != nil {
		return 0, err
	}
	switch b {
	case 0xD0:
		b, err := dec.r.ReadByte()
		return int64(int8(b)), err
	case 0xD1:
		bs := dec.b[:2]
		_, err := dec.fr.Read(bs)
		return int64(int16(binary.BigEndian.Uint16(bs))), err
	case 0xD2:
		bs := dec.b[:4]
		_, err := dec.fr.Read(bs)
		return int64(int32(binary.BigEndian.Uint32(bs))), err
	case 0xD3:
		bs := dec.b[:8]
		_, err := dec.fr.Read(bs)
		return int64(binary.BigEndian.Uint64(bs)), err
	default:
		return 0, ErrorIncorrectType{"int", b}
	}
}

func decodeInt(t reflect.Type) interface{} {
	return decodeFunc(_decodeInt)
}
func _decodeInt(dec *Decoder, p unsafe.Pointer) error {
	v, err := readInt(dec)
	*(*int)(p) = int(v)
	return err
}
func decodeInt8(t reflect.Type) interface{} {
	return decodeFunc(_decodeInt8)
}
func _decodeInt8(dec *Decoder, p unsafe.Pointer) error {
	v, err := readInt(dec)
	*(*int8)(p) = int8(v)
	return err
}
func decodeInt16(t reflect.Type) interface{} {
	return decodeFunc(_decodeInt16)
}
func _decodeInt16(dec *Decoder, p unsafe.Pointer) error {
	v, err := readInt(dec)
	*(*int16)(p) = int16(v)
	return err
}
func decodeInt32(t reflect.Type) interface{} {
	return decodeFunc(_decodeInt32)
}
func _decodeInt32(dec *Decoder, p unsafe.Pointer) error {
	v, err := readInt(dec)
	*(*int32)(p) = int32(v)
	return err
}
func decodeInt64(t reflect.Type) interface{} {
	return decodeFunc(_decodeInt64)
}
func _decodeInt64(dec *Decoder, p unsafe.Pointer) error {
	v, err := readInt(dec)
	*(*int64)(p) = v
	return err
}

func readUint(dec *Decoder) (uint64, error) {
	b, err := dec.r.ReadByte()
	if err != nil {
		return 0, err
	}
	switch b {
	case 0xCC:
		b, err := dec.r.ReadByte()
		return uint64(b), err
	case 0xCD:
		bs := dec.b[:2]
		_, err := dec.fr.Read(bs)
		return uint64(binary.BigEndian.Uint16(bs)), err
	case 0xCE:
		bs := dec.b[:4]
		_, err := dec.fr.Read(bs)
		return uint64(binary.BigEndian.Uint32(bs)), err
	case 0xCF:
		bs := dec.b[:8]
		_, err := dec.fr.Read(bs)
		return uint64(binary.BigEndian.Uint64(bs)), err
	default:
		return 0, ErrorIncorrectType{"uint", b}
	}
}

func decodeUint(t reflect.Type) interface{} {
	return decodeFunc(_decodeUint)
}
func _decodeUint(dec *Decoder, p unsafe.Pointer) error {
	v, err := readUint(dec)
	*(*uint)(p) = uint(v)
	return err
}
func decodeUint8(t reflect.Type) interface{} {
	return decodeFunc(_decodeUint8)
}
func _decodeUint8(dec *Decoder, p unsafe.Pointer) error {
	v, err := readUint(dec)
	*(*uint8)(p) = uint8(v)
	return err
}
func decodeUint16(t reflect.Type) interface{} {
	return decodeFunc(_decodeUint16)
}
func _decodeUint16(dec *Decoder, p unsafe.Pointer) error {
	v, err := readUint(dec)
	*(*uint16)(p) = uint16(v)
	return err
}
func decodeUint32(t reflect.Type) interface{} {
	return decodeFunc(_decodeUint32)
}
func _decodeUint32(dec *Decoder, p unsafe.Pointer) error {
	v, err := readUint(dec)
	*(*uint32)(p) = uint32(v)
	return err
}
func decodeUint64(t reflect.Type) interface{} {
	return decodeFunc(_decodeUint64)
}
func _decodeUint64(dec *Decoder, p unsafe.Pointer) error {
	v, err := readUint(dec)
	*(*uint64)(p) = v
	return err
}

func decodeFloat32(t reflect.Type) interface{} {
	return decodeFunc(_decodeFloat32)
}

func _decodeFloat32(dec *Decoder, p unsafe.Pointer) error {
	b := dec.b[:5]
	_, err := dec.fr.Read(b)
	if err != nil {
		return err
	}
	if b[0] != 0xCA {
		return ErrorIncorrectType{"float32", b[0]}
	}
	*(*uint32)(p) = binary.BigEndian.Uint32(b[1:])
	return nil
}

func decodeFloat64(t reflect.Type) interface{} {
	return decodeFunc(_decodeFloat64)
}

func _decodeFloat64(dec *Decoder, p unsafe.Pointer) error {
	b := dec.b[:9]
	_, err := dec.fr.Read(b)
	if err != nil {
		return err
	}
	if b[0] != 0xCB {
		return ErrorIncorrectType{"float64", b[0]}
	}
	*(*uint64)(p) = binary.BigEndian.Uint64(b[1:])
	return nil
}

func decodeSlice(t reflect.Type) interface{} {
	if t.Elem().Kind() == reflect.Uint8 {
		return decodeFunc(_decodeByteSlice)
	}
	slDec := getTypeDecoder(t.Elem())
	size := t.Elem().Size()
	_, ok := slDec.(decodeFunc)
	return decodeReflectFunc(func(dec *Decoder, v reflect.Value) error {
		p := unsafe.Pointer(v.UnsafeAddr())
		b, err := dec.r.ReadByte()
		if err != nil {
			return err
		}
		var l int
		switch {
		case b&0xF0 == 0x90:
			l = int(b & 0xF)
		case b == 0xDC:
			bs := dec.b[:2]
			_, err := dec.fr.Read(bs)
			if err != nil {
				return err
			}
			l = int(binary.BigEndian.Uint16(bs))
		case b == 0xDD:
			bs := dec.b[:4]
			_, err := dec.fr.Read(bs)
			if err != nil {
				return err
			}
			l = int(binary.BigEndian.Uint32(bs))
		default:
			return ErrorIncorrectType{"[]slice", b}
		}

		slice := (*reflect.SliceHeader)(p)

		if slice.Len != l {
			v := reflect.MakeSlice(t, l, l)
			ptr := reflect.New(t)
			ptr.Elem().Set(v)
			slice = (*reflect.SliceHeader)(unsafe.Pointer(ptr.Pointer()))
		}
		rSlice := (*reflect.SliceHeader)(p)
		rSlice.Cap = slice.Cap
		rSlice.Len = slice.Len
		rSlice.Data = slice.Data

		if ok {
			for i := 0; i < l; i++ {
				err := (slDec.(decodeFunc))(dec, unsafe.Pointer(slice.Data+size*uintptr(i)))
				if err != nil {
					return err
				}
			}
		} else {
			for i := 0; i < l; i++ {
				vI := v.Index(i)
				if vI.Kind() == reflect.Interface {
					vI = vI.Addr()
				}
				err := (slDec.(decodeReflectFunc))(dec, vI)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func _decodeByteSlice(dec *Decoder, p unsafe.Pointer) error {
	b, err := dec.r.ReadByte()
	if err != nil {
		return err
	}
	var l int
	switch b {
	case 0xC4:
		bs := dec.b[:1]
		_, err := dec.fr.Read(bs)
		if err != nil {
			return err
		}
		l = int(bs[0])
	case 0xC5:
		bs := dec.b[:2]
		_, err := dec.fr.Read(bs)
		if err != nil {
			return err
		}
		l = int(binary.BigEndian.Uint16(bs))
	case 0xC6:
		bs := dec.b[:4]
		_, err := dec.fr.Read(bs)
		if err != nil {
			return err
		}
		l = int(binary.BigEndian.Uint32(bs))
	default:
		return ErrorIncorrectType{"[]byte", b}
	}
	by := *(*[]byte)(p)
	if len(by) != l {
		by = make([]byte, l)
	}
	_, err = dec.fr.Read(by)
	*(*[]byte)(p) = by
	return err
}

func decodeArray(t reflect.Type) interface{} {
	slType := reflect.SliceOf(t.Elem())
	slDec := getTypeDecoder(slType)
	_, ok := slDec.(decodeFunc)
	l := t.Len()
	return decodeReflectFunc(func(dec *Decoder, v reflect.Value) error {
		p := v.UnsafeAddr()
		var err error
		if ok {
			slice := reflect.SliceHeader{
				Len: l, Cap: l,
				Data: p,
			}
			err = (slDec.(decodeFunc))(dec, unsafe.Pointer(&slice))
			if slice.Len != l {
				return ErrorIncorrectLength{l, slice.Len}
			}
		} else {
			slice := v.Slice(0, v.Len())
			ptr := reflect.New(slType)
			ptr.Elem().Set(slice)
			err = (slDec.(decodeReflectFunc))(dec, ptr.Elem())
			if slice.Len() != l {
				return ErrorIncorrectLength{l, slice.Len()}
			}
		}
		return err
	})
}

func decodeMap(t reflect.Type) interface{} {
	if t.Key().Kind() != reflect.String || t.Elem().Kind() != reflect.Interface {
		panic("Only maps of type map[string]interface{} are supported")
	}
	kType := t.Key()
	kDec := getTypeDecoder(kType)
	_, kOk := kDec.(decodeFunc)
	eType := t.Elem()
	eDec := getTypeDecoder(eType)
	_, eOk := eDec.(decodeFunc)
	return decodeReflectFunc(func(dec *Decoder, v reflect.Value) error {
		b, err := dec.r.ReadByte()
		if err != nil {
			return err
		}
		var l int
		switch {
		case b&0xF0 == 0x80:
			l = int(b & 0xF)
		case b == 0xDE:
			bs := dec.b[:2]
			_, err := dec.fr.Read(bs)
			if err != nil {
				return err
			}
			l = int(binary.BigEndian.Uint16(bs))
		case b == 0xDF:
			bs := dec.b[:4]
			_, err := dec.fr.Read(bs)
			if err != nil {
				return err
			}
			l = int(binary.BigEndian.Uint32(bs))
		default:
			return ErrorIncorrectType{"map", b}
		}
		m := reflect.MakeMap(t)
		for i := 0; i < l; i++ {
			k := reflect.New(kType)
			if kOk {
				err := (kDec.(decodeFunc))(dec, unsafe.Pointer(k.Pointer()))
				if err != nil {
					return err
				}
			} else {
				err := (kDec.(decodeReflectFunc))(dec, k)
				if err != nil {
					return err
				}
			}

			e := reflect.New(eType)
			if eOk {
				err := (eDec.(decodeFunc))(dec, unsafe.Pointer(e.Pointer()))
				if err != nil {
					return err
				}
			} else {
				err := (eDec.(decodeReflectFunc))(dec, e)
				if err != nil {
					return err
				}
			}

			m.SetMapIndex(k.Elem(), e.Elem())
		}
		v.Set(m)
		return nil
	})
}

func decodeInterface(t reflect.Type) interface{} {
	return decodeReflectFunc(_decodeInterface)
}

func _decodeInterface(dec *Decoder, v reflect.Value) error {
	b, err := dec.r.ReadByte()
	if err != nil {
		return err
	}
	var l int
	switch b {
	case 0xC7:
		bs := dec.b[:2]
		dec.fr.Read(bs)
		l = int(bs[0])
	case 0xC8:
		bs := dec.b[:3]
		dec.fr.Read(bs)
		l = int(binary.BigEndian.Uint16(bs))
	case 0xC9:
		bs := dec.b[:5]
		dec.fr.Read(bs)
		l = int(binary.BigEndian.Uint32(bs))
	default:
		panic((ErrorIncorrectType{"interface{}", b}).Error())
		return ErrorIncorrectType{"interface{}", b}
	}
	by := make([]byte, l)
	dec.fr.Read(by)
	iDec := &Decoder{
		r: bufio.NewReader(bytes.NewReader(by)),
	}

	id := new(string)
	err = _decodeString(iDec, unsafe.Pointer(id))

	eng := getEngineById(*id)
	if err != nil {
		return err
	}

	val := reflect.New(eng._type)

	if decode, ok := eng.decode.(decodeFunc); ok {
		err = decode(iDec, unsafe.Pointer(val.Elem().UnsafeAddr()))
	} else {
		err = (eng.decode.(decodeReflectFunc))(iDec, val.Elem())
	}
	if err != nil {
		return err
	}

	v.Elem().Set(val.Elem())
	return nil
}
