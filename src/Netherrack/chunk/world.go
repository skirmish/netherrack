package chunk

import (
	"Soulsand"
)

func init() {
	go worldWatcher()
}

type World struct {
	Name                      string
	chunkChannel              chan *ChunkRequest
	chunkJoinChannel          chan *chunkEntityRequest
	chunkLeaveChannel         chan *chunkEntityRequest
	chunkJoinWatcherChannel   chan *chunkWatcherRequest
	chunkLeaveWatcherChannel  chan *chunkWatcherRequest
	chunkMessageChannel       chan *chunkMessage
	chunkBlocksRequestChannel chan *chunkBlocksRequest
	chunkKillChannel          chan *ChunkPosition
	chunks                    map[ChunkPosition]*Chunk
}

func NewWorld() *World {
	world := &World{
		chunkChannel:              make(chan *ChunkRequest, 500),
		chunkJoinChannel:          make(chan *chunkEntityRequest, 500),
		chunkLeaveChannel:         make(chan *chunkEntityRequest, 500),
		chunkJoinWatcherChannel:   make(chan *chunkWatcherRequest, 500),
		chunkLeaveWatcherChannel:  make(chan *chunkWatcherRequest, 500),
		chunkMessageChannel:       make(chan *chunkMessage, 5000),
		chunkBlocksRequestChannel: make(chan *chunkBlocksRequest, 5000),
		chunkKillChannel:          make(chan *ChunkPosition, 500),
		chunks:                    make(map[ChunkPosition]*Chunk),
	}
	go world.chunkWatcher()
	return world
}

func (world *World) GetBlocks(x, y, z, w, h, d int) *Blocks {
	out := &Blocks{
		make([]byte, w*h*d),
		make([]byte, w*h*d),
		w, h, d,
	}
	chans := make([]chan []byte, w*h*d)
	for by := 0; by < h; by++ {
		for bx := 0; bx < w; bx++ {
			for bz := 0; bz < d; bz++ {
				ret := make(chan []byte, 1)
				tx := x + bx
				ty := y + by
				tz := z + bz
				cx := tx >> 4
				cz := tz >> 4
				world.chunkBlocksRequestChannel <- &chunkBlocksRequest{
					ChunkPosition{int32(cx), int32(cz)},
					tx - cx*16,
					ty,
					tz - cz*16,
					ret,
				}
				chans[bx|(bz*w)|(by*w*d)] = ret
			}
		}
	}
	for by := 0; by < h; by++ {
		for bx := 0; bx < w; bx++ {
			for bz := 0; bz < d; bz++ {
				res := <-chans[bx|(bz*w)|(by*w*d)]
				out.setBlock(bx, by, bz, res[0])
				out.setMeta(bx, by, bz, res[1])
			}
		}
	}
	return out
}

func (world *World) GetChunk(x, z int32, ret chan [][]byte) {
	world.chunkChannel <- &ChunkRequest{
		X: x, Z: z,
		Ret: ret,
	}
}

func (world *World) JoinChunkAsWatcher(x, z int32, pl Soulsand.Player) {
	world.chunkJoinWatcherChannel <- &chunkWatcherRequest{
		Pos: ChunkPosition{
			X: x, Z: z,
		},
		P: pl,
	}
}

func (world *World) LeaveChunkAsWatcher(x, z int32, pl Soulsand.Player) {
	world.chunkLeaveWatcherChannel <- &chunkWatcherRequest{
		Pos: ChunkPosition{
			X: x, Z: z,
		},
		P: pl,
	}
}

func (world *World) JoinChunk(x, z int32, e Soulsand.Entity) {
	world.chunkJoinChannel <- &chunkEntityRequest{
		Pos: ChunkPosition{
			X: x, Z: z,
		},
		E: e,
	}
}

func (world *World) LeaveChunk(x, z int32, e Soulsand.Entity) {
	world.chunkLeaveChannel <- &chunkEntityRequest{
		Pos: ChunkPosition{
			X: x, Z: z,
		},
		E: e,
	}
}

func (world *World) SendChunkMessage(x, z, id int32, msg func(Soulsand.SyncPlayer)) {
	world.chunkMessageChannel <- &chunkMessage{
		ChunkPosition{x, z},
		msg,
		id,
	}
}

func (world *World) getChunk(cp ChunkPosition) *Chunk {
	ch, ok := world.chunks[cp]
	if !ok {
		ch = CreateChunk(cp.X, cp.Z)
		ch.World = world
		go chunkControler(ch)
		world.chunks[cp] = ch
	}
	return ch
}

var (
	worldEvent = make(chan func(), 200)
	worlds     = make(map[string]*World)
)

func worldWatcher() {
	for {
		f := <-worldEvent
		f()
	}
}

func GetWorld(name string) *World {
	res := make(chan *World, 1)
	worldEvent <- func() {
		w, ok := worlds[name]
		if !ok {
			w = NewWorld()
			w.Name = name
			worlds[name] = w
		}
		res <- w
	}
	return <-res
}
