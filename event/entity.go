package event

import (
	"github.com/NetherrackDev/soulsand"
)

var _ soulsand.EventEntityTick = &EntityTick{}

type EntityTick struct {
	Event

	entity soulsand.SyncEntity
	tick   uint64
}

func NewEntityTick(entity soulsand.SyncEntity, tick uint64) (string, *EntityTick) {
	return "EventEntityTick", &EntityTick{
		entity: entity,
		tick:   tick,
	}
}

func (et *EntityTick) GetEntity() soulsand.SyncEntity {
	return et.entity
}

func (et *EntityTick) GetTick() uint64 {
	return et.tick
}
