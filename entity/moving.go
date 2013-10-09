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
)

type Moving struct {
	LastCX, LastCZ      int32
	LastX, LastY, LastZ float64
	LastYaw, LastPitch  float32
}

//Updates the entity's movement and moves the chunk its in if required
func (me *Moving) UpdateMovement(super Entity, ce *Common) (movedChunk bool) {
	ce.CX, ce.CZ = int32(ce.X)>>4, int32(ce.Z)>>4
	if ce.CX != me.LastCX || ce.CZ != me.LastCZ {
		movedChunk = true
		//TODO: Move the entity to the next chunk
	}
	me.LastCX, me.LastCZ = ce.CX, ce.CZ

	if ce.CurrentTick%2 == 0 {
		moved := false
		if ce.CurrentTick%(10*5) == 0 {
			moved = true
			ce.World.QueuePacket(int(ce.CX), int(ce.CZ), ce.Uuid, protocol.EntityTeleport{
				EntityID: ce.ID,
				X:        int32(ce.X * 32),
				Y:        int32(ce.Y * 32),
				Z:        int32(ce.Z * 32),
				Yaw:      int8((ce.Yaw / 360) * 256),
				Pitch:    int8((ce.Pitch / 360) * 256),
			})
		} else {
			dx := ce.X - me.LastX
			dy := ce.Y - me.LastY
			dz := ce.Z - me.LastZ
			dyaw := ce.Yaw - me.LastYaw
			dpitch := ce.Pitch - me.LastPitch
			if dx >= 4 || dy >= 4 || dz >= 4 {
				moved = true
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
				moved = true
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
				moved = true
				ce.World.QueuePacket(int(ce.CX), int(ce.CZ), ce.Uuid, protocol.EntityMove{
					EntityID: ce.ID,
					DX:       int8(dx * 32),
					DY:       int8(dy * 32),
					DZ:       int8(dz * 32),
				})
			} else if dyaw != 0 || dpitch != 0 {
				moved = true
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
			if moved {
				ce.World.UpdateSpawnData(int(ce.CX), int(ce.CZ), super, true, super.SpawnPackets())
			}
		}

		me.LastX, me.LastY, me.LastZ = ce.X, ce.Y, ce.Z
		me.LastYaw, me.LastPitch = ce.Yaw, ce.Pitch
	}

	return
}
