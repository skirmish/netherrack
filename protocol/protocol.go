package protocol

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/thinkofdeath/soulsand"
	"github.com/thinkofdeath/soulsand/server"
	"io"
	"io/ioutil"
	"log"
	mrand "math/rand"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"
)

var PROTOVERSION byte

type Conn struct {
	conn net.Conn

	output io.Writer
	input  io.Reader
}

var (
	certBytes  []byte
	privateKey *rsa.PrivateKey
)

func init() {
	p, err := rsa.GenerateKey(rand.Reader, 1024)
	privateKey = p
	if err != nil {
		log.Println(err)
		return
	}
	privateKey.Precompute()

	certBytes, err = x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Println(err)
		return
	}
}

func NewConnection(conn net.Conn) (*Conn, string) {
	connection := &Conn{conn: conn, output: conn, input: conn}
	protoVersion, username, _, _ := connection.ReadHandshake()
	if protoVersion != PROTOVERSION {
		if protoVersion < PROTOVERSION {
			connection.WriteDisconnect("Client out of date")
		} else {
			connection.WriteDisconnect("Server out of date")
		}
		runtime.Goexit()
	}
	if server.GetFlag(soulsand.RANDOM_NAMES) {
		ext := fmt.Sprintf("%d", mrand.Int31n(9999))
		if len(username)+len(ext) > 16 {
			username = username[:16-len(ext)] + ext
		} else {
			username += ext
		}
	}
	log.Printf("Player %s connecting\n", username)

	token := make([]byte, 16)
	rand.Read(token)

	sByte := make([]byte, 4)
	binary.BigEndian.PutUint32(sByte, uint32(mrand.Int()))
	serverID := hex.EncodeToString(sByte)

	connection.WriteEncryptionKeyRequest(serverID, certBytes, token)

	if connection.readUByte() != 0xFC {
		connection.WriteDisconnect("Protocol error")
		runtime.Goexit()
	}

	sharedSecret, verifyTokenResponse := connection.ReadEncryptionKeyResponse()

	sharedSecretKey, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, sharedSecret)
	if err != nil {
		connection.WriteDisconnect("Protocol error")
		runtime.Goexit()
	}

	verifyTokenResponse, err = rsa.DecryptPKCS1v15(rand.Reader, privateKey, verifyTokenResponse)
	if err != nil {
		connection.WriteDisconnect("Protocol error")
		runtime.Goexit()
	}

	if !bytes.Equal(token, verifyTokenResponse) {
		connection.WriteDisconnect("Protocol error")
		runtime.Goexit()
	}

	//Auth client
	sha := sha1.New()
	sha.Write([]byte(serverID))
	sha.Write([]byte(sharedSecretKey))
	sha.Write([]byte(certBytes))
	hash := sha.Sum(make([]byte, 0))
	isNegative := (hash[0] & 0x80) == 0x80
	if isNegative {
		hash = twosCompliment(hash)
	}
	buf := hex.EncodeToString(hash)
	if isNegative {
		buf = "-" + buf
	}
	hashString := strings.TrimLeft(buf, "0")

	if !server.GetFlag(soulsand.OFFLINE_MODE) {
		response, err := http.Get(fmt.Sprintf("http://session.minecraft.net/game/checkserver.jsp?user=%s&serverId=%s", username, hashString))
		if err != nil {
			connection.WriteDisconnect("Failed to connect to minecraft auth servers")
			runtime.Goexit()
		}
		defer response.Body.Close()
		responseBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			connection.WriteDisconnect("Failed to connect to minecraft auth servers")
			runtime.Goexit()
		}
		if string(responseBytes) != "YES" {
			connection.WriteDisconnect("Auth failed")
			runtime.Goexit()
		}
	}

	connection.WriteEncryptionKeyResponse([]byte{}, []byte{})

	aesCipher, err := aes.NewCipher(sharedSecretKey)
	if err != nil {
		connection.WriteDisconnect("Protocol error")
		runtime.Goexit()
	}

	connection.input = cipher.StreamReader{
		R: connection.conn,
		S: NewCFB8Decrypt(aesCipher, sharedSecretKey),
	}

	connection.output = cipher.StreamWriter{
		W: connection.conn,
		S: NewCFB8Encrypt(aesCipher, sharedSecretKey),
	}

	cdPacket := make([]byte, 2)
	connection.Read(cdPacket)
	if !bytes.Equal(cdPacket, []byte{0xCD, 0x00}) {
		connection.WriteDisconnect("Protocol error")
		log.Println(cdPacket)
		runtime.Goexit()
	}

	return connection, username
}

func twosCompliment(p []byte) []byte {
	carry := true
	for i := len(p) - 1; i >= 0; i-- {
		p[i] = ^p[i]
		if carry {
			carry = p[i] == 0xFF
			p[i]++
		}
	}
	return p
}

func (c *Conn) Write(b []byte) (int, error) {
	c.conn.SetWriteDeadline(time.Now().Add(15 * time.Second))
	n, err := c.output.Write(b)
	if err != nil {
		runtime.Goexit()
	}
	return n, err
}

func (c *Conn) Read(b []byte) (int, error) {
	c.conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	n, err := io.ReadFull(c.input, b)
	if err != nil {
		runtime.Goexit()
	}
	return n, err
}
