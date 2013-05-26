package player

import (
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/effect"
	"github.com/NetherrackDev/soulsand/gamemode"
	"github.com/NetherrackDev/soulsand/sound"
)

func (player *Player) PlayEffect(x, y, z int, eff effect.Type, data int, relative bool) error {
	return player.RunSync(func(soulsand.SyncEntity) {
		player.connection.WriteSoundOrParticleEffect(int32(eff), int32(x), byte(y), int32(z), int32(data), !relative)
	})
}

func (player *Player) PlaySound(x, y, z float64, name sound.Type, volume float32, pitch int8) error {
	if len(name) == 0 {
		return nil
	}
	return player.RunSync(func(soulsand.SyncEntity) {
		player.connection.WriteNamedSoundEffect(string(name), int32(x*8), int32(y*8), int32(z*8), volume, pitch)
	})
}

func (player *Player) SetGamemode(mode gamemode.Type) error {
	return player.RunSync(func(soulsand.SyncEntity) {
		player.SetGamemodeSync(mode)
	})
}

func (player *Player) Gamemode() (gamemode.Type, error) {
	val, err := player.CallSync(func(e soulsand.SyncEntity, ret chan interface{}) {
		ret <- player.gamemode
	})
	if err == nil {
		return val.(gamemode.Type), err
	}
	return 0, err
}

func (player *Player) Locale() (string, error) {
	val, err := player.CallSync(func(e soulsand.SyncEntity, ret chan interface{}) {
		ret <- player.settings.locale
	})
	if err == nil {
		return val.(string), err
	}
	return "", err
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
	x, y, z := player.PositionSync()
	yaw, pitch := player.LookSync()
	player.connection.WritePlayerPositionLook(x, y, z, y+1.6, yaw, pitch, false)
}

func (player *Player) DisplayName() (string, error) {
	val, err := player.CallSync(func(e soulsand.SyncEntity, ret chan interface{}) {
		ret <- player.displayName
	})
	if err == nil {
		return val.(string), err
	}
	return "", err
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
