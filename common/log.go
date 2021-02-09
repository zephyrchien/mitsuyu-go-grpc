package common

import (
	"fmt"
	"io"
)

const (
	LOG_NONE  = 0
	LOG_ERR   = 1
	LOG_INFO  = 2
	LOG_DEBUG = 3
)

type Logger struct {
	level int
	err   chan error
	info  chan string
	debug chan string
	done  chan struct{}
}

func NewLogger(levelStr string) *Logger {
	var err chan error
	var info chan string
	var debug chan string
	var level int
	switch levelStr {
	case "err":
		level = LOG_ERR
	case "info":
		level = LOG_INFO
	case "debug":
		level = LOG_DEBUG
	default:
		level = LOG_NONE
	}
	if level >= LOG_ERR {
		err = make(chan error, 5)
	}
	if level >= LOG_INFO {
		info = make(chan string, 10)
	}
	if level >= LOG_DEBUG {
		debug = make(chan string, 20)
	}
	return &Logger{level: level, err: err, info: info, debug: debug}
}

func (l *Logger) GetLevel() int {
	return l.level
}

func (l *Logger) GetErr() chan error {
	return l.err
}

func (l *Logger) GetInfo() chan string {
	return l.info
}

func (l *Logger) GetDebug() chan string {
	return l.debug
}

func (l *Logger) Errorf(err error) {
	if l.level >= LOG_ERR {
		l.err <- err
	}
}

func (l *Logger) Infof(info string) {
	if l.level >= LOG_INFO {
		l.info <- info
	}
}

func (l *Logger) Debugf(debug string) {
	if l.level >= LOG_DEBUG {
		l.debug <- debug
	}
}

func (l *Logger) StartLog(dst io.Writer) {
	l.done = make(chan struct{}, 0)
	for {
		select {
		case <-l.done:
			return
		case e := <-l.err:
			fmt.Fprintf(dst, "%v", e)
		case i := <-l.info:
			fmt.Fprintf(dst, "%s", i)
		case d := <-l.debug:
			fmt.Fprintf(dst, "%s", d)
		}
	}
}

func (l *Logger) StopLog() {
	defer func() {
		recover()
	}()
	close(l.done)
}
