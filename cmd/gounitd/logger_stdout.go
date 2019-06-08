package main

import (
	"log"
)

type loggerStdout struct {
	prefix string
}

func newLoggerStdout(prefix string) *loggerStdout {
	return &loggerStdout{
		prefix: prefix,
	}
}

func (l *loggerStdout) Printf(s string, args ...interface{}) {
	log.Printf("["+l.prefix+"] "+s, args...)
}
