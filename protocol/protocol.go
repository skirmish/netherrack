package protocol

import (
	"runtime"
	//"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"io"
	"log"
	"net"
	"time"
)

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
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
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

func NewConnection(conn net.Conn, auth bool) *Conn {
	connection := &Conn{conn: conn}
	return connection
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
	n, err := io.ReadFull(c.conn, b)
	if err != nil {
		runtime.Goexit()
	}
	return n, err
}
