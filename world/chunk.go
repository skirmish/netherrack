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
	"github.com/NetherrackDev/netherrack/protocol"
)

var defaultSkyLight [(16 * 16 * 16) / 2]byte

func init() {
	for i := range defaultSkyLight {
		defaultSkyLight[i] = 0xFF //Full bright
	}
}

type Watcher interface {
	//Queues a packet to be sent to the watcher
	QueuePacket(packet protocol.Packet)
	//Returns the watcher's UUID
	UUID() string
}

//A chunk loaded locally in a flat byte arrays
type Chunk struct {
	X, Z   int
	world  *World `ignore:"true"`
	system System `ignore:"true"`

	Sections  [16]*ChunkSection
	HeightMap [16 * 16]byte

	join  chan Watcher `ignore:"true"`
	leave chan Watcher `ignore:"true"`

	watchers     map[string]Watcher `ignore:"true"`
	closeChannel chan chan bool     `ignore:"true"`
}

type ChunkSection struct {
	Blocks     [16 * 16 * 16]byte
	Data       [(16 * 16 * 16) / 2]byte
	BlockLight [(16 * 16 * 16) / 2]byte
	SkyLight   [(16 * 16 * 16) / 2]byte
	BlockCount uint
}

//Inits the chunk. Should only be called by the world
func (c *Chunk) Init(world *World, gen Generator, system System) {
	c.world = world
	c.system = system
	c.join = make(chan Watcher, 20)
	c.leave = make(chan Watcher, 20)
	c.watchers = make(map[string]Watcher)
	c.closeChannel = make(chan chan bool)
	go c.run(gen)
}

func (c *Chunk) run(gen Generator) {
	if gen != nil {
		gen.Generate(c)
		<-c.world.SaveLimiter
		c.system.SaveChunk(c.X, c.Z, c)
		c.world.SaveLimiter <- struct{}{}
	}
	for {
		select {
		case watcher := <-c.join:
			watchers := make([]Watcher, 0, 1)
			c.watchers[watcher.UUID()] = watcher
			watchers = append(watchers, watcher)
		getWatchers:
			for {
				select {
				case watcher := <-c.join:
					c.watchers[watcher.UUID()] = watcher
					watchers = append(watchers, watcher)
				default:
					break getWatchers
				}
			}
			zl := <-c.world.SendLimiter
			data, primaryBitMap := c.genPacketData(zl)
			packet := protocol.ChunkData{
				X: int32(c.X), Z: int32(c.Z),
				GroundUp:       true,
				PrimaryBitMap:  primaryBitMap,
				CompressedData: data,
			}
			for _, watcher := range watchers {
				watcher.QueuePacket(packet)
			}
			c.world.SendLimiter <- zl
		case watcher := <-c.leave:
			delete(c.watchers, watcher.UUID())
			watcher.QueuePacket(protocol.ChunkData{
				X: int32(c.X), Z: int32(c.Z),
				GroundUp:       true,
				CompressedData: []byte{},
			})
			if len(c.watchers) == 0 {
				c.world.RequestClose <- c
			}
		case ret := <-c.closeChannel:
			if len(c.watchers) == 0 && len(c.join) == 0 {
				ret <- true
				return
			} else {
				ret <- false
			}
		}
	}
}

//Sets the block at the coordinates
func (c *Chunk) SetBlock(x, y, z int, b byte) {
	section := c.Sections[y>>4]
	if section == nil {
		if b == 0 {
			return
		}
		section = &ChunkSection{}
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
func (c *Chunk) Block(x, y, z int) byte {
	section := c.Sections[y>>4]
	if section == nil {
		return 0
	}
	return section.Blocks[x|(z<<4)|((y&0xF)<<8)]
}

func (c *Chunk) genPacketData(cache cachedCompressor) ([]byte, uint16) {
	var mask uint16
	buf, zl := cache.buf, cache.zl
	buf.Reset()
	zl.Reset(buf)
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
	zl.Flush()
	return buf.Bytes(), mask
}

func (c *Chunk) Close() bool {
	ret := make(chan bool, 1)
	c.closeChannel <- ret
	return <-ret
}

//Adds the watcher to the chunk
func (c *Chunk) Join(watcher Watcher) {
	c.join <- watcher
}

//Removes the watcher to the chunk
func (c *Chunk) Leave(watcher Watcher) {
	c.leave <- watcher
}

//Returns the chunk's location
func (c *Chunk) Location() (x, z int) {
	return c.X, c.Z
}
