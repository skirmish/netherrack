package Soulsand

import ()

type SyncChunk interface {
	//Gets the block in the chunk at the coordinates
	GetBlock(x, y, z int) byte
	//Gets the metadata in the chunk at the coordinates
	GetMeta(x, y, z int) byte
	//Sets the block in the chunk at the coordinates
	SetBlock(x, y, z int, blType byte)
	//Sets the metadata in the chunk at the coordinates
	SetMeta(x, y, z int, data byte)
}
