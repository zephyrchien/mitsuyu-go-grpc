package server

import (
	//"context"
	"crypto/tls"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip" // install gzip
	"google.golang.org/grpc/metadata"
	"mitsuyu/client"
	"mitsuyu/common"
	"mitsuyu/mitsuyu"
	"mitsuyu/transport"
	"net"
	"os"
	"strings"
	"sync"
)

const BUFFERSIZE = 4096

type Server struct {
	addr        string
	serviceName string
	tls         *tls.Config
	logger      *common.Logger
	done        chan struct{}
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
	s.logger = common.NewLogger(config.LogLevel)
	return s, nil
}

func (s *Server) Addr() string {
	return s.addr
}

func (s *Server) GetLogger() *common.Logger {
	return s.logger
}

func (s *Server) Run() {
	s.done = make(chan struct{}, 0)
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		fmt.Printf("Server: Unable to bind %s, %v\n", s.addr, err)
		os.Exit(0)
	}
	defer lis.Close()
	var opts []grpc.ServerOption
	if s.tls != nil {
		creds := credentials.NewTLS(s.tls)
		opts = append(opts, grpc.Creds(creds))
	}
	ss := grpc.NewServer(opts...)
	mitsuyu.RegisterMitsuyuServer(ss, s, s.serviceName)
	go ss.Serve(lis)
	defer ss.Stop()
	<-s.done
}

func (s *Server) Stop() {
	defer func() {
		recover()
	}()
	close(s.done)
}

// grpc functions
func (s *Server) Proxy(stream mitsuyu.Mitsuyu_ProxyServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return fmt.Errorf("Proxy: Unknown headers")
	}
	in := transport.NewGRPCStreamServer(stream)
	var reuse bool
	for {
		header, err := in.SniffHeader()
		if err != nil {
			return fmt.Errorf("Proxy: %v", err)
		}
		// parse ctrl info
		if reuse,err=parseReuse(md,header);err!=nil{
			return fmt.Errorf("%v",err)
		}
		// start proxy
		out, err := decideDestination(md, header)
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		wg := new(sync.WaitGroup)
		wg.Add(2)
		go forward(wg, in, out)
		go reverse(wg, in, out)
		wg.Wait()
		if !reuse {
			break
		}
	}

	return nil
}

func parseReuse(md metadata.MD, header []byte) (is bool, err error){
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("Proxy: Unable to parse metadata %v", e)
		}
	}()
	reuse := md.Get("reuse")[0]
	// for a reused conn, parse metadata from data.head instead
	if is = reuse == "true";is {
		if header == nil {
			return is,fmt.Errorf("Reuse: Empty header\n")
		}
		mdstrs:=strings.Split(string(header),"\r\n")
		if len(mdstrs)%2 == 1 {
			return is,fmt.Errorf("Reuse: Unable to parse header\n")
		}
		var key string
		for i, s := range mdstrs {
			if i%2 == 0 {
				key = s
				continue
			}
			md.Set(key,s)
		}
	}
	return is, nil
}

func decideDestination(md metadata.MD, header []byte) (out transport.Outbound, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("Proxy: Unable to decide destination %v", e)
		}
	}()
	var isdn, host, port, dns, next string
	isdn = md.Get("isdn")[0]
	host = md.Get("xxhost")[0]
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
	if ip, err := ipLookup(host, dns); err != nil {
		addr = host + ":" + port
	} else {
		addr = ip + ":" + port
	}
	return net.Dial("tcp", addr)
}

func forward(wg *sync.WaitGroup, in *transport.GRPCStreamServer, out transport.Outbound) {
	defer out.Close()
	buf:=make([]byte,BUFFERSIZE)
	for {
		n, err := in.Read(buf)
		if err != nil {
			break
		}
		if _, err = out.Write(buf[:n]); err != nil {
			break
		}
	}
	wg.Done()
}

func reverse(wg *sync.WaitGroup, in *transport.GRPCStreamServer, out transport.Outbound) {
	defer out.Close()
	buf := make([]byte, BUFFERSIZE)
	for {
		n, err := out.Read(buf)
		if err != nil {
			in.WriteAll(&mitsuyu.Data{Head: []byte{transport.CMD,transport.CMD_EOF}})
			break
		}
		if err = in.WriteAll(&mitsuyu.Data{Data: buf[:n]}); err != nil {
			break
		}
	}
	wg.Done()
}
