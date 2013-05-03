package protocol

import (
	"bytes"
	"encoding/binary"
	"github.com/thinkofdeath/netherrack/items"
	"github.com/thinkofdeath/netherrack/nbt"
	"github.com/thinkofdeath/soulsand"
	"math"
)

func (c *Conn) readBool() bool {
	return c.readUByte() == 1
}

func (c *Conn) ReadUByte() byte {
	data := make([]byte, 1)
	c.Read(data)
	return data[0]
}

func (c *Conn) readUByte() byte {
	data := make([]byte, 1)
	c.Read(data)
	return data[0]
}

func (c *Conn) readByte() int8 {
	return int8(c.readUByte())
}

func (c *Conn) readUShort() uint16 {
	data := make([]byte, 2)
	c.Read(data)
	return binary.BigEndian.Uint16(data)
}

func (c *Conn) readShort() int16 {
	return int16(c.readUShort())
}

func (c *Conn) readInt() int32 {
	data := make([]byte, 4)
	c.Read(data)
	return int32(binary.BigEndian.Uint32(data))
}

func (c *Conn) readLong() int64 {
	data := make([]byte, 8)
	c.Read(data)
	return int64(binary.BigEndian.Uint64(data))
}

func (c *Conn) readFloat() float32 {
	data := make([]byte, 4)
	c.Read(data)
	return math.Float32frombits(binary.BigEndian.Uint32(data))
}

func (c *Conn) readDouble() float64 {
	data := make([]byte, 8)
	c.Read(data)
	return math.Float64frombits(binary.BigEndian.Uint64(data))
}

func (c *Conn) readString() string {
	length := c.readShort()
	runes := make([]rune, length)
	for i, _ := range runes {
		runes[i] = rune(c.readUShort())
	}
	return string(runes)
}

func (c *Conn) readSlot() (itemstack soulsand.ItemStack) {
	itemstackRaw := &items.ItemStack{}
	itemstack = itemstackRaw
	itemstackRaw.ID = c.readShort()
	if itemstackRaw.ID == -1 {
		return
	}
	itemstackRaw.Count = c.readUByte()
	itemstackRaw.Damage = c.readShort()
	if l := c.readShort(); l != -1 {
		data := make([]byte, l)
		c.Read(data)
		itemstackRaw.Tag = nbt.Parse(bytes.NewReader(data))
	}
	return
}
