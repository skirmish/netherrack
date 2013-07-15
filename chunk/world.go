package chunk

import (
	"github.com/NetherrackDev/netherrack/nbt"
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/effect"
	"github.com/NetherrackDev/soulsand/sound"
	"runtime"
	"sync"
	"time"
)

//Compile time checks
var _ soulsand.World = &World{}

func init() {
	go worldWatcher()
}

type World struct {
	sync.RWMutex
	Name      string
	chunks    map[ChunkPosition]*Chunk
	players   map[string]soulsand.Player
	settings  nbt.Type
	dataLock  sync.RWMutex
	regions   map[uint64]*region
	generator soulsand.ChunkGenerator
}

func NewWorld(name string) *World {
	world := &World{
		Name:      name,
		players:   make(map[string]soulsand.Player),
		chunks:    make(map[ChunkPosition]*Chunk),
		regions:   make(map[uint64]*region),
		generator: defaultGenerator(0),
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
		<-ticker.C
		world.tick()
	}
}

func (world *World) tick() {
	world.Lock()
	defer world.Unlock()
	time, _ := world.settings.GetLong("Time", 0)
	time++
	world.settings.Set("Time", time)
	dayTime, _ := world.settings.GetLong("DayTime", 0)
	dayTime = (dayTime + 1) % 24000
	world.settings.Set("DayTime", dayTime)
	if time%20 == 0 {
		world.updateTime()
	}
	if time%6000 == 0 {
		world.saveNoLock()
	}
	if len(world.players) == 0 && len(world.chunks) == 0 {
		worldEvent <- func() {
			delete(worlds, world.Name)
		}
		runtime.Goexit()
	}
}

func (world *World) Spawn() (x, y, z int) {
	world.RLock()
	defer world.RUnlock()
	spawnX, _ := world.settings.GetInt("SpawnX", 0)
	spawnY, _ := world.settings.GetInt("SpawnY", 80)
	spawnZ, _ := world.settings.GetInt("SpawnZ", 0)
	return int(spawnX), int(spawnY), int(spawnZ)
}

func (world *World) updateTime() {
	time, _ := world.settings.GetLong("Time", 0)
	dayTime, _ := world.settings.GetLong("DayTime", 0)
	for _, p := range world.players {
		p.RunSync(func(se soulsand.SyncEntity) {
			sp := se.(soulsand.SyncPlayer)
			sp.Connection().WriteTimeUpdate(time, dayTime)
		})
	}
}

func (world *World) AddPlayer(player soulsand.Player) {
	world.Lock()
	defer world.Unlock()
	world.players[player.Name()] = player
}

func (world *World) RemovePlayer(player soulsand.Player) {
	world.Lock()
	defer world.Unlock()
	delete(world.players, player.Name())
}

func (world *World) RunSync(x, z int, f func(soulsand.SyncChunk)) {
	world.RLock()
	chunk := world.getChunk(ChunkPosition{int32(x), int32(z)})
	world.RUnlock()
	chunk.eventChannel <- f
}

func (world *World) PlayEffect(x, y, z int, eff effect.Type, data int, relative bool) {
	world.RunSync(x>>4, z>>4, func(c soulsand.SyncChunk) {
		chunk := c.(*Chunk)
		for _, p := range chunk.Players {
			p.PlayEffect(x, y, z, eff, data, relative)
		}
	})
}

func (world *World) PlaySound(x, y, z float64, name sound.Type, volume float32, pitch int8) {
	world.RunSync(int(x)>>4, int(z)>>4, func(c soulsand.SyncChunk) {
		chunk := c.(*Chunk)
		for _, p := range chunk.Players {
			p.PlaySound(x, y, z, name, volume, pitch)
		}
	})
}

func (world *World) SetBlock(x, y, z int, block, meta byte) {
	if y < 0 || y >= 255 {
		return
	}
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

func (world *World) Block(x, y, z int) (block, meta byte) {
	if y < 0 || y >= 255 {
		return 0, 0
	}
	cx := x >> 4
	cz := z >> 4
	ret := make(chan []byte, 1)
	world.RunSync(cx, cz, func(c soulsand.SyncChunk) {
		ret <- []byte{c.Block(x-cx*16, y, z-cz*16),
			c.Meta(x-cx*16, y, z-cz*16)}
	})
	d := <-ret
	return d[0], d[1]
}

func (world *World) Blocks(x, y, z, w, h, d int) *Blocks {
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
					out.setBlock(bx, by, bz, c.Block(tx-cx*16, ty, tz-cz*16))
					out.setMeta(bx, by, bz, c.Meta(tx-cx*16, ty, tz-cz*16))
					wait.Done()
				})
			}
		}
	}
	wait.Wait()
	return out
}

func (world *World) GetChunkData(x, z int32, ret chan [][]byte, stop chan struct{}) {
	world.RLock()
	chunk := world.getChunk(ChunkPosition{int32(x), int32(z)})
	world.RUnlock()
	chunk.requests <- &ChunkRequest{
		X: x, Z: z,
		Stop: stop,
		Ret:  ret,
	}
}

func (world *World) JoinChunkAsWatcher(x, z int32, pl soulsand.Player) {
	world.RLock()
	pos := ChunkPosition{
		X: x, Z: z,
	}
	chunk := world.getChunk(pos)
	world.RUnlock()
	chunk.watcherJoin <- &chunkWatcherRequest{
		Pos: pos,
		P:   pl,
	}
}

func (world *World) LeaveChunkAsWatcher(x, z int32, pl soulsand.Player) {
	world.RLock()
	pos := ChunkPosition{
		X: x, Z: z,
	}
	chunk := world.getChunk(pos)
	world.RUnlock()
	chunk.watcherLeave <- &chunkWatcherRequest{
		Pos: pos,
		P:   pl,
	}
}

func (world *World) JoinChunk(x, z int32, e soulsand.Entity) {
	world.RLock()
	pos := ChunkPosition{
		X: x, Z: z,
	}
	chunk := world.getChunk(pos)
	world.RUnlock()
	chunk.entityJoin <- &chunkEntityRequest{
		Pos: pos,
		E:   e,
	}
}

func (world *World) LeaveChunk(x, z int32, e soulsand.Entity) {
	world.RLock()
	pos := ChunkPosition{
		X: x, Z: z,
	}
	chunk := world.getChunk(pos)
	world.RUnlock()
	chunk.entityLeave <- &chunkEntityRequest{
		Pos: pos,
		E:   e,
	}
}

func (world *World) SendChunkMessage(x, z, id int32, msg func(soulsand.SyncEntity)) {
	world.RLock()
	pos := ChunkPosition{
		X: x, Z: z,
	}
	chunk := world.getChunk(pos)
	world.RUnlock()
	chunk.messageChannel <- &chunkMessage{
		pos,
		msg,
		id,
	}
}

func (world *World) getChunk(cp ChunkPosition) *Chunk {
	ch, ok := world.chunks[cp]
	if !ok {
		ch = CreateChunk(cp.X, cp.Z)
		ch.World = world
		go chunkController(ch)
		world.chunks[cp] = ch
	}
	return ch
}

func (world *World) grabChunk(x, z int32) *Chunk {
	world.RLock()
	defer world.RUnlock()
	return world.getChunk(ChunkPosition{x, z})
}

func (world *World) chunkExists(x, z int32) bool {
	world.dataLock.RLock()
	r, ok := world.regions[(uint64(x>>5)&0xFFFFFFFF)|uint64(z>>5)<<32]
	world.dataLock.RUnlock()
	if !ok {
		return false
	}
	r.RLock()
	defer r.RUnlock()
	return r.chunkExists(x, z)
}

func (world *World) chunkLoaded(x, z int32) bool {
	world.RLock()
	defer world.RUnlock()
	_, ok := world.chunks[ChunkPosition{x, z}]
	return ok
}

func (world *World) SetTime(time int64) {
	world.Lock()
	defer world.Unlock()
	world.settings.Set("DayTime", time)
}

func (world *World) killChunk(cp ChunkPosition) {
	world.Lock()
	defer world.Unlock()
	delete(world.chunks, cp)
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

func GetWorldCount() int {
	res := make(chan int, 1)
	worldEvent <- func() {
		res <- len(worlds)
	}
	return <-res
}
