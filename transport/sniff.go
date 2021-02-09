package transport

import (
	"bytes"
	"fmt"
	"github.com/ZephyrChien/Mitsuyu/common"
	"strings"
)

func GetDomainName(in Inbound) {
	buf := make([]byte, 1024)
	n, err := in.Read(buf)
	if err != nil {
		return
	}
	buffer := bytes.NewBuffer(buf[:n])
	in.SetBuffer(buffer)
	host, err := SniffHost(buf[:n])
	if err != nil {
		return
	}
	addr := &common.Addr{Isdn: true, Host: host, Port: in.Addr().Port}
	in.SetAddr(addr)
}

func SniffHost(buf []byte) (string, error) {
	host, err := SniffFromHTTP(buf)
	if err != nil {
		host, err = SniffFromHTTPS(buf)
	}
	if err != nil {
		return "", err
	}
	return host, nil
}

func SniffFromHTTP(buf []byte) (string, error) {
	var errorNotHTTP = fmt.Errorf("Common: sniff from HTTP failed")
	// borrow from http.go
	buff := bytes.Split(buf, []byte("\r\n"))
	_, _, err := parseFirstLine(string(buff[0]))
	if err != nil {
		return "", errorNotHTTP
	}
	header := parseHeader(buff)
	host := header.Get("Host")
	if host == "" {
		return "", errorNotHTTP
	}
	return host, nil
}

func SniffFromHTTPS(buf []byte) (string, error) {
	var errorNotHTTPS = fmt.Errorf("Common: sniff from HTTPS failed")
	// tls
	if len(buf) < 5 || buf[0] != 0x16 {
		return "", errorNotHTTPS
	}
	// version
	if buf[1] != 0x03 || buf[2] < 0x01 || buf[2] > 0x04 {
		return "", errorNotHTTPS
	}
	// max length
	var maxlen = int(buf[3])<<8 | int(buf[4])
	if maxlen+5 > len(buf) || maxlen < 42 {
		return "", errorNotHTTPS
	}
	var index = 43
	var next = 0
	// session id
	next = countNext(buf[index:], 1)
	if next == -1 || index+next+1+2 > maxlen+5 {
		return "", errorNotHTTPS
	}
	index = index + next + 1
	// cipher suits
	next = countNext(buf[index:], 2)
	if next == -1 || next%2 == 1 || index+next+2+1 > maxlen+5 {
		return "", errorNotHTTPS
	}
	index = index + next + 2
	// compression method
	next = countNext(buf[index:], 1)
	if next == -1 || index+next+1+2 > maxlen+5 {
		return "", errorNotHTTPS
	}
	index = index + next + 1
	// extensions
	next = countNext(buf[index:], 2)
	if next == -1 || index+next+2 != maxlen+5 {
		return "", errorNotHTTPS
	}
	index += 2
	for {
		if index+4 > maxlen+5 {
			return "", errorNotHTTPS
		}
		if buf[index] == 0x00 && buf[index+1] == 0x00 {
			break
		}
		next = countNext(buf[index+2:], 2)
		index = index + next + 4
	}
	snilen := countNext(buf[index+2:], 2)
	host := parseSNI(buf[index+4 : index+snilen+4])
	if host == "" {
		return "", errorNotHTTPS
	}
	return host, nil

}

func countNext(buf []byte, n int) int {
	var step int
	if n == 1 {
		step = int(buf[0])
	} else if n == 2 {
		step = int(buf[0])<<8 | int(buf[1])
	} else {
		return -1
	}
	if step+n > len(buf) {
		return -1
	}
	return step
}

func parseSNI(buf []byte) string {
	maxlen := int(buf[0])<<8 | int(buf[1])
	buf = buf[2:]
	if len(buf) != maxlen {
		return ""
	}
	for {
		if len(buf) < 3 {
			return ""
		}
		nameType := buf[0]
		nameLen := int(buf[1])<<8 | int(buf[2])
		buf = buf[3:]
		if len(buf) < nameLen {
			return ""
		}
		if nameType == 0x00 {
			host := string(buf[:nameLen])
			if strings.HasSuffix(host, ".") {
				return ""
			}
			return host
		}
	}

}
