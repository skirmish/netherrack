/*
   Copyright 2013 Matthew Collins (purggames@gmail.com)

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package world

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/NetherrackDev/netherrack/format/msgpack"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const msgpackName = "Msgpack"

func init() {
	AddSystem(msgpackName, func() System {
		return &MsgpackSystem{}
	})
}

type MsgpackSystem struct {
	Name string
	path string `msgpack:"ignore"`

	regionLock sync.RWMutex       `msgpack:"ignore"`
	regions    map[uint64]*region `msgpack:"ignore"`
	needsSave  bool               `msgpack:"ignore"`
}

//Loads or creates the system
func (mw *MsgpackSystem) Init(path string) {
	mw.path = path
	mw.regions = make(map[uint64]*region)
	level := filepath.Join(path, "level.nether")
	_, err := os.Stat(level)
	if err == nil {
		mw.readLevel(level)
	}
	mw.writeLevel(level)
	os.MkdirAll(filepath.Join(mw.path, "data"), 0777)
	os.MkdirAll(filepath.Join(mw.path, "regions"), 0777)
	go mw.run()
}

func (mw *MsgpackSystem) run() {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		<-t.C
		if mw.needsSave {
			mw.needsSave = false
			for _, reg := range mw.regions {
				if reg.needsSave { //A bit racey but not an issue
					reg.Lock()
					reg.needsSave = false
					reg.Save()
					reg.Unlock()
				}
			}
		}
	}
}

//Returns the chunk at the coordinates, also returns if the chunk existed
//before this
func (mw *MsgpackSystem) Chunk(x, z int) (*Chunk, bool) {
	reg := mw.region(x>>5, z>>5)
	reg.RLock()
	rx := x & 0x1F
	rz := z & 0x1F
	idx := rx | (rz << 5)
	offset := reg.Offsets[idx]
	count := reg.SectionCounts[idx]
	reg.RUnlock()
	if offset == 0 {
		return &Chunk{X: x, Z: z}, false
	}

	section := io.NewSectionReader(reg.file, int64(offset)*regionSectionSize, int64(count)*regionSectionSize)
	gz, err := gzip.NewReader(section)
	if err != nil {
		panic(err)
	}
	chunk := &Chunk{}
	err = msgpack.Read(gz, chunk)
	gz.Close()

	if err != nil {
		panic(err)
	}
	return chunk, true
}

//Saves the chunk back to storage.
func (mw *MsgpackSystem) SaveChunk(x, z int, chunk *Chunk) {
	reg := mw.region(x>>5, z>>5)
	reg.Lock()
	rx := x & 0x1F
	rz := z & 0x1F
	idx := rx | (rz << 5)
	if off := reg.Offsets[idx]; off != 0 {
		for i := off; i < off+reg.SectionCounts[idx]; i++ {
			reg.usedLocations[i] = false
		}
	}
	reg.Unlock()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	msgpack.Write(gz, chunk)
	gz.Close()

	reg.Lock()
	offset := len(reg.usedLocations)
	count := (buf.Len() / regionSectionSize) + 1
check:
	for i := 3; i < len(reg.usedLocations); i++ {
		if !reg.usedLocations[i] {
			for j := 1; j < count; j++ {
				if reg.usedLocations[i+j] {
					continue check
				}
			}
			offset = i
			break
		}
	}
	if len(reg.usedLocations) < offset+count {
		temp := reg.usedLocations
		reg.usedLocations = make([]bool, offset+count)
		copy(reg.usedLocations, temp)
	}
	for i := offset; i < offset+count; i++ {
		reg.usedLocations[i] = true
	}
	reg.Offsets[idx] = offset
	reg.SectionCounts[idx] = count
	reg.needsSave = true
	mw.needsSave = true
	reg.chunkcount++
	reg.Unlock()

	n, err := reg.file.WriteAt(buf.Bytes(), int64(offset)*regionSectionSize)
	if err != nil || n != buf.Len() {
		panic(err)
	}
}

const maxRegionHeaderSize = regionSectionSize * 3
const regionSectionSize = 1024 * 4

type region struct {
	Offsets       [32 * 32]int
	SectionCounts [32 * 32]int

	file          *os.File `msgpack:"ignore"`
	usedLocations []bool   `msgpack:"ignore"`
	sync.RWMutex  `msgpack:"ignore"`
	needsSave     bool `msgpack:"ignore"`
	chunkcount    uint `msgpack:"ignore"`
}

func (r *region) Save() {
	r.file.Seek(0, 0)
	msgpack.Write(r.file, r)
}

func (mw *MsgpackSystem) region(x, z int) *region {
	key := (uint64(int32(x)) & 0xFFFFFFFF) | ((uint64(int32(z)) & 0xFFFFFFFF) << 32)
	mw.regionLock.RLock()
	reg, ok := mw.regions[key]
	mw.regionLock.RUnlock()
	if ok {
		return reg
	}
	mw.regionLock.Lock()
	reg, ok = mw.regions[key]
	if ok {
		mw.regionLock.Unlock()
		return reg
	}
	reg = &region{}
	file, err := os.OpenFile(filepath.Join(mw.path, "regions", fmt.Sprintf("%016X.region", key)), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	msgpack.Read(file, reg)
	reg.file = file
	reg.usedLocations = []bool{true, true, true}
	for i, off := range reg.Offsets {
		if off == 0 {
			continue
		}
		count := reg.SectionCounts[i]
		if len(reg.usedLocations) < off+count {
			temp := reg.usedLocations
			reg.usedLocations = make([]bool, off+count)
			copy(reg.usedLocations, temp)
		}
		for j := off; j < off+count; j++ {
			reg.usedLocations[j] = true
		}
	}
	reg.Save()

	mw.regions[key] = reg
	mw.regionLock.Unlock()
	return reg
}

//Closes the chunk in the system
func (mw *MsgpackSystem) CloseChunk(x, z int, chunk *Chunk) {
	reg := mw.region(x>>5, z>>5)
	mw.regionLock.Lock()
	reg.Lock()
	reg.chunkcount--
	if reg.chunkcount == 0 {
		//Unload
		delete(mw.regions, (uint64(int32(x>>5))&0xFFFFFFFF)|((uint64(int32(z>>5))&0xFFFFFFFF)<<32))
	}
	reg.Unlock()
	mw.regionLock.Unlock()
}

func (mw *MsgpackSystem) readLevel(level string) {
	f, err := os.Open(level)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	err = msgpack.Read(f, mw)
	if err != nil {
		panic(err)
	}
}

func (mw *MsgpackSystem) writeLevel(level string) {
	f, err := os.Create(level)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	err = msgpack.Write(f, mw)
	if err != nil {
		panic(err)
	}
}

//Gets the name of the system
func (mw *MsgpackSystem) SystemName() string {
	return msgpackName
}

//Writes the passed struct/struct pointer to the data folder
//with the name key.nether.
func (mw *MsgpackSystem) Write(key string, v interface{}) error {
	f, err := os.Create(filepath.Join(mw.path, "data", key+".nether"))
	if err != nil {
		return err
	}
	defer f.Close()
	return msgpack.Write(f, v)
}

//Reads key into the passed struct pointer
func (mw *MsgpackSystem) Read(key string, v interface{}) error {
	f, err := os.Open(filepath.Join(mw.path, "data", key+".nether"))
	if err != nil {
		return err
	}
	defer f.Close()
	return msgpack.Read(f, v)
}
