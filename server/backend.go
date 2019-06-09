package server

import (
	"net/http"
	"os"
	"regexp"

	"github.com/labstack/gommon/log"
	"github.com/trafficstars/fasthttp"
	"github.com/trafficstars/fasthttpsocket"
)

type Backend struct {
	Server                *Server
	Socket                *fasthttpsocket.SocketClient
	Address               string
	URLRegexp             *regexp.Regexp
	Connections           int
	UnixSocketPermissions os.FileMode
	Return                uint
	SetHeadersMap         map[string]string
}

func newBackend(srv *Server, cfg ConfigBackend) (*Backend, error) {
	b := &Backend{
		Server:        srv,
		Address:       cfg.Address,
		Connections:   cfg.Connections,
		Return:        cfg.Return,
		SetHeadersMap: cfg.SetHeaders.ToMap(),
	}
	var err error
	b.URLRegexp, err = regexp.Compile(cfg.URLRegexp)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (b *Backend) Start() error {
	if b.Return > 0 {
		return nil
	}
	var err error
	b.Socket, err = fasthttpsocket.NewSocketClient(fasthttpsocket.Config{
		Address:               b.Address,
		UnixSocketPermissions: b.UnixSocketPermissions,
		Logger:                log.New(`[gounit-backend]`),
	})
	if err != nil {
		return err
	}
	err = b.Socket.Start(b.Connections)
	if err != nil {
		b.Socket = nil
	}
	return err
}

func (b *Backend) IsFits(ctx *fasthttp.RequestCtx) bool {
	if b.URLRegexp == nil {
		return true
	}
	return b.URLRegexp.Match(ctx.Request.URI().FullURI())
}

func (b *Backend) HandleRequest(f *Frontend, ctx *fasthttp.RequestCtx) error {
	if b.Server.AccessLogger != nil {
		b.Server.AccessLogger.Printf("request to %v: %v", b.Address, string(ctx.URI().FullURI()))
	}
	var err error
	if b.Return > 0 {
		ctx.Response.SetStatusCode(int(b.Return))
	} else {
		if b.SetHeadersMap != nil {
			b.Server.SetHeaders(ctx, f.SetHeadersMap)
		}
		err = b.Socket.SendAndReceive(ctx)
		if err != nil {
			ctx.SetStatusCode(http.StatusBadGateway)
			if b.Server.ErrorLogger != nil {
				b.Server.ErrorLogger.Printf("cannot send request to %v: %v", b.Address, err)
			}
		}
	}
	if b.Server.AccessLogger != nil {
		b.Server.AccessLogger.Printf("request-end (%v)", b.Address)
	}
	return err
}
