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

package entity

import (
	"github.com/NetherrackDev/netherrack/world"
	"sync"
)

type Entity interface {
	//Returns the entity's UUID
	UUID() string
	//Spawns the entity for the watcher
	SpawnFor(world.Watcher)
	//Spawns the entity for the watcher
	DespawnFor(world.Watcher)
	//Returns wether the entity is saveable to the chunk
	Saveable() bool
}

//Contains methods that a entity needs for a server
type Server interface {
}

var (
	usedEntityIDs = map[int32]bool{}
	lastID        int32
	ueiLock       sync.Mutex
)

func GetID() int32 {
	ueiLock.Lock()
	defer ueiLock.Unlock()
	for {
		if !usedEntityIDs[lastID] {
			usedEntityIDs[lastID] = true
			id := lastID
			lastID++
			return id
		}
		lastID++
	}
}

func FreeID(id int32) {
	ueiLock.Lock()
	defer ueiLock.Unlock()
	delete(usedEntityIDs, id)
}
