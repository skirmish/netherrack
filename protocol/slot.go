package protocol

import (
	"bytes"
	"compress/gzip"
	"github.com/NetherrackDev/netherrack/items"
	"github.com/NetherrackDev/soulsand"
)

func (c *Conn) writeItemstack(itemstack soulsand.ItemStack) {
	if itemstack == nil {
		out := NewByteWriter(2)
		out.WriteShort(-1)
		c.Write(out.Bytes())
		return
	}
	slotData := itemstack.(*items.ItemStack)
	slotData.Lock.RLock()
	defer slotData.Lock.RUnlock()
	var out *byteWriter
	if slotData.ID() == -1 {
		out = NewByteWriter(2)
	} else {
		out = NewByteWriter(2 + 1 + 2 + 2)
	}

	dataLength := int16(-1)
	var data []byte
	if slotData.Tag != nil {
		var buf bytes.Buffer
		gz, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
		defer gz.Close()
		slotData.Tag.WriteTo(gz, "tag")
		gz.Flush()
		data = buf.Bytes()
		dataLength = int16(len(data))
	}

	out.WriteShort(slotData.ID())
	if slotData.ID() == -1 {
		c.Write(out.Bytes())
		return
	}
	out.WriteUByte(slotData.Count())
	out.WriteShort(slotData.Data())
	out.WriteShort(dataLength)
	c.Write(out.Bytes())
	if dataLength != -1 {
		c.Write(data)
	}
}
