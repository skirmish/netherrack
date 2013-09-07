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
