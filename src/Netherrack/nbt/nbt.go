package nbt

import (
	"encoding/binary"
	"io"
	"math"
)

func ParseCompound(data Reader) *Compound {
	typeID := data.ReadUByte()
	if typeID != 10 {
		return nil
	}
	return parseCompound(data, true)
}

func parseCompound(data Reader, hasName bool) *Compound {
	compound := &Compound{}
	if hasName {
		compound.Name = data.ReadString()
	}
	compound.Tags = make(map[string]WriterTo)
compLoop:
	for {
		typeID := data.ReadUByte()
		switch typeID {
		case 1: //Byte
			bTag := Byte{}
			bTag.Name = data.ReadString()
			bTag.Value = data.ReadByte()
			compound.Tags[bTag.Name] = bTag
		case 2: //Short
			sTag := Short{}
			sTag.Name = data.ReadString()
			sTag.Value = data.ReadShort()
			compound.Tags[sTag.Name] = sTag
		case 3: //Int
			iTag := Int{}
			iTag.Name = data.ReadString()
			iTag.Value = data.ReadInt()
			compound.Tags[iTag.Name] = iTag
		case 4: //Long
			lTag := Long{}
			lTag.Name = data.ReadString()
			lTag.Value = data.ReadLong()
			compound.Tags[lTag.Name] = lTag
		case 5: //Float
			fTag := Float{}
			fTag.Name = data.ReadString()
			fTag.Value = data.ReadFloat()
			compound.Tags[fTag.Name] = fTag
		case 6: //Double
			dTag := Double{}
			dTag.Name = data.ReadString()
			dTag.Value = data.ReadDouble()
			compound.Tags[dTag.Name] = dTag
		case 7: //Byte Array
			bArray := ByteArray{}
			bArray.Name = data.ReadString()
			bData := make([]byte, data.ReadInt())
			data.R.Read(bData)
			bArray.Values = make([]int8, len(bData))
			for i := 0; i < len(bData); i++ {
				bArray.Values[i] = int8(bData[i])
			}
			compound.Tags[bArray.Name] = bArray
		case 8: //String
			stringTag := String{}
			stringTag.Name = data.ReadString()
			stringTag.Value = data.ReadString()
			compound.Tags[stringTag.Name] = stringTag
		case 9: //List
			list := parseList(data, true)
			compound.Tags[list.Name] = *list
		case 10: //Compound
			com := parseCompound(data, true)
			compound.Tags[com.Name] = *com
		case 11: //Int Array
			iArray := IntArray{}
			iArray.Name = data.ReadString()
			bData := make([]byte, data.ReadInt()*4)
			data.R.Read(bData)
			iArray.Values = make([]int32, len(bData)/4)
			for i := 0; i < len(bData); i++ {
				iArray.Values[i] = int32(bData[i])
			}
			compound.Tags[iArray.Name] = iArray
		case 0: //End
			break compLoop
		}
	}
	return compound
}

func parseList(data Reader, hasName bool) *List {
	list := &List{}
	if hasName {
		list.Name = data.ReadString()
	}
	list.Type = data.ReadUByte()
	count := int(data.ReadInt())
	list.Tags = make([]WriterTo, count)
	for i := 0; i < count; i++ {
		switch list.Type {
		case 1: //Byte
			bTag := Byte{}
			bTag.Value = data.ReadByte()
			list.Tags[i] = bTag
		case 2: //Short
			sTag := Short{}
			sTag.Value = data.ReadShort()
			list.Tags[i] = sTag
		case 3: //Int
			iTag := Int{}
			iTag.Value = data.ReadInt()
			list.Tags[i] = iTag
		case 4: //Long
			lTag := Long{}
			lTag.Value = data.ReadLong()
			list.Tags[i] = lTag
		case 5: //Float
			fTag := Float{}
			fTag.Value = data.ReadFloat()
			list.Tags[i] = fTag
		case 6: //Double
			dTag := Double{}
			dTag.Value = data.ReadDouble()
			list.Tags[i] = dTag
		case 7: //Byte Array
			bArray := ByteArray{}
			bData := make([]byte, data.ReadInt())
			data.R.Read(bData)
			bArray.Values = make([]int8, len(bData))
			for i := 0; i < len(bData); i++ {
				bArray.Values[i] = int8(bData[i])
			}
			list.Tags[i] = bArray
		case 8: //String
			stringTag := String{}
			stringTag.Value = data.ReadString()
			list.Tags[i] = stringTag
		case 9: //List
			list := parseList(data, false)
			list.Tags[i] = *list
		case 10: //Compound
			com := parseCompound(data, false)
			list.Tags[i] = *com
		case 11: //Int Array
			iArray := IntArray{}
			bData := make([]byte, data.ReadInt()*4)
			data.R.Read(bData)
			iArray.Values = make([]int32, len(bData)/4)
			for i := 0; i < len(bData); i++ {
				iArray.Values[i] = int32(bData[i])
			}
			list.Tags[i] = iArray
		}
	}
	return list
}

func (com Compound) WriteTo(w io.Writer, hasName bool) {
	var data []byte
	if hasName {
		data = make([]byte, 1+2+len(com.Name))
		data[0] = 10
		binary.BigEndian.PutUint16(data[1:3], uint16(len(com.Name)))
		copy(data[3:], []byte(com.Name))
	} else {
		data = make([]byte, 0)
	}
	w.Write(data)
	for _, tag := range com.Tags {
		b := tag.(WriterTo)
		b.WriteTo(w, true)
	}
	w.Write([]byte{0})
}

func (nbyte Byte) WriteTo(w io.Writer, hasName bool) {
	var data []byte
	if hasName {
		data = make([]byte, 1+2+len(nbyte.Name)+1)
		data[0] = 1
		binary.BigEndian.PutUint16(data[1:3], uint16(len(nbyte.Name)))
		copy(data[3:], []byte(nbyte.Name))
	} else {
		data = make([]byte, 1)
	}
	data[len(data)-1] = byte(nbyte.Value)
	w.Write(data)
}

func (nshort Short) WriteTo(w io.Writer, hasName bool) {
	var data []byte
	if hasName {
		data = make([]byte, 1+2+len(nshort.Name)+2)
		data[0] = 2
		binary.BigEndian.PutUint16(data[1:3], uint16(len(nshort.Name)))
		copy(data[3:], []byte(nshort.Name))
	} else {
		data = make([]byte, 2)
	}
	binary.BigEndian.PutUint16(data[len(data)-2:], uint16(nshort.Value))
	w.Write(data)
}

func (nint Int) WriteTo(w io.Writer, hasName bool) {
	var data []byte
	if hasName {
		data = make([]byte, 1+2+len(nint.Name)+4)
		data[0] = 3
		binary.BigEndian.PutUint16(data[1:3], uint16(len(nint.Name)))
		copy(data[3:], []byte(nint.Name))
	} else {
		data = make([]byte, 4)
	}
	binary.BigEndian.PutUint32(data[len(data)-4:], uint32(nint.Value))
	w.Write(data)
}

func (nlong Long) WriteTo(w io.Writer, hasName bool) {
	var data []byte
	if hasName {
		data = make([]byte, 1+2+len(nlong.Name)+8)
		data[0] = 4
		binary.BigEndian.PutUint16(data[1:3], uint16(len(nlong.Name)))
		copy(data[3:], []byte(nlong.Name))
	} else {
		data = make([]byte, 8)
	}
	binary.BigEndian.PutUint64(data[len(data)-8:], uint64(nlong.Value))
	w.Write(data)
}

func (nfloat Float) WriteTo(w io.Writer, hasName bool) {
	var data []byte
	if hasName {
		data = make([]byte, 1+2+len(nfloat.Name)+4)
		data[0] = 5
		binary.BigEndian.PutUint16(data[1:3], uint16(len(nfloat.Name)))
		copy(data[3:], []byte(nfloat.Name))
	} else {
		data = make([]byte, 4)
	}
	binary.BigEndian.PutUint32(data[len(data)-4:], uint32(math.Float32bits(nfloat.Value)))
	w.Write(data)
}

func (ndouble Double) WriteTo(w io.Writer, hasName bool) {
	var data []byte
	if hasName {
		data = make([]byte, 1+2+len(ndouble.Name)+8)
		data[0] = 6
		binary.BigEndian.PutUint16(data[1:3], uint16(len(ndouble.Name)))
		copy(data[3:], []byte(ndouble.Name))
	} else {
		data = make([]byte, 8)
	}
	binary.BigEndian.PutUint64(data[len(data)-8:], uint64(math.Float64bits(ndouble.Value)))
	w.Write(data)
}

func (nbarray ByteArray) WriteTo(w io.Writer, hasName bool) {
	var data []byte
	if hasName {
		data = make([]byte, 1+2+len(nbarray.Name)+4+len(nbarray.Values))
		data[0] = 7
		binary.BigEndian.PutUint16(data[1:3], uint16(len(nbarray.Name)))
		copy(data[3:], []byte(nbarray.Name))
	} else {
		data = make([]byte, len(nbarray.Values)+4)
	}
	binary.BigEndian.PutUint32(data[len(data)-len(nbarray.Values)-4:], uint32(len(nbarray.Values)))
	pos := 0
	start := len(data) - len(nbarray.Values)
	for i := start; i < start+len(nbarray.Values); i++ {
		data[i] = byte(nbarray.Values[pos])
		pos++
	}
	w.Write(data)
}

func (nstring String) WriteTo(w io.Writer, hasName bool) {
	var data []byte
	if hasName {
		data = make([]byte, 1+2+len(nstring.Name)+2+len(nstring.Value))
		data[0] = 8
		binary.BigEndian.PutUint16(data[1:3], uint16(len(nstring.Name)))
		copy(data[3:], []byte(nstring.Name))
	} else {
		data = make([]byte, 2+len(nstring.Value))
	}
	binary.BigEndian.PutUint16(data[len(data)-len(nstring.Value)-2:], uint16(len(nstring.Value)))
	copy(data[len(data)-len(nstring.Value):], []byte(nstring.Value))
	w.Write(data)
}

func (niarray IntArray) WriteTo(w io.Writer, hasName bool) {
	var data []byte
	if hasName {
		data = make([]byte, 1+2+len(niarray.Name)+4+len(niarray.Values)*4)
		data[0] = 11
		binary.BigEndian.PutUint16(data[1:3], uint16(len(niarray.Name)))
		copy(data[3:], []byte(niarray.Name))
	} else {
		data = make([]byte, 4+len(niarray.Values)*4)
	}
	binary.BigEndian.PutUint32(data[len(data)-len(niarray.Values)*4-4:], uint32(len(niarray.Values)))
	pos := 0
	start := len(data) - len(niarray.Values)*4
	for i := start; i < start+len(niarray.Values); i += 4 {
		binary.BigEndian.PutUint32(data[i:i+4], uint32(niarray.Values[pos]))
		pos++
	}
	w.Write(data)
}

func (nlist List) WriteTo(w io.Writer, hasName bool) {
	var data []byte
	if hasName {
		data = make([]byte, 1+2+len(nlist.Name)+5)
		data[0] = 9
		binary.BigEndian.PutUint16(data[1:3], uint16(len(nlist.Name)))
		copy(data[3:], []byte(nlist.Name))
	} else {
		data = make([]byte, 5)
	}
	data[len(data)-5] = nlist.Type
	binary.BigEndian.PutUint32(data[len(data)-4:], uint32(len(nlist.Tags)))
	w.Write(data)
	for _, tag := range nlist.Tags {
		b := tag.(WriterTo)
		b.WriteTo(w, false)
	}
}

type (
	Reader struct {
		R io.Reader
	}
	Byte struct {
		Name  string
		Value int8
	}
	Short struct {
		Name  string
		Value int16
	}
	Int struct {
		Name  string
		Value int32
	}
	Long struct {
		Name  string
		Value int64
	}
	Float struct {
		Name  string
		Value float32
	}
	Double struct {
		Name  string
		Value float64
	}
	ByteArray struct {
		Name   string
		Values []int8
	}
	String struct {
		Name  string
		Value string
	}
	List struct {
		Name string
		Type byte
		Tags []WriterTo
	}
	Compound struct {
		Name string
		Tags map[string]WriterTo
	}
	IntArray struct {
		Name   string
		Values []int32
	}
	WriterTo interface {
		WriteTo(io.Writer, bool)
	}
)

func NewCompound() *Compound {
	return &Compound{Tags: make(map[string]WriterTo)}
}

func (r *Reader) ReadUByte() byte {
	b := make([]byte, 1)
	r.R.Read(b)
	return b[0]
}

func (r *Reader) ReadByte() int8 {
	return int8(r.ReadUByte())
}

func (r *Reader) ReadUShort() uint16 {
	b := make([]byte, 2)
	r.R.Read(b)
	return binary.BigEndian.Uint16(b)
}

func (r *Reader) ReadShort() int16 {
	return int16(r.ReadUShort())
}

func (r *Reader) ReadInt() int32 {
	b := make([]byte, 4)
	r.R.Read(b)
	return int32(binary.BigEndian.Uint32(b))
}

func (r *Reader) ReadLong() int64 {
	b := make([]byte, 8)
	r.R.Read(b)
	return int64(binary.BigEndian.Uint64(b))
}

func (r *Reader) ReadFloat() float32 {
	b := make([]byte, 4)
	r.R.Read(b)
	return math.Float32frombits(binary.BigEndian.Uint32(b))
}

func (r *Reader) ReadDouble() float64 {
	b := make([]byte, 8)
	r.R.Read(b)
	return math.Float64frombits(binary.BigEndian.Uint64(b))
}

func (r *Reader) ReadString() string {
	l := int(r.ReadUShort())
	d := make([]byte, l)
	r.R.Read(d)
	return string(d)
}

const (
	TypeEnd byte = iota
	TypeByte
	TypeShort
	TypeInt
	TypeLong
	TypeFloat
	TypeDouble
	TypeByteArray
	TypeString
	TypeList
	TypeCompound
	TypeIntArray
)
