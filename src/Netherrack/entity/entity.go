package entity

import (
	"Netherrack/system"
	"Soulsand"
)

//Compile time checks
var _ Soulsand.Entity = &Entity{}
var _ Soulsand.SyncEntity = &Entity{}

type Entity struct {
	EID         int32
	CurrentTick uint64

	World Soulsand.World
	Chunk struct {
		X, Z   int32
		LX, LZ int32
	}

	Position struct {
		X, Y, Z             float64
		Yaw, Pitch          float32
		LastX, LastY, LastZ float64
	}
	Velocity struct {
		X, Y, Z float64
	}

	EventChannel chan func(Soulsand.SyncEntity)
	Spawnable

	isActive bool
}

type Spawnable interface {
	CreateSpawn() func(Soulsand.SyncPlayer)
	CreateDespawn() func(Soulsand.SyncPlayer)
}

func (e *Entity) Init(s Spawnable) {
	e.EventChannel = make(chan func(Soulsand.SyncEntity), 500)
	e.Spawnable = s
	e.isActive = true
	e.EID = system.GetFreeEntityID(e)
}

func (e *Entity) Finalise() {
	e.kill()
}

func (e *Entity) kill() {
	if !e.isActive {
		return
	}
	e.isActive = false
	system.FreeEntityID(e)
}

func (e *Entity) GetID() int32 {
	return e.EID
}

func (e *Entity) SendMoveUpdate() (movedChunk bool) {
	e.Position.X += e.Velocity.X
	e.Position.Y += e.Velocity.Y
	e.Position.Z += e.Velocity.Z

	dx := (e.Position.X - e.Position.LastX) * 32
	dy := (e.Position.Y - e.Position.LastY) * 32
	dz := (e.Position.Z - e.Position.LastZ) * 32

	e.Position.LastX = e.Position.X
	e.Position.LastY = e.Position.Y
	e.Position.LastZ = e.Position.Z

	if dx >= 4 || dy >= 4 || dz >= 4 || e.CurrentTick%100 == 0 {
		e.World.SendChunkMessage(e.Chunk.X, e.Chunk.Z, e.GetID(), entityTeleport(e.GetID(), e.Position.X, e.Position.Y, e.Position.Z, e.Position.Yaw, e.Position.Pitch))
	} else {
		e.World.SendChunkMessage(e.Chunk.X, e.Chunk.Z, e.GetID(), entityRelativeLookMove(e.GetID(), int8(dx), int8(dy), int8(dz), int8(int((e.Position.Yaw/180.0)*128)), int8((e.Position.Pitch/180.0)*128)))
	}

	e.Chunk.X = int32(e.Position.X) >> 4
	e.Chunk.Z = int32(e.Position.Z) >> 4

	if e.Chunk.X != e.Chunk.LX || e.Chunk.Z != e.Chunk.LZ {
		e.World.LeaveChunk(e.Chunk.LX, e.Chunk.LZ, e)
		e.World.JoinChunk(e.Chunk.X, e.Chunk.Z, e)
		e.World.SendChunkMessage(e.Chunk.LX, e.Chunk.LZ, e.GetID(), entityTrySpawn(e.Chunk.X, e.Chunk.Z, e.CreateSpawn()))
		e.World.SendChunkMessage(e.Chunk.X, e.Chunk.Z, e.GetID(), entityTryDespawn(e.Chunk.LX, e.Chunk.LZ, e.CreateDespawn()))
		e.Chunk.LX = e.Chunk.X
		e.Chunk.LZ = e.Chunk.Z
		movedChunk = true
	}
	return
}

func entityTrySpawn(cx, cz int32, f func(Soulsand.SyncPlayer)) func(Soulsand.SyncPlayer) {
	return func(p Soulsand.SyncPlayer) {
		player := p.(interface {
			AsEntity() Entity
			GetViewDistanceSync() int
		})
		entity := player.AsEntity()
		vd := int32(player.GetViewDistanceSync())
		if cx < entity.Chunk.X-vd || cx >= entity.Chunk.X+vd+1 || cz < entity.Chunk.Z-vd || cz >= entity.Chunk.Z+vd+1 {
			f(p.(Soulsand.SyncPlayer))
		}
	}
}

func entityTryDespawn(cx, cz int32, f func(Soulsand.SyncPlayer)) func(Soulsand.SyncPlayer) {
	return func(p Soulsand.SyncPlayer) {
		player := p.(interface {
			AsEntity() Entity
			GetViewDistanceSync() int
		})
		entity := player.AsEntity()
		vd := int32(player.GetViewDistanceSync())
		if cx < entity.Chunk.X-vd || cx >= entity.Chunk.X+vd+1 || cz < entity.Chunk.Z-vd || cz >= entity.Chunk.Z+vd+1 {
			f(p.(Soulsand.SyncPlayer))
		}
	}
}

func entityTeleport(id int32, x, y, z float64, yaw, pitch float32) func(Soulsand.SyncPlayer) {
	return func(p Soulsand.SyncPlayer) {
		player := p.(interface {
			SendEntityTeleport(eID, x, y, z int32, yaw, pitch int8)
		})
		player.SendEntityTeleport(
			id,
			int32(x*32),
			int32(y*32),
			int32(z*32),
			int8(int((yaw/180.0)*128)),
			int8((pitch/180.0)*128))
	}
}

func entityRelativeLookMove(id int32, dx, dy, dz, yaw, pitch int8) func(Soulsand.SyncPlayer) {
	return func(p Soulsand.SyncPlayer) {
		player := p.(interface {
			SendEntityLookMove(eID int32, dX, dY, dZ int8, yaw, pitch int8)
			SendEntityHeadLook(eID int32, hYaw int8)
		})
		player.SendEntityLookMove(id, dx, dy, dz, yaw, pitch)
		player.SendEntityHeadLook(id, yaw)
	}
}

type MetadataItem struct {
	Index byte
	Type  byte
	Value interface{}
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

/*type Spawnable interface {
	Spawn(interface{})
	Despawn(interface{})
	GetEID() int32
}*/
