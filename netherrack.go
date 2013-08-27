package netherrack

import (
	"encoding/binary"
	// "github.com/NetherrackDev/netherrack/log"
	"github.com/NetherrackDev/netherrack/protocol"
	"log"
	"net"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

const (
	//The currently supported protocol verison
	ProtocolVersion       = 74
	protocolVersionString = "74" //Save int-string conversion in list ping
	//The currently supported Minecraft version
	MinecraftVersion = "1.6.2"
)

//Stores server related infomation
type Server struct {
	listener net.Listener

	listData struct {
		sync.RWMutex
		MessageOfTheDay string
	}
}

func (server *Server) init() {

}

//Creates a server that will be bound to the passed ip:port
func NewServer() *Server {
	server := &Server{}

	server.listData.MessageOfTheDay = "Netherrack Server"

	return server
}

//Starts the minecraft server. This will block while the server is running
func (server *Server) Start(address string) error {
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

func (server *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	mcConn := protocol.CreateConnection(conn)
	packetId := mcConn.ReadUByte()
	if packetId == 0xFE {
		mcConn.ReadUByte()

		/*
			In 1.6 extra data was added containing the clients
			protocol version and the ip:port of the server they are
			connected to. This infomation is not sent by older clients
			so a custom read deadline should be set to prevent waiting
			too long causing the client to see a large ping time.
		*/

		extraData := make([]byte, 1)
		conn.SetReadDeadline(time.Now().Add(150 * time.Nanosecond))
		if _, err := conn.Read(extraData); err == nil && extraData[0] == 0xFA {
			//Extra data found
			channel, data := mcConn.ReadPluginMessage()

			if channel != "MC|PingHost" {
				return
			}
			if len(data) < 3 {
				return
			}
			protoVersion := data[0]
			offset, host := readString(data[1:])
			if len(data) < offset+1+4 {
				log.Println(offset, len(data), host)
				return
			}
			port := binary.BigEndian.Uint32(data[offset+1:])
			_, _, _ = protoVersion, host, port //TODO: Something with the new infomation
		}

		mcConn.WriteDisconnect(server.buildServerPing())
	}
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
