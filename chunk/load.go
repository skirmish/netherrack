package chunk

import (
	"compress/gzip"
	"compress/zlib"
	"errors"
	"github.com/NetherrackDev/netherrack/log"
	"github.com/NetherrackDev/netherrack/nbt"
	"io"
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

	if biomes, ok := level.GetUByteArray("Biomes", nil); ok {
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				chunk.SetBiome(x, z, biomes[x|z<<4])
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
		blockBytes, _ := section.GetUByteArray("Blocks", nil)
		data, _ := section.GetUByteArray("Data", nil)
		blockLight, _ := section.GetUByteArray("BlockLight", nil)
		skyLight, _ := section.GetUByteArray("SkyLight", nil)
		chunkSection := chunk.SubChunks[sectionY]
		if chunkSection == nil {
			chunkSection = &SubChunk{}
			chunk.SubChunks[sectionY] = chunkSection
		}
		chunkSection.Type = blockBytes
		chunkSection.MetaData = data
		chunkSection.BlockLight = blockLight
		chunkSection.SkyLight = skyLight

		//Update infomation on the chunk
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				for y := 0; y < 16; y++ {
					i := (y << 8) | (z << 4) | x
					if chunkSection.Type[i] != 0 {
						chunkSection.blocks++
					}
					idx := i >> 1
					if i&1 == 0 {
						if chunkSection.BlockLight[idx]&0xF > 0 {
							chunkSection.blockLights++
						}
						if chunkSection.SkyLight[idx]&0xF > 0 {
							chunkSection.skyLights++
						}
					} else {
						if chunkSection.BlockLight[idx]>>4 > 0 {
							chunkSection.blockLights++
						}
						if chunkSection.SkyLight[idx]>>4 > 0 {
							chunkSection.skyLights++
						}
					}
				}
			}
		}
	}
	chunk.needsSave = false
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
	chunkNBT, err = nbt.ParseBytes(r, true)
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
