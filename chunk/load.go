package chunk

import (
	"compress/gzip"
	"compress/zlib"
	"errors"
	"github.com/NetherrackDev/netherrack/nbt"
	"io"
	"log"
	"os"
)

func (chunk *Chunk) tryLoad() byte {

	region := chunk.World.getRegion(chunk.X>>5, chunk.Z>>5)
	region.addChunk()

	region.RLock()
	defer region.RUnlock()
	if !region.chunkExists(chunk.X, chunk.Z) {
		return 0
	}

	chunkNBT, err := chunk.loadNBT()
	if err != nil {
		log.Println(err.Error())
		return 2
	}

	level, ok := chunkNBT.GetCompound("Level", false)
	if !ok {
		log.Println("Currupted chunk")
		return 2
	}

	if biomes, ok := level.GetByteArray("Biomes", nil); ok {
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				chunk.SetBiome(x, z, byte(biomes[x|z<<4]))
			}
		}
	}

	heightMap, _ := level.GetIntArray("HeightMap", nil)
	chunk.heightMap = heightMap

	sections, ok := level.GetList("Sections", false)
	if !ok {
		log.Println("Currupted chunk")
		return 2
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
	return 1
}

func (chunk *Chunk) loadNBT() (chunkNBT nbt.Type, err error) {
	region := chunk.World.getRegion(chunk.X>>5, chunk.Z>>5)
	offset := region.getOffset(chunk.X, chunk.Z)

	regionFile := region.file
	headerBytes := make([]byte, 5)
	regionFile.ReadAt(headerBytes, int64(offset)*SECTOR_SIZE)
	//size := binary.BigEndian.Uint32(headerBytes)
	compressionType := headerBytes[4]
	fv := &fileView{regionFile, int64(offset)*SECTOR_SIZE + 5}
	var r io.Reader
	if compressionType == 2 {
		var zl io.ReadCloser
		zl, err = zlib.NewReader(fv)
		if err != nil {
			return
		}
		defer zl.Close()
		r = zl
	} else if compressionType == 1 {
		var gz io.ReadCloser
		gz, err = gzip.NewReader(fv)
		if err != nil {
			return
		}
		defer gz.Close()
		r = gz
	} else {
		err = errors.New("Unsupported compression type")
		return
	}
	chunkNBT, err = nbt.Parse(r)
	if err != nil {
		return
	}
	return
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
