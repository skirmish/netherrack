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

import ()

type World struct {
	system System
}

func (world *World) run() {

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
