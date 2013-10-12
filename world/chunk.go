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
	"encoding/binary"
	"github.com/NetherrackDev/netherrack/protocol"
	"time"
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

type Entity interface {
	//Returns the entity's UUID
	UUID() string
	//Returns wether the entity is saveable to the chunk
	Saveable() bool
}

//A chunk loaded locally in a flat byte arrays
type Chunk struct {
	X, Z   int
	world  *World
	system System

	Sections  [16]*ChunkSection
	Biome     [16 * 16]byte
	HeightMap [16 * 16]byte
	needsSave bool

	blockChanges []blockChange

	join              chan Watcher
	leave             chan Watcher
	blockPlace        chan blockChange
	blockGet          chan blockGet
	chunkPacket       chan chunkPacket
	entity            chan entityChunk
	entitySpawnUpdate chan packetUpdate

	watchers        map[string]Watcher
	entities        map[string]Entity
	Entities        map[string]Entity
	entitySpawnData map[string]spawnData
	closeChannel    chan chan bool
}

type ChunkSection struct {
	Blocks     [16 * 16 * 16]byte
	Data       [(16 * 16 * 16) / 2]byte
	BlockLight [(16 * 16 * 16) / 2]byte
	SkyLight   [(16 * 16 * 16) / 2]byte
	BlockCount uint
}

type spawnData struct {
	spawn   []protocol.Packet
	despawn []protocol.Packet
}

//Inits the chunk. Should only be called by the world
func (c *Chunk) Init(world *World, gen Generator, system System) {
	c.world = world
	c.system = system
	c.join = make(chan Watcher, 20)
	c.leave = make(chan Watcher, 20)
	c.blockPlace = make(chan blockChange, 50)
	c.blockGet = make(chan blockGet, 50)
	c.chunkPacket = make(chan chunkPacket, 50)
	c.entity = make(chan entityChunk, 50)
	c.entitySpawnUpdate = make(chan packetUpdate, 50)
	c.watchers = make(map[string]Watcher)
	c.entities = make(map[string]Entity)
	c.entitySpawnData = make(map[string]spawnData)
	c.Entities = make(map[string]Entity)
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
	t := time.NewTicker(5 * time.Minute)
	var blockUpdate <-chan time.Time
	defer t.Stop()
	for {
		select {
		case <-blockUpdate:
			blockUpdate = nil
			data := make([]byte, len(c.blockChanges)*4)
			for i, bc := range c.blockChanges {
				binary.BigEndian.PutUint32(data[i*4:],
					(uint32(bc.Data)&0xf)|
						(uint32(bc.Block)<<4)|
						(uint32(bc.Y)<<16)|
						(uint32(bc.Z&0xF)<<24)|
						(uint32(bc.X&0xF)<<28))
			}
			packet := protocol.MultiBlockChange{
				X: protocol.VarInt(c.X), Z: protocol.VarInt(c.Z),
				RecordCount: int16(len(c.blockChanges)),
				Data:        data,
			}
			for _, w := range c.watchers {
				w.QueuePacket(packet)
			}
			c.blockChanges = nil
		case <-t.C:
			if c.needsSave {
				<-c.world.SaveLimiter
				c.needsSave = false
				c.system.SaveChunk(c.X, c.Z, c)
				c.world.SaveLimiter <- struct{}{}
			}
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
				for uuid, e := range c.entitySpawnData {
					if watcher.UUID() != uuid {
						for _, p := range e.spawn {
							watcher.QueuePacket(p)
						}
					}
				}
			}
			c.world.SendLimiter <- zl
		case watcher := <-c.leave:
			delete(c.watchers, watcher.UUID())
			watcher.QueuePacket(protocol.ChunkData{
				X: int32(c.X), Z: int32(c.Z),
				GroundUp:       true,
				CompressedData: []byte{},
			})
			for uuid, e := range c.entitySpawnData {
				if watcher.UUID() != uuid {
					for _, p := range e.despawn {
						watcher.QueuePacket(p)
					}
				}
			}
			if len(c.watchers) == 0 {
				c.world.RequestClose <- c
			}
		case bp := <-c.blockPlace:
			x, z := bp.X&0xF, bp.Z&0xF
			c.SetBlock(x, bp.Y, z, bp.Block)
			c.SetData(x, bp.Y, z, bp.Data)
			c.blockChanges = append(c.blockChanges, bp)
			if blockUpdate == nil {
				blockUpdate = time.After(time.Second / 10) //Allow for multiple block changes to be grouped
			}
		case bg := <-c.blockGet:
			x, z := bg.X&0xF, bg.Z&0xF
			block := c.Block(x, bg.Y, z)
			data := c.Data(x, bg.Y, z)
			bg.Ret <- [2]byte{block, data}
		case packet := <-c.chunkPacket:
			if packet.UUID != "" {
				for _, w := range c.watchers {
					if w.UUID() != packet.UUID {
						w.QueuePacket(packet.Packet)
					}
				}
			} else {
				for _, w := range c.watchers {
					w.QueuePacket(packet.Packet)
				}
			}
		case ec := <-c.entity:
			if ec.Add {
				c.entities[ec.Entity.UUID()] = ec.Entity
				if ec.Entity.Saveable() {
					c.Entities[ec.Entity.UUID()] = ec.Entity
				}
				sData := spawnData{
					spawn:   ec.Spawn,
					despawn: ec.Despawn,
				}
				c.entitySpawnData[ec.Entity.UUID()] = sData
				for _, w := range c.watchers {
					if w.UUID() != ec.Entity.UUID() {
						for _, p := range sData.spawn {
							w.QueuePacket(p)
						}
					}
				}
			} else {
				delete(c.entities, ec.Entity.UUID())
				sData := c.entitySpawnData[ec.Entity.UUID()]
				delete(c.entitySpawnData, ec.Entity.UUID())
				if ec.Entity.Saveable() {
					delete(c.Entities, ec.Entity.UUID())
				}
				for _, w := range c.watchers {
					if w.UUID() != ec.Entity.UUID() {
						for _, p := range sData.despawn {
							w.QueuePacket(p)
						}
					}
				}
			}
		case ret := <-c.closeChannel:
			if len(c.watchers) == 0 && len(c.join) == 0 && !c.needsSave {
				c.system.CloseChunk(c.X, c.Z, c)
				ret <- true
				return
			}
			ret <- false
			if c.needsSave {
				t.Stop()
				t = time.NewTicker(100 * time.Millisecond)
			}
		case update := <-c.entitySpawnUpdate:
			sData := c.entitySpawnData[update.entity.UUID()]
			if update.spawn {
				sData.spawn = update.packets
			} else {
				sData.despawn = update.packets
			}
			c.entitySpawnData[update.entity.UUID()] = sData
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
	c.needsSave = true
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

//Sets the block at the coordinates
func (c *Chunk) SetData(x, y, z int, d byte) {
	section := c.Sections[y>>4]
	if section == nil {
		return
	}
	c.needsSave = true
	idx := (x | (z << 4) | ((y & 0xF) << 8))
	data := section.Data[idx>>1]
	if idx&1 == 0 {
		section.Data[idx>>1] = (data & 0xF0) | (d & 0xF)
		return
	}
	section.Data[idx>>1] = (data & 0xF) | ((d & 0xF) << 4)
}

//Gets the block at the coordinates
func (c *Chunk) Data(x, y, z int) byte {
	section := c.Sections[y>>4]
	if section == nil {
		return 0
	}
	idx := (x | (z << 4) | ((y & 0xF) << 8))
	d := section.Data[idx>>1]
	if idx&1 == 0 {
		return d & 0xF
	}
	return d >> 4
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
	zl.Write(c.Biome[:])
	zl.Flush()
	ret := make([]byte, buf.Len())
	copy(ret, buf.Bytes())
	return ret, mask
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
