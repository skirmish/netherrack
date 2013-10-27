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
	"compress/zlib"
	"github.com/NetherrackDev/netherrack/protocol"
	"time"
)

//Dimensions normally control the lighting and skycolour
type Dimension int8

const (
	Overworld Dimension = 0
	Nether    Dimension = -1
	End       Dimension = 1
)

type World struct {
	Name string

	system    System
	generator Generator

	tryClose chan TryClose

	loadedChunks map[uint64]*Chunk

	joinChunk       chan joinChunk
	leaveChunk      chan joinChunk
	entityChunk     chan entityChunk
	placeBlock      chan blockChange
	getBlock        chan blockGet
	light           chan lightEvent
	chunkPacket     chan chunkPacket
	updateSpawnData chan packetUpdate
	timeOfDay       chan chan int64

	//The limiters were added because trying to send/save all the chunks
	//at once caused large amounts of memory usage
	SendLimiter  chan cachedCompressor
	SaveLimiter  chan struct{}
	RequestClose chan *Chunk

	worldData struct {
		Dimension  Dimension
		AgeOfWorld int64
		TimeOfDay  int64
	}
}

type TryClose struct {
	Ret   chan struct{}
	Done  chan bool
	World *World
}

type cachedCompressor struct {
	buf *bytes.Buffer
	zl  *zlib.Writer
}

func (world *World) init() {
	world.joinChunk = make(chan joinChunk, 500)
	world.leaveChunk = make(chan joinChunk, 500)

	world.placeBlock = make(chan blockChange, 1000)
	world.getBlock = make(chan blockGet, 1000)
	world.light = make(chan lightEvent, 1000)

	world.chunkPacket = make(chan chunkPacket, 1000)
	world.entityChunk = make(chan entityChunk, 500)
	world.RequestClose = make(chan *Chunk, 20)
	world.updateSpawnData = make(chan packetUpdate, 200)
	world.timeOfDay = make(chan chan int64, 100)
}

func (world *World) run() {
	var done chan bool
	defer func() { done <- true }()
	defer world.system.Close()
	world.generator.Load(world)
	world.loadedChunks = make(map[uint64]*Chunk)
	world.SendLimiter = make(chan cachedCompressor, 20)
	for i := 0; i < cap(world.SendLimiter); i++ {
		buf := &bytes.Buffer{}
		world.SendLimiter <- cachedCompressor{buf, zlib.NewWriter(buf)}
	}
	world.SaveLimiter = make(chan struct{}, 5)
	for i := 0; i < cap(world.SaveLimiter); i++ {
		world.SaveLimiter <- struct{}{}
	}
	defer world.system.Write("levelData", &world.worldData)
	tick := time.NewTicker(time.Second / 10)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			world.worldData.AgeOfWorld += 2
			world.worldData.TimeOfDay += 2
			if world.worldData.AgeOfWorld%(20*60*5) == 0 {
				world.system.Write("levelData", &world.worldData)
			}
			if len(world.loadedChunks) == 0 {
				ret := make(chan struct{})
				done = make(chan bool)
				world.tryClose <- TryClose{
					Ret:   ret,
					Done:  done,
					World: world,
				}
				<-ret
				if len(world.joinChunk) > 0 {
					done <- false
					continue
				}
				return
			}
		case ret := <-world.timeOfDay:
			ret <- world.worldData.TimeOfDay
		case jc := <-world.joinChunk:
			world.chunk(jc.x, jc.z).Join(jc.watcher)
		case lc := <-world.leaveChunk:
			world.chunk(lc.x, lc.z).Leave(lc.watcher)

		case bp := <-world.placeBlock:
			cx, cz := bp.X>>4, bp.Z>>4
			world.chunk(cx, cz).blockPlace <- bp
		case bg := <-world.getBlock:
			cx, cz := bg.X>>4, bg.Z>>4
			world.chunk(cx, cz).blockGet <- bg
		case l := <-world.light:
			cx, cz := l.X>>4, l.Z>>4
			world.chunk(cx, cz).light <- l

		case cp := <-world.chunkPacket:
			world.chunk(cp.X, cp.Z).chunkPacket <- cp
		case ec := <-world.entityChunk:
			world.chunk(ec.X, ec.Z).entity <- ec
		case chunk := <-world.RequestClose:
			if chunk.Close() {
				delete(world.loadedChunks, chunkKey(chunk.X, chunk.Z))
			}
		case pu := <-world.updateSpawnData:
			world.chunk(pu.X, pu.Z).entitySpawnUpdate <- pu
		}
	}
}

func (world *World) TimeOfDay() int64 {
	ret := make(chan int64, 1)
	world.timeOfDay <- ret
	return <-ret
}

type packetUpdate struct {
	entity  Entity
	X, Z    int
	spawn   bool
	packets []protocol.Packet
}

func (world *World) UpdateSpawnData(x, z int, entity Entity, spawn bool, packets []protocol.Packet) {
	world.updateSpawnData <- packetUpdate{
		X: z, Z: z,
		entity:  entity,
		spawn:   spawn,
		packets: packets,
	}
}

type entityChunk struct {
	Add     bool
	X, Z    int
	Entity  Entity
	Spawn   []protocol.Packet
	Despawn []protocol.Packet
}

func (world *World) AddEntity(x, z int, entity Entity, spawn []protocol.Packet, despawn []protocol.Packet) {
	world.entityChunk <- entityChunk{
		Add:     true,
		X:       x,
		Z:       z,
		Entity:  entity,
		Spawn:   spawn,
		Despawn: despawn,
	}
}

func (world *World) RemoveEntity(x, z int, entity Entity) {
	world.entityChunk <- entityChunk{
		Add:    false,
		X:      x,
		Z:      z,
		Entity: entity,
	}

}

type blockChange struct {
	X, Y, Z     int
	Block, Data byte
}

type chunkPacket struct {
	Packet protocol.Packet
	UUID   string
	X, Z   int
}

//Sends the packet to all watchers of the chunk apart from the watcher with
//the passed uuid (leave blank to send to all)
func (world *World) QueuePacket(x, z int, uuid string, packet protocol.Packet) {
	world.chunkPacket <- chunkPacket{
		packet,
		uuid,
		x, z,
	}
}

type lightEvent struct {
	X, Y, Z int
	Sky     bool
	Value   interface{}
}

type lightChange struct {
	Level byte
}

func (world *World) setLight(x, y, z int, level byte, sky bool) {
	if y < 0 || y >= 255 {
		return
	}
	world.light <- lightEvent{x, y, z, sky, lightChange{level}}
}

type lightGet struct {
	Ret chan byte
}

//Gets the block and data at the location
func (world *World) Light(x, y, z int, sky bool) (level byte) {
	if y < 0 || y >= 255 {
		if sky {
			return 15
		}
		return 0
	}
	ret := make(chan byte, 1)
	world.light <- lightEvent{
		x, y, z, sky, lightGet{ret},
	}
	d := <-ret
	return d
}

//Sets the block and data at the location
func (world *World) SetBlock(x, y, z int, block, data byte) {
	world.placeBlock <- blockChange{
		x, y, z, block, data,
	}
}

type blockGet struct {
	X, Y, Z int
	Ret     chan [2]byte
}

//Gets the block and data at the location
func (world *World) Block(x, y, z int) (block, data byte) {
	ret := make(chan [2]byte, 1)
	world.getBlock <- blockGet{
		x, y, z, ret,
	}
	d := <-ret
	return d[0], d[1]
}

type joinChunk struct {
	x, z    int
	watcher Watcher
}

//Adds the watcher to the chunk at the coordinates. If the chunk isn't loaded
//then it will be loaded.
func (world *World) JoinChunk(x, z int, watcher Watcher) {
	world.joinChunk <- joinChunk{x, z, watcher}
}

//Removes the watcher to the chunk at the coordinates.
func (world *World) LeaveChunk(x, z int, watcher Watcher) {
	world.leaveChunk <- joinChunk{x, z, watcher}
}

//Writes the value into the world's system's storage. This method
//is safe to call from different goroutines when the key is different.
func (world *World) Write(key string, value interface{}) error {
	return world.system.Write(key, value)
}

//Reads the value into the world's system's storage. This method
//is safe to call from different goroutines when the key is different.
func (world *World) Read(key string, value interface{}) error {
	return world.system.Read(key, value)
}

//Gets the loaded chunk or loads it if it isn't loaded
func (world *World) chunk(x, z int) *Chunk {
	if chunk, ok := world.loadedChunks[chunkKey(x, z)]; ok {
		return chunk
	}
	chunk, ok := world.system.Chunk(x, z)
	world.loadedChunks[chunkKey(x, z)] = chunk
	if !ok {
		chunk.Init(world, world.generator, world.system)
	} else {
		chunk.Init(world, nil, world.system)
	}
	return chunk
}

//Returns the worlds dimension
func (world *World) Dimension() Dimension {
	return world.worldData.Dimension
}

func chunkKey(x, z int) uint64 {
	return (uint64(int32(x)) & 0xFFFFFFFF) | ((uint64(int32(z)) & 0xFFFFFFFF) << 32)
}
