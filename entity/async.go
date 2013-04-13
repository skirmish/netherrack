package entity

import (
	"bitbucket.org/Thinkofdeath/soulsand"
)

func (e *Entity) Remove() {
	e.EventChannel <- func(soulsand.SyncEntity) {
		e.kill()
	}
}

func (e *Entity) IsAlive() bool {
	res := make(chan bool, 1)
	e.EventChannel <- func(soulsand.SyncEntity) {
		res <- e.isActive
	}
	return <-res
}

func (e *Entity) SetPosition(x, y, z float64) {
	e.EventChannel <- func(entity soulsand.SyncEntity) {
		e.Position.X, e.Position.Y, e.Position.Z = x, y, z
		e.Position.LastX, e.Position.LastY, e.Position.LastZ = x, y, z
		e.World.SendChunkMessage(e.Chunk.X, e.Chunk.Z, e.GetID(), entityTeleport(e.GetID(), e.Position.X, e.Position.Y, e.Position.Z, e.Position.Yaw, e.Position.Pitch))
		if player, ok := entity.(interface {
			UpdatePosition()
		}); ok {
			player.UpdatePosition()
		}
	}
}

func (e *Entity) GetPosition() (float64, float64, float64) {
	res := make(chan []float64, 1)
	e.EventChannel <- func(soulsand.SyncEntity) {
		pos := make([]float64, 3)
		pos[0] = e.Position.X
		pos[1] = e.Position.Y
		pos[2] = e.Position.Z
		res <- pos
	}
	out := <-res
	return out[0], out[1], out[2]
}

func (e *Entity) SetVelocity(x, y, z float64) {
	e.EventChannel <- func(entity soulsand.SyncEntity) {
		e.Velocity.X, e.Velocity.Y, e.Velocity.Z = x, y, z
	}
}

func (e *Entity) GetVelocity() (float64, float64, float64) {
	res := make(chan []float64, 1)
	e.EventChannel <- func(soulsand.SyncEntity) {
		pos := make([]float64, 3)
		pos[0] = e.Velocity.X
		pos[1] = e.Velocity.Y
		pos[2] = e.Velocity.Z
		res <- pos
	}
	out := <-res
	return out[0], out[1], out[2]
}
