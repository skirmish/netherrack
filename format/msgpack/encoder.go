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
	"io"
	"reflect"
	"unsafe"
)

type Encoder struct {
	w io.Writer
	b [9]byte //A buffer to save allocations
}

//Creates a new encoder that will write to the writer
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
	}
}

//Encodes the value into to the writer in the msgpack format
func (enc *Encoder) Encode(e interface{}) error {
	return enc.EncodeValue(reflect.ValueOf(e))
}

//Encodes the value into to the writer in the msgpack format
func (enc *Encoder) EncodeValue(value reflect.Value) error {
	if value.Kind() != reflect.Ptr {
		panic("Encode requires a pointer type")
	}
	value = value.Elem()
	eng := getEngine(value.Type())

	if encode, ok := eng.encode.(encodeFunc); ok {
		return encode(enc, unsafe.Pointer(value.UnsafeAddr()))
	} else {
		encode := eng.encode.(encodeReflectFunc)
		return encode(enc, value)
	}
}
