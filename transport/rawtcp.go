package transport

import (
	"bytes"
	"github.com/ZephyrChien/Mitsuyu/common"
	"net"
	"strings"
)

type RawTCP struct {
	proto  string
	addr   *common.Addr
	buffer *bytes.Buffer
	conn   net.Conn
}

func NewRawTCP(conn net.Conn) *RawTCP {
	return &RawTCP{conn: conn}
}

func NewRawTCPWithSniff(buf []byte, conn net.Conn) (*RawTCP, error) {
	var host, port, proto string
	var err error
	if host, err = SniffFromHTTP(buf); err == nil {
		s := strings.SplitN(host, ":", 2)
		host = s[0]
		if len(s) == 1 {
			port = "80"
		} else {
			port = s[1]
		}
		proto = "http"
	} else if host, err = SniffFromHTTPS(buf); err == nil {
		port = "443"
		proto = "https"
	}
	if err != nil {
		return nil, err
	}
	var isdn = net.ParseIP(host) == nil
	addr := &common.Addr{Isdn: isdn, Host: host, Port: port}
	buffer := bytes.NewBuffer(buf)
	return &RawTCP{proto: proto, addr: addr, buffer: buffer, conn: conn}, nil
}

func (c *RawTCP) Addr() *common.Addr {
	return c.addr
}

func (c *RawTCP) Proto() string {
	return c.proto
}

func (c *RawTCP) SetAddr(addr *common.Addr) {
	c.addr = addr
}

func (c *RawTCP) SetBuffer(buffer *bytes.Buffer) {
	c.buffer = buffer
}

func (c *RawTCP) ShallowRead(b *[]byte) (int, error) {
	return c.conn.Read(*b)
}

func (c *RawTCP) Read(b []byte) (int, error) {
	if c.buffer == nil {
		return c.conn.Read(b)
	}
	n, err := c.buffer.Read(b)
	c.buffer = nil
	return n, err
}

func (c *RawTCP) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c *RawTCP) Close() error {
	return c.conn.Close()
}
