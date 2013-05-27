package chunk

import (
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/blocks"
)

func (chunk *Chunk) Relight() {
	//Clear lights & Sky lights

	north := make(map[uint16]byte)
	south := make(map[uint16]byte)
	east := make(map[uint16]byte)
	west := make(map[uint16]byte)

	northSky := make(map[uint16]byte)
	southSky := make(map[uint16]byte)
	eastSky := make(map[uint16]byte)
	westSky := make(map[uint16]byte)

	var skyLightQueue *lightInfo
	var blockLightQueue *lightInfo

	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			height := chunk.heightMap[x|z<<4]
			max := 256 - height

			if nHeight := 256 - chunk.HeightX(x-1, z); nHeight < max {
				max = nHeight
			}
			if nHeight := 256 - chunk.HeightX(x+1, z); nHeight < max {
				max = nHeight
			}
			if nHeight := 256 - chunk.HeightX(x, z-1); nHeight < max {
				max = nHeight
			}
			if nHeight := 256 - chunk.HeightX(x, z+1); nHeight < max {
				max = nHeight
			}

			maxI := 256 - int(max)
			heightI := int(height)
			for y := heightI; y < 256; y++ {
				if y <= maxI {
					skyLightQueue = skyLightQueue.Append(LightInfoGet(x, y, z, 15))
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

	for bp, light := range chunk.skyLights {
		x, y, z := bp.Position()
		skyLightQueue = skyLightQueue.Append(LightInfoGet(x, y, z, light))
	}

	for bp, light := range chunk.lights {
		x, y, z := bp.Position()
		blockLightQueue = blockLightQueue.Append(LightInfoGet(x, y, z, light))
	}

	if skyLightQueue != nil {
		current := skyLightQueue.root
		var next *lightInfo
		for ; current != nil; current = next {
			info := current
			x := info.x
			z := info.z
			y := info.y
			light := info.light
			next = info.next
			info.Free()

			if chunk.SkyLight(x, y, z) >= light {
				continue
			}
			chunk.SetSkyLight(x, y, z, light)

			block := blocks.GetBlockById(chunk.Block(x, y, z))
			newLight := int8(light) - int8(block.LightFiltered()) - 1
			if newLight <= 0 {
				continue
			}

			if y > 0 && y < 255 {
				block := blocks.GetBlockById(chunk.Block(x, y-1, z))
				newerLight := int8(light) - int8(block.LightFiltered())
				if (newerLight == 15 && block.StopsSkylight()) || chunk.SkyLight(x, y+1, z) != 15 {
					newerLight--
				}
				if int8(chunk.SkyLight(x, y-1, z)) < newerLight {
					skyLightQueue = skyLightQueue.Append(LightInfoGet(x, y-1, z, byte(newerLight)))
				}
			}
			if y < 255 {
				skyLightQueue = chunk.checkSkyLight(skyLightQueue, newLight, x, y, z, 0, 1, 0)
			}

			if x > 0 {
				skyLightQueue = chunk.checkSkyLight(skyLightQueue, newLight, x, y, z, -1, 0, 0)
			} else {
				pos := uint16(z | y<<4)
				if oLight := int8(westSky[pos]); oLight < newLight {
					westSky[pos] = byte(newLight)
				}
			}
			if x < 15 {
				skyLightQueue = chunk.checkSkyLight(skyLightQueue, newLight, x, y, z, 1, 0, 0)
			} else {
				pos := uint16(z | y<<4)
				if oLight := int8(eastSky[pos]); oLight < newLight {
					eastSky[pos] = byte(newLight)
				}
			}

			if z > 0 {
				skyLightQueue = chunk.checkSkyLight(skyLightQueue, newLight, x, y, z, 0, 0, -1)
			} else {
				pos := uint16(x | y<<4)
				if oLight := int8(northSky[pos]); oLight < newLight {
					northSky[pos] = byte(newLight)
				}
			}
			if z < 15 {
				skyLightQueue = chunk.checkSkyLight(skyLightQueue, newLight, x, y, z, 0, 0, 1)
			} else {
				pos := uint16(x | y<<4)
				if oLight := int8(southSky[pos]); oLight < newLight {
					southSky[pos] = byte(newLight)
				}
			}
		}
	}

	if blockLightQueue != nil {
		current := blockLightQueue.root
		var next *lightInfo
		for ; current != nil; current = next {
			info := current
			x := info.x
			z := info.z
			y := info.y
			light := info.light
			next = info.next
			info.Free()

			if chunk.BlockLight(x, y, z) >= light {
				continue
			}

			chunk.SetBlockLight(x, y, z, light)

			block := blocks.GetBlockById(chunk.Block(x, y, z))
			newLight := int8(light) - int8(block.LightFiltered()) - 1
			if newLight <= 0 {
				continue
			}

			if y > 0 {
				blockLightQueue = chunk.checkBlockLight(blockLightQueue, newLight, x, y, z, 0, -1, 0)
			}
			if y < 255 {
				blockLightQueue = chunk.checkBlockLight(blockLightQueue, newLight, x, y, z, 0, 1, 0)
			}

			if x > 0 {
				blockLightQueue = chunk.checkBlockLight(blockLightQueue, newLight, x, y, z, -1, 0, 0)
			} else {
				pos := uint16(z | y<<4)
				if oLight := int8(west[pos]); oLight < newLight {
					west[pos] = byte(newLight)
				}
			}
			if x < 15 {
				blockLightQueue = chunk.checkBlockLight(blockLightQueue, newLight, x, y, z, 1, 0, 0)
			} else {
				pos := uint16(z | y<<4)
				if oLight := int8(east[pos]); oLight < newLight {
					east[pos] = byte(newLight)
				}
			}

			if z > 0 {
				blockLightQueue = chunk.checkBlockLight(blockLightQueue, newLight, x, y, z, 0, 0, -1)
			} else {
				pos := uint16(x | y<<4)
				if oLight := int8(north[pos]); oLight < newLight {
					north[pos] = byte(newLight)
				}
			}
			if z < 15 {
				blockLightQueue = chunk.checkBlockLight(blockLightQueue, newLight, x, y, z, 0, 0, 1)
			} else {
				pos := uint16(x | y<<4)
				if oLight := int8(south[pos]); oLight < newLight {
					south[pos] = byte(newLight)
				}
			}
		}
	}

	depth := 0
	if chunk.relightDepth == 0 {
		depth = 2
	} else if chunk.relightDepth == 2 {
		depth = 1
	}

	if chunk.relightDepth != 1 && chunk.World.chunkLoaded(chunk.X, chunk.Z-1) { //North
		oldNorth := chunk.lightInfo.north
		newNorth := north //No need for locks as the map will never be changed past this point
		oldNorthSky := chunk.lightInfo.northSky
		newNorthSky := northSky
		chunk.World.RunSync(int(chunk.X), int(chunk.Z-1), func(c soulsand.SyncChunk) {
			otherChunk := c.(*Chunk)
			if otherChunk.relightDepth < depth {
				otherChunk.relightDepth = depth
			}
			relight := false
			for pos, _ := range oldNorth {
				x := pos & 0xF
				y := pos >> 4
				delete(otherChunk.lights, createBlockPosition(int(x), int(y), 15))
			}
			for pos, light := range newNorth {
				if oLight := oldNorth[pos]; oLight != light {
					relight = true
				}
				x := pos & 0xF
				y := pos >> 4
				otherChunk.lights[createBlockPosition(int(x), int(y), 15)] = light
			}
			for pos, _ := range oldNorthSky {
				x := pos & 0xF
				y := pos >> 4
				delete(otherChunk.skyLights, createBlockPosition(int(x), int(y), 15))
			}
			for pos, light := range newNorthSky {
				if oLight := oldNorthSky[pos]; oLight != light {
					relight = true
				}
				x := pos & 0xF
				y := pos >> 4
				otherChunk.skyLights[createBlockPosition(int(x), int(y), 15)] = light
			}
			if relight {
				otherChunk.needsRelight = true
			}
		})
	}
	if chunk.relightDepth != 1 && chunk.World.chunkLoaded(chunk.X, chunk.Z+1) { //South
		oldSouth := chunk.lightInfo.south
		newSouth := south
		oldSouthSky := chunk.lightInfo.southSky
		newSouthSky := southSky
		chunk.World.RunSync(int(chunk.X), int(chunk.Z+1), func(c soulsand.SyncChunk) {
			otherChunk := c.(*Chunk)
			if otherChunk.relightDepth < depth {
				otherChunk.relightDepth = depth
			}
			relight := false
			for pos, _ := range oldSouth {
				x := pos & 0xF
				y := pos >> 4
				delete(otherChunk.lights, createBlockPosition(int(x), int(y), 0))
			}
			for pos, light := range newSouth {
				if oLight := oldSouth[pos]; oLight != light {
					relight = true
				}
				x := pos & 0xF
				y := pos >> 4
				otherChunk.lights[createBlockPosition(int(x), int(y), 0)] = light
			}
			for pos, _ := range oldSouthSky {
				x := pos & 0xF
				y := pos >> 4
				delete(otherChunk.skyLights, createBlockPosition(int(x), int(y), 15))
			}
			for pos, light := range newSouthSky {
				if oLight := oldSouthSky[pos]; oLight != light {
					relight = true
				}
				x := pos & 0xF
				y := pos >> 4
				otherChunk.skyLights[createBlockPosition(int(x), int(y), 15)] = light
			}
			if relight {
				otherChunk.needsRelight = true
			}
		})
	}
	if chunk.relightDepth != 1 && chunk.World.chunkLoaded(chunk.X+1, chunk.Z) { //East
		oldEast := chunk.lightInfo.east
		newEast := east
		oldEastSky := chunk.lightInfo.eastSky
		newEastSky := eastSky
		chunk.World.RunSync(int(chunk.X+1), int(chunk.Z), func(c soulsand.SyncChunk) {
			otherChunk := c.(*Chunk)
			if otherChunk.relightDepth < depth {
				otherChunk.relightDepth = depth
			}
			relight := false
			for pos, _ := range oldEast {
				z := pos & 0xF
				y := pos >> 4
				delete(otherChunk.lights, createBlockPosition(0, int(y), int(z)))
			}
			for pos, light := range newEast {
				if oLight := oldEast[pos]; oLight != light {
					relight = true
				}
				z := pos & 0xF
				y := pos >> 4
				otherChunk.lights[createBlockPosition(0, int(y), int(z))] = light
			}
			for pos, _ := range oldEastSky {
				x := pos & 0xF
				y := pos >> 4
				delete(otherChunk.skyLights, createBlockPosition(int(x), int(y), 15))
			}
			for pos, light := range newEastSky {
				if oLight := oldEastSky[pos]; oLight != light {
					relight = true
				}
				x := pos & 0xF
				y := pos >> 4
				otherChunk.skyLights[createBlockPosition(int(x), int(y), 15)] = light
			}
			if relight {
				otherChunk.needsRelight = true
			}
		})
	}
	if chunk.relightDepth != 1 && chunk.World.chunkLoaded(chunk.X-1, chunk.Z) { //West
		oldWest := chunk.lightInfo.west
		newWest := west
		oldWestSky := chunk.lightInfo.westSky
		newWestSky := westSky
		chunk.World.RunSync(int(chunk.X-1), int(chunk.Z), func(c soulsand.SyncChunk) {
			otherChunk := c.(*Chunk)
			if otherChunk.relightDepth < depth {
				otherChunk.relightDepth = depth
			}
			relight := false
			for pos, _ := range oldWest {
				z := pos & 0xF
				y := pos >> 4
				delete(otherChunk.lights, createBlockPosition(15, int(y), int(z)))
			}
			for pos, light := range newWest {
				if oLight := oldWest[pos]; oLight != light {
					relight = true
				}
				z := pos & 0xF
				y := pos >> 4
				otherChunk.lights[createBlockPosition(15, int(y), int(z))] = light
			}
			for pos, _ := range oldWestSky {
				x := pos & 0xF
				y := pos >> 4
				delete(otherChunk.skyLights, createBlockPosition(int(x), int(y), 15))
			}
			for pos, light := range newWestSky {
				if oLight := oldWestSky[pos]; oLight != light {
					relight = true
				}
				x := pos & 0xF
				y := pos >> 4
				otherChunk.skyLights[createBlockPosition(int(x), int(y), 15)] = light
			}
			if relight {
				otherChunk.needsRelight = true
			}
		})
	}

	chunk.lightInfo.north = north
	chunk.lightInfo.south = south
	chunk.lightInfo.east = east
	chunk.lightInfo.west = west

	chunk.lightInfo.northSky = northSky
	chunk.lightInfo.southSky = southSky
	chunk.lightInfo.eastSky = eastSky
	chunk.lightInfo.westSky = westSky

	chunk.relightDepth = 0
	chunk.needsRelight = false
}

func (chunk *Chunk) checkSkyLight(skyLightQueue *lightInfo, light int8, x, y, z, ox, oy, oz int) *lightInfo {
	// block := blocks.GetBlockById(chunk.Block(x+ox, y+oy, z+oz))
	// newLight := int8(light) - int8(block.LightFiltered()) - 1
	if int8(chunk.SkyLight(x+ox, y+oy, z+oz)) < light {
		skyLightQueue = skyLightQueue.Append(LightInfoGet(x+ox, y+oy, z+oz, byte(light)))
	}
	return skyLightQueue
}

func (chunk *Chunk) checkBlockLight(blockLightQueue *lightInfo, light int8, x, y, z, ox, oy, oz int) *lightInfo {
	// block := blocks.GetBlockById(chunk.Block(x+ox, y+oy, z+oz))
	// newLight := int8(light) - int8(block.LightFiltered()) - 1
	if int8(chunk.BlockLight(x+ox, y+oy, z+oz)) < light {
		blockLightQueue = blockLightQueue.Append(LightInfoGet(x+ox, y+oy, z+oz, byte(light)))
	}
	return blockLightQueue
}
