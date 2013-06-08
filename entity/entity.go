package entity

import (
	"github.com/NetherrackDev/netherrack/entity/metadata"
	"github.com/NetherrackDev/netherrack/event"
	"github.com/NetherrackDev/netherrack/internal"
	"github.com/NetherrackDev/netherrack/system"
	"github.com/NetherrackDev/soulsand"
)

//Compile time checks
var _ soulsand.Entity = &Entity{}
var _ soulsand.SyncEntity = &Entity{}

type Entity struct {
	event.Source

	EID         int32
	CurrentTick uint64

	world internal.World
	Chunk struct {
		X, Z   int32
		LX, LZ int32
	}

	position struct {
		X, Y, Z             float64
		Yaw, Pitch          float32
		LastX, LastY, LastZ float64
	}
	velocity struct {
		X, Y, Z float64
	}

	EventChannel chan func(soulsand.SyncEntity)
	EntityDead   chan struct{}
	Owner        Spawnable

	metadata metadata.Type

	isActive bool
}

type Spawnable interface {
	soulsand.SyncEntity
	CreateSpawn() func(soulsand.SyncEntity)
	CreateDespawn() func(soulsand.SyncEntity)
}

func (e *Entity) Init(s Spawnable) {
	e.EventChannel = make(chan func(soulsand.SyncEntity), 500)
	e.EntityDead = make(chan struct{}, 1)
	e.Owner = s
	e.isActive = true
	e.EID = system.GetFreeEntityID(e)
	e.metadata = metadata.Type{
		0: int8(0),
	}
}

func (e *Entity) Finalise() {
	e.EntityDead <- struct{}{}
	e.kill()
}

func (e *Entity) Tick() {
	e.CurrentTick++
	e.Fire(event.NewEntityTick(e.Owner, e.CurrentTick))
}

func (e *Entity) kill() {
	if !e.isActive {
		return
	}
	e.Owner = nil
	e.isActive = false
	system.FreeEntityID(e)
}

func (e *Entity) ID() int32 {
	return e.EID
}

func (e *Entity) SendMoveUpdate() (movedChunk bool) {
	e.position.X += e.velocity.X
	e.position.Y += e.velocity.Y
	e.position.Z += e.velocity.Z

	dx := (e.position.X - e.position.LastX) * 32
	dy := (e.position.Y - e.position.LastY) * 32
	dz := (e.position.Z - e.position.LastZ) * 32

	e.position.LastX = e.position.X
	e.position.LastY = e.position.Y
	e.position.LastZ = e.position.Z

	if dx >= 4 || dy >= 4 || dz >= 4 || e.CurrentTick%100 == 0 {
		e.world.SendChunkMessage(e.Chunk.X, e.Chunk.Z, e.ID(), entityTeleport(e.ID(), e.position.X, e.position.Y, e.position.Z, e.position.Yaw, e.position.Pitch))
	} else {
		e.world.SendChunkMessage(e.Chunk.X, e.Chunk.Z, e.ID(), entityRelativeLookMove(e.ID(), int8(dx), int8(dy), int8(dz), int8(int((e.position.Yaw/180.0)*128)), int8((e.position.Pitch/180.0)*128)))
	}

	e.Chunk.X = int32(e.position.X) >> 4
	e.Chunk.Z = int32(e.position.Z) >> 4

	if e.Chunk.X != e.Chunk.LX || e.Chunk.Z != e.Chunk.LZ {
		e.world.LeaveChunk(e.Chunk.LX, e.Chunk.LZ, e.Owner.(soulsand.Entity))
		e.world.JoinChunk(e.Chunk.X, e.Chunk.Z, e.Owner.(soulsand.Entity))
		e.world.SendChunkMessage(e.Chunk.LX, e.Chunk.LZ, e.ID(), entityTryDespawn(e.Chunk.X, e.Chunk.Z, e.Owner.CreateDespawn()))
		e.world.SendChunkMessage(e.Chunk.X, e.Chunk.Z, e.ID(), entityTrySpawn(e.Chunk.LX, e.Chunk.LZ, e.Owner.CreateSpawn()))
		e.Chunk.LX = e.Chunk.X
		e.Chunk.LZ = e.Chunk.Z
		movedChunk = true
	}
	return
}

func (e *Entity) CreateSpawn() func(soulsand.SyncEntity) {
	return e.Owner.CreateSpawn()
}

func (e *Entity) CreateDespawn() func(soulsand.SyncEntity) {
	return e.Owner.CreateDespawn()
}

func entityTrySpawn(cx, cz int32, f func(soulsand.SyncEntity)) func(soulsand.SyncEntity) {
	return func(p soulsand.SyncEntity) {
		player := p.(interface {
			AsEntity() Entity
			soulsand.SyncPlayer
		})
		entity := player.AsEntity()
		vd := int32(player.ViewDistanceSync())
		if cx < entity.Chunk.X-vd || cx >= entity.Chunk.X+vd+1 || cz < entity.Chunk.Z-vd || cz >= entity.Chunk.Z+vd+1 {
			f(p)
		}
	}
}

func entityTryDespawn(cx, cz int32, f func(soulsand.SyncEntity)) func(soulsand.SyncEntity) {
	return func(p soulsand.SyncEntity) {
		player := p.(interface {
			AsEntity() Entity
			soulsand.SyncPlayer
		})
		entity := player.AsEntity()
		vd := int32(player.ViewDistanceSync())
		if cx < entity.Chunk.X-vd || cx >= entity.Chunk.X+vd+1 || cz < entity.Chunk.Z-vd || cz >= entity.Chunk.Z+vd+1 {
			f(p)
		}
	}
}

func entityTeleport(id int32, x, y, z float64, yaw, pitch float32) func(soulsand.SyncEntity) {
	return func(p soulsand.SyncEntity) {
		player := p.(soulsand.SyncPlayer)
		player.Connection().WriteEntityTeleport(
			id,
			int32(x*32),
			int32(y*32),
			int32(z*32),
			int8(int((yaw/180.0)*128)),
			int8((pitch/180.0)*128))
	}
}

func entityRelativeLookMove(id int32, dx, dy, dz, yaw, pitch int8) func(soulsand.SyncEntity) {
	return func(p soulsand.SyncEntity) {
		player := p.(soulsand.SyncPlayer)
		player.Connection().WriteEntityLookAndRelativeMove(id, dx, dy, dz, yaw, pitch)
		player.Connection().WriteEntityHeadLook(id, yaw)
	}
}

func (*Entity) CheckCube(x1, y1, z1, w1, h1, d1, x2, y2, z2, w2, h2, d2 float64) bool {
	if x1 > x2+w2 {
		return false
	}
	if x2 > x1+w1 {
		return false
	}
	if y1 > y2+h2 {
		return false
	}
	if y2 > y1+h1 {
		return false
	}
	if z1 > z2+d2 {
		return false
	}
	if z2 > z1+d1 {
		return false
	}
	return true
}
