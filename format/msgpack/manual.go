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
	"encoding/binary"
	"errors"
	"io"
	"math"
)

//Shortcuts for simple types to save an alloc
func skip(r io.Reader, de *msgDecoder) error {
	bs := de.b[:1]
	_, err := r.Read(bs)
	if err != nil {
		return err
	}
	switch bs[0] {
	case 0xC0, 0xC2, 0xC3: //0 byte
	case 0xCC, 0xD0: //1 byte
		bs = de.b[:1]
		_, err := r.Read(bs)
		if err != nil {
			return err
		}
	case 0xCD, 0xD1: //2 byte
		bs = de.b[:2]
		_, err := r.Read(bs)
		if err != nil {
			return err
		}
	case 0xCE, 0xD2, 0xCA: //4 byte
		bs = de.b[:4]
		_, err := r.Read(bs)
		if err != nil {
			return err
		}
	case 0xCF, 0xD3, 0xCB: //8 byte
		bs = de.b[:8]
		_, err := r.Read(bs)
		if err != nil {
			return err
		}
	}
	_, err = fallbackRead(r, de)
	return err
}

//If the type isn't known at read time then this will manually check the type
//and return the value
func fallbackRead(r io.Reader, de *msgDecoder) (interface{}, error) {
	bs := de.b[:1]
	_, err := r.Read(bs)
	if err != nil {
		return nil, err
	}
	switch bs[0] {
	case 0xC0: //Nil
		return nil, nil
	case 0xC2: //False
		return false, nil
	case 0xC3: //True
		return true, nil

	case 0xCC: //8-Bit Signed int
		bs = de.b[:1]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		return int8(bs[0]), nil
	case 0xD0: //8-Bit Unsigned int
		bs = de.b[:1]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		return byte(bs[0]), nil

	case 0xCD: //16-Bit Signed int
		bs = de.b[:2]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		return int16(binary.LittleEndian.Uint16(bs)), nil
	case 0xD1: //16-Bit Unsigned int
		bs = de.b[:2]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		return binary.LittleEndian.Uint16(bs), nil

	case 0xCE: //32-Bit Unsigned int
		bs = de.b[:4]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		return int32(binary.LittleEndian.Uint32(bs)), nil
	case 0xD2: //32-Bit Signed int
		bs = de.b[:4]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		return binary.LittleEndian.Uint32(bs), nil
	case 0xCA: //32-Bit Float
		bs = de.b[:4]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		return math.Float32frombits(binary.LittleEndian.Uint32(bs)), nil

	case 0xCF: //64-Bit Signed int
		bs = de.b[:8]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		return int64(binary.LittleEndian.Uint64(bs)), nil
	case 0xD3: //64-Bit Unsigned int
		bs = de.b[:8]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		return binary.LittleEndian.Uint64(bs), nil
	case 0xCB: //64-Bit Float
		bs = de.b[:8]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		return math.Float64frombits(binary.LittleEndian.Uint64(bs)), nil

	case 0xDC: //Array 16-Bit
		bs = de.b[:2]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		l := binary.LittleEndian.Uint16(bs)
		return readArray(r, de, int(l))
	case 0xDD: //Array 32-Bit
		bs = de.b[:4]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		l := binary.LittleEndian.Uint32(bs)
		return readArray(r, de, int(l))

	case 0xDE: //Map 16-Bit
		bs = de.b[:2]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		l := binary.LittleEndian.Uint16(bs)
		return readMap(r, de, int(l))
	case 0xDF: //Map 32-Bit
		bs = de.b[:4]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		l := binary.LittleEndian.Uint32(bs)
		return readMap(r, de, int(l))

	case 0xD9: //8-Bit String
		bs = de.b[:1]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		l := bs[0]
		b := make([]byte, l)
		_, err = r.Read(b)
		if err != nil {
			return nil, err
		}
		return string(b), nil
	case 0xDA: //16-Bit String
		bs = de.b[:2]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		l := binary.LittleEndian.Uint16(bs)
		b := make([]byte, l)
		_, err = r.Read(b)
		if err != nil {
			return nil, err
		}
		return string(b), nil
	case 0xDB: //32-Bit String
		bs = de.b[:4]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		l := binary.LittleEndian.Uint32(bs)
		b := make([]byte, l)
		_, err = r.Read(b)
		if err != nil {
			return nil, err
		}
		return string(b), nil

	case 0xC4: //8-Bit Byte array
		bs = de.b[:1]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		l := bs[0]
		b := make([]byte, l)
		_, err = r.Read(b)
		if err != nil {
			return nil, err
		}
		return b, nil
	case 0xC5: //16-Bit Byte array
		bs = de.b[:2]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		l := binary.LittleEndian.Uint16(bs)
		b := make([]byte, l)
		_, err = r.Read(b)
		if err != nil {
			return nil, err
		}
		return b, nil
	case 0xC6: //32-Bit Byte array
		bs = de.b[:4]
		_, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		l := binary.LittleEndian.Uint32(bs)
		b := make([]byte, l)
		_, err = r.Read(b)
		if err != nil {
			return nil, err
		}
		return b, nil

	default: //Special types
		if bs[0]&0x80 == 0x00 {
			return int8(bs[0] & 0x7F), nil
		}
		if bs[0]&0xE0 == 0xE0 {
			return -int8(bs[0] & 0x1F), nil
		}
		if bs[0]&0xF0 == 0x90 {
			l := bs[0] & 0x0F
			return readArray(r, de, int(l))
		}
		if bs[0]&0xF0 == 0x80 {
			l := bs[0] & 0x0F
			return readMap(r, de, int(l))
		}
		if bs[0]&0xE0 == 0xA0 {
			l := bs[0] & 0x1F
			b := make([]byte, l)
			_, err := r.Read(b)
			if err != nil {
				return nil, err
			}
			return string(b), nil
		}
	}
	return nil, ErrorUnknownType
}

var ErrorUnknownType = errors.New("format/msgpack: Unknown type")

//Reads an array of an unknown type
func readArray(r io.Reader, de *msgDecoder, l int) ([]interface{}, error) {
	out := make([]interface{}, l)
	var err error
	for i := range out {
		out[i], err = fallbackRead(r, de)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

//Reads an map of an unknown type
func readMap(r io.Reader, de *msgDecoder, l int) (map[interface{}]interface{}, error) {
	out := make(map[interface{}]interface{}, l)
	for i := 0; i < l; i++ {
		key, err := fallbackRead(r, de)
		if err != nil {
			return nil, err
		}
		val, err := fallbackRead(r, de)
		if err != nil {
			return nil, err
		}
		out[key] = val
	}
	return out, nil
}
