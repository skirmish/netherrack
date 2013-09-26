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
	"github.com/NetherrackDev/netherrack/message"
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
	//Sends the packet to every player on the server
	QueuePacket(packet protocol.Packet)
	//Sends the message to every player on the server
	SendMessage(msg *message.Message)
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
	spawnFor      chan world.Watcher
	despawnFor    chan world.Watcher

	rand   *rand.Rand
	pingID int32

	event struct {
		sync.RWMutex
		blockPlace chan<- BlockPlacement
		blockDig   chan<- BlockDig
	}
}

func NewPlayer(uuid, username string, conn *protocol.Conn, server Server) *Player {
	p := &Player{
		Username:      username,
		Conn:          conn,
		packetQueue:   make(chan protocol.Packet, 200),
		readPackets:   make(chan protocol.Packet, 20),
		errorChannel:  make(chan error, 1),
		closedChannel: make(chan struct{}, 1),
		spawnFor:      make(chan world.Watcher, 2),
		despawnFor:    make(chan world.Watcher, 2),
		Server:        server,
		rand:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	p.CommonEntity.Server = server
	p.CommonEntity.ID = entity.GetID()
	p.CommonEntity.Uuid = uuid
	p.pingID = -1
	go p.packetReader()
	return p
}

//Queues a packet to be sent to the player
func (p *Player) QueuePacket(packet protocol.Packet) {
	select {
	case p.packetQueue <- packet:
	case _, _ = <-p.closedChannel:
	}
}

//Processes incomming and outgoing packets. Blocks until the player leaves
//or is kicked.
func (p *Player) Start() {
	defer p.close()

	p.World = p.Server.DefaultWorld()

	p.Conn.WritePacket(protocol.LoginRequest{
		EntityID:   p.ID,
		LevelType:  "netherrack", //Not used by the client
		Gamemode:   1,
		Dimension:  int8(p.World.Dimension()),
		Difficulty: 3,
		MaxPlayers: 127,
	})
	p.Conn.WritePacket(protocol.PluginMessage{
		Channel: "MC|Brand",
		Data:    []byte("Netherrack"),
	})
	p.Conn.WritePacket(protocol.PlayerPositionLook{
		X:        0,
		Y:        90,
		Stance:   90 + 1.6,
		Z:        0,
		Yaw:      0,
		Pitch:    0,
		OnGround: true,
	})
	p.spawn()
	defer p.despawn()

	for x := -10; x <= 10; x++ {
		for z := -10; z <= 10; z++ {
			p.World.JoinChunk(x, z, p)
		}
	}
	tick := time.NewTicker(time.Second / 10)
	var currentTick uint64
	defer tick.Stop()
	for {
		select {
		case err := <-p.errorChannel:
			log.Printf("Player %s error: %s\n", p.Username, err)
			return
		case <-tick.C:
			if currentTick%(15*10) == 0 { //Every 15 seconds
				if p.pingID != -1 {
					p.disconnect("Timed out")
					continue
				}
				p.pingID = p.rand.Int31()
				p.Conn.WritePacket(protocol.KeepAlive{p.pingID})
			}
			lcx, lcz := p.LastCX, p.LastCZ
			if p.UpdateMovement(p) {
				for x := lcx - 10; x <= lcx+10; x++ {
					for z := lcz - 10; z <= lcz+10; z++ {
						if x < p.CX-10 || x > p.CX+10 || z < p.CZ-10 || z > p.CZ+10 {
							p.World.LeaveChunk(int(x), int(z), p)
						}
					}
				}
				for x := p.CX - 10; x <= p.CX+10; x++ {
					for z := p.CZ - 10; z <= p.CZ+10; z++ {
						if x < lcx-10 || x > lcx+10 || z < lcz-10 || z > lcz+10 {
							p.World.JoinChunk(int(x), int(z), p)
						}
					}
				}
			}
			currentTick++
		case packet := <-p.packetQueue:
			p.Conn.WritePacket(packet)
		case packet := <-p.readPackets:
			p.processPacket(packet)
		case watcher := <-p.spawnFor:
			packets := p.spawnPackets()
			for _, packet := range packets {
				watcher.QueuePacket(packet)
			}
		case watcher := <-p.despawnFor:
			packets := p.despawnPackets()
			for _, packet := range packets {
				watcher.QueuePacket(packet)
			}
		}
	}
}

func (p *Player) spawn() {
	p.World.AddEntity(int(p.CX), int(p.CZ), p)
}

func (p *Player) despawn() {
	p.World.RemoveEntity(int(p.CX), int(p.CZ), p)
}

func (p *Player) spawnPackets() []protocol.Packet {
	return []protocol.Packet{
		protocol.SpawnNamedEntity{
			EntityID:    p.ID,
			PlayerName:  p.Username,
			PlayerUUID:  p.Uuid,
			X:           int32(p.X * 32),
			Y:           int32(p.Y * 32),
			Z:           int32(p.Z * 32),
			Yaw:         int8((p.Yaw / 360) * 256),
			Pitch:       int8((p.Pitch / 360) * 256),
			CurrentItem: 0,
			Metadata:    map[byte]interface{}{0: int8(0)},
		},
	}
}

func (p *Player) despawnPackets() []protocol.Packet {
	return []protocol.Packet{
		protocol.EntityDestroy{[]int32{p.ID}},
	}
}

func (p *Player) Saveable() bool {
	return false
}

func (p *Player) SpawnFor(watcher world.Watcher) {
	p.spawnFor <- watcher
}

func (p *Player) DespawnFor(watcher world.Watcher) {
	p.despawnFor <- watcher
}

//Acts on the passed packet
//TODO: Kick player on wrong packet
func (p *Player) processPacket(packet protocol.Packet) {
	switch packet := packet.(type) {
	case protocol.PlayerDigging:
		p.event.RLock()
		if p.event.blockDig != nil {
			res := make(chan struct{}, 1)
			p.event.blockDig <- BlockDig{
				Packet: packet,
				Return: res,
			}
			<-res
		}
		p.event.RUnlock()
	case protocol.PlayerBlockPlacement:
		p.event.RLock()
		if p.event.blockPlace != nil {
			res := make(chan struct{}, 1)
			p.event.blockPlace <- BlockPlacement{
				Packet: packet,
				Return: res,
			}
			<-res
		}
		p.event.RUnlock()
	case protocol.Player:
	case protocol.PlayerLook:
		yaw := math.Mod(float64(packet.Yaw), 360)
		if yaw < 0 {
			yaw = 360 + yaw
		}
		p.Yaw = float32(yaw)
		p.Pitch = packet.Pitch
	case protocol.PlayerPosition:
		p.X, p.Y, p.Z = packet.X, packet.Y, packet.Z
	case protocol.PlayerPositionLook:
		p.X, p.Y, p.Z = packet.X, packet.Y, packet.Z
		yaw := math.Mod(float64(packet.Yaw), 360)
		if yaw < 0 {
			yaw = 360 + yaw
		}
		p.Yaw = float32(yaw)
		p.Pitch = packet.Pitch
	case protocol.KeepAlive:
		if p.pingID == -1 {
			return
		}
		if p.pingID != packet.KeepAliveID {
			p.disconnect("Incorrect KeepAliveID")
			return
		}
		p.pingID = -1
	case protocol.Disconnect:
		p.disconnect(packet.Reason)
	}
}

func (p *Player) disconnect(reason string) {
	p.Conn.WritePacket(protocol.Disconnect{reason})
	p.errorChannel <- errors.New(reason)
}

//Close and cleanup the player. The packetReader will close
//once the orginal net.Conn is closed.
func (p *Player) close() {
	close(p.closedChannel)
	for x := p.CX - 10; x <= p.CX+10; x++ {
		for z := p.CZ - 10; z <= p.CZ+10; z++ {
			p.World.LeaveChunk(int(x), int(z), p)
		}
	}
	entity.FreeID(p.ID)
}

//Reads incomming packets and passes them to the watcher
func (p *Player) packetReader() {
	for {
		packet, err := p.Conn.ReadPacket()
		if err != nil {
			p.errorChannel <- err
			return
		}
		p.readPackets <- packet
	}
}
