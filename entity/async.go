package entity

import (
	"github.com/NetherrackDev/soulsand"
)

func (e *Entity) Remove() error {
	return e.RunSync(func(soulsand.SyncEntity) {
		e.RemoveSync()
	})
}

func (e *Entity) Alive() (bool, error) {
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
		e.position.X, e.position.Y, e.position.Z = x, y, z
		e.position.LastX, e.position.LastY, e.position.LastZ = x, y, z
		e.world.SendChunkMessage(e.Chunk.X, e.Chunk.Z, e.ID(), entityTeleport(e.ID(), e.position.X, e.position.Y, e.position.Z, e.position.Yaw, e.position.Pitch))
		if player, ok := entity.(interface {
			UpdatePosition()
		}); ok {
			player.UpdatePosition()
		}
	})
}

func (e *Entity) Position() (float64, float64, float64, error) {
	val, err := e.CallSync(func(et soulsand.SyncEntity, ret chan interface{}) {
		ret <- []float64{e.position.X, e.position.Y, e.position.Z}
	})
	if err == nil {
		out := val.([]float64)
		return out[0], out[1], out[2], err
	}
	return 0, 0, 0, err
}

func (e *Entity) SetVelocity(x, y, z float64) error {
	return e.RunSync(func(soulsand.SyncEntity) {
		e.velocity.X, e.velocity.Y, e.velocity.Z = x, y, z
	})
}

func (e *Entity) Velocity() (float64, float64, float64, error) {
	val, err := e.CallSync(func(et soulsand.SyncEntity, ret chan interface{}) {
		ret <- []float64{e.velocity.X, e.velocity.Y, e.velocity.Z}
	})
	if err == nil {
		out := val.([]float64)
		return out[0], out[1], out[2], err
	}
	return 0, 0, 0, err
}

func (e *Entity) SetWorld(world soulsand.World) error {
	return e.RunSync(func(soulsand.SyncEntity) {
		e.SetWorldSync(world)
	})
}

func (e *Entity) World() (soulsand.World, error) {
	val, err := e.CallSync(func(et soulsand.SyncEntity, ret chan interface{}) {
		ret <- e.WorldSync()
	})
	if err != nil {
		return nil, err
	}
	return val.(soulsand.World), nil
}

func (e *Entity) SetLook(yaw, pitch float32) error {
	return e.RunSync(func(soulsand.SyncEntity) {
		e.SetLookSync(yaw, pitch)
	})
}

func (e *Entity) Look() (float32, float32, error) {
	val, err := e.CallSync(func(et soulsand.SyncEntity, ret chan interface{}) {
		y, p := e.LookSync()
		res := []float32{y, p}
		ret <- res
	})
	if err != nil {
		return 0, 0, err
	}
	res := val.([]float32)
	return res[0], res[1], nil
}
