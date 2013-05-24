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
	"github.com/NetherrackDev/soulsand/command"
	"github.com/NetherrackDev/soulsand/gamemode"
	"github.com/NetherrackDev/soulsand/locale"
	"github.com/NetherrackDev/soulsand/server"
	"log"
	"net"
	"runtime"
	"runtime/debug"
	"testing"
	"time"
)

type Player struct {
	entity.Entity
	event.Source

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
	command.Add("say [string]", func(p soulsand.CommandSender, msg string) {
		system.Broadcast(fmt.Sprintf("["+soulsand.ColourPurple+"Server"+soulsand.ChatReset+"]:"+soulsand.ColourPink+" %s", msg))
	})
	command.Add("gc", func(caller soulsand.CommandSender) {
		runtime.GC()
		debug.FreeOSMemory()
		caller.SendMessageSync("GC Run")
	})
	command.Add("stats", func(caller soulsand.CommandSender) {
		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)
		caller.SendMessageSync(fmt.Sprintf("Sys: %dkb", stats.Sys/1024))
		caller.SendMessageSync(fmt.Sprintf("Alloc: %dkb", stats.Alloc/1024))
		caller.SendMessageSync(fmt.Sprintf("TotalAlloc: %dkb", stats.TotalAlloc/1024))
	})
	command.Add("relight", func(caller soulsand.CommandSender) {
		if player, ok := caller.(*Player); ok {
			player.World.RunSync(int(player.Chunk.X), int(player.Chunk.Z), func(chunk soulsand.SyncChunk) {
				start := time.Now().UnixNano()
				chunk.Relight()
				end := time.Now().UnixNano()
				player.SendMessage(fmt.Sprintf("Time taken: %.3fms", float64(end-start)/1000000.0))
				player.RunSync(func(soulsand.SyncEntity) {
					player.World.GetChunk(player.Chunk.X, player.Chunk.Z, player.ChunkChannel, player.EntityDead)
				})
			})
		} else {
			caller.SendMessageSync("This can only be used by a player")
		}
	})
	command.Add("bench light", func(caller soulsand.CommandSender) {
		if player, ok := caller.(*Player); ok {
			player.World.RunSync(int(player.Chunk.X), int(player.Chunk.Z), func(chunk soulsand.SyncChunk) {
				res := testing.Benchmark(func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						chunk.Relight()
					}
				})
				player.SendMessage(soulsand.ColourCyan + "Benchmark complete")
				player.SendMessage(soulsand.ColourBrightGreen + fmt.Sprintf("AllocsPerOp: %d", res.AllocsPerOp()))
				player.SendMessage(soulsand.ColourBrightGreen + fmt.Sprintf("AllocedBytesPerOp: %d", res.AllocedBytesPerOp()))
				player.SendMessage(soulsand.ColourBrightGreen + fmt.Sprintf("MsPerOp: %.2f", float64(res.NsPerOp())/1000000.0))
			})
		} else {
			caller.SendMessageSync("This can only be used by a player")
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

	if server.GetPlayer(player.name) != nil {
		player.connection.WriteDisconnect(locale.Get(player.GetLocaleSync(), "disconnect.reason.loggedin"))
		runtime.Goexit()
	}
	player.Entity.Init(player)
	defer player.Entity.Finalise()
	defer player.closeConnection()
	player.World = server.GetWorld("main").(internal.World)
	player.World.AddPlayer(player)
	defer player.leaveWorld()
	player.gamemode = server.GetDefaultGamemode()

	spawnX, spawnY, spawnZ := player.World.GetSpawn()

	player.Position.X = float64(spawnX)
	player.Position.Y = float64(spawnY)
	player.Position.Z = float64(spawnZ)
	player.Chunk.LX = int32(player.Position.X) >> 4
	player.Chunk.X = int32(player.Position.X) >> 4
	player.Chunk.LZ = int32(player.Position.Z) >> 4
	player.Chunk.Z = int32(player.Position.Z) >> 4
	player.displayName = player.name
	player.inventory = inventory.CreatePlayerInventory()
	player.inventory.AddWatcher(player)
	defer player.inventory.RemoveWatcher(player)

	player.connection.WriteLoginRequest(player.EID, "flat", int8(player.gamemode), 0, 3, 32)

	eventType, ev := event.NewJoin(player, locale.Get(player.GetLocaleSync(), "disconnect.reason.unknown"))
	if system.EventSource.Fire(eventType, ev) {
		player.connection.WriteDisconnect(ev.Reason)
		runtime.Goexit()
	}

	player.connection.WriteSpawnPosition(int32(spawnX), int32(spawnY), int32(spawnZ))
	player.connection.WritePlayerPositionLook(player.Position.X, player.Position.Y, player.Position.Z, player.Position.Y+1.6, player.Position.Yaw, player.Position.Pitch, false)

	log.Printf("Player \"%s\" logged in with %d", player.name, player.EID)

	system.AddPlayer(player)
	defer system.RemovePlayer(player)
	defer player.Fire(event.NewLeave(player))

	defer player.cleanChunks()

	player.World.JoinChunk(player.Chunk.X, player.Chunk.Z, player)
	player.spawn()
	defer player.despawn()

	go player.dataWatcher()

	player.loop()
}

func (player *Player) loop() {
	timer := time.NewTicker(time.Second / 20)
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
			player.CurrentTick++
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
			log.Println("Error")
			runtime.Goexit()
		}
	}
}

func (p *Player) leaveWorld() {
	p.World.RemovePlayer(p)
}

func (player *Player) cleanChunks() {
	player.World.LeaveChunk(player.Chunk.X, player.Chunk.Z, player)
	vd := int32(player.settings.viewDistance)
	for x := player.Chunk.X - vd; x < player.Chunk.X+vd+1; x++ {
		for z := player.Chunk.Z - vd; z < player.Chunk.Z+vd+1; z++ {
			player.World.LeaveChunkAsWatcher(x, z, player)
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
					p.World.LeaveChunkAsWatcher(x, z, p)
					p.connection.WriteChunkDataUnload(x, z)
				}
			}
		}
		for x := p.Chunk.X - vd; x < p.Chunk.X+vd+1; x++ {
			for z := p.Chunk.Z - vd; z < p.Chunk.Z+vd+1; z++ {
				if x < lx-vd || x >= lx+vd+1 || z < lz-vd || z >= lz+vd+1 {
					p.World.GetChunk(x, z, p.ChunkChannel, p.EntityDead)
					p.World.JoinChunkAsWatcher(x, z, p)
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
				p.World.LeaveChunkAsWatcher(x, z, p)
			}
		}
	}
	for x := p.Chunk.X - int32(p.settings.viewDistance); x < p.Chunk.X+int32(p.settings.viewDistance)+1; x++ {
		for z := p.Chunk.Z - int32(p.settings.viewDistance); z < p.Chunk.Z+int32(p.settings.viewDistance)+1; z++ {
			p.World.GetChunk(x, z, p.ChunkChannel, p.EntityDead)
			p.World.JoinChunkAsWatcher(x, z, p)
		}
	}
}

func (p *Player) spawn() {
	p.World.SendChunkMessage(p.Chunk.X, p.Chunk.Z, p.GetID(), p.CreateSpawn())
}

func (p *Player) despawn() {
	p.World.SendChunkMessage(p.Chunk.X, p.Chunk.Z, p.GetID(), p.CreateDespawn())
}

func (p *Player) dataWatcher() {
	defer func() { p.errorChannel <- struct{}{} }()
	for {
		id := p.connection.ReadUByte()
		p.currentPacketChannel <- id
		<-p.readPacketChannel
	}
}
