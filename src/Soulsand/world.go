package Soulsand

import ()

type World interface {
	//Gets the block & meta at the coordinates
	GetBlock(x, y, z int) []byte
	//Gets the block & meta at the coordinates
	SetBlock(x, y, z int, block, meta byte)
	//Run the function on the chunk at the coordinates
	RunSync(x, z int, f func(SyncChunk))
}
