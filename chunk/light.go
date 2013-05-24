package chunk

import (
	"github.com/NetherrackDev/soulsand/blocks"
)

type lightInfo struct {
	x     int
	y     int
	z     int
	light byte
	next  *lightInfo
	root  *lightInfo
}

func (l *lightInfo) Append(l2 *lightInfo) *lightInfo {
	if l != nil {
		l.next = l2
		l2.root = l.root
	} else {
		l2.root = l2
	}
	return l2
}

func (chunk *Chunk) Relight() {
	//Clear lights & Sky lights

	var skyLightQueue *lightInfo
	var blockLightQueue *lightInfo

	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			height := chunk.heightMap[x|z<<4]
			max := 256 - height
			if x > 0 {
				if nHeight := 256 - chunk.heightMap[(x-1)|z<<4]; nHeight < max {
					max = nHeight
				}
			}
			if x < 15 {
				if nHeight := 256 - chunk.heightMap[(x+1)|z<<4]; nHeight < max {
					max = nHeight
				}
			}
			if z > 0 {
				if nHeight := 256 - chunk.heightMap[x|(z-1)<<4]; nHeight < max {
					max = nHeight
				}
			}
			if z < 15 {
				if nHeight := 256 - chunk.heightMap[x|(z+1)<<4]; nHeight < max {
					max = nHeight
				}
			}
			maxI := 256 - int(max)
			heightI := int(height)
			for y := heightI; y < 256; y++ {
				if y <= maxI {
					skyLightQueue = skyLightQueue.Append(&lightInfo{
						x:     x,
						y:     y,
						z:     z,
						light: 15,
					})
					chunk.SetSkyLight(x, y, z, 0)
				} else {
					chunk.SetSkyLight(x, y, z, 15)
				}
				chunk.SetBlockLight(x, y, z, 0)
			}
			for y := 0; y < heightI; y++ {
				chunk.SetSkyLight(x, y, z, 0)
				chunk.SetBlockLight(x, y, z, 0)
			}
		}
	}

	for bp, light := range chunk.lights {
		x, y, z := bp.GetPosition()
		blockLightQueue = blockLightQueue.Append(&lightInfo{
			x:     x,
			y:     y,
			z:     z,
			light: light,
		})
	}

	if skyLightQueue != nil {
		current := skyLightQueue.root
		for ; current != nil; current = current.next {
			info := current
			info.root = nil
			x := info.x
			z := info.z
			y := info.y
			light := info.light

			if chunk.GetSkyLight(x, y, z) >= light {
				continue
			}

			chunk.SetSkyLight(x, y, z, light)

			if y > 0 || y < 255 {
				block := blocks.GetBlockById(chunk.GetBlock(x, y-1, z))
				newLight := int8(light) - int8(block.LightFiltered())
				if (newLight == 15 && block.StopsSkylight()) || chunk.GetSkyLight(x, y+1, z) != 15 {
					newLight--
				}
				if int8(chunk.GetSkyLight(x, y-1, z)) < newLight {
					skyLightQueue = skyLightQueue.Append(&lightInfo{
						x:     x,
						y:     y - 1,
						z:     z,
						light: byte(newLight),
					})
				}
			}
			if y < 255 {
				skyLightQueue = chunk.checkSkyLight(skyLightQueue, light, x, y, z, 0, 1, 0)
			}

			if x > 0 {
				skyLightQueue = chunk.checkSkyLight(skyLightQueue, light, x, y, z, -1, 0, 0)
			}
			if x < 15 {
				skyLightQueue = chunk.checkSkyLight(skyLightQueue, light, x, y, z, 1, 0, 0)
			}

			if z > 0 {
				skyLightQueue = chunk.checkSkyLight(skyLightQueue, light, x, y, z, 0, 0, -1)
			}
			if z < 15 {
				skyLightQueue = chunk.checkSkyLight(skyLightQueue, light, x, y, z, 0, 0, 1)
			}
		}
	}

	if blockLightQueue != nil {
		current := blockLightQueue.root
		for ; current != nil; current = current.next {
			info := current
			info.root = nil
			x := info.x
			z := info.z
			y := info.y
			light := info.light

			if chunk.GetBlockLight(x, y, z) >= light {
				continue
			}

			chunk.SetBlockLight(x, y, z, light)

			if y > 0 {
				blockLightQueue = chunk.checkBlockLight(blockLightQueue, light, x, y, z, 0, -1, 0)
			}
			if y < 255 {
				blockLightQueue = chunk.checkBlockLight(blockLightQueue, light, x, y, z, 0, 1, 0)
			}

			if x > 0 {
				blockLightQueue = chunk.checkBlockLight(blockLightQueue, light, x, y, z, -1, 0, 0)
			}
			if x < 15 {
				blockLightQueue = chunk.checkBlockLight(blockLightQueue, light, x, y, z, 1, 0, 0)
			}

			if z > 0 {
				blockLightQueue = chunk.checkBlockLight(blockLightQueue, light, x, y, z, 0, 0, -1)
			}
			if z < 15 {
				blockLightQueue = chunk.checkBlockLight(blockLightQueue, light, x, y, z, 0, 0, 1)
			}
		}
	}
	chunk.needsRelight = false
}

func (chunk *Chunk) checkBlockLightRemove(blockRemoveLightQueue *lightInfo, light byte, x, y, z, ox, oy, oz int) *lightInfo {
	block := blocks.GetBlockById(chunk.GetBlock(x+ox, y+oy, z+oz))
	newLight := int8(light) - int8(block.LightFiltered()) - 1
	if int8(chunk.GetBlockLight(x+ox, y+oy, z+oz)) <= newLight {
		blockRemoveLightQueue = blockRemoveLightQueue.Append(&lightInfo{
			x:     x + ox,
			y:     y + oy,
			z:     z + oz,
			light: byte(newLight),
		})
	}
	return blockRemoveLightQueue
}

func (chunk *Chunk) checkSkyLight(skyLightQueue *lightInfo, light byte, x, y, z, ox, oy, oz int) *lightInfo {
	block := blocks.GetBlockById(chunk.GetBlock(x+ox, y+oy, z+oz))
	newLight := int8(light) - int8(block.LightFiltered()) - 1
	if int8(chunk.GetSkyLight(x+ox, y+oy, z+oz)) < newLight {
		skyLightQueue = skyLightQueue.Append(&lightInfo{
			x:     x + ox,
			y:     y + oy,
			z:     z + oz,
			light: byte(newLight),
		})
	}
	return skyLightQueue
}

func (chunk *Chunk) checkBlockLight(blockLightQueue *lightInfo, light byte, x, y, z, ox, oy, oz int) *lightInfo {
	block := blocks.GetBlockById(chunk.GetBlock(x+ox, y+oy, z+oz))
	newLight := int8(light) - int8(block.LightFiltered()) - 1
	if int8(chunk.GetBlockLight(x+ox, y+oy, z+oz)) < newLight {
		blockLightQueue = blockLightQueue.Append(&lightInfo{
			x:     x + ox,
			y:     y + oy,
			z:     z + oz,
			light: byte(newLight),
		})
	}
	return blockLightQueue
}
