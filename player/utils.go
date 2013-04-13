package player

import (
	"bitbucket.org/Thinkofdeath/netherrack/entity"
	"Soulsand"
	"Soulsand/effect"
	"Soulsand/gamemode"
)

func (player *Player) PlayEffect(x, y, z int, eff effect.Type, data int, relative bool) {
	player.RunSync(func(Soulsand.SyncPlayer) {
		player.connection.WriteSoundParticleEffect(int32(eff), int32(x), byte(y), int32(z), int32(data), !relative)
	})
}

func (player *Player) SetGamemode(mode gamemode.Type) {
	player.playerEventChannel <- func(Soulsand.SyncPlayer) {
		player.gamemode = mode
	}
}

func (player *Player) GetGamemode() gamemode.Type {
	res := make(chan gamemode.Type, 1)
	player.playerEventChannel <- func(Soulsand.SyncPlayer) {
		res <- player.gamemode
	}
	return <-res
}

func (player *Player) GetConnection() Soulsand.UnsafeConnection {
	return &player.connection
}

func (player *Player) AsEntity() entity.Entity {
	return player.Entity
}

func (player *Player) RunSync(f func(Soulsand.SyncPlayer)) {
	player.playerEventChannel <- f
}

func (player *Player) GetLocale() string {
	res := make(chan string, 1)
	player.playerEventChannel <- func(Soulsand.SyncPlayer) {
		res <- player.settings.locale
	}
	return <-res
}

func (player *Player) SetExperienceBar(position float32) {
	player.playerEventChannel <- func(Soulsand.SyncPlayer) {
		player.experienceBar = position
		player.UpdateExperience()
	}
}

func (player *Player) SetLevel(level int) {
	player.playerEventChannel <- func(Soulsand.SyncPlayer) {
		player.level = int16(level)
		player.UpdateExperience()
	}
}

func (player *Player) UpdateExperience() {
	player.connection.WriteSetExperience(player.experienceBar, player.level, 0)
}

func (player *Player) UpdatePosition() {
	player.connection.WritePlayerPositionLook(player.Position.X, player.Position.Y, player.Position.Z, player.Position.Y+1.6, player.Position.Yaw, player.Position.Pitch, false)
}

func (player *Player) GetDisplayName() string {
	res := make(chan string, 1)
	player.playerEventChannel <- func(Soulsand.SyncPlayer) {
		res <- player.displayName
	}
	return <-res
}

func (player *Player) SetDisplayName(name string) {
	player.playerEventChannel <- func(Soulsand.SyncPlayer) {
		player.displayName = name
	}
}

func (player *Player) GetName() string {
	return player.name
}

func (player *Player) SendMessage(message string) {
	player.playerEventChannel <- func(Soulsand.SyncPlayer) {
		player.SendMessageSync(message)
	}
}

func (player *Player) SendEntityAttach(eID int32, vID int32) {
	player.connection.WriteAttachEntity(eID, vID)
}

func (player *Player) SendSpawnMob(eID int32, t int8, x, y, z int32, yaw, pitch, hYaw int8, velX, velY, velZ int16, metadata map[byte]entity.MetadataItem) {
	player.connection.WriteSpawnMob(eID, t, x, y, z, yaw, pitch, hYaw, velX, velY, velZ, metadata)
}

func (player *Player) SendEntityTeleport(eID, x, y, z int32, yaw, pitch int8) {
	player.connection.WriteEntityTeleport(eID, x, y, z, yaw, pitch)
}

func (player *Player) SendEntityLookMove(eID int32, dX, dY, dZ int8, yaw, pitch int8) {
	player.connection.WriteEntityLookRelativeMove(eID, dX, dY, dZ, yaw, pitch)
}

func (player *Player) SendEntityHeadLook(eID int32, hYaw int8) {
	player.connection.WriteEntityHeadLook(eID, hYaw)
}

func (player *Player) SendEntityDestroy(ids []int32) {
	player.connection.WriteDestroyEntity(ids)
}
