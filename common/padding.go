package common

import (
	"bytes"
	"math/rand"
	"time"
)

func PaddingBytes(srclen int) []byte {
	if srclen > 256 || srclen <= 0 {
		return nil
	}
	rand.Seed(time.Now().UnixNano())
	l := 256 - srclen + rand.Intn(256)
	padd := bytes.Repeat([]byte{byte(srclen)}, l)
	return padd
}
