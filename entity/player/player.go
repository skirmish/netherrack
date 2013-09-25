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
	"github.com/NetherrackDev/netherrack/world"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

//Contains methods that a player needs for a server
type Server interface {
	entity.Server
	//Returns the default world for the server
	DefaultWorld() *world.World
	//Gets the world by name, loads the world if it isn't loaded
	World(name string) *world.World
}

//A local player is a player connected directly to this server
type Player struct {
	entity.CommonEntity

	Conn     *protocol.Conn
	uuid     string
	Username string
	Server   Server

	packetQueue   chan protocol.Packet
	readPackets   chan protocol.Packet
	errorChannel  chan error
	closedChannel chan struct{}

	rand   *rand.Rand
	pingID int32

	event struct {
		sync.RWMutex
		blockPlace chan<- BlockPlacement
		blockDig   chan<- BlockDig
	}
}

func NewPlayer(uuid, username string, conn *protocol.Conn, server Server) *Player {
	lp := &Player{
		uuid:          uuid,
		Username:      username,
		Conn:          conn,
		packetQueue:   make(chan protocol.Packet, 200),
		readPackets:   make(chan protocol.Packet, 20),
		errorChannel:  make(chan error, 1),
		closedChannel: make(chan struct{}, 1),
		Server:        server,
		rand:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	lp.CommonEntity.Server = server
	lp.pingID = -1
	go lp.packetReader()
	return lp
}

//Queues a packet to be sent to the player
func (lp *Player) QueuePacket(packet protocol.Packet) {
	select {
	case lp.packetQueue <- packet:
	case _, _ = <-lp.closedChannel:
	}
}

//Processes incomming and outgoing packets. Blocks until the player leaves
//or is kicked.
func (lp *Player) Start() {
	defer lp.close()

	lp.World = lp.Server.DefaultWorld()

	lp.Conn.WritePacket(protocol.LoginRequest{
		EntityID:   0,
		LevelType:  "netherrack", //Not used by the client
		Gamemode:   1,
		Dimension:  int8(lp.World.Dimension()),
		Difficulty: 3,
		MaxPlayers: 127,
	})
	lp.Conn.WritePacket(protocol.PluginMessage{
		Channel: "MC|Brand",
		Data:    []byte("Netherrack"),
	})
	lp.Conn.WritePacket(protocol.PlayerPositionLook{
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
			log.Printf("Player %s error: %s\n", lp.Username, err)
			return
		case <-tick.C:
			if currentTick%(15*10) == 0 { //Every 15 seconds
				if lp.pingID != -1 {
					lp.disconnect("Timed out")
					continue
				}
				lp.pingID = lp.rand.Int31()
				lp.Conn.WritePacket(protocol.KeepAlive{lp.pingID})
			}
			lcx, lcz := lp.LastCX, lp.LastCZ
			if lp.UpdateMovement(lp) {
				for x := lcx - 10; x <= lcx+10; x++ {
					for z := lcz - 10; z <= lcz+10; z++ {
						if x < lp.CX-10 || x > lp.CX+10 || z < lp.CZ-10 || z > lp.CZ+10 {
							lp.World.LeaveChunk(int(x), int(z), lp)
						}
					}
				}
				for x := lp.CX - 10; x <= lp.CX+10; x++ {
					for z := lp.CZ - 10; z <= lp.CZ+10; z++ {
						if x < lcx-10 || x > lcx+10 || z < lcz-10 || z > lcz+10 {
							lp.World.JoinChunk(int(x), int(z), lp)
						}
					}
				}
			}
			currentTick++
		case packet := <-lp.packetQueue:
			lp.Conn.WritePacket(packet)
		case packet := <-lp.readPackets:
			lp.processPacket(packet)
		}
	}
}

//Acts on the passed packet
//TODO: Kick player on wrong packet
func (lp *Player) processPacket(packet protocol.Packet) {
	switch packet := packet.(type) {
	case protocol.PlayerDigging:
		lp.event.RLock()
		if lp.event.blockDig != nil {
			res := make(chan struct{}, 1)
			lp.event.blockDig <- BlockDig{
				Packet: packet,
				Return: res,
			}
			<-res
		}
		lp.event.RUnlock()
	case protocol.PlayerBlockPlacement:
		lp.event.RLock()
		if lp.event.blockPlace != nil {
			res := make(chan struct{}, 1)
			lp.event.blockPlace <- BlockPlacement{
				Packet: packet,
				Return: res,
			}
			<-res
		}
		lp.event.RUnlock()
	case protocol.Player:
	case protocol.PlayerLook:
		yaw := math.Mod(float64(packet.Yaw), 360)
		if yaw < 0 {
			yaw = 360 + yaw
		}
		lp.Yaw = float32(yaw)
		lp.Pitch = packet.Pitch
	case protocol.PlayerPosition:
		lp.X, lp.Y, lp.Z = packet.X, packet.Y, packet.Z
	case protocol.PlayerPositionLook:
		lp.X, lp.Y, lp.Z = packet.X, packet.Y, packet.Z
		yaw := math.Mod(float64(packet.Yaw), 360)
		if yaw < 0 {
			yaw = 360 + yaw
		}
		lp.Yaw = float32(yaw)
		lp.Pitch = packet.Pitch
	case protocol.KeepAlive:
		if lp.pingID == -1 {
			return
		}
		if lp.pingID != packet.KeepAliveID {
			lp.disconnect("Incorrect KeepAliveID")
			return
		}
		lp.pingID = -1
	case protocol.Disconnect:
		lp.disconnect(packet.Reason)
	}
}

func (lp *Player) disconnect(reason string) {
	lp.Conn.WritePacket(protocol.Disconnect{reason})
	lp.errorChannel <- errors.New(reason)
}

//Close and cleanup the player. The packetReader will close
//once the orginal net.Conn is closed.
func (lp *Player) close() {
	close(lp.closedChannel)
	for x := lp.CX - 10; x <= lp.CX+10; x++ {
		for z := lp.CZ - 10; z <= lp.CZ+10; z++ {
			lp.World.LeaveChunk(int(x), int(z), lp)
		}
	}
}

//Reads incomming packets and passes them to the watcher
func (lp *Player) packetReader() {
	for {
		packet, err := lp.Conn.ReadPacket()
		if err != nil {
			lp.errorChannel <- err
			return
		}
		lp.readPackets <- packet
	}
}

//Returns the player's UUID
func (lp *Player) UUID() string {
	return lp.uuid
}
