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

package nbt

import (
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

//Reads in the nbt format from r into v. Panics if
//v is not a pointer to a struct.
func Read(r io.Reader, v interface{}) (err error) {
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Ptr {
		panic("format/nbt: v is not a pointer")
	}
	val := ptr.Elem()
	if val.Kind() != reflect.Struct {
		panic("format/nbt: v isn't a pointer to a struct")
	}
	//Catch any errors to prevent crashing
	defer func() {
		/*if e := recover(); e != nil {
			if er, ok := e.(error); ok {
				err = er
			} else {
				err = fmt.Errorf("format/nbt: %s", e)
			}
			return
		}*/
	}()

	fs := fields(val.Type())
	de := &msgDecoder{}
	_, t, err := readPrefix(r, de)
	if err != nil {
		return err
	}
	if t != 10 {
		return errors.New("format/nbt: Not an nbt file")
	}
	return read(fullReader{r}, de, fs, val)
}

//Reads a struct map from r into val
func read(r io.Reader, de *msgDecoder, fs map[string]interface{}, val reflect.Value) error {
	name, t, err := readPrefix(r, de)
	for ; t != 0; name, t, err = readPrefix(r, de) {
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
			if t != f.requiredType {
				return ErrorIncorrectType
			}
			val := val.Field(f.sField)
			err := f.read(r, de, val)
			if err != nil {
				return err
			}
		} else {
			if t != 10 {
				return ErrorIncorrectType
			}
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

func readPrefix(r io.Reader, de *msgDecoder) (name string, t byte, err error) {
	bs := de.b[:1]
	_, err = r.Read(bs)
	if err != nil {
		return
	}
	t = bs[0]
	if t != 0 {
		name, err = readString(r, de)
	}
	return
}

//Writes v into w in the nbt format.
func Write(w io.Writer, v interface{}, name string) (err error) {
	defer func() {
		/*if e := recover(); e != nil {
			if er, ok := e.(error); ok {
				err = er
			} else {
				err = fmt.Errorf("format/nbt: %s", e)
			}
			return
		}*/
	}()
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	fs := fields(val.Type())
	en := &msgEncoder{}

	//Write name
	bs := en.b[:3]
	bs[0] = 10
	binary.BigEndian.PutUint16(bs[1:], uint16(len(name)))
	_, err = w.Write(bs)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(name))
	if err != nil {
		return err
	}

	return write(w, en, fs, val)
}

//Writes val into w using the struct map fs
func write(w io.Writer, en *msgEncoder, fs map[string]interface{}, val reflect.Value) error {
	for _, i := range fs {
		if f, ok := i.(field); ok {
			writePrefix(en, w, f.name, f.requiredType)
			v := val.Field(f.sField)
			err := f.write(w, en, v)
			if err != nil {
				return err
			}
		} else {
			fs := i.(fieldStruct)
			writePrefix(en, w, fs.name, 10)
			v := val.Field(fs.sField)
			err := write(w, en, fs.m, v)
			if err != nil {
				return err
			}
		}
	}
	bs := en.b[:1]
	bs[0] = 0
	_, err := w.Write(bs)
	return err
}

func writePrefix(en *msgEncoder, w io.Writer, name []byte, t byte) error {
	bs := en.b[:3]
	bs[0] = t
	binary.BigEndian.PutUint16(bs[1:], uint16(len(name)))
	_, err := w.Write(bs)
	if err != nil {
		return err
	}
	_, err = w.Write(name)
	if err != nil {
		return err
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
		var name string
		if !f.Anonymous {
			name = f.Name
			if tName := f.Tag.Get("nbt"); len(tName) > 0 {
				name = tName
			}
			if name == "ignore" || f.Tag.Get("ignore") == "true" {
				continue
			}
		} else {
			name = f.Type.Name()
			if tName := f.Tag.Get("nbt"); len(tName) > 0 {
				name = tName
			}
			if f.Tag.Get("ignore") == "true" {
				continue
			}
		}
		fs[name] = compileField(f, name)
	}
	return fs
}

//A field contains the methods needed to read and write the
//field.
type field struct {
	sField       int
	name         []byte
	write        encoder
	read         decoder
	requiredType byte
}

//A special type of field which contains a struct
type fieldStruct struct {
	sField int
	name   []byte
	m      map[string]interface{}
}

type encoder func(w io.Writer, en *msgEncoder, field reflect.Value) error
type decoder func(r io.Reader, de *msgDecoder, field reflect.Value) error

//Returns the field or fields needed to fully write the struct's field
func compileField(sf reflect.StructField, name string) interface{} {
	f := field{sField: sf.Index[0]}

	f.name = []byte(name)

	switch sf.Type.Kind() {
	case reflect.Struct:
		return fieldStruct{f.sField, f.name, compileStruct(sf.Type)}
	case reflect.Bool:
		f.write = encodeBool
		f.read = decodeBool
		f.requiredType = 1
	case reflect.Int8:
		f.write = encodeInt8
		f.read = decodeInt8
		f.requiredType = 1
	case reflect.Int16:
		f.write = encodeInt16
		f.read = decodeInt16
		f.requiredType = 2
	case reflect.Int32:
		f.write = encodeInt32
		f.read = decodeInt32
		f.requiredType = 3
	case reflect.Int64:
		f.write = encodeInt64
		f.read = decodeInt64
		f.requiredType = 4
	case reflect.String:
		f.write = encodeString
		f.read = decodeString
		f.requiredType = 8
	case reflect.Map:
		f.requiredType = 10
		elem := sf.Type.Elem()
		var elemField interface{}
		name := "map:" + sf.Name
		if elem.Kind() != reflect.Interface {
			elemField = compileField(reflect.StructField{Type: elem, Index: []int{0}}, name)
		}
		f.write = func(w io.Writer, en *msgEncoder, fi reflect.Value) error {
			keys := fi.MapKeys()
			for _, key := range keys {
				if f, ok := elemField.(field); ok {
					v := fi.MapIndex(key)
					writePrefix(en, w, []byte(key.String()), f.requiredType)
					err := f.write(w, en, v)
					if err != nil {
						return err
					}
				} else {
					if elemField == nil {
						v := fi.MapIndex(key).Elem()
						temp := compileField(reflect.StructField{Type: v.Type(), Index: []int{0}}, "")
						if f, ok := temp.(field); ok {
							writePrefix(en, w, []byte(key.String()), f.requiredType)
							err := f.write(w, en, v)
							if err != nil {
								return err
							}
						} else {
							writePrefix(en, w, []byte(key.String()), 10)
							fs := temp.(fieldStruct)
							err := write(w, en, fs.m, v)
							if err != nil {
								return err
							}
						}
					} else {
						writePrefix(en, w, []byte(key.String()), 10)
						fs := elemField.(fieldStruct)
						v := fi.MapIndex(key)
						err := write(w, en, fs.m, v)
						if err != nil {
							return err
						}
					}

				}
			}
			bs := en.b[:1]
			bs[0] = 0
			_, err := w.Write(bs)
			return err
		}
		f.read = func(r io.Reader, de *msgDecoder, fi reflect.Value) error {

			ma := reflect.MakeMap(sf.Type)

			name, t, err := readPrefix(r, de)
			for ; t != 0; name, t, err = readPrefix(r, de) {
				if err != nil {
					return err
				}
				keyVal := reflect.ValueOf(name)

				var val reflect.Value
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
		f.requiredType = 9
		elem := sf.Type.Elem()
		switch elem.Kind() {
		case reflect.Uint8: //Short-cut for byte arrays
			f.requiredType = 7
			f.write = func(w io.Writer, en *msgEncoder, fi reflect.Value) error {
				l := fi.Len()
				bs := en.b[:4]
				binary.BigEndian.PutUint32(bs, uint32(l))
				_, err := w.Write(bs)
				if err != nil {
					return err
				}
				_, err = w.Write(fi.Bytes())
				return err
			}
			f.read = func(r io.Reader, de *msgDecoder, fi reflect.Value) error {
				bs := de.b[:4]
				_, err := r.Read(bs)
				if err != nil {
					return err
				}
				l := binary.BigEndian.Uint32(bs)
				out := make([]byte, l)
				_, err = r.Read(out)
				if err != nil {
					return nil
				}
				fi.SetBytes(out)
				return nil
			}
		case reflect.Int32: //Short-cut for int32 arrays
			f.requiredType = 11
			f.write = func(w io.Writer, en *msgEncoder, fi reflect.Value) error {
				l := fi.Len()
				bs := en.b[:4]
				binary.BigEndian.PutUint32(bs, uint32(l))
				_, err := w.Write(bs)
				if err != nil {
					return err
				}
				data := fi.Interface().([]int32)
				for i := range data {
					binary.BigEndian.PutUint32(bs, uint32(data[i]))
					_, err := w.Write(bs)
					if err != nil {
						return err
					}
				}
				return err
			}
			f.read = func(r io.Reader, de *msgDecoder, fi reflect.Value) error {
				bs := de.b[:4]
				_, err := r.Read(bs)
				if err != nil {
					return err
				}
				l := binary.BigEndian.Uint32(bs)
				out := make([]int32, l)
				for i := range out {
					_, err := r.Read(bs)
					if err != nil {
						return err
					}
					out[i] = int32(binary.BigEndian.Uint32(bs))
				}
				fi.Set(reflect.ValueOf(out))
				return nil
			}
		default:
			name := "slice:" + sf.Name
			elemField := compileField(reflect.StructField{Type: elem, Index: []int{0}}, name)
			f.write = func(w io.Writer, en *msgEncoder, fi reflect.Value) error {
				l := fi.Len()
				bs := en.b[:5]
				binary.BigEndian.PutUint32(bs[1:], uint32(l))
				if f, ok := elemField.(field); ok {
					bs[0] = f.requiredType
				} else {
					bs[0] = 10
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
				bs := de.b[:5]
				_, err := r.Read(bs)
				if err != nil {
					return err
				}
				if f, ok := elemField.(field); ok {
					if bs[0] != f.requiredType {
						return ErrorIncorrectType
					}
				} else {
					if bs[0] != 10 {
						return ErrorIncorrectType
					}
				}
				l := int(binary.BigEndian.Uint32(bs[1:]))
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
		f.requiredType = 5
		f.write = encodeFloat32
		f.read = decodeFloat32
	case reflect.Float64:
		f.requiredType = 6
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

var ErrorIncorrectType = errors.New("format/nbt: Incorrect type")

func encodeBool(w io.Writer, en *msgEncoder, field reflect.Value) error {
	bs := en.b[:1]
	if field.Bool() {
		bs[0] = 1
	} else {
		bs[0] = 0
	}
	_, err := w.Write(bs)
	return err
}

func decodeBool(r io.Reader, de *msgDecoder, field reflect.Value) error {
	bs := de.b[:1]
	_, err := r.Read(bs)
	field.SetBool(bs[0] != 0x00)
	return err
}

func encodeInt8(w io.Writer, en *msgEncoder, field reflect.Value) error {
	val := int8(field.Int())
	bs := en.b[:1]
	bs[0] = byte(val)
	_, err := w.Write(bs)
	return err
}

func decodeInt8(r io.Reader, de *msgDecoder, field reflect.Value) error {
	bs := de.b[:1]
	_, err := r.Read(bs)
	if err != nil {
		return err
	}
	field.SetInt(int64(int8(bs[0])))
	return nil
}

func encodeInt16(w io.Writer, en *msgEncoder, field reflect.Value) error {
	val := int16(field.Int())
	bs := en.b[:2]
	binary.BigEndian.PutUint16(bs, uint16(val))
	_, err := w.Write(bs)
	return err
}

func decodeInt16(r io.Reader, de *msgDecoder, field reflect.Value) error {
	bs := de.b[:2]
	_, err := r.Read(bs)
	if err != nil {
		return err
	}
	field.SetInt(int64(int16(binary.BigEndian.Uint16(bs))))
	return nil
}

func encodeInt32(w io.Writer, en *msgEncoder, field reflect.Value) error {
	val := int32(field.Int())
	bs := en.b[:4]
	binary.BigEndian.PutUint32(bs, uint32(val))
	_, err := w.Write(bs)
	return err
}

func decodeInt32(r io.Reader, de *msgDecoder, field reflect.Value) error {
	bs := de.b[:4]
	_, err := r.Read(bs)
	if err != nil {
		return err
	}
	field.SetInt(int64(int32(binary.BigEndian.Uint32(bs))))
	return nil
}

func encodeInt64(w io.Writer, en *msgEncoder, field reflect.Value) error {
	val := field.Int()
	bs := en.b[:8]
	binary.BigEndian.PutUint64(bs, uint64(val))
	_, err := w.Write(bs)
	return err
}

func decodeInt64(r io.Reader, de *msgDecoder, field reflect.Value) error {
	bs := de.b[:8]
	_, err := r.Read(bs)
	if err != nil {
		return err
	}
	field.SetInt(int64(binary.BigEndian.Uint64(bs)))
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
	bs := en.b[:2]
	binary.BigEndian.PutUint16(bs, uint16(l))
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
	bs := de.b[:2]
	_, err := r.Read(bs)
	if err != nil {
		return "", err
	}
	l := binary.BigEndian.Uint16(bs)
	b := make([]byte, l)
	_, err = r.Read(b)
	return string(b), err
}

func encodeFloat32(w io.Writer, en *msgEncoder, field reflect.Value) error {
	bs := en.b[:4]
	binary.BigEndian.PutUint32(bs, math.Float32bits(float32(field.Float())))
	_, err := w.Write(bs)
	if err != nil {
		return err
	}
	return nil
}

func decodeFloat32(r io.Reader, de *msgDecoder, field reflect.Value) error {
	bs := de.b[:4]
	_, err := r.Read(bs)
	if err != nil {
		return err
	}
	field.SetFloat(float64(math.Float32frombits(binary.BigEndian.Uint32(bs))))
	return nil
}

func encodeFloat64(w io.Writer, en *msgEncoder, field reflect.Value) error {
	bs := en.b[:8]
	binary.BigEndian.PutUint64(bs, math.Float64bits(field.Float()))
	_, err := w.Write(bs)
	if err != nil {
		return err
	}
	return nil
}

func decodeFloat64(r io.Reader, de *msgDecoder, field reflect.Value) error {
	bs := de.b[:8]
	_, err := r.Read(bs)
	if err != nil {
		return err
	}
	field.SetFloat(math.Float64frombits(binary.BigEndian.Uint64(bs)))
	return nil
}
