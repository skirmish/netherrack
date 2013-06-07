package chunk

import (
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/blocks"
	"time"
)

type lightOperation interface {
	Execute(*Chunk)
}

func getBlockLightLocal(chunk *Chunk, x, y, z int) byte {
	return getLightLocal(chunk, x, y, z, false)
}

func getSkyLightLocal(chunk *Chunk, x, y, z int) byte {
	return getLightLocal(chunk, x, y, z, true)
}

func getLightLocal(chunk *Chunk, x, y, z int, sky bool) (light byte) {
	if sky {
		light = 15
	}
	if x < 0 || x > 15 || z < 0 || z > 15 || y < 0 || y > 255 {
		if x < 0 || x > 15 || z < 0 || z > 15 {
			cx, cz := ((chunk.X<<4)+int32(x))>>4, ((chunk.Z<<4)+int32(z))>>4
			if chunk.World.chunkLoaded(cx, cz) {
				otherChunk := chunk.World.grabChunk(cx, cz)
				if otherChunk == nil {
					return
				}
				ret := make(chan byte, 1)
				select {
				case otherChunk.lightChannel <- chunkLightRequest{
					byte(chunk.X<<4 + int32(x) - cx<<4), byte(y), byte(chunk.Z<<4 + int32(z) - cz<<4),
					!sky,
					ret,
				}:
				default:
					return
				}
				timeOut := time.NewTimer(time.Second)
				defer timeOut.Stop()
				for {
					select {
					case light := <-ret:
						return light
					case req := <-chunk.lightChannel:
						if req.blockLight {
							req.ret <- chunk.BlockLight(int(req.x), int(req.y), int(req.z))
						} else {
							req.ret <- chunk.SkyLight(int(req.x), int(req.y), int(req.z))
						}
					case f := <-chunk.eventChannel:
						f(chunk)
					case <-timeOut.C:
						return
					}
				}
			}
			return
		}
		return
	}
	if sky {
		light = chunk.SkyLight(x, y, z)
	} else {
		light = chunk.BlockLight(x, y, z)
	}
	return
}

type blockLightAdd struct {
	x     int8
	y     byte
	z     int8
	light byte
}

func (bla blockLightAdd) Execute(chunk *Chunk) {
	chunk.SetBlockLight(int(bla.x), int(bla.y), int(bla.z), bla.light)

	chunk.pendingLightOperations.Push(blockLightUpdate{bla.x - 1, bla.y, bla.z})
	chunk.pendingLightOperations.Push(blockLightUpdate{bla.x + 1, bla.y, bla.z})
	chunk.pendingLightOperations.Push(blockLightUpdate{bla.x, bla.y - 1, bla.z})
	chunk.pendingLightOperations.Push(blockLightUpdate{bla.x, bla.y + 1, bla.z})
	chunk.pendingLightOperations.Push(blockLightUpdate{bla.x, bla.y, bla.z - 1})
	chunk.pendingLightOperations.Push(blockLightUpdate{bla.x, bla.y, bla.z + 1})
}

type blockLightUpdate struct {
	x int8
	y byte
	z int8
}

func (blu blockLightUpdate) Execute(chunk *Chunk) {
	if blu.y < 0 || blu.y > 255 {
		return
	}
	if blu.x < 0 || blu.x > 15 || blu.z < 0 || blu.z > 15 {
		x, z := int(blu.x), int(blu.z)
		cx, cz := ((chunk.X<<4)+int32(x))>>4, ((chunk.Z<<4)+int32(z))>>4
		if chunk.World.chunkLoaded(cx, cz) {
			chunk.World.RunSync(int(cx), int(cz), func(oc soulsand.SyncChunk) {
				otherChunk := oc.(*Chunk)
				blu.x = int8(chunk.X<<4 + int32(x) - cx<<4)
				blu.z = int8(chunk.Z<<4 + int32(z) - cz<<4)
				//otherChunk.pendingLightOperations.Push(blu)
				blu.Execute(otherChunk)
			})
			return
		}
		return
	}

	canTravel := [6]bool{}

	var newLight int8
	x, y, z := int(blu.x), int(blu.y), int(blu.z)
	if light := int8(getBlockLightLocal(chunk, x-1, y, z)); light > newLight {
		newLight = light
	} else {
		canTravel[0] = true
	}
	if light := int8(getBlockLightLocal(chunk, x+1, y, z)); light > newLight {
		newLight = light
	} else {
		canTravel[1] = true
	}
	if light := int8(getBlockLightLocal(chunk, x, y-1, z)); light > newLight {
		newLight = light
	} else {
		canTravel[2] = true
	}
	if light := int8(getBlockLightLocal(chunk, x, y+1, z)); light > newLight {
		newLight = light
	} else {
		canTravel[3] = true
	}
	if light := int8(getBlockLightLocal(chunk, x, y, z-1)); light > newLight {
		newLight = light
	} else {
		canTravel[4] = true
	}
	if light := int8(getBlockLightLocal(chunk, x, y, z+1)); light > newLight {
		newLight = light
	} else {
		canTravel[5] = true
	}

	block := blocks.GetBlockById(chunk.Block(x, y, z))

	newLight -= 1 + int8(block.LightFiltered())
	if newLight < 0 {
		return
	}

	if light := chunk.BlockLight(x, y, z); light >= byte(newLight) {
		return
	}

	chunk.SetBlockLight(x, y, z, byte(newLight))

	if canTravel[0] {
		chunk.pendingLightOperations.Push(blockLightUpdate{blu.x - 1, blu.y, blu.z})
	}
	if canTravel[1] {
		chunk.pendingLightOperations.Push(blockLightUpdate{blu.x + 1, blu.y, blu.z})
	}
	if canTravel[2] {
		chunk.pendingLightOperations.Push(blockLightUpdate{blu.x, blu.y - 1, blu.z})
	}
	if canTravel[3] {
		chunk.pendingLightOperations.Push(blockLightUpdate{blu.x, blu.y + 1, blu.z})
	}
	if canTravel[4] {
		chunk.pendingLightOperations.Push(blockLightUpdate{blu.x, blu.y, blu.z - 1})
	}
	if canTravel[5] {
		chunk.pendingLightOperations.Push(blockLightUpdate{blu.x, blu.y, blu.z + 1})
	}
}

type blockLightRemove struct {
	x int8
	y byte
	z int8
}

func (blr blockLightRemove) Execute(chunk *Chunk) {
	chunk.SetBlockLight(int(blr.x), int(blr.y), int(blr.z), 0)

	chunk.pendingLightOperations.Push(blockLightRemoveUpdate{blr.x - 1, blr.y, blr.z})
	chunk.pendingLightOperations.Push(blockLightRemoveUpdate{blr.x + 1, blr.y, blr.z})
	chunk.pendingLightOperations.Push(blockLightRemoveUpdate{blr.x, blr.y - 1, blr.z})
	chunk.pendingLightOperations.Push(blockLightRemoveUpdate{blr.x, blr.y + 1, blr.z})
	chunk.pendingLightOperations.Push(blockLightRemoveUpdate{blr.x, blr.y, blr.z - 1})
	chunk.pendingLightOperations.Push(blockLightRemoveUpdate{blr.x, blr.y, blr.z + 1})
}

type blockLightRemoveUpdate struct {
	x int8
	y byte
	z int8
}

func (blu blockLightRemoveUpdate) Execute(chunk *Chunk) {
	if blu.y < 0 || blu.y > 255 {
		return
	}
	if blu.x < 0 || blu.x > 15 || blu.z < 0 || blu.z > 15 {
		x, z := int(blu.x), int(blu.z)
		cx, cz := ((chunk.X<<4)+int32(x))>>4, ((chunk.Z<<4)+int32(z))>>4
		if chunk.World.chunkLoaded(cx, cz) {
			chunk.World.RunSync(int(cx), int(cz), func(oc soulsand.SyncChunk) {
				otherChunk := oc.(*Chunk)
				blu.x = int8(chunk.X<<4 + int32(x) - cx<<4)
				blu.z = int8(chunk.Z<<4 + int32(z) - cz<<4)
				//otherChunk.pendingLightOperations.Push(blu)
				blu.Execute(otherChunk)
			})
		}
		return
	}

	canTravel := [6]bool{}

	var newLight int8
	x, y, z := int(blu.x), int(blu.y), int(blu.z)

	if light := int8(getBlockLightLocal(chunk, x-1, y, z)); light >= newLight {
		newLight = light
		canTravel[0] = true
	}
	if light := int8(getBlockLightLocal(chunk, x+1, y, z)); light >= newLight {
		newLight = light
		canTravel[1] = true
	}
	if light := int8(getBlockLightLocal(chunk, x, y-1, z)); light >= newLight {
		newLight = light
		canTravel[2] = true
	}
	if light := int8(getBlockLightLocal(chunk, x, y+1, z)); light >= newLight {
		newLight = light
		canTravel[3] = true
	}
	if light := int8(getBlockLightLocal(chunk, x, y, z-1)); light >= newLight {
		newLight = light
		canTravel[4] = true
	}
	if light := int8(getBlockLightLocal(chunk, x, y, z+1)); light >= newLight {
		newLight = light
		canTravel[5] = true
	}
	orgZero := true
	if newLight != 0 {
		orgZero = false
		block := blocks.GetBlockById(chunk.Block(x, y, z))
		if block.LightLevel() > 0 {
			//chunk.pendingLightOperations.Push(blockLightAdd{blu.x, blu.y, blu.z, block.LightLevel()})
			chunk.brokenLights.Add(blockLightAdd{blu.x, blu.y, blu.z, block.LightLevel()})
		}
		newLight -= 1 + int8(block.LightFiltered())
	}
	if newLight < 0 {
		return
	}

	if light := chunk.BlockLight(x, y, z); light <= byte(newLight) {
		return
	}

	chunk.SetBlockLight(x, y, z, byte(newLight))

	if orgZero {
		return
	}

	if canTravel[0] {
		chunk.pendingLightOperations.Push(blockLightRemoveUpdate{blu.x - 1, blu.y, blu.z})
	}
	if canTravel[1] {
		chunk.pendingLightOperations.Push(blockLightRemoveUpdate{blu.x + 1, blu.y, blu.z})
	}
	if canTravel[2] {
		chunk.pendingLightOperations.Push(blockLightRemoveUpdate{blu.x, blu.y - 1, blu.z})
	}
	if canTravel[3] {
		chunk.pendingLightOperations.Push(blockLightRemoveUpdate{blu.x, blu.y + 1, blu.z})
	}
	if canTravel[4] {
		chunk.pendingLightOperations.Push(blockLightRemoveUpdate{blu.x, blu.y, blu.z - 1})
	}
	if canTravel[5] {
		chunk.pendingLightOperations.Push(blockLightRemoveUpdate{blu.x, blu.y, blu.z + 1})
	}
}

type blockRemove struct {
	x int8
	y byte
	z int8
}

func (br blockRemove) Execute(chunk *Chunk) {
	chunk.pendingLightOperations.Push(skyLightUpdate{br.x, br.y, br.z})
	chunk.pendingLightOperations.Push(blockLightUpdate{br.x, br.y, br.z})
}

type blockAdd struct {
	x int8
	y byte
	z int8
}

func (ba blockAdd) Execute(chunk *Chunk) {

	x, y, z := int(ba.x), int(ba.y), int(ba.z)
	block := blocks.GetBlockById(chunk.Block(x, y, z))
	light := int8(chunk.BlockLight(x, y, z)) - int8(block.LightFiltered())
	if light < 0 {
		light = 0
	}
	chunk.SetBlockLight(x, y, z, byte(light))

	light = int8(chunk.SkyLight(x, y, z)) - int8(block.LightFiltered())
	if light < 0 {
		light = 0
	}
	chunk.SetSkyLight(x, y, z, byte(light))

	chunk.pendingLightOperations.Push(blockLightRemoveUpdate{ba.x - 1, ba.y, ba.z})
	chunk.pendingLightOperations.Push(blockLightRemoveUpdate{ba.x + 1, ba.y, ba.z})
	chunk.pendingLightOperations.Push(blockLightRemoveUpdate{ba.x, ba.y - 1, ba.z})
	chunk.pendingLightOperations.Push(blockLightRemoveUpdate{ba.x, ba.y + 1, ba.z})
	chunk.pendingLightOperations.Push(blockLightRemoveUpdate{ba.x, ba.y, ba.z - 1})
	chunk.pendingLightOperations.Push(blockLightRemoveUpdate{ba.x, ba.y, ba.z + 1})

	chunk.pendingLightOperations.Push(skyLightRemoveUpdate{ba.x - 1, ba.y, ba.z})
	chunk.pendingLightOperations.Push(skyLightRemoveUpdate{ba.x + 1, ba.y, ba.z})
	chunk.pendingLightOperations.Push(skyLightRemoveUpdate{ba.x, ba.y - 1, ba.z})
	chunk.pendingLightOperations.Push(skyLightRemoveUpdate{ba.x, ba.y + 1, ba.z})
	chunk.pendingLightOperations.Push(skyLightRemoveUpdate{ba.x, ba.y, ba.z - 1})
	chunk.pendingLightOperations.Push(skyLightRemoveUpdate{ba.x, ba.y, ba.z + 1})
}

type skyLightUpdate struct {
	x int8
	y byte
	z int8
}

func (blu skyLightUpdate) Execute(chunk *Chunk) {
	if blu.y < 0 || blu.y > 255 {
		return
	}
	if blu.x < 0 || blu.x > 15 || blu.z < 0 || blu.z > 15 {
		x, z := int(blu.x), int(blu.z)
		cx, cz := ((chunk.X<<4)+int32(x))>>4, ((chunk.Z<<4)+int32(z))>>4
		if chunk.World.chunkLoaded(cx, cz) {
			chunk.World.RunSync(int(cx), int(cz), func(oc soulsand.SyncChunk) {
				otherChunk := oc.(*Chunk)
				blu.x = int8(chunk.X<<4 + int32(x) - cx<<4)
				blu.z = int8(chunk.Z<<4 + int32(z) - cz<<4)
				//otherChunk.pendingLightOperations.Push(blu)
				blu.Execute(otherChunk)
			})
			return
		}
		return
	}

	canTravel := [6]bool{}

	var newLight int8
	infDown := false

	x, y, z := int(blu.x), int(blu.y), int(blu.z)

	if light := int8(getSkyLightLocal(chunk, x-1, y, z)); light > newLight {
		newLight = light
	} else {
		canTravel[0] = true
	}
	if light := int8(getSkyLightLocal(chunk, x+1, y, z)); light > newLight {
		newLight = light
	} else {
		canTravel[1] = true
	}
	if light := int8(getSkyLightLocal(chunk, x, y-1, z)); light > newLight {
		newLight = light
	} else {
		canTravel[2] = true
	}
	if light := int8(getSkyLightLocal(chunk, x, y+1, z)); light > newLight {
		newLight = light
	} else {
		canTravel[3] = true
	}
	if light := int8(getSkyLightLocal(chunk, x, y, z-1)); light > newLight {
		newLight = light
	} else {
		canTravel[4] = true
	}
	if light := int8(getSkyLightLocal(chunk, x, y, z+1)); light > newLight {
		newLight = light
	} else {
		canTravel[5] = true
	}
	if y >= int(chunk.Height(x, z)) {
		newLight = 15
		infDown = true
	}
	block := blocks.GetBlockById(chunk.Block(x, y, z))

	newLight -= 1 + int8(block.LightFiltered())

	if block.StopsSkylight() {
		infDown = false
	}
	if infDown {
		newLight++
	}
	if newLight < 0 {
		return
	}

	if light := chunk.SkyLight(x, y, z); light >= byte(newLight) {
		return
	}

	chunk.SetSkyLight(x, y, z, byte(newLight))

	if canTravel[0] {
		chunk.pendingLightOperations.Push(skyLightUpdate{blu.x - 1, blu.y, blu.z})
	}
	if canTravel[1] {
		chunk.pendingLightOperations.Push(skyLightUpdate{blu.x + 1, blu.y, blu.z})
	}
	if canTravel[2] {
		chunk.pendingLightOperations.Push(skyLightUpdate{blu.x, blu.y - 1, blu.z})
	}
	if canTravel[3] {
		chunk.pendingLightOperations.Push(skyLightUpdate{blu.x, blu.y + 1, blu.z})
	}
	if canTravel[4] {
		chunk.pendingLightOperations.Push(skyLightUpdate{blu.x, blu.y, blu.z - 1})
	}
	if canTravel[5] {
		chunk.pendingLightOperations.Push(skyLightUpdate{blu.x, blu.y, blu.z + 1})
	}
}

type skyLightRemoveUpdate struct {
	x int8
	y byte
	z int8
}

func (blu skyLightRemoveUpdate) Execute(chunk *Chunk) {
	if blu.y < 0 || blu.y > 255 {
		return
	}
	if blu.x < 0 || blu.x > 15 || blu.z < 0 || blu.z > 15 {
		x, z := int(blu.x), int(blu.z)
		cx, cz := ((chunk.X<<4)+int32(x))>>4, ((chunk.Z<<4)+int32(z))>>4
		if chunk.World.chunkLoaded(cx, cz) {
			chunk.World.RunSync(int(cx), int(cz), func(oc soulsand.SyncChunk) {
				otherChunk := oc.(*Chunk)
				blu.x = int8(chunk.X<<4 + int32(x) - cx<<4)
				blu.z = int8(chunk.Z<<4 + int32(z) - cz<<4)
				//otherChunk.pendingLightOperations.Push(blu)
				blu.Execute(otherChunk)
			})
		}
		return
	}

	canTravel := [6]bool{}

	var newLight int8
	infDown := false
	x, y, z := int(blu.x), int(blu.y), int(blu.z)

	if light := int8(getSkyLightLocal(chunk, x-1, y, z)); light >= newLight {
		newLight = light
		canTravel[0] = true
	}
	if light := int8(getSkyLightLocal(chunk, x+1, y, z)); light >= newLight {
		newLight = light
		canTravel[1] = true
	}
	if light := int8(getSkyLightLocal(chunk, x, y-1, z)); light >= newLight {
		newLight = light
		canTravel[2] = true
	}
	if light := int8(getSkyLightLocal(chunk, x, y+1, z)); light >= newLight {
		newLight = light
		canTravel[3] = true
	}
	if light := int8(getSkyLightLocal(chunk, x, y, z-1)); light >= newLight {
		newLight = light
		canTravel[4] = true
	}
	if light := int8(getSkyLightLocal(chunk, x, y, z+1)); light >= newLight {
		newLight = light
		canTravel[5] = true
	}

	if y >= int(chunk.Height(x, z)) {
		newLight = 15
		infDown = true
	}
	orgZero := true
	if newLight != 0 {
		orgZero = false
		block := blocks.GetBlockById(chunk.Block(x, y, z))
		newLight -= 1 + int8(block.LightFiltered())
	}
	if infDown && newLight == 14 {
		newLight++
	}

	if newLight < 0 {
		return
	}

	if light := chunk.SkyLight(x, y, z); light <= byte(newLight) {
		return
	}

	chunk.SetSkyLight(x, y, z, byte(newLight))

	if orgZero {
		return
	}

	if canTravel[0] {
		chunk.pendingLightOperations.Push(skyLightRemoveUpdate{blu.x - 1, blu.y, blu.z})
	}
	if canTravel[1] {
		chunk.pendingLightOperations.Push(skyLightRemoveUpdate{blu.x + 1, blu.y, blu.z})
	}
	if canTravel[2] {
		chunk.pendingLightOperations.Push(skyLightRemoveUpdate{blu.x, blu.y - 1, blu.z})
	}
	if canTravel[3] {
		chunk.pendingLightOperations.Push(skyLightRemoveUpdate{blu.x, blu.y + 1, blu.z})
	}
	if canTravel[4] {
		chunk.pendingLightOperations.Push(skyLightRemoveUpdate{blu.x, blu.y, blu.z - 1})
	}
	if canTravel[5] {
		chunk.pendingLightOperations.Push(skyLightRemoveUpdate{blu.x, blu.y, blu.z + 1})
	}
}
