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
	"github.com/NetherrackDev/netherrack/entity"
	"github.com/NetherrackDev/netherrack/protocol"
	"github.com/NetherrackDev/netherrack/world"
)

//Player is an interface that covers general methods for a player
//wether they are local or not
type Player interface {
	entity.Entity

	//Queues a packet to be sent to the player
	QueuePacket(packet protocol.Packet)
}

//Contains methods that a player needs for a server
type Server interface {
	entity.Server
	//Returns the default world for the server
	DefaultWorld() world.World
	//Gets the world by name, loads the world if it isn't loaded
	World(name string) world.World
}
