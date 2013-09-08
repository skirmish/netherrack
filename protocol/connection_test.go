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

package protocol

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestPacketsID(t *testing.T) {
	for i, packetType := range packets {
		if packetType == nil {
			continue
		}
		if reflect.New(packetType).Elem().Interface().(Packet).ID() != byte(i) {
			t.Fatalf("Id mis-match: %d", i)
		}
	}
}

func TestPackets(t *testing.T) {
	for _, packetType := range packets {
		if packetType == nil {
			continue
		}
		fields(packetType)
	}
}

func TestConnection(t *testing.T) {
	var buf bytes.Buffer
	conn := &Conn{
		Out: &buf,
	}
	conn.WritePacket(KeepAlive{0x60})
	if !bytes.Equal(buf.Bytes(), []byte{0x00, 0x00, 0x00, 0x00, 0x60}) {
		t.Error(buf.Bytes())
		t.FailNow()
	}
}

func TestSlot(t *testing.T) {
	var buf bytes.Buffer
	conn := &Conn{
		Out: &buf,
	}
	conn.WritePacket(EntityEquipment{
		EntityID: 0x88,
		Slot:     0x03,
		Item: Slot{
			ID:     0x05,
			Count:  0x01,
			Damage: 0x60,
			Tag:    []byte{69, 69},
		},
	})
	if !bytes.Equal(buf.Bytes(), []byte{5, 0, 0, 0, 136, 0, 3, 0, 5, 1, 0, 96, 0, 2, 69, 69}) {
		t.Error(buf.Bytes())
		t.FailNow()
	}
}
func BenchmarkSimpleWrite(b *testing.B) {
	packet := KeepAlive{55}
	conn := &Conn{}
	var buf bytes.Buffer
	conn.Out = &buf
	conn.WritePacket(packet)
	conn.Out = ioutil.Discard
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.WritePacket(packet)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkSimpleRead(b *testing.B) {
	packet := KeepAlive{55}
	conn := &Conn{}
	var buf bytes.Buffer
	conn.Out = &buf
	conn.WritePacket(packet)
	conn.Out = ioutil.Discard
	reader := bytes.NewReader(buf.Bytes())
	conn.In = reader
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.ReadPacket()
		reader.Seek(0, 0)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkLongCompile(b *testing.B) {
	packet := LoginRequest{}
	t := reflect.TypeOf(packet)
	delete(fieldCache.m, t)
	delete(fieldCache.create, t)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fields(t)
		delete(fieldCache.m, t)
		delete(fieldCache.create, t)
	}
}

func BenchmarkLongWrite(b *testing.B) {
	packet := LoginRequest{
		EntityID:   5745,
		LevelType:  "largeBiomes",
		Gamemode:   0,
		Dimension:  -1,
		Difficulty: 3,
		MaxPlayers: 60,
	}
	conn := &Conn{}
	var buf bytes.Buffer
	conn.Out = &buf
	conn.WritePacket(packet)
	conn.Out = ioutil.Discard
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.WritePacket(packet)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkLongRead(b *testing.B) {
	packet := LoginRequest{
		EntityID:   5745,
		LevelType:  "largeBiomes",
		Gamemode:   0,
		Dimension:  -1,
		Difficulty: 3,
		MaxPlayers: 60,
	}
	conn := &Conn{}
	var buf bytes.Buffer
	conn.Out = &buf
	conn.WritePacket(packet)
	conn.Out = ioutil.Discard
	reader := bytes.NewReader(buf.Bytes())
	conn.In = reader
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.ReadPacket()
		reader.Seek(0, 0)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkSlotAllWrite(b *testing.B) {
	packet := EntityEquipment{
		EntityID: 0x88,
		Slot:     0x03,
		Item: Slot{
			ID:     0x05,
			Count:  0x01,
			Damage: 0x60,
			Tag:    []byte{69, 69},
		},
	}
	conn := &Conn{}
	var buf bytes.Buffer
	conn.Out = &buf
	conn.WritePacket(packet)
	conn.Out = ioutil.Discard
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.WritePacket(packet)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkSlotAllRead(b *testing.B) {
	packet := EntityEquipment{
		EntityID: 0x88,
		Slot:     0x03,
		Item: Slot{
			ID:     0x05,
			Count:  0x01,
			Damage: 0x60,
			Tag:    []byte{69, 69},
		},
	}
	conn := &Conn{}
	var buf bytes.Buffer
	conn.Out = &buf
	conn.WritePacket(packet)
	conn.Out = ioutil.Discard
	reader := bytes.NewReader(buf.Bytes())
	conn.In = reader
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.ReadPacket()
		reader.Seek(0, 0)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkSlotNoTagWrite(b *testing.B) {
	packet := EntityEquipment{
		EntityID: 0x88,
		Slot:     0x03,
		Item: Slot{
			ID:     0x05,
			Count:  0x01,
			Damage: 0x60,
		},
	}
	conn := &Conn{}
	var buf bytes.Buffer
	conn.Out = &buf
	conn.WritePacket(packet)
	conn.Out = ioutil.Discard
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.WritePacket(packet)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkSlotNoTagRead(b *testing.B) {
	packet := EntityEquipment{
		EntityID: 0x88,
		Slot:     0x03,
		Item: Slot{
			ID:     0x05,
			Count:  0x01,
			Damage: 0x60,
		},
	}
	conn := &Conn{}
	var buf bytes.Buffer
	conn.Out = &buf
	conn.WritePacket(packet)
	conn.Out = ioutil.Discard
	reader := bytes.NewReader(buf.Bytes())
	conn.In = reader
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.ReadPacket()
		reader.Seek(0, 0)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkSlotEmptyWrite(b *testing.B) {
	packet := EntityEquipment{
		EntityID: 0x88,
		Slot:     0x03,
		Item: Slot{
			ID: -1,
		},
	}
	conn := &Conn{}
	var buf bytes.Buffer
	conn.Out = &buf
	conn.WritePacket(packet)
	conn.Out = ioutil.Discard
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.WritePacket(packet)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkSlotEmptyRead(b *testing.B) {
	packet := EntityEquipment{
		EntityID: 0x88,
		Slot:     0x03,
		Item: Slot{
			ID: -1,
		},
	}
	conn := &Conn{}
	var buf bytes.Buffer
	conn.Out = &buf
	conn.WritePacket(packet)
	conn.Out = ioutil.Discard
	reader := bytes.NewReader(buf.Bytes())
	conn.In = reader
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.ReadPacket()
		reader.Seek(0, 0)
	}
	b.SetBytes(int64(buf.Len()))
}
