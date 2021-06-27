package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
	"mitsuyu/common"
	"mitsuyu/mitsuyu"
	"mitsuyu/transport"
	"net"
	"os"
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
	padding		  int
	compress      string
	serviceName   string
	strategyGroup []*common.Strategy
	logger        *common.Logger
	conns         *common.Connector
	stats         *common.Statistician
	done          chan struct{}
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

	c.padding,_ = strconv.Atoi(config.Padding)
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

	// load log level
	c.logger = common.NewLogger(config.LogLevel)

	// enable statistic
	c.conns = common.NewConnector()
	// enable network limit
	uplimit, _ := strconv.Atoi(config.UpLimit)
	downlimit, _ := strconv.Atoi(config.DownLimit)
	c.stats = common.NewStatistician(uplimit*1024, downlimit*1024)
	return c, nil
}

func (c *Client) Local() string {
	return c.local
}

func (c *Client) Remote() string {
	return c.remote
}

func (c *Client) SetLocal(local string) {
	c.local = local
}

func (c *Client) SetRemote(remote string) {
	c.remote = remote
}

func (c *Client) SetTLSSNI(sni string) {
	c.tls = &tls.Config{
		ServerName:         sni,
		InsecureSkipVerify: false,
	}
}

func (c *Client) SetCompress(b bool) {
	if b {
		c.compress = "true"
	} else {
		c.compress = "false"
	}
}

func (c *Client) GetServiceName() string {
	return c.serviceName
}

func (c *Client) GetSummary() []string {
	ss := make([]string, 0, 6)
	ss = append(ss, fmt.Sprintf("service: %s", c.serviceName))
	ss = append(ss, fmt.Sprintf("local_addr: %s", c.local))
	ss = append(ss, fmt.Sprintf("remote_addr: %s", c.remote))
	ss = append(ss, fmt.Sprintf("use_tls: %t", c.tls != nil))
	ss = append(ss, fmt.Sprintf("tls_sni: %s", c.tls.ServerName))
	ss = append(ss, fmt.Sprintf("compress: %s", c.compress))
	return ss
}

func (c *Client) GetLogger() *common.Logger {
	return c.logger
}

func (c *Client) GetConnector() *common.Connector {
	return c.conns
}

func (c *Client) GetStatistician() *common.Statistician {
	return c.stats
}

func (c *Client) Run() {
	// log info
	c.logger.Infof("__boot__\n")
	c.done = make(chan struct{}, 0)
	lis, err := net.Listen("tcp", c.local)
	if err != nil {
		fmt.Printf("Client: Unable to bind %s, %v\n", c.local, err)
		os.Exit(0)
	}
	defer lis.Close()
	for {
		select {
		case <-c.done:
			// log info
			c.logger.Infof("__shutdown__\n")
			return
		default:
			conn, err := lis.Accept()
			if err != nil {
				// log err
				c.logger.Errorf(fmt.Errorf("Client: Accept failed, %v\n", err))
				continue
			}
			go c.deliver(conn)
		}
	}
}

func (c *Client) Stop() {
	defer func() {
		recover()
	}()
	close(c.done)
	// log info
	c.logger.Infof("request shutdown\n")
}

func (c *Client) CallMitsuyuProxy(md metadata.MD) (*transport.GRPCStreamClient, error) {
	// log debug
	c.logger.Debugf("Outbound: Dial gRPC\n")

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
		// log error
		c.logger.Errorf(fmt.Errorf("Outbound: Dial gRPC timeout, %v\n", err))
		return nil, err
	}
	cc := mitsuyu.NewMitsuyuClient(grpcConn, c.serviceName)

	ctx := metadata.NewOutgoingContext(context.Background(), md)
	// call grpc func
	var callopts []grpc.CallOption
	if c.compress == "true" {
		callopts = append(callopts, grpc.UseCompressor(gzip.Name))
	}
	// log debug
	c.logger.Debugf("Outbound: Create stream\n")
	stream, err := cc.Proxy(ctx, callopts...)
	if err != nil {
		// log error
		c.logger.Errorf(fmt.Errorf("Outbound: Failed to create stream, %v\n", err))
		return nil, err
	}
	ccc := transport.NewGRPCStreamClient(grpcConn, stream)
	return ccc, nil
}

func (c *Client) deliver(conn net.Conn) {
	// log debug
	c.logger.Debugf("Client: Select inbound protocol\n")
	defer conn.Close()
	// log debug
	c.logger.Debugf("Client: Read first package\n")
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		// log error
		c.logger.Errorf(fmt.Errorf("Client: Failed to read first package, %v\n", err))
		return
	}
	if s5, err := transport.Socks5Handshake(buf[:n], conn); err == nil {
		c.handle(s5)
	} else if h, err := transport.HttpHandshake(buf[:n], conn); err == nil {
		c.handle(h)
	} else if rawTCP, err := transport.NewRawTCPFromRedirect(buf[:n], conn); err == nil &&
		rawTCP.Addr().Host+":"+rawTCP.Addr().Port != c.local {
		c.handle(rawTCP)
	} else if rawTCP, err := transport.NewRawTCPWithSniff(buf[:n], conn); err == nil {
		c.handle(rawTCP)
	} else {
		// log error
		c.logger.Errorf(fmt.Errorf("Client: Unknown protocol\n"))
	}
}

func (c *Client) handle(in transport.Inbound) {
	if !in.Addr().Isdn {
		transport.GetDomainName(in)
	}
	// statistic
	c.conns.RecordOpen(in.Addr().Host)

	md := metadata.New(map[string]string{
		"xxhost": in.Addr().Host,
		"port":   in.Addr().Port,
		"isdn":   strconv.FormatBool(in.Addr().Isdn),
		"dns":    "default",
		"next":   "null",
	})
	// log debug
	c.logger.Debugf("Inbound: Prepare metadata\n")
	if allow := c.applyClientStrategy(in.Addr(), md); !allow {
		c.logger.Infof(fmt.Sprintf("%-6s|%s:%s|blocked\n", in.Proto(), in.Addr().Host, in.Addr().Port))
		return
	}

	ccc, err := c.CallMitsuyuProxy(md)
	if err != nil {
		return
	}
	stream := ccc.GetStream()
	wg := new(sync.WaitGroup)
	wg.Add(2)

	c.logger.Infof(fmt.Sprintf("%-6s|%s:%s|dns=%s\n", in.Proto(), in.Addr().Host, in.Addr().Port, md.Get("dns")[0]))
	// forward
	go func() {
		defer ccc.Close()
		defer in.Close()
		buf := make([]byte, BUFFERSIZE)
		// log debug
		c.logger.Debugf("Proxy: Start forward proxy\n")
		for {
			n, err := in.Read(buf)
			if err != nil {
				break
			}
			padd := common.PaddingBytes(n,c.padding)
			if err = stream.Send(&mitsuyu.Data{Data: buf[:n], Tail: padd}); err != nil {
				break
			}
			// statistic uptraffic
			c.stats.RecordUplink(n + len(padd))
		}
		if h, ok := in.(*transport.Http); ok && !h.IsTun() {
			time.Sleep(4 * time.Second)
		}
		// log debug
		c.logger.Debugf("Proxy: Finish forward proxy\n")
		wg.Done()
	}()
	// reverse
	go func() {
		defer ccc.Close()
		defer in.Close()
		// statistic
		var n = 0
		// log debug
		c.logger.Debugf("Proxy: Start reverse proxy\n")
		for {
			r, err := stream.Recv()
			if err != nil {
				break
			}
			if n, err = in.Write(r.GetData()); err != nil {
				break
			}
			// statistic
			c.stats.RecordDownlink(n)
		}
		// log debug
		c.logger.Debugf("Proxy: Finish reverse proxy\n")
		wg.Done()
	}()
	wg.Wait()
	// statistic
	c.conns.RecordClose(in.Addr().Host)
	// log debug
	c.logger.Debugf("Proxy: Done\n")
}

func (c *Client) applyClientStrategy(addr *common.Addr, md metadata.MD) (allow bool) {
	// log debug
	c.logger.Debugf("Strategy: Match rules\n")

	if blockReservedAddr(addr) {
		// log debug
		c.logger.Debugf("Strategy: Block private address\n")
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
		// log debug
		c.logger.Debugf("Strategy: Apply rules\n")
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
