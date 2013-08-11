package netherrack

import (
	"encoding/binary"
	"fmt"
	"github.com/NetherrackDev/netherrack/event"
	"github.com/NetherrackDev/netherrack/log"
	"github.com/NetherrackDev/netherrack/protocol"
	"net"
	"net/http"
	"runtime"
	"runtime/debug"
	"time"
)

const (
	ProtocolVersion  = 74
	MinecraftVersion = "1.6.2"
)

type Server struct {
	event.Source

	connection struct {
		Ip   string
		Port int
	}
}

func (server *Server) init() {
	server.Source.Init()
}

func NewServer(ip string, port int) *Server {
	server := &Server{}
	server.connection.Ip = ip
	server.connection.Port = port
	return server
}

func (server *Server) Start() {
	log.Printf("NumProcs: %d\n", runtime.GOMAXPROCS(-1))
	debug.SetGCPercent(10)
	go func() {
		log.Println(http.ListenAndServe(server.connection.Ip+":25567", nil))
	}()
	log.Println("Starting Netherrack server")

	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.connection.Ip, server.connection.Port))
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			panic(err)
		}
		go server.handleConnection(conn)
	}
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

	}
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
