package player

import (
	"bitbucket.org/Thinkofdeath/soulsand"
	"bitbucket.org/Thinkofdeath/soulsand/effect"
	"bitbucket.org/Thinkofdeath/soulsand/gamemode"
)

func (player *Player) PlayEffect(x, y, z int, eff effect.Type, data int, relative bool) error {
	return player.RunSync(func(soulsand.SyncEntity) {
		player.connection.WriteSoundParticleEffect(int32(eff), int32(x), byte(y), int32(z), int32(data), !relative)
	})
}

func (player *Player) SetGamemode(mode gamemode.Type) error {
	return player.RunSync(func(soulsand.SyncEntity) {
		player.gamemode = mode
	})
}

func (player *Player) GetGamemode() (gamemode.Type, error) {
	val, err := player.CallSync(func(e soulsand.SyncEntity, ret chan interface{}) {
		ret <- player.gamemode
	})
	return val.(gamemode.Type), err
}

func (player *Player) GetLocale() (string, error) {
	val, err := player.CallSync(func(e soulsand.SyncEntity, ret chan interface{}) {
		ret <- player.settings.locale
	})
	return val.(string), err
}

func (player *Player) SetExperienceBar(position float32) error {
	return player.RunSync(func(soulsand.SyncEntity) {
		player.experienceBar = position
		player.UpdateExperience()
	})
}

func (player *Player) SetLevel(level int) error {
	return player.RunSync(func(soulsand.SyncEntity) {
		player.level = int16(level)
		player.UpdateExperience()
	})
}

func (player *Player) UpdateExperience() {
	player.connection.WriteSetExperience(player.experienceBar, player.level, 0)
}

func (player *Player) UpdatePosition() {
	player.connection.WritePlayerPositionLook(player.Position.X, player.Position.Y, player.Position.Z, player.Position.Y+1.6, player.Position.Yaw, player.Position.Pitch, false)
}

func (player *Player) GetDisplayName() (string, error) {
	val, err := player.CallSync(func(e soulsand.SyncEntity, ret chan interface{}) {
		ret <- player.displayName
	})
	return val.(string), err
}

func (player *Player) SetDisplayName(name string) error {
	return player.RunSync(func(soulsand.SyncEntity) {
		player.displayName = name
	})
}

func (player *Player) SendMessage(message string) error {
	return player.RunSync(func(soulsand.SyncEntity) {
		player.SendMessageSync(message)
	})
}
