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
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

var encodeGenerators map[reflect.Kind]func(t reflect.Type) interface{}

func init() {
	encodeGenerators = map[reflect.Kind]func(t reflect.Type) interface{}{
		reflect.Struct:    encodeStruct,
		reflect.String:    encodeString,
		reflect.Bool:      encodeBool,
		reflect.Ptr:       encodePtr,
		reflect.Int:       encodeInt,
		reflect.Int8:      encodeInt8,
		reflect.Int16:     encodeInt16,
		reflect.Int32:     encodeInt32,
		reflect.Int64:     encodeInt64,
		reflect.Uint:      encodeUint,
		reflect.Uint8:     encodeUint8,
		reflect.Uint16:    encodeUint16,
		reflect.Uint32:    encodeUint32,
		reflect.Uint64:    encodeUint64,
		reflect.Float32:   encodeFloat32,
		reflect.Float64:   encodeFloat64,
		reflect.Slice:     encodeSlice,
		reflect.Array:     encodeArray,
		reflect.Map:       encodeMap,
		reflect.Interface: encodeInterface,
	}
}

func getTypeEncoder(t reflect.Type) interface{} {
	gen, ok := encodeGenerators[t.Kind()]
	if !ok {
		panic(fmt.Sprintf("Unhandled kind %s", t.Kind()))
	}
	return gen(t)
}

type eField struct {
	offset       uintptr
	name         string
	index        []int
	encode       interface{}
	needsReflect bool
}

func encodeStruct(t reflect.Type) interface{} {
	var fields []eField
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
		e := eField{
			offset: f.Offset,
			name:   f.Name,
			encode: getTypeEncoder(f.Type),
			index:  f.Index,
		}
		_, e.needsReflect = e.encode.(encodeReflectFunc)
		fields = append(fields, e)
	}
	var b []byte
	l := len(fields)
	switch {
	case l <= 15:
		b = make([]byte, 1)
		b[0] = 0x80 | byte(l)
	case l <= 0xFFFF:
		b = make([]byte, 3)
		b[0] = 0xDE
		binary.BigEndian.PutUint16(b[1:], uint16(l))
	case l <= 0xFFFFFFFF:
		b = make([]byte, 5)
		b[0] = 0xDF
		binary.BigEndian.PutUint32(b[1:], uint32(l))
	}
	return encodeReflectFunc(func(enc *Encoder, v reflect.Value) error {
		p := v.UnsafeAddr()
		enc.w.Write(b)
		for _, e := range fields {
			_encodeString(enc, unsafe.Pointer(&e.name))
			if !e.needsReflect {
				encode := e.encode.(encodeFunc)
				if err := encode(enc, unsafe.Pointer(p+e.offset)); err != nil {
					return err
				}
			} else {
				encode := e.encode.(encodeReflectFunc)
				if err := encode(enc, v.FieldByIndex(e.index)); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func encodeString(t reflect.Type) interface{} {
	return encodeFunc(_encodeString)
}

func _encodeString(enc *Encoder, p unsafe.Pointer) error {
	str := (*reflect.StringHeader)(p)
	var b []byte
	l := str.Len
	switch {
	case l <= 31:
		b = enc.b[:1]
		b[0] = 0xA0 | byte(l)
	case l <= 0xFF:
		b = enc.b[:2]
		b[0] = 0xD9
		b[1] = byte(l)
	case l <= 0xFFFF:
		b = enc.b[:3]
		b[0] = 0xDA
		binary.BigEndian.PutUint16(b[1:], uint16(l))
	case l <= 0xFFFFFFFF:
		b = enc.b[:5]
		b[0] = 0xDB
		binary.BigEndian.PutUint32(b[1:], uint32(l))
	}
	_, err := enc.w.Write(b)
	if err != nil {
		return err
	}
	_, err = enc.w.Write(*(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: str.Data,
		Len:  str.Len,
		Cap:  str.Len,
	})))
	return err
}

func encodeBool(t reflect.Type) interface{} {
	return encodeFunc(_encodeBool)
}

func _encodeBool(enc *Encoder, p unsafe.Pointer) error {
	b := enc.b[:1]
	if *(*bool)(p) {
		b[0] = 0xC3
	} else {
		b[0] = 0xC2
	}
	_, err := enc.w.Write(b)
	return err
}

func encodePtr(t reflect.Type) interface{} {
	elemEnc := getTypeEncoder(t.Elem())
	_, ok := elemEnc.(encodeFunc)
	return encodeReflectFunc(func(enc *Encoder, v reflect.Value) error {
		p := unsafe.Pointer(v.UnsafeAddr())
		rp := *(*uintptr)(p)
		if rp == 0 {
			b := enc.b[:1]
			b[0] = 0xC0
			_, err := enc.w.Write(b)
			return err
		}
		if ok {
			eFunc := elemEnc.(encodeFunc)
			return eFunc(enc, unsafe.Pointer(rp))
		} else {
			return (elemEnc.(encodeReflectFunc))(enc, v.Elem())
		}
		return nil
	})
}

func writeInt(enc *Encoder, v int64) error {
	var err error
	switch {
	case v <= 127 && v >= -128:
		bs := enc.b[:2]
		bs[0] = 0xD0
		bs[1] = byte(int8(v))
		_, err = enc.w.Write(bs)
	case v <= 32767 && v >= -32768:
		bs := enc.b[:3]
		bs[0] = 0xD1
		binary.BigEndian.PutUint16(bs[1:], uint16(int16(v)))
		_, err = enc.w.Write(bs)
	case v <= 2147483647 && v >= -2147483648:
		bs := enc.b[:5]
		bs[0] = 0xD2
		binary.BigEndian.PutUint32(bs[1:], uint32(int32(v)))
		_, err = enc.w.Write(bs)
	default:
		bs := enc.b[:9]
		bs[0] = 0xD3
		binary.BigEndian.PutUint64(bs[1:], uint64(int64(v)))
		_, err = enc.w.Write(bs)
	}
	return err
}

func encodeInt(t reflect.Type) interface{} {
	return encodeFunc(_encodeInt)
}
func _encodeInt(enc *Encoder, p unsafe.Pointer) error {
	return writeInt(enc, int64(*(*int)(p)))
}
func encodeInt8(t reflect.Type) interface{} {
	return encodeFunc(_encodeInt8)
}
func _encodeInt8(enc *Encoder, p unsafe.Pointer) error {
	return writeInt(enc, int64(*(*int8)(p)))
}
func encodeInt16(t reflect.Type) interface{} {
	return encodeFunc(_encodeInt16)
}
func _encodeInt16(enc *Encoder, p unsafe.Pointer) error {
	return writeInt(enc, int64(*(*int16)(p)))
}
func encodeInt32(t reflect.Type) interface{} {
	return encodeFunc(_encodeInt32)
}
func _encodeInt32(enc *Encoder, p unsafe.Pointer) error {
	return writeInt(enc, int64(*(*int32)(p)))
}
func encodeInt64(t reflect.Type) interface{} {
	return encodeFunc(_encodeInt64)
}
func _encodeInt64(enc *Encoder, p unsafe.Pointer) error {
	return writeInt(enc, *(*int64)(p))
}

func writeUint(enc *Encoder, v uint64) error {
	var err error
	switch {
	case v <= 0xFF:
		bs := enc.b[:2]
		bs[0] = 0xCC
		bs[1] = byte(v)
		_, err = enc.w.Write(bs)
	case v <= 0xFFFF:
		bs := enc.b[:3]
		bs[0] = 0xCD
		binary.BigEndian.PutUint16(bs[1:], uint16(v))
		_, err = enc.w.Write(bs)
	case v <= 0xFFFFFFFF:
		bs := enc.b[:5]
		bs[0] = 0xCE
		binary.BigEndian.PutUint32(bs[1:], uint32(v))
		_, err = enc.w.Write(bs)
	default:
		bs := enc.b[:9]
		bs[0] = 0xCF
		binary.BigEndian.PutUint64(bs[1:], v)
		_, err = enc.w.Write(bs)
	}
	return err
}

func encodeUint(t reflect.Type) interface{} {
	return encodeFunc(_encodeUint)
}
func _encodeUint(enc *Encoder, p unsafe.Pointer) error {
	return writeUint(enc, uint64(*(*uint)(p)))
}
func encodeUint8(t reflect.Type) interface{} {
	return encodeFunc(_encodeUint8)
}
func _encodeUint8(enc *Encoder, p unsafe.Pointer) error {
	return writeUint(enc, uint64(*(*uint8)(p)))
}
func encodeUint16(t reflect.Type) interface{} {
	return encodeFunc(_encodeUint16)
}
func _encodeUint16(enc *Encoder, p unsafe.Pointer) error {
	return writeUint(enc, uint64(*(*uint16)(p)))
}
func encodeUint32(t reflect.Type) interface{} {
	return encodeFunc(_encodeUint32)
}
func _encodeUint32(enc *Encoder, p unsafe.Pointer) error {
	return writeUint(enc, uint64(*(*uint32)(p)))
}
func encodeUint64(t reflect.Type) interface{} {
	return encodeFunc(_encodeUint64)
}
func _encodeUint64(enc *Encoder, p unsafe.Pointer) error {
	return writeUint(enc, *(*uint64)(p))
}

func encodeFloat32(t reflect.Type) interface{} {
	return encodeFunc(_encodeFloat32)
}

func _encodeFloat32(enc *Encoder, p unsafe.Pointer) error {
	b := enc.b[:5]
	b[0] = 0xCA
	binary.BigEndian.PutUint32(b[1:], *(*uint32)(p))
	_, err := enc.w.Write(b)
	return err
}

func encodeFloat64(t reflect.Type) interface{} {
	return encodeFunc(_encodeFloat64)
}

func _encodeFloat64(enc *Encoder, p unsafe.Pointer) error {
	b := enc.b[:9]
	b[0] = 0xCB
	binary.BigEndian.PutUint64(b[1:], *(*uint64)(p))
	_, err := enc.w.Write(b)
	return err
}

func encodeSlice(t reflect.Type) interface{} {
	if t.Elem().Kind() == reflect.Uint8 {
		return encodeFunc(_encodeByteSlice)
	}
	slEnc := getTypeEncoder(t.Elem())
	size := t.Elem().Size()
	_, ok := slEnc.(encodeFunc)
	return encodeReflectFunc(func(enc *Encoder, v reflect.Value) error {
		p := v.Pointer()
		var b []byte
		l := v.Len()
		switch {
		case l <= 15:
			b = enc.b[:1]
			b[0] = 0x90 | byte(l)
		case l <= 0xFFFF:
			b = enc.b[:3]
			b[0] = 0xDC
			binary.BigEndian.PutUint16(b[1:], uint16(l))
		case l <= 0xFFFFFFFF:
			b = enc.b[:5]
			b[0] = 0xDD
			binary.BigEndian.PutUint32(b[1:], uint32(l))
		}
		_, err := enc.w.Write(b)
		if err != nil {
			return err
		}
		if ok {
			encF := slEnc.(encodeFunc)
			for i := 0; i < l; i++ {
				err := encF(enc, unsafe.Pointer(p+size*uintptr(i)))
				if err != nil {
					return err
				}
			}
		} else {
			encF := slEnc.(encodeReflectFunc)
			for i := 0; i < l; i++ {
				err := encF(enc, v.Index(i))
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func _encodeByteSlice(enc *Encoder, p unsafe.Pointer) error {
	b := *(*[]byte)(p)
	l := len(b)
	var bs []byte
	switch {
	case l <= 0xFF:
		bs = enc.b[:2]
		bs[0] = 0xC4
		bs[1] = byte(l)
	case l <= 0xFFFF:
		bs = enc.b[:3]
		bs[0] = 0xC5
		binary.BigEndian.PutUint16(bs[1:], uint16(l))
	case l <= 0xFFFFFFFF:
		bs = enc.b[:6]
		bs[0] = 0xC6
		binary.BigEndian.PutUint32(bs[1:], uint32(l))
	}
	_, err := enc.w.Write(bs)
	if err != nil {
		return err
	}
	_, err = enc.w.Write(b)
	return err
}

func encodeArray(t reflect.Type) interface{} {
	//Cheat by turning the array into a slice
	slEnc := getTypeEncoder(reflect.SliceOf(t.Elem()))
	l := t.Len()
	_, ok := slEnc.(encodeFunc)
	return encodeReflectFunc(func(enc *Encoder, v reflect.Value) error {
		if ok {
			slice := reflect.SliceHeader{ //Convert to slice
				Data: v.UnsafeAddr(),
				Len:  l,
				Cap:  l,
			}
			encF := slEnc.(encodeFunc)
			return encF(enc, unsafe.Pointer(&slice))
		} else {
			encF := slEnc.(encodeReflectFunc)
			return encF(enc, v.Slice(0, v.Len()))
		}
	})
}

func encodeMap(t reflect.Type) interface{} {
	kType := t.Key()
	kEnc := getTypeEncoder(kType)
	_, kOk := kEnc.(encodeFunc)
	eType := t.Elem()
	eEnc := getTypeEncoder(eType)
	_, eOk := eEnc.(encodeFunc)
	return encodeReflectFunc(func(enc *Encoder, v reflect.Value) error {
		var b []byte
		l := v.Len()
		switch {
		case l <= 15:
			b = enc.b[:1]
			b[0] = 0x80 | byte(l)
		case l <= 0xFFFF:
			b = enc.b[:3]
			b[0] = 0xDE
			binary.BigEndian.PutUint16(b[1:], uint16(l))
		case l <= 0xFFFFFFFF:
			b = enc.b[:5]
			b[0] = 0xDF
			binary.BigEndian.PutUint32(b[1:], uint32(l))
		}
		enc.w.Write(b)
		for _, k := range v.MapKeys() {
			val := v.MapIndex(k)

			kPtr := reflect.New(kType)
			kPtr.Elem().Set(k)
			if kOk {
				encF := kEnc.(encodeFunc)
				err := encF(enc, unsafe.Pointer(kPtr.Elem().UnsafeAddr()))
				if err != nil {
					return err
				}
			} else {
				encF := kEnc.(encodeReflectFunc)
				err := encF(enc, kPtr.Elem())
				if err != nil {
					return err
				}
			}

			ePtr := reflect.New(eType)
			ePtr.Elem().Set(val)
			if eOk {
				encF := eEnc.(encodeFunc)
				err := encF(enc, unsafe.Pointer(ePtr.Elem().UnsafeAddr()))
				if err != nil {
					return err
				}
			} else {
				encF := eEnc.(encodeReflectFunc)
				err := encF(enc, ePtr.Elem())
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func encodeInterface(t reflect.Type) interface{} {
	return encodeReflectFunc(_encodeInterface)
}

func _encodeInterface(enc *Encoder, v reflect.Value) error {
	var buf bytes.Buffer
	iEnc := &Encoder{
		w: &buf,
	}

	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	eng := getEngine(v.Type())

	err := _encodeString(iEnc, unsafe.Pointer(&eng.id))
	if err != nil {
		return err
	}

	vPtr := reflect.New(v.Type())
	vPtr.Elem().Set(v)
	if e, ok := eng.encode.(encodeFunc); ok {
		err := e(iEnc, unsafe.Pointer(vPtr.Elem().UnsafeAddr()))
		if err != nil {
			return err
		}
	} else {
		e := eng.encode.(encodeReflectFunc)
		err := e(iEnc, vPtr.Elem())
		if err != nil {
			return err
		}
	}
	l := buf.Len()
	var bs []byte
	switch {
	case l <= 0xFF:
		bs = enc.b[:3]
		bs[0] = 0xC7
		bs[1] = byte(l)
		bs[2] = 5
	case l <= 0xFFFF:
		bs = enc.b[:4]
		bs[0] = 0xC8
		binary.BigEndian.PutUint16(bs[1:], uint16(l))
		bs[3] = 5
	case l <= 0xFFFFFFFF:
		bs = enc.b[:6]
		bs[0] = 0xC9
		binary.BigEndian.PutUint32(bs[1:], uint32(l))
		bs[5] = 5
	}
	_, err = enc.w.Write(bs)
	if err != nil {
		return err
	}
	_, err = buf.WriteTo(enc.w)
	return err
}
