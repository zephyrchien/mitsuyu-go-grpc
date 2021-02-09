package transport

import (
	"bytes"
	"github.com/ZephyrChien/Mitsuyu/common"
)

// socks5, http, tcp
type Inbound interface {
	Addr() *common.Addr
	Proto() string
	SetAddr(*common.Addr)
	SetBuffer(*bytes.Buffer)
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
}

// grpc(client), tcp
type Outbound interface {
	//ShallowRead(b *[]byte) (int, error)
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
}
