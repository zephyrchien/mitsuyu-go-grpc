package transport

import (
	"errors"
	"bytes"
	"google.golang.org/grpc"
	"mitsuyu/mitsuyu"
)

const CMD = 0xff
const CMD_EOF = 0x00

var ERR_CMD_EOF = errors.New("CMD: Close")

type GRPCStreamClient struct {
	conn   *grpc.ClientConn
	stream mitsuyu.Mitsuyu_ProxyClient
}

type GRPCStreamServer struct {
	buffer *bytes.Buffer
	stream mitsuyu.Mitsuyu_ProxyServer
}

// client

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

func (c *GRPCStreamClient) ReadAll() (*mitsuyu.Data, error) {
	return c.stream.Recv()
}

func (c *GRPCStreamClient) Write(b []byte) (int, error) {
	return len(b), c.stream.Send(&mitsuyu.Data{Data: b})
}

func (c *GRPCStreamClient) WriteAll(data *mitsuyu.Data) error {
	return c.stream.Send(data)
}

func (c *GRPCStreamClient) Close() error {
	return c.conn.Close()
}

// server

func NewGRPCStreamServer(stream mitsuyu.Mitsuyu_ProxyServer) *GRPCStreamServer {
	return &GRPCStreamServer{stream: stream}
}

func (s *GRPCStreamServer) GetStream() mitsuyu.Mitsuyu_ProxyServer {
	return s.stream
}

func (s *GRPCStreamServer) Read(b []byte) (int, error) {
	if s.buffer != nil {
		n, err := s.buffer.Read(b)
		s.buffer = nil
		return n, err
	}
	r, err := s.stream.Recv()
	if err != nil {
		return 0, err
	}
	if head:=r.GetHead();len(head)==2 &&head[0]==CMD&&head[1]==CMD_EOF{
		return 2, ERR_CMD_EOF
	}
	n := copy(b, r.GetData())
	return n, nil
}

func (s *GRPCStreamServer) ReadAll() (*mitsuyu.Data, error) {
	return s.stream.Recv()
}

func (s *GRPCStreamServer) Write(b []byte) (int, error) {
	return len(b), s.stream.Send(&mitsuyu.Data{Data: b})
}

func (s *GRPCStreamServer) WriteAll(data *mitsuyu.Data) error {
	return s.stream.Send(data)
}

func (s *GRPCStreamServer) SniffHeader() ([]byte, error) {
	r, err := s.stream.Recv()
	if err != nil {
		return nil, err
	}
	s.buffer = bytes.NewBuffer(r.GetData())
	return r.GetHead(), nil
}
