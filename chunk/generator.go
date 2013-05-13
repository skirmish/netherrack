package chunk

import (
	"compress/zlib"
	"github.com/NetherrackDev/netherrack/nbt"
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/blocks"
	"os"
)

func (chunk *Chunk) generate() {
	if !chunk.tryLoad() {
		chunk.World.generator.Generate(int(chunk.X), int(chunk.Z), chunk)
		chunk.Relight()
	}
}

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
					chunk.SetBlock(x, y, z, blocks.Dandelion.Id())
				} else {
					chunk.SetSkyLight(x, y, z, 15)
					chunk.SetBlock(x, y, z, blocks.Rose.Id())
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
			//chunk.SetBlock(x, y, z, blocks.BrownMushroom.Id())

			chunk.SetSkyLight(x, y, z, light)

			if y > 0 {
				block := blocks.GetBlockById(chunk.GetBlock(x, y-1, z))
				newLight := int8(light) - int8(block.LightFiltered())
				if newLight == 15 && block.StopsSkylight() {
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

type defaultGenerator int

func (defaultGenerator) Generate(x, z int, chunk soulsand.SyncChunk) {
	for y := 0; y < 256; y++ {
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				if y <= 64 {
					chunk.SetBlock(x, y, z, blocks.Wool.Id())
					if x == 0 || x == 15 || z == 0 || z == 15 {
						chunk.SetMeta(x, y, z, 1)
					} else {
						chunk.SetMeta(x, y, z, byte((y>>4)+4))
					}
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

func (chunk *Chunk) tryLoad() bool {
	region := chunk.World.getRegion(chunk.X>>5, chunk.Z>>5)
	region.addChunk()
	if !region.chunkExists(chunk.X, chunk.Z) {
		return false
	}

	region.RLock()
	defer region.RUnlock()
	offset := region.getOffset(chunk.X, chunk.Z)
	regionFile := region.file

	headerBytes := make([]byte, 5)
	regionFile.ReadAt(headerBytes, int64(offset)*SECTOR_SIZE)
	//size:= binary.BigEndian.Uint32(headerBytes)
	compressionType := headerBytes[4]

	if compressionType != 2 {
		panic("Unsupported compression type")
	}
	fv := &fileView{regionFile, int64(offset)*SECTOR_SIZE + 5}
	zl, err := zlib.NewReader(fv)
	if err != nil {
		panic("Compression error")
	}
	chunkNBT := nbt.Parse(zl)

	level, ok := chunkNBT.GetCompound("Level", false)
	if !ok {
		panic("Currupted chunk")
	}

	if biomes, ok := level.GetByteArray("Biomes", nil); ok {
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				chunk.SetBiome(x, z, byte(biomes[x+z*16]))
			}
		}
	}

	heightMap, _ := level.GetIntArray("HeightMap", nil)
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			chunk.heightMap[x|z<<4] = heightMap[x|z<<4]
		}
	}

	sections, ok := level.GetList("Sections", false)
	if !ok {
		panic("Currupted chunk")
	}
	for _, s := range sections {
		section := s.(nbt.Type)
		sectionY, _ := section.GetByte("Y", 0)
		blocks, _ := section.GetByteArray("Blocks", nil)
		data, _ := section.GetByteArray("Data", nil)
		blockLight, _ := section.GetByteArray("BlockLight", nil)
		skyLight, _ := section.GetByteArray("SkyLight", nil)
		for y := 0; y < 16; y++ {
			for z := 0; z < 16; z++ {
				for x := 0; x < 16; x++ {
					i := x | z<<4 | y<<8
					chunk.SetBlock(x, int(sectionY)*16+y, z, byte(blocks[i]))
					if i&1 == 0 {
						chunk.SetMeta(x, int(sectionY)*16+y, z, byte(data[i>>1])&0xF)
						chunk.SetBlockLight(x, int(sectionY)*16+y, z, byte(blockLight[i>>1])&0xF)
						chunk.SetSkyLight(x, int(sectionY)*16+y, z, byte(skyLight[i>>1])&0xF)
						continue
					}
					chunk.SetMeta(x, int(sectionY)*16+y, z, byte(data[i>>1]>>4))
					chunk.SetBlockLight(x, int(sectionY)*16+y, z, byte(blockLight[i>>1])>>4)
					chunk.SetSkyLight(x, int(sectionY)*16+y, z, byte(skyLight[i>>1])>>4)
				}
			}
		}
	}
	chunk.needsRelight = false
	return true
}

type fileView struct {
	file   *os.File
	offset int64
}

func (fv *fileView) Read(b []byte) (int, error) {
	n, err := fv.file.ReadAt(b, fv.offset)
	fv.offset += int64(n)
	return n, err
}
