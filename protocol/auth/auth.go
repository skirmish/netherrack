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

package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NetherrackDev/netherrack/protocol"
	"net/http"
)

var Instance = Authenticator{}

type Authenticator struct{}

var ErrorAuthFailed = errors.New("Authentication failed")

type jsonResponse struct {
	ID string `json:"id"`
}

//Checks the users against the Mojang login servers
func (Authenticator) Authenticate(handshake protocol.Handshake, serverID string, sharedSecret, publicKey []byte) (string, error) {

	response, err := http.Get(fmt.Sprintf("https://sessionserver.mojang.com/session/minecraft/hasJoined?username=%s&serverId=%s", handshake.Username, serverID))
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	dec := json.NewDecoder(response.Body)
	res := &jsonResponse{}
	err = dec.Decode(res)
	if err != nil {
		return "", ErrorAuthFailed
	}

	if len(res.ID) != 32 {
		return "", ErrorAuthFailed
	}

	return res.ID, nil
}

func twosCompliment(p []byte) {
	carry := true
	for i := len(p) - 1; i >= 0; i-- {
		p[i] = ^p[i]
		if carry {
			carry = p[i] == 0xFF
			p[i]++
		}
	}
}
