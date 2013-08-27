package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

func init() {
	fieldCache.m = make(map[reflect.Type][]field)
	fieldCache.create = make(map[reflect.Type]*sync.WaitGroup)
}

type Conn struct {
	In  io.Reader
	Out io.Writer

	b [8]byte
}

func (conn *Conn) WritePacket(packet Packet) {
	conn.Out.Write([]byte{packet.ID()})

	conn.writeInterface(packet)
}

func (conn *Conn) writeInterface(data interface{}) {
	val := reflect.ValueOf(data)
	fs := fields(val.Type())
	for _, f := range fs {
		if f.condition(val) {
			v := val.FieldByIndex(f.sField.Index)
			f.write(conn, v)
		}
	}
}

var fieldCache struct {
	sync.RWMutex
	m      map[reflect.Type][]field
	create map[reflect.Type]*sync.WaitGroup
}

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

	sy.Done()
	fieldCache.Lock()
	fieldCache.m[t] = fs
	fieldCache.Unlock()
	return fs
}

func compileStruct(t reflect.Type, ind []int) []field {
	var fs []field
	count := t.NumField()
	for i := 0; i < count; i++ {
		f := t.Field(i)
		fs = append(fs, compileField(f, t, ind)...)
	}
	return fs
}

type field struct {
	sField    reflect.StructField
	condition func(root reflect.Value) bool
	write     func(conn *Conn, field reflect.Value)
}

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
		ind := checkField.Index
		val, _ := strconv.ParseInt(args[2], 10, 64)
		switch args[1] {
		case "!=":
			f.condition = func(root reflect.Value) bool {
				return root.FieldByIndex(ind).Int() != val
			}
		case "==":
			f.condition = func(root reflect.Value) bool {
				return root.FieldByIndex(ind).Int() == val
			}
		}
	} else {
		f.condition = condAlways
	}

	switch sf.Type.Kind() {
	case reflect.Int8:
		f.write = encodeInt8
	case reflect.Uint8:
		f.write = encodeUint8
	case reflect.Int16:
		f.write = encodeInt16
	case reflect.Uint16:
		f.write = encodeUint16
	case reflect.Int32:
		f.write = encodeInt32
	case reflect.Int64:
		f.write = encodeInt64
	case reflect.Float32:
		f.write = encodeFloat32
	case reflect.Float64:
		f.write = encodeFloat64
	case reflect.String:
		f.write = encodeString
	case reflect.Slice:
		e := sf.Type.Elem()
		var write func(conn *Conn, field reflect.Value)
		switch e.Kind() {
		case reflect.Uint8:
			write = encodeByteSlice
		default:
			panic("Unknown slice type")
		}
		nilValue, _ := strconv.Atoi(sf.Tag.Get("nil"))
		lType := sf.Tag.Get("ltype")
		f.write = func(conn *Conn, field reflect.Value) {
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
			}
			conn.Out.Write(bs)
			if !field.IsNil() {
				write(conn, field)
			}
		}
	case reflect.Struct:
		return compileStruct(sf.Type, sf.Index)
	default:
		panic(fmt.Errorf("Unhandled type %s", sf.Type.Kind().String()))
	}
	return []field{f}
}

func encodeByteSlice(conn *Conn, field reflect.Value) {
	conn.Out.Write(field.Bytes())
}

func encodeString(conn *Conn, field reflect.Value) {
	val := []rune(field.String())
	length := len(val)
	bs := make([]byte, length*2+2)
	binary.BigEndian.PutUint16(bs, uint16(length))
	for i, r := range val {
		binary.BigEndian.PutUint16(bs[2+i*2:], uint16(r))
	}
	conn.Out.Write(bs)
}

func encodeInt8(conn *Conn, field reflect.Value) {
	bs := conn.b[:1]
	bs[0] = byte(field.Int())
	conn.Out.Write(bs)
}

func encodeUint8(conn *Conn, field reflect.Value) {
	bs := conn.b[:1]
	bs[0] = byte(field.Uint())
	conn.Out.Write(bs)
}

func encodeInt16(conn *Conn, field reflect.Value) {
	bs := conn.b[:2]
	binary.BigEndian.PutUint16(bs, uint16(field.Int()))
	conn.Out.Write(bs)
}

func encodeUint16(conn *Conn, field reflect.Value) {
	bs := conn.b[:2]
	binary.BigEndian.PutUint16(bs, uint16(field.Uint()))
	conn.Out.Write(bs)
}

func encodeInt32(conn *Conn, field reflect.Value) {
	bs := conn.b[:4]
	binary.BigEndian.PutUint32(bs, uint32(field.Int()))
	conn.Out.Write(bs)
}

func encodeInt64(conn *Conn, field reflect.Value) {
	bs := conn.b[:8]
	binary.BigEndian.PutUint64(bs, uint64(field.Int()))
	conn.Out.Write(bs)
}

func encodeFloat32(conn *Conn, field reflect.Value) {
	bs := conn.b[:4]
	binary.BigEndian.PutUint32(bs, math.Float32bits(float32(field.Float())))
	conn.Out.Write(bs)
}

func encodeFloat64(conn *Conn, field reflect.Value) {
	bs := conn.b[:8]
	binary.BigEndian.PutUint64(bs, math.Float64bits(field.Float()))
	conn.Out.Write(bs)
}

func condAlways(root reflect.Value) bool { return true }
