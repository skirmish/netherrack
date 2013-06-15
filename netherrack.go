package netherrack

import (
	"github.com/NetherrackDev/netherrack/chunk"
	"github.com/NetherrackDev/netherrack/event"
	"github.com/NetherrackDev/netherrack/network"
	"github.com/NetherrackDev/netherrack/protocol"
	"github.com/NetherrackDev/netherrack/system"
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/chat"
	"github.com/NetherrackDev/soulsand/command"
	"github.com/NetherrackDev/soulsand/gamemode"
	"github.com/NetherrackDev/soulsand/locale"
	"github.com/NetherrackDev/soulsand/log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"
)

//Compile time checks
var _ soulsand.Server = &Server{}

func init() {
	//log.SetFlags(log.Lshortfile | log.Ltime)
	setDefaultLocaleStrings()
	locale.Load("data/lang")
	server := &Server{}
	server.init()
	soulsand.SetServer(server, provider{})

	command.Add("safestop", func(sender soulsand.CommandSender) {
		sender.SendMessageSync(chat.New().Colour(chat.Aqua).Text("Waiting for worlds to empty"))
		//TODO: Kick players and prevent them from joining
		go func() {
			for chunk.GetWorldCount() > 0 {
				time.Sleep(time.Second)
			}
			sender.SendMessageSync(chat.New().Text("Killing server"))
			os.Exit(0)
		}()
	})
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

	command.Parse()

	server.ProtoVersion = 70
	protocol.PROTOVERSION = byte(server.ProtoVersion)
	server.ListPing.Version = "13w24b"

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

func (server *Server) DefaultGamemode() gamemode.Type {
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

func (server *Server) EntityCount() int {
	return system.GetEntityCount()
}

func (server *Server) World(name string) soulsand.World {
	return chunk.GetWorld(name)
}

func (server *Server) Player(name string) soulsand.Player {
	return system.GetPlayer(name)
}

func (server *Server) Players() []soulsand.Player {
	return system.GetPlayers()
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

func (server *Server) ListPingData() []string {
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

func (server *Server) Flag(flag uint64) bool {
	return server.flags&flag != 0
}
