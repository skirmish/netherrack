package netherrack

import (
	"github.com/NetherrackDev/netherrack/chunk"
	"github.com/NetherrackDev/netherrack/event"
	"github.com/NetherrackDev/netherrack/network"
	"github.com/NetherrackDev/netherrack/protocol"
	"github.com/NetherrackDev/netherrack/system"
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/gamemode"
	"github.com/NetherrackDev/soulsand/locale"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"runtime/debug"
	"strconv"
)

//Compile time checks
var _ soulsand.Server = &Server{}

func init() {
	log.SetFlags(log.Lshortfile | log.Ltime)
	setDefaultLocaleStrings()
	locale.Load("data/lang")
	server := &Server{}
	server.init()
	soulsand.SetServer(server, &provider{})
}

type Server struct {
	event.Source

	flags        uint64
	ProtoVersion int
	ListPing     struct {
		MessageOfTheDay string
		Version         string
		MaxPlayers      int
	}
	event chan func()

	Config struct {
		Gamemode gamemode.Type
	}
}

func (server *Server) init() {
	server.Source.Init()
	system.EventSource = server.Source
}

func (server *Server) Start(ip string, port int) {
	log.Printf("NumProcs: %d\n", runtime.GOMAXPROCS(-1))
	debug.SetGCPercent(10)
	go func() {
		log.Println(http.ListenAndServe(ip+":25567", nil))
	}()
	log.Println("Starting Netherrack server")

	server.ProtoVersion = 61
	protocol.PROTOVERSION = byte(server.ProtoVersion)
	server.ListPing.Version = "1.5.2"

	server.event = make(chan func(), 1000)

	go server.watcher()
	go network.Listen(ip, port)
}

func (server *Server) watcher() {
	for {
		select {
		case f := <-server.event:
			f()
		}
	}
}

func (server *Server) GetDefaultGamemode() gamemode.Type {
	res := make(chan gamemode.Type, 1)
	server.event <- func() {
		res <- server.Config.Gamemode
	}
	return <-res
}

func (server *Server) SetDefaultGamemode(mode gamemode.Type) {
	server.event <- func() {
		server.Config.Gamemode = mode
	}
}

func (server *Server) GetEntityCount() int {
	return system.GetEntityCount()
}

func (server *Server) GetWorld(name string) soulsand.World {
	return chunk.GetWorld(name)
}

func (server *Server) GetPlayer(name string) soulsand.Player {
	return system.GetPlayer(name)
}

func (server *Server) SetMessageOfTheDay(message string) {
	server.event <- func() {
		server.ListPing.MessageOfTheDay = message
	}
}

func (server *Server) SetMaxPlayers(max int) {
	server.event <- func() {
		server.ListPing.MaxPlayers = max
	}
}

func (server *Server) GetListPingData() []string {
	res := make(chan []string, 1)
	server.event <- func() {
		res <- []string{
			server.ListPing.MessageOfTheDay,
			strconv.Itoa(server.ProtoVersion),
			server.ListPing.Version,
			strconv.Itoa(0),
			strconv.Itoa(server.ListPing.MaxPlayers)}
	}
	return <-res
}

func (server *Server) SetFlag(flag uint64, value bool) {
	if value {
		server.flags |= flag
	} else {
		server.flags &= ^flag
	}
}

func (server *Server) GetFlag(flag uint64) bool {
	return server.flags&flag != 0
}
