package server

import (
	"errors"
	"net"
	"strings"
	"time"

	"github.com/trafficstars/fasthttp"
)

var (
	ErrInvalidListenAddress = errors.New(`[gounit-frontend] Invalid listen address`)
	ErrEmptyListenAddress   = errors.New(`[gounit-frontend] Empty listen address`)
)

type Frontend struct {
	Server             *Server
	ListenFamily       string
	ListenAddress      string
	HTTP               *fasthttp.Server
	Concurrency        int
	IsControl          bool
	ReadBufferSize     uint
	WriteBufferSize    uint
	SetHeadersMap      map[string]string
	MaxRequestBodySize uint
	MaxConnsPerIP      uint
}

func newFrontend(srv *Server, cfg ConfigFrontend) (*Frontend, error) {
	if cfg.Listen == `` {
		return nil, ErrEmptyListenAddress
	}
	words := strings.SplitN(cfg.Listen, `:`, 2)
	if len(words) < 2 {
		return nil, ErrInvalidListenAddress
	}
	f := &Frontend{
		Server:             srv,
		ListenFamily:       words[0],
		ListenAddress:      words[1],
		Concurrency:        cfg.Concurrency,
		IsControl:          cfg.IsControl,
		SetHeadersMap:      cfg.SetHeaders.ToMap(),
		ReadBufferSize:     cfg.ReadBufferSize,
		WriteBufferSize:    cfg.WriteBufferSize,
		MaxRequestBodySize: cfg.MaxRequestBodySize,
	}
	return f, nil
}

func (f *Frontend) handleRequest(ctx *fasthttp.RequestCtx) {
	if f.IsControl {
		writeMetrics(ctx)
		return
	}
	if f.SetHeadersMap != nil {
		f.Server.SetHeaders(ctx, f.SetHeadersMap)
	}
	f.Server.HandleRequest(f, ctx)
}

func (f *Frontend) Start() error {
	f.HTTP = &fasthttp.Server{
		Handler:                       f.handleRequest,
		Name:                          "",
		Concurrency:                   f.Concurrency,
		DisableKeepalive:              false,
		ReadBufferSize:                int(f.ReadBufferSize),
		WriteBufferSize:               int(f.WriteBufferSize),
		ReadTimeout:                   0,
		WriteTimeout:                  0,
		MaxConnsPerIP:                 int(f.MaxConnsPerIP),
		MaxRequestsPerConn:            0,
		MaxKeepaliveDuration:          0,
		TCPKeepalive:                  true,
		TCPKeepalivePeriod:            60,
		MaxRequestBodySize:            int(f.MaxRequestBodySize),
		ReduceMemoryUsage:             false,
		GetOnly:                       false,
		LogAllErrors:                  false,
		DisableHeaderNamesNormalizing: false,
		NoDefaultServerHeader:         false,
		NoDefaultContentType:          false,
		ConnState:                     nil,
		Logger:                        nil,
	}

	ln, err := net.Listen(f.ListenFamily, f.ListenAddress)
	if err != nil {
		return err
	}
	go func(ln net.Listener, srv *fasthttp.Server) {
		if srv.TCPKeepalive {
			if tcpln, ok := ln.(*net.TCPListener); ok {
				srv.Serve(tcpKeepaliveListener{
					TCPListener:     tcpln,
					keepalivePeriod: srv.TCPKeepalivePeriod,
				})
				return
			}
		}
		srv.Serve(ln)
	}(ln, f.HTTP)
	return nil
}

type tcpKeepaliveListener struct {
	*net.TCPListener
	keepalivePeriod time.Duration
}

func (ln tcpKeepaliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	if ln.keepalivePeriod > 0 {
		tc.SetKeepAlivePeriod(ln.keepalivePeriod)
	}
	return tc, nil
}
