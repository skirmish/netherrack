package network

import (
	"encoding/binary"
	"fmt"
	"github.com/NetherrackDev/netherrack/player"
	"github.com/NetherrackDev/netherrack/protocol"
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/log"
	"net"
	"strings"
	"time"
)

//Start listening for connections
func Listen(ip string, port int) {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(fmt.Sprintf("Listening on %s:%d", ip, port))
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("Connection from " + conn.RemoteAddr().String())
		go handleInitialConnection(conn)
	}
}

type netherrackServer interface {
	ListPingData() []string
}

//Work out if it is a listPing or a handshake
func handleInitialConnection(conn net.Conn) {
	defer conn.Close()
	firstByte := make([]byte, 1)
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	n, err := conn.Read(firstByte)
	if n != 1 || err != nil {
		log.Println(err)
		return
	}
	if firstByte[0] == 0xFE { //listPing
		connection := protocol.CreateConnection(conn)
		connection.Read(make([]byte, 1))

		//1.6 added extra info to the list ping
		extra := make([]byte, 1)
		//Do this manually because connection.Read sets the deadline to 15 seconds but this data may not be there
		conn.SetReadDeadline(time.Now().Add(150 * time.Nanosecond))
		if _, err := conn.Read(extra); err == nil {
			log.Println("Got extra data")
			if extra[0] != 0xFA {
				return
			}
			channel, data := connection.ReadPluginMessage()
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

		listData := soulsand.GetServer().(netherrackServer).ListPingData()
		ping := "§1\x00" + strings.Join(listData, "\x00")
		connection.WriteDisconnect(ping)
	} else { //handshake
		player.HandlePlayer(conn)
	}
	log.Println(conn.RemoteAddr().String() + " closed")
}

func readString(data []byte) (off int, str string) {
	length := binary.BigEndian.Uint16(data)
	off += 2 + int(length)*2
	runes := make([]rune, length)
	for i, _ := range runes {
		runes[i] = rune(binary.BigEndian.Uint16(data[2+i*2:]))
	}
	str = string(runes)
	return
}
