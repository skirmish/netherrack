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

type Generator interface {
	//Generates the terrain in the passed chunk
	Generate(chunk *Chunk)
	//Returns the generator's name
	Name() string
	//Loads the generator's settings from the world
	Load(world *World)
	//Saves the generator's settings to the world
	Save(world *World)
}

var generators = map[string]func() Generator{}

//Add a generator to be used for gemerating worlds.
//Should only be called at init.
func AddGenerator(name string, f func() Generator) {
	generators[name] = f
}
