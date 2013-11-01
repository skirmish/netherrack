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
	"github.com/NetherrackDev/netherrack/message"
)

//Ping is json encoded before being sent to the the client.
//Version and Players.Online will automaticly be replaced with
//the correct details
type Ping struct {
	Version     PingVersion     `json:"version"`
	Players     PingPlayers     `json:"players"`
	Description message.Message `json:"description"`
	Favicon     string          `json:"favicon,omitempty"`
}

//PingVersion is the version section of minecraft's server list ping
type PingVersion struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

//PingPlayers contains the number of currently online players and the
//max number of players. It can optionally contain a sample list of players.
//Online will automatically be replaced with the number of online players
type PingPlayers struct {
	Max    int          `json:"max"`
	Online int          `json:"online"`
	Sample []PingPlayer `json:"sample,omitempty"`
}

//PingPlayer is a player to be displayed on the sample player list in the
//minecraft client
type PingPlayer struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}
