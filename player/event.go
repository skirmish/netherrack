package player

import (
	"github.com/NetherrackDev/netherrack/entity/metadata"
	"github.com/NetherrackDev/soulsand"
)

func (p *Player) CreateSpawn() func(soulsand.SyncEntity) {
	id := p.ID()
	name := p.Name()
	x, y, z := p.PositionSync()
	yaw, pitch := p.LookSync()
	metadata := (p.EntityMetadata().(metadata.Type)).Clone()
	return func(p soulsand.SyncEntity) {
		player := p.(*Player)
		player.connection.WriteSpawnNamedEntity(
			id,
			name,
			int32(x*32),
			int32(y*32),
			int32(z*32),
			int8((yaw/180.0)*128),
			int8((pitch/180.0)*128),
			0,
			metadata)
	}
}

func (p *Player) CreateDespawn() func(soulsand.SyncEntity) {
	id := p.ID()
	return func(p soulsand.SyncEntity) {
		player := p.(*Player)
		player.connection.WriteDestroyEntity([]int32{id})
	}
}
