package event

import (
	"github.com/NetherrackDev/soulsand"
)

var _ soulsand.EventEntityTick = &EntityTick{}
var _ soulsand.EventEntitySpawnFor = &EntitySpawnFor{}
var _ soulsand.EventEntityDespawnFor = &EntityDespawnFor{}

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

type EntitySpawnFor struct {
	Event

	entity soulsand.Entity
	player soulsand.SyncPlayer
}

func NewEntitySpawnFor(entity soulsand.Entity, player soulsand.SyncPlayer) (string, *EntitySpawnFor) {
	return "EventEntitySpawnFor", &EntitySpawnFor{
		entity: entity,
		player: player,
	}
}

func (esf *EntitySpawnFor) GetEntity() soulsand.Entity {
	return esf.entity
}

func (esf *EntitySpawnFor) GetPlayer() soulsand.SyncPlayer {
	return esf.player
}

type EntityDespawnFor struct {
	Event

	entity soulsand.Entity
	player soulsand.SyncPlayer
}

func NewEntityDespawnFor(entity soulsand.Entity, player soulsand.SyncPlayer) (string, *EntityDespawnFor) {
	return "EventEntitySpawnFor", &EntityDespawnFor{
		entity: entity,
		player: player,
	}
}

func (esf *EntityDespawnFor) GetEntity() soulsand.Entity {
	return esf.entity
}

func (esf *EntityDespawnFor) GetPlayer() soulsand.SyncPlayer {
	return esf.player
}
