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
	"github.com/NetherrackDev/netherrack/protocol"
	"github.com/NetherrackDev/netherrack/protocol/auth"
	"github.com/NetherrackDev/netherrack/world"
	"log"
	"net"
	"net/http"
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
	MinecraftVersion = "1.6.2"
)

var protocolVersionString = strconv.Itoa(ProtocolVersion) //Save int-string conversion in list ping

//Stores server related infomation
type Server struct {
	listener net.Listener
	running  bool

	listData struct {
		sync.RWMutex
		MessageOfTheDay string
	}

	worlds struct {
		sync.RWMutex
		m   map[string]world.World
		def world.World
	}

	authenticator protocol.Authenticator
}

//Creates a server
func NewServer() *Server {
	server := &Server{
		authenticator: auth.Instance,
	}
	server.listData.MessageOfTheDay = "Netherrack Server"

	return server
}

//Starts the minecraft server. This will block while the server is running
func (server *Server) Start(address string) error {
	server.running = true
	log.Printf("NumProcs: %d\n", runtime.GOMAXPROCS(-1))
	debug.SetGCPercent(10)
	go func() {
		log.Println(http.ListenAndServe(":25567", nil))
	}()
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

//Returns the default world for the server
func (server *Server) DefaultWorld() (world world.World) {
	server.worlds.RLock()
	world = server.worlds.def
	server.worlds.RUnlock()
	return
}

func (server *Server) World(name string) world.World {
	return nil
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
		conn.SetReadDeadline(time.Now().Add(9000 * time.Nanosecond))
		mcConn.Deadliner = nil
		packet, err := mcConn.ReadPacket()
		if err != nil {
			if extraData, ok := packet.(protocol.PluginMessage); ok {
				if extraData.Channel != "MC|PingHost" {
					goto sendPing
				}
				if len(extraData.Data) < 3 {
					goto sendPing
				}
				protoVersion := extraData.Data[0]
				offset, host := readString(extraData.Data[1:])
				if len(extraData.Data) < offset+1+4 {
					goto sendPing
				}
				port := binary.BigEndian.Uint32(extraData.Data[offset+1:])
				_, _, _ = protoVersion, host, port //TODO: Something with the new infomation
			}
		}
	sendPing:
		mcConn.Deadliner = conn
		mcConn.WritePacket(protocol.Disconnect{server.buildServerPing()})
		return
	}
	if _, ok := packet.(protocol.Handshake); !ok {
		mcConn.WritePacket(protocol.Disconnect{"Protocol Error"})
		return
	}

	username, err := mcConn.Login(packet.(protocol.Handshake), server.authenticator)
	if err != nil {
		log.Printf("Player %s login error: %s", username, err)
		mcConn.WritePacket(protocol.Disconnect{err.Error()})
		return
	}

	p := player.NewLocalPlayer(username, mcConn, server)
	p.Start()
}

func (server *Server) buildServerPing() string {
	server.listData.RLock()
	defer server.listData.RUnlock()
	return strings.Join([]string{
		"§1",
		protocolVersionString,
		MinecraftVersion,
		server.listData.MessageOfTheDay,
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
