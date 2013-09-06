package auth

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/NetherrackDev/netherrack/protocol"
	"io/ioutil"
	"net/http"
	"strings"
)

var Instance = Authenticator{}

type Authenticator struct{}

var ErrorAuthFailed = errors.New("Authentication failed")

//Checks the users against the Mojang login servers
func (Authenticator) Authenticate(handshake protocol.Handshake, serverID string, sharedSecret, publicKey []byte) error {
	sha := sha1.New()
	sha.Write([]byte(serverID))
	sha.Write(sharedSecret)
	sha.Write(publicKey)
	hash := sha.Sum(nil)

	negative := (hash[0] & 0x80) == 0x80
	if negative {
		twosCompliment(hash)
	}

	buf := hex.EncodeToString(hash)
	if negative {
		buf = "-" + buf
	}
	hashString := strings.TrimLeft(buf, "0")

	response, err := http.Get(fmt.Sprintf("http://session.minecraft.net/game/checkserver.jsp?user=%s&serverId=%s", handshake.Username, hashString))
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if string(responseBytes) != "YES" {
		return ErrorAuthFailed
	}

	return nil
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
