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
	"io/ioutil"
	"testing"
)

func BenchmarkWriteMap(b *testing.B) {
	var buf bytes.Buffer
	Write(&buf, mapData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Write(ioutil.Discard, mapData)
	}
	b.SetBytes(int64(buf.Len()))
}

type testStruct struct {
	NestedCompoundTest testStruct2   `msgpack:"nested compound test"`
	IntTest            int32         `msgpack:"intTest"`
	ByteTest           byte          `msgpack:"byteTest"`
	StringTest         string        `msgpack:"stringTest"`
	ListTest           []int64       `msgpack:"listTest"`
	DoubleTest         float64       `msgpack:"doubleTest"`
	FloatTest          float32       `msgpack:"floatTest"`
	LongTest           int64         `msgpack:"longTest"`
	ListTestCompound   []testStruct3 `msgpack:"list (compound)"`
	ByteArrayTest      []byte        `msgpack:"byteArrayTest"`
	ShortTest          int16         `msgpack:"shortTest"`
}

type testStruct2 struct {
	Egg testStruct2Value `msgpack:"egg"`
	Ham testStruct2Value `msgpack:"ham"`
}

type testStruct2Value struct {
	Name  string  `msgpack:"name"`
	Value float32 `msgpack:"value"`
}

type testStruct3 struct {
	CreatedOn int64  `msgpack:"created-on"`
	Name      string `msgpack:"name"`
}

func BenchmarkWriteStruct(b *testing.B) {
	var buf bytes.Buffer
	data := testStruct{
		NestedCompoundTest: testStruct2{
			Egg: testStruct2Value{
				Name:  "Eggbert",
				Value: 0.5,
			},
			Ham: testStruct2Value{
				Name:  "Hampus",
				Value: 0.75,
			},
		},
		IntTest:    2147483647,
		ByteTest:   127,
		StringTest: "HELLO WORLD THIS IS A TEST STRING \xc3\x85\xc3\x84\xc3\x96!",
		ListTest: []int64{
			11,
			12,
			13,
			14,
			15,
		},
		DoubleTest: 0.49312871321823148,
		FloatTest:  0.49823147058486938,
		LongTest:   9223372036854775807,
		ListTestCompound: []testStruct3{
			testStruct3{
				CreatedOn: 1264099775885,
				Name:      "Compound tag #0",
			},
			testStruct3{
				CreatedOn: 1264099775885,
				Name:      "Compound tag #1",
			},
		},
		ByteArrayTest: []byte{1, 2, 3, 4, 5, 6, 7, 8, 10},
		ShortTest:     32767,
	}
	Write(&buf, data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Write(ioutil.Discard, data)
	}
	b.SetBytes(int64(buf.Len()))
}

type benchSlice struct {
	Val []byte
}

func BenchmarkByteSliceWrite(b *testing.B) {
	var buf bytes.Buffer
	data := benchSlice{
		Val: make([]byte, 16*16*256),
	}
	Write(&buf, data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Write(ioutil.Discard, data)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkByteSliceRead(b *testing.B) {
	var buf bytes.Buffer
	data := benchSlice{
		Val: make([]byte, 16*16*256),
	}
	Write(&buf, data)

	reader := bytes.NewReader(buf.Bytes())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Read(reader, &data)
	}
	b.SetBytes(int64(buf.Len()))
}

type benchArray struct {
	Val [16 * 16 * 256]byte
}

func BenchmarkByteArrayWrite(b *testing.B) {
	var buf bytes.Buffer
	data := &benchArray{}
	Write(&buf, data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Write(ioutil.Discard, data)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkByteArrayRead(b *testing.B) {
	var buf bytes.Buffer
	data := &benchArray{}
	Write(&buf, data)

	reader := bytes.NewReader(buf.Bytes())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Read(reader, data)
	}
	b.SetBytes(int64(buf.Len()))
}
