package main

import (
	"log"
)

type logger struct {
	prefix string
}

func newLogger(prefix string) *logger {
	return &logger{
		prefix: prefix,
	}
}

func (l *logger) Printf(s string, args ...interface{}) {
	log.Printf("["+l.prefix+"] "+s, args...)
}
