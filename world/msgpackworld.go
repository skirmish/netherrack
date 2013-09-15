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

const msgpackName = "Msgpack"

func init() {
	AddSystem(msgpackName, func() System {
		return &MsgpackWorld{}
	})
}

type MsgpackWorld struct {
	Name string
	path string `msgpack:"ignore"`
}

//Loads or creates the system
func (mw *MsgpackWorld) Init(path string) {
	mw.path = path
	level := filepath.Join(path, "level.nether")
	_, err := os.Stat(level)
	if err == nil {
		mw.readLevel(level)
	}
	mw.writeLevel(level)
}

func (mw *MsgpackWorld) readLevel(level string) {
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

func (mw *MsgpackWorld) writeLevel(level string) {
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
func (mw *MsgpackWorld) SystemName() string {
	return msgpackName
}

//Writes the passed struct/struct pointer to the data folder
//with the name key.nether.
func (mw *MsgpackWorld) Write(key string, v interface{}) error {
	f, err := os.Create(filepath.Join(mw.path, "data", key+".nether"))
	if err != nil {
		return err
	}
	defer f.Close()
	return msgpack.Write(f, v)
}

//Reads key into the passed struct pointer
func (mw *MsgpackWorld) Read(key string, v interface{}) error {
	f, err := os.Open(filepath.Join(mw.path, "data", key+".nether"))
	if err != nil {
		return err
	}
	defer f.Close()
	return msgpack.Read(f, v)
}
