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
	"github.com/NetherrackDev/netherrack/world"
)

type CommonEntity struct {
	//Exported for setting when embedded
	Server Server
	//Exported for setting when embedded
	World *world.World

	CX, CZ, LastCX, LastCZ int32
	X, Y, Z                float64
	Yaw, Pitch             float32
}

//Updates the entity's movement and moves the chunk its in if required
func (ce *CommonEntity) UpdateMovement(super Entity) (movedChunk bool) {
	ce.CX, ce.CZ = int32(ce.X)>>4, int32(ce.Z)>>4
	if ce.CX != ce.LastCX || ce.CZ != ce.LastCZ {
		movedChunk = true
		//TODO: Move the entity to the next chunk
	}
	ce.LastCX, ce.LastCZ = ce.CX, ce.CZ
	return
}