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
	"errors"
	"github.com/NetherrackDev/netherrack/entity/player"
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
	MinecraftVersion = "13w38a"
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
}

//Creates a server
func NewServer() *Server {
	server := &Server{
		authenticator: auth.Instance,
	}
	server.worlds.m = make(map[string]*world.World)
	server.worlds.waitMap = make(map[string]*sync.WaitGroup)

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

	for {
		conn, err := listen.Accept()
		if err != nil {
			return err
		}
		go server.handleConnection(conn)
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

	packet, err := mcConn.ReadPacket()
	if err != nil {
		mcConn.WritePacket(protocol.Disconnect{"Protocol Error"})
		return
	}
	if _, ok := packet.(protocol.ServerListPing); ok {
		/*
			In 1.6 extra data was added containing the clients
			protocol version and the ip:port of the server they are
			connected to. This infomation is not sent by older clients
			so a custom read deadline should be set to prevent waiting
			too long causing the client to see a large ping time.
		*/
		conn.SetReadDeadline(time.Now().Add(time.Millisecond))
		mcConn.Deadliner = nil
		packet, err := mcConn.ReadPacket()

		var response string

		if err == nil {
			if extraData, ok := packet.(protocol.PluginMessage); ok {
				if extraData.Channel != "MC|PingHost" {
					err = errors.New("Incorrect channel")
					goto oldPing
				}
				if len(extraData.Data) < 3 {
					err = errors.New("Incorrect size")
					goto oldPing
				}
				protoVersion := extraData.Data[0]
				offset, host := readString(extraData.Data[1:])
				if len(extraData.Data) < offset+1+4 {
					err = errors.New("Incorrect size")
					goto oldPing
				}
				port := binary.BigEndian.Uint32(extraData.Data[offset+1:])

				server.event.RLock()
				if server.event.pingEvent != nil {
					res := make(chan string, 1)
					pingEvent := PingEvent{
						Addr:            conn.RemoteAddr(),
						ProtocolVersion: protoVersion,
						TargetHost:      host,
						TargetPort:      int(port),
						Response:        res,
					}
					server.event.pingEvent <- pingEvent
					response = <-res
				}
				server.event.RUnlock()
			}
		}
	oldPing:
		if err != nil {
			server.event.RLock()
			if server.event.oldPingEvent != nil {
				res := make(chan string, 1)
				event := OldPingEvent{conn.RemoteAddr(), res}
				server.event.oldPingEvent <- event
				response = <-res
			}
			server.event.RUnlock()
		}
		mcConn.Deadliner = conn
		if response == "" {
			response = server.buildServerPing()
		}
		mcConn.WritePacket(protocol.Disconnect{response})
		return
	}
	if _, ok := packet.(protocol.Handshake); !ok {
		mcConn.WritePacket(protocol.Disconnect{"Protocol Error"})
		return
	}

	username, uuid, err := mcConn.Login(packet.(protocol.Handshake), server.authenticator)
	if err != nil {
		log.Printf("Player %s(%s) login error: %s", uuid, username, err)
		mcConn.WritePacket(protocol.Disconnect{err.Error()})
		return
	}

	p := player.NewPlayer(uuid, username, mcConn, server)

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
			return
		}
	}
	server.event.RUnlock()

	p.Start()
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
