package transport

import (
	"github.com/ZephyrChien/Mitsuyu/mitsuyu"
	"google.golang.org/grpc"
)

type GRPCStreamClient struct {
	conn   *grpc.ClientConn
	stream mitsuyu.Mitsuyu_ProxyClient
}

func NewGRPCStreamClient(conn *grpc.ClientConn, stream mitsuyu.Mitsuyu_ProxyClient) *GRPCStreamClient {
	return &GRPCStreamClient{conn: conn, stream: stream}
}

func (c *GRPCStreamClient) GetStream() mitsuyu.Mitsuyu_ProxyClient {
	return c.stream
}

func (c *GRPCStreamClient) Read(b []byte) (int, error) {
	r, err := c.stream.Recv()
	if err != nil {
		return 0, err
	}
	n := copy(b, r.GetData())
	return n, nil
}
func (c *GRPCStreamClient) ShallowRead(b *[]byte) (int, error) {
	r, err := c.stream.Recv()
	if err != nil {
		return 0, err
	}
	*b = r.GetData()
	return len(*b), nil
}

func (c *GRPCStreamClient) Write(b []byte) (int, error) {
	return len(b), c.stream.Send(&mitsuyu.Data{Data: b})
}

func (c *GRPCStreamClient) Close() error {
	return c.conn.Close()
}
