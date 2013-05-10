package internal

import (
	"github.com/NetherrackDev/soulsand"
)

type World interface {
	soulsand.World
	//Adds the entity to the chunk
	JoinChunk(x, z int32, e soulsand.Entity)
	//Removes the entity from the chunk
	LeaveChunk(x, z int32, e soulsand.Entity)
	//Adds the player to the chunk so that it will recieve events from this chunk
	JoinChunkAsWatcher(x, z int32, pl soulsand.Player)
	//Removes the player so that will stop recieving events from this chunk
	LeaveChunkAsWatcher(x, z int32, pl soulsand.Player)
	//Sends a message to the chunk
	SendChunkMessage(x, z, id int32, msg func(soulsand.SyncEntity))
	//Send a chunk packet to the channel
	GetChunk(x, z int32, ret chan [][]byte, stop chan struct{})

	AddPlayer(player soulsand.Player)

	RemovePlayer(player soulsand.Player)
}
