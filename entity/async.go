package entity

import (
	"github.com/thinkofdeath/soulsand"
)

func (e *Entity) Remove() error {
	return e.RunSync(func(soulsand.SyncEntity) {
		e.kill()
	})
}

func (e *Entity) IsAlive() (bool, error) {
	res := make(chan bool, 1)
	e.EventChannel <- func(soulsand.SyncEntity) {
		res <- e.isActive
	}
	val, err := e.CallSync(func(et soulsand.SyncEntity, ret chan interface{}) {
		ret <- e.isActive
	})
	if err == nil {
		return val.(bool), err
	}
	return false, err
}

func (e *Entity) SetPosition(x, y, z float64) error {
	return e.RunSync(func(entity soulsand.SyncEntity) {
		e.Position.X, e.Position.Y, e.Position.Z = x, y, z
		e.Position.LastX, e.Position.LastY, e.Position.LastZ = x, y, z
		e.World.SendChunkMessage(e.Chunk.X, e.Chunk.Z, e.GetID(), entityTeleport(e.GetID(), e.Position.X, e.Position.Y, e.Position.Z, e.Position.Yaw, e.Position.Pitch))
		if player, ok := entity.(interface {
			UpdatePosition()
		}); ok {
			player.UpdatePosition()
		}
	})
}

func (e *Entity) GetPosition() (float64, float64, float64, error) {
	val, err := e.CallSync(func(et soulsand.SyncEntity, ret chan interface{}) {
		ret <- []float64{e.Position.X, e.Position.Y, e.Position.Z}
	})
	if err == nil {
		out := val.([]float64)
		return out[0], out[1], out[2], err
	}
	return 0, 0, 0, err
}

func (e *Entity) SetVelocity(x, y, z float64) error {
	return e.RunSync(func(soulsand.SyncEntity) {
		e.Velocity.X, e.Velocity.Y, e.Velocity.Z = x, y, z
	})
}

func (e *Entity) GetVelocity() (float64, float64, float64, error) {
	val, err := e.CallSync(func(et soulsand.SyncEntity, ret chan interface{}) {
		ret <- []float64{e.Velocity.X, e.Velocity.Y, e.Velocity.Z}
	})
	if err == nil {
		out := val.([]float64)
		return out[0], out[1], out[2], err
	}
	return 0, 0, 0, err
}
