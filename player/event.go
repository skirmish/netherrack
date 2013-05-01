package player

import (
	"github.com/thinkofdeath/netherrack/entity/metadata"
	"github.com/thinkofdeath/soulsand"
)

func (p *Player) CreateSpawn() func(soulsand.SyncEntity) {
	id := p.GetID()
	name := p.GetName()
	x, y, z := p.Position.X, p.Position.Y, p.Position.Z
	yaw, pitch := p.Position.Yaw, p.Position.Pitch
	metadata := (p.GetEntityMetadata().(metadata.Type)).Clone()
	return func(p soulsand.SyncEntity) {
		player := p.(*Player)
		player.connection.WriteSpawnNamedEntity(
			id,
			name,
			int32(x*32),
			int32(y*32),
			int32(z*32),
			int8(int((yaw/180.0)*128)),
			int8((pitch/180.0)*128),
			0,
			metadata)
	}
}

func (p *Player) CreateDespawn() func(soulsand.SyncEntity) {
	id := p.GetID()
	return func(p soulsand.SyncEntity) {
		player := p.(*Player)
		player.connection.WriteDestroyEntity([]int32{id})
	}
}
