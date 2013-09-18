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

package player

import (
	"errors"
	"github.com/NetherrackDev/netherrack/entity"
	"github.com/NetherrackDev/netherrack/protocol"
	"log"
	"math/rand"
	"time"
)

//A local player is a player connected directly to this server
type LocalPlayer struct {
	entity.LocalEntity

	conn     *protocol.Conn
	username string
	Server   Server

	packetQueue   chan protocol.Packet
	readPackets   chan protocol.Packet
	errorChannel  chan error
	closedChannel chan struct{}

	pingID int32
}

func NewLocalPlayer(username string, conn *protocol.Conn, server Server) *LocalPlayer {
	lp := &LocalPlayer{
		username:      username,
		conn:          conn,
		packetQueue:   make(chan protocol.Packet, 200),
		readPackets:   make(chan protocol.Packet, 20),
		errorChannel:  make(chan error, 1),
		closedChannel: make(chan struct{}, 1),
		Server:        server,
	}
	lp.LocalEntity.Server = server
	lp.pingID = -1
	go lp.packetReader()
	return lp
}

//Queues a packet to be sent to the player
func (lp *LocalPlayer) QueuePacket(packet protocol.Packet) {
	select {
	case lp.packetQueue <- packet:
	case _, _ = <-lp.closedChannel:
	}
}

//Processes incomming and outgoing packets. Blocks until the player leaves
//or is kicked.
func (lp *LocalPlayer) Start() {
	defer lp.close()

	lp.World = lp.Server.DefaultWorld()

	lp.conn.WritePacket(protocol.LoginRequest{
		EntityID:   0,
		LevelType:  "netherrack", //Not used by the client
		Gamemode:   1,
		Dimension:  0,
		Difficulty: 3,
		MaxPlayers: 127,
	})
	lp.conn.WritePacket(protocol.PlayerPositionLook{
		X:        0,
		Y:        90,
		Stance:   90 + 1.6,
		Z:        0,
		Yaw:      0,
		Pitch:    0,
		OnGround: true,
	})

	for x := -10; x <= 10; x++ {
		for z := -10; z <= 10; z++ {
			lp.World.JoinChunk(x, z, lp)
		}
	}
	tick := time.NewTicker(time.Second / 10)
	var currentTick uint64
	defer tick.Stop()
	for {
		select {
		case err := <-lp.errorChannel:
			log.Printf("Player %s error: %s\n", lp.username, err)
			return
		case <-tick.C:
			if currentTick%(15*10) == 0 { //Every 15 seconds
				lp.pingID = rand.Int31()
				lp.conn.WritePacket(protocol.KeepAlive{lp.pingID})
			}
			currentTick++
		case packet := <-lp.packetQueue:
			lp.conn.WritePacket(packet)
		case packet := <-lp.readPackets:
			lp.processPacket(packet)
		}
	}
}

//Acts on the passed packet
func (lp *LocalPlayer) processPacket(packet protocol.Packet) {
	switch packet := packet.(type) {
	case protocol.KeepAlive:
		if lp.pingID == -1 {
			return
		}
		if lp.pingID != packet.KeepAliveID {
			lp.disconnect("Incorrect KeepAliveID")
			return
		}
		lp.pingID = -1
	default:
		log.Printf("Unhandled packet %02X(%+v) from %s\n", packet.ID(), packet, lp.username)
	}
}

func (lp *LocalPlayer) disconnect(reason string) {
	lp.conn.WritePacket(protocol.Disconnect{reason})
	lp.errorChannel <- errors.New(reason)
}

//Close and cleanup the player. The packetReader will close
//once the orginal net.Conn is closed.
func (lp *LocalPlayer) close() {
	close(lp.closedChannel)
}

//Reads incomming packets and passes them to the watcher
func (lp *LocalPlayer) packetReader() {
	for {
		packet, err := lp.conn.ReadPacket()
		if err != nil {
			lp.errorChannel <- err
			return
		}
		lp.readPackets <- packet
	}
}
