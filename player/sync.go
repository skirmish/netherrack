package player

import (
	"github.com/NetherrackDev/netherrack/entity"
	"github.com/NetherrackDev/netherrack/internal"
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/chat"
	"github.com/NetherrackDev/soulsand/gamemode"
	"math"
)

func (player *Player) setViewDistance(viewDistance int) {
	player.settings.oldViewDistance = player.settings.viewDistance
	player.settings.viewDistance = int(math.Pow(2, 4-float64(viewDistance)))
	if player.settings.viewDistance > 10 {
		player.settings.viewDistance = 10
	}
}

func (player *Player) Name() string {
	return player.name
}

func (player *Player) HightlightedSlot() int {
	return player.CurrentSlot
}

func (player *Player) ViewDistanceSync() int {
	return player.settings.viewDistance
}

func (player *Player) LocaleSync() string {
	return player.settings.locale
}

func (player *Player) SendMessageSync(msg *chat.Message) {
	player.connection.WriteChatMessage(msg.String())
}

func (player *Player) DisplayNameSync() string {
	return player.displayName
}

func (player *Player) SetDisplayNameSync(name string) {
	player.displayName = name
}

func (player *Player) SetGamemodeSync(mode gamemode.Type) {
	player.gamemode = mode
	player.connection.WriteChangeGameState(3, int8(mode))
}

func (player *Player) GamemodeSync() gamemode.Type {
	return player.gamemode
}

func (player *Player) Connection() soulsand.UnsafeConnection {
	return player.connection
}

func (player *Player) AsEntity() entity.Entity {
	return player.Entity
}

func (player *Player) OpenInventorySync(inv soulsand.Inventory) {
	netherInv := inv.(internal.Inventory)
	title := netherInv.Name()
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
	player.connection.WriteOpenWindow(5, netherInv.WindowType(), title, int8(netherInv.Size()), useTitle, 0)
}

func (player *Player) Inventory() soulsand.PlayerInventory {
	return player.inventory
}
