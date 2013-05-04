package chunk

import (
	"compress/zlib"
	"fmt"
	"github.com/thinkofdeath/netherrack/nbt"
	"github.com/thinkofdeath/soulsand/blocks"
	"os"
	"path/filepath"
)

func (chunk *Chunk) generate() {
	if !chunk.tryLoad() {
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
}

func (chunk *Chunk) tryLoad() bool {
	chunk.World.dataLock.RLock()
	defer chunk.World.dataLock.RUnlock()
	rx := chunk.X >> 5
	rz := chunk.Z >> 5
	region, err := os.Open(filepath.Join("worlds", chunk.World.Name, "region", fmt.Sprintf("r.%d.%d.mca", rx, rz)))
	defer region.Close()
	if err != nil {
		return false
	}
	region.Seek(int64(((chunk.X-(rx<<5))+((chunk.Z-(rz<<5))<<5))*4), 0)

	data := make([]byte, 4)
	region.Read(data)
	offset := int32(data[2]) | int32(data[1])<<8 | int32(data[0])<<16
	count := data[3]
	if offset == 0 || count == 0 {
		return false
	}

	region.Seek(int64(offset)*4096, 0)
	headerBytes := make([]byte, 5)
	region.Read(headerBytes)
	//size:= binary.BigEndian.Uint32(headerBytes)
	compressionType := headerBytes[4]

	if compressionType != 2 {
		panic("Unsupported compression type")
	}

	zl, err := zlib.NewReader(region)
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
