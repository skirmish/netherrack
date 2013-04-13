package player

import ()

func (player *Player) GetViewDistanceSync() int {
	return player.settings.viewDistance
}

func (player *Player) GetLocaleSync() string {
	return player.settings.locale
}

func (player *Player) SendMessageSync(msg string) {
	player.connection.WriteChatMessage(msg)
}

func (player *Player) GetDisplayNameSync() string {
	return player.displayName
}

func (player *Player) SetDisplayNameSync(name string) {
	player.displayName = name
}
