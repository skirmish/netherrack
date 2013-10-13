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

type EntityComponent struct {
	ID     int32
	Uuid   string
	Server Server
	World  *world.World

	CurrentTick uint64

	Systems []System
}

func (e *EntityComponent) Init(root interface{}) {
	e.Systems = Systems(root)
}

func (e *EntityComponent) Update(root interface{}) {
	for _, s := range e.Systems {
		s.Update(root)
	}
}

//Returns the entity's UUID
func (e *EntityComponent) UUID() string {
	return e.Uuid
}

func (e *EntityComponent) Entity() *EntityComponent {
	return e
}

func init() {
	RegisterSystem(TickSystem{})
}

type TickSystem struct{}

type tickable interface {
	Entity() *EntityComponent
}

func (TickSystem) Valid(e interface{}) bool {
	_, ok := e.(tickable)
	return ok
}

func (TickSystem) Priority() Priority { return Highest }

func (TickSystem) Update(entity interface{}) {
	ti := entity.(tickable)
	e := ti.Entity()
	e.CurrentTick++
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
