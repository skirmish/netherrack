package chunk

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"github.com/NetherrackDev/netherrack/nbt"
)

func (chunk *Chunk) Save() {

	region := chunk.World.getRegion(chunk.X>>5, chunk.Z>>5)
	relX, relZ := chunk.X-(region.x<<5), chunk.Z-(region.z<<5)

	chunkNBT := chunk.getNBT()
	level, _ := chunkNBT.GetCompound("Level", true)

	level.Set("xPos", chunk.X)
	level.Set("zPos", chunk.Z)

	lastUpdate, _ := level.GetLong("LastUpdate", 0)
	level.Set("LastUpdate", lastUpdate)
	terrainPopulated, _ := level.GetByte("TerrainPopulated", 1)
	level.Set("TerrainPopulated", terrainPopulated)
	inhabitedTime, _ := level.GetLong("InhabitedTime", 0)
	level.Set("InhabitedTime", inhabitedTime)

	biomeData := make([]int8, 16*16)
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			biomeData[x|z<<4] = int8(chunk.GetBiome(x, z))
		}
	}
	level.Set("Biomes", biomeData)
	level.Set("HeightMap", chunk.heightMap)

	entities, _ := level.GetList("Entities", true)
	level.Set("Entities", entities)
	tileEntities, _ := level.GetList("TileEntities", true)
	level.Set("TileEntities", tileEntities)

	tileTicks, ok := level.GetList("TileTicks", false)
	if ok {
		level.Set("TileTicks", tileTicks)
	}

	sections := make([]interface{}, 0, 5)
	for i, section := range chunk.SubChunks {
		if section != nil {
			nbtSection := nbt.NewNBT()
			nbtSection.Set("Y", int8(i))
			sections = append(sections, nbtSection)
			blocks := make([]int8, 16*16*16)
			data := make([]int8, 16*16*16*0.5)
			blockLight := make([]int8, 16*16*16*0.5)
			skyLight := make([]int8, 16*16*16*0.5)
			for i := 0; i < 16*16*16; i++ {
				blocks[i] = int8(section.Type[i])
			}
			for i := 0; i < 16*16*16*0.5; i++ {
				data[i] = int8(section.MetaData[i])
				blockLight[i] = int8(section.BlockLight[i])
				skyLight[i] = int8(section.SkyLight[i])
			}
			nbtSection.Set("Blocks", blocks)
			nbtSection.Set("Data", data)
			nbtSection.Set("BlockLight", blockLight)
			nbtSection.Set("SkyLight", skyLight)
		}
	}
	level.Set("Sections", sections)

	var data bytes.Buffer
	zl, err := zlib.NewWriterLevel(&data, zlib.BestSpeed)
	if err != nil {
		panic(err)
	}
	chunkNBT.WriteTo(zl, fmt.Sprintf("Chunk [%d,%d]", relX, relZ))
	zl.Close()
	/*gz, err := gzip.NewWriterLevel(&data, gzip.BestSpeed)
	if err != nil {
		panic(err)
	}
	chunkNBT.WriteTo(gz, fmt.Sprintf("Chunk [%d,%d]", relX, relZ))
	gz.Close()*/
	region.Lock()
	size := data.Len()
	count := ((size + 5) / SECTOR_SIZE) + 1
	offset := len(region.usedSectors)
check:
	for i := 0; i < len(region.usedSectors); i++ {
		if !region.usedSectors[i] {
			for j := 1; j < count; j++ {
				if j >= len(region.usedSectors) {
					break
				}
				if region.usedSectors[j] {
					continue check
				}
			}
			offset = i
			break
		}
	}

	i := relX | relZ<<5
	region.counts[i] = byte(count)
	region.offsets[i] = int32(offset)

	if region.offsets[i]+int32(region.counts[i]) >= int32(len(region.usedSectors)) {
		temp := region.usedSectors
		region.usedSectors = make([]bool, len(region.usedSectors)+64)
		copy(region.usedSectors, temp)
	}
	for j := region.offsets[i]; j < region.offsets[i]+int32(region.counts[i]); j++ {
		region.usedSectors[j] = true
	}

	regionFile := region.file

	/*stat, err := regionFile.Stat()
	if err != nil {
		region.Unlock()
		panic(err)
	}
	if int64(offset+count)*SECTOR_SIZE > stat.Size() {
		regionFile.Truncate(int64(offset+count) * SECTOR_SIZE)
	}*/
	region.Unlock()

	chunkLocation := make([]byte, 4)
	off := region.offsets[i]
	chunkLocation[2], chunkLocation[1], chunkLocation[0] = byte(off), byte(off>>8), byte(off>>16)
	chunkLocation[3] = region.counts[i]
	regionFile.WriteAt(chunkLocation, int64(relX|relZ*32)*4)

	headerBytes := make([]byte, 5)
	binary.BigEndian.PutUint32(headerBytes, uint32(size+1))
	headerBytes[4] = 2
	regionFile.WriteAt(headerBytes, int64(offset)*SECTOR_SIZE)
	data.Write(make([]byte, count*SECTOR_SIZE-data.Len()-5))
	_, err = regionFile.WriteAt(data.Bytes(), int64(offset)*SECTOR_SIZE+5)
	if err != nil {
		panic(err)
	}
	err = regionFile.Sync()
	if err != nil {
		panic(err)
	}
	chunk.needsSave = false
}

func (chunk *Chunk) getNBT() (chunkNBT nbt.Type) {
	region := chunk.World.getRegion(chunk.X>>5, chunk.Z>>5)
	region.RLock()
	defer region.RUnlock()
	relX, relZ := chunk.X-(region.x<<5), chunk.Z-(region.z<<5)

	if region.chunkExists(chunk.X, chunk.Z) {
		var err error
		chunkNBT, err = chunk.loadNBT()
		if err == nil {
			offset := region.getOffset(chunk.X, chunk.Z)
			off := int(offset)
			count := int(region.counts[relX|relZ<<5])
			region.RUnlock()
			region.freeSectors(off, count)
			region.RLock()
		} else {
			chunkNBT = nbt.NewNBT()
		}

	} else {
		chunkNBT = nbt.NewNBT()
	}
	return
}

func (region *region) freeSectors(off, count int) {
	region.Lock()
	defer region.Unlock()
	for i := off; i < off+count; i++ {
		region.usedSectors[i] = false
	}
}
