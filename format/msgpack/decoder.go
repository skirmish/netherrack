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
	"io"
	"reflect"
	"unsafe"
)

type Decoder struct {
	r *bufio.Reader
	b [9]byte //A buffer to save allocations
}

type fullReader struct{ io.Reader }

func (fr fullReader) Read(b []byte) (int, error) {
	return io.ReadFull(fr.Reader, b)
}

//Creates a new decoder that will read to the reader
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: bufio.NewReader(fullReader{r}),
	}
}

//Decodes the value from the reader in the msgpack format
func (dec *Decoder) Decode(e interface{}) error {
	return dec.DecodeValue(reflect.ValueOf(e))
}

//Decodes the value from the reader in the msgpack format
func (dec *Decoder) DecodeValue(value reflect.Value) error {
	if value.Kind() != reflect.Ptr {
		panic("Decode requires a pointer type")
	}
	value = value.Elem()
	eng := getEngine(value.Type())

	if decode, ok := eng.decode.(decodeFunc); ok {
		return decode(dec, unsafe.Pointer(value.UnsafeAddr()))
	} else {
		decode := eng.decode.(decodeReflectFunc)
		return decode(dec, value)
	}
}
