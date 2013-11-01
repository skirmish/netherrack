/*
   Copyright 2013 Matthew Collins (purggames@gmail.com)

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

/*
	Netherrack is a minecraft server framework for Go (http://golang.org).

	This is still a work in progress and there is still a lot of features
	that are currently missing. The api for this will mostly likely change
	a lot whilst i'm developing this and minecraft changes.
*/
package netherrack

import (
	"encoding/json"
	"github.com/NetherrackDev/netherrack/entity/player"
	"github.com/NetherrackDev/netherrack/message"
	"github.com/NetherrackDev/netherrack/protocol"
	"github.com/NetherrackDev/netherrack/protocol/auth"
	"github.com/NetherrackDev/netherrack/world"
	"log"
	"net"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

const (
	//The currently supported protocol verison
	ProtocolVersion = protocol.Version
	//The currently supported Minecraft version
	MinecraftVersion = "1.7.2"
)

var (
	defaultPingVersion = PingVersion{
		Name:     MinecraftVersion,
		Protocol: ProtocolVersion,
	}
)

//Server is the contains loaded worlds and handles
//network connections from players. The Handler must be set
//before use
type Server struct {
	listener net.Listener
	running  bool

	worlds struct {
		sync.RWMutex
		m        map[string]*world.World
		waitMap  map[string]*sync.WaitGroup
		def      string
		tryClose chan world.TryClose
	}

	authenticator protocol.Authenticator

	Handler ServerHandler

	global struct {
		packet chan protocol.Packet
		add    chan *player.Player
		remove chan *player.Player
	}

	ping struct {
		sync.RWMutex
		data Ping
	}

	playerCount int32
}

//NewServer creates a server which has its internals setup. This must be
//used to create a Server
func NewServer() *Server {
	server := &Server{
		authenticator: auth.Instance,
	}
	server.worlds.m = make(map[string]*world.World)
	server.worlds.waitMap = make(map[string]*sync.WaitGroup)
	server.worlds.tryClose = make(chan world.TryClose, 2)
	server.global.packet = make(chan protocol.Packet, 200)
	server.global.add = make(chan *player.Player, 20)
	server.global.remove = make(chan *player.Player, 20)
	return server
}

//Start starts the minecraft server. This will block while the server is running
//until the server stops
func (server *Server) Start(address string) error {
	server.running = true
	log.Printf("NumProcs: %d\n", runtime.GOMAXPROCS(-1))
	debug.SetGCPercent(10)

	log.Println("Starting Netherrack server")

	listen, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	server.listener = listen

	go server.globalServer()
	go server.worldServer()
	for {
		conn, err := listen.Accept()
		if err != nil {
			return err
		}
		go server.handleConnection(conn)
	}
}

//Handles unloading worlds
func (server *Server) worldServer() {
	for {
		tc := <-server.worlds.tryClose
		//A world is trying to close, so we lock the server's
		//world handling so new players can't try and join whilist
		//its saving
		server.worlds.Lock()
		tc.Ret <- struct{}{}
		//If a player managed to join before the lock was obtained
		//then the world will return false on this channel and
		//stay loaded
		if <-tc.Done {
			log.Println("World closed " + tc.World.Name)
			delete(server.worlds.m, tc.World.Name)
			delete(server.worlds.waitMap, tc.World.Name)
		}
		server.worlds.Unlock()
	}
}

//Handles sending packets to all players on the server
func (server *Server) globalServer() {
	players := map[string]*player.Player{}
	for {
		select {
		case packet := <-server.global.packet:
			for _, p := range players {
				p.QueuePacket(packet)
			}
		case p := <-server.global.add:
			players[p.UUID()] = p
		case p := <-server.global.remove:
			players[p.UUID()] = p
		}
	}
}

//Addr returns the address the server is currently listening on
//once started.
func (server *Server) Addr() net.Addr {
	return server.listener.Addr()
}

//SetAuthenticator changes the authenticator usered by the server. This panics
//if the server is started.
func (server *Server) SetAuthenticator(auth protocol.Authenticator) {
	if server.running {
		panic("Server is running")
	}
	server.authenticator = auth
}

//SetDefaultWorld sets the default world for the server. This panics
//if the server is started.
func (server *Server) SetDefaultWorld(def string) {
	if server.running {
		panic("Server is running")
	}
	server.worlds.def = def
}

//DefaultWorld returns the default world for the server
func (server *Server) DefaultWorld() *world.World {
	return server.World(server.worlds.def)
}

//LoadWorld gets the world by name. If the world isn't loaded but it exists
//then it is loaded. If it doesn't exist then it is created using
//the passed system.
func (server *Server) LoadWorld(name string, system world.System, gen world.Generator, dimension world.Dimension) *world.World {
	w := server.World(name)
	if w != nil {
		return w
	}
	server.worlds.Lock()
	if w = server.worlds.m[name]; w != nil { //Double check now we have the lock
		return w
	}
	wait, ok := server.worlds.waitMap[name]
	if ok { //Another goroutine is currently loading the world
		server.worlds.Unlock()
		wait.Wait() //Wait for completion
		server.worlds.RLock()
		w := server.worlds.m[name] //Get the loaded world
		server.worlds.RUnlock()

		//This can happen if World is called before LoadWorld.
		//This will try again until it loads the world.
		//TODO: Maybe make it so this isn't needed
		if w == nil {
			w = server.LoadWorld(name, system, gen, dimension)
		}
		return w
	}
	wait = &sync.WaitGroup{}
	server.worlds.waitMap[name] = wait
	wait.Add(1) //Force other goroutines to wait for this world
	server.worlds.Unlock()
	w = world.LoadWorld(name, system, gen, dimension, server.worlds.tryClose)

	server.worlds.Lock()
	server.worlds.m[name] = w //Put the loaded world in the map
	server.worlds.Unlock()

	wait.Done() //Unpause other goroutines
	return w
}

//World gets the world by name. If the world isn't loaded but it exists
//then it is loaded.
func (server *Server) World(name string) *world.World {
	//Check if the world is loaded
	server.worlds.RLock()
	w := server.worlds.m[name]
	server.worlds.RUnlock()

	if w == nil { //World isn't loaded
		server.worlds.Lock()
		if w = server.worlds.m[name]; w != nil { //Double check now we have the lock
			return w
		}
		wait, ok := server.worlds.waitMap[name]
		if ok { //Another goroutine is currently loading the world
			server.worlds.Unlock()
			wait.Wait() //Wait for completion
			server.worlds.RLock()
			w := server.worlds.m[name] //Get the loaded world
			server.worlds.RUnlock()
			return w
		}
		wait = &sync.WaitGroup{}
		server.worlds.waitMap[name] = wait
		wait.Add(1) //Force other goroutines to wait for this world
		server.worlds.Unlock()
		w = world.GetWorld(name, server.worlds.tryClose)

		server.worlds.Lock()
		server.worlds.m[name] = w //Put the loaded world in the map
		delete(server.worlds.waitMap, name)
		server.worlds.Unlock()

		wait.Done() //Unpause other goroutines
	}
	return w
}

func (server *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	mcConn := &protocol.Conn{
		Out:            conn,
		In:             conn,
		Deadliner:      conn,
		ReadDirection:  protocol.Serverbound,
		WriteDirection: protocol.Clientbound,
	}

	packet, err := mcConn.ReadPacket()
	if err != nil {
		//The client either had a early connection issue or
		//it isn't a minecraft client
		return
	}
	handshake, ok := packet.(protocol.Handshake)
	if !ok {
		//Client sent the wrong packet. This shouldn't
		//happen because in the handshaking protocol (default state)
		//only has the handshake packet as a valid packet
		return
	}

	//Status ping
	if handshake.State == 1 {
		mcConn.State = protocol.Status
		packet, err := mcConn.ReadPacket()
		if _, ok := packet.(protocol.StatusGet); !ok || err != nil {
			return
		}

		ping := server.getPing()
		ping.Version = defaultPingVersion

		online := atomic.LoadInt32(&server.playerCount)
		ping.Players.Online = int(online)

		by, err := json.Marshal(ping)
		if err != nil {
			panic(err)
		}
		mcConn.WritePacket(protocol.StatusResponse{string(by)})
		packet, err = mcConn.ReadPacket()
		if err != nil {
			return
		}
		cPing, ok := packet.(protocol.ClientStatusPing)
		if !ok {
			return
		}
		mcConn.WritePacket(protocol.StatusPing{Time: cPing.Time})
		return
	}

	if handshake.State != 2 {
		return
	}

	defer log.Printf("Killed %s", conn.RemoteAddr())
	log.Printf("Connection %s", conn.RemoteAddr())

	username, uuid, err := mcConn.Login(packet.(protocol.Handshake), server.authenticator)
	if err != nil {
		log.Printf("Player %s(%s) login error: %s", uuid, username, err)
		mcConn.WritePacket(protocol.LoginDisconnect{(&message.Message{Text: err.Error(), Color: message.Red}).JSONString()})
		return
	}

	p := player.NewPlayer(uuid, username, mcConn, server)

	//Adds the player to server
	server.global.add <- p
	atomic.AddInt32(&server.playerCount, 1)
	defer func() {
		server.global.remove <- p
		atomic.AddInt32(&server.playerCount, -1)
	}()

	ok, msg := server.Handler.PlayerJoin(p)
	if ok {
		mcConn.WritePacket(protocol.Disconnect{msg})
		server.global.remove <- p
		return
	}

	//This will block until the player logouts or is kicked
	p.Start()
}

//QueuePacket queues the packet to be send to every player on the server
func (server *Server) QueuePacket(packet protocol.Packet) {
	server.global.packet <- packet
}

//SendMessage sends the message to every player on the server
func (server *Server) SendMessage(msg *message.Message) {
	server.QueuePacket(protocol.ServerMessage{msg.JSONString()})
}

//SetPing set the ping to be showed on minecraft's server browser
func (server *Server) SetPing(ping Ping) {
	server.ping.Lock()
	defer server.ping.Unlock()
	server.ping.data = ping
}

func (server *Server) getPing() Ping {
	server.ping.RLock()
	defer server.ping.RUnlock()
	return server.ping.data
}
