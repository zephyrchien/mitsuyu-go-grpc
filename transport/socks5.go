package transport

import (
	"fmt"
	"github.com/ZephyrChien/Mitsuyu/common"
	"net"
	"strconv"
)

type Socks5 struct {
	conn net.Conn
	addr *common.Addr
}

func (s5 *Socks5) Addr() *common.Addr {
	return s5.addr
}

func (s5 *Socks5) Proto() string {
	return "socks5"
}

func (s5 *Socks5) Read(b []byte) (int, error) {
	return s5.conn.Read(b)
}
func (s5 *Socks5) Write(b []byte) (int, error) {
	return s5.conn.Write(b)
}
func (s5 *Socks5) Close() error {
	return s5.conn.Close()
}

func Socks5Handshake(buf []byte, conn net.Conn) (*Socks5, error) {
	var err error
	var addr *common.Addr
	if err = selMethod(buf); err != nil {
		return nil, wrapErrorSocks5(err)
	}
	if err = sendNoAuthMethod(conn); err != nil {
		return nil, wrapErrorSocks5(err)
	}
	if addr, err = recvDataRequest(conn); err != nil {
		return nil, wrapErrorSocks5(err)
	}
	if err = sendDataReply(conn); err != nil {
		return nil, wrapErrorSocks5(err)
	}
	s5 := &Socks5{conn: conn, addr: addr}
	return s5, nil
}

// Select noauth as auth method
func selMethod(buf []byte) error {
	if buf[0] != 0x05 {
		return fmt.Errorf("invalid version %d", buf[0])
	}

	var nmethods, method int8
	nmethods = int8(buf[1])
	for i := int8(0); i < nmethods; i++ {
		method = int8(buf[i+2])
		if method == 0x00 {
			break
		}
	}
	if method != 0x00 {
		return fmt.Errorf("method not supported")
	}
	return nil
}

func sendNoAuthMethod(conn net.Conn) error {
	_, err := conn.Write([]byte{0x05, 0x00})
	return err
}

// Receive data request
func recvDataRequest(conn net.Conn) (*common.Addr, error) {
	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	if buf[0] != 0x05 || buf[1] != 0x01 || n < 6 {
		return nil, fmt.Errorf("bad data request")
	}
	port := uint16(buf[n-2])<<8 | uint16(buf[n-1])
	atyp := int8(buf[3])
	var host string
	switch atyp {
	case 0x01:
		host = net.IP([]byte{buf[4], buf[5], buf[6], buf[7]}).String()
	case 0x03:
		l := int8(buf[4])
		b := make([]byte, 0, l)
		for i := int8(0); i < l; i++ {
			b = append(b, buf[5+i])
		}
		host = string(b)
	case 0x04:
		b := make([]byte, 16)
		for i := 0; i < 16; i++ {
			b = append(b, buf[4+i])
		}
		host = net.IP(b).String()
	}
	return &common.Addr{Isdn: atyp == 0x03, Host: host, Port: strconv.Itoa(int(port))}, nil
}

func sendDataReply(conn net.Conn) error {
	r := [10]byte{0: 0x05, 3: 0x01}
	_, err := conn.Write(r[:])
	return err
}

func wrapErrorSocks5(err error) error {
	return fmt.Errorf("Socks5 handshake: %v", err)
}
