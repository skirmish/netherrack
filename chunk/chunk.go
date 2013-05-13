package chunk

import (
	"github.com/NetherrackDev/soulsand/blocks"

	"github.com/NetherrackDev/soulsand"
)

var (
	_ soulsand.SyncChunk = &Chunk{}
)

type Chunk struct {
	World          *World
	X, Z           int32
	SubChunks      []*SubChunk
	Biome          []byte
	Players        map[string]soulsand.Player
	Entitys        map[int32]soulsand.Entity
	requests       chan *ChunkRequest
	watcherJoin    chan *chunkWatcherRequest
	watcherLeave   chan *chunkWatcherRequest
	entityJoin     chan *chunkEntityRequest
	entityLeave    chan *chunkEntityRequest
	messageChannel chan *chunkMessage
	eventChannel   chan func(soulsand.SyncChunk)
	blockQueue     []blockChange
	heightMap      []int32
	lights         map[blockPosition]byte
	needsRelight   bool
}
type SubChunk struct {
	Type       []byte
	MetaData   []byte
	BlockLight []byte
	SkyLight   []byte
	blocks     uint
}

type blockPosition uint32

func createBlockPosition(x, y, z int) blockPosition {
	return blockPosition(uint32(x) | uint32(z)<<4 | uint32(y)<<8)
}

func (bp blockPosition) GetPosition() (x, y, z int) {
	x = int(bp & 0xF)
	z = int((bp >> 4) & 0xF)
	y = int((bp >> 8) & 0xFF)
	return
}

type ChunkPosition struct {
	X, Z int32
}
type blockChange struct {
	X, Y, Z     int
	Block, Meta byte
}
type ChunkRequest struct {
	X, Z int32
	Stop chan struct{}
	Ret  chan [][]byte
}
type chunkEntityRequest struct {
	Pos ChunkPosition
	E   soulsand.Entity
}
type chunkWatcherRequest struct {
	Pos ChunkPosition
	P   soulsand.Player
}
type chunkMessage struct {
	Pos ChunkPosition
	Msg func(soulsand.SyncEntity)
	ID  int32
}
type chunkBlocksRequest struct {
	Pos     ChunkPosition
	X, Y, Z int
	Ret     chan []byte
}
type chunkEvent struct {
	Pos ChunkPosition
	F   func(soulsand.SyncChunk)
}

func (c *Chunk) GetPlayerMap() map[string]soulsand.Player {
	return c.Players
}

func (c *Chunk) AddChange(x, y, z int, block, meta byte) {
	c.blockQueue = append(c.blockQueue, blockChange{x, y, z, block, meta})
}

func (c *Chunk) SetBlock(x, y, z int, blType byte) {
	c.needsRelight = true
	sec := y >> 4
	ind := ((y & 15) << 8) | (z << 4) | x
	section := c.SubChunks[sec]
	if section.Type[ind] == 0 && blType != 0 {
		section.blocks++
		block := blocks.GetBlockById(blType)
		bp := createBlockPosition(x, y, z)
		if light := block.LightLevel(); light != 0 {
			if y+1 > int(c.heightMap[x|z<<4]) {
				c.heightMap[x|z<<4] = int32(y) + 1
			}
			c.lights[bp] = light
		} else if _, ok := c.lights[bp]; ok {
			if y+1 == int(c.heightMap[x|z<<4]) {
				var ty int
				for ty = y - 1; ty >= 0; ty-- {
					if block := blocks.GetBlockById(c.GetBlock(x, ty, z)); block.LightFiltered() != 0 || block.StopsSkylight() {
						c.heightMap[x|z<<4] = int32(ty) + 1
						break
					}
				}
				if ty == 0 {
					c.heightMap[x|z<<4] = 0
				}
			}
			delete(c.lights, bp)
		}
	} else if section.Type[ind] != 0 && blType == 0 {
		section.blocks--
		bp := createBlockPosition(x, y, z)
		if _, ok := c.lights[bp]; ok {
			delete(c.lights, bp)
		}
		if y+1 == int(c.heightMap[x|z<<4]) {
			var ty int
			for ty = y - 1; ty >= 0; ty-- {
				if block := blocks.GetBlockById(c.GetBlock(x, ty, z)); block.LightFiltered() != 0 || block.StopsSkylight() {
					c.heightMap[x|z<<4] = int32(ty) + 1
					break
				}
			}
			if ty == 0 {
				c.heightMap[x|z<<4] = 0
			}
		}
	}
	section.Type[ind] = blType
}

func (c *Chunk) GetBlock(x, y, z int) byte {
	sec := y >> 4
	ind := ((y & 15) << 8) | (z << 4) | x
	return c.SubChunks[sec].Type[ind]
}
func (c *Chunk) SetMeta(x, y, z int, data byte) {
	sec := y >> 4
	i := ((y & 15) << 8) | (z << 4) | x
	section := c.SubChunks[sec]
	if i&1 == 0 {
		section.MetaData[i>>1] &= 0xF0
		section.MetaData[i>>1] |= data & 0xF
	} else {
		section.MetaData[i>>1] &= 0xF
		section.MetaData[i>>1] |= data << 4
	}
}

func (c *Chunk) GetMeta(x, y, z int) byte {
	sec := y >> 4
	i := ((y & 15) << 8) | (z << 4) | x
	if i&1 == 0 {
		return c.SubChunks[sec].MetaData[i>>1] & 0xF
	}
	return c.SubChunks[sec].MetaData[i>>1] >> 4
}

func (c *Chunk) SetBlockLight(x, y, z int, data byte) {
	sec := y >> 4
	section := c.SubChunks[sec]
	i := ((y & 15) << 8) | (z << 4) | x
	idx := i >> 1
	if i&1 == 0 {
		section.BlockLight[idx] &= 0xF0
		section.BlockLight[idx] |= data & 0xF
	} else {
		section.BlockLight[idx] &= 0xF
		section.BlockLight[idx] |= data << 4
	}
}

func (c *Chunk) GetBlockLight(x, y, z int) byte {
	sec := y >> 4
	section := c.SubChunks[sec]
	i := ((y & 15) << 8) | (z << 4) | x
	idx := i >> 1
	if i&1 == 0 {
		return section.BlockLight[idx] & 0xF

	} else {
		return section.BlockLight[idx] >> 4
	}
}

func (c *Chunk) SetSkyLight(x, y, z int, data byte) {
	sec := y >> 4
	section := c.SubChunks[sec]
	i := ((y & 15) << 8) | (z << 4) | x
	idx := i >> 1
	if i&1 == 0 {
		section.SkyLight[idx] &= 0xF0
		section.SkyLight[idx] |= data & 0xF
	} else {
		section.SkyLight[idx] &= 0xF
		section.SkyLight[idx] |= data << 4
	}
}

func (c *Chunk) GetSkyLight(x, y, z int) byte {
	sec := y >> 4
	section := c.SubChunks[sec]
	i := ((y & 15) << 8) | (z << 4) | x
	idx := i >> 1
	if i&1 == 0 {
		return section.SkyLight[idx] & 0xF

	} else {
		return section.SkyLight[idx] >> 4
	}
}

func (c *Chunk) SetBiome(x, z int, biome byte) {
	c.Biome[x|(z<<4)] = biome
}

func CreateChunk(x, z int32) *Chunk {
	chunk := &Chunk{
		X:              x,
		Z:              z,
		SubChunks:      make([]*SubChunk, 16),
		Biome:          make([]byte, 16*16),
		Players:        make(map[string]soulsand.Player),
		Entitys:        make(map[int32]soulsand.Entity),
		requests:       make(chan *ChunkRequest, 500),
		watcherJoin:    make(chan *chunkWatcherRequest, 200),
		watcherLeave:   make(chan *chunkWatcherRequest, 200),
		entityJoin:     make(chan *chunkEntityRequest, 200),
		entityLeave:    make(chan *chunkEntityRequest, 200),
		messageChannel: make(chan *chunkMessage, 1000),
		eventChannel:   make(chan func(soulsand.SyncChunk), 500),
		blockQueue:     make([]blockChange, 0, 3),
		heightMap:      make([]int32, 16*16),
		lights:         make(map[blockPosition]byte),
	}
	for i := 0; i < 16; i++ {
		chunk.SubChunks[i] = CreateSubChunk()
	}
	return chunk
}

func CreateSubChunk() *SubChunk {
	subChunk := &SubChunk{
		Type:       make([]byte, 16*16*16),
		MetaData:   make([]byte, (16*16*16)/2),
		BlockLight: make([]byte, (16*16*16)/2),
		SkyLight:   make([]byte, (16*16*16)/2),
	}
	return subChunk
}

type chunkMessageEvent interface {
	Run(interface{})
	GetEID() int32
}
