package player

import (
	"bitbucket.org/Thinkofdeath/netherrack/entity"
	"bitbucket.org/Thinkofdeath/netherrack/event"
	"bitbucket.org/Thinkofdeath/netherrack/internal"
	"bitbucket.org/Thinkofdeath/netherrack/system"
	"bitbucket.org/Thinkofdeath/soulsand"
	"bitbucket.org/Thinkofdeath/soulsand/command"
	sevent "bitbucket.org/Thinkofdeath/soulsand/event"
	"bitbucket.org/Thinkofdeath/soulsand/gamemode"
	"bitbucket.org/Thinkofdeath/soulsand/locale"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"runtime"
	"time"
)

type Player struct {
	entity.Entity
	event.Source

	connection Connection
	name       string

	errorChannel         chan bool
	currentPacketChannel chan byte
	readPacketChannel    chan bool
	ChunkChannel         chan [][]byte
	playerEventChannel   chan func(soulsand.SyncPlayer)

	currentTickID int32

	Inventory struct {
		CurrentSlot int
	}

	displayName       string
	IgnoreMoveUpdates bool

	experienceBar float32
	level         int16

	gamemode gamemode.Type

	settings struct {
		locale       string
		viewDistance int
		chatFlags    byte
		difficulty   byte
		showCape     bool
	}
}

func init() {
	command.Add("say $s[]", func(p soulsand.CommandSender, msg string) {
		system.Broadcast(fmt.Sprintf("["+soulsand.ColourPurple+"Server"+soulsand.ChatReset+"]:"+soulsand.ColourPink+" %s", msg))
	})
	command.Add("gc", func(caller soulsand.CommandSender) {
		runtime.GC()
		caller.SendMessageSync("GC Run")
	})
}

//Checks to make sure it matches the API
var (
	_ soulsand.Player     = &Player{}
	_ soulsand.SyncPlayer = &Player{}
)

func HandlePlayer(conn net.Conn) {

	player := &Player{}
	player.Source.Init()
	player.settings.viewDistance = 10
	player.errorChannel = make(chan bool, 1)
	player.currentPacketChannel = make(chan byte)
	player.readPacketChannel = make(chan bool, 2)
	player.playerEventChannel = make(chan func(soulsand.SyncPlayer), 1000)
	player.ChunkChannel = make(chan [][]byte, 500)
	defer func() {
		player.readPacketChannel <- true
	}()

	player.connection.conn = conn
	player.connection.player = player
	player.connection.Login()

	if soulsand.GetServer().GetPlayer(player.name) != nil {
		player.connection.WriteDisconnect(locale.Get(player.GetLocaleSync(), "disconnect.reason.loggedin"))
		runtime.Goexit()
	}
	ev := event.NewJoin(player, locale.Get(player.GetLocaleSync(), "disconnect.reason.unknown"))
	if system.EventSource.Fire(sevent.PLAYER_JOIN, ev) {
		player.connection.WriteDisconnect(ev.Reason)
		runtime.Goexit()
	}

	player.Entity.Init(player)
	defer player.Entity.Finalise()
	player.World = soulsand.GetServer().GetWorld("main").(internal.World)
	player.gamemode = soulsand.GetServer().GetDefaultGamemode()

	player.Position.X = 0
	player.Position.Y = 100
	player.Position.Z = 0
	player.Chunk.LX = 0
	player.Chunk.X = 0
	player.Chunk.LZ = 0
	player.Chunk.Z = 0
	player.displayName = player.name

	player.connection.WriteLoginRequest(player.EID, "flat", int8(player.gamemode), 0, 3, 32)
	player.connection.WriteSpawnPosition(0, 100, 0)
	player.connection.WritePlayerPositionLook(player.Position.X, player.Position.Y, player.Position.Z, player.Position.Y+1.6, player.Position.Yaw, player.Position.Pitch, false)

	log.Printf("Player \"%s\" logged in with %d", player.name, player.EID)

	defer func() {
		player.World.LeaveChunk(player.Chunk.X, player.Chunk.Z, player)
		for x := player.Chunk.X - int32(player.settings.viewDistance); x < player.Chunk.X+int32(player.settings.viewDistance)+1; x++ {
			for z := player.Chunk.Z - int32(player.settings.viewDistance); z < player.Chunk.Z+int32(player.settings.viewDistance)+1; z++ {
				player.World.LeaveChunkAsWatcher(x, z, player)
			}
		}
		timer := time.After(10 * time.Second)
	emptyChannels:
		for {
			select {
			case <-player.ChunkChannel:
			case f := <-player.playerEventChannel:
				f(player)
			case f := <-player.EventChannel:
				f(player)
			case <-timer:
				break emptyChannels
			}
		}
		log.Println("Player disconnected")
	}()

	system.AddPlayer(player)
	defer system.RemovePlayer(player)
	defer player.Fire(sevent.PLAYER_LEAVE, event.NewLeave(player))

	vd := int32(player.settings.viewDistance)
	for x := -vd; x < vd+1; x++ {
		for z := -vd; z < vd+1; z++ {
			player.World.GetChunk(int32(x), int32(z), player.ChunkChannel)
			player.World.JoinChunkAsWatcher(int32(x), int32(z), player)
		}
	}
	player.World.JoinChunk(player.Chunk.X, player.Chunk.Z, player)
	player.spawn()
	defer player.despawn()

	go player.dataWatcher()
	defer log.Println("Player disconnecting")

	timer := time.NewTicker(time.Second / 20)
	defer timer.Stop()

	for {
		select {
		case chunkData := <-player.ChunkChannel:
			cx := int32(binary.BigEndian.Uint32(chunkData[0][1:5]))
			cz := int32(binary.BigEndian.Uint32(chunkData[0][5:9]))
			vd := int32(player.settings.viewDistance)
			if !(cx < player.Chunk.X-vd || cx >= player.Chunk.X+vd+1 || cz < player.Chunk.Z-vd || cz >= player.Chunk.Z+vd+1) {
				player.connection.outStream.Write(chunkData[0])
				player.connection.outStream.Write(chunkData[1])
			}
		case f := <-player.playerEventChannel:
			f(player)
		case f := <-player.EventChannel:
			f(player)
		case <-timer.C:
			player.CurrentTick++
			if player.CurrentTick%100 == 0 {
				player.currentTickID = int32(player.CurrentTick)
				player.connection.WriteKeepAlive(player.currentTickID)
			}
			player.SendMoveUpdate()
		case pId := <-player.currentPacketChannel:
			if pFunc, ok := packets[pId]; ok {
				pFunc(&player.connection)
			} else {
				log.Printf("Unknown packet 0x%X\n", pId)
				runtime.Goexit()
			}
			player.readPacketChannel <- true
		case <-player.errorChannel:
			log.Println("Error")
			runtime.Goexit()
		}
	}
}

func (p *Player) SendMoveUpdate() {
	lx, lz := p.Chunk.LX, p.Chunk.LZ
	if p.Entity.SendMoveUpdate() {
		vd := int32(p.settings.viewDistance)
		for x := lx - vd; x < lx+vd+1; x++ {
			for z := lz - vd; z < lz+vd+1; z++ {
				if x < p.Chunk.X-vd || x >= p.Chunk.X+vd+1 || z < p.Chunk.Z-vd || z >= p.Chunk.Z+vd+1 {
					p.connection.WriteChunkDataUnload(x, z)
					p.World.LeaveChunkAsWatcher(x, z, p)
				}
			}
		}
		for x := p.Chunk.X - vd; x < p.Chunk.X+vd+1; x++ {
			for z := p.Chunk.Z - vd; z < p.Chunk.Z+vd+1; z++ {
				if x < lx-vd || x >= lx+vd+1 || z < lz-vd || z >= lz+vd+1 {
					p.World.GetChunk(x, z, p.ChunkChannel)
					p.World.JoinChunkAsWatcher(x, z, p)
				}
			}
		}
	}
}

func (p *Player) chunkReload(old int) {
	for x := p.Chunk.X - int32(old); x < p.Chunk.X+int32(old)+1; x++ {
		for z := p.Chunk.X - int32(old); z < p.Chunk.Z+int32(old)+1; z++ {
			p.connection.WriteChunkDataUnload(x, z)
			p.World.LeaveChunkAsWatcher(x, z, p)
		}
	}
	for x := p.Chunk.X - int32(p.settings.viewDistance); x < p.Chunk.X+int32(p.settings.viewDistance)+1; x++ {
		for z := p.Chunk.X - int32(p.settings.viewDistance); z < p.Chunk.Z+int32(p.settings.viewDistance)+1; z++ {
			p.World.GetChunk(x, z, p.ChunkChannel)
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
	defer func() { p.errorChannel <- true }()
	for {
		id := p.connection.ReadUByte()
		p.currentPacketChannel <- id
		<-p.readPacketChannel
	}
}
