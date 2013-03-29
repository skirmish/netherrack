package Soulsand

import ()

type SyncChunk interface {
	//Gets the block in a the chunk at the coordinates
	GetBlock(x, y, z int) byte
	//Gets the metadata in a the chunk at the coordinates
	GetMeta(x, y, z int) byte
}
