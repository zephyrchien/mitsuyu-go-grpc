package client

import (
	"bytes"
	"errors"
	"fmt"
	"mitsuyu/common"
	"mitsuyu/mitsuyu"
	"mitsuyu/transport"
	"sync"
	"time"

	"google.golang.org/grpc/metadata"
)

type stream struct {
	conn     *transport.GRPCStreamClient
	deadline time.Time
}

type pool struct {
	cap		int
	lock    sync.Mutex
	streams []*stream
}

var StreamPool *pool

func (c *Client) newStream() (*stream, error) {
	md := metadata.New(map[string]string{
		"xxhost": "null",
		"port":   "null",
		"isdn":   "null",
		"dns":    "null",
		"next":   "null",
		"reuse":  "true",
	})
	conn, err := c.CallMitsuyuProxy(md)
	if err != nil {
		return nil, err
	}
	return &stream{
		conn:     conn,
		deadline: time.Now().Add(time.Second * time.Duration(c.timeout)),
	}, nil
}

func (s *stream) send(data *mitsuyu.Data) error {
	return s.conn.WriteAll(data)
}

func (s *stream) recv() (*mitsuyu.Data, error) {
	return s.conn.ReadAll()
}

func (s *stream) close() error {
	return s.conn.Close()
}

func (s *stream) timeout() bool {
	if time.Now().After(s.deadline) {
		return true
	}
	return false
}

func (p *pool) size() int {
	return len(p.streams)
}

func (p *pool) push(s *stream) error{
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.cap !=0 && len(p.streams) >= p.cap {
		return errors.New("Exceed max capacity")
	}
	p.streams = append(p.streams, s)
	return nil
}

func (p *pool) pop() (*stream, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if len(p.streams) < 1 {
		return nil, errors.New("No available stream")
	}
	s := p.streams[0]
	if len(p.streams) > 1 {
		p.streams = p.streams[1:]
	} else {
		p.streams = nil
	}
	return s, nil
}

func (c *Client) handleReuse(in transport.Inbound, md metadata.MD) {
	// log debug
	c.logger.Debugf("Reuse: Enabled\n")

	md.Set("reuse", "true")
	mdbin := new(bytes.Buffer)

	for key, val := range md {
		if mdbin.Len() != 0 {
			mdbin.WriteString("\r\n")
		}
		mdbin.WriteString(key)
		mdbin.WriteString("\r\n")
		mdbin.WriteString(val[0])
	}

	var stream *stream
	var err error
	if stream, err = StreamPool.pop(); err != nil {
		c.logger.Debugf("Reuse: No available connection\n")
		if stream, err = c.newStream(); err != nil {
			return
		}
		c.logger.Infof(fmt.Sprintf("%-6s|[new]%s:%s|dns=%s\n", in.Proto(), in.Addr().Host, in.Addr().Port, md.Get("dns")[0]))
	}else{
		c.logger.Infof(fmt.Sprintf("%-6s|[reuse]%s:%s|dns=%s\n", in.Proto(), in.Addr().Host, in.Addr().Port, md.Get("dns")[0]))
	}

	wg := new(sync.WaitGroup)
	wg.Add(2)
	// forward
	go func() {
		defer in.Close()
		buf := make([]byte, BUFFERSIZE)
		// log debug
		c.logger.Debugf("Proxy: Start forward proxy\n")
		for {
			n, err := in.Read(buf)
			if err != nil {
				stream.send(&mitsuyu.Data{Head: []byte{transport.CMD,transport.CMD_EOF}})
				break
			}
			padd := common.PaddingBytes(n, c.padding)
			if err = stream.send(&mitsuyu.Data{Data: buf[:n], Tail: padd, Head: mdbin.Bytes()}); err != nil {
				break
			}
			// statistic uptraffic
			c.stats.RecordUplink(n + len(padd) + mdbin.Len())

			mdbin.Reset()
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
		defer in.Close()
		// statistic
		var n = 0
		// log debug
		c.logger.Debugf("Proxy: Start reverse proxy\n")
		for {
			r, err := stream.recv()
			if err != nil {
				break
			}
			if head:=r.GetHead();len(head)==2 && head[0]==transport.CMD&&head[1]==transport.CMD_EOF{
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
	if stream.timeout() {
		c.logger.Debugf("Reuse: Timeout exceeded, close\n")
		stream.close()
	} else {
		StreamPool.push(stream)
		c.logger.Debugf("Reuse: Recycle\n")
	}
}
