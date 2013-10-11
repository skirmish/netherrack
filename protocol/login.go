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

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
)

var (
	publicKeyBytes []byte
	privateKey     *rsa.PrivateKey
)

const Version = 0

func init() {
	var err error
	privateKey, err = rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	privateKey.Precompute()

	publicKeyBytes, err = x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}
}

var (
	ErrorOutOfDateServer = errors.New("Server out of date")
	ErrorOutOfDateClient = errors.New("Client out of date")
	ErrorVerifyFailed    = errors.New("Verify Token incorrect")
	ErrorEncryption      = errors.New("Encryption error")
)

//An Authenticator should check if the user is who they say they are.
//Normaly this involves checking against the Mojang servers.
type Authenticator interface {
	Authenticate(username string, serverID string, sharedSecret, publicKey []byte) (uuid string, err error)
}

//Auths the user and returns their username.
//Uses infomation from http://wiki.vg/Protocol_Encryption
func (conn *Conn) Login(handshake Handshake, authenticator Authenticator) (username string, uuid string, err error) {
	if handshake.ProtocolVersion != Version {
		if handshake.ProtocolVersion < Version {
			return "", uuid, ErrorOutOfDateClient
		} else {
			return "", uuid, ErrorOutOfDateServer
		}
	}

	conn.State = Login

	packet, err := conn.ReadPacket()
	if err != nil {
		return
	}
	lStart, ok := packet.(LoginStart)
	if !ok {
		err = fmt.Errorf("Unexpected packet: %x", packet.ID())
		return
	}
	username = lStart.Username

	verifyToken := make([]byte, 16) //Used by the server to check encryption is working correctly
	rand.Read(verifyToken)

	var serverID = "-"
	if authenticator != nil {
		serverBytes := make([]byte, 10)
		rand.Read(serverBytes)
		serverID = hex.EncodeToString(serverBytes)
	}

	conn.WritePacket(EncryptionKeyRequest{
		ServerID:    serverID,
		PublicKey:   publicKeyBytes,
		VerifyToken: verifyToken,
	})

	packet, err = conn.ReadPacket()
	if err != nil {
		return
	}
	encryptionResponse, ok := packet.(EncryptionKeyResponse)
	if !ok {
		err = fmt.Errorf("Unexpected packet: %x", packet.ID())
		return
	}

	sharedSecret, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptionResponse.SharedSecret)
	if err != nil {
		return
	}

	verifyTokenResponse, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptionResponse.VerifyToken)
	if err != nil {
		return
	}
	if !bytes.Equal(verifyToken, verifyTokenResponse) {
		return
	}

	if authenticator != nil {
		if uuid, err = authenticator.Authenticate(username, serverID, sharedSecret, publicKeyBytes); err != nil {
			return
		}
	} else {
		idBytes := make([]byte, 16)
		rand.Read(idBytes)
		uuid = hex.EncodeToString(idBytes)
	}
	//conn.WritePacket(EncryptionKeyResponse{})

	aesCipher, err := aes.NewCipher(sharedSecret)
	if err != nil {
		return
	}

	conn.In = cipher.StreamReader{
		R: conn.In,
		S: newCFB8Decrypt(aesCipher, sharedSecret),
	}
	conn.Out = cipher.StreamWriter{
		W: conn.Out,
		S: newCFB8Encrypt(aesCipher, sharedSecret),
	}
	conn.WritePacket(LoginSuccess{uuid, username})
	conn.State = Play

	return
}
