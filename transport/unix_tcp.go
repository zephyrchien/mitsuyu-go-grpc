// +build !win

package transport

import (
	"bytes"
	"fmt"
	"mitsuyu/common"
	"net"
	"strconv"
	"syscall"
)

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
