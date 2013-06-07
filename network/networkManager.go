package network

import (
	"encoding/binary"
	"fmt"
	"github.com/NetherrackDev/netherrack/log"
	"github.com/NetherrackDev/netherrack/player"
	"github.com/NetherrackDev/soulsand"
	"net"
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

//Work out if it is a listPing or a handshake
func handleInitialConnection(conn net.Conn) {
	defer conn.Close()
	firstByte := make([]byte, 1)
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	n, err := conn.Read(firstByte)
	if n != 1 || err != nil {
		log.Println(err)
		return
	}
	if firstByte[0] == 0xFE { //listPing
		conn.Read(make([]byte, 1))
		listData := soulsand.GetServer().GetListPingData()
		motd := []rune(listData[0])
		versionNo := []rune(listData[1])
		version := []rune(listData[2])
		currentPlayers := []rune(listData[3])
		maxPlayers := []rune(listData[4])
		strLen := 3 + len(versionNo) + 1 + len(version) + 1 + len(motd) + 1 + len(currentPlayers) + 1 + len(maxPlayers)
		msg := make([]byte, 1+2+strLen*2)
		msg[0] = 0xFF
		binary.BigEndian.PutUint16(msg[1:3], uint16(strLen))
		binary.BigEndian.PutUint16(msg[3:5], uint16(0xA7))
		binary.BigEndian.PutUint16(msg[5:7], uint16(0x31))

		pos := 9
		for _, b := range versionNo {
			binary.BigEndian.PutUint16(msg[pos:pos+2], uint16(b))
			pos += 2
		}
		pos += 2
		for _, b := range version {
			binary.BigEndian.PutUint16(msg[pos:pos+2], uint16(b))
			pos += 2
		}
		pos += 2
		for _, b := range motd {
			binary.BigEndian.PutUint16(msg[pos:pos+2], uint16(b))
			pos += 2
		}
		pos += 2
		for _, b := range currentPlayers {
			binary.BigEndian.PutUint16(msg[pos:pos+2], uint16(b))
			pos += 2
		}
		pos += 2
		for _, b := range maxPlayers {
			binary.BigEndian.PutUint16(msg[pos:pos+2], uint16(b))
			pos += 2
		}
		conn.Write(msg)
	} else { //handshake
		player.HandlePlayer(conn)
	}
	log.Println(conn.RemoteAddr().String() + " closed")
}
