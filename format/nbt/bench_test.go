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
	"bytes"
	"io/ioutil"
	"testing"
)

func BenchmarkWriteMap(b *testing.B) {
	var buf bytes.Buffer
	Write(&buf, mapData, "")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Write(ioutil.Discard, mapData, "")
	}
	b.SetBytes(int64(buf.Len()))
}

type testStruct struct {
	NestedCompoundTest testStruct2   `nbt:"nested compound test"`
	IntTest            int32         `nbt:"intTest"`
	ByteTest           int8          `nbt:"byteTest"`
	StringTest         string        `nbt:"stringTest"`
	ListTest           []int64       `nbt:"listTest"`
	DoubleTest         float64       `nbt:"doubleTest"`
	FloatTest          float32       `nbt:"floatTest"`
	LongTest           int64         `nbt:"longTest"`
	ListTestCompound   []testStruct3 `nbt:"list (compound)"`
	ByteArrayTest      []byte        `nbt:"byteArrayTest"`
	ShortTest          int16         `nbt:"shortTest"`
}

type testStruct2 struct {
	Egg testStruct2Value `nbt:"egg"`
	Ham testStruct2Value `nbt:"ham"`
}

type testStruct2Value struct {
	Name  string  `nbt:"name"`
	Value float32 `nbt:"value"`
}

type testStruct3 struct {
	CreatedOn int64  `nbt:"created-on"`
	Name      string `nbt:"name"`
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
	Write(&buf, data, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Write(ioutil.Discard, data, "")
	}
	b.SetBytes(int64(buf.Len()))
}
