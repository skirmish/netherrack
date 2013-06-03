package chunk

import (
	"github.com/NetherrackDev/netherrack/util"
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/blocks"
)

var (
	_ soulsand.SyncChunk = &Chunk{}
)

type Chunk struct {
	World                  *World
	X, Z                   int32
	SubChunks              []*SubChunk
	biome                  []byte
	Players                map[string]soulsand.Player
	Entitys                map[int32]soulsand.Entity
	requests               chan *ChunkRequest
	watcherJoin            chan *chunkWatcherRequest
	watcherLeave           chan *chunkWatcherRequest
	entityJoin             chan *chunkEntityRequest
	entityLeave            chan *chunkEntityRequest
	messageChannel         chan *chunkMessage
	eventChannel           chan func(soulsand.SyncChunk)
	lightChannel           chan chunkLightRequest
	blockQueue             []blockChange
	heightMap              []int32
	needsSave              bool
	pendingLightOperations util.Stack
	brokenLights           util.Queue
}
type SubChunk struct {
	Type        []byte
	MetaData    []byte
	BlockLight  []byte
	SkyLight    []byte
	blocks      int32
	blockLights int32
	skyLights   int32
}

type chunkLightRequest struct {
	x, y, z    byte
	blockLight bool
	ret        chan byte
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
	sec := y >> 4
	section := c.SubChunks[sec]
	if section == nil {
		if blType != 0 {
			section = CreateSubChunk()
			c.SubChunks[sec] = section
		} else {
			return
		}
	}
	ind := ((y & 15) << 8) | (z << 4) | x
	if section.Type[ind] == 0 && blType != 0 {
		section.blocks++
		block := blocks.GetBlockById(blType)
		if y+1 > int(c.heightMap[x|z<<4]) && (block.LightFiltered() != 0 || block.StopsSkylight()) {
			c.heightMap[x|z<<4] = int32(y) + 1
		} else if y+1 == int(c.heightMap[x|z<<4]) {
			var ty int
			for ty = y - 1; ty >= 0; ty-- {
				if block := blocks.GetBlockById(c.Block(x, ty, z)); block.LightFiltered() != 0 || block.StopsSkylight() {
					c.heightMap[x|z<<4] = int32(ty) + 1
					break
				}
			}
			if ty == 0 {
				c.heightMap[x|z<<4] = 0
			}
		}
	} else if section.Type[ind] != 0 && blType == 0 {
		section.blocks--
		c.SetMeta(x, y, z, 0)
		if y+1 == int(c.heightMap[x|z<<4]) {
			var ty int
			for ty = y - 1; ty >= 0; ty-- {
				if block := blocks.GetBlockById(c.Block(x, ty, z)); block.LightFiltered() != 0 || block.StopsSkylight() {
					c.heightMap[x|z<<4] = int32(ty) + 1
					break
				}
			}
			if ty == 0 {
				c.heightMap[x|z<<4] = 0
			}
		}
	}
	if section.Type[ind] != blType {
		oldBlock := blocks.GetBlockById(section.Type[ind])
		block := blocks.GetBlockById(blType)
		bx, by, bz := int8(x), byte(y), int8(z)
		if light := oldBlock.LightLevel(); light != 0 {
			c.pendingLightOperations.Push(blockLightRemove{bx, by, bz})
		} else if oldBlock.Id() != 0 {
			c.pendingLightOperations.Push(blockRemove{bx, by, bz})
		}
		if light := block.LightLevel(); light != 0 {
			c.pendingLightOperations.Push(blockLightAdd{bx, by, bz, light})
		} else if blType != 0 {
			c.pendingLightOperations.Push(blockAdd{bx, by, bz})
		}
		section.Type[ind] = blType
	}

	if section.blocks == 0 && section.skyLights == 0 && section.blockLights == 0 {
		c.SubChunks[sec] = nil
	}
}

func (c *Chunk) Block(x, y, z int) byte {
	sec := y >> 4
	if section := c.SubChunks[sec]; section != nil {
		ind := ((y & 15) << 8) | (z << 4) | x
		return section.Type[ind]
	} else {
		return 0
	}
}
func (c *Chunk) SetMeta(x, y, z int, data byte) {
	sec := y >> 4
	section := c.SubChunks[sec]
	if section == nil {
		if data != 0 {
			section = CreateSubChunk()
			c.SubChunks[sec] = section
		} else {
			return
		}
	}
	i := ((y & 15) << 8) | (z << 4) | x
	if i&1 == 0 {
		section.MetaData[i>>1] &= 0xF0
		section.MetaData[i>>1] |= data & 0xF
	} else {
		section.MetaData[i>>1] &= 0xF
		section.MetaData[i>>1] |= data << 4
	}

	if section.blocks == 0 && section.skyLights == 0 && section.blockLights == 0 {
		c.SubChunks[sec] = nil
	}
}

func (c *Chunk) Meta(x, y, z int) byte {
	sec := y >> 4
	if section := c.SubChunks[sec]; section != nil {
		i := ((y & 15) << 8) | (z << 4) | x
		if i&1 == 0 {
			return section.MetaData[i>>1] & 0xF
		}
		return section.MetaData[i>>1] >> 4
	} else {
		return 0
	}
}

func (c *Chunk) SetBlockLight(x, y, z int, light byte) {
	sec := y >> 4
	section := c.SubChunks[sec]
	if section == nil {
		if light != 0 {
			section = CreateSubChunk()
			c.SubChunks[sec] = section
		} else {
			return
		}
	}
	i := ((y & 15) << 8) | (z << 4) | x
	idx := i >> 1
	if i&1 == 0 {
		orgLight := section.BlockLight[idx] & 0xF
		if orgLight == 0 && light != 0 {
			section.blockLights++
		} else if orgLight != 0 && light == 0 {
			section.blockLights--
		}
		section.BlockLight[idx] &= 0xF0
		section.BlockLight[idx] |= light & 0xF
	} else {
		orgLight := section.BlockLight[idx] >> 4
		if orgLight == 0 && light != 0 {
			section.blockLights++
		} else if orgLight != 0 && light == 0 {
			section.blockLights--
		}
		section.BlockLight[idx] &= 0xF
		section.BlockLight[idx] |= light << 4
	}

	if section.blocks == 0 && section.skyLights == 0 && section.blockLights == 0 {
		c.SubChunks[sec] = nil
	}
}

func (c *Chunk) BlockLight(x, y, z int) byte {
	sec := y >> 4
	if section := c.SubChunks[sec]; section != nil {
		i := ((y & 15) << 8) | (z << 4) | x
		idx := i >> 1
		if i&1 == 0 {
			return section.BlockLight[idx] & 0xF

		} else {
			return section.BlockLight[idx] >> 4
		}
	} else {
		return 0
	}
}

func (c *Chunk) SetSkyLight(x, y, z int, light byte) {
	sec := y >> 4
	section := c.SubChunks[sec]
	if section == nil {
		if light != 15 {
			section = CreateSubChunk()
			c.SubChunks[sec] = section
		} else {
			return
		}
	}
	i := ((y & 15) << 8) | (z << 4) | x
	idx := i >> 1
	if i&1 == 0 {
		orgLight := section.SkyLight[idx] & 0xF
		if orgLight == 15 && light != 15 {
			section.skyLights++
		} else if orgLight != 15 && light == 15 {
			section.skyLights--
		}
		section.SkyLight[idx] &= 0xF0
		section.SkyLight[idx] |= light & 0xF
	} else {
		orgLight := section.SkyLight[idx] >> 4
		if orgLight == 15 && light != 15 {
			section.skyLights++
		} else if orgLight != 15 && light == 15 {
			section.skyLights--
		}
		section.SkyLight[idx] &= 0xF
		section.SkyLight[idx] |= light << 4
	}

	if section.blocks == 0 && section.skyLights == 0 && section.blockLights == 0 {
		c.SubChunks[sec] = nil
	}
}

func (c *Chunk) SkyLight(x, y, z int) byte {
	sec := y >> 4
	if section := c.SubChunks[sec]; section != nil {
		i := ((y & 15) << 8) | (z << 4) | x
		idx := i >> 1
		if i&1 == 0 {
			return section.SkyLight[idx] & 0xF

		} else {
			return section.SkyLight[idx] >> 4
		}
	} else {
		return 15
	}
}

func (c *Chunk) SetBiome(x, z int, biome byte) {
	c.needsSave = true
	c.biome[x|(z<<4)] = biome
}

func (c *Chunk) Biome(x, z int) byte {
	return c.biome[x|(z<<4)]
}

func (c *Chunk) Height(x, z int) int32 {
	return c.heightMap[x|z<<4]
}

func CreateChunk(x, z int32) *Chunk {
	chunk := &Chunk{
		X:                      x,
		Z:                      z,
		SubChunks:              make([]*SubChunk, 16),
		biome:                  make([]byte, 16*16),
		Players:                make(map[string]soulsand.Player),
		Entitys:                make(map[int32]soulsand.Entity),
		requests:               make(chan *ChunkRequest, 500),
		watcherJoin:            make(chan *chunkWatcherRequest, 200),
		watcherLeave:           make(chan *chunkWatcherRequest, 200),
		entityJoin:             make(chan *chunkEntityRequest, 200),
		entityLeave:            make(chan *chunkEntityRequest, 200),
		messageChannel:         make(chan *chunkMessage, 1000),
		eventChannel:           make(chan func(soulsand.SyncChunk), 5000),
		lightChannel:           make(chan chunkLightRequest, 8),
		blockQueue:             make([]blockChange, 0, 3),
		heightMap:              make([]int32, 16*16),
		pendingLightOperations: util.NewStack(),
		brokenLights:           util.NewQueue(),
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
	copy(subChunk.SkyLight, emptySection.SkyLight)
	return subChunk
}

type chunkMessageEvent interface {
	Run(interface{})
	GetEID() int32
}

var emptySection *SubChunk

func init() {
	emptySection = &SubChunk{
		Type:       make([]byte, 16*16*16),
		MetaData:   make([]byte, (16*16*16)/2),
		BlockLight: make([]byte, (16*16*16)/2),
		SkyLight:   make([]byte, (16*16*16)/2),
	}
	for i := 0; i < len(emptySection.SkyLight); i++ {
		emptySection.SkyLight[i] = 0xFF
	}
}
