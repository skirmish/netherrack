package nbt

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func BenchmarkWrite(b *testing.B) {
	var buf bytes.Buffer
	data := Type{
		"nested compound test": Type{
			"egg": Type{
				"name":  "Eggbert",
				"value": float32(0.5),
			},
			"ham": Type{
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
			Type{
				"created-on": int64(1264099775885),
				"name":       "Compound tag #0",
			},
			Type{
				"created-on": int64(1264099775885),
				"name":       "Compound tag #1",
			},
		},
		"byteArrayTest": []byte{1, 2, 3, 4, 5, 6, 7, 8, 10},
		"shortTest":     int16(32767),
	}
	data.WriteTo(&buf, "Level")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data.WriteTo(ioutil.Discard, "Level")
	}
	b.SetBytes(int64(buf.Len()))
}
