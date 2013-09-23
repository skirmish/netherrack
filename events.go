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

package netherrack

import (
	"github.com/NetherrackDev/netherrack/entity/player"
	"net"
)

type OldPingEvent struct {
	Addr     net.Addr
	Response chan<- string
}

func (server *Server) SetOldPingEvent(event chan<- OldPingEvent) {
	server.event.Lock()
	server.event.oldPingEvent = event
	server.event.Unlock()
}

type PingEvent struct {
	Addr            net.Addr
	ProtocolVersion byte
	TargetHost      string
	TargetPort      int
	Response        chan<- string
}

func (server *Server) SetPingEvent(event chan<- PingEvent) {
	server.event.Lock()
	server.event.pingEvent = event
	server.event.Unlock()
}

type PlayerJoinEvent struct {
	Player *player.Player
	Return chan<- struct{}
}

func (server *Server) SetPlayerJoinEvent(event chan<- PlayerJoinEvent) {
	server.event.Lock()
	server.event.playerJoinEvent = event
	server.event.Unlock()
}
