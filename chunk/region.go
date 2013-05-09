package chunk

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

type region struct {
	sync.RWMutex
	world       *World
	chunkCount  int32
	x, z        int32
	file        *os.File
	offsets     []int32
	counts      []byte
	usedSectors []bool
}

const SECTOR_SIZE = 4096

func (world *World) getRegion(x, z int32) *region {
	world.dataLock.RLock()
	r, ok := world.regions[(uint64(x)&0xFFFFFFFF)|uint64(z)<<32]
	world.dataLock.RUnlock()
	if !ok {
		world.dataLock.Lock()
		defer world.dataLock.Unlock()
		if r, ok = world.regions[(uint64(x)&0xFFFFFFFF)|uint64(z)<<32]; ok { //cover the case of multiple goroutines requesting a lock
			return r
		}
		r = &region{}
		r.world = world
		r.init(x, z)
		world.regions[(uint64(x)&0xFFFFFFFF)|uint64(z)<<32] = r
	}
	return r
}

func (region *region) addChunk() {
	atomic.AddInt32(&region.chunkCount, 1)
}

func (region *region) removeChunk() {
	if atomic.AddInt32(&region.chunkCount, -1) == 0 {
		region.world.dataLock.Lock()
		defer region.world.dataLock.Unlock()
		delete(region.world.regions, (uint64(region.x)&0xFFFFFFFF)|uint64(region.z)<<32)
	}
}

func (region *region) init(rx, rz int32) {
	region.x, region.z = rx, rz

	os.MkdirAll(filepath.Join("worlds", region.world.Name, "region"), os.ModeDir|os.ModePerm)

	path := filepath.Join("worlds", region.world.Name, "region", fmt.Sprintf("r.%d.%d.mca", rx, rz))
	regionFile, err := os.Open(path)
	if err != nil {
		regionFile, err = os.Create(path)
		if err != nil {
			panic(err)
		}
		regionFile.Truncate(SECTOR_SIZE * 2)
		regionFile.Seek(0, 0)
	}
	region.file = regionFile

	stat, err := regionFile.Stat()
	if err != nil {
		panic(err)
	}
	region.usedSectors = make([]bool, (stat.Size()+SECTOR_SIZE/2)/SECTOR_SIZE)
	region.usedSectors[0], region.usedSectors[1] = true, true

	region.offsets = make([]int32, 32*32)
	region.counts = make([]byte, 32*32)

	chunkLocation := make([]byte, 4)
	for i := 0; i < 32*32; i++ {
		regionFile.Read(chunkLocation)
		region.offsets[i] = int32(chunkLocation[2]) | int32(chunkLocation[1])<<8 | int32(chunkLocation[0])<<16
		region.counts[i] = chunkLocation[3]
		for j := region.offsets[i]; j < region.offsets[i]+int32(region.counts[i]); j++ {
			region.usedSectors[j] = true
		}
	}
	regionFile.Seek(0, 0)
}

func (region *region) getOffset(x, z int32) int32 {
	relX, relZ := x-(region.x<<5), z-(region.z<<5)
	return region.offsets[relX|relZ<<5]
}

func (region *region) chunkExists(x, z int32) bool {
	relX, relZ := x-(region.x<<5), z-(region.z<<5)
	if region.offsets[relX|relZ<<5] == 0 || region.counts[relX|relZ<<5] == 0 {
		return false
	}
	return true
}
