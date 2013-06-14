package event

import (
	"github.com/NetherrackDev/soulsand"
)

var _ soulsand.EventEntityTick = &EntityTick{}
var _ soulsand.EventEntitySpawnFor = &EntitySpawnFor{}
var _ soulsand.EventEntityDespawnFor = &EntityDespawnFor{}
var _ soulsand.EventEntitySpawn = &EntitySpawn{}
var _ soulsand.EventEntityDespawn = &EntitySpawn{}

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

func (et *EntityTick) Entity() soulsand.SyncEntity {
	return et.entity
}

func (et *EntityTick) Tick() uint64 {
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

func (esf *EntitySpawnFor) Entity() soulsand.Entity {
	return esf.entity
}

func (esf *EntitySpawnFor) Player() soulsand.SyncPlayer {
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

func (esf *EntityDespawnFor) Entity() soulsand.Entity {
	return esf.entity
}

func (esf *EntityDespawnFor) Player() soulsand.SyncPlayer {
	return esf.player
}

type EntitySpawn struct {
	Event

	entity soulsand.SyncEntity
}

func NewEntitySpawn(entity soulsand.SyncEntity) (string, *EntitySpawn) {
	return "EventEntitySpawn", &EntitySpawn{
		entity: entity,
	}
}

func (es *EntitySpawn) Entity() soulsand.SyncEntity {
	return es.entity
}

type EntityDespawn struct {
	Event

	entity soulsand.SyncEntity
}

func NewEntityDespawn(entity soulsand.SyncEntity) (string, *EntityDespawn) {
	return "EventEntityDespawn", &EntityDespawn{
		entity: entity,
	}
}

func (ed *EntityDespawn) Entity() soulsand.SyncEntity {
	return ed.entity
}
