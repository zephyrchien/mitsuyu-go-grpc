package transport

import (
	"net"
)

type RawTCP struct {
	conn net.Conn
}

func NewRawTCP(conn net.Conn) *RawTCP {
	return &RawTCP{conn: conn}
}

func (c *RawTCP) ShallowRead(b *[]byte) (int, error) {
	return c.conn.Read(*b)
}

func (c *RawTCP) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}
func (c *RawTCP) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c *RawTCP) Close() error {
	return c.conn.Close()
}
