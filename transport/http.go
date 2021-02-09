package transport

import (
	"bytes"
	"fmt"
	"github.com/ZephyrChien/Mitsuyu/common"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type Http struct {
	conn   net.Conn
	addr   *common.Addr
	method string
	proto  string
	buffer *bytes.Buffer
}

func (h *Http) Addr() *common.Addr {
	return h.addr
}

func (h *Http) Proto() string {
	return h.proto
}
func (h *Http) IsTun() bool {
	return h.method == "connect"
}

func (h *Http) SetAddr(addr *common.Addr) {
	h.addr = addr
}

func (h *Http) SetBuffer(buffer *bytes.Buffer) {
	h.buffer = buffer
}

func (h *Http) Read(b []byte) (int, error) {
	if h.buffer == nil {
		return h.conn.Read(b)
	}
	n, err := h.buffer.Read(b)
	h.buffer = nil
	return n, err
}
func (h *Http) Write(b []byte) (int, error) {
	return h.conn.Write(b)
}
func (h *Http) Close() error {
	return h.conn.Close()
}

func HttpHandshake(buf []byte, conn net.Conn) (*Http, error) {
	buff := bytes.Split(buf, []byte("\r\n"))
	method, link, err := parseFirstLine(string(buff[0]))
	if err != nil {
		return nil, wrapErrorHttp(err)
	}

	// https
	if method == "connect" {
		addr := parseHost(link, "443")
		if err = handshakeTunnel(conn); err != nil {
			return nil, wrapErrorHttp(err)
		}
		return &Http{conn: conn, addr: addr, method: "connect", proto: "https"}, nil
	}

	// plain http
	header := parseHeader(buff)
	buffer := new(bytes.Buffer)
	if u, err := url.Parse(link); err != nil {
		buffer.Write(buff[0])
	} else {
		buffer.Write([]byte(fmt.Sprintf("%s %s HTTP/1.1", strings.ToUpper(method), u.RequestURI())))
	}
	buffer.Write([]byte("\r\n"))
	header.Write(buffer)
	buffer.Write([]byte("\r\n"))
	addr := parseHost(header.Get("Host"), "80")
	if l := len(buff); len(buff[l-1]) != 0 || len(buff[l-2]) != 0 {
		buffer.Write(buff[l-1])
		ll, b := buffer.Len(), buffer.Bytes()
		if _, err = conn.Read(b[ll:]); err != nil {
			return nil, err
		}
	}
	return &Http{conn: conn, addr: addr, method: method, proto: "http", buffer: buffer}, nil
}

func parseFirstLine(line string) (method, link string, err error) {
	var methods = [...]string{"head", "get", "post", "connect"}
	strs := strings.SplitN(line, " ", 3)
	if len(strs) != 3 {
		err = fmt.Errorf("Unable to parse header")
		return
	}
	method = strings.ToLower(strings.TrimSpace(strs[0]))
	link = strings.ToLower(strings.TrimSpace(strs[1]))
	version := strings.ToLower(strings.TrimSpace(strs[2]))
	if version != "http/1.1" {
		err = fmt.Errorf("Support HTTP/1.1 only")
		return
	}
	ok := false
	for _, m := range methods {
		if method == m {
			ok = true
			break
		}
	}
	if !ok {
		err = fmt.Errorf("Unable to parse header")
		return
	}
	return method, link, nil
}

func parseHeader(buf [][]byte) http.Header {
	header := make(http.Header, 20)
	for i, l := 1, len(buf); i < l; i++ {
		b := bytes.SplitN(buf[i], []byte(":"), 2)
		if len(b) != 2 {
			continue
		}
		key := bytes.TrimSpace(b[0])
		values := bytes.Split(bytes.TrimSpace(b[1]), []byte(","))
		for _, v := range values {
			header.Add(string(key), string(bytes.TrimSpace(v)))
		}
	}
	header.Del("Proxy-Connection")
	header.Del("Proxy-Authenticate")
	header.Del("Proxy-Authorization")
	header.Del("TE")
	header.Del("Trailers")
	header.Del("Transfer-Encoding")
	header.Del("Upgrade")
	connections := header.Get("Connection")
	if connections == "" {
		return header
	}
	for _, h := range strings.Split(connections, ",") {
		header.Del(strings.TrimSpace(h))
	}
	return header
}

func parseHost(link, dport string) *common.Addr {
	var host, port string
	hs := strings.SplitN(link, ":", 2)
	host = hs[0]
	if len(hs) == 2 {
		port = hs[1]
	} else {
		port = dport
	}
	isdn := net.ParseIP(host) == nil
	return &common.Addr{Host: host, Port: port, Isdn: isdn}
}

func handshakeTunnel(conn net.Conn) error {
	_, err := conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	return err
}
func wrapErrorHttp(err error) error {
	return fmt.Errorf("Http handshake: %v", err)
}
