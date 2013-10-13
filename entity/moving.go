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

package entity

import (
	"github.com/NetherrackDev/netherrack/protocol"
	"github.com/NetherrackDev/netherrack/world"
)

type PositionComponent struct {
	CX, CZ     int32
	X, Y, Z    float64
	Yaw, Pitch float32
}

func (p *PositionComponent) Position() *PositionComponent {
	return p
}

type LastPositionComponent struct {
	MovedChunk          bool
	LastCX, LastCZ      int32
	LastX, LastY, LastZ float64
	LastYaw, LastPitch  float32
}

func (lp *LastPositionComponent) LastPosition() *LastPositionComponent {
	return lp
}

func init() {
	RegisterSystem(SystemMovable{})
}

type SystemMovable struct{}

type movable interface {
	world.Entity
	Entity() *EntityComponent
	Position() *PositionComponent
	LastPosition() *LastPositionComponent
	SpawnPackets() []protocol.Packet
}

func (SystemMovable) Valid(e interface{}) bool {
	_, ok := e.(movable)
	return ok
}

func (SystemMovable) Priority() Priority { return Normal }

//Updates the entity's movement and moves the chunk its in if required
func (SystemMovable) Update(entity interface{}) {
	mov := entity.(movable)
	e := mov.Entity()
	p := mov.Position()
	m := mov.LastPosition()
	p.CX, p.CZ = int32(p.X)>>4, int32(p.Z)>>4
	if p.CX != m.LastCX || p.CZ != m.LastCZ {
		m.MovedChunk = true
		//TODO: Move the entity to the next chunk
	}
	m.LastCX, m.LastCZ = p.CX, p.CZ

	if e.CurrentTick%2 == 0 {
		moved := false
		if e.CurrentTick%(10*5) == 0 {
			moved = true
			e.World.QueuePacket(int(p.CX), int(p.CZ), e.Uuid, protocol.EntityTeleport{
				EntityID: e.ID,
				X:        int32(p.X * 32),
				Y:        int32(p.Y * 32),
				Z:        int32(p.Z * 32),
				Yaw:      int8((p.Yaw / 360) * 256),
				Pitch:    int8((p.Pitch / 360) * 256),
			})
		} else {
			dx := p.X - m.LastX
			dy := p.Y - m.LastY
			dz := p.Z - m.LastZ
			dyaw := p.Yaw - m.LastYaw
			dpitch := p.Pitch - m.LastPitch
			if dx >= 4 || dy >= 4 || dz >= 4 {
				moved = true
				e.World.QueuePacket(int(p.CX), int(p.CZ), e.Uuid, protocol.EntityTeleport{
					EntityID: e.ID,
					X:        int32(p.X * 32),
					Y:        int32(p.Y * 32),
					Z:        int32(p.Z * 32),
					Yaw:      int8((p.Yaw / 360) * 256),
					Pitch:    int8((p.Pitch / 360) * 256),
				})
				e.World.QueuePacket(int(p.CX), int(p.CZ), e.Uuid, protocol.EntityHeadLook{
					EntityID: e.ID,
					HeadYaw:  int8((p.Yaw / 360) * 256),
				})
			} else if (dx != 0 || dy != 0 || dz != 0) && (dyaw != 0 || dpitch != 0) {
				moved = true
				e.World.QueuePacket(int(p.CX), int(p.CZ), e.Uuid, protocol.EntityLookMove{
					EntityID: e.ID,
					DX:       int8(dx * 32),
					DY:       int8(dy * 32),
					DZ:       int8(dz * 32),
					Yaw:      int8((p.Yaw / 360) * 256),
					Pitch:    int8((p.Pitch / 360) * 256),
				})
				e.World.QueuePacket(int(p.CX), int(p.CZ), e.Uuid, protocol.EntityHeadLook{
					EntityID: e.ID,
					HeadYaw:  int8((p.Yaw / 360) * 256),
				})
			} else if dx != 0 || dy != 0 || dz != 0 {
				moved = true
				e.World.QueuePacket(int(p.CX), int(p.CZ), e.Uuid, protocol.EntityMove{
					EntityID: e.ID,
					DX:       int8(dx * 32),
					DY:       int8(dy * 32),
					DZ:       int8(dz * 32),
				})
			} else if dyaw != 0 || dpitch != 0 {
				moved = true
				e.World.QueuePacket(int(p.CX), int(p.CZ), e.Uuid, protocol.EntityLook{
					EntityID: e.ID,
					Yaw:      int8((p.Yaw / 360) * 256),
					Pitch:    int8((p.Pitch / 360) * 256),
				})
				e.World.QueuePacket(int(p.CX), int(p.CZ), e.Uuid, protocol.EntityHeadLook{
					EntityID: e.ID,
					HeadYaw:  int8((p.Yaw / 360) * 256),
				})
			}
			if moved {
				e.World.UpdateSpawnData(int(p.CX), int(p.CZ), mov, true, mov.SpawnPackets())
			}
		}

		m.LastX, m.LastY, m.LastZ = p.X, p.Y, p.Z
		m.LastYaw, m.LastPitch = p.Yaw, p.Pitch
	}

	return
}
