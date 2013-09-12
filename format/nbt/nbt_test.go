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
	"reflect"
	"testing"
)

type typeBool struct {
	Val bool
}

func TestBool(t *testing.T) {
	v := typeBool{true}

	var buf bytes.Buffer
	err := Write(&buf, &v, "")
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
	Val8  int8
	Val16 int16
	Val32 int32
	Val64 int64
}

func TestInt(t *testing.T) {
	v := typeInt{-0x0E, 0x0ECC, -0x0ECCCCCC, 0x0ECCCCCCCCCCCCCC}
	org := v

	var buf bytes.Buffer
	err := Write(&buf, &v, "")
	if err != nil {
		t.Fatal(err)
	}

	v = typeInt{}
	err = Read(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, org) {
		t.Fail()
	}
}

type typeString struct {
	Val string
}

func TestString(t *testing.T) {
	v := typeString{"Hello world"}

	var buf bytes.Buffer
	err := Write(&buf, &v, "")
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
	err := Write(&buf, &v, "")
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
	"listTest": []int64{
		int64(11),
		int64(12),
		int64(13),
		int64(14),
		int64(15),
	},
	"doubleTest": float64(0.49312871321823148),
	"floatTest":  float32(0.49823147058486938),
	"longTest":   int64(9223372036854775807),
	"listTest (compound)": []map[string]interface{}{
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
	t.Skip("Broken")
	var buf bytes.Buffer
	err := Write(&buf, mapData, "")
	if err != nil {
		t.Fatal(err)
	}
	out := testDataMap{map[string]interface{}{}}
	err = Read(&buf, &out)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(mapData, out) {
		t.Errorf("Wanted: %v", mapData)
		t.Errorf("Got: %v", out)
		t.Fail()
	}
}

type testArrayBytes struct {
	Val []byte
}

func TestArrayBytes(t *testing.T) {
	v := testArrayBytes{[]byte{5, 6, 7, 8, 9, 22}}
	org := v

	var buf bytes.Buffer
	err := Write(&buf, &v, "")
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

type testArrayInt struct {
	Val []int32
}

func TestArrayInt(t *testing.T) {
	v := testArrayInt{[]int32{5, -6, 7, -8, 9, -22}}
	org := v

	var buf bytes.Buffer
	err := Write(&buf, &v, "")
	if err != nil {
		t.Fatal(err)
	}

	v = testArrayInt{}
	err = Read(&buf, &v)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, org) {
		t.Fail()
	}
}

type testArray struct {
	Val1 []float32
	Val2 []testArrayStruct
	Val3 []string
}
type testArrayStruct struct {
	Name string
}

func TestArray(t *testing.T) {
	v := testArray{
		[]float32{0.5, 0.3, 0.6, 0.1},
		[]testArrayStruct{
			testArrayStruct{"Bob"},
			testArrayStruct{"Jim"},
		},
		[]string{"Hello", "World"},
	}
	org := v

	var buf bytes.Buffer
	err := Write(&buf, &v, "")
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
