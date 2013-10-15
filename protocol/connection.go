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

package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	fieldCache.m = make(map[reflect.Type][]field)
	fieldCache.create = make(map[reflect.Type]*sync.WaitGroup)
}

type VarInt int

//Conn has WritePacket and ReadPacket methods that allow
//Go structs to be used in sending and recieving minecraft
//packets.
//
//Supported types
//    uint8
//    int8                 //Java byte
//    uint16
//    int16                //Java short
//    int32                //Java int
//    int64                //Java long
//    float32              //Java float
//    float64              //Java double
//    string               //UTF-16 string with a int16 length prefix
//    structs              //Not struct pointers
//    []type               //Above are the supported types
//    map[byte]interface{} //Encoded as entity metadata
//    VarInt
type Conn struct {
	In  io.Reader
	Out io.Writer

	Deadliner Deadliner

	//Used on the write goroutine
	b [8]byte
	//Used on the read goroutine
	rb [8]byte

	State          State
	ReadDirection  Direction
	WriteDirection Direction
}

type Deadliner interface {
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

//Reads a minecraft packet from conn
func (conn *Conn) ReadPacket() (Packet, error) {
	if conn.Deadliner != nil {
		conn.Deadliner.SetReadDeadline(time.Now().Add(10 * time.Second))
	}

	l, err := readVarInt(conn)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, l)
	_, err = io.ReadFull(conn.In, buf)
	if err != nil {
		return nil, err
	}

	temp := conn.In
	conn.In = bytes.NewReader(buf)

	id, err := readVarInt(conn)
	if err != nil {
		return nil, err
	}
	if int(id) == _MapChunkBulkID { //Its hard to parse normally
		p := MapChunkBulk{}
		binary.Read(conn.In, binary.BigEndian, &p.ChunkCount)
		binary.Read(conn.In, binary.BigEndian, &p.DataLength)
		binary.Read(conn.In, binary.BigEndian, &p.SkyLight)
		p.Data = make([]byte, p.DataLength)
		p.Meta = make([]ChunkMeta, p.ChunkCount)
		conn.In.Read(p.Data)
		binary.Read(conn.In, binary.BigEndian, p.Meta)
		conn.In = temp
		return p, nil
	}

	st := packets[conn.State][conn.ReadDirection]
	if id < 0 || int(id) >= len(st) {
		return nil, fmt.Errorf("Invalid packet %02X", id)
	}
	ty := st[id]
	if ty == nil {
		return nil, fmt.Errorf("Invalid packet %02X", id)
	}

	val := reflect.New(ty).Elem()

	fs := fields(ty)
	for _, f := range fs {
		if f.condition(val) {
			v := val.FieldByIndex(f.sField.Index)
			err := f.read(conn, v)
			if err != nil {
				return nil, err
			}
		}
	}
	conn.In = temp
	return val.Interface().(Packet), nil
}

//Writes the packet to conn
func (conn *Conn) WritePacket(packet Packet) {
	if conn.Deadliner != nil {
		conn.Deadliner.SetWriteDeadline(time.Now().Add(10 * time.Second))
	}

	var buf bytes.Buffer
	temp := conn.Out
	conn.Out = &buf

	val := reflect.ValueOf(packet)
	ty := val.Type()

	id, ok := packetsToID[conn.WriteDirection][ty]
	if !ok {
		panic("Invalid Packet")
	}

	writeVarInt(conn, VarInt(id))

	if id == _MapChunkBulkID { //Its hard to parse normally
		p := packet.(MapChunkBulk)
		binary.Write(conn.Out, binary.BigEndian, &p.ChunkCount)
		binary.Write(conn.Out, binary.BigEndian, &p.DataLength)
		binary.Write(conn.Out, binary.BigEndian, &p.SkyLight)
		binary.Write(conn.Out, binary.BigEndian, p.Data)
		binary.Write(conn.Out, binary.BigEndian, p.Meta)
		conn.Out = temp
		writeVarInt(conn, VarInt(buf.Len()))
		buf.WriteTo(conn.Out)
		return
	}

	fs := fields(ty)
	for _, f := range fs {
		if f.condition(val) {
			v := val.FieldByIndex(f.sField.Index)
			f.write(conn, v)
		}
	}
	conn.Out = temp
	writeVarInt(conn, VarInt(buf.Len()))
	buf.WriteTo(conn.Out)
}

var fieldCache struct {
	sync.RWMutex
	m      map[reflect.Type][]field
	create map[reflect.Type]*sync.WaitGroup
}

//Returns the fields for the type t. This method caches the results making
//later calls cheap.
func fields(t reflect.Type) []field {
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

	fs = compileStruct(t, nil)

	fieldCache.Lock()
	fieldCache.m[t] = fs
	fieldCache.Unlock()
	sy.Done()
	return fs
}

//Loops through the fields of the struct and returns a slice of fields.
//ind is the offset of the struct in the root struct.
func compileStruct(t reflect.Type, ind []int) []field {
	var fs []field = []field{}
	count := t.NumField()
	for i := 0; i < count; i++ {
		f := t.Field(i)
		fs = append(fs, compileField(f, t, ind)...)
	}
	return fs
}

//A field contains the methods needed to read and write the
//field. It also has the condition that the field requires
//to be written.
type field struct {
	sField    reflect.StructField
	condition func(root reflect.Value) bool
	write     encoder
	read      decoder
}

//Returns the field or fields needed to fully write the struct's field
func compileField(sf reflect.StructField, t reflect.Type, ind []int) []field {
	temp := sf.Index[0]
	sf.Index = make([]int, len(ind)+1)
	copy(sf.Index, ind)
	sf.Index[len(ind)] = temp
	f := field{sField: sf}

	cond := sf.Tag.Get("if")
	if len(cond) > 0 {
		var args [3]string
		pos := strings.Index(cond, ",")
		args[0] = cond[:pos]
		cond = cond[pos+1:]
		pos = strings.Index(cond, ",")
		args[1] = cond[:pos]
		args[2] = cond[pos+1:]
		checkField, ok := t.FieldByName(args[0])
		if !ok {
			panic(fmt.Errorf("Unknown field: %s", args[0]))
		}
		in := make([]int, len(checkField.Index)+len(ind))
		copy(in, ind)
		copy(in[len(ind):], checkField.Index)

		valsStr := strings.Split(args[2], "|")
		vals := make([]int64, len(valsStr))
		for i := range vals {
			vals[i], _ = strconv.ParseInt(valsStr[i], 10, 64)
		}

		switch args[1] {
		case "!=":
			f.condition = func(root reflect.Value) bool {
				val := root.FieldByIndex(in).Int()
				for _, v := range vals {
					if v != val {
						return true
					}
				}
				return false
			}
		case "==":
			f.condition = func(root reflect.Value) bool {
				val := root.FieldByIndex(ind).Int()
				for _, v := range vals {
					if v == val {
						return true
					}
				}
				return false
			}
		}
	} else {
		f.condition = condAlways
	}

	switch sf.Type.Kind() {
	case reflect.Bool:
		f.write = encodeBool
		f.read = decodeBool
	case reflect.Int8:
		f.write = encodeInt8
		f.read = decodeInt8
	case reflect.Uint8:
		f.write = encodeUint8
		f.read = decodeUint8
	case reflect.Int16:
		f.write = encodeInt16
		f.read = decodeInt16
	case reflect.Uint16:
		f.write = encodeUint16
		f.read = decodeUint16
	case reflect.Int32:
		f.write = encodeInt32
		f.read = decodeInt32
	case reflect.Int64:
		f.write = encodeInt64
		f.read = decodeInt64
	case reflect.Float32:
		f.write = encodeFloat32
		f.read = decodeFloat32
	case reflect.Float64:
		f.write = encodeFloat64
		f.read = decodeFloat64
	case reflect.String:
		f.write = encodeString
		f.read = decodeString
	case reflect.Slice:
		e := sf.Type.Elem()
		f.write, f.read = getSliceCoders(e, sf)
	case reflect.Map:
		if sf.Tag.Get("metadata") == "true" {
			f.write = encodeMetadata
			f.read = decodeMetadata
		} else {
			panic("Maps NYI")
		}
	case reflect.Struct:
		return compileStruct(sf.Type, sf.Index)
	case reflect.Int: //VarInt
		f.write = encodeVarInt
		f.read = decodeVarInt
	default:
		panic(fmt.Errorf("Unhandled type %s for %s", sf.Type.Kind().String(), sf.Name))
	}
	if f.write == nil {
		panic(fmt.Errorf("Missing write for type %s", sf.Type.Kind()))
	}
	if f.read == nil {
		panic(fmt.Errorf("Missing read for type %s", sf.Type.Kind()))
	}
	return []field{f}
}

//Returns the encoding and decoder methods required to write and read the slice
func getSliceCoders(e reflect.Type, sf reflect.StructField) (encoder, decoder) {
	var write encoder
	var read decoder
	noLoop := false

	nilValue, err := strconv.Atoi(sf.Tag.Get("nil"))
	if err != nil || len(sf.Tag.Get("nil")) == 0 {
		nilValue = 0
	}
	lType := sf.Tag.Get("ltype")

	switch e.Kind() {
	case reflect.Bool:
		write = encodeBool
		read = decodeBool
	case reflect.Uint8:
		write = encodeByteSlice
		read = func(conn *Conn, field reflect.Value) error {
			var l int
			var bs []byte
			switch lType {
			case "int8":
				bs = conn.rb[:1]
				_, err := io.ReadFull(conn.In, bs)
				if err != nil {
					return err
				}
				l = int(int8(bs[0]))
			case "int16":
				bs = conn.rb[:2]
				_, err := io.ReadFull(conn.In, bs)
				if err != nil {
					return err
				}
				l = int(int16(binary.BigEndian.Uint16(bs)))
			case "int32":
				bs = conn.rb[:4]
				_, err := io.ReadFull(conn.In, bs)
				if err != nil {
					return err
				}
				l = int(int32(binary.BigEndian.Uint32(bs)))
			case "varint":
				ll, err := readVarInt(conn)
				if err != nil {
					return err
				}
				l = int(ll)
			case "nil":
			default:
				panic("Unknown length type")
			}
			if l != nilValue {
				b := make([]byte, l)
				_, err := io.ReadFull(conn.In, b)
				if err != nil {
					return err
				}
				field.SetBytes(b)
			}
			return nil
		}
		noLoop = true
	case reflect.Int8:
		write = encodeInt8
		read = decodeInt8
	case reflect.Int16:
		write = encodeInt16
		read = decodeInt16
	case reflect.Uint16:
		write = encodeUint16
		read = decodeUint16
	case reflect.Int32:
		write = encodeInt32
		read = decodeInt32
	case reflect.Int64:
		write = encodeInt64
		read = decodeInt64
	case reflect.Float32:
		write = encodeFloat32
		read = decodeFloat32
	case reflect.Float64:
		write = encodeFloat64
		read = decodeFloat64
	case reflect.String:
		write = encodeString
		read = decodeString
	case reflect.Struct:
		structFields := fields(e)
		write = func(conn *Conn, field reflect.Value) {
			for _, f := range structFields {
				if f.condition(field) {
					v := field.FieldByIndex(f.sField.Index)
					f.write(conn, v)
				}
			}
		}
		read = func(conn *Conn, field reflect.Value) error {
			for _, f := range structFields {
				if f.condition(field) {
					v := field.FieldByIndex(f.sField.Index)
					if err := f.read(conn, v); err != nil {
						return err
					}
				}
			}
			return nil
		}
	default:
		panic("Unknown slice type " + e.Kind().String())
	}

	if !noLoop {
		loopWrite := write
		write = func(conn *Conn, field reflect.Value) {
			l := field.Len()
			for i := 0; i < l; i++ {
				loopWrite(conn, field.Index(i))
			}
		}
		loopRead := read
		read = func(conn *Conn, field reflect.Value) error {
			var l int
			var bs []byte
			switch lType {
			case "int8":
				bs = conn.rb[:1]
				_, err := io.ReadFull(conn.In, bs)
				if err != nil {
					return err
				}
				l = int(int8(bs[0]))
			case "int16":
				bs = conn.rb[:2]
				_, err := io.ReadFull(conn.In, bs)
				if err != nil {
					return err
				}
				l = int(int16(binary.BigEndian.Uint16(bs)))
			case "int32":
				bs = conn.rb[:4]
				_, err := io.ReadFull(conn.In, bs)
				if err != nil {
					return err
				}
				l = int(int32(binary.BigEndian.Uint32(bs)))
			case "varint":
				ll, err := readVarInt(conn)
				if err != nil {
					return err
				}
				l = int(ll)
			case "nil":
			default:
				panic("Unknown length type")
			}
			if l != nilValue {
				slice := reflect.MakeSlice(sf.Type, l, l)
				for i := 0; i < l; i++ {
					if err := loopRead(conn, slice.Index(i)); err != nil {
						return err
					}
				}
				field.Set(slice)
			}
			return nil
		}
	}

	retwrite := func(conn *Conn, field reflect.Value) {
		var l int
		if field.IsNil() {
			l = nilValue
		} else {
			l = field.Len()
		}
		var bs []byte
		switch lType {
		case "int8":
			bs = conn.b[:1]
			bs[0] = byte(l)
		case "int16":
			bs = conn.b[:2]
			binary.BigEndian.PutUint16(bs, uint16(l))
		case "int32":
			bs = conn.b[:4]
			binary.BigEndian.PutUint32(bs, uint32(l))
		case "varint":
			writeVarInt(conn, VarInt(l))
		case "nil":
		default:
			panic("Unknown length type")
		}
		conn.Out.Write(bs)
		if !field.IsNil() {
			write(conn, field)
		}
	}
	return retwrite, read
}

type encoder func(conn *Conn, field reflect.Value)
type decoder func(conn *Conn, field reflect.Value) error

func encodeMetadata(conn *Conn, field reflect.Value) {
	m := field.Interface().(map[byte]interface{})
	index := []byte{0}
	var ty byte = 0 //Type
	var bs []byte
	for i, v := range m {
		manual := false
		switch v := v.(type) {
		case int8:
			ty = 0
			bs = conn.b[:1]
			bs[0] = byte(v)
		case int16:
			ty = 1
			bs = conn.b[:2]
			binary.BigEndian.PutUint16(bs, uint16(v))
		case int32:
			ty = 2
			bs = conn.b[:4]
			binary.BigEndian.PutUint32(bs, uint32(v))
		case float32:
			ty = 3
			bs = conn.b[:4]
			binary.BigEndian.PutUint32(bs, math.Float32bits(v))
		case string:
			manual = true
			ty = 4
			index[0] = (i & 0x1F) | (ty << 5)
			conn.Out.Write(index)
			writeVarInt(conn, VarInt(len(v)))
			conn.Out.Write([]byte(v))
		case Slot:
			manual = true
			ty = 5
			index[0] = (i & 0x1F) | (ty << 5)
			conn.Out.Write(index)
			val := reflect.ValueOf(v)
			fs := fields(val.Type())
			for _, f := range fs {
				if f.condition(val) {
					v := val.FieldByIndex(f.sField.Index)
					f.write(conn, v)
				}
			}
		}
		if !manual {
			index[0] = (i & 0x1F) | (ty << 5)
			conn.Out.Write(index)
			conn.Out.Write(bs)
		}
	}
	conn.Out.Write([]byte{0x7F})
}

var slotType = reflect.TypeOf((*Slot)(nil)).Elem()

func decodeMetadata(conn *Conn, field reflect.Value) error {
	index := make([]byte, 1)
	_, err := io.ReadFull(conn.In, index)
	if err != nil {
		return err
	}

	m := map[byte]interface{}{}
	for index[0] != 0x7F {
		i := index[0] & 0x1F
		ty := index[0] >> 5
		var v interface{}
		var bs []byte
		switch ty {
		case 0:
			bs = conn.rb[:1]
			_, err := io.ReadFull(conn.In, bs)
			if err != nil {
				return err
			}
			v = int8(bs[0])

		case 1:
			bs = conn.rb[:2]
			_, err := io.ReadFull(conn.In, bs)
			if err != nil {
				return err
			}
			v = int16(binary.BigEndian.Uint16(bs))

		case 2:
			bs = conn.rb[:4]
			_, err := io.ReadFull(conn.In, bs)
			if err != nil {
				return err
			}
			v = int32(binary.BigEndian.Uint32(bs))

		case 3:
			bs = conn.rb[:4]
			_, err := io.ReadFull(conn.In, bs)
			if err != nil {
				return err
			}
			v = math.Float32frombits(binary.BigEndian.Uint32(bs))

		case 4:
			l, err := readVarInt(conn)
			if err != nil {
				return err
			}
			b := make([]byte, l)
			_, err = conn.In.Read(b)
			v = string(string(b))
		case 5:
			val := reflect.New(slotType).Elem()

			fs := fields(val.Type())
			for _, f := range fs {
				if f.condition(val) {
					v := val.FieldByIndex(f.sField.Index)

					if err := f.read(conn, v); err != nil {
						return err
					}
				}
			}
			v = val.Interface()
		}
		m[i] = v
		_, err := io.ReadFull(conn.In, index)
		if err != nil {
			return err
		}
	}
	field.Set(reflect.ValueOf(m))
	return nil
}

func encodeByteSlice(conn *Conn, field reflect.Value) {
	conn.Out.Write(field.Bytes())
}

func encodeString(conn *Conn, field reflect.Value) {
	writeVarInt(conn, VarInt(field.Len()))
	conn.Out.Write([]byte(field.String()))
}

func decodeString(conn *Conn, field reflect.Value) error {
	l, err := readVarInt(conn)
	if err != nil {
		return err
	}
	b := make([]byte, l)
	_, err = conn.In.Read(b)
	field.SetString(string(b))
	return err
}

func encodeBool(conn *Conn, field reflect.Value) {
	bs := conn.b[:1]
	if field.Bool() {
		bs[0] = 1
	} else {
		bs[0] = 0
	}
	conn.Out.Write(bs)
}

func decodeBool(conn *Conn, field reflect.Value) error {
	bs := conn.rb[:1]
	_, err := io.ReadFull(conn.In, bs)
	if err != nil {
		return err
	}
	if bs[0] == 1 {
		field.SetBool(true)
	}
	return nil
}

func encodeInt8(conn *Conn, field reflect.Value) {
	bs := conn.b[:1]
	bs[0] = byte(field.Int())
	conn.Out.Write(bs)
}

func decodeInt8(conn *Conn, field reflect.Value) error {
	bs := conn.rb[:1]
	_, err := io.ReadFull(conn.In, bs)
	if err != nil {
		return err
	}
	field.SetInt(int64(bs[0]))
	return nil
}

func encodeUint8(conn *Conn, field reflect.Value) {
	bs := conn.b[:1]
	bs[0] = byte(field.Uint())
	conn.Out.Write(bs)
}

func decodeUint8(conn *Conn, field reflect.Value) error {
	bs := conn.rb[:1]
	_, err := io.ReadFull(conn.In, bs)
	if err != nil {
		return err
	}
	field.SetUint(uint64(bs[0]))
	return nil
}

func encodeInt16(conn *Conn, field reflect.Value) {
	bs := conn.b[:2]
	binary.BigEndian.PutUint16(bs, uint16(field.Int()))
	conn.Out.Write(bs)
}

func decodeInt16(conn *Conn, field reflect.Value) error {
	bs := conn.rb[:2]
	_, err := io.ReadFull(conn.In, bs)
	if err != nil {
		return err
	}
	field.SetInt(int64(binary.BigEndian.Uint16(bs)))
	return nil
}

func encodeUint16(conn *Conn, field reflect.Value) {
	bs := conn.b[:2]
	binary.BigEndian.PutUint16(bs, uint16(field.Uint()))
	conn.Out.Write(bs)
}

func decodeUint16(conn *Conn, field reflect.Value) error {
	bs := conn.rb[:2]
	_, err := io.ReadFull(conn.In, bs)
	if err != nil {
		return err
	}
	field.SetUint(uint64(binary.BigEndian.Uint16(bs)))
	return nil
}

func encodeInt32(conn *Conn, field reflect.Value) {
	bs := conn.b[:4]
	binary.BigEndian.PutUint32(bs, uint32(field.Int()))
	conn.Out.Write(bs)
}

func decodeInt32(conn *Conn, field reflect.Value) error {
	bs := conn.rb[:4]
	_, err := io.ReadFull(conn.In, bs)
	if err != nil {
		return err
	}
	field.SetInt(int64(binary.BigEndian.Uint32(bs)))
	return nil
}

func encodeInt64(conn *Conn, field reflect.Value) {
	bs := conn.b[:8]
	binary.BigEndian.PutUint64(bs, uint64(field.Int()))
	conn.Out.Write(bs)
}

func decodeInt64(conn *Conn, field reflect.Value) error {
	bs := conn.rb[:8]
	_, err := io.ReadFull(conn.In, bs)
	if err != nil {
		return err
	}
	field.SetInt(int64(binary.BigEndian.Uint64(bs)))
	return nil
}

func encodeFloat32(conn *Conn, field reflect.Value) {
	bs := conn.b[:4]
	binary.BigEndian.PutUint32(bs, math.Float32bits(float32(field.Float())))
	conn.Out.Write(bs)
}

func decodeFloat32(conn *Conn, field reflect.Value) error {
	bs := conn.rb[:4]
	_, err := io.ReadFull(conn.In, bs)
	if err != nil {
		return err
	}
	field.SetFloat(float64(math.Float32frombits(binary.BigEndian.Uint32(bs))))
	return nil
}

func encodeFloat64(conn *Conn, field reflect.Value) {
	bs := conn.b[:8]
	binary.BigEndian.PutUint64(bs, math.Float64bits(field.Float()))
	conn.Out.Write(bs)
}

func decodeFloat64(conn *Conn, field reflect.Value) error {
	bs := conn.rb[:8]
	_, err := io.ReadFull(conn.In, bs)
	if err != nil {
		return err
	}
	field.SetFloat(math.Float64frombits(binary.BigEndian.Uint64(bs)))
	return nil
}

type byteReader struct {
	io.Reader
	buf [1]byte
}

func (b byteReader) ReadByte() (byte, error) {
	bs := b.buf[:]
	_, err := b.Read(bs)
	return bs[0], err
}

func readVarInt(conn *Conn) (VarInt, error) {
	x, err := binary.ReadUvarint(byteReader{conn.In, [1]byte{}})
	return VarInt(int32(uint32(x))), err
}

//Modified from encoding/binary
func writeVarInt(conn *Conn, i VarInt) {
	bs := conn.b[:]
	n := binary.PutUvarint(bs, uint64(uint32(i)))
	conn.Out.Write(bs[:n])
}

func encodeVarInt(conn *Conn, field reflect.Value) {
	writeVarInt(conn, VarInt(field.Int()))
}

func decodeVarInt(conn *Conn, field reflect.Value) error {
	i, err := readVarInt(conn)
	field.SetInt(int64(i))
	return err
}

func condAlways(root reflect.Value) bool { return true }
