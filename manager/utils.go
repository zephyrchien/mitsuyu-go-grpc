package manager

import (
	"fmt"
)

type LogRecorder struct {
	ch chan string
}

func NewLogRecorder() *LogRecorder {
	ch := make(chan string, 20)
	return &LogRecorder{ch: ch}
}

func (r *LogRecorder) Write(b []byte) (n int, err error) {
	n = len(b)
	if n == 0 {
		return n, fmt.Errorf("Write failed")
	}
	r.ch <- string(b[:n-1])
	return n, nil
}

func (r *LogRecorder) GetChan() chan string {
	return r.ch
}
