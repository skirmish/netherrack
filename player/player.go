package player

import (
	"encoding/binary"
	"fmt"
	"github.com/NetherrackDev/netherrack/entity"
	"github.com/NetherrackDev/netherrack/event"
	"github.com/NetherrackDev/netherrack/internal"
	"github.com/NetherrackDev/netherrack/inventory"
	"github.com/NetherrackDev/netherrack/protocol"
	"github.com/NetherrackDev/netherrack/system"
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/chat"
	"github.com/NetherrackDev/soulsand/command"
	"github.com/NetherrackDev/soulsand/gamemode"
	"github.com/NetherrackDev/soulsand/locale"
	"github.com/NetherrackDev/soulsand/log"
	"github.com/NetherrackDev/soulsand/server"
	"net"
	"runtime"
	"runtime/debug"
	"time"
)

type Player struct {
	entity.Entity

	connection *protocol.Conn
	name       string

	errorChannel         chan struct{}
	currentPacketChannel chan byte
	readPacketChannel    chan struct{}
	ChunkChannel         chan [][]byte

	currentTickID int32

	CurrentSlot   int
	inventory     *inventory.PlayerInventory
	openInventory internal.Inventory

	displayName       string
	IgnoreMoveUpdates bool

	experienceBar float32
	level         int16

	gamemode gamemode.Type

	settings struct {
		locale          string
		viewDistance    int
		oldViewDistance int
		chatFlags       byte
		difficulty      byte
		showCape        bool
	}
}

func init() {
	command.Add("time set [int]", func(caller soulsand.CommandSender, time int) {
		if player, ok := caller.(*Player); ok {
			player.WorldInternal().SetTime(int64(time))
		} else {
			caller.SendMessageSync(chat.New().Colour(chat.Red).Text("This can only be used by a player"))
		}
	})
	command.Add("say [string]", func(p soulsand.CommandSender, msg string) {
		system.Broadcast(chat.New().Text("[").
			Colour(chat.DarkPurple).Text("Server").
			Text("]: ").
			Colour(chat.LightPurple).Text(msg))
	})
	command.Add("gc", func(caller soulsand.CommandSender) {
		runtime.GC()
		debug.FreeOSMemory()
		caller.SendMessageSync(chat.New().Text("GC Ran"))
	})
	command.Add("stats", func(caller soulsand.CommandSender) {
		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)
		caller.SendMessageSync(chat.New().Text(fmt.Sprintf("Sys: %dkb", stats.Sys/1024)))
		caller.SendMessageSync(chat.New().Text(fmt.Sprintf("Alloc: %dkb", stats.Alloc/1024)))
		caller.SendMessageSync(chat.New().Text(fmt.Sprintf("TotalAlloc: %dkb", stats.TotalAlloc/1024)))
	})
	command.Add("chunk resend", func(caller soulsand.CommandSender) {
		if player, ok := caller.(*Player); ok {
			player.WorldInternal().GetChunkData(player.Chunk.X, player.Chunk.Z, player.ChunkChannel, player.EntityDead)
		} else {
			caller.SendMessageSync(chat.New().Colour(chat.Red).Text("This can only be used by a player"))
		}
	})
}

//Checks to make sure it matches the API
var (
	_ soulsand.Player     = &Player{}
	_ soulsand.SyncPlayer = &Player{}
)

func finalPlayer(player *Player) {
	log.Println("Player gone")
}

func HandlePlayer(conn net.Conn) {

	player := &Player{}
	runtime.SetFinalizer(player, finalPlayer)
	player.Source.Init()
	player.settings.viewDistance = 0
	player.errorChannel = make(chan struct{}, 1)
	player.currentPacketChannel = make(chan byte)
	player.readPacketChannel = make(chan struct{}, 2)
	player.ChunkChannel = make(chan [][]byte, 500)

	player.connection, player.name = protocol.NewConnection(conn)

	if server.Player(player.name) != nil {
		player.connection.WriteDisconnect(locale.Get(player.LocaleSync(), "disconnect.reason.loggedin"))
		runtime.Goexit()
	}
	player.Entity.Init(player)
	defer player.Entity.Finalise()
	defer player.closeConnection()
	player.SetWorldSync(server.World("main").(internal.World))
	player.WorldInternal().AddPlayer(player)
	defer player.leaveWorld()
	player.gamemode = server.DefaultGamemode()

	spawnX, spawnY, spawnZ := player.WorldInternal().Spawn()

	player.SetPositionSync(float64(spawnX), float64(spawnY), float64(spawnZ))
	x, _, z := player.PositionSync()
	player.Chunk.LX = int32(x) >> 4
	player.Chunk.X = int32(x) >> 4
	player.Chunk.LZ = int32(z) >> 4
	player.Chunk.Z = int32(z) >> 4
	player.displayName = player.name
	player.inventory = inventory.CreatePlayerInventory()
	player.inventory.AddWatcher(player)
	defer player.inventory.RemoveWatcher(player)

	player.connection.WriteLoginRequest(player.EID, "flat", int8(player.gamemode), 0, 3, 32)

	eventType, ev := event.NewJoin(player, locale.Get(player.LocaleSync(), "disconnect.reason.unknown"))
	if system.EventSource.Fire(eventType, ev) {
		player.connection.WriteDisconnect(ev.Reason)
		runtime.Goexit()
	}

	player.connection.WriteSpawnPosition(int32(spawnX), int32(spawnY), int32(spawnZ))
	x, y, z := player.PositionSync()
	yaw, pitch := player.LookSync()
	player.connection.WritePlayerPositionLook(x, y, z, y+1.6, yaw, pitch, false)

	log.Printf("Player \"%s\" logged in with %d", player.name, player.EID)

	system.AddPlayer(player)
	defer system.RemovePlayer(player)
	defer player.Fire(event.NewLeave(player))

	defer player.cleanChunks()

	player.WorldInternal().JoinChunk(player.Chunk.X, player.Chunk.Z, player)
	player.Spawn()
	defer player.Despawn()

	go player.dataWatcher()

	player.loop()
}

func (player *Player) loop() {
	timer := time.NewTicker(time.Second / 10)
	defer timer.Stop()
	for {
		select {
		case chunkData := <-player.ChunkChannel:
			cx := int32(binary.BigEndian.Uint32(chunkData[0][1:5]))
			cz := int32(binary.BigEndian.Uint32(chunkData[0][5:9]))
			vd := int32(player.settings.viewDistance)
			if !(cx < player.Chunk.X-vd || cx >= player.Chunk.X+vd+1 || cz < player.Chunk.Z-vd || cz >= player.Chunk.Z+vd+1) {
				player.connection.Write(chunkData[0])
				player.connection.Write(chunkData[1])
			}
		case f := <-player.EventChannel:
			f(player)
		case <-timer.C:
			player.Tick()
			if player.CurrentTick%100 == 0 {
				player.currentTickID = int32(player.CurrentTick)
				player.connection.WriteKeepAlive(player.currentTickID)
			}
			if player.settings.oldViewDistance != player.settings.viewDistance {
				player.chunkReload(player.settings.oldViewDistance)
				player.settings.oldViewDistance = player.settings.viewDistance
			}
			player.SendMoveUpdate()
		case pId := <-player.currentPacketChannel:
			if pFunc, ok := packets[pId]; ok {
				pFunc(player.connection, player)
			} else {
				log.Printf("Unknown packet 0x%X\n", pId)
				runtime.Goexit()
			}
			player.readPacketChannel <- struct{}{}
		case <-player.errorChannel:
			runtime.Goexit()
		case <-player.EntityDead:
			player.WasKilled = true
			player.EntityDead <- struct{}{}
		}
	}
}

func (p *Player) leaveWorld() {
	p.WorldInternal().RemovePlayer(p)
}

func (player *Player) cleanChunks() {
	player.WorldInternal().LeaveChunk(player.Chunk.X, player.Chunk.Z, player)
	vd := int32(player.settings.viewDistance)
	for x := player.Chunk.X - vd; x < player.Chunk.X+vd+1; x++ {
		for z := player.Chunk.Z - vd; z < player.Chunk.Z+vd+1; z++ {
			player.WorldInternal().LeaveChunkAsWatcher(x, z, player)
		}
	}
	log.Println("Player disconnected")
}

func (player *Player) closeConnection() {
	player.readPacketChannel <- struct{}{}
}

func (p *Player) SendMoveUpdate() {
	lx, lz := p.Chunk.LX, p.Chunk.LZ
	if p.Entity.SendMoveUpdate() {
		vd := int32(p.settings.viewDistance)
		for x := lx - vd; x < lx+vd+1; x++ {
			for z := lz - vd; z < lz+vd+1; z++ {
				if x < p.Chunk.X-vd || x >= p.Chunk.X+vd+1 || z < p.Chunk.Z-vd || z >= p.Chunk.Z+vd+1 {
					p.WorldInternal().LeaveChunkAsWatcher(x, z, p)
					p.connection.WriteChunkDataUnload(x, z)
				}
			}
		}
		for x := p.Chunk.X - vd; x < p.Chunk.X+vd+1; x++ {
			for z := p.Chunk.Z - vd; z < p.Chunk.Z+vd+1; z++ {
				if x < lx-vd || x >= lx+vd+1 || z < lz-vd || z >= lz+vd+1 {
					p.WorldInternal().GetChunkData(x, z, p.ChunkChannel, p.EntityDead)
					p.WorldInternal().JoinChunkAsWatcher(x, z, p)
				}
			}
		}
	}
}

func (p *Player) chunkReload(old int) {
	if old != 0 {
		for x := p.Chunk.X - int32(old); x < p.Chunk.X+int32(old)+1; x++ {
			for z := p.Chunk.Z - int32(old); z < p.Chunk.Z+int32(old)+1; z++ {
				p.connection.WriteChunkDataUnload(x, z)
				p.WorldInternal().LeaveChunkAsWatcher(x, z, p)
			}
		}
	}
	for x := p.Chunk.X - int32(p.settings.viewDistance); x < p.Chunk.X+int32(p.settings.viewDistance)+1; x++ {
		for z := p.Chunk.Z - int32(p.settings.viewDistance); z < p.Chunk.Z+int32(p.settings.viewDistance)+1; z++ {
			p.WorldInternal().GetChunkData(x, z, p.ChunkChannel, p.EntityDead)
			p.WorldInternal().JoinChunkAsWatcher(x, z, p)
		}
	}
}

func (p *Player) dataWatcher() {
	defer func() { p.errorChannel <- struct{}{} }()
	for {
		id := p.connection.ReadUByte()
		p.currentPacketChannel <- id
		<-p.readPacketChannel
	}
}
