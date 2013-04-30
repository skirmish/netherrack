package player

import (
	"github.com/thinkofdeath/netherrack/entity"
	"github.com/thinkofdeath/netherrack/entity/metadata"
	"github.com/thinkofdeath/netherrack/internal"
	"github.com/thinkofdeath/soulsand"
	"github.com/thinkofdeath/soulsand/gamemode"
)

func (player *Player) GetName() string {
	return player.name
}

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

func (player *Player) SetGamemodeSync(mode gamemode.Type) {
	player.gamemode = mode
	player.connection.WriteChangeGameState(3, int8(mode))
}

func (player *Player) GetGamemodeSync() gamemode.Type {
	return player.gamemode
}

func (player *Player) GetConnection() soulsand.UnsafeConnection {
	return player.connection
}

func (player *Player) AsEntity() entity.Entity {
	return player.Entity
}

func (player *Player) OpenInventorySync(inv soulsand.Inventory) {
	netherInv := inv.(internal.Inventory)
	title := inv.GetName()
	useTitle := true
	if len(title) == 0 {
		title = ""
		useTitle = false
	}
	player.connection.WriteOpenWindow(5, netherInv.GetWindowType(), title, int8(netherInv.GetSize()-27-9), useTitle)
}

func (player *Player) SendEntityAttach(eID int32, vID int32) {
	player.connection.WriteAttachEntity(eID, vID)
}

func (player *Player) SendSpawnMob(eID int32, t int8, x, y, z int32, yaw, pitch, hYaw int8, velX, velY, velZ int16, data metadata.Type) {
	player.connection.WriteSpawnMob(eID, t, x, y, z, yaw, pitch, hYaw, velX, velY, velZ, data)
}

func (player *Player) SendEntityTeleport(eID, x, y, z int32, yaw, pitch int8) {
	player.connection.WriteEntityTeleport(eID, x, y, z, yaw, pitch)
}

func (player *Player) SendEntityLookMove(eID int32, dX, dY, dZ int8, yaw, pitch int8) {
	player.connection.WriteEntityLookAndRelativeMove(eID, dX, dY, dZ, yaw, pitch)
}

func (player *Player) SendEntityHeadLook(eID int32, hYaw int8) {
	player.connection.WriteEntityHeadLook(eID, hYaw)
}

func (player *Player) SendEntityDestroy(ids []int32) {
	player.connection.WriteDestroyEntity(ids)
}
