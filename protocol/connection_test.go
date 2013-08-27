package protocol

import (
	"bytes"
	"io/ioutil"
	"testing"
)

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

func BenchmarkConnectionSimple(b *testing.B) {
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

func BenchmarkConnectionLong(b *testing.B) {
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

func BenchmarkSlotAll(b *testing.B) {
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

func BenchmarkSlotNoTag(b *testing.B) {
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

func BenchmarkSlotEmpty(b *testing.B) {
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
