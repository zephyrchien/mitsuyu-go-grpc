package transport

import (
	"github.com/ZephyrChien/Mitsuyu/common"
)

// socks5, http, tcp
type Inbound interface {
	Addr() *common.Addr
	Proto() string
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
