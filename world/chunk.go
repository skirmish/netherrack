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

var defaultSkyLight [(16 * 16 * 16) / 2]byte

func init() {
	for i := range defaultSkyLight {
		defaultSkyLight[i] = 0xFF //Full bright
	}
}

type Chunk interface {
	//Returns the chunk's location
	Location() (x, z int)
	//Inits the chunk. Should only be called by the world
	Init(world *World, gen Generator, system System)
	//Adds the watcher to the chunk
	Join(watcher Watcher)
	//Sets the block at the coordinates
	SetBlock(x, y, z int, b byte)
	//Gets the block at the coordinates
	Block(x, y, z int) byte
}

type Watcher interface {
	//Queues a packet to be sent to the watcher
	QueuePacket(packet protocol.Packet)
}

//A chunk loaded locally in a flat byte arrays
type byteChunk struct {
	X, Z   int
	world  *World `ignore:"true"`
	system System `ignore:"true"`

	Sections  [16]*byteChunkSection
	HeightMap [16 * 16]byte

	join chan Watcher `ignore:"true"`
}

type byteChunkSection struct {
	Blocks     [16 * 16 * 16]byte
	Data       [(16 * 16 * 16) / 2]byte
	BlockLight [(16 * 16 * 16) / 2]byte
	SkyLight   [(16 * 16 * 16) / 2]byte
	BlockCount uint
}

//Inits the chunk. Should only be called by the world
func (c *byteChunk) Init(world *World, gen Generator, system System) {
	c.world = world
	c.system = system
	c.join = make(chan Watcher, 20)
	go c.run(gen)
}

func (c *byteChunk) run(gen Generator) {
	if gen != nil {
		gen.Generate(c)
		c.system.SaveChunk(c.X, c.Z, c)
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

//Sets the block at the coordinates
func (c *byteChunk) SetBlock(x, y, z int, b byte) {
	section := c.Sections[y>>4]
	if section == nil {
		if b == 0 {
			return
		}
		section = &byteChunkSection{}
		copy(section.SkyLight[:], defaultSkyLight[:])
		c.Sections[y>>4] = section
	}
	idx := x | (z << 4) | ((y & 0xF) << 8)
	if section.Blocks[idx] != 0 && b == 0 {
		section.BlockCount--
		if y == int(c.HeightMap[x|(z<<4)]) {
			for i := y - 1; i >= 0; i-- {
				if c.Block(x, i, z) != 0 {
					c.HeightMap[x|(z<<4)] = byte(i)
					break
				}
			}
			if y == int(c.HeightMap[x|(z<<4)]) {
				c.HeightMap[x|(z<<4)] = 0
			}
		}
	} else if section.Blocks[idx] == 0 && b != 0 {
		section.BlockCount++
		if y > int(c.HeightMap[x|(z<<4)]) {
			c.HeightMap[x|(z<<4)] = byte(y)
		}
	}
	section.Blocks[idx] = b
	if section.BlockCount == 0 {
		c.Sections[y>>4] = nil
	}
}

//Gets the block at the coordinates
func (c *byteChunk) Block(x, y, z int) byte {
	section := c.Sections[y>>4]
	if section == nil {
		return 0
	}
	return section.Blocks[x|(z<<4)|((y&0xF)<<8)]
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

//Adds the watcher to the chunk
func (c *byteChunk) Join(watcher Watcher) {
	c.join <- watcher
}

func (c *byteChunk) close() {

}

//Returns the chunk's location
func (c *byteChunk) Location() (x, z int) {
	return c.X, c.Z
}
