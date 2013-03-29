package chunk

import (
	"Soulsand"
	"sync"
)

//Compile time checks
var _ Soulsand.World = &World{}

func init() {
	go worldWatcher()
}

type World struct {
	Name                     string
	chunkChannel             chan *ChunkRequest
	chunkJoinChannel         chan *chunkEntityRequest
	chunkLeaveChannel        chan *chunkEntityRequest
	chunkJoinWatcherChannel  chan *chunkWatcherRequest
	chunkLeaveWatcherChannel chan *chunkWatcherRequest
	chunkMessageChannel      chan *chunkMessage
	chunkKillChannel         chan *ChunkPosition
	chunkEventChannel        chan chunkEvent
	chunks                   map[ChunkPosition]*Chunk
}

func NewWorld() *World {
	world := &World{
		chunkChannel:             make(chan *ChunkRequest, 500),
		chunkJoinChannel:         make(chan *chunkEntityRequest, 500),
		chunkLeaveChannel:        make(chan *chunkEntityRequest, 500),
		chunkJoinWatcherChannel:  make(chan *chunkWatcherRequest, 500),
		chunkLeaveWatcherChannel: make(chan *chunkWatcherRequest, 500),
		chunkMessageChannel:      make(chan *chunkMessage, 5000),
		chunkKillChannel:         make(chan *ChunkPosition, 500),
		chunkEventChannel:        make(chan chunkEvent, 5000),
		chunks:                   make(map[ChunkPosition]*Chunk),
	}
	go world.chunkWatcher()
	return world
}

func (world *World) RunSync(x, z int, f func(Soulsand.SyncChunk)) {
	world.chunkEventChannel <- chunkEvent{
		Pos: ChunkPosition{int32(x), int32(z)},
		F:   f,
	}
}

func (world *World) GetBlock(x, y, z int) []byte {
	cx := x >> 4
	cz := z >> 4
	ret := make(chan []byte, 1)
	world.RunSync(cx, cz, func(c Soulsand.SyncChunk) {
		ret <- []byte{c.GetBlock(x-cx*16, y, z-cz*16),
			c.GetMeta(x-cx*16, y, z-cz*16)}
	})
	return <-ret
}

func (world *World) GetBlocks(x, y, z, w, h, d int) *Blocks {
	out := &Blocks{
		make([]byte, w*h*d),
		make([]byte, w*h*d),
		w, h, d,
	}
	var wait sync.WaitGroup
	for by := 0; by < h; by++ {
		for bx := 0; bx < w; bx++ {
			for bz := 0; bz < d; bz++ {
				tx := x + bx
				ty := y + by
				tz := z + bz
				cx := tx >> 4
				cz := tz >> 4
				wait.Add(1)
				world.RunSync(cx, cz, func(c Soulsand.SyncChunk) {
					out.setBlock(bx, by, bz, c.GetBlock(tx-cx*16, ty, tz-cz*16))
					out.setMeta(bx, by, bz, c.GetMeta(tx-cx*16, ty, tz-cz*16))
					wait.Done()
				})
			}
		}
	}
	wait.Wait()
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
