package chunk

import (
	"compress/zlib"
	"github.com/thinkofdeath/netherrack/nbt"
	"github.com/thinkofdeath/soulsand"
	"github.com/thinkofdeath/soulsand/blocks"
	"log"
	"os"
	"time"
)

func (chunk *Chunk) generate() {
	if !chunk.tryLoad() {
		chunk.World.generator.Generate(int(chunk.X), int(chunk.Z), chunk)
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

	start := time.Now().UnixNano()

	var skyLightQueue *lightInfo
	var blockLightQueue *lightInfo

	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			for y := 0; y < 256; y++ {
				if y >= int(chunk.heightMap[x|z<<4]) {
					skyLightQueue = skyLightQueue.Append(&lightInfo{
						x:     x,
						y:     y,
						z:     z,
						light: 15,
					})
				}
				block := blocks.GetBlockById(chunk.GetBlock(x, y, z))
				if light := block.LightLevel(); light != 0 {
					blockLightQueue = blockLightQueue.Append(&lightInfo{
						x:     x,
						y:     y,
						z:     z,
						light: light,
					})

				}
			}
		}
	}

	if skyLightQueue != nil {
		current := skyLightQueue.root
		for current != nil {
			info := current
			current = current.next
			info.root = nil
			info.next = nil
			x := info.x
			z := info.z
			y := info.y
			light := info.light
			if light <= chunk.GetSkyLight(x, y, z) {
				continue
			}
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
				block := blocks.GetBlockById(chunk.GetBlock(x, y+1, z))
				newLight := int8(light) - int8(block.LightFiltered()) - 1
				if int8(chunk.GetSkyLight(x, y+1, z)) < newLight {
					skyLightQueue = skyLightQueue.Append(&lightInfo{
						x:     x,
						y:     y + 1,
						z:     z,
						light: byte(newLight),
					})
				}
			}

			if x > 0 {
				block := blocks.GetBlockById(chunk.GetBlock(x-1, y, z))
				newLight := int8(light) - int8(block.LightFiltered()) - 1
				if int8(chunk.GetSkyLight(x-1, y, z)) < newLight {
					skyLightQueue = skyLightQueue.Append(&lightInfo{
						x:     x - 1,
						y:     y,
						z:     z,
						light: byte(newLight),
					})
				}
			}
			if x < 15 {
				block := blocks.GetBlockById(chunk.GetBlock(x+1, y, z))
				newLight := int8(light) - int8(block.LightFiltered()) - 1
				if int8(chunk.GetSkyLight(x+1, y, z)) < newLight {
					skyLightQueue = skyLightQueue.Append(&lightInfo{
						x:     x + 1,
						y:     y,
						z:     z,
						light: byte(newLight),
					})
				}
			}

			if z > 0 {
				block := blocks.GetBlockById(chunk.GetBlock(x, y, z-1))
				newLight := int8(light) - int8(block.LightFiltered()) - 1
				if int8(chunk.GetSkyLight(x, y, z-1)) < newLight {
					skyLightQueue = skyLightQueue.Append(&lightInfo{
						x:     x,
						y:     y,
						z:     z - 1,
						light: byte(newLight),
					})
				}
			}
			if z < 15 {
				block := blocks.GetBlockById(chunk.GetBlock(x, y, z+1))
				newLight := int8(light) - int8(block.LightFiltered()) - 1
				if int8(chunk.GetSkyLight(x, y, z+1)) < newLight {
					skyLightQueue = skyLightQueue.Append(&lightInfo{
						x:     x,
						y:     y,
						z:     z + 1,
						light: byte(newLight),
					})
				}
			}
		}
	}
	if blockLightQueue != nil {
		current := blockLightQueue.root
		for current != nil {
			info := current
			current = current.next
			info.root = nil
			info.next = nil
			x := info.x
			z := info.z
			y := info.y
			light := info.light
			if light <= chunk.GetBlockLight(x, y, z) {
				continue
			}
			chunk.SetBlockLight(x, y, z, light)

			if y > 0 {
				block := blocks.GetBlockById(chunk.GetBlock(x, y-1, z))
				newLight := int8(light) - int8(block.LightFiltered()) - 1
				if int8(chunk.GetBlockLight(x, y-1, z)) < newLight {
					blockLightQueue = blockLightQueue.Append(&lightInfo{
						x:     x,
						y:     y - 1,
						z:     z,
						light: byte(newLight),
					})
				}
			}

			if y < 255 {
				block := blocks.GetBlockById(chunk.GetBlock(x, y+1, z))
				newLight := int8(light) - int8(block.LightFiltered()) - 1
				if int8(chunk.GetBlockLight(x, y+1, z)) < newLight {
					blockLightQueue = blockLightQueue.Append(&lightInfo{
						x:     x,
						y:     y + 1,
						z:     z,
						light: byte(newLight),
					})
				}
			}

			if x > 0 {
				block := blocks.GetBlockById(chunk.GetBlock(x-1, y, z))
				newLight := int8(light) - int8(block.LightFiltered()) - 1
				if int8(chunk.GetBlockLight(x-1, y, z)) < newLight {
					blockLightQueue = blockLightQueue.Append(&lightInfo{
						x:     x - 1,
						y:     y,
						z:     z,
						light: byte(newLight),
					})
				}
			}
			if x < 15 {
				block := blocks.GetBlockById(chunk.GetBlock(x+1, y, z))
				newLight := int8(light) - int8(block.LightFiltered()) - 1
				if int8(chunk.GetBlockLight(x+1, y, z)) < newLight {
					blockLightQueue = blockLightQueue.Append(&lightInfo{
						x:     x + 1,
						y:     y,
						z:     z,
						light: byte(newLight),
					})
				}
			}

			if z > 0 {
				block := blocks.GetBlockById(chunk.GetBlock(x, y, z-1))
				newLight := int8(light) - int8(block.LightFiltered()) - 1
				if int8(chunk.GetBlockLight(x, y, z-1)) < newLight {
					blockLightQueue = blockLightQueue.Append(&lightInfo{
						x:     x,
						y:     y,
						z:     z - 1,
						light: byte(newLight),
					})
				}
			}
			if z < 15 {
				block := blocks.GetBlockById(chunk.GetBlock(x, y, z+1))
				newLight := int8(light) - int8(block.LightFiltered()) - 1
				if int8(chunk.GetBlockLight(x, y, z+1)) < newLight {
					blockLightQueue = blockLightQueue.Append(&lightInfo{
						x:     x,
						y:     y,
						z:     z + 1,
						light: byte(newLight),
					})
				}
			}
		}
	}

	taken := time.Now().UnixNano() - start
	log.Printf("Time: %dms\n", taken/1000000)
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
			chunk.heightMap[x|z<<4] = heightMap[z|x<<4]
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
