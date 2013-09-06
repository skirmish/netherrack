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

const Version = 74

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
	Authenticate(handshake Handshake, serverID string, sharedSecret, publicKey []byte) error
}

//Auths the user and returns their username
func (conn *Conn) Login(handshake Handshake, authenticator Authenticator) (string, error) {
	if handshake.ProtocolVersion != Version {
		if handshake.ProtocolVersion < Version {
			return "", ErrorOutOfDateClient
		} else {
			return "", ErrorOutOfDateServer
		}
	}

	verifyToken := make([]byte, 16) //Used by the server to check encryption is working correctly
	rand.Read(verifyToken)

	serverBytes := make([]byte, 10)
	rand.Read(serverBytes)
	serverID := hex.EncodeToString(serverBytes)

	conn.WritePacket(EncryptionKeyRequest{
		ServerID:    serverID,
		PublicKey:   publicKeyBytes,
		VerifyToken: verifyToken,
	})

	packet, err := conn.ReadPacket()
	if err != nil {
		return handshake.Username, err
	}
	encryptionResponse, ok := packet.(EncryptionKeyResponse)
	if !ok {
		err = fmt.Errorf("Unexpected packet: %x", packet.ID())
		return handshake.Username, err
	}

	sharedSecret, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptionResponse.SharedSecret)
	if err != nil {
		return handshake.Username, err
	}

	verifyTokenResponse, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptionResponse.VerifyToken)
	if err != nil {
		return handshake.Username, err
	}

	if !bytes.Equal(verifyToken, verifyTokenResponse) {
		return handshake.Username, ErrorVerifyFailed
	}

	if err := authenticator.Authenticate(handshake, serverID, sharedSecret, publicKeyBytes); err != nil {
		return handshake.Username, err
	}

	conn.WritePacket(EncryptionKeyResponse{})

	aesCipher, err := aes.NewCipher(sharedSecret)
	if err != nil {
		return handshake.Username, err
	}

	conn.In = cipher.StreamReader{
		R: conn.In,
		S: newCFB8Decrypt(aesCipher, sharedSecret),
	}
	conn.Out = cipher.StreamWriter{
		W: conn.Out,
		S: newCFB8Encrypt(aesCipher, sharedSecret),
	}

	packet, err = conn.ReadPacket()
	if err != nil {
		return handshake.Username, err
	}
	var clientStatuses ClientStatuses
	clientStatuses, ok = packet.(ClientStatuses)
	if !ok {
		err = fmt.Errorf("Unexpected packet: %x", packet.ID())
		return handshake.Username, err
	}
	if clientStatuses.Payload != 0x00 {
		return handshake.Username, ErrorEncryption
	}

	return handshake.Username, nil
}
