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

package protocol

type LoginDisconnect struct {
	Data string
}

func (LoginDisconnect) ID() byte { return 0x00 }

type EncryptionKeyResponse struct {
	SharedSecret []byte `ltype:"int16"`
	VerifyToken  []byte `ltype:"int16"`
}

func (EncryptionKeyResponse) ID() byte { return 0x01 }

type LoginSuccess struct {
	UUID     string
	Username string
}

func (LoginSuccess) ID() byte { return 0x02 }

type LoginStart struct {
	Username string
}

func (LoginStart) ID() byte { return 0x00 }

type EncryptionKeyRequest struct {
	ServerID    string
	PublicKey   []byte `ltype:"int16"`
	VerifyToken []byte `ltype:"int16"`
}

func (EncryptionKeyRequest) ID() byte { return 0x01 }
