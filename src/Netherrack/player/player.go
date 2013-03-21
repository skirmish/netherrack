package player

import (
	"Soulsand/command"
	"Netherrack/entity"
	"Netherrack/event"
	"Soulsand/locale"
	"Soulsand"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"runtime"
	"time"
)

type Player struct {
	entity.Entity

	connection Connection
	name       string

	errorChannel         chan bool
	currentPacketChannel chan byte
	readPacketChannel    chan bool
	ChunkChannel         chan [][]byte
	playerEventChannel   chan func(Soulsand.SyncPlayer)

	currentTickID int32

	Inventory struct {
		CurrentSlot int
	}

	displayName       string
	IgnoreMoveUpdates bool

	experienceBar float32
	level         int16

	gamemode Soulsand.Gamemode

	settings struct {
		locale       string
		viewDistance int
		chatFlags    byte
		difficulty   byte
		showCape     bool
	}
}

func init() {
	command.Add("say $s[]", func(p interface{}, args []interface{}) {
		event.Broadcast(fmt.Sprintf("["+Soulsand.ColourPurple+"Server"+Soulsand.ChatReset+"]:"+Soulsand.ColourPink+" %s", args[0].(string)))
	})
}

func HandlePlayer(conn net.Conn) {

	player := &Player{}
	player.settings.viewDistance = 10
	player.errorChannel = make(chan bool, 1)
	player.currentPacketChannel = make(chan byte)
	player.readPacketChannel = make(chan bool, 2)
	player.playerEventChannel = make(chan func(Soulsand.SyncPlayer), 1000)
	player.ChunkChannel = make(chan [][]byte, 500)
	defer func() {
		player.readPacketChannel <- true
	}()

	player.connection.conn = conn
	player.connection.player = player
	player.connection.Login()
	
	if Soulsand.GetServer().GetPlayer(player.name) != nil {
		player.connection.WriteDisconnect(locale.Get(player.GetLocaleSync(), "disconnect.reason.loggedin"))
		runtime.Goexit()
	}

	player.Init(player)
	defer player.Finalise()
	player.World = Soulsand.GetServer().GetWorld("main")
	player.gamemode = Soulsand.GetServer().GetDefaultGamemode()

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

	event.AddPlayer(player)
	defer event.RemovePlayer(player)

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

	//DEBUG
	player.connection.WriteCreateScoreboard("debug", "Debug Info", false)
	player.connection.WriteDisplayScoreboard(1, "debug")
	var stats runtime.MemStats
	//DEBUG

	go player.dataWatcher()
	defer log.Println("Player disconnecting")

	player.connection.WritePlayerListItem("thinkofdeath", true, 0)

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
			if player.CurrentTick%20 == 0 {
				runtime.ReadMemStats(&stats)
				player.connection.WriteUpdateScore(Soulsand.ColourCyan+"Memory(MB)", "debug", int32(stats.Alloc/1048576))
				player.connection.WriteUpdateScore(Soulsand.ColourCyan+"Goroutines", "debug", int32(runtime.NumGoroutine()))
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
