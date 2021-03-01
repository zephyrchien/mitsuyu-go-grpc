package common

import (
	"bytes"
	"math/rand"
	"time"
)

func PaddingBytes(srclen, minlen int) []byte {
	if minlen == 0 {
		return nil
	}
	if srclen > minlen || srclen <= 0 {
		return nil
	}
	rand.Seed(time.Now().UnixNano())
	l := minlen - srclen + rand.Intn(256)
	padd := bytes.Repeat([]byte{byte(srclen)}, l)
	return padd
}
