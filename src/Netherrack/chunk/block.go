package chunk

import ()

type Blocks struct {
	raw     []byte
	rawMeta []byte
	X, Y, Z int
}

func (blocks *Blocks) GetBlock(x, y, z int) byte {
	return blocks.raw[x+(z*blocks.X)+(y*blocks.X*blocks.Z)]
}

func (blocks *Blocks) setBlock(x, y, z int, b byte) {
	blocks.raw[x+(z*blocks.X)+(y*blocks.X*blocks.Z)] = b
}

func (blocks *Blocks) GetMeta(x, y, z int) byte {
	return blocks.rawMeta[x+(z*blocks.X)+(y*blocks.X*blocks.Z)]
}

func (blocks *Blocks) setMeta(x, y, z int, b byte) {
	blocks.rawMeta[x+(z*blocks.X)+(y*blocks.X*blocks.Z)] = b
}

func (blocks *Blocks) GetRaw() []byte {
	return blocks.raw
}
