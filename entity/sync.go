package entity

import (
	"bitbucket.org/Thinkofdeath/soulsand"
	"errors"
)

func (e *Entity) GetPositionSync() (float64, float64, float64) {
	return e.Position.X, e.Position.Y, e.Position.Z
}

func (e *Entity) SetPositionSync(x, y, z float64) {
	e.Position.X, e.Position.Y, e.Position.Z = x, y, z
}

func (e *Entity) RunSync(f func(soulsand.SyncEntity)) error {
	select {
	case e.EventChannel <- f:
	case <-e.EntityDead:
		e.EntityDead <- struct{}{}
		return errors.New("Entity removed")
	}
	return nil
}

func (e *Entity) CallSync(f func(soulsand.SyncEntity, chan interface{})) (interface{}, error) {
	ret := make(chan interface{}, 1)
	err := e.RunSync(func(soulsand.SyncEntity) {
		f(e, ret)
	})
	if err == nil {
		select {
		case val := <-ret:
			return val, err
		case <-e.EntityDead:
			e.EntityDead <- struct{}{}
			return nil, errors.New("Entity removed")
		}
	}
	return nil, err
}
