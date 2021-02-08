package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/ZephyrChien/Mitsuyu/common"
	"github.com/ZephyrChien/Mitsuyu/mitsuyu"
	"github.com/ZephyrChien/Mitsuyu/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const BUFFERSIZE = 4096

type Client struct {
	local         string
	remote        string
	tls           *tls.Config
	compress      string
	serviceName   string
	strategyGroup []*common.Strategy
}

func New(config *common.ClientConfig) (*Client, error) {
	c := new(Client)
	if config.Local == "" || config.Remote == "" {
		return nil, fmt.Errorf("Common: Invalid address")
	}
	strs := strings.Split(config.Remote, ":")
	var remoteHost, remotePort string
	strslen := len(strs)
	if strslen < 2 {
		return nil, fmt.Errorf("Common: Invalid remote address")
	}
	remotePort = strs[strslen-1]
	remoteHost = strings.Join(strs[:strslen-1], ":")
	c.local = config.Local
	c.remote = remoteHost + ":" + remotePort

	c.serviceName = config.ServiceName

	c.compress = config.Compress

	// load tls config
	if config.TLS == "true" {
		sni := config.TLSSNI
		if sni == "" {
			sni = remoteHost
		}
		var certpool *x509.CertPool
		if config.TLSCA != "" {
			if cafile, err := ioutil.ReadFile(config.TLSCA); err != nil {
				return nil, fmt.Errorf("Common: Unable to load ca-file")
			} else {
				certpool = x509.NewCertPool()
				if ok := certpool.AppendCertsFromPEM(cafile); !ok {
					certpool = nil
				}
			}
		}
		c.tls = &tls.Config{
			RootCAs:            certpool,
			ServerName:         sni,
			InsecureSkipVerify: config.TLSVerify == "false",
		}
	}

	// load strategy
	c.strategyGroup = config.StrategyGroup
	return c, nil
}

func (c *Client) Local() string {
	return c.local
}

func (c *Client) Remote() string {
	return c.remote
}

func (c *Client) Run() {
	lis, err := net.Listen("tcp", c.local)
	if err != nil {
		return
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			continue
		}
		go c.deliver(conn)
	}
}

func (c *Client) CallMitsuyuProxy(md metadata.MD) (*transport.GRPCStreamClient, error) {
	var dialopts []grpc.DialOption
	if c.tls != nil {
		creds := credentials.NewTLS(c.tls)
		dialopts = append(dialopts, grpc.WithTransportCredentials(creds))
	} else {
		dialopts = append(dialopts, grpc.WithInsecure())
	}
	dialopts = append(dialopts, grpc.WithBlock())
	// dial
	ctxx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	grpcConn, err := grpc.DialContext(ctxx, c.remote, dialopts...)
	if err != nil {
		return nil, err
	}
	cc := mitsuyu.NewMitsuyuClient(grpcConn, c.serviceName)

	ctx := metadata.NewOutgoingContext(context.Background(), md)
	// call grpc func
	var callopts []grpc.CallOption
	if c.compress == "true" {
		callopts = append(callopts, grpc.UseCompressor(gzip.Name))
	}
	stream, err := cc.Proxy(ctx, callopts...)
	if err != nil {
		return nil, err
	}
	ccc := transport.NewGRPCStreamClient(grpcConn, stream)
	return ccc, nil
}

func (c *Client) deliver(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}
	if s5, err := transport.Socks5Handshake(buf[:n], conn); err == nil {
		c.handle(s5)
	} else if h, err := transport.HttpHandshake(buf[:n], conn); err == nil {
		c.handle(h)
	} else if rawTCP, err := transport.NewRawTCPWithSniff(buf[:n], conn); err == nil {
		c.handle(rawTCP)
	} else {
		fmt.Errorf("Inbound: Unknown protocol")
	}
}

func (c *Client) handle(in transport.Inbound) {
	md := metadata.New(map[string]string{
		"xxhost": in.Addr().Host,
		"port":   in.Addr().Port,
		"isdn":   strconv.FormatBool(in.Addr().Isdn),
		"dns":    "default",
		"next":   "null",
	})
	if allow := c.applyClientStrategy(in.Addr(), md); !allow {
		fmt.Printf("%-6s| %s:%s [blocked]\n", in.Proto(), in.Addr().Host, in.Addr().Port)
		return
	}

	ccc, err := c.CallMitsuyuProxy(md)
	if err != nil {
		return
	}
	stream := ccc.GetStream()
	wg := new(sync.WaitGroup)
	wg.Add(2)

	fmt.Printf("%-6s| %s:%s [dns: %s]\n", in.Proto(), in.Addr().Host, in.Addr().Port, md.Get("dns")[0])
	// forward
	go func() {
		defer ccc.Close()
		defer in.Close()
		buf := make([]byte, BUFFERSIZE)
		for {
			n, err := in.Read(buf)
			if err != nil {
				break
			}
			if err = stream.Send(&mitsuyu.Data{Data: buf[:n]}); err != nil {
				break
			}
		}
		if h, ok := in.(*transport.Http); ok && !h.IsTun() {
			time.Sleep(4 * time.Second)
		}
		wg.Done()
	}()
	// reverse
	go func() {
		defer ccc.Close()
		defer in.Close()
		for {
			r, err := stream.Recv()
			if err != nil {
				break
			}
			if _, err = in.Write(r.GetData()); err != nil {
				break
			}
		}
		wg.Done()
	}()
	wg.Wait()
}

func (c *Client) applyClientStrategy(addr *common.Addr, md metadata.MD) (allow bool) {
	if blockReservedAddr(addr) {
		return false
	}
	var matched = false
	var index = 0
	for i, rules := range c.strategyGroup {
		if matchRules(addr, rules) {
			index = i
			matched = true
			break
		}
	}
	if matched {
		if c.strategyGroup[index].Block == "true" {
			return false
		}
		if dns := c.strategyGroup[index].DNS; dns != "" {
			md.Set("dns", dns)
		}
		if next := c.strategyGroup[index].Next; next != "" {
			md.Set("next", next)
			md.Set("next_service_name", c.serviceName)
		}
	}
	return true
}
