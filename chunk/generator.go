package chunk

import (
	"compress/zlib"
	"github.com/thinkofdeath/netherrack/nbt"
	"github.com/thinkofdeath/soulsand/blocks"
	"os"
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
	region := chunk.World.getRegion(chunk.X>>5, chunk.Z>>5)
	if !region.chunkExists(chunk.X, chunk.Z) {
		return false
	}

	region.addChunk()

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
