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

type CommonEntity struct {
	ID   int32
	Uuid string
	//Exported for setting when embedded
	Server Server
	//Exported for setting when embedded
	World *world.World

	CX, CZ, LastCX, LastCZ int32
	LastX, LastY, LastZ    float64
	X, Y, Z                float64
	Yaw, Pitch             float32
	LastYaw, LastPitch     float32

	currentTick uint64
}

//Updates the entity's movement and moves the chunk its in if required
func (ce *CommonEntity) UpdateMovement(super Entity) (movedChunk bool) {
	ce.CX, ce.CZ = int32(ce.X)>>4, int32(ce.Z)>>4
	if ce.CX != ce.LastCX || ce.CZ != ce.LastCZ {
		movedChunk = true
		//TODO: Move the entity to the next chunk
	}
	ce.LastCX, ce.LastCZ = ce.CX, ce.CZ

	if ce.currentTick%2 == 0 {
		if ce.currentTick%(10*5) == 0 {
			ce.World.QueuePacket(int(ce.CX), int(ce.CZ), ce.Uuid, protocol.EntityTeleport{
				EntityID: ce.ID,
				X:        int32(ce.X * 32),
				Y:        int32(ce.Y * 32),
				Z:        int32(ce.Z * 32),
				Yaw:      int8((ce.Yaw / 360) * 256),
				Pitch:    int8((ce.Pitch / 360) * 256),
			})
		} else {
			dx := ce.X - ce.LastX
			dy := ce.Y - ce.LastY
			dz := ce.Z - ce.LastZ
			dyaw := ce.Yaw - ce.LastYaw
			dpitch := ce.Pitch - ce.LastPitch
			if dx >= 4 || dy >= 4 || dz >= 4 {
				ce.World.QueuePacket(int(ce.CX), int(ce.CZ), ce.Uuid, protocol.EntityTeleport{
					EntityID: ce.ID,
					X:        int32(ce.X * 32),
					Y:        int32(ce.Y * 32),
					Z:        int32(ce.Z * 32),
					Yaw:      int8((ce.Yaw / 360) * 256),
					Pitch:    int8((ce.Pitch / 360) * 256),
				})
				ce.World.QueuePacket(int(ce.CX), int(ce.CZ), ce.Uuid, protocol.EntityHeadLook{
					EntityID: ce.ID,
					HeadYaw:  int8((ce.Yaw / 360) * 256),
				})
			} else if (dx != 0 || dy != 0 || dz != 0) && (dyaw != 0 || dpitch != 0) {
				ce.World.QueuePacket(int(ce.CX), int(ce.CZ), ce.Uuid, protocol.EntityLookMove{
					EntityID: ce.ID,
					DX:       int8(dx * 32),
					DY:       int8(dy * 32),
					DZ:       int8(dz * 32),
					Yaw:      int8((ce.Yaw / 360) * 256),
					Pitch:    int8((ce.Pitch / 360) * 256),
				})
				ce.World.QueuePacket(int(ce.CX), int(ce.CZ), ce.Uuid, protocol.EntityHeadLook{
					EntityID: ce.ID,
					HeadYaw:  int8((ce.Yaw / 360) * 256),
				})
			} else if dx != 0 || dy != 0 || dz != 0 {
				ce.World.QueuePacket(int(ce.CX), int(ce.CZ), ce.Uuid, protocol.EntityMove{
					EntityID: ce.ID,
					DX:       int8(dx * 32),
					DY:       int8(dy * 32),
					DZ:       int8(dz * 32),
				})
			} else if dyaw != 0 || dpitch != 0 {
				ce.World.QueuePacket(int(ce.CX), int(ce.CZ), ce.Uuid, protocol.EntityLook{
					EntityID: ce.ID,
					Yaw:      int8((ce.Yaw / 360) * 256),
					Pitch:    int8((ce.Pitch / 360) * 256),
				})
				ce.World.QueuePacket(int(ce.CX), int(ce.CZ), ce.Uuid, protocol.EntityHeadLook{
					EntityID: ce.ID,
					HeadYaw:  int8((ce.Yaw / 360) * 256),
				})
			}
		}

		ce.LastX, ce.LastY, ce.LastZ = ce.X, ce.Y, ce.Z
		ce.LastYaw, ce.LastPitch = ce.Yaw, ce.Pitch
	}
	ce.currentTick++

	return
}

//Returns the entity's UUID
func (ce *CommonEntity) UUID() string {
	return ce.Uuid
}
