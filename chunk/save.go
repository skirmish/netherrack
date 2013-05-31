package chunk

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"github.com/NetherrackDev/netherrack/nbt"
)

func (chunk *Chunk) Save() {
	chunk.needsSave = false
	return

	region := chunk.World.getRegion(chunk.X>>5, chunk.Z>>5)
	relX, relZ := chunk.X-(region.x<<5), chunk.Z-(region.z<<5)

	//The reason that the old chunk is loaded first is so that any extra/unsupported NBT
	//tags are not lost during saving
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
			biomeData[x|z<<4] = int8(chunk.Biome(x, z))
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
			nbtSection.Set("Blocks", section.Type)
			nbtSection.Set("Data", section.MetaData)
			nbtSection.Set("BlockLight", section.BlockLight)
			nbtSection.Set("SkyLight", section.SkyLight)
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

	//Lock the region to allow free sectors to be found
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
		//Increase the size of the usedSectors array
		temp := region.usedSectors
		region.usedSectors = make([]bool, len(region.usedSectors)+64)
		copy(region.usedSectors, temp)
	}
	for j := region.offsets[i]; j < region.offsets[i]+int32(region.counts[i]); j++ {
		region.usedSectors[j] = true
	}

	regionFile := region.file
	region.Unlock()

	chunkLocation := make([]byte, 4)
	off := region.offsets[i]
	chunkLocation[2], chunkLocation[1], chunkLocation[0] = byte(off), byte(off>>8), byte(off>>16)
	chunkLocation[3] = region.counts[i]
	regionFile.WriteAt(chunkLocation, int64(relX|relZ*32)*4)

	headerBytes := make([]byte, 5)
	binary.BigEndian.PutUint32(headerBytes, uint32(size+1))
	headerBytes[4] = 2
	_, err = regionFile.WriteAt(headerBytes, int64(offset)*SECTOR_SIZE)
	if err != nil {
		panic(err)
	}
	_, err = regionFile.WriteAt(data.Bytes(), int64(offset)*SECTOR_SIZE+5)
	if err != nil {
		panic(err)
	}
	//Pad the file to allow normal minecraft to load it
	_, err = regionFile.WriteAt([]byte{0}, int64(offset+count)*SECTOR_SIZE-1)
	if err != nil {
		panic(err)
	}
	//Sync the changes to the file
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
