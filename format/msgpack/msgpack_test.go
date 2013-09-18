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
	"reflect"
	"testing"
)

type typePointer struct {
	Val  *bool
	Val2 *bool
	Val3 *int
}

func TestPtr(t *testing.T) {
	temp := true
	temp2 := 5673
	v := typePointer{Val2: &temp, Val3: &temp2}

	var buf bytes.Buffer
	err := Write(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}
	v = typePointer{new(bool), nil, nil}
	err = Read(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	if v.Val != nil || *v.Val2 != true || *v.Val3 != temp2 {
		t.Fail()
	}
}

type typeBool struct {
	Val bool
}

func TestBool(t *testing.T) {
	v := typeBool{true}

	var buf bytes.Buffer
	err := Write(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}
	v = typeBool{}
	err = Read(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	if !v.Val {
		t.Fail()
	}
}

type typeInt struct {
	Val8   int8
	ValU8  uint8
	Val16  int16
	ValU16 uint16
	Val32  int32
	ValU32 uint32
	Val64  int64
	ValU64 uint64
}

func TestInt(t *testing.T) {
	v := typeInt{-1, 2, -3, 4, -5, 6, -7, 8}

	var buf bytes.Buffer
	err := Write(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	v = typeInt{}
	err = Read(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, typeInt{-1, 2, -3, 4, -5, 6, -7, 8}) {
		t.Fail()
	}
}

type typeString struct {
	Val string
}

func TestString(t *testing.T) {
	v := typeString{"Hello world"}

	var buf bytes.Buffer
	err := Write(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	v = typeString{}
	err = Read(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	if v.Val != "Hello world" {
		t.Fail()
	}
}

type typeFloat struct {
	Val32 float32
	Val64 float64
}

func TestFloat(t *testing.T) {
	v := typeFloat{5, 7}
	org := v

	var buf bytes.Buffer
	err := Write(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	v = typeFloat{}
	err = Read(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, org) {
		t.Fail()
	}
}

type testDataMap struct {
	Data map[string]interface{}
}

var mapData = testDataMap{map[string]interface{}{
	"nested compound test": map[string]interface{}{
		"egg": map[string]interface{}{
			"name":  "Eggbert",
			"value": float32(0.5),
		},
		"ham": map[string]interface{}{
			"name":  "Hampus",
			"value": float32(0.75),
		},
	},
	"intTest":    int32(2147483647),
	"byteTest":   int8(127),
	"stringTest": "HELLO WORLD THIS IS A TEST STRING \xc3\x85\xc3\x84\xc3\x96!",
	"listTest": []interface{}{
		int64(11),
		int64(12),
		int64(13),
		int64(14),
		int64(15),
	},
	"doubleTest": float64(0.49312871321823148),
	"floatTest":  float32(0.49823147058486938),
	"longTest":   int64(9223372036854775807),
	"listTest (compound)": []interface{}{
		map[string]interface{}{
			"created-on": int64(1264099775885),
			"name":       "Compound tag #0",
		},
		map[string]interface{}{
			"created-on": int64(1264099775885),
			"name":       "Compound tag #1",
		},
	},
	"byteArrayTest": []byte{1, 2, 3, 4, 5, 6, 7, 8, 10},
	"shortTest":     int16(32767),
}}

func TestMap(t *testing.T) {
	t.Skip("Currently broken")
	var buf bytes.Buffer
	Write(&buf, mapData)
	out := testDataMap{map[string]interface{}{}}
	Read(&buf, &out)
	if !reflect.DeepEqual(mapData, out) {
		t.Errorf("Wanted: %v", mapData)
		t.Errorf("Got: %v", out)
		t.Fail()
	}
}

type testSliceBytes struct {
	Val []byte
}

func TestSliceBytes(t *testing.T) {
	v := testSliceBytes{[]byte{5, 6, 7, 8, 9, 22}}
	org := v

	var buf bytes.Buffer
	err := Write(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	v = testSliceBytes{}
	err = Read(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, org) {
		t.Fail()
	}
}

type testArrayBytes struct {
	Val [6]byte
}

func TestArrayBytes(t *testing.T) {
	v := testArrayBytes{[6]byte{5, 6, 7, 8, 9, 22}}
	org := v

	var buf bytes.Buffer
	err := Write(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	v = testArrayBytes{}
	err = Read(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, org) {
		t.Fail()
	}
}

type testSlice struct {
	Val  []int
	Val2 []uint
	Val3 []float32
	Val4 []testSliceStruct
	Val5 []string
	Val6 []*testSliceStruct
}
type testSliceStruct struct {
	Name string
}

func TestSlice(t *testing.T) {
	v := testSlice{
		[]int{5, -6, 7, -8, 9, -22},
		[]uint{5, 363, 73, 7, 784, 6},
		[]float32{0.5, 0.3, 0.6, 0.1},
		[]testSliceStruct{
			testSliceStruct{"Bob"},
			testSliceStruct{"Jim"},
		},
		[]string{"Hello", "World"},
		[]*testSliceStruct{
			&testSliceStruct{"Bob"},
			&testSliceStruct{"Jim"},
		},
	}
	org := v

	var buf bytes.Buffer
	err := Write(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	v = testSlice{}
	err = Read(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, org) {
		t.Fail()
	}
}

type testArray struct {
	Val  [6]int
	Val2 [6]uint
	Val3 [4]float32
	Val4 [2]testArrayStruct
	Val5 [2]string
	Val6 [2]*testArrayStruct
}
type testArrayStruct struct {
	Name string
}

func TestArray(t *testing.T) {
	v := testArray{
		[6]int{5, -6, 7, -8, 9, -22},
		[6]uint{5, 363, 73, 7, 784, 6},
		[4]float32{0.5, 0.3, 0.6, 0.1},
		[2]testArrayStruct{
			testArrayStruct{"Bob"},
			testArrayStruct{"Jim"},
		},
		[2]string{"Hello", "World"},
		[2]*testArrayStruct{
			&testArrayStruct{"Bob"},
			&testArrayStruct{"Jim"},
		},
	}
	org := v

	var buf bytes.Buffer
	err := Write(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	v = testArray{}
	err = Read(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, org) {
		t.Fail()
	}
}

type testSliceInterface struct {
	Val []interface{}
}

func TestSliceInterface(t *testing.T) {
	t.Skip("Currently broken")
	v := testSliceInterface{
		[]interface{}{
			"Hello",
			5,
			0.3,
		},
	}
	org := v

	var buf bytes.Buffer
	err := Write(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	v = testSliceInterface{}
	err = Read(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, org) {
		t.Fail()
	}
}

type testEmbeded struct {
	Hello int
	EmbededStruct
}

type EmbededStruct struct {
	Cake string
}

func TestEmbeded(t *testing.T) {
	v := testEmbeded{5, EmbededStruct{"Hello"}}
	org := v

	var buf bytes.Buffer
	err := Write(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	v = testEmbeded{}
	err = Read(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, org) {
		t.Fail()
	}
}
