package Soulsand

import ()

type World interface {
	//Adds the entity to the chunk
	JoinChunk(x, z int32, e Entity)
	//Removes the entity from the chunk
	LeaveChunk(x, z int32, e Entity)
	//Adds the player to the chunk so that it will recieve events from this chunk
	JoinChunkAsWatcher(x, z int32, pl Player)
	//Removes the player so that will stop recieving events from this chunk
	LeaveChunkAsWatcher(x, z int32, pl Player)
	//Sends a message to the chunk
	SendChunkMessage(x, z, id int32, msg func(SyncPlayer))
	//Send a chunk packet to the channel
	GetChunk(x, z int32, ret chan [][]byte)
}
