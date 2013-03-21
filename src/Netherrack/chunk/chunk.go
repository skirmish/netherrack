package chunk

import (
	"Soulsand"
)

type (
	Chunk struct {
		World          *World
		X, Z           int32
		SubChunks      []*SubChunk
		Biome          []byte
		Players        map[int32]Soulsand.Player
		Entitys        map[int32]Soulsand.Entity
		requests       chan *ChunkRequest
		watcherJoin    chan *chunkWatcherRequest
		watcherLeave   chan *chunkWatcherRequest
		entityJoin     chan *chunkEntityRequest
		entityLeave    chan *chunkEntityRequest
		messageChannel chan *chunkMessage
		blockRequest   chan *chunkBlocksRequest
	}
	SubChunk struct {
		Type       []byte
		MetaData   []byte
		BlockLight []byte
		SkyLight   []byte
		Blocks     uint
	}
	ChunkPosition struct {
		X, Z int32
	}
	ChunkRequest struct {
		X, Z int32
		Ret  chan [][]byte
	}
	chunkEntityRequest struct {
		Pos ChunkPosition
		E   Soulsand.Entity
	}
	chunkWatcherRequest struct {
		Pos ChunkPosition
		P   Soulsand.Player
	}
	chunkMessage struct {
		Pos ChunkPosition
		Msg func(Soulsand.SyncPlayer)
		ID  int32
	}
	chunkBlocksRequest struct {
		Pos     ChunkPosition
		X, Y, Z int
		Ret     chan []byte
	}
)

var (
/*chunkChannel              chan *ChunkRequest        = make(chan *ChunkRequest, 500)
chunkJoinChannel          chan *chunkEntityRequest  = make(chan *chunkEntityRequest, 500)
chunkLeaveChannel         chan *chunkEntityRequest  = make(chan *chunkEntityRequest, 500)
chunkJoinWatcherChannel   chan *chunkWatcherRequest = make(chan *chunkWatcherRequest, 500)
chunkLeaveWatcherChannel  chan *chunkWatcherRequest = make(chan *chunkWatcherRequest, 500)
chunkMessageChannel       chan *chunkMessage        = make(chan *chunkMessage, 5000)
chunkBlocksRequestChannel chan *chunkBlocksRequest  = make(chan *chunkBlocksRequest, 5000)
chunkKillChannel          chan *ChunkPosition       = make(chan *ChunkPosition, 500)
chunks                    map[ChunkPosition]*Chunk  = make(map[ChunkPosition]*Chunk)*/
)

func (world *World) chunkWatcher() {
	for {
		select {
		case pos := <-world.chunkKillChannel:
			delete(world.chunks, *pos)
		case cr := <-world.chunkChannel:
			cp := ChunkPosition{X: cr.X, Z: cr.Z}
			world.getChunk(cp).requests <- cr
		case msg := <-world.chunkJoinWatcherChannel:
			world.getChunk(msg.Pos).watcherJoin <- msg
		case msg := <-world.chunkLeaveWatcherChannel:
			world.getChunk(msg.Pos).watcherLeave <- msg
		case msg := <-world.chunkJoinChannel:
			world.getChunk(msg.Pos).entityJoin <- msg
		case msg := <-world.chunkLeaveChannel:
			world.getChunk(msg.Pos).entityLeave <- msg
		case msg := <-world.chunkMessageChannel:
			world.getChunk(msg.Pos).messageChannel <- msg
		case msg := <-world.chunkBlocksRequestChannel:
			world.getChunk(msg.Pos).blockRequest <- msg
		}

	}
}

func (c *Chunk) SetBlock(x, y, z int, blType byte) {
	sec := y >> 4
	if c.SubChunks[sec] == nil {
		c.SubChunks[sec] = CreateSubChunk()
	}
	ind := ((y & 15) << 8) | (z << 4) | x
	if c.SubChunks[sec].Type[ind] == 0 && blType != 0 {
		c.SubChunks[sec].Blocks++
	} else if c.SubChunks[sec].Type[ind] != 0 && blType == 0 {
		c.SubChunks[sec].Blocks--
	}
	c.SubChunks[sec].Type[ind] = blType
	if c.SubChunks[sec].Blocks == 0 {
		c.SubChunks[sec] = nil
	}
}

func (c *Chunk) GetBlock(x, y, z int) byte {
	sec := y >> 4
	if sec < 0 || sec >= 16 || c.SubChunks[sec] == nil {
		return 0
	}
	ind := ((y & 15) << 8) | (z << 4) | x
	return c.SubChunks[sec].Type[ind]
}
func (c *Chunk) SetMeta(x, y, z int, data byte) {
	sec := y >> 4
	if c.SubChunks[sec] == nil {
		return
	}
	i := ((y & 15) << 8) | (z << 4) | x
	if i&1 == 0 {
		c.SubChunks[sec].MetaData[i>>1] &= 0xF0
		c.SubChunks[sec].MetaData[i>>1] |= data & 0xF
	} else {
		c.SubChunks[sec].MetaData[i>>1] &= 0xF
		c.SubChunks[sec].MetaData[i>>1] |= (data & 0xF) << 4
	}
}

func (c *Chunk) GetMeta(x, y, z int) byte {
	sec := y >> 4
	if sec < 0 || sec >= 16 || c.SubChunks[sec] == nil {
		return 0
	}
	i := ((y & 15) << 8) | (z << 4) | x
	if i&1 == 0 {
		return c.SubChunks[sec].MetaData[i>>1] & 0xF
	} else {
		return c.SubChunks[sec].MetaData[i>>1] >> 4
	}
	return 0
}

func (c *Chunk) SetBlockLight(x, y, z int, data byte) {
	sec := y >> 4
	if c.SubChunks[sec] == nil {
		return
	}
	i := ((y & 15) << 8) | (z << 4) | x
	if i&1 == 0 {
		c.SubChunks[sec].BlockLight[i>>1] &= 0xF0
		c.SubChunks[sec].BlockLight[i>>1] |= data & 0xF
	} else {
		c.SubChunks[sec].BlockLight[i>>1] &= 0xF
		c.SubChunks[sec].BlockLight[i>>1] |= (data & 0xF) << 4
	}
}
func (c *Chunk) SetSkyLight(x, y, z int, data byte) {
	sec := y >> 4
	if c.SubChunks[sec] == nil {
		return
	}
	i := ((y & 15) << 8) | (z << 4) | x
	if i&1 == 0 {
		c.SubChunks[sec].SkyLight[i>>1] &= 0xF0
		c.SubChunks[sec].SkyLight[i>>1] |= data & 0xF
	} else {
		c.SubChunks[sec].SkyLight[i>>1] &= 0xF
		c.SubChunks[sec].SkyLight[i>>1] |= (data & 0xF) << 4
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
		Biome:          make([]byte, 256),
		Players:        make(map[int32]Soulsand.Player),
		Entitys:        make(map[int32]Soulsand.Entity),
		requests:       make(chan *ChunkRequest, 500),
		watcherJoin:    make(chan *chunkWatcherRequest, 200),
		watcherLeave:   make(chan *chunkWatcherRequest, 200),
		entityJoin:     make(chan *chunkEntityRequest, 200),
		entityLeave:    make(chan *chunkEntityRequest, 200),
		messageChannel: make(chan *chunkMessage, 1000),
		blockRequest:   make(chan *chunkBlocksRequest, 200),
	}
	/*for i := 0; i < 16; i++ {
		chunk.SubChunks[i] = CreateSubChunk()
	}*/
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

/*type entityType interface {
	GetEID() int32
	SpawnFor(chan interface{})
	DespawnFor(chan interface{})
}*/
