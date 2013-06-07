package creeper

import (
	"github.com/NetherrackDev/netherrack/entity"
	"github.com/NetherrackDev/netherrack/entity/metadata"
	"github.com/NetherrackDev/netherrack/internal"
	"github.com/NetherrackDev/netherrack/log"
	"github.com/NetherrackDev/soulsand"
	"runtime"
	"time"
)

var _ soulsand.EntityCreeper = &Type{}
var _ entity.Spawnable = &Type{}
var _ interface {
	Pause()
} = &Type{}

type Type struct {
	entity.Entity
}

func New(x, y, z float64, world soulsand.World) soulsand.EntityCreeper {
	creeper := &Type{}
	creeper.Source.Init()

	creeper.SetWorldSync(world.(internal.World))
	creeper.SetPositionSync(x, y, z)

	go creeper.loop()
	return creeper
}

func (creeper *Type) spawn() {
	creeper.WorldInternal().SendChunkMessage(creeper.Chunk.X, creeper.Chunk.Z, creeper.ID(), creeper.CreateSpawn())
}

func (creeper *Type) despawn() {
	creeper.WorldInternal().SendChunkMessage(creeper.Chunk.X, creeper.Chunk.Z, creeper.ID(), creeper.CreateDespawn())
}

func (creeper *Type) loop() {
	creeper.Entity.Init(creeper)
	defer creeper.Entity.Finalise()
	x, _, z := creeper.PositionSync()
	creeper.Chunk.X, creeper.Chunk.Z = int32(x)>>4, int32(z)>>4
	creeper.WorldInternal().JoinChunk(creeper.Chunk.X, creeper.Chunk.Z, creeper)
	defer creeper.leaveWorld()

	creeper.spawn()
	defer creeper.despawn()

	timer := time.NewTicker(time.Second / 10)
	defer timer.Stop()
	for {
		select {
		case f := <-creeper.EventChannel:
			f(creeper)
		case <-timer.C:
			creeper.CurrentTick++
			creeper.SendMoveUpdate()
		}
	}
}

func (creeper *Type) Pause() {
	log.Println("Paused")
	creeper.RunSync(func(soulsand.SyncEntity) {
		runtime.Goexit()
	})
}

func (creeper *Type) leaveWorld() {
	x, _, z := creeper.PositionSync()
	creeper.Chunk.X, creeper.Chunk.Z = int32(x)>>4, int32(z)>>4
	creeper.WorldInternal().LeaveChunk(creeper.Chunk.X, creeper.Chunk.Z, creeper)
}

func (creeper *Type) CreateSpawn() func(soulsand.SyncEntity) {
	id := creeper.ID()
	metadata := (creeper.EntityMetadata().(metadata.Type)).Clone()
	x, y, z := creeper.PositionSync()
	yaw, pitch := creeper.LookSync()
	velX, velY, velZ := creeper.VelocitySync()
	return func(p soulsand.SyncEntity) {
		player := p.(soulsand.SyncPlayer)
		player.Connection().WriteSpawnMob(id,
			50,
			int32(x*32),
			int32(y*32),
			int32(z*32),
			int8((pitch/180.0)*128),
			int8((pitch/180.0)*128),
			int8((yaw/180.0)*128),
			int16(velX*8000),
			int16(velY*8000),
			int16(velZ*8000),
			metadata)
	}
}

func (creeper *Type) CreateDespawn() func(soulsand.SyncEntity) {
	id := creeper.ID()
	return func(p soulsand.SyncEntity) {
		player := p.(soulsand.SyncPlayer)
		player.Connection().WriteDestroyEntity([]int32{id})
	}
}
