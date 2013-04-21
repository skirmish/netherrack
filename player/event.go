package player

import (
	"github.com/thinkofdeath/netherrack/entity"
	"github.com/thinkofdeath/soulsand"
)

func (p *Player) CreateSpawn() func(soulsand.SyncEntity) {
	id := p.GetID()
	name := p.GetName()
	x, y, z := p.Position.X, p.Position.Y, p.Position.Z
	yaw, pitch := p.Position.Yaw, p.Position.Pitch
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
			map[byte]entity.MetadataItem{
				0: entity.MetadataItem{
					Index: 0,
					Type:  0,
					Value: int8(0),
				},
			})
	}
}

func (p *Player) CreateDespawn() func(soulsand.SyncEntity) {
	id := p.GetID()
	return func(p soulsand.SyncEntity) {
		player := p.(*Player)
		player.SendEntityDestroy([]int32{id})
	}
}
