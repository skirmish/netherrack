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

package player

import (
	"github.com/NetherrackDev/netherrack/protocol"
)

type BlockPlacement struct {
	Packet protocol.PlayerBlockPlacement
	Return chan<- struct{}
}

func (p *Player) SetBlockPlacementEvent(event chan<- BlockPlacement) {
	p.event.Lock()
	p.event.blockPlace = event
	p.event.Unlock()
}

type BlockDig struct {
	Packet protocol.PlayerDigging
	Return chan<- struct{}
}

func (p *Player) SetBlockDigEvent(event chan<- BlockDig) {
	p.event.Lock()
	p.event.blockDig = event
	p.event.Unlock()
}

type EnterWorld struct {
	Packet *protocol.LoginRequest
	Return chan<- struct{}
}

func (p *Player) SetEnterWorldEvent(event chan<- EnterWorld) {
	p.event.Lock()
	p.event.enterWorld = event
	p.event.Unlock()
}
