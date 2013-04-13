package chunk

import (
	"bitbucket.org/Thinkofdeath/soulsand/blocks"
)

func generateChunk(chunk *Chunk) {
	for y := 0; y < 256; y++ {
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				if y <= 64 {
					chunk.SetBlock(x, y, z, blocks.Wool)
					if x == 0 || x == 15 || z == 0 || z == 15 {
						chunk.SetMeta(x, y, z, 1)
					} else {
						chunk.SetMeta(x, y, z, byte((y>>4)+4))
					}
				} else {
					chunk.SetSkyLight(x, y, z, 15)
					chunk.SetBlockLight(x, y, z, 15)
				}
			}
		}
	}
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			chunk.SetBiome(x, z, 1)
		}
	}
}
