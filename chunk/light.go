package chunk

import (
	"github.com/NetherrackDev/netherrack/debug"
	"github.com/NetherrackDev/soulsand/blocks"
	"os"
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
	debug.Start("SkyLight")
	stack := make([]lightInfo, 16*16*16)
	stackPointer := 0

	for len(skyLights) > 0 {
		x, y, z := skyLights[0].GetPosition()
		skyLights = skyLights[1:]

		stackPointer = chunk.propagateSkyLight(stack, stackPointer, x, y-1, z, 15)
		stackPointer = chunk.propagateSkyLight(stack, stackPointer, x, y+1, z, 14)
		stackPointer = chunk.propagateSkyLight(stack, stackPointer, x-1, y, z, 14)
		stackPointer = chunk.propagateSkyLight(stack, stackPointer, x+1, y, z, 14)
		stackPointer = chunk.propagateSkyLight(stack, stackPointer, x, y, z-1, 14)
		stackPointer = chunk.propagateSkyLight(stack, stackPointer, x, y, z+1, 14)

		debug.StepIn("StackLoop")
		for stackPointer != 0 {
			stackPointer--
			current := stack[stackPointer]
			stackPointer = chunk.propagateSkyLight(stack, stackPointer, current.x, current.y, current.z, current.light)
		}
		debug.StepOut()
	}
	debug.Stop()
	f, _ := os.Create("relight.txt")
	defer f.Close()
	debug.Print(f, true)

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

func (chunk *Chunk) propagateSkyLight(stack []lightInfo, stackPointer, x, y, z int, light byte) int {
	debug.StepIn("Propagate SkyLight")
	defer debug.StepOut()
	if y < 0 || y > 255 || x < 0 || x > 15 || z < 0 || z > 15 {
		return stackPointer
	}
	block := blocks.GetBlockById(chunk.GetBlock(x, y, z))
	light = light - block.LightFiltered()
	if light == 0 || light > 15 || light <= chunk.GetSkyLight(x, y, z) {
		return stackPointer
	}
	chunk.SetSkyLight(x, y, z, light)

	if light == 15 && !block.StopsSkylight() {
		//stackP = append(stackP, lightInfo{x, y - 1, z, light})
		stack[stackPointer] = lightInfo{x, y - 1, z, light}
		stackPointer++
	} else {
		///stackP = append(stackP, lightInfo{x, y - 1, z, light - 1})
		stack[stackPointer] = lightInfo{x, y - 1, z, light - 1}
		stackPointer++
	}
	//stackP = append(stackP, lightInfo{x, y + 1, z, light - 1})
	stack[stackPointer] = lightInfo{x, y + 1, z, light - 1}
	stackPointer++

	//stackP = append(stackP, lightInfo{x - 1, y, z, light - 1})
	stack[stackPointer] = lightInfo{x - 1, y, z, light - 1}
	stackPointer++

	//stackP = append(stackP, lightInfo{x + 1, y, z, light - 1})
	stack[stackPointer] = lightInfo{x + 1, y, z, light - 1}
	stackPointer++

	//stackP = append(stackP, lightInfo{x, y, z - 1, light - 1})
	stack[stackPointer] = lightInfo{x, y, z - 1, light - 1}
	stackPointer++

	//stackP = append(stackP, lightInfo{x, y, z + 1, light - 1})
	stack[stackPointer] = lightInfo{x, y, z + 1, light - 1}
	stackPointer++

	return stackPointer
}
