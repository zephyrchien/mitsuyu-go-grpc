// +build win

package transport

import (
	"fmt"
	"net"
)

func NewRawTCPFromRedirect(buf []byte, conn net.Conn) (*RawTCP, error) {
	return nil, fmt.Errorf("RawTCP: Support unix only")
}
