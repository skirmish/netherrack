package player

import (
	"github.com/NetherrackDev/netherrack/event"
	"github.com/NetherrackDev/netherrack/log"
	"github.com/NetherrackDev/netherrack/protocol"
	"github.com/NetherrackDev/netherrack/system"
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/blocks"
	"github.com/NetherrackDev/soulsand/command"
	"github.com/NetherrackDev/soulsand/effect"
	"github.com/NetherrackDev/soulsand/gamemode"
	"runtime"
)

var packets map[byte]func(c *protocol.Conn, player *Player) = map[byte]func(c *protocol.Conn, player *Player){
	0x00: func(c *protocol.Conn, player *Player) { //Keep Alive
		id := c.ReadKeepAlive()
		if id != player.currentTickID {
			runtime.Goexit()
		}
	},
	0x03: func(c *protocol.Conn, player *Player) { //Chat Message
		msg := c.ReadChatMessage()
		if len(msg) <= 0 {
			return
		}
		if msg[0] == '/' {
			command.Exec(msg[1:], player)
		} else {
			eventType, ev := event.NewMessage(player, msg)
			if !player.Fire(eventType, ev) {
				system.Broadcast(ev.Message())
			}
		}
	},
	0x07: func(c *protocol.Conn, player *Player) { //Use Entity
		c.ReadUseEntity()
	},
	0x0A: func(c *protocol.Conn, player *Player) { //Player
		c.ReadPlayer()
	},
	0x0B: func(c *protocol.Conn, player *Player) { //Player Position
		x, y, _, z, _ := c.ReadPlayerPosition()
		if !player.IgnoreMoveUpdates {
			player.SetPositionSync(x, y, z)
		}
	},
	0x0C: func(c *protocol.Conn, player *Player) { //Player Look
		yaw, pitch, _ := c.ReadPlayerLook()
		player.SetLookSync(yaw, pitch)
	},
	0x0D: func(c *protocol.Conn, player *Player) { //Player Position and Look
		x, y, _, z, yaw, pitch, _ := c.ReadPlayerPositionLook()
		if !player.IgnoreMoveUpdates {
			player.SetPositionSync(x, y, z)
		}
		player.SetLookSync(yaw, pitch)
	},
	0x0E: func(c *protocol.Conn, player *Player) { //Player Digging
		status, bx, by, bz, face := c.ReadPlayerDigging()
		if status != 2 && !(status == 0 && player.gamemode == gamemode.Creative) {
			return
		}
		x := int(bx)
		y := int(by)
		z := int(bz)
		if !player.Fire(event.NewPlayerBlockBreak(player, x, y, z, face, status)) {
			player.WorldInternal().RunSync(x>>4, z>>4, func(ch soulsand.SyncChunk) {
				chunk := ch.(interface {
					GetPlayerMap() map[string]soulsand.Player
				})
				rx := x - ((x >> 4) << 4)
				rz := z - ((z >> 4) << 4)
				block := ch.Block(rx, y, rz)
				meta := ch.Meta(rx, y, rz)
				m := chunk.GetPlayerMap()
				for _, p := range m {
					if p.Name() != player.Name() {
						p.PlayEffect(x, y, z, effect.BlockBreak, int(block)|(int(meta)<<12), true)
					}
				}
			})
			player.WorldInternal().SetBlock(x, y, z, 0, 0)
		} else {
			data := player.WorldSync().Block(x, y, z)
			bId, bData := data[0], data[1]
			player.connection.WriteBlockChange(bx, by, bz, int16(bId), bData)
		}
	},
	0x0F: func(c *protocol.Conn, player *Player) { //Player Block Placement
		bx, by, bz, direction, _, _, _ := c.ReadPlayerBlockPlacement()
		x := int(bx)
		y := int(by)
		z := int(bz)
		if x == -1 && z == -1 && y == 255 {
			return
		}
		switch direction {
		case 0:
			y--
		case 1:
			y++
		case 2:
			z--
		case 3:
			z++
		case 4:
			x--
		case 5:
			x++
		}
		if item := player.inventory.GetHotbarSlot(player.CurrentSlot); item != nil && item.ID() < 256 {
			id, e := event.NewPlayerBlockPlace(player, x, y, z, item)
			if !player.Fire(id, e) {
				item = e.Block()
				id := byte(item.ID())
				data := byte(item.Data())
				player.WorldInternal().SetBlock(x, y, z, id, data)
				player.PlaySound(float64(x)+0.5, float64(y)+0.5, float64(z)+0.5, blocks.GetBlockById(id).PlacementSound(), 1, 50)
			} else {
				data := player.WorldSync().Block(x, y, z)
				bId, bData := data[0], data[1]
				player.connection.WriteBlockChange(bx, by, bz, int16(bId), bData)
			}
		} else {
			player.Fire(event.NewPlayerRightClick(player))
		}
	},
	0x10: func(c *protocol.Conn, player *Player) { //Held Item Change
		slotID := c.ReadHeldItemChange()
		player.CurrentSlot = int(slotID)
	},
	0x12: func(c *protocol.Conn, player *Player) { //Animation
		_, ani := c.ReadAnimation()
		if ani == 1 {
			player.Fire(event.NewPlayerLeftClick(player))
		}
	},
	0x13: func(c *protocol.Conn, player *Player) { //Entity Action
		_, action := c.ReadEntityAction()
		switch action {
		case 4: //Start sprinting
			player.connection.WriteEntityProperties(player.ID(), map[string]float64{"generic.movementSpeed": 0.13})
		case 5: //Stop sprinting
			player.connection.WriteEntityProperties(player.ID(), map[string]float64{"generic.movementSpeed": 0.1})
		}
	},
	0x1B: func(c *protocol.Conn, player *Player) { //Steer Vehicle (0x1B)
		c.ReadSteerVehicle()
	},
	0x65: func(c *protocol.Conn, player *Player) { //Close Window
		id := c.ReadCloseWindow()
		if id == 5 && player.openInventory != nil {
			player.openInventory.RemoveWatcher(player)
		}
	},
	0x66: func(c *protocol.Conn, player *Player) { //Click Window
		c.ReadClickWindow()
	},
	0x6A: func(c *protocol.Conn, player *Player) { //Confirm Transaction
		c.ReadConfirmTransaction()
	},
	0x6B: func(c *protocol.Conn, player *Player) { //Creative Inventory Action
		slot, item := c.ReadCreativeInventoryAction()
		player.inventory.SetSlot(int(slot), item)
	},
	0x6C: func(c *protocol.Conn, player *Player) { //Enchant Item
		c.ReadEnchantItem()
	},
	0x82: func(c *protocol.Conn, player *Player) { //Update Sign
		c.ReadUpdateSign()
	},
	0xCA: func(c *protocol.Conn, player *Player) { //Player Abilities
		c.ReadPlayerAbilities()
	},
	0xCB: func(c *protocol.Conn, player *Player) { //Tab-complete
		text := c.ReadTabComplete()
		c.WriteTabComplete(command.Complete(text[1:]))
	},
	0xCC: func(c *protocol.Conn, player *Player) { //Client Settings
		locale, viewDistance, chatFlags, difficulty, showCape := c.ReadClientSettings()
		player.settings.locale = locale
		player.setViewDistance(int(viewDistance))
		player.settings.chatFlags = byte(chatFlags)
		player.settings.difficulty = byte(difficulty)
		player.settings.showCape = showCape
	},
	0xCD: func(c *protocol.Conn, player *Player) { //Client Statuses
		c.ReadClientStatuses()
	},
	0xFA: func(c *protocol.Conn, player *Player) { //Plugin Message
		c.ReadPluginMessage()
	},
	0xFF: func(c *protocol.Conn, player *Player) { //Disconnect
		log.Printf("Player %s disconnect %s\n", player.Name(), c.ReadDisconnect())
		runtime.Goexit()
	},
}
