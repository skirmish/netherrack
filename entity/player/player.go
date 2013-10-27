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
	entity.EntityComponent
	//Parts
	entity.PositionComponent
	entity.LastPositionComponent

	conn     *protocol.Conn
	uuid     string
	Username string
	Server   Server

	packetQueue   chan protocol.Packet
	readPackets   chan protocol.Packet
	errorChannel  chan error
	ClosedChannel chan struct{}

	rand   *rand.Rand
	pingID int32

	Handler PlayerHandler

	LockChan chan chan struct{}

	permission map[string]bool
}

func NewPlayer(uuid, username string, conn *protocol.Conn, server Server) *Player {
	p := &Player{
		Username:      username,
		conn:          conn,
		packetQueue:   make(chan protocol.Packet, 200),
		readPackets:   make(chan protocol.Packet, 20),
		errorChannel:  make(chan error, 1),
		ClosedChannel: make(chan struct{}),
		Server:        server,
		rand:          rand.New(rand.NewSource(time.Now().UnixNano())),
		LockChan:      make(chan chan struct{}),
		permission:    map[string]bool{},
	}
	p.EntityComponent.Server = server
	p.ID = entity.GetID()
	p.Uuid = uuid
	p.pingID = -1
	p.Init(p)
	go p.packetReader()
	go p.packetWriter()
	return p
}

func (p *Player) Player() *Player {
	return p
}

//Sends a message to the player
func (p *Player) SendMessage(msg *message.Message) {
	p.QueuePacket(protocol.ServerMessage{msg.JSONString()})
}

//Queues a packet to be sent to the player
func (p *Player) QueuePacket(packet protocol.Packet) {
	select {
	case p.packetQueue <- packet:
	case <-p.ClosedChannel:
	}
}

func (p *Player) CanCall(command string) bool {
	//thinkofdeath's uuid. Gives all permissions for testing
	//reasons.
	if p.Uuid == "4566e69fc90748ee8d71d7ba5aa00d20" {
		return true
	}
	return p.permission[command]
}

//Processes incomming and outgoing packets. Blocks until the player leaves
//or is kicked.
func (p *Player) Start() {
	defer p.close()

	p.World = p.Server.DefaultWorld()
	p.X, p.Y, p.Z = 0, 70, 0

	login := &protocol.JoinGame{
		EntityID:   p.ID,
		Gamemode:   0,
		Dimension:  int8(p.World.Dimension()),
		Difficulty: 0,
		MaxPlayers: 0,
		LevelType:  "default",
	}

	p.Handler.EnterWorld(login)
	defer p.Handler.Leave()

	p.QueuePacket(*login)
	p.QueuePacket(protocol.PluginMessage{
		Channel: "MC|Brand",
		Data:    []byte("Netherrack"),
	})
	p.QueuePacket(protocol.PlayerPositionLook{
		X:        p.X,
		Y:        p.Y,
		Z:        p.Z,
		Yaw:      p.Yaw,
		Pitch:    p.Pitch,
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
	defer tick.Stop()
	for {
		select {
		case err := <-p.errorChannel:
			log.Printf("Player %s error: %s\n", p.Username, err)
			return
		case <-tick.C:
			if p.CurrentTick%60 == 0 {
				p.QueuePacket(protocol.TimeUpdate{
					TimeOfDay: -16000, //p.World.TimeOfDay(),
				})
			}
			if p.CurrentTick%(15*10) == 0 { //Every 15 seconds
				if p.pingID != -1 {
					p.disconnect("Timed out")
					continue
				}
				p.pingID = p.rand.Int31()
				p.QueuePacket(protocol.KeepAlive{p.pingID})
			}
			lcx, lcz := p.LastCX, p.LastCZ
			p.Update(p)
			if p.MovedChunk {
				p.MovedChunk = false
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
		case packet := <-p.readPackets:
			p.processPacket(packet)
		case lock := <-p.LockChan:
			<-lock
		}
	}
}

func (p *Player) spawn() {
	p.World.AddEntity(int(p.CX), int(p.CZ), p, p.SpawnPackets(), p.DespawnPackets())
}

func (p *Player) despawn() {
	p.World.RemoveEntity(int(p.CX), int(p.CZ), p)
}

func (p *Player) SpawnPackets() []protocol.Packet {
	return []protocol.Packet{
		protocol.SpawnPlayer{
			EntityID:    protocol.VarInt(p.ID),
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

func (p *Player) DespawnPackets() []protocol.Packet {
	return []protocol.Packet{
		protocol.EntityDestroy{[]int32{p.ID}},
	}
}

func (p *Player) Saveable() bool {
	return false
}

//Acts on the passed packet
//TODO: Kick player on wrong packet
func (p *Player) processPacket(packet protocol.Packet) {
	switch packet := packet.(type) {
	case protocol.ChatMessage:
		p.Handler.Chat(packet.Message)
	case protocol.PlayerDigging:
		p.Handler.BlockDig(packet)
	case protocol.PlayerBlockPlacement:
		p.Handler.BlockPlacement(packet)
	case protocol.ClientPlayer:
	case protocol.ClientPlayerLook:
		yaw := math.Mod(float64(packet.Yaw), 360)
		if yaw < 0 {
			yaw = 360 + yaw
		}
		p.Yaw = float32(yaw)
		p.Pitch = packet.Pitch
	case protocol.ClientPlayerPosition:
		p.X, p.Y, p.Z = packet.X, packet.Y, packet.Z
	case protocol.ClientPlayerPositionLook:
		p.X, p.Y, p.Z = packet.X, packet.Y, packet.Z
		yaw := math.Mod(float64(packet.Yaw), 360)
		if yaw < 0 {
			yaw = 360 + yaw
		}
		p.Yaw = float32(yaw)
		p.Pitch = packet.Pitch
	case protocol.ClientKeepAlive:
		if p.pingID == -1 {
			return
		}
		if p.pingID != packet.KeepAliveID {
			p.disconnect("Incorrect KeepAliveID")
			return
		}
		p.pingID = -1
	}
}

func (p *Player) disconnect(reason string) {
	p.QueuePacket(protocol.Disconnect{reason})
	p.errorChannel <- errors.New(reason)
}

//Close and cleanup the player. The packetReader will close
//once the orginal net.Conn is closed.
func (p *Player) close() {
	close(p.ClosedChannel)
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
		packet, err := p.conn.ReadPacket()
		if err != nil {
			p.errorChannel <- err
			return
		}
		select {
		case p.readPackets <- packet:
		case <-p.ClosedChannel:
			return
		}
	}
}

func (p *Player) packetWriter() {
	for {
		select {
		case packet := <-p.packetQueue:
			p.conn.WritePacket(packet)
		case <-p.ClosedChannel:
			return
		}
	}
}
