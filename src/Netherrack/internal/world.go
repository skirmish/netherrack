package internal

import (
	"Soulsand"
)

type World interface {
	Soulsand.World
	//Adds the entity to the chunk
	JoinChunk(x, z int32, e Soulsand.Entity)
	//Removes the entity from the chunk
	LeaveChunk(x, z int32, e Soulsand.Entity)
	//Adds the player to the chunk so that it will recieve events from this chunk
	JoinChunkAsWatcher(x, z int32, pl Soulsand.Player)
	//Removes the player so that will stop recieving events from this chunk
	LeaveChunkAsWatcher(x, z int32, pl Soulsand.Player)
	//Sends a message to the chunk
	SendChunkMessage(x, z, id int32, msg func(Soulsand.SyncPlayer))
	//Send a chunk packet to the channel
	GetChunk(x, z int32, ret chan [][]byte)
}
