package chunk

import (
	"github.com/thinkofdeath/netherrack/nbt"
	"github.com/thinkofdeath/soulsand"
	"github.com/thinkofdeath/soulsand/effect"
	"sync"
	"time"
)

//Compile time checks
var _ soulsand.World = &World{}

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
	worldEventChannel        chan func(soulsand.World)
	chunks                   map[ChunkPosition]*Chunk
	players                  map[string]soulsand.Player
	settings                 nbt.Type
	dataLock                 sync.RWMutex
}

func NewWorld(name string) *World {
	world := &World{
		Name:                     name,
		chunkChannel:             make(chan *ChunkRequest, 500),
		chunkJoinChannel:         make(chan *chunkEntityRequest, 500),
		chunkLeaveChannel:        make(chan *chunkEntityRequest, 500),
		chunkJoinWatcherChannel:  make(chan *chunkWatcherRequest, 500),
		chunkLeaveWatcherChannel: make(chan *chunkWatcherRequest, 500),
		chunkMessageChannel:      make(chan *chunkMessage, 5000),
		chunkKillChannel:         make(chan *ChunkPosition, 500),
		chunkEventChannel:        make(chan chunkEvent, 5000),
		worldEventChannel:        make(chan func(soulsand.World), 200),
		players:                  make(map[string]soulsand.Player),
		chunks:                   make(map[ChunkPosition]*Chunk),
	}
	go world.chunkWatcher()
	return world
}

func (world *World) chunkWatcher() {
	world.loadLevel()
	defer world.save()
	ticker := time.NewTicker(time.Second / 20)
	defer ticker.Stop()
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
		case msg := <-world.chunkEventChannel:
			world.getChunk(msg.Pos).eventChannel <- msg.F
		case f := <-world.worldEventChannel:
			f(world)
		case <-ticker.C:
			time, _ := world.settings.GetLong("Time", 0)
			time++
			world.settings.Set("Time", time)
			dayTime, _ := world.settings.GetLong("DayTime", 0)
			dayTime++
			world.settings.Set("DayTime", dayTime)
			if dayTime%20 == 0 {
				world.updateTime()
			}
			if time%6000 == 0 {
				world.save()
			}
		}

	}
}

func (world *World) GetSpawn() (x, y, z int) {
	ret := make(chan struct{}, 1)
	world.worldEventChannel <- func(soulsand.World) {
		spawnX, _ := world.settings.GetInt("SpawnX", 0)
		spawnY, _ := world.settings.GetInt("SpawnY", 64)
		spawnZ, _ := world.settings.GetInt("SpawnZ", 0)
		x, y, z = int(spawnX), int(spawnY), int(spawnZ)
		ret <- struct{}{}
	}
	<-ret
	return
}

func (world *World) updateTime() {
	time, _ := world.settings.GetLong("Time", 0)
	dayTime, _ := world.settings.GetLong("DayTime", 0)
	for _, p := range world.players {
		p.RunSync(func(se soulsand.SyncEntity) {
			sp := se.(soulsand.SyncPlayer)
			sp.GetConnection().WriteTimeUpdate(time, dayTime)
		})
	}
}

func (world *World) AddPlayer(player soulsand.Player) {
	world.worldEventChannel <- func(soulsand.World) {
		world.players[player.GetName()] = player
	}
}

func (world *World) RemovePlayer(player soulsand.Player) {
	world.worldEventChannel <- func(soulsand.World) {
		delete(world.players, player.GetName())
	}
}

func (world *World) RunSync(x, z int, f func(soulsand.SyncChunk)) {
	world.chunkEventChannel <- chunkEvent{
		Pos: ChunkPosition{int32(x), int32(z)},
		F:   f,
	}
}

func (world *World) PlayEffect(x, y, z int, eff effect.Type, data int, relative bool) {
	world.RunSync(x>>4, z>>4, func(c soulsand.SyncChunk) {
		chunk := c.(*Chunk)
		for _, p := range chunk.Players {
			p.PlayEffect(x, y, z, eff, data, relative)
		}
	})
}

func (world *World) SetBlock(x, y, z int, block, meta byte) {
	cx := x >> 4
	cz := z >> 4
	world.RunSync(cx, cz, func(c soulsand.SyncChunk) {
		rx := x - cx*16
		rz := z - cz*16
		c.SetBlock(rx, y, rz, block)
		c.SetMeta(rx, y, rz, meta)
		c.(*Chunk).AddChange(rx, y, rz, block, meta)
	})
}

func (world *World) GetBlock(x, y, z int) []byte {
	cx := x >> 4
	cz := z >> 4
	ret := make(chan []byte, 1)
	world.RunSync(cx, cz, func(c soulsand.SyncChunk) {
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
				world.RunSync(cx, cz, func(c soulsand.SyncChunk) {
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

func (world *World) GetChunk(x, z int32, ret chan [][]byte, stop chan struct{}) {
	world.chunkChannel <- &ChunkRequest{
		X: x, Z: z,
		Stop: stop,
		Ret:  ret,
	}
}

func (world *World) JoinChunkAsWatcher(x, z int32, pl soulsand.Player) {
	world.chunkJoinWatcherChannel <- &chunkWatcherRequest{
		Pos: ChunkPosition{
			X: x, Z: z,
		},
		P: pl,
	}
}

func (world *World) LeaveChunkAsWatcher(x, z int32, pl soulsand.Player) {
	world.chunkLeaveWatcherChannel <- &chunkWatcherRequest{
		Pos: ChunkPosition{
			X: x, Z: z,
		},
		P: pl,
	}
}

func (world *World) JoinChunk(x, z int32, e soulsand.Entity) {
	world.chunkJoinChannel <- &chunkEntityRequest{
		Pos: ChunkPosition{
			X: x, Z: z,
		},
		E: e,
	}
}

func (world *World) LeaveChunk(x, z int32, e soulsand.Entity) {
	world.chunkLeaveChannel <- &chunkEntityRequest{
		Pos: ChunkPosition{
			X: x, Z: z,
		},
		E: e,
	}
}

func (world *World) SendChunkMessage(x, z, id int32, msg func(soulsand.SyncEntity)) {
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
			w = NewWorld(name)
			worlds[name] = w
		}
		res <- w
	}
	return <-res
}
