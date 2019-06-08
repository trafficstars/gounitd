package main

import (
	"fmt"
	"log"
	"log/syslog"
)

type loggerSyslog struct {
	prefix string
	writer *syslog.Writer
}

func newLoggerSyslog(prefix string) *loggerSyslog {
	var err error
	r := &loggerSyslog{
		prefix: prefix,
	}
	r.writer, err = syslog.New(syslog.LOG_ERR, "gounitd")
	if err != nil {
		panic(err)
	}
	return r
}

func (l *loggerSyslog) Printf(s string, args ...interface{}) {
	msg := fmt.Sprintf("["+l.prefix+"] "+s, args...)
	err := l.writer.Err(msg)
	if err != nil {
		log.Print(`[error] unable to write to syslog message: "`, msg, `"`)
	}
}
