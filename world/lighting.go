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
	"github.com/NetherrackDev/netherrack/blocks"
)

type lightMode int

const (
	lightAdd lightMode = iota
	lightRemove
	lightUpdate
)

type lightRequest struct {
	LightLevel byte
	Mode       lightMode
	World      *World
	X, Y, Z    int
}

func lightWorker(lightChan chan lightRequest, ret chan struct{}) {
lightLoop:
	for {
		select {
		case req := <-lightChan:
			switch req.Mode {
			case lightAdd: //Only for block lighting
				req.World.setLight(req.X, req.Y, req.Z, req.LightLevel, false)
				updates := propagateLight(req.World, req.X, req.Y, req.Z, req.LightLevel, false)
				for len(updates) != 0 {
					u := updates[0]
					nu := propagateLight(u.world, u.x, u.y, u.z, u.level, false)
					copy(updates, updates[1:])
					updates = updates[:len(updates)-1]
					updates = append(updates, nu...)
				}
			case lightRemove:
				req.World.setLight(req.X, req.Y, req.Z, 0, false)
				updates, reprops := removeLight(req.World, req.X, req.Y, req.Z, req.LightLevel, false)
				for len(updates) != 0 {
					u := updates[0]
					nu, nr := removeLight(u.world, u.x, u.y, u.z, u.level, false)
					copy(updates, updates[1:])
					updates = updates[:len(updates)-1]
					updates = append(updates, nu...)
					reprops = append(reprops, nr...)
				}

				for _, u := range reprops {
					max := int8(0)
					for _, off := range offsets {
						l := int8(u.world.Light(u.x+off.x, u.y+off.y, u.z+off.z, false))
						if l > max {
							max = l
						}
					}
					b, _ := u.world.Block(u.x, u.y, u.z)
					bl := blocks.Blocks[b]
					max -= 1
					if bl.LightFiltered != 0 {
						max -= int8(bl.LightFiltered) - 1
					}

					if max > 0 {
						u.world.setLight(u.x, u.y, u.z, byte(max), false)
						updates := propagateLight(u.world, u.x, u.y, u.z, byte(max), false)
						for len(updates) != 0 {
							u2 := updates[0]
							nu := propagateLight(u2.world, u2.x, u2.y, u2.z, u2.level, false)
							updates = append(append(updates[:0], updates[1:]...), nu...)
						}
					} else {
						u.world.setLight(u.x, u.y, u.z, 0, false)
					}
				}

			case lightUpdate:
				max := int8(0)
				for _, off := range offsets {
					l := int8(req.World.Light(req.X+off.x, req.Y+off.y, req.Z+off.z, false))
					if l > max {
						max = l
					}
				}
				b, _ := req.World.Block(req.X, req.Y, req.Z)
				bl := blocks.Blocks[b]
				max -= 1
				if bl.LightFiltered != 0 {
					max -= int8(bl.LightFiltered) - 1
				}

				if max > 0 {
					req.World.setLight(req.X, req.Y, req.Z, byte(max), false)
					updates := propagateLight(req.World, req.X, req.Y, req.Z, byte(max), false)
					for len(updates) != 0 {
						u := updates[0]
						nu := propagateLight(u.world, u.x, u.y, u.z, u.level, false)
						updates = append(append(updates[:0], updates[1:]...), nu...)
					}
				} else {
					req.World.setLight(req.X, req.Y, req.Z, 0, false)
				}
			}
		default:
			break lightLoop
		}
	}
	ret <- struct{}{}
}

type offset struct {
	x, y, z int
}

var offsets = [6]offset{
	{1, 0, 0},
	{-1, 0, 0},
	{0, 1, 0},
	{0, -1, 0},
	{0, 0, 1},
	{0, 0, -1},
}

type lightPropData struct {
	world   *World
	x, y, z int
	level   byte
}

func removeLight(world *World, x, y, z int, level byte, sky bool) (updates []lightPropData, reprops []lightPropData) {
	for _, off := range offsets {
		b, _ := world.Block(x+off.x, y+off.y, z+off.z)
		l := int8(world.Light(x+off.x, y+off.y, z+off.z, sky))
		if l == 0 {
			continue
		}
		tempLevel := int8(level)
		tempLevel -= 1
		bl := blocks.Blocks[b]
		if bl.LightFiltered != 0 {
			tempLevel -= int8(bl.LightFiltered) - 1
		}
		if tempLevel == l {
			world.setLight(x+off.x, y+off.y, z+off.z, 0, sky)
			updates = append(updates, lightPropData{world, x + off.x, y + off.y, z + off.z, byte(tempLevel)})
		} else if l > tempLevel {
			reprops = append(updates, lightPropData{world, x, y, z, byte(0)})
		}
	}
	return
}

func propagateLight(world *World, x, y, z int, level byte, sky bool) (updates []lightPropData) {
	for _, off := range offsets {
		b, _ := world.Block(x+off.x, y+off.y, z+off.z)
		l := int8(world.Light(x+off.x, y+off.y, z+off.z, sky))
		tempLevel := int8(level)
		tempLevel -= 1
		bl := blocks.Blocks[b]
		if bl.LightFiltered != 0 {
			tempLevel -= int8(bl.LightFiltered) - 1

		}
		if tempLevel > l {
			world.setLight(x+off.x, y+off.y, z+off.z, byte(tempLevel), sky)
			updates = append(updates, lightPropData{world, x + off.x, y + off.y, z + off.z, byte(tempLevel)})
		}
	}
	return
}
