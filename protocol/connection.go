package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
)

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
	t := val.Type()
	count := t.NumField()
	for i := 0; i < count; i++ {
		conn.writeField(val, val.Field(i), t.Field(i))
	}
}

//TODO: Implement metadata
func (conn *Conn) writeField(root reflect.Value, field reflect.Value, sField reflect.StructField) {
	if strings.HasPrefix(string(sField.Tag), "if") { //if should always be first if used
		cond := sField.Tag.Get("if")
		var args [3]string
		pos := strings.Index(cond, ",")
		args[0] = cond[:pos]
		cond = cond[pos+1:]
		pos = strings.Index(cond, ",")
		args[1] = cond[:pos]
		args[2] = cond[pos+1:]
		checkField := root.FieldByName(args[0])
		val, _ := strconv.ParseInt(args[2], 10, 64)
		switch args[1] {
		case "!=":
			if checkField.Int() == val {
				return
			}
		case "==":
			if checkField.Int() != val {
				return
			}
		}
	}
	var bs []byte
	switch field.Kind() {
	case reflect.Bool:
		bs = conn.b[:1]
		if field.Bool() {
			bs[0] = 1
		} else {
			bs[0] = 0
		}
	case reflect.Uint8:
		conn.b[0] = byte(field.Uint())
		bs = conn.b[:1]
	case reflect.Int8:
		conn.b[0] = byte(field.Int())
		bs = conn.b[:1]
	case reflect.Int16:
		bs = conn.b[:2]
		binary.BigEndian.PutUint16(bs, uint16(field.Int()))
	case reflect.Uint16:
		bs = conn.b[:2]
		binary.BigEndian.PutUint16(bs, uint16(field.Uint()))
	case reflect.Int32:
		bs = conn.b[:4]
		binary.BigEndian.PutUint32(bs, uint32(field.Int()))
	case reflect.Int64:
		bs = conn.b[:8]
		binary.BigEndian.PutUint64(bs, uint64(field.Int()))
	case reflect.Float32:
		bs = conn.b[:4]
		binary.BigEndian.PutUint32(bs, math.Float32bits(float32(field.Float())))
	case reflect.Float64:
		bs = conn.b[:8]
		binary.BigEndian.PutUint64(bs, math.Float64bits(field.Float()))
	case reflect.String:
		val := []rune(field.String())
		length := len(val)
		bs = make([]byte, length*2+2)
		binary.BigEndian.PutUint16(bs, uint16(length))
		for i, r := range val {
			binary.BigEndian.PutUint16(bs[2+i*2:], uint16(r))
		}
	case reflect.Slice:
		l := field.Len()
		if field.IsNil() {
			l, _ = strconv.Atoi(sField.Tag.Get("nil"))
		}
		switch sField.Tag.Get("ltype") {
		case "int8":
			bs = conn.b[:1]
			bs[0] = byte(l)
		case "int16":
			bs = conn.b[:2]
			binary.BigEndian.PutUint16(bs, uint16(l))
		}
		conn.Out.Write(bs)
		bs = nil
		for i := 0; i < l; i++ {
			conn.writeField(root, field.Index(i), reflect.StructField{}) //StructField shouldn't be need in a slice
		}
	case reflect.Struct:
		bs = nil
		conn.writeInterface(field.Interface())
	default:
		panic(fmt.Errorf("Unknown type: %t for %s", field.Kind(), sField.Name))
	}
	if bs != nil {
		conn.Out.Write(bs)
	}
}
