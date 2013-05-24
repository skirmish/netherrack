package chunk

import (
	"github.com/NetherrackDev/soulsand/blocks"
	//"log"
	//"time"
)

type lightInfo struct {
	x, y, z int
	light   byte
}

func (chunk *Chunk) Relight() {

	//start := time.Now()

	skyLights := make([]blockPosition, 0, 16*16*150)

	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			for y := 0; y < 256; y++ {
				if y >= int(chunk.heightMap[x|z<<4]) {
					skyLights = append(skyLights, createBlockPosition(x, y, z))
					chunk.SetSkyLight(x, y, z, 15)
				} else {
					chunk.SetSkyLight(x, y, z, 0)
				}
				chunk.SetBlockLight(x, y, z, 0)
			}
		}
	}

	//log.Printf("Clear: %.2f", float64(time.Now().Sub(start).Nanoseconds())/1000000.0)
	//start = time.Now()

	for pos, light := range chunk.lights {
		x, y, z := pos.GetPosition()
		chunk.SetBlockLight(x, y, z, light)
		chunk.propagateBlockLight(x, y, z, light)

		chunk.propagateBlockLight(x, y-1, z, light-1)
		chunk.propagateBlockLight(x, y+1, z, light-1)
		chunk.propagateBlockLight(x-1, y, z, light-1)
		chunk.propagateBlockLight(x+1, y, z, light-1)
		chunk.propagateBlockLight(x, y, z-1, light-1)
		chunk.propagateBlockLight(x, y, z+1, light-1)
	}

	//log.Printf("BlockLight: %.2f", float64(time.Now().Sub(start).Nanoseconds())/1000000.0)
	//start = time.Now()

	stack := make([]lightInfo, 0, 16*16*16)

	for len(skyLights) > 0 {
		x, y, z := skyLights[0].GetPosition()
		skyLights = skyLights[1:]

		chunk.propagateSkyLight(&stack, x, y-1, z, 15)
		chunk.propagateSkyLight(&stack, x, y+1, z, 14)
		chunk.propagateSkyLight(&stack, x-1, y, z, 14)
		chunk.propagateSkyLight(&stack, x+1, y, z, 14)
		chunk.propagateSkyLight(&stack, x, y, z-1, 14)
		chunk.propagateSkyLight(&stack, x, y, z+1, 14)

		for len(stack) > 0 {
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			chunk.propagateSkyLight(&stack, current.x, current.y, current.z, current.light)
		}
	}

	//log.Printf("SkyLight: %.2f", float64(time.Now().Sub(start).Nanoseconds())/1000000.0)

	chunk.needsRelight = false
}

func (chunk *Chunk) propagateBlockLight(x, y, z int, light byte) {
	if y < 0 || y > 255 || x < 0 || x > 15 || z < 0 || z > 15 {
		return
	}
	block := blocks.GetBlockById(chunk.GetBlock(x, y, z))
	light = light - block.LightFiltered()
	if light == 0 || light > 15 || light <= chunk.GetBlockLight(x, y, z) {
		return
	}
	chunk.SetBlockLight(x, y, z, light)

	chunk.propagateBlockLight(x, y-1, z, light-1)
	chunk.propagateBlockLight(x, y+1, z, light-1)
	chunk.propagateBlockLight(x-1, y, z, light-1)
	chunk.propagateBlockLight(x+1, y, z, light-1)
	chunk.propagateBlockLight(x, y, z-1, light-1)
	chunk.propagateBlockLight(x, y, z+1, light-1)
}

func (chunk *Chunk) propagateSkyLight(stack *[]lightInfo, x, y, z int, light byte) {
	if y < 0 || y > 255 || x < 0 || x > 15 || z < 0 || z > 15 {
		return
	}
	block := blocks.GetBlockById(chunk.GetBlock(x, y, z))
	light = light - block.LightFiltered()
	if light == 0 || light > 15 || light <= chunk.GetSkyLight(x, y, z) {
		return
	}
	chunk.SetSkyLight(x, y, z, light)

	stackP := *stack

	if light == 15 && !block.StopsSkylight() {
		stackP = append(stackP, lightInfo{x, y - 1, z, light})
	} else {
		stackP = append(stackP, lightInfo{x, y - 1, z, light - 1})
	}
	stackP = append(stackP, lightInfo{x, y + 1, z, light - 1})
	stackP = append(stackP, lightInfo{x - 1, y, z, light - 1})
	stackP = append(stackP, lightInfo{x + 1, y, z, light - 1})
	stackP = append(stackP, lightInfo{x, y, z - 1, light - 1})
	stackP = append(stackP, lightInfo{x, y, z + 1, light - 1})
}
