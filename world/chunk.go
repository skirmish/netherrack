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

package world

import (
	"bytes"
	"compress/zlib"
	"github.com/NetherrackDev/netherrack/protocol"
)

type Chunk interface {
	//Returns the chunk's location
	Location() (x, z int)
	//Inits the chunk. Should only be called by the world
	Init(world *World, gen Generator)
	//Adds the watcher to the chunk
	Join(watcher Watcher)
}

type Watcher interface {
	//Queues a packet to be sent to the watcher
	QueuePacket(packet protocol.Packet)
}

//A chunk loaded locally in a flat byte arrays
type byteChunk struct {
	X, Z  int
	world *World `ignore:"true"`

	Sections [16]*byteChunkSection

	join chan Watcher
}

type byteChunkSection struct {
	Blocks     [16 * 16 * 16]byte
	Data       [(16 * 16 * 16) / 2]byte
	BlockLight [(16 * 16 * 16) / 2]byte
	SkyLight   [(16 * 16 * 16) / 2]byte
}

func (c *byteChunk) Init(world *World, gen Generator) {
	c.world = world
	c.join = make(chan Watcher, 20)
	c.Sections[0] = &byteChunkSection{}
	c.Sections[0].Blocks[0] = 1
	c.Sections[0].Blocks[1] = 1
	go c.run(gen)
}

func (c *byteChunk) run(gen Generator) {
	if gen != nil {
		gen.Generate(c)
	}
	defer c.close()
	for {
		select {
		case watcher := <-c.join:
			watchers := make([]Watcher, 0, 1)
			watchers = append(watchers, watcher)
		getWatchers:
			for {
				select {
				case watcher := <-c.join:
					watchers = append(watchers, watcher)
				default:
					break getWatchers
				}
			}
			data, primaryBitMap := c.genPacketData()
			packet := protocol.ChunkData{
				X: int32(c.X), Z: int32(c.Z),
				GroundUp:       true,
				PrimaryBitMap:  primaryBitMap,
				AddBitMap:      0x0,
				CompressedData: data,
			}
			for _, watcher := range watchers {
				watcher.QueuePacket(packet)
			}
		}
	}
}

func (c *byteChunk) genPacketData() ([]byte, uint16) {
	var buf bytes.Buffer
	var mask uint16
	zl := zlib.NewWriter(&buf)
	for i, sec := range c.Sections {
		if sec == nil {
			continue
		}
		zl.Write(sec.Blocks[:])
		mask |= 1 << uint(i)
	}

	for _, sec := range c.Sections {
		if sec == nil {
			continue
		}
		zl.Write(sec.Data[:])
	}

	for _, sec := range c.Sections {
		if sec == nil {
			continue
		}
		zl.Write(sec.BlockLight[:])
	}

	for _, sec := range c.Sections {
		if sec == nil {
			continue
		}
		zl.Write(sec.SkyLight[:])
	}
	zl.Close()
	return buf.Bytes(), mask
}

func (c *byteChunk) Join(watcher Watcher) {
	c.join <- watcher
}

func (c *byteChunk) close() {

}

func (c *byteChunk) Location() (x, z int) {
	return c.X, c.Z
}
