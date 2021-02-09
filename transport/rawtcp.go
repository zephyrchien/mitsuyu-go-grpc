package transport

import (
	"bytes"
	"fmt"
	"github.com/ZephyrChien/Mitsuyu/common"
	"net"
	"strconv"
	"strings"
	"syscall"
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

func NewRawTCPFromRedirect(buf []byte, conn net.Conn) (*RawTCP, error) {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return nil, fmt.Errorf("RawTCP: Not a tcp connection")
	}
	fd, err := tcpConn.File() // be placed in blocking mode
	if err != nil {
		return nil, fmt.Errorf("RawTCP: %v", err)
	}
	defer fd.Close()
	addrbuf, err := syscall.GetsockoptIPv6Mreq(int(fd.Fd()), syscall.IPPROTO_IP, 80)
	if err != nil || len(addrbuf.Multiaddr) != 16 {
		return nil, fmt.Errorf("RawTCP: %v", err)
	}
	if err = syscall.SetNonblock(int(fd.Fd()), true); err != nil {
		return nil, fmt.Errorf("RawTCP: %v", err)
	}
	b := addrbuf.Multiaddr
	port := strconv.Itoa(int(b[2])<<8 | int(b[3]))
	host := net.IP(b[4:8]).String()
	addr := &common.Addr{Isdn: false, Host: host, Port: port}
	buffer := bytes.NewBuffer(buf)
	return &RawTCP{proto: "TCP", addr: addr, buffer: buffer, conn: conn}, nil
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
