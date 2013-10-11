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

package netherrack

import (
	"encoding/binary"
	"github.com/NetherrackDev/netherrack/entity/player"
	"github.com/NetherrackDev/netherrack/message"
	"github.com/NetherrackDev/netherrack/protocol"
	"github.com/NetherrackDev/netherrack/protocol/auth"
	"github.com/NetherrackDev/netherrack/world"
	"log"
	"net"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	//The currently supported protocol verison
	ProtocolVersion = protocol.Version
	//The currently supported Minecraft version
	MinecraftVersion = "13w41a"
)

var protocolVersionString = strconv.Itoa(ProtocolVersion) //Save int-string conversion in list ping

//Stores server related infomation
type Server struct {
	listener net.Listener
	running  bool

	worlds struct {
		sync.RWMutex
		m       map[string]*world.World
		waitMap map[string]*sync.WaitGroup
		def     *world.World
	}

	authenticator protocol.Authenticator

	event struct {
		sync.RWMutex
		oldPingEvent    chan<- OldPingEvent
		pingEvent       chan<- PingEvent
		playerJoinEvent chan<- PlayerJoinEvent
	}

	chat struct {
		packet chan protocol.Packet
		add    chan *player.Player
		remove chan *player.Player
	}
}

//Creates a server
func NewServer() *Server {
	server := &Server{
		authenticator: auth.Instance,
	}
	server.worlds.m = make(map[string]*world.World)
	server.worlds.waitMap = make(map[string]*sync.WaitGroup)
	server.chat.packet = make(chan protocol.Packet, 200)
	server.chat.add = make(chan *player.Player, 20)
	server.chat.remove = make(chan *player.Player, 20)
	return server
}

//Starts the minecraft server. This will block while the server is running
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

	go server.chatServer()
	for {
		conn, err := listen.Accept()
		if err != nil {
			return err
		}
		go server.handleConnection(conn)
	}
}

func (server *Server) chatServer() {
	players := map[string]*player.Player{}
	for {
		select {
		case packet := <-server.chat.packet:
			for _, p := range players {
				p.QueuePacket(packet)
			}
		case p := <-server.chat.add:
			players[p.UUID()] = p
		case p := <-server.chat.remove:
			players[p.UUID()] = p
		}
	}
}

//Returns the address the server is currently listening on
//once started.
func (server *Server) Addr() net.Addr {
	return server.listener.Addr()
}

//Changes the authenticator usered by the server. This panics
//if the server is started.
func (server *Server) SetAuthenticator(auth protocol.Authenticator) {
	if server.running {
		panic("Server is running")
	}
	server.authenticator = auth
}

//Sets the default world for the server
func (server *Server) SetDefaultWorld(world *world.World) {
	server.worlds.Lock()
	server.worlds.def = world
	server.worlds.Unlock()
}

//Returns the default world for the server
func (server *Server) DefaultWorld() (world *world.World) {
	server.worlds.RLock()
	world = server.worlds.def
	server.worlds.RUnlock()
	return
}

//Gets the world by name. If the world isn't loaded but it exists
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
		//TODO: Make it so this isn't needed
		if w == nil {
			w = server.LoadWorld(name, system, gen, dimension)
		}
		return w
	}
	wait = &sync.WaitGroup{}
	server.worlds.waitMap[name] = wait
	wait.Add(1) //Force other goroutines to wait for this world
	server.worlds.Unlock()
	w = world.LoadWorld(name, system, gen, dimension)

	server.worlds.Lock()
	server.worlds.m[name] = w //Put the loaded world in the map
	server.worlds.Unlock()

	wait.Done() //Unpause other goroutines
	return w
}

//Gets the world by name. If the world isn't loaded but it exists
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
		w = world.GetWorld(name)

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
	defer time.Sleep(time.Second / 2) //Allow for last messages to be sent before closing

	mcConn := &protocol.Conn{
		Out:       conn,
		In:        conn,
		Deadliner: conn,
	}

	log.Println("Connection")

	packet, err := mcConn.ReadPacket()
	if err != nil {
		return
	}

	handshake, ok := packet.(protocol.Handshake)
	if !ok {
		log.Println("Wrong packet")
		return
	}

	if handshake.State == 1 {
		mcConn.State = protocol.Status
		packet, err := mcConn.ReadPacket()
		if _, ok := packet.(protocol.StatusGet); !ok || err != nil {
			return
		}
		//TODO: Un-hard code this
		mcConn.WritePacket(protocol.StatusResponse{`{"description":{"text":"A Minecraft Server","color":"red"},"players":{"max":20,"online":1,"sample":[{"name":"Thinkofdeath","id":""}]},"version":{"name":"13w41b","protocol":0}}`})
		packet, err = mcConn.ReadPacket()
		if err != nil {
			panic(err)
			return
		}
		mcConn.WritePacket(packet)
		return
	}

	if handshake.State != 2 {
		log.Println("Invalid state")
		return
	}

	username, uuid, err := mcConn.Login(packet.(protocol.Handshake), server.authenticator)
	if err != nil {
		log.Printf("Player %s(%s) login error: %s", uuid, username, err)
		mcConn.WritePacket(protocol.LoginDisconnect{err.Error()})
		return
	}

	p := player.NewPlayer(uuid, username, mcConn, server)

	server.chat.add <- p

	server.event.RLock()
	if server.event.playerJoinEvent != nil {
		res := make(chan string, 1)
		event := PlayerJoinEvent{
			Player: p,
			Return: res,
		}
		server.event.playerJoinEvent <- event
		if msg := <-res; msg != "" {
			mcConn.WritePacket(protocol.Disconnect{msg})
			server.chat.remove <- p
			return
		}
	}
	server.event.RUnlock()

	p.Start()

	server.chat.remove <- p
}

//Sends the packet to every player on the server
func (server *Server) QueuePacket(packet protocol.Packet) {
	server.chat.packet <- packet
}

//Sends the message to every player on the server
func (server *Server) SendMessage(msg *message.Message) {
	server.QueuePacket(protocol.ServerMessage{msg.JSONString()})
}

func (server *Server) buildServerPing() string {
	return strings.Join([]string{
		"§1",
		protocolVersionString,
		MinecraftVersion,
		"Netherrack Server",
		"0",
		"100",
	}, "\x00")
}

func readString(data []byte) (off int, str string) {
	length := binary.BigEndian.Uint16(data)
	if int(length)*2+2 > len(data) {
		runtime.Goexit()
	}
	off += 2 + int(length)*2
	runes := make([]rune, length)
	for i, _ := range runes {
		runes[i] = rune(binary.BigEndian.Uint16(data[2+i*2:]))
	}
	str = string(runes)
	return
}
