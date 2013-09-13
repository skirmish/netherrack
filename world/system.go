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
	"github.com/NetherrackDev/netherrack/format/msgpack"
	"os"
	"path/filepath"
)

type System interface {
	//Loads or creates the system
	Init(path string)
	//Gets the name of the system
	SystemName() string
	//Writes the passed struct/struct pointer to the system's storage
	//with the key 'name'.
	Write(name string, v interface{}) error
	//Reads 'name' into the passed struct pointer
	Read(name string, v interface{}) error
}

var systems = map[string]func() System{}

//Add a system to be used for loading/saving worlds.
//Should only be called at init.
func AddSystem(name string, f func() System) {
	systems[name] = f
}

//Loads the world by name using the passed system if
//the world doesn't exists.
func LoadWorld(name string, system System) *World {
	metapath := filepath.Join("./worlds/", name, "netherrack.meta")
	_, err := os.Stat(metapath)
	if err == nil {
		//Load the world
		return GetWorld(name)
	}

	//Create the world
	meta := netherrackMeta{
		SystemName: system.SystemName(),
	}
	os.MkdirAll(filepath.Join("./worlds/", name), 0777)
	f, err := os.Create(metapath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	msgpack.Write(f, meta)

	w := &World{
		system: system,
	}
	system.Init(filepath.Join("./worlds/", name))
	go w.run()

	return w
}

type netherrackMeta struct {
	SystemName string
}

//Loads the world by name
func GetWorld(name string) *World {
	root := filepath.Join("./worlds/", name)
	f, err := os.Open(filepath.Join(root, "netherrack.meta"))
	if err != nil {
		return nil
	}
	defer f.Close()

	meta := netherrackMeta{}
	msgpack.Read(f, &meta)

	sys, ok := systems[meta.SystemName]
	if !ok {
		panic("Unknown world system")
	}
	system := sys()

	w := &World{
		system: system,
	}
	system.Init(filepath.Join("./worlds/", name))
	go w.run()

	return w
}
