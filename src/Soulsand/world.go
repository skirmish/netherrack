package Soulsand

import ()

type World interface {
	//Gets the block & meta at the coordinates
	GetBlock(x, y, z int) []byte
}
