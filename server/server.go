package server

import (
	//"context"
	"crypto/tls"
	"fmt"
	"github.com/ZephyrChien/Mitsuyu/client"
	"github.com/ZephyrChien/Mitsuyu/common"
	"github.com/ZephyrChien/Mitsuyu/mitsuyu"
	"github.com/ZephyrChien/Mitsuyu/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip" // install gzip
	"google.golang.org/grpc/metadata"
	"net"
	"sync"
)

const BUFFERSIZE = 4096

type Server struct {
	addr        string
	serviceName string
	tls         *tls.Config
	mitsuyu.UnimplementedMitsuyuServer
}

func New(config *common.ServerConfig) (*Server, error) {
	s := &Server{addr: config.Addr, serviceName: config.ServiceName}
	if config.TLS == "true" {
		cert, err := tls.LoadX509KeyPair(config.TLSCert, config.TLSKey)
		if err != nil {
			return nil, fmt.Errorf("Common: Invalid cert or key")
		}
		s.tls = &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.NoClientCert,
		}
	}
	return s, nil
}

func (s *Server) Addr() string {
	return s.addr
}

func (s *Server) Serve() {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return
	}
	var opts []grpc.ServerOption
	if s.tls != nil {
		creds := credentials.NewTLS(s.tls)
		opts = append(opts, grpc.Creds(creds))
	}
	ss := grpc.NewServer(opts...)
	mitsuyu.RegisterMitsuyuServer(ss, s, s.serviceName)
	ss.Serve(lis)
}

// grpc functions
func (s *Server) Proxy(stream mitsuyu.Mitsuyu_ProxyServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return fmt.Errorf("Proxy: Unknown headers")
	}

	// start proxy
	out, err := decideDestination(md)
	if err != nil {
		return fmt.Errorf("Proxy: %v", err)
	}
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go forward(wg, out, stream)
	go reverse(wg, out, stream)
	wg.Wait()
	return nil
}

func decideDestination(md metadata.MD) (out transport.Outbound, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("Proxy: Unable to decide destination %v", e)
		}
	}()
	var isdn, host, port, dns, next string
	isdn = md.Get("isdn")[0]
	host = md.Get("host")[0]
	port = md.Get("port")[0]
	dns = md.Get("dns")[0]
	next = md.Get("next")[0]
	// proxy chain
	if next != "null" {
		md.Set("next", "null")
		serviceNames := md.Get("next_service_name")
		if len(serviceNames) == 0 {
			return nil, fmt.Errorf("Proxy: Invalid next service name")
		}
		conf := &common.ClientConfig{
			Local:       "null",
			Remote:      next,
			ServiceName: serviceNames[0],
			TLS:         "true",
			TLSVerify:   "true",
			Compress:    "true",
		}
		c, err := client.New(conf)
		if err != nil {
			return nil, err
		}
		return c.CallMitsuyuProxy(md)
	}
	// dns
	var addr string
	if dns == "default" || isdn == "false" {
		return net.Dial("tcp", host+":"+port)
	}
	if addr, err = ipLookup(host, dns); err != nil {
		addr = host + ":" + port
	}
	return net.Dial("tcp", addr)
}

func forward(wg *sync.WaitGroup, out transport.Outbound, stream mitsuyu.Mitsuyu_ProxyServer) {
	defer out.Close()
	for {
		r, err := stream.Recv()
		if err != nil {
			break
		}
		if _, err = out.Write(r.GetData()); err != nil {
			break
		}
	}
	wg.Done()
}

func reverse(wg *sync.WaitGroup, out transport.Outbound, stream mitsuyu.Mitsuyu_ProxyServer) {
	defer out.Close()
	buf := make([]byte, BUFFERSIZE)
	for {
		n, err := out.Read(buf)
		if err != nil {
			break
		}
		if err = stream.Send(&mitsuyu.Data{Data: buf[:n]}); err != nil {
			break
		}
	}
	wg.Done()
}
