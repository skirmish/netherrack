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
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"sync"
)

func init() {
	fieldCache.m = make(map[reflect.Type]map[string]interface{})
	fieldCache.create = make(map[reflect.Type]*sync.WaitGroup)
}

type fullReader struct {
	io.Reader
}

func (f fullReader) Read(p []byte) (int, error) {
	return io.ReadFull(f.Reader, p)
}

//Reads in the msgpack format from r into v. Panics if
//v is not a pointer to a struct.
func Read(r io.Reader, v interface{}) (err error) {
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Ptr {
		panic("format/msgpack: v is not a pointer")
	}
	val := ptr.Elem()
	if val.Kind() != reflect.Struct {
		panic("format/msgpack: v isn't a pointer to a struct")
	}
	//Catch any errors to prevent crashing
	defer func() {
		if e := recover(); e != nil {
			if er, ok := e.(error); ok {
				err = er
			} else {
				err = fmt.Errorf("format/msgpack: %s", e)
			}
			return
		}
	}()

	fs := fields(val.Type())
	de := &msgDecoder{}
	return read(fullReader{r}, de, fs, val)
}

//Reads a struct map from r into val
func read(r io.Reader, de *msgDecoder, fs map[string]interface{}, val reflect.Value) error {
	m := de.b[:1]
	_, err := r.Read(m)
	if err != nil {
		return err
	}

	count := 0
	if m[0]&0xF0 == 0x80 {
		count = int(m[0] & 0x0F)
	} else if m[0] == 0xDE {
		m := de.b[:2]
		count = int(binary.LittleEndian.Uint16(m))
	} else if m[0] == 0xDF {
		m := de.b[:4]
		count = int(binary.LittleEndian.Uint32(m))
	}

	for i := 0; i < count; i++ {
		name, err := readString(r, de)
		if err != nil {
			return err
		}
		i, ok := fs[name]
		if !ok {
			err := skip(r, de)
			if err != nil {
				return err
			}
			continue
		}
		if f, ok := i.(field); ok {
			val := val.Field(f.sField)
			err := f.read(r, de, val)
			if err != nil {
				return err
			}
		} else {
			f := i.(fieldStruct)
			val := val.Field(f.sField)
			err := read(r, de, f.m, val)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

//Writes v into w in the msgpack format.
func Write(w io.Writer, v interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if er, ok := e.(error); ok {
				err = er
			} else {
				err = fmt.Errorf("format/msgpack: %s", e)
			}
			return
		}
	}()
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	fs := fields(val.Type())
	en := &msgEncoder{}
	return write(w, en, fs, val)
}

//Writes val into w using the struct map fs
func write(w io.Writer, en *msgEncoder, fs map[string]interface{}, val reflect.Value) error {
	count := len(fs)
	var bs []byte
	if count <= 15 {
		bs = en.b[:1]
		bs[0] = 0x80 | byte(count)
	} else if count <= 0xFFFF {
		bs = en.b[:3]
		bs[0] = 0xDE
		binary.LittleEndian.PutUint16(bs[1:], uint16(count))
	} else {
		bs = en.b[:5]
		bs[0] = 0xDF
		binary.LittleEndian.PutUint32(bs[1:], uint32(count))
	}
	w.Write(bs)
	for _, i := range fs {
		if f, ok := i.(field); ok {
			w.Write(f.nameBytes)
			v := val.Field(f.sField)
			err := f.write(w, en, v)
			if err != nil {
				return err
			}
		} else {
			fs := i.(fieldStruct)
			w.Write(fs.nameBytes)
			v := val.Field(fs.sField)
			err := write(w, en, fs.m, v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

var fieldCache struct {
	sync.RWMutex
	m      map[reflect.Type]map[string]interface{}
	create map[reflect.Type]*sync.WaitGroup
}

//Returns the fields for the type t. This method caches the results making
//later calls cheap.
func fields(t reflect.Type) map[string]interface{} {
	fieldCache.RLock()
	fs := fieldCache.m[t]
	fieldCache.RUnlock()

	//Cached version exists
	if fs != nil {
		return fs
	}
	//This is to prevent multiple goroutines computing the same thing
	fieldCache.Lock()
	var sy *sync.WaitGroup
	if sy, ok := fieldCache.create[t]; ok {
		fieldCache.Unlock()
		sy.Wait()
		return fields(t)
	}
	sy = &sync.WaitGroup{}
	fieldCache.create[t] = sy
	sy.Add(1)
	fieldCache.Unlock()

	fs = compileStruct(t)

	fieldCache.Lock()
	fieldCache.m[t] = fs
	fieldCache.Unlock()
	sy.Done()
	return fs
}

//Loops through the fields of the struct and returns a slice of fields.
//ind is the offset of the struct in the root struct.
func compileStruct(t reflect.Type) map[string]interface{} {
	fs := map[string]interface{}{}
	count := t.NumField()
	for i := 0; i < count; i++ {
		f := t.Field(i)
		if !f.Anonymous {
			name := f.Name
			if tName := f.Tag.Get("msgpack"); len(tName) > 0 {
				name = tName
			}
			if name == "ignore" {
				continue
			}
			fs[name] = compileField(f, name)
		}
	}
	return fs
}

//A field contains the methods needed to read and write the
//field.
type field struct {
	sField    int
	nameBytes []byte
	write     encoder
	read      decoder
}

//A special type of field which contains a struct
type fieldStruct struct {
	sField    int
	nameBytes []byte
	m         map[string]interface{}
}

type encoder func(w io.Writer, en *msgEncoder, field reflect.Value) error
type decoder func(r io.Reader, de *msgDecoder, field reflect.Value) error

//Returns the field or fields needed to fully write the struct's field
func compileField(sf reflect.StructField, name string) interface{} {
	f := field{sField: sf.Index[0]}

	var buf bytes.Buffer
	writeString(&buf, &msgEncoder{}, name)
	f.nameBytes = buf.Bytes()

	switch sf.Type.Kind() {
	case reflect.Ptr:
		fs := compileField(reflect.StructField{Type: sf.Type.Elem(), Index: []int{0}}, "*"+sf.Name)
		if fi, ok := fs.(field); ok {
			elem := sf.Type.Elem()
			zero := reflect.Zero(sf.Type)
			f.read = func(r io.Reader, de *msgDecoder, field reflect.Value) error {
				buf := bufio.NewReader(r)
				if b, err := buf.Peek(1); b[0] == 0xC0 {
					buf.ReadByte()
					field.Set(zero)
					return err
				}
				val := reflect.New(elem)
				if err := fi.read(buf, de, val); err != nil {
					return err
				}
				field.Set(val)
				return nil
			}
			f.write = func(w io.Writer, en *msgEncoder, field reflect.Value) error {
				if field.IsNil() {
					w.Write([]byte{0xC0})
					return nil
				}
				val := field.Elem()
				if err := fi.write(w, en, val); err != nil {
					return err
				}
				return nil
			}
		}
	case reflect.Struct:
		return fieldStruct{f.sField, f.nameBytes, compileStruct(sf.Type)}
	case reflect.Bool:
		f.write = encodeBool
		f.read = decodeBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		f.write = encodeInt
		f.read = decodeInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		f.write = encodeUint
		f.read = decodeUint
	case reflect.String:
		f.write = encodeString
		f.read = decodeString
	case reflect.Map:
		key := sf.Type.Key()
		elem := sf.Type.Elem()
		var keyField interface{}
		var elemField interface{}
		name := "map:" + sf.Name
		keyField = compileField(reflect.StructField{Type: key, Index: []int{0}}, name)
		if elem.Kind() != reflect.Interface {
			elemField = compileField(reflect.StructField{Type: elem, Index: []int{0}}, name)
		}
		f.write = func(w io.Writer, en *msgEncoder, fi reflect.Value) error {
			count := fi.Len()
			var bs []byte
			if count <= 15 {
				bs = en.b[:1]
				bs[0] = 0x80 | byte(count)
			} else if count <= 0xFFFF {
				bs = en.b[:3]
				bs[0] = 0xDE
				binary.LittleEndian.PutUint16(bs[1:], uint16(count))
			} else {
				bs = en.b[:5]
				bs[0] = 0xDF
				binary.LittleEndian.PutUint32(bs[1:], uint32(count))
			}
			_, err := w.Write(bs)
			if err != nil {
				return err
			}
			keys := fi.MapKeys()
			for _, key := range keys {
				if f, ok := keyField.(field); ok {
					v := key
					err := f.write(w, en, v)
					if err != nil {
						return err
					}
				} else {
					fs := keyField.(fieldStruct)
					v := key
					err := write(w, en, fs.m, v)
					if err != nil {
						return err
					}
				}
				if f, ok := elemField.(field); ok {
					v := fi.MapIndex(key)
					err := f.write(w, en, v)
					if err != nil {
						return err
					}
				} else {
					if elemField == nil {
						v := fi.MapIndex(key).Elem()
						temp := compileField(reflect.StructField{Type: v.Type(), Index: []int{0}}, "")
						if f, ok := temp.(field); ok {
							err := f.write(w, en, v)
							if err != nil {
								return err
							}
						} else {
							fs := temp.(fieldStruct)
							err := write(w, en, fs.m, v)
							if err != nil {
								return err
							}
						}
					} else {
						fs := elemField.(fieldStruct)
						v := fi.MapIndex(key)
						err := write(w, en, fs.m, v)
						if err != nil {
							return err
						}
					}
				}
			}
			return nil
		}
		f.read = func(r io.Reader, de *msgDecoder, fi reflect.Value) error {
			m := de.b[:1]
			_, err := r.Read(m)
			if err != nil {
				return err
			}

			count := 0
			if m[0]&0xF0 == 0x80 {
				count = int(m[0] & 0x0F)
			} else if m[0] == 0xDE {
				m := de.b[:2]
				count = int(binary.LittleEndian.Uint16(m))
			} else if m[0] == 0xDF {
				m := de.b[:4]
				count = int(binary.LittleEndian.Uint32(m))
			}

			ma := reflect.MakeMap(sf.Type)

			for i := 0; i < count; i++ {
				val := reflect.New(key).Elem()
				if f, ok := keyField.(field); ok {
					err := f.read(r, de, val)
					if err != nil {
						return err
					}
				} else {
					f := keyField.(fieldStruct)
					err := read(r, de, f.m, val)
					if err != nil {
						return err
					}
				}
				keyVal := val

				if f, ok := elemField.(field); ok {
					val = reflect.New(elem)
					err := f.read(r, de, val)
					if err != nil {
						return err
					}
				} else {
					if elemField == nil {
						v, err := fallbackRead(r, de)
						if err != nil {
							return err
						}
						val = reflect.ValueOf(v)
					} else {
						val = reflect.New(elem)
						fs := elemField.(fieldStruct)
						err := read(r, de, fs.m, val)
						if err != nil {
							return err
						}
					}
				}
				ma.SetMapIndex(keyVal, val)
			}
			fi.Set(ma)
			return nil
		}
	case reflect.Slice:
		elem := sf.Type.Elem()
		switch elem.Kind() {
		case reflect.Uint8: //Short-cut for byte arrays
			f.write = func(w io.Writer, en *msgEncoder, fi reflect.Value) error {
				l := fi.Len()
				var bs []byte
				if l <= 0xFF {
					bs = en.b[:2]
					bs[0] = 0xC4
					bs[1] = byte(l)
				} else if l <= 0xFFFF {
					bs = en.b[:3]
					bs[0] = 0xC5
					binary.LittleEndian.PutUint16(bs[1:], uint16(l))
				} else {
					bs = en.b[:5]
					bs[0] = 0xC6
					binary.LittleEndian.PutUint32(bs[1:], uint32(l))
				}
				_, err := w.Write(bs)
				if err != nil {
					return err
				}
				_, err = w.Write(fi.Bytes())
				return err
			}
			f.read = func(r io.Reader, de *msgDecoder, fi reflect.Value) error {
				bs := de.b[:1]
				_, err := r.Read(bs)
				if err != nil {
					return err
				}
				var l int
				switch bs[0] {
				case 0xC4:
					_, err := r.Read(bs)
					if err != nil {
						return err
					}
					l = int(bs[0])
				case 0xC5:
					bs = de.b[:2]
					_, err := r.Read(bs)
					if err != nil {
						return err
					}
					l = int(binary.LittleEndian.Uint16(bs))
				case 0xC6:
					bs = de.b[:4]
					_, err := r.Read(bs)
					if err != nil {
						return err
					}
					l = int(binary.LittleEndian.Uint32(bs))
				}
				out := make([]byte, l)
				_, err = r.Read(out)
				if err != nil {
					return nil
				}
				fi.SetBytes(out)
				return nil
			}
		case reflect.Interface: //Manual method
			f.write = func(w io.Writer, en *msgEncoder, fi reflect.Value) error {
				l := fi.Len()
				var bs []byte
				if l <= 0xF {
					bs = en.b[:1]
					bs[0] = 0x90 | byte(l)
				} else if l <= 0xFFFF {
					bs = en.b[:3]
					bs[0] = 0xDC
					binary.LittleEndian.PutUint16(bs[1:], uint16(l))
				} else {
					bs = en.b[:5]
					bs[0] = 0xDD
					binary.LittleEndian.PutUint32(bs[1:], uint32(l))
				}
				_, err := w.Write(bs)
				if err != nil {
					return err
				}
				for i := 0; i < l; i++ {
					v := fi.Index(i).Elem()
					temp := compileField(reflect.StructField{Type: v.Type(), Index: []int{0}}, "")
					if f, ok := temp.(field); ok {
						err := f.write(w, en, v)
						if err != nil {
							return err
						}
					} else {
						fs := temp.(fieldStruct)
						err := write(w, en, fs.m, v)
						if err != nil {
							return err
						}
					}
				}
				return nil
			}
			f.read = func(r io.Reader, de *msgDecoder, fi reflect.Value) error {
				var l int
				bs := de.b[:1]
				_, err := r.Read(bs)
				if err != nil {
					return err
				}
				switch bs[0] {
				case 0xDC: //Array 16-Bit
					bs = de.b[:2]
					_, err := r.Read(bs)
					if err != nil {
						return err
					}
					l = int(binary.LittleEndian.Uint16(bs))
				case 0xDD: //Array 32-Bit
					bs = de.b[:4]
					_, err := r.Read(bs)
					if err != nil {
						return err
					}
					l = int(binary.LittleEndian.Uint32(bs))
				default:
					if bs[0]&0xF0 == 0x90 {
						l = int(bs[0] & 0x0F)
					}
				}
				val, err := readArray(r, de, l)
				if err != nil {
					return err
				}
				fi.Set(reflect.ValueOf(val))
				return nil
			}
		default:
			name := "slice:" + sf.Name
			elemField := compileField(reflect.StructField{Type: elem, Index: []int{0}}, name)
			f.write = func(w io.Writer, en *msgEncoder, fi reflect.Value) error {
				l := fi.Len()
				var bs []byte
				if l <= 0xF {
					bs = en.b[:1]
					bs[0] = 0x90 | byte(l)
				} else if l <= 0xFFFF {
					bs = en.b[:3]
					bs[0] = 0xDC
					binary.LittleEndian.PutUint16(bs[1:], uint16(l))
				} else {
					bs = en.b[:5]
					bs[0] = 0xDD
					binary.LittleEndian.PutUint32(bs[1:], uint32(l))
				}
				_, err := w.Write(bs)
				if err != nil {
					return err
				}
				if f, ok := elemField.(field); ok {
					for i := 0; i < l; i++ {
						v := fi.Index(i)
						err := f.write(w, en, v)
						if err != nil {
							return err
						}
					}
				} else {
					f := elemField.(fieldStruct)
					for i := 0; i < l; i++ {
						v := fi.Index(i)
						err := write(w, en, f.m, v)
						if err != nil {
							return err
						}
					}
				}
				return nil
			}
			f.read = func(r io.Reader, de *msgDecoder, fi reflect.Value) error {
				var l int
				bs := de.b[:1]
				_, err := r.Read(bs)
				if err != nil {
					return err
				}
				switch bs[0] {
				case 0xDC: //Array 16-Bit
					bs = de.b[:2]
					_, err := r.Read(bs)
					if err != nil {
						return err
					}
					l = int(binary.LittleEndian.Uint16(bs))
				case 0xDD: //Array 32-Bit
					bs = de.b[:4]
					_, err := r.Read(bs)
					if err != nil {
						return err
					}
					l = int(binary.LittleEndian.Uint32(bs))
				default:
					if bs[0]&0xF0 == 0x90 {
						l = int(bs[0] & 0x0F)
					}
				}
				val := reflect.MakeSlice(sf.Type, l, l)
				if f, ok := elemField.(field); ok {
					for i := 0; i < l; i++ {
						v := val.Index(i)
						err := f.read(r, de, v)
						if err != nil {
							return err
						}
					}
				} else {
					f := elemField.(fieldStruct)
					for i := 0; i < l; i++ {
						v := val.Index(i)
						err := read(r, de, f.m, v)
						if err != nil {
							return err
						}
					}
				}
				fi.Set(val)
				return nil
			}
		}
	case reflect.Float32:
		f.write = encodeFloat32
		f.read = decodeFloat32
	case reflect.Float64:
		f.write = encodeFloat64
		f.read = decodeFloat64
	default:
		panic(fmt.Errorf("Unhandled type %s for %s", sf.Type.Kind().String(), sf.Name))
	}
	return f
}

type msgEncoder struct {
	b [8]byte
}

type msgDecoder struct {
	b [8]byte
}

var ErrorIncorrectType = errors.New("format/msgpack: Incorrect type")

func encodeBool(w io.Writer, en *msgEncoder, field reflect.Value) error {
	bs := en.b[:1]
	if field.Bool() {
		bs[0] = 0xC3
	} else {
		bs[0] = 0xC2
	}
	_, err := w.Write(bs)
	return err
}

func decodeBool(r io.Reader, de *msgDecoder, field reflect.Value) error {
	bs := de.b[:1]
	_, err := r.Read(bs)
	if bs[0] != 0xC3 && bs[0] != 0xC2 {
		return ErrorIncorrectType
	}
	field.SetBool(bs[0] == 0xC3)
	return err
}

func encodeInt(w io.Writer, en *msgEncoder, field reflect.Value) error {
	val := field.Int()
	var err error
	switch {
	case val <= 127 && val >= -128:
		bs := en.b[:2]
		bs[0] = 0xD0
		bs[1] = byte(val)
		_, err = w.Write(bs)
	case val <= 32767 && val >= -32768:
		bs := en.b[:3]
		bs[0] = 0xD1
		binary.LittleEndian.PutUint16(bs[1:], uint16(val))
		_, err = w.Write(bs)
	case val <= 2147483647 && val >= -2147483648:
		bs := en.b[:5]
		bs[0] = 0xD2
		binary.LittleEndian.PutUint32(bs[1:], uint32(val))
		_, err = w.Write(bs)
	default:
		bs := en.b[:1]
		bs[0] = 0xD3
		_, err = w.Write(bs)
		bs = en.b[:8]
		binary.LittleEndian.PutUint64(bs, uint64(val))
		_, err = w.Write(bs)
	}
	return err
}

func decodeInt(r io.Reader, de *msgDecoder, field reflect.Value) error {
	bs := de.b[:1]
	_, err := r.Read(bs)
	if err != nil {
		return err
	}
	switch bs[0] {
	case 0xD0:
		bs = de.b[:1]
		_, err := r.Read(bs)
		if err != nil {
			return err
		}
		field.SetInt(int64(int8(bs[0])))
	case 0xD1:
		bs = de.b[:2]
		_, err := r.Read(bs)
		if err != nil {
			return err
		}
		field.SetInt(int64(int16(binary.LittleEndian.Uint16(bs))))
	case 0xD2:
		bs = de.b[:4]
		_, err := r.Read(bs)
		if err != nil {
			return err
		}
		field.SetInt(int64(int32(binary.LittleEndian.Uint32(bs))))
	case 0xD3:
		bs = de.b[:8]
		_, err := r.Read(bs)
		if err != nil {
			return err
		}
		field.SetInt(int64(binary.LittleEndian.Uint64(bs)))
	default:
		return ErrorIncorrectType
	}
	return nil
}

func encodeUint(w io.Writer, en *msgEncoder, field reflect.Value) error {
	val := field.Uint()
	var err error
	switch {
	case val <= 0xFF:
		bs := en.b[:2]
		bs[0] = 0xCC
		bs[1] = byte(val)
		_, err = w.Write(bs)
	case val <= 0xFF:
		bs := en.b[:3]
		bs[0] = 0xCD
		binary.LittleEndian.PutUint16(bs[1:], uint16(val))
		_, err = w.Write(bs)
	case val <= 0xFFFF:
		bs := en.b[:5]
		bs[0] = 0xCE
		binary.LittleEndian.PutUint32(bs[1:], uint32(val))
		_, err = w.Write(bs)
	default:
		bs := en.b[:1]
		bs[0] = 0xCF
		_, err = w.Write(bs)
		bs = en.b[:8]
		binary.LittleEndian.PutUint64(bs[1:], val)
		_, err = w.Write(bs)
	}
	return err
}

func decodeUint(r io.Reader, de *msgDecoder, field reflect.Value) error {
	bs := de.b[:1]
	_, err := r.Read(bs)
	if err != nil {
		return err
	}
	switch bs[0] {
	case 0xCC:
		bs = de.b[:1]
		_, err := r.Read(bs)
		if err != nil {
			return err
		}
		field.SetUint(uint64(bs[0]))
	case 0xCD:
		bs = de.b[:2]
		_, err := r.Read(bs)
		if err != nil {
			return err
		}
		field.SetUint(uint64(binary.LittleEndian.Uint16(bs)))
	case 0xCE:
		bs = de.b[:4]
		_, err := r.Read(bs)
		if err != nil {
			return err
		}
		field.SetUint(uint64(binary.LittleEndian.Uint32(bs)))
	case 0xCF:
		bs = de.b[:8]
		_, err := r.Read(bs)
		if err != nil {
			return err
		}
		field.SetUint(binary.LittleEndian.Uint64(bs))
	default:
		return ErrorIncorrectType
	}
	return nil
}

func encodeString(w io.Writer, en *msgEncoder, field reflect.Value) error {
	return writeString(w, en, field.String())
}

func decodeString(r io.Reader, de *msgDecoder, field reflect.Value) error {
	str, err := readString(r, de)
	if err != nil {
		return err
	}
	field.SetString(str)
	return nil
}

func writeString(w io.Writer, en *msgEncoder, str string) error {
	l := len(str)
	var bs []byte
	switch {
	case l <= 31:
		bs = en.b[:1]
		bs[0] = 0xA0 | byte(l)
	case l <= 0xFF:
		bs = en.b[:2]
		bs[0] = 0xD9
		bs[1] = byte(l)
	case l <= 0xFFFF:
		bs = en.b[:3]
		bs[0] = 0xDA
		binary.LittleEndian.PutUint16(bs[1:], uint16(l))
	default:
		bs = en.b[:5]
		bs[0] = 0xDB
		binary.LittleEndian.PutUint32(bs[1:], uint32(l))
	}
	_, err := w.Write(bs)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(str))
	if err != nil {
		return err
	}
	return nil
}

func readString(r io.Reader, de *msgDecoder) (string, error) {
	l := 0
	bs := de.b[:1]
	_, err := r.Read(bs)
	if err != nil {
		return "", err
	}
	if bs[0]&0xE0 == 0xA0 {
		l = int(bs[0] & 0x1F)
	} else {
		switch bs[0] {
		case 0xD9:
			bs = de.b[:1]
			_, err := r.Read(bs)
			if err != nil {
				return "", err
			}
			l = int(bs[0])
		case 0xDA:
			bs = de.b[:2]
			_, err := r.Read(bs)
			if err != nil {
				return "", err
			}
			l = int(binary.LittleEndian.Uint16(bs))
		case 0xDB:
			bs = de.b[:4]
			_, err := r.Read(bs)
			if err != nil {
				return "", err
			}
			l = int(binary.LittleEndian.Uint32(bs))
		default:
			return "", ErrorIncorrectType
		}
	}
	str := make([]byte, l)
	_, err = r.Read(str)
	if err != nil {
		return "", err
	}
	return string(str), nil
}

func encodeFloat32(w io.Writer, en *msgEncoder, field reflect.Value) error {
	bs := en.b[:1]
	bs[0] = 0xCA
	_, err := w.Write(bs)
	if err != nil {
		return err
	}
	bs = en.b[:4]
	binary.LittleEndian.PutUint32(bs, math.Float32bits(float32(field.Float())))
	_, err = w.Write(bs)
	if err != nil {
		return err
	}
	return nil
}

func decodeFloat32(r io.Reader, de *msgDecoder, field reflect.Value) error {
	_, err := r.Read(de.b[:1])
	if err != nil {
		return err
	}
	if de.b[0] != 0xCA {
		return ErrorIncorrectType
	}
	bs := de.b[:4]
	_, err = r.Read(bs)
	if err != nil {
		return err
	}
	field.SetFloat(float64(math.Float32frombits(binary.LittleEndian.Uint32(bs))))
	return nil
}

func encodeFloat64(w io.Writer, en *msgEncoder, field reflect.Value) error {
	bs := en.b[:1]
	bs[0] = 0xCB
	_, err := w.Write(bs)
	if err != nil {
		return err
	}
	bs = en.b[:8]
	binary.LittleEndian.PutUint64(bs, math.Float64bits(field.Float()))
	_, err = w.Write(bs)
	if err != nil {
		return err
	}
	return nil
}

func decodeFloat64(r io.Reader, de *msgDecoder, field reflect.Value) error {
	_, err := r.Read(de.b[:1])
	if err != nil {
		return err
	}
	if de.b[0] != 0xCB {
		return ErrorIncorrectType
	}
	bs := de.b[:8]
	_, err = r.Read(bs)
	if err != nil {
		return err
	}
	field.SetFloat(math.Float64frombits(binary.LittleEndian.Uint64(bs)))
	return nil
}
