package player

import (
	"github.com/NetherrackDev/netherrack/entity"
	"github.com/NetherrackDev/netherrack/internal"
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/gamemode"
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
	title := netherInv.GetName()
	useTitle := true
	if len(title) == 0 {
		title = ""
		useTitle = false
	}
	if player.openInventory != nil {
		//Close already open inventory
		player.openInventory.RemoveWatcher(player)
		player.connection.WriteCloseWindow(5)
	}
	netherInv.AddWatcher(player)
	player.openInventory = netherInv
	player.connection.WriteOpenWindow(5, netherInv.GetWindowType(), title, int8(netherInv.GetSize()-27-9), useTitle)
}
