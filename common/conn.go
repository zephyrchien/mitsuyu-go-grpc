package common

import (
	"fmt"
)

type Connector struct {
	enable bool
	done   chan struct{}
	open   chan string
	close  chan string
	conns  map[string]int
}

func NewConnector() *Connector {
	open := make(chan string, 20)
	close := make(chan string, 20)
	conns := make(map[string]int)
	return &Connector{enable: false, open: open, close: close, conns: conns}
}

func (c *Connector) Config(b bool) {
	c.enable = b
}

func (c *Connector) GetOpen() chan string {
	return c.open
}

func (c *Connector) GetClose() chan string {
	return c.close
}

func (c *Connector) RecordOpen(dname string) {
	if c.enable {
		c.open <- dname
	}
}

func (c *Connector) RecordClose(dname string) {
	if c.enable {
		c.close <- dname
	}
}

func (c *Connector) StartRecord() {
	c.done = make(chan struct{}, 0)
	for {
		select {
		case <-c.done:
			return
		case dname := <-c.open:
			if n, ok := c.conns[dname]; ok {
				c.conns[dname] = n + 1
			} else {
				c.conns[dname] = 1
			}
		case dname := <-c.close:
			if n, ok := c.conns[dname]; ok {
				if n <= 1 {
					delete(c.conns, dname)
				} else {
					c.conns[dname] = n - 1
				}
			}
		}
	}
}

func (c *Connector) StopRecord() {
	defer func() {
		recover()
	}()
	close(c.done)
}

func (c *Connector) GetReport() []string {
	r := make([]string, 0, len(c.conns))
	for dname, n := range c.conns {
		r = append(r, fmt.Sprintf("[%d]%s", n, dname))
	}
	return r
}
