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

/*
This is designed to replicate vanilla's superflat worlds.
Currently only basic worlds are supported.
*/
package flat

import (
	"bytes"
	"github.com/NetherrackDev/netherrack/world"
	"strconv"
	"strings"
)

const name = "superflat"

func init() {
	world.AddGenerator(name, func() world.Generator { return &SuperFlat{} })
}

//Vanilla presets
var (
	ClassicFlat    = ParseString("2;7,2x3,2;1;village")
	TunnelersBream = ParseString("2;7,230x1,5x3,2;3;stronghold,biome_1,decoration,dungeon,mineshaft")
	WaterWorld     = ParseString("2;7,5x1,5x3,5x12,90x9;1;biome_1,village")
	Overworld      = ParseString("2;7,59x1,3x3,2;1;stronghold,biome_1,village,decoration,dungeon,lake,mineshaft,lava_lake")
	SnowyKingdom   = ParseString("2;7,59x1,3x3,2,78;12;biome_1,village")
	BottomlessPit  = ParseString("2;2x4,3x3,2;1;biome_1,village")
	Desert         = ParseString("2;7,3x1,52x24,8x12;2;stronghold,biome_1,village,decoration,dungeon,mineshaft")
	RedstoneReady  = ParseString("2;7,3x1,52x24;2;")
)

//A vanilla superflat world generator
type SuperFlat struct {
	layers []layer  `msgpack:"ignore"`
	biome  byte     `msgpack:"ignore"`
	extra  []string `msgpack:"ignore"`
}

type layer struct {
	count int
	block byte
}

//ParseString reads in a vanilla superflat world code and returns
//a superflat world generator.
//
//For example the code
//    2;7,2x3,2;1;village,stronghold
//works like this
//    2;                   //Version, only 2 is supported
//    7,2x3,2;             //A comma seperated list of blocks
//        7,               //One layer of bedrock
//        2x3,             //Two layers of dirt
//        2                //One layer of grass
//    1;                   //Plains biome
//    village,stronghold   //A comma seperated list of features for the world (Not supported)
func ParseString(code string) *SuperFlat {
	sf := &SuperFlat{}
	sf.parseString(code)
	return sf
}

func (sf *SuperFlat) parseString(code string) {
	args := strings.Split(code, ";")
	if args[0] != "2" {
		panic("Unsupported version " + args[0])
	}

	l := strings.Split(args[1], ",")
	sf.layers = make([]layer, len(l))
	for i, lay := range l {
		sf.layers[i].count = 1
		if strings.Contains(lay, "x") {
			parts := strings.Split(lay, "x")
			count, err := strconv.Atoi(parts[0])
			if err != nil {
				panic(err)
			}
			sf.layers[i].count = count
			block, err := strconv.Atoi(parts[1])
			if err != nil {
				panic(err)
			}
			sf.layers[i].block = byte(block)
			continue
		}
		block, err := strconv.Atoi(lay)
		if err != nil {
			panic(err)
		}
		sf.layers[i].block = byte(block)
	}

	biome, err := strconv.Atoi(args[2])
	if err != nil {
		panic(err)
	}
	sf.biome = byte(biome)

	e := strings.Split(args[3], ",")
	sf.extra = e
}

//Generates the chunk.
func (sf *SuperFlat) Generate(chunk *world.Chunk) {
	y := 0
	for _, layer := range sf.layers {
		for i := 0; i < layer.count; i++ {
			for x := 0; x < 16; x++ {
				for z := 0; z < 16; z++ {
					chunk.SetBlock(x, y, z, layer.block)
				}
			}
			y++
		}
	}
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			chunk.Biome[x|z<<4] = sf.biome
		}
	}

}

//The name of the generator
func (sf *SuperFlat) Name() string {
	return name
}

//The vinilla superflat world code used to generate this world
func (sf *SuperFlat) GenerationCode() string {
	var buf bytes.Buffer
	buf.WriteString("2;") //Version
	for i, layer := range sf.layers {
		if layer.count == 1 {
			buf.WriteString(strconv.Itoa(int(layer.block)))
		} else {
			buf.WriteString(strconv.Itoa(int(layer.count)))
			buf.WriteString("x")
			buf.WriteString(strconv.Itoa(int(layer.block)))
		}
		if i < len(sf.layers)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString(";")
	buf.WriteString(strconv.Itoa(int(sf.biome)))
	buf.WriteString(";")
	for i, extra := range sf.extra {
		buf.WriteString(extra)
		if i < len(sf.extra)-1 {
			buf.WriteString(",")
		}
	}
	return buf.String()
}

const saveKey = "SuperFlatGenerator"

type savableGenerator struct {
	GenerationCode string
}

//Saves the generator's settings to the world's storage
func (sf *SuperFlat) Load(w *world.World) {
	var sav savableGenerator
	w.Read(saveKey, &sav)
	sf.parseString(sav.GenerationCode)
}

//Saves the generator's settings to the world's storage
func (sf *SuperFlat) Save(w *world.World) {
	sav := savableGenerator{sf.GenerationCode()}
	w.Write(saveKey, &sav)
}
